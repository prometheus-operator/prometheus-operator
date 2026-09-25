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

	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

func TestValidateAlertmanagerConfigWithGlobal(t *testing.T) {
	t.Parallel()

	global := &monitoringv1alpha1.AlertmanagerConfig{
		Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
			MuteTimeIntervals: []monitoringv1alpha1.MuteTimeInterval{
				{
					Name: "public_holidays",
					TimeIntervals: []monitoringv1alpha1.TimeInterval{
						{
							Months: []monitoringv1alpha1.MonthRange{"January"},
						},
					},
				},
			},
		},
	}

	for _, tc := range []struct {
		name      string
		in        *monitoringv1alpha1.AlertmanagerConfig
		global    *monitoringv1alpha1.AlertmanagerConfig
		expectErr bool
	}{
		{
			name: "route references mute time interval from global config",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{Name: "default"},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver:          "default",
						MuteTimeIntervals: []string{"public_holidays"},
					},
				},
			},
			global: global,
		},
		{
			name: "route references mute time interval missing from local and global config",
			in: &monitoringv1alpha1.AlertmanagerConfig{
				Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
					Receivers: []monitoringv1alpha1.Receiver{
						{Name: "default"},
					},
					Route: &monitoringv1alpha1.Route{
						Receiver:          "default",
						MuteTimeIntervals: []string{"public_holidays"},
					},
				},
			},
			global:    nil,
			expectErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateAlertmanagerConfigWithGlobal(tc.in, tc.global)
			if tc.expectErr && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tc.expectErr && err != nil {
				t.Fatalf("expected no error but got: %v", err)
			}
		})
	}
}
