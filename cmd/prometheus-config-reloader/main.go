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
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/thanos-io/thanos/pkg/reloader"

	"github.com/prometheus-operator/prometheus-operator/internal/goruntime"
	logging "github.com/prometheus-operator/prometheus-operator/internal/log"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"github.com/prometheus-operator/prometheus-operator/pkg/versionutil"
)

const (
	defaultWatchInterval = 3 * time.Minute  // 3 minutes was the value previously hardcoded in github.com/thanos-io/thanos/pkg/reloader.
	defaultDelayInterval = 1 * time.Second  // 1 second seems a reasonable amount of time for the kubelet to update the secrets/configmaps.
	defaultRetryInterval = 5 * time.Second  // 5 seconds was the value previously hardcoded in github.com/thanos-io/thanos/pkg/reloader.
	defaultReloadTimeout = 30 * time.Second // 30 seconds was the default value

	defaultGOMemlimitRatio = "0.0"

	httpReloadMethod   = "http"
	signalReloadMethod = "signal"

	statefulsetOrdinalEnvvar = "STATEFULSET_ORDINAL_NUMBER"
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

	memlimitRatio := app.Flag("auto-gomemlimit-ratio", "The ratio of reserved GOMEMLIMIT memory to the detected maximum container or system memory. Default: 0 (disabled)").Default(defaultGOMemlimitRatio).Float64()

	watchedDir := app.Flag("watched-dir", "directory to watch non-recursively").Strings()

	reloadMethod := app.Flag("reload-method", "method used to reload the configuration").Default(httpReloadMethod).Enum(httpReloadMethod, signalReloadMethod)
	processName := app.Flag("process-executable-name", "executable name used to match the process when using the signal reload method").Default("prometheus").String()

	createStatefulsetOrdinalFrom := app.Flag(
		"statefulset-ordinal-from-envvar",
		fmt.Sprintf("parse this environment variable to create %s, containing the statefulset ordinal number", statefulsetOrdinalEnvvar)).
		Default(operator.PodNameEnvVar).String()

	listenAddress := app.Flag(
		"listen-address",
		"address on which to expose metrics (disabled when empty)").
		String()

	webConfig := app.Flag(
		"web-config-file",
		"[EXPERIMENTAL] Path to configuration file that can enable TLS or authentication. See: https://prometheus.io/docs/prometheus/latest/configuration/https/",
	).Default("").String()

	var logConfig logging.Config
	app.Flag(
		"log-format",
		fmt.Sprintf("log format to use. Possible values: %s", strings.Join(logging.AvailableLogFormats, ", "))).
		Default(logging.FormatLogFmt).StringVar(&logConfig.Format)

	app.Flag(
		"log-level",
		fmt.Sprintf("log level to use. Possible values: %s", strings.Join(logging.AvailableLogLevels, ", "))).
		Default(logging.LevelInfo).StringVar(&logConfig.Level)

	reloadURL := app.Flag("reload-url", "URL to trigger the configuration").
		Default("http://127.0.0.1:9090/-/reload").URL()

	runtimeInfoURL := app.Flag("runtimeinfo-url", "URL to check the status of the runtime configuration").
		Default("http://127.0.0.1:9090/api/v1/status/runtimeinfo").URL()

	versionutil.RegisterIntoKingpinFlags(app)

	if _, err := app.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(2)
	}

	if versionutil.ShouldPrintVersion() {
		versionutil.Print(os.Stdout, "prometheus-config-reloader")
		os.Exit(0)
	}

	logger, err := logging.NewLogger(logConfig)
	if err != nil {
		stdlog.Fatal(err)
	}

	err = web.Validate(*webConfig)
	if err != nil {
		level.Error(logger).Log("msg", "Unable to validate web configuration file", "err", err)
		os.Exit(2)
	}

	if createStatefulsetOrdinalFrom != nil {
		if err := createOrdinalEnvvar(*createStatefulsetOrdinalFrom); err != nil {
			level.Warn(logger).Log("msg", fmt.Sprintf("Failed setting %s", statefulsetOrdinalEnvvar))
		}
	}

	level.Info(logger).Log("msg", "Starting prometheus-config-reloader", "version", version.Info())
	level.Info(logger).Log("build_context", version.BuildContext())
	goruntime.SetMaxProcs(logger)
	goruntime.SetMemLimit(logger, *memlimitRatio)

	r := prometheus.NewRegistry()
	r.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	var (
		g           run.Group
		ctx, cancel = context.WithCancel(context.Background())
	)

	{
		opts := reloader.Options{
			CfgFile:       *cfgFile,
			CfgOutputFile: *cfgSubstFile,
			WatchedDirs:   *watchedDir,
			DelayInterval: *delayInterval,
			WatchInterval: *watchInterval,
			RetryInterval: *retryInterval,
		}

		switch *reloadMethod {
		case signalReloadMethod:
			opts.RuntimeInfoURL = *runtimeInfoURL
			opts.ProcessName = *processName
		default:
			opts.ReloadURL = *reloadURL
			opts.HTTPClient = createHTTPClient(reloadTimeout)
		}

		rel := reloader.New(
			logger,
			r,
			&opts,
		)

		g.Add(func() error {
			return rel.Watch(ctx)
		}, func(error) {
			cancel()
		})
	}

	if *listenAddress != "" && *watchInterval != 0 {
		http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{Registry: r}))
		http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"up"}`))
		})

		srv := &http.Server{}

		g.Add(func() error {
			level.Info(logger).Log("msg", "Starting web server for metrics", "listen", *listenAddress)
			return web.ListenAndServe(srv, &web.FlagConfig{
				WebListenAddresses: &[]string{*listenAddress},
				WebConfigFile:      webConfig,
			}, logger)
		}, func(error) {
			srv.Close()
		})
	}

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	g.Add(func() error {
		select {
		case <-term:
			level.Info(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
		case <-ctx.Done():
		}

		return nil
	}, func(error) {})

	if err := g.Run(); err != nil {
		level.Error(logger).Log("msg", "Failed to run", "err", err)
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
