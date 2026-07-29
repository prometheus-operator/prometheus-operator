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

package operator

import (
	"context"
	"log/slog"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/prometheus-operator/prometheus-operator/pkg/k8s"
)

type noopSyncer struct{}

func (noopSyncer) Sync(context.Context, string) error         { return nil }
func (noopSyncer) UpdateStatus(context.Context, string) error { return nil }

type noopGetter struct{}

func (noopGetter) Get(string) (runtime.Object, error) { return nil, nil }

func newTestResourceReconciler(t *testing.T) *ResourceReconciler {
	t.Helper()

	return NewResourceReconciler(
		slog.New(slog.DiscardHandler),
		noopSyncer{},
		noopGetter{},
		NewMetrics(prometheus.NewRegistry()),
		"Prometheus",
		prometheus.NewRegistry(),
		"",
	)
}

// TestOnUpdateDeletionInProgress ensures that the reconciler enqueues the
// object for reconciliation as soon as it gets marked for deletion, even if
// its generation, labels and annotations are unchanged. This is what allows
// the controller to run its deletion logic (e.g. remove the status cleanup
// finalizer) when the object is being deleted.
func TestOnUpdateDeletionInProgress(t *testing.T) {
	rr := newTestResourceReconciler(t)

	old := &metav1.PartialObjectMetadata{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "prometheus",
			Namespace:  "default",
			Finalizers: []string{k8s.StatusCleanupFinalizerName},
		},
	}

	now := metav1.Now()
	cur := old.DeepCopy()
	cur.DeletionTimestamp = &now

	rr.OnUpdate(old, cur)

	require.Equal(t, 1, rr.reconcileQ.Len())
}

// TestOnUpdateNoStateChange ensures that the reconciler doesn't enqueue the
// object when neither its state nor its deletion status changed, to avoid
// hot reconciliation loops.
func TestOnUpdateNoStateChange(t *testing.T) {
	rr := newTestResourceReconciler(t)

	old := &metav1.PartialObjectMetadata{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "prometheus",
			Namespace:       "default",
			ResourceVersion: "1",
			Finalizers:      []string{k8s.StatusCleanupFinalizerName},
		},
	}

	cur := old.DeepCopy()
	cur.ResourceVersion = "2"

	rr.OnUpdate(old, cur)

	require.Equal(t, 0, rr.reconcileQ.Len())
}

// TestOnUpdateDeletionInProgressMissedEvent ensures that the reconciler still
// enqueues the object while its deletion is in progress even when the old
// object already had the deletion timestamp set, e.g. because the informer
// missed the update event that originally set it.
func TestOnUpdateDeletionInProgressMissedEvent(t *testing.T) {
	rr := newTestResourceReconciler(t)

	now := metav1.Now()
	old := &metav1.PartialObjectMetadata{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "prometheus",
			Namespace:         "default",
			ResourceVersion:   "1",
			Finalizers:        []string{k8s.StatusCleanupFinalizerName},
			DeletionTimestamp: &now,
		},
	}

	cur := old.DeepCopy()
	cur.ResourceVersion = "2"

	rr.OnUpdate(old, cur)

	require.Equal(t, 1, rr.reconcileQ.Len())
}

// TestOnUpdateDeletionInProgressWithoutFinalizer ensures that the reconciler
// still skips objects being deleted when they don't carry the status
// cleanup finalizer, since the controller has nothing left to do for them.
func TestOnUpdateDeletionInProgressWithoutFinalizer(t *testing.T) {
	rr := newTestResourceReconciler(t)

	old := &metav1.PartialObjectMetadata{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus",
			Namespace: "default",
		},
	}

	now := metav1.Now()
	cur := old.DeepCopy()
	cur.DeletionTimestamp = &now

	rr.OnUpdate(old, cur)

	require.Equal(t, 0, rr.reconcileQ.Len())
}
