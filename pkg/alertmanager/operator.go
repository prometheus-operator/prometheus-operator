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

package alertmanager

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/listwatch"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/mitchellh/hashstructure"
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
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	resyncPeriod = 5 * time.Minute
)

var (
	managedByOperatorLabel      = "managed-by"
	managedByOperatorLabelValue = "prometheus-operator"
	managedByOperatorLabels     = map[string]string{
		managedByOperatorLabel: managedByOperatorLabelValue,
	}
)

// Operator manages life cycle of Alertmanager deployments and
// monitoring configurations.
type Operator struct {
	kclient kubernetes.Interface
	mclient monitoringclient.Interface
	logger  log.Logger

	nsAlrtInf    cache.SharedIndexInformer
	nsAlrtCfgInf cache.SharedIndexInformer

	alrtInfs    *informers.ForResource
	alrtCfgInfs *informers.ForResource
	secrInfs    *informers.ForResource
	ssetInfs    *informers.ForResource

	queue workqueue.RateLimitingInterface

	metrics *operator.Metrics

	config Config
}

type Config struct {
	Host                         string
	LocalHost                    string
	ClusterDomain                string
	ReloaderConfig               operator.ReloaderConfig
	AlertmanagerDefaultBaseImage string
	Namespaces                   operator.Namespaces
	Labels                       operator.Labels
	AlertManagerSelector         string
	SecretListWatchSelector      string
}

// New creates a new controller.
func New(ctx context.Context, c operator.Config, logger log.Logger, r prometheus.Registerer) (*Operator, error) {
	cfg, err := k8sutil.NewClusterConfig(c.Host, c.TLSInsecure, &c.TLSConfig)
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

	o := &Operator{
		kclient: client,
		mclient: mclient,
		logger:  logger,
		queue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "alertmanager"),
		metrics: operator.NewMetrics("alertmanager", r),
		config: Config{
			Host:                         c.Host,
			LocalHost:                    c.LocalHost,
			ClusterDomain:                c.ClusterDomain,
			ReloaderConfig:               c.ReloaderConfig,
			AlertmanagerDefaultBaseImage: c.AlertmanagerDefaultBaseImage,
			Namespaces:                   c.Namespaces,
			Labels:                       c.Labels,
			AlertManagerSelector:         c.AlertManagerSelector,
			SecretListWatchSelector:      c.SecretListWatchSelector,
		},
	}

	if err := o.bootstrap(ctx); err != nil {
		return nil, err
	}

	return o, nil
}

func (c *Operator) bootstrap(ctx context.Context) error {
	var err error

	if _, err := labels.Parse(c.config.AlertManagerSelector); err != nil {
		return errors.Wrap(err, "can not parse alertmanager selector value")
	}

	c.alrtInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.config.Namespaces.AlertmanagerAllowList,
			c.config.Namespaces.DenyList,
			c.mclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = c.config.AlertManagerSelector
			},
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.AlertmanagerName),
	)
	if err != nil {
		return errors.Wrap(err, "error creating alertmanager informers")
	}

	var alertmanagerStores []cache.Store
	for _, informer := range c.alrtInfs.GetInformers() {
		alertmanagerStores = append(alertmanagerStores, informer.Informer().GetStore())
	}
	c.metrics.MustRegister(newAlertmanagerCollectorForStores(alertmanagerStores...))

	c.alrtCfgInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.config.Namespaces.AllowList,
			c.config.Namespaces.DenyList,
			c.mclient,
			resyncPeriod,
			nil,
		),
		monitoringv1alpha1.SchemeGroupVersion.WithResource(monitoringv1alpha1.AlertmanagerConfigName),
	)
	if err != nil {
		return errors.Wrap(err, "error creating alertmanagerconfig informers")
	}

	secretListWatchSelector, err := fields.ParseSelector(c.config.SecretListWatchSelector)
	if err != nil {
		return errors.Wrap(err, "can not parse secrets selector value")
	}
	c.secrInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			c.config.Namespaces.AllowList,
			c.config.Namespaces.DenyList,
			c.kclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.FieldSelector = secretListWatchSelector.String()
			},
		),
		v1.SchemeGroupVersion.WithResource("secrets"),
	)
	if err != nil {
		return errors.Wrap(err, "error creating secret informers")
	}

	c.ssetInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			c.config.Namespaces.AlertmanagerAllowList,
			c.config.Namespaces.DenyList,
			c.kclient,
			resyncPeriod,
			nil,
		),
		appsv1.SchemeGroupVersion.WithResource("statefulsets"),
	)
	if err != nil {
		return errors.Wrap(err, "error creating statefulset informers")
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
	c.nsAlrtCfgInf = newNamespaceInformer(c, c.config.Namespaces.AllowList)
	if listwatch.IdenticalNamespaces(c.config.Namespaces.AllowList, c.config.Namespaces.AlertmanagerAllowList) {
		c.nsAlrtInf = c.nsAlrtCfgInf
	} else {
		c.nsAlrtInf = newNamespaceInformer(c, c.config.Namespaces.AlertmanagerAllowList)
	}

	return nil
}

// waitForCacheSync waits for the informers' caches to be synced.
func (c *Operator) waitForCacheSync(ctx context.Context) error {
	for _, infs := range []struct {
		name                 string
		informersForResource *informers.ForResource
	}{
		{"Alertmanager", c.alrtInfs},
		{"AlertmanagerConfig", c.alrtCfgInfs},
		{"Secret", c.secrInfs},
		{"StatefulSet", c.ssetInfs},
	} {
		for _, inf := range infs.informersForResource.GetInformers() {
			if !operator.WaitForNamedCacheSync(ctx, "alertmanager", log.With(c.logger, "informer", infs.name), inf.Informer()) {
				return errors.Errorf("failed to sync cache for %s informer", infs.name)
			}
		}
	}

	for _, inf := range []struct {
		name     string
		informer cache.SharedIndexInformer
	}{
		{"AlertmanagerNamespace", c.nsAlrtInf},
		{"AlertmanagerConfigNamespace", c.nsAlrtCfgInf},
	} {
		if !operator.WaitForNamedCacheSync(ctx, "alertmanager", log.With(c.logger, "informer", inf.name), inf.informer) {
			return errors.Errorf("failed to sync cache for %s informer", inf.name)
		}
	}

	level.Info(c.logger).Log("msg", "successfully synced all caches")
	return nil
}

// addHandlers adds the eventhandlers to the informers.
func (c *Operator) addHandlers() {
	c.alrtInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleAlertmanagerAdd,
		DeleteFunc: c.handleAlertmanagerDelete,
		UpdateFunc: c.handleAlertmanagerUpdate,
	})
	c.alrtCfgInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleAlertmanagerConfigAdd,
		DeleteFunc: c.handleAlertmanagerConfigDelete,
		UpdateFunc: c.handleAlertmanagerConfigUpdate,
	})
	c.secrInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleSecretAdd,
		DeleteFunc: c.handleSecretDelete,
		UpdateFunc: c.handleSecretUpdate,
	})
	c.ssetInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleStatefulSetAdd,
		DeleteFunc: c.handleStatefulSetDelete,
		UpdateFunc: c.handleStatefulSetUpdate,
	})

	// The controller needs to watch the namespaces in which the
	// alertmanagerconfigs live because a label change on a namespace may
	// trigger a configuration change.
	// It doesn't need to watch on addition/deletion though because it's
	// already covered by the event handlers on alertmanagerconfigs.
	c.nsAlrtCfgInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: c.handleNamespaceUpdate,
	})
}

func (c *Operator) handleAlertmanagerConfigAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "AlertmanagerConfig added")
		c.metrics.TriggerByCounter(monitoringv1alpha1.AlertmanagerConfigKind, "add").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleAlertmanagerConfigUpdate(old, cur interface{}) {
	if old.(*monitoringv1alpha1.AlertmanagerConfig).ResourceVersion == cur.(*monitoringv1alpha1.AlertmanagerConfig).ResourceVersion {
		return
	}

	o, ok := c.getObject(cur)
	if ok {
		level.Debug(c.logger).Log("msg", "AlertmanagerConfig updated")
		c.metrics.TriggerByCounter(monitoringv1alpha1.AlertmanagerConfigKind, "update").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleAlertmanagerConfigDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "AlertmanagerConfig delete")
		c.metrics.TriggerByCounter(monitoringv1alpha1.AlertmanagerConfigKind, "delete").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Do we need to enqueue secrets just for the namespace or in general?
func (c *Operator) handleSecretDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "Secret deleted")
		c.metrics.TriggerByCounter("Secret", "delete").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleSecretUpdate(old, cur interface{}) {
	if old.(*v1.Secret).ResourceVersion == cur.(*v1.Secret).ResourceVersion {
		return
	}

	o, ok := c.getObject(cur)
	if ok {
		level.Debug(c.logger).Log("msg", "Secret updated")
		c.metrics.TriggerByCounter("Secret", "update").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleSecretAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "Secret added")
		c.metrics.TriggerByCounter("Secret", "add").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// enqueueForNamespace enqueues all Alertmanager object keys that belong to the
// given namespace or select objects in the given namespace.
func (c *Operator) enqueueForNamespace(nsName string) {
	nsObject, exists, err := c.nsAlrtCfgInf.GetStore().GetByKey(nsName)
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "get namespace to enqueue Alertmanager instances failed",
			"err", err,
		)
		return
	}
	if !exists {
		level.Error(c.logger).Log(
			"msg", fmt.Sprintf("get namespace to enqueue Alertmanager instances failed: namespace %q does not exist", nsName),
		)
		return
	}
	ns := nsObject.(*v1.Namespace)

	err = c.alrtInfs.ListAll(labels.Everything(), func(obj interface{}) {
		// Check for Alertmanager instances in the namespace.
		am := obj.(*monitoringv1.Alertmanager)
		if am.Namespace == nsName {
			c.enqueue(am)
			return
		}

		// Check for Alertmanager instances selecting AlertmanagerConfigs in
		// the namespace.
		acNSSelector, err := metav1.LabelSelectorAsSelector(am.Spec.AlertmanagerConfigNamespaceSelector)
		if err != nil {
			level.Error(c.logger).Log(
				"msg", fmt.Sprintf("failed to convert AlertmanagerConfigNamespaceSelector of %q to selector", am.Name),
				"err", err,
			)
			return
		}

		if acNSSelector.Matches(labels.Set(ns.Labels)) {
			c.enqueue(am)
			return
		}
	})
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "listing all Alertmanager instances from cache failed",
			"err", err,
		)
	}
}

// Run the controller.
func (c *Operator) Run(ctx context.Context) error {
	defer c.queue.ShutDown()

	errChan := make(chan error)
	go func() {
		v, err := c.kclient.Discovery().ServerVersion()
		if err != nil {
			errChan <- errors.Wrap(err, "communicating with server failed")
			return
		}
		level.Info(c.logger).Log("msg", "connection established", "cluster-version", v)
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
		level.Info(c.logger).Log("msg", "CRD API endpoints ready")
	case <-ctx.Done():
		return nil
	}

	go c.worker(ctx)

	go c.alrtInfs.Start(ctx.Done())
	go c.alrtCfgInfs.Start(ctx.Done())
	go c.secrInfs.Start(ctx.Done())
	go c.ssetInfs.Start(ctx.Done())
	go c.nsAlrtCfgInf.Run(ctx.Done())
	if c.nsAlrtInf != c.nsAlrtCfgInf {
		go c.nsAlrtInf.Run(ctx.Done())
	}
	if err := c.waitForCacheSync(ctx); err != nil {
		return err
	}
	c.addHandlers()

	c.metrics.Ready().Set(1)
	<-ctx.Done()
	return nil
}

func (c *Operator) keyFunc(obj interface{}) (string, bool) {
	k, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		level.Error(c.logger).Log("msg", "creating key failed", "err", err)
		return k, false
	}
	return k, true
}

func (c *Operator) getObject(obj interface{}) (metav1.Object, bool) {
	ts, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		obj = ts.Obj
	}

	o, err := meta.Accessor(obj)
	if err != nil {
		level.Error(c.logger).Log("msg", "get object failed", "err", err)
		return nil, false
	}
	return o, true
}

// enqueue adds a key to the queue. If obj is a key already it gets added
// directly. Otherwise, the key is extracted via keyFunc.
func (c *Operator) enqueue(obj interface{}) {
	if obj == nil {
		return
	}

	key, ok := obj.(string)
	if !ok {
		key, ok = c.keyFunc(obj)
		if !ok {
			return
		}
	}

	c.queue.Add(key)
}

// worker runs a worker thread that just dequeues items, processes them
// and marks them done. It enforces that the syncHandler is never invoked
// concurrently with the same key.
func (c *Operator) worker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *Operator) processNextWorkItem(ctx context.Context) bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	c.metrics.ReconcileCounter().Inc()
	err := c.sync(ctx, key.(string))
	c.metrics.SetSyncStatus(key.(string), err == nil)
	if err == nil {
		c.queue.Forget(key)
		return true
	}

	c.metrics.ReconcileErrorsCounter().Inc()
	utilruntime.HandleError(errors.Wrap(err, fmt.Sprintf("Sync %q failed", key)))
	c.queue.AddRateLimited(key)

	return true
}

func (c *Operator) alertmanagerForStatefulSet(sset interface{}) *monitoringv1.Alertmanager {
	key, ok := c.keyFunc(sset)
	if !ok {
		return nil
	}

	match, aKey := statefulSetKeyToAlertmanagerKey(key)
	if !match {
		level.Debug(c.logger).Log("msg", "StatefulSet key did not match an Alertmanager key format", "key", key)
		return nil
	}

	a, err := c.alrtInfs.Get(aKey)
	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		level.Error(c.logger).Log("msg", "Alertmanager lookup failed", "err", err)
		return nil
	}

	return a.(*monitoringv1.Alertmanager)
}

func statefulSetNameFromAlertmanagerName(name string) string {
	return "alertmanager-" + name
}

func statefulSetKeyToAlertmanagerKey(key string) (bool, string) {
	r := regexp.MustCompile("^(.+)/alertmanager-(.+)$")

	matches := r.FindAllStringSubmatch(key, 2)
	if len(matches) != 1 {
		return false, ""
	}
	if len(matches[0]) != 3 {
		return false, ""
	}
	return true, matches[0][1] + "/" + matches[0][2]
}

func alertmanagerKeyToStatefulSetKey(key string) string {
	keyParts := strings.Split(key, "/")
	return keyParts[0] + "/alertmanager-" + keyParts[1]
}

func (c *Operator) handleAlertmanagerAdd(obj interface{}) {
	key, ok := c.keyFunc(obj)
	if !ok {
		return
	}

	level.Debug(c.logger).Log("msg", "Alertmanager added", "key", key)
	c.metrics.TriggerByCounter(monitoringv1.AlertmanagersKind, "add").Inc()
	checkAlertmanagerSpecDeprecation(key, obj.(*monitoringv1.Alertmanager), c.logger)
	c.enqueue(key)
}

func (c *Operator) handleAlertmanagerDelete(obj interface{}) {
	key, ok := c.keyFunc(obj)
	if !ok {
		return
	}

	level.Debug(c.logger).Log("msg", "Alertmanager deleted", "key", key)
	c.metrics.TriggerByCounter(monitoringv1.AlertmanagersKind, "delete").Inc()
	c.enqueue(key)
}

func (c *Operator) handleAlertmanagerUpdate(old, cur interface{}) {
	if old.(*monitoringv1.Alertmanager).ResourceVersion == cur.(*monitoringv1.Alertmanager).ResourceVersion {
		return
	}

	key, ok := c.keyFunc(cur)
	if !ok {
		return
	}

	level.Debug(c.logger).Log("msg", "Alertmanager updated", "key", key)
	c.metrics.TriggerByCounter(monitoringv1.AlertmanagersKind, "update").Inc()
	checkAlertmanagerSpecDeprecation(key, cur.(*monitoringv1.Alertmanager), c.logger)
	c.enqueue(key)
}

func (c *Operator) handleStatefulSetDelete(obj interface{}) {
	if a := c.alertmanagerForStatefulSet(obj); a != nil {
		level.Debug(c.logger).Log("msg", "StatefulSet delete")
		c.metrics.TriggerByCounter("StatefulSet", "delete").Inc()

		c.enqueue(a)
	}
}

func (c *Operator) handleStatefulSetAdd(obj interface{}) {
	if a := c.alertmanagerForStatefulSet(obj); a != nil {
		level.Debug(c.logger).Log("msg", "StatefulSet added")
		c.metrics.TriggerByCounter("StatefulSet", "add").Inc()

		c.enqueue(a)
	}
}

func (c *Operator) handleStatefulSetUpdate(oldo, curo interface{}) {
	old := oldo.(*appsv1.StatefulSet)
	cur := curo.(*appsv1.StatefulSet)

	level.Debug(c.logger).Log("msg", "update handler", "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	// Periodic resync may resend the deployment without changes in-between.
	// Also breaks loops created by updating the resource ourselves.
	if old.ResourceVersion == cur.ResourceVersion {
		return
	}

	// Wake up Alertmanager resource the deployment belongs to.
	if a := c.alertmanagerForStatefulSet(cur); a != nil {
		level.Debug(c.logger).Log("msg", "StatefulSet updated")
		c.metrics.TriggerByCounter("StatefulSet", "update").Inc()

		c.enqueue(a)
	}
}

func (c *Operator) handleNamespaceUpdate(oldo, curo interface{}) {
	old := oldo.(*v1.Namespace)
	cur := curo.(*v1.Namespace)

	level.Debug(c.logger).Log("msg", "update handler", "namespace", cur.GetName(), "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	// Periodic resync may resend the Namespace without changes
	// in-between.
	if old.ResourceVersion == cur.ResourceVersion {
		return
	}

	level.Debug(c.logger).Log("msg", "Namespace updated", "namespace", cur.GetName())
	c.metrics.TriggerByCounter("Namespace", "update").Inc()

	// Check for Alertmanager instances selecting AlertmanagerConfigs in the namespace.
	err := c.alrtInfs.ListAll(labels.Everything(), func(obj interface{}) {
		a := obj.(*monitoringv1.Alertmanager)

		sync, err := k8sutil.LabelSelectionHasChanged(old.Labels, cur.Labels, a.Spec.AlertmanagerConfigNamespaceSelector)
		if err != nil {
			level.Error(c.logger).Log(
				"err", err,
				"name", a.Name,
				"namespace", a.Namespace,
			)
			return
		}

		if sync {
			c.enqueue(a)
		}
	})
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "listing all Alertmanager instances from cache failed",
			"err", err,
		)
	}
}

func (c *Operator) sync(ctx context.Context, key string) error {
	aobj, err := c.alrtInfs.Get(key)

	if apierrors.IsNotFound(err) {
		c.metrics.ForgetObject(key)
		// Dependent resources are cleaned up by K8s via OwnerReferences
		return nil
	}
	if err != nil {
		return err
	}

	am := aobj.(*monitoringv1.Alertmanager)
	am = am.DeepCopy()
	am.APIVersion = monitoringv1.SchemeGroupVersion.String()
	am.Kind = monitoringv1.AlertmanagersKind

	if am.Spec.Paused {
		return nil
	}

	logger := log.With(c.logger, "key", key)
	level.Info(logger).Log("msg", "sync alertmanager")

	assetStore := assets.NewStore(c.kclient.CoreV1(), c.kclient.CoreV1())

	if err := c.provisionAlertmanagerConfiguration(ctx, am, assetStore); err != nil {
		return errors.Wrap(err, "provision alertmanager configuration")
	}

	if err := c.createOrUpdateTLSAssetSecret(ctx, am, assetStore); err != nil {
		return errors.Wrap(err, "creating tls asset secret failed")
	}

	// Create governing service if it doesn't exist.
	svcClient := c.kclient.CoreV1().Services(am.Namespace)
	if err = k8sutil.CreateOrUpdateService(ctx, svcClient, makeStatefulSetService(am, c.config)); err != nil {
		return errors.Wrap(err, "synchronizing governing service failed")
	}

	obj, err := c.ssetInfs.Get(alertmanagerKeyToStatefulSetKey(key))
	exists := !apierrors.IsNotFound(err)
	if err != nil && exists {
		return errors.Wrap(err, "failed to retrieve statefulset")
	}

	oldSpec := appsv1.StatefulSetSpec{}
	if obj != nil {
		ss := obj.(*appsv1.StatefulSet)
		oldSpec = ss.Spec
	}

	newSSetInputHash, err := createSSetInputHash(*am, c.config, oldSpec)
	if err != nil {
		return err
	}

	sset, err := makeStatefulSet(am, c.config, newSSetInputHash)
	if err != nil {
		return errors.Wrap(err, "failed to make statefulset")
	}
	operator.SanitizeSTS(sset)

	ssetClient := c.kclient.AppsV1().StatefulSets(am.Namespace)

	var oldSSetInputHash string
	if obj != nil {
		oldSSetInputHash = obj.(*appsv1.StatefulSet).ObjectMeta.Annotations[sSetInputHashName]
	}
	if newSSetInputHash == oldSSetInputHash {
		level.Debug(logger).Log("msg", "new statefulset generation inputs match current, skipping any actions")
		return nil
	}

	if !exists {
		level.Debug(logger).Log("msg", "no current statefulset found")
		level.Debug(logger).Log("msg", "creating statefulset")
		if _, err := ssetClient.Create(ctx, sset, metav1.CreateOptions{}); err != nil {
			return errors.Wrap(err, "creating statefulset failed")
		}
		return nil
	}

	err = k8sutil.UpdateStatefulSet(ctx, ssetClient, sset)
	sErr, ok := err.(*apierrors.StatusError)

	if ok && sErr.ErrStatus.Code == 422 && sErr.ErrStatus.Reason == metav1.StatusReasonInvalid {
		c.metrics.StsDeleteCreateCounter().Inc()

		// Gather only reason for failed update
		failMsg := make([]string, len(sErr.ErrStatus.Details.Causes))
		for i, cause := range sErr.ErrStatus.Details.Causes {
			failMsg[i] = cause.Message
		}

		level.Info(logger).Log("msg", "recreating AlertManager StatefulSet because the update operation wasn't possible", "reason", strings.Join(failMsg, ", "))
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

func createSSetInputHash(a monitoringv1.Alertmanager, c Config, s appsv1.StatefulSetSpec) (string, error) {
	hash, err := hashstructure.Hash(struct {
		A monitoringv1.Alertmanager
		C Config
		S appsv1.StatefulSetSpec
	}{a, c, s},
		nil,
	)
	if err != nil {
		return "", errors.Wrap(
			err,
			"failed to calculate combined hash of Alertmanager CRD and config",
		)
	}

	return fmt.Sprintf("%d", hash), nil
}

func (c *Operator) provisionAlertmanagerConfiguration(ctx context.Context, am *monitoringv1.Alertmanager, store *assets.Store) error {
	namespacedLogger := log.With(c.logger, "alertmanager", am.Name, "namespace", am.Namespace)

	// Validate AlertManager Config Inputs at AlertManager CRD level
	if err := validateConfigInputs(am); err != nil {
		return err
	}

	secretName := defaultConfigSecretName(am.Name)
	if am.Spec.ConfigSecret != "" {
		secretName = am.Spec.ConfigSecret
	}

	// Tentatively retrieve the secret containing the user-provided Alertmanager
	// configuration.
	secret, err := c.kclient.CoreV1().Secrets(am.Namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "get base configuration secret")
	}

	var secretData map[string][]byte
	if secret != nil {
		secretData = secret.Data
	}

	rawBaseConfig := []byte(`route:
  receiver: 'null'
receivers:
- name: 'null'`)
	if len(secretData[alertmanagerConfigFile]) > 0 {
		rawBaseConfig = secretData[alertmanagerConfigFile]
	} else {
		if secret == nil {
			level.Info(namespacedLogger).Log("msg", "base config secret not found", "secret", secretName)
		} else {
			level.Info(namespacedLogger).
				Log("msg", "key not found in base config secret", "secret", secretName, "key", alertmanagerConfigFile)
		}
	}

	// If no AlertmanagerConfig selectors are configured, the user wants to
	// manage configuration themselves.
	if am.Spec.AlertmanagerConfigSelector == nil {
		level.Debug(namespacedLogger).
			Log("msg", "no AlertmanagerConfig selector specified, copying base config as-is",
				"base config secret", secretName, "mounted config secret", generatedConfigSecretName(am.Name))

		err = c.createOrUpdateGeneratedConfigSecret(ctx, am, rawBaseConfig, secretData)
		if err != nil {
			return errors.Wrap(err, "create or update generated config secret failed")
		}
		return nil
	}

	baseConfig, err := loadCfg(string(rawBaseConfig))
	if err != nil {
		return errors.Wrap(err, "base config from Secret could not be parsed")
	}

	amVersion := operator.StringValOrDefault(am.Spec.Version, operator.DefaultAlertmanagerVersion)
	version, err := semver.ParseTolerant(amVersion)
	if err != nil {
		return errors.Wrap(err, "failed to parse alertmanager version")
	}

	amConfigs, err := c.selectAlertmanagerConfigs(ctx, am, store)
	if err != nil {
		return errors.Wrap(err, "selecting AlertmanagerConfigs failed")
	}

	generator := newConfigGenerator(namespacedLogger, version, store)
	generatedConfig, err := generator.generateConfig(ctx, *baseConfig, amConfigs)
	if err != nil {
		return errors.Wrap(err, "generating Alertmanager config yaml failed")
	}

	err = c.createOrUpdateGeneratedConfigSecret(ctx, am, generatedConfig, secretData)
	if err != nil {
		return errors.Wrap(err, "create or update generated config secret failed")
	}

	return nil
}

func (c *Operator) createOrUpdateGeneratedConfigSecret(ctx context.Context, am *monitoringv1.Alertmanager, conf []byte, additionalData map[string][]byte) error {
	boolTrue := true
	sClient := c.kclient.CoreV1().Secrets(am.Namespace)

	generatedConfigSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   generatedConfigSecretName(am.Name),
			Labels: c.config.Labels.Merge(managedByOperatorLabels),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         am.APIVersion,
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               am.Kind,
					Name:               am.Name,
					UID:                am.UID,
				},
			},
		},
		Data: map[string][]byte{},
	}

	for k, v := range additionalData {
		generatedConfigSecret.Data[k] = v
	}
	generatedConfigSecret.Data[alertmanagerConfigFile] = conf

	err := k8sutil.CreateOrUpdateSecret(ctx, sClient, generatedConfigSecret)
	if err != nil {
		return errors.Wrap(err, "failed to update generated config secret")
	}

	return nil
}

func (c *Operator) selectAlertmanagerConfigs(ctx context.Context, am *monitoringv1.Alertmanager, store *assets.Store) (map[string]*monitoringv1alpha1.AlertmanagerConfig, error) {
	namespaces := []string{}

	// If 'AlertmanagerConfigNamespaceSelector' is nil, only check own namespace.
	if am.Spec.AlertmanagerConfigNamespaceSelector == nil {
		namespaces = append(namespaces, am.Namespace)

		level.Debug(c.logger).Log("msg", "selecting AlertmanagerConfigs from alertmanager's namespace", "namespace", am.Namespace, "alertmanager", am.Name)
	} else {
		amConfigNSSelector, err := metav1.LabelSelectorAsSelector(am.Spec.AlertmanagerConfigNamespaceSelector)
		if err != nil {
			return nil, err
		}

		err = cache.ListAll(c.nsAlrtCfgInf.GetStore(), amConfigNSSelector, func(obj interface{}) {
			namespaces = append(namespaces, obj.(*v1.Namespace).Name)
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list namespaces")
		}

		level.Debug(c.logger).Log("msg", "filtering namespaces to select AlertmanagerConfigs from", "namespaces", strings.Join(namespaces, ","), "namespace", am.Namespace, "alertmanager", am.Name)
	}

	// Selectors (<namespace>/<name>) might overlap. Deduplicate them along the keyFunc.
	amConfigs := make(map[string]*monitoringv1alpha1.AlertmanagerConfig)

	amConfigSelector, err := metav1.LabelSelectorAsSelector(am.Spec.AlertmanagerConfigSelector)
	if err != nil {
		return nil, err
	}

	for _, ns := range namespaces {
		err := c.alrtCfgInfs.ListAllByNamespace(ns, amConfigSelector, func(obj interface{}) {
			k, ok := c.keyFunc(obj)
			if ok {
				amConfigs[k] = obj.(*monitoringv1alpha1.AlertmanagerConfig)
			}
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list alertmanager configs in namespace %s", ns)
		}
	}

	var rejected int
	res := make(map[string]*monitoringv1alpha1.AlertmanagerConfig, len(amConfigs))

	for namespaceAndName, amc := range amConfigs {
		if err := checkAlertmanagerConfig(ctx, amc, store); err != nil {
			rejected++
			level.Warn(c.logger).Log(
				"msg", "skipping alertmanagerconfig",
				"error", err.Error(),
				"alertmanagerconfig", namespaceAndName,
				"namespace", am.Namespace,
				"alertmanager", am.Name,
			)
			continue
		}

		res[namespaceAndName] = amc
	}

	amcKeys := []string{}
	for k := range res {
		amcKeys = append(amcKeys, k)
	}
	level.Debug(c.logger).Log("msg", "selected AlertmanagerConfigs", "alertmanagerconfigs", strings.Join(amcKeys, ","), "namespace", am.Namespace, "prometheus", am.Name)

	if amKey, ok := c.keyFunc(am); ok {
		c.metrics.SetSelectedResources(amKey, monitoringv1alpha1.AlertmanagerConfigKind, len(res))
		c.metrics.SetRejectedResources(amKey, monitoringv1alpha1.AlertmanagerConfigKind, rejected)
	}

	return res, nil
}

// checkAlertmanagerConfig verifies that an AlertmanagerConfig object is valid
// and has no missing references to other objects.
func checkAlertmanagerConfig(ctx context.Context, amc *monitoringv1alpha1.AlertmanagerConfig, store *assets.Store) error {
	receiverNames, err := checkReceivers(ctx, amc, store)
	if err != nil {
		return err
	}

	return validateAlertManagerRoutes(amc.Spec.Route, receiverNames, true)
}

func checkReceivers(ctx context.Context, amc *monitoringv1alpha1.AlertmanagerConfig, store *assets.Store) (map[string]struct{}, error) {
	receiverNames, err := validateReceivers(amc.Spec.Receivers)
	if err != nil {
		return nil, errors.Wrap(err, "checkReceivers: failed to validateReceivers")
	}

	for i, receiver := range amc.Spec.Receivers {
		amcKey := fmt.Sprintf("alertmanagerConfig/%s/%s/%d", amc.GetNamespace(), amc.GetName(), i)

		err = checkPagerDutyConfigs(ctx, receiver.PagerDutyConfigs, amc.GetNamespace(), amcKey, store)
		if err != nil {
			return nil, err
		}

		err = checkOpsGenieConfigs(ctx, receiver.OpsGenieConfigs, amc.GetNamespace(), amcKey, store)
		if err != nil {
			return nil, err
		}
		err = checkSlackConfigs(ctx, receiver.SlackConfigs, amc.GetNamespace(), amcKey, store)
		if err != nil {
			return nil, err
		}

		err = checkWebhookConfigs(ctx, receiver.WebhookConfigs, amc.GetNamespace(), amcKey, store)
		if err != nil {
			return nil, err
		}

		err = checkWechatConfigs(ctx, receiver.WeChatConfigs, amc.GetNamespace(), amcKey, store)
		if err != nil {
			return nil, err
		}

		err = checkEmailConfigs(ctx, receiver.EmailConfigs, amc.GetNamespace(), amcKey, store)
		if err != nil {
			return nil, err
		}

		err = checkVictorOpsConfigs(ctx, receiver.VictorOpsConfigs, amc.GetNamespace(), amcKey, store)
		if err != nil {
			return nil, err
		}

		err = checkPushoverConfigs(ctx, receiver.PushoverConfigs, amc.GetNamespace(), amcKey, store)
		if err != nil {
			return nil, err
		}
	}

	return receiverNames, nil
}

func checkPagerDutyConfigs(ctx context.Context, configs []monitoringv1alpha1.PagerDutyConfig, namespace string, key string, store *assets.Store) error {
	for i, config := range configs {
		pagerDutyConfigKey := fmt.Sprintf("%s/pagerduty/%d", key, i)

		if config.RoutingKey != nil {
			if _, err := store.GetSecretKey(ctx, namespace, *config.RoutingKey); err != nil {
				return err
			}
		}

		if config.ServiceKey != nil {
			if _, err := store.GetSecretKey(ctx, namespace, *config.ServiceKey); err != nil {
				return err
			}
		}

		if err := configureHTTPConfigInStore(ctx, config.HTTPConfig, namespace, pagerDutyConfigKey, store); err != nil {
			return err
		}
	}

	return nil
}

func checkOpsGenieConfigs(ctx context.Context, configs []monitoringv1alpha1.OpsGenieConfig, namespace string, key string, store *assets.Store) error {
	for i, config := range configs {
		opsgenieConfigKey := fmt.Sprintf("%s/opsgenie/%d", key, i)

		if config.APIKey != nil {
			if _, err := store.GetSecretKey(ctx, namespace, *config.APIKey); err != nil {
				return err
			}
		}

		if err := configureHTTPConfigInStore(ctx, config.HTTPConfig, namespace, opsgenieConfigKey, store); err != nil {
			return err
		}
	}

	return nil
}

func checkSlackConfigs(ctx context.Context, configs []monitoringv1alpha1.SlackConfig, namespace string, key string, store *assets.Store) error {
	for i, config := range configs {
		slackConfigKey := fmt.Sprintf("%s/slack/%d", key, i)

		if config.APIURL != nil {
			if _, err := store.GetSecretKey(ctx, namespace, *config.APIURL); err != nil {
				return err
			}
		}

		if err := configureHTTPConfigInStore(ctx, config.HTTPConfig, namespace, slackConfigKey, store); err != nil {
			return err
		}
	}

	return nil
}

func checkWebhookConfigs(ctx context.Context, configs []monitoringv1alpha1.WebhookConfig, namespace string, key string, store *assets.Store) error {
	for i, config := range configs {
		webhookConfigKey := fmt.Sprintf("%s/webhook/%d", key, i)

		if config.URLSecret != nil {
			if _, err := store.GetSecretKey(ctx, namespace, *config.URLSecret); err != nil {
				return err
			}
		}

		if err := configureHTTPConfigInStore(ctx, config.HTTPConfig, namespace, webhookConfigKey, store); err != nil {
			return err
		}
	}

	return nil
}

func checkWechatConfigs(ctx context.Context, configs []monitoringv1alpha1.WeChatConfig, namespace string, key string, store *assets.Store) error {
	for i, config := range configs {
		wechatConfigKey := fmt.Sprintf("%s/wechat/%d", key, i)

		if config.APISecret != nil {
			if _, err := store.GetSecretKey(ctx, namespace, *config.APISecret); err != nil {
				return err
			}
		}

		if err := configureHTTPConfigInStore(ctx, config.HTTPConfig, namespace, wechatConfigKey, store); err != nil {
			return err
		}
	}

	return nil
}

func checkEmailConfigs(ctx context.Context, configs []monitoringv1alpha1.EmailConfig, namespace string, key string, store *assets.Store) error {
	for _, config := range configs {
		if config.AuthPassword != nil {
			if _, err := store.GetSecretKey(ctx, namespace, *config.AuthPassword); err != nil {
				return err
			}
		}
		if config.AuthSecret != nil {
			if _, err := store.GetSecretKey(ctx, namespace, *config.AuthSecret); err != nil {
				return err
			}
		}

		if err := store.AddSafeTLSConfig(ctx, namespace, config.TLSConfig); err != nil {
			return err
		}
	}

	return nil
}

func checkVictorOpsConfigs(ctx context.Context, configs []monitoringv1alpha1.VictorOpsConfig, namespace string, key string, store *assets.Store) error {
	for i, config := range configs {

		if config.APIKey != nil {
			if _, err := store.GetSecretKey(ctx, namespace, *config.APIKey); err != nil {
				return err
			}
		}

		victoropsConfigKey := fmt.Sprintf("%s/victorops/%d", key, i)
		if err := configureHTTPConfigInStore(ctx, config.HTTPConfig, namespace, victoropsConfigKey, store); err != nil {
			return err
		}
	}

	return nil
}

func checkPushoverConfigs(ctx context.Context, configs []monitoringv1alpha1.PushoverConfig, namespace string, key string, store *assets.Store) error {
	checkSecret := func(secret *v1.SecretKeySelector, name string) error {
		if secret == nil {
			return errors.Errorf("mandatory field %s is empty", name)
		}
		s, err := store.GetSecretKey(ctx, namespace, *secret)
		if err != nil {
			return err
		}
		if s == "" {
			return errors.New("mandatory field userKey is empty")
		}
		return nil
	}

	for i, config := range configs {

		if err := checkSecret(config.UserKey, "userKey"); err != nil {
			return err
		}
		if err := checkSecret(config.Token, "token"); err != nil {
			return err
		}

		pushoverConfigKey := fmt.Sprintf("%s/pushover/%d", key, i)
		if err := configureHTTPConfigInStore(ctx, config.HTTPConfig, namespace, pushoverConfigKey, store); err != nil {
			return err
		}
	}

	return nil
}

// configureHTTPConfigInStore configure the asset store for HTTPConfigs.
func configureHTTPConfigInStore(ctx context.Context, httpConfig *monitoringv1alpha1.HTTPConfig, namespace string, key string, store *assets.Store) error {
	if httpConfig == nil {
		return nil
	}

	var err error
	if httpConfig.BearerTokenSecret != nil {
		if err = store.AddBearerToken(ctx, namespace, *httpConfig.BearerTokenSecret, key); err != nil {
			return err
		}
	}

	if err = store.AddSafeAuthorizationCredentials(ctx, namespace, httpConfig.Authorization, key); err != nil {
		return err
	}

	if err = store.AddBasicAuth(ctx, namespace, httpConfig.BasicAuth, key); err != nil {
		return err
	}

	return store.AddSafeTLSConfig(ctx, namespace, httpConfig.TLSConfig)
}

func (c *Operator) createOrUpdateTLSAssetSecret(ctx context.Context, am *monitoringv1.Alertmanager, store *assets.Store) error {
	boolTrue := true
	sClient := c.kclient.CoreV1().Secrets(am.Namespace)

	tlsAssetsSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   tlsAssetsSecretName(am.Name),
			Labels: c.config.Labels.Merge(managedByOperatorLabels),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         am.APIVersion,
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               am.Kind,
					Name:               am.Name,
					UID:                am.UID,
				},
			},
		},
		Data: make(map[string][]byte, len(store.TLSAssets)),
	}

	for key, asset := range store.TLSAssets {
		tlsAssetsSecret.Data[key.String()] = []byte(asset)
	}

	err := k8sutil.CreateOrUpdateSecret(ctx, sClient, tlsAssetsSecret)
	if err != nil {
		return errors.Wrap(err, "failed to create TLS assets secret for Alertmanager")
	}

	return nil
}

//checkAlertmanagerSpecDeprecation checks for deprecated fields in the prometheus spec and logs a warning if applicable
func checkAlertmanagerSpecDeprecation(key string, a *monitoringv1.Alertmanager, logger log.Logger) {
	deprecationWarningf := "alertmanager key=%v, field %v is deprecated, '%v' field should be used instead"
	if a.Spec.BaseImage != "" {
		level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, key, "spec.baseImage", "spec.image"))
	}
	if a.Spec.Tag != "" {
		level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, key, "spec.tag", "spec.image"))
	}
	if a.Spec.SHA != "" {
		level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, key, "spec.sha", "spec.image"))
	}
}

func ListOptions(name string) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
			"app.kubernetes.io/name": "alertmanager",
			"alertmanager":           name,
		})).String(),
	}
}

func Status(ctx context.Context, kclient kubernetes.Interface, a *monitoringv1.Alertmanager) (*monitoringv1.AlertmanagerStatus, []v1.Pod, error) {
	res := &monitoringv1.AlertmanagerStatus{Paused: a.Spec.Paused}

	pods, err := kclient.CoreV1().Pods(a.Namespace).List(ctx, ListOptions(a.Name))
	if err != nil {
		return nil, nil, errors.Wrap(err, "retrieving pods of failed")
	}
	sset, err := kclient.AppsV1().StatefulSets(a.Namespace).Get(ctx, statefulSetNameFromAlertmanagerName(a.Name), metav1.GetOptions{})
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

func tlsAssetsSecretName(name string) string {
	return fmt.Sprintf("%s-tls-assets", prefixedName(name))
}
