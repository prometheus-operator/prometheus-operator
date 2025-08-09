---
title: "Monitoring v1beta1 API Reference"
description: "Generated API reference for monitoring.coreos.com/v1beta1"
draft: false
images: []
menu: "operator"
weight: 154
toc: true
---
> This page is automatically generated with `gen-crd-api-reference-docs`.
<h2 id="monitoring.coreos.com/v1alpha1">monitoring.coreos.com/v1alpha1</h2>
Resource Types:
<ul><li>
<a href="#monitoring.coreos.com/v1alpha1.AlertmanagerConfig">AlertmanagerConfig</a>
</li><li>
<a href="#monitoring.coreos.com/v1alpha1.PrometheusAgent">PrometheusAgent</a>
</li><li>
<a href="#monitoring.coreos.com/v1alpha1.ScrapeConfig">ScrapeConfig</a>
</li></ul>
<h3 id="monitoring.coreos.com/v1alpha1.AlertmanagerConfig">AlertmanagerConfig
</h3>
<div>
<p>AlertmanagerConfig configures the Prometheus Alertmanager,
specifying how alerts should be grouped, inhibited and notified to external systems.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code><br/>
string</td>
<td>
<code>
monitoring.coreos.com/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>AlertmanagerConfig</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.AlertmanagerConfigSpec">
AlertmanagerConfigSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>route</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.Route">
Route
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The Alertmanager route definition for alerts matching the resource&rsquo;s
namespace. If present, it will be added to the generated Alertmanager
configuration as a first-level route.</p>
</td>
</tr>
<tr>
<td>
<code>receivers</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.Receiver">
[]Receiver
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of receivers.</p>
</td>
</tr>
<tr>
<td>
<code>inhibitRules</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.InhibitRule">
[]InhibitRule
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of inhibition rules. The rules will only apply to alerts matching
the resource&rsquo;s namespace.</p>
</td>
</tr>
<tr>
<td>
<code>muteTimeIntervals</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.MuteTimeInterval">
[]MuteTimeInterval
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of MuteTimeInterval specifying when the routes should be muted.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.PrometheusAgent">PrometheusAgent
</h3>
<div>
<p>The <code>PrometheusAgent</code> custom resource definition (CRD) defines a desired <a href="https://prometheus.io/blog/2021/11/16/agent/">Prometheus Agent</a> setup to run in a Kubernetes cluster.</p>
<p>The CRD is very similar to the <code>Prometheus</code> CRD except for features which aren&rsquo;t available in agent mode like rule evaluation, persistent storage and Thanos sidecar.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code><br/>
string</td>
<td>
<code>
monitoring.coreos.com/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>PrometheusAgent</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.PrometheusAgentSpec">
PrometheusAgentSpec
</a>
</em>
</td>
<td>
<p>Specification of the desired behavior of the Prometheus agent. More info:
<a href="https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status">https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status</a></p>
<br/>
<br/>
<table>
<tr>
<td>
<code>mode</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.PrometheusAgentMode">
PrometheusAgentMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Mode defines how the Prometheus operator deploys the PrometheusAgent pod(s).</p>
<p>(Alpha) Using this field requires the <code>PrometheusAgentDaemonSet</code> feature gate to be enabled.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
Monitoring v1.EmbeddedObjectMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata configures labels and annotations which are propagated to the Prometheus pods.</p>
<p>The following items are reserved and cannot be overridden:
* &ldquo;prometheus&rdquo; label, set to the name of the Prometheus object.
* &ldquo;app.kubernetes.io/instance&rdquo; label, set to the name of the Prometheus object.
* &ldquo;app.kubernetes.io/managed-by&rdquo; label, set to &ldquo;prometheus-operator&rdquo;.
* &ldquo;app.kubernetes.io/name&rdquo; label, set to &ldquo;prometheus&rdquo;.
* &ldquo;app.kubernetes.io/version&rdquo; label, set to the Prometheus version.
* &ldquo;operator.prometheus.io/name&rdquo; label, set to the name of the Prometheus object.
* &ldquo;operator.prometheus.io/shard&rdquo; label, set to the shard number of the Prometheus object.
* &ldquo;kubectl.kubernetes.io/default-container&rdquo; annotation, set to &ldquo;prometheus&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitorSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ServiceMonitors to be selected for target discovery. An empty label
selector matches all objects. A null label selector matches no objects.</p>
<p>If <code>spec.serviceMonitorSelector</code>, <code>spec.podMonitorSelector</code>, <code>spec.probeSelector</code>
and <code>spec.scrapeConfigSelector</code> are null, the Prometheus configuration is unmanaged.
The Prometheus operator will ensure that the Prometheus configuration&rsquo;s
Secret exists, but it is the responsibility of the user to provide the raw
gzipped Prometheus configuration under the <code>prometheus.yaml.gz</code> key.
This behavior is <em>deprecated</em> and will be removed in the next major version
of the custom resource definition. It is recommended to use
<code>spec.additionalScrapeConfigs</code> instead.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitorNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for ServicedMonitors discovery. An empty label selector
matches all namespaces. A null label selector (default value) matches the current
namespace only.</p>
</td>
</tr>
<tr>
<td>
<code>podMonitorSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>PodMonitors to be selected for target discovery. An empty label selector
matches all objects. A null label selector matches no objects.</p>
<p>If <code>spec.serviceMonitorSelector</code>, <code>spec.podMonitorSelector</code>, <code>spec.probeSelector</code>
and <code>spec.scrapeConfigSelector</code> are null, the Prometheus configuration is unmanaged.
The Prometheus operator will ensure that the Prometheus configuration&rsquo;s
Secret exists, but it is the responsibility of the user to provide the raw
gzipped Prometheus configuration under the <code>prometheus.yaml.gz</code> key.
This behavior is <em>deprecated</em> and will be removed in the next major version
of the custom resource definition. It is recommended to use
<code>spec.additionalScrapeConfigs</code> instead.</p>
</td>
</tr>
<tr>
<td>
<code>podMonitorNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for PodMonitors discovery. An empty label selector
matches all namespaces. A null label selector (default value) matches the current
namespace only.</p>
</td>
</tr>
<tr>
<td>
<code>probeSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Probes to be selected for target discovery. An empty label selector
matches all objects. A null label selector matches no objects.</p>
<p>If <code>spec.serviceMonitorSelector</code>, <code>spec.podMonitorSelector</code>, <code>spec.probeSelector</code>
and <code>spec.scrapeConfigSelector</code> are null, the Prometheus configuration is unmanaged.
The Prometheus operator will ensure that the Prometheus configuration&rsquo;s
Secret exists, but it is the responsibility of the user to provide the raw
gzipped Prometheus configuration under the <code>prometheus.yaml.gz</code> key.
This behavior is <em>deprecated</em> and will be removed in the next major version
of the custom resource definition. It is recommended to use
<code>spec.additionalScrapeConfigs</code> instead.</p>
</td>
</tr>
<tr>
<td>
<code>probeNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for Probe discovery. An empty label
selector matches all namespaces. A null label selector matches the
current namespace only.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeConfigSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ScrapeConfigs to be selected for target discovery. An empty label
selector matches all objects. A null label selector matches no objects.</p>
<p>If <code>spec.serviceMonitorSelector</code>, <code>spec.podMonitorSelector</code>, <code>spec.probeSelector</code>
and <code>spec.scrapeConfigSelector</code> are null, the Prometheus configuration is unmanaged.
The Prometheus operator will ensure that the Prometheus configuration&rsquo;s
Secret exists, but it is the responsibility of the user to provide the raw
gzipped Prometheus configuration under the <code>prometheus.yaml.gz</code> key.
This behavior is <em>deprecated</em> and will be removed in the next major version
of the custom resource definition. It is recommended to use
<code>spec.additionalScrapeConfigs</code> instead.</p>
<p>Note that the ScrapeConfig custom resource definition is currently at Alpha level.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeConfigNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespaces to match for ScrapeConfig discovery. An empty label selector
matches all namespaces. A null label selector matches the current
namespace only.</p>
<p>Note that the ScrapeConfig custom resource definition is currently at Alpha level.</p>
</td>
</tr>
<tr>
<td>
<code>version</code><br/>
<em>
string
</em>
</td>
<td>
<p>Version of Prometheus being deployed. The operator uses this information
to generate the Prometheus StatefulSet + configuration files.</p>
<p>If not specified, the operator assumes the latest upstream version of
Prometheus available at the time when the version of the operator was
released.</p>
</td>
</tr>
<tr>
<td>
<code>paused</code><br/>
<em>
bool
</em>
</td>
<td>
<p>When a Prometheus deployment is paused, no actions except for deletion
will be performed on the underlying objects.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Container image name for Prometheus. If specified, it takes precedence
over the <code>spec.baseImage</code>, <code>spec.tag</code> and <code>spec.sha</code> fields.</p>
<p>Specifying <code>spec.version</code> is still necessary to ensure the Prometheus
Operator knows which version of Prometheus is being configured.</p>
<p>If neither <code>spec.image</code> nor <code>spec.baseImage</code> are defined, the operator
will use the latest upstream version of Prometheus available at the time
when the operator was released.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#pullpolicy-v1-core">
Kubernetes core/v1.PullPolicy
</a>
</em>
</td>
<td>
<p>Image pull policy for the &lsquo;prometheus&rsquo;, &lsquo;init-config-reloader&rsquo; and &lsquo;config-reloader&rsquo; containers.
See <a href="https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy">https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy</a> for more details.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>An optional list of references to Secrets in the same namespace
to use for pulling images from registries.
See <a href="http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod">http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod</a></p>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Number of replicas of each shard to deploy for a Prometheus deployment.
<code>spec.replicas</code> multiplied by <code>spec.shards</code> is the total number of Pods
created.</p>
<p>Default: 1</p>
</td>
</tr>
<tr>
<td>
<code>shards</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Number of shards to distribute the scraped targets onto.</p>
<p><code>spec.replicas</code> multiplied by <code>spec.shards</code> is the total number of Pods
being created.</p>
<p>When not defined, the operator assumes only one shard.</p>
<p>Note that scaling down shards will not reshard data onto the remaining
instances, it must be manually moved. Increasing shards will not reshard
data either but it will continue to be available from the same
instances. To query globally, use either
* Thanos sidecar + querier for query federation and Thanos Ruler for rules.
* Remote-write to send metrics to a central location.</p>
<p>By default, the sharding of targets is performed on:
* The <code>__address__</code> target&rsquo;s metadata label for PodMonitor,
ServiceMonitor and ScrapeConfig resources.
* The <code>__param_target__</code> label for Probe resources.</p>
<p>Users can define their own sharding implementation by setting the
<code>__tmp_hash</code> label during the target discovery with relabeling
configuration (either in the monitoring resources or via scrape class).</p>
<p>You can also disable sharding on a specific target by setting the
<code>__tmp_disable_sharding</code> label with relabeling configuration. When
the label value isn&rsquo;t empty, all Prometheus shards will scrape the target.</p>
</td>
</tr>
<tr>
<td>
<code>replicaExternalLabelName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of Prometheus external label used to denote the replica name.
The external label will <em>not</em> be added when the field is set to the
empty string (<code>&quot;&quot;</code>).</p>
<p>Default: &ldquo;prometheus_replica&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>prometheusExternalLabelName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of Prometheus external label used to denote the Prometheus instance
name. The external label will <em>not</em> be added when the field is set to
the empty string (<code>&quot;&quot;</code>).</p>
<p>Default: &ldquo;prometheus&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>logLevel</code><br/>
<em>
string
</em>
</td>
<td>
<p>Log level for Prometheus and the config-reloader sidecar.</p>
</td>
</tr>
<tr>
<td>
<code>logFormat</code><br/>
<em>
string
</em>
</td>
<td>
<p>Log format for Log level for Prometheus and the config-reloader sidecar.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<p>Interval between consecutive scrapes.</p>
<p>Default: &ldquo;30s&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>scrapeTimeout</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<p>Number of seconds to wait until a scrape request times out.
The value cannot be greater than the scrape interval otherwise the operator will reject the resource.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeProtocols</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ScrapeProtocol">
[]Monitoring v1.ScrapeProtocol
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The protocols to negotiate during a scrape. It tells clients the
protocols supported by Prometheus in order of preference (from most to least preferred).</p>
<p>If unset, Prometheus uses its default value.</p>
<p>It requires Prometheus &gt;= v2.49.0.</p>
<p><code>PrometheusText1.0.0</code> requires Prometheus &gt;= v3.0.0.</p>
</td>
</tr>
<tr>
<td>
<code>externalLabels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>The labels to add to any time series or alerts when communicating with
external systems (federation, remote storage, Alertmanager).
Labels defined by <code>spec.replicaExternalLabelName</code> and
<code>spec.prometheusExternalLabelName</code> take precedence over this list.</p>
</td>
</tr>
<tr>
<td>
<code>enableRemoteWriteReceiver</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Enable Prometheus to be used as a receiver for the Prometheus remote
write protocol.</p>
<p>WARNING: This is not considered an efficient way of ingesting samples.
Use it with caution for specific low-volume use cases.
It is not suitable for replacing the ingestion via scraping and turning
Prometheus into a push-based metrics collection system.
For more information see <a href="https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver">https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver</a></p>
<p>It requires Prometheus &gt;= v2.33.0.</p>
</td>
</tr>
<tr>
<td>
<code>enableOTLPReceiver</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enable Prometheus to be used as a receiver for the OTLP Metrics protocol.</p>
<p>Note that the OTLP receiver endpoint is automatically enabled if <code>.spec.otlpConfig</code> is defined.</p>
<p>It requires Prometheus &gt;= v2.47.0.</p>
</td>
</tr>
<tr>
<td>
<code>remoteWriteReceiverMessageVersions</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.RemoteWriteMessageVersion">
[]Monitoring v1.RemoteWriteMessageVersion
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of the protobuf message versions to accept when receiving the
remote writes.</p>
<p>It requires Prometheus &gt;= v2.54.0.</p>
</td>
</tr>
<tr>
<td>
<code>enableFeatures</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.EnableFeature">
[]Monitoring v1.EnableFeature
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enable access to Prometheus feature flags. By default, no features are enabled.</p>
<p>Enabling features which are disabled by default is entirely outside the
scope of what the maintainers will support and by doing so, you accept
that this behaviour may break at any time without notice.</p>
<p>For more information see <a href="https://prometheus.io/docs/prometheus/latest/feature_flags/">https://prometheus.io/docs/prometheus/latest/feature_flags/</a></p>
</td>
</tr>
<tr>
<td>
<code>externalUrl</code><br/>
<em>
string
</em>
</td>
<td>
<p>The external URL under which the Prometheus service is externally
available. This is necessary to generate correct URLs (for instance if
Prometheus is accessible behind an Ingress resource).</p>
</td>
</tr>
<tr>
<td>
<code>routePrefix</code><br/>
<em>
string
</em>
</td>
<td>
<p>The route prefix Prometheus registers HTTP handlers for.</p>
<p>This is useful when using <code>spec.externalURL</code>, and a proxy is rewriting
HTTP routes of a request, and the actual ExternalURL is still true, but
the server serves requests under a different route prefix. For example
for use with <code>kubectl proxy</code>.</p>
</td>
</tr>
<tr>
<td>
<code>storage</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.StorageSpec">
Monitoring v1.StorageSpec
</a>
</em>
</td>
<td>
<p>Storage defines the storage used by Prometheus.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<p>Volumes allows the configuration of additional volumes on the output
StatefulSet definition. Volumes specified will be appended to other
volumes that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<p>VolumeMounts allows the configuration of additional VolumeMounts.</p>
<p>VolumeMounts will be appended to other VolumeMounts in the &lsquo;prometheus&rsquo;
container, that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>persistentVolumeClaimRetentionPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#statefulsetpersistentvolumeclaimretentionpolicy-v1-apps">
Kubernetes apps/v1.StatefulSetPersistentVolumeClaimRetentionPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The field controls if and how PVCs are deleted during the lifecycle of a StatefulSet.
The default behavior is all PVCs are retained.
This is an alpha field from kubernetes 1.23 until 1.26 and a beta field from 1.26.
It requires enabling the StatefulSetAutoDeletePVC feature gate.</p>
</td>
</tr>
<tr>
<td>
<code>web</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.PrometheusWebSpec">
Monitoring v1.PrometheusWebSpec
</a>
</em>
</td>
<td>
<p>Defines the configuration of the Prometheus web server.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>Defines the resources requests and limits of the &lsquo;prometheus&rsquo; container.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>Defines on which Nodes the Pods are scheduled.</p>
</td>
</tr>
<tr>
<td>
<code>serviceAccountName</code><br/>
<em>
string
</em>
</td>
<td>
<p>ServiceAccountName is the name of the ServiceAccount to use to run the
Prometheus Pods.</p>
</td>
</tr>
<tr>
<td>
<code>automountServiceAccountToken</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>AutomountServiceAccountToken indicates whether a service account token should be automatically mounted in the pod.
If the field isn&rsquo;t set, the operator mounts the service account token by default.</p>
<p><strong>Warning:</strong> be aware that by default, Prometheus requires the service account token for Kubernetes service discovery.
It is possible to use strategic merge patch to project the service account token into the &lsquo;prometheus&rsquo; container.</p>
</td>
</tr>
<tr>
<td>
<code>secrets</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Secrets is a list of Secrets in the same namespace as the Prometheus
object, which shall be mounted into the Prometheus Pods.
Each Secret is added to the StatefulSet definition as a volume named <code>secret-&lt;secret-name&gt;</code>.
The Secrets are mounted into /etc/prometheus/secrets/<secret-name> in the &lsquo;prometheus&rsquo; container.</p>
</td>
</tr>
<tr>
<td>
<code>configMaps</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>ConfigMaps is a list of ConfigMaps in the same namespace as the Prometheus
object, which shall be mounted into the Prometheus Pods.
Each ConfigMap is added to the StatefulSet definition as a volume named <code>configmap-&lt;configmap-name&gt;</code>.
The ConfigMaps are mounted into /etc/prometheus/configmaps/<configmap-name> in the &lsquo;prometheus&rsquo; container.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the Pods&rsquo; affinity scheduling rules if specified.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the Pods&rsquo; tolerations if specified.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.TopologySpreadConstraint">
[]Monitoring v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the pod&rsquo;s topology spread constraints if specified.</p>
</td>
</tr>
<tr>
<td>
<code>remoteWrite</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.RemoteWriteSpec">
[]Monitoring v1.RemoteWriteSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the list of remote write configurations.</p>
</td>
</tr>
<tr>
<td>
<code>otlp</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OTLPConfig">
Monitoring v1.OTLPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Settings related to the OTLP receiver feature.
It requires Prometheus &gt;= v2.55.0.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecurityContext holds pod-level security attributes and common container settings.
This defaults to the default PodSecurityContext.</p>
</td>
</tr>
<tr>
<td>
<code>dnsPolicy</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.DNSPolicy">
Monitoring v1.DNSPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the DNS policy for the pods.</p>
</td>
</tr>
<tr>
<td>
<code>dnsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.PodDNSConfig">
Monitoring v1.PodDNSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the DNS configuration for the pods.</p>
</td>
</tr>
<tr>
<td>
<code>listenLocal</code><br/>
<em>
bool
</em>
</td>
<td>
<p>When true, the Prometheus server listens on the loopback address
instead of the Pod IP&rsquo;s address.</p>
</td>
</tr>
<tr>
<td>
<code>enableServiceLinks</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Indicates whether information about services should be injected into pod&rsquo;s environment variables</p>
</td>
</tr>
<tr>
<td>
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Containers allows injecting additional containers or modifying operator
generated containers. This can be used to allow adding an authentication
proxy to the Pods or to change the behavior of an operator generated
container. Containers described here modify an operator generated
container if they share the same name and modifications are done via a
strategic merge patch.</p>
<p>The names of containers managed by the operator are:
* <code>prometheus</code>
* <code>config-reloader</code>
* <code>thanos-sidecar</code></p>
<p>Overriding containers is entirely outside the scope of what the
maintainers will support and by doing so, you accept that this behaviour
may break at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>initContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>InitContainers allows injecting initContainers to the Pod definition. Those
can be used to e.g.  fetch secrets for injection into the Prometheus
configuration from external sources. Any errors during the execution of
an initContainer will lead to a restart of the Pod. More info:
<a href="https://kubernetes.io/docs/concepts/workloads/pods/init-containers/">https://kubernetes.io/docs/concepts/workloads/pods/init-containers/</a>
InitContainers described here modify an operator generated init
containers if they share the same name and modifications are done via a
strategic merge patch.</p>
<p>The names of init container name managed by the operator are:
* <code>init-config-reloader</code>.</p>
<p>Overriding init containers is entirely outside the scope of what the
maintainers will support and by doing so, you accept that this behaviour
may break at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>additionalScrapeConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AdditionalScrapeConfigs allows specifying a key of a Secret containing
additional Prometheus scrape configurations. Scrape configurations
specified are appended to the configurations generated by the Prometheus
Operator. Job configurations specified must have the form as specified
in the official Prometheus documentation:
<a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config</a>.
As scrape configs are appended, the user is responsible to make sure it
is valid. Note that using this feature may expose the possibility to
break upgrades of Prometheus. It is advised to review Prometheus release
notes to ensure that no incompatible scrape configs are going to break
Prometheus after the upgrade.</p>
</td>
</tr>
<tr>
<td>
<code>apiserverConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.APIServerConfig">
Monitoring v1.APIServerConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>APIServerConfig allows specifying a host and auth methods to access the
Kuberntees API server.
If null, Prometheus is assumed to run inside of the cluster: it will
discover the API servers automatically and use the Pod&rsquo;s CA certificate
and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/.</p>
</td>
</tr>
<tr>
<td>
<code>priorityClassName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Priority class assigned to the Pods.</p>
</td>
</tr>
<tr>
<td>
<code>portName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Port name used for the pods and governing service.
Default: &ldquo;web&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>arbitraryFSAccessThroughSMs</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ArbitraryFSAccessThroughSMsConfig">
Monitoring v1.ArbitraryFSAccessThroughSMsConfig
</a>
</em>
</td>
<td>
<p>When true, ServiceMonitor, PodMonitor and Probe object are forbidden to
reference arbitrary files on the file system of the &lsquo;prometheus&rsquo;
container.
When a ServiceMonitor&rsquo;s endpoint specifies a <code>bearerTokenFile</code> value
(e.g.  &lsquo;/var/run/secrets/kubernetes.io/serviceaccount/token&rsquo;), a
malicious target can get access to the Prometheus service account&rsquo;s
token in the Prometheus&rsquo; scrape request. Setting
<code>spec.arbitraryFSAccessThroughSM</code> to &lsquo;true&rsquo; would prevent the attack.
Users should instead provide the credentials using the
<code>spec.bearerTokenSecret</code> field.</p>
</td>
</tr>
<tr>
<td>
<code>overrideHonorLabels</code><br/>
<em>
bool
</em>
</td>
<td>
<p>When true, Prometheus resolves label conflicts by renaming the labels in the scraped data
to “exported_” for all targets created from ServiceMonitor, PodMonitor and
ScrapeConfig objects. Otherwise the HonorLabels field of the service or pod monitor applies.
In practice,<code>overrideHonorLaels:true</code> enforces <code>honorLabels:false</code>
for all ServiceMonitor, PodMonitor and ScrapeConfig objects.</p>
</td>
</tr>
<tr>
<td>
<code>overrideHonorTimestamps</code><br/>
<em>
bool
</em>
</td>
<td>
<p>When true, Prometheus ignores the timestamps for all the targets created
from service and pod monitors.
Otherwise the HonorTimestamps field of the service or pod monitor applies.</p>
</td>
</tr>
<tr>
<td>
<code>ignoreNamespaceSelectors</code><br/>
<em>
bool
</em>
</td>
<td>
<p>When true, <code>spec.namespaceSelector</code> from all PodMonitor, ServiceMonitor
and Probe objects will be ignored. They will only discover targets
within the namespace of the PodMonitor, ServiceMonitor and Probe
object.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedNamespaceLabel</code><br/>
<em>
string
</em>
</td>
<td>
<p>When not empty, a label will be added to:</p>
<ol>
<li>All metrics scraped from <code>ServiceMonitor</code>, <code>PodMonitor</code>, <code>Probe</code> and <code>ScrapeConfig</code> objects.</li>
<li>All metrics generated from recording rules defined in <code>PrometheusRule</code> objects.</li>
<li>All alerts generated from alerting rules defined in <code>PrometheusRule</code> objects.</li>
<li>All vector selectors of PromQL expressions defined in <code>PrometheusRule</code> objects.</li>
</ol>
<p>The label will not added for objects referenced in <code>spec.excludedFromEnforcement</code>.</p>
<p>The label&rsquo;s name is this field&rsquo;s value.
The label&rsquo;s value is the namespace of the <code>ServiceMonitor</code>,
<code>PodMonitor</code>, <code>Probe</code>, <code>PrometheusRule</code> or <code>ScrapeConfig</code> object.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedSampleLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>When defined, enforcedSampleLimit specifies a global limit on the number
of scraped samples that will be accepted. This overrides any
<code>spec.sampleLimit</code> set by ServiceMonitor, PodMonitor, Probe objects
unless <code>spec.sampleLimit</code> is greater than zero and less than
<code>spec.enforcedSampleLimit</code>.</p>
<p>It is meant to be used by admins to keep the overall number of
samples/series under a desired limit.</p>
<p>When both <code>enforcedSampleLimit</code> and <code>sampleLimit</code> are defined and greater than zero, the following rules apply:
* Scrape objects without a defined sampleLimit value will inherit the global sampleLimit value (Prometheus &gt;= 2.45.0) or the enforcedSampleLimit value (Prometheus &lt; v2.45.0).
If Prometheus version is &gt;= 2.45.0 and the <code>enforcedSampleLimit</code> is greater than the <code>sampleLimit</code>, the <code>sampleLimit</code> will be set to <code>enforcedSampleLimit</code>.
* Scrape objects with a sampleLimit value less than or equal to enforcedSampleLimit keep their specific value.
* Scrape objects with a sampleLimit value greater than enforcedSampleLimit are set to enforcedSampleLimit.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedTargetLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>When defined, enforcedTargetLimit specifies a global limit on the number
of scraped targets. The value overrides any <code>spec.targetLimit</code> set by
ServiceMonitor, PodMonitor, Probe objects unless <code>spec.targetLimit</code> is
greater than zero and less than <code>spec.enforcedTargetLimit</code>.</p>
<p>It is meant to be used by admins to to keep the overall number of
targets under a desired limit.</p>
<p>When both <code>enforcedTargetLimit</code> and <code>targetLimit</code> are defined and greater than zero, the following rules apply:
* Scrape objects without a defined targetLimit value will inherit the global targetLimit value (Prometheus &gt;= 2.45.0) or the enforcedTargetLimit value (Prometheus &lt; v2.45.0).
If Prometheus version is &gt;= 2.45.0 and the <code>enforcedTargetLimit</code> is greater than the <code>targetLimit</code>, the <code>targetLimit</code> will be set to <code>enforcedTargetLimit</code>.
* Scrape objects with a targetLimit value less than or equal to enforcedTargetLimit keep their specific value.
* Scrape objects with a targetLimit value greater than enforcedTargetLimit are set to enforcedTargetLimit.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedLabelLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>When defined, enforcedLabelLimit specifies a global limit on the number
of labels per sample. The value overrides any <code>spec.labelLimit</code> set by
ServiceMonitor, PodMonitor, Probe objects unless <code>spec.labelLimit</code> is
greater than zero and less than <code>spec.enforcedLabelLimit</code>.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
<p>When both <code>enforcedLabelLimit</code> and <code>labelLimit</code> are defined and greater than zero, the following rules apply:
* Scrape objects without a defined labelLimit value will inherit the global labelLimit value (Prometheus &gt;= 2.45.0) or the enforcedLabelLimit value (Prometheus &lt; v2.45.0).
If Prometheus version is &gt;= 2.45.0 and the <code>enforcedLabelLimit</code> is greater than the <code>labelLimit</code>, the <code>labelLimit</code> will be set to <code>enforcedLabelLimit</code>.
* Scrape objects with a labelLimit value less than or equal to enforcedLabelLimit keep their specific value.
* Scrape objects with a labelLimit value greater than enforcedLabelLimit are set to enforcedLabelLimit.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedLabelNameLengthLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>When defined, enforcedLabelNameLengthLimit specifies a global limit on the length
of labels name per sample. The value overrides any <code>spec.labelNameLengthLimit</code> set by
ServiceMonitor, PodMonitor, Probe objects unless <code>spec.labelNameLengthLimit</code> is
greater than zero and less than <code>spec.enforcedLabelNameLengthLimit</code>.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
<p>When both <code>enforcedLabelNameLengthLimit</code> and <code>labelNameLengthLimit</code> are defined and greater than zero, the following rules apply:
* Scrape objects without a defined labelNameLengthLimit value will inherit the global labelNameLengthLimit value (Prometheus &gt;= 2.45.0) or the enforcedLabelNameLengthLimit value (Prometheus &lt; v2.45.0).
If Prometheus version is &gt;= 2.45.0 and the <code>enforcedLabelNameLengthLimit</code> is greater than the <code>labelNameLengthLimit</code>, the <code>labelNameLengthLimit</code> will be set to <code>enforcedLabelNameLengthLimit</code>.
* Scrape objects with a labelNameLengthLimit value less than or equal to enforcedLabelNameLengthLimit keep their specific value.
* Scrape objects with a labelNameLengthLimit value greater than enforcedLabelNameLengthLimit are set to enforcedLabelNameLengthLimit.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedLabelValueLengthLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>When not null, enforcedLabelValueLengthLimit defines a global limit on the length
of labels value per sample. The value overrides any <code>spec.labelValueLengthLimit</code> set by
ServiceMonitor, PodMonitor, Probe objects unless <code>spec.labelValueLengthLimit</code> is
greater than zero and less than <code>spec.enforcedLabelValueLengthLimit</code>.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
<p>When both <code>enforcedLabelValueLengthLimit</code> and <code>labelValueLengthLimit</code> are defined and greater than zero, the following rules apply:
* Scrape objects without a defined labelValueLengthLimit value will inherit the global labelValueLengthLimit value (Prometheus &gt;= 2.45.0) or the enforcedLabelValueLengthLimit value (Prometheus &lt; v2.45.0).
If Prometheus version is &gt;= 2.45.0 and the <code>enforcedLabelValueLengthLimit</code> is greater than the <code>labelValueLengthLimit</code>, the <code>labelValueLengthLimit</code> will be set to <code>enforcedLabelValueLengthLimit</code>.
* Scrape objects with a labelValueLengthLimit value less than or equal to enforcedLabelValueLengthLimit keep their specific value.
* Scrape objects with a labelValueLengthLimit value greater than enforcedLabelValueLengthLimit are set to enforcedLabelValueLengthLimit.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedKeepDroppedTargets</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>When defined, enforcedKeepDroppedTargets specifies a global limit on the number of targets
dropped by relabeling that will be kept in memory. The value overrides
any <code>spec.keepDroppedTargets</code> set by
ServiceMonitor, PodMonitor, Probe objects unless <code>spec.keepDroppedTargets</code> is
greater than zero and less than <code>spec.enforcedKeepDroppedTargets</code>.</p>
<p>It requires Prometheus &gt;= v2.47.0.</p>
<p>When both <code>enforcedKeepDroppedTargets</code> and <code>keepDroppedTargets</code> are defined and greater than zero, the following rules apply:
* Scrape objects without a defined keepDroppedTargets value will inherit the global keepDroppedTargets value (Prometheus &gt;= 2.45.0) or the enforcedKeepDroppedTargets value (Prometheus &lt; v2.45.0).
If Prometheus version is &gt;= 2.45.0 and the <code>enforcedKeepDroppedTargets</code> is greater than the <code>keepDroppedTargets</code>, the <code>keepDroppedTargets</code> will be set to <code>enforcedKeepDroppedTargets</code>.
* Scrape objects with a keepDroppedTargets value less than or equal to enforcedKeepDroppedTargets keep their specific value.
* Scrape objects with a keepDroppedTargets value greater than enforcedKeepDroppedTargets are set to enforcedKeepDroppedTargets.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedBodySizeLimit</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ByteSize">
Monitoring v1.ByteSize
</a>
</em>
</td>
<td>
<p>When defined, enforcedBodySizeLimit specifies a global limit on the size
of uncompressed response body that will be accepted by Prometheus.
Targets responding with a body larger than this many bytes will cause
the scrape to fail.</p>
<p>It requires Prometheus &gt;= v2.28.0.</p>
<p>When both <code>enforcedBodySizeLimit</code> and <code>bodySizeLimit</code> are defined and greater than zero, the following rules apply:
* Scrape objects without a defined bodySizeLimit value will inherit the global bodySizeLimit value (Prometheus &gt;= 2.45.0) or the enforcedBodySizeLimit value (Prometheus &lt; v2.45.0).
If Prometheus version is &gt;= 2.45.0 and the <code>enforcedBodySizeLimit</code> is greater than the <code>bodySizeLimit</code>, the <code>bodySizeLimit</code> will be set to <code>enforcedBodySizeLimit</code>.
* Scrape objects with a bodySizeLimit value less than or equal to enforcedBodySizeLimit keep their specific value.
* Scrape objects with a bodySizeLimit value greater than enforcedBodySizeLimit are set to enforcedBodySizeLimit.</p>
</td>
</tr>
<tr>
<td>
<code>nameValidationScheme</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.NameValidationSchemeOptions">
Monitoring v1.NameValidationSchemeOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the validation scheme for metric and label names.</p>
<p>It requires Prometheus &gt;= v2.55.0.</p>
</td>
</tr>
<tr>
<td>
<code>nameEscapingScheme</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.NameEscapingSchemeOptions">
Monitoring v1.NameEscapingSchemeOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the character escaping scheme that will be requested when scraping
for metric and label names that do not conform to the legacy Prometheus
character set.</p>
<p>It requires Prometheus &gt;= v3.4.0.</p>
</td>
</tr>
<tr>
<td>
<code>convertClassicHistogramsToNHCB</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to convert all scraped classic histograms into a native
histogram with custom buckets.</p>
<p>It requires Prometheus &gt;= v3.4.0.</p>
</td>
</tr>
<tr>
<td>
<code>minReadySeconds</code><br/>
<em>
uint32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Minimum number of seconds for which a newly created Pod should be ready
without any of its container crashing for it to be considered available.
Defaults to 0 (pod will be considered available as soon as it is ready)</p>
<p>This is an alpha field from kubernetes 1.22 until 1.24 which requires
enabling the StatefulSetMinReadySeconds feature gate.</p>
</td>
</tr>
<tr>
<td>
<code>hostAliases</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.HostAlias">
[]Monitoring v1.HostAlias
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional list of hosts and IPs that will be injected into the Pod&rsquo;s
hosts file if specified.</p>
</td>
</tr>
<tr>
<td>
<code>additionalArgs</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Argument">
[]Monitoring v1.Argument
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AdditionalArgs allows setting additional arguments for the &lsquo;prometheus&rsquo; container.</p>
<p>It is intended for e.g. activating hidden flags which are not supported by
the dedicated configuration options yet. The arguments are passed as-is to the
Prometheus container which may cause issues if they are invalid or not supported
by the given Prometheus version.</p>
<p>In case of an argument conflict (e.g. an argument which is already set by the
operator itself) or when providing an invalid argument, the reconciliation will
fail and an error will be logged.</p>
</td>
</tr>
<tr>
<td>
<code>walCompression</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures compression of the write-ahead log (WAL) using Snappy.</p>
<p>WAL compression is enabled by default for Prometheus &gt;= 2.20.0</p>
<p>Requires Prometheus v2.11.0 and above.</p>
</td>
</tr>
<tr>
<td>
<code>excludedFromEnforcement</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ObjectReference">
[]Monitoring v1.ObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of references to PodMonitor, ServiceMonitor, Probe and PrometheusRule objects
to be excluded from enforcing a namespace label of origin.</p>
<p>It is only applicable if <code>spec.enforcedNamespaceLabel</code> set to true.</p>
</td>
</tr>
<tr>
<td>
<code>hostNetwork</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Use the host&rsquo;s network namespace if true.</p>
<p>Make sure to understand the security implications if you want to enable
it (<a href="https://kubernetes.io/docs/concepts/configuration/overview/">https://kubernetes.io/docs/concepts/configuration/overview/</a> ).</p>
<p>When hostNetwork is enabled, this will set the DNS policy to
<code>ClusterFirstWithHostNet</code> automatically (unless <code>.spec.DNSPolicy</code> is set
to a different value).</p>
</td>
</tr>
<tr>
<td>
<code>podTargetLabels</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodTargetLabels are appended to the <code>spec.podTargetLabels</code> field of all
PodMonitor and ServiceMonitor objects.</p>
</td>
</tr>
<tr>
<td>
<code>tracingConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.PrometheusTracingConfig">
Monitoring v1.PrometheusTracingConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TracingConfig configures tracing in Prometheus.</p>
<p>This is an <em>experimental feature</em>, it may change in any upcoming release
in a breaking way.</p>
</td>
</tr>
<tr>
<td>
<code>bodySizeLimit</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ByteSize">
Monitoring v1.ByteSize
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>BodySizeLimit defines per-scrape on response body size.
Only valid in Prometheus versions 2.45.0 and newer.</p>
<p>Note that the global limit only applies to scrape objects that don&rsquo;t specify an explicit limit value.
If you want to enforce a maximum limit for all scrape objects, refer to enforcedBodySizeLimit.</p>
</td>
</tr>
<tr>
<td>
<code>sampleLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.
Only valid in Prometheus versions 2.45.0 and newer.</p>
<p>Note that the global limit only applies to scrape objects that don&rsquo;t specify an explicit limit value.
If you want to enforce a maximum limit for all scrape objects, refer to enforcedSampleLimit.</p>
</td>
</tr>
<tr>
<td>
<code>targetLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>TargetLimit defines a limit on the number of scraped targets that will be accepted.
Only valid in Prometheus versions 2.45.0 and newer.</p>
<p>Note that the global limit only applies to scrape objects that don&rsquo;t specify an explicit limit value.
If you want to enforce a maximum limit for all scrape objects, refer to enforcedTargetLimit.</p>
</td>
</tr>
<tr>
<td>
<code>labelLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on number of labels that will be accepted for a sample.
Only valid in Prometheus versions 2.45.0 and newer.</p>
<p>Note that the global limit only applies to scrape objects that don&rsquo;t specify an explicit limit value.
If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelLimit.</p>
</td>
</tr>
<tr>
<td>
<code>labelNameLengthLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on length of labels name that will be accepted for a sample.
Only valid in Prometheus versions 2.45.0 and newer.</p>
<p>Note that the global limit only applies to scrape objects that don&rsquo;t specify an explicit limit value.
If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelNameLengthLimit.</p>
</td>
</tr>
<tr>
<td>
<code>labelValueLengthLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on length of labels value that will be accepted for a sample.
Only valid in Prometheus versions 2.45.0 and newer.</p>
<p>Note that the global limit only applies to scrape objects that don&rsquo;t specify an explicit limit value.
If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelValueLengthLimit.</p>
</td>
</tr>
<tr>
<td>
<code>keepDroppedTargets</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on the number of targets dropped by relabeling
that will be kept in memory. 0 means no limit.</p>
<p>It requires Prometheus &gt;= v2.47.0.</p>
<p>Note that the global limit only applies to scrape objects that don&rsquo;t specify an explicit limit value.
If you want to enforce a maximum limit for all scrape objects, refer to enforcedKeepDroppedTargets.</p>
</td>
</tr>
<tr>
<td>
<code>reloadStrategy</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ReloadStrategyType">
Monitoring v1.ReloadStrategyType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the strategy used to reload the Prometheus configuration.
If not specified, the configuration is reloaded using the /-/reload HTTP endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>maximumStartupDurationSeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the maximum time that the <code>prometheus</code> container&rsquo;s startup probe will wait before being considered failed. The startup probe will return success after the WAL replay is complete.
If set, the value should be greater than 60 (seconds). Otherwise it will be equal to 600 seconds (15 minutes).</p>
</td>
</tr>
<tr>
<td>
<code>scrapeClasses</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ScrapeClass">
[]Monitoring v1.ScrapeClass
</a>
</em>
</td>
<td>
<p>List of scrape classes to expose to scraping objects such as
PodMonitors, ServiceMonitors, Probes and ScrapeConfigs.</p>
<p>This is an <em>experimental feature</em>, it may change in any upcoming release
in a breaking way.</p>
</td>
</tr>
<tr>
<td>
<code>serviceDiscoveryRole</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ServiceDiscoveryRole">
Monitoring v1.ServiceDiscoveryRole
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the service discovery role used to discover targets from
<code>ServiceMonitor</code> objects and Alertmanager endpoints.</p>
<p>If set, the value should be either &ldquo;Endpoints&rdquo; or &ldquo;EndpointSlice&rdquo;.
If unset, the operator assumes the &ldquo;Endpoints&rdquo; role.</p>
</td>
</tr>
<tr>
<td>
<code>tsdb</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.TSDBSpec">
Monitoring v1.TSDBSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the runtime reloadable configuration of the timeseries database(TSDB).
It requires Prometheus &gt;= v2.39.0 or PrometheusAgent &gt;= v2.54.0.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeFailureLogFile</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>File to which scrape failures are logged.
Reloading the configuration will reopen the file.</p>
<p>If the filename has an empty path, e.g. &lsquo;file.log&rsquo;, The Prometheus Pods
will mount the file into an emptyDir volume at <code>/var/log/prometheus</code>.
If a full path is provided, e.g. &lsquo;/var/log/prometheus/file.log&rsquo;, you
must mount a volume in the specified directory and it must be writable.
It requires Prometheus &gt;= v2.55.0.</p>
</td>
</tr>
<tr>
<td>
<code>serviceName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The name of the service name used by the underlying StatefulSet(s) as the governing service.
If defined, the Service  must be created before the Prometheus/PrometheusAgent resource in the same namespace and it must define a selector that matches the pod labels.
If empty, the operator will create and manage a headless service named <code>prometheus-operated</code> for Prometheus resources,
or <code>prometheus-agent-operated</code> for PrometheusAgent resources.
When deploying multiple Prometheus/PrometheusAgent resources in the same namespace, it is recommended to specify a different value for each.
See <a href="https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id">https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id</a> for more details.</p>
</td>
</tr>
<tr>
<td>
<code>runtime</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.RuntimeConfig">
Monitoring v1.RuntimeConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RuntimeConfig configures the values for the Prometheus process behavior</p>
</td>
</tr>
<tr>
<td>
<code>terminationGracePeriodSeconds</code><br/>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional duration in seconds the pod needs to terminate gracefully.
Value must be non-negative integer. The value zero indicates stop immediately via
the kill signal (no opportunity to shut down) which may lead to data corruption.</p>
<p>Defaults to 600 seconds.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.PrometheusStatus">
Monitoring v1.PrometheusStatus
</a>
</em>
</td>
<td>
<p>Most recent observed status of the Prometheus cluster. Read-only.
More info:
<a href="https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status">https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.ScrapeConfig">ScrapeConfig
</h3>
<div>
<p>ScrapeConfig defines a namespaced Prometheus scrape_config to be aggregated across
multiple namespaces into the Prometheus configuration.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code><br/>
string</td>
<td>
<code>
monitoring.coreos.com/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>ScrapeConfig</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">
ScrapeConfigSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>jobName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The value of the <code>job</code> label assigned to the scraped metrics by default.</p>
<p>The <code>job_name</code> field in the rendered scrape configuration is always controlled by the
operator to prevent duplicate job names, which Prometheus does not allow. Instead the
<code>job</code> label is set by means of relabeling configs.</p>
</td>
</tr>
<tr>
<td>
<code>staticConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.StaticConfig">
[]StaticConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StaticConfigs defines a list of static targets with a common label set.</p>
</td>
</tr>
<tr>
<td>
<code>fileSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.FileSDConfig">
[]FileSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>FileSDConfigs defines a list of file service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>httpSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPSDConfig">
[]HTTPSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTPSDConfigs defines a list of HTTP service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>kubernetesSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.KubernetesSDConfig">
[]KubernetesSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>KubernetesSDConfigs defines a list of Kubernetes service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>consulSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.ConsulSDConfig">
[]ConsulSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ConsulSDConfigs defines a list of Consul service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>dnsSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.DNSSDConfig">
[]DNSSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DNSSDConfigs defines a list of DNS service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>ec2SDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.EC2SDConfig">
[]EC2SDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>EC2SDConfigs defines a list of EC2 service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>azureSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.AzureSDConfig">
[]AzureSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AzureSDConfigs defines a list of Azure service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>gceSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.GCESDConfig">
[]GCESDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>GCESDConfigs defines a list of GCE service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>openstackSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.OpenStackSDConfig">
[]OpenStackSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>OpenStackSDConfigs defines a list of OpenStack service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>digitalOceanSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.DigitalOceanSDConfig">
[]DigitalOceanSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DigitalOceanSDConfigs defines a list of DigitalOcean service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>kumaSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.KumaSDConfig">
[]KumaSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>KumaSDConfigs defines a list of Kuma service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>eurekaSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.EurekaSDConfig">
[]EurekaSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>EurekaSDConfigs defines a list of Eureka service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>dockerSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.DockerSDConfig">
[]DockerSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DockerSDConfigs defines a list of Docker service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>linodeSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.LinodeSDConfig">
[]LinodeSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>LinodeSDConfigs defines a list of Linode service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>hetznerSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HetznerSDConfig">
[]HetznerSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HetznerSDConfigs defines a list of Hetzner service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>nomadSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.NomadSDConfig">
[]NomadSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NomadSDConfigs defines a list of Nomad service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>dockerSwarmSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.DockerSwarmSDConfig">
[]DockerSwarmSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DockerswarmSDConfigs defines a list of Dockerswarm service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>puppetDBSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.PuppetDBSDConfig">
[]PuppetDBSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PuppetDBSDConfigs defines a list of PuppetDB service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>lightSailSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.LightSailSDConfig">
[]LightSailSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>LightsailSDConfigs defines a list of Lightsail service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>ovhcloudSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.OVHCloudSDConfig">
[]OVHCloudSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>OVHCloudSDConfigs defines a list of OVHcloud service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>scalewaySDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.ScalewaySDConfig">
[]ScalewaySDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ScalewaySDConfigs defines a list of Scaleway instances and baremetal service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>ionosSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.IonosSDConfig">
[]IonosSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>IonosSDConfigs defines a list of IONOS service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>relabelings</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.RelabelConfig">
[]Monitoring v1.RelabelConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RelabelConfigs defines how to rewrite the target&rsquo;s labels before scraping.
Prometheus Operator automatically adds relabelings for a few standard Kubernetes fields.
The original scrape job&rsquo;s name is available via the <code>__tmp_prometheus_job_name</code> label.
More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config</a></p>
</td>
</tr>
<tr>
<td>
<code>metricsPath</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>MetricsPath HTTP path to scrape for metrics. If empty, Prometheus uses the default value (e.g. /metrics).</p>
</td>
</tr>
<tr>
<td>
<code>scrapeInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ScrapeInterval is the interval between consecutive scrapes.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeTimeout</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ScrapeTimeout is the number of seconds to wait until a scrape request times out.
The value cannot be greater than the scrape interval otherwise the operator will reject the resource.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeProtocols</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ScrapeProtocol">
[]Monitoring v1.ScrapeProtocol
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The protocols to negotiate during a scrape. It tells clients the
protocols supported by Prometheus in order of preference (from most to least preferred).</p>
<p>If unset, Prometheus uses its default value.</p>
<p>It requires Prometheus &gt;= v2.49.0.</p>
</td>
</tr>
<tr>
<td>
<code>fallbackScrapeProtocol</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ScrapeProtocol">
Monitoring v1.ScrapeProtocol
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The protocol to use if a scrape returns blank, unparseable, or otherwise invalid Content-Type.</p>
<p>It requires Prometheus &gt;= v3.0.0.</p>
</td>
</tr>
<tr>
<td>
<code>honorTimestamps</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data.</p>
</td>
</tr>
<tr>
<td>
<code>trackTimestampsStaleness</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>TrackTimestampsStaleness whether Prometheus tracks staleness of
the metrics that have an explicit timestamp present in scraped data.
Has no effect if <code>honorTimestamps</code> is false.
It requires Prometheus &gt;= v2.48.0.</p>
</td>
</tr>
<tr>
<td>
<code>honorLabels</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>HonorLabels chooses the metric&rsquo;s labels on collisions with target labels.</p>
</td>
</tr>
<tr>
<td>
<code>params</code><br/>
<em>
map[string][]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional HTTP URL parameters</p>
</td>
</tr>
<tr>
<td>
<code>scheme</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures the protocol scheme used for requests.
If empty, Prometheus uses HTTP by default.</p>
</td>
</tr>
<tr>
<td>
<code>enableCompression</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>When false, Prometheus will request uncompressed response from the scraped target.</p>
<p>It requires Prometheus &gt;= v2.49.0.</p>
<p>If unset, Prometheus uses true by default.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>BasicAuth information to use on every scrape request.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization header to use on every scrape request.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>OAuth2 configuration to use on every scrape request.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration to use on every scrape request</p>
</td>
</tr>
<tr>
<td>
<code>sampleLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.</p>
</td>
</tr>
<tr>
<td>
<code>targetLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>TargetLimit defines a limit on the number of scraped targets that will be accepted.</p>
</td>
</tr>
<tr>
<td>
<code>labelLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on number of labels that will be accepted for a sample.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>labelNameLengthLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on length of labels name that will be accepted for a sample.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>labelValueLengthLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on length of labels value that will be accepted for a sample.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeClassicHistograms</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to scrape a classic histogram that is also exposed as a native histogram.
It requires Prometheus &gt;= v2.45.0.</p>
</td>
</tr>
<tr>
<td>
<code>nativeHistogramBucketLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>If there are more than this many buckets in a native histogram,
buckets will be merged to stay within the limit.
It requires Prometheus &gt;= v2.45.0.</p>
</td>
</tr>
<tr>
<td>
<code>nativeHistogramMinBucketFactor</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/api/resource#Quantity">
k8s.io/apimachinery/pkg/api/resource.Quantity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If the growth factor of one bucket to the next is smaller than this,
buckets will be merged to increase the factor sufficiently.
It requires Prometheus &gt;= v2.50.0.</p>
</td>
</tr>
<tr>
<td>
<code>convertClassicHistogramsToNHCB</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to convert all scraped classic histograms into a native histogram with custom buckets.
It requires Prometheus &gt;= v3.0.0.</p>
</td>
</tr>
<tr>
<td>
<code>keepDroppedTargets</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on the number of targets dropped by relabeling
that will be kept in memory. 0 means no limit.</p>
<p>It requires Prometheus &gt;= v2.47.0.</p>
</td>
</tr>
<tr>
<td>
<code>metricRelabelings</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.RelabelConfig">
[]Monitoring v1.RelabelConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>MetricRelabelConfigs to apply to samples before ingestion.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>nameValidationScheme</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.NameValidationSchemeOptions">
Monitoring v1.NameValidationSchemeOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the validation scheme for metric and label names.</p>
<p>It requires Prometheus &gt;= v3.0.0.</p>
</td>
</tr>
<tr>
<td>
<code>nameEscapingScheme</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.NameEscapingSchemeOptions">
Monitoring v1.NameEscapingSchemeOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Metric name escaping mode to request through content negotiation.</p>
<p>It requires Prometheus &gt;= v3.4.0.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeClass</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The scrape class to apply.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.AlertmanagerConfigSpec">AlertmanagerConfigSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.AlertmanagerConfig">AlertmanagerConfig</a>)
</p>
<div>
<p>AlertmanagerConfigSpec is a specification of the desired behavior of the
Alertmanager configuration.
By default, the Alertmanager configuration only applies to alerts for which
the <code>namespace</code> label is equal to the namespace of the AlertmanagerConfig
resource (see the <code>.spec.alertmanagerConfigMatcherStrategy</code> field of the
Alertmanager CRD).</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>route</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.Route">
Route
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The Alertmanager route definition for alerts matching the resource&rsquo;s
namespace. If present, it will be added to the generated Alertmanager
configuration as a first-level route.</p>
</td>
</tr>
<tr>
<td>
<code>receivers</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.Receiver">
[]Receiver
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of receivers.</p>
</td>
</tr>
<tr>
<td>
<code>inhibitRules</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.InhibitRule">
[]InhibitRule
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of inhibition rules. The rules will only apply to alerts matching
the resource&rsquo;s namespace.</p>
</td>
</tr>
<tr>
<td>
<code>muteTimeIntervals</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.MuteTimeInterval">
[]MuteTimeInterval
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of MuteTimeInterval specifying when the routes should be muted.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.AttachMetadata">AttachMetadata
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.KubernetesSDConfig">KubernetesSDConfig</a>)
</p>
<div>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>node</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Attaches node metadata to discovered targets.
When set to true, Prometheus must have the <code>get</code> permission on the
<code>Nodes</code> objects.
Only valid for Pod, Endpoint and Endpointslice roles.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.AuthenticationMethodType">AuthenticationMethodType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.AzureSDConfig">AzureSDConfig</a>)
</p>
<div>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;ManagedIdentity&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;OAuth&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;SDK&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.AzureSDConfig">AzureSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>AzureSDConfig allow retrieving scrape targets from Azure VMs.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#azure_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#azure_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>environment</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The Azure environment.</p>
</td>
</tr>
<tr>
<td>
<code>authenticationMethod</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.AuthenticationMethodType">
AuthenticationMethodType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<h1>The authentication method, either <code>OAuth</code> or <code>ManagedIdentity</code> or <code>SDK</code>.</h1>
<p>See <a href="https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview">https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview</a>
SDK authentication method uses environment variables by default.
See <a href="https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication">https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication</a></p>
</td>
</tr>
<tr>
<td>
<code>subscriptionID</code><br/>
<em>
string
</em>
</td>
<td>
<p>The subscription ID. Always required.</p>
</td>
</tr>
<tr>
<td>
<code>tenantID</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional tenant ID. Only required with the OAuth authentication method.</p>
</td>
</tr>
<tr>
<td>
<code>clientID</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional client ID. Only required with the OAuth authentication method.</p>
</td>
</tr>
<tr>
<td>
<code>clientSecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional client secret. Only required with the OAuth authentication method.</p>
</td>
</tr>
<tr>
<td>
<code>resourceGroup</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional resource group name. Limits discovery to this resource group.
Requires  Prometheus v2.35.0 and above</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RefreshInterval configures the refresh interval at which Prometheus will re-read the instance list.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The port to scrape metrics from. If using the public IP address, this must
instead be specified in the relabeling rule.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>BasicAuth information to authenticate against the target HTTP endpoint.
More info: <a href="https://prometheus.io/docs/operating/configuration/#endpoints">https://prometheus.io/docs/operating/configuration/#endpoints</a>
Cannot be set at the same time as <code>authorization</code>, or <code>oAuth2</code>.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization header configuration to authenticate against the target HTTP endpoint.
Cannot be set at the same time as <code>oAuth2</code>, or <code>basicAuth</code>.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional OAuth 2.0 configuration to authenticate against the target HTTP endpoint.
Cannot be set at the same time as <code>authorization</code>, or <code>basicAuth</code>.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration applying to the target HTTP endpoint.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.ConsulSDConfig">ConsulSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>ConsulSDConfig defines a Consul service discovery configuration
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#consul_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#consul_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>server</code><br/>
<em>
string
</em>
</td>
<td>
<p>Consul server address. A valid string consisting of a hostname or IP followed by an optional port number.</p>
</td>
</tr>
<tr>
<td>
<code>pathPrefix</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Prefix for URIs for when consul is behind an API gateway (reverse proxy).</p>
<p>It requires Prometheus &gt;= 2.45.0.</p>
</td>
</tr>
<tr>
<td>
<code>tokenRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Consul ACL TokenRef, if not provided it will use the ACL from the local Consul Agent.</p>
</td>
</tr>
<tr>
<td>
<code>datacenter</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Consul Datacenter name, if not provided it will use the local Consul Agent Datacenter.</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespaces are only supported in Consul Enterprise.</p>
<p>It requires Prometheus &gt;= 2.28.0.</p>
</td>
</tr>
<tr>
<td>
<code>partition</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Admin Partitions are only supported in Consul Enterprise.</p>
</td>
</tr>
<tr>
<td>
<code>scheme</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTP Scheme default &ldquo;http&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>services</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>A list of services for which targets are retrieved. If omitted, all services are scraped.</p>
</td>
</tr>
<tr>
<td>
<code>tags</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>An optional list of tags used to filter nodes for a given service. Services must contain all tags in the list.
Starting with Consul 1.14, it is recommended to use <code>filter</code> with the <code>ServiceTags</code> selector instead.</p>
</td>
</tr>
<tr>
<td>
<code>tagSeparator</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The string by which Consul tags are joined into the tag label.
If unset, Prometheus uses its default value.</p>
</td>
</tr>
<tr>
<td>
<code>nodeMeta</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Node metadata key/value pairs to filter nodes for a given service.
Starting with Consul 1.14, it is recommended to use <code>filter</code> with the <code>NodeMeta</code> selector instead.</p>
</td>
</tr>
<tr>
<td>
<code>filter</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Filter expression used to filter the catalog results.
See <a href="https://www.consul.io/api-docs/catalog#list-services">https://www.consul.io/api-docs/catalog#list-services</a>
It requires Prometheus &gt;= 3.0.0.</p>
</td>
</tr>
<tr>
<td>
<code>allowStale</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Allow stale Consul results (see <a href="https://www.consul.io/api/features/consistency.html">https://www.consul.io/api/features/consistency.html</a>). Will reduce load on Consul.
If unset, Prometheus uses its default value.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The time after which the provided names are refreshed.
On large setup it might be a good idea to increase this value because the catalog will change all the time.
If unset, Prometheus uses its default value.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional BasicAuth information to authenticate against the Consul Server.
More info: <a href="https://prometheus.io/docs/operating/configuration/#endpoints">https://prometheus.io/docs/operating/configuration/#endpoints</a>
Cannot be set at the same time as <code>authorization</code>, or <code>oauth2</code>.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional Authorization header configuration to authenticate against the Consul Server.
Cannot be set at the same time as <code>basicAuth</code>, or <code>oauth2</code>.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional OAuth2.0 configuration.
Cannot be set at the same time as <code>basicAuth</code>, or <code>authorization</code>.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.
If unset, Prometheus uses its default value.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.
If unset, Prometheus uses its default value.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration to connect to the Consul API.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.DNSRecordType">DNSRecordType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.DNSSDConfig">DNSSDConfig</a>)
</p>
<div>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;A&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;AAAA&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;MX&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;NS&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;SRV&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.DNSSDConfig">DNSSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>DNSSDConfig allows specifying a set of DNS domain names which are periodically queried to discover a list of targets.
The DNS servers to be contacted are read from /etc/resolv.conf.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#dns_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#dns_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>names</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>A list of DNS domain names to be queried.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RefreshInterval configures the time after which the provided names are refreshed.
If not set, Prometheus uses its default value.</p>
</td>
</tr>
<tr>
<td>
<code>type</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.DNSRecordType">
DNSRecordType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The type of DNS query to perform. One of SRV, A, AAAA, MX or NS.
If not set, Prometheus uses its default value.</p>
<p>When set to NS, it requires Prometheus &gt;= v2.49.0.
When set to MX, it requires Prometheus &gt;= v2.38.0</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The port number used if the query type is not SRV
Ignored for SRV records</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.DayOfMonthRange">DayOfMonthRange
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.TimeInterval">TimeInterval</a>)
</p>
<div>
<p>DayOfMonthRange is an inclusive range of days of the month beginning at 1</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>start</code><br/>
<em>
int
</em>
</td>
<td>
<p>Start of the inclusive range</p>
</td>
</tr>
<tr>
<td>
<code>end</code><br/>
<em>
int
</em>
</td>
<td>
<p>End of the inclusive range</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.DigitalOceanSDConfig">DigitalOceanSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>DigitalOceanSDConfig allow retrieving scrape targets from DigitalOcean&rsquo;s Droplets API.
This service discovery uses the public IPv4 address by default, by that can be changed with relabeling
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#digitalocean_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#digitalocean_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization header configuration to authenticate against the DigitalOcean API.
Cannot be set at the same time as <code>oauth2</code>.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional OAuth 2.0 configuration.
Cannot be set at the same time as <code>authorization</code>.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration applying to the target HTTP endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The port to scrape metrics from.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Refresh interval to re-read the instance list.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.DiscordConfig">DiscordConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Receiver">Receiver</a>)
</p>
<div>
<p>DiscordConfig configures notifications via Discord.
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#discord_config">https://prometheus.io/docs/alerting/latest/configuration/#discord_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>sendResolved</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether or not to notify about resolved alerts.</p>
</td>
</tr>
<tr>
<td>
<code>apiURL</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>The secret&rsquo;s key that contains the Discord webhook URL.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>title</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The template of the message&rsquo;s title.</p>
</td>
</tr>
<tr>
<td>
<code>message</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The template of the message&rsquo;s body.</p>
</td>
</tr>
<tr>
<td>
<code>content</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The template of the content&rsquo;s body.</p>
</td>
</tr>
<tr>
<td>
<code>username</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The username of the message sender.</p>
</td>
</tr>
<tr>
<td>
<code>avatarURL</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.URL">
URL
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The avatar url of the message sender.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">
HTTPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTP client configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.DockerSDConfig">DockerSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>Docker SD configurations allow retrieving scrape targets from Docker Engine hosts.
This SD discovers &ldquo;containers&rdquo; and will create a target for each network IP and
port the container is configured to expose.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#docker_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#docker_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>host</code><br/>
<em>
string
</em>
</td>
<td>
<p>Address of the docker daemon</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration applying to the target HTTP endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>The port to scrape metrics from.</p>
</td>
</tr>
<tr>
<td>
<code>hostNetworkingHost</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The host to use if the container is in host networking mode.</p>
</td>
</tr>
<tr>
<td>
<code>matchFirstNetwork</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether to match the first network if the container has multiple networks defined.
If unset, Prometheus uses true by default.
It requires Prometheus &gt;= v2.54.1.</p>
</td>
</tr>
<tr>
<td>
<code>filters</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.Filters">
Filters
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional filters to limit the discovery process to a subset of the available resources.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Time after which the container is refreshed.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>BasicAuth information to use on every scrape request.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization header configuration to authenticate against the Docker API.
Cannot be set at the same time as <code>oauth2</code>.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional OAuth 2.0 configuration.
Cannot be set at the same time as <code>authorization</code>.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.DockerSwarmSDConfig">DockerSwarmSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>DockerSwarmSDConfig configurations allow retrieving scrape targets from Docker Swarm engine.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#dockerswarm_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#dockerswarm_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>host</code><br/>
<em>
string
</em>
</td>
<td>
<p>Address of the Docker daemon</p>
</td>
</tr>
<tr>
<td>
<code>role</code><br/>
<em>
string
</em>
</td>
<td>
<p>Role of the targets to retrieve. Must be <code>Services</code>, <code>Tasks</code>, or <code>Nodes</code>.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The port to scrape metrics from, when <code>role</code> is nodes, and for discovered
tasks and services that don&rsquo;t have published ports.</p>
</td>
</tr>
<tr>
<td>
<code>filters</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.Filters">
Filters
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional filters to limit the discovery process to a subset of available
resources.
The available filters are listed in the upstream documentation:
Services: <a href="https://docs.docker.com/engine/api/v1.40/#operation/ServiceList">https://docs.docker.com/engine/api/v1.40/#operation/ServiceList</a>
Tasks: <a href="https://docs.docker.com/engine/api/v1.40/#operation/TaskList">https://docs.docker.com/engine/api/v1.40/#operation/TaskList</a>
Nodes: <a href="https://docs.docker.com/engine/api/v1.40/#operation/NodeList">https://docs.docker.com/engine/api/v1.40/#operation/NodeList</a></p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The time after which the service discovery data is refreshed.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional HTTP basic authentication information.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization header configuration to authenticate against the target HTTP endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional OAuth 2.0 configuration.
Cannot be set at the same time as <code>authorization</code>, or <code>basicAuth</code>.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration to use on every scrape request</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.EC2SDConfig">EC2SDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>EC2SDConfig allow retrieving scrape targets from AWS EC2 instances.
The private IP address is used by default, but may be changed to the public IP address with relabeling.
The IAM credentials used must have the ec2:DescribeInstances permission to discover scrape targets
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#ec2_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#ec2_sd_config</a></p>
<p>The EC2 service discovery requires AWS API keys or role ARN for authentication.
BasicAuth, Authorization and OAuth2 fields are not present on purpose.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>region</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The AWS region.</p>
</td>
</tr>
<tr>
<td>
<code>accessKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AccessKey is the AWS API key.</p>
</td>
</tr>
<tr>
<td>
<code>secretKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecretKey is the AWS API secret.</p>
</td>
</tr>
<tr>
<td>
<code>roleARN</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>AWS Role ARN, an alternative to using AWS API keys.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The port to scrape metrics from. If using the public IP address, this must
instead be specified in the relabeling rule.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RefreshInterval configures the refresh interval at which Prometheus will re-read the instance list.</p>
</td>
</tr>
<tr>
<td>
<code>filters</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.Filters">
Filters
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Filters can be used optionally to filter the instance list by other criteria.
Available filter criteria can be found here:
<a href="https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html">https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html</a>
Filter API documentation: <a href="https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Filter.html">https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Filter.html</a>
It requires Prometheus &gt;= v2.3.0</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration to connect to the AWS EC2 API.
It requires Prometheus &gt;= v2.41.0</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.
It requires Prometheus &gt;= v2.41.0</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.
It requires Prometheus &gt;= v2.41.0</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.EmailConfig">EmailConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Receiver">Receiver</a>)
</p>
<div>
<p>EmailConfig configures notifications via Email.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>sendResolved</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether or not to notify about resolved alerts.</p>
</td>
</tr>
<tr>
<td>
<code>to</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The email address to send notifications to.</p>
</td>
</tr>
<tr>
<td>
<code>from</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The sender address.</p>
</td>
</tr>
<tr>
<td>
<code>hello</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The hostname to identify to the SMTP server.</p>
</td>
</tr>
<tr>
<td>
<code>smarthost</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The SMTP host and port through which emails are sent. E.g. example.com:25</p>
</td>
</tr>
<tr>
<td>
<code>authUsername</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The username to use for authentication.</p>
</td>
</tr>
<tr>
<td>
<code>authPassword</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>The secret&rsquo;s key that contains the password to use for authentication.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>authSecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>The secret&rsquo;s key that contains the CRAM-MD5 secret.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>authIdentity</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The identity to use for authentication.</p>
</td>
</tr>
<tr>
<td>
<code>headers</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.KeyValue">
[]KeyValue
</a>
</em>
</td>
<td>
<p>Further headers email header key/value pairs. Overrides any headers
previously set by the notification implementation.</p>
</td>
</tr>
<tr>
<td>
<code>html</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The HTML body of the email notification.</p>
</td>
</tr>
<tr>
<td>
<code>text</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The text body of the email notification.</p>
</td>
</tr>
<tr>
<td>
<code>requireTLS</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>The SMTP TLS requirement.
Note that Go does not support unencrypted connections to remote SMTP endpoints.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.EurekaSDConfig">EurekaSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>Eureka SD configurations allow retrieving scrape targets using the Eureka REST API.
Prometheus will periodically check the REST endpoint and create a target for every app instance.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#eureka_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#eureka_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>server</code><br/>
<em>
string
</em>
</td>
<td>
<p>The URL to connect to the Eureka server.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>BasicAuth information to use on every scrape request.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization header to use on every scrape request.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional OAuth 2.0 configuration.
Cannot be set at the same time as <code>authorization</code> or <code>basic_auth</code>.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration applying to the target HTTP endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Refresh interval to re-read the instance list.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.FileSDConfig">FileSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>FileSDConfig defines a Prometheus file service discovery configuration
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#file_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#file_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>files</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.SDFile">
[]SDFile
</a>
</em>
</td>
<td>
<p>List of files to be used for file discovery. Recommendation: use absolute paths. While relative paths work, the
prometheus-operator project makes no guarantees about the working directory where the configuration file is
stored.
Files must be mounted using Prometheus.ConfigMaps or Prometheus.Secrets.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RefreshInterval configures the refresh interval at which Prometheus will reload the content of the files.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.Filter">Filter
</h3>
<div>
<p>Filter name and value pairs to limit the discovery process to a subset of available resources.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the Filter.</p>
</td>
</tr>
<tr>
<td>
<code>values</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Value to filter on.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.Filters">Filters
(<code>[]github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1.Filter</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.DockerSDConfig">DockerSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DockerSwarmSDConfig">DockerSwarmSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.EC2SDConfig">EC2SDConfig</a>)
</p>
<div>
</div>
<h3 id="monitoring.coreos.com/v1alpha1.GCESDConfig">GCESDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>GCESDConfig configures scrape targets from GCP GCE instances.
The private IP address is used by default, but may be changed to
the public IP address with relabeling.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#gce_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#gce_sd_config</a></p>
<p>The GCE service discovery will load the Google Cloud credentials
from the file specified by the GOOGLE_APPLICATION_CREDENTIALS environment variable.
See <a href="https://cloud.google.com/kubernetes-engine/docs/tutorials/authenticating-to-cloud-platform">https://cloud.google.com/kubernetes-engine/docs/tutorials/authenticating-to-cloud-platform</a></p>
<p>A pre-requisite for using GCESDConfig is that a Secret containing valid
Google Cloud credentials is mounted into the Prometheus or PrometheusAgent
pod via the <code>.spec.secrets</code> field and that the GOOGLE_APPLICATION_CREDENTIALS
environment variable is set to /etc/prometheus/secrets/<secret-name>/<credentials-filename.json>.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>project</code><br/>
<em>
string
</em>
</td>
<td>
<p>The Google Cloud Project ID</p>
</td>
</tr>
<tr>
<td>
<code>zone</code><br/>
<em>
string
</em>
</td>
<td>
<p>The zone of the scrape targets. If you need multiple zones use multiple GCESDConfigs.</p>
</td>
</tr>
<tr>
<td>
<code>filter</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Filter can be used optionally to filter the instance list by other criteria
Syntax of this filter is described in the filter query parameter section:
<a href="https://cloud.google.com/compute/docs/reference/latest/instances/list">https://cloud.google.com/compute/docs/reference/latest/instances/list</a></p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RefreshInterval configures the refresh interval at which Prometheus will re-read the instance list.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The port to scrape metrics from. If using the public IP address, this must
instead be specified in the relabeling rule.</p>
</td>
</tr>
<tr>
<td>
<code>tagSeparator</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The tag separator is used to separate the tags on concatenation</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.HTTPConfig">HTTPConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.DiscordConfig">DiscordConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.MSTeamsConfig">MSTeamsConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.MSTeamsV2Config">MSTeamsV2Config</a>, <a href="#monitoring.coreos.com/v1alpha1.OpsGenieConfig">OpsGenieConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.PagerDutyConfig">PagerDutyConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.PushoverConfig">PushoverConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.SNSConfig">SNSConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.SlackConfig">SlackConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.TelegramConfig">TelegramConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.VictorOpsConfig">VictorOpsConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.WeChatConfig">WeChatConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.WebexConfig">WebexConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.WebhookConfig">WebhookConfig</a>)
</p>
<div>
<p>HTTPConfig defines a client HTTP configuration.
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#http_config">https://prometheus.io/docs/alerting/latest/configuration/#http_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization header configuration for the client.
This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>BasicAuth for the client.
This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>OAuth2 client credentials used to fetch a token for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>bearerTokenSecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The secret&rsquo;s key that contains the bearer token to be used by the client
for authentication.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration for the client.</p>
</td>
</tr>
<tr>
<td>
<code>proxyURL</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional proxy URL.</p>
<p>If defined, this field takes precedence over <code>proxyUrl</code>.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>FollowRedirects specifies whether the client should follow HTTP 3xx redirects.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.HTTPSDConfig">HTTPSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>HTTPSDConfig defines a prometheus HTTP service discovery configuration
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#http_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#http_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>url</code><br/>
<em>
string
</em>
</td>
<td>
<p>URL from which the targets are fetched.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RefreshInterval configures the refresh interval at which Prometheus will re-query the
endpoint to update the target list.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>BasicAuth information to authenticate against the target HTTP endpoint.
More info: <a href="https://prometheus.io/docs/operating/configuration/#endpoints">https://prometheus.io/docs/operating/configuration/#endpoints</a>
Cannot be set at the same time as <code>authorization</code>, or <code>oAuth2</code>.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization header configuration to authenticate against the target HTTP endpoint.
Cannot be set at the same time as <code>oAuth2</code>, or <code>basicAuth</code>.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional OAuth 2.0 configuration to authenticate against the target HTTP endpoint.
Cannot be set at the same time as <code>authorization</code>, or <code>basicAuth</code>.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration applying to the target HTTP endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.HetznerSDConfig">HetznerSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>HetznerSDConfig allow retrieving scrape targets from Hetzner Cloud API and Robot API.
This service discovery uses the public IPv4 address by default, but that can be changed with relabeling
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#hetzner_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#hetzner_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>role</code><br/>
<em>
string
</em>
</td>
<td>
<p>The Hetzner role of entities that should be discovered.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>BasicAuth information to use on every scrape request, required when role is robot.
Role hcloud does not support basic auth.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization header configuration, required when role is hcloud.
Role robot does not support bearer token authentication.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional OAuth 2.0 configuration.
Cannot be used at the same time as <code>basic_auth</code> or <code>authorization</code>.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration to use on every scrape request.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>The port to scrape metrics from.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The time after which the servers are refreshed.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.InhibitRule">InhibitRule
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.AlertmanagerConfigSpec">AlertmanagerConfigSpec</a>)
</p>
<div>
<p>InhibitRule defines an inhibition rule that allows to mute alerts when other
alerts are already firing.
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#inhibit_rule">https://prometheus.io/docs/alerting/latest/configuration/#inhibit_rule</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>targetMatch</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.Matcher">
[]Matcher
</a>
</em>
</td>
<td>
<p>Matchers that have to be fulfilled in the alerts to be muted. The
operator enforces that the alert matches the resource&rsquo;s namespace.</p>
</td>
</tr>
<tr>
<td>
<code>sourceMatch</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.Matcher">
[]Matcher
</a>
</em>
</td>
<td>
<p>Matchers for which one or more alerts have to exist for the inhibition
to take effect. The operator enforces that the alert matches the
resource&rsquo;s namespace.</p>
</td>
</tr>
<tr>
<td>
<code>equal</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Labels that must have an equal value in the source and target alert for
the inhibition to take effect.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.IonosSDConfig">IonosSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>IonosSDConfig configurations allow retrieving scrape targets from IONOS resources.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#ionos_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#ionos_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>datacenterID</code><br/>
<em>
string
</em>
</td>
<td>
<p>The unique ID of the IONOS data center.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Port to scrape the metrics from.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Refresh interval to re-read the list of resources.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<p>Authorization` header configuration, required when using IONOS.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration to use when connecting to the IONOS API.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether the HTTP requests should follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether to enable HTTP2.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether to enable OAuth2.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.K8SSelectorConfig">K8SSelectorConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.KubernetesSDConfig">KubernetesSDConfig</a>)
</p>
<div>
<p>K8SSelectorConfig is Kubernetes Selector Config</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>role</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.KubernetesRole">
KubernetesRole
</a>
</em>
</td>
<td>
<p>Role specifies the type of Kubernetes resource to limit the service discovery to.
Accepted values are: Node, Pod, Endpoints, EndpointSlice, Service, Ingress.</p>
</td>
</tr>
<tr>
<td>
<code>label</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>An optional label selector to limit the service discovery to resources with specific labels and label values.
e.g: <code>node.kubernetes.io/instance-type=master</code></p>
</td>
</tr>
<tr>
<td>
<code>field</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>An optional field selector to limit the service discovery to resources which have fields with specific values.
e.g: <code>metadata.name=foobar</code></p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.KeyValue">KeyValue
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.EmailConfig">EmailConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.OpsGenieConfig">OpsGenieConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.PagerDutyConfig">PagerDutyConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.VictorOpsConfig">VictorOpsConfig</a>)
</p>
<div>
<p>KeyValue defines a (key, value) tuple.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>key</code><br/>
<em>
string
</em>
</td>
<td>
<p>Key of the tuple.</p>
</td>
</tr>
<tr>
<td>
<code>value</code><br/>
<em>
string
</em>
</td>
<td>
<p>Value of the tuple.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.KubernetesRole">KubernetesRole
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.K8SSelectorConfig">K8SSelectorConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.KubernetesSDConfig">KubernetesSDConfig</a>)
</p>
<div>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Endpoints&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;EndpointSlice&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Ingress&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Node&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Pod&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Service&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.KubernetesSDConfig">KubernetesSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>KubernetesSDConfig allows retrieving scrape targets from Kubernetes&rsquo; REST API.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiServer</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The API server address consisting of a hostname or IP address followed
by an optional port number.
If left empty, Prometheus is assumed to run inside
of the cluster. It will discover API servers automatically and use the pod&rsquo;s
CA certificate and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/.</p>
</td>
</tr>
<tr>
<td>
<code>role</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.KubernetesRole">
KubernetesRole
</a>
</em>
</td>
<td>
<p>Role of the Kubernetes entities that should be discovered.
Role <code>Endpointslice</code> requires Prometheus &gt;= v2.21.0</p>
</td>
</tr>
<tr>
<td>
<code>namespaces</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.NamespaceDiscovery">
NamespaceDiscovery
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional namespace discovery. If omitted, Prometheus discovers targets across all namespaces.</p>
</td>
</tr>
<tr>
<td>
<code>attachMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.AttachMetadata">
AttachMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional metadata to attach to discovered targets.
It requires Prometheus &gt;= v2.35.0 when using the <code>Pod</code> role and
Prometheus &gt;= v2.37.0 for <code>Endpoints</code> and <code>Endpointslice</code> roles.</p>
</td>
</tr>
<tr>
<td>
<code>selectors</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.K8SSelectorConfig">
[]K8SSelectorConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Selector to select objects.
It requires Prometheus &gt;= v2.17.0</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>BasicAuth information to use on every scrape request.
Cannot be set at the same time as <code>authorization</code>, or <code>oauth2</code>.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization header to use on every scrape request.
Cannot be set at the same time as <code>basicAuth</code>, or <code>oauth2</code>.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional OAuth 2.0 configuration.
Cannot be set at the same time as <code>authorization</code>, or <code>basicAuth</code>.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration to connect to the Kubernetes API.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.KumaSDConfig">KumaSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>KumaSDConfig allow retrieving scrape targets from Kuma&rsquo;s control plane.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kuma_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kuma_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>server</code><br/>
<em>
string
</em>
</td>
<td>
<p>Address of the Kuma Control Plane&rsquo;s MADS xDS server.</p>
</td>
</tr>
<tr>
<td>
<code>clientID</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Client id is used by Kuma Control Plane to compute Monitoring Assignment for specific Prometheus backend.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The time to wait between polling update requests.</p>
</td>
</tr>
<tr>
<td>
<code>fetchTimeout</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The time after which the monitoring assignments are refreshed.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration to use on every scrape request</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>BasicAuth information to use on every scrape request.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization header to use on every scrape request.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional OAuth 2.0 configuration.
Cannot be set at the same time as <code>authorization</code>, or <code>basicAuth</code>.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.LightSailSDConfig">LightSailSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>LightSailSDConfig configurations allow retrieving scrape targets from AWS Lightsail instances.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#lightsail_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#lightsail_sd_config</a>
TODO: Need to document that we will not be supporting the <code>_file</code> fields.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>region</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The AWS region.</p>
</td>
</tr>
<tr>
<td>
<code>accessKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AccessKey is the AWS API key.</p>
</td>
</tr>
<tr>
<td>
<code>secretKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecretKey is the AWS API secret.</p>
</td>
</tr>
<tr>
<td>
<code>roleARN</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>AWS Role ARN, an alternative to using AWS API keys.</p>
</td>
</tr>
<tr>
<td>
<code>endpoint</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Custom endpoint to be used.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Refresh interval to re-read the list of instances.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Port to scrape the metrics from.
If using the public IP address, this must instead be specified in the relabeling rule.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional HTTP basic authentication information.
Cannot be set at the same time as <code>authorization</code>, or <code>oauth2</code>.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional <code>authorization</code> HTTP header configuration.
Cannot be set at the same time as <code>basicAuth</code>, or <code>oauth2</code>.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional OAuth2.0 configuration.
Cannot be set at the same time as <code>basicAuth</code>, or <code>authorization</code>.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration to connect to the Puppet DB.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether the HTTP requests should follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether to enable HTTP2.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.LinodeSDConfig">LinodeSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>LinodeSDConfig configurations allow retrieving scrape targets from Linode&rsquo;s Linode APIv4.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#linode_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#linode_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>region</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional region to filter on.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Default port to scrape metrics from.</p>
</td>
</tr>
<tr>
<td>
<code>tagSeparator</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The string by which Linode Instance tags are joined into the tag label.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Time after which the linode instances are refreshed.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization header configuration.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional OAuth 2.0 configuration.
Cannot be used at the same time as <code>authorization</code>.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration applying to the target HTTP endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.MSTeamsConfig">MSTeamsConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Receiver">Receiver</a>)
</p>
<div>
<p>MSTeamsConfig configures notifications via Microsoft Teams.
It requires Alertmanager &gt;= 0.26.0.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>sendResolved</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to notify about resolved alerts.</p>
</td>
</tr>
<tr>
<td>
<code>webhookUrl</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>MSTeams webhook URL.</p>
</td>
</tr>
<tr>
<td>
<code>title</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Message title template.</p>
</td>
</tr>
<tr>
<td>
<code>summary</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Message summary template.
It requires Alertmanager &gt;= 0.27.0.</p>
</td>
</tr>
<tr>
<td>
<code>text</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Message body template.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">
HTTPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTP client configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.MSTeamsV2Config">MSTeamsV2Config
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Receiver">Receiver</a>)
</p>
<div>
<p>MSTeamsV2Config configures notifications via Microsoft Teams using the new message format with adaptive cards as required by flows
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#msteamsv2_config">https://prometheus.io/docs/alerting/latest/configuration/#msteamsv2_config</a>
It requires Alertmanager &gt;= 0.28.0.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>sendResolved</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to notify about resolved alerts.</p>
</td>
</tr>
<tr>
<td>
<code>webhookURL</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>MSTeams incoming webhook URL.</p>
</td>
</tr>
<tr>
<td>
<code>title</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Message title template.</p>
</td>
</tr>
<tr>
<td>
<code>text</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Message body template.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">
HTTPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTP client configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.MatchType">MatchType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Matcher">Matcher</a>)
</p>
<div>
<p>MatchType is a comparison operator on a Matcher</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;=&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;!=&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;!~&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;=~&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.Matcher">Matcher
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.InhibitRule">InhibitRule</a>, <a href="#monitoring.coreos.com/v1alpha1.Route">Route</a>)
</p>
<div>
<p>Matcher defines how to match on alert&rsquo;s labels.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Label to match.</p>
</td>
</tr>
<tr>
<td>
<code>value</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Label value to match.</p>
</td>
</tr>
<tr>
<td>
<code>matchType</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.MatchType">
MatchType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Match operation available with AlertManager &gt;= v0.22.0 and
takes precedence over Regex (deprecated) if non-empty.</p>
</td>
</tr>
<tr>
<td>
<code>regex</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to match on equality (false) or regular-expression (true).
Deprecated: for AlertManager &gt;= v0.22.0, <code>matchType</code> should be used instead.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.Month">Month
(<code>string</code> alias)</h3>
<div>
<p>Month of the year</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;april&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;august&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;december&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;february&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;january&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;july&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;june&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;march&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;may&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;november&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;october&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;september&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.MonthRange">MonthRange
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.TimeInterval">TimeInterval</a>)
</p>
<div>
<p>MonthRange is an inclusive range of months of the year beginning in January
Months can be specified by name (e.g &lsquo;January&rsquo;) by numerical month (e.g &lsquo;1&rsquo;) or as an inclusive range (e.g &lsquo;January:March&rsquo;, &lsquo;1:3&rsquo;, &lsquo;1:March&rsquo;)</p>
</div>
<h3 id="monitoring.coreos.com/v1alpha1.MuteTimeInterval">MuteTimeInterval
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.AlertmanagerConfigSpec">AlertmanagerConfigSpec</a>)
</p>
<div>
<p>MuteTimeInterval specifies the periods in time when notifications will be muted</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the time interval</p>
</td>
</tr>
<tr>
<td>
<code>timeIntervals</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.TimeInterval">
[]TimeInterval
</a>
</em>
</td>
<td>
<p>TimeIntervals is a list of TimeInterval</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.NamespaceDiscovery">NamespaceDiscovery
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.KubernetesSDConfig">KubernetesSDConfig</a>)
</p>
<div>
<p>NamespaceDiscovery is the configuration for discovering
Kubernetes namespaces.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ownNamespace</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Includes the namespace in which the Prometheus pod runs to the list of watched namespaces.</p>
</td>
</tr>
<tr>
<td>
<code>names</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of namespaces where to watch for resources.
If empty and <code>ownNamespace</code> isn&rsquo;t true, Prometheus watches for resources in all namespaces.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.NomadSDConfig">NomadSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>NomadSDConfig configurations allow retrieving scrape targets from Nomad&rsquo;s Service API.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#nomad_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#nomad_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>allowStale</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>The information to access the Nomad API. It is to be defined
as the Nomad documentation requires.</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>region</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>server</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>tagSeparator</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>BasicAuth information to use on every scrape request.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization header to use on every scrape request.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional OAuth 2.0 configuration.
Cannot be set at the same time as <code>authorization</code> or <code>basic_auth</code>.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration applying to the target HTTP endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.OVHCloudSDConfig">OVHCloudSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>OVHCloudSDConfig configurations allow retrieving scrape targets from OVHcloud&rsquo;s dedicated servers and VPS using their API.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#ovhcloud_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#ovhcloud_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>applicationKey</code><br/>
<em>
string
</em>
</td>
<td>
<p>Access key to use. <a href="https://api.ovh.com">https://api.ovh.com</a>.</p>
</td>
</tr>
<tr>
<td>
<code>applicationSecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>consumerKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>service</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.OVHService">
OVHService
</a>
</em>
</td>
<td>
<p>Service of the targets to retrieve. Must be <code>VPS</code> or <code>DedicatedServer</code>.</p>
</td>
</tr>
<tr>
<td>
<code>endpoint</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Custom endpoint to be used.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Refresh interval to re-read the resources list.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.OVHService">OVHService
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.OVHCloudSDConfig">OVHCloudSDConfig</a>)
</p>
<div>
<p>Service of the targets to retrieve. Must be <code>VPS</code> or <code>DedicatedServer</code>.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;DedicatedServer&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;VPS&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.OpenStackRole">OpenStackRole
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.OpenStackSDConfig">OpenStackSDConfig</a>)
</p>
<div>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Hypervisor&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Instance&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;LoadBalancer&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.OpenStackSDConfig">OpenStackSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>OpenStackSDConfig allow retrieving scrape targets from OpenStack Nova instances.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#openstack_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#openstack_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>role</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.OpenStackRole">
OpenStackRole
</a>
</em>
</td>
<td>
<p>The OpenStack role of entities that should be discovered.</p>
<p>Note: The <code>LoadBalancer</code> role requires Prometheus &gt;= v3.2.0.</p>
</td>
</tr>
<tr>
<td>
<code>region</code><br/>
<em>
string
</em>
</td>
<td>
<p>The OpenStack Region.</p>
</td>
</tr>
<tr>
<td>
<code>identityEndpoint</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>IdentityEndpoint specifies the HTTP endpoint that is required to work with
the Identity API of the appropriate version.</p>
</td>
</tr>
<tr>
<td>
<code>username</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Username is required if using Identity V2 API. Consult with your provider&rsquo;s
control panel to discover your account&rsquo;s username.
In Identity V3, either userid or a combination of username
and domainId or domainName are needed</p>
</td>
</tr>
<tr>
<td>
<code>userid</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>UserID</p>
</td>
</tr>
<tr>
<td>
<code>password</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Password for the Identity V2 and V3 APIs. Consult with your provider&rsquo;s
control panel to discover your account&rsquo;s preferred method of authentication.</p>
</td>
</tr>
<tr>
<td>
<code>domainName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>At most one of domainId and domainName must be provided if using username
with Identity V3. Otherwise, either are optional.</p>
</td>
</tr>
<tr>
<td>
<code>domainID</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>DomainID</p>
</td>
</tr>
<tr>
<td>
<code>projectName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The ProjectId and ProjectName fields are optional for the Identity V2 API.
Some providers allow you to specify a ProjectName instead of the ProjectId.
Some require both. Your provider&rsquo;s authentication policies will determine
how these fields influence authentication.</p>
</td>
</tr>
<tr>
<td>
<code>projectID</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProjectID</p>
</td>
</tr>
<tr>
<td>
<code>applicationCredentialName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The ApplicationCredentialID or ApplicationCredentialName fields are
required if using an application credential to authenticate. Some providers
allow you to create an application credential to authenticate rather than a
password.</p>
</td>
</tr>
<tr>
<td>
<code>applicationCredentialId</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ApplicationCredentialID</p>
</td>
</tr>
<tr>
<td>
<code>applicationCredentialSecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The applicationCredentialSecret field is required if using an application
credential to authenticate.</p>
</td>
</tr>
<tr>
<td>
<code>allTenants</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether the service discovery should list all instances for all projects.
It is only relevant for the &lsquo;instance&rsquo; role and usually requires admin permissions.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Refresh interval to re-read the instance list.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The port to scrape metrics from. If using the public IP address, this must
instead be specified in the relabeling rule.</p>
</td>
</tr>
<tr>
<td>
<code>availability</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Availability of the endpoint to connect to.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration applying to the target HTTP endpoint.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.OpsGenieConfig">OpsGenieConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Receiver">Receiver</a>)
</p>
<div>
<p>OpsGenieConfig configures notifications via OpsGenie.
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#opsgenie_config">https://prometheus.io/docs/alerting/latest/configuration/#opsgenie_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>sendResolved</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether or not to notify about resolved alerts.</p>
</td>
</tr>
<tr>
<td>
<code>apiKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The secret&rsquo;s key that contains the OpsGenie API key.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>apiURL</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The URL to send OpsGenie API requests to.</p>
</td>
</tr>
<tr>
<td>
<code>message</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Alert text limited to 130 characters.</p>
</td>
</tr>
<tr>
<td>
<code>description</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Description of the incident.</p>
</td>
</tr>
<tr>
<td>
<code>source</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Backlink to the sender of the notification.</p>
</td>
</tr>
<tr>
<td>
<code>tags</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Comma separated list of tags attached to the notifications.</p>
</td>
</tr>
<tr>
<td>
<code>note</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Additional alert note.</p>
</td>
</tr>
<tr>
<td>
<code>priority</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Priority level of alert. Possible values are P1, P2, P3, P4, and P5.</p>
</td>
</tr>
<tr>
<td>
<code>updateAlerts</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to update message and description of the alert in OpsGenie if it already exists
By default, the alert is never updated in OpsGenie, the new message only appears in activity log.</p>
</td>
</tr>
<tr>
<td>
<code>details</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.KeyValue">
[]KeyValue
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>A set of arbitrary key/value pairs that provide further detail about the incident.</p>
</td>
</tr>
<tr>
<td>
<code>responders</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.OpsGenieConfigResponder">
[]OpsGenieConfigResponder
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of responders responsible for notifications.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">
HTTPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTP client configuration.</p>
</td>
</tr>
<tr>
<td>
<code>entity</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional field that can be used to specify which domain alert is related to.</p>
</td>
</tr>
<tr>
<td>
<code>actions</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Comma separated list of actions that will be available for the alert.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.OpsGenieConfigResponder">OpsGenieConfigResponder
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.OpsGenieConfig">OpsGenieConfig</a>)
</p>
<div>
<p>OpsGenieConfigResponder defines a responder to an incident.
One of <code>id</code>, <code>name</code> or <code>username</code> has to be defined.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>id</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ID of the responder.</p>
</td>
</tr>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of the responder.</p>
</td>
</tr>
<tr>
<td>
<code>username</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Username of the responder.</p>
</td>
</tr>
<tr>
<td>
<code>type</code><br/>
<em>
string
</em>
</td>
<td>
<p>Type of responder.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.PagerDutyConfig">PagerDutyConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Receiver">Receiver</a>)
</p>
<div>
<p>PagerDutyConfig configures notifications via PagerDuty.
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#pagerduty_config">https://prometheus.io/docs/alerting/latest/configuration/#pagerduty_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>sendResolved</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether or not to notify about resolved alerts.</p>
</td>
</tr>
<tr>
<td>
<code>routingKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The secret&rsquo;s key that contains the PagerDuty integration key (when using
Events API v2). Either this field or <code>serviceKey</code> needs to be defined.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>serviceKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The secret&rsquo;s key that contains the PagerDuty service key (when using
integration type &ldquo;Prometheus&rdquo;). Either this field or <code>routingKey</code> needs to
be defined.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>url</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The URL to send requests to.</p>
</td>
</tr>
<tr>
<td>
<code>client</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Client identification.</p>
</td>
</tr>
<tr>
<td>
<code>clientURL</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Backlink to the sender of notification.</p>
</td>
</tr>
<tr>
<td>
<code>description</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Description of the incident.</p>
</td>
</tr>
<tr>
<td>
<code>severity</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Severity of the incident.</p>
</td>
</tr>
<tr>
<td>
<code>class</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The class/type of the event.</p>
</td>
</tr>
<tr>
<td>
<code>group</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>A cluster or grouping of sources.</p>
</td>
</tr>
<tr>
<td>
<code>component</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The part or component of the affected system that is broken.</p>
</td>
</tr>
<tr>
<td>
<code>details</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.KeyValue">
[]KeyValue
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Arbitrary key/value pairs that provide further detail about the incident.</p>
</td>
</tr>
<tr>
<td>
<code>pagerDutyImageConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.PagerDutyImageConfig">
[]PagerDutyImageConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>A list of image details to attach that provide further detail about an incident.</p>
</td>
</tr>
<tr>
<td>
<code>pagerDutyLinkConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.PagerDutyLinkConfig">
[]PagerDutyLinkConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>A list of link details to attach that provide further detail about an incident.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">
HTTPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTP client configuration.</p>
</td>
</tr>
<tr>
<td>
<code>source</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Unique location of the affected system.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.PagerDutyImageConfig">PagerDutyImageConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.PagerDutyConfig">PagerDutyConfig</a>)
</p>
<div>
<p>PagerDutyImageConfig attaches images to an incident</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>src</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Src of the image being attached to the incident</p>
</td>
</tr>
<tr>
<td>
<code>href</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional URL; makes the image a clickable link.</p>
</td>
</tr>
<tr>
<td>
<code>alt</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Alt is the optional alternative text for the image.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.PagerDutyLinkConfig">PagerDutyLinkConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.PagerDutyConfig">PagerDutyConfig</a>)
</p>
<div>
<p>PagerDutyLinkConfig attaches text links to an incident</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>href</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Href is the URL of the link to be attached</p>
</td>
</tr>
<tr>
<td>
<code>alt</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Text that describes the purpose of the link, and can be used as the link&rsquo;s text.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.ParsedRange">ParsedRange
</h3>
<div>
<p>ParsedRange is an integer representation of a range</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>start</code><br/>
<em>
int
</em>
</td>
<td>
<p>Start is the beginning of the range</p>
</td>
</tr>
<tr>
<td>
<code>end</code><br/>
<em>
int
</em>
</td>
<td>
<p>End of the range</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.PrometheusAgentMode">PrometheusAgentMode
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.PrometheusAgentSpec">PrometheusAgentSpec</a>)
</p>
<div>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;DaemonSet&#34;</p></td>
<td><p>Deploys PrometheusAgent as DaemonSet.</p>
</td>
</tr><tr><td><p>&#34;StatefulSet&#34;</p></td>
<td><p>Deploys PrometheusAgent as StatefulSet.</p>
</td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.PrometheusAgentSpec">PrometheusAgentSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.PrometheusAgent">PrometheusAgent</a>)
</p>
<div>
<p>PrometheusAgentSpec is a specification of the desired behavior of the Prometheus agent. More info:
<a href="https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status">https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>mode</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.PrometheusAgentMode">
PrometheusAgentMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Mode defines how the Prometheus operator deploys the PrometheusAgent pod(s).</p>
<p>(Alpha) Using this field requires the <code>PrometheusAgentDaemonSet</code> feature gate to be enabled.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
Monitoring v1.EmbeddedObjectMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata configures labels and annotations which are propagated to the Prometheus pods.</p>
<p>The following items are reserved and cannot be overridden:
* &ldquo;prometheus&rdquo; label, set to the name of the Prometheus object.
* &ldquo;app.kubernetes.io/instance&rdquo; label, set to the name of the Prometheus object.
* &ldquo;app.kubernetes.io/managed-by&rdquo; label, set to &ldquo;prometheus-operator&rdquo;.
* &ldquo;app.kubernetes.io/name&rdquo; label, set to &ldquo;prometheus&rdquo;.
* &ldquo;app.kubernetes.io/version&rdquo; label, set to the Prometheus version.
* &ldquo;operator.prometheus.io/name&rdquo; label, set to the name of the Prometheus object.
* &ldquo;operator.prometheus.io/shard&rdquo; label, set to the shard number of the Prometheus object.
* &ldquo;kubectl.kubernetes.io/default-container&rdquo; annotation, set to &ldquo;prometheus&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitorSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ServiceMonitors to be selected for target discovery. An empty label
selector matches all objects. A null label selector matches no objects.</p>
<p>If <code>spec.serviceMonitorSelector</code>, <code>spec.podMonitorSelector</code>, <code>spec.probeSelector</code>
and <code>spec.scrapeConfigSelector</code> are null, the Prometheus configuration is unmanaged.
The Prometheus operator will ensure that the Prometheus configuration&rsquo;s
Secret exists, but it is the responsibility of the user to provide the raw
gzipped Prometheus configuration under the <code>prometheus.yaml.gz</code> key.
This behavior is <em>deprecated</em> and will be removed in the next major version
of the custom resource definition. It is recommended to use
<code>spec.additionalScrapeConfigs</code> instead.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitorNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for ServicedMonitors discovery. An empty label selector
matches all namespaces. A null label selector (default value) matches the current
namespace only.</p>
</td>
</tr>
<tr>
<td>
<code>podMonitorSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>PodMonitors to be selected for target discovery. An empty label selector
matches all objects. A null label selector matches no objects.</p>
<p>If <code>spec.serviceMonitorSelector</code>, <code>spec.podMonitorSelector</code>, <code>spec.probeSelector</code>
and <code>spec.scrapeConfigSelector</code> are null, the Prometheus configuration is unmanaged.
The Prometheus operator will ensure that the Prometheus configuration&rsquo;s
Secret exists, but it is the responsibility of the user to provide the raw
gzipped Prometheus configuration under the <code>prometheus.yaml.gz</code> key.
This behavior is <em>deprecated</em> and will be removed in the next major version
of the custom resource definition. It is recommended to use
<code>spec.additionalScrapeConfigs</code> instead.</p>
</td>
</tr>
<tr>
<td>
<code>podMonitorNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for PodMonitors discovery. An empty label selector
matches all namespaces. A null label selector (default value) matches the current
namespace only.</p>
</td>
</tr>
<tr>
<td>
<code>probeSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Probes to be selected for target discovery. An empty label selector
matches all objects. A null label selector matches no objects.</p>
<p>If <code>spec.serviceMonitorSelector</code>, <code>spec.podMonitorSelector</code>, <code>spec.probeSelector</code>
and <code>spec.scrapeConfigSelector</code> are null, the Prometheus configuration is unmanaged.
The Prometheus operator will ensure that the Prometheus configuration&rsquo;s
Secret exists, but it is the responsibility of the user to provide the raw
gzipped Prometheus configuration under the <code>prometheus.yaml.gz</code> key.
This behavior is <em>deprecated</em> and will be removed in the next major version
of the custom resource definition. It is recommended to use
<code>spec.additionalScrapeConfigs</code> instead.</p>
</td>
</tr>
<tr>
<td>
<code>probeNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for Probe discovery. An empty label
selector matches all namespaces. A null label selector matches the
current namespace only.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeConfigSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ScrapeConfigs to be selected for target discovery. An empty label
selector matches all objects. A null label selector matches no objects.</p>
<p>If <code>spec.serviceMonitorSelector</code>, <code>spec.podMonitorSelector</code>, <code>spec.probeSelector</code>
and <code>spec.scrapeConfigSelector</code> are null, the Prometheus configuration is unmanaged.
The Prometheus operator will ensure that the Prometheus configuration&rsquo;s
Secret exists, but it is the responsibility of the user to provide the raw
gzipped Prometheus configuration under the <code>prometheus.yaml.gz</code> key.
This behavior is <em>deprecated</em> and will be removed in the next major version
of the custom resource definition. It is recommended to use
<code>spec.additionalScrapeConfigs</code> instead.</p>
<p>Note that the ScrapeConfig custom resource definition is currently at Alpha level.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeConfigNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespaces to match for ScrapeConfig discovery. An empty label selector
matches all namespaces. A null label selector matches the current
namespace only.</p>
<p>Note that the ScrapeConfig custom resource definition is currently at Alpha level.</p>
</td>
</tr>
<tr>
<td>
<code>version</code><br/>
<em>
string
</em>
</td>
<td>
<p>Version of Prometheus being deployed. The operator uses this information
to generate the Prometheus StatefulSet + configuration files.</p>
<p>If not specified, the operator assumes the latest upstream version of
Prometheus available at the time when the version of the operator was
released.</p>
</td>
</tr>
<tr>
<td>
<code>paused</code><br/>
<em>
bool
</em>
</td>
<td>
<p>When a Prometheus deployment is paused, no actions except for deletion
will be performed on the underlying objects.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Container image name for Prometheus. If specified, it takes precedence
over the <code>spec.baseImage</code>, <code>spec.tag</code> and <code>spec.sha</code> fields.</p>
<p>Specifying <code>spec.version</code> is still necessary to ensure the Prometheus
Operator knows which version of Prometheus is being configured.</p>
<p>If neither <code>spec.image</code> nor <code>spec.baseImage</code> are defined, the operator
will use the latest upstream version of Prometheus available at the time
when the operator was released.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#pullpolicy-v1-core">
Kubernetes core/v1.PullPolicy
</a>
</em>
</td>
<td>
<p>Image pull policy for the &lsquo;prometheus&rsquo;, &lsquo;init-config-reloader&rsquo; and &lsquo;config-reloader&rsquo; containers.
See <a href="https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy">https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy</a> for more details.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>An optional list of references to Secrets in the same namespace
to use for pulling images from registries.
See <a href="http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod">http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod</a></p>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Number of replicas of each shard to deploy for a Prometheus deployment.
<code>spec.replicas</code> multiplied by <code>spec.shards</code> is the total number of Pods
created.</p>
<p>Default: 1</p>
</td>
</tr>
<tr>
<td>
<code>shards</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Number of shards to distribute the scraped targets onto.</p>
<p><code>spec.replicas</code> multiplied by <code>spec.shards</code> is the total number of Pods
being created.</p>
<p>When not defined, the operator assumes only one shard.</p>
<p>Note that scaling down shards will not reshard data onto the remaining
instances, it must be manually moved. Increasing shards will not reshard
data either but it will continue to be available from the same
instances. To query globally, use either
* Thanos sidecar + querier for query federation and Thanos Ruler for rules.
* Remote-write to send metrics to a central location.</p>
<p>By default, the sharding of targets is performed on:
* The <code>__address__</code> target&rsquo;s metadata label for PodMonitor,
ServiceMonitor and ScrapeConfig resources.
* The <code>__param_target__</code> label for Probe resources.</p>
<p>Users can define their own sharding implementation by setting the
<code>__tmp_hash</code> label during the target discovery with relabeling
configuration (either in the monitoring resources or via scrape class).</p>
<p>You can also disable sharding on a specific target by setting the
<code>__tmp_disable_sharding</code> label with relabeling configuration. When
the label value isn&rsquo;t empty, all Prometheus shards will scrape the target.</p>
</td>
</tr>
<tr>
<td>
<code>replicaExternalLabelName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of Prometheus external label used to denote the replica name.
The external label will <em>not</em> be added when the field is set to the
empty string (<code>&quot;&quot;</code>).</p>
<p>Default: &ldquo;prometheus_replica&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>prometheusExternalLabelName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of Prometheus external label used to denote the Prometheus instance
name. The external label will <em>not</em> be added when the field is set to
the empty string (<code>&quot;&quot;</code>).</p>
<p>Default: &ldquo;prometheus&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>logLevel</code><br/>
<em>
string
</em>
</td>
<td>
<p>Log level for Prometheus and the config-reloader sidecar.</p>
</td>
</tr>
<tr>
<td>
<code>logFormat</code><br/>
<em>
string
</em>
</td>
<td>
<p>Log format for Log level for Prometheus and the config-reloader sidecar.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<p>Interval between consecutive scrapes.</p>
<p>Default: &ldquo;30s&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>scrapeTimeout</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<p>Number of seconds to wait until a scrape request times out.
The value cannot be greater than the scrape interval otherwise the operator will reject the resource.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeProtocols</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ScrapeProtocol">
[]Monitoring v1.ScrapeProtocol
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The protocols to negotiate during a scrape. It tells clients the
protocols supported by Prometheus in order of preference (from most to least preferred).</p>
<p>If unset, Prometheus uses its default value.</p>
<p>It requires Prometheus &gt;= v2.49.0.</p>
<p><code>PrometheusText1.0.0</code> requires Prometheus &gt;= v3.0.0.</p>
</td>
</tr>
<tr>
<td>
<code>externalLabels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>The labels to add to any time series or alerts when communicating with
external systems (federation, remote storage, Alertmanager).
Labels defined by <code>spec.replicaExternalLabelName</code> and
<code>spec.prometheusExternalLabelName</code> take precedence over this list.</p>
</td>
</tr>
<tr>
<td>
<code>enableRemoteWriteReceiver</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Enable Prometheus to be used as a receiver for the Prometheus remote
write protocol.</p>
<p>WARNING: This is not considered an efficient way of ingesting samples.
Use it with caution for specific low-volume use cases.
It is not suitable for replacing the ingestion via scraping and turning
Prometheus into a push-based metrics collection system.
For more information see <a href="https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver">https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver</a></p>
<p>It requires Prometheus &gt;= v2.33.0.</p>
</td>
</tr>
<tr>
<td>
<code>enableOTLPReceiver</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enable Prometheus to be used as a receiver for the OTLP Metrics protocol.</p>
<p>Note that the OTLP receiver endpoint is automatically enabled if <code>.spec.otlpConfig</code> is defined.</p>
<p>It requires Prometheus &gt;= v2.47.0.</p>
</td>
</tr>
<tr>
<td>
<code>remoteWriteReceiverMessageVersions</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.RemoteWriteMessageVersion">
[]Monitoring v1.RemoteWriteMessageVersion
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of the protobuf message versions to accept when receiving the
remote writes.</p>
<p>It requires Prometheus &gt;= v2.54.0.</p>
</td>
</tr>
<tr>
<td>
<code>enableFeatures</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.EnableFeature">
[]Monitoring v1.EnableFeature
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enable access to Prometheus feature flags. By default, no features are enabled.</p>
<p>Enabling features which are disabled by default is entirely outside the
scope of what the maintainers will support and by doing so, you accept
that this behaviour may break at any time without notice.</p>
<p>For more information see <a href="https://prometheus.io/docs/prometheus/latest/feature_flags/">https://prometheus.io/docs/prometheus/latest/feature_flags/</a></p>
</td>
</tr>
<tr>
<td>
<code>externalUrl</code><br/>
<em>
string
</em>
</td>
<td>
<p>The external URL under which the Prometheus service is externally
available. This is necessary to generate correct URLs (for instance if
Prometheus is accessible behind an Ingress resource).</p>
</td>
</tr>
<tr>
<td>
<code>routePrefix</code><br/>
<em>
string
</em>
</td>
<td>
<p>The route prefix Prometheus registers HTTP handlers for.</p>
<p>This is useful when using <code>spec.externalURL</code>, and a proxy is rewriting
HTTP routes of a request, and the actual ExternalURL is still true, but
the server serves requests under a different route prefix. For example
for use with <code>kubectl proxy</code>.</p>
</td>
</tr>
<tr>
<td>
<code>storage</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.StorageSpec">
Monitoring v1.StorageSpec
</a>
</em>
</td>
<td>
<p>Storage defines the storage used by Prometheus.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<p>Volumes allows the configuration of additional volumes on the output
StatefulSet definition. Volumes specified will be appended to other
volumes that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<p>VolumeMounts allows the configuration of additional VolumeMounts.</p>
<p>VolumeMounts will be appended to other VolumeMounts in the &lsquo;prometheus&rsquo;
container, that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>persistentVolumeClaimRetentionPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#statefulsetpersistentvolumeclaimretentionpolicy-v1-apps">
Kubernetes apps/v1.StatefulSetPersistentVolumeClaimRetentionPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The field controls if and how PVCs are deleted during the lifecycle of a StatefulSet.
The default behavior is all PVCs are retained.
This is an alpha field from kubernetes 1.23 until 1.26 and a beta field from 1.26.
It requires enabling the StatefulSetAutoDeletePVC feature gate.</p>
</td>
</tr>
<tr>
<td>
<code>web</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.PrometheusWebSpec">
Monitoring v1.PrometheusWebSpec
</a>
</em>
</td>
<td>
<p>Defines the configuration of the Prometheus web server.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>Defines the resources requests and limits of the &lsquo;prometheus&rsquo; container.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>Defines on which Nodes the Pods are scheduled.</p>
</td>
</tr>
<tr>
<td>
<code>serviceAccountName</code><br/>
<em>
string
</em>
</td>
<td>
<p>ServiceAccountName is the name of the ServiceAccount to use to run the
Prometheus Pods.</p>
</td>
</tr>
<tr>
<td>
<code>automountServiceAccountToken</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>AutomountServiceAccountToken indicates whether a service account token should be automatically mounted in the pod.
If the field isn&rsquo;t set, the operator mounts the service account token by default.</p>
<p><strong>Warning:</strong> be aware that by default, Prometheus requires the service account token for Kubernetes service discovery.
It is possible to use strategic merge patch to project the service account token into the &lsquo;prometheus&rsquo; container.</p>
</td>
</tr>
<tr>
<td>
<code>secrets</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Secrets is a list of Secrets in the same namespace as the Prometheus
object, which shall be mounted into the Prometheus Pods.
Each Secret is added to the StatefulSet definition as a volume named <code>secret-&lt;secret-name&gt;</code>.
The Secrets are mounted into /etc/prometheus/secrets/<secret-name> in the &lsquo;prometheus&rsquo; container.</p>
</td>
</tr>
<tr>
<td>
<code>configMaps</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>ConfigMaps is a list of ConfigMaps in the same namespace as the Prometheus
object, which shall be mounted into the Prometheus Pods.
Each ConfigMap is added to the StatefulSet definition as a volume named <code>configmap-&lt;configmap-name&gt;</code>.
The ConfigMaps are mounted into /etc/prometheus/configmaps/<configmap-name> in the &lsquo;prometheus&rsquo; container.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the Pods&rsquo; affinity scheduling rules if specified.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the Pods&rsquo; tolerations if specified.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.TopologySpreadConstraint">
[]Monitoring v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the pod&rsquo;s topology spread constraints if specified.</p>
</td>
</tr>
<tr>
<td>
<code>remoteWrite</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.RemoteWriteSpec">
[]Monitoring v1.RemoteWriteSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the list of remote write configurations.</p>
</td>
</tr>
<tr>
<td>
<code>otlp</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OTLPConfig">
Monitoring v1.OTLPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Settings related to the OTLP receiver feature.
It requires Prometheus &gt;= v2.55.0.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecurityContext holds pod-level security attributes and common container settings.
This defaults to the default PodSecurityContext.</p>
</td>
</tr>
<tr>
<td>
<code>dnsPolicy</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.DNSPolicy">
Monitoring v1.DNSPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the DNS policy for the pods.</p>
</td>
</tr>
<tr>
<td>
<code>dnsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.PodDNSConfig">
Monitoring v1.PodDNSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the DNS configuration for the pods.</p>
</td>
</tr>
<tr>
<td>
<code>listenLocal</code><br/>
<em>
bool
</em>
</td>
<td>
<p>When true, the Prometheus server listens on the loopback address
instead of the Pod IP&rsquo;s address.</p>
</td>
</tr>
<tr>
<td>
<code>enableServiceLinks</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Indicates whether information about services should be injected into pod&rsquo;s environment variables</p>
</td>
</tr>
<tr>
<td>
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Containers allows injecting additional containers or modifying operator
generated containers. This can be used to allow adding an authentication
proxy to the Pods or to change the behavior of an operator generated
container. Containers described here modify an operator generated
container if they share the same name and modifications are done via a
strategic merge patch.</p>
<p>The names of containers managed by the operator are:
* <code>prometheus</code>
* <code>config-reloader</code>
* <code>thanos-sidecar</code></p>
<p>Overriding containers is entirely outside the scope of what the
maintainers will support and by doing so, you accept that this behaviour
may break at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>initContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>InitContainers allows injecting initContainers to the Pod definition. Those
can be used to e.g.  fetch secrets for injection into the Prometheus
configuration from external sources. Any errors during the execution of
an initContainer will lead to a restart of the Pod. More info:
<a href="https://kubernetes.io/docs/concepts/workloads/pods/init-containers/">https://kubernetes.io/docs/concepts/workloads/pods/init-containers/</a>
InitContainers described here modify an operator generated init
containers if they share the same name and modifications are done via a
strategic merge patch.</p>
<p>The names of init container name managed by the operator are:
* <code>init-config-reloader</code>.</p>
<p>Overriding init containers is entirely outside the scope of what the
maintainers will support and by doing so, you accept that this behaviour
may break at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>additionalScrapeConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AdditionalScrapeConfigs allows specifying a key of a Secret containing
additional Prometheus scrape configurations. Scrape configurations
specified are appended to the configurations generated by the Prometheus
Operator. Job configurations specified must have the form as specified
in the official Prometheus documentation:
<a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config</a>.
As scrape configs are appended, the user is responsible to make sure it
is valid. Note that using this feature may expose the possibility to
break upgrades of Prometheus. It is advised to review Prometheus release
notes to ensure that no incompatible scrape configs are going to break
Prometheus after the upgrade.</p>
</td>
</tr>
<tr>
<td>
<code>apiserverConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.APIServerConfig">
Monitoring v1.APIServerConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>APIServerConfig allows specifying a host and auth methods to access the
Kuberntees API server.
If null, Prometheus is assumed to run inside of the cluster: it will
discover the API servers automatically and use the Pod&rsquo;s CA certificate
and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/.</p>
</td>
</tr>
<tr>
<td>
<code>priorityClassName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Priority class assigned to the Pods.</p>
</td>
</tr>
<tr>
<td>
<code>portName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Port name used for the pods and governing service.
Default: &ldquo;web&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>arbitraryFSAccessThroughSMs</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ArbitraryFSAccessThroughSMsConfig">
Monitoring v1.ArbitraryFSAccessThroughSMsConfig
</a>
</em>
</td>
<td>
<p>When true, ServiceMonitor, PodMonitor and Probe object are forbidden to
reference arbitrary files on the file system of the &lsquo;prometheus&rsquo;
container.
When a ServiceMonitor&rsquo;s endpoint specifies a <code>bearerTokenFile</code> value
(e.g.  &lsquo;/var/run/secrets/kubernetes.io/serviceaccount/token&rsquo;), a
malicious target can get access to the Prometheus service account&rsquo;s
token in the Prometheus&rsquo; scrape request. Setting
<code>spec.arbitraryFSAccessThroughSM</code> to &lsquo;true&rsquo; would prevent the attack.
Users should instead provide the credentials using the
<code>spec.bearerTokenSecret</code> field.</p>
</td>
</tr>
<tr>
<td>
<code>overrideHonorLabels</code><br/>
<em>
bool
</em>
</td>
<td>
<p>When true, Prometheus resolves label conflicts by renaming the labels in the scraped data
to “exported_” for all targets created from ServiceMonitor, PodMonitor and
ScrapeConfig objects. Otherwise the HonorLabels field of the service or pod monitor applies.
In practice,<code>overrideHonorLaels:true</code> enforces <code>honorLabels:false</code>
for all ServiceMonitor, PodMonitor and ScrapeConfig objects.</p>
</td>
</tr>
<tr>
<td>
<code>overrideHonorTimestamps</code><br/>
<em>
bool
</em>
</td>
<td>
<p>When true, Prometheus ignores the timestamps for all the targets created
from service and pod monitors.
Otherwise the HonorTimestamps field of the service or pod monitor applies.</p>
</td>
</tr>
<tr>
<td>
<code>ignoreNamespaceSelectors</code><br/>
<em>
bool
</em>
</td>
<td>
<p>When true, <code>spec.namespaceSelector</code> from all PodMonitor, ServiceMonitor
and Probe objects will be ignored. They will only discover targets
within the namespace of the PodMonitor, ServiceMonitor and Probe
object.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedNamespaceLabel</code><br/>
<em>
string
</em>
</td>
<td>
<p>When not empty, a label will be added to:</p>
<ol>
<li>All metrics scraped from <code>ServiceMonitor</code>, <code>PodMonitor</code>, <code>Probe</code> and <code>ScrapeConfig</code> objects.</li>
<li>All metrics generated from recording rules defined in <code>PrometheusRule</code> objects.</li>
<li>All alerts generated from alerting rules defined in <code>PrometheusRule</code> objects.</li>
<li>All vector selectors of PromQL expressions defined in <code>PrometheusRule</code> objects.</li>
</ol>
<p>The label will not added for objects referenced in <code>spec.excludedFromEnforcement</code>.</p>
<p>The label&rsquo;s name is this field&rsquo;s value.
The label&rsquo;s value is the namespace of the <code>ServiceMonitor</code>,
<code>PodMonitor</code>, <code>Probe</code>, <code>PrometheusRule</code> or <code>ScrapeConfig</code> object.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedSampleLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>When defined, enforcedSampleLimit specifies a global limit on the number
of scraped samples that will be accepted. This overrides any
<code>spec.sampleLimit</code> set by ServiceMonitor, PodMonitor, Probe objects
unless <code>spec.sampleLimit</code> is greater than zero and less than
<code>spec.enforcedSampleLimit</code>.</p>
<p>It is meant to be used by admins to keep the overall number of
samples/series under a desired limit.</p>
<p>When both <code>enforcedSampleLimit</code> and <code>sampleLimit</code> are defined and greater than zero, the following rules apply:
* Scrape objects without a defined sampleLimit value will inherit the global sampleLimit value (Prometheus &gt;= 2.45.0) or the enforcedSampleLimit value (Prometheus &lt; v2.45.0).
If Prometheus version is &gt;= 2.45.0 and the <code>enforcedSampleLimit</code> is greater than the <code>sampleLimit</code>, the <code>sampleLimit</code> will be set to <code>enforcedSampleLimit</code>.
* Scrape objects with a sampleLimit value less than or equal to enforcedSampleLimit keep their specific value.
* Scrape objects with a sampleLimit value greater than enforcedSampleLimit are set to enforcedSampleLimit.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedTargetLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>When defined, enforcedTargetLimit specifies a global limit on the number
of scraped targets. The value overrides any <code>spec.targetLimit</code> set by
ServiceMonitor, PodMonitor, Probe objects unless <code>spec.targetLimit</code> is
greater than zero and less than <code>spec.enforcedTargetLimit</code>.</p>
<p>It is meant to be used by admins to to keep the overall number of
targets under a desired limit.</p>
<p>When both <code>enforcedTargetLimit</code> and <code>targetLimit</code> are defined and greater than zero, the following rules apply:
* Scrape objects without a defined targetLimit value will inherit the global targetLimit value (Prometheus &gt;= 2.45.0) or the enforcedTargetLimit value (Prometheus &lt; v2.45.0).
If Prometheus version is &gt;= 2.45.0 and the <code>enforcedTargetLimit</code> is greater than the <code>targetLimit</code>, the <code>targetLimit</code> will be set to <code>enforcedTargetLimit</code>.
* Scrape objects with a targetLimit value less than or equal to enforcedTargetLimit keep their specific value.
* Scrape objects with a targetLimit value greater than enforcedTargetLimit are set to enforcedTargetLimit.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedLabelLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>When defined, enforcedLabelLimit specifies a global limit on the number
of labels per sample. The value overrides any <code>spec.labelLimit</code> set by
ServiceMonitor, PodMonitor, Probe objects unless <code>spec.labelLimit</code> is
greater than zero and less than <code>spec.enforcedLabelLimit</code>.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
<p>When both <code>enforcedLabelLimit</code> and <code>labelLimit</code> are defined and greater than zero, the following rules apply:
* Scrape objects without a defined labelLimit value will inherit the global labelLimit value (Prometheus &gt;= 2.45.0) or the enforcedLabelLimit value (Prometheus &lt; v2.45.0).
If Prometheus version is &gt;= 2.45.0 and the <code>enforcedLabelLimit</code> is greater than the <code>labelLimit</code>, the <code>labelLimit</code> will be set to <code>enforcedLabelLimit</code>.
* Scrape objects with a labelLimit value less than or equal to enforcedLabelLimit keep their specific value.
* Scrape objects with a labelLimit value greater than enforcedLabelLimit are set to enforcedLabelLimit.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedLabelNameLengthLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>When defined, enforcedLabelNameLengthLimit specifies a global limit on the length
of labels name per sample. The value overrides any <code>spec.labelNameLengthLimit</code> set by
ServiceMonitor, PodMonitor, Probe objects unless <code>spec.labelNameLengthLimit</code> is
greater than zero and less than <code>spec.enforcedLabelNameLengthLimit</code>.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
<p>When both <code>enforcedLabelNameLengthLimit</code> and <code>labelNameLengthLimit</code> are defined and greater than zero, the following rules apply:
* Scrape objects without a defined labelNameLengthLimit value will inherit the global labelNameLengthLimit value (Prometheus &gt;= 2.45.0) or the enforcedLabelNameLengthLimit value (Prometheus &lt; v2.45.0).
If Prometheus version is &gt;= 2.45.0 and the <code>enforcedLabelNameLengthLimit</code> is greater than the <code>labelNameLengthLimit</code>, the <code>labelNameLengthLimit</code> will be set to <code>enforcedLabelNameLengthLimit</code>.
* Scrape objects with a labelNameLengthLimit value less than or equal to enforcedLabelNameLengthLimit keep their specific value.
* Scrape objects with a labelNameLengthLimit value greater than enforcedLabelNameLengthLimit are set to enforcedLabelNameLengthLimit.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedLabelValueLengthLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>When not null, enforcedLabelValueLengthLimit defines a global limit on the length
of labels value per sample. The value overrides any <code>spec.labelValueLengthLimit</code> set by
ServiceMonitor, PodMonitor, Probe objects unless <code>spec.labelValueLengthLimit</code> is
greater than zero and less than <code>spec.enforcedLabelValueLengthLimit</code>.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
<p>When both <code>enforcedLabelValueLengthLimit</code> and <code>labelValueLengthLimit</code> are defined and greater than zero, the following rules apply:
* Scrape objects without a defined labelValueLengthLimit value will inherit the global labelValueLengthLimit value (Prometheus &gt;= 2.45.0) or the enforcedLabelValueLengthLimit value (Prometheus &lt; v2.45.0).
If Prometheus version is &gt;= 2.45.0 and the <code>enforcedLabelValueLengthLimit</code> is greater than the <code>labelValueLengthLimit</code>, the <code>labelValueLengthLimit</code> will be set to <code>enforcedLabelValueLengthLimit</code>.
* Scrape objects with a labelValueLengthLimit value less than or equal to enforcedLabelValueLengthLimit keep their specific value.
* Scrape objects with a labelValueLengthLimit value greater than enforcedLabelValueLengthLimit are set to enforcedLabelValueLengthLimit.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedKeepDroppedTargets</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>When defined, enforcedKeepDroppedTargets specifies a global limit on the number of targets
dropped by relabeling that will be kept in memory. The value overrides
any <code>spec.keepDroppedTargets</code> set by
ServiceMonitor, PodMonitor, Probe objects unless <code>spec.keepDroppedTargets</code> is
greater than zero and less than <code>spec.enforcedKeepDroppedTargets</code>.</p>
<p>It requires Prometheus &gt;= v2.47.0.</p>
<p>When both <code>enforcedKeepDroppedTargets</code> and <code>keepDroppedTargets</code> are defined and greater than zero, the following rules apply:
* Scrape objects without a defined keepDroppedTargets value will inherit the global keepDroppedTargets value (Prometheus &gt;= 2.45.0) or the enforcedKeepDroppedTargets value (Prometheus &lt; v2.45.0).
If Prometheus version is &gt;= 2.45.0 and the <code>enforcedKeepDroppedTargets</code> is greater than the <code>keepDroppedTargets</code>, the <code>keepDroppedTargets</code> will be set to <code>enforcedKeepDroppedTargets</code>.
* Scrape objects with a keepDroppedTargets value less than or equal to enforcedKeepDroppedTargets keep their specific value.
* Scrape objects with a keepDroppedTargets value greater than enforcedKeepDroppedTargets are set to enforcedKeepDroppedTargets.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedBodySizeLimit</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ByteSize">
Monitoring v1.ByteSize
</a>
</em>
</td>
<td>
<p>When defined, enforcedBodySizeLimit specifies a global limit on the size
of uncompressed response body that will be accepted by Prometheus.
Targets responding with a body larger than this many bytes will cause
the scrape to fail.</p>
<p>It requires Prometheus &gt;= v2.28.0.</p>
<p>When both <code>enforcedBodySizeLimit</code> and <code>bodySizeLimit</code> are defined and greater than zero, the following rules apply:
* Scrape objects without a defined bodySizeLimit value will inherit the global bodySizeLimit value (Prometheus &gt;= 2.45.0) or the enforcedBodySizeLimit value (Prometheus &lt; v2.45.0).
If Prometheus version is &gt;= 2.45.0 and the <code>enforcedBodySizeLimit</code> is greater than the <code>bodySizeLimit</code>, the <code>bodySizeLimit</code> will be set to <code>enforcedBodySizeLimit</code>.
* Scrape objects with a bodySizeLimit value less than or equal to enforcedBodySizeLimit keep their specific value.
* Scrape objects with a bodySizeLimit value greater than enforcedBodySizeLimit are set to enforcedBodySizeLimit.</p>
</td>
</tr>
<tr>
<td>
<code>nameValidationScheme</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.NameValidationSchemeOptions">
Monitoring v1.NameValidationSchemeOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the validation scheme for metric and label names.</p>
<p>It requires Prometheus &gt;= v2.55.0.</p>
</td>
</tr>
<tr>
<td>
<code>nameEscapingScheme</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.NameEscapingSchemeOptions">
Monitoring v1.NameEscapingSchemeOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the character escaping scheme that will be requested when scraping
for metric and label names that do not conform to the legacy Prometheus
character set.</p>
<p>It requires Prometheus &gt;= v3.4.0.</p>
</td>
</tr>
<tr>
<td>
<code>convertClassicHistogramsToNHCB</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to convert all scraped classic histograms into a native
histogram with custom buckets.</p>
<p>It requires Prometheus &gt;= v3.4.0.</p>
</td>
</tr>
<tr>
<td>
<code>minReadySeconds</code><br/>
<em>
uint32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Minimum number of seconds for which a newly created Pod should be ready
without any of its container crashing for it to be considered available.
Defaults to 0 (pod will be considered available as soon as it is ready)</p>
<p>This is an alpha field from kubernetes 1.22 until 1.24 which requires
enabling the StatefulSetMinReadySeconds feature gate.</p>
</td>
</tr>
<tr>
<td>
<code>hostAliases</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.HostAlias">
[]Monitoring v1.HostAlias
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional list of hosts and IPs that will be injected into the Pod&rsquo;s
hosts file if specified.</p>
</td>
</tr>
<tr>
<td>
<code>additionalArgs</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Argument">
[]Monitoring v1.Argument
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AdditionalArgs allows setting additional arguments for the &lsquo;prometheus&rsquo; container.</p>
<p>It is intended for e.g. activating hidden flags which are not supported by
the dedicated configuration options yet. The arguments are passed as-is to the
Prometheus container which may cause issues if they are invalid or not supported
by the given Prometheus version.</p>
<p>In case of an argument conflict (e.g. an argument which is already set by the
operator itself) or when providing an invalid argument, the reconciliation will
fail and an error will be logged.</p>
</td>
</tr>
<tr>
<td>
<code>walCompression</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures compression of the write-ahead log (WAL) using Snappy.</p>
<p>WAL compression is enabled by default for Prometheus &gt;= 2.20.0</p>
<p>Requires Prometheus v2.11.0 and above.</p>
</td>
</tr>
<tr>
<td>
<code>excludedFromEnforcement</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ObjectReference">
[]Monitoring v1.ObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of references to PodMonitor, ServiceMonitor, Probe and PrometheusRule objects
to be excluded from enforcing a namespace label of origin.</p>
<p>It is only applicable if <code>spec.enforcedNamespaceLabel</code> set to true.</p>
</td>
</tr>
<tr>
<td>
<code>hostNetwork</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Use the host&rsquo;s network namespace if true.</p>
<p>Make sure to understand the security implications if you want to enable
it (<a href="https://kubernetes.io/docs/concepts/configuration/overview/">https://kubernetes.io/docs/concepts/configuration/overview/</a> ).</p>
<p>When hostNetwork is enabled, this will set the DNS policy to
<code>ClusterFirstWithHostNet</code> automatically (unless <code>.spec.DNSPolicy</code> is set
to a different value).</p>
</td>
</tr>
<tr>
<td>
<code>podTargetLabels</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodTargetLabels are appended to the <code>spec.podTargetLabels</code> field of all
PodMonitor and ServiceMonitor objects.</p>
</td>
</tr>
<tr>
<td>
<code>tracingConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.PrometheusTracingConfig">
Monitoring v1.PrometheusTracingConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TracingConfig configures tracing in Prometheus.</p>
<p>This is an <em>experimental feature</em>, it may change in any upcoming release
in a breaking way.</p>
</td>
</tr>
<tr>
<td>
<code>bodySizeLimit</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ByteSize">
Monitoring v1.ByteSize
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>BodySizeLimit defines per-scrape on response body size.
Only valid in Prometheus versions 2.45.0 and newer.</p>
<p>Note that the global limit only applies to scrape objects that don&rsquo;t specify an explicit limit value.
If you want to enforce a maximum limit for all scrape objects, refer to enforcedBodySizeLimit.</p>
</td>
</tr>
<tr>
<td>
<code>sampleLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.
Only valid in Prometheus versions 2.45.0 and newer.</p>
<p>Note that the global limit only applies to scrape objects that don&rsquo;t specify an explicit limit value.
If you want to enforce a maximum limit for all scrape objects, refer to enforcedSampleLimit.</p>
</td>
</tr>
<tr>
<td>
<code>targetLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>TargetLimit defines a limit on the number of scraped targets that will be accepted.
Only valid in Prometheus versions 2.45.0 and newer.</p>
<p>Note that the global limit only applies to scrape objects that don&rsquo;t specify an explicit limit value.
If you want to enforce a maximum limit for all scrape objects, refer to enforcedTargetLimit.</p>
</td>
</tr>
<tr>
<td>
<code>labelLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on number of labels that will be accepted for a sample.
Only valid in Prometheus versions 2.45.0 and newer.</p>
<p>Note that the global limit only applies to scrape objects that don&rsquo;t specify an explicit limit value.
If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelLimit.</p>
</td>
</tr>
<tr>
<td>
<code>labelNameLengthLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on length of labels name that will be accepted for a sample.
Only valid in Prometheus versions 2.45.0 and newer.</p>
<p>Note that the global limit only applies to scrape objects that don&rsquo;t specify an explicit limit value.
If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelNameLengthLimit.</p>
</td>
</tr>
<tr>
<td>
<code>labelValueLengthLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on length of labels value that will be accepted for a sample.
Only valid in Prometheus versions 2.45.0 and newer.</p>
<p>Note that the global limit only applies to scrape objects that don&rsquo;t specify an explicit limit value.
If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelValueLengthLimit.</p>
</td>
</tr>
<tr>
<td>
<code>keepDroppedTargets</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on the number of targets dropped by relabeling
that will be kept in memory. 0 means no limit.</p>
<p>It requires Prometheus &gt;= v2.47.0.</p>
<p>Note that the global limit only applies to scrape objects that don&rsquo;t specify an explicit limit value.
If you want to enforce a maximum limit for all scrape objects, refer to enforcedKeepDroppedTargets.</p>
</td>
</tr>
<tr>
<td>
<code>reloadStrategy</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ReloadStrategyType">
Monitoring v1.ReloadStrategyType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the strategy used to reload the Prometheus configuration.
If not specified, the configuration is reloaded using the /-/reload HTTP endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>maximumStartupDurationSeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the maximum time that the <code>prometheus</code> container&rsquo;s startup probe will wait before being considered failed. The startup probe will return success after the WAL replay is complete.
If set, the value should be greater than 60 (seconds). Otherwise it will be equal to 600 seconds (15 minutes).</p>
</td>
</tr>
<tr>
<td>
<code>scrapeClasses</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ScrapeClass">
[]Monitoring v1.ScrapeClass
</a>
</em>
</td>
<td>
<p>List of scrape classes to expose to scraping objects such as
PodMonitors, ServiceMonitors, Probes and ScrapeConfigs.</p>
<p>This is an <em>experimental feature</em>, it may change in any upcoming release
in a breaking way.</p>
</td>
</tr>
<tr>
<td>
<code>serviceDiscoveryRole</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ServiceDiscoveryRole">
Monitoring v1.ServiceDiscoveryRole
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the service discovery role used to discover targets from
<code>ServiceMonitor</code> objects and Alertmanager endpoints.</p>
<p>If set, the value should be either &ldquo;Endpoints&rdquo; or &ldquo;EndpointSlice&rdquo;.
If unset, the operator assumes the &ldquo;Endpoints&rdquo; role.</p>
</td>
</tr>
<tr>
<td>
<code>tsdb</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.TSDBSpec">
Monitoring v1.TSDBSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the runtime reloadable configuration of the timeseries database(TSDB).
It requires Prometheus &gt;= v2.39.0 or PrometheusAgent &gt;= v2.54.0.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeFailureLogFile</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>File to which scrape failures are logged.
Reloading the configuration will reopen the file.</p>
<p>If the filename has an empty path, e.g. &lsquo;file.log&rsquo;, The Prometheus Pods
will mount the file into an emptyDir volume at <code>/var/log/prometheus</code>.
If a full path is provided, e.g. &lsquo;/var/log/prometheus/file.log&rsquo;, you
must mount a volume in the specified directory and it must be writable.
It requires Prometheus &gt;= v2.55.0.</p>
</td>
</tr>
<tr>
<td>
<code>serviceName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The name of the service name used by the underlying StatefulSet(s) as the governing service.
If defined, the Service  must be created before the Prometheus/PrometheusAgent resource in the same namespace and it must define a selector that matches the pod labels.
If empty, the operator will create and manage a headless service named <code>prometheus-operated</code> for Prometheus resources,
or <code>prometheus-agent-operated</code> for PrometheusAgent resources.
When deploying multiple Prometheus/PrometheusAgent resources in the same namespace, it is recommended to specify a different value for each.
See <a href="https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id">https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id</a> for more details.</p>
</td>
</tr>
<tr>
<td>
<code>runtime</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.RuntimeConfig">
Monitoring v1.RuntimeConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RuntimeConfig configures the values for the Prometheus process behavior</p>
</td>
</tr>
<tr>
<td>
<code>terminationGracePeriodSeconds</code><br/>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional duration in seconds the pod needs to terminate gracefully.
Value must be non-negative integer. The value zero indicates stop immediately via
the kill signal (no opportunity to shut down) which may lead to data corruption.</p>
<p>Defaults to 600 seconds.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.PuppetDBSDConfig">PuppetDBSDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>PuppetDBSDConfig configurations allow retrieving scrape targets from PuppetDB resources.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#puppetdb_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#puppetdb_sd_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>url</code><br/>
<em>
string
</em>
</td>
<td>
<p>The URL of the PuppetDB root query endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>query</code><br/>
<em>
string
</em>
</td>
<td>
<p>Puppet Query Language (PQL) query. Only resources are supported.
<a href="https://puppet.com/docs/puppetdb/latest/api/query/v4/pql.html">https://puppet.com/docs/puppetdb/latest/api/query/v4/pql.html</a></p>
</td>
</tr>
<tr>
<td>
<code>includeParameters</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to include the parameters as meta labels.
Note: Enabling this exposes parameters in the Prometheus UI and API. Make sure
that you don&rsquo;t have secrets exposed as parameters if you enable this.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Refresh interval to re-read the list of resources.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Port to scrape the metrics from.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional HTTP basic authentication information.
Cannot be set at the same time as <code>authorization</code>, or <code>oauth2</code>.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional <code>authorization</code> HTTP header configuration.
Cannot be set at the same time as <code>basicAuth</code>, or <code>oauth2</code>.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional OAuth2.0 configuration.
Cannot be set at the same time as <code>basicAuth</code>, or <code>authorization</code>.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration to connect to the Puppet DB.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether the HTTP requests should follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether to enable HTTP2.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.PushoverConfig">PushoverConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Receiver">Receiver</a>)
</p>
<div>
<p>PushoverConfig configures notifications via Pushover.
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#pushover_config">https://prometheus.io/docs/alerting/latest/configuration/#pushover_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>sendResolved</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether or not to notify about resolved alerts.</p>
</td>
</tr>
<tr>
<td>
<code>userKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The secret&rsquo;s key that contains the recipient user&rsquo;s user key.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.
Either <code>userKey</code> or <code>userKeyFile</code> is required.</p>
</td>
</tr>
<tr>
<td>
<code>userKeyFile</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The user key file that contains the recipient user&rsquo;s user key.
Either <code>userKey</code> or <code>userKeyFile</code> is required.
It requires Alertmanager &gt;= v0.26.0.</p>
</td>
</tr>
<tr>
<td>
<code>token</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The secret&rsquo;s key that contains the registered application&rsquo;s API token, see <a href="https://pushover.net/apps">https://pushover.net/apps</a>.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.
Either <code>token</code> or <code>tokenFile</code> is required.</p>
</td>
</tr>
<tr>
<td>
<code>tokenFile</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The token file that contains the registered application&rsquo;s API token, see <a href="https://pushover.net/apps">https://pushover.net/apps</a>.
Either <code>token</code> or <code>tokenFile</code> is required.
It requires Alertmanager &gt;= v0.26.0.</p>
</td>
</tr>
<tr>
<td>
<code>title</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Notification title.</p>
</td>
</tr>
<tr>
<td>
<code>message</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Notification message.</p>
</td>
</tr>
<tr>
<td>
<code>url</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>A supplementary URL shown alongside the message.</p>
</td>
</tr>
<tr>
<td>
<code>urlTitle</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>A title for supplementary URL, otherwise just the URL is shown</p>
</td>
</tr>
<tr>
<td>
<code>ttl</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The time to live definition for the alert notification</p>
</td>
</tr>
<tr>
<td>
<code>device</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The name of a device to send the notification to</p>
</td>
</tr>
<tr>
<td>
<code>sound</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The name of one of the sounds supported by device clients to override the user&rsquo;s default sound choice</p>
</td>
</tr>
<tr>
<td>
<code>priority</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Priority, see <a href="https://pushover.net/api#priority">https://pushover.net/api#priority</a></p>
</td>
</tr>
<tr>
<td>
<code>retry</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>How often the Pushover servers will send the same notification to the user.
Must be at least 30 seconds.</p>
</td>
</tr>
<tr>
<td>
<code>expire</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>How long your notification will continue to be retried for, unless the user
acknowledges the notification.</p>
</td>
</tr>
<tr>
<td>
<code>html</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether notification message is HTML or plain text.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">
HTTPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTP client configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.Receiver">Receiver
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.AlertmanagerConfigSpec">AlertmanagerConfigSpec</a>)
</p>
<div>
<p>Receiver defines one or more notification integrations.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the receiver. Must be unique across all items from the list.</p>
</td>
</tr>
<tr>
<td>
<code>opsgenieConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.OpsGenieConfig">
[]OpsGenieConfig
</a>
</em>
</td>
<td>
<p>List of OpsGenie configurations.</p>
</td>
</tr>
<tr>
<td>
<code>pagerdutyConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.PagerDutyConfig">
[]PagerDutyConfig
</a>
</em>
</td>
<td>
<p>List of PagerDuty configurations.</p>
</td>
</tr>
<tr>
<td>
<code>discordConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.DiscordConfig">
[]DiscordConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of Discord configurations.</p>
</td>
</tr>
<tr>
<td>
<code>slackConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.SlackConfig">
[]SlackConfig
</a>
</em>
</td>
<td>
<p>List of Slack configurations.</p>
</td>
</tr>
<tr>
<td>
<code>webhookConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.WebhookConfig">
[]WebhookConfig
</a>
</em>
</td>
<td>
<p>List of webhook configurations.</p>
</td>
</tr>
<tr>
<td>
<code>wechatConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.WeChatConfig">
[]WeChatConfig
</a>
</em>
</td>
<td>
<p>List of WeChat configurations.</p>
</td>
</tr>
<tr>
<td>
<code>emailConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.EmailConfig">
[]EmailConfig
</a>
</em>
</td>
<td>
<p>List of Email configurations.</p>
</td>
</tr>
<tr>
<td>
<code>victoropsConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.VictorOpsConfig">
[]VictorOpsConfig
</a>
</em>
</td>
<td>
<p>List of VictorOps configurations.</p>
</td>
</tr>
<tr>
<td>
<code>pushoverConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.PushoverConfig">
[]PushoverConfig
</a>
</em>
</td>
<td>
<p>List of Pushover configurations.</p>
</td>
</tr>
<tr>
<td>
<code>snsConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.SNSConfig">
[]SNSConfig
</a>
</em>
</td>
<td>
<p>List of SNS configurations</p>
</td>
</tr>
<tr>
<td>
<code>telegramConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.TelegramConfig">
[]TelegramConfig
</a>
</em>
</td>
<td>
<p>List of Telegram configurations.</p>
</td>
</tr>
<tr>
<td>
<code>webexConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.WebexConfig">
[]WebexConfig
</a>
</em>
</td>
<td>
<p>List of Webex configurations.</p>
</td>
</tr>
<tr>
<td>
<code>msteamsConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.MSTeamsConfig">
[]MSTeamsConfig
</a>
</em>
</td>
<td>
<p>List of MSTeams configurations.
It requires Alertmanager &gt;= 0.26.0.</p>
</td>
</tr>
<tr>
<td>
<code>msteamsv2Configs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.MSTeamsV2Config">
[]MSTeamsV2Config
</a>
</em>
</td>
<td>
<p>List of MSTeamsV2 configurations.
It requires Alertmanager &gt;= 0.28.0.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.Route">Route
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.AlertmanagerConfigSpec">AlertmanagerConfigSpec</a>)
</p>
<div>
<p>Route defines a node in the routing tree.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>receiver</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of the receiver for this route. If not empty, it should be listed in
the <code>receivers</code> field.</p>
</td>
</tr>
<tr>
<td>
<code>groupBy</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of labels to group by.
Labels must not be repeated (unique list).
Special label &ldquo;&hellip;&rdquo; (aggregate by all possible labels), if provided, must be the only element in the list.</p>
</td>
</tr>
<tr>
<td>
<code>groupWait</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>How long to wait before sending the initial notification.
Must match the regular expression<code>^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$</code>
Example: &ldquo;30s&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>groupInterval</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>How long to wait before sending an updated notification.
Must match the regular expression<code>^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$</code>
Example: &ldquo;5m&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>repeatInterval</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>How long to wait before repeating the last notification.
Must match the regular expression<code>^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$</code>
Example: &ldquo;4h&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>matchers</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.Matcher">
[]Matcher
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of matchers that the alert&rsquo;s labels should match. For the first
level route, the operator removes any existing equality and regexp
matcher on the <code>namespace</code> label and adds a <code>namespace: &lt;object
namespace&gt;</code> matcher.</p>
</td>
</tr>
<tr>
<td>
<code>continue</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Boolean indicating whether an alert should continue matching subsequent
sibling nodes. It will always be overridden to true for the first-level
route by the Prometheus operator.</p>
</td>
</tr>
<tr>
<td>
<code>routes</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON">
[]k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1.JSON
</a>
</em>
</td>
<td>
<p>Child routes.</p>
</td>
</tr>
<tr>
<td>
<code>muteTimeIntervals</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Note: this comment applies to the field definition above but appears
below otherwise it gets included in the generated manifest.
CRD schema doesn&rsquo;t support self-referential types for now (see
<a href="https://github.com/kubernetes/kubernetes/issues/62872)">https://github.com/kubernetes/kubernetes/issues/62872)</a>. We have to use
an alternative type to circumvent the limitation. The downside is that
the Kube API can&rsquo;t validate the data beyond the fact that it is a valid
JSON representation.
MuteTimeIntervals is a list of MuteTimeInterval names that will mute this route when matched,</p>
</td>
</tr>
<tr>
<td>
<code>activeTimeIntervals</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ActiveTimeIntervals is a list of MuteTimeInterval names when this route should be active.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.SDFile">SDFile
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.FileSDConfig">FileSDConfig</a>)
</p>
<div>
<p>SDFile represents a file used for service discovery</p>
</div>
<h3 id="monitoring.coreos.com/v1alpha1.SNSConfig">SNSConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Receiver">Receiver</a>)
</p>
<div>
<p>SNSConfig configures notifications via AWS SNS.
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#sns_configs">https://prometheus.io/docs/alerting/latest/configuration/#sns_configs</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>sendResolved</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether or not to notify about resolved alerts.</p>
</td>
</tr>
<tr>
<td>
<code>apiURL</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The SNS API URL i.e. <a href="https://sns.us-east-2.amazonaws.com">https://sns.us-east-2.amazonaws.com</a>.
If not specified, the SNS API URL from the SNS SDK will be used.</p>
</td>
</tr>
<tr>
<td>
<code>sigv4</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Sigv4">
Monitoring v1.Sigv4
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures AWS&rsquo;s Signature Verification 4 signing process to sign requests.</p>
</td>
</tr>
<tr>
<td>
<code>topicARN</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SNS topic ARN, i.e. arn:aws:sns:us-east-2:698519295917:My-Topic
If you don&rsquo;t specify this value, you must specify a value for the PhoneNumber or TargetARN.</p>
</td>
</tr>
<tr>
<td>
<code>subject</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Subject line when the message is delivered to email endpoints.</p>
</td>
</tr>
<tr>
<td>
<code>phoneNumber</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Phone number if message is delivered via SMS in E.164 format.
If you don&rsquo;t specify this value, you must specify a value for the TopicARN or TargetARN.</p>
</td>
</tr>
<tr>
<td>
<code>targetARN</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The  mobile platform endpoint ARN if message is delivered via mobile notifications.
If you don&rsquo;t specify this value, you must specify a value for the topic_arn or PhoneNumber.</p>
</td>
</tr>
<tr>
<td>
<code>message</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The message content of the SNS notification.</p>
</td>
</tr>
<tr>
<td>
<code>attributes</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SNS message attributes.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">
HTTPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTP client configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.ScalewayRole">ScalewayRole
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScalewaySDConfig">ScalewaySDConfig</a>)
</p>
<div>
<p>Role of the targets to retrieve. Must be <code>Instance</code> or <code>Baremetal</code>.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Baremetal&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Instance&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.ScalewaySDConfig">ScalewaySDConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>ScalewaySDConfig configurations allow retrieving scrape targets from Scaleway instances and baremetal services.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scaleway_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scaleway_sd_config</a>
TODO: Need to document that we will not be supporting the <code>_file</code> fields.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>accessKey</code><br/>
<em>
string
</em>
</td>
<td>
<p>Access key to use. <a href="https://console.scaleway.com/project/credentials">https://console.scaleway.com/project/credentials</a></p>
</td>
</tr>
<tr>
<td>
<code>secretKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Secret key to use when listing targets.</p>
</td>
</tr>
<tr>
<td>
<code>projectID</code><br/>
<em>
string
</em>
</td>
<td>
<p>Project ID of the targets.</p>
</td>
</tr>
<tr>
<td>
<code>role</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.ScalewayRole">
ScalewayRole
</a>
</em>
</td>
<td>
<p>Service of the targets to retrieve. Must be <code>Instance</code> or <code>Baremetal</code>.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The port to scrape metrics from.</p>
</td>
</tr>
<tr>
<td>
<code>apiURL</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>API URL to use when doing the server listing requests.</p>
</td>
</tr>
<tr>
<td>
<code>zone</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Zone is the availability zone of your targets (e.g. fr-par-1).</p>
</td>
</tr>
<tr>
<td>
<code>nameFilter</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NameFilter specify a name filter (works as a LIKE) to apply on the server listing request.</p>
</td>
</tr>
<tr>
<td>
<code>tagsFilter</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>TagsFilter specify a tag filter (a server needs to have all defined tags to be listed) to apply on the server listing request.</p>
</td>
</tr>
<tr>
<td>
<code>refreshInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Refresh interval to re-read the list of instances.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration to use on every scrape request</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfig">ScrapeConfig</a>)
</p>
<div>
<p>ScrapeConfigSpec is a specification of the desired configuration for a scrape configuration.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>jobName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The value of the <code>job</code> label assigned to the scraped metrics by default.</p>
<p>The <code>job_name</code> field in the rendered scrape configuration is always controlled by the
operator to prevent duplicate job names, which Prometheus does not allow. Instead the
<code>job</code> label is set by means of relabeling configs.</p>
</td>
</tr>
<tr>
<td>
<code>staticConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.StaticConfig">
[]StaticConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StaticConfigs defines a list of static targets with a common label set.</p>
</td>
</tr>
<tr>
<td>
<code>fileSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.FileSDConfig">
[]FileSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>FileSDConfigs defines a list of file service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>httpSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPSDConfig">
[]HTTPSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTPSDConfigs defines a list of HTTP service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>kubernetesSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.KubernetesSDConfig">
[]KubernetesSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>KubernetesSDConfigs defines a list of Kubernetes service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>consulSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.ConsulSDConfig">
[]ConsulSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ConsulSDConfigs defines a list of Consul service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>dnsSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.DNSSDConfig">
[]DNSSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DNSSDConfigs defines a list of DNS service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>ec2SDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.EC2SDConfig">
[]EC2SDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>EC2SDConfigs defines a list of EC2 service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>azureSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.AzureSDConfig">
[]AzureSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AzureSDConfigs defines a list of Azure service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>gceSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.GCESDConfig">
[]GCESDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>GCESDConfigs defines a list of GCE service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>openstackSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.OpenStackSDConfig">
[]OpenStackSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>OpenStackSDConfigs defines a list of OpenStack service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>digitalOceanSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.DigitalOceanSDConfig">
[]DigitalOceanSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DigitalOceanSDConfigs defines a list of DigitalOcean service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>kumaSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.KumaSDConfig">
[]KumaSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>KumaSDConfigs defines a list of Kuma service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>eurekaSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.EurekaSDConfig">
[]EurekaSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>EurekaSDConfigs defines a list of Eureka service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>dockerSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.DockerSDConfig">
[]DockerSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DockerSDConfigs defines a list of Docker service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>linodeSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.LinodeSDConfig">
[]LinodeSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>LinodeSDConfigs defines a list of Linode service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>hetznerSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HetznerSDConfig">
[]HetznerSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HetznerSDConfigs defines a list of Hetzner service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>nomadSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.NomadSDConfig">
[]NomadSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NomadSDConfigs defines a list of Nomad service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>dockerSwarmSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.DockerSwarmSDConfig">
[]DockerSwarmSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DockerswarmSDConfigs defines a list of Dockerswarm service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>puppetDBSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.PuppetDBSDConfig">
[]PuppetDBSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PuppetDBSDConfigs defines a list of PuppetDB service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>lightSailSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.LightSailSDConfig">
[]LightSailSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>LightsailSDConfigs defines a list of Lightsail service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>ovhcloudSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.OVHCloudSDConfig">
[]OVHCloudSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>OVHCloudSDConfigs defines a list of OVHcloud service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>scalewaySDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.ScalewaySDConfig">
[]ScalewaySDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ScalewaySDConfigs defines a list of Scaleway instances and baremetal service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>ionosSDConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.IonosSDConfig">
[]IonosSDConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>IonosSDConfigs defines a list of IONOS service discovery configurations.</p>
</td>
</tr>
<tr>
<td>
<code>relabelings</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.RelabelConfig">
[]Monitoring v1.RelabelConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RelabelConfigs defines how to rewrite the target&rsquo;s labels before scraping.
Prometheus Operator automatically adds relabelings for a few standard Kubernetes fields.
The original scrape job&rsquo;s name is available via the <code>__tmp_prometheus_job_name</code> label.
More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config</a></p>
</td>
</tr>
<tr>
<td>
<code>metricsPath</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>MetricsPath HTTP path to scrape for metrics. If empty, Prometheus uses the default value (e.g. /metrics).</p>
</td>
</tr>
<tr>
<td>
<code>scrapeInterval</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ScrapeInterval is the interval between consecutive scrapes.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeTimeout</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ScrapeTimeout is the number of seconds to wait until a scrape request times out.
The value cannot be greater than the scrape interval otherwise the operator will reject the resource.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeProtocols</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ScrapeProtocol">
[]Monitoring v1.ScrapeProtocol
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The protocols to negotiate during a scrape. It tells clients the
protocols supported by Prometheus in order of preference (from most to least preferred).</p>
<p>If unset, Prometheus uses its default value.</p>
<p>It requires Prometheus &gt;= v2.49.0.</p>
</td>
</tr>
<tr>
<td>
<code>fallbackScrapeProtocol</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.ScrapeProtocol">
Monitoring v1.ScrapeProtocol
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The protocol to use if a scrape returns blank, unparseable, or otherwise invalid Content-Type.</p>
<p>It requires Prometheus &gt;= v3.0.0.</p>
</td>
</tr>
<tr>
<td>
<code>honorTimestamps</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data.</p>
</td>
</tr>
<tr>
<td>
<code>trackTimestampsStaleness</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>TrackTimestampsStaleness whether Prometheus tracks staleness of
the metrics that have an explicit timestamp present in scraped data.
Has no effect if <code>honorTimestamps</code> is false.
It requires Prometheus &gt;= v2.48.0.</p>
</td>
</tr>
<tr>
<td>
<code>honorLabels</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>HonorLabels chooses the metric&rsquo;s labels on collisions with target labels.</p>
</td>
</tr>
<tr>
<td>
<code>params</code><br/>
<em>
map[string][]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional HTTP URL parameters</p>
</td>
</tr>
<tr>
<td>
<code>scheme</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures the protocol scheme used for requests.
If empty, Prometheus uses HTTP by default.</p>
</td>
</tr>
<tr>
<td>
<code>enableCompression</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>When false, Prometheus will request uncompressed response from the scraped target.</p>
<p>It requires Prometheus &gt;= v2.49.0.</p>
<p>If unset, Prometheus uses true by default.</p>
</td>
</tr>
<tr>
<td>
<code>enableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.BasicAuth">
Monitoring v1.BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>BasicAuth information to use on every scrape request.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeAuthorization">
Monitoring v1.SafeAuthorization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Authorization header to use on every scrape request.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.OAuth2">
Monitoring v1.OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>OAuth2 configuration to use on every scrape request.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.SafeTLSConfig">
Monitoring v1.SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration to use on every scrape request</p>
</td>
</tr>
<tr>
<td>
<code>sampleLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.</p>
</td>
</tr>
<tr>
<td>
<code>targetLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>TargetLimit defines a limit on the number of scraped targets that will be accepted.</p>
</td>
</tr>
<tr>
<td>
<code>labelLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on number of labels that will be accepted for a sample.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>labelNameLengthLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on length of labels name that will be accepted for a sample.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>labelValueLengthLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on length of labels value that will be accepted for a sample.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeClassicHistograms</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to scrape a classic histogram that is also exposed as a native histogram.
It requires Prometheus &gt;= v2.45.0.</p>
</td>
</tr>
<tr>
<td>
<code>nativeHistogramBucketLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>If there are more than this many buckets in a native histogram,
buckets will be merged to stay within the limit.
It requires Prometheus &gt;= v2.45.0.</p>
</td>
</tr>
<tr>
<td>
<code>nativeHistogramMinBucketFactor</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/api/resource#Quantity">
k8s.io/apimachinery/pkg/api/resource.Quantity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If the growth factor of one bucket to the next is smaller than this,
buckets will be merged to increase the factor sufficiently.
It requires Prometheus &gt;= v2.50.0.</p>
</td>
</tr>
<tr>
<td>
<code>convertClassicHistogramsToNHCB</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to convert all scraped classic histograms into a native histogram with custom buckets.
It requires Prometheus &gt;= v3.0.0.</p>
</td>
</tr>
<tr>
<td>
<code>keepDroppedTargets</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on the number of targets dropped by relabeling
that will be kept in memory. 0 means no limit.</p>
<p>It requires Prometheus &gt;= v2.47.0.</p>
</td>
</tr>
<tr>
<td>
<code>metricRelabelings</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.RelabelConfig">
[]Monitoring v1.RelabelConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>MetricRelabelConfigs to apply to samples before ingestion.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>proxyURL</code> defines the HTTP proxy server to use.</p>
</td>
</tr>
<tr>
<td>
<code>noProxy</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>noProxy</code> is a comma-separated string that can contain IPs, CIDR notation, domain names
that should be excluded from proxying. IP and domain names can
contain port numbers.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyFromEnvironment</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the proxy configuration defined by environment variables (HTTP_PROXY, HTTPS_PROXY, and NO_PROXY).</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>proxyConnectHeader</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
map[string][]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ProxyConnectHeader optionally specifies headers to send to
proxies during CONNECT requests.</p>
<p>It requires Prometheus &gt;= v2.43.0, Alertmanager &gt;= v0.25.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
<tr>
<td>
<code>nameValidationScheme</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.NameValidationSchemeOptions">
Monitoring v1.NameValidationSchemeOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the validation scheme for metric and label names.</p>
<p>It requires Prometheus &gt;= v3.0.0.</p>
</td>
</tr>
<tr>
<td>
<code>nameEscapingScheme</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.NameEscapingSchemeOptions">
Monitoring v1.NameEscapingSchemeOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Metric name escaping mode to request through content negotiation.</p>
<p>It requires Prometheus &gt;= v3.4.0.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeClass</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The scrape class to apply.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.SlackAction">SlackAction
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.SlackConfig">SlackConfig</a>)
</p>
<div>
<p>SlackAction configures a single Slack action that is sent with each
notification.
See <a href="https://api.slack.com/docs/message-attachments#action_fields">https://api.slack.com/docs/message-attachments#action_fields</a> and
<a href="https://api.slack.com/docs/message-buttons">https://api.slack.com/docs/message-buttons</a> for more information.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>text</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>url</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>style</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>value</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>confirm</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.SlackConfirmationField">
SlackConfirmationField
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.SlackConfig">SlackConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Receiver">Receiver</a>)
</p>
<div>
<p>SlackConfig configures notifications via Slack.
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#slack_config">https://prometheus.io/docs/alerting/latest/configuration/#slack_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>sendResolved</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether or not to notify about resolved alerts.</p>
</td>
</tr>
<tr>
<td>
<code>apiURL</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The secret&rsquo;s key that contains the Slack webhook URL.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>channel</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The channel or user to send notifications to.</p>
</td>
</tr>
<tr>
<td>
<code>username</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>color</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>title</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>titleLink</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>pretext</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>text</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>fields</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.SlackField">
[]SlackField
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>A list of Slack fields that are sent with each notification.</p>
</td>
</tr>
<tr>
<td>
<code>shortFields</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>footer</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>fallback</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>callbackId</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>iconEmoji</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>iconURL</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>imageURL</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>thumbURL</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>linkNames</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>mrkdwnIn</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>actions</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.SlackAction">
[]SlackAction
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>A list of Slack actions that are sent with each notification.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">
HTTPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTP client configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.SlackConfirmationField">SlackConfirmationField
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.SlackAction">SlackAction</a>)
</p>
<div>
<p>SlackConfirmationField protect users from destructive actions or
particularly distinguished decisions by asking them to confirm their button
click one more time.
See <a href="https://api.slack.com/docs/interactive-message-field-guide#confirmation_fields">https://api.slack.com/docs/interactive-message-field-guide#confirmation_fields</a>
for more information.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>text</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>title</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>okText</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>dismissText</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.SlackField">SlackField
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.SlackConfig">SlackConfig</a>)
</p>
<div>
<p>SlackField configures a single Slack field that is sent with each notification.
Each field must contain a title, value, and optionally, a boolean value to indicate if the field
is short enough to be displayed next to other fields designated as short.
See <a href="https://api.slack.com/docs/message-attachments#fields">https://api.slack.com/docs/message-attachments#fields</a> for more information.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>title</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>value</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>short</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.StaticConfig">StaticConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>StaticConfig defines a Prometheus static configuration.
See <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>targets</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.Target">
[]Target
</a>
</em>
</td>
<td>
<p>List of targets for this static configuration.</p>
</td>
</tr>
<tr>
<td>
<code>labels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Labels assigned to all metrics scraped from the targets.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.Target">Target
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.StaticConfig">StaticConfig</a>)
</p>
<div>
<p>Target represents a target for Prometheus to scrape
kubebuilder:validation:MinLength:=1</p>
</div>
<h3 id="monitoring.coreos.com/v1alpha1.TelegramConfig">TelegramConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Receiver">Receiver</a>)
</p>
<div>
<p>TelegramConfig configures notifications via Telegram.
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#telegram_config">https://prometheus.io/docs/alerting/latest/configuration/#telegram_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>sendResolved</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to notify about resolved alerts.</p>
</td>
</tr>
<tr>
<td>
<code>apiURL</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The Telegram API URL i.e. <a href="https://api.telegram.org">https://api.telegram.org</a>.
If not specified, default API URL will be used.</p>
</td>
</tr>
<tr>
<td>
<code>botToken</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Telegram bot token. It is mutually exclusive with <code>botTokenFile</code>.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
<p>Either <code>botToken</code> or <code>botTokenFile</code> is required.</p>
</td>
</tr>
<tr>
<td>
<code>botTokenFile</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>File to read the Telegram bot token from. It is mutually exclusive with <code>botToken</code>.
Either <code>botToken</code> or <code>botTokenFile</code> is required.</p>
<p>It requires Alertmanager &gt;= v0.26.0.</p>
</td>
</tr>
<tr>
<td>
<code>chatID</code><br/>
<em>
int64
</em>
</td>
<td>
<p>The Telegram chat ID.</p>
</td>
</tr>
<tr>
<td>
<code>messageThreadID</code><br/>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>The Telegram Group Topic ID.
It requires Alertmanager &gt;= 0.26.0.</p>
</td>
</tr>
<tr>
<td>
<code>message</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Message template</p>
</td>
</tr>
<tr>
<td>
<code>disableNotifications</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Disable telegram notifications</p>
</td>
</tr>
<tr>
<td>
<code>parseMode</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Parse mode for telegram message</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">
HTTPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTP client configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.Time">Time
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.TimeRange">TimeRange</a>)
</p>
<div>
<p>Time defines a time in 24hr format</p>
</div>
<h3 id="monitoring.coreos.com/v1alpha1.TimeInterval">TimeInterval
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.MuteTimeInterval">MuteTimeInterval</a>)
</p>
<div>
<p>TimeInterval describes intervals of time</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>times</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.TimeRange">
[]TimeRange
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Times is a list of TimeRange</p>
</td>
</tr>
<tr>
<td>
<code>weekdays</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.WeekdayRange">
[]WeekdayRange
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Weekdays is a list of WeekdayRange</p>
</td>
</tr>
<tr>
<td>
<code>daysOfMonth</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.DayOfMonthRange">
[]DayOfMonthRange
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DaysOfMonth is a list of DayOfMonthRange</p>
</td>
</tr>
<tr>
<td>
<code>months</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.MonthRange">
[]MonthRange
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Months is a list of MonthRange</p>
</td>
</tr>
<tr>
<td>
<code>years</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.YearRange">
[]YearRange
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Years is a list of YearRange</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.TimeRange">TimeRange
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.TimeInterval">TimeInterval</a>)
</p>
<div>
<p>TimeRange defines a start and end time in 24hr format</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>startTime</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.Time">
Time
</a>
</em>
</td>
<td>
<p>StartTime is the start time in 24hr format.</p>
</td>
</tr>
<tr>
<td>
<code>endTime</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.Time">
Time
</a>
</em>
</td>
<td>
<p>EndTime is the end time in 24hr format.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.URL">URL
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.DiscordConfig">DiscordConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.WebexConfig">WebexConfig</a>)
</p>
<div>
<p>URL represents a valid URL</p>
</div>
<h3 id="monitoring.coreos.com/v1alpha1.VictorOpsConfig">VictorOpsConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Receiver">Receiver</a>)
</p>
<div>
<p>VictorOpsConfig configures notifications via VictorOps.
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#victorops_config">https://prometheus.io/docs/alerting/latest/configuration/#victorops_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>sendResolved</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether or not to notify about resolved alerts.</p>
</td>
</tr>
<tr>
<td>
<code>apiKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The secret&rsquo;s key that contains the API key to use when talking to the VictorOps API.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>apiUrl</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The VictorOps API URL.</p>
</td>
</tr>
<tr>
<td>
<code>routingKey</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>A key used to map the alert to a team.</p>
</td>
</tr>
<tr>
<td>
<code>messageType</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Describes the behavior of the alert (CRITICAL, WARNING, INFO).</p>
</td>
</tr>
<tr>
<td>
<code>entityDisplayName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Contains summary of the alerted problem.</p>
</td>
</tr>
<tr>
<td>
<code>stateMessage</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Contains long explanation of the alerted problem.</p>
</td>
</tr>
<tr>
<td>
<code>monitoringTool</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The monitoring tool the state message is from.</p>
</td>
</tr>
<tr>
<td>
<code>customFields</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.KeyValue">
[]KeyValue
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Additional custom fields for notification.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">
HTTPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The HTTP client&rsquo;s configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.WeChatConfig">WeChatConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Receiver">Receiver</a>)
</p>
<div>
<p>WeChatConfig configures notifications via WeChat.
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#wechat_config">https://prometheus.io/docs/alerting/latest/configuration/#wechat_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>sendResolved</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether or not to notify about resolved alerts.</p>
</td>
</tr>
<tr>
<td>
<code>apiSecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The secret&rsquo;s key that contains the WeChat API key.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>apiURL</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The WeChat API URL.</p>
</td>
</tr>
<tr>
<td>
<code>corpID</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The corp id for authentication.</p>
</td>
</tr>
<tr>
<td>
<code>agentID</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>toUser</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>toParty</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>toTag</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>message</code><br/>
<em>
string
</em>
</td>
<td>
<p>API request data as defined by the WeChat API.</p>
</td>
</tr>
<tr>
<td>
<code>messageType</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">
HTTPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTP client configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.WebexConfig">WebexConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Receiver">Receiver</a>)
</p>
<div>
<p>WebexConfig configures notification via Cisco Webex
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#webex_config">https://prometheus.io/docs/alerting/latest/configuration/#webex_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>sendResolved</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to notify about resolved alerts.</p>
</td>
</tr>
<tr>
<td>
<code>apiURL</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.URL">
URL
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The Webex Teams API URL i.e. <a href="https://webexapis.com/v1/messages">https://webexapis.com/v1/messages</a>
Provide if different from the default API URL.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">
HTTPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The HTTP client&rsquo;s configuration.
You must supply the bot token via the <code>httpConfig.authorization</code> field.</p>
</td>
</tr>
<tr>
<td>
<code>message</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Message template</p>
</td>
</tr>
<tr>
<td>
<code>roomID</code><br/>
<em>
string
</em>
</td>
<td>
<p>ID of the Webex Teams room where to send the messages.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.WebhookConfig">WebhookConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.Receiver">Receiver</a>)
</p>
<div>
<p>WebhookConfig configures notifications via a generic receiver supporting the webhook payload.
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#webhook_config">https://prometheus.io/docs/alerting/latest/configuration/#webhook_config</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>sendResolved</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether or not to notify about resolved alerts.</p>
</td>
</tr>
<tr>
<td>
<code>url</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The URL to send HTTP POST requests to. <code>urlSecret</code> takes precedence over
<code>url</code>. One of <code>urlSecret</code> and <code>url</code> should be defined.</p>
</td>
</tr>
<tr>
<td>
<code>urlSecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The secret&rsquo;s key that contains the webhook URL to send HTTP requests to.
<code>urlSecret</code> takes precedence over <code>url</code>. One of <code>urlSecret</code> and <code>url</code>
should be defined.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">
HTTPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTP client configuration.</p>
</td>
</tr>
<tr>
<td>
<code>maxAlerts</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Maximum number of alerts to be sent per webhook message. When 0, all alerts are included.</p>
</td>
</tr>
<tr>
<td>
<code>timeout</code><br/>
<em>
<a href="../v1/api.md#monitoring.coreos.com/v1.Duration">
Monitoring v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The maximum time to wait for a webhook request to complete, before failing the
request and allowing it to be retried.
It requires Alertmanager &gt;= v0.28.0.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.Weekday">Weekday
(<code>string</code> alias)</h3>
<div>
<p>Weekday is day of the week</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;friday&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;monday&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;saturday&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;sunday&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;thursday&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;tuesday&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;wednesday&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1alpha1.WeekdayRange">WeekdayRange
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.TimeInterval">TimeInterval</a>)
</p>
<div>
<p>WeekdayRange is an inclusive range of days of the week beginning on Sunday
Days can be specified by name (e.g &lsquo;Sunday&rsquo;) or as an inclusive range (e.g &lsquo;Monday:Friday&rsquo;)</p>
</div>
<h3 id="monitoring.coreos.com/v1alpha1.YearRange">YearRange
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.TimeInterval">TimeInterval</a>)
</p>
<div>
<p>YearRange is an inclusive range of years</p>
</div>
<hr/>
