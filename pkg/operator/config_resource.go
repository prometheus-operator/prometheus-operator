// Copyright 2025 The prometheus-operator Authors
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
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
)

const (
	statusSubResource = "status"

	// InvalidConfiguration is a generic reason for selected resources that are not valid.
	InvalidConfiguration = "InvalidConfiguration"
)

// ConfigurationResource is a type constraint that permits only the specific pointer types for configuration resources
// selectable by Prometheus, PrometheusAgent, Alertmanager or ThanosRuler.
type ConfigurationResource interface {
	*monitoringv1.ServiceMonitor | *monitoringv1.PodMonitor | *monitoringv1.Probe | *monitoringv1alpha1.ScrapeConfig | *monitoringv1.PrometheusRule | *monitoringv1alpha1.AlertmanagerConfig
}

// TypedConfigurationResource is a generic type that holds a configuration resource with its validation status.
type TypedConfigurationResource[T ConfigurationResource] struct {
	resource   T
	err        error  // Error encountered during selection or validation (nil if valid).
	reason     string // Reason for rejection; empty if accepted.
	generation int64  // Generation of the desired state (spec).
}

// TypedResourcesSelection represents a map of configuration resources selected by Prometheus or PrometheusAgent.
type TypedResourcesSelection[T ConfigurationResource] map[string]TypedConfigurationResource[T]

func NewTypedConfigurationResource[T ConfigurationResource](res T, err error, reason string, generation int64) TypedConfigurationResource[T] {
	return TypedConfigurationResource[T]{
		resource:   res,
		err:        err,
		reason:     reason,
		generation: generation,
	}
}

func (r *TypedConfigurationResource[T]) Resource() T {
	return r.resource
}

// Conditions returns a list of conditions based on the validation status of the configuration resource.
func (r *TypedConfigurationResource[T]) Conditions() []monitoringv1.ConfigResourceCondition {
	condition := monitoringv1.ConfigResourceCondition{
		Type:               monitoringv1.Accepted,
		Status:             monitoringv1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             r.reason,
		ObservedGeneration: r.generation,
	}

	if r.err != nil {
		condition.Status = monitoringv1.ConditionFalse
		condition.Message = r.err.Error()
	}

	return []monitoringv1.ConfigResourceCondition{condition}
}

// ValidResources returns only the resources which the operator considers to be valid.
// The keys of the returned map identify the resources using the `<namespace>/<name>` format.
func (resources TypedResourcesSelection[T]) ValidResources() map[string]T {
	validRes := make(map[string]T)
	for k, res := range resources {
		if res.err == nil {
			validRes[k] = res.resource
		}
	}
	return validRes
}

// ConfigResourceSyncer patches the status of configuration resources.
type ConfigResourceSyncer struct {
	client   dynamic.Interface
	accessor *Accessor

	// GroupVersionResource and metadata of the Workload.
	gvr      schema.GroupVersionResource
	workload metav1.Object
}

func NewConfigResourceSyncer(workload RuntimeObject, client dynamic.Interface, accessor *Accessor) *ConfigResourceSyncer {
	return &ConfigResourceSyncer{
		client:   client,
		accessor: accessor,
		gvr:      toGroupVersionResource(workload),
		workload: workload,
	}
}

type patch []patchOperation

type patchOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}

// GetBindingIndex returns the index of the workload binding in the slice.
// The return value is negative if there's no binding for the workload.
func (crs *ConfigResourceSyncer) GetBindingIndex(bindings []monitoringv1.WorkloadBinding) int {
	for i, binding := range bindings {
		if binding.Namespace == crs.workload.GetNamespace() &&
			binding.Name == crs.workload.GetName() &&
			binding.Group == crs.gvr.Group &&
			binding.Resource == crs.gvr.Resource {
			return i
		}
	}

	return -1
}

func (crs *ConfigResourceSyncer) newBinding(conditions []monitoringv1.ConfigResourceCondition) monitoringv1.WorkloadBinding {
	return monitoringv1.WorkloadBinding{
		Namespace:  crs.workload.GetNamespace(),
		Name:       crs.workload.GetName(),
		Resource:   crs.gvr.Resource,
		Group:      crs.gvr.Group,
		Conditions: conditions,
	}
}

func (crs *ConfigResourceSyncer) newUnstructuredBinding(conditions []monitoringv1.ConfigResourceCondition) (map[string]any, error) {
	b, err := json.Marshal(crs.newBinding(conditions))
	if err != nil {
		return nil, err
	}
	content := map[string]any{}
	if err = json.Unmarshal(b, &content); err != nil {
		return nil, err
	}
	return content, nil
}

type RuntimeObject interface {
	runtime.Object
	metav1.Object
}

type ConfigurationObject interface {
	RuntimeObject
	Bindings() []monitoringv1.WorkloadBinding
}

// UpdateBinding updates the workload's binding in the configuration resource's
// status subresource.
// If the binding is up-to-date, this is a no-operation.
func (crs *ConfigResourceSyncer) UpdateBinding(ctx context.Context, configResource ConfigurationObject, conditions []monitoringv1.ConfigResourceCondition) error {
	bindings := configResource.Bindings()

	if len(bindings) == 0 {
		// When the bindings slice is empty, update the status instead of patch
		// to avoid race conditions when the resource is selected by 2 (or
		// more) workloads.
		// Patch semantics don't allow to ensure that the second patch
		// operation doesn't overwrite the first change. Using ResourceVersion
		// (aka optimistic locking), we make sure that the second operation
		// would fail (and be retried).
		obj := &unstructured.Unstructured{
			Object: map[string]any{},
		}

		binding, err := crs.newUnstructuredBinding(conditions)
		if err != nil {
			return err
		}

		if err := unstructured.SetNestedSlice(obj.Object, []any{binding}, "status", "bindings"); err != nil {
			return err
		}

		obj.SetGroupVersionKind(configResource.GetObjectKind().GroupVersionKind())
		obj.SetKind(configResource.GetObjectKind().GroupVersionKind().Kind)
		obj.SetName(configResource.GetName())
		obj.SetResourceVersion(configResource.GetResourceVersion())

		if _, err = crs.client.Resource(toGroupVersionResource(configResource)).Namespace(configResource.GetNamespace()).UpdateStatus(
			ctx,
			obj,
			metav1.UpdateOptions{
				FieldManager:    PrometheusOperatorFieldManager,
				FieldValidation: metav1.FieldValidationStrict,
			},
		); err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}

		return nil
	}

	patch, err := crs.updateBindingPatch(bindings, conditions)
	if err != nil {
		return fmt.Errorf("failed to build patch status: %w", err)
	}

	if len(patch) == 0 {
		return nil
	}

	if _, err = crs.client.Resource(toGroupVersionResource(configResource)).Namespace(configResource.GetNamespace()).Patch(
		ctx,
		configResource.GetName(),
		types.JSONPatchType,
		patch,
		metav1.PatchOptions{
			FieldManager:    PrometheusOperatorFieldManager,
			FieldValidation: metav1.FieldValidationStrict,
		},
		statusSubResource,
	); err != nil {
		return fmt.Errorf("failed to patch status: %w", err)
	}

	return nil
}

// RemoveBinding removes the workload's binding from the configuration
// resource's status subresource.
// If the workload has no binding, this a no-operation.
func (crs *ConfigResourceSyncer) RemoveBinding(ctx context.Context, configResource ConfigurationObject) error {
	bindings := configResource.Bindings()
	p, err := crs.removeBindingPatch(bindings)
	if err != nil {
		return err
	}

	if len(p) == 0 {
		// Binding not found.
		return nil
	}

	_, err = crs.client.Resource(toGroupVersionResource(configResource)).Namespace(configResource.GetNamespace()).Patch(
		ctx,
		configResource.GetName(),
		types.JSONPatchType,
		p,
		metav1.PatchOptions{
			FieldManager:    PrometheusOperatorFieldManager,
			FieldValidation: metav1.FieldValidationStrict,
		},
		statusSubResource,
	)

	return err
}

// CleanupBindings removes the workload's binding from all configuration
// resources that are not in the resourceSelection.
func CleanupBindings[T ConfigurationResource](
	ctx context.Context,
	listerFunc func(labels.Selector, cache.AppendFunc) error,
	resourceSelection TypedResourcesSelection[T],
	csr *ConfigResourceSyncer,
) error {
	var err error
	listErr := listerFunc(labels.Everything(), func(o any) {
		if err != nil {
			// Stop processing on the first error.
			return
		}

		k, ok := csr.accessor.MetaNamespaceKey(o)
		if !ok {
			return
		}

		if _, found := resourceSelection[k]; found {
			return
		}

		obj, ok := o.(ConfigurationObject)
		if !ok {
			return
		}
		if err = k8sutil.AddTypeInformationToObject(obj); err != nil {
			err = fmt.Errorf("failed to add type information: %w", err)
			return
		}

		var gvk = obj.GetObjectKind().GroupVersionKind()

		if err = csr.RemoveBinding(ctx, obj); err != nil {
			err = fmt.Errorf("failed to remove workload binding from %s %s status: %w", gvk.Kind, k, err)
		}
	})
	if listErr != nil {
		return fmt.Errorf("listing all items from cache failed: %w", listErr)
	}
	return err
}

func toGroupVersionResource(o runtime.Object) schema.GroupVersionResource {
	gvk := o.GetObjectKind().GroupVersionKind()
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: monitoring.KindToResource(gvk.Kind),
	}
}

// updateBindingPatch returns a RFC-6902 JSON patch which updates the
// conditions of the resource's status.
// If the binding doesn't exist, the patch adds it to the status.
// If the binding is already up-to-date, the return value is empty.
func (crs *ConfigResourceSyncer) updateBindingPatch(bindings []monitoringv1.WorkloadBinding, conditions []monitoringv1.ConfigResourceCondition) ([]byte, error) {
	i := crs.GetBindingIndex(bindings)
	if i < 0 {
		// Append the workload binding to the slice.
		return json.Marshal(patch{
			patchOperation{
				Op:    "add",
				Path:  "/status/bindings/-",
				Value: crs.newBinding(conditions),
			},
		})
	}

	// No need to update the binding if the conditions haven't changed
	if equalConfigResourceConditions(bindings[i].Conditions, conditions) {
		return nil, nil
	}

	return json.Marshal(
		append(
			crs.testBindingExists(i),
			patchOperation{
				Op:    "replace",
				Path:  fmt.Sprintf("/status/bindings/%d/conditions", i),
				Value: conditions,
			},
		),
	)
}

// removeBindingPatch returns a RFC-6902 JSON patch which removes the
// workload binding from the resource's status.
// If the binding doesn't exist, the return value is empty.
func (crs *ConfigResourceSyncer) removeBindingPatch(bindings []monitoringv1.WorkloadBinding) ([]byte, error) {
	i := crs.GetBindingIndex(bindings)
	if i < 0 {
		return nil, nil
	}

	return json.Marshal(
		append(
			crs.testBindingExists(i),
			patchOperation{
				Op:   "remove",
				Path: fmt.Sprintf("/status/bindings/%d", i),
			}),
	)
}

func (crs *ConfigResourceSyncer) testBindingExists(i int) patch {
	return []patchOperation{
		{
			Op:    "test",
			Path:  fmt.Sprintf("/status/bindings/%d/name", i),
			Value: crs.workload.GetName(),
		},
		{
			Op:    "test",
			Path:  fmt.Sprintf("/status/bindings/%d/namespace", i),
			Value: crs.workload.GetNamespace(),
		},
		{
			Op:    "test",
			Path:  fmt.Sprintf("/status/bindings/%d/resource", i),
			Value: crs.gvr.Resource,
		},
		{
			Op:    "test",
			Path:  fmt.Sprintf("/status/bindings/%d/group", i),
			Value: crs.gvr.Group,
		},
	}
}

// equalConfigResourceConditions returns true when both slices are equal semantically.
func equalConfigResourceConditions(a, b []monitoringv1.ConfigResourceCondition) bool {
	if len(a) != len(b) {
		return false
	}

	ac, bc := slices.Clone(a), slices.Clone(b)

	slices.SortFunc(ac, func(a, b monitoringv1.ConfigResourceCondition) int {
		return cmp.Compare(a.Type, b.Type)
	})
	slices.SortFunc(bc, func(a, b monitoringv1.ConfigResourceCondition) int {
		return cmp.Compare(a.Type, b.Type)
	})

	return slices.EqualFunc(ac, bc, func(a, b monitoringv1.ConfigResourceCondition) bool {
		return a.Type == b.Type &&
			a.Status == b.Status &&
			a.Reason == b.Reason &&
			a.Message == b.Message &&
			a.ObservedGeneration == b.ObservedGeneration
	})
}
