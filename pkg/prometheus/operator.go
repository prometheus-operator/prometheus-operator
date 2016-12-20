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

package prometheus

import (
	"fmt"
	"strings"
	"time"

	"github.com/coreos/prometheus-operator/pkg/analytics"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
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
	metav1 "k8s.io/client-go/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/labels"
	utilruntime "k8s.io/client-go/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	TPRGroup   = "monitoring.coreos.com"
	TPRVersion = "v1alpha1"

	TPRPrometheusesKind    = "prometheuses"
	TPRServiceMonitorsKind = "servicemonitors"

	tprServiceMonitor = "service-monitor." + TPRGroup
	tprPrometheus     = "prometheus." + TPRGroup
)

// Operator manages lify cycle of Prometheus deployments and
// monitoring configurations.
type Operator struct {
	kclient *kubernetes.Clientset
	pclient *rest.RESTClient
	logger  log.Logger

	promInf cache.SharedIndexInformer
	smonInf cache.SharedIndexInformer
	cmapInf cache.SharedIndexInformer
	ssetInf cache.SharedIndexInformer

	queue *queue.Queue

	host string
}

// Config defines configuration parameters for the Operator.
type Config struct {
	Host        string
	TLSInsecure bool
	TLSConfig   rest.TLSClientConfig
}

// New creates a new controller.
func New(conf Config, logger log.Logger) (*Operator, error) {
	cfg, err := k8sutil.NewClusterConfig(conf.Host, conf.TLSInsecure, &conf.TLSConfig)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	promclient, err := NewPrometheusRESTClient(*cfg)
	if err != nil {
		return nil, err
	}
	c := &Operator{
		kclient: client,
		pclient: promclient,
		logger:  logger,
		queue:   queue.New(),
		host:    cfg.Host,
	}

	c.promInf = cache.NewSharedIndexInformer(
		NewPrometheusListWatch(c.pclient),
		&spec.Prometheus{}, resyncPeriod, cache.Indexers{},
	)
	c.smonInf = cache.NewSharedIndexInformer(
		NewServiceMonitorListWatch(c.pclient),
		&spec.ServiceMonitor{}, resyncPeriod, cache.Indexers{},
	)
	c.cmapInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.Core().RESTClient(), "configmaps", api.NamespaceAll, nil),
		&v1.ConfigMap{}, resyncPeriod, cache.Indexers{},
	)
	c.ssetInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.Apps().RESTClient(), "statefulsets", api.NamespaceAll, nil),
		&v1beta1.StatefulSet{}, resyncPeriod, cache.Indexers{},
	)

	c.promInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleAddPrometheus,
		DeleteFunc: c.handleDeletePrometheus,
		UpdateFunc: c.handleUpdatePrometheus,
	})
	c.smonInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleSmonAdd,
		DeleteFunc: c.handleSmonDelete,
		UpdateFunc: c.handleSmonUpdate,
	})
	c.cmapInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.handleConfigmapDelete,
	})
	c.ssetInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleAddStatefulSet,
		DeleteFunc: c.handleDeleteStatefulSet,
		UpdateFunc: c.handleUpdateStatefulSet,
	})

	return c, nil
}

// Run the controller.
func (c *Operator) Run(stopc <-chan struct{}) error {
	defer c.queue.ShutDown()

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

	go c.worker()

	go c.promInf.Run(stopc)
	go c.smonInf.Run(stopc)
	go c.cmapInf.Run(stopc)
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

func (c *Operator) handleAddPrometheus(obj interface{}) {
	key, ok := c.keyFunc(obj)
	if !ok {
		return
	}

	analytics.PrometheusCreated()
	c.logger.Log("msg", "Prometheus added", "key", key)
	c.enqueue(key)
}

func (c *Operator) handleDeletePrometheus(obj interface{}) {
	key, ok := c.keyFunc(obj)
	if !ok {
		return
	}

	analytics.PrometheusDeleted()
	c.logger.Log("msg", "Prometheus deleted", "key", key)
	c.enqueue(key)
}

func (c *Operator) handleUpdatePrometheus(old, cur interface{}) {
	key, ok := c.keyFunc(cur)
	if !ok {
		return
	}

	c.logger.Log("msg", "Prometheus updated", "key", key)
	c.enqueue(key)
}

func (c *Operator) handleSmonAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleSmonUpdate(old, cur interface{}) {
	o, ok := c.getObject(cur)
	if ok {
		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleSmonDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleConfigmapDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if !ok {
		return
	}

	key, ok := c.keyFunc(o)
	if !ok {
		return
	}
	key = strings.TrimSuffix(key, "-rules")

	_, exists, err := c.promInf.GetIndexer().GetByKey(key)
	if err != nil {
		c.logger.Log("msg", "index lookup failed", "err", err)
	}
	if exists {
		c.enqueue(key)
	}
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

// enqueueForNamespace enqueues all Prometheus object keys that belong to the given namespace.
func (c *Operator) enqueueForNamespace(ns string) {
	cache.ListAll(c.promInf.GetStore(), labels.Everything(), func(obj interface{}) {
		p := obj.(*spec.Prometheus)
		if p.Namespace == ns {
			c.enqueue(p)
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

func (c *Operator) prometheusForStatefulSet(ps interface{}) *spec.Prometheus {
	key, ok := c.keyFunc(ps)
	if !ok {
		return nil
	}
	// Namespace/Name are one-to-one so the key will find the respective Prometheus resource.
	p, exists, err := c.promInf.GetStore().GetByKey(key)
	if err != nil {
		c.logger.Log("msg", "Prometheus lookup failed", "err", err)
		return nil
	}
	if !exists {
		return nil
	}
	return p.(*spec.Prometheus)
}

func (c *Operator) handleDeleteStatefulSet(obj interface{}) {
	if ps := c.prometheusForStatefulSet(obj); ps != nil {
		c.enqueue(ps)
	}
}

func (c *Operator) handleAddStatefulSet(obj interface{}) {
	if ps := c.prometheusForStatefulSet(obj); ps != nil {
		c.enqueue(ps)
	}
}

func (c *Operator) handleUpdateStatefulSet(oldo, curo interface{}) {
	old := oldo.(*v1beta1.StatefulSet)
	cur := curo.(*v1beta1.StatefulSet)

	c.logger.Log("msg", "update handler", "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	// Periodic resync may resend the deployment without changes in-between.
	// Also breaks loops created by updating the resource ourselves.
	if old.ResourceVersion == cur.ResourceVersion {
		return
	}

	if ps := c.prometheusForStatefulSet(cur); ps != nil {
		c.enqueue(ps)
	}
}

func (c *Operator) sync(key string) error {
	obj, exists, err := c.promInf.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		// TODO(fabxc): we want to do server side deletion due to the variety of
		// resources we create.
		// Doing so just based on the deletion event is not reliable, so
		// we have to garbage collect the controller-created resources in some other way.
		//
		// Let's rely on the index key matching that of the created configmap and StatefulSet for now.
		// This does not work if we delete Prometheus resources as the
		// controller is not running – that could be solved via garbage collection later.
		return c.destroyPrometheus(key)
	}

	p := obj.(*spec.Prometheus)
	if p.Spec.Paused {
		return nil
	}

	c.logger.Log("msg", "sync prometheus", "key", key)

	// If no service monitor selectors are configured, the user wants to manage
	// configuration himself.
	if p.Spec.ServiceMonitorSelector != nil {
		// We just always regenerate the configuration to be safe.
		if err := c.createConfig(p); err != nil {
			return err
		}
	}

	// Create ConfigMaps if they don't exist.
	cmClient := c.kclient.Core().ConfigMaps(p.Namespace)
	if _, err := cmClient.Create(makeEmptyConfig(p.Name)); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	if _, err := cmClient.Create(makeEmptyRules(p.Name)); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	// Create governing service if it doesn't exist.
	svcClient := c.kclient.Core().Services(p.Namespace)
	if _, err := svcClient.Create(makeStatefulSetService(p)); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create statefulset service: %s", err)
	}

	ssetClient := c.kclient.Apps().StatefulSets(p.Namespace)
	// Ensure we have a StatefulSet running Prometheus deployed.
	obj, exists, err = c.ssetInf.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}

	if !exists {
		if _, err := ssetClient.Create(makeStatefulSet(*p, nil)); err != nil {
			return fmt.Errorf("create statefulset: %s", err)
		}
		return nil
	}
	if _, err := ssetClient.Update(makeStatefulSet(*p, obj.(*v1beta1.StatefulSet))); err != nil {
		return err
	}

	return c.syncVersion(key, p)
}

func ListOptions(name string) v1.ListOptions {
	return v1.ListOptions{
		LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
			"app":        "prometheus",
			"prometheus": name,
		})).String(),
	}
}

// syncVersion ensures that all running pods for a Prometheus have the required version.
// It kills pods with the wrong version one-after-one and lets the StatefulSet controller
// create new pods.
//
// TODO(fabxc): remove this once the StatefulSet controller learns how to do rolling updates.
func (c *Operator) syncVersion(key string, p *spec.Prometheus) error {
	status, oldPods, err := PrometheusStatus(c.kclient, p)
	if err != nil {
		return err
	}

	// If the StatefulSet is still busy scaling, don't interfere by killing pods.
	// We enqueue ourselves again to until the StatefulSet is ready.
	if status.Replicas != p.Spec.Replicas {
		return fmt.Errorf("scaling in progress")
	}
	if status.Replicas == 0 {
		return nil
	}
	if len(oldPods) == 0 {
		return nil
	}
	if status.UnavailableReplicas > 0 {
		return fmt.Errorf("waiting for pods to become ready")
	}

	// TODO(fabxc): delete oldest pod first.
	if err := c.kclient.Core().Pods(p.Namespace).Delete(oldPods[0].Name, nil); err != nil {
		return err
	}
	// If there are further pods that need updating, we enqueue ourselves again.
	if len(oldPods) > 1 {
		return fmt.Errorf("%d out-of-date pods remaining", len(oldPods)-1)
	}
	return nil
}

func PrometheusStatus(kclient *kubernetes.Clientset, p *spec.Prometheus) (*spec.PrometheusStatus, []*v1.Pod, error) {
	res := &spec.PrometheusStatus{}

	pods, err := kclient.Core().Pods(p.Namespace).List(ListOptions(p.Name))
	if err != nil {
		return nil, nil, err
	}

	res.Replicas = int32(len(pods.Items))

	var oldPods []*v1.Pod
	for _, pod := range pods.Items {
		ready, err := k8sutil.PodRunningAndReady(pod)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot determine pod ready state: %s", err)
		}
		if ready {
			res.AvailableReplicas++
			// TODO(fabxc): detect other fields of the pod template that are mutable.
			if strings.HasSuffix(pod.Spec.Containers[0].Image, p.Spec.Version) {
				res.UpdatedReplicas++
			} else {
				oldPods = append(oldPods, &pod)
			}
			continue
		}
		res.UnavailableReplicas++
	}

	return res, oldPods, nil
}

func (c *Operator) destroyPrometheus(key string) error {
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

	// TODO(fabxc): temprorary solution until StatefulSet status provides necessary info to know
	// whether scale-down completed.
	for {
		pods, err := podClient.List(ListOptions(sset.Name))
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

	// Delete the auto-generate configuration.
	// TODO(fabxc): add an ownerRef at creation so we don't delete config maps
	// manually created for Prometheus servers with no ServiceMonitor selectors.
	cm := c.kclient.Core().ConfigMaps(sset.Namespace)

	if err := cm.Delete(sset.Name, nil); err != nil {
		return err
	}
	if err := cm.Delete(fmt.Sprintf("%s-rules", sset.Name), nil); err != nil {
		return err
	}
	return nil
}

func (c *Operator) createConfig(p *spec.Prometheus) error {
	smons, err := c.selectServiceMonitors(p)
	if err != nil {
		return err
	}
	// Update config map based on the most recent configuration.
	b, err := generateConfig(p, smons)
	if err != nil {
		return fmt.Errorf("generating config failed: %s", err)
	}

	cm := &v1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name: p.Name,
		},
		Data: map[string]string{
			"prometheus.yaml": string(b),
		},
	}

	cmClient := c.kclient.CoreV1().ConfigMaps(p.Namespace)

	_, err = cmClient.Get(p.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = cmClient.Create(cm)
	} else if err == nil {
		_, err = cmClient.Update(cm)
	}
	return err
}

func (c *Operator) selectServiceMonitors(p *spec.Prometheus) (map[string]*spec.ServiceMonitor, error) {
	// Selectors might overlap. Deduplicate them along the keyFunc.
	res := make(map[string]*spec.ServiceMonitor)

	selector, err := metav1.LabelSelectorAsSelector(p.Spec.ServiceMonitorSelector)
	if err != nil {
		return nil, err
	}

	// Only service monitors within the same namespace as the Prometheus
	// object can belong to it.
	cache.ListAllByNamespace(c.smonInf.GetIndexer(), p.Namespace, selector, func(obj interface{}) {
		k, ok := c.keyFunc(obj)
		if ok {
			res[k] = obj.(*spec.ServiceMonitor)
		}
	})

	return res, nil
}

func (c *Operator) createTPRs() error {
	tprs := []*extensionsobj.ThirdPartyResource{
		{
			ObjectMeta: v1.ObjectMeta{
				Name: tprServiceMonitor,
			},
			Versions: []extensionsobj.APIVersion{
				{Name: TPRVersion},
			},
			Description: "Prometheus monitoring for a service",
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name: tprPrometheus,
			},
			Versions: []extensionsobj.APIVersion{
				{Name: TPRVersion},
			},
			Description: "Managed Prometheus server",
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
	err := k8sutil.WaitForTPRReady(c.kclient.CoreV1().RESTClient(), TPRGroup, TPRVersion, TPRPrometheusesKind)
	if err != nil {
		return err
	}
	return k8sutil.WaitForTPRReady(c.kclient.CoreV1().RESTClient(), TPRGroup, TPRVersion, TPRServiceMonitorsKind)
}
