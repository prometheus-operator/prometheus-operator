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
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/coreos/prometheus-operator/pkg/analytics"
	"github.com/coreos/prometheus-operator/pkg/client/monitoring"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
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
	extensionsobjold "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	configFilename = "prometheus.yaml"

	resyncPeriod = 5 * time.Minute
)

// Operator manages lify cycle of Prometheus deployments and
// monitoring configurations.
type Operator struct {
	kclient   kubernetes.Interface
	mclient   monitoring.Interface
	crdclient apiextensionsclient.Interface
	logger    log.Logger

	promInf cache.SharedIndexInformer
	smonInf cache.SharedIndexInformer
	cmapInf cache.SharedIndexInformer
	secrInf cache.SharedIndexInformer
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
	Host                         string
	KubeletObject                string
	TLSInsecure                  bool
	StatefulSetUpdatesAvailable  bool
	TLSConfig                    rest.TLSClientConfig
	ConfigReloaderImage          string
	PrometheusConfigReloader     string
	AlertmanagerDefaultBaseImage string
	PrometheusDefaultBaseImage   string
}

type BasicAuthCredentials struct {
	username string
	password string
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

	mclient, err := monitoring.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	crdclient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating apiextensions client failed")
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
		crdclient:              crdclient,
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
			ListFunc:  mclient.MonitoringV1().Prometheuses(api.NamespaceAll).List,
			WatchFunc: mclient.MonitoringV1().Prometheuses(api.NamespaceAll).Watch,
		},
		&monitoringv1.Prometheus{}, resyncPeriod, cache.Indexers{},
	)
	c.promInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleAddPrometheus,
		DeleteFunc: c.handleDeletePrometheus,
		UpdateFunc: c.handleUpdatePrometheus,
	})

	c.smonInf = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  mclient.MonitoringV1().ServiceMonitors(api.NamespaceAll).List,
			WatchFunc: mclient.MonitoringV1().ServiceMonitors(api.NamespaceAll).Watch,
		},
		&monitoringv1.ServiceMonitor{}, resyncPeriod, cache.Indexers{},
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
		AddFunc:    c.handleConfigMapAdd,
		DeleteFunc: c.handleConfigMapDelete,
		UpdateFunc: c.handleConfigMapUpdate,
	})
	c.secrInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.Core().RESTClient(), "secrets", api.NamespaceAll, nil),
		&v1.Secret{}, resyncPeriod, cache.Indexers{},
	)
	c.secrInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleSecretAdd,
		DeleteFunc: c.handleSecretDelete,
		UpdateFunc: c.handleSecretUpdate,
	})

	c.ssetInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.AppsV1beta1().RESTClient(), "statefulsets", api.NamespaceAll, nil),
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

func (c *Operator) RegisterMetrics(r prometheus.Registerer) {
	r.MustRegister(NewPrometheusCollector(c.promInf.GetStore()))
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

		mv, err := k8sutil.GetMinorVersion(c.kclient.Discovery())
		if mv < 7 {
			c.config.StatefulSetUpdatesAvailable = false
		}

		c.config.StatefulSetUpdatesAvailable = true
		if err := c.createCRDs(); err != nil {
			errChan <- errors.Wrap(err, "creating CRDs failed")
			return
		}
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
		c.logger.Log("msg", "CRD API endpoints ready")
	case <-stopc:
		return nil
	}

	go c.worker()

	go c.promInf.Run(stopc)
	go c.smonInf.Run(stopc)
	go c.cmapInf.Run(stopc)
	go c.secrInf.Run(stopc)
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

// nodeAddresses returns the provided node's address, based on the priority:
// 1. NodeInternalIP
// 2. NodeExternalIP
// 3. NodeLegacyHostIP
// 3. NodeHostName
//
// Copied from github.com/prometheus/prometheus/discovery/kubernetes/node.go
func nodeAddress(node *v1.Node) (string, map[v1.NodeAddressType][]string, error) {
	m := map[v1.NodeAddressType][]string{}
	for _, a := range node.Status.Addresses {
		m[a.Type] = append(m[a.Type], a.Address)
	}

	if addresses, ok := m[v1.NodeInternalIP]; ok {
		return addresses[0], m, nil
	}
	if addresses, ok := m[v1.NodeExternalIP]; ok {
		return addresses[0], m, nil
	}
	// NodeLegacyHostIP support has been removed in 1.7, this is here for prolonged 1.6 support.
	if addresses, ok := m[v1.NodeAddressType("LegacyHostIP")]; ok {
		return addresses[0], m, nil
	}
	if addresses, ok := m[v1.NodeHostName]; ok {
		return addresses[0], m, nil
	}
	return "", m, fmt.Errorf("host address unknown")
}

func (c *Operator) syncNodeEndpoints() {
	eps := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.kubeletObjectName,
			Labels: map[string]string{
				"k8s-app": "kubelet",
			},
		},
		Subsets: []v1.EndpointSubset{
			{
				Ports: []v1.EndpointPort{
					{
						Name: "https-metrics",
						Port: 10250,
					},
					{
						Name: "http-metrics",
						Port: 10255,
					},
					{
						Name: "cadvisor",
						Port: 4194,
					},
				},
			},
		},
	}

	cache.ListAll(c.nodeInf.GetStore(), labels.Everything(), func(obj interface{}) {
		n := obj.(*v1.Node)
		address, _, err := nodeAddress(n)
		if err != nil {
			c.logger.Log("msg", "failed to determine hostname for node", "err", err, "node", n.Name)
			return
		}
		eps.Subsets[0].Addresses = append(eps.Subsets[0].Addresses, v1.EndpointAddress{
			IP:       address,
			NodeName: &n.Name,
			TargetRef: &v1.ObjectReference{
				Kind:       "Node",
				Name:       n.Name,
				UID:        n.UID,
				APIVersion: n.APIVersion,
			},
		})
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
				{
					Name: "https-metrics",
					Port: 10250,
				},
			},
		},
	}

	err := k8sutil.CreateOrUpdateService(c.kclient.CoreV1().Services(c.kubeletObjectNamespace), svc)
	if err != nil {
		c.logger.Log("msg", "synchronizing kubelet service object failed", "err", err, "namespace", c.kubeletObjectNamespace, "name", svc.Name)
	}

	err = k8sutil.CreateOrUpdateEndpoints(c.kclient.CoreV1().Endpoints(c.kubeletObjectNamespace), eps)
	if err != nil {
		c.logger.Log("msg", "synchronizing kubelet endpoints object failed", "err", err, "namespace", c.kubeletObjectNamespace, "name", eps.Name)
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

func (c *Operator) handleSecretDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleSecretUpdate(old, cur interface{}) {
	o, ok := c.getObject(cur)
	if ok {
		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleSecretAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleConfigMapAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleConfigMapDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleConfigMapUpdate(old, cur interface{}) {
	o, ok := c.getObject(cur)
	if ok {
		c.enqueueForNamespace(o.GetNamespace())
	}
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

// enqueueForNamespace enqueues all Prometheus object keys that belong to the given namespace.
func (c *Operator) enqueueForNamespace(ns string) {
	cache.ListAll(c.promInf.GetStore(), labels.Everything(), func(obj interface{}) {
		p := obj.(*monitoringv1.Prometheus)
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

	utilruntime.HandleError(errors.Wrap(err, fmt.Sprintf("Sync %q failed", key)))
	c.queue.AddRateLimited(key)

	return true
}

func (c *Operator) prometheusForStatefulSet(sset interface{}) *monitoringv1.Prometheus {
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
	return p.(*monitoringv1.Prometheus)
}

func prometheusNameFromStatefulSetName(name string) string {
	return strings.TrimPrefix(name, "prometheus-")
}

func statefulSetNameFromPrometheusName(name string) string {
	return "prometheus-" + name
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

	p := obj.(*monitoringv1.Prometheus)
	if p.Spec.Paused {
		return nil
	}

	c.logger.Log("msg", "sync prometheus", "key", key)

	ruleFileConfigMaps, err := c.ruleFileConfigMaps(p)
	if err != nil {
		return errors.Wrap(err, "retrieving rule file configmaps failed")
	}

	// If no service monitor selectors are configured, the user wants to manage
	// configuration himself.
	if p.Spec.ServiceMonitorSelector != nil {
		// We just always regenerate the configuration to be safe.
		if err := c.createConfig(p, ruleFileConfigMaps); err != nil {
			return errors.Wrap(err, "creating config failed")
		}
	}

	// Create Secret if it doesn't exist.
	s, err := makeEmptyConfig(p.Name, ruleFileConfigMaps)
	if err != nil {
		return errors.Wrap(err, "generating empty config secret failed")
	}
	if _, err := c.kclient.Core().Secrets(p.Namespace).Create(s); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "creating empty config file failed")
	}

	// Create governing service if it doesn't exist.
	svcClient := c.kclient.Core().Services(p.Namespace)
	if err := k8sutil.CreateOrUpdateService(svcClient, makeStatefulSetService(p)); err != nil {
		return errors.Wrap(err, "synchronizing governing service failed")
	}

	ssetClient := c.kclient.AppsV1beta1().StatefulSets(p.Namespace)
	// Ensure we have a StatefulSet running Prometheus deployed.
	obj, exists, err = c.ssetInf.GetIndexer().GetByKey(prometheusKeyToStatefulSetKey(key))
	if err != nil {
		return errors.Wrap(err, "retrieving statefulset failed")
	}

	if !exists {
		sset, err := makeStatefulSet(*p, nil, &c.config, ruleFileConfigMaps)
		if err != nil {
			return errors.Wrap(err, "creating statefulset failed")
		}
		if _, err := ssetClient.Create(sset); err != nil {
			return errors.Wrap(err, "creating statefulset failed")
		}
		return nil
	}
	sset, err := makeStatefulSet(*p, obj.(*v1beta1.StatefulSet), &c.config, ruleFileConfigMaps)
	if err != nil {
		return errors.Wrap(err, "updating statefulset failed")
	}
	if _, err := ssetClient.Update(sset); err != nil {
		return errors.Wrap(err, "updating statefulset failed")
	}

	err = c.syncVersion(key, p)
	if err != nil {
		return errors.Wrap(err, "syncing version failed")
	}

	return nil
}

func (c *Operator) ruleFileConfigMaps(p *monitoringv1.Prometheus) ([]*v1.ConfigMap, error) {
	res := []*v1.ConfigMap{}

	ruleSelector, err := metav1.LabelSelectorAsSelector(p.Spec.RuleSelector)
	if err != nil {
		return nil, err
	}

	cache.ListAllByNamespace(c.cmapInf.GetIndexer(), p.Namespace, ruleSelector, func(obj interface{}) {
		_, ok := c.keyFunc(obj)
		if ok {
			res = append(res, obj.(*v1.ConfigMap))
		}
	})

	return res, nil
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
// TODO(brancz): remove this once the 1.6 support is removed.
func (c *Operator) syncVersion(key string, p *monitoringv1.Prometheus) error {
	if c.config.StatefulSetUpdatesAvailable {
		return nil
	}

	status, oldPods, err := PrometheusStatus(c.kclient, p)
	if err != nil {
		return errors.Wrap(err, "retrieving Prometheus status failed")
	}

	// If the StatefulSet is still busy scaling, don't interfere by killing pods.
	// We enqueue ourselves again to until the StatefulSet is ready.
	expectedReplicas := int32(1)
	if p.Spec.Replicas != nil {
		expectedReplicas = *p.Spec.Replicas
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
	if err := c.kclient.Core().Pods(p.Namespace).Delete(oldPods[0].Name, nil); err != nil {
		return err
	}

	// If there are further pods that need updating, we enqueue ourselves again.
	if len(oldPods) > 1 {
		return fmt.Errorf("%d out-of-date pods remaining", len(oldPods)-1)
	}
	return nil
}

// PrometheusStatus evaluates the current status of a Prometheus deployment with respect
// to its specified resource object. It return the status and a list of pods that
// are not updated.
func PrometheusStatus(kclient kubernetes.Interface, p *monitoringv1.Prometheus) (*monitoringv1.PrometheusStatus, []v1.Pod, error) {
	res := &monitoringv1.PrometheusStatus{Paused: p.Spec.Paused}

	pods, err := kclient.Core().Pods(p.Namespace).List(ListOptions(p.Name))
	if err != nil {
		return nil, nil, errors.Wrap(err, "retrieving pods of failed")
	}
	sset, err := kclient.AppsV1beta1().StatefulSets(p.Namespace).Get(statefulSetNameFromPrometheusName(p.Name), metav1.GetOptions{})
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

// needsUpdate checks whether the given pod conforms with the pod template spec
// for various attributes that are influenced by the Prometheus CRD settings.
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

func (c *Operator) destroyPrometheus(key string) error {
	ssetKey := prometheusKeyToStatefulSetKey(key)
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
	ssetClient := c.kclient.AppsV1beta1().StatefulSets(sset.Namespace)

	if _, err := ssetClient.Update(sset); err != nil {
		return errors.Wrap(err, "updating statefulset for scale-down failed")
	}

	podClient := c.kclient.Core().Pods(sset.Namespace)

	// TODO(fabxc): temprorary solution until StatefulSet status provides necessary info to know
	// whether scale-down completed.
	for {
		pods, err := podClient.List(ListOptions(prometheusNameFromStatefulSetName(sset.Name)))
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

	// Delete the auto-generate configuration.
	// TODO(fabxc): add an ownerRef at creation so we don't delete Secrets
	// manually created for Prometheus servers with no ServiceMonitor selectors.
	s := c.kclient.Core().Secrets(sset.Namespace)
	secret, err := s.Get(sset.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "retrieving config Secret failed")
	}
	if apierrors.IsNotFound(err) {
		// Secret does not exist so nothing to clean up
		return nil
	}

	value, found := secret.Labels[managedByOperatorLabel]
	if found && value == managedByOperatorLabelValue {
		if err := s.Delete(sset.Name, nil); err != nil {
			return errors.Wrap(err, "deleting config Secret failed")
		}
	}

	return nil
}

func (c *Operator) loadBasicAuthSecrets(mons map[string]*monitoringv1.ServiceMonitor, s *v1.SecretList) (map[string]BasicAuthCredentials, error) {

	secrets := map[string]BasicAuthCredentials{}

	for _, mon := range mons {

		for i, ep := range mon.Spec.Endpoints {

			if ep.BasicAuth != nil {

				var username string
				var password string

				for _, secret := range s.Items {

					if secret.Name == ep.BasicAuth.Username.Name {

						if u, ok := secret.Data[ep.BasicAuth.Username.Key]; ok {
							username = string(u)
						} else {
							return nil, fmt.Errorf("Secret password of servicemonitor %s not found.", mon.Name)
						}

					}

					if secret.Name == ep.BasicAuth.Password.Name {

						if p, ok := secret.Data[ep.BasicAuth.Password.Key]; ok {
							password = string(p)
						} else {
							return nil, fmt.Errorf("Secret username of servicemonitor %s not found.",
								mon.Name)
						}

					}
				}

				if username == "" && password == "" {
					return nil, fmt.Errorf("Could not generate basicAuth for servicemonitor %s. Username and password are empty.",
						mon.Name)
				} else {
					secrets[fmt.Sprintf("%s/%s/%d", mon.Namespace, mon.Name, i)] =
						BasicAuthCredentials{
							username: username,
							password: password,
						}
				}

			}
		}
	}

	return secrets, nil

}

func (c *Operator) createConfig(p *monitoringv1.Prometheus, ruleFileConfigMaps []*v1.ConfigMap) error {
	smons, err := c.selectServiceMonitors(p)
	if err != nil {
		return errors.Wrap(err, "selecting ServiceMonitors failed")
	}

	sClient := c.kclient.CoreV1().Secrets(p.Namespace)

	listSecrets, err := sClient.List(metav1.ListOptions{})

	if err != nil {
		return err
	}

	basicAuthSecrets, err := c.loadBasicAuthSecrets(smons, listSecrets)

	if err != nil {
		return err
	}

	// Update secret based on the most recent configuration.
	conf, err := generateConfig(p, smons, len(ruleFileConfigMaps), basicAuthSecrets)
	if err != nil {
		return errors.Wrap(err, "generating config failed")
	}

	s, err := makeConfigSecret(p.Name, ruleFileConfigMaps)
	if err != nil {
		return errors.Wrap(err, "generating base secret failed")
	}
	s.ObjectMeta.Annotations = map[string]string{
		"generated": "true",
	}
	s.Data[configFilename] = []byte(conf)

	curSecret, err := sClient.Get(s.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		c.logger.Log("msg", "creating configuration")
		_, err = sClient.Create(s)
		return err
	}

	generatedConf := s.Data[configFilename]
	generatedConfigMaps := s.Data[configMapsFilename]
	curConfig, curConfigFound := curSecret.Data[configFilename]
	curConfigMaps, curConfigMapsFound := curSecret.Data[configMapsFilename]
	if curConfigFound && curConfigMapsFound {
		if bytes.Equal(curConfig, generatedConf) && bytes.Equal(curConfigMaps, generatedConfigMaps) {
			c.logger.Log("msg", "updating config skipped, no configuration change")
			return nil
		} else {
			c.logger.Log("msg", "current config or current configmaps has changed")
		}
	} else {
		c.logger.Log("msg", "no current config or current configmaps found", "currentConfigFound", curConfigFound, "currentConfigMapsFound", curConfigMapsFound)
	}

	c.logger.Log("msg", "updating configuration")
	_, err = sClient.Update(s)
	return err
}

func (c *Operator) selectServiceMonitors(p *monitoringv1.Prometheus) (map[string]*monitoringv1.ServiceMonitor, error) {
	// Selectors might overlap. Deduplicate them along the keyFunc.
	res := make(map[string]*monitoringv1.ServiceMonitor)

	selector, err := metav1.LabelSelectorAsSelector(p.Spec.ServiceMonitorSelector)
	if err != nil {
		return nil, err
	}

	// Only service monitors within the same namespace as the Prometheus
	// object can belong to it.
	cache.ListAllByNamespace(c.smonInf.GetIndexer(), p.Namespace, selector, func(obj interface{}) {
		k, ok := c.keyFunc(obj)
		if ok {
			res[k] = obj.(*monitoringv1.ServiceMonitor)
		}
	})

	return res, nil
}

func (c *Operator) createTPRs() error {
	tprs := []*extensionsobjold.ThirdPartyResource{
		k8sutil.NewPrometheusTPRDefinition(),
		k8sutil.NewServiceMonitorTPRDefinition(),
	}
	tprClient := c.kclient.Extensions().ThirdPartyResources()

	for _, tpr := range tprs {
		if _, err := tprClient.Create(tpr); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
		c.logger.Log("msg", "TPR created", "tpr", tpr.Name)
	}

	// We have to wait for the TPRs to be ready. Otherwise the initial watch may fail.
	err := k8sutil.WaitForCRDReady(c.mclient.MonitoringV1alpha1().Prometheuses(api.NamespaceAll).List)
	if err != nil {
		return err
	}
	return k8sutil.WaitForCRDReady(c.mclient.MonitoringV1alpha1().ServiceMonitors(api.NamespaceAll).List)
}

func (c *Operator) createCRDs() error {
	crds := []*extensionsobj.CustomResourceDefinition{
		k8sutil.NewPrometheusCustomResourceDefinition(),
		k8sutil.NewServiceMonitorCustomResourceDefinition(),
	}

	crdClient := c.crdclient.ApiextensionsV1beta1().CustomResourceDefinitions()

	for _, crd := range crds {
		if _, err := crdClient.Create(crd); err != nil && !apierrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "creating CRD: %s", crd.Spec.Names.Kind)
		}
		c.logger.Log("msg", "CRD created", "crd", crd.Spec.Names.Kind)
	}

	// We have to wait for the CRDs to be ready. Otherwise the initial watch may fail.
	err := k8sutil.WaitForCRDReady(c.mclient.MonitoringV1().Prometheuses(api.NamespaceAll).List)
	if err != nil {
		return err
	}
	return k8sutil.WaitForCRDReady(c.mclient.MonitoringV1().ServiceMonitors(api.NamespaceAll).List)
}
