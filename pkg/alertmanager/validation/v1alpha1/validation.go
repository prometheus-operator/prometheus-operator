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

package v1alpha1

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	"k8s.io/utils/ptr"

	"github.com/prometheus-operator/prometheus-operator/pkg/alertmanager/validation"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

var durationRe = regexp.MustCompile(`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$`)

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
			return nil, fmt.Errorf("%q receiver is not unique: %w", receiver.Name, err)
		}
		receiverNames[receiver.Name] = struct{}{}

		if err := validateOpsGenieConfigs(receiver.OpsGenieConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'opsgenieConfigs'%w", receiver.Name, err)
		}

		if err = validatePagerDutyConfigs(receiver.PagerDutyConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'pagerdutyConfigs'%w", receiver.Name, err)
		}

		if err := validateDiscordConfigs(receiver.DiscordConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'discordConfigs'%w", receiver.Name, err)
		}

		if err := validateSlackConfigs(receiver.SlackConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'slackConfigs'%w", receiver.Name, err)
		}

		if err := validateWebhookConfigs(receiver.WebhookConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'webhookConfigs'%w", receiver.Name, err)
		}

		if err := validateWechatConfigs(receiver.WeChatConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'wechatConfigs'%w", receiver.Name, err)
		}

		if err := validateEmailConfig(receiver.EmailConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'emailConfigs'%w", receiver.Name, err)
		}

		if err := validateVictorOpsConfigs(receiver.VictorOpsConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'victoropsConfis%w", receiver.Name, err)
		}

		if err := validatePushoverConfigs(receiver.PushoverConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'pushoverConfigs'%w", receiver.Name, err)
		}

		if err := validateSnsConfigs(receiver.SNSConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'snsConfigs'%w", receiver.Name, err)
		}

		if err := validateTelegramConfigs(receiver.TelegramConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'telegramConfigs'%w", receiver.Name, err)
		}

		if err := validateWebexConfigs(receiver.WebexConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'webexConfigs'%w", receiver.Name, err)
		}

		if err := validateMSTeamsConfigs(receiver.MSTeamsConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'msteamsConfigs'%w", receiver.Name, err)
		}

		if err := validateMSTeamsV2Configs(receiver.MSTeamsV2Configs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'msteamsv2Configs'%w", receiver.Name, err)
		}

		if err := validateRocketchatConfigs(receiver.RocketChatConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate receiver %s: 'rocketchatConfigs'%w", receiver.Name, err)
		}
	}

	return receiverNames, nil
}

func validatePagerDutyConfigs(configs []monitoringv1alpha1.PagerDutyConfig) error {
	for i, conf := range configs {
		if err := validation.ValidateURLPtr((*string)(conf.URL)); err != nil {
			return fmt.Errorf("[%d]: url: %w", i, err)
		}

		if err := validation.ValidateURLPtr((*string)(conf.ClientURL)); err != nil {
			return fmt.Errorf("[%d]: clientURL: %w", i, err)
		}

		if conf.RoutingKey == nil && conf.ServiceKey == nil {
			return errors.New("[%d]: one of 'routingKey' or 'serviceKey' is required")
		}

		for j, lc := range conf.PagerDutyLinkConfigs {
			if err := validation.ValidateURLPtr((*string)(lc.Href)); err != nil {
				return fmt.Errorf("[%d]: pagerDutyLinkConfigs[%d]: href: %w", i, j, err)
			}
		}

		for j, ic := range conf.PagerDutyImageConfigs {
			if err := validation.ValidateURLPtr((*string)(ic.Href)); err != nil {
				return fmt.Errorf("[%d]: pagerDutyImageConfigs[%d]: href: %w", i, j, err)
			}
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}
	return nil
}

func validateOpsGenieConfigs(configs []monitoringv1alpha1.OpsGenieConfig) error {
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

func validateDiscordConfigs(configs []monitoringv1alpha1.DiscordConfig) error {
	for i, config := range configs {
		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}
	return nil
}

func validateRocketchatConfigs(configs []monitoringv1alpha1.RocketChatConfig) error {
	for i, config := range configs {
		if err := validation.ValidateURLPtr((*string)(config.APIURL)); err != nil {
			return fmt.Errorf("[%d]: 'apiURL': %w", i, err)
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func validateSlackConfigs(configs []monitoringv1alpha1.SlackConfig) error {
	for i, config := range configs {
		if err := config.Validate(); err != nil {
			return err
		}

		if err := validation.ValidateURLPtr((*string)(config.IconURL)); err != nil {
			return fmt.Errorf("[%d]: 'iconURL': %w", i, err)
		}

		if err := validation.ValidateURLPtr((*string)(config.ImageURL)); err != nil {
			return fmt.Errorf("[%d]: 'imageURL': %w", i, err)
		}

		if err := validation.ValidateURLPtr((*string)(config.ThumbURL)); err != nil {
			return fmt.Errorf("[%d]: 'thumbURL': %w", i, err)
		}

		for j, sa := range config.Actions {
			if err := validation.ValidateURLPtr((*string)(sa.URL)); err != nil {
				return fmt.Errorf("[%d]: invalid 'action'[%d]: 'url': %w", i, j, err)
			}
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}
	return nil
}

func validateWebhookConfigs(configs []monitoringv1alpha1.WebhookConfig) error {
	for i, config := range configs {
		if config.URL == nil && config.URLSecret == nil {
			return fmt.Errorf("[%d]: one of 'url' or 'urlSecret' must be specified", i)
		}

		if err := validation.ValidateURLPtr((*string)(config.URL)); err != nil {
			return fmt.Errorf("[%d]: url: %w", i, err)
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func validateWechatConfigs(configs []monitoringv1alpha1.WeChatConfig) error {
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

func validateEmailConfig(configs []monitoringv1alpha1.EmailConfig) error {
<<<<<<< HEAD
	for i, config := range configs {
		if config.To == "" {
			return fmt.Errorf("[%d]: missing 'to' address", i)
=======
	for _, config := range configs {
		if ptr.Deref(config.To, "") == "" {
			return errors.New("missing 'to' address")
>>>>>>> main
		}

		if ptr.Deref(config.Smarthost, "") != "" {
			_, _, err := net.SplitHostPort(*config.Smarthost)
			if err != nil {
<<<<<<< HEAD
				return fmt.Errorf("[%d]: invalid 'smarthost' %s: %w", i, config.Smarthost, err)
=======
				return fmt.Errorf("invalid 'smarthost' %s: %w", *config.Smarthost, err)
>>>>>>> main
			}
		}

		if config.Headers != nil {
			// Header names are case-insensitive, check for collisions.
			normalizedHeaders := map[string]struct{}{}
			for _, v := range config.Headers {
				normalized := strings.ToLower(v.Key)
				if _, ok := normalizedHeaders[normalized]; ok {
					return fmt.Errorf("[%d]: duplicate header %q", i, normalized)
				}
				normalizedHeaders[normalized] = struct{}{}
			}
		}
	}
	return nil
}

func validateVictorOpsConfigs(configs []monitoringv1alpha1.VictorOpsConfig) error {
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
					return fmt.Errorf("[%d]: usage of reserved word %q is not allowed in custom fields", i, v.Key)
				}
			}
		}

		if config.RoutingKey == "" {
			return fmt.Errorf("[%d]: missing 'routingKey' key", i)
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

func validatePushoverConfigs(configs []monitoringv1alpha1.PushoverConfig) error {
	for i, config := range configs {
		if config.UserKey == nil && config.UserKeyFile == nil {
			return fmt.Errorf("[%d]: one of 'userKey' or 'userKeyFile' must be configured", i)
		}

		if config.Token == nil && config.TokenFile == nil {
			return fmt.Errorf("[%d]: one of 'token' or 'tokenFile' must be configured", i)
		}

		if config.HTML != nil && *config.HTML && config.Monospace != nil && *config.Monospace {
			return fmt.Errorf("[%d]: 'html' and 'monospace' options are mutually exclusive", i)
		}

		if err := validation.ValidateURLPtr((*string)(config.URL)); err != nil {
			return fmt.Errorf("[%d]: url: %w", i, err)
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func validateSnsConfigs(configs []monitoringv1alpha1.SNSConfig) error {
	for i, config := range configs {
<<<<<<< HEAD
		if (config.TargetARN == "") != (config.TopicARN == "") != (config.PhoneNumber == "") {
			return fmt.Errorf("[%d]: must provide one of 'targetARN', 'topicARN', or 'phoneNumber'", i)
=======
		if (ptr.Deref(config.TargetARN, "") == "") != (ptr.Deref(config.TopicARN, "") == "") != (ptr.Deref(config.PhoneNumber, "") == "") {
			return fmt.Errorf("[%d]: must provide either a targetARN, topicARN, or phoneNumber for SNS config", i)
		}

		if err := validation.ValidateURLPtr((*string)(config.ApiURL)); err != nil {
			return fmt.Errorf("[%d]: apiURL: %w", i, err)
>>>>>>> main
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}
	return nil
}

func validateTelegramConfigs(configs []monitoringv1alpha1.TelegramConfig) error {
	for i, config := range configs {
		if config.BotToken == nil && config.BotTokenFile == nil {
			return fmt.Errorf("[%d]: mandatory field 'botToken' or 'botTokenfile' is empty", i)
		}

		if config.ChatID == 0 {
			return fmt.Errorf("[%d]: mandatory field 'chatID' is empty", i)
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

func validateWebexConfigs(configs []monitoringv1alpha1.WebexConfig) error {
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

func validateMSTeamsConfigs(configs []monitoringv1alpha1.MSTeamsConfig) error {
	for i, config := range configs {
		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func validateMSTeamsV2Configs(configs []monitoringv1alpha1.MSTeamsV2Config) error {
	for i, config := range configs {
		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
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

	if r.GroupInterval != "" && !durationRe.MatchString(r.GroupInterval) {
		return fmt.Errorf("groupInterval %s does not match required regex: %s", r.GroupInterval, durationRe.String())

	}
	if r.GroupWait != "" && !durationRe.MatchString(r.GroupWait) {
		return fmt.Errorf("groupWait %s does not match required regex: %s", r.GroupWait, durationRe.String())
	}

	if r.RepeatInterval != "" && !durationRe.MatchString(r.RepeatInterval) {
		return fmt.Errorf("repeatInterval %s does not match required regex: %s", r.RepeatInterval, durationRe.String())
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
