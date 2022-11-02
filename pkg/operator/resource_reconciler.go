// Copyright 2022 The prometheus-operator Authors
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
	"reflect"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Syncer knows how to synchronize statefulset-based resources.
type Syncer interface {
	// Sync the state of the object identified by its key.
	Sync(context.Context, string) error
	// UpdateStatus updates the status of the object identified by its key.
	UpdateStatus(context.Context, string) error
	// Resolve returns the resource associated to the statefulset.
	Resolve(*appsv1.StatefulSet) metav1.Object
}

// ReconcilerMetrics tracks reconciler metrics.
type ReconcilerMetrics interface {
	TriggerByCounter(string, HandlerEvent) prometheus.Counter
}

// ResourceReconciler reacts on changes for statefulset-based resources and
// triggers synchronization of the resources.
//
// ResourceReconciler implements the cache.ResourceEventHandler interface and
// it can subscribe to resource events like this:
//
// var statefulSetInformer, resourceInformer cache.SharedInformer
// ...
// rr := NewResourceReconciler(...)
// statefulSetInformer.AddEventHandler(rr)
// resourceInformer.AddEventHandler(rr)
//
// ResourceReconciler will trigger object and status reconciliations based on
// the events received from the informer.
type ResourceReconciler struct {
	logger log.Logger

	resourceKind string

	syncer Syncer

	reconcileTotal    prometheus.Counter
	reconcileErrors   prometheus.Counter
	reconcileDuration prometheus.Histogram

	metrics ReconcilerMetrics

	// Queue to trigger state reconciliations of  objects.
	reconcileQ workqueue.RateLimitingInterface
	// Queue to trigger status updates of Prometheus objects.
	statusQ workqueue.RateLimitingInterface

	g errgroup.Group
}

// NewResourceReconciler returns a reconciler for the "kind" resource.
func NewResourceReconciler(
	l log.Logger,
	syncer Syncer,
	metrics ReconcilerMetrics,
	kind string,
	reg prometheus.Registerer,
) *ResourceReconciler {
	reconcileTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_reconcile_operations_total",
		Help: "Total number of reconcile operations",
	})

	reconcileErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_reconcile_errors_total",
		Help: "Number of errors that occurred during reconcile operations",
	})

	reconcileDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "prometheus_operator_reconcile_duration_seconds",
		Help:    "Histogram of reconcile operations",
		Buckets: []float64{.1, .5, 1, 5, 10},
	})

	reg.MustRegister(reconcileTotal, reconcileErrors, reconcileDuration)

	qname := strings.ToLower(kind)

	for _, t := range []string{"StatefulSet", kind} {
		for _, e := range []HandlerEvent{AddEvent, DeleteEvent, UpdateEvent} {
			metrics.TriggerByCounter(t, e)
		}
	}

	return &ResourceReconciler{
		logger:       l,
		resourceKind: kind,
		syncer:       syncer,

		reconcileTotal:    reconcileTotal,
		reconcileErrors:   reconcileErrors,
		reconcileDuration: reconcileDuration,
		metrics:           metrics,

		reconcileQ: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), qname),
		statusQ:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), qname+"_status"),
	}
}

// DeletionInProgress returns true if the object deletion has been requested.
func (rr *ResourceReconciler) DeletionInProgress(o metav1.Object) bool {
	if o.GetDeletionTimestamp() != nil {
		level.Debug(rr.logger).Log(
			"msg", "object deletion in progress",
			"object", fmt.Sprintf("%s/%s", o.GetNamespace(), o.GetName()),
		)
		return true
	}
	return false
}

// hasObjectChanged returns true if the objects have different resource revisions.
func (rr *ResourceReconciler) hasObjectChanged(old, cur metav1.Object) bool {
	if old.GetResourceVersion() != cur.GetResourceVersion() {
		level.Debug(rr.logger).Log(
			"msg", "different resource versions",
			"current", cur.GetResourceVersion(),
			"old", old.GetResourceVersion(),
			"object", fmt.Sprintf("%s/%s", cur.GetNamespace(), cur.GetName()),
		)
		return true
	}

	return false
}

// hasStateChanged returns true if the 2 objects are different in a way that
// the controller should reconcile the actual state against the desired state.
// It helps preventing hot loops when the controller updates the status
// subresource for instance.
func (rr *ResourceReconciler) hasStateChanged(old, cur metav1.Object) bool {
	if old.GetGeneration() != cur.GetGeneration() {
		level.Debug(rr.logger).Log(
			"msg", "different generations",
			"current", cur.GetGeneration(),
			"old", old.GetGeneration(),
			"object", fmt.Sprintf("%s/%s", cur.GetNamespace(), cur.GetName()),
		)
		return true
	}

	if !reflect.DeepEqual(old.GetLabels(), cur.GetLabels()) {
		level.Debug(rr.logger).Log(
			"msg", "different labels",
			"current", fmt.Sprintf("%v", cur.GetLabels()),
			"old", fmt.Sprintf("%v", old.GetLabels()),
			"object", fmt.Sprintf("%s/%s", cur.GetNamespace(), cur.GetName()),
		)
		return true

	}
	if !reflect.DeepEqual(old.GetAnnotations(), cur.GetAnnotations()) {
		level.Debug(rr.logger).Log(
			"msg", "different annotations",
			"current", fmt.Sprintf("%v", cur.GetAnnotations()),
			"old", fmt.Sprintf("%v", old.GetAnnotations()),
			"object", fmt.Sprintf("%s/%s", cur.GetNamespace(), cur.GetName()),
		)
		return true
	}

	return false
}

// objectKey returns the `namespace/name` key of a Kubernetes object, typically
// retrieved from a controller's cache.
func (rr *ResourceReconciler) objectKey(obj interface{}) (string, bool) {
	k, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		level.Error(rr.logger).Log("msg", "creating key failed", "err", err)
		return "", false
	}

	return k, true
}

// OnAdd implements the cache.ResourceEventHandler interface.
func (rr *ResourceReconciler) OnAdd(obj interface{}) {
	if _, ok := obj.(*appsv1.StatefulSet); ok {
		rr.onStatefulSetAdd(obj.(*appsv1.StatefulSet))
		return
	}

	key, ok := rr.objectKey(obj)
	if !ok {
		return
	}

	level.Debug(rr.logger).Log("msg", fmt.Sprintf("%s added", rr.resourceKind), "key", key)
	rr.metrics.TriggerByCounter(rr.resourceKind, AddEvent).Inc()

	rr.reconcileQ.Add(key)
}

// OnUpdate implements the cache.ResourceEventHandler interface.
func (rr *ResourceReconciler) OnUpdate(old, cur interface{}) {
	if _, ok := cur.(*appsv1.StatefulSet); ok {
		rr.onStatefulSetUpdate(old.(*appsv1.StatefulSet), cur.(*appsv1.StatefulSet))
		return
	}

	key, ok := rr.objectKey(cur)
	if !ok {
		return
	}

	mOld, err := meta.Accessor(old)
	if err != nil {
		level.Error(rr.logger).Log("err", fmt.Sprintf("failed to get object meta: %s", err), "key", key)
	}

	mCur, err := meta.Accessor(cur)
	if err != nil {
		level.Error(rr.logger).Log("err", fmt.Sprintf("failed to get object meta: %s", err), "key", key)
	}

	if rr.DeletionInProgress(mCur) {
		return
	}

	if !rr.hasStateChanged(mOld, mCur) {
		return
	}

	level.Debug(rr.logger).Log("msg", fmt.Sprintf("%s updated", rr.resourceKind), "key", key)
	rr.metrics.TriggerByCounter(rr.resourceKind, UpdateEvent).Inc()

	rr.reconcileQ.Add(key)
}

// OnDelete implements the cache.ResourceEventHandler interface.
func (rr *ResourceReconciler) OnDelete(obj interface{}) {
	if _, ok := obj.(*appsv1.StatefulSet); ok {
		rr.onStatefulSetDelete(obj.(*appsv1.StatefulSet))
		return
	}

	key, ok := rr.objectKey(obj)
	if !ok {
		return
	}

	level.Debug(rr.logger).Log("msg", fmt.Sprintf("%s deleted", rr.resourceKind), "key", key)
	rr.metrics.TriggerByCounter(rr.resourceKind, DeleteEvent).Inc()

	rr.reconcileQ.Add(key)
}

func (rr *ResourceReconciler) onStatefulSetAdd(ss *appsv1.StatefulSet) {
	obj := rr.syncer.Resolve(ss)
	if obj == nil {
		return
	}

	level.Debug(rr.logger).Log("msg", "StatefulSet added")
	rr.metrics.TriggerByCounter("StatefulSet", AddEvent).Inc()

	rr.EnqueueForReconciliation(obj)
}

func (rr *ResourceReconciler) onStatefulSetUpdate(old, cur *appsv1.StatefulSet) {
	level.Debug(rr.logger).Log("msg", "update handler", "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	if rr.DeletionInProgress(cur) {
		return
	}

	if !rr.hasObjectChanged(old, cur) {
		return
	}

	obj := rr.syncer.Resolve(cur)
	if obj == nil {
		return
	}

	level.Debug(rr.logger).Log("msg", "StatefulSet updated")
	rr.metrics.TriggerByCounter("StatefulSet", UpdateEvent).Inc()

	if !rr.hasStateChanged(old, cur) {
		// If the statefulset state (spec, labels or annotations) hasn't
		// changed, the operator can only update the status subresource instead
		// of doing a full reconciliation.
		rr.EnqueueForStatus(obj)
		return
	}

	rr.EnqueueForReconciliation(obj)
}

func (rr *ResourceReconciler) onStatefulSetDelete(ss *appsv1.StatefulSet) {
	obj := rr.syncer.Resolve(ss)
	if obj == nil {
		return
	}

	level.Debug(rr.logger).Log("msg", "StatefulSet delete")
	rr.metrics.TriggerByCounter("StatefulSet", DeleteEvent).Inc()

	rr.EnqueueForReconciliation(obj)
}

// EnqueueForReconciliation asks for reconciling the object.
func (rr *ResourceReconciler) EnqueueForReconciliation(obj metav1.Object) {
	rr.reconcileQ.Add(obj.GetNamespace() + "/" + obj.GetName())
}

// EnqueueForStatus asks for updating the status of the object.
func (rr *ResourceReconciler) EnqueueForStatus(obj metav1.Object) {
	rr.statusQ.Add(obj.GetNamespace() + "/" + obj.GetName())
}

// Run the goroutines responsible for processing the reconciliation and status
// queues.
func (rr *ResourceReconciler) Run(ctx context.Context) {
	// Goroutine that reconciles the desired state of objects.
	rr.g.Go(func() error {
		for rr.processNextReconcileItem(ctx) {
		}
		return nil
	})

	// Goroutine that reconciles the status of objects.
	rr.g.Go(func() error {
		for rr.processNextStatusItem(ctx) {
		}
		return nil
	})
}

// Stop the processing queues and wait for goroutines to exit.
func (rr *ResourceReconciler) Stop() {
	rr.reconcileQ.ShutDown()
	rr.statusQ.ShutDown()

	_ = rr.g.Wait()
}

// processNextReconcileItem dequeues items, processes them, and marks them done.
// It is guaranteed that the sync() method is never invoked concurrently with
// the same key.
// Before returning, the object's key is automatically added to the status queue.
func (rr *ResourceReconciler) processNextReconcileItem(ctx context.Context) bool {
	item, quit := rr.reconcileQ.Get()
	if quit {
		return false
	}

	key := item.(string)
	defer rr.reconcileQ.Done(key)
	defer rr.statusQ.Add(key) // enqueues the object's key to update the status subresource

	rr.reconcileTotal.Inc()
	startTime := time.Now()
	err := rr.syncer.Sync(ctx, key)
	rr.reconcileDuration.Observe(time.Since(startTime).Seconds())

	if err == nil {
		rr.reconcileQ.Forget(key)
		return true
	}

	rr.reconcileErrors.Inc()
	utilruntime.HandleError(errors.Wrap(err, fmt.Sprintf("sync %q failed", key)))
	rr.reconcileQ.AddRateLimited(key)

	return true
}

func (rr *ResourceReconciler) processNextStatusItem(ctx context.Context) bool {
	item, quit := rr.statusQ.Get()
	if quit {
		return false
	}

	key := item.(string)
	defer rr.statusQ.Done(key)

	err := rr.syncer.UpdateStatus(ctx, key)
	if err == nil {
		rr.statusQ.Forget(key)
		return true
	}

	utilruntime.HandleError(errors.Wrap(err, fmt.Sprintf("status %q failed", key)))
	rr.statusQ.AddRateLimited(key)

	return true
}
