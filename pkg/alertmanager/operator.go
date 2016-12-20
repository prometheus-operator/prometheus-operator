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
	"strings"
	"time"

	"github.com/coreos/prometheus-operator/pkg/analytics"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/coreos/prometheus-operator/pkg/prometheus"
	"github.com/coreos/prometheus-operator/pkg/queue"
	"github.com/coreos/prometheus-operator/pkg/spec"

	"github.com/go-kit/kit/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	apierrors "k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/meta"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	extensionsobj "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/labels"
	utilruntime "k8s.io/client-go/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	TPRGroup   = "monitoring.coreos.com"
	TPRVersion = "v1alpha1"

	TPRAlertmanagersKind = "alertmanagers"

	tprAlertmanager = "alertmanager." + TPRGroup
)

// Operator manages lify cycle of Alertmanager deployments and
// monitoring configurations.
type Operator struct {
	kclient *kubernetes.Clientset
	pclient *rest.RESTClient
	logger  log.Logger

	alrtInf cache.SharedIndexInformer
	ssetInf cache.SharedIndexInformer

	queue *queue.Queue

	host string
}

// New creates a new controller.
func New(c prometheus.Config, logger log.Logger) (*Operator, error) {
	cfg, err := k8sutil.NewClusterConfig(c.Host, c.TLSInsecure, &c.TLSConfig)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	promclient, err := prometheus.NewPrometheusRESTClient(*cfg)
	if err != nil {
		return nil, err
	}

	return &Operator{
		kclient: client,
		pclient: promclient,
		logger:  logger,
		queue:   queue.New(),
		host:    cfg.Host,
	}, nil
}

// Run the controller.
func (c *Operator) Run(stopc <-chan struct{}) error {
	defer c.queue.ShutDown()
	go c.worker()

	errChan := make(chan error)
	go func() {
		v, err := c.kclient.Discovery().ServerVersion()
		if err != nil {
			errChan <- fmt.Errorf("communicating with server failed: %s", err)
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

	c.alrtInf = cache.NewSharedIndexInformer(
		NewAlertmanagerListWatch(c.pclient),
		&spec.Alertmanager{}, resyncPeriod, cache.Indexers{},
	)
	c.ssetInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.Apps().RESTClient(), "statefulsets", api.NamespaceAll, nil),
		&v1beta1.StatefulSet{}, resyncPeriod, cache.Indexers{},
	)

	c.alrtInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleAlertmanagerAdd,
		DeleteFunc: c.handleAlertmanagerDelete,
		UpdateFunc: c.handleAlertmanagerUpdate,
	})
	c.ssetInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleStatefulSetAdd,
		DeleteFunc: c.handleStatefulSetDelete,
		UpdateFunc: c.handleStatefulSetUpdate,
	})

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

func (c *Operator) getObject(obj interface{}) (meta.Object, bool) {
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
		am := obj.(*spec.Alertmanager)
		if am.Namespace == ns {
			c.enqueue(am)
		}
	})
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (c *Operator) worker() {
	for {
		key, quit := c.queue.Get()
		if quit {
			return
		}
		if err := c.sync(key.(string)); err != nil {
			utilruntime.HandleError(fmt.Errorf("reconciliation failed, re-enqueueing: %s", err))
			// We only mark the item as done after waiting. In the meantime
			// other items can be processed but the same item won't be processed again.
			// This is a trivial form of rate-limiting that is sufficient for our throughput
			// and latency expectations.
			go func() {
				time.Sleep(3 * time.Second)
				c.queue.Done(key)
			}()
			continue
		}

		c.queue.Done(key)
	}
}

func (c *Operator) alertmanagerForStatefulSet(ps interface{}) *spec.Alertmanager {
	key, ok := c.keyFunc(ps)
	if !ok {
		return nil
	}
	// Namespace/Name are one-to-one so the key will find the respective Alertmanager resource.
	a, exists, err := c.alrtInf.GetStore().GetByKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("get Alertmanager resource: %s", err))
		return nil
	}
	if !exists {
		return nil
	}
	return a.(*spec.Alertmanager)
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

	am := obj.(*spec.Alertmanager)

	c.logger.Log("msg", "sync alertmanager", "key", key)

	// Create governing service if it doesn't exist.
	svcClient := c.kclient.Core().Services(am.Namespace)
	if _, err := svcClient.Create(makeStatefulSetService(am)); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create statefulset service: %s", err)
	}

	ssetClient := c.kclient.Apps().StatefulSets(am.Namespace)
	// Ensure we have a StatefulSet running Alertmanager deployed.
	obj, exists, err = c.ssetInf.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}

	if !exists {
		if _, err := ssetClient.Create(makeStatefulSet(am, nil)); err != nil {
			return fmt.Errorf("create statefulset: %s", err)
		}
		return nil
	}
	if _, err := ssetClient.Update(makeStatefulSet(am, obj.(*v1beta1.StatefulSet))); err != nil {
		return err
	}

	return c.syncVersion(am)
}

func listOptions(name string) v1.ListOptions {
	return v1.ListOptions{
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
func (c *Operator) syncVersion(am *spec.Alertmanager) error {
	podClient := c.kclient.Core().Pods(am.Namespace)

	pods, err := podClient.List(listOptions(am.Name))
	if err != nil {
		return err
	}

	// If the StatefulSet is still busy scaling, don't interfere by killing pods.
	// We enqueue ourselves again to until the StatefulSet is ready.
	if len(pods.Items) != int(am.Spec.Replicas) {
		return fmt.Errorf("scaling in progress")
	}
	if len(pods.Items) == 0 {
		return nil
	}

	var oldPods []*v1.Pod
	allReady := true
	// Only proceed if all existing pods are running and ready.
	for _, pod := range pods.Items {
		ready, err := k8sutil.PodRunningAndReady(pod)
		if err != nil {
			c.logger.Log("msg", "cannot determine pod ready state", "err", err)
		}
		if ready {
			// TODO(fabxc): detect other fields of the pod template that are mutable.
			if !strings.HasSuffix(pod.Spec.Containers[0].Image, am.Spec.Version) {
				oldPods = append(oldPods, &pod)
			}
			continue
		}
		allReady = false
	}

	if len(oldPods) == 0 {
		return nil
	}
	if !allReady {
		return fmt.Errorf("waiting for pods to become ready")
	}

	// TODO(fabxc): delete oldest pod first.
	if err := podClient.Delete(oldPods[0].Name, nil); err != nil {
		return err
	}
	// If there are further pods that need updating, we enqueue ourselves again.
	if len(oldPods) > 1 {
		return fmt.Errorf("%d out-of-date pods remaining", len(oldPods)-1)
	}
	return nil
}

func (c *Operator) destroyAlertmanager(key string) error {
	obj, exists, err := c.ssetInf.GetStore().GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	sset := obj.(*v1beta1.StatefulSet)
	*sset.Spec.Replicas = 0

	// Update the replica count to 0 and wait for all pods to be deleted.
	ssetClient := c.kclient.Apps().StatefulSets(sset.Namespace)

	if _, err := ssetClient.Update(sset); err != nil {
		return err
	}

	podClient := c.kclient.Core().Pods(sset.Namespace)

	// TODO(fabxc): temporary solution until StatefulSet status provides necessary info to know
	// whether scale-down completed.
	for {
		pods, err := podClient.List(listOptions(sset.Name))
		if err != nil {
			return err
		}
		if len(pods.Items) == 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// StatefulSet scaled down, we can delete it.
	if err := ssetClient.Delete(sset.Name, nil); err != nil {
		return err
	}

	return nil
}

func (c *Operator) createTPRs() error {
	tprs := []*extensionsobj.ThirdPartyResource{
		{
			ObjectMeta: v1.ObjectMeta{
				Name: tprAlertmanager,
			},
			Versions: []extensionsobj.APIVersion{
				{Name: TPRVersion},
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
	return k8sutil.WaitForTPRReady(c.kclient.CoreV1().RESTClient(), TPRGroup, TPRVersion, TPRAlertmanagersKind)
}
