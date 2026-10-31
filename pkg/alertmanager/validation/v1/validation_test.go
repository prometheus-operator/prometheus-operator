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

package v1

import (
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestValidateGlobalTelegramConfig(t *testing.T) {
	testCases := []struct {
		name      string
		config    *monitoringv1.AlertmanagerGlobalConfig
		expectErr bool
		errMsg    *string
	}{
		{
			name: "Test validate global telegram config - botToken specified",
			config: &monitoringv1.AlertmanagerGlobalConfig{
				TelegramConfig: &monitoringv1.GlobalTelegramConfig{
					BotToken: &v1.SecretKeySelector{
						Key:                  "token",
						LocalObjectReference: v1.LocalObjectReference{Name: "telegram"},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Test validate global telegram config - botTokenFile specified",
			config: &monitoringv1.AlertmanagerGlobalConfig{
				TelegramConfig: &monitoringv1.GlobalTelegramConfig{
					BotTokenFile: new("/bot/token/file"),
				},
			},
			expectErr: false,
		},
		{
			name: "Test fail to validate global telegram config - botToken and botTokenFile specified",
			config: &monitoringv1.AlertmanagerGlobalConfig{
				TelegramConfig: &monitoringv1.GlobalTelegramConfig{
					BotToken: &v1.SecretKeySelector{
						Key:                  "token",
						LocalObjectReference: v1.LocalObjectReference{Name: "telegram"},
					},
					BotTokenFile: new("/bot/token/file"),
				},
			},
			expectErr: true,
			errMsg:    new("only one of 'botToken' or 'botTokenfile' must be configured"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAlertmanagerGlobalConfig(tc.config)
			if !tc.expectErr {
				require.NoError(t, err)
				return
			}

			t.Logf("err: %s", err)
			if tc.errMsg != nil {
				assert.EqualError(t, err, *tc.errMsg)
				return
			}
			require.Error(t, err)
		})
	}
}
