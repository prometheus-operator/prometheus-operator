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
	"compress/gzip"
	"fmt"
	"reflect"
	"strings"
	"time"

	monitoring "github.com/coreos/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringclient "github.com/coreos/prometheus-operator/pkg/client/versioned"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/coreos/prometheus-operator/pkg/listwatch"
	"github.com/coreos/prometheus-operator/pkg/operator"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/mitchellh/hashstructure"
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
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	resyncPeriod = 5 * time.Minute
)

// Operator manages life cycle of Prometheus deployments and
// monitoring configurations.
type Operator struct {
	kclient   kubernetes.Interface
	mclient   monitoringclient.Interface
	crdclient apiextensionsclient.Interface
	logger    log.Logger

	promInf cache.SharedIndexInformer
	smonInf cache.SharedIndexInformer
	pmonInf cache.SharedIndexInformer
	ruleInf cache.SharedIndexInformer
	cmapInf cache.SharedIndexInformer
	secrInf cache.SharedIndexInformer
	ssetInf cache.SharedIndexInformer
	nsInf   cache.SharedIndexInformer

	queue workqueue.RateLimitingInterface

	metrics *operator.Metrics

	nodeAddressLookupErrors prometheus.Counter
	nodeEndpointSyncs       prometheus.Counter
	nodeEndpointSyncErrors  prometheus.Counter

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
	Host                          string
	KubeletObject                 string
	TLSInsecure                   bool
	TLSConfig                     rest.TLSClientConfig
	ConfigReloaderImage           string
	ConfigReloaderCPU             string
	ConfigReloaderMemory          string
	PrometheusConfigReloaderImage string
	AlertmanagerDefaultBaseImage  string
	PrometheusDefaultBaseImage    string
	ThanosDefaultBaseImage        string
	Namespaces                    Namespaces
	Labels                        Labels
	CrdKinds                      monitoringv1.CrdKinds
	EnableValidation              bool
	LocalHost                     string
	LogLevel                      string
	LogFormat                     string
	ManageCRDs                    bool
	PromSelector                  string
	AlertManagerSelector          string
	ThanosRulerSelector           string
}

type Namespaces struct {
	// allow list/deny list for common custom resources
	AllowList, DenyList []string
	// allow list for prometheus/alertmanager custom resources
	PrometheusAllowList, AlertmanagerAllowList, ThanosRulerAllowList []string
}

// BasicAuthCredentials represents a username password pair to be used with
// basic http authentication, see https://tools.ietf.org/html/rfc7617.
type BasicAuthCredentials struct {
	username string
	password string
}

// BearerToken represents a bearer token, see
// https://tools.ietf.org/html/rfc6750.
type BearerToken string

// TLSAsset represents any TLS related opaque string, e.g. CA files, client
// certificates.
type TLSAsset string

// New creates a new controller.
func New(conf Config, logger log.Logger, r prometheus.Registerer) (*Operator, error) {
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

	if _, err := labels.Parse(conf.PromSelector); err != nil {
		return nil, errors.Wrap(err, "can not parse prometheus selector value")
	}

	if _, err := labels.Parse(conf.AlertManagerSelector); err != nil {
		return nil, errors.Wrap(err, "can not parse alertmanager selector value")
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
		configGenerator:        newConfigGenerator(logger),
		metrics:                operator.NewMetrics("prometheus", r),
		nodeAddressLookupErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prometheus_operator_node_address_lookup_errors_total",
			Help: "Number of times a node IP address could not be determined",
		}),
		nodeEndpointSyncs: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prometheus_operator_node_syncs_total",
			Help: "Number of node endpoints synchronisations",
		}),
		nodeEndpointSyncErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prometheus_operator_node_syncs_failed_total",
			Help: "Number of node endpoints synchronisation failures",
		}),
	}
	c.metrics.MustRegister(c.nodeAddressLookupErrors, c.nodeEndpointSyncs, c.nodeEndpointSyncErrors)

	c.promInf = cache.NewSharedIndexInformer(
		c.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.PrometheusAllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						options.LabelSelector = c.config.PromSelector
						return mclient.MonitoringV1().Prometheuses(namespace).List(options)
					},
					WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
						options.LabelSelector = c.config.PromSelector
						return mclient.MonitoringV1().Prometheuses(namespace).Watch(options)
					},
				}
			}),
		),
		&monitoringv1.Prometheus{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	c.metrics.MustRegister(NewPrometheusCollector(c.promInf.GetStore()))

	c.smonInf = cache.NewSharedIndexInformer(
		c.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.AllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						return mclient.MonitoringV1().ServiceMonitors(namespace).List(options)
					},
					WatchFunc: mclient.MonitoringV1().ServiceMonitors(namespace).Watch,
				}
			}),
		),
		&monitoringv1.ServiceMonitor{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	c.pmonInf = cache.NewSharedIndexInformer(
		c.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.AllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						return mclient.MonitoringV1().PodMonitors(namespace).List(options)
					},
					WatchFunc: mclient.MonitoringV1().PodMonitors(namespace).Watch,
				}
			}),
		),
		&monitoringv1.PodMonitor{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	c.ruleInf = cache.NewSharedIndexInformer(
		c.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.AllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
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

	c.cmapInf = cache.NewSharedIndexInformer(
		c.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.PrometheusAllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						options.LabelSelector = labelPrometheusName
						return c.kclient.CoreV1().ConfigMaps(namespace).List(options)
					},
					WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
						options.LabelSelector = labelPrometheusName
						return c.kclient.CoreV1().ConfigMaps(namespace).Watch(options)
					},
				}
			}),
		),
		&v1.ConfigMap{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	c.secrInf = cache.NewSharedIndexInformer(
		c.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.PrometheusAllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return cache.NewListWatchFromClient(c.kclient.CoreV1().RESTClient(), "secrets", namespace, fields.Everything())
			}),
		),
		&v1.Secret{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	c.ssetInf = cache.NewSharedIndexInformer(
		c.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.PrometheusAllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return cache.NewListWatchFromClient(c.kclient.AppsV1().RESTClient(), "statefulsets", namespace, fields.Everything())
			}),
		),
		&appsv1.StatefulSet{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	// nsResyncPeriod is used to control how often the namespace informer
	// should resync. If the unprivileged ListerWatcher is used, then the
	// informer must resync more often because it cannot watch for
	// namespace changes.
	nsResyncPeriod := 15 * time.Second
	// If the only namespace is v1.NamespaceAll, then the client must be
	// privileged and a regular cache.ListWatch will be used. In this case
	// watching works and we do not need to resync so frequently.
	if listwatch.IsAllNamespaces(c.config.Namespaces.AllowList) {
		nsResyncPeriod = resyncPeriod
	}
	c.nsInf = cache.NewSharedIndexInformer(
		c.metrics.NewInstrumentedListerWatcher(
			listwatch.NewUnprivilegedNamespaceListWatchFromClient(c.logger, c.kclient.CoreV1().RESTClient(), c.config.Namespaces.AllowList, c.config.Namespaces.DenyList, fields.Everything()),
		),
		&v1.Namespace{}, nsResyncPeriod, cache.Indexers{},
	)

	return c, nil
}

// waitForCacheSync waits for the informers' caches to be synced.
func (c *Operator) waitForCacheSync(stopc <-chan struct{}) error {
	ok := true
	informers := []struct {
		name     string
		informer cache.SharedIndexInformer
	}{
		{"Prometheus", c.promInf},
		{"ServiceMonitor", c.smonInf},
		{"PodMonitor", c.pmonInf},
		{"PrometheusRule", c.ruleInf},
		{"ConfigMap", c.cmapInf},
		{"Secret", c.secrInf},
		{"StatefulSet", c.ssetInf},
		{"Namespace", c.nsInf},
	}
	for _, inf := range informers {
		if !cache.WaitForCacheSync(stopc, inf.informer.HasSynced) {
			level.Error(c.logger).Log("msg", fmt.Sprintf("failed to sync %s cache", inf.name))
			ok = false
		} else {
			level.Debug(c.logger).Log("msg", fmt.Sprintf("successfully synced %s cache", inf.name))
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
	c.promInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handlePrometheusAdd,
		DeleteFunc: c.handlePrometheusDelete,
		UpdateFunc: c.handlePrometheusUpdate,
	})
	c.smonInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleSmonAdd,
		DeleteFunc: c.handleSmonDelete,
		UpdateFunc: c.handleSmonUpdate,
	})
	c.pmonInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handlePmonAdd,
		DeleteFunc: c.handlePmonDelete,
		UpdateFunc: c.handlePmonUpdate,
	})
	c.ruleInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleRuleAdd,
		DeleteFunc: c.handleRuleDelete,
		UpdateFunc: c.handleRuleUpdate,
	})
	c.cmapInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleConfigMapAdd,
		DeleteFunc: c.handleConfigMapDelete,
		UpdateFunc: c.handleConfigMapUpdate,
	})
	c.secrInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleSecretAdd,
		DeleteFunc: c.handleSecretDelete,
		UpdateFunc: c.handleSecretUpdate,
	})
	c.ssetInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleStatefulSetAdd,
		DeleteFunc: c.handleStatefulSetDelete,
		UpdateFunc: c.handleStatefulSetUpdate,
	})
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
	go c.pmonInf.Run(stopc)
	go c.ruleInf.Run(stopc)
	go c.cmapInf.Run(stopc)
	go c.secrInf.Run(stopc)
	go c.ssetInf.Run(stopc)
	go c.nsInf.Run(stopc)
	if err := c.waitForCacheSync(stopc); err != nil {
		return err
	}
	c.addHandlers()

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
	c.metrics.TriggerByCounter(monitoringv1.PrometheusesKind, "add").Inc()
	checkPrometheusSpecDeprecation(key, obj.(*monitoringv1.Prometheus), c.logger)
	c.enqueue(key)
}

func (c *Operator) handlePrometheusDelete(obj interface{}) {
	key, ok := c.keyFunc(obj)
	if !ok {
		return
	}

	level.Debug(c.logger).Log("msg", "Prometheus deleted", "key", key)
	c.metrics.TriggerByCounter(monitoringv1.PrometheusesKind, "delete").Inc()
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
	c.metrics.TriggerByCounter(monitoringv1.PrometheusesKind, "update").Inc()
	checkPrometheusSpecDeprecation(key, cur.(*monitoringv1.Prometheus), c.logger)
	c.enqueue(key)
}

func (c *Operator) reconcileNodeEndpoints(stopc <-chan struct{}) {
	c.syncNodeEndpointsWithLogError()
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-stopc:
			return
		case <-ticker.C:
			c.syncNodeEndpointsWithLogError()
		}
	}
}

// nodeAddresses returns the provided node's address, based on the priority:
// 1. NodeInternalIP
// 2. NodeExternalIP
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
	return "", m, fmt.Errorf("host address unknown")
}

func getNodeAddresses(nodes *v1.NodeList) ([]v1.EndpointAddress, []error) {
	addresses := make([]v1.EndpointAddress, 0)
	errs := make([]error, 0)

	for _, n := range nodes.Items {
		address, _, err := nodeAddress(n)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "failed to determine hostname for node (%s)", n.Name))
			continue
		}
		addresses = append(addresses, v1.EndpointAddress{
			IP: address,
			TargetRef: &v1.ObjectReference{
				Kind:       "Node",
				Name:       n.Name,
				UID:        n.UID,
				APIVersion: n.APIVersion,
			},
		})
	}

	return addresses, errs
}

func (c *Operator) syncNodeEndpointsWithLogError() {
	c.nodeEndpointSyncs.Inc()
	err := c.syncNodeEndpoints()
	if err != nil {
		c.nodeEndpointSyncErrors.Inc()
		level.Error(c.logger).Log("msg", "syncing nodes into Endpoints object failed", "err", err)
	}
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

	addresses, errs := getNodeAddresses(nodes)
	if len(errs) > 0 {
		for _, err := range errs {
			level.Warn(c.logger).Log("err", err)
		}
		c.nodeAddressLookupErrors.Add(float64(len(errs)))
	}
	eps.Subsets[0].Addresses = addresses

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
		c.metrics.TriggerByCounter(monitoringv1.ServiceMonitorsKind, "add").Inc()

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
		c.metrics.TriggerByCounter(monitoringv1.ServiceMonitorsKind, "update").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleSmonDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ServiceMonitor delete")
		c.metrics.TriggerByCounter(monitoringv1.ServiceMonitorsKind, "delete").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handlePmonAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PodMonitor added")
		c.metrics.TriggerByCounter(monitoringv1.PodMonitorsKind, "add").Inc()
		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handlePmonUpdate(old, cur interface{}) {
	if old.(*monitoringv1.PodMonitor).ResourceVersion == cur.(*monitoringv1.PodMonitor).ResourceVersion {
		return
	}

	o, ok := c.getObject(cur)
	if ok {
		level.Debug(c.logger).Log("msg", "PodMonitor updated")
		c.metrics.TriggerByCounter(monitoringv1.PodMonitorsKind, "update").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handlePmonDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PodMonitor delete")
		c.metrics.TriggerByCounter(monitoringv1.PodMonitorsKind, "delete").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleRuleAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PrometheusRule added")
		c.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, "add").Inc()

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
		c.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, "update").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleRuleDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PrometheusRule deleted")
		c.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, "delete").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

// TODO: Do we need to enque secrets just for the namespace or in general?
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

// TODO: Do we need to enque configmaps just for the namespace or in general?
func (c *Operator) handleConfigMapAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ConfigMap added")
		c.metrics.TriggerByCounter("ConfigMap", "add").Inc()

		c.enqueueForNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleConfigMapDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ConfigMap deleted")
		c.metrics.TriggerByCounter("ConfigMap", "delete").Inc()

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
		c.metrics.TriggerByCounter("ConfigMap", "update").Inc()

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

// enqueueForNamespace enqueues all Prometheus object keys that belong to the
// given namespace or select objects in the given namespace.
func (c *Operator) enqueueForNamespace(nsName string) {
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
		)
		return
	}
	ns := nsObject.(*v1.Namespace)

	err = cache.ListAll(c.promInf.GetStore(), labels.Everything(), func(obj interface{}) {
		// Check for Prometheus instances in the NS.
		p := obj.(*monitoringv1.Prometheus)
		if p.Namespace == nsName {
			c.enqueue(p)
			return
		}

		// Check for Prometheus instances selecting ServiceMonitors in
		// the NS.
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

		// Check for Prometheus instances selecting PrometheusRules in
		// the NS.
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

// worker runs a worker thread that just dequeues items, processes them, and
// marks them done. It enforces that the syncHandler is never invoked
// concurrently with the same key.
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

	c.metrics.ReconcileErrorsCounter().Inc()
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
		c.metrics.TriggerByCounter("StatefulSet", "delete").Inc()

		c.enqueue(ps)
	}
}

func (c *Operator) handleStatefulSetAdd(obj interface{}) {
	if ps := c.prometheusForStatefulSet(obj); ps != nil {
		level.Debug(c.logger).Log("msg", "StatefulSet added")
		c.metrics.TriggerByCounter("StatefulSet", "add").Inc()

		c.enqueue(ps)
	}
}

func (c *Operator) handleStatefulSetUpdate(oldo, curo interface{}) {
	old := oldo.(*appsv1.StatefulSet)
	cur := curo.(*appsv1.StatefulSet)

	level.Debug(c.logger).Log("msg", "update handler", "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	// Periodic resync may resend the StatefulSet without changes
	// in-between. Also breaks loops created by updating the resource
	// ourselves.
	if old.ResourceVersion == cur.ResourceVersion {
		return
	}

	if ps := c.prometheusForStatefulSet(cur); ps != nil {
		level.Debug(c.logger).Log("msg", "StatefulSet updated")
		c.metrics.TriggerByCounter("StatefulSet", "update").Inc()

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
	p = p.DeepCopy()
	p.APIVersion = monitoringv1.SchemeGroupVersion.String()
	p.Kind = monitoringv1.PrometheusesKind

	if p.Spec.Paused {
		return nil
	}

	level.Info(c.logger).Log("msg", "sync prometheus", "key", key)
	ruleConfigMapNames, err := c.createOrUpdateRuleConfigMaps(p)
	if err != nil {
		return err
	}

	// If no service monitor selectors are configured, the user wants to
	// manage configuration themselves.
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
		if _, err := c.kclient.CoreV1().Secrets(p.Namespace).Create(s); err != nil && !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "creating empty config file failed")
		}
	}
	if !apierrors.IsNotFound(err) && err != nil {
		return err
	}

	if err := c.createOrUpdateTLSAssetSecret(p); err != nil {
		return errors.Wrap(err, "creating tls asset secret failed")
	}

	// Create governing service if it doesn't exist.
	svcClient := c.kclient.CoreV1().Services(p.Namespace)
	if err := k8sutil.CreateOrUpdateService(svcClient, makeStatefulSetService(p, c.config)); err != nil {
		return errors.Wrap(err, "synchronizing governing service failed")
	}

	ssetClient := c.kclient.AppsV1().StatefulSets(p.Namespace)
	// Ensure we have a StatefulSet running Prometheus deployed.
	obj, exists, err = c.ssetInf.GetIndexer().GetByKey(prometheusKeyToStatefulSetKey(key))
	if err != nil {
		return errors.Wrap(err, "retrieving statefulset failed")
	}

	spec := appsv1.StatefulSetSpec{}
	if obj != nil {
		ss := obj.(*appsv1.StatefulSet)
		spec = ss.Spec
	}
	newSSetInputHash, err := createSSetInputHash(*p, c.config, ruleConfigMapNames, spec)
	if err != nil {
		return err
	}

	sset, err := makeStatefulSet(*p, &c.config, ruleConfigMapNames, newSSetInputHash)
	if err != nil {
		return errors.Wrap(err, "making statefulset failed")
	}
	sanitizeSTS(sset)

	if !exists {
		level.Debug(c.logger).Log("msg", "no current Prometheus statefulset found")
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

	level.Debug(c.logger).Log("msg", "updating current Prometheus statefulset")

	_, err = ssetClient.Update(sset)
	sErr, ok := err.(*apierrors.StatusError)

	if ok && sErr.ErrStatus.Code == 422 && sErr.ErrStatus.Reason == metav1.StatusReasonInvalid {
		c.metrics.StsDeleteCreateCounter().Inc()
		level.Info(c.logger).Log("msg", "resolving illegal update of Prometheus StatefulSet", "details", sErr.ErrStatus.Details)
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

// sanitizeSTS removes values for APIVersion and Kind from the VolumeClaimTemplates.
// This prevents update failures due to these fields changing when applied.
func sanitizeSTS(sts *appsv1.StatefulSet) {
	for i := range sts.Spec.VolumeClaimTemplates {
		sts.Spec.VolumeClaimTemplates[i].APIVersion = ""
		sts.Spec.VolumeClaimTemplates[i].Kind = ""
	}
}

//checkPrometheusSpecDeprecation checks for deprecated fields in the prometheus spec and logs a warning if applicable
func checkPrometheusSpecDeprecation(key string, p *monitoringv1.Prometheus, logger log.Logger) {
	deprecationWarningf := "prometheus key=%v, field %v is deprecated, '%v' field should be used instead"
	if p.Spec.BaseImage != "" {
		level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, key, "spec.baseImage", "spec.image"))
	}
	if p.Spec.Tag != "" {
		level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, key, "spec.tag", "spec.image"))
	}
	if p.Spec.SHA != "" {
		level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, key, "spec.sha", "spec.image"))
	}
	if p.Spec.Thanos != nil {
		if p.Spec.BaseImage != "" {
			level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, key, "spec.thanos.baseImage", "spec.thanos.image"))
		}
		if p.Spec.Tag != "" {
			level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, key, "spec.thanos.tag", "spec.thanos.image"))
		}
		if p.Spec.SHA != "" {
			level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, key, "spec.thanos.sha", "spec.thanos.image"))
		}
	}
}

func createSSetInputHash(p monitoringv1.Prometheus, c Config, ruleConfigMapNames []string, ss interface{}) (string, error) {
	hash, err := hashstructure.Hash(struct {
		P monitoringv1.Prometheus
		C Config
		S interface{}
		R []string `hash:"set"`
	}{p, c, ss, ruleConfigMapNames},
		nil,
	)
	if err != nil {
		return "", errors.Wrap(
			err,
			"failed to calculate combined hash of Prometheus StatefulSet, Prometheus CRD, config and"+
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

// PrometheusStatus evaluates the current status of a Prometheus deployment with
// respect to its specified resource object. It return the status and a list of
// pods that are not updated.
func PrometheusStatus(kclient kubernetes.Interface, p *monitoringv1.Prometheus) (*monitoringv1.PrometheusStatus, []v1.Pod, error) {
	res := &monitoringv1.PrometheusStatus{Paused: p.Spec.Paused}

	pods, err := kclient.CoreV1().Pods(p.Namespace).List(ListOptions(p.Name))
	if err != nil {
		return nil, nil, errors.Wrap(err, "retrieving pods of failed")
	}
	sset, err := kclient.AppsV1().StatefulSets(p.Namespace).Get(statefulSetNameFromPrometheusName(p.Name), metav1.GetOptions{})
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

func (c *Operator) loadAdditionalScrapeConfigsSecret(additionalScrapeConfigs *v1.SecretKeySelector, s *v1.SecretList) ([]byte, error) {
	if additionalScrapeConfigs != nil {
		for _, secret := range s.Items {
			if secret.Name == additionalScrapeConfigs.Name {
				if c, ok := secret.Data[additionalScrapeConfigs.Key]; ok {
					return c, nil
				}

				return nil, fmt.Errorf("key %v could not be found in Secret %v", additionalScrapeConfigs.Key, additionalScrapeConfigs.Name)
			}
		}
		if additionalScrapeConfigs.Optional == nil || !*additionalScrapeConfigs.Optional {
			return nil, fmt.Errorf("secret %v could not be found", additionalScrapeConfigs.Name)
		}
		level.Debug(c.logger).Log("msg", fmt.Sprintf("secret %v could not be found", additionalScrapeConfigs.Name))
	}
	return nil, nil
}

func extractCredKey(secret *v1.Secret, sel v1.SecretKeySelector, cred string) (string, error) {
	if s, ok := secret.Data[sel.Key]; ok {
		return string(s), nil
	}
	return "", fmt.Errorf("secret %s key %q in secret %q not found", cred, sel.Key, sel.Name)
}

func getCredFromSecret(c corev1client.SecretInterface, sel v1.SecretKeySelector, cred string, cacheKey string, cache map[string]*v1.Secret) (_ string, err error) {
	var s *v1.Secret
	var ok bool

	if s, ok = cache[cacheKey]; !ok {
		if s, err = c.Get(sel.Name, metav1.GetOptions{}); err != nil {
			return "", fmt.Errorf("unable to fetch %s secret %q: %s", cred, sel.Name, err)
		}
		cache[cacheKey] = s
	}
	return extractCredKey(s, sel, cred)
}

func getCredFromConfigMap(c corev1client.ConfigMapInterface, sel v1.ConfigMapKeySelector, cred string, cacheKey string, cache map[string]*v1.ConfigMap) (_ string, err error) {
	var s *v1.ConfigMap
	var ok bool

	if s, ok = cache[cacheKey]; !ok {
		if s, err = c.Get(sel.Name, metav1.GetOptions{}); err != nil {
			return "", fmt.Errorf("unable to fetch %s configmap %q: %s", cred, sel.Name, err)
		}
		cache[cacheKey] = s
	}

	if a, ok := s.Data[sel.Key]; ok {
		return string(a), nil
	}
	return "", fmt.Errorf("config %s key %q in configmap %q not found", cred, sel.Key, sel.Name)
}

func loadBasicAuthSecretFromAPI(basicAuth *monitoringv1.BasicAuth, c corev1client.CoreV1Interface, ns string, cache map[string]*v1.Secret) (BasicAuthCredentials, error) {
	var username string
	var password string
	var err error

	sClient := c.Secrets(ns)

	if username, err = getCredFromSecret(sClient, basicAuth.Username, "username", ns+"/"+basicAuth.Username.Name, cache); err != nil {
		return BasicAuthCredentials{}, err
	}

	if password, err = getCredFromSecret(sClient, basicAuth.Password, "password", ns+"/"+basicAuth.Password.Name, cache); err != nil {
		return BasicAuthCredentials{}, err
	}

	return BasicAuthCredentials{username: username, password: password}, nil
}

func loadBasicAuthSecret(basicAuth *monitoringv1.BasicAuth, s *v1.SecretList) (BasicAuthCredentials, error) {
	var username string
	var password string
	var err error

	for _, secret := range s.Items {

		if secret.Name == basicAuth.Username.Name {
			if username, err = extractCredKey(&secret, basicAuth.Username, "username"); err != nil {
				return BasicAuthCredentials{}, err
			}
		}

		if secret.Name == basicAuth.Password.Name {
			if password, err = extractCredKey(&secret, basicAuth.Password, "password"); err != nil {
				return BasicAuthCredentials{}, err
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

func gzipConfig(buf *bytes.Buffer, conf []byte) error {
	w := gzip.NewWriter(buf)
	defer w.Close()
	if _, err := w.Write(conf); err != nil {
		return err
	}
	return nil
}

func (c *Operator) loadBasicAuthSecrets(
	mons map[string]*monitoringv1.ServiceMonitor,
	remoteReads []monitoringv1.RemoteReadSpec,
	remoteWrites []monitoringv1.RemoteWriteSpec,
	apiserverConfig *monitoringv1.APIServerConfig,
	SecretsInPromNS *v1.SecretList,
) (map[string]BasicAuthCredentials, error) {

	secrets := map[string]BasicAuthCredentials{}
	nsSecretCache := make(map[string]*v1.Secret)
	for _, mon := range mons {
		for i, ep := range mon.Spec.Endpoints {
			if ep.BasicAuth != nil {
				credentials, err := loadBasicAuthSecretFromAPI(ep.BasicAuth, c.kclient.CoreV1(), mon.Namespace, nsSecretCache)
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

func (c *Operator) loadBearerTokensFromSecrets(mons map[string]*monitoringv1.ServiceMonitor) (map[string]BearerToken, error) {
	tokens := map[string]BearerToken{}
	nsSecretCache := make(map[string]*v1.Secret)

	for _, mon := range mons {
		for i, ep := range mon.Spec.Endpoints {
			if ep.BearerTokenSecret.Name == "" {
				continue
			}

			sClient := c.kclient.CoreV1().Secrets(mon.Namespace)
			token, err := getCredFromSecret(
				sClient,
				ep.BearerTokenSecret,
				"bearertoken",
				mon.Namespace+"/"+ep.BearerTokenSecret.Name,
				nsSecretCache,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to extract endpoint bearertoken for servicemonitor %v from secret %v in namespace %v",
					mon.Name, ep.BearerTokenSecret.Name, mon.Namespace,
				)
			}

			tokens[fmt.Sprintf("serviceMonitor/%s/%s/%d", mon.Namespace, mon.Name, i)] = BearerToken(token)
		}
	}

	return tokens, nil
}

func (c *Operator) loadTLSAssets(mons map[string]*monitoringv1.ServiceMonitor) (map[string]TLSAsset, error) {
	assets := map[string]TLSAsset{}
	nsSecretCache := make(map[string]*v1.Secret)
	nsConfigMapCache := make(map[string]*v1.ConfigMap)

	for _, mon := range mons {
		for _, ep := range mon.Spec.Endpoints {
			if ep.TLSConfig == nil {
				continue
			}

			prefix := mon.Namespace + "/"
			secretSelectors := map[string]*v1.SecretKeySelector{}
			configMapSelectors := map[string]*v1.ConfigMapKeySelector{}
			if ep.TLSConfig.CA != (monitoringv1.SecretOrConfigMap{}) {
				switch {
				case ep.TLSConfig.CA.Secret != nil:
					secretSelectors[prefix+ep.TLSConfig.CA.Secret.Name+"/"+ep.TLSConfig.CA.Secret.Key] = ep.TLSConfig.CA.Secret
				case ep.TLSConfig.CA.ConfigMap != nil:
					configMapSelectors[prefix+ep.TLSConfig.CA.ConfigMap.Name+"/"+ep.TLSConfig.CA.ConfigMap.Key] = ep.TLSConfig.CA.ConfigMap
				}
			}
			if ep.TLSConfig.Cert != (monitoringv1.SecretOrConfigMap{}) {
				switch {
				case ep.TLSConfig.Cert.Secret != nil:
					secretSelectors[prefix+ep.TLSConfig.Cert.Secret.Name+"/"+ep.TLSConfig.Cert.Secret.Key] = ep.TLSConfig.Cert.Secret
				case ep.TLSConfig.Cert.ConfigMap != nil:
					configMapSelectors[prefix+ep.TLSConfig.Cert.ConfigMap.Name+"/"+ep.TLSConfig.Cert.ConfigMap.Key] = ep.TLSConfig.Cert.ConfigMap
				}
			}
			if ep.TLSConfig.KeySecret != nil {
				secretSelectors[prefix+ep.TLSConfig.KeySecret.Name+"/"+ep.TLSConfig.KeySecret.Key] = ep.TLSConfig.KeySecret
			}

			for key, selector := range secretSelectors {
				sClient := c.kclient.CoreV1().Secrets(mon.Namespace)
				asset, err := getCredFromSecret(
					sClient,
					*selector,
					"tls config",
					key,
					nsSecretCache,
				)
				if err != nil {
					return nil, fmt.Errorf(
						"failed to extract endpoint tls asset for servicemonitor %v from secret %v and key %v in namespace %v",
						mon.Name, selector.Name, selector.Key, mon.Namespace,
					)
				}

				assets[fmt.Sprintf(
					"%v_%v_%v",
					mon.Namespace,
					selector.Name,
					selector.Key,
				)] = TLSAsset(asset)
			}

			for key, selector := range configMapSelectors {
				sClient := c.kclient.CoreV1().ConfigMaps(mon.Namespace)
				asset, err := getCredFromConfigMap(
					sClient,
					*selector,
					"tls config",
					key,
					nsConfigMapCache,
				)
				if err != nil {
					return nil, fmt.Errorf(
						"failed to extract endpoint tls asset for servicemonitor %v from configmap %v and key %v in namespace %v",
						mon.Name, selector.Name, selector.Key, mon.Namespace,
					)
				}

				assets[fmt.Sprintf(
					"%v_%v_%v",
					mon.Namespace,
					selector.Name,
					selector.Key,
				)] = TLSAsset(asset)
			}

		}
	}

	return assets, nil
}

func (c *Operator) createOrUpdateConfigurationSecret(p *monitoringv1.Prometheus, ruleConfigMapNames []string) error {
	smons, err := c.selectServiceMonitors(p)
	if err != nil {
		return errors.Wrap(err, "selecting ServiceMonitors failed")
	}

	pmons, err := c.selectPodMonitors(p)
	if err != nil {
		return errors.Wrap(err, "selecting PodMonitors failed")
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

	bearerTokens, err := c.loadBearerTokensFromSecrets(smons)
	if err != nil {
		return err
	}

	additionalScrapeConfigs, err := c.loadAdditionalScrapeConfigsSecret(p.Spec.AdditionalScrapeConfigs, SecretsInPromNS)
	if err != nil {
		return errors.Wrap(err, "loading additional scrape configs from Secret failed")
	}
	additionalAlertRelabelConfigs, err := c.loadAdditionalScrapeConfigsSecret(p.Spec.AdditionalAlertRelabelConfigs, SecretsInPromNS)
	if err != nil {
		return errors.Wrap(err, "loading additional alert relabel configs from Secret failed")
	}
	additionalAlertManagerConfigs, err := c.loadAdditionalScrapeConfigsSecret(p.Spec.AdditionalAlertManagerConfigs, SecretsInPromNS)
	if err != nil {
		return errors.Wrap(err, "loading additional alert manager configs from Secret failed")
	}

	// Update secret based on the most recent configuration.
	conf, err := c.configGenerator.generateConfig(
		p,
		smons,
		pmons,
		basicAuthSecrets,
		bearerTokens,
		additionalScrapeConfigs,
		additionalAlertRelabelConfigs,
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

	// Compress config to avoid 1mb secret limit for a while
	var buf bytes.Buffer
	if err = gzipConfig(&buf, conf); err != nil {
		return errors.Wrap(err, "couldnt gzip config")
	}
	s.Data[configFilename] = buf.Bytes()

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

func (c *Operator) createOrUpdateTLSAssetSecret(p *monitoringv1.Prometheus) error {
	boolTrue := true
	sClient := c.kclient.CoreV1().Secrets(p.Namespace)

	smons, err := c.selectServiceMonitors(p)
	if err != nil {
		return errors.Wrap(err, "selecting ServiceMonitors failed")
	}

	tlsAssets, err := c.loadTLSAssets(smons)
	if err != nil {
		return err
	}

	tlsAssetsSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   tlsAssetsSecretName(p.Name),
			Labels: c.config.Labels.Merge(managedByOperatorLabels),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         p.APIVersion,
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               p.Kind,
					Name:               p.Name,
					UID:                p.UID,
				},
			},
		},
		Data: map[string][]byte{},
	}

	for key, asset := range tlsAssets {
		tlsAssetsSecret.Data[key] = []byte(asset)
	}

	_, err = sClient.Get(tlsAssetsSecret.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(
				err,
				"failed to check whether tls assets secret already exists for Prometheus %v in namespace %v",
				p.Name,
				p.Namespace,
			)
		}
		_, err = sClient.Create(tlsAssetsSecret)
		level.Debug(c.logger).Log("msg", "created tlsAssetsSecret", "secretname", tlsAssetsSecret.Name)

	} else {
		_, err = sClient.Update(tlsAssetsSecret)
		level.Debug(c.logger).Log("msg", "updated tlsAssetsSecret", "secretname", tlsAssetsSecret.Name)
	}

	if err != nil {
		return errors.Wrapf(err, "failed to create TLS assets secret for Prometheus %v in namespace %v", p.Name, p.Namespace)
	}

	return nil
}

func (c *Operator) selectServiceMonitors(p *monitoringv1.Prometheus) (map[string]*monitoringv1.ServiceMonitor, error) {
	namespaces := []string{}
	// Selectors (<namespace>/<name>) might overlap. Deduplicate them along the keyFunc.
	res := make(map[string]*monitoringv1.ServiceMonitor)

	servMonSelector, err := metav1.LabelSelectorAsSelector(p.Spec.ServiceMonitorSelector)
	if err != nil {
		return nil, err
	}

	// If 'ServiceMonitorNamespaceSelector' is nil only check own namespace.
	if p.Spec.ServiceMonitorNamespaceSelector == nil {
		namespaces = append(namespaces, p.Namespace)
	} else {
		servMonNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.ServiceMonitorNamespaceSelector)
		if err != nil {
			return nil, err
		}

		namespaces, err = c.listMatchingNamespaces(servMonNSSelector)
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

	// If denied by Prometheus spec, filter out all service monitors that access
	// the file system.
	if p.Spec.ArbitraryFSAccessThroughSMs.Deny {
		for namespaceAndName, sm := range res {
			for _, endpoint := range sm.Spec.Endpoints {
				if err := testForArbitraryFSAccess(endpoint); err != nil {
					delete(res, namespaceAndName)
					level.Warn(c.logger).Log(
						"msg", "skipping servicemonitor",
						"error", err.Error(),
						"servicemonitor", namespaceAndName,
						"namespace", p.Namespace,
						"prometheus", p.Name,
					)
				}
			}
		}
	}

	serviceMonitors := []string{}
	for k := range res {
		serviceMonitors = append(serviceMonitors, k)
	}
	level.Debug(c.logger).Log("msg", "selected ServiceMonitors", "servicemonitors", strings.Join(serviceMonitors, ","), "namespace", p.Namespace, "prometheus", p.Name)

	return res, nil
}

func (c *Operator) selectPodMonitors(p *monitoringv1.Prometheus) (map[string]*monitoringv1.PodMonitor, error) {
	namespaces := []string{}
	// Selectors might overlap. Deduplicate them along the keyFunc.
	res := make(map[string]*monitoringv1.PodMonitor)

	podMonSelector, err := metav1.LabelSelectorAsSelector(p.Spec.PodMonitorSelector)
	if err != nil {
		return nil, err
	}

	// If 'PodMonitorNamespaceSelector' is nil only check own namespace.
	if p.Spec.PodMonitorNamespaceSelector == nil {
		namespaces = append(namespaces, p.Namespace)
	} else {
		podMonNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.PodMonitorNamespaceSelector)
		if err != nil {
			return nil, err
		}

		namespaces, err = c.listMatchingNamespaces(podMonNSSelector)
		if err != nil {
			return nil, err
		}
	}

	level.Debug(c.logger).Log("msg", "filtering namespaces to select PodMonitors from", "namespaces", strings.Join(namespaces, ","), "namespace", p.Namespace, "prometheus", p.Name)

	podMonitors := []string{}
	for _, ns := range namespaces {
		cache.ListAllByNamespace(c.pmonInf.GetIndexer(), ns, podMonSelector, func(obj interface{}) {
			k, ok := c.keyFunc(obj)
			if ok {
				res[k] = obj.(*monitoringv1.PodMonitor)
				podMonitors = append(podMonitors, k)
			}
		})
	}

	level.Debug(c.logger).Log("msg", "selected PodMonitors", "podmonitors", strings.Join(podMonitors, ","), "namespace", p.Namespace, "prometheus", p.Name)

	return res, nil
}

func testForArbitraryFSAccess(e monitoringv1.Endpoint) error {
	if e.BearerTokenFile != "" {
		return errors.New("it accesses file system via bearer token file which Prometheus specification prohibits")
	}

	tlsConf := e.TLSConfig
	if tlsConf == nil {
		return nil
	}

	if err := e.TLSConfig.Validate(); err != nil {
		return err
	}

	if tlsConf.CAFile != "" || tlsConf.CertFile != "" || tlsConf.KeyFile != "" {
		return errors.New("it accesses file system via tls config which Prometheus specification prohibits")
	}

	return nil
}

// listMatchingNamespaces lists all the namespaces that match the provided
// selector.
func (c *Operator) listMatchingNamespaces(selector labels.Selector) ([]string, error) {
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
		k8sutil.NewCustomResourceDefinition(c.config.CrdKinds.Prometheus, monitoring.GroupName, c.config.Labels.LabelsMap, c.config.EnableValidation),
		k8sutil.NewCustomResourceDefinition(c.config.CrdKinds.ServiceMonitor, monitoring.GroupName, c.config.Labels.LabelsMap, c.config.EnableValidation),
		k8sutil.NewCustomResourceDefinition(c.config.CrdKinds.PodMonitor, monitoring.GroupName, c.config.Labels.LabelsMap, c.config.EnableValidation),
		k8sutil.NewCustomResourceDefinition(c.config.CrdKinds.PrometheusRule, monitoring.GroupName, c.config.Labels.LabelsMap, c.config.EnableValidation),
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
		{
			monitoringv1.PrometheusesKind,
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.PrometheusAllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						return c.mclient.MonitoringV1().Prometheuses(namespace).List(options)
					},
				}
			}).List,
		},
		{
			monitoringv1.ServiceMonitorsKind,
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.AllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						return c.mclient.MonitoringV1().ServiceMonitors(namespace).List(options)
					},
				}
			}).List,
		},
		{
			monitoringv1.PodMonitorsKind,
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.AllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						return c.mclient.MonitoringV1().PodMonitors(namespace).List(options)
					},
				}
			}).List,
		},
		{
			monitoringv1.PrometheusRuleKind,
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.AllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						return c.mclient.MonitoringV1().PrometheusRules(namespace).List(options)
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
