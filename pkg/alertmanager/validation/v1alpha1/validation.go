// Copyright The prometheus-operator Authors
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
	"errors"
	"fmt"
	"net"
	"strings"

	"k8s.io/utils/ptr"

	"github.com/prometheus-operator/prometheus-operator/pkg/alertmanager/validation"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

// ValidateAlertmanagerConfig checks that the given resource complies with the
// semantics of the Alertmanager configuration.
// In particular, it verifies things that can't be modelized with the OpenAPI
// specification such as routes should refer to an existing receiver.
func ValidateAlertmanagerConfig(amc *monitoringv1alpha1.AlertmanagerConfig) error {
	receivers, err := validateReceivers(amc.Spec.Receivers)
	if err != nil {
		return err
	}

	muteTimeIntervals, err := validateMuteTimeIntervals(amc.Spec.MuteTimeIntervals)
	if err != nil {
		return err
	}

	return validateRoute(amc.Spec.Route, receivers, muteTimeIntervals, true)
}

func validateReceivers(receivers []monitoringv1alpha1.Receiver) (map[string]struct{}, error) {
	var err error
	receiverNames := make(map[string]struct{})

	for _, receiver := range receivers {
		if _, found := receiverNames[receiver.Name]; found {
			return nil, fmt.Errorf("%q receiver is not unique", receiver.Name)
		}
		receiverNames[receiver.Name] = struct{}{}

		receiverValidationFailedFormat := func(err error) (map[string]struct{}, error) {
			return nil, fmt.Errorf("failed to validate receiver %q: %w", receiver.Name, err)
		}

		if err := validateOpsGenieConfigs(receiver.OpsGenieConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err = validatePagerDutyConfigs(receiver.PagerDutyConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err := validateDiscordConfigs(receiver.DiscordConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err := validateSlackConfigs(receiver.SlackConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err := validateWebhookConfigs(receiver.WebhookConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err := validateWechatConfigs(receiver.WeChatConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err := validateEmailConfig(receiver.EmailConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err := validateVictorOpsConfigs(receiver.VictorOpsConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err := validatePushoverConfigs(receiver.PushoverConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err := validateSNSConfigs(receiver.SNSConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err := validateTelegramConfigs(receiver.TelegramConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err := validateWebexConfigs(receiver.WebexConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err := validateMSTeamsConfigs(receiver.MSTeamsConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err := validateMSTeamsV2Configs(receiver.MSTeamsV2Configs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err := validateRocketchatConfigs(receiver.RocketChatConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}

		if err := validateJiraConfigs(receiver.JiraConfigs); err != nil {
			return receiverValidationFailedFormat(err)
		}
	}

	return receiverNames, nil
}

func validatePagerDutyConfigs(configs []monitoringv1alpha1.PagerDutyConfig) error {
	v := func(conf monitoringv1alpha1.PagerDutyConfig) error {
		if err := validation.ValidateURLPtr((*string)(conf.URL)); err != nil {
			return fmt.Errorf("invalid 'url': %w", err)
		}

		if conf.ClientURL != nil && *conf.ClientURL != "" {
			if err := validation.ValidateTemplateURL(*conf.ClientURL); err != nil {
				return fmt.Errorf("invalid 'clientURL': %w", err)
			}
		}

		if conf.RoutingKey == nil && conf.ServiceKey == nil {
			return errors.New("one of 'routingKey' or 'serviceKey' is required")
		}

		for j, lc := range conf.PagerDutyLinkConfigs {
			if lc.Href != nil && *lc.Href != "" {
				if err := validation.ValidateTemplateURL(*lc.Href); err != nil {
					return fmt.Errorf("'pagerDutyLinkConfigs'[%d]: invalid 'href': %w", j, err)
				}
			}
		}

		for j, ic := range conf.PagerDutyImageConfigs {
			if ic.Href != nil && *ic.Href != "" {
				if err := validation.ValidateTemplateURL(*ic.Href); err != nil {
					return fmt.Errorf("'pagerDutyImageConfigs'[%d]: invalid 'href': %w", j, err)
				}
			}
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'pagerdutyConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validateOpsGenieConfigs(configs []monitoringv1alpha1.OpsGenieConfig) error {
	v := func(conf monitoringv1alpha1.OpsGenieConfig) error {
		if err := conf.Validate(); err != nil {
			return err
		}

		if err := validation.ValidateURLPtr((*string)(conf.APIURL)); err != nil {
			return fmt.Errorf("invalid 'apiURL': %w", err)
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'opsgenieConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validateDiscordConfigs(configs []monitoringv1alpha1.DiscordConfig) error {
	v := func(conf monitoringv1alpha1.DiscordConfig) error {
		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'discordConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validateRocketchatConfigs(configs []monitoringv1alpha1.RocketChatConfig) error {
	v := func(conf monitoringv1alpha1.RocketChatConfig) error {
		if err := validation.ValidateURLPtr((*string)(conf.APIURL)); err != nil {
			return fmt.Errorf("invalid 'apiURL': %w", err)
		}

		if err := validation.ValidateTemplateURLPtr(conf.IconURL); err != nil {
			return fmt.Errorf("invalid 'iconURL': %w", err)
		}

		if err := validation.ValidateTemplateURLPtr(conf.ImageURL); err != nil {
			return fmt.Errorf("invalid 'imageURL': %w", err)
		}

		if err := validation.ValidateTemplateURLPtr(conf.ThumbURL); err != nil {
			return fmt.Errorf("invalid 'thumbURL': %w", err)
		}

		for j, a := range conf.Actions {
			if err := validation.ValidateTemplateURLPtr(a.URL); err != nil {
				return fmt.Errorf("'actions'[%d]: invalid 'url': %w", j, err)
			}
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'rocketchatConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validateSlackConfigs(configs []monitoringv1alpha1.SlackConfig) error {
	v := func(conf monitoringv1alpha1.SlackConfig) error {
		if err := conf.Validate(); err != nil {
			return err
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'slackConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validateWebhookConfigs(configs []monitoringv1alpha1.WebhookConfig) error {
	v := func(conf monitoringv1alpha1.WebhookConfig) error {
		if conf.URL == nil && conf.URLSecret == nil {
			return errors.New("one of 'url' or 'urlSecret' must be specified")
		}

		if err := validation.ValidateTemplateURLPtr(conf.URL); err != nil {
			return fmt.Errorf("invalid 'url': %w", err)
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'webhookConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validateWechatConfigs(configs []monitoringv1alpha1.WeChatConfig) error {
	v := func(conf monitoringv1alpha1.WeChatConfig) error {
		if err := validation.ValidateURLPtr((*string)(conf.APIURL)); err != nil {
			return fmt.Errorf("invalid 'apiURL': %w", err)
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'wechatConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validateEmailConfig(configs []monitoringv1alpha1.EmailConfig) error {
	v := func(conf monitoringv1alpha1.EmailConfig) error {
		if ptr.Deref(conf.To, "") == "" {
			return errors.New("missing 'to' address")
		}

		if ptr.Deref(conf.Smarthost, "") != "" {
			_, _, err := net.SplitHostPort(*conf.Smarthost)
			if err != nil {
				return fmt.Errorf("invalid 'smarthost' %q: %w", *conf.Smarthost, err)
			}
		}

		if conf.Headers != nil {
			// Header names are case-insensitive, check for collisions.
			normalizedHeaders := map[string]struct{}{}
			for _, h := range conf.Headers {
				normalized := strings.ToLower(h.Key)
				if _, ok := normalizedHeaders[normalized]; ok {
					return fmt.Errorf("duplicate header %q", normalized)
				}
				normalizedHeaders[normalized] = struct{}{}
			}
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'emailConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validateVictorOpsConfigs(configs []monitoringv1alpha1.VictorOpsConfig) error {
	v := func(conf monitoringv1alpha1.VictorOpsConfig) error {
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

		if len(conf.CustomFields) > 0 {
			for _, f := range conf.CustomFields {
				if _, ok := reservedFields[f.Key]; ok {
					return fmt.Errorf("usage of reserved word %q is not allowed in custom fields", f.Key)
				}
			}
		}

		if conf.RoutingKey == "" {
			return errors.New("missing 'routingKey' key")
		}

		if err := validation.ValidateURLPtr((*string)(conf.APIURL)); err != nil {
			return fmt.Errorf("invalid 'apiURL': %w", err)
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'victoropsConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validatePushoverConfigs(configs []monitoringv1alpha1.PushoverConfig) error {
	v := func(conf monitoringv1alpha1.PushoverConfig) error {
		if conf.UserKey == nil && conf.UserKeyFile == nil {
			return errors.New("one of 'userKey' or 'userKeyFile' must be configured")
		}

		if conf.Token == nil && conf.TokenFile == nil {
			return errors.New("one of 'token' or 'tokenFile' must be configured")
		}

		if conf.HTML != nil && *conf.HTML && conf.Monospace != nil && *conf.Monospace {
			return errors.New("'html' and 'monospace' options are mutually exclusive")
		}

		if conf.URL != "" {
			if err := validation.ValidateTemplateURL(conf.URL); err != nil {
				return fmt.Errorf("invalid 'url': %w", err)
			}
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'pushoverConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validateSNSConfigs(configs []monitoringv1alpha1.SNSConfig) error {
	v := func(conf monitoringv1alpha1.SNSConfig) error {
		if (ptr.Deref(conf.TargetARN, "") == "") != (ptr.Deref(conf.TopicARN, "") == "") != (ptr.Deref(conf.PhoneNumber, "") == "") {
			return errors.New("must provide one of 'targetARN', 'topicARN', or 'phoneNumber'")
		}

		if conf.ApiURL != nil {
			if err := validation.ValidateTemplateURL(*conf.ApiURL); err != nil {
				return fmt.Errorf("invalid 'apiURL': %w", err)
			}
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'snsConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validateTelegramConfigs(configs []monitoringv1alpha1.TelegramConfig) error {
	v := func(conf monitoringv1alpha1.TelegramConfig) error {
		if conf.BotToken == nil && conf.BotTokenFile == nil {
			return errors.New("mandatory field botToken or botTokenfile is empty")
		}

		if conf.BotToken != nil && conf.BotTokenFile != nil {
			return errors.New("only one of 'botToken' or 'botTokenfile' must be configured")
		}

		if conf.ChatID == 0 {
			return errors.New("mandatory field 'chatID' is empty")
		}

		if err := validation.ValidateURLPtr((*string)(conf.APIURL)); err != nil {
			return fmt.Errorf("invalid 'apiURL': %w", err)
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'telegramConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validateWebexConfigs(configs []monitoringv1alpha1.WebexConfig) error {
	v := func(conf monitoringv1alpha1.WebexConfig) error {
		if err := validation.ValidateURLPtr((*string)(conf.APIURL)); err != nil {
			return fmt.Errorf("invalid 'apiURL': %w", err)
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'webexConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validateMSTeamsConfigs(configs []monitoringv1alpha1.MSTeamsConfig) error {
	v := func(conf monitoringv1alpha1.MSTeamsConfig) error {
		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'msteamsConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validateMSTeamsV2Configs(configs []monitoringv1alpha1.MSTeamsV2Config) error {
	v := func(conf monitoringv1alpha1.MSTeamsV2Config) error {
		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'msteamsv2Configs'[%d]: %w", i, err)
		}
	}

	return nil
}

func validateJiraConfigs(configs []monitoringv1alpha1.JiraConfig) error {
	v := func(conf monitoringv1alpha1.JiraConfig) error {
		if conf.Project == "" {
			return errors.New("invalid 'project': this is a required field")
		}

		if err := validation.ValidateURLPtr((*string)(conf.APIURL)); err != nil {
			return fmt.Errorf("apiURL: %w", err)
		}

		if conf.IssueType == "" {
			return errors.New("invalid 'issueType': this is a required field")
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("'httpConfig': %w", err)
		}

		if err := conf.Validate(); err != nil {
			return err
		}

		return nil
	}

	for i, conf := range configs {
		if err := v(conf); err != nil {
			return fmt.Errorf("'jiraConfigs'[%d]: %w", i, err)
		}
	}

	return nil
}

// validateRoute verifies that the given route and all its children are
// semantically valid.  because of the self-referential issues mentioned in
// https://github.com/kubernetes/kubernetes/issues/62872 it is not currently
// possible to apply OpenAPI validation to a v1alpha1.Route.
func validateRoute(r *monitoringv1alpha1.Route, receivers, muteTimeIntervals map[string]struct{}, topLevelRoute bool) error {
	if r == nil {
		return nil
	}

	if r.Receiver == "" {
		if topLevelRoute {
			return errors.New("root route must define a receiver")
		}
	} else {
		if _, found := receivers[r.Receiver]; !found {
			return fmt.Errorf("receiver %q not found", r.Receiver)
		}
	}

	if groupLen := len(r.GroupBy); groupLen > 0 {
		groupedBy := make(map[string]struct{}, groupLen)
		for _, str := range r.GroupBy {
			if _, found := groupedBy[str]; found {
				return fmt.Errorf("duplicate values not permitted in route 'groupBy': %v", r.GroupBy)
			}
			groupedBy[str] = struct{}{}
		}
		if _, found := groupedBy["..."]; found && groupLen > 1 {
			return fmt.Errorf("'...' must be a sole value in route 'groupBy': %v", r.GroupBy)
		}
	}

	for _, namedMuteTimeInterval := range r.MuteTimeIntervals {
		if _, found := muteTimeIntervals[namedMuteTimeInterval]; !found {
			return fmt.Errorf("mute time interval %q not found", namedMuteTimeInterval)
		}
	}

	for _, namedActiveTimeInterval := range r.ActiveTimeIntervals {
		if _, found := muteTimeIntervals[namedActiveTimeInterval]; !found {
			return fmt.Errorf("time interval %q not found", namedActiveTimeInterval)
		}
	}

	for i, m := range r.Matchers {
		if err := m.Validate(); err != nil {
			return fmt.Errorf("matcher[%d]: %w", i, err)
		}
	}

	// Unmarshal the child routes and validate them recursively.
	children, err := r.ChildRoutes()
	if err != nil {
		return err
	}

	for i := range children {
		if err := validateRoute(&children[i], receivers, muteTimeIntervals, false); err != nil {
			return fmt.Errorf("route[%d]: %w", i, err)
		}
	}

	return nil
}

func validateMuteTimeIntervals(muteTimeIntervals []monitoringv1alpha1.MuteTimeInterval) (map[string]struct{}, error) {
	muteTimeIntervalNames := make(map[string]struct{}, len(muteTimeIntervals))

	for i, mti := range muteTimeIntervals {
		if err := mti.Validate(); err != nil {
			return nil, fmt.Errorf("mute time interval[%d] is invalid: %w", i, err)
		}
		muteTimeIntervalNames[mti.Name] = struct{}{}
	}
	return muteTimeIntervalNames, nil
}
