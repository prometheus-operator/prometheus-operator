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
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"github.com/prometheus/common/version"
	"golang.org/x/sync/errgroup"
	appsv1 "k8s.io/api/apps/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	eventsv1 "k8s.io/api/events/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	k8sflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	crd "github.com/prometheus-operator/prometheus-operator/example"
	"github.com/prometheus-operator/prometheus-operator/internal/goruntime"
	logging "github.com/prometheus-operator/prometheus-operator/internal/log"
	"github.com/prometheus-operator/prometheus-operator/internal/metrics"
	"github.com/prometheus-operator/prometheus-operator/pkg/admission"
	alertmanagercontroller "github.com/prometheus-operator/prometheus-operator/pkg/alertmanager"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8s"
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
	logger *slog.Logger,
	kclient kubernetes.Interface,
	allowedNamespaces []string,
	groupVersion schema.GroupVersion,
	resource string,
	attributes ...k8s.ResourceAttribute,
) (bool, error) {
	installed, err := k8s.IsAPIGroupVersionResourceSupported(kclient.Discovery(), groupVersion, resource)
	if err != nil {
		return false, fmt.Errorf("failed to check presence of resource %q (group %q): %w", resource, groupVersion, err)
	}

	if !installed {
		logger.Warn(fmt.Sprintf("resource %q (group: %q) not installed in the cluster", resource, groupVersion))
		return false, nil
	}

	allowed, errs, err := k8s.IsAllowed(ctx, kclient.AuthorizationV1().SelfSubjectAccessReviews(), allowedNamespaces, attributes...)
	if err != nil {
		return false, fmt.Errorf("failed to check permissions on resource %q (group %q): %w", resource, groupVersion, err)
	}

	if !allowed {
		for _, reason := range errs {
			logger.Warn(fmt.Sprintf("missing permission on resource %q (group: %q)", resource, groupVersion), "reason", reason)
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

	disableUnmanagedPrometheusConfiguration bool

	// Parameters for the kubelet endpoints controller.
	kubeletObject        string
	kubeletSelector      operator.LabelSelector
	nodeAddressPriority  operator.NodeAddressPriority
	kubeletEndpoints     bool
	kubeletEndpointSlice bool
	kubeletSyncPeriod    time.Duration
	kubeletHTTPMetrics   bool

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
	fs.BoolVar(&kubeletEndpointSlice, "kubelet-endpointslice", false, "Create EndpointSlice objects for kubelet targets.")
	fs.BoolVar(&kubeletEndpoints, "kubelet-endpoints", true, "Create Endpoints objects for kubelet targets.")
	fs.DurationVar(&kubeletSyncPeriod, "kubelet-sync-period", 3*time.Minute, "How often the operator reconciles the kubelet Endpoints and EndpointSlice objects (e.g., 10s, 2m, 1h30m).")
	fs.BoolVar(&kubeletHTTPMetrics, "kubelet-http-metrics", true, "Include HTTP metrics port (10255) in kubelet service. Set to false if your cluster has disabled the insecure kubelet read-only port (e.g., GKE 1.32+).")

	// The Prometheus config reloader image is released along with the
	// Prometheus Operator image, tagged with the same semver version. Default to
	// the Prometheus Operator version if no Prometheus config reloader image is
	// specified.
	fs.StringVar(&cfg.ReloaderConfig.Image, "prometheus-config-reloader", operator.DefaultPrometheusConfigReloaderImage, "Prometheus config reloader image")
	fs.Var(&cfg.ReloaderConfig.CPURequests, "config-reloader-cpu-request", "Config Reloader CPU requests. Value \"0\" disables it and causes no request to be configured.")
	fs.Var(&cfg.ReloaderConfig.CPULimits, "config-reloader-cpu-limit", "Config Reloader CPU limits. Value \"0\" disables it and causes no limit to be configured.")
	fs.Var(&cfg.ReloaderConfig.MemoryRequests, "config-reloader-memory-request", "Config Reloader memory requests. Value \"0\" disables it and causes no request to be configured.")
	fs.Var(&cfg.ReloaderConfig.MemoryLimits, "config-reloader-memory-limit", "Config Reloader memory limits. Value \"0\" disables it and causes no limit to be configured.")
	fs.BoolVar(&cfg.ReloaderConfig.EnableProbes, "enable-config-reloader-probes", false, "Enable liveness, readiness, and startup probes for the config-reloader container. Default: false")

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
	fs.BoolVar(&cfg.WatchObjectRefsInAllNamespaces, "watch-referenced-objects-in-all-namespaces", false, "When true the operator watches for configmaps and secrets in both workload and configuration resource namespaces.\nWhen false (default), the operator will only watch for secrets and configmaps in:\n* Workload namespaces for Prometheus and PrometheusAgent resources.\n* Configuration namespaces for Alertmanager resources.")

	fs.Var(&cfg.Annotations, "annotations", "Annotations to be add to all resources created by the operator")
	fs.Var(&cfg.Labels, "labels", "Labels to be add to all resources created by the operator")

	fs.StringVar(&cfg.LocalHost, "localhost", "localhost", "EXPERIMENTAL (could be removed in future releases) - Host used to communicate between local services on a pod. Fixes issues where localhost resolves incorrectly.")
	fs.StringVar(&cfg.ClusterDomain, "cluster-domain", "", "The domain of the cluster. This is used to generate service FQDNs. If this is not specified, DNS search domain expansion is used instead.")

	fs.Var(&cfg.PromSelector, "prometheus-instance-selector", "Label selector to filter Prometheus and PrometheusAgent Custom Resources to watch.")
	fs.Var(&cfg.AlertmanagerSelector, "alertmanager-instance-selector", "Label selector to filter Alertmanager Custom Resources to watch.")
	fs.Var(&cfg.ThanosRulerSelector, "thanos-ruler-instance-selector", "Label selector to filter ThanosRuler Custom Resources to watch.")
	fs.Var(&cfg.SecretListWatchFieldSelector, "secret-field-selector", "Field selector to filter Secrets to watch")
	fs.Var(&cfg.SecretListWatchLabelSelector, "secret-label-selector", "Label selector to filter Secrets to watch")
	fs.Var(&cfg.ConfigMapListWatchFieldSelector, "configmap-field-selector", "Field selector to filter ConfigMaps to watch")
	fs.Var(&cfg.ConfigMapListWatchLabelSelector, "configmap-label-selector", "Label selector to filter ConfigMaps to watch")

	fs.Float64Var(&memlimitRatio, "auto-gomemlimit-ratio", defaultMemlimitRatio, "The ratio of reserved GOMEMLIMIT memory to the detected maximum container or system memory. The value should be greater than 0.0 and less than 1.0. Default: 0.0 (disabled).")
	fs.BoolVar(&disableUnmanagedPrometheusConfiguration, "disable-unmanaged-prometheus-configuration", false, "Disable support for unmanaged Prometheus configuration when all resource selectors are nil. As stated in the API documentation, unmanaged Prometheus configuration is a deprecated feature which can be avoided with '.spec.additionalScrapeConfigs' or the ScrapeConfig CRD. Default: false.")
	cfg.RegisterFeatureGatesFlags(fs, featureGates)

	logging.RegisterFlags(fs, &logConfig)
	versionutil.RegisterFlags(fs)

	// No need to check for errors because Parse would exit on error.
	_ = fs.Parse(os.Args[1:])
}

// checkStatusSubresourcePermissions returns true when the operator has the
// required permissions to update the status subresource of the provided
// configuration resources.
func checkStatusSubresourcePermissions(
	ctx context.Context,
	logger *slog.Logger,
	kclient kubernetes.Interface,
	gvrs []schema.GroupVersionResource,
) bool {
	ok := true
	for _, gvr := range gvrs {
		allowed, errs, err := k8s.IsAllowed(
			ctx,
			kclient.AuthorizationV1().SelfSubjectAccessReviews(),
			cfg.Namespaces.AllowList.Slice(),
			k8s.ResourceAttribute{
				Group:    gvr.Group,
				Version:  gvr.Version,
				Resource: fmt.Sprintf("%s/status", gvr.Resource),
				Verbs:    []string{"update"},
			},
		)
		if err != nil {
			ok = false
			logger.Error("failed to check permissions on status subresource", "err", err, "resource", gvr.String())
			continue
		}

		if allowed {
			continue
		}

		ok = false
		for _, reason := range errs {
			logger.Error("missing permission on status subresource", "reason", reason, "resource", gvr.String())
		}
	}

	return ok
}

func run(fs *flag.FlagSet) int {
	parseFlags(fs)

	if versionutil.ShouldPrintVersion() {
		versionutil.Print(os.Stdout, "prometheus-operator")
		return 0
	}

	// Determine command (default to "start")
	cmd := "start"
	if fs.NArg() > 0 {
		cmd = fs.Arg(0)
	}

	// Route to appropriate command handler
	switch cmd {
	case "start":
		return start()
	case "crds":
		return crds()
	case "full-crds":
		return fullCrds()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		fmt.Fprintln(os.Stderr, "Available commands: crds, full-crds, start")
		return 1
	}
}

// start runs the Prometheus Operator.
func start() int {
	logger, err := logging.NewLoggerSlog(logConfig)
	if err != nil {
		stdlog.Fatal(err)
	}
	klog.SetSlogLogger(logger)

	if err := cfg.Gates.UpdateFeatureGates(*featureGates.Map); err != nil {
		logger.Error("failed to update feature gates", "error", err)
		return 1
	}

	logger.Info("Starting Prometheus Operator", "version", version.Info(), "build_context", version.BuildContext(), "feature_gates", cfg.Gates.String())
	logger.Info("Operator's configuration",
		"watch_referenced_objects_in_all_namespaces", cfg.WatchObjectRefsInAllNamespaces,
		"controller_id", cfg.ControllerID,
		"enable_config_reloader_probes", cfg.ReloaderConfig.EnableProbes)
	goruntime.SetMaxProcs(logger)
	goruntime.SetMemLimit(logger, memlimitRatio)

	if len(cfg.Namespaces.AllowList) > 0 && len(cfg.Namespaces.DenyList) > 0 {
		return 1
	}
	if err := cfg.Namespaces.Finalize(); err != nil {
		logger.Error("failed to parse namespaces configuration", "configuration", cfg.Namespaces.String(), "error", err)
		return 1
	}
	logger.Info("Namespaces filtering configuration ", "config", cfg.Namespaces.String())

	ctx, cancel := context.WithCancel(context.Background())
	wg, ctx := errgroup.WithContext(ctx)
	r := metrics.NewRegistry("prometheus_operator")

	k8s.MustRegisterClientGoMetrics(r)

	restConfig, err := k8s.NewClusterConfig(k8s.ClusterConfig{
		Host:      apiServer,
		TLSConfig: tlsClientConfig,
		AsUser:    impersonateUser,
	})

	if err != nil {
		logger.Error("failed to create Kubernetes client configuration", "err", err)
		cancel()
		return 1
	}

	kclient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		logger.Error("failed to create Kubernetes client", "err", err)
		cancel()
		return 1
	}

	kubernetesVersion, err := kclient.Discovery().ServerVersion()
	if err != nil {
		logger.Error("failed to request Kubernetes server version", "err", err)
		cancel()
		return 1
	}

	cfg.KubernetesVersion, err = semver.ParseTolerant(kubernetesVersion.String())
	if err != nil {
		// If the Kubernetes version can't be parsed, assume v1.16.0 since this
		// is the minimal requirement for Prometheus Operator.
		cfg.KubernetesVersion = semver.MustParse("1.16.0")
		logger.Warn("failed to parse Kubernetes version", "version", kubernetesVersion.String(), "err", err)
	}
	logger.Info("connection established", "kubernetes_version", cfg.KubernetesVersion.String())

	var (
		alertmanagerControllerOptions = []alertmanagercontroller.ControllerOption{}
		promAgentControllerOptions    = []prometheusagentcontroller.ControllerOption{}
		promControllerOptions         = []prometheuscontroller.ControllerOption{}
		thanosControllerOptions       = []thanoscontroller.ControllerOption{}
	)
	if disableUnmanagedPrometheusConfiguration {
		logger.Info("Disabling support for unmanaged Prometheus configurations")
		promControllerOptions = append(promControllerOptions, prometheuscontroller.WithoutUnmanagedConfiguration())
	}
	// Check if we can read the storage classs
	canReadStorageClass, err := checkPrerequisites(
		ctx,
		logger,
		kclient,
		nil,
		storagev1.SchemeGroupVersion,
		storagev1.SchemeGroupVersion.WithResource("storageclasses").Resource,
		k8s.ResourceAttribute{
			Group:    storagev1.GroupName,
			Version:  storagev1.SchemeGroupVersion.Version,
			Resource: storagev1.SchemeGroupVersion.WithResource("storageclasses").Resource,
			Verbs:    []string{"get"},
		},
	)
	if err != nil {
		logger.Error("failed to check StorageClass support", "err", err)
		cancel()
		return 1
	}
	if canReadStorageClass {
		alertmanagerControllerOptions = append(alertmanagerControllerOptions, alertmanagercontroller.WithStorageClassValidation())
		promAgentControllerOptions = append(promAgentControllerOptions, prometheusagentcontroller.WithStorageClassValidation())
		promControllerOptions = append(promControllerOptions, prometheuscontroller.WithStorageClassValidation())
		thanosControllerOptions = append(thanosControllerOptions, thanoscontroller.WithStorageClassValidation())
	}

	canEmitEvents, reasons, err := k8s.IsAllowed(ctx, kclient.AuthorizationV1().SelfSubjectAccessReviews(), nil,
		k8s.ResourceAttribute{
			Group:    eventsv1.GroupName,
			Version:  eventsv1.SchemeGroupVersion.Version,
			Resource: eventsv1.SchemeGroupVersion.WithResource("events").Resource,
			Verbs:    []string{"create", "patch"},
		})
	if err != nil {
		logger.Error("failed to check Events support", "err", err)
		cancel()
		return 1
	}

	if !canEmitEvents {
		for _, reason := range reasons {
			logger.Warn("missing permission to emit events", "reason", reason)
		}
	}
	cfg.EventRecorderFactory = operator.NewEventRecorderFactory(canEmitEvents)

	scrapeConfigSupported, err := checkPrerequisites(
		ctx,
		logger,
		kclient,
		cfg.Namespaces.AllowList.Slice(),
		monitoringv1alpha1.SchemeGroupVersion,
		monitoringv1alpha1.ScrapeConfigName,
		k8s.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1alpha1.Version,
			Resource: monitoringv1alpha1.ScrapeConfigName,
			Verbs:    []string{"get", "list", "watch"},
		},
	)
	if err != nil {
		logger.Error("failed to check ScrapeConfig support", "err", err)
		cancel()
		return 1
	}
	if scrapeConfigSupported {
		promControllerOptions = append(promControllerOptions, prometheuscontroller.WithScrapeConfig())
		promAgentControllerOptions = append(promAgentControllerOptions, prometheusagentcontroller.WithScrapeConfig())
	}

	// EndpointSlice v1 became available with Kubernetes v1.21.0.
	endpointSliceSupported := cfg.KubernetesVersion.GTE(semver.MustParse("1.21.0"))
	logger.Info("Kubernetes API capabilities", "endpointslices", endpointSliceSupported)
	if endpointSliceSupported {
		promControllerOptions = append(promControllerOptions, prometheuscontroller.WithEndpointSlice())
		promAgentControllerOptions = append(promAgentControllerOptions, prometheusagentcontroller.WithEndpointSlice())
	}

	prometheusSupported, err := checkPrerequisites(
		ctx,
		logger,
		kclient,
		cfg.Namespaces.PrometheusAllowList.Slice(),
		monitoringv1.SchemeGroupVersion,
		monitoringv1.PrometheusName,
		k8s.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1.Version,
			Resource: monitoringv1.PrometheusName,
			Verbs:    []string{"get", "list", "watch"},
		},
		k8s.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1.Version,
			Resource: fmt.Sprintf("%s/status", monitoringv1.PrometheusName),
			Verbs:    []string{"update"},
		},
	)
	if err != nil {
		logger.Error("failed to check Prometheus support", "err", err)
		cancel()
		return 1
	}

	var po *prometheuscontroller.Operator
	if prometheusSupported {
		if cfg.Gates.Enabled(operator.StatusForConfigurationResourcesFeature) {
			if !checkStatusSubresourcePermissions(
				ctx,
				logger,
				kclient,
				[]schema.GroupVersionResource{
					monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ServiceMonitorName),
					monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PodMonitorName),
					monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ProbeName),
					monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PrometheusRuleName),
				},
			) {
				cancel()
				return 1
			}

			promControllerOptions = append(promControllerOptions, prometheuscontroller.WithConfigResourceStatus())
		}

		po, err = prometheuscontroller.New(ctx, restConfig, cfg, logger, r, promControllerOptions...)
		if err != nil {
			logger.Error("instantiating prometheus controller failed", "err", err)
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
		k8s.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1alpha1.Version,
			Resource: monitoringv1alpha1.PrometheusAgentName,
			Verbs:    []string{"get", "list", "watch"},
		},
		k8s.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1alpha1.Version,
			Resource: fmt.Sprintf("%s/status", monitoringv1alpha1.PrometheusAgentName),
			Verbs:    []string{"update"},
		},
	)
	if err != nil {
		logger.Error("failed to check PrometheusAgent support", "err", err)
		cancel()
		return 1
	}

	// If Prometheus Agent runs in DaemonSet mode, check if
	// the operator has proper RBAC permissions on the DaemonSet resource.
	if cfg.Gates.Enabled(operator.PrometheusAgentDaemonSetFeature) {
		allowed, errs, err := k8s.IsAllowed(ctx,
			kclient.AuthorizationV1().SelfSubjectAccessReviews(),
			cfg.Namespaces.PrometheusAllowList.Slice(),
			k8s.ResourceAttribute{
				Group:    appsv1.SchemeGroupVersion.Group,
				Version:  appsv1.SchemeGroupVersion.Version,
				Resource: "daemonsets",
				Verbs:    []string{"get", "list", "watch", "create", "update", "delete"},
			})
		if err != nil {
			logger.Error("failed to check permissions on DaemonSet resource", "err", err)
			cancel()
			return 1
		}
		if !allowed {
			for _, reason := range errs {
				logger.Error("missing permissions to manage Daemonset resource for Prometheus Agent", "reason", reason)
				cancel()
				return 1
			}
		}
	}

	var pao *prometheusagentcontroller.Operator
	if prometheusAgentSupported {
		if cfg.Gates.Enabled(operator.StatusForConfigurationResourcesFeature) {
			if !checkStatusSubresourcePermissions(
				ctx,
				logger,
				kclient,
				[]schema.GroupVersionResource{
					monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ServiceMonitorName),
					monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PodMonitorName),
					monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ProbeName),
				},
			) {
				cancel()
				return 1
			}

			promAgentControllerOptions = append(promAgentControllerOptions, prometheusagentcontroller.WithConfigResourceStatus())
		}

		pao, err = prometheusagentcontroller.New(ctx, restConfig, cfg, logger, r, promAgentControllerOptions...)
		if err != nil {
			logger.Error("instantiating prometheus-agent controller failed", "err", err)
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
		k8s.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1.Version,
			Resource: monitoringv1.AlertmanagerName,
			Verbs:    []string{"get", "list", "watch"},
		},
		k8s.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1.Version,
			Resource: fmt.Sprintf("%s/status", monitoringv1.AlertmanagerName),
			Verbs:    []string{"update"},
		},
	)
	if err != nil {
		logger.Error("failed to check Alertmanager support", "err", err)
		cancel()
		return 1
	}

	var ao *alertmanagercontroller.Operator
	if alertmanagerSupported {
		if cfg.Gates.Enabled(operator.StatusForConfigurationResourcesFeature) {
			// TODO: check permissions when implementing the AlertmanagerConfig status subresource.
			alertmanagerControllerOptions = append(alertmanagerControllerOptions, alertmanagercontroller.WithConfigResourceStatus())
		}

		ao, err = alertmanagercontroller.New(ctx, restConfig, cfg, logger, r, alertmanagerControllerOptions...)
		if err != nil {
			logger.Error("instantiating alertmanager controller failed", "err", err)
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
		k8s.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1.Version,
			Resource: monitoringv1.ThanosRulerName,
			Verbs:    []string{"get", "list", "watch"},
		},
		k8s.ResourceAttribute{
			Group:    monitoring.GroupName,
			Version:  monitoringv1.Version,
			Resource: fmt.Sprintf("%s/status", monitoringv1.ThanosRulerName),
			Verbs:    []string{"update"},
		},
	)
	if err != nil {
		logger.Error("failed to check ThanosRuler support", "err", err)
		cancel()
		return 1
	}

	var to *thanoscontroller.Operator
	if thanosRulerSupported {
		if cfg.Gates.Enabled(operator.StatusForConfigurationResourcesFeature) {
			if !checkStatusSubresourcePermissions(
				ctx,
				logger,
				kclient,
				[]schema.GroupVersionResource{
					monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PrometheusRuleName),
				},
			) {
				cancel()
				return 1
			}

			thanosControllerOptions = append(thanosControllerOptions, thanoscontroller.WithConfigResourceStatus())
		}

		to, err = thanoscontroller.New(ctx, restConfig, cfg, logger, r, thanosControllerOptions...)
		if err != nil {
			logger.Error("instantiating thanos controller failed", "err", err)
			cancel()
			return 1
		}
	}

	var kec *kubelet.Controller
	if kubeletObject != "" {
		opts := []kubelet.ControllerOption{
			kubelet.WithNodeAddressPriority(nodeAddressPriority.String()),
			kubelet.WithSyncPeriod(kubeletSyncPeriod),
			kubelet.WithHTTPMetrics(kubeletHTTPMetrics),
		}

		kubeletService := strings.Split(kubeletObject, "/")
		if len(kubeletService) != 2 {
			logger.Error(fmt.Sprintf("malformatted kubelet object string %q, must be in format \"namespace/name\"", kubeletObject))
			cancel()
			return 1
		}

		if kubeletEndpointSlice {
			allowed, errs, err := k8s.IsAllowed(
				ctx,
				kclient.AuthorizationV1().SelfSubjectAccessReviews(),
				[]string{kubeletService[0]},
				k8s.ResourceAttribute{
					Group:    discoveryv1.SchemeGroupVersion.Group,
					Version:  discoveryv1.SchemeGroupVersion.Version,
					Resource: "endpointslices",
					Verbs:    []string{"get", "list", "create", "update", "delete"},
				})
			if err != nil {
				logger.Error(fmt.Sprintf("failed to check permissions on resource 'endpointslices' (group %q)", discoveryv1.SchemeGroupVersion.Group), "err", err)
				cancel()
				return 1
			}

			if !allowed {
				for _, reason := range errs {
					logger.Warn(fmt.Sprintf("missing permission on resource 'endpointslices' (group: %q)", discoveryv1.SchemeGroupVersion.Group), "reason", reason)
				}
			} else {
				opts = append(opts, kubelet.WithEndpointSlice())
			}
		}

		if kubeletEndpoints {
			opts = append(opts, kubelet.WithEndpoints())
		}

		if kec, err = kubelet.New(
			logger.With("component", "kubelet_endpoints"),
			kclient,
			r,
			kubeletService[1],
			kubeletService[0],
			kubeletSelector,
			cfg.Annotations,
			cfg.Labels,
			opts...,
		); err != nil {
			logger.Error("instantiating kubelet endpoints controller failed", "err", err)
			cancel()
			return 1
		}
	}

	if po == nil && pao == nil && ao == nil && to == nil && kec == nil {
		logger.Error("no controller can be started, check the RBAC permissions of the service account")
		cancel()
		return 1
	}

	// Setup the web server.
	mux := http.NewServeMux()
	admit := admission.New(logger.With("component", "admissionwebhook"), model.LegacyValidation)
	admit.Register(mux)

	r.MustRegister(cfg.Gates)

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
		logger.Error("failed to create web server", "err", err)
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
		logger.Info("received SIGTERM, exiting gracefully...")
	case <-ctx.Done():
	}

	if err := srv.Shutdown(ctx); err != nil {
		logger.Warn("server shutdown error", "err", err)
	}

	cancel()
	if err := wg.Wait(); err != nil {
		logger.Warn("unhandled error received. Exiting...", "err", err)
		return 1
	}

	return 0
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [arguments] [<command>]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  start      Run the operator (default)")
		fmt.Fprintln(os.Stderr, "  crds       Print the CRDs in YAML format to standard output")
		fmt.Fprintln(os.Stderr, "  full-crds  Print the full CRDs (with all fields) in YAML format to standard output")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Arguments:")
		fs.PrintDefaults()
	}
	os.Exit(run(fs))
}

// crds prints all embedded CRDs to stdout.
func crds() int {
	if err := crd.PrintAll(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error printing CRDs: %v\n", err)
		return 1
	}
	return 0
}

// fullCrds prints all embedded full CRDs to stdout.
func fullCrds() int {
	if err := crd.PrintAllFull(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error printing full CRDs: %v\n", err)
		return 1
	}
	return 0
}
