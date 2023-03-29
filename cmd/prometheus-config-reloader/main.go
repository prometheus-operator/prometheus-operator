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
	"crypto/tls"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	logging "github.com/prometheus-operator/prometheus-operator/internal/log"
	"github.com/prometheus-operator/prometheus-operator/pkg/versionutil"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/thanos-io/thanos/pkg/reloader"
)

const (
	defaultWatchInterval = 3 * time.Minute  // 3 minutes was the value previously hardcoded in github.com/thanos-io/thanos/pkg/reloader.
	defaultDelayInterval = 1 * time.Second  // 1 second seems a reasonable amount of time for the kubelet to update the secrets/configmaps.
	defaultRetryInterval = 5 * time.Second  // 5 seconds was the value previously hardcoded in github.com/thanos-io/thanos/pkg/reloader.
	defaultReloadTimeout = 30 * time.Second // 30 seconds was the default value

	statefulsetOrdinalEnvvar            = "STATEFULSET_ORDINAL_NUMBER"
	statefulsetOrdinalFromEnvvarDefault = "POD_NAME"
)

func main() {
	app := kingpin.New("prometheus-config-reloader", "")
	cfgFile := app.Flag("config-file", "config file watched by the reloader").
		String()

	cfgSubstFile := app.Flag("config-envsubst-file", "output file for environment variable substituted config file").
		String()

	watchInterval := app.Flag("watch-interval", "how often the reloader re-reads the configuration file and directories; when set to 0, the program runs only once and exits").Default(defaultWatchInterval.String()).Duration()
	delayInterval := app.Flag("delay-interval", "how long the reloader waits before reloading after it has detected a change").Default(defaultDelayInterval.String()).Duration()
	retryInterval := app.Flag("retry-interval", "how long the reloader waits before retrying in case the endpoint returned an error").Default(defaultRetryInterval.String()).Duration()
	reloadTimeout := app.Flag("reload-timeout", "how long the reloader waits for a response from the reload URL").Default(defaultReloadTimeout.String()).Duration()

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
		fmt.Sprintf("log format to use. Possible values: %s", strings.Join(logging.AvailableLogFormats, ", "))).
		Default(logging.FormatLogFmt).String()

	logLevel := app.Flag(
		"log-level",
		fmt.Sprintf("log level to use. Possible values: %s", strings.Join(logging.AvailableLogLevels, ", "))).
		Default(logging.LevelInfo).String()

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

	logger, err := logging.NewLogger(*logLevel, *logFormat)
	if err != nil {
		stdlog.Fatal(err)
	}

	if createStatefulsetOrdinalFrom != nil {
		if err := createOrdinalEnvvar(*createStatefulsetOrdinalFrom); err != nil {
			level.Warn(logger).Log("msg", fmt.Sprintf("Failed setting %s", statefulsetOrdinalEnvvar))
		}
	}

	level.Info(logger).Log("msg", "Starting prometheus-config-reloader", "version", version.Info())
	level.Info(logger).Log("build_context", version.BuildContext())

	r := prometheus.NewRegistry()
	r.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
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

		client := createHTTPClient(reloadTimeout)
		rel.SetHttpClient(client)

		g.Add(func() error {
			return rel.Watch(ctx)
		}, func(error) {
			cancel()
		})
	}
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"up"}`))
	}
	if *listenAddress != "" && *watchInterval != 0 {
		g.Add(func() error {
			level.Info(logger).Log("msg", "Starting web server for metrics", "listen", *listenAddress)
			http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{Registry: r}))
			http.HandleFunc("/healthz", f)
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

func createHTTPClient(timeout *time.Duration) http.Client {
	transport := (http.DefaultTransport.(*http.Transport)).Clone() // Use the default transporter for production and future changes ready settings.

	transport.DialContext = (&net.Dialer{
		Timeout:   *timeout, // How long should we wait to connect to Prometheus
		KeepAlive: -1,       // Keep alive probe is unnecessary
	}).DialContext

	transport.DisableKeepAlives = true                        // Connection pooling isn't applicable here.
	transport.MaxConnsPerHost = transport.MaxIdleConnsPerHost // Can only have x connections per host, if it is higher than this value something is wrong. Set to max idle as this is a sensible default.

	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true, // TLS certificate verification is disabled by default.
	}

	return http.Client{
		Timeout:   *timeout, // This timeout includes DNS + connect + sending request + reading response
		Transport: transport,
	}
}

func createOrdinalEnvvar(fromName string) error {
	reg := regexp.MustCompile(`\d+$`)
	val := reg.FindString(os.Getenv(fromName))
	return os.Setenv(statefulsetOrdinalEnvvar, val)
}
