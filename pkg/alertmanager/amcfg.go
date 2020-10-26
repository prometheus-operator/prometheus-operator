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
	"context"
	"fmt"
	"path"
	"sort"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus/alertmanager/config"
	commoncfg "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"
)

// Customization of Config type from alertmanager repo:
// https://github.com/prometheus/alertmanager/blob/master/config/config.go
//
// Custom global type to get around obfuscation of secret values when
// marshalling. See the following issue for details:
// https://github.com/prometheus/alertmanager/issues/1985
type alertmanagerConfig struct {
	Global       *globalConfig  `yaml:"global,omitempty" json:"global,omitempty"`
	Route        *route         `yaml:"route,omitempty" json:"route,omitempty"`
	InhibitRules []*inhibitRule `yaml:"inhibit_rules,omitempty" json:"inhibit_rules,omitempty"`
	Receivers    []*receiver    `yaml:"receivers,omitempty" json:"receivers,omitempty"`
	Templates    []string       `yaml:"templates" json:"templates"`
}

type globalConfig struct {
	// ResolveTimeout is the time after which an alert is declared resolved
	// if it has not been updated.
	ResolveTimeout model.Duration `yaml:"resolve_timeout" json:"resolve_timeout"`

	HTTPConfig *commoncfg.HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	SMTPFrom         string          `yaml:"smtp_from,omitempty" json:"smtp_from,omitempty"`
	SMTPHello        string          `yaml:"smtp_hello,omitempty" json:"smtp_hello,omitempty"`
	SMTPSmarthost    config.HostPort `yaml:"smtp_smarthost,omitempty" json:"smtp_smarthost,omitempty"`
	SMTPAuthUsername string          `yaml:"smtp_auth_username,omitempty" json:"smtp_auth_username,omitempty"`
	SMTPAuthPassword string          `yaml:"smtp_auth_password,omitempty" json:"smtp_auth_password,omitempty"`
	SMTPAuthSecret   string          `yaml:"smtp_auth_secret,omitempty" json:"smtp_auth_secret,omitempty"`
	SMTPAuthIdentity string          `yaml:"smtp_auth_identity,omitempty" json:"smtp_auth_identity,omitempty"`
	SMTPRequireTLS   bool            `yaml:"smtp_require_tls,omitempty" json:"smtp_require_tls,omitempty"`
	SlackAPIURL      *config.URL     `yaml:"slack_api_url,omitempty" json:"slack_api_url,omitempty"`
	PagerdutyURL     *config.URL     `yaml:"pagerduty_url,omitempty" json:"pagerduty_url,omitempty"`
	HipchatAPIURL    *config.URL     `yaml:"hipchat_api_url,omitempty" json:"hipchat_api_url,omitempty"`
	HipchatAuthToken string          `yaml:"hipchat_auth_token,omitempty" json:"hipchat_auth_token,omitempty"`
	OpsGenieAPIURL   *config.URL     `yaml:"opsgenie_api_url,omitempty" json:"opsgenie_api_url,omitempty"`
	OpsGenieAPIKey   string          `yaml:"opsgenie_api_key,omitempty" json:"opsgenie_api_key,omitempty"`
	WeChatAPIURL     *config.URL     `yaml:"wechat_api_url,omitempty" json:"wechat_api_url,omitempty"`
	WeChatAPISecret  string          `yaml:"wechat_api_secret,omitempty" json:"wechat_api_secret,omitempty"`
	WeChatAPICorpID  string          `yaml:"wechat_api_corp_id,omitempty" json:"wechat_api_corp_id,omitempty"`
	VictorOpsAPIURL  *config.URL     `yaml:"victorops_api_url,omitempty" json:"victorops_api_url,omitempty"`
	VictorOpsAPIKey  string          `yaml:"victorops_api_key,omitempty" json:"victorops_api_key,omitempty"`
}

type route struct {
	Receiver       string            `yaml:"receiver,omitempty" json:"receiver,omitempty"`
	GroupByStr     []string          `yaml:"group_by,omitempty" json:"group_by,omitempty"`
	Match          map[string]string `yaml:"match,omitempty" json:"match,omitempty"`
	MatchRE        map[string]string `yaml:"match_re,omitempty" json:"match_re,omitempty"`
	Continue       bool              `yaml:"continue,omitempty" json:"continue,omitempty"`
	Routes         []*route          `yaml:"routes,omitempty" json:"routes,omitempty"`
	GroupWait      string            `yaml:"group_wait,omitempty" json:"group_wait,omitempty"`
	GroupInterval  string            `yaml:"group_interval,omitempty" json:"group_interval,omitempty"`
	RepeatInterval string            `yaml:"repeat_interval,omitempty" json:"repeat_interval,omitempty"`
}

type inhibitRule struct {
	TargetMatch   map[string]string `yaml:"target_match,omitempty" json:"target_match,omitempty"`
	TargetMatchRE map[string]string `yaml:"target_match_re,omitempty" json:"target_match_re,omitempty"`
	SourceMatch   map[string]string `yaml:"source_match,omitempty" json:"source_match,omitempty"`
	SourceMatchRE map[string]string `yaml:"source_match_re,omitempty" json:"source_match_re,omitempty"`
	Equal         []string          `yaml:"equal,omitempty" json:"equal,omitempty"`
}

type receiver struct {
	Name             string             `yaml:"name" json:"name"`
	OpsgenieConfigs  []*opsgenieConfig  `yaml:"opsgenie_configs,omitempty" json:"opsgenie_configs,omitempty"`
	PagerdutyConfigs []*pagerdutyConfig `yaml:"pagerduty_configs,omitempty" json:"pagerduty_configs,omitempty"`
	WebhookConfigs   []*webhookConfig   `yaml:"webhook_configs,omitempty" json:"webhook_configs,omitempty"`
	WeChatConfigs    []*weChatConfig    `yaml:"wechat_configs,omitempty" json:"wechat_config,omitempty"`
}

type webhookConfig struct {
	VSendResolved bool              `yaml:"send_resolved" json:"send_resolved"`
	URL           string            `yaml:"url,omitempty" json:"url,omitempty"`
	HTTPConfig    *httpClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`
	MaxAlerts     int32             `yaml:"max_alerts,omitempty" json:"max_alerts,omitempty"`
}

type pagerdutyConfig struct {
	VSendResolved bool              `yaml:"send_resolved" json:"send_resolved"`
	HTTPConfig    *httpClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`
	ServiceKey    string            `yaml:"service_key,omitempty" json:"service_key,omitempty"`
	RoutingKey    string            `yaml:"routing_key,omitempty" json:"routing_key,omitempty"`
	URL           string            `yaml:"url,omitempty" json:"url,omitempty"`
	Client        string            `yaml:"client,omitempty" json:"client,omitempty"`
	ClientURL     string            `yaml:"client_url,omitempty" json:"client_url,omitempty"`
	Description   string            `yaml:"description,omitempty" json:"description,omitempty"`
	Details       map[string]string `yaml:"details,omitempty" json:"details,omitempty"`
	Images        []pagerdutyImage  `yaml:"images,omitempty" json:"images,omitempty"`
	Links         []pagerdutyLink   `yaml:"links,omitempty" json:"links,omitempty"`
	Severity      string            `yaml:"severity,omitempty" json:"severity,omitempty"`
	Class         string            `yaml:"class,omitempty" json:"class,omitempty"`
	Component     string            `yaml:"component,omitempty" json:"component,omitempty"`
	Group         string            `yaml:"group,omitempty" json:"group,omitempty"`
}

type opsgenieConfig struct {
	VSendResolved bool                `yaml:"send_resolved" json:"send_resolved"`
	HTTPConfig    *httpClientConfig   `yaml:"http_config,omitempty" json:"http_config,omitempty"`
	APIKey        string              `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	APIURL        string              `yaml:"api_url,omitempty" json:"api_url,omitempty"`
	Message       string              `yaml:"message,omitempty" json:"message,omitempty"`
	Description   string              `yaml:"description,omitempty" json:"description,omitempty"`
	Source        string              `yaml:"source,omitempty" json:"source,omitempty"`
	Details       map[string]string   `yaml:"details,omitempty" json:"details,omitempty"`
	Responders    []opsgenieResponder `yaml:"responders,omitempty" json:"responders,omitempty"`
	Tags          string              `yaml:"tags,omitempty" json:"tags,omitempty"`
	Note          string              `yaml:"note,omitempty" json:"note,omitempty"`
	Priority      string              `yaml:"priority,omitempty" json:"priority,omitempty"`
}

type weChatConfig struct {
	VSendResolved bool              `yaml:"send_resolved" json:"send_resolved"`
	APISecret     string            `yaml:"api_secret,omitempty" json:"api_secret,omitempty"`
	APIURL        string            `yaml:"api_url,omitempty" json:"api_url,omitempty"`
	CorpID        string            `yaml:"corp_id,omitempty" json:"corp_id,omitempty"`
	AgentID       string            `yaml:"agent_id,omitempty" json:"agent_id,omitempty"`
	ToUser        string            `yaml:"to_user,omitempty" json:"to_user,omitempty"`
	ToParty       string            `yaml:"to_party,omitempty" json:"to_party,omitempty"`
	ToTag         string            `yaml:"to_tag,omitempty" json:"to_tag,omitempty"`
	Message       string            `yaml:"message,omitempty" json:"message,omitempty"`
	MessageType   string            `yaml:"message_type,omitempty" json:"message,omitempty"`
	HTTPConfig    *httpClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`
}

type httpClientConfig struct {
	BasicAuth       *basicAuth          `yaml:"basic_auth,omitempty"`
	BearerToken     string              `yaml:"bearer_token,omitempty"`
	BearerTokenFile string              `yaml:"bearer_token_file,omitempty"`
	ProxyURL        string              `yaml:"proxy_url,omitempty"`
	TLSConfig       commoncfg.TLSConfig `yaml:"tls_config,omitempty"`
}

type basicAuth struct {
	Username     string `yaml:"username"`
	Password     string `yaml:"password,omitempty"`
	PasswordFile string `yaml:"password_file,omitempty"`
}

type pagerdutyLink struct {
	Href string `yaml:"href,omitempty" json:"href,omitempty"`
	Text string `yaml:"text,omitempty" json:"text,omitempty"`
}

type pagerdutyImage struct {
	Src  string `yaml:"src,omitempty" json:"src,omitempty"`
	Alt  string `yaml:"alt,omitempty" json:"alt,omitempty"`
	Href string `yaml:"href,omitempty" json:"href,omitempty"`
}

type opsgenieResponder struct {
	ID       string `yaml:"id,omitempty" json:"id,omitempty"`
	Name     string `yaml:"name,omitempty" json:"name,omitempty"`
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Type     string `yaml:"type,omitempty" json:"type,omitempty"`
}

func loadCfg(s string) (*alertmanagerConfig, error) {
	// Run upstream Load function to get any validation checks that it runs.
	_, err := config.Load(s)
	if err != nil {
		return nil, err
	}

	cfg := &alertmanagerConfig{}
	err = yaml.UnmarshalStrict([]byte(s), cfg)

	return cfg, nil
}

func (c alertmanagerConfig) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("<error creating config string: %s>", err)
	}
	return string(b)
}

type configGenerator struct {
	logger log.Logger
	store  *assets.Store
}

func newConfigGenerator(logger log.Logger, store *assets.Store) *configGenerator {
	cg := &configGenerator{
		logger: logger,
		store:  store,
	}
	return cg
}

func (cg *configGenerator) generateConfig(
	ctx context.Context,
	baseConfig alertmanagerConfig,
	amConfigs map[string]*monitoringv1alpha1.AlertmanagerConfig,
) ([]byte, error) {
	// amConfigIdentifiers is a sorted slice of keys from
	// amConfigs map, used to always generate the config in the
	// same order
	amConfigIdentifiers := make([]string, len(amConfigs))
	i := 0
	for k := range amConfigs {
		amConfigIdentifiers[i] = k
		i++
	}
	sort.Strings(amConfigIdentifiers)

	subRoutes := []*route{}
	for _, amConfigIdentifier := range amConfigIdentifiers {
		crKey := types.NamespacedName{
			Name:      amConfigs[amConfigIdentifier].Name,
			Namespace: amConfigs[amConfigIdentifier].Namespace,
		}

		// add routes to subRoutes
		if amConfigs[amConfigIdentifier].Spec.Route != nil {
			subRoutes = append(subRoutes, convertRoute(amConfigs[amConfigIdentifier].Spec.Route, crKey, true))
		}

		// add receivers to baseConfig.Receivers
		for _, receiver := range amConfigs[amConfigIdentifier].Spec.Receivers {
			receivers, err := cg.convertReceiver(ctx, &receiver, crKey)
			if err != nil {
				return nil, errors.Wrapf(err, "AlertmanagerConfig %s", crKey.String())
			}
			baseConfig.Receivers = append(baseConfig.Receivers, receivers)
		}

		// add inhibitRules to baseConfig.InhibitRules
		for _, inhibitRule := range amConfigs[amConfigIdentifier].Spec.InhibitRules {
			baseConfig.InhibitRules = append(baseConfig.InhibitRules, convertInhibitRule(&inhibitRule, crKey))
		}
	}

	// Append subroutes from base to the end, then replace with the new slice
	subRoutes = append(subRoutes, baseConfig.Route.Routes...)
	baseConfig.Route.Routes = subRoutes

	return yaml.Marshal(baseConfig)
}

func convertRoute(in *monitoringv1alpha1.Route, crKey types.NamespacedName, firstLevelRoute bool) *route {

	// Enforce continue to be true for main Route in a CR
	cont := in.Continue
	if firstLevelRoute {
		cont = true
	}

	match := map[string]string{}
	matchRE := map[string]string{}

	for _, matcher := range in.Matchers {
		if matcher.Regex {
			matchRE[matcher.Name] = matcher.Value
		} else {
			match[matcher.Name] = matcher.Value
		}
	}
	if firstLevelRoute {
		match["namespace"] = crKey.Namespace
		delete(matchRE, "namespace")
	}

	// Set to nil if empty so that it doesn't show up in resulting yaml
	if len(match) == 0 {
		match = nil
	}
	// Set to nil if empty so that it doesn't show up in resulting yaml
	if len(matchRE) == 0 {
		matchRE = nil
	}

	var routes []*route
	if len(in.Routes) > 0 {
		routes := make([]*route, len(in.Routes))
		for i := range in.Routes {
			routes[i] = convertRoute(&in.Routes[i], crKey, false)
		}
	}

	receiver := prefixReceiverName(in.Receiver, crKey)

	return &route{
		Receiver:       receiver,
		GroupByStr:     in.GroupBy,
		GroupWait:      in.GroupWait,
		GroupInterval:  in.GroupInterval,
		RepeatInterval: in.RepeatInterval,
		Continue:       cont,
		Match:          match,
		MatchRE:        matchRE,
		Routes:         routes,
	}
}

func (cg *configGenerator) convertReceiver(ctx context.Context, in *monitoringv1alpha1.Receiver, crKey types.NamespacedName) (*receiver, error) {
	var pagerdutyConfigs []*pagerdutyConfig

	if l := len(in.PagerDutyConfigs); l > 0 {
		pagerdutyConfigs = make([]*pagerdutyConfig, l)
		for i := range in.PagerDutyConfigs {
			receiver, err := cg.convertPagerdutyConfig(ctx, in.PagerDutyConfigs[i], crKey)
			if err != nil {
				return nil, errors.Wrapf(err, "PagerDutyConfig[%d]", i)
			}
			pagerdutyConfigs[i] = receiver
		}
	}

	var webhookConfigs []*webhookConfig
	if l := len(in.WebhookConfigs); l > 0 {
		webhookConfigs = make([]*webhookConfig, l)
		for i := range in.WebhookConfigs {
			receiver, err := cg.convertWebhookConfig(ctx, in.WebhookConfigs[i], crKey)
			if err != nil {
				return nil, errors.Wrapf(err, "WebhookConfig[%d]", i)
			}
			webhookConfigs[i] = receiver
		}
	}

	var opsgenieConfigs []*opsgenieConfig
	if l := len(in.OpsGenieConfigs); l > 0 {
		opsgenieConfigs = make([]*opsgenieConfig, l)
		for i := range in.OpsGenieConfigs {
			receiver, err := cg.convertOpsgenieConfig(ctx, in.OpsGenieConfigs[i], crKey)
			if err != nil {
				return nil, errors.Wrapf(err, "OpsGenieConfigs[%d]", i)
			}
			opsgenieConfigs[i] = receiver
		}
	}

	var weChatConfigs []*weChatConfig
	if l := len(in.WeChatConfigs); l > 0 {
		weChatConfigs = make([]*weChatConfig, l)
		for i := range in.WeChatConfigs {
			receiver, err := cg.convertWeChatConfig(ctx, in.WeChatConfigs[i], crKey)
			if err != nil {
				return nil, errors.Wrapf(err, "WeChatConfig[%d]", i)
			}
			weChatConfigs[i] = receiver
		}
	}

	return &receiver{
		Name:             prefixReceiverName(in.Name, crKey),
		OpsgenieConfigs:  opsgenieConfigs,
		PagerdutyConfigs: pagerdutyConfigs,
		WebhookConfigs:   webhookConfigs,
		WeChatConfigs:    weChatConfigs,
	}, nil
}

func (cg *configGenerator) convertWebhookConfig(ctx context.Context, in monitoringv1alpha1.WebhookConfig, crKey types.NamespacedName) (*webhookConfig, error) {
	out := &webhookConfig{}

	if in.SendResolved != nil {
		out.VSendResolved = *in.SendResolved
	}

	if in.URLSecret != nil {
		url, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.URLSecret)
		if err != nil {
			return nil, errors.Errorf("failed to get key %q from secret %q", in.URLSecret.Key, in.URLSecret.Name)
		}
		out.URL = url
	} else if in.URL != nil {
		out.URL = *in.URL
	}

	if in.HTTPConfig != nil {
		httpConfig, err := cg.convertHTTPConfig(ctx, *in.HTTPConfig, crKey)
		if err != nil {
			return nil, err
		}
		out.HTTPConfig = httpConfig
	}

	if in.MaxAlerts != nil {
		out.MaxAlerts = *in.MaxAlerts
	}

	return out, nil
}

func (cg *configGenerator) convertPagerdutyConfig(ctx context.Context, in monitoringv1alpha1.PagerDutyConfig, crKey types.NamespacedName) (*pagerdutyConfig, error) {
	out := &pagerdutyConfig{}

	if in.SendResolved != nil {
		out.VSendResolved = *in.SendResolved
	}

	if in.RoutingKey != nil {
		routingKey, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.RoutingKey)
		if err != nil {
			return nil, errors.Errorf("failed to get routing key %q from secret %q", in.RoutingKey.Key, in.RoutingKey.Name)
		}
		out.RoutingKey = routingKey
	}

	if in.ServiceKey != nil {
		serviceKey, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.ServiceKey)
		if err != nil {
			return nil, errors.Errorf("failed to get service key %q from secret %q", in.ServiceKey.Key, in.ServiceKey.Name)
		}
		out.ServiceKey = serviceKey
	}

	if in.URL != nil {
		out.URL = *in.URL
	}

	if in.Client != nil {
		out.Client = *in.Client
	}

	if in.ClientURL != nil {
		out.ClientURL = *in.ClientURL
	}

	if in.Description != nil {
		out.Description = *in.Description
	}

	if in.Severity != nil {
		out.Severity = *in.Severity
	}

	if in.Class != nil {
		out.Class = *in.Class
	}

	if in.Group != nil {
		out.Group = *in.Group
	}

	if in.Component != nil {
		out.Component = *in.Component
	}

	var details map[string]string
	if l := len(in.Details); l > 0 {
		details = make(map[string]string, l)
		for _, d := range in.Details {
			details[d.Key] = d.Value
		}
	}
	out.Details = details

	if in.HTTPConfig != nil {
		httpConfig, err := cg.convertHTTPConfig(ctx, *in.HTTPConfig, crKey)
		if err != nil {
			return nil, err
		}
		out.HTTPConfig = httpConfig
	}

	return out, nil
}

func (cg *configGenerator) convertOpsgenieConfig(ctx context.Context, in monitoringv1alpha1.OpsGenieConfig, crKey types.NamespacedName) (*opsgenieConfig, error) {
	out := &opsgenieConfig{}

	if in.SendResolved != nil {
		out.VSendResolved = *in.SendResolved
	}

	if in.APIKey != nil {
		apiKey, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.APIKey)
		if err != nil {
			return nil, errors.Errorf("failed to get api key %q from secret %q", in.APIKey.Key, in.APIKey.Name)
		}
		out.APIKey = apiKey
	}

	if in.APIURL != nil {
		out.APIURL = *in.APIURL
	}

	if in.Message != nil {
		out.Message = *in.Message
	}

	if in.Description != nil {
		out.Description = *in.Description
	}

	if in.Source != nil {
		out.Source = *in.Source
	}

	if in.Tags != nil {
		out.Tags = *in.Tags
	}

	if in.Note != nil {
		out.Note = *in.Note
	}

	if in.Priority != nil {
		out.Priority = *in.Priority
	}

	var details map[string]string
	if l := len(in.Details); l > 0 {
		details = make(map[string]string, l)
		for _, d := range in.Details {
			details[d.Key] = d.Value
		}
	}
	out.Details = details

	var responders []opsgenieResponder
	if l := len(in.Responders); l > 0 {
		responders = make([]opsgenieResponder, l)
		for _, r := range in.Responders {
			var responder opsgenieResponder = opsgenieResponder{
				ID:       r.ID,
				Name:     r.Name,
				Username: r.Username,
				Type:     r.Type,
			}
			responders = append(responders, responder)
		}
	}
	out.Responders = responders

	if in.HTTPConfig != nil {
		httpConfig, err := cg.convertHTTPConfig(ctx, *in.HTTPConfig, crKey)
		if err != nil {
			return nil, err
		}
		out.HTTPConfig = httpConfig
	}

	return out, nil
}

func (cg *configGenerator) convertWeChatConfig(ctx context.Context, in monitoringv1alpha1.WeChatConfig, crKey types.NamespacedName) (*weChatConfig, error) {

	out := &weChatConfig{}

	if in.SendResolved != nil {
		out.VSendResolved = *in.SendResolved
	}

	if in.APISecret != nil {
		apiSecret, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.APISecret)
		if err != nil {
			return nil, errors.Errorf("failed to get secret %q", in.APISecret)
		}
		out.APISecret = apiSecret
	}

	if in.APIURL != nil {
		out.APIURL = *in.APIURL
	}

	if in.CorpID != nil {
		out.CorpID = *in.CorpID
	}

	if in.AgentID != nil {
		out.AgentID = *in.AgentID
	}

	if in.ToUser != nil {
		out.ToUser = *in.ToUser
	}

	if in.ToParty != nil {
		out.ToParty = *in.ToParty
	}

	if in.ToTag != nil {
		out.ToTag = *in.ToTag
	}

	if in.Message != nil {
		out.Message = *in.Message
	}

	if in.MessageType != nil {
		out.MessageType = *in.MessageType
	}

	if in.HTTPConfig != nil {
		httpConfig, err := cg.convertHTTPConfig(ctx, *in.HTTPConfig, crKey)
		if err != nil {
			return nil, err
		}
		out.HTTPConfig = httpConfig
	}

	return out, nil
}

func convertInhibitRule(in *monitoringv1alpha1.InhibitRule, crKey types.NamespacedName) *inhibitRule {
	sourceMatch := map[string]string{}
	sourceMatchRE := map[string]string{}
	for _, sm := range in.SourceMatch {
		if sm.Regex {
			sourceMatchRE[sm.Name] = sm.Value
		} else {
			sourceMatch[sm.Name] = sm.Value
		}
	}

	sourceMatch["namespace"] = crKey.Namespace
	delete(sourceMatchRE, "namespace")

	// Set to nil if empty so that it doesn't show up in resulting yaml
	if len(sourceMatchRE) == 0 {
		sourceMatchRE = nil
	}

	targetMatch := map[string]string{}
	targetMatchRE := map[string]string{}
	for _, tm := range in.TargetMatch {
		if tm.Regex {
			targetMatchRE[tm.Name] = tm.Value
		} else {
			targetMatch[tm.Name] = tm.Value
		}
	}

	targetMatch["namespace"] = crKey.Namespace
	delete(targetMatchRE, "namespace")

	// Set to nil if empty so that it doesn't show up in resulting yaml
	if len(targetMatchRE) == 0 {
		targetMatchRE = nil
	}

	equal := in.Equal
	if len(equal) == 0 {
		equal = nil
	}

	return &inhibitRule{
		SourceMatch:   sourceMatch,
		SourceMatchRE: sourceMatchRE,
		TargetMatch:   targetMatch,
		TargetMatchRE: targetMatchRE,
		Equal:         equal,
	}
}

func prefixReceiverName(receiverName string, crKey types.NamespacedName) string {
	if receiverName == "" {
		return ""
	}
	return crKey.Namespace + "-" + crKey.Name + "-" + receiverName
}

func (cg *configGenerator) convertHTTPConfig(ctx context.Context, in monitoringv1alpha1.HTTPConfig, crKey types.NamespacedName) (*httpClientConfig, error) {
	out := &httpClientConfig{}

	if in.ProxyURL != nil {
		out.ProxyURL = *in.ProxyURL
	}

	if in.BasicAuth != nil {
		username, err := cg.store.GetSecretKey(ctx, crKey.Namespace, in.BasicAuth.Username)
		if err != nil {
			return nil, errors.Errorf("failed to get BasicAuth username key %q from secret %q", in.BasicAuth.Username.Key, in.BasicAuth.Username.Name)
		}

		password, err := cg.store.GetSecretKey(ctx, crKey.Namespace, in.BasicAuth.Password)
		if err != nil {
			return nil, errors.Errorf("failed to get BasicAuth password key %q from secret %q", in.BasicAuth.Password.Key, in.BasicAuth.Password.Name)
		}

		if username != "" && password != "" {
			out.BasicAuth = &basicAuth{Username: username, Password: password}
		}
	}

	if in.TLSConfig != nil {
		out.TLSConfig = cg.convertTLSConfig(ctx, in.TLSConfig, crKey)
	}

	if in.BearerTokenSecret != nil {
		bearerToken, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.BearerTokenSecret)
		if err != nil {
			return nil, errors.Errorf("failed to get bearer token key %q from secret %q", in.BearerTokenSecret.Key, in.BearerTokenSecret.Name)
		}
		out.BearerToken = bearerToken
	}

	return out, nil
}

func (cg *configGenerator) convertTLSConfig(ctx context.Context, in *monitoringv1.SafeTLSConfig, crKey types.NamespacedName) commoncfg.TLSConfig {
	out := commoncfg.TLSConfig{
		ServerName:         in.ServerName,
		InsecureSkipVerify: in.InsecureSkipVerify,
	}

	if in.CA != (monitoringv1.SecretOrConfigMap{}) {
		out.CAFile = path.Join(tlsAssetsDir, assets.TLSAssetKeyFromSelector(crKey.Namespace, in.CA).String())
	}
	if in.Cert != (monitoringv1.SecretOrConfigMap{}) {
		out.CertFile = path.Join(tlsAssetsDir, assets.TLSAssetKeyFromSelector(crKey.Namespace, in.Cert).String())
	}
	if in.KeySecret != nil {
		out.KeyFile = path.Join(tlsAssetsDir, assets.TLSAssetKeyFromSecretSelector(crKey.Namespace, in.KeySecret).String())
	}

	return out
}
