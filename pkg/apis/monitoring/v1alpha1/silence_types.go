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

package v1alpha1

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SilenceKind    = "Silence"
	SilenceName    = "silences"
	SilenceKindKey = "silence"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories="prometheus-operator",shortName="sil"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type == 'Ready')].status"
// +kubebuilder:printcolumn:name="Expires At",type="date",JSONPath=".spec.expiresAt"
// +kubebuilder:printcolumn:name="Comment",type="string",JSONPath=".spec.comment"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// Silence defines a silence for Alertmanager instances managed by the prometheus-operator.
// The Silence custom resource definition (CRD) enables GitOps-friendly management of
// Alertmanager silences through Kubernetes resources, providing version control,
// RBAC integration, and audit trails for silence operations.
type Silence struct {
	// TypeMeta defines the versioned schema of this representation of an object.
	metav1.TypeMeta `json:",inline"`
	// metadata defines ObjectMeta as the metadata that all persisted resources
	// must have, which includes all objects users can create.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// spec defines the desired state of the Silence.
	// +required
	Spec SilenceSpec `json:"spec"`
	// status defines the observed state of the Silence.
	// +optional
	Status SilenceStatus `json:"status,omitempty"`
}

// SilenceList contains a list of Silence objects.
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SilenceList struct {
	// TypeMeta defines the versioned schema of this representation of an object.
	metav1.TypeMeta `json:",inline"`
	// metadata defines ListMeta as metadata for collection responses.
	metav1.ListMeta `json:"metadata,omitempty"`
	// items is the list of Silence objects.
	Items []Silence `json:"items"`
}

// SilenceSpec defines the desired state of a Silence.
type SilenceSpec struct {
	// comment provides context about the silence, explaining the reason for
	// creating it (e.g., "Scheduled maintenance window").
	// +required
	Comment string `json:"comment"`
	// expiresAt specifies when the silence expires and alert notifications
	// resume. This field is required to prevent indefinite silences.
	// +required
	ExpiresAt metav1.Time `json:"expiresAt"`
	// matchers define the alerts that should be silenced. An alert matches
	// the silence if it matches all matchers.
	// +kubebuilder:validation:MinItems=1
	// +required
	Matchers []SilenceMatcher `json:"matchers"`
}

// SilenceMatcher defines a label matcher for selecting alerts to silence.
type SilenceMatcher struct {
	// name is the label name to match against.
	// +kubebuilder:validation:MinLength=1
	// +required
	Name string `json:"name"`
	// value is the label value to match against.
	// +required
	Value string `json:"value"`
	// matchType defines the type of match to perform.
	// Supported values are:
	// - "=" (equality match)
	// - "!=" (inequality match)
	// - "=~" (regex match)
	// - "!~" (negative regex match)
	// +kubebuilder:validation:Enum="=";"!=";"=~";"!~"
	// +kubebuilder:default:="="
	// +optional
	MatchType string `json:"matchType,omitempty"`
}

// SilenceStatus defines the observed state of a Silence.
type SilenceStatus struct {
	// observedGeneration is the most recent generation observed for this Silence.
	// It corresponds to the metadata generation field, indicating whether the
	// status reflects the latest changes to the resource.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// bindings tracks the Alertmanager instances that this Silence is applied to.
	// Each binding represents one Alertmanager resource that has selected this Silence.
	// +optional
	Bindings []SilenceBinding `json:"bindings,omitempty"`
	// conditions represent the latest available observations of the Silence's state.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []monitoringv1.Condition `json:"conditions,omitempty"`
}

// SilenceBinding represents the binding between a Silence and an Alertmanager instance.
type SilenceBinding struct {
	// name is the name of the Alertmanager resource.
	// +required
	Name string `json:"name"`
	// namespace is the namespace of the Alertmanager resource.
	// +required
	Namespace string `json:"namespace"`
	// silenceID is the unique identifier of the silence in the Alertmanager API.
	// This is populated after the silence is successfully created.
	// +optional
	SilenceID string `json:"silenceID,omitempty"`
	// lastSyncTime is the last time the silence was successfully synchronized
	// with this Alertmanager instance.
	// +optional
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty"`
	// syncedInstances is the number of Alertmanager instances that have
	// successfully received this silence.
	// +optional
	SyncedInstances int32 `json:"syncedInstances,omitempty"`
	// totalInstances is the total number of Alertmanager instances that
	// should have this silence.
	// +optional
	TotalInstances int32 `json:"totalInstances,omitempty"`
	// conditions represent the latest available observations of this binding's state.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []monitoringv1.Condition `json:"conditions,omitempty"`
}
