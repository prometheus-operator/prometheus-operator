# Zone aware sharding for Prometheus

* **Owners:**
  * [arnecls](https://github.com/arnecls)
* **Status:**
  * `Accepted`
* **Related Tickets:**
  * [#6437](https://github.com/prometheus-operator/prometheus-operator/issues/6437)
* **Other docs:**
  * [Well known kubernetes labels](https://kubernetes.io/docs/reference/labels-annotations-taints/#topologykubernetesiozone)
  * [AWS zone names](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-availability-zones)
  * [GCP zone names](https://cloud.google.com/compute/docs/regions-zones#available)
  * [Shard Autoscaling](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/proposals/202310-shard-autoscaling.md) design proposal.

This proposal describes how we can implement zone-aware sharding by adding
support for custom labels and zone configuration options to the existing
Prometheus configuration resources.

## Why

When running large, multi-zone clusters, Prometheus scraping can lead to an
increase in inter-zone traffic costs. A solution would be to deploy 1 Prometheus shard
per zone and configure each shard to scrape only the targets local to its zone. The
current sharding implementation can't solve the issue though. While
it's possible to customize the label (`__address__` by default) used for distributing the
targets to the Prometheus instances, there's no way to configure a single Prometheus
resource so that each shard is bound to a specific zone.

## Goals

* Define a set of configuration options required to allow zone aware sharding
* Define the relabel configuration to be generated for zone aware sharding
* Schedule Prometheus pods to their respective zones.
* Stay backwards compatible to the current mechanism by default

## Non-goals

* Implement mechanisms to automatically fix configuration errors by the user
* Support mixed environments (kubernetes and non-kubernetes targets are scraped)
* Support Kubernetes clusters before 1.26 (topology label support)
* Implement zone aware scraping for targets defined via
  `.spec.additionalScrapeConfigs` and `ScrapeConfig` custom resources.

## How

> [!NOTE]
> Due to the size of this feature, it will be placed behind a
> [feature gate](https://github.com/prometheus-operator/prometheus-operator/blob/main/pkg/operator/feature_gates.go)
> to allow incremental testing.

### Algorithm

In order to do calculate a stable assignment, following parameters are required:

1. `num_shards`: The number of Prometheus shards
2. `shard_index`: A number of the range `[0..num_shards-1]` identifying a single
   Prometheus instance inside a given shard
3. `zones`: A list of the zones to be scraped
4. `zone_label`: A label denoting the zone of a target
5. `address`: The content of the `__address__` label

It has to be noted that `zone_label` is supposed to contain a value from the
`zones` list.
The `num_shards` value is referring to the currently available `.spec.shards`
from the [Prometheus custom resource definition](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api-reference/api.md#monitoring.coreos.com/v1.Prometheus).

Given these values, a target is to be scraped by a given Prometheus instance
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

#### Edge cases

We have introduced asserts in the above section to warn about edge cases that
might lead to duplicate data or data loss.

By the above algorithm. Prometheus instances will be distributed in an
alternating fashion by using the already existing shard index.
This leads to the following edge cases:

When `num_shards` is 10 and `len(zones)` is 3 as in `[A..C]`.
`shards_per_zone` is 3. This yields the following distribution

```text
0 1 2 | 3 4 5 | 6 7 8 | 9 | shard index
A B C | A B C | A B C | A | zone
0 0 0 | 1 1 1 | 2 2 2 | 0 | assignment index
```

In this case the 2nd assert will warn about double scraping of instances in
zone A, as the same targets are being assigned to instance 0 and 9.

When `num_shards` is 2 and `len(zones)` is 3 as in `[A..C]`.
`shards_per_zone` is 1. This yields the following distribution

```text
0 1   | shard index
A B C | zone
0 0 0 | assignment index
```

In this case targets in zone C are not being scraped.

Both cases should lead to an error during reconciliation, causing the change to
not be rolled out. The first case (double scraping) is not as severe as a zone
not being scraped but it is otherwise hard to spot.
It's also to be mentioned that replicas are to be used to achieve redundant
scraping.

### Topology field discovery

The kubernetes service discovery currently does not expose any topology field.
Such a field would have to be added, otherwise users would have to inject such
a label themselves.

A good candidate for such fields are the `topology.kubernetes.io/*` labels
which should be present on all nodes.

There are two ways to handle this:

1. A change to the Prometheus kubernetes discovery service to add the required
   label to all targets.
2. The operator could do this discovery and add a relabel rule based on the
   node name.

The second solution would require the operator to constantly update the relabel
configuration. This could lead to increased load on clusters with agressive
autoscaling as well as race conditions for pods on newly created nodes, as the
config change is not atomic/instant.

As of that, a change to the kubernetes service discovery is considered the more
stable, and thus preferrable solution. It will require additional permissions
for Prometheus in case it is not already allowed to read node objects.

### API changes

> [!NOTE]
> This proposal is mutually exclusive to [DaemonSet mode](202405-agent-daemonset.md),
> as Prometheus always scrapes a single node in that case.
> Defining a `shardingStrategy` when `DaemonSet mode` is active, should lead to
> a reconciliation error.

Following the algorithm presented above, we suggest the following configuration
options to be added to the [Prometheus](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api-reference/api.md#monitoring.coreos.com/v1.Prometheus) and PrometheusAgent custom resource definitions.

All values used in this snippet should also be the defaults for their
corresponding keys.

```yaml
spec:
  shardingStrategy:
     # Select a sharding mode. Can be 'Classic' or 'Topology'.
     # Defaults to `Classic`.
    mode: 'Classic'

    # The following section is only valid if "mode" is set to "Topology"
    topology:
      # Prometheus external label used to communicate the topology zone.
      # If not defined, it defaults to "zone".
      # If defined to an empty string, no external label is added to the Prometheus configuration.
      externalLabelName: "zone"
      # All topology values to be used by the cluster, i.e. a list of all
      # zones in use.
      values: []
```

The `topology` section does not use the term `zone`. This will prevent API
changes in case other topologies, like regions, need to be supported in future
releases.

Both modes do not contain an explicit overwrite of the label used for sharding.
This feature is already possible by generating a `__tmp_hash` label through
[scrape classes](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api-reference/api.md#monitoring.coreos.com/v1.ScrapeClass).

In case of the `Topology` mode, two labels are used for sharding. One is used
to determine the correct topology of a target, the other one is used to allow
sharding inside a specfic topology (e.g. zone).
The second label implements the exact same mechanics as the `Classic` mode and
thus uses the same `__tmp_hash` overwrite mechanics.
To allow overwrites for the topology determination label, a custom label named
`__tmp_topology` can be generated, following the same idea.

The `externalLabelName` should be added by default to allow debugging. It also
gives some general, valuable insights for multi-zone setups.

It is possible to change the `mode` field from `Classic` to `Topology` (and
vice-versa) without service interruption.

### Generated configuration

The following examples are based on the algorithm above.
Please note that `shard_index` has to be provided by the operator during
config generation.

We use a replica count of 2 in all examples to illustrate that this value
does not have any effect, as both replicas will have the same `shared_index`
assigned.

#### Classic mode

Given the following configuration

```yaml
spec:
  shards: 4
  replicas: 2
  shardingStrategy:
    mode: 'Classic'
```

we would get the following output for `shard_index == 2`

```yaml
- source_labels: ['__address__', '__tmp_hash']
  target_label: '__tmp_hash'
  regex: '(.+);'
  replacement: '$1'
  action: 'replace'
- source_labels: ['__tmp_hash']
  target_label: '__tmp_hash'
  modulus: 4
  action: 'hashmod'
- source_labels: ['__tmp_hash']
  regex: '2'                    # shard_index
  action: 'keep'
```

#### Topology mode

Given the following configuration

```yaml
spec:
  shards: 4
  replicas: 2
  shardingStrategy:
    mode: 'Topology'
    topology:
      values:
        - 'europe-west4-a'
        - 'europe-west4-b'
```

we would get the following output for `shard_index == 2`:

```yaml
# zones := shardingStrategy.topology.values
# shards_per_zone := max(1, floor(shards / len(zones)))

# topology determination
- source_labels: ['__meta_kubernetes_endpointslice_endpoint_zone', __tmp_topology]
  target_label: '__tmp_topology'
  regex: '(.+);'
- source_labels: ['__meta_kubernetes_node_label_topology_kubernetes_io_zone', '__meta_kubernetes_node_labelpresent_topology_kubernetes_io_zone', '__tmp_topology']
  regex: '(.+);true;'
  target_label: '__tmp_topology'
- source_labels: ['__tmp_topology']
  regex: 'europe-west4-a'          # zones[shard_index % shards_per_zone]
  action: 'keep'

# In-topology sharding
- source_labels: ['__address__', '__tmp_hash']
  target_label: '__tmp_hash'
  regex: "(.+);"
  action: 'replace'
- source_labels: ['__tmp_hash']
  target_label: '__tmp_hash'
  modulus: 4
  action: 'hashmod'
- source_labels: [ '__tmp_hash' ]
  regex: '1'                       # floor(shard_index / shards_per_zone)
  action: 'keep'
```

> [!NOTE]
> Node metadata will need to be added when using certain monitors.
> This requires an additional flag in the `kubernetes_sd_configs` section like this:
>
> ```yaml
> kubernetes_sd_configs:
>   - attach_metadata:
>       node: true
> ```

### Prometheus instance zone assignment

To make sure that Prometheus instances are deployed to the correct zone (of their
assigned target), we need to generate a [node affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity)
or a [node selector](https://kubernetes.io/docs/tasks/configure-pod-container/assign-pods-nodes/#create-a-pod-that-gets-scheduled-to-your-chosen-node).

As node selectors are simpler to manage, and node affinities might run into
ordering issues when a user defines their own affinities, node selectors should
be used.
If a `nodeSelector` has already been defined, it will be merged with the node
selector generated here. If the same key was used, the value will be
replaced with the generated value.

Given this input:

```yaml
spec:
  nodeSelector:
    'foo': 'bar'
    'topology.kubernetes.io/zone': 'will be replaced'
  shards: 4
  replicas: 2
  scrapeClasses:
    - name: 'topology'
      default: true
      attachMetadata:
        node: true
  shardingStrategy:
    mode: 'Topology'
    topology:
      values:
        - 'europe-west4-a'
        - 'europe-west4-b'
```

The following snippet would be generated for `shared_index == 2`:

```yaml
# zones := shardingStrategy.topology.values
# shards_per_zone := max(1, floor(shards / len(zones)))
spec:
  nodeSelector:
    # Existing nodeSelectors using 'topology.kubernetes.io/zone' will be
    # replaced with the generated value:
    # zones[shard_index % shards_per_zone]
    'topology.kubernetes.io/zone': 'europe-west4-a'
    # Existing nodeSelectors will be kept
    'foo': 'bar'
```

## Alternatives

We could allow users to define the complete relabel and node selector logic
themselves. This would be more flexible, but also way harder to configure.

By abstracting into `shardingStrategy`, we can cover the most common cases
without requiring users to have deep knowledge about Prometheus relabel
configuration.

A field `additionalRelabelConfig` was discussed to allow arbitrary logic to be
added before the sharding configuration. It was decided that this would
duplicate the functionality of [scrape classes](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api-reference/api.md#monitoring.coreos.com/v1.ScrapeClass)
found in, e.g., the [Prometheus](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api-reference/api.md#prometheusspec)
custom resource definition.

The use of `sourceLabel` fields instead of the `__tmp-*` label mechanic was
discussed. It was agreed to not introduce new fields as an existing feature
would have to be changed, and to not further increase the number of API fields.

An overwrite for the topology node label was discussed. This field was not
added as there was no clear use-case yet. The general structure was kept so
that it will still be possible to add such a field in a future release.

## Action Plan

A rough plan of the steps required to implement the feature:

1. Add the `PrometheusTopologySharding` feature gate.
2. Implement the API changes with pre-flight validations.
3. Implement the node selector update when `mode: Topology`.
4. Implement the external label name when `mode: Topology`.
5. Implement the target sharding when `mode: Topology`.
