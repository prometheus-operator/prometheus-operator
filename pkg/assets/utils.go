// Copyright 2020 The prometheus-operator Authors
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
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
)

// TLSAssetKey is a key for a TLS asset.
type TLSAssetKey struct {
	from string
	ns   string
	name string
	key  string
}

// TLSAssetKeyFromSecretSelector returns a TLSAssetKey struct from a secret key selector.
func TLSAssetKeyFromSecretSelector(ns string, sel *v1.SecretKeySelector) TLSAssetKey {
	return TLSAssetKeyFromSelector(
		ns,
		monitoringv1.SecretOrConfigMap{
			Secret: sel,
		},
	)
}

// TLSAssetKeyFromSelector returns a TLSAssetKey struct from a secret or configmap key selector.
func TLSAssetKeyFromSelector(ns string, sel monitoringv1.SecretOrConfigMap) TLSAssetKey {
	if sel.Secret != nil {
		return TLSAssetKey{
			from: "secret",
			ns:   ns,
			name: sel.Secret.Name,
			key:  sel.Secret.Key,
		}
	}
	return TLSAssetKey{
		from: "configmap",
		ns:   ns,
		name: sel.ConfigMap.Name,
		key:  sel.ConfigMap.Key,
	}
}

// String implements the fmt.Stringer interface.
func (k TLSAssetKey) String() string {
	return fmt.Sprintf("%s_%s_%s_%s", k.from, k.ns, k.name, k.key)
}
