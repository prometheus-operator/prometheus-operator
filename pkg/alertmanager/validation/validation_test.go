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

package validation

import (
	"net/url"
	"testing"

	"github.com/prometheus/alertmanager/config/common"
	"github.com/stretchr/testify/require"
)

func TestValidateUrl(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		expectErr    bool
		expectResult func() *common.URL
	}{
		{
			name:      "invalid url",
			in:        "https://!^example.com",
			expectErr: true,
		},
		{
			name:      "missing host",
			in:        "http://",
			expectErr: true,
		},
		{
			name:      "missing scheme",
			in:        "example.com",
			expectErr: true,
		},
		{
			name:      "invalid scheme",
			in:        "tcp://example.com",
			expectErr: true,
		},
		{
			name: "valid URL",
			in:   "https://u:p@example.com",
			expectResult: func() *common.URL {
				u, _ := url.Parse("https://u:p@example.com")
				return &common.URL{URL: u}
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
			require.Equal(t, u.String(), res.String())
		})
	}
}
func TestValidateSecretUrl(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		expectErr    bool
		expectResult func() *common.URL
	}{
		{
			name:      "invalid URL",
			in:        "https://!^example.com",
			expectErr: true,
		},
		{
			name:      "missing host",
			in:        "http://",
			expectErr: true,
		},
		{
			name:      "missing scheme",
			in:        "example.com",
			expectErr: true,
		},
		{
			name:      "invalid scheme",
			in:        "tcp://example.com",
			expectErr: true,
		},
		{
			name: "Test happy path",
			in:   "https://u:p@example.com",
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
