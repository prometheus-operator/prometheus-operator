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

	corev1 "k8s.io/api/core/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
)

// tlsCredentials are the credentials used for web TLS.
type TLSCredentials struct {
	// mountPath is the directory where TLS credentials are intended to be mounted
	mountPath string
	// keySecret is the Kubernetes secret containing the TLS key.
	keySecret corev1.SecretKeySelector
	// cert is the Kubernetes secret or configmap containing the TLS certificate.
	cert monitoringv1.SecretOrConfigMap
	// clientCA is the Kubernetes secret or configmap containing the client CA certificate.
	clientCA monitoringv1.SecretOrConfigMap
}

// newTLSCredentials creates new TLSCredentials from secrets of configmaps.
// TODO: should we keep this, than to make all the fields public for use in mtls_config.go?
func NewTLSCredentials(
	mountPath string,
	key corev1.SecretKeySelector,
	cert monitoringv1.SecretOrConfigMap,
	clientCA monitoringv1.SecretOrConfigMap,
) *TLSCredentials {
	return &TLSCredentials{
		mountPath: mountPath,
		keySecret: key,
		cert:      cert,
		clientCA:  clientCA,
	}
}

// getMountParameters creates volumes and volume mounts referencing the TLS credentials.
func (a TLSCredentials) GetMountParameters(volumePrefix string) ([]corev1.Volume, []corev1.VolumeMount, error) {
	var (
		volumes []corev1.Volume
		mounts  []corev1.VolumeMount
		err     error
	)

	prefix := volumePrefix + "secret-key-"
	volumes, mounts, err = a.mountParamsForSecret(volumes, mounts, a.keySecret, prefix, a.GetKeyMountPath())
	if err != nil {
		return nil, nil, err
	}

	switch {
	case a.cert.Secret != nil:
		prefix := volumePrefix + "secret-cert-"
		volumes, mounts, err = a.mountParamsForSecret(volumes, mounts, *a.cert.Secret, prefix, a.GetCertMountPath())
		if err != nil {
			return nil, nil, err
		}
	case a.cert.ConfigMap != nil:
		prefix := volumePrefix + "configmap-cert-"
		volumes, mounts, err = a.mountParamsForConfigmap(volumes, mounts, *a.cert.ConfigMap, prefix, a.GetCertMountPath())
		if err != nil {
			return nil, nil, err
		}
	}

	switch {
	case a.clientCA.Secret != nil:
		prefix := volumePrefix + "secret-client-ca-"
		volumes, mounts, err = a.mountParamsForSecret(volumes, mounts, *a.clientCA.Secret, prefix, a.GetCAMountPath())
		if err != nil {
			return nil, nil, err
		}
	case a.clientCA.ConfigMap != nil:
		prefix := volumePrefix + "configmap-client-ca-"
		volumes, mounts, err = a.mountParamsForConfigmap(volumes, mounts, *a.clientCA.ConfigMap, prefix, a.GetCAMountPath())
		if err != nil {
			return nil, nil, err
		}
	}

	return volumes, mounts, nil
}

func (a TLSCredentials) mountParamsForSecret(
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

func (a TLSCredentials) mountParamsForConfigmap(
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

// getKeyMountPath is the mount path of the TLS key inside a prometheus container.
func (a TLSCredentials) GetKeyMountPath() string {
	secret := monitoringv1.SecretOrConfigMap{Secret: &a.keySecret}
	return a.tlsPathForSelector(secret, "key")
}

// getKeyFilename returns the filename (key) of the key.
func (a TLSCredentials) getKeyFilename() string {
	return a.keySecret.Key
}

// getCertMountPath is the mount path of the TLS certificate inside a prometheus container,.
func (a TLSCredentials) GetCertMountPath() string {
	if a.cert.ConfigMap != nil || a.cert.Secret != nil {
		return a.tlsPathForSelector(a.cert, "cert")
	}

	return ""
}

// getCertFilename returns the filename (key) of the certificate.
func (a TLSCredentials) getCertFilename() string {
	if a.cert.Secret != nil {
		return a.cert.Secret.Key
	} else if a.cert.ConfigMap != nil {
		return a.cert.ConfigMap.Key
	}

	return ""
}

// getCAMountPath is the mount path of the client CA certificate inside a prometheus container.
func (a TLSCredentials) GetCAMountPath() string {
	if a.clientCA.ConfigMap != nil || a.clientCA.Secret != nil {
		return a.tlsPathForSelector(a.clientCA, "ca")
	}

	return ""
}

// getCAFilename is the mount path of the client CA certificate inside a prometheus container.
func (a TLSCredentials) getCAFilename() string {
	if a.clientCA.Secret != nil {
		return a.clientCA.Secret.Key
	} else if a.clientCA.ConfigMap != nil {
		return a.clientCA.ConfigMap.Key
	}

	return ""
}

func (a *TLSCredentials) tlsPathForSelector(sel monitoringv1.SecretOrConfigMap, mountType string) string {
	var filename string
	if sel.Secret != nil {
		filename = fmt.Sprintf("secret/%s-%s", sel.Secret.Name, mountType)
	} else {
		filename = fmt.Sprintf("configmap/%s-%s", sel.ConfigMap.Name, mountType)
	}

	return path.Join(a.mountPath, filename)
}
