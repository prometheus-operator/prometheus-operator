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
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/ptr"

	"github.com/prometheus-operator/prometheus-operator/pkg/k8s"
)

// Syncer knows how to synchronize statefulset-based or daemonset-based resources.
type Syncer interface {
	// Sync the state of the object identified by its key.
	Sync(context.Context, string) error
	// UpdateStatus updates the status of the object identified by its key.
	UpdateStatus(context.Context, string) error
}

// OwnedResourceOwner returns an object from its "<namespace>/<name>" key.
type OwnedResourceOwner interface {
	Get(string) (runtime.Object, error)
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
// rr := NewResourceReconciler(..., "Prometheus", ...)
// statefulSetInformer.AddEventHandler(rr)
// resourceInformer.AddEventHandler(rr)
//
// ResourceReconciler will trigger object and status reconciliations based on
// the events received from the informer.
type ResourceReconciler struct {
	logger *slog.Logger

	resourceKind string

	syncer Syncer
	getter OwnedResourceOwner

	reconcileTotal    prometheus.Counter
	reconcileErrors   prometheus.Counter
	reconcileDuration prometheus.Histogram
	statusTotal       prometheus.Counter
	statusErrors      prometheus.Counter

	metrics ReconcilerMetrics

	// Queue to trigger state reconciliations of  objects.
	reconcileQ workqueue.TypedRateLimitingInterface[string]
	// Queue to trigger status updates of Prometheus objects.
	statusQ workqueue.TypedRateLimitingInterface[string]

	g errgroup.Group

	controllerID string
}

var (
	_ = cache.ResourceEventHandler(&ResourceReconciler{})
)

const (
	controllerIDAnnotation = "operator.prometheus.io/controller-id"
)

type workQueueMetricsProvider struct {
	depth                          *prometheus.GaugeVec
	addsTotal                      *prometheus.CounterVec
	latency                        *prometheus.HistogramVec
	workDuration                   *prometheus.HistogramVec
	unfinishedWorkSeconds          *prometheus.GaugeVec
	longestRunningProcessorSeconds *prometheus.GaugeVec
	retriesTotal                   *prometheus.CounterVec
}

var _ = workqueue.MetricsProvider(&workQueueMetricsProvider{})

func newWorkQueueMetricsProvider(reg prometheus.Registerer) *workQueueMetricsProvider {
	mp := &workQueueMetricsProvider{
		depth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "prometheus_operator_workqueue_depth",
				Help: "Depth of the queue",
			},
			[]string{"name"},
		),
		addsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "prometheus_operator_workqueue_adds_total",
				Help: "Total number of additions to the queue",
			},
			[]string{"name"},
		),
		latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:                            "prometheus_operator_workqueue_latency_seconds",
				Help:                            "Histogram of latency for the queue",
				Buckets:                         []float64{.1, .5, 1, 5, 10},
				NativeHistogramBucketFactor:     1.1,
				NativeHistogramMaxBucketNumber:  100,
				NativeHistogramMinResetDuration: 1 * time.Hour,
			},
			[]string{"name"},
		),
		workDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:                            "prometheus_operator_workqueue_work_duration_seconds",
				Help:                            "Histogram of work duration for the queue",
				Buckets:                         []float64{.1, .5, 1, 5, 10},
				NativeHistogramBucketFactor:     1.1,
				NativeHistogramMaxBucketNumber:  100,
				NativeHistogramMinResetDuration: 1 * time.Hour,
			},
			[]string{"name"},
		),
		unfinishedWorkSeconds: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "prometheus_operator_workqueue_unfinished_work_seconds",
				Help: "How many seconds has been spent by processing work which is not yet finished. A growing number indicates a stuck thread.",
			},
			[]string{"name"},
		),
		longestRunningProcessorSeconds: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "prometheus_operator_workqueue_longest_running_processor_seconds",
				Help: "How many seconds has the longest running (unfinished) processor spent.",
			},
			[]string{"name"},
		),
		retriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "prometheus_operator_workqueue_retries_total",
				Help: "Total number of retries",
			},
			[]string{"name"},
		),
	}

	reg.MustRegister(
		mp.depth,
		mp.addsTotal,
		mp.latency,
		mp.workDuration,
		mp.unfinishedWorkSeconds,
		mp.longestRunningProcessorSeconds,
		mp.retriesTotal,
	)

	return mp
}

func (mp *workQueueMetricsProvider) NewDepthMetric(name string) workqueue.GaugeMetric {
	return mp.depth.WithLabelValues(name)
}

func (mp *workQueueMetricsProvider) NewAddsMetric(name string) workqueue.CounterMetric {
	return mp.addsTotal.WithLabelValues(name)
}

func (mp *workQueueMetricsProvider) NewLatencyMetric(name string) workqueue.HistogramMetric {
	return mp.latency.WithLabelValues(name)
}

func (mp *workQueueMetricsProvider) NewWorkDurationMetric(name string) workqueue.HistogramMetric {
	return mp.workDuration.WithLabelValues(name)
}

func (mp *workQueueMetricsProvider) NewUnfinishedWorkSecondsMetric(name string) workqueue.SettableGaugeMetric {
	return mp.unfinishedWorkSeconds.WithLabelValues(name)
}

func (mp *workQueueMetricsProvider) NewLongestRunningProcessorSecondsMetric(name string) workqueue.SettableGaugeMetric {
	return mp.longestRunningProcessorSeconds.WithLabelValues(name)
}

func (mp *workQueueMetricsProvider) NewRetriesMetric(name string) workqueue.CounterMetric {
	return mp.retriesTotal.WithLabelValues(name)
}

// NewResourceReconciler returns a reconciler for the "kind" resource.
func NewResourceReconciler(
	l *slog.Logger,
	syncer Syncer,
	getter OwnedResourceOwner,
	metrics ReconcilerMetrics,
	kind string,
	reg prometheus.Registerer,
	controllerID string,
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
		Name:                            "prometheus_operator_reconcile_duration_seconds",
		Help:                            "Histogram of reconcile operations",
		Buckets:                         []float64{.1, .5, 1, 5, 10},
		NativeHistogramBucketFactor:     1.1,
		NativeHistogramMaxBucketNumber:  100,
		NativeHistogramMinResetDuration: 1 * time.Hour,
	})

	statusTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_status_update_operations_total",
		Help: "Total number of update operations to status subresources",
	})

	statusErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_status_update_errors_total",
		Help: "Number of errors that occurred during update operations to status subresources",
	})

	reg.MustRegister(reconcileTotal, reconcileErrors, reconcileDuration, statusTotal, statusErrors)
	mp := newWorkQueueMetricsProvider(reg)

	qname := strings.ToLower(kind)

	// TODO: Support reconciling metrics for DaemonSet resource
	for _, t := range []string{"StatefulSet", kind} {
		for _, e := range []HandlerEvent{AddEvent, DeleteEvent, UpdateEvent} {
			metrics.TriggerByCounter(t, e)
		}
	}

	return &ResourceReconciler{
		logger:       l,
		resourceKind: kind,
		syncer:       syncer,
		getter:       getter,

		reconcileTotal:    reconcileTotal,
		reconcileErrors:   reconcileErrors,
		reconcileDuration: reconcileDuration,
		statusTotal:       statusTotal,
		statusErrors:      statusErrors,
		metrics:           metrics,
		controllerID:      controllerID,

		reconcileQ: workqueue.NewTypedRateLimitingQueueWithConfig[string](
			workqueue.DefaultTypedControllerRateLimiter[string](),
			workqueue.TypedRateLimitingQueueConfig[string]{
				Name:            qname,
				MetricsProvider: mp,
			},
		),
		statusQ: workqueue.NewTypedRateLimitingQueueWithConfig[string](
			workqueue.DefaultTypedControllerRateLimiter[string](),
			workqueue.TypedRateLimitingQueueConfig[string]{
				Name:            qname + "_status",
				MetricsProvider: mp,
			},
		),
	}
}

// DeletionInProgress returns true if the object deletion has been requested.
func (rr *ResourceReconciler) DeletionInProgress(o metav1.Object) bool {
	if o.GetDeletionTimestamp() != nil {
		rr.logger.Debug("object deletion in progress",
			"object", KeyForObject(o),
		)
		return true
	}
	return false
}

// hasObjectChanged returns true if the objects have different resource revisions.
func (rr *ResourceReconciler) hasObjectChanged(old, cur metav1.Object) bool {
	if old.GetResourceVersion() != cur.GetResourceVersion() {
		rr.logger.Debug("different resource versions",
			"current", cur.GetResourceVersion(),
			"old", old.GetResourceVersion(),
			"object", KeyForObject(cur),
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
		rr.logger.Debug("different generations",
			"current", cur.GetGeneration(),
			"old", old.GetGeneration(),
			"object", KeyForObject(cur),
		)
		return true
	}

	if !reflect.DeepEqual(old.GetLabels(), cur.GetLabels()) {
		rr.logger.Debug("different labels",
			"current", fmt.Sprintf("%v", cur.GetLabels()),
			"old", fmt.Sprintf("%v", old.GetLabels()),
			"object", KeyForObject(cur),
		)
		return true

	}
	if !reflect.DeepEqual(old.GetAnnotations(), cur.GetAnnotations()) {
		rr.logger.Debug("different annotations",
			"current", fmt.Sprintf("%v", cur.GetAnnotations()),
			"old", fmt.Sprintf("%v", old.GetAnnotations()),
			"object", KeyForObject(cur),
		)
		return true
	}

	return false
}

// objectKey returns the `namespace/name` key of a Kubernetes object, typically
// retrieved from a controller's cache.
func (rr *ResourceReconciler) objectKey(obj any) (string, bool) {
	k, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		rr.logger.Error("creating key failed", "err", err)
		return "", false
	}

	return k, true
}

func (rr *ResourceReconciler) resolve(obj metav1.Object) metav1.Object {
	for _, or := range obj.GetOwnerReferences() {
		if !ptr.Deref(or.Controller, false) {
			continue
		}

		if or.Kind != rr.resourceKind {
			continue
		}

		owner, err := rr.getter.Get(KeyForObject(&metav1.ObjectMeta{Name: or.Name, Namespace: obj.GetNamespace()}))
		if err != nil {
			if !apierrors.IsNotFound(err) {
				rr.logger.Error("failed to resolve controller owner", "err", err, "namespace", obj.GetNamespace(), "name", obj.GetName(), "kind", rr.resourceKind)
			}

			return nil
		}

		owner = owner.DeepCopyObject()
		o, err := meta.Accessor(owner)
		if err != nil {
			rr.logger.Error("failed to get owner meta", "err", err, "gvk", owner.GetObjectKind().GroupVersionKind().String(), "namespace", obj.GetNamespace(), "name", obj.GetName(), "kind", rr.resourceKind)
		}

		return o
	}

	rr.logger.Debug("no known controller owner", "namespace", obj.GetNamespace(), "name", obj.GetName())
	return nil
}

// OnAdd implements the cache.ResourceEventHandler interface.
func (rr *ResourceReconciler) OnAdd(obj any, _ bool) {

	switch v := obj.(type) {
	case *appsv1.DaemonSet:
		rr.onDaemonSetAdd(v)
		return
	case *appsv1.StatefulSet:
		rr.onStatefulSetAdd(v)
		return
	}

	key, ok := rr.objectKey(obj)
	if !ok {
		return
	}

	objMeta, err := meta.Accessor(obj)
	if err != nil {
		return
	}

	if !rr.isManagedByController(objMeta) {
		return
	}

	rr.logger.Debug(fmt.Sprintf("%s added", rr.resourceKind), "key", key)
	rr.metrics.TriggerByCounter(rr.resourceKind, AddEvent).Inc()

	rr.reconcileQ.Add(key)
}

// OnUpdate implements the cache.ResourceEventHandler interface.
func (rr *ResourceReconciler) OnUpdate(old, cur any) {
	switch v := cur.(type) {
	case *appsv1.DaemonSet:
		rr.onDaemonSetUpdate(old.(*appsv1.DaemonSet), v)
		return
	case *appsv1.StatefulSet:
		rr.onStatefulSetUpdate(old.(*appsv1.StatefulSet), v)
		return
	}

	key, ok := rr.objectKey(cur)
	if !ok {
		return
	}

	mOld, err := meta.Accessor(old)
	if err != nil {
		rr.logger.Error("failed to get old object meta", "err", err, "key", key)
		return
	}

	mCur, err := meta.Accessor(cur)
	if err != nil {
		rr.logger.Error("failed to get current object meta", "err", err, "key", key)
		return
	}

	if !rr.isManagedByController(mCur) {
		return
	}

	if !k8s.HasStatusCleanupFinalizer(mCur) && rr.DeletionInProgress(mCur) {
		return
	}

	if !rr.hasStateChanged(mOld, mCur) {
		return
	}

	rr.logger.Debug(fmt.Sprintf("%s updated", rr.resourceKind), "key", key)
	rr.metrics.TriggerByCounter(rr.resourceKind, UpdateEvent).Inc()

	rr.reconcileQ.Add(key)
}

// OnDelete implements the cache.ResourceEventHandler interface.
func (rr *ResourceReconciler) OnDelete(obj any) {
	switch v := obj.(type) {
	case *appsv1.DaemonSet:
		rr.onDaemonSetDelete(v)
		return
	case *appsv1.StatefulSet:
		rr.onStatefulSetDelete(v)
		return
	}

	key, ok := rr.objectKey(obj)
	if !ok {
		return
	}

	objMeta, err := meta.Accessor(obj)
	if err != nil {
		return
	}

	if !rr.isManagedByController(objMeta) {
		return
	}

	rr.logger.Debug(fmt.Sprintf("%s deleted", rr.resourceKind), "key", key)
	rr.metrics.TriggerByCounter(rr.resourceKind, DeleteEvent).Inc()

	rr.reconcileQ.Add(key)
}

func (rr *ResourceReconciler) onStatefulSetAdd(ss *appsv1.StatefulSet) {
	obj := rr.resolve(ss)
	if obj == nil {
		return
	}

	rr.logger.Debug("StatefulSet added")
	rr.metrics.TriggerByCounter("StatefulSet", AddEvent).Inc()

	rr.EnqueueForReconciliation(obj)
}

func (rr *ResourceReconciler) onDaemonSetAdd(ds *appsv1.DaemonSet) {
	obj := rr.resolve(ds)
	if obj == nil {
		return
	}

	rr.logger.Debug("DaemonSet added")

	rr.EnqueueForReconciliation(obj)
}

func (rr *ResourceReconciler) onStatefulSetUpdate(old, cur *appsv1.StatefulSet) {
	rr.logger.Debug("update handler", "resource", "statefulset", "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	if rr.DeletionInProgress(cur) {
		return
	}

	if !rr.hasObjectChanged(old, cur) {
		return
	}

	obj := rr.resolve(cur)
	if obj == nil {
		return
	}

	rr.logger.Debug("StatefulSet updated")
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

func (rr *ResourceReconciler) onDaemonSetUpdate(old, cur *appsv1.DaemonSet) {
	rr.logger.Debug("update handler", "resource", "daemonset", "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	if rr.DeletionInProgress(cur) {
		return
	}

	if !rr.hasObjectChanged(old, cur) {
		return
	}

	obj := rr.resolve(cur)
	if obj == nil {
		return
	}

	rr.logger.Debug("DaemonSet updated")
	if !rr.hasStateChanged(old, cur) {
		// If the daemonset state (spec, labels or annotations) hasn't
		// changed, the operator can only update the status subresource instead
		// of doing a full reconciliation.
		// TODO: Uncomment this when Prometheus Agent DaemonSet's status has been supported.
		// rr.EnqueueForStatus(obj)
		return
	}

	rr.EnqueueForReconciliation(obj)
}

func (rr *ResourceReconciler) onStatefulSetDelete(ss *appsv1.StatefulSet) {
	obj := rr.resolve(ss)
	if obj == nil {
		return
	}

	rr.logger.Debug("StatefulSet delete")
	rr.metrics.TriggerByCounter("StatefulSet", DeleteEvent).Inc()

	rr.EnqueueForReconciliation(obj)
}

func (rr *ResourceReconciler) onDaemonSetDelete(ds *appsv1.DaemonSet) {
	obj := rr.resolve(ds)
	if obj == nil {
		return
	}

	rr.logger.Debug("DaemonSet delete")

	rr.EnqueueForReconciliation(obj)
}

// EnqueueForReconciliation asks for reconciling the object.
func (rr *ResourceReconciler) EnqueueForReconciliation(obj metav1.Object) {
	if !rr.isManagedByController(obj) {
		return
	}

	rr.reconcileQ.Add(KeyForObject(obj))
}

// EnqueueForStatus asks for updating the status of the object.
func (rr *ResourceReconciler) EnqueueForStatus(obj metav1.Object) {
	if !rr.isManagedByController(obj) {
		return
	}

	rr.statusQ.Add(KeyForObject(obj))
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
	key, quit := rr.reconcileQ.Get()
	if quit {
		return false
	}

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
	utilruntime.HandleError(fmt.Errorf("sync %q failed: %w", key, err))
	rr.reconcileQ.AddRateLimited(key)

	return true
}

func (rr *ResourceReconciler) processNextStatusItem(ctx context.Context) bool {
	key, quit := rr.statusQ.Get()
	if quit {
		return false
	}

	defer rr.statusQ.Done(key)

	rr.statusTotal.Inc()
	err := rr.syncer.UpdateStatus(ctx, key)
	if err == nil {
		rr.statusQ.Forget(key)
		return true
	}

	rr.statusErrors.Inc()
	utilruntime.HandleError(fmt.Errorf("status %q failed: %w", key, err))
	rr.statusQ.AddRateLimited(key)

	return true
}

// isManagedByController returns true if the controller is the "owner" of the object.
// Whether it's owner is determined by the value of 'controllerID'
// annotation. If the value matches the controllerID then it owns it.
func (rr *ResourceReconciler) isManagedByController(obj metav1.Object) bool {
	var controllerID string

	if obj.GetAnnotations() != nil {
		controllerID = obj.GetAnnotations()[controllerIDAnnotation]
	}

	if controllerID != rr.controllerID {
		rr.logger.Debug("skipping object not managed by the controller", "object", KeyForObject(obj), "object_id", controllerID, "controller_id", rr.controllerID)
		return false
	}

	return true
}

// KeyForObject returns a string key identifying the given object.
// For cluster-scoped resources, the key is `<name>`.
// For namespace-scoped resources, the key is `<namespace>/<name>`.
func KeyForObject(o metav1.Object) string {
	if o.GetNamespace() == "" {
		return o.GetName()
	}

	return o.GetNamespace() + "/" + o.GetName()
}
