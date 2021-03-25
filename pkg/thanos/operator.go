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

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	resyncPeriod     = 5 * time.Minute
	thanosRulerLabel = "thanos-ruler"
)

// Operator manages life cycle of Thanos deployments and
// monitoring configurations.
type Operator struct {
	kclient kubernetes.Interface
	mclient monitoringclient.Interface
	logger  log.Logger

	thanosRulerInfs *informers.ForResource
	cmapInfs        *informers.ForResource
	ruleInfs        *informers.ForResource
	ssetInfs        *informers.ForResource

	nsThanosRulerInf cache.SharedIndexInformer
	nsRuleInf        cache.SharedIndexInformer

	queue workqueue.RateLimitingInterface

	metrics *operator.Metrics

	config Config
}

// Config defines configuration parameters for the Operator.
type Config struct {
	Host                   string
	TLSInsecure            bool
	TLSConfig              rest.TLSClientConfig
	ReloaderConfig         operator.ReloaderConfig
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

	mclient, err := monitoringclient.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating monitoring client failed")
	}

	if _, err := labels.Parse(conf.ThanosRulerSelector); err != nil {
		return nil, errors.Wrap(err, "can not parse thanos ruler selector value")
	}

	o := &Operator{
		kclient: client,
		mclient: mclient,
		logger:  logger,
		queue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "thanos"),
		metrics: operator.NewMetrics("thanos", r),
		config: Config{
			Host:                   conf.Host,
			TLSInsecure:            conf.TLSInsecure,
			TLSConfig:              conf.TLSConfig,
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

	o.cmapInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			o.config.Namespaces.ThanosRulerAllowList,
			o.config.Namespaces.DenyList,
			o.kclient,
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
	ok := true

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
				ok = false
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
			ok = false
		}
	}

	if !ok {
		return errors.New("failed to sync caches")
	}

	level.Info(o.logger).Log("msg", "successfully synced all caches")
	return nil
}

// addHandlers adds the eventhandlers to the informers.
func (o *Operator) addHandlers() {
	o.thanosRulerInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.handleThanosRulerAdd,
		DeleteFunc: o.handleThanosRulerDelete,
		UpdateFunc: o.handleThanosRulerUpdate,
	})
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
	o.ssetInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.handleStatefulSetAdd,
		DeleteFunc: o.handleStatefulSetDelete,
		UpdateFunc: o.handleStatefulSetUpdate,
	})
}

// Run the controller.
func (o *Operator) Run(ctx context.Context) error {
	defer o.queue.ShutDown()

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

	go o.worker(ctx)

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
	o.addHandlers()

	o.metrics.Ready().Set(1)
	<-ctx.Done()
	return nil
}

func (o *Operator) keyFunc(obj interface{}) (string, bool) {
	k, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		level.Error(o.logger).Log("msg", "creating key failed", "err", err)
		return k, false
	}
	return k, true
}

func (o *Operator) handleThanosRulerAdd(obj interface{}) {
	key, ok := o.keyFunc(obj)
	if !ok {
		return
	}

	level.Debug(o.logger).Log("msg", "ThanosRuler added", "key", key)
	o.metrics.TriggerByCounter(monitoringv1.ThanosRulerKind, "add").Inc()
	o.enqueue(key)
}

func (o *Operator) handleThanosRulerDelete(obj interface{}) {
	key, ok := o.keyFunc(obj)
	if !ok {
		return
	}

	level.Debug(o.logger).Log("msg", "ThanosRuler deleted", "key", key)
	o.metrics.TriggerByCounter(monitoringv1.ThanosRulerKind, "delete").Inc()
	o.enqueue(key)
}

func (o *Operator) handleThanosRulerUpdate(old, cur interface{}) {
	if old.(*monitoringv1.ThanosRuler).ResourceVersion == cur.(*monitoringv1.ThanosRuler).ResourceVersion {
		return
	}

	key, ok := o.keyFunc(cur)
	if !ok {
		return
	}

	level.Debug(o.logger).Log("msg", "ThanosRuler updated", "key", key)
	o.metrics.TriggerByCounter(monitoringv1.ThanosRulerKind, "update").Inc()
	o.enqueue(key)
}

// TODO: Do we need to enqueue configmaps just for the namespace or in general?
func (o *Operator) handleConfigMapAdd(obj interface{}) {
	meta, ok := o.getObjectMeta(obj)
	if ok {
		level.Debug(o.logger).Log("msg", "ConfigMap added")
		o.metrics.TriggerByCounter("ConfigMap", "add").Inc()

		o.enqueueForThanosRulerNamespace(meta.GetNamespace())
	}
}

func (o *Operator) handleConfigMapDelete(obj interface{}) {
	meta, ok := o.getObjectMeta(obj)
	if ok {
		level.Debug(o.logger).Log("msg", "ConfigMap deleted")
		o.metrics.TriggerByCounter("ConfigMap", "delete").Inc()

		o.enqueueForThanosRulerNamespace(meta.GetNamespace())
	}
}

func (o *Operator) handleConfigMapUpdate(old, cur interface{}) {
	if old.(*v1.ConfigMap).ResourceVersion == cur.(*v1.ConfigMap).ResourceVersion {
		return
	}

	meta, ok := o.getObjectMeta(cur)
	if ok {
		level.Debug(o.logger).Log("msg", "ConfigMap updated")
		o.metrics.TriggerByCounter("ConfigMap", "update").Inc()

		o.enqueueForThanosRulerNamespace(meta.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (o *Operator) handleRuleAdd(obj interface{}) {
	meta, ok := o.getObjectMeta(obj)
	if ok {
		level.Debug(o.logger).Log("msg", "PrometheusRule added")
		o.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, "add").Inc()

		o.enqueueForRulesNamespace(meta.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (o *Operator) handleRuleUpdate(old, cur interface{}) {
	if old.(*monitoringv1.PrometheusRule).ResourceVersion == cur.(*monitoringv1.PrometheusRule).ResourceVersion {
		return
	}

	meta, ok := o.getObjectMeta(cur)
	if ok {
		level.Debug(o.logger).Log("msg", "PrometheusRule updated")
		o.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, "update").Inc()

		o.enqueueForRulesNamespace(meta.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (o *Operator) handleRuleDelete(obj interface{}) {
	meta, ok := o.getObjectMeta(obj)
	if ok {
		level.Debug(o.logger).Log("msg", "PrometheusRule deleted")
		o.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, "delete").Inc()

		o.enqueueForRulesNamespace(meta.GetNamespace())
	}
}

func (o *Operator) thanosForStatefulSet(sset interface{}) *monitoringv1.ThanosRuler {
	key, ok := o.keyFunc(sset)
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

func (o *Operator) handleStatefulSetAdd(obj interface{}) {
	if ps := o.thanosForStatefulSet(obj); ps != nil {
		level.Debug(o.logger).Log("msg", "StatefulSet added")
		o.metrics.TriggerByCounter("StatefulSet", "add").Inc()

		o.enqueue(ps)
	}
}

func (o *Operator) handleStatefulSetDelete(obj interface{}) {
	if ps := o.thanosForStatefulSet(obj); ps != nil {
		level.Debug(o.logger).Log("msg", "StatefulSet delete")
		o.metrics.TriggerByCounter("StatefulSet", "delete").Inc()

		o.enqueue(ps)
	}
}

func (o *Operator) handleStatefulSetUpdate(oldo, curo interface{}) {
	old := oldo.(*appsv1.StatefulSet)
	cur := curo.(*appsv1.StatefulSet)

	level.Debug(o.logger).Log("msg", "update handler", "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	// Periodic resync may resend the StatefulSet without changes
	// in-between. Also breaks loops created by updating the resource
	// ourselves.
	if old.ResourceVersion == cur.ResourceVersion {
		return
	}

	if ps := o.thanosForStatefulSet(cur); ps != nil {
		level.Debug(o.logger).Log("msg", "StatefulSet updated")
		o.metrics.TriggerByCounter("StatefulSet", "update").Inc()

		o.enqueue(ps)
	}
}

// enqueue adds a key to the queue. If obj is a key already it gets added
// directly. Otherwise, the key is extracted via keyFunc.
func (o *Operator) enqueue(obj interface{}) {
	if obj == nil {
		return
	}

	key, ok := obj.(string)
	if !ok {
		key, ok = o.keyFunc(obj)
		if !ok {
			return
		}
	}

	o.queue.Add(key)
}

// worker runs a worker thread that just dequeues items, processes them, and
// marks them done. It enforces that the syncHandler is never invoked
// concurrently with the same key.
func (o *Operator) worker(ctx context.Context) {
	for o.processNextWorkItem(ctx) {
	}
}

func (o *Operator) processNextWorkItem(ctx context.Context) bool {
	key, quit := o.queue.Get()
	if quit {
		return false
	}
	defer o.queue.Done(key)

	o.metrics.ReconcileCounter().Inc()
	err := o.sync(ctx, key.(string))
	o.metrics.SetSyncStatus(key.(string), err == nil)
	if err == nil {
		o.queue.Forget(key)
		return true
	}

	o.metrics.ReconcileErrorsCounter().Inc()
	utilruntime.HandleError(errors.Wrap(err, fmt.Sprintf("Sync %q failed", key)))
	o.queue.AddRateLimited(key)

	return true
}

func (o *Operator) sync(ctx context.Context, key string) error {
	trobj, err := o.thanosRulerInfs.Get(key)
	if apierrors.IsNotFound(err) {
		o.metrics.ForgetObject(key)
		// Dependent resources are cleaned up by K8s via OwnerReferences
		return nil
	}
	if err != nil {
		return err
	}

	tr := trobj.(*monitoringv1.ThanosRuler)
	tr = tr.DeepCopy()
	tr.APIVersion = monitoringv1.SchemeGroupVersion.String()
	tr.Kind = monitoringv1.ThanosRulerKind

	if tr.Spec.Paused {
		return nil
	}

	level.Info(o.logger).Log("msg", "sync thanos-ruler", "key", key)

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
	ssetClient := o.kclient.AppsV1().StatefulSets(tr.Namespace)
	obj, err := o.ssetInfs.Get(thanosKeyToStatefulSetKey(key))

	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "retrieving statefulset failed")
	}

	exists := !apierrors.IsNotFound(err)

	if !exists {
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

	spec := appsv1.StatefulSetSpec{}
	if obj != nil {
		ss := obj.(*appsv1.StatefulSet)
		spec = ss.Spec
	}

	newSSetInputHash, err := createSSetInputHash(*tr, o.config, ruleConfigMapNames, spec)
	if err != nil {
		return err
	}

	sset, err := makeStatefulSet(tr, o.config, ruleConfigMapNames, newSSetInputHash)
	if err != nil {
		return errors.Wrap(err, "making the statefulset, to update, failed")
	}

	operator.SanitizeSTS(sset)

	oldSSetInputHash := obj.(*appsv1.StatefulSet).ObjectMeta.Annotations[sSetInputHashName]
	if newSSetInputHash == oldSSetInputHash {
		level.Debug(o.logger).Log("msg", "new statefulset generation inputs match current, skipping any actions")
		return nil
	}

	err = k8sutil.UpdateStatefulSet(ctx, ssetClient, sset)
	sErr, ok := err.(*apierrors.StatusError)

	if ok && sErr.ErrStatus.Code == 422 && sErr.ErrStatus.Reason == metav1.StatusReasonInvalid {
		o.metrics.StsDeleteCreateCounter().Inc()
		level.Info(o.logger).Log("msg", "resolving illegal update of ThanosRuler StatefulSet", "details", sErr.ErrStatus.Details)
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

// listMatchingNamespaces lists all the namespaces that match the provided
// selector.
func (o *Operator) listMatchingNamespaces(selector labels.Selector) ([]string, error) {
	var ns []string
	err := cache.ListAll(o.nsRuleInf.GetStore(), selector, func(obj interface{}) {
		ns = append(ns, obj.(*v1.Namespace).Name)
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list namespaces")
	}
	return ns, nil
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
			"app":            thanosRulerLabel,
			thanosRulerLabel: name,
		})).String(),
	}
}

// RulerStatus evaluates the current status of a ThanosRuler deployment with
// respect to its specified resource object. It returns the status and a list of
// pods that are not updated.
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
			"msg", fmt.Sprintf("get namespace to enqueue ThanosRuler instances failed: namespace %q does not exist", nsName),
		)
		return
	}
	ns := nsObject.(*v1.Namespace)

	err = o.thanosRulerInfs.ListAll(labels.Everything(), func(obj interface{}) {
		// Check for ThanosRuler instances in the namespace.
		tr := obj.(*monitoringv1.ThanosRuler)
		if tr.Namespace == nsName {
			o.enqueue(tr)
			return
		}

		// Check for ThanosRuler instances selecting PrometheusRules in
		// the namespace.
		ruleNSSelector, err := metav1.LabelSelectorAsSelector(tr.Spec.RuleNamespaceSelector)
		if err != nil {
			level.Error(o.logger).Log(
				"msg", fmt.Sprintf("failed to convert RuleNamespaceSelector of %q to selector", tr.Name),
				"err", err,
			)
			return
		}

		if ruleNSSelector.Matches(labels.Set(ns.Labels)) {
			o.enqueue(tr)
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

func (o *Operator) getObjectMeta(obj interface{}) (metav1.Object, bool) {
	ts, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		obj = ts.Obj
	}

	metaObj, err := meta.Accessor(obj)
	if err != nil {
		level.Error(o.logger).Log("msg", "get object failed", "err", err)
		return nil, false
	}
	return metaObj, true
}
