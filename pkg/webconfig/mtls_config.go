package webconfig

import (
	"fmt"
	"path"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	mTLSVolumeName         = "mtls-config"
	mTLSConfigFile         = "mtls-config.yaml"
	volumePrefixMTLSConfig = "mtls-config-"
)

type MTLSConfig struct {
	tlsConfig            *monitoringv1.AMClusterTLSConfig
	tlsCredentialsClient *tlsCredentials
	tlsCredentialsServer *tlsCredentials
	mountingDir          string
	secretName           string
}

func NewMTLSConfig(mountingDir string, secretName string, clusterTLSConfig *monitoringv1.AMClusterTLSConfig) (*MTLSConfig, error) {
	if err := clusterTLSConfig.Validate(); err != nil {
		return nil, err
	}

	var tlsCredsClient *tlsCredentials
	if clusterTLSConfig != nil && clusterTLSConfig.TLSClientConfig != nil {
		tlsCredsClient = newTLSCredentials(mountingDir, clusterTLSConfig.TLSClientConfig.TLSKeySecret, clusterTLSConfig.TLSClientConfig.TLSCert, clusterTLSConfig.TLSClientConfig.ServerCA)
	}

	var tlsCredsServer *tlsCredentials
	if clusterTLSConfig != nil && clusterTLSConfig.TLSServerConfig != nil {
		tlsCredsServer = newTLSCredentials(mountingDir, clusterTLSConfig.TLSServerConfig.TLSKeySecret, clusterTLSConfig.TLSServerConfig.TLSCert, clusterTLSConfig.TLSServerConfig.ClientCA)
	}

	return &MTLSConfig{
		tlsConfig:            clusterTLSConfig,
		tlsCredentialsClient: tlsCredsClient,
		tlsCredentialsServer: tlsCredsServer,
		mountingDir:          mountingDir,
		secretName:           secretName,
	}, nil
}

func (c *MTLSConfig) MakeMTLSConfigFileSecret(labels map[string]string, ownerReference metav1.OwnerReference) (*v1.Secret, error) {
	data, err := c.generateConfigFileContents()
	if err != nil {
		return nil, err
	}

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            c.secretName,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Data: map[string][]byte{
			mTLSConfigFile: data,
		},
	}, nil
}

// Config file format is defined at https://github.com/prometheus/alertmanager/blob/main/docs/https.md#gossip-traffic
func (c MTLSConfig) generateConfigFileContents() ([]byte, error) {
	// mTLS requires both client and server certs to be present
	if c.tlsConfig == nil || c.tlsCredentialsClient == nil || c.tlsCredentialsServer == nil {
		return []byte{}, nil
	}

	mtlsConfigServer := yaml.MapSlice{}
	mtlsConfigClient := yaml.MapSlice{}

	// server part

	if certPath := c.tlsCredentialsServer.getCertMountPath(); certPath != "" {
		mtlsConfigServer = append(mtlsConfigServer, yaml.MapItem{Key: "cert_file", Value: certPath})
	}

	if keyPath := c.tlsCredentialsServer.getKeyMountPath(); keyPath != "" {
		mtlsConfigServer = append(mtlsConfigServer, yaml.MapItem{Key: "key_file", Value: keyPath})
	}

	if c.tlsConfig.TLSServerConfig.ClientAuthType != "" {
		mtlsConfigServer = append(mtlsConfigServer, yaml.MapItem{
			Key:   "client_auth_type",
			Value: c.tlsConfig.TLSServerConfig.ClientAuthType,
		})
	}

	if caPath := c.tlsCredentialsServer.getCAMountPath(); caPath != "" {
		mtlsConfigServer = append(mtlsConfigServer, yaml.MapItem{Key: "client_ca_file", Value: caPath})
	}

	if c.tlsConfig.TLSServerConfig.MinVersion != "" {
		mtlsConfigServer = append(mtlsConfigServer, yaml.MapItem{
			Key:   "min_version",
			Value: c.tlsConfig.TLSServerConfig.MinVersion,
		})
	}

	if c.tlsConfig.TLSServerConfig.MaxVersion != "" {
		mtlsConfigServer = append(mtlsConfigServer, yaml.MapItem{
			Key:   "max_version",
			Value: c.tlsConfig.TLSServerConfig.MaxVersion,
		})
	}

	if len(c.tlsConfig.TLSServerConfig.CipherSuites) != 0 {
		mtlsConfigServer = append(mtlsConfigServer, yaml.MapItem{
			Key:   "cipher_suites",
			Value: c.tlsConfig.TLSServerConfig.CipherSuites,
		})
	}

	if c.tlsConfig.TLSServerConfig.PreferServerCipherSuites != nil {
		mtlsConfigServer = append(mtlsConfigServer, yaml.MapItem{
			Key:   "prefer_server_cipher_suites",
			Value: c.tlsConfig.TLSServerConfig.PreferServerCipherSuites,
		})
	}

	if len(c.tlsConfig.TLSServerConfig.CurvePreferences) != 0 {
		mtlsConfigServer = append(mtlsConfigServer, yaml.MapItem{
			Key:   "curve_preferences",
			Value: c.tlsConfig.TLSServerConfig.CurvePreferences,
		})
	}

	// client part

	if certPath := c.tlsCredentialsClient.getCertMountPath(); certPath != "" {
		mtlsConfigClient = append(mtlsConfigClient, yaml.MapItem{Key: "cert_file", Value: certPath})
	}

	if keyPath := c.tlsCredentialsClient.getKeyMountPath(); keyPath != "" {
		mtlsConfigClient = append(mtlsConfigClient, yaml.MapItem{Key: "key_file", Value: keyPath})
	}

	if caPath := c.tlsCredentialsClient.getCAMountPath(); caPath != "" {
		mtlsConfigClient = append(mtlsConfigClient, yaml.MapItem{Key: "ca_file", Value: caPath})
	}

	if serverName := c.tlsConfig.TLSClientConfig.ServerName; serverName != "" {
		mtlsConfigClient = append(mtlsConfigClient, yaml.MapItem{Key: "server_name", Value: serverName})
	}

	if c.tlsConfig.TLSClientConfig.InsecureSkipVerify != nil {
		mtlsConfigClient = append(mtlsConfigClient, yaml.MapItem{
			Key:   "insecure_skip_verify",
			Value: c.tlsConfig.TLSClientConfig.InsecureSkipVerify,
		})
	}

	cfg := yaml.MapSlice{
		{
			Key:   "tls_server_config",
			Value: mtlsConfigServer,
		},
		{
			Key:   "tls_client_config",
			Value: mtlsConfigClient,
		},
	}

	return yaml.Marshal(cfg)
}

func (c MTLSConfig) makeArg(filePath string) string {
	return fmt.Sprintf("--cluster.tls-config=%s", filePath)
}

func (c MTLSConfig) makeVolume() v1.Volume {
	return v1.Volume{
		Name: mTLSVolumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: c.secretName,
			},
		},
	}
}

func (c MTLSConfig) makeVolumeMount(filePath string) v1.VolumeMount {
	return v1.VolumeMount{
		Name:      mTLSVolumeName,
		SubPath:   mTLSConfigFile,
		ReadOnly:  true,
		MountPath: filePath,
	}
}

// GetMountParameters returns volumes and volume mounts referencing the config file
// and the associated TLS credentials.
// In addition, GetMountParameters returns a "--cluster.tls-config" command line option pointing
// to the file in the volume mount.
func (c MTLSConfig) GetMountParameters() (string, []v1.Volume, []v1.VolumeMount) {
	destinationPath := path.Join(c.mountingDir, mTLSConfigFile)

	var volumes []v1.Volume
	var mounts []v1.VolumeMount

	arg := c.makeArg(destinationPath)
	cfgVolume := c.makeVolume()
	volumes = append(volumes, cfgVolume)

	cfgMount := c.makeVolumeMount(destinationPath)
	mounts = append(mounts, cfgMount)

	if c.tlsConfig != nil {
		tlsVolumes, tlsMounts := c.tlsCredentialsClient.getMountParameters(volumePrefixMTLSConfig)
		volumes = append(volumes, tlsVolumes...)
		mounts = append(mounts, tlsMounts...)
		tlsVolumes, tlsMounts = c.tlsCredentialsServer.getMountParameters(volumePrefixMTLSConfig)
		// deduplicate volume by names
		for idx, volumeToAppend := range tlsVolumes {
			duplicated := false
			for _, volumeAppended := range volumes {
				if volumeToAppend.Name == volumeAppended.Name {
					duplicated = true
					break
				}
			}
			if duplicated {
				continue
			}
			volumes = append(volumes, volumeToAppend)
			mounts = append(mounts, tlsMounts[idx])
		}
	}

	return arg, volumes, mounts
}
