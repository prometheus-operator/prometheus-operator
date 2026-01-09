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
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"path"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/timeinterval"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	sortutil "github.com/prometheus-operator/prometheus-operator/internal/sortutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/alertmanager/validation"
	validationv1 "github.com/prometheus-operator/prometheus-operator/pkg/alertmanager/validation/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
)

const inhibitRuleNamespaceKey = "namespace"

// alertmanagerConfigFrom returns a valid alertmanagerConfig from b
// or returns an error if
// 1. s fails validation provided by upstream
// 2. s fails to unmarshal into internal type
// 3. the unmarshalled output is invalid.
func alertmanagerConfigFromBytes(b []byte) (*alertmanagerConfig, error) {
	// Run upstream Load function to get any validation checks that it runs.
	_, err := config.Load(string(b))
	if err != nil {
		return nil, err
	}

	cfg := &alertmanagerConfig{}
	err = yaml.UnmarshalStrict(b, cfg)
	if err != nil {
		return nil, err
	}

	if err := checkAlertmanagerConfigRootRoute(cfg.Route); err != nil {
		return nil, fmt.Errorf("check AlertmanagerConfig root route failed: %w", err)
	}

	return cfg, nil
}

func checkAlertmanagerConfigRootRoute(rootRoute *route) error {
	if rootRoute == nil {
		return errors.New("root route must exist")
	}

	if rootRoute.Continue {
		return errors.New("cannot have continue in root route")
	}

	if rootRoute.Receiver == "" {
		return errors.New("root route's receiver must exist")
	}

	if len(rootRoute.Matchers) > 0 || len(rootRoute.Match) > 0 || len(rootRoute.MatchRE) > 0 {
		return errors.New("'matchers' not permitted on root route")
	}

	if len(rootRoute.MuteTimeIntervals) > 0 {
		return errors.New("'mute_time_intervals' not permitted on root route")
	}

	if len(rootRoute.ActiveTimeIntervals) > 0 {
		return errors.New("'active_time_intervals' not permitted on root route")
	}

	return nil
}

func (c *alertmanagerConfig) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("<error creating config string: %s>", err)
	}
	return string(b)
}

type enforcer interface {
	processRoute(types.NamespacedName, *route) *route
	processInhibitRule(types.NamespacedName, *inhibitRule) *inhibitRule
}

// continueToNextRoute is an enforcer that always sets `continue: true` for the
// top-level route.
type continueToNextRoute struct {
	e enforcer
}

var _ enforcer = &continueToNextRoute{}

func (cte *continueToNextRoute) processRoute(crKey types.NamespacedName, r *route) *route {
	r = cte.e.processRoute(crKey, r)
	r.Continue = true

	return r
}

func (cte *continueToNextRoute) processInhibitRule(crKey types.NamespacedName, ir *inhibitRule) *inhibitRule {
	return cte.e.processInhibitRule(crKey, ir)
}

// noopEnforcer is a passthrough enforcer.
type noopEnforcer struct{}

var _ enforcer = &noopEnforcer{}

func (ne *noopEnforcer) processInhibitRule(_ types.NamespacedName, ir *inhibitRule) *inhibitRule {
	return ir
}

func (ne *noopEnforcer) processRoute(_ types.NamespacedName, r *route) *route {
	return r
}

// namespaceEnforcer enforces a namespace label matcher.
type namespaceEnforcer struct {
	matchersV2Allowed bool
}

var _ enforcer = &namespaceEnforcer{}

// processInhibitRule for namespaceEnforcer modifies the inhibition rule to
// match alerts originating only from the given namespace.
func (ne *namespaceEnforcer) processInhibitRule(crKey types.NamespacedName, ir *inhibitRule) *inhibitRule {
	// Inhibition rule created from AlertmanagerConfig resources should only match
	// alerts that come from the same namespace.
	delete(ir.SourceMatchRE, inhibitRuleNamespaceKey)
	delete(ir.TargetMatchRE, inhibitRuleNamespaceKey)

	if !ne.matchersV2Allowed {
		ir.SourceMatch[inhibitRuleNamespaceKey] = crKey.Namespace
		ir.TargetMatch[inhibitRuleNamespaceKey] = crKey.Namespace

		return ir
	}

	v2NamespaceMatcher := monitoringv1alpha1.Matcher{
		Name:      inhibitRuleNamespaceKey,
		Value:     crKey.Namespace,
		MatchType: monitoringv1alpha1.MatchEqual,
	}.String()

	if !contains(v2NamespaceMatcher, ir.SourceMatchers) {
		ir.SourceMatchers = append(ir.SourceMatchers, v2NamespaceMatcher)
	}
	if !contains(v2NamespaceMatcher, ir.TargetMatchers) {
		ir.TargetMatchers = append(ir.TargetMatchers, v2NamespaceMatcher)
	}

	delete(ir.SourceMatch, inhibitRuleNamespaceKey)
	delete(ir.TargetMatch, inhibitRuleNamespaceKey)

	return ir
}

// processRoute on namespaceEnforcer modifies the route configuration to match alerts
// originating only from the given namespace.
func (ne *namespaceEnforcer) processRoute(crKey types.NamespacedName, r *route) *route {
	// Routes created from AlertmanagerConfig resources should only match
	// alerts that come from the same namespace.
	if ne.matchersV2Allowed {
		r.Matchers = append(r.Matchers, monitoringv1alpha1.Matcher{
			Name:      "namespace",
			Value:     crKey.Namespace,
			MatchType: monitoringv1alpha1.MatchEqual,
		}.String())
	} else {
		r.Match["namespace"] = crKey.Namespace
	}

	return r
}

type otherNamespaceEnforcer struct {
	alertmanagerNamespace string
	namespaceEnforcer
}

var _ enforcer = &otherNamespaceEnforcer{}

func (one *otherNamespaceEnforcer) processInhibitRule(crKey types.NamespacedName, ir *inhibitRule) *inhibitRule {
	if crKey.Namespace == one.alertmanagerNamespace {
		return ir
	}
	return one.namespaceEnforcer.processInhibitRule(crKey, ir)
}

func (one *otherNamespaceEnforcer) processRoute(crKey types.NamespacedName, r *route) *route {
	if crKey.Namespace == one.alertmanagerNamespace {
		return r
	}
	return one.namespaceEnforcer.processRoute(crKey, r)
}

// ConfigBuilder knows how to build an Alertmanager configuration from a raw
// configuration and/or AlertmanagerConfig objects.
// The API is public because it's used by Grafana Alloy (https://github.com/grafana/alloy).
// Note that the project makes no API stability guarantees.
type ConfigBuilder struct {
	cfg       *alertmanagerConfig
	logger    *slog.Logger
	amVersion semver.Version
	store     *assets.StoreBuilder
	enforcer  enforcer
}

func NewConfigBuilder(logger *slog.Logger, amVersion semver.Version, store *assets.StoreBuilder, am *monitoringv1.Alertmanager) *ConfigBuilder {
	cg := &ConfigBuilder{
		logger:    logger,
		amVersion: amVersion,
		store:     store,
		enforcer:  getEnforcer(am.Spec.AlertmanagerConfigMatcherStrategy, amVersion, am.Namespace),
	}
	return cg
}

func getEnforcer(matcherStrategy monitoringv1.AlertmanagerConfigMatcherStrategy, amVersion semver.Version, amNamespace string) enforcer {
	var e enforcer
	switch matcherStrategy.Type {
	case monitoringv1.NoneConfigMatcherStrategyType:
		e = &noopEnforcer{}
	case monitoringv1.OnNamespaceExceptForAlertmanagerNamespaceConfigMatcherStrategyType:
		e = &otherNamespaceEnforcer{
			alertmanagerNamespace: amNamespace,
			namespaceEnforcer: namespaceEnforcer{
				matchersV2Allowed: amVersion.GTE(semver.MustParse("0.22.0")),
			},
		}
	default:
		e = &namespaceEnforcer{
			matchersV2Allowed: amVersion.GTE(semver.MustParse("0.22.0")),
		}
	}

	return &continueToNextRoute{e: e}
}

func (cb *ConfigBuilder) MarshalJSON() ([]byte, error) {
	return yaml.Marshal(cb.cfg)
}

// initializeFromAlertmanagerConfig initializes the configuration from an AlertmanagerConfig object.
func (cb *ConfigBuilder) initializeFromAlertmanagerConfig(ctx context.Context, globalConfig *monitoringv1.AlertmanagerGlobalConfig, amConfig *monitoringv1alpha1.AlertmanagerConfig) error {
	globalAlertmanagerConfig := &alertmanagerConfig{}

	crKey := types.NamespacedName{
		Namespace: amConfig.Namespace,
		Name:      amConfig.Name,
	}

	if err := checkAlertmanagerConfigResource(ctx, amConfig, cb.amVersion, cb.store); err != nil {
		return err
	}

	if err := cb.checkAlertmanagerGlobalConfigResource(ctx, globalConfig, crKey.Namespace); err != nil {
		return err
	}

	global, err := cb.convertGlobalConfig(ctx, globalConfig, crKey)
	if err != nil {
		return err
	}
	globalAlertmanagerConfig.Global = global

	// This is need to check required fields are set either at global or receiver level at later step.
	if global != nil {
		cb.cfg = &alertmanagerConfig{
			Global: global,
		}
	}

	// Add inhibitRules to globalAlertmanagerConfig.InhibitRules without enforce namespace
	for _, inhibitRule := range amConfig.Spec.InhibitRules {
		globalAlertmanagerConfig.InhibitRules = append(globalAlertmanagerConfig.InhibitRules, cb.convertInhibitRule(&inhibitRule))
	}

	// Add routes to globalAlertmanagerConfig.Route without enforce namespace
	globalAlertmanagerConfig.Route = cb.convertRoute(amConfig.Spec.Route, crKey)

	for _, receiver := range amConfig.Spec.Receivers {
		receivers, err := cb.convertReceiver(ctx, &receiver, crKey)
		if err != nil {
			return err
		}
		globalAlertmanagerConfig.Receivers = append(globalAlertmanagerConfig.Receivers, receivers)
	}

	for _, muteTimeInterval := range amConfig.Spec.MuteTimeIntervals {
		mti, err := convertMuteTimeInterval(&muteTimeInterval, crKey)
		if err != nil {
			return err
		}
		globalAlertmanagerConfig.MuteTimeIntervals = append(globalAlertmanagerConfig.MuteTimeIntervals, mti)
	}

	if err := globalAlertmanagerConfig.sanitize(cb.amVersion, cb.logger); err != nil {
		return err
	}

	if err := checkAlertmanagerConfigRootRoute(globalAlertmanagerConfig.Route); err != nil {
		return err
	}

	cb.cfg = globalAlertmanagerConfig
	return nil
}

// InitializeFromRawConfiguration initializes the configuration from raw data.
func (cb *ConfigBuilder) InitializeFromRawConfiguration(b []byte) error {
	globalAlertmanagerConfig, err := alertmanagerConfigFromBytes(b)
	if err != nil {
		return err
	}

	cb.cfg = globalAlertmanagerConfig
	return nil
}

// AddAlertmanagerConfigs adds AlertmanagerConfig objects to the current configuration.
func (cb *ConfigBuilder) AddAlertmanagerConfigs(ctx context.Context, amConfigs map[string]*monitoringv1alpha1.AlertmanagerConfig) error {
	subRoutes := make([]*route, 0, len(amConfigs))
	for _, amConfigIdentifier := range sortutil.SortedKeys(amConfigs) {
		crKey := types.NamespacedName{
			Name:      amConfigs[amConfigIdentifier].Name,
			Namespace: amConfigs[amConfigIdentifier].Namespace,
		}

		// Add inhibitRules to baseConfig.InhibitRules.
		for _, inhibitRule := range amConfigs[amConfigIdentifier].Spec.InhibitRules {
			cb.cfg.InhibitRules = append(cb.cfg.InhibitRules,
				cb.enforcer.processInhibitRule(
					crKey,
					cb.convertInhibitRule(
						&inhibitRule,
					),
				),
			)
		}

		// Skip early if there's no route definition.
		if amConfigs[amConfigIdentifier].Spec.Route == nil {
			continue
		}

		subRoutes = append(subRoutes,
			cb.enforcer.processRoute(
				crKey,
				cb.convertRoute(
					amConfigs[amConfigIdentifier].Spec.Route,
					crKey,
				),
			),
		)

		for _, receiver := range amConfigs[amConfigIdentifier].Spec.Receivers {
			receivers, err := cb.convertReceiver(ctx, &receiver, crKey)
			if err != nil {
				return fmt.Errorf("AlertmanagerConfig %s: %w", crKey.String(), err)
			}
			cb.cfg.Receivers = append(cb.cfg.Receivers, receivers)
		}

		for _, muteTimeInterval := range amConfigs[amConfigIdentifier].Spec.MuteTimeIntervals {
			mti, err := convertMuteTimeInterval(&muteTimeInterval, crKey)
			if err != nil {
				return fmt.Errorf("AlertmanagerConfig %s: %w", crKey.String(), err)
			}
			cb.cfg.MuteTimeIntervals = append(cb.cfg.MuteTimeIntervals, mti)
		}
	}

	// For alerts to be processed by the AlertmanagerConfig routes, they need
	// to appear before the routes defined in the main configuration.
	// Because all first-level AlertmanagerConfig routes have "continue: true",
	// alerts will fall through.
	cb.cfg.Route.Routes = append(subRoutes, cb.cfg.Route.Routes...)

	return cb.cfg.sanitize(cb.amVersion, cb.logger)
}

func (cb *ConfigBuilder) getValidURLFromSecret(ctx context.Context, namespace string, selector v1.SecretKeySelector) (string, error) {
	url, err := cb.store.GetSecretKey(ctx, namespace, selector)
	if err != nil {
		return "", fmt.Errorf("failed to get URL: %w", err)
	}

	url = strings.TrimSpace(url)
	if _, err := validation.ValidateURL(url); err != nil {
		return url, fmt.Errorf("invalid URL %q in key %q from secret %q: %w", url, selector.Key, selector.Name, err)
	}
	return url, nil
}

func (cb *ConfigBuilder) convertGlobalConfig(ctx context.Context, in *monitoringv1.AlertmanagerGlobalConfig, crKey types.NamespacedName) (*globalConfig, error) {
	if in == nil {
		return nil, nil
	}

	out := &globalConfig{}

	if in.SMTPConfig != nil {
		if err := cb.convertSMTPConfig(ctx, out, *in.SMTPConfig, crKey); err != nil {
			return nil, fmt.Errorf("invalid global smtpConfig: %w", err)
		}
	}

	if in.HTTPConfigWithProxy != nil {
		v1alpha1Config := monitoringv1alpha1.HTTPConfig{
			Authorization: in.HTTPConfigWithProxy.Authorization,
			BasicAuth:     in.HTTPConfigWithProxy.BasicAuth,
			OAuth2:        in.HTTPConfigWithProxy.OAuth2,
			//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
			BearerTokenSecret: in.HTTPConfigWithProxy.BearerTokenSecret,
			TLSConfig:         in.HTTPConfigWithProxy.TLSConfig,
			ProxyConfig:       in.HTTPConfigWithProxy.ProxyConfig,
			FollowRedirects:   in.HTTPConfigWithProxy.FollowRedirects,
			EnableHTTP2:       in.HTTPConfigWithProxy.EnableHTTP2,
		}

		httpConfig, err := cb.convertHTTPConfig(ctx, &v1alpha1Config, crKey)
		if err != nil {
			return nil, fmt.Errorf("invalid global httpConfig: %w", err)
		}

		if err := configureHTTPConfigInStore(ctx, &v1alpha1Config, crKey.Namespace, cb.store); err != nil {
			return nil, err
		}

		out.HTTPConfig = httpConfig
	}

	if in.ResolveTimeout != "" {
		timeout, err := model.ParseDuration(string(in.ResolveTimeout))
		if err != nil {
			return nil, fmt.Errorf("parse resolve timeout: %w", err)
		}
		out.ResolveTimeout = &timeout
	}

	if in.SlackAPIURL != nil {
		slackAPIURL, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.SlackAPIURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get Slack API URL: %w", err)
		}
		u, err := url.Parse(slackAPIURL)
		if err != nil {
			return nil, fmt.Errorf("parse slack API URL: %w", err)
		}
		out.SlackAPIURL = &config.URL{URL: u}
	}

	if in.OpsGenieAPIURL != nil {
		opsgenieAPIURL, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.OpsGenieAPIURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get OpsGenie API URL: %w", err)
		}
		u, err := url.Parse(opsgenieAPIURL)
		if err != nil {
			return nil, fmt.Errorf("parse OpsGenie API URL: %w", err)
		}
		out.OpsGenieAPIURL = &config.URL{URL: u}
	}

	if in.OpsGenieAPIKey != nil {
		opsGenieAPIKey, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.OpsGenieAPIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get OpsGenie API KEY: %w", err)
		}
		out.OpsGenieAPIKey = opsGenieAPIKey
	}

	if in.PagerdutyURL != nil {
		u, err := url.Parse(string(*in.PagerdutyURL))
		if err != nil {
			return nil, fmt.Errorf("parse Pagerduty URL: %w", err)
		}
		out.PagerdutyURL = &config.URL{URL: u}
	}

	if err := cb.convertGlobalTelegramConfig(out, in.TelegramConfig); err != nil {
		return nil, fmt.Errorf("invalid global telegram config: %w", err)
	}

	if err := cb.convertGlobalJiraConfig(out, in.JiraConfig); err != nil {
		return nil, fmt.Errorf("invalid global jira config: %w", err)
	}

	if err := cb.convertGlobalRocketChatConfig(ctx, out, in.RocketChatConfig, crKey); err != nil {
		return nil, fmt.Errorf("invalid global rocket chat config: %w", err)
	}

	if err := cb.convertGlobalWebexConfig(out, in.WebexConfig); err != nil {
		return nil, fmt.Errorf("invalid global webex config: %w", err)
	}

	if err := cb.convertGlobalWeChatConfig(ctx, out, in.WeChatConfig, crKey); err != nil {
		return nil, fmt.Errorf("invalid global wechat config: %w", err)
	}

	if err := cb.convertGlobalVictorOpsConfig(ctx, out, in.VictorOpsConfig, crKey); err != nil {
		return nil, fmt.Errorf("invalid global victorops config: %w", err)
	}

	return out, nil
}

func (cb *ConfigBuilder) convertRoute(in *monitoringv1alpha1.Route, crKey types.NamespacedName) *route {
	if in == nil {
		return nil
	}

	matchers, match, matchRE := cb.convertMatchersV2(in.Matchers)

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
			routes[i] = cb.convertRoute(&children[i], crKey)
		}
	}

	receiver := makeNamespacedString(in.Receiver, crKey)

	var prefixedMuteTimeIntervals []string
	if len(in.MuteTimeIntervals) > 0 {
		for _, mti := range in.MuteTimeIntervals {
			prefixedMuteTimeIntervals = append(prefixedMuteTimeIntervals, makeNamespacedString(mti, crKey))
		}
	}

	var prefixedActiveTimeIntervals []string
	if len(in.ActiveTimeIntervals) > 0 {
		for _, ati := range in.ActiveTimeIntervals {
			prefixedActiveTimeIntervals = append(prefixedActiveTimeIntervals, makeNamespacedString(ati, crKey))
		}
	}

	return &route{
		Receiver:            receiver,
		GroupByStr:          in.GroupBy,
		GroupWait:           in.GroupWait,
		GroupInterval:       in.GroupInterval,
		RepeatInterval:      in.RepeatInterval,
		Continue:            in.Continue,
		Match:               match,
		MatchRE:             matchRE,
		Matchers:            matchers,
		Routes:              routes,
		MuteTimeIntervals:   prefixedMuteTimeIntervals,
		ActiveTimeIntervals: prefixedActiveTimeIntervals,
	}
}

// convertReceiver converts a monitoringv1alpha1.Receiver to an alertmanager.receiver.
func (cb *ConfigBuilder) convertReceiver(ctx context.Context, in *monitoringv1alpha1.Receiver, crKey types.NamespacedName) (*receiver, error) {
	var pagerdutyConfigs []*pagerdutyConfig
	if l := len(in.PagerDutyConfigs); l > 0 {
		pagerdutyConfigs = make([]*pagerdutyConfig, l)
		for i := range in.PagerDutyConfigs {
			receiver, err := cb.convertPagerdutyConfig(ctx, in.PagerDutyConfigs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("PagerDutyConfig[%d]: %w", i, err)
			}
			pagerdutyConfigs[i] = receiver
		}
	}

	var discordConfigs []*discordConfig
	if l := len(in.DiscordConfigs); l > 0 {
		discordConfigs = make([]*discordConfig, l)
		for i := range in.DiscordConfigs {
			receiver, err := cb.convertDiscordConfig(ctx, in.DiscordConfigs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("DiscordConfig[%d]: %w", i, err)
			}
			discordConfigs[i] = receiver
		}
	}

	var slackConfigs []*slackConfig
	if l := len(in.SlackConfigs); l > 0 {
		slackConfigs = make([]*slackConfig, l)
		for i := range in.SlackConfigs {
			receiver, err := cb.convertSlackConfig(ctx, in.SlackConfigs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("SlackConfig[%d]: %w", i, err)
			}
			slackConfigs[i] = receiver
		}
	}

	var webhookConfigs []*webhookConfig
	if l := len(in.WebhookConfigs); l > 0 {
		webhookConfigs = make([]*webhookConfig, l)
		for i := range in.WebhookConfigs {
			receiver, err := cb.convertWebhookConfig(ctx, in.WebhookConfigs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("WebhookConfig[%d]: %w", i, err)
			}
			webhookConfigs[i] = receiver
		}
	}

	var opsgenieConfigs []*opsgenieConfig
	if l := len(in.OpsGenieConfigs); l > 0 {
		opsgenieConfigs = make([]*opsgenieConfig, l)
		for i := range in.OpsGenieConfigs {
			receiver, err := cb.convertOpsgenieConfig(ctx, in.OpsGenieConfigs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("OpsGenieConfigs[%d]: %w", i, err)
			}
			opsgenieConfigs[i] = receiver
		}
	}

	var weChatConfigs []*weChatConfig
	if l := len(in.WeChatConfigs); l > 0 {
		weChatConfigs = make([]*weChatConfig, l)
		for i := range in.WeChatConfigs {
			receiver, err := cb.convertWeChatConfig(ctx, in.WeChatConfigs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("WeChatConfig[%d]: %w", i, err)
			}
			weChatConfigs[i] = receiver
		}
	}

	var emailConfigs []*emailConfig
	if l := len(in.EmailConfigs); l > 0 {
		emailConfigs = make([]*emailConfig, l)
		for i := range in.EmailConfigs {
			receiver, err := cb.convertEmailConfig(ctx, in.EmailConfigs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("EmailConfig[%d]: %w", i, err)
			}
			emailConfigs[i] = receiver
		}
	}

	var victorOpsConfigs []*victorOpsConfig
	if l := len(in.VictorOpsConfigs); l > 0 {
		victorOpsConfigs = make([]*victorOpsConfig, l)
		for i := range in.VictorOpsConfigs {
			receiver, err := cb.convertVictorOpsConfig(ctx, in.VictorOpsConfigs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("VictorOpsConfig[%d]: %w", i, err)
			}
			victorOpsConfigs[i] = receiver
		}
	}

	var pushoverConfigs []*pushoverConfig
	if l := len(in.PushoverConfigs); l > 0 {
		pushoverConfigs = make([]*pushoverConfig, l)
		for i := range in.PushoverConfigs {
			receiver, err := cb.convertPushoverConfig(ctx, in.PushoverConfigs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("PushoverConfig[%d]: %w", i, err)
			}
			pushoverConfigs[i] = receiver
		}
	}

	var snsConfigs []*snsConfig
	if l := len(in.SNSConfigs); l > 0 {
		snsConfigs = make([]*snsConfig, l)
		for i := range in.SNSConfigs {
			receiver, err := cb.convertSnsConfig(ctx, in.SNSConfigs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("SNSConfig[%d]: %w", i, err)
			}
			snsConfigs[i] = receiver
		}
	}

	var telegramConfigs []*telegramConfig
	if l := len(in.TelegramConfigs); l > 0 {
		telegramConfigs = make([]*telegramConfig, l)
		for i := range in.TelegramConfigs {
			receiver, err := cb.convertTelegramConfig(ctx, in.TelegramConfigs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("TelegramConfig[%d]: %w", i, err)
			}
			telegramConfigs[i] = receiver
		}
	}

	var msTeamsConfigs []*msTeamsConfig
	if l := len(in.MSTeamsConfigs); l > 0 {
		msTeamsConfigs = make([]*msTeamsConfig, l)
		for i := range in.MSTeamsConfigs {
			receiver, err := cb.convertMSTeamsConfig(ctx, in.MSTeamsConfigs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("MSTeamsConfig[%d]: %w", i, err)
			}
			msTeamsConfigs[i] = receiver
		}
	}

	var msTeamsV2Configs []*msTeamsV2Config
	if l := len(in.MSTeamsV2Configs); l > 0 {
		msTeamsV2Configs = make([]*msTeamsV2Config, l)
		for i := range in.MSTeamsV2Configs {
			receiver, err := cb.convertMSTeamsV2Config(ctx, in.MSTeamsV2Configs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("MSTeamsConfigV2[%d]: %w", i, err)
			}
			msTeamsV2Configs[i] = receiver
		}
	}

	var webexConfigs []*webexConfig
	if l := len(in.WebexConfigs); l > 0 {
		webexConfigs = make([]*webexConfig, l)
		for i := range in.WebexConfigs {
			receiver, err := cb.convertWebexConfig(ctx, in.WebexConfigs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("WebexConfig[%d]: %w", i, err)
			}
			webexConfigs[i] = receiver
		}
	}

	var rocketchatConfigs []*rocketChatConfig
	if l := len(in.RocketChatConfigs); l > 0 {
		rocketchatConfigs = make([]*rocketChatConfig, l)
		for i := range in.RocketChatConfigs {
			receiver, err := cb.convertRocketChatConfig(ctx, in.RocketChatConfigs[i], crKey)
			if err != nil {
				return nil, fmt.Errorf("RocketChatConfig[%d]: %w", i, err)
			}
			rocketchatConfigs[i] = receiver
		}
	}

	return &receiver{
		Name:              makeNamespacedString(in.Name, crKey),
		OpsgenieConfigs:   opsgenieConfigs,
		PagerdutyConfigs:  pagerdutyConfigs,
		DiscordConfigs:    discordConfigs,
		SlackConfigs:      slackConfigs,
		WebhookConfigs:    webhookConfigs,
		WeChatConfigs:     weChatConfigs,
		EmailConfigs:      emailConfigs,
		VictorOpsConfigs:  victorOpsConfigs,
		PushoverConfigs:   pushoverConfigs,
		SNSConfigs:        snsConfigs,
		TelegramConfigs:   telegramConfigs,
		WebexConfigs:      webexConfigs,
		MSTeamsConfigs:    msTeamsConfigs,
		MSTeamsV2Configs:  msTeamsV2Configs,
		RocketChatConfigs: rocketchatConfigs,
	}, nil
}

func (cb *ConfigBuilder) convertRocketChatConfig(ctx context.Context, in monitoringv1alpha1.RocketChatConfig, crKey types.NamespacedName) (*rocketChatConfig, error) {
	out := &rocketChatConfig{
		SendResolved: in.SendResolved,
	}

	token, err := cb.store.GetSecretKey(ctx, crKey.Namespace, in.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to get RocketChat token: %w", err)
	}
	out.Token = &token

	tokenID, err := cb.store.GetSecretKey(ctx, crKey.Namespace, in.TokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to get RocketChat token ID: %w", err)
	}
	out.TokenID = &tokenID

	httpConfig, err := cb.convertHTTPConfig(ctx, in.HTTPConfig, crKey)
	if err != nil {
		return nil, err
	}
	out.HTTPConfig = httpConfig

	return out, nil
}

func (cb *ConfigBuilder) convertWebhookConfig(ctx context.Context, in monitoringv1alpha1.WebhookConfig, crKey types.NamespacedName) (*webhookConfig, error) {
	out := &webhookConfig{
		VSendResolved: in.SendResolved,
	}

	if in.URLSecret != nil {
		url, err := cb.getValidURLFromSecret(ctx, crKey.Namespace, *in.URLSecret)
		if err != nil {
			return nil, err
		}
		out.URL = url
	} else if in.URL != nil {
		url, err := validation.ValidateURL(string(*in.URL))
		if err != nil {
			return nil, err
		}
		out.URL = url.String()
	}

	httpConfig, err := cb.convertHTTPConfig(ctx, in.HTTPConfig, crKey)
	if err != nil {
		return nil, err
	}
	out.HTTPConfig = httpConfig

	if in.MaxAlerts > 0 {
		out.MaxAlerts = in.MaxAlerts
	}

	if in.Timeout != nil {
		if *in.Timeout != "" {
			timeout, err := model.ParseDuration(string(*in.Timeout))
			if err != nil {
				return nil, err
			}
			out.Timeout = &timeout
		}
	}

	return out, nil
}

func (cb *ConfigBuilder) convertDiscordConfig(ctx context.Context, in monitoringv1alpha1.DiscordConfig, crKey types.NamespacedName) (*discordConfig, error) {
	out := &discordConfig{
		VSendResolved: in.SendResolved,
	}

	if in.Title != nil && *in.Title != "" {
		out.Title = *in.Title
	}

	if in.Message != nil && *in.Message != "" {
		out.Message = *in.Message
	}

	if in.Content != nil && *in.Content != "" {
		out.Content = *in.Content
	}

	if in.Username != nil && *in.Username != "" {
		out.Username = *in.Username
	}

	if in.AvatarURL != nil && *in.AvatarURL != "" {
		out.AvatarURL = (string)(*in.AvatarURL)
	}

	url, err := cb.getValidURLFromSecret(ctx, crKey.Namespace, in.APIURL)
	if err != nil {
		return nil, err
	}
	out.WebhookURL = url

	httpConfig, err := cb.convertHTTPConfig(ctx, in.HTTPConfig, crKey)
	if err != nil {
		return nil, err
	}
	out.HTTPConfig = httpConfig

	return out, nil
}

func (cb *ConfigBuilder) convertSlackConfig(ctx context.Context, in monitoringv1alpha1.SlackConfig, crKey types.NamespacedName) (*slackConfig, error) {
	out := &slackConfig{
		VSendResolved: in.SendResolved,
		Channel:       ptr.Deref(in.Channel, ""),
		Username:      ptr.Deref(in.Username, ""),
		Color:         ptr.Deref(in.Color, ""),
		Title:         ptr.Deref(in.Title, ""),
		Pretext:       ptr.Deref(in.Pretext, ""),
		Text:          ptr.Deref(in.Text, ""),
		ShortFields:   ptr.Deref(in.ShortFields, false),
		Footer:        ptr.Deref(in.Footer, ""),
		Fallback:      ptr.Deref(in.Fallback, ""),
		CallbackID:    ptr.Deref(in.CallbackID, ""),
		IconEmoji:     ptr.Deref(in.IconEmoji, ""),
		LinkNames:     ptr.Deref(in.LinkNames, false),
		MrkdwnIn:      in.MrkdwnIn,
	}

	if in.APIURL != nil {
		url, err := cb.getValidURLFromSecret(ctx, crKey.Namespace, *in.APIURL)
		if err != nil {
			return nil, err
		}
		out.APIURL = url
	}

	if ptr.Deref(in.TitleLink, "") != "" {
		out.TitleLink = string(*in.TitleLink)
	}
	if ptr.Deref(in.IconURL, "") != "" {
		out.TitleLink = string(*in.IconURL)
	}
	if ptr.Deref(in.ImageURL, "") != "" {
		out.TitleLink = string(*in.ImageURL)
	}
	if ptr.Deref(in.ThumbURL, "") != "" {
		out.TitleLink = string(*in.ThumbURL)
	}

	var actions []slackAction
	if l := len(in.Actions); l > 0 {
		actions = make([]slackAction, l)
		for i, a := range in.Actions {
			action := slackAction{
				Type:  a.Type,
				Text:  a.Text,
				Style: ptr.Deref(a.Style, ""),
				Name:  ptr.Deref(a.Name, ""),
				Value: ptr.Deref(a.Value, ""),
			}

			if ptr.Deref(a.URL, "") != "" {
				action.URL = string(*a.URL)
			}

			if a.ConfirmField != nil {
				action.ConfirmField = &slackConfirmationField{
					Text:        a.ConfirmField.Text,
					Title:       ptr.Deref(a.ConfirmField.Title, ""),
					OkText:      ptr.Deref(a.ConfirmField.OkText, ""),
					DismissText: ptr.Deref(a.ConfirmField.DismissText, ""),
				}
			}

			actions[i] = action
		}
		out.Actions = actions
	}

	if l := len(in.Fields); l > 0 {
		fields := make([]slackField, l)
		for i, f := range in.Fields {
			field := slackField{
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

	httpConfig, err := cb.convertHTTPConfig(ctx, in.HTTPConfig, crKey)
	if err != nil {
		return nil, err
	}
	out.HTTPConfig = httpConfig

	if in.Timeout != nil {
		if *in.Timeout != "" {
			timeout, err := model.ParseDuration(string(*in.Timeout))
			if err != nil {
				return nil, err
			}
			out.Timeout = &timeout
		}
	}

	return out, nil
}

func (cb *ConfigBuilder) convertPagerdutyConfig(ctx context.Context, in monitoringv1alpha1.PagerDutyConfig, crKey types.NamespacedName) (*pagerdutyConfig, error) {
	out := &pagerdutyConfig{
		VSendResolved: in.SendResolved,
		Class:         ptr.Deref(in.Class, ""),
		Client:        ptr.Deref(in.Client, ""),
		Component:     ptr.Deref(in.Component, ""),
		Description:   ptr.Deref(in.Description, ""),
		Group:         ptr.Deref(in.Group, ""),
		Severity:      ptr.Deref(in.Severity, ""),
	}

	if in.URL != nil {
		out.URL = string(*in.URL)
	}

	if in.ClientURL != nil {
		out.ClientURL = string(*in.ClientURL)
	}

	if in.RoutingKey != nil {
		routingKey, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.RoutingKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get routing key: %w", err)
		}
		out.RoutingKey = routingKey
	}

	if in.ServiceKey != nil {
		serviceKey, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.ServiceKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get service key: %w", err)
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

	var linkConfigs []pagerdutyLink
	if l := len(in.PagerDutyLinkConfigs); l > 0 {
		linkConfigs = make([]pagerdutyLink, l)
		for i, lc := range in.PagerDutyLinkConfigs {
			linkConfigs[i] = pagerdutyLink{
				Text: ptr.Deref(lc.Text, ""),
			}
			if lc.Href != nil {
				linkConfigs[i].Href = string(*lc.Href)
			}
		}
	}
	out.Links = linkConfigs

	var imageConfig []pagerdutyImage
	if l := len(in.PagerDutyImageConfigs); l > 0 {
		imageConfig = make([]pagerdutyImage, l)
		for i, ic := range in.PagerDutyImageConfigs {
			imageConfig[i] = pagerdutyImage{
				Src: ptr.Deref(ic.Src, ""),
				Alt: ptr.Deref(ic.Alt, ""),
			}
			if ic.Href != nil {
				imageConfig[i].Href = string(*ic.Href)
			}
		}
	}
	out.Images = imageConfig

	httpConfig, err := cb.convertHTTPConfig(ctx, in.HTTPConfig, crKey)
	if err != nil {
		return nil, err
	}
	out.HTTPConfig = httpConfig

	if in.Source != nil {
		out.Source = *in.Source
	}

	if in.Timeout != nil {
		if *in.Timeout != "" {
			timeout, err := model.ParseDuration(string(*in.Timeout))
			if err != nil {
				return nil, err
			}
			out.Timeout = &timeout
		}
	}

	return out, nil
}

func (cb *ConfigBuilder) convertOpsgenieConfig(ctx context.Context, in monitoringv1alpha1.OpsGenieConfig, crKey types.NamespacedName) (*opsgenieConfig, error) {
	out := &opsgenieConfig{
		VSendResolved: in.SendResolved,
		Message:       in.Message,
		Description:   in.Description,
		Source:        in.Source,
		Tags:          in.Tags,
		Note:          in.Note,
		Priority:      in.Priority,
		Actions:       in.Actions,
		Entity:        in.Entity,
		UpdateAlerts:  in.UpdateAlerts,
	}

	if in.APIURL != nil {
		out.APIURL = string(*in.APIURL)
	}

	if in.APIKey != nil {
		apiKey, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.APIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get API key: %w", err)
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
			responder := opsgenieResponder{
				ID:       r.ID,
				Name:     r.Name,
				Username: r.Username,
				Type:     r.Type,
			}
			responders = append(responders, responder)
		}
	}
	out.Responders = responders

	httpConfig, err := cb.convertHTTPConfig(ctx, in.HTTPConfig, crKey)
	if err != nil {
		return nil, err
	}
	out.HTTPConfig = httpConfig

	return out, nil
}

func (cb *ConfigBuilder) convertWeChatConfig(ctx context.Context, in monitoringv1alpha1.WeChatConfig, crKey types.NamespacedName) (*weChatConfig, error) {
	out := &weChatConfig{
		VSendResolved: in.SendResolved,
		CorpID:        in.CorpID,
		AgentID:       in.AgentID,
		ToUser:        in.ToUser,
		ToParty:       in.ToParty,
		ToTag:         in.ToTag,
		Message:       in.Message,
		MessageType:   in.MessageType,
	}

	if in.APIURL != nil {
		out.APIURL = string(*in.APIURL)
	}

	if in.APISecret != nil {
		apiSecret, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.APISecret)
		if err != nil {
			return nil, fmt.Errorf("failed to get API secret: %w", err)
		}
		out.APISecret = apiSecret
	}

	httpConfig, err := cb.convertHTTPConfig(ctx, in.HTTPConfig, crKey)
	if err != nil {
		return nil, err
	}
	out.HTTPConfig = httpConfig

	return out, nil
}

func (cb *ConfigBuilder) convertWebexConfig(ctx context.Context, in monitoringv1alpha1.WebexConfig, crKey types.NamespacedName) (*webexConfig, error) {
	out := &webexConfig{
		VSendResolved: in.SendResolved,
		RoomID:        in.RoomID,
	}

	if in.APIURL != nil {
		out.APIURL = string(*in.APIURL)
	}

	if in.Message != nil {
		out.Message = *in.Message
	}

	httpConfig, err := cb.convertHTTPConfig(ctx, in.HTTPConfig, crKey)
	if err != nil {
		return nil, err
	}
	out.HTTPConfig = httpConfig

	return out, nil
}

func (cb *ConfigBuilder) convertEmailConfig(ctx context.Context, in monitoringv1alpha1.EmailConfig, crKey types.NamespacedName) (*emailConfig, error) {
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

	if in.Smarthost == "" {
		if cb.cfg.Global == nil || cb.cfg.Global.SMTPSmarthost.Host == "" {
			return nil, fmt.Errorf("SMTP smarthost is a mandatory field, it is neither specified at global config nor at receiver level")
		}
	}

	if in.From == "" {
		if cb.cfg.Global == nil || cb.cfg.Global.SMTPFrom == "" {
			return nil, fmt.Errorf("SMTP from is a mandatory field, it is neither specified at global config nor at receiver level")
		}
	}

	if in.Smarthost != "" {
		out.Smarthost.Host, out.Smarthost.Port, _ = net.SplitHostPort(in.Smarthost)
	}

	if in.AuthPassword != nil {
		authPassword, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.AuthPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to get auth password: %w", err)
		}
		out.AuthPassword = authPassword
	}

	if in.AuthSecret != nil {
		authSecret, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.AuthSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to get auth secret: %w", err)
		}
		out.AuthSecret = authSecret
	}

	if l := len(in.Headers); l > 0 {
		headers := make(map[string]string, l)
		for _, d := range in.Headers {
			headers[d.Key] = d.Value
		}
		out.Headers = headers
	}

	if in.TLSConfig != nil {
		out.TLSConfig = cb.convertTLSConfig(in.TLSConfig, crKey)
	}

	return out, nil
}

func (cb *ConfigBuilder) convertVictorOpsConfig(ctx context.Context, in monitoringv1alpha1.VictorOpsConfig, crKey types.NamespacedName) (*victorOpsConfig, error) {
	out := &victorOpsConfig{
		VSendResolved:     in.SendResolved,
		RoutingKey:        in.RoutingKey,
		MessageType:       ptr.Deref(in.MessageType, ""),
		EntityDisplayName: ptr.Deref(in.EntityDisplayName, ""),
		StateMessage:      ptr.Deref(in.StateMessage, ""),
		MonitoringTool:    ptr.Deref(in.MonitoringTool, ""),
	}

	if in.APIKey != nil {
		apiKey, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.APIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get API key: %w", err)
		}
		out.APIKey = apiKey
	}

	if in.APIURL != nil {
		out.APIURL = string(*in.APIURL)
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
				return nil, fmt.Errorf("VictorOps config contains custom field %s which cannot be used as it conflicts with the fixed/static fields", d.Key)
			}
			customFields[d.Key] = d.Value
		}
	}
	out.CustomFields = customFields

	httpConfig, err := cb.convertHTTPConfig(ctx, in.HTTPConfig, crKey)
	if err != nil {
		return nil, err
	}
	out.HTTPConfig = httpConfig

	return out, nil
}

func (cb *ConfigBuilder) convertPushoverConfig(ctx context.Context, in monitoringv1alpha1.PushoverConfig, crKey types.NamespacedName) (*pushoverConfig, error) {
	out := &pushoverConfig{
		VSendResolved: in.SendResolved,
		Title:         ptr.Deref(in.Title, ""),
		Message:       ptr.Deref(in.Message, ""),
		URLTitle:      ptr.Deref(in.URLTitle, ""),
		Priority:      ptr.Deref(in.Priority, ""),
		HTML:          in.HTML,
		Monospace:     in.Monospace,
	}

	if ptr.Deref(in.URL, "") != "" {
		out.URL = string(*in.URL)
	}

	if in.TTL != nil {
		out.TTL = string(*in.TTL)
	}

	if in.Device != nil {
		out.Device = *in.Device
	}

	{
		userKey, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.UserKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get user key: %w", err)
		}
		if userKey == "" {
			return nil, fmt.Errorf("mandatory field %q is empty", "userKey")
		}
		out.UserKey = userKey
	}

	{
		token, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to get token: %w", err)
		}
		if token == "" {
			return nil, fmt.Errorf("mandatory field %q is empty", "token")
		}
		out.Token = token
	}

	{
		if ptr.Deref(in.Retry, "") != "" {
			retry, err := model.ParseDuration(*in.Retry)
			if err != nil {
				return nil, fmt.Errorf("parse resolve retry: %w", err)
			}
			out.Retry = &retry
		}

		if ptr.Deref(in.Expire, "") != "" {
			expire, err := model.ParseDuration(*in.Expire)
			if err != nil {
				return nil, fmt.Errorf("parse resolve expire: %w", err)
			}
			out.Expire = &expire
		}
	}

	httpConfig, err := cb.convertHTTPConfig(ctx, in.HTTPConfig, crKey)
	if err != nil {
		return nil, err
	}
	out.HTTPConfig = httpConfig

	return out, nil
}

func (cb *ConfigBuilder) convertTelegramConfig(ctx context.Context, in monitoringv1alpha1.TelegramConfig, crKey types.NamespacedName) (*telegramConfig, error) {
	out := &telegramConfig{
		VSendResolved:        in.SendResolved,
		ChatID:               in.ChatID,
		Message:              in.Message,
		DisableNotifications: false,
		ParseMode:            in.ParseMode,
	}

	if in.APIURL != nil {
		out.APIUrl = string(*in.APIURL)
	}

	httpConfig, err := cb.convertHTTPConfig(ctx, in.HTTPConfig, crKey)
	if err != nil {
		return nil, err
	}
	out.HTTPConfig = httpConfig

	if in.MessageThreadID != nil {
		out.MessageThreadID = int(*in.MessageThreadID)
	}

	if in.BotToken != nil {
		botToken, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.BotToken)
		if err != nil {
			return nil, fmt.Errorf("failed to get bot token: %w", err)
		}
		if botToken == "" {
			return nil, fmt.Errorf("mandatory field %q is empty", "botToken")
		}
		out.BotToken = botToken
	}

	return out, nil
}

func (cb *ConfigBuilder) convertSnsConfig(ctx context.Context, in monitoringv1alpha1.SNSConfig, crKey types.NamespacedName) (*snsConfig, error) {
	out := &snsConfig{
		VSendResolved: in.SendResolved,
		APIUrl:        in.ApiURL,
		TopicARN:      in.TopicARN,
		PhoneNumber:   in.PhoneNumber,
		TargetARN:     in.TargetARN,
		Subject:       in.Subject,
		Message:       in.Message,
		Attributes:    in.Attributes,
	}

	httpConfig, err := cb.convertHTTPConfig(ctx, in.HTTPConfig, crKey)
	if err != nil {
		return nil, err
	}
	out.HTTPConfig = httpConfig

	if in.Sigv4 != nil {
		out.Sigv4 = sigV4Config{
			Region:  in.Sigv4.Region,
			Profile: in.Sigv4.Profile,
			RoleARN: in.Sigv4.RoleArn,
		}

		if in.Sigv4.AccessKey != nil && in.Sigv4.SecretKey != nil {
			accessKey, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.Sigv4.AccessKey)
			if err != nil {
				return nil, fmt.Errorf("failed to get access key: %w", err)
			}

			secretKey, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.Sigv4.SecretKey)
			if err != nil {
				return nil, fmt.Errorf("failed to get AWS secret key: %w", err)

			}
			out.Sigv4.AccessKey = accessKey
			out.Sigv4.SecretKey = secretKey
		}
	}

	return out, nil
}

func (cb *ConfigBuilder) convertMSTeamsConfig(
	ctx context.Context, in monitoringv1alpha1.MSTeamsConfig, crKey types.NamespacedName,
) (*msTeamsConfig, error) {
	out := &msTeamsConfig{
		SendResolved: in.SendResolved,
	}

	if in.Title != nil {
		out.Title = *in.Title
	}

	if in.Text != nil {
		out.Text = *in.Text
	}

	if in.Summary != nil {
		out.Summary = *in.Summary
	}

	webHookURL, err := cb.store.GetSecretKey(ctx, crKey.Namespace, in.WebhookURL)
	if err != nil {
		return nil, err
	}

	out.WebhookURL = webHookURL

	httpConfig, err := cb.convertHTTPConfig(ctx, in.HTTPConfig, crKey)
	if err != nil {
		return nil, err
	}
	out.HTTPConfig = httpConfig

	return out, nil
}

func (cb *ConfigBuilder) convertMSTeamsV2Config(
	ctx context.Context, in monitoringv1alpha1.MSTeamsV2Config, crKey types.NamespacedName,
) (*msTeamsV2Config, error) {
	out := &msTeamsV2Config{
		SendResolved: in.SendResolved,
	}

	if in.Title != nil {
		out.Title = *in.Title
	}

	if in.Text != nil {
		out.Text = *in.Text
	}

	if in.WebhookURL != nil {
		webHookURL, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.WebhookURL)
		if err != nil {
			return nil, err
		}

		out.WebhookURL = webHookURL
	}

	httpConfig, err := cb.convertHTTPConfig(ctx, in.HTTPConfig, crKey)
	if err != nil {
		return nil, err
	}
	out.HTTPConfig = httpConfig

	return out, nil
}

func (cb *ConfigBuilder) convertInhibitRule(in *monitoringv1alpha1.InhibitRule) *inhibitRule {
	sourceMatchers, sourceMatch, sourceMatchRE := cb.convertMatchersV2(in.SourceMatch)
	targetMatchers, targetMatch, targetMatchRE := cb.convertMatchersV2(in.TargetMatch)

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

func (cb *ConfigBuilder) convertMatchersV2(ms []monitoringv1alpha1.Matcher) ([]string, map[string]string, map[string]string) {
	matchersV2Allowed := cb.amVersion.GTE(semver.MustParse("0.22.0"))

	var matchers []string
	match := map[string]string{}
	matchRE := map[string]string{}

	for _, m := range ms {
		if m.MatchType != "" {
			matchers = append(matchers, m.String())
			continue
		}

		if matchersV2Allowed {
			if m.Regex {
				matchers = append(matchers, inhibitRuleRegexToV2(m.Name, m.Value))
			} else {
				matchers = append(matchers, inhibitRuleToV2(m.Name, m.Value))
			}
			continue
		}

		if m.Regex {
			matchRE[m.Name] = m.Value
		} else {
			match[m.Name] = m.Value
		}
	}

	return matchers, match, matchRE
}

func convertMuteTimeInterval(in *monitoringv1alpha1.MuteTimeInterval, crKey types.NamespacedName) (*timeInterval, error) {
	muteTimeInterval := &timeInterval{}

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
	return crKey.Namespace + "/" + crKey.Name + "/" + in
}

func (cb *ConfigBuilder) convertSMTPConfig(ctx context.Context, out *globalConfig, in monitoringv1.GlobalSMTPConfig, crKey types.NamespacedName) error {
	if in.From != nil {
		out.SMTPFrom = *in.From
	}
	if in.Hello != nil {
		out.SMTPHello = *in.Hello
	}
	if in.AuthUsername != nil {
		out.SMTPAuthUsername = *in.AuthUsername
	}
	if in.AuthIdentity != nil {
		out.SMTPAuthIdentity = *in.AuthIdentity
	}
	out.SMTPRequireTLS = in.RequireTLS

	if in.SmartHost != nil {
		out.SMTPSmarthost.Host = in.SmartHost.Host
		out.SMTPSmarthost.Port = in.SmartHost.Port
	}

	if in.TLSConfig != nil {
		out.SMTPTLSConfig = cb.convertTLSConfig(in.TLSConfig, crKey)
	}

	if in.AuthPassword != nil {
		authPassword, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.AuthPassword)
		if err != nil {
			return err
		}
		out.SMTPAuthPassword = authPassword
	}

	if in.AuthSecret != nil {
		authSecret, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.AuthSecret)
		if err != nil {
			return err
		}
		out.SMTPAuthSecret = authSecret
	}

	return nil
}

// convertHTTPConfig converts the HTTPConfig CRD field to the internal configuration struct.
func (cb *ConfigBuilder) convertHTTPConfig(ctx context.Context, in *monitoringv1alpha1.HTTPConfig, crKey types.NamespacedName) (*httpClientConfig, error) {
	if in == nil {
		return nil, nil
	}

	proxyConfig, err := cb.convertProxyConfig(ctx, in.ProxyConfig, crKey)
	if err != nil {
		return nil, err
	}

	// in.ProxyURL comes from the common v1.ProxyConfig struct and is
	// serialized as `proxyUrl` while in.ProxyURLOriginal is serialized as
	// `proxyURL`. ProxyURLOriginal existed first in the CRD spec hence it
	// can't be removed till the next API bump and should take precedence over
	// in.ProxyURL.
	if ptr.Deref(in.ProxyURLOriginal, "") != "" {
		proxyConfig.ProxyURL = *in.ProxyURLOriginal
	}

	out := &httpClientConfig{
		proxyConfig:     proxyConfig,
		FollowRedirects: in.FollowRedirects,
		EnableHTTP2:     in.EnableHTTP2,
	}

	if in.BasicAuth != nil {
		username, err := cb.store.GetSecretKey(ctx, crKey.Namespace, in.BasicAuth.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to get BasicAuth username: %w", err)
		}

		password, err := cb.store.GetSecretKey(ctx, crKey.Namespace, in.BasicAuth.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to get BasicAuth password: %w", err)
		}

		if username != "" || password != "" {
			out.BasicAuth = &basicAuth{Username: username, Password: password}
		}
	}

	if in.Authorization != nil {
		credentials, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.Authorization.Credentials)
		if err != nil {
			return nil, fmt.Errorf("failed to get Authorization credentials: %w", err)
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
		out.TLSConfig = cb.convertTLSConfig(in.TLSConfig, crKey)
	}

	if in.BearerTokenSecret != nil {
		bearerToken, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.BearerTokenSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to get bearer token: %w", err)
		}
		out.BearerToken = bearerToken
	}

	if in.OAuth2 != nil {
		clientID, err := cb.store.GetKey(ctx, crKey.Namespace, in.OAuth2.ClientID)
		if err != nil {
			return nil, fmt.Errorf("failed to get clientID: %w", err)
		}

		clientSecret, err := cb.store.GetSecretKey(ctx, crKey.Namespace, in.OAuth2.ClientSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to get client secret: %w", err)
		}
		proxyConfig, err := cb.convertProxyConfig(ctx, in.OAuth2.ProxyConfig, crKey)
		if err != nil {
			return nil, err
		}
		out.OAuth2 = &oauth2{
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			Scopes:         in.OAuth2.Scopes,
			TokenURL:       in.OAuth2.TokenURL,
			EndpointParams: in.OAuth2.EndpointParams,
			proxyConfig:    proxyConfig,
		}
	}

	return out, nil
}

func (cb *ConfigBuilder) convertTLSConfig(in *monitoringv1.SafeTLSConfig, crKey types.NamespacedName) *tlsConfig {
	out := tlsConfig{}

	if in.ServerName != nil {
		out.ServerName = *in.ServerName
	}

	if in.InsecureSkipVerify != nil {
		out.InsecureSkipVerify = *in.InsecureSkipVerify
	}

	s := cb.store.ForNamespace(crKey.Namespace)

	if in.CA != (monitoringv1.SecretOrConfigMap{}) {
		out.CAFile = path.Join(tlsAssetsDir, s.TLSAsset(in.CA))
	}

	if in.Cert != (monitoringv1.SecretOrConfigMap{}) {
		out.CertFile = path.Join(tlsAssetsDir, s.TLSAsset(in.Cert))
	}

	if in.KeySecret != nil {
		out.KeyFile = path.Join(tlsAssetsDir, s.TLSAsset(in.KeySecret))
	}

	if in.MinVersion != nil {
		out.MinVersion = string(*in.MinVersion)
	}

	if in.MaxVersion != nil {
		out.MaxVersion = string(*in.MaxVersion)
	}

	return &out
}

func (cb *ConfigBuilder) convertProxyConfig(ctx context.Context, in monitoringv1.ProxyConfig, crKey types.NamespacedName) (proxyConfig, error) {
	out := proxyConfig{}

	if in.ProxyURL != nil {
		out.ProxyURL = *in.ProxyURL
	}

	if in.NoProxy != nil {
		out.NoProxy = *in.NoProxy
	}

	if in.ProxyFromEnvironment != nil {
		out.ProxyFromEnvironment = *in.ProxyFromEnvironment
	}

	if len(in.ProxyConnectHeader) > 0 {
		proxyConnectHeader := make(map[string][]string, len(in.ProxyConnectHeader))
		for k, v := range in.ProxyConnectHeader {
			proxyConnectHeader[k] = []string{}
			for _, vv := range v {
				value, err := cb.store.GetSecretKey(ctx, crKey.Namespace, vv)
				if err != nil {
					return out, fmt.Errorf("failed to get proxyConnectHeader secretKey: %w", err)
				}
				proxyConnectHeader[k] = append(proxyConnectHeader[k], value)
			}
		}
		out.ProxyConnectHeader = proxyConnectHeader
	}

	return out, nil
}

func (cb *ConfigBuilder) convertGlobalTelegramConfig(out *globalConfig, in *monitoringv1.GlobalTelegramConfig) error {
	if in == nil {
		return nil
	}

	if cb.amVersion.LT(semver.MustParse("0.24.0")) {
		return fmt.Errorf("telegram integration requires Alertmanager >= 0.24.0")
	}

	if in.APIURL != nil {
		u, err := url.Parse(string(*in.APIURL))
		if err != nil {
			return fmt.Errorf("failed to parse Telegram API URL: %w", err)
		}
		out.TelegramAPIURL = &config.URL{URL: u}
	}

	return nil
}

func (cb *ConfigBuilder) convertGlobalJiraConfig(out *globalConfig, in *monitoringv1.GlobalJiraConfig) error {
	if in == nil {
		return nil
	}

	if cb.amVersion.LT(semver.MustParse("0.28.0")) {
		return errors.New("jira integration requires Alertmanager >= 0.28.0")
	}

	if in.APIURL != nil {
		u, err := url.Parse(string(*in.APIURL))
		if err != nil {
			return fmt.Errorf("failed to parse Jira API URL: %w", err)
		}
		out.JiraAPIURL = &config.URL{URL: u}
	}

	return nil
}

func (cb *ConfigBuilder) convertGlobalRocketChatConfig(ctx context.Context, out *globalConfig, in *monitoringv1.GlobalRocketChatConfig, crKey types.NamespacedName) error {
	if in == nil {
		return nil
	}

	if cb.amVersion.LT(semver.MustParse("0.28.0")) {
		return errors.New("rocket chat integration requires Alertmanager >= 0.28.0")
	}

	if in.APIURL != nil {
		u, err := url.Parse(string(*in.APIURL))
		if err != nil {
			return fmt.Errorf("failed to parse Rocket Chat API URL: %w", err)
		}
		out.RocketChatAPIURL = &config.URL{URL: u}
	}

	if in.Token != nil {
		token, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.Token)
		if err != nil {
			return fmt.Errorf("failed to get Rocket Chat Token: %w", err)
		}
		out.RocketChatToken = token
	}

	if in.TokenID != nil {
		tokenID, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.TokenID)
		if err != nil {
			return fmt.Errorf("failed to get Rocket Chat Token ID: %w", err)
		}
		out.RocketChatTokenID = tokenID
	}

	return nil
}

func (cb *ConfigBuilder) convertGlobalWebexConfig(out *globalConfig, in *monitoringv1.GlobalWebexConfig) error {
	if in == nil {
		return nil
	}

	if cb.amVersion.LT(semver.MustParse("0.25.0")) {
		return fmt.Errorf(`webex integration requires Alertmanager >= 0.25.0`)
	}

	if in.APIURL != nil {
		u, err := url.Parse(string(*in.APIURL))
		if err != nil {
			return fmt.Errorf("parse Webex API URL: %w", err)
		}
		out.WebexAPIURL = &config.URL{URL: u}
	}

	return nil
}

func (cb *ConfigBuilder) convertGlobalWeChatConfig(ctx context.Context, out *globalConfig, in *monitoringv1.GlobalWeChatConfig, crKey types.NamespacedName) error {
	if in == nil {
		return nil
	}

	if in.APIURL != nil {
		u, err := url.Parse(string(*in.APIURL))
		if err != nil {
			return fmt.Errorf("wechat API URL: %w", err)
		}
		out.WeChatAPIURL = &config.URL{URL: u}
	}

	if in.APISecret != nil {
		apiSecret, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.APISecret)
		if err != nil {
			return fmt.Errorf("failed to get WeChat Secret: %w", err)
		}
		out.WeChatAPISecret = apiSecret
	}

	if in.APICorpID != nil {
		out.WeChatAPICorpID = *in.APICorpID
	}

	return nil
}

func (cb *ConfigBuilder) convertGlobalVictorOpsConfig(ctx context.Context, out *globalConfig, in *monitoringv1.GlobalVictorOpsConfig, crKey types.NamespacedName) error {
	if in == nil {
		return nil
	}

	if in.APIURL != nil {
		u, err := url.Parse(string(*in.APIURL))
		if err != nil {
			return fmt.Errorf("failed to parse VictorOps API URL: %w", err)
		}
		out.VictorOpsAPIURL = &config.URL{URL: u}
	}

	if in.APIKey != nil {
		apiSecret, err := cb.store.GetSecretKey(ctx, crKey.Namespace, *in.APIKey)
		if err != nil {
			return fmt.Errorf("failed to get VictorOps Secret: %w", err)
		}
		out.VictorOpsAPIKey = apiSecret
	}

	return nil
}

// sanitize the config against a specific Alertmanager version
// types may be sanitized in one of two ways:
// 1. stripping the unsupported config and log a warning
// 2. error which ensures that config will not be reconciled - this will be logged by a calling function.
func (c *alertmanagerConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if c == nil {
		return nil
	}

	if err := c.Global.sanitize(amVersion, logger); err != nil {
		return err
	}

	for _, receiver := range c.Receivers {
		if err := receiver.sanitize(amVersion, logger); err != nil {
			return err
		}
	}

	for i, rule := range c.InhibitRules {
		if err := rule.sanitize(amVersion, logger); err != nil {
			return fmt.Errorf("inhibit_rules[%d]: %w", i, err)
		}
	}

	if len(c.MuteTimeIntervals) > 0 && amVersion.LT(semver.MustParse("0.22.0")) {
		// mute time intervals are unsupported < 0.22.0, and we already log the situation
		// when handling the routes so just set to nil
		c.MuteTimeIntervals = nil
	}

	if len(c.TimeIntervals) > 0 && amVersion.LT(semver.MustParse("0.24.0")) {
		// time intervals are unsupported < 0.24.0, and we already log the situation
		// when handling the routes so just set to nil
		c.TimeIntervals = nil
	}

	for _, ti := range c.MuteTimeIntervals {
		if err := ti.sanitize(amVersion, logger); err != nil {
			return fmt.Errorf("mute_time_intervals[%s]: %w", ti.Name, err)
		}
	}

	for _, ti := range c.TimeIntervals {
		if err := ti.sanitize(amVersion, logger); err != nil {
			return fmt.Errorf("time_intervals[%s]: %w", ti.Name, err)
		}
	}

	return c.Route.sanitize(amVersion, logger)
}

// sanitize globalConfig.
func (gc *globalConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if gc == nil {
		return nil
	}

	if gc.HTTPConfig != nil {
		if err := gc.HTTPConfig.sanitize(amVersion, logger); err != nil {
			return err
		}
	}

	if gc.SMTPTLSConfig != nil && amVersion.LT(semver.MustParse("0.28.0")) {
		msg := "'smtp_tls_config' supported in Alertmanager >= 0.28.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		gc.SMTPTLSConfig = nil
	}

	// We need to sanitize the config for slack globally
	// As of v0.22.0 Alertmanager config supports passing URL via file name
	if gc.SlackAPIURLFile != "" {
		if gc.SlackAPIURL != nil {
			msg := "'slack_api_url' and 'slack_api_url_file' are mutually exclusive - 'slack_api_url' has taken precedence"
			logger.Warn(msg)
			gc.SlackAPIURLFile = ""
		}

		if amVersion.LT(semver.MustParse("0.22.0")) {
			msg := "'slack_api_url_file' supported in Alertmanager >= 0.22.0 only - dropping field from provided config"
			logger.Warn(msg, "current_version", amVersion.String())
			gc.SlackAPIURLFile = ""
		}
	}

	if gc.SlackAppToken != "" && amVersion.LT(semver.MustParse("0.30.0")) {
		msg := "'slack_app_token' supported in Alertmanager >= 0.30.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		gc.SlackAppToken = ""
	}

	if gc.SlackAppTokenFile != "" && amVersion.LT(semver.MustParse("0.30.0")) {
		msg := "'slack_app_token_file' supported in Alertmanager >= 0.30.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		gc.SlackAppTokenFile = ""
	}

	if gc.SlackAppURL != nil && amVersion.LT(semver.MustParse("0.30.0")) {
		msg := "'slack_app_url' supported in Alertmanager >= 0.30.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		gc.SlackAppURL = nil
	}

	if gc.SlackAppToken != "" && gc.SlackAppTokenFile != "" {
		msg := "'slack_app_token' and 'slack_app_token_file' are mutually exclusive - 'slack_app_token' has taken precedence"
		logger.Warn(msg)
		gc.SlackAppTokenFile = ""
	}

	if (gc.SlackAppToken != "" || gc.SlackAppTokenFile != "") && (gc.SlackAPIURL != nil || gc.SlackAPIURLFile != "") {
		if gc.SlackAPIURL != nil && gc.SlackAppURL != nil && gc.SlackAPIURL.String() != gc.SlackAppURL.String() {
			return fmt.Errorf("at most one of slack_app_token/slack_app_token_file & slack_api_url/slack_api_url_file must be configured (unless slack_api_url matches slack_app_url)")
		}
	}

	if gc.OpsGenieAPIKeyFile != "" && amVersion.LT(semver.MustParse("0.24.0")) {
		msg := "'opsgenie_api_key_file' supported in Alertmanager >= 0.24.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		gc.OpsGenieAPIKeyFile = ""
	}

	if gc.SMTPAuthPasswordFile != "" && amVersion.LT(semver.MustParse("0.25.0")) {
		msg := "'smtp_auth_password_file' supported in Alertmanager >= 0.25.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		gc.SMTPAuthPasswordFile = ""
	}

	if gc.SMTPAuthPassword != "" && gc.SMTPAuthPasswordFile != "" {
		msg := "'smtp_auth_password' and 'smtp_auth_password_file' are mutually exclusive - 'smtp_auth_password' has taken precedence"
		logger.Warn(msg)
		gc.SMTPAuthPasswordFile = ""
	}

	if gc.VictorOpsAPIKeyFile != "" && amVersion.LT(semver.MustParse("0.25.0")) {
		msg := "'victorops_api_key_file' supported in Alertmanager >= 0.25.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		gc.VictorOpsAPIKeyFile = ""
	}

	if gc.VictorOpsAPIKey != "" && gc.VictorOpsAPIKeyFile != "" {
		msg := "'victorops_api_key' and 'victorops_api_key_file' are mutually exclusive - 'victorops_api_key' has taken precedence"
		logger.Warn(msg)
		gc.VictorOpsAPIKeyFile = ""
	}

	return nil
}

func (hc *httpClientConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if hc == nil {
		return nil
	}

	if hc.Authorization != nil && !amVersion.GTE(semver.MustParse("0.22.0")) {
		return fmt.Errorf("'authorization' set in 'http_config' but supported in Alertmanager >= 0.22.0 only")
	}

	if hc.OAuth2 != nil && !amVersion.GTE(semver.MustParse("0.22.0")) {
		return fmt.Errorf("'oauth2' set in 'http_config' but supported in Alertmanager >= 0.22.0 only")
	}

	if hc.FollowRedirects != nil && !amVersion.GTE(semver.MustParse("0.22.0")) {
		msg := "'follow_redirects' set in 'http_config' but supported in Alertmanager >= 0.22.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		hc.FollowRedirects = nil
	}

	if hc.EnableHTTP2 != nil && !amVersion.GTE(semver.MustParse("0.25.0")) {
		msg := "'enable_http2' set in 'http_config' but supported in Alertmanager >= 0.25.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		hc.EnableHTTP2 = nil
	}

	if err := hc.TLSConfig.sanitize(amVersion, logger); err != nil {
		return err
	}

	if err := hc.proxyConfig.sanitize(amVersion, logger); err != nil {
		return err
	}

	return hc.OAuth2.sanitize(amVersion, logger)
}

var tlsVersions = map[string]int{
	"":      0x0000,
	"TLS13": tls.VersionTLS13,
	"TLS12": tls.VersionTLS12,
	"TLS11": tls.VersionTLS11,
	"TLS10": tls.VersionTLS10,
}

func (tc *tlsConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if tc == nil {
		return nil
	}

	if tc.MinVersion != "" && !amVersion.GTE(semver.MustParse("0.25.0")) {
		msg := "'min_version' set in 'tls_config' but supported in Alertmanager >= 0.25.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		tc.MinVersion = ""
	}

	if tc.MaxVersion != "" && !amVersion.GTE(semver.MustParse("0.25.0")) {
		msg := "'max_version' set in 'tls_config' but supported in Alertmanager >= 0.25.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		tc.MaxVersion = ""
	}

	minVersion, found := tlsVersions[tc.MinVersion]
	if !found {
		return fmt.Errorf("unknown TLS version: %s", tc.MinVersion)
	}

	maxVersion, found := tlsVersions[tc.MaxVersion]
	if !found {
		return fmt.Errorf("unknown TLS version: %s", tc.MaxVersion)
	}

	if minVersion != 0 && maxVersion != 0 && minVersion > maxVersion {
		return fmt.Errorf("max TLS version %q must be greater than or equal to min TLS version %q", tc.MaxVersion, tc.MinVersion)
	}

	return nil
}

func (pc *proxyConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if pc == nil {
		return nil
	}

	// All proxy options are supported starting from v0.26.0. Below this
	// version, only 'proxy_url' is supported.
	if amVersion.GTE(semver.MustParse("0.26.0")) {
		if len(pc.ProxyConnectHeader) > 0 && (!pc.ProxyFromEnvironment && pc.ProxyURL == "") {
			return fmt.Errorf("if 'proxy_connect_header' is configured, 'proxy_url' or 'proxy_from_environment' must also be configured")
		}

		if pc.ProxyFromEnvironment && pc.ProxyURL != "" {
			return fmt.Errorf("if 'proxy_from_environment' is configured, 'proxy_url' must not be configured")
		}

		if pc.ProxyFromEnvironment && pc.NoProxy != "" {
			return fmt.Errorf("if 'proxy_from_environment' is configured, 'no_proxy' must not be configured")
		}

		if pc.ProxyURL == "" && pc.NoProxy != "" {
			return fmt.Errorf("if 'no_proxy' is configured, 'proxy_url' must also be configured")
		}

		return nil
	}

	if pc.ProxyFromEnvironment {
		msg := "'proxy_from_environment' set to true but supported in Alertmanager >= 0.26.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		pc.ProxyFromEnvironment = false
	}

	if pc.NoProxy != "" {
		msg := "'no_proxy' configured but supported in Alertmanager >= 0.26.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		pc.NoProxy = ""
	}

	if len(pc.ProxyConnectHeader) > 0 {
		msg := "'proxy_connect_header' configured but supported in Alertmanager >= 0.26.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		pc.ProxyConnectHeader = nil
	}

	return nil
}

func (o *oauth2) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if o == nil {
		return nil
	}

	if (o.ProxyURL != "" || o.NoProxy != "" || len(o.ProxyConnectHeader) > 0) &&
		!amVersion.GTE(semver.MustParse("0.25.0")) {
		msg := "'proxyConfig' set in 'oauth2' but supported in Alertmanager >= 0.25.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		o.ProxyURL = ""
		o.NoProxy = ""
		o.ProxyFromEnvironment = false
		o.ProxyConnectHeader = nil
	}

	return nil
}

// sanitize the receiver.
func (r *receiver) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if r == nil {
		return nil
	}
	withLogger := logger.With("receiver", r.Name)

	for _, conf := range r.EmailConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.OpsgenieConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.PagerdutyConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.PushoverConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.SlackConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.VictorOpsConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.WebhookConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.WeChatConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.SNSConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.TelegramConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.DiscordConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.WebexConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.MSTeamsConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.MSTeamsV2Configs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.JiraConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.RocketChatConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.MattermostConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	for _, conf := range r.IncidentioConfigs {
		if err := conf.sanitize(amVersion, withLogger); err != nil {
			return err
		}
	}

	return nil
}

func (ec *emailConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if ec.AuthPasswordFile != "" && amVersion.LT(semver.MustParse("0.25.0")) {
		msg := "'auth_password_file' supported in Alertmanager >= 0.25.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		ec.AuthPasswordFile = ""
	}

	if ec.AuthPassword != "" && ec.AuthPasswordFile != "" {
		logger.Warn("'auth_password' and 'auth_password_file' are mutually exclusive for email receiver config - 'auth_password' has taken precedence")
		ec.AuthPasswordFile = ""
	}

	return nil
}

func (ogc *opsgenieConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if err := ogc.HTTPConfig.sanitize(amVersion, logger); err != nil {
		return err
	}

	lessThanV0_24 := amVersion.LT(semver.MustParse("0.24.0"))

	if ogc.Actions != "" && lessThanV0_24 {
		msg := "opsgenie_config 'actions' supported in Alertmanager >= 0.24.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		ogc.Actions = ""
	}

	if ogc.Entity != "" && lessThanV0_24 {
		msg := "opsgenie_config 'entity' supported in Alertmanager >= 0.24.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		ogc.Entity = ""
	}
	if ogc.UpdateAlerts != nil && lessThanV0_24 {
		msg := "update_alerts 'entity' supported in Alertmanager >= 0.24.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		ogc.UpdateAlerts = nil
	}
	for _, responder := range ogc.Responders {
		if err := responder.sanitize(amVersion); err != nil {
			return err
		}
	}

	if ogc.APIKey != "" && ogc.APIKeyFile != "" {
		logger.Warn("'api_key' and 'api_key_file' are mutually exclusive for OpsGenie receiver config - 'api_key' has taken precedence")
		ogc.APIKeyFile = ""
	}

	if ogc.APIKeyFile == "" {
		return nil
	}

	if lessThanV0_24 {
		msg := "'api_key_file' supported in Alertmanager >= 0.24.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		ogc.APIKeyFile = ""
	}

	return nil
}

func (ops *opsgenieResponder) sanitize(amVersion semver.Version) error {
	if ops.Type == "teams" && amVersion.LT(semver.MustParse("0.24.0")) {
		return fmt.Errorf("'teams' set in 'opsgenieResponder' but supported in Alertmanager >= 0.24.0 only")
	}
	return nil
}

func (pdc *pagerdutyConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	lessThanV0_25 := amVersion.LT(semver.MustParse("0.25.0"))
	lessThanV0_30 := amVersion.LT(semver.MustParse("0.30.0"))

	if pdc.Source != "" && lessThanV0_25 {
		msg := "'source' supported in Alertmanager >= 0.25.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		pdc.Source = ""
	}

	if pdc.RoutingKeyFile != "" && lessThanV0_25 {
		msg := "'routing_key_file' supported in Alertmanager >= 0.25.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		pdc.RoutingKeyFile = ""
	}

	if pdc.ServiceKeyFile != "" && lessThanV0_25 {
		msg := "'service_key_file' supported in Alertmanager >= 0.25.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		pdc.ServiceKeyFile = ""
	}

	if pdc.ServiceKey != "" && pdc.ServiceKeyFile != "" {
		msg := "'service_key' and 'service_key_file' are mutually exclusive for pagerdury receiver config - 'service_key' has taken precedence"
		logger.Warn(msg)
		pdc.ServiceKeyFile = ""
	}

	if pdc.RoutingKey != "" && pdc.RoutingKeyFile != "" {
		msg := "'routing_key' and 'routing_key_file' are mutually exclusive for pagerdury receiver config - 'routing_key' has taken precedence"
		logger.Warn(msg)
		pdc.RoutingKeyFile = ""
	}

	if pdc.Timeout != nil && lessThanV0_30 {
		msg := "'timeout' supported in Alertmanager >= 0.30.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		pdc.Timeout = nil
	}

	return pdc.HTTPConfig.sanitize(amVersion, logger)
}

func (poc *pushoverConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	lessThanV0_26 := amVersion.LT(semver.MustParse("0.26.0"))
	lessThanV0_27 := amVersion.LT(semver.MustParse("0.27.0"))
	lessThanV0_29 := amVersion.LT(semver.MustParse("0.29.0"))

	if poc.UserKeyFile != "" && lessThanV0_26 {
		msg := "'user_key_file' supported in Alertmanager >= 0.26.0 only - dropping field from pushover receiver config"
		logger.Warn(msg, "current_version", amVersion.String())
		poc.UserKeyFile = ""
	}

	if poc.UserKey == "" && poc.UserKeyFile == "" {
		return fmt.Errorf("missing mandatory field user_key or user_key_file")
	}

	if poc.UserKey != "" && poc.UserKeyFile != "" {
		msg := "'user_key' and 'user_key_file' are mutually exclusive for pushover receiver config - 'user_key' has taken precedence"
		logger.Warn(msg)
		poc.UserKeyFile = ""
	}

	if poc.TokenFile != "" && lessThanV0_26 {
		msg := "'token_file' supported in Alertmanager >= 0.26.0 only - dropping field from pushover receiver config"
		logger.Warn(msg, "current_version", amVersion.String())
		poc.TokenFile = ""
	}

	if poc.Token == "" && poc.TokenFile == "" {
		return fmt.Errorf("missing mandatory field token or token_file")
	}

	if poc.Token != "" && poc.TokenFile != "" {
		msg := "'token' and 'token_file' are mutually exclusive for pushover receiver config - 'token' has taken precedence"
		logger.Warn(msg)
		poc.TokenFile = ""
	}

	if poc.TTL != "" && lessThanV0_27 {
		msg := "'ttl' supported in Alertmanager >= 0.27.0 only - dropping field from pushover receiver config"
		logger.Warn(msg, "current_version", amVersion.String())
		poc.TTL = ""
	}

	if poc.Device != "" && lessThanV0_26 {
		msg := "'device' supported in Alertmanager >= 0.26.0 only - dropping field from pushover receiver config"
		logger.Warn(msg, "current_version", amVersion.String())
		poc.Device = ""
	}

	if poc.Monospace != nil && *poc.Monospace && lessThanV0_29 {
		msg := "'monospace' supported in Alertmanager >= 0.29.0 only - dropping field from pushover receiver config"
		logger.Warn(msg, "current_version", amVersion.String())
		*poc.Monospace = false
	}

	if poc.HTML != nil && *poc.HTML && poc.Monospace != nil && *poc.Monospace {
		return errors.New("either monospace or html must be configured")
	}

	if poc.URL != "" {
		if _, err := validation.ValidateURL(poc.URL); err != nil {
			return fmt.Errorf("invalid 'url': %w", err)
		}
	}

	return poc.HTTPConfig.sanitize(amVersion, logger)
}

func (sc *slackConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	lessThanV0_30 := amVersion.LT(semver.MustParse("0.30.0"))

	if err := sc.HTTPConfig.sanitize(amVersion, logger); err != nil {
		return err
	}

	if sc.Timeout != nil && lessThanV0_30 {
		msg := "'timeout' supported in Alertmanager >= 0.30.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		sc.Timeout = nil
	}

	if sc.AppToken != "" && lessThanV0_30 {
		msg := "'app_token' supported in Alertmanager >= 0.30.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		sc.AppToken = ""
	}

	if sc.AppTokenFile != "" && lessThanV0_30 {
		msg := "'app_token_file' supported in Alertmanager >= 0.30.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		sc.AppTokenFile = ""
	}

	if sc.AppURL != "" && lessThanV0_30 {
		msg := "'app_url' supported in Alertmanager >= 0.30.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		sc.AppURL = ""
	}

	if sc.AppToken != "" && sc.AppTokenFile != "" {
		msg := "'app_token' and 'app_token_file' are mutually exclusive for slack receiver config - 'app_token' has taken precedence"
		logger.Warn(msg)
		sc.AppTokenFile = ""
	}

	if (sc.AppToken != "" || sc.AppTokenFile != "") && (sc.APIURL != "" || sc.APIURLFile != "") {
		return fmt.Errorf("at most one of app_token/app_token_file & api_url/api_url_file must be configured")
	}

	if sc.APIURLFile == "" {
		return nil
	}

	// We need to sanitize the config for slack receivers
	// As of v0.22.0 Alertmanager config supports passing URL via file name
	if sc.APIURLFile != "" && amVersion.LT(semver.MustParse("0.22.0")) {
		msg := "'api_url_file' supported in Alertmanager >= 0.22.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		sc.APIURLFile = ""
	}

	if sc.APIURL != "" && sc.APIURLFile != "" {
		msg := "'api_url' and 'api_url_file' are mutually exclusive for slack receiver config - 'api_url' has taken precedence"
		logger.Warn(msg)
		sc.APIURLFile = ""
	}

	return nil
}

func (voc *victorOpsConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if err := voc.HTTPConfig.sanitize(amVersion, logger); err != nil {
		return err
	}

	if voc.APIKeyFile != "" && amVersion.LT(semver.MustParse("0.25.0")) {
		msg := "'api_key_file' supported in Alertmanager >= 0.25.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		voc.APIKeyFile = ""
	}

	if voc.APIKey != "" && voc.APIKeyFile != "" {
		msg := "'api_key' and 'api_key_file' are mutually exclusive for victorops receiver config - 'api_url' has taken precedence"
		logger.Warn(msg)
		voc.APIKeyFile = ""
	}

	return nil
}

func (whc *webhookConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if err := whc.HTTPConfig.sanitize(amVersion, logger); err != nil {
		return err
	}

	if whc.URLFile != "" && amVersion.LT(semver.MustParse("0.26.0")) {
		msg := "'url_file' supported in Alertmanager >= 0.26.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		whc.URLFile = ""
	}

	if whc.URL != "" && whc.URLFile != "" {
		msg := "'url' and 'url_file' are mutually exclusive for webhook receiver config - 'url' has taken precedence"
		logger.Warn(msg)
		whc.URLFile = ""
	}

	if whc.Timeout != nil && amVersion.LT(semver.MustParse("0.28.0")) {
		msg := "'timeout' supported in Alertmanager >= 0.28.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		whc.Timeout = nil
	}

	if whc.URL != "" {
		if _, err := validation.ValidateURL(whc.URL); err != nil {
			return fmt.Errorf("invalid 'url': %w", err)
		}
	}

	return nil
}

func (tc *msTeamsConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if amVersion.LT(semver.MustParse("0.26.0")) {
		return fmt.Errorf(`invalid syntax in receivers config; msteams integration is only available in Alertmanager >= 0.26.0`)
	}

	if tc.WebhookURL == "" {
		return fmt.Errorf("mandatory field %q is empty", "webhook_url")
	}

	if _, err := validation.ValidateURL(tc.WebhookURL); err != nil {
		return fmt.Errorf("invalid 'webhook_url': %w", err)
	}

	if tc.Summary != "" && amVersion.LT(semver.MustParse("0.27.0")) {
		msg := "'summary' supported in Alertmanager >= 0.27.0 only - dropping field `summary` from msteams config"
		logger.Warn(msg, "current_version", amVersion.String())
		tc.Summary = ""
	}

	return tc.HTTPConfig.sanitize(amVersion, logger)
}

func (tc *msTeamsV2Config) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	msTeamsV2Allowed := amVersion.GTE(semver.MustParse("0.28.0"))
	if !msTeamsV2Allowed {
		return fmt.Errorf(`invalid syntax in receivers config; msteams v2 integration is available in Alertmanager >= 0.28.0`)
	}

	if tc.WebhookURL == "" && len(tc.WebhookURLFile) == 0 {
		return errors.New("no webhook_url or webhook_url_file provided")
	}

	if tc.WebhookURL != "" && len(tc.WebhookURLFile) != 0 {
		return errors.New("both webhook_url and webhook_url_file cannot be set at the same time")
	}

	return tc.HTTPConfig.sanitize(amVersion, logger)
}

func (wcc *weChatConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	return wcc.HTTPConfig.sanitize(amVersion, logger)
}

func (sc *snsConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if sc.APIUrl != "" {
		if _, err := validation.ValidateURL(sc.APIUrl); err != nil {
			return fmt.Errorf("invalid 'api_url': %w", err)
		}
	}

	return sc.HTTPConfig.sanitize(amVersion, logger)
}

func (tc *telegramConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	lessThanV0_26 := amVersion.LT(semver.MustParse("0.26.0"))
	telegramAllowed := amVersion.GTE(semver.MustParse("0.24.0"))

	if !telegramAllowed {
		return fmt.Errorf(`invalid syntax in receivers config; telegram integration is available in Alertmanager >= 0.24.0`)
	}

	if tc.ChatID == 0 {
		return fmt.Errorf("mandatory field %q is empty", "chatID")
	}

	if tc.BotTokenFile != "" && lessThanV0_26 {
		msg := "'bot_token_file' supported in Alertmanager >= 0.26.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		tc.BotTokenFile = ""
	}

	if tc.BotToken == "" && tc.BotTokenFile == "" {
		return fmt.Errorf("missing mandatory field botToken or botTokenFile")
	}

	if tc.BotToken != "" && tc.BotTokenFile != "" {
		msg := "'bot_token' and 'bot_token_file' are mutually exclusive for telegram receiver config - 'bot_token' has taken precedence"
		logger.Warn(msg)
		tc.BotTokenFile = ""
	}

	if tc.MessageThreadID != 0 && lessThanV0_26 {
		msg := "'message_thread_id' supported in Alertmanager >= 0.26.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		tc.MessageThreadID = 0
	}

	if tc.APIUrl != "" {
		if _, err := validation.ValidateURL(tc.APIUrl); err != nil {
			return fmt.Errorf("invalid 'api_url': %w", err)
		}
	}

	return tc.HTTPConfig.sanitize(amVersion, logger)
}

func (dc *discordConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	discordAllowed := amVersion.GTE(semver.MustParse("0.25.0"))
	lessThanV0_28 := amVersion.LT(semver.MustParse("0.28.0"))

	if !discordAllowed {
		return fmt.Errorf(`invalid syntax in receivers config; discord integration is available in Alertmanager >= 0.25.0`)
	}

	if dc.Content != "" && lessThanV0_28 {
		msg := "'content' supported in Alertmanager >= 0.28.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		dc.Content = ""
	}

	if dc.Username != "" && lessThanV0_28 {
		msg := "'username' supported in Alertmanager >= 0.28.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		dc.Username = ""
	}

	if dc.AvatarURL != "" && lessThanV0_28 {
		msg := "'avatar_url' supported in Alertmanager >= 0.28.0 only - dropping field from provided config"
		logger.Warn(msg, "current_version", amVersion.String())
		dc.AvatarURL = ""
	}

	return dc.HTTPConfig.sanitize(amVersion, logger)
}

func (tc *webexConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	webexAllowed := amVersion.GTE(semver.MustParse("0.25.0"))
	if !webexAllowed {
		return fmt.Errorf(`invalid syntax in receivers config; webex integration is available in Alertmanager >= 0.25.0`)
	}

	if tc.RoomID == "" {
		return fmt.Errorf("mandatory field %q is empty", "room_id")
	}

	return tc.HTTPConfig.sanitize(amVersion, logger)
}

func (jc *jiraConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	jiraConfigAllowed := amVersion.GTE(semver.MustParse("0.28.0"))
	if !jiraConfigAllowed {
		return fmt.Errorf(`invalid syntax in receivers config; jira integration is available in Alertmanager >= 0.28.0`)
	}

	if jc.Project == "" {
		return fmt.Errorf("missing project in jira_config")
	}
	if jc.IssueType == "" {
		return errors.New("missing issue_type in jira_config")
	}

	apiTypeAllowed := amVersion.GTE(semver.MustParse("0.29.0"))
	if jc.APIType != "" {
		if !apiTypeAllowed {
			msg := "'api_type' supported in Alertmanager >= 0.29.0 only - dropping field from provided config"
			logger.Warn(msg, "current_version", amVersion.String())
			jc.APIType = ""
		} else {
			if jc.APIType != "auto" && jc.APIType != "cloud" && jc.APIType != "datacenter" {
				return fmt.Errorf("invalid 'api_type': a value must be 'auto', 'cloud' or 'datacenter'")
			}
		}
	}

	return jc.HTTPConfig.sanitize(amVersion, logger)
}

func (rc *rocketChatConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	rocketChatAllowed := amVersion.GTE(semver.MustParse("0.28.0"))
	if !rocketChatAllowed {
		return fmt.Errorf(`invalid syntax in receivers config; rocketchat integration is available in Alertmanager >= 0.28.0`)
	}

	if rc.Token != nil && len(rc.TokenFile) > 0 {
		return fmt.Errorf("at most one of token & token_file must be configured")
	}
	if rc.TokenID != nil && len(rc.TokenIDFile) > 0 {
		return fmt.Errorf("at most one of token_id & token_id_file must be configured")
	}

	return rc.HTTPConfig.sanitize(amVersion, logger)
}

func (mc *mattermostConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	mattermostAllowed := amVersion.GTE(semver.MustParse("0.30.0"))
	if !mattermostAllowed {
		return fmt.Errorf(`invalid syntax in receivers config; mattermost integration is available in Alertmanager >= 0.30.0`)
	}

	if mc.WebhookURL == "" && mc.WebhookURLFile == "" {
		return fmt.Errorf(`one of 'webhook_url' or 'webhook_url_file' must be configured`)
	}

	if mc.WebhookURL != "" && mc.WebhookURLFile != "" {
		msg := "'webhook_url' and 'webhook_url_file' are mutually exclusive for mattermost receiver config - 'webhook_url' has taken precedence"
		logger.Warn(msg)
		mc.WebhookURLFile = ""
	}

	return mc.HTTPConfig.sanitize(amVersion, logger)
}

func (ic *incidentioConfig) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	incidentioAllowed := amVersion.GTE(semver.MustParse("0.29.0"))
	if !incidentioAllowed {
		return fmt.Errorf("invalid syntax in receivers config; incident.io integration is available in Alertmanager >= 0.29.0")
	}

	if ic.URL == "" && ic.URLFile == "" {
		return errors.New("one of url or url_file must be configured")
	}

	if ic.URL != "" && ic.URLFile != "" {
		return errors.New("at most one of url & url_file must be configured")
	}

	if ic.URL != "" {
		if _, err := validation.ValidateURL(ic.URL); err != nil {
			return fmt.Errorf("invalid url: %w", err)
		}
	}

	if ic.AlertSourceToken != "" && ic.AlertSourceTokenFile != "" {
		return errors.New("at most one of alert_source_token & alert_source_token_file must be configured")
	}

	if ic.HTTPConfig != nil && ic.HTTPConfig.Authorization != nil && (ic.AlertSourceToken != "" || ic.AlertSourceTokenFile != "") {
		return errors.New("cannot specify alert_source_token or alert_source_token_file when using http_config.authorization")
	}

	if ic.AlertSourceToken == "" && ic.AlertSourceTokenFile == "" && (ic.HTTPConfig == nil || ic.HTTPConfig.Authorization == nil) {
		return errors.New("at least one of alert_source_token, alert_source_token_file or http_config.authorization must be configured")
	}

	return ic.HTTPConfig.sanitize(amVersion, logger)
}

func (ir *inhibitRule) sanitize(amVersion semver.Version, logger *slog.Logger) error {
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
		logger.Warn(msg, "source_match", ir.SourceMatch, "target_match", ir.TargetMatch, "source_match_re", ir.SourceMatchRE, "target_match_re", ir.TargetMatchRE)
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

func (ti *timeInterval) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if amVersion.GTE(semver.MustParse("0.25.0")) {
		return nil
	}

	for i, tis := range ti.TimeIntervals {
		if tis.Location != nil {
			logger.Warn("time_interval location is supported in Alertmanager >= 0.25.0 only - dropping config")
			ti.TimeIntervals[i].Location = nil
		}
	}

	return nil
}

// sanitize a route and all its child routes.
// Warns if the config is using deprecated syntax against a later version.
// Returns an error if the config could potentially break routing logic.
func (r *route) sanitize(amVersion semver.Version, logger *slog.Logger) error {
	if r == nil {
		return nil
	}

	matchersV2Allowed := amVersion.GTE(semver.MustParse("0.22.0"))
	muteTimeIntervalsAllowed := matchersV2Allowed
	activeTimeIntervalsAllowed := amVersion.GTE(semver.MustParse("0.24.0"))
	withLogger := logger.With("receiver", r.Receiver)

	if !matchersV2Allowed && checkNotEmptyStrSlice(r.Matchers) {
		return fmt.Errorf(`invalid syntax in route config for 'matchers' comparison based matching is supported in Alertmanager >= 0.22.0 only (matchers=%v)`, r.Matchers)
	}

	if matchersV2Allowed && checkNotEmptyMap(r.Match, r.MatchRE) {
		msg := "'matchers' field is using a deprecated syntax which will be removed in future versions"
		withLogger.Warn(msg, "match", fmt.Sprint(r.Match), "match_re", fmt.Sprint(r.MatchRE))
	}

	if !muteTimeIntervalsAllowed {
		msg := "named mute time intervals in route is supported in Alertmanager >= 0.22.0 only - dropping config"
		withLogger.Warn(msg, "mute_time_intervals", fmt.Sprint(r.MuteTimeIntervals))
		r.MuteTimeIntervals = nil
	}

	if !activeTimeIntervalsAllowed {
		msg := "active time intervals in route is supported in Alertmanager >= 0.24.0 only - dropping config"
		withLogger.Warn(msg, "active_time_intervals", fmt.Sprint(r.ActiveTimeIntervals))
		r.ActiveTimeIntervals = nil
	}

	for i, child := range r.Routes {
		if err := child.sanitize(amVersion, logger); err != nil {
			return fmt.Errorf("route[%d]: %w", i, err)
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
// is equal to the provided value with all whitespace removed.
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

func (cb *ConfigBuilder) checkAlertmanagerGlobalConfigResource(
	ctx context.Context,
	gc *monitoringv1.AlertmanagerGlobalConfig,
	namespace string,
) error {
	if gc == nil {
		return nil
	}

	// Perform semantic validation irrespective of the Alertmanager version.
	if err := validationv1.ValidateAlertmanagerGlobalConfig(gc); err != nil {
		return err
	}

	// Perform more specific validations which depend on the Alertmanager
	// version. It also retrieves data from referenced secrets and configmaps
	// (and fails in case of missing/invalid references).
	if err := cb.checkGlobalWeChatConfig(ctx, gc.WeChatConfig, namespace); err != nil {
		return err
	}

	return nil
}

func (cb *ConfigBuilder) checkGlobalWeChatConfig(
	ctx context.Context,
	wc *monitoringv1.GlobalWeChatConfig,
	namespace string,
) error {
	if wc == nil {
		return nil
	}

	if wc.APISecret != nil {
		if _, err := cb.store.GetSecretKey(ctx, namespace, *wc.APISecret); err != nil {
			return err
		}
	}

	return nil
}
