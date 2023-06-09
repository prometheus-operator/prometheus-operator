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

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

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
		match, key := StatefulSetKeyToPrometheusKey(c.input)
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
