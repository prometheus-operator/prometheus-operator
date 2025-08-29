// Copyright 2024 The prometheus-operator Authors
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

package assets

import (
	v1 "k8s.io/api/core/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// StoreGetter can get data from ConfigMap/Secret objects via key selectors.
// It will return an error if the key selector didn't match.
type StoreGetter interface {
	GetSecretOrConfigMapKey(key monitoringv1.SecretOrConfigMap) (string, error)
	GetConfigMapKey(key v1.ConfigMapKeySelector) (string, error)
	GetSecretKey(key v1.SecretKeySelector) ([]byte, error)
	TLSAsset(key any) string
}
