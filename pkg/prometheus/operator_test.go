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
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

func TestKeyToStatefulSetKey(t *testing.T) {
	cases := []struct {
		p        monitoringv1.PrometheusInterface
		name     string
		shard    int
		expected string
	}{
		{
			p:        &monitoringv1.Prometheus{},
			name:     "namespace/test",
			shard:    0,
			expected: "namespace/prometheus-test",
		},
		{
			p:        &monitoringv1alpha1.PrometheusAgent{},
			name:     "namespace/test",
			shard:    1,
			expected: "namespace/prom-agent-test-shard-1",
		},
	}

	for _, c := range cases {
		got := KeyToStatefulSetKey(c.p, c.name, c.shard)
		require.Equal(t, c.expected, got, "Expected key %q got %q", c.expected, got)
	}
}

func TestValidateRemoteWriteConfig(t *testing.T) {
	cases := []struct {
		name      string
		spec      monitoringv1.RemoteWriteSpec
		expectErr bool
		version   string
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
		{
			name: "with_no_azure_managed_identity_and_no_azure_oAuth_and_no_azure_sdk",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
				},
			},
			expectErr: true,
		},
		{
			name: "with_azure_managed_identity_and_azure_oAuth",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					ManagedIdentity: &monitoringv1.ManagedIdentity{
						ClientID: ptr.To("client-id"),
					},
					OAuth: &monitoringv1.AzureOAuth{
						TenantID: "00000000-a12b-3cd4-e56f-000000000000",
						ClientID: "00000000-0000-0000-0000-000000000000",
						ClientSecret: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "azure-oauth-secret",
							},
							Key: "secret-key",
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "with_azure_managed_identity_and_azure_sdk",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					ManagedIdentity: &monitoringv1.ManagedIdentity{
						ClientID: ptr.To("client-id"),
					},
					SDK: &monitoringv1.AzureSDK{
						TenantID: ptr.To("00000000-a12b-3cd4-e56f-000000000000"),
					},
				},
			},
			expectErr: true,
		},
		{
			name: "with_azure_managed_identity_empty_client_id",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					ManagedIdentity: &monitoringv1.ManagedIdentity{
						ClientID: ptr.To(""),
					},
				},
			},
			version:   "3.4.0",
			expectErr: true,
		},
		{
			name: "with_azure_managed_identity_empty_client_id_v3.5.0",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					ManagedIdentity: &monitoringv1.ManagedIdentity{
						ClientID: ptr.To(""),
					},
				},
			},
			version: "3.5.0",
		},
		{
			name: "with_azure_sdk_and_azure_oAuth",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					SDK: &monitoringv1.AzureSDK{
						TenantID: ptr.To("00000000-a12b-3cd4-e56f-000000000000"),
					},
					OAuth: &monitoringv1.AzureOAuth{
						TenantID: "00000000-a12b-3cd4-e56f-000000000000",
						ClientID: "00000000-0000-0000-0000-000000000000",
						ClientSecret: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "azure-oauth-secret",
							},
							Key: "secret-key",
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "with_invalid_azure_oAuth_clientID",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					OAuth: &monitoringv1.AzureOAuth{
						TenantID: "00000000-a12b-3cd4-e56f-000000000000",
						ClientID: "invalid",
						ClientSecret: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "azure-oauth-secret",
							},
							Key: "secret-key",
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "rw_azuread_with_workload_identity",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					WorkloadIdentity: &monitoringv1.AzureWorkloadIdentity{
						ClientID: "00000000-a12b-3cd4-e56f-000000000000",
						TenantID: "11111111-a12b-3cd4-e56f-000000000000",
					},
				},
			},
		},
		{
			name: "with_invalid_workload_identity_clientID",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					WorkloadIdentity: &monitoringv1.AzureWorkloadIdentity{
						ClientID: "invalid-uuid",
						TenantID: "11111111-a12b-3cd4-e56f-000000000000",
					},
				},
			},
			expectErr: true,
		},
		{
			name: "with_invalid_workload_identity_tenantID",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					WorkloadIdentity: &monitoringv1.AzureWorkloadIdentity{
						ClientID: "00000000-a12b-3cd4-e56f-000000000000",
						TenantID: "invalid-uuid",
					},
				},
			},
			expectErr: true,
		},
		{
			name: "with_workload_identity_and_managed_identity",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					ManagedIdentity: &monitoringv1.ManagedIdentity{
						ClientID: ptr.To("00000000-a12b-3cd4-e56f-000000000000"),
					},
					WorkloadIdentity: &monitoringv1.AzureWorkloadIdentity{
						ClientID: "00000000-a12b-3cd4-e56f-000000000000",
						TenantID: "11111111-a12b-3cd4-e56f-000000000000",
					},
				},
			},
			expectErr: true,
		},
		{
			name: "with_workload_identity_and_oauth",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					OAuth: &monitoringv1.AzureOAuth{
						TenantID: "00000000-a12b-3cd4-e56f-000000000000",
						ClientID: "00000000-0000-0000-0000-000000000000",
						ClientSecret: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "azure-oauth-secret",
							},
							Key: "secret-key",
						},
					},
					WorkloadIdentity: &monitoringv1.AzureWorkloadIdentity{
						ClientID: "00000000-a12b-3cd4-e56f-000000000000",
						TenantID: "11111111-a12b-3cd4-e56f-000000000000",
					},
				},
			},
			expectErr: true,
		},
		{
			name: "with_workload_identity_and_sdk",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					SDK: &monitoringv1.AzureSDK{
						TenantID: ptr.To("00000000-a12b-3cd4-e56f-000000000000"),
					},
					WorkloadIdentity: &monitoringv1.AzureWorkloadIdentity{
						ClientID: "00000000-a12b-3cd4-e56f-000000000000",
						TenantID: "11111111-a12b-3cd4-e56f-000000000000",
					},
				},
			},
			expectErr: true,
		},
		{
			name: "with_no_azure_auth_method_including_workload_identity",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
				},
			},
			expectErr: true,
		},
		{
			name:    "with_workload_identity_unsupported_version",
			version: "v3.6.0",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					WorkloadIdentity: &monitoringv1.AzureWorkloadIdentity{
						ClientID: "00000000-a12b-3cd4-e56f-000000000000",
						TenantID: "11111111-a12b-3cd4-e56f-000000000000",
					},
				},
			},
			expectErr: true,
		},
		{
			name:    "with_oauth_unsupported_version",
			version: "v2.47.0",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					OAuth: &monitoringv1.AzureOAuth{
						TenantID: "00000000-a12b-3cd4-e56f-000000000000",
						ClientID: "00000000-0000-0000-0000-000000000000",
						ClientSecret: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "azure-oauth-secret",
							},
							Key: "secret-key",
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name:    "with_sdk_unsupported_version",
			version: "v2.51.0",
			spec: monitoringv1.RemoteWriteSpec{
				URL: "http://example.com",
				AzureAD: &monitoringv1.AzureAD{
					Cloud: ptr.To("AzureGovernment"),
					SDK: &monitoringv1.AzureSDK{
						TenantID: ptr.To("00000000-a12b-3cd4-e56f-000000000000"),
					},
				},
			},
			expectErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "test",
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Version: tc.version,
					},
				},
			}
			cg := mustNewConfigGenerator(t, p)

			err := cg.validateRemoteWriteSpec(tc.spec)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
