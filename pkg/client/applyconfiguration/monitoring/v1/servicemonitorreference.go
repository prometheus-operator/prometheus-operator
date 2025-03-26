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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

// ServiceMonitorReferenceApplyConfiguration represents a declarative configuration of the ServiceMonitorReference type for use
// with apply.
type ServiceMonitorReferenceApplyConfiguration struct {
	Resource   *string                       `json:"resource,omitempty"`
	Name       *string                       `json:"name,omitempty"`
	Namespace  *string                       `json:"namespace,omitempty"`
	Conditions []ConditionApplyConfiguration `json:"conditions,omitempty"`
}

// ServiceMonitorReferenceApplyConfiguration constructs a declarative configuration of the ServiceMonitorReference type for use with
// apply.
func ServiceMonitorReference() *ServiceMonitorReferenceApplyConfiguration {
	return &ServiceMonitorReferenceApplyConfiguration{}
}

// WithResource sets the Resource field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Resource field is set to the value of the last call.
func (b *ServiceMonitorReferenceApplyConfiguration) WithResource(value string) *ServiceMonitorReferenceApplyConfiguration {
	b.Resource = &value
	return b
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *ServiceMonitorReferenceApplyConfiguration) WithName(value string) *ServiceMonitorReferenceApplyConfiguration {
	b.Name = &value
	return b
}

// WithNamespace sets the Namespace field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Namespace field is set to the value of the last call.
func (b *ServiceMonitorReferenceApplyConfiguration) WithNamespace(value string) *ServiceMonitorReferenceApplyConfiguration {
	b.Namespace = &value
	return b
}

// WithConditions adds the given value to the Conditions field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Conditions field.
func (b *ServiceMonitorReferenceApplyConfiguration) WithConditions(values ...*ConditionApplyConfiguration) *ServiceMonitorReferenceApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithConditions")
		}
		b.Conditions = append(b.Conditions, *values[i])
	}
	return b
}
