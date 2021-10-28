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

package alertmanager

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

func ValidateConfig(amc *monitoringv1alpha1.AlertmanagerConfig) error {
	receivers, err := validateReceivers(amc.Spec.Receivers)
	if err != nil {
		return err
	}

	return validateAlertManagerRoutes(amc.Spec.Route, receivers, true)
}

func validateReceivers(receivers []monitoringv1alpha1.Receiver) (map[string]struct{}, error) {
	var err error
	receiverNames := make(map[string]struct{})

	for _, receiver := range receivers {
		if _, found := receiverNames[receiver.Name]; found {
			return nil, errors.Errorf("%q receiver is not unique", receiver.Name)
		}
		receiverNames[receiver.Name] = struct{}{}

		if err = validatePagerDutyConfigs(receiver.PagerDutyConfigs); err != nil {
			return nil, errors.Wrapf(err, "failed to validate 'pagerDutyConfig' - receiver %s", receiver.Name)
		}

		if err := validateOpsGenieConfigs(receiver.OpsGenieConfigs); err != nil {
			return nil, errors.Wrapf(err, "failed to validate 'opsGenieConfig' - receiver %s", receiver.Name)
		}

		if err := validateSlackConfigs(receiver.SlackConfigs); err != nil {
			return nil, errors.Wrapf(err, "failed to validate 'slackConfig' - receiver %s", receiver.Name)
		}

		if err := validateWebhookConfigs(receiver.WebhookConfigs); err != nil {
			return nil, errors.Wrapf(err, "failed to validate 'slackConfig' - receiver %s", receiver.Name)
		}

		if err := validateWechatConfigs(receiver.WeChatConfigs); err != nil {
			return nil, errors.Wrapf(err, "failed to validate 'weChatConfig' - receiver %s", receiver.Name)
		}

		if err := validateEmailConfig(receiver.EmailConfigs); err != nil {
			return nil, errors.Wrapf(err, "failed to validate 'emailConfig' - receiver %s", receiver.Name)

		}

		if err := validateVictorOpsConfigs(receiver.VictorOpsConfigs); err != nil {
			return nil, errors.Wrapf(err, "failed to validate 'victorOpsConfig' - receiver %s", receiver.Name)
		}

		if err := validatePushoverConfigs(receiver.PushoverConfigs); err != nil {
			return nil, errors.Wrapf(err, "failed to validate 'pushOverConfig' - receiver %s", receiver.Name)
		}
	}

	return receiverNames, nil
}

// validatePagerDutyConfigs is a no-op
func validatePagerDutyConfigs(configs []monitoringv1alpha1.PagerDutyConfig) error {
	return nil
}

func validateOpsGenieConfigs(configs []monitoringv1alpha1.OpsGenieConfig) error {
	for _, config := range configs {
		if err := config.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func validateSlackConfigs(configs []monitoringv1alpha1.SlackConfig) error {
	for _, config := range configs {
		if err := config.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func validateWebhookConfigs(configs []monitoringv1alpha1.WebhookConfig) error {
	for _, config := range configs {
		if config.URL == nil && config.URLSecret == nil {
			return errors.New("one of 'url' or 'urlSecret' must be specified")
		}
	}
	return nil
}

func validateWechatConfigs(configs []monitoringv1alpha1.WeChatConfig) error {
	for _, config := range configs {
		if config.APIURL != "" {
			if _, err := url.Parse(config.APIURL); err != nil {
				return errors.Wrap(err, "weChat 'apiURL' not valid")
			}
		}
	}
	return nil
}

func validateEmailConfig(configs []monitoringv1alpha1.EmailConfig) error {
	for _, config := range configs {
		if config.To == "" {
			return errors.New("missing to address in email config")
		}

		if config.Smarthost != "" {
			_, _, err := net.SplitHostPort(config.Smarthost)
			if err != nil {
				return errors.New("invalid email field 'smarthost'")
			}
		}

		if config.Headers != nil {
			// Header names are case-insensitive, check for collisions.
			normalizedHeaders := map[string]struct{}{}
			for _, v := range config.Headers {
				normalized := strings.Title(v.Key)
				if _, ok := normalizedHeaders[normalized]; ok {
					return fmt.Errorf("duplicate header %q in email config", normalized)
				}
				normalizedHeaders[normalized] = struct{}{}
			}
		}
	}
	return nil
}

func validateVictorOpsConfigs(configs []monitoringv1alpha1.VictorOpsConfig) error {
	for _, config := range configs {

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
			return errors.New("missing 'routingKey' key in VictorOps config")
		}
	}
	return nil
}

func validatePushoverConfigs(configs []monitoringv1alpha1.PushoverConfig) error {
	for _, config := range configs {
		if config.UserKey == nil {
			return errors.Errorf("mandatory field %q is empty", "userKey")
		}

		if config.Token == nil {
			return errors.Errorf("mandatory field %q is empty", "token")
		}

		if config.Retry != "" {
			_, err := time.ParseDuration(config.Retry)
			if err != nil {
				return errors.New("invalid retry duration")
			}
		}
		if config.Expire != "" {
			_, err := time.ParseDuration(config.Expire)
			if err != nil {
				return errors.New("invalid expire duration")
			}
		}
	}

	return nil
}

// validateAlertManagerRoutes verifies that the given route and all its children are semantically valid.
func validateAlertManagerRoutes(r *monitoringv1alpha1.Route, receivers map[string]struct{}, topLevelRoute bool) error {
	if r == nil {
		return nil
	}

	if _, found := receivers[r.Receiver]; !found && (r.Receiver != "" || topLevelRoute) {
		return errors.Errorf("receiver %q not found", r.Receiver)
	}

	children, err := r.ChildRoutes()
	if err != nil {
		return err
	}

	for i := range children {
		if err := validateAlertManagerRoutes(&children[i], receivers, false); err != nil {
			return errors.Wrapf(err, "route[%d]", i)
		}
	}

	return nil
}
