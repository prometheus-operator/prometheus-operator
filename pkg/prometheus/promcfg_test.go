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
	"bytes"
	"fmt"
	"testing"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/kylelemons/godebug/pretty"
)

func TestConfigGeneration(t *testing.T) {
	for _, v := range CompatibilityMatrix {
		cfg, err := generateTestConfig(v)
		if err != nil {
			t.Fatal(err)
		}

		for i := 0; i < 1000; i++ {
			testcfg, err := generateTestConfig(v)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(cfg, testcfg) {
				t.Fatalf("Config generation is not deterministic.\n\n\nFirst generation: \n\n%s\n\nDifferent generation: \n\n%s\n\n", string(cfg), string(testcfg))
			}
		}
	}
}

func TestNamespaceSetCorrectly(t *testing.T) {
	sm := &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testservicemonitor1",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group1",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{"test"},
			},
		},
	}

	cg := &configGenerator{}

	c := cg.generateK8SSDConfig(getNamespacesFromServiceMonitor(sm), nil, nil)
	s, err := yaml.Marshal(yaml.MapSlice{c})
	if err != nil {
		t.Fatal(err)
	}

	expected := `kubernetes_sd_configs:
- role: endpoints
  namespaces:
    names:
    - test
`

	result := string(s)

	if expected != result {
		t.Fatalf("Unexpected result.\n\nGot:\n\n%s\n\nExpected:\n\n%s\n\n", result, expected)
	}
}

func TestK8SSDConfigGeneration(t *testing.T) {
	sm := &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testservicemonitor1",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group1",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{"test"},
			},
		},
	}

	cg := &configGenerator{}

	testcases := []struct {
		apiserverConfig  *monitoringv1.APIServerConfig
		basicAuthSecrets map[string]BasicAuthCredentials
		expected         string
	}{
		{
			nil,
			nil,
			`kubernetes_sd_configs:
- role: endpoints
  namespaces:
    names:
    - test
`,
		},
		{
			&monitoringv1.APIServerConfig{
				Host:            "example.com",
				BasicAuth:       &monitoringv1.BasicAuth{},
				BearerToken:     "bearer_token",
				BearerTokenFile: "bearer_token_file",
				TLSConfig:       nil,
			},
			map[string]BasicAuthCredentials{
				"apiserver": {
					"foo",
					"bar",
				},
			},
			`kubernetes_sd_configs:
- role: endpoints
  namespaces:
    names:
    - test
  api_server: example.com
  basic_auth:
    username: foo
    password: bar
  bearer_token: bearer_token
  bearer_token_file: bearer_token_file
`,
		},
	}

	for _, tc := range testcases {
		c := cg.generateK8SSDConfig(
			getNamespacesFromServiceMonitor(sm),
			tc.apiserverConfig,
			tc.basicAuthSecrets,
		)
		s, err := yaml.Marshal(yaml.MapSlice{c})
		if err != nil {
			t.Fatal(err)
		}
		result := string(s)

		if result != tc.expected {
			t.Fatalf("Unexpected result.\n\nGot:\n\n%s\n\nExpected:\n\n%s\n\n", result, tc.expected)
		}
	}
}

func TestAlertmanagerBearerToken(t *testing.T) {
	cg := &configGenerator{}
	cfg, err := cg.generateConfig(
		&monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: monitoringv1.PrometheusSpec{
				Alerting: &monitoringv1.AlertingSpec{
					Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
						{
							Name:            "alertmanager-main",
							Namespace:       "default",
							Port:            intstr.FromString("web"),
							BearerTokenFile: "/some/file/on/disk",
						},
					},
				},
			},
		},
		nil,
		map[string]BasicAuthCredentials{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	// If this becomes an endless sink of maintenance, then we should just
	// change this to check that just the `bearer_token_file` is set with
	// something like json-path.
	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
rule_files: []
scrape_configs: []
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  alertmanagers:
  - path_prefix: /
    scheme: http
    kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
        - default
    bearer_token_file: /some/file/on/disk
    relabel_configs:
    - action: keep
      source_labels:
      - __meta_kubernetes_service_name
      regex: alertmanager-main
    - action: keep
      source_labels:
      - __meta_kubernetes_endpoint_port_name
      regex: web
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestAdditionalAlertRelabelConfigs(t *testing.T) {
	cg := &configGenerator{}
	cfg, err := cg.generateConfig(
		&monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: monitoringv1.PrometheusSpec{
				Alerting: &monitoringv1.AlertingSpec{
					Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
						{
							Name:      "alertmanager-main",
							Namespace: "default",
							Port:      intstr.FromString("web"),
						},
					},
				},
			},
		},
		nil,
		map[string]BasicAuthCredentials{},
		nil,
		[]byte(`- action: drop
  source_labels: [__meta_kubernetes_node_name]
  regex: spot-(.+)

`),
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
rule_files: []
scrape_configs: []
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  - action: drop
    source_labels:
    - __meta_kubernetes_node_name
    regex: spot-(.+)
  alertmanagers:
  - path_prefix: /
    scheme: http
    kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
        - default
    relabel_configs:
    - action: keep
      source_labels:
      - __meta_kubernetes_service_name
      regex: alertmanager-main
    - action: keep
      source_labels:
      - __meta_kubernetes_endpoint_port_name
      regex: web
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestAdditionalAlertmanagers(t *testing.T) {
	cg := &configGenerator{}
	cfg, err := cg.generateConfig(
		&monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: monitoringv1.PrometheusSpec{
				Alerting: &monitoringv1.AlertingSpec{
					Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
						{
							Name:      "alertmanager-main",
							Namespace: "default",
							Port:      intstr.FromString("web"),
						},
					},
				},
			},
		},
		nil,
		map[string]BasicAuthCredentials{},
		nil,
		nil,
		[]byte(`- static_configs:
  - targets:
    - localhost
`),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
rule_files: []
scrape_configs: []
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  alertmanagers:
  - path_prefix: /
    scheme: http
    kubernetes_sd_configs:
    - role: endpoints
      namespaces:
        names:
        - default
    relabel_configs:
    - action: keep
      source_labels:
      - __meta_kubernetes_service_name
      regex: alertmanager-main
    - action: keep
      source_labels:
      - __meta_kubernetes_endpoint_port_name
      regex: web
  - static_configs:
    - targets:
      - localhost
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestTargetLabels(t *testing.T) {
	cg := &configGenerator{}
	cfg, err := cg.generateConfig(
		&monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: monitoringv1.PrometheusSpec{
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
			},
		},
		map[string]*monitoringv1.ServiceMonitor{
			"testservicemonitor1": &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					TargetLabels: []string{"example", "env"},
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:     "web",
							Interval: "30s",
						},
					},
				},
			},
		},
		map[string]BasicAuthCredentials{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
rule_files: []
scrape_configs:
- job_name: default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_service_label_example
    target_label: example
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_service_label_env
    target_label: env
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  alertmanagers: []
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestPodTargetLabels(t *testing.T) {
	cg := &configGenerator{}
	cfg, err := cg.generateConfig(
		&monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: monitoringv1.PrometheusSpec{
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
			},
		},
		map[string]*monitoringv1.ServiceMonitor{
			"testservicemonitor1": &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testservicemonitor1",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					PodTargetLabels: []string{"example", "env"},
					Endpoints: []monitoringv1.Endpoint{
						{
							Port:     "web",
							Interval: "30s",
						},
					},
				},
			},
		},
		map[string]BasicAuthCredentials{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
rule_files: []
scrape_configs:
- job_name: default/testservicemonitor1/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  scrape_interval: 30s
  relabel_configs:
  - action: keep
    source_labels:
    - __meta_kubernetes_endpoint_port_name
    regex: web
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_pod_label_example
    target_label: example
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_pod_label_env
    target_label: env
    regex: (.+)
    replacement: ${1}
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
  - target_label: endpoint
    replacement: web
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  alertmanagers: []
`

	result := string(cfg)

	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func TestEmptyEndointPorts(t *testing.T) {
	cg := &configGenerator{}
	cfg, err := cg.generateConfig(
		&monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		},
		map[string]*monitoringv1.ServiceMonitor{
			"test": &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: monitoringv1.ServiceMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"foo": "bar",
						},
					},
					Endpoints: []monitoringv1.Endpoint{
						// Add a single endpoint with empty configuration.
						{},
					},
				},
			},
		},
		map[string]BasicAuthCredentials{},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	// If this becomes an endless sink of maintenance, then we should just
	// change this to check that just the `bearer_token_file` is set with
	// something like json-path.
	expected := `global:
  evaluation_interval: 30s
  scrape_interval: 30s
  external_labels:
    prometheus: default/test
    prometheus_replica: $(POD_NAME)
rule_files: []
scrape_configs:
- job_name: default/test/0
  honor_labels: false
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - default
  relabel_configs:
  - action: keep
    source_labels:
    - __meta_kubernetes_service_label_foo
    regex: bar
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Node;(.*)
    replacement: ${1}
    target_label: node
  - source_labels:
    - __meta_kubernetes_endpoint_address_target_kind
    - __meta_kubernetes_endpoint_address_target_name
    separator: ;
    regex: Pod;(.*)
    replacement: ${1}
    target_label: pod
  - source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: service
  - source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod
  - source_labels:
    - __meta_kubernetes_service_name
    target_label: job
    replacement: ${1}
alerting:
  alert_relabel_configs:
  - action: labeldrop
    regex: prometheus_replica
  alertmanagers: []
`

	result := string(cfg)
	if expected != result {
		fmt.Println(pretty.Compare(expected, result))
		t.Fatal("expected Prometheus configuration and actual configuration do not match")
	}
}

func generateTestConfig(version string) ([]byte, error) {
	cg := &configGenerator{}
	replicas := int32(1)
	return cg.generateConfig(
		&monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: monitoringv1.PrometheusSpec{
				Alerting: &monitoringv1.AlertingSpec{
					Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
						{
							Name:      "alertmanager-main",
							Namespace: "default",
							Port:      intstr.FromString("web"),
						},
					},
				},
				ExternalLabels: map[string]string{
					"label1": "value1",
					"label2": "value2",
				},
				Version:  version,
				Replicas: &replicas,
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": "group1",
					},
				},
				RuleSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"role": "rulefile",
					},
				},
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("400Mi"),
					},
				},
				RemoteRead: []monitoringv1.RemoteReadSpec{{
					URL: "https://example.com/remote_read",
				}},
				RemoteWrite: []monitoringv1.RemoteWriteSpec{{
					URL: "https://example.com/remote_write",
				}},
			},
		},
		makeServiceMonitors(),
		map[string]BasicAuthCredentials{},
		nil,
		nil,
		nil,
		nil,
	)
}

func makeServiceMonitors() map[string]*monitoringv1.ServiceMonitor {
	res := map[string]*monitoringv1.ServiceMonitor{}

	res["servicemonitor1"] = &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testservicemonitor1",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group1",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group": "group1",
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "web",
					Interval: "30s",
				},
			},
		},
	}

	res["servicemonitor2"] = &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testservicemonitor2",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group2",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group":  "group2",
					"group3": "group3",
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "web",
					Interval: "30s",
				},
			},
		},
	}

	res["servicemonitor3"] = &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testservicemonitor3",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group4",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group":  "group4",
					"group3": "group5",
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "web",
					Interval: "30s",
					Path:     "/federate",
					Params:   map[string][]string{"metrics[]": {"{__name__=~\"job:.*\"}"}},
				},
			},
		},
	}

	res["servicemonitor4"] = &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testservicemonitor4",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group6",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group":  "group6",
					"group3": "group7",
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "web",
					Interval: "30s",
					MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
						{
							Action:       "drop",
							Regex:        "my-job-pod-.+",
							SourceLabels: []string{"pod_name"},
						},
						{
							Action:       "drop",
							Regex:        "test",
							SourceLabels: []string{"namespace"},
						},
					},
				},
			},
		},
	}

	res["servicemonitor5"] = &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testservicemonitor4",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group8",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group":  "group8",
					"group3": "group9",
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "web",
					Interval: "30s",
					RelabelConfigs: []*monitoringv1.RelabelConfig{
						{
							Action:       "replace",
							Regex:        "(.*)",
							Replacement:  "$1",
							SourceLabels: []string{"__meta_kubernetes_pod_ready"},
							TargetLabel:  "pod_ready",
						},
						{
							Action:       "replace",
							Regex:        "(.*)",
							Replacement:  "$1",
							SourceLabels: []string{"__meta_kubernetes_pod_node_name"},
							TargetLabel:  "nodename",
						},
					},
				},
			},
		},
	}

	return res
}
