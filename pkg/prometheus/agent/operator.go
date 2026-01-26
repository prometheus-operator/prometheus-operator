// Copyright 2023 The prometheus-operator Authors
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

package prometheusagent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure"
	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8s"
	"github.com/prometheus-operator/prometheus-operator/pkg/listwatch"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
)

const (
	resyncPeriod              = 5 * time.Minute
	controllerName            = "prometheusagent-controller"
	applicationNameLabelValue = "prometheus-agent"

	noSelectedResourcesMessage = "No ServiceMonitor, PodMonitor, Probe, and ScrapeConfig have been selected."
)

// Operator manages life cycle of Prometheus agent deployments and
// monitoring configurations.
type Operator struct {
	kclient  kubernetes.Interface
	mdClient metadata.Interface
	mclient  monitoringclient.Interface

	logger *slog.Logger

	accessor *operator.Accessor

	controllerID string

	nsPromInf cache.SharedIndexInformer
	nsMonInf  cache.SharedIndexInformer

	promInfs  *informers.ForResource
	smonInfs  *informers.ForResource
	pmonInfs  *informers.ForResource
	probeInfs *informers.ForResource
	sconInfs  *informers.ForResource
	cmapInfs  *informers.ForResource
	secrInfs  *informers.ForResource
	ssetInfs  *informers.ForResource
	dsetInfs  *informers.ForResource

	rr *operator.ResourceReconciler

	metrics         *operator.Metrics
	reconciliations *operator.ReconciliationTracker

	config prompkg.Config

	endpointSliceSupported bool // Whether the Kubernetes API supports the EndpointSlice kind.
	scrapeConfigSupported  bool
	canReadStorageClass    bool

	newEventRecorder operator.NewEventRecorderFunc

	statusReporter prompkg.StatusReporter

	daemonSetFeatureGateEnabled  bool
	configResourcesStatusEnabled bool

	finalizerSyncer *operator.FinalizerSyncer
}

type ControllerOption func(*Operator)

// WithEndpointSlice tells that the Kubernetes API supports the Endpointslice resource.
func WithEndpointSlice() ControllerOption {
	return func(o *Operator) {
		o.endpointSliceSupported = true
	}
}

// WithScrapeConfig tells that the controller manages ScrapeConfig objects.
func WithScrapeConfig() ControllerOption {
	return func(o *Operator) {
		o.scrapeConfigSupported = true
	}
}

// WithStorageClassValidation tells that the controller should verify that the
// Prometheus spec references a valid StorageClass name.
func WithStorageClassValidation() ControllerOption {
	return func(o *Operator) {
		o.canReadStorageClass = true
	}
}

// WithConfigResourceStatus tells that the controller can manage the status of
// configuration resources.
func WithConfigResourceStatus() ControllerOption {
	return func(o *Operator) {
		o.configResourcesStatusEnabled = true
	}
}

// New creates a new controller.
func New(ctx context.Context, restConfig *rest.Config, c operator.Config, logger *slog.Logger, r prometheus.Registerer, options ...ControllerOption) (*Operator, error) {
	logger = logger.With("component", controllerName)

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("instantiating kubernetes client failed: %w", err)
	}

	mdClient, err := metadata.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("instantiating metadata client failed: %w", err)
	}

	mclient, err := monitoringclient.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("instantiating monitoring client failed: %w", err)
	}

	// All the metrics exposed by the controller get the controller="prometheus-agent" label.
	r = prometheus.WrapRegistererWith(prometheus.Labels{"controller": "prometheus-agent"}, r)

	o := &Operator{
		kclient:  client,
		mdClient: mdClient,
		mclient:  mclient,
		logger:   logger,
		config: prompkg.Config{
			LocalHost:                  c.LocalHost,
			ReloaderConfig:             c.ReloaderConfig,
			PrometheusDefaultBaseImage: c.PrometheusDefaultBaseImage,
			ThanosDefaultBaseImage:     c.ThanosDefaultBaseImage,
			Annotations:                c.Annotations,
			Labels:                     c.Labels,
		},
		metrics:                      operator.NewMetrics(r),
		reconciliations:              &operator.ReconciliationTracker{},
		controllerID:                 c.ControllerID,
		newEventRecorder:             c.EventRecorderFactory(client, controllerName),
		configResourcesStatusEnabled: c.Gates.Enabled(operator.StatusForConfigurationResourcesFeature),
		finalizerSyncer:              operator.NewNoopFinalizerSyncer(),
	}
	o.metrics.MustRegister(
		o.reconciliations,
	)
	for _, opt := range options {
		opt(o)
	}

	if o.configResourcesStatusEnabled {
		o.finalizerSyncer = operator.NewFinalizerSyncer(mdClient, monitoringv1alpha1.SchemeGroupVersion.WithResource(monitoringv1alpha1.PrometheusAgentName))
	}

	o.promInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.Namespaces.PrometheusAllowList,
			c.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = c.PromSelector.String()
			},
		),
		monitoringv1alpha1.SchemeGroupVersion.WithResource(monitoringv1alpha1.PrometheusAgentName),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating prometheus-agent informers: %w", err)
	}

	var promStores []cache.Store
	for _, informer := range o.promInfs.GetInformers() {
		promStores = append(promStores, informer.Informer().GetStore())
	}
	o.metrics.MustRegister(prompkg.NewCollectorForStores(promStores...))

	o.rr = operator.NewResourceReconciler(
		o.logger,
		o,
		o.promInfs,
		o.metrics,
		monitoringv1alpha1.PrometheusAgentsKind,
		r,
		o.controllerID,
	)

	o.smonInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.Namespaces.AllowList,
			c.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ServiceMonitorName),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating servicemonitor informers: %w", err)
	}

	o.pmonInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.Namespaces.AllowList,
			c.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PodMonitorName),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating podmonitor informers: %w", err)
	}

	o.probeInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.Namespaces.AllowList,
			c.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ProbeName),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating probe informers: %w", err)
	}

	if o.scrapeConfigSupported {
		o.sconInfs, err = informers.NewInformersForResource(
			informers.NewMonitoringInformerFactories(
				c.Namespaces.AllowList,
				c.Namespaces.DenyList,
				mclient,
				resyncPeriod,
				nil,
			),
			monitoringv1alpha1.SchemeGroupVersion.WithResource(monitoringv1alpha1.ScrapeConfigName),
		)
		if err != nil {
			return nil, fmt.Errorf("error creating scrapeconfig informers: %w", err)
		}
	}

	allowList := c.Namespaces.PrometheusAllowList
	if c.WatchObjectRefsInAllNamespaces {
		allowList = operator.MergeAllowLists(
			c.Namespaces.PrometheusAllowList,
			c.Namespaces.AllowList,
		)
	}

	o.cmapInfs, err = informers.NewInformersForResourceWithTransform(
		informers.NewMetadataInformerFactory(
			allowList,
			c.Namespaces.DenyList,
			o.mdClient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.FieldSelector = c.ConfigMapListWatchFieldSelector.String()
				options.LabelSelector = c.ConfigMapListWatchLabelSelector.String()
			},
		),
		v1.SchemeGroupVersion.WithResource(string(v1.ResourceConfigMaps)),
		informers.PartialObjectMetadataStrip(operator.ConfigMapGVK()),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating configmap informers: %w", err)
	}

	o.secrInfs, err = informers.NewInformersForResourceWithTransform(
		informers.NewMetadataInformerFactory(
			allowList,
			c.Namespaces.DenyList,
			o.mdClient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.FieldSelector = c.SecretListWatchFieldSelector.String()
				options.LabelSelector = c.SecretListWatchLabelSelector.String()
			},
		),
		v1.SchemeGroupVersion.WithResource(string(v1.ResourceSecrets)),
		informers.PartialObjectMetadataStrip(operator.SecretGVK()),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating secrets informers: %w", err)
	}

	o.ssetInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			c.Namespaces.PrometheusAllowList,
			c.Namespaces.DenyList,
			o.kclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = prompkg.LabelSelectorForStatefulSets(prometheusMode)
			},
		),
		appsv1.SchemeGroupVersion.WithResource("statefulsets"),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating statefulset informers: %w", err)
	}

	if c.Gates.Enabled(operator.PrometheusAgentDaemonSetFeature) {
		o.daemonSetFeatureGateEnabled = true

		o.dsetInfs, err = informers.NewInformersForResource(
			informers.NewKubeInformerFactories(
				c.Namespaces.PrometheusAllowList,
				c.Namespaces.DenyList,
				o.kclient,
				resyncPeriod,
				func(options *metav1.ListOptions) {
					options.LabelSelector = fmt.Sprintf(
						"%s,%s,%s in (%s)",
						operator.ManagedByOperatorLabelSelector(),
						prompkg.PrometheusNameLabelName,
						prompkg.PrometheusModeLabelName,
						prometheusMode,
					)
				},
			),
			appsv1.SchemeGroupVersion.WithResource("daemonsets"),
		)
		if err != nil {
			return nil, fmt.Errorf("error creating daemonset informers: %w", err)
		}
	}

	newNamespaceInformer := func(o *Operator, allowList map[string]struct{}) (cache.SharedIndexInformer, error) {
		lw, privileged, err := listwatch.NewNamespaceListWatchFromClient(
			ctx,
			o.logger,
			c.KubernetesVersion,
			o.kclient.CoreV1(),
			o.kclient.AuthorizationV1().SelfSubjectAccessReviews(),
			allowList,
			c.Namespaces.DenyList,
		)
		if err != nil {
			return nil, err
		}

		logger.Debug("creating namespace informer", "privileged", privileged)
		return cache.NewSharedIndexInformer(
			o.metrics.NewInstrumentedListerWatcher(lw),
			&v1.Namespace{}, resyncPeriod, cache.Indexers{},
		), nil
	}

	o.nsMonInf, err = newNamespaceInformer(o, c.Namespaces.AllowList)
	if err != nil {
		return nil, err
	}

	if listwatch.IdenticalNamespaces(c.Namespaces.AllowList, c.Namespaces.PrometheusAllowList) {
		o.nsPromInf = o.nsMonInf
	} else {
		o.nsPromInf, err = newNamespaceInformer(o, c.Namespaces.PrometheusAllowList)
		if err != nil {
			return nil, err
		}
	}

	o.statusReporter = prompkg.StatusReporter{
		Kclient:         o.kclient,
		Reconciliations: o.reconciliations,
		SsetInfs:        o.ssetInfs,
		Rr:              o.rr,
	}

	return o, nil
}

// Run the controller.
func (c *Operator) Run(ctx context.Context) error {
	go c.rr.Run(ctx)
	defer c.rr.Stop()

	go c.promInfs.Start(ctx.Done())
	go c.smonInfs.Start(ctx.Done())
	go c.pmonInfs.Start(ctx.Done())
	go c.probeInfs.Start(ctx.Done())
	if c.scrapeConfigSupported {
		go c.sconInfs.Start(ctx.Done())
	}
	go c.cmapInfs.Start(ctx.Done())
	go c.secrInfs.Start(ctx.Done())
	go c.ssetInfs.Start(ctx.Done())
	if c.dsetInfs != nil {
		go c.dsetInfs.Start(ctx.Done())
	}
	go c.nsMonInf.Run(ctx.Done())
	if c.nsPromInf != c.nsMonInf {
		go c.nsPromInf.Run(ctx.Done())
	}
	if err := c.waitForCacheSync(ctx); err != nil {
		return err
	}

	// Refresh the status of the existing Prometheus agent objects.
	_ = c.promInfs.ListAll(labels.Everything(), func(obj any) {
		c.RefreshStatusFor(obj.(*monitoringv1alpha1.PrometheusAgent))
	})

	c.addHandlers()

	// TODO(simonpasquier): watch for PrometheusAgent pods instead of polling.
	go operator.StatusPoller(ctx, c)

	c.metrics.Ready().Set(1)
	<-ctx.Done()
	return nil
}

// Iterate implements the operator.StatusReconciler interface.
func (c *Operator) Iterate(processFn func(metav1.Object, []monitoringv1.Condition)) {
	if err := c.promInfs.ListAll(labels.Everything(), func(o any) {
		p := o.(*monitoringv1alpha1.PrometheusAgent)
		processFn(p, p.Status.Conditions)
	}); err != nil {
		c.logger.Error("failed to list PrometheusAgent objects", "err", err)
	}
}

// RefreshStatusFor implements the operator.StatusReconciler interface.
func (c *Operator) RefreshStatusFor(o metav1.Object) {
	c.rr.EnqueueForStatus(o)
}

// waitForCacheSync waits for the informers' caches to be synced.
func (c *Operator) waitForCacheSync(ctx context.Context) error {
	for _, infs := range []struct {
		name                 string
		informersForResource *informers.ForResource
	}{
		{"PrometheusAgent", c.promInfs},
		{"ServiceMonitor", c.smonInfs},
		{"PodMonitor", c.pmonInfs},
		{"Probe", c.probeInfs},
		{"ScrapeConfig", c.sconInfs},
		{"ConfigMap", c.cmapInfs},
		{"Secret", c.secrInfs},
		{"StatefulSet", c.ssetInfs},
		{"DaemonSet", c.dsetInfs},
	} {
		// Skipping informers that were not started. If prerequisites for a CRD were not met, their informer will be
		// nil. ScrapeConfig is one example.
		if infs.informersForResource == nil {
			continue
		}

		for _, inf := range infs.informersForResource.GetInformers() {
			if !operator.WaitForNamedCacheSync(ctx, "prometheusagent", c.logger.With("informer", infs.name), inf.Informer()) {
				return fmt.Errorf("failed to sync cache for %s informer", infs.name)
			}
		}
	}

	for _, inf := range []struct {
		name     string
		informer cache.SharedIndexInformer
	}{
		{"PromNamespace", c.nsPromInf},
		{"MonNamespace", c.nsMonInf},
	} {
		if !operator.WaitForNamedCacheSync(ctx, "prometheusagent", c.logger.With("informer", inf.name), inf.informer) {
			return fmt.Errorf("failed to sync cache for %s informer", inf.name)
		}
	}

	c.logger.Info("successfully synced all caches")
	return nil
}

// addHandlers adds the eventhandlers to the informers.
func (c *Operator) addHandlers() {
	c.promInfs.AddEventHandler(c.rr)

	c.ssetInfs.AddEventHandler(c.rr)

	if c.dsetInfs != nil {
		c.dsetInfs.AddEventHandler(c.rr)
	}

	c.smonInfs.AddEventHandler(operator.NewEventHandler(
		c.logger,
		c.accessor,
		c.metrics,
		monitoringv1.ServiceMonitorsKind,
		c.enqueueForMonitorNamespace,
		operator.WithFilter(
			operator.AnyFilter(
				operator.GenerationChanged,
				operator.LabelsChanged,
			),
		),
	))

	c.pmonInfs.AddEventHandler(operator.NewEventHandler(
		c.logger,
		c.accessor,
		c.metrics,
		monitoringv1.PodMonitorsKind,
		c.enqueueForMonitorNamespace,
		operator.WithFilter(
			operator.AnyFilter(
				operator.GenerationChanged,
				operator.LabelsChanged,
			),
		),
	))

	c.probeInfs.AddEventHandler(operator.NewEventHandler(
		c.logger,
		c.accessor,
		c.metrics,
		monitoringv1.ProbesKind,
		c.enqueueForMonitorNamespace,
		operator.WithFilter(
			operator.AnyFilter(
				operator.GenerationChanged,
				operator.LabelsChanged,
			),
		),
	))

	if c.sconInfs != nil {
		c.sconInfs.AddEventHandler(operator.NewEventHandler(
			c.logger,
			c.accessor,
			c.metrics,
			monitoringv1alpha1.ScrapeConfigsKind,
			c.enqueueForMonitorNamespace,
			operator.WithFilter(
				operator.AnyFilter(
					operator.GenerationChanged,
					operator.LabelsChanged,
				),
			),
		))
	}

	hasRefFunc := operator.HasReferenceFunc(
		c.promInfs,
		c.reconciliations,
	)
	c.cmapInfs.AddEventHandler(operator.NewEventHandler(
		c.logger,
		c.accessor,
		c.metrics,
		operator.ConfigMapGVK().Kind,
		c.enqueueForPrometheusNamespace,
		operator.WithFilter(operator.ResourceVersionChanged),
		operator.WithFilter(hasRefFunc),
	))

	c.secrInfs.AddEventHandler(operator.NewEventHandler(
		c.logger,
		c.accessor,
		c.metrics,
		operator.SecretGVK().Kind,
		c.enqueueForPrometheusNamespace,
		operator.WithFilter(operator.ResourceVersionChanged),
		operator.WithFilter(hasRefFunc),
	))

	// The controller needs to watch the namespaces in which the service/pod
	// monitors and rules live because a label change on a namespace may
	// trigger a configuration change.
	// It doesn't need to watch on addition/deletion though because it's
	// already covered by the event handlers on service/pod monitors and rules.
	_, _ = c.nsMonInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: c.handleMonitorNamespaceUpdate,
	})
}

// Sync implements the operator.Syncer interface.
func (c *Operator) Sync(ctx context.Context, key string) error {
	c.reconciliations.ResetStatus(key)
	err := c.sync(ctx, key)
	c.reconciliations.SetStatus(key, err)

	return err
}

func (c *Operator) sync(ctx context.Context, key string) error {
	p, err := operator.GetObjectFromKey[*monitoringv1alpha1.PrometheusAgent](c.promInfs, key)
	if err != nil {
		return err
	}

	if p == nil {
		c.reconciliations.ForgetObject(key)
		// Dependent resources are cleaned up by K8s via OwnerReferences
		return nil
	}

	logger := c.logger.With("key", key)
	logger.Info("sync prometheusagent")

	finalizersAdded, err := c.finalizerSyncer.Sync(ctx, p, c.rr.DeletionInProgress(p), func() error { return nil })
	if err != nil {
		return err
	}

	if finalizersAdded {
		// Since the object has been updated, let's trigger another sync.
		c.rr.EnqueueForReconciliation(p)
		return nil
	}

	// Check if the Agent instance is marked for deletion.
	if c.rr.DeletionInProgress(p) {
		c.reconciliations.ForgetObject(key)
		return nil
	}

	if p.Spec.Paused {
		logger.Info("no action taken (the resource is paused)")
		return nil
	}

	if ptr.Deref(p.Spec.Mode, "") == monitoringv1alpha1.DaemonSetPrometheusAgentMode && !c.daemonSetFeatureGateEnabled {
		return fmt.Errorf("feature gate for Prometheus Agent's DaemonSet mode is not enabled")
	}

	// Generate the configuration data.
	var (
		assetStore = assets.NewStoreBuilder(c.kclient.CoreV1(), c.kclient.CoreV1())
		opts       = []prompkg.ConfigGeneratorOption{}
	)
	if c.endpointSliceSupported {
		opts = append(opts, prompkg.WithEndpointSliceSupport())
	}
	if ptr.Deref(p.Spec.Mode, "") == monitoringv1alpha1.DaemonSetPrometheusAgentMode {
		opts = append(opts, prompkg.WithDaemonSet())
	}

	cg, err := prompkg.NewConfigGenerator(logger, p, opts...)
	if err != nil {
		return err
	}

	if err := c.createOrUpdateConfigurationSecret(ctx, logger, p, cg, assetStore); err != nil {
		return fmt.Errorf("creating config failed: %w", err)
	}
	c.reconciliations.UpdateReferenceTracker(key, assetStore.RefTracker())

	tlsAssets, err := operator.ReconcileShardedSecret(ctx, assetStore.TLSAssets(), c.kclient, prompkg.NewTLSAssetSecret(p, c.config))
	if err != nil {
		return fmt.Errorf("failed to reconcile the TLS secrets: %w", err)
	}

	if err := c.createOrUpdateWebConfigSecret(ctx, p); err != nil {
		return fmt.Errorf("synchronizing web config secret failed: %w", err)
	}

	switch ptr.Deref(p.Spec.Mode, "") {
	case monitoringv1alpha1.DaemonSetPrometheusAgentMode:
		err = c.syncDaemonSet(ctx, key, p, cg, tlsAssets)
	default:
		if err := operator.CheckStorageClass(ctx, c.canReadStorageClass, c.kclient, p.Spec.Storage); err != nil {
			return err
		}

		err = c.syncStatefulSet(ctx, key, p, cg, tlsAssets)
	}

	return err
}

func (c *Operator) syncDaemonSet(ctx context.Context, key string, p *monitoringv1alpha1.PrometheusAgent, cg *prompkg.ConfigGenerator, tlsAssets *operator.ShardedSecret) error {
	logger := c.logger.With("key", key)

	dsetClient := c.kclient.AppsV1().DaemonSets(p.Namespace)

	var notFound bool
	if _, err := c.dsetInfs.Get(keyToDaemonSetKey(p, key)); err != nil {
		notFound = apierrors.IsNotFound(err)
		if !notFound {
			return fmt.Errorf("retrieving daemonset failed: %w", err)
		}
	}

	dset, err := makeDaemonSet(
		p,
		c.config,
		cg,
		tlsAssets)
	if err != nil {
		return fmt.Errorf("making daemonset failed: %w", err)
	}

	if notFound {
		logger.Debug("creating daemonset")
		if _, err := dsetClient.Create(ctx, dset, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("creating daemonset failed: %w", err)
		}

		logger.Info("daemonset successfully created")
		return nil
	}

	err = k8s.UpdateDaemonSet(ctx, dsetClient, dset)
	sErr, ok := err.(*apierrors.StatusError)

	if ok && sErr.ErrStatus.Code == 422 && sErr.ErrStatus.Reason == metav1.StatusReasonInvalid {
		// Gather only reason for failed update
		failMsg := make([]string, len(sErr.ErrStatus.Details.Causes))
		for i, cause := range sErr.ErrStatus.Details.Causes {
			failMsg[i] = cause.Message
		}

		logger.Info("recreating DaemonSet because the update operation wasn't possible", "reason", strings.Join(failMsg, ", "))

		propagationPolicy := metav1.DeletePropagationForeground
		if err := dsetClient.Delete(ctx, dset.GetName(), metav1.DeleteOptions{PropagationPolicy: &propagationPolicy}); err != nil {
			return fmt.Errorf("failed to delete DaemonSet to avoid forbidden action: %w", err)
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("updating DaemonSet failed: %w", err)
	}

	return nil
}

func (c *Operator) syncStatefulSet(ctx context.Context, key string, p *monitoringv1alpha1.PrometheusAgent, cg *prompkg.ConfigGenerator, tlsAssets *operator.ShardedSecret) error {
	logger := c.logger.With("key", key)

	if p.Spec.ServiceName != nil {
		svcClient := c.kclient.CoreV1().Services(p.Namespace)
		selectorLabels := makeSelectorLabels(p.Name)

		if err := k8s.EnsureCustomGoverningService(ctx, p.Namespace, *p.Spec.ServiceName, svcClient, selectorLabels); err != nil {
			return err
		}
	} else {
		svc := prompkg.BuildStatefulSetService(
			governingServiceName,
			map[string]string{
				operator.ApplicationNameLabelKey: applicationNameLabelValue,
			},
			p,
			c.config,
		)

		if _, err := k8s.CreateOrUpdateService(ctx, c.kclient.CoreV1().Services(p.Namespace), svc); err != nil {
			return fmt.Errorf("synchronizing default governing service failed: %w", err)
		}
	}

	ssetClient := c.kclient.AppsV1().StatefulSets(p.Namespace)

	// Ensure we have a StatefulSet running Prometheus Agent deployed and that StatefulSet names are created correctly.
	expected := prompkg.ExpectedStatefulSetShardNames(p)
	for shard, ssetName := range expected {
		logger := logger.With("statefulset", ssetName, "shard", fmt.Sprintf("%d", shard))
		logger.Debug("reconciling statefulset")

		var notFound bool
		obj, err := c.ssetInfs.Get(prompkg.KeyToStatefulSetKey(p, key, shard))
		if err != nil {
			notFound = apierrors.IsNotFound(err)
			if !notFound {
				return fmt.Errorf("retrieving statefulset failed: %w", err)
			}
		}

		existingStatefulSet := &appsv1.StatefulSet{}
		if obj != nil {
			existingStatefulSet = obj.(*appsv1.StatefulSet)
			if c.rr.DeletionInProgress(existingStatefulSet) {
				// We want to avoid entering a hot-loop of update/delete cycles
				// here since the sts was marked for deletion in foreground,
				// which means it may take some time before the finalizers
				// complete and the resource disappears from the API. The
				// deletion timestamp will have been set when the initial
				// delete request was issued. In that case, we avoid further
				// processing.
				continue
			}
		}

		newSSetInputHash, err := createSSetInputHash(*p, c.config, tlsAssets, existingStatefulSet.Spec)
		if err != nil {
			return err
		}

		sset, err := makeStatefulSet(
			ssetName,
			p,
			c.config,
			cg,
			newSSetInputHash,
			int32(shard),
			tlsAssets)
		if err != nil {
			return fmt.Errorf("making statefulset failed: %w", err)
		}
		operator.SanitizeSTS(sset)

		if notFound {
			logger.Debug("creating statefulset")
			if _, err := k8s.CreateStatefulSetOrPatchLabels(ctx, ssetClient, sset); err != nil {
				return fmt.Errorf("failed to create statefulset: %w", err)
			}
			continue
		}

		if newSSetInputHash == existingStatefulSet.Annotations[operator.InputHashAnnotationKey] {
			logger.Debug("new statefulset generation inputs match current, skipping any actions")
			continue
		}

		logger.Debug("updating current statefulset because of hash divergence",
			"new_hash", newSSetInputHash,
			"existing_hash", existingStatefulSet.Annotations[operator.InputHashAnnotationKey],
		)

		if err = k8s.ForceUpdateStatefulSet(ctx, ssetClient, sset, func(reason string) {
			c.metrics.StsDeleteCreateCounter().Inc()
			logger.Info("recreating StatefulSet because the update operation wasn't possible", "reason", reason)
		}); err != nil {
			return err
		}
	}

	ssets := map[string]struct{}{}
	for _, ssetName := range expected {
		ssets[ssetName] = struct{}{}
	}

	var deleteErrs []error
	err := c.ssetInfs.ListAllByNamespace(p.Namespace, labels.SelectorFromSet(labels.Set{prompkg.PrometheusNameLabelName: p.Name, prompkg.PrometheusModeLabelName: prometheusMode}), func(obj any) {
		s := obj.(*appsv1.StatefulSet)

		if _, ok := ssets[s.Name]; ok {
			// Do not delete statefulsets that we still expect to exist. This
			// is to cleanup StatefulSets when shards are reduced.
			return
		}

		if c.rr.DeletionInProgress(s) {
			return
		}

		if delErr := ssetClient.Delete(ctx, s.GetName(), metav1.DeleteOptions{PropagationPolicy: ptr.To(metav1.DeletePropagationForeground)}); delErr != nil {
			if !apierrors.IsNotFound(delErr) {
				deleteErrs = append(deleteErrs, fmt.Errorf("failed to delete StatefulSet %s: %w", s.GetName(), delErr))
			}
		}
	})
	if err != nil {
		return fmt.Errorf("listing StatefulSet resources failed: %w", err)
	}
	if len(deleteErrs) > 0 {
		return fmt.Errorf("failed to clean up excess StatefulSets: %w", errors.Join(deleteErrs...))
	}

	return nil
}

func (c *Operator) createOrUpdateConfigurationSecret(ctx context.Context, logger *slog.Logger, p *monitoringv1alpha1.PrometheusAgent, cg *prompkg.ConfigGenerator, store *assets.StoreBuilder) error {
	resourceSelector, err := prompkg.NewResourceSelector(logger, p, store, c.nsMonInf, c.metrics, c.newEventRecorder(p))
	if err != nil {
		return err
	}

	smons, err := resourceSelector.SelectServiceMonitors(ctx, c.smonInfs.ListAllByNamespace)
	if err != nil {
		return fmt.Errorf("selecting ServiceMonitors failed: %w", err)
	}

	pmons, err := resourceSelector.SelectPodMonitors(ctx, c.pmonInfs.ListAllByNamespace)
	if err != nil {
		return fmt.Errorf("selecting PodMonitors failed: %w", err)
	}

	bmons, err := resourceSelector.SelectProbes(ctx, c.probeInfs.ListAllByNamespace)
	if err != nil {
		return fmt.Errorf("selecting Probes failed: %w", err)
	}

	var scrapeConfigs operator.TypedResourcesSelection[*monitoringv1alpha1.ScrapeConfig]
	if c.sconInfs != nil {
		scrapeConfigs, err = resourceSelector.SelectScrapeConfigs(ctx, c.sconInfs.ListAllByNamespace)
		if err != nil {
			return fmt.Errorf("selecting ScrapeConfigs failed: %w", err)
		}
	}

	if len(smons)+len(pmons)+len(bmons)+len(scrapeConfigs) == 0 {
		c.reconciliations.SetReasonAndMessage(operator.KeyForObject(p), operator.NoSelectedResourcesReason, noSelectedResourcesMessage)
	}

	if err := cg.AddRemoteWriteToStore(ctx, store, p.GetNamespace(), p.Spec.RemoteWrite); err != nil {
		return err
	}

	if err := prompkg.AddAPIServerConfigToStore(ctx, store, p.GetNamespace(), p.Spec.APIServerConfig); err != nil {
		return err
	}

	if err := prompkg.AddScrapeClassesToStore(ctx, store, p.GetNamespace(), p.Spec.ScrapeClasses); err != nil {
		return fmt.Errorf("failed to process scrape classes: %w", err)
	}

	sClient := c.kclient.CoreV1().Secrets(p.Namespace)
	additionalScrapeConfigs, err := k8s.LoadSecretRef(ctx, logger, sClient, p.Spec.AdditionalScrapeConfigs)
	if err != nil {
		return fmt.Errorf("loading additional scrape configs from Secret failed: %w", err)
	}

	// Update secret based on the most recent configuration.
	conf, err := cg.GenerateAgentConfiguration(
		smons.ValidResources(),
		pmons.ValidResources(),
		bmons.ValidResources(),
		scrapeConfigs.ValidResources(),
		store,
		additionalScrapeConfigs,
	)
	if err != nil {
		return fmt.Errorf("generating config failed: %w", err)
	}

	// Compress config to avoid 1mb secret limit for a while
	s, err := prompkg.MakeConfigurationSecret(p, c.config, conf)
	if err != nil {
		return fmt.Errorf("creating compressed secret failed: %w", err)
	}

	logger.Debug("updating Prometheus configuration secret")
	return k8s.CreateOrUpdateSecret(ctx, sClient, s)
}

func createSSetInputHash(p monitoringv1alpha1.PrometheusAgent, c prompkg.Config, tlsAssets *operator.ShardedSecret, ssSpec appsv1.StatefulSetSpec) (string, error) {
	var http2 *bool
	if p.Spec.Web != nil && p.Spec.Web.HTTPConfig != nil {
		http2 = p.Spec.Web.HTTPConfig.HTTP2
	}

	// The controller should ignore any changes to RevisionHistoryLimit field because
	// it may be modified by external actors.
	// See https://github.com/prometheus-operator/prometheus-operator/issues/5712
	ssSpec.RevisionHistoryLimit = nil

	hash, err := hashstructure.Hash(struct {
		PrometheusLabels      map[string]string
		PrometheusAnnotations map[string]string
		PrometheusGeneration  int64
		PrometheusWebHTTP2    *bool
		Config                prompkg.Config
		StatefulSetSpec       appsv1.StatefulSetSpec
		ShardedSecret         *operator.ShardedSecret
	}{
		PrometheusLabels:      p.Labels,
		PrometheusAnnotations: p.Annotations,
		PrometheusGeneration:  p.Generation,
		PrometheusWebHTTP2:    http2,
		Config:                c,
		StatefulSetSpec:       ssSpec,
		ShardedSecret:         tlsAssets,
	},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to calculate combined hash: %w", err)
	}

	return fmt.Sprintf("%d", hash), nil
}

// UpdateStatus updates the status subresource of the object identified by the given
// key.
// UpdateStatus implements the operator.Syncer interface.
func (c *Operator) UpdateStatus(ctx context.Context, key string) error {
	p, err := operator.GetObjectFromKey[*monitoringv1alpha1.PrometheusAgent](c.promInfs, key)

	if err != nil {
		return err
	}

	if p == nil {
		return nil
	}
	// Check if the Agent instance is marked for deletion.
	if c.rr.DeletionInProgress(p) {
		return nil
	}
	pStatus, err := c.statusReporter.Process(ctx, p, key)
	if err != nil {
		return fmt.Errorf("failed to get prometheus agent status: %w", err)
	}
	p.Status = *pStatus

	selectorLabels := makeSelectorLabels(p.Name)
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: selectorLabels})
	if err != nil {
		return fmt.Errorf("failed to create selector for prometheus agent scale status: %w", err)
	}
	p.Status.Selector = selector.String()
	p.Status.Shards = ptr.Deref(p.Spec.Shards, 1)

	if _, err = c.mclient.MonitoringV1alpha1().PrometheusAgents(p.Namespace).ApplyStatus(ctx, prompkg.ApplyConfigurationFromPrometheusAgent(p, true), metav1.ApplyOptions{FieldManager: k8s.PrometheusOperatorFieldManager, Force: true}); err != nil {
		c.logger.Info("failed to apply prometheus status subresource, trying again without scale fields", "err", err)
		// Try again, but this time does not update scale subresource.
		if _, err = c.mclient.MonitoringV1alpha1().PrometheusAgents(p.Namespace).ApplyStatus(ctx, prompkg.ApplyConfigurationFromPrometheusAgent(p, false), metav1.ApplyOptions{FieldManager: k8s.PrometheusOperatorFieldManager, Force: true}); err != nil {
			return fmt.Errorf("failed to Apply prometheus agent status subresource: %w", err)
		}
	}

	return nil
}

func (c *Operator) createOrUpdateWebConfigSecret(ctx context.Context, p *monitoringv1alpha1.PrometheusAgent) error {
	var fields monitoringv1.WebConfigFileFields
	if p.Spec.Web != nil {
		fields = p.Spec.Web.WebConfigFileFields
	}

	webConfig, err := webconfig.New(
		prompkg.WebConfigDir,
		prompkg.WebConfigSecretName(p),
		fields,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize web config: %w", err)
	}

	s := &v1.Secret{}
	operator.UpdateObject(
		s,
		operator.WithLabels(c.config.Labels),
		operator.WithAnnotations(c.config.Annotations),
		operator.WithManagingOwner(p),
	)

	if err := webConfig.CreateOrUpdateWebConfigSecret(ctx, c.kclient.CoreV1().Secrets(p.Namespace), s); err != nil {
		return fmt.Errorf("failed to reconcile web config secret: %w", err)
	}

	return nil
}

func (c *Operator) enqueueForPrometheusNamespace(nsName string) {
	c.enqueueForNamespace(c.nsPromInf.GetStore(), nsName)
}

func (c *Operator) enqueueForMonitorNamespace(nsName string) {
	c.enqueueForNamespace(c.nsMonInf.GetStore(), nsName)
}

// enqueueForNamespace enqueues all Prometheus object keys that belong to the
// given namespace or select objects in the given namespace.
func (c *Operator) enqueueForNamespace(store cache.Store, nsName string) {
	nsObject, found, err := store.GetByKey(nsName)
	if err != nil {
		c.logger.Error(
			"get namespace to enqueue Prometheus instances failed",
			"err", err,
		)
		return
	}
	if !found {
		c.logger.Error(fmt.Sprintf("get namespace to enqueue Prometheus instances failed: namespace %q does not exist", nsName))
		return
	}
	ns := nsObject.(*v1.Namespace)

	err = c.promInfs.ListAll(labels.Everything(), func(obj any) {
		// Check for Prometheus Agent instances in the namespace.
		p := obj.(*monitoringv1alpha1.PrometheusAgent)
		if p.Namespace == nsName {
			c.rr.EnqueueForReconciliation(p)
			return
		}

		// Check for Prometheus instances selecting ServiceMonitors in
		// the namespace.
		smNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.ServiceMonitorNamespaceSelector)
		if err != nil {
			c.logger.Error(
				fmt.Sprintf("failed to convert ServiceMonitorNamespaceSelector of %q to selector", p.Name),
				"err", err,
			)
			return
		}

		if smNSSelector.Matches(labels.Set(ns.Labels)) {
			c.rr.EnqueueForReconciliation(p)
			return
		}

		// Check for Prometheus instances selecting PodMonitors in the NS.
		pmNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.PodMonitorNamespaceSelector)
		if err != nil {
			c.logger.Error(
				fmt.Sprintf("failed to convert PodMonitorNamespaceSelector of %q to selector", p.Name),
				"err", err,
			)
			return
		}

		if pmNSSelector.Matches(labels.Set(ns.Labels)) {
			c.rr.EnqueueForReconciliation(p)
			return
		}

		// Check for Prometheus instances selecting Probes in the NS.
		bmNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.ProbeNamespaceSelector)
		if err != nil {
			c.logger.Error(
				fmt.Sprintf("failed to convert ProbeNamespaceSelector of %q to selector", p.Name),
				"err", err,
			)
			return
		}

		if bmNSSelector.Matches(labels.Set(ns.Labels)) {
			c.rr.EnqueueForReconciliation(p)
			return
		}
		// Check for Prometheus instances selecting Probes in the NS.
		ScrapeConfigNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.ScrapeConfigNamespaceSelector)
		if err != nil {
			c.logger.Error(
				fmt.Sprintf("failed to convert ScrapeConfigNamespaceSelector of %q to selector", p.Name),
				"err", err,
			)
			return
		}

		if ScrapeConfigNSSelector.Matches(labels.Set(ns.Labels)) {
			c.rr.EnqueueForReconciliation(p)
			return
		}
	})
	if err != nil {
		c.logger.Error("listing all Prometheus instances from cache failed",
			"err", err,
		)
	}

}

func (c *Operator) handleMonitorNamespaceUpdate(oldo, curo any) {
	old := oldo.(*v1.Namespace)
	cur := curo.(*v1.Namespace)

	c.logger.Debug("update handler", "namespace", cur.GetName(), "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	// Periodic resync may resend the Namespace without changes
	// in-between.
	if old.ResourceVersion == cur.ResourceVersion {
		return
	}

	c.logger.Debug("Monitor namespace updated", "namespace", cur.GetName())
	c.metrics.TriggerByCounter("Namespace", operator.UpdateEvent).Inc()

	// Check for Prometheus Agent instances selecting ServiceMonitors, PodMonitors,
	// Probes and ScrapeConfigs in the namespace.
	err := c.promInfs.ListAll(labels.Everything(), func(obj any) {
		p := obj.(*monitoringv1alpha1.PrometheusAgent)

		for name, selector := range map[string]*metav1.LabelSelector{
			"PodMonitors":     p.Spec.PodMonitorNamespaceSelector,
			"Probes":          p.Spec.ProbeNamespaceSelector,
			"ScrapeConfigs":   p.Spec.ScrapeConfigNamespaceSelector,
			"ServiceMonitors": p.Spec.ServiceMonitorNamespaceSelector,
		} {

			sync, err := k8s.LabelSelectionHasChanged(old.Labels, cur.Labels, selector)
			if err != nil {
				c.logger.Error(
					"failed to detect label selection change",
					"err", err,
					"name", p.Name,
					"namespace", p.Namespace,
					"subresource", name,
				)
				return
			}

			if sync {
				c.rr.EnqueueForReconciliation(p)
				return
			}
		}
	})
	if err != nil {
		c.logger.Error("listing all Prometheus Agent instances from cache failed",
			"err", err,
		)
	}
}

func makeSelectorLabels(name string) map[string]string {
	return map[string]string{
		operator.ApplicationNameLabelKey:     applicationNameLabelValue,
		operator.ManagedByLabelKey:           operator.ManagedByLabelValue,
		operator.ApplicationInstanceLabelKey: name,
		prompkg.PrometheusNameLabelName:      name,
	}
}

func keyToDaemonSetKey(p monitoringv1.PrometheusInterface, key string) string {
	keyParts := strings.Split(key, "/")
	return fmt.Sprintf("%s/%s", keyParts[0], fmt.Sprintf("%s-%s", prompkg.Prefix(p), keyParts[1]))
}
