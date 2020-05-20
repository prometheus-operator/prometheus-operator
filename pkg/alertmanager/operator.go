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
	"strings"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/listwatch"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"gopkg.in/yaml.v2"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/alertmanager/config"
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
		},
	}

	o.alrtInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			o.config.Namespaces.AlertmanagerAllowList,
			o.config.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = o.config.AlertManagerSelector
			},
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.AlertmanagerName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating alertmanager informers")
	}

	o.alrtCfgInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			o.config.Namespaces.AllowList,
			o.config.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = o.config.AlertManagerSelector
			},
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.AlertmanagerConfigName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating alertmanagerconfig informers")
	}

	o.ssetInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			o.config.Namespaces.AlertmanagerAllowList,
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
				listwatch.NewUnprivilegedNamespaceListWatchFromClient(o.logger, o.kclient.CoreV1().RESTClient(), allowList, o.config.Namespaces.DenyList, fields.Everything()),
			),
			&v1.Namespace{}, nsResyncPeriod, cache.Indexers{},
		)

		return nsInf
	}
	o.nsAlrtCfgInf = newNamespaceInformer(o, o.config.Namespaces.AllowList)
	if listwatch.IdenticalNamespaces(o.config.Namespaces.AllowList, o.config.Namespaces.AlertmanagerAllowList) {
		o.nsAlrtInf = o.nsAlrtCfgInf
	} else {
		o.nsAlrtInf = newNamespaceInformer(o, o.config.Namespaces.AlertmanagerAllowList)
	}

	return o, nil
}

// waitForCacheSync waits for the informers' caches to be synced.
func (c *Operator) waitForCacheSync(ctx context.Context) error {
	ok := true

	for _, infs := range []struct {
		name                 string
		informersForResource *informers.ForResource
	}{
		{"Alertmanager", c.alrtInfs},
		{"AlertmanagerConfig", c.alrtCfgInfs},
		{"StatefulSet", c.ssetInfs},
	} {
		for _, inf := range infs.informersForResource.GetInformers() {
			if !operator.WaitForCacheSync(ctx, log.With(c.logger, "informer", infs.name), inf.Informer()) {
				ok = false
			}
		}
	}

	if !ok {
		return errors.New("failed to sync caches")
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
	c.ssetInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleStatefulSetAdd,
		DeleteFunc: c.handleStatefulSetDelete,
		UpdateFunc: c.handleStatefulSetUpdate,
	})
}

func (c *Operator) handleAlertmanagerConfigAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "AlertmanagerConfig added")
		c.metrics.TriggerByCounter(monitoringv1.AlertmanagerConfigKind, "add").Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleAlertmanagerConfigUpdate(old, cur interface{}) {
	if old.(*monitoringv1.AlertmanagerConfig).ResourceVersion == cur.(*monitoringv1.AlertmanagerConfig).ResourceVersion {
		return
	}

	o, ok := c.getObject(cur)
	if ok {
		level.Debug(c.logger).Log("msg", "AlertmanagerConfig updated")
		c.metrics.TriggerByCounter(monitoringv1.AlertmanagerConfigKind, "update").Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleAlertmanagerConfigDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "AlertmanagerConfig delete")
		c.metrics.TriggerByCounter(monitoringv1.AlertmanagerConfigKind, "delete").Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

func (c *Operator) enqueueForMonitorNamespace(nsName string) {
	c.enqueueForNamespace(c.nsAlrtCfgInf.GetStore(), nsName)
}

// enqueueForNamespace enqueues all Alertmanager object keys that belong to the
// given namespace or select objects in the given namespace.
func (c *Operator) enqueueForNamespace(store cache.Store, nsName string) {
	nsObject, exists, err := store.GetByKey(nsName)
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

	objs, err := c.alrtInfs.List(labels.Everything())
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "listing all Alertmanager instances from cache failed",
			"err", err,
		)
		return
	}

	for _, obj := range objs {
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

	aKey := statefulSetKeyToAlertmanagerKey(key)
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

func alertmanagerNameFromStatefulSetName(name string) string {
	return strings.TrimPrefix(name, "alertmanager-")
}

func statefulSetNameFromAlertmanagerName(name string) string {
	return "alertmanager-" + name
}

func statefulSetKeyToAlertmanagerKey(key string) string {
	keyParts := strings.Split(key, "/")
	return keyParts[0] + "/" + strings.TrimPrefix(keyParts[1], "alertmanager-")
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
		c.enqueue(a)
	}
}

func (c *Operator) handleStatefulSetAdd(obj interface{}) {
	if a := c.alertmanagerForStatefulSet(obj); a != nil {
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
		c.enqueue(a)
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

	level.Info(c.logger).Log("msg", "sync alertmanager", "key", key)

	if err := c.provisionAlertmanagerConfiguration(context.TODO(), am); err != nil {
		return errors.Wrap(err, "provision alertmanager configuration")
	}

	// Create governing service if it doesn't exist.
	svcClient := c.kclient.CoreV1().Services(am.Namespace)
	if err = k8sutil.CreateOrUpdateService(ctx, svcClient, makeStatefulSetService(am, c.config)); err != nil {
		return errors.Wrap(err, "synchronizing governing service failed")
	}

	ssetClient := c.kclient.AppsV1().StatefulSets(am.Namespace)
	// Ensure we have a StatefulSet running Alertmanager deployed.
	obj, err := c.ssetInfs.Get(alertmanagerKeyToStatefulSetKey(key))

	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "retrieving statefulset failed")
	}

	if apierrors.IsNotFound(err) {
		sset, err := makeStatefulSet(am, nil, c.config)
		if err != nil {
			return errors.Wrap(err, "making the statefulset, to create, failed")
		}
		operator.SanitizeSTS(sset)
		if _, err := ssetClient.Create(ctx, sset, metav1.CreateOptions{}); err != nil {
			return errors.Wrap(err, "creating statefulset failed")
		}
		return nil
	}

	sset, err := makeStatefulSet(am, obj.(*appsv1.StatefulSet), c.config)
	if err != nil {
		return errors.Wrap(err, "making the statefulset, to update, failed")
	}

	operator.SanitizeSTS(sset)
	_, err = ssetClient.Update(ctx, sset, metav1.UpdateOptions{})
	sErr, ok := err.(*apierrors.StatusError)

	if ok && sErr.ErrStatus.Code == 422 && sErr.ErrStatus.Reason == metav1.StatusReasonInvalid {
		c.metrics.StsDeleteCreateCounter().Inc()
		level.Info(c.logger).Log("msg", "resolving illegal update of Alertmanager StatefulSet", "details", sErr.ErrStatus.Details)
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

func (c *Operator) provisionAlertmanagerConfiguration(ctx context.Context, am *monitoringv1.Alertmanager) error {
	secretName := defaultConfigSecretName(am.Name)
	if am.Spec.ConfigSecret != "" {
		secretName = am.Spec.ConfigSecret
	}

	secret, err := c.kclient.CoreV1().Secrets(am.Namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "get base configuration secret")
	}

	var baseConfigYaml yaml.MapSlice
	baseConfig, _ := secret.Data["alertmanager.yaml"]
	_, err = config.Load(string(baseConfig))
	if err != nil {
		level.Warn(c.logger).Log("msg", "base configuration loaded from Secret could not be parsed", "secret", secretName)
		baseConfig = []byte{}
	}

	err = yaml.Unmarshal(baseConfig, &baseConfigYaml)
	if err != nil {
		return errors.Wrap(err, "unmarshalling base config")
	}

	foundRoute := false
	foundReceivers := false
	foundInhibitRules := false
	for _, i := range baseConfigYaml {
		if k, ok := i.Key.(string); ok {
			if k == "route" {
				foundRoute = true
			}
			if k == "receivers" {
				foundReceivers = true
			}
			if k == "route" {
				foundInhibitRules = true
			}
		}
	}

	if !foundRoute {
		return errors.New("default route must be configured")
	}

	// build routes, receivers and inhibit rules from AlertmanagerConfigs here

	if foundReceivers {
		// just append receivrs from AlertmanagerConfig CRs, otherwise must first create receivers field
	}

	if foundInhibitRules {
		// just append inhibition rules from AlertmanagerConfig CRs, otherwise must first create inhibit rules field
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
			"app":          "alertmanager",
			"alertmanager": name,
		})).String(),
	}
}

func AlertmanagerStatus(ctx context.Context, kclient kubernetes.Interface, a *monitoringv1.Alertmanager) (*monitoringv1.AlertmanagerStatus, []v1.Pod, error) {
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
