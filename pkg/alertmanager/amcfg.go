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
	"net"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus/alertmanager/config"
	commoncfg "github.com/prometheus/common/config"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"
)

func loadCfg(s string) (*alertmanagerConfig, error) {
	// Run upstream Load function to get any validation checks that it runs.
	_, err := config.Load(s)
	if err != nil {
		return nil, err
	}

	cfg := &alertmanagerConfig{}
	err = yaml.UnmarshalStrict([]byte(s), cfg)
	if err != nil {
		return nil, err
	}

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
	// same order.
	amConfigIdentifiers := make([]string, len(amConfigs))
	i := 0
	for k := range amConfigs {
		amConfigIdentifiers[i] = k
		i++
	}
	sort.Strings(amConfigIdentifiers)

	subRoutes := make([]*route, 0, len(amConfigs))
	for _, amConfigIdentifier := range amConfigIdentifiers {
		crKey := types.NamespacedName{
			Name:      amConfigs[amConfigIdentifier].Name,
			Namespace: amConfigs[amConfigIdentifier].Namespace,
		}

		// Add inhibitRules to baseConfig.InhibitRules.
		for _, inhibitRule := range amConfigs[amConfigIdentifier].Spec.InhibitRules {
			baseConfig.InhibitRules = append(baseConfig.InhibitRules, convertInhibitRule(&inhibitRule, crKey))
		}

		// Skip early if there's no route definition.
		if amConfigs[amConfigIdentifier].Spec.Route == nil {
			continue
		}

		subRoutes = append(subRoutes, convertRoute(amConfigs[amConfigIdentifier].Spec.Route, crKey, true))

		for _, receiver := range amConfigs[amConfigIdentifier].Spec.Receivers {
			receivers, err := cg.convertReceiver(ctx, &receiver, crKey)
			if err != nil {
				return nil, errors.Wrapf(err, "AlertmanagerConfig %s", crKey.String())
			}
			baseConfig.Receivers = append(baseConfig.Receivers, receivers)
		}
	}

	// For alerts to be processed by the AlertmanagerConfig routes, they need
	// to appear before the routes defined in the main configuration.
	// Because all first-level AlertmanagerConfig routes have "continue: true",
	// alerts will fallthrough.
	baseConfig.Route.Routes = append(subRoutes, baseConfig.Route.Routes...)

	return yaml.Marshal(baseConfig)
}

func convertRoute(in *monitoringv1alpha1.Route, crKey types.NamespacedName, firstLevelRoute bool) *route {
	// Enforce "continue" to be true for the top-level route.
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

	// Set to nil if empty so that it doesn't show up in the resulting yaml.
	if len(match) == 0 {
		match = nil
	}
	// Set to nil if empty so that it doesn't show up in the resulting yaml.
	if len(matchRE) == 0 {
		matchRE = nil
	}

	var routes []*route
	if len(in.Routes) > 0 {
		routes = make([]*route, len(in.Routes))
		children, err := in.ChildRoutes()
		if err != nil {
			// The controller should already have checked that ChildRoutes()
			// doesn't return an error when selecting AlertmanagerConfig CRDs.
			// If there's an error here, we have a serious bug in the code.
			panic(err)
		}
		for i := range children {
			routes[i] = convertRoute(&children[i], crKey, false)
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

	var slackConfigs []*slackConfig
	if l := len(in.SlackConfigs); l > 0 {
		slackConfigs = make([]*slackConfig, l)
		for i := range in.SlackConfigs {
			receiver, err := cg.convertSlackConfig(ctx, in.SlackConfigs[i], crKey)
			if err != nil {
				return nil, errors.Wrapf(err, "SlackConfig[%d]", i)
			}
			slackConfigs[i] = receiver
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

	var emailConfigs []*emailConfig
	if l := len(in.EmailConfigs); l > 0 {
		emailConfigs = make([]*emailConfig, l)
		for i := range in.EmailConfigs {
			receiver, err := cg.convertEmailConfig(ctx, in.EmailConfigs[i], crKey)
			if err != nil {
				return nil, errors.Wrapf(err, "EmailConfig[%d]", i)
			}
			emailConfigs[i] = receiver
		}
	}

	var victorOpsConfigs []*victorOpsConfig
	if l := len(in.VictorOpsConfigs); l > 0 {
		victorOpsConfigs = make([]*victorOpsConfig, l)
		for i := range in.VictorOpsConfigs {
			receiver, err := cg.convertVictorOpsConfig(ctx, in.VictorOpsConfigs[i], crKey)
			if err != nil {
				return nil, errors.Wrapf(err, "VictorOpsConfig[%d]", i)
			}
			victorOpsConfigs[i] = receiver
		}
	}

	var pushoverConfigs []*pushoverConfig
	if l := len(in.PushoverConfigs); l > 0 {
		pushoverConfigs = make([]*pushoverConfig, l)
		for i := range in.PushoverConfigs {
			receiver, err := cg.convertPushoverConfig(ctx, in.PushoverConfigs[i], crKey)
			if err != nil {
				return nil, errors.Wrapf(err, "PushoverConfig[%d]", i)
			}
			pushoverConfigs[i] = receiver
		}
	}

	return &receiver{
		Name:             prefixReceiverName(in.Name, crKey),
		OpsgenieConfigs:  opsgenieConfigs,
		PagerdutyConfigs: pagerdutyConfigs,
		SlackConfigs:     slackConfigs,
		WebhookConfigs:   webhookConfigs,
		WeChatConfigs:    weChatConfigs,
		EmailConfigs:     emailConfigs,
		VictorOpsConfigs: victorOpsConfigs,
		PushoverConfigs:  pushoverConfigs,
	}, nil
}

func (cg *configGenerator) convertWebhookConfig(ctx context.Context, in monitoringv1alpha1.WebhookConfig, crKey types.NamespacedName) (*webhookConfig, error) {
	out := &webhookConfig{
		VSendResolved: in.SendResolved,
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

	if in.MaxAlerts > 0 {
		out.MaxAlerts = in.MaxAlerts
	}

	return out, nil
}

func (cg *configGenerator) convertSlackConfig(ctx context.Context, in monitoringv1alpha1.SlackConfig, crKey types.NamespacedName) (*slackConfig, error) {
	out := &slackConfig{
		VSendResolved: in.SendResolved,
		Channel:       in.Channel,
		Username:      in.Username,
		Color:         in.Color,
		Title:         in.Title,
		TitleLink:     in.TitleLink,
		Pretext:       in.Pretext,
		Text:          in.Text,
		ShortFields:   in.ShortFields,
		Footer:        in.Footer,
		Fallback:      in.Fallback,
		CallbackID:    in.CallbackID,
		IconEmoji:     in.IconEmoji,
		IconURL:       in.IconURL,
		ImageURL:      in.ImageURL,
		ThumbURL:      in.ThumbURL,
		LinkNames:     in.LinkNames,
		MrkdwnIn:      in.MrkdwnIn,
	}

	if in.APIURL != nil {
		url, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.APIURL)
		if err != nil {
			return nil, errors.Errorf("failed to get key %q from secret %q", in.APIURL.Key, in.APIURL.Name)
		}
		out.APIURL = url
	}

	var actions []slackAction
	if l := len(in.Actions); l > 0 {
		actions = make([]slackAction, l)
		for i, a := range in.Actions {
			var action slackAction = slackAction{
				Type:  a.Type,
				Text:  a.Text,
				URL:   a.URL,
				Style: a.Style,
				Name:  a.Name,
				Value: a.Value,
			}

			if a.ConfirmField != nil {

				action.ConfirmField = &slackConfirmationField{
					Text:        a.ConfirmField.Text,
					Title:       a.ConfirmField.Title,
					OkText:      a.ConfirmField.OkText,
					DismissText: a.ConfirmField.DismissText,
				}
			}

			actions[i] = action
		}
		out.Actions = actions
	}

	if l := len(in.Fields); l > 0 {
		var fields []slackField = make([]slackField, l)
		for i, f := range in.Fields {
			var field slackField = slackField{
				Title: f.Title,
				Value: f.Value,
			}

			if f.Short != nil {
				field.Short = *f.Short
			}
			fields[i] = field
		}
		out.Fields = fields
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

func (cg *configGenerator) convertPagerdutyConfig(ctx context.Context, in monitoringv1alpha1.PagerDutyConfig, crKey types.NamespacedName) (*pagerdutyConfig, error) {
	out := &pagerdutyConfig{
		VSendResolved: in.SendResolved,
		Class:         in.Class,
		Client:        in.Client,
		ClientURL:     in.ClientURL,
		Component:     in.Component,
		Description:   in.Description,
		Group:         in.Group,
		Severity:      in.Severity,
		URL:           in.URL,
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
	out := &opsgenieConfig{
		VSendResolved: in.SendResolved,
		APIURL:        in.APIURL,
		Message:       in.Message,
		Description:   in.Description,
		Source:        in.Source,
		Tags:          in.Tags,
		Note:          in.Note,
		Priority:      in.Priority,
	}

	if in.APIKey != nil {
		apiKey, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.APIKey)
		if err != nil {
			return nil, errors.Errorf("failed to get api key %q from secret %q", in.APIKey.Key, in.APIKey.Name)
		}
		out.APIKey = apiKey
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
		responders = make([]opsgenieResponder, 0, l)
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

	out := &weChatConfig{
		VSendResolved: in.SendResolved,
		APIURL:        in.APIURL,
		CorpID:        in.CorpID,
		AgentID:       in.AgentID,
		ToUser:        in.ToUser,
		ToParty:       in.ToParty,
		ToTag:         in.ToTag,
		Message:       in.Message,
		MessageType:   in.MessageType,
	}

	if in.APISecret != nil {
		apiSecret, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.APISecret)
		if err != nil {
			return nil, errors.Errorf("failed to get secret %q", in.APISecret)
		}
		out.APISecret = apiSecret
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

func (cg *configGenerator) convertEmailConfig(ctx context.Context, in monitoringv1alpha1.EmailConfig, crKey types.NamespacedName) (*emailConfig, error) {
	out := &emailConfig{
		VSendResolved: in.SendResolved,
		To:            in.To,
		From:          in.From,
		Hello:         in.Hello,
		AuthUsername:  in.AuthUsername,
		AuthIdentity:  in.AuthIdentity,
		HTML:          in.HTML,
		Text:          in.Text,
		RequireTLS:    in.RequireTLS,
	}

	if in.To == "" {
		return nil, errors.New("missing to address in email config")
	}

	if in.Smarthost != "" {
		host, port, err := net.SplitHostPort(in.Smarthost)
		if err != nil {
			return nil, errors.New("failed to extract host and port from Smarthost")
		}
		out.Smarthost.Host = host
		out.Smarthost.Port = port
	}

	if in.AuthPassword != nil {
		authPassword, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.AuthPassword)
		if err != nil {
			return nil, errors.Errorf("failed to get secret %q", in.AuthPassword)
		}
		out.AuthPassword = authPassword
	}

	if in.AuthSecret != nil {
		authSecret, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.AuthSecret)
		if err != nil {
			return nil, errors.Errorf("failed to get secret %q", in.AuthSecret)
		}
		out.AuthSecret = authSecret
	}

	var headers map[string]string
	if l := len(in.Headers); l > 0 {
		headers = make(map[string]string, l)

		var key string
		for _, d := range in.Headers {
			key = strings.Title(d.Key)
			if _, ok := headers[key]; ok {
				return nil, errors.Errorf("duplicate header %q in email config", key)
			}
			headers[key] = d.Value
		}
	}
	out.Headers = headers

	if in.TLSConfig != nil {
		out.TLSConfig = cg.convertTLSConfig(ctx, in.TLSConfig, crKey)
	}

	return out, nil
}

func (cg *configGenerator) convertVictorOpsConfig(ctx context.Context, in monitoringv1alpha1.VictorOpsConfig, crKey types.NamespacedName) (*victorOpsConfig, error) {
	out := &victorOpsConfig{
		VSendResolved:     in.SendResolved,
		APIURL:            in.APIURL,
		RoutingKey:        in.RoutingKey,
		MessageType:       in.MessageType,
		EntityDisplayName: in.EntityDisplayName,
		StateMessage:      in.StateMessage,
		MonitoringTool:    in.MonitoringTool,
	}

	if in.APIKey != nil {
		apiKey, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.APIKey)
		if err != nil {
			return nil, errors.Errorf("failed to get secret %q", in.APIKey)
		}
		out.APIKey = apiKey
	}
	if in.RoutingKey == "" {
		return nil, errors.New("missing Routing key in VictorOps config")
	}

	var customFields map[string]string
	if l := len(in.CustomFields); l > 0 {
		// from https://github.com/prometheus/alertmanager/blob/a7f9fdadbecbb7e692d2cd8d3334e3d6de1602e1/config/notifiers.go#L497
		reservedFields := map[string]struct{}{
			"routing_key":         {},
			"message_type":        {},
			"state_message":       {},
			"entity_display_name": {},
			"monitoring_tool":     {},
			"entity_id":           {},
			"entity_state":        {},
		}
		customFields = make(map[string]string, l)
		for _, d := range in.CustomFields {
			if _, ok := reservedFields[d.Key]; ok {
				return nil, errors.Errorf("VictorOps config contains custom field %s which cannot be used as it conflicts with the fixed/static fields", d.Key)
			}
			customFields[d.Key] = d.Value
		}
	}
	out.CustomFields = customFields

	if in.HTTPConfig != nil {
		httpConfig, err := cg.convertHTTPConfig(ctx, *in.HTTPConfig, crKey)
		if err != nil {
			return nil, err
		}
		out.HTTPConfig = httpConfig
	}
	return out, nil
}

func (cg *configGenerator) convertPushoverConfig(ctx context.Context, in monitoringv1alpha1.PushoverConfig, crKey types.NamespacedName) (*pushoverConfig, error) {
	out := &pushoverConfig{
		VSendResolved: in.SendResolved,
		Title:         in.Title,
		Message:       in.Message,
		URL:           in.URL,
		URLTitle:      in.URLTitle,
		Priority:      in.Priority,
		HTML:          in.HTML,
	}

	{
		if in.UserKey == nil {
			return nil, errors.Errorf("mandatory field %q is empty", "userKey")
		}
		userKey, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.UserKey)
		if err != nil {
			return nil, errors.Errorf("failed to get secret %q", in.UserKey)
		}
		if userKey == "" {
			return nil, errors.Errorf("mandatory field %q is empty", "userKey")
		}
		out.UserKey = userKey
	}

	{
		if in.Token == nil {
			return nil, errors.Errorf("mandatory field %q is empty", "token")
		}
		token, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.Token)
		if err != nil {
			return nil, errors.Errorf("failed to get secret %q", in.Token)
		}
		if token == "" {
			return nil, errors.Errorf("mandatory field %q is empty", "token")
		}
		out.Token = token
	}

	if in.Retry != "" {
		retry, err := time.ParseDuration(in.Retry)
		if err != nil {
			return nil, errors.Errorf("failed to parse Retry duration: %s", err)
		}
		out.Retry = duration(retry)
	}

	if in.Expire != "" {
		expire, err := time.ParseDuration(in.Expire)
		if err != nil {
			return nil, errors.Errorf("failed to parse Expire duration: %s", err)
		}
		out.Expire = duration(expire)
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
	out := &httpClientConfig{
		ProxyURL: in.ProxyURL,
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
