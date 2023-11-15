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
		c TLSConfig

		err    bool
		assert func(*testing.T, *tls.Config)
	}{
		{
			c: TLSConfig{},

			assert: func(t *testing.T, c *tls.Config) {
				require.Nil(t, c)
			},
		},
		{
			c: TLSConfig{
				Enabled:      true,
				ClientCAFile: "ca.crt",
			},
			err: true,
		},
		{
			c: TLSConfig{
				Enabled: true,
			},

			assert: func(t *testing.T, c *tls.Config) {
				require.Nil(t, c)
			},
		},
		{
			c: TLSConfig{
				Enabled:    true,
				CertFile:   "server.crt",
				KeyFile:    "server.key",
				MinVersion: "VersionTLSXX",
			},

			err: true,
		},
		{
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
			c: TLSConfig{
				Enabled:      true,
				CertFile:     "server.crt",
				KeyFile:      "server.key",
				CipherSuites: operator.StringSet(map[string]struct{}{"foo": {}}),
			},

			err: true,
		},
		{
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
	} {
		t.Run("", func(t *testing.T) {
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
