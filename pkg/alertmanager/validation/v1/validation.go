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

package v1

import (
	"fmt"

	"github.com/prometheus-operator/prometheus-operator/pkg/alertmanager/validation"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func ValidateAlertmanagerGlobalConfig(gc *monitoringv1.AlertmanagerGlobalConfig) error {
	if gc == nil {
		return nil
	}

	if err := gc.HTTPConfigWithProxy.Validate(); err != nil {
		return fmt.Errorf("httpConfig: %w", err)
	}

	if err := validatingTelegramConfig(gc.TelegramConfig); err != nil {
		return fmt.Errorf("telegram: %w", err)
	}

	if err := validateGlobalJiraConfig(gc.JiraConfig); err != nil {
		return fmt.Errorf("jira: %w", err)
	}

	if err := validateGlobalVictorOpsConfig(gc.VictorOpsConfig); err != nil {
		return fmt.Errorf("victorops: %w", err)
	}

	if err := validateGlobalRocketChatConfig(gc.RocketChatConfig); err != nil {
		return fmt.Errorf("rocketChat: %w", err)
	}

	if err := validateGlobalWebexConfig(gc.WebexConfig); err != nil {
		return fmt.Errorf("webex: %w", err)
	}

	if err := validateGlobalWeChatConfig(gc.WeChatConfig); err != nil {
		return fmt.Errorf("wechatConfig: %w", err)
	}

	return nil
}

func validatingTelegramConfig(tc *monitoringv1.GlobalTelegramConfig) error {
	if tc == nil {
		return nil
	}

	if err := validation.ValidateURLPtr((*string)(tc.APIURL)); err != nil {
		return fmt.Errorf("invalid apiURL: %w", err)
	}

	return nil
}

func validateGlobalJiraConfig(jc *monitoringv1.GlobalJiraConfig) error {
	if jc == nil {
		return nil
	}

	if err := validation.ValidateURLPtr((*string)(jc.APIURL)); err != nil {
		return fmt.Errorf("invalid apiURL: %w", err)
	}

	return nil
}

func validateGlobalVictorOpsConfig(vc *monitoringv1.GlobalVictorOpsConfig) error {
	if vc == nil {
		return nil
	}

	if err := validation.ValidateURLPtr((*string)(vc.APIURL)); err != nil {
		return fmt.Errorf("invalid apiURL: %w", err)
	}

	return nil
}

func validateGlobalRocketChatConfig(rc *monitoringv1.GlobalRocketChatConfig) error {
	if rc == nil {
		return nil
	}

	if err := validation.ValidateURLPtr((*string)(rc.APIURL)); err != nil {
		return fmt.Errorf("invalid apiURL: %w", err)
	}

	return nil
}

func validateGlobalWebexConfig(wc *monitoringv1.GlobalWebexConfig) error {
	if wc == nil {
		return nil
	}

	if err := validation.ValidateURLPtr((*string)(wc.APIURL)); err != nil {
		return fmt.Errorf("invalid apiURL: %w", err)
	}

	return nil
}

func validateGlobalWeChatConfig(wc *monitoringv1.GlobalWeChatConfig) error {
	if wc == nil {
		return nil
	}

	if err := validation.ValidateURLPtr((*string)(wc.APIURL)); err != nil {
		return fmt.Errorf("invalid apiURL: %w", err)
	}

	return nil
}
