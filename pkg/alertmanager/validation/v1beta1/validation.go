// Copyright 2021 The prometheus-operator Authors
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

package v1beta1

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"k8s.io/utils/ptr"

	"github.com/prometheus-operator/prometheus-operator/pkg/alertmanager/validation"
	monitoringv1beta1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1beta1"
)

// ValidateAlertmanagerConfig checks that the given resource complies with the
// semantics of the Alertmanager configuration.
// In particular, it verifies things that can't be modelized with the OpenAPI
// specification such as routes should refer to an existing receiver.
func ValidateAlertmanagerConfig(amc *monitoringv1beta1.AlertmanagerConfig) error {
	receivers, err := validateReceivers(amc.Spec.Receivers)
	if err != nil {
		return err
	}

	timeIntervals, err := validateTimeIntervals(amc.Spec.TimeIntervals)
	if err != nil {
		return err
	}

	return validateRoute(amc.Spec.Route, receivers, timeIntervals, true)
}

func validateReceivers(receivers []monitoringv1beta1.Receiver) (map[string]struct{}, error) {
	var err error
	receiverNames := make(map[string]struct{})

	for _, receiver := range receivers {
		if _, found := receiverNames[receiver.Name]; found {
			return nil, fmt.Errorf("%q receiver is not unique", receiver.Name)
		}
		receiverNames[receiver.Name] = struct{}{}

		if err = validatePagerDutyConfigs(receiver.PagerDutyConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'pagerDutyConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validateOpsGenieConfigs(receiver.OpsGenieConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'opsGenieConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validateSlackConfigs(receiver.SlackConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'slackConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validateWebhookConfigs(receiver.WebhookConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'webhookConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validateWechatConfigs(receiver.WeChatConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'weChatConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validateEmailConfig(receiver.EmailConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'emailConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validateVictorOpsConfigs(receiver.VictorOpsConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'victorOpsConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validatePushoverConfigs(receiver.PushoverConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'pushOverConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validateSnsConfigs(receiver.SNSConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'snsConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validateTelegramConfigs(receiver.TelegramConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'telegramConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validateDiscordConfigs(receiver.DiscordConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'discordConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validateWebexConfigs(receiver.WebexConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'webexConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validateMSTeamsConfigs(receiver.MSTeamsConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'msteamsConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validateRocketchatConfigs(receiver.RocketChatConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'rocketchatConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validateMSTeamsV2Configs(receiver.MSTeamsV2Configs); err != nil {
			return nil, fmt.Errorf("failed to validate 'msteamsv2Config' - receiver %s: %w", receiver.Name, err)
		}
	}

	return receiverNames, nil
}

func validatePagerDutyConfigs(configs []monitoringv1beta1.PagerDutyConfig) error {
	for i, conf := range configs {
		if err := validation.ValidateURLPtr((*string)(conf.URL)); err != nil {
			return fmt.Errorf("[%d]: url: %w", i, err)
		}

		if conf.ClientURL != nil && *conf.ClientURL != "" {
			if err := validation.ValidateTemplateURL(*conf.ClientURL); err != nil {
				return fmt.Errorf("[%d]: clientURL: %w", i, err)
			}
		}

		if conf.RoutingKey == nil && conf.ServiceKey == nil {
			return errors.New("one of 'routingKey' or 'serviceKey' is required")
		}

		for j, lc := range conf.PagerDutyLinkConfigs {
			if lc.Href != nil && *lc.Href != "" {
				if err := validation.ValidateTemplateURL(*lc.Href); err != nil {
					return fmt.Errorf("[%d]: pagerDutyLinkConfigs[%d]: href: %w", i, j, err)
				}
			}
		}

		for j, ic := range conf.PagerDutyImageConfigs {
			if ic.Href != nil && *ic.Href != "" {
				if err := validation.ValidateTemplateURL(*ic.Href); err != nil {
					return fmt.Errorf("[%d]: pagerDutyImageConfigs[%d]: href: %w", i, j, err)
				}
			}
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func validateOpsGenieConfigs(configs []monitoringv1beta1.OpsGenieConfig) error {
	for i, config := range configs {
		if err := config.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}

		if err := validation.ValidateURLPtr((*string)(config.APIURL)); err != nil {
			return fmt.Errorf("[%d]: apiURL: %w", i, err)
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func validateSlackConfigs(configs []monitoringv1beta1.SlackConfig) error {
	for _, config := range configs {
		if err := config.Validate(); err != nil {
			return err
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func validateWebhookConfigs(configs []monitoringv1beta1.WebhookConfig) error {
	for i, config := range configs {
		if config.URL == nil && config.URLSecret == nil {
			return fmt.Errorf("[%d]: one of 'url' or 'urlSecret' must be specified", i)
		}

		if err := validation.ValidateTemplateURLPtr(config.URL); err != nil {
			return fmt.Errorf("[%d]: url: %w", i, err)
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func validateWechatConfigs(configs []monitoringv1beta1.WeChatConfig) error {
	for i, config := range configs {
		if err := validation.ValidateURLPtr((*string)(config.APIURL)); err != nil {
			return fmt.Errorf("[%d]: apiURL: %w", i, err)
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func validateEmailConfig(configs []monitoringv1beta1.EmailConfig) error {
	for _, config := range configs {
		if ptr.Deref(config.To, "") == "" {
			return errors.New("missing 'to' address")
		}

		if ptr.Deref(config.Smarthost, "") != "" {
			_, _, err := net.SplitHostPort(*config.Smarthost)
			if err != nil {
				return fmt.Errorf("invalid 'smarthost' %s: %w", *config.Smarthost, err)
			}
		}

		if config.Headers != nil {
			// Header names are case-insensitive, check for collisions.
			normalizedHeaders := map[string]struct{}{}
			for _, v := range config.Headers {
				normalized := strings.ToLower(v.Key)
				if _, ok := normalizedHeaders[normalized]; ok {
					return fmt.Errorf("duplicate header %q", normalized)
				}
				normalizedHeaders[normalized] = struct{}{}
			}
		}
	}
	return nil
}

func validateVictorOpsConfigs(configs []monitoringv1beta1.VictorOpsConfig) error {
	for i, config := range configs {

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

		if len(config.CustomFields) > 0 {
			for _, v := range config.CustomFields {
				if _, ok := reservedFields[v.Key]; ok {
					return fmt.Errorf("usage of reserved word %q is not allowed in custom fields", v.Key)
				}
			}
		}

		if config.RoutingKey == "" {
			return errors.New("missing 'routingKey' key")
		}

		if err := validation.ValidateURLPtr((*string)(config.APIURL)); err != nil {
			return fmt.Errorf("[%d]: apiURL: %w", i, err)
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func validatePushoverConfigs(configs []monitoringv1beta1.PushoverConfig) error {
	for i, config := range configs {
		if config.UserKey == nil && config.UserKeyFile == nil {
			return fmt.Errorf("one of userKey or userKeyFile must be configured")
		}

		if config.Token == nil && config.TokenFile == nil {
			return fmt.Errorf("one of token or tokenFile must be configured")
		}

		if config.HTML != nil && *config.HTML && config.Monospace != nil && *config.Monospace {
			return fmt.Errorf("html and monospace options are mutually exclusive")
		}

		if config.URL != "" {
			if err := validation.ValidateTemplateURL(config.URL); err != nil {
				return fmt.Errorf("[%d]: url: %w", i, err)
			}
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func validateSnsConfigs(configs []monitoringv1beta1.SNSConfig) error {
	for i, config := range configs {
		if (ptr.Deref(config.TargetARN, "") == "") != (ptr.Deref(config.TopicARN, "") == "") != (ptr.Deref(config.PhoneNumber, "") == "") {
			return fmt.Errorf("[%d]: must provide either a targetARN, topicARN, or phoneNumber for SNS config", i)
		}

		if config.ApiURL != nil {
			if err := validation.ValidateTemplateURL(*config.ApiURL); err != nil {
				return fmt.Errorf("[%d]: apiURL: %w", i, err)
			}
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func validateTelegramConfigs(configs []monitoringv1beta1.TelegramConfig) error {
	for i, config := range configs {
		if config.BotToken == nil && config.BotTokenFile == nil {
			return fmt.Errorf("[%d]: mandatory field botToken or botTokenfile is empty", i)
		}

		if config.ChatID == 0 {
			return fmt.Errorf("[%d]: mandatory field %q is empty", i, "chatID")
		}

		if err := validation.ValidateURLPtr((*string)(config.APIURL)); err != nil {
			return fmt.Errorf("[%d]: apiURL: %w", i, err)
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func validateWebexConfigs(configs []monitoringv1beta1.WebexConfig) error {
	for i, config := range configs {
		if err := validation.ValidateURLPtr((*string)(config.APIURL)); err != nil {
			return fmt.Errorf("[%d]: apiURL: %w", i, err)
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func validateDiscordConfigs(configs []monitoringv1beta1.DiscordConfig) error {
	for _, config := range configs {
		if err := config.HTTPConfig.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func validateRocketchatConfigs(configs []monitoringv1beta1.RocketChatConfig) error {
	for i, config := range configs {
		if err := validation.ValidateURLPtr((*string)(config.APIURL)); err != nil {
			return fmt.Errorf("[%d]: apiURL: %w", i, err)
		}

		if err := validation.ValidateTemplateURLPtr(config.IconURL); err != nil {
			return fmt.Errorf("[%d]: invalid 'iconURL': %w", i, err)
		}

		if err := validation.ValidateTemplateURLPtr(config.ImageURL); err != nil {
			return fmt.Errorf("[%d]: invalid 'imageURL': %w", i, err)
		}

		if err := validation.ValidateTemplateURLPtr(config.ThumbURL); err != nil {
			return fmt.Errorf("[%d]: invalid 'thumbURL': %w", i, err)
		}

		for j, a := range config.Actions {
			if err := validation.ValidateTemplateURLPtr(a.URL); err != nil {
				return fmt.Errorf("%d: actions[%d]: invalid 'url': %w", i, j, err)
			}
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func validateMSTeamsConfigs(configs []monitoringv1beta1.MSTeamsConfig) error {
	for _, config := range configs {
		if err := config.HTTPConfig.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func validateMSTeamsV2Configs(configs []monitoringv1beta1.MSTeamsV2Config) error {
	for _, config := range configs {
		if err := config.HTTPConfig.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// validateRoute verifies that the given route and all its children are
// semantically valid.  because of the self-referential issues mentioned in
// https://github.com/kubernetes/kubernetes/issues/62872 it is not currently
// possible to apply OpenAPI validation to a v1beta1.Route.
func validateRoute(r *monitoringv1beta1.Route, receivers, timeIntervals map[string]struct{}, topLevelRoute bool) error {
	if r == nil {
		return nil
	}

	if r.Receiver == "" {
		if topLevelRoute {
			return fmt.Errorf("root route must define a receiver")
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

	for _, namedTimeInterval := range r.MuteTimeIntervals {
		if _, found := timeIntervals[namedTimeInterval]; !found {
			return fmt.Errorf("time interval %q not found", namedTimeInterval)
		}
	}

	for _, namedTimeInterval := range r.ActiveTimeIntervals {
		if _, found := timeIntervals[namedTimeInterval]; !found {
			return fmt.Errorf("time interval %q not found", namedTimeInterval)
		}
	}

	for i, v := range r.Matchers {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("matcher[%d]: %w", i, err)
		}
	}

	// Unmarshal the child routes and validate them recursively.
	children, err := r.ChildRoutes()
	if err != nil {
		return err
	}

	for i := range children {
		if err := validateRoute(&children[i], receivers, timeIntervals, false); err != nil {
			return fmt.Errorf("route[%d]: %w", i, err)
		}
	}

	return nil
}

func validateTimeIntervals(timeIntervals []monitoringv1beta1.TimeInterval) (map[string]struct{}, error) {
	timeIntervalNames := make(map[string]struct{}, len(timeIntervals))

	for i, ti := range timeIntervals {
		if err := ti.Validate(); err != nil {
			return nil, fmt.Errorf("time interval[%d] is invalid: %w", i, err)
		}
		timeIntervalNames[ti.Name] = struct{}{}
	}
	return timeIntervalNames, nil
}
