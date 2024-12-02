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
	"context"
	"path"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	webconfig "github.com/prometheus-operator/prometheus-operator/pkg/webconfig"
)

const (
	cmdflag            = "cluster.tls-config"
	volumeName         = "cluster-tls-config"
	configFile         = "cluster-tls-config.yaml"
	serverVolumePrefix = "cluster-server-tls-"
	clientVolumePrefix = "cluster-client-tls-"
	serverTLSCredDir   = "server-tls"
	clientTLSCredDir   = "client-tls"
)

// Config is the web configuration for prometheus and alertmanager instance.
//
// Config can make a secret which holds the web config contents, as well as
// volumes and volume mounts for referencing the secret and the
// necessary TLS credentials.
type ClusterTLSConfig struct {
	serverTLSConfig      *monitoringv1.WebTLSConfig
	clientTLSConfig      *monitoringv1.SafeTLSConfig
	serverTLSCredentials *webconfig.TLSCredentials
	clientTLSCredentials *webconfig.TLSCredentials
	mountingDir          string
	secretName           string
}

// New creates a new Config.
func New(mountingDir string, secretName string, clusterTLSConfig monitoringv1.ClusterTLSConfig) (*ClusterTLSConfig, error) {
	serverTLSConfig := clusterTLSConfig.ServerTLS
	if err := serverTLSConfig.Validate(); err != nil {
		return nil, err
	}

	clientTLSConfig := clusterTLSConfig.ClientTLS
	if err := clientTLSConfig.Validate(); err != nil {
		return nil, err
	}

	var serverTLSCreds *webconfig.TLSCredentials
	var clientTLSCreds *webconfig.TLSCredentials

	if serverTLSConfig != nil {
		serverTLSCreds = webconfig.NewTLSCredentials(path.Join(mountingDir, serverTLSCredDir), serverTLSConfig.KeySecret, serverTLSConfig.KeyFile, serverTLSConfig.Cert, serverTLSConfig.CertFile, serverTLSConfig.ClientCA, serverTLSConfig.ClientCAFile)
	}
	if clientTLSConfig != nil {
		clientTLSCreds = webconfig.NewTLSCredentials(path.Join(mountingDir, clientTLSCredDir), *clientTLSConfig.KeySecret, "", clientTLSConfig.Cert, "", clientTLSConfig.CA, "")
	}

	return &ClusterTLSConfig{
		serverTLSConfig:      serverTLSConfig,
		clientTLSConfig:      clientTLSConfig,
		serverTLSCredentials: serverTLSCreds,
		clientTLSCredentials: clientTLSCreds,
		mountingDir:          mountingDir,
		secretName:           secretName,
	}, nil
}

// GetMountParameters returns volumes and volume mounts referencing the mtls config file
// and the associated TLS credentials.
// In addition, GetMountParameters returns a cluster.tls-config command line option pointing
// to the file in the volume mount.
func (c ClusterTLSConfig) GetMountParameters() (monitoringv1.Argument, []v1.Volume, []v1.VolumeMount, error) {
	destinationPath := path.Join(c.mountingDir, configFile)

	var volumes []v1.Volume
	var mounts []v1.VolumeMount

	arg := c.makeArg(destinationPath)
	cfgVolume := c.makeVolume()
	volumes = append(volumes, cfgVolume)

	cfgMount := c.makeVolumeMount(destinationPath)
	mounts = append(mounts, cfgMount)

	// The server and client TLS credentials are mounted in different paths: ~/{mountDir}/{serverTLSCredDir}
	// and ~/{mountDir}/{clientTLSCredDir} respectively.
	if c.serverTLSCredentials != nil {
		servertlsVolumes, servertlsMounts, err := c.serverTLSCredentials.GetMountParameters(serverVolumePrefix)
		if err != nil {
			return monitoringv1.Argument{}, nil, nil, err
		}
		volumes = append(volumes, servertlsVolumes...)
		mounts = append(mounts, servertlsMounts...)
	}

	if c.clientTLSCredentials != nil {
		clienttlsVolumes, clienttlsMounts, err := c.clientTLSCredentials.GetMountParameters(clientVolumePrefix)
		if err != nil {
			return monitoringv1.Argument{}, nil, nil, err
		}
		volumes = append(volumes, clienttlsVolumes...)
		mounts = append(mounts, clienttlsMounts...)
	}

	return arg, volumes, mounts, nil
}

// CreateOrUpdateClusterTLSConfigSecret create or update a Kubernetes secret with the data for the cluster tls config file.
// The format of the cluster tls config file is available in the official prometheus documentation:
// https://github.com/prometheus/alertmanager/blob/main/docs/https.md#gossip-traffic/
func (c ClusterTLSConfig) CreateOrUpdateClusterTLSConfigSecret(ctx context.Context, secretClient clientv1.SecretInterface, s *v1.Secret) error {
	data, err := c.generateClusterTLSConfigFileContents()
	if err != nil {
		return err
	}

	s.Name = c.secretName
	s.Data = map[string][]byte{
		configFile: data,
	}

	return k8sutil.CreateOrUpdateSecret(ctx, secretClient, s)
}

func (c ClusterTLSConfig) generateClusterTLSConfigFileContents() ([]byte, error) {
	cfg := yaml.MapSlice{}

	c.addServerTLSConfigToYaml(cfg)
	c.addClientTLSConfigToYaml(cfg)

	return yaml.Marshal(cfg)
}

func (c ClusterTLSConfig) addServerTLSConfigToYaml(cfg yaml.MapSlice) yaml.MapSlice {
	mtlsServerConfig := yaml.MapSlice{}

	if certPath := c.serverTLSCredentials.GetCertMountPath(); certPath != "" {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "cert_file", Value: certPath})
	}

	if keyPath := c.serverTLSCredentials.GetKeyMountPath(); keyPath != "" {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "key_file", Value: keyPath})
	}

	if c.serverTLSConfig.ClientAuthType != "" {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "client_auth_type",
			Value: c.serverTLSConfig.ClientAuthType,
		})
	}

	if caPath := c.serverTLSCredentials.GetCAMountPath(); caPath != "" {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{Key: "client_ca_file", Value: caPath})
	}

	if c.serverTLSConfig.MinVersion != "" {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "min_version",
			Value: c.serverTLSConfig.MinVersion,
		})
	}

	if c.serverTLSConfig.MaxVersion != "" {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "max_version",
			Value: c.serverTLSConfig.MaxVersion,
		})
	}

	if len(c.serverTLSConfig.CipherSuites) != 0 {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "cipher_suites",
			Value: c.serverTLSConfig.CipherSuites,
		})
	}

	if c.serverTLSConfig.PreferServerCipherSuites != nil {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "prefer_server_cipher_suites",
			Value: c.serverTLSConfig.PreferServerCipherSuites,
		})
	}

	if len(c.serverTLSConfig.CurvePreferences) != 0 {
		mtlsServerConfig = append(mtlsServerConfig, yaml.MapItem{
			Key:   "curve_preferences",
			Value: c.serverTLSConfig.CurvePreferences,
		})
	}

	return append(cfg, yaml.MapItem{Key: "tls_server_config", Value: mtlsServerConfig})
}

func (c ClusterTLSConfig) addClientTLSConfigToYaml(cfg yaml.MapSlice) yaml.MapSlice {

	mtlsClientConfig := yaml.MapSlice{}

	if certPath := c.clientTLSCredentials.GetCertMountPath(); certPath != "" {
		mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{Key: "cert_file", Value: certPath})
	}

	if keyPath := c.clientTLSCredentials.GetKeyMountPath(); keyPath != "" {
		mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{Key: "key_file", Value: keyPath})
	}

	if caPath := c.clientTLSCredentials.GetCAMountPath(); caPath != "" {
		mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{Key: "ca_file", Value: caPath})
	}

	if serverName := c.clientTLSConfig.ServerName; serverName != nil {
		mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{Key: "server_name", Value: serverName})
	}

	mtlsClientConfig = append(mtlsClientConfig, yaml.MapItem{
		Key:   "insecure_skip_verify",
		Value: c.clientTLSConfig.InsecureSkipVerify,
	})
	return append(cfg, yaml.MapItem{Key: "tls_client_config", Value: mtlsClientConfig})
}

func (c ClusterTLSConfig) makeArg(filePath string) monitoringv1.Argument {
	return monitoringv1.Argument{Name: "cluster.tls-config", Value: filePath}
}

func (c ClusterTLSConfig) makeVolume() v1.Volume {
	return v1.Volume{
		Name: volumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: c.secretName,
			},
		},
	}
}

func (c ClusterTLSConfig) makeVolumeMount(filePath string) v1.VolumeMount {
	return v1.VolumeMount{
		Name:      volumeName,
		SubPath:   configFile,
		ReadOnly:  true,
		MountPath: filePath,
	}
}
