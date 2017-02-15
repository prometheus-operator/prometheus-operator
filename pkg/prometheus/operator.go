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
	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"

	"github.com/go-kit/kit/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	extensionsobj "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/workqueue"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	tprServiceMonitor = "service-monitor." + v1alpha1.TPRGroup
	tprPrometheus     = "prometheus." + v1alpha1.TPRGroup

	resyncPeriod = 5 * time.Minute
)

// Operator manages lify cycle of Prometheus deployments and
// monitoring configurations.
type Operator struct {
	kclient *kubernetes.Clientset
	mclient *v1alpha1.MonitoringV1alpha1Client
	logger  log.Logger

	promInf cache.SharedIndexInformer
	smonInf cache.SharedIndexInformer
	cmapInf cache.SharedIndexInformer
	ssetInf cache.SharedIndexInformer
	nodeInf cache.SharedIndexInformer

	queue workqueue.RateLimitingInterface

	host                   string
	kubeletObjectName      string
	kubeletObjectNamespace string
	kubeletSyncEnabled     bool
	config                 Config
}

// Config defines configuration parameters for the Operator.
type Config struct {
	Host                string
	KubeletObject       string
	TLSInsecure         bool
	TLSConfig           rest.TLSClientConfig
	ConfigReloaderImage string
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

	mclient, err := v1alpha1.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	kubeletObjectName := ""
	kubeletObjectNamespace := ""
	kubeletSyncEnabled := false

	if conf.KubeletObject != "" {
		parts := strings.Split(conf.KubeletObject, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformatted kubelet object string, must be in format \"namespace/name\"")
		}
		kubeletObjectNamespace = parts[0]
		kubeletObjectName = parts[1]
		kubeletSyncEnabled = true
	}

	c := &Operator{
		kclient:                client,
		mclient:                mclient,
		logger:                 logger,
		queue:                  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "prometheus"),
		host:                   cfg.Host,
		kubeletObjectName:      kubeletObjectName,
		kubeletObjectNamespace: kubeletObjectNamespace,
		kubeletSyncEnabled:     kubeletSyncEnabled,
		config:                 conf,
	}

	c.promInf = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  mclient.Prometheuses(api.NamespaceAll).List,
			WatchFunc: mclient.Prometheuses(api.NamespaceAll).Watch,
		},
		&v1alpha1.Prometheus{}, resyncPeriod, cache.Indexers{},
	)
	c.promInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleAddPrometheus,
		DeleteFunc: c.handleDeletePrometheus,
		UpdateFunc: c.handleUpdatePrometheus,
	})

	c.smonInf = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  mclient.ServiceMonitors(api.NamespaceAll).List,
			WatchFunc: mclient.ServiceMonitors(api.NamespaceAll).Watch,
		},
		&v1alpha1.ServiceMonitor{}, resyncPeriod, cache.Indexers{},
	)
	c.smonInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleSmonAdd,
		DeleteFunc: c.handleSmonDelete,
		UpdateFunc: c.handleSmonUpdate,
	})

	c.cmapInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.Core().RESTClient(), "configmaps", api.NamespaceAll, nil),
		&v1.ConfigMap{}, resyncPeriod, cache.Indexers{},
	)
	c.cmapInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.handleConfigmapDelete,
	})

	c.ssetInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.Apps().RESTClient(), "statefulsets", api.NamespaceAll, nil),
		&v1beta1.StatefulSet{}, resyncPeriod, cache.Indexers{},
	)
	c.ssetInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleAddStatefulSet,
		DeleteFunc: c.handleDeleteStatefulSet,
		UpdateFunc: c.handleUpdateStatefulSet,
	})

	if kubeletSyncEnabled {
		c.nodeInf = cache.NewSharedIndexInformer(
			cache.NewListWatchFromClient(c.kclient.Core().RESTClient(), "nodes", api.NamespaceAll, nil),
			&v1.Node{}, resyncPeriod, cache.Indexers{},
		)
		c.nodeInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAddNode,
			DeleteFunc: c.handleDeleteNode,
			UpdateFunc: c.handleUpdateNode,
		})
	}

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

	if c.kubeletSyncEnabled {
		go c.nodeInf.Run(stopc)
	}

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

func (c *Operator) handleAddNode(obj interface{})         { c.syncNodeEndpoints() }
func (c *Operator) handleDeleteNode(obj interface{})      { c.syncNodeEndpoints() }
func (c *Operator) handleUpdateNode(old, cur interface{}) { c.syncNodeEndpoints() }

func (c *Operator) syncNodeEndpoints() {
	endpoints := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.kubeletObjectName,
			Labels: map[string]string{
				"k8s-app": "kubelet",
			},
		},
		Subsets: []v1.EndpointSubset{
			v1.EndpointSubset{
				Ports: []v1.EndpointPort{
					v1.EndpointPort{
						Name: "https-metrics",
						Port: 10250,
					},
				},
			},
		},
	}

	cache.ListAll(c.nodeInf.GetStore(), labels.Everything(), func(obj interface{}) {
		n := obj.(*v1.Node)
		for _, a := range n.Status.Addresses {
			if a.Type == v1.NodeInternalIP {
				endpoints.Subsets[0].Addresses = append(endpoints.Subsets[0].Addresses, v1.EndpointAddress{
					IP:       a.Address,
					Hostname: n.Name,
					NodeName: &n.Name,
					TargetRef: &v1.ObjectReference{
						Kind:       "Node",
						Name:       n.Name,
						UID:        n.UID,
						APIVersion: n.APIVersion,
					},
				})
			}
		}
	})

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.kubeletObjectName,
			Labels: map[string]string{
				"k8s-app": "kubelet",
			},
		},
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: "https-metrics",
					Port: 10250,
				},
			},
		},
	}

	_, err := c.kclient.CoreV1().Services(c.kubeletObjectNamespace).Update(svc)
	if err != nil && !apierrors.IsNotFound(err) {
		c.logger.Log("msg", "updating kubelet service object failed", "err", err)
	}
	if apierrors.IsNotFound(err) {
		_, err = c.kclient.CoreV1().Services(c.kubeletObjectNamespace).Create(svc)
		if err != nil {
			c.logger.Log("msg", "creating kubelet service object failed", "err", err)
		}
	}

	_, err = c.kclient.CoreV1().Endpoints(c.kubeletObjectNamespace).Update(endpoints)
	if err != nil && !apierrors.IsNotFound(err) {
		c.logger.Log("msg", "updating kubelet enpoints object failed", "err", err)
	}
	if apierrors.IsNotFound(err) {
		_, err = c.kclient.CoreV1().Endpoints(c.kubeletObjectNamespace).Create(endpoints)
		if err != nil {
			c.logger.Log("msg", "creating kubelet enpoints object failed", "err", err)
		}
	}
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

func (c *Operator) getObject(obj interface{}) (apimetav1.Object, bool) {
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
		p := obj.(*v1alpha1.Prometheus)
		if p.Namespace == ns {
			c.enqueue(p)
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

	utilruntime.HandleError(fmt.Errorf("Sync %q failed with %v", key, err))
	c.queue.AddRateLimited(key)

	return true
}

func (c *Operator) prometheusForStatefulSet(sset interface{}) *v1alpha1.Prometheus {
	key, ok := c.keyFunc(sset)
	if !ok {
		return nil
	}

	promKey := statefulSetKeyToPrometheusKey(key)
	p, exists, err := c.promInf.GetStore().GetByKey(promKey)
	if err != nil {
		c.logger.Log("msg", "Prometheus lookup failed", "err", err)
		return nil
	}
	if !exists {
		return nil
	}
	return p.(*v1alpha1.Prometheus)
}

func prometheusNameFromStatefulSetName(name string) string {
	return strings.TrimPrefix(name, "prometheus-")
}

func statefulSetKeyToPrometheusKey(key string) string {
	keyParts := strings.Split(key, "/")
	return keyParts[0] + "/" + strings.TrimPrefix(keyParts[1], "prometheus-")
}

func prometheusKeyToStatefulSetKey(key string) string {
	keyParts := strings.Split(key, "/")
	return keyParts[0] + "/prometheus-" + keyParts[1]
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

	p := obj.(*v1alpha1.Prometheus)
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
	obj, exists, err = c.ssetInf.GetIndexer().GetByKey(prometheusKeyToStatefulSetKey(key))
	if err != nil {
		return err
	}

	if !exists {
		if _, err := ssetClient.Create(makeStatefulSet(*p, nil, &c.config)); err != nil {
			return fmt.Errorf("create statefulset: %s", err)
		}
		return nil
	}
	if _, err := ssetClient.Update(makeStatefulSet(*p, obj.(*v1beta1.StatefulSet), &c.config)); err != nil {
		return err
	}

	return c.syncVersion(key, p)
}

func ListOptions(name string) metav1.ListOptions {
	return metav1.ListOptions{
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
func (c *Operator) syncVersion(key string, p *v1alpha1.Prometheus) error {
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

func PrometheusStatus(kclient *kubernetes.Clientset, p *v1alpha1.Prometheus) (*v1alpha1.PrometheusStatus, []v1.Pod, error) {
	res := &v1alpha1.PrometheusStatus{Paused: p.Spec.Paused}

	pods, err := kclient.Core().Pods(p.Namespace).List(ListOptions(p.Name))
	if err != nil {
		return nil, nil, err
	}

	res.Replicas = int32(len(pods.Items))

	var oldPods []v1.Pod
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
				oldPods = append(oldPods, pod)
			}
			continue
		}
		res.UnavailableReplicas++
	}

	return res, oldPods, nil
}

func (c *Operator) destroyPrometheus(key string) error {
	ssetKey := prometheusKeyToStatefulSetKey(key)
	obj, exists, err := c.ssetInf.GetStore().GetByKey(ssetKey)
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
		pods, err := podClient.List(ListOptions(prometheusNameFromStatefulSetName(sset.Name)))
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

func (c *Operator) createConfig(p *v1alpha1.Prometheus) error {
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
		ObjectMeta: apimetav1.ObjectMeta{
			Name: configConfigMapName(p.Name),
		},
		Data: map[string]string{
			"prometheus.yaml": string(b),
		},
	}

	cmClient := c.kclient.CoreV1().ConfigMaps(p.Namespace)

	_, err = cmClient.Get(cm.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = cmClient.Create(cm)
	} else if err == nil {
		_, err = cmClient.Update(cm)
	}
	return err
}

func (c *Operator) selectServiceMonitors(p *v1alpha1.Prometheus) (map[string]*v1alpha1.ServiceMonitor, error) {
	// Selectors might overlap. Deduplicate them along the keyFunc.
	res := make(map[string]*v1alpha1.ServiceMonitor)

	selector, err := metav1.LabelSelectorAsSelector(p.Spec.ServiceMonitorSelector)
	if err != nil {
		return nil, err
	}

	// Only service monitors within the same namespace as the Prometheus
	// object can belong to it.
	cache.ListAllByNamespace(c.smonInf.GetIndexer(), p.Namespace, selector, func(obj interface{}) {
		k, ok := c.keyFunc(obj)
		if ok {
			res[k] = obj.(*v1alpha1.ServiceMonitor)
		}
	})

	return res, nil
}

func (c *Operator) createTPRs() error {
	tprs := []*extensionsobj.ThirdPartyResource{
		{
			ObjectMeta: apimetav1.ObjectMeta{
				Name: tprServiceMonitor,
			},
			Versions: []extensionsobj.APIVersion{
				{Name: v1alpha1.TPRVersion},
			},
			Description: "Prometheus monitoring for a service",
		},
		{
			ObjectMeta: apimetav1.ObjectMeta{
				Name: tprPrometheus,
			},
			Versions: []extensionsobj.APIVersion{
				{Name: v1alpha1.TPRVersion},
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
	err := k8sutil.WaitForTPRReady(c.kclient.CoreV1().RESTClient(), v1alpha1.TPRGroup, v1alpha1.TPRVersion, v1alpha1.TPRPrometheusName)
	if err != nil {
		return err
	}
	return k8sutil.WaitForTPRReady(c.kclient.CoreV1().RESTClient(), v1alpha1.TPRGroup, v1alpha1.TPRVersion, v1alpha1.TPRServiceMonitorName)
}
