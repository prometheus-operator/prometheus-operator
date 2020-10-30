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
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
)

func strPtr(str string) *string {
	return &str
}

func TestCheckAlertmanagerConfig(t *testing.T) {
	c := fake.NewSimpleClientset(
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret",
				Namespace: "ns1",
			},
			Data: map[string][]byte{
				"key1": []byte("val1"),
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
						Routes: []monitoringv1alpha1.Route{{
							Receiver: "recv2",
						}},
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
								URL: strPtr("http://test.local"),
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
								CorpID: strPtr("testingCorpID"),
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
								CorpID: strPtr("testingCorpID"),
								APIURL: strPtr("http://::invalid-url"),
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
								CorpID: strPtr("testingCorpID"),
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
								CorpID: strPtr("testingCorpID"),
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
	} {
		t.Run(tc.amConfig.Name, func(t *testing.T) {
			store := assets.NewStore(c.CoreV1(), c.CoreV1())
			err := checkAlertmanagerConfig(context.Background(), tc.amConfig, store)
			if tc.ok {
				if err != nil {
					t.Fatalf("expecting no error but got %q", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expecting error but got none")
			}
		})
	}
}

func TestListOptions(t *testing.T) {
	for i := 0; i < 1000; i++ {
		o := ListOptions("test")
		if o.LabelSelector != "app=alertmanager,alertmanager=test" && o.LabelSelector != "alertmanager=test,app=alertmanager" {
			t.Fatalf("LabelSelector not computed correctly\n\nExpected: \"app=alertmanager,alertmanager=test\"\n\nGot:      %#+v", o.LabelSelector)
		}
	}
}
