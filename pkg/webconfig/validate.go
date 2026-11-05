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

package webconfig

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
)

// ValidateTLSAssets verifies that web TLS secret and configmap references exist
// and contain the configured keys before mounting them into the workload.
func ValidateTLSAssets(ctx context.Context, namespace string, store *assets.StoreBuilder, tls *monitoringv1.WebTLSConfig) error {
	if tls == nil {
		return nil
	}

	if err := tls.Validate(); err != nil {
		return err
	}

	if tls.KeySecret != (corev1.SecretKeySelector{}) {
		if _, err := store.GetSecretKey(ctx, namespace, tls.KeySecret); err != nil {
			return fmt.Errorf("invalid TLS key secret: %w", err)
		}
	}

	if tls.Cert != (monitoringv1.SecretOrConfigMap{}) {
		if _, err := store.GetKey(ctx, namespace, tls.Cert); err != nil {
			return fmt.Errorf("invalid TLS certificate reference: %w", err)
		}
	}

	if tls.ClientCA != (monitoringv1.SecretOrConfigMap{}) {
		if _, err := store.GetKey(ctx, namespace, tls.ClientCA); err != nil {
			return fmt.Errorf("invalid TLS client CA reference: %w", err)
		}
	}

	return nil
}
