// Copyright 2018 The prometheus-operator Authors
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
	"encoding/json"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMarshallServiceMonitor(t *testing.T) {
	sm := &ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"group": "group1",
			},
		},
		Spec: ServiceMonitorSpec{
			NamespaceSelector: NamespaceSelector{
				MatchNames: []string{"test"},
			},
			Endpoints: []Endpoint{
				{
					Port: "metric",
				},
			},
		},
	}
	expected := `{"metadata":{"name":"test","namespace":"default","creationTimestamp":null,"labels":{"group":"group1"}},"spec":{"endpoints":[{"port":"metric","bearerTokenSecret":{"key":""}}],"selector":{},"namespaceSelector":{"matchNames":["test"]}}}`

	r, err := json.Marshal(sm)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	rs := string(r)
	if rs != expected {
		t.Fatalf("Got %s expected: %s ", rs, expected)
	}
}

func TestValidateSecretOrConfigMap(t *testing.T) {
	for _, good := range []SecretOrConfigMap{
		SecretOrConfigMap{},
		SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
		SecretOrConfigMap{ConfigMap: &v1.ConfigMapKeySelector{}},
	} {
		if err := good.Validate(); err != nil {
			t.Errorf("expected validation of %+v not to fail, err: %s", good, err)
		}
	}

	bad := SecretOrConfigMap{Secret: &v1.SecretKeySelector{}, ConfigMap: &v1.ConfigMapKeySelector{}}
	if err := bad.Validate(); err == nil {
		t.Errorf("expected validation of %+v to fail, but got no error", bad)
	}
}

func TestValidateSafeTLSConfig(t *testing.T) {
	for _, tc := range []struct {
		config *SafeTLSConfig
		err    bool
	}{
		{
			// CA, Cert, and KeySecret.
			config: &SafeTLSConfig{
				CA:        SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				Cert:      SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				KeySecret: &v1.SecretKeySelector{},
			},
			err: false,
		},
		{
			// Without CA cert.
			config: &SafeTLSConfig{
				Cert:      SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				KeySecret: &v1.SecretKeySelector{},
			},
			err: false,
		},
		{
			// Without Cert.
			config: &SafeTLSConfig{
				CA:        SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				KeySecret: &v1.SecretKeySelector{},
			},
			err: true,
		},
		{
			// Without KeySecret.
			config: &SafeTLSConfig{
				CA:   SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				Cert: SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
			},
			err: true,
		},
	} {
		t.Run("", func(t *testing.T) {
			err := tc.config.Validate()
			if tc.err && err == nil {
				t.Fatalf("expected validation of %+v to fail, but got no error", tc.config)
			}
			if !tc.err && err != nil {
				t.Fatalf("expected validation of %+v not to fail, err: %s", tc.config, err)
			}
		})
	}
}

func TestValidateTLSConfig(t *testing.T) {
	for _, tc := range []struct {
		config *TLSConfig
		err    bool
	}{
		{
			// CAFile, CertFile, and KeyFile.
			config: &TLSConfig{
				CAFile:   "cafile",
				CertFile: "certfile",
				KeyFile:  "keyfile",
			},
			err: false,
		},
		{
			// Without CAFile.
			config: &TLSConfig{
				CertFile: "certfile",
				KeyFile:  "keyfile",
			},
			err: false,
		},
		{
			// Without CertFile.
			config: &TLSConfig{
				CAFile:  "cafile",
				KeyFile: "keyfile",
			},
			err: true,
		},
		{
			// Without KeyFile.
			config: &TLSConfig{
				CAFile:   "cafile",
				CertFile: "certfile",
			},
			err: true,
		},
		{
			// CertSecret and KeyFile.
			config: &TLSConfig{
				CAFile:  "cafile",
				KeyFile: "keyfile",
				SafeTLSConfig: SafeTLSConfig{
					Cert: SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				},
			},
			err: false,
		},
		{
			// CertFile and KeySecret.
			config: &TLSConfig{
				CAFile:   "cafile",
				CertFile: "certfile",
				SafeTLSConfig: SafeTLSConfig{
					KeySecret: &v1.SecretKeySelector{},
				},
			},
			err: false,
		},
		{
			// CA, Cert, and KeySecret.
			config: &TLSConfig{
				SafeTLSConfig: SafeTLSConfig{
					CA:        SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
					Cert:      SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
					KeySecret: &v1.SecretKeySelector{},
				},
			},
			err: false,
		},
		{
			// Without CA and CAFile.
			config: &TLSConfig{
				SafeTLSConfig: SafeTLSConfig{
					Cert:      SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
					KeySecret: &v1.SecretKeySelector{},
				},
			},
			err: false,
		},
		{
			// Without Cert and CertFile.
			config: &TLSConfig{
				SafeTLSConfig: SafeTLSConfig{
					CA:        SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
					KeySecret: &v1.SecretKeySelector{},
				},
			},
			err: true,
		},
		{
			// Without KeySecret and KeyFile.
			config: &TLSConfig{
				SafeTLSConfig: SafeTLSConfig{
					CA:   SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
					Cert: SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				},
			},
			err: true,
		},
	} {
		t.Run("", func(t *testing.T) {
			err := tc.config.Validate()
			if tc.err && err == nil {
				t.Fatalf("expected validation of %+v to fail, but got no error", tc.config)
			}
			if !tc.err && err != nil {
				t.Fatalf("expected validation of %+v not to fail, err: %s", tc.config, err)
			}
		})
	}
}
