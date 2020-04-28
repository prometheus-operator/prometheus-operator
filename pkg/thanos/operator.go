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
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/coreos/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringclient "github.com/coreos/prometheus-operator/pkg/client/versioned"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/coreos/prometheus-operator/pkg/listwatch"
	"github.com/coreos/prometheus-operator/pkg/operator"
	"github.com/mitchellh/hashstructure"

	prometheusoperator "github.com/coreos/prometheus-operator/pkg/prometheus"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
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
	kclient   kubernetes.Interface
	mclient   monitoringclient.Interface
	crdclient apiextensionsclient.Interface
	logger    log.Logger

	thanosRulerInf   cache.SharedIndexInformer
	nsThanosRulerInf cache.SharedIndexInformer
	nsRuleInf        cache.SharedIndexInformer
	cmapInf          cache.SharedIndexInformer
	ruleInf          cache.SharedIndexInformer
	ssetInf          cache.SharedIndexInformer

	queue workqueue.RateLimitingInterface

	metrics *operator.Metrics

	config Config
}

// Config defines configuration parameters for the Operator.
type Config struct {
	Host                   string
	TLSInsecure            bool
	TLSConfig              rest.TLSClientConfig
	ConfigReloaderImage    string
	ConfigReloaderCPU      string
	ConfigReloaderMemory   string
	ThanosDefaultBaseImage string
	Namespaces             prometheusoperator.Namespaces
	Labels                 prometheusoperator.Labels
	CrdKinds               monitoringv1.CrdKinds
	EnableValidation       bool
	LocalHost              string
	LogLevel               string
	LogFormat              string
	ManageCRDs             bool
	ThanosRulerSelector    string
}

// New creates a new controller.
func New(conf prometheusoperator.Config, logger log.Logger, r prometheus.Registerer) (*Operator, error) {
	cfg, err := k8sutil.NewClusterConfig(conf.Host, conf.TLSInsecure, &conf.TLSConfig)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating cluster config failed")
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating kubernetes client failed")
	}

	crdclient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating apiextensions client failed")
	}

	mclient, err := monitoringclient.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating monitoring client failed")
	}

	if _, err := labels.Parse(conf.ThanosRulerSelector); err != nil {
		return nil, errors.Wrap(err, "can not parse thanos ruler selector value")
	}

	o := &Operator{
		kclient:   client,
		mclient:   mclient,
		crdclient: crdclient,
		logger:    logger,
		queue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "thanos"),
		metrics:   operator.NewMetrics("thanos", r),
		config: Config{
			Host:                   conf.Host,
			TLSInsecure:            conf.TLSInsecure,
			TLSConfig:              conf.TLSConfig,
			ConfigReloaderImage:    conf.ConfigReloaderImage,
			ConfigReloaderCPU:      conf.ConfigReloaderCPU,
			ConfigReloaderMemory:   conf.ConfigReloaderMemory,
			ThanosDefaultBaseImage: conf.ThanosDefaultBaseImage,
			Namespaces:             conf.Namespaces,
			Labels:                 conf.Labels,
			CrdKinds:               conf.CrdKinds,
			EnableValidation:       conf.EnableValidation,
			LocalHost:              conf.LocalHost,
			LogLevel:               conf.LogLevel,
			LogFormat:              conf.LogFormat,
			ManageCRDs:             conf.ManageCRDs,
			ThanosRulerSelector:    conf.ThanosRulerSelector,
		},
	}

	o.cmapInf = cache.NewSharedIndexInformer(
		o.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(o.logger, o.config.Namespaces.ThanosRulerAllowList, o.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						options.LabelSelector = labelThanosRulerName
						return o.kclient.CoreV1().ConfigMaps(namespace).List(options)
					},
					WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
						options.LabelSelector = labelThanosRulerName
						return o.kclient.CoreV1().ConfigMaps(namespace).Watch(options)
					},
				}
			}),
		),
		&v1.ConfigMap{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	o.thanosRulerInf = cache.NewSharedIndexInformer(
		o.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(o.logger, o.config.Namespaces.ThanosRulerAllowList, o.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						options.LabelSelector = o.config.ThanosRulerSelector
						return o.mclient.MonitoringV1().ThanosRulers(namespace).List(options)
					},
					WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
						options.LabelSelector = o.config.ThanosRulerSelector
						return o.mclient.MonitoringV1().ThanosRulers(namespace).Watch(options)
					},
				}
			}),
		),
		&monitoringv1.ThanosRuler{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	o.metrics.MustRegister(NewThanosRulerCollector(o.thanosRulerInf.GetStore()))
	o.ruleInf = cache.NewSharedIndexInformer(
		o.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(o.logger, o.config.Namespaces.AllowList, o.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						return mclient.MonitoringV1().PrometheusRules(namespace).List(options)
					},
					WatchFunc: mclient.MonitoringV1().PrometheusRules(namespace).Watch,
				}
			}),
		),
		&monitoringv1.PrometheusRule{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	o.ssetInf = cache.NewSharedIndexInformer(
		listwatch.MultiNamespaceListerWatcher(o.logger, o.config.Namespaces.ThanosRulerAllowList, o.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
			return cache.NewListWatchFromClient(o.kclient.AppsV1().RESTClient(), "statefulsets", namespace, fields.Everything())
		}),
		&appsv1.StatefulSet{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

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
				listwatch.NewUnprivilegedNamespaceListWatchFromClient(o.logger, o.kclient.CoreV1().RESTClient(), allowList, o.config.Namespaces.DenyList, fields.Everything()),
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
func (o *Operator) waitForCacheSync(stopc <-chan struct{}) error {
	ok := true
	informers := []struct {
		name     string
		informer cache.SharedIndexInformer
	}{
		{"ThanosRuler", o.thanosRulerInf},
		{"ThanosRulerNamespace", o.nsThanosRulerInf},
		{"RuleNamespace", o.nsRuleInf},
		{"ConfigMap", o.cmapInf},
		{"PrometheusRule", o.ruleInf},
		{"StatefulSet", o.ssetInf},
	}
	for _, inf := range informers {
		if !cache.WaitForCacheSync(stopc, inf.informer.HasSynced) {
			level.Error(o.logger).Log("msg", fmt.Sprintf("failed to sync %s cache", inf.name))
			ok = false
		} else {
			level.Debug(o.logger).Log("msg", fmt.Sprintf("successfully synced %s cache", inf.name))
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
	o.thanosRulerInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.handleThanosRulerAdd,
		DeleteFunc: o.handleThanosRulerDelete,
		UpdateFunc: o.handleThanosRulerUpdate,
	})
	o.cmapInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.handleConfigMapAdd,
		DeleteFunc: o.handleConfigMapDelete,
		UpdateFunc: o.handleConfigMapUpdate,
	})
	o.ruleInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.handleRuleAdd,
		DeleteFunc: o.handleRuleDelete,
		UpdateFunc: o.handleRuleUpdate,
	})
	o.ssetInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.handleStatefulSetAdd,
		DeleteFunc: o.handleStatefulSetDelete,
		UpdateFunc: o.handleStatefulSetUpdate,
	})
}

// Run the controller.
func (o *Operator) Run(stopc <-chan struct{}) error {
	defer o.queue.ShutDown()

	errChan := make(chan error)
	go func() {
		v, err := o.kclient.Discovery().ServerVersion()
		if err != nil {
			errChan <- errors.Wrap(err, "communicating with server failed")
			return
		}
		level.Info(o.logger).Log("msg", "connection established", "cluster-version", v)

		if o.config.ManageCRDs {
			if err := o.createCRDs(); err != nil {
				errChan <- errors.Wrap(err, "creating CRDs failed")
				return
			}
		}
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
		level.Info(o.logger).Log("msg", "CRD API endpoints ready")
	case <-stopc:
		return nil
	}

	go o.worker()

	go o.thanosRulerInf.Run(stopc)
	go o.cmapInf.Run(stopc)
	go o.ruleInf.Run(stopc)
	go o.nsRuleInf.Run(stopc)
	if o.nsRuleInf != o.nsThanosRulerInf {
		go o.nsThanosRulerInf.Run(stopc)
	}
	go o.ssetInf.Run(stopc)
	if err := o.waitForCacheSync(stopc); err != nil {
		return err
	}
	o.addHandlers()

	<-stopc
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

// TODO: Do we need to enque configmaps just for the namespace or in general?
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

// TODO: Don't enque just for the namespace
func (o *Operator) handleRuleAdd(obj interface{}) {
	meta, ok := o.getObjectMeta(obj)
	if ok {
		level.Debug(o.logger).Log("msg", "PrometheusRule added")
		o.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, "add").Inc()

		o.enqueueForRulesNamespace(meta.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
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

// TODO: Don't enque just for the namespace
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
	tr, exists, err := o.thanosRulerInf.GetStore().GetByKey(thanosKey)
	if err != nil {
		level.Error(o.logger).Log("msg", "ThanosRuler lookup failed", "err", err)
		return nil
	}
	if !exists {
		return nil
	}
	return tr.(*monitoringv1.ThanosRuler)
}

func thanosNameFromStatefulSetName(name string) string {
	return strings.TrimPrefix(name, "thanos-ruler-")
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
func (o *Operator) worker() {
	for o.processNextWorkItem() {
	}
}

func (o *Operator) processNextWorkItem() bool {
	key, quit := o.queue.Get()
	if quit {
		return false
	}
	defer o.queue.Done(key)

	err := o.sync(key.(string))
	if err == nil {
		o.queue.Forget(key)
		return true
	}

	o.metrics.ReconcileErrorsCounter().Inc()
	utilruntime.HandleError(errors.Wrap(err, fmt.Sprintf("Sync %q failed", key)))
	o.queue.AddRateLimited(key)

	return true
}

func (o *Operator) sync(key string) error {
	obj, exists, err := o.thanosRulerInf.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		// Dependent resources are cleaned up by K8s via OwnerReferences
		return nil
	}

	tr := obj.(*monitoringv1.ThanosRuler)
	tr = tr.DeepCopy()
	tr.APIVersion = monitoringv1.SchemeGroupVersion.String()
	tr.Kind = monitoringv1.ThanosRulerKind

	if tr.Spec.Paused {
		return nil
	}

	level.Info(o.logger).Log("msg", "sync thanos-ruler", "key", key)

	ruleConfigMapNames, err := o.createOrUpdateRuleConfigMaps(tr)
	if err != nil {
		return err
	}

	// Create governing service if it doesn't exist.
	svcClient := o.kclient.CoreV1().Services(tr.Namespace)
	if err = k8sutil.CreateOrUpdateService(svcClient, makeStatefulSetService(tr, o.config)); err != nil {
		return errors.Wrap(err, "synchronizing governing service failed")
	}

	// Ensure we have a StatefulSet running Thanos deployed.
	ssetClient := o.kclient.AppsV1().StatefulSets(tr.Namespace)
	obj, exists, err = o.ssetInf.GetIndexer().GetByKey(thanosKeyToStatefulSetKey(key))
	if err != nil {
		return errors.Wrap(err, "retrieving statefulset failed")
	}

	if !exists {
		sset, err := makeStatefulSet(tr, o.config, ruleConfigMapNames, "")
		if err != nil {
			return errors.Wrap(err, "making thanos statefulset config failed")
		}
		operator.SanitizeSTS(sset)
		if _, err := ssetClient.Create(sset); err != nil {
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

	_, err = ssetClient.Update(sset)
	sErr, ok := err.(*apierrors.StatusError)

	if ok && sErr.ErrStatus.Code == 422 && sErr.ErrStatus.Reason == metav1.StatusReasonInvalid {
		o.metrics.StsDeleteCreateCounter().Inc()
		level.Info(o.logger).Log("msg", "resolving illegal update of ThanosRuler StatefulSet", "details", sErr.ErrStatus.Details)
		propagationPolicy := metav1.DeletePropagationForeground
		if err := ssetClient.Delete(sset.GetName(), &metav1.DeleteOptions{PropagationPolicy: &propagationPolicy}); err != nil {
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

func (o *Operator) createCRDs() error {
	crds := []*extensionsobj.CustomResourceDefinition{
		k8sutil.NewCustomResourceDefinition(o.config.CrdKinds.ThanosRuler, monitoring.GroupName, o.config.Labels.LabelsMap, o.config.EnableValidation),
	}

	crdClient := o.crdclient.ApiextensionsV1beta1().CustomResourceDefinitions()

	for _, crd := range crds {
		oldCRD, err := crdClient.Get(crd.Name, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "getting CRD: %s", crd.Spec.Names.Kind)
		}
		if apierrors.IsNotFound(err) {
			if _, err := crdClient.Create(crd); err != nil {
				return errors.Wrapf(err, "creating CRD: %s", crd.Spec.Names.Kind)
			}
			level.Info(o.logger).Log("msg", "CRD created", "crd", crd.Spec.Names.Kind)
		}
		if err == nil {
			crd.ResourceVersion = oldCRD.ResourceVersion
			if _, err := crdClient.Update(crd); err != nil {
				return errors.Wrapf(err, "creating CRD: %s", crd.Spec.Names.Kind)
			}
			level.Info(o.logger).Log("msg", "CRD updated", "crd", crd.Spec.Names.Kind)
		}
	}

	crdListFuncs := []struct {
		name     string
		listFunc func(opts metav1.ListOptions) (runtime.Object, error)
	}{
		{
			monitoringv1.ThanosRulerKind,
			listwatch.MultiNamespaceListerWatcher(o.logger, o.config.Namespaces.ThanosRulerAllowList, o.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						return o.mclient.MonitoringV1().Prometheuses(namespace).List(options)
					},
				}
			}).List,
		},
		{
			monitoringv1.PrometheusRuleKind,
			listwatch.MultiNamespaceListerWatcher(o.logger, o.config.Namespaces.AllowList, o.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						return o.mclient.MonitoringV1().PrometheusRules(namespace).List(options)
					},
				}
			}).List,
		},
	}

	for _, crdListFunc := range crdListFuncs {
		err := k8sutil.WaitForCRDReady(crdListFunc.listFunc)
		if err != nil {
			return errors.Wrapf(err, "waiting for %v crd failed", crdListFunc.name)
		}
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
			"app":            thanosRulerLabel,
			thanosRulerLabel: name,
		})).String(),
	}
}

// ThanosRulerStatus evaluates the current status of a ThanosRuler deployment with
// respect to its specified resource object. It return the status and a list of
// pods that are not updated.
func ThanosRulerStatus(kclient kubernetes.Interface, tr *monitoringv1.ThanosRuler) (*monitoringv1.ThanosRulerStatus, []v1.Pod, error) {
	res := &monitoringv1.ThanosRulerStatus{Paused: tr.Spec.Paused}

	pods, err := kclient.CoreV1().Pods(tr.Namespace).List(ListOptions(tr.Name))
	if err != nil {
		return nil, nil, errors.Wrap(err, "retrieving pods of failed")
	}
	sset, err := kclient.AppsV1().StatefulSets(tr.Namespace).Get(statefulSetNameFromThanosName(tr.Name), metav1.GetOptions{})
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

	err = cache.ListAll(o.thanosRulerInf.GetStore(), labels.Everything(), func(obj interface{}) {
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
