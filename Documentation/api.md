---
title: "API reference"
description: "Prometheus operator generated API reference docs"
draft: false
images: []
menu: "operator"
weight: 211
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
</li><li>
<a href="#monitoring.coreos.com/v1.ThanosRuler">ThanosRuler</a>
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
<code>imagePullPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#pullpolicy-v1-core">
Kubernetes core/v1.PullPolicy
</a>
</em>
</td>
<td>
<p>Image pull policy for the &lsquo;alertmanager&rsquo;, &lsquo;init-config-reloader&rsquo; and &lsquo;config-reloader&rsquo; containers.
See <a href="https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy">https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy</a> for more details.</p>
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
Each Secret is added to the StatefulSet definition as a volume named <code>secret-&lt;secret-name&gt;</code>.
The Secrets are mounted into <code>/etc/alertmanager/secrets/&lt;secret-name&gt;</code> in the &lsquo;alertmanager&rsquo; container.</p>
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
Each ConfigMap is added to the StatefulSet definition as a volume named <code>configmap-&lt;configmap-name&gt;</code>.
The ConfigMaps are mounted into <code>/etc/alertmanager/configmaps/&lt;configmap-name&gt;</code> in the &lsquo;alertmanager&rsquo; container.</p>
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
instance. If empty, it defaults to <code>alertmanager-&lt;alertmanager-name&gt;</code>.</p>
<p>The Alertmanager configuration should be available under the
<code>alertmanager.yaml</code> key. Additional keys from the original secret are
copied to the generated secret and mounted into the
<code>/etc/alertmanager/config</code> directory in the <code>alertmanager</code> container.</p>
<p>If either the secret or the <code>alertmanager.yaml</code> key is missing, the
operator provisions a minimal Alertmanager configuration with one empty
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
Defaults to <code>web</code>.</p>
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
<code>alertmanagerConfigMatcherStrategy</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertmanagerConfigMatcherStrategy">
AlertmanagerConfigMatcherStrategy
</a>
</em>
</td>
<td>
<p>The AlertmanagerConfigMatcherStrategy defines how AlertmanagerConfig objects match the alerts.
In the future more options may be added.</p>
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
This is an alpha field from kubernetes 1.22 until 1.24 which requires enabling the StatefulSetMinReadySeconds feature gate.</p>
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
If the service account has <code>automountServiceAccountToken: true</code>, set the field to <code>false</code> to opt out of automounting API credentials.</p>
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
<p>Most recent observed status of the Alertmanager cluster. Read-only.
More info:
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
<code>attachMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AttachMetadata">
AttachMetadata
</a>
</em>
</td>
<td>
<p>Attaches node metadata to discovered targets.
Requires Prometheus v2.35.0 and above.</p>
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
If not specified, the Prometheus global scrape timeout is used.</p>
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
<p>PodMetadata configures labels and annotations which are propagated to the Prometheus pods.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for ServicedMonitors discovery. An empty label selector
matches all namespaces. A null label selector matches the current
namespace only.</p>
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
<p><em>Experimental</em> PodMonitors to be selected for target discovery. An empty
label selector matches all objects. A null label selector matches no
objects.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for PodMonitors discovery. An empty label selector
matches all namespaces. A null label selector matches the current
namespace only.</p>
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
<p><em>Experimental</em> Probes to be selected for target discovery. An empty
label selector matches all objects. A null label selector matches no
objects.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> Namespaces to match for Probe discovery. An empty label
selector matches all namespaces. A null label selector matches the
current namespace only.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeConfigSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> ScrapeConfigs to be selected for target discovery. An
empty label selector matches all objects. A null label selector matches
no objects.</p>
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
<code>scrapeConfigNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for ScrapeConfig discovery. An empty label selector
matches all namespaces. A null label selector matches the current
current namespace only.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#pullpolicy-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
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
<em>(Optional)</em>
<p>EXPERIMENTAL: Number of shards to distribute targets onto. <code>spec.replicas</code>
multiplied by <code>spec.shards</code> is the total number of Pods created.</p>
<p>Note that scaling down shards will not reshard data onto remaining
instances, it must be manually moved. Increasing shards will not reshard
data either but it will continue to be available from the same
instances. To query globally, use Thanos sidecar and Thanos querier or
remote write data to a central location.</p>
<p>Sharding is performed on the content of the <code>__address__</code> target meta-label
for PodMonitors and ServiceMonitors and <code>__param_target__</code> for Probes.</p>
<p>Default: 1</p>
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Number of seconds to wait until a scrape request times out.</p>
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
<code>enableFeatures</code><br/>
<em>
[]string
</em>
</td>
<td>
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
<a href="#monitoring.coreos.com/v1.StorageSpec">
StorageSpec
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
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
<code>web</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusWebSpec">
PrometheusWebSpec
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
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
<a href="#monitoring.coreos.com/v1.RemoteWriteSpec">
[]RemoteWriteSpec
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
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
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
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<a href="#monitoring.coreos.com/v1.APIServerConfig">
APIServerConfig
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
<a href="#monitoring.coreos.com/v1.ArbitraryFSAccessThroughSMsConfig">
ArbitraryFSAccessThroughSMsConfig
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
<p>When true, <code>spec.namespaceSelector</code> from all PodMonitor, ServiceMonitor
and Probe objects will be ignored. They will only discover targets
within the namespace of the PodMonitor, ServiceMonitor and Probe
objec.</p>
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
<p>When not empty, a label will be added to</p>
<ol>
<li>All metrics scraped from <code>ServiceMonitor</code>, <code>PodMonitor</code>, <code>Probe</code> and <code>ScrapeConfig</code> objects.</li>
<li>All metrics generated from recording rules defined in <code>PrometheusRule</code> objects.</li>
<li>All alerts generated from alerting rules defined in <code>PrometheusRule</code> objects.</li>
<li>All vector selectors of PromQL expressions defined in <code>PrometheusRule</code> objects.</li>
</ol>
<p>The label will not added for objects referenced in <code>spec.excludedFromEnforcement</code>.</p>
<p>The label&rsquo;s name is this field&rsquo;s value.
The label&rsquo;s value is the namespace of the <code>ServiceMonitor</code>,
<code>PodMonitor</code>, <code>Probe</code> or <code>PrometheusRule</code> object.</p>
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
unless <code>spec.sampleLimit</code> is greater than zero and less than than
<code>spec.enforcedSampleLimit</code>.</p>
<p>It is meant to be used by admins to keep the overall number of
samples/series under a desired limit.</p>
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
<p>When defined, enforcedBodySizeLimit specifies a global limit on the size
of uncompressed response body that will be accepted by Prometheus.
Targets responding with a body larger than this many bytes will cause
the scrape to fail.</p>
<p>It requires Prometheus &gt;= v2.28.0.</p>
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
<a href="#monitoring.coreos.com/v1.HostAlias">
[]HostAlias
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
<a href="#monitoring.coreos.com/v1.Argument">
[]Argument
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
<a href="#monitoring.coreos.com/v1.ObjectReference">
[]ObjectReference
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
it (<a href="https://kubernetes.io/docs/concepts/configuration/overview/">https://kubernetes.io/docs/concepts/configuration/overview/</a>).</p>
<p>When hostNetwork is enabled, this will set the DNS policy to
<code>ClusterFirstWithHostNet</code> automatically.</p>
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
<a href="#monitoring.coreos.com/v1.PrometheusTracingConfig">
PrometheusTracingConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>EXPERIMENTAL: TracingConfig configures tracing in Prometheus. This is an
experimental feature, it may change in any upcoming release in a
breaking way.</p>
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
<p><em>Deprecated: use &lsquo;spec.image&rsquo; instead.</em></p>
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
<p><em>Deprecated: use &lsquo;spec.image&rsquo; instead. The image&rsquo;s tag can be specified
as part of the image name.</em></p>
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
<p><em>Deprecated: use &lsquo;spec.image&rsquo; instead. The image&rsquo;s digest can be
specified as part of the image name.</em></p>
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
<p>How long to retain the Prometheus data.</p>
<p>Default: &ldquo;24h&rdquo; if <code>spec.retention</code> and <code>spec.retentionSize</code> are empty.</p>
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
<p>Maximum number of bytes used by the Prometheus data.</p>
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
<p>When true, the Prometheus compaction is disabled.</p>
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
<p>Defines the configuration of the Prometheus rules&rsquo; engine.</p>
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
<em>(Optional)</em>
<p>Defines the list of PrometheusRule objects to which the namespace label
enforcement doesn&rsquo;t apply.
This is only relevant when <code>spec.enforcedNamespaceLabel</code> is set to true.
<em>Deprecated: use <code>spec.excludedFromEnforcement</code> instead.</em></p>
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
<em>(Optional)</em>
<p>PrometheusRule objects to be selected for rule evaluation. An empty
label selector matches all objects. A null label selector matches no
objects.</p>
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
<em>(Optional)</em>
<p>Namespaces to match for PrometheusRule discovery. An empty label selector
matches all namespaces. A null label selector matches the current
namespace only.</p>
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
<em>(Optional)</em>
<p>QuerySpec defines the configuration of the Promethus query service.</p>
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
<em>(Optional)</em>
<p>Defines the settings related to Alertmanager.</p>
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
<em>(Optional)</em>
<p>AdditionalAlertRelabelConfigs specifies a key of a Secret containing
additional Prometheus alert relabel configurations. The alert relabel
configurations are appended to the configuration generated by the
Prometheus Operator. They must be formatted according to the official
Prometheus documentation:</p>
<p><a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs</a></p>
<p>The user is responsible for making sure that the configurations are valid</p>
<p>Note that using this feature may expose the possibility to break
upgrades of Prometheus. It is advised to review Prometheus release notes
to ensure that no incompatible alert relabel configs are going to break
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
<em>(Optional)</em>
<p>AdditionalAlertManagerConfigs specifies a key of a Secret containing
additional Prometheus Alertmanager configurations. The Alertmanager
configurations are appended to the configuration generated by the
Prometheus Operator. They must be formatted according to the official
Prometheus documentation:</p>
<p><a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config</a></p>
<p>The user is responsible for making sure that the configurations are valid</p>
<p>Note that using this feature may expose the possibility to break
upgrades of Prometheus. It is advised to review Prometheus release notes
to ensure that no incompatible AlertManager configs are going to break
Prometheus after the upgrade.</p>
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
<em>(Optional)</em>
<p>Defines the list of remote read configurations.</p>
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
<em>(Optional)</em>
<p>Defines the configuration of the optional Thanos sidecar.</p>
<p>This section is experimental, it may change significantly without
deprecation notice in any release.</p>
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
<p>queryLogFile specifies where the file to which PromQL queries are logged.</p>
<p>If the filename has an empty path, e.g. &lsquo;query.log&rsquo;, The Prometheus Pods
will mount the file into an emptyDir volume at <code>/var/log/prometheus</code>.
If a full path is provided, e.g. &lsquo;/var/log/prometheus/query.log&rsquo;, you
must mount a volume in the specified directory and it must be writable.
This is because the prometheus container runs with a read-only root
filesystem for security reasons.
Alternatively, the location can be set to a standard I/O stream, e.g.
<code>/dev/stdout</code>, to log query information to the default Prometheus log
stream.</p>
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
<p>AllowOverlappingBlocks enables vertical compaction and vertical query
merge in Prometheus.</p>
<p><em>Deprecated: this flag has no effect for Prometheus &gt;= 2.39.0 where overlapping blocks are enabled by default.</em></p>
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
<em>(Optional)</em>
<p>Exemplars related settings that are runtime reloadable.
It requires to enable the <code>exemplar-storage</code> feature flag to be effective.</p>
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
<p>Interval between rule evaluations.
Default: &ldquo;30s&rdquo;</p>
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
<p>Enables access to the Prometheus web admin API.</p>
<p>WARNING: Enabling the admin APIs enables mutating endpoints, to delete data,
shutdown Prometheus, and more. Enabling this should be done with care and the
user is advised to add additional authentication authorization via a proxy to
ensure only clients authorized to perform these actions can do so.</p>
<p>For more information:
<a href="https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis">https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis</a></p>
</td>
</tr>
<tr>
<td>
<code>tsdb</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TSDBSpec">
TSDBSpec
</a>
</em>
</td>
<td>
<p>Defines the runtime reloadable configuration of the timeseries database
(TSDB).</p>
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
<code>attachMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AttachMetadata">
AttachMetadata
</a>
</em>
</td>
<td>
<p>Attaches node metadata to discovered targets.
Requires Prometheus v2.37.0 and above.</p>
</td>
</tr>
</table>
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
<td><code>ThanosRuler</code></td>
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
<code>version</code><br/>
<em>
string
</em>
</td>
<td>
<p>Version of Thanos to be deployed.</p>
</td>
</tr>
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
<code>imagePullPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#pullpolicy-v1-core">
Kubernetes core/v1.PullPolicy
</a>
</em>
</td>
<td>
<p>Image pull policy for the &lsquo;thanos&rsquo;, &lsquo;init-config-reloader&rsquo; and &lsquo;config-reloader&rsquo; containers.
See <a href="https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy">https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy</a> for more details.</p>
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
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<p>VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
VolumeMounts specified will be appended to other VolumeMounts in the ruler container,
that are generated as a result of StorageSpec objects.</p>
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
Defaults to <code>web</code>.</p>
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
This is an alpha field from kubernetes 1.22 until 1.24 which requires enabling the StatefulSetMinReadySeconds feature gate.</p>
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
<p>AdditionalArgs allows setting additional arguments for the ThanosRuler container.
It is intended for e.g. activating hidden flags which are not supported by
the dedicated configuration options yet. The arguments are passed as-is to the
ThanosRuler container which may cause issues if they are invalid or not supported
by the given ThanosRuler version.
In case of an argument conflict (e.g. an argument which is already set by the
operator itself) or when providing an invalid argument the reconciliation will
fail and an error will be logged.</p>
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
<p>Most recent observed status of the ThanosRuler cluster. Read-only.
More info:
<a href="https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status">https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status</a></p>
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
<p>APIServerConfig defines how the Prometheus server connects to the Kubernetes API server.</p>
<p>More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config</a></p>
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
<p>Kubernetes API address consisting of a hostname or IP address followed
by an optional port number.</p>
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
<p>BasicAuth configuration for the API server.</p>
<p>Cannot be set at the same time as <code>authorization</code>, <code>bearerToken</code>, or
<code>bearerTokenFile</code>.</p>
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
<p>Cannot be set at the same time as <code>basicAuth</code>, <code>authorization</code>, or <code>bearerToken</code>.</p>
<p><em>Deprecated: this will be removed in a future release. Prefer using <code>authorization</code>.</em></p>
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
<em>(Optional)</em>
<p>TLS Config to use for the API server.</p>
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
<em>(Optional)</em>
<p>Authorization section for the API server.</p>
<p>Cannot be set at the same time as <code>basicAuth</code>, <code>bearerToken</code>, or
<code>bearerTokenFile</code>.</p>
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
<p><em>Warning: this field shouldn&rsquo;t be used because the token value appears
in clear-text. Prefer using <code>authorization</code>.</em></p>
<p><em>Deprecated: this will be removed in a future release.</em></p>
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
<h3 id="monitoring.coreos.com/v1.AlertmanagerConfigMatcherStrategy">AlertmanagerConfigMatcherStrategy
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerSpec">AlertmanagerSpec</a>)
</p>
<div>
<p>AlertmanagerConfigMatcherStrategy defines the strategy used by AlertmanagerConfig objects to match alerts.</p>
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
<p>If set to <code>OnNamespace</code>, the operator injects a label matcher matching the namespace of the AlertmanagerConfig object for all its routes and inhibition rules.
<code>None</code> will not add any additional matchers other than the ones specified in the AlertmanagerConfig.
Default is <code>OnNamespace</code>.</p>
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
<tr>
<td>
<code>templates</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SecretOrConfigMap">
[]SecretOrConfigMap
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Custom notification templates.</p>
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
containing Alertmanager IPs to fire alerts against.</p>
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
<p>Namespace of the Endpoints object.</p>
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
<p>Name of the Endpoints object in the namespace.</p>
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
<p>Port on which the Alertmanager API is exposed.</p>
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
<em>(Optional)</em>
<p>TLS Config to use for Alertmanager.</p>
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
<p>BasicAuth configuration for Alertmanager.</p>
<p>Cannot be set at the same time as <code>bearerTokenFile</code>, or <code>authorization</code>.</p>
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
<p>File to read bearer token for Alertmanager.</p>
<p>Cannot be set at the same time as <code>basicAuth</code>, or <code>authorization</code>.</p>
<p><em>Deprecated: this will be removed in a future release. Prefer using <code>authorization</code>.</em></p>
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
<em>(Optional)</em>
<p>Authorization section for Alertmanager.</p>
<p>Cannot be set at the same time as <code>basicAuth</code>, or <code>bearerTokenFile</code>.</p>
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
<p>Version of the Alertmanager API that Prometheus uses to send alerts.
It can be &ldquo;v1&rdquo; or &ldquo;v2&rdquo;.</p>
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
<em>(Optional)</em>
<p>Timeout is a per-target Alertmanager timeout when pushing alerts.</p>
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
<em>(Optional)</em>
<p>Whether to enable HTTP2.</p>
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
<code>smtp</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.GlobalSMTPConfig">
GlobalSMTPConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures global SMTP parameters.</p>
</td>
</tr>
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
<tr>
<td>
<code>slackApiUrl</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>The default Slack API URL.</p>
</td>
</tr>
<tr>
<td>
<code>opsGenieApiUrl</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>The default OpsGenie API URL.</p>
</td>
</tr>
<tr>
<td>
<code>opsGenieApiKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>The default OpsGenie API Key.</p>
</td>
</tr>
<tr>
<td>
<code>pagerdutyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<p>The default Pagerduty URL.</p>
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
<code>imagePullPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#pullpolicy-v1-core">
Kubernetes core/v1.PullPolicy
</a>
</em>
</td>
<td>
<p>Image pull policy for the &lsquo;alertmanager&rsquo;, &lsquo;init-config-reloader&rsquo; and &lsquo;config-reloader&rsquo; containers.
See <a href="https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy">https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy</a> for more details.</p>
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
Each Secret is added to the StatefulSet definition as a volume named <code>secret-&lt;secret-name&gt;</code>.
The Secrets are mounted into <code>/etc/alertmanager/secrets/&lt;secret-name&gt;</code> in the &lsquo;alertmanager&rsquo; container.</p>
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
Each ConfigMap is added to the StatefulSet definition as a volume named <code>configmap-&lt;configmap-name&gt;</code>.
The ConfigMaps are mounted into <code>/etc/alertmanager/configmaps/&lt;configmap-name&gt;</code> in the &lsquo;alertmanager&rsquo; container.</p>
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
instance. If empty, it defaults to <code>alertmanager-&lt;alertmanager-name&gt;</code>.</p>
<p>The Alertmanager configuration should be available under the
<code>alertmanager.yaml</code> key. Additional keys from the original secret are
copied to the generated secret and mounted into the
<code>/etc/alertmanager/config</code> directory in the <code>alertmanager</code> container.</p>
<p>If either the secret or the <code>alertmanager.yaml</code> key is missing, the
operator provisions a minimal Alertmanager configuration with one empty
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
Defaults to <code>web</code>.</p>
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
<code>alertmanagerConfigMatcherStrategy</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertmanagerConfigMatcherStrategy">
AlertmanagerConfigMatcherStrategy
</a>
</em>
</td>
<td>
<p>The AlertmanagerConfigMatcherStrategy defines how AlertmanagerConfig objects match the alerts.
In the future more options may be added.</p>
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
This is an alpha field from kubernetes 1.22 until 1.24 which requires enabling the StatefulSetMinReadySeconds feature gate.</p>
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
If the service account has <code>automountServiceAccountToken: true</code>, set the field to <code>false</code> to opt out of automounting API credentials.</p>
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
<p>AlertmanagerStatus is the most recent observed status of the Alertmanager cluster. Read-only.
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
<p>Total number of non-terminated pods targeted by this Alertmanager
object (their labels match the selector).</p>
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
object that have the desired version spec.</p>
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
<p>Total number of unavailable pods targeted by this Alertmanager object.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Condition">
[]Condition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The current state of the Alertmanager object.</p>
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
<tr>
<td>
<code>getConcurrency</code><br/>
<em>
uint32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Maximum number of GET requests processed concurrently. This corresponds to the
Alertmanager&rsquo;s <code>--web.get-concurrency</code> flag.</p>
</td>
</tr>
<tr>
<td>
<code>timeout</code><br/>
<em>
uint32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Timeout for HTTP requests. This corresponds to the Alertmanager&rsquo;s
<code>--web.timeout</code> flag.</p>
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
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosSpec">ThanosSpec</a>)
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
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PodMonitorSpec">PodMonitorSpec</a>, <a href="#monitoring.coreos.com/v1.ServiceMonitorSpec">ServiceMonitorSpec</a>)
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
<p>Defines the authentication type. The value is case-insensitive.</p>
<p>&ldquo;Basic&rdquo; is not a supported value.</p>
<p>Default: &ldquo;Bearer&rdquo;</p>
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
<p>Selects a key of a Secret in the namespace that contains the credentials for authentication.</p>
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
<p>File to read a secret from, mutually exclusive with <code>credentials</code>.</p>
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
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.APIServerConfig">APIServerConfig</a>, <a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>, <a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPSDConfig">HTTPSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>, <a href="#monitoring.coreos.com/v1beta1.HTTPConfig">HTTPConfig</a>)
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
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.PrometheusAgentSpec">PrometheusAgentSpec</a>)
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
<p>PodMetadata configures labels and annotations which are propagated to the Prometheus pods.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for ServicedMonitors discovery. An empty label selector
matches all namespaces. A null label selector matches the current
namespace only.</p>
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
<p><em>Experimental</em> PodMonitors to be selected for target discovery. An empty
label selector matches all objects. A null label selector matches no
objects.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for PodMonitors discovery. An empty label selector
matches all namespaces. A null label selector matches the current
namespace only.</p>
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
<p><em>Experimental</em> Probes to be selected for target discovery. An empty
label selector matches all objects. A null label selector matches no
objects.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> Namespaces to match for Probe discovery. An empty label
selector matches all namespaces. A null label selector matches the
current namespace only.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeConfigSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> ScrapeConfigs to be selected for target discovery. An
empty label selector matches all objects. A null label selector matches
no objects.</p>
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
<code>scrapeConfigNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for ScrapeConfig discovery. An empty label selector
matches all namespaces. A null label selector matches the current
current namespace only.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#pullpolicy-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
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
<em>(Optional)</em>
<p>EXPERIMENTAL: Number of shards to distribute targets onto. <code>spec.replicas</code>
multiplied by <code>spec.shards</code> is the total number of Pods created.</p>
<p>Note that scaling down shards will not reshard data onto remaining
instances, it must be manually moved. Increasing shards will not reshard
data either but it will continue to be available from the same
instances. To query globally, use Thanos sidecar and Thanos querier or
remote write data to a central location.</p>
<p>Sharding is performed on the content of the <code>__address__</code> target meta-label
for PodMonitors and ServiceMonitors and <code>__param_target__</code> for Probes.</p>
<p>Default: 1</p>
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Number of seconds to wait until a scrape request times out.</p>
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
<code>enableFeatures</code><br/>
<em>
[]string
</em>
</td>
<td>
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
<a href="#monitoring.coreos.com/v1.StorageSpec">
StorageSpec
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
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
<code>web</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusWebSpec">
PrometheusWebSpec
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
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
<a href="#monitoring.coreos.com/v1.RemoteWriteSpec">
[]RemoteWriteSpec
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
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
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
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<a href="#monitoring.coreos.com/v1.APIServerConfig">
APIServerConfig
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
<a href="#monitoring.coreos.com/v1.ArbitraryFSAccessThroughSMsConfig">
ArbitraryFSAccessThroughSMsConfig
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
<p>When true, <code>spec.namespaceSelector</code> from all PodMonitor, ServiceMonitor
and Probe objects will be ignored. They will only discover targets
within the namespace of the PodMonitor, ServiceMonitor and Probe
objec.</p>
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
<p>When not empty, a label will be added to</p>
<ol>
<li>All metrics scraped from <code>ServiceMonitor</code>, <code>PodMonitor</code>, <code>Probe</code> and <code>ScrapeConfig</code> objects.</li>
<li>All metrics generated from recording rules defined in <code>PrometheusRule</code> objects.</li>
<li>All alerts generated from alerting rules defined in <code>PrometheusRule</code> objects.</li>
<li>All vector selectors of PromQL expressions defined in <code>PrometheusRule</code> objects.</li>
</ol>
<p>The label will not added for objects referenced in <code>spec.excludedFromEnforcement</code>.</p>
<p>The label&rsquo;s name is this field&rsquo;s value.
The label&rsquo;s value is the namespace of the <code>ServiceMonitor</code>,
<code>PodMonitor</code>, <code>Probe</code> or <code>PrometheusRule</code> object.</p>
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
unless <code>spec.sampleLimit</code> is greater than zero and less than than
<code>spec.enforcedSampleLimit</code>.</p>
<p>It is meant to be used by admins to keep the overall number of
samples/series under a desired limit.</p>
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
<p>When defined, enforcedBodySizeLimit specifies a global limit on the size
of uncompressed response body that will be accepted by Prometheus.
Targets responding with a body larger than this many bytes will cause
the scrape to fail.</p>
<p>It requires Prometheus &gt;= v2.28.0.</p>
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
<a href="#monitoring.coreos.com/v1.HostAlias">
[]HostAlias
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
<a href="#monitoring.coreos.com/v1.Argument">
[]Argument
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
<a href="#monitoring.coreos.com/v1.ObjectReference">
[]ObjectReference
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
it (<a href="https://kubernetes.io/docs/concepts/configuration/overview/">https://kubernetes.io/docs/concepts/configuration/overview/</a>).</p>
<p>When hostNetwork is enabled, this will set the DNS policy to
<code>ClusterFirstWithHostNet</code> automatically.</p>
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
<a href="#monitoring.coreos.com/v1.PrometheusTracingConfig">
PrometheusTracingConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>EXPERIMENTAL: TracingConfig configures tracing in Prometheus. This is an
experimental feature, it may change in any upcoming release in a
breaking way.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.Condition">Condition
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerStatus">AlertmanagerStatus</a>, <a href="#monitoring.coreos.com/v1.PrometheusStatus">PrometheusStatus</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerStatus">ThanosRulerStatus</a>)
</p>
<div>
<p>Condition represents the state of the resources associated with the
Prometheus, Alertmanager or ThanosRuler resource.</p>
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
<a href="#monitoring.coreos.com/v1.ConditionType">
ConditionType
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
<a href="#monitoring.coreos.com/v1.ConditionStatus">
ConditionStatus
</a>
</em>
</td>
<td>
<p>Status of the condition.</p>
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
<tr>
<td>
<code>observedGeneration</code><br/>
<em>
int64
</em>
</td>
<td>
<p>ObservedGeneration represents the .metadata.generation that the
condition was set based upon. For instance, if <code>.metadata.generation</code> is
currently 12, but the <code>.status.conditions[].observedGeneration</code> is 9, the
condition is out of date with respect to the current state of the
instance.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ConditionStatus">ConditionStatus
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.Condition">Condition</a>)
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
<h3 id="monitoring.coreos.com/v1.ConditionType">ConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.Condition">Condition</a>)
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
<td><p>Available indicates whether enough pods are ready to provide the
service.
The possible status values for this condition type are:
- True: all pods are running and ready, the service is fully available.
- Degraded: some pods aren&rsquo;t ready, the service is partially available.
- False: no pods are running, the service is totally unavailable.
- Unknown: the operator couldn&rsquo;t determine the condition status.</p>
</td>
</tr><tr><td><p>&#34;Reconciled&#34;</p></td>
<td><p>Reconciled indicates whether the operator has reconciled the state of
the underlying resources with the object&rsquo;s spec.
The possible status values for this condition type are:
- True: the reconciliation was successful.
- False: the reconciliation failed.
- Unknown: the operator couldn&rsquo;t determine the condition status.</p>
</td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1.Duration">Duration
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>, <a href="#monitoring.coreos.com/v1.AlertmanagerGlobalConfig">AlertmanagerGlobalConfig</a>, <a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.MetadataConfig">MetadataConfig</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>, <a href="#monitoring.coreos.com/v1.PrometheusTracingConfig">PrometheusTracingConfig</a>, <a href="#monitoring.coreos.com/v1.QuerySpec">QuerySpec</a>, <a href="#monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1.Rule">Rule</a>, <a href="#monitoring.coreos.com/v1.RuleGroup">RuleGroup</a>, <a href="#monitoring.coreos.com/v1.TSDBSpec">TSDBSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosSpec">ThanosSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.FileSDConfig">FileSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPSDConfig">HTTPSDConfig</a>)
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
<p>Defines the desired characteristics of a volume requested by a pod author.
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
When the AnyVolumeDataSource feature gate is enabled, dataSource contents will be copied to dataSourceRef,
and dataSourceRef contents will be copied to dataSource when dataSourceRef.namespace is not specified.
If the namespace is specified, then dataSourceRef will not be copied to dataSource.</p>
</td>
</tr>
<tr>
<td>
<code>dataSourceRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#typedobjectreference-v1-core">
Kubernetes core/v1.TypedObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>dataSourceRef specifies the object from which to populate the volume with data, if a non-empty
volume is desired. This may be any object from a non-empty API group (non
core object) or a PersistentVolumeClaim object.
When this field is specified, volume binding will only succeed if the type of
the specified object matches some installed volume populator or dynamic
provisioner.
This field will replace the functionality of the dataSource field and as such
if both fields are non-empty, they must have the same value. For backwards
compatibility, when namespace isn&rsquo;t specified in dataSourceRef,
both fields (dataSource and dataSourceRef) will be set to the same
value automatically if one of them is empty and the other is non-empty.
When namespace is specified in dataSourceRef,
dataSource isn&rsquo;t set to the same value and must be empty.
There are three important differences between dataSource and dataSourceRef:
* While dataSource only allows two specific types of objects, dataSourceRef
allows any non-core object, as well as PersistentVolumeClaim objects.
* While dataSource ignores disallowed values (dropping them), dataSourceRef
preserves all values, and generates an error if a disallowed value is
specified.
* While dataSource only allows local objects, dataSourceRef allows objects
in any namespaces.
(Beta) Using this field requires the AnyVolumeDataSource feature gate to be enabled.
(Alpha) Using the namespace field of dataSourceRef requires the CrossNamespaceVolumeDataSource feature gate to be enabled.</p>
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
<p><em>Deprecated: this field is never set.</em></p>
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
<p>HTTP scheme to use for scraping.
<code>http</code> and <code>https</code> are the expected values unless you rewrite the <code>__scheme__</code> label via relabeling.
If empty, Prometheus uses the default value <code>http</code>.</p>
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
<tr>
<td>
<code>filterRunning</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Drop pods that are not running. (Failed, Succeeded). Enabled by default.
More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase">https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase</a></p>
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
<em>(Optional)</em>
<p>Maximum number of exemplars stored in memory for all series.</p>
<p>exemplar-storage itself must be enabled using the <code>spec.enableFeature</code>
option for exemplars to be scraped in the first place.</p>
<p>If not set, Prometheus uses its default value. A value of zero or less
than zero disables the storage.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.GlobalSMTPConfig">GlobalSMTPConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerGlobalConfig">AlertmanagerGlobalConfig</a>)
</p>
<div>
<p>GlobalSMTPConfig configures global SMTP parameters.
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
<code>from</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The default SMTP From header field.</p>
</td>
</tr>
<tr>
<td>
<code>smartHost</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.HostPort">
HostPort
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The default SMTP smarthost used for sending emails.</p>
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
<p>The default hostname to identify to the SMTP server.</p>
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
<p>SMTP Auth using CRAM-MD5, LOGIN and PLAIN. If empty, Alertmanager doesn&rsquo;t authenticate to the SMTP server.</p>
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
<em>(Optional)</em>
<p>SMTP Auth using LOGIN and PLAIN.</p>
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
<p>SMTP Auth using PLAIN</p>
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
<em>(Optional)</em>
<p>SMTP Auth using CRAM-MD5.</p>
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
<p>The default SMTP TLS requirement.
Note that Go does not support unencrypted connections to remote SMTP endpoints.</p>
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
<h3 id="monitoring.coreos.com/v1.HostPort">HostPort
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.GlobalSMTPConfig">GlobalSMTPConfig</a>)
</p>
<div>
<p>HostPort represents a &ldquo;host:port&rdquo; network address.</p>
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
<p>Defines the host&rsquo;s address, it can be a DNS name or a literal IP address.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
string
</em>
</td>
<td>
<p>Defines the host&rsquo;s port, it can be a literal port number or a port name.</p>
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
<p>Defines whether metric metadata is sent to the remote storage or not.</p>
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
<p>Defines how frequently metric metadata is sent to the remote storage.</p>
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
<h3 id="monitoring.coreos.com/v1.NonEmptyDuration">NonEmptyDuration
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.Rule">Rule</a>)
</p>
<div>
<p>NonEmptyDuration is a valid time duration that can be parsed by Prometheus model.ParseDuration() function.
Compared to Duration,  NonEmptyDuration enforces a minimum length of 1.
Supported units: y, w, d, h, m, s, ms
Examples: <code>30s</code>, <code>1m</code>, <code>1h20m15s</code>, <code>15d</code></p>
</div>
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
<p>Name of the referent. When not set, all resources in the namespace are matched.</p>
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
<p>HTTP scheme to use for scraping.
<code>http</code> and <code>https</code> are the expected values unless you rewrite the <code>__scheme__</code> label via relabeling.
If empty, Prometheus uses the default value <code>http</code>.</p>
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
<tr>
<td>
<code>filterRunning</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Drop pods that are not running. (Failed, Succeeded). Enabled by default.
More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase">https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase</a></p>
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
<p>Certificate authority used when verifying server certificates.</p>
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
<p>Client certificate to present when doing client-authentication.</p>
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
<code>attachMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AttachMetadata">
AttachMetadata
</a>
</em>
</td>
<td>
<p>Attaches node metadata to discovered targets.
Requires Prometheus v2.35.0 and above.</p>
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
If not specified, the Prometheus global scrape timeout is used.</p>
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
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ProbeTLSConfig">ProbeTLSConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>)
</p>
<div>
<p>ProbeTLSConfig specifies TLS configuration parameters for the prober.</p>
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
<p>Certificate authority used when verifying server certificates.</p>
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
<p>Client certificate to present when doing client-authentication.</p>
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
<code>http</code> and <code>https</code> are the expected values unless you rewrite the <code>__scheme__</code> label via relabeling.
If empty, Prometheus uses the default value <code>http</code>.</p>
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
<h3 id="monitoring.coreos.com/v1.PrometheusRuleExcludeConfig">PrometheusRuleExcludeConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>)
</p>
<div>
<p>PrometheusRuleExcludeConfig enables users to configure excluded
PrometheusRule names and their namespaces to be ignored while enforcing
namespace label for alerts and metrics.</p>
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
<p>Namespace of the excluded PrometheusRule object.</p>
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
<p>Name of the excluded PrometheusRule object.</p>
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
<p>PodMetadata configures labels and annotations which are propagated to the Prometheus pods.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for ServicedMonitors discovery. An empty label selector
matches all namespaces. A null label selector matches the current
namespace only.</p>
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
<p><em>Experimental</em> PodMonitors to be selected for target discovery. An empty
label selector matches all objects. A null label selector matches no
objects.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for PodMonitors discovery. An empty label selector
matches all namespaces. A null label selector matches the current
namespace only.</p>
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
<p><em>Experimental</em> Probes to be selected for target discovery. An empty
label selector matches all objects. A null label selector matches no
objects.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> Namespaces to match for Probe discovery. An empty label
selector matches all namespaces. A null label selector matches the
current namespace only.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeConfigSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> ScrapeConfigs to be selected for target discovery. An
empty label selector matches all objects. A null label selector matches
no objects.</p>
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
<code>scrapeConfigNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for ScrapeConfig discovery. An empty label selector
matches all namespaces. A null label selector matches the current
current namespace only.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#pullpolicy-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
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
<em>(Optional)</em>
<p>EXPERIMENTAL: Number of shards to distribute targets onto. <code>spec.replicas</code>
multiplied by <code>spec.shards</code> is the total number of Pods created.</p>
<p>Note that scaling down shards will not reshard data onto remaining
instances, it must be manually moved. Increasing shards will not reshard
data either but it will continue to be available from the same
instances. To query globally, use Thanos sidecar and Thanos querier or
remote write data to a central location.</p>
<p>Sharding is performed on the content of the <code>__address__</code> target meta-label
for PodMonitors and ServiceMonitors and <code>__param_target__</code> for Probes.</p>
<p>Default: 1</p>
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Number of seconds to wait until a scrape request times out.</p>
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
<code>enableFeatures</code><br/>
<em>
[]string
</em>
</td>
<td>
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
<a href="#monitoring.coreos.com/v1.StorageSpec">
StorageSpec
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
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
<code>web</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusWebSpec">
PrometheusWebSpec
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
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
<a href="#monitoring.coreos.com/v1.RemoteWriteSpec">
[]RemoteWriteSpec
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
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
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
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<a href="#monitoring.coreos.com/v1.APIServerConfig">
APIServerConfig
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
<a href="#monitoring.coreos.com/v1.ArbitraryFSAccessThroughSMsConfig">
ArbitraryFSAccessThroughSMsConfig
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
<p>When true, <code>spec.namespaceSelector</code> from all PodMonitor, ServiceMonitor
and Probe objects will be ignored. They will only discover targets
within the namespace of the PodMonitor, ServiceMonitor and Probe
objec.</p>
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
<p>When not empty, a label will be added to</p>
<ol>
<li>All metrics scraped from <code>ServiceMonitor</code>, <code>PodMonitor</code>, <code>Probe</code> and <code>ScrapeConfig</code> objects.</li>
<li>All metrics generated from recording rules defined in <code>PrometheusRule</code> objects.</li>
<li>All alerts generated from alerting rules defined in <code>PrometheusRule</code> objects.</li>
<li>All vector selectors of PromQL expressions defined in <code>PrometheusRule</code> objects.</li>
</ol>
<p>The label will not added for objects referenced in <code>spec.excludedFromEnforcement</code>.</p>
<p>The label&rsquo;s name is this field&rsquo;s value.
The label&rsquo;s value is the namespace of the <code>ServiceMonitor</code>,
<code>PodMonitor</code>, <code>Probe</code> or <code>PrometheusRule</code> object.</p>
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
unless <code>spec.sampleLimit</code> is greater than zero and less than than
<code>spec.enforcedSampleLimit</code>.</p>
<p>It is meant to be used by admins to keep the overall number of
samples/series under a desired limit.</p>
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
<p>When defined, enforcedBodySizeLimit specifies a global limit on the size
of uncompressed response body that will be accepted by Prometheus.
Targets responding with a body larger than this many bytes will cause
the scrape to fail.</p>
<p>It requires Prometheus &gt;= v2.28.0.</p>
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
<a href="#monitoring.coreos.com/v1.HostAlias">
[]HostAlias
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
<a href="#monitoring.coreos.com/v1.Argument">
[]Argument
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
<a href="#monitoring.coreos.com/v1.ObjectReference">
[]ObjectReference
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
it (<a href="https://kubernetes.io/docs/concepts/configuration/overview/">https://kubernetes.io/docs/concepts/configuration/overview/</a>).</p>
<p>When hostNetwork is enabled, this will set the DNS policy to
<code>ClusterFirstWithHostNet</code> automatically.</p>
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
<a href="#monitoring.coreos.com/v1.PrometheusTracingConfig">
PrometheusTracingConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>EXPERIMENTAL: TracingConfig configures tracing in Prometheus. This is an
experimental feature, it may change in any upcoming release in a
breaking way.</p>
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
<p><em>Deprecated: use &lsquo;spec.image&rsquo; instead.</em></p>
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
<p><em>Deprecated: use &lsquo;spec.image&rsquo; instead. The image&rsquo;s tag can be specified
as part of the image name.</em></p>
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
<p><em>Deprecated: use &lsquo;spec.image&rsquo; instead. The image&rsquo;s digest can be
specified as part of the image name.</em></p>
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
<p>How long to retain the Prometheus data.</p>
<p>Default: &ldquo;24h&rdquo; if <code>spec.retention</code> and <code>spec.retentionSize</code> are empty.</p>
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
<p>Maximum number of bytes used by the Prometheus data.</p>
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
<p>When true, the Prometheus compaction is disabled.</p>
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
<p>Defines the configuration of the Prometheus rules&rsquo; engine.</p>
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
<em>(Optional)</em>
<p>Defines the list of PrometheusRule objects to which the namespace label
enforcement doesn&rsquo;t apply.
This is only relevant when <code>spec.enforcedNamespaceLabel</code> is set to true.
<em>Deprecated: use <code>spec.excludedFromEnforcement</code> instead.</em></p>
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
<em>(Optional)</em>
<p>PrometheusRule objects to be selected for rule evaluation. An empty
label selector matches all objects. A null label selector matches no
objects.</p>
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
<em>(Optional)</em>
<p>Namespaces to match for PrometheusRule discovery. An empty label selector
matches all namespaces. A null label selector matches the current
namespace only.</p>
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
<em>(Optional)</em>
<p>QuerySpec defines the configuration of the Promethus query service.</p>
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
<em>(Optional)</em>
<p>Defines the settings related to Alertmanager.</p>
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
<em>(Optional)</em>
<p>AdditionalAlertRelabelConfigs specifies a key of a Secret containing
additional Prometheus alert relabel configurations. The alert relabel
configurations are appended to the configuration generated by the
Prometheus Operator. They must be formatted according to the official
Prometheus documentation:</p>
<p><a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs</a></p>
<p>The user is responsible for making sure that the configurations are valid</p>
<p>Note that using this feature may expose the possibility to break
upgrades of Prometheus. It is advised to review Prometheus release notes
to ensure that no incompatible alert relabel configs are going to break
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
<em>(Optional)</em>
<p>AdditionalAlertManagerConfigs specifies a key of a Secret containing
additional Prometheus Alertmanager configurations. The Alertmanager
configurations are appended to the configuration generated by the
Prometheus Operator. They must be formatted according to the official
Prometheus documentation:</p>
<p><a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config</a></p>
<p>The user is responsible for making sure that the configurations are valid</p>
<p>Note that using this feature may expose the possibility to break
upgrades of Prometheus. It is advised to review Prometheus release notes
to ensure that no incompatible AlertManager configs are going to break
Prometheus after the upgrade.</p>
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
<em>(Optional)</em>
<p>Defines the list of remote read configurations.</p>
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
<em>(Optional)</em>
<p>Defines the configuration of the optional Thanos sidecar.</p>
<p>This section is experimental, it may change significantly without
deprecation notice in any release.</p>
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
<p>queryLogFile specifies where the file to which PromQL queries are logged.</p>
<p>If the filename has an empty path, e.g. &lsquo;query.log&rsquo;, The Prometheus Pods
will mount the file into an emptyDir volume at <code>/var/log/prometheus</code>.
If a full path is provided, e.g. &lsquo;/var/log/prometheus/query.log&rsquo;, you
must mount a volume in the specified directory and it must be writable.
This is because the prometheus container runs with a read-only root
filesystem for security reasons.
Alternatively, the location can be set to a standard I/O stream, e.g.
<code>/dev/stdout</code>, to log query information to the default Prometheus log
stream.</p>
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
<p>AllowOverlappingBlocks enables vertical compaction and vertical query
merge in Prometheus.</p>
<p><em>Deprecated: this flag has no effect for Prometheus &gt;= 2.39.0 where overlapping blocks are enabled by default.</em></p>
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
<em>(Optional)</em>
<p>Exemplars related settings that are runtime reloadable.
It requires to enable the <code>exemplar-storage</code> feature flag to be effective.</p>
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
<p>Interval between rule evaluations.
Default: &ldquo;30s&rdquo;</p>
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
<p>Enables access to the Prometheus web admin API.</p>
<p>WARNING: Enabling the admin APIs enables mutating endpoints, to delete data,
shutdown Prometheus, and more. Enabling this should be done with care and the
user is advised to add additional authentication authorization via a proxy to
ensure only clients authorized to perform these actions can do so.</p>
<p>For more information:
<a href="https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis">https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis</a></p>
</td>
</tr>
<tr>
<td>
<code>tsdb</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TSDBSpec">
TSDBSpec
</a>
</em>
</td>
<td>
<p>Defines the runtime reloadable configuration of the timeseries database
(TSDB).</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.PrometheusStatus">PrometheusStatus
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.Prometheus">Prometheus</a>, <a href="#monitoring.coreos.com/v1alpha1.PrometheusAgent">PrometheusAgent</a>)
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
<a href="#monitoring.coreos.com/v1.Condition">
[]Condition
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
<h3 id="monitoring.coreos.com/v1.PrometheusTracingConfig">PrometheusTracingConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>)
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
<code>clientType</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Client used to export the traces. Supported values are <code>http</code> or <code>grpc</code>.</p>
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
<p>Endpoint to send the traces to. Should be provided in format <host>:<port>.</p>
</td>
</tr>
<tr>
<td>
<code>samplingFraction</code><br/>
<em>
k8s.io/apimachinery/pkg/api/resource.Quantity
</em>
</td>
<td>
<em>(Optional)</em>
<p>Sets the probability a given trace will be sampled. Must be a float from 0 through 1.</p>
</td>
</tr>
<tr>
<td>
<code>insecure</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>If disabled, the client will use a secure connection.</p>
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
<em>(Optional)</em>
<p>Key-value pairs to be used as headers associated with gRPC or HTTP requests.</p>
</td>
</tr>
<tr>
<td>
<code>compression</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Compression key for supported compression types. The only supported value is <code>gzip</code>.</p>
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
<em>(Optional)</em>
<p>Maximum time the exporter will wait for each batch export.</p>
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
<em>(Optional)</em>
<p>TLS Config to use when sending traces.</p>
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
<p>PrometheusWebSpec defines the configuration of the Prometheus web server.</p>
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
<em>(Optional)</em>
<p>The prometheus web page title.</p>
</td>
</tr>
<tr>
<td>
<code>maxConnections</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the maximum number of simultaneous connections
A zero value means that Prometheus doesn&rsquo;t accept any incoming connection.</p>
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
<em>(Optional)</em>
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
<em>(Optional)</em>
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
<em>(Optional)</em>
<p>Maximum number of samples a single query can load into memory. Note that
queries will fail if they would load more samples than this into memory,
so this also limits the number of samples a query can return.</p>
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
<em>(Optional)</em>
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
<p>Capacity is the number of samples to buffer per shard before we start
dropping them.</p>
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
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1.ProbeTargetIngress">ProbeTargetIngress</a>, <a href="#monitoring.coreos.com/v1.ProbeTargetStaticConfig">ProbeTargetStaticConfig</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>RelabelConfig allows dynamic rewriting of the label set for targets, alerts,
scraped samples and remote write samples.</p>
<p>More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config</a></p>
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
<em>(Optional)</em>
<p>The source labels select values from existing labels. Their content is
concatenated using the configured Separator and matched against the
configured regular expression.</p>
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
<p>Separator is the string between concatenated SourceLabels.</p>
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
<p>Label to which the resulting string is written in a replacement.</p>
<p>It is mandatory for <code>Replace</code>, <code>HashMod</code>, <code>Lowercase</code>, <code>Uppercase</code>,
<code>KeepEqual</code> and <code>DropEqual</code> actions.</p>
<p>Regex capture groups are available.</p>
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
<p>Regular expression against which the extracted value is matched.</p>
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
<p>Only applicable when the action is <code>HashMod</code>.</p>
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
<p>Replacement value against which a Replace action is performed if the
regular expression matches.</p>
<p>Regex capture groups are available.</p>
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
<p>Action to perform based on the regex matching.</p>
<p><code>Uppercase</code> and <code>Lowercase</code> actions require Prometheus &gt;= v2.36.0.
<code>DropEqual</code> and <code>KeepEqual</code> actions require Prometheus &gt;= v2.41.0.</p>
<p>Default: &ldquo;Replace&rdquo;</p>
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
<p>The name of the remote read queue, it must be unique if specified. The
name is used in metrics and logging in order to differentiate read
configurations.</p>
<p>It requires Prometheus &gt;= v2.15.0.</p>
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
<em>(Optional)</em>
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
<em>(Optional)</em>
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
<code>oauth2</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>OAuth2 configuration for the URL.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
<p>Cannot be set at the same time as <code>authorization</code>, or <code>basicAuth</code>.</p>
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
<p>BasicAuth configuration for the URL.</p>
<p>Cannot be set at the same time as <code>authorization</code>, or <code>oauth2</code>.</p>
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
<p>File from which to read the bearer token for the URL.</p>
<p><em>Deprecated: this will be removed in a future release. Prefer using <code>authorization</code>.</em></p>
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
<em>(Optional)</em>
<p>Authorization section for the URL.</p>
<p>It requires Prometheus &gt;= v2.26.0.</p>
<p>Cannot be set at the same time as <code>basicAuth</code>, or <code>oauth2</code>.</p>
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
<p><em>Warning: this field shouldn&rsquo;t be used because the token value appears
in clear-text. Prefer using <code>authorization</code>.</em></p>
<p><em>Deprecated: this will be removed in a future release.</em></p>
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
<em>(Optional)</em>
<p>TLS Config to use for the URL.</p>
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
<code>followRedirects</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configure whether HTTP requests follow HTTP 3xx redirects.</p>
<p>It requires Prometheus &gt;= v2.26.0.</p>
</td>
</tr>
<tr>
<td>
<code>filterExternalLabels</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use the external labels as selectors for the remote read endpoint.</p>
<p>It requires Prometheus &gt;= v2.34.0.</p>
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
name is used in metrics and logging in order to differentiate queues.</p>
<p>It requires Prometheus &gt;= v2.15.0.</p>
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
<em>(Optional)</em>
<p>Enables sending of exemplars over remote write. Note that
exemplar-storage itself must be enabled using the <code>spec.enableFeature</code>
option for exemplars to be scraped in the first place.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
</td>
</tr>
<tr>
<td>
<code>sendNativeHistograms</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enables sending of native histograms, also known as sparse histograms
over remote write.</p>
<p>It requires Prometheus &gt;= v2.40.0.</p>
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
<em>(Optional)</em>
<p>Custom HTTP headers to be sent along with each remote write request.
Be aware that headers that are set by Prometheus itself can&rsquo;t be overwritten.</p>
<p>It requires Prometheus &gt;= v2.25.0.</p>
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
<em>(Optional)</em>
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
<em>(Optional)</em>
<p>OAuth2 configuration for the URL.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
<p>Cannot be set at the same time as <code>sigv4</code>, <code>authorization</code>, or <code>basicAuth</code>.</p>
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
<p>BasicAuth configuration for the URL.</p>
<p>Cannot be set at the same time as <code>sigv4</code>, <code>authorization</code>, or <code>oauth2</code>.</p>
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
<p>File from which to read bearer token for the URL.</p>
<p><em>Deprecated: this will be removed in a future release. Prefer using <code>authorization</code>.</em></p>
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
<em>(Optional)</em>
<p>Authorization section for the URL.</p>
<p>It requires Prometheus &gt;= v2.26.0.</p>
<p>Cannot be set at the same time as <code>sigv4</code>, <code>basicAuth</code>, or <code>oauth2</code>.</p>
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
<p>Sigv4 allows to configures AWS&rsquo;s Signature Verification 4 for the URL.</p>
<p>It requires Prometheus &gt;= v2.26.0.</p>
<p>Cannot be set at the same time as <code>authorization</code>, <code>basicAuth</code>, or <code>oauth2</code>.</p>
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
<p><em>Warning: this field shouldn&rsquo;t be used because the token value appears
in clear-text. Prefer using <code>authorization</code>.</em></p>
<p><em>Deprecated: this will be removed in a future release.</em></p>
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
<em>(Optional)</em>
<p>TLS Config to use for the URL.</p>
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
<em>(Optional)</em>
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
<em>(Optional)</em>
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
<p>Name of the time series to output to. Must be a valid metric name.
Only one of <code>record</code> and <code>alert</code> must be set.</p>
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
<p>Name of the alert. Must be a valid label value.
Only one of <code>record</code> and <code>alert</code> must be set.</p>
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
<p>PromQL expression to evaluate.</p>
</td>
</tr>
<tr>
<td>
<code>for</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Alerts are considered firing once they have been returned for this long.</p>
</td>
</tr>
<tr>
<td>
<code>keep_firing_for</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.NonEmptyDuration">
NonEmptyDuration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>KeepFiringFor defines how long an alert will continue firing after the condition that triggered it has cleared.</p>
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
<p>Labels to add or overwrite.</p>
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
<p>Annotations to add to each alert.
Only valid for alerting rules.</p>
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
<p>RuleGroup is a list of sequentially evaluated recording and alerting rules.</p>
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
<p>Name of the rule group.</p>
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
<em>(Optional)</em>
<p>Interval determines how often rules in the group are evaluated.</p>
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
<em>(Optional)</em>
<p>List of alerting and recording rules.</p>
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
<p>PartialResponseStrategy is only used by ThanosRuler and will
be ignored by Prometheus instances.
More info: <a href="https://github.com/thanos-io/thanos/blob/main/docs/components/rule.md#partial-response">https://github.com/thanos-io/thanos/blob/main/docs/components/rule.md#partial-response</a></p>
</td>
</tr>
<tr>
<td>
<code>limit</code><br/>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>Limit the number of alerts an alerting rule and series a recording
rule can produce.
Limit is supported starting with Prometheus &gt;= 2.31 and Thanos Ruler &gt;= 0.24.</p>
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
<p>Defines the parameters of the Prometheus rules&rsquo; engine.</p>
<p>Any update to these parameters trigger a restart of the pods.</p>
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
<p>Max time to tolerate prometheus outage for restoring &lsquo;for&rsquo; state of
alert.</p>
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
<p>Minimum duration between alert and restored &lsquo;for&rsquo; state.</p>
<p>This is maintained only for alerts with a configured &lsquo;for&rsquo; time greater
than the grace period.</p>
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
<p>Minimum amount of time to wait before resending an alert to
Alertmanager.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.SafeAuthorization">SafeAuthorization
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>, <a href="#monitoring.coreos.com/v1.Authorization">Authorization</a>, <a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPSDConfig">HTTPSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>, <a href="#monitoring.coreos.com/v1beta1.HTTPConfig">HTTPConfig</a>)
</p>
<div>
<p>SafeAuthorization specifies a subset of the Authorization struct, that is
safe for use because it doesn&rsquo;t provide access to the Prometheus container&rsquo;s
filesystem.</p>
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
<p>Defines the authentication type. The value is case-insensitive.</p>
<p>&ldquo;Basic&rdquo; is not a supported value.</p>
<p>Default: &ldquo;Bearer&rdquo;</p>
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
<p>Selects a key of a Secret in the namespace that contains the credentials for authentication.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.SafeTLSConfig">SafeTLSConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpointTLSConfig">PodMetricsEndpointTLSConfig</a>, <a href="#monitoring.coreos.com/v1.ProbeTLSConfig">ProbeTLSConfig</a>, <a href="#monitoring.coreos.com/v1.TLSConfig">TLSConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.EmailConfig">EmailConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPSDConfig">HTTPSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>, <a href="#monitoring.coreos.com/v1beta1.EmailConfig">EmailConfig</a>, <a href="#monitoring.coreos.com/v1beta1.HTTPConfig">HTTPConfig</a>)
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
<p>Certificate authority used when verifying server certificates.</p>
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
<p>Client certificate to present when doing client-authentication.</p>
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
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerConfiguration">AlertmanagerConfiguration</a>, <a href="#monitoring.coreos.com/v1.OAuth2">OAuth2</a>, <a href="#monitoring.coreos.com/v1.SafeTLSConfig">SafeTLSConfig</a>, <a href="#monitoring.coreos.com/v1.WebTLSConfig">WebTLSConfig</a>)
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
<code>attachMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AttachMetadata">
AttachMetadata
</a>
</em>
</td>
<td>
<p>Attaches node metadata to discovered targets.
Requires Prometheus v2.37.0 and above.</p>
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
sign requests.</p>
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
<em>(Optional)</em>
<p>AccessKey is the AWS API key. If not specified, the environment variable
<code>AWS_ACCESS_KEY_ID</code> is used.</p>
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
<em>(Optional)</em>
<p>SecretKey is the AWS API secret. If not specified, the environment
variable <code>AWS_SECRET_ACCESS_KEY</code> is used.</p>
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
If no storage option is specified, then by default an <a href="https://kubernetes.io/docs/concepts/storage/volumes/#emptydir">EmptyDir</a> will be used.</p>
<p>If multiple storage options are specified, priority will be given as follows:
1. emptyDir
2. ephemeral
3. volumeClaimTemplate</p>
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
<p><em>Deprecated: subPath usage will be removed in a future release.</em></p>
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
<p>EmptyDirVolumeSource to be used by the StatefulSet.
If specified, it takes precedence over <code>ephemeral</code> and <code>volumeClaimTemplate</code>.
More info: <a href="https://kubernetes.io/docs/concepts/storage/volumes/#emptydir">https://kubernetes.io/docs/concepts/storage/volumes/#emptydir</a></p>
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
<p>EphemeralVolumeSource to be used by the StatefulSet.
This is a beta field in k8s 1.21 and GA in 1.15.
For lower versions, starting with k8s 1.19, it requires enabling the GenericEphemeralVolume feature gate.
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
<p>Defines the PVC spec to be used by the Prometheus StatefulSets.
The easiest way to use a volume that cannot be automatically provisioned
is to use a label selector alongside manually created PersistentVolumes.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.TLSConfig">TLSConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.APIServerConfig">APIServerConfig</a>, <a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>, <a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.PrometheusTracingConfig">PrometheusTracingConfig</a>, <a href="#monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosSpec">ThanosSpec</a>)
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
<p>Certificate authority used when verifying server certificates.</p>
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
<p>Client certificate to present when doing client-authentication.</p>
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
<h3 id="monitoring.coreos.com/v1.TSDBSpec">TSDBSpec
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
<code>outOfOrderTimeWindow</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Configures how old an out-of-order/out-of-bounds sample can be with
respect to the TSDB max time.</p>
<p>An out-of-order/out-of-bounds sample is ingested into the TSDB as long as
the timestamp of the sample is &gt;= (TSDB.MaxTime - outOfOrderTimeWindow).</p>
<p>Out of order ingestion is an experimental feature.</p>
<p>It requires Prometheus &gt;= v2.39.0.</p>
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
<code>version</code><br/>
<em>
string
</em>
</td>
<td>
<p>Version of Thanos to be deployed.</p>
</td>
</tr>
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
<code>imagePullPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#pullpolicy-v1-core">
Kubernetes core/v1.PullPolicy
</a>
</em>
</td>
<td>
<p>Image pull policy for the &lsquo;thanos&rsquo;, &lsquo;init-config-reloader&rsquo; and &lsquo;config-reloader&rsquo; containers.
See <a href="https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy">https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy</a> for more details.</p>
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
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<p>VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
VolumeMounts specified will be appended to other VolumeMounts in the ruler container,
that are generated as a result of StorageSpec objects.</p>
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
Defaults to <code>web</code>.</p>
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
This is an alpha field from kubernetes 1.22 until 1.24 which requires enabling the StatefulSetMinReadySeconds feature gate.</p>
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
<p>AdditionalArgs allows setting additional arguments for the ThanosRuler container.
It is intended for e.g. activating hidden flags which are not supported by
the dedicated configuration options yet. The arguments are passed as-is to the
ThanosRuler container which may cause issues if they are invalid or not supported
by the given ThanosRuler version.
In case of an argument conflict (e.g. an argument which is already set by the
operator itself) or when providing an invalid argument the reconciliation will
fail and an error will be logged.</p>
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
<p>ThanosRulerStatus is the most recent observed status of the ThanosRuler. Read-only.
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
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Condition">
[]Condition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The current state of the Alertmanager object.</p>
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
<p>ThanosSpec defines the configuration of the Thanos sidecar.</p>
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
<em>(Optional)</em>
<p>Container image name for Thanos. If specified, it takes precedence over
the <code>spec.thanos.baseImage</code>, <code>spec.thanos.tag</code> and <code>spec.thanos.sha</code>
fields.</p>
<p>Specifying <code>spec.thanos.version</code> is still necessary to ensure the
Prometheus Operator knows which version of Thanos is being configured.</p>
<p>If neither <code>spec.thanos.image</code> nor <code>spec.thanos.baseImage</code> are defined,
the operator will use the latest upstream version of Thanos available at
the time when the operator was released.</p>
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
<em>(Optional)</em>
<p>Version of Thanos being deployed. The operator uses this information
to generate the Prometheus StatefulSet + configuration files.</p>
<p>If not specified, the operator assumes the latest upstream release of
Thanos available at the time when the version of the operator was
released.</p>
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
<em>(Optional)</em>
<p><em>Deprecated: use &lsquo;image&rsquo; instead. The image&rsquo;s tag can be specified as
part of the image name.</em></p>
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
<em>(Optional)</em>
<p><em>Deprecated: use &lsquo;image&rsquo; instead.  The image digest can be specified
as part of the image name.</em></p>
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
<em>(Optional)</em>
<p><em>Deprecated: use &lsquo;image&rsquo; instead.</em></p>
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
<p>Defines the resources requests and limits of the Thanos sidecar.</p>
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
<em>(Optional)</em>
<p>Defines the Thanos sidecar&rsquo;s configuration to upload TSDB blocks to object storage.</p>
<p>More info: <a href="https://thanos.io/tip/thanos/storage.md/">https://thanos.io/tip/thanos/storage.md/</a></p>
<p>objectStorageConfigFile takes precedence over this field.</p>
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
<em>(Optional)</em>
<p>Defines the Thanos sidecar&rsquo;s configuration file to upload TSDB blocks to object storage.</p>
<p>More info: <a href="https://thanos.io/tip/thanos/storage.md/">https://thanos.io/tip/thanos/storage.md/</a></p>
<p>This field takes precedence over objectStorageConfig.</p>
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
<p><em>Deprecated: use <code>grpcListenLocal</code> and <code>httpListenLocal</code> instead.</em></p>
</td>
</tr>
<tr>
<td>
<code>grpcListenLocal</code><br/>
<em>
bool
</em>
</td>
<td>
<p>When true, the Thanos sidecar listens on the loopback interface instead
of the Pod IP&rsquo;s address for the gRPC endpoints.</p>
<p>It has no effect if <code>listenLocal</code> is true.</p>
</td>
</tr>
<tr>
<td>
<code>httpListenLocal</code><br/>
<em>
bool
</em>
</td>
<td>
<p>When true, the Thanos sidecar listens on the loopback interface instead
of the Pod IP&rsquo;s address for the HTTP endpoints.</p>
<p>It has no effect if <code>listenLocal</code> is true.</p>
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
<em>(Optional)</em>
<p>Defines the tracing configuration for the Thanos sidecar.</p>
<p>More info: <a href="https://thanos.io/tip/thanos/tracing.md/">https://thanos.io/tip/thanos/tracing.md/</a></p>
<p>This is an experimental feature, it may change in any upcoming release
in a breaking way.</p>
<p>tracingConfigFile takes precedence over this field.</p>
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
<p>Defines the tracing configuration file for the Thanos sidecar.</p>
<p>More info: <a href="https://thanos.io/tip/thanos/tracing.md/">https://thanos.io/tip/thanos/tracing.md/</a></p>
<p>This is an experimental feature, it may change in any upcoming release
in a breaking way.</p>
<p>This field takes precedence over tracingConfig.</p>
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
<em>(Optional)</em>
<p>Configures the TLS parameters for the gRPC server providing the StoreAPI.</p>
<p>Note: Currently only the <code>caFile</code>, <code>certFile</code>, and <code>keyFile</code> fields are supported.</p>
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
<p>Log level for the Thanos sidecar.</p>
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
<p>Log format for the Thanos sidecar.</p>
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
<p>Defines the start of time range limit served by the Thanos sidecar&rsquo;s StoreAPI.
The field&rsquo;s value should be a constant time in RFC3339 format or a time
duration relative to current time, such as -1d or 2h45m. Valid duration
units are ms, s, m, h, d, w, y.</p>
</td>
</tr>
<tr>
<td>
<code>blockSize</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>BlockDuration controls the size of TSDB blocks produced by Prometheus.
The default value is 2h to match the upstream Prometheus defaults.</p>
<p>WARNING: Changing the block duration can impact the performance and
efficiency of the entire Prometheus/Thanos stack due to how it interacts
with memory and Thanos compactors. It is recommended to keep this value
set to a multiple of 120 times your longest scrape or rule interval. For
example, 30s * 120 = 1h.</p>
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
<p>ReadyTimeout is the maximum time that the Thanos sidecar will wait for
Prometheus to start.</p>
</td>
</tr>
<tr>
<td>
<code>getConfigInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>How often to retrieve the Prometheus configuration.</p>
</td>
</tr>
<tr>
<td>
<code>getConfigTimeout</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Maximum time to wait when retrieving the Prometheus configuration.</p>
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
<em>(Optional)</em>
<p>VolumeMounts allows configuration of additional VolumeMounts for Thanos.
VolumeMounts specified will be appended to other VolumeMounts in the
&lsquo;thanos-sidecar&rsquo; container.</p>
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
<em>(Optional)</em>
<p>AdditionalArgs allows setting additional arguments for the Thanos container.
The arguments are passed as-is to the Thanos container which may cause issues
if they are invalid or not supported the given Thanos version.
In case of an argument conflict (e.g. an argument which is already set by the
operator itself) or when providing an invalid argument, the reconciliation will
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
</li><li>
<a href="#monitoring.coreos.com/v1alpha1.PrometheusAgent">PrometheusAgent</a>
</li><li>
<a href="#monitoring.coreos.com/v1alpha1.ScrapeConfig">ScrapeConfig</a>
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
<p>PrometheusAgent defines a Prometheus agent deployment.</p>
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
<code>podMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
EmbeddedObjectMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata configures labels and annotations which are propagated to the Prometheus pods.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for ServicedMonitors discovery. An empty label selector
matches all namespaces. A null label selector matches the current
namespace only.</p>
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
<p><em>Experimental</em> PodMonitors to be selected for target discovery. An empty
label selector matches all objects. A null label selector matches no
objects.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for PodMonitors discovery. An empty label selector
matches all namespaces. A null label selector matches the current
namespace only.</p>
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
<p><em>Experimental</em> Probes to be selected for target discovery. An empty
label selector matches all objects. A null label selector matches no
objects.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> Namespaces to match for Probe discovery. An empty label
selector matches all namespaces. A null label selector matches the
current namespace only.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeConfigSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> ScrapeConfigs to be selected for target discovery. An
empty label selector matches all objects. A null label selector matches
no objects.</p>
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
<code>scrapeConfigNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for ScrapeConfig discovery. An empty label selector
matches all namespaces. A null label selector matches the current
current namespace only.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#pullpolicy-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
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
<em>(Optional)</em>
<p>EXPERIMENTAL: Number of shards to distribute targets onto. <code>spec.replicas</code>
multiplied by <code>spec.shards</code> is the total number of Pods created.</p>
<p>Note that scaling down shards will not reshard data onto remaining
instances, it must be manually moved. Increasing shards will not reshard
data either but it will continue to be available from the same
instances. To query globally, use Thanos sidecar and Thanos querier or
remote write data to a central location.</p>
<p>Sharding is performed on the content of the <code>__address__</code> target meta-label
for PodMonitors and ServiceMonitors and <code>__param_target__</code> for Probes.</p>
<p>Default: 1</p>
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Number of seconds to wait until a scrape request times out.</p>
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
<code>enableFeatures</code><br/>
<em>
[]string
</em>
</td>
<td>
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
<a href="#monitoring.coreos.com/v1.StorageSpec">
StorageSpec
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
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
<code>web</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusWebSpec">
PrometheusWebSpec
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
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
<a href="#monitoring.coreos.com/v1.RemoteWriteSpec">
[]RemoteWriteSpec
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
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
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
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<a href="#monitoring.coreos.com/v1.APIServerConfig">
APIServerConfig
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
<a href="#monitoring.coreos.com/v1.ArbitraryFSAccessThroughSMsConfig">
ArbitraryFSAccessThroughSMsConfig
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
<p>When true, <code>spec.namespaceSelector</code> from all PodMonitor, ServiceMonitor
and Probe objects will be ignored. They will only discover targets
within the namespace of the PodMonitor, ServiceMonitor and Probe
objec.</p>
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
<p>When not empty, a label will be added to</p>
<ol>
<li>All metrics scraped from <code>ServiceMonitor</code>, <code>PodMonitor</code>, <code>Probe</code> and <code>ScrapeConfig</code> objects.</li>
<li>All metrics generated from recording rules defined in <code>PrometheusRule</code> objects.</li>
<li>All alerts generated from alerting rules defined in <code>PrometheusRule</code> objects.</li>
<li>All vector selectors of PromQL expressions defined in <code>PrometheusRule</code> objects.</li>
</ol>
<p>The label will not added for objects referenced in <code>spec.excludedFromEnforcement</code>.</p>
<p>The label&rsquo;s name is this field&rsquo;s value.
The label&rsquo;s value is the namespace of the <code>ServiceMonitor</code>,
<code>PodMonitor</code>, <code>Probe</code> or <code>PrometheusRule</code> object.</p>
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
unless <code>spec.sampleLimit</code> is greater than zero and less than than
<code>spec.enforcedSampleLimit</code>.</p>
<p>It is meant to be used by admins to keep the overall number of
samples/series under a desired limit.</p>
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
<p>When defined, enforcedBodySizeLimit specifies a global limit on the size
of uncompressed response body that will be accepted by Prometheus.
Targets responding with a body larger than this many bytes will cause
the scrape to fail.</p>
<p>It requires Prometheus &gt;= v2.28.0.</p>
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
<a href="#monitoring.coreos.com/v1.HostAlias">
[]HostAlias
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
<a href="#monitoring.coreos.com/v1.Argument">
[]Argument
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
<a href="#monitoring.coreos.com/v1.ObjectReference">
[]ObjectReference
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
it (<a href="https://kubernetes.io/docs/concepts/configuration/overview/">https://kubernetes.io/docs/concepts/configuration/overview/</a>).</p>
<p>When hostNetwork is enabled, this will set the DNS policy to
<code>ClusterFirstWithHostNet</code> automatically.</p>
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
<a href="#monitoring.coreos.com/v1.PrometheusTracingConfig">
PrometheusTracingConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>EXPERIMENTAL: TracingConfig configures tracing in Prometheus. This is an
experimental feature, it may change in any upcoming release in a
breaking way.</p>
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
<code>relabelings</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RelabelConfig">
[]RelabelConfig
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
<code>basicAuth</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
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
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
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
<code>tlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS configuration to use on every scrape request</p>
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>BasicAuth information to authenticate against the target HTTP endpoint.
More info: <a href="https://prometheus.io/docs/operating/configuration/#endpoints">https://prometheus.io/docs/operating/configuration/#endpoints</a></p>
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
<em>(Optional)</em>
<p>Authorization header configuration to authenticate against the target HTTP endpoint.</p>
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
<p>TLS configuration applying to the target HTTP endpoint.</p>
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
<code>podMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
EmbeddedObjectMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata configures labels and annotations which are propagated to the Prometheus pods.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for ServicedMonitors discovery. An empty label selector
matches all namespaces. A null label selector matches the current
namespace only.</p>
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
<p><em>Experimental</em> PodMonitors to be selected for target discovery. An empty
label selector matches all objects. A null label selector matches no
objects.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for PodMonitors discovery. An empty label selector
matches all namespaces. A null label selector matches the current
namespace only.</p>
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
<p><em>Experimental</em> Probes to be selected for target discovery. An empty
label selector matches all objects. A null label selector matches no
objects.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> Namespaces to match for Probe discovery. An empty label
selector matches all namespaces. A null label selector matches the
current namespace only.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeConfigSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p><em>Experimental</em> ScrapeConfigs to be selected for target discovery. An
empty label selector matches all objects. A null label selector matches
no objects.</p>
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
<code>scrapeConfigNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Namespaces to match for ScrapeConfig discovery. An empty label selector
matches all namespaces. A null label selector matches the current
current namespace only.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#pullpolicy-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
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
<em>(Optional)</em>
<p>EXPERIMENTAL: Number of shards to distribute targets onto. <code>spec.replicas</code>
multiplied by <code>spec.shards</code> is the total number of Pods created.</p>
<p>Note that scaling down shards will not reshard data onto remaining
instances, it must be manually moved. Increasing shards will not reshard
data either but it will continue to be available from the same
instances. To query globally, use Thanos sidecar and Thanos querier or
remote write data to a central location.</p>
<p>Sharding is performed on the content of the <code>__address__</code> target meta-label
for PodMonitors and ServiceMonitors and <code>__param_target__</code> for Probes.</p>
<p>Default: 1</p>
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<p>Number of seconds to wait until a scrape request times out.</p>
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
<code>enableFeatures</code><br/>
<em>
[]string
</em>
</td>
<td>
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
<a href="#monitoring.coreos.com/v1.StorageSpec">
StorageSpec
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
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
<code>web</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PrometheusWebSpec">
PrometheusWebSpec
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
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
<a href="#monitoring.coreos.com/v1.RemoteWriteSpec">
[]RemoteWriteSpec
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
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
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
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
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
<a href="#monitoring.coreos.com/v1.APIServerConfig">
APIServerConfig
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
<a href="#monitoring.coreos.com/v1.ArbitraryFSAccessThroughSMsConfig">
ArbitraryFSAccessThroughSMsConfig
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
<p>When true, <code>spec.namespaceSelector</code> from all PodMonitor, ServiceMonitor
and Probe objects will be ignored. They will only discover targets
within the namespace of the PodMonitor, ServiceMonitor and Probe
objec.</p>
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
<p>When not empty, a label will be added to</p>
<ol>
<li>All metrics scraped from <code>ServiceMonitor</code>, <code>PodMonitor</code>, <code>Probe</code> and <code>ScrapeConfig</code> objects.</li>
<li>All metrics generated from recording rules defined in <code>PrometheusRule</code> objects.</li>
<li>All alerts generated from alerting rules defined in <code>PrometheusRule</code> objects.</li>
<li>All vector selectors of PromQL expressions defined in <code>PrometheusRule</code> objects.</li>
</ol>
<p>The label will not added for objects referenced in <code>spec.excludedFromEnforcement</code>.</p>
<p>The label&rsquo;s name is this field&rsquo;s value.
The label&rsquo;s value is the namespace of the <code>ServiceMonitor</code>,
<code>PodMonitor</code>, <code>Probe</code> or <code>PrometheusRule</code> object.</p>
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
unless <code>spec.sampleLimit</code> is greater than zero and less than than
<code>spec.enforcedSampleLimit</code>.</p>
<p>It is meant to be used by admins to keep the overall number of
samples/series under a desired limit.</p>
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
<p>When defined, enforcedBodySizeLimit specifies a global limit on the size
of uncompressed response body that will be accepted by Prometheus.
Targets responding with a body larger than this many bytes will cause
the scrape to fail.</p>
<p>It requires Prometheus &gt;= v2.28.0.</p>
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
<a href="#monitoring.coreos.com/v1.HostAlias">
[]HostAlias
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
<a href="#monitoring.coreos.com/v1.Argument">
[]Argument
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
<a href="#monitoring.coreos.com/v1.ObjectReference">
[]ObjectReference
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
it (<a href="https://kubernetes.io/docs/concepts/configuration/overview/">https://kubernetes.io/docs/concepts/configuration/overview/</a>).</p>
<p>When hostNetwork is enabled, this will set the DNS policy to
<code>ClusterFirstWithHostNet</code> automatically.</p>
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
<a href="#monitoring.coreos.com/v1.PrometheusTracingConfig">
PrometheusTracingConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>EXPERIMENTAL: TracingConfig configures tracing in Prometheus. This is an
experimental feature, it may change in any upcoming release in a
breaking way.</p>
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
<p>The secret&rsquo;s key that contains the recipient user&rsquo;s user key.
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
<p>The secret&rsquo;s key that contains the registered application&rsquo;s API token, see <a href="https://pushover.net/apps">https://pushover.net/apps</a>.
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
<code>relabelings</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RelabelConfig">
[]RelabelConfig
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
<code>basicAuth</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
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
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
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
<code>tlsConfig</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
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
<em>(Optional)</em>
<p>List of targets for this static configuration.</p>
</td>
</tr>
<tr>
<td>
<code>labels</code><br/>
<em>
map[github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.LabelName]string
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
<p>Target represents a target for Prometheus to scrape</p>
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
<p>The Alertmanager route definition for alerts matching the resource&rsquo;s
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
the resource&rsquo;s namespace.</p>
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
<p>The Alertmanager route definition for alerts matching the resource&rsquo;s
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
the resource&rsquo;s namespace.</p>
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
operator enforces that the alert matches the resource&rsquo;s namespace.</p>
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
<p>The secret&rsquo;s key that contains the recipient user&rsquo;s user key.
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
<p>The secret&rsquo;s key that contains the registered application&rsquo;s API token, see <a href="https://pushover.net/apps">https://pushover.net/apps</a>.
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
MuteTimeIntervals is a list of TimeInterval names that will mute this route when matched.</p>
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
<p>ActiveTimeIntervals is a list of TimeInterval names when this route should be active.</p>
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
