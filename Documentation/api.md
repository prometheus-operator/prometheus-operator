---
title: "API reference"
description: "Prometheus operator generated API reference docs"
draft: false
images: []
menu:
docs:
parent: "operator"
weight: 208
toc: true
---
> This page is automatically generated with `gen-crd-api-reference-docs`.
<p>Packages:</p>
<ul>
<li>
<a href="#monitoring.coreos.com%2fv1">monitoring.coreos.com/v1</a>
</li>
<li>
<a href="#monitoring.coreos.com%2fv1alpha1">monitoring.coreos.com/v1alpha1</a>
</li>
<li>
<a href="#monitoring.coreos.com%2fv1beta1">monitoring.coreos.com/v1beta1</a>
</li>
</ul>
<h2 id="monitoring.coreos.com/v1">monitoring.coreos.com/v1</h2>
Resource Types:
<ul><li>
<a href="#monitoring.coreos.com/v1.Alertmanager">Alertmanager</a>
</li><li>
<a href="#monitoring.coreos.com/v1.PodMonitor">PodMonitor</a>
</li><li>
<a href="#monitoring.coreos.com/v1.Probe">Probe</a>
</li><li>
<a href="#monitoring.coreos.com/v1.Prometheus">Prometheus</a>
</li><li>
<a href="#monitoring.coreos.com/v1.PrometheusRule">PrometheusRule</a>
</li><li>
<a href="#monitoring.coreos.com/v1.ServiceMonitor">ServiceMonitor</a>
</li></ul>
<h3 id="monitoring.coreos.com/v1.Alertmanager">Alertmanager
</h3>
<div>
<p>Alertmanager describes an Alertmanager cluster.</p>
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
monitoring.coreos.com/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>Alertmanager</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
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
<a href="#monitoring.coreos.com/v1.AlertmanagerSpec">
AlertmanagerSpec
</a>
</em>
</td>
<td>
<p>Specification of the desired behavior of the Alertmanager cluster. More info:
<a href="https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status">https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status</a></p>
<br/>
<br/>
<table>
<tr>
<td>
<code>podMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
EmbeddedObjectMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata configures Labels and Annotations which are propagated to the alertmanager pods.</p>
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
<p>Image if specified has precedence over baseImage, tag and sha
combinations. Specifying the version is still necessary to ensure the
Prometheus Operator knows what version of Alertmanager is being
configured.</p>
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
<p>Version the cluster should be on.</p>
</td>
</tr>
<tr>
<td>
<code>tag</code><br/>
<em>
string
</em>
</td>
<td>
<p>Tag of Alertmanager container image to be deployed. Defaults to the value of <code>version</code>.
Version is ignored if Tag is set.
Deprecated: use &lsquo;image&rsquo; instead.  The image tag can be specified
as part of the image URL.</p>
</td>
</tr>
<tr>
<td>
<code>sha</code><br/>
<em>
string
</em>
</td>
<td>
<p>SHA of Alertmanager container image to be deployed. Defaults to the value of <code>version</code>.
Similar to a tag, but the SHA explicitly deploys an immutable container image.
Version and Tag are ignored if SHA is set.
Deprecated: use &lsquo;image&rsquo; instead.  The image digest can be specified
as part of the image URL.</p>
</td>
</tr>
<tr>
<td>
<code>baseImage</code><br/>
<em>
string
</em>
</td>
<td>
<p>Base image that is used to deploy pods, without tag.
Deprecated: use &lsquo;image&rsquo; instead</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>An optional list of references to secrets in the same namespace
to use for pulling prometheus and alertmanager images from registries
see <a href="http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod">http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod</a></p>
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
<p>Secrets is a list of Secrets in the same namespace as the Alertmanager
object, which shall be mounted into the Alertmanager Pods.
The Secrets are mounted into /etc/alertmanager/secrets/<secret-name>.</p>
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
<p>ConfigMaps is a list of ConfigMaps in the same namespace as the Alertmanager
object, which shall be mounted into the Alertmanager Pods.
The ConfigMaps are mounted into /etc/alertmanager/configmaps/<configmap-name>.</p>
</td>
</tr>
<tr>
<td>
<code>configSecret</code><br/>
<em>
string
</em>
</td>
<td>
<p>ConfigSecret is the name of a Kubernetes Secret in the same namespace as the
Alertmanager object, which contains the configuration for this Alertmanager
instance. If empty, it defaults to &lsquo;alertmanager-<alertmanager-name>&rsquo;.</p>
<p>The Alertmanager configuration should be available under the
<code>alertmanager.yaml</code> key. Additional keys from the original secret are
copied to the generated secret.</p>
<p>If either the secret or the <code>alertmanager.yaml</code> key is missing, the
operator provisions an Alertmanager configuration with one empty
receiver (effectively dropping alert notifications).</p>
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
<p>Log level for Alertmanager to be configured with.</p>
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
<p>Log format for Alertmanager to be configured with.</p>
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
<p>Size is the expected size of the alertmanager cluster. The controller will
eventually make the size of the running cluster equal to the expected
size.</p>
</td>
</tr>
<tr>
<td>
<code>retention</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.GoDuration">
GoDuration
</a>
</em>
</td>
<td>
<p>Time duration Alertmanager shall retain data for. Default is &lsquo;120h&rsquo;,
and must match the regular expression <code>[0-9]+(ms|s|m|h)</code> (milliseconds seconds minutes hours).</p>
</td>
</tr>
<tr>
<td>
<code>storage</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.StorageSpec">
StorageSpec
</a>
</em>
</td>
<td>
<p>Storage is the definition of how storage will be used by the Alertmanager
instances.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<p>Volumes allows configuration of additional volumes on the output StatefulSet definition.
Volumes specified will be appended to other volumes that are generated as a result of
StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<p>VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
VolumeMounts specified will be appended to other VolumeMounts in the alertmanager container,
that are generated as a result of StorageSpec objects.</p>
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
<p>The external URL the Alertmanager instances will be available under. This is
necessary to generate correct URLs. This is necessary if Alertmanager is not
served from root of a DNS name.</p>
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
<p>The route prefix Alertmanager registers HTTP handlers for. This is useful,
if using ExternalURL and a proxy is rewriting HTTP routes of a request,
and the actual ExternalURL is still true, but the server serves requests
under a different route prefix. For example for use with <code>kubectl proxy</code>.</p>
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
<p>If set to true all actions on the underlying managed objects are not
goint to be performed, except for delete actions.</p>
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
<p>Define which Nodes the Pods are scheduled on.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>Define resources requests and limits for single Pods.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s scheduling constraints.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s topology spread constraints.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<p>SecurityContext holds pod-level security attributes and common container settings.
This defaults to the default PodSecurityContext.</p>
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
<code>listenLocal</code><br/>
<em>
bool
</em>
</td>
<td>
<p>ListenLocal makes the Alertmanager server listen on loopback, so that it
does not bind against the Pod IP. Note this is only for the Alertmanager
UI, not the gossip communication.</p>
</td>
</tr>
<tr>
<td>
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>Containers allows injecting additional containers. This is meant to
allow adding an authentication proxy to an Alertmanager pod.
Containers described here modify an operator generated container if they
share the same name and modifications are done via a strategic merge
patch. The current container names are: <code>alertmanager</code> and
<code>config-reloader</code>. Overriding containers is entirely outside the scope
of what the maintainers will support and by doing so, you accept that
this behaviour may break at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>initContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
fetch secrets for injection into the Alertmanager configuration from external sources. Any
errors during the execution of an initContainer will lead to a restart of the Pod. More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/init-containers/">https://kubernetes.io/docs/concepts/workloads/pods/init-containers/</a>
Using initContainers for any use case other then secret fetching is entirely outside the scope
of what the maintainers will support and by doing so, you accept that this behaviour may break
at any time without notice.</p>
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
<p>Priority class assigned to the Pods</p>
</td>
</tr>
<tr>
<td>
<code>additionalPeers</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>AdditionalPeers allows injecting a set of additional Alertmanagers to peer with to form a highly available cluster.</p>
</td>
</tr>
<tr>
<td>
<code>clusterAdvertiseAddress</code><br/>
<em>
string
</em>
</td>
<td>
<p>ClusterAdvertiseAddress is the explicit address to advertise in cluster.
Needs to be provided for non RFC1918 <a href="public">1</a> addresses.
[1] RFC1918: <a href="https://tools.ietf.org/html/rfc1918">https://tools.ietf.org/html/rfc1918</a></p>
</td>
</tr>
<tr>
<td>
<code>clusterGossipInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.GoDuration">
GoDuration
</a>
</em>
</td>
<td>
<p>Interval between gossip attempts.</p>
</td>
</tr>
<tr>
<td>
<code>clusterPushpullInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.GoDuration">
GoDuration
</a>
</em>
</td>
<td>
<p>Interval between pushpull attempts.</p>
</td>
</tr>
<tr>
<td>
<code>clusterPeerTimeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.GoDuration">
GoDuration
</a>
</em>
</td>
<td>
<p>Timeout for cluster peering.</p>
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
This defaults to web</p>
</td>
</tr>
<tr>
<td>
<code>forceEnableClusterMode</code><br/>
<em>
bool
</em>
</td>
<td>
<p>ForceEnableClusterMode ensures Alertmanager does not deactivate the cluster mode when running with a single replica.
Use case is e.g. spanning an Alertmanager cluster across Kubernetes clusters with a single replica in each.</p>
</td>
</tr>
<tr>
<td>
<code>alertmanagerConfigSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>AlertmanagerConfigs to be selected for to merge and configure Alertmanager with.</p>
</td>
</tr>
<tr>
<td>
<code>alertmanagerConfigNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to be selected for AlertmanagerConfig discovery. If nil, only
check own namespace.</p>
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
<p>Minimum number of seconds for which a newly created pod should be ready
without any of its container crashing for it to be considered available.
Defaults to 0 (pod will be considered available as soon as it is ready)
This is an alpha field and requires enabling StatefulSetMinReadySeconds feature gate.</p>
</td>
</tr>
<tr>
<td>
<code>hostAliases</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.HostAlias">
[]HostAlias
</a>
</em>
</td>
<td>
<p>Pods&rsquo; hostAliases configuration</p>
</td>
</tr>
<tr>
<td>
<code>web</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertmanagerWebSpec">
AlertmanagerWebSpec
</a>
</em>
</td>
<td>
<p>Defines the web command line flags when starting Alertmanager.</p>
</td>
</tr>
<tr>
<td>
<code>alertmanagerConfiguration</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertmanagerConfiguration">
AlertmanagerConfiguration
</a>
</em>
</td>
<td>
<p>EXPERIMENTAL: alertmanagerConfiguration specifies the configuration of Alertmanager.
If defined, it takes precedence over the <code>configSecret</code> field.
This field may change in future releases.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertmanagerStatus">
AlertmanagerStatus
</a>
</em>
</td>
<td>
<p>Most recent observed status of the Alertmanager cluster. Read-only. Not
included when requesting from the apiserver, only from the Prometheus
Operator API itself. More info:
<a href="https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status">https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.PodMonitor">PodMonitor
</h3>
<div>
<p>PodMonitor defines monitoring for a set of pods.</p>
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
monitoring.coreos.com/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>PodMonitor</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
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
<a href="#monitoring.coreos.com/v1.PodMonitorSpec">
PodMonitorSpec
</a>
</em>
</td>
<td>
<p>Specification of desired Pod selection for target discovery by Prometheus.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>jobLabel</code><br/>
<em>
string
</em>
</td>
<td>
<p>The label to use to retrieve the job name from.</p>
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
<p>PodTargetLabels transfers labels on the Kubernetes Pod onto the target.</p>
</td>
</tr>
<tr>
<td>
<code>podMetricsEndpoints</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">
[]PodMetricsEndpoint
</a>
</em>
</td>
<td>
<p>A list of endpoints allowed as part of this PodMonitor.</p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Selector to select Pod objects.</p>
</td>
</tr>
<tr>
<td>
<code>namespaceSelector</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.NamespaceSelector">
NamespaceSelector
</a>
</em>
</td>
<td>
<p>Selector to select which namespaces the Endpoints objects are discovered from.</p>
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
<p>Per-scrape limit on length of labels value that will be accepted for a sample.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>attachMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AttachMetadata">
AttachMetadata
</a>
</em>
</td>
<td>
<p>Attaches node metadata to discovered targets. Only valid for role: pod.
Only valid in Prometheus versions 2.35.0 and newer.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.Probe">Probe
</h3>
<div>
<p>Probe defines monitoring for a set of static targets or ingresses.</p>
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
monitoring.coreos.com/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>Probe</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
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
<a href="#monitoring.coreos.com/v1.ProbeSpec">
ProbeSpec
</a>
</em>
</td>
<td>
<p>Specification of desired Ingress selection for target discovery by Prometheus.</p>
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
<p>The job name assigned to scraped metrics by default.</p>
</td>
</tr>
<tr>
<td>
<code>prober</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ProberSpec">
ProberSpec
</a>
</em>
</td>
<td>
<p>Specification for the prober to use for probing targets.
The prober.URL parameter is required. Targets cannot be probed if left empty.</p>
</td>
</tr>
<tr>
<td>
<code>module</code><br/>
<em>
string
</em>
</td>
<td>
<p>The module to use for probing specifying how to probe the target.
Example module configuring in the blackbox exporter:
<a href="https://github.com/prometheus/blackbox_exporter/blob/master/example.yml">https://github.com/prometheus/blackbox_exporter/blob/master/example.yml</a></p>
</td>
</tr>
<tr>
<td>
<code>targets</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ProbeTargets">
ProbeTargets
</a>
</em>
</td>
<td>
<p>Targets defines a set of static or dynamically discovered targets to probe.</p>
</td>
</tr>
<tr>
<td>
<code>interval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Interval at which targets are probed using the configured prober.
If not specified Prometheus&rsquo; global scrape interval is used.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeTimeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Timeout for scraping metrics from the Prometheus exporter.
If not specified, the Prometheus global scrape interval is used.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ProbeTLSConfig">
ProbeTLSConfig
</a>
</em>
</td>
<td>
<p>TLS configuration to use when scraping the endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>bearerTokenSecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Secret to mount to read bearer token for scraping targets. The secret
needs to be in the same namespace as the probe and accessible by
the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
</a>
</em>
</td>
<td>
<p>BasicAuth allow an endpoint to authenticate over basic authentication.
More info: <a href="https://prometheus.io/docs/operating/configuration/#endpoint">https://prometheus.io/docs/operating/configuration/#endpoint</a></p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
</a>
</em>
</td>
<td>
<p>OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>metricRelabelings</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RelabelConfig">
[]RelabelConfig
</a>
</em>
</td>
<td>
<p>MetricRelabelConfigs to apply to samples before ingestion.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
</a>
</em>
</td>
<td>
<p>Authorization section for this endpoint</p>
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
<p>Per-scrape limit on length of labels value that will be accepted for a sample.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.Prometheus">Prometheus
</h3>
<div>
<p>Prometheus defines a Prometheus deployment.</p>
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
monitoring.coreos.com/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>Prometheus</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
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
<a href="#monitoring.coreos.com/v1.PrometheusSpec">
PrometheusSpec
</a>
</em>
</td>
<td>
<p>Specification of the desired behavior of the Prometheus cluster. More info:
<a href="https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status">https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status</a></p>
<br/>
<br/>
<table>
<tr>
<td>
<code>podMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
EmbeddedObjectMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata configures Labels and Annotations which are propagated to the prometheus pods.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitorSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ServiceMonitors to be selected for target discovery. <em>Deprecated:</em> if
neither this nor podMonitorSelector are specified, configuration is
unmanaged.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitorNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespace&rsquo;s labels to match for ServiceMonitor discovery. If nil, only
check own namespace.</p>
</td>
</tr>
<tr>
<td>
<code>podMonitorSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> PodMonitors to be selected for target discovery.
<em>Deprecated:</em> if neither this nor serviceMonitorSelector are specified,
configuration is unmanaged.</p>
</td>
</tr>
<tr>
<td>
<code>podMonitorNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespace&rsquo;s labels to match for PodMonitor discovery. If nil, only
check own namespace.</p>
</td>
</tr>
<tr>
<td>
<code>probeSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> Probes to be selected for target discovery.</p>
</td>
</tr>
<tr>
<td>
<code>probeNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> Namespaces to be selected for Probe discovery. If nil, only check own namespace.</p>
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
<p>Version of Prometheus to be deployed.</p>
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
<p>Image if specified has precedence over baseImage, tag and sha
combinations. Specifying the version is still necessary to ensure the
Prometheus Operator knows what version of Prometheus is being
configured.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>An optional list of references to secrets in the same namespace
to use for pulling prometheus and alertmanager images from registries
see <a href="http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod">http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod</a></p>
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
<p>Number of replicas of each shard to deploy for a Prometheus deployment.
Number of replicas multiplied by shards is the total number of Pods
created.</p>
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
<p>EXPERIMENTAL: Number of shards to distribute targets onto. Number of
replicas multiplied by shards is the total number of Pods created. Note
that scaling down shards will not reshard data onto remaining instances,
it must be manually moved. Increasing shards will not reshard data
either but it will continue to be available from the same instances. To
query globally use Thanos sidecar and Thanos querier or remote write
data to a central location. Sharding is done on the content of the
<code>__address__</code> target meta-label.</p>
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
<p>Name of Prometheus external label used to denote replica name.
Defaults to the value of <code>prometheus_replica</code>. External label will
<em>not</em> be added when value is set to empty string (<code>&quot;&quot;</code>).</p>
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
<p>Name of Prometheus external label used to denote Prometheus instance
name. Defaults to the value of <code>prometheus</code>. External label will
<em>not</em> be added when value is set to empty string (<code>&quot;&quot;</code>).</p>
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
<p>Log level for Prometheus to be configured with.</p>
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
<p>Log format for Prometheus to be configured with.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Interval between consecutive scrapes. Default: <code>30s</code></p>
</td>
</tr>
<tr>
<td>
<code>scrapeTimeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Number of seconds to wait for target to respond before erroring.</p>
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
external systems (federation, remote storage, Alertmanager).</p>
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
<p>Enable Prometheus to be used as a receiver for the Prometheus remote write protocol. Defaults to the value of <code>false</code>.
WARNING: This is not considered an efficient way of ingesting samples.
Use it with caution for specific low-volume use cases.
It is not suitable for replacing the ingestion via scraping and turning
Prometheus into a push-based metrics collection system.
For more information see <a href="https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver">https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver</a>
Only valid in Prometheus versions 2.33.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>enableFeatures</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Enable access to Prometheus disabled features. By default, no features are enabled.
Enabling disabled features is entirely outside the scope of what the maintainers will
support and by doing so, you accept that this behaviour may break at any
time without notice.
For more information see <a href="https://prometheus.io/docs/prometheus/latest/disabled_features/">https://prometheus.io/docs/prometheus/latest/disabled_features/</a></p>
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
<p>The external URL the Prometheus instances will be available under. This is
necessary to generate correct URLs. This is necessary if Prometheus is not
served from root of a DNS name.</p>
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
<p>The route prefix Prometheus registers HTTP handlers for. This is useful,
if using ExternalURL and a proxy is rewriting HTTP routes of a request,
and the actual ExternalURL is still true, but the server serves requests
under a different route prefix. For example for use with <code>kubectl proxy</code>.</p>
</td>
</tr>
<tr>
<td>
<code>storage</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.StorageSpec">
StorageSpec
</a>
</em>
</td>
<td>
<p>Storage spec to specify how storage shall be used.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<p>Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will
be appended to other volumes that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<p>VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
VolumeMounts specified will be appended to other VolumeMounts in the prometheus container,
that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>web</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusWebSpec">
PrometheusWebSpec
</a>
</em>
</td>
<td>
<p>Defines the web command line flags when starting Prometheus.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>Define resources requests and limits for single Pods.</p>
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
<p>Define which Nodes the Pods are scheduled on.</p>
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
<code>secrets</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Secrets is a list of Secrets in the same namespace as the Prometheus
object, which shall be mounted into the Prometheus Pods.
The Secrets are mounted into /etc/prometheus/secrets/<secret-name>.</p>
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
The ConfigMaps are mounted into /etc/prometheus/configmaps/<configmap-name>.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s scheduling constraints.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s topology spread constraints.</p>
</td>
</tr>
<tr>
<td>
<code>remoteWrite</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RemoteWriteSpec">
[]RemoteWriteSpec
</a>
</em>
</td>
<td>
<p>remoteWrite is the list of remote write configurations.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<p>SecurityContext holds pod-level security attributes and common container settings.
This defaults to the default PodSecurityContext.</p>
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
<p>ListenLocal makes the Prometheus server listen on loopback, so that it
does not bind against the Pod IP.</p>
</td>
</tr>
<tr>
<td>
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>Containers allows injecting additional containers or modifying operator
generated containers. This can be used to allow adding an authentication
proxy to a Prometheus pod or to change the behavior of an operator
generated container. Containers described here modify an operator
generated container if they share the same name and modifications are
done via a strategic merge patch. The current container names are:
<code>prometheus</code>, <code>config-reloader</code>, and <code>thanos-sidecar</code>. Overriding
containers is entirely outside the scope of what the maintainers will
support and by doing so, you accept that this behaviour may break at any
time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>initContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
fetch secrets for injection into the Prometheus configuration from external sources. Any errors
during the execution of an initContainer will lead to a restart of the Pod. More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/init-containers/">https://kubernetes.io/docs/concepts/workloads/pods/init-containers/</a>
InitContainers described here modify an operator
generated init containers if they share the same name and modifications are
done via a strategic merge patch. The current init container name is:
<code>init-config-reloader</code>. Overriding init containers is entirely outside the
scope of what the maintainers will support and by doing so, you accept that
this behaviour may break at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>additionalScrapeConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
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
<a href="#monitoring.coreos.com/v1.APIServerConfig">
APIServerConfig
</a>
</em>
</td>
<td>
<p>APIServerConfig allows specifying a host and auth methods to access apiserver.
If left empty, Prometheus is assumed to run inside of the cluster
and will discover API servers automatically and use the pod&rsquo;s CA certificate
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
<p>Priority class assigned to the Pods</p>
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
This defaults to web</p>
</td>
</tr>
<tr>
<td>
<code>arbitraryFSAccessThroughSMs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ArbitraryFSAccessThroughSMsConfig">
ArbitraryFSAccessThroughSMsConfig
</a>
</em>
</td>
<td>
<p>ArbitraryFSAccessThroughSMs configures whether configuration
based on a service monitor can access arbitrary files on the file system
of the Prometheus container e.g. bearer token files.</p>
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
<p>When true, Prometheus resolves label conflicts by renaming the labels in
the scraped data to &ldquo;exported_<label value>&rdquo; for all targets created
from service and pod monitors.
Otherwise the HonorLabels field of the service or pod monitor applies.</p>
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
<p>IgnoreNamespaceSelectors if set to true will ignore NamespaceSelector
settings from all PodMonitor, ServiceMonitor and Probe objects. They will
only discover endpoints within the namespace of the PodMonitor,
ServiceMonitor and Probe objects.
Defaults to false.</p>
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
<p>EnforcedNamespaceLabel If set, a label will be added to</p>
<ol>
<li>all user-metrics (created by <code>ServiceMonitor</code>, <code>PodMonitor</code> and <code>Probe</code> objects) and</li>
<li>in all <code>PrometheusRule</code> objects (except the ones excluded in <code>prometheusRulesExcludedFromEnforce</code>) to
<ul>
<li>alerting &amp; recording rules and</li>
<li>the metrics used in their expressions (<code>expr</code>).</li>
</ul></li>
</ol>
<p>Label name is this field&rsquo;s value.
Label value is the namespace of the created object (mentioned above).</p>
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
<p>EnforcedSampleLimit defines global limit on number of scraped samples
that will be accepted. This overrides any SampleLimit set per
ServiceMonitor or/and PodMonitor. It is meant to be used by admins to
enforce the SampleLimit to keep overall number of samples/series under
the desired limit.
Note that if SampleLimit is lower that value will be taken instead.</p>
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
<p>EnforcedTargetLimit defines a global limit on the number of scraped
targets.  This overrides any TargetLimit set per ServiceMonitor or/and
PodMonitor.  It is meant to be used by admins to enforce the TargetLimit
to keep the overall number of targets under the desired limit.
Note that if TargetLimit is lower, that value will be taken instead,
except if either value is zero, in which case the non-zero value will be
used.  If both values are zero, no limit is enforced.</p>
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
<p>Per-scrape limit on number of labels that will be accepted for a sample. If
more than this number of labels are present post metric-relabeling, the
entire scrape will be treated as failed. 0 means no limit.
Only valid in Prometheus versions 2.27.0 and newer.</p>
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
<p>Per-scrape limit on length of labels name that will be accepted for a sample.
If a label name is longer than this number post metric-relabeling, the entire
scrape will be treated as failed. 0 means no limit.
Only valid in Prometheus versions 2.27.0 and newer.</p>
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
<p>Per-scrape limit on length of labels value that will be accepted for a sample.
If a label value is longer than this number post metric-relabeling, the
entire scrape will be treated as failed. 0 means no limit.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedBodySizeLimit</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
</a>
</em>
</td>
<td>
<p>EnforcedBodySizeLimit defines the maximum size of uncompressed response body
that will be accepted by Prometheus. Targets responding with a body larger than this many bytes
will cause the scrape to fail. Example: 100MB.
If defined, the limit will apply to all service/pod monitors and probes.
This is an experimental feature, this behaviour could
change or be removed in the future.
Only valid in Prometheus versions 2.28.0 and newer.</p>
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
<p>Minimum number of seconds for which a newly created pod should be ready
without any of its container crashing for it to be considered available.
Defaults to 0 (pod will be considered available as soon as it is ready)
This is an alpha field and requires enabling StatefulSetMinReadySeconds feature gate.</p>
</td>
</tr>
<tr>
<td>
<code>hostAliases</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.HostAlias">
[]HostAlias
</a>
</em>
</td>
<td>
<p>Pods&rsquo; hostAliases configuration</p>
</td>
</tr>
<tr>
<td>
<code>additionalArgs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Argument">
[]Argument
</a>
</em>
</td>
<td>
<p>AdditionalArgs allows setting additional arguments for the Prometheus container.
It is intended for e.g. activating hidden flags which are not supported by
the dedicated configuration options yet. The arguments are passed as-is to the
Prometheus container which may cause issues if they are invalid or not supporeted
by the given Prometheus version.
In case of an argument conflict (e.g. an argument which is already set by the
operator itself) or when providing an invalid argument the reconciliation will
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
<p>Enable compression of the write-ahead log using Snappy. This flag is
only available in versions of Prometheus &gt;= 2.11.0.</p>
</td>
</tr>
<tr>
<td>
<code>excludedFromEnforcement</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ObjectReference">
[]ObjectReference
</a>
</em>
</td>
<td>
<p>List of references to PodMonitor, ServiceMonitor, Probe and PrometheusRule objects
to be excluded from enforcing a namespace label of origin.
Applies only if enforcedNamespaceLabel set to true.</p>
</td>
</tr>
<tr>
<td>
<code>baseImage</code><br/>
<em>
string
</em>
</td>
<td>
<p>Base image to use for a Prometheus deployment.
Deprecated: use &lsquo;image&rsquo; instead</p>
</td>
</tr>
<tr>
<td>
<code>tag</code><br/>
<em>
string
</em>
</td>
<td>
<p>Tag of Prometheus container image to be deployed. Defaults to the value of <code>version</code>.
Version is ignored if Tag is set.
Deprecated: use &lsquo;image&rsquo; instead.  The image tag can be specified
as part of the image URL.</p>
</td>
</tr>
<tr>
<td>
<code>sha</code><br/>
<em>
string
</em>
</td>
<td>
<p>SHA of Prometheus container image to be deployed. Defaults to the value of <code>version</code>.
Similar to a tag, but the SHA explicitly deploys an immutable container image.
Version and Tag are ignored if SHA is set.
Deprecated: use &lsquo;image&rsquo; instead.  The image digest can be specified
as part of the image URL.</p>
</td>
</tr>
<tr>
<td>
<code>retention</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Time duration Prometheus shall retain data for. Default is &lsquo;24h&rsquo; if
retentionSize is not set, and must match the regular expression <code>[0-9]+(ms|s|m|h|d|w|y)</code>
(milliseconds seconds minutes hours days weeks years).</p>
</td>
</tr>
<tr>
<td>
<code>retentionSize</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
</a>
</em>
</td>
<td>
<p>Maximum amount of disk space used by blocks.</p>
</td>
</tr>
<tr>
<td>
<code>disableCompaction</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Disable prometheus compaction.</p>
</td>
</tr>
<tr>
<td>
<code>rules</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Rules">
Rules
</a>
</em>
</td>
<td>
<p>/&ndash;rules.*/ command-line arguments.</p>
</td>
</tr>
<tr>
<td>
<code>prometheusRulesExcludedFromEnforce</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusRuleExcludeConfig">
[]PrometheusRuleExcludeConfig
</a>
</em>
</td>
<td>
<p>PrometheusRulesExcludedFromEnforce - list of prometheus rules to be excluded from enforcing
of adding namespace labels. Works only if enforcedNamespaceLabel set to true.
Make sure both ruleNamespace and ruleName are set for each pair.
Deprecated: use excludedFromEnforcement instead.</p>
</td>
</tr>
<tr>
<td>
<code>query</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.QuerySpec">
QuerySpec
</a>
</em>
</td>
<td>
<p>QuerySpec defines the query command line flags when starting Prometheus.</p>
</td>
</tr>
<tr>
<td>
<code>ruleSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>A selector to select which PrometheusRules to mount for loading alerting/recording
rules from. Until (excluding) Prometheus Operator v0.24.0 Prometheus
Operator will migrate any legacy rule ConfigMaps to PrometheusRule custom
resources selected by RuleSelector. Make sure it does not match any config
maps that you do not want to be migrated.</p>
</td>
</tr>
<tr>
<td>
<code>ruleNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to be selected for PrometheusRules discovery. If unspecified, only
the same namespace as the Prometheus object is in is used.</p>
</td>
</tr>
<tr>
<td>
<code>alerting</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertingSpec">
AlertingSpec
</a>
</em>
</td>
<td>
<p>Define details regarding alerting.</p>
</td>
</tr>
<tr>
<td>
<code>remoteRead</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RemoteReadSpec">
[]RemoteReadSpec
</a>
</em>
</td>
<td>
<p>remoteRead is the list of remote read configurations.</p>
</td>
</tr>
<tr>
<td>
<code>additionalAlertRelabelConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>AdditionalAlertRelabelConfigs allows specifying a key of a Secret containing
additional Prometheus alert relabel configurations. Alert relabel configurations
specified are appended to the configurations generated by the Prometheus
Operator. Alert relabel configurations specified must have the form as specified
in the official Prometheus documentation:
<a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs</a>.
As alert relabel configs are appended, the user is responsible to make sure it
is valid. Note that using this feature may expose the possibility to
break upgrades of Prometheus. It is advised to review Prometheus release
notes to ensure that no incompatible alert relabel configs are going to break
Prometheus after the upgrade.</p>
</td>
</tr>
<tr>
<td>
<code>additionalAlertManagerConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>AdditionalAlertManagerConfigs allows specifying a key of a Secret containing
additional Prometheus AlertManager configurations. AlertManager configurations
specified are appended to the configurations generated by the Prometheus
Operator. Job configurations specified must have the form as specified
in the official Prometheus documentation:
<a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config</a>.
As AlertManager configs are appended, the user is responsible to make sure it
is valid. Note that using this feature may expose the possibility to
break upgrades of Prometheus. It is advised to review Prometheus release
notes to ensure that no incompatible AlertManager configs are going to break
Prometheus after the upgrade.</p>
</td>
</tr>
<tr>
<td>
<code>thanos</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ThanosSpec">
ThanosSpec
</a>
</em>
</td>
<td>
<p>Thanos configuration allows configuring various aspects of a Prometheus
server in a Thanos environment.</p>
<p>This section is experimental, it may change significantly without
deprecation notice in any release.</p>
<p>This is experimental and may change significantly without backward
compatibility in any release.</p>
</td>
</tr>
<tr>
<td>
<code>queryLogFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>QueryLogFile specifies the file to which PromQL queries are logged.
If the filename has an empty path, e.g. &lsquo;query.log&rsquo;, prometheus-operator will mount the file into an
emptyDir volume at <code>/var/log/prometheus</code>. If a full path is provided, e.g. /var/log/prometheus/query.log, you must mount a volume
in the specified directory and it must be writable. This is because the prometheus container runs with a read-only root filesystem for security reasons.
Alternatively, the location can be set to a stdout location such as <code>/dev/stdout</code> to log
query information to the default Prometheus log stream.
This is only available in versions of Prometheus &gt;= 2.16.0.
For more details, see the Prometheus docs (<a href="https://prometheus.io/docs/guides/query-log/">https://prometheus.io/docs/guides/query-log/</a>)</p>
</td>
</tr>
<tr>
<td>
<code>allowOverlappingBlocks</code><br/>
<em>
bool
</em>
</td>
<td>
<p>AllowOverlappingBlocks enables vertical compaction and vertical query merge in Prometheus.
This is still experimental in Prometheus so it may change in any upcoming release.</p>
</td>
</tr>
<tr>
<td>
<code>exemplars</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Exemplars">
Exemplars
</a>
</em>
</td>
<td>
<p>Exemplars related settings that are runtime reloadable.
It requires to enable the exemplar storage feature to be effective.</p>
</td>
</tr>
<tr>
<td>
<code>evaluationInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Interval between consecutive evaluations. Default: <code>30s</code></p>
</td>
</tr>
<tr>
<td>
<code>enableAdminAPI</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Enable access to prometheus web admin API. Defaults to the value of <code>false</code>.
WARNING: Enabling the admin APIs enables mutating endpoints, to delete data,
shutdown Prometheus, and more. Enabling this should be done with care and the
user is advised to add additional authentication authorization via a proxy to
ensure only clients authorized to perform these actions can do so.
For more information see <a href="https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis">https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis</a></p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusStatus">
PrometheusStatus
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
<h3 id="monitoring.coreos.com/v1.PrometheusRule">PrometheusRule
</h3>
<div>
<p>PrometheusRule defines recording and alerting rules for a Prometheus instance</p>
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
monitoring.coreos.com/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>PrometheusRule</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
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
<a href="#monitoring.coreos.com/v1.PrometheusRuleSpec">
PrometheusRuleSpec
</a>
</em>
</td>
<td>
<p>Specification of desired alerting rule definitions for Prometheus.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>groups</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RuleGroup">
[]RuleGroup
</a>
</em>
</td>
<td>
<p>Content of Prometheus rule file</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ServiceMonitor">ServiceMonitor
</h3>
<div>
<p>ServiceMonitor defines monitoring for a set of services.</p>
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
monitoring.coreos.com/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>ServiceMonitor</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
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
<a href="#monitoring.coreos.com/v1.ServiceMonitorSpec">
ServiceMonitorSpec
</a>
</em>
</td>
<td>
<p>Specification of desired Service selection for target discovery by
Prometheus.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>jobLabel</code><br/>
<em>
string
</em>
</td>
<td>
<p>JobLabel selects the label from the associated Kubernetes service which will be used as the <code>job</code> label for all metrics.</p>
<p>For example:
If in <code>ServiceMonitor.spec.jobLabel: foo</code> and in <code>Service.metadata.labels.foo: bar</code>,
then the <code>job=&quot;bar&quot;</code> label is added to all metrics.</p>
<p>If the value of this field is empty or if the label doesn&rsquo;t exist for the given Service, the <code>job</code> label of the metrics defaults to the name of the Kubernetes Service.</p>
</td>
</tr>
<tr>
<td>
<code>targetLabels</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>TargetLabels transfers labels from the Kubernetes <code>Service</code> onto the created metrics.</p>
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
<p>PodTargetLabels transfers labels on the Kubernetes <code>Pod</code> onto the created metrics.</p>
</td>
</tr>
<tr>
<td>
<code>endpoints</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Endpoint">
[]Endpoint
</a>
</em>
</td>
<td>
<p>A list of endpoints allowed as part of this ServiceMonitor.</p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Selector to select Endpoints objects.</p>
</td>
</tr>
<tr>
<td>
<code>namespaceSelector</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.NamespaceSelector">
NamespaceSelector
</a>
</em>
</td>
<td>
<p>Selector to select which namespaces the Kubernetes Endpoints objects are discovered from.</p>
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
<p>Per-scrape limit on length of labels value that will be accepted for a sample.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.APIServerConfig">APIServerConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>)
</p>
<div>
<p>APIServerConfig defines a host and auth methods to access apiserver.
More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config</a></p>
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
<p>Host of apiserver.
A valid string consisting of a hostname or IP followed by an optional port number</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
</a>
</em>
</td>
<td>
<p>BasicAuth allow an endpoint to authenticate over basic authentication</p>
</td>
</tr>
<tr>
<td>
<code>bearerToken</code><br/>
<em>
string
</em>
</td>
<td>
<p>Bearer token for accessing apiserver.</p>
</td>
</tr>
<tr>
<td>
<code>bearerTokenFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>File to read bearer token for accessing apiserver.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TLSConfig">
TLSConfig
</a>
</em>
</td>
<td>
<p>TLS Config to use for accessing apiserver.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Authorization">
Authorization
</a>
</em>
</td>
<td>
<p>Authorization section for accessing apiserver</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AlertingSpec">AlertingSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>)
</p>
<div>
<p>AlertingSpec defines parameters for alerting configuration of Prometheus servers.</p>
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
<code>alertmanagers</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">
[]AlertmanagerEndpoints
</a>
</em>
</td>
<td>
<p>AlertmanagerEndpoints Prometheus should fire alerts against.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AlertmanagerConfiguration">AlertmanagerConfiguration
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerSpec">AlertmanagerSpec</a>)
</p>
<div>
<p>AlertmanagerConfiguration defines the Alertmanager configuration.</p>
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
<p>The name of the AlertmanagerConfig resource which is used to generate the Alertmanager configuration.
It must be defined in the same namespace as the Alertmanager object.
The operator will not enforce a <code>namespace</code> label for routes and inhibition rules.</p>
</td>
</tr>
<tr>
<td>
<code>global</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertmanagerGlobalConfig">
AlertmanagerGlobalConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the global parameters of the Alertmanager configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertingSpec">AlertingSpec</a>)
</p>
<div>
<p>AlertmanagerEndpoints defines a selection of a single Endpoints object
containing alertmanager IPs to fire alerts against.</p>
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
<code>namespace</code><br/>
<em>
string
</em>
</td>
<td>
<p>Namespace of Endpoints object.</p>
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
<p>Name of Endpoints object in Namespace.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/util/intstr#IntOrString">
k8s.io/apimachinery/pkg/util/intstr.IntOrString
</a>
</em>
</td>
<td>
<p>Port the Alertmanager API is exposed on.</p>
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
<p>Scheme to use when firing alerts.</p>
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
<p>Prefix for the HTTP path alerts are pushed to.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TLSConfig">
TLSConfig
</a>
</em>
</td>
<td>
<p>TLS Config to use for alertmanager connection.</p>
</td>
</tr>
<tr>
<td>
<code>bearerTokenFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>BearerTokenFile to read from filesystem to use when authenticating to
Alertmanager.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
</a>
</em>
</td>
<td>
<p>Authorization section for this alertmanager endpoint</p>
</td>
</tr>
<tr>
<td>
<code>apiVersion</code><br/>
<em>
string
</em>
</td>
<td>
<p>Version of the Alertmanager API that Prometheus uses to send alerts. It
can be &ldquo;v1&rdquo; or &ldquo;v2&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>timeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Timeout is a per-target Alertmanager timeout when pushing alerts.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AlertmanagerGlobalConfig">AlertmanagerGlobalConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerConfiguration">AlertmanagerConfiguration</a>)
</p>
<div>
<p>AlertmanagerGlobalConfig configures parameters that are valid in all other configuration contexts.
See <a href="https://prometheus.io/docs/alerting/latest/configuration/#configuration-file">https://prometheus.io/docs/alerting/latest/configuration/#configuration-file</a></p>
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
<code>resolveTimeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>ResolveTimeout is the default value used by alertmanager if the alert does
not include EndsAt, after this time passes it can declare the alert as resolved if it has not been updated.
This has no impact on alerts from Prometheus, as they always include EndsAt.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.HTTPConfig">
HTTPConfig
</a>
</em>
</td>
<td>
<p>HTTP client configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AlertmanagerSpec">AlertmanagerSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.Alertmanager">Alertmanager</a>)
</p>
<div>
<p>AlertmanagerSpec is a specification of the desired behavior of the Alertmanager cluster. More info:
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
<code>podMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
EmbeddedObjectMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata configures Labels and Annotations which are propagated to the alertmanager pods.</p>
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
<p>Image if specified has precedence over baseImage, tag and sha
combinations. Specifying the version is still necessary to ensure the
Prometheus Operator knows what version of Alertmanager is being
configured.</p>
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
<p>Version the cluster should be on.</p>
</td>
</tr>
<tr>
<td>
<code>tag</code><br/>
<em>
string
</em>
</td>
<td>
<p>Tag of Alertmanager container image to be deployed. Defaults to the value of <code>version</code>.
Version is ignored if Tag is set.
Deprecated: use &lsquo;image&rsquo; instead.  The image tag can be specified
as part of the image URL.</p>
</td>
</tr>
<tr>
<td>
<code>sha</code><br/>
<em>
string
</em>
</td>
<td>
<p>SHA of Alertmanager container image to be deployed. Defaults to the value of <code>version</code>.
Similar to a tag, but the SHA explicitly deploys an immutable container image.
Version and Tag are ignored if SHA is set.
Deprecated: use &lsquo;image&rsquo; instead.  The image digest can be specified
as part of the image URL.</p>
</td>
</tr>
<tr>
<td>
<code>baseImage</code><br/>
<em>
string
</em>
</td>
<td>
<p>Base image that is used to deploy pods, without tag.
Deprecated: use &lsquo;image&rsquo; instead</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>An optional list of references to secrets in the same namespace
to use for pulling prometheus and alertmanager images from registries
see <a href="http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod">http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod</a></p>
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
<p>Secrets is a list of Secrets in the same namespace as the Alertmanager
object, which shall be mounted into the Alertmanager Pods.
The Secrets are mounted into /etc/alertmanager/secrets/<secret-name>.</p>
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
<p>ConfigMaps is a list of ConfigMaps in the same namespace as the Alertmanager
object, which shall be mounted into the Alertmanager Pods.
The ConfigMaps are mounted into /etc/alertmanager/configmaps/<configmap-name>.</p>
</td>
</tr>
<tr>
<td>
<code>configSecret</code><br/>
<em>
string
</em>
</td>
<td>
<p>ConfigSecret is the name of a Kubernetes Secret in the same namespace as the
Alertmanager object, which contains the configuration for this Alertmanager
instance. If empty, it defaults to &lsquo;alertmanager-<alertmanager-name>&rsquo;.</p>
<p>The Alertmanager configuration should be available under the
<code>alertmanager.yaml</code> key. Additional keys from the original secret are
copied to the generated secret.</p>
<p>If either the secret or the <code>alertmanager.yaml</code> key is missing, the
operator provisions an Alertmanager configuration with one empty
receiver (effectively dropping alert notifications).</p>
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
<p>Log level for Alertmanager to be configured with.</p>
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
<p>Log format for Alertmanager to be configured with.</p>
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
<p>Size is the expected size of the alertmanager cluster. The controller will
eventually make the size of the running cluster equal to the expected
size.</p>
</td>
</tr>
<tr>
<td>
<code>retention</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.GoDuration">
GoDuration
</a>
</em>
</td>
<td>
<p>Time duration Alertmanager shall retain data for. Default is &lsquo;120h&rsquo;,
and must match the regular expression <code>[0-9]+(ms|s|m|h)</code> (milliseconds seconds minutes hours).</p>
</td>
</tr>
<tr>
<td>
<code>storage</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.StorageSpec">
StorageSpec
</a>
</em>
</td>
<td>
<p>Storage is the definition of how storage will be used by the Alertmanager
instances.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<p>Volumes allows configuration of additional volumes on the output StatefulSet definition.
Volumes specified will be appended to other volumes that are generated as a result of
StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<p>VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
VolumeMounts specified will be appended to other VolumeMounts in the alertmanager container,
that are generated as a result of StorageSpec objects.</p>
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
<p>The external URL the Alertmanager instances will be available under. This is
necessary to generate correct URLs. This is necessary if Alertmanager is not
served from root of a DNS name.</p>
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
<p>The route prefix Alertmanager registers HTTP handlers for. This is useful,
if using ExternalURL and a proxy is rewriting HTTP routes of a request,
and the actual ExternalURL is still true, but the server serves requests
under a different route prefix. For example for use with <code>kubectl proxy</code>.</p>
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
<p>If set to true all actions on the underlying managed objects are not
goint to be performed, except for delete actions.</p>
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
<p>Define which Nodes the Pods are scheduled on.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>Define resources requests and limits for single Pods.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s scheduling constraints.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s topology spread constraints.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<p>SecurityContext holds pod-level security attributes and common container settings.
This defaults to the default PodSecurityContext.</p>
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
<code>listenLocal</code><br/>
<em>
bool
</em>
</td>
<td>
<p>ListenLocal makes the Alertmanager server listen on loopback, so that it
does not bind against the Pod IP. Note this is only for the Alertmanager
UI, not the gossip communication.</p>
</td>
</tr>
<tr>
<td>
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>Containers allows injecting additional containers. This is meant to
allow adding an authentication proxy to an Alertmanager pod.
Containers described here modify an operator generated container if they
share the same name and modifications are done via a strategic merge
patch. The current container names are: <code>alertmanager</code> and
<code>config-reloader</code>. Overriding containers is entirely outside the scope
of what the maintainers will support and by doing so, you accept that
this behaviour may break at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>initContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
fetch secrets for injection into the Alertmanager configuration from external sources. Any
errors during the execution of an initContainer will lead to a restart of the Pod. More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/init-containers/">https://kubernetes.io/docs/concepts/workloads/pods/init-containers/</a>
Using initContainers for any use case other then secret fetching is entirely outside the scope
of what the maintainers will support and by doing so, you accept that this behaviour may break
at any time without notice.</p>
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
<p>Priority class assigned to the Pods</p>
</td>
</tr>
<tr>
<td>
<code>additionalPeers</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>AdditionalPeers allows injecting a set of additional Alertmanagers to peer with to form a highly available cluster.</p>
</td>
</tr>
<tr>
<td>
<code>clusterAdvertiseAddress</code><br/>
<em>
string
</em>
</td>
<td>
<p>ClusterAdvertiseAddress is the explicit address to advertise in cluster.
Needs to be provided for non RFC1918 <a href="public">1</a> addresses.
[1] RFC1918: <a href="https://tools.ietf.org/html/rfc1918">https://tools.ietf.org/html/rfc1918</a></p>
</td>
</tr>
<tr>
<td>
<code>clusterGossipInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.GoDuration">
GoDuration
</a>
</em>
</td>
<td>
<p>Interval between gossip attempts.</p>
</td>
</tr>
<tr>
<td>
<code>clusterPushpullInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.GoDuration">
GoDuration
</a>
</em>
</td>
<td>
<p>Interval between pushpull attempts.</p>
</td>
</tr>
<tr>
<td>
<code>clusterPeerTimeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.GoDuration">
GoDuration
</a>
</em>
</td>
<td>
<p>Timeout for cluster peering.</p>
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
This defaults to web</p>
</td>
</tr>
<tr>
<td>
<code>forceEnableClusterMode</code><br/>
<em>
bool
</em>
</td>
<td>
<p>ForceEnableClusterMode ensures Alertmanager does not deactivate the cluster mode when running with a single replica.
Use case is e.g. spanning an Alertmanager cluster across Kubernetes clusters with a single replica in each.</p>
</td>
</tr>
<tr>
<td>
<code>alertmanagerConfigSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>AlertmanagerConfigs to be selected for to merge and configure Alertmanager with.</p>
</td>
</tr>
<tr>
<td>
<code>alertmanagerConfigNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to be selected for AlertmanagerConfig discovery. If nil, only
check own namespace.</p>
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
<p>Minimum number of seconds for which a newly created pod should be ready
without any of its container crashing for it to be considered available.
Defaults to 0 (pod will be considered available as soon as it is ready)
This is an alpha field and requires enabling StatefulSetMinReadySeconds feature gate.</p>
</td>
</tr>
<tr>
<td>
<code>hostAliases</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.HostAlias">
[]HostAlias
</a>
</em>
</td>
<td>
<p>Pods&rsquo; hostAliases configuration</p>
</td>
</tr>
<tr>
<td>
<code>web</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertmanagerWebSpec">
AlertmanagerWebSpec
</a>
</em>
</td>
<td>
<p>Defines the web command line flags when starting Alertmanager.</p>
</td>
</tr>
<tr>
<td>
<code>alertmanagerConfiguration</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertmanagerConfiguration">
AlertmanagerConfiguration
</a>
</em>
</td>
<td>
<p>EXPERIMENTAL: alertmanagerConfiguration specifies the configuration of Alertmanager.
If defined, it takes precedence over the <code>configSecret</code> field.
This field may change in future releases.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AlertmanagerStatus">AlertmanagerStatus
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.Alertmanager">Alertmanager</a>)
</p>
<div>
<p>AlertmanagerStatus is the most recent observed status of the Alertmanager cluster. Read-only. Not
included when requesting from the apiserver, only from the Prometheus
Operator API itself. More info:
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
<code>paused</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Represents whether any actions on the underlying managed objects are
being performed. Only delete actions will be performed.</p>
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
<p>Total number of non-terminated pods targeted by this Alertmanager
cluster (their labels match the selector).</p>
</td>
</tr>
<tr>
<td>
<code>updatedReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Total number of non-terminated pods targeted by this Alertmanager
cluster that have the desired version spec.</p>
</td>
</tr>
<tr>
<td>
<code>availableReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Total number of available pods (ready for at least minReadySeconds)
targeted by this Alertmanager cluster.</p>
</td>
</tr>
<tr>
<td>
<code>unavailableReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Total number of unavailable pods targeted by this Alertmanager cluster.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AlertmanagerWebSpec">AlertmanagerWebSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerSpec">AlertmanagerSpec</a>)
</p>
<div>
<p>AlertmanagerWebSpec defines the web command line flags when starting Alertmanager.</p>
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
<code>tlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.WebTLSConfig">
WebTLSConfig
</a>
</em>
</td>
<td>
<p>Defines the TLS parameters for HTTPS.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.WebHTTPConfig">
WebHTTPConfig
</a>
</em>
</td>
<td>
<p>Defines HTTP parameters for web server.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ArbitraryFSAccessThroughSMsConfig">ArbitraryFSAccessThroughSMsConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>)
</p>
<div>
<p>ArbitraryFSAccessThroughSMsConfig enables users to configure, whether
a service monitor selected by the Prometheus instance is allowed to use
arbitrary files on the file system of the Prometheus container. This is the case
when e.g. a service monitor specifies a BearerTokenFile in an endpoint. A
malicious user could create a service monitor selecting arbitrary secret files
in the Prometheus container. Those secrets would then be sent with a scrape
request by Prometheus to a malicious target. Denying the above would prevent the
attack, users can instead use the BearerTokenSecret field.</p>
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
<code>deny</code><br/>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.Argument">Argument
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.ThanosSpec">ThanosSpec</a>)
</p>
<div>
<p>Argument as part of the AdditionalArgs list.</p>
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
<p>Name of the argument, e.g. &ldquo;scrape.discovery-reload-interval&rdquo;.</p>
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
<p>Argument value, e.g. 30s. Can be empty for name-only arguments (e.g. &ndash;storage.tsdb.no-lockfile)</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AttachMetadata">AttachMetadata
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PodMonitorSpec">PodMonitorSpec</a>)
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
<p>When set to true, Prometheus must have permissions to get Nodes.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.Authorization">Authorization
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.APIServerConfig">APIServerConfig</a>, <a href="#monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>)
</p>
<div>
<p>Authorization contains optional <code>Authorization</code> header configuration.
This section is only understood by versions of Prometheus &gt;= 2.26.0.</p>
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
<p>Set the authentication type. Defaults to Bearer, Basic will cause an
error</p>
</td>
</tr>
<tr>
<td>
<code>credentials</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>The secret&rsquo;s key that contains the credentials of the request</p>
</td>
</tr>
<tr>
<td>
<code>credentialsFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>File to read a secret from, mutually exclusive with Credentials (from SafeAuthorization)</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AuthorizationValidationError">AuthorizationValidationError
</h3>
<div>
<p>AuthorizationValidationError is returned by Authorization.Validate()
on semantically invalid configurations.</p>
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
<code>err</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.BasicAuth">BasicAuth
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.APIServerConfig">APIServerConfig</a>, <a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1beta1.HTTPConfig">HTTPConfig</a>)
</p>
<div>
<p>BasicAuth allow an endpoint to authenticate over basic authentication
More info: <a href="https://prometheus.io/docs/operating/configuration/#endpoints">https://prometheus.io/docs/operating/configuration/#endpoints</a></p>
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
<code>username</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>The secret in the service monitor namespace that contains the username
for authentication.</p>
</td>
</tr>
<tr>
<td>
<code>password</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>The secret in the service monitor namespace that contains the password
for authentication.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ByteSize">ByteSize
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>)
</p>
<div>
<p>ByteSize is a valid memory size type based on powers-of-2, so 1KB is 1024B.
Supported units: B, KB, KiB, MB, MiB, GB, GiB, TB, TiB, PB, PiB, EB, EiB Ex: <code>512MB</code>.</p>
</div>
<h3 id="monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>)
</p>
<div>
<p>CommonPrometheusFields are the options available to both the Prometheus server and agent.</p>
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
<code>podMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
EmbeddedObjectMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata configures Labels and Annotations which are propagated to the prometheus pods.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitorSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ServiceMonitors to be selected for target discovery. <em>Deprecated:</em> if
neither this nor podMonitorSelector are specified, configuration is
unmanaged.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitorNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespace&rsquo;s labels to match for ServiceMonitor discovery. If nil, only
check own namespace.</p>
</td>
</tr>
<tr>
<td>
<code>podMonitorSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> PodMonitors to be selected for target discovery.
<em>Deprecated:</em> if neither this nor serviceMonitorSelector are specified,
configuration is unmanaged.</p>
</td>
</tr>
<tr>
<td>
<code>podMonitorNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespace&rsquo;s labels to match for PodMonitor discovery. If nil, only
check own namespace.</p>
</td>
</tr>
<tr>
<td>
<code>probeSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> Probes to be selected for target discovery.</p>
</td>
</tr>
<tr>
<td>
<code>probeNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> Namespaces to be selected for Probe discovery. If nil, only check own namespace.</p>
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
<p>Version of Prometheus to be deployed.</p>
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
<p>Image if specified has precedence over baseImage, tag and sha
combinations. Specifying the version is still necessary to ensure the
Prometheus Operator knows what version of Prometheus is being
configured.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>An optional list of references to secrets in the same namespace
to use for pulling prometheus and alertmanager images from registries
see <a href="http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod">http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod</a></p>
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
<p>Number of replicas of each shard to deploy for a Prometheus deployment.
Number of replicas multiplied by shards is the total number of Pods
created.</p>
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
<p>EXPERIMENTAL: Number of shards to distribute targets onto. Number of
replicas multiplied by shards is the total number of Pods created. Note
that scaling down shards will not reshard data onto remaining instances,
it must be manually moved. Increasing shards will not reshard data
either but it will continue to be available from the same instances. To
query globally use Thanos sidecar and Thanos querier or remote write
data to a central location. Sharding is done on the content of the
<code>__address__</code> target meta-label.</p>
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
<p>Name of Prometheus external label used to denote replica name.
Defaults to the value of <code>prometheus_replica</code>. External label will
<em>not</em> be added when value is set to empty string (<code>&quot;&quot;</code>).</p>
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
<p>Name of Prometheus external label used to denote Prometheus instance
name. Defaults to the value of <code>prometheus</code>. External label will
<em>not</em> be added when value is set to empty string (<code>&quot;&quot;</code>).</p>
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
<p>Log level for Prometheus to be configured with.</p>
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
<p>Log format for Prometheus to be configured with.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Interval between consecutive scrapes. Default: <code>30s</code></p>
</td>
</tr>
<tr>
<td>
<code>scrapeTimeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Number of seconds to wait for target to respond before erroring.</p>
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
external systems (federation, remote storage, Alertmanager).</p>
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
<p>Enable Prometheus to be used as a receiver for the Prometheus remote write protocol. Defaults to the value of <code>false</code>.
WARNING: This is not considered an efficient way of ingesting samples.
Use it with caution for specific low-volume use cases.
It is not suitable for replacing the ingestion via scraping and turning
Prometheus into a push-based metrics collection system.
For more information see <a href="https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver">https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver</a>
Only valid in Prometheus versions 2.33.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>enableFeatures</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Enable access to Prometheus disabled features. By default, no features are enabled.
Enabling disabled features is entirely outside the scope of what the maintainers will
support and by doing so, you accept that this behaviour may break at any
time without notice.
For more information see <a href="https://prometheus.io/docs/prometheus/latest/disabled_features/">https://prometheus.io/docs/prometheus/latest/disabled_features/</a></p>
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
<p>The external URL the Prometheus instances will be available under. This is
necessary to generate correct URLs. This is necessary if Prometheus is not
served from root of a DNS name.</p>
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
<p>The route prefix Prometheus registers HTTP handlers for. This is useful,
if using ExternalURL and a proxy is rewriting HTTP routes of a request,
and the actual ExternalURL is still true, but the server serves requests
under a different route prefix. For example for use with <code>kubectl proxy</code>.</p>
</td>
</tr>
<tr>
<td>
<code>storage</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.StorageSpec">
StorageSpec
</a>
</em>
</td>
<td>
<p>Storage spec to specify how storage shall be used.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<p>Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will
be appended to other volumes that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<p>VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
VolumeMounts specified will be appended to other VolumeMounts in the prometheus container,
that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>web</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusWebSpec">
PrometheusWebSpec
</a>
</em>
</td>
<td>
<p>Defines the web command line flags when starting Prometheus.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>Define resources requests and limits for single Pods.</p>
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
<p>Define which Nodes the Pods are scheduled on.</p>
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
<code>secrets</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Secrets is a list of Secrets in the same namespace as the Prometheus
object, which shall be mounted into the Prometheus Pods.
The Secrets are mounted into /etc/prometheus/secrets/<secret-name>.</p>
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
The ConfigMaps are mounted into /etc/prometheus/configmaps/<configmap-name>.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s scheduling constraints.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s topology spread constraints.</p>
</td>
</tr>
<tr>
<td>
<code>remoteWrite</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RemoteWriteSpec">
[]RemoteWriteSpec
</a>
</em>
</td>
<td>
<p>remoteWrite is the list of remote write configurations.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<p>SecurityContext holds pod-level security attributes and common container settings.
This defaults to the default PodSecurityContext.</p>
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
<p>ListenLocal makes the Prometheus server listen on loopback, so that it
does not bind against the Pod IP.</p>
</td>
</tr>
<tr>
<td>
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>Containers allows injecting additional containers or modifying operator
generated containers. This can be used to allow adding an authentication
proxy to a Prometheus pod or to change the behavior of an operator
generated container. Containers described here modify an operator
generated container if they share the same name and modifications are
done via a strategic merge patch. The current container names are:
<code>prometheus</code>, <code>config-reloader</code>, and <code>thanos-sidecar</code>. Overriding
containers is entirely outside the scope of what the maintainers will
support and by doing so, you accept that this behaviour may break at any
time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>initContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
fetch secrets for injection into the Prometheus configuration from external sources. Any errors
during the execution of an initContainer will lead to a restart of the Pod. More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/init-containers/">https://kubernetes.io/docs/concepts/workloads/pods/init-containers/</a>
InitContainers described here modify an operator
generated init containers if they share the same name and modifications are
done via a strategic merge patch. The current init container name is:
<code>init-config-reloader</code>. Overriding init containers is entirely outside the
scope of what the maintainers will support and by doing so, you accept that
this behaviour may break at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>additionalScrapeConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
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
<a href="#monitoring.coreos.com/v1.APIServerConfig">
APIServerConfig
</a>
</em>
</td>
<td>
<p>APIServerConfig allows specifying a host and auth methods to access apiserver.
If left empty, Prometheus is assumed to run inside of the cluster
and will discover API servers automatically and use the pod&rsquo;s CA certificate
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
<p>Priority class assigned to the Pods</p>
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
This defaults to web</p>
</td>
</tr>
<tr>
<td>
<code>arbitraryFSAccessThroughSMs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ArbitraryFSAccessThroughSMsConfig">
ArbitraryFSAccessThroughSMsConfig
</a>
</em>
</td>
<td>
<p>ArbitraryFSAccessThroughSMs configures whether configuration
based on a service monitor can access arbitrary files on the file system
of the Prometheus container e.g. bearer token files.</p>
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
<p>When true, Prometheus resolves label conflicts by renaming the labels in
the scraped data to &ldquo;exported_<label value>&rdquo; for all targets created
from service and pod monitors.
Otherwise the HonorLabels field of the service or pod monitor applies.</p>
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
<p>IgnoreNamespaceSelectors if set to true will ignore NamespaceSelector
settings from all PodMonitor, ServiceMonitor and Probe objects. They will
only discover endpoints within the namespace of the PodMonitor,
ServiceMonitor and Probe objects.
Defaults to false.</p>
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
<p>EnforcedNamespaceLabel If set, a label will be added to</p>
<ol>
<li>all user-metrics (created by <code>ServiceMonitor</code>, <code>PodMonitor</code> and <code>Probe</code> objects) and</li>
<li>in all <code>PrometheusRule</code> objects (except the ones excluded in <code>prometheusRulesExcludedFromEnforce</code>) to
<ul>
<li>alerting &amp; recording rules and</li>
<li>the metrics used in their expressions (<code>expr</code>).</li>
</ul></li>
</ol>
<p>Label name is this field&rsquo;s value.
Label value is the namespace of the created object (mentioned above).</p>
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
<p>EnforcedSampleLimit defines global limit on number of scraped samples
that will be accepted. This overrides any SampleLimit set per
ServiceMonitor or/and PodMonitor. It is meant to be used by admins to
enforce the SampleLimit to keep overall number of samples/series under
the desired limit.
Note that if SampleLimit is lower that value will be taken instead.</p>
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
<p>EnforcedTargetLimit defines a global limit on the number of scraped
targets.  This overrides any TargetLimit set per ServiceMonitor or/and
PodMonitor.  It is meant to be used by admins to enforce the TargetLimit
to keep the overall number of targets under the desired limit.
Note that if TargetLimit is lower, that value will be taken instead,
except if either value is zero, in which case the non-zero value will be
used.  If both values are zero, no limit is enforced.</p>
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
<p>Per-scrape limit on number of labels that will be accepted for a sample. If
more than this number of labels are present post metric-relabeling, the
entire scrape will be treated as failed. 0 means no limit.
Only valid in Prometheus versions 2.27.0 and newer.</p>
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
<p>Per-scrape limit on length of labels name that will be accepted for a sample.
If a label name is longer than this number post metric-relabeling, the entire
scrape will be treated as failed. 0 means no limit.
Only valid in Prometheus versions 2.27.0 and newer.</p>
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
<p>Per-scrape limit on length of labels value that will be accepted for a sample.
If a label value is longer than this number post metric-relabeling, the
entire scrape will be treated as failed. 0 means no limit.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedBodySizeLimit</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
</a>
</em>
</td>
<td>
<p>EnforcedBodySizeLimit defines the maximum size of uncompressed response body
that will be accepted by Prometheus. Targets responding with a body larger than this many bytes
will cause the scrape to fail. Example: 100MB.
If defined, the limit will apply to all service/pod monitors and probes.
This is an experimental feature, this behaviour could
change or be removed in the future.
Only valid in Prometheus versions 2.28.0 and newer.</p>
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
<p>Minimum number of seconds for which a newly created pod should be ready
without any of its container crashing for it to be considered available.
Defaults to 0 (pod will be considered available as soon as it is ready)
This is an alpha field and requires enabling StatefulSetMinReadySeconds feature gate.</p>
</td>
</tr>
<tr>
<td>
<code>hostAliases</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.HostAlias">
[]HostAlias
</a>
</em>
</td>
<td>
<p>Pods&rsquo; hostAliases configuration</p>
</td>
</tr>
<tr>
<td>
<code>additionalArgs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Argument">
[]Argument
</a>
</em>
</td>
<td>
<p>AdditionalArgs allows setting additional arguments for the Prometheus container.
It is intended for e.g. activating hidden flags which are not supported by
the dedicated configuration options yet. The arguments are passed as-is to the
Prometheus container which may cause issues if they are invalid or not supporeted
by the given Prometheus version.
In case of an argument conflict (e.g. an argument which is already set by the
operator itself) or when providing an invalid argument the reconciliation will
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
<p>Enable compression of the write-ahead log using Snappy. This flag is
only available in versions of Prometheus &gt;= 2.11.0.</p>
</td>
</tr>
<tr>
<td>
<code>excludedFromEnforcement</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ObjectReference">
[]ObjectReference
</a>
</em>
</td>
<td>
<p>List of references to PodMonitor, ServiceMonitor, Probe and PrometheusRule objects
to be excluded from enforcing a namespace label of origin.
Applies only if enforcedNamespaceLabel set to true.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.Duration">Duration
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>, <a href="#monitoring.coreos.com/v1.AlertmanagerGlobalConfig">AlertmanagerGlobalConfig</a>, <a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.MetadataConfig">MetadataConfig</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>, <a href="#monitoring.coreos.com/v1.QuerySpec">QuerySpec</a>, <a href="#monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosSpec">ThanosSpec</a>)
</p>
<div>
<p>Duration is a valid time duration that can be parsed by Prometheus model.ParseDuration() function.
Supported units: y, w, d, h, m, s, ms
Examples: <code>30s</code>, <code>1m</code>, <code>1h20m15s</code>, <code>15d</code></p>
</div>
<h3 id="monitoring.coreos.com/v1.EmbeddedObjectMetadata">EmbeddedObjectMetadata
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerSpec">AlertmanagerSpec</a>, <a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.EmbeddedPersistentVolumeClaim">EmbeddedPersistentVolumeClaim</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>)
</p>
<div>
<p>EmbeddedObjectMetadata contains a subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta
Only fields which are relevant to embedded resources are included.</p>
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
<em>(Optional)</em>
<p>Name must be unique within a namespace. Is required when creating resources, although
some resources may allow a client to request the generation of an appropriate name
automatically. Name is primarily intended for creation idempotence and configuration
definition.
Cannot be updated.
More info: <a href="http://kubernetes.io/docs/user-guide/identifiers#names">http://kubernetes.io/docs/user-guide/identifiers#names</a></p>
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
<p>Map of string keys and values that can be used to organize and categorize
(scope and select) objects. May match selectors of replication controllers
and services.
More info: <a href="http://kubernetes.io/docs/user-guide/labels">http://kubernetes.io/docs/user-guide/labels</a></p>
</td>
</tr>
<tr>
<td>
<code>annotations</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Annotations is an unstructured key value map stored with a resource that may be
set by external tools to store and retrieve arbitrary metadata. They are not
queryable and should be preserved when modifying objects.
More info: <a href="http://kubernetes.io/docs/user-guide/annotations">http://kubernetes.io/docs/user-guide/annotations</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.EmbeddedPersistentVolumeClaim">EmbeddedPersistentVolumeClaim
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.StorageSpec">StorageSpec</a>)
</p>
<div>
<p>EmbeddedPersistentVolumeClaim is an embedded version of k8s.io/api/core/v1.PersistentVolumeClaim.
It contains TypeMeta and a reduced ObjectMeta.</p>
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
<code>metadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
EmbeddedObjectMetadata
</a>
</em>
</td>
<td>
<p>EmbeddedMetadata contains metadata relevant to an EmbeddedResource.</p>
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#persistentvolumeclaimspec-v1-core">
Kubernetes core/v1.PersistentVolumeClaimSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Spec defines the desired characteristics of a volume requested by a pod author.
More info: <a href="https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims">https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims</a></p>
<br/>
<br/>
<table>
<tr>
<td>
<code>accessModes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#persistentvolumeaccessmode-v1-core">
[]Kubernetes core/v1.PersistentVolumeAccessMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>accessModes contains the desired access modes the volume should have.
More info: <a href="https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1">https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1</a></p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>selector is a label query over volumes to consider for binding.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>resources represents the minimum resources the volume should have.
If RecoverVolumeExpansionFailure feature is enabled users are allowed to specify resource requirements
that are lower than previous value but must still be higher than capacity recorded in the
status field of the claim.
More info: <a href="https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources">https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources</a></p>
</td>
</tr>
<tr>
<td>
<code>volumeName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>volumeName is the binding reference to the PersistentVolume backing this claim.</p>
</td>
</tr>
<tr>
<td>
<code>storageClassName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>storageClassName is the name of the StorageClass required by the claim.
More info: <a href="https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1">https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1</a></p>
</td>
</tr>
<tr>
<td>
<code>volumeMode</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#persistentvolumemode-v1-core">
Kubernetes core/v1.PersistentVolumeMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>volumeMode defines what type of volume is required by the claim.
Value of Filesystem is implied when not included in claim spec.</p>
</td>
</tr>
<tr>
<td>
<code>dataSource</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#typedlocalobjectreference-v1-core">
Kubernetes core/v1.TypedLocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>dataSource field can be used to specify either:
* An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot)
* An existing PVC (PersistentVolumeClaim)
If the provisioner or an external controller can support the specified data source,
it will create a new volume based on the contents of the specified data source.
If the AnyVolumeDataSource feature gate is enabled, this field will always have
the same contents as the DataSourceRef field.</p>
</td>
</tr>
<tr>
<td>
<code>dataSourceRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#typedlocalobjectreference-v1-core">
Kubernetes core/v1.TypedLocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>dataSourceRef specifies the object from which to populate the volume with data, if a non-empty
volume is desired. This may be any local object from a non-empty API group (non
core object) or a PersistentVolumeClaim object.
When this field is specified, volume binding will only succeed if the type of
the specified object matches some installed volume populator or dynamic
provisioner.
This field will replace the functionality of the DataSource field and as such
if both fields are non-empty, they must have the same value. For backwards
compatibility, both fields (DataSource and DataSourceRef) will be set to the same
value automatically if one of them is empty and the other is non-empty.
There are two important differences between DataSource and DataSourceRef:
* While DataSource only allows two specific types of objects, DataSourceRef
allows any non-core object, as well as PersistentVolumeClaim objects.
* While DataSource ignores disallowed values (dropping them), DataSourceRef
preserves all values, and generates an error if a disallowed value is
specified.
(Beta) Using this field requires the AnyVolumeDataSource feature gate to be enabled.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#persistentvolumeclaimstatus-v1-core">
Kubernetes core/v1.PersistentVolumeClaimStatus
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Status represents the current information/status of a persistent volume claim.
Read-only.
More info: <a href="https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims">https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.Endpoint">Endpoint
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ServiceMonitorSpec">ServiceMonitorSpec</a>)
</p>
<div>
<p>Endpoint defines a scrapeable endpoint serving Prometheus metrics.</p>
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
<code>port</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the service port this endpoint refers to. Mutually exclusive with targetPort.</p>
</td>
</tr>
<tr>
<td>
<code>targetPort</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/util/intstr#IntOrString">
k8s.io/apimachinery/pkg/util/intstr.IntOrString
</a>
</em>
</td>
<td>
<p>Name or number of the target port of the Pod behind the Service, the port must be specified with container port property. Mutually exclusive with port.</p>
</td>
</tr>
<tr>
<td>
<code>path</code><br/>
<em>
string
</em>
</td>
<td>
<p>HTTP path to scrape for metrics.
If empty, Prometheus uses the default value (e.g. <code>/metrics</code>).</p>
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
<p>HTTP scheme to use for scraping.</p>
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
<p>Optional HTTP URL parameters</p>
</td>
</tr>
<tr>
<td>
<code>interval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Interval at which metrics should be scraped
If not specified Prometheus&rsquo; global scrape interval is used.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeTimeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Timeout after which the scrape is ended
If not specified, the Prometheus global scrape timeout is used unless it is less than <code>Interval</code> in which the latter is used.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TLSConfig">
TLSConfig
</a>
</em>
</td>
<td>
<p>TLS configuration to use when scraping the endpoint</p>
</td>
</tr>
<tr>
<td>
<code>bearerTokenFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>File to read bearer token for scraping targets.</p>
</td>
</tr>
<tr>
<td>
<code>bearerTokenSecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Secret to mount to read bearer token for scraping targets. The secret
needs to be in the same namespace as the service monitor and accessible by
the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
</a>
</em>
</td>
<td>
<p>Authorization section for this endpoint</p>
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
<p>HonorLabels chooses the metric&rsquo;s labels on collisions with target labels.</p>
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
<p>HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
</a>
</em>
</td>
<td>
<p>BasicAuth allow an endpoint to authenticate over basic authentication
More info: <a href="https://prometheus.io/docs/operating/configuration/#endpoints">https://prometheus.io/docs/operating/configuration/#endpoints</a></p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
</a>
</em>
</td>
<td>
<p>OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>metricRelabelings</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RelabelConfig">
[]RelabelConfig
</a>
</em>
</td>
<td>
<p>MetricRelabelConfigs to apply to samples before ingestion.</p>
</td>
</tr>
<tr>
<td>
<code>relabelings</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RelabelConfig">
[]RelabelConfig
</a>
</em>
</td>
<td>
<p>RelabelConfigs to apply to samples before scraping.
Prometheus Operator automatically adds relabelings for a few standard Kubernetes fields.
The original scrape job&rsquo;s name is available via the <code>__tmp_prometheus_job_name</code> label.
More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config</a></p>
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
<p>ProxyURL eg <a href="http://proxyserver:2195">http://proxyserver:2195</a> Directs scrapes to proxy through this endpoint.</p>
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
<p>FollowRedirects configures whether scrape requests follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHttp2</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.Exemplars">Exemplars
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>)
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
<code>maxSize</code><br/>
<em>
int64
</em>
</td>
<td>
<p>Maximum number of exemplars stored in memory for all series.
If not set, Prometheus uses its default value.
A value of zero or less than zero disables the storage.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.GoDuration">GoDuration
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerSpec">AlertmanagerSpec</a>)
</p>
<div>
<p>GoDuration is a valid time duration that can be parsed by Go&rsquo;s time.ParseDuration() function.
Supported units: h, m, s, ms
Examples: <code>45ms</code>, <code>30s</code>, <code>1m</code>, <code>1h20m15s</code></p>
</div>
<h3 id="monitoring.coreos.com/v1.HTTPConfig">HTTPConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerGlobalConfig">AlertmanagerGlobalConfig</a>)
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
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
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
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
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
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The secret&rsquo;s key that contains the bearer token to be used by the client
for authentication.
The secret needs to be in the same namespace as the Alertmanager
object and accessible by the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
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
<h3 id="monitoring.coreos.com/v1.HostAlias">HostAlias
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerSpec">AlertmanagerSpec</a>, <a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>)
</p>
<div>
<p>HostAlias holds the mapping between IP and hostnames that will be injected as an entry in the
pod&rsquo;s hosts file.</p>
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
<code>ip</code><br/>
<em>
string
</em>
</td>
<td>
<p>IP address of the host file entry.</p>
</td>
</tr>
<tr>
<td>
<code>hostnames</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Hostnames for the above IP address.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.LabelName">LabelName
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.RelabelConfig">RelabelConfig</a>)
</p>
<div>
<p>LabelName is a valid Prometheus label name which may only contain ASCII letters, numbers, as well as underscores.</p>
</div>
<h3 id="monitoring.coreos.com/v1.MetadataConfig">MetadataConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>)
</p>
<div>
<p>MetadataConfig configures the sending of series metadata to the remote storage.</p>
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
<code>send</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Whether metric metadata is sent to the remote storage or not.</p>
</td>
</tr>
<tr>
<td>
<code>sendInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>How frequently metric metadata is sent to the remote storage.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.NamespaceSelector">NamespaceSelector
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PodMonitorSpec">PodMonitorSpec</a>, <a href="#monitoring.coreos.com/v1.ProbeTargetIngress">ProbeTargetIngress</a>, <a href="#monitoring.coreos.com/v1.ServiceMonitorSpec">ServiceMonitorSpec</a>)
</p>
<div>
<p>NamespaceSelector is a selector for selecting either all namespaces or a
list of namespaces.
If <code>any</code> is true, it takes precedence over <code>matchNames</code>.
If <code>matchNames</code> is empty and <code>any</code> is false, it means that the objects are
selected from the current namespace.</p>
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
<code>any</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Boolean describing whether all namespaces are selected in contrast to a
list restricting them.</p>
</td>
</tr>
<tr>
<td>
<code>matchNames</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>List of namespace names to select from.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.OAuth2">OAuth2
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1beta1.HTTPConfig">HTTPConfig</a>)
</p>
<div>
<p>OAuth2 allows an endpoint to authenticate with OAuth2.
More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#oauth2">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#oauth2</a></p>
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
<code>clientId</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SecretOrConfigMap">
SecretOrConfigMap
</a>
</em>
</td>
<td>
<p>The secret or configmap containing the OAuth2 client id</p>
</td>
</tr>
<tr>
<td>
<code>clientSecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>The secret containing the OAuth2 client secret</p>
</td>
</tr>
<tr>
<td>
<code>tokenUrl</code><br/>
<em>
string
</em>
</td>
<td>
<p>The URL to fetch the token from</p>
</td>
</tr>
<tr>
<td>
<code>scopes</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>OAuth2 scopes used for the token request</p>
</td>
</tr>
<tr>
<td>
<code>endpointParams</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>Parameters to append to the token URL</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.OAuth2ValidationError">OAuth2ValidationError
</h3>
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
<code>err</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ObjectReference">ObjectReference
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>)
</p>
<div>
<p>ObjectReference references a PodMonitor, ServiceMonitor, Probe or PrometheusRule object.</p>
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
<code>group</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Group of the referent. When not specified, it defaults to <code>monitoring.coreos.com</code></p>
</td>
</tr>
<tr>
<td>
<code>resource</code><br/>
<em>
string
</em>
</td>
<td>
<p>Resource of the referent.</p>
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
<p>Namespace of the referent.
More info: <a href="https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/">https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/</a></p>
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
<p>Name of the referent. When not set, all resources are matched.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PodMonitorSpec">PodMonitorSpec</a>)
</p>
<div>
<p>PodMetricsEndpoint defines a scrapeable endpoint of a Kubernetes Pod serving Prometheus metrics.</p>
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
<code>port</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the pod port this endpoint refers to. Mutually exclusive with targetPort.</p>
</td>
</tr>
<tr>
<td>
<code>targetPort</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/util/intstr#IntOrString">
k8s.io/apimachinery/pkg/util/intstr.IntOrString
</a>
</em>
</td>
<td>
<p>Deprecated: Use &lsquo;port&rsquo; instead.</p>
</td>
</tr>
<tr>
<td>
<code>path</code><br/>
<em>
string
</em>
</td>
<td>
<p>HTTP path to scrape for metrics.
If empty, Prometheus uses the default value (e.g. <code>/metrics</code>).</p>
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
<p>HTTP scheme to use for scraping.</p>
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
<p>Optional HTTP URL parameters</p>
</td>
</tr>
<tr>
<td>
<code>interval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Interval at which metrics should be scraped
If not specified Prometheus&rsquo; global scrape interval is used.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeTimeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Timeout after which the scrape is ended
If not specified, the Prometheus global scrape interval is used.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PodMetricsEndpointTLSConfig">
PodMetricsEndpointTLSConfig
</a>
</em>
</td>
<td>
<p>TLS configuration to use when scraping the endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>bearerTokenSecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Secret to mount to read bearer token for scraping targets. The secret
needs to be in the same namespace as the pod monitor and accessible by
the Prometheus Operator.</p>
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
<p>HonorLabels chooses the metric&rsquo;s labels on collisions with target labels.</p>
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
<p>HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
</a>
</em>
</td>
<td>
<p>BasicAuth allow an endpoint to authenticate over basic authentication.
More info: <a href="https://prometheus.io/docs/operating/configuration/#endpoint">https://prometheus.io/docs/operating/configuration/#endpoint</a></p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
</a>
</em>
</td>
<td>
<p>OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
</a>
</em>
</td>
<td>
<p>Authorization section for this endpoint</p>
</td>
</tr>
<tr>
<td>
<code>metricRelabelings</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RelabelConfig">
[]RelabelConfig
</a>
</em>
</td>
<td>
<p>MetricRelabelConfigs to apply to samples before ingestion.</p>
</td>
</tr>
<tr>
<td>
<code>relabelings</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RelabelConfig">
[]RelabelConfig
</a>
</em>
</td>
<td>
<p>RelabelConfigs to apply to samples before scraping.
Prometheus Operator automatically adds relabelings for a few standard Kubernetes fields.
The original scrape job&rsquo;s name is available via the <code>__tmp_prometheus_job_name</code> label.
More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config</a></p>
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
<p>ProxyURL eg <a href="http://proxyserver:2195">http://proxyserver:2195</a> Directs scrapes to proxy through this endpoint.</p>
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
<p>FollowRedirects configures whether scrape requests follow HTTP 3xx redirects.</p>
</td>
</tr>
<tr>
<td>
<code>enableHttp2</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Whether to enable HTTP2.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.PodMetricsEndpointTLSConfig">PodMetricsEndpointTLSConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>)
</p>
<div>
<p>PodMetricsEndpointTLSConfig specifies TLS configuration parameters.</p>
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
<code>ca</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SecretOrConfigMap">
SecretOrConfigMap
</a>
</em>
</td>
<td>
<p>Struct containing the CA cert to use for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>cert</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SecretOrConfigMap">
SecretOrConfigMap
</a>
</em>
</td>
<td>
<p>Struct containing the client cert file for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>keySecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Secret containing the client key file for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>serverName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Used to verify the hostname for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>insecureSkipVerify</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Disable target certificate validation.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.PodMonitorSpec">PodMonitorSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PodMonitor">PodMonitor</a>)
</p>
<div>
<p>PodMonitorSpec contains specification parameters for a PodMonitor.</p>
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
<code>jobLabel</code><br/>
<em>
string
</em>
</td>
<td>
<p>The label to use to retrieve the job name from.</p>
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
<p>PodTargetLabels transfers labels on the Kubernetes Pod onto the target.</p>
</td>
</tr>
<tr>
<td>
<code>podMetricsEndpoints</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">
[]PodMetricsEndpoint
</a>
</em>
</td>
<td>
<p>A list of endpoints allowed as part of this PodMonitor.</p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Selector to select Pod objects.</p>
</td>
</tr>
<tr>
<td>
<code>namespaceSelector</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.NamespaceSelector">
NamespaceSelector
</a>
</em>
</td>
<td>
<p>Selector to select which namespaces the Endpoints objects are discovered from.</p>
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
<p>Per-scrape limit on length of labels value that will be accepted for a sample.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>attachMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AttachMetadata">
AttachMetadata
</a>
</em>
</td>
<td>
<p>Attaches node metadata to discovered targets. Only valid for role: pod.
Only valid in Prometheus versions 2.35.0 and newer.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ProbeSpec">ProbeSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.Probe">Probe</a>)
</p>
<div>
<p>ProbeSpec contains specification parameters for a Probe.</p>
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
<p>The job name assigned to scraped metrics by default.</p>
</td>
</tr>
<tr>
<td>
<code>prober</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ProberSpec">
ProberSpec
</a>
</em>
</td>
<td>
<p>Specification for the prober to use for probing targets.
The prober.URL parameter is required. Targets cannot be probed if left empty.</p>
</td>
</tr>
<tr>
<td>
<code>module</code><br/>
<em>
string
</em>
</td>
<td>
<p>The module to use for probing specifying how to probe the target.
Example module configuring in the blackbox exporter:
<a href="https://github.com/prometheus/blackbox_exporter/blob/master/example.yml">https://github.com/prometheus/blackbox_exporter/blob/master/example.yml</a></p>
</td>
</tr>
<tr>
<td>
<code>targets</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ProbeTargets">
ProbeTargets
</a>
</em>
</td>
<td>
<p>Targets defines a set of static or dynamically discovered targets to probe.</p>
</td>
</tr>
<tr>
<td>
<code>interval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Interval at which targets are probed using the configured prober.
If not specified Prometheus&rsquo; global scrape interval is used.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeTimeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Timeout for scraping metrics from the Prometheus exporter.
If not specified, the Prometheus global scrape interval is used.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ProbeTLSConfig">
ProbeTLSConfig
</a>
</em>
</td>
<td>
<p>TLS configuration to use when scraping the endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>bearerTokenSecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Secret to mount to read bearer token for scraping targets. The secret
needs to be in the same namespace as the probe and accessible by
the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
</a>
</em>
</td>
<td>
<p>BasicAuth allow an endpoint to authenticate over basic authentication.
More info: <a href="https://prometheus.io/docs/operating/configuration/#endpoint">https://prometheus.io/docs/operating/configuration/#endpoint</a></p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
</a>
</em>
</td>
<td>
<p>OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>metricRelabelings</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RelabelConfig">
[]RelabelConfig
</a>
</em>
</td>
<td>
<p>MetricRelabelConfigs to apply to samples before ingestion.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
</a>
</em>
</td>
<td>
<p>Authorization section for this endpoint</p>
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
<p>Per-scrape limit on length of labels value that will be accepted for a sample.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ProbeTLSConfig">ProbeTLSConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>)
</p>
<div>
<p>ProbeTLSConfig specifies TLS configuration parameters.</p>
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
<code>ca</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SecretOrConfigMap">
SecretOrConfigMap
</a>
</em>
</td>
<td>
<p>Struct containing the CA cert to use for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>cert</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SecretOrConfigMap">
SecretOrConfigMap
</a>
</em>
</td>
<td>
<p>Struct containing the client cert file for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>keySecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Secret containing the client key file for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>serverName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Used to verify the hostname for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>insecureSkipVerify</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Disable target certificate validation.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ProbeTargetIngress">ProbeTargetIngress
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ProbeTargets">ProbeTargets</a>)
</p>
<div>
<p>ProbeTargetIngress defines the set of Ingress objects considered for probing.
The operator configures a target for each host/path combination of each ingress object.</p>
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
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Selector to select the Ingress objects.</p>
</td>
</tr>
<tr>
<td>
<code>namespaceSelector</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.NamespaceSelector">
NamespaceSelector
</a>
</em>
</td>
<td>
<p>From which namespaces to select Ingress objects.</p>
</td>
</tr>
<tr>
<td>
<code>relabelingConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RelabelConfig">
[]RelabelConfig
</a>
</em>
</td>
<td>
<p>RelabelConfigs to apply to the label set of the target before it gets
scraped.
The original ingress address is available via the
<code>__tmp_prometheus_ingress_address</code> label. It can be used to customize the
probed URL.
The original scrape job&rsquo;s name is available via the <code>__tmp_prometheus_job_name</code> label.
More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ProbeTargetStaticConfig">ProbeTargetStaticConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ProbeTargets">ProbeTargets</a>)
</p>
<div>
<p>ProbeTargetStaticConfig defines the set of static targets considered for probing.</p>
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
<code>static</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>The list of hosts to probe.</p>
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
<p>Labels assigned to all metrics scraped from the targets.</p>
</td>
</tr>
<tr>
<td>
<code>relabelingConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RelabelConfig">
[]RelabelConfig
</a>
</em>
</td>
<td>
<p>RelabelConfigs to apply to the label set of the targets before it gets
scraped.
More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ProbeTargets">ProbeTargets
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>)
</p>
<div>
<p>ProbeTargets defines how to discover the probed targets.
One of the <code>staticConfig</code> or <code>ingress</code> must be defined.
If both are defined, <code>staticConfig</code> takes precedence.</p>
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
<code>staticConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ProbeTargetStaticConfig">
ProbeTargetStaticConfig
</a>
</em>
</td>
<td>
<p>staticConfig defines the static list of targets to probe and the
relabeling configuration.
If <code>ingress</code> is also defined, <code>staticConfig</code> takes precedence.
More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#static_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#static_config</a>.</p>
</td>
</tr>
<tr>
<td>
<code>ingress</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ProbeTargetIngress">
ProbeTargetIngress
</a>
</em>
</td>
<td>
<p>ingress defines the Ingress objects to probe and the relabeling
configuration.
If <code>staticConfig</code> is also defined, <code>staticConfig</code> takes precedence.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ProbeTargetsValidationError">ProbeTargetsValidationError
</h3>
<div>
<p>ProbeTargetsValidationError is returned by ProbeTargets.Validate()
on semantically invalid configurations.</p>
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
<code>err</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ProberSpec">ProberSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>)
</p>
<div>
<p>ProberSpec contains specification parameters for the Prober used for probing.</p>
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
<p>Mandatory URL of the prober.</p>
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
<p>HTTP scheme to use for scraping.
Defaults to <code>http</code>.</p>
</td>
</tr>
<tr>
<td>
<code>path</code><br/>
<em>
string
</em>
</td>
<td>
<p>Path to collect metrics from.
Defaults to <code>/probe</code>.</p>
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
<p>Optional ProxyURL.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.PrometheusCondition">PrometheusCondition
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusStatus">PrometheusStatus</a>)
</p>
<div>
<p>PrometheusCondition represents the state of the resources associated with the Prometheus resource.</p>
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
<a href="#monitoring.coreos.com/v1.PrometheusConditionType">
PrometheusConditionType
</a>
</em>
</td>
<td>
<p>Type of the condition being reported.</p>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusConditionStatus">
PrometheusConditionStatus
</a>
</em>
</td>
<td>
<p>status of the condition.</p>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>lastTransitionTime is the time of the last update to the current status property.</p>
</td>
</tr>
<tr>
<td>
<code>reason</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Reason for the condition&rsquo;s last transition.</p>
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
<p>Human-readable message indicating details for the condition&rsquo;s last transition.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.PrometheusConditionStatus">PrometheusConditionStatus
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusCondition">PrometheusCondition</a>)
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
<tbody><tr><td><p>&#34;Degraded&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;False&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;True&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Unknown&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1.PrometheusConditionType">PrometheusConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusCondition">PrometheusCondition</a>)
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
<tbody><tr><td><p>&#34;Available&#34;</p></td>
<td><p>Available indicates whether enough Prometheus pods are ready to provide the service.</p>
</td>
</tr><tr><td><p>&#34;Reconciled&#34;</p></td>
<td><p>Reconciled indicates that the operator has reconciled the state of the underlying resources with the Prometheus object spec.</p>
</td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1.PrometheusRuleExcludeConfig">PrometheusRuleExcludeConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>)
</p>
<div>
<p>PrometheusRuleExcludeConfig enables users to configure excluded PrometheusRule names and their namespaces
to be ignored while enforcing namespace label for alerts and metrics.</p>
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
<code>ruleNamespace</code><br/>
<em>
string
</em>
</td>
<td>
<p>RuleNamespace - namespace of excluded rule</p>
</td>
</tr>
<tr>
<td>
<code>ruleName</code><br/>
<em>
string
</em>
</td>
<td>
<p>RuleNamespace - name of excluded rule</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.PrometheusRuleSpec">PrometheusRuleSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusRule">PrometheusRule</a>)
</p>
<div>
<p>PrometheusRuleSpec contains specification parameters for a Rule.</p>
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
<code>groups</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RuleGroup">
[]RuleGroup
</a>
</em>
</td>
<td>
<p>Content of Prometheus rule file</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.Prometheus">Prometheus</a>)
</p>
<div>
<p>PrometheusSpec is a specification of the desired behavior of the Prometheus cluster. More info:
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
<code>podMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
EmbeddedObjectMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata configures Labels and Annotations which are propagated to the prometheus pods.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitorSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ServiceMonitors to be selected for target discovery. <em>Deprecated:</em> if
neither this nor podMonitorSelector are specified, configuration is
unmanaged.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitorNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespace&rsquo;s labels to match for ServiceMonitor discovery. If nil, only
check own namespace.</p>
</td>
</tr>
<tr>
<td>
<code>podMonitorSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> PodMonitors to be selected for target discovery.
<em>Deprecated:</em> if neither this nor serviceMonitorSelector are specified,
configuration is unmanaged.</p>
</td>
</tr>
<tr>
<td>
<code>podMonitorNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespace&rsquo;s labels to match for PodMonitor discovery. If nil, only
check own namespace.</p>
</td>
</tr>
<tr>
<td>
<code>probeSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> Probes to be selected for target discovery.</p>
</td>
</tr>
<tr>
<td>
<code>probeNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> Namespaces to be selected for Probe discovery. If nil, only check own namespace.</p>
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
<p>Version of Prometheus to be deployed.</p>
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
<p>Image if specified has precedence over baseImage, tag and sha
combinations. Specifying the version is still necessary to ensure the
Prometheus Operator knows what version of Prometheus is being
configured.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>An optional list of references to secrets in the same namespace
to use for pulling prometheus and alertmanager images from registries
see <a href="http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod">http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod</a></p>
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
<p>Number of replicas of each shard to deploy for a Prometheus deployment.
Number of replicas multiplied by shards is the total number of Pods
created.</p>
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
<p>EXPERIMENTAL: Number of shards to distribute targets onto. Number of
replicas multiplied by shards is the total number of Pods created. Note
that scaling down shards will not reshard data onto remaining instances,
it must be manually moved. Increasing shards will not reshard data
either but it will continue to be available from the same instances. To
query globally use Thanos sidecar and Thanos querier or remote write
data to a central location. Sharding is done on the content of the
<code>__address__</code> target meta-label.</p>
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
<p>Name of Prometheus external label used to denote replica name.
Defaults to the value of <code>prometheus_replica</code>. External label will
<em>not</em> be added when value is set to empty string (<code>&quot;&quot;</code>).</p>
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
<p>Name of Prometheus external label used to denote Prometheus instance
name. Defaults to the value of <code>prometheus</code>. External label will
<em>not</em> be added when value is set to empty string (<code>&quot;&quot;</code>).</p>
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
<p>Log level for Prometheus to be configured with.</p>
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
<p>Log format for Prometheus to be configured with.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Interval between consecutive scrapes. Default: <code>30s</code></p>
</td>
</tr>
<tr>
<td>
<code>scrapeTimeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Number of seconds to wait for target to respond before erroring.</p>
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
external systems (federation, remote storage, Alertmanager).</p>
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
<p>Enable Prometheus to be used as a receiver for the Prometheus remote write protocol. Defaults to the value of <code>false</code>.
WARNING: This is not considered an efficient way of ingesting samples.
Use it with caution for specific low-volume use cases.
It is not suitable for replacing the ingestion via scraping and turning
Prometheus into a push-based metrics collection system.
For more information see <a href="https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver">https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver</a>
Only valid in Prometheus versions 2.33.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>enableFeatures</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Enable access to Prometheus disabled features. By default, no features are enabled.
Enabling disabled features is entirely outside the scope of what the maintainers will
support and by doing so, you accept that this behaviour may break at any
time without notice.
For more information see <a href="https://prometheus.io/docs/prometheus/latest/disabled_features/">https://prometheus.io/docs/prometheus/latest/disabled_features/</a></p>
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
<p>The external URL the Prometheus instances will be available under. This is
necessary to generate correct URLs. This is necessary if Prometheus is not
served from root of a DNS name.</p>
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
<p>The route prefix Prometheus registers HTTP handlers for. This is useful,
if using ExternalURL and a proxy is rewriting HTTP routes of a request,
and the actual ExternalURL is still true, but the server serves requests
under a different route prefix. For example for use with <code>kubectl proxy</code>.</p>
</td>
</tr>
<tr>
<td>
<code>storage</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.StorageSpec">
StorageSpec
</a>
</em>
</td>
<td>
<p>Storage spec to specify how storage shall be used.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<p>Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will
be appended to other volumes that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<p>VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
VolumeMounts specified will be appended to other VolumeMounts in the prometheus container,
that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>web</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusWebSpec">
PrometheusWebSpec
</a>
</em>
</td>
<td>
<p>Defines the web command line flags when starting Prometheus.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>Define resources requests and limits for single Pods.</p>
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
<p>Define which Nodes the Pods are scheduled on.</p>
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
<code>secrets</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Secrets is a list of Secrets in the same namespace as the Prometheus
object, which shall be mounted into the Prometheus Pods.
The Secrets are mounted into /etc/prometheus/secrets/<secret-name>.</p>
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
The ConfigMaps are mounted into /etc/prometheus/configmaps/<configmap-name>.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s scheduling constraints.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s topology spread constraints.</p>
</td>
</tr>
<tr>
<td>
<code>remoteWrite</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RemoteWriteSpec">
[]RemoteWriteSpec
</a>
</em>
</td>
<td>
<p>remoteWrite is the list of remote write configurations.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<p>SecurityContext holds pod-level security attributes and common container settings.
This defaults to the default PodSecurityContext.</p>
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
<p>ListenLocal makes the Prometheus server listen on loopback, so that it
does not bind against the Pod IP.</p>
</td>
</tr>
<tr>
<td>
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>Containers allows injecting additional containers or modifying operator
generated containers. This can be used to allow adding an authentication
proxy to a Prometheus pod or to change the behavior of an operator
generated container. Containers described here modify an operator
generated container if they share the same name and modifications are
done via a strategic merge patch. The current container names are:
<code>prometheus</code>, <code>config-reloader</code>, and <code>thanos-sidecar</code>. Overriding
containers is entirely outside the scope of what the maintainers will
support and by doing so, you accept that this behaviour may break at any
time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>initContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
fetch secrets for injection into the Prometheus configuration from external sources. Any errors
during the execution of an initContainer will lead to a restart of the Pod. More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/init-containers/">https://kubernetes.io/docs/concepts/workloads/pods/init-containers/</a>
InitContainers described here modify an operator
generated init containers if they share the same name and modifications are
done via a strategic merge patch. The current init container name is:
<code>init-config-reloader</code>. Overriding init containers is entirely outside the
scope of what the maintainers will support and by doing so, you accept that
this behaviour may break at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>additionalScrapeConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
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
<a href="#monitoring.coreos.com/v1.APIServerConfig">
APIServerConfig
</a>
</em>
</td>
<td>
<p>APIServerConfig allows specifying a host and auth methods to access apiserver.
If left empty, Prometheus is assumed to run inside of the cluster
and will discover API servers automatically and use the pod&rsquo;s CA certificate
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
<p>Priority class assigned to the Pods</p>
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
This defaults to web</p>
</td>
</tr>
<tr>
<td>
<code>arbitraryFSAccessThroughSMs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ArbitraryFSAccessThroughSMsConfig">
ArbitraryFSAccessThroughSMsConfig
</a>
</em>
</td>
<td>
<p>ArbitraryFSAccessThroughSMs configures whether configuration
based on a service monitor can access arbitrary files on the file system
of the Prometheus container e.g. bearer token files.</p>
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
<p>When true, Prometheus resolves label conflicts by renaming the labels in
the scraped data to &ldquo;exported_<label value>&rdquo; for all targets created
from service and pod monitors.
Otherwise the HonorLabels field of the service or pod monitor applies.</p>
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
<p>IgnoreNamespaceSelectors if set to true will ignore NamespaceSelector
settings from all PodMonitor, ServiceMonitor and Probe objects. They will
only discover endpoints within the namespace of the PodMonitor,
ServiceMonitor and Probe objects.
Defaults to false.</p>
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
<p>EnforcedNamespaceLabel If set, a label will be added to</p>
<ol>
<li>all user-metrics (created by <code>ServiceMonitor</code>, <code>PodMonitor</code> and <code>Probe</code> objects) and</li>
<li>in all <code>PrometheusRule</code> objects (except the ones excluded in <code>prometheusRulesExcludedFromEnforce</code>) to
<ul>
<li>alerting &amp; recording rules and</li>
<li>the metrics used in their expressions (<code>expr</code>).</li>
</ul></li>
</ol>
<p>Label name is this field&rsquo;s value.
Label value is the namespace of the created object (mentioned above).</p>
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
<p>EnforcedSampleLimit defines global limit on number of scraped samples
that will be accepted. This overrides any SampleLimit set per
ServiceMonitor or/and PodMonitor. It is meant to be used by admins to
enforce the SampleLimit to keep overall number of samples/series under
the desired limit.
Note that if SampleLimit is lower that value will be taken instead.</p>
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
<p>EnforcedTargetLimit defines a global limit on the number of scraped
targets.  This overrides any TargetLimit set per ServiceMonitor or/and
PodMonitor.  It is meant to be used by admins to enforce the TargetLimit
to keep the overall number of targets under the desired limit.
Note that if TargetLimit is lower, that value will be taken instead,
except if either value is zero, in which case the non-zero value will be
used.  If both values are zero, no limit is enforced.</p>
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
<p>Per-scrape limit on number of labels that will be accepted for a sample. If
more than this number of labels are present post metric-relabeling, the
entire scrape will be treated as failed. 0 means no limit.
Only valid in Prometheus versions 2.27.0 and newer.</p>
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
<p>Per-scrape limit on length of labels name that will be accepted for a sample.
If a label name is longer than this number post metric-relabeling, the entire
scrape will be treated as failed. 0 means no limit.
Only valid in Prometheus versions 2.27.0 and newer.</p>
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
<p>Per-scrape limit on length of labels value that will be accepted for a sample.
If a label value is longer than this number post metric-relabeling, the
entire scrape will be treated as failed. 0 means no limit.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedBodySizeLimit</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
</a>
</em>
</td>
<td>
<p>EnforcedBodySizeLimit defines the maximum size of uncompressed response body
that will be accepted by Prometheus. Targets responding with a body larger than this many bytes
will cause the scrape to fail. Example: 100MB.
If defined, the limit will apply to all service/pod monitors and probes.
This is an experimental feature, this behaviour could
change or be removed in the future.
Only valid in Prometheus versions 2.28.0 and newer.</p>
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
<p>Minimum number of seconds for which a newly created pod should be ready
without any of its container crashing for it to be considered available.
Defaults to 0 (pod will be considered available as soon as it is ready)
This is an alpha field and requires enabling StatefulSetMinReadySeconds feature gate.</p>
</td>
</tr>
<tr>
<td>
<code>hostAliases</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.HostAlias">
[]HostAlias
</a>
</em>
</td>
<td>
<p>Pods&rsquo; hostAliases configuration</p>
</td>
</tr>
<tr>
<td>
<code>additionalArgs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Argument">
[]Argument
</a>
</em>
</td>
<td>
<p>AdditionalArgs allows setting additional arguments for the Prometheus container.
It is intended for e.g. activating hidden flags which are not supported by
the dedicated configuration options yet. The arguments are passed as-is to the
Prometheus container which may cause issues if they are invalid or not supporeted
by the given Prometheus version.
In case of an argument conflict (e.g. an argument which is already set by the
operator itself) or when providing an invalid argument the reconciliation will
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
<p>Enable compression of the write-ahead log using Snappy. This flag is
only available in versions of Prometheus &gt;= 2.11.0.</p>
</td>
</tr>
<tr>
<td>
<code>excludedFromEnforcement</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ObjectReference">
[]ObjectReference
</a>
</em>
</td>
<td>
<p>List of references to PodMonitor, ServiceMonitor, Probe and PrometheusRule objects
to be excluded from enforcing a namespace label of origin.
Applies only if enforcedNamespaceLabel set to true.</p>
</td>
</tr>
<tr>
<td>
<code>baseImage</code><br/>
<em>
string
</em>
</td>
<td>
<p>Base image to use for a Prometheus deployment.
Deprecated: use &lsquo;image&rsquo; instead</p>
</td>
</tr>
<tr>
<td>
<code>tag</code><br/>
<em>
string
</em>
</td>
<td>
<p>Tag of Prometheus container image to be deployed. Defaults to the value of <code>version</code>.
Version is ignored if Tag is set.
Deprecated: use &lsquo;image&rsquo; instead.  The image tag can be specified
as part of the image URL.</p>
</td>
</tr>
<tr>
<td>
<code>sha</code><br/>
<em>
string
</em>
</td>
<td>
<p>SHA of Prometheus container image to be deployed. Defaults to the value of <code>version</code>.
Similar to a tag, but the SHA explicitly deploys an immutable container image.
Version and Tag are ignored if SHA is set.
Deprecated: use &lsquo;image&rsquo; instead.  The image digest can be specified
as part of the image URL.</p>
</td>
</tr>
<tr>
<td>
<code>retention</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Time duration Prometheus shall retain data for. Default is &lsquo;24h&rsquo; if
retentionSize is not set, and must match the regular expression <code>[0-9]+(ms|s|m|h|d|w|y)</code>
(milliseconds seconds minutes hours days weeks years).</p>
</td>
</tr>
<tr>
<td>
<code>retentionSize</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
</a>
</em>
</td>
<td>
<p>Maximum amount of disk space used by blocks.</p>
</td>
</tr>
<tr>
<td>
<code>disableCompaction</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Disable prometheus compaction.</p>
</td>
</tr>
<tr>
<td>
<code>rules</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Rules">
Rules
</a>
</em>
</td>
<td>
<p>/&ndash;rules.*/ command-line arguments.</p>
</td>
</tr>
<tr>
<td>
<code>prometheusRulesExcludedFromEnforce</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusRuleExcludeConfig">
[]PrometheusRuleExcludeConfig
</a>
</em>
</td>
<td>
<p>PrometheusRulesExcludedFromEnforce - list of prometheus rules to be excluded from enforcing
of adding namespace labels. Works only if enforcedNamespaceLabel set to true.
Make sure both ruleNamespace and ruleName are set for each pair.
Deprecated: use excludedFromEnforcement instead.</p>
</td>
</tr>
<tr>
<td>
<code>query</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.QuerySpec">
QuerySpec
</a>
</em>
</td>
<td>
<p>QuerySpec defines the query command line flags when starting Prometheus.</p>
</td>
</tr>
<tr>
<td>
<code>ruleSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>A selector to select which PrometheusRules to mount for loading alerting/recording
rules from. Until (excluding) Prometheus Operator v0.24.0 Prometheus
Operator will migrate any legacy rule ConfigMaps to PrometheusRule custom
resources selected by RuleSelector. Make sure it does not match any config
maps that you do not want to be migrated.</p>
</td>
</tr>
<tr>
<td>
<code>ruleNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to be selected for PrometheusRules discovery. If unspecified, only
the same namespace as the Prometheus object is in is used.</p>
</td>
</tr>
<tr>
<td>
<code>alerting</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertingSpec">
AlertingSpec
</a>
</em>
</td>
<td>
<p>Define details regarding alerting.</p>
</td>
</tr>
<tr>
<td>
<code>remoteRead</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RemoteReadSpec">
[]RemoteReadSpec
</a>
</em>
</td>
<td>
<p>remoteRead is the list of remote read configurations.</p>
</td>
</tr>
<tr>
<td>
<code>additionalAlertRelabelConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>AdditionalAlertRelabelConfigs allows specifying a key of a Secret containing
additional Prometheus alert relabel configurations. Alert relabel configurations
specified are appended to the configurations generated by the Prometheus
Operator. Alert relabel configurations specified must have the form as specified
in the official Prometheus documentation:
<a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs</a>.
As alert relabel configs are appended, the user is responsible to make sure it
is valid. Note that using this feature may expose the possibility to
break upgrades of Prometheus. It is advised to review Prometheus release
notes to ensure that no incompatible alert relabel configs are going to break
Prometheus after the upgrade.</p>
</td>
</tr>
<tr>
<td>
<code>additionalAlertManagerConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>AdditionalAlertManagerConfigs allows specifying a key of a Secret containing
additional Prometheus AlertManager configurations. AlertManager configurations
specified are appended to the configurations generated by the Prometheus
Operator. Job configurations specified must have the form as specified
in the official Prometheus documentation:
<a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config</a>.
As AlertManager configs are appended, the user is responsible to make sure it
is valid. Note that using this feature may expose the possibility to
break upgrades of Prometheus. It is advised to review Prometheus release
notes to ensure that no incompatible AlertManager configs are going to break
Prometheus after the upgrade.</p>
</td>
</tr>
<tr>
<td>
<code>thanos</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ThanosSpec">
ThanosSpec
</a>
</em>
</td>
<td>
<p>Thanos configuration allows configuring various aspects of a Prometheus
server in a Thanos environment.</p>
<p>This section is experimental, it may change significantly without
deprecation notice in any release.</p>
<p>This is experimental and may change significantly without backward
compatibility in any release.</p>
</td>
</tr>
<tr>
<td>
<code>queryLogFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>QueryLogFile specifies the file to which PromQL queries are logged.
If the filename has an empty path, e.g. &lsquo;query.log&rsquo;, prometheus-operator will mount the file into an
emptyDir volume at <code>/var/log/prometheus</code>. If a full path is provided, e.g. /var/log/prometheus/query.log, you must mount a volume
in the specified directory and it must be writable. This is because the prometheus container runs with a read-only root filesystem for security reasons.
Alternatively, the location can be set to a stdout location such as <code>/dev/stdout</code> to log
query information to the default Prometheus log stream.
This is only available in versions of Prometheus &gt;= 2.16.0.
For more details, see the Prometheus docs (<a href="https://prometheus.io/docs/guides/query-log/">https://prometheus.io/docs/guides/query-log/</a>)</p>
</td>
</tr>
<tr>
<td>
<code>allowOverlappingBlocks</code><br/>
<em>
bool
</em>
</td>
<td>
<p>AllowOverlappingBlocks enables vertical compaction and vertical query merge in Prometheus.
This is still experimental in Prometheus so it may change in any upcoming release.</p>
</td>
</tr>
<tr>
<td>
<code>exemplars</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Exemplars">
Exemplars
</a>
</em>
</td>
<td>
<p>Exemplars related settings that are runtime reloadable.
It requires to enable the exemplar storage feature to be effective.</p>
</td>
</tr>
<tr>
<td>
<code>evaluationInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Interval between consecutive evaluations. Default: <code>30s</code></p>
</td>
</tr>
<tr>
<td>
<code>enableAdminAPI</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Enable access to prometheus web admin API. Defaults to the value of <code>false</code>.
WARNING: Enabling the admin APIs enables mutating endpoints, to delete data,
shutdown Prometheus, and more. Enabling this should be done with care and the
user is advised to add additional authentication authorization via a proxy to
ensure only clients authorized to perform these actions can do so.
For more information see <a href="https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis">https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.PrometheusStatus">PrometheusStatus
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.Prometheus">Prometheus</a>)
</p>
<div>
<p>PrometheusStatus is the most recent observed status of the Prometheus cluster.
More info:
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
<code>paused</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Represents whether any actions on the underlying managed objects are
being performed. Only delete actions will be performed.</p>
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
<p>Total number of non-terminated pods targeted by this Prometheus deployment
(their labels match the selector).</p>
</td>
</tr>
<tr>
<td>
<code>updatedReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Total number of non-terminated pods targeted by this Prometheus deployment
that have the desired version spec.</p>
</td>
</tr>
<tr>
<td>
<code>availableReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Total number of available pods (ready for at least minReadySeconds)
targeted by this Prometheus deployment.</p>
</td>
</tr>
<tr>
<td>
<code>unavailableReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Total number of unavailable pods targeted by this Prometheus deployment.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusCondition">
[]PrometheusCondition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The current state of the Prometheus deployment.</p>
</td>
</tr>
<tr>
<td>
<code>shardStatuses</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ShardStatus">
[]ShardStatus
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The list has one entry per shard. Each entry provides a summary of the shard status.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.PrometheusWebSpec">PrometheusWebSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>)
</p>
<div>
<p>PrometheusWebSpec defines the web command line flags when starting Prometheus.</p>
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
<code>tlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.WebTLSConfig">
WebTLSConfig
</a>
</em>
</td>
<td>
<p>Defines the TLS parameters for HTTPS.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.WebHTTPConfig">
WebHTTPConfig
</a>
</em>
</td>
<td>
<p>Defines HTTP parameters for web server.</p>
</td>
</tr>
<tr>
<td>
<code>pageTitle</code><br/>
<em>
string
</em>
</td>
<td>
<p>The prometheus web page title</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.QuerySpec">QuerySpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>)
</p>
<div>
<p>QuerySpec defines the query command line flags when starting Prometheus.</p>
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
<code>lookbackDelta</code><br/>
<em>
string
</em>
</td>
<td>
<p>The delta difference allowed for retrieving metrics during expression evaluations.</p>
</td>
</tr>
<tr>
<td>
<code>maxConcurrency</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Number of concurrent queries that can be run at once.</p>
</td>
</tr>
<tr>
<td>
<code>maxSamples</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Maximum number of samples a single query can load into memory. Note that queries will fail if they would load more samples than this into memory, so this also limits the number of samples a query can return.</p>
</td>
</tr>
<tr>
<td>
<code>timeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Maximum time a query may take before being aborted.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.QueueConfig">QueueConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>)
</p>
<div>
<p>QueueConfig allows the tuning of remote write&rsquo;s queue_config parameters.
This object is referenced in the RemoteWriteSpec object.</p>
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
<code>capacity</code><br/>
<em>
int
</em>
</td>
<td>
<p>Capacity is the number of samples to buffer per shard before we start dropping them.</p>
</td>
</tr>
<tr>
<td>
<code>minShards</code><br/>
<em>
int
</em>
</td>
<td>
<p>MinShards is the minimum number of shards, i.e. amount of concurrency.</p>
</td>
</tr>
<tr>
<td>
<code>maxShards</code><br/>
<em>
int
</em>
</td>
<td>
<p>MaxShards is the maximum number of shards, i.e. amount of concurrency.</p>
</td>
</tr>
<tr>
<td>
<code>maxSamplesPerSend</code><br/>
<em>
int
</em>
</td>
<td>
<p>MaxSamplesPerSend is the maximum number of samples per send.</p>
</td>
</tr>
<tr>
<td>
<code>batchSendDeadline</code><br/>
<em>
string
</em>
</td>
<td>
<p>BatchSendDeadline is the maximum time a sample will wait in buffer.</p>
</td>
</tr>
<tr>
<td>
<code>maxRetries</code><br/>
<em>
int
</em>
</td>
<td>
<p>MaxRetries is the maximum number of times to retry a batch on recoverable errors.</p>
</td>
</tr>
<tr>
<td>
<code>minBackoff</code><br/>
<em>
string
</em>
</td>
<td>
<p>MinBackoff is the initial retry delay. Gets doubled for every retry.</p>
</td>
</tr>
<tr>
<td>
<code>maxBackoff</code><br/>
<em>
string
</em>
</td>
<td>
<p>MaxBackoff is the maximum retry delay.</p>
</td>
</tr>
<tr>
<td>
<code>retryOnRateLimit</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Retry upon receiving a 429 status code from the remote-write storage.
This is experimental feature and might change in the future.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.RelabelConfig">RelabelConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1.ProbeTargetIngress">ProbeTargetIngress</a>, <a href="#monitoring.coreos.com/v1.ProbeTargetStaticConfig">ProbeTargetStaticConfig</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>)
</p>
<div>
<p>RelabelConfig allows dynamic rewriting of the label set, being applied to samples before ingestion.
It defines <code>&lt;metric_relabel_configs&gt;</code>-section of Prometheus configuration.
More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#metric_relabel_configs">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#metric_relabel_configs</a></p>
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
<code>sourceLabels</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.LabelName">
[]LabelName
</a>
</em>
</td>
<td>
<p>The source labels select values from existing labels. Their content is concatenated
using the configured separator and matched against the configured regular expression
for the replace, keep, and drop actions.</p>
</td>
</tr>
<tr>
<td>
<code>separator</code><br/>
<em>
string
</em>
</td>
<td>
<p>Separator placed between concatenated source label values. default is &lsquo;;&rsquo;.</p>
</td>
</tr>
<tr>
<td>
<code>targetLabel</code><br/>
<em>
string
</em>
</td>
<td>
<p>Label to which the resulting value is written in a replace action.
It is mandatory for replace actions. Regex capture groups are available.</p>
</td>
</tr>
<tr>
<td>
<code>regex</code><br/>
<em>
string
</em>
</td>
<td>
<p>Regular expression against which the extracted value is matched. Default is &lsquo;(.*)&rsquo;</p>
</td>
</tr>
<tr>
<td>
<code>modulus</code><br/>
<em>
uint64
</em>
</td>
<td>
<p>Modulus to take of the hash of the source label values.</p>
</td>
</tr>
<tr>
<td>
<code>replacement</code><br/>
<em>
string
</em>
</td>
<td>
<p>Replacement value against which a regex replace is performed if the
regular expression matches. Regex capture groups are available. Default is &lsquo;$1&rsquo;</p>
</td>
</tr>
<tr>
<td>
<code>action</code><br/>
<em>
string
</em>
</td>
<td>
<p>Action to perform based on regex matching. Default is &lsquo;replace&rsquo;.
uppercase and lowercase actions require Prometheus &gt;= 2.36.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>)
</p>
<div>
<p>RemoteReadSpec defines the configuration for Prometheus to read back samples
from a remote endpoint.</p>
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
<p>The URL of the endpoint to query from.</p>
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
<p>The name of the remote read queue, it must be unique if specified. The name
is used in metrics and logging in order to differentiate read
configurations.  Only valid in Prometheus versions 2.15.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>requiredMatchers</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>An optional list of equality matchers which have to be present
in a selector to query the remote read endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>remoteTimeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Timeout for requests to the remote read endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>headers</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>Custom HTTP headers to be sent along with each remote read request.
Be aware that headers that are set by Prometheus itself can&rsquo;t be overwritten.
Only valid in Prometheus versions 2.26.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>readRecent</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Whether reads should be made for queries for time ranges that
the local storage should have complete data for.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
</a>
</em>
</td>
<td>
<p>BasicAuth for the URL.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
</a>
</em>
</td>
<td>
<p>OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>bearerToken</code><br/>
<em>
string
</em>
</td>
<td>
<p>Bearer token for remote read.</p>
</td>
</tr>
<tr>
<td>
<code>bearerTokenFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>File to read bearer token for remote read.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Authorization">
Authorization
</a>
</em>
</td>
<td>
<p>Authorization section for remote read</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TLSConfig">
TLSConfig
</a>
</em>
</td>
<td>
<p>TLS Config to use for remote read.</p>
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
<p>Optional ProxyURL.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>)
</p>
<div>
<p>RemoteWriteSpec defines the configuration to write samples from Prometheus
to a remote endpoint.</p>
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
<p>The URL of the endpoint to send samples to.</p>
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
<p>The name of the remote write queue, it must be unique if specified. The
name is used in metrics and logging in order to differentiate queues.
Only valid in Prometheus versions 2.15.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>sendExemplars</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Enables sending of exemplars over remote write. Note that
exemplar-storage itself must be enabled using the enableFeature option
for exemplars to be scraped in the first place.  Only valid in
Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>remoteTimeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Timeout for requests to the remote write endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>headers</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>Custom HTTP headers to be sent along with each remote write request.
Be aware that headers that are set by Prometheus itself can&rsquo;t be overwritten.
Only valid in Prometheus versions 2.25.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>writeRelabelConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RelabelConfig">
[]RelabelConfig
</a>
</em>
</td>
<td>
<p>The list of remote write relabel configurations.</p>
</td>
</tr>
<tr>
<td>
<code>oauth2</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
</a>
</em>
</td>
<td>
<p>OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
</a>
</em>
</td>
<td>
<p>BasicAuth for the URL.</p>
</td>
</tr>
<tr>
<td>
<code>bearerToken</code><br/>
<em>
string
</em>
</td>
<td>
<p>Bearer token for remote write.</p>
</td>
</tr>
<tr>
<td>
<code>bearerTokenFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>File to read bearer token for remote write.</p>
</td>
</tr>
<tr>
<td>
<code>authorization</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Authorization">
Authorization
</a>
</em>
</td>
<td>
<p>Authorization section for remote write</p>
</td>
</tr>
<tr>
<td>
<code>sigv4</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Sigv4">
Sigv4
</a>
</em>
</td>
<td>
<p>Sigv4 allows to configures AWS&rsquo;s Signature Verification 4</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TLSConfig">
TLSConfig
</a>
</em>
</td>
<td>
<p>TLS Config to use for remote write.</p>
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
<p>Optional ProxyURL.</p>
</td>
</tr>
<tr>
<td>
<code>queueConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.QueueConfig">
QueueConfig
</a>
</em>
</td>
<td>
<p>QueueConfig allows tuning of the remote write queue parameters.</p>
</td>
</tr>
<tr>
<td>
<code>metadataConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.MetadataConfig">
MetadataConfig
</a>
</em>
</td>
<td>
<p>MetadataConfig configures the sending of series metadata to the remote storage.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.Rule">Rule
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.RuleGroup">RuleGroup</a>)
</p>
<div>
<p>Rule describes an alerting or recording rule
See Prometheus documentation: <a href="https://www.prometheus.io/docs/prometheus/latest/configuration/alerting_rules/">alerting</a> or <a href="https://www.prometheus.io/docs/prometheus/latest/configuration/recording_rules/#recording-rules">recording</a> rule</p>
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
<code>record</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>alert</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>expr</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/util/intstr#IntOrString">
k8s.io/apimachinery/pkg/util/intstr.IntOrString
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>for</code><br/>
<em>
string
</em>
</td>
<td>
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
</td>
</tr>
<tr>
<td>
<code>annotations</code><br/>
<em>
map[string]string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.RuleGroup">RuleGroup
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusRuleSpec">PrometheusRuleSpec</a>)
</p>
<div>
<p>RuleGroup is a list of sequentially evaluated recording and alerting rules.
Note: PartialResponseStrategy is only used by ThanosRuler and will
be ignored by Prometheus instances.  Valid values for this field are &lsquo;warn&rsquo;
or &lsquo;abort&rsquo;.  More info: <a href="https://github.com/thanos-io/thanos/blob/main/docs/components/rule.md#partial-response">https://github.com/thanos-io/thanos/blob/main/docs/components/rule.md#partial-response</a></p>
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
</td>
</tr>
<tr>
<td>
<code>interval</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>rules</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Rule">
[]Rule
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>partial_response_strategy</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.Rules">Rules
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>)
</p>
<div>
<p>/&ndash;rules.*/ command-line arguments</p>
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
<code>alert</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RulesAlert">
RulesAlert
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.RulesAlert">RulesAlert
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.Rules">Rules</a>)
</p>
<div>
<p>/&ndash;rules.alert.*/ command-line arguments</p>
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
<code>forOutageTolerance</code><br/>
<em>
string
</em>
</td>
<td>
<p>Max time to tolerate prometheus outage for restoring &lsquo;for&rsquo; state of alert.</p>
</td>
</tr>
<tr>
<td>
<code>forGracePeriod</code><br/>
<em>
string
</em>
</td>
<td>
<p>Minimum duration between alert and restored &lsquo;for&rsquo; state.
This is maintained only for alerts with configured &lsquo;for&rsquo; time greater than grace period.</p>
</td>
</tr>
<tr>
<td>
<code>resendDelay</code><br/>
<em>
string
</em>
</td>
<td>
<p>Minimum amount of time to wait before resending an alert to Alertmanager.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.SafeAuthorization">SafeAuthorization
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>, <a href="#monitoring.coreos.com/v1.Authorization">Authorization</a>, <a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1beta1.HTTPConfig">HTTPConfig</a>)
</p>
<div>
<p>SafeAuthorization specifies a subset of the Authorization struct, that is
safe for use in Endpoints (no CredentialsFile field)</p>
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
<p>Set the authentication type. Defaults to Bearer, Basic will cause an
error</p>
</td>
</tr>
<tr>
<td>
<code>credentials</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>The secret&rsquo;s key that contains the credentials of the request</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.SafeTLSConfig">SafeTLSConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpointTLSConfig">PodMetricsEndpointTLSConfig</a>, <a href="#monitoring.coreos.com/v1.ProbeTLSConfig">ProbeTLSConfig</a>, <a href="#monitoring.coreos.com/v1.TLSConfig">TLSConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.EmailConfig">EmailConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1beta1.EmailConfig">EmailConfig</a>, <a href="#monitoring.coreos.com/v1beta1.HTTPConfig">HTTPConfig</a>)
</p>
<div>
<p>SafeTLSConfig specifies safe TLS configuration parameters.</p>
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
<code>ca</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SecretOrConfigMap">
SecretOrConfigMap
</a>
</em>
</td>
<td>
<p>Struct containing the CA cert to use for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>cert</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SecretOrConfigMap">
SecretOrConfigMap
</a>
</em>
</td>
<td>
<p>Struct containing the client cert file for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>keySecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Secret containing the client key file for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>serverName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Used to verify the hostname for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>insecureSkipVerify</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Disable target certificate validation.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.SecretOrConfigMap">SecretOrConfigMap
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.OAuth2">OAuth2</a>, <a href="#monitoring.coreos.com/v1.SafeTLSConfig">SafeTLSConfig</a>, <a href="#monitoring.coreos.com/v1.WebTLSConfig">WebTLSConfig</a>)
</p>
<div>
<p>SecretOrConfigMap allows to specify data as a Secret or ConfigMap. Fields are mutually exclusive.</p>
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
<code>secret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Secret containing data to use for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>configMap</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#configmapkeyselector-v1-core">
Kubernetes core/v1.ConfigMapKeySelector
</a>
</em>
</td>
<td>
<p>ConfigMap containing data to use for the targets.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.SecretOrConfigMapValidationError">SecretOrConfigMapValidationError
</h3>
<div>
<p>SecretOrConfigMapValidationError is returned by SecretOrConfigMap.Validate()
on semantically invalid configurations.</p>
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
<code>err</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ServiceMonitorSpec">ServiceMonitorSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ServiceMonitor">ServiceMonitor</a>)
</p>
<div>
<p>ServiceMonitorSpec contains specification parameters for a ServiceMonitor.</p>
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
<code>jobLabel</code><br/>
<em>
string
</em>
</td>
<td>
<p>JobLabel selects the label from the associated Kubernetes service which will be used as the <code>job</code> label for all metrics.</p>
<p>For example:
If in <code>ServiceMonitor.spec.jobLabel: foo</code> and in <code>Service.metadata.labels.foo: bar</code>,
then the <code>job=&quot;bar&quot;</code> label is added to all metrics.</p>
<p>If the value of this field is empty or if the label doesn&rsquo;t exist for the given Service, the <code>job</code> label of the metrics defaults to the name of the Kubernetes Service.</p>
</td>
</tr>
<tr>
<td>
<code>targetLabels</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>TargetLabels transfers labels from the Kubernetes <code>Service</code> onto the created metrics.</p>
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
<p>PodTargetLabels transfers labels on the Kubernetes <code>Pod</code> onto the created metrics.</p>
</td>
</tr>
<tr>
<td>
<code>endpoints</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Endpoint">
[]Endpoint
</a>
</em>
</td>
<td>
<p>A list of endpoints allowed as part of this ServiceMonitor.</p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Selector to select Endpoints objects.</p>
</td>
</tr>
<tr>
<td>
<code>namespaceSelector</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.NamespaceSelector">
NamespaceSelector
</a>
</em>
</td>
<td>
<p>Selector to select which namespaces the Kubernetes Endpoints objects are discovered from.</p>
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
<p>Per-scrape limit on length of labels value that will be accepted for a sample.
Only valid in Prometheus versions 2.27.0 and newer.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ShardStatus">ShardStatus
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusStatus">PrometheusStatus</a>)
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
<code>shardID</code><br/>
<em>
string
</em>
</td>
<td>
<p>Identifier of the shard.</p>
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
<p>Total number of pods targeted by this shard.</p>
</td>
</tr>
<tr>
<td>
<code>updatedReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Total number of non-terminated pods targeted by this shard
that have the desired spec.</p>
</td>
</tr>
<tr>
<td>
<code>availableReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Total number of available pods (ready for at least minReadySeconds)
targeted by this shard.</p>
</td>
</tr>
<tr>
<td>
<code>unavailableReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Total number of unavailable pods targeted by this shard.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.Sigv4">Sigv4
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.SNSConfig">SNSConfig</a>, <a href="#monitoring.coreos.com/v1beta1.SNSConfig">SNSConfig</a>)
</p>
<div>
<p>Sigv4 optionally configures AWS&rsquo;s Signature Verification 4 signing process to
sign requests. Cannot be set at the same time as basic_auth or authorization.</p>
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
<p>Region is the AWS region. If blank, the region from the default credentials chain used.</p>
</td>
</tr>
<tr>
<td>
<code>accessKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>AccessKey is the AWS API key. If blank, the environment variable <code>AWS_ACCESS_KEY_ID</code> is used.</p>
</td>
</tr>
<tr>
<td>
<code>secretKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>SecretKey is the AWS API secret. If blank, the environment variable <code>AWS_SECRET_ACCESS_KEY</code> is used.</p>
</td>
</tr>
<tr>
<td>
<code>profile</code><br/>
<em>
string
</em>
</td>
<td>
<p>Profile is the named AWS profile used to authenticate.</p>
</td>
</tr>
<tr>
<td>
<code>roleArn</code><br/>
<em>
string
</em>
</td>
<td>
<p>RoleArn is the named AWS profile used to authenticate.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.StorageSpec">StorageSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerSpec">AlertmanagerSpec</a>, <a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>)
</p>
<div>
<p>StorageSpec defines the configured storage for a group Prometheus servers.
If no storage option is specified, then by default an <a href="https://kubernetes.io/docs/concepts/storage/volumes/#emptydir">EmptyDir</a> will be used.
If multiple storage options are specified, priority will be given as follows: EmptyDir, Ephemeral, and lastly VolumeClaimTemplate.</p>
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
<code>disableMountSubPath</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Deprecated: subPath usage will be disabled by default in a future release, this option will become unnecessary.
DisableMountSubPath allows to remove any subPath usage in volume mounts.</p>
</td>
</tr>
<tr>
<td>
<code>emptyDir</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#emptydirvolumesource-v1-core">
Kubernetes core/v1.EmptyDirVolumeSource
</a>
</em>
</td>
<td>
<p>EmptyDirVolumeSource to be used by the Prometheus StatefulSets. If specified, used in place of any volumeClaimTemplate. More
info: <a href="https://kubernetes.io/docs/concepts/storage/volumes/#emptydir">https://kubernetes.io/docs/concepts/storage/volumes/#emptydir</a></p>
</td>
</tr>
<tr>
<td>
<code>ephemeral</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#ephemeralvolumesource-v1-core">
Kubernetes core/v1.EphemeralVolumeSource
</a>
</em>
</td>
<td>
<p>EphemeralVolumeSource to be used by the Prometheus StatefulSets.
This is a beta field in k8s 1.21, for lower versions, starting with k8s 1.19, it requires enabling the GenericEphemeralVolume feature gate.
More info: <a href="https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/#generic-ephemeral-volumes">https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/#generic-ephemeral-volumes</a></p>
</td>
</tr>
<tr>
<td>
<code>volumeClaimTemplate</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.EmbeddedPersistentVolumeClaim">
EmbeddedPersistentVolumeClaim
</a>
</em>
</td>
<td>
<p>A PVC spec to be used by the Prometheus StatefulSets.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.TLSConfig">TLSConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.APIServerConfig">APIServerConfig</a>, <a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>, <a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosSpec">ThanosSpec</a>)
</p>
<div>
<p>TLSConfig extends the safe TLS configuration with file parameters.</p>
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
<code>ca</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SecretOrConfigMap">
SecretOrConfigMap
</a>
</em>
</td>
<td>
<p>Struct containing the CA cert to use for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>cert</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SecretOrConfigMap">
SecretOrConfigMap
</a>
</em>
</td>
<td>
<p>Struct containing the client cert file for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>keySecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Secret containing the client key file for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>serverName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Used to verify the hostname for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>insecureSkipVerify</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Disable target certificate validation.</p>
</td>
</tr>
<tr>
<td>
<code>caFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>Path to the CA cert in the Prometheus container to use for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>certFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>Path to the client cert file in the Prometheus container for the targets.</p>
</td>
</tr>
<tr>
<td>
<code>keyFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>Path to the client key file in the Prometheus container for the targets.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.TLSConfigValidationError">TLSConfigValidationError
</h3>
<div>
<p>TLSConfigValidationError is returned by TLSConfig.Validate() on semantically
invalid tls configurations.</p>
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
<code>err</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ThanosRuler">ThanosRuler
</h3>
<div>
<p>ThanosRuler defines a ThanosRuler deployment.</p>
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
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
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
<a href="#monitoring.coreos.com/v1.ThanosRulerSpec">
ThanosRulerSpec
</a>
</em>
</td>
<td>
<p>Specification of the desired behavior of the ThanosRuler cluster. More info:
<a href="https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status">https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status</a></p>
<br/>
<br/>
<table>
<tr>
<td>
<code>podMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
EmbeddedObjectMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata contains Labels and Annotations gets propagated to the thanos ruler pods.</p>
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
<p>Thanos container image URL.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>An optional list of references to secrets in the same namespace
to use for pulling thanos images from registries
see <a href="http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod">http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod</a></p>
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
<p>When a ThanosRuler deployment is paused, no actions except for deletion
will be performed on the underlying objects.</p>
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
<p>Number of thanos ruler instances to deploy.</p>
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
<p>Define which Nodes the Pods are scheduled on.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>Resources defines the resource requirements for single Pods.
If not provided, no requests/limits will be set</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s scheduling constraints.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s topology spread constraints.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<p>SecurityContext holds pod-level security attributes and common container settings.
This defaults to the default PodSecurityContext.</p>
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
<p>Priority class assigned to the Pods</p>
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
Thanos Ruler Pods.</p>
</td>
</tr>
<tr>
<td>
<code>storage</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.StorageSpec">
StorageSpec
</a>
</em>
</td>
<td>
<p>Storage spec to specify how storage shall be used.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<p>Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will
be appended to other volumes that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>objectStorageConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>ObjectStorageConfig configures object storage in Thanos.
Alternative to ObjectStorageConfigFile, and lower order priority.</p>
</td>
</tr>
<tr>
<td>
<code>objectStorageConfigFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>ObjectStorageConfigFile specifies the path of the object storage configuration file.
When used alongside with ObjectStorageConfig, ObjectStorageConfigFile takes precedence.</p>
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
<p>ListenLocal makes the Thanos ruler listen on loopback, so that it
does not bind against the Pod IP.</p>
</td>
</tr>
<tr>
<td>
<code>queryEndpoints</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>QueryEndpoints defines Thanos querier endpoints from which to query metrics.
Maps to the &ndash;query flag of thanos ruler.</p>
</td>
</tr>
<tr>
<td>
<code>queryConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Define configuration for connecting to thanos query instances.
If this is defined, the QueryEndpoints field will be ignored.
Maps to the <code>query.config</code> CLI argument.
Only available with thanos v0.11.0 and higher.</p>
</td>
</tr>
<tr>
<td>
<code>alertmanagersUrl</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Define URLs to send alerts to Alertmanager.  For Thanos v0.10.0 and higher,
AlertManagersConfig should be used instead.  Note: this field will be ignored
if AlertManagersConfig is specified.
Maps to the <code>alertmanagers.url</code> arg.</p>
</td>
</tr>
<tr>
<td>
<code>alertmanagersConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Define configuration for connecting to alertmanager.  Only available with thanos v0.10.0
and higher.  Maps to the <code>alertmanagers.config</code> arg.</p>
</td>
</tr>
<tr>
<td>
<code>ruleSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>A label selector to select which PrometheusRules to mount for alerting and
recording.</p>
</td>
</tr>
<tr>
<td>
<code>ruleNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to be selected for Rules discovery. If unspecified, only
the same namespace as the ThanosRuler object is in is used.</p>
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
<p>EnforcedNamespaceLabel enforces adding a namespace label of origin for each alert
and metric that is user created. The label value will always be the namespace of the object that is
being created.</p>
</td>
</tr>
<tr>
<td>
<code>excludedFromEnforcement</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ObjectReference">
[]ObjectReference
</a>
</em>
</td>
<td>
<p>List of references to PrometheusRule objects
to be excluded from enforcing a namespace label of origin.
Applies only if enforcedNamespaceLabel set to true.</p>
</td>
</tr>
<tr>
<td>
<code>prometheusRulesExcludedFromEnforce</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusRuleExcludeConfig">
[]PrometheusRuleExcludeConfig
</a>
</em>
</td>
<td>
<p>PrometheusRulesExcludedFromEnforce - list of Prometheus rules to be excluded from enforcing
of adding namespace labels. Works only if enforcedNamespaceLabel set to true.
Make sure both ruleNamespace and ruleName are set for each pair
Deprecated: use excludedFromEnforcement instead.</p>
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
<p>Log level for ThanosRuler to be configured with.</p>
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
<p>Log format for ThanosRuler to be configured with.</p>
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
This defaults to web</p>
</td>
</tr>
<tr>
<td>
<code>evaluationInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Interval between consecutive evaluations.</p>
</td>
</tr>
<tr>
<td>
<code>retention</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Time duration ThanosRuler shall retain data for. Default is &lsquo;24h&rsquo;,
and must match the regular expression <code>[0-9]+(ms|s|m|h|d|w|y)</code> (milliseconds seconds minutes hours days weeks years).</p>
</td>
</tr>
<tr>
<td>
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>Containers allows injecting additional containers or modifying operator generated
containers. This can be used to allow adding an authentication proxy to a ThanosRuler pod or
to change the behavior of an operator generated container. Containers described here modify
an operator generated container if they share the same name and modifications are done via a
strategic merge patch. The current container names are: <code>thanos-ruler</code> and <code>config-reloader</code>.
Overriding containers is entirely outside the scope of what the maintainers will support and by doing
so, you accept that this behaviour may break at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>initContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
fetch secrets for injection into the ThanosRuler configuration from external sources. Any
errors during the execution of an initContainer will lead to a restart of the Pod.
More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/init-containers/">https://kubernetes.io/docs/concepts/workloads/pods/init-containers/</a>
Using initContainers for any use case other then secret fetching is entirely outside the scope
of what the maintainers will support and by doing so, you accept that this behaviour may break
at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>tracingConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>TracingConfig configures tracing in Thanos. This is an experimental feature, it may change in any upcoming release in a breaking way.</p>
</td>
</tr>
<tr>
<td>
<code>tracingConfigFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>TracingConfig specifies the path of the tracing configuration file.
When used alongside with TracingConfig, TracingConfigFile takes precedence.</p>
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
<p>Labels configure the external label pairs to ThanosRuler. A default replica label
<code>thanos_ruler_replica</code> will be always added  as a label with the value of the pod&rsquo;s name and it will be dropped in the alerts.</p>
</td>
</tr>
<tr>
<td>
<code>alertDropLabels</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>AlertDropLabels configure the label names which should be dropped in ThanosRuler alerts.
The replica label <code>thanos_ruler_replica</code> will always be dropped in alerts.</p>
</td>
</tr>
<tr>
<td>
<code>externalPrefix</code><br/>
<em>
string
</em>
</td>
<td>
<p>The external URL the Thanos Ruler instances will be available under. This is
necessary to generate correct URLs. This is necessary if Thanos Ruler is not
served from root of a DNS name.</p>
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
<p>The route prefix ThanosRuler registers HTTP handlers for. This allows thanos UI to be served on a sub-path.</p>
</td>
</tr>
<tr>
<td>
<code>grpcServerTlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TLSConfig">
TLSConfig
</a>
</em>
</td>
<td>
<p>GRPCServerTLSConfig configures the gRPC server from which Thanos Querier reads
recorded rule data.
Note: Currently only the CAFile, CertFile, and KeyFile fields are supported.
Maps to the &lsquo;&ndash;grpc-server-tls-*&rsquo; CLI args.</p>
</td>
</tr>
<tr>
<td>
<code>alertQueryUrl</code><br/>
<em>
string
</em>
</td>
<td>
<p>The external Query URL the Thanos Ruler will set in the &lsquo;Source&rsquo; field
of all alerts.
Maps to the &lsquo;&ndash;alert.query-url&rsquo; CLI arg.</p>
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
<p>Minimum number of seconds for which a newly created pod should be ready
without any of its container crashing for it to be considered available.
Defaults to 0 (pod will be considered available as soon as it is ready)
This is an alpha field and requires enabling StatefulSetMinReadySeconds feature gate.</p>
</td>
</tr>
<tr>
<td>
<code>alertRelabelConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>AlertRelabelConfigs configures alert relabeling in ThanosRuler.
Alert relabel configurations must have the form as specified in the official Prometheus documentation:
<a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs</a>
Alternative to AlertRelabelConfigFile, and lower order priority.</p>
</td>
</tr>
<tr>
<td>
<code>alertRelabelConfigFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>AlertRelabelConfigFile specifies the path of the alert relabeling configuration file.
When used alongside with AlertRelabelConfigs, alertRelabelConfigFile takes precedence.</p>
</td>
</tr>
<tr>
<td>
<code>hostAliases</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.HostAlias">
[]HostAlias
</a>
</em>
</td>
<td>
<p>Pods&rsquo; hostAliases configuration</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ThanosRulerStatus">
ThanosRulerStatus
</a>
</em>
</td>
<td>
<p>Most recent observed status of the ThanosRuler cluster. Read-only. Not
included when requesting from the apiserver, only from the ThanosRuler
Operator API itself. More info:
<a href="https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status">https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ThanosRuler">ThanosRuler</a>)
</p>
<div>
<p>ThanosRulerSpec is a specification of the desired behavior of the ThanosRuler. More info:
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
<code>podMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
EmbeddedObjectMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata contains Labels and Annotations gets propagated to the thanos ruler pods.</p>
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
<p>Thanos container image URL.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>An optional list of references to secrets in the same namespace
to use for pulling thanos images from registries
see <a href="http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod">http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod</a></p>
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
<p>When a ThanosRuler deployment is paused, no actions except for deletion
will be performed on the underlying objects.</p>
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
<p>Number of thanos ruler instances to deploy.</p>
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
<p>Define which Nodes the Pods are scheduled on.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>Resources defines the resource requirements for single Pods.
If not provided, no requests/limits will be set</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s scheduling constraints.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<p>If specified, the pod&rsquo;s topology spread constraints.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<p>SecurityContext holds pod-level security attributes and common container settings.
This defaults to the default PodSecurityContext.</p>
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
<p>Priority class assigned to the Pods</p>
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
Thanos Ruler Pods.</p>
</td>
</tr>
<tr>
<td>
<code>storage</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.StorageSpec">
StorageSpec
</a>
</em>
</td>
<td>
<p>Storage spec to specify how storage shall be used.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<p>Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will
be appended to other volumes that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>objectStorageConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>ObjectStorageConfig configures object storage in Thanos.
Alternative to ObjectStorageConfigFile, and lower order priority.</p>
</td>
</tr>
<tr>
<td>
<code>objectStorageConfigFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>ObjectStorageConfigFile specifies the path of the object storage configuration file.
When used alongside with ObjectStorageConfig, ObjectStorageConfigFile takes precedence.</p>
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
<p>ListenLocal makes the Thanos ruler listen on loopback, so that it
does not bind against the Pod IP.</p>
</td>
</tr>
<tr>
<td>
<code>queryEndpoints</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>QueryEndpoints defines Thanos querier endpoints from which to query metrics.
Maps to the &ndash;query flag of thanos ruler.</p>
</td>
</tr>
<tr>
<td>
<code>queryConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Define configuration for connecting to thanos query instances.
If this is defined, the QueryEndpoints field will be ignored.
Maps to the <code>query.config</code> CLI argument.
Only available with thanos v0.11.0 and higher.</p>
</td>
</tr>
<tr>
<td>
<code>alertmanagersUrl</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Define URLs to send alerts to Alertmanager.  For Thanos v0.10.0 and higher,
AlertManagersConfig should be used instead.  Note: this field will be ignored
if AlertManagersConfig is specified.
Maps to the <code>alertmanagers.url</code> arg.</p>
</td>
</tr>
<tr>
<td>
<code>alertmanagersConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Define configuration for connecting to alertmanager.  Only available with thanos v0.10.0
and higher.  Maps to the <code>alertmanagers.config</code> arg.</p>
</td>
</tr>
<tr>
<td>
<code>ruleSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>A label selector to select which PrometheusRules to mount for alerting and
recording.</p>
</td>
</tr>
<tr>
<td>
<code>ruleNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to be selected for Rules discovery. If unspecified, only
the same namespace as the ThanosRuler object is in is used.</p>
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
<p>EnforcedNamespaceLabel enforces adding a namespace label of origin for each alert
and metric that is user created. The label value will always be the namespace of the object that is
being created.</p>
</td>
</tr>
<tr>
<td>
<code>excludedFromEnforcement</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ObjectReference">
[]ObjectReference
</a>
</em>
</td>
<td>
<p>List of references to PrometheusRule objects
to be excluded from enforcing a namespace label of origin.
Applies only if enforcedNamespaceLabel set to true.</p>
</td>
</tr>
<tr>
<td>
<code>prometheusRulesExcludedFromEnforce</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusRuleExcludeConfig">
[]PrometheusRuleExcludeConfig
</a>
</em>
</td>
<td>
<p>PrometheusRulesExcludedFromEnforce - list of Prometheus rules to be excluded from enforcing
of adding namespace labels. Works only if enforcedNamespaceLabel set to true.
Make sure both ruleNamespace and ruleName are set for each pair
Deprecated: use excludedFromEnforcement instead.</p>
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
<p>Log level for ThanosRuler to be configured with.</p>
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
<p>Log format for ThanosRuler to be configured with.</p>
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
This defaults to web</p>
</td>
</tr>
<tr>
<td>
<code>evaluationInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Interval between consecutive evaluations.</p>
</td>
</tr>
<tr>
<td>
<code>retention</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Time duration ThanosRuler shall retain data for. Default is &lsquo;24h&rsquo;,
and must match the regular expression <code>[0-9]+(ms|s|m|h|d|w|y)</code> (milliseconds seconds minutes hours days weeks years).</p>
</td>
</tr>
<tr>
<td>
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>Containers allows injecting additional containers or modifying operator generated
containers. This can be used to allow adding an authentication proxy to a ThanosRuler pod or
to change the behavior of an operator generated container. Containers described here modify
an operator generated container if they share the same name and modifications are done via a
strategic merge patch. The current container names are: <code>thanos-ruler</code> and <code>config-reloader</code>.
Overriding containers is entirely outside the scope of what the maintainers will support and by doing
so, you accept that this behaviour may break at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>initContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
fetch secrets for injection into the ThanosRuler configuration from external sources. Any
errors during the execution of an initContainer will lead to a restart of the Pod.
More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/init-containers/">https://kubernetes.io/docs/concepts/workloads/pods/init-containers/</a>
Using initContainers for any use case other then secret fetching is entirely outside the scope
of what the maintainers will support and by doing so, you accept that this behaviour may break
at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>tracingConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>TracingConfig configures tracing in Thanos. This is an experimental feature, it may change in any upcoming release in a breaking way.</p>
</td>
</tr>
<tr>
<td>
<code>tracingConfigFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>TracingConfig specifies the path of the tracing configuration file.
When used alongside with TracingConfig, TracingConfigFile takes precedence.</p>
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
<p>Labels configure the external label pairs to ThanosRuler. A default replica label
<code>thanos_ruler_replica</code> will be always added  as a label with the value of the pod&rsquo;s name and it will be dropped in the alerts.</p>
</td>
</tr>
<tr>
<td>
<code>alertDropLabels</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>AlertDropLabels configure the label names which should be dropped in ThanosRuler alerts.
The replica label <code>thanos_ruler_replica</code> will always be dropped in alerts.</p>
</td>
</tr>
<tr>
<td>
<code>externalPrefix</code><br/>
<em>
string
</em>
</td>
<td>
<p>The external URL the Thanos Ruler instances will be available under. This is
necessary to generate correct URLs. This is necessary if Thanos Ruler is not
served from root of a DNS name.</p>
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
<p>The route prefix ThanosRuler registers HTTP handlers for. This allows thanos UI to be served on a sub-path.</p>
</td>
</tr>
<tr>
<td>
<code>grpcServerTlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TLSConfig">
TLSConfig
</a>
</em>
</td>
<td>
<p>GRPCServerTLSConfig configures the gRPC server from which Thanos Querier reads
recorded rule data.
Note: Currently only the CAFile, CertFile, and KeyFile fields are supported.
Maps to the &lsquo;&ndash;grpc-server-tls-*&rsquo; CLI args.</p>
</td>
</tr>
<tr>
<td>
<code>alertQueryUrl</code><br/>
<em>
string
</em>
</td>
<td>
<p>The external Query URL the Thanos Ruler will set in the &lsquo;Source&rsquo; field
of all alerts.
Maps to the &lsquo;&ndash;alert.query-url&rsquo; CLI arg.</p>
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
<p>Minimum number of seconds for which a newly created pod should be ready
without any of its container crashing for it to be considered available.
Defaults to 0 (pod will be considered available as soon as it is ready)
This is an alpha field and requires enabling StatefulSetMinReadySeconds feature gate.</p>
</td>
</tr>
<tr>
<td>
<code>alertRelabelConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>AlertRelabelConfigs configures alert relabeling in ThanosRuler.
Alert relabel configurations must have the form as specified in the official Prometheus documentation:
<a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs</a>
Alternative to AlertRelabelConfigFile, and lower order priority.</p>
</td>
</tr>
<tr>
<td>
<code>alertRelabelConfigFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>AlertRelabelConfigFile specifies the path of the alert relabeling configuration file.
When used alongside with AlertRelabelConfigs, alertRelabelConfigFile takes precedence.</p>
</td>
</tr>
<tr>
<td>
<code>hostAliases</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.HostAlias">
[]HostAlias
</a>
</em>
</td>
<td>
<p>Pods&rsquo; hostAliases configuration</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ThanosRulerStatus">ThanosRulerStatus
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ThanosRuler">ThanosRuler</a>)
</p>
<div>
<p>ThanosRulerStatus is the most recent observed status of the ThanosRuler. Read-only. Not
included when requesting from the apiserver, only from the Prometheus
Operator API itself. More info:
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
<code>paused</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Represents whether any actions on the underlying managed objects are
being performed. Only delete actions will be performed.</p>
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
<p>Total number of non-terminated pods targeted by this ThanosRuler deployment
(their labels match the selector).</p>
</td>
</tr>
<tr>
<td>
<code>updatedReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Total number of non-terminated pods targeted by this ThanosRuler deployment
that have the desired version spec.</p>
</td>
</tr>
<tr>
<td>
<code>availableReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Total number of available pods (ready for at least minReadySeconds)
targeted by this ThanosRuler deployment.</p>
</td>
</tr>
<tr>
<td>
<code>unavailableReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Total number of unavailable pods targeted by this ThanosRuler deployment.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ThanosSpec">ThanosSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>)
</p>
<div>
<p>ThanosSpec defines parameters for a Prometheus server within a Thanos deployment.</p>
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
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<p>Image if specified has precedence over baseImage, tag and sha
combinations. Specifying the version is still necessary to ensure the
Prometheus Operator knows what version of Thanos is being
configured.</p>
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
<p>Version describes the version of Thanos to use.</p>
</td>
</tr>
<tr>
<td>
<code>tag</code><br/>
<em>
string
</em>
</td>
<td>
<p>Tag of Thanos sidecar container image to be deployed. Defaults to the value of <code>version</code>.
Version is ignored if Tag is set.
Deprecated: use &lsquo;image&rsquo; instead.  The image tag can be specified
as part of the image URL.</p>
</td>
</tr>
<tr>
<td>
<code>sha</code><br/>
<em>
string
</em>
</td>
<td>
<p>SHA of Thanos container image to be deployed. Defaults to the value of <code>version</code>.
Similar to a tag, but the SHA explicitly deploys an immutable container image.
Version and Tag are ignored if SHA is set.
Deprecated: use &lsquo;image&rsquo; instead.  The image digest can be specified
as part of the image URL.</p>
</td>
</tr>
<tr>
<td>
<code>baseImage</code><br/>
<em>
string
</em>
</td>
<td>
<p>Thanos base image if other than default.
Deprecated: use &lsquo;image&rsquo; instead</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>Resources defines the resource requirements for the Thanos sidecar.
If not provided, no requests/limits will be set</p>
</td>
</tr>
<tr>
<td>
<code>objectStorageConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>ObjectStorageConfig configures object storage in Thanos.
Alternative to ObjectStorageConfigFile, and lower order priority.</p>
</td>
</tr>
<tr>
<td>
<code>objectStorageConfigFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>ObjectStorageConfigFile specifies the path of the object storage configuration file.
When used alongside with ObjectStorageConfig, ObjectStorageConfigFile takes precedence.</p>
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
<p>ListenLocal makes the Thanos sidecar listen on loopback, so that it
does not bind against the Pod IP.</p>
</td>
</tr>
<tr>
<td>
<code>tracingConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>TracingConfig configures tracing in Thanos. This is an experimental feature, it may change in any upcoming release in a breaking way.</p>
</td>
</tr>
<tr>
<td>
<code>tracingConfigFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>TracingConfig specifies the path of the tracing configuration file.
When used alongside with TracingConfig, TracingConfigFile takes precedence.</p>
</td>
</tr>
<tr>
<td>
<code>grpcServerTlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TLSConfig">
TLSConfig
</a>
</em>
</td>
<td>
<p>GRPCServerTLSConfig configures the gRPC server from which Thanos Querier reads
recorded rule data.
Note: Currently only the CAFile, CertFile, and KeyFile fields are supported.
Maps to the &lsquo;&ndash;grpc-server-tls-*&rsquo; CLI args.</p>
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
<p>LogLevel for Thanos sidecar to be configured with.</p>
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
<p>LogFormat for Thanos sidecar to be configured with.</p>
</td>
</tr>
<tr>
<td>
<code>minTime</code><br/>
<em>
string
</em>
</td>
<td>
<p>MinTime for Thanos sidecar to be configured with. Option can be a constant time in RFC3339 format or time duration relative to current time, such as -1d or 2h45m. Valid duration units are ms, s, m, h, d, w, y.</p>
</td>
</tr>
<tr>
<td>
<code>readyTimeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>ReadyTimeout is the maximum time Thanos sidecar will wait for Prometheus to start. Eg 10m</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<p>VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
VolumeMounts specified will be appended to other VolumeMounts in the thanos-sidecar container.</p>
</td>
</tr>
<tr>
<td>
<code>additionalArgs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Argument">
[]Argument
</a>
</em>
</td>
<td>
<p>AdditionalArgs allows setting additional arguments for the Thanos container.
The arguments are passed as-is to the Thanos container which may cause issues
if they are invalid or not supporeted the given Thanos version.
In case of an argument conflict (e.g. an argument which is already set by the
operator itself) or when providing an invalid argument the reconciliation will
fail and an error will be logged.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.WebConfigFileFields">WebConfigFileFields
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerWebSpec">AlertmanagerWebSpec</a>, <a href="#monitoring.coreos.com/v1.PrometheusWebSpec">PrometheusWebSpec</a>)
</p>
<div>
<p>WebConfigFileFields defines the file content for &ndash;web.config.file flag.</p>
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
<code>tlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.WebTLSConfig">
WebTLSConfig
</a>
</em>
</td>
<td>
<p>Defines the TLS parameters for HTTPS.</p>
</td>
</tr>
<tr>
<td>
<code>httpConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.WebHTTPConfig">
WebHTTPConfig
</a>
</em>
</td>
<td>
<p>Defines HTTP parameters for web server.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.WebHTTPConfig">WebHTTPConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.WebConfigFileFields">WebConfigFileFields</a>)
</p>
<div>
<p>WebHTTPConfig defines HTTP parameters for web server.</p>
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
<code>http2</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Enable HTTP/2 support. Note that HTTP/2 is only supported with TLS.
When TLSConfig is not configured, HTTP/2 will be disabled.
Whenever the value of the field changes, a rolling update will be triggered.</p>
</td>
</tr>
<tr>
<td>
<code>headers</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.WebHTTPHeaders">
WebHTTPHeaders
</a>
</em>
</td>
<td>
<p>List of headers that can be added to HTTP responses.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.WebHTTPHeaders">WebHTTPHeaders
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.WebHTTPConfig">WebHTTPConfig</a>)
</p>
<div>
<p>WebHTTPHeaders defines the list of headers that can be added to HTTP responses.</p>
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
<code>contentSecurityPolicy</code><br/>
<em>
string
</em>
</td>
<td>
<p>Set the Content-Security-Policy header to HTTP responses.
Unset if blank.</p>
</td>
</tr>
<tr>
<td>
<code>xFrameOptions</code><br/>
<em>
string
</em>
</td>
<td>
<p>Set the X-Frame-Options header to HTTP responses.
Unset if blank. Accepted values are deny and sameorigin.
<a href="https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options">https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options</a></p>
</td>
</tr>
<tr>
<td>
<code>xContentTypeOptions</code><br/>
<em>
string
</em>
</td>
<td>
<p>Set the X-Content-Type-Options header to HTTP responses.
Unset if blank. Accepted value is nosniff.
<a href="https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options">https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options</a></p>
</td>
</tr>
<tr>
<td>
<code>xXSSProtection</code><br/>
<em>
string
</em>
</td>
<td>
<p>Set the X-XSS-Protection header to all responses.
Unset if blank.
<a href="https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-XSS-Protection">https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-XSS-Protection</a></p>
</td>
</tr>
<tr>
<td>
<code>strictTransportSecurity</code><br/>
<em>
string
</em>
</td>
<td>
<p>Set the Strict-Transport-Security header to HTTP responses.
Unset if blank.
Please make sure that you use this with care as this header might force
browsers to load Prometheus and the other applications hosted on the same
domain and subdomains over HTTPS.
<a href="https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security">https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.WebTLSConfig">WebTLSConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.WebConfigFileFields">WebConfigFileFields</a>)
</p>
<div>
<p>WebTLSConfig defines the TLS parameters for HTTPS.</p>
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
<code>keySecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Secret containing the TLS key for the server.</p>
</td>
</tr>
<tr>
<td>
<code>cert</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SecretOrConfigMap">
SecretOrConfigMap
</a>
</em>
</td>
<td>
<p>Contains the TLS certificate for the server.</p>
</td>
</tr>
<tr>
<td>
<code>clientAuthType</code><br/>
<em>
string
</em>
</td>
<td>
<p>Server policy for client authentication. Maps to ClientAuth Policies.
For more detail on clientAuth options:
<a href="https://golang.org/pkg/crypto/tls/#ClientAuthType">https://golang.org/pkg/crypto/tls/#ClientAuthType</a></p>
</td>
</tr>
<tr>
<td>
<code>client_ca</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SecretOrConfigMap">
SecretOrConfigMap
</a>
</em>
</td>
<td>
<p>Contains the CA certificate for client certificate authentication to the server.</p>
</td>
</tr>
<tr>
<td>
<code>minVersion</code><br/>
<em>
string
</em>
</td>
<td>
<p>Minimum TLS version that is acceptable. Defaults to TLS12.</p>
</td>
</tr>
<tr>
<td>
<code>maxVersion</code><br/>
<em>
string
</em>
</td>
<td>
<p>Maximum TLS version that is acceptable. Defaults to TLS13.</p>
</td>
</tr>
<tr>
<td>
<code>cipherSuites</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>List of supported cipher suites for TLS versions up to TLS 1.2. If empty,
Go default cipher suites are used. Available cipher suites are documented
in the go documentation: <a href="https://golang.org/pkg/crypto/tls/#pkg-constants">https://golang.org/pkg/crypto/tls/#pkg-constants</a></p>
</td>
</tr>
<tr>
<td>
<code>preferServerCipherSuites</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Controls whether the server selects the
client&rsquo;s most preferred cipher suite, or the server&rsquo;s most preferred
cipher suite. If true then the server&rsquo;s preference, as expressed in
the order of elements in cipherSuites, is used.</p>
</td>
</tr>
<tr>
<td>
<code>curvePreferences</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Elliptic curves that will be used in an ECDHE handshake, in preference
order. Available curves are documented in the go documentation:
<a href="https://golang.org/pkg/crypto/tls/#CurveID">https://golang.org/pkg/crypto/tls/#CurveID</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.WebTLSConfigError">WebTLSConfigError
</h3>
<div>
<p>WebTLSConfigError is returned by WebTLSConfig.Validate() on
semantically invalid configurations.</p>
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
<code>err</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<hr/>
<h2 id="monitoring.coreos.com/v1alpha1">monitoring.coreos.com/v1alpha1</h2>
Resource Types:
<ul><li>
<a href="#monitoring.coreos.com/v1alpha1.AlertmanagerConfig">AlertmanagerConfig</a>
</li></ul>
<h3 id="monitoring.coreos.com/v1alpha1.AlertmanagerConfig">AlertmanagerConfig
</h3>
<div>
<p>AlertmanagerConfig defines a namespaced AlertmanagerConfig to be aggregated
across multiple namespaces configuring one Alertmanager cluster.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
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
<p>The Alertmanager route definition for alerts matching the resource’s
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
the resource’s namespace.</p>
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
<h3 id="monitoring.coreos.com/v1alpha1.AlertmanagerConfigSpec">AlertmanagerConfigSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.AlertmanagerConfig">AlertmanagerConfig</a>)
</p>
<div>
<p>AlertmanagerConfigSpec is a specification of the desired behavior of the Alertmanager configuration.
By definition, the Alertmanager configuration only applies to alerts for which
the <code>namespace</code> label is equal to the namespace of the AlertmanagerConfig resource.</p>
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
<p>The Alertmanager route definition for alerts matching the resource’s
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
the resource’s namespace.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
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
<h3 id="monitoring.coreos.com/v1alpha1.HTTPConfig">HTTPConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1alpha1.OpsGenieConfig">OpsGenieConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.PagerDutyConfig">PagerDutyConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.PushoverConfig">PushoverConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.SNSConfig">SNSConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.SlackConfig">SlackConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.TelegramConfig">TelegramConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.VictorOpsConfig">VictorOpsConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.WeChatConfig">WeChatConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.WebhookConfig">WebhookConfig</a>)
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
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
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
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
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
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
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
operator enforces that the alert matches the resource’s namespace.</p>
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
resource’s namespace.</p>
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
Deprecated as of AlertManager &gt;= v0.22.0 where a user should use MatchType instead.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>The secret&rsquo;s key that contains the recipient user’s user key.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>token</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>The secret&rsquo;s key that contains the registered application’s API token, see <a href="https://pushover.net/apps">https://pushover.net/apps</a>.
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
<p>List of matchers that the alert’s labels should match. For the first
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
</tbody>
</table>
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
<a href="#monitoring.coreos.com/v1.Sigv4">
Sigv4
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>Telegram bot token
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<h2 id="monitoring.coreos.com/v1beta1">monitoring.coreos.com/v1beta1</h2>
Resource Types:
<ul><li>
<a href="#monitoring.coreos.com/v1beta1.AlertmanagerConfig">AlertmanagerConfig</a>
</li></ul>
<h3 id="monitoring.coreos.com/v1beta1.AlertmanagerConfig">AlertmanagerConfig
</h3>
<div>
<p>AlertmanagerConfig defines a namespaced AlertmanagerConfig to be aggregated
across multiple namespaces configuring one Alertmanager cluster.</p>
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
monitoring.coreos.com/v1beta1
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
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
<a href="#monitoring.coreos.com/v1beta1.AlertmanagerConfigSpec">
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
<a href="#monitoring.coreos.com/v1beta1.Route">
Route
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The Alertmanager route definition for alerts matching the resource’s
namespace. If present, it will be added to the generated Alertmanager
configuration as a first-level route.</p>
</td>
</tr>
<tr>
<td>
<code>receivers</code><br/>
<em>
<a href="#monitoring.coreos.com/v1beta1.Receiver">
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
<a href="#monitoring.coreos.com/v1beta1.InhibitRule">
[]InhibitRule
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of inhibition rules. The rules will only apply to alerts matching
the resource’s namespace.</p>
</td>
</tr>
<tr>
<td>
<code>timeIntervals</code><br/>
<em>
<a href="#monitoring.coreos.com/v1beta1.TimeInterval">
[]TimeInterval
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of TimeInterval specifying when the routes should be muted or active.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1beta1.AlertmanagerConfigSpec">AlertmanagerConfigSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.AlertmanagerConfig">AlertmanagerConfig</a>)
</p>
<div>
<p>AlertmanagerConfigSpec is a specification of the desired behavior of the Alertmanager configuration.
By definition, the Alertmanager configuration only applies to alerts for which
the <code>namespace</code> label is equal to the namespace of the AlertmanagerConfig resource.</p>
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
<a href="#monitoring.coreos.com/v1beta1.Route">
Route
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The Alertmanager route definition for alerts matching the resource’s
namespace. If present, it will be added to the generated Alertmanager
configuration as a first-level route.</p>
</td>
</tr>
<tr>
<td>
<code>receivers</code><br/>
<em>
<a href="#monitoring.coreos.com/v1beta1.Receiver">
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
<a href="#monitoring.coreos.com/v1beta1.InhibitRule">
[]InhibitRule
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of inhibition rules. The rules will only apply to alerts matching
the resource’s namespace.</p>
</td>
</tr>
<tr>
<td>
<code>timeIntervals</code><br/>
<em>
<a href="#monitoring.coreos.com/v1beta1.TimeInterval">
[]TimeInterval
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of TimeInterval specifying when the routes should be muted or active.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1beta1.DayOfMonthRange">DayOfMonthRange
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.TimePeriod">TimePeriod</a>)
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
<h3 id="monitoring.coreos.com/v1beta1.EmailConfig">EmailConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Receiver">Receiver</a>)
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
<a href="#monitoring.coreos.com/v1beta1.SecretKeySelector">
SecretKeySelector
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
<a href="#monitoring.coreos.com/v1beta1.SecretKeySelector">
SecretKeySelector
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
<a href="#monitoring.coreos.com/v1beta1.KeyValue">
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
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
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
<h3 id="monitoring.coreos.com/v1beta1.HTTPConfig">HTTPConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.OpsGenieConfig">OpsGenieConfig</a>, <a href="#monitoring.coreos.com/v1beta1.PagerDutyConfig">PagerDutyConfig</a>, <a href="#monitoring.coreos.com/v1beta1.PushoverConfig">PushoverConfig</a>, <a href="#monitoring.coreos.com/v1beta1.SNSConfig">SNSConfig</a>, <a href="#monitoring.coreos.com/v1beta1.SlackConfig">SlackConfig</a>, <a href="#monitoring.coreos.com/v1beta1.TelegramConfig">TelegramConfig</a>, <a href="#monitoring.coreos.com/v1beta1.VictorOpsConfig">VictorOpsConfig</a>, <a href="#monitoring.coreos.com/v1beta1.WeChatConfig">WeChatConfig</a>, <a href="#monitoring.coreos.com/v1beta1.WebhookConfig">WebhookConfig</a>)
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
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
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
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
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
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1beta1.SecretKeySelector">
SecretKeySelector
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
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
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
<h3 id="monitoring.coreos.com/v1beta1.InhibitRule">InhibitRule
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.AlertmanagerConfigSpec">AlertmanagerConfigSpec</a>)
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
<a href="#monitoring.coreos.com/v1beta1.Matcher">
[]Matcher
</a>
</em>
</td>
<td>
<p>Matchers that have to be fulfilled in the alerts to be muted. The
operator enforces that the alert matches the resource’s namespace.</p>
</td>
</tr>
<tr>
<td>
<code>sourceMatch</code><br/>
<em>
<a href="#monitoring.coreos.com/v1beta1.Matcher">
[]Matcher
</a>
</em>
</td>
<td>
<p>Matchers for which one or more alerts have to exist for the inhibition
to take effect. The operator enforces that the alert matches the
resource’s namespace.</p>
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
<h3 id="monitoring.coreos.com/v1beta1.KeyValue">KeyValue
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.EmailConfig">EmailConfig</a>, <a href="#monitoring.coreos.com/v1beta1.OpsGenieConfig">OpsGenieConfig</a>, <a href="#monitoring.coreos.com/v1beta1.PagerDutyConfig">PagerDutyConfig</a>, <a href="#monitoring.coreos.com/v1beta1.VictorOpsConfig">VictorOpsConfig</a>)
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
<h3 id="monitoring.coreos.com/v1beta1.MatchType">MatchType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Matcher">Matcher</a>)
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
<h3 id="monitoring.coreos.com/v1beta1.Matcher">Matcher
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.InhibitRule">InhibitRule</a>, <a href="#monitoring.coreos.com/v1beta1.Route">Route</a>)
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
<a href="#monitoring.coreos.com/v1beta1.MatchType">
MatchType
</a>
</em>
</td>
<td>
<p>Match operator, one of <code>=</code> (equal to), <code>!=</code> (not equal to), <code>=~</code> (regex
match) or <code>!~</code> (not regex match).
Negative operators (<code>!=</code> and <code>!~</code>) require Alertmanager &gt;= v0.22.0.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1beta1.Month">Month
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
<h3 id="monitoring.coreos.com/v1beta1.MonthRange">MonthRange
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.TimePeriod">TimePeriod</a>)
</p>
<div>
<p>MonthRange is an inclusive range of months of the year beginning in January
Months can be specified by name (e.g &lsquo;January&rsquo;) by numerical month (e.g &lsquo;1&rsquo;) or as an inclusive range (e.g &lsquo;January:March&rsquo;, &lsquo;1:3&rsquo;, &lsquo;1:March&rsquo;)</p>
</div>
<h3 id="monitoring.coreos.com/v1beta1.OpsGenieConfig">OpsGenieConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Receiver">Receiver</a>)
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
<a href="#monitoring.coreos.com/v1beta1.SecretKeySelector">
SecretKeySelector
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
<code>details</code><br/>
<em>
<a href="#monitoring.coreos.com/v1beta1.KeyValue">
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
<a href="#monitoring.coreos.com/v1beta1.OpsGenieConfigResponder">
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
<a href="#monitoring.coreos.com/v1beta1.HTTPConfig">
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
<h3 id="monitoring.coreos.com/v1beta1.OpsGenieConfigResponder">OpsGenieConfigResponder
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.OpsGenieConfig">OpsGenieConfig</a>)
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
<h3 id="monitoring.coreos.com/v1beta1.PagerDutyConfig">PagerDutyConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Receiver">Receiver</a>)
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
<a href="#monitoring.coreos.com/v1beta1.SecretKeySelector">
SecretKeySelector
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
<a href="#monitoring.coreos.com/v1beta1.SecretKeySelector">
SecretKeySelector
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
<a href="#monitoring.coreos.com/v1beta1.KeyValue">
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
<a href="#monitoring.coreos.com/v1beta1.PagerDutyImageConfig">
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
<a href="#monitoring.coreos.com/v1beta1.PagerDutyLinkConfig">
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
<a href="#monitoring.coreos.com/v1beta1.HTTPConfig">
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
<h3 id="monitoring.coreos.com/v1beta1.PagerDutyImageConfig">PagerDutyImageConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.PagerDutyConfig">PagerDutyConfig</a>)
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
<h3 id="monitoring.coreos.com/v1beta1.PagerDutyLinkConfig">PagerDutyLinkConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.PagerDutyConfig">PagerDutyConfig</a>)
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
<h3 id="monitoring.coreos.com/v1beta1.ParsedRange">ParsedRange
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
<h3 id="monitoring.coreos.com/v1beta1.PushoverConfig">PushoverConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Receiver">Receiver</a>)
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
<a href="#monitoring.coreos.com/v1beta1.SecretKeySelector">
SecretKeySelector
</a>
</em>
</td>
<td>
<p>The secret&rsquo;s key that contains the recipient user’s user key.
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
</td>
</tr>
<tr>
<td>
<code>token</code><br/>
<em>
<a href="#monitoring.coreos.com/v1beta1.SecretKeySelector">
SecretKeySelector
</a>
</em>
</td>
<td>
<p>The secret&rsquo;s key that contains the registered application’s API token, see <a href="https://pushover.net/apps">https://pushover.net/apps</a>.
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
<a href="#monitoring.coreos.com/v1beta1.HTTPConfig">
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
<h3 id="monitoring.coreos.com/v1beta1.Receiver">Receiver
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.AlertmanagerConfigSpec">AlertmanagerConfigSpec</a>)
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
<a href="#monitoring.coreos.com/v1beta1.OpsGenieConfig">
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
<a href="#monitoring.coreos.com/v1beta1.PagerDutyConfig">
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
<code>slackConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1beta1.SlackConfig">
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
<a href="#monitoring.coreos.com/v1beta1.WebhookConfig">
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
<a href="#monitoring.coreos.com/v1beta1.WeChatConfig">
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
<a href="#monitoring.coreos.com/v1beta1.EmailConfig">
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
<a href="#monitoring.coreos.com/v1beta1.VictorOpsConfig">
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
<a href="#monitoring.coreos.com/v1beta1.PushoverConfig">
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
<a href="#monitoring.coreos.com/v1beta1.SNSConfig">
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
<a href="#monitoring.coreos.com/v1beta1.TelegramConfig">
[]TelegramConfig
</a>
</em>
</td>
<td>
<p>List of Telegram configurations.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1beta1.Route">Route
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.AlertmanagerConfigSpec">AlertmanagerConfigSpec</a>)
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
<a href="#monitoring.coreos.com/v1beta1.Matcher">
[]Matcher
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of matchers that the alert’s labels should match. For the first
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
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1beta1.SNSConfig">SNSConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Receiver">Receiver</a>)
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
<a href="#monitoring.coreos.com/v1.Sigv4">
Sigv4
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
<a href="#monitoring.coreos.com/v1beta1.HTTPConfig">
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
<h3 id="monitoring.coreos.com/v1beta1.SecretKeySelector">SecretKeySelector
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.EmailConfig">EmailConfig</a>, <a href="#monitoring.coreos.com/v1beta1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1beta1.OpsGenieConfig">OpsGenieConfig</a>, <a href="#monitoring.coreos.com/v1beta1.PagerDutyConfig">PagerDutyConfig</a>, <a href="#monitoring.coreos.com/v1beta1.PushoverConfig">PushoverConfig</a>, <a href="#monitoring.coreos.com/v1beta1.SlackConfig">SlackConfig</a>, <a href="#monitoring.coreos.com/v1beta1.TelegramConfig">TelegramConfig</a>, <a href="#monitoring.coreos.com/v1beta1.VictorOpsConfig">VictorOpsConfig</a>, <a href="#monitoring.coreos.com/v1beta1.WeChatConfig">WeChatConfig</a>, <a href="#monitoring.coreos.com/v1beta1.WebhookConfig">WebhookConfig</a>)
</p>
<div>
<p>SecretKeySelector selects a key of a Secret.</p>
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
<p>The name of the secret in the object&rsquo;s namespace to select from.</p>
</td>
</tr>
<tr>
<td>
<code>key</code><br/>
<em>
string
</em>
</td>
<td>
<p>The key of the secret to select from.  Must be a valid secret key.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1beta1.SlackAction">SlackAction
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.SlackConfig">SlackConfig</a>)
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
<a href="#monitoring.coreos.com/v1beta1.SlackConfirmationField">
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
<h3 id="monitoring.coreos.com/v1beta1.SlackConfig">SlackConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Receiver">Receiver</a>)
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
<a href="#monitoring.coreos.com/v1beta1.SecretKeySelector">
SecretKeySelector
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
<a href="#monitoring.coreos.com/v1beta1.SlackField">
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
<a href="#monitoring.coreos.com/v1beta1.SlackAction">
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
<a href="#monitoring.coreos.com/v1beta1.HTTPConfig">
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
<h3 id="monitoring.coreos.com/v1beta1.SlackConfirmationField">SlackConfirmationField
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.SlackAction">SlackAction</a>)
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
<h3 id="monitoring.coreos.com/v1beta1.SlackField">SlackField
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.SlackConfig">SlackConfig</a>)
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
<h3 id="monitoring.coreos.com/v1beta1.TelegramConfig">TelegramConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Receiver">Receiver</a>)
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
<a href="#monitoring.coreos.com/v1beta1.SecretKeySelector">
SecretKeySelector
</a>
</em>
</td>
<td>
<p>Telegram bot token
The secret needs to be in the same namespace as the AlertmanagerConfig
object and accessible by the Prometheus Operator.</p>
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
<a href="#monitoring.coreos.com/v1beta1.HTTPConfig">
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
<h3 id="monitoring.coreos.com/v1beta1.Time">Time
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.TimeRange">TimeRange</a>)
</p>
<div>
<p>Time defines a time in 24hr format</p>
</div>
<h3 id="monitoring.coreos.com/v1beta1.TimeInterval">TimeInterval
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.AlertmanagerConfigSpec">AlertmanagerConfigSpec</a>)
</p>
<div>
<p>TimeInterval specifies the periods in time when notifications will be muted or active.</p>
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
<p>Name of the time interval.</p>
</td>
</tr>
<tr>
<td>
<code>timeIntervals</code><br/>
<em>
<a href="#monitoring.coreos.com/v1beta1.TimePeriod">
[]TimePeriod
</a>
</em>
</td>
<td>
<p>TimeIntervals is a list of TimePeriod.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1beta1.TimePeriod">TimePeriod
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.TimeInterval">TimeInterval</a>)
</p>
<div>
<p>TimePeriod describes periods of time.</p>
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
<a href="#monitoring.coreos.com/v1beta1.TimeRange">
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
<a href="#monitoring.coreos.com/v1beta1.WeekdayRange">
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
<a href="#monitoring.coreos.com/v1beta1.DayOfMonthRange">
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
<a href="#monitoring.coreos.com/v1beta1.MonthRange">
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
<a href="#monitoring.coreos.com/v1beta1.YearRange">
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
<h3 id="monitoring.coreos.com/v1beta1.TimeRange">TimeRange
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.TimePeriod">TimePeriod</a>)
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
<a href="#monitoring.coreos.com/v1beta1.Time">
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
<a href="#monitoring.coreos.com/v1beta1.Time">
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
<h3 id="monitoring.coreos.com/v1beta1.VictorOpsConfig">VictorOpsConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Receiver">Receiver</a>)
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
<a href="#monitoring.coreos.com/v1beta1.SecretKeySelector">
SecretKeySelector
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
<a href="#monitoring.coreos.com/v1beta1.KeyValue">
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
<a href="#monitoring.coreos.com/v1beta1.HTTPConfig">
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
<h3 id="monitoring.coreos.com/v1beta1.WeChatConfig">WeChatConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Receiver">Receiver</a>)
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
<a href="#monitoring.coreos.com/v1beta1.SecretKeySelector">
SecretKeySelector
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
<a href="#monitoring.coreos.com/v1beta1.HTTPConfig">
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
<h3 id="monitoring.coreos.com/v1beta1.WebhookConfig">WebhookConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Receiver">Receiver</a>)
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
<a href="#monitoring.coreos.com/v1beta1.SecretKeySelector">
SecretKeySelector
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
<a href="#monitoring.coreos.com/v1beta1.HTTPConfig">
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
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1beta1.Weekday">Weekday
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
<h3 id="monitoring.coreos.com/v1beta1.WeekdayRange">WeekdayRange
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.TimePeriod">TimePeriod</a>)
</p>
<div>
<p>WeekdayRange is an inclusive range of days of the week beginning on Sunday
Days can be specified by name (e.g &lsquo;Sunday&rsquo;) or as an inclusive range (e.g &lsquo;Monday:Friday&rsquo;)</p>
</div>
<h3 id="monitoring.coreos.com/v1beta1.YearRange">YearRange
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.TimePeriod">TimePeriod</a>)
</p>
<div>
<p>YearRange is an inclusive range of years</p>
</div>
<hr/>
