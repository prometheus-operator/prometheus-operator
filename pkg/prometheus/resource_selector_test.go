// Copyright 2023 The prometheus-operator Authors
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
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

var (
	certsDir = "../../test/e2e/tls_certs/"
)

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
}

func TestSelectProbes(t *testing.T) {
	for _, tc := range []struct {
		scenario    string
		updateSpec  func(*monitoringv1.ProbeSpec)
		promVersion string
		valid       bool
		scrapeClass *string
	}{
		{
			scenario: "url starting with http",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "http://blackbox-exporter.example.com"
			},
			valid: false,
		},
		{
			scenario: "url starting with https",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "https://blackbox-exporter.example.com"
			},
			valid: false,
		},
		{
			scenario: "url starting with ftp",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "ftp://fileserver.com"
			},
			valid: false,
		},
		{
			scenario: "ip address as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "192.168.178.3"
			},
			valid: true,
		},
		{
			scenario: "ip address:port as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "192.168.178.3:9090"
			},
			valid: true,
		},
		{
			scenario: "dnsname as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "blackbox-exporter.example.com"
			},
			valid: true,
		},
		{
			scenario: "dnsname:port as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "blackbox-exporter.example.com:8080"
			},
			valid: true,
		},
		{
			scenario: "hostname as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "localhost"
			},
			valid: true,
		},
		{
			scenario: "hostname starting with a digit as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "12-exporter.example.com"
			},
			valid: true,
		},
		{
			scenario: "ipv6 address as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "::1"
			},
			valid: true,
		},
		{
			scenario: "ipv6 full address as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "2001:db8::1"
			},
			valid: true,
		},
		{
			scenario: "ipv6 address with port as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "[::1]:9090"
			},
			valid: true,
		},
		{
			scenario: "ipv6 full address with port as prober url",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.URL = "[2001:db8::1]:9090"
			},
			valid: true,
		},
		{
			scenario: "invalid proxyconfig due to invalid proxyurl",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyURL:             ptr.To("http://xxx-${dev}.svc.cluster.local:80"),
					NoProxy:              ptr.To("0.0.0.0"),
					ProxyFromEnvironment: ptr.To(false),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "invalid proxyconfig due to proxy environment set to true and proxyurl defined",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyURL:             ptr.To("http://no-proxy.com"),
					ProxyFromEnvironment: ptr.To(true),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "invalid proxyconfig due to proxy environment set to true and noproxy defined",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.ProxyConfig = monitoringv1.ProxyConfig{
					NoProxy:              ptr.To("0.0.0.0"),
					ProxyFromEnvironment: ptr.To(true),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "invalid proxyconfig due to invalid secret secret key",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyURL:             ptr.To("http://no-proxy.com"),
					ProxyFromEnvironment: ptr.To(true),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "invalid_key",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "invalid proxyconfig due to proxy from environment set to false and proxyurl and noproxy not defined",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyFromEnvironment: ptr.To(false),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "valid proxyconfig",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.ProberSpec.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyURL:             ptr.To("http://no-proxy.com"),
					NoProxy:              ptr.To("0.0.0.0"),
					ProxyFromEnvironment: ptr.To(false),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "valid metric relabeling config",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.MetricRelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  "valid",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "utf-8 metric relabeling config with prom2",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.MetricRelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  " invalid label name",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			promVersion: "2.55.0",
			valid:       false,
		},
		{
			scenario: "utf-8 metric relabeling config with prom3",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.MetricRelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  " invalid label name",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			promVersion: "3.5.0",
			valid:       true,
		},
		{
			scenario: "valid static relabeling config",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.Targets.StaticConfig.RelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  "valid",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "utf-8 static relabeling config with prom2",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.Targets.StaticConfig.RelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  " invalid label name",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			promVersion: "2.55.0",
			valid:       false,
		},
		{
			scenario: "utf-8 static relabeling config with prom3",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.Targets.StaticConfig.RelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  " invalid label name",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			promVersion: "3.5.0",
			valid:       true,
		},
		{
			scenario: "valid ingress relabeling config",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.Targets.StaticConfig = nil
				ps.Targets.Ingress = &monitoringv1.ProbeTargetIngress{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "utf-8 ingress relabeling config with prom2",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.Targets.Ingress = &monitoringv1.ProbeTargetIngress{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				}
			},
			promVersion: "2.55.0",
			valid:       false,
		},
		{
			scenario: "utf-8 ingress relabeling config with prom3",
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.Targets.Ingress = &monitoringv1.ProbeTargetIngress{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				}
			},
			promVersion: "3.5.0",
			valid:       true,
		},
		{
			scenario:    "inexistent scrape class",
			scrapeClass: ptr.To("inexistent"),
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.Targets.StaticConfig = nil
				ps.Targets.Ingress = &monitoringv1.ProbeTargetIngress{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario:    "existent scrape class",
			scrapeClass: ptr.To("existent"),
			updateSpec: func(ps *monitoringv1.ProbeSpec) {
				ps.Targets.StaticConfig = nil
				ps.Targets.Ingress = &monitoringv1.ProbeTargetIngress{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				}
			},
			valid: true,
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			cs := fake.NewSimpleClientset(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"key1": []byte("val1"),
					},
				},
			)

			p := &monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: tc.promVersion,
						ScrapeClasses: []monitoringv1.ScrapeClass{
							{
								Name: "existent",
							},
						},
					},
				},
			}
			rs, err := NewResourceSelector(
				newLogger(),
				p,
				assets.NewStoreBuilder(cs.CoreV1(), cs.CoreV1()),
				nil,
				operator.NewMetrics(prometheus.NewPedanticRegistry()),
				operator.NewFakeRecorder(1, p),
			)
			require.NoError(t, err)

			probe := &monitoringv1.Probe{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: monitoringv1.ProbeSpec{
					ProberSpec: monitoringv1.ProberSpec{
						URL: "example.com:80",
					},
					Targets: monitoringv1.ProbeTargets{
						StaticConfig: &monitoringv1.ProbeTargetStaticConfig{},
					},
				},
			}

			if tc.scrapeClass != nil {
				probe.Spec.ScrapeClassName = tc.scrapeClass
			}

			tc.updateSpec(&probe.Spec)

			probes, err := rs.SelectProbes(context.Background(), func(_ string, _ labels.Selector, appendFn cache.AppendFunc) error {
				appendFn(probe)
				return nil
			})

			require.NoError(t, err)

			valid := probes.ValidResources()

			require.Len(t, probes, 1)
			if tc.valid {
				require.Len(t, valid, 1)
			} else {
				require.Len(t, valid, 0)
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
		t.Run(tc.scenario, func(t *testing.T) {
			for _, endpoint := range tc.smSpec.Endpoints {
				err := validateScrapeIntervalAndTimeout(&tc.prometheus, endpoint.Interval, endpoint.ScrapeTimeout)
				t.Logf("err %v", err)
				if tc.expectedErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
			}
		})
	}
}

func TestSelectServiceMonitors(t *testing.T) {
	ca, err := os.ReadFile(certsDir + "ca.crt")
	require.NoError(t, err)

	cert, err := os.ReadFile(certsDir + "client.crt")
	require.NoError(t, err)

	key, err := os.ReadFile(certsDir + "client.key")
	require.NoError(t, err)

	for _, tc := range []struct {
		scenario    string
		updateSpec  func(*monitoringv1.ServiceMonitorSpec)
		promVersion string
		valid       bool
		scrapeClass *string
	}{
		{
			scenario: "valid metric relabeling config",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			valid: true,
		},
		{
			scenario: "utf-8 metric relabeling config with prom2",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			promVersion: "2.55.0",
			valid:       false,
		},
		{
			scenario: "utf-8 metric relabeling config with prom3",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			promVersion: "3.5.0",
			valid:       true,
		},
		{
			scenario: "valid relabeling config",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			valid: true,
		},
		{
			scenario: "utf-8 relabeling config with prom2",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			promVersion: "2.55.0",
			valid:       false,
		},
		{
			scenario: "utf-8 relabeling config with prom3",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			promVersion: "3.5.0",
			valid:       true,
		},
		{
			scenario: "valid TLS config with CA, cert and key",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
						HTTPConfigWithTLSFiles: monitoringv1.HTTPConfigWithTLSFiles{
							TLSConfig: &monitoringv1.TLSConfig{
								SafeTLSConfig: monitoringv1.SafeTLSConfig{
									CA: monitoringv1.SecretOrConfigMap{
										Secret: &v1.SecretKeySelector{
											Key: "ca",
											LocalObjectReference: v1.LocalObjectReference{
												Name: "secret",
											},
										},
									},
									Cert: monitoringv1.SecretOrConfigMap{
										Secret: &v1.SecretKeySelector{
											Key: "cert",
											LocalObjectReference: v1.LocalObjectReference{
												Name: "secret",
											},
										},
									},
									KeySecret: &v1.SecretKeySelector{
										Key: "key",
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
									},
								},
							},
						},
					},
				})
			},
			valid: true,
		},
		{
			scenario: "invalid TLS config with both CA and CAFile",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
						HTTPConfigWithTLSFiles: monitoringv1.HTTPConfigWithTLSFiles{
							TLSConfig: &monitoringv1.TLSConfig{
								SafeTLSConfig: monitoringv1.SafeTLSConfig{
									CA: monitoringv1.SecretOrConfigMap{
										Secret: &v1.SecretKeySelector{
											Key: "ca",
											LocalObjectReference: v1.LocalObjectReference{
												Name: "secret",
											},
										},
									},
								},
								TLSFilesConfig: monitoringv1.TLSFilesConfig{
									CAFile: "/etc/secrets/tls/ca.crt",
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario: "invalid TLS config with both CA Secret and Configmap",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
						HTTPConfigWithTLSFiles: monitoringv1.HTTPConfigWithTLSFiles{
							TLSConfig: &monitoringv1.TLSConfig{
								SafeTLSConfig: monitoringv1.SafeTLSConfig{
									CA: monitoringv1.SecretOrConfigMap{
										Secret: &v1.SecretKeySelector{
											Key: "ca",
											LocalObjectReference: v1.LocalObjectReference{
												Name: "secret",
											},
										},
										ConfigMap: &v1.ConfigMapKeySelector{
											Key: "ca",
											LocalObjectReference: v1.LocalObjectReference{
												Name: "configmap",
											},
										},
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario: "invalid TLS config with invalid CA data",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
						HTTPConfigWithTLSFiles: monitoringv1.HTTPConfigWithTLSFiles{
							TLSConfig: &monitoringv1.TLSConfig{
								SafeTLSConfig: monitoringv1.SafeTLSConfig{
									CA: monitoringv1.SecretOrConfigMap{
										Secret: &v1.SecretKeySelector{
											Key: "invalid_ca",
											LocalObjectReference: v1.LocalObjectReference{
												Name: "secret",
											},
										},
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario: "invalid TLS config with cert and missing key",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
						HTTPConfigWithTLSFiles: monitoringv1.HTTPConfigWithTLSFiles{
							TLSConfig: &monitoringv1.TLSConfig{
								SafeTLSConfig: monitoringv1.SafeTLSConfig{
									Cert: monitoringv1.SecretOrConfigMap{
										Secret: &v1.SecretKeySelector{
											Key: "cert",
											LocalObjectReference: v1.LocalObjectReference{
												Name: "secret",
											},
										},
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario: "invalid TLS config with key and missing cert",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
						HTTPConfigWithTLSFiles: monitoringv1.HTTPConfigWithTLSFiles{
							TLSConfig: &monitoringv1.TLSConfig{
								SafeTLSConfig: monitoringv1.SafeTLSConfig{
									KeySecret: &v1.SecretKeySelector{
										Key: "key",
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario: "invalid TLS config with key and invalid cert",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
						HTTPConfigWithTLSFiles: monitoringv1.HTTPConfigWithTLSFiles{
							TLSConfig: &monitoringv1.TLSConfig{
								SafeTLSConfig: monitoringv1.SafeTLSConfig{
									Cert: monitoringv1.SecretOrConfigMap{
										Secret: &v1.SecretKeySelector{
											Key: "invalid_ca",
											LocalObjectReference: v1.LocalObjectReference{
												Name: "secret",
											},
										},
									},
									KeySecret: &v1.SecretKeySelector{
										Key: "key",
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario: "valid proxy config",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				})
			},
			valid: true,
		},
		{
			scenario: "invalid proxy config with invalid secret key",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "invalid_key",
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario: "invalid proxy config due to invalid proxy url",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://xxx-${dev}.svc.cluster.local:80"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario: "invalid proxy config with noProxy defined but proxy from environment set to true",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
						ProxyConfig: monitoringv1.ProxyConfig{
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(true),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario: "invalid proxy config with proxy url defined but proxy from environment set to true",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							ProxyFromEnvironment: ptr.To(true),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario: "invalid proxy config only with proxy connect header defined",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario:    "inexistent Scrape Class",
			scrapeClass: ptr.To("inexistent"),
			updateSpec: func(_ *monitoringv1.ServiceMonitorSpec) {
			},
			valid: false,
		},
		{
			scenario:    "existent Scrape Class",
			scrapeClass: ptr.To("existent"),
			updateSpec: func(_ *monitoringv1.ServiceMonitorSpec) {
			},
			valid: true,
		},
		{
			scenario: "utf-8 mixed endpoints with prom2",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			promVersion: "2.55.0",
			valid:       false,
		},
		{
			scenario: "utf-8 mixed endpoints with prom3",
			updateSpec: func(sm *monitoringv1.ServiceMonitorSpec) {
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
				sm.Endpoints = append(sm.Endpoints, monitoringv1.Endpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			promVersion: "3.5.0",
			valid:       true,
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			cs := fake.NewSimpleClientset(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"ca":         ca,
						"invalid_ca": []byte("garbage"),
						"cert":       cert,
						"key":        key,
						"key1":       []byte("val1"),
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "configmap",
						Namespace: "test",
					},
					Data: map[string]string{
						"ca":   string(ca),
						"cert": string(cert),
					},
				},
			)

			p := &monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: tc.promVersion,
						ScrapeClasses: []monitoringv1.ScrapeClass{
							{
								Name: "existent",
							},
						},
					},
				},
			}
			rs, err := NewResourceSelector(
				newLogger(),
				p,
				assets.NewStoreBuilder(cs.CoreV1(), cs.CoreV1()),
				nil,
				operator.NewMetrics(prometheus.NewPedanticRegistry()),
				operator.NewFakeRecorder(1, p),
			)
			require.NoError(t, err)

			sm := &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: monitoringv1.ServiceMonitorSpec{},
			}

			if tc.scrapeClass != nil {
				sm.Spec.ScrapeClassName = tc.scrapeClass
			}

			tc.updateSpec(&sm.Spec)

			sms, err := rs.SelectServiceMonitors(context.Background(), func(_ string, _ labels.Selector, appendFn cache.AppendFunc) error {
				appendFn(sm)
				return nil
			})

			require.NoError(t, err)
			require.Len(t, sms, 1)

			valid := sms.ValidResources()
			if tc.valid {
				require.Len(t, valid, 1)
			} else {
				require.Empty(t, valid)
			}
		})
	}
}

func TestSelectPodMonitors(t *testing.T) {
	for _, tc := range []struct {
		scenario    string
		updateSpec  func(*monitoringv1.PodMonitorSpec)
		promVersion string
		valid       bool
		scrapeClass *string
	}{
		{
			scenario: "valid metric relabeling config",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			valid: true,
		},
		{
			scenario: "utf-8 metric relabeling config with prom2",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			promVersion: "2.55.0",
			valid:       false,
		},
		{
			scenario: "utf-8 metric relabeling config with prom3",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			promVersion: "3.5.0",
			valid:       true,
		},
		{
			scenario: "valid relabeling config",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			valid: true,
		},
		{
			scenario: "utf-8 relabeling config with prom2",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			promVersion: "2.55.0",
			valid:       false,
		},
		{
			scenario: "utf-8 relabeling config with prom3",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					RelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			promVersion: "3.5.0",
			valid:       true,
		},
		{
			scenario: "valid proxy config",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					HTTPConfigWithProxy: monitoringv1.HTTPConfigWithProxy{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				})
			},
			valid: true,
		},
		{
			scenario: "invalid proxy config with invalid secret key",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					HTTPConfigWithProxy: monitoringv1.HTTPConfigWithProxy{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "invalid_key",
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario: "invalid proxy config due to invalid proxy url",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					HTTPConfigWithProxy: monitoringv1.HTTPConfigWithProxy{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://xxx-${dev}.svc.cluster.local:80"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario: "invalid proxy config with noProxy defined but proxy from environment set to true",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					HTTPConfigWithProxy: monitoringv1.HTTPConfigWithProxy{
						ProxyConfig: monitoringv1.ProxyConfig{
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(true),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario: "invalid proxy config with proxy url defined but proxy from environment set to true",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					HTTPConfigWithProxy: monitoringv1.HTTPConfigWithProxy{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							ProxyFromEnvironment: ptr.To(true),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario: "invalid proxy config only with proxy connect header defined",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					HTTPConfigWithProxy: monitoringv1.HTTPConfigWithProxy{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				})
			},
			valid: false,
		},
		{
			scenario:    "Inexistent Scrape Class",
			scrapeClass: ptr.To("inexistent"),
			updateSpec: func(_ *monitoringv1.PodMonitorSpec) {
			},
			valid: false,
		},
		{
			scenario:    "existent Scrape Class",
			scrapeClass: ptr.To("existent"),
			updateSpec: func(_ *monitoringv1.PodMonitorSpec) {
			},
			valid: true,
		},
		{
			scenario: "utf-8 mixed Endpoints with prom2",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			promVersion: "2.55.0",
			valid:       false,
		},
		{
			scenario: "utf-8 mixed Endpoints with prom3",
			updateSpec: func(pm *monitoringv1.PodMonitorSpec) {
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  " invalid label name",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
				pm.PodMetricsEndpoints = append(pm.PodMetricsEndpoints, monitoringv1.PodMetricsEndpoint{
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							Action:       "Replace",
							TargetLabel:  "valid",
							SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
						},
					},
				})
			},
			promVersion: "3.5.0",
			valid:       true,
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			cs := fake.NewSimpleClientset(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"key1": []byte("val1"),
					},
				},
			)

			p := &monitoringv1.Prometheus{
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: tc.promVersion,
						ScrapeClasses: []monitoringv1.ScrapeClass{
							{
								Name: "existent",
							},
						},
					},
				},
			}
			rs, err := NewResourceSelector(
				newLogger(),
				p,
				assets.NewStoreBuilder(cs.CoreV1(), cs.CoreV1()),
				nil,
				operator.NewMetrics(prometheus.NewPedanticRegistry()),
				operator.NewFakeRecorder(1, p),
			)
			require.NoError(t, err)

			pm := &monitoringv1.PodMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			}

			if tc.scrapeClass != nil {
				pm.Spec.ScrapeClassName = tc.scrapeClass
			}

			tc.updateSpec(&pm.Spec)

			sms, err := rs.SelectPodMonitors(context.Background(), func(_ string, _ labels.Selector, appendFn cache.AppendFunc) error {
				appendFn(pm)
				return nil
			})

			require.NoError(t, err)
			require.Len(t, sms, 1)

			valid := sms.ValidResources()
			if tc.valid {
				require.Len(t, valid, 1)
			} else {
				require.Empty(t, valid)
			}
		})
	}
}

func TestSelectScrapeConfigs(t *testing.T) {
	ca, err := os.ReadFile(certsDir + "ca.crt")
	require.NoError(t, err)

	cert, err := os.ReadFile(certsDir + "client.crt")
	require.NoError(t, err)

	key, err := os.ReadFile(certsDir + "client.key")
	require.NoError(t, err)
	for _, tc := range []struct {
		scenario    string
		updateSpec  func(*monitoringv1alpha1.ScrapeConfigSpec)
		valid       bool
		promVersion string
		scrapeClass *string
	}{
		{
			scenario: "valid relabeling config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.RelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  "valid",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "utf-8 relabeling config with prom2",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.RelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  " invalid label name",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			promVersion: "2.55.0",
			valid:       false,
		},
		{
			scenario: "utf-8 relabeling config with prom3",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.RelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  " invalid label name",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			promVersion: "3.5.0",
			valid:       true,
		},
		{
			scenario: "valid metric relabeling config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.MetricRelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  "valid",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "utf-8 metric relabeling config with prom2",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.MetricRelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  " invalid label name",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			promVersion: "2.55.0",
			valid:       false,
		},
		{
			scenario: "utf-8 metric relabeling config with prom3",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.MetricRelabelConfigs = []monitoringv1.RelabelConfig{
					{
						Action:       "Replace",
						TargetLabel:  " invalid label name",
						SourceLabels: []monitoringv1.LabelName{"foo", "bar"},
					},
				}
			},
			promVersion: "3.5.0",
			valid:       true,
		},
		{
			scenario: "valid proxy config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyURL:             ptr.To("http://no-proxy.com"),
					NoProxy:              ptr.To("0.0.0.0"),
					ProxyFromEnvironment: ptr.To(false),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "invalid proxy config with proxyConnectHeaders but no proxyUrl defined or proxyFromEnvironment set to true",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "invalid proxy config with proxy from environment set to true but proxyUrl defined",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyURL:             ptr.To("http://no-proxy.com"),
					ProxyFromEnvironment: ptr.To(true),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "invalid proxy config with proxyFromEnvironment set to true but noProxy defined",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					NoProxy:              ptr.To("0.0.0.0"),
					ProxyFromEnvironment: ptr.To(true),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "invalid proxy config with invalid secret key",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyURL:             ptr.To("http://no-proxy.com"),
					NoProxy:              ptr.To("0.0.0.0"),
					ProxyFromEnvironment: ptr.To(false),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "invalid-key",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "invalid proxy config with noProxy defined and but no proxyUrl defined",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					NoProxy: ptr.To("0.0.0.0"),
				}
			},
			valid: false,
		},
		{
			scenario: "valid proxy config with multi header values",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyURL:             ptr.To("http://no-proxy.com"),
					NoProxy:              ptr.To("0.0.0.0"),
					ProxyFromEnvironment: ptr.To(false),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "invalid proxy config with one invalid secret key",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ProxyConfig = monitoringv1.ProxyConfig{
					ProxyURL:             ptr.To("http://no-proxy.com"),
					NoProxy:              ptr.To("0.0.0.0"),
					ProxyFromEnvironment: ptr.To(false),
					ProxyConnectHeader: map[string][]v1.SecretKeySelector{
						"header": {
							{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "invalid-key",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "staticConfig with valid Labels",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.StaticConfigs = []monitoringv1alpha1.StaticConfig{
					{
						Labels: map[string]string{"owner": "prometheus"},
					},
				}
			},
			valid: true,
		},
		{
			scenario:    "staticConfig with utf-8 label",
			promVersion: "3.0.0",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.StaticConfigs = []monitoringv1alpha1.StaticConfig{
					{
						Labels: map[string]string{"测试服务": "prometheus"},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "staticConfig with invalid utf-8 label",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.StaticConfigs = []monitoringv1alpha1.StaticConfig{
					{
						Labels: map[string]string{"\xff": "prometheus"},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "HTTP SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HTTPSDConfigs = []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://example.com",
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid:       false,
			promVersion: "2.29.0",
		},
		{
			scenario: "HTTP SD config with invalid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HTTPSDConfigs = []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://example.com",
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							ProxyFromEnvironment: ptr.To(true),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid:       false,
			promVersion: "2.29.0",
		},
		{
			scenario: "HTTP SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HTTPSDConfigs = []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://example.com",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid:       false,
			promVersion: "2.29.0",
		},
		{
			scenario: "HTTP SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HTTPSDConfigs = []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://example.com",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid:       false,
			promVersion: "2.29.0",
		},
		{
			scenario: "HTTP SD proxy config with invalid secret key",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HTTPSDConfigs = []monitoringv1alpha1.HTTPSDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "invalid-key",
									},
								},
							},
						},
					},
				}
			},
			valid:       false,
			promVersion: "2.29.0",
		},
		{
			scenario: "HTTP SD config in unsupported Prometheus version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HTTPSDConfigs = []monitoringv1alpha1.HTTPSDConfig{
					{
						URL: "http://example.com",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			promVersion: "2.27.0",
			valid:       false,
		},
		{
			scenario: "Kubernetes SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleNode,
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Kubernetes SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleNode,
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Kubernetes SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleNode,
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Kubernetes SD config with invalid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleNode,
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Kubernetes SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleNode,
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Kubernetes SD config with invalid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleNode,
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							ProxyFromEnvironment: ptr.To(true),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Kubernetes SD config with invalid label selector",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Selectors: []monitoringv1alpha1.K8SSelectorConfig{
							{
								Label: ptr.To("app=example,env!=production,release in (v1, v2)"),
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Kubernetes SD config with invalid field selector",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Selectors: []monitoringv1alpha1.K8SSelectorConfig{
							{
								Field: ptr.To("status.phase=Running,metadata.name!=worker"),
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Kubernetes SD config with valid Selector Role",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleNode,
						Selectors: []monitoringv1alpha1.K8SSelectorConfig{
							{
								Role: monitoringv1alpha1.KubernetesRoleNode,
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Kubernetes SD config with invalid Selector Role",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleNode,
						Selectors: []monitoringv1alpha1.K8SSelectorConfig{
							{
								Role: monitoringv1alpha1.KubernetesRolePod,
							},
						},
					},
				}

			},
			valid: false,
		},
		{
			scenario: "Kubernetes SD config with Role Pod",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRolePod,
					},
				}
			},
			promVersion: "2.51.0",
			valid:       true,
		},
		{
			scenario: "Kubernetes SD config with Role Pod but wrong version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRolePod,
					},
				}
			},
			promVersion: "2.31.0",
			valid:       false,
		},
		{
			scenario: "Kubernetes SD config with Role Endpoint",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleEndpoint,
					},
				}
			},
			promVersion: "2.51.0",
			valid:       true,
		},
		{
			scenario: "Kubernetes SD config with Role Endpoint but wrong version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleEndpoint,
					},
				}
			},
			promVersion: "2.31.0",
			valid:       false,
		},
		{
			scenario: "Kubernetes SD config with Role EndpointSlice",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleEndpointSlice,
					},
				}
			},
			promVersion: "2.51.0",
			valid:       true,
		},
		{
			scenario: "Kubernetes SD config with Role EndpointSlice but wrong version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleEndpointSlice,
					},
				}
			},
			promVersion: "2.31.0",
			valid:       false,
		},
		{
			scenario: "Kubernetes SD config with valid label and field selectors",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Role: monitoringv1alpha1.KubernetesRoleNode,
						Selectors: []monitoringv1alpha1.K8SSelectorConfig{
							{
								Role:  monitoringv1alpha1.KubernetesRoleNode,
								Label: ptr.To("app=example,env!=production,release in (v1, v2)"),
								Field: ptr.To("status.phase=Running,metadata.name!=worker"),
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Kubernetes SD config with only apiServer specified",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						APIServer: ptr.To("https://kube-api-server-address:6443"),
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Kubernetes SD config with only namespaces.ownNamespace specified",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						Namespaces: &monitoringv1alpha1.NamespaceDiscovery{
							IncludeOwnNamespace: ptr.To(true),
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Kubernetes SD config with both apiServer and namespaces.ownNamespace specified",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KubernetesSDConfigs = []monitoringv1alpha1.KubernetesSDConfig{
					{
						APIServer: ptr.To("https://kube-api-server-address:6443"),
						Namespaces: &monitoringv1alpha1.NamespaceDiscovery{
							IncludeOwnNamespace: ptr.To(true),
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Consul SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ConsulSDConfigs = []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "example.com",
						TokenRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Consul SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ConsulSDConfigs = []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "example.com",
						TokenRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "wrong",
							},
							Key: "key1",
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Consul SD config with no secret ref provided",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ConsulSDConfigs = []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "example.com",
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Consul SD config with filter",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ConsulSDConfigs = []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "example.com",
						Filter: ptr.To("Meta.env == \"qa\""),
					},
				}
			},
			valid:       true,
			promVersion: "3.0.0",
		},
		{
			scenario: "Consul SD config with filter but unsupported prometheus version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ConsulSDConfigs = []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "example.com",
						Filter: ptr.To("Meta.env == \"qa\""),
					},
				}
			},
			valid:       false,
			promVersion: "2.55.0",
		},
		{
			scenario: "Consul SD proxy config with invalid secret key",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ConsulSDConfigs = []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "example.com",
						TokenRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
					},
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "foo",
										},
										Key: "invalid-key",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Consul SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ConsulSDConfigs = []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "server",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Consul SD config with invalid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ConsulSDConfigs = []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "server",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Consul SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ConsulSDConfigs = []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "server",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Consul SD config with invalid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ConsulSDConfigs = []monitoringv1alpha1.ConsulSDConfig{
					{
						Server: "server",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "DNS SD config with no port specified for type other than SRV record",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DNSSDConfigs = []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"node.demo.do.prometheus.io"},
						Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeA),
					},
				}
			},
			promVersion: "2.51.0",
			valid:       false,
		},
		{
			scenario: "DNS SD config with MX record type and correct version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DNSSDConfigs = []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"node.demo.do.prometheus.io"},
						Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeMX),
						Port:  ptr.To(int32(9900)),
					},
				}
			},
			promVersion: "2.51.0",
			valid:       true,
		},
		{
			scenario: "DNS SD config with A record type and correct version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DNSSDConfigs = []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"node.demo.do.prometheus.io"},
						Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeA),
						Port:  ptr.To(int32(9900)),
					},
				}
			},
			promVersion: "2.51.0",
			valid:       true,
		},
		{
			scenario: "DNS SD config with port specified for type other than SRV record",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DNSSDConfigs = []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"node.demo.do.prometheus.io"},
						Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeA),
						Port:  ptr.To(int32(9900)),
					},
				}
			},
			valid: true,
		},
		{
			scenario: "DNS SD config with NS record type and correct version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DNSSDConfigs = []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"node.demo.do.prometheus.io"},
						Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeNS),
						Port:  ptr.To(int32(9900)),
					},
				}
			},
			promVersion: "2.51.0",
			valid:       true,
		},
		{
			scenario: "DNS SD config with MX record type and correct version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DNSSDConfigs = []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"node.demo.do.prometheus.io"},
						Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeMX),
						Port:  ptr.To(int32(9900)),
					},
				}
			},
			promVersion: "2.51.0",
			valid:       true,
		},
		{
			scenario: "DNS SD config with A record type and correct version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DNSSDConfigs = []monitoringv1alpha1.DNSSDConfig{
					{
						Names: []string{"node.demo.do.prometheus.io"},
						Type:  ptr.To(monitoringv1alpha1.DNSRecordTypeA),
						Port:  ptr.To(int32(9900)),
					},
				}
			},
			promVersion: "2.51.0",
			valid:       true,
		},
		{
			scenario: "EC2 SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EC2SDConfigs = []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
						AccessKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
						SecretKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key2",
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "EC2 SD config with no secret ref provided",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EC2SDConfigs = []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
					},
				}
			},
			valid: true,
		},
		{
			scenario: "EC2 SD config with invalid secret ref for secretKey",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EC2SDConfigs = []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
						AccessKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
						SecretKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "wrong",
							},
							Key: "key2",
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "EC2 SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EC2SDConfigs = []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "EC2 SD config with valid HTTPS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EC2SDConfigs = []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
						RefreshInterval: ptr.To(monitoringv1.Duration("30s")),
						EnableHTTP2:     ptr.To(true),
					},
				}
			},
			promVersion: "2.52.0",
			valid:       true,
		},
		{
			scenario: "EC2 SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EC2SDConfigs = []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "EC2 SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EC2SDConfigs = []monitoringv1alpha1.EC2SDConfig{
					{
						Region: ptr.To("us-east-1"),
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			promVersion: "2.52.0",
			valid:       true,
		},
		{
			scenario: "Azure SD config with valid options for OAuth authentication method",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						TenantID: ptr.To("BBBB222B-B2B2-2B22-B222-2BB2222BB2B2"),
						ClientID: ptr.To("333333CC-3C33-3333-CCC3-33C3CCCCC33C"),
						ClientSecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Azure SD config with no client secret ref provided for OAuth authentication method",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeOAuth),
						TenantID:             ptr.To("BBBB222B-B2B2-2B22-B222-2BB2222BB2B2"),
						ClientID:             ptr.To("333333CC-3C33-3333-CCC3-33C3CCCCC33C"),
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Azure SD config with no tenant id provided for OAuth authentication method",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeOAuth),
						ClientID:             ptr.To("333333CC-3C33-3333-CCC3-33C3CCCCC33C"),
						ClientSecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Azure SD config with no client id provided for OAuth authentication method",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeOAuth),
						TenantID:             ptr.To("BBBB222B-B2B2-2B22-B222-2BB2222BB2B2"),
						ClientSecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Azure SD config without options provided for ManagedIdentity authentication method",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeManagedIdentity),
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Azure SD config without options provided for SDK authentication method",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeSDK),
					},
				}
			},
			promVersion: "2.52.0",
			valid:       true,
		},
		{
			scenario: "Azure SD config with SDK authentication method but unsupported prometheus version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeSDK),
					},
				}
			},
			promVersion: "2.51.0",
			valid:       false,
		},
		{
			scenario: "Azure SD config with ResourceGroup and prometheus version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeManagedIdentity),
						ResourceGroup:        ptr.To("my-resource-group"),
					},
				}
			},
			promVersion: "2.51.0",
			valid:       true,
		},
		{
			scenario: "Azure SD config with ResourceGroup but unsupported prometheus version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						AuthenticationMethod: ptr.To(monitoringv1alpha1.AuthMethodTypeManagedIdentity),
						ResourceGroup:        ptr.To("my-resource-group"),
					},
				}
			},
			promVersion: "2.34.0",
			valid:       false,
		},
		{
			scenario: "Azure SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						SubscriptionID: "11111111-1111-1111-1111-111111111111",
						TenantID:       ptr.To("BBBB222B-B2B2-2B22-B222-2BB2222BB2B2"),
						ClientID:       ptr.To("333333CC-3C33-3333-CCC3-33C3CCCCC33C"),
						ClientSecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Azure SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.AzureSDConfigs = []monitoringv1alpha1.AzureSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "OpenStack SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OpenStackSDConfigs = []monitoringv1alpha1.OpenStackSDConfig{
					{
						Role:   monitoringv1alpha1.OpenStackRoleInstance,
						Region: "RegionOne",
						Password: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
						ApplicationCredentialSecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key2",
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "OpenStack SD config with invalid secret ref for password",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OpenStackSDConfigs = []monitoringv1alpha1.OpenStackSDConfig{
					{
						Password: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "invalid",
							},
							Key: "key1",
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "OpenStack SD config with invalid secret ref for application credentials",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OpenStackSDConfigs = []monitoringv1alpha1.OpenStackSDConfig{
					{
						ApplicationCredentialSecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key3",
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "OpenStack SD config with no secret ref provided",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OpenStackSDConfigs = []monitoringv1alpha1.OpenStackSDConfig{
					{
						Role:   monitoringv1alpha1.OpenStackRoleHypervisor,
						Region: "RegionTwo",
					},
				}
			},
			valid: true,
		},
		{
			scenario: "OpenStack SD config loadbalancer role in unsupported Prometheus version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OpenStackSDConfigs = []monitoringv1alpha1.OpenStackSDConfig{
					{
						Role:   monitoringv1alpha1.OpenStackRoleLoadBalancer,
						Region: "RegionTwo",
					},
				}
			},
			valid:       false,
			promVersion: "3.1.0",
		},
		{
			scenario: "OpenStack SD config loadbalancer role in supported Prometheus version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OpenStackSDConfigs = []monitoringv1alpha1.OpenStackSDConfig{
					{
						Role:   monitoringv1alpha1.OpenStackRoleLoadBalancer,
						Region: "RegionTwo",
					},
				}
			},
			valid:       true,
			promVersion: "3.2.0",
		},
		{
			scenario: "DigitalOcean SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DigitalOceanSDConfigs = []monitoringv1alpha1.DigitalOceanSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			promVersion: "2.40.0",
			valid:       true,
		},
		{
			scenario: "DigitalOcean SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DigitalOceanSDConfigs = []monitoringv1alpha1.DigitalOceanSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			promVersion: "2.40.0",
			valid:       false,
		},
		{
			scenario: "Digital Ocean SD config in unsupported Prometheus version",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DigitalOceanSDConfigs = []monitoringv1alpha1.DigitalOceanSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			promVersion: "2.11.0",
			valid:       false,
		},
		{
			scenario: "Kuma SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KumaSDConfigs = []monitoringv1alpha1.KumaSDConfig{
					{
						Server: "http://example.com",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Kuma SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KumaSDConfigs = []monitoringv1alpha1.KumaSDConfig{
					{
						Server: "http://example.com",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Kuma SD config with invalid server",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KumaSDConfigs = []monitoringv1alpha1.KumaSDConfig{
					{
						Server: "aaaaaa",
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Kuma SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KumaSDConfigs = []monitoringv1alpha1.KumaSDConfig{
					{
						Server: "http://example.com",
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Kuma SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.KumaSDConfigs = []monitoringv1alpha1.KumaSDConfig{
					{
						Server: "http://example.com",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Eureka SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EurekaSDConfigs = []monitoringv1alpha1.EurekaSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Eureka SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EurekaSDConfigs = []monitoringv1alpha1.EurekaSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Eureka SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EurekaSDConfigs = []monitoringv1alpha1.EurekaSDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Eureka SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.EurekaSDConfigs = []monitoringv1alpha1.EurekaSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},

		{
			scenario: "Docker SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSDConfigs = []monitoringv1alpha1.DockerSDConfig{
					{
						Host: "hostAddress",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Docker SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSDConfigs = []monitoringv1alpha1.DockerSDConfig{
					{
						Host: "hostAddress",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Docker SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSDConfigs = []monitoringv1alpha1.DockerSDConfig{
					{
						Host: "hostAddress",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Docker SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSDConfigs = []monitoringv1alpha1.DockerSDConfig{
					{
						Host: "hostAddress",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Linode SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LinodeSDConfigs = []monitoringv1alpha1.LinodeSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Linode SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LinodeSDConfigs = []monitoringv1alpha1.LinodeSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Linode SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LinodeSDConfigs = []monitoringv1alpha1.LinodeSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Linode SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LinodeSDConfigs = []monitoringv1alpha1.LinodeSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Hetzner SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HetznerSDConfigs = []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Hetzner SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HetznerSDConfigs = []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Hetzener SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HetznerSDConfigs = []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Hetzner SD config with invalid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HetznerSDConfigs = []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Hetzner SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HetznerSDConfigs = []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Hetzner SD config with invalid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.HetznerSDConfigs = []monitoringv1alpha1.HetznerSDConfig{
					{
						Role: "hcloud",
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							ProxyFromEnvironment: ptr.To(true),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Nomad SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.NomadSDConfigs = []monitoringv1alpha1.NomadSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Nomad SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.NomadSDConfigs = []monitoringv1alpha1.NomadSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Nomad SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.NomadSDConfigs = []monitoringv1alpha1.NomadSDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Nomad SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.NomadSDConfigs = []monitoringv1alpha1.NomadSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Dockerswarm SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSwarmSDConfigs = []monitoringv1alpha1.DockerSwarmSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Dockerswarm SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSwarmSDConfigs = []monitoringv1alpha1.DockerSwarmSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Dockerswarm SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSwarmSDConfigs = []monitoringv1alpha1.DockerSwarmSDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Dockerswarm SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.DockerSwarmSDConfigs = []monitoringv1alpha1.DockerSwarmSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "PuppetDB SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.PuppetDBSDConfigs = []monitoringv1alpha1.PuppetDBSDConfig{
					{
						URL: "https://example.com",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "PuppetDB SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.PuppetDBSDConfigs = []monitoringv1alpha1.PuppetDBSDConfig{
					{
						URL: "https://example.com",
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "PuppetDB SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.PuppetDBSDConfigs = []monitoringv1alpha1.PuppetDBSDConfig{
					{
						URL: "https://example.com",
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "PuppetDB SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.PuppetDBSDConfigs = []monitoringv1alpha1.PuppetDBSDConfig{
					{
						URL: "https://example.com",
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "PuppetDB SD config with invalid URL",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.PuppetDBSDConfigs = []monitoringv1alpha1.PuppetDBSDConfig{
					{
						URL: "www.percent-off.com",
					},
				}
			},
			valid: false,
		},
		{
			scenario: "LightSail SD config with valid TLS Config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "cert",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
							KeySecret: &v1.SecretKeySelector{
								Key: "key",
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "LightSail SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			promVersion: "3.7.0",
			valid:       false,
		},
		{
			scenario: "LightSail SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "LightSail SD config with invalid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							ProxyFromEnvironment: ptr.To(true),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "LightSail SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						Authorization: &monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},

		{
			scenario: "LightSail SD config with valid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						Region: ptr.To("us-east-1"),
						AccessKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
						SecretKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key2",
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "LightSail SD config with no secret ref provided",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						Region: ptr.To("us-east-1"),
					},
				}
			},
			valid: true,
		},
		{
			scenario: "LightSail SD config with invalid secret ref for accessKey",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						Region: ptr.To("us-east-1"),
						AccessKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "wrong",
							},
							Key: "key1",
						},
						SecretKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key2",
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "LightSail SD config with invalid secret ref for secretKey",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.LightSailSDConfigs = []monitoringv1alpha1.LightSailSDConfig{
					{
						Region: ptr.To("us-east-1"),
						AccessKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
						SecretKey: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "wrong",
							},
							Key: "key2",
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "OVHCloud SD config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OVHCloudSDConfigs = []monitoringv1alpha1.OVHCloudSDConfig{
					{
						ApplicationKey: "ApplicationKey",
						ApplicationSecret: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
						ConsumerKey: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key2",
						},
						Service:  monitoringv1alpha1.OVHServiceVPS,
						Endpoint: ptr.To("127.0.0.1"),
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Scaleway SD config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ScalewaySDConfigs = []monitoringv1alpha1.ScalewaySDConfig{
					{
						AccessKey: "AccessKey",
						SecretKey: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "secret",
							},
							Key: "key1",
						},
						ProjectID: "1",
						Role:      monitoringv1alpha1.ScalewayRoleInstance,

						Zone:       ptr.To("beijing-1"),
						Port:       ptr.To(int32(23456)),
						ApiURL:     ptr.To("https://api.scaleway.com/"),
						NameFilter: ptr.To("name"),
						TagsFilter: []string{"aa", "bb"},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Scaleway SD config with invalid secret ref for secretKey",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ScalewaySDConfigs = []monitoringv1alpha1.ScalewaySDConfig{
					{
						AccessKey: "AccessKey",
						SecretKey: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "wrong",
							},
							Key: "key1",
						},
						ProjectID: "1",
						Role:      monitoringv1alpha1.ScalewayRoleInstance,
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Scaleway SD config with invalid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ScalewaySDConfigs = []monitoringv1alpha1.ScalewaySDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							ProxyFromEnvironment: ptr.To(true),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Scaleway SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.ScalewaySDConfigs = []monitoringv1alpha1.ScalewaySDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									Key: "invalid_ca",
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Ionos SD config with valid TLS config",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.IonosSDConfigs = []monitoringv1alpha1.IonosSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
									Key: "ca",
								},
							},
							Cert: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
									Key: "cert",
								},
							},
							KeySecret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "secret",
								},
								Key: "key",
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Ionos SD config with invalid TLS config with invalid CA data",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.IonosSDConfigs = []monitoringv1alpha1.IonosSDConfig{
					{
						TLSConfig: &monitoringv1.SafeTLSConfig{
							CA: monitoringv1.SecretOrConfigMap{
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "secret",
									},
									Key: "invalid-ca",
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Ionos SD config with valid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.IonosSDConfigs = []monitoringv1alpha1.IonosSDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							NoProxy:              ptr.To("0.0.0.0"),
							ProxyFromEnvironment: ptr.To(false),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid: true,
		},
		{
			scenario: "Ionos SD config with invalid proxy settings",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.IonosSDConfigs = []monitoringv1alpha1.IonosSDConfig{
					{
						ProxyConfig: monitoringv1.ProxyConfig{
							ProxyURL:             ptr.To("http://no-proxy.com"),
							ProxyFromEnvironment: ptr.To(true),
							ProxyConnectHeader: map[string][]v1.SecretKeySelector{
								"header": {
									{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "secret",
										},
										Key: "key1",
									},
								},
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Ionos SD config with invalid secret ref",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.IonosSDConfigs = []monitoringv1alpha1.IonosSDConfig{
					{
						Authorization: monitoringv1.SafeAuthorization{
							Credentials: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "wrong",
								},
								Key: "key1",
							},
						},
					},
				}
			},
			valid: false,
		},
		{
			scenario: "Inexistent Scrape Class",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OpenStackSDConfigs = []monitoringv1alpha1.OpenStackSDConfig{
					{
						Role:   monitoringv1alpha1.OpenStackRoleHypervisor,
						Region: "RegionTwo",
					},
				}
			},
			valid:       false,
			scrapeClass: ptr.To("inexistent"),
		},
		{
			scenario: "inexistent Scrape Class",
			updateSpec: func(sc *monitoringv1alpha1.ScrapeConfigSpec) {
				sc.OpenStackSDConfigs = []monitoringv1alpha1.OpenStackSDConfig{
					{
						Role:   monitoringv1alpha1.OpenStackRoleHypervisor,
						Region: "RegionTwo",
					},
				}
			},
			valid:       true,
			scrapeClass: ptr.To("existent"),
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			cs := fake.NewSimpleClientset(
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "test",
					},
					Data: map[string][]byte{
						"key1":       []byte("val1"),
						"key2":       []byte("val2"),
						"ca":         ca,
						"invalid_ca": []byte("garbage"),
						"cert":       cert,
						"key":        key,
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "configmap",
						Namespace: "test",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
			)

			p := &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: tc.promVersion,
						ScrapeClasses: []monitoringv1.ScrapeClass{
							{
								Name: "existent",
							},
						},
					},
				},
			}
			rs, err := NewResourceSelector(
				newLogger(),
				p,
				assets.NewStoreBuilder(cs.CoreV1(), cs.CoreV1()),
				nil,
				operator.NewMetrics(prometheus.NewPedanticRegistry()),
				operator.NewFakeRecorder(1, p),
			)
			require.NoError(t, err)

			sc := &monitoringv1alpha1.ScrapeConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			}

			if tc.scrapeClass != nil {
				sc.Spec.ScrapeClassName = tc.scrapeClass
			}

			tc.updateSpec(&sc.Spec)

			sms, err := rs.SelectScrapeConfigs(context.Background(), func(_ string, _ labels.Selector, appendFn cache.AppendFunc) error {
				appendFn(sc)
				return nil
			})

			require.NoError(t, err)
			require.Len(t, sms, 1)

			valid := sms.ValidResources()
			if tc.valid {
				require.Len(t, valid, 1)
			} else {
				require.Empty(t, valid)
			}
		})
	}
}

func TestSelectPodMonitorsWithInvalidAuthentication(t *testing.T) {
	storeBuilder := assets.NewTestStoreBuilder(
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"secret": []byte("xxx"),
			},
		},
	)
	secretKey := v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: "foo",
		},
		Key: "secret",
	}

	for _, tc := range []struct {
		name       string
		updateFunc func(pe *monitoringv1.PodMetricsEndpoint)
	}{
		{
			name: "duplicate bearerTokenSecret and authorization",
			updateFunc: func(pe *monitoringv1.PodMetricsEndpoint) {
				pe.BearerTokenSecret = &secretKey
				pe.Authorization = &monitoringv1.SafeAuthorization{
					Credentials: &secretKey,
				}
			},
		},
		{
			name: "duplicate bearerTokenSecret and basicAuth",
			updateFunc: func(pe *monitoringv1.PodMetricsEndpoint) {
				pe.BearerTokenSecret = &secretKey
				pe.BasicAuth = &monitoringv1.BasicAuth{
					Username: secretKey,
					Password: secretKey,
				}
			},
		},
		{
			name: "duplicate bearerTokenSecret and oauth2",
			updateFunc: func(pe *monitoringv1.PodMetricsEndpoint) {
				pe.BearerTokenSecret = &secretKey
				pe.OAuth2 = &monitoringv1.OAuth2{
					ClientID: monitoringv1.SecretOrConfigMap{
						Secret: &secretKey,
					},
					ClientSecret: secretKey,
					TokenURL:     "http://example.com",
				}
			},
		},
		{
			name: "duplicate authorization and basicAuth",
			updateFunc: func(pe *monitoringv1.PodMetricsEndpoint) {
				pe.Authorization = &monitoringv1.SafeAuthorization{
					Credentials: &secretKey,
				}
				pe.BasicAuth = &monitoringv1.BasicAuth{
					Username: secretKey,
					Password: secretKey,
				}
			},
		},
		{
			name: "duplicate authorization and oauth2",
			updateFunc: func(pe *monitoringv1.PodMetricsEndpoint) {
				pe.Authorization = &monitoringv1.SafeAuthorization{
					Credentials: &secretKey,
				}
				pe.OAuth2 = &monitoringv1.OAuth2{
					ClientID: monitoringv1.SecretOrConfigMap{
						Secret: &secretKey,
					},
					ClientSecret: secretKey,
					TokenURL:     "http://example.com",
				}
			},
		},
		{
			name: "duplicate basicAuth and oauth2",
			updateFunc: func(pe *monitoringv1.PodMetricsEndpoint) {
				pe.BasicAuth = &monitoringv1.BasicAuth{
					Username: secretKey,
					Password: secretKey,
				}
				pe.OAuth2 = &monitoringv1.OAuth2{
					ClientID: monitoringv1.SecretOrConfigMap{
						Secret: &secretKey,
					},
					ClientSecret: secretKey,
					TokenURL:     "http://example.com",
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := defaultPrometheus()

			pme := monitoringv1.PodMetricsEndpoint{
				Port:     ptr.To("web"),
				Interval: "30s",
			}
			tc.updateFunc(&pme)
			pm := &monitoringv1.PodMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Labels: map[string]string{
						"group": "group1",
					},
				},
				Spec: monitoringv1.PodMonitorSpec{
					PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{pme},
				},
			}

			rs, err := NewResourceSelector(
				newLogger(),
				p,
				storeBuilder,
				nil,
				operator.NewMetrics(prometheus.NewPedanticRegistry()),
				operator.NewFakeRecorder(1, p),
			)
			require.NoError(t, err)

			pms, err := rs.SelectPodMonitors(context.Background(), func(_ string, _ labels.Selector, appendFn cache.AppendFunc) error {
				appendFn(pm)
				return nil
			})

			require.NoError(t, err)
			require.Len(t, pms, 1)

			valid := pms.ValidResources()
			require.Empty(t, valid)
		})
	}
}
