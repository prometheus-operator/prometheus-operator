// Copyright 2016 The prometheus-operator Authors
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

package k8sutil

import (
	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	"testing"
)

func TestPodLabelsAnnotations(t *testing.T) {
	build := func(name, image string, ports ...v1.ContainerPort) v1.Container {
		return v1.Container{
			Name:  name,
			Image: image,
			Ports: ports,
		}
	}

	port1A := v1.ContainerPort{
		Name:          "portA",
		ContainerPort: 1,
	}
	port1B := v1.ContainerPort{
		Name:          "portB",
		ContainerPort: 1,
	}
	port2A := v1.ContainerPort{
		Name:          "portA",
		ContainerPort: 2,
	}

	testCases := []struct {
		name    string
		base    []v1.Container
		patches []v1.Container
		result  []v1.Container
	}{
		// sanity checks
		{
			name: "everything nil",
		}, {
			name:   "no patch",
			base:   []v1.Container{build("c1", "image:A")},
			result: []v1.Container{build("c1", "image:A")},
		}, {
			name:    "no Base",
			patches: []v1.Container{build("c1", "image:A")},
			result:  []v1.Container{build("c1", "image:A")},
		}, {
			name:    "no conflict",
			base:    []v1.Container{build("c1", "image:A")},
			patches: []v1.Container{build("c2", "image:A")},
			result:  []v1.Container{build("c1", "image:A"), build("c2", "image:A")},
		}, {
			name:    "no conflict with port",
			base:    []v1.Container{build("c1", "image:A", port1A)},
			patches: []v1.Container{build("c2", "image:A", port1B)},
			result:  []v1.Container{build("c1", "image:A", port1A), build("c2", "image:A", port1B)},
		},
		// string conflicts
		{
			name:    "one conflict",
			base:    []v1.Container{build("c1", "image:A")},
			patches: []v1.Container{build("c1", "image:B")},
			result:  []v1.Container{build("c1", "image:B")},
		}, {
			name:    "one conflict with ports",
			base:    []v1.Container{build("c1", "image:A", port1A)},
			patches: []v1.Container{build("c1", "image:B", port1A)},
			result:  []v1.Container{build("c1", "image:B", port1A)},
		}, {
			name:    "out of order conflict",
			base:    []v1.Container{build("c1", "image:A"), build("c2", "image:A")},
			patches: []v1.Container{build("c2", "image:B"), build("c1", "image:B")},
			result:  []v1.Container{build("c1", "image:B"), build("c2", "image:B")},
		},
		// struct conflict
		{
			name:    "port name conflict",
			base:    []v1.Container{build("c1", "image:A", port1A)},
			patches: []v1.Container{build("c1", "image:A", port2A)},
			result:  []v1.Container{build("c1", "image:A", port2A, port1A)}, // port ordering doesn't matter here
		},
		{
			name:    "port value conflict",
			base:    []v1.Container{build("c1", "image:A", port1A)},
			patches: []v1.Container{build("c1", "image:A", port1B)},
			result:  []v1.Container{build("c1", "image:A", port1B)}, // port ordering doesn't matter here
		},
		{
			name:    "empty image, add port",
			base:    []v1.Container{build("c1", "image:A")},
			patches: []v1.Container{build("c1", "", port1A)},
			result:  []v1.Container{build("c1", "image:A", port1A)},
		},
	}

	for _, tc := range testCases {
		result, err := MergePatchContainers(tc.base, tc.patches)
		require.NoError(t, err)
		if diff := pretty.Compare(result, tc.result); diff != "" {
			t.Fatalf("Test %s: patch result did not match. diff: %s.", tc.name, diff)
		}
	}
}
