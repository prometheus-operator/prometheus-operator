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

package admission

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// PrometheusRules is used to separate the PrometheusRules CRD wrapper from the underlying Prometheus rules.
type PrometheusRules struct {
	// TypeMeta defines the versioned schema of this representation of an object.
	metav1.TypeMeta `json:",inline"`
	// metadata defines ObjectMeta as the metadata that all persisted resources.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// spec defines the specification of desired alerting rule definitions for Prometheus.
	Spec runtime.RawExtension `json:"spec"`
}

type RuleGroups struct {
	// groups defines alerting rules groups.
	Groups []RuleGroup `json:"groups"`
}

type RuleGroup struct {
	// rules defines alerting rules.
	Rules []Rule `json:"rules"`
}

type Rule struct {
	// labels defines labels to be added to rules.
	Labels map[string]any `json:"labels,omitempty"`
	// annotations defines annotations to add to each alert.
	Annotations map[string]any `json:"annotations,omitempty"`
}
