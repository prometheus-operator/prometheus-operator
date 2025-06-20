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
	"k8s.io/apimachinery/pkg/types"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
)

// SyncFinalizers adds or removes the finalizer form the workload resource(Prometheus, PrometheusAgent, Alertmanager and ThanosRuler).
// It returns true if the finalizers were modified, otherwise false. The second return value is an error, if any.
func SyncFinalizers[T metav1.Object](ctx context.Context, p T, key string, mclient monitoringclient.Interface, reconciliations *ReconciliationTracker, logger *slog.Logger, deletionInProgress bool, configResourcesStatusEnabled bool) (bool, error) {
	if !configResourcesStatusEnabled {
		return false, nil
	}

	// The resource isn't being deleted, add the finalizer if missing.
	if !deletionInProgress {
		// Add finalizer to the Prometheus resource if it doesn't have one.
		finalizers := p.GetFinalizers()
		patchBytes, err := k8sutil.FinalizerAddPatch(finalizers, k8sutil.StatusCleanupFinalizerName)
		if err != nil {
			return false, fmt.Errorf("failed to marshal patch: %w", err)
		}

		if len(patchBytes) == 0 {
			return false, nil
		}
		if err = updateObject[T](ctx, mclient, p, patchBytes); err != nil {
			return false, fmt.Errorf("failed to add %s finalizer: %w", k8sutil.StatusCleanupFinalizerName, err)
		}
		logger.Debug("added finalizer to object")
		return true, nil
	}

	// If the Prometheus instance is marked for deletion, we remove the finalizer.
	finalizers := p.GetFinalizers()
	patchBytes, err := k8sutil.FinalizerDeletePatch(finalizers, k8sutil.StatusCleanupFinalizerName)
	if err != nil {
		return false, fmt.Errorf("failed to marshal patch: %w", err)
	}
	if len(patchBytes) == 0 {
		reconciliations.ForgetObject(key)
		return false, nil
	}

	if err = updateObject[T](ctx, mclient, p, patchBytes); err != nil {
		return false, fmt.Errorf("failed to remove %s finalizer: %w", k8sutil.StatusCleanupFinalizerName, err)
	}
	logger.Debug("removed finalizer from object")
	reconciliations.ForgetObject(key)

	return true, nil
}

// updateObject updates the given object in the cluster using a JSON patch.
func updateObject[T metav1.Object](
	ctx context.Context,
	mclient monitoringclient.Interface,
	p T,
	patchBytes []byte,
) error {
	var err error
	opts := metav1.PatchOptions{FieldManager: PrometheusOperatorFieldManager}

	switch obj := any(p).(type) {
	case *monitoringv1.Prometheus:
		_, err = mclient.MonitoringV1().
			Prometheuses(obj.GetNamespace()).
			Patch(ctx, obj.GetName(), types.JSONPatchType, patchBytes, opts)

	case *monitoringv1.Alertmanager:
		_, err = mclient.MonitoringV1().
			Alertmanagers(obj.GetNamespace()).
			Patch(ctx, obj.GetName(), types.JSONPatchType, patchBytes, opts)

	case *monitoringv1.ThanosRuler:
		_, err = mclient.MonitoringV1().
			ThanosRulers(obj.GetNamespace()).
			Patch(ctx, obj.GetName(), types.JSONPatchType, patchBytes, opts)

	case *monitoringv1alpha1.PrometheusAgent:
		_, err = mclient.MonitoringV1alpha1().
			PrometheusAgents(obj.GetNamespace()).
			Patch(ctx, obj.GetName(), types.JSONPatchType, patchBytes, opts)

	default:
		return fmt.Errorf("unknown object type %T", p)
	}

	return err
}
