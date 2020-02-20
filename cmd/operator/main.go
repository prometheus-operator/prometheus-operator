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
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/coreos/prometheus-operator/pkg/admission"
	alertmanagercontroller "github.com/coreos/prometheus-operator/pkg/alertmanager"
	"github.com/coreos/prometheus-operator/pkg/api"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	prometheuscontroller "github.com/coreos/prometheus-operator/pkg/prometheus"
	thanoscontroller "github.com/coreos/prometheus-operator/pkg/thanos"
	"github.com/coreos/prometheus-operator/pkg/version"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

const (
	logLevelAll   = "all"
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
	logLevelNone  = "none"
)

const (
	logFormatLogfmt = "logfmt"
	logFormatJson   = "json"
)

var (
	ns             = namespaces{}
	deniedNs       = namespaces{}
	prometheusNs   = namespaces{}
	alertmanagerNs = namespaces{}
	thanosRulerNs  = namespaces{}
)

type namespaces map[string]struct{}

// Set implements the flagset.Value interface.
func (n namespaces) Set(value string) error {
	if n == nil {
		return errors.New("expected n of type namespaces to be initialized")
	}
	ns := strings.Split(value, ",")
	for i := range ns {
		n[ns[i]] = struct{}{}
	}
	return nil
}

// String implements the flagset.Value interface.
func (n namespaces) String() string {
	return strings.Join(n.asSlice(), ",")
}

func (n namespaces) asSlice() []string {
	var ns []string
	for k := range n {
		ns = append(ns, k)
	}
	return ns
}

func serve(srv *http.Server, listener net.Listener, logger log.Logger) func() error {
	return func() error {
		logger.Log("msg", "Staring insecure server on :8080")
		if err := srv.Serve(listener); err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}

var (
	cfg                prometheuscontroller.Config
	availableLogLevels = []string{
		logLevelAll,
		logLevelDebug,
		logLevelInfo,
		logLevelWarn,
		logLevelError,
		logLevelNone,
	}
	availableLogFormats = []string{
		logFormatLogfmt,
		logFormatJson,
	}
)

func init() {
	cfg.CrdKinds = monitoringv1.DefaultCrdKinds
	flagset := flag.CommandLine
	klog.InitFlags(flagset)
	flagset.StringVar(&cfg.Host, "apiserver", "", "API Server addr, e.g. ' - NOT RECOMMENDED FOR PRODUCTION - http://127.0.0.1:8080'. Omit parameter to run in on-cluster mode and utilize the service account token.")
	flagset.StringVar(&cfg.TLSConfig.CertFile, "cert-file", "", " - NOT RECOMMENDED FOR PRODUCTION - Path to public TLS certificate file.")
	flagset.StringVar(&cfg.TLSConfig.KeyFile, "key-file", "", "- NOT RECOMMENDED FOR PRODUCTION - Path to private TLS certificate file.")
	flagset.StringVar(&cfg.TLSConfig.CAFile, "ca-file", "", "- NOT RECOMMENDED FOR PRODUCTION - Path to TLS CA file.")
	flagset.StringVar(&cfg.KubeletObject, "kubelet-service", "", "Service/Endpoints object to write kubelets into in format \"namespace/name\"")
	flagset.BoolVar(&cfg.TLSInsecure, "tls-insecure", false, "- NOT RECOMMENDED FOR PRODUCTION - Don't verify API server's CA certificate.")
	// The Prometheus config reloader image is released along with the
	// Prometheus Operator image, tagged with the same semver version. Default to
	// the Prometheus Operator version if no Prometheus config reloader image is
	// specified.
	flagset.StringVar(&cfg.PrometheusConfigReloaderImage, "prometheus-config-reloader", fmt.Sprintf("quay.io/coreos/prometheus-config-reloader:v%v", version.Version), "Prometheus config reloader image")
	flagset.StringVar(&cfg.ConfigReloaderImage, "config-reloader-image", "jimmidyson/configmap-reload:v0.3.0", "Reload Image")
	flagset.StringVar(&cfg.ConfigReloaderCPU, "config-reloader-cpu", "100m", "Config Reloader CPU. Value \"0\" disables it and causes no limit to be configured.")
	flagset.StringVar(&cfg.ConfigReloaderMemory, "config-reloader-memory", "25Mi", "Config Reloader Memory. Value \"0\" disables it and causes no limit to be configured.")
	flagset.StringVar(&cfg.AlertmanagerDefaultBaseImage, "alertmanager-default-base-image", "quay.io/prometheus/alertmanager", "Alertmanager default base image")
	flagset.StringVar(&cfg.PrometheusDefaultBaseImage, "prometheus-default-base-image", "quay.io/prometheus/prometheus", "Prometheus default base image")
	flagset.StringVar(&cfg.ThanosDefaultBaseImage, "thanos-default-base-image", "quay.io/thanos/thanos", "Thanos default base image")
	flagset.Var(ns, "namespaces", "Namespaces to scope the interaction of the Prometheus Operator and the apiserver (allow list). This is mutually exclusive with --deny-namespaces.")
	flagset.Var(deniedNs, "deny-namespaces", "Namespaces not to scope the interaction of the Prometheus Operator (deny list). This is mutually exclusive with --namespaces.")
	flagset.Var(prometheusNs, "prometheus-instance-namespaces", "Namespaces where Prometheus custom resources and corresponding Secrets, Configmaps and StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for Prometheus custom resources.")
	flagset.Var(alertmanagerNs, "alertmanager-instance-namespaces", "Namespaces where Alertmanager custom resources and corresponding StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for Alertmanager custom resources.")
	flagset.Var(thanosRulerNs, "thanos-ruler-instance-namespaces", "Namespaces where ThanosRuler custom resources and corresponding StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for ThanosRuler custom resources.")
	flagset.Var(&cfg.Labels, "labels", "Labels to be add to all resources created by the operator")
	flagset.Var(&cfg.CrdKinds, "crd-kinds", " - EXPERIMENTAL (could be removed in future releases) - customize CRD kind names")
	flagset.BoolVar(&cfg.EnableValidation, "with-validation", true, "Include the validation spec in the CRD")
	flagset.StringVar(&cfg.LocalHost, "localhost", "localhost", "EXPERIMENTAL (could be removed in future releases) - Host used to communicate between local services on a pod. Fixes issues where localhost resolves incorrectly.")
	flagset.StringVar(&cfg.LogLevel, "log-level", logLevelInfo, fmt.Sprintf("Log level to use. Possible values: %s", strings.Join(availableLogLevels, ", ")))
	flagset.StringVar(&cfg.LogFormat, "log-format", logFormatLogfmt, fmt.Sprintf("Log format to use. Possible values: %s", strings.Join(availableLogFormats, ", ")))
	flagset.BoolVar(&cfg.ManageCRDs, "manage-crds", true, "Manage all CRDs with the Prometheus Operator.")
	flagset.StringVar(&cfg.PromSelector, "prometheus-instance-selector", "", "Label selector to filter Prometheus CRDs to manage")
	flagset.StringVar(&cfg.AlertManagerSelector, "alertmanager-instance-selector", "", "Label selector to filter AlertManager CRDs to manage")
	flagset.StringVar(&cfg.ThanosRulerSelector, "thanos-ruler-instance-selector", "", "Label selector to filter ThanosRuler CRDs to manage")
	flagset.Parse(os.Args[1:])

	cfg.Namespaces.AllowList = ns.asSlice()
	if len(cfg.Namespaces.AllowList) == 0 {
		cfg.Namespaces.AllowList = append(cfg.Namespaces.AllowList, v1.NamespaceAll)
	}

	cfg.Namespaces.DenyList = deniedNs.asSlice()
	cfg.Namespaces.PrometheusAllowList = prometheusNs.asSlice()
	cfg.Namespaces.AlertmanagerAllowList = alertmanagerNs.asSlice()
	cfg.Namespaces.ThanosRulerAllowList = thanosRulerNs.asSlice()

	if len(cfg.Namespaces.PrometheusAllowList) == 0 {
		cfg.Namespaces.PrometheusAllowList = cfg.Namespaces.AllowList
	}

	if len(cfg.Namespaces.AlertmanagerAllowList) == 0 {
		cfg.Namespaces.AlertmanagerAllowList = cfg.Namespaces.AllowList
	}

	if len(cfg.Namespaces.ThanosRulerAllowList) == 0 {
		cfg.Namespaces.ThanosRulerAllowList = cfg.Namespaces.AllowList
	}
}

func Main() int {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	if cfg.LogFormat == logFormatJson {
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	}
	switch cfg.LogLevel {
	case logLevelAll:
		logger = level.NewFilter(logger, level.AllowAll())
	case logLevelDebug:
		logger = level.NewFilter(logger, level.AllowDebug())
	case logLevelInfo:
		logger = level.NewFilter(logger, level.AllowInfo())
	case logLevelWarn:
		logger = level.NewFilter(logger, level.AllowWarn())
	case logLevelError:
		logger = level.NewFilter(logger, level.AllowError())
	case logLevelNone:
		logger = level.NewFilter(logger, level.AllowNone())
	default:
		fmt.Fprintf(os.Stderr, "log level %v unknown, %v are possible values", cfg.LogLevel, availableLogLevels)
		return 1
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	logger.Log("msg", fmt.Sprintf("Starting Prometheus Operator version '%v'.", version.Version))

	if len(ns) > 0 && len(deniedNs) > 0 {
		fmt.Fprint(os.Stderr, "--namespaces and --deny-namespaces are mutually exclusive. Please provide only one of them.\n")
		return 1
	}

	r := prometheus.NewRegistry()
	po, err := prometheuscontroller.New(cfg, log.With(logger, "component", "prometheusoperator"), r)
	if err != nil {
		fmt.Fprint(os.Stderr, "instantiating prometheus controller failed: ", err)
		return 1
	}

	ao, err := alertmanagercontroller.New(cfg, log.With(logger, "component", "alertmanageroperator"), r)
	if err != nil {
		fmt.Fprint(os.Stderr, "instantiating alertmanager controller failed: ", err)
		return 1
	}

	to, err := thanoscontroller.New(cfg, log.With(logger, "component", "thanosoperator"), r)
	if err != nil {
		fmt.Fprint(os.Stderr, "instantiating thanos controller failed: ", err)
		return 1
	}

	mux := http.NewServeMux()
	web, err := api.New(cfg, log.With(logger, "component", "api"))
	if err != nil {
		fmt.Fprint(os.Stderr, "instantiating api failed: ", err)
		return 1
	}
	admit := admission.New(log.With(logger, "component", "admissionwebhook"))

	web.Register(mux)
	admit.Register(mux)
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Fprint(os.Stderr, "listening port 8080 failed", err)
		return 1
	}

	validationTriggeredCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_rule_validation_triggered_total",
		Help: "Number of times a prometheusRule object triggered validation",
	})

	validationErrorsCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_rule_validation_errors_total",
		Help: "Number of errors that occurred while validating a prometheusRules object",
	})

	r.MustRegister(
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		validationTriggeredCounter,
		validationErrorsCounter,
	)

	admit.RegisterMetrics(
		&validationTriggeredCounter,
		&validationErrorsCounter,
	)

	mux.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	ctx, cancel := context.WithCancel(context.Background())
	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error { return po.Run(ctx.Done()) })
	wg.Go(func() error { return ao.Run(ctx.Done()) })
	wg.Go(func() error { return to.Run(ctx.Done()) })

	srv := &http.Server{Handler: mux}
	wg.Go(serve(srv, l, logger))

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	select {
	case <-term:
		logger.Log("msg", "Received SIGTERM, exiting gracefully...")
	case <-ctx.Done():
	}

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log("msg", "Server shutdown error", "err", err)
	}

	cancel()
	if err := wg.Wait(); err != nil {
		logger.Log("msg", "Unhandled error received. Exiting...", "err", err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(Main())
}
