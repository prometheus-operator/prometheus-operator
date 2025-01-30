// Copyright 2024 The prometheus-operator Authors
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

package clustertlsconfig

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"k8s.io/utils/ptr"
)

func addServerTLSConfigToYaml(c ClusterTLSConfig, cfg yaml.MapSlice) yaml.MapSlice {
	tls := c.serverTLSConfig
	if tls == nil {
		return nil
	}

	mtlsServerConfig := yaml.MapSlice{}
	tlsRefs := c.serverTLSReferences

	switch {
	case ptr.Deref(tls.KeyFile, "") != "":
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "key_file", Value: *tls.KeyFile})
	case tlsRefs.GetKeyMountPath() != "":
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "key_file", Value: filepath.Join(tlsRefs.GetKeyMountPath(), tlsRefs.GetKeyFilename())})
	}

	switch {
	case ptr.Deref(tls.CertFile, "") != "":
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "cert_file", Value: *tls.CertFile})
	case tlsRefs.GetCertMountPath() != "":
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "cert_file", Value: filepath.Join(tlsRefs.GetCertMountPath(), tlsRefs.GetCertFilename())})
	}

	if ptr.Deref(tls.ClientAuthType, "") != "" {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "client_auth_type",
			Value: *tls.ClientAuthType,
		})
	}

	switch {
	case ptr.Deref(tls.ClientCAFile, "") != "":
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "client_ca_file", Value: *tls.ClientCAFile})
	case tlsRefs.GetCAMountPath() != "":
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "client_ca_file", Value: filepath.Join(tlsRefs.GetCAMountPath(), tlsRefs.GetCAFilename())})
	}

	if ptr.Deref(tls.MinVersion, "") != "" {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "min_version",
			Value: *tls.MinVersion,
		})
	}

	if ptr.Deref(tls.MaxVersion, "") != "" {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "max_version",
			Value: *tls.MaxVersion,
		})
	}

	if len(tls.CipherSuites) != 0 {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "cipher_suites",
			Value: tls.CipherSuites,
		})
	}

	if tls.PreferServerCipherSuites != nil {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "prefer_server_cipher_suites",
			Value: tls.PreferServerCipherSuites,
		})
	}

	if len(tls.CurvePreferences) != 0 {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "curve_preferences",
			Value: tls.CurvePreferences,
		})
	}

	return append(cfg, yaml.MapItem{Key: "tls_server_config", Value: mtlsServerConfig})
}

func addClientTLSConfigToYaml(c ClusterTLSConfig, cfg yaml.MapSlice) yaml.MapSlice {
	tls := c.clientTLSConfig
	if tls == nil {
		return nil
	}

	mtlsClientConfig := yaml.MapSlice{}
	tlsRefs := c.clientTLSReferences

	if keyPath := tlsRefs.GetKeyMountPath(); keyPath != "" {
		mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{Key: "key_file", Value: fmt.Sprintf("%s/%s", keyPath, tlsRefs.GetKeyFilename())})
	}

	if certPath := tlsRefs.GetCertMountPath(); certPath != "" {
		mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{Key: "cert_file", Value: fmt.Sprintf("%s/%s", certPath, tlsRefs.GetCertFilename())})
	}

	if caPath := tlsRefs.GetCAMountPath(); caPath != "" {
		mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{Key: "ca_file", Value: fmt.Sprintf("%s/%s", caPath, tlsRefs.GetCAFilename())})
	}

	if serverName := tls.ServerName; serverName != nil {
		mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{Key: "server_name", Value: serverName})
	}

	if tls.InsecureSkipVerify != nil {
		mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{
			Key:   "insecure_skip_verify",
			Value: tls.InsecureSkipVerify,
		})
	}

	return append(cfg, yaml.MapItem{Key: "tls_client_config", Value: mtlsClientConfig})
}
