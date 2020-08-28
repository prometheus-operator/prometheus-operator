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
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/listwatch"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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
	kclient kubernetes.Interface
	mclient monitoringclient.Interface
	logger  log.Logger

	promInf   cache.SharedIndexInformer
	smonInf   cache.SharedIndexInformer
	pmonInf   cache.SharedIndexInformer
	probeInf  cache.SharedIndexInformer
	ruleInf   cache.SharedIndexInformer
	cmapInf   cache.SharedIndexInformer
	secrInf   cache.SharedIndexInformer
	ssetInf   cache.SharedIndexInformer
	nsPromInf cache.SharedIndexInformer
	nsMonInf  cache.SharedIndexInformer

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
	ClusterDomain                 string
	KubeletObject                 string
	ListenAddress                 string
	TLSInsecure                   bool
	TLSConfig                     rest.TLSClientConfig
	ServerTLSConfig               operator.TLSServerConfig
	ConfigReloaderImage           string
	ConfigReloaderCPU             string
	ConfigReloaderMemory          string
	PrometheusConfigReloaderImage string
	AlertmanagerDefaultBaseImage  string
	PrometheusDefaultBaseImage    string
	ThanosDefaultBaseImage        string
	Namespaces                    Namespaces
	Labels                        Labels
	LocalHost                     string
	LogLevel                      string
	LogFormat                     string
	PromSelector                  string
	AlertManagerSelector          string
	ThanosRulerSelector           string
	SecretListWatchSelector       string
}

type Namespaces struct {
	// allow list/deny list for common custom resources
	AllowList, DenyList map[string]struct{}
	// allow list for prometheus/alertmanager custom resources
	PrometheusAllowList, AlertmanagerAllowList, ThanosRulerAllowList map[string]struct{}
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

	secretListWatchSelector, err := fields.ParseSelector(conf.SecretListWatchSelector)
	if err != nil {
		return nil, errors.Wrap(err, "can not parse secrets selector value")
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
						return mclient.MonitoringV1().Prometheuses(namespace).List(context.TODO(), options)
					},
					WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
						options.LabelSelector = c.config.PromSelector
						return mclient.MonitoringV1().Prometheuses(namespace).Watch(context.TODO(), options)
					},
				}
			}),
		),
		&monitoringv1.Prometheus{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	c.metrics.MustRegister(NewPrometheusCollector(c.promInf.GetStore()))
	c.metrics.MustRegister(operator.NewStoreCollector("prometheus", c.promInf.GetStore()))

	c.smonInf = cache.NewSharedIndexInformer(
		c.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.AllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						return mclient.MonitoringV1().ServiceMonitors(namespace).List(context.TODO(), options)
					},
					WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
						return mclient.MonitoringV1().ServiceMonitors(namespace).Watch(context.TODO(), options)
					},
				}
			}),
		),
		&monitoringv1.ServiceMonitor{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	c.metrics.MustRegister(operator.NewStoreCollector("servicemonitor", c.smonInf.GetStore()))

	c.pmonInf = cache.NewSharedIndexInformer(
		c.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.AllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						return mclient.MonitoringV1().PodMonitors(namespace).List(context.TODO(), options)
					},
					WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
						return mclient.MonitoringV1().PodMonitors(namespace).Watch(context.TODO(), options)
					},
				}
			}),
		),
		&monitoringv1.PodMonitor{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	c.metrics.MustRegister(operator.NewStoreCollector("podmonitor", c.pmonInf.GetStore()))

	c.probeInf = cache.NewSharedIndexInformer(
		c.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.AllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (object runtime.Object, err error) {
						return mclient.MonitoringV1().Probes(namespace).List(context.TODO(), options)
					},
					WatchFunc: func(options metav1.ListOptions) (w watch.Interface, err error) {
						return mclient.MonitoringV1().Probes(namespace).Watch(context.TODO(), options)
					},
				}
			}),
		),
		&monitoringv1.Probe{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	c.metrics.MustRegister(operator.NewStoreCollector("probe", c.probeInf.GetStore()))

	c.ruleInf = cache.NewSharedIndexInformer(
		c.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.AllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						return mclient.MonitoringV1().PrometheusRules(namespace).List(context.TODO(), options)
					},
					WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
						return mclient.MonitoringV1().PrometheusRules(namespace).Watch(context.TODO(), options)
					},
				}
			}),
		),
		&monitoringv1.PrometheusRule{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	c.metrics.MustRegister(operator.NewStoreCollector("prometheurule", c.ruleInf.GetStore()))

	c.cmapInf = cache.NewSharedIndexInformer(
		c.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.PrometheusAllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return &cache.ListWatch{
					ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
						options.LabelSelector = labelPrometheusName
						return c.kclient.CoreV1().ConfigMaps(namespace).List(context.TODO(), options)
					},
					WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
						options.LabelSelector = labelPrometheusName
						return c.kclient.CoreV1().ConfigMaps(namespace).Watch(context.TODO(), options)
					},
				}
			}),
		),
		&v1.ConfigMap{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	c.secrInf = cache.NewSharedIndexInformer(
		c.metrics.NewInstrumentedListerWatcher(
			listwatch.MultiNamespaceListerWatcher(c.logger, c.config.Namespaces.PrometheusAllowList, c.config.Namespaces.DenyList, func(namespace string) cache.ListerWatcher {
				return cache.NewListWatchFromClient(c.kclient.CoreV1().RESTClient(), "secrets", namespace, secretListWatchSelector)
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
	c.nsMonInf = newNamespaceInformer(c, c.config.Namespaces.AllowList)
	if listwatch.IdenticalNamespaces(c.config.Namespaces.AllowList, c.config.Namespaces.PrometheusAllowList) {
		c.nsPromInf = c.nsMonInf
	} else {
		c.nsPromInf = newNamespaceInformer(c, c.config.Namespaces.PrometheusAllowList)
	}

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
		{"Probe", c.probeInf},
		{"PrometheusRule", c.ruleInf},
		{"ConfigMap", c.cmapInf},
		{"Secret", c.secrInf},
		{"StatefulSet", c.ssetInf},
		{"PromNamespace", c.nsPromInf},
		{"MonNamespace", c.nsMonInf},
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
	c.probeInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleBmonAdd,
		UpdateFunc: c.handleBmonUpdate,
		DeleteFunc: c.handleBmonDelete,
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
	go c.probeInf.Run(stopc)
	go c.ruleInf.Run(stopc)
	go c.cmapInf.Run(stopc)
	go c.secrInf.Run(stopc)
	go c.ssetInf.Run(stopc)
	go c.nsMonInf.Run(stopc)
	if c.nsPromInf != c.nsMonInf {
		go c.nsPromInf.Run(stopc)
	}
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

	nodes, err := c.kclient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
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

		c.enqueueForMonitorNamespace(o.GetNamespace())
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

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleSmonDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ServiceMonitor delete")
		c.metrics.TriggerByCounter(monitoringv1.ServiceMonitorsKind, "delete").Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handlePmonAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PodMonitor added")
		c.metrics.TriggerByCounter(monitoringv1.PodMonitorsKind, "add").Inc()
		c.enqueueForMonitorNamespace(o.GetNamespace())
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

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handlePmonDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PodMonitor delete")
		c.metrics.TriggerByCounter(monitoringv1.PodMonitorsKind, "delete").Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleBmonAdd(obj interface{}) {
	if o, ok := c.getObject(obj); ok {
		level.Debug(c.logger).Log("msg", "Probe added")
		c.metrics.TriggerByCounter(monitoringv1.ProbesKind, "add").Inc()
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleBmonUpdate(old, cur interface{}) {
	if old.(*monitoringv1.Probe).ResourceVersion == cur.(*monitoringv1.Probe).ResourceVersion {
		return
	}

	if o, ok := c.getObject(cur); ok {
		level.Debug(c.logger).Log("msg", "Probe updated")
		c.metrics.TriggerByCounter(monitoringv1.ProbesKind, "update")
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleBmonDelete(obj interface{}) {
	if o, ok := c.getObject(obj); ok {
		level.Debug(c.logger).Log("msg", "Probe delete")
		c.metrics.TriggerByCounter(monitoringv1.ProbesKind, "delete").Inc()
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleRuleAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PrometheusRule added")
		c.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, "add").Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
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

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enque just for the namespace
func (c *Operator) handleRuleDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PrometheusRule deleted")
		c.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, "delete").Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Do we need to enque secrets just for the namespace or in general?
func (c *Operator) handleSecretDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "Secret deleted")
		c.metrics.TriggerByCounter("Secret", "delete").Inc()

		c.enqueueForPrometheusNamespace(o.GetNamespace())
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

		c.enqueueForPrometheusNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleSecretAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "Secret added")
		c.metrics.TriggerByCounter("Secret", "add").Inc()

		c.enqueueForPrometheusNamespace(o.GetNamespace())
	}
}

// TODO: Do we need to enque configmaps just for the namespace or in general?
func (c *Operator) handleConfigMapAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ConfigMap added")
		c.metrics.TriggerByCounter("ConfigMap", "add").Inc()

		c.enqueueForPrometheusNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleConfigMapDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ConfigMap deleted")
		c.metrics.TriggerByCounter("ConfigMap", "delete").Inc()

		c.enqueueForPrometheusNamespace(o.GetNamespace())
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

		c.enqueueForPrometheusNamespace(o.GetNamespace())
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

func (c *Operator) enqueueForPrometheusNamespace(nsName string) {
	c.enqueueForNamespace(c.nsPromInf.GetStore(), nsName)
}

func (c *Operator) enqueueForMonitorNamespace(nsName string) {
	c.enqueueForNamespace(c.nsMonInf.GetStore(), nsName)
}

// enqueueForNamespace enqueues all Prometheus object keys that belong to the
// given namespace or select objects in the given namespace.
func (c *Operator) enqueueForNamespace(store cache.Store, nsName string) {
	nsObject, exists, err := store.GetByKey(nsName)
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
		// Check for Prometheus instances in the namespace.
		p := obj.(*monitoringv1.Prometheus)
		if p.Namespace == nsName {
			c.enqueue(p)
			return
		}

		// Check for Prometheus instances selecting ServiceMonitors in
		// the namespace.
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

		// Check for Prometheus instances selecting PodMonitors in the NS.
		pmNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.PodMonitorNamespaceSelector)
		if err != nil {
			level.Error(c.logger).Log(
				"msg", fmt.Sprintf("failed to convert PodMonitorNamespaceSelector of %q to selector", p.Name),
				"err", err,
			)
			return
		}

		if pmNSSelector.Matches(labels.Set(ns.Labels)) {
			c.enqueue(p)
			return
		}

		// Check for Prometheus instances selecting Probes in the NS.
		bmNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.ProbeNamespaceSelector)
		if err != nil {
			level.Error(c.logger).Log(
				"msg", fmt.Sprintf("failed to convert ProbeNamespaceSelector of %q to selector", p.Name),
				"err", err,
			)
			return
		}

		if bmNSSelector.Matches(labels.Set(ns.Labels)) {
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

	c.metrics.ReconcileCounter().Inc()
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

	assetStore := newAssetStore(c.kclient.CoreV1(), c.kclient.CoreV1())

	if err := c.createOrUpdateConfigurationSecret(p, ruleConfigMapNames, assetStore); err != nil {
		return errors.Wrap(err, "creating config failed")
	}

	if err := c.createOrUpdateTLSAssetSecret(p, assetStore); err != nil {
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
	operator.SanitizeSTS(sset)

	if !exists {
		level.Debug(c.logger).Log("msg", "no current Prometheus statefulset found")
		level.Debug(c.logger).Log("msg", "creating Prometheus statefulset")
		if _, err := ssetClient.Create(context.TODO(), sset, metav1.CreateOptions{}); err != nil {
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

	_, err = ssetClient.Update(context.TODO(), sset, metav1.UpdateOptions{})
	sErr, ok := err.(*apierrors.StatusError)

	if ok && sErr.ErrStatus.Code == 422 && sErr.ErrStatus.Reason == metav1.StatusReasonInvalid {
		c.metrics.StsDeleteCreateCounter().Inc()
		level.Info(c.logger).Log("msg", "resolving illegal update of Prometheus StatefulSet", "details", sErr.ErrStatus.Details)
		propagationPolicy := metav1.DeletePropagationForeground
		if err := ssetClient.Delete(context.TODO(), sset.GetName(), metav1.DeleteOptions{PropagationPolicy: &propagationPolicy}); err != nil {
			return errors.Wrap(err, "failed to delete StatefulSet to avoid forbidden action")
		}
		return nil
	}

	if err != nil {
		return errors.Wrap(err, "updating StatefulSet failed")
	}

	return nil
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

	if p.Spec.ServiceMonitorSelector == nil && p.Spec.PodMonitorSelector == nil && p.Spec.ProbeSelector == nil {
		level.Warn(logger).Log("msg", "neither serviceMonitorSelector nor podMonitorSelector, nor probeSelector specified. Custom configuration is deprecated, use additionalScrapeConfigs instead")
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

	pods, err := kclient.CoreV1().Pods(p.Namespace).List(context.TODO(), ListOptions(p.Name))
	if err != nil {
		return nil, nil, errors.Wrap(err, "retrieving pods of failed")
	}
	sset, err := kclient.AppsV1().StatefulSets(p.Namespace).Get(context.TODO(), statefulSetNameFromPrometheusName(p.Name), metav1.GetOptions{})
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

func gzipConfig(buf *bytes.Buffer, conf []byte) error {
	w := gzip.NewWriter(buf)
	defer w.Close()
	if _, err := w.Write(conf); err != nil {
		return err
	}
	return nil
}

func (c *Operator) createOrUpdateConfigurationSecret(p *monitoringv1.Prometheus, ruleConfigMapNames []string, store *assetStore) error {
	// If no service or pod monitor selectors are configured, the user wants to
	// manage configuration themselves. Do create an empty Secret if it doesn't
	// exist.
	if p.Spec.ServiceMonitorSelector == nil && p.Spec.PodMonitorSelector == nil &&
		p.Spec.ProbeSelector == nil {
		level.Debug(c.logger).Log("msg", "neither ServiceMonitor nor PodMonitor, nor Probe selector specified, leaving configuration unmanaged", "prometheus", p.Name, "namespace", p.Namespace)

		s, err := makeEmptyConfigurationSecret(p, c.config)
		if err != nil {
			return errors.Wrap(err, "generating empty config secret failed")
		}
		sClient := c.kclient.CoreV1().Secrets(p.Namespace)
		_, err = sClient.Get(context.TODO(), s.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			if _, err := c.kclient.CoreV1().Secrets(p.Namespace).Create(context.TODO(), s, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
				return errors.Wrap(err, "creating empty config file failed")
			}
		}
		if !apierrors.IsNotFound(err) && err != nil {
			return err
		}

		return nil
	}

	smons, err := c.selectServiceMonitors(p, store)
	if err != nil {
		return errors.Wrap(err, "selecting ServiceMonitors failed")
	}

	pmons, err := c.selectPodMonitors(p)
	if err != nil {
		return errors.Wrap(err, "selecting PodMonitors failed")
	}

	bmons, err := c.selectProbes(p)
	if err != nil {
		return errors.Wrap(err, "selecting Probes failed")
	}

	sClient := c.kclient.CoreV1().Secrets(p.Namespace)
	SecretsInPromNS, err := sClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for i, remote := range p.Spec.RemoteRead {
		if err := store.addBasicAuth(p.GetNamespace(), remote.BasicAuth, fmt.Sprintf("remoteRead/%d", i)); err != nil {
			return errors.Wrapf(err, "remote read %d", i)
		}
	}

	for i, remote := range p.Spec.RemoteWrite {
		if err := store.addBasicAuth(p.GetNamespace(), remote.BasicAuth, fmt.Sprintf("remoteWrite/%d", i)); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
	}

	if p.Spec.APIServerConfig != nil {
		if err := store.addBasicAuth(p.GetNamespace(), p.Spec.APIServerConfig.BasicAuth, "apiserver"); err != nil {
			return errors.Wrap(err, "apiserver config")
		}
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
		bmons,
		store.basicAuthAssets,
		store.bearerTokenAssets,
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

	curSecret, err := sClient.Get(context.TODO(), s.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		level.Debug(c.logger).Log("msg", "creating configuration")
		_, err = sClient.Create(context.TODO(), s, metav1.CreateOptions{})
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
	_, err = sClient.Update(context.TODO(), s, metav1.UpdateOptions{})
	return err
}

func (c *Operator) createOrUpdateTLSAssetSecret(p *monitoringv1.Prometheus, store *assetStore) error {
	boolTrue := true
	sClient := c.kclient.CoreV1().Secrets(p.Namespace)

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

	for i, rw := range p.Spec.RemoteWrite {
		if err := store.addTLSConfig(p.GetNamespace(), rw.TLSConfig); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
	}

	for key, asset := range store.tlsAssets {
		tlsAssetsSecret.Data[key.String()] = []byte(asset)
	}

	_, err := sClient.Get(context.TODO(), tlsAssetsSecret.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(
				err,
				"failed to check whether tls assets secret already exists for Prometheus %v in namespace %v",
				p.Name,
				p.Namespace,
			)
		}
		_, err = sClient.Create(context.TODO(), tlsAssetsSecret, metav1.CreateOptions{})
		level.Debug(c.logger).Log("msg", "created tlsAssetsSecret", "secretname", tlsAssetsSecret.Name)

	} else {
		_, err = sClient.Update(context.TODO(), tlsAssetsSecret, metav1.UpdateOptions{})
		level.Debug(c.logger).Log("msg", "updated tlsAssetsSecret", "secretname", tlsAssetsSecret.Name)
	}

	if err != nil {
		return errors.Wrapf(err, "failed to create TLS assets secret for Prometheus %v in namespace %v", p.Name, p.Namespace)
	}

	return nil
}

func (c *Operator) selectServiceMonitors(p *monitoringv1.Prometheus, store *assetStore) (map[string]*monitoringv1.ServiceMonitor, error) {
	namespaces := []string{}
	// Selectors (<namespace>/<name>) might overlap. Deduplicate them along the keyFunc.
	serviceMonitors := make(map[string]*monitoringv1.ServiceMonitor)

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
				serviceMonitors[k] = obj.(*monitoringv1.ServiceMonitor)
			}
		})
	}

	res := make(map[string]*monitoringv1.ServiceMonitor, len(serviceMonitors))
	for namespaceAndName, sm := range serviceMonitors {
		var err error

		for i, endpoint := range sm.Spec.Endpoints {
			// If denied by Prometheus spec, filter out all service monitors that access
			// the file system.
			if p.Spec.ArbitraryFSAccessThroughSMs.Deny {
				if err = testForArbitraryFSAccess(endpoint); err != nil {
					break
				}
			}

			smKey := fmt.Sprintf("serviceMonitor/%s/%s/%d", sm.GetNamespace(), sm.GetName(), i)

			if err = store.addBearerToken(sm.GetNamespace(), endpoint.BearerTokenSecret, smKey); err != nil {
				break
			}

			if err = store.addBasicAuth(sm.GetNamespace(), endpoint.BasicAuth, smKey); err != nil {
				break
			}

			if err = store.addTLSConfig(sm.GetNamespace(), endpoint.TLSConfig); err != nil {
				break
			}
		}

		if err != nil {
			level.Warn(c.logger).Log(
				"msg", "skipping servicemonitor",
				"error", err.Error(),
				"servicemonitor", namespaceAndName,
				"namespace", p.Namespace,
				"prometheus", p.Name,
			)
			continue
		}

		res[namespaceAndName] = sm
	}

	smKeys := []string{}
	for k := range res {
		smKeys = append(smKeys, k)
	}
	level.Debug(c.logger).Log("msg", "selected ServiceMonitors", "servicemonitors", strings.Join(smKeys, ","), "namespace", p.Namespace, "prometheus", p.Name)

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

func (c *Operator) selectProbes(p *monitoringv1.Prometheus) (map[string]*monitoringv1.Probe, error) {
	namespaces := []string{}
	// Selectors might overlap. Deduplicate them along the keyFunc.
	res := make(map[string]*monitoringv1.Probe)

	bMonSelector, err := metav1.LabelSelectorAsSelector(p.Spec.ProbeSelector)
	if err != nil {
		return nil, err
	}

	// If 'ProbeNamespaceSelector' is nil only check own namespace.
	if p.Spec.ProbeNamespaceSelector == nil {
		namespaces = append(namespaces, p.Namespace)
	} else {
		bMonNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.ProbeNamespaceSelector)
		if err != nil {
			return nil, err
		}

		namespaces, err = c.listMatchingNamespaces(bMonNSSelector)
		if err != nil {
			return nil, err
		}
	}

	level.Debug(c.logger).Log("msg", "filtering namespaces to select Probes from", "namespaces", strings.Join(namespaces, ","), "namespace", p.Namespace, "prometheus", p.Name)

	probes := make([]string, 0)
	for _, ns := range namespaces {
		cache.ListAllByNamespace(c.probeInf.GetIndexer(), ns, bMonSelector, func(obj interface{}) {
			if k, ok := c.keyFunc(obj); ok {
				res[k] = obj.(*monitoringv1.Probe)
				probes = append(probes, k)
			}
		})
	}

	level.Debug(c.logger).Log("msg", "selected Probes", "probes", strings.Join(probes, ","), "namespace", p.Namespace, "prometheus", p.Name)

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
	err := cache.ListAll(c.nsMonInf.GetStore(), selector, func(obj interface{}) {
		ns = append(ns, obj.(*v1.Namespace).Name)
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list namespaces")
	}
	return ns, nil
}

type tlsAssetKey struct {
	from string
	ns   string
	name string
	key  string
}

func tlsAssetKeyFromSecretSelector(ns string, sel *v1.SecretKeySelector) tlsAssetKey {
	return tlsAssetKey{
		from: "configmap",
		ns:   ns,
		name: sel.Name,
		key:  sel.Key,
	}
}

func tlsAssetKeyFromConfigMapSelector(ns string, sel *v1.ConfigMapKeySelector) tlsAssetKey {
	return tlsAssetKey{
		from: "secret",
		ns:   ns,
		name: sel.Name,
		key:  sel.Key,
	}
}

func (k tlsAssetKey) String() string {
	return fmt.Sprintf("%s_%s_%s_%s", k.from, k.ns, k.name, k.key)
}

type assetStore struct {
	cmClient corev1client.ConfigMapsGetter
	sClient  corev1client.SecretsGetter
	objStore cache.Store

	tlsAssets         map[tlsAssetKey]TLSAsset
	bearerTokenAssets map[string]BearerToken
	basicAuthAssets   map[string]BasicAuthCredentials
}

func newAssetStore(cmClient corev1client.ConfigMapsGetter, sClient corev1client.SecretsGetter) *assetStore {
	assetStore := &assetStore{
		cmClient:          cmClient,
		sClient:           sClient,
		tlsAssets:         make(map[tlsAssetKey]TLSAsset),
		bearerTokenAssets: make(map[string]BearerToken),
		basicAuthAssets:   make(map[string]BasicAuthCredentials),
	}

	assetStore.objStore = cache.NewStore(assetStore.keyFunc)

	return assetStore
}

func (a *assetStore) keyFunc(obj interface{}) (string, error) {
	switch v := obj.(type) {
	case *v1.ConfigMap:
		return fmt.Sprintf("0/%s/%s", v.GetNamespace(), v.GetName()), nil
	case *v1.Secret:
		return fmt.Sprintf("1/%s/%s", v.GetNamespace(), v.GetName()), nil
	}
	return "", errors.Errorf("unsupported type: %T", obj)
}

func (a *assetStore) addTLSConfig(ns string, tlsConfig *monitoringv1.TLSConfig) error {
	if tlsConfig == nil {
		return nil
	}

	if tlsConfig.CA != (monitoringv1.SecretOrConfigMap{}) {
		switch {
		case tlsConfig.CA.Secret != nil:
			ca, err := a.getSecretKey(ns, *tlsConfig.CA.Secret)
			if err != nil {
				return errors.Wrap(err, "failed to get CA")
			}
			a.tlsAssets[tlsAssetKeyFromSecretSelector(ns, tlsConfig.CA.Secret)] = TLSAsset(ca)

		case tlsConfig.CA.ConfigMap != nil:
			ca, err := a.getConfigMapKey(ns, *tlsConfig.CA.ConfigMap)
			if err != nil {
				return errors.Wrap(err, "failed to get CA")
			}
			a.tlsAssets[tlsAssetKeyFromConfigMapSelector(ns, tlsConfig.CA.ConfigMap)] = TLSAsset(ca)
		}
	}

	if tlsConfig.Cert != (monitoringv1.SecretOrConfigMap{}) {
		switch {
		case tlsConfig.Cert.Secret != nil:
			cert, err := a.getSecretKey(ns, *tlsConfig.Cert.Secret)
			if err != nil {
				return errors.Wrap(err, "failed to get cert")
			}
			a.tlsAssets[tlsAssetKeyFromSecretSelector(ns, tlsConfig.Cert.Secret)] = TLSAsset(cert)

		case tlsConfig.Cert.ConfigMap != nil:
			cert, err := a.getConfigMapKey(ns, *tlsConfig.Cert.ConfigMap)
			if err != nil {
				return errors.Wrap(err, "failed to get cert")
			}
			a.tlsAssets[tlsAssetKeyFromConfigMapSelector(ns, tlsConfig.Cert.ConfigMap)] = TLSAsset(cert)
		}
	}

	if tlsConfig.KeySecret != nil {
		key, err := a.getSecretKey(ns, *tlsConfig.KeySecret)
		if err != nil {
			return errors.Wrap(err, "failed to get key")
		}
		a.tlsAssets[tlsAssetKeyFromSecretSelector(ns, tlsConfig.KeySecret)] = TLSAsset(key)
	}

	return nil
}

func (a *assetStore) addBasicAuth(ns string, ba *monitoringv1.BasicAuth, key string) error {
	if ba == nil {
		return nil
	}

	username, err := a.getSecretKey(ns, ba.Username)
	if err != nil {
		return errors.Wrap(err, "failed to get basic auth username")
	}

	password, err := a.getSecretKey(ns, ba.Password)
	if err != nil {
		return errors.Wrap(err, "failed to get basic auth password")
	}

	a.basicAuthAssets[key] = BasicAuthCredentials{
		username: username,
		password: password,
	}

	return nil
}

func (a *assetStore) addBearerToken(ns string, sel v1.SecretKeySelector, key string) error {
	if sel.Name == "" {
		return nil
	}

	bearerToken, err := a.getSecretKey(ns, sel)
	if err != nil {
		return errors.Wrap(err, "failed to get bearer token")
	}

	a.bearerTokenAssets[key] = BearerToken(bearerToken)

	return nil
}

func (a *assetStore) getConfigMapKey(namespace string, sel v1.ConfigMapKeySelector) (string, error) {
	obj, exists, err := a.objStore.Get(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sel.Name,
			Namespace: namespace,
		},
	})
	if err != nil {
		return "", errors.Wrapf(err, "unexpected store error when getting configmap %q", sel.Name)
	}

	if !exists {
		cm, err := a.cmClient.ConfigMaps(namespace).Get(context.TODO(), sel.Name, metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrapf(err, "unable to get configmap %q", sel.Name)
		}
		if err = a.objStore.Add(cm); err != nil {
			return "", errors.Wrapf(err, "unexpected store error when adding configmap %q", sel.Name)
		}
		obj = cm
	}

	cm := obj.(*v1.ConfigMap)
	if _, found := cm.Data[sel.Key]; !found {
		return "", errors.Errorf("key %q in configmap %q not found", sel.Key, sel.Name)
	}

	return cm.Data[sel.Key], nil
}

func (a *assetStore) getSecretKey(namespace string, sel v1.SecretKeySelector) (string, error) {
	obj, exists, err := a.objStore.Get(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sel.Name,
			Namespace: namespace,
		},
	})
	if err != nil {
		return "", errors.Wrapf(err, "unexpected store error when getting secret %q", sel.Name)
	}

	if !exists {
		secret, err := a.sClient.Secrets(namespace).Get(context.TODO(), sel.Name, metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrapf(err, "unable to get secret %q", sel.Name)
		}
		if err = a.objStore.Add(secret); err != nil {
			return "", errors.Wrapf(err, "unexpected store error when adding secret %q", sel.Name)
		}
		obj = secret
	}

	secret := obj.(*v1.Secret)
	if _, found := secret.Data[sel.Key]; !found {
		return "", errors.Errorf("key %q in secret %q not found", sel.Key, sel.Name)
	}

	return string(secret.Data[sel.Key]), nil
}
