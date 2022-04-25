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
	"fmt"
	"reflect"
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"github.com/prometheus/prometheus/model/relabel"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kylelemons/godebug/pretty"
)

func TestListOptions(t *testing.T) {
	for i := 0; i < 1000; i++ {
		o := ListOptions("test")
		if o.LabelSelector != "app.kubernetes.io/name=prometheus,prometheus=test" && o.LabelSelector != "prometheus=test,app.kubernetes.io/name=prometheus" {
			t.Fatalf("LabelSelector not computed correctly\n\nExpected: \"app.kubernetes.io/name=prometheus,prometheus=test\"\n\nGot:      %#+v", o.LabelSelector)
		}
	}
}

func TestCreateStatefulSetInputHash(t *testing.T) {
	p1 := monitoringv1.Prometheus{}
	p1.Spec.Version = "v1.7.0"
	p2 := monitoringv1.Prometheus{}
	p2.Spec.Version = "v1.7.2"
	c := operator.Config{}

	p1Hash, err := createSSetInputHash(p1, c, []string{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	p2Hash, err := createSSetInputHash(p2, c, []string{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if p1Hash == p2Hash {
		t.Fatal("expected two different Prometheus CRDs to result in two different hash but got equal hash")
	}
}

func TestGetNodeAddresses(t *testing.T) {
	cases := []struct {
		name              string
		nodes             *v1.NodeList
		expectedAddresses []string
		expectedErrors    int
	}{
		{
			name: "simple",
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node-0",
						},
						Status: v1.NodeStatus{
							Addresses: []v1.NodeAddress{
								{
									Address: "10.0.0.1",
									Type:    v1.NodeInternalIP,
								},
							},
						},
					},
				},
			},
			expectedAddresses: []string{"10.0.0.1"},
			expectedErrors:    0,
		},
		{
			// Replicates #1815
			name: "missing ip on one node",
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node-0",
						},
						Status: v1.NodeStatus{
							Addresses: []v1.NodeAddress{
								{
									Address: "node-0",
									Type:    v1.NodeHostName,
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node-1",
						},
						Status: v1.NodeStatus{
							Addresses: []v1.NodeAddress{
								{
									Address: "10.0.0.1",
									Type:    v1.NodeInternalIP,
								},
							},
						},
					},
				},
			},
			expectedAddresses: []string{"10.0.0.1"},
			expectedErrors:    1,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			addrs, errs := getNodeAddresses(c.nodes)
			if len(errs) != c.expectedErrors {
				t.Errorf("Expected %d errors, got %d. Errors: %v", c.expectedErrors, len(errs), errs)
			}
			ips := make([]string, 0)
			for _, addr := range addrs {
				ips = append(ips, addr.IP)
			}
			if !reflect.DeepEqual(ips, c.expectedAddresses) {
				t.Error(pretty.Compare(ips, c.expectedAddresses))
			}
		})
	}
}

func TestStatefulSetKeyToPrometheusKey(t *testing.T) {
	cases := []struct {
		input         string
		expectedKey   string
		expectedMatch bool
	}{
		{
			input:         "namespace/prometheus-test",
			expectedKey:   "namespace/test",
			expectedMatch: true,
		},
		{
			input:         "namespace/prometheus-test-shard-1",
			expectedKey:   "namespace/test",
			expectedMatch: true,
		},
		{
			input:         "allns-z-thanosrulercreatedeletecluster-qcwdmj-0/thanos-ruler-test",
			expectedKey:   "",
			expectedMatch: false,
		},
	}

	for _, c := range cases {
		match, key := statefulSetKeyToPrometheusKey(c.input)
		if c.expectedKey != key {
			t.Fatalf("Expected prometheus key %q got %q", c.expectedKey, key)
		}
		if c.expectedMatch != match {
			notExp := ""
			if !c.expectedMatch {
				notExp = "not "
			}
			t.Fatalf("Expected input %sto be matching a prometheus key, but did not", notExp)
		}
	}
}

func TestPrometheusKeyToStatefulSetKey(t *testing.T) {
	cases := []struct {
		name     string
		shard    int
		expected string
	}{
		{
			name:     "namespace/test",
			shard:    0,
			expected: "namespace/prometheus-test",
		},
		{
			name:     "namespace/test",
			shard:    1,
			expected: "namespace/prometheus-test-shard-1",
		},
	}

	for _, c := range cases {
		got := prometheusKeyToStatefulSetKey(c.name, c.shard)
		if c.expected != got {
			t.Fatalf("Expected key %q got %q", c.expected, got)
		}
	}
}

func TestValidateRemoteWriteConfig(t *testing.T) {
	cases := []struct {
		name      string
		spec      monitoringv1.RemoteWriteSpec
		expectErr bool
	}{
		{
			name: "with_OAuth2",
			spec: monitoringv1.RemoteWriteSpec{
				OAuth2: &monitoringv1.OAuth2{},
			},
		}, {
			name: "with_SigV4",
			spec: monitoringv1.RemoteWriteSpec{
				Sigv4: &monitoringv1.Sigv4{},
			},
		},
		{
			name: "with_OAuth2_and_SigV4",
			spec: monitoringv1.RemoteWriteSpec{
				OAuth2: &monitoringv1.OAuth2{},
				Sigv4:  &monitoringv1.Sigv4{},
			},
			expectErr: true,
		}, {
			name: "with_OAuth2_and_BasicAuth",
			spec: monitoringv1.RemoteWriteSpec{
				OAuth2:    &monitoringv1.OAuth2{},
				BasicAuth: &monitoringv1.BasicAuth{},
			},
			expectErr: true,
		}, {
			name: "with_BasicAuth_and_SigV4",
			spec: monitoringv1.RemoteWriteSpec{
				BasicAuth: &monitoringv1.BasicAuth{},
				Sigv4:     &monitoringv1.Sigv4{},
			},
			expectErr: true,
		}, {
			name: "with_BasicAuth_and_SigV4_and_OAuth2",
			spec: monitoringv1.RemoteWriteSpec{
				BasicAuth: &monitoringv1.BasicAuth{},
				Sigv4:     &monitoringv1.Sigv4{},
				OAuth2:    &monitoringv1.OAuth2{},
			},
			expectErr: true,
		},
	}
	for _, c := range cases {
		test := c
		t.Run(test.name, func(t *testing.T) {
			err := validateRemoteWriteSpec(test.spec)
			if err != nil && !test.expectErr {
				t.Fatalf("unexpected error occurred: %v", err)
			}
			if err == nil && test.expectErr {
				t.Fatalf("expected an error, got nil")
			}
		})
	}
}

func TestValidateRelabelConfig(t *testing.T) {
	defaultRegexp, err := relabel.DefaultRelabelConfig.Regex.MarshalYAML()
	if err != nil {
		t.Errorf("Could not marshal relabel.DefaultRelabelConfig.Regex: %v", err)
	}
	defaultRegex, ok := defaultRegexp.(string)
	if !ok {
		t.Errorf("Could not assert marshaled defaultRegexp as string: %v", defaultRegexp)
	}

	defaultSourceLabels := []monitoringv1.LabelName{}
	for _, label := range relabel.DefaultRelabelConfig.SourceLabels {
		defaultSourceLabels = append(defaultSourceLabels, monitoringv1.LabelName(label))
	}

	for _, tc := range []struct {
		scenario      string
		relabelConfig monitoringv1.RelabelConfig
		expectedErr   bool
	}{
		// Test invalid regex expression
		{
			scenario: "Invalid regex",
			relabelConfig: monitoringv1.RelabelConfig{
				Regex: "invalid regex)",
			},
			expectedErr: true,
		},
		// Test invalid target label
		{
			scenario: "invalid target label",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "replace",
				TargetLabel: "l\\${3}",
			},
			expectedErr: true,
		},
		// Test empty target label for action replace
		{
			scenario: "empty target label for replace action",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "replace",
				TargetLabel: "",
			},
			expectedErr: true,
		},
		// Test empty target label for action hashmod
		{
			scenario: "empty target label for hashmod action",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "hashmod",
				TargetLabel: "",
			},
			expectedErr: true,
		},
		// Test invalid hashmod relabel config
		{
			scenario: "invalid hashmod config",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"instance"},
				Action:       "hashmod",
				Modulus:      0,
				TargetLabel:  "__tmp_hashmod",
			},
			expectedErr: true,
		},
		// Test invalid labelmap relabel config
		{
			scenario: "invalid labelmap config",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "labelmap",
				Regex:       "__meta_kubernetes_service_label_(.+)",
				Replacement: "some-name-value",
			},
			expectedErr: true,
		},
		// Test valid labelmap relabel config when replacement not specified
		{
			scenario: "valid labelmap config",
			relabelConfig: monitoringv1.RelabelConfig{
				Action: "labelmap",
				Regex:  "__meta_kubernetes_service_label_(.+)",
			},
		},
		// Test valid labelmap relabel config with replacement specified
		{
			scenario: "valid labelmap config",
			relabelConfig: monitoringv1.RelabelConfig{
				Action:      "labelmap",
				Regex:       "__meta_kubernetes_service_label_(.+)",
				Replacement: "${2}",
			},
		},
		// Test invalid labelkeep relabel config
		{
			scenario: "invalid labelkeep config",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"instance"},
				Action:       "labelkeep",
				TargetLabel:  "__tmp_labelkeep",
			},
			expectedErr: true,
		},
		// Test valid labelkeep relabel config
		{
			scenario: "valid labelkeep config",
			relabelConfig: monitoringv1.RelabelConfig{
				Action: "labelkeep",
			},
		},
		// Test valid labeldrop relabel config
		{
			scenario: "valid labeldrop config",
			relabelConfig: monitoringv1.RelabelConfig{
				Action: "labeldrop",
				Regex:  "replica",
			},
		},
		{
			scenario: "valid labeldrop config with default values",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: defaultSourceLabels,
				Separator:    relabel.DefaultRelabelConfig.Separator,
				TargetLabel:  relabel.DefaultRelabelConfig.TargetLabel,
				Regex:        defaultRegex,
				Modulus:      relabel.DefaultRelabelConfig.Modulus,
				Replacement:  relabel.DefaultRelabelConfig.Replacement,
				Action:       "labeldrop",
			},
		},
		// Test valid relabel config
		{
			scenario: "valid hashmod config",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"instance"},
				Action:       "hashmod",
				Modulus:      10,
				TargetLabel:  "__tmp_hashmod",
			},
		},
		// Test valid relabel config
		{
			scenario: "valid replace config",
			relabelConfig: monitoringv1.RelabelConfig{
				SourceLabels: []monitoringv1.LabelName{"__address__"},
				Action:       "replace",
				Regex:        "([^:]+)(?::\\d+)?",
				Replacement:  "$1:80",
				TargetLabel:  "__address__",
			},
		},
	} {
		t.Run(fmt.Sprintf("case %s", tc.scenario), func(t *testing.T) {
			err := validateRelabelConfig(tc.relabelConfig)
			if err != nil && !tc.expectedErr {
				t.Fatalf("expected no error, got: %v", err)
			}
			if err == nil && tc.expectedErr {
				t.Fatalf("expected an error, got nil")
			}
		})
	}
}

func TestValidateProberUrl(t *testing.T) {
	for _, tc := range []struct {
		scenario    string
		proberSpec  monitoringv1.ProberSpec
		expectedErr bool
	}{
		{
			scenario: "url starting with http",
			proberSpec: monitoringv1.ProberSpec{
				URL: "http://blackbox-exporter.example.com",
			},
			expectedErr: true,
		},
		{
			scenario: "url starting with https",
			proberSpec: monitoringv1.ProberSpec{
				URL: "https://blackbox-exporter.example.com",
			},
			expectedErr: true,
		},
		{
			scenario: "url starting with ftp",
			proberSpec: monitoringv1.ProberSpec{
				URL: "ftp://fileserver.com",
			},
			expectedErr: true,
		},
		{
			scenario: "ip address as prober url",
			proberSpec: monitoringv1.ProberSpec{
				URL: "192.168.178.3",
			},
		},
		{
			scenario: "ip address:port as prober url",
			proberSpec: monitoringv1.ProberSpec{
				URL: "192.168.178.3:9090",
			},
		},
		{
			scenario: "dnsname as prober url",
			proberSpec: monitoringv1.ProberSpec{
				URL: "blackbox-exporter.example.com",
			},
		},
		{
			scenario: "dnsname:port as prober url",
			proberSpec: monitoringv1.ProberSpec{
				URL: "blackbox-exporter.example.com:8080",
			},
		},
		{
			scenario: "hostname as prober url",
			proberSpec: monitoringv1.ProberSpec{
				URL: "localhost",
			},
		},
		{
			scenario: "hostname starting with a digit as prober url",
			proberSpec: monitoringv1.ProberSpec{
				URL: "12-exporter.example.com",
			},
		},
	} {
		t.Run(fmt.Sprintf("case %s %s", tc.scenario, tc.proberSpec.URL), func(t *testing.T) {
			err := validateProberURL(tc.proberSpec.URL)
			if err != nil && !tc.expectedErr {
				t.Fatalf("expected no error, got: %v", err)
			}
			if err == nil && tc.expectedErr {
				t.Fatalf("expected an error, got nil")
			}
		})
	}
}

func TestValidateScrapeIntervalAndTimeout(t *testing.T) {
	for _, tc := range []struct {
		scenario    string
		prometheus  monitoringv1.Prometheus
		smSpec      monitoringv1.ServiceMonitorSpec
		expectedErr bool
	}{
		{
			scenario: "scrape interval and timeout specified at service monitor spec but invalid #1",
			smSpec: monitoringv1.ServiceMonitorSpec{
				Endpoints: []monitoringv1.Endpoint{
					{
						Interval:      "30s",
						ScrapeTimeout: "45s",
					},
				},
			},
			expectedErr: true,
		},
		{
			scenario: "scrape interval and timeout specified at service monitor spec but invalid #2",
			smSpec: monitoringv1.ServiceMonitorSpec{
				Endpoints: []monitoringv1.Endpoint{
					{
						Interval:      "30 s",
						ScrapeTimeout: "10s",
					},
				},
			},
			expectedErr: true,
		},
		{
			scenario: "scrape interval and timeout specified at service monitor spec are valid",
			smSpec: monitoringv1.ServiceMonitorSpec{
				Endpoints: []monitoringv1.Endpoint{
					{
						Interval:      "15s",
						ScrapeTimeout: "10s",
					},
				},
			},
		},
		{
			scenario: "only scrape timeout specified at service monitor spec but invalid compared to global scrapeInterval",
			prometheus: monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						ScrapeInterval: "15s",
					},
				},
			},
			smSpec: monitoringv1.ServiceMonitorSpec{
				Endpoints: []monitoringv1.Endpoint{
					{
						ScrapeTimeout: "20s",
					},
				},
			},
			expectedErr: true,
		},
		{
			scenario: "only scrape timeout specified at service monitor spec but invalid compared to default global scrapeInterval",
			smSpec: monitoringv1.ServiceMonitorSpec{
				Endpoints: []monitoringv1.Endpoint{
					{
						ScrapeTimeout: "60s",
					},
				},
			},
			expectedErr: true,
		},
	} {
		t.Run(fmt.Sprintf("case %s", tc.scenario), func(t *testing.T) {
			for _, endpoint := range tc.smSpec.Endpoints {
				err := validateScrapeIntervalAndTimeout(&tc.prometheus, endpoint.Interval, endpoint.ScrapeTimeout)
				t.Logf("err %v", err)
				if err != nil && !tc.expectedErr {
					t.Fatalf("expected no error, got: %v", err)
				}
				if err == nil && tc.expectedErr {
					t.Fatalf("expected an error, got nil")
				}
			}
		})
	}
}
