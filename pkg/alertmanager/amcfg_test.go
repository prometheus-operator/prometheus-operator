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
	"log/slog"
	"net/url"
	"testing"
	"time"

	"github.com/blang/semver/v4"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/timeinterval"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"gotest.tools/v3/golden"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

func mustMarshalRoute(r monitoringv1alpha1.Route) []byte {
	b, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	return b
}

func TestInitializeFromAlertmanagerConfig(t *testing.T) {
	myrouteJSON := mustMarshalRoute(monitoringv1alpha1.Route{
		Receiver: "myreceiver",
		Matchers: []monitoringv1alpha1.Matcher{
			{
				Name:  "mykey",
				Value: "myvalue",
				Regex: false,
			},
			{
				Name:  "mykey1",
				Value: "myvalue1",
				Regex: false,
			},
		},
	})

	version21, err := semver.ParseTolerant("v0.21.0")
	require.NoError(t, err)

	version24, err := semver.ParseTolerant("v0.24.0")
	require.NoError(t, err)

	version26, err := semver.ParseTolerant("v0.26.0")
	require.NoError(t, err)

	version28, err := semver.ParseTolerant("v0.28.0")
	require.NoError(t, err)

	pagerdutyURL := "example.pagerduty.com"
	invalidPagerdutyURL := "://example.pagerduty.com"

	telegramAPIURL := "https://telegram.example.com"
	invalidTelegramAPIURL := "://telegram.example.com"

	jiraAPIURL := "https://jira.example.com"
	invalidJiraAPIURL := "://jira.example.com"

	rocketChatAPIURL := "https://rocketchat.example.com"
	invalidRocketChatAPIURL := "://rocketchat.example.com"

	webexAPIURL := "https://webex.example.com"
	invalidWebexAPIURL := "://webex.example.com"

	weChatAPIURL := "https://wechat.example.com"
	invalidWeChatAPIURL := "://wechat.example.com"
	wechatCorpID := "mywechatcorpid"

	victorOpsAPIURL := "https://victorops.example.com"
	invalidVictorOpsAPIURL := "://victorops.example.com"

	tests := []struct {
		name            string
		amVersion       *semver.Version
		globalConfig    *monitoringv1.AlertmanagerGlobalConfig
		matcherStrategy monitoringv1.AlertmanagerConfigMatcherStrategy
		amConfig        *monitoringv1alpha1.AlertmanagerConfig
		wantErr         bool
		golden          string
	}{
		{
			name:      "valid global config",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				SMTPConfig: &monitoringv1.GlobalSMTPConfig{
					From: ptr.To("from"),
					SmartHost: &monitoringv1.HostPort{
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
					TLSConfig: &monitoringv1.SafeTLSConfig{
						InsecureSkipVerify: ptr.To(true),
						MinVersion:         ptr.To(monitoringv1.TLSVersion12),
						MaxVersion:         ptr.To(monitoringv1.TLSVersion13),
					},
				},
				ResolveTimeout: "30s",
				HTTPConfigWithProxy: &monitoringv1.HTTPConfigWithProxy{
					HTTPConfig: monitoringv1.HTTPConfig{
						HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
							OAuth2: &monitoringv1.OAuth2{
								ClientID: monitoringv1.SecretOrConfigMap{
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_global_config.golden",
		},
		{
			name:      "valid global config with global HTTPConfig CA",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				HTTPConfigWithProxy: &monitoringv1.HTTPConfigWithProxy{
					HTTPConfig: monitoringv1.HTTPConfig{
						HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
							FollowRedirects: ptr.To(true),
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								ConfigMap: &corev1.ConfigMapKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "proxy-ca-certificate",
									},
									Key: "certificate",
								},
							},
						},
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_global_config_with_global_httpconfig_ca.golden",
		},
		{
			name: "valid global config with Slack API URL",
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_global_config_with_Slack_API_URL.golden",
		},
		{
			name: "global config with invalid Slack API URL",
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
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
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name: "global config with missing Slack API URL",
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
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
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name: "valid global config with OpsGenie API URL",
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_global_config_with_OpsGenie_API_URL.golden",
		},
		{
			name: "global config with invalid OpsGenie API URL",
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
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
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name: "global config with missing OpsGenie API URL",
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
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
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name: "valid global config with OpsGenie API KEY",
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_global_config_with_OpsGenie_API_KEY.golden",
		},
		{
			name: "global config with missing OpsGenie API KEY",
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
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
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name: "valid global config with Pagerduty URL",
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				PagerdutyURL: ptr.To(monitoringv1.URL(pagerdutyURL)),
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_global_config_with_Pagerduty_URL.golden",
		},
		{
			name: "global config with invalid Pagerduty URL",
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				PagerdutyURL: ptr.To(monitoringv1.URL(invalidPagerdutyURL)),
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
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
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
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				HTTPConfigWithProxy: &monitoringv1.HTTPConfigWithProxy{
					HTTPConfig: monitoringv1.HTTPConfig{
						HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
							FollowRedirects: ptr.To(true),
						},
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
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "globalConfig_has_null_resolve_timeout.golden",
		},
		{
			name: "globalConfig httpconfig/proxyconfig has null secretKey for proxyConnectHeader",
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				HTTPConfigWithProxy: &monitoringv1.HTTPConfigWithProxy{
					HTTPConfig: monitoringv1.HTTPConfig{
						HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
							FollowRedirects: ptr.To(true),
						},
					},
					ProxyConfig: monitoringv1.ProxyConfig{
						ProxyURL: ptr.To("http://example.com"),
						NoProxy:  ptr.To("svc.cluster.local"),
						ProxyConnectHeader: map[string][]corev1.SecretKeySelector{
							"header": {
								{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "no-secret",
									},
									Key: "proxy-header",
								},
							},
						},
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
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "valid globalConfig httpconfig/proxyconfig/proxyConnectHeader with amVersion24",
			amVersion: &version24,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				HTTPConfigWithProxy: &monitoringv1.HTTPConfigWithProxy{
					ProxyConfig: monitoringv1.ProxyConfig{
						ProxyURL: ptr.To("http://example.com"),
						NoProxy:  ptr.To("svc.cluster.local"),
						ProxyConnectHeader: map[string][]corev1.SecretKeySelector{
							"header": {
								{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "secret",
									},
									Key: "proxy-header",
								},
							},
						},
					},
					HTTPConfig: monitoringv1.HTTPConfig{
						HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
							FollowRedirects: ptr.To(true),
						},
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
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_globalConfig_httpconfig_proxyconfig_proxyConnectHeader_with_amVersion24.golden",
		},
		{
			name:      "valid globalConfig httpconfig/proxyconfig/proxyConnectHeader with amVersion26",
			amVersion: &version26,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				HTTPConfigWithProxy: &monitoringv1.HTTPConfigWithProxy{
					ProxyConfig: monitoringv1.ProxyConfig{
						ProxyURL: ptr.To("http://example.com"),
						NoProxy:  ptr.To("svc.cluster.local"),
						ProxyConnectHeader: map[string][]corev1.SecretKeySelector{
							"header": {
								{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "secret",
									},
									Key: "proxy-header",
								},
							},
						},
					},
					HTTPConfig: monitoringv1.HTTPConfig{
						HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
							FollowRedirects: ptr.To(true),
						},
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
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_globalConfig_httpconfig_proxyconfig_proxyConnectHeader_with_amVersion26.golden",
		},
		{
			name: "invalid alertmanagerConfig with invalid child routes",
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
						Routes: []apiextensionsv1.JSON{
							{
								Raw: []byte(`{"receiver": "recv2", "matchers": [{"severity":"!=critical$"}]}`),
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "valid global config SMTP tlsconfig email receiver",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				SMTPConfig: &monitoringv1.GlobalSMTPConfig{
					From: ptr.To("from"),
					SmartHost: &monitoringv1.HostPort{
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
					TLSConfig: &monitoringv1.SafeTLSConfig{
						InsecureSkipVerify: ptr.To(true),
						MinVersion:         ptr.To(monitoringv1.TLSVersion12),
						MaxVersion:         ptr.To(monitoringv1.TLSVersion13),
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
						{
							Name: "myreceiver",
							EmailConfigs: []monitoringv1alpha1.EmailConfig{
								{
									SendResolved: ptr.To(true),
									Smarthost:    "abc:1234",
									From:         "a",
									To:           "b",
									AuthUsername: "foo",
								},
							},
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_global_config_smtp_tlsconfig_email_receiver.golden",
		},
		{
			name:      "valid global config with amVersion21 ",
			amVersion: &version21,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				SMTPConfig: &monitoringv1.GlobalSMTPConfig{
					From: ptr.To("from"),
					SmartHost: &monitoringv1.HostPort{
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
					TLSConfig: &monitoringv1.SafeTLSConfig{
						InsecureSkipVerify: ptr.To(true),
						MinVersion:         ptr.To(monitoringv1.TLSVersion12),
						MaxVersion:         ptr.To(monitoringv1.TLSVersion13),
					},
				},
				ResolveTimeout: "30s",
				HTTPConfigWithProxy: &monitoringv1.HTTPConfigWithProxy{
					HTTPConfig: monitoringv1.HTTPConfig{
						HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
							OAuth2: &monitoringv1.OAuth2{
								ClientID: monitoringv1.SecretOrConfigMap{
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_global_config_with_amVersion21.golden",
		},
		{
			name:      "valid global config telegram api url",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				TelegramConfig: &monitoringv1.GlobalTelegramConfig{
					APIURL: ptr.To(monitoringv1.URL(telegramAPIURL)),
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_global_config_with_Telegram_API_URL.golden",
		},
		{
			name:      "invalid global config telegram api url",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				TelegramConfig: &monitoringv1.GlobalTelegramConfig{
					APIURL: ptr.To(monitoringv1.URL(invalidTelegramAPIURL)),
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "invalid global config telegram api url version not supported",
			amVersion: &version21,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				TelegramConfig: &monitoringv1.GlobalTelegramConfig{
					APIURL: ptr.To(monitoringv1.URL(telegramAPIURL)),
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "valid global config jira api url",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				JiraConfig: &monitoringv1.GlobalJiraConfig{
					APIURL: ptr.To(monitoringv1.URL(jiraAPIURL)),
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_global_config_with_Jira_API_URL.golden",
		},
		{
			name:      "invalid global config jira api url",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				JiraConfig: &monitoringv1.GlobalJiraConfig{
					APIURL: ptr.To(monitoringv1.URL(invalidJiraAPIURL)),
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "invalid global config jira api url version not supported",
			amVersion: &version26,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				JiraConfig: &monitoringv1.GlobalJiraConfig{
					APIURL: ptr.To(monitoringv1.URL(jiraAPIURL)),
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "valid global config rocket chat",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				RocketChatConfig: &monitoringv1.GlobalRocketChatConfig{
					APIURL: ptr.To(monitoringv1.URL(rocketChatAPIURL)),
					Token: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "rocketchat",
						},
						Key: "token",
					},
					TokenID: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "rocketchat",
						},
						Key: "token_id",
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_global_config_with_RocketChat.golden",
		},
		{
			name:      "invalid global config rocket chat api url",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				RocketChatConfig: &monitoringv1.GlobalRocketChatConfig{
					APIURL: ptr.To(monitoringv1.URL(invalidRocketChatAPIURL)),
					Token: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "rocketchat",
						},
						Key: "token",
					},
					TokenID: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "rocketchat",
						},
						Key: "token_id",
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "invalid global config rocket chat token missing secret",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				RocketChatConfig: &monitoringv1.GlobalRocketChatConfig{
					APIURL: ptr.To(monitoringv1.URL(rocketChatAPIURL)),
					Token: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "rocketchat-missing",
						},
						Key: "token",
					},
					TokenID: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "rocketchat",
						},
						Key: "token_id",
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "invalid global config rocket chat token id missing secret",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				RocketChatConfig: &monitoringv1.GlobalRocketChatConfig{
					APIURL: ptr.To(monitoringv1.URL(rocketChatAPIURL)),
					Token: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "rocketchat",
						},
						Key: "token",
					},
					TokenID: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "rocketchat-missing",
						},
						Key: "token_id",
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "invalid global config rocket chat version not supported",
			amVersion: &version26,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				RocketChatConfig: &monitoringv1.GlobalRocketChatConfig{
					APIURL: ptr.To(monitoringv1.URL(rocketChatAPIURL)),
					Token: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "rocketchat",
						},
						Key: "token",
					},
					TokenID: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "rocketchat",
						},
						Key: "token_id",
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "valid global config webex api url",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				WebexConfig: &monitoringv1.GlobalWebexConfig{
					APIURL: ptr.To(monitoringv1.URL(webexAPIURL)),
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_global_config_with_Webex_API_URL.golden",
		},
		{
			name:      "invalid global config webex api url",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				WebexConfig: &monitoringv1.GlobalWebexConfig{
					APIURL: ptr.To(monitoringv1.URL(invalidWebexAPIURL)),
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "invalid global config webex api url version not supported",
			amVersion: &version24,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				WebexConfig: &monitoringv1.GlobalWebexConfig{
					APIURL: ptr.To(monitoringv1.URL(webexAPIURL)),
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "valid global config wechat config",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				WeChatConfig: &monitoringv1.GlobalWeChatConfig{
					APIURL: ptr.To(monitoringv1.URL(weChatAPIURL)),
					APISecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "wechat",
						},
						Key: "api_secret",
					},
					APICorpID: &wechatCorpID,
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_global_config_with_WeChat_Config.golden",
		},
		{
			name:      "invalid global config wechat config api url",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				WeChatConfig: &monitoringv1.GlobalWeChatConfig{
					APIURL: ptr.To(monitoringv1.URL(invalidWeChatAPIURL)),
					APISecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "wechat",
						},
						Key: "api_secret",
					},
					APICorpID: &wechatCorpID,
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "invalid global config wechat config missing secret",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				WeChatConfig: &monitoringv1.GlobalWeChatConfig{
					APIURL: ptr.To(monitoringv1.URL(weChatAPIURL)),
					APISecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "wechat-missing",
						},
						Key: "api_secret",
					},
					APICorpID: &wechatCorpID,
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "valid global config victorops",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				VictorOpsConfig: &monitoringv1.GlobalVictorOpsConfig{
					APIURL: ptr.To(monitoringv1.URL(victorOpsAPIURL)),
					APIKey: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "victorops",
						},
						Key: "api_key",
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			golden: "valid_global_config_with_VictorOps.golden",
		},
		{
			name:      "invalid global config victorops api url",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				VictorOpsConfig: &monitoringv1.GlobalVictorOpsConfig{
					APIURL: ptr.To(monitoringv1.URL(invalidVictorOpsAPIURL)),
					APIKey: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "victorops",
						},
						Key: "api_key",
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "invalid global config victorops api key missing",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				VictorOpsConfig: &monitoringv1.GlobalVictorOpsConfig{
					APIURL: ptr.To(monitoringv1.URL(victorOpsAPIURL)),
					APIKey: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "victorops-missing",
						},
						Key: "api_key",
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
		{
			name:      "invalid httpConfig global config basicAuth and oauth2 are both specified",
			amVersion: &version28,
			globalConfig: &monitoringv1.AlertmanagerGlobalConfig{
				ResolveTimeout: "30s",
				HTTPConfigWithProxy: &monitoringv1.HTTPConfigWithProxy{
					HTTPConfig: monitoringv1.HTTPConfig{
						HTTPConfigWithoutTLS: monitoringv1.HTTPConfigWithoutTLS{
							OAuth2: &monitoringv1.OAuth2{
								ClientID: monitoringv1.SecretOrConfigMap{
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
							BasicAuth: &monitoringv1.BasicAuth{
								Username: corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "webhook-client-secret",
									},
									Key: "test",
								},
								Password: corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "webhook-client-secret",
									},
									Key: "test",
								},
							},
							FollowRedirects: ptr.To(true),
						},
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
						{
							Name: "myreceiver",
						},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver: "null",
						Routes: []apiextensionsv1.JSON{
							{
								Raw: myrouteJSON,
							},
						},
					},
				},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespace",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		if tt.amVersion == nil {
			version, err := semver.ParseTolerant("v0.22.2")
			require.NoError(t, err)
			tt.amVersion = &version
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
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "proxy-ca-certificate",
					Namespace: "mynamespace",
				},
				Data: map[string]string{
					"certificate": `-----BEGIN CERTIFICATE-----
MIIDbTCCAlWgAwIBAgIUM7xicCKY+p54CUpjWsTl7KTssbIwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAgFw0yNTExMTExNjU5MTRaGA8yMTI1
MTAxODE2NTkxNFowRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUx
ITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAJZlFSLI4/t27LnnYnv+EsFXoyjryinP4NbaHjjB
gEEEHMo+GgL8cOu3VbDuzgpC3opJ+AHGsltXc+gZ86YPu8EzwiKB+Ci4p8K5z9+g
QW8WX9lpZn7z5WRm53llVLDZY/vCSzQ5KFQ8V/0vYJfxOYUapilH9mQqENnaw9dz
0VckluLgSLKA/A95p8Rp2Zt1tAtwjD3ClRQ1wricymbt+5qVt45zLC6MD32WizyV
vjCUc2kCZDyjHPZIauLoo0rQAiO/mX8nUJpxGf8yp/Rs7hL1tBAWSj9FFBvdUwz+
z9qRyE6ojp1HVSDpGyLsRXZwqiP5IL72iZlcoDRr1+zmWTMCAwEAAaNTMFEwHQYD
VR0OBBYEFC5yBxXPVqkvyq05rr7OzIdS1h2GMB8GA1UdIwQYMBaAFC5yBxXPVqkv
yq05rr7OzIdS1h2GMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEB
AFO+7n/RSw18NRZ8Z8uFLvZwWCc/PeZ5uo23m4AAVLr5vwWCIeh+DYG4u4xhyKd5
B3U7zgxU2/pmSS1kPDAIBooVe90C804OxfL3/QOurRC9Eugi441DvpkJ2Uy91PWA
E5G2s2fZRWUytKl0I7YqyeDlP96V34qi6P8e3GAWGoGExjJzQVomYXeVU/0eQkOC
Z8Ja2z8jw1xUKxfurno8wsAgFAQLuUZ0sTpwHBtwzFEdIeaAHBbNkkuGq7leIw/u
83OdaXOYthY8wG5jRwDRQSA0FGayQrKPj1+II2VMgU/ApF5zs7Gid32pi4iVyu1i
9MFjFOe4ShaqsQ9HgZuAZls=
-----END CERTIFICATE-----`,
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
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rocketchat",
					Namespace: "mynamespace",
				},
				Data: map[string][]byte{
					"token":    []byte("mytoken"),
					"token_id": []byte("mytokenid"),
				},
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wechat",
					Namespace: "mynamespace",
				},
				Data: map[string][]byte{
					"api_secret": []byte("mywechatsecret"),
				},
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "victorops",
					Namespace: "mynamespace",
				},
				Data: map[string][]byte{
					"api_key": []byte("myvictoropsapikey"),
				},
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "mynamespace",
				},
				Data: map[string][]byte{
					"proxy-header": []byte("value"),
				},
			},
		)
		cb := NewConfigBuilder(
			newNopLogger(t),
			*tt.amVersion,
			assets.NewStoreBuilder(kclient.CoreV1(), kclient.CoreV1()),
			&monitoringv1.Alertmanager{
				ObjectMeta: metav1.ObjectMeta{Namespace: "alertmanager-namespace"},
				Spec:       monitoringv1.AlertmanagerSpec{AlertmanagerConfigMatcherStrategy: tt.matcherStrategy},
			},
		)
		t.Run(tt.name, func(t *testing.T) {
			err := cb.initializeFromAlertmanagerConfig(context.TODO(), tt.globalConfig, tt.amConfig)
			if tt.wantErr {
				t.Logf("err: %s", err)
				require.Error(t, err)
				return
			}

			amConfigurations, err := yaml.Marshal(cb.cfg)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigurations), tt.golden)
		})
	}
}

func TestGenerateConfig(t *testing.T) {
	type testCase struct {
		name            string
		kclient         kubernetes.Interface
		baseConfig      alertmanagerConfig
		amVersion       *semver.Version
		matcherStrategy monitoringv1.AlertmanagerConfigMatcherStrategy
		amConfigs       map[string]*monitoringv1alpha1.AlertmanagerConfig
		golden          string
		expectedError   bool
	}
	version24, err := semver.ParseTolerant("v0.24.0")
	require.NoError(t, err)

	version26, err := semver.ParseTolerant("v0.26.0")
	require.NoError(t, err)

	version27, err := semver.ParseTolerant("v0.27.0")
	require.NoError(t, err)

	version28, err := semver.ParseTolerant("v0.28.0")
	require.NoError(t, err)

	globalSlackAPIURL, err := url.Parse("http://slack.example.com")
	require.NoError(t, err)

	testCases := []testCase{
		{
			name:    "skeleton base, no CRs",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			golden: "skeleton_base_no_CRs.golden",
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
			golden: "skeleton_base_with_global_send_revolved_no_CRs.golden",
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
			golden: "skeleton_base_with_global_smtp_require_tls_set_to_false,_no_CRs.golden",
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
			golden: "skeleton_base_with_global_smtp_require_tls_set_to_true_no_CRs.golden",
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
			golden: "skeleton_base_with_inhibit_rules_no_CRs.golden",
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
			golden: "base_with_sub_route_and_matchers_no_CRs.golden",
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
			golden: "skeleton_base_with_mute_time_intervals_no_CRs.golden",
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
			golden: "skeleton_base_with_sns_receiver_no_CRs.golden",
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
			golden: "skeleton_base_with_active_time_intervals_no_CRs.golden",
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
			name:    "skeleton base, CR with sub-routes",
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
							Routes: []apiextensionsv1.JSON{
								{
									Raw: mustMarshalRoute(monitoringv1alpha1.Route{
										Receiver: "test2",
										GroupBy:  []string{"job", "instance"},
										Continue: true,
										Matchers: []monitoringv1alpha1.Matcher{
											{Name: "job", Value: "foo", MatchType: "="},
										},
									}),
								},
								{
									Raw: mustMarshalRoute(monitoringv1alpha1.Route{
										Receiver: "test3",
										GroupBy:  []string{"job", "instance"},
										Matchers: []monitoringv1alpha1.Matcher{
											{Name: "job", Value: "bar", MatchType: "="},
										},
									}),
								},
							},
						},
						Receivers: []monitoringv1alpha1.Receiver{
							{Name: "test"}, {Name: "test2"}, {Name: "test3"},
						},
					},
				},
			},
			golden: "skeleton_base_CR_with_subroutes.golden",
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
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
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
			name:    "skeleton base in same namespace as alertmanager, simple CR with namespaceMatcher disabled for alertmanager namespace",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespaceExceptForAlertmanagerNamespace",
			},
			amConfigs: map[string]*monitoringv1alpha1.AlertmanagerConfig{
				"alertmanager-namespace": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myamc",
						Namespace: "alertmanager-namespace",
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
			golden: "skeleton_base_same_namespace_simple_CR_with_namespaceMatcher_disabled_other_namespaces.golden",
		},
		{
			name:    "skeleton base in different namespace to alertmanager, simple CR with namespaceMatcher disabled for alertmanager namespace",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route:     &route{Receiver: "null"},
				Receivers: []*receiver{{Name: "null"}},
			},
			matcherStrategy: monitoringv1.AlertmanagerConfigMatcherStrategy{
				Type: "OnNamespaceExceptForAlertmanagerNamespace",
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
			golden: "skeleton_base_diff_namespace_simple_CR_with_namespaceMatcher_disabled_other_namespaces.golden",
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
										Src:  ptr.To("https://some-image.com"),
										Href: ptr.To(monitoringv1alpha1.URL("https://some-image.com")),
										Alt:  ptr.To("some-image"),
									},
								},
								PagerDutyLinkConfigs: []monitoringv1alpha1.PagerDutyLinkConfig{
									{
										Href: ptr.To(monitoringv1alpha1.URL("https://some-link.com")),
										Text: ptr.To("some-link"),
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
								URL: ptr.To(monitoringv1alpha1.URL("http://test.url")),
								HTTPConfig: &monitoringv1alpha1.HTTPConfig{
									OAuth2: &monitoringv1.OAuth2{
										ClientID: monitoringv1.SecretOrConfigMap{
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
								APIURL: ptr.To(monitoringv1alpha1.URL("https://example.com/")),
							}},
						}},
					},
				},
			},
			golden: "CR_with_Opsgenie_Receiver_Valid_APIURL.golden",
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
			name: "CR with Pushover Receiver",
			kclient: fake.NewSimpleClientset(
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "am-pushover-test-receiver",
						Namespace: "mynamespace",
					},
					Data: map[string][]byte{
						"userkey": []byte("userkeySecret"),
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "am-pushover-token-receiver",
						Namespace: "mynamespace",
					},
					Data: map[string][]byte{
						"token": []byte("tokenSecret"),
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
							PushoverConfigs: []monitoringv1alpha1.PushoverConfig{{
								UserKey: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "am-pushover-test-receiver",
									},
									Key: "userkey",
								},
								Token: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "am-pushover-token-receiver",
									},
									Key: "token",
								},
								Retry:  ptr.To("5m"),
								Expire: ptr.To("30s"),
								HTML:   ptr.To(true),
							}},
						}},
					},
				},
			},
			golden: "CR_with_Pushover_Receiver.golden",
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
								APIURL: ptr.To(monitoringv1alpha1.URL("https://api.telegram.org")),
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
									Sigv4: &monitoringv1.Sigv4{
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
									Sigv4: &monitoringv1.Sigv4{
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
		{
			name:      "CR with MSTeams Receiver with Summary",
			amVersion: &version27,
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
										Title:   ptr.To("test title"),
										Summary: ptr.To("test summary"),
										Text:    ptr.To("test text"),
									},
								},
							},
						},
					},
				},
			},
			golden: "CR_with_MSTeams_Receiver_Summary.golden",
		},
		{
			name:      "CR with MSTeams Receiver with Partial Conf",
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
									},
								},
							},
						},
					},
				},
			},
			golden: "CR_with_MSTeams_Receiver_Partial_Conf.golden",
		},
		{
			name:      "CR with MSTeamsV2 Receiver",
			amVersion: &version28,
			kclient: fake.NewSimpleClientset(
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ms-teams-secret",
						Namespace: "mynamespace",
					},
					Data: map[string][]byte{
						"url": []byte("https://prod-108.westeurope.logic.azure.com:443/workflows/id"),
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
								MSTeamsV2Configs: []monitoringv1alpha1.MSTeamsV2Config{
									{
										WebhookURL: &corev1.SecretKeySelector{
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
			golden: "CR_with_MSTeamsV2_Receiver.golden",
		},
		{
			name:      "CR with MSTeamsV2 Receiver with Partial Conf",
			amVersion: &version28,
			kclient: fake.NewSimpleClientset(
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ms-teams-secret",
						Namespace: "mynamespace",
					},
					Data: map[string][]byte{
						"url": []byte("https://prod-108.westeurope.logic.azure.com:443/workflows/id"),
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
								MSTeamsV2Configs: []monitoringv1alpha1.MSTeamsV2Config{
									{
										WebhookURL: &corev1.SecretKeySelector{
											Key: "url",
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "ms-teams-secret",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			golden: "CR_with_MSTeamsV2_Receiver_Partial_Conf.golden",
		},
		{
			name:      "CR with EmailConfig with Required Fields specified at Receiver level",
			amVersion: &version26,
			kclient:   fake.NewSimpleClientset(),
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
								EmailConfigs: []monitoringv1alpha1.EmailConfig{
									{
										Smarthost: "example.com:25",
										From:      "admin@example.com",
										To:        "customers@example.com",
									},
								},
							},
						},
					},
				},
			},
			golden: "CR_with_EmailConfig_Receiver_Conf.golden",
		},
		{
			name:      "CR with EmailConfig Missing SmartHost Field",
			amVersion: &version26,
			kclient:   fake.NewSimpleClientset(),
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
								EmailConfigs: []monitoringv1alpha1.EmailConfig{
									{
										From: "admin@example.com",
										To:   "customers@example.com",
									},
								},
							},
						},
					},
				},
			},
			expectedError: true,
		},
		{
			name:      "CR with EmailConfig Missing SMTP From Field",
			amVersion: &version26,
			kclient:   fake.NewSimpleClientset(),
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
								EmailConfigs: []monitoringv1alpha1.EmailConfig{
									{
										From: "admin@example.com",
										To:   "customers@example.com",
									},
								},
							},
						},
					},
				},
			},
			expectedError: true,
		},
		{
			name:      "CR with EmailConfig Missing Required Fields from Receiver level but specified at Global level",
			amVersion: &version26,
			kclient:   fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Global: &globalConfig{
					SMTPSmarthost: config.HostPort{
						Host: "smtp.example.org",
						Port: "587",
					},
					SMTPFrom: "admin@globaltest.com",
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
						Receivers: []monitoringv1alpha1.Receiver{
							{
								Name: "test",
								EmailConfigs: []monitoringv1alpha1.EmailConfig{
									{
										To: "customers@example.com",
									},
								},
							},
						},
					},
				},
			},
			golden: "CR_with_EmailConfig_Receiver_Global_Defaults_Conf.golden",
		},
		{
			name:      "CR with WebhookConfig with Timeout Setup",
			amVersion: &version28,
			kclient:   fake.NewSimpleClientset(),
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
								WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
									{
										URL:     ptr.To(monitoringv1alpha1.URL("https://example.com/")),
										Timeout: ptr.To(monitoringv1.Duration("5s")),
									},
								},
							},
						},
					},
				},
			},
			golden: "CR_with_WebhookConfig_with_Timeout_Setup.golden",
		},
		{
			name:      "CR with WebhookConfig with Timeout Setup Older Version",
			amVersion: &version26,
			kclient:   fake.NewSimpleClientset(),
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
								WebhookConfigs: []monitoringv1alpha1.WebhookConfig{
									{
										URL:     ptr.To(monitoringv1alpha1.URL("https://example.com/")),
										Timeout: ptr.To(monitoringv1.Duration("5s")),
									},
								},
							},
						},
					},
				},
			},
			golden: "CR_with_WebhookConfig_with_Timeout_Setup_Older_Version.golden",
		},
	}

	logger := newNopLogger(t)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewStoreBuilder(tc.kclient.CoreV1(), tc.kclient.CoreV1())

			if tc.amVersion == nil {
				version, err := semver.ParseTolerant("v0.22.2")
				require.NoError(t, err)
				tc.amVersion = &version
			}

			cb := NewConfigBuilder(logger, *tc.amVersion, store,
				&monitoringv1.Alertmanager{
					ObjectMeta: metav1.ObjectMeta{Namespace: "alertmanager-namespace"},
					Spec:       monitoringv1.AlertmanagerSpec{AlertmanagerConfigMatcherStrategy: tc.matcherStrategy},
				},
			)
			cb.cfg = &tc.baseConfig

			if tc.expectedError {
				require.Error(t, cb.AddAlertmanagerConfigs(context.Background(), tc.amConfigs))
				return
			}
			require.NoError(t, cb.AddAlertmanagerConfigs(context.Background(), tc.amConfigs))

			cfgBytes, err := cb.MarshalJSON()
			require.NoError(t, err)

			// Verify the generated yaml is as expected
			golden.Assert(t, string(cfgBytes), tc.golden)

			// Verify the generated config is something that Alertmanager will be happy with
			_, err = alertmanagerConfigFromBytes(cfgBytes)
			require.NoError(t, err)
		})
	}
}

func TestSanitizeConfig(t *testing.T) {
	logger := newNopLogger(t)
	versionFileURLAllowed := semver.Version{Major: 0, Minor: 22}
	versionFileURLNotAllowed := semver.Version{Major: 0, Minor: 21}

	matcherV2SyntaxAllowed := semver.Version{Major: 0, Minor: 22}
	matcherV2SyntaxNotAllowed := semver.Version{Major: 0, Minor: 21}

	versionOpsGenieAPIKeyFileAllowed := semver.Version{Major: 0, Minor: 24}
	versionOpsGenieAPIKeyFileNotAllowed := semver.Version{Major: 0, Minor: 23}

	versionDiscordAllowed := semver.Version{Major: 0, Minor: 25}
	versionDiscordNotAllowed := semver.Version{Major: 0, Minor: 24}

	versionDiscordMessageFieldsAllowed := semver.Version{Major: 0, Minor: 28}
	versionDiscordMessageFieldsNotAllowed := semver.Version{Major: 0, Minor: 27}

	versionMSteamsV2Allowed := semver.Version{Major: 0, Minor: 28}
	versionMSteamsV2NotAllowed := semver.Version{Major: 0, Minor: 27}

	versionWebexAllowed := semver.Version{Major: 0, Minor: 25}
	versionWebexNotAllowed := semver.Version{Major: 0, Minor: 24}

	versionTelegramBotTokenFileAllowed := semver.Version{Major: 0, Minor: 26}
	versionTelegramBotTokenFileNotAllowed := semver.Version{Major: 0, Minor: 25}

	versionTelegramMessageThreadIDAllowed := semver.Version{Major: 0, Minor: 26}
	versionTelegramMessageThreadIDNotAllowed := semver.Version{Major: 0, Minor: 25}

	versionMSTeamsSummaryAllowed := semver.Version{Major: 0, Minor: 27}
	versionMSTeamsSummaryNotAllowed := semver.Version{Major: 0, Minor: 26}

	versionSMTPTLSConfigAllowed := semver.Version{Major: 0, Minor: 28}
	versionSMTPTLSConfigNotAllowed := semver.Version{Major: 0, Minor: 27}

	versionMattermostConfigAllowed := semver.Version{Major: 0, Minor: 30}
	versionMattermostConfigNotAllowed := semver.Version{Major: 0, Minor: 29}

	versionTimeoutConfigAllowed := semver.Version{Major: 0, Minor: 30}
	versionTimeoutConfigNotAllowed := semver.Version{Major: 0, Minor: 29}

	versionSlackAppConfigAllowed := semver.Version{Major: 0, Minor: 30}
	versionSlackAppConfigNotAllowed := semver.Version{Major: 0, Minor: 29}

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		expectErr      bool
		golden         string
	}{
		{
			name:           "Test smtp_tls_config is dropped for unsupported versions",
			againstVersion: versionSMTPTLSConfigNotAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SMTPTLSConfig: &tlsConfig{
						CAFile: "/var/kubernetes/secrets/tls/ca.txt",
					},
				},
			},
			golden: "test_smtp_tls_config_is_dropped_for_unsupported_versions.golden",
		},
		{
			name:           "Test smtp_tls_config is added for supported versions",
			againstVersion: versionSMTPTLSConfigAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SMTPTLSConfig: &tlsConfig{
						CAFile:     "/var/kubernetes/secrets/tls/ca.txt",
						MinVersion: "TLS12",
						MaxVersion: "TLS13",
					},
				},
			},
			golden: "test_smtp_tls_config_is_added_for_supported_versions.golden",
		},
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
			golden: "test_slack_api_url_takes_precedence_in_global_config.golden",
		},
		{
			name:           "Test slack_api_url_file is dropped for unsupported versions",
			againstVersion: versionFileURLNotAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SlackAPIURLFile: "/test",
				},
			},
			golden: "test_slack_api_url_file is dropped_for_unsupported_versions.golden",
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
			golden: "test_api_url_takes_precedence_in_slack_config.golden",
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
			golden: "test_api_url_file_is_dropped_in_slack_config_for_unsupported_versions.golden",
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
			golden: "test_slack_config_happy_path.golden",
		},
		{
			name:           "Test timeout is dropped in slack config for unsupported versions",
			againstVersion: versionTimeoutConfigNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						SlackConfigs: []*slackConfig{
							{
								Timeout: ptr.To(model.Duration(time.Minute)),
							},
						},
					},
				},
			},
			golden: "test_slack_timeout_is_dropped_in_slack_config_for_unsupported_versions.golden",
		},
		{
			name:           "Test timeout is added in slack config for supported versions",
			againstVersion: versionTimeoutConfigAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						SlackConfigs: []*slackConfig{
							{
								Timeout: ptr.To(model.Duration(time.Minute)),
							},
						},
					},
				},
			},
			golden: "test_slack_timeout_is_added_in_slack_config_for_supported_versions.golden",
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
			golden: "test_inhibit_rules_happy_path.golden",
		},
		{
			name:           "opsgenie_api_key_file config",
			againstVersion: versionOpsGenieAPIKeyFileAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					OpsGenieAPIKeyFile: "/test",
				},
			},
			golden: "opsgenie_api_key_file_config.golden",
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
			golden: "api_key_file_field_for_OpsGenie_config.golden",
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
			golden: "api_key_file_and_api_key_fields_for_OpsGenie_config.golden",
		},
		{
			name:           "opsgenie_api_key_file is dropped for unsupported versions",
			againstVersion: versionOpsGenieAPIKeyFileNotAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					OpsGenieAPIKeyFile: "/test",
				},
			},
			golden: "opsgenie_api_key_file_is_dropped_for_unsupported_versions.golden",
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
			golden: "api_key_file_is_dropped_for_unsupported_versions.golden",
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
			golden: "discord_config_for_supported_versions.golden",
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
			name:           "Test content is dropped in discord config for unsupported versions",
			againstVersion: versionDiscordMessageFieldsNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						DiscordConfigs: []*discordConfig{
							{
								WebhookURL: "http://example.com",
								Content:    "content added for unsupported version",
							},
						},
					},
				},
			},
			golden: "test_content_field_dropped_in_discord_config_for_unsupported_versions.golden",
		},
		{
			name:           "Test content is added in discord config for supported versions",
			againstVersion: versionDiscordMessageFieldsAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						DiscordConfigs: []*discordConfig{
							{
								WebhookURL: "http://example.com",
								Content:    "content added for supported version",
							},
						},
					},
				},
			},
			golden: "test_content_field_added_in_discord_config_for_supported_versions.golden",
		},
		{
			name:           "Test username is dropped in discord config for unsupported versions",
			againstVersion: versionDiscordMessageFieldsNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						DiscordConfigs: []*discordConfig{
							{
								WebhookURL: "http://example.com",
								Username:   "discord_admin",
							},
						},
					},
				},
			},
			golden: "test_username_field_dropped_in_discord_config_for_unsupported_versions.golden",
		},
		{
			name:           "Test username is added in discord config for supported versions",
			againstVersion: versionDiscordMessageFieldsAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						DiscordConfigs: []*discordConfig{
							{
								WebhookURL: "http://example.com",
								Username:   "discord_admin",
							},
						},
					},
				},
			},
			golden: "test_username_field_added_in_discord_config_for_supported_versions.golden",
		},
		{
			name:           "Test avatar_url is dropped in discord config for unsupported versions",
			againstVersion: versionDiscordMessageFieldsNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						DiscordConfigs: []*discordConfig{
							{
								WebhookURL: "http://example.com",
								AvatarURL:  "http://example.com/discord_avatar",
							},
						},
					},
				},
			},
			golden: "test_avatar_url_field_dropped_in_discord_config_for_unsupported_versions.golden",
		},
		{
			name:           "Test avatar_url is added in discord config for supported versions",
			againstVersion: versionDiscordMessageFieldsAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						DiscordConfigs: []*discordConfig{
							{
								WebhookURL: "http://example.com",
								AvatarURL:  "http://example.com/discord_avatar",
							},
						},
					},
				},
			},
			golden: "test_avatar_url_field_added_in_discord_config_for_supported_versions.golden",
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
			golden: "webex_config_for_supported_versions.golden",
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
			name:           "msteamsv2_config for supported versions",
			againstVersion: versionMSteamsV2Allowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						MSTeamsV2Configs: []*msTeamsV2Config{
							{
								WebhookURL: "http://example.com",
							},
						},
					},
				},
			},
			golden: "msteamsv2_config_for_supported_versions.golden",
		},
		{
			name:           "msteamsv2_config returns error for unsupported versions",
			againstVersion: versionMSteamsV2NotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						MSTeamsV2Configs: []*msTeamsV2Config{
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
			name:           "msteamsv2_config no webhook url or webhook url file set",
			againstVersion: versionMSteamsV2Allowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						MSTeamsV2Configs: []*msTeamsV2Config{
							{},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "msteamsv2_config both webhook url and webhook url file set",
			againstVersion: versionMSteamsV2Allowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						MSTeamsV2Configs: []*msTeamsV2Config{
							{
								WebhookURL:     "http://example.com",
								WebhookURLFile: "/var/secrets/webhook-url-file",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "msteamsv2_config with webhook url file set",
			againstVersion: versionMSteamsV2Allowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						MSTeamsV2Configs: []*msTeamsV2Config{
							{
								WebhookURLFile: "/var/secrets/webhook-url-file",
							},
						},
					},
				},
			},
			golden: "msteamsv2_config_with_webhook_config_file_set.golden",
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
			golden: "bot_token_file_field_for_Telegram_config.golden",
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
			golden: "bot_token_file_and_bot_token_fields_for_Telegram_config.golden",
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
			golden: "bot_token specified and bot_token_file_is_dropped_for_unsupported_versions.golden",
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
		{
			name:           "message_thread_id field for Telegram config",
			againstVersion: versionTelegramMessageThreadIDAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "telegram",
						TelegramConfigs: []*telegramConfig{
							{
								ChatID:          12345,
								MessageThreadID: 123,
								BotToken:        "test",
							},
						},
					},
				},
			},
			golden: "message_thread_id_field_for_Telegram_config.golden",
		},
		{
			name:           "message_thread_id is dropped for unsupported versions",
			againstVersion: versionTelegramMessageThreadIDNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "telegram",
						TelegramConfigs: []*telegramConfig{
							{
								ChatID:          12345,
								MessageThreadID: 123,
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "summary is dropped for unsupported versions for MSTeams config",
			againstVersion: versionMSTeamsSummaryNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "msteams",
						MSTeamsConfigs: []*msTeamsConfig{
							{
								WebhookURL: "http://example.com",
								Title:      "test title",
								Summary:    "test summary",
								Text:       "test text",
							},
						},
					},
				},
			},
			golden: "summary_is_dropped_for_unsupported_versions_for_MSTeams_config.golden",
		},
		{
			name:           "summary add in supported versions for MSTeams config",
			againstVersion: versionMSTeamsSummaryAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "msteams",
						MSTeamsConfigs: []*msTeamsConfig{
							{
								WebhookURL: "http://example.com",
								Title:      "test title",
								Summary:    "test summary",
								Text:       "test text",
							},
						},
					},
				},
			},
			golden: "summary_add_in_supported_versions_for_MSTeams_config.golden",
		},
		{
			name:           "Test config version mattermost allowed",
			againstVersion: versionMattermostConfigAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						MattermostConfigs: []*mattermostConfig{
							{
								WebhookURL: "www.test.com",
								Text:       "test text",
							},
						},
					},
				},
			},
			golden: "test_config_version_mattermost_allowed.golden",
		},
		{
			name:           "Test drop config version mattermost not allowed",
			againstVersion: versionMattermostConfigNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						MattermostConfigs: []*mattermostConfig{
							{
								WebhookURL: "www.test.com",
								Text:       "test text",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "Test webhook_url takes precedence in mattermost config",
			againstVersion: versionMattermostConfigAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						MattermostConfigs: []*mattermostConfig{
							{
								WebhookURL:     "www.test.com",
								WebhookURLFile: "/test",
								Text:           "test text",
							},
						},
					},
				},
			},
			golden: "test_webhook_url_takes_precedence_in_mattermost_config.golden",
		},
		{
			name:           "Test timeout is dropped in pagerduty config for unsupported versions",
			againstVersion: versionTimeoutConfigNotAllowed,

			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PagerdutyConfigs: []*pagerdutyConfig{
							{
								Timeout: ptr.To(model.Duration(time.Minute)),
							},
						},
					},
				},
			},
			golden: "test_pagerduty_timeout_is_dropped_in_pagerduty_config_for_unsupported_versions.golden",
		},
		{
			name:           "Test timeout is added in pagerduty config for supported versions",
			againstVersion: versionTimeoutConfigAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PagerdutyConfigs: []*pagerdutyConfig{
							{
								Timeout: ptr.To(model.Duration(time.Minute)),
							},
						},
					},
				},
			},
			golden: "test_pagerduty_timeout_is_added_in_pagerduty_config_for_supported_versions.golden",
		},
		{
			name:           "Test slack_app_token is dropped for unsupported versions",
			againstVersion: versionSlackAppConfigNotAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SlackAppToken: "xoxb-token",
					SlackAppURL: &config.URL{
						URL: &url.URL{
							Scheme: "https",
							Host:   "slack.com",
							Path:   "/api/chat.postMessage",
						},
					},
				},
			},
			golden: "test_slack_app_token_is_dropped_for_unsupported_versions.golden",
		},
		{
			name:           "Test slack_app_url is dropped for unsupported versions",
			againstVersion: versionSlackAppConfigNotAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SlackAppURL: &config.URL{
						URL: &url.URL{
							Scheme: "https",
							Host:   "slack.com",
							Path:   "/api/chat.postMessage",
						},
					},
				},
			},
			golden: "test_slack_app_url_is_dropped_for_unsupported_versions.golden",
		},
		{
			name:           "Test slack_app_token and slack_app_url preserved for supported versions",
			againstVersion: versionSlackAppConfigAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SlackAppToken: "xoxb-token",
					SlackAppURL: &config.URL{
						URL: &url.URL{
							Scheme: "https",
							Host:   "slack.com",
							Path:   "/api/chat.postMessage",
						},
					},
				},
			},
			golden: "test_slack_app_token_and_slack_app_url_preserved_for_supported_versions.golden",
		},
		{
			name:           "Test slack_app_token takes precedence over slack_app_token_file",
			againstVersion: versionSlackAppConfigAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SlackAppToken:     "xoxb-token",
					SlackAppTokenFile: "/var/secrets/token",
					SlackAppURL: &config.URL{
						URL: &url.URL{
							Scheme: "https",
							Host:   "slack.com",
							Path:   "/api/chat.postMessage",
						},
					},
				},
			},
			golden: "test_slack_app_token_takes_precedence_over_slack_app_token_file.golden",
		},
		{
			name:           "Test slack_app_token and slack_api_url both configured with different URLs fails",
			againstVersion: versionSlackAppConfigAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SlackAPIURL: &config.URL{
						URL: &url.URL{
							Scheme: "https",
							Host:   "hooks.slack.com",
							Path:   "/services/XXX",
						},
					},
					SlackAppToken: "xoxb-token",
					SlackAppURL: &config.URL{
						URL: &url.URL{
							Scheme: "https",
							Host:   "slack.com",
							Path:   "/api/chat.postMessage",
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "Test slack_app_token and slack_api_url with same URL is allowed",
			againstVersion: versionSlackAppConfigAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SlackAPIURL: &config.URL{
						URL: &url.URL{
							Scheme: "https",
							Host:   "slack.com",
							Path:   "/api/chat.postMessage",
						},
					},
					SlackAppToken: "xoxb-token",
					SlackAppURL: &config.URL{
						URL: &url.URL{
							Scheme: "https",
							Host:   "slack.com",
							Path:   "/api/chat.postMessage",
						},
					},
				},
			},
			golden: "test_slack_app_token_and_slack_api_url_with_same_url_is_allowed.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			routeCfg, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(routeCfg), tc.golden)
		})
	}
}

func TestHTTPClientConfig(t *testing.T) {
	logger := newNopLogger(t)

	httpConfigV25Allowed := semver.Version{Major: 0, Minor: 25}
	httpConfigV25NotAllowed := semver.Version{Major: 0, Minor: 24}

	versionAuthzAllowed := semver.Version{Major: 0, Minor: 22}
	versionAuthzNotAllowed := semver.Version{Major: 0, Minor: 21}

	httpConfigV26Allowed := semver.Version{Major: 0, Minor: 26}
	httpConfigV26NotAllowed := semver.Version{Major: 0, Minor: 25}

	// test the http config independently since all receivers rely on same behaviour
	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *httpClientConfig
		expectErr      bool
		golden         string
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
			golden:         "test_happy_path.golden",
		},
		{
			name: "HTTP client config fields preserved with v0.25.0",
			in: &httpClientConfig{
				OAuth2: &oauth2{
					ClientID:         "a",
					ClientSecret:     "b",
					ClientSecretFile: "c",
					TokenURL:         "d",
					proxyConfig: proxyConfig{
						ProxyURL: "http://example.com/",
					},
				},
				EnableHTTP2: ptr.To(false),
				TLSConfig: &tlsConfig{
					MinVersion: "TLS12",
					MaxVersion: "TLS12",
				},
			},
			againstVersion: httpConfigV25Allowed,
			golden:         "HTTP_client_config_fields_preserved_with_v0_25_0.golden",
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
			golden:         "HTTP_client_config_with_min_TLS_version_only.golden",
		},
		{
			name: "HTTP client config with max TLS version only",
			in: &httpClientConfig{
				TLSConfig: &tlsConfig{
					MaxVersion: "TLS12",
				},
			},
			againstVersion: httpConfigV25Allowed,
			golden:         "HTTP_client_config_with_max_TLS_version_only.golden",
		},
		{
			name: "HTTP client config TLS min version > max version",
			in: &httpClientConfig{
				OAuth2: &oauth2{
					ClientID:         "a",
					ClientSecret:     "b",
					ClientSecretFile: "c",
					TokenURL:         "d",
					proxyConfig: proxyConfig{
						ProxyURL: "http://example.com/",
					},
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
					proxyConfig: proxyConfig{
						ProxyURL: "http://example.com/",
					},
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
					proxyConfig: proxyConfig{
						ProxyURL: "http://example.com/",
					},
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
					proxyConfig: proxyConfig{
						ProxyURL: "http://example.com/",
					},
				},
				EnableHTTP2: ptr.To(false),
				TLSConfig: &tlsConfig{
					MinVersion: "TLS12",
					MaxVersion: "TLS12",
				},
			},
			againstVersion: httpConfigV25NotAllowed,
			golden:         "test_HTTP_client_config_fields_dropped_before_v0_25_0.golden",
		},
		{
			name: "Test HTTP client config oauth2 proxyConfig fields dropped before v0.25.0",
			in: &httpClientConfig{
				OAuth2: &oauth2{
					ClientID:         "a",
					ClientSecret:     "b",
					ClientSecretFile: "c",
					TokenURL:         "d",
					proxyConfig: proxyConfig{
						ProxyURL:             "http://example.com/",
						NoProxy:              "http://proxy.io/",
						ProxyFromEnvironment: true,
					},
				},
				EnableHTTP2: ptr.To(false),
			},
			againstVersion: httpConfigV25NotAllowed,
			golden:         "test_HTTP_client_config_oauth2_proxyConfig_fields_dropped_before_v0_25_0.golden",
		},
		{
			name: "Test HTTP client config oauth2 proxyConfig fields",
			in: &httpClientConfig{
				OAuth2: &oauth2{
					ClientID:         "a",
					ClientSecret:     "b",
					ClientSecretFile: "c",
					TokenURL:         "d",
					proxyConfig: proxyConfig{
						ProxyURL:             "http://example.com/",
						NoProxy:              "http://proxy.io/",
						ProxyFromEnvironment: true,
					},
				},
			},
			againstVersion: httpConfigV25Allowed,
			golden:         "Test_HTTP_client_config_oauth2_proxyConfig_fields.golden",
		},
		{
			name: "no_proxy and proxy_connect_header fields dropped before v0.26.0",
			in: &httpClientConfig{
				proxyConfig: proxyConfig{
					NoProxy: "example.com",
					ProxyConnectHeader: map[string][]string{
						"X-Foo": {"Bar"},
					},
				},
			},
			againstVersion: httpConfigV26NotAllowed,
			golden:         "no_proxy_and_proxy_connect_header_fields_dropped_before_v0_26_0.golden",
		},
		{
			name: "no_proxy/proxy_connect_header fields preserved after v0.26.0",
			in: &httpClientConfig{
				proxyConfig: proxyConfig{
					ProxyURL: "http://example.com",
					NoProxy:  "svc.cluster.local",
					ProxyConnectHeader: map[string][]string{
						"X-Foo": {"Bar"},
					},
				},
			},
			againstVersion: httpConfigV26Allowed,
			golden:         "no_proxy_proxy_connect_header_fields_preserved_after_v0_26_0.golden",
		},
		{
			name: "proxy_from_environment field dropped before v0.26.0",
			in: &httpClientConfig{
				proxyConfig: proxyConfig{
					ProxyFromEnvironment: true,
				},
			},
			againstVersion: httpConfigV26NotAllowed,
			golden:         "proxy_from_environment_field_dropped_before_v0_26_0.golden",
		},
		{
			name: "proxy_from_environment field preserved after v0.26.0",
			in: &httpClientConfig{
				proxyConfig: proxyConfig{
					ProxyFromEnvironment: true,
				},
			},
			againstVersion: httpConfigV26Allowed,
			golden:         "proxy_from_environment_field_preserved_after_v0_26_0.golden",
		},
		{
			name: "proxy_from_environment and proxy_url configured return an error",
			in: &httpClientConfig{
				proxyConfig: proxyConfig{
					ProxyFromEnvironment: true,
					ProxyURL:             "http://example.com",
				},
			},
			againstVersion: httpConfigV26Allowed,
			expectErr:      true,
		},
		{
			name: "proxy_from_environment and no_proxy configured return an error",
			in: &httpClientConfig{
				proxyConfig: proxyConfig{
					ProxyFromEnvironment: true,
					NoProxy:              "svc.cluster.local",
				},
			},
			againstVersion: httpConfigV26Allowed,
			expectErr:      true,
		},
		{
			name: "no_proxy configured alone returns an error",
			in: &httpClientConfig{
				proxyConfig: proxyConfig{
					NoProxy: "svc.cluster.local",
				},
			},
			againstVersion: httpConfigV26Allowed,
			expectErr:      true,
		},
		{
			name: "proxy_connect_header configured alone returns an error",
			in: &httpClientConfig{
				proxyConfig: proxyConfig{
					ProxyConnectHeader: map[string][]string{
						"X-Foo": {"Bar"},
					},
				},
			},
			againstVersion: httpConfigV26Allowed,
			expectErr:      true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			amConfigs, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigs), tc.golden)
		})
	}
}

func TestTimeInterval(t *testing.T) {
	logger := newNopLogger(t)

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		golden         string
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
			golden: "time_intervals_and_active_time_intervals_in_route_config.golden",
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
			golden: "time_intervals_is_dropped_for_unsupported_versions.golden",
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
			golden: "active_time_intervals_is_dropped_for_unsupported_versions.golden",
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
			golden: "location_is_dropped_for_unsupported_versions.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			require.NoError(t, err)

			amConfigs, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigs), tc.golden)
		})
	}
}
func TestSanitizePushoverReceiverConfig(t *testing.T) {
	logger := newNopLogger(t)

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		expectErr      bool
		golden         string
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
			golden: "test_pushover_user_key_token_takes_precedence_in_pushover_config.golden",
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
		{
			name:           "Test pushover HTML and Monospace are mutually exclusive",
			againstVersion: semver.Version{Major: 0, Minor: 29},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey:   "foo",
								Token:     "bar",
								HTML:      ptr.To(true),
								Monospace: ptr.To(true),
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "Test pushover HTML allowed without Monospace",
			againstVersion: semver.Version{Major: 0, Minor: 29},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey: "foo",
								Token:   "bar",
								HTML:    ptr.To(true),
							},
						},
					},
				},
			},
			golden: "test_pushover_html_allowed_without_monospace.golden",
		},
		{
			name:           "Test pushover Monospace allowed without HTML",
			againstVersion: semver.Version{Major: 0, Minor: 29},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey:   "foo",
								Token:     "bar",
								Monospace: ptr.To(true),
							},
						},
					},
				},
			},
			golden: "test_pushover_monospace_allowed_without_html.golden",
		},
		{
			name:           "Test pushover Monospace dropped in unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 28},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey:   "foo",
								Token:     "bar",
								Monospace: ptr.To(true),
							},
						},
					},
				},
			},
			golden: "test_pushover_monospace_dropped_unsupported_version.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			amConfigs, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigs), tc.golden)
		})
	}
}
func TestSanitizeEmailConfig(t *testing.T) {
	logger := newNopLogger(t)

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		golden         string
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
			golden: "test_smtp_auth_password_takes_precedence_in_global_config.golden",
		},
		{
			name:           "Test smtp_auth_password_file is dropped for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 24},
			in: &alertmanagerConfig{
				Global: &globalConfig{
					SMTPAuthPasswordFile: "bar",
				},
			},
			golden: "test_smtp_auth_password_file_is_dropped_for_unsupported_versions.golden",
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
			golden: "test_smtp_auth_password_takes_precedence_in_email_config.golden",
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
			golden: "test_smtp_auth_password_file_is_dropped_in_slack_config_for_unsupported_versions.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			require.NoError(t, err)

			amConfigs, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigs), tc.golden)
		})
	}
}

func TestSanitizeVictorOpsConfig(t *testing.T) {
	logger := newNopLogger(t)

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		golden         string
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
			golden: "test_victorops_api_key_takes_precedence_in_global_config.golden",
		},
		{
			name:           "Test victorops_api_key_file is dropped for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 24},
			in: &alertmanagerConfig{
				Global: &globalConfig{
					VictorOpsAPIKeyFile: "bar",
				},
			},
			golden: "test_victorops_api_key_file_is_dropped_for_unsupported_versions.golden",
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
			golden: "test_api_key_takes_precedence_in_victorops_config.golden",
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
			golden: "test_api_key_file_is_dropped_in_victorops_config_for_unsupported_versions.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			require.NoError(t, err)

			amConfigs, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigs), tc.golden)
		})
	}
}

func TestSanitizeWebhookConfig(t *testing.T) {
	logger := newNopLogger(t)

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		golden         string
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
			golden: "test_webhook_url_file_is_dropped_in_webhook_config_for_unsupported_versions.golden",
		},
		{
			name:           "Test url takes precedence in webhook config",
			againstVersion: semver.Version{Major: 0, Minor: 26},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						WebhookConfigs: []*webhookConfig{
							{
								URL:     "http://example.com/foo",
								URLFile: "bar",
							},
						},
					},
				},
			},
			golden: "test_url_takes_precedence_in_webhook_config.golden",
		},
		{
			name:           "Test timeout is dropped in webhook config for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 26},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						WebhookConfigs: []*webhookConfig{
							{
								Timeout: ptr.To(model.Duration(time.Minute)),
							},
						},
					},
				},
			},
			golden: "test_webhook_timeout_is_dropped_in_webhook_config_for_unsupported_versions.golden",
		},
		{
			name:           "Test timeout is added in webhook config for supported versions",
			againstVersion: semver.Version{Major: 0, Minor: 28},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						WebhookConfigs: []*webhookConfig{
							{
								Timeout: ptr.To(model.Duration(time.Minute)),
							},
						},
					},
				},
			},
			golden: "test_webhook_timeout_is_added_in_webhook_config_for_supported_versions.golden",
		},
		{
			name:           "Test invalid url returns error",
			againstVersion: semver.Version{Major: 0, Minor: 26},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						WebhookConfigs: []*webhookConfig{
							{
								URL: "not-a-valid-url",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "Test valid url passes validation",
			againstVersion: semver.Version{Major: 0, Minor: 26},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						WebhookConfigs: []*webhookConfig{
							{
								URL: "http://example.com/webhook",
							},
						},
					},
				},
			},
			golden: "test_webhook_valid_url_passes.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			amConfigs, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigs), tc.golden)
		})
	}
}

func TestSanitizePushoverConfig(t *testing.T) {
	logger := newNopLogger(t)

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		golden         string
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
			golden: "test_pushover_user_key_file_is_dropped_in_pushover_config_for_unsupported_versions.golden",
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
			golden: "test_pushover_token_file_is_dropped_in_pushover_config_for_unsupported_versions.golden",
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
			golden: "test_user_key_takes_precedence_in_pushover_config.golden",
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
			golden: "test_token_takes_precedence_in_pushover_config.golden",
		},
		{
			name:           "Test invalid url returns error",
			againstVersion: semver.Version{Major: 0, Minor: 26},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey: "key",
								Token:   "token",
								URL:     "not-a-valid-url",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "Test valid url passes validation",
			againstVersion: semver.Version{Major: 0, Minor: 26},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						PushoverConfigs: []*pushoverConfig{
							{
								UserKey: "key",
								Token:   "token",
								URL:     "http://example.com",
							},
						},
					},
				},
			},
			golden: "test_pushover_valid_url_passes.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			amConfigs, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigs), tc.golden)
		})
	}
}

func TestSanitizePagerDutyConfig(t *testing.T) {
	logger := newNopLogger(t)

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		golden         string
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
			golden: "test_routing_key_takes_precedence_in_pagerdouty_config.golden",
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
			golden: "test_routing_key_file_is_dropped_in_pagerduty_config_for_unsupported_versions.golden",
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
			golden: "test_service_key_takes_precedence_in_pagerduty_config.golden",
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
			golden: "test_service_key_file_is_dropped_in_pagerduty_config_for_unsupported_versions.golden",
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
			golden: "test_source_is_dropped_in_pagerduty_config_for_unsupported_versions.golden",
		},
		{
			name:           "Test source is added in pagerduty config for supported versions",
			againstVersion: semver.Version{Major: 0, Minor: 25},
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
			golden: "test_source_is_added_in_pagerduty_config_for_supported_versions.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			require.NoError(t, err)

			amPagerDutyCfg, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amPagerDutyCfg), tc.golden)
		})
	}
}

func TestSanitizeJiraConfig(t *testing.T) {
	logger := newNopLogger(t)
	versionJiraAllowed := semver.Version{Major: 0, Minor: 28}
	versionJiraNotAllowed := semver.Version{Major: 0, Minor: 27}

	versionAPITypeAllowed := semver.Version{Major: 0, Minor: 29}
	versionAPITypeNotAllowed := semver.Version{Major: 0, Minor: 28}
	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		golden         string
		expectErr      bool
	}{
		{
			name:           "jira_configs returns error for unsupported versions",
			againstVersion: versionJiraNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						JiraConfigs: []*jiraConfig{
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
			name:           "jira_configs allows for supported versions",
			againstVersion: versionJiraAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						JiraConfigs: []*jiraConfig{
							{
								APIURL:    "http://issues.example.com",
								Project:   "Monitoring",
								IssueType: "Bug",
							},
						},
					},
				},
			},
			golden: "jira_configs_for_supported_versions.golden",
		},
		{
			name:           "jira_configs returns error for missing mandatory fields",
			againstVersion: versionJiraAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						JiraConfigs: []*jiraConfig{
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
			name:           "jira_configs with send_resolved",
			againstVersion: versionJiraAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						JiraConfigs: []*jiraConfig{
							{
								APIURL:       "http://issues.example.com",
								Project:      "Monitoring",
								IssueType:    "Bug",
								SendResolved: ptr.To(true),
							},
						},
					},
				},
			},
			golden: "jira_configs_with_send_resolved.golden",
		},
		{
			name:           "jira_configs with api_type",
			againstVersion: versionAPITypeAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						JiraConfigs: []*jiraConfig{
							{
								APIURL:    "http://issues.example.com",
								Project:   "Monitoring",
								APIType:   "datacenter",
								IssueType: "Bug",
							},
						},
					},
				},
			},
			golden: "jira_config_with_api_type.golden",
		},
		{
			name:           "jira_configs with api_type version not supported",
			againstVersion: versionAPITypeNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						JiraConfigs: []*jiraConfig{
							{
								APIURL:    "http://issues.example.com",
								Project:   "Monitoring",
								APIType:   "datacenter",
								IssueType: "Bug",
							},
						},
					},
				},
			},
			golden: "jira_config_with_api_type_version_not_supported.golden",
		},
		{
			name:           "jira_configs with invalid api_type",
			againstVersion: versionAPITypeAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						JiraConfigs: []*jiraConfig{
							{
								APIURL:    "http://issues.example.com",
								Project:   "Monitoring",
								APIType:   "onpremise",
								IssueType: "Bug",
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

			amConfigs, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigs), tc.golden)
		})
	}
}

func TestSanitizeDiscordConfig(t *testing.T) {
	logger := newNopLogger(t)

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		golden         string
	}{
		{
			name:           "Test Username field is dropped in discord config for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 27},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "discord",
						DiscordConfigs: []*discordConfig{
							{
								Username:   "content",
								WebhookURL: "http://example.com",
								Message:    "test message",
							},
						},
					},
				},
			},
			golden: "Discord_username_dropped_in_unsupported_versions_config.golden",
		},
		{
			name:           "Test Username field add in discord config for supported versions",
			againstVersion: semver.Version{Major: 0, Minor: 28},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "discord",
						DiscordConfigs: []*discordConfig{
							{
								Username:   "content",
								WebhookURL: "http://example.com",
								Message:    "test message",
							},
						},
					},
				},
			},
			golden: "Discord_username_add_in_supported_versions_config.golden",
		},
		{
			name:           "Test AvatarURL field is dropped in discord config for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 27},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "discord",
						DiscordConfigs: []*discordConfig{
							{
								AvatarURL:  "content",
								WebhookURL: "http://example.com",
								Message:    "test message",
							},
						},
					},
				},
			},
			golden: "Discord_avatarURL_dropped_in_unsupported_versions_config.golden",
		},
		{
			name:           "Test AvatarURL field add in discord config for supported versions",
			againstVersion: semver.Version{Major: 0, Minor: 28},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "discord",
						DiscordConfigs: []*discordConfig{
							{
								AvatarURL:  "content",
								WebhookURL: "http://example.com",
								Message:    "test message",
							},
						},
					},
				},
			},
			golden: "Discord_avatarURL_add_in_supported_versions_config.golden",
		},
		{
			name:           "Test Content field is dropped in discord config for unsupported versions",
			againstVersion: semver.Version{Major: 0, Minor: 27},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "discord",
						DiscordConfigs: []*discordConfig{
							{
								Content:    "content",
								WebhookURL: "http://example.com",
								Message:    "test message",
							},
						},
					},
				},
			},
			golden: "Discord_content_dropped_in_unsupported_versions_config.golden",
		},
		{
			name:           "Test Content field add in discord config for supported versions",
			againstVersion: semver.Version{Major: 0, Minor: 28},
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						Name: "discord",
						DiscordConfigs: []*discordConfig{
							{
								Content:    "test content",
								WebhookURL: "http://example.com",
								Message:    "test message",
							},
						},
					},
				},
			},
			golden: "Discord_content_add_in_supported_versions_config.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			require.NoError(t, err)

			amConfigs, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigs), tc.golden)
		})
	}

}

func TestSanitizeRocketChatConfig(t *testing.T) {
	logger := newNopLogger(t)
	versionRocketChatAllowed := semver.Version{Major: 0, Minor: 28}
	versionRocketChatNotAllowed := semver.Version{Major: 0, Minor: 27}
	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		golden         string
		expectErr      bool
	}{
		{
			name:           "rocketchat_configs returns error for unsupported versions",
			againstVersion: versionRocketChatNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						RocketChatConfigs: []*rocketChatConfig{
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
			name:           "rocketchat_config with send_resolved",
			againstVersion: versionRocketChatAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						RocketChatConfigs: []*rocketChatConfig{
							{
								APIURL:       "http://example.com",
								SendResolved: ptr.To(true),
							},
						},
					},
				},
			},
			golden: "rocketchat_config_with_send_resolved.golden",
		},
		{
			name:           "rocketchat_configs allows for supported versions",
			againstVersion: versionRocketChatAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						RocketChatConfigs: []*rocketChatConfig{
							{
								APIURL: "http://example.com",
							},
						},
					},
				},
			},
			golden: "rocketchat_configs_for_supported_versions.golden",
		},
		{
			name:           "rocketchat_configs both token or token_file set",
			againstVersion: versionRocketChatAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						RocketChatConfigs: []*rocketChatConfig{
							{
								APIURL:    "http://example.com",
								Token:     ptr.To("aaaa-bbbb-cccc-dddd"),
								TokenFile: "/var/kubernetes/secrets/token",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "rocketchat_configs both token_id or token_id_file set",
			againstVersion: versionRocketChatAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						RocketChatConfigs: []*rocketChatConfig{
							{
								APIURL:      "http://example.com",
								TokenID:     ptr.To("t123456"),
								TokenIDFile: "/var/kubernetes/secrets/token-id",
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

			amConfigs, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigs), tc.golden)
		})
	}
}

func TestSanitizeIncidentioConfig(t *testing.T) {
	logger := newNopLogger(t)
	versionIncidentioAllowed := semver.Version{Major: 0, Minor: 29}
	versionIncidentioNotAllowed := semver.Version{Major: 0, Minor: 28}
	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		golden         string
		expectErr      bool
	}{
		{
			name:           "incidentio_configs returns error for unsupported versions",
			againstVersion: versionIncidentioNotAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						IncidentioConfigs: []*incidentioConfig{
							{
								URL:              "http://example.com",
								AlertSourceToken: "token123",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "incidentio_configs allows for supported versions",
			againstVersion: versionIncidentioAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						IncidentioConfigs: []*incidentioConfig{
							{
								URL:              "http://example.com",
								AlertSourceToken: "token123",
							},
						},
					},
				},
			},
			golden: "incidentio_configs_for_supported_versions.golden",
		},
		{
			name:           "incidentio_configs invalid url returns error",
			againstVersion: versionIncidentioAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						IncidentioConfigs: []*incidentioConfig{
							{
								URL:              "not-a-valid-url",
								AlertSourceToken: "token123",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "incidentio_configs both url and url_file set",
			againstVersion: versionIncidentioAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						IncidentioConfigs: []*incidentioConfig{
							{
								URL:              "http://example.com",
								URLFile:          "/var/kubernetes/secrets/url",
								AlertSourceToken: "token123",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "incidentio_configs both alert_source_token and alert_source_token_file set",
			againstVersion: versionIncidentioAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						IncidentioConfigs: []*incidentioConfig{
							{
								URL:                  "http://example.com",
								AlertSourceToken:     "token123",
								AlertSourceTokenFile: "/var/kubernetes/secrets/token",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "incidentio_configs missing url and url_file",
			againstVersion: versionIncidentioAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						IncidentioConfigs: []*incidentioConfig{
							{
								AlertSourceToken: "token123",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "incidentio_configs missing authentication",
			againstVersion: versionIncidentioAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						IncidentioConfigs: []*incidentioConfig{
							{
								URL: "http://example.com",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "incidentio_configs with http_config.authorization and alert_source_token",
			againstVersion: versionIncidentioAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						IncidentioConfigs: []*incidentioConfig{
							{
								URL:              "http://example.com",
								AlertSourceToken: "token123",
								HTTPConfig: &httpClientConfig{
									Authorization: &authorization{
										Type:        "Bearer",
										Credentials: "creds123",
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
			name:           "incidentio_configs with http_config.authorization only",
			againstVersion: versionIncidentioAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						IncidentioConfigs: []*incidentioConfig{
							{
								URL: "http://example.com",
								HTTPConfig: &httpClientConfig{
									Authorization: &authorization{
										Type:        "Bearer",
										Credentials: "creds123",
									},
								},
							},
						},
					},
				},
			},
			golden: "incidentio_configs_with_http_authorization.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			amConfigs, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigs), tc.golden)
		})
	}
}
func TestSanitizeRoute(t *testing.T) {
	logger := newNopLogger(t)
	matcherV2SyntaxAllowed := semver.Version{Major: 0, Minor: 22}
	matcherV2SyntaxNotAllowed := semver.Version{Major: 0, Minor: 21}

	namespaceLabel := "namespace"
	namespaceValue := "test-ns"

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *route
		expectErr      bool
		golden         string
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
			golden: "test_route_with_new_syntax_no_child_routes.golden",
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
			golden: "test_route_with_new_syntax_supported_with_child_routes.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			routeCfg, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(routeCfg), tc.golden)
		})
	}
}

// We want to ensure that the imported types from config.MuteTimeInterval
// and any others with custom marshalling/unmarshalling are parsed
// into the internal struct as expected.
func TestLoadConfig(t *testing.T) {
	testCase := []struct {
		name     string
		expected *alertmanagerConfig
		golden   string
	}{
		{
			name: "mute_time_intervals field",
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
			golden: "mute_time_intervals_field.golden",
		},
		{
			name: "Global opsgenie_api_key_file field",
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
			golden: "Global_opsgenie_api_key_file_field.golden",
		},
		{
			name: "OpsGenie entity and actions fields",
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
			golden: "OpsGenie_entity_and_actions_fields.golden",
		},
		{
			name: "Discord url field",
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
			golden: "Discord_url_field.golden",
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ac, err := alertmanagerConfigFromBytes(golden.Get(t, tc.golden))
			require.NoError(t, err)
			require.Equal(t, tc.expected, ac)
		})
	}
}

func TestConvertHTTPConfig(t *testing.T) {
	testCases := []struct {
		name    string
		cfg     monitoringv1alpha1.HTTPConfig
		version string
		golden  string
	}{
		{
			name:   "no proxy",
			cfg:    monitoringv1alpha1.HTTPConfig{},
			golden: "no_proxy.golden",
		},
		{
			name: "proxyURL only",
			cfg: monitoringv1alpha1.HTTPConfig{
				ProxyURLOriginal: ptr.To("http://example.com"),
			},
			golden: "proxy_url_only.golden",
		},
		{
			name: "proxyUrl only",
			cfg: monitoringv1alpha1.HTTPConfig{
				ProxyConfig: monitoringv1.ProxyConfig{
					ProxyURL: ptr.To("http://example.com"),
				},
			},
			golden: "proxy_config_only.golden",
		},
		{
			name: "proxyUrl and proxyURL",
			cfg: monitoringv1alpha1.HTTPConfig{
				ProxyURLOriginal: ptr.To("http://example.com"),
				ProxyConfig: monitoringv1.ProxyConfig{
					ProxyURL: ptr.To("http://bad.example.com"),
				},
			},
			golden: "proxy_url_and_proxy_config.golden",
		},
		{
			name: "proxyUrl and empty proxyURL",
			cfg: monitoringv1alpha1.HTTPConfig{
				ProxyURLOriginal: ptr.To(""),
				ProxyConfig: monitoringv1.ProxyConfig{
					ProxyURL: ptr.To("http://example.com"),
				},
			},
			golden: "proxy_url_empty_proxy_config.golden",
		},
		{
			name: "enableHTTP2",
			cfg: monitoringv1alpha1.HTTPConfig{
				EnableHTTP2: ptr.To(false),
			},
			golden: "http_config_enable_http2_supported.golden",
		},
		{
			name: "enableHTTP2 not supported",
			cfg: monitoringv1alpha1.HTTPConfig{
				EnableHTTP2: ptr.To(false),
			},
			version: "v0.24.0",
			golden:  "http_config_enable_http2_not_supported.golden",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			version := operator.DefaultAlertmanagerVersion
			if tc.version != "" {
				version = tc.version
			}
			v, err := semver.ParseTolerant(version)
			require.NoError(t, err)

			logger := newNopLogger(t)
			cb := NewConfigBuilder(
				logger,
				v,
				nil,
				&monitoringv1.Alertmanager{
					ObjectMeta: metav1.ObjectMeta{Namespace: "alertmanager-namespace"},
					Spec: monitoringv1.AlertmanagerSpec{
						Version: tc.version,
					},
				},
			)

			cfg, err := cb.convertHTTPConfig(context.Background(), &tc.cfg, types.NamespacedName{})
			require.NoError(t, err)

			err = cfg.sanitize(v, logger)
			require.NoError(t, err)

			cfgBytes, err := yaml.Marshal(cfg)
			require.NoError(t, err)

			golden.Assert(t, string(cfgBytes), tc.golden)
		})
	}
}

func TestSanitizeTelegramConfig(t *testing.T) {
	logger := newNopLogger(t)
	versionTelegramExampleAllowed := semver.Version{Major: 0, Minor: 26}

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		golden         string
		expectErr      bool
	}{
		{
			name:           "telegram invalid api_url returns error",
			againstVersion: versionTelegramExampleAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						TelegramConfigs: []*telegramConfig{
							{
								APIUrl:   "not-a-valid-url",
								BotToken: "token",
								ChatID:   12345,
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "telegram valid api_url passes validation",
			againstVersion: versionTelegramExampleAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						TelegramConfigs: []*telegramConfig{
							{
								APIUrl:   "http://example.com",
								BotToken: "token",
								ChatID:   12345,
							},
						},
					},
				},
			},
			golden: "telegram_valid_url_passes.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			amConfigs, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigs), tc.golden)
		})
	}
}

func TestSanitizeMSTeamsConfig(t *testing.T) {
	logger := newNopLogger(t)
	versionMSTeamsExampleAllowed := semver.Version{Major: 0, Minor: 27}

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		golden         string
		expectErr      bool
	}{
		{
			name:           "msteams invalid webhook_url returns error",
			againstVersion: versionMSTeamsExampleAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						MSTeamsConfigs: []*msTeamsConfig{
							{
								WebhookURL: "not-a-valid-url",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "msteams valid webhook_url passes validation",
			againstVersion: versionMSTeamsExampleAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						MSTeamsConfigs: []*msTeamsConfig{
							{
								WebhookURL: "http://example.com/webhook",
							},
						},
					},
				},
			},
			golden: "msteams_valid_url_passes.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			amConfigs, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigs), tc.golden)
		})
	}
}

func TestSanitizeSNSConfig(t *testing.T) {
	logger := newNopLogger(t)
	versionSNSAllowed := semver.Version{Major: 0, Minor: 25}

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		golden         string
		expectErr      bool
	}{
		{
			name:           "sns invalid api_url returns error",
			againstVersion: versionSNSAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						SNSConfigs: []*snsConfig{
							{
								APIUrl:   "not-a-valid-url",
								TopicARN: "arn:aws:sns:us-east-1:123456789012:test",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:           "sns valid api_url passes validation",
			againstVersion: versionSNSAllowed,
			in: &alertmanagerConfig{
				Receivers: []*receiver{
					{
						SNSConfigs: []*snsConfig{
							{
								APIUrl:   "https://sns.us-east-1.amazonaws.com",
								TopicARN: "arn:aws:sns:us-east-1:123456789012:test",
							},
						},
					},
				},
			},
			golden: "sns_valid_url_passes.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.sanitize(tc.againstVersion, logger)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			amConfigs, err := yaml.Marshal(tc.in)
			require.NoError(t, err)

			golden.Assert(t, string(amConfigs), tc.golden)
		})
	}
}

func newNopLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.DiscardHandler)
}
