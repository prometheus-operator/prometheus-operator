package webconfig

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	mTLSVolumeName = "mtls-config"
	mTLSConfigFile = "mtls-config.yaml"
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
