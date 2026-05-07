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

package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
)

// polls is the number of times ServerResourcesForGroupVersion function in this file runs.
// We use this to test whether checkInstalledWithTimeout polls several times even if ServerResourcesForGroupVersion returns false first time.
var polls = 0

type mockKubernetesClient struct {
	kubernetes.Interface
	discovery discovery.DiscoveryInterface
}

func newMockKubernetesClient() kubernetes.Interface {
	return &mockKubernetesClient{
		discovery: &mockDiscoveryClient{},
	}
}

func (m *mockKubernetesClient) Discovery() discovery.DiscoveryInterface {
	return m.discovery
}

type mockDiscoveryClient struct {
	discovery.DiscoveryInterface
}

func (m *mockDiscoveryClient) ServerResourcesForGroupVersion(_ string) (*metav1.APIResourceList, error) {
	// Make the function runs for 10 second so that
	// we can control how many time this is called until timeout
	time.Sleep(10 * time.Second)

	polls++

	// checkInstalledWithTimeout will only return true after 3 polls.
	if polls >= 3 {
		return &metav1.APIResourceList{
			APIResources: []metav1.APIResource{
				metav1.APIResource{
					Name: "true",
				},
			},
		}, nil
	}

	return &metav1.APIResourceList{
		APIResources: []metav1.APIResource{
			metav1.APIResource{
				Name: "false",
			},
		},
	}, nil
}

func TestWaitForCRDInstalled(t *testing.T) {
	ctx := context.Background()
	client := newMockKubernetesClient()

	// Because we set the timeout as 5 seconds here, and ServerResourcesForGroupVersion runs for 10 second,
	// there will be only 1 poll, and we expectcheckInstalledWithTimeout to return false.
	installed, err := checkInstalledWithTimeout(ctx, client, storagev1.SchemeGroupVersion, "true", 5*time.Second)
	require.NoError(t, err)
	require.Equal(t, polls, 1)
	require.False(t, installed)

	// Now the timeout is 50 seconds, ServerResourcesForGroupVersion will be polled more than 1 time.
	// Because ServerResourcesForGroupVersion has already run at least 3 times now, we expectcheckInstalledWithTimeout to return true.
	installed, err = checkInstalledWithTimeout(ctx, client, storagev1.SchemeGroupVersion, "true", 50*time.Second)
	require.NoError(t, err)
	require.GreaterOrEqual(t, polls, 3)
	require.True(t, installed)
}

func TestSetCRDToWaitFor(t *testing.T) {
	crds := &crdsList{}

	require.NoError(t, crds.Set("storage class"))
	require.NoError(t, crds.Set("Scrape Config"))
	require.NoError(t, crds.Set("PROMETHEUS"))
	require.NoError(t, crds.Set("proMeTheusAgent"))
	require.NoError(t, crds.Set("ALERT MANAGER"))
	require.NoError(t, crds.Set("thaNos ruLer"))

	require.Error(t, crds.Set("foo"))
}
