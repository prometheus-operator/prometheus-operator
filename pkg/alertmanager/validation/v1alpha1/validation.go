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
			return nil, fmt.Errorf("%q receiver is not unique: %w", receiver.Name, err)
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

		if err := validateWebexConfigs(receiver.WebexConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'webexConfig' - receiver %s: %w", receiver.Name, err)
		}

		if err := validateDiscordConfigs(receiver.DiscordConfigs); err != nil {
			return nil, fmt.Errorf("failed to validate 'discordConfig' - receiver %s: %w", receiver.Name, err)
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

func validatePagerDutyConfigs(configs []monitoringv1alpha1.PagerDutyConfig) error {
	for i, conf := range configs {
		if err := validation.ValidateURLPtr((*string)(conf.URL)); err != nil {
			return fmt.Errorf("[%d]: url: %w", i, err)
		}

		if conf.ClientURL != nil && *conf.ClientURL != "" {
			if err := validation.ValidateTemplateURL(*conf.ClientURL); err != nil {
				return fmt.Errorf("[%d]: clientURL: %w", i, err)
			}
		}

		if err := validation.ValidatePagerDutyConfig(conf.RoutingKey != nil, conf.ServiceKey != nil); err != nil {
			return err
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
	for _, config := range configs {
		if err := config.HTTPConfig.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func validateRocketchatConfigs(configs []monitoringv1alpha1.RocketChatConfig) error {
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
				return fmt.Errorf("[%d]: actions[%d]: invalid 'url': %w", i, j, err)
			}
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}

	return nil
}

func validateSlackConfigs(configs []monitoringv1alpha1.SlackConfig) error {
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

func validateWebhookConfigs(configs []monitoringv1alpha1.WebhookConfig) error {
	for i, config := range configs {
		if err := validation.ValidateWebhookConfig(config.URL != nil, config.URLSecret != nil); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
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
	for _, config := range configs {
		if ptr.Deref(config.To, "") == "" {
			return errors.New("missing 'to' address")
		}

		if smarthost := ptr.Deref(config.Smarthost, ""); smarthost != "" {
			if err := validation.ValidateSmarthost(smarthost); err != nil {
				return err
			}
		}

		if config.Headers != nil {
			keys := make([]string, 0, len(config.Headers))
			for _, v := range config.Headers {
				keys = append(keys, v.Key)
			}
			if err := validation.ValidateEmailHeaders(keys); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateVictorOpsConfigs(configs []monitoringv1alpha1.VictorOpsConfig) error {
	for i, config := range configs {
		if len(config.CustomFields) > 0 {
			keys := make([]string, 0, len(config.CustomFields))
			for _, v := range config.CustomFields {
				keys = append(keys, v.Key)
			}
			if err := validation.ValidateVictorOpsCustomFields(keys); err != nil {
				return err
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

func validatePushoverConfigs(configs []monitoringv1alpha1.PushoverConfig) error {
	for i, config := range configs {
		if err := validation.ValidatePushoverConfig(
			config.UserKey != nil,
			config.UserKeyFile != nil,
			config.Token != nil,
			config.TokenFile != nil,
			config.HTML != nil && *config.HTML,
			config.Monospace != nil && *config.Monospace,
		); err != nil {
			return err
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

func validateSnsConfigs(configs []monitoringv1alpha1.SNSConfig) error {
	for i, config := range configs {
		if err := validation.ValidateSNSConfig(
			ptr.Deref(config.TargetARN, "") != "",
			ptr.Deref(config.TopicARN, "") != "",
			ptr.Deref(config.PhoneNumber, "") != "",
		); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
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

func validateTelegramConfigs(configs []monitoringv1alpha1.TelegramConfig) error {
	for i, config := range configs {
		if err := validation.ValidateTelegramConfig(
			config.BotToken != nil,
			config.BotTokenFile != nil,
			config.ChatID,
		); err != nil {
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
	for _, config := range configs {
		if err := config.HTTPConfig.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func validateMSTeamsV2Configs(configs []monitoringv1alpha1.MSTeamsV2Config) error {
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
// possible to apply OpenAPI validation to a v1alpha1.Route.
func validateRoute(r *monitoringv1alpha1.Route, receivers, muteTimeIntervals map[string]struct{}, topLevelRoute bool) error {
	if r == nil {
		return nil
	}

	if err := validation.ValidateRouteReceiver(r.Receiver, receivers, topLevelRoute); err != nil {
		return err
	}

	if err := validation.ValidateRouteGroupBy(r.GroupBy); err != nil {
		return err
	}

	for _, namedMuteTimeInterval := range r.MuteTimeIntervals {
		if err := validation.ValidateTimeIntervalReference(namedMuteTimeInterval, muteTimeIntervals, true); err != nil {
			return err
		}
	}

	for _, namedActiveTimeInterval := range r.ActiveTimeIntervals {
		if err := validation.ValidateTimeIntervalReference(namedActiveTimeInterval, muteTimeIntervals, false); err != nil {
			return err
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
