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

import "gopkg.in/yaml.v2"

func addServerTLSConfigToYaml(c ClusterTLSConfig, cfg yaml.MapSlice) yaml.MapSlice {
	serverTLSConfig := c.serverTLSConfig
	if serverTLSConfig == nil {
		return nil
	}

	mtlsServerConfig := yaml.MapSlice{}
	serverTLSCredentials := c.serverTLSCredentials

	switch {
	case serverTLSCredentials.GetKeyFile() != "":
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "key_file", Value: serverTLSCredentials.GetKeyFile()})
	case serverTLSCredentials.GetKeyMountPath() != "":
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "key_file", Value: serverTLSCredentials.GetKeyMountPath()})
	}

	switch {
	case serverTLSCredentials.GetCertFile() != "":
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "cert_file", Value: serverTLSCredentials.GetCertFile()})
	case serverTLSCredentials.GetCertMountPath() != "":
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "cert_file", Value: serverTLSCredentials.GetCertMountPath()})
	}

	if serverTLSConfig.ClientAuthType != "" {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "client_auth_type",
			Value: serverTLSConfig.ClientAuthType,
		})
	}

	switch {
	case serverTLSCredentials.GetCAFile() != "":
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "ca_file", Value: serverTLSCredentials.GetCAFile()})
	case serverTLSCredentials.GetCAMountPath() != "":
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "ca_file", Value: serverTLSCredentials.GetCAMountPath()})
	}

	if serverTLSConfig.MinVersion != "" {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "min_version",
			Value: serverTLSConfig.MinVersion,
		})
	}

	if serverTLSConfig.MaxVersion != "" {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "max_version",
			Value: serverTLSConfig.MaxVersion,
		})
	}

	if len(serverTLSConfig.CipherSuites) != 0 {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "cipher_suites",
			Value: serverTLSConfig.CipherSuites,
		})
	}

	if serverTLSConfig.PreferServerCipherSuites != nil {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "prefer_server_cipher_suites",
			Value: serverTLSConfig.PreferServerCipherSuites,
		})
	}

	if len(serverTLSConfig.CurvePreferences) != 0 {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "curve_preferences",
			Value: serverTLSConfig.CurvePreferences,
		})
	}

	return append(cfg, yaml.MapItem{Key: "tls_server_config", Value: mtlsServerConfig})
}

func addClientTLSConfigToYaml(c ClusterTLSConfig, cfg yaml.MapSlice) yaml.MapSlice {
	clientTLSConfig := c.clientTLSConfig
	if clientTLSConfig == nil {
		return nil
	}

	clientTLSCredentials := c.clientTLSCredentials
	mtlsClientConfig := yaml.MapSlice{}

	if keyPath := clientTLSCredentials.GetKeyMountPath(); keyPath != "" {
		mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{Key: "key_file", Value: clientTLSCredentials.GetKeyMountPath()})
	}

	if certPath := clientTLSCredentials.GetCertMountPath(); certPath != "" {
		mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{Key: "cert_file", Value: certPath})
	}

	if caPath := clientTLSCredentials.GetCAMountPath(); caPath != "" {
		mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{Key: "ca_file", Value: caPath})
	}

	if serverName := clientTLSConfig.ServerName; serverName != nil {
		mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{Key: "server_name", Value: serverName})
	}

	if clientTLSConfig.InsecureSkipVerify != nil {
		mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{
			Key:   "insecure_skip_verify",
			Value: clientTLSConfig.InsecureSkipVerify,
		})
	}

	return append(cfg, yaml.MapItem{Key: "tls_client_config", Value: mtlsClientConfig})
}
