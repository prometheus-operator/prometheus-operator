// Copyright The prometheus-operator Authors
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
	"encoding/json"
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

func TestAlertmanagerConfigConversion(t *testing.T) {
	for _, tc := range []struct {
		name         string
		v1beta1Spec  AlertmanagerConfigSpec
		v1alpha1Spec v1alpha1.AlertmanagerConfigSpec
	}{
		{
			name: "email without threading",
			v1beta1Spec: AlertmanagerConfigSpec{
				Receivers: []Receiver{{
					Name:         "email",
					EmailConfigs: []EmailConfig{{To: new("team@example.com")}},
				}},
			},
			v1alpha1Spec: v1alpha1.AlertmanagerConfigSpec{
				Receivers: []v1alpha1.Receiver{{
					Name:         "email",
					EmailConfigs: []v1alpha1.EmailConfig{{To: new("team@example.com")}},
				}},
			},
		},
		{
			name: "email with threading",
			v1beta1Spec: AlertmanagerConfigSpec{
				Receivers: []Receiver{{
					Name: "email",
					EmailConfigs: []EmailConfig{{
						To:        new("team@example.com"),
						Threading: &EmailThreadingConfig{ThreadByDate: ThreadByDateTypeDaily},
					}},
				}},
			},
			v1alpha1Spec: v1alpha1.AlertmanagerConfigSpec{
				Receivers: []v1alpha1.Receiver{{
					Name: "email",
					EmailConfigs: []v1alpha1.EmailConfig{{
						To:        new("team@example.com"),
						Threading: &v1alpha1.EmailThreadingConfig{ThreadByDate: v1alpha1.ThreadByDateTypeDaily},
					}},
				}},
			},
		},
		{
			name: "pushover with ttl",
			v1beta1Spec: AlertmanagerConfigSpec{
				Receivers: []Receiver{{
					Name:            "pushover",
					PushoverConfigs: []PushoverConfig{{TTL: new(monitoringv1.Duration("1h"))}},
				}},
			},
			v1alpha1Spec: v1alpha1.AlertmanagerConfigSpec{
				Receivers: []v1alpha1.Receiver{{
					Name:            "pushover",
					PushoverConfigs: []v1alpha1.PushoverConfig{{TTL: new(monitoringv1.Duration("1h"))}},
				}},
			},
		},
		{
			name: "all receiver types with optional fields unset",
			v1beta1Spec: AlertmanagerConfigSpec{
				Receivers: []Receiver{{
					Name:              "all",
					OpsGenieConfigs:   []OpsGenieConfig{{}},
					PagerDutyConfigs:  []PagerDutyConfig{{}},
					DiscordConfigs:    []DiscordConfig{{}},
					SlackConfigs:      []SlackConfig{{}},
					WebhookConfigs:    []WebhookConfig{{}},
					WeChatConfigs:     []WeChatConfig{{}},
					EmailConfigs:      []EmailConfig{{}},
					VictorOpsConfigs:  []VictorOpsConfig{{}},
					PushoverConfigs:   []PushoverConfig{{}},
					SNSConfigs:        []SNSConfig{{}},
					TelegramConfigs:   []TelegramConfig{{}},
					WebexConfigs:      []WebexConfig{{}},
					MSTeamsConfigs:    []MSTeamsConfig{{}},
					MSTeamsV2Configs:  []MSTeamsV2Config{{}},
					RocketChatConfigs: []RocketChatConfig{{}},
				}},
			},
			v1alpha1Spec: v1alpha1.AlertmanagerConfigSpec{
				Receivers: []v1alpha1.Receiver{{
					Name:              "all",
					OpsGenieConfigs:   []v1alpha1.OpsGenieConfig{{}},
					PagerDutyConfigs:  []v1alpha1.PagerDutyConfig{{}},
					DiscordConfigs:    []v1alpha1.DiscordConfig{{}},
					SlackConfigs:      []v1alpha1.SlackConfig{{}},
					WebhookConfigs:    []v1alpha1.WebhookConfig{{}},
					WeChatConfigs:     []v1alpha1.WeChatConfig{{}},
					EmailConfigs:      []v1alpha1.EmailConfig{{}},
					VictorOpsConfigs:  []v1alpha1.VictorOpsConfig{{}},
					PushoverConfigs:   []v1alpha1.PushoverConfig{{}},
					SNSConfigs:        []v1alpha1.SNSConfig{{}},
					TelegramConfigs:   []v1alpha1.TelegramConfig{{}},
					WebexConfigs:      []v1alpha1.WebexConfig{{}},
					MSTeamsConfigs:    []v1alpha1.MSTeamsConfig{{}},
					MSTeamsV2Configs:  []v1alpha1.MSTeamsV2Config{{}},
					RocketChatConfigs: []v1alpha1.RocketChatConfig{{}},
				}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("v1beta1 to v1alpha1", func(t *testing.T) {
				src := &AlertmanagerConfig{Spec: tc.v1beta1Spec}
				dst := &v1alpha1.AlertmanagerConfig{}

				if err := src.ConvertTo(dst); err != nil {
					t.Fatalf("expected no error but got %v", err)
				}

				assertJSONEqual(t, tc.v1alpha1Spec, dst.Spec)
			})

			t.Run("v1alpha1 to v1beta1", func(t *testing.T) {
				src := &v1alpha1.AlertmanagerConfig{Spec: tc.v1alpha1Spec}
				dst := &AlertmanagerConfig{}

				if err := dst.ConvertFrom(src); err != nil {
					t.Fatalf("expected no error but got %v", err)
				}

				assertJSONEqual(t, tc.v1beta1Spec, dst.Spec)
			})
		})
	}
}

func TestConvertRouteLabels(t *testing.T) {
	tests := []struct {
		name           string
		inputLabels    []KeyValue
		expectedLabels []KeyValue
	}{
		{
			name:           "nil labels",
			inputLabels:    nil,
			expectedLabels: nil,
		},
		{
			name:           "empty labels",
			inputLabels:    []KeyValue{},
			expectedLabels: []KeyValue{},
		},
		{
			name: "simple labels",
			inputLabels: []KeyValue{
				{Key: "severity", Value: "critical"},
				{Key: "team", Value: "platform"},
			},
			expectedLabels: []KeyValue{
				{Key: "severity", Value: "critical"},
				{Key: "team", Value: "platform"},
			},
		},
		{
			name: "labels with template expressions",
			inputLabels: []KeyValue{
				{Key: "environment", Value: "{{ .Labels.env }}"},
				{Key: "region", Value: "{{ .ExternalLabels.region }}"},
			},
			expectedLabels: []KeyValue{
				{Key: "environment", Value: "{{ .Labels.env }}"},
				{Key: "region", Value: "{{ .ExternalLabels.region }}"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("convertRouteFrom preserves labels", func(t *testing.T) {
				var alphaLabels []v1alpha1.KeyValue
				if tc.inputLabels != nil {
					alphaLabels = make([]v1alpha1.KeyValue, len(tc.inputLabels))
					for i, kv := range tc.inputLabels {
						alphaLabels[i] = v1alpha1.KeyValue{Key: kv.Key, Value: kv.Value}
					}
				}

				alphaRoute := &v1alpha1.Route{
					Receiver: "test",
					Labels:   alphaLabels,
				}

				betaRoute, err := convertRouteFrom(alphaRoute)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if tc.inputLabels == nil {
					if betaRoute.Labels != nil {
						t.Fatalf("expected nil labels, got %v", betaRoute.Labels)
					}
					return
				}

				if len(betaRoute.Labels) != len(tc.expectedLabels) {
					t.Fatalf("expected %d labels, got %d", len(tc.expectedLabels), len(betaRoute.Labels))
				}

				for i, expected := range tc.expectedLabels {
					if betaRoute.Labels[i].Key != expected.Key || betaRoute.Labels[i].Value != expected.Value {
						t.Errorf("expected label[%d] = {%s, %s}, got {%s, %s}",
							i, expected.Key, expected.Value,
							betaRoute.Labels[i].Key, betaRoute.Labels[i].Value)
					}
				}
			})

			t.Run("convertRouteTo preserves labels", func(t *testing.T) {
				betaRoute := &Route{
					Receiver: "test",
					Labels:   tc.inputLabels,
				}

				alphaRoute, err := convertRouteTo(betaRoute)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if tc.inputLabels == nil {
					if alphaRoute.Labels != nil {
						t.Fatalf("expected nil labels, got %v", alphaRoute.Labels)
					}
					return
				}

				if len(alphaRoute.Labels) != len(tc.expectedLabels) {
					t.Fatalf("expected %d labels, got %d", len(tc.expectedLabels), len(alphaRoute.Labels))
				}

				for i, expected := range tc.expectedLabels {
					if alphaRoute.Labels[i].Key != expected.Key || alphaRoute.Labels[i].Value != expected.Value {
						t.Errorf("expected label[%d] = {%s, %s}, got {%s, %s}",
							i, expected.Key, expected.Value,
							alphaRoute.Labels[i].Key, alphaRoute.Labels[i].Value)
					}
				}
			})
		})
	}
}

// assertJSONEqual compares the JSON representations because the conversion
// functions return empty slices for nil inputs which serialize the same way.
func assertJSONEqual(t *testing.T, expected, got any) {
	t.Helper()

	e, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("failed to marshal expected value: %v", err)
	}

	g, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("failed to marshal got value: %v", err)
	}

	if string(e) != string(g) {
		t.Fatalf("wanted %s, but got %s", e, g)
	}
}
