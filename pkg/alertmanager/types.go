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

package alertmanager

import (
	"github.com/prometheus/alertmanager/config"
	commoncfg "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

// Customization of Config type from alertmanager repo:
// https://github.com/prometheus/alertmanager/blob/main/config/config.go
//
// Custom global type to get around obfuscation of secret values when
// marshalling. See the following issue for details:
// https://github.com/prometheus/alertmanager/issues/1985
type alertmanagerConfig struct {
	Global            *globalConfig   `yaml:"global,omitempty"`
	Route             *route          `yaml:"route,omitempty"`
	InhibitRules      []*inhibitRule  `yaml:"inhibit_rules,omitempty"`
	Receivers         []*receiver     `yaml:"receivers,omitempty"`
	MuteTimeIntervals []*timeInterval `yaml:"mute_time_intervals,omitempty"`
	TimeIntervals     []*timeInterval `yaml:"time_intervals,omitempty"`
	Templates         []string        `yaml:"templates"`
}

type globalConfig struct {
	// ResolveTimeout is the time after which an alert is declared resolved
	// if it has not been updated.
	ResolveTimeout *model.Duration `yaml:"resolve_timeout,omitempty"`

	HTTPConfig *httpClientConfig `yaml:"http_config,omitempty"`

	SMTPFrom              string          `yaml:"smtp_from,omitempty"`
	SMTPHello             string          `yaml:"smtp_hello,omitempty"`
	SMTPSmarthost         config.HostPort `yaml:"smtp_smarthost,omitempty"`
	SMTPAuthUsername      string          `yaml:"smtp_auth_username,omitempty"`
	SMTPAuthPassword      string          `yaml:"smtp_auth_password,omitempty"`
	SMTPAuthPasswordFile  string          `yaml:"smtp_auth_password_file,omitempty"`
	SMTPAuthSecret        string          `yaml:"smtp_auth_secret,omitempty"`
	SMTPAuthSecretFile    string          `yaml:"smtp_auth_secret_file,omitempty"`
	SMTPAuthIdentity      string          `yaml:"smtp_auth_identity,omitempty"`
	SMTPRequireTLS        *bool           `yaml:"smtp_require_tls,omitempty"`
	SMTPTLSConfig         *tlsConfig      `yaml:"smtp_tls_config,omitempty"`
	SMTPForceImplicitTLS  *bool           `yaml:"smtp_force_implicit_tls,omitempty"`
	SlackAPIURL           *config.URL     `yaml:"slack_api_url,omitempty"`
	SlackAPIURLFile       string          `yaml:"slack_api_url_file,omitempty"`
	PagerdutyURL          *config.URL     `yaml:"pagerduty_url,omitempty"`
	HipchatAPIURL         *config.URL     `yaml:"hipchat_api_url,omitempty"`
	HipchatAuthToken      string          `yaml:"hipchat_auth_token,omitempty"`
	OpsGenieAPIURL        *config.URL     `yaml:"opsgenie_api_url,omitempty"`
	OpsGenieAPIKey        string          `yaml:"opsgenie_api_key,omitempty"`
	OpsGenieAPIKeyFile    string          `yaml:"opsgenie_api_key_file,omitempty"`
	WeChatAPIURL          *config.URL     `yaml:"wechat_api_url,omitempty"`
	WeChatAPISecret       string          `yaml:"wechat_api_secret,omitempty"`
	WeChatAPICorpID       string          `yaml:"wechat_api_corp_id,omitempty"`
	VictorOpsAPIURL       *config.URL     `yaml:"victorops_api_url,omitempty"`
	VictorOpsAPIKey       string          `yaml:"victorops_api_key,omitempty"`
	VictorOpsAPIKeyFile   string          `yaml:"victorops_api_key_file,omitempty"`
	TelegramAPIURL        *config.URL     `yaml:"telegram_api_url,omitempty"`
	TelegramBotToken      string          `yaml:"telegram_bot_token,omitempty"`
	TelegramBotTokenFile  string          `yaml:"telegram_bot_token_file,omitempty"`
	WebexAPIURL           *config.URL     `yaml:"webex_api_url,omitempty"`
	JiraAPIURL            *config.URL     `yaml:"jira_api_url,omitempty"`
	RocketChatAPIURL      *config.URL     `yaml:"rocketchat_api_url,omitempty"`
	RocketChatToken       string          `yaml:"rocketchat_token,omitempty"`
	RocketChatTokenFile   string          `yaml:"rocketchat_token_file,omitempty"`
	RocketChatTokenID     string          `yaml:"rocketchat_token_id,omitempty"`
	RocketChatTokenIDFile string          `yaml:"rocketchat_token_id_file,omitempty"`
	SlackAppToken         string          `yaml:"slack_app_token,omitempty"`
	SlackAppTokenFile     string          `yaml:"slack_app_token_file,omitempty"`
	SlackAppURL           *config.URL     `yaml:"slack_app_url,omitempty"`
	WeChatAPISecretFile   string          `yaml:"wechat_api_secret_file,omitempty"`
}

type route struct {
	Receiver            string            `yaml:"receiver,omitempty"`
	GroupByStr          []string          `yaml:"group_by,omitempty"`
	Match               map[string]string `yaml:"match,omitempty"`
	MatchRE             map[string]string `yaml:"match_re,omitempty"`
	Matchers            []string          `yaml:"matchers,omitempty"`
	Continue            bool              `yaml:"continue,omitempty"`
	Routes              []*route          `yaml:"routes,omitempty"`
	GroupWait           string            `yaml:"group_wait,omitempty"`
	GroupInterval       string            `yaml:"group_interval,omitempty"`
	RepeatInterval      string            `yaml:"repeat_interval,omitempty"`
	MuteTimeIntervals   []string          `yaml:"mute_time_intervals,omitempty"`
	ActiveTimeIntervals []string          `yaml:"active_time_intervals,omitempty"`
}

type inhibitRule struct {
	TargetMatch    map[string]string `yaml:"target_match,omitempty"`
	TargetMatchRE  map[string]string `yaml:"target_match_re,omitempty"`
	TargetMatchers []string          `yaml:"target_matchers,omitempty"`
	SourceMatch    map[string]string `yaml:"source_match,omitempty"`
	SourceMatchRE  map[string]string `yaml:"source_match_re,omitempty"`
	SourceMatchers []string          `yaml:"source_matchers,omitempty"`
	Equal          []string          `yaml:"equal,omitempty"`
}

type receiver struct {
	Name              string              `yaml:"name"`
	OpsgenieConfigs   []*opsgenieConfig   `yaml:"opsgenie_configs,omitempty"`
	PagerdutyConfigs  []*pagerdutyConfig  `yaml:"pagerduty_configs,omitempty"`
	SlackConfigs      []*slackConfig      `yaml:"slack_configs,omitempty"`
	WebhookConfigs    []*webhookConfig    `yaml:"webhook_configs,omitempty"`
	WeChatConfigs     []*weChatConfig     `yaml:"wechat_configs,omitempty"`
	EmailConfigs      []*emailConfig      `yaml:"email_configs,omitempty"`
	PushoverConfigs   []*pushoverConfig   `yaml:"pushover_configs,omitempty"`
	VictorOpsConfigs  []*victorOpsConfig  `yaml:"victorops_configs,omitempty"`
	SNSConfigs        []*snsConfig        `yaml:"sns_configs,omitempty"`
	TelegramConfigs   []*telegramConfig   `yaml:"telegram_configs,omitempty"`
	DiscordConfigs    []*discordConfig    `yaml:"discord_configs,omitempty"`
	WebexConfigs      []*webexConfig      `yaml:"webex_configs,omitempty"`
	MSTeamsConfigs    []*msTeamsConfig    `yaml:"msteams_configs,omitempty"`
	MSTeamsV2Configs  []*msTeamsV2Config  `yaml:"msteamsv2_configs,omitempty"`
	JiraConfigs       []*jiraConfig       `yaml:"jira_configs,omitempty"`
	RocketChatConfigs []*rocketChatConfig `yaml:"rocketchat_configs,omitempty"`
	MattermostConfigs []*mattermostConfig `yaml:"mattermost_configs,omitempty"`
	IncidentioConfigs []*incidentioConfig `yaml:"incidentio_configs,omitempty"`
}

type webhookConfig struct {
	VSendResolved *bool             `yaml:"send_resolved,omitempty"`
	URL           string            `yaml:"url,omitempty"`
	URLFile       string            `yaml:"url_file,omitempty"`
	HTTPConfig    *httpClientConfig `yaml:"http_config,omitempty"`
	MaxAlerts     int32             `yaml:"max_alerts,omitempty"`
	Timeout       *model.Duration   `yaml:"timeout,omitempty"`
}

type pagerdutyConfig struct {
	VSendResolved  *bool             `yaml:"send_resolved,omitempty"`
	HTTPConfig     *httpClientConfig `yaml:"http_config,omitempty"`
	ServiceKey     string            `yaml:"service_key,omitempty"`
	ServiceKeyFile string            `yaml:"service_key_file,omitempty"`
	RoutingKey     string            `yaml:"routing_key,omitempty"`
	RoutingKeyFile string            `yaml:"routing_key_file,omitempty"`
	URL            string            `yaml:"url,omitempty"`
	Client         string            `yaml:"client,omitempty"`
	ClientURL      string            `yaml:"client_url,omitempty"`
	Description    string            `yaml:"description,omitempty"`
	Details        map[string]any    `yaml:"details,omitempty"`
	Images         []pagerdutyImage  `yaml:"images,omitempty"`
	Links          []pagerdutyLink   `yaml:"links,omitempty"`
	Severity       string            `yaml:"severity,omitempty"`
	Class          string            `yaml:"class,omitempty"`
	Component      string            `yaml:"component,omitempty"`
	Group          string            `yaml:"group,omitempty"`
	Source         string            `yaml:"source,omitempty"`
	Timeout        *model.Duration   `yaml:"timeout,omitempty"`
}

type opsgenieConfig struct {
	VSendResolved *bool               `yaml:"send_resolved,omitempty"`
	HTTPConfig    *httpClientConfig   `yaml:"http_config,omitempty"`
	APIKey        string              `yaml:"api_key,omitempty"`
	APIKeyFile    string              `yaml:"api_key_file,omitempty"`
	APIURL        string              `yaml:"api_url,omitempty"`
	Message       string              `yaml:"message,omitempty"`
	Description   string              `yaml:"description,omitempty"`
	Source        string              `yaml:"source,omitempty"`
	Details       map[string]string   `yaml:"details,omitempty"`
	Responders    []opsgenieResponder `yaml:"responders,omitempty"`
	Tags          string              `yaml:"tags,omitempty"`
	Note          string              `yaml:"note,omitempty"`
	Priority      string              `yaml:"priority,omitempty"`
	UpdateAlerts  *bool               `yaml:"update_alerts,omitempty"`
	Entity        string              `yaml:"entity,omitempty"`
	Actions       string              `yaml:"actions,omitempty"`
}

type weChatConfig struct {
	VSendResolved *bool             `yaml:"send_resolved,omitempty"`
	APISecret     string            `yaml:"api_secret,omitempty"`
	APIURL        string            `yaml:"api_url,omitempty"`
	CorpID        string            `yaml:"corp_id,omitempty"`
	AgentID       string            `yaml:"agent_id,omitempty"`
	ToUser        string            `yaml:"to_user,omitempty"`
	ToParty       string            `yaml:"to_party,omitempty"`
	ToTag         string            `yaml:"to_tag,omitempty"`
	Message       string            `yaml:"message,omitempty"`
	MessageType   string            `yaml:"message_type,omitempty"`
	HTTPConfig    *httpClientConfig `yaml:"http_config,omitempty"`
	APISecretFile string            `yaml:"api_secret_file,omitempty"`
}

type slackConfig struct {
	VSendResolved *bool             `yaml:"send_resolved,omitempty"`
	HTTPConfig    *httpClientConfig `yaml:"http_config,omitempty"`
	APIURL        string            `yaml:"api_url,omitempty"`
	APIURLFile    string            `yaml:"api_url_file,omitempty"`
	AppToken      string            `yaml:"app_token,omitempty"`
	AppTokenFile  string            `yaml:"app_token_file,omitempty"`
	AppURL        string            `yaml:"app_url,omitempty"`
	Channel       string            `yaml:"channel,omitempty"`
	Username      string            `yaml:"username,omitempty"`
	Color         string            `yaml:"color,omitempty"`
	Title         string            `yaml:"title,omitempty"`
	TitleLink     string            `yaml:"title_link,omitempty"`
	Pretext       string            `yaml:"pretext,omitempty"`
	Text          string            `yaml:"text,omitempty"`
	Fields        []slackField      `yaml:"fields,omitempty"`
	ShortFields   bool              `yaml:"short_fields,omitempty"`
	Footer        string            `yaml:"footer,omitempty"`
	Fallback      string            `yaml:"fallback,omitempty"`
	CallbackID    string            `yaml:"callback_id,omitempty"`
	IconEmoji     string            `yaml:"icon_emoji,omitempty"`
	IconURL       string            `yaml:"icon_url,omitempty"`
	ImageURL      string            `yaml:"image_url,omitempty"`
	ThumbURL      string            `yaml:"thumb_url,omitempty"`
	LinkNames     bool              `yaml:"link_names,omitempty"`
	MrkdwnIn      []string          `yaml:"mrkdwn_in,omitempty"`
	Actions       []slackAction     `yaml:"actions,omitempty"`
	Timeout       *model.Duration   `yaml:"timeout,omitempty"`
	MessageText   string            `yaml:"message_text,omitempty"`
}

type httpClientConfig struct {
	Authorization   *authorization     `yaml:"authorization,omitempty"`
	BasicAuth       *basicAuth         `yaml:"basic_auth,omitempty"`
	OAuth2          *oauth2            `yaml:"oauth2,omitempty"`
	BearerToken     string             `yaml:"bearer_token,omitempty"`
	BearerTokenFile string             `yaml:"bearer_token_file,omitempty"`
	TLSConfig       *tlsConfig         `yaml:"tls_config,omitempty"`
	FollowRedirects *bool              `yaml:"follow_redirects,omitempty"`
	EnableHTTP2     *bool              `yaml:"enable_http2,omitempty"`
	HTTPHeaders     *commoncfg.Headers `yaml:"http_headers,omitempty"`
	proxyConfig     `yaml:",inline"`
}

type proxyConfig struct {
	ProxyURL             string              `yaml:"proxy_url,omitempty"`
	NoProxy              string              `yaml:"no_proxy,omitempty"`
	ProxyFromEnvironment bool                `yaml:"proxy_from_environment,omitempty"`
	ProxyConnectHeader   map[string][]string `yaml:"proxy_connect_header,omitempty"`
}

type tlsConfig struct {
	CAFile             string `yaml:"ca_file,omitempty"`
	CertFile           string `yaml:"cert_file,omitempty"`
	KeyFile            string `yaml:"key_file,omitempty"`
	ServerName         string `yaml:"server_name,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
	MinVersion         string `yaml:"min_version,omitempty"`
	MaxVersion         string `yaml:"max_version,omitempty"`
}

type authorization struct {
	Type            string `yaml:"type,omitempty"`
	Credentials     string `yaml:"credentials,omitempty"`
	CredentialsFile string `yaml:"credentials_file,omitempty"`
}

type basicAuth struct {
	Username     string `yaml:"username,omitempty"`
	Password     string `yaml:"password,omitempty"`
	PasswordFile string `yaml:"password_file,omitempty"`
}

type oauth2 struct {
	ClientID         string            `yaml:"client_id"`
	ClientSecret     string            `yaml:"client_secret"`
	ClientSecretFile string            `yaml:"client_secret_file,omitempty"`
	Scopes           []string          `yaml:"scopes,omitempty"`
	TokenURL         string            `yaml:"token_url"`
	EndpointParams   map[string]string `yaml:"endpoint_params,omitempty"`
	proxyConfig      `yaml:",inline"`

	TLSConfig *tlsConfig `yaml:"tls_config,omitempty"`
}

type pagerdutyLink struct {
	Href string `yaml:"href,omitempty"`
	Text string `yaml:"text,omitempty"`
}

type pagerdutyImage struct {
	Src  string `yaml:"src,omitempty"`
	Alt  string `yaml:"alt,omitempty"`
	Href string `yaml:"href,omitempty"`
}

type opsgenieResponder struct {
	ID       string `yaml:"id,omitempty"`
	Name     string `yaml:"name,omitempty"`
	Username string `yaml:"username,omitempty"`
	Type     string `yaml:"type,omitempty"`
}

type slackField struct {
	Title string `yaml:"title,omitempty"`
	Value string `yaml:"value,omitempty"`
	Short bool   `yaml:"short,omitempty"`
}

type slackAction struct {
	Type         string                  `yaml:"type,omitempty"`
	Text         string                  `yaml:"text,omitempty"`
	URL          string                  `yaml:"url,omitempty"`
	Style        string                  `yaml:"style,omitempty"`
	Name         string                  `yaml:"name,omitempty"`
	Value        string                  `yaml:"value,omitempty"`
	ConfirmField *slackConfirmationField `yaml:"confirm,omitempty"`
}

type slackConfirmationField struct {
	Text        string `yaml:"text,omitempty"`
	Title       string `yaml:"title,omitempty"`
	OkText      string `yaml:"ok_text,omitempty"`
	DismissText string `yaml:"dismiss_text,omitempty"`
}

type emailConfig struct {
	VSendResolved    *bool             `yaml:"send_resolved,omitempty"`
	To               string            `yaml:"to,omitempty"`
	From             string            `yaml:"from,omitempty"`
	Hello            string            `yaml:"hello,omitempty"`
	Smarthost        config.HostPort   `yaml:"smarthost,omitempty"`
	AuthUsername     string            `yaml:"auth_username,omitempty"`
	AuthPassword     string            `yaml:"auth_password,omitempty"`
	AuthPasswordFile string            `yaml:"auth_password_file,omitempty"`
	AuthSecret       string            `yaml:"auth_secret,omitempty"`
	AuthSecretFile   string            `yaml:"auth_secret_file,omitempty"`
	AuthIdentity     string            `yaml:"auth_identity,omitempty"`
	Headers          map[string]string `yaml:"headers,omitempty"`
	HTML             *string           `yaml:"html,omitempty"`
	Text             *string           `yaml:"text,omitempty"`
	RequireTLS       *bool             `yaml:"require_tls,omitempty"`
	TLSConfig        *tlsConfig        `yaml:"tls_config,omitempty"`
	ForceImplicitTLS *bool             `yaml:"force_implicit_tls,omitempty"`
}

type pushoverConfig struct {
	VSendResolved *bool             `yaml:"send_resolved,omitempty"`
	HTTPConfig    *httpClientConfig `yaml:"http_config,omitempty"`
	UserKey       string            `yaml:"user_key,omitempty"`
	UserKeyFile   string            `yaml:"user_key_file,omitempty"`
	Token         string            `yaml:"token,omitempty"`
	TokenFile     string            `yaml:"token_file,omitempty"`
	Title         string            `yaml:"title,omitempty"`
	Message       string            `yaml:"message,omitempty"`
	URL           string            `yaml:"url,omitempty"`
	URLTitle      string            `yaml:"url_title,omitempty"`
	TTL           string            `yaml:"ttl,omitempty"`
	Device        string            `yaml:"device,omitempty"`
	Sound         string            `yaml:"sound,omitempty"`
	Priority      string            `yaml:"priority,omitempty"`
	Retry         *model.Duration   `yaml:"retry,omitempty"`
	Expire        *model.Duration   `yaml:"expire,omitempty"`
	HTML          *bool             `yaml:"html,omitempty"`
	Monospace     *bool             `yaml:"monospace,omitempty"`
}

type snsConfig struct {
	VSendResolved *bool             `yaml:"send_resolved,omitempty"`
	HTTPConfig    *httpClientConfig `yaml:"http_config,omitempty"`
	APIUrl        string            `yaml:"api_url,omitempty"`
	Sigv4         sigV4Config       `yaml:"sigv4,omitempty"`
	TopicARN      string            `yaml:"topic_arn,omitempty"`
	PhoneNumber   string            `yaml:"phone_number,omitempty"`
	TargetARN     string            `yaml:"target_arn,omitempty"`
	Subject       string            `yaml:"subject,omitempty"`
	Message       string            `yaml:"message,omitempty"`
	Attributes    map[string]string `yaml:"attributes,omitempty"`
}

type telegramConfig struct {
	VSendResolved        *bool             `yaml:"send_resolved,omitempty"`
	APIUrl               string            `yaml:"api_url,omitempty"`
	BotToken             string            `yaml:"bot_token,omitempty"`
	BotTokenFile         string            `yaml:"bot_token_file,omitempty"`
	ChatID               int64             `yaml:"chat_id,omitempty"`
	ChatIDFile           string            `yaml:"chat_id_file,omitempty"`
	MessageThreadID      int               `yaml:"message_thread_id,omitempty"`
	Message              string            `yaml:"message,omitempty"`
	DisableNotifications bool              `yaml:"disable_notifications,omitempty"`
	ParseMode            string            `yaml:"parse_mode,omitempty"`
	HTTPConfig           *httpClientConfig `yaml:"http_config,omitempty"`
}

type discordConfig struct {
	VSendResolved *bool             `yaml:"send_resolved,omitempty"`
	HTTPConfig    *httpClientConfig `yaml:"http_config,omitempty"`
	WebhookURL    string            `yaml:"webhook_url,omitempty"`
	Title         string            `yaml:"title,omitempty"`
	Message       string            `yaml:"message,omitempty"`
	Content       string            `yaml:"content,omitempty"`
	Username      string            `yaml:"username,omitempty"`
	AvatarURL     string            `yaml:"avatar_url,omitempty"`
}

type webexConfig struct {
	VSendResolved *bool             `yaml:"send_resolved,omitempty"`
	HTTPConfig    *httpClientConfig `yaml:"http_config,omitempty"`
	APIURL        string            `yaml:"api_url,omitempty"`
	Message       string            `yaml:"message,omitempty"`
	RoomID        string            `yaml:"room_id"`
}

type sigV4Config struct {
	Region    string `yaml:"region,omitempty"`
	AccessKey string `yaml:"access_key,omitempty"`
	SecretKey string `yaml:"secret_key,omitempty"`
	Profile   string `yaml:"profile,omitempty"`
	RoleARN   string `yaml:"role_arn,omitempty"`
}

type victorOpsConfig struct {
	VSendResolved     *bool             `yaml:"send_resolved,omitempty"`
	HTTPConfig        *httpClientConfig `yaml:"http_config,omitempty"`
	APIKey            string            `yaml:"api_key,omitempty"`
	APIKeyFile        string            `yaml:"api_key_file,omitempty"`
	APIURL            string            `yaml:"api_url,omitempty"`
	RoutingKey        string            `yaml:"routing_key,omitempty"`
	MessageType       string            `yaml:"message_type,omitempty"`
	StateMessage      string            `yaml:"state_message,omitempty"`
	EntityDisplayName string            `yaml:"entity_display_name,omitempty"`
	MonitoringTool    string            `yaml:"monitoring_tool,omitempty"`
	CustomFields      map[string]string `yaml:"custom_fields,omitempty"`
}

type msTeamsConfig struct {
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
	WebhookURL   string            `yaml:"webhook_url"`
	Title        string            `yaml:"title,omitempty"`
	Summary      string            `yaml:"summary,omitempty"`
	Text         string            `yaml:"text,omitempty"`
	HTTPConfig   *httpClientConfig `yaml:"http_config,omitempty"`
}

type msTeamsV2Config struct {
	SendResolved   *bool             `yaml:"send_resolved,omitempty"`
	WebhookURL     string            `yaml:"webhook_url,omitempty"`
	WebhookURLFile string            `yaml:"webhook_url_file,omitempty"`
	Title          string            `yaml:"title,omitempty"`
	Text           string            `yaml:"text,omitempty"`
	HTTPConfig     *httpClientConfig `yaml:"http_config,omitempty"`
}

type jiraConfig struct {
	HTTPConfig        *httpClientConfig `yaml:"http_config,omitempty"`
	SendResolved      *bool             `yaml:"send_resolved,omitempty"`
	APIURL            string            `yaml:"api_url,omitempty"`
	Project           string            `yaml:"project,omitempty"`
	Summary           string            `yaml:"summary,omitempty"`
	Description       string            `yaml:"description,omitempty"`
	Labels            []string          `yaml:"labels,omitempty"`
	Priority          string            `yaml:"priority,omitempty"`
	IssueType         string            `yaml:"issue_type,omitempty"`
	ReopenTransition  string            `yaml:"reopen_transition,omitempty"`
	ResolveTransition string            `yaml:"resolve_transition,omitempty"`
	WontFixResolution string            `yaml:"wont_fix_resolution,omitempty"`
	ReopenDuration    model.Duration    `yaml:"reopen_duration,omitempty"`
	Fields            map[string]any    `yaml:"fields,omitempty"`
	APIType           string            `yaml:"api_type,omitempty"`
}

type rocketchatAttachmentField struct {
	Short *bool  `yaml:"short"`
	Title string `yaml:"title,omitempty"`
	Value string `yaml:"value,omitempty"`
}

type rocketchatAttachmentAction struct {
	Type               string `yaml:"type,omitempty"`
	Text               string `yaml:"text,omitempty"`
	URL                string `yaml:"url,omitempty"`
	ImageURL           string `yaml:"image_url,omitempty"`
	IsWebView          bool   `yaml:"is_webview"`
	WebviewHeightRatio string `yaml:"webview_height_ratio,omitempty"`
	Msg                string `yaml:"msg,omitempty"`
	MsgInChatWindow    bool   `yaml:"msg_in_chat_window"`
	MsgProcessingType  string `yaml:"msg_processing_type,omitempty"`
}

type rocketChatConfig struct {
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
	HTTPConfig   *httpClientConfig `yaml:"http_config,omitempty"`
	APIURL       string            `yaml:"api_url,omitempty"`
	TokenID      *string           `yaml:"token_id,omitempty"`
	TokenIDFile  string            `yaml:"token_id_file,omitempty"`
	Token        *string           `yaml:"token,omitempty"`
	TokenFile    string            `yaml:"token_file,omitempty"`
	// RocketChat channel override, (like #other-channel or @username).
	Channel     string                        `yaml:"channel,omitempty"`
	Color       string                        `yaml:"color,omitempty"`
	Title       string                        `yaml:"title,omitempty"`
	TitleLink   string                        `yaml:"title_link,omitempty"`
	Text        string                        `yaml:"text,omitempty"`
	Fields      []*rocketchatAttachmentField  `yaml:"fields,omitempty"`
	ShortFields bool                          `yaml:"short_fields"`
	Emoji       string                        `yaml:"emoji,omitempty"`
	IconURL     string                        `yaml:"icon_url,omitempty"`
	ImageURL    string                        `yaml:"image_url,omitempty"`
	ThumbURL    string                        `yaml:"thumb_url,omitempty"`
	LinkNames   bool                          `yaml:"link_names"`
	Actions     []*rocketchatAttachmentAction `yaml:"actions,omitempty"`
}

type mattermostConfig struct {
	SendResolved   *bool                         `yaml:"send_resolved,omitempty" json:"send_resolved,omitempty"`
	WebhookURL     string                        `yaml:"webhook_url,omitempty" json:"webhook_url,omitempty"`
	WebhookURLFile string                        `yaml:"webhook_url_file,omitempty" json:"webhook_url_file,omitempty"`
	Channel        string                        `yaml:"channel,omitempty" json:"channel,omitempty"`
	Username       string                        `yaml:"username,omitempty" json:"username,omitempty"`
	Text           string                        `yaml:"text,omitempty" json:"text,omitempty"`
	IconURL        string                        `yaml:"icon_url,omitempty" json:"icon_url,omitempty"`
	IconEmoji      string                        `yaml:"icon_emoji,omitempty" json:"icon_emoji,omitempty"`
	Attachments    []*mattermostAttachmentConfig `yaml:"attachments,omitempty" json:"attachments,omitempty"`
	Props          *mattermostPropsConfig        `yaml:"props,omitempty" json:"props,omitempty"`
	Priority       *mattermostPriorityConfig     `yaml:"priority,omitempty" json:"priority,omitempty"`
	HTTPConfig     *httpClientConfig             `yaml:"http_config,omitempty" json:"http_config,omitempty"`
}

type mattermostAttachmentConfig struct {
	Fallback   string            `yaml:"fallback,omitempty" json:"fallback,omitempty"`
	Color      string            `yaml:"color,omitempty" json:"color,omitempty"`
	Pretext    string            `yaml:"pretext,omitempty" json:"pretext,omitempty"`
	Text       string            `yaml:"text,omitempty" json:"text,omitempty"`
	AuthorName string            `yaml:"author_name,omitempty" json:"author_name,omitempty"`
	AuthorLink string            `yaml:"author_link,omitempty" json:"author_link,omitempty"`
	AuthorIcon string            `yaml:"author_icon,omitempty" json:"author_icon,omitempty"`
	Title      string            `yaml:"title,omitempty" json:"title,omitempty"`
	TitleLink  string            `yaml:"title_link,omitempty" json:"title_link,omitempty"`
	Fields     []mattermostField `yaml:"fields,omitempty" json:"fields,omitempty"`
	ThumbURL   string            `yaml:"thumb_url,omitempty" json:"thumb_url,omitempty"`
	Footer     string            `yaml:"footer,omitempty" json:"footer,omitempty"`
	FooterIcon string            `yaml:"footer_icon,omitempty" json:"footer_icon,omitempty"`
	ImageURL   string            `yaml:"image_url,omitempty" json:"image_url,omitempty"`
}

type mattermostField struct {
	Title string `yaml:"title,omitempty" json:"title,omitempty"`
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
	Short bool   `yaml:"short,omitempty" json:"short,omitempty"`
}

type mattermostPropsConfig struct {
	Card *string `yaml:"card,omitempty" json:"card,omitempty"`
}

type mattermostPriorityConfig struct {
	Priority                string `yaml:"priority,omitempty" json:"priority,omitempty"`
	RequestedAck            *bool  `yaml:"requested_ack,omitempty" json:"requested_ack,omitempty"`
	PersistentNotifications *bool  `yaml:"persistent_notifications,omitempty" json:"persistent_notifications,omitempty"`
}

type incidentioConfig struct {
	VSendResolved        *bool             `yaml:"send_resolved,omitempty"`
	HTTPConfig           *httpClientConfig `yaml:"http_config,omitempty"`
	URL                  string            `yaml:"url,omitempty"`
	URLFile              string            `yaml:"url_file,omitempty"`
	AlertSourceToken     string            `yaml:"alert_source_token,omitempty"`
	AlertSourceTokenFile string            `yaml:"alert_source_token_file,omitempty"`
	MaxAlerts            *int32            `yaml:"max_alerts,omitempty"`
	Timeout              *model.Duration   `yaml:"timeout,omitempty"`
}

type timeInterval config.TimeInterval
