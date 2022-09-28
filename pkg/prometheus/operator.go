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
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
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

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/listwatch"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
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

	nsPromInf cache.SharedIndexInformer
	nsMonInf  cache.SharedIndexInformer

	promInfs  *informers.ForResource
	smonInfs  *informers.ForResource
	pmonInfs  *informers.ForResource
	probeInfs *informers.ForResource
	ruleInfs  *informers.ForResource
	cmapInfs  *informers.ForResource
	secrInfs  *informers.ForResource
	ssetInfs  *informers.ForResource

	// Queue to trigger reconciliations of Prometheus objects.
	reconcileQueue workqueue.RateLimitingInterface
	// Queue to trigger status updates of Prometheus objects.
	statusQueue workqueue.RateLimitingInterface

	metrics         *operator.Metrics
	reconciliations *operator.ReconciliationTracker

	nodeAddressLookupErrors prometheus.Counter
	nodeEndpointSyncs       prometheus.Counter
	nodeEndpointSyncErrors  prometheus.Counter

	host                   string
	kubeletObjectName      string
	kubeletObjectNamespace string
	kubeletSyncEnabled     bool
	config                 operator.Config
	endpointSliceSupported bool
}

// New creates a new controller.
func New(ctx context.Context, conf operator.Config, logger log.Logger, r prometheus.Registerer) (*Operator, error) {
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
		reconcileQueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "prometheus"),
		statusQueue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "prometheus_status"),
		host:                   cfg.Host,
		kubeletObjectName:      kubeletObjectName,
		kubeletObjectNamespace: kubeletObjectNamespace,
		kubeletSyncEnabled:     kubeletSyncEnabled,
		config:                 conf,
		metrics:                operator.NewMetrics("prometheus", r),
		reconciliations:        &operator.ReconciliationTracker{},
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
	c.metrics.MustRegister(
		c.nodeAddressLookupErrors,
		c.nodeEndpointSyncs,
		c.nodeEndpointSyncErrors,
		c.reconciliations,
	)

	c.promInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.config.Namespaces.PrometheusAllowList,
			c.config.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = c.config.PromSelector
			},
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PrometheusName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating prometheus informers")
	}

	var promStores []cache.Store
	for _, informer := range c.promInfs.GetInformers() {
		promStores = append(promStores, informer.Informer().GetStore())
	}
	c.metrics.MustRegister(newPrometheusCollectorForStores(promStores...))

	c.smonInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.config.Namespaces.AllowList,
			c.config.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ServiceMonitorName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating servicemonitor informers")
	}

	c.pmonInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.config.Namespaces.AllowList,
			c.config.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PodMonitorName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating podmonitor informers")
	}

	c.probeInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.config.Namespaces.AllowList,
			c.config.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ProbeName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating probe informers")
	}

	c.ruleInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.config.Namespaces.AllowList,
			c.config.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PrometheusRuleName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating prometheusrule informers")
	}

	c.cmapInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			c.config.Namespaces.PrometheusAllowList,
			c.config.Namespaces.DenyList,
			c.kclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = labelPrometheusName
			},
		),
		v1.SchemeGroupVersion.WithResource(string(v1.ResourceConfigMaps)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating configmap informers")
	}

	c.secrInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			c.config.Namespaces.PrometheusAllowList,
			c.config.Namespaces.DenyList,
			c.kclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.FieldSelector = secretListWatchSelector.String()
			},
		),
		v1.SchemeGroupVersion.WithResource(string(v1.ResourceSecrets)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating secrets informers")
	}

	c.ssetInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			c.config.Namespaces.PrometheusAllowList,
			c.config.Namespaces.DenyList,
			c.kclient,
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
				listwatch.NewUnprivilegedNamespaceListWatchFromClient(ctx, o.logger, o.kclient.CoreV1().RESTClient(), allowList, o.config.Namespaces.DenyList, fields.Everything()),
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

	endpointSliceSupported, err := k8sutil.IsAPIGroupVersionResourceSupported(c.kclient.Discovery(), "discovery.k8s.io", "endpointslices")
	if err != nil {
		level.Warn(c.logger).Log("msg", "failed to check if the API supports the endpointslice resources", "err ", err)
	}
	level.Info(c.logger).Log("msg", "Kubernetes API capabilities", "endpointslices", endpointSliceSupported)
	c.endpointSliceSupported = endpointSliceSupported
	return c, nil
}

// waitForCacheSync waits for the informers' caches to be synced.
func (c *Operator) waitForCacheSync(ctx context.Context) error {
	for _, infs := range []struct {
		name                 string
		informersForResource *informers.ForResource
	}{
		{"Prometheus", c.promInfs},
		{"ServiceMonitor", c.smonInfs},
		{"PodMonitor", c.pmonInfs},
		{"PrometheusRule", c.ruleInfs},
		{"Probe", c.probeInfs},
		{"ConfigMap", c.cmapInfs},
		{"Secret", c.secrInfs},
		{"StatefulSet", c.ssetInfs},
	} {
		for _, inf := range infs.informersForResource.GetInformers() {
			if !operator.WaitForNamedCacheSync(ctx, "prometheus", log.With(c.logger, "informer", infs.name), inf.Informer()) {
				return errors.Errorf("failed to sync cache for %s informer", infs.name)
			}
		}
	}

	for _, inf := range []struct {
		name     string
		informer cache.SharedIndexInformer
	}{
		{"PromNamespace", c.nsPromInf},
		{"MonNamespace", c.nsMonInf},
	} {
		if !operator.WaitForNamedCacheSync(ctx, "prometheus", log.With(c.logger, "informer", inf.name), inf.informer) {
			return errors.Errorf("failed to sync cache for %s informer", inf.name)
		}
	}

	level.Info(c.logger).Log("msg", "successfully synced all caches")
	return nil
}

// addHandlers adds the eventhandlers to the informers.
func (c *Operator) addHandlers() {
	c.promInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handlePrometheusAdd,
		DeleteFunc: c.handlePrometheusDelete,
		UpdateFunc: c.handlePrometheusUpdate,
	})

	c.smonInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleSmonAdd,
		DeleteFunc: c.handleSmonDelete,
		UpdateFunc: c.handleSmonUpdate,
	})

	c.pmonInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handlePmonAdd,
		DeleteFunc: c.handlePmonDelete,
		UpdateFunc: c.handlePmonUpdate,
	})
	c.probeInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleBmonAdd,
		UpdateFunc: c.handleBmonUpdate,
		DeleteFunc: c.handleBmonDelete,
	})
	c.ruleInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleRuleAdd,
		DeleteFunc: c.handleRuleDelete,
		UpdateFunc: c.handleRuleUpdate,
	})
	c.cmapInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleConfigMapAdd,
		DeleteFunc: c.handleConfigMapDelete,
		UpdateFunc: c.handleConfigMapUpdate,
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

	// The controller needs to watch the namespaces in which the service/pod
	// monitors and rules live because a label change on a namespace may
	// trigger a configuration change.
	// It doesn't need to watch on addition/deletion though because it's
	// already covered by the event handlers on service/pod monitors and rules.
	c.nsMonInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: c.handleMonitorNamespaceUpdate,
	})
}

// Run the controller.
func (c *Operator) Run(ctx context.Context) error {
	defer c.reconcileQueue.ShutDown()
	defer c.statusQueue.ShutDown()

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

	// Start the goroutine that reconciles the desired state of Prometheus objects.
	go func(ctx context.Context) {
		for c.processNextReconcileItem(ctx) {
		}
	}(ctx)

	// Start the goroutine that reconciles the status of Prometheus objects.
	go func(ctx context.Context) {
		for c.processNextStatusItem(ctx) {
		}
	}(ctx)

	go c.promInfs.Start(ctx.Done())
	go c.smonInfs.Start(ctx.Done())
	go c.pmonInfs.Start(ctx.Done())
	go c.probeInfs.Start(ctx.Done())
	go c.ruleInfs.Start(ctx.Done())
	go c.cmapInfs.Start(ctx.Done())
	go c.secrInfs.Start(ctx.Done())
	go c.ssetInfs.Start(ctx.Done())
	go c.nsMonInf.Run(ctx.Done())
	if c.nsPromInf != c.nsMonInf {
		go c.nsPromInf.Run(ctx.Done())
	}
	if err := c.waitForCacheSync(ctx); err != nil {
		return err
	}

	// Refresh the status of the existing Prometheus objects.
	_ = c.promInfs.ListAll(labels.Everything(), func(obj interface{}) {
		c.addToStatusQueue(obj)
	})

	c.addHandlers()

	if c.kubeletSyncEnabled {
		go c.reconcileNodeEndpoints(ctx)
	}

	// Run a goroutine that refreshes regularly the Prometheus objects that
	// aren't fully available to keep the status up-to-date with the pod
	// conditions. In practice when a new version of the statefulset is rolled
	// out and the updated pod is crashlooping, the statefulset status won't
	// see any update because the number of ready/updated replicas doesn't
	// change. Without the periodic refresh, the Prometheus object's status
	// would report "containers with incomplete status: [init-config-reloader]"
	// forever.
	// TODO(simonpasquier): watch for Prometheus pods instead of polling.
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := c.promInfs.ListAll(labels.Everything(), func(o interface{}) {
					p := o.(*monitoringv1.Prometheus)
					for _, cond := range p.Status.Conditions {
						if cond.Type == monitoringv1.PrometheusAvailable && cond.Status != monitoringv1.PrometheusConditionTrue {
							c.addToStatusQueue(p)
							break
						}
					}
				})
				if err != nil {
					level.Error(c.logger).Log("msg", "failed to list Prometheus objects", "err", err)
				}
			}
		}
	}()

	c.metrics.Ready().Set(1)
	<-ctx.Done()
	return nil
}

func (c *Operator) keyFunc(obj interface{}) (string, bool) {
	k, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		level.Error(c.logger).Log("msg", "creating key failed", "err", err)
		return "", false
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
	c.addToReconcileQueue(key)
}

func (c *Operator) handlePrometheusDelete(obj interface{}) {
	key, ok := c.keyFunc(obj)
	if !ok {
		return
	}

	level.Debug(c.logger).Log("msg", "Prometheus deleted", "key", key)
	c.metrics.TriggerByCounter(monitoringv1.PrometheusesKind, "delete").Inc()
	c.addToReconcileQueue(key)
}

// hasStateChanged returns true if the 2 objects are different in a way that
// the controller should reconcile the actual state against the desired state.
// It helps preventing hot loops when the controller updates the status
// subresource for instance.
func (c *Operator) hasStateChanged(old, cur metav1.Object) bool {
	if old.GetGeneration() != cur.GetGeneration() {
		level.Debug(c.logger).Log(
			"msg", "different generations",
			"current", cur.GetGeneration(),
			"old", old.GetGeneration(),
			"object", fmt.Sprintf("%s/%s", cur.GetNamespace(), cur.GetName()),
		)
		return true
	}

	if !reflect.DeepEqual(old.GetLabels(), cur.GetLabels()) {
		level.Debug(c.logger).Log(
			"msg", "different labels",
			"current", fmt.Sprintf("%v", cur.GetLabels()),
			"old", fmt.Sprintf("%v", old.GetLabels()),
			"object", fmt.Sprintf("%s/%s", cur.GetNamespace(), cur.GetName()),
		)
		return true

	}
	if !reflect.DeepEqual(old.GetAnnotations(), cur.GetAnnotations()) {
		level.Debug(c.logger).Log(
			"msg", "different annotations",
			"current", fmt.Sprintf("%v", cur.GetAnnotations()),
			"old", fmt.Sprintf("%v", old.GetAnnotations()),
			"object", fmt.Sprintf("%s/%s", cur.GetNamespace(), cur.GetName()),
		)
		return true
	}

	return false
}

// hasObjectChanged returns true if the 2 objects are different.
func (c *Operator) hasObjectChanged(old, cur metav1.Object) bool {
	if old.GetResourceVersion() != cur.GetResourceVersion() {
		level.Debug(c.logger).Log(
			"msg", "different resource versions",
			"current", cur.GetResourceVersion(),
			"old", old.GetResourceVersion(),
			"object", fmt.Sprintf("%s/%s", cur.GetNamespace(), cur.GetName()),
		)
		return true
	}

	return false
}

func (c *Operator) handlePrometheusUpdate(old, cur interface{}) {
	if !c.hasStateChanged(&old.(*monitoringv1.Prometheus).ObjectMeta, &cur.(*monitoringv1.Prometheus).ObjectMeta) {
		return
	}

	key, ok := c.keyFunc(cur)
	if !ok {
		return
	}

	level.Debug(c.logger).Log("msg", "Prometheus updated", "key", key)
	c.metrics.TriggerByCounter(monitoringv1.PrometheusesKind, "update").Inc()
	checkPrometheusSpecDeprecation(key, cur.(*monitoringv1.Prometheus), c.logger)
	c.addToReconcileQueue(key)
}

func (c *Operator) reconcileNodeEndpoints(ctx context.Context) {
	c.syncNodeEndpointsWithLogError(ctx)
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.syncNodeEndpointsWithLogError(ctx)
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

func (c *Operator) syncNodeEndpointsWithLogError(ctx context.Context) {
	level.Debug(c.logger).Log("msg", "Syncing nodes into Endpoints object")

	c.nodeEndpointSyncs.Inc()
	err := c.syncNodeEndpoints(ctx)
	if err != nil {
		c.nodeEndpointSyncErrors.Inc()
		level.Error(c.logger).Log("msg", "Syncing nodes into Endpoints object failed", "err", err)
	}
}

func (c *Operator) syncNodeEndpoints(ctx context.Context) error {
	logger := log.With(c.logger, "operation", "syncNodeEndpoints")
	eps := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.kubeletObjectName,
			Labels: c.config.Labels.Merge(map[string]string{
				"k8s-app":                      "kubelet",
				"app.kubernetes.io/name":       "kubelet",
				"app.kubernetes.io/managed-by": "prometheus-operator",
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

	nodes, err := c.kclient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "listing nodes failed")
	}

	level.Debug(logger).Log("msg", "Nodes retrieved from the Kubernetes API", "num_nodes", len(nodes.Items))

	addresses, errs := getNodeAddresses(nodes)
	if len(errs) > 0 {
		for _, err := range errs {
			level.Warn(logger).Log("err", err)
		}
		c.nodeAddressLookupErrors.Add(float64(len(errs)))
	}
	level.Debug(logger).Log("msg", "Nodes converted to endpoint addresses", "num_addresses", len(addresses))

	eps.Subsets[0].Addresses = addresses

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.kubeletObjectName,
			Labels: c.config.Labels.Merge(map[string]string{
				"k8s-app":                      "kubelet",
				"app.kubernetes.io/name":       "kubelet",
				"app.kubernetes.io/managed-by": "prometheus-operator",
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

	level.Debug(logger).Log("msg", "Updating Kubernetes service", "service", c.kubeletObjectName, "ns", c.kubeletObjectNamespace)
	err = k8sutil.CreateOrUpdateService(ctx, c.kclient.CoreV1().Services(c.kubeletObjectNamespace), svc)
	if err != nil {
		return errors.Wrap(err, "synchronizing kubelet service object failed")
	}

	level.Debug(logger).Log("msg", "Updating Kubernetes endpoint", "endpoint", c.kubeletObjectName, "ns", c.kubeletObjectNamespace)
	err = k8sutil.CreateOrUpdateEndpoints(ctx, c.kclient.CoreV1().Endpoints(c.kubeletObjectNamespace), eps)
	if err != nil {
		return errors.Wrap(err, "synchronizing kubelet endpoints object failed")
	}

	return nil
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleSmonAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ServiceMonitor added")
		c.metrics.TriggerByCounter(monitoringv1.ServiceMonitorsKind, "add").Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
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

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleSmonDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ServiceMonitor delete")
		c.metrics.TriggerByCounter(monitoringv1.ServiceMonitorsKind, "delete").Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handlePmonAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PodMonitor added")
		c.metrics.TriggerByCounter(monitoringv1.PodMonitorsKind, "add").Inc()
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
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

// TODO: Don't enqueue just for the namespace
func (c *Operator) handlePmonDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PodMonitor delete")
		c.metrics.TriggerByCounter(monitoringv1.PodMonitorsKind, "delete").Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleBmonAdd(obj interface{}) {
	if o, ok := c.getObject(obj); ok {
		level.Debug(c.logger).Log("msg", "Probe added")
		c.metrics.TriggerByCounter(monitoringv1.ProbesKind, "add").Inc()
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
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

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleBmonDelete(obj interface{}) {
	if o, ok := c.getObject(obj); ok {
		level.Debug(c.logger).Log("msg", "Probe delete")
		c.metrics.TriggerByCounter(monitoringv1.ProbesKind, "delete").Inc()
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleRuleAdd(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PrometheusRule added")
		c.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, "add").Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
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

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleRuleDelete(obj interface{}) {
	o, ok := c.getObject(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PrometheusRule deleted")
		c.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, "delete").Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Do we need to enqueue secrets just for the namespace or in general?
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

// TODO: Do we need to enqueue configmaps just for the namespace or in general?
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

// addToReconcileQueue adds the object to the reconciliation queue.
func (c *Operator) addToReconcileQueue(obj interface{}) {
	c.addToQueue(obj, c.reconcileQueue)
}

// addToStatusQueue adds the object to the status queue.
func (c *Operator) addToStatusQueue(obj interface{}) {
	c.addToQueue(obj, c.statusQueue)
}

// addToQueue adds the object to the given queue.
// If the object is a string, it gets added directly. Otherwise, the object's
// key is extracted via keyFunc.
func (c *Operator) addToQueue(obj interface{}, q workqueue.Interface) {
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

	q.Add(key)
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

	err = c.promInfs.ListAll(labels.Everything(), func(obj interface{}) {
		// Check for Prometheus instances in the namespace.
		p := obj.(*monitoringv1.Prometheus)
		if p.Namespace == nsName {
			c.addToReconcileQueue(p)
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
			c.addToReconcileQueue(p)
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
			c.addToReconcileQueue(p)
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
			c.addToReconcileQueue(p)
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
			c.addToReconcileQueue(p)
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

// processNextReconcileItem dequeues items, processes them, and marks them done.
// It is guaranteed that the sync() method is never invoked concurrently with
// the same key.
// Before returning, the object's key is automatically added to the status queue.
func (c *Operator) processNextReconcileItem(ctx context.Context) bool {
	item, quit := c.reconcileQueue.Get()
	if quit {
		return false
	}
	key := item.(string)
	defer c.reconcileQueue.Done(key)
	defer c.addToStatusQueue(key) // enqueues the object's key to update the status subresource

	c.metrics.ReconcileCounter().Inc()
	startTime := time.Now()
	err := c.sync(ctx, key)
	c.metrics.ReconcileDurationHistogram().Observe(time.Since(startTime).Seconds())
	c.reconciliations.SetStatus(key, err)

	if err == nil {
		c.reconcileQueue.Forget(key)
		return true
	}

	c.metrics.ReconcileErrorsCounter().Inc()
	utilruntime.HandleError(errors.Wrap(err, fmt.Sprintf("sync %q failed", key)))
	c.reconcileQueue.AddRateLimited(key)

	return true
}

func (c *Operator) processNextStatusItem(ctx context.Context) bool {
	key, quit := c.statusQueue.Get()
	if quit {
		return false
	}
	defer c.statusQueue.Done(key)

	err := c.status(ctx, key.(string))
	if err == nil {
		c.statusQueue.Forget(key)
		return true
	}

	utilruntime.HandleError(errors.Wrap(err, fmt.Sprintf("status %q failed", key)))
	c.statusQueue.AddRateLimited(key)

	return true
}

func (c *Operator) prometheusForStatefulSet(sset interface{}) *monitoringv1.Prometheus {
	key, ok := c.keyFunc(sset)
	if !ok {
		return nil
	}

	match, promKey := statefulSetKeyToPrometheusKey(key)
	if !match {
		level.Debug(c.logger).Log("msg", "StatefulSet key did not match a Prometheus key format", "key", key)
		return nil
	}

	p, err := c.promInfs.Get(promKey)
	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		level.Error(c.logger).Log("msg", "Prometheus lookup failed", "err", err)
		return nil
	}

	return p.(*monitoringv1.Prometheus)
}

func statefulSetNameFromPrometheusName(name string, shard int) string {
	if shard == 0 {
		return fmt.Sprintf("prometheus-%s", name)
	}
	return fmt.Sprintf("prometheus-%s-shard-%d", name, shard)
}

var prometheusKeyInShardStatefulSet = regexp.MustCompile("^(.+)/prometheus-(.+)-shard-[1-9][0-9]*$")
var prometheusKeyInStatefulSet = regexp.MustCompile("^(.+)/prometheus-(.+)$")

func statefulSetKeyToPrometheusKey(key string) (bool, string) {
	r := prometheusKeyInStatefulSet
	if prometheusKeyInShardStatefulSet.MatchString(key) {
		r = prometheusKeyInShardStatefulSet
	}

	matches := r.FindAllStringSubmatch(key, 2)
	if len(matches) != 1 {
		return false, ""
	}
	if len(matches[0]) != 3 {
		return false, ""
	}
	return true, matches[0][1] + "/" + matches[0][2]
}

func prometheusKeyToStatefulSetKey(key string, shard int) string {
	keyParts := strings.Split(key, "/")
	return fmt.Sprintf("%s/%s", keyParts[0], statefulSetNameFromPrometheusName(keyParts[1], shard))
}

func (c *Operator) handleStatefulSetDelete(obj interface{}) {
	ps := c.prometheusForStatefulSet(obj)
	if ps == nil {
		return
	}

	level.Debug(c.logger).Log("msg", "StatefulSet delete")
	c.metrics.TriggerByCounter("StatefulSet", "delete").Inc()

	c.addToReconcileQueue(ps)
}

func (c *Operator) handleStatefulSetAdd(obj interface{}) {
	ps := c.prometheusForStatefulSet(obj)
	if ps == nil {
		return
	}

	level.Debug(c.logger).Log("msg", "StatefulSet added")
	c.metrics.TriggerByCounter("StatefulSet", "add").Inc()

	c.addToReconcileQueue(ps)
}

func (c *Operator) handleStatefulSetUpdate(oldo, curo interface{}) {
	old := oldo.(*appsv1.StatefulSet)
	cur := curo.(*appsv1.StatefulSet)

	level.Debug(c.logger).Log("msg", "update handler", "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	if !c.hasObjectChanged(old, cur) {
		return
	}

	p := c.prometheusForStatefulSet(cur)
	if p == nil {
		return
	}

	level.Debug(c.logger).Log("msg", "StatefulSet updated")
	c.metrics.TriggerByCounter("StatefulSet", "update").Inc()

	if !c.hasStateChanged(old, cur) {
		// If the statefulset state (spec, labels or annotations) hasn't
		// changed, the operator can only update the status subresource instead
		// of doing a full reconciliation.
		c.addToStatusQueue(p)
		return
	}

	c.addToReconcileQueue(p)
}

func (c *Operator) handleMonitorNamespaceUpdate(oldo, curo interface{}) {
	old := oldo.(*v1.Namespace)
	cur := curo.(*v1.Namespace)

	level.Debug(c.logger).Log("msg", "update handler", "namespace", cur.GetName(), "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	// Periodic resync may resend the Namespace without changes
	// in-between.
	if old.ResourceVersion == cur.ResourceVersion {
		return
	}

	level.Debug(c.logger).Log("msg", "Monitor namespace updated", "namespace", cur.GetName())
	c.metrics.TriggerByCounter("Namespace", "update").Inc()

	// Check for Prometheus instances selecting ServiceMonitors, PodMonitors,
	// Probes and PrometheusRules in the namespace.
	err := c.promInfs.ListAll(labels.Everything(), func(obj interface{}) {
		p := obj.(*monitoringv1.Prometheus)

		for name, selector := range map[string]*metav1.LabelSelector{
			"PodMonitors":     p.Spec.PodMonitorNamespaceSelector,
			"Probes":          p.Spec.ProbeNamespaceSelector,
			"PrometheusRules": p.Spec.RuleNamespaceSelector,
			"ServiceMonitors": p.Spec.ServiceMonitorNamespaceSelector,
		} {

			sync, err := k8sutil.LabelSelectionHasChanged(old.Labels, cur.Labels, selector)
			if err != nil {
				level.Error(c.logger).Log(
					"err", err,
					"name", p.Name,
					"namespace", p.Namespace,
					"subresource", name,
				)
				return
			}

			if sync {
				c.addToReconcileQueue(p)
				return
			}
		}
	})
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "listing all Prometheus instances from cache failed",
			"err", err,
		)
	}
}

func (c *Operator) sync(ctx context.Context, key string) error {
	pobj, err := c.promInfs.Get(key)

	if apierrors.IsNotFound(err) {
		c.reconciliations.ForgetObject(key)
		// Dependent resources are cleaned up by K8s via OwnerReferences
		return nil
	}
	if err != nil {
		return err
	}

	p := pobj.(*monitoringv1.Prometheus)
	p = p.DeepCopy()
	if err := k8sutil.AddTypeInformationToObject(p); err != nil {
		return errors.Wrap(err, "failed to set Prometheus type information")
	}

	logger := log.With(c.logger, "key", key)
	if p.Spec.Paused {
		level.Info(logger).Log("msg", "the resource is paused, not reconciling")
		return nil
	}

	level.Info(logger).Log("msg", "sync prometheus")
	ruleConfigMapNames, err := c.createOrUpdateRuleConfigMaps(ctx, p)
	if err != nil {
		return err
	}

	assetStore := assets.NewStore(c.kclient.CoreV1(), c.kclient.CoreV1())

	if err := c.createOrUpdateConfigurationSecret(ctx, p, ruleConfigMapNames, assetStore); err != nil {
		return errors.Wrap(err, "creating config failed")
	}

	tlsAssets, err := c.createOrUpdateTLSAssetSecrets(ctx, p, assetStore)
	if err != nil {
		return errors.Wrap(err, "creating tls asset secret failed")
	}

	if err := c.createOrUpdateWebConfigSecret(ctx, p); err != nil {
		return errors.Wrap(err, "synchronizing web config secret failed")
	}

	// Create governing service if it doesn't exist.
	svcClient := c.kclient.CoreV1().Services(p.Namespace)
	if err := k8sutil.CreateOrUpdateService(ctx, svcClient, makeStatefulSetService(p, c.config)); err != nil {
		return errors.Wrap(err, "synchronizing governing service failed")
	}

	ssetClient := c.kclient.AppsV1().StatefulSets(p.Namespace)

	// Ensure we have a StatefulSet running Prometheus deployed and that StatefulSet names are created correctly.
	expected := expectedStatefulSetShardNames(p)
	for shard, ssetName := range expected {
		logger := log.With(logger, "statefulset", ssetName, "shard", fmt.Sprintf("%d", shard))
		level.Debug(logger).Log("msg", "reconciling statefulset")

		obj, err := c.ssetInfs.Get(prometheusKeyToStatefulSetKey(key, shard))
		exists := !apierrors.IsNotFound(err)
		if err != nil && !apierrors.IsNotFound(err) {
			return errors.Wrap(err, "retrieving statefulset failed")
		}

		existingStatefulSet := &appsv1.StatefulSet{}
		if obj != nil {
			existingStatefulSet = obj.(*appsv1.StatefulSet)
			if existingStatefulSet.DeletionTimestamp != nil {
				// We want to avoid entering a hot-loop of update/delete cycles
				// here since the sts was marked for deletion in foreground,
				// which means it may take some time before the finalizers
				// complete and the resource disappears from the API. The
				// deletion timestamp will have been set when the initial
				// delete request was issued. In that case, we avoid further
				// processing.
				level.Info(logger).Log(
					"msg", "halting update of StatefulSet",
					"reason", "resource has been marked for deletion",
					"resource_name", existingStatefulSet.GetName(),
				)
				continue
			}
		}

		newSSetInputHash, err := createSSetInputHash(*p, c.config, ruleConfigMapNames, tlsAssets, existingStatefulSet.Spec)
		if err != nil {
			return err
		}

		sset, err := makeStatefulSet(logger, ssetName, *p, &c.config, ruleConfigMapNames, newSSetInputHash, int32(shard), tlsAssets.ShardNames())
		if err != nil {
			return errors.Wrap(err, "making statefulset failed")
		}
		operator.SanitizeSTS(sset)

		if !exists {
			level.Debug(logger).Log("msg", "no current statefulset found")
			level.Debug(logger).Log("msg", "creating statefulset")
			if _, err := ssetClient.Create(ctx, sset, metav1.CreateOptions{}); err != nil {
				return errors.Wrap(err, "creating statefulset failed")
			}
			continue
		}

		if newSSetInputHash == existingStatefulSet.ObjectMeta.Annotations[sSetInputHashName] {
			level.Debug(logger).Log("msg", "new statefulset generation inputs match current, skipping any actions")
			continue
		}

		level.Debug(logger).Log(
			"msg", "updating current statefulset because of hash divergence",
			"new_hash", newSSetInputHash,
			"existing_hash", existingStatefulSet.ObjectMeta.Annotations[sSetInputHashName],
		)

		err = k8sutil.UpdateStatefulSet(ctx, ssetClient, sset)
		sErr, ok := err.(*apierrors.StatusError)

		if ok && sErr.ErrStatus.Code == 422 && sErr.ErrStatus.Reason == metav1.StatusReasonInvalid {
			c.metrics.StsDeleteCreateCounter().Inc()

			// Gather only reason for failed update
			failMsg := make([]string, len(sErr.ErrStatus.Details.Causes))
			for i, cause := range sErr.ErrStatus.Details.Causes {
				failMsg[i] = cause.Message
			}

			level.Info(logger).Log("msg", "recreating StatefulSet because the update operation wasn't possible", "reason", strings.Join(failMsg, ", "))

			propagationPolicy := metav1.DeletePropagationForeground
			if err := ssetClient.Delete(ctx, sset.GetName(), metav1.DeleteOptions{PropagationPolicy: &propagationPolicy}); err != nil {
				return errors.Wrap(err, "failed to delete StatefulSet to avoid forbidden action")
			}
			continue
		}

		if err != nil {
			return errors.Wrap(err, "updating StatefulSet failed")
		}
	}

	ssets := map[string]struct{}{}
	for _, ssetName := range expected {
		ssets[ssetName] = struct{}{}
	}

	err = c.ssetInfs.ListAllByNamespace(p.Namespace, labels.SelectorFromSet(labels.Set{prometheusNameLabelName: p.Name}), func(obj interface{}) {
		s := obj.(*appsv1.StatefulSet)

		if _, ok := ssets[s.Name]; ok {
			// Do not delete statefulsets that we still expect to exist. This
			// is to cleanup StatefulSets when shards are reduced.
			return
		}

		// Deletion already in progress.
		if s.DeletionTimestamp != nil {
			return
		}

		propagationPolicy := metav1.DeletePropagationForeground
		if err := ssetClient.Delete(ctx, s.GetName(), metav1.DeleteOptions{PropagationPolicy: &propagationPolicy}); err != nil {
			level.Error(c.logger).Log("err", err, "name", s.GetName(), "namespace", s.GetNamespace())
		}
	})
	if err != nil {
		return errors.Wrap(err, "listing StatefulSet resources failed")
	}

	return nil
}

// status updates the status subresource of the object identified by the given
// key.
func (c *Operator) status(ctx context.Context, key string) error {
	pobj, err := c.promInfs.Get(key)

	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	p := pobj.(*monitoringv1.Prometheus)
	p = p.DeepCopy()

	pStatus := monitoringv1.PrometheusStatus{
		Paused: p.Spec.Paused,
	}

	logger := log.With(c.logger, "key", key)
	level.Info(logger).Log("msg", "update prometheus status")

	var (
		availableCondition = monitoringv1.PrometheusCondition{
			Type:   monitoringv1.PrometheusAvailable,
			Status: monitoringv1.PrometheusConditionTrue,
			LastTransitionTime: metav1.Time{
				Time: time.Now().UTC(),
			},
			ObservedGeneration: p.Generation,
		}
		messages []string
	)

	for shard := range expectedStatefulSetShardNames(p) {
		ssetName := prometheusKeyToStatefulSetKey(key, shard)
		logger := log.With(logger, "statefulset", ssetName, "shard", shard)

		obj, err := c.ssetInfs.Get(ssetName)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// Object not yet in the store or already deleted.
				level.Info(logger).Log("msg", "not found")
				continue
			}
			return errors.Wrap(err, "failed to retrieve statefulset")
		}

		sset := obj.(*appsv1.StatefulSet)
		if sset.DeletionTimestamp != nil {
			level.Debug(logger).Log("msg", "deletion in progress")
			continue
		}

		stsReporter, err := newStatefulSetReporter(ctx, c.kclient, sset)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve statefulset state")
		}

		pStatus.Replicas += int32(len(stsReporter.pods))
		pStatus.UpdatedReplicas += int32(len(stsReporter.Updated()))
		pStatus.AvailableReplicas += int32(len(stsReporter.Ready()))
		pStatus.UnavailableReplicas += int32(len(stsReporter.pods) - len(stsReporter.Ready()))

		pStatus.ShardStatuses = append(
			pStatus.ShardStatuses,
			monitoringv1.ShardStatus{
				ShardID:             strconv.Itoa(shard),
				Replicas:            int32(len(stsReporter.pods)),
				UpdatedReplicas:     int32(len(stsReporter.Updated())),
				AvailableReplicas:   int32(len(stsReporter.Ready())),
				UnavailableReplicas: int32(len(stsReporter.pods) - len(stsReporter.Ready())),
			},
		)

		if len(stsReporter.Ready()) == len(stsReporter.pods) {
			// All pods are ready (or the desired number of replicas is zero).
			continue
		}

		if len(stsReporter.Ready()) == 0 {
			availableCondition.Reason = "NoPodReady"
			availableCondition.Status = monitoringv1.PrometheusConditionFalse
		} else if availableCondition.Status != monitoringv1.PrometheusConditionFalse {
			availableCondition.Reason = "SomePodsNotReady"
			availableCondition.Status = monitoringv1.PrometheusConditionDegraded
		}

		for _, p := range stsReporter.pods {
			if m := p.Message(); m != "" {
				messages = append(messages, fmt.Sprintf("shard %d: pod %s: %s", shard, p.Name, m))
			}
		}
	}

	availableCondition.Message = strings.Join(messages, "\n")

	// Compute the Reconciled ConditionType.
	reconciledCondition := monitoringv1.PrometheusCondition{
		Type:   monitoringv1.PrometheusReconciled,
		Status: monitoringv1.PrometheusConditionTrue,
		LastTransitionTime: metav1.Time{
			Time: time.Now().UTC(),
		},
		ObservedGeneration: p.Generation,
	}
	reconciliationStatus, found := c.reconciliations.GetStatus(key)
	if !found {
		reconciledCondition.Status = monitoringv1.PrometheusConditionUnknown
		reconciledCondition.Reason = "NotFound"
		reconciledCondition.Message = fmt.Sprintf("object %q not found", key)
	} else {
		if !reconciliationStatus.Ok() {
			reconciledCondition.Status = monitoringv1.PrometheusConditionFalse
		}
		reconciledCondition.Reason = reconciliationStatus.Reason()
		reconciledCondition.Message = reconciliationStatus.Message()
	}

	// Update the last transition times only if the status of the available condition has changed.
	for _, condition := range p.Status.Conditions {
		if condition.Type == availableCondition.Type && condition.Status == availableCondition.Status {
			availableCondition.LastTransitionTime = condition.LastTransitionTime
			continue
		}

		if condition.Type == reconciledCondition.Type && condition.Status == reconciledCondition.Status {
			reconciledCondition.LastTransitionTime = condition.LastTransitionTime
		}
	}

	pStatus.Conditions = append(pStatus.Conditions, availableCondition, reconciledCondition)

	p.Status = pStatus
	if _, err = c.mclient.MonitoringV1().Prometheuses(p.Namespace).UpdateStatus(ctx, p, metav1.UpdateOptions{}); err != nil {
		return errors.Wrap(err, "failed to update status subresource")
	}

	return nil
}

// checkPrometheusSpecDeprecation checks for deprecated fields in the prometheus spec and logs a warning if applicable
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

func createSSetInputHash(p monitoringv1.Prometheus, c operator.Config, ruleConfigMapNames []string, tlsAssets *operator.ShardedSecret, ssSpec appsv1.StatefulSetSpec) (string, error) {
	var http2 *bool
	if p.Spec.Web != nil && p.Spec.Web.WebConfigFileFields.HTTPConfig != nil {
		http2 = p.Spec.Web.WebConfigFileFields.HTTPConfig.HTTP2
	}

	hash, err := hashstructure.Hash(struct {
		PrometheusLabels      map[string]string
		PrometheusAnnotations map[string]string
		PrometheusGeneration  int64
		PrometheusWebHTTP2    *bool
		Config                operator.Config
		StatefulSetSpec       appsv1.StatefulSetSpec
		RuleConfigMaps        []string `hash:"set"`
		Assets                []string `hash:"set"`
	}{
		PrometheusLabels:      p.Labels,
		PrometheusAnnotations: p.Annotations,
		PrometheusGeneration:  p.Generation,
		PrometheusWebHTTP2:    http2,
		Config:                c,
		StatefulSetSpec:       ssSpec,
		RuleConfigMaps:        ruleConfigMapNames,
		Assets:                tlsAssets.ShardNames(),
	},
		nil,
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to calculate combined hash")
	}

	return fmt.Sprintf("%d", hash), nil
}

func ListOptions(name string) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
			"app.kubernetes.io/name": "prometheus",
			"prometheus":             name,
		})).String(),
	}
}

type pod v1.Pod

//// Ready returns true if the pod matches with the statefulset's revision.
//func (p *pod) Updated() bool {
//   return p.revision == p.Labels["controller-revision-hash"]
//}

// Ready returns true if the pod is ready.
func (p *pod) Ready() bool {
	if p.Status.Phase != v1.PodRunning {
		return false
	}

	for _, cond := range p.Status.Conditions {
		if cond.Type != v1.PodReady {
			continue
		}
		return cond.Status == v1.ConditionTrue
	}

	return false
}

// Message returns a human-readable and terse message about the state of the pod.
func (p *pod) Message() string {
	for _, condType := range []v1.PodConditionType{
		v1.PodScheduled,    // Check first that the pod is scheduled.
		v1.PodInitialized,  // Then that init containers have been started successfully.
		v1.ContainersReady, // Then that all containers are ready.
		v1.PodReady,        // And finally that the pod is ready.
	} {
		for _, cond := range p.Status.Conditions {
			if cond.Type == condType && cond.Status == v1.ConditionFalse {
				return cond.Message
			}
		}
	}

	return ""
}

type statefulSetReporter struct {
	pods []*pod
	sset *appsv1.StatefulSet
}

// Updated returns the list of pods that match with the statefulset's revision.
func (sr *statefulSetReporter) Updated() []*pod {
	return sr.filterPods(func(p *pod) bool {
		return sr.IsUpdated(p)
	})
}

// IsUpdated returns true if the given pod matches with the statefulset's revision.
func (sr *statefulSetReporter) IsUpdated(p *pod) bool {
	return sr.sset.Status.UpdateRevision == p.Labels["controller-revision-hash"]
}

// Ready returns the list of pods that are ready.
func (sr *statefulSetReporter) Ready() []*pod {
	return sr.filterPods(func(p *pod) bool {
		return p.Ready()
	})
}

func (sr *statefulSetReporter) filterPods(f func(*pod) bool) []*pod {
	pods := make([]*pod, 0, len(sr.pods))

	for _, p := range sr.pods {
		if f(p) {
			pods = append(pods, p)
		}
	}

	return pods
}

// getPodsState returns the state of pods which are targeted by the given StatefulSet.
func newStatefulSetReporter(ctx context.Context, kclient kubernetes.Interface, sset *appsv1.StatefulSet) (*statefulSetReporter, error) {
	ls, err := metav1.LabelSelectorAsSelector(sset.Spec.Selector)
	if err != nil {
		// Something is really broken if the statefulset's selector isn't valid.
		panic(err)
	}

	pods, err := kclient.CoreV1().Pods(sset.Namespace).List(ctx, metav1.ListOptions{LabelSelector: ls.String()})
	if err != nil {
		return nil, err
	}

	stsReporter := &statefulSetReporter{
		sset: sset,
		pods: make([]*pod, 0, len(pods.Items)),
	}
	for _, p := range pods.Items {
		var found bool
		for _, owner := range p.ObjectMeta.OwnerReferences {
			if owner.Kind == "StatefulSet" && owner.Name == sset.Name {
				found = true
				break
			}
		}

		if !found {
			continue
		}

		stsReporter.pods = append(stsReporter.pods, (func(p pod) *pod { return &p })(pod(p)))
	}

	return stsReporter, nil
}

// Status evaluates the current status of a Prometheus deployment with
// respect to its specified resource object. It returns the status and a list of
// pods that are not updated.
// TODO(simonpasquier): remove once the status subresource is considered stable.
func Status(ctx context.Context, kclient kubernetes.Interface, p *monitoringv1.Prometheus) (monitoringv1.PrometheusStatus, []v1.Pod, error) {
	res := monitoringv1.PrometheusStatus{Paused: p.Spec.Paused}

	var oldPods []v1.Pod
	for _, ssetName := range expectedStatefulSetShardNames(p) {
		sset, err := kclient.AppsV1().StatefulSets(p.Namespace).Get(ctx, ssetName, metav1.GetOptions{})
		if err != nil {
			return monitoringv1.PrometheusStatus{}, nil, errors.Wrapf(err, "failed to retrieve statefulset %s/%s", p.Namespace, ssetName)
		}

		stsReporter, err := newStatefulSetReporter(ctx, kclient, sset)
		if err != nil {
			return monitoringv1.PrometheusStatus{}, nil, errors.Wrapf(err, "failed to retrieve pods state for statefulset %s/%s", p.Namespace, ssetName)
		}

		res.Replicas += int32(len(stsReporter.pods))
		res.UpdatedReplicas += int32(len(stsReporter.Updated()))
		res.AvailableReplicas += int32(len(stsReporter.Ready()))
		res.UnavailableReplicas += int32(len(stsReporter.pods) - len(stsReporter.Ready()))

		for _, p := range stsReporter.pods {
			if p.Ready() && !stsReporter.IsUpdated(p) {
				oldPods = append(oldPods, v1.Pod(*p))
			}
		}
	}

	return res, oldPods, nil
}

func (c *Operator) loadConfigFromSecret(sks *v1.SecretKeySelector, s *v1.SecretList) ([]byte, error) {
	if sks == nil {
		return nil, nil
	}

	for _, secret := range s.Items {
		if secret.Name == sks.Name {
			if c, ok := secret.Data[sks.Key]; ok {
				return c, nil
			}

			return nil, fmt.Errorf("key %v could not be found in secret %v", sks.Key, sks.Name)
		}
	}

	if sks.Optional == nil || !*sks.Optional {
		return nil, fmt.Errorf("secret %v could not be found", sks.Name)
	}

	level.Debug(c.logger).Log("msg", fmt.Sprintf("secret %v could not be found", sks.Name))
	return nil, nil
}

func (c *Operator) createOrUpdateConfigurationSecret(ctx context.Context, p *monitoringv1.Prometheus, ruleConfigMapNames []string, store *assets.Store) error {
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
		_, err = sClient.Get(ctx, s.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			if _, err := c.kclient.CoreV1().Secrets(p.Namespace).Create(ctx, s, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
				return errors.Wrap(err, "creating empty config file failed")
			}
		}
		if !apierrors.IsNotFound(err) && err != nil {
			return err
		}

		return nil
	}

	smons, err := c.selectServiceMonitors(ctx, p, store)
	if err != nil {
		return errors.Wrap(err, "selecting ServiceMonitors failed")
	}

	pmons, err := c.selectPodMonitors(ctx, p, store)
	if err != nil {
		return errors.Wrap(err, "selecting PodMonitors failed")
	}

	bmons, err := c.selectProbes(ctx, p, store)
	if err != nil {
		return errors.Wrap(err, "selecting Probes failed")
	}
	sClient := c.kclient.CoreV1().Secrets(p.Namespace)
	SecretsInPromNS, err := sClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for i, remote := range p.Spec.RemoteRead {
		if err := store.AddBasicAuth(ctx, p.GetNamespace(), remote.BasicAuth, fmt.Sprintf("remoteRead/%d", i)); err != nil {
			return errors.Wrapf(err, "remote read %d", i)
		}
		if err := store.AddOAuth2(ctx, p.GetNamespace(), remote.OAuth2, fmt.Sprintf("remoteRead/%d", i)); err != nil {
			return errors.Wrapf(err, "remote read %d", i)
		}
		if err := store.AddTLSConfig(ctx, p.GetNamespace(), remote.TLSConfig); err != nil {
			return errors.Wrapf(err, "remote read %d", i)
		}
		if err := store.AddAuthorizationCredentials(ctx, p.GetNamespace(), remote.Authorization, fmt.Sprintf("remoteRead/auth/%d", i)); err != nil {
			return errors.Wrapf(err, "remote read %d", i)
		}
	}

	for i, remote := range p.Spec.RemoteWrite {
		if err := validateRemoteWriteSpec(remote); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		key := fmt.Sprintf("remoteWrite/%d", i)
		if err := store.AddBasicAuth(ctx, p.GetNamespace(), remote.BasicAuth, key); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		if err := store.AddOAuth2(ctx, p.GetNamespace(), remote.OAuth2, key); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		if err := store.AddTLSConfig(ctx, p.GetNamespace(), remote.TLSConfig); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		if err := store.AddAuthorizationCredentials(ctx, p.GetNamespace(), remote.Authorization, fmt.Sprintf("remoteWrite/auth/%d", i)); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		if err := store.AddSigV4(ctx, p.GetNamespace(), remote.Sigv4, key); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
	}

	if p.Spec.APIServerConfig != nil {
		if err := store.AddBasicAuth(ctx, p.GetNamespace(), p.Spec.APIServerConfig.BasicAuth, "apiserver"); err != nil {
			return errors.Wrap(err, "apiserver config")
		}
		if err := store.AddAuthorizationCredentials(ctx, p.GetNamespace(), p.Spec.APIServerConfig.Authorization, "apiserver/auth"); err != nil {
			return errors.Wrapf(err, "apiserver config")
		}
	}
	if p.Spec.Alerting != nil {
		for i, am := range p.Spec.Alerting.Alertmanagers {
			if err := store.AddSafeAuthorizationCredentials(ctx, p.GetNamespace(), am.Authorization, fmt.Sprintf("alertmanager/auth/%d", i)); err != nil {
				return errors.Wrapf(err, "apiserver config")
			}
		}
	}

	additionalScrapeConfigs, err := c.loadConfigFromSecret(p.Spec.AdditionalScrapeConfigs, SecretsInPromNS)
	if err != nil {
		return errors.Wrap(err, "loading additional scrape configs from Secret failed")
	}
	additionalAlertRelabelConfigs, err := c.loadConfigFromSecret(p.Spec.AdditionalAlertRelabelConfigs, SecretsInPromNS)
	if err != nil {
		return errors.Wrap(err, "loading additional alert relabel configs from Secret failed")
	}
	additionalAlertManagerConfigs, err := c.loadConfigFromSecret(p.Spec.AdditionalAlertManagerConfigs, SecretsInPromNS)
	if err != nil {
		return errors.Wrap(err, "loading additional alert manager configs from Secret failed")
	}

	cg, err := NewConfigGenerator(c.logger, p, c.endpointSliceSupported)
	if err != nil {
		return err
	}

	// Update secret based on the most recent configuration.
	conf, err := cg.Generate(
		p,
		smons,
		pmons,
		bmons,
		store,
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
	if err = operator.GzipConfig(&buf, conf); err != nil {
		return errors.Wrap(err, "couldn't gzip config")
	}
	s.Data[configFilename] = buf.Bytes()

	level.Debug(c.logger).Log("msg", "updating Prometheus configuration secret")

	return k8sutil.CreateOrUpdateSecret(ctx, sClient, s)
}

func (c *Operator) createOrUpdateTLSAssetSecrets(ctx context.Context, p *monitoringv1.Prometheus, store *assets.Store) (*operator.ShardedSecret, error) {
	labels := c.config.Labels.Merge(managedByOperatorLabels)
	template := newTLSAssetSecret(p, labels)

	sSecret := operator.NewShardedSecret(template, tlsAssetsSecretName(p.Name))

	for k, v := range store.TLSAssets {
		sSecret.AppendData(k.String(), []byte(v))
	}

	sClient := c.kclient.CoreV1().Secrets(p.Namespace)

	if err := sSecret.StoreSecrets(ctx, sClient); err != nil {
		return nil, errors.Wrapf(err, "failed to create TLS assets secret for Prometheus")
	}

	level.Debug(c.logger).Log("msg", "tls-asset secret: stored")

	return sSecret, nil
}

func newTLSAssetSecret(p *monitoringv1.Prometheus, labels map[string]string) *v1.Secret {
	boolTrue := true
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   tlsAssetsSecretName(p.Name),
			Labels: labels,
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
}

func (c *Operator) createOrUpdateWebConfigSecret(ctx context.Context, p *monitoringv1.Prometheus) error {
	boolTrue := true

	var fields monitoringv1.WebConfigFileFields
	if p.Spec.Web != nil {
		fields = p.Spec.Web.WebConfigFileFields
	}

	webConfig, err := webconfig.New(
		webConfigDir,
		webConfigSecretName(p.Name),
		fields,
	)
	if err != nil {
		return errors.Wrap(err, "failed to initialize web config")
	}

	secretClient := c.kclient.CoreV1().Secrets(p.Namespace)
	ownerReference := metav1.OwnerReference{
		APIVersion:         p.APIVersion,
		BlockOwnerDeletion: &boolTrue,
		Controller:         &boolTrue,
		Kind:               p.Kind,
		Name:               p.Name,
		UID:                p.UID,
	}
	secretLabels := c.config.Labels.Merge(managedByOperatorLabels)

	if err := webConfig.CreateOrUpdateWebConfigSecret(ctx, secretClient, secretLabels, ownerReference); err != nil {
		return errors.Wrap(err, "failed to reconcile web config secret")
	}

	return nil
}

func (c *Operator) selectServiceMonitors(ctx context.Context, p *monitoringv1.Prometheus, store *assets.Store) (map[string]*monitoringv1.ServiceMonitor, error) {
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
		err := c.smonInfs.ListAllByNamespace(ns, servMonSelector, func(obj interface{}) {
			k, ok := c.keyFunc(obj)
			if ok {
				svcMon := obj.(*monitoringv1.ServiceMonitor).DeepCopy()
				if err := k8sutil.AddTypeInformationToObject(svcMon); err != nil {
					level.Error(c.logger).Log("msg", "failed to set ServiceMonitor type information", "namespace", ns, "err", err)
					return
				}
				serviceMonitors[k] = svcMon
			}
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list service monitors in namespace %s", ns)
		}
	}

	var rejected int
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

			if err = store.AddBearerToken(ctx, sm.GetNamespace(), endpoint.BearerTokenSecret, smKey); err != nil {
				break
			}

			if err = store.AddBasicAuth(ctx, sm.GetNamespace(), endpoint.BasicAuth, smKey); err != nil {
				break
			}

			if endpoint.TLSConfig != nil {
				if err = store.AddTLSConfig(ctx, sm.GetNamespace(), endpoint.TLSConfig); err != nil {
					break
				}
			}

			if err = store.AddOAuth2(ctx, sm.GetNamespace(), endpoint.OAuth2, smKey); err != nil {
				break
			}

			smAuthKey := fmt.Sprintf("serviceMonitor/auth/%s/%s/%d", sm.GetNamespace(), sm.GetName(), i)
			if err = store.AddSafeAuthorizationCredentials(ctx, sm.GetNamespace(), endpoint.Authorization, smAuthKey); err != nil {
				break
			}

			if err = validateScrapeIntervalAndTimeout(p, endpoint.Interval, endpoint.ScrapeTimeout); err != nil {
				break
			}

			for _, rl := range endpoint.RelabelConfigs {
				if rl.Action != "" {
					if err = validateRelabelConfig(*p, *rl); err != nil {
						break
					}
				}
			}

			for _, rl := range endpoint.MetricRelabelConfigs {
				if rl.Action != "" {
					if err = validateRelabelConfig(*p, *rl); err != nil {
						break
					}
				}
			}
		}

		if err != nil {
			rejected++
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

	if pKey, ok := c.keyFunc(p); ok {
		c.metrics.SetSelectedResources(pKey, monitoringv1.ServiceMonitorsKind, len(res))
		c.metrics.SetRejectedResources(pKey, monitoringv1.ServiceMonitorsKind, rejected)
	}

	return res, nil
}

func (c *Operator) selectPodMonitors(ctx context.Context, p *monitoringv1.Prometheus, store *assets.Store) (map[string]*monitoringv1.PodMonitor, error) {
	namespaces := []string{}
	// Selectors (<namespace>/<name>) might overlap. Deduplicate them along the keyFunc.
	podMonitors := make(map[string]*monitoringv1.PodMonitor)

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

	for _, ns := range namespaces {
		err := c.pmonInfs.ListAllByNamespace(ns, podMonSelector, func(obj interface{}) {
			k, ok := c.keyFunc(obj)
			if ok {
				podMon := obj.(*monitoringv1.PodMonitor).DeepCopy()
				if err := k8sutil.AddTypeInformationToObject(podMon); err != nil {
					level.Error(c.logger).Log("msg", "failed to set PodMonitor type information", "namespace", ns, "err", err)
					return
				}
				podMonitors[k] = podMon
			}
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list pod monitors in namespace %s", ns)
		}
	}

	var rejected int
	res := make(map[string]*monitoringv1.PodMonitor, len(podMonitors))
	for namespaceAndName, pm := range podMonitors {
		var err error

		for i, endpoint := range pm.Spec.PodMetricsEndpoints {
			pmKey := fmt.Sprintf("podMonitor/%s/%s/%d", pm.GetNamespace(), pm.GetName(), i)

			if err = store.AddBearerToken(ctx, pm.GetNamespace(), endpoint.BearerTokenSecret, pmKey); err != nil {
				break
			}

			if err = store.AddBasicAuth(ctx, pm.GetNamespace(), endpoint.BasicAuth, pmKey); err != nil {
				break
			}

			if endpoint.TLSConfig != nil {
				if err = store.AddSafeTLSConfig(ctx, pm.GetNamespace(), &endpoint.TLSConfig.SafeTLSConfig); err != nil {
					break
				}
			}

			if err = store.AddOAuth2(ctx, pm.GetNamespace(), endpoint.OAuth2, pmKey); err != nil {
				break
			}

			pmAuthKey := fmt.Sprintf("podMonitor/auth/%s/%s/%d", pm.GetNamespace(), pm.GetName(), i)
			if err = store.AddSafeAuthorizationCredentials(ctx, pm.GetNamespace(), endpoint.Authorization, pmAuthKey); err != nil {
				break
			}

			if err = validateScrapeIntervalAndTimeout(p, endpoint.Interval, endpoint.ScrapeTimeout); err != nil {
				break
			}

			for _, rl := range endpoint.RelabelConfigs {
				if rl.Action != "" {
					if err = validateRelabelConfig(*p, *rl); err != nil {
						break
					}
				}
			}

			for _, rl := range endpoint.MetricRelabelConfigs {
				if rl.Action != "" {
					if err = validateRelabelConfig(*p, *rl); err != nil {
						break
					}
				}
			}
		}

		if err != nil {
			rejected++
			level.Warn(c.logger).Log(
				"msg", "skipping podmonitor",
				"error", err.Error(),
				"podmonitor", namespaceAndName,
				"namespace", p.Namespace,
				"prometheus", p.Name,
			)
			continue
		}

		res[namespaceAndName] = pm
	}

	pmKeys := []string{}
	for k := range res {
		pmKeys = append(pmKeys, k)
	}
	level.Debug(c.logger).Log("msg", "selected PodMonitors", "podmonitors", strings.Join(pmKeys, ","), "namespace", p.Namespace, "prometheus", p.Name)

	if pKey, ok := c.keyFunc(p); ok {
		c.metrics.SetSelectedResources(pKey, monitoringv1.PodMonitorsKind, len(res))
		c.metrics.SetRejectedResources(pKey, monitoringv1.PodMonitorsKind, rejected)
	}

	return res, nil
}

func (c *Operator) selectProbes(ctx context.Context, p *monitoringv1.Prometheus, store *assets.Store) (map[string]*monitoringv1.Probe, error) {
	namespaces := []string{}
	// Selectors might overlap. Deduplicate them along the keyFunc.
	probes := make(map[string]*monitoringv1.Probe)

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

	for _, ns := range namespaces {
		err := c.probeInfs.ListAllByNamespace(ns, bMonSelector, func(obj interface{}) {
			if k, ok := c.keyFunc(obj); ok {
				probe := obj.(*monitoringv1.Probe).DeepCopy()
				if err := k8sutil.AddTypeInformationToObject(probe); err != nil {
					level.Error(c.logger).Log("msg", "failed to set Probe type information", "namespace", ns, "err", err)
					return
				}
				probes[k] = probe
			}
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list probes in namespace %s", ns)
		}
	}

	var rejected int
	res := make(map[string]*monitoringv1.Probe, len(probes))

	for probeName, probe := range probes {
		rejectFn := func(probe *monitoringv1.Probe, err error) {
			rejected++
			level.Warn(c.logger).Log(
				"msg", "skipping probe",
				"error", err.Error(),
				"probe", probe,
				"namespace", p.Namespace,
				"prometheus", p.Name,
			)
		}

		if err = probe.Spec.Targets.Validate(); err != nil {
			rejectFn(probe, err)
			continue
		}

		pnKey := fmt.Sprintf("probe/%s/%s", probe.GetNamespace(), probe.GetName())
		if err = store.AddBearerToken(ctx, probe.GetNamespace(), probe.Spec.BearerTokenSecret, pnKey); err != nil {
			rejectFn(probe, err)
			continue
		}

		if err = store.AddBasicAuth(ctx, probe.GetNamespace(), probe.Spec.BasicAuth, pnKey); err != nil {
			rejectFn(probe, err)
			continue
		}

		if probe.Spec.TLSConfig != nil {
			if err = store.AddSafeTLSConfig(ctx, probe.GetNamespace(), &probe.Spec.TLSConfig.SafeTLSConfig); err != nil {
				rejectFn(probe, err)
				continue
			}
		}
		pnAuthKey := fmt.Sprintf("probe/auth/%s/%s", probe.GetNamespace(), probe.GetName())
		if err = store.AddSafeAuthorizationCredentials(ctx, probe.GetNamespace(), probe.Spec.Authorization, pnAuthKey); err != nil {
			rejectFn(probe, err)
			continue
		}

		if err = store.AddOAuth2(ctx, probe.GetNamespace(), probe.Spec.OAuth2, pnKey); err != nil {
			rejectFn(probe, err)
			continue
		}

		if err = validateScrapeIntervalAndTimeout(p, probe.Spec.Interval, probe.Spec.ScrapeTimeout); err != nil {
			rejectFn(probe, err)
			continue
		}

		for _, rl := range probe.Spec.MetricRelabelConfigs {
			if rl.Action != "" {
				if err = validateRelabelConfig(*p, *rl); err != nil {
					rejectFn(probe, err)
					continue
				}
			}
		}
		if err = validateProberURL(probe.Spec.ProberSpec.URL); err != nil {
			err := errors.Wrapf(err, "%s url specified in proberSpec is invalid, it should be of the format `hostname` or `hostname:port`", probe.Spec.ProberSpec.URL)
			rejectFn(probe, err)
			continue
		}
		res[probeName] = probe
	}

	probeKeys := make([]string, 0)
	for k := range res {
		probeKeys = append(probeKeys, k)
	}
	level.Debug(c.logger).Log("msg", "selected Probes", "probes", strings.Join(probeKeys, ","), "namespace", p.Namespace, "prometheus", p.Name)

	if pKey, ok := c.keyFunc(p); ok {
		c.metrics.SetSelectedResources(pKey, monitoringv1.ProbesKind, len(res))
		c.metrics.SetRejectedResources(pKey, monitoringv1.ProbesKind, rejected)
	}

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

// validateRemoteWriteSpec checks that mutually exclusive configurations are not
// included in the Prometheus remoteWrite configuration section.
// Reference:
// https://github.com/prometheus/prometheus/blob/main/docs/configuration/configuration.md#remote_write
func validateRemoteWriteSpec(spec monitoringv1.RemoteWriteSpec) error {
	var nonNilFields []string
	for k, v := range map[string]interface{}{
		"basicAuth":     spec.BasicAuth,
		"oauth2":        spec.OAuth2,
		"authorization": spec.Authorization,
		"sigv4":         spec.Sigv4,
	} {
		if reflect.ValueOf(v).IsNil() {
			continue
		}
		nonNilFields = append(nonNilFields, fmt.Sprintf("%q", k))
	}

	if len(nonNilFields) > 1 {
		return errors.Errorf("%s can't be set at the same time, at most one of them must be defined", strings.Join(nonNilFields, " and "))
	}

	return nil
}

func validateRelabelConfig(p monitoringv1.Prometheus, rc monitoringv1.RelabelConfig) error {
	relabelTarget := regexp.MustCompile(`^(?:(?:[a-zA-Z_]|\$(?:\{\w+\}|\w+))+\w*)+$`)
	promVersion := operator.StringValOrDefault(p.Spec.Version, operator.DefaultPrometheusVersion)
	version, err := semver.ParseTolerant(promVersion)
	if err != nil {
		return errors.Wrap(err, "failed to parse Prometheus version")
	}
	minimumVersion := version.GTE(semver.MustParse("2.36.0"))

	if (rc.Action == string(relabel.Lowercase) || rc.Action == string(relabel.Uppercase)) && !minimumVersion {
		return errors.Errorf("%s relabel action is only supported from Prometheus version 2.36.0", rc.Action)
	}

	if _, err := relabel.NewRegexp(rc.Regex); err != nil {
		return errors.Wrapf(err, "invalid regex %s for relabel configuration", rc.Regex)
	}

	if rc.Modulus == 0 && rc.Action == string(relabel.HashMod) {
		return errors.Errorf("relabel configuration for hashmod requires non-zero modulus")
	}

	if (rc.Action == string(relabel.Replace) || rc.Action == string(relabel.HashMod) || rc.Action == string(relabel.Lowercase) || rc.Action == string(relabel.Uppercase)) && rc.TargetLabel == "" {
		return errors.Errorf("relabel configuration for %s action needs targetLabel value", rc.Action)
	}

	if (rc.Action == string(relabel.Replace) || rc.Action == string(relabel.Lowercase) || rc.Action == string(relabel.Uppercase)) && !relabelTarget.MatchString(rc.TargetLabel) {
		return errors.Errorf("%q is invalid 'target_label' for %s action", rc.TargetLabel, rc.Action)
	}

	if (rc.Action == string(relabel.Lowercase) || rc.Action == string(relabel.Uppercase)) && !(rc.Replacement == relabel.DefaultRelabelConfig.Replacement || rc.Replacement == "") {
		return errors.Errorf("'replacement' can not be set for %s action", rc.Action)
	}

	if rc.Action == string(relabel.LabelMap) {
		if rc.Replacement != "" && !relabelTarget.MatchString(rc.Replacement) {
			return errors.Errorf("%q is invalid 'replacement' for %s action", rc.Replacement, rc.Action)
		}
	}

	if rc.Action == string(relabel.HashMod) && !model.LabelName(rc.TargetLabel).IsValid() {
		return errors.Errorf("%q is invalid 'target_label' for %s action", rc.TargetLabel, rc.Action)
	}

	if rc.Action == string(relabel.LabelDrop) || rc.Action == string(relabel.LabelKeep) {
		if len(rc.SourceLabels) != 0 ||
			!(rc.TargetLabel == "" ||
				rc.TargetLabel == relabel.DefaultRelabelConfig.TargetLabel) ||
			!(rc.Modulus == uint64(0) ||
				rc.Modulus == relabel.DefaultRelabelConfig.Modulus) ||
			!(rc.Separator == "" ||
				rc.Separator == relabel.DefaultRelabelConfig.Separator) ||
			!(rc.Replacement == relabel.DefaultRelabelConfig.Replacement ||
				rc.Replacement == "") {
			return errors.Errorf("%s action requires only 'regex', and no other fields", rc.Action)
		}
	}
	return nil
}

func validateProberURL(url string) error {
	hostPort := strings.Split(url, ":")

	if !govalidator.IsHost(hostPort[0]) {
		return errors.Errorf("invalid host: %q", hostPort[0])
	}

	// handling cases with url specified as host:port
	if len(hostPort) > 1 {
		if !govalidator.IsPort(hostPort[1]) {
			return errors.Errorf("invalid port: %q", hostPort[1])
		}
	}
	return nil
}

func validateScrapeIntervalAndTimeout(p *monitoringv1.Prometheus, scrapeInterval, scrapeTimeout monitoringv1.Duration) error {
	if scrapeTimeout == "" {
		return nil
	}
	if scrapeInterval == "" {
		scrapeInterval = p.Spec.ScrapeInterval
	}
	return operator.CompareScrapeTimeoutToScrapeInterval(scrapeTimeout, scrapeInterval)
}
