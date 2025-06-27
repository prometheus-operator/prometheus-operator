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

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
)

// SyncFinalizers ensures the `monitoring.coreos.com/status-cleanup` finalizer is correctly set on the given workload resource
// (Prometheus, PrometheusAgent, Alertmanager, or ThanosRuler). It adds the finalizer if necessary, or removes it when appropriate.
//
// Returns true if the finalizer list was modified, otherwise false.
// If the object is being deleted, it is also removed from the reconciliation tracker.
// The second return value indicates any error encountered during the operation.
func SyncFinalizers(ctx context.Context, p metav1.Object, key string, mdClient metadata.Interface, reconciliations *ReconciliationTracker, logger *slog.Logger, deletionInProgress bool, configResourcesStatusEnabled bool) (bool, error) {
	if !configResourcesStatusEnabled {
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
		if err = updateObject(ctx, mdClient, p, patchBytes); err != nil {
			return false, fmt.Errorf("failed to add %s finalizer: %w", k8sutil.StatusCleanupFinalizerName, err)
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
		reconciliations.ForgetObject(key)
		return false, nil
	}

	if err = updateObject(ctx, mdClient, p, patchBytes); err != nil {
		return false, fmt.Errorf("failed to remove %s finalizer: %w", k8sutil.StatusCleanupFinalizerName, err)
	}
	logger.Debug("removed finalizer from object")
	reconciliations.ForgetObject(key)

	return true, nil
}

// updateObject applies a JSON patch to update the metadata of the given workload object (Prometheus, PrometheusAgent, Alertmanager, or ThanosRuler) in the cluster.
func updateObject(
	ctx context.Context,
	mdClient metadata.Interface,
	p metav1.Object,
	patchBytes []byte,
) error {
	var err error
	var gvr schema.GroupVersionResource

	switch any(p).(type) {
	case *monitoringv1.Prometheus:
		gvr = monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PrometheusName)

	case *monitoringv1.Alertmanager:
		gvr = monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.AlertmanagerName)

	case *monitoringv1.ThanosRuler:
		gvr = monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ThanosRulerName)

	case *monitoringv1alpha1.PrometheusAgent:
		gvr = monitoringv1alpha1.SchemeGroupVersion.WithResource(monitoringv1alpha1.PrometheusAgentName)

	default:
		return fmt.Errorf("unknown object type %T", p)
	}

	_, err = mdClient.Resource(gvr).
		Namespace(p.GetNamespace()).
		Patch(ctx, p.GetName(), types.JSONPatchType, patchBytes, metav1.PatchOptions{FieldManager: PrometheusOperatorFieldManager})

	return err
}
