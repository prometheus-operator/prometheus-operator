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
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/listwatch"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/tools/cache"
)

const (
	resyncPeriod     = 5 * time.Minute
	thanosRulerLabel = "thanos-ruler"
)

// Operator manages life cycle of Thanos deployments and
// monitoring configurations.
type Operator struct {
	kclient  kubernetes.Interface
	mdClient metadata.Interface
	mclient  monitoringclient.Interface
	logger   log.Logger
	accessor *operator.Accessor

	thanosRulerInfs *informers.ForResource
	cmapInfs        *informers.ForResource
	ruleInfs        *informers.ForResource
	ssetInfs        *informers.ForResource

	rr *operator.ResourceReconciler

	nsThanosRulerInf cache.SharedIndexInformer
	nsRuleInf        cache.SharedIndexInformer

	metrics         *operator.Metrics
	reconciliations *operator.ReconciliationTracker

	config Config
}

// Config defines configuration parameters for the Operator.
type Config struct {
	ReloaderConfig         operator.ContainerConfig
	ThanosDefaultBaseImage string
	Namespaces             operator.Namespaces
	Labels                 operator.Labels
	LocalHost              string
	LogLevel               string
	LogFormat              string
	ThanosRulerSelector    string
}

// New creates a new controller.
func New(ctx context.Context, conf operator.Config, logger log.Logger, r prometheus.Registerer) (*Operator, error) {
	cfg, err := k8sutil.NewClusterConfig(conf.Host, conf.TLSInsecure, &conf.TLSConfig)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating cluster config failed")
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating kubernetes client failed")
	}

	mdClient, err := metadata.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating metadata client failed")
	}

	mclient, err := monitoringclient.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating monitoring client failed")
	}

	if _, err := labels.Parse(conf.ThanosRulerSelector); err != nil {
		return nil, errors.Wrap(err, "can not parse thanos ruler selector value")
	}

	// All the metrics exposed by the controller get the controller="thanos" label.
	r = prometheus.WrapRegistererWith(prometheus.Labels{"controller": "thanos"}, r)

	o := &Operator{
		kclient:         client,
		mdClient:        mdClient,
		mclient:         mclient,
		logger:          logger,
		accessor:        operator.NewAccessor(logger),
		metrics:         operator.NewMetrics(r),
		reconciliations: &operator.ReconciliationTracker{},
		config: Config{
			ReloaderConfig:         conf.ReloaderConfig,
			ThanosDefaultBaseImage: conf.ThanosDefaultBaseImage,
			Namespaces:             conf.Namespaces,
			Labels:                 conf.Labels,
			LocalHost:              conf.LocalHost,
			LogLevel:               conf.LogLevel,
			LogFormat:              conf.LogFormat,
			ThanosRulerSelector:    conf.ThanosRulerSelector,
		},
	}

	o.rr = operator.NewResourceReconciler(
		o.logger,
		o,
		o.metrics,
		monitoringv1.ThanosRulerKind,
		r,
	)

	o.cmapInfs, err = informers.NewInformersForResource(
		informers.NewMetadataInformerFactory(
			o.config.Namespaces.ThanosRulerAllowList,
			o.config.Namespaces.DenyList,
			o.mdClient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = labelThanosRulerName
			},
		),
		v1.SchemeGroupVersion.WithResource(string(v1.ResourceConfigMaps)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating configmap informers")
	}

	o.thanosRulerInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			o.config.Namespaces.ThanosRulerAllowList,
			o.config.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = o.config.ThanosRulerSelector
			},
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ThanosRulerName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating thanosruler informers")
	}

	var thanosStores []cache.Store
	for _, informer := range o.thanosRulerInfs.GetInformers() {
		thanosStores = append(thanosStores, informer.Informer().GetStore())
	}
	o.metrics.MustRegister(newThanosRulerCollectorForStores(thanosStores...))

	o.ruleInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			o.config.Namespaces.AllowList,
			o.config.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PrometheusRuleName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating prometheusrule informers")
	}

	o.ssetInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			o.config.Namespaces.ThanosRulerAllowList,
			o.config.Namespaces.DenyList,
			o.kclient,
			resyncPeriod,
			nil,
		),
		appsv1.SchemeGroupVersion.WithResource("statefulsets"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating statefulset informers")
	}

	newNamespaceInformer := func(o *Operator, allowList map[string]struct{}) cache.SharedIndexInformer {
		// nsResyncPeriod is used to control how often the namespace informer
		// should resync. If the unprivileged ListerWatcher is used, then the
		// informer must resync more often because it cannot watch for
		// namespace changes.
		nsResyncPeriod := 15 * time.Second
		// If the only namespace is v1.NamespaceAll, then the client must be
		// privileged and a regular cache.ListWatch will be used. In this case
		// watching works and we do not need to resync so frequently.
		if listwatch.IsAllNamespaces(allowList) {
			nsResyncPeriod = resyncPeriod
		}
		nsInf := cache.NewSharedIndexInformer(
			o.metrics.NewInstrumentedListerWatcher(
				listwatch.NewUnprivilegedNamespaceListWatchFromClient(ctx, o.logger, o.kclient.CoreV1().RESTClient(), allowList, o.config.Namespaces.DenyList, fields.Everything()),
			),
			&v1.Namespace{}, nsResyncPeriod, cache.Indexers{},
		)

		return nsInf
	}
	o.nsRuleInf = newNamespaceInformer(o, o.config.Namespaces.AllowList)
	if listwatch.IdenticalNamespaces(o.config.Namespaces.AllowList, o.config.Namespaces.ThanosRulerAllowList) {
		o.nsThanosRulerInf = o.nsRuleInf
	} else {
		o.nsThanosRulerInf = newNamespaceInformer(o, o.config.Namespaces.ThanosRulerAllowList)
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
			if !operator.WaitForNamedCacheSync(ctx, "thanos", log.With(o.logger, "informer", infs.name), inf.Informer()) {
				return errors.Errorf("failed to sync cache for %s informer", infs.name)
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
		if !operator.WaitForNamedCacheSync(ctx, "thanos", log.With(o.logger, "informer", inf.name), inf.informer) {
			return errors.Errorf("failed to sync cache for %s informer", inf.name)
		}
	}

	level.Info(o.logger).Log("msg", "successfully synced all caches")
	return nil
}

// addHandlers adds the eventhandlers to the informers.
func (o *Operator) addHandlers() {
	o.thanosRulerInfs.AddEventHandler(o.rr)
	o.ssetInfs.AddEventHandler(o.rr)

	o.cmapInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.handleConfigMapAdd,
		DeleteFunc: o.handleConfigMapDelete,
		UpdateFunc: o.handleConfigMapUpdate,
	})
	o.ruleInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.handleRuleAdd,
		DeleteFunc: o.handleRuleDelete,
		UpdateFunc: o.handleRuleUpdate,
	})

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
	errChan := make(chan error)
	go func() {
		v, err := o.kclient.Discovery().ServerVersion()
		if err != nil {
			errChan <- errors.Wrap(err, "communicating with server failed")
			return
		}
		level.Info(o.logger).Log("msg", "connection established", "cluster-version", v)
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
		level.Info(o.logger).Log("msg", "CRD API endpoints ready")
	case <-ctx.Done():
		return nil
	}

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
		level.Error(o.logger).Log("msg", "failed to list ThanosRuler objects", "err", err)
	}
}

// RefreshStatus implements the operator.StatusReconciler interface.
func (o *Operator) RefreshStatusFor(obj metav1.Object) {
	o.rr.EnqueueForStatus(obj)
}

// TODO: Do we need to enqueue configmaps just for the namespace or in general?
func (o *Operator) handleConfigMapAdd(obj interface{}) {
	meta, ok := o.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(o.logger).Log("msg", "ConfigMap added")
		o.metrics.TriggerByCounter("ConfigMap", operator.AddEvent).Inc()

		o.enqueueForThanosRulerNamespace(meta.GetNamespace())
	}
}

func (o *Operator) handleConfigMapDelete(obj interface{}) {
	meta, ok := o.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(o.logger).Log("msg", "ConfigMap deleted")
		o.metrics.TriggerByCounter("ConfigMap", operator.DeleteEvent).Inc()

		o.enqueueForThanosRulerNamespace(meta.GetNamespace())
	}
}

func (o *Operator) handleConfigMapUpdate(old, cur interface{}) {

	oldMeta, ok := o.accessor.ObjectMetadata(old)
	if !ok {
		return
	}

	curMeta, ok := o.accessor.ObjectMetadata(cur)
	if !ok {
		return
	}

	if oldMeta.GetResourceVersion() == curMeta.GetResourceVersion() {
		return
	}

	level.Debug(o.logger).Log("msg", "ConfigMap updated")
	o.metrics.TriggerByCounter("ConfigMap", operator.UpdateEvent).Inc()
	o.enqueueForThanosRulerNamespace(curMeta.GetNamespace())
}

// TODO: Don't enqueue just for the namespace
func (o *Operator) handleRuleAdd(obj interface{}) {
	meta, ok := o.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(o.logger).Log("msg", "PrometheusRule added")
		o.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, operator.AddEvent).Inc()

		o.enqueueForRulesNamespace(meta.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (o *Operator) handleRuleUpdate(old, cur interface{}) {
	if old.(*monitoringv1.PrometheusRule).ResourceVersion == cur.(*monitoringv1.PrometheusRule).ResourceVersion {
		return
	}

	meta, ok := o.accessor.ObjectMetadata(cur)
	if ok {
		level.Debug(o.logger).Log("msg", "PrometheusRule updated")
		o.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, operator.UpdateEvent).Inc()

		o.enqueueForRulesNamespace(meta.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (o *Operator) handleRuleDelete(obj interface{}) {
	meta, ok := o.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(o.logger).Log("msg", "PrometheusRule deleted")
		o.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, operator.DeleteEvent).Inc()

		o.enqueueForRulesNamespace(meta.GetNamespace())
	}
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
		level.Error(o.logger).Log("msg", "ThanosRuler lookup failed", "err", err)
		return nil
	}

	return tr.(*monitoringv1.ThanosRuler)
}

func statefulSetNameFromThanosName(name string) string {
	return "thanos-ruler-" + name
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

	level.Debug(o.logger).Log("msg", "update handler", "namespace", cur.GetName(), "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	// Periodic resync may resend the Namespace without changes in-between.
	if old.ResourceVersion == cur.ResourceVersion {
		return
	}

	level.Debug(o.logger).Log("msg", "Namespace updated", "namespace", cur.GetName())
	o.metrics.TriggerByCounter("Namespace", operator.UpdateEvent).Inc()

	// Check for ThanosRuler instances selecting PrometheusRules in the namespace.
	err := o.thanosRulerInfs.ListAll(labels.Everything(), func(obj interface{}) {
		tr := obj.(*monitoringv1.ThanosRuler)

		sync, err := k8sutil.LabelSelectionHasChanged(old.Labels, cur.Labels, tr.Spec.RuleNamespaceSelector)
		if err != nil {
			level.Error(o.logger).Log(
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
		level.Error(o.logger).Log(
			"msg", "listing all ThanosRuler instances from cache failed",
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
		return errors.Wrap(err, "failed to set ThanosRuler type information")
	}

	if tr.Spec.Paused {
		return nil
	}

	logger := log.With(o.logger, "key", key)
	level.Info(logger).Log("msg", "sync thanos-ruler")

	ruleConfigMapNames, err := o.createOrUpdateRuleConfigMaps(ctx, tr)
	if err != nil {
		return err
	}

	// Create governing service if it doesn't exist.
	svcClient := o.kclient.CoreV1().Services(tr.Namespace)
	if err = k8sutil.CreateOrUpdateService(ctx, svcClient, makeStatefulSetService(tr, o.config)); err != nil {
		return errors.Wrap(err, "synchronizing governing service failed")
	}

	// Ensure we have a StatefulSet running Thanos deployed.
	existingStatefulSet, err := o.getStatefulSetFromThanosRulerKey(key)
	if err != nil {
		return err
	}

	if existingStatefulSet == nil {
		ssetClient := o.kclient.AppsV1().StatefulSets(tr.Namespace)
		sset, err := makeStatefulSet(tr, o.config, ruleConfigMapNames, "")
		if err != nil {
			return errors.Wrap(err, "making thanos statefulset config failed")
		}

		operator.SanitizeSTS(sset)
		if _, err := ssetClient.Create(ctx, sset, metav1.CreateOptions{}); err != nil {
			return errors.Wrap(err, "creating thanos statefulset failed")
		}

		return nil
	}

	if o.rr.DeletionInProgress(existingStatefulSet) {
		return nil
	}

	newSSetInputHash, err := createSSetInputHash(*tr, o.config, ruleConfigMapNames, existingStatefulSet.Spec)
	if err != nil {
		return err
	}

	sset, err := makeStatefulSet(tr, o.config, ruleConfigMapNames, newSSetInputHash)
	if err != nil {
		return errors.Wrap(err, "making the statefulset, to update, failed")
	}

	operator.SanitizeSTS(sset)

	if newSSetInputHash == existingStatefulSet.ObjectMeta.Annotations[sSetInputHashName] {
		level.Debug(logger).Log("msg", "new statefulset generation inputs match current, skipping any actions")
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

		level.Info(logger).Log("msg", "recreating ThanosRuler StatefulSet because the update operation wasn't possible", "reason", strings.Join(failMsg, ", "))
		propagationPolicy := metav1.DeletePropagationForeground
		if err := ssetClient.Delete(ctx, sset.GetName(), metav1.DeleteOptions{PropagationPolicy: &propagationPolicy}); err != nil {
			return errors.Wrap(err, "failed to delete StatefulSet to avoid forbidden action")
		}
		return nil
	}

	if err != nil {
		return errors.Wrap(err, "updating StatefulSet failed")
	}

	return nil
}

// getThanosRulerFromKey returns a copy of the ThanosRuler object identified by key.
// If the object is not found, it returns a nil pointer.
func (o *Operator) getThanosRulerFromKey(key string) (*monitoringv1.ThanosRuler, error) {
	obj, err := o.thanosRulerInfs.Get(key)
	if err != nil {
		if apierrors.IsNotFound(err) {
			level.Info(o.logger).Log("msg", "ThanosRuler not found", "key", key)
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to retrieve ThanosRuler from informer")
	}

	return obj.(*monitoringv1.ThanosRuler).DeepCopy(), nil
}

// getStatefulSetFromThanosRulerKey returns a copy of the StatefulSet object
// corresponding to the ThanosRuler object identified by key.
// If the object is not found, it returns a nil pointer.
func (o *Operator) getStatefulSetFromThanosRulerKey(key string) (*appsv1.StatefulSet, error) {
	ssetName := thanosKeyToStatefulSetKey(key)

	obj, err := o.ssetInfs.Get(ssetName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			level.Info(o.logger).Log("msg", "StatefulSet not found", "key", ssetName)
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to retrieve StatefulSet from informer")
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
		return errors.Wrap(err, "failed to get StatefulSet")
	}

	if sset == nil || o.rr.DeletionInProgress(sset) {
		return nil
	}

	stsReporter, err := operator.NewStatefulSetReporter(ctx, o.kclient, sset)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve statefulset state")
	}

	availableCondition := stsReporter.Update(tr)
	reconciledCondition := o.reconciliations.GetCondition(key, tr.Generation)
	tr.Status.Conditions = operator.UpdateConditions(tr.Status.Conditions, availableCondition, reconciledCondition)

	tr.Status.Paused = tr.Spec.Paused

	if _, err = o.mclient.MonitoringV1().ThanosRulers(tr.Namespace).UpdateStatus(ctx, tr, metav1.UpdateOptions{}); err != nil {
		return errors.Wrap(err, "failed to update status subresource")
	}

	return nil
}

func createSSetInputHash(tr monitoringv1.ThanosRuler, c Config, ruleConfigMapNames []string, ss interface{}) (string, error) {
	hash, err := hashstructure.Hash(struct {
		TR monitoringv1.ThanosRuler
		C  Config
		S  interface{}
		R  []string `hash:"set"`
	}{tr, c, ss, ruleConfigMapNames},
		nil,
	)
	if err != nil {
		return "", errors.Wrap(
			err,
			"failed to calculate combined hash of ThanosRuler StatefulSet, ThanosRuler CRD, config and"+
				" rule ConfigMap names",
		)
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

// RulerStatus evaluates the current status of a ThanosRuler deployment with
// respect to its specified resource object. It returns the status and a list of
// pods that are not updated.
// TODO(simonpasquier): remove after 0.66.0 is released.
func RulerStatus(ctx context.Context, kclient kubernetes.Interface, tr *monitoringv1.ThanosRuler) (*monitoringv1.ThanosRulerStatus, []v1.Pod, error) {
	res := &monitoringv1.ThanosRulerStatus{Paused: tr.Spec.Paused}

	pods, err := kclient.CoreV1().Pods(tr.Namespace).List(ctx, ListOptions(tr.Name))
	if err != nil {
		return nil, nil, errors.Wrap(err, "retrieving pods of failed")
	}
	sset, err := kclient.AppsV1().StatefulSets(tr.Namespace).Get(ctx, statefulSetNameFromThanosName(tr.Name), metav1.GetOptions{})
	if err != nil {
		return nil, nil, errors.Wrap(err, "retrieving stateful set failed")
	}

	res.Replicas = int32(len(pods.Items))

	var oldPods []v1.Pod
	for _, pod := range pods.Items {
		ready, err := k8sutil.PodRunningAndReady(pod)
		if err != nil {
			return nil, nil, errors.Wrap(err, "cannot determine pod ready state")
		}
		if ready {
			res.AvailableReplicas++
			// TODO(fabxc): detect other fields of the pod template
			// that are mutable.
			if needsUpdate(&pod, sset.Spec.Template) {
				oldPods = append(oldPods, pod)
			} else {
				res.UpdatedReplicas++
			}
			continue
		}
		res.UnavailableReplicas++
	}

	return res, oldPods, nil
}

func needsUpdate(pod *v1.Pod, tmpl v1.PodTemplateSpec) bool {
	c1 := pod.Spec.Containers[0]
	c2 := tmpl.Spec.Containers[0]

	if c1.Image != c2.Image {
		return true
	}

	if !reflect.DeepEqual(c1.Args, c2.Args) {
		return true
	}

	return false
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
		level.Error(o.logger).Log(
			"msg", "get namespace to enqueue ThanosRuler instances failed",
			"err", err,
		)
		return
	}
	if !exists {
		level.Error(o.logger).Log(
			"msg", "get namespace to enqueue ThanosRuler instances failed: namespace does not exist",
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
			level.Error(o.logger).Log(
				"err", errors.Wrap(err, "failed to convert RuleNamespaceSelector"),
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
		level.Error(o.logger).Log(
			"msg", "listing all ThanosRuler instances from cache failed",
			"err", err,
		)
	}
}
