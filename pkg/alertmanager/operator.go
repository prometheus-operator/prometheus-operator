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
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/coreos/prometheus-operator/pkg/analytics"
	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	prometheusoperator "github.com/coreos/prometheus-operator/pkg/prometheus"

	"github.com/coreos/prometheus-operator/third_party/workqueue"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	extensionsobj "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
)

const (
	tprAlertmanager = "alertmanager." + v1alpha1.TPRGroup

	resyncPeriod = 5 * time.Minute
)

// Operator manages lify cycle of Alertmanager deployments and
// monitoring configurations.
type Operator struct {
	kclient *kubernetes.Clientset
	mclient *v1alpha1.MonitoringV1alpha1Client
	logger  log.Logger

	alrtInf cache.SharedIndexInformer
	ssetInf cache.SharedIndexInformer

	queue workqueue.RateLimitingInterface

	config Config
}

type Config struct {
	Host                         string
	ConfigReloaderImage          string
	AlertmanagerDefaultBaseImage string
}

// New creates a new controller.
func New(c prometheusoperator.Config, logger log.Logger) (*Operator, error) {
	cfg, err := k8sutil.NewClusterConfig(c.Host, c.TLSInsecure, &c.TLSConfig)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating cluster config failed")
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating kubernetes client failed")
	}

	mclient, err := v1alpha1.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating monitoring client failed")
	}

	o := &Operator{
		kclient: client,
		mclient: mclient,
		logger:  logger,
		queue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "alertmanager"),
		config:  Config{Host: c.Host, ConfigReloaderImage: c.ConfigReloaderImage, AlertmanagerDefaultBaseImage: c.AlertmanagerDefaultBaseImage},
	}

	o.alrtInf = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  o.mclient.Alertmanagers(api.NamespaceAll).List,
			WatchFunc: o.mclient.Alertmanagers(api.NamespaceAll).Watch,
		},
		&v1alpha1.Alertmanager{}, resyncPeriod, cache.Indexers{},
	)
	o.ssetInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(o.kclient.Apps().RESTClient(), "statefulsets", api.NamespaceAll, nil),
		&v1beta1.StatefulSet{}, resyncPeriod, cache.Indexers{},
	)

	o.alrtInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.handleAlertmanagerAdd,
		DeleteFunc: o.handleAlertmanagerDelete,
		UpdateFunc: o.handleAlertmanagerUpdate,
	})
	o.ssetInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.handleStatefulSetAdd,
		DeleteFunc: o.handleStatefulSetDelete,
		UpdateFunc: o.handleStatefulSetUpdate,
	})

	return o, nil
}

func (c *Operator) RegisterMetrics(r prometheus.Registerer) {
	r.MustRegister(NewAlertmanagerCollector(c.alrtInf.GetStore()))
}

// Run the controller.
func (c *Operator) Run(stopc <-chan struct{}) error {
	defer c.queue.ShutDown()

	errChan := make(chan error)
	go func() {
		v, err := c.kclient.Discovery().ServerVersion()
		if err != nil {
			errChan <- errors.Wrap(err, "communicating with server failed")
			return
		}
		c.logger.Log("msg", "connection established", "cluster-version", v)

		if err := c.createTPRs(); err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
		c.logger.Log("msg", "TPR API endpoints ready")
	case <-stopc:
		return nil
	}

	go c.worker()

	go c.alrtInf.Run(stopc)
	go c.ssetInf.Run(stopc)

	<-stopc
	return nil
}

func (c *Operator) keyFunc(obj interface{}) (string, bool) {
	k, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		c.logger.Log("msg", "creating key failed", "err", err)
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
		c.logger.Log("msg", "get object failed", "err", err)
		return nil, false
	}
	return o, true
}

// enqueue adds a key to the queue. If obj is a key already it gets added directly.
// Otherwise, the key is extracted via keyFunc.
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

// enqueueForNamespace enqueues all Alertmanager object keys that belong to the given namespace.
func (c *Operator) enqueueForNamespace(ns string) {
	cache.ListAll(c.alrtInf.GetStore(), labels.Everything(), func(obj interface{}) {
		am := obj.(*v1alpha1.Alertmanager)
		if am.Namespace == ns {
			c.enqueue(am)
		}
	})
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (c *Operator) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *Operator) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.sync(key.(string))
	if err == nil {
		c.queue.Forget(key)
		return true
	}

	utilruntime.HandleError(errors.Wrap(err, fmt.Sprintf("Sync %q failed", key)))
	c.queue.AddRateLimited(key)

	return true
}

func (c *Operator) alertmanagerForStatefulSet(sset interface{}) *v1alpha1.Alertmanager {
	key, ok := c.keyFunc(sset)
	if !ok {
		return nil
	}

	aKey := statefulSetKeyToAlertmanagerKey(key)
	a, exists, err := c.alrtInf.GetStore().GetByKey(aKey)
	if err != nil {
		c.logger.Log("msg", "Alertmanager lookup failed", "err", err)
		return nil
	}
	if !exists {
		return nil
	}
	return a.(*v1alpha1.Alertmanager)
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

	analytics.AlertmanagerCreated()
	c.logger.Log("msg", "Alertmanager added", "key", key)
	c.enqueue(key)
}

func (c *Operator) handleAlertmanagerDelete(obj interface{}) {
	key, ok := c.keyFunc(obj)
	if !ok {
		return
	}

	analytics.AlertmanagerDeleted()
	c.logger.Log("msg", "Alertmanager deleted", "key", key)
	c.enqueue(key)
}

func (c *Operator) handleAlertmanagerUpdate(old, cur interface{}) {
	key, ok := c.keyFunc(cur)
	if !ok {
		return
	}

	c.logger.Log("msg", "Alertmanager updated", "key", key)
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
	old := oldo.(*v1beta1.StatefulSet)
	cur := curo.(*v1beta1.StatefulSet)

	c.logger.Log("msg", "update handler", "old", old.ResourceVersion, "cur", cur.ResourceVersion)

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

func (c *Operator) sync(key string) error {
	obj, exists, err := c.alrtInf.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		// TODO(fabxc): we want to do server side deletion due to the variety of
		// resources we create.
		// Doing so just based on the deletion event is not reliable, so
		// we have to garbage collect the controller-created resources in some other way.
		//
		// Let's rely on the index key matching that of the created configmap and replica
		// set for now. This does not work if we delete Alertmanager resources as the
		// controller is not running – that could be solved via garbage collection later.
		return c.destroyAlertmanager(key)
	}

	am := obj.(*v1alpha1.Alertmanager)
	if am.Spec.Paused {
		return nil
	}

	c.logger.Log("msg", "sync alertmanager", "key", key)

	// Create governing service if it doesn't exist.
	svcClient := c.kclient.Core().Services(am.Namespace)
	if err = k8sutil.CreateOrUpdateService(svcClient, makeStatefulSetService(am)); err != nil {
		return errors.Wrap(err, "synchronizing governing service failed")
	}

	ssetClient := c.kclient.Apps().StatefulSets(am.Namespace)
	// Ensure we have a StatefulSet running Alertmanager deployed.
	obj, exists, err = c.ssetInf.GetIndexer().GetByKey(alertmanagerKeyToStatefulSetKey(key))
	if err != nil {
		return errors.Wrap(err, "retrieving statefulset failed")
	}

	if !exists {
		sset, err := makeStatefulSet(am, nil, c.config)
		if err != nil {
			return errors.Wrap(err, "making the statefulset, to create, failed")
		}
		if _, err := ssetClient.Create(sset); err != nil {
			return errors.Wrap(err, "creating statefulset failed")
		}
		return nil
	}

	sset, err := makeStatefulSet(am, obj.(*v1beta1.StatefulSet), c.config)
	if err != nil {
		return errors.Wrap(err, "making the statefulset, to update, failed")
	}
	if _, err := ssetClient.Update(sset); err != nil {
		return errors.Wrap(err, "updating statefulset failed")
	}

	return c.syncVersion(am)
}

func ListOptions(name string) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
			"app":          "alertmanager",
			"alertmanager": name,
		})).String(),
	}
}

// syncVersion ensures that all running pods for a Alertmanager have the required version.
// It kills pods with the wrong version one-after-one and lets the StatefulSet controller
// create new pods.
//
// TODO(fabxc): remove this once the StatefulSet controller learns how to do rolling updates.
func (c *Operator) syncVersion(a *v1alpha1.Alertmanager) error {
	status, oldPods, err := AlertmanagerStatus(c.kclient, a)
	if err != nil {
		return errors.Wrap(err, "retrieving Alertmanager status failed")
	}

	// If the StatefulSet is still busy scaling, don't interfere by killing pods.
	// We enqueue ourselves again to until the StatefulSet is ready.
	expectedReplicas := int32(1)
	if a.Spec.Replicas != nil {
		expectedReplicas = *a.Spec.Replicas
	}
	if status.Replicas != expectedReplicas {
		return fmt.Errorf("scaling in progress, %d expected replicas, %d found replicas", expectedReplicas, status.Replicas)
	}
	if status.Replicas == 0 {
		return nil
	}
	if len(oldPods) == 0 {
		return nil
	}
	if status.UnavailableReplicas > 0 {
		return fmt.Errorf("waiting for %d unavailable pods to become ready", status.UnavailableReplicas)
	}

	// TODO(fabxc): delete oldest pod first.
	if err := c.kclient.Core().Pods(a.Namespace).Delete(oldPods[0].Name, nil); err != nil {
		return err
	}
	// If there are further pods that need updating, we enqueue ourselves again.
	if len(oldPods) > 1 {
		return fmt.Errorf("%d out-of-date pods remaining", len(oldPods)-1)
	}
	return nil
}

func AlertmanagerStatus(kclient *kubernetes.Clientset, a *v1alpha1.Alertmanager) (*v1alpha1.AlertmanagerStatus, []v1.Pod, error) {
	res := &v1alpha1.AlertmanagerStatus{Paused: a.Spec.Paused}

	pods, err := kclient.Core().Pods(a.Namespace).List(ListOptions(a.Name))
	if err != nil {
		return nil, nil, errors.Wrap(err, "retrieving pods of failed")
	}
	sset, err := kclient.Apps().StatefulSets(a.Namespace).Get(statefulSetNameFromAlertmanagerName(a.Name), metav1.GetOptions{})
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
			// TODO(fabxc): detect other fields of the pod template that are mutable.
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

func (c *Operator) destroyAlertmanager(key string) error {
	ssetKey := alertmanagerKeyToStatefulSetKey(key)
	obj, exists, err := c.ssetInf.GetStore().GetByKey(ssetKey)
	if err != nil {
		return errors.Wrap(err, "retrieving statefulset from cache failed")
	}
	if !exists {
		return nil
	}
	sset := obj.(*v1beta1.StatefulSet)
	*sset.Spec.Replicas = 0

	// Update the replica count to 0 and wait for all pods to be deleted.
	ssetClient := c.kclient.Apps().StatefulSets(sset.Namespace)

	if _, err := ssetClient.Update(sset); err != nil {
		return errors.Wrap(err, "updating statefulset for scale-down failed")
	}

	podClient := c.kclient.Core().Pods(sset.Namespace)

	// TODO(fabxc): temporary solution until StatefulSet status provides necessary info to know
	// whether scale-down completed.
	for {
		pods, err := podClient.List(ListOptions(alertmanagerNameFromStatefulSetName(sset.Name)))
		if err != nil {
			return errors.Wrap(err, "retrieving pods of statefulset failed")
		}
		if len(pods.Items) == 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// StatefulSet scaled down, we can delete it.
	if err := ssetClient.Delete(sset.Name, nil); err != nil {
		return errors.Wrap(err, "deleting statefulset failed")
	}

	return nil
}

func (c *Operator) createTPRs() error {
	tprs := []*extensionsobj.ThirdPartyResource{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: tprAlertmanager,
			},
			Versions: []extensionsobj.APIVersion{
				{Name: v1alpha1.TPRVersion},
			},
			Description: "Managed Alertmanager cluster",
		},
	}
	tprClient := c.kclient.Extensions().ThirdPartyResources()

	for _, tpr := range tprs {
		if _, err := tprClient.Create(tpr); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
		c.logger.Log("msg", "TPR created", "tpr", tpr.Name)
	}

	// We have to wait for the TPRs to be ready. Otherwise the initial watch may fail.
	return k8sutil.WaitForTPRReady(c.kclient.CoreV1().RESTClient(), v1alpha1.TPRGroup, v1alpha1.TPRVersion, v1alpha1.TPRAlertmanagerName)
}
