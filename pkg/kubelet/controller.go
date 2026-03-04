// Copyright 2023 The prometheus-operator Authors
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

package kubelet

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"

	"github.com/prometheus-operator/prometheus-operator/pkg/k8s"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

const (
	applicationNameLabelValue = "kubelet"

	maxEndpointsPerSlice = 512

	endpointsLabel     = "endpoints"
	endpointSliceLabel = "endpointslice"

	httpsPort        = int32(10250)
	httpsPortName    = "https-metrics"
	httpPort         = int32(10255)
	httpPortName     = "http-metrics"
	cAdvisorPort     = int32(4194)
	cAdvisorPortName = "cadvisor"
)

type Controller struct {
	logger *slog.Logger

	kclient kubernetes.Interface

	nodeAddressLookupErrors prometheus.Counter
	nodeEndpointSyncs       *prometheus.CounterVec
	nodeEndpointSyncErrors  *prometheus.CounterVec

	kubeletObjectName      string
	kubeletObjectNamespace string
	kubeletSelector        string

	annotations operator.Map
	labels      operator.Map

	nodeAddressPriority  string
	maxEndpointsPerSlice int

	manageEndpointSlice bool
	manageEndpoints     bool
	syncPeriod          time.Duration

	// httpMetricsEnabled controls whether to include the insecure HTTP metrics
	// port (10255) in the kubelet Service. Set to false when the cluster has
	// disabled the insecure kubelet read-only port (e.g., GKE 1.32+).
	httpMetricsEnabled bool
}

type ControllerOption func(*Controller)

func WithEndpointSlice() ControllerOption {
	return func(c *Controller) {
		c.manageEndpointSlice = true
	}
}

func WithMaxEndpointsPerSlice(v int) ControllerOption {
	return func(c *Controller) {
		c.maxEndpointsPerSlice = v
	}
}

func WithEndpoints() ControllerOption {
	return func(c *Controller) {
		c.manageEndpoints = true
	}
}

func WithNodeAddressPriority(s string) ControllerOption {
	return func(c *Controller) {
		c.nodeAddressPriority = s
	}
}

func WithSyncPeriod(d time.Duration) ControllerOption {
	return func(c *Controller) {
		c.syncPeriod = d
	}
}

// WithHTTPMetrics controls whether to include the insecure HTTP metrics port
// (10255) in the kubelet Service. When disabled, only the secure HTTPS port
// (10250) and cAdvisor port (4194) are included. This is useful when the
// cluster has disabled the insecure kubelet read-only port (e.g., GKE 1.32+).
func WithHTTPMetrics(enabled bool) ControllerOption {
	return func(c *Controller) {
		c.httpMetricsEnabled = enabled
	}
}

func New(
	logger *slog.Logger,
	kclient kubernetes.Interface,
	r prometheus.Registerer,
	kubeletServiceName string,
	kubeletServiceNamespace string,
	kubeletSelector operator.LabelSelector,
	commonAnnotations operator.Map,
	commonLabels operator.Map,
	opts ...ControllerOption,
) (*Controller, error) {
	c := &Controller{
		kclient: kclient,

		nodeAddressLookupErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prometheus_operator_node_address_lookup_errors_total",
			Help: "Number of times a node IP address could not be determined",
		}),
		nodeEndpointSyncs: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "prometheus_operator_node_syncs_total",
				Help: "Total number of synchronisations for the given resource",
			},
			[]string{"resource"},
		),
		nodeEndpointSyncErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "prometheus_operator_node_syncs_failed_total",
				Help: "Total number of failed synchronisations for the given resource",
			},
			[]string{"resource"},
		),

		kubeletObjectName:      kubeletServiceName,
		kubeletObjectNamespace: kubeletServiceNamespace,
		kubeletSelector:        kubeletSelector.String(),
		maxEndpointsPerSlice:   maxEndpointsPerSlice,

		annotations: commonAnnotations,
		labels:      commonLabels,
	}

	for _, opt := range opts {
		opt(c)
	}

	if !c.manageEndpoints && !c.manageEndpointSlice {
		return nil, fmt.Errorf("at least one of endpoints or endpointslice needs to be enabled")
	}

	for _, v := range []string{
		endpointsLabel,
		endpointSliceLabel,
	} {
		c.nodeEndpointSyncs.WithLabelValues(v)
		c.nodeEndpointSyncErrors.WithLabelValues(v)
	}

	if r == nil {
		r = prometheus.NewRegistry()
	}
	r.MustRegister(
		c.nodeAddressLookupErrors,
		c.nodeEndpointSyncs,
		c.nodeEndpointSyncErrors,
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "prometheus_operator_kubelet_managed_resource",
				Help: "",
				ConstLabels: prometheus.Labels{
					"resource": endpointsLabel,
				},
			},
			func() float64 {
				if c.manageEndpoints {
					return 1.0
				}
				return 0.0
			},
		),
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "prometheus_operator_kubelet_managed_resource",
				Help: "",
				ConstLabels: prometheus.Labels{
					"resource": endpointSliceLabel,
				},
			},
			func() float64 {
				if c.manageEndpointSlice {
					return 1.0
				}
				return 0.0
			},
		),
	)

	c.logger = logger.With("kubelet_object", fmt.Sprintf("%s/%s", c.kubeletObjectNamespace, c.kubeletObjectName))

	return c, nil
}

func (c *Controller) Run(ctx context.Context) error {
	c.logger.Info("Starting controller")

	ticker := time.NewTicker(c.syncPeriod)
	defer ticker.Stop()
	for {
		c.sync(ctx)

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// nodeAddress returns the provided node's address, based on the priority:
// 1. NodeInternalIP
// 2. NodeExternalIP
//
// Copied from github.com/prometheus/prometheus/discovery/kubernetes/node.go.
func (c *Controller) nodeAddress(node v1.Node) (string, map[v1.NodeAddressType][]string, error) {
	m := map[v1.NodeAddressType][]string{}
	for _, a := range node.Status.Addresses {
		m[a.Type] = append(m[a.Type], a.Address)
	}

	switch c.nodeAddressPriority {
	case "internal":
		if addresses, ok := m[v1.NodeInternalIP]; ok {
			return addresses[0], m, nil
		}
		if addresses, ok := m[v1.NodeExternalIP]; ok {
			return addresses[0], m, nil
		}
	case "external":
		if addresses, ok := m[v1.NodeExternalIP]; ok {
			return addresses[0], m, nil
		}
		if addresses, ok := m[v1.NodeInternalIP]; ok {
			return addresses[0], m, nil
		}
	}

	return "", m, fmt.Errorf("host address unknown")
}

// nodeReadyConditionKnown checks the node for a known Ready condition. If the
// condition is Unknown then that node's kubelet has not recently sent any node
// status, so we should not add this node to the kubelet endpoint and scrape
// it.
func nodeReadyConditionKnown(node v1.Node) bool {
	for _, c := range node.Status.Conditions {
		if c.Type == v1.NodeReady && c.Status != v1.ConditionUnknown {
			return true
		}
	}
	return false
}

type nodeAddress struct {
	apiVersion string
	ipAddress  string
	name       string
	uid        types.UID
	ipv4       bool
	ready      bool
}

func (na *nodeAddress) discoveryV1Endpoint() discoveryv1.Endpoint {
	return discoveryv1.Endpoint{
		Addresses: []string{na.ipAddress},
		Conditions: discoveryv1.EndpointConditions{
			Ready: ptr.To(true),
		},
		NodeName: ptr.To(na.name),
		TargetRef: &v1.ObjectReference{
			Kind:       "Node",
			Name:       na.name,
			UID:        na.uid,
			APIVersion: na.apiVersion,
		},
	}
}

func (na *nodeAddress) v1EndpointAddress() v1.EndpointAddress {
	return v1.EndpointAddress{
		IP:       na.ipAddress,
		NodeName: ptr.To(na.name),
		TargetRef: &v1.ObjectReference{
			Kind:       "Node",
			Name:       na.name,
			UID:        na.uid,
			APIVersion: na.apiVersion,
		},
	}
}

func (c *Controller) getNodeAddresses(nodes []v1.Node) ([]nodeAddress, []error) {
	var (
		addresses         = make([]nodeAddress, 0, len(nodes))
		readyKnownNodes   = map[string]string{}
		readyUnknownNodes = map[string]string{}

		errs []error
	)

	for _, n := range nodes {
		address, _, err := c.nodeAddress(n)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to determine hostname for node %q (priority: %s): %w", n.Name, c.nodeAddressPriority, err))
			continue
		}

		ip := net.ParseIP(address)
		if ip == nil {
			errs = append(errs, fmt.Errorf("failed to parse IP address %q for node %q (priority: %s): %w", address, n.Name, c.nodeAddressPriority, err))
			continue
		}

		na := nodeAddress{
			ipAddress:  address,
			name:       n.Name,
			uid:        n.UID,
			apiVersion: n.APIVersion,
			ipv4:       ip.To4() != nil,
			ready:      nodeReadyConditionKnown(n),
		}
		addresses = append(addresses, na)

		if !na.ready {
			c.logger.Info("Node Ready condition is Unknown", "node", n.GetName())
			readyUnknownNodes[address] = n.Name
			continue
		}

		readyKnownNodes[address] = n.Name
	}

	// We want to remove any nodes that have an unknown ready state *and* a
	// duplicate IP address. If this is the case, we want to keep just the node
	// with the duplicate IP address that has a known ready state. This also
	// ensures that order of addresses are preserved.
	addressesFinal := make([]nodeAddress, 0)
	for _, address := range addresses {
		knownNodeName, foundKnown := readyKnownNodes[address.ipAddress]
		_, foundUnknown := readyUnknownNodes[address.ipAddress]
		if foundKnown && foundUnknown && address.name != knownNodeName {
			continue
		}

		addressesFinal = append(addressesFinal, address)
	}

	return addressesFinal, errs
}

func (c *Controller) sync(ctx context.Context) {
	c.logger.Debug("Synchronizing nodes")

	//TODO(simonpasquier): add failed/attempted counters.
	nodeList, err := c.kclient.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: c.kubeletSelector})
	if err != nil {
		c.logger.Error("Failed to list nodes", "err", err)
		return
	}

	// Sort the nodes slice by their name.
	nodes := nodeList.Items
	slices.SortStableFunc(nodes, func(a, b v1.Node) int {
		return strings.Compare(a.Name, b.Name)
	})
	c.logger.Debug("Nodes retrieved from the Kubernetes API", "num_nodes", len(nodes))

	addresses, errs := c.getNodeAddresses(nodes)
	if len(errs) > 0 {
		for _, err := range errs {
			c.logger.Warn(err.Error())
		}
		c.nodeAddressLookupErrors.Add(float64(len(errs)))
	}
	c.logger.Debug("Nodes converted to endpoint addresses", "num_addresses", len(addresses))

	svc, err := c.syncService(ctx)
	if err != nil {
		c.logger.Error("Failed to synchronize kubelet service", "err", err)
	}

	if c.manageEndpoints {
		c.nodeEndpointSyncs.WithLabelValues(endpointsLabel).Inc()
		if err = c.syncEndpoints(ctx, addresses); err != nil {
			c.nodeEndpointSyncErrors.WithLabelValues(endpointsLabel).Inc()
			c.logger.Error("Failed to synchronize kubelet endpoints", "err", err)
		}
	}

	if c.manageEndpointSlice {
		c.nodeEndpointSyncs.WithLabelValues(endpointSliceLabel).Inc()
		if err = c.syncEndpointSlice(ctx, svc, addresses); err != nil {
			c.nodeEndpointSyncErrors.WithLabelValues(endpointSliceLabel).Inc()
			c.logger.Error("Failed to synchronize kubelet endpointslice", "err", err)
		}
	}
}

func (c *Controller) syncEndpoints(ctx context.Context, addresses []nodeAddress) error {
	c.logger.Debug("Sync endpoints")

	//nolint:staticcheck // Ignore SA1019 Endpoints is marked as deprecated.
	eps := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.kubeletObjectName,
			Annotations: c.annotations,
			Labels: c.labels.Merge(map[string]string{
				"k8s-app":                        applicationNameLabelValue,
				operator.ApplicationNameLabelKey: applicationNameLabelValue,
				operator.ManagedByLabelKey:       operator.ManagedByLabelValue,
			}),
		},
		//nolint:staticcheck // Ignore SA1019 Endpoints is marked as deprecated.
		Subsets: []v1.EndpointSubset{
			{
				Addresses: make([]v1.EndpointAddress, len(addresses)),
				Ports:     c.endpointPorts(),
			},
		},
	}

	if c.manageEndpointSlice {
		// Tell the endpointslice mirroring controller that it shouldn't manage
		// the endpoints object since this controller is in charge.
		eps.Labels[discoveryv1.LabelSkipMirror] = "true"
	}

	for i, na := range addresses {
		eps.Subsets[0].Addresses[i] = na.v1EndpointAddress()
	}

	c.logger.Debug("Updating Kubernetes endpoint")
	err := k8s.CreateOrUpdateEndpoints(ctx, c.kclient.CoreV1().Endpoints(c.kubeletObjectNamespace), eps)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) syncService(ctx context.Context) (*v1.Service, error) {
	c.logger.Debug("Sync service")

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.kubeletObjectName,
			Annotations: c.annotations,
			Labels: c.labels.Merge(map[string]string{
				"k8s-app":                        applicationNameLabelValue,
				operator.ApplicationNameLabelKey: applicationNameLabelValue,
				operator.ManagedByLabelKey:       operator.ManagedByLabelValue,
			}),
		},
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: v1.ClusterIPNone,
			Ports:     c.servicePorts(),
		},
	}

	c.logger.Debug("Updating Kubernetes service", "service", c.kubeletObjectName)
	return k8s.CreateOrUpdateService(ctx, c.kclient.CoreV1().Services(c.kubeletObjectNamespace), svc)
}

func (c *Controller) syncEndpointSlice(ctx context.Context, svc *v1.Service, addresses []nodeAddress) error {
	c.logger.Debug("Sync endpointslice")

	// Get the list of endpointslice objects associated to the service.
	client := c.kclient.DiscoveryV1().EndpointSlices(c.kubeletObjectNamespace)
	l, err := client.List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set{discoveryv1.LabelServiceName: c.kubeletObjectName}.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to list endpointslice: %w", err)
	}

	epsl := []discoveryv1.EndpointSlice{}
	if len(l.Items) > 0 {
		epsl = l.Items
	}

	nodeAddressIdx := make(map[string]nodeAddress, len(addresses))
	for _, a := range addresses {
		nodeAddressIdx[a.ipAddress] = a
	}

	// Iterate over the existing endpoints to update their state or remove them
	// if the IP address isn't associated to a node anymore.
	for i, eps := range epsl {
		endpoints := make([]discoveryv1.Endpoint, 0, len(eps.Endpoints))
		for _, ep := range eps.Endpoints {
			if len(ep.Addresses) != 1 {
				c.logger.Warn("Got more than 1 address for the endpoint", "name", eps.Name, "num", len(ep.Addresses))
				continue
			}

			a, found := nodeAddressIdx[ep.Addresses[0]]
			if !found {
				// The node doesn't exist anymore.
				continue
			}

			endpoints = append(endpoints, a.discoveryV1Endpoint())
			delete(nodeAddressIdx, a.ipAddress)
		}

		epsl[i].Endpoints = endpoints
	}

	// Append new nodes into the existing endpointslices.
	for _, a := range addresses {
		if _, found := nodeAddressIdx[a.ipAddress]; !found {
			// Already processed.
			continue
		}

		for i := range epsl {
			if a.ipv4 != (epsl[i].AddressType == discoveryv1.AddressTypeIPv4) {
				// Not the same address type.
				continue
			}

			if len(epsl[i].Endpoints) >= c.maxEndpointsPerSlice {
				// The endpoints slice is full.
				continue
			}

			epsl[i].Endpoints = append(epsl[i].Endpoints, a.discoveryV1Endpoint())
			delete(nodeAddressIdx, a.ipAddress)

			break
		}
	}

	// Create new endpointslice object(s) for the new nodes which couldn't be
	// appended to the existing endpointslices.
	var (
		ipv4Eps *discoveryv1.EndpointSlice
		ipv6Eps *discoveryv1.EndpointSlice
	)
	for _, a := range addresses {
		if _, found := nodeAddressIdx[a.ipAddress]; !found {
			// Already processed.
			continue
		}

		if ipv4Eps != nil && c.fullCapacity(ipv4Eps.Endpoints) {
			epsl = append(epsl, *ipv4Eps)
			ipv4Eps = nil
		}

		if ipv6Eps != nil && c.fullCapacity(ipv6Eps.Endpoints) {
			epsl = append(epsl, *ipv6Eps)
			ipv6Eps = nil
		}

		eps := ipv4Eps
		if !a.ipv4 {
			eps = ipv6Eps
		}

		if eps == nil {
			eps = &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: c.kubeletObjectName + "-",
					Annotations:  c.annotations,
					Labels: c.labels.Merge(map[string]string{
						discoveryv1.LabelServiceName:     c.kubeletObjectName,
						discoveryv1.LabelManagedBy:       operator.ManagedByLabelValue,
						"k8s-app":                        applicationNameLabelValue,
						operator.ApplicationNameLabelKey: applicationNameLabelValue,
						operator.ManagedByLabelKey:       operator.ManagedByLabelValue,
					}),
					OwnerReferences: []metav1.OwnerReference{{
						APIVersion:         "v1",
						BlockOwnerDeletion: ptr.To(true),
						Controller:         ptr.To(true),
						Kind:               "Service",
						Name:               c.kubeletObjectName,
						UID:                svc.UID,
					},
					},
				},
				Ports: c.endpointSlicePorts(),
			}

			if a.ipv4 {
				eps.AddressType = discoveryv1.AddressTypeIPv4
				ipv4Eps = eps
			} else {
				eps.AddressType = discoveryv1.AddressTypeIPv6
				ipv6Eps = eps
			}
		}

		eps.Endpoints = append(eps.Endpoints, a.discoveryV1Endpoint())
		delete(nodeAddressIdx, a.ipAddress)
	}

	if ipv4Eps != nil {
		epsl = append(epsl, *ipv4Eps)
	}

	if ipv6Eps != nil {
		epsl = append(epsl, *ipv6Eps)
	}

	for _, eps := range epsl {
		if len(eps.Endpoints) == 0 {
			fmt.Println("delete")
			c.logger.Debug("Deleting endpointslice object", "name", eps.Name)
			err := client.Delete(ctx, eps.Name, metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("failed to delete endpoinslice: %w", err)
			}

			continue
		}

		c.logger.Debug("Updating endpointslice object", "name", eps.Name)
		err := k8s.CreateOrUpdateEndpointSlice(ctx, client, &eps)
		if err != nil {
			return fmt.Errorf("failed to update endpoinslice: %w", err)
		}
	}

	return nil
}

func (c *Controller) fullCapacity(eps []discoveryv1.Endpoint) bool {
	return len(eps) >= c.maxEndpointsPerSlice
}

// servicePorts returns the list of ServicePort for the kubelet Service.
// If httpMetricsEnabled is false, the insecure HTTP port (10255) is excluded.
func (c *Controller) servicePorts() []v1.ServicePort {
	ports := []v1.ServicePort{
		{
			Name: httpsPortName,
			Port: httpsPort,
		},
		{
			Name: cAdvisorPortName,
			Port: cAdvisorPort,
		},
	}

	if c.httpMetricsEnabled {
		ports = append(ports, v1.ServicePort{
			Name: httpPortName,
			Port: httpPort,
		})
	}

	return ports
}

// endpointPorts returns the list of EndpointPort for the kubelet Endpoints.
// If httpMetricsEnabled is false, the insecure HTTP port (10255) is excluded.
func (c *Controller) endpointPorts() []v1.EndpointPort {
	ports := []v1.EndpointPort{
		{
			Name: httpsPortName,
			Port: httpsPort,
		},
		{
			Name: cAdvisorPortName,
			Port: cAdvisorPort,
		},
	}

	if c.httpMetricsEnabled {
		ports = append(ports, v1.EndpointPort{
			Name: httpPortName,
			Port: httpPort,
		})
	}

	return ports
}

// endpointSlicePorts returns the list of EndpointPort for the kubelet EndpointSlice.
// If httpMetricsEnabled is false, the insecure HTTP port (10255) is excluded.
func (c *Controller) endpointSlicePorts() []discoveryv1.EndpointPort {
	ports := []discoveryv1.EndpointPort{
		{
			Name: ptr.To(httpsPortName),
			Port: ptr.To(httpsPort),
		},
		{
			Name: ptr.To(cAdvisorPortName),
			Port: ptr.To(cAdvisorPort),
		},
	}

	if c.httpMetricsEnabled {
		ports = append(ports, discoveryv1.EndpointPort{
			Name: ptr.To(httpPortName),
			Port: ptr.To(httpPort),
		})
	}

	return ports
}
