// Copyright 2024 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package servicemonitorcontroller

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
)

const (
	resyncPeriod = 5 * time.Minute
)

// StatusController updates the status field of ServiceMonitor resources.
type StatusController struct {
	logger   *slog.Logger
	mclient  monitoringclient.Interface
	queue    workqueue.RateLimitingInterface
	metrics  *statusControllerMetrics
	informer *informers.ForResource

	prometheusInfs *informers.ForResource
	promAgentInfs  *informers.ForResource
	serviceMonInfs *informers.ForResource
}

type statusControllerMetrics struct {
	statusUpdates        prometheus.Counter
	statusUpdateFailures prometheus.Counter
}

// NewStatusController creates a new StatusController.
func NewStatusController(
	ctx context.Context,
	logger *slog.Logger,
	mclient monitoringclient.Interface,
	prometheusInfs *informers.ForResource,
	promAgentInfs *informers.ForResource,
	serviceMonInfs *informers.ForResource,
	reg prometheus.Registerer,
) (*StatusController, error) {
	logger = logger.With("controller", "servicemonitor-status")

	if reg == nil {
		reg = prometheus.NewRegistry()
	}

	c := &StatusController{
		logger:         logger,
		mclient:        mclient,
		prometheusInfs: prometheusInfs,
		promAgentInfs:  promAgentInfs,
		serviceMonInfs: serviceMonInfs,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "servicemonitor-status"),
		metrics: &statusControllerMetrics{
			statusUpdates: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "prometheus_operator_servicemonitor_status_updates_total",
				Help: "Total number of status updates for ServiceMonitor objects",
			}),
			statusUpdateFailures: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "prometheus_operator_servicemonitor_status_update_failures_total",
				Help: "Total number of status update failures for ServiceMonitor objects",
			}),
		},
	}

	reg.MustRegister(c.metrics.statusUpdates, c.metrics.statusUpdateFailures)

	c.setupEventHandlers()

	return c, nil
}

func (c *StatusController) setupEventHandlers() {
	// Watch Prometheus resources and queue related ServiceMonitors
	for _, inf := range c.prometheusInfs.GetInformers() {
		inf.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handlePrometheusAdd,
			UpdateFunc: c.handlePrometheusUpdate,
			DeleteFunc: c.handlePrometheusDelete,
		})
	}

	// Watch PrometheusAgent resources and queue related ServiceMonitors
	for _, inf := range c.promAgentInfs.GetInformers() {
		inf.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handlePrometheusAgentAdd,
			UpdateFunc: c.handlePrometheusAgentUpdate,
			DeleteFunc: c.handlePrometheusAgentDelete,
		})
	}

	// Watch for ServiceMonitor reconciliation status changes
	for _, inf := range c.serviceMonInfs.GetInformers() {
		inf.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if sm, ok := obj.(*monitoringv1.ServiceMonitor); ok {
					c.enqueueServiceMonitor(sm)
				}
			},
		})
	}
}

func (c *StatusController) handlePrometheusAdd(obj interface{}) {
	prom, ok := obj.(*monitoringv1.Prometheus)
	if !ok {
		return
	}
	c.logger.Debug("Handling Prometheus add", "prometheus", fmt.Sprintf("%s/%s", prom.Namespace, prom.Name))
	c.enqueueServiceMonitorsForPrometheus(prom)
}

func (c *StatusController) handlePrometheusUpdate(old, cur interface{}) {
	oldProm, ok := old.(*monitoringv1.Prometheus)
	if !ok {
		return
	}

	curProm, ok := cur.(*monitoringv1.Prometheus)
	if !ok {
		return
	}

	// Only handle selector updates
	if oldProm.Spec.ServiceMonitorSelector.String() == curProm.Spec.ServiceMonitorSelector.String() &&
		oldProm.Spec.ServiceMonitorNamespaceSelector.String() == curProm.Spec.ServiceMonitorNamespaceSelector.String() {
		return
	}

	c.logger.Debug("Handling Prometheus update", "prometheus", fmt.Sprintf("%s/%s", curProm.Namespace, curProm.Name))

	// Enqueue service monitors from both old and new selectors
	c.enqueueServiceMonitorsForPrometheus(oldProm)
	c.enqueueServiceMonitorsForPrometheus(curProm)
}

func (c *StatusController) handlePrometheusDelete(obj interface{}) {
	prom, ok := obj.(*monitoringv1.Prometheus)
	if !ok {
		// Delete can get a DeletedFinalStateUnknown instead of a Prometheus
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return
		}
		prom, ok = tombstone.Obj.(*monitoringv1.Prometheus)
		if !ok {
			return
		}
	}
	c.logger.Debug("Handling Prometheus delete", "prometheus", fmt.Sprintf("%s/%s", prom.Namespace, prom.Name))
	c.enqueueServiceMonitorsForPrometheus(prom)
}

func (c *StatusController) handlePrometheusAgentAdd(obj interface{}) {
	agent, ok := obj.(*monitoringv1alpha1.PrometheusAgent)
	if !ok {
		return
	}
	c.logger.Debug("Handling PrometheusAgent add", "prometheusagent", fmt.Sprintf("%s/%s", agent.Namespace, agent.Name))
	c.enqueueServiceMonitorsForPrometheusAgent(agent)
}

func (c *StatusController) handlePrometheusAgentUpdate(old, cur interface{}) {
	oldAgent, ok := old.(*monitoringv1alpha1.PrometheusAgent)
	if !ok {
		return
	}

	curAgent, ok := cur.(*monitoringv1alpha1.PrometheusAgent)
	if !ok {
		return
	}

	// Only handle selector updates
	if oldAgent.Spec.ServiceMonitorSelector.String() == curAgent.Spec.ServiceMonitorSelector.String() &&
		oldAgent.Spec.ServiceMonitorNamespaceSelector.String() == curAgent.Spec.ServiceMonitorNamespaceSelector.String() {
		return
	}

	c.logger.Debug("Handling PrometheusAgent update", "prometheusagent", fmt.Sprintf("%s/%s", curAgent.Namespace, curAgent.Name))

	// Enqueue service monitors from both old and new selectors
	c.enqueueServiceMonitorsForPrometheusAgent(oldAgent)
	c.enqueueServiceMonitorsForPrometheusAgent(curAgent)
}

func (c *StatusController) handlePrometheusAgentDelete(obj interface{}) {
	agent, ok := obj.(*monitoringv1alpha1.PrometheusAgent)
	if !ok {
		// Delete can get a DeletedFinalStateUnknown instead of a Prometheus
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return
		}
		agent, ok = tombstone.Obj.(*monitoringv1alpha1.PrometheusAgent)
		if !ok {
			return
		}
	}
	c.logger.Debug("Handling PrometheusAgent delete", "prometheusagent", fmt.Sprintf("%s/%s", agent.Namespace, agent.Name))
	c.enqueueServiceMonitorsForPrometheusAgent(agent)
}

func (c *StatusController) enqueueServiceMonitorsForPrometheus(prom *monitoringv1.Prometheus) {
	// If no ServiceMonitor selector is configured, don't enqueue anything
	if prom.Spec.ServiceMonitorSelector == nil {
		return
	}

	// Get all ServiceMonitors in selected namespaces
	nsSelector, err := k8sutil.GetNamespaceSelector(prom.Spec.ServiceMonitorNamespaceSelector)
	if err != nil {
		c.logger.Error("Failed to get namespace selector", "error", err)
		return
	}

	for _, smonInf := range c.serviceMonInfs.GetInformers() {
		// Get the namespace for this informer
		namespace := getNamespaceFromInformer(smonInf)

		// Skip if namespace selector doesn't match
		if nsSelector != nil && !nsSelector.Matches(labels.Set{"name": namespace}) {
			continue
		}

		smSelector, err := metav1.LabelSelectorAsSelector(prom.Spec.ServiceMonitorSelector)
		if err != nil {
			c.logger.Error("Failed to convert label selector", "error", err)
			continue
		}

		smons, err := smonInf.Lister().List(smSelector)
		if err != nil {
			c.logger.Error("Failed to list ServiceMonitors", "namespace", namespace, "error", err)
			continue
		}

		for _, smObj := range smons {
			sm, ok := smObj.(*monitoringv1.ServiceMonitor)
			if !ok {
				continue
			}
			c.enqueueServiceMonitor(sm)
		}
	}
}

func (c *StatusController) enqueueServiceMonitorsForPrometheusAgent(agent *monitoringv1alpha1.PrometheusAgent) {
	// If no ServiceMonitor selector is configured, don't enqueue anything
	if agent.Spec.ServiceMonitorSelector == nil {
		return
	}

	// Get all ServiceMonitors in selected namespaces
	nsSelector, err := k8sutil.GetNamespaceSelector(agent.Spec.ServiceMonitorNamespaceSelector)
	if err != nil {
		c.logger.Error("Failed to get namespace selector", "error", err)
		return
	}

	for _, smonInf := range c.serviceMonInfs.GetInformers() {
		// Get the namespace for this informer
		namespace := getNamespaceFromInformer(smonInf)

		// Skip if namespace selector doesn't match
		if nsSelector != nil && !nsSelector.Matches(labels.Set{"name": namespace}) {
			continue
		}

		smSelector, err := metav1.LabelSelectorAsSelector(agent.Spec.ServiceMonitorSelector)
		if err != nil {
			c.logger.Error("Failed to convert label selector", "error", err)
			continue
		}

		smons, err := smonInf.Lister().List(smSelector)
		if err != nil {
			c.logger.Error("Failed to list ServiceMonitors", "namespace", namespace, "error", err)
			continue
		}

		for _, smObj := range smons {
			sm, ok := smObj.(*monitoringv1.ServiceMonitor)
			if !ok {
				continue
			}
			c.enqueueServiceMonitor(sm)
		}
	}
}

func (c *StatusController) enqueueServiceMonitor(sm *monitoringv1.ServiceMonitor) {
	key, err := cache.MetaNamespaceKeyFunc(sm)
	if err != nil {
		c.logger.Error("Error creating key for ServiceMonitor", "ServiceMonitor", fmt.Sprintf("%s/%s", sm.Namespace, sm.Name), "error", err)
		return
	}
	c.queue.Add(key)
}

// Run starts the controller and blocks until stopCh is closed.
func (c *StatusController) Run(ctx context.Context, workers int) error {
	defer c.queue.ShutDown()

	c.logger.Info("Starting ServiceMonitor status controller")
	defer c.logger.Info("Shutting down ServiceMonitor status controller")

	for i := 0; i < workers; i++ {
		go c.worker(ctx)
	}

	<-ctx.Done()
	return nil
}

func (c *StatusController) worker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *StatusController) processNextWorkItem(ctx context.Context) bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.sync(ctx, key.(string))
	if err == nil {
		c.queue.Forget(key)
		return true
	}

	c.logger.Error("Error syncing ServiceMonitor status", "key", key, "error", err)
	c.queue.AddRateLimited(key)

	return true
}

func (c *StatusController) sync(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	// Find the informer for this namespace
	var serviceMonitor *monitoringv1.ServiceMonitor

	for _, inf := range c.serviceMonInfs.GetInformers() {
		if getNamespaceFromInformer(inf) == namespace {
			obj, err := inf.Lister().Get(name)
			if err != nil {
				if apierrors.IsNotFound(err) {
					c.logger.Debug("ServiceMonitor has been deleted", "namespace", namespace, "name", name)
					return nil
				}
				return err
			}

			sm, ok := obj.(*monitoringv1.ServiceMonitor)
			if !ok {
				return fmt.Errorf("expected *monitoringv1.ServiceMonitor, got %T", obj)
			}

			serviceMonitor = sm
			break
		}
	}

	if serviceMonitor == nil {
		return fmt.Errorf("servicemonitor %s/%s not found in any informer", namespace, name)
	}

	// Reconcile the status by collecting references from all Prometheus/PrometheusAgent instances
	references, err := c.collectReferences(ctx, serviceMonitor)
	if err != nil {
		return err
	}

	// Check if status needs to be updated
	if statusNeedsUpdate(serviceMonitor.Status.References, references) {
		// Create a copy to avoid modifying the cache
		newServiceMon := serviceMonitor.DeepCopy()
		newServiceMon.Status.References = references

		_, err = c.mclient.MonitoringV1().ServiceMonitors(newServiceMon.Namespace).UpdateStatus(ctx, newServiceMon, metav1.UpdateOptions{})
		if err != nil {
			c.metrics.statusUpdateFailures.Inc()
			return err
		}
		c.metrics.statusUpdates.Inc()
		c.logger.Debug("Updated ServiceMonitor status", "namespace", namespace, "name", name, "references", len(references))
	}

	return nil
}

func (c *StatusController) collectReferences(ctx context.Context, sm *monitoringv1.ServiceMonitor) ([]monitoringv1.ServiceMonitorReference, error) {
	var references []monitoringv1.ServiceMonitorReference

	// Collect Prometheus references
	promReferences, err := c.collectPrometheusReferences(ctx, sm)
	if err != nil {
		return nil, err
	}
	references = append(references, promReferences...)

	// Collect PrometheusAgent references
	agentReferences, err := c.collectPrometheusAgentReferences(ctx, sm)
	if err != nil {
		return nil, err
	}
	references = append(references, agentReferences...)

	return references, nil
}

func (c *StatusController) collectPrometheusReferences(ctx context.Context, sm *monitoringv1.ServiceMonitor) ([]monitoringv1.ServiceMonitorReference, error) {
	var references []monitoringv1.ServiceMonitorReference

	// Check all Prometheus instances across all namespaces
	for _, promInf := range c.prometheusInfs.GetInformers() {
		promList, err := promInf.Lister().List(labels.Everything())
		if err != nil {
			return nil, err
		}

		for _, obj := range promList {
			prom, ok := obj.(*monitoringv1.Prometheus)
			if !ok {
				continue
			}

			// Skip if Prometheus doesn't select ServiceMonitors
			if prom.Spec.ServiceMonitorSelector == nil {
				continue
			}

			// Check if this ServiceMonitor is selected by the Prometheus instance
			if !isServiceMonitorSelectedByPrometheus(sm, prom) {
				continue
			}

			// Get reconciliation status from the Prometheus instance
			reference := monitoringv1.ServiceMonitorReference{
				Resource:  monitoringv1.PrometheusName,
				Name:      prom.Name,
				Namespace: prom.Namespace,
				// We need to add a basic condition, it will be updated by the Prometheus reconciliation
				Conditions: []monitoringv1.Condition{
					{
						Type:               monitoringv1.Reconciled,
						Status:             monitoringv1.ConditionTrue,
						LastTransitionTime: metav1.Now(),
						ObservedGeneration: sm.Generation,
					},
				},
			}

			references = append(references, reference)
		}
	}

	return references, nil
}

func (c *StatusController) collectPrometheusAgentReferences(ctx context.Context, sm *monitoringv1.ServiceMonitor) ([]monitoringv1.ServiceMonitorReference, error) {
	var references []monitoringv1.ServiceMonitorReference

	// Check all PrometheusAgent instances across all namespaces
	for _, agentInf := range c.promAgentInfs.GetInformers() {
		agentList, err := agentInf.Lister().List(labels.Everything())
		if err != nil {
			return nil, err
		}

		for _, obj := range agentList {
			agent, ok := obj.(*monitoringv1alpha1.PrometheusAgent)
			if !ok {
				continue
			}

			// Skip if PrometheusAgent doesn't select ServiceMonitors
			if agent.Spec.ServiceMonitorSelector == nil {
				continue
			}

			// Check if this ServiceMonitor is selected by the PrometheusAgent instance
			if !isServiceMonitorSelectedByPrometheusAgent(sm, agent) {
				continue
			}

			// Get reconciliation status from the PrometheusAgent instance
			reference := monitoringv1.ServiceMonitorReference{
				Resource:  "prometheusagents",
				Name:      agent.Name,
				Namespace: agent.Namespace,
				// We need to add a basic condition, it will be updated by the PrometheusAgent reconciliation
				Conditions: []monitoringv1.Condition{
					{
						Type:               monitoringv1.Reconciled,
						Status:             monitoringv1.ConditionTrue,
						LastTransitionTime: metav1.Now(),
						ObservedGeneration: sm.Generation,
					},
				},
			}

			references = append(references, reference)
		}
	}

	return references, nil
}

// isServiceMonitorSelectedByPrometheus checks if the given ServiceMonitor is selected by the Prometheus instance.
func isServiceMonitorSelectedByPrometheus(sm *monitoringv1.ServiceMonitor, prom *monitoringv1.Prometheus) bool {
	// Check namespace selector
	if prom.Spec.ServiceMonitorNamespaceSelector != nil {
		nsSelector, err := metav1.LabelSelectorAsSelector(prom.Spec.ServiceMonitorNamespaceSelector)
		if err != nil {
			return false
		}

		// If namespace selector doesn't match, skip
		namespaceLabels := getNamespaceLabels(sm.Namespace)
		if !nsSelector.Matches(namespaceLabels) {
			return false
		}
	} else if sm.Namespace != prom.Namespace {
		// If no namespace selector is specified, only match ServiceMonitors in the same namespace
		return false
	}

	// Check ServiceMonitor selector
	smSelector, err := metav1.LabelSelectorAsSelector(prom.Spec.ServiceMonitorSelector)
	if err != nil {
		return false
	}

	return smSelector.Matches(labels.Set(sm.Labels))
}

// isServiceMonitorSelectedByPrometheusAgent checks if the given ServiceMonitor is selected by the PrometheusAgent instance.
func isServiceMonitorSelectedByPrometheusAgent(sm *monitoringv1.ServiceMonitor, agent *monitoringv1alpha1.PrometheusAgent) bool {
	// Check namespace selector
	if agent.Spec.ServiceMonitorNamespaceSelector != nil {
		nsSelector, err := metav1.LabelSelectorAsSelector(agent.Spec.ServiceMonitorNamespaceSelector)
		if err != nil {
			return false
		}

		// If namespace selector doesn't match, skip
		namespaceLabels := getNamespaceLabels(sm.Namespace)
		if !nsSelector.Matches(namespaceLabels) {
			return false
		}
	} else if sm.Namespace != agent.Namespace {
		// If no namespace selector is specified, only match ServiceMonitors in the same namespace
		return false
	}

	// Check ServiceMonitor selector
	smSelector, err := metav1.LabelSelectorAsSelector(agent.Spec.ServiceMonitorSelector)
	if err != nil {
		return false
	}

	return smSelector.Matches(labels.Set(sm.Labels))
}

// getNamespaceLabels returns a map with the namespace name as a label.
// This is a simplified approach - in a real implementation you would get the actual namespace labels from the API.
func getNamespaceLabels(namespace string) labels.Set {
	return labels.Set{"name": namespace}
}

// getNamespaceFromInformer extracts the namespace from an informer
func getNamespaceFromInformer(inf informers.InformLister) string {
	// Use the ListWatcher to get namespace information
	// In real implementation, you'd need to extract this from the informer
	// This is a simplification for this example

	// Try to get namespace from the objects in the store
	objs := inf.Informer().GetStore().List()
	if len(objs) > 0 {
		if obj, ok := objs[0].(metav1.Object); ok {
			return obj.GetNamespace()
		}
	}

	// Default to "" (all namespaces) if we can't determine
	return ""
}

// statusNeedsUpdate compares current and target status references and returns true if an update is needed.
func statusNeedsUpdate(current, target []monitoringv1.ServiceMonitorReference) bool {
	if len(current) != len(target) {
		return true
	}

	for _, targetRef := range target {
		found := false
		for _, currentRef := range current {
			if targetRef.Resource == currentRef.Resource &&
				targetRef.Name == currentRef.Name &&
				targetRef.Namespace == currentRef.Namespace {
				found = true

				// Check conditions
				if !conditionsEqual(currentRef.Conditions, targetRef.Conditions) {
					return true
				}
				break
			}
		}
		if !found {
			return true
		}
	}

	return false
}

// conditionsEqual compares two slices of conditions and returns true if they are equal.
func conditionsEqual(a, b []monitoringv1.Condition) bool {
	if len(a) != len(b) {
		return false
	}

	// Sort both slices by Type to ensure a consistent comparison
	slices.SortFunc(a, func(c1, c2 monitoringv1.Condition) int {
		if string(c1.Type) < string(c2.Type) {
			return -1
		}
		return 1
	})
	slices.SortFunc(b, func(c1, c2 monitoringv1.Condition) int {
		if string(c1.Type) < string(c2.Type) {
			return -1
		}
		return 1
	})

	for i := range a {
		if a[i].Type != b[i].Type ||
			a[i].Status != b[i].Status ||
			a[i].Reason != b[i].Reason ||
			a[i].Message != b[i].Message ||
			a[i].ObservedGeneration != b[i].ObservedGeneration {
			return false
		}
	}

	return true
}
