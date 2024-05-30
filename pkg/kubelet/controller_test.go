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
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetNodeAddresses(t *testing.T) {
	for _, c := range []struct {
		name              string
		nodes             *v1.NodeList
		expectedAddresses []string
		expectedErrors    int
	}{
		{
			name: "simple",
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node-0",
						},
						Status: v1.NodeStatus{
							Addresses: []v1.NodeAddress{
								{
									Address: "10.0.0.1",
									Type:    v1.NodeInternalIP,
								},
							},
							Conditions: []v1.NodeCondition{
								{
									Type:   v1.NodeReady,
									Status: v1.ConditionTrue,
								},
							},
						},
					},
				},
			},
			expectedAddresses: []string{"10.0.0.1"},
			expectedErrors:    0,
		},
		{
			// Replicates #1815
			name: "missing ip on one node",
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node-0",
						},
						Status: v1.NodeStatus{
							Addresses: []v1.NodeAddress{
								{
									Address: "node-0",
									Type:    v1.NodeHostName,
								},
							},
							Conditions: []v1.NodeCondition{
								{
									Type:   v1.NodeReady,
									Status: v1.ConditionTrue,
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node-1",
						},
						Status: v1.NodeStatus{
							Addresses: []v1.NodeAddress{
								{
									Address: "10.0.0.1",
									Type:    v1.NodeInternalIP,
								},
							},
							Conditions: []v1.NodeCondition{
								{
									Type:   v1.NodeReady,
									Status: v1.ConditionTrue,
								},
							},
						},
					},
				},
			},
			expectedAddresses: []string{"10.0.0.1"},
			expectedErrors:    1,
		},
		{
			name: "not ready node unique ip",
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node-0",
						},
						Status: v1.NodeStatus{
							Addresses: []v1.NodeAddress{
								{
									Address: "10.0.0.1",
									Type:    v1.NodeInternalIP,
								},
							},
							Conditions: []v1.NodeCondition{
								{
									Type:   v1.NodeReady,
									Status: v1.ConditionTrue,
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node-1",
						},
						Status: v1.NodeStatus{
							Addresses: []v1.NodeAddress{
								{
									Address: "10.0.0.2",
									Type:    v1.NodeInternalIP,
								},
							},
							Conditions: []v1.NodeCondition{
								{
									Type:   v1.NodeReady,
									Status: v1.ConditionUnknown,
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node-2",
						},
						Status: v1.NodeStatus{
							Addresses: []v1.NodeAddress{
								{
									Address: "10.0.0.3",
									Type:    v1.NodeInternalIP,
								},
							},
							Conditions: []v1.NodeCondition{
								{
									Type:   v1.NodeReady,
									Status: v1.ConditionFalse,
								},
							},
						},
					},
				},
			},
			expectedAddresses: []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"},
			expectedErrors:    0,
		},
		{
			name: "not ready node duplicate ip",
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node-0",
						},
						Status: v1.NodeStatus{
							Addresses: []v1.NodeAddress{
								{
									Address: "10.0.0.1",
									Type:    v1.NodeInternalIP,
								},
							},
							Conditions: []v1.NodeCondition{
								{
									Type:   v1.NodeReady,
									Status: v1.ConditionTrue,
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node-1",
						},
						Status: v1.NodeStatus{
							Addresses: []v1.NodeAddress{
								{
									Address: "10.0.0.1",
									Type:    v1.NodeInternalIP,
								},
							},
							Conditions: []v1.NodeCondition{
								{
									Type:   v1.NodeReady,
									Status: v1.ConditionUnknown,
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node-2",
						},
						Status: v1.NodeStatus{
							Addresses: []v1.NodeAddress{
								{
									Address: "10.0.0.3",
									Type:    v1.NodeInternalIP,
								},
							},
							Conditions: []v1.NodeCondition{
								{
									Type:   v1.NodeReady,
									Status: v1.ConditionFalse,
								},
							},
						},
					},
				},
			},
			expectedAddresses: []string{"10.0.0.1", "10.0.0.3"},
			expectedErrors:    0,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			controller := Controller{
				nodeAddressPriority: "internal",
			}

			addrs, errs := controller.getNodeAddresses(c.nodes)
			require.Len(t, errs, c.expectedErrors)
			checkNodeAddresses(t, addrs, c.expectedAddresses)
		})
	}
}

func TestNodeAddressPriority(t *testing.T) {
	nodes := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-0",
				},
				Status: v1.NodeStatus{
					Addresses: []v1.NodeAddress{
						{
							Address: "192.168.0.100",
							Type:    v1.NodeInternalIP,
						},
						{
							Address: "203.0.113.100",
							Type:    v1.NodeExternalIP,
						},
					},
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Status: v1.NodeStatus{
					Addresses: []v1.NodeAddress{
						{
							Address: "104.27.131.189",
							Type:    v1.NodeExternalIP,
						},
						{
							Address: "192.168.1.100",
							Type:    v1.NodeInternalIP,
						},
					},
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
		},
	}

	internalC := Controller{
		nodeAddressPriority: "internal",
	}
	actualAddresses, errs := internalC.getNodeAddresses(nodes)
	require.Empty(t, errs)
	expectedAddresses := []string{"192.168.0.100", "192.168.1.100"}
	checkNodeAddresses(t, actualAddresses, expectedAddresses)

	externalC := Controller{
		nodeAddressPriority: "external",
	}
	actualAddresses, errs = externalC.getNodeAddresses(nodes)
	require.Empty(t, errs)
	expectedAddresses = []string{"203.0.113.100", "104.27.131.189"}
	checkNodeAddresses(t, actualAddresses, expectedAddresses)
}

func checkNodeAddresses(t *testing.T, actualAddresses []v1.EndpointAddress, expectedAddresses []string) {
	ips := make([]string, 0)
	for _, addr := range actualAddresses {
		ips = append(ips, addr.IP)
	}

	require.Equal(t, expectedAddresses, ips)
}
