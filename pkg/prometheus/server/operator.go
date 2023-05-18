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
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/tools/cache"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/listwatch"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
)

const (
	resyncPeriod = 5 * time.Minute
)

// Operator manages life cycle of Prometheus deployments and
// monitoring configurations.
type Operator struct {
	kclient  kubernetes.Interface
	mdClient metadata.Interface
	mclient  monitoringclient.Interface
	logger   log.Logger
	accessor *operator.Accessor

	nsPromInf cache.SharedIndexInformer
	nsMonInf  cache.SharedIndexInformer

	promInfs  *informers.ForResource
	smonInfs  *informers.ForResource
	pmonInfs  *informers.ForResource
	probeInfs *informers.ForResource
	sconInfs  *informers.ForResource
	ruleInfs  *informers.ForResource
	cmapInfs  *informers.ForResource
	secrInfs  *informers.ForResource
	ssetInfs  *informers.ForResource

	rr *operator.ResourceReconciler

	metrics         *operator.Metrics
	reconciliations *operator.ReconciliationTracker

	nodeAddressLookupErrors prometheus.Counter
	nodeEndpointSyncs       prometheus.Counter
	nodeEndpointSyncErrors  prometheus.Counter

	kubeletObjectName      string
	kubeletObjectNamespace string
	kubeletSyncEnabled     bool
	config                 operator.Config
	endpointSliceSupported bool
	scrapeConfigSupported  bool
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

	mdClient, err := metadata.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating metadata client failed")
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

	// Check prerequisites for ScrapeConfig
	verbs := map[string][]string{
		monitoringv1alpha1.ScrapeConfigName: {"get", "list", "watch"},
	}
	var scrapeConfigSupported bool
	cc, err := k8sutil.NewCRDChecker(conf.Host, conf.TLSInsecure, &conf.TLSConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new CRDChecker object")
	}

	var namespaces = make([]string, 0, len(conf.Namespaces.AllowList))
	for k := range conf.Namespaces.AllowList {
		namespaces = append(namespaces, k)
	}

	err = cc.CheckPrerequisites(ctx,
		namespaces,
		verbs,
		monitoringv1alpha1.SchemeGroupVersion.String(),
		monitoringv1alpha1.ScrapeConfigName)
	switch {
	case errors.Is(err, k8sutil.ErrPrerequiresitesFailed):
		level.Warn(logger).Log("msg", "ScrapeConfig CRD disabled because prerequisites are not met", "err", err)
	case err != nil:
		return nil, errors.Wrap(err, "failed to check prerequisites for the ScrapeConfig CRD ")
	default:
		scrapeConfigSupported = true
	}

	// All the metrics exposed by the controller get the controller="prometheus" label.
	r = prometheus.WrapRegistererWith(prometheus.Labels{"controller": "prometheus"}, r)

	c := &Operator{
		kclient:                client,
		mdClient:               mdClient,
		mclient:                mclient,
		logger:                 logger,
		accessor:               operator.NewAccessor(logger),
		kubeletObjectName:      kubeletObjectName,
		kubeletObjectNamespace: kubeletObjectNamespace,
		kubeletSyncEnabled:     kubeletSyncEnabled,
		config:                 conf,
		metrics:                operator.NewMetrics(r),
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
		scrapeConfigSupported: scrapeConfigSupported,
	}
	c.metrics.MustRegister(
		c.nodeAddressLookupErrors,
		c.nodeEndpointSyncs,
		c.nodeEndpointSyncErrors,
		c.reconciliations,
	)

	c.rr = operator.NewResourceReconciler(
		c.logger,
		c,
		c.metrics,
		monitoringv1.PrometheusesKind,
		r,
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
	c.metrics.MustRegister(prompkg.NewCollectorForStores(promStores...))

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

	if c.scrapeConfigSupported {
		c.sconInfs, err = informers.NewInformersForResource(
			informers.NewMonitoringInformerFactories(
				c.config.Namespaces.AllowList,
				c.config.Namespaces.DenyList,
				mclient,
				resyncPeriod,
				nil,
			),
			monitoringv1alpha1.SchemeGroupVersion.WithResource(monitoringv1alpha1.ScrapeConfigName),
		)
		if err != nil {
			return nil, errors.Wrap(err, "error creating scrapeconfigs informers")
		}
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
		informers.NewMetadataInformerFactory(
			c.config.Namespaces.PrometheusAllowList,
			c.config.Namespaces.DenyList,
			c.mdClient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = prompkg.LabelPrometheusName
			},
		),
		v1.SchemeGroupVersion.WithResource(string(v1.ResourceConfigMaps)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating configmap informers")
	}

	c.secrInfs, err = informers.NewInformersForResource(
		informers.NewMetadataInformerFactory(
			c.config.Namespaces.PrometheusAllowList,
			c.config.Namespaces.DenyList,
			c.mdClient,
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
		{"ScrapeConfig", c.sconInfs},
		{"ConfigMap", c.cmapInfs},
		{"Secret", c.secrInfs},
		{"StatefulSet", c.ssetInfs},
	} {
		// Skipping informers that were not started. If prerequisites for a CRD were not met, their informer will be
		// nil. ScrapeConfig is one example.
		if infs.informersForResource == nil {
			continue
		}

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
	c.promInfs.AddEventHandler(c.rr)

	c.ssetInfs.AddEventHandler(c.rr)

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
	if c.sconInfs != nil {
		c.sconInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleScrapeConfigAdd,
			UpdateFunc: c.handleScrapeConfigUpdate,
			DeleteFunc: c.handleScrapeConfigDelete,
		})
	}
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

	// The controller needs to watch the namespaces in which the service/pod
	// monitors and rules live because a label change on a namespace may
	// trigger a configuration change.
	// It doesn't need to watch on addition/deletion though because it's
	// already covered by the event handlers on service/pod monitors and rules.
	_, _ = c.nsMonInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: c.handleMonitorNamespaceUpdate,
	})
}

// Run the controller.
func (c *Operator) Run(ctx context.Context) error {
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

	go c.rr.Run(ctx)
	defer c.rr.Stop()

	go c.promInfs.Start(ctx.Done())
	go c.smonInfs.Start(ctx.Done())
	go c.pmonInfs.Start(ctx.Done())
	go c.probeInfs.Start(ctx.Done())
	if c.scrapeConfigSupported {
		go c.sconInfs.Start(ctx.Done())
	}
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
		c.RefreshStatusFor(obj.(*monitoringv1.Prometheus))
	})

	c.addHandlers()

	if c.kubeletSyncEnabled {
		go c.reconcileNodeEndpoints(ctx)
	}

	// TODO(simonpasquier): watch for Prometheus pods instead of polling.
	go operator.StatusPoller(ctx, c)

	c.metrics.Ready().Set(1)
	<-ctx.Done()
	return nil
}

// Iterate implements the operator.StatusReconciler interface.
func (c *Operator) Iterate(processFn func(metav1.Object, []monitoringv1.Condition)) {
	if err := c.promInfs.ListAll(labels.Everything(), func(o interface{}) {
		p := o.(*monitoringv1.Prometheus)
		processFn(p, p.Status.Conditions)
	}); err != nil {
		level.Error(c.logger).Log("msg", "failed to list Prometheus objects", "err", err)
	}
}

// RefreshStatus implements the operator.StatusReconciler interface.
func (c *Operator) RefreshStatusFor(o metav1.Object) {
	c.rr.EnqueueForStatus(o)
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
	o, ok := c.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ServiceMonitor added")
		c.metrics.TriggerByCounter(monitoringv1.ServiceMonitorsKind, operator.AddEvent).Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleSmonUpdate(old, cur interface{}) {
	if old.(*monitoringv1.ServiceMonitor).ResourceVersion == cur.(*monitoringv1.ServiceMonitor).ResourceVersion {
		return
	}

	o, ok := c.accessor.ObjectMetadata(cur)
	if ok {
		level.Debug(c.logger).Log("msg", "ServiceMonitor updated")
		c.metrics.TriggerByCounter(monitoringv1.ServiceMonitorsKind, operator.UpdateEvent).Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleSmonDelete(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ServiceMonitor delete")
		c.metrics.TriggerByCounter(monitoringv1.ServiceMonitorsKind, operator.DeleteEvent).Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handlePmonAdd(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PodMonitor added")
		c.metrics.TriggerByCounter(monitoringv1.PodMonitorsKind, operator.AddEvent).Inc()
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handlePmonUpdate(old, cur interface{}) {
	if old.(*monitoringv1.PodMonitor).ResourceVersion == cur.(*monitoringv1.PodMonitor).ResourceVersion {
		return
	}

	o, ok := c.accessor.ObjectMetadata(cur)
	if ok {
		level.Debug(c.logger).Log("msg", "PodMonitor updated")
		c.metrics.TriggerByCounter(monitoringv1.PodMonitorsKind, operator.UpdateEvent).Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handlePmonDelete(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PodMonitor delete")
		c.metrics.TriggerByCounter(monitoringv1.PodMonitorsKind, operator.DeleteEvent).Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleBmonAdd(obj interface{}) {
	if o, ok := c.accessor.ObjectMetadata(obj); ok {
		level.Debug(c.logger).Log("msg", "Probe added")
		c.metrics.TriggerByCounter(monitoringv1.ProbesKind, operator.AddEvent).Inc()
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleBmonUpdate(old, cur interface{}) {
	if old.(*monitoringv1.Probe).ResourceVersion == cur.(*monitoringv1.Probe).ResourceVersion {
		return
	}

	if o, ok := c.accessor.ObjectMetadata(cur); ok {
		level.Debug(c.logger).Log("msg", "Probe updated")
		c.metrics.TriggerByCounter(monitoringv1.ProbesKind, operator.UpdateEvent)
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleBmonDelete(obj interface{}) {
	if o, ok := c.accessor.ObjectMetadata(obj); ok {
		level.Debug(c.logger).Log("msg", "Probe delete")
		c.metrics.TriggerByCounter(monitoringv1.ProbesKind, operator.DeleteEvent).Inc()
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleScrapeConfigAdd(obj interface{}) {
	if o, ok := c.accessor.ObjectMetadata(obj); ok {
		level.Debug(c.logger).Log("msg", "ScrapeConfig added")
		c.metrics.TriggerByCounter(monitoringv1alpha1.ScrapeConfigsKind, operator.AddEvent).Inc()
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleScrapeConfigUpdate(old, cur interface{}) {
	if old.(*monitoringv1alpha1.ScrapeConfig).ResourceVersion == cur.(*monitoringv1alpha1.ScrapeConfig).ResourceVersion {
		return
	}

	if o, ok := c.accessor.ObjectMetadata(cur); ok {
		level.Debug(c.logger).Log("msg", "ScrapeConfig updated")
		c.metrics.TriggerByCounter(monitoringv1alpha1.ScrapeConfigsKind, operator.UpdateEvent)
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleScrapeConfigDelete(obj interface{}) {
	if o, ok := c.accessor.ObjectMetadata(obj); ok {
		level.Debug(c.logger).Log("msg", "ScrapeConfig deleted")
		c.metrics.TriggerByCounter(monitoringv1alpha1.ScrapeConfigsKind, operator.DeleteEvent).Inc()
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleRuleAdd(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PrometheusRule added")
		c.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, operator.AddEvent).Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleRuleUpdate(old, cur interface{}) {
	if old.(*monitoringv1.PrometheusRule).ResourceVersion == cur.(*monitoringv1.PrometheusRule).ResourceVersion {
		return
	}

	o, ok := c.accessor.ObjectMetadata(cur)
	if ok {
		level.Debug(c.logger).Log("msg", "PrometheusRule updated")
		c.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, operator.UpdateEvent).Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleRuleDelete(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PrometheusRule deleted")
		c.metrics.TriggerByCounter(monitoringv1.PrometheusRuleKind, operator.DeleteEvent).Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Do we need to enqueue secrets just for the namespace or in general?
func (c *Operator) handleSecretDelete(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if !ok {
		return
	}
	level.Debug(c.logger).Log("msg", "Secret deleted")
	c.metrics.TriggerByCounter("Secret", operator.DeleteEvent).Inc()
	c.enqueueForPrometheusNamespace(o.GetNamespace())
}

func (c *Operator) handleSecretUpdate(old, cur interface{}) {
	oldObj, ok := c.accessor.ObjectMetadata(old)
	if !ok {
		return
	}

	curObj, ok := c.accessor.ObjectMetadata(cur)
	if !ok {
		return
	}

	if oldObj.GetResourceVersion() == curObj.GetResourceVersion() {
		return
	}

	level.Debug(c.logger).Log("msg", "Secret updated")
	c.metrics.TriggerByCounter("Secret", operator.UpdateEvent).Inc()
	c.enqueueForPrometheusNamespace(curObj.GetNamespace())
}

func (c *Operator) handleSecretAdd(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if !ok {
		return
	}

	level.Debug(c.logger).Log("msg", "Secret added")
	c.metrics.TriggerByCounter("Secret", operator.AddEvent).Inc()
	c.enqueueForPrometheusNamespace(o.GetNamespace())
}

// TODO: Do we need to enqueue configmaps just for the namespace or in general?
func (c *Operator) handleConfigMapAdd(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if !ok {
		return
	}

	level.Debug(c.logger).Log("msg", "ConfigMap added")
	c.metrics.TriggerByCounter("ConfigMap", operator.AddEvent).Inc()
	c.enqueueForPrometheusNamespace(o.GetNamespace())
}

func (c *Operator) handleConfigMapDelete(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if !ok {
		return
	}

	level.Debug(c.logger).Log("msg", "ConfigMap deleted")
	c.metrics.TriggerByCounter("ConfigMap", operator.DeleteEvent).Inc()
	c.enqueueForPrometheusNamespace(o.GetNamespace())
}

func (c *Operator) handleConfigMapUpdate(old, cur interface{}) {
	oldObj, ok := c.accessor.ObjectMetadata(old)
	if !ok {
		return
	}

	curObj, ok := c.accessor.ObjectMetadata(cur)
	if !ok {
		return
	}

	if oldObj.GetResourceVersion() == curObj.GetResourceVersion() {
		return
	}

	level.Debug(c.logger).Log("msg", "ConfigMap updated")
	c.metrics.TriggerByCounter("ConfigMap", operator.UpdateEvent).Inc()
	c.enqueueForPrometheusNamespace(curObj.GetNamespace())
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
			c.rr.EnqueueForReconciliation(p)
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
			c.rr.EnqueueForReconciliation(p)
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
			c.rr.EnqueueForReconciliation(p)
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
			c.rr.EnqueueForReconciliation(p)
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
			c.rr.EnqueueForReconciliation(p)
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

// Resolve implements the operator.Syncer interface.
func (c *Operator) Resolve(ss *appsv1.StatefulSet) metav1.Object {
	key, ok := c.accessor.MetaNamespaceKey(ss)
	if !ok {
		return nil
	}

	match, promKey := prompkg.StatefulSetKeyToPrometheusKey(key)
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
	c.metrics.TriggerByCounter("Namespace", operator.UpdateEvent).Inc()

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
				c.rr.EnqueueForReconciliation(p)
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

// Sync implements the operator.Syncer interface.
func (c *Operator) Sync(ctx context.Context, key string) error {
	err := c.sync(ctx, key)
	c.reconciliations.SetStatus(key, err)

	return err
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
	logDeprecatedFields(logger, p)

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
	cg, err := prompkg.NewConfigGenerator(c.logger, p, c.endpointSliceSupported)
	if err != nil {
		return err
	}

	if err := c.createOrUpdateConfigurationSecret(ctx, p, cg, ruleConfigMapNames, assetStore); err != nil {
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
	expected := prompkg.ExpectedStatefulSetShardNames(p)
	for shard, ssetName := range expected {
		logger := log.With(logger, "statefulset", ssetName, "shard", fmt.Sprintf("%d", shard))
		level.Debug(logger).Log("msg", "reconciling statefulset")

		obj, err := c.ssetInfs.Get(prompkg.KeyToStatefulSetKey(p, key, shard))
		exists := !apierrors.IsNotFound(err)
		if err != nil && !apierrors.IsNotFound(err) {
			return errors.Wrap(err, "retrieving statefulset failed")
		}

		existingStatefulSet := &appsv1.StatefulSet{}
		if obj != nil {
			existingStatefulSet = obj.(*appsv1.StatefulSet)
			if c.rr.DeletionInProgress(existingStatefulSet) {
				// We want to avoid entering a hot-loop of update/delete cycles
				// here since the sts was marked for deletion in foreground,
				// which means it may take some time before the finalizers
				// complete and the resource disappears from the API. The
				// deletion timestamp will have been set when the initial
				// delete request was issued. In that case, we avoid further
				// processing.
				continue
			}
		}

		newSSetInputHash, err := createSSetInputHash(*p, c.config, ruleConfigMapNames, tlsAssets, existingStatefulSet.Spec)
		if err != nil {
			return err
		}

		sset, err := makeStatefulSet(
			logger,
			ssetName,
			p,
			p.Spec.BaseImage, p.Spec.Tag, p.Spec.SHA,
			p.Spec.Retention,
			p.Spec.RetentionSize,
			p.Spec.Rules,
			p.Spec.Query,
			p.Spec.AllowOverlappingBlocks,
			p.Spec.EnableAdminAPI,
			p.Spec.QueryLogFile,
			p.Spec.Thanos,
			p.Spec.DisableCompaction,
			&c.config,
			cg,
			ruleConfigMapNames,
			newSSetInputHash,
			int32(shard),
			tlsAssets.ShardNames())
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

		if newSSetInputHash == existingStatefulSet.ObjectMeta.Annotations[prompkg.SSetInputHashName] {
			level.Debug(logger).Log("msg", "new statefulset generation inputs match current, skipping any actions")
			continue
		}

		level.Debug(logger).Log(
			"msg", "updating current statefulset because of hash divergence",
			"new_hash", newSSetInputHash,
			"existing_hash", existingStatefulSet.ObjectMeta.Annotations[prompkg.SSetInputHashName],
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

	err = c.ssetInfs.ListAllByNamespace(p.Namespace, labels.SelectorFromSet(labels.Set{prompkg.PrometheusNameLabelName: p.Name, prompkg.PrometheusModeLabeLName: prometheusMode}), func(obj interface{}) {
		s := obj.(*appsv1.StatefulSet)

		if _, ok := ssets[s.Name]; ok {
			// Do not delete statefulsets that we still expect to exist. This
			// is to cleanup StatefulSets when shards are reduced.
			return
		}

		if c.rr.DeletionInProgress(s) {
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

// UpdateStatus updates the status subresource of the object identified by the given
// key.
// UpdateStatus implements the operator.Syncer interface.
func (c *Operator) UpdateStatus(ctx context.Context, key string) error {
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
		availableCondition = monitoringv1.Condition{
			Type:   monitoringv1.Available,
			Status: monitoringv1.ConditionTrue,
			LastTransitionTime: metav1.Time{
				Time: time.Now().UTC(),
			},
			ObservedGeneration: p.Generation,
		}
		messages []string
		replicas = 1
	)

	if p.Spec.Replicas != nil {
		replicas = int(*p.Spec.Replicas)
	}

	for shard := range prompkg.ExpectedStatefulSetShardNames(p) {
		ssetName := prompkg.KeyToStatefulSetKey(p, key, shard)
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
		if c.rr.DeletionInProgress(sset) {
			continue
		}

		stsReporter, err := operator.NewStatefulSetReporter(ctx, c.kclient, sset)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve statefulset state")
		}

		pStatus.Replicas += int32(len(stsReporter.Pods))
		pStatus.UpdatedReplicas += int32(len(stsReporter.UpdatedPods()))
		pStatus.AvailableReplicas += int32(len(stsReporter.ReadyPods()))
		pStatus.UnavailableReplicas += int32(len(stsReporter.Pods) - len(stsReporter.ReadyPods()))

		pStatus.ShardStatuses = append(
			pStatus.ShardStatuses,
			monitoringv1.ShardStatus{
				ShardID:             strconv.Itoa(shard),
				Replicas:            int32(len(stsReporter.Pods)),
				UpdatedReplicas:     int32(len(stsReporter.UpdatedPods())),
				AvailableReplicas:   int32(len(stsReporter.ReadyPods())),
				UnavailableReplicas: int32(len(stsReporter.Pods) - len(stsReporter.ReadyPods())),
			},
		)

		if len(stsReporter.ReadyPods()) >= replicas {
			// All pods are ready (or the desired number of replicas is zero).
			continue
		}

		if len(stsReporter.ReadyPods()) == 0 {
			availableCondition.Reason = "NoPodReady"
			availableCondition.Status = monitoringv1.ConditionFalse
		} else if availableCondition.Status != monitoringv1.ConditionFalse {
			availableCondition.Reason = "SomePodsNotReady"
			availableCondition.Status = monitoringv1.ConditionDegraded
		}

		for _, p := range stsReporter.Pods {
			if m := p.Message(); m != "" {
				messages = append(messages, fmt.Sprintf("shard %d: pod %s: %s", shard, p.Name, m))
			}
		}
	}

	availableCondition.Message = strings.Join(messages, "\n")

	reconciledCondition := c.reconciliations.GetCondition(key, p.Generation)
	pStatus.Conditions = operator.UpdateConditions(pStatus.Conditions, availableCondition, reconciledCondition)

	p.Status = pStatus
	if _, err = c.mclient.MonitoringV1().Prometheuses(p.Namespace).UpdateStatus(ctx, p, metav1.UpdateOptions{}); err != nil {
		return errors.Wrap(err, "failed to update status subresource")
	}

	return nil
}

func logDeprecatedFields(logger log.Logger, p *monitoringv1.Prometheus) {
	deprecationWarningf := "field %q is deprecated, field %q should be used instead"
	if p.Spec.BaseImage != "" {
		level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, "spec.baseImage", "spec.image"))
	}

	if p.Spec.Tag != "" {
		level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, "spec.tag", "spec.image"))
	}

	if p.Spec.SHA != "" {
		level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, "spec.sha", "spec.image"))
	}

	if p.Spec.Thanos != nil {
		if p.Spec.BaseImage != "" {
			level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, "spec.thanos.baseImage", "spec.thanos.image"))
		}
		if p.Spec.Tag != "" {
			level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, "spec.thanos.tag", "spec.thanos.image"))
		}
		if p.Spec.SHA != "" {
			level.Warn(logger).Log("msg", fmt.Sprintf(deprecationWarningf, "spec.thanos.sha", "spec.thanos.image"))
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

func (c *Operator) createOrUpdateConfigurationSecret(ctx context.Context, p *monitoringv1.Prometheus, cg *prompkg.ConfigGenerator, ruleConfigMapNames []string, store *assets.Store) error {
	// If no service or pod monitor selectors are configured, the user wants to
	// manage configuration themselves. Do create an empty Secret if it doesn't
	// exist.
	if p.Spec.ServiceMonitorSelector == nil && p.Spec.PodMonitorSelector == nil &&
		p.Spec.ProbeSelector == nil {
		level.Debug(c.logger).Log("msg", "neither ServiceMonitor nor PodMonitor, nor Probe selector specified, leaving configuration unmanaged", "prometheus", p.Name, "namespace", p.Namespace)

		s, err := prompkg.MakeEmptyConfigurationSecret(p, c.config)
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

	resourceSelector := prompkg.NewResourceSelector(c.logger, p, store, c.nsMonInf, c.metrics)

	smons, err := resourceSelector.SelectServiceMonitors(ctx, c.smonInfs.ListAllByNamespace)
	if err != nil {
		return errors.Wrap(err, "selecting ServiceMonitors failed")
	}

	pmons, err := resourceSelector.SelectPodMonitors(ctx, c.pmonInfs.ListAllByNamespace)
	if err != nil {
		return errors.Wrap(err, "selecting PodMonitors failed")
	}

	bmons, err := resourceSelector.SelectProbes(ctx, c.probeInfs.ListAllByNamespace)
	if err != nil {
		return errors.Wrap(err, "selecting Probes failed")
	}

	var scrapeConfigs map[string]*monitoringv1alpha1.ScrapeConfig
	if c.sconInfs != nil {
		scrapeConfigs, err = resourceSelector.SelectScrapeConfigs(c.sconInfs.ListAllByNamespace)
		if err != nil {
			return errors.Wrap(err, "selecting ScrapeConfigs failed")
		}
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
		if err := prompkg.ValidateRemoteWriteSpec(remote); err != nil {
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
			if err := store.AddBasicAuth(ctx, p.GetNamespace(), am.BasicAuth, fmt.Sprintf("alertmanager/auth/%d", i)); err != nil {
				return errors.Wrapf(err, "alerting")
			}
			if err := store.AddSafeAuthorizationCredentials(ctx, p.GetNamespace(), am.Authorization, fmt.Sprintf("alertmanager/auth/%d", i)); err != nil {
				return errors.Wrapf(err, "alerting")
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

	// Update secret based on the most recent configuration.
	conf, err := cg.GenerateServerConfiguration(
		p.Spec.EvaluationInterval,
		p.Spec.QueryLogFile,
		p.Spec.RuleSelector,
		p.Spec.Exemplars,
		p.Spec.TSDB,
		p.Spec.Alerting,
		p.Spec.RemoteRead,
		smons,
		pmons,
		bmons,
		scrapeConfigs,
		store,
		additionalScrapeConfigs,
		additionalAlertRelabelConfigs,
		additionalAlertManagerConfigs,
		ruleConfigMapNames,
	)
	if err != nil {
		return errors.Wrap(err, "generating config failed")
	}

	s := prompkg.MakeConfigSecret(p, c.config)
	s.ObjectMeta.Annotations = map[string]string{
		"generated": "true",
	}

	// Compress config to avoid 1mb secret limit for a while
	var buf bytes.Buffer
	if err = operator.GzipConfig(&buf, conf); err != nil {
		return errors.Wrap(err, "couldn't gzip config")
	}
	s.Data[prompkg.ConfigFilename] = buf.Bytes()

	level.Debug(c.logger).Log("msg", "updating Prometheus configuration secret")

	return k8sutil.CreateOrUpdateSecret(ctx, sClient, s)
}

func (c *Operator) createOrUpdateTLSAssetSecrets(ctx context.Context, p *monitoringv1.Prometheus, store *assets.Store) (*operator.ShardedSecret, error) {
	labels := c.config.Labels.Merge(prompkg.ManagedByOperatorLabels)
	template := prompkg.NewTLSAssetSecret(p, labels)

	sSecret := operator.NewShardedSecret(template, prompkg.TLSAssetsSecretName(p))

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

func (c *Operator) createOrUpdateWebConfigSecret(ctx context.Context, p *monitoringv1.Prometheus) error {
	boolTrue := true

	var fields monitoringv1.WebConfigFileFields
	if p.Spec.Web != nil {
		fields = p.Spec.Web.WebConfigFileFields
	}

	webConfig, err := webconfig.New(
		prompkg.WebConfigDir,
		prompkg.WebConfigSecretName(p),
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
	secretLabels := c.config.Labels.Merge(prompkg.ManagedByOperatorLabels)

	if err := webConfig.CreateOrUpdateWebConfigSecret(ctx, secretClient, secretLabels, ownerReference); err != nil {
		return errors.Wrap(err, "failed to reconcile web config secret")
	}

	return nil
}
