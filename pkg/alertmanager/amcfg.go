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

	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/timeinterval"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"
)

const inhibitRuleNamespaceKey = "namespace"

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
	logger    log.Logger
	amVersion semver.Version
	store     *assets.Store
}

func newConfigGenerator(logger log.Logger, amVersion semver.Version, store *assets.Store) *configGenerator {
	cg := &configGenerator{
		logger:    logger,
		amVersion: amVersion,
		store:     store,
	}
	return cg
}

// validateConfigInputs runs extra validation on the AlertManager fields which can't be done at the CRD schema validation level.
func validateConfigInputs(am *monitoringv1.Alertmanager) error {
	if am.Spec.Retention != "" {
		if err := operator.ValidateDurationField(am.Spec.Retention); err != nil {
			return errors.Wrap(err, "invalid retention value specified")
		}
	}
	if am.Spec.ClusterGossipInterval != "" {
		if err := operator.ValidateDurationField(am.Spec.ClusterGossipInterval); err != nil {
			return errors.Wrap(err, "invalid clusterGossipInterval value specified")
		}
	}

	if am.Spec.ClusterPushpullInterval != "" {
		if err := operator.ValidateDurationField(am.Spec.ClusterPushpullInterval); err != nil {
			return errors.Wrap(err, "invalid clusterPushpullInterval value specified")
		}
	}

	if am.Spec.ClusterPeerTimeout != "" {
		if err := operator.ValidateDurationField(am.Spec.ClusterPeerTimeout); err != nil {
			return errors.Wrap(err, "invalid clusterPeerTimeout value specified")
		}
	}
	return nil
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
			baseConfig.InhibitRules = append(baseConfig.InhibitRules, cg.convertInhibitRule(&inhibitRule, crKey))
		}

		// Skip early if there's no route definition.
		if amConfigs[amConfigIdentifier].Spec.Route == nil {
			continue
		}

		subRoutes = append(subRoutes,
			cg.enforceNamespaceForRoute(
				cg.convertRoute(
					amConfigs[amConfigIdentifier].Spec.Route, crKey),
				amConfigs[amConfigIdentifier].Namespace,
			),
		)

		for _, receiver := range amConfigs[amConfigIdentifier].Spec.Receivers {
			receivers, err := cg.convertReceiver(ctx, &receiver, crKey)
			if err != nil {
				return nil, errors.Wrapf(err, "AlertmanagerConfig %s", crKey.String())
			}
			baseConfig.Receivers = append(baseConfig.Receivers, receivers)
		}

		for _, muteTimeInterval := range amConfigs[amConfigIdentifier].Spec.MuteTimeIntervals {
			mti, err := convertMuteTimeInterval(&muteTimeInterval, crKey)
			if err != nil {
				return nil, errors.Wrapf(err, "AlertmanagerConfig %s", crKey.String())
			}
			baseConfig.MuteTimeIntervals = append(baseConfig.MuteTimeIntervals, mti)
		}
	}

	// For alerts to be processed by the AlertmanagerConfig routes, they need
	// to appear before the routes defined in the main configuration.
	// Because all first-level AlertmanagerConfig routes have "continue: true",
	// alerts will fall through.
	baseConfig.Route.Routes = append(subRoutes, baseConfig.Route.Routes...)

	generatedConf := &baseConfig
	if err := generatedConf.sanitize(cg.amVersion, cg.logger); err != nil {
		return nil, err
	}
	return yaml.Marshal(generatedConf)
}

// enforceNamespaceForRoute modifies the route configuration to match alerts
// originating only from the given namespace.
func (cg *configGenerator) enforceNamespaceForRoute(r *route, namespace string) *route {
	matchersV2Allowed := cg.amVersion.GTE(semver.MustParse("0.22.0"))
	// Routes created from AlertmanagerConfig resources should only match
	// alerts that come from the same namespace.
	if matchersV2Allowed {
		r.Matchers = append(r.Matchers, monitoringv1alpha1.Matcher{
			Name:      "namespace",
			Value:     namespace,
			MatchType: monitoringv1alpha1.MatchEqual,
		}.String())
	} else {
		r.Match["namespace"] = namespace
	}

	// Alerts should still be evaluated by the following routes.
	r.Continue = true

	return r
}

func (cg *configGenerator) convertRoute(in *monitoringv1alpha1.Route, crKey types.NamespacedName) *route {
	var matchers []string

	// deprecated
	match := map[string]string{}
	matchRE := map[string]string{}

	for _, matcher := range in.Matchers {
		// prefer matchers to deprecated config
		if matcher.MatchType != "" {
			matchers = append(matchers, matcher.String())
			continue
		}

		if matcher.Regex {
			matchRE[matcher.Name] = matcher.Value
		} else {
			match[matcher.Name] = matcher.Value
		}
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
			routes[i] = cg.convertRoute(&children[i], crKey)
		}
	}

	receiver := makeNamespacedString(in.Receiver, crKey)

	var prefixedMuteTimeIntervals []string
	if len(in.MuteTimeIntervals) > 0 {
		for _, mti := range in.MuteTimeIntervals {
			prefixedMuteTimeIntervals = append(prefixedMuteTimeIntervals, makeNamespacedString(mti, crKey))
		}
	}

	return &route{
		Receiver:          receiver,
		GroupByStr:        in.GroupBy,
		GroupWait:         in.GroupWait,
		GroupInterval:     in.GroupInterval,
		RepeatInterval:    in.RepeatInterval,
		Continue:          in.Continue,
		Match:             match,
		MatchRE:           matchRE,
		Matchers:          matchers,
		Routes:            routes,
		MuteTimeIntervals: prefixedMuteTimeIntervals,
	}
}

// convertReceiver converts a monitoringv1alpha1.Receiver to an alertmanager.receiver
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
		Name:             makeNamespacedString(in.Name, crKey),
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
		out.URL = strings.TrimSpace(url)
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
		out.APIURL = strings.TrimSpace(url)
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

	if in.Smarthost != "" {
		out.Smarthost.Host, out.Smarthost.Port, _ = net.SplitHostPort(in.Smarthost)
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

	if l := len(in.Headers); l > 0 {
		headers := make(map[string]string, l)
		for _, d := range in.Headers {
			headers[strings.Title(d.Key)] = d.Value
		}
		out.Headers = headers
	}

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
		token, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.Token)
		if err != nil {
			return nil, errors.Errorf("failed to get secret %q", in.Token)
		}
		if token == "" {
			return nil, errors.Errorf("mandatory field %q is empty", "token")
		}
		out.Token = token
	}

	{
		if in.Retry != "" {
			retry, _ := time.ParseDuration(in.Retry)
			out.Retry = duration(retry)
		}

		if in.Expire != "" {
			expire, _ := time.ParseDuration(in.Expire)
			out.Expire = duration(expire)
		}
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

func (cg *configGenerator) convertInhibitRule(in *monitoringv1alpha1.InhibitRule, crKey types.NamespacedName) *inhibitRule {
	matchersV2Allowed := cg.amVersion.GTE(semver.MustParse("0.22.0"))
	var sourceMatchers []string
	var targetMatchers []string

	v2NamespaceMatcher := monitoringv1alpha1.Matcher{
		Name:      inhibitRuleNamespaceKey,
		Value:     crKey.Namespace,
		MatchType: monitoringv1alpha1.MatchEqual,
	}.String()

	// todo (pgough) the following config are deprecated and can be removed when
	// support matrix has reached >= 0.22.0
	sourceMatch := map[string]string{}
	sourceMatchRE := map[string]string{}
	targetMatch := map[string]string{}
	targetMatchRE := map[string]string{}

	for _, sm := range in.SourceMatch {
		// prefer matchers to deprecated syntax
		if sm.MatchType != "" {
			sourceMatchers = append(sourceMatchers, sm.String())
			continue
		}

		if matchersV2Allowed {
			if sm.Regex {
				sourceMatchers = append(sourceMatchers, inhibitRuleRegexToV2(sm.Name, sm.Value))
			} else {
				sourceMatchers = append(sourceMatchers, inhibitRuleToV2(sm.Name, sm.Value))
			}
			continue
		}

		if sm.Regex {
			sourceMatchRE[sm.Name] = sm.Value
		} else {
			sourceMatch[sm.Name] = sm.Value
		}
	}

	delete(sourceMatchRE, inhibitRuleNamespaceKey)
	if matchersV2Allowed {
		if !contains(v2NamespaceMatcher, sourceMatchers) {
			sourceMatchers = append(sourceMatchers, v2NamespaceMatcher)
		}
		delete(sourceMatch, inhibitRuleNamespaceKey)
	} else {
		sourceMatch[inhibitRuleNamespaceKey] = crKey.Namespace
	}

	for _, tm := range in.TargetMatch {
		// prefer matchers to deprecated config
		if tm.MatchType != "" {
			targetMatchers = append(targetMatchers, tm.String())
			continue
		}

		if matchersV2Allowed {
			if tm.Regex {
				targetMatchers = append(targetMatchers, inhibitRuleRegexToV2(tm.Name, tm.Value))
			} else {
				targetMatchers = append(targetMatchers, inhibitRuleToV2(tm.Name, tm.Value))
			}
			continue
		}

		if tm.Regex {
			targetMatchRE[tm.Name] = tm.Value
		} else {
			targetMatch[tm.Name] = tm.Value
		}
	}

	delete(targetMatchRE, inhibitRuleNamespaceKey)
	if matchersV2Allowed {
		if !contains(v2NamespaceMatcher, targetMatchers) {
			targetMatchers = append(targetMatchers, v2NamespaceMatcher)
		}
		delete(targetMatch, inhibitRuleNamespaceKey)
	} else {
		targetMatch[inhibitRuleNamespaceKey] = crKey.Namespace
	}

	return &inhibitRule{
		SourceMatch:    sourceMatch,
		SourceMatchRE:  sourceMatchRE,
		SourceMatchers: sourceMatchers,
		TargetMatch:    targetMatch,
		TargetMatchRE:  targetMatchRE,
		TargetMatchers: targetMatchers,
		Equal:          in.Equal,
	}
}

func convertMuteTimeInterval(in *monitoringv1alpha1.MuteTimeInterval, crKey types.NamespacedName) (*muteTimeInterval, error) {
	muteTimeInterval := &muteTimeInterval{}

	for _, timeInterval := range in.TimeIntervals {
		ti := timeinterval.TimeInterval{}

		for _, time := range timeInterval.Times {
			parsedTime, err := time.Parse()
			if err != nil {
				return nil, err
			}
			ti.Times = append(ti.Times, timeinterval.TimeRange{
				StartMinute: parsedTime.Start,
				EndMinute:   parsedTime.End,
			})
		}

		for _, wd := range timeInterval.Weekdays {
			parsedWeekday, err := wd.Parse()
			if err != nil {
				return nil, err
			}
			ti.Weekdays = append(ti.Weekdays, timeinterval.WeekdayRange{
				InclusiveRange: timeinterval.InclusiveRange{
					Begin: parsedWeekday.Start,
					End:   parsedWeekday.End,
				},
			})
		}

		for _, dom := range timeInterval.DaysOfMonth {
			ti.DaysOfMonth = append(ti.DaysOfMonth, timeinterval.DayOfMonthRange{
				InclusiveRange: timeinterval.InclusiveRange{
					Begin: dom.Start,
					End:   dom.End,
				},
			})
		}

		for _, month := range timeInterval.Months {
			parsedMonth, err := month.Parse()
			if err != nil {
				return nil, err
			}
			ti.Months = append(ti.Months, timeinterval.MonthRange{
				InclusiveRange: timeinterval.InclusiveRange{
					Begin: parsedMonth.Start,
					End:   parsedMonth.End,
				},
			})
		}

		for _, year := range timeInterval.Years {
			parsedYear, err := year.Parse()
			if err != nil {
				return nil, err
			}
			ti.Years = append(ti.Years, timeinterval.YearRange{
				InclusiveRange: timeinterval.InclusiveRange{
					Begin: parsedYear.Start,
					End:   parsedYear.End,
				},
			})
		}

		muteTimeInterval.Name = makeNamespacedString(in.Name, crKey)
		muteTimeInterval.TimeIntervals = append(muteTimeInterval.TimeIntervals, ti)
	}

	return muteTimeInterval, nil
}

func makeNamespacedString(in string, crKey types.NamespacedName) string {
	if in == "" {
		return ""
	}
	return crKey.Namespace + "-" + crKey.Name + "-" + in
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

		if username != "" || password != "" {
			out.BasicAuth = &basicAuth{Username: username, Password: password}
		}
	}

	if in.Authorization != nil {
		credentials, err := cg.store.GetSecretKey(ctx, crKey.Namespace, *in.Authorization.Credentials)
		if err != nil {
			return nil, errors.Errorf("failed to get Authorization credentials key %q from secret %q", in.Authorization.Credentials.Key, in.Authorization.Credentials.Name)
		}

		if credentials != "" {
			authorizationType := in.Authorization.Type
			if authorizationType == "" {
				authorizationType = "Bearer"
			}
			out.Authorization = &authorization{Type: authorizationType, Credentials: credentials}
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

func (cg *configGenerator) convertTLSConfig(ctx context.Context, in *monitoringv1.SafeTLSConfig, crKey types.NamespacedName) tlsConfig {
	out := tlsConfig{
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

// sanitize the config against a specific AlertManager version
// types may be sanitized in one of two ways:
// 1. stripping the unsupported config and log a warning
// 2. error which ensures that config will not be reconciled - this will be logged by a calling function
func (c *alertmanagerConfig) sanitize(amVersion semver.Version, logger log.Logger) error {
	if c == nil {
		return nil
	}

	c.Global.sanitize(amVersion, logger)

	for _, receiver := range c.Receivers {
		receiver.sanitize(amVersion, logger)
	}

	for i, rule := range c.InhibitRules {
		if err := rule.sanitize(amVersion, logger); err != nil {
			return errors.Wrapf(err, "inhibit_rules[%d]", i)
		}
	}

	if len(c.MuteTimeIntervals) > 0 && !amVersion.GTE(semver.MustParse("0.22.0")) {
		// mute time intervals are unsupported < 0.22.0, and we already log the situation
		// when handling the routes so just set to nil
		c.MuteTimeIntervals = nil
	}

	return c.Route.sanitize(amVersion, logger)
}

// sanitize globalConfig
func (gc *globalConfig) sanitize(amVersion semver.Version, logger log.Logger) {
	if gc == nil {
		return
	}

	if gc.HTTPConfig != nil {
		gc.HTTPConfig.sanitize(amVersion, logger)
	}

	// We need to sanitize the config for slack globally
	// As of v0.22.0 AlertManager config supports passing URL via file name
	fileURLAllowed := amVersion.GTE(semver.MustParse("0.22.0"))
	if gc.SlackAPIURLFile != "" {

		if gc.SlackAPIURL != nil {
			msg := "'slack_api_url' and 'slack_api_url_file' are mutually exclusive - 'slack_api_url' has taken precedence"
			level.Warn(logger).Log("msg", msg)
			gc.SlackAPIURLFile = ""
		}

		if !fileURLAllowed {
			msg := "'slack_api_url_file' supported in AlertManager >= 0.22.0 only - dropping field from provided config"
			level.Warn(logger).Log("msg", msg, "current_version", amVersion.String())
			gc.SlackAPIURLFile = ""
		}
	}
}

// sanitize httpClientConfig
func (hc *httpClientConfig) sanitize(amVersion semver.Version, logger log.Logger) {
	if hc == nil {
		return
	}
	// we don't need to do any sanitization in this case and return early
	if hc.Authorization == nil {
		return
	}

	if hc.BasicAuth != nil {
		msg := "'basicAuth' and 'authorization' are mutually exclusive, 'basicAuth' has taken precedence"
		level.Warn(logger).Log("msg", msg)
		hc.Authorization = nil
	}
	// we could have returned here but useful to grab the log and bubble up the warning
	if httpAuthzAllowed := amVersion.GTE(semver.MustParse("0.22.0")); !httpAuthzAllowed {
		msg := "'authorization' set in 'http_config' but  supported in AlertManager >= 0.22.0 only - dropping field from provided config"
		level.Warn(logger).Log("msg", msg, "current_version", amVersion.String())
		hc.Authorization = nil
	}
}

// sanitize the receiver
func (r *receiver) sanitize(amVersion semver.Version, logger log.Logger) {
	if r == nil {
		return
	}
	withLogger := log.With(logger, "receiver", r.Name)

	for _, conf := range r.OpsgenieConfigs {
		conf.sanitize(amVersion, withLogger)
	}

	for _, conf := range r.PagerdutyConfigs {
		conf.sanitize(amVersion, withLogger)
	}

	for _, conf := range r.PagerdutyConfigs {
		conf.sanitize(amVersion, withLogger)
	}

	for _, conf := range r.PushoverConfigs {
		conf.sanitize(amVersion, withLogger)
	}

	for _, conf := range r.SlackConfigs {
		conf.sanitize(amVersion, withLogger)
	}

	for _, conf := range r.VictorOpsConfigs {
		conf.sanitize(amVersion, withLogger)
	}

	for _, conf := range r.WebhookConfigs {
		conf.sanitize(amVersion, withLogger)
	}

	for _, conf := range r.WeChatConfigs {
		conf.sanitize(amVersion, withLogger)
	}
}

func (ogc *opsgenieConfig) sanitize(amVersion semver.Version, logger log.Logger) {
	ogc.HTTPConfig.sanitize(amVersion, logger)
}

func (pdc *pagerdutyConfig) sanitize(amVersion semver.Version, logger log.Logger) {
	pdc.HTTPConfig.sanitize(amVersion, logger)
}

func (poc *pushoverConfig) sanitize(amVersion semver.Version, logger log.Logger) {
	poc.HTTPConfig.sanitize(amVersion, logger)
}

func (sc *slackConfig) sanitize(amVersion semver.Version, logger log.Logger) {
	sc.HTTPConfig.sanitize(amVersion, logger)

	if sc.APIURLFile == "" {
		return
	}
	// We need to sanitize the config for slack receivers
	// As of v0.22.0 AlertManager config supports passing URL via file name
	fileURLAllowed := amVersion.GTE(semver.MustParse("0.22.0"))
	if sc.APIURL != "" {
		msg := "'api_url' and 'api_url_file' are mutually exclusive for slack receiver config - 'api_url' has taken precedence"
		level.Warn(logger).Log("msg", msg)
		sc.APIURLFile = ""
	}

	if !fileURLAllowed {
		msg := "'api_url_file' supported in AlertManager >= 0.22.0 only - dropping field from provided config"
		level.Warn(logger).Log("msg", msg, "current_version", amVersion.String())
		sc.APIURLFile = ""
	}
}

func (voc *victorOpsConfig) sanitize(amVersion semver.Version, logger log.Logger) {
	voc.HTTPConfig.sanitize(amVersion, logger)
}

func (whc *webhookConfig) sanitize(amVersion semver.Version, logger log.Logger) {
	whc.HTTPConfig.sanitize(amVersion, logger)
}

func (wcc *weChatConfig) sanitize(amVersion semver.Version, logger log.Logger) {
	wcc.HTTPConfig.sanitize(amVersion, logger)
}

func (ir *inhibitRule) sanitize(amVersion semver.Version, logger log.Logger) error {
	matchersV2Allowed := amVersion.GTE(semver.MustParse("0.22.0"))

	if !matchersV2Allowed {
		// check if rule has provided invalid syntax and error if true
		if checkNotEmptyStrSlice(ir.SourceMatchers, ir.TargetMatchers) {
			msg := fmt.Sprintf(`target_matchers and source_matchers matching is supported in Alertmanager >= 0.22.0 only (target_matchers=%v, source_matchers=%v)`, ir.TargetMatchers, ir.SourceMatchers)
			return errors.New(msg)
		}
		return nil
	}

	// we log a warning if the rule continues to use deprecated values in addition
	// to the namespace label we have injected - but we won't convert these
	if checkNotEmptyMap(ir.SourceMatch, ir.TargetMatch, ir.SourceMatchRE, ir.TargetMatchRE) {
		msg := "inhibit rule is using a deprecated match syntax which will be removed in future versions"
		level.Warn(logger).Log("msg", msg, "source_match", ir.SourceMatch, "target_match", ir.TargetMatch, "source_match_re", ir.SourceMatchRE, "target_match_re", ir.TargetMatchRE)
	}

	// ensure empty data structures are assigned nil so their yaml output is sanitized
	ir.TargetMatch = convertMapToNilIfEmpty(ir.TargetMatch)
	ir.TargetMatchRE = convertMapToNilIfEmpty(ir.TargetMatchRE)
	ir.SourceMatch = convertMapToNilIfEmpty(ir.SourceMatch)
	ir.SourceMatchRE = convertMapToNilIfEmpty(ir.SourceMatchRE)
	ir.TargetMatchers = convertSliceToNilIfEmpty(ir.TargetMatchers)
	ir.SourceMatchers = convertSliceToNilIfEmpty(ir.SourceMatchers)
	ir.Equal = convertSliceToNilIfEmpty(ir.Equal)

	return nil
}

// sanitize a route and all its child routes.
// Warns if the config is using deprecated syntax against a later version.
// Returns an error if the config could potentially break routing logic
func (r *route) sanitize(amVersion semver.Version, logger log.Logger) error {
	if r == nil {
		return nil
	}

	matchersV2Allowed := amVersion.GTE(semver.MustParse("0.22.0"))
	muteTimeIntervalsAllowed := matchersV2Allowed
	withLogger := log.With(logger, "receiver", r.Receiver)

	if !matchersV2Allowed && checkNotEmptyStrSlice(r.Matchers) {
		return fmt.Errorf(`invalid syntax in route config for 'matchers' comparison based matching is supported in Alertmanager >= 0.22.0 only (matchers=%v)`, r.Matchers)
	}

	if matchersV2Allowed && checkNotEmptyMap(r.Match, r.MatchRE) {
		msg := "'matchers' field is using a deprecated syntax which will be removed in future versions"
		level.Warn(withLogger).Log("msg", msg, "match", r.Match, "match_re", r.MatchRE)
	}

	if !muteTimeIntervalsAllowed {
		msg := "named mute time intervals in route is supported in Alertmanager >= 0.22.0 only - dropping config"
		level.Warn(withLogger).Log("msg", msg, "mute_time_intervals", r.MuteTimeIntervals)
		r.MuteTimeIntervals = nil
	}

	for i, child := range r.Routes {
		if err := child.sanitize(amVersion, logger); err != nil {
			return errors.Wrapf(err, "route[%d]", i)
		}
	}
	// Set to nil if empty so that it doesn't show up in the resulting yaml.
	r.Match = convertMapToNilIfEmpty(r.Match)
	r.MatchRE = convertMapToNilIfEmpty(r.MatchRE)
	r.Matchers = convertSliceToNilIfEmpty(r.Matchers)
	return nil
}

func checkNotEmptyMap(in ...map[string]string) bool {
	for _, input := range in {
		if len(input) > 0 {
			return true
		}
	}
	return false
}

func checkNotEmptyStrSlice(in ...[]string) bool {
	for _, input := range in {
		if len(input) > 0 {
			return true
		}
	}
	return false
}

func convertMapToNilIfEmpty(in map[string]string) map[string]string {
	if len(in) > 0 {
		return in
	}
	return nil
}

func convertSliceToNilIfEmpty(in []string) []string {
	if len(in) > 0 {
		return in
	}
	return nil
}

// contains will return true if any slice value with all whitespace removed
// is equal to the provided value with all whitespace removed
func contains(value string, in []string) bool {
	for _, str := range in {
		if strings.ReplaceAll(value, " ", "") == strings.ReplaceAll(str, " ", "") {
			return true
		}
	}
	return false
}

func inhibitRuleToV2(name, value string) string {
	return monitoringv1alpha1.Matcher{
		Name:      name,
		Value:     value,
		MatchType: monitoringv1alpha1.MatchEqual,
	}.String()
}

func inhibitRuleRegexToV2(name, value string) string {
	return monitoringv1alpha1.Matcher{
		Name:      name,
		Value:     value,
		MatchType: monitoringv1alpha1.MatchRegexp,
	}.String()
}

func checkIsV2Matcher(in ...[]monitoringv1alpha1.Matcher) bool {
	for _, input := range in {
		for _, matcher := range input {
			if matcher.MatchType != "" {
				return true
			}
		}
	}
	return false
}
