# Zone aware sharding for prometheus

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

- Implement mechanisms to fix configuration errors by the user
- Support mixed environments (kubernetes and non-kubernetes targets are scraped)

## Algorithm

In order to do calculate a stable assignment, following parameters are required:

1. `num_shards`: The number of prometheus shards
2. `shard_index`: A number of the range `[0..num_shards[` identifying a single prometheus 
   instance
3. `zones`: A list of the zones to be scraped
4. `zone_label`: A label denoting the zone of a target
5. `address`: The content of the `__address__` label

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
the `num_shards` value without redistribution of shards or data.
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
0 1 2 | 3 4 5 | 6 7 8 | 9 | shard index
A B C | A B C | A B C | A | zone
0 0 0 | 1 1 1 | 2 2 2 | 0 | assignment index
```

In this case the 2nd assert will warn about double scraping of instances in
zone A, as the same targets are being assigned to instance 0 and 9.

When `num_shards` is 2 and `len(zones)` is 3 as in `[A..C]`.
`shards_per_zone` is 1. This yields the following distribution

```
0 1   | shard index
A B C | zone
0 0 0 | assignment index
```

In this case the 1st assert will warn about zone C not being scraped.

This second case - a zone not being scraped - should lead to an error in the
operator, causing the change to not be rolled out, as data would be lost.
The first case - double scraping - should at minimum cause a warning.

## Required configuration options

Following the algorithm presented above, we suggest the following configuration
options to be added to the [Prometheus custom resource definition](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#monitoring.coreos.com/v1.Prometheus).

All values used in this snippet should also be the defaults for their 
corresponding keys.

```yaml
spec:
  shardingStrategy:
     # Select a sharding mode. Can be 'Classic' or 'Topology'
    mode: 'Classic'    

    # Optional: array of relabel configurations to insert before the sharding
    # relabel configuration.
    additionalRelabelConfig: [] 

    # The following section is only valid if "mode" is set to "Classic"
    classic:
        # Metric label used for sharding.
        sourceLabel: '__address__'
      
    # The following section is only valid if "mode" is set to "Topology"
    # 'topology.kubernetes.io/zone' for 'topology'
    topology: 
        # A kubernetes node label defining the topology to be sharded on
        nodeLabel: 'topology.kubernetes.io/zone'
        
        # A prometheus metric containing the topology value of a given target
        sourceLabel: '__meta_kubernetes_pod_label_topology_kubernetes_io_zone'

        # All topology values to be used by nodeLabel and sourceLabel
        values: []
```

The `additionalRelabelConfig` section is meant to allow the `sourceLabel` to be
generated if needed. This should allow enough flexibility to cover edgecases
not anticiapted by this proposal.

It is also to be noted that the `topology` section does not use the term `zone`.
This makes the feature more flexible in case a user needs to shard on e.g.
regions instead.

## Generated configuration

In order to illustrate configuration generator mechanics, the following
examples use Helm templating syntax.
Variables are based on the proposed configuration structure as defined above.
Functions from the [sprig](https://masterminds.github.io/sprig/) library are 
used, as they are also present in Helm.

Values generated by the operator, like `shard_index` are retrieved from an
`.operator` object.

### classic mode

Given the following configuration

```yaml
spec:
  shards: 4
  shardingStrategy:
    mode: 'Classic'    
    additionalRelabelConfig: [] 
    classic:
      sourceLabel: '__address__'
```

And the relabel configuration being genereated through this template

```yaml
{{ .spec.shardingStrategy.additionalRelabelConfig | toYaml }}
- source_labels: [ '{{ default "__address__" .spec.shardingStrategy.classic.sourceLabel }}' ] 
  separator: ;
  modulus: {{ .spec.shards }}
  target_label: '__tmp_hash'
  replacement: '$1'
  action: 'hashmod'
- source_labels: ['__tmp_hash']
  separator: ;
  regex: '{{ .operator.shardIndex }}'
  replacement: '$1'
  action: 'keep'
```

we would get the following output for `shard_index == 2`

```yaml
- source_labels: [ '__address__' ] 
  separator: ;
  modulus: 4
  target_label: '__tmp_hash'
  replacement: '$1'
  action: 'hashmod'
- source_labels: ['__tmp_hash']
  separator: ;
  regex: '2'
  replacement: $1
  action: 'keep'
```

### topology mode

Given the following configuration

```yaml
spec:
  shards: 4
  shardingStrategy:
    mode: 'Topology'    
    additionalRelabelConfig: [] 
    topology:
      nodeLabel: 'topology.kubernetes.io/zone'
      sourceLabel: '__meta_kubernetes_pod_label_topology_kubernetes_io_zone'
      values:
        - europe-west4-a
        - europe-west4-b
```

And the relabel configuration being genereated through this template

```yaml
{{ $shardsPerZone := max 1 (.spec.shards | div (len .spec.shardingStrategy.topology.values)) }}
{{ $promZoneIdx := .operator.shardIndex | mod $shardsPerZone }}

{{ .spec.shardingStrategy.additionalRelabelConfig | toYaml }}
- source_labels: [ '{{ default "__meta_kubernetes_pod_label_topology_kubernetes_io_zone" .spec.shardingStrategy.topology.sourceLabel }}' ]
  separator: ;
  regex: '{{ index .spec.shardingStrategy.topology $promZoneIdx }}'
  replacement: '$1'
  action: 'keep'
- source_labels: [ '__address__' ] 
  separator: ;
  modulus: {{ $shardsPerZone }}
  target_label: '__tmp_hash'
  replacement: '$1'
  action: 'hashmod'
- source_labels: ['__tmp_hash']
  separator: ;
  regex: '{{ .operator.shardIndex | div $shardsPerZone }}'
  replacement: '$1'
  action: 'keep'
```

we would get the following output for `shard_index == 2`:

```yaml
- source_labels: [ '__meta_kubernetes_pod_label_topology_kubernetes_io_zone' ] 
  separator: ;
  regex: 'europe-west4-a'
  replacement: $1
  action: keep
- source_labels: [ '__address__' ] 
  separator: ;
  modulus: 2
  target_label: '__tmp_hash'
  replacement: '$1'
  action: 'hashmod'
- source_labels: [ '__tmp_hash' ]
  separator: ;
  regex: '1'
  replacement: '$1'
  action: 'keep'
```

## Prometheus instance zone assignment

To make sure that prometheus instances are deployed to the correct zone (of their
assigned target), we need to generate a [node affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity)
or a [node selector](https://kubernetes.io/docs/tasks/configure-pod-container/assign-pods-nodes/#create-a-pod-that-gets-scheduled-to-your-chosen-node).

As node selectors are simpler to manage, and node affinities might run into
ordering issues when a user defines their own affinities, node selectors should
be used.

We use the same Helm templating notation as above to illustrate the mechanics.

```yaml
{{ $shardsPerZone := max 1 (.spec.shards | div (len .spec.shardingStrategy.topology.values)) }}
spec:
  nodeSelector:
    '{{ .spec.shardingStrategy.topology.nodeLabel }}': '{{ index .spec.shardingStrategy.topology.values (.operator.shardIndex | div $shardsPerZone) }}'
```

Given this input:

```yaml
spec:
  shards: 4
  shardingStrategy:
    mode: 'Topology'    
    additionalRelabelConfig: [] 
    topology:
      nodeLabel: 'topology.kubernetes.io/zone'
      sourceLabel: '__meta_kubernetes_pod_label_topology_kubernetes_io_zone'
      values:
        - europe-west4-a
        - europe-west4-b
```

The following snippet would be generated for `shared_index == 2`:

```yaml
spec:
  nodeSelector:
    'topology.kubernetes.io/zone': 'europe-west4-a'
```


## Alternatives

N/A

## Action Plan

N/A
