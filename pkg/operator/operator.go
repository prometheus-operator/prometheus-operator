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
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
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

// Metrics represents metrics associated to an operator.
type Metrics struct {
	reg prometheus.Registerer

	listCounter            prometheus.Counter
	listFailedCounter      prometheus.Counter
	watchCounter           prometheus.Counter
	watchFailedCounter     prometheus.Counter
	reconcileCounter       prometheus.Counter
	reconcileErrorsCounter prometheus.Counter
	stsDeleteCreateCounter prometheus.Counter
	// triggerByCounter is a set of counters keeping track of the amount
	// of times Prometheus Operator was triggered to reconcile its created
	// objects. It is split in the dimensions of Kubernetes objects and
	// corresponding actions (add, delete, update).
	triggerByCounter *prometheus.CounterVec
	ready            prometheus.Gauge

	// mtx protects all fields below.
	mtx       sync.RWMutex
	syncs     map[string]bool
	resources map[resourceKey]map[string]int
}

type resourceKey struct {
	resource string
	state    resourceState
}

// NewMetrics initializes operator metrics and registers them with the given registerer.
// All metrics have a "controller=<name>" label.
func NewMetrics(name string, r prometheus.Registerer) *Metrics {
	reg := prometheus.WrapRegistererWith(prometheus.Labels{"controller": name}, r)
	m := Metrics{
		reg: reg,
		reconcileCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prometheus_operator_reconcile_operations_total",
			Help: "Total number of reconcile operations",
		}),
		reconcileErrorsCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prometheus_operator_reconcile_errors_total",
			Help: "Number of errors that occurred during reconcile operations",
		}),
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

		syncs:     make(map[string]bool),
		resources: make(map[resourceKey]map[string]int),
	}

	m.reg.MustRegister(
		m.reconcileCounter,
		m.reconcileErrorsCounter,
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

// ReconcileCounter returns a counter to track attempted reconciliations.
func (m *Metrics) ReconcileCounter() prometheus.Counter {
	return m.reconcileCounter
}

// ReconcileErrorsCounter returns a counter to track reconciliation errors.
func (m *Metrics) ReconcileErrorsCounter() prometheus.Counter {
	return m.reconcileErrorsCounter
}

// StsDeleteCreateCounter returns a counter to track statefulset's recreations.
func (m *Metrics) StsDeleteCreateCounter() prometheus.Counter {
	return m.stsDeleteCreateCounter
}

// TriggerByCounter returns a counter to track operator actions by operation (add/delete/update) and action.
func (m *Metrics) TriggerByCounter(triggeredBy, action string) prometheus.Counter {
	return m.triggerByCounter.WithLabelValues(triggeredBy, action)
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

// SetSyncStatus tracks the status of the last sync operation for the given object.
func (m *Metrics) SetSyncStatus(objKey string, success bool) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.syncs[objKey] = success
}

// ForgetObject removes the metrics tracked for the given object's key.
// It should be called when the controller detects that the object has been deleted.
func (m *Metrics) ForgetObject(objKey string) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	delete(m.syncs, objKey)

	for k := range m.resources {
		delete(m.resources[k], objKey)
	}
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
	ch <- syncsDesc
}

// Collect implements the prometheus.Collector interface.
func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	var ok, failed float64
	for _, success := range m.syncs {
		if success {
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
	ret, err := i.next.List(options)
	if err != nil {
		i.listFailed.Inc()
	}
	return ret, err
}

// Watch implements the cache.ListerWatcher interface.
func (i *instrumentedListerWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	i.watchTotal.Inc()
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

// WaitForCacheSync synchronizes the informer's cache and will log a warning
// every minute if the operation hasn't completed yet.
// Under normal circumstances, the cache sync should be fast. If it takes more
// than 1 minute, it means that something is stuck and the message will
// indicate to the admin which informer is the culprit.
// See https://github.com/prometheus-operator/prometheus-operator/issues/3347.
func WaitForNamedCacheSync(ctx context.Context, controllerName string, logger log.Logger, inf cache.SharedIndexInformer) bool {
	ctx, cancel := context.WithCancel(ctx)

	done := make(chan struct{})
	go func() {
		t := time.NewTicker(time.Minute)
		defer t.Stop()

		select {
		case <-t.C:
			level.Warn(logger).Log("msg", "cache sync not yet completed")
		case <-ctx.Done():
			close(done)
		}
	}()

	ok := cache.WaitForNamedCacheSync(controllerName, ctx.Done(), inf.HasSynced)
	if !ok {
		level.Error(logger).Log("msg", "failed to sync cache")
	} else {
		level.Debug(logger).Log("msg", "successfully synced cache")
	}

	// Stop the logging goroutine and wait for its exit.
	cancel()
	<-done

	return ok
}
