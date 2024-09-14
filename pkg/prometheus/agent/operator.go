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
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure"
	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/listwatch"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
)

const (
	resyncPeriod   = 5 * time.Minute
	controllerName = "prometheusagent-controller"
)

var prometheusAgentKeyInShardStatefulSet = regexp.MustCompile("^(.+)/prom-agent-(.+)-shard-[1-9][0-9]*$")
var prometheusAgentKey = regexp.MustCompile("^(.+)/prom-agent-(.+)$")

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

	endpointSliceSupported bool // Whether the Kubernetes API suports the EndpointSlice kind.
	scrapeConfigSupported  bool
	canReadStorageClass    bool

	eventRecorder record.EventRecorder

	statusReporter prompkg.StatusReporter

	daemonSetFeatureGateEnabled bool
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
		metrics:         operator.NewMetrics(r),
		reconciliations: &operator.ReconciliationTracker{},
		controllerID:    c.ControllerID,
		eventRecorder:   c.EventRecorderFactory(client, controllerName),
	}
	o.metrics.MustRegister(
		o.reconciliations,
	)
	for _, opt := range options {
		opt(o)
	}

	o.rr = operator.NewResourceReconciler(
		o.logger,
		o,
		o.metrics,
		monitoringv1alpha1.PrometheusAgentsKind,
		r,
		o.controllerID,
	)

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

	o.cmapInfs, err = informers.NewInformersForResourceWithTransform(
		informers.NewMetadataInformerFactory(
			c.Namespaces.PrometheusAllowList,
			c.Namespaces.DenyList,
			o.mdClient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = prompkg.LabelPrometheusName
			},
		),
		v1.SchemeGroupVersion.WithResource(string(v1.ResourceConfigMaps)),
		informers.PartialObjectMetadataStrip,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating configmap informers: %w", err)
	}

	o.secrInfs, err = informers.NewInformersForResourceWithTransform(
		informers.NewMetadataInformerFactory(
			c.Namespaces.PrometheusAllowList,
			c.Namespaces.DenyList,
			o.mdClient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.FieldSelector = c.SecretListWatchFieldSelector.String()
				options.LabelSelector = c.SecretListWatchLabelSelector.String()
			},
		),
		v1.SchemeGroupVersion.WithResource(string(v1.ResourceSecrets)),
		informers.PartialObjectMetadataStrip,
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
			nil,
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
				nil,
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
	_ = c.promInfs.ListAll(labels.Everything(), func(obj interface{}) {
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
	if err := c.promInfs.ListAll(labels.Everything(), func(o interface{}) {
		p := o.(*monitoringv1alpha1.PrometheusAgent)
		processFn(p, p.Status.Conditions)
	}); err != nil {
		c.logger.Error("failed to list PrometheusAgent objects", "err", err)
	}
}

// RefreshStatus implements the operator.StatusReconciler interface.
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
	))

	c.pmonInfs.AddEventHandler(operator.NewEventHandler(
		c.logger,
		c.accessor,
		c.metrics,
		monitoringv1.PodMonitorsKind,
		c.enqueueForMonitorNamespace,
	))

	c.probeInfs.AddEventHandler(operator.NewEventHandler(
		c.logger,
		c.accessor,
		c.metrics,
		monitoringv1.ProbesKind,
		c.enqueueForMonitorNamespace,
	))

	if c.sconInfs != nil {
		c.sconInfs.AddEventHandler(operator.NewEventHandler(
			c.logger,
			c.accessor,
			c.metrics,
			monitoringv1alpha1.ScrapeConfigsKind,
			c.enqueueForMonitorNamespace,
		))
	}

	c.cmapInfs.AddEventHandler(operator.NewEventHandler(
		c.logger,
		c.accessor,
		c.metrics,
		"ConfigMap",
		c.enqueueForPrometheusNamespace,
	))

	c.secrInfs.AddEventHandler(operator.NewEventHandler(
		c.logger,
		c.accessor,
		c.metrics,
		"Secret",
		c.enqueueForPrometheusNamespace,
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

// Resolve implements the operator.Syncer interface.
func (c *Operator) Resolve(obj interface{}) metav1.Object {
	key, ok := c.accessor.MetaNamespaceKey(obj)
	if !ok {
		return nil
	}

	var match bool
	var promKey string

	switch obj.(type) {
	case *appsv1.DaemonSet:
		match, promKey = daemonSetKeyToPrometheusAgentKey(key)
	case *appsv1.StatefulSet:
		match, promKey = statefulSetKeyToPrometheusAgentKey(key)
	}

	if !match {
		c.logger.Debug("StatefulSet key did not match a Prometheus Agent key format", "key", key)
		return nil
	}

	p, err := c.promInfs.Get(promKey)
	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		c.logger.Error("Prometheus lookup failed", "err", err)
		return nil
	}

	return p.(*monitoringv1alpha1.PrometheusAgent)
}

// statefulSetKeyToPrometheusAgentKey checks if StatefulSet key can be converted to Prometheus Agent key
// and do the conversion if it's true. The case of Prometheus Agent key in sharding is also handled.
func statefulSetKeyToPrometheusAgentKey(key string) (bool, string) {
	r := prometheusAgentKey
	if prometheusAgentKeyInShardStatefulSet.MatchString(key) {
		r = prometheusAgentKeyInShardStatefulSet
	}

	matches := r.FindAllStringSubmatch(key, 2)
	if len(matches) != 1 {
		return false, ""
	}
	if len(matches[0]) != 3 {
		return false, ""
	}
	return true, matches[0][1] + "/" + matches[0][2]
}

// daemonSetKeyToPrometheusAgentKey checks if DaemonSet key can be converted to Prometheus Agent key
// and do the conversion if it's true.
func daemonSetKeyToPrometheusAgentKey(key string) (bool, string) {
	r := prometheusAgentKey

	matches := r.FindAllStringSubmatch(key, 2)
	if len(matches) != 1 {
		return false, ""
	}
	if len(matches[0]) != 3 {
		return false, ""
	}
	return true, matches[0][1] + "/" + matches[0][2]
}

// Sync implements the operator.Syncer interface.
// TODO: Consider refactoring the common code between syncDaemonSet() and syncStatefulSet().
func (c *Operator) Sync(ctx context.Context, key string) error {
	pobj, err := c.promInfs.Get(key)

	if apierrors.IsNotFound(err) {
		c.reconciliations.ForgetObject(key)
		// Dependent resources are cleaned up by K8s via OwnerReferences
		return nil
	}
	if err != nil {
		return err
	}

	p := pobj.(*monitoringv1alpha1.PrometheusAgent)
	p = p.DeepCopy()
	if ptr.Deref(p.Spec.Mode, "StatefulSet") == "DaemonSet" {
		err = c.syncDaemonSet(ctx, key, p)
	} else {
		err = c.syncStatefulSet(ctx, key, p)
	}
	c.reconciliations.SetStatus(key, err)
	return err
}

func (c *Operator) syncDaemonSet(ctx context.Context, key string, p *monitoringv1alpha1.PrometheusAgent) error {
	if !c.daemonSetFeatureGateEnabled {
		return fmt.Errorf("feature gate for Prometheus Agent's DaemonSet mode is not enabled")
	}

	if err := k8sutil.AddTypeInformationToObject(p); err != nil {
		return fmt.Errorf("failed to set Prometheus type information: %w", err)
	}

	logger := c.logger.With("key", key)

	// Check if the Agent instance is marked for deletion.
	if c.rr.DeletionInProgress(p) {
		return nil
	}

	if p.Spec.Paused {
		logger.Info("the resource is paused, not reconciling")
		return nil
	}

	logger.Info("sync prometheus")

	opts := []prompkg.ConfigGeneratorOption{}
	if c.endpointSliceSupported {
		opts = append(opts, prompkg.WithEndpointSliceSupport())
	}
	cg, err := prompkg.NewConfigGenerator(c.logger, p, opts...)
	if err != nil {
		return err
	}

	assetStore := assets.NewStoreBuilder(c.kclient.CoreV1(), c.kclient.CoreV1())
	if err := c.createOrUpdateConfigurationSecret(ctx, p, cg, assetStore); err != nil {
		return fmt.Errorf("creating config failed: %w", err)
	}

	tlsAssets, err := operator.ReconcileShardedSecret(ctx, assetStore.TLSAssets(), c.kclient, prompkg.NewTLSAssetSecret(p, c.config))
	if err != nil {
		return fmt.Errorf("failed to reconcile the TLS secrets: %w", err)
	}

	if err := c.createOrUpdateWebConfigSecret(ctx, p); err != nil {
		return fmt.Errorf("synchronizing web config secret failed: %w", err)
	}

	dsetClient := c.kclient.AppsV1().DaemonSets(p.Namespace)

	logger.Debug("reconciling daemonset")

	_, err = c.dsetInfs.Get(keyToDaemonSetKey(p, key))
	exists := !apierrors.IsNotFound(err)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("retrieving daemonset failed: %w", err)
	}

	dset, err := makeDaemonSet(
		p,
		&c.config,
		cg,
		tlsAssets)
	if err != nil {
		return fmt.Errorf("making daemonset failed: %w", err)
	}

	if !exists {
		logger.Debug("no current daemonset found")
		logger.Debug("creating daemonset")
		if _, err := dsetClient.Create(ctx, dset, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("creating daemonset failed: %w", err)
		}

		logger.Info("daemonset successfully created")
		return nil
	}

	err = k8sutil.UpdateDaemonSet(ctx, dsetClient, dset)
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

func (c *Operator) syncStatefulSet(ctx context.Context, key string, p *monitoringv1alpha1.PrometheusAgent) error {
	if err := k8sutil.AddTypeInformationToObject(p); err != nil {
		return fmt.Errorf("failed to set Prometheus type information: %w", err)
	}

	logger := c.logger.With("key", key)

	// Check if the Agent instance is marked for deletion.
	if c.rr.DeletionInProgress(p) {
		return nil
	}

	if p.Spec.Paused {
		logger.Info("the resource is paused, not reconciling")
		return nil
	}

	logger.Info("sync prometheus")

	if err := operator.CheckStorageClass(ctx, c.canReadStorageClass, c.kclient, p.Spec.Storage); err != nil {
		return err
	}

	opts := []prompkg.ConfigGeneratorOption{}
	if c.endpointSliceSupported {
		opts = append(opts, prompkg.WithEndpointSliceSupport())
	}
	cg, err := prompkg.NewConfigGenerator(c.logger, p, opts...)
	if err != nil {
		return err
	}

	assetStore := assets.NewStoreBuilder(c.kclient.CoreV1(), c.kclient.CoreV1())
	if err := c.createOrUpdateConfigurationSecret(ctx, p, cg, assetStore); err != nil {
		return fmt.Errorf("creating config failed: %w", err)
	}

	tlsAssets, err := operator.ReconcileShardedSecret(ctx, assetStore.TLSAssets(), c.kclient, prompkg.NewTLSAssetSecret(p, c.config))
	if err != nil {
		return fmt.Errorf("failed to reconcile the TLS secrets: %w", err)
	}

	if err := c.createOrUpdateWebConfigSecret(ctx, p); err != nil {
		return fmt.Errorf("synchronizing web config secret failed: %w", err)
	}

	// Create governing service if it doesn't exist.
	svcClient := c.kclient.CoreV1().Services(p.Namespace)
	if err := k8sutil.CreateOrUpdateService(ctx, svcClient, makeStatefulSetService(p, c.config)); err != nil {
		return fmt.Errorf("synchronizing governing service failed: %w", err)
	}

	ssetClient := c.kclient.AppsV1().StatefulSets(p.Namespace)

	// Ensure we have a StatefulSet running Prometheus Agent deployed and that StatefulSet names are created correctly.
	expected := prompkg.ExpectedStatefulSetShardNames(p)
	for shard, ssetName := range expected {
		logger := logger.With("statefulset", ssetName, "shard", fmt.Sprintf("%d", shard))
		logger.Debug("reconciling statefulset")

		obj, err := c.ssetInfs.Get(prompkg.KeyToStatefulSetKey(p, key, shard))
		exists := !apierrors.IsNotFound(err)
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("retrieving statefulset failed: %w", err)
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
			&c.config,
			cg,
			newSSetInputHash,
			int32(shard),
			tlsAssets)
		if err != nil {
			return fmt.Errorf("making statefulset failed: %w", err)
		}
		operator.SanitizeSTS(sset)

		if !exists {
			logger.Debug("no current statefulset found")
			logger.Debug("creating statefulset")
			if _, err := ssetClient.Create(ctx, sset, metav1.CreateOptions{}); err != nil {
				return fmt.Errorf("creating statefulset failed: %w", err)
			}
			continue
		}

		if newSSetInputHash == existingStatefulSet.ObjectMeta.Annotations[operator.InputHashAnnotationName] {
			logger.Debug("new statefulset generation inputs match current, skipping any actions")
			continue
		}

		logger.Debug("updating current statefulset because of hash divergence",
			"new_hash", newSSetInputHash,
			"existing_hash", existingStatefulSet.ObjectMeta.Annotations[operator.InputHashAnnotationName],
		)

		err = k8sutil.UpdateStatefulSet(ctx, ssetClient, sset)
		sErr, ok := err.(*apierrors.StatusError)

		if ok && sErr.ErrStatus.Code == 422 && sErr.ErrStatus.Reason == metav1.StatusReasonInvalid {
			c.metrics.StsDeleteCreateCounter().Inc()

			// Gather only reason for failed update
			failMsg := make([]string, len(sErr.ErrStatus.Details.Causes))
			for i, cause := range sErr.ErrStatus.Details.Causes {
				failMsg[i] = cause.Message
			}

			logger.Info("recreating StatefulSet because the update operation wasn't possible", "reason", strings.Join(failMsg, ", "))

			propagationPolicy := metav1.DeletePropagationForeground
			if err := ssetClient.Delete(ctx, sset.GetName(), metav1.DeleteOptions{PropagationPolicy: &propagationPolicy}); err != nil {
				return fmt.Errorf("failed to delete StatefulSet to avoid forbidden action: %w", err)
			}
			continue
		}

		if err != nil {
			return fmt.Errorf("updating StatefulSet failed: %w", err)
		}
	}

	ssets := map[string]struct{}{}
	for _, ssetName := range expected {
		ssets[ssetName] = struct{}{}
	}

	err = c.ssetInfs.ListAllByNamespace(p.Namespace, labels.SelectorFromSet(labels.Set{prompkg.PrometheusNameLabelName: p.Name, prompkg.PrometheusModeLabeLName: prometheusMode}), func(obj interface{}) {
		s := obj.(*appsv1.StatefulSet)

		if _, ok := ssets[s.Name]; ok {
			// Do not delete statefulsets that we still expect to exist. This
			// is to cleanup StatefulSets when shards are reduced.
			return
		}

		if c.rr.DeletionInProgress(s) {
			return
		}

		propagationPolicy := metav1.DeletePropagationForeground
		if err := ssetClient.Delete(ctx, s.GetName(), metav1.DeleteOptions{PropagationPolicy: &propagationPolicy}); err != nil {
			c.logger.Error("", "err", err, "name", s.GetName(), "namespace", s.GetNamespace())
		}
	})
	if err != nil {
		return fmt.Errorf("listing StatefulSet resources failed: %w", err)
	}

	return nil
}

func (c *Operator) createOrUpdateConfigurationSecret(ctx context.Context, p *monitoringv1alpha1.PrometheusAgent, cg *prompkg.ConfigGenerator, store *assets.StoreBuilder) error {
	resourceSelector, err := prompkg.NewResourceSelector(c.logger, p, store, c.nsMonInf, c.metrics, c.eventRecorder)
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

	var scrapeConfigs map[string]*monitoringv1alpha1.ScrapeConfig
	if c.sconInfs != nil {
		scrapeConfigs, err = resourceSelector.SelectScrapeConfigs(ctx, c.sconInfs.ListAllByNamespace)
		if err != nil {
			return fmt.Errorf("selecting ScrapeConfigs failed: %w", err)
		}
	}

	if err := prompkg.AddRemoteWritesToStore(ctx, store, p.GetNamespace(), p.Spec.RemoteWrite); err != nil {
		return err
	}

	if err := prompkg.AddAPIServerConfigToStore(ctx, store, p.GetNamespace(), p.Spec.APIServerConfig); err != nil {
		return err
	}

	if err := prompkg.AddScrapeClassesToStore(ctx, store, p.GetNamespace(), p.Spec.ScrapeClasses); err != nil {
		return fmt.Errorf("failed to process scrape classes: %w", err)
	}

	sClient := c.kclient.CoreV1().Secrets(p.Namespace)
	additionalScrapeConfigs, err := k8sutil.LoadSecretRef(ctx, c.logger, sClient, p.Spec.AdditionalScrapeConfigs)
	if err != nil {
		return fmt.Errorf("loading additional scrape configs from Secret failed: %w", err)
	}

	// Update secret based on the most recent configuration.
	conf, err := cg.GenerateAgentConfiguration(
		smons,
		pmons,
		bmons,
		scrapeConfigs,
		p.Spec.TSDB,
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

	c.logger.Debug("updating Prometheus configuration secret")
	return k8sutil.CreateOrUpdateSecret(ctx, sClient, s)
}

func createSSetInputHash(p monitoringv1alpha1.PrometheusAgent, c prompkg.Config, tlsAssets *operator.ShardedSecret, ssSpec appsv1.StatefulSetSpec) (string, error) {
	var http2 *bool
	if p.Spec.Web != nil && p.Spec.Web.WebConfigFileFields.HTTPConfig != nil {
		http2 = p.Spec.Web.WebConfigFileFields.HTTPConfig.HTTP2
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
	pobj, err := c.promInfs.Get(key)

	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	p := pobj.(*monitoringv1alpha1.PrometheusAgent)
	p = p.DeepCopy()

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

	if _, err = c.mclient.MonitoringV1alpha1().PrometheusAgents(p.Namespace).ApplyStatus(ctx, prompkg.ApplyConfigurationFromPrometheusAgent(p, true), metav1.ApplyOptions{FieldManager: operator.PrometheusOperatorFieldManager, Force: true}); err != nil {
		c.logger.Info("failed to apply prometheus status subresource, trying again without scale fields", "err", err)
		// Try again, but this time does not update scale subresource.
		if _, err = c.mclient.MonitoringV1alpha1().PrometheusAgents(p.Namespace).ApplyStatus(ctx, prompkg.ApplyConfigurationFromPrometheusAgent(p, false), metav1.ApplyOptions{FieldManager: operator.PrometheusOperatorFieldManager, Force: true}); err != nil {
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
	nsObject, exists, err := store.GetByKey(nsName)
	if err != nil {
		c.logger.Error(
			"get namespace to enqueue Prometheus instances failed",
			"err", err,
		)
		return
	}
	if !exists {
		c.logger.Error(fmt.Sprintf("get namespace to enqueue Prometheus instances failed: namespace %q does not exist", nsName))
		return
	}
	ns := nsObject.(*v1.Namespace)

	err = c.promInfs.ListAll(labels.Everything(), func(obj interface{}) {
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

		c.logger.Info("we are gonna check if it Matches")

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

func (c *Operator) handleMonitorNamespaceUpdate(oldo, curo interface{}) {
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
	// and Probes in the namespace.
	err := c.promInfs.ListAll(labels.Everything(), func(obj interface{}) {
		p := obj.(*monitoringv1alpha1.PrometheusAgent)

		for name, selector := range map[string]*metav1.LabelSelector{
			"PodMonitors":     p.Spec.PodMonitorNamespaceSelector,
			"Probes":          p.Spec.ProbeNamespaceSelector,
			"ServiceMonitors": p.Spec.ServiceMonitorNamespaceSelector,
		} {

			sync, err := k8sutil.LabelSelectionHasChanged(old.Labels, cur.Labels, selector)
			if err != nil {
				c.logger.Error(
					"",
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

func ListOptions(name string) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
			"app.kubernetes.io/name":       "prometheus-agent",
			"app.kubernetes.io/managed-by": "prometheus-operator",
			"app.kubernetes.io/instance":   name,
		})).String(),
	}
}

func makeSelectorLabels(name string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":        "prometheus-agent",
		"app.kubernetes.io/managed-by":  "prometheus-operator",
		"app.kubernetes.io/instance":    name,
		prompkg.PrometheusNameLabelName: name,
	}
}

func keyToDaemonSetKey(p monitoringv1.PrometheusInterface, key string) string {
	keyParts := strings.Split(key, "/")
	return fmt.Sprintf("%s/%s", keyParts[0], fmt.Sprintf("%s-%s", prompkg.Prefix(p), keyParts[1]))
}
