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

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	corev1 "k8s.io/api/core/v1"
)

const (
	volumePrefix = "web-config-tls-"
)

// tlsCredentials are the credentials used for web TLS.
type tlsCredentials struct {
	// mountPath is the directory where TLS credentials are intended to be mounted
	mountPath string

	// keySecret is the kubernetes secret containing the TLS key
	keySecret corev1.SecretKeySelector
	// cert is the kubernetes secret or configmap containing the TLS certificate
	cert monitoringv1.SecretOrConfigMap
	// clientCA is the kubernetes secret or configmap containing the client CA certificate
	clientCA monitoringv1.SecretOrConfigMap
}

// newTLSCredentials creates new tlsCredentials from secrets of configmaps.
func newTLSCredentials(
	mountPath string,
	key corev1.SecretKeySelector,
	cert monitoringv1.SecretOrConfigMap,
	clientCA monitoringv1.SecretOrConfigMap,
) *tlsCredentials {
	return &tlsCredentials{
		mountPath: mountPath,
		keySecret: key,
		cert:      cert,
		clientCA:  clientCA,
	}
}

// getMountParameters creates volumes and volume mounts referencing the TLS credentials.
func (a tlsCredentials) getMountParameters() ([]corev1.Volume, []corev1.VolumeMount, error) {
	var (
		volumes []corev1.Volume
		mounts  []corev1.VolumeMount
		err     error
	)

	prefix := volumePrefix + "secret-key-"
	volumes, mounts, err = a.mountParamsForSecret(volumes, mounts, &a.keySecret, prefix, a.getKeyMountPath())
	if err != nil {
		return nil, nil, err
	}

	if a.cert.Secret != nil {
		prefix := volumePrefix + "secret-cert-"
		volumes, mounts, err = a.mountParamsForSecret(volumes, mounts, a.cert.Secret, prefix, a.getCertMountPath())
		if err != nil {
			return nil, nil, err
		}
	} else if a.cert.ConfigMap != nil {
		prefix := volumePrefix + "configmap-cert-"
		volumes, mounts, err = a.mountParamsForConfigmap(volumes, mounts, a.cert.ConfigMap, prefix, a.getCertMountPath())
		if err != nil {
			return nil, nil, err
		}
	}

	if a.clientCA.Secret != nil {
		prefix := volumePrefix + "secret-client-ca-"
		volumes, mounts, err = a.mountParamsForSecret(volumes, mounts, a.clientCA.Secret, prefix, a.getCAMountPath())
		if err != nil {
			return nil, nil, err
		}
	} else if a.clientCA.ConfigMap != nil {
		prefix := volumePrefix + "configmap-client-ca-"
		volumes, mounts, err = a.mountParamsForConfigmap(volumes, mounts, a.clientCA.ConfigMap, prefix, a.getCAMountPath())
		if err != nil {
			return nil, nil, err
		}
	}

	return volumes, mounts, nil
}

func (a tlsCredentials) mountParamsForSecret(
	volumes []corev1.Volume,
	mounts []corev1.VolumeMount,
	secret *corev1.SecretKeySelector,
	volumePrefix string,
	mountPath string,
) ([]corev1.Volume, []corev1.VolumeMount, error) {
	vn := k8sutil.NewResourceNamerWithPrefix(volumePrefix)
	volumeName, err := vn.UniqueVolumeName(secret.Name)
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

	mounts = append(mounts, corev1.VolumeMount{
		Name:      volumeName,
		ReadOnly:  true,
		MountPath: mountPath,
		SubPath:   secret.Key,
	})

	return volumes, mounts, nil
}

func (a tlsCredentials) mountParamsForConfigmap(
	volumes []corev1.Volume,
	mounts []corev1.VolumeMount,
	configMap *corev1.ConfigMapKeySelector,
	volumePrefix string,
	mountPath string,
) ([]corev1.Volume, []corev1.VolumeMount, error) {
	vn := k8sutil.NewResourceNamerWithPrefix(volumePrefix)
	volumeName, err := vn.UniqueVolumeName(configMap.Name)
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

	mounts = append(mounts, corev1.VolumeMount{
		Name:      volumeName,
		ReadOnly:  true,
		MountPath: mountPath,
		SubPath:   configMap.Key,
	})

	return volumes, mounts, nil
}

// getKeyMountPath is the mount path of the TLS key inside a prometheus container.
func (a tlsCredentials) getKeyMountPath() string {
	secret := monitoringv1.SecretOrConfigMap{Secret: &a.keySecret}
	return a.tlsPathForSelector(secret)
}

// getCertMountPath is the mount path of the TLS certificate inside a prometheus container,
func (a tlsCredentials) getCertMountPath() string {
	if a.cert.ConfigMap != nil || a.cert.Secret != nil {
		return a.tlsPathForSelector(a.cert)
	}

	return ""
}

// getCAMountPath is the mount path of the client CA certificate inside a prometheus container.
func (a tlsCredentials) getCAMountPath() string {
	if a.clientCA.ConfigMap != nil || a.clientCA.Secret != nil {
		return a.tlsPathForSelector(a.clientCA)
	}

	return ""
}

func (a *tlsCredentials) tlsPathForSelector(sel monitoringv1.SecretOrConfigMap) string {
	var filename string
	if sel.Secret != nil {
		filename = fmt.Sprintf("secret_%s_%s", sel.Secret.Name, sel.Secret.Key)
	} else {
		filename = fmt.Sprintf("configmap_%s_%s", sel.ConfigMap.Name, sel.ConfigMap.Key)
	}

	return path.Join(a.mountPath, filename)
}
