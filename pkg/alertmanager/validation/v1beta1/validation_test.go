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

package v1beta1

import (
	"testing"

	monitoringv1beta1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1beta1"
	"k8s.io/utils/pointer"
)

func TestValidateAlertmanagerConfig(t *testing.T) {
	testCases := []struct {
		name      string
		in        *monitoringv1beta1.AlertmanagerConfig
		expectErr bool
	}{
		{
			name: "Test fail to validate on duplicate receiver",
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
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
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							OpsGenieConfigs: []monitoringv1beta1.OpsGenieConfig{
								{
									Responders: []monitoringv1beta1.OpsGenieConfigResponder{
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
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							SlackConfigs: []monitoringv1beta1.SlackConfig{
								{
									Actions: []monitoringv1beta1.SlackAction{
										{
											Type: "a",
											Text: "b",
											URL:  "www.test.com",
											Name: "c",
											ConfirmField: &monitoringv1beta1.SlackConfirmationField{
												Text: "d",
											},
										},
									},
									Fields: []monitoringv1beta1.SlackField{
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
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							WebhookConfigs: []monitoringv1beta1.WebhookConfig{
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
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							WeChatConfigs: []monitoringv1beta1.WeChatConfig{
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
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							EmailConfigs: []monitoringv1beta1.EmailConfig{
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
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							EmailConfigs: []monitoringv1beta1.EmailConfig{
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
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							VictorOpsConfigs: []monitoringv1beta1.VictorOpsConfig{
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
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							VictorOpsConfigs: []monitoringv1beta1.VictorOpsConfig{
								{
									RoutingKey: "a",
									CustomFields: []monitoringv1beta1.KeyValue{
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
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							PushoverConfigs: []monitoringv1beta1.PushoverConfig{
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
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							PushoverConfigs: []monitoringv1beta1.PushoverConfig{
								{
									UserKey: &monitoringv1beta1.SecretKeySelector{
										Name: "creds",
										Key:  "user",
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
			name: "Test fail to validate routes - parent route has no receiver",
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
					},
					Route: &monitoringv1beta1.Route{
						Receiver: "will-not-be-found",
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate routes with duplicate groupBy",
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
					},
					Route: &monitoringv1beta1.Route{
						Receiver: "same",
						GroupBy:  []string{"job", "job"},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate routes with exclusive value and other in groupBy",
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
					},
					Route: &monitoringv1beta1.Route{
						Receiver: "same",
						GroupBy:  []string{"job", "..."},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test fail to validate routes - named mute time interval does not exist",
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
					},
					Route: &monitoringv1beta1.Route{
						Receiver:          "same",
						MuteTimeIntervals: []string{"awol"},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test happy path",
			in: &monitoringv1beta1.AlertmanagerConfig{
				Spec: monitoringv1beta1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1beta1.Receiver{
						{
							Name: "same",
						},
						{
							Name: "different",
							OpsGenieConfigs: []monitoringv1beta1.OpsGenieConfig{
								{
									Responders: []monitoringv1beta1.OpsGenieConfigResponder{
										{
											ID:       "a",
											Name:     "b",
											Username: "c",
										},
									},
								},
							},
							SlackConfigs: []monitoringv1beta1.SlackConfig{
								{
									Actions: []monitoringv1beta1.SlackAction{
										{
											Type: "a",
											Text: "b",
											URL:  "https://www.test.com",
											Name: "c",
											ConfirmField: &monitoringv1beta1.SlackConfirmationField{
												Text: "d",
											},
										},
									},
									Fields: []monitoringv1beta1.SlackField{
										{
											Title: "a",
											Value: "b",
										},
									},
								},
							},
							WebhookConfigs: []monitoringv1beta1.WebhookConfig{
								{
									URL: pointer.String("https://www.test.com"),
									URLSecret: &monitoringv1beta1.SecretKeySelector{
										Name: "creds",
										Key:  "url",
									},
								},
							},
							WeChatConfigs: []monitoringv1beta1.WeChatConfig{
								{
									APIURL: "https://test.com",
								},
							},
							EmailConfigs: []monitoringv1beta1.EmailConfig{
								{
									To:        "a",
									Smarthost: "b:8080",
									Headers: []monitoringv1beta1.KeyValue{
										{
											Key:   "c",
											Value: "d",
										},
									},
								},
							},
							VictorOpsConfigs: []monitoringv1beta1.VictorOpsConfig{
								{
									RoutingKey: "a",
									CustomFields: []monitoringv1beta1.KeyValue{
										{
											Key:   "b",
											Value: "c",
										},
									},
								},
							},
							PushoverConfigs: []monitoringv1beta1.PushoverConfig{
								{
									UserKey: &monitoringv1beta1.SecretKeySelector{
										Name: "creds",
										Key:  "user",
									},
									Token: &monitoringv1beta1.SecretKeySelector{
										Name: "creds",
										Key:  "token",
									},
									Retry:  "10m",
									Expire: "5m",
								},
							},
						},
					},
					Route: &monitoringv1beta1.Route{
						Receiver:          "same",
						GroupBy:           []string{"..."},
						MuteTimeIntervals: []string{"weekdays-only"},
					},
					TimeIntervals: []monitoringv1beta1.TimeInterval{
						{
							Name: "weekdays-only",
							TimeIntervals: []monitoringv1beta1.TimePeriod{
								{
									Weekdays: []monitoringv1beta1.WeekdayRange{
										monitoringv1beta1.WeekdayRange("Saturday"),
										monitoringv1beta1.WeekdayRange("Sunday"),
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
