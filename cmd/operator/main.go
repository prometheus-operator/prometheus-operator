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

	rbacproxytls "github.com/brancz/kube-rbac-proxy/pkg/tls"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	klog "k8s.io/klog/v2"

	logging "github.com/prometheus-operator/prometheus-operator/internal/log"
	"github.com/prometheus-operator/prometheus-operator/pkg/admission"
	alertmanagercontroller "github.com/prometheus-operator/prometheus-operator/pkg/alertmanager"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prometheusagentcontroller "github.com/prometheus-operator/prometheus-operator/pkg/prometheus/agent"
	prometheuscontroller "github.com/prometheus-operator/prometheus-operator/pkg/prometheus/server"
	"github.com/prometheus-operator/prometheus-operator/pkg/server"
	thanoscontroller "github.com/prometheus-operator/prometheus-operator/pkg/thanos"
	"github.com/prometheus-operator/prometheus-operator/pkg/versionutil"
)

const (
	defaultOperatorTLSDir = "/etc/tls/private"
)

const (
	defaultReloaderCPU    = "10m"
	defaultReloaderMemory = "50Mi"
)

var (
	ns                   = namespaces{}
	deniedNs             = namespaces{}
	prometheusNs         = namespaces{}
	alertmanagerNs       = namespaces{}
	alertmanagerConfigNs = namespaces{}
	thanosRulerNs        = namespaces{}
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
	var ns = make([]string, 0, len(n))
	for k := range n {
		ns = append(ns, k)
	}
	return ns
}

// Helper function for checking CRD prerequisites
func checkPrerequisites(ctx context.Context, logger log.Logger, cc *k8sutil.CRDChecker, allowedNamespaces []string, verbs map[string][]string, groupVersion, resourceName string) (bool, error) {
	err := cc.CheckPrerequisites(ctx, allowedNamespaces, verbs, groupVersion, resourceName)
	switch {
	case errors.Is(err, k8sutil.ErrPrerequiresitesFailed):
		level.Warn(logger).Log("msg", fmt.Sprintf("%s CRD not supported", resourceName), "reason", err)
		return false, nil
	case err != nil:
		level.Error(logger).Log("msg", fmt.Sprintf("failed to check prerequisites for the %s CRD ", resourceName), "err", err)
		return false, err
	default:
		return true, nil
	}
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
	cfg = operator.Config{}

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
	flagset.StringVar(&cfg.KubeletSelector, "kubelet-selector", "", "Label selector to filter nodes.")
	flagset.BoolVar(&cfg.TLSInsecure, "tls-insecure", false, "- NOT RECOMMENDED FOR PRODUCTION - Don't verify API server's CA certificate.")
	// The Prometheus config reloader image is released along with the
	// Prometheus Operator image, tagged with the same semver version. Default to
	// the Prometheus Operator version if no Prometheus config reloader image is
	// specified.
	flagset.StringVar(&cfg.ReloaderConfig.Image, "prometheus-config-reloader", operator.DefaultPrometheusConfigReloaderImage, "Prometheus config reloader image")
	flagset.StringVar(&cfg.ReloaderConfig.CPURequest, "config-reloader-cpu-request", defaultReloaderCPU, "Config Reloader CPU request. Value \"0\" disables it and causes no request to be configured. Flag overrides `--config-reloader-cpu` value for the CPU request")
	flagset.StringVar(&cfg.ReloaderConfig.CPULimit, "config-reloader-cpu-limit", defaultReloaderCPU, "Config Reloader CPU limit. Value \"0\" disables it and causes no limit to be configured. Flag overrides `--config-reloader-cpu` for the CPU limit")
	flagset.StringVar(&cfg.ReloaderConfig.MemoryRequest, "config-reloader-memory-request", defaultReloaderMemory, "Config Reloader Memory request. Value \"0\" disables it and causes no request to be configured. Flag overrides `--config-reloader-memory` for the memory request")
	flagset.StringVar(&cfg.ReloaderConfig.MemoryLimit, "config-reloader-memory-limit", defaultReloaderMemory, "Config Reloader Memory limit. Value \"0\" disables it and causes no limit to be configured. Flag overrides `--config-reloader-memory` for the memory limit")
	flagset.BoolVar(&cfg.ReloaderConfig.EnableProbes, "enable-config-reloader-probes", false, "Enable liveness and readiness for the config-reloader container. Default: false")
	flagset.StringVar(&cfg.AlertmanagerDefaultBaseImage, "alertmanager-default-base-image", operator.DefaultAlertmanagerBaseImage, "Alertmanager default base image (path without tag/version)")
	flagset.StringVar(&cfg.PrometheusDefaultBaseImage, "prometheus-default-base-image", operator.DefaultPrometheusBaseImage, "Prometheus default base image (path without tag/version)")
	flagset.StringVar(&cfg.ThanosDefaultBaseImage, "thanos-default-base-image", operator.DefaultThanosBaseImage, "Thanos default base image (path without tag/version)")
	flagset.Var(ns, "namespaces", "Namespaces to scope the interaction of the Prometheus Operator and the apiserver (allow list). This is mutually exclusive with --deny-namespaces.")
	flagset.Var(deniedNs, "deny-namespaces", "Namespaces not to scope the interaction of the Prometheus Operator (deny list). This is mutually exclusive with --namespaces.")
	flagset.Var(prometheusNs, "prometheus-instance-namespaces", "Namespaces where Prometheus and PrometheusAgent custom resources and corresponding Secrets, Configmaps and StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for Prometheus custom resources.")
	flagset.Var(alertmanagerNs, "alertmanager-instance-namespaces", "Namespaces where Alertmanager custom resources and corresponding StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for Alertmanager custom resources.")
	flagset.Var(alertmanagerConfigNs, "alertmanager-config-namespaces", "Namespaces where AlertmanagerConfig custom resources and corresponding Secrets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for AlertmanagerConfig custom resources.")
	flagset.Var(thanosRulerNs, "thanos-ruler-instance-namespaces", "Namespaces where ThanosRuler custom resources and corresponding StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for ThanosRuler custom resources.")
	flagset.Var(&cfg.Annotations, "annotations", "Annotations to be add to all resources created by the operator")
	flagset.Var(&cfg.Labels, "labels", "Labels to be add to all resources created by the operator")
	flagset.StringVar(&cfg.LocalHost, "localhost", "localhost", "EXPERIMENTAL (could be removed in future releases) - Host used to communicate between local services on a pod. Fixes issues where localhost resolves incorrectly.")
	flagset.StringVar(&cfg.ClusterDomain, "cluster-domain", "", "The domain of the cluster. This is used to generate service FQDNs. If this is not specified, DNS search domain expansion is used instead.")
	flagset.StringVar(&cfg.LogLevel, "log-level", "info", fmt.Sprintf("Log level to use. Possible values: %s", strings.Join(logging.AvailableLogLevels, ", ")))
	flagset.StringVar(&cfg.LogFormat, "log-format", "logfmt", fmt.Sprintf("Log format to use. Possible values: %s", strings.Join(logging.AvailableLogFormats, ", ")))
	flagset.StringVar(&cfg.PromSelector, "prometheus-instance-selector", "", "Label selector to filter Prometheus and PrometheusAgent Custom Resources to watch.")
	flagset.StringVar(&cfg.AlertManagerSelector, "alertmanager-instance-selector", "", "Label selector to filter AlertManager Custom Resources to watch.")
	flagset.StringVar(&cfg.ThanosRulerSelector, "thanos-ruler-instance-selector", "", "Label selector to filter ThanosRuler Custom Resources to watch.")
	flagset.StringVar(&cfg.SecretListWatchSelector, "secret-field-selector", "", "Field selector to filter Secrets to watch")
}

func Main() int {
	versionutil.RegisterFlags()
	// No need to check for errors because Parse would exit on error.
	_ = flagset.Parse(os.Args[1:])

	if versionutil.ShouldPrintVersion() {
		versionutil.Print(os.Stdout, "prometheus-operator")
		return 0
	}

	logger, err := logging.NewLogger(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		stdlog.Fatal(err)
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

	// Above level 6, the k8s client would log bearer tokens in clear-text.
	klog.ClampLevel(6)
	klog.SetLogger(log.With(logger, "component", "k8s_client_runtime"))

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
	cfg.Namespaces.AlertmanagerConfigAllowList = alertmanagerConfigNs
	cfg.Namespaces.ThanosRulerAllowList = thanosRulerNs

	if len(cfg.Namespaces.PrometheusAllowList) == 0 {
		cfg.Namespaces.PrometheusAllowList = cfg.Namespaces.AllowList
	}

	if len(cfg.Namespaces.AlertmanagerAllowList) == 0 {
		cfg.Namespaces.AlertmanagerAllowList = cfg.Namespaces.AllowList
	}

	if len(cfg.Namespaces.AlertmanagerConfigAllowList) == 0 {
		cfg.Namespaces.AlertmanagerConfigAllowList = cfg.Namespaces.AllowList
	}

	if len(cfg.Namespaces.ThanosRulerAllowList) == 0 {
		cfg.Namespaces.ThanosRulerAllowList = cfg.Namespaces.AllowList
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg, ctx := errgroup.WithContext(ctx)
	r := prometheus.NewRegistry()

	k8sutil.MustRegisterClientGoMetrics(r)

	allowedNamespaces := namespaces(cfg.Namespaces.AllowList).asSlice()

	cc, err := k8sutil.NewCRDChecker(cfg.Host, cfg.TLSInsecure, &cfg.TLSConfig)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create new CRDChecker object ", "err", err)
		cancel()
		return 1
	}

	scrapeConfigSupported, err := checkPrerequisites(
		ctx,
		logger,
		cc,
		allowedNamespaces,
		map[string][]string{
			monitoringv1alpha1.ScrapeConfigName: {"get", "list", "watch"},
		},
		monitoringv1alpha1.SchemeGroupVersion.String(),
		monitoringv1alpha1.ScrapeConfigName,
	)

	if err != nil {
		cancel()
		return 1
	}

	po, err := prometheuscontroller.New(ctx, cfg, log.With(logger, "component", "prometheusoperator"), r, scrapeConfigSupported)
	if err != nil {
		fmt.Fprintln(os.Stderr, "instantiating prometheus controller failed: ", err)
		cancel()
		return 1
	}

	prometheusAgentSupported, err := checkPrerequisites(
		ctx,
		logger,
		cc,
		allowedNamespaces,
		map[string][]string{
			monitoringv1alpha1.PrometheusAgentName:                           {"get", "list", "watch"},
			fmt.Sprintf("%s/status", monitoringv1alpha1.PrometheusAgentName): {"update"},
		},
		monitoringv1alpha1.SchemeGroupVersion.String(),
		monitoringv1alpha1.PrometheusAgentName,
	)
	if err != nil {
		cancel()
		return 1
	}

	var pao *prometheusagentcontroller.Operator
	if prometheusAgentSupported {
		pao, err = prometheusagentcontroller.New(ctx, cfg, log.With(logger, "component", "prometheusagentoperator"), r, scrapeConfigSupported)
		if err != nil {
			level.Error(logger).Log("msg", "instantiating prometheus-agent controller failed", "err", err)
			cancel()
			return 1
		}
	}

	ao, err := alertmanagercontroller.New(ctx, cfg, log.With(logger, "component", "alertmanageroperator"), r)
	if err != nil {
		fmt.Fprintln(os.Stderr, "instantiating alertmanager controller failed: ", err)
		cancel()
		return 1
	}

	to, err := thanoscontroller.New(ctx, cfg, log.With(logger, "component", "thanosoperator"), r)
	if err != nil {
		fmt.Fprintln(os.Stderr, "instantiating thanos controller failed: ", err)
		cancel()
		return 1
	}

	mux := http.NewServeMux()

	admit := admission.New(log.With(logger, "component", "admissionwebhook"))
	admit.Register(mux)
	l, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		fmt.Fprintln(os.Stderr, "listening failed", cfg.ListenAddress, err)
		cancel()
		return 1
	}

	var tlsConfig *tls.Config
	if serverTLS {
		if rawTLSCipherSuites != "" {
			cfg.ServerTLSConfig.CipherSuites = strings.Split(rawTLSCipherSuites, ",")
		}
		tlsConfig, err = server.NewTLSConfig(logger, cfg.ServerTLSConfig.CertFile, cfg.ServerTLSConfig.KeyFile,
			cfg.ServerTLSConfig.ClientCAFile, cfg.ServerTLSConfig.MinVersion, cfg.ServerTLSConfig.CipherSuites)
		if tlsConfig == nil || err != nil {
			fmt.Fprintln(os.Stderr, "invalid TLS config", err)
			cancel()
			return 1
		}
	}

	validationTriggeredCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_rule_validation_triggered_total",
		Help: "DEPRECATED, removed in v0.57.0: Number of times a prometheusRule object triggered validation",
	})

	validationErrorsCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_rule_validation_errors_total",
		Help: "DEPRECATED, removed in v0.57.0: Number of errors that occurred while validating a prometheusRules object",
	})

	alertManagerConfigValidationTriggered := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_alertmanager_config_validation_triggered_total",
		Help: "DEPRECATED, removed in v0.57.0: Number of times an alertmanagerconfig object triggered validation",
	})

	alertManagerConfigValidationError := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_alertmanager_config_validation_errors_total",
		Help: "DEPRECATED, removed in v0.57.0: Number of errors that occurred while validating a alertmanagerconfig object",
	})

	r.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		validationTriggeredCounter,
		validationErrorsCounter,
		alertManagerConfigValidationTriggered,
		alertManagerConfigValidationError,
		version.NewCollector("prometheus_operator"),
	)

	admit.RegisterMetrics(
		validationTriggeredCounter,
		validationErrorsCounter,
		alertManagerConfigValidationTriggered,
		alertManagerConfigValidationError,
	)

	mux.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	wg.Go(func() error { return po.Run(ctx) })
	if pao != nil {
		wg.Go(func() error { return pao.Run(ctx) })
	}
	wg.Go(func() error { return ao.Run(ctx) })
	wg.Go(func() error { return to.Run(ctx) })

	if tlsConfig != nil {
		r, err := rbacproxytls.NewCertReloader(
			cfg.ServerTLSConfig.CertFile,
			cfg.ServerTLSConfig.KeyFile,
			cfg.ServerTLSConfig.ReloadInterval,
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to initialize certificate reloader", err)
			cancel()
			return 1
		}

		tlsConfig.GetCertificate = r.GetCertificate

		wg.Go(func() error {
			for {
				// r.Watch will wait ReloadInterval, so this is not
				// a hot loop
				if err := r.Watch(ctx); err != nil {
					level.Warn(logger).Log("msg", "error watching certificate reloader",
						"err", err)
				} else {
					return nil
				}
			}
		})
	}
	srv := &http.Server{
		Handler:           mux,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		// use flags on standard logger to align with base logger and get consistent parsed fields form adapter:
		// use shortfile flag to get proper 'caller' field (avoid being wrongly parsed/extracted from message)
		// and no datetime related flag to keep 'ts' field from base logger (with controlled format)
		ErrorLog: stdlog.New(log.NewStdlibAdapter(logger), "", stdlog.Lshortfile),
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
