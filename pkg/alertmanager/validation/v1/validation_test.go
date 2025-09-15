// Copyright 2021 The prometheus-operator Authors
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

package v1

import (
	"testing"

	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestValidateAlertmanager(t *testing.T) {
	testCases := []struct {
		name      string
		in        *monitoringv1.Alertmanager
		expectErr bool
	}{
		{
			name: "Test HTTP Config both environment and url defined",
			in: &monitoringv1.Alertmanager{
				Spec: monitoringv1.AlertmanagerSpec{
					AlertmanagerConfiguration: &monitoringv1.AlertmanagerConfiguration{
						Global: &monitoringv1.AlertmanagerGlobalConfig{
							HTTPConfig: &monitoringv1.HTTPConfig{
								ProxyConfig: monitoringv1.ProxyConfig{
									ProxyFromEnvironment: ptr.To(true),
									ProxyURL:             ptr.To("https://proxy.example.com"),
								},
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Test HTTP Config both noproxy and url defined",
			in: &monitoringv1.Alertmanager{
				Spec: monitoringv1.AlertmanagerSpec{
					AlertmanagerConfiguration: &monitoringv1.AlertmanagerConfiguration{
						Global: &monitoringv1.AlertmanagerGlobalConfig{
							HTTPConfig: &monitoringv1.HTTPConfig{
								ProxyConfig: monitoringv1.ProxyConfig{
									NoProxy:  ptr.To("https://noproxy.example.com"),
									ProxyURL: ptr.To("https://proxy.example.com"),
								},
							},
						},
					},
				},
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAlertmanager(tc.in)
			if tc.expectErr && err == nil {
				t.Error("expected error but got none")
			}

			if err != nil {
				if tc.expectErr {
					return
				}
				t.Errorf("got error but expected none -%s", err.Error())
			}
		})
	}
}
