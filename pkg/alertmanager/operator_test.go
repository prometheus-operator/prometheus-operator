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

package alertmanager

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	authv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	monitoringfake "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/fake"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

func TestCreateStatefulSetInputHash(t *testing.T) {
	falseVal := false

	for _, tc := range []struct {
		name string
		a, b monitoringv1.Alertmanager

		equal bool
	}{
		{
			name: "different generations",
			a: monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
				},
				Spec: monitoringv1.AlertmanagerSpec{
					Version: "v0.0.1",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("200Mi"),
						},
					},
				},
			},
			b: monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
				},
				Spec: monitoringv1.AlertmanagerSpec{
					Version: "v0.0.1",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("100Mi"),
						},
					},
				},
			},

			equal: false,
		},
		{
			// different resource.Quantity produce the same hash because the
			// struct contains private fields that aren't integrated into the
			// hash computation.
			name: "different specs but same hash",
			a: monitoringv1.Alertmanager{
				Spec: monitoringv1.AlertmanagerSpec{
					Version: "v0.0.1",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("200Mi"),
						},
					},
				},
			},
			b: monitoringv1.Alertmanager{
				Spec: monitoringv1.AlertmanagerSpec{
					Version: "v0.0.1",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("100Mi"),
						},
					},
				},
			},

			equal: true,
		},
		{
			name: "different labels",
			a: monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"foo": "bar"},
				},
				Spec: monitoringv1.AlertmanagerSpec{
					Version: "v0.0.1",
				},
			},
			b: monitoringv1.Alertmanager{
				Spec: monitoringv1.AlertmanagerSpec{
					Version: "v0.0.1",
				},
			},

			equal: false,
		},
		{
			name: "different annotations",
			a: monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"foo": "bar"},
				},
				Spec: monitoringv1.AlertmanagerSpec{
					Version: "v0.0.1",
				},
			},
			b: monitoringv1.Alertmanager{
				Spec: monitoringv1.AlertmanagerSpec{
					Version: "v0.0.1",
				},
			},

			equal: false,
		},
		{
			name: "different web http2",
			a: monitoringv1.Alertmanager{
				Spec: monitoringv1.AlertmanagerSpec{
					Version: "v0.0.1",
					Web: &monitoringv1.AlertmanagerWebSpec{
						WebConfigFileFields: monitoringv1.WebConfigFileFields{
							HTTPConfig: &monitoringv1.WebHTTPConfig{
								HTTP2: &falseVal,
							},
						},
					},
				},
			},
			b: monitoringv1.Alertmanager{
				Spec: monitoringv1.AlertmanagerSpec{
					Version: "v0.0.1",
				},
			},

			equal: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a1Hash, err := createSSetInputHash(tc.a, Config{}, &operator.ShardedSecret{}, appsv1.StatefulSetSpec{})
			require.NoError(t, err)

			a2Hash, err := createSSetInputHash(tc.b, Config{}, &operator.ShardedSecret{}, appsv1.StatefulSetSpec{})
			require.NoError(t, err)

			if !tc.equal {
				require.NotEqual(t, a1Hash, a2Hash, "expected two different Alertmanager CRDs to produce different hashes but got equal hash")
				return
			}

			require.Equal(t, a1Hash, a2Hash, "expected two Alertmanager CRDs to produce the same hash but got different hash")

			a2Hash, err = createSSetInputHash(tc.a, Config{}, &operator.ShardedSecret{}, appsv1.StatefulSetSpec{Replicas: ptr.To(int32(2))})
			require.NoError(t, err)

			require.NotEqual(t, a1Hash, a2Hash, "expected same Alertmanager CRDs with different statefulset specs to produce different hashes but got equal hash")
		})
	}
}

// Test to exercise the function checkAlertmanagerConfigResource
// and validate that semantic validation is in place for all the fields in the
// AlertmanagerConfig CR. The validation is performed by the operator
// after selecting AlertmanagerConfig resources and before generating the
// Alertmanager configuration.
func TestCheckAlertmanagerConfig(t *testing.T) {
	version, err := semver.ParseTolerant(operator.DefaultAlertmanagerVersion)
	require.NoError(t, err)

	c := fake.NewSimpleClientset(
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret",
				Namespace: "ns1",
			},
			Data: map[string][]byte{
				"key1":                 []byte("https://val1.com"),
				"template-url":         []byte("{{ .labels.url }}"),
				"invalid-url":          []byte("://foo"),
				"invalid-template-url": []byte("{{ .labels.url"),
			},
		},
	)

	for _, tc := range []struct {
		amConfig *monitoringv1alpha1.AlertmanagerConfig
		ok       bool
	}{
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "empty",
					Namespace: "ns1",
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "inhibitRulesOnly",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					InhibitRules: []monitoringv1alpha1.InhibitRule{
						{
							SourceMatch: []monitoringv1alpha1.Matcher{
								{
									Name:  "alertname",
									Value: "NodeNotReady",
								},
							},
							TargetMatch: []monitoringv1alpha1.Matcher{
								{
									Name:  "alertname",
									Value: "TargetDown",
								},
							},
							Equal: []string{"node"},
						},
					},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "missing-receiver",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "not-existing",
					},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-receivers",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
						Routes: []apiextensionsv1.JSON{
							{Raw: []byte(`{"receiver": "recv2"}`)},
						},
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
					}, {
						Name: "recv2",
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nested-routes-without-receiver",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
						Routes: []apiextensionsv1.JSON{
							{Raw: []byte(`{"routes": [{"receiver": "recv2"}]}`)},
						},
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
					}, {
						Name: "recv2",
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "top-level-route-without-receiver",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-receivers",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
					}, {
						Name: "recv1",
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "missing-pagerduty-service-key",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						PagerDutyConfigs: []monitoringv1alpha1.PagerDutyConfig{{
							ServiceKey: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
								Key:                  "not-existing",
							},
						}},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-pagerduty-service-key",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						PagerDutyConfigs: []monitoringv1alpha1.PagerDutyConfig{{
							ServiceKey: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
								Key:                  "key1",
							},
						}},
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "missing-pagerduty-routing-key",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						PagerDutyConfigs: []monitoringv1alpha1.PagerDutyConfig{{
							RoutingKey: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
								Key:                  "not-existing",
							},
						}},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-pagerduty-routing-key",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						PagerDutyConfigs: []monitoringv1alpha1.PagerDutyConfig{{
							RoutingKey: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
								Key:                  "key1",
							},
						}},
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "webhook-without-url",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
							{},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "webhook-with-url",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
							{
								URL: ptr.To("http://test.example.com"),
							},
						},
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "webhook-with-invalid-url",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
							{
								URL: ptr.To("http:test.example.com"),
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "webhook-with-template-url",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
							{
								URL: ptr.To("{{ .labels.url }}"),
							},
						},
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "webhook-with-invalid-template-url",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
							{
								URL: ptr.To("{{ .labels.value "),
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "webhook-with-url-secret",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
							{
								URLSecret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
									Key:                  "key1",
								},
							},
						},
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "webhook-with-template-url-secret",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
							{
								URLSecret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
									Key:                  "template-url",
								},
							},
						},
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "webhook-with-wrong-url-secret",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
							{
								URLSecret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
									Key:                  "not-existing",
								},
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "webhook-with-invalid-template-url-secret",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
							{
								URLSecret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
									Key:                  "invalid-template-url",
								},
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wechat-valid",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						WeChatConfigs: []monitoringv1alpha1.WeChatConfig{
							{
								CorpID: ptr.To("testingCorpID"),
							},
						},
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wechat-invalid-url",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						WeChatConfigs: []monitoringv1alpha1.WeChatConfig{
							{
								CorpID: ptr.To("testingCorpID"),
								APIURL: ptr.To(monitoringv1alpha1.URL("http://::invalid-url")),
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wechat-invalid-secret",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						WeChatConfigs: []monitoringv1alpha1.WeChatConfig{
							{
								CorpID: ptr.To("testingCorpID"),
								APISecret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
									Key:                  "not-existing",
								},
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wechat-valid-secret",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						WeChatConfigs: []monitoringv1alpha1.WeChatConfig{
							{
								CorpID: ptr.To("testingCorpID"),
								APISecret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
									Key:                  "key1",
								},
							},
						},
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-empty-slack",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						SlackConfigs: []monitoringv1alpha1.SlackConfig{
							{},
						},
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "slack-with-empty-action-type",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						SlackConfigs: []monitoringv1alpha1.SlackConfig{
							{
								Actions: []monitoringv1alpha1.SlackAction{
									{
										Type: "",
									},
								},
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "slack-with-empty-action-text",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						SlackConfigs: []monitoringv1alpha1.SlackConfig{
							{
								Actions: []monitoringv1alpha1.SlackAction{
									{
										Type: "type",
										Text: "",
									},
								},
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "slack-with-empty-action-url-and-name",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						SlackConfigs: []monitoringv1alpha1.SlackConfig{
							{
								Actions: []monitoringv1alpha1.SlackAction{
									{
										Type: "type",
										Text: "text",
									},
								},
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "slack-with-valid-action-url",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						SlackConfigs: []monitoringv1alpha1.SlackConfig{
							{
								Actions: []monitoringv1alpha1.SlackAction{
									{
										Type: "type",
										Text: "text",
										URL:  "http://localhost",
									},
								},
							},
						},
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "slack-with-valid-action-name",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						SlackConfigs: []monitoringv1alpha1.SlackConfig{
							{
								Actions: []monitoringv1alpha1.SlackAction{
									{
										Type: "type",
										Text: "text",
										Name: ptr.To("my-action"),
									},
								},
							},
						},
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "slack-with-invalid-action-confirm-field",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						SlackConfigs: []monitoringv1alpha1.SlackConfig{
							{
								Actions: []monitoringv1alpha1.SlackAction{
									{
										Type: "type",
										Text: "text",
										Name: ptr.To("my-action"),
										ConfirmField: &monitoringv1alpha1.SlackConfirmationField{
											Text: "",
										},
									},
								},
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "slack-with-vali-action-confirm-field",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						SlackConfigs: []monitoringv1alpha1.SlackConfig{
							{
								Actions: []monitoringv1alpha1.SlackAction{
									{
										Type: "type",
										Text: "text",
										Name: ptr.To("my-action"),
										ConfirmField: &monitoringv1alpha1.SlackConfirmationField{
											Text: "text",
										},
									},
								},
							},
						},
					}},
				},
			},
			ok: true,
		}, {
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "slack-with-empty-field-title",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						SlackConfigs: []monitoringv1alpha1.SlackConfig{
							{
								Fields: []monitoringv1alpha1.SlackField{
									{
										Title: "",
									},
								},
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "slack-with-empty-field-value",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						SlackConfigs: []monitoringv1alpha1.SlackConfig{
							{
								Fields: []monitoringv1alpha1.SlackField{
									{
										Title: "title",
										Value: "",
									},
								},
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "slack-with-valid-field",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						SlackConfigs: []monitoringv1alpha1.SlackConfig{
							{
								Fields: []monitoringv1alpha1.SlackField{
									{
										Title: "title",
										Value: "value",
									},
								},
							},
						},
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "slack-with-valid-action-and-field",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						SlackConfigs: []monitoringv1alpha1.SlackConfig{
							{
								Actions: []monitoringv1alpha1.SlackAction{
									{
										Type: "type",
										Text: "text",
										Name: ptr.To("my-action"),
										ConfirmField: &monitoringv1alpha1.SlackConfirmationField{
											Text: "text",
										},
									},
								},
								Fields: []monitoringv1alpha1.SlackField{
									{
										Title: "title",
										Value: "value",
									},
								},
							},
						},
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "subroute-with-unknown-field",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: []byte(`{"receiver": "recv2", "matchers": [{"severity":"!=critical$"}]}`),
							},
						},
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
					}, {
						Name: "recv2",
					}},
				},
			},
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "slack-with-invalid-url-in-secret",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						SlackConfigs: []monitoringv1alpha1.SlackConfig{
							{
								APIURL: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
									Key:                  "invalid-url",
								},
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "subroute-with-invalid-matcher",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: []byte(`{"receiver": "recv2", "matchers": [{"name": "severity", "value": "critical", "matchType": "!!"}]}`),
							},
						},
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
					}, {
						Name: "recv2",
					}},
				},
			},
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "subroute-with-empty-matcher-name",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: []byte(`{"receiver": "recv2", "matchers": [{"name": "", "value": "critical", "matchType": "!="}]}`),
							},
						},
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
					}, {
						Name: "recv2",
					}},
				},
			},
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "subroute-with-missing-receiver",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: []byte(`{"receiver": "recv2", "matchers": [{"name": "severity", "value": "critical", "matchType": "!="}]}`),
							},
						},
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
					}},
				},
			},
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-subroute-definition",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: []byte(`{"receiver": "recv2", "matchers": [{"name": "severity", "value": "critical", "matchType": "!="}]}`),
							},
						},
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
					}, {
						Name: "recv2",
					}},
				},
			},
			ok: true,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "discord-with-invalid-url-in-secret",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						DiscordConfigs: []monitoringv1alpha1.DiscordConfig{
							{
								APIURL: v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
									Key:                  "invalid-url",
								},
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sns-with-invalid-api-url",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						SNSConfigs: []monitoringv1alpha1.SNSConfig{
							{
								ApiURL: ptr.To("https:://sns.us-east-2.amazonaws.com"),
								Sigv4: &monitoringv1.Sigv4{
									Region:  "us-east-2",
									RoleArn: "test-roleARN",
								},
								TopicARN: ptr.To("test-topicARN"),
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "msteams-with-missing-webhook-url-secret",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						MSTeamsConfigs: []monitoringv1alpha1.MSTeamsConfig{
							{
								WebhookURL: v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{Name: "not-existing-secret"},
									Key:                  "url",
								},
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "msteams-with-missing-webhook-url-key",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						MSTeamsConfigs: []monitoringv1alpha1.MSTeamsConfig{
							{
								WebhookURL: v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
									Key:                  "not-existing",
								},
							},
						},
					}},
				},
			},
			ok: false,
		},
		{
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "msteams-with-valid-webhook-url-secret",
					Namespace: "ns1",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Route: &monitoringv1alpha1.Route{
						Receiver: "recv1",
					},
					Receivers: []monitoringv1alpha1.Receiver{{
						Name: "recv1",
						MSTeamsConfigs: []monitoringv1alpha1.MSTeamsConfig{
							{
								WebhookURL: v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
									Key:                  "key1",
								},
							},
						},
					}},
				},
			},
			ok: true,
		},
	} {
		t.Run(tc.amConfig.Name, func(t *testing.T) {
			store := assets.NewStoreBuilder(c.CoreV1(), c.CoreV1())

			err := checkAlertmanagerConfigResource(context.Background(), tc.amConfig, version, store)
			if tc.ok {
				require.NoError(t, err)
				return
			}

			t.Logf("err: %s", err)
			require.Error(t, err)
		})
	}
}

// Test to exercise the function provisionAlertmanagerConfiguration
// and validate that the operator is able to generate an Alertmanager
// configuration depending on the method chosen by the user.
// Alertmanager can be configured using either the AlertmanagerConfig resource
// or a Kubernetes secret.
func TestProvisionAlertmanagerConfiguration(t *testing.T) {
	for _, tc := range []struct {
		am      *monitoringv1.Alertmanager
		objects []runtime.Object

		ok           bool
		expectedKeys []string
	}{
		{
			am: &monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "empty",
					Namespace: "test",
				},
			},
			ok: true,
		},
		{
			am: &monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "empty-with-selector",
					Namespace: "test",
				},
				Spec: monitoringv1.AlertmanagerSpec{
					AlertmanagerConfigSelector: &metav1.LabelSelector{},
				},
			},
			ok: true,
		},
		{
			am: &monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-user-config-in-secret-with-no-config-selector",
					Namespace: "test",
				},
				Spec: monitoringv1.AlertmanagerSpec{
					ConfigSecret: "amconfig",
				},
			},
			objects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "amconfig",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"alertmanager.yaml": []byte(`invalid`),
					},
				},
			},
			ok: true,
		},
		{
			am: &monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-user-config-provided-to-operator",
					Namespace: "test",
				},
				Spec: monitoringv1.AlertmanagerSpec{
					ConfigSecret: "amconfig",
					AlertmanagerConfigSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"test": "test"},
					},
				},
			},
			objects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "amconfig",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"alertmanager.yaml": []byte(`invalid`),
					},
				},
			},
			ok: false,
		},
		{
			am: &monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-user-config",
					Namespace: "test",
				},
				Spec: monitoringv1.AlertmanagerSpec{
					ConfigSecret: "amconfig",
				},
			},
			objects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "amconfig",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"alertmanager.yaml": []byte(`{route: {receiver: empty}, receivers: [{name: empty}]}`),
					},
				},
			},
			ok: true,
		},
		{
			am: &monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-user-config-with-selector",
					Namespace: "test",
				},
				Spec: monitoringv1.AlertmanagerSpec{
					ConfigSecret:               "amconfig",
					AlertmanagerConfigSelector: &metav1.LabelSelector{},
				},
			},
			objects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "amconfig",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"alertmanager.yaml": []byte(`{route: {receiver: empty}, receivers: [{name: empty}]}`),
					},
				},
			},
			ok: true,
		},
		{
			am: &monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-user-config-and-additional-secret",
					Namespace: "test",
				},
				Spec: monitoringv1.AlertmanagerSpec{
					ConfigSecret: "amconfig",
				},
			},
			objects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "amconfig",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"alertmanager.yaml": []byte(`{route: {receiver: empty}, receivers: [{name: empty}]}`),
						"key1":              []byte(`val1`),
					},
				},
			},
			ok:           true,
			expectedKeys: []string{"key1"},
		},
		{
			am: &monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-user-config-and-additional-secret-with-selectors",
					Namespace: "test",
				},
				Spec: monitoringv1.AlertmanagerSpec{
					ConfigSecret:               "amconfig",
					AlertmanagerConfigSelector: &metav1.LabelSelector{},
				},
			},
			objects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "amconfig",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"alertmanager.yaml": []byte(`{route: {receiver: empty}, receivers: [{name: empty}]}`),
						"key1":              []byte(`val1`),
					},
				},
			},
			ok:           true,
			expectedKeys: []string{"key1"},
		},
		{
			am: &monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-user-config-but-additional-secret",
					Namespace: "test",
				},
				Spec: monitoringv1.AlertmanagerSpec{
					ConfigSecret: "amconfig",
				},
			},
			objects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "amconfig",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"key1": []byte(`val1`),
					},
				},
			},
			ok:           true,
			expectedKeys: []string{"key1"},
		},
		{
			am: &monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-user-config-but-additional-secret-with-selectors",
					Namespace: "test",
				},
				Spec: monitoringv1.AlertmanagerSpec{
					ConfigSecret:               "amconfig",
					AlertmanagerConfigSelector: &metav1.LabelSelector{},
				},
			},
			objects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "amconfig",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"key1": []byte(`val1`),
					},
				},
			},
			ok:           true,
			expectedKeys: []string{"key1"},
		},
	} {
		t.Run(tc.am.Name, func(t *testing.T) {
			c := fake.NewSimpleClientset(tc.objects...)

			o := &Operator{
				kclient:          c,
				mclient:          monitoringfake.NewSimpleClientset(),
				ssarClient:       &alwaysAllowed{},
				logger:           slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
				metrics:          operator.NewMetrics(prometheus.NewRegistry()),
				newEventRecorder: func(related runtime.Object) *operator.EventRecorder { return operator.NewFakeRecorder(1, related) },
			}

			err := o.bootstrap(
				context.Background(),
				operator.Config{
					Namespaces: operator.Namespaces{
						AlertmanagerConfigAllowList: map[string]struct{}{
							v1.NamespaceAll: {},
						},
						AlertmanagerAllowList: map[string]struct{}{
							"foo": {},
						},
					},
				},
			)
			require.NoError(t, err)

			store := assets.NewStoreBuilder(c.CoreV1(), c.CoreV1())
			err = o.provisionAlertmanagerConfiguration(context.Background(), tc.am, store)

			if !tc.ok {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			secret, err := c.CoreV1().Secrets(tc.am.Namespace).Get(context.Background(), generatedConfigSecretName(tc.am.Name), metav1.GetOptions{})
			require.NoError(t, err)

			expected := append(tc.expectedKeys, alertmanagerConfigFileCompressed)
			require.Equal(t, len(secret.Data), len(expected), "expecting %d items to be present in the generated secret but got %d", len(expected), len(secret.Data))

			for _, k := range expected {
				_, found := secret.Data[k]
				require.True(t, found, "expecting key %q to be present in the generated secret but got nothing", k)
			}
		})
	}
}

// alwaysAllowed implements SelfSubjectAccessReviewInterface.
type alwaysAllowed struct{}

func (*alwaysAllowed) Create(_ context.Context, _ *authv1.SelfSubjectAccessReview, _ metav1.CreateOptions) (*authv1.SelfSubjectAccessReview, error) {
	return &authv1.SelfSubjectAccessReview{
		Status: authv1.SubjectAccessReviewStatus{
			Allowed: true,
		},
	}, nil
}
