// Copyright 2020 The prometheus-operator Authors
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
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/timeinterval"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/pointer"

	monitoringingv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
)

func TestInitializeFromAlertmanagerConfig(t *testing.T) {
	myroute := monitoringv1alpha1.Route{
		Receiver: "myreceiver",
		Matchers: []monitoringv1alpha1.Matcher{
			{
				Name:  "mykey",
				Value: "myvalue",
				Regex: false,
			},
		},
	}

	myrouteJSON, _ := json.Marshal(myroute)
	pagerdutyURL := "example.pagerduty.com"
	invalidPagerdutyURL := "://example.pagerduty.com"

	tests := []struct {
		name            string
		globalConfig    *monitoringingv1.AlertmanagerGlobalConfig
		matcherStrategy monitoringingv1.AlertmanagerConfigMatcherStrategy
		amConfig        *monitoringv1alpha1.AlertmanagerConfig
		want            *alertmanagerConfig
		wantErr         bool
	}{
		{
			name: "valid global config",
			globalConfig: &monitoringingv1.AlertmanagerGlobalConfig{
				SMTPConfig: &monitoringingv1.GlobalSMTPConfig{
					From: pointer.String("from"),
					SmartHost: &monitoringingv1.HostPort{
						Host: "smtp.example.org",
						Port: "587",
					},
					Hello:        pointer.String("smtp.example.org"),
					AuthUsername: pointer.String("dev@smtp.example.org"),
					AuthPassword: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "smtp-auth",
						},
						Key: "password",
					},
					AuthIdentity: pointer.String("dev@smtp.example.org"),
					AuthSecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "smtp-auth",
						},
						Key: "secret",
					},
					RequireTLS: pointer.Bool(true),
				},
				ResolveTimeout: "30s",
				HTTPConfig: &monitoringingv1.HTTPConfig{
					OAuth2: &monitoringingv1.OAuth2{
						ClientID: monitoringingv1.SecretOrConfigMap{
							ConfigMap: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "webhook-client-id",
								},
								Key: "test",
							},
						},
						ClientSecret: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "webhook-client-secret",
							},
							Key: "test",
						},
						TokenURL: "https://test.com",
						Scopes:   []string{"any"},
						EndpointParams: map[string]string{
							"some": "value",
						},
					},
					FollowRedirects: pointer.Bool(true),
				},
			},
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "global-config",
					Namespace: "mynamespace",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "null",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []v1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringingv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			want: &alertmanagerConfig{
				Global: &globalConfig{
					ResolveTimeout: func(d model.Duration) *model.Duration { return &d }(model.Duration(30 * time.Second)),
					SMTPFrom:       "from",
					SMTPSmarthost: config.HostPort{
						Host: "smtp.example.org",
						Port: "587",
					},
					SMTPHello:        "smtp.example.org",
					SMTPAuthUsername: "dev@smtp.example.org",
					SMTPAuthPassword: "password",
					SMTPAuthIdentity: "dev@smtp.example.org",
					SMTPAuthSecret:   "secret",
					SMTPRequireTLS:   pointer.Bool(true),
					HTTPConfig: &httpClientConfig{
						OAuth2: &oauth2{
							ClientID:     "clientID",
							ClientSecret: "clientSecret",
							Scopes:       []string{"any"},
							TokenURL:     "https://test.com",
							EndpointParams: map[string]string{
								"some": "value",
							},
						},
						FollowRedirects: pointer.Bool(true),
					},
				},
				Receivers: []*receiver{
					{
						Name: "mynamespace/global-config/null",
					},
				},
				Route: &route{
					Receiver: "mynamespace/global-config/null",
					Routes: []*route{
						{
							Receiver: "mynamespace/global-config/myreceiver",
							Match: map[string]string{
								"mykey": "myvalue",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid global config with Slack API URL",
			globalConfig: &monitoringingv1.AlertmanagerGlobalConfig{
				SlackAPIURL: &corev1.SecretKeySelector{
					Key: "url",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "slack",
					},
				},
			},
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "global-config",
					Namespace: "mynamespace",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "null",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []v1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringingv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			want: &alertmanagerConfig{
				Global: &globalConfig{
					SlackAPIURL: parseURL(t, "https://slack.example.com"),
				},
				Receivers: []*receiver{
					{
						Name: "mynamespace/global-config/null",
					},
				},
				Route: &route{
					Receiver: "mynamespace/global-config/null",
					Routes: []*route{
						{
							Receiver: "mynamespace/global-config/myreceiver",
							Match: map[string]string{
								"mykey": "myvalue",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "global config with invalid Slack API URL",
			globalConfig: &monitoringingv1.AlertmanagerGlobalConfig{
				SlackAPIURL: &corev1.SecretKeySelector{
					Key: "invalid_url",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "slack",
					},
				},
			},
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "global-config",
					Namespace: "mynamespace",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "null",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []v1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringingv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name: "global config with missing Slack API URL",
			globalConfig: &monitoringingv1.AlertmanagerGlobalConfig{
				SlackAPIURL: &corev1.SecretKeySelector{
					Key: "url",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "not_existing",
					},
				},
			},
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "global-config",
					Namespace: "mynamespace",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "null",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []v1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringingv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name: "valid global config with OpsGenie API URL",
			globalConfig: &monitoringingv1.AlertmanagerGlobalConfig{
				OpsGenieAPIURL: &corev1.SecretKeySelector{
					Key: "url",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "opsgenie",
					},
				},
			},
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "global-config",
					Namespace: "mynamespace",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "null",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []v1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringingv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			want: &alertmanagerConfig{
				Global: &globalConfig{
					OpsGenieAPIURL: parseURL(t, "https://opsgenie.example.com"),
				},
				Receivers: []*receiver{
					{
						Name: "mynamespace/global-config/null",
					},
				},
				Route: &route{
					Receiver: "mynamespace/global-config/null",
					Routes: []*route{
						{
							Receiver: "mynamespace/global-config/myreceiver",
							Match: map[string]string{
								"mykey": "myvalue",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "global config with invalid OpsGenie API URL",
			globalConfig: &monitoringingv1.AlertmanagerGlobalConfig{
				OpsGenieAPIURL: &corev1.SecretKeySelector{
					Key: "invalid_url",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "opsgenie",
					},
				},
			},
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "global-config",
					Namespace: "mynamespace",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "null",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []v1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringingv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name: "global config with missing OpsGenie API URL",
			globalConfig: &monitoringingv1.AlertmanagerGlobalConfig{
				OpsGenieAPIURL: &corev1.SecretKeySelector{
					Key: "url",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "not_existing",
					},
				},
			},
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "global-config",
					Namespace: "mynamespace",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "null",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []v1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringingv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name: "valid global config with OpsGenie API KEY",
			globalConfig: &monitoringingv1.AlertmanagerGlobalConfig{
				OpsGenieAPIKey: &corev1.SecretKeySelector{
					Key: "api_key",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "opsgenie",
					},
				},
			},
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "global-config",
					Namespace: "mynamespace",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "null",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []v1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringingv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			want: &alertmanagerConfig{
				Global: &globalConfig{
					OpsGenieAPIKey: "mykey",
				},
				Receivers: []*receiver{
					{
						Name: "mynamespace/global-config/null",
					},
				},
				Route: &route{
					Receiver: "mynamespace/global-config/null",
					Routes: []*route{
						{
							Receiver: "mynamespace/global-config/myreceiver",
							Match: map[string]string{
								"mykey": "myvalue",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "global config with missing OpsGenie API KEY",
			globalConfig: &monitoringingv1.AlertmanagerGlobalConfig{
				OpsGenieAPIKey: &corev1.SecretKeySelector{
					Key: "api_key",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "not_existing",
					},
				},
			},
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "global-config",
					Namespace: "mynamespace",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "null",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []v1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringingv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name: "valid global config with Pagerduty URL",
			globalConfig: &monitoringingv1.AlertmanagerGlobalConfig{
				PagerdutyURL: &pagerdutyURL,
			},
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "global-config",
					Namespace: "mynamespace",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "null",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []v1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringingv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			want: &alertmanagerConfig{
				Global: &globalConfig{
					PagerdutyURL: parseURL(t, pagerdutyURL),
				},
				Receivers: []*receiver{
					{
						Name: "mynamespace/global-config/null",
					},
				},
				Route: &route{
					Receiver: "mynamespace/global-config/null",
					Routes: []*route{
						{
							Receiver: "mynamespace/global-config/myreceiver",
							Match: map[string]string{
								"mykey": "myvalue",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "global config with invalid Pagerduty URL",
			globalConfig: &monitoringingv1.AlertmanagerGlobalConfig{
				PagerdutyURL: &invalidPagerdutyURL,
			},
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "global-config",
					Namespace: "mynamespace",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "null",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []v1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringingv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name: "missing route",
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "global-config",
					Namespace: "mynamespace",
				},
			},
			wantErr: true,
		},
		{
			name: "globalConfig has null resolve timeout",
			globalConfig: &monitoringingv1.AlertmanagerGlobalConfig{
				HTTPConfig: &monitoringingv1.HTTPConfig{
					FollowRedirects: pointer.Bool(true),
				},
			},
			amConfig: &monitoringv1alpha1.AlertmanagerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "global-config",
					Namespace: "mynamespace",
				},
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{
							Name: "null",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
					},
				},
			},
			matcherStrategy: monitoringingv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			want: &alertmanagerConfig{
				Global: &globalConfig{
					HTTPConfig: &httpClientConfig{
						FollowRedirects: pointer.Bool(true),
					},
				},
				Receivers: []*receiver{
					{
						Name: "mynamespace/global-config/null",
					},
				},
				Route: &route{
					Receiver: "mynamespace/global-config/null",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		version, err := semver.ParseTolerant("v0.22.2")
		if err != nil {
			t.Fatal(err)
		}
		kclient := fake.NewSimpleClientset(
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "webhook-client-id",
					Namespace: "mynamespace",
				},
				Data: map[string]string{
					"test": "clientID",
				},
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "smtp-auth",
					Namespace: "mynamespace",
				},
				Data: map[string][]byte{
					"password": []byte("password"),
					"secret":   []byte("secret"),
				},
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "webhook-client-secret",
					Namespace: "mynamespace",
				},
				Data: map[string][]byte{
					"test": []byte("clientSecret"),
				},
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "slack",
					Namespace: "mynamespace",
				},
				Data: map[string][]byte{
					"url":         []byte("https://slack.example.com"),
					"invalid_url": []byte("://slack.example.com"),
				},
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "opsgenie",
					Namespace: "mynamespace",
				},
				Data: map[string][]byte{
					"url":         []byte("https://opsgenie.example.com"),
					"invalid_url": []byte("://opsgenie.example.com"),
					"api_key":     []byte("mykey"),
				},
			},
		)
		cb := newConfigBuilder(
			log.NewNopLogger(),
			version,
			assets.NewStore(kclient.CoreV1(), kclient.CoreV1()),
			tt.matcherStrategy,
		)
		t.Run(tt.name, func(t *testing.T) {
			err := cb.initializeFromAlertmanagerConfig(context.TODO(), tt.globalConfig, tt.amConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("initializeFromAlertmanagerConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.Equal(t, tt.want, cb.cfg)
		})
	}
}

func TestGenerateConfig(t *testing.T) {
	type testCase struct {
		name            string
		kclient         kubernetes.Interface
		baseConfig      alertmanagerConfig
		amVersion       *semver.Version
		matcherStrategy monitoringingv1.AlertmanagerConfigMatcherStrategy
		amConfigs       map[string]*monitoringv1alpha1.AlertmanagerConfig
		expected        string
	}
	version24, err := semver.ParseTolerant("v0.24.0")
	if err != nil {
		t.Fatal(err)
	}

	globalSlackAPIURL, err := url.Parse("http://slack.example.com")
	if err != nil {
		t.Fatal("Could not parse slack API URL")
	}

	testCases := []testCase{
		{
			name:    "skeleton base, no CRs",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{},
			expected: `route:
  receiver: "null"
receivers:
- name: "null"
templates: []
`,
		},
		{
			name:    "skeleton base with global send_revolved, no CRs",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Global: &globalConfig{
					ResolveTimeout: func(d model.Duration) *model.Duration { return &d }(model.Duration(time.Minute)),
				},
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{},
			expected: `global:
  resolve_timeout: 1m
route:
  receiver: "null"
receivers:
- name: "null"
templates: []
`,
		},
		{
			name:    "skeleton base with global smtp_require_tls set to false, no CRs",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Global: &globalConfig{
					SMTPRequireTLS: func(b bool) *bool { return &b }(false),
				},
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{},
			expected: `global:
  smtp_require_tls: false
route:
  receiver: "null"
receivers:
- name: "null"
templates: []
`,
		},
		{
			name:    "skeleton base with global smtp_require_tls set to true, no CRs",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Global: &globalConfig{
					SMTPRequireTLS: func(b bool) *bool { return &b }(true),
				},
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{},
			expected: `global:
  smtp_require_tls: true
route:
  receiver: "null"
receivers:
- name: "null"
templates: []
`,
		},
		{
			name:    "skeleton base with inhibit rules, no CRs",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				InhibitRules: []*inhibitRule{
					{
						SourceMatchers: []string{"test!=dropped", "expect=~this-value"},
						TargetMatchers: []string{"test!=dropped", "expect=~this-value"},
					},
				},
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{},
			expected: `route:
  receiver: "null"
inhibit_rules:
- target_matchers:
  - test!=dropped
  - expect=~this-value
  source_matchers:
  - test!=dropped
  - expect=~this-value
receivers:
- name: "null"
templates: []
`,
		},
		{
			name:    "base with sub route and matchers, no CRs",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route: &route{
					Receiver: "null",
					Routes: []*route{{
						Matchers: []string{"namespace=custom-test"},
						Receiver: "custom",
					}},
				},
				Receivers: []*receiver{
					{Name: "null"},
					{Name: "custom"},
				},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: custom
    matchers:
    - namespace=custom-test
receivers:
- name: "null"
- name: custom
templates: []
`,
		},
		{
			name:    "skeleton base with mute time intervals, no CRs",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
				MuteTimeIntervals: []*timeInterval{
					{
						Name: "maintenance_windows",
						TimeIntervals: []timeinterval.TimeInterval{
							{
								Months: []timeinterval.MonthRange{
									{
										InclusiveRange: timeinterval.InclusiveRange{
											Begin: 1,
											End:   1,
										},
									},
								},
								DaysOfMonth: []timeinterval.DayOfMonthRange{
									{
										InclusiveRange: timeinterval.InclusiveRange{
											Begin: 7,
											End:   7,
										},
									},
									{
										InclusiveRange: timeinterval.InclusiveRange{
											Begin: 18,
											End:   18,
										},
									},
									{
										InclusiveRange: timeinterval.InclusiveRange{
											Begin: 28,
											End:   28,
										},
									},
								},
								Times: []timeinterval.TimeRange{
									{
										StartMinute: 1020,
										EndMinute:   1440,
									},
								},
							},
						},
					},
				},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{},
			expected: `route:
  receiver: "null"
receivers:
- name: "null"
mute_time_intervals:
- name: maintenance_windows
  time_intervals:
  - times:
    - start_time: "17:00"
      end_time: "24:00"
    days_of_month: ["7", "18", "28"]
    months: ["1"]
templates: []
`,
		},
		{
			name:    "skeleton base with sns receiver, no CRs",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route: &route{Receiver: "sns-test"},
				Receivers: []*receiver{
					{
						Name: "sns-test",
						SNSConfigs: []*snsConfig{
							{
								APIUrl:      "https://sns.us-west-2.amazonaws.com",
								TopicARN:    "arn:test",
								PhoneNumber: "+12345",
								TargetARN:   "arn:target",
								Subject:     "testing",
								Sigv4: sigV4Config{
									Region:    "us-west-2",
									AccessKey: "key",
									SecretKey: "secret",
									Profile:   "dev",
									RoleARN:   "arn:dev",
								},
							},
						},
					},
				},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{},
			expected: `route:
  receiver: sns-test
receivers:
- name: sns-test
  sns_configs:
  - api_url: https://sns.us-west-2.amazonaws.com
    sigv4:
      region: us-west-2
      access_key: key
      secret_key: secret
      profile: dev
      role_arn: arn:dev
    topic_arn: arn:test
    phone_number: "+12345"
    target_arn: arn:target
    subject: testing
templates: []
`,
		},
		{
			name:      "skeleton base with active_time_intervals, no CRs",
			amVersion: &version24,
			kclient:   fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route: &route{
					Receiver: "null",
					Routes: []*route{
						{
							Receiver:            "null",
							ActiveTimeIntervals: []string{"workdays"},
						},
					},
				},
				Receivers: []*receiver{{Name: "null"}},
				TimeIntervals: []*timeInterval{
					{
						Name: "workdays",
						TimeIntervals: []timeinterval.TimeInterval{
							{
								Weekdays: []timeinterval.WeekdayRange{
									{
										InclusiveRange: timeinterval.InclusiveRange{Begin: 1, End: 5},
									},
								},
							},
						},
					},
				},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: "null"
    active_time_intervals:
    - workdays
receivers:
- name: "null"
time_intervals:
- name: workdays
  time_intervals:
  - weekdays: ['monday:friday']
templates: []
`,
		},
		{
			name:    "skeleton base, simple CR",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test",
							GroupBy:  []string{"job"},
						},
						Receivers: []monitoringv1alpha1.Receiver{{Name: "test"}},
					},
				},
			},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test
    group_by:
    - job
    matchers:
    - namespace="mynamespace"
    continue: true
receivers:
- name: "null"
- name: mynamespace/myamc/test
templates: []
`,
		},
		{
			name:    "skeleton base, simple CR with namespaceMatcher disabled",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			matcherStrategy: monitoringingv1.AlertmanagerConfigMatcherStrategy{
				Type: "None",
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test",
							GroupBy:  []string{"job"},
						},
						Receivers: []monitoringv1alpha1.Receiver{{Name: "test"}},
					},
				},
			},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test
    group_by:
    - job
    continue: true
receivers:
- name: "null"
- name: mynamespace/myamc/test
templates: []
`,
		},
		{
			name:    "skeleton base, CR with inhibition rules only (deprecated matchers not converted)",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			amVersion: &semver.Version{
				Major: 0,
				Minor: 20,
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
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
			},
			expected: `route:
  receiver: "null"
inhibit_rules:
- target_match:
    alertname: TargetDown
    namespace: mynamespace
  source_match:
    alertname: NodeNotReady
    namespace: mynamespace
  equal:
  - node
receivers:
- name: "null"
templates: []
`,
		},
		{
			name:    "skeleton base, CR with inhibition rules only (deprecated matchers are converted)",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						InhibitRules: []monitoringv1alpha1.InhibitRule{
							{
								SourceMatch: []monitoringv1alpha1.Matcher{
									{
										Name:  "alertname",
										Value: "NodeNotReady",
										Regex: true,
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
			},
			expected: `route:
  receiver: "null"
inhibit_rules:
- target_matchers:
  - alertname="TargetDown"
  - namespace="mynamespace"
  source_matchers:
  - alertname=~"NodeNotReady"
  - namespace="mynamespace"
  equal:
  - node
receivers:
- name: "null"
templates: []
`,
		},
		{
			name:    "skeleton base, CR with inhibition rules only",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						InhibitRules: []monitoringv1alpha1.InhibitRule{
							{
								SourceMatch: []monitoringv1alpha1.Matcher{
									{
										Name:      "alertname",
										MatchType: monitoringv1alpha1.MatchRegexp,
										Value:     "NodeNotReady",
									},
								},
								TargetMatch: []monitoringv1alpha1.Matcher{
									{
										Name:      "alertname",
										MatchType: monitoringv1alpha1.MatchNotEqual,
										Value:     "TargetDown",
									},
								},
								Equal: []string{"node"},
							},
						},
					},
				},
			},
			expected: `route:
  receiver: "null"
inhibit_rules:
- target_matchers:
  - alertname!="TargetDown"
  - namespace="mynamespace"
  source_matchers:
  - alertname=~"NodeNotReady"
  - namespace="mynamespace"
  equal:
  - node
receivers:
- name: "null"
templates: []
`,
		},
		{
			name:    "base with subroute - deprecated matching pattern, simple CR",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route: &route{
					Receiver: "null",
					Routes:   []*route{{Receiver: "null"}},
				},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test",
						},
						Receivers: []monitoringv1alpha1.Receiver{{Name: "test"}},
					},
				},
			},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test
    matchers:
    - namespace="mynamespace"
    continue: true
  - receiver: "null"
receivers:
- name: "null"
- name: mynamespace/myamc/test
templates: []
`,
		},
		{
			name: "CR with Pagerduty Receiver",
			kclient: fake.NewSimpleClientset(
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "am-pd-test-receiver",
						Namespace: "mynamespace",
					},
					Data: map[string][]byte{
						"routingKey": []byte("1234abc"),
					},
				},
			),
			baseConfig: alertmanagerConfig{
				Route: &route{
					Receiver: "null",
				},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test-pd",
						},
						Receivers: []monitoringv1alpha1.Receiver{{
							Name: "test-pd",
							PagerDutyConfigs: []monitoringv1alpha1.PagerDutyConfig{{
								RoutingKey: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "am-pd-test-receiver",
									},
									Key: "routingKey",
								},
								PagerDutyImageConfigs: []monitoringv1alpha1.PagerDutyImageConfig{
									{
										Src:  "https://some-image.com",
										Href: "https://some-image.com",
										Alt:  "some-image",
									},
								},
								PagerDutyLinkConfigs: []monitoringv1alpha1.PagerDutyLinkConfig{
									{
										Href: "https://some-link.com",
										Text: "some-link",
									},
								},
							}},
						}},
					},
				},
			},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test-pd
    matchers:
    - namespace="mynamespace"
    continue: true
receivers:
- name: "null"
- name: mynamespace/myamc/test-pd
  pagerduty_configs:
  - routing_key: 1234abc
    images:
    - src: https://some-image.com
      alt: some-image
      href: https://some-image.com
    links:
    - href: https://some-link.com
      text: some-link
templates: []
`,
		},
		{
			name: "CR with Webhook Receiver and custom http config (oauth2)",
			kclient: fake.NewSimpleClientset(
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "webhook-client-id",
						Namespace: "mynamespace",
					},
					Data: map[string]string{
						"test": "clientID",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "webhook-client-secret",
						Namespace: "mynamespace",
					},
					Data: map[string][]byte{
						"test": []byte("clientSecret"),
					},
				},
			),
			baseConfig: alertmanagerConfig{
				Route: &route{
					Receiver: "null",
				},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test",
						},
						Receivers: []monitoringv1alpha1.Receiver{{
							Name: "test",
							WebhookConfigs: []monitoringv1alpha1.WebhookConfig{{
								URL: func(s string) *string {
									return &s
								}("http://test.url"),
								HTTPConfig: &monitoringv1alpha1.HTTPConfig{
									OAuth2: &monitoringingv1.OAuth2{
										ClientID: monitoringingv1.SecretOrConfigMap{
											ConfigMap: &corev1.ConfigMapKeySelector{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: "webhook-client-id",
												},
												Key: "test",
											},
										},
										ClientSecret: corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "webhook-client-secret",
											},
											Key: "test",
										},
										TokenURL: "https://test.com",
										Scopes:   []string{"any"},
										EndpointParams: map[string]string{
											"some": "value",
										},
									},
									FollowRedirects: pointer.Bool(true),
								},
							}},
						}},
					},
				},
			},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test
    matchers:
    - namespace="mynamespace"
    continue: true
receivers:
- name: "null"
- name: mynamespace/myamc/test
  webhook_configs:
  - url: http://test.url
    http_config:
      oauth2:
        client_id: clientID
        client_secret: clientSecret
        scopes:
        - any
        token_url: https://test.com
        endpoint_params:
          some: value
      follow_redirects: true
templates: []
`,
		},
		{
			name: "CR with Opsgenie Receiver",
			kclient: fake.NewSimpleClientset(
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "am-og-test-receiver",
						Namespace: "mynamespace",
					},
					Data: map[string][]byte{
						"apiKey": []byte("1234abc"),
					},
				},
			),
			baseConfig: alertmanagerConfig{
				Route: &route{
					Receiver: "null",
				},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test",
						},
						Receivers: []monitoringv1alpha1.Receiver{{
							Name: "test",
							OpsGenieConfigs: []monitoringv1alpha1.OpsGenieConfig{{
								APIKey: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "am-og-test-receiver",
									},
									Key: "apiKey",
								},
							}},
						}},
					},
				},
			},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test
    matchers:
    - namespace="mynamespace"
    continue: true
receivers:
- name: "null"
- name: mynamespace/myamc/test
  opsgenie_configs:
  - api_key: 1234abc
templates: []
`,
		},
		{
			name: "CR with Opsgenie Team Responder",
			kclient: fake.NewSimpleClientset(
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "am-og-test-receiver",
						Namespace: "mynamespace",
					},
					Data: map[string][]byte{
						"apiKey": []byte("1234abc"),
					},
				},
			),
			baseConfig: alertmanagerConfig{
				Route: &route{
					Receiver: "null",
				},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test",
						},
						Receivers: []monitoringv1alpha1.Receiver{{
							Name: "test",
							OpsGenieConfigs: []monitoringv1alpha1.OpsGenieConfig{{
								APIKey: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "am-og-test-receiver",
									},
									Key: "apiKey",
								},
								Responders: []monitoringv1alpha1.OpsGenieConfigResponder{{
									Name: "myname",
									Type: "team",
								}},
							}},
						}},
					},
				},
			},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test
    matchers:
    - namespace="mynamespace"
    continue: true
receivers:
- name: "null"
- name: mynamespace/myamc/test
  opsgenie_configs:
  - api_key: 1234abc
    responders:
    - name: myname
      type: team
templates: []
`,
		},
		{
			name: "CR with WeChat Receiver",
			kclient: fake.NewSimpleClientset(
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "am-wechat-test-receiver",
						Namespace: "mynamespace",
					},
					Data: map[string][]byte{
						"apiSecret": []byte("wechatsecret"),
					},
				},
			),
			baseConfig: alertmanagerConfig{
				Route: &route{
					Receiver: "null",
				},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test",
						},
						Receivers: []monitoringv1alpha1.Receiver{{
							Name: "test",
							WeChatConfigs: []monitoringv1alpha1.WeChatConfig{{
								APISecret: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "am-wechat-test-receiver",
									},
									Key: "apiSecret",
								},
								CorpID: "wechatcorpid",
							}},
						}},
					},
				},
			},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test
    matchers:
    - namespace="mynamespace"
    continue: true
receivers:
- name: "null"
- name: mynamespace/myamc/test
  wechat_configs:
  - api_secret: wechatsecret
    corp_id: wechatcorpid
templates: []
`,
		},

		{
			name:      "CR with Telegram Receiver",
			amVersion: &version24,
			kclient: fake.NewSimpleClientset(
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "am-telegram-test-receiver",
						Namespace: "mynamespace",
					},
					Data: map[string][]byte{
						"botToken": []byte("bipbop"),
					},
				},
			),
			baseConfig: alertmanagerConfig{
				Route: &route{
					Receiver: "null",
				},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test",
						},
						Receivers: []monitoringv1alpha1.Receiver{{
							Name: "test",
							TelegramConfigs: []monitoringv1alpha1.TelegramConfig{{
								APIURL: "https://api.telegram.org",
								BotToken: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "am-telegram-test-receiver",
									},
									Key: "botToken",
								},
								ChatID: 12345,
							}},
						}},
					},
				},
			},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test
    matchers:
    - namespace="mynamespace"
    continue: true
receivers:
- name: "null"
- name: mynamespace/myamc/test
  telegram_configs:
  - api_url: https://api.telegram.org
    bot_token: bipbop
    chat_id: 12345
templates: []
`,
		},

		{

			name:    "CR with Slack Receiver and global Slack URL",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Global: &globalConfig{
					SlackAPIURL: &config.URL{URL: globalSlackAPIURL},
				},
				Route: &route{
					Receiver: "null",
				},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test",
						},
						Receivers: []monitoringv1alpha1.Receiver{{
							Name: "test",
							SlackConfigs: []monitoringv1alpha1.SlackConfig{{
								Actions: []monitoringv1alpha1.SlackAction{
									{
										Type: "type",
										Text: "text",
										Name: "my-action",
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
							}},
						}},
					},
				},
			},
			expected: `global:
  slack_api_url: http://slack.example.com
route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test
    matchers:
    - namespace="mynamespace"
    continue: true
receivers:
- name: "null"
- name: mynamespace/myamc/test
  slack_configs:
  - fields:
    - title: title
      value: value
    actions:
    - type: type
      text: text
      name: my-action
      confirm:
        text: text
templates: []
`,
		},
		{

			name:    "CR with Slack Receiver and global Slack URL File",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Global: &globalConfig{
					SlackAPIURLFile: "/etc/test",
				},
				Route: &route{
					Receiver: "null",
				},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test",
						},
						Receivers: []monitoringv1alpha1.Receiver{{
							Name: "test",
							SlackConfigs: []monitoringv1alpha1.SlackConfig{{
								Actions: []monitoringv1alpha1.SlackAction{
									{
										Type: "type",
										Text: "text",
										Name: "my-action",
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
							}},
						}},
					},
				},
			},
			expected: `global:
  slack_api_url_file: /etc/test
route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test
    matchers:
    - namespace="mynamespace"
    continue: true
receivers:
- name: "null"
- name: mynamespace/myamc/test
  slack_configs:
  - fields:
    - title: title
      value: value
    actions:
    - type: type
      text: text
      name: my-action
      confirm:
        text: text
templates: []
`,
		},
		{

			name: "CR with SNS Receiver with Access and Key",
			kclient: fake.NewSimpleClientset(
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "am-sns-test",
						Namespace: "mynamespace",
					},
					Data: map[string][]byte{
						"key":    []byte("xyz"),
						"secret": []byte("123"),
					},
				}),
			baseConfig: alertmanagerConfig{
				Route: &route{
					Receiver: "null",
				},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test",
						},
						Receivers: []monitoringv1alpha1.Receiver{{
							Name: "test",
							SNSConfigs: []monitoringv1alpha1.SNSConfig{
								{
									ApiURL: "https://sns.us-east-2.amazonaws.com",
									Sigv4: &monitoringingv1.Sigv4{
										Region: "us-east-2",
										AccessKey: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "am-sns-test",
											},
											Key: "key",
										},
										SecretKey: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "am-sns-test",
											},
											Key: "secret",
										},
									},
									TopicARN: "test-topicARN",
								},
							},
						}},
					},
				},
			},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test
    matchers:
    - namespace="mynamespace"
    continue: true
receivers:
- name: "null"
- name: mynamespace/myamc/test
  sns_configs:
  - api_url: https://sns.us-east-2.amazonaws.com
    sigv4:
      region: us-east-2
      access_key: xyz
      secret_key: "123"
    topic_arn: test-topicARN
templates: []
`,
		},
		{

			name: "CR with SNS Receiver with roleARN",
			kclient: fake.NewSimpleClientset(
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "am-sns-test",
						Namespace: "mynamespace",
					},
					Data: map[string][]byte{
						"key":    []byte("xyz"),
						"secret": []byte("123"),
					},
				}),
			baseConfig: alertmanagerConfig{
				Route: &route{
					Receiver: "null",
				},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test",
						},
						Receivers: []monitoringv1alpha1.Receiver{{
							Name: "test",
							SNSConfigs: []monitoringv1alpha1.SNSConfig{
								{
									ApiURL: "https://sns.us-east-2.amazonaws.com",
									Sigv4: &monitoringingv1.Sigv4{
										Region:  "us-east-2",
										RoleArn: "test-roleARN",
									},
									TopicARN: "test-topicARN",
								},
							},
						}},
					},
				},
			},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test
    matchers:
    - namespace="mynamespace"
    continue: true
receivers:
- name: "null"
- name: mynamespace/myamc/test
  sns_configs:
  - api_url: https://sns.us-east-2.amazonaws.com
    sigv4:
      region: us-east-2
      role_arn: test-roleARN
    topic_arn: test-topicARN
templates: []
`,
		},
		{

			name:    "CR with Mute Time Intervals",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Global: &globalConfig{
					SlackAPIURLFile: "/etc/test",
				},
				Route: &route{
					Receiver: "null",
				},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver:          "test",
							MuteTimeIntervals: []string{"test"},
						},
						MuteTimeIntervals: []monitoringv1alpha1.MuteTimeInterval{
							{
								Name: "test",
								TimeIntervals: []monitoringv1alpha1.TimeInterval{
									{
										Times: []monitoringv1alpha1.TimeRange{
											{
												StartTime: "08:00",
												EndTime:   "17:00",
											},
										},
										Weekdays: []monitoringv1alpha1.WeekdayRange{
											monitoringv1alpha1.WeekdayRange("Saturday"),
											monitoringv1alpha1.WeekdayRange("Sunday"),
										},
										Months: []monitoringv1alpha1.MonthRange{
											"January:March",
										},
										DaysOfMonth: []monitoringv1alpha1.DayOfMonthRange{
											{
												Start: 1,
												End:   10,
											},
										},
										Years: []monitoringv1alpha1.YearRange{
											"2030:2050",
										},
									},
								},
							},
						},
						Receivers: []monitoringv1alpha1.Receiver{{
							Name: "test",
							SlackConfigs: []monitoringv1alpha1.SlackConfig{{
								Actions: []monitoringv1alpha1.SlackAction{
									{
										Type: "type",
										Text: "text",
										Name: "my-action",
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
							}},
						}},
					},
				},
			},
			expected: `global:
  slack_api_url_file: /etc/test
route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test
    matchers:
    - namespace="mynamespace"
    continue: true
    mute_time_intervals:
    - mynamespace/myamc/test
receivers:
- name: "null"
- name: mynamespace/myamc/test
  slack_configs:
  - fields:
    - title: title
      value: value
    actions:
    - type: type
      text: text
      name: my-action
      confirm:
        text: text
mute_time_intervals:
- name: mynamespace/myamc/test
  time_intervals:
  - times:
    - start_time: "08:00"
      end_time: "17:00"
    weekdays: [saturday, sunday]
    days_of_month: ["1:10"]
    months: ["1:3"]
    years: ['2030:2050']
templates: []
`,
		},
		{
			name:    "CR with Active Time Intervals",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Global: &globalConfig{
					SlackAPIURLFile: "/etc/test",
				},
				Route: &route{
					Receiver: "null",
				},
				Receivers: []*receiver{{Name: "null"}},
			},
			amVersion: &version24,
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"mynamespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "mynamespace",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver:            "test",
							ActiveTimeIntervals: []string{"test"},
						},
						MuteTimeIntervals: []monitoringv1alpha1.MuteTimeInterval{
							{
								Name: "test",
								TimeIntervals: []monitoringv1alpha1.TimeInterval{
									{
										Times: []monitoringv1alpha1.TimeRange{
											{
												StartTime: "08:00",
												EndTime:   "17:00",
											},
										},
										Weekdays: []monitoringv1alpha1.WeekdayRange{
											monitoringv1alpha1.WeekdayRange("Saturday"),
											monitoringv1alpha1.WeekdayRange("Sunday"),
										},
										Months: []monitoringv1alpha1.MonthRange{
											"January:March",
										},
										DaysOfMonth: []monitoringv1alpha1.DayOfMonthRange{
											{
												Start: 1,
												End:   10,
											},
										},
										Years: []monitoringv1alpha1.YearRange{
											"2030:2050",
										},
									},
								},
							},
						},
						Receivers: []monitoringv1alpha1.Receiver{{
							Name: "test",
							SlackConfigs: []monitoringv1alpha1.SlackConfig{{
								Actions: []monitoringv1alpha1.SlackAction{
									{
										Type: "type",
										Text: "text",
										Name: "my-action",
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
							}},
						}},
					},
				},
			},
			expected: `global:
  slack_api_url_file: /etc/test
route:
  receiver: "null"
  routes:
  - receiver: mynamespace/myamc/test
    matchers:
    - namespace="mynamespace"
    continue: true
    active_time_intervals:
    - mynamespace/myamc/test
receivers:
- name: "null"
- name: mynamespace/myamc/test
  slack_configs:
  - fields:
    - title: title
      value: value
    actions:
    - type: type
      text: text
      name: my-action
      confirm:
        text: text
mute_time_intervals:
- name: mynamespace/myamc/test
  time_intervals:
  - times:
    - start_time: "08:00"
      end_time: "17:00"
    weekdays: [saturday, sunday]
    days_of_month: ["1:10"]
    months: ["1:3"]
    years: ['2030:2050']
templates: []
`,
		},
	}

	logger := log.NewNopLogger()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewStore(tc.kclient.CoreV1(), tc.kclient.CoreV1())

			if tc.amVersion == nil {
				version, err := semver.ParseTolerant("v0.22.2")
				if err != nil {
					t.Fatal(err)
				}
				tc.amVersion = &version
			}

			cb := newConfigBuilder(logger, *tc.amVersion, store, tc.matcherStrategy)
			cb.cfg = &tc.baseConfig

			if err := cb.addAlertmanagerConfigs(context.Background(), tc.amConfigs); err != nil {
				t.Fatal(err)
			}

			cfgBytes, err := cb.marshalJSON()
			if err != nil {
				t.Fatal(err)
			}

			// Verify the generated yaml is as expected
			if diff := cmp.Diff(tc.expected, string(cfgBytes)); diff != "" {
				t.Errorf("Unexpected result (-want +got):\n%s", diff)
			}

			// Verify the generated config is something that Alertmanager will be happy with
			_, err = alertmanagerConfigFromBytes(cfgBytes)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSanitizeConfig(t *testing.T) {
	logger := log.NewNopLogger()
	versionFileURLAllowed := semver.Version{Major: 0, Minor: 22}
	versionFileURLNotAllowed := semver.Version{Major: 0, Minor: 21}

	matcherV2SyntaxAllowed := semver.Version{Major: 0, Minor: 22}
	matcherV2SyntaxNotAllowed := semver.Version{Major: 0, Minor: 21}

	versionOpsGenieAPIKeyFileAllowed := semver.Version{Major: 0, Minor: 24}
	versionOpsGenieAPIKeyFileNotAllowed := semver.Version{Major: 0, Minor: 23}

	versionDiscordAllowed := semver.Version{Major: 0, Minor: 25}
	versionDiscordNotAllowed := semver.Version{Major: 0, Minor: 24}

	versionWebexAllowed := semver.Version{Major: 0, Minor: 25}
	versionWebexNotAllowed := semver.Version{Major: 0, Minor: 24}

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		expect         alertmanagerConfig
		expectErr      bool
	}{
		{
			name:           "Test slack_api_url takes precedence in global config",
			againstVersion: versionFileURLAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SlackAPIURL: &config.URL{
						URL: &url.URL{
							Host: "www.test.com",
						}},
					SlackAPIURLFile: "/test",
				},
			},
			expect: alertmanagerConfig{
				Global: &globalConfig{
					SlackAPIURL: &config.URL{
						URL: &url.URL{
							Host: "www.test.com",
						}},
					SlackAPIURLFile: "",
				},
			},
		},
		{
			name:           "Test slack_api_url_file is dropped for unsupported versions",
			againstVersion: versionFileURLNotAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SlackAPIURLFile: "/test",
				},
			},
			expect: alertmanagerConfig{
				Global: &globalConfig{
					SlackAPIURLFile: "",
				},
			},
		},
		{
			name:           "Test api_url takes precedence in slack config",
			againstVersion: versionFileURLAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						SlackConfigs: []*slackConfig{
							{
								APIURL:     "www.test.com",
								APIURLFile: "/test",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						SlackConfigs: []*slackConfig{
							{
								APIURL:     "www.test.com",
								APIURLFile: "",
							},
						},
					},
				},
			},
		},
		{
			name:           "Test api_url_file is dropped in slack config for unsupported versions",
			againstVersion: versionFileURLNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						SlackConfigs: []*slackConfig{
							{
								APIURLFile: "/test",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						SlackConfigs: []*slackConfig{
							{
								APIURLFile: "",
							},
						},
					},
				},
			},
		},
		{
			name:           "Test slack config happy path",
			againstVersion: versionFileURLAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SlackAPIURLFile: "/test",
				},
				Receivers: []*receiver{
					{
						SlackConfigs: []*slackConfig{
							{
								APIURLFile: "/test/case",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Global: &globalConfig{
					SlackAPIURLFile: "/test",
				},
				Receivers: []*receiver{
					{
						SlackConfigs: []*slackConfig{
							{
								APIURLFile: "/test/case",
							},
						},
					},
				},
			},
		},
		{
			name:           "Test inhibit rules error with unsupported syntax",
			againstVersion: matcherV2SyntaxNotAllowed,
			in: &alertmanagerConfig{
				InhibitRules: []*inhibitRule{
					{
						// this rule is marked as invalid. we must error out despite a valid config @[1]
						TargetMatch: map[string]string{
							"dropped": "as-side-effect",
						},
						TargetMatchers: []string{"drop=~me"},
						SourceMatch: map[string]string{
							"dropped": "as-side-effect",
						},
						SourceMatchers: []string{"drop=~me"},
					},
					{
						// test we continue to support both syntax
						TargetMatch: map[string]string{
							"keep": "me-for-now",
						},
						SourceMatch: map[string]string{
							"keep": "me-for-now",
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "Test inhibit rules happy path",
			againstVersion: matcherV2SyntaxAllowed,
			in: &alertmanagerConfig{
				InhibitRules: []*inhibitRule{
					{
						// test we continue to support both syntax
						TargetMatch: map[string]string{
							"keep": "me-for-now",
						},
						TargetMatchers: []string{"keep=~me"},
						SourceMatch: map[string]string{
							"keep": "me-for-now",
						},
						SourceMatchers: []string{"keep=me"},
					},
				},
			},
			expect: alertmanagerConfig{
				InhibitRules: []*inhibitRule{
					{
						TargetMatch: map[string]string{
							"keep": "me-for-now",
						},
						TargetMatchers: []string{"keep=~me"},
						SourceMatch: map[string]string{
							"keep": "me-for-now",
						},
						SourceMatchers: []string{"keep=me"},
					},
				},
			},
		},
		{
			name:           "opsgenie_api_key_file config",
			againstVersion: versionOpsGenieAPIKeyFileAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					OpsGenieAPIKeyFile: "/test",
				},
			},
			expect: alertmanagerConfig{
				Global: &globalConfig{
					OpsGenieAPIKeyFile: "/test",
				},
			},
		},
		{
			name:           "api_key_file field for OpsGenie config",
			againstVersion: versionOpsGenieAPIKeyFileAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "opsgenie",
						OpsgenieConfigs: []*opsgenieConfig{
							{
								APIKeyFile: "/test",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "opsgenie",
						OpsgenieConfigs: []*opsgenieConfig{
							{
								APIKeyFile: "/test",
							},
						},
					},
				},
			},
		},
		{
			name:           "api_key_file and api_key fields for OpsGenie config",
			againstVersion: versionOpsGenieAPIKeyFileAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "opsgenie",
						OpsgenieConfigs: []*opsgenieConfig{
							{
								APIKey:     "test",
								APIKeyFile: "/test",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "opsgenie",
						OpsgenieConfigs: []*opsgenieConfig{
							{
								APIKey: "test",
							},
						},
					},
				},
			},
		},
		{
			name:           "opsgenie_api_key_file is dropped for unsupported versions",
			againstVersion: versionOpsGenieAPIKeyFileNotAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					OpsGenieAPIKeyFile: "/test",
				},
			},
			expect: alertmanagerConfig{
				Global: &globalConfig{},
			},
		},
		{
			name:           "api_key_file is dropped for unsupported versions",
			againstVersion: versionOpsGenieAPIKeyFileNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "opsgenie",
						OpsgenieConfigs: []*opsgenieConfig{
							{
								APIKeyFile: "/test",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name:            "opsgenie",
						OpsgenieConfigs: []*opsgenieConfig{{}},
					},
				},
			},
		},
		{
			name:           "discord_config for supported versions",
			againstVersion: versionDiscordAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						DiscordConfigs: []*discordConfig{
							{
								WebhookURL: "http://example.com",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						DiscordConfigs: []*discordConfig{
							{
								WebhookURL: "http://example.com",
							},
						},
					},
				},
			},
		},
		{
			name:           "discord_config returns error for unsupported versions",
			againstVersion: versionDiscordNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						DiscordConfigs: []*discordConfig{
							{
								WebhookURL: "http://example.com",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "webex_config for supported versions",
			againstVersion: versionWebexAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						WebexConfigs: []*webexConfig{
							{
								APIURL: "http://example.com",
								RoomID: "foo",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						WebexConfigs: []*webexConfig{
							{
								APIURL: "http://example.com",
								RoomID: "foo",
							},
						},
					},
				},
			},
		},
		{
			name:           "webex_config returns error for unsupported versions",
			againstVersion: versionWebexNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						WebexConfigs: []*webexConfig{
							{
								APIURL: "http://example.com",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "webex_config returns error for missing mandatory field",
			againstVersion: versionWebexAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						WebexConfigs: []*webexConfig{
							{
								APIURL: "http://example.com",
							},
						},
					},
				},
			},
			expectErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got: %q", err)
			}

			require.Equal(t, tc.expect, *tc.in)
		})
	}
}

func TestHTTPClientConfig(t *testing.T) {
	logger := log.NewNopLogger()

	httpConfigV25Allowed := semver.Version{Major: 0, Minor: 25}
	httpConfigV25NotAllowed := semver.Version{Major: 0, Minor: 24}

	versionAuthzAllowed := semver.Version{Major: 0, Minor: 22}
	versionAuthzNotAllowed := semver.Version{Major: 0, Minor: 21}

	// test the http config independently since all receivers rely on same behaviour
	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *httpClientConfig
		expect         httpClientConfig
		expectErr      bool
	}{
		{
			name: "Test happy path",
			in: &httpClientConfig{
				Authorization: &authorization{
					Type:            "any",
					Credentials:     "some",
					CredentialsFile: "/must/keep",
				},
			},
			againstVersion: versionAuthzAllowed,
			expect: httpClientConfig{
				Authorization: &authorization{
					Type:            "any",
					Credentials:     "some",
					CredentialsFile: "/must/keep",
				},
			},
		},
		{
			name: "HTTP client config fields preserved with v0.25.0",
			in: &httpClientConfig{
				OAuth2: &oauth2{
					ClientID:         "a",
					ClientSecret:     "b",
					ClientSecretFile: "c",
					TokenURL:         "d",
					ProxyURL:         "http://example.com/",
				},
				EnableHTTP2: pointer.Bool(false),
				TLSConfig: &tlsConfig{
					MinVersion: "TLS12",
					MaxVersion: "TLS12",
				},
			},
			againstVersion: httpConfigV25Allowed,
			expect: httpClientConfig{
				OAuth2: &oauth2{
					ClientID:         "a",
					ClientSecret:     "b",
					ClientSecretFile: "c",
					TokenURL:         "d",
					ProxyURL:         "http://example.com/",
				},
				EnableHTTP2: pointer.Bool(false),
				TLSConfig: &tlsConfig{
					MinVersion: "TLS12",
					MaxVersion: "TLS12",
				},
			},
		},
		{
			name:           "Test authorization causes error for unsupported versions",
			againstVersion: versionAuthzNotAllowed,
			in: &httpClientConfig{
				Authorization: &authorization{
					Type:            "any",
					Credentials:     "some",
					CredentialsFile: "/must/drop",
				},
			},
			expectErr: true,
		},
		{
			name:           "Test oauth2 causes error for unsupported versions",
			againstVersion: versionAuthzNotAllowed,
			in: &httpClientConfig{
				OAuth2: &oauth2{
					ClientID:         "a",
					ClientSecret:     "b",
					ClientSecretFile: "c",
					TokenURL:         "d",
				},
			},
			expectErr: true,
		},
		{
			name: "HTTP client config with min TLS version only",
			in: &httpClientConfig{
				TLSConfig: &tlsConfig{
					MinVersion: "TLS12",
				},
			},
			againstVersion: httpConfigV25Allowed,
			expect: httpClientConfig{
				TLSConfig: &tlsConfig{
					MinVersion: "TLS12",
				},
			},
		},
		{
			name: "HTTP client config with max TLS version only",
			in: &httpClientConfig{
				TLSConfig: &tlsConfig{
					MaxVersion: "TLS12",
				},
			},
			againstVersion: httpConfigV25Allowed,
			expect: httpClientConfig{
				TLSConfig: &tlsConfig{
					MaxVersion: "TLS12",
				},
			},
		},
		{
			name: "HTTP client config TLS min version > max version",
			in: &httpClientConfig{
				OAuth2: &oauth2{
					ClientID:         "a",
					ClientSecret:     "b",
					ClientSecretFile: "c",
					TokenURL:         "d",
					ProxyURL:         "http://example.com/",
				},
				EnableHTTP2: pointer.Bool(false),
				TLSConfig: &tlsConfig{
					MinVersion: "TLS13",
					MaxVersion: "TLS12",
				},
			},
			againstVersion: httpConfigV25Allowed,
			expectErr:      true,
		},
		{
			name: "HTTP client config TLS min version unknown",
			in: &httpClientConfig{
				OAuth2: &oauth2{
					ClientID:         "a",
					ClientSecret:     "b",
					ClientSecretFile: "c",
					TokenURL:         "d",
					ProxyURL:         "http://example.com/",
				},
				EnableHTTP2: pointer.Bool(false),
				TLSConfig: &tlsConfig{
					MinVersion: "TLS14",
				},
			},
			againstVersion: httpConfigV25Allowed,
			expectErr:      true,
		},
		{
			name: "HTTP client config TLS max version unknown",
			in: &httpClientConfig{
				OAuth2: &oauth2{
					ClientID:         "a",
					ClientSecret:     "b",
					ClientSecretFile: "c",
					TokenURL:         "d",
					ProxyURL:         "http://example.com/",
				},
				EnableHTTP2: pointer.Bool(false),
				TLSConfig: &tlsConfig{
					MaxVersion: "TLS14",
				},
			},
			againstVersion: httpConfigV25Allowed,
			expectErr:      true,
		},
		{
			name: "Test HTTP client config fields dropped before v0.25.0",
			in: &httpClientConfig{
				OAuth2: &oauth2{
					ClientID:         "a",
					ClientSecret:     "b",
					ClientSecretFile: "c",
					TokenURL:         "d",
					ProxyURL:         "http://example.com/",
				},
				EnableHTTP2: pointer.Bool(false),
				TLSConfig: &tlsConfig{
					MinVersion: "TLS12",
					MaxVersion: "TLS12",
				},
			},
			againstVersion: httpConfigV25NotAllowed,
			expect: httpClientConfig{
				OAuth2: &oauth2{
					ClientID:         "a",
					ClientSecret:     "b",
					ClientSecretFile: "c",
					TokenURL:         "d",
				},
				TLSConfig: &tlsConfig{},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expect, *tc.in)
		})
	}
}

func TestTimeInterval(t *testing.T) {
	logger := log.NewNopLogger()

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		expect         alertmanagerConfig
		expectErr      bool
	}{
		{
			name:           "time_intervals and active_time_intervals in Route config",
			againstVersion: semver.Version{Major: 0, Minor: 24},
			in: &alertmanagerConfig{
				TimeIntervals: []*timeInterval{
					{
						Name:          "weekend",
						TimeIntervals: []timeinterval.TimeInterval{},
					},
				},
				Route: &route{
					ActiveTimeIntervals: []string{
						"weekend",
					},
				},
			},
			expect: alertmanagerConfig{
				TimeIntervals: []*timeInterval{
					{
						Name:          "weekend",
						TimeIntervals: []timeinterval.TimeInterval{},
					},
				},
				Route: &route{
					ActiveTimeIntervals: []string{
						"weekend",
					},
				},
			},
		},
		{
			name:           "time_intervals is dropped for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 23},
			in: &alertmanagerConfig{
				TimeIntervals: []*timeInterval{
					{
						Name:          "weekend",
						TimeIntervals: []timeinterval.TimeInterval{},
					},
				},
			},
			expect: alertmanagerConfig{},
		},
		{
			name:           "active_time_intervals is dropped for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 23},
			in: &alertmanagerConfig{
				TimeIntervals: []*timeInterval{
					{
						Name:          "weekend",
						TimeIntervals: []timeinterval.TimeInterval{},
					},
				},
				Route: &route{
					ActiveTimeIntervals: []string{
						"weekend",
					},
				},
			},
			expect: alertmanagerConfig{
				Route: &route{},
			},
		},
		{
			name:           "location is dropped for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 24},
			in: &alertmanagerConfig{
				MuteTimeIntervals: []*timeInterval{
					{
						Name: "workdays",
						TimeIntervals: []timeinterval.TimeInterval{
							{
								Weekdays: []timeinterval.WeekdayRange{
									{
										InclusiveRange: timeinterval.InclusiveRange{Begin: 1, End: 5},
									},
								},
								Location: &timeinterval.Location{
									Location: time.Local,
								},
							},
						},
					},
				},
				TimeIntervals: []*timeInterval{
					{
						Name: "sunday",
						TimeIntervals: []timeinterval.TimeInterval{
							{
								Weekdays: []timeinterval.WeekdayRange{
									{
										InclusiveRange: timeinterval.InclusiveRange{Begin: 0, End: 0},
									},
								},
								Location: &timeinterval.Location{
									Location: time.Local,
								},
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				MuteTimeIntervals: []*timeInterval{
					{
						Name: "workdays",
						TimeIntervals: []timeinterval.TimeInterval{
							{
								Weekdays: []timeinterval.WeekdayRange{
									{
										InclusiveRange: timeinterval.InclusiveRange{Begin: 1, End: 5},
									},
								},
							},
						},
					},
				},
				TimeIntervals: []*timeInterval{
					{
						Name: "sunday",
						TimeIntervals: []timeinterval.TimeInterval{
							{
								Weekdays: []timeinterval.WeekdayRange{
									{
										InclusiveRange: timeinterval.InclusiveRange{Begin: 0, End: 0},
									},
								},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got: %q", err)
			}

			require.Equal(t, tc.expect, *tc.in)
		})
	}
}

func TestSanitizeEmailConfig(t *testing.T) {
	logger := log.NewNopLogger()

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		expect         alertmanagerConfig
		expectErr      bool
	}{
		{
			name:           "Test smtp_auth_password takes precedence in global config",
			againstVersion: semver.Version{Major: 0, Minor: 25},
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SMTPAuthPassword:     "foo",
					SMTPAuthPasswordFile: "bar",
				},
			},
			expect: alertmanagerConfig{
				Global: &globalConfig{
					SMTPAuthPassword: "foo",
				},
			},
		},
		{
			name:           "Test smtp_auth_password_file is dropped for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 24},
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SMTPAuthPasswordFile: "bar",
				},
			},
			expect: alertmanagerConfig{
				Global: &globalConfig{
					SMTPAuthPasswordFile: "",
				},
			},
		},
		{
			name:           "Test smtp_auth_password takes precedence in email config",
			againstVersion: semver.Version{Major: 0, Minor: 25},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						EmailConfigs: []*emailConfig{
							{
								AuthPassword:     "foo",
								AuthPasswordFile: "bar",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						EmailConfigs: []*emailConfig{
							{
								AuthPassword: "foo",
							},
						},
					},
				},
			},
		},
		{
			name:           "Test smtp_auth_password_file is dropped in slack config for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 24},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						EmailConfigs: []*emailConfig{
							{
								AuthPasswordFile: "bar",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						EmailConfigs: []*emailConfig{
							{
								AuthPasswordFile: "",
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got: %q", err)
			}

			require.Equal(t, tc.expect, *tc.in)
		})
	}
}

func TestSanitizeVictorOpsConfig(t *testing.T) {
	logger := log.NewNopLogger()

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		expect         alertmanagerConfig
		expectErr      bool
	}{
		{
			name:           "Test victorops_api_key takes precedence in global config",
			againstVersion: semver.Version{Major: 0, Minor: 25},
			in: &alertmanagerConfig{
				Global: &globalConfig{
					VictorOpsAPIKey:     "foo",
					VictorOpsAPIKeyFile: "bar",
				},
			},
			expect: alertmanagerConfig{
				Global: &globalConfig{
					VictorOpsAPIKey: "foo",
				},
			},
		},
		{
			name:           "Test victorops_api_key_file is dropped for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 24},
			in: &alertmanagerConfig{
				Global: &globalConfig{
					VictorOpsAPIKeyFile: "bar",
				},
			},
			expect: alertmanagerConfig{
				Global: &globalConfig{
					VictorOpsAPIKeyFile: "",
				},
			},
		},
		{
			name:           "Test api_key takes precedence in victorops config",
			againstVersion: semver.Version{Major: 0, Minor: 25},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						VictorOpsConfigs: []*victorOpsConfig{
							{
								APIKey:     "foo",
								APIKeyFile: "bar",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						VictorOpsConfigs: []*victorOpsConfig{
							{
								APIKey: "foo",
							},
						},
					},
				},
			},
		},
		{
			name:           "Test api_key_file is dropped in victorops config for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 24},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						VictorOpsConfigs: []*victorOpsConfig{
							{
								APIKeyFile: "bar",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						VictorOpsConfigs: []*victorOpsConfig{
							{
								APIKeyFile: "",
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got: %q", err)
			}

			require.Equal(t, tc.expect, *tc.in)
		})
	}
}

func TestSanitizePagerDutyConfig(t *testing.T) {
	logger := log.NewNopLogger()

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		expect         alertmanagerConfig
		expectErr      bool
	}{
		{
			name:           "Test routing_key takes precedence in pagerduty config",
			againstVersion: semver.Version{Major: 0, Minor: 25},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PagerdutyConfigs: []*pagerdutyConfig{
							{
								RoutingKey:     "foo",
								RoutingKeyFile: "bar",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						PagerdutyConfigs: []*pagerdutyConfig{
							{
								RoutingKey: "foo",
							},
						},
					},
				},
			},
		},
		{
			name:           "Test routing_key_file is dropped in pagerduty config for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 24},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PagerdutyConfigs: []*pagerdutyConfig{
							{
								RoutingKeyFile: "bar",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						PagerdutyConfigs: []*pagerdutyConfig{
							{
								RoutingKeyFile: "",
							},
						},
					},
				},
			},
		},
		{
			name:           "Test service_key takes precedence in pagerduty config",
			againstVersion: semver.Version{Major: 0, Minor: 25},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PagerdutyConfigs: []*pagerdutyConfig{
							{
								ServiceKey:     "foo",
								ServiceKeyFile: "bar",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						PagerdutyConfigs: []*pagerdutyConfig{
							{
								ServiceKey: "foo",
							},
						},
					},
				},
			},
		},
		{
			name:           "Test service_key_file is dropped in pagerduty config for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 24},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PagerdutyConfigs: []*pagerdutyConfig{
							{
								ServiceKeyFile: "bar",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						PagerdutyConfigs: []*pagerdutyConfig{
							{
								ServiceKeyFile: "",
							},
						},
					},
				},
			},
		},
		{
			name:           "Test source is dropped in pagerduty config for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 24},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PagerdutyConfigs: []*pagerdutyConfig{
							{
								Source: "foo",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						PagerdutyConfigs: []*pagerdutyConfig{
							{
								Source: "",
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got: %q", err)
			}

			require.Equal(t, tc.expect, *tc.in)
		})
	}
}

func TestSanitizeRoute(t *testing.T) {
	logger := log.NewNopLogger()
	matcherV2SyntaxAllowed := semver.Version{Major: 0, Minor: 22}
	matcherV2SyntaxNotAllowed := semver.Version{Major: 0, Minor: 21}

	namespaceLabel := "namespace"
	namespaceValue := "test-ns"

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *route
		expectErr      bool
		expect         route
	}{
		{
			name:           "Test route with new syntax not supported fails",
			againstVersion: matcherV2SyntaxNotAllowed,
			in: &route{
				Receiver: "test",
				Match: map[string]string{
					namespaceLabel: namespaceValue,
				},
				Matchers: []string{fmt.Sprintf("%s=%s", namespaceLabel, namespaceValue)},
				Continue: true,
				Routes: []*route{
					{
						Match: map[string]string{
							"keep": "me",
						},
						Matchers: []string{"strip=~me"},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "Test route with new syntax supported and no child routes",
			againstVersion: matcherV2SyntaxAllowed,
			in: &route{
				Receiver: "test",
				Match: map[string]string{
					namespaceLabel: namespaceValue,
				},
				Matchers: []string{fmt.Sprintf("%s=%s", namespaceLabel, namespaceValue)},
				Continue: true,
			},
			expect: route{
				Receiver: "test",
				Match: map[string]string{
					namespaceLabel: namespaceValue,
				},
				Matchers: []string{fmt.Sprintf("%s=%s", namespaceLabel, namespaceValue)},
				Continue: true,
			},
		},
		{
			name:           "Test route with new syntax supported with child routes",
			againstVersion: matcherV2SyntaxAllowed,
			in: &route{
				Receiver: "test",
				Match: map[string]string{
					"some": "value",
				},
				Matchers: []string{fmt.Sprintf("%s=%s", namespaceLabel, namespaceValue)},
				Continue: true,
				Routes: []*route{
					{
						Match: map[string]string{
							"keep": "me",
						},
						Matchers: []string{"keep=~me"},
					},
				},
			},
			expect: route{
				Receiver: "test",
				Match: map[string]string{
					"some": "value",
				},
				Matchers: []string{fmt.Sprintf("%s=%s", namespaceLabel, namespaceValue)},
				Continue: true,
				Routes: []*route{
					{
						Match: map[string]string{
							"keep": "me",
						},
						Matchers: []string{"keep=~me"},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("wanted %v but got error %s", tc.expect, err.Error())
			}

			require.Equal(t, tc.expect, *tc.in)
		})
	}
}

// We want to ensure that the imported types from config.MuteTimeInterval
// and any others with custom marshalling/unmarshalling are parsed
// into the internal struct as expected
func TestLoadConfig(t *testing.T) {
	testCase := []struct {
		name     string
		rawConf  []byte
		expected *alertmanagerConfig
	}{
		{
			name: "mute_time_intervals field",
			rawConf: []byte(`route:
  receiver: "null"
receivers:
- name: "null"
mute_time_intervals:
- name: maintenance_windows
  time_intervals:
  - times:
    - start_time: "17:00"
      end_time: "24:00"
    days_of_month: ["7", "18", "28"]
    months: ["january"]
templates: []
`),
			expected: &alertmanagerConfig{
				Global: nil,
				Route: &route{
					Receiver: "null",
				},
				Receivers: []*receiver{
					{
						Name: "null",
					},
				},
				MuteTimeIntervals: []*timeInterval{
					{
						Name: "maintenance_windows",
						TimeIntervals: []timeinterval.TimeInterval{
							{
								Months: []timeinterval.MonthRange{
									{
										InclusiveRange: timeinterval.InclusiveRange{
											Begin: 1,
											End:   1,
										},
									},
								},
								DaysOfMonth: []timeinterval.DayOfMonthRange{
									{
										InclusiveRange: timeinterval.InclusiveRange{
											Begin: 7,
											End:   7,
										},
									},
									{
										InclusiveRange: timeinterval.InclusiveRange{
											Begin: 18,
											End:   18,
										},
									},
									{
										InclusiveRange: timeinterval.InclusiveRange{
											Begin: 28,
											End:   28,
										},
									},
								},
								Times: []timeinterval.TimeRange{
									{
										StartMinute: 1020,
										EndMinute:   1440,
									},
								},
							},
						},
					},
				},
				Templates: []string{},
			},
		},
		{
			name: "Global opsgenie_api_key_file field",
			rawConf: []byte(`route:
  receiver: "null"
receivers:
- name: "null"
global:
  opsgenie_api_key_file: "xxx"
templates: []
`),
			expected: &alertmanagerConfig{
				Global: &globalConfig{
					OpsGenieAPIKeyFile: "xxx",
				},
				Route: &route{
					Receiver: "null",
				},
				Receivers: []*receiver{
					{
						Name: "null",
					},
				},
				Templates: []string{},
			},
		},
		{
			name: "OpsGenie entity and actions fields",
			rawConf: []byte(`route:
  receiver: "opsgenie"
receivers:
- name: "opsgenie"
  opsgenie_configs:
  - entity: entity1
    actions: action1,action2
    api_key: xxx
templates: []
`),
			expected: &alertmanagerConfig{
				Route: &route{
					Receiver: "opsgenie",
				},
				Receivers: []*receiver{
					{
						Name: "opsgenie",
						OpsgenieConfigs: []*opsgenieConfig{
							{
								Entity:  "entity1",
								Actions: "action1,action2",
								APIKey:  "xxx",
							},
						},
					},
				},
				Templates: []string{},
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ac, err := alertmanagerConfigFromBytes(tc.rawConf)
			if err != nil {
				t.Fatalf("expecing no error, got %v", err)
			}
			require.Equal(t, tc.expected, ac)
		})
	}
}

func parseURL(t *testing.T, u string) *config.URL {
	t.Helper()
	url, err := url.Parse(u)
	if err != nil {
		t.Fatalf("failed to parse URL %q: %s", u, err)
	}
	return &config.URL{URL: url}
}
