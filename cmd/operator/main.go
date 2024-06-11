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
	"flag"
	"fmt"
	stdlog "log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	k8sflag "k8s.io/component-base/cli/flag"
	"k8s.io/utils/ptr"

	"github.com/prometheus-operator/prometheus-operator/internal/goruntime"
	logging "github.com/prometheus-operator/prometheus-operator/internal/log"
	"github.com/prometheus-operator/prometheus-operator/pkg/admission"
	alertmanagercontroller "github.com/prometheus-operator/prometheus-operator/pkg/alertmanager"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/kubelet"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prometheusagentcontroller "github.com/prometheus-operator/prometheus-operator/pkg/prometheus/agent"
	prometheuscontroller "github.com/prometheus-operator/prometheus-operator/pkg/prometheus/server"
	"github.com/prometheus-operator/prometheus-operator/pkg/server"
	thanoscontroller "github.com/prometheus-operator/prometheus-operator/pkg/thanos"
	"github.com/prometheus-operator/prometheus-operator/pkg/versionutil"
)

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

const (
	defaultReloaderCPU    = "10m"
	defaultReloaderMemory = "50Mi"

	defaultMemlimitRatio = 0.0
)

var (
	cfg = operator.DefaultConfig(defaultReloaderCPU, defaultReloaderMemory)

	logConfig logging.Config

	impersonateUser string
	apiServer       string
	tlsClientConfig rest.TLSClientConfig

	memlimitRatio float64

	serverConfig = server.DefaultConfig(":8080", false)

	// Parameters for the kubelet endpoints controller.
	kubeletObject       string
	kubeletSelector     operator.LabelSelector
	nodeAddressPriority operator.NodeAddressPriority

	featureGates = k8sflag.NewMapStringBool(ptr.To(map[string]bool{}))
)

func parseFlags(fs *flag.FlagSet) {
	// Web server settings.
	server.RegisterFlags(fs, &serverConfig)

	// Kubernetes client-go settings.
	fs.StringVar(&impersonateUser, "as", "", "Username to impersonate. User could be a regular user or a service account in a namespace.")
	fs.StringVar(&apiServer, "apiserver", "", "API Server addr, e.g. ' - NOT RECOMMENDED FOR PRODUCTION - http://127.0.0.1:8080'. Omit parameter to run in on-cluster mode and utilize the service account token.")
	fs.StringVar(&tlsClientConfig.CertFile, "cert-file", "", " - NOT RECOMMENDED FOR PRODUCTION - Path to public TLS certificate file.")
	fs.StringVar(&tlsClientConfig.KeyFile, "key-file", "", "- NOT RECOMMENDED FOR PRODUCTION - Path to private TLS certificate file.")
	fs.StringVar(&tlsClientConfig.CAFile, "ca-file", "", "- NOT RECOMMENDED FOR PRODUCTION - Path to TLS CA file.")
	fs.BoolVar(&tlsClientConfig.Insecure, "tls-insecure", false, "- NOT RECOMMENDED FOR PRODUCTION - Don't verify API server's CA certificate.")

	fs.StringVar(&kubeletObject, "kubelet-service", "", "Service/Endpoints object to write kubelets into in format \"namespace/name\"")
	fs.Var(&kubeletSelector, "kubelet-selector", "Label selector to filter nodes.")
	fs.Var(&nodeAddressPriority, "kubelet-node-address-priority", "Node address priority used by kubelet. Either 'internal' or 'external'. Default: 'internal'.")

	// The Prometheus config reloader image is released along with the
	// Prometheus Operator image, tagged with the same semver version. Default to
	// the Prometheus Operator version if no Prometheus config reloader image is
	// specified.
	fs.StringVar(&cfg.ReloaderConfig.Image, "prometheus-config-reloader", operator.DefaultPrometheusConfigReloaderImage, "Prometheus config reloader image")
	fs.Var(&cfg.ReloaderConfig.CPURequests, "config-reloader-cpu-request", "Config Reloader CPU requests. Value \"0\" disables it and causes no request to be configured.")
	fs.Var(&cfg.ReloaderConfig.CPULimits, "config-reloader-cpu-limit", "Config Reloader CPU limits. Value \"0\" disables it and causes no limit to be configured.")
	fs.Var(&cfg.ReloaderConfig.MemoryRequests, "config-reloader-memory-request", "Config Reloader memory requests. Value \"0\" disables it and causes no request to be configured.")
	fs.Var(&cfg.ReloaderConfig.MemoryLimits, "config-reloader-memory-limit", "Config Reloader memory limits. Value \"0\" disables it and causes no limit to be configured.")
	fs.BoolVar(&cfg.ReloaderConfig.EnableProbes, "enable-config-reloader-probes", false, "Enable liveness and readiness for the config-reloader container. Default: false")

	fs.StringVar(&cfg.AlertmanagerDefaultBaseImage, "alertmanager-default-base-image", operator.DefaultAlertmanagerBaseImage, "Alertmanager default base image (path without tag/version)")
	fs.StringVar(&cfg.PrometheusDefaultBaseImage, "prometheus-default-base-image", operator.DefaultPrometheusBaseImage, "Prometheus default base image (path without tag/version)")
	fs.StringVar(&cfg.ThanosDefaultBaseImage, "thanos-default-base-image", operator.DefaultThanosBaseImage, "Thanos default base image (path without tag/version)")
	fs.StringVar(&cfg.ControllerID, "controller-id", "", "Value used by the operator to filter Alertmanager, Prometheus, PrometheusAgent and ThanosRuler objects that it should reconcile. If the value isn't empty, the operator only reconciles objects with an `operator.prometheus.io/controller-id` annotation of the same value. Otherwise the operator reconciles all objects without the annotation or with an empty annotation value.")

	fs.Var(cfg.Namespaces.AllowList, "namespaces", "Namespaces to scope the interaction of the Prometheus Operator and the apiserver (allow list). This is mutually exclusive with --deny-namespaces.")
	fs.Var(cfg.Namespaces.DenyList, "deny-namespaces", "Namespaces not to scope the interaction of the Prometheus Operator (deny list). This is mutually exclusive with --namespaces.")
	fs.Var(cfg.Namespaces.PrometheusAllowList, "prometheus-instance-namespaces", "Namespaces where Prometheus and PrometheusAgent custom resources and corresponding Secrets, Configmaps and StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for Prometheus custom resources.")
	fs.Var(cfg.Namespaces.AlertmanagerAllowList, "alertmanager-instance-namespaces", "Namespaces where Alertmanager custom resources and corresponding StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for Alertmanager custom resources.")
	fs.Var(cfg.Namespaces.AlertmanagerConfigAllowList, "alertmanager-config-namespaces", "Namespaces where AlertmanagerConfig custom resources and corresponding Secrets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for AlertmanagerConfig custom resources.")
	fs.Var(cfg.Namespaces.ThanosRulerAllowList, "thanos-ruler-instance-namespaces", "Namespaces where ThanosRuler custom resources and corresponding StatefulSets are watched/created. If set this takes precedence over --namespaces or --deny-namespaces for ThanosRuler custom resources.")

	fs.Var(&cfg.Annotations, "annotations", "Annotations to be add to all resources created by the operator")
	fs.Var(&cfg.Labels, "labels", "Labels to be add to all resources created by the operator")

	fs.StringVar(&cfg.LocalHost, "localhost", "localhost", "EXPERIMENTAL (could be removed in future releases) - Host used to communicate between local services on a pod. Fixes issues where localhost resolves incorrectly.")
	fs.StringVar(&cfg.ClusterDomain, "cluster-domain", "", "The domain of the cluster. This is used to generate service FQDNs. If this is not specified, DNS search domain expansion is used instead.")

	fs.Var(&cfg.PromSelector, "prometheus-instance-selector", "Label selector to filter Prometheus and PrometheusAgent Custom Resources to watch.")
	fs.Var(&cfg.AlertmanagerSelector, "alertmanager-instance-selector", "Label selector to filter Alertmanager Custom Resources to watch.")
	fs.Var(&cfg.ThanosRulerSelector, "thanos-ruler-instance-selector", "Label selector to filter ThanosRuler Custom Resources to watch.")
	fs.Var(&cfg.SecretListWatchSelector, "secret-field-selector", "Field selector to filter Secrets to watch")

	fs.Float64Var(&memlimitRatio, "auto-gomemlimit-ratio", defaultMemlimitRatio, "The ratio of reserved GOMEMLIMIT memory to the detected maximum container or system memory. The value should be greater than 0.0 and less than 1.0. Default: 0.0 (disabled).")

	cfg.RegisterFeatureGatesFlags(fs, featureGates)

	logging.RegisterFlags(fs, &logConfig)
	versionutil.RegisterFlags(fs)

	// No need to check for errors because Parse would exit on error.
	_ = fs.Parse(os.Args[1:])
}

func run(fs *flag.FlagSet) int {
	parseFlags(fs)

	if versionutil.ShouldPrintVersion() {
		versionutil.Print(os.Stdout, "prometheus-operator")
		return 0
	}

	logger, err := logging.NewLogger(logConfig)
	if err != nil {
		stdlog.Fatal(err)
	}

	if err := cfg.Gates.UpdateFeatureGates(*featureGates.Map); err != nil {
		level.Error(logger).Log("error", err)
		return 1
	}

	level.Info(logger).Log("msg", "Starting Prometheus Operator", "version", version.Info())
	level.Info(logger).Log("build_context", version.BuildContext())
	level.Info(logger).Log("feature_gates", cfg.Gates.String())
	goruntime.SetMaxProcs(logger)
	goruntime.SetMemLimit(logger, memlimitRatio)

	if len(cfg.Namespaces.AllowList) > 0 && len(cfg.Namespaces.DenyList) > 0 {
		level.Error(logger).Log(
			"msg", "--namespaces and --deny-namespaces are mutually exclusive, only one should be provided",
			"namespaces", cfg.Namespaces.AllowList,
			"deny_namespaces", cfg.Namespaces.DenyList,
		)
		return 1
	}
	cfg.Namespaces.Finalize()
	level.Info(logger).Log("msg", "namespaces filtering configuration ", "config", cfg.Namespaces.String())

	ctx, cancel := context.WithCancel(context.Background())
	wg, ctx := errgroup.WithContext(ctx)
	r := prometheus.NewRegistry()

	k8sutil.MustRegisterClientGoMetrics(r)

	restConfig, err := k8sutil.NewClusterConfig(k8sutil.ClusterConfig{
		Host:      apiServer,
		TLSConfig: tlsClientConfig,
		AsUser:    impersonateUser,
	})

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

	promControllerOptions := make([]prometheuscontroller.ControllerOptions, 0)
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
	if canReadStorageClass {
		promControllerOptions = append(promControllerOptions, prometheuscontroller.WithStorageClassValidation())
	}

	canEmitEvents, reasons, err := k8sutil.IsAllowed(ctx, kclient.AuthorizationV1().SelfSubjectAccessReviews(), nil,
		k8sutil.ResourceAttribute{
			Group:    corev1.GroupName,
			Version:  corev1.SchemeGroupVersion.Version,
			Resource: corev1.SchemeGroupVersion.WithResource("events").Resource,
			Verbs:    []string{"create", "patch"},
		})
	if err != nil {
		level.Error(logger).Log("msg", "failed to check Events support", "err", err)
		cancel()
		return 1
	}

	if !canEmitEvents {
		for _, reason := range reasons {
			level.Warn(logger).Log("msg", "missing permission to emit events", "reason", reason)
		}
	}
	eventRecorderFactory := operator.NewEventRecorderFactory(canEmitEvents)

	scrapeConfigSupported, err := checkPrerequisites(
		ctx,
		logger,
		kclient,
		cfg.Namespaces.AllowList.Slice(),
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
	if scrapeConfigSupported {
		promControllerOptions = append(promControllerOptions, prometheuscontroller.WithScrapeConfig())
	}

	prometheusSupported, err := checkPrerequisites(
		ctx,
		logger,
		kclient,
		cfg.Namespaces.PrometheusAllowList.Slice(),
		monitoringv1.SchemeGroupVersion,
		monitoringv1.PrometheusName,
		k8sutil.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1.Version,
			Resource: monitoringv1.PrometheusName,
			Verbs:    []string{"get", "list", "watch"},
		},
		k8sutil.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1.Version,
			Resource: fmt.Sprintf("%s/status", monitoringv1.PrometheusName),
			Verbs:    []string{"update"},
		},
	)
	if err != nil {
		level.Error(logger).Log("msg", "failed to check Prometheus support", "err", err)
		cancel()
		return 1
	}

	var po *prometheuscontroller.Operator
	if prometheusSupported {
		po, err = prometheuscontroller.New(ctx, restConfig, cfg, logger, r, eventRecorderFactory, promControllerOptions...)
		if err != nil {
			level.Error(logger).Log("msg", "instantiating prometheus controller failed", "err", err)
			cancel()
			return 1
		}
	}

	prometheusAgentSupported, err := checkPrerequisites(
		ctx,
		logger,
		kclient,
		cfg.Namespaces.PrometheusAllowList.Slice(),
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
		pao, err = prometheusagentcontroller.New(ctx, restConfig, cfg, logger, r, scrapeConfigSupported, canReadStorageClass, eventRecorderFactory)
		if err != nil {
			level.Error(logger).Log("msg", "instantiating prometheus-agent controller failed", "err", err)
			cancel()
			return 1
		}
	}

	alertmanagerSupported, err := checkPrerequisites(
		ctx,
		logger,
		kclient,
		cfg.Namespaces.AlertmanagerAllowList.Slice(),
		monitoringv1.SchemeGroupVersion,
		monitoringv1.AlertmanagerName,
		k8sutil.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1.Version,
			Resource: monitoringv1.AlertmanagerName,
			Verbs:    []string{"get", "list", "watch"},
		},
		k8sutil.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1.Version,
			Resource: fmt.Sprintf("%s/status", monitoringv1.AlertmanagerName),
			Verbs:    []string{"update"},
		},
	)
	if err != nil {
		level.Error(logger).Log("msg", "failed to check Alertmanager support", "err", err)
		cancel()
		return 1
	}

	var ao *alertmanagercontroller.Operator
	if alertmanagerSupported {
		ao, err = alertmanagercontroller.New(ctx, restConfig, cfg, logger, r, canReadStorageClass, eventRecorderFactory)
		if err != nil {
			level.Error(logger).Log("msg", "instantiating alertmanager controller failed", "err", err)
			cancel()
			return 1
		}
	}

	thanosRulerSupported, err := checkPrerequisites(
		ctx,
		logger,
		kclient,
		cfg.Namespaces.ThanosRulerAllowList.Slice(),
		monitoringv1.SchemeGroupVersion,
		monitoringv1.ThanosRulerName,
		k8sutil.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1.Version,
			Resource: monitoringv1.ThanosRulerName,
			Verbs:    []string{"get", "list", "watch"},
		},
		k8sutil.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1.Version,
			Resource: fmt.Sprintf("%s/status", monitoringv1.ThanosRulerName),
			Verbs:    []string{"update"},
		},
	)
	if err != nil {
		level.Error(logger).Log("msg", "failed to check ThanosRuler support", "err", err)
		cancel()
		return 1
	}

	var to *thanoscontroller.Operator
	if thanosRulerSupported {
		to, err = thanoscontroller.New(ctx, restConfig, cfg, logger, r, canReadStorageClass, eventRecorderFactory)
		if err != nil {
			level.Error(logger).Log("msg", "instantiating thanos controller failed", "err", err)
			cancel()
			return 1
		}
	}

	var kec *kubelet.Controller
	if kubeletObject != "" {
		if kec, err = kubelet.New(
			log.With(logger, "component", "kubelet_endpoints"),
			restConfig,
			r,
			kubeletObject,
			kubeletSelector,
			cfg.Annotations,
			cfg.Labels,
			nodeAddressPriority,
		); err != nil {
			level.Error(logger).Log("msg", "instantiating kubelet endpoints controller failed", "err", err)
			cancel()
			return 1
		}
	}

	if po == nil && pao == nil && ao == nil && to == nil && kec == nil {
		level.Error(logger).Log("msg", "no controller can be started, check the RBAC permissions of the service account")
		cancel()
		return 1
	}

	// Setup the web server.
	mux := http.NewServeMux()

	admit := admission.New(log.With(logger, "component", "admissionwebhook"))
	admit.Register(mux)

	r.MustRegister(
		collectors.NewGoCollector(
			collectors.WithGoCollectorRuntimeMetrics(
				collectors.MetricsScheduler,
				collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile(`^/sync/.*`)},
			),
		),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		versioncollector.NewCollector("prometheus_operator"),
		cfg.Gates,
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

	srv, err := server.NewServer(logger, &serverConfig, mux)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create web server", "err", err)
		cancel()
		return 1
	}

	// Start the web server.
	wg.Go(func() error { return srv.Serve(ctx) })

	// Start the controllers.
	if po != nil {
		wg.Go(func() error { return po.Run(ctx) })
	}
	if pao != nil {
		wg.Go(func() error { return pao.Run(ctx) })
	}
	if ao != nil {
		wg.Go(func() error { return ao.Run(ctx) })
	}
	if to != nil {
		wg.Go(func() error { return to.Run(ctx) })
	}
	if kec != nil {
		wg.Go(func() error { return kec.Run(ctx) })
	}

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	select {
	case <-term:
		level.Info(logger).Log("msg", "received SIGTERM, exiting gracefully...")
	case <-ctx.Done():
	}

	if err := srv.Shutdown(ctx); err != nil {
		level.Warn(logger).Log("msg", "server shutdown error", "err", err)
	}

	cancel()
	if err := wg.Wait(); err != nil {
		level.Warn(logger).Log("msg", "unhandled error received. Exiting...", "err", err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(run(flag.CommandLine))
}
