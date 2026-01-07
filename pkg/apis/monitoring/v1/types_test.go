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
	"k8s.io/apimachinery/pkg/api/resource"
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
		{
			name: "maxVersion more than minVersion",
			config: &SafeTLSConfig{
				MinVersion: func(v TLSVersion) *TLSVersion { return &v }(TLSVersion10),
				MaxVersion: func(v TLSVersion) *TLSVersion { return &v }(TLSVersion13),
			},
			err: false,
		},
		{
			name: "maxVersion equal to minVersion",
			config: &SafeTLSConfig{
				MinVersion: func(v TLSVersion) *TLSVersion { return &v }(TLSVersion13),
				MaxVersion: func(v TLSVersion) *TLSVersion { return &v }(TLSVersion13),
			},
			err: false,
		},
		{
			name: "maxVersion is less than minVersion",
			config: &SafeTLSConfig{
				MinVersion: func(v TLSVersion) *TLSVersion { return &v }(TLSVersion13),
				MaxVersion: func(v TLSVersion) *TLSVersion { return &v }(TLSVersion10),
			},
			err: true,
		},
		{
			name:   "SafeTLSConfig nil",
			config: nil,
			err:    false,
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
				TLSFilesConfig: TLSFilesConfig{
					CAFile:   "cafile",
					CertFile: "certfile",
					KeyFile:  "keyfile",
				},
			},
			err: false,
		},
		{
			name: "certFile and keyFile",
			config: &TLSConfig{
				TLSFilesConfig: TLSFilesConfig{
					CertFile: "certfile",
					KeyFile:  "keyfile",
				},
			},
			err: false,
		},
		{
			name: "caFile and keyFile",
			config: &TLSConfig{
				TLSFilesConfig: TLSFilesConfig{
					CAFile:  "cafile",
					KeyFile: "keyfile",
				},
			},
			err: true,
		},
		{
			name: "caFile and certFile",
			config: &TLSConfig{
				TLSFilesConfig: TLSFilesConfig{
					CAFile:   "cafile",
					CertFile: "certfile",
				},
			},
			err: true,
		},
		{
			name: "caFile, cert and keyFile",
			config: &TLSConfig{
				TLSFilesConfig: TLSFilesConfig{
					CAFile:  "cafile",
					KeyFile: "keyfile",
				},
				SafeTLSConfig: SafeTLSConfig{
					Cert: SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				},
			},
			err: false,
		},
		{
			name: "caFile, certFile and keySecret",
			config: &TLSConfig{
				TLSFilesConfig: TLSFilesConfig{
					CAFile:   "cafile",
					CertFile: "certfile",
				},
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
		{
			name: "maxVersion more than minVersion",
			config: &TLSConfig{
				SafeTLSConfig: SafeTLSConfig{
					MinVersion: func(v TLSVersion) *TLSVersion { return &v }(TLSVersion10),
					MaxVersion: func(v TLSVersion) *TLSVersion { return &v }(TLSVersion13),
				},
			},
			err: false,
		},
		{
			name: "maxVersion equal to minVersion",
			config: &TLSConfig{
				SafeTLSConfig: SafeTLSConfig{
					MinVersion: func(v TLSVersion) *TLSVersion { return &v }(TLSVersion13),
					MaxVersion: func(v TLSVersion) *TLSVersion { return &v }(TLSVersion13),
				},
			},
			err: false,
		},
		{
			name: "maxVersion is less than minVersion",
			config: &TLSConfig{
				SafeTLSConfig: SafeTLSConfig{
					MinVersion: func(v TLSVersion) *TLSVersion { return &v }(TLSVersion13),
					MaxVersion: func(v TLSVersion) *TLSVersion { return &v }(TLSVersion10),
				},
			},
			err: true,
		},
		{
			name:   "tlsconfig nil",
			config: nil,
			err:    false,
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

func TestValidateWebTlsConfig(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config *WebTLSConfig
		err    bool
	}{
		{
			name: "caFile, certFile and keyFile",
			config: &WebTLSConfig{
				ClientCAFile: func(s string) *string { return &s }("cafile"),
				CertFile:     func(s string) *string { return &s }("certfile"),
				KeyFile:      func(s string) *string { return &s }("keyfile"),
			},
		},
		{
			name: "certFile and keyFile",
			config: &WebTLSConfig{
				CertFile: func(s string) *string { return &s }("certfile"),
				KeyFile:  func(s string) *string { return &s }("keyfile"),
			},
		},
		{
			name: "caFile and keyFile",
			config: &WebTLSConfig{
				ClientCAFile: func(s string) *string { return &s }("cafile"),
				KeyFile:      func(s string) *string { return &s }("keyfile"),
			},
			err: true,
		},
		{
			name: "caFile and certFile",
			config: &WebTLSConfig{
				ClientCAFile: func(s string) *string { return &s }("cafile"),
				CertFile:     func(s string) *string { return &s }("certfile"),
			},
			err: true,
		},
		{
			name: "caFile, cert and keyFile",
			config: &WebTLSConfig{
				ClientCAFile: func(s string) *string { return &s }("cafile"),
				KeyFile:      func(s string) *string { return &s }("keyfile"),
				Cert:         SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
			},
		},
		{
			name: "caFile, certFile and keySecret",
			config: &WebTLSConfig{
				ClientCAFile: func(s string) *string { return &s }("cafile"),
				CertFile:     func(s string) *string { return &s }("certfile"),
				KeySecret: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "test-secret",
					},
					Key: "tls.key",
				},
			},
		},
		{
			name: "ca, cert and keySecret",
			config: &WebTLSConfig{
				Cert:     SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				ClientCA: SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				KeySecret: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "test-secret",
					},
					Key: "tls.key",
				},
			},
		},
		{
			name: "cert and keySecret",
			config: &WebTLSConfig{
				ClientCA: SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				KeySecret: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "test-secret",
					},
					Key: "tls.key",
				},
			},
			err: true,
		},
		{
			name: "ca and cert",
			config: &WebTLSConfig{
				ClientCA: SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				Cert:     SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
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

func TestValidateOAuth2(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config *OAuth2
		err    bool
	}{
		{
			name: "SafeTLSConfig nil",
			config: &OAuth2{
				ClientID:     SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				ClientSecret: v1.SecretKeySelector{},
				TokenURL:     "http://tokenurl.org",
				TLSConfig:    nil,
			},
			err: false,
		},
		{
			name: "SafeTLSConfig not nil",
			config: &OAuth2{
				ClientID:     SecretOrConfigMap{Secret: &v1.SecretKeySelector{}},
				ClientSecret: v1.SecretKeySelector{},
				TokenURL:     "http://tokenurl.org",
				TLSConfig: &SafeTLSConfig{
					MinVersion: func(v TLSVersion) *TLSVersion { return &v }(TLSVersion10),
					MaxVersion: func(v TLSVersion) *TLSVersion { return &v }(TLSVersion13),
				},
			},
			err: false,
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

func TestValidateTracingConfig(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config *TracingConfig
		err    bool
	}{
		{
			name: "TLSConfig nil",
			config: &TracingConfig{
				TLSConfig: nil,
			},
			err: false,
		},
		{
			name: "SamplingFraction simple value",
			config: &TracingConfig{
				SamplingFraction: func(v resource.Quantity) *resource.Quantity { return &v }(resource.MustParse("0.56")),
			},
			err: false,
		},
		{
			name: "SamplingFraction > 1",
			config: &TracingConfig{
				SamplingFraction: resource.NewQuantity(10, resource.DecimalSI),
			},
			err: true,
		},
		{
			name: "SamplingFraction < 0",
			config: &TracingConfig{
				SamplingFraction: resource.NewQuantity(-1, resource.DecimalSI),
			},
			err: true,
		},
		{
			name: "SamplingFraction == 0",
			config: &TracingConfig{
				SamplingFraction: resource.NewQuantity(0, resource.DecimalSI),
			},
			err: false,
		},
		{
			name: "SamplingFraction == 1",
			config: &TracingConfig{
				SamplingFraction: resource.NewQuantity(1, resource.DecimalSI),
			},
			err: false,
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
