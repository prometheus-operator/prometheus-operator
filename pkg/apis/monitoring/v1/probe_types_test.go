// Copyright 2018 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateProbeTargets(t *testing.T) {
	tests := []struct {
		name         string
		probeTargets ProbeTargets
		wantErr      bool
	}{

		{
			name: "probe with static config target",
			probeTargets: ProbeTargets{
				StaticConfig: &ProbeTargetStaticConfig{
					Targets: []string{"/probe"},
					Labels:  map[string]string{"app": "foo"},
				},
			},
			wantErr: false,
		},
		{
			name: "probe with ingress target",
			probeTargets: ProbeTargets{
				Ingress: &ProbeTargetIngress{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "foo",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "one of staticConfig and ingress is required",
			probeTargets: ProbeTargets{
				StaticConfig: nil,
				Ingress:      nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.probeTargets.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
