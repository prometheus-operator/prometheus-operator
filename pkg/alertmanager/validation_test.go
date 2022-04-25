// Copyright 2021 The prometheus-operator Authors
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
	"net/url"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"

	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus/alertmanager/config"
)

func TestValidateAlertmanagerConfig(t *testing.T) {
	testCases := []struct {
		name      string
		in        *monitoringv1alpha1.AlertmanagerConfig
		expectErr bool
	}{
		{
			name: "Test fail to validate on duplicate receiver",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "same",
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate on opsgenie config - missing required fields",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							OpsGenieConfigs: []monitoringv1alpha1.OpsGenieConfig{
								{
									Responders: []monitoringv1alpha1.OpsGenieConfigResponder{
										{},
									},
								},
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate on slack config - valid action fields - invalid fields field",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							SlackConfigs: []monitoringv1alpha1.SlackConfig{
								{
									Actions: []monitoringv1alpha1.SlackAction{
										{
											Type: "a",
											Text: "b",
											URL:  "www.test.com",
											Name: "c",
											ConfirmField: &monitoringv1alpha1.SlackConfirmationField{
												Text: "d",
											},
										},
									},
									Fields: []monitoringv1alpha1.SlackField{
										{},
									},
								},
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate webhook config - missing required fields",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
								{},
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate wechat config - invalid URL",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							WeChatConfigs: []monitoringv1alpha1.WeChatConfig{
								{
									APIURL: "http://%><invalid.com",
								},
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate email config - missing to field",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							EmailConfigs: []monitoringv1alpha1.EmailConfig{
								{},
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate email config - invalid smarthost",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							EmailConfigs: []monitoringv1alpha1.EmailConfig{
								{
									To:        "a",
									Smarthost: "invalid",
								},
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate VictorOpsConfigs - missing routing key",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							VictorOpsConfigs: []monitoringv1alpha1.VictorOpsConfig{
								{},
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate VictorOpsConfigs - reservedFields",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							VictorOpsConfigs: []monitoringv1alpha1.VictorOpsConfig{
								{
									RoutingKey: "a",
									CustomFields: []monitoringv1alpha1.KeyValue{
										{
											Key:   "routing_key",
											Value: "routing_key",
										},
									},
								},
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate PushoverConfigs - missing user key",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							PushoverConfigs: []monitoringv1alpha1.PushoverConfig{
								{},
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate PushoverConfigs - missing token",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							PushoverConfigs: []monitoringv1alpha1.PushoverConfig{
								{
									UserKey: &v1.SecretKeySelector{},
								},
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate routes - parent route has no receiver",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "will-not-be-found",
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate routes with duplicate groupBy",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "same",
						GroupBy:  []string{"job", "job"},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate routes with exclusive value and other in groupBy",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "same",
						GroupBy:  []string{"job", "..."},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate routes - named mute time interval does not exist",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver:          "same",
						MuteTimeIntervals: []string{"awol"},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test happy path",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							OpsGenieConfigs: []monitoringv1alpha1.OpsGenieConfig{
								{
									Responders: []monitoringv1alpha1.OpsGenieConfigResponder{
										{
											ID:       "a",
											Name:     "b",
											Username: "c",
										},
									},
								},
							},
							SlackConfigs: []monitoringv1alpha1.SlackConfig{
								{
									Actions: []monitoringv1alpha1.SlackAction{
										{
											Type: "a",
											Text: "b",
											URL:  "https://www.test.com",
											Name: "c",
											ConfirmField: &monitoringv1alpha1.SlackConfirmationField{
												Text: "d",
											},
										},
									},
									Fields: []monitoringv1alpha1.SlackField{
										{
											Title: "a",
											Value: "b",
										},
									},
								},
							},
							WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
								{
									URL:       strToPtr("https://www.test.com"),
									URLSecret: &v1.SecretKeySelector{},
								},
							},
							WeChatConfigs: []monitoringv1alpha1.WeChatConfig{
								{
									APIURL: "https://test.com",
								},
							},
							EmailConfigs: []monitoringv1alpha1.EmailConfig{
								{
									To:        "a",
									Smarthost: "b:8080",
									Headers: []monitoringv1alpha1.KeyValue{
										{
											Key:   "c",
											Value: "d",
										},
									},
								},
							},
							VictorOpsConfigs: []monitoringv1alpha1.VictorOpsConfig{
								{
									RoutingKey: "a",
									CustomFields: []monitoringv1alpha1.KeyValue{
										{
											Key:   "b",
											Value: "c",
										},
									},
								},
							},
							PushoverConfigs: []monitoringv1alpha1.PushoverConfig{
								{
									UserKey: &v1.SecretKeySelector{},
									Token:   &v1.SecretKeySelector{},
									Retry:   "10m",
									Expire:  "5m",
								},
							},
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver:          "same",
						GroupBy:           []string{"..."},
						MuteTimeIntervals: []string{"weekdays-only"},
					},
					MuteTimeIntervals: []monitoringv1alpha1.MuteTimeInterval{
						{
							Name: "weekdays-only",
							TimeIntervals: []monitoringv1alpha1.TimeInterval{
								{
									Weekdays: []monitoringv1alpha1.WeekdayRange{
										monitoringv1alpha1.WeekdayRange("Saturday"),
										monitoringv1alpha1.WeekdayRange("Sunday"),
									},
								},
							},
						},
					},
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAlertmanagerConfig(tc.in)
			if tc.expectErr && err == nil {
				t.Error("expected error but got none")
			}

			if err != nil {
				if tc.expectErr {
					return
				}
				t.Errorf("got error but expected none -%s", err.Error())
			}
		})
	}
}

func TestValidateUrl(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		expectErr    bool
		expectResult func() *config.URL
	}{
		{
			name:      "Test invalid url returns error",
			in:        "https://!^invalid.com",
			expectErr: true,
		},
		{
			name:      "Test missing scheme returns error",
			in:        "is.normally.valid",
			expectErr: true,
		},
		{
			name: "Test happy path",
			in:   "https://u:p@is.compliant.with.upstream.unmarshal",
			expectResult: func() *config.URL {
				u, _ := url.Parse("https://u:p@is.compliant.with.upstream.unmarshal")
				return &config.URL{URL: u}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := ValidateURL(tc.in)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			res := tc.expectResult()
			if !reflect.DeepEqual(u, res) {
				t.Fatalf("wanted %v but got %v", res, u)
			}
		})
	}
}

func strToPtr(s string) *string {
	return &s
}
