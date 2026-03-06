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

package k8s

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// MergePatchContainers adds patches to base using a strategic merge patch and
// iterating by container name, failing on the first error.
func MergePatchContainers(base, patches []v1.Container) ([]v1.Container, error) {
	var out []v1.Container

	containersByName := make(map[string]v1.Container)
	for _, c := range patches {
		containersByName[c.Name] = c
	}

	// Patch the containers that exist in base.
	for _, container := range base {
		patchContainer, ok := containersByName[container.Name]
		if !ok {
			// This container didn't need to be patched.
			out = append(out, container)
			continue
		}

		containerBytes, err := json.Marshal(container)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for container %s: %w", container.Name, err)
		}

		patchBytes, err := json.Marshal(patchContainer)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for patch container %s: %w", container.Name, err)
		}

		// Calculate the patch result.
		jsonResult, err := strategicpatch.StrategicMergePatch(containerBytes, patchBytes, v1.Container{})
		if err != nil {
			return nil, fmt.Errorf("failed to generate merge patch for container %s: %w", container.Name, err)
		}

		var patchResult v1.Container
		if err := json.Unmarshal(jsonResult, &patchResult); err != nil {
			return nil, fmt.Errorf("failed to unmarshal merged container %s: %w", container.Name, err)
		}
		sanitizeProbeHandlers(&patchResult)

		// Add the patch result and remove the corresponding key from the to do list.
		out = append(out, patchResult)
		delete(containersByName, container.Name)
	}

	// Append containers that are left in containersByName.
	// Iterate over patches to preserve the order.
	for _, c := range patches {
		if container, found := containersByName[c.Name]; found {
			sanitizeProbeHandlers(&container)
			out = append(out, container)
		}
	}

	return out, nil
}

// sanitizeProbeHandlers ensures each probe has at most one handler type.
// Strategic merge can leave multiple handler types set (e.g. HTTPGet from base + TCPSocket from patch),
// which causes Kubernetes "may not specify more than 1 handler type" validation error.
//
// Note: Using a TCPSocket or Exec probe as an override means the readiness probe
// bypasses the Prometheus /-/ready endpoint. Consequently, it will not account for
// internal state such as WAL replay or TSDB initialization. The Pod may be marked
// 'Ready' by Kubernetes before Prometheus is actually capable of serving queries.
// This is a known trade-off when prioritizing security (avoiding plaintext credentials)
// over granular readiness checks.
func sanitizeProbeHandlers(c *v1.Container) {
	for _, p := range []*v1.Probe{c.StartupProbe, c.ReadinessProbe, c.LivenessProbe} {
		if p == nil {
			continue
		}
		p.ProbeHandler = sanitizeProbeHandler(p.ProbeHandler)
	}
}

func sanitizeProbeHandler(h v1.ProbeHandler) v1.ProbeHandler {
	out := v1.ProbeHandler{}
	// Keep exactly one handler. Prefer patch-friendly types (TCPSocket, Exec) over default HTTPGet
	// so user override to tcpSocket/exec is respected after merge.
	if h.GRPC != nil {
		out.GRPC = h.GRPC
		return out
	}
	if h.TCPSocket != nil {
		out.TCPSocket = h.TCPSocket
		return out
	}
	if h.Exec != nil {
		out.Exec = h.Exec
		return out
	}
	if h.HTTPGet != nil {
		out.HTTPGet = h.HTTPGet
		return out
	}
	return out
}
