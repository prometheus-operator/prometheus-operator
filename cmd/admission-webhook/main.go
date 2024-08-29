// Copyright 2022 The prometheus-operator Authors
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
	"flag"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"

	"github.com/prometheus-operator/prometheus-operator/internal/goruntime"
	logging "github.com/prometheus-operator/prometheus-operator/internal/log"
	"github.com/prometheus-operator/prometheus-operator/pkg/admission"
	"github.com/prometheus-operator/prometheus-operator/pkg/server"
	"github.com/prometheus-operator/prometheus-operator/pkg/versionutil"
)

const defaultGOMemlimitRatio = 0.0

func main() {
	var (
		serverConfig  server.Config = server.DefaultConfig(":8443", true)
		flagset                     = flag.CommandLine
		logConfig     logging.Config
		memlimitRatio float64
	)

	server.RegisterFlags(flagset, &serverConfig)
	versionutil.RegisterFlags(flagset)
	logging.RegisterFlags(flagset, &logConfig)

	flagset.Float64Var(&memlimitRatio, "auto-gomemlimit-ratio", defaultGOMemlimitRatio, "The ratio of reserved GOMEMLIMIT memory to the detected maximum container or system memory. The value should be greater than 0.0 and less than 1.0. Default: 0.0 (disabled).")

	_ = flagset.Parse(os.Args[1:])

	if versionutil.ShouldPrintVersion() {
		versionutil.Print(os.Stdout, "admission-webhook")
		return
	}

	logger, err := logging.NewLoggerSlog(logConfig)
	if err != nil {
		stdlog.Fatal(err)
	}

	// We're currently migrating our logging library from go-kit to slog.
	// The go-kit logger is being removed in small PRs. For now, we are creating 2 loggers to avoid breaking changes and
	// to have a smooth transition.
	goKitLogger, err := logging.NewLogger(logConfig)
	if err != nil {
		stdlog.Fatal(err)
	}

	goruntime.SetMaxProcs(logger)
	goruntime.SetMemLimit(logger, memlimitRatio)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg, ctx := errgroup.WithContext(ctx)

	mux := http.NewServeMux()
	admit := admission.New(logger.With("component", "admissionwebhook"))
	admit.Register(mux)

	r := prometheus.NewRegistry()
	r.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		version.NewCollector("prometheus_operator_admission_webhook"),
	)
	mux.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"up"}`))
	})

	srv, err := server.NewServer(goKitLogger, &serverConfig, mux)
	if err != nil {
		logger.Error("failed to create web server", "err", err)
		os.Exit(1)
	}

	wg.Go(func() error {
		return srv.Serve(ctx)
	})

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-term:
		logger.Info("Received signal, exiting gracefully...", "signal", sig.String())
	case <-ctx.Done():
	}

	if err := srv.Shutdown(ctx); err != nil {
		logger.Warn("Server shutdown error", "err", err)
	}

	cancel()
	if err := wg.Wait(); err != nil {
		logger.Warn("Unhandled error received. Exiting...", "err", err)
		os.Exit(1)
	}
}
