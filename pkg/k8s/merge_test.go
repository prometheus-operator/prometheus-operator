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

package k8s

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestPodLabelsAnnotations(t *testing.T) {
	build := func(name, image string, ports ...corev1.ContainerPort) corev1.Container {
		return corev1.Container{
			Name:  name,
			Image: image,
			Ports: ports,
		}
	}

	port1A := corev1.ContainerPort{
		Name:          "portA",
		ContainerPort: 1,
	}
	port1B := corev1.ContainerPort{
		Name:          "portB",
		ContainerPort: 1,
	}
	port2A := corev1.ContainerPort{
		Name:          "portA",
		ContainerPort: 2,
	}

	testCases := []struct {
		name    string
		base    []corev1.Container
		patches []corev1.Container
		result  []corev1.Container
	}{
		// sanity checks
		{
			name: "everything nil",
		}, {
			name:   "no patch",
			base:   []corev1.Container{build("c1", "image:A")},
			result: []corev1.Container{build("c1", "image:A")},
		}, {
			name:    "no Base",
			patches: []corev1.Container{build("c1", "image:A")},
			result:  []corev1.Container{build("c1", "image:A")},
		}, {
			name:    "no conflict",
			base:    []corev1.Container{build("c1", "image:A")},
			patches: []corev1.Container{build("c2", "image:A")},
			result:  []corev1.Container{build("c1", "image:A"), build("c2", "image:A")},
		}, {
			name:    "no conflict with port",
			base:    []corev1.Container{build("c1", "image:A", port1A)},
			patches: []corev1.Container{build("c2", "image:A", port1B)},
			result:  []corev1.Container{build("c1", "image:A", port1A), build("c2", "image:A", port1B)},
		},
		// string conflicts
		{
			name:    "one conflict",
			base:    []corev1.Container{build("c1", "image:A")},
			patches: []corev1.Container{build("c1", "image:B")},
			result:  []corev1.Container{build("c1", "image:B")},
		}, {
			name:    "one conflict with ports",
			base:    []corev1.Container{build("c1", "image:A", port1A)},
			patches: []corev1.Container{build("c1", "image:B", port1A)},
			result:  []corev1.Container{build("c1", "image:B", port1A)},
		}, {
			name:    "out of order conflict",
			base:    []corev1.Container{build("c1", "image:A"), build("c2", "image:A")},
			patches: []corev1.Container{build("c2", "image:B"), build("c1", "image:B")},
			result:  []corev1.Container{build("c1", "image:B"), build("c2", "image:B")},
		},
		// struct conflict
		{
			name:    "port name conflict",
			base:    []corev1.Container{build("c1", "image:A", port1A)},
			patches: []corev1.Container{build("c1", "image:A", port2A)},
			result:  []corev1.Container{build("c1", "image:A", port2A, port1A)}, // port ordering doesn't matter here
		},
		{
			name:    "port value conflict",
			base:    []corev1.Container{build("c1", "image:A", port1A)},
			patches: []corev1.Container{build("c1", "image:A", port1B)},
			result:  []corev1.Container{build("c1", "image:A", port1B)}, // port ordering doesn't matter here
		},
		{
			name:    "empty image, add port",
			base:    []corev1.Container{build("c1", "image:A")},
			patches: []corev1.Container{build("c1", "", port1A)},
			result:  []corev1.Container{build("c1", "image:A", port1A)},
		},
	}

	for _, tc := range testCases {
		result, err := MergePatchContainers(tc.base, tc.patches)
		require.NoError(t, err)
		diff := pretty.Compare(result, tc.result)
		require.Equal(t, "", diff, "Test %s: patch result did not match. diff: %s.", tc.name, diff)
	}
}

func TestMergePatchContainersProbeHandlerSanitization(t *testing.T) {
	port := intstr.FromInt32(8080)

	httpGet := &corev1.HTTPGetAction{Path: "/health", Port: port}
	tcpSocket := &corev1.TCPSocketAction{Port: port}
	exec := &corev1.ExecAction{Command: []string{"cat", "/tmp/healthy"}}
	grpc := &corev1.GRPCAction{Port: 8080}

	probe := func(h corev1.ProbeHandler) *corev1.Probe {
		return &corev1.Probe{ProbeHandler: h}
	}

	testCases := []struct {
		name    string
		base    []corev1.Container
		patches []corev1.Container
		result  []corev1.Container
	}{
		{
			// Strategic merge leaves both HTTPGet (from base) and TCPSocket (from patch) set;
			// sanitization must keep only the patch handler.
			name: "readiness probe: patch overrides HTTPGet with TCPSocket",
			base: []corev1.Container{{
				Name:           "c1",
				ReadinessProbe: probe(corev1.ProbeHandler{HTTPGet: httpGet}),
			}},
			patches: []corev1.Container{{
				Name:           "c1",
				ReadinessProbe: probe(corev1.ProbeHandler{TCPSocket: tcpSocket}),
			}},
			result: []corev1.Container{{
				Name:           "c1",
				ReadinessProbe: probe(corev1.ProbeHandler{TCPSocket: tcpSocket}),
			}},
		},
		{
			name: "readiness probe: patch overrides HTTPGet with Exec",
			base: []corev1.Container{{
				Name:           "c1",
				ReadinessProbe: probe(corev1.ProbeHandler{HTTPGet: httpGet}),
			}},
			patches: []corev1.Container{{
				Name:           "c1",
				ReadinessProbe: probe(corev1.ProbeHandler{Exec: exec}),
			}},
			result: []corev1.Container{{
				Name:           "c1",
				ReadinessProbe: probe(corev1.ProbeHandler{Exec: exec}),
			}},
		},
		{
			name: "readiness probe: patch overrides HTTPGet with GRPC",
			base: []corev1.Container{{
				Name:           "c1",
				ReadinessProbe: probe(corev1.ProbeHandler{HTTPGet: httpGet}),
			}},
			patches: []corev1.Container{{
				Name:           "c1",
				ReadinessProbe: probe(corev1.ProbeHandler{GRPC: grpc}),
			}},
			result: []corev1.Container{{
				Name:           "c1",
				ReadinessProbe: probe(corev1.ProbeHandler{GRPC: grpc}),
			}},
		},
		{
			// Same handler type in base and patch: no conflict, no change needed.
			name: "readiness probe: same handler type in base and patch",
			base: []corev1.Container{{
				Name:           "c1",
				ReadinessProbe: probe(corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Path: "/old", Port: port}}),
			}},
			patches: []corev1.Container{{
				Name:           "c1",
				ReadinessProbe: probe(corev1.ProbeHandler{HTTPGet: httpGet}),
			}},
			result: []corev1.Container{{
				Name:           "c1",
				ReadinessProbe: probe(corev1.ProbeHandler{HTTPGet: httpGet}),
			}},
		},
		{
			// Patch does not touch the probe: base probe is preserved, no sanitization needed.
			name: "patch does not override probe",
			base: []corev1.Container{{
				Name:           "c1",
				Image:          "image:A",
				ReadinessProbe: probe(corev1.ProbeHandler{HTTPGet: httpGet}),
			}},
			patches: []corev1.Container{{
				Name:  "c1",
				Image: "image:B",
			}},
			result: []corev1.Container{{
				Name:           "c1",
				Image:          "image:B",
				ReadinessProbe: probe(corev1.ProbeHandler{HTTPGet: httpGet}),
			}},
		},
		{
			// A brand-new container supplied entirely by the user via patches has no base to
			// merge with. Even if it contains a malformed probe (multiple handlers), we leave
			// it untouched — it is the user's responsibility to supply a valid spec.
			name: "new container from patch with multiple probe handlers is not sanitized",
			patches: []corev1.Container{{
				Name:           "c1",
				ReadinessProbe: probe(corev1.ProbeHandler{HTTPGet: httpGet, TCPSocket: tcpSocket}),
			}},
			result: []corev1.Container{{
				Name:           "c1",
				ReadinessProbe: probe(corev1.ProbeHandler{HTTPGet: httpGet, TCPSocket: tcpSocket}),
			}},
		},
		{
			// Sanitization applies to all probe types, not just ReadinessProbe.
			name: "liveness and startup probes are also sanitized",
			base: []corev1.Container{{
				Name:         "c1",
				StartupProbe: probe(corev1.ProbeHandler{HTTPGet: httpGet}),
				LivenessProbe: probe(corev1.ProbeHandler{HTTPGet: httpGet}),
			}},
			patches: []corev1.Container{{
				Name:          "c1",
				StartupProbe:  probe(corev1.ProbeHandler{TCPSocket: tcpSocket}),
				LivenessProbe: probe(corev1.ProbeHandler{Exec: exec}),
			}},
			result: []corev1.Container{{
				Name:          "c1",
				StartupProbe:  probe(corev1.ProbeHandler{TCPSocket: tcpSocket}),
				LivenessProbe: probe(corev1.ProbeHandler{Exec: exec}),
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := MergePatchContainers(tc.base, tc.patches)
			require.NoError(t, err)
			diff := pretty.Compare(result, tc.result)
			require.Equal(t, "", diff, "patch result did not match. diff:\n%s", diff)
		})
	}
}

func TestMergePatchContainersOrderPreserved(t *testing.T) {
	build := func(name, image string) corev1.Container {
		return corev1.Container{
			Name:  name,
			Image: image,
		}
	}

	for range 10 {
		result, err := MergePatchContainers(
			[]corev1.Container{
				build("c1", "image:base"),
				build("c2", "image:base"),
			},
			[]corev1.Container{
				build("c1", "image:A"),
				build("c3", "image:B"),
				build("c4", "image:C"),
				build("c5", "image:D"),
				build("c6", "image:E"),
			},
		)
		require.NoError(t, err)

		diff := pretty.Compare(
			result,
			[]corev1.Container{
				build("c1", "image:A"),
				build("c2", "image:base"),
				build("c3", "image:B"),
				build("c4", "image:C"),
				build("c5", "image:D"),
				build("c6", "image:E"),
			},
		)
		require.Equal(t, "", diff, "patch result did not match. diff:\n%s", diff)
	}
}
