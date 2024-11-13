# Status subresource for Prometheus operator CRDs

* **Owners:**
  * [arnecls](https://github.com/arnecls)

* **Related Tickets:**
  * [#6437](https://github.com/prometheus-operator/prometheus-operator/issues/6437)

* **Other docs:** 
  * [Well known kubernetes labels](https://kubernetes.io/docs/reference/labels-annotations-taints/#topologykubernetesiozone)
  * [AWS zone names](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-availability-zones)
  * [GCP zone names](https://cloud.google.com/compute/docs/regions-zones#available)

This proposal describes how we can implement zone-aware sharding by adding
support for custom labels and zone configuration options to the existing 
prometheus configuration resources. 

## Why

When running large, multi-zone clusters, prometheus scraping can lead to an
increase in inter-zone traffic costs. The current sharding mechanics will
allow multiple instances of prometheus to run, but as the `__address__` label
is hard coded, all instances will always fetch all zones.

It is not sufficient to simply switch this label with another label, as 
multiple prometheus instances per zone might be required. Furthermore we
must be able to calculate the "assignment" of a specific instance to a zone,
and that assignment must be stable.

## Goals

- Define a set of configuration options required to allow zone aware sharding
- Define the relabel configuration to be generated for zone aware sharding
- Define changes to the prometheus pod spec to support zone stickyness
- Stay backwards compatible to the current mechanism by default

## Non-goals

- Implement mechanisms to avoid configuration errors by the user
- Support mixed environments (kubernetes and non-kubernetes targets are scraped)

## Algorithm

In order to do calculate a stable assignment, following parameters are required:

1. `num_shards`: The number of prometheus shards
1. `shard_index`: A number of the range `[0..num_shards[` identifying a single prometheus instance
1. `zones`: A list of the zones to be scraped
2. `zone_label`: A label denoting the zone of a target
3. `address`: The content of the `__address__` label

It has to be noted that `zone_label` is supposed to contain a value from the
`zones` list.
The `num_shards` value is referring to the currently available `.spec.shards`
from the [Prometheus custom resource definition](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#monitoring.coreos.com/v1.Prometheus).

Given these values, a target is to be scraped by a given prometheus instance 
by using the following algorithm:

```go
assert(num_shards >= len(zones))     # Error: zone(s) will not get scraped
assert(num_shards % len(zones) == 0) # Warning: multi-scraping of instances

shards_per_zone := max(1, floor(num_shards / len(zones)))
prom_zone_idx := shard_index % shards_per_zone
    
if zone_label == zones[prom_zone_idx] {
    
    assignment_idx := floor(shard_index / shards_per_zone)

    if hash(address) % shards_per_zone == assignment_idx {
        do_scrape()
    }
}
```

By using modulo to calculate the `prom_zone_idx`, instances will be distributed
to zones in the sense of `A,B,C,A,B,C` and so on. This allows a modification of
the `num_shards` value without redistribution of data.  
This was preferred over allowing the number of `zones` to change, as this is
less likely to happen.

### Edge cases

We have introduced asserts in the above section to warn about edge cases that
might lead to duplicate data or data loss.

By the above algorithm. prometheus instances will be distributed in an
alternating fashion by using the already existing shard index.
This leads to the following edge cases:

When `num_shards` is 10 and `len(zones)` is 3 as in `[A..C]`.
`shards_per_zone` is 3. This yields the following distribution

```
0 1 2 | 3 4 5 | 6 7 8 | 9 | instance
A B C | A B C | A B C | A | zone
0 0 0 | 1 1 1 | 2 2 2 | 0 | assignment index
```

In this case the 2nd assert will warn about double scraping of instances in
zone A, as the same targets are being assigned to instance 0 and 9.

When `num_shards` is 2 and `len(zones)` is 3 as in `[A..C]`.
`shards_per_zone` is 1. This yields the following distribution

```
0 1   | instance
A B C | zone
0 0 0 | assignment index
```

In this case the 1st assert will warn about zone C not being scraped.

## Required configuration options

Following the algorithm presented above, we suggest the following configuration
options to be added to the [Prometheus custom resource definition](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#monitoring.coreos.com/v1.Prometheus):

```yaml
spec:
  shardingStrategy:
    mode: 'classic'             # can be 'classic' or 'topology'
    label: '__address__'        # defaults to '__address__` for 'classic' and 
                                # 'topology.kubernetes.io/zone' for 'topology'
    topology: []                # expects list of zones for 'toplogy'
    additionalRelabelConfig: [] # Optional: array of relabel configurations
```

The `additionalRelabelConfig` section is meant to allow the `label` to be
generated if needed. This should allow enough flexibility to cover edgecases
not anticiapted by this proposal.

## Generated configuration

The following configuration uses Helm templating syntax and the proposed
configuration structure as proposed above. Functions from the [sprig](https://masterminds.github.io/sprig/)
library are used, as they are also present in Helm.

Values generated by the operator, like `shard_index` are retrieved from an
`.operator` object.

### classic mode

```yaml
  {{ $label := .spec.shardingStrategy.label | replace "/" "_" | replace "." "_" }}
  {{ .spec.shardingStrategy.additionalRelabelConfig | toYaml }}
  - source_labels: [ {{ default "__address__" $label }} ] 
    separator: ;
    modulus: {{ .spec.shards }}
    target_label: __tmp_hash
    replacement: $1
    action: hashmod
  - source_labels: [__tmp_hash]
    separator: ;
    regex: {{ .operator.shardIndex }}
    replacement: $1
    action: keep
```

### topology mode

```yaml
  {{ $labelSuffix := .spec.shardingStrategy.label | replace "/" "_" | replace "." "_" }}
  {{ $shardsPerZone := max 1 (.spec.shards | div (len .spec.shardingStrategy.Topology)) }}
  {{ $promZoneIdx := .operator.shardIndex | mod $shardsPerZone }}
  
  {{ .spec.shardingStrategy.additionalRelabelConfig | toYaml }}
  - source_labels: [ '__meta_kubernetes_pod_label_{{ default "topology_kubernetes_io_zone" $labelSuffix }}' ]
    separator: ;
    regex: {{ index .spec.shardingStrategy.topology $promZoneIdx }}
    replacement: $1
    action: keep
  - source_labels: [ '__address__' ] 
    separator: ;
    modulus: {{ $shardsPerZone }}
    target_label: __tmp_hash
    replacement: $1
    action: hashmod
  - source_labels: [__tmp_hash]
    separator: ;
    regex: {{ .operator.shardIndex | div $shardsPerZone }}
    replacement: $1
    action: keep
```

## Prometheus instance zone assignment

To make sure that prometheus instances are deployed to the correct zone (of their 
assigned target), we need to generate [node affinities](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity).

The corresponding section in a kubernetes podspec will look like as follow.  
We use the same Helm templating notation as above.

```yaml
{{ $shardsPerZone := max 1 (.spec.shards | div (len .spec.shardingStrategy.Topology)) }}
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: topology.kubernetes.io/zone
            operator: In
            values:
            - '{{ index .spec.shardingStrategy.Topology (.operator.shardIndex | div $shardsPerZone) }}'
```

## Alternatives

N/A

## Action Plan

N/A
