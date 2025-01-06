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
	"fmt"
	"path"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
)

const (
	volumePrefix = "web-config-tls-"
)

// tlsReferences represent TLS material referenced from secrets/configmaps.
type tlsReferences struct {
	// mountPath is the directory where the TLS files are intended to be mounted.
	mountPath string

	// keySecret is the Kubernetes Secret containing the TLS private key.
	keySecret corev1.SecretKeySelector
	// cert is the Kubernetes Secret or ConfigMap containing the TLS certificate.
	cert monitoringv1.SecretOrConfigMap
	// clientCA is the Kubernetes Secret or ConfigMap containing the client CA certificate.
	clientCA monitoringv1.SecretOrConfigMap
}

func newTLSReferences(mountPath string, cfg monitoringv1.WebTLSConfig) *tlsReferences {
	return &tlsReferences{
		mountPath: mountPath,
		keySecret: cfg.KeySecret,
		cert:      cfg.Cert,
		clientCA:  cfg.ClientCA,
	}
}

// getMountParameters creates volumes and volume mounts referencing the TLS credentials.
func (tr *tlsReferences) getMountParameters() ([]corev1.Volume, []corev1.VolumeMount, error) {
	var (
		volumes []corev1.Volume
		mounts  []corev1.VolumeMount
		err     error
	)

	prefix := volumePrefix + "secret-key-"
	volumes, mounts, err = tr.mountParamsForSecret(volumes, mounts, tr.keySecret, prefix, tr.getKeyMountPath())
	if err != nil {
		return nil, nil, err
	}

	switch {
	case tr.cert.Secret != nil:
		prefix := volumePrefix + "secret-cert-"
		volumes, mounts, err = tr.mountParamsForSecret(volumes, mounts, *tr.cert.Secret, prefix, tr.getCertMountPath())
		if err != nil {
			return nil, nil, err
		}
	case tr.cert.ConfigMap != nil:
		prefix := volumePrefix + "configmap-cert-"
		volumes, mounts, err = tr.mountParamsForConfigmap(volumes, mounts, *tr.cert.ConfigMap, prefix, tr.getCertMountPath())
		if err != nil {
			return nil, nil, err
		}
	}

	switch {
	case tr.clientCA.Secret != nil:
		prefix := volumePrefix + "secret-client-ca-"
		volumes, mounts, err = tr.mountParamsForSecret(volumes, mounts, *tr.clientCA.Secret, prefix, tr.getCAMountPath())
		if err != nil {
			return nil, nil, err
		}
	case tr.clientCA.ConfigMap != nil:
		prefix := volumePrefix + "configmap-client-ca-"
		volumes, mounts, err = tr.mountParamsForConfigmap(volumes, mounts, *tr.clientCA.ConfigMap, prefix, tr.getCAMountPath())
		if err != nil {
			return nil, nil, err
		}
	}

	return volumes, mounts, nil
}

func (tr *tlsReferences) mountParamsForSecret(
	volumes []corev1.Volume,
	mounts []corev1.VolumeMount,
	secret corev1.SecretKeySelector,
	volumePrefix string,
	mountPath string,
) ([]corev1.Volume, []corev1.VolumeMount, error) {
	vn := k8sutil.NewResourceNamerWithPrefix(volumePrefix)
	volumeName, err := vn.UniqueDNS1123Label(secret.Name)
	if err != nil {
		return nil, nil, err
	}

	volumes = append(volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secret.Name,
			},
		},
	})

	// We're mounting the volume in full as subPath mounts can't receive updates. This is important when renewing
	// certificates. Prometheus and Alertmanager then load certificates on every request, there is no need to tell them
	// to reload their configuration.
	//
	// References:
	// * https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod
	// * https://github.com/prometheus-operator/prometheus-operator/issues/5527
	// * https://github.com/prometheus-operator/prometheus-operator/pull/5535#discussion_r1194940482
	mounts = append(mounts, corev1.VolumeMount{
		Name:      volumeName,
		ReadOnly:  true,
		MountPath: mountPath,
	})

	return volumes, mounts, nil
}

func (tr *tlsReferences) mountParamsForConfigmap(
	volumes []corev1.Volume,
	mounts []corev1.VolumeMount,
	configMap corev1.ConfigMapKeySelector,
	volumePrefix string,
	mountPath string,
) ([]corev1.Volume, []corev1.VolumeMount, error) {
	vn := k8sutil.NewResourceNamerWithPrefix(volumePrefix)
	volumeName, err := vn.UniqueDNS1123Label(configMap.Name)
	if err != nil {
		return nil, nil, err
	}

	volumes = append(volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMap.Name,
				},
			},
		},
	})

	// We're mounting the volume in full as subPath mounts can't receive updates. This is important when renewing
	// certificates. Prometheus and Alertmanager then load certificates on every request, there is no need to tell them
	// to reload their configuration.
	//
	// References:
	// * https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod
	// * https://github.com/prometheus-operator/prometheus-operator/issues/5527
	// * https://github.com/prometheus-operator/prometheus-operator/pull/5535#discussion_r1194940482
	mounts = append(mounts, corev1.VolumeMount{
		Name:      volumeName,
		ReadOnly:  true,
		MountPath: mountPath,
	})

	return volumes, mounts, nil
}

// getKeyMountPath is the mount path of the private key inside a container.
func (tr *tlsReferences) getKeyMountPath() string {
	secret := monitoringv1.SecretOrConfigMap{Secret: &tr.keySecret}
	return tr.tlsPathForSelector(secret, "key")
}

// getKeyFilename returns the filename (key) of the private key.
func (tr *tlsReferences) getKeyFilename() string {
	return tr.keySecret.Key
}

// getCertMountPath is the mount path of the TLS certificate inside a container.
func (tr *tlsReferences) getCertMountPath() string {
	if tr.cert.ConfigMap != nil || tr.cert.Secret != nil {
		return tr.tlsPathForSelector(tr.cert, "cert")
	}

	return ""
}

// getCertFilename returns the filename (key) of the certificate.
func (tr *tlsReferences) getCertFilename() string {
	if tr.cert.Secret != nil {
		return tr.cert.Secret.Key
	} else if tr.cert.ConfigMap != nil {
		return tr.cert.ConfigMap.Key
	}

	return ""
}

// getCAMountPath is the mount path of the client CA certificate inside a container.
func (tr *tlsReferences) getCAMountPath() string {
	if tr.clientCA.ConfigMap != nil || tr.clientCA.Secret != nil {
		return tr.tlsPathForSelector(tr.clientCA, "ca")
	}

	return ""
}

// getCAFilename retruns the filename (key) of the client CA certificate.
func (tr *tlsReferences) getCAFilename() string {
	if tr.clientCA.Secret != nil {
		return tr.clientCA.Secret.Key
	} else if tr.clientCA.ConfigMap != nil {
		return tr.clientCA.ConfigMap.Key
	}

	return ""
}

func (tr *tlsReferences) tlsPathForSelector(sel monitoringv1.SecretOrConfigMap, mountType string) string {
	var filename string
	if sel.Secret != nil {
		filename = filepath.Join("secret", fmt.Sprintf("%s-%s", sel.Secret.Name, mountType))
	} else {
		filename = filepath.Join("configmap", fmt.Sprintf("%s-%s", sel.ConfigMap.Name, mountType))
	}

	return path.Join(tr.mountPath, filename)
}
