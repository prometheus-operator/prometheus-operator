package operator

// This module contains class for naming Prometheus and PrometheusAgent objects
// and sub-objects consistently.

import (
	"fmt"
)

var (
	minShards = int32(1)
)

type Nomenclator struct {
	kind   string
	prefix string
	name   string
	shards int32
}

func NewNomenclator(kind, prefix, name string, shards *int32) *Nomenclator {
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

func (n *Nomenclator) Kind() string {
	return n.kind
}

func (n *Nomenclator) BaseName() string {
	return n.name
}

func (n *Nomenclator) ConfigSecretName() string {
	return n.prefixedName()
}

func (n *Nomenclator) TLSAssetsSecretName() string {
	return fmt.Sprintf("%s-tls-assets", n.prefixedName())
}

func (n *Nomenclator) WebConfigSecretName() string {
	return fmt.Sprintf("%s-web-config", n.prefixedName())
}

func (n *Nomenclator) VolumeName() string {
	return fmt.Sprintf("%s-db", n.prefixedName())
}

func (n *Nomenclator) PrometheusNameByShard(shard int32) string {
	base := n.prefixedName()
	if shard == 0 {
		return base
	}
	return fmt.Sprintf("%s-shard-%d", base, shard)
}

func (n *Nomenclator) ExpectedStatefulSetShardNames() []string {
	res := []string{}
	for i := int32(0); i < n.shards; i++ {
		res = append(res, n.PrometheusNameByShard(i))
	}
	return res
}
