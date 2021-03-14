// Copyright 2016 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/prometheus-operator/prometheus-operator/pkg/versionutil"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/thanos-io/thanos/pkg/reloader"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	logFormatLogfmt = "logfmt"
	logFormatJSON   = "json"

	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
	logLevelNone  = "none"

	defaultWatchInterval = 3 * time.Minute // 3 minutes was the value previously hardcoded in github.com/thanos-io/thanos/pkg/reloader.
	defaultDelayInterval = 1 * time.Second // 1 second seems a reasonable amount of time for the kubelet to update the secrets/configmaps.
	defaultRetryInterval = 5 * time.Second // 5 seconds was the value previously hardcoded in github.com/thanos-io/thanos/pkg/reloader.

	statefulsetOrdinalEnvvar            = "STATEFULSET_ORDINAL_NUMBER"
	statefulsetOrdinalFromEnvvarDefault = "POD_NAME"
)

var (
	availableLogFormats = []string{
		logFormatLogfmt,
		logFormatJSON,
	}
	availableLogLevels = []string{
		logLevelDebug,
		logLevelInfo,
		logLevelWarn,
		logLevelError,
		logLevelNone,
	}
)

func main() {
	app := kingpin.New("prometheus-config-reloader", "")
	cfgFile := app.Flag("config-file", "config file watched by the reloader").
		String()

	cfgSubstFile := app.Flag("config-envsubst-file", "output file for environment variable substituted config file").
		String()

	watchInterval := app.Flag("watch-interval", "how often the reloader re-reads the configuration file and directories").Default(defaultWatchInterval.String()).Duration()
	delayInterval := app.Flag("delay-interval", "how long the reloader waits before reloading after it has detected a change").Default(defaultDelayInterval.String()).Duration()
	retryInterval := app.Flag("retry-interval", "how long the reloader waits before retrying in case the endpoint returned an error").Default(defaultRetryInterval.String()).Duration()

	watchedDir := app.Flag("watched-dir", "directory to watch non-recursively").Strings()

	createStatefulsetOrdinalFrom := app.Flag(
		"statefulset-ordinal-from-envvar",
		fmt.Sprintf("parse this environment variable to create %s, containing the statefulset ordinal number", statefulsetOrdinalEnvvar)).
		Default(statefulsetOrdinalFromEnvvarDefault).String()

	listenAddress := app.Flag(
		"listen-address",
		"address on which to expose metrics (disabled when empty)").
		String()

	logFormat := app.Flag(
		"log-format",
		fmt.Sprintf("log format to use. Possible values: %s", strings.Join(availableLogFormats, ", "))).
		Default(logFormatLogfmt).String()

	logLevel := app.Flag(
		"log-level",
		fmt.Sprintf("log level to use. Possible values: %s", strings.Join(availableLogLevels, ", "))).
		Default(logLevelInfo).String()

	reloadURL := app.Flag("reload-url", "reload URL to trigger Prometheus reload on").
		Default("http://127.0.0.1:9090/-/reload").URL()

	versionutil.RegisterIntoKingpinFlags(app)

	if _, err := app.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if versionutil.ShouldPrintVersion() {
		versionutil.Print(os.Stdout, "prometheus-config-reloader")
		os.Exit(0)
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))

	if *logFormat == logFormatJSON {
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	}

	switch *logLevel {
	case logLevelDebug:
		logger = level.NewFilter(logger, level.AllowDebug())
	case logLevelWarn:
		logger = level.NewFilter(logger, level.AllowWarn())
	case logLevelError:
		logger = level.NewFilter(logger, level.AllowError())
	case logLevelNone:
		logger = level.NewFilter(logger, level.AllowNone())
	default:
		logger = level.NewFilter(logger, level.AllowInfo())
	}

	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	if createStatefulsetOrdinalFrom != nil {
		if err := createOrdinalEnvvar(*createStatefulsetOrdinalFrom); err != nil {
			level.Warn(logger).Log("msg", fmt.Sprintf("Failed setting %s", statefulsetOrdinalEnvvar))
		}
	}

	level.Info(logger).Log("msg", "Starting prometheus-config-reloader", "version", version.Info())
	level.Info(logger).Log("build_context", version.BuildContext())

	r := prometheus.NewRegistry()
	r.MustRegister(
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)

	var g run.Group
	{
		ctx, cancel := context.WithCancel(context.Background())
		rel := reloader.New(
			logger,
			r,
			&reloader.Options{
				ReloadURL:     *reloadURL,
				CfgFile:       *cfgFile,
				CfgOutputFile: *cfgSubstFile,
				WatchedDirs:   *watchedDir,
				DelayInterval: *delayInterval,
				WatchInterval: *watchInterval,
				RetryInterval: *retryInterval,
			},
		)

		g.Add(func() error {
			return rel.Watch(ctx)
		}, func(error) {
			cancel()
		})
	}

	if *listenAddress != "" {
		g.Add(func() error {
			level.Info(logger).Log("msg", "Starting web server for metrics", "listen", *listenAddress)
			http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{Registry: r}))
			return http.ListenAndServe(*listenAddress, nil)
		}, func(err error) {
			level.Error(logger).Log("err", err)
		})
	}

	if err := g.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func createOrdinalEnvvar(fromName string) error {
	reg := regexp.MustCompile(`\d+$`)
	val := reg.FindString(os.Getenv(fromName))
	return os.Setenv(statefulsetOrdinalEnvvar, val)
}
