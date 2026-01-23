// Copyright 2025 The prometheus-operator Authors
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

package v1alpha1

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	RemoteWriteKind    = "RemoteWrite"
	RemoteWriteName    = "remotewrites"
	RemoteWriteKindKey = "remotewrite"
)

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="rw"
// +kubebuilder:storageversion

// RemoteWrite defines a remote write endpoint for Prometheus to send metrics to.
// It allows users in different namespaces to configure their own remote write
// destinations without modifying the central Prometheus resource.
type RemoteWrite struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired remote write configuration.
	// +required
	Spec monitoringv1.RemoteWriteSpec `json:"spec"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *RemoteWrite) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// +k8s:openapi-gen=true

// RemoteWriteList is a list of RemoteWrite resources.
type RemoteWriteList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of RemoteWrite resources.
	Items []RemoteWrite `json:"items"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *RemoteWriteList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}
