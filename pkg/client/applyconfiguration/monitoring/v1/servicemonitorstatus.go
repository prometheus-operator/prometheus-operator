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

// ServiceMonitorStatusApplyConfiguration represents a declarative configuration of the ServiceMonitorStatus type for use
// with apply.
type ServiceMonitorStatusApplyConfiguration struct {
	References []ServiceMonitorReferenceApplyConfiguration `json:"references,omitempty"`
}

// ServiceMonitorStatusApplyConfiguration constructs a declarative configuration of the ServiceMonitorStatus type for use with
// apply.
func ServiceMonitorStatus() *ServiceMonitorStatusApplyConfiguration {
	return &ServiceMonitorStatusApplyConfiguration{}
}

// WithReferences adds the given value to the References field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the References field.
func (b *ServiceMonitorStatusApplyConfiguration) WithReferences(values ...*ServiceMonitorReferenceApplyConfiguration) *ServiceMonitorStatusApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithReferences")
		}
		b.References = append(b.References, *values[i])
	}
	return b
}
