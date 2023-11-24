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
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestValidateSecretOrConfigMap(t *testing.T) {
	for _, good := range []SecretOrConfigMap{
		{},
		{Secret: &v1.SecretKeySelector{}},
		{ConfigMap: &v1.ConfigMapKeySelector{}},
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
		name   string
		config *SafeTLSConfig

		err bool
	}{
		{
			name: "ca, cert and keySecret",
			config: &SafeTLSConfig{
				CA:        SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				Cert:      SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				KeySecret: &v1.SecretKeySelector{},
			},
			err: false,
		},
		{
			name: "cert and keySecret",
			config: &SafeTLSConfig{
				Cert:      SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				KeySecret: &v1.SecretKeySelector{},
			},
			err: false,
		},
		{
			name: "ca and keySecret",
			config: &SafeTLSConfig{
				CA:        SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				KeySecret: &v1.SecretKeySelector{},
			},
			err: true,
		},
		{
			name: "ca and cert",
			config: &SafeTLSConfig{
				CA: SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{},
				},
				Cert: SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{},
				},
			},
			err: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
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

func TestValidateTLSConfig(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config *TLSConfig

		err bool
	}{
		{
			name: "caFile, certFile and keyFile",
			config: &TLSConfig{
				CAFile:   "cafile",
				CertFile: "certfile",
				KeyFile:  "keyfile",
			},
			err: false,
		},
		{
			name: "certFile and keyFile",
			config: &TLSConfig{
				CertFile: "certfile",
				KeyFile:  "keyfile",
			},
			err: false,
		},
		{
			name: "caFile and keyFile",
			config: &TLSConfig{
				CAFile:  "cafile",
				KeyFile: "keyfile",
			},
			err: true,
		},
		{
			name: "caFile and certFile",
			config: &TLSConfig{
				CAFile:   "cafile",
				CertFile: "certfile",
			},
			err: true,
		},
		{
			name: "caFile, cert and keyFile",
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
			name: "caFile, certFile and keySecret",
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
			name: "ca, cert and keySecret",
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
			name: "cert and keySecret",
			config: &TLSConfig{
				SafeTLSConfig: SafeTLSConfig{
					Cert:      SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
					KeySecret: &v1.SecretKeySelector{},
				},
			},
			err: false,
		},
		{
			name: "ca and keySecret",
			config: &TLSConfig{
				SafeTLSConfig: SafeTLSConfig{
					CA:        SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
					KeySecret: &v1.SecretKeySelector{},
				},
			},
			err: true,
		},
		{
			name: "ca and cert",
			config: &TLSConfig{
				SafeTLSConfig: SafeTLSConfig{
					CA:   SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
					Cert: SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				},
			},
			err: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
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

func TestValidateAuthorization(t *testing.T) {
	creds := &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: "key"},
	}
	for _, tc := range []struct {
		name   string
		config *Authorization
		err    bool
	}{
		{
			name: "minimal example",
			config: &Authorization{
				SafeAuthorization: SafeAuthorization{
					Credentials: creds,
				},
			},
			err: false,
		},
		{
			name: "explicit Bearer type",
			config: &Authorization{
				SafeAuthorization: SafeAuthorization{
					Type:        "Bearer",
					Credentials: creds,
				},
			},
			err: false,
		},
		{
			name: "custom type",
			config: &Authorization{
				SafeAuthorization: SafeAuthorization{
					Type:        "token",
					Credentials: creds,
				},
			},
			err: false,
		},
		{
			name: "type Basic not allowed",
			config: &Authorization{
				SafeAuthorization: SafeAuthorization{
					Type:        "Basic",
					Credentials: creds,
				},
			},
			err: true,
		},
		{
			name: "conflict between credentials and credentials_file",
			config: &Authorization{
				SafeAuthorization: SafeAuthorization{
					Credentials: creds,
				},
				CredentialsFile: "/some/file",
			},
			err: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.err && err == nil {
				t.Fatal("expected error but got none")
			}

			if !tc.err && err != nil {
				t.Fatalf("expected no error but got: %s", err)
			}
		})
	}
}
