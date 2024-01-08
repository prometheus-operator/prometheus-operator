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
						},
					},
				},
			},
			expectedAddresses: []string{"10.0.0.1"},
			expectedErrors:    1,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			addrs, errs := getNodeAddresses(c.nodes)
			require.Len(t, errs, c.expectedErrors)

			ips := make([]string, 0)
			for _, addr := range addrs {
				ips = append(ips, addr.IP)
			}

			require.Equal(t, c.expectedAddresses, ips)
		})
	}
}
