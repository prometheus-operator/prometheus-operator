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

package prometheus

import (
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// TestBuildProbesWithClientAuth verifies that when TLS is configured with
// client authentication, the probes use HTTP instead of HTTPS to avoid
// kubelet failing to authenticate the probes.
func TestBuildProbesWithClientAuth(t *testing.T) {
	cpf := &monitoringv1.CommonPrometheusFields{
		Web: &monitoringv1.PrometheusWebSpec{
			WebConfigFileFields: monitoringv1.WebConfigFileFields{
				TLSConfig: &monitoringv1.WebTLSConfig{
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test-prometheus-serving-certs",
							},
							Key: "tls.crt",
						},
					},
					KeySecret: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-prometheus-serving-certs",
						},
						Key: "tls.key",
					},
					ClientCA: monitoringv1.SecretOrConfigMap{
						Secret: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test-prometheus-client-ca",
							},
							Key: "ca.crt",
						},
					},
					ClientAuthType: ptr.To("RequireAndVerifyClientCert"),
				},
			},
		},
	}

	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: *cpf,
		},
	}

	cg, err := NewConfigGenerator(nil, p)
	require.NoError(t, err)

	// Build probes
	startupProbe, readinessProbe, livenessProbe := cg.BuildProbes()

	// Verify that all probes use HTTP scheme when client auth is required
	// This prevents kubelet from failing to authenticate the probes
	for _, probe := range []*corev1.Probe{startupProbe, readinessProbe, livenessProbe} {
		require.NotNil(t, probe.HTTPGet, "Probe must have HTTPGet defined")
		require.Equal(t, corev1.URISchemeHTTP, probe.HTTPGet.Scheme,
			"Probe scheme should be HTTP when client auth is required, not HTTPS")
	}
}

// TestBuildProbesWithTLSWithoutClientAuth verifies that when TLS is configured
// without client authentication, the probes correctly use HTTPS.
func TestBuildProbesWithTLSWithoutClientAuth(t *testing.T) {
	cpf := &monitoringv1.CommonPrometheusFields{
		Web: &monitoringv1.PrometheusWebSpec{
			WebConfigFileFields: monitoringv1.WebConfigFileFields{
				TLSConfig: &monitoringv1.WebTLSConfig{
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test-prometheus-serving-certs",
							},
							Key: "tls.crt",
						},
					},
					KeySecret: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-prometheus-serving-certs",
						},
						Key: "tls.key",
					},
				},
			},
		},
	}

	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: *cpf,
		},
	}

	cg, err := NewConfigGenerator(nil, p)
	require.NoError(t, err)

	// Build probes
	startupProbe, readinessProbe, livenessProbe := cg.BuildProbes()

	// Verify that all probes use HTTPS scheme when TLS is configured without client auth
	for _, probe := range []*corev1.Probe{startupProbe, readinessProbe, livenessProbe} {
		require.NotNil(t, probe.HTTPGet, "Probe must have HTTPGet defined")
		require.Equal(t, corev1.URISchemeHTTPS, probe.HTTPGet.Scheme,
			"Probe scheme should be HTTPS when TLS is configured without client auth")
	}
}

// TestBuildProbesWithoutTLS verifies that when no TLS is configured, the
// probes correctly use HTTP scheme.
func TestBuildProbesWithoutTLS(t *testing.T) {
	p := &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{},
		},
	}

	cg, err := NewConfigGenerator(nil, p)
	require.NoError(t, err)

	// Build probes
	startupProbe, readinessProbe, livenessProbe := cg.BuildProbes()

	// Verify that all probes use HTTP scheme when no TLS is configured
	for _, probe := range []*corev1.Probe{startupProbe, readinessProbe, livenessProbe} {
		require.NotNil(t, probe.HTTPGet, "Probe must have HTTPGet defined")
		require.Equal(t, corev1.URISchemeHTTP, probe.HTTPGet.Scheme,
			"Probe scheme should be HTTP when no TLS is configured")
	}
}
