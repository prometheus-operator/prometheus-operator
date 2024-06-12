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

	v1 "k8s.io/api/core/v1"
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
						ClientID: "client-id",
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
						ClientID: "client-id",
					},
					SDK: &monitoringv1.AzureSDK{
						TenantID: ptr.To("00000000-a12b-3cd4-e56f-000000000000"),
					},
				},
			},
			expectErr: true,
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
	}
	for _, c := range cases {
		test := c
		t.Run(test.name, func(t *testing.T) {
			err := ValidateRemoteWriteSpec(test.spec)
			if err != nil && !test.expectErr {
				t.Fatalf("unexpected error occurred: %v", err)
			}
			if err == nil && test.expectErr {
				t.Fatalf("expected an error, got nil")
			}
		})
	}
}
