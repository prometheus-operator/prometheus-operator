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

package operator

import (
	"reflect"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	DefaultReloaderTestConfig = &Config{
		ReloaderConfig: ContainerConfig{
			CPURequest:    "100m",
			CPULimit:      "100m",
			MemoryRequest: "50Mi",
			MemoryLimit:   "50Mi",
			Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
		},
	}
)

func TestSidecarsResources(t *testing.T, makeStatefulSet func(reloaderConfig ContainerConfig) *appsv1.StatefulSet) {
	for _, tc := range []struct {
		name              string
		reloaderConfig    ContainerConfig
		expectedResources v1.ResourceRequirements
	}{
		{
			name: "no_resources",
			reloaderConfig: ContainerConfig{
				CPURequest:    "0",
				CPULimit:      "0",
				MemoryRequest: "0",
				MemoryLimit:   "0",
				Image:         DefaultReloaderTestConfig.ReloaderConfig.Image,
			},
			expectedResources: v1.ResourceRequirements{
				Limits:   v1.ResourceList{},
				Requests: v1.ResourceList{},
			},
		},
		{
			name: "no_requests",
			reloaderConfig: ContainerConfig{
				CPURequest:    "0",
				CPULimit:      "100m",
				MemoryRequest: "0",
				MemoryLimit:   "50Mi",
				Image:         DefaultReloaderTestConfig.ReloaderConfig.Image,
			},
			expectedResources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("100m"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
				Requests: v1.ResourceList{},
			},
		},
		{
			name: "no_limits",
			reloaderConfig: ContainerConfig{
				CPURequest:    "100m",
				CPULimit:      "0",
				MemoryRequest: "50Mi",
				MemoryLimit:   "0",
				Image:         DefaultReloaderTestConfig.ReloaderConfig.Image,
			},
			expectedResources: v1.ResourceRequirements{
				Limits: v1.ResourceList{},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("100m"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
		},
		{
			name: "no_CPU_resources",
			reloaderConfig: ContainerConfig{
				CPURequest:    "0",
				CPULimit:      "0",
				MemoryRequest: "50Mi",
				MemoryLimit:   "50Mi",
				Image:         DefaultReloaderTestConfig.ReloaderConfig.Image,
			},
			expectedResources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
				Requests: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
		},
		{
			name: "no_CPU_requests",
			reloaderConfig: ContainerConfig{
				CPURequest:    "0",
				CPULimit:      "100m",
				MemoryRequest: "50Mi",
				MemoryLimit:   "50Mi",
				Image:         DefaultReloaderTestConfig.ReloaderConfig.Image,
			},
			expectedResources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("100m"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
				Requests: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
		},
		{
			name: "no_CPU_limits",
			reloaderConfig: ContainerConfig{
				CPURequest:    "100m",
				CPULimit:      "0",
				MemoryRequest: "50Mi",
				MemoryLimit:   "50Mi",
				Image:         DefaultReloaderTestConfig.ReloaderConfig.Image,
			},
			expectedResources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("100m"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
		},
		{
			name: "no_memory_resources",
			reloaderConfig: ContainerConfig{
				CPURequest:    "100m",
				CPULimit:      "100m",
				MemoryRequest: "0",
				MemoryLimit:   "0",
				Image:         DefaultReloaderTestConfig.ReloaderConfig.Image,
			},
			expectedResources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("100m"),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("100m"),
				},
			},
		},
		{
			name: "no_memory_requests",
			reloaderConfig: ContainerConfig{
				CPURequest:    "100m",
				CPULimit:      "100m",
				MemoryRequest: "0",
				MemoryLimit:   "50Mi",
				Image:         DefaultReloaderTestConfig.ReloaderConfig.Image,
			},
			expectedResources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("100m"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("100m"),
				},
			},
		},
		{
			name: "no_memory_limits",
			reloaderConfig: ContainerConfig{
				CPURequest:    "100m",
				CPULimit:      "100m",
				MemoryRequest: "50Mi",
				MemoryLimit:   "0",
				Image:         DefaultReloaderTestConfig.ReloaderConfig.Image,
			},
			expectedResources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("100m"),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("100m"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sset := makeStatefulSet(tc.reloaderConfig)
			foundContainer := false

			for _, c := range sset.Spec.Template.Spec.Containers {
				if strings.HasSuffix(c.Name, "config-reloader") {
					foundContainer = true
				}
				if strings.HasSuffix(c.Name, "config-reloader") && !reflect.DeepEqual(c.Resources, tc.expectedResources) {
					t.Fatalf("Expected resource requests/limits:\n\n%s\n\nGot:\n\n%s", tc.expectedResources.String(), c.Resources.String())
				}
			}

			if !foundContainer {
				t.Fatalf("Expected to find a config-reloader container but it did")
			}
		})
	}
}
