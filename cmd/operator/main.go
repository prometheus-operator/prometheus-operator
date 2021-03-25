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
	"errors"
	"flag"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus-operator/prometheus-operator/pkg/admission"
	alertmanagercontroller "github.com/prometheus-operator/prometheus-operator/pkg/alertmanager"
	"github.com/prometheus-operator/prometheus-operator/pkg/api"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prometheuscontroller "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
	thanoscontroller "github.com/prometheus-operator/prometheus-operator/pkg/thanos"
	"github.com/prometheus-operator/prometheus-operator/pkg/versionutil"

	rbacproxytls "github.com/brancz/kube-rbac-proxy/pkg/tls"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	klog "k8s.io/klog"
	klogv2 "k8s.io/klog/v2"
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
	logFormatJSON   = "json"
)

const (
	defaultOperatorTLSDir = "/etc/tls/private"
)

const (
	defaultReloaderCPU    = "100m"
	defaultReloaderMemory = "50Mi"
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
	for _, ns := range strings.Split(value, ",") {
		n[ns] = struct{}{}
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
		level.Info(logger).Log("msg", "Starting insecure server on "+listener.Addr().String())
		if err := srv.Serve(listener); err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}

func serveTLS(srv *http.Server, listener net.Listener, logger log.Logger) func() error {
	return func() error {
		level.Info(logger).Log("msg", "Starting secure server on "+listener.Addr().String())
		if err := srv.ServeTLS(listener, "", ""); err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}

var (
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
		logFormatJSON,
	}
	cfg             = operator.Config{}
	rcCPU, rcMemory string

	rawTLSCipherSuites string
	serverTLS          bool

	flagset = flag.CommandLine
)

func init() {
	// With migration to klog-gokit, calling klogv2.InitFlags(flagset) is not applicable.
	flagset.StringVar(&cfg.ListenAddress, "web.listen-address", ":8080", "Address on which to expose metrics and web interface.")
	flagset.BoolVar(&serverTLS, "web.enable-tls", false, "Activate prometheus operator web server TLS.  "+
		" This is useful for example when using the rule validation webhook.")
	flagset.StringVar(&cfg.ServerTLSConfig.CertFile, "web.cert-file", defaultOperatorTLSDir+"/tls.crt", "Cert file to be used for operator web server endpoints.")
	flagset.StringVar(&cfg.ServerTLSConfig.KeyFile, "web.key-file", defaultOperatorTLSDir+"/tls.key", "Private key matching the cert file to be used for operator web server endpoints.")
	flagset.StringVar(&cfg.ServerTLSConfig.ClientCAFile, "web.client-ca-file", defaultOperatorTLSDir+"/tls-ca.crt", "Client CA certificate file to be used for operator web server endpoints.")
	flagset.DurationVar(&cfg.ServerTLSConfig.ReloadInterval, "web.tls-reload-interval", time.Minute, "The interval at which to watch for TLS certificate changes, by default set to 1 minute. (default 1m0s).")
	flagset.StringVar(&cfg.ServerTLSConfig.MinVersion, "web.tls-min-version", "VersionTLS13",
		"Minimum TLS version supported. Value must match version names from https://golang.org/pkg/crypto/tls/#pkg-constants.")
	flagset.StringVar(&rawTLSCipherSuites, "web.tls-cipher-suites", "", "Comma-separated list of cipher suites for the server."+
		" Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants)."+
		"If omitted, the default Go cipher suites will be used."+
		"Note that TLS 1.3 ciphersuites are not configurable.")
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
	flagset.StringVar(&cfg.ReloaderConfig.Image, "prometheus-config-reloader", operator.DefaultPrometheusConfigReloaderImage, "Prometheus config reloader image")
	// TODO(JJJJJones): remove the '--config-reloader-cpu' flag before releasing v0.48.
	flagset.StringVar(&rcCPU, "config-reloader-cpu", defaultReloaderCPU, "Config Reloader CPU request & limit. Value \"0\" disables it and causes no request/limit to be configured. Deprecated, it will be removed in v0.48.0.")
	flagset.StringVar(&cfg.ReloaderConfig.CPURequest, "config-reloader-cpu-request", "", "Config Reloader CPU request. Value \"0\" disables it and causes no request to be configured. Flag overrides `--config-reloader-cpu` value for the CPU request")
	flagset.StringVar(&cfg.ReloaderConfig.CPULimit, "config-reloader-cpu-limit", "", "Config Reloader CPU limit. Value \"0\" disables it and causes no limit to be configured. Flag overrides `--config-reloader-cpu` for the CPU limit")
	// TODO(JJJJJones): remove the '--config-reloader-memory' flag before releasing v0.48.
	flagset.StringVar(&rcMemory, "config-reloader-memory", defaultReloaderMemory, "Config Reloader Memory request & limit. Value \"0\" disables it and causes no request/limit to be configured. Deprecated, it will be removed in v0.48.0.")
	flagset.StringVar(&cfg.ReloaderConfig.MemoryRequest, "config-reloader-memory-request", "", "Config Reloader Memory request. Value \"0\" disables it and causes no request to be configured. Flag overrides `--config-reloader-memory` for the memory request")
	flagset.StringVar(&cfg.ReloaderConfig.MemoryLimit, "config-reloader-memory-limit", "", "Config Reloader Memory limit. Value \"0\" disables it and causes no limit to be configured. Flag overrides `--config-reloader-memory` for the memory limit")
	flagset.StringVar(&cfg.AlertmanagerDefaultBaseImage, "alertmanager-default-base-image", operator.DefaultAlertmanagerBaseImage, "Alertmanager default base image (path without tag/version)")
	flagset.StringVar(&cfg.PrometheusDefaultBaseImage, "prometheus-default-base-image", operator.DefaultPrometheusBaseImage, "Prometheus default base image (path without tag/version)")
	flagset.StringVar(&cfg.ThanosDefaultBaseImage, "thanos-default-base-image", operator.DefaultThanosBaseImage, "Thanos default base image (path without tag/version)")
	flagset.Var(ns, "namespaces", "Namespaces to scope the interaction of the Prometheus Operator and the apiserver (allow list). This is mutually exclusive with --deny-namespaces.")
	flagset.Var(deniedNs, "deny-namespaces", "Namespaces not to scope the interaction of the Prometheus Operator (deny list). This is mutually exclusive with --namespaces.")
	flagset.Var(prometheusNs, "prometheus-instance-namespaces", "Namespaces where Prometheus custom resources and corresponding Secrets, Configmaps and StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for Prometheus custom resources.")
	flagset.Var(alertmanagerNs, "alertmanager-instance-namespaces", "Namespaces where Alertmanager custom resources and corresponding StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for Alertmanager custom resources.")
	flagset.Var(thanosRulerNs, "thanos-ruler-instance-namespaces", "Namespaces where ThanosRuler custom resources and corresponding StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for ThanosRuler custom resources.")
	flagset.Var(&cfg.Labels, "labels", "Labels to be add to all resources created by the operator")
	flagset.StringVar(&cfg.LocalHost, "localhost", "localhost", "EXPERIMENTAL (could be removed in future releases) - Host used to communicate between local services on a pod. Fixes issues where localhost resolves incorrectly.")
	flagset.StringVar(&cfg.ClusterDomain, "cluster-domain", "", "The domain of the cluster. This is used to generate service FQDNs. If this is not specified, DNS search domain expansion is used instead.")
	flagset.StringVar(&cfg.LogLevel, "log-level", logLevelInfo, fmt.Sprintf("Log level to use. Possible values: %s", strings.Join(availableLogLevels, ", ")))
	flagset.StringVar(&cfg.LogFormat, "log-format", logFormatLogfmt, fmt.Sprintf("Log format to use. Possible values: %s", strings.Join(availableLogFormats, ", ")))
	flagset.StringVar(&cfg.PromSelector, "prometheus-instance-selector", "", "Label selector to filter Prometheus Custom Resources to watch.")
	flagset.StringVar(&cfg.AlertManagerSelector, "alertmanager-instance-selector", "", "Label selector to filter AlertManager Custom Resources to watch.")
	flagset.StringVar(&cfg.ThanosRulerSelector, "thanos-ruler-instance-selector", "", "Label selector to filter ThanosRuler Custom Resources to watch.")
	flagset.StringVar(&cfg.SecretListWatchSelector, "secret-field-selector", "", "Field selector to filter Secrets to watch")
}

func Main() int {
	versionutil.RegisterFlags()
	flagset.Parse(os.Args[1:])

	if versionutil.ShouldPrintVersion() {
		versionutil.Print(os.Stdout, "prometheus-operator")
		return 0
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	if cfg.LogFormat == logFormatJSON {
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

	// Prioritize specific flags over generic ones
	// If rcCPU has been set but not config.CPURequest/config.CPULimit, default "" to rcCPU
	// If rcMemory has been set but not config.MemoryRequest or config.MemoryLimit, default "" to rcMemory
	if rcCPU != "" {
		if cfg.ReloaderConfig.CPURequest == "" {
			cfg.ReloaderConfig.CPURequest = rcCPU
		}
		if cfg.ReloaderConfig.CPULimit == "" {
			cfg.ReloaderConfig.CPULimit = rcCPU
		}
	}
	if rcMemory != "" {
		if cfg.ReloaderConfig.MemoryRequest == "" {
			cfg.ReloaderConfig.MemoryRequest = rcMemory
		}
		if cfg.ReloaderConfig.MemoryLimit == "" {
			cfg.ReloaderConfig.MemoryLimit = rcMemory
		}
	}
	// Check validity of reloader resource values given to flags
	_, err1 := resource.ParseQuantity(cfg.ReloaderConfig.CPULimit)
	if err1 != nil {
		fmt.Fprintf(os.Stderr, "The CPU limit specified for reloader \"%v\" is not a valid quantity! %v\n", cfg.ReloaderConfig.CPULimit, err1)
	}
	_, err2 := resource.ParseQuantity(cfg.ReloaderConfig.CPURequest)
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "The CPU request specified for reloader \"%v\" is not a valid quantity! %v\n", cfg.ReloaderConfig.CPURequest, err2)
	}
	_, err3 := resource.ParseQuantity(cfg.ReloaderConfig.MemoryLimit)
	if err3 != nil {
		fmt.Fprintf(os.Stderr, "The Memory limit specified for reloader \"%v\" is not a valid quantity! %v\n", cfg.ReloaderConfig.MemoryLimit, err3)
	}
	_, err4 := resource.ParseQuantity(cfg.ReloaderConfig.MemoryRequest)
	if err4 != nil {
		fmt.Fprintf(os.Stderr, "The Memory request specified for reloader \"%v\" is not a valid quantity! %v\n", cfg.ReloaderConfig.MemoryRequest, err4)
	}
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return 1
	}
	// Warn users that could be utilizing --config-reloader-cpu and --config-reloader-memory of their deprecated nature
	if rcCPU != defaultReloaderCPU {
		level.Warn(logger).Log("msg", "The '--config-reloader-cpu' flag has been deprecated! Please use '--config-reloader-cpu-limit' and '--config-reloader-cpu-request' according to your use-case.")
	}
	if rcMemory != defaultReloaderMemory {
		level.Warn(logger).Log("msg", "The '--config-reloader-memory' flag has been deprecated! Please use '--config-reloader-memory-limit' and '--config-reloader-memory-request' according to your use-case.")
	}

	// Above level 6, the k8s client would log bearer tokens in clear-text.
	klog.ClampLevel(6)
	klog.SetLogger(log.With(logger, "component", "k8s_client_runtime"))
	klogv2.ClampLevel(6)
	klogv2.SetLogger(log.With(logger, "component", "k8s_client_runtime"))

	level.Info(logger).Log("msg", "Starting Prometheus Operator", "version", version.Info())
	level.Info(logger).Log("build_context", version.BuildContext())

	if len(ns) > 0 && len(deniedNs) > 0 {
		fmt.Fprint(os.Stderr, "--namespaces and --deny-namespaces are mutually exclusive. Please provide only one of them.\n")
		return 1
	}

	cfg.Namespaces.AllowList = ns
	if len(cfg.Namespaces.AllowList) == 0 {
		cfg.Namespaces.AllowList[v1.NamespaceAll] = struct{}{}
	}

	cfg.Namespaces.DenyList = deniedNs
	cfg.Namespaces.PrometheusAllowList = prometheusNs
	cfg.Namespaces.AlertmanagerAllowList = alertmanagerNs
	cfg.Namespaces.ThanosRulerAllowList = thanosRulerNs

	if len(cfg.Namespaces.PrometheusAllowList) == 0 {
		cfg.Namespaces.PrometheusAllowList = cfg.Namespaces.AllowList
	}

	if len(cfg.Namespaces.AlertmanagerAllowList) == 0 {
		cfg.Namespaces.AlertmanagerAllowList = cfg.Namespaces.AllowList
	}

	if len(cfg.Namespaces.ThanosRulerAllowList) == 0 {
		cfg.Namespaces.ThanosRulerAllowList = cfg.Namespaces.AllowList
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg, ctx := errgroup.WithContext(ctx)
	r := prometheus.NewRegistry()

	k8sutil.MustRegisterClientGoMetrics(r)

	po, err := prometheuscontroller.New(ctx, cfg, log.With(logger, "component", "prometheusoperator"), r)
	if err != nil {
		fmt.Fprint(os.Stderr, "instantiating prometheus controller failed: ", err)
		cancel()
		return 1
	}

	ao, err := alertmanagercontroller.New(ctx, cfg, log.With(logger, "component", "alertmanageroperator"), r)
	if err != nil {
		fmt.Fprint(os.Stderr, "instantiating alertmanager controller failed: ", err)
		cancel()
		return 1
	}

	to, err := thanoscontroller.New(ctx, cfg, log.With(logger, "component", "thanosoperator"), r)
	if err != nil {
		fmt.Fprint(os.Stderr, "instantiating thanos controller failed: ", err)
		cancel()
		return 1
	}

	mux := http.NewServeMux()
	web, err := api.New(cfg, log.With(logger, "component", "api"))
	if err != nil {
		fmt.Fprint(os.Stderr, "instantiating api failed: ", err)
		cancel()
		return 1
	}
	admit := admission.New(log.With(logger, "component", "admissionwebhook"))

	web.Register(mux)
	admit.Register(mux)
	l, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		fmt.Fprint(os.Stderr, "listening failed", cfg.ListenAddress, err)
		cancel()
		return 1
	}

	var tlsConfig *tls.Config = nil
	if serverTLS {
		if rawTLSCipherSuites != "" {
			cfg.ServerTLSConfig.CipherSuites = strings.Split(rawTLSCipherSuites, ",")
		}
		tlsConfig, err = operator.NewTLSConfig(logger, cfg.ServerTLSConfig.CertFile, cfg.ServerTLSConfig.KeyFile,
			cfg.ServerTLSConfig.ClientCAFile, cfg.ServerTLSConfig.MinVersion, cfg.ServerTLSConfig.CipherSuites)
		if tlsConfig == nil || err != nil {
			fmt.Fprint(os.Stderr, "invalid TLS config", err)
			cancel()
			return 1
		}
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
		version.NewCollector("prometheus_operator"),
	)

	admit.RegisterMetrics(
		validationTriggeredCounter,
		validationErrorsCounter,
	)

	mux.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	wg.Go(func() error { return po.Run(ctx) })
	wg.Go(func() error { return ao.Run(ctx) })
	wg.Go(func() error { return to.Run(ctx) })

	if tlsConfig != nil {
		r, err := rbacproxytls.NewCertReloader(
			cfg.ServerTLSConfig.CertFile,
			cfg.ServerTLSConfig.KeyFile,
			cfg.ServerTLSConfig.ReloadInterval,
		)
		if err != nil {
			fmt.Fprint(os.Stderr, "failed to initialize certificate reloader", err)
			cancel()
			return 1
		}

		tlsConfig.GetCertificate = r.GetCertificate

		wg.Go(func() error {
			t := time.NewTicker(cfg.ServerTLSConfig.ReloadInterval)
			for {
				select {
				case <-t.C:
				case <-ctx.Done():
					return nil
				}
				if err := r.Watch(ctx); err != nil {
					level.Warn(logger).Log("msg", "error reloading server TLS certificate",
						"err", err)
				} else {
					return nil
				}
			}
		})
	}
	srv := &http.Server{
		Handler:   mux,
		TLSConfig: tlsConfig,
		ErrorLog:  stdlog.New(log.NewStdlibAdapter(logger), "", stdlog.LstdFlags),
	}
	if srv.TLSConfig == nil {
		wg.Go(serve(srv, l, logger))
	} else {
		wg.Go(serveTLS(srv, l, logger))
	}

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	select {
	case <-term:
		level.Info(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
	case <-ctx.Done():
	}

	if err := srv.Shutdown(ctx); err != nil {
		level.Warn(logger).Log("msg", "Server shutdown error", "err", err)
	}

	cancel()
	if err := wg.Wait(); err != nil {
		level.Warn(logger).Log("msg", "Unhandled error received. Exiting...", "err", err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(Main())
}
