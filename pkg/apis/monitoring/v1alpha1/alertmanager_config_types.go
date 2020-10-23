// Copyright 2020 The prometheus-operator Authors
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
	"regexp"

	"github.com/pkg/errors"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	Version = "v1alpha1"

	AlertmanagerConfigKind    = "AlertmanagerConfig"
	AlertmanagerConfigName    = "alertmanagerconfigs"
	AlertmanagerConfigKindKey = "alertmanagerconfig"
)

var (
	opsGenieTypeRe = regexp.MustCompile("^(team|user|escalation|schedule)$")
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
	Continue       bool      `json:"continue,omitempty"`
	Routes         []Route   `json:"routes,omitempty"`
}

type Receiver struct {
	Name             string            `json:"name"`
	OpsGenieConfigs  []OpsGenieConfig  `json:"opsgenieConfigs,omitempty"`
	PagerDutyConfigs []PagerDutyConfig `json:"pagerDutyConfigs,omitempty"`
	WebhookConfigs   []WebhookConfig   `json:"webhookConfigs,omitempty"`
}

type PagerDutyConfig struct {
	SendResolved *bool                   `json:"sendResolved,omitempty"`
	RoutingKey   *v1.SecretKeySelector   `json:"routingKey,omitempty"`
	ServiceKey   *v1.SecretKeySelector   `json:"serviceKey,omitempty"`
	URL          *string                 `json:"url,omitempty"`
	Client       *string                 `json:"client,omitempty"`
	ClientURL    *string                 `json:"clientURL,omitempty"`
	Description  *string                 `json:"description,omitempty"`
	Severity     *string                 `json:"severity,omitempty"`
	Class        *string                 `json:"class,omitempty"`
	Group        *string                 `json:"group,omitempty"`
	Component    *string                 `json:"component,omitempty"`
	Details      []PagerDutyConfigDetail `json:"details,omitempty"`
	HTTPConfig   *HTTPConfig             `json:"httpConfig,omitempty"`
}

type WebhookConfig struct {
	SendResolved *bool                 `json:"sendResolved,omitempty"`
	URL          *string               `json:"url,omitempty"`
	URLSecret    *v1.SecretKeySelector `json:"urlSecret,omitempty"`
	HTTPConfig   *HTTPConfig           `json:"httpConfig,omitempty"`
	MaxAlerts    *int32                `json:"maxAlerts,omitempty"`
}

type OpsGenieConfig struct {
	SendResolved *bool                     `json:"sendResolved,omitempty"`
	APIKey       *v1.SecretKeySelector     `json:"apiKey,omitempty"`
	APIURL       *string                   `json:"apiURL,omitempty"`
	Message      *string                   `json:"message,omitempty"`
	Description  *string                   `json:"description,omitempty"`
	Source       *string                   `json:"source,omitempty"`
	Tags         *string                   `json:"tags,omitempty"`
	Note         *string                   `json:"note,omitempty"`
	Priority     *string                   `json:"priority,omitempty"`
	Details      []OpsGenieConfigDetail    `json:"details,omitempty"`
	Responders   []OpsGenieConfigResponder `json:"responders,omitempty"`
	HTTPConfig   *HTTPConfig               `json:"httpConfig,omitempty"`
}

// Validate ensures OpsGenieConfig is valid
func (o *OpsGenieConfig) Validate() error {
	for _, responder := range o.Responders {
		if err := responder.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type OpsGenieConfigResponder struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Username string `json:"username,omitempty"`
	Type     string `json:"type,omitempty"`
}

// Validate ensures OpsGenieConfigResponder is valid
func (r *OpsGenieConfigResponder) Validate() error {
	if r.ID == "" && r.Name == "" && r.Username == "" {
		return errors.Errorf("responder must have at least an ID, a Name or an Username defined")
	}

	if !opsGenieTypeRe.MatchString(r.Type) {
		return errors.Errorf("responder type should match team, user, escalation or schedule")
	}

	return nil
}

type HTTPConfig struct {
	BasicAuth         *monitoringv1.BasicAuth     `json:"basicAuth,omitempty"`
	BearerTokenSecret *v1.SecretKeySelector       `json:"bearerTokenSecret,omitempty"`
	TLSConfig         *monitoringv1.SafeTLSConfig `json:"tlsConfig,omitempty"`
	ProxyURL          *string                     `json:"proxyURL,omitempty"`
}

type OpsGenieConfigDetail struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PagerDutyConfigDetail struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type InhibitRule struct {
	TargetMatch []Matcher `json:"targetMatch,omitempty"`
	SourceMatch []Matcher `json:"sourceMatch,omitempty"`
	Equal       []string  `json:"equal,omitempty"`
}

type Matcher struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Regex bool   `json:"regex,omitempty"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *AlertmanagerConfig) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// DeepCopyObject implements the runtime.Object interface.
func (l *AlertmanagerConfigList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}
