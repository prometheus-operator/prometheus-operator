// Copyright 2016 The prometheus-operator Authors
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

package framework

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"github.com/pkg/errors"
)

func MakeConfigMapWithCert(kubeClient kubernetes.Interface, ns, name, keyKey, certKey, caKey string,
	keyBytes, certBytes, caBytes []byte) *v1.ConfigMap {

	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Data:       map[string]string{},
	}

	if keyBytes != nil {
		cm.Data[keyKey] = string(keyBytes)
	}

	if certBytes != nil {
		cm.Data[certKey] = string(certBytes)
	}

	if caBytes != nil {
		cm.Data[caKey] = string(caBytes)
	}

	return cm
}

func (f *Framework) WaitForConfigMapExist(ctx context.Context, ns, name string) (*v1.ConfigMap, error) {
	var configMap *v1.ConfigMap
	err := wait.PollUntilContextTimeout(ctx, 2*time.Second, f.DefaultTimeout, false, func(ctx context.Context) (bool, error) {
		var err error
		configMap, err = f.
			KubeClient.
			CoreV1().
			ConfigMaps(ns).
			Get(ctx, name, metav1.GetOptions{})

		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return true, nil
	})

	return configMap, errors.Wrapf(err, "waiting for ConfigMap '%v' in namespace '%v'", name, ns)
}

func (f *Framework) WaitForConfigMapNotExist(ctx context.Context, ns, name string) error {
	err := wait.PollUntilContextTimeout(ctx, 2*time.Second, f.DefaultTimeout, false, func(ctx context.Context) (bool, error) {
		var err error
		_, err = f.
			KubeClient.
			CoreV1().
			ConfigMaps(ns).
			Get(ctx, name, metav1.GetOptions{})

		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	})

	return errors.Wrapf(err, "waiting for ConfigMap '%v' in namespace '%v' to not exist", name, ns)
}
