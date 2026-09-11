// Copyright The prometheus-operator Authors
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

package k8s

import (
	"context"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/utils/ptr"
)

// LoadSecretRef returns the data from a secret key reference.
// If the reference is set as optional and the secret or key isn't found, the
// function returns no error.
func LoadSecretRef(ctx context.Context, logger *slog.Logger, client typedcorev1.SecretInterface, sks *corev1.SecretKeySelector) ([]byte, error) {
	if sks == nil {
		return nil, nil
	}

	// Unless explicitly defined, references aren't optional.
	optional := ptr.Deref(sks.Optional, false)

	secret, err := client.Get(ctx, sks.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) && optional {
			logger.Debug(fmt.Sprintf("secret %v could not be found", sks.Name))
			return nil, nil
		}

		return nil, err
	}

	b, found := secret.Data[sks.Key]
	if !found {
		if optional {
			logger.Debug(fmt.Sprintf("secret %v could not be found", sks.Name))
			return nil, nil
		}

		return nil, fmt.Errorf("key %v could not be found in secret %v", sks.Key, sks.Name)
	}

	return b, nil
}

// LoadConfigMapRef returns the data from a configmap key reference.
// If the reference is set as optional and the configmap or key isn't found, the
// function returns no error.
func LoadConfigMapRef(ctx context.Context, logger *slog.Logger, client typedcorev1.ConfigMapInterface, cmks *corev1.ConfigMapKeySelector) ([]byte, error) {
	if cmks == nil {
		return nil, nil
	}

	// Unless explicitly defined, references aren't optional.
	optional := ptr.Deref(cmks.Optional, false)

	configmap, err := client.Get(ctx, cmks.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) && optional {
			logger.Debug(fmt.Sprintf("configmap %v could not be found", cmks.Name))
			return nil, nil
		}

		return nil, err
	}

	b, found := configmap.Data[cmks.Key]
	if !found {
		if optional {
			logger.Debug(fmt.Sprintf("configmap %v could not be found", cmks.Name))
			return nil, nil
		}

		return nil, fmt.Errorf("key %v could not be found in configmap %v", cmks.Key, cmks.Name)
	}

	return []byte(b), nil
}
