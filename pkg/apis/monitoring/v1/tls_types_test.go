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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTLSConfigWithServerNameTemplate(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config *SafeTLSConfig

		err bool
		exp string
	}{
		{
			name: "serverName with valid template",
			config: &SafeTLSConfig{
				ServerName: new(TemplateString(`{{ .Name }}.{{ .Namespace }}.svc.cluster.local.`)),
			},
			exp: "foo.bar.svc.cluster.local.",
		},
		{
			name: "serverName with valid label index template",
			config: &SafeTLSConfig{
				ServerName: new(TemplateString(`{{ index .Labels "tls-server-name" }}`)),
			},
			exp: "servername-from-labels",
		},
		{
			name: "serverName with valid annotation index template",
			config: &SafeTLSConfig{
				ServerName: new(TemplateString(`{{ index .Annotations "tls-server-name" }}`)),
			},
			exp: "servername-from-annotations",
		},
		{
			name: "serverName with invalid template (unclosed action)",
			config: &SafeTLSConfig{
				ServerName: new(TemplateString(`{{ .Name`)),
			},
			err: true,
		},
		{
			name: "serverName with invalid template (unbounded variable)",
			config: &SafeTLSConfig{
				ServerName: new(TemplateString(`{{ .Value }}`)),
			},
			err: true,
		},
		{
			name: "serverName without template syntax",
			config: &SafeTLSConfig{
				ServerName: new(TemplateString("static-server-name.example.com")),
			},
			exp: "static-server-name.example.com",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := &metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
				Labels: map[string]string{
					"tls-server-name": "servername-from-labels",
				},
				Annotations: map[string]string{
					"tls-server-name": "servername-from-annotations",
				},
			}
			err := tc.config.ValidateWithTemplateSupport(o)
			if tc.err {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got %s", err)
			}

			s, err := tc.config.ServerName.Render(o)
			if err != nil {
				t.Fatalf("expected no error but got %s", err)
			}

			if s != tc.exp {
				t.Fatalf("expected %q but got %q", tc.exp, s)
			}
		})
	}
}
