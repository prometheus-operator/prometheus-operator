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
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/timeinterval"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"

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
					From: ptr.To("from"),
					SmartHost: &monitoringingv1.HostPort{
						Host: "smtp.example.org",
						Port: "587",
					},
					Hello:        ptr.To("smtp.example.org"),
					AuthUsername: ptr.To("dev@smtp.example.org"),
					AuthPassword: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "smtp-auth",
						},
						Key: "password",
					},
					AuthIdentity: ptr.To("dev@smtp.example.org"),
					AuthSecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "smtp-auth",
						},
						Key: "secret",
					},
					RequireTLS: ptr.To(true),
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
					FollowRedirects: ptr.To(true),
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
					ResolveTimeout: ptr.To(model.Duration(30 * time.Second)),
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
					SMTPRequireTLS:   ptr.To(true),
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
						FollowRedirects: ptr.To(true),
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
					FollowRedirects: ptr.To(true),
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
						FollowRedirects: ptr.To(true),
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
		golden          string
	}
	version24, err := semver.ParseTolerant("v0.24.0")
	if err != nil {
		t.Fatal(err)
	}

	version26, err := semver.ParseTolerant("v0.26.0")
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
			golden:    "skeleton_base_no_CRs.golden",
		},
		{
			name:    "skeleton base with global send_revolved, no CRs",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Global: &globalConfig{
					ResolveTimeout: ptr.To(model.Duration(time.Minute)),
				},
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{},
			golden:    "skeleton_base_with_global_send_revolved_no_CRs.golden",
		},
		{
			name:    "skeleton base with global smtp_require_tls set to false, no CRs",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Global: &globalConfig{
					SMTPRequireTLS: ptr.To(false),
				},
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{},
			golden:    "skeleton_base_with_global_smtp_require_tls_set_to_false,_no_CRs.golden",
		},
		{
			name:    "skeleton base with global smtp_require_tls set to true, no CRs",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Global: &globalConfig{
					SMTPRequireTLS: ptr.To(true),
				},
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{},
			golden:    "skeleton_base_with_global_smtp_require_tls_set_to_true_no_CRs.golden",
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
			golden:    "skeleton_base_with_inhibit_rules_no_CRs.golden",
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
			golden:    "base_with_sub_route_and_matchers_no_CRs.golden",
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
			golden:    "skeleton_base_with_mute_time_intervals_no_CRs.golden",
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
			golden:    "skeleton_base_with_sns_receiver_no_CRs.golden",
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
			golden:    "skeleton_base_with_active_time_intervals_no_CRs.golden",
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
			golden: "skeleton_base_simple_CR.golden",
		},
		{
			name:    "multiple AlertmanagerConfig objects",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route: &route{
					Receiver: "null",
					Routes: []*route{
						{
							Receiver:   "watchdog",
							Matchers:   []string{"alertname=Watchdog"},
							GroupByStr: []string{"alertname"},
						},
					},
				},
				Receivers: []*receiver{{Name: "null"}, {Name: "watchdog"}},
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"ns1/amc1": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "amc1",
						Namespace: "ns1",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test1",
							GroupBy:  []string{"job"},
						},
						Receivers: []monitoringv1alpha1.Receiver{{Name: "test1"}},
					},
				},
				"ns1/amc2": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "amc2",
						Namespace: "ns1",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test2",
							GroupBy:  []string{"instance"},
						},
						Receivers: []monitoringv1alpha1.Receiver{{Name: "test2"}},
					},
				},
				"ns2/amc1": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "amc1",
						Namespace: "ns2",
					},
					Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
						Route: &monitoringv1alpha1.Route{
							Receiver: "test3",
							GroupBy:  []string{"job", "instance"},
						},
						Receivers: []monitoringv1alpha1.Receiver{{Name: "test3"}},
					},
				},
			},
			golden: "skeleton_base_multiple_alertmanagerconfigs.golden",
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
			golden: "skeleton_base_simple_CR_with_namespaceMatcher_disabled.golden",
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
			golden: "skeleton_base_CR_with_inhibition_rules_only_deprecated_matchers_not_converted.golden",
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
			golden: "skeleton_base_CR_with_inhibition_rules_only_deprecated_matchers_are_converted.golden",
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
			golden: "skeleton_base,_CR_with_inhibition_rules_only.golden",
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
			golden: "base_with_subroute_deprecated_matching_pattern_simple_CR.golden",
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
			golden: "CR_with_Pagerduty_Receiver.golden",
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
								URL: ptr.To("http://test.url"),
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
									FollowRedirects: ptr.To(true),
								},
							}},
						}},
					},
				},
			},
			golden: "CR_with_Webhook_Receiver_and_custom_http_config_oauth2.golden",
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
			golden: "CR_with_Opsgenie_Receiver.golden",
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
			golden: "CR_with_Opsgenie_Team_Responder.golden",
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
			golden: "CR_with_WeChat_Receiver.golden",
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
			golden: "CR_with_Telegram_Receiver.golden",
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
			golden: "CR_with_Slack_Receiver_and_global_Slack_URL.golden",
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
			golden: "CR_with_Slack_Receiver_and_global_Slack_URL_File.golden",
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
			golden: "CR_with_SNS_Receiver_with_Access_and_Key.golden",
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
			golden: "CR_with_SNS_Receiver_with_roleARN.golden",
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
			golden: "CR_with_Mute_Time_Intervals.golden",
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
			golden: "CR_with_Active_Time_Intervals.golden",
		},
		{
			name:      "CR with MSTeams Receiver",
			amVersion: &version26,
			kclient: fake.NewSimpleClientset(
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ms-teams-secret",
						Namespace: "mynamespace",
					},
					Data: map[string][]byte{
						"url": []byte("https://webhook.office.com/webhookb2/id/IncomingWebhook/id"),
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
						Receivers: []monitoringv1alpha1.Receiver{
							{
								Name: "test",
								MSTeamsConfigs: []monitoringv1alpha1.MSTeamsConfig{
									{
										WebhookURL: corev1.SecretKeySelector{
											Key: "url",
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "ms-teams-secret",
											},
										},
										Title: ptr.To("test title"),
										Text:  ptr.To("test text"),
									},
								},
							},
						},
					},
				},
			},
			golden: "CR_with_MSTeams_Receiver.golden",
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
			golden.Assert(t, string(cfgBytes), tc.golden)

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

	versionTelegramBotTokenFileAllowed := semver.Version{Major: 0, Minor: 26}
	versionTelegramBotTokenFileNotAllowed := semver.Version{Major: 0, Minor: 25}

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
		{
			name:           "bot_token_file field for Telegram config",
			againstVersion: versionTelegramBotTokenFileAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "telegram",
						TelegramConfigs: []*telegramConfig{
							{
								ChatID:       12345,
								BotTokenFile: "/test",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "telegram",
						TelegramConfigs: []*telegramConfig{
							{
								ChatID:       12345,
								BotTokenFile: "/test",
							},
						},
					},
				},
			},
		},
		{
			name:           "bot_token_file and bot_token fields for Telegram config",
			againstVersion: versionTelegramBotTokenFileAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "telegram",
						TelegramConfigs: []*telegramConfig{
							{
								ChatID:       12345,
								BotToken:     "test",
								BotTokenFile: "/test",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "telegram",
						TelegramConfigs: []*telegramConfig{{
							ChatID:   12345,
							BotToken: "test",
						}},
					},
				},
			},
		},
		{
			name:           "bot_token not specified and bot_token_file is dropped for unsupported versions",
			againstVersion: versionTelegramBotTokenFileNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "telegram",
						TelegramConfigs: []*telegramConfig{
							{
								ChatID:       12345,
								BotTokenFile: "/test",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "bot_token specified and bot_token_file is dropped for unsupported versions",
			againstVersion: versionTelegramBotTokenFileNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "telegram",
						TelegramConfigs: []*telegramConfig{
							{
								ChatID:       12345,
								BotToken:     "test",
								BotTokenFile: "/test",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "telegram",
						TelegramConfigs: []*telegramConfig{{
							ChatID:   12345,
							BotToken: "test",
						}},
					},
				},
			},
		},
		{
			name:           "bot_token and bot_token_file empty",
			againstVersion: versionTelegramBotTokenFileAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "telegram",
						TelegramConfigs: []*telegramConfig{
							{
								ChatID: 12345,
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
				EnableHTTP2: ptr.To(false),
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
				EnableHTTP2: ptr.To(false),
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
				EnableHTTP2: ptr.To(false),
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
				EnableHTTP2: ptr.To(false),
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
				EnableHTTP2: ptr.To(false),
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
				EnableHTTP2: ptr.To(false),
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
func TestSanitizePushoverReceiverConfig(t *testing.T) {
	logger := log.NewNopLogger()

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		expect         alertmanagerConfig
		expectErr      bool
	}{
		{
			name:           "Test pushover user_key/token takes precedence in pushover config",
			againstVersion: semver.Version{Major: 0, Minor: 26},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey:     "foo",
								UserKeyFile: "/path/use_key_file",
								Token:       "bar",
								TokenFile:   "/path/token_file",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey:     "foo",
								UserKeyFile: "",
								Token:       "bar",
								TokenFile:   "",
							},
						},
					},
				},
			},
		},
		{
			name:           "Test pushover token or token_file must be configured",
			againstVersion: semver.Version{Major: 0, Minor: 26},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey: "foo",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "Test pushover user_key or user_key_file must be configured",
			againstVersion: semver.Version{Major: 0, Minor: 26},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								Token: "bar",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "Test pushover user_key/token_file dropped in pushover config for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 25},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey:   "foo",
								TokenFile: "/path/token_file",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "Test pushover user_key_file/token dropped in pushover config for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 25},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKeyFile: "/path/use_key_file",
								Token:       "bar",
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
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
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

func TestSanitizeWebhookConfig(t *testing.T) {
	logger := log.NewNopLogger()

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		expect         alertmanagerConfig
		expectErr      bool
	}{
		{
			name:           "Test webhook_url_file is dropped in webhook config for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 25},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						WebhookConfigs: []*webhookConfig{
							{
								URLFile: "foo",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						WebhookConfigs: []*webhookConfig{
							{
								URLFile: "",
							},
						},
					},
				},
			},
		},
		{
			name:           "Test url takes precedence in webhook config",
			againstVersion: semver.Version{Major: 0, Minor: 26},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						WebhookConfigs: []*webhookConfig{
							{
								URL:     "foo",
								URLFile: "bar",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						WebhookConfigs: []*webhookConfig{
							{
								URL: "foo",
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

func TestSanitizePushoverConfig(t *testing.T) {
	logger := log.NewNopLogger()

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		expect         alertmanagerConfig
		expectErr      bool
	}{
		{
			name:           "Test pushover_user_key_file is dropped in pushover config for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 25},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey:     "key",
								UserKeyFile: "foo",
								Token:       "token",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey:     "key",
								UserKeyFile: "",
								Token:       "token",
							},
						},
					},
				},
			},
		},
		{
			name:           "Test pushover_token_file is dropped in pushover config for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 25},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey:   "key",
								Token:     "token",
								TokenFile: "foo",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey:   "key",
								Token:     "token",
								TokenFile: "",
							},
						},
					},
				},
			},
		},
		{
			name:           "Test user_key takes precedence in pushover config",
			againstVersion: semver.Version{Major: 0, Minor: 26},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey:     "foo",
								UserKeyFile: "bar",
								Token:       "token",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey: "foo",
								Token:   "token",
							},
						},
					},
				},
			},
		},
		{
			name:           "Test token takes precedence in pushover config",
			againstVersion: semver.Version{Major: 0, Minor: 26},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey:   "foo",
								Token:     "foo",
								TokenFile: "bar",
							},
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey: "foo",
								Token:   "foo",
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
		expected *alertmanagerConfig
		golden   string
	}{
		{
			name:   "mute_time_intervals field",
			golden: "mute_time_intervals_field.golden",
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
			name:   "Global opsgenie_api_key_file field",
			golden: "Global_opsgenie_api_key_file_field.golden",
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
			name:   "OpsGenie entity and actions fields",
			golden: "OpsGenie_entity_and_actions_fields.golden",
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
		{
			name:   "Discord url field",
			golden: "Discord_url_field.golden",
			expected: &alertmanagerConfig{
				Route: &route{
					Receiver: "discord",
				},
				Receivers: []*receiver{
					{
						Name: "discord",
						DiscordConfigs: []*discordConfig{
							{
								WebhookURL: "http://example.com",
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
			ac, err := alertmanagerConfigFromBytes(golden.Get(t, tc.golden))
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
