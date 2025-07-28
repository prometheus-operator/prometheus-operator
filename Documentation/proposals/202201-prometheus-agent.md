# Prometheus Agent support

## Status

`Implemented`

## Summary

The Prometheus 2.32.0 release introduces the Prometheus Agent, a mode optimized for remote-write dominant scenarios. This document proposes extending the Prometheus Operator to allow running a Prometheus Agent with different deployment strategies.

## Background

The Prometheus Operator in its current state does not allow a simple way of deploying the Prometheus agent. A potential workaround has been described in a [Github comment](https://github.com/prometheus-operator/prometheus-operator/issues/3989#issuecomment-974137486), where the agent can be deployed through the existing Prometheus CRD by explicitly setting command-line arguments specific to the agent mode.

As described in the comment, one significant problem with this approach is that the Prometheus Operator always generates `alerts` and `rules` sections in the Prometheus config file. These sections are not allowed when running the agent so users need to take additional actions to pause reconciliation of the Prometheus CR, tweak the generated secret and then unpause reconciliation in order to resolve the problem. Alternatively, users can apply a strategic merge patch to the prometheus container as described in the kube-prometheus docs: [https://github.com/prometheus-operator/kube-prometheus/blob/main/docs/customizations/prometheus-agent.md](https://github.com/prometheus-operator/kube-prometheus/blob/main/docs/customizations/prometheus-agent.md)

While this workaround can be used as a stop-gap solution to unblock users in the short term, it has the drawback of needing additional steps which require understanding implementation details of the operator itself. In addition to this, overriding the value of the argument `--config.file` also requires knowledge of Prometheus Operator internals.

A lot of the fields supported by the current PrometheusSpec are not applicable to the agent mode. These fields are documented in the PrometheusAgent CRD section.

Finally, the Prometheus agent is significantly different from the Prometheus server in the way that it fits in a monitoring stack. Therefore, running it as a StatefulSet might not be the only possible deployment strategy, users might want to run it as a DaemonSet or a Deployment instead.

## Proposal

This document proposes introducing a PrometheusAgent CRD to allow users to run Prometheus in agent mode. Having a separate CRD allows the Prometheus and PrometheusAgent CRDs to evolve independently and expose parameters specific to each Prometheus mode.

For example, the PrometheusAgent CRD could have a `strategy` field indicating the deployment strategy for the agent, but no `alerting` field since alerts are not supported in agent mode. Even though there will be an upfront cost for introducing a new CRD, having separate APIs would simplify long-term maintenance by allowing the use of CRD validation mechanisms provided by Kubernetes.

In addition, dedicated APIs with mode-specific fields are self documenting since they remove the need to explicitly document which fields and field values are allowed or required for each individual mode. Users will also be able to get an easier overview of the different parameters they could set for each mode, which leads to a better user experience when using the operator.

Finally, the advantage of using a separate CRD is the possibility of using an alpha API version, which would clearly indicate that the CRD is still under development. The Prometheus CRD, on the other hand, has already been declared as v1 and adding experimental fields to it will be challenging from both documentation and implementation aspects.

### Prometheus Agent CRD

The PrometheusAgent CRD would be similar to the Prometheus CRD, with the exception of removing fields which are not applicable to the prometheus agent mode.

Here is the list of fields we want to exclude:
* `retention`
* `retentionSize`
* `disableCompaction`
* `evaluationInterval`
* `rules`
* `query`
* `ruleSelector`
* `ruleNamespaceSelector`
* `alerting`
* `remoteRead`
* `additionalAlertRelabelConfigs`
* `additionalAlertManagerConfigs`
* `thanos`
* `prometheusRulesExcludedFromEnforce`
* `queryLogFile`
* `allowOverlappingBlocks`

The `enabledFeatures` field can be validated for agent-specific features only, which include: `expand-external-labels`, `extra-scrape-metrics` and `new-service-discovery-manager`.

Finally, the `remoteWrite` field should be made required only for the agent since it is a mandatory configuration section in agent mode.

### Deployment Strategies

When using Prometheus in server mode, scraped samples are stored in memory and on disk. These samples need to be preserved during disruptions, such as pod replacements or cluster maintenance operations which cause evictions. Because of this, the Prometheus Operator currently deploys Prometheus instances as Kubernetes StatefulSets.

On the other hand, when running Prometheus in agent mode, samples are sent to a remote write target immediately, and are not kept locally for a long time. The only use-case for storing samples locally is to allow retries when remote write targets are not available. This is achieved by keeping scraped samples in a WAL for 2h at most. Samples which have been successfully sent to remote write targets are immediately removed from local storage.

Since the Prometheus agent has slightly different storage requirements, this proposal suggests allowing users to choose different deployment strategies.

#### Running the agent with cluster-wide scope

Even though the Prometheus agent has very little need for storage, there are still scenarios where sample data can be lost if persistent storage is not used. If a remote write target is unavailable and an agent pod is evicted at the same time, the samples collected during the unavailability window of the remote write target will be completely lost.

For this reason, the cluster-wide strategy would be implemented by deploying a StatefulSet, similarly to how `Prometheus` CRs are currently reconciled. This also allows for reusing existing code from the operator and delivering a working solution faster and with fewer changes. Familiarity with how StatefulSets work, together with the possibility to reuse existing code, were the primary reasons for choosing StatefulSets for this strategy over Deployments.

The following table documents the problems that could occur with a Deployment and StatefulSet strategy in different situations.

<table>
  <tr>
   <td>
   </td>
   <td><strong>Pod update</strong>
   </td>
   <td><strong>Network outage during pod update</strong>
   </td>
   <td><strong>Network outage during node drain</strong>
   </td>
   <td><strong>Cloud k8s node rotation</strong>
   </td>
   <td><strong>Non-graceful pod deletion</strong>
   </td>
  </tr>
  <tr>
   <td><strong>Deployment with emptyDir volume</strong>
   </td>
   <td>No delay in scrapes if the new pod is created before the old one is terminated
   </td>
   <td>Unsent samples will be lost. 
<p>
EmptyDir is tied to a pod <em>and</em> node, and data from the old pod will not be preserved.
   </td>
   <td>Unsent samples will be lost. 
<p>
EmptyDir is tied to a pod <em>and</em> node, and data from the old pod will not be preserved.
   </td>
   <td>Unsent samples will be lost
   </td>
   <td>Unsent samples will be lost. 
<p>
EmptyDir is tied to a pod <em>and</em> node, and data from the old pod will not be preserved.
   </td>
  </tr>
  <tr>
   <td><strong>Statefulset with a PVC</strong>
   </td>
   <td>Potential delay in a subsequent scrape due to recreation of the pod
   </td>
   <td>No data loss, the volume will contain all unsent data
   </td>
   <td>No data loss, the volume will contain all unsent data
   </td>
   <td>No data loss if a new pod scheduled to the same AZ node. May be stuck in pending state otherwise
   </td>
   <td>No data loss, the volume will contain all unsent data
   </td>
  </tr>
  <tr>
   <td><strong>Deployment or STS with replicas</strong>
   </td>
   <td>No delay, mitigated by replicas
   </td>
   <td>Unsent data will be lost if last replica terminated before network outage resolves
   </td>
   <td>No data loss, as other replicas are running on other nodes
   </td>
   <td>No data loss, as other replicas running on other nodes
   </td>
   <td>No data loss as other replicas untouched
   </td>
  </tr>
</table>

#### Running the agent with node-specific scope

This strategy has a built-in auto-scaling mechanism since each agent will scrape only a subset of the targets. As the cluster grows and more nodes are added to it, new agent instances will automatically be scheduled to scrape pods on those nodes. Even though the load distribution will not be perfect (targets on certain nodes might produce far more metrics than targets on other nodes), it is a simple way of adding some sort of load management.

Another advantage is that persistent storage can now be handled by mounting a host volume, a strategy commonly used by log collectors. The need for persistent storage is described in the StatefulSet strategy section.

The Grafana Agent config exposes a `host_filter` boolean flag which, when enabled, instructs the agent to only filter targets from the same node, in addition to the scrape config already provided. With this option, the same config can be used for agents running on multiple nodes, and the agents will automatically scrape targets from their own nodes. Such a config option is not yet available in Prometheus. An issue has already been raised [[3]](https://github.com/prometheus/prometheus/issues/9637) and there is an open PR for addressing it [[4]](https://github.com/prometheus/prometheus/pull/10004).

Until the upstream work has been completed, it could be possible to implement this strategy with a few tweaks:
* the operator could use the [downward API](https://kubernetes.io/docs/tasks/inject-data-application/downward-api-volume-expose-pod-information/#capabilities-of-the-downward-api) to inject the node name in the pods.
* the operator's config reloader already supports expansion of environment variables.

With this setup, the unexpanded Prometheus configuration would look as follows

```yaml
relabel_configs:
- source_labels: [__meta_kubernetes_pod_node_name]
  action: keep
  regex: $NODE_NAME

in the pod definition:
spec:
- container: config-reloader
  env:
- name: NODE_NAME
  valueFrom:
  fieldRef:
  fieldPath: spec.nodeName
```

## Additional implementation details

There has been a suggestion in [a Github comment](https://github.com/prometheus-operator/prometheus-operator/issues/3989#issuecomment-821249404) to introduce a ScrapeConfig CRD in parallel to adding the PrometheusAgent CRD, and “translate” PrometheusAgent CRs to ScrapeConfig CRs. The main challenge with this approach is that it significantly increases the scope of the work that needs to be done to support deploying Prometheus agents.

A leaner alternative would be to focus on implementing the PrometheusAgent CRD by reusing code from the existing Prometheus controller. The ScrapeConfig can then be introduced separately, and the PrometheusAgent can be the first CRD which gets migrated to it.

### Implementation steps

The first step in the implementation process would include creating the PrometheusAgent CRD and deploying the agent as a StatefulSet, similar to how the Prometheus CRD is currently reconciled. This will allow for reusing a lot of the existing codebase from the Prometheus controller and the new CRD can be released in a timely manner.

Subsequent steps would include iterating on users' feedback and either implementing different deployment strategies, or refining the existing one.

## References
* [1] [https://github.com/grafana/agent/blob/5bf8cf452fa76c75387e30b6373630923679221c/production/kubernetes/agent-bare.yaml#L43](https://github.com/grafana/agent/blob/5bf8cf452fa76c75387e30b6373630923679221c/production/kubernetes/agent-bare.yaml#L43)
* [2] [https://github.com/open-telemetry/opentelemetry-operator#deployment-modes](https://github.com/open-telemetry/opentelemetry-operator#deployment-modes)
* [3] [https://github.com/prometheus/prometheus/issues/9637](https://github.com/prometheus/prometheus/issues/9637)
* [4] [https://github.com/prometheus/prometheus/pull/10004](https://github.com/prometheus/prometheus/pull/10004)
