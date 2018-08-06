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

	"github.com/coreos/prometheus-operator/pkg/client/monitoring"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
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
	ruleInf cache.SharedIndexInformer
	cmapInf cache.SharedIndexInformer
	secrInf cache.SharedIndexInformer
	ssetInf cache.SharedIndexInformer
	nsInf   cache.SharedIndexInformer

	queue workqueue.RateLimitingInterface

	statefulsetErrors prometheus.Counter

	// triggerByCounterVec is a set of counters keeping track of the amount of
	// times Prometheus Operator was triggered to reconcile its created objects. It
	// is split in the dimensions of Kubernetes objects and corresponding actions
	// (add, delete, update).
	triggerByCounterVec *prometheus.CounterVec

	host                   string
	kubeletObjectName      string
	kubeletObjectNamespace string
	kubeletSyncEnabled     bool
	config                 Config

	configGenerator *configGenerator
}

type Labels struct {
	LabelsString string
	LabelsMap    map[string]string
}

// Implement the flag.Value interface
func (labels *Labels) String() string {
	return labels.LabelsString
}

// Merge labels create a new map with labels merged.
func (labels *Labels) Merge(otherLabels map[string]string) map[string]string {
	mergedLabels := map[string]string{}

	for key, value := range otherLabels {
		mergedLabels[key] = value
	}

	for key, value := range labels.LabelsMap {
		mergedLabels[key] = value
	}
	return mergedLabels
}

// Set implements the flag.Set interface.
func (labels *Labels) Set(value string) error {
	m := map[string]string{}
	if value != "" {
		splited := strings.Split(value, ",")
		for _, pair := range splited {
			sp := strings.Split(pair, "=")
			m[sp[0]] = sp[1]
		}
	}
	(*labels).LabelsMap = m
	(*labels).LabelsString = value
	return nil
}

// Config defines configuration parameters for the Operator.
type Config struct {
	Host                         string
	KubeletObject                string
	TLSInsecure                  bool
	TLSConfig                    rest.TLSClientConfig
	ConfigReloaderImage          string
	PrometheusConfigReloader     string
	AlertmanagerDefaultBaseImage string
	PrometheusDefaultBaseImage   string
	ThanosDefaultBaseImage       string
	Namespace                    string
	Labels                       Labels
	CrdGroup                     string
	CrdKinds                     monitoringv1.CrdKinds
	EnableValidation             bool
	DisableAutoUserGroup         bool
	LocalHost                    string
	LogLevel                     string
	LogFormat                    string
	ManageCRDs                   bool
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

	mclient, err := monitoring.NewForConfig(&conf.CrdKinds, conf.CrdGroup, cfg)
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
		configGenerator:        NewConfigGenerator(logger),
	}

	c.promInf = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  mclient.MonitoringV1().Prometheuses(c.config.Namespace).List,
			WatchFunc: mclient.MonitoringV1().Prometheuses(c.config.Namespace).Watch,
		},
		&monitoringv1.Prometheus{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	c.promInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handlePrometheusAdd,
		DeleteFunc: c.handlePrometheusDelete,
		UpdateFunc: c.handlePrometheusUpdate,
	})

	c.smonInf = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  mclient.MonitoringV1().ServiceMonitors(c.config.Namespace).List,
			WatchFunc: mclient.MonitoringV1().ServiceMonitors(c.config.Namespace).Watch,
		},
		&monitoringv1.ServiceMonitor{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	c.smonInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleSmonAdd,
		DeleteFunc: c.handleSmonDelete,
		UpdateFunc: c.handleSmonUpdate,
	})

	c.ruleInf = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  mclient.MonitoringV1().PrometheusRules(c.config.Namespace).List,
			WatchFunc: mclient.MonitoringV1().PrometheusRules(c.config.Namespace).Watch,
		},
		&monitoringv1.PrometheusRule{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	c.ruleInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleRuleAdd,
		DeleteFunc: c.handleRuleDelete,
		UpdateFunc: c.handleRuleUpdate,
	})

	c.cmapInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.Core().RESTClient(), "configmaps", c.config.Namespace, fields.Everything()),
		&v1.ConfigMap{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	c.cmapInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleConfigMapAdd,
		DeleteFunc: c.handleConfigMapDelete,
		UpdateFunc: c.handleConfigMapUpdate,
	})
	c.secrInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.Core().RESTClient(), "secrets", c.config.Namespace, fields.Everything()),
		&v1.Secret{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	c.secrInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleSecretAdd,
		DeleteFunc: c.handleSecretDelete,
		UpdateFunc: c.handleSecretUpdate,
	})

	c.ssetInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.AppsV1beta2().RESTClient(), "statefulsets", c.config.Namespace, fields.Everything()),
		&appsv1.StatefulSet{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	c.ssetInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleStatefulSetAdd,
		DeleteFunc: c.handleStatefulSetDelete,
		UpdateFunc: c.handleStatefulSetUpdate,
	})

	c.nsInf = cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(c.kclient.Core().RESTClient(), "namespaces", metav1.NamespaceAll, fields.Everything()),
		&v1.Namespace{}, resyncPeriod, cache.Indexers{},
	)

	return c, nil
}

// RegisterMetrics registers Prometheus metrics on the given Prometheus registerer.
func (c *Operator) RegisterMetrics(r prometheus.Registerer) {
	c.statefulsetErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_prometheus_reconcile_errors_total",
		Help: "Number of errors that occurred while reconciling the alertmanager statefulset",
	})

	c.triggerByCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "prometheus_operator_triggered_total",
		Help: "Number of times a Kubernetes object add, delete or update event" +
			" triggered Prometheus Operator to reconcile an object",
	},
		[]string{"triggered_by", "action"},
	)

	r.MustRegister(
		c.statefulsetErrors,
		c.triggerByCounterVec,
		NewPrometheusCollector(c.promInf.GetStore()),
	)
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
		level.Info(c.logger).Log("msg", "connection established", "cluster-version", v)

		if c.config.ManageCRDs {
			if err := c.createCRDs(); err != nil {
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
		level.Info(c.logger).Log("msg", "CRD API endpoints ready")
	case <-stopc:
		return nil
	}

	go c.worker()

	go c.promInf.Run(stopc)
	go c.smonInf.Run(stopc)
	go c.ruleInf.Run(stopc)
	go c.cmapInf.Run(stopc)
	go c.secrInf.Run(stopc)
	go c.ssetInf.Run(stopc)
	if c.config.Namespace == v1.NamespaceAll {
		go c.nsInf.Run(stopc)
	}

	if c.kubeletSyncEnabled {
		go c.reconcileNodeEndpoints(stopc)
	}

	<-stopc
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

func (c *Operator) handlePrometheusAdd(obj interface{}) {
	key, ok := c.keyFunc(obj)
	if !ok {
		return
	}

	level.Debug(c.logger).Log("msg", "Prometheus added", "key", key)
	c.triggerByCounterVec.WithLabelValues(monitoringv1.PrometheusesKind, "add").Inc()
	c.enqueue(key)
}

func (c *Operator) handlePrometheusDelete(obj interface{}) {
	key, ok := c.keyFunc(obj)
	if !ok {
		return
	}

	level.Debug(c.logger).Log("msg", "Prometheus deleted", "key", key)
	c.triggerByCounterVec.WithLabelValues(monitoringv1.PrometheusesKind, "delete").Inc()
	c.enqueue(key)
}

func (c *Operator) handlePrometheusUpdate(old, cur interface{}) {
	if old.(*monitoringv1.Prometheus).ResourceVersion == cur.(*monitoringv1.Prometheus).ResourceVersion {
		return
	}

	key, ok := c.keyFunc(cur)
	if !ok {
		return
	}

	level.Debug(c.logger).Log("msg", "Prometheus updated", "key", key)
	c.triggerByCounterVec.WithLabelValues(monitoringv1.PrometheusesKind, "update").Inc()
	c.enqueue(key)
}

func (c *Operator) reconcileNodeEndpoints(stopc <-chan struct{}) {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-stopc:
			return
		case <-ticker.C:
			err := c.syncNodeEndpoints()
			if err != nil {
				level.Error(c.logger).Log("msg", "syncing nodes into Endpoints object failed", "err", err)
			}
		}
	}
}

// nodeAddresses returns the provided node's address, based on the priority:
// 1. NodeInternalIP
// 2. NodeExternalIP
// 3. NodeLegacyHostIP
// 3. NodeHostName
//
// Copied from github.com/prometheus/prometheus/discovery/kubernetes/node.go
func nodeAddress(node v1.Node) (string, map[v1.NodeAddressType][]string, error) {
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
	if addresses, ok := m[v1.NodeHostName]; ok {
		return addresses[0], m, nil
	}
	return "", m, fmt.Errorf("host address unknown")
}

func (c *Operator) syncNodeEndpoints() error {
	eps := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.kubeletObjectName,
			Labels: c.config.Labels.Merge(map[string]string{
				"k8s-app": "kubelet",
			}),
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

	nodes, err := c.kclient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "listing nodes failed")
	}

	for _, n := range nodes.Items {
		address, _, err := nodeAddress(n)
		if err != nil {
			return errors.Wrapf(err, "failed to determine hostname for node (%s)", n.Name)
		}
		eps.Subsets[0].Addresses = append(eps.Subsets[0].Addresses, v1.EndpointAddress{
			IP: address,
			TargetRef: &v1.ObjectReference{
				Kind:       "Node",
				Name:       n.Name,
				UID:        n.UID,
				APIVersion: n.APIVersion,
			},
		})
	}

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.kubeletObjectName,
			Labels: c.config.Labels.Merge(map[string]string{
				"k8s-app": "kubelet",
			}),
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

	err = k8sutil.CreateOrUpdateService(c.kclient.CoreV1().Services(c.kubeletObjectNamespace), svc)
	if err != nil {
		return errors.Wrap(err, "synchronizing kubelet service object failed")
	}

	err = k8sutil.CreateOrUpdateEndpoints(c.kclient.CoreV1().Endpoints(c.kubeletObjectNamespace), eps)
	if err != nil {
		return errors.Wrap(err, "synchronizing kubelet endpoints object failed")
	}

	return nil
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleSmonAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ServiceMonitor added")
		c.triggerByCounterVec.WithLabelValues(monitoringv1.ServiceMonitorsKind, "add").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleSmonUpdate(old, cur interface{}) {
	if old.(*monitoringv1.ServiceMonitor).ResourceVersion == cur.(*monitoringv1.ServiceMonitor).ResourceVersion {
		return
	}

	o, ok := c.getObject(cur)
	if ok {
		level.Debug(c.logger).Log("msg", "ServiceMonitor updated")
		c.triggerByCounterVec.WithLabelValues(monitoringv1.ServiceMonitorsKind, "update").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleSmonDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ServiceMonitor delete")
		c.triggerByCounterVec.WithLabelValues(monitoringv1.ServiceMonitorsKind, "delete").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleRuleAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PrometheusRule added")
		c.triggerByCounterVec.WithLabelValues(monitoringv1.PrometheusRuleKind, "add").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleRuleUpdate(old, cur interface{}) {
	if old.(*monitoringv1.PrometheusRule).ResourceVersion == cur.(*monitoringv1.PrometheusRule).ResourceVersion {
		return
	}

	o, ok := c.getObject(cur)
	if ok {
		level.Debug(c.logger).Log("msg", "PrometheusRule updated")
		c.triggerByCounterVec.WithLabelValues(monitoringv1.PrometheusRuleKind, "update").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleRuleDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PrometheusRule deleted")
		c.triggerByCounterVec.WithLabelValues(monitoringv1.PrometheusRuleKind, "delete").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Do we need to enque secrets just for the namespace or in general?
func (c *Operator) handleSecretDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "Secret deleted")
		c.triggerByCounterVec.WithLabelValues("Secret", "delete").Inc()

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
		c.triggerByCounterVec.WithLabelValues("Secret", "update").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleSecretAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "Secret added")
		c.triggerByCounterVec.WithLabelValues("Secret", "add").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Do we need to enque configmaps just for the namespace or in general?
func (c *Operator) handleConfigMapAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ConfigMap added")
		c.triggerByCounterVec.WithLabelValues("ConfigMap", "add").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleConfigMapDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ConfigMap deleted")
		c.triggerByCounterVec.WithLabelValues("ConfigMap", "delete").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleConfigMapUpdate(old, cur interface{}) {
	if old.(*v1.ConfigMap).ResourceVersion == cur.(*v1.ConfigMap).ResourceVersion {
		return
	}

	o, ok := c.getObject(cur)
	if ok {
		level.Debug(c.logger).Log("msg", "ConfigMap updated")
		c.triggerByCounterVec.WithLabelValues("ConfigMap", "update").Inc()

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
		level.Error(c.logger).Log("msg", "get object failed", "err", err)
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

// enqueueForNamespace enqueues all Prometheus object keys that belong to the
// given namespace or select objects in the given namespace.
func (c *Operator) enqueueForNamespace(nsName string) {
	// This variable is only accessed if the operator is watching all namespaces.
	var ns *v1.Namespace
	// If the operator is watching all namespaces, then it is running the namespace informer
	// and can search for the namespace object in the cache.
	if c.config.Namespace == v1.NamespaceAll {
		nsObject, exists, err := c.nsInf.GetStore().GetByKey(nsName)
		if err != nil {
			level.Error(c.logger).Log(
				"msg", "get namespace to enqueue Prometheus instances failed",
				"err", err,
			)
			return
		}
		if !exists {
			level.Error(c.logger).Log(
				"msg", fmt.Sprintf("get namespace to enqueue Prometheus instances failed: namespace %q does not exist", nsName),
				"err", err,
			)
			return
		}
		ns = nsObject.(*v1.Namespace)
	}

	err := cache.ListAll(c.promInf.GetStore(), labels.Everything(), func(obj interface{}) {
		// Check for Prometheus instances in the NS.
		p := obj.(*monitoringv1.Prometheus)
		if p.Namespace == nsName {
			c.enqueue(p)
			return
		}

		// If the operator is only watching one namespace, then it is not possible to
		// act on a ServiceMonitor or Rule in different namespace, so return early.
		if c.config.Namespace != v1.NamespaceAll {
			return
		}

		// Check for Prometheus instances selecting ServiceMonitors in the NS.
		smNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.ServiceMonitorNamespaceSelector)
		if err != nil {
			level.Error(c.logger).Log(
				"msg", fmt.Sprintf("failed to convert ServiceMonitorNamespaceSelector of %q to selector", p.Name),
				"err", err,
			)
			return
		}

		if smNSSelector.Matches(labels.Set(ns.Labels)) {
			c.enqueue(p)
			return
		}

		// Check for Prometheus instances selecting PrometheusRules in the NS.
		ruleNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.RuleNamespaceSelector)
		if err != nil {
			level.Error(c.logger).Log(
				"msg", fmt.Sprintf("failed to convert RuleNamespaceSelector of %q to selector", p.Name),
				"err", err,
			)
			return
		}

		if ruleNSSelector.Matches(labels.Set(ns.Labels)) {
			c.enqueue(p)
			return
		}
	})
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "listing all Prometheus instances from cache failed",
			"err", err,
		)
	}
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

	c.statefulsetErrors.Inc()
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
		level.Error(c.logger).Log("msg", "Prometheus lookup failed", "err", err)
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

func (c *Operator) handleStatefulSetDelete(obj interface{}) {
	if ps := c.prometheusForStatefulSet(obj); ps != nil {
		level.Debug(c.logger).Log("msg", "StatefulSet delete")
		c.triggerByCounterVec.WithLabelValues("StatefulSet", "delete").Inc()

		c.enqueue(ps)
	}
}

func (c *Operator) handleStatefulSetAdd(obj interface{}) {
	if ps := c.prometheusForStatefulSet(obj); ps != nil {
		level.Debug(c.logger).Log("msg", "StatefulSet added")
		c.triggerByCounterVec.WithLabelValues("StatefulSet", "add").Inc()

		c.enqueue(ps)
	}
}

func (c *Operator) handleStatefulSetUpdate(oldo, curo interface{}) {
	old := oldo.(*appsv1.StatefulSet)
	cur := curo.(*appsv1.StatefulSet)

	level.Debug(c.logger).Log("msg", "update handler", "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	// Periodic resync may resend the StatefulSet without changes in-between.
	// Also breaks loops created by updating the resource ourselves.
	if old.ResourceVersion == cur.ResourceVersion {
		return
	}

	if ps := c.prometheusForStatefulSet(cur); ps != nil {
		level.Debug(c.logger).Log("msg", "StatefulSet updated")
		c.triggerByCounterVec.WithLabelValues("StatefulSet", "update").Inc()

		c.enqueue(ps)
	}
}

func (c *Operator) sync(key string) error {
	obj, exists, err := c.promInf.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		// Dependent resources are cleaned up by K8s via OwnerReferences
		return nil
	}

	p := obj.(*monitoringv1.Prometheus)
	if p.Spec.Paused {
		return nil
	}

	level.Info(c.logger).Log("msg", "sync prometheus", "key", key)

	ruleConfigMapNames, err := c.createOrUpdateRuleConfigMaps(p)
	if err != nil {
		return err
	}

	// If no service monitor selectors are configured, the user wants to manage
	// configuration themselves.
	if p.Spec.ServiceMonitorSelector != nil {
		// We just always regenerate the configuration to be safe.
		if err := c.createOrUpdateConfigurationSecret(p, ruleConfigMapNames); err != nil {
			return errors.Wrap(err, "creating config failed")
		}
	}

	// Create empty Secret if it doesn't exist. See comment above.
	s, err := makeEmptyConfigurationSecret(p, c.config)
	if err != nil {
		return errors.Wrap(err, "generating empty config secret failed")
	}
	sClient := c.kclient.CoreV1().Secrets(p.Namespace)
	_, err = sClient.Get(s.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		if _, err := c.kclient.Core().Secrets(p.Namespace).Create(s); err != nil && !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "creating empty config file failed")
		}
		return err
	}

	// Create governing service if it doesn't exist.
	svcClient := c.kclient.Core().Services(p.Namespace)
	if err := k8sutil.CreateOrUpdateService(svcClient, makeStatefulSetService(p, c.config)); err != nil {
		return errors.Wrap(err, "synchronizing governing service failed")
	}

	ssetClient := c.kclient.AppsV1beta2().StatefulSets(p.Namespace)
	// Ensure we have a StatefulSet running Prometheus deployed.
	obj, exists, err = c.ssetInf.GetIndexer().GetByKey(prometheusKeyToStatefulSetKey(key))
	if err != nil {
		return errors.Wrap(err, "retrieving statefulset failed")
	}

	newSSetInputHash, err := createSSetInputHash(*p, c.config, ruleConfigMapNames)
	if err != nil {
		return err
	}

	if !exists {
		level.Debug(c.logger).Log("msg", "no current Prometheus statefulset found")
		sset, err := makeStatefulSet(*p, "", &c.config, ruleConfigMapNames, newSSetInputHash)
		if err != nil {
			return errors.Wrap(err, "making statefulset failed")
		}

		level.Debug(c.logger).Log("msg", "creating Prometheus statefulset")
		if _, err := ssetClient.Create(sset); err != nil {
			return errors.Wrap(err, "creating statefulset failed")
		}
		return nil
	}

	oldSSetInputHash := obj.(*appsv1.StatefulSet).ObjectMeta.Annotations[sSetInputHashName]
	if newSSetInputHash == oldSSetInputHash {
		level.Debug(c.logger).Log("msg", "new statefulset generation inputs match current, skipping any actions")
		return nil
	}

	sset, err := makeStatefulSet(*p, obj.(*appsv1.StatefulSet).Spec.PodManagementPolicy, &c.config, ruleConfigMapNames, newSSetInputHash)
	if err != nil {
		return errors.Wrap(err, "making statefulset failed")
	}

	level.Debug(c.logger).Log("msg", "updating current Prometheus statefulset")
	if _, err := ssetClient.Update(sset); err != nil {
		return errors.Wrap(err, "updating statefulset failed")
	}

	return nil
}

func createSSetInputHash(p monitoringv1.Prometheus, c Config, ruleConfigMapNames []string) (string, error) {
	hash, err := hashstructure.Hash(struct {
		P monitoringv1.Prometheus
		C Config
		R []string `hash:"set"`
	}{p, c, ruleConfigMapNames},
		nil,
	)
	if err != nil {
		return "", errors.Wrap(
			err,
			"failed to calculate combined hash of Prometheus CRD, config and"+
				" rule ConfigMap names",
		)
	}

	return fmt.Sprintf("%d", hash), nil
}

func ListOptions(name string) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
			"app":        "prometheus",
			"prometheus": name,
		})).String(),
	}
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
	sset, err := kclient.AppsV1beta2().StatefulSets(p.Namespace).Get(statefulSetNameFromPrometheusName(p.Name), metav1.GetOptions{})
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

func loadAdditionalScrapeConfigsSecret(additionalScrapeConfigs *v1.SecretKeySelector, s *v1.SecretList) ([]byte, error) {
	if additionalScrapeConfigs != nil {
		for _, secret := range s.Items {
			if secret.Name == additionalScrapeConfigs.Name {
				if c, ok := secret.Data[additionalScrapeConfigs.Key]; ok {
					return c, nil
				}

				return nil, fmt.Errorf("key %v could not be found in Secret %v", additionalScrapeConfigs.Key, additionalScrapeConfigs.Name)
			}
		}
		return nil, fmt.Errorf("secret %v could not be found", additionalScrapeConfigs.Name)
	}
	return nil, nil
}

func loadBasicAuthSecret(basicAuth *monitoringv1.BasicAuth, s *v1.SecretList) (BasicAuthCredentials, error) {
	var username string
	var password string

	for _, secret := range s.Items {

		if secret.Name == basicAuth.Username.Name {

			if u, ok := secret.Data[basicAuth.Username.Key]; ok {
				username = string(u)
			} else {
				return BasicAuthCredentials{}, fmt.Errorf("secret username key %q in secret %q not found", basicAuth.Username.Key, secret.Name)
			}

		}

		if secret.Name == basicAuth.Password.Name {

			if p, ok := secret.Data[basicAuth.Password.Key]; ok {
				password = string(p)
			} else {
				return BasicAuthCredentials{}, fmt.Errorf("secret password key %q in secret %q not found", basicAuth.Password.Key, secret.Name)
			}

		}
		if username != "" && password != "" {
			break
		}
	}

	if username == "" && password == "" {
		return BasicAuthCredentials{}, fmt.Errorf("basic auth username and password secret not found")
	}

	return BasicAuthCredentials{username: username, password: password}, nil

}

func (c *Operator) loadBasicAuthSecrets(
	mons map[string]*monitoringv1.ServiceMonitor,
	remoteReads []monitoringv1.RemoteReadSpec,
	remoteWrites []monitoringv1.RemoteWriteSpec,
	apiserverConfig *monitoringv1.APIServerConfig,
	SecretsInPromNS *v1.SecretList,
) (map[string]BasicAuthCredentials, error) {

	sMonSecretMap := make(map[string]*v1.SecretList)
	for _, mon := range mons {
		smNamespace := mon.Namespace
		if sMonSecretMap[smNamespace] == nil {
			msClient := c.kclient.CoreV1().Secrets(smNamespace)
			listSecrets, err := msClient.List(metav1.ListOptions{})

			if err != nil {
				return nil, errors.Wrapf(err, "failed to retrieve secrets in namespace '%v' for servicemonitor '%v'", smNamespace, mon.Name)
			}
			sMonSecretMap[smNamespace] = listSecrets
		}
	}

	secrets := map[string]BasicAuthCredentials{}
	for _, mon := range mons {
		for i, ep := range mon.Spec.Endpoints {
			if ep.BasicAuth != nil {
				credentials, err := loadBasicAuthSecret(ep.BasicAuth, sMonSecretMap[mon.Namespace])
				if err != nil {
					return nil, fmt.Errorf("could not generate basicAuth for servicemonitor %s. %s", mon.Name, err)
				}
				secrets[fmt.Sprintf("serviceMonitor/%s/%s/%d", mon.Namespace, mon.Name, i)] = credentials
			}
		}
	}

	for i, remote := range remoteReads {
		if remote.BasicAuth != nil {
			credentials, err := loadBasicAuthSecret(remote.BasicAuth, SecretsInPromNS)
			if err != nil {
				return nil, fmt.Errorf("could not generate basicAuth for remote_read config %d. %s", i, err)
			}
			secrets[fmt.Sprintf("remoteRead/%d", i)] = credentials
		}
	}

	for i, remote := range remoteWrites {
		if remote.BasicAuth != nil {
			credentials, err := loadBasicAuthSecret(remote.BasicAuth, SecretsInPromNS)
			if err != nil {
				return nil, fmt.Errorf("could not generate basicAuth for remote_write config %d. %s", i, err)
			}
			secrets[fmt.Sprintf("remoteWrite/%d", i)] = credentials
		}
	}

	// load apiserver basic auth secret
	if apiserverConfig != nil && apiserverConfig.BasicAuth != nil {
		credentials, err := loadBasicAuthSecret(apiserverConfig.BasicAuth, SecretsInPromNS)
		if err != nil {
			return nil, fmt.Errorf("could not generate basicAuth for apiserver config. %s", err)
		}
		secrets["apiserver"] = credentials
	}

	return secrets, nil

}

func (c *Operator) createOrUpdateConfigurationSecret(p *monitoringv1.Prometheus, ruleConfigMapNames []string) error {
	smons, err := c.selectServiceMonitors(p)
	if err != nil {
		return errors.Wrap(err, "selecting ServiceMonitors failed")
	}

	sClient := c.kclient.CoreV1().Secrets(p.Namespace)
	SecretsInPromNS, err := sClient.List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	basicAuthSecrets, err := c.loadBasicAuthSecrets(smons, p.Spec.RemoteRead, p.Spec.RemoteWrite, p.Spec.APIServerConfig, SecretsInPromNS)
	if err != nil {
		return err
	}

	additionalScrapeConfigs, err := loadAdditionalScrapeConfigsSecret(p.Spec.AdditionalScrapeConfigs, SecretsInPromNS)
	if err != nil {
		return errors.Wrap(err, "loading additional scrape configs from Secret failed")
	}
	additionalAlertManagerConfigs, err := loadAdditionalScrapeConfigsSecret(p.Spec.AdditionalAlertManagerConfigs, SecretsInPromNS)
	if err != nil {
		return errors.Wrap(err, "loading additional alert manager configs from Secret failed")
	}

	// Update secret based on the most recent configuration.
	conf, err := c.configGenerator.generateConfig(
		p,
		smons,
		basicAuthSecrets,
		additionalScrapeConfigs,
		additionalAlertManagerConfigs,
		ruleConfigMapNames,
	)
	if err != nil {
		return errors.Wrap(err, "generating config failed")
	}

	s := makeConfigSecret(p, c.config)
	s.ObjectMeta.Annotations = map[string]string{
		"generated": "true",
	}
	s.Data[configFilename] = []byte(conf)

	curSecret, err := sClient.Get(s.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		level.Debug(c.logger).Log("msg", "creating configuration")
		_, err = sClient.Create(s)
		return err
	}

	var (
		generatedConf             = s.Data[configFilename]
		curConfig, curConfigFound = curSecret.Data[configFilename]
	)
	if curConfigFound {
		if bytes.Equal(curConfig, generatedConf) {
			level.Debug(c.logger).Log("msg", "updating Prometheus configuration secret skipped, no configuration change")
			return nil
		}
		level.Debug(c.logger).Log("msg", "current Prometheus configuration has changed")
	} else {
		level.Debug(c.logger).Log("msg", "no current Prometheus configuration secret found", "currentConfigFound", curConfigFound)
	}

	level.Debug(c.logger).Log("msg", "updating Prometheus configuration secret")
	_, err = sClient.Update(s)
	return err
}

func (c *Operator) selectServiceMonitors(p *monitoringv1.Prometheus) (map[string]*monitoringv1.ServiceMonitor, error) {
	namespaces := []string{}
	// Selectors might overlap. Deduplicate them along the keyFunc.
	res := make(map[string]*monitoringv1.ServiceMonitor)

	servMonSelector, err := metav1.LabelSelectorAsSelector(p.Spec.ServiceMonitorSelector)
	if err != nil {
		return nil, err
	}

	// If 'ServiceMonitorNamespaceSelector' is nil, only check own namespace.
	if p.Spec.ServiceMonitorNamespaceSelector == nil {
		namespaces = append(namespaces, p.Namespace)
	} else {
		servMonNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.ServiceMonitorNamespaceSelector)
		if err != nil {
			return nil, err
		}

		namespaces, err = c.listNamespaces(servMonNSSelector)
		if err != nil {
			return nil, err
		}
	}

	level.Debug(c.logger).Log("msg", "filtering namespaces to select ServiceMonitors from", "namespaces", strings.Join(namespaces, ","), "namespace", p.Namespace, "prometheus", p.Name)

	for _, ns := range namespaces {
		cache.ListAllByNamespace(c.smonInf.GetIndexer(), ns, servMonSelector, func(obj interface{}) {
			k, ok := c.keyFunc(obj)
			if ok {
				res[k] = obj.(*monitoringv1.ServiceMonitor)
			}
		})
	}

	serviceMonitors := []string{}
	for k := range res {
		serviceMonitors = append(serviceMonitors, k)
	}
	level.Debug(c.logger).Log("msg", "selected ServiceMonitors", "servicemonitors", strings.Join(serviceMonitors, ","), "namespace", p.Namespace, "prometheus", p.Name)

	return res, nil
}

// listNamespaces lists all the namespaces that match the provided selector.
// If the operator is watching a single namespace rather than all namespaces,
// then this func returns only the watched namespace, since all reconciliation
// must occur in this namespace.
func (c *Operator) listNamespaces(selector labels.Selector) ([]string, error) {
	if c.config.Namespace != v1.NamespaceAll {
		return []string{c.config.Namespace}, nil
	}
	var ns []string
	err := cache.ListAll(c.nsInf.GetStore(), selector, func(obj interface{}) {
		ns = append(ns, obj.(*v1.Namespace).Name)
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list namespaces")
	}
	return ns, nil
}

func (c *Operator) createCRDs() error {
	crds := []*extensionsobj.CustomResourceDefinition{
		k8sutil.NewCustomResourceDefinition(c.config.CrdKinds.Prometheus, c.config.CrdGroup, c.config.Labels.LabelsMap, c.config.EnableValidation),
		k8sutil.NewCustomResourceDefinition(c.config.CrdKinds.ServiceMonitor, c.config.CrdGroup, c.config.Labels.LabelsMap, c.config.EnableValidation),
		k8sutil.NewCustomResourceDefinition(c.config.CrdKinds.PrometheusRule, c.config.CrdGroup, c.config.Labels.LabelsMap, c.config.EnableValidation),
	}

	crdClient := c.crdclient.ApiextensionsV1beta1().CustomResourceDefinitions()

	for _, crd := range crds {
		oldCRD, err := crdClient.Get(crd.Name, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "getting CRD: %s", crd.Spec.Names.Kind)
		}
		if apierrors.IsNotFound(err) {
			if _, err := crdClient.Create(crd); err != nil {
				return errors.Wrapf(err, "creating CRD: %s", crd.Spec.Names.Kind)
			}
			level.Info(c.logger).Log("msg", "CRD created", "crd", crd.Spec.Names.Kind)
		}
		if err == nil {
			crd.ResourceVersion = oldCRD.ResourceVersion
			if _, err := crdClient.Update(crd); err != nil {
				return errors.Wrapf(err, "creating CRD: %s", crd.Spec.Names.Kind)
			}
			level.Info(c.logger).Log("msg", "CRD updated", "crd", crd.Spec.Names.Kind)
		}
	}

	crdListFuncs := []struct {
		name     string
		listFunc func(opts metav1.ListOptions) (runtime.Object, error)
	}{
		{"Prometheus", c.mclient.MonitoringV1().Prometheuses(c.config.Namespace).List},
		{"ServiceMonitor", c.mclient.MonitoringV1().ServiceMonitors(c.config.Namespace).List},
		{"PrometheusRule", c.mclient.MonitoringV1().PrometheusRules(c.config.Namespace).List},
	}

	for _, crdListFunc := range crdListFuncs {
		err := k8sutil.WaitForCRDReady(crdListFunc.listFunc)
		if err != nil {
			return errors.Wrapf(err, "waiting for %v crd failed", crdListFunc.name)
		}
	}

	return nil
}
