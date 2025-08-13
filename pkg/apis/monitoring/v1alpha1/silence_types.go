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
// +kubebuilder:printcolumn:name="Expires At",type="string",JSONPath=".spec.expiresAt"
// +kubebuilder:printcolumn:name="Alertmanager",type="string",JSONPath=".spec.alertmanagerRef.name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// Silence enables GitOps-friendly management of Alertmanager silences through
// Kubernetes resources by extending the existing Alertmanager controller.
type Silence struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the Silence.
	Spec SilenceSpec `json:"spec"`
	// Most recently observed status of the Silence.
	// +optional
	Status SilenceStatus `json:"status,omitempty"`
}

// SilenceSpec defines the desired state of Silence
type SilenceSpec struct {
	// Comment provides context about the silence, typically explaining
	// the reason for creating it (e.g., "Scheduled maintenance window").
	// +optional
	Comment string `json:"comment,omitempty"`

	// ExpiresAt specifies when the silence expires and alert notifications
	// resume. This field is required to prevent indefinite silences.
	ExpiresAt metav1.Time `json:"expiresAt"`

	// Matchers define the alert matching rules for this silence.
	// An alert is silenced only if all matchers match the alert's labels.
	// +kubebuilder:validation:MinItems=1
	Matchers []SilenceMatcher `json:"matchers"`

	// AlertmanagerRef references a specific Alertmanager instance to apply this silence to.
	// If not specified, the silence applies to all Alertmanager instances in the namespace.
	// +optional
	AlertmanagerRef *AlertmanagerRef `json:"alertmanagerRef,omitempty"`
}

// SilenceMatcher defines a single matcher for alert labels in a silence.
// Matchers follow Alertmanager's label matching semantics.
type SilenceMatcher struct {
	// Name specifies the alert label name to match against.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Value specifies the alert label value to match against.
	Value string `json:"value"`

	// MatchType defines how the matcher compares the label value.
	// "=" for exact equality, "!=" for inequality,
	// "=~" for regex match, "!~" for regex non-match.
	// +kubebuilder:validation:Enum="=";"!=";"=~";"!~"
	// +kubebuilder:default="="
	MatchType string `json:"matchType"`
}

// AlertmanagerRef references a specific Alertmanager instance for targeted silence application
type AlertmanagerRef struct {
	// Name of the target Alertmanager instance
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Namespace of the target Alertmanager instance.
	// Defaults to the namespace of the Silence object if not specified.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// SilenceStatus tracks the synchronization state across multiple Alertmanager instances
type SilenceStatus struct {
	// Conditions represent the latest available observations of the silence's state
	// using standard Kubernetes condition semantics.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []SilenceCondition `json:"conditions,omitempty"`

	// SilenceID contains the Alertmanager-assigned unique identifier for this silence.
	// Set once the silence is successfully created via Alertmanager REST API v2.
	// +optional
	SilenceID string `json:"silenceID,omitempty"`

	// ObservedGeneration represents the .metadata.generation that was last processed
	// by the controller to track if updates are needed.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// SilenceCondition represents the state of a silence at a certain point using
// standard Kubernetes condition semantics for status reporting.
type SilenceCondition struct {
	// Type of the condition (Ready, Synced, Failed).
	Type SilenceConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status metav1.ConditionStatus `json:"status"`
	// LastTransitionTime is the last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Reason is a unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Message is a human-readable message indicating details about the last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// SilenceConditionType represents the type of condition for a Silence.
// +kubebuilder:validation:Enum=Ready;Synced;Failed
type SilenceConditionType string

const (
	// SilenceReady indicates the silence has been successfully created/updated via Alertmanager REST API.
	SilenceReady SilenceConditionType = "Ready"
	// SilenceSynced indicates the silence is synchronized across all target Alertmanager instances.
	SilenceSynced SilenceConditionType = "Synced"
	// SilenceFailed indicates an error occurred during Alertmanager API interaction.
	SilenceFailed SilenceConditionType = "Failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SilenceList is a list of Silence resources
type SilenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Silence `json:"items"`
}
