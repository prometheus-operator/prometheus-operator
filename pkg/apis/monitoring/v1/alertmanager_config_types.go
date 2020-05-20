// Copyright 2018 The prometheus-operator Authors
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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	AlertmanagerConfigKind    = "AlertmanagerConfig"
	AlertmanagerConfigName    = "alertmanagerconfigs"
	AlertmanagerConfigKindKey = "alertmanagerconfig"
)

// AlertmanagerConfig defines a namespaced AlertmanagerConfig to be aggregated across multiple namespaces configuring one Alertmanager.
// +genclient
// +k8s:openapi-gen=true
type AlertmanagerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AlertmanagerConfigSpec `json:"spec"`
}

// AlertmanagerConfigList is a list of AlertmanagerConfig.
// +k8s:openapi-gen=true
type AlertmanagerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of AlertmanagerConfig
	Items []*AlertmanagerConfig `json:"items"`
}

type AlertmanagerConfigSpec struct {
	Route        *Route        `json:"route,omitempty"`
	Receivers    []Receiver    `json:"receivers,omitempty"`
	InhibitRules []InhibitRule `json:"inhibitRules,omitempty"`
}

type Route struct {
	Receiver       string    `json:"receiver,omitempty"`
	GroupBy        []string  `json:"groupBy,omitempty"`
	GroupWait      string    `json:"groupWait,omitempty"`
	GroupInterval  string    `json:"groupInterval,omitempty"`
	RepeatInterval string    `json:"repeatInterval,omitempty"`
	Matchers       []Matcher `json:"matchers,omitempty"`
	Continue       *bool     `json:"continue,omitempty"`
	Routes         []Route   `json:"routes,omitempty"`
}

type Receiver struct {
	Name             string            `json:"name"`
	PagerDutyConfigs []PagerDutyConfig `json:"pagerDutyConfigs,omitempty"`
}

type PagerDutyConfig struct {
	SendResolved *bool                   `json:"sendResolved,omitempty"`
	RoutingKey   *string                 `json:"routingKey,omitempty"`
	ServiceKey   *string                 `json:"serviceKey,omitempty"`
	URL          *string                 `json:"url,omitempty"`
	Client       *string                 `json:"client,omitempty"`
	ClientURL    *string                 `json:"clientURL,omitempty"`
	Description  *string                 `json:"description,omitempty"`
	Severity     *string                 `json:"severity,omitempty"`
	Details      []PagerDutyConfigDetail `json:"details,omitempty"`
	HTTPConfig   *HTTPConfig             `json:"httpConfig,omitempty"`
}

type HTTPConfig struct {
	BasicAuth         *BasicAuth            `json:"basicAuth,omitempty"`
	BearerTokenSecret *v1.SecretKeySelector `json:"bearerTokenSecret,omitempty"`
	TLSConfig         *TLSConfig            `json:"tlsConfig,omitempty"`
	ProxyURL          *string               `json:"proxyURL,omitempty"`
}

type PagerDutyConfigDetail struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type InhibitRule struct {
	TargetMatch []Matcher `json:"matcher,omitempty"`
	SourceMatch []Matcher `json:"matcher,omitempty"`
	Equal       []string  `json:"equal,omitempty"`
}

type Matcher struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Regex *bool  `json:"regex,omitempty"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *AlertmanagerConfig) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// DeepCopyObject implements the runtime.Object interface.
func (l *AlertmanagerConfigList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}
