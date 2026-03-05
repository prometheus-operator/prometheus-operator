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

package prometheus

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure"
	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
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
	"github.com/prometheus-operator/prometheus-operator/pkg/prometheus/validation"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
)

const (
	resyncPeriod              = 5 * time.Minute
	controllerName            = "prometheus-controller"
	applicationNameLabelValue = "prometheus"

	noSelectedResourcesMessage = "No ServiceMonitor, PodMonitor, Probe, ScrapeConfig, and PrometheusRule have been selected."

	unmanagedConfigurationReason  = "ConfigurationUnmanaged"
	unmanagedConfigurationMessage = "the operator doesn't manage the Prometheus configuration secret because neither serviceMonitorSelector nor podMonitorSelector, nor probeSelector, nor scrapeConfigSelector is specified. Unmanaged Prometheus configuration is deprecated, use additionalScrapeConfigs or the ScrapeConfig Custom Resource Definition instead. Unmanaged Prometheus configuration can also be disabled from the operator's command-line (check './operator --help')."
)

// Operator manages the life cycle of Prometheus deployments and
// monitoring configurations.
type Operator struct {
	kclient  kubernetes.Interface
	dclient  dynamic.Interface
	mdClient metadata.Interface
	mclient  monitoringclient.Interface

	logger   *slog.Logger
	accessor *operator.Accessor
	config   prompkg.Config

	controllerID string

	nsPromInf cache.SharedIndexInformer
	nsMonInf  cache.SharedIndexInformer

	promInfs  *informers.ForResource
	smonInfs  *informers.ForResource
	pmonInfs  *informers.ForResource
	probeInfs *informers.ForResource
	sconInfs  *informers.ForResource
	ruleInfs  *informers.ForResource
	cmapInfs  *informers.ForResource
	secrInfs  *informers.ForResource
	ssetInfs  *informers.ForResource

	rr *operator.ResourceReconciler

	metrics         *operator.Metrics
	reconciliations *operator.ReconciliationTracker
	statusReporter  prompkg.StatusReporter

	endpointSliceSupported        bool
	scrapeConfigSupported         bool
	canReadStorageClass           bool
	disableUnmanagedConfiguration bool
	retentionPoliciesEnabled      bool
	configResourcesStatusEnabled  bool

	newEventRecorder operator.NewEventRecorderFunc
	finalizerSyncer  *operator.FinalizerSyncer
}

type ControllerOption func(*Operator)

// selectedConfigResources return the configuration resources (serviceMonitors, podMonitors, probes, prometheusRules and scrapeConfigs)
// selected by Prometheus.
type selectedConfigResources struct {
	sMons         operator.TypedResourcesSelection[*monitoringv1.ServiceMonitor]
	pMons         operator.TypedResourcesSelection[*monitoringv1.PodMonitor]
	bMons         operator.TypedResourcesSelection[*monitoringv1.Probe]
	scrapeConfigs operator.TypedResourcesSelection[*monitoringv1alpha1.ScrapeConfig]
	rules         operator.PrometheusRuleSelection
}

func (s *selectedConfigResources) Len() int {
	return len(s.sMons) +
		len(s.pMons) +
		len(s.bMons) +
		len(s.scrapeConfigs) +
		s.rules.SelectedLen()
}

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

// WithoutUnmanagedConfiguration tells that the controller should not support
// unmanaged configurations.
func WithoutUnmanagedConfiguration() ControllerOption {
	return func(o *Operator) {
		o.disableUnmanagedConfiguration = true
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
func New(ctx context.Context, restConfig *rest.Config, c operator.Config, logger *slog.Logger, r prometheus.Registerer, opts ...ControllerOption) (*Operator, error) {
	logger = logger.With("component", controllerName)

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("instantiating kubernetes client failed: %w", err)
	}

	dclient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("instantiating dynamic client failed: %w", err)
	}

	mdClient, err := metadata.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("instantiating metadata client failed: %w", err)
	}

	mclient, err := monitoringclient.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("instantiating monitoring client failed: %w", err)
	}

	// All the metrics exposed by the controller get the controller="prometheus" label.
	r = prometheus.WrapRegistererWith(prometheus.Labels{"controller": "prometheus"}, r)

	o := &Operator{
		kclient:  client,
		dclient:  dclient,
		mdClient: mdClient,
		mclient:  mclient,
		logger:   logger,
		accessor: operator.NewAccessor(logger),

		config: prompkg.Config{
			LocalHost:                  c.LocalHost,
			ReloaderConfig:             c.ReloaderConfig,
			PrometheusDefaultBaseImage: c.PrometheusDefaultBaseImage,
			ThanosDefaultBaseImage:     c.ThanosDefaultBaseImage,
			Annotations:                c.Annotations,
			Labels:                     c.Labels,
		},
		metrics:         operator.NewMetrics(r),
		reconciliations: &operator.ReconciliationTracker{},

		controllerID:             c.ControllerID,
		newEventRecorder:         c.EventRecorderFactory(client, controllerName),
		retentionPoliciesEnabled: c.Gates.Enabled(operator.PrometheusShardRetentionPolicyFeature),
		finalizerSyncer:          operator.NewNoopFinalizerSyncer(),
	}
	for _, opt := range opts {
		opt(o)
	}

	if o.configResourcesStatusEnabled {
		o.finalizerSyncer = operator.NewFinalizerSyncer(mdClient, monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PrometheusName))
	}

	o.metrics.MustRegister(o.reconciliations)

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
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PrometheusName),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating prometheus informers: %w", err)
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
		monitoringv1.PrometheusesKind,
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
			return nil, fmt.Errorf("error creating scrapeconfigs informers: %w", err)
		}
	}
	o.ruleInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.Namespaces.AllowList,
			c.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PrometheusRuleName),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating prometheusrule informers: %w", err)
	}

	allowList := c.Namespaces.PrometheusAllowList
	if c.WatchObjectRefsInAllNamespaces {
		allowList = operator.MergeAllowLists(c.Namespaces.PrometheusAllowList, c.Namespaces.AllowList)
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

		o.logger.Debug("creating namespace informer", "privileged", privileged)
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

// waitForCacheSync waits for the informers' caches to be synced.
func (c *Operator) waitForCacheSync(ctx context.Context) error {
	for _, infs := range []struct {
		name                 string
		informersForResource *informers.ForResource
	}{
		{"Prometheus", c.promInfs},
		{"ServiceMonitor", c.smonInfs},
		{"PodMonitor", c.pmonInfs},
		{"PrometheusRule", c.ruleInfs},
		{"Probe", c.probeInfs},
		{"ScrapeConfig", c.sconInfs},
		{"ConfigMap", c.cmapInfs},
		{"Secret", c.secrInfs},
		{"StatefulSet", c.ssetInfs},
	} {
		// Skipping informers that were not started. If prerequisites for a CRD were not met, their informer will be
		// nil. ScrapeConfig is one example.
		if infs.informersForResource == nil {
			continue
		}

		for _, inf := range infs.informersForResource.GetInformers() {
			if !operator.WaitForNamedCacheSync(ctx, "prometheus", c.logger.With("informer", infs.name), inf.Informer()) {
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
		if !operator.WaitForNamedCacheSync(ctx, "prometheus", c.logger.With("informer", inf.name), inf.informer) {
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

	c.ruleInfs.AddEventHandler(operator.NewEventHandler(
		c.logger,
		c.accessor,
		c.metrics,
		monitoringv1.PrometheusRuleKind,
		c.enqueueForMonitorNamespace,
		operator.WithFilter(
			operator.AnyFilter(
				operator.GenerationChanged,
				operator.LabelsChanged,
			),
		),
	))

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
	go c.ruleInfs.Start(ctx.Done())
	go c.cmapInfs.Start(ctx.Done())
	go c.secrInfs.Start(ctx.Done())
	go c.ssetInfs.Start(ctx.Done())
	go c.nsMonInf.Run(ctx.Done())
	if c.nsPromInf != c.nsMonInf {
		go c.nsPromInf.Run(ctx.Done())
	}
	if err := c.waitForCacheSync(ctx); err != nil {
		return err
	}

	// Refresh the status of the existing Prometheus objects.
	_ = c.promInfs.ListAll(labels.Everything(), func(obj any) {
		c.RefreshStatusFor(obj.(*monitoringv1.Prometheus))
	})

	c.addHandlers()

	// TODO(simonpasquier): watch for Prometheus pods instead of polling.
	go operator.StatusPoller(ctx, c)

	c.metrics.Ready().Set(1)
	<-ctx.Done()
	return nil
}

// Iterate implements the operator.StatusReconciler interface.
func (c *Operator) Iterate(processFn func(operator.StatusGetter)) {
	if err := c.promInfs.ListAll(labels.Everything(), func(o any) {
		processFn(o.(*monitoringv1.Prometheus))
	}); err != nil {
		c.logger.Error("failed to list Prometheus objects", "err", err)
	}
}

// RefreshStatusFor implements the operator.StatusReconciler interface.
func (c *Operator) RefreshStatusFor(o metav1.Object) {
	c.rr.EnqueueForStatus(o)
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
		c.logger.Error(
			fmt.Sprintf("get namespace to enqueue Prometheus instances failed: namespace %q does not exist", nsName),
		)
		return
	}
	ns := nsObject.(*v1.Namespace)

	err = c.promInfs.ListAll(labels.Everything(), func(obj any) {
		// Check for Prometheus instances in the namespace.
		p := obj.(*monitoringv1.Prometheus)
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

		// Check for Prometheus instances selecting PrometheusRules in
		// the NS.
		ruleNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.RuleNamespaceSelector)
		if err != nil {
			c.logger.Error(
				fmt.Sprintf("failed to convert RuleNamespaceSelector of %q to selector", p.Name),
				"err", err,
			)
			return
		}

		if ruleNSSelector.Matches(labels.Set(ns.Labels)) {
			c.rr.EnqueueForReconciliation(p)
			return
		}
		// Check for Prometheus instances selecting ScrapeConfigs in
		// the NS.
		scrapeConfigNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.ScrapeConfigNamespaceSelector)
		if err != nil {
			c.logger.Error(
				fmt.Sprintf("failed to convert ScrapeConfigNamespaceSelector of %q to selector", p.Name),
				"err", err,
			)
			return
		}

		if scrapeConfigNSSelector.Matches(labels.Set(ns.Labels)) {
			c.rr.EnqueueForReconciliation(p)
			return
		}
	})
	if err != nil {
		c.logger.Error(
			"listing all Prometheus instances from cache failed",
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

	// Check for Prometheus instances selecting ServiceMonitors, PodMonitors,
	// Probes, PrometheusRules and ScrapeConfigs in the namespace.
	err := c.promInfs.ListAll(labels.Everything(), func(obj any) {
		p := obj.(*monitoringv1.Prometheus)

		for name, selector := range map[string]*metav1.LabelSelector{
			"PodMonitors":     p.Spec.PodMonitorNamespaceSelector,
			"Probes":          p.Spec.ProbeNamespaceSelector,
			"PrometheusRules": p.Spec.RuleNamespaceSelector,
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
		c.logger.Error(
			"listing all Prometheus instances from cache failed",
			"err", err,
		)
	}
}

// Sync implements the operator.Syncer interface.
func (c *Operator) Sync(ctx context.Context, key string) error {
	c.reconciliations.ResetStatus(key)
	err := c.sync(ctx, key)
	c.reconciliations.SetStatus(key, err)

	return err
}

func (c *Operator) sync(ctx context.Context, key string) error {
	p, err := operator.GetObjectFromKey[*monitoringv1.Prometheus](c.promInfs, key)
	if err != nil {
		return err
	}

	if p == nil {
		c.reconciliations.ForgetObject(key)
		// Dependent resources are cleaned up by K8s via OwnerReferences
		return nil
	}

	logger := c.logger.With("key", key)
	logger.Info("sync prometheus")

	finalizerAdded, err := c.finalizerSyncer.Sync(ctx, p, c.rr.DeletionInProgress(p), func() error {
		return c.configResStatusCleanup(ctx, p)
	})
	if err != nil {
		return err
	}

	if finalizerAdded {
		// Since the finalizer has been added to the object, let's trigger another sync.
		c.rr.EnqueueForReconciliation(p)
		return nil
	}

	if c.rr.DeletionInProgress(p) {
		c.reconciliations.ForgetObject(key)
		return nil
	}

	if p.Spec.Paused {
		logger.Info("no action taken (the resource is paused)")
		return nil
	}

	c.recordDeprecatedFields(key, logger, p)

	if err := operator.CheckStorageClass(ctx, c.canReadStorageClass, c.kclient, p.Spec.Storage); err != nil {
		return err
	}

	assetStore := assets.NewStoreBuilder(c.kclient.CoreV1(), c.kclient.CoreV1())

	// Select configuration resources.
	resources, err := c.getSelectedConfigResources(ctx, logger, p, assetStore)
	if err != nil {
		return err
	}

	if resources.Len() == 0 {
		c.reconciliations.SetReasonAndMessage(key, operator.NoSelectedResourcesReason, noSelectedResourcesMessage)
	}

	ruleConfigMapNames, err := c.createOrUpdateRuleConfigMaps(ctx, p, resources.rules, logger)
	if err != nil {
		return err
	}

	opts := []prompkg.ConfigGeneratorOption{}
	if c.endpointSliceSupported {
		opts = append(opts, prompkg.WithEndpointSliceSupport())
	}
	cg, err := prompkg.NewConfigGenerator(logger, p, opts...)
	if err != nil {
		return err
	}

	if err := c.createOrUpdateConfigurationSecret(ctx, logger, p, cg, ruleConfigMapNames, assetStore, resources); err != nil {
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

	if err := c.createOrUpdateThanosConfigSecret(ctx, p); err != nil {
		return fmt.Errorf("failed to reconcile Thanos config secret: %w", err)
	}

	if p.Spec.ServiceName != nil {
		svcClient := c.kclient.CoreV1().Services(p.Namespace)
		selectorLabels := makeSelectorLabels(p.Name)

		if err := k8s.EnsureCustomGoverningService(ctx, p.Namespace, *p.Spec.ServiceName, svcClient, selectorLabels); err != nil {
			return err
		}
	} else {
		// Reconcile the default governing service.
		svc := prompkg.BuildStatefulSetService(
			governingServiceName,
			map[string]string{
				operator.ApplicationNameLabelKey: applicationNameLabelValue,
			},
			p,
			c.config,
		)

		if p.Spec.Thanos != nil {
			svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
				Name:       "grpc",
				Port:       10901,
				TargetPort: intstr.FromString("grpc"),
			})
		}

		if _, err := k8s.CreateOrUpdateService(ctx, c.kclient.CoreV1().Services(p.Namespace), svc); err != nil {
			return fmt.Errorf("synchronizing default governing service failed: %w", err)
		}
	}

	ssetClient := c.kclient.AppsV1().StatefulSets(p.Namespace)

	// Ensure we have a StatefulSet running Prometheus deployed and that StatefulSet names are created correctly.
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

		newSSetInputHash, err := createSSetInputHash(*p, c.config, ruleConfigMapNames, tlsAssets, existingStatefulSet.Spec)
		if err != nil {
			return err
		}

		sset, err := makeStatefulSet(
			ssetName,
			p,
			c.config,
			cg,
			ruleConfigMapNames,
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

		logger.Debug(
			"updating current statefulset because of hash divergence",
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
	err = c.ssetInfs.ListAllByNamespace(p.Namespace, labels.SelectorFromSet(labels.Set{prompkg.PrometheusNameLabelName: p.Name, prompkg.PrometheusModeLabelName: prometheusMode}), func(obj any) {
		s := obj.(*appsv1.StatefulSet)

		if _, ok := ssets[s.Name]; ok {
			// Do not delete statefulsets that we still expect to exist. This
			// is to cleanup StatefulSets when shards are reduced.
			return
		}

		if c.rr.DeletionInProgress(s) {
			return
		}

		shouldRetain, retainErr := c.shouldRetain(p)
		if retainErr != nil {
			deleteErrs = append(deleteErrs, fmt.Errorf("failed to determine if StatefulSet %s should be retained: %w", s.GetName(), retainErr))
			return
		}
		if shouldRetain {
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

	err = c.updateConfigResourcesStatus(ctx, p, *resources)

	return err
}

// updateConfigResourcesStatus updates the status of the selected configuration
// resources (ServiceMonitor, PodMonitor, ScrapeConfig and Probe).
func (c *Operator) updateConfigResourcesStatus(ctx context.Context, p *monitoringv1.Prometheus, resources selectedConfigResources) error {
	if !c.configResourcesStatusEnabled {
		return nil
	}

	var configResourceSyncer = operator.NewConfigResourceSyncer(p, c.dclient, c.accessor)

	// Update the status of selected serviceMonitors.
	for key, configResource := range resources.sMons {
		if err := configResourceSyncer.UpdateBinding(ctx, configResource.Resource(), configResource.Conditions()); err != nil {
			return fmt.Errorf("failed to update ServiceMonitor %s status: %w", key, err)
		}
	}

	// Update the status of selected podMonitors.
	for key, configResource := range resources.pMons {
		if err := configResourceSyncer.UpdateBinding(ctx, configResource.Resource(), configResource.Conditions()); err != nil {
			return fmt.Errorf("failed to update PodMonitor %s status: %w", key, err)
		}
	}

	// Update the status of selected probes.
	for key, configResource := range resources.bMons {
		if err := configResourceSyncer.UpdateBinding(ctx, configResource.Resource(), configResource.Conditions()); err != nil {
			return fmt.Errorf("failed to update Probe %s status: %w", key, err)
		}
	}

	// Update the status of selected scrapeConfigs.
	for key, configResource := range resources.scrapeConfigs {
		if err := configResourceSyncer.UpdateBinding(ctx, configResource.Resource(), configResource.Conditions()); err != nil {
			return fmt.Errorf("failed to update ScrapeConfig %s status: %w", key, err)
		}
	}

	// Update the status of selected prometheusRules.
	for key, configResource := range resources.rules.Selected() {
		if err := configResourceSyncer.UpdateBinding(ctx, configResource.Resource(), configResource.Conditions()); err != nil {
			return fmt.Errorf("failed to update PrometheusRule %s status: %w", key, err)
		}
	}

	// Remove bindings from serviceMonitors which reference the
	// workload but aren't selected anymore.
	if err := operator.CleanupBindings(ctx, c.smonInfs.ListAll, resources.sMons, configResourceSyncer); err != nil {
		return fmt.Errorf("failed to remove bindings for service monitors: %w", err)
	}

	// Remove bindings from podMonitors which reference the
	// workload but aren't selected anymore.
	if err := operator.CleanupBindings(ctx, c.pmonInfs.ListAll, resources.pMons, configResourceSyncer); err != nil {
		return fmt.Errorf("failed to remove bindings for pod monitors: %w", err)
	}

	// Remove bindings from scrapeConfigs which reference the
	// workload but aren't selected anymore.
	// Only cleanup if ScrapeConfig support is enabled (sconInfs is initialized).
	if c.sconInfs != nil {
		if err := operator.CleanupBindings(ctx, c.sconInfs.ListAll, resources.scrapeConfigs, configResourceSyncer); err != nil {
			return fmt.Errorf("failed to remove bindings for scrapeConfigs: %w", err)
		}
	}

	// Remove bindings from probes which reference the
	// workload but aren't selected anymore.
	if err := operator.CleanupBindings(ctx, c.probeInfs.ListAll, resources.bMons, configResourceSyncer); err != nil {
		return fmt.Errorf("failed to remove bindings for probes: %w", err)
	}

	// Remove bindings from prometheusRules which reference the
	// workload but aren't selected anymore.
	if err := operator.CleanupBindings(ctx, c.ruleInfs.ListAll, resources.rules.Selected(), configResourceSyncer); err != nil {
		return fmt.Errorf("failed to remove bindings for prometheusRules: %w", err)
	}
	return nil
}

// configResStatusCleanup removes prometheus bindings from the configuration resources (ServiceMonitor, PodMonitor, ScrapeConfig, PrometheusRule and PodMonitor).
func (c *Operator) configResStatusCleanup(ctx context.Context, p *monitoringv1.Prometheus) error {
	if !c.configResourcesStatusEnabled {
		return nil
	}

	var configResourceSyncer = operator.NewConfigResourceSyncer(p, c.dclient, c.accessor)

	// Remove bindings from all serviceMonitors which reference the workload.
	if err := operator.CleanupBindings(ctx, c.smonInfs.ListAll, operator.TypedResourcesSelection[*monitoringv1.ServiceMonitor]{}, configResourceSyncer); err != nil {
		return fmt.Errorf("failed to remove bindings for service monitors: %w", err)
	}

	// Remove bindings from all podMonitors which reference the workload.
	if err := operator.CleanupBindings(ctx, c.pmonInfs.ListAll, operator.TypedResourcesSelection[*monitoringv1.PodMonitor]{}, configResourceSyncer); err != nil {
		return fmt.Errorf("failed to remove bindings for pod monitors: %w", err)
	}

	// Remove bindings from all scrapeConfigs which reference the workload.
	// Only cleanup if ScrapeConfig support is enabled (sconInfs is initialized).
	if c.sconInfs != nil {
		if err := operator.CleanupBindings(ctx, c.sconInfs.ListAll, operator.TypedResourcesSelection[*monitoringv1alpha1.ScrapeConfig]{}, configResourceSyncer); err != nil {
			return fmt.Errorf("failed to remove bindings for scrapeConfigs: %w", err)
		}
	}

	// Remove bindings from all probes which reference the workload.
	if err := operator.CleanupBindings(ctx, c.probeInfs.ListAll, operator.TypedResourcesSelection[*monitoringv1.Probe]{}, configResourceSyncer); err != nil {
		return fmt.Errorf("failed to remove bindings for probes: %w", err)
	}

	// Remove bindings from all prometheusRules which reference the workload.
	if err := operator.CleanupBindings(ctx, c.ruleInfs.ListAll, operator.TypedResourcesSelection[*monitoringv1.PrometheusRule]{}, configResourceSyncer); err != nil {
		return fmt.Errorf("failed to remove bindings for prometheusRule: %w", err)
	}
	return nil
}

// As the ShardRetentionPolicy feature evolves, should retain will evolve accordingly.
// For now, shouldRetain just returns the appropriate boolean based on the retention type.
func (c *Operator) shouldRetain(p *monitoringv1.Prometheus) (bool, error) {
	if !c.retentionPoliciesEnabled {
		// Feature-gate is disabled, default behavior is always to delete.
		return false, nil
	}
	if p.Spec.ShardRetentionPolicy == nil {
		// ShardRetentionPolicy not configured, default behavior is to delete.
		return false, nil
	}
	if ptr.Deref(p.Spec.ShardRetentionPolicy.WhenScaled,
		monitoringv1.DeleteWhenScaledRetentionType) == monitoringv1.RetainWhenScaledRetentionType {
		return true, nil
	}

	return false, nil
}

// UpdateStatus updates the status subresource of the object identified by the given
// key.
// UpdateStatus implements the operator.Syncer interface.
func (c *Operator) UpdateStatus(ctx context.Context, key string) error {
	p, err := operator.GetObjectFromKey[*monitoringv1.Prometheus](c.promInfs, key)
	if err != nil {
		return err
	}

	if p == nil {
		return nil
	}

	if c.rr.DeletionInProgress(p) {
		return nil
	}
	pStatus, err := c.statusReporter.Process(ctx, p, key)
	if err != nil {
		return fmt.Errorf("failed to get prometheus status: %w", err)
	}

	p.Status = *pStatus
	selectorLabels := makeSelectorLabels(p.Name)
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: selectorLabels})
	if err != nil {
		return fmt.Errorf("failed to create selector for prometheus scale status: %w", err)
	}
	p.Status.Selector = selector.String()
	p.Status.Shards = ptr.Deref(p.Spec.Shards, 1)

	if _, err = c.mclient.MonitoringV1().Prometheuses(p.Namespace).ApplyStatus(ctx, prompkg.ApplyConfigurationFromPrometheus(p, true), metav1.ApplyOptions{FieldManager: k8s.PrometheusOperatorFieldManager, Force: true}); err != nil {
		c.logger.Info("failed to apply prometheus status subresource, trying again without scale fields", "err", err)
		// Try again, but this time does not update scale subresource.
		if _, err = c.mclient.MonitoringV1().Prometheuses(p.Namespace).ApplyStatus(ctx, prompkg.ApplyConfigurationFromPrometheus(p, false), metav1.ApplyOptions{FieldManager: k8s.PrometheusOperatorFieldManager, Force: true}); err != nil {
			return fmt.Errorf("failed to apply prometheus status subresource: %w", err)
		}
	}

	return nil
}

func (c *Operator) recordDeprecatedFields(key string, logger *slog.Logger, p *monitoringv1.Prometheus) {
	deprecationWarningf := "field %q is deprecated, field %q should be used instead"
	var deprecations []string

	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if p.Spec.BaseImage != "" {
		deprecations = append(deprecations, fmt.Sprintf(deprecationWarningf, "spec.baseImage", "spec.image"))
	}

	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if p.Spec.Tag != "" {
		deprecations = append(deprecations, fmt.Sprintf(deprecationWarningf, "spec.tag", "spec.image"))
	}

	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if p.Spec.SHA != "" {
		deprecations = append(deprecations, fmt.Sprintf(deprecationWarningf, "spec.sha", "spec.image"))
	}

	if p.Spec.Thanos != nil {
		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if p.Spec.Thanos.BaseImage != nil && *p.Spec.Thanos.BaseImage != "" {
			deprecations = append(deprecations, fmt.Sprintf(deprecationWarningf, "spec.thanos.baseImage", "spec.thanos.image"))
		}

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if p.Spec.Thanos.Tag != nil && *p.Spec.Thanos.Tag != "" {
			deprecations = append(deprecations, fmt.Sprintf(deprecationWarningf, "spec.thanos.tag", "spec.thanos.image"))
		}

		//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
		if p.Spec.Thanos.SHA != nil && *p.Spec.Thanos.SHA != "" {
			deprecations = append(deprecations, fmt.Sprintf(deprecationWarningf, "spec.thanos.sha", "spec.thanos.image"))
		}
	}

	if len(deprecations) > 0 {
		for _, m := range deprecations {
			logger.Warn(m)
		}
		c.reconciliations.SetReasonAndMessage(key, operator.DeprecatedFieldsInUseReason, strings.Join(deprecations, "; "))
		return
	}

	if c.unmanagedPrometheusConfiguration(p) {
		logger.Warn(unmanagedConfigurationMessage)
		c.reconciliations.SetReasonAndMessage(key, unmanagedConfigurationReason, unmanagedConfigurationMessage)
	}
}

func (c *Operator) unmanagedPrometheusConfiguration(p *monitoringv1.Prometheus) bool {
	return !c.disableUnmanagedConfiguration &&
		p.Spec.ServiceMonitorSelector == nil &&
		p.Spec.PodMonitorSelector == nil &&
		p.Spec.ProbeSelector == nil &&
		p.Spec.ScrapeConfigSelector == nil
}

func createSSetInputHash(p monitoringv1.Prometheus, c prompkg.Config, ruleConfigMapNames []string, tlsAssets *operator.ShardedSecret, ssSpec appsv1.StatefulSetSpec) (string, error) {
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
		RuleConfigMaps        []string `hash:"set"`
		ShardedSecret         *operator.ShardedSecret
	}{
		PrometheusLabels:      p.Labels,
		PrometheusAnnotations: p.Annotations,
		PrometheusGeneration:  p.Generation,
		PrometheusWebHTTP2:    http2,
		Config:                c,
		StatefulSetSpec:       ssSpec,
		RuleConfigMaps:        ruleConfigMapNames,
		ShardedSecret:         tlsAssets,
	},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to calculate combined hash: %w", err)
	}

	return fmt.Sprintf("%d", hash), nil
}

// getSeletedConfigResources returns all the configuration resources (PodMonitor, ServiceMonitor, Probes and ScrapeConfigs) selected by the Prometheus.
func (c *Operator) getSelectedConfigResources(ctx context.Context, logger *slog.Logger, p *monitoringv1.Prometheus, store *assets.StoreBuilder) (*selectedConfigResources, error) {
	resourceSelector, err := prompkg.NewResourceSelector(logger, p, store, c.nsMonInf, c.metrics, c.newEventRecorder(p))

	if err != nil {
		return nil, err
	}
	smons, err := resourceSelector.SelectServiceMonitors(ctx, c.smonInfs.ListAllByNamespace)
	if err != nil {
		return nil, fmt.Errorf("selecting ServiceMonitors failed: %w", err)
	}

	pmons, err := resourceSelector.SelectPodMonitors(ctx, c.pmonInfs.ListAllByNamespace)
	if err != nil {
		return nil, fmt.Errorf("selecting PodMonitors failed: %w", err)
	}

	bmons, err := resourceSelector.SelectProbes(ctx, c.probeInfs.ListAllByNamespace)
	if err != nil {
		return nil, fmt.Errorf("selecting Probes failed: %w", err)
	}

	var scrapeConfigs operator.TypedResourcesSelection[*monitoringv1alpha1.ScrapeConfig]
	if c.sconInfs != nil {
		scrapeConfigs, err = resourceSelector.SelectScrapeConfigs(ctx, c.sconInfs.ListAllByNamespace)
		if err != nil {
			return nil, fmt.Errorf("selecting ScrapeConfigs failed: %w", err)
		}
	}

	rules, err := c.selectPrometheusRules(p, logger)
	if err != nil {
		return nil, fmt.Errorf("selecting PrometheusRule failed: %w", err)
	}

	return &selectedConfigResources{
		sMons:         smons,
		bMons:         bmons,
		pMons:         pmons,
		scrapeConfigs: scrapeConfigs,
		rules:         rules,
	}, nil
}

func (c *Operator) createOrUpdateConfigurationSecret(ctx context.Context, logger *slog.Logger, p *monitoringv1.Prometheus, cg *prompkg.ConfigGenerator, ruleConfigMapNames []string, store *assets.StoreBuilder, resources *selectedConfigResources) error {
	// If no service/pod monitor and probe selectors are configured, the user
	// wants to manage configuration themselves. Let's create an empty Secret
	// if it doesn't exist.
	if c.unmanagedPrometheusConfiguration(p) {

		s, err := prompkg.MakeConfigurationSecret(p, c.config, nil)
		if err != nil {
			return fmt.Errorf("failed to generate empty configuration secret: %w", err)
		}

		sClient := c.kclient.CoreV1().Secrets(p.Namespace)
		_, err = sClient.Get(ctx, s.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			logger.Debug("creating an empty configuration secret")
			if _, err := c.kclient.CoreV1().Secrets(p.Namespace).Create(ctx, s, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
				return fmt.Errorf("failed to create an empty configuration secret: %w", err)
			}

			return nil
		}

		return err
	}

	if err := prompkg.AddRemoteReadsToStore(ctx, store, p.GetNamespace(), p.Spec.RemoteRead); err != nil {
		return err
	}

	if err := cg.AddRemoteWriteToStore(ctx, store, p.GetNamespace(), p.Spec.RemoteWrite); err != nil {
		return err
	}

	if err := prompkg.AddAPIServerConfigToStore(ctx, store, p.GetNamespace(), p.Spec.APIServerConfig); err != nil {
		return err
	}

	if p.Spec.Alerting != nil {
		ams := p.Spec.Alerting.Alertmanagers

		for i, am := range ams {
			if err := validateAlertmanagerEndpoints(p, am); err != nil {
				return fmt.Errorf("alertmanager %d: %w", i, err)
			}
		}

		if err := addAlertmanagerEndpointsToStore(ctx, store, p.GetNamespace(), ams); err != nil {
			return err
		}
	}

	if err := prompkg.AddScrapeClassesToStore(ctx, store, p.GetNamespace(), p.Spec.ScrapeClasses); err != nil {
		return fmt.Errorf("failed to process scrape classes: %w", err)
	}

	sClient := c.kclient.CoreV1().Secrets(p.Namespace)
	additionalScrapeConfigs, err := k8s.LoadSecretRef(ctx, logger, sClient, p.Spec.AdditionalScrapeConfigs)
	if err != nil {
		return fmt.Errorf("loading additional scrape configs from Secret failed: %w", err)
	}
	additionalAlertRelabelConfigs, err := k8s.LoadSecretRef(ctx, logger, sClient, p.Spec.AdditionalAlertRelabelConfigs)
	if err != nil {
		return fmt.Errorf("loading additional alert relabel configs from Secret failed: %w", err)
	}
	additionalAlertManagerConfigs, err := k8s.LoadSecretRef(ctx, logger, sClient, p.Spec.AdditionalAlertManagerConfigs)
	if err != nil {
		return fmt.Errorf("loading additional alert manager configs from Secret failed: %w", err)
	}

	// Update secret based on the most recent configuration.
	conf, err := cg.GenerateServerConfiguration(
		p,
		resources.sMons.ValidResources(),
		resources.pMons.ValidResources(),
		resources.bMons.ValidResources(),
		resources.scrapeConfigs.ValidResources(),
		store,
		additionalScrapeConfigs,
		additionalAlertRelabelConfigs,
		additionalAlertManagerConfigs,
		ruleConfigMapNames,
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

func (c *Operator) createOrUpdateWebConfigSecret(ctx context.Context, p *monitoringv1.Prometheus) error {
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

func (c *Operator) createOrUpdateThanosConfigSecret(ctx context.Context, p *monitoringv1.Prometheus) error {
	secret, err := buildPrometheusHTTPClientConfigSecret(p)
	if err != nil {
		return fmt.Errorf("failed to build Thanos HTTP client config secret: :%w", err)
	}

	operator.UpdateObject(
		secret,
		operator.WithLabels(c.config.Labels),
		operator.WithAnnotations(c.config.Annotations),
		operator.WithManagingOwner(p),
	)

	return k8s.CreateOrUpdateSecret(ctx, c.kclient.CoreV1().Secrets(secret.Namespace), secret)
}

func makeSelectorLabels(name string) map[string]string {
	return map[string]string{
		operator.ManagedByLabelKey:           operator.ManagedByLabelValue,
		operator.ApplicationNameLabelKey:     applicationNameLabelValue,
		operator.ApplicationInstanceLabelKey: name,
		prompkg.PrometheusNameLabelName:      name,
		"prometheus":                         name,
	}
}

func validateAlertmanagerEndpoints(p *monitoringv1.Prometheus, am monitoringv1.AlertmanagerEndpoints) error {
	var nonNilFields []string

	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if am.BearerTokenFile != "" {
		nonNilFields = append(nonNilFields, fmt.Sprintf("%q", "bearerTokenFile"))
	}

	for k, v := range map[string]any{
		"basicAuth":     am.BasicAuth,
		"authorization": am.Authorization,
		"sigv4":         am.Sigv4,
	} {
		if reflect.ValueOf(v).IsNil() {
			continue
		}
		nonNilFields = append(nonNilFields, fmt.Sprintf("%q", k))
	}

	if len(nonNilFields) > 1 {
		return fmt.Errorf("%s can't be set at the same time, at most one of them must be defined", strings.Join(nonNilFields, " and "))
	}

	lcv, err := validation.NewLabelConfigValidator(p)
	if err != nil {
		return err
	}

	if err := lcv.Validate(am.RelabelConfigs); err != nil {
		return fmt.Errorf("invalid relabelings: %w", err)
	}

	if err := lcv.Validate(am.AlertRelabelConfigs); err != nil {
		return fmt.Errorf("invalid alertRelabelings: %w", err)
	}

	return nil
}

func addAlertmanagerEndpointsToStore(ctx context.Context, store *assets.StoreBuilder, namespace string, ams []monitoringv1.AlertmanagerEndpoints) error {
	for i, am := range ams {
		if err := store.AddBasicAuth(ctx, namespace, am.BasicAuth); err != nil {
			return fmt.Errorf("alertmanager %d: %w", i, err)
		}

		if err := store.AddSafeAuthorizationCredentials(ctx, namespace, am.Authorization); err != nil {
			return fmt.Errorf("alertmanager %d: %w", i, err)
		}

		if err := store.AddSigV4(ctx, namespace, am.Sigv4); err != nil {
			return fmt.Errorf("alertmanager %d: %w", i, err)
		}

		if err := store.AddTLSConfig(ctx, namespace, am.TLSConfig); err != nil {
			return fmt.Errorf("alertmanager %d: %w", i, err)
		}

		if err := store.AddProxyConfig(ctx, namespace, am.ProxyConfig); err != nil {
			return fmt.Errorf("alertmanager: %d: %w", i, err)
		}
	}

	return nil
}
