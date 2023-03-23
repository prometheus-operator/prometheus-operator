// Copyright 2023 The prometheus-operator Authors
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
	PrometheusAgentsKind   = "PrometheusAgent"
	PrometheusAgentName    = "prometheusagents"
	PrometheusAgentKindKey = "prometheusagent"
)

func (l *PrometheusAgent) GetCommonPrometheusFields() monitoringv1.CommonPrometheusFields {
	return l.Spec.CommonPrometheusFields
}

func (l *PrometheusAgent) SetCommonPrometheusFields(f monitoringv1.CommonPrometheusFields) {
	l.Spec.CommonPrometheusFields = f
}

func (l *PrometheusAgent) GetTypeMeta() metav1.TypeMeta {
	return l.TypeMeta
}

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="promagent"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version",description="The version of Prometheus agent"
// +kubebuilder:printcolumn:name="Desired",type="integer",JSONPath=".spec.replicas",description="The number of desired replicas"
// +kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.availableReplicas",description="The number of ready replicas"
// +kubebuilder:printcolumn:name="Reconciled",type="string",JSONPath=".status.conditions[?(@.type == 'Reconciled')].status"
// +kubebuilder:printcolumn:name="Available",type="string",JSONPath=".status.conditions[?(@.type == 'Available')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Paused",type="boolean",JSONPath=".status.paused",description="Whether the resource reconciliation is paused or not",priority=1
// +kubebuilder:subresource:status

// PrometheusAgent defines a Prometheus agent deployment.
type PrometheusAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the Prometheus agent. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Spec PrometheusAgentSpec `json:"spec"`
	// Most recent observed status of the Prometheus cluster. Read-only.
	// More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Status monitoringv1.PrometheusStatus `json:"status,omitempty"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *PrometheusAgent) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// PrometheusAgentList is a list of Prometheus agents.
// +k8s:openapi-gen=true
type PrometheusAgentList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of Prometheus agents
	Items []*PrometheusAgent `json:"items"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *PrometheusAgentList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// PrometheusAgentSpec is a specification of the desired behavior of the Prometheus agent. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type PrometheusAgentSpec struct {
	monitoringv1.CommonPrometheusFields `json:",inline"`
}
