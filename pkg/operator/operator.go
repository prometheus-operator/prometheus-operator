// Copyright 2019 The prometheus-operator Authors
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
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/scheme"
)

const (
	PrometheusOperatorFieldManager = "PrometheusOperator"

	InvalidConfigurationEvent = "InvalidConfiguration"
)

var (
	syncsDesc = prometheus.NewDesc(
		"prometheus_operator_syncs",
		"Number of objects per sync status (ok/failed)",
		[]string{"status"},
		nil,
	)
	resourcesDesc = prometheus.NewDesc(
		"prometheus_operator_managed_resources",
		"Number of resources managed by the operator's controller per state (selected/rejected)",
		[]string{"resource", "state"},
		nil,
	)
)

type ReconciliationStatus struct {
	err error
}

func (rs ReconciliationStatus) Reason() string {
	if rs.Ok() {
		return ""
	}

	return "ReconciliationFailed"
}

func (rs ReconciliationStatus) Message() string {
	if rs.Ok() {
		return ""
	}

	return rs.err.Error()
}

func (rs ReconciliationStatus) Ok() bool {
	return rs.err == nil
}

// ReconciliationTracker tracks reconciliation status per object.
//
// It only uses their `<namespace>/<name>` key to identify objects.
//
// The zero ReconciliationTracker is ready to use.
type ReconciliationTracker struct {
	once sync.Once

	// mtx protects all fields below.
	mtx            sync.RWMutex
	statusByObject map[string]ReconciliationStatus
	refTracker     map[string]ReferenceTracker
}

// ReferenceTracker returns true if it has a reference to the object.
type ReferenceTracker interface {
	Has(runtime.Object) bool
}

func (rt *ReconciliationTracker) init() {
	rt.once.Do(func() {
		rt.statusByObject = map[string]ReconciliationStatus{}
		rt.refTracker = map[string]ReferenceTracker{}
	})
}

// HasRefTo returns true if the object identified by key has a direct or
// indirect reference to obj (secret or configmap).
func (rt *ReconciliationTracker) HasRefTo(key string, obj runtime.Object) bool {
	rt.mtx.Lock()
	defer rt.mtx.Unlock()

	refTracker, found := rt.refTracker[key]
	if !found {
		return false
	}

	return refTracker.Has(obj)
}

// UpdateReferenceTracker updates the reference tracker for the object identified by key.
func (rt *ReconciliationTracker) UpdateReferenceTracker(key string, refTracker ReferenceTracker) {
	rt.init()
	rt.mtx.Lock()
	defer rt.mtx.Unlock()

	rt.refTracker[key] = refTracker
}

// SetStatus updates the last reconciliation status for the object identified by key.
func (rt *ReconciliationTracker) SetStatus(key string, err error) {
	rt.init()
	rt.mtx.Lock()
	defer rt.mtx.Unlock()

	rt.statusByObject[key] = ReconciliationStatus{err: err}
}

// GetStatus returns the last reconciliation status for the given object.
// The second value indicates whether the object is known or not.
func (rt *ReconciliationTracker) getStatus(k string) (ReconciliationStatus, bool) {
	rt.mtx.Lock()
	defer rt.mtx.Unlock()

	s, found := rt.statusByObject[k]
	if !found {
		return ReconciliationStatus{}, false
	}

	return s, true
}

// GetCondition returns a monitoringv1.Condition for the last-known
// reconciliation status of the given object.
func (rt *ReconciliationTracker) GetCondition(k string, gen int64) monitoringv1.Condition {
	condition := monitoringv1.Condition{
		Type:   monitoringv1.Reconciled,
		Status: monitoringv1.ConditionTrue,
		LastTransitionTime: metav1.Time{
			Time: time.Now().UTC(),
		},
		ObservedGeneration: gen,
	}

	reconciliationStatus, found := rt.getStatus(k)
	if !found {
		condition.Status = monitoringv1.ConditionUnknown
		condition.Reason = "NotFound"
		condition.Message = fmt.Sprintf("object %q not found", k)
	} else {
		if !reconciliationStatus.Ok() {
			condition.Status = monitoringv1.ConditionFalse
		}
		condition.Reason = reconciliationStatus.Reason()
		condition.Message = reconciliationStatus.Message()
	}

	return condition
}

// ForgetObject removes the given object from the tracker.
// It should be called when the controller detects that the object has been deleted.
func (rt *ReconciliationTracker) ForgetObject(key string) {
	rt.mtx.Lock()
	defer rt.mtx.Unlock()

	if rt.statusByObject == nil {
		return
	}

	delete(rt.statusByObject, key)
	delete(rt.refTracker, key)
}

// Describe implements the prometheus.Collector interface.
func (rt *ReconciliationTracker) Describe(ch chan<- *prometheus.Desc) {
	ch <- syncsDesc
}

// Collect implements the prometheus.Collector interface.
func (rt *ReconciliationTracker) Collect(ch chan<- prometheus.Metric) {
	rt.mtx.RLock()
	defer rt.mtx.RUnlock()

	var ok, failed float64
	for _, st := range rt.statusByObject {
		if st.Ok() {
			ok++
		} else {
			failed++
		}
	}

	ch <- prometheus.MustNewConstMetric(
		syncsDesc,
		prometheus.GaugeValue,
		ok,
		"ok",
	)
	ch <- prometheus.MustNewConstMetric(
		syncsDesc,
		prometheus.GaugeValue,
		failed,
		"failed",
	)
}

// Metrics represents metrics associated to an operator.
type Metrics struct {
	reg prometheus.Registerer

	listCounter            prometheus.Counter
	listFailedCounter      prometheus.Counter
	watchCounter           prometheus.Counter
	watchFailedCounter     prometheus.Counter
	stsDeleteCreateCounter prometheus.Counter
	// triggerByCounter is a set of counters keeping track of the amount
	// of times Prometheus Operator was triggered to reconcile its created
	// objects. It is split in the dimensions of Kubernetes objects and
	// corresponding actions (add, delete, update).
	triggerByCounter *prometheus.CounterVec
	ready            prometheus.Gauge

	// mtx protects all fields below.
	mtx       sync.RWMutex
	resources map[resourceKey]map[string]int
}

type resourceKey struct {
	resource string
	state    resourceState
}

// NewMetrics initializes operator metrics and registers them with the given registerer.
func NewMetrics(r prometheus.Registerer) *Metrics {
	m := Metrics{
		reg: r,
		triggerByCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "prometheus_operator_triggered_total",
			Help: "Number of times a Kubernetes object add, delete or update event" +
				" triggered the Prometheus Operator to reconcile an object",
		}, []string{"triggered_by", "action"}),
		stsDeleteCreateCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prometheus_operator_reconcile_sts_delete_create_total",
			Help: "Number of times that reconciling a statefulset required deleting and re-creating it",
		}),
		listCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prometheus_operator_list_operations_total",
			Help: "Total number of list operations",
		}),
		listFailedCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prometheus_operator_list_operations_failed_total",
			Help: "Total number of list operations that failed",
		}),
		watchCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prometheus_operator_watch_operations_total",
			Help: "Total number of watch operations",
		}),
		watchFailedCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prometheus_operator_watch_operations_failed_total",
			Help: "Total number of watch operations that failed",
		}),
		ready: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "prometheus_operator_ready",
			Help: "1 when the controller is ready to reconcile resources, 0 otherwise",
		}),

		resources: make(map[resourceKey]map[string]int),
	}

	m.reg.MustRegister(
		m.triggerByCounter,
		m.stsDeleteCreateCounter,
		m.listCounter,
		m.listFailedCounter,
		m.watchCounter,
		m.watchFailedCounter,
		m.ready,
		&m,
	)

	return &m
}

type EventRecorderFactory func(client kubernetes.Interface, component string) record.EventRecorder

func NewEventRecorderFactory(emitEvents bool) EventRecorderFactory {
	return func(client kubernetes.Interface, component string) record.EventRecorder {
		eventBroadcaster := record.NewBroadcaster()
		eventBroadcaster.StartStructuredLogging(0)

		if emitEvents {
			eventBroadcaster.StartRecordingToSink(&typedv1.EventSinkImpl{Interface: client.CoreV1().Events("")})
		}

		return eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: component})
	}
}

// StsDeleteCreateCounter returns a counter to track statefulset's recreations.
func (m *Metrics) StsDeleteCreateCounter() prometheus.Counter {
	return m.stsDeleteCreateCounter
}

type HandlerEvent string

const (
	AddEvent    = HandlerEvent("add")
	DeleteEvent = HandlerEvent("delete")
	UpdateEvent = HandlerEvent("update")
)

// TriggerByCounter returns a counter to track operator actions by resource type and action (add/delete/update).
func (m *Metrics) TriggerByCounter(triggeredBy string, action HandlerEvent) prometheus.Counter {
	return m.triggerByCounter.With(prometheus.Labels{"triggered_by": triggeredBy, "action": string(action)})
}

const (
	selected int = iota
	rejected
)

type resourceState int

func (r resourceState) String() string {
	switch int(r) {
	case selected:
		return "selected"
	case rejected:
		return "rejected"
	}
	return ""
}

// SetSelectedResources sets the number of resources that the controller selected for the given object's key.
func (m *Metrics) SetSelectedResources(objKey, resource string, v int) {
	m.setResources(objKey, resourceKey{resource: resource, state: resourceState(selected)}, v)
}

// SetRejectedResources sets the number of resources that the controller rejected for the given object's key.
func (m *Metrics) SetRejectedResources(objKey, resource string, v int) {
	m.setResources(objKey, resourceKey{resource: resource, state: resourceState(rejected)}, v)
}

func (m *Metrics) setResources(objKey string, resKey resourceKey, v int) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if _, found := m.resources[resKey]; !found {
		m.resources[resourceKey{resource: resKey.resource, state: resourceState(selected)}] = make(map[string]int)
		m.resources[resourceKey{resource: resKey.resource, state: resourceState(rejected)}] = make(map[string]int)
	}

	m.resources[resKey][objKey] = v
}

// Ready returns a gauge to track whether the controller is ready or not.
func (m *Metrics) Ready() prometheus.Gauge {
	return m.ready
}

// MustRegister registers metrics with the Metrics registerer.
func (m *Metrics) MustRegister(metrics ...prometheus.Collector) {
	m.reg.MustRegister(metrics...)
}

// Describe implements the prometheus.Collector interface.
func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- resourcesDesc
}

// Collect implements the prometheus.Collector interface.
func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	for rKey := range m.resources {
		var total int
		for _, v := range m.resources[rKey] {
			total += v
		}
		ch <- prometheus.MustNewConstMetric(
			resourcesDesc,
			prometheus.GaugeValue,
			float64(total),
			rKey.resource,
			rKey.state.String(),
		)
	}
}

type instrumentedListerWatcher struct {
	next        cache.ListerWatcher
	listTotal   prometheus.Counter
	listFailed  prometheus.Counter
	watchTotal  prometheus.Counter
	watchFailed prometheus.Counter
}

// NewInstrumentedListerWatcher returns a cache.ListerWatcher with instrumentation.
func (m *Metrics) NewInstrumentedListerWatcher(lw cache.ListerWatcher) cache.ListerWatcher {
	return &instrumentedListerWatcher{
		next:        lw,
		listTotal:   m.listCounter,
		listFailed:  m.listFailedCounter,
		watchTotal:  m.watchCounter,
		watchFailed: m.watchFailedCounter,
	}
}

// List implements the cache.ListerWatcher interface.
func (i *instrumentedListerWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	i.listTotal.Inc()
	//nolint:staticcheck // Ignore SA1019 the function is deprecated.
	ret, err := i.next.List(options)
	if err != nil {
		i.listFailed.Inc()
	}
	return ret, err
}

// Watch implements the cache.ListerWatcher interface.
func (i *instrumentedListerWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	i.watchTotal.Inc()
	//nolint:staticcheck // Ignore SA1019 the function is deprecated.
	ret, err := i.next.Watch(options)
	if err != nil {
		i.watchFailed.Inc()
	}
	return ret, err
}

// SanitizeSTS removes values for APIVersion and Kind from the VolumeClaimTemplates.
// This prevents update failures due to these fields changing when applied.
// See https://github.com/kubernetes/kubernetes/issues/87583
func SanitizeSTS(sts *appsv1.StatefulSet) {
	for i := range sts.Spec.VolumeClaimTemplates {
		sts.Spec.VolumeClaimTemplates[i].APIVersion = ""
		sts.Spec.VolumeClaimTemplates[i].Kind = ""
	}
}

// WaitForNamedCacheSync synchronizes the informer's cache and will log a
// warning every minute if the operation hasn't completed yet, until it reaches
// a timeout of 10 minutes.
// Under normal circumstances, the cache sync should be fast. If it takes more
// than 1 minute, it means that something is stuck and the message will
// indicate to the admin which informer is the culprit.
// See https://github.com/prometheus-operator/prometheus-operator/issues/3347.
func WaitForNamedCacheSync(ctx context.Context, controllerName string, logger *slog.Logger, inf cache.SharedIndexInformer) bool {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	t := time.NewTicker(time.Minute)
	defer t.Stop()

	go func() {
		for {
			select {
			case <-t.C:
				logger.Warn("cache sync not yet completed")
			case <-ctx.Done():
				return
			}
		}
	}()

	ok := cache.WaitForNamedCacheSync(controllerName, ctx.Done(), inf.HasSynced)
	if !ok {
		logger.Error("failed to sync cache")
	} else {
		logger.Debug("successfully synced cache")
	}

	return ok
}

// ConfigMapGVK returns the GroupVersionKind representing ConfigMap objects.
func ConfigMapGVK() schema.GroupVersionKind {
	return v1.SchemeGroupVersion.WithKind("ConfigMap")
}

// SecretGVK returns the GroupVersionKind representing Secret objects.
func SecretGVK() schema.GroupVersionKind {
	return v1.SchemeGroupVersion.WithKind("Secret")
}
