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
	"strings"
	"time"

	"github.com/go-kit/log"
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

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	monitoringv1ac "github.com/prometheus-operator/prometheus-operator/pkg/client/applyconfiguration/monitoring/v1"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/listwatch"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
)

const (
	resyncPeriod     = 5 * time.Minute
	thanosRulerLabel = "thanos-ruler"
	controllerName   = "thanos-controller"
)

// Operator manages life cycle of Thanos deployments and
// monitoring configurations.
type Operator struct {
	kclient  kubernetes.Interface
	mdClient metadata.Interface
	mclient  monitoringclient.Interface
	// We're currently migrating our logging library from go-kit to slog.
	// The go-kit logger is being removed in small PRs. For now, we are creating 2 loggers to avoid breaking changes and
	// to have a smooth transition.
	goKitLogger log.Logger
	logger      *slog.Logger
	accessor    *operator.Accessor

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

	eventRecorder record.EventRecorder

	config Config
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

// New creates a new controller.
func New(ctx context.Context, restConfig *rest.Config, c operator.Config, goKitLogger log.Logger, logger *slog.Logger, r prometheus.Registerer, options ...ControllerOption) (*Operator, error) {
	goKitLogger = log.With(goKitLogger, "component", controllerName)
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

	// All the metrics exposed by the controller get the controller="thanos" label.
	r = prometheus.WrapRegistererWith(prometheus.Labels{"controller": "thanos"}, r)

	o := &Operator{
		kclient:         client,
		mdClient:        mdClient,
		mclient:         mclient,
		goKitLogger:     goKitLogger,
		logger:          logger,
		accessor:        operator.NewAccessor(logger),
		metrics:         operator.NewMetrics(r),
		eventRecorder:   c.EventRecorderFactory(client, controllerName),
		reconciliations: &operator.ReconciliationTracker{},
		controllerID:    c.ControllerID,
		config: Config{
			ReloaderConfig:         c.ReloaderConfig,
			ThanosDefaultBaseImage: c.ThanosDefaultBaseImage,
			Annotations:            c.Annotations,
			Labels:                 c.Labels,
			LocalHost:              c.LocalHost,
		},
	}
	for _, opt := range options {
		opt(o)
	}

	o.rr = operator.NewResourceReconciler(
		o.logger,
		o,
		o.metrics,
		monitoringv1.ThanosRulerKind,
		r,
		o.controllerID,
	)

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
			nil,
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
		"ConfigMap",
		o.enqueueForThanosRulerNamespace,
	))

	o.ruleInfs.AddEventHandler(operator.NewEventHandler(
		o.logger,
		o.accessor,
		o.metrics,
		monitoringv1.PrometheusRuleKind,
		o.enqueueForRulesNamespace,
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
	_ = o.thanosRulerInfs.ListAll(labels.Everything(), func(obj interface{}) {
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
func (o *Operator) Iterate(processFn func(metav1.Object, []monitoringv1.Condition)) {
	if err := o.thanosRulerInfs.ListAll(labels.Everything(), func(o interface{}) {
		a := o.(*monitoringv1.ThanosRuler)
		processFn(a, a.Status.Conditions)
	}); err != nil {
		o.logger.Error("failed to list ThanosRuler objects", "err", err)
	}
}

// RefreshStatus implements the operator.StatusReconciler interface.
func (o *Operator) RefreshStatusFor(obj metav1.Object) {
	o.rr.EnqueueForStatus(obj)
}

// Resolve implements the operator.Syncer interface.
func (o *Operator) Resolve(ss *appsv1.StatefulSet) metav1.Object {
	key, ok := o.accessor.MetaNamespaceKey(ss)
	if !ok {
		return nil
	}

	thanosKey := statefulSetKeyToThanosKey(key)
	tr, err := o.thanosRulerInfs.Get(thanosKey)
	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		o.logger.Error("ThanosRuler lookup failed", "err", err)
		return nil
	}

	return tr.(*monitoringv1.ThanosRuler)
}

func statefulSetKeyToThanosKey(key string) string {
	keyParts := strings.Split(key, "/")
	return keyParts[0] + "/" + strings.TrimPrefix(keyParts[1], "thanos-ruler-")
}

func thanosKeyToStatefulSetKey(key string) string {
	keyParts := strings.Split(key, "/")
	return keyParts[0] + "/thanos-ruler-" + keyParts[1]
}

func (o *Operator) handleNamespaceUpdate(oldo, curo interface{}) {
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
	err := o.thanosRulerInfs.ListAll(labels.Everything(), func(obj interface{}) {
		tr := obj.(*monitoringv1.ThanosRuler)

		sync, err := k8sutil.LabelSelectionHasChanged(old.Labels, cur.Labels, tr.Spec.RuleNamespaceSelector)
		if err != nil {
			o.logger.Error("",
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
	err := o.sync(ctx, key)
	o.reconciliations.SetStatus(key, err)

	return err
}

func (o *Operator) sync(ctx context.Context, key string) error {
	trobj, err := o.thanosRulerInfs.Get(key)
	if apierrors.IsNotFound(err) {
		o.reconciliations.ForgetObject(key)
		// Dependent resources are cleaned up by K8s via OwnerReferences
		return nil
	}
	if err != nil {
		return err
	}

	tr := trobj.(*monitoringv1.ThanosRuler)
	tr = tr.DeepCopy()
	if err := k8sutil.AddTypeInformationToObject(tr); err != nil {
		return fmt.Errorf("failed to set ThanosRuler type information: %w", err)
	}

	// Check if the Thanos instance is marked for deletion.
	if o.rr.DeletionInProgress(tr) {
		return nil
	}

	if tr.Spec.Paused {
		return nil
	}

	logger := o.logger.With("key", key)
	logger.Info("sync thanos-ruler")

	if err := operator.CheckStorageClass(ctx, o.canReadStorageClass, o.kclient, tr.Spec.Storage); err != nil {
		return err
	}

	ruleConfigMapNames, err := o.createOrUpdateRuleConfigMaps(ctx, tr)
	if err != nil {
		return err
	}

	assetStore := assets.NewStoreBuilder(o.kclient.CoreV1(), o.kclient.CoreV1())

	tlsAssets, err := operator.ReconcileShardedSecret(ctx, assetStore.TLSAssets(), o.kclient, newTLSAssetSecret(tr, o.config))
	if err != nil {
		return fmt.Errorf("failed to reconcile the TLS secrets: %w", err)
	}

	if err := o.createOrUpdateWebConfigSecret(ctx, tr); err != nil {
		return fmt.Errorf("failed to synchronize web config secret: %w", err)
	}

	// Create governing service if it doesn't exist.
	svcClient := o.kclient.CoreV1().Services(tr.Namespace)
	if err = k8sutil.CreateOrUpdateService(ctx, svcClient, makeStatefulSetService(tr, o.config)); err != nil {
		return fmt.Errorf("synchronizing governing service failed: %w", err)
	}

	// Ensure we have a StatefulSet running Thanos deployed.
	existingStatefulSet, err := o.getStatefulSetFromThanosRulerKey(key)
	if err != nil {
		return err
	}

	if existingStatefulSet == nil {
		ssetClient := o.kclient.AppsV1().StatefulSets(tr.Namespace)
		sset, err := makeStatefulSet(tr, o.config, ruleConfigMapNames, "", tlsAssets)
		if err != nil {
			return fmt.Errorf("making thanos statefulset config failed: %w", err)
		}

		operator.SanitizeSTS(sset)
		if _, err := ssetClient.Create(ctx, sset, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("creating thanos statefulset failed: %w", err)
		}

		return nil
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

	if newSSetInputHash == existingStatefulSet.ObjectMeta.Annotations[operator.InputHashAnnotationName] {
		logger.Debug("new statefulset generation inputs match current, skipping any actions")
		return nil
	}

	ssetClient := o.kclient.AppsV1().StatefulSets(tr.Namespace)
	err = k8sutil.UpdateStatefulSet(ctx, ssetClient, sset)
	sErr, ok := err.(*apierrors.StatusError)

	if ok && sErr.ErrStatus.Code == 422 && sErr.ErrStatus.Reason == metav1.StatusReasonInvalid {
		o.metrics.StsDeleteCreateCounter().Inc()

		// Gather only reason for failed update
		failMsg := make([]string, len(sErr.ErrStatus.Details.Causes))
		for i, cause := range sErr.ErrStatus.Details.Causes {
			failMsg[i] = cause.Message
		}

		logger.Info("recreating ThanosRuler StatefulSet because the update operation wasn't possible", "reason", strings.Join(failMsg, ", "))
		propagationPolicy := metav1.DeletePropagationForeground
		if err := ssetClient.Delete(ctx, sset.GetName(), metav1.DeleteOptions{PropagationPolicy: &propagationPolicy}); err != nil {
			return fmt.Errorf("failed to delete StatefulSet to avoid forbidden action: %w", err)
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("updating StatefulSet failed: %w", err)
	}

	return nil
}

// getThanosRulerFromKey returns a copy of the ThanosRuler object identified by key.
// If the object is not found, it returns a nil pointer.
func (o *Operator) getThanosRulerFromKey(key string) (*monitoringv1.ThanosRuler, error) {
	obj, err := o.thanosRulerInfs.Get(key)
	if err != nil {
		if apierrors.IsNotFound(err) {
			o.logger.Info("ThanosRuler not found", "key", key)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve ThanosRuler from informer: %w", err)
	}

	return obj.(*monitoringv1.ThanosRuler).DeepCopy(), nil
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
	tr, err := o.getThanosRulerFromKey(key)
	if err != nil {
		return err
	}

	if tr == nil || o.rr.DeletionInProgress(tr) {
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

	if _, err = o.mclient.MonitoringV1().ThanosRulers(tr.Namespace).ApplyStatus(ctx, applyConfigurationFromThanosRuler(tr), metav1.ApplyOptions{FieldManager: operator.PrometheusOperatorFieldManager, Force: true}); err != nil {
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

func ListOptions(name string) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
			"app.kubernetes.io/name": thanosRulerLabel,
			thanosRulerLabel:         name,
		})).String(),
	}
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

	err = o.thanosRulerInfs.ListAll(labels.Everything(), func(obj interface{}) {
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
