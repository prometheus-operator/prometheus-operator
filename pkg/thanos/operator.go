// Copyright 2020 The prometheus-operator Authors
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

package thanos

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/mitchellh/hashstructure"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	monitoringv1ac "github.com/prometheus-operator/prometheus-operator/pkg/client/applyconfiguration/monitoring/v1"
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
	applicationNameLabelValue = "thanos-ruler"
	controllerName            = "thanos-controller"
	rwConfigFile              = "remote-write.yaml"

	noSelectedResourcesMessage = "No PrometheusRule have been selected."
)

var minRemoteWriteVersion = semver.MustParse("0.24.0")

// Operator manages life cycle of Thanos deployments and
// monitoring configurations.
type Operator struct {
	kclient  kubernetes.Interface
	dclient  dynamic.Interface
	mdClient metadata.Interface
	mclient  monitoringclient.Interface

	logger   *slog.Logger
	accessor *operator.Accessor

	controllerID string

	thanosRulerInfs *informers.ForResource
	cmapInfs        *informers.ForResource
	ruleInfs        *informers.ForResource
	ssetInfs        *informers.ForResource

	rr *operator.ResourceReconciler

	nsThanosRulerInf cache.SharedIndexInformer
	nsRuleInf        cache.SharedIndexInformer

	metrics             *operator.Metrics
	reconciliations     *operator.ReconciliationTracker
	canReadStorageClass bool

	newEventRecorder operator.NewEventRecorderFunc

	config Config

	configResourcesStatusEnabled bool

	finalizerSyncer *operator.FinalizerSyncer
}

// Config defines the operator's parameters for the Thanos controller.
// Whenever the value of one of these parameters is changed, it triggers an
// update of the managed statefulsets.
type Config struct {
	LocalHost              string
	ReloaderConfig         operator.ContainerConfig
	ThanosDefaultBaseImage string
	Annotations            operator.Map
	Labels                 operator.Map
}

type ControllerOption func(*Operator)

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

	// All the metrics exposed by the controller get the controller="thanos" label.
	r = prometheus.WrapRegistererWith(prometheus.Labels{"controller": "thanos"}, r)

	o := &Operator{
		kclient:          client,
		dclient:          dclient,
		mdClient:         mdClient,
		mclient:          mclient,
		logger:           logger,
		accessor:         operator.NewAccessor(logger),
		metrics:          operator.NewMetrics(r),
		newEventRecorder: c.EventRecorderFactory(client, controllerName),
		reconciliations:  &operator.ReconciliationTracker{},
		controllerID:     c.ControllerID,
		config: Config{
			ReloaderConfig:         c.ReloaderConfig,
			ThanosDefaultBaseImage: c.ThanosDefaultBaseImage,
			Annotations:            c.Annotations,
			Labels:                 c.Labels,
			LocalHost:              c.LocalHost,
		},
		finalizerSyncer: operator.NewNoopFinalizerSyncer(),
	}
	for _, opt := range options {
		opt(o)
	}

	if o.configResourcesStatusEnabled {
		o.finalizerSyncer = operator.NewFinalizerSyncer(mdClient, monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ThanosRulerName))
	}

	o.cmapInfs, err = informers.NewInformersForResource(
		informers.NewMetadataInformerFactory(
			c.Namespaces.ThanosRulerAllowList,
			c.Namespaces.DenyList,
			o.mdClient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = labelThanosRulerName
			},
		),
		v1.SchemeGroupVersion.WithResource(string(v1.ResourceConfigMaps)),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating configmap informers: %w", err)
	}

	o.thanosRulerInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.Namespaces.ThanosRulerAllowList,
			c.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = c.ThanosRulerSelector.String()
			},
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ThanosRulerName),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating thanosruler informers: %w", err)
	}

	var thanosStores []cache.Store
	for _, informer := range o.thanosRulerInfs.GetInformers() {
		thanosStores = append(thanosStores, informer.Informer().GetStore())
	}
	o.metrics.MustRegister(newThanosRulerCollectorForStores(thanosStores...))

	o.rr = operator.NewResourceReconciler(
		o.logger,
		o,
		o.thanosRulerInfs,
		o.metrics,
		monitoringv1.ThanosRulerKind,
		r,
		o.controllerID,
	)

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

	o.ssetInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			c.Namespaces.ThanosRulerAllowList,
			c.Namespaces.DenyList,
			o.kclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = labelSelectorForStatefulSets()
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
			c.Namespaces.DenyList)
		if err != nil {
			return nil, err
		}

		o.logger.Debug("creating namespace informer", "privileged", privileged)
		return cache.NewSharedIndexInformer(
			o.metrics.NewInstrumentedListerWatcher(lw),
			&v1.Namespace{},
			resyncPeriod,
			cache.Indexers{},
		), nil
	}

	o.nsRuleInf, err = newNamespaceInformer(o, c.Namespaces.AllowList)
	if err != nil {
		return nil, err
	}

	if listwatch.IdenticalNamespaces(c.Namespaces.AllowList, c.Namespaces.ThanosRulerAllowList) {
		o.nsThanosRulerInf = o.nsRuleInf
	} else {
		o.nsThanosRulerInf, err = newNamespaceInformer(o, c.Namespaces.ThanosRulerAllowList)
		if err != nil {
			return nil, err
		}
	}

	return o, nil
}

// waitForCacheSync waits for the informers' caches to be synced.
func (o *Operator) waitForCacheSync(ctx context.Context) error {
	for _, infs := range []struct {
		name                 string
		informersForResource *informers.ForResource
	}{
		{"ThanosRuler", o.thanosRulerInfs},
		{"ConfigMap", o.cmapInfs},
		{"PrometheusRule", o.ruleInfs},
		{"StatefulSet", o.ssetInfs},
	} {
		for _, inf := range infs.informersForResource.GetInformers() {
			if !operator.WaitForNamedCacheSync(ctx, "thanos", o.logger.With("informer", infs.name), inf.Informer()) {
				return fmt.Errorf("failed to sync cache for %s informer", infs.name)
			}
		}
	}

	for _, inf := range []struct {
		name     string
		informer cache.SharedIndexInformer
	}{
		{"ThanosRulerNamespace", o.nsThanosRulerInf},
		{"RuleNamespace", o.nsRuleInf},
	} {
		if !operator.WaitForNamedCacheSync(ctx, "thanos", o.logger.With("informer", inf.name), inf.informer) {
			return fmt.Errorf("failed to sync cache for %s informer", inf.name)
		}
	}

	o.logger.Info("successfully synced all caches")
	return nil
}

// addHandlers adds the eventhandlers to the informers.
func (o *Operator) addHandlers() {
	o.thanosRulerInfs.AddEventHandler(o.rr)
	o.ssetInfs.AddEventHandler(o.rr)

	o.cmapInfs.AddEventHandler(operator.NewEventHandler(
		o.logger,
		o.accessor,
		o.metrics,
		operator.ConfigMapGVK().Kind,
		o.enqueueForThanosRulerNamespace,
		operator.WithFilter(operator.ResourceVersionChanged),
	))

	o.ruleInfs.AddEventHandler(operator.NewEventHandler(
		o.logger,
		o.accessor,
		o.metrics,
		monitoringv1.PrometheusRuleKind,
		o.enqueueForRulesNamespace,
		operator.WithFilter(
			operator.AnyFilter(
				operator.GenerationChanged,
				operator.LabelsChanged,
			),
		),
	))

	// The controller needs to watch the namespaces in which the rules live
	// because a label change on a namespace may trigger a configuration
	// change.
	// It doesn't need to watch on addition/deletion though because it's
	// already covered by the event handlers on rules.
	_, _ = o.nsRuleInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: o.handleNamespaceUpdate,
	})
}

// Run the controller.
func (o *Operator) Run(ctx context.Context) error {
	go o.rr.Run(ctx)
	defer o.rr.Stop()

	go o.thanosRulerInfs.Start(ctx.Done())
	go o.cmapInfs.Start(ctx.Done())
	go o.ruleInfs.Start(ctx.Done())
	go o.nsRuleInf.Run(ctx.Done())
	if o.nsRuleInf != o.nsThanosRulerInf {
		go o.nsThanosRulerInf.Run(ctx.Done())
	}
	go o.ssetInfs.Start(ctx.Done())
	if err := o.waitForCacheSync(ctx); err != nil {
		return err
	}

	// Refresh the status of the existing ThanosRuler objects.
	_ = o.thanosRulerInfs.ListAll(labels.Everything(), func(obj any) {
		o.rr.EnqueueForStatus(obj.(*monitoringv1.ThanosRuler))
	})

	o.addHandlers()

	// TODO(simonpasquier): watch for ThanosRuler pods instead of polling.
	go operator.StatusPoller(ctx, o)

	o.metrics.Ready().Set(1)
	<-ctx.Done()
	return nil
}

// Iterate implements the operator.StatusReconciler interface.
func (o *Operator) Iterate(processFn func(operator.StatusGetter)) {
	if err := o.thanosRulerInfs.ListAll(labels.Everything(), func(o any) {
		processFn(o.(*monitoringv1.ThanosRuler))
	}); err != nil {
		o.logger.Error("failed to list ThanosRuler objects", "err", err)
	}
}

// RefreshStatusFor implements the operator.StatusReconciler interface.
func (o *Operator) RefreshStatusFor(obj metav1.Object) {
	o.rr.EnqueueForStatus(obj)
}

func thanosKeyToStatefulSetKey(key string) string {
	keyParts := strings.Split(key, "/")
	return keyParts[0] + "/thanos-ruler-" + keyParts[1]
}

func (o *Operator) handleNamespaceUpdate(oldo, curo any) {
	old := oldo.(*v1.Namespace)
	cur := curo.(*v1.Namespace)

	o.logger.Debug("update handler", "namespace", cur.GetName(), "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	// Periodic resync may resend the Namespace without changes in-between.
	if old.ResourceVersion == cur.ResourceVersion {
		return
	}

	o.logger.Debug("Namespace updated", "namespace", cur.GetName())
	o.metrics.TriggerByCounter("Namespace", operator.UpdateEvent).Inc()

	// Check for ThanosRuler instances selecting PrometheusRules in the namespace.
	err := o.thanosRulerInfs.ListAll(labels.Everything(), func(obj any) {
		tr := obj.(*monitoringv1.ThanosRuler)

		sync, err := k8s.LabelSelectionHasChanged(old.Labels, cur.Labels, tr.Spec.RuleNamespaceSelector)
		if err != nil {
			o.logger.Error(
				"failed to detect label selection change",
				"err", err,
				"name", tr.Name,
				"namespace", tr.Namespace,
			)
			return
		}

		if sync {
			o.rr.EnqueueForReconciliation(tr)
		}
	})
	if err != nil {
		o.logger.Error("listing all ThanosRuler instances from cache failed",
			"err", err,
		)
	}
}

// Sync implements the operator.Syncer interface.
func (o *Operator) Sync(ctx context.Context, key string) error {
	o.reconciliations.ResetStatus(key)
	err := o.sync(ctx, key)
	o.reconciliations.SetStatus(key, err)

	return err
}

func (o *Operator) sync(ctx context.Context, key string) error {
	tr, err := operator.GetObjectFromKey[*monitoringv1.ThanosRuler](o.thanosRulerInfs, key)
	if err != nil {
		return err
	}

	if tr == nil {
		o.reconciliations.ForgetObject(key)
		// Dependent resources are cleaned up by K8s via OwnerReferences
		return nil
	}

	logger := o.logger.With("key", key)
	logger.Info("sync thanos-ruler")

	finalizerAdded, err := o.finalizerSyncer.Sync(ctx, tr, o.rr.DeletionInProgress(tr), func() error {
		return o.configResStatusCleanup(ctx, tr)
	})
	if err != nil {
		return err
	}

	if finalizerAdded {
		// Since the finalizer has been added to the object, let's trigger another sync.
		o.rr.EnqueueForReconciliation(tr)
		return nil
	}

	// Check if the Thanos instance is marked for deletion.
	if o.rr.DeletionInProgress(tr) {
		o.reconciliations.ForgetObject(key)
		return nil
	}

	if tr.Spec.Paused {
		logger.Info("no action taken (the resource is paused)")
		return nil
	}

	o.recordDeprecatedFields(key, logger, tr)

	if err := operator.CheckStorageClass(ctx, o.canReadStorageClass, o.kclient, tr.Spec.Storage); err != nil {
		return err
	}

	selectedRules, err := o.selectPrometheusRules(tr, logger)
	if err != nil {
		return err
	}

	if selectedRules.SelectedLen() == 0 {
		o.reconciliations.SetReasonAndMessage(key, operator.NoSelectedResourcesReason, noSelectedResourcesMessage)
	}

	ruleConfigMapNames, err := o.createOrUpdateRuleConfigMaps(ctx, tr, selectedRules, logger)
	if err != nil {
		return err
	}

	assetStore := assets.NewStoreBuilder(o.kclient.CoreV1(), o.kclient.CoreV1())

	if err := o.createOrUpdateRulerConfigSecret(ctx, assetStore, tr); err != nil {
		return fmt.Errorf("failed to synchronize ruler config secret: %w", err)
	}

	tlsAssets, err := operator.ReconcileShardedSecret(ctx, assetStore.TLSAssets(), o.kclient, newTLSAssetSecret(tr, o.config))
	if err != nil {
		return fmt.Errorf("failed to reconcile the TLS secrets: %w", err)
	}

	if err := o.createOrUpdateWebConfigSecret(ctx, tr); err != nil {
		return fmt.Errorf("failed to synchronize web config secret: %w", err)
	}

	svcClient := o.kclient.CoreV1().Services(tr.Namespace)
	if tr.Spec.ServiceName != nil {
		selectorLabels := makeSelectorLabels(tr.Name)
		if err := k8s.EnsureCustomGoverningService(ctx, tr.Namespace, *tr.Spec.ServiceName, svcClient, selectorLabels); err != nil {
			return err
		}
	} else {
		// Create governing service if it doesn't exist.
		if _, err = k8s.CreateOrUpdateService(ctx, svcClient, makeStatefulSetService(tr, o.config)); err != nil {
			return fmt.Errorf("synchronizing governing service failed: %w", err)
		}
	}

	// Ensure we have a StatefulSet running Thanos deployed.
	existingStatefulSet, err := o.getStatefulSetFromThanosRulerKey(key)
	if err != nil {
		return err
	}

	shouldCreate := false
	if existingStatefulSet == nil {
		shouldCreate = true
		existingStatefulSet = &appsv1.StatefulSet{}
	}

	if o.rr.DeletionInProgress(existingStatefulSet) {
		return nil
	}

	newSSetInputHash, err := createSSetInputHash(*tr, o.config, tlsAssets, ruleConfigMapNames, existingStatefulSet.Spec)
	if err != nil {
		return err
	}

	sset, err := makeStatefulSet(tr, o.config, ruleConfigMapNames, newSSetInputHash, tlsAssets)
	if err != nil {
		return fmt.Errorf("failed to generate statefulset: %w", err)
	}

	operator.SanitizeSTS(sset)

	// Update the status of selected configuration resources (PrometheusRules).
	// This must be called before the StatefulSet creation/update to ensure
	// config resource bindings are updated on first reconciliation.
	if err = o.updateConfigResourcesStatus(ctx, tr, selectedRules); err != nil {
		return err
	}

	ssetClient := o.kclient.AppsV1().StatefulSets(tr.Namespace)
	if shouldCreate {
		logger.Debug("creating statefulset")
		if _, err := k8s.CreateStatefulSetOrPatchLabels(ctx, ssetClient, sset); err != nil {
			return fmt.Errorf("failed to create thanos statefulset: %w", err)
		}

		return nil
	}

	if newSSetInputHash == existingStatefulSet.Annotations[operator.InputHashAnnotationKey] {
		logger.Debug("new statefulset generation inputs match current, skipping any actions", "hash", newSSetInputHash)
		return nil
	}

	logger.Debug("new hash differs from the existing value", "new", newSSetInputHash, "existing", existingStatefulSet.Annotations[operator.InputHashAnnotationKey])
	if err = k8s.ForceUpdateStatefulSet(ctx, ssetClient, sset, func(reason string) {
		o.metrics.StsDeleteCreateCounter().Inc()
		logger.Info("recreating StatefulSet because the update operation wasn't possible", "reason", reason)
	}); err != nil {
		return err
	}

	return nil
}

func (o *Operator) recordDeprecatedFields(key string, logger *slog.Logger, tr *monitoringv1.ThanosRuler) {
	deprecationWarningf := "field %q is deprecated, field %q should be used instead"
	var deprecations []string

	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if len(tr.Spec.PrometheusRulesExcludedFromEnforce) > 0 {
		deprecations = append(deprecations, fmt.Sprintf(deprecationWarningf, "spec.prometheusRulesExcludedFromEnforce", "spec.excludedFromEnforcement"))
	}

	if len(deprecations) > 0 {
		for _, m := range deprecations {
			logger.Warn(m)
		}
		o.reconciliations.SetReasonAndMessage(key, operator.DeprecatedFieldsInUseReason, strings.Join(deprecations, "; "))
	}
}

// updateConfigResourcesStatus updates the status of the selected configuration
// resources (PrometheusRules).
func (o *Operator) updateConfigResourcesStatus(ctx context.Context, tr *monitoringv1.ThanosRuler, rules operator.PrometheusRuleSelection) error {
	if !o.configResourcesStatusEnabled {
		return nil
	}

	var configResourceSyncer = operator.NewConfigResourceSyncer(tr, o.dclient, o.accessor)

	for key, configResource := range rules.Selected() {
		if err := configResourceSyncer.UpdateBinding(ctx, configResource.Resource(), configResource.Conditions()); err != nil {
			return fmt.Errorf("failed to update PrometheusRule %s status: %w", key, err)
		}
	}

	if err := operator.CleanupBindings(ctx, o.ruleInfs.ListAll, rules.Selected(), configResourceSyncer); err != nil {
		return fmt.Errorf("failed to remove bindings for prometheusRules: %w", err)
	}
	return nil
}

// configResStatusCleanup removes thanosRuler bindings from the configuration resources (PrometheusRule).
func (o *Operator) configResStatusCleanup(ctx context.Context, tr *monitoringv1.ThanosRuler) error {
	if !o.configResourcesStatusEnabled {
		return nil
	}

	var configResourceSyncer = operator.NewConfigResourceSyncer(tr, o.dclient, o.accessor)

	if err := operator.CleanupBindings(ctx, o.ruleInfs.ListAll, operator.TypedResourcesSelection[*monitoringv1.PrometheusRule]{}, configResourceSyncer); err != nil {
		return fmt.Errorf("failed to remove bindings for prometheusRule: %w", err)
	}
	return nil
}

// getStatefulSetFromThanosRulerKey returns a copy of the StatefulSet object
// corresponding to the ThanosRuler object identified by key.
// If the object is not found, it returns a nil pointer without error.
func (o *Operator) getStatefulSetFromThanosRulerKey(key string) (*appsv1.StatefulSet, error) {
	ssetName := thanosKeyToStatefulSetKey(key)

	obj, err := o.ssetInfs.Get(ssetName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			o.logger.Info("StatefulSet not found", "key", ssetName)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve StatefulSet from informer: %w", err)
	}

	return obj.(*appsv1.StatefulSet).DeepCopy(), nil
}

// UpdateStatus implements the operator.Syncer interface.
func (o *Operator) UpdateStatus(ctx context.Context, key string) error {
	tr, err := operator.GetObjectFromKey[*monitoringv1.ThanosRuler](o.thanosRulerInfs, key)
	if err != nil {
		return err
	}

	if tr == nil {
		return nil
	}

	if o.rr.DeletionInProgress(tr) {
		return nil
	}

	sset, err := o.getStatefulSetFromThanosRulerKey(key)
	if err != nil {
		return fmt.Errorf("failed to get StatefulSet: %w", err)
	}

	if sset != nil && o.rr.DeletionInProgress(sset) {
		return nil
	}

	stsReporter, err := operator.NewStatefulSetReporter(ctx, o.kclient, sset)
	if err != nil {
		return fmt.Errorf("failed to retrieve statefulset state: %w", err)
	}

	availableCondition := stsReporter.Update(tr)
	reconciledCondition := o.reconciliations.GetCondition(key, tr.Generation)
	tr.Status.Conditions = operator.UpdateConditions(tr.Status.Conditions, availableCondition, reconciledCondition)
	tr.Status.Paused = tr.Spec.Paused

	if _, err = o.mclient.MonitoringV1().ThanosRulers(tr.Namespace).ApplyStatus(ctx, applyConfigurationFromThanosRuler(tr), metav1.ApplyOptions{FieldManager: k8s.PrometheusOperatorFieldManager, Force: true}); err != nil {
		return fmt.Errorf("failed to apply status subresource: %w", err)
	}

	return nil
}

func createSSetInputHash(tr monitoringv1.ThanosRuler, c Config, tlsAssets *operator.ShardedSecret, ruleConfigMapNames []string, ss appsv1.StatefulSetSpec) (string, error) {

	// The controller should ignore any changes to RevisionHistoryLimit field because
	// it may be modified by external actors.
	// See https://github.com/prometheus-operator/prometheus-operator/issues/5712
	ss.RevisionHistoryLimit = nil

	hash, err := hashstructure.Hash(struct {
		ThanosRulerLabels      map[string]string
		ThanosRulerAnnotations map[string]string
		ThanosRulerGeneration  int64
		Config                 Config
		StatefulSetSpec        appsv1.StatefulSetSpec
		RuleConfigMaps         []string `hash:"set"`
		ShardedSecret          *operator.ShardedSecret
	}{
		ThanosRulerLabels:      tr.Labels,
		ThanosRulerAnnotations: tr.Annotations,
		ThanosRulerGeneration:  tr.Generation,
		Config:                 c,
		StatefulSetSpec:        ss,
		RuleConfigMaps:         ruleConfigMapNames,
		ShardedSecret:          tlsAssets,
	},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to calculate combined hash: %w", err)
	}

	return fmt.Sprintf("%d", hash), nil
}

func (o *Operator) enqueueForThanosRulerNamespace(nsName string) {
	o.enqueueForNamespace(o.nsThanosRulerInf.GetStore(), nsName)
}

func (o *Operator) enqueueForRulesNamespace(nsName string) {
	o.enqueueForNamespace(o.nsRuleInf.GetStore(), nsName)
}

// enqueueForNamespace enqueues all ThanosRuler object keys that belong to the
// given namespace or select objects in the given namespace.
func (o *Operator) enqueueForNamespace(store cache.Store, nsName string) {
	nsObject, exists, err := store.GetByKey(nsName)
	if err != nil {
		o.logger.Error("get namespace to enqueue ThanosRuler instances failed",
			"err", err,
		)
		return
	}
	if !exists {
		o.logger.Error("get namespace to enqueue ThanosRuler instances failed: namespace does not exist",
			"namespace", nsName,
		)
		return
	}
	ns := nsObject.(*v1.Namespace)

	err = o.thanosRulerInfs.ListAll(labels.Everything(), func(obj any) {
		// Check for ThanosRuler instances in the namespace.
		tr := obj.(*monitoringv1.ThanosRuler)
		if tr.Namespace == nsName {
			o.rr.EnqueueForReconciliation(tr)
			return
		}

		// Check for ThanosRuler instances selecting PrometheusRules in
		// the namespace.
		ruleNSSelector, err := metav1.LabelSelectorAsSelector(tr.Spec.RuleNamespaceSelector)
		if err != nil {
			o.logger.Error("",
				"err", fmt.Errorf("failed to convert RuleNamespaceSelector: %w", err),
				"name", tr.Name,
				"namespace", tr.Namespace,
				"selector", tr.Spec.RuleNamespaceSelector,
			)
			return
		}

		if ruleNSSelector.Matches(labels.Set(ns.Labels)) {
			o.rr.EnqueueForReconciliation(tr)
			return
		}
	})
	if err != nil {
		o.logger.Error("listing all ThanosRuler instances from cache failed",
			"err", err,
		)
	}
}

func (o *Operator) createOrUpdateWebConfigSecret(ctx context.Context, tr *monitoringv1.ThanosRuler) error {
	var fields monitoringv1.WebConfigFileFields
	if tr.Spec.Web != nil {
		fields = tr.Spec.Web.WebConfigFileFields
	}

	webConfig, err := webconfig.New(
		webConfigDir,
		webConfigSecretName(tr.Name),
		fields,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize the web config: %w", err)
	}

	s := &v1.Secret{}
	operator.UpdateObject(
		s,
		operator.WithLabels(o.config.Labels),
		operator.WithAnnotations(o.config.Annotations),
		operator.WithManagingOwner(tr),
	)

	if err := webConfig.CreateOrUpdateWebConfigSecret(ctx, o.kclient.CoreV1().Secrets(tr.Namespace), s); err != nil {
		return fmt.Errorf("failed to update the web config secret: %w", err)
	}

	return nil
}

func applyConfigurationFromThanosRuler(a *monitoringv1.ThanosRuler) *monitoringv1ac.ThanosRulerApplyConfiguration {
	trac := monitoringv1ac.ThanosRulerStatus().
		WithPaused(a.Status.Paused).
		WithReplicas(a.Status.Replicas).
		WithAvailableReplicas(a.Status.AvailableReplicas).
		WithUpdatedReplicas(a.Status.UpdatedReplicas).
		WithUnavailableReplicas(a.Status.UnavailableReplicas)

	for _, condition := range a.Status.Conditions {
		trac.WithConditions(
			monitoringv1ac.Condition().
				WithType(condition.Type).
				WithStatus(condition.Status).
				WithLastTransitionTime(condition.LastTransitionTime).
				WithReason(condition.Reason).
				WithMessage(condition.Message).
				WithObservedGeneration(condition.ObservedGeneration),
		)
	}

	return monitoringv1ac.ThanosRuler(a.Name, a.Namespace).WithStatus(trac)
}

func newTLSAssetSecret(tr *monitoringv1.ThanosRuler, config Config) *v1.Secret {
	s := &v1.Secret{
		Data: map[string][]byte{},
	}

	operator.UpdateObject(
		s,
		operator.WithLabels(config.Labels),
		operator.WithAnnotations(config.Annotations),
		operator.WithManagingOwner(tr),
		operator.WithName(tlsAssetsSecretName(tr.Name)),
		operator.WithNamespace(tr.GetObjectMeta().GetNamespace()),
	)

	return s
}

// In cases where an existing selector label is modified, or a new one is added, new sts cannot match existing pods.
// We should try to avoid removing such immutable fields whenever possible since doing
// so forces us to enter the 'recreate cycle' and can potentially lead to downtime.
// The requirement to make a change here should be carefully evaluated.
func makeSelectorLabels(name string) map[string]string {
	return map[string]string{
		operator.ApplicationNameLabelKey:     applicationNameLabelValue,
		operator.ManagedByLabelKey:           operator.ManagedByLabelValue,
		operator.ApplicationInstanceLabelKey: name,
		"thanos-ruler":                       name,
	}
}

// labelSelectorForStatefulSets returns a label selector which selects
// all ThanosRuler statefulsets.
func labelSelectorForStatefulSets() string {
	return fmt.Sprintf(
		"%s in (%s),%s in (%s)",
		operator.ManagedByLabelKey, operator.ManagedByLabelValue,
		operator.ApplicationNameLabelKey, applicationNameLabelValue,
	)
}

func (o *Operator) createOrUpdateRulerConfigSecret(ctx context.Context, store *assets.StoreBuilder, tr *monitoringv1.ThanosRuler) error {
	sClient := o.kclient.CoreV1().Secrets(tr.GetNamespace())

	s := &v1.Secret{
		Data: map[string][]byte{},
	}

	operator.UpdateObject(
		s,
		operator.WithName(rulerConfigSecretName(tr.Name)),
		operator.WithAnnotations(o.config.Annotations),
		operator.WithLabels(o.config.Labels),
		operator.WithOwner(tr),
	)

	thanosVersion := operator.StringValOrDefault(ptr.Deref(tr.Spec.Version, ""), operator.DefaultThanosVersion)
	version, err := semver.ParseTolerant(thanosVersion)
	if err != nil {
		return fmt.Errorf("failed to parse Thanos Ruler version %q: %w", thanosVersion, err)
	}

	if len(tr.Spec.RemoteWrite) > 0 {
		if version.LT(minRemoteWriteVersion) {
			return fmt.Errorf("thanos remote-write configuration requires at least version %q: current version %q", minRemoteWriteVersion, version)
		}
	}

	// resetFieldFn resets the value of a field in the RemoteWriteSpec struct
	// if the field isn't supported by the current version.
	// It also logs a warning message reporting the minimum version required.
	resetFieldFn := func(minVersion string) func(string, any) {
		return func(field string, v any) {
			elem := reflect.ValueOf(v).Elem()
			if elem.IsNil() {
				return
			}
			o.logger.Warn(fmt.Sprintf("ignoring %q not supported by Thanos", field), "minimum_version", minVersion)
			elem.Set(reflect.Zero(elem.Type()))
		}
	}

	for i, rw := range tr.Spec.RemoteWrite {
		// Thanos does not support azureAD.workloadIdentity in any version
		if rw.AzureAD != nil && rw.AzureAD.WorkloadIdentity != nil {
			reset := resetFieldFn("none")
			reset("azureAD.workloadIdentity", &rw.AzureAD.WorkloadIdentity)
		}

		// Thanos does not support azureAD.scope in any version
		if rw.AzureAD != nil && rw.AzureAD.Scope != nil {
			reset := resetFieldFn("none")
			reset("azureAD.scope", &rw.AzureAD.Scope)
		}

		// Thanos v0.40.0 is equivalent to Prometheus v3.5.1 which allows empty clientId values.
		if version.LT(semver.MustParse("0.40.0")) {
			if rw.AzureAD != nil && rw.AzureAD.ManagedIdentity != nil {
				if ptr.Deref(rw.AzureAD.ManagedIdentity.ClientID, "") == "" {
					return fmt.Errorf("remoteWrite[%d]: azureAD.managedIdentity.clientId is required with Thanos < 0.40.0, current = %s", i, version)
				}
			}
		}
		// Thanos v0.38.0 is equivalent to Prometheus v3.1.0.
		if version.LT(semver.MustParse("0.38.0")) {
			reset := resetFieldFn("0.38.0")
			reset("roundRobinDNS", &rw.RoundRobinDNS) // requires >= 3.1.0
		}

		// Thanos v0.37.0 is equivalent to Prometheus v2.55.1.
		if version.LT(semver.MustParse("0.37.0")) {
			reset := resetFieldFn("0.37.0")
			reset("messageVersion", &rw.MessageVersion) // requires >= 2.54.0
		}

		// Thanos v0.36.0 is equivalent to Prometheus v2.52.2.
		if version.LT(semver.MustParse("0.36.0")) {
			reset := resetFieldFn("0.36.0")
			if rw.AzureAD != nil {
				reset("azureAD.sdk", &rw.AzureAD.SDK) // requires >= v2.52.2
			}
		}

		// Thanos v0.32.0 is equivalent to Prometheus v2.48.0.
		if version.LT(semver.MustParse("0.32.0")) {
			reset := resetFieldFn("0.32.0")
			if rw.QueueConfig != nil {
				reset("queueConfig.sampleAgeLimit", &rw.QueueConfig.SampleAgeLimit) // requires >= v2.50.0
			}
			reset("noProxy", &rw.NoProxy)                           // requires >= v2.48.0
			reset("proxyFromEnvironment", &rw.ProxyFromEnvironment) // requires >= v2.48.0
			reset("proxyConnectHeader", &rw.ProxyConnectHeader)     // requires >= v2.48.0
		}

		// Thanos v0.31.0 is equivalent to Prometheus v2.42.0.
		if version.LT(semver.MustParse("0.31.0")) {
			reset := resetFieldFn("0.31.0")
			if rw.AzureAD != nil {
				reset("azureAD.oauth", &rw.AzureAD.OAuth) // requires >= v2.48.0
			}
			reset("azureAD", &rw.AzureAD) // requires >= v2.45.0
		}

		// Thanos v0.30.0 is equivalent to Prometheus v2.40.7.
		if version.LT(semver.MustParse("0.30.0")) {
			reset := resetFieldFn("0.30.0")
			if rw.TLSConfig != nil {
				reset("tlsConfig.maxVersion", &rw.TLSConfig.MaxVersion) // requires >= v2.41.0
			}
			reset("sendNativeHistograms", &rw.SendNativeHistograms) // requires >= v2.40.0
		}

		// Thanos v0.28.0 is equivalent to Prometheus v2.38.0.
		if version.LT(semver.MustParse("0.28.0")) {
			reset := resetFieldFn("0.28.0")
			if rw.TLSConfig != nil {
				reset("tlsConfig.minVersion", &rw.TLSConfig.MinVersion) // >= requires v2.35.0
			}
		}

		// Thanos v0.24.0 is equivalent to Prometheus v2.32.0.
	}

	cg, err := prompkg.NewConfigGenerator(o.logger, nil, prompkg.WithoutVersionCheck())
	if err != nil {
		return err
	}

	err = cg.AddRemoteWriteToStore(ctx, store, tr.Namespace, tr.Spec.RemoteWrite)
	if err != nil {
		return err
	}

	rwConfig, err := yaml.Marshal(
		yaml.MapSlice{
			cg.GenerateRemoteWriteConfig(tr.Spec.RemoteWrite, store.ForNamespace(tr.Namespace)),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to marshal remote-write configuration: %w", err)
	}
	s.Data[rwConfigFile] = rwConfig

	if err = k8s.CreateOrUpdateSecret(ctx, sClient, s); err != nil {
		return err
	}

	return nil
}
