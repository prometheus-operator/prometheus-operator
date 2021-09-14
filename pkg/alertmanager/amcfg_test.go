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
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

func TestGenerateConfig(t *testing.T) {
	type testCase struct {
		name       string
		kclient    kubernetes.Interface
		baseConfig alertmanagerConfig
		amConfigs  map[string]*monitoringv1alpha1.AlertmanagerConfig
		expected   string
	}

	ctx := context.Background()
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
			name:    "base with sub route, no CRs",
			kclient: fake.NewSimpleClientset(),
			baseConfig: alertmanagerConfig{
				Route: &route{
					Receiver: "null",
					Routes: []*route{{
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
receivers:
- name: "null"
- name: custom
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
						},
						Receivers: []monitoringv1alpha1.Receiver{{Name: "test"}},
					},
				},
			},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: mynamespace-myamc-test
    match:
      namespace: mynamespace
    continue: true
receivers:
- name: "null"
- name: mynamespace-myamc-test
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
			name:    "base with subroute, simple CR",
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
  - receiver: mynamespace-myamc-test
    match:
      namespace: mynamespace
    continue: true
  - receiver: "null"
receivers:
- name: "null"
- name: mynamespace-myamc-test
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
							}},
						}},
					},
				},
			},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: mynamespace-myamc-test-pd
    match:
      namespace: mynamespace
    continue: true
receivers:
- name: "null"
- name: mynamespace-myamc-test-pd
  pagerduty_configs:
  - routing_key: 1234abc
templates: []
`,
		},
		{
			name:    "CR with Webhook Receiver",
			kclient: fake.NewSimpleClientset(),
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
							}},
						}},
					},
				},
			},
			expected: `route:
  receiver: "null"
  routes:
  - receiver: mynamespace-myamc-test
    match:
      namespace: mynamespace
    continue: true
receivers:
- name: "null"
- name: mynamespace-myamc-test
  webhook_configs:
  - url: http://test.url
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
  - receiver: mynamespace-myamc-test
    match:
      namespace: mynamespace
    continue: true
receivers:
- name: "null"
- name: mynamespace-myamc-test
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
  - receiver: mynamespace-myamc-test
    match:
      namespace: mynamespace
    continue: true
receivers:
- name: "null"
- name: mynamespace-myamc-test
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
  - receiver: mynamespace-myamc-test
    match:
      namespace: mynamespace
    continue: true
receivers:
- name: "null"
- name: mynamespace-myamc-test
  wechat_configs:
  - api_secret: wechatsecret
    corp_id: wechatcorpid
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
  - receiver: mynamespace-myamc-test
    match:
      namespace: mynamespace
    continue: true
receivers:
- name: "null"
- name: mynamespace-myamc-test
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
  - receiver: mynamespace-myamc-test
    match:
      namespace: mynamespace
    continue: true
receivers:
- name: "null"
- name: mynamespace-myamc-test
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
	}

	version, err := semver.ParseTolerant("v0.22.2")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := assets.NewStore(tc.kclient.CoreV1(), tc.kclient.CoreV1())
			cg := newConfigGenerator(nil, version, store)
			cfgBytes, err := cg.generateConfig(ctx, tc.baseConfig, tc.amConfigs)
			if err != nil {
				t.Fatal(err)
			}

			result := string(cfgBytes)

			// Verify the generated yaml is as expected
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("Unexpected result (-want +got):\n%s", diff)
			}

			// Verify the generated config is something that Alertmanager will be happy with
			_, err = config.Load(result)
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

	versionAuthzAllowed := versionFileURLAllowed
	versionAuthzNotAllowed := versionFileURLNotAllowed

	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *alertmanagerConfig
		expect         alertmanagerConfig
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
			name:           "Test basicAuth takes precedence over authorization in http config",
			againstVersion: versionFileURLNotAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					HTTPConfig: &httpClientConfig{
						Authorization: &authorization{
							Type:            "any",
							Credentials:     "some",
							CredentialsFile: "/must/drop",
						},
						BasicAuth: &basicAuth{
							Username:     "tester",
							Password:     "testing",
							PasswordFile: "/test",
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Global: &globalConfig{
					HTTPConfig: &httpClientConfig{
						Authorization: nil,
						BasicAuth: &basicAuth{
							Username:     "tester",
							Password:     "testing",
							PasswordFile: "/test",
						},
					},
				},
			},
		},
		{
			name:           "Test authorization is dropped in global http config for unsupported versions",
			againstVersion: versionAuthzNotAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					HTTPConfig: &httpClientConfig{
						Authorization: &authorization{
							Type:            "any",
							Credentials:     "some",
							CredentialsFile: "/must/drop",
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Global: &globalConfig{
					HTTPConfig: &httpClientConfig{
						Authorization: nil,
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
			name:           "Test http config happy path",
			againstVersion: versionAuthzAllowed,
			in: &alertmanagerConfig{
				Global: &globalConfig{
					HTTPConfig: &httpClientConfig{
						Authorization: &authorization{
							Type:            "any",
							Credentials:     "some",
							CredentialsFile: "/must/keep",
						},
					},
				},
			},
			expect: alertmanagerConfig{
				Global: &globalConfig{
					HTTPConfig: &httpClientConfig{
						Authorization: &authorization{
							Type:            "any",
							Credentials:     "some",
							CredentialsFile: "/must/keep",
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.in.sanitize(tc.againstVersion, logger)
			out := *tc.in
			if !reflect.DeepEqual(out, tc.expect) {
				t.Fatalf("wanted %v but got %v", tc.expect, out)
			}
		})
	}

	// test the http config independently since all receivers rely on same behaviour
	for _, tc := range []struct {
		name           string
		againstVersion semver.Version
		in             *httpClientConfig
		expect         httpClientConfig
	}{
		{
			name: "Test authorization is dropped in http config for unsupported versions",
			in: &httpClientConfig{
				Authorization: &authorization{
					Type:            "any",
					Credentials:     "some",
					CredentialsFile: "/must/drop",
				},
			},
			againstVersion: versionAuthzNotAllowed,
			expect:         httpClientConfig{},
		},
		{
			name: "Test authorization is dropped in favour of basicAuth for http config",
			in: &httpClientConfig{
				Authorization: &authorization{
					Type:            "any",
					Credentials:     "some",
					CredentialsFile: "/must/drop",
				},
				BasicAuth: &basicAuth{
					Username:     "tester",
					Password:     "testing",
					PasswordFile: "/test",
				},
			},
			againstVersion: versionAuthzNotAllowed,
			expect: httpClientConfig{
				BasicAuth: &basicAuth{
					Username:     "tester",
					Password:     "testing",
					PasswordFile: "/test",
				},
			},
		},
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.in.sanitize(tc.againstVersion, logger)
			out := *tc.in
			if !reflect.DeepEqual(out, tc.expect) {
				t.Fatalf("wanted %v but got %v", tc.expect, out)
			}
		})
	}

}
