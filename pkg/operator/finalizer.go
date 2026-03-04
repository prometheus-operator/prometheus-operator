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

package operator

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/metadata"

	"github.com/prometheus-operator/prometheus-operator/pkg/k8s"
)

// FinalizerSyncer holds the configuration and dependencies
// required to perform finalizer synchronization.
type FinalizerSyncer struct {
	mdClient metadata.Interface
	gvr      schema.GroupVersionResource
	disabled bool
}

func NewFinalizerSyncer(mdClient metadata.Interface, gvr schema.GroupVersionResource) *FinalizerSyncer {
	return &FinalizerSyncer{
		mdClient: mdClient,
		gvr:      gvr,
	}
}

func NewNoopFinalizerSyncer() *FinalizerSyncer {
	return &FinalizerSyncer{
		disabled: true,
	}
}

// Sync ensures the `monitoring.coreos.com/status-cleanup` finalizer is correctly set on the given workload resource
// (Prometheus, PrometheusAgent, Alertmanager, or ThanosRuler). It adds the finalizer if necessary, or removes it when appropriate.
//
// Returns true if the finalizer was added, otherwise false.
// The second return value indicates any error encountered during the operation.
func (s *FinalizerSyncer) Sync(ctx context.Context, p metav1.Object, deletionInProgress bool, statusCleanup func() error) (bool, error) {
	if s.disabled {
		return false, nil
	}

	finalizers := p.GetFinalizers()

	// The resource isn't being deleted, add the finalizer if missing.
	if !deletionInProgress {
		// Add finalizer to the workload resource if it doesn't have one.
		patchBytes, err := k8s.FinalizerAddPatch(finalizers, k8s.StatusCleanupFinalizerName)
		if err != nil {
			return false, fmt.Errorf("failed to marshal patch: %w", err)
		}

		if len(patchBytes) == 0 {
			return false, nil
		}
		if err = s.updateObject(ctx, p, patchBytes); err != nil {
			return false, fmt.Errorf("failed to add %q finalizer: %w", k8s.StatusCleanupFinalizerName, err)
		}
		return true, nil
	}

	// Remove the workload bindings from the status of config resources.
	if err := statusCleanup(); err != nil {
		return false, fmt.Errorf("failed to clean up config resources status: %w", err)
	}

	// If the workload instance is marked for deletion, we remove the finalizer.
	patchBytes, err := k8s.FinalizerDeletePatch(finalizers, k8s.StatusCleanupFinalizerName)
	if err != nil {
		return false, fmt.Errorf("failed to marshal patch: %w", err)
	}
	if len(patchBytes) == 0 {
		return false, nil
	}

	if err = s.updateObject(ctx, p, patchBytes); err != nil {
		return false, fmt.Errorf("failed to remove %q finalizer: %w", k8s.StatusCleanupFinalizerName, err)
	}

	return false, nil
}

// updateObject applies a JSON patch to update the metadata of the given workload object (Prometheus, PrometheusAgent, Alertmanager, or ThanosRuler) in the cluster.
func (s *FinalizerSyncer) updateObject(
	ctx context.Context,
	p metav1.Object,
	patchBytes []byte,
) error {
	_, err := s.mdClient.Resource(s.gvr).
		Namespace(p.GetNamespace()).
		Patch(ctx, p.GetName(), types.JSONPatchType, patchBytes, metav1.PatchOptions{FieldManager: k8s.PrometheusOperatorFieldManager})

	return err
}
