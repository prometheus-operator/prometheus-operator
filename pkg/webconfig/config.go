// Copyright 2021 The prometheus-operator Authors
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

package webconfig

import (
	"context"
	"fmt"
	"path"
	"strings"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	volumeName = "web-config"
	configFile = "web-config.yaml"
)

// Config is the web configuration for prometheus and alertmanager instance.
//
// Config can make a secret which holds the web config contents, as well as
// volumes and volume mounts for referencing the secret and the
// necessary TLS credentials.
type Config struct {
	tlsConfig      *monitoringv1.WebTLSConfig
	httpConfig     *monitoringv1.WebHTTPConfig
	tlsCredentials *tlsCredentials
	mountingDir    string
	secretName     string
}

// New creates a new Config.
func New(mountingDir string, secretName string, configFileFields monitoringv1.WebConfigFileFields) (*Config, error) {
	tlsConfig := configFileFields.TLSConfig

	if err := tlsConfig.Validate(); err != nil {
		return nil, err
	}

	var tlsCreds *tlsCredentials
	if tlsConfig != nil {
		tlsCreds = newTLSCredentials(mountingDir, tlsConfig.KeySecret, tlsConfig.Cert, tlsConfig.ClientCA)
	}

	return &Config{
		tlsConfig:      tlsConfig,
		httpConfig:     configFileFields.HTTPConfig,
		tlsCredentials: tlsCreds,
		mountingDir:    mountingDir,
		secretName:     secretName,
	}, nil
}

// GetMountParameters returns volumes and volume mounts referencing the config file
// and the associated TLS credentials.
// In addition, GetMountParameters returns a web.config.file command line option pointing
// to the file in the volume mount.
func (c Config) GetMountParameters() (monitoringv1.Argument, []v1.Volume, []v1.VolumeMount, error) {
	destinationPath := path.Join(c.mountingDir, configFile)

	var volumes []v1.Volume
	var mounts []v1.VolumeMount

	arg := c.makeArg(destinationPath)
	cfgVolume := c.makeVolume()
	volumes = append(volumes, cfgVolume)

	cfgMount := c.makeVolumeMount(destinationPath)
	mounts = append(mounts, cfgMount)

	if c.tlsCredentials != nil {
		tlsVolumes, tlsMounts, err := c.tlsCredentials.getMountParameters()
		if err != nil {
			return monitoringv1.Argument{}, nil, nil, err
		}
		volumes = append(volumes, tlsVolumes...)
		mounts = append(mounts, tlsMounts...)
	}

	return arg, volumes, mounts, nil
}

// CreateOrUpdateWebConfigSecret create or update a Kubernetes secret with the data for the web config file.
// The format of the web config file is available in the official prometheus documentation:
// https://prometheus.io/docs/prometheus/latest/configuration/https/#https-and-authentication
func (c Config) CreateOrUpdateWebConfigSecret(ctx context.Context, secretClient clientv1.SecretInterface, labels map[string]string, ownerReference metav1.OwnerReference) error {
	data, err := c.generateConfigFileContents()
	if err != nil {
		return err
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            c.secretName,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Data: map[string][]byte{
			configFile: data,
		},
	}

	return k8sutil.CreateOrUpdateSecret(ctx, secretClient, secret)
}

func (c Config) generateConfigFileContents() ([]byte, error) {
	if c.tlsConfig == nil && c.httpConfig == nil {
		return []byte{}, nil
	}

	cfg := yaml.MapSlice{}

	cfg = c.addTLSServerConfigToYaml(cfg)
	cfg = c.addHTTPServerConfigToYaml(cfg)

	return yaml.Marshal(cfg)
}

func (c Config) addTLSServerConfigToYaml(cfg yaml.MapSlice) yaml.MapSlice {
	tls := c.tlsConfig
	if tls == nil {
		return cfg
	}

	tlsServerConfig := yaml.MapSlice{}
	if certPath := c.tlsCredentials.getCertMountPath(); certPath != "" {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{Key: "cert_file", Value: fmt.Sprintf("%s/%s", certPath, c.tlsCredentials.getCertFilename())})
	}

	if keyPath := c.tlsCredentials.getKeyMountPath(); keyPath != "" {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{Key: "key_file", Value: fmt.Sprintf("%s/%s", keyPath, c.tlsCredentials.getKeyFilename())})
	}

	if tls.ClientAuthType != "" {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{
			Key:   "client_auth_type",
			Value: tls.ClientAuthType,
		})
	}

	if caPath := c.tlsCredentials.getCAMountPath(); caPath != "" {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{Key: "client_ca_file", Value: fmt.Sprintf("%s/%s", caPath, c.tlsCredentials.getCAFilename())})
	}

	if tls.MinVersion != "" {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{
			Key:   "min_version",
			Value: tls.MinVersion,
		})
	}

	if tls.MaxVersion != "" {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{
			Key:   "max_version",
			Value: tls.MaxVersion,
		})
	}

	if len(tls.CipherSuites) != 0 {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{
			Key:   "cipher_suites",
			Value: tls.CipherSuites,
		})
	}

	if tls.PreferServerCipherSuites != nil {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{
			Key:   "prefer_server_cipher_suites",
			Value: tls.PreferServerCipherSuites,
		})
	}

	if len(tls.CurvePreferences) != 0 {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{
			Key:   "curve_preferences",
			Value: tls.CurvePreferences,
		})
	}

	return append(cfg, yaml.MapItem{Key: "tls_server_config", Value: tlsServerConfig})
}

func (c Config) addHTTPServerConfigToYaml(cfg yaml.MapSlice) yaml.MapSlice {
	http := c.httpConfig
	if http == nil {
		return cfg
	}

	httpServerConfig := yaml.MapSlice{}

	if http.HTTP2 != nil {
		httpServerConfig = append(httpServerConfig, yaml.MapItem{Key: "http2", Value: *http.HTTP2})
	}

	headers := http.Headers
	if headers == nil {
		return append(cfg, yaml.MapItem{Key: "http_server_config", Value: httpServerConfig})
	}

	headersConfig := yaml.MapSlice{}

	if headers.ContentSecurityPolicy != "" {
		headersConfig = append(headersConfig, yaml.MapItem{
			Key:   "Content-Security-Policy",
			Value: headers.ContentSecurityPolicy,
		})
	}

	if headers.StrictTransportSecurity != "" {
		headersConfig = append(headersConfig, yaml.MapItem{
			Key: "Strict-Transport-Security", Value: headers.StrictTransportSecurity,
		})
	}

	if headers.XContentTypeOptions != "" {
		headersConfig = append(headersConfig, yaml.MapItem{
			Key: "X-Content-Type-Options", Value: strings.ToLower(headers.XContentTypeOptions),
		})
	}

	if headers.XFrameOptions != "" {
		headersConfig = append(headersConfig, yaml.MapItem{
			Key: "X-Frame-Options", Value: strings.ToLower(headers.XFrameOptions),
		})
	}

	if headers.XXSSProtection != "" {
		headersConfig = append(headersConfig, yaml.MapItem{
			Key: "X-XSS-Protection", Value: headers.XXSSProtection,
		})
	}

	httpServerConfig = append(httpServerConfig, yaml.MapItem{Key: "headers", Value: headersConfig})

	return append(cfg, yaml.MapItem{Key: "http_server_config", Value: httpServerConfig})
}

func (c Config) makeArg(filePath string) monitoringv1.Argument {
	return monitoringv1.Argument{Name: "web.config.file", Value: filePath}
}

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

func (c Config) makeVolumeMount(filePath string) v1.VolumeMount {
	return v1.VolumeMount{
		Name:      volumeName,
		SubPath:   configFile,
		ReadOnly:  true,
		MountPath: filePath,
	}
}
