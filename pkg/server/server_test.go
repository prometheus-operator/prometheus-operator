// Copyright 2020 The prometheus-operator Authors
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

package server

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

func TestConvertTLSConfig(t *testing.T) {
	for _, tc := range []struct {
		name string
		c    TLSConfig

		err    bool
		assert func(*testing.T, *tls.Config)
	}{
		{
			name: "TLS disabled",
			c:    TLSConfig{},

			assert: func(t *testing.T, c *tls.Config) {
				require.Nil(t, c)
			},
		},
		{
			name: "error when client CA is configured without cert/key",
			c: TLSConfig{
				Enabled:      true,
				ClientCAFile: "ca.crt",
			},
			err: true,
		},
		{
			name: "TLS enabled but no cert/key provided",
			c: TLSConfig{
				Enabled: true,
			},

			assert: func(t *testing.T, c *tls.Config) {
				require.Nil(t, c)
			},
		},
		{
			name: "error when invalid TLS version",
			c: TLSConfig{
				Enabled:    true,
				CertFile:   "server.crt",
				KeyFile:    "server.key",
				MinVersion: "VersionTLSXX",
			},

			err: true,
		},
		{
			name: "valid TLS config with default version",
			c: TLSConfig{
				Enabled:  true,
				CertFile: "server.crt",
				KeyFile:  "server.key",
			},

			assert: func(t *testing.T, c *tls.Config) {
				require.NotNil(t, c)
				require.Equal(t, tls.VersionTLS12, int(c.MinVersion))
			},
		},
		{
			name: "error when invalid cipher suite",
			c: TLSConfig{
				Enabled:      true,
				CertFile:     "server.crt",
				KeyFile:      "server.key",
				CipherSuites: operator.StringSet(map[string]struct{}{"foo": {}}),
			},

			err: true,
		},
		{
			name: "valid cipher suite",
			c: TLSConfig{
				Enabled:      true,
				CertFile:     "server.crt",
				KeyFile:      "server.key",
				CipherSuites: operator.StringSet(map[string]struct{}{"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA": {}}),
			},

			assert: func(t *testing.T, c *tls.Config) {
				require.NotNil(t, c)
				require.Equal(t, []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA}, c.CipherSuites)
			},
		},
		{
			name: "no curves defined",
			c: TLSConfig{
				Enabled:  true,
				CertFile: "server.crt",
				KeyFile:  "server.key",
				Curves:   operator.StringSet{},
			},

			assert: func(t *testing.T, c *tls.Config) {
				require.NotNil(t, c)
				require.Len(t, c.CurvePreferences, 0)
			},
		},
		{
			name: "single valid curve",
			c: TLSConfig{
				Enabled:  true,
				CertFile: "server.crt",
				KeyFile:  "server.key",
				Curves:   operator.StringSet(map[string]struct{}{"CurveP256": {}}),
			},

			assert: func(t *testing.T, c *tls.Config) {
				require.NotNil(t, c)
				require.Equal(t, []tls.CurveID{tls.CurveP256}, c.CurvePreferences)
			},
		},
		{
			name: "multiple valid curves",
			c: TLSConfig{
				Enabled:  true,
				CertFile: "server.crt",
				KeyFile:  "server.key",
				Curves:   operator.StringSet(map[string]struct{}{"CurveP256": {}, "X25519": {}}),
			},

			assert: func(t *testing.T, c *tls.Config) {
				require.NotNil(t, c)
				require.Len(t, c.CurvePreferences, 2)
				require.Contains(t, c.CurvePreferences, tls.CurveP256)
				require.Contains(t, c.CurvePreferences, tls.X25519)
			},
		},
		{
			name: "all supported curves",
			c: TLSConfig{
				Enabled:  true,
				CertFile: "server.crt",
				KeyFile:  "server.key",
				Curves: operator.StringSet(map[string]struct{}{
					"CurveP256":      {},
					"CurveP384":      {},
					"CurveP521":      {},
					"X25519":         {},
					"X25519MLKEM768": {},
				}),
			},

			assert: func(t *testing.T, c *tls.Config) {
				require.NotNil(t, c)
				require.Len(t, c.CurvePreferences, 5)
				require.Contains(t, c.CurvePreferences, tls.CurveP256)
				require.Contains(t, c.CurvePreferences, tls.CurveP384)
				require.Contains(t, c.CurvePreferences, tls.CurveP521)
				require.Contains(t, c.CurvePreferences, tls.X25519)
				require.Contains(t, c.CurvePreferences, tls.X25519MLKEM768)
			},
		},
		{
			name: "error when invalid curve",
			c: TLSConfig{
				Enabled:  true,
				CertFile: "server.crt",
				KeyFile:  "server.key",
				Curves:   operator.StringSet(map[string]struct{}{"InvalidCurve": {}}),
			},

			err: true,
		},
		{
			name: "error when mix of valid and invalid curves",
			c: TLSConfig{
				Enabled:  true,
				CertFile: "server.crt",
				KeyFile:  "server.key",
				Curves:   operator.StringSet(map[string]struct{}{"CurveP256": {}, "InvalidCurve": {}}),
			},

			err: true,
		},
		{
			name: "error when empty string curve name",
			c: TLSConfig{
				Enabled:  true,
				CertFile: "server.crt",
				KeyFile:  "server.key",
				Curves:   operator.StringSet(map[string]struct{}{"": {}}),
			},

			err: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, err := tc.c.Convert(nil)
			if tc.err {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tc.assert != nil {
				tc.assert(t, c)
			}
		})
	}
}
