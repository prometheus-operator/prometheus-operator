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
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
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
						ClientSecret: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
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
						ClientSecret: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
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
						ClientSecret: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
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
						ClientSecret: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
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
						ClientSecret: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
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

type fakeStatefulSetGetter []appsv1.StatefulSet

func (ssg fakeStatefulSetGetter) Get(key string) (runtime.Object, error) {
	for _, sset := range ssg {
		if key == fmt.Sprintf("%s/%s", sset.Namespace, sset.Name) {
			return &sset, nil
		}
	}

	return nil, apierrors.NewNotFound(schema.GroupResource{}, "")
}

type fakeReconciledConditionGetter struct{}

func (rcg *fakeReconciledConditionGetter) GetCondition(_ string, n int64) monitoringv1.Condition {
	return monitoringv1.Condition{
		Type:   monitoringv1.Reconciled,
		Status: monitoringv1.ConditionTrue,
		LastTransitionTime: metav1.Time{
			Time: time.Now().UTC(),
		},
		ObservedGeneration: n,
	}
}

type fakeDeletionChecker struct{}

func (dc *fakeDeletionChecker) DeletionInProgress(_ metav1.Object) bool { return false }

func fakeStatefulSet(name string) appsv1.StatefulSet {
	return appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  "ns",
			Generation: 45,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{},
		},
		Status: appsv1.StatefulSetStatus{
			UpdateRevision: name + "-ffffffff",
		},
	}
}

func fakeReadyPod(sts string, ordinal int, ready bool) corev1.Pod {
	status := corev1.ConditionFalse
	if ready {
		status = corev1.ConditionTrue
	}
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:       fmt.Sprintf("%s-%d", sts, ordinal),
			Namespace:  "ns",
			Generation: 47,
			Labels: map[string]string{
				"controller-revision-hash": sts + "-ffffffff",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					Name: sts,
					Kind: "StatefulSet",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{{
				Type:   corev1.PodReady,
				Status: status,
			}},
		},
	}
}

func TestCombinedStatus(t *testing.T) {
	for _, tc := range []struct {
		name     string
		statuses []monitoringv1.ConditionStatus
		exp      monitoringv1.ConditionStatus
	}{
		{
			name:     "nil slice",
			statuses: nil,
			exp:      monitoringv1.ConditionTrue,
		},
		{
			name:     "empty slice",
			statuses: []monitoringv1.ConditionStatus{},
			exp:      monitoringv1.ConditionTrue,
		},
		{
			name:     "all True",
			statuses: []monitoringv1.ConditionStatus{monitoringv1.ConditionTrue, monitoringv1.ConditionTrue},
			exp:      monitoringv1.ConditionTrue,
		},
		{
			name:     "single False",
			statuses: []monitoringv1.ConditionStatus{monitoringv1.ConditionFalse},
			exp:      monitoringv1.ConditionFalse,
		},
		{
			name:     "False short-circuits remaining statuses",
			statuses: []monitoringv1.ConditionStatus{monitoringv1.ConditionTrue, monitoringv1.ConditionFalse, monitoringv1.ConditionDegraded},
			exp:      monitoringv1.ConditionFalse,
		},
		{
			name:     "single Degraded",
			statuses: []monitoringv1.ConditionStatus{monitoringv1.ConditionDegraded},
			exp:      monitoringv1.ConditionDegraded,
		},
		{
			name:     "Degraded with True",
			statuses: []monitoringv1.ConditionStatus{monitoringv1.ConditionTrue, monitoringv1.ConditionDegraded},
			exp:      monitoringv1.ConditionDegraded,
		},
		{
			name:     "single Unknown",
			statuses: []monitoringv1.ConditionStatus{monitoringv1.ConditionUnknown},
			exp:      monitoringv1.ConditionUnknown,
		},
		{
			name:     "Unknown with True",
			statuses: []monitoringv1.ConditionStatus{monitoringv1.ConditionTrue, monitoringv1.ConditionUnknown},
			exp:      monitoringv1.ConditionUnknown,
		},
		{
			name:     "Degraded overrides Unknown",
			statuses: []monitoringv1.ConditionStatus{monitoringv1.ConditionUnknown, monitoringv1.ConditionDegraded},
			exp:      monitoringv1.ConditionDegraded,
		},
		{
			name:     "Degraded then Unknown stays Degraded",
			statuses: []monitoringv1.ConditionStatus{monitoringv1.ConditionDegraded, monitoringv1.ConditionUnknown},
			exp:      monitoringv1.ConditionDegraded,
		},
		{
			name:     "False takes priority over Degraded",
			statuses: []monitoringv1.ConditionStatus{monitoringv1.ConditionDegraded, monitoringv1.ConditionFalse},
			exp:      monitoringv1.ConditionFalse,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.exp, combinedStatus(tc.statuses))
		})
	}
}

func TestCombinedReason(t *testing.T) {
	for _, tc := range []struct {
		name    string
		reasons []string
		exp     string
	}{
		{
			name:    "nil slice",
			reasons: nil,
			exp:     "",
		},
		{
			name:    "empty slice",
			reasons: []string{},
			exp:     "",
		},
		{
			name:    "all empty strings",
			reasons: []string{"", ""},
			exp:     "",
		},
		{
			name:    "single reason",
			reasons: []string{"ReasonA"},
			exp:     "ReasonA",
		},
		{
			name:    "duplicate reasons",
			reasons: []string{"ReasonA", "ReasonA"},
			exp:     "ReasonA",
		},
		{
			name:    "two distinct reasons joined alphabetically",
			reasons: []string{"ReasonB", "ReasonA"},
			exp:     "ReasonAAndReasonB",
		},
		{
			name:    "empty strings are ignored",
			reasons: []string{"", "ReasonA", ""},
			exp:     "ReasonA",
		},
		{
			name:    "distinct reasons with duplicates and empty strings",
			reasons: []string{"ReasonB", "", "ReasonA", "ReasonB"},
			exp:     "ReasonAAndReasonB",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.exp, combinedReason(tc.reasons))
		})
	}
}

func TestStatefulSetReporterProcess(t *testing.T) {
	for _, tc := range []struct {
		name  string
		p     monitoringv1.Prometheus
		ssets []appsv1.StatefulSet
		pods  []corev1.Pod
		exp   *monitoringv1.PrometheusStatus
	}{
		{
			name: "prometheus with (replicas=1,shards=1) and no statefulset",
			p: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test",
					Namespace:  "ns",
					Generation: 42,
				},
				Spec: monitoringv1.PrometheusSpec{},
			},
			exp: &monitoringv1.PrometheusStatus{
				Conditions: []monitoringv1.Condition{
					{
						Type:               monitoringv1.Available,
						Status:             monitoringv1.ConditionFalse,
						Reason:             "StatefulSetNotFound",
						ObservedGeneration: 42,
					},
					{
						Type:               monitoringv1.Reconciled,
						Status:             monitoringv1.ConditionTrue,
						ObservedGeneration: 42,
					},
				},
				ShardStatuses: []monitoringv1.ShardStatus{
					{
						ShardID: "0",
					},
				},
			},
		},
		{
			name: "prometheus with (replicas=1,shards=1) and no pods",
			p: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test",
					Namespace:  "ns",
					Generation: 42,
				},
				Spec: monitoringv1.PrometheusSpec{},
			},
			ssets: []appsv1.StatefulSet{
				fakeStatefulSet("prometheus-test"),
			},
			exp: &monitoringv1.PrometheusStatus{
				Conditions: []monitoringv1.Condition{
					{
						Type:               monitoringv1.Available,
						Status:             monitoringv1.ConditionFalse,
						Reason:             "NoPodReady",
						ObservedGeneration: 42,
					},
					{
						Type:               monitoringv1.Reconciled,
						Status:             monitoringv1.ConditionTrue,
						ObservedGeneration: 42,
					},
				},
				ShardStatuses: []monitoringv1.ShardStatus{
					{
						ShardID: "0",
					},
				},
			},
		},
		{
			name: "prometheus with (replicas=1,shards=1) with no ready pod",
			p: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test",
					Namespace:  "ns",
					Generation: 42,
				},
				Spec: monitoringv1.PrometheusSpec{},
			},
			ssets: []appsv1.StatefulSet{
				fakeStatefulSet("prometheus-test"),
			},
			pods: []corev1.Pod{
				fakeReadyPod("prometheus-test", 0, false),
			},
			exp: &monitoringv1.PrometheusStatus{
				Conditions: []monitoringv1.Condition{
					{
						Type:               monitoringv1.Available,
						Status:             monitoringv1.ConditionFalse,
						Reason:             "NoPodReady",
						ObservedGeneration: 42,
					},
					{
						Type:               monitoringv1.Reconciled,
						Status:             monitoringv1.ConditionTrue,
						ObservedGeneration: 42,
					},
				},
				ShardStatuses: []monitoringv1.ShardStatus{
					{
						ShardID:             "0",
						Replicas:            1,
						UpdatedReplicas:     1,
						UnavailableReplicas: 1,
					},
				},
				Replicas:            1,
				UpdatedReplicas:     1,
				UnavailableReplicas: 1,
			},
		},
		{
			name: "prometheus with (replicas=1,shards=1) with ready pod",
			p: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test",
					Namespace:  "ns",
					Generation: 42,
				},
				Spec: monitoringv1.PrometheusSpec{},
			},
			ssets: []appsv1.StatefulSet{
				fakeStatefulSet("prometheus-test"),
			},
			pods: []corev1.Pod{
				fakeReadyPod("prometheus-test", 0, true),
			},
			exp: &monitoringv1.PrometheusStatus{
				Conditions: []monitoringv1.Condition{
					{
						Type:               monitoringv1.Available,
						Status:             monitoringv1.ConditionTrue,
						Reason:             "",
						ObservedGeneration: 42,
					},
					{
						Type:               monitoringv1.Reconciled,
						Status:             monitoringv1.ConditionTrue,
						ObservedGeneration: 42,
					},
				},
				ShardStatuses: []monitoringv1.ShardStatus{
					{
						ShardID:           "0",
						Replicas:          1,
						UpdatedReplicas:   1,
						AvailableReplicas: 1,
					},
				},
				Replicas:          1,
				UpdatedReplicas:   1,
				AvailableReplicas: 1,
			},
		},
		{
			name: "prometheus with (replicas=2,shards=1) with ready and not ready pods",
			p: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test",
					Namespace:  "ns",
					Generation: 42,
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Replicas: ptr.To(int32(2)),
					},
				},
			},
			ssets: []appsv1.StatefulSet{
				fakeStatefulSet("prometheus-test"),
			},
			pods: []corev1.Pod{
				fakeReadyPod("prometheus-test", 0, true),
				fakeReadyPod("prometheus-test", 1, false),
			},
			exp: &monitoringv1.PrometheusStatus{
				Conditions: []monitoringv1.Condition{
					{
						Type:               monitoringv1.Available,
						Status:             monitoringv1.ConditionDegraded,
						Reason:             "SomePodsNotReady",
						ObservedGeneration: 42,
					},
					{
						Type:               monitoringv1.Reconciled,
						Status:             monitoringv1.ConditionTrue,
						ObservedGeneration: 42,
					},
				},
				ShardStatuses: []monitoringv1.ShardStatus{
					{
						ShardID:             "0",
						Replicas:            2,
						UpdatedReplicas:     2,
						AvailableReplicas:   1,
						UnavailableReplicas: 1,
					},
				},
				Replicas:            2,
				UpdatedReplicas:     2,
				AvailableReplicas:   1,
				UnavailableReplicas: 1,
			},
		},
		{
			name: "prometheus with (replicas=1,shards=2) with ready pods",
			p: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test",
					Namespace:  "ns",
					Generation: 42,
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Replicas: ptr.To(int32(1)),
						Shards:   ptr.To(int32(2)),
					},
				},
			},
			ssets: []appsv1.StatefulSet{
				fakeStatefulSet("prometheus-test"),
				fakeStatefulSet("prometheus-test-shard-1"),
			},
			pods: []corev1.Pod{
				fakeReadyPod("prometheus-test", 0, true),
				fakeReadyPod("prometheus-test-shard-1", 0, true),
			},
			exp: &monitoringv1.PrometheusStatus{
				Conditions: []monitoringv1.Condition{
					{
						Type:               monitoringv1.Available,
						Status:             monitoringv1.ConditionTrue,
						Reason:             "",
						ObservedGeneration: 42,
					},
					{
						Type:               monitoringv1.Reconciled,
						Status:             monitoringv1.ConditionTrue,
						ObservedGeneration: 42,
					},
				},
				ShardStatuses: []monitoringv1.ShardStatus{
					{
						ShardID:           "0",
						Replicas:          1,
						UpdatedReplicas:   1,
						AvailableReplicas: 1,
					},
					{
						ShardID:           "1",
						Replicas:          1,
						UpdatedReplicas:   1,
						AvailableReplicas: 1,
					},
				},
				Replicas:          2,
				UpdatedReplicas:   2,
				AvailableReplicas: 2,
			},
		},
		{
			name: "prometheus with (replicas=1,shards=2) with ready and not ready pods",
			p: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test",
					Namespace:  "ns",
					Generation: 42,
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Replicas: ptr.To(int32(1)),
						Shards:   ptr.To(int32(2)),
					},
				},
			},
			ssets: []appsv1.StatefulSet{
				fakeStatefulSet("prometheus-test"),
				fakeStatefulSet("prometheus-test-shard-1"),
			},
			pods: []corev1.Pod{
				fakeReadyPod("prometheus-test", 0, false),
				fakeReadyPod("prometheus-test-shard-1", 0, true),
			},
			exp: &monitoringv1.PrometheusStatus{
				Conditions: []monitoringv1.Condition{
					{
						Type:               monitoringv1.Available,
						Status:             monitoringv1.ConditionFalse,
						Reason:             "NoPodReady",
						ObservedGeneration: 42,
					},
					{
						Type:               monitoringv1.Reconciled,
						Status:             monitoringv1.ConditionTrue,
						ObservedGeneration: 42,
					},
				},
				ShardStatuses: []monitoringv1.ShardStatus{
					{
						ShardID:             "0",
						Replicas:            1,
						UpdatedReplicas:     1,
						UnavailableReplicas: 1,
					},
					{
						ShardID:           "1",
						Replicas:          1,
						UpdatedReplicas:   1,
						AvailableReplicas: 1,
					},
				},
				Replicas:            2,
				UpdatedReplicas:     2,
				AvailableReplicas:   1,
				UnavailableReplicas: 1,
			},
		},
		{
			name: "prometheus with (replicas=2,shards=2) with ready and not ready pods",
			p: monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test",
					Namespace:  "ns",
					Generation: 42,
				},
				Spec: monitoringv1.PrometheusSpec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						Replicas: ptr.To(int32(2)),
						Shards:   ptr.To(int32(2)),
					},
				},
			},
			ssets: []appsv1.StatefulSet{
				fakeStatefulSet("prometheus-test"),
				fakeStatefulSet("prometheus-test-shard-1"),
			},
			pods: []corev1.Pod{
				fakeReadyPod("prometheus-test", 0, false),
				fakeReadyPod("prometheus-test", 1, true),
				fakeReadyPod("prometheus-test-shard-1", 0, true),
				fakeReadyPod("prometheus-test-shard-1", 1, true),
			},
			exp: &monitoringv1.PrometheusStatus{
				Conditions: []monitoringv1.Condition{
					{
						Type:               monitoringv1.Available,
						Status:             monitoringv1.ConditionDegraded,
						Reason:             "SomePodsNotReady",
						ObservedGeneration: 42,
					},
					{
						Type:               monitoringv1.Reconciled,
						Status:             monitoringv1.ConditionTrue,
						ObservedGeneration: 42,
					},
				},
				ShardStatuses: []monitoringv1.ShardStatus{
					{
						ShardID:             "0",
						Replicas:            2,
						UpdatedReplicas:     2,
						AvailableReplicas:   1,
						UnavailableReplicas: 1,
					},
					{
						ShardID:           "1",
						Replicas:          2,
						UpdatedReplicas:   2,
						AvailableReplicas: 2,
					},
				},
				Replicas:            4,
				UpdatedReplicas:     4,
				AvailableReplicas:   3,
				UnavailableReplicas: 1,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := fake.NewClientset()
			for _, pod := range tc.pods {
				c.Tracker().Add(&pod)
			}

			sr := NewStatusReporter(
				c,
				&fakeReconciledConditionGetter{},
				fakeStatefulSetGetter(tc.ssets),
				&fakeDeletionChecker{},
				operator.NoneRepairPolicy,
			)

			logger := slog.New(slog.DiscardHandler)
			status, err := sr.Process(context.Background(), logger, &tc.p, fmt.Sprintf("%s/%s", tc.p.Namespace, tc.p.Name))
			if tc.exp == nil {
				require.Error(t, err)
				return
			}

			for i := range status.Conditions {
				status.Conditions[i].LastTransitionTime = metav1.NewTime(time.Time{})
				status.Conditions[i].Message = ""
			}

			require.Equal(t, tc.exp, status)
		})
	}
}
