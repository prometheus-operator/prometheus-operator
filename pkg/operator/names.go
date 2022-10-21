// Copyright 2022 The prometheus-operator Authors
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

package operator

import (
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// Nomenclator objects are used for naming Prometheus and PrometheusAgent objects
// and sub-objects consistently.
type Nomenclator struct {
	kind   string
	prefix string
	name   string
	label  string
	shards int32
}

func NewNomenclator(kind, prefix, name, label string, shards *int32) *Nomenclator {
	nc := Nomenclator{
		kind:   kind,
		prefix: prefix,
		name:   name,
	}

	nc.shards = 1
	if shards != nil && *shards > 1 {
		nc.shards = *shards
	}

	return &nc
}

func (n *Nomenclator) prefixedName() string {
	return fmt.Sprintf("%s-%s", n.prefix, n.name)
}

// Kind returns owner object's kind name
func (n *Nomenclator) Kind() string {
	return n.kind
}

// Kind returns owner object's kind name
func (n *Nomenclator) Prefix() string {
	return n.prefix
}

// BaseName returns owner object's name
func (n *Nomenclator) BaseName() string {
	return n.name
}

//
func (n *Nomenclator) NameLabelName() string {
	return n.label
}

// ConfigSecretName returns name of ConfigSecret owned by Nomenclator's owner
func (n *Nomenclator) ConfigSecretName() string {
	return n.prefixedName()
}

// TLSAssetsSecretName returns name of TLSAssets owned by Nomenclator's owner
func (n *Nomenclator) TLSAssetsSecretName() string {
	return fmt.Sprintf("%s-tls-assets", n.prefixedName())
}

// WebConfigSecretName returns name of Secret of WebConfig owned by Nomenclator's owner
func (n *Nomenclator) WebConfigSecretName() string {
	return fmt.Sprintf("%s-web-config", n.prefixedName())
}

// VokumeName returns name of Volume owned by Nomenclator's owner
func (n *Nomenclator) VolumeName() string {
	return fmt.Sprintf("%s-db", n.prefixedName())
}

func (n *Nomenclator) prometheusNameByShard(shard int32) string {
	base := n.prefixedName()
	if shard == 0 {
		return base
	}
	return fmt.Sprintf("%s-shard-%d", base, shard)
}

// ExpectedStatefulSetShardNames retuns slice of Statefulset's shard names
func (n *Nomenclator) ExpectedStatefulSetShardNames() []string {
	res := []string{}
	for i := int32(0); i < n.shards; i++ {
		res = append(res, n.prometheusNameByShard(i))
	}
	return res
}

func (n *Nomenclator) SubPathForStorage(s *monitoringv1.StorageSpec) string {
	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if s == nil || s.DisableMountSubPath {
		return ""
	}

	return n.VolumeName()
}
