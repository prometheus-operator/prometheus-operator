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
)

func TestKubeletEndpointSliceMigration(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ctx := context.Background()
	ns := framework.CreateNamespace(ctx, t, testCtx)

	t.Log("Starting Kubelet EndpointSlice migration test")

	// Phase 1: Test with Endpoints enabled
	t.Run("Phase1_EndpointsOnly", func(t *testing.T) {
		t.Log("Phase 1: Testing with kubelet-endpoints=true and kubelet-endpointslice=false")

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

		// Register cleanup
		for _, f := range finalizers {
			testCtx.AddFinalizerFn(f)
		}

		// Wait for operator to be ready
		time.Sleep(10 * time.Second)

		// Verify Endpoints resource is created
		err = waitForKubeletEndpoints(ctx, t)
		require.NoError(t, err, "Failed to find kubelet Endpoints resource")

		// Verify EndpointSlice resources are NOT created (should be 0)
		endpointSlices, err := getKubeletEndpointSlices(ctx)
		require.NoError(t, err)
		require.Empty(t, endpointSlices, "EndpointSlices should not be created when only Endpoints is enabled")

		t.Log("Phase 1 completed successfully - Endpoints created, EndpointSlices not created")
	})

	// Phase 2: Switch to EndpointSlices
	t.Run("Phase2_EndpointSlicesOnly", func(t *testing.T) {
		t.Log("Phase 2: Switching to kubelet-endpoints=false and kubelet-endpointslice=true")

		// Redeploy operator with EndpointSlices enabled and Endpoints disabled
		finalizers, err := framework.CreateOrUpdatePrometheusOperatorWithOpts(ctx, operatorFramework.PrometheusOperatorOpts{
			Namespace:           ns,
			ClusterRoleBindings: true,
			AdditionalArgs: []string{
				fmt.Sprintf("--kubelet-service=%s/%s", kubeletServiceNamespace, kubeletServiceName),
				"--kubelet-endpoints=false",
				"--kubelet-endpointslice=true",
			},
		})
		require.NoError(t, err)

		// Register cleanup
		for _, f := range finalizers {
			testCtx.AddFinalizerFn(f)
		}

		// Wait for operator to reconcile
		time.Sleep(15 * time.Second)

		// Verify EndpointSlice resources are created
		err = waitForKubeletEndpointSlices(ctx, t)
		require.NoError(t, err, "Failed to find kubelet EndpointSlice resources")

		// Verify the EndpointSlices contain valid kubelet targets
		err = verifyKubeletEndpointSliceTargets(ctx, t)
		require.NoError(t, err, "EndpointSlice targets verification failed")

		t.Log("Phase 2 completed successfully - EndpointSlices created and contain valid targets")
	})

	// Phase 3: Verify migration worked correctly
	t.Run("Phase3_MigrationVerification", func(t *testing.T) {
		t.Log("Phase 3: Verifying migration from Endpoints to EndpointSlices worked correctly")

		// Verify service is still present
		_, err := framework.KubeClient.CoreV1().Services(kubeletServiceNamespace).Get(ctx, kubeletServiceName, metav1.GetOptions{})
		require.NoError(t, err, "Kubelet service should still exist after migration")

		// Get EndpointSlices and verify they match expected kubelet ports
		endpointSlices, err := getKubeletEndpointSlices(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, endpointSlices, "EndpointSlices should exist after migration")

		// Verify EndpointSlices have the expected ports (https-metrics, http-metrics, cadvisor)
		expectedPorts := map[string]int32{
			"https-metrics": 10250,
			"http-metrics":  10255,
			"cadvisor":      4194,
		}

		for _, eps := range endpointSlices {
			require.NotEmpty(t, eps.Endpoints, "EndpointSlice should have endpoints")

			// Verify ports
			portMap := make(map[string]int32)
			for _, port := range eps.Ports {
				if port.Name != nil && port.Port != nil {
					portMap[*port.Name] = *port.Port
				}
			}

			for expectedName, expectedPort := range expectedPorts {
				actualPort, exists := portMap[expectedName]
				require.True(t, exists, "Expected port %s not found in EndpointSlice", expectedName)
				require.Equal(t, expectedPort, actualPort, "Port %s has incorrect value", expectedName)
			}
		}

		t.Log("Phase 3 completed successfully - Migration verification passed")
	})
}

// waitForKubeletEndpoints waits for the kubelet Endpoints resource to be created
func waitForKubeletEndpoints(ctx context.Context, t *testing.T) error {
	timeout := 60 * time.Second
	interval := 2 * time.Second

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for kubelet Endpoints resource")
		default:
			endpoints, err := framework.KubeClient.CoreV1().Endpoints(kubeletServiceNamespace).Get(ctx, kubeletServiceName, metav1.GetOptions{})
			if err == nil && endpoints != nil && len(endpoints.Subsets) > 0 && len(endpoints.Subsets[0].Addresses) > 0 {
				t.Logf("Found kubelet Endpoints resource with %d addresses", len(endpoints.Subsets[0].Addresses))
				return nil
			}
			time.Sleep(interval)
		}
	}
}

// waitForKubeletEndpointSlices waits for kubelet EndpointSlice resources to be created
func waitForKubeletEndpointSlices(ctx context.Context, t *testing.T) error {
	timeout := 60 * time.Second
	interval := 2 * time.Second

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for kubelet EndpointSlice resources")
		default:
			endpointSlices, err := getKubeletEndpointSlices(ctx)
			if err == nil && len(endpointSlices) > 0 {
				// Check if any EndpointSlice has endpoints
				for _, eps := range endpointSlices {
					if len(eps.Endpoints) > 0 {
						t.Logf("Found kubelet EndpointSlice resources: %d slices with endpoints", len(endpointSlices))
						return nil
					}
				}
			}
			time.Sleep(interval)
		}
	}
}

// getKubeletEndpointSlices retrieves all EndpointSlice resources for the kubelet service
func getKubeletEndpointSlices(ctx context.Context) ([]discoveryv1.EndpointSlice, error) {
	endpointSlices, err := framework.KubeClient.DiscoveryV1().EndpointSlices(kubeletServiceNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set{discoveryv1.LabelServiceName: kubeletServiceName}.String(),
	})
	if err != nil {
		return nil, err
	}
	return endpointSlices.Items, nil
}

// verifyKubeletEndpointSliceTargets verifies that EndpointSlice targets are valid kubelet endpoints
func verifyKubeletEndpointSliceTargets(ctx context.Context, t *testing.T) error {
	endpointSlices, err := getKubeletEndpointSlices(ctx)
	if err != nil {
		return err
	}

	if len(endpointSlices) == 0 {
		return fmt.Errorf("no EndpointSlice resources found for kubelet service")
	}

	// Get all nodes to verify endpoints correspond to actual nodes
	nodes, err := framework.KubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %v", err)
	}

	// Collect all endpoint addresses from EndpointSlices
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

	// Verify we have endpoints
	if len(endpointAddresses) == 0 {
		return fmt.Errorf("no endpoint addresses found in EndpointSlices")
	}

	// Verify we have node names
	if len(endpointNodeNames) == 0 {
		return fmt.Errorf("no node names found in EndpointSlice endpoints")
	}

	t.Logf("Found %d endpoint addresses and %d node names in EndpointSlices", len(endpointAddresses), len(endpointNodeNames))

	// Verify that node names in endpoints correspond to actual nodes
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
