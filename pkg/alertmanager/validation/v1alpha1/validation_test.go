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

package v1alpha1

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

func TestValidateWebexConfigsRequireAuthorization(t *testing.T) {
	withAuthorization := &monitoringv1alpha1.HTTPConfig{
		Authorization: &monitoringv1.SafeAuthorization{
			Credentials: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "webex"},
				Key:                  "token",
			},
		},
	}

	for _, tc := range []struct {
		name      string
		config    monitoringv1alpha1.WebexConfig
		expectErr bool
	}{
		{
			name:      "no httpConfig",
			config:    monitoringv1alpha1.WebexConfig{RoomID: "room-1"},
			expectErr: true,
		},
		{
			name:      "httpConfig without authorization",
			config:    monitoringv1alpha1.WebexConfig{RoomID: "room-1", HTTPConfig: &monitoringv1alpha1.HTTPConfig{}},
			expectErr: true,
		},
		{
			name:   "httpConfig with authorization",
			config: monitoringv1alpha1.WebexConfig{RoomID: "room-1", HTTPConfig: withAuthorization},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateWebexConfigs([]monitoringv1alpha1.WebexConfig{tc.config})
			if tc.expectErr && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tc.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
