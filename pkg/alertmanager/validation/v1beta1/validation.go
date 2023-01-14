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
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/prometheus-operator/prometheus-operator/pkg/alertmanager/validation"
	monitoringv1beta1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1beta1"
)

var durationRe = regexp.MustCompile(`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$`)

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

	return validateAlertManagerRoutes(amc.Spec.Route, receivers, timeIntervals, true)
}

func validateReceivers(receivers []monitoringv1beta1.Receiver) (map[string]struct{}, error) {
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
			return nil, errors.Wrapf(err, "failed to validate 'webhookConfig' - receiver %s", receiver.Name)
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

		if err := validateSnsConfigs(receiver.SNSConfigs); err != nil {
			return nil, errors.Wrapf(err, "failed to validate 'snsConfig' - receiver %s", receiver.Name)
		}
	}

	return receiverNames, nil
}

func validatePagerDutyConfigs(configs []monitoringv1beta1.PagerDutyConfig) error {
	for _, conf := range configs {
		if conf.URL != "" {
			if _, err := validation.ValidateURL(conf.URL); err != nil {
				return errors.Wrap(err, "pagerduty validation failed for 'url'")
			}
		}
		if conf.RoutingKey == nil && conf.ServiceKey == nil {
			return errors.New("one of 'routingKey' or 'serviceKey' is required")
		}

		if err := conf.HTTPConfig.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func validateOpsGenieConfigs(configs []monitoringv1beta1.OpsGenieConfig) error {
	for _, config := range configs {
		if err := config.Validate(); err != nil {
			return err
		}
		if config.APIURL != "" {
			if _, err := validation.ValidateURL(config.APIURL); err != nil {
				return errors.Wrap(err, "invalid 'apiURL'")
			}
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return err
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
	for _, config := range configs {
		if config.URL == nil && config.URLSecret == nil {
			return errors.New("one of 'url' or 'urlSecret' must be specified")
		}
		if config.URL != nil {
			if _, err := validation.ValidateURL(*config.URL); err != nil {
				return errors.Wrapf(err, "invalid 'url'")
			}
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func validateWechatConfigs(configs []monitoringv1beta1.WeChatConfig) error {
	for _, config := range configs {
		if config.APIURL != "" {
			if _, err := validation.ValidateURL(config.APIURL); err != nil {
				return errors.Wrap(err, "invalid 'apiURL'")
			}
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func validateEmailConfig(configs []monitoringv1beta1.EmailConfig) error {
	for _, config := range configs {
		if config.To == "" {
			return errors.New("missing 'to' address")
		}

		if config.Smarthost != "" {
			_, _, err := net.SplitHostPort(config.Smarthost)
			if err != nil {
				return errors.Wrapf(err, "invalid field 'smarthost': %s", config.Smarthost)
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
			return errors.New("missing 'routingKey' key")
		}

		if config.APIURL != "" {
			if _, err := validation.ValidateURL(config.APIURL); err != nil {
				return errors.Wrapf(err, "'apiURL' %s invalid", config.APIURL)
			}
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func validatePushoverConfigs(configs []monitoringv1beta1.PushoverConfig) error {
	for _, config := range configs {
		if config.UserKey == nil {
			return errors.Errorf("mandatory field %q is empty", "userKey")
		}

		if config.Token == nil {
			return errors.Errorf("mandatory field %q is empty", "token")
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func validateSnsConfigs(configs []monitoringv1beta1.SNSConfig) error {
	for _, config := range configs {
		if (config.TargetARN == "") != (config.TopicARN == "") != (config.PhoneNumber == "") {
			return fmt.Errorf("must provide either a Target ARN, Topic ARN, or Phone Number for SNS config")
		}

		if err := config.HTTPConfig.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// validateAlertManagerRoutes verifies that the given route and all its children are semantically valid.
// because of the self-referential issues mentioned in https://github.com/kubernetes/kubernetes/issues/62872
// it is not currently possible to apply OpenAPI validation to a v1beta1.Route
func validateAlertManagerRoutes(r *monitoringv1beta1.Route, receivers, timeIntervals map[string]struct{}, topLevelRoute bool) error {
	if r == nil {
		return nil
	}

	if r.Receiver == "" {
		if topLevelRoute {
			return errors.Errorf("root route must define a receiver")
		}
	} else {
		if _, found := receivers[r.Receiver]; !found {
			return errors.Errorf("receiver %q not found", r.Receiver)
		}
	}

	if groupLen := len(r.GroupBy); groupLen > 0 {
		groupedBy := make(map[string]struct{}, groupLen)
		for _, str := range r.GroupBy {
			if _, found := groupedBy[str]; found {
				return errors.Errorf("duplicate values not permitted in route 'groupBy': %v", r.GroupBy)
			}
			groupedBy[str] = struct{}{}
		}
		if _, found := groupedBy["..."]; found && groupLen > 1 {
			return errors.Errorf("'...' must be a sole value in route 'groupBy': %v", r.GroupBy)
		}
	}

	for _, namedTimeInterval := range r.MuteTimeIntervals {
		if _, found := timeIntervals[namedTimeInterval]; !found {
			return errors.Errorf("time interval %q not found", namedTimeInterval)
		}
	}

	for _, namedTimeInterval := range r.ActiveTimeIntervals {
		if _, found := timeIntervals[namedTimeInterval]; !found {
			return errors.Errorf("time interval %q not found", namedTimeInterval)
		}
	}

	// validate that if defaults are set, they match regex
	if r.GroupInterval != "" && !durationRe.MatchString(r.GroupInterval) {
		return errors.Errorf("groupInterval %s does not match required regex: %s", r.GroupInterval, durationRe.String())

	}
	if r.GroupWait != "" && !durationRe.MatchString(r.GroupWait) {
		return errors.Errorf("groupWait %s does not match required regex: %s", r.GroupWait, durationRe.String())
	}

	if r.RepeatInterval != "" && !durationRe.MatchString(r.RepeatInterval) {
		return errors.Errorf("repeatInterval %s does not match required regex: %s", r.RepeatInterval, durationRe.String())
	}

	children, err := r.ChildRoutes()
	if err != nil {
		return err
	}

	for i := range children {
		if err := validateAlertManagerRoutes(&children[i], receivers, timeIntervals, false); err != nil {
			return errors.Wrapf(err, "route[%d]", i)
		}
	}

	return nil
}

func validateTimeIntervals(timeIntervals []monitoringv1beta1.TimeInterval) (map[string]struct{}, error) {
	timeIntervalNames := make(map[string]struct{}, len(timeIntervals))

	for i, ti := range timeIntervals {
		if err := ti.Validate(); err != nil {
			return nil, errors.Wrapf(err, "time interval[%d] is invalid", i)
		}
		timeIntervalNames[ti.Name] = struct{}{}
	}
	return timeIntervalNames, nil
}
