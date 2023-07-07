// Copyright 2017 The prometheus-operator Authors
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
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

func TestListOptions(t *testing.T) {
	for i := 0; i < 1000; i++ {
		o := ListOptions("test")
		if o.LabelSelector != "app.kubernetes.io/name=prometheus,prometheus=test" && o.LabelSelector != "prometheus=test,app.kubernetes.io/name=prometheus" {
			t.Fatalf("LabelSelector not computed correctly\n\nExpected: \"app.kubernetes.io/name=prometheus,prometheus=test\"\n\nGot:      %#+v", o.LabelSelector)
		}
	}
}

func TestCreateStatefulSetInputHash(t *testing.T) {
	falseVal := false

	for _, tc := range []struct {
		name string
		a, b monitoringv1.Prometheus

		equal bool
	}{
		{
			name: "different generations",
			a: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.0",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("200Mi"),
							},
						},
					},
				},
			},
			b: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.0",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("100Mi"),
							},
						},
					},
				},
			},
		},
		{
			// differrent resource.Quantity produce the same hash because the
			// struct contains private fields that aren't integrated into the
			// hash computation.
			name: "different specs but same hash",
			a: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.0",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("200Mi"),
							},
						},
					},
				},
			},
			b: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.0",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("100Mi"),
							},
						},
					},
				},
			},
			equal: true,
		},
		{
			name: "same hash with different status",
			a: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Paused:  true,
						Version: "v1.7.2",
					},
				},
			},
			b: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Paused:  true,
						Version: "v1.7.2",
					},
				},
				Status: monitoringv1.PrometheusStatus{
					Paused: true,
				},
			},

			equal: true,
		},
		{
			name: "different labels",
			a: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"foo": "bar"},
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.2",
					},
				},
			},
			b: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.2",
					},
				},
			},

			equal: false,
		},
		{
			name: "different annotations",
			a: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"foo": "bar"},
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.2",
					},
				},
			},
			b: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.2",
					},
				},
			},

			equal: false,
		},
		{
			name: "different web http2",
			a: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.2",
						Web: &monitoringv1.PrometheusWebSpec{
							WebConfigFileFields: monitoringv1.WebConfigFileFields{
								HTTPConfig: &monitoringv1.WebHTTPConfig{
									HTTP2: &falseVal,
								},
							},
						},
					},
				},
			},
			b: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: "v1.7.2",
					},
				},
			},

			equal: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := operator.Config{}

			p1Hash, err := createSSetInputHash(tc.a, c, []string{}, nil, appsv1.StatefulSetSpec{})
			if err != nil {
				t.Fatal(err)
			}

			p2Hash, err := createSSetInputHash(tc.b, c, []string{}, nil, appsv1.StatefulSetSpec{})
			if err != nil {
				t.Fatal(err)
			}

			if !tc.equal {
				if p1Hash == p2Hash {
					t.Fatal("expected two different Prometheus CRDs to produce different hashes but got equal hash")
				}
				return
			}

			if p1Hash != p2Hash {
				t.Fatal("expected two Prometheus CRDs to produce the same hash but got different hash")
			}

			p2Hash, err = createSSetInputHash(tc.a, c, []string{}, nil, appsv1.StatefulSetSpec{Replicas: func(i int32) *int32 { return &i }(2)})
			if err != nil {
				t.Fatal(err)
			}

			if p1Hash == p2Hash {
				t.Fatal("expected same Prometheus CRDs with different statefulset specs to produce different hashes but got equal hash")
			}
		})
	}
}

func TestGetNodeAddresses(t *testing.T) {
	cases := []struct {
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
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			addrs, errs := getNodeAddresses(c.nodes)
			if len(errs) != c.expectedErrors {
				t.Errorf("Expected %d errors, got %d. Errors: %v", c.expectedErrors, len(errs), errs)
			}
			ips := make([]string, 0)
			for _, addr := range addrs {
				ips = append(ips, addr.IP)
			}
			if !reflect.DeepEqual(ips, c.expectedAddresses) {
				t.Error(pretty.Compare(ips, c.expectedAddresses))
			}
		})
	}
}
