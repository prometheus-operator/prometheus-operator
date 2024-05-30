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
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

const resyncPeriod = 3 * time.Minute

type Controller struct {
	logger log.Logger

	kclient kubernetes.Interface

	nodeAddressLookupErrors prometheus.Counter
	nodeEndpointSyncs       prometheus.Counter
	nodeEndpointSyncErrors  prometheus.Counter

	kubeletObjectName      string
	kubeletObjectNamespace string
	kubeletSelector        string

	annotations operator.Map
	labels      operator.Map

	nodeAddressPriority string
}

func New(
	logger log.Logger,
	restConfig *rest.Config,
	r prometheus.Registerer,
	kubeletObject string,
	kubeletSelector operator.LabelSelector,
	commonAnnotations operator.Map,
	commonLabels operator.Map,
	nodeAddressPriority operator.NodeAddressPriority,
) (*Controller, error) {
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("instantiating kubernetes client failed: %w", err)
	}

	c := &Controller{
		kclient: client,

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

		kubeletSelector: kubeletSelector.String(),

		annotations: commonAnnotations,
		labels:      commonLabels,

		nodeAddressPriority: nodeAddressPriority.String(),
	}

	r.MustRegister(
		c.nodeAddressLookupErrors,
		c.nodeEndpointSyncs,
		c.nodeEndpointSyncErrors,
	)

	parts := strings.Split(kubeletObject, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("malformatted kubelet object string %q, must be in format \"namespace/name\"", kubeletObject)
	}
	c.kubeletObjectNamespace = parts[0]
	c.kubeletObjectName = parts[1]

	c.logger = log.With(logger, "kubelet_object", kubeletObject)

	return c, nil
}

func (c *Controller) Run(ctx context.Context) error {
	ticker := time.NewTicker(resyncPeriod)
	defer ticker.Stop()
	for {
		c.syncNodeEndpointsWithLogError(ctx)

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

func (c *Controller) getNodeAddresses(nodes *v1.NodeList) ([]v1.EndpointAddress, []error) {
	addresses := make([]v1.EndpointAddress, 0)
	errs := make([]error, 0)
	readyKnownNodes := make(map[string]string)
	readyUnknownNodes := make(map[string]string)

	for _, n := range nodes.Items {
		address, _, err := c.nodeAddress(n)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to determine hostname for node (%s): %w", n.Name, err))
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

		if !nodeReadyConditionKnown(n) {
			if c.logger != nil {
				level.Info(c.logger).Log("msg", "Node Ready condition is Unknown", "node", n.GetName())
			}
			readyUnknownNodes[address] = n.Name
			continue
		}
		readyKnownNodes[address] = n.Name
	}

	// We want to remove any nodes that have an unknown ready state *and* a
	// duplicate IP address. If this is the case, we want to keep just the node
	// with the duplicate IP address that has a known ready state. This also
	// ensures that order of addresses are preserved.
	addressesFinal := make([]v1.EndpointAddress, 0)
	for _, address := range addresses {
		knownNodeName, foundKnown := readyKnownNodes[address.IP]
		_, foundUnknown := readyUnknownNodes[address.IP]
		if foundKnown && foundUnknown && address.TargetRef.Name != knownNodeName {
			continue
		}
		addressesFinal = append(addressesFinal, address)
	}

	return addressesFinal, errs
}

func (c *Controller) syncNodeEndpointsWithLogError(ctx context.Context) {
	level.Debug(c.logger).Log("msg", "Synchronizing nodes")

	c.nodeEndpointSyncs.Inc()
	err := c.syncNodeEndpoints(ctx)
	if err != nil {
		c.nodeEndpointSyncErrors.Inc()
		level.Error(c.logger).Log("msg", "Failed to synchronize nodes", "err", err)
	}
}

func (c *Controller) syncNodeEndpoints(ctx context.Context) error {
	eps := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.kubeletObjectName,
			Annotations: c.annotations,
			Labels: c.labels.Merge(map[string]string{
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

	nodes, err := c.kclient.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: c.kubeletSelector})
	if err != nil {
		return fmt.Errorf("listing nodes failed: %w", err)
	}

	level.Debug(c.logger).Log("msg", "Nodes retrieved from the Kubernetes API", "num_nodes", len(nodes.Items))

	addresses, errs := c.getNodeAddresses(nodes)
	if len(errs) > 0 {
		for _, err := range errs {
			level.Warn(c.logger).Log("err", err)
		}
		c.nodeAddressLookupErrors.Add(float64(len(errs)))
	}
	level.Debug(c.logger).Log("msg", "Nodes converted to endpoint addresses", "num_addresses", len(addresses))

	eps.Subsets[0].Addresses = addresses

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.kubeletObjectName,
			Annotations: c.annotations,
			Labels: c.labels.Merge(map[string]string{
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

	level.Debug(c.logger).Log("msg", "Updating Kubernetes service", "service")
	err = k8sutil.CreateOrUpdateService(ctx, c.kclient.CoreV1().Services(c.kubeletObjectNamespace), svc)
	if err != nil {
		return fmt.Errorf("synchronizing kubelet service object failed: %w", err)
	}

	level.Debug(c.logger).Log("msg", "Updating Kubernetes endpoint")
	err = k8sutil.CreateOrUpdateEndpoints(ctx, c.kclient.CoreV1().Endpoints(c.kubeletObjectNamespace), eps)
	if err != nil {
		return fmt.Errorf("synchronizing kubelet endpoints object failed: %w", err)
	}

	return nil
}
