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

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	operatorFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
)

const (
	kubeletServiceName      = "prometheus-operator-kubelet"
	kubeletServiceNamespace = "kube-system"

	resourceTimeout = 90 * time.Second
	pollInterval    = 5 * time.Second
	stabilizeTime   = 20 * time.Second
)

func testKubeletEndpointSliceMigration(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ctx := context.Background()
	ns := framework.CreateNamespace(ctx, t, testCtx)

	t.Log("Starting Kubelet EndpointSlice migration test")

	t.Log("Phase 1: Deploying operator with kubelet-endpoints=true")
	finalizers, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(ctx, operatorFramework.PrometheusOperatorOpts{
		Namespace:           ns,
		ClusterRoleBindings: true,
		AdditionalArgs: []string{
			fmt.Sprintf("--kubelet-service=%s/%s", kubeletServiceNamespace, kubeletServiceName),
			"--kubelet-endpoints=true",
			"--kubelet-endpointslice=false",
		},
	})
	require.NoError(t, err)

	for _, f := range finalizers {
		testCtx.AddFinalizerFn(f)
	}

	time.Sleep(stabilizeTime)

	err = waitForKubeletEndpoints(ctx, t)
	require.NoError(t, err, "Failed to verify kubelet Endpoints resource creation")

	t.Log("Successfully verified Endpoints resource creation")

	t.Log("Phase 2: Switching operator to kubelet-endpointslice=true")
	_, err = framework.CreateOrUpdatePrometheusOperatorWithOpts(ctx, operatorFramework.PrometheusOperatorOpts{
		Namespace:           ns,
		ClusterRoleBindings: true,
		AdditionalArgs: []string{
			fmt.Sprintf("--kubelet-service=%s/%s", kubeletServiceNamespace, kubeletServiceName),
			"--kubelet-endpoints=false",
			"--kubelet-endpointslice=true",
		},
	})
	require.NoError(t, err)

	time.Sleep(stabilizeTime)

	err = waitForKubeletEndpointSlices(ctx, t)
	require.NoError(t, err, "Failed to verify kubelet EndpointSlice resource creation")

	err = verifyKubeletEndpointSliceTargets(ctx, t)
	require.NoError(t, err, "EndpointSlice targets verification failed")

	t.Log("Successfully verified EndpointSlice resource creation and target validation")
	t.Log("Kubelet EndpointSlice migration test completed successfully")
}

func waitForKubeletEndpoints(ctx context.Context, t *testing.T) error {
	ctx, cancel := context.WithTimeout(ctx, resourceTimeout)
	defer cancel()

	t.Log("Waiting for kubelet Endpoints resource creation...")
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for kubelet Endpoints resource after %v", resourceTimeout)
		default:
			endpoints, err := framework.KubeClient.CoreV1().Endpoints(kubeletServiceNamespace).Get(ctx, kubeletServiceName, metav1.GetOptions{})
			if err == nil && endpoints != nil && len(endpoints.Subsets) > 0 && len(endpoints.Subsets[0].Addresses) > 0 {
				t.Logf("Found kubelet Endpoints resource with %d addresses", len(endpoints.Subsets[0].Addresses))
				return nil
			}
			if err != nil {
				t.Logf("Error getting kubelet Endpoints: %v", err)
			}
			time.Sleep(pollInterval)
		}
	}
}

func waitForKubeletEndpointSlices(ctx context.Context, t *testing.T) error {
	ctx, cancel := context.WithTimeout(ctx, resourceTimeout)
	defer cancel()

	t.Log("Waiting for kubelet EndpointSlice resource creation...")
	for {
		select {
		case <-ctx.Done():
			endpointSlices, err := getKubeletEndpointSlices(ctx)
			if err != nil {
				t.Logf("Error getting EndpointSlices during timeout: %v", err)
			} else {
				t.Logf("EndpointSlices found during timeout: %d", len(endpointSlices))
				for i, eps := range endpointSlices {
					t.Logf("EndpointSlice %d: %s, endpoints: %d", i, eps.Name, len(eps.Endpoints))
				}
			}
			return fmt.Errorf("timeout waiting for kubelet EndpointSlice resources after %v", resourceTimeout)
		default:
			endpointSlices, err := getKubeletEndpointSlices(ctx)
			if err != nil {
				t.Logf("Error getting EndpointSlices: %v", err)
				time.Sleep(pollInterval)
				continue
			}

			if len(endpointSlices) > 0 {
				totalEndpoints := 0
				for _, eps := range endpointSlices {
					totalEndpoints += len(eps.Endpoints)
				}
				if totalEndpoints > 0 {
					t.Logf("Found kubelet EndpointSlice resources: %d slices with %d total endpoints", len(endpointSlices), totalEndpoints)
					return nil
				}
				t.Logf("Found %d EndpointSlice resources but no endpoints yet", len(endpointSlices))
			}
			time.Sleep(pollInterval)
		}
	}
}

func getKubeletEndpointSlices(ctx context.Context) ([]discoveryv1.EndpointSlice, error) {
	endpointSlices, err := framework.KubeClient.DiscoveryV1().EndpointSlices(kubeletServiceNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set{discoveryv1.LabelServiceName: kubeletServiceName}.String(),
	})
	if err != nil {
		return nil, err
	}
	return endpointSlices.Items, nil
}

func verifyKubeletEndpointSliceTargets(ctx context.Context, t *testing.T) error {
	endpointSlices, err := getKubeletEndpointSlices(ctx)
	if err != nil {
		return err
	}

	if len(endpointSlices) == 0 {
		return fmt.Errorf("no EndpointSlice resources found for kubelet service")
	}

	nodes, err := framework.KubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %v", err)
	}

	var endpointAddresses []string
	var endpointNodeNames []string

	for _, eps := range endpointSlices {
		for _, endpoint := range eps.Endpoints {
			endpointAddresses = append(endpointAddresses, endpoint.Addresses...)
			if endpoint.NodeName != nil {
				endpointNodeNames = append(endpointNodeNames, *endpoint.NodeName)
			}
		}
	}

	if len(endpointAddresses) == 0 {
		return fmt.Errorf("no endpoint addresses found in EndpointSlices")
	}

	if len(endpointNodeNames) == 0 {
		return fmt.Errorf("no node names found in EndpointSlice endpoints")
	}

	t.Logf("Found %d endpoint addresses and %d node names in EndpointSlices", len(endpointAddresses), len(endpointNodeNames))

	nodeNameSet := make(map[string]bool)
	for _, node := range nodes.Items {
		nodeNameSet[node.Name] = true
	}

	for _, nodeName := range endpointNodeNames {
		if !nodeNameSet[nodeName] {
			return fmt.Errorf("EndpointSlice references non-existent node: %s", nodeName)
		}
	}

	t.Logf("All EndpointSlice targets correspond to valid nodes")
	return nil
}
