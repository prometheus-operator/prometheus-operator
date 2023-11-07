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
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	logging "github.com/prometheus-operator/prometheus-operator/internal/log"
	"github.com/prometheus-operator/prometheus-operator/pkg/admission"
	alertmanagercontroller "github.com/prometheus-operator/prometheus-operator/pkg/alertmanager"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
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

var (
	ns                   = namespaces{}
	deniedNs             = namespaces{}
	prometheusNs         = namespaces{}
	alertmanagerNs       = namespaces{}
	alertmanagerConfigNs = namespaces{}
	thanosRulerNs        = namespaces{}
)

type namespaces map[string]struct{}

// Set implements the flag.Value interface.
func (n namespaces) Set(value string) error {
	if n == nil {
		return errors.New("expected n of type namespaces to be initialized")
	}
	for _, ns := range strings.Split(value, ",") {
		n[ns] = struct{}{}
	}
	return nil
}

// String implements the flag.Value interface.
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

// checkPrerequisites verifies that the CRD is installed in the cluster and
// that the operator has enough permissions to manage the resource.
func checkPrerequisites(
	ctx context.Context,
	logger log.Logger,
	kclient kubernetes.Interface,
	allowedNamespaces []string,
	groupVersion schema.GroupVersion,
	resource string,
	attributes ...k8sutil.ResourceAttribute,
) (bool, error) {
	installed, err := k8sutil.IsAPIGroupVersionResourceSupported(kclient.Discovery(), groupVersion, resource)
	if err != nil {
		return false, fmt.Errorf("failed to check presence of resource %q (group %q): %w", resource, groupVersion, err)
	}

	if !installed {
		level.Warn(logger).Log("msg", fmt.Sprintf("resource %q (group: %q) not installed in the cluster", resource, groupVersion))
		return false, nil
	}

	allowed, errs, err := k8sutil.IsAllowed(ctx, kclient.AuthorizationV1().SelfSubjectAccessReviews(), allowedNamespaces, attributes...)
	if err != nil {
		return false, fmt.Errorf("failed to check permissions on resource %q (group %q): %w", resource, groupVersion, err)
	}

	if !allowed {
		for _, reason := range errs {
			level.Warn(logger).Log("msg", fmt.Sprintf("missing permission on resource %q (group: %q)", resource, groupVersion), "reason", reason)
		}
		return false, nil
	}

	return true, nil
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
		level.Info(logger).Log("msg", "Starting secure server on "+listener.Addr().String(), "http2", enableHTTP2)
		if err := srv.ServeTLS(listener, "", ""); err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}

const (
	defaultReloaderCPU    = "10m"
	defaultReloaderMemory = "50Mi"
)

var (
	cfg = operator.DefaultConfig(defaultReloaderCPU, defaultReloaderMemory)

	rawTLSCipherSuites string
	enableHTTP2        bool
	serverTLS          bool

	flagset = flag.CommandLine
)

func init() {
	flagset.StringVar(&cfg.ListenAddress, "web.listen-address", ":8080", "Address on which to expose metrics and web interface.")
	// Mitigate CVE-2023-44487 by disabling HTTP2 by default until the Go
	// standard library and golang.org/x/net are fully fixed.
	// Right now, it is possible for authenticated and unauthenticated users to
	// hold open HTTP2 connections and consume huge amounts of memory.
	// See:
	// * https://github.com/kubernetes/kubernetes/pull/121120
	// * https://github.com/kubernetes/kubernetes/issues/121197
	// * https://github.com/golang/go/issues/63417#issuecomment-1758858612
	flagset.BoolVar(&enableHTTP2, "web.enable-http2", false, "Enable HTTP2 connections.")
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

	flagset.StringVar(&cfg.ImpersonateUser, "as", "", "Username to impersonate. User could be a regular user or a service account in a namespace.")
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
	flagset.Var(&cfg.ReloaderConfig.CPURequests, "config-reloader-cpu-request", "Config Reloader CPU requests. Value \"0\" disables it and causes no request to be configured.")
	flagset.Var(&cfg.ReloaderConfig.CPULimits, "config-reloader-cpu-limit", "Config Reloader CPU limits. Value \"0\" disables it and causes no limit to be configured.")
	flagset.Var(&cfg.ReloaderConfig.MemoryRequests, "config-reloader-memory-request", "Config Reloader memory requests. Value \"0\" disables it and causes no request to be configured.")
	flagset.Var(&cfg.ReloaderConfig.MemoryLimits, "config-reloader-memory-limit", "Config Reloader memory limits. Value \"0\" disables it and causes no limit to be configured.")
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

func run() int {
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

	level.Info(logger).Log("msg", "Starting Prometheus Operator", "version", version.Info())
	level.Info(logger).Log("build_context", version.BuildContext())

	if len(ns) > 0 && len(deniedNs) > 0 {
		level.Error(logger).Log(
			"msg", "--namespaces and --deny-namespaces are mutually exclusive, only one should be provided",
			"namespaces", ns,
			"deny_namespaces", deniedNs,
		)
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

	restConfig, err := k8sutil.NewClusterConfig(cfg.Host, cfg.TLSInsecure, &cfg.TLSConfig, cfg.ImpersonateUser)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create Kubernetes client configuration", "err", err)
		cancel()
		return 1
	}

	kclient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create Kubernetes client", "err", err)
		cancel()
		return 1
	}

	kubernetesVersion, err := kclient.Discovery().ServerVersion()
	if err != nil {
		level.Error(logger).Log("msg", "failed to request Kubernetes server version", "err", err)
		cancel()
		return 1
	}
	cfg.KubernetesVersion = *kubernetesVersion
	level.Info(logger).Log("msg", "connection established", "cluster-version", cfg.KubernetesVersion)
	// Check if we can read the storage classs
	canReadStorageClass, err := checkPrerequisites(
		ctx,
		logger,
		kclient,
		nil,
		storagev1.SchemeGroupVersion,
		storagev1.SchemeGroupVersion.WithResource("storageclasses").Resource,
		k8sutil.ResourceAttribute{
			Group:    storagev1.GroupName,
			Version:  storagev1.SchemeGroupVersion.Version,
			Resource: storagev1.SchemeGroupVersion.WithResource("storageclasses").Resource,
			Verbs:    []string{"get"},
		},
	)

	if err != nil {
		level.Error(logger).Log("msg", "failed to check StorageClass support", "err", err)
		cancel()
		return 1
	}

	scrapeConfigSupported, err := checkPrerequisites(
		ctx,
		logger,
		kclient,
		namespaces(cfg.Namespaces.AllowList).asSlice(),
		monitoringv1alpha1.SchemeGroupVersion,
		monitoringv1alpha1.ScrapeConfigName,
		k8sutil.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1alpha1.Version,
			Resource: monitoringv1alpha1.ScrapeConfigName,
			Verbs:    []string{"get", "list", "watch"},
		},
	)
	if err != nil {
		level.Error(logger).Log("msg", "failed to check ScrapeConfig support", "err", err)
		cancel()
		return 1
	}

	po, err := prometheuscontroller.New(ctx, restConfig, cfg, log.With(logger, "component", "prometheusoperator"), r, scrapeConfigSupported, canReadStorageClass)
	if err != nil {
		level.Error(logger).Log("msg", "instantiating prometheus controller failed", "err", err)
		cancel()
		return 1
	}

	prometheusAgentSupported, err := checkPrerequisites(
		ctx,
		logger,
		kclient,
		namespaces(cfg.Namespaces.PrometheusAllowList).asSlice(),
		monitoringv1alpha1.SchemeGroupVersion,
		monitoringv1alpha1.PrometheusAgentName,
		k8sutil.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1alpha1.Version,
			Resource: monitoringv1alpha1.PrometheusAgentName,
			Verbs:    []string{"get", "list", "watch"},
		},
		k8sutil.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1alpha1.Version,
			Resource: fmt.Sprintf("%s/status", monitoringv1alpha1.PrometheusAgentName),
			Verbs:    []string{"update"},
		},
	)
	if err != nil {
		level.Error(logger).Log("msg", "failed to check PrometheusAgent support", "err", err)
		cancel()
		return 1
	}

	var pao *prometheusagentcontroller.Operator
	if prometheusAgentSupported {
		pao, err = prometheusagentcontroller.New(ctx, restConfig, cfg, log.With(logger, "component", "prometheusagentoperator"), r, scrapeConfigSupported, canReadStorageClass)
		if err != nil {
			level.Error(logger).Log("msg", "instantiating prometheus-agent controller failed", "err", err)
			cancel()
			return 1
		}
	}

	ao, err := alertmanagercontroller.New(ctx, restConfig, cfg, log.With(logger, "component", "alertmanageroperator"), r, canReadStorageClass)
	if err != nil {
		level.Error(logger).Log("msg", "instantiating alertmanager controller failed", "err", err)
		cancel()
		return 1
	}

	to, err := thanoscontroller.New(ctx, restConfig, cfg, log.With(logger, "component", "thanosoperator"), r, canReadStorageClass)
	if err != nil {
		level.Error(logger).Log("msg", "instantiating thanos controller failed", "err", err)
		cancel()
		return 1
	}

	mux := http.NewServeMux()

	admit := admission.New(log.With(logger, "component", "admissionwebhook"))
	admit.Register(mux)
	l, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		level.Error(logger).Log("msg", "listening failed", "address", cfg.ListenAddress, "err", err)
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
			level.Error(logger).Log("msg", "invalid TLS config", "err", err)
			cancel()
			return 1
		}
	}

	r.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		version.NewCollector("prometheus_operator"),
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
			level.Error(logger).Log("msg", "failed to initialize certificate reloader", "err", err)
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
	if !enableHTTP2 {
		srv.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
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
	os.Exit(run())
}
