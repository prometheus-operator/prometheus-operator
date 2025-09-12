// Copyright 2024 The prometheus-operator Authors
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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestCommonPrometheusFields_ValidateResources(t *testing.T) {
	tests := []struct {
		name      string
		resources v1.ResourceRequirements
		wantErr   bool
	}{
		{
			name: "valid resources",
			resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("100m"),
					v1.ResourceMemory: resource.MustParse("128Mi"),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("200m"),
					v1.ResourceMemory: resource.MustParse("256Mi"),
				},
			},
			wantErr: false,
		},
		{
			name: "CPU limit smaller than request",
			resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("200m"),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("100m"),
				},
			},
			wantErr: true,
		},
		{
			name: "memory limit smaller than request",
			resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("256Mi"),
				},
				Limits: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
			wantErr: true,
		},
		{
			name: "no limits specified",
			resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("100m"),
					v1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
			wantErr: false,
		},
		{
			name: "no requests specified",
			resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("200m"),
					v1.ResourceMemory: resource.MustParse("256Mi"),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := &CommonPrometheusFields{
				Resources: tt.resources,
			}
			err := fields.ValidateResources()
			if (err != nil) != tt.wantErr {
				t.Errorf("CommonPrometheusFields.ValidateResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAlertmanagerSpec_ValidateResources(t *testing.T) {
	tests := []struct {
		name      string
		resources v1.ResourceRequirements
		wantErr   bool
	}{
		{
			name: "valid resources",
			resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("100m"),
					v1.ResourceMemory: resource.MustParse("128Mi"),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("200m"),
					v1.ResourceMemory: resource.MustParse("256Mi"),
				},
			},
			wantErr: false,
		},
		{
			name: "CPU limit smaller than request",
			resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("200m"),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("100m"),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &AlertmanagerSpec{
				Resources: tt.resources,
			}
			err := spec.ValidateResources()
			if (err != nil) != tt.wantErr {
				t.Errorf("AlertmanagerSpec.ValidateResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestThanosRulerSpec_ValidateResources(t *testing.T) {
	tests := []struct {
		name      string
		resources v1.ResourceRequirements
		wantErr   bool
	}{
		{
			name: "valid resources",
			resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("100m"),
					v1.ResourceMemory: resource.MustParse("128Mi"),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("200m"),
					v1.ResourceMemory: resource.MustParse("256Mi"),
				},
			},
			wantErr: false,
		},
		{
			name: "memory limit smaller than request",
			resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("256Mi"),
				},
				Limits: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &ThanosRulerSpec{
				Resources: tt.resources,
			}
			err := spec.ValidateResources()
			if (err != nil) != tt.wantErr {
				t.Errorf("ThanosRulerSpec.ValidateResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
