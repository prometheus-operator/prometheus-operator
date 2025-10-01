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
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	webconfig "github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
)

const (
	cliFlag            = "cluster.tls-config"
	volumeName         = "cluster-tls-config"
	serverVolumePrefix = "cluster-tls-server-config-"
	clientVolumePrefix = "cluster-tls-client-config-"
	serverTLSCredDir   = "server_tls"
	clientTLSCredDir   = "client_tls"

	// ConfigFileKey is the secret's key containing the YAML configuration.
	ConfigFileKey = "cluster-tls-config.yaml"
)

// Config is the Alertmanager cluster's mTLS configuration.
//
// Config can make a secret which holds the cluster configuration as well as
// volumes and volume mounts for referencing the secret and the necessary TLS
// credentials.
type Config struct {
	clusterTLSConfig    *monitoringv1.ClusterTLSConfig
	serverTLSReferences *webconfig.TLSReferences
	clientTLSReferences *webconfig.TLSReferences
	mountingDir         string
	secretName          string
}

// New creates a new ClusterTLSConfig.
// All volumes related to the cluster TLS config will be mounted via the `mountingDir`.
// The Secret where the cluster TLS config will be stored will be named `secretName`.
// All volumes containing TLS credentials related to cluster TLS configuration will be prefixed with "cluster-tls-server-config-"
// or "cluster-tls-client-config-" respectively, for server and client credentials.
func New(mountingDir string, a *monitoringv1.Alertmanager) (*Config, error) {
	clusterTLSConfig := a.Spec.ClusterTLS
	secretName := fmt.Sprintf("alertmanager-%s-cluster-tls-config", a.Name)

	if clusterTLSConfig == nil {
		return &Config{
			mountingDir: mountingDir,
			secretName:  secretName,
		}, nil
	}

	var (
		clientTLSCreds *webconfig.TLSReferences
		serverTLSCreds *webconfig.TLSReferences
	)

	serverTLSConfig := clusterTLSConfig.ServerTLS
	if err := serverTLSConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid server TLS configuration: %w", err)
	}

	clientTLSConfig := clusterTLSConfig.ClientTLS
	if err := clientTLSConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid client TLS configuration: %w", err)
	}
	if reflect.ValueOf(clientTLSConfig.Cert).IsZero() {
		return nil, errors.New("invalid client TLS configuration: certificate is required")
	}

	serverTLSCreds = webconfig.NewTLSReferences(path.Join(mountingDir, serverTLSCredDir), serverTLSConfig.KeySecret, serverTLSConfig.Cert, serverTLSConfig.ClientCA)
	clientTLSCreds = webconfig.NewTLSReferences(path.Join(mountingDir, clientTLSCredDir), *clientTLSConfig.KeySecret, clientTLSConfig.Cert, clientTLSConfig.CA)

	return &Config{
		clusterTLSConfig:    clusterTLSConfig,
		serverTLSReferences: serverTLSCreds,
		clientTLSReferences: clientTLSCreds,
		mountingDir:         mountingDir,
		secretName:          secretName,
	}, nil
}

// GetMountParameters returns volumes and volume mounts referencing the cluster TLS config file
// and the associated TLS credentials.
// In addition, GetMountParameters returns a cluster.tls-config command line option pointing
// to the cluster TLS config file in the volume mount.
// All TLS credentials related to cluster TLS configuration will be prefixed with "cluster-tls-server-config-"
// or "cluster-tls-client-config-" respectively, for server and client credentials.
// The server and client TLS credentials are mounted in different paths: ~/{mountingDir}/server-tls/
// and ~/{mountingDir}/client-tls/ respectively.
func (c Config) GetMountParameters() (*monitoringv1.Argument, []v1.Volume, []v1.VolumeMount, error) {
	destinationPath := path.Join(c.mountingDir, ConfigFileKey)

	var arg *monitoringv1.Argument
	// Only return an argument if the cluster TLS config and it's server component are defined.
	if c.clusterTLSConfig != nil {
		arg = c.makeArg(destinationPath)
	}

	var volumes []v1.Volume
	cfgVolume := c.makeVolume()
	volumes = append(volumes, cfgVolume)

	var mounts []v1.VolumeMount
	cfgMount := c.makeVolumeMount(destinationPath)
	mounts = append(mounts, cfgMount)

	if c.serverTLSReferences != nil {
		servertlsVolumes, servertlsMounts, err := c.serverTLSReferences.GetMountParameters(serverVolumePrefix)
		if err != nil {
			return nil, nil, nil, err
		}
		volumes = append(volumes, servertlsVolumes...)
		mounts = append(mounts, servertlsMounts...)
	}

	if c.clientTLSReferences != nil {
		clienttlsVolumes, clienttlsMounts, err := c.clientTLSReferences.GetMountParameters(clientVolumePrefix)
		if err != nil {
			return nil, nil, nil, err
		}
		volumes = append(volumes, clienttlsVolumes...)
		mounts = append(mounts, clienttlsMounts...)
	}

	return arg, volumes, mounts, nil
}

// ClusterTLSConfiguration create or update a Kubernetes secret with the data for the cluster TLS config file.
// The format of the cluster TLS config file is available in the official prometheus documentation:
// https://github.com/prometheus/alertmanager/blob/main/docs/https.md#gossip-traffic/
func (c Config) ClusterTLSConfiguration() ([]byte, error) {
	if c.clusterTLSConfig == nil {
		return []byte{}, nil
	}
	data, err := c.generateConfigFileContents()
	if err != nil {
		return nil, err
	}
	return data, nil
}

// generateConfigFileContents() generates the contents of cluster-tls-config.yaml
// from the Config in the form of an array of bytes.
func (c Config) generateConfigFileContents() ([]byte, error) {
	cfg := yaml.MapSlice{}

	cfg = c.addServerTLSConfigToYaml(cfg)
	cfg = c.addClientTLSConfigToYaml(cfg)

	return yaml.Marshal(cfg)
}

// makeArg() returns an argument with the name "cluster.tls-config" with the filePath
// as its value.
func (c Config) makeArg(filePath string) *monitoringv1.Argument {
	return &monitoringv1.Argument{Name: cliFlag, Value: filePath}
}

// makeVolume() creates a Volume with volumeName = "cluster-tls-config" which stores
// the secret which contains the cluster TLS config.
func (c Config) makeVolume() v1.Volume {
	return v1.Volume{
		Name: volumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: c.secretName,
			},
		},
	}
}

// makeVolumeMount() creates a VolumeMount, mounting the cluster_tls_config.yaml SubPath
// to the given filePath.
func (c Config) makeVolumeMount(filePath string) v1.VolumeMount {
	return v1.VolumeMount{
		Name:      volumeName,
		SubPath:   ConfigFileKey,
		ReadOnly:  true,
		MountPath: filePath,
	}
}

func (c Config) GetSecretName() string {
	return c.secretName
}

func (c Config) addServerTLSConfigToYaml(cfg yaml.MapSlice) yaml.MapSlice {
	tls := c.clusterTLSConfig.ServerTLS

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

func (c Config) addClientTLSConfigToYaml(cfg yaml.MapSlice) yaml.MapSlice {
	tls := c.clusterTLSConfig.ClientTLS

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
