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
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/metadata"

	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
)

// FinalizerSyncer holds the configuration and dependencies
// required to perform finalizer synchronization.
type FinalizerSyncer struct {
	mdClient                     metadata.Interface
	gvr                          schema.GroupVersionResource
	configResourcesStatusEnabled bool
}

func NewFinalizerSyncer(
	mdClient metadata.Interface,
	gvr schema.GroupVersionResource,
	configResourcesStatusEnabled bool,
) *FinalizerSyncer {
	return &FinalizerSyncer{
		mdClient:                     mdClient,
		gvr:                          gvr,
		configResourcesStatusEnabled: configResourcesStatusEnabled,
	}
}

// Sync ensures the `monitoring.coreos.com/status-cleanup` finalizer is correctly set on the given workload resource
// (Prometheus, PrometheusAgent, Alertmanager, or ThanosRuler). It adds the finalizer if necessary, or removes it when appropriate.
//
// Returns true if the finalizer list was modified, otherwise false.
// The second return value indicates any error encountered during the operation.
func (s *FinalizerSyncer) Sync(ctx context.Context, p metav1.Object, logger *slog.Logger, deletionInProgress bool) (bool, error) {
	if !s.configResourcesStatusEnabled {
		return false, nil
	}

	finalizers := p.GetFinalizers()

	// The resource isn't being deleted, add the finalizer if missing.
	if !deletionInProgress {
		// Add finalizer to the workload resource if it doesn't have one.
		patchBytes, err := k8sutil.FinalizerAddPatch(finalizers, k8sutil.StatusCleanupFinalizerName)
		if err != nil {
			return false, fmt.Errorf("failed to marshal patch: %w", err)
		}

		if len(patchBytes) == 0 {
			return false, nil
		}
		if err = s.updateObject(ctx, p, patchBytes); err != nil {
			return false, fmt.Errorf("failed to add %q finalizer: %w", k8sutil.StatusCleanupFinalizerName, err)
		}
		logger.Debug("added finalizer to object")
		return true, nil
	}

	// If the workload instance is marked for deletion, we remove the finalizer.
	patchBytes, err := k8sutil.FinalizerDeletePatch(finalizers, k8sutil.StatusCleanupFinalizerName)
	if err != nil {
		return false, fmt.Errorf("failed to marshal patch: %w", err)
	}
	if len(patchBytes) == 0 {
		return false, nil
	}

	if err = s.updateObject(ctx, p, patchBytes); err != nil {
		return false, fmt.Errorf("failed to remove %q finalizer: %w", k8sutil.StatusCleanupFinalizerName, err)
	}
	logger.Debug("removed finalizer from object")

	return true, nil
}

// updateObject applies a JSON patch to update the metadata of the given workload object (Prometheus, PrometheusAgent, Alertmanager, or ThanosRuler) in the cluster.
func (s *FinalizerSyncer) updateObject(
	ctx context.Context,
	p metav1.Object,
	patchBytes []byte,
) error {
	_, err := s.mdClient.Resource(s.gvr).
		Namespace(p.GetNamespace()).
		Patch(ctx, p.GetName(), types.JSONPatchType, patchBytes, metav1.PatchOptions{FieldManager: PrometheusOperatorFieldManager})

	return err
}
