// Copyright The prometheus-operator Authors
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
	"strings"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"
	"golang.org/x/sync/errgroup"

	"github.com/prometheus-operator/prometheus-operator/internal/goruntime"
	logging "github.com/prometheus-operator/prometheus-operator/internal/log"
	"github.com/prometheus-operator/prometheus-operator/internal/metrics"
	"github.com/prometheus-operator/prometheus-operator/pkg/admission"
	"github.com/prometheus-operator/prometheus-operator/pkg/server"
	"github.com/prometheus-operator/prometheus-operator/pkg/versionutil"
)

const defaultGOMemlimitRatio = 0.0
const defaultValidationScheme = "legacy"

func main() {
	var (
		serverConfig         = server.DefaultConfig(":8443", true)
		flagset              = flag.CommandLine
		logConfig            logging.Config
		memlimitRatio        float64
		nameValidationScheme string
		promqlOptionsStr     string
	)

	server.RegisterFlags(flagset, &serverConfig)
	versionutil.RegisterFlags(flagset)
	logging.RegisterFlags(flagset, &logConfig)

	flagset.Float64Var(&memlimitRatio, "auto-gomemlimit-ratio", defaultGOMemlimitRatio, "The ratio of reserved GOMEMLIMIT memory to the detected maximum container or system memory. The value should be greater than 0.0 and less than 1.0. Default: 0.0 (disabled).")
	flagset.StringVar(&nameValidationScheme, "name-validation-scheme", defaultValidationScheme, "The name validation scheme to use ('legacy' or 'utf8').")
	flagset.StringVar(&promqlOptionsStr, "promql-options", "", "Comma-separated list of PromQL parser options to enable. Valid values: experimental-functions, duration-expression-parsing, extended-range-selectors, binop-fill-modifiers.")

	_ = flagset.Parse(os.Args[1:])

	if versionutil.ShouldPrintVersion() {
		versionutil.Print(os.Stdout, "admission-webhook")
		return
	}

	logger, err := logging.NewLoggerSlog(logConfig)
	if err != nil {
		stdlog.Fatal(err)
	}

	goruntime.SetMemLimit(logger, memlimitRatio)

	// Parse and validate the name validation scheme
	var validationScheme model.ValidationScheme
	switch nameValidationScheme {
	case "utf8":
		validationScheme = model.UTF8Validation
	case "legacy":
		validationScheme = model.LegacyValidation
	default:
		logger.Error("invalid name validation scheme", "scheme", nameValidationScheme, "supported", []string{"utf8", "legacy"})
		os.Exit(1)
	}

	var parserOptions parser.Options
	for opt := range strings.SplitSeq(promqlOptionsStr, ",") {
		switch strings.TrimSpace(opt) {
		case "":
			// skip empty string (flag not set)
		case "experimental-functions":
			parserOptions.EnableExperimentalFunctions = true
		case "duration-expression-parsing":
			parserOptions.ExperimentalDurationExpr = true
		case "extended-range-selectors":
			parserOptions.EnableExtendedRangeSelectors = true
		case "binop-fill-modifiers":
			parserOptions.EnableBinopFillModifiers = true
		default:
			logger.Error("invalid -promql-options value", "value", opt, "supported", []string{"experimental-functions", "duration-expression-parsing", "extended-range-selectors", "binop-fill-modifiers"})
			os.Exit(1)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg, ctx := errgroup.WithContext(ctx)

	mux := http.NewServeMux()
	admit := admission.New(logger.With("component", "admissionwebhook"), validationScheme, parserOptions)
	admit.Register(mux)

	r := metrics.NewRegistry("prometheus_operator_admission_webhook")

	mux.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"up"}`))
	})

	srv, err := server.NewServer(logger, &serverConfig, mux)
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
