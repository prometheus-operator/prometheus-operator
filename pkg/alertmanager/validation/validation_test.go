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

func TestValidateTemplateURL(t *testing.T) {
	for _, tc := range []struct {
		name string
		url  string
		err  bool
	}{
		{
			name: "plain URL",
			url:  "https://example.com",
		},
		{
			name: "URL with credentials",
			url:  "https://user:pass@example.com/path",
		},
		{
			name: "template with field access",
			url:  `{{ .CommonAnnotations.client }}`,
		},
		{
			name: "template with toUpper",
			url:  `https://example.com/{{ .CommonLabels.namespace | toUpper }}`,
		},
		{
			name: "template with match",
			url:  `{{ if match "^env-" .CommonLabels.namespace }}https://env.example.com{{ else }}https://example.com{{ end }}`,
		},
		{
			name: "template with reReplaceAll",
			url:  `https://example.com/{{ reReplaceAll "[^a-z]" "" .CommonLabels.env }}`,
		},
		{
			name: "template with safeUrl",
			url:  `{{ safeUrl "https://example.com" }}`,
		},
		{
			name: "template with join",
			url:  `https://example.com/{{ join "/" .CommonLabels.path }}`,
		},
		{
			name: "template with routeLabels",
			url:  `https://example.com/{{ routeLabels "severity" }}`,
		},
		{
			name: "invalid URL",
			url:  "not-a-valid-url",
			err:  true,
		},
		{
			name: "invalid template - unclosed action",
			url:  `{{ .CommonAnnotations.client`,
			err:  true,
		},
		{
			name: "invalid template - unknown function",
			url:  `{{ nonexistentFunc .CommonAnnotations.client }}`,
			err:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTemplateURL(tc.url)
			if tc.err {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got: %s", err)
			}
		})
	}
}

func TestValidateLabelValueTemplate(t *testing.T) {
	for _, tc := range []struct {
		name string
		tmpl string
		err  bool
	}{
		{
			name: "empty string",
			tmpl: "",
		},
		{
			name: "plain string without template",
			tmpl: "critical",
		},
		{
			name: "valid template expression",
			tmpl: "{{ .Labels.severity }}",
		},
		{
			name: "valid template with mixed text",
			tmpl: "prefix-{{ .Labels.env }}-suffix",
		},
		{
			name: "valid template with multiple expressions",
			tmpl: "{{ .Labels.team }}/{{ .Labels.env }}",
		},
		{
			name: "valid template accessing nested fields",
			tmpl: "{{ .CommonAnnotations.summary }}",
		},
		{
			name: "valid template with toUpper",
			tmpl: "{{ .Labels.env | toUpper }}",
		},
		{
			name: "valid template with match",
			tmpl: `{{ if match "^prod" .Labels.env }}prod{{ else }}nonprod{{ end }}`,
		},
		{
			name: "invalid template - unclosed action",
			tmpl: "{{ .Labels.severity",
			err:  true,
		},
		{
			name: "invalid template - bad function",
			tmpl: "{{ nonexistentFunc .Labels.severity }}",
			err:  true,
		},
		{
			name: "invalid template - unclosed action at end",
			tmpl: "text {{ .Labels.foo",
			err:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateLabelValueTemplate(tc.tmpl)
			if tc.err {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got: %s", err)
			}
		})
	}
}
