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
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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

// AlertmanagerConfig defines a namespaced AlertmanagerConfig to be aggregated
// across multiple namespaces configuring one Alertmanager cluster.
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

// AlertmanagerConfigSpec is a specification of the desired behavior of the Alertmanager configuration.
// By definition, the Alertmanager configuration only applies to alerts for which
// the `namespace` label is equal to the namespace of the AlertmanagerConfig resource.
type AlertmanagerConfigSpec struct {
	// The Alertmanager route definition for alerts matching the resource’s
	// namespace. It will be added to the generated Alertmanager configuration
	// as a first-level route.
	Route *Route `json:"route,omitempty"`
	// List of receivers.
	Receivers []Receiver `json:"receivers,omitempty"`
	// List of inhibition rules. The rules will only apply to alerts matching
	// the resource’s namespace.
	InhibitRules []InhibitRule `json:"inhibitRules,omitempty"`
}

// Route defines a node in the routing tree.
type Route struct {
	// Name of the receiver for this route. If present, it should be listed in
	// the `receivers` field. The field can be omitted only for nested routes
	// otherwise it is mandatory.
	Receiver string `json:"receiver,omitempty"`
	// List of labels to group by.
	GroupBy []string `json:"groupBy,omitempty"`
	// How long to wait before sending the initial notification. Must match the
	// regular expression `[0-9]+(ms|s|m|h)` (milliseconds seconds minutes
	// hours).
	GroupWait string `json:"groupWait,omitempty"`
	// How long to wait before sending an updated notification. Must match the
	// regular expression `[0-9]+(ms|s|m|h)` (milliseconds seconds minutes
	// hours).
	GroupInterval string `json:"groupInterval,omitempty"`
	// How long to wait before repeating the last notification. Must match the
	// regular expression `[0-9]+(ms|s|m|h)` (milliseconds seconds minutes
	// hours).
	RepeatInterval string `json:"repeatInterval,omitempty"`
	// List of matchers that the alert’s labels should match. For the first
	// level route, the operator removes any existing equality and regexp
	// matcher on the `namespace` label and adds a `namespace: <object
	// namespace>` matcher.
	Matchers []Matcher `json:"matchers,omitempty"`
	// Boolean indicating whether an alert should continue matching subsequent
	// sibling nodes. It will always be overridden to true for the first-level
	// route by the Prometheus operator.
	Continue bool `json:"continue,omitempty"`
	// Child routes.
	Routes []apiextensionsv1.JSON `json:"routes,omitempty"`
	// Note: this comment applies to the field definition above but appears
	// below otherwise it gets included in the generated manifest.
	// CRD schema doesn't support self referential types for now (see
	// https://github.com/kubernetes/kubernetes/issues/62872). We have to use
	// an alternative type to circumvent the limitation. The downside is that
	// the Kube API can't validate the data beyond the fact that it is a valid
	// JSON representation.
}

// ChildRoutes extracts the child routes.
func (r *Route) ChildRoutes() ([]Route, error) {
	out := make([]Route, len(r.Routes))

	for i, v := range r.Routes {
		if err := json.Unmarshal(v.Raw, &out[i]); err != nil {
			return nil, fmt.Errorf("route[%d]: %w", i, err)
		}
	}

	return out, nil
}

// Receiver defines one or more notification integrations.
type Receiver struct {
	// Name of the receiver. Must be unique across all items from the list.
	Name string `json:"name"`
	// List of OpsGenie configurations.
	OpsGenieConfigs []OpsGenieConfig `json:"opsgenieConfigs,omitempty"`
	// List of PagerDuty configurations.
	PagerDutyConfigs []PagerDutyConfig `json:"pagerdutyConfigs,omitempty"`
	// List of Slack configurations.
	SlackConfigs []SlackConfig `json:"slackConfigs,omitempty"`
	// List of webhook configurations.
	WebhookConfigs []WebhookConfig `json:"webhookConfigs,omitempty"`
	// List of WeChat configurations.
	WeChatConfigs []WeChatConfig `json:"wechatConfigs,omitempty"`
	// List of Email configurations.
	EmailConfigs []EmailConfig `json:"emailConfigs,omitempty"`
	// List of VictorOps configurations.
	VictorOpsConfigs []VictorOpsConfig `json:"victoropsConfigs,omitempty"`
	// List of Pushover configurations.
	PushoverConfigs []PushoverConfig `json:"pushoverConfigs,omitempty"`
}

// PagerDutyConfig configures notifications via PagerDuty.
// See https://prometheus.io/docs/alerting/latest/configuration/#pagerduty_config
type PagerDutyConfig struct {
	// Whether or not to notify about resolved alerts.
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The secret's key that contains the PagerDuty integration key (when using
	// Events API v2). Either this field or `serviceKey` needs to be defined.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	RoutingKey *v1.SecretKeySelector `json:"routingKey,omitempty"`
	// The secret's key that contains the PagerDuty service key (when using
	// integration type "Prometheus"). Either this field or `routingKey` needs to
	// be defined.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	ServiceKey *v1.SecretKeySelector `json:"serviceKey,omitempty"`
	// The URL to send requests to.
	URL *string `json:"url,omitempty"`
	// Client identification.
	Client *string `json:"client,omitempty"`
	// Backlink to the sender of notification.
	ClientURL *string `json:"clientURL,omitempty"`
	// Description of the incident.
	Description *string `json:"description,omitempty"`
	// Severity of the incident.
	Severity *string `json:"severity,omitempty"`
	// The class/type of the event.
	Class *string `json:"class,omitempty"`
	// A cluster or grouping of sources.
	Group *string `json:"group,omitempty"`
	// The part or component of the affected system that is broken.
	Component *string `json:"component,omitempty"`
	// Arbitrary key/value pairs that provide further detail about the incident.
	Details []KeyValue `json:"details,omitempty"`
	// HTTP client configuration.
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
}

// SlackConfig configures notifications via Slack.
// See https://prometheus.io/docs/alerting/latest/configuration/#slack_config
type SlackConfig struct {
	// Whether or not to notify about resolved alerts.
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The secret's key that contains the Slack webhook URL.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	APIURL *v1.SecretKeySelector `json:"apiURL,omitempty"`
	// The channel or user to send notifications to.
	Channel   *string `json:"channel,omitempty"`
	Username  *string `json:"username,omitempty"`
	Color     *string `json:"color,omitempty"`
	Title     *string `json:"title,omitempty"`
	TitleLink *string `json:"titleLink,omitempty"`
	Pretext   *string `json:"pretext,omitempty"`
	Text      *string `json:"text,omitempty"`
	// A list of Slack fields that are sent with each notification.
	Fields      []SlackField `json:"fields,omitempty"`
	ShortFields *bool        `json:"shortFields,omitempty"`
	Footer      *string      `json:"footer,omitempty"`
	Fallback    *string      `json:"fallback,omitempty"`
	CallbackID  *string      `json:"callbackId,omitempty"`
	IconEmoji   *string      `json:"iconEmoji,omitempty"`
	IconURL     *string      `json:"iconURL,omitempty"`
	ImageURL    *string      `json:"imageURL,omitempty"`
	ThumbURL    *string      `json:"thumbURL,omitempty"`
	LinkNames   *bool        `json:"linkNames,omitempty"`
	MrkdwnIn    []string     `json:"mrkdwnIn,omitempty"`
	// A list of Slack actions that are sent with each notification.
	Actions []SlackAction `json:"actions,omitempty"`
	// HTTP client configuration.
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
}

// Validate ensures SlackConfig is valid.
func (sc *SlackConfig) Validate() error {
	for _, action := range sc.Actions {
		if err := action.Validate(); err != nil {
			return err
		}
	}
	for _, field := range sc.Fields {
		if err := field.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// SlackAction configures a single Slack action that is sent with each
// notification.
// See https://api.slack.com/docs/message-attachments#action_fields and
// https://api.slack.com/docs/message-buttons for more information.
type SlackAction struct {
	Type         string                  `json:"type"`
	Text         string                  `json:"text"`
	URL          string                  `json:"url,omitempty"`
	Style        string                  `json:"style,omitempty"`
	Name         string                  `json:"name,omitempty"`
	Value        string                  `json:"value,omitempty"`
	ConfirmField *SlackConfirmationField `json:"confirm,omitempty"`
}

// Validate ensures SlackAction is valid.
func (sa *SlackAction) Validate() error {
	if sa.Type == "" {
		return errors.New("missing type in Slack action configuration")
	}
	if sa.Text == "" {
		return errors.New("missing text in Slack action configuration")
	}
	if sa.URL == "" && sa.Name == "" {
		return errors.New("missing name or url in Slack action configuration")
	}
	if sa.ConfirmField != nil {
		if err := sa.ConfirmField.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// SlackConfirmationField protect users from destructive actions or
// particularly distinguished decisions by asking them to confirm their button
// click one more time.
// See https://api.slack.com/docs/interactive-message-field-guide#confirmation_fields
// for more information.
type SlackConfirmationField struct {
	Text        string  `json:"text"`
	Title       *string `json:"title,omitempty"`
	OkText      *string `json:"okText,omitempty"`
	DismissText *string `json:"dismissText,omitempty"`
}

// Validate ensures SlackConfirmationField is valid
func (scf *SlackConfirmationField) Validate() error {
	if scf.Text == "" {
		return errors.New("missing text in Slack confirmation configuration")
	}
	return nil
}

// SlackField configures a single Slack field that is sent with each notification.
// Each field must contain a title, value, and optionally, a boolean value to indicate if the field
// is short enough to be displayed next to other fields designated as short.
// See https://api.slack.com/docs/message-attachments#fields for more information.
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short *bool  `json:"short,omitempty"`
}

// Validate ensures SlackField is valid
func (sf *SlackField) Validate() error {
	if sf.Title == "" {
		return errors.New("missing title in Slack field configuration")
	}
	if sf.Value == "" {
		return errors.New("missing value in Slack field configuration")
	}
	return nil
}

// WebhookConfig configures notifications via a generic receiver supporting the webhook payload.
// See https://prometheus.io/docs/alerting/latest/configuration/#webhook_config
type WebhookConfig struct {
	// Whether or not to notify about resolved alerts.
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The URL to send HTTP POST requests to. `urlSecret` takes precedence over
	// `url`. One of `urlSecret` and `url` should be defined.
	URL *string `json:"url,omitempty"`
	// The secret's key that contains the webhook URL to send HTTP requests to.
	// `urlSecret` takes precedence over `url`. One of `urlSecret` and `url`
	// should be defined.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	URLSecret *v1.SecretKeySelector `json:"urlSecret,omitempty"`
	// HTTP client configuration.
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
	// Maximum number of alerts to be sent per webhook message.
	MaxAlerts *int32 `json:"maxAlerts,omitempty"`
}

// OpsGenieConfig configures notifications via OpsGenie.
// See https://prometheus.io/docs/alerting/latest/configuration/#opsgenie_config
type OpsGenieConfig struct {
	// Whether or not to notify about resolved alerts.
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The secret's key that contains the OpsGenie API key.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	APIKey *v1.SecretKeySelector `json:"apiKey,omitempty"`
	// The URL to send OpsGenie API requests to.
	APIURL *string `json:"apiURL,omitempty"`
	// Alert text limited to 130 characters.
	Message *string `json:"message,omitempty"`
	// Description of the incident.
	Description *string `json:"description,omitempty"`
	// Backlink to the sender of the notification.
	Source *string `json:"source,omitempty"`
	// Comma separated list of tags attached to the notifications.
	Tags *string `json:"tags,omitempty"`
	// Additional alert note.
	Note *string `json:"note,omitempty"`
	// Priority level of alert. Possible values are P1, P2, P3, P4, and P5.
	Priority *string `json:"priority,omitempty"`
	// A set of arbitrary key/value pairs that provide further detail about the incident.
	Details []KeyValue `json:"details,omitempty"`
	// List of responders responsible for notifications.
	Responders []OpsGenieConfigResponder `json:"responders,omitempty"`
	// HTTP client configuration.
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
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

// OpsGenieConfigResponder defines a responder to an incident.
// One of id, name or username has to be defined.
type OpsGenieConfigResponder struct {
	// ID of the responder.
	ID string `json:"id,omitempty"`
	// Name of the responder.
	Name string `json:"name,omitempty"`
	// Username of the responder.
	Username string `json:"username,omitempty"`
	// Type of responder.
	Type string `json:"type,omitempty"`
}

// Validate ensures OpsGenieConfigResponder is valid
func (r *OpsGenieConfigResponder) Validate() error {
	if r.ID == "" && r.Name == "" && r.Username == "" {
		return errors.New("responder must have at least an ID, a Name or an Username defined")
	}

	if !opsGenieTypeRe.MatchString(r.Type) {
		return errors.New("responder type should match team, user, escalation or schedule")
	}

	return nil
}

// HTTPConfig defines a client HTTP configuration.
// See https://prometheus.io/docs/alerting/latest/configuration/#http_config
type HTTPConfig struct {
	// BasicAuth for the client.
	BasicAuth *monitoringv1.BasicAuth `json:"basicAuth,omitempty"`
	// The secret's key that contains the bearer token to be used by the client
	// for authentication.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	BearerTokenSecret *v1.SecretKeySelector `json:"bearerTokenSecret,omitempty"`
	// TLS configuration for the client.
	TLSConfig *monitoringv1.SafeTLSConfig `json:"tlsConfig,omitempty"`
	// Optional proxy URL.
	ProxyURL *string `json:"proxyURL,omitempty"`
}

// WeChatConfig configures notifications via WeChat.
// See https://prometheus.io/docs/alerting/latest/configuration/#wechat_config
type WeChatConfig struct {
	// Whether or not to notify about resolved alerts.
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The secret's key that contains the WeChat API key.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	APISecret *v1.SecretKeySelector `json:"apiSecret,omitempty"`
	// The WeChat API URL.
	APIURL *string `json:"apiURL,omitempty"`
	// The corp id for authentication.
	CorpID  *string `json:"corpID,omitempty"`
	AgentID *string `json:"agentID,omitempty"`
	ToUser  *string `json:"toUser,omitempty"`
	ToParty *string `json:"toParty,omitempty"`
	ToTag   *string `json:"toTag,omitempty"`
	// API request data as defined by the WeChat API.
	Message     *string `json:"message,omitempty"`
	MessageType *string `json:"messageType,omitempty"`
	// HTTP client configuration.
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
}

// EmailConfig configures notifications via Email.
type EmailConfig struct {
	// Whether or not to notify about resolved alerts.
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The email address to send notifications to.
	To *string `json:"to,omitempty"`
	// The sender address.
	From *string `json:"from,omitempty"`
	// The hostname to identify to the SMTP server.
	Hello *string `json:"hello,omitempty"`
	// The SMTP host through which emails are sent.
	Smarthost *string `json:"smarthost,omitempty"`
	// SMTP authentication information.
	AuthUsername *string               `json:"authUsername,omitempty"`
	AuthPassword *v1.SecretKeySelector `json:"authPassword,omitempty"`
	AuthSecret   *v1.SecretKeySelector `json:"authSecret,omitempty"`
	AuthIdentity *string               `json:"authIdentity,omitempty"`
	// Further headers email header key/value pairs. Overrides any headers
	// previously set by the notification implementation.
	Headers []KeyValue `json:"headers,omitempty"`
	// The HTML body of the email notification.
	HTML *string `json:"html,omitempty"`
	// The text body of the email notification.
	Text *string `json:"text,omitempty"`
	// The SMTP TLS requirement.
	// Note that Go does not support unencrypted connections to remote SMTP endpoints.
	RequireTLS *bool `json:"requireTLS,omitempty"`
	// TLS configuration
	TLSConfig *monitoringv1.SafeTLSConfig `json:"tlsConfig,omitempty"`
}

// VictorOpsConfig configures notifications via VictorOps.
// See https://prometheus.io/docs/alerting/latest/configuration/#victorops_config
type VictorOpsConfig struct {
	// Whether or not to notify about resolved alerts.
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The API key to use when talking to the VictorOps API.
	APIKey *v1.SecretKeySelector `json:"apiKey,omitempty"`
	// The VictorOps API URL.
	APIURL *string `json:"apiUrl,omitempty"`
	// A key used to map the alert to a team.
	RoutingKey *string `json:"routingKey"`
	// Describes the behavior of the alert (CRITICAL, WARNING, INFO).
	MessageType *string `json:"messageType,omitempty"`
	// Contains summary of the alerted problem.
	EntityDisplayName *string `json:"entityDisplayName,omitempty"`
	// Contains long explanation of the alerted problem.
	StateMessage *string `json:"stateMessage,omitempty"`
	// The monitoring tool the state message is from.
	MonitoringTool *string `json:"monitoringTool,omitempty"`
	// Additional custom fields for notification.
	CustomFields []KeyValue `json:"customFields,omitempty"`
	// The HTTP client's configuration.
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
}

// PushoverConfig configures notifications via Pushover.
// See https://prometheus.io/docs/alerting/latest/configuration/#pushover_config
type PushoverConfig struct {
	// Whether or not to notify about resolved alerts.
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The recipient user’s user key.
	UserKey *v1.SecretKeySelector `json:"userKey,omitempty"`
	// Your registered application’s API token, see https://pushover.net/apps
	Token *v1.SecretKeySelector `json:"token,omitempty"`
	// Notification title.
	Title *string `json:"title,omitempty"`
	// Notification message.
	Message *string `json:"message,omitempty"`
	// A supplementary URL shown alongside the message.
	URL *string `json:"url,omitempty"`
	// A title for supplementary URL, otherwise just the URL is shown
	URLTitle *string `json:"urlTitle,omitempty"`
	// The name of one of the sounds supported by device clients to override the user's default sound choice
	Sound *string `json:"sound,omitempty"`
	// Priority, see https://pushover.net/api#priority
	Priority *string `json:"priority,omitempty"`
	// How often the Pushover servers will send the same notification to the user.
	// Must be at least 30 seconds.
	Retry *string `json:"retry,omitempty"`
	// How long your notification will continue to be retried for, unless the user
	// acknowledges the notification.
	Expire *string `json:"expire,omitempty"`
	// Whether notification message is HTML or plain text.
	HTML *bool `json:"html,omitempty"`
	// HTTP client configuration.
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
}

// InhibitRule defines an inhibition rule that allows to mute alerts when other
// alerts are already firing.
// See https://prometheus.io/docs/alerting/latest/configuration/#inhibit_rule
type InhibitRule struct {
	// Matchers that have to be fulfilled in the alerts to be muted. The
	// operator enforces that the alert matches the resource’s namespace.
	TargetMatch []Matcher `json:"targetMatch,omitempty"`
	// Matchers for which one or more alerts have to exist for the inhibition
	// to take effect. The operator enforces that the alert matches the
	// resource’s namespace.
	SourceMatch []Matcher `json:"sourceMatch,omitempty"`
	// Labels that must have an equal value in the source and target alert for
	// the inhibition to take effect.
	Equal []string `json:"equal,omitempty"`
}

// KeyValue defines a (key, value) tuple.
type KeyValue struct {
	// Key of the tuple.
	Key string `json:"key"`
	// Value of the tuple.
	Value string `json:"value"`
}

// Matcher defines how to match on alert's labels.
type Matcher struct {
	// Label to match.
	Name string `json:"name"`
	// Label value to match.
	Value string `json:"value"`
	// Whether to match on equality (false) or regular-expression (true).
	Regex bool `json:"regex,omitempty"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *AlertmanagerConfig) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// DeepCopyObject implements the runtime.Object interface.
func (l *AlertmanagerConfigList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}
