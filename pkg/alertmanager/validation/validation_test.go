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

package validation

import (
	"net/url"
	"testing"

	"github.com/prometheus/alertmanager/config"
	"github.com/stretchr/testify/require"
)

func TestValidateUrl(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		expectErr    bool
		expectResult func() *config.URL
	}{
		{
			name:      "Test invalid url returns error",
			in:        "https://!^invalid.com",
			expectErr: true,
		},
		{
			name:      "Test missing scheme returns error",
			in:        "is.normally.valid",
			expectErr: true,
		},
		{
			name: "Test happy path",
			in:   "https://u:p@is.compliant.with.upstream.unmarshal",
			expectResult: func() *config.URL {
				u, _ := url.Parse("https://u:p@is.compliant.with.upstream.unmarshal")
				return &config.URL{URL: u}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := ValidateURL(tc.in)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			res := tc.expectResult()
			require.Equal(t, u, res, "wanted %v but got %v", res, u)
		})
	}
}
func TestValidateSecretUrl(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		expectErr    bool
		expectResult func() *config.URL
	}{
		{
			name:      "Test invalid url returns error",
			in:        "https://!^invalid.com",
			expectErr: true,
		},
		{
			name:      "Test missing scheme returns error",
			in:        "is.normally.valid",
			expectErr: true,
		},
		{
			name: "Test happy path",
			in:   "https://u:p@is.compliant.with.upstream.unmarshal",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSecretURL(tc.in)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

		})
	}
}

func TestValidateVictorOpsCustomFields(t *testing.T) {
	tests := []struct {
		name      string
		keys      []string
		expectErr bool
	}{
		{
			name:      "empty keys is valid",
			keys:      []string{},
			expectErr: false,
		},
		{
			name:      "valid custom field",
			keys:      []string{"my_custom_field"},
			expectErr: false,
		},
		{
			name:      "reserved field routing_key",
			keys:      []string{"routing_key"},
			expectErr: true,
		},
		{
			name:      "reserved field entity_id",
			keys:      []string{"entity_id"},
			expectErr: true,
		},
		{
			name:      "mix of valid and reserved",
			keys:      []string{"valid_field", "message_type"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateVictorOpsCustomFields(tc.keys)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateEmailHeaders(t *testing.T) {
	tests := []struct {
		name      string
		keys      []string
		expectErr bool
	}{
		{
			name:      "empty headers is valid",
			keys:      []string{},
			expectErr: false,
		},
		{
			name:      "single header is valid",
			keys:      []string{"Content-Type"},
			expectErr: false,
		},
		{
			name:      "different headers is valid",
			keys:      []string{"Content-Type", "X-Custom"},
			expectErr: false,
		},
		{
			name:      "duplicate headers is invalid",
			keys:      []string{"Content-Type", "content-type"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEmailHeaders(tc.keys)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidatePushoverConfig(t *testing.T) {
	tests := []struct {
		name           string
		hasUserKey     bool
		hasUserKeyFile bool
		hasToken       bool
		hasTokenFile   bool
		html           bool
		monospace      bool
		expectErr      bool
	}{
		{
			name:       "userKey and token present",
			hasUserKey: true,
			hasToken:   true,
			expectErr:  false,
		},
		{
			name:           "userKeyFile and tokenFile present",
			hasUserKeyFile: true,
			hasTokenFile:   true,
			expectErr:      false,
		},
		{
			name:      "missing userKey and userKeyFile",
			hasToken:  true,
			expectErr: true,
		},
		{
			name:       "missing token and tokenFile",
			hasUserKey: true,
			expectErr:  true,
		},
		{
			name:       "html and monospace both true",
			hasUserKey: true,
			hasToken:   true,
			html:       true,
			monospace:  true,
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePushoverConfig(tc.hasUserKey, tc.hasUserKeyFile, tc.hasToken, tc.hasTokenFile, tc.html, tc.monospace)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateWebhookConfig(t *testing.T) {
	tests := []struct {
		name         string
		hasURL       bool
		hasURLSecret bool
		expectErr    bool
	}{
		{
			name:      "has URL",
			hasURL:    true,
			expectErr: false,
		},
		{
			name:         "has URLSecret",
			hasURLSecret: true,
			expectErr:    false,
		},
		{
			name:      "missing both",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateWebhookConfig(tc.hasURL, tc.hasURLSecret)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidatePagerDutyConfig(t *testing.T) {
	tests := []struct {
		name          string
		hasRoutingKey bool
		hasServiceKey bool
		expectErr     bool
	}{
		{
			name:          "has routingKey",
			hasRoutingKey: true,
			expectErr:     false,
		},
		{
			name:          "has serviceKey",
			hasServiceKey: true,
			expectErr:     false,
		},
		{
			name:      "missing both",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePagerDutyConfig(tc.hasRoutingKey, tc.hasServiceKey)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateSmarthost(t *testing.T) {
	tests := []struct {
		name      string
		smarthost string
		expectErr bool
	}{
		{
			name:      "valid host:port",
			smarthost: "smtp.example.com:587",
			expectErr: false,
		},
		{
			name:      "invalid format",
			smarthost: "invalid",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSmarthost(tc.smarthost)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateRouteGroupBy(t *testing.T) {
	tests := []struct {
		name      string
		groupBy   []string
		expectErr bool
	}{
		{
			name:      "empty is valid",
			groupBy:   []string{},
			expectErr: false,
		},
		{
			name:      "single value is valid",
			groupBy:   []string{"alertname"},
			expectErr: false,
		},
		{
			name:      "ellipsis alone is valid",
			groupBy:   []string{"..."},
			expectErr: false,
		},
		{
			name:      "duplicate values",
			groupBy:   []string{"alertname", "alertname"},
			expectErr: true,
		},
		{
			name:      "ellipsis with other values",
			groupBy:   []string{"...", "alertname"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRouteGroupBy(tc.groupBy)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateRouteReceiver(t *testing.T) {
	receivers := map[string]struct{}{
		"default": {},
		"slack":   {},
	}

	tests := []struct {
		name       string
		receiver   string
		isTopLevel bool
		expectErr  bool
	}{
		{
			name:       "valid receiver",
			receiver:   "default",
			isTopLevel: true,
			expectErr:  false,
		},
		{
			name:       "missing receiver on top-level",
			receiver:   "",
			isTopLevel: true,
			expectErr:  true,
		},
		{
			name:       "missing receiver on child is valid",
			receiver:   "",
			isTopLevel: false,
			expectErr:  false,
		},
		{
			name:       "receiver not found",
			receiver:   "nonexistent",
			isTopLevel: true,
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRouteReceiver(tc.receiver, receivers, tc.isTopLevel)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
