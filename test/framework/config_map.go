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
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func MakeConfigMapWithCert(ns, name, keyKey, certKey, caKey string,
	keyBytes, certBytes, caBytes []byte) *corev1.ConfigMap {

	cm := &corev1.ConfigMap{
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

func (f *Framework) WaitForConfigMapExist(ctx context.Context, ns, name string) (*corev1.ConfigMap, error) {
	var (
		configMap *corev1.ConfigMap
		getErr    error
	)
	err := wait.PollUntilContextTimeout(ctx, 2*time.Second, f.DefaultTimeout, true, func(ctx context.Context) (bool, error) {
		configMap, getErr = f.
			KubeClient.
			CoreV1().
			ConfigMaps(ns).
			Get(ctx, name, metav1.GetOptions{})

		if getErr != nil {
			return false, nil
		}

		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %w", err, getErr)
	}

	return configMap, nil
}

func (f *Framework) WaitForConfigMapNotExist(ctx context.Context, ns, name string) error {
	var getErr error
	err := wait.PollUntilContextTimeout(ctx, 2*time.Second, f.DefaultTimeout, true, func(ctx context.Context) (bool, error) {
		_, getErr = f.
			KubeClient.
			CoreV1().
			ConfigMaps(ns).
			Get(ctx, name, metav1.GetOptions{})

		if getErr != nil {
			if apierrors.IsNotFound(getErr) {
				return true, nil
			}

			return false, nil
		}

		getErr = errors.New("configmap found")
		return false, nil
	})

	if err != nil {
		return fmt.Errorf("%w: %w", err, getErr)
	}
	return nil
}
