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

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/applyconfigurations/meta/v1"
)

// EmbeddedPersistentVolumeClaimApplyConfiguration represents an declarative configuration of the EmbeddedPersistentVolumeClaim type for use
// with apply.
type EmbeddedPersistentVolumeClaimApplyConfiguration struct {
	v1.TypeMetaApplyConfiguration             `json:",inline"`
	*EmbeddedObjectMetadataApplyConfiguration `json:"metadata,omitempty"`
	Spec                                      *corev1.PersistentVolumeClaimSpec   `json:"spec,omitempty"`
	Status                                    *corev1.PersistentVolumeClaimStatus `json:"status,omitempty"`
	PropagateOwnerReferences                  *bool                               `json:"propagateOwnerReferences,omitempty"`
}

// EmbeddedPersistentVolumeClaimApplyConfiguration constructs an declarative configuration of the EmbeddedPersistentVolumeClaim type for use with
// apply.
func EmbeddedPersistentVolumeClaim() *EmbeddedPersistentVolumeClaimApplyConfiguration {
	b := &EmbeddedPersistentVolumeClaimApplyConfiguration{}
	b.WithKind("EmbeddedPersistentVolumeClaim")
	b.WithAPIVersion("monitoring.coreos.com/v1")
	return b
}

// WithKind sets the Kind field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Kind field is set to the value of the last call.
func (b *EmbeddedPersistentVolumeClaimApplyConfiguration) WithKind(value string) *EmbeddedPersistentVolumeClaimApplyConfiguration {
	b.Kind = &value
	return b
}

// WithAPIVersion sets the APIVersion field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the APIVersion field is set to the value of the last call.
func (b *EmbeddedPersistentVolumeClaimApplyConfiguration) WithAPIVersion(value string) *EmbeddedPersistentVolumeClaimApplyConfiguration {
	b.APIVersion = &value
	return b
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *EmbeddedPersistentVolumeClaimApplyConfiguration) WithName(value string) *EmbeddedPersistentVolumeClaimApplyConfiguration {
	b.ensureEmbeddedObjectMetadataApplyConfigurationExists()
	b.Name = &value
	return b
}

// WithLabels puts the entries into the Labels field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Labels field,
// overwriting an existing map entries in Labels field with the same key.
func (b *EmbeddedPersistentVolumeClaimApplyConfiguration) WithLabels(entries map[string]string) *EmbeddedPersistentVolumeClaimApplyConfiguration {
	b.ensureEmbeddedObjectMetadataApplyConfigurationExists()
	if b.Labels == nil && len(entries) > 0 {
		b.Labels = make(map[string]string, len(entries))
	}
	for k, v := range entries {
		b.Labels[k] = v
	}
	return b
}

// WithAnnotations puts the entries into the Annotations field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Annotations field,
// overwriting an existing map entries in Annotations field with the same key.
func (b *EmbeddedPersistentVolumeClaimApplyConfiguration) WithAnnotations(entries map[string]string) *EmbeddedPersistentVolumeClaimApplyConfiguration {
	b.ensureEmbeddedObjectMetadataApplyConfigurationExists()
	if b.Annotations == nil && len(entries) > 0 {
		b.Annotations = make(map[string]string, len(entries))
	}
	for k, v := range entries {
		b.Annotations[k] = v
	}
	return b
}

func (b *EmbeddedPersistentVolumeClaimApplyConfiguration) ensureEmbeddedObjectMetadataApplyConfigurationExists() {
	if b.EmbeddedObjectMetadataApplyConfiguration == nil {
		b.EmbeddedObjectMetadataApplyConfiguration = &EmbeddedObjectMetadataApplyConfiguration{}
	}
}

// WithSpec sets the Spec field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Spec field is set to the value of the last call.
func (b *EmbeddedPersistentVolumeClaimApplyConfiguration) WithSpec(value corev1.PersistentVolumeClaimSpec) *EmbeddedPersistentVolumeClaimApplyConfiguration {
	b.Spec = &value
	return b
}

// WithStatus sets the Status field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Status field is set to the value of the last call.
func (b *EmbeddedPersistentVolumeClaimApplyConfiguration) WithStatus(value corev1.PersistentVolumeClaimStatus) *EmbeddedPersistentVolumeClaimApplyConfiguration {
	b.Status = &value
	return b
}

// WithPropagateOwnerReferences sets the PropagateOwnerReferences field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the PropagateOwnerReferences field is set to the value of the last call.
func (b *EmbeddedPersistentVolumeClaimApplyConfiguration) WithPropagateOwnerReferences(value bool) *EmbeddedPersistentVolumeClaimApplyConfiguration {
	b.PropagateOwnerReferences = &value
	return b
}
