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
	serverVolumePrefix = "cluster-tls-server-config-"
	clientVolumePrefix = "cluster-tls-client-config-"
	serverTLSCredDir   = "server_tls"
	clientTLSCredDir   = "client_tls"
)

// Config is the web configuration for prometheus and alertmanager instance.
//
// Config can make a secret which holds the web config contents, as well as
// volumes and volume mounts for referencing the secret and the
// necessary TLS credentials.
type ClusterTLSConfig struct {
	serverTLSConfig     *monitoringv1.WebTLSConfig
	clientTLSConfig     *monitoringv1.SafeTLSConfig
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
func New(mountingDir string, secretName string, clusterTLSConfig monitoringv1.ClusterTLSConfigFields) (*ClusterTLSConfig, error) {
	serverTLSConfig := clusterTLSConfig.ServerTLS
	if err := serverTLSConfig.Validate(); err != nil {
		return nil, err
	}

	clientTLSConfig := clusterTLSConfig.ClientTLS
	if err := clientTLSConfig.Validate(); err != nil {
		return nil, err
	}

	var serverTLSCreds *webconfig.TLSReferences
	var clientTLSCreds *webconfig.TLSReferences

	if serverTLSConfig != nil {
		serverTLSCreds = webconfig.NewTLSReferences(path.Join(mountingDir, serverTLSCredDir), serverTLSConfig.KeySecret, serverTLSConfig.Cert, serverTLSConfig.ClientCA)
	}
	if clientTLSConfig != nil {
		clientTLSCreds = webconfig.NewTLSReferences(path.Join(mountingDir, clientTLSCredDir), *clientTLSConfig.KeySecret, clientTLSConfig.Cert, clientTLSConfig.CA)
	}

	return &ClusterTLSConfig{
		serverTLSConfig:     serverTLSConfig,
		clientTLSConfig:     clientTLSConfig,
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
func (c ClusterTLSConfig) GetMountParameters() (monitoringv1.Argument, []v1.Volume, []v1.VolumeMount, error) {
	destinationPath := path.Join(c.mountingDir, configFile)

	var volumes []v1.Volume
	var mounts []v1.VolumeMount

	arg := c.makeArg(destinationPath)
	cfgVolume := c.makeVolume()
	volumes = append(volumes, cfgVolume)

	cfgMount := c.makeVolumeMount(destinationPath)
	mounts = append(mounts, cfgMount)

	if c.serverTLSReferences != nil {
		servertlsVolumes, servertlsMounts, err := c.serverTLSReferences.GetMountParameters(serverVolumePrefix)
		if err != nil {
			return monitoringv1.Argument{}, nil, nil, err
		}
		volumes = append(volumes, servertlsVolumes...)
		mounts = append(mounts, servertlsMounts...)
	}

	if c.clientTLSReferences != nil {
		clienttlsVolumes, clienttlsMounts, err := c.clientTLSReferences.GetMountParameters(clientVolumePrefix)
		if err != nil {
			return monitoringv1.Argument{}, nil, nil, err
		}
		volumes = append(volumes, clienttlsVolumes...)
		mounts = append(mounts, clienttlsMounts...)
	}

	return arg, volumes, mounts, nil
}

// CreateOrUpdateClusterTLSConfigSecret create or update a Kubernetes secret with the data for the cluster TLS config file.
// The format of the cluster TLS config file is available in the official prometheus documentation:
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

// generateClusterTLSConfigFileContents() generates the contents of cluster-tls-config.yaml
// from the ClusterTLSConfig in the form of an array of bytes.
func (c ClusterTLSConfig) generateClusterTLSConfigFileContents() ([]byte, error) {
	if c.serverTLSConfig == nil && c.clientTLSConfig == nil {
		return []byte{}, nil
	}

	cfg := yaml.MapSlice{}

	cfg = addServerTLSConfigToYaml(c, cfg)
	cfg = addClientTLSConfigToYaml(c, cfg)

	return yaml.Marshal(cfg)
}

// makeArg() returns an argument with the name "cluster.tls-config" with the filePath
// as its value.
func (c ClusterTLSConfig) makeArg(filePath string) monitoringv1.Argument {
	return monitoringv1.Argument{Name: cmdflag, Value: filePath}
}

// makeVolume() creates a Volume with volumeName = "cluster-tls-config" which stores
// the secret which contains the cluster TLS config.
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

// makeVolumeMount() creates a VolumeMount, mounting the cluster_tls_config.yaml SubPath
// to the given filePath.
func (c ClusterTLSConfig) makeVolumeMount(filePath string) v1.VolumeMount {
	return v1.VolumeMount{
		Name:      volumeName,
		SubPath:   configFile,
		ReadOnly:  true,
		MountPath: filePath,
	}
}
