---
title: "API reference"
description: "Prometheus operator generated API reference docs"
draft: false
images: []
menu: "operator"
weight: 151
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
<p>The <code>Alertmanager</code> custom resource definition (CRD) defines a desired <a href="https://prometheus.io/docs/alerting">Alertmanager</a> setup to run in a Kubernetes cluster. It allows to specify many options such as the number of replicas, persistent storage and many more.</p>
<p>For each <code>Alertmanager</code> resource, the Operator deploys a <code>StatefulSet</code> in the same namespace. When there are two or more configured replicas, the Operator runs the Alertmanager instances in high-availability mode.</p>
<p>The resource defines via label and namespace selectors which <code>AlertmanagerConfig</code> objects should be associated to the deployed Alertmanager instances.</p>
</div>
<table>
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
<p>PodMetadata configures labels and annotations which are propagated to the Alertmanager pods.</p>
<p>The following items are reserved and cannot be overridden:
* &ldquo;alertmanager&rdquo; label, set to the name of the Alertmanager instance.
* &ldquo;app.kubernetes.io/instance&rdquo; label, set to the name of the Alertmanager instance.
* &ldquo;app.kubernetes.io/managed-by&rdquo; label, set to &ldquo;prometheus-operator&rdquo;.
* &ldquo;app.kubernetes.io/name&rdquo; label, set to &ldquo;alertmanager&rdquo;.
* &ldquo;app.kubernetes.io/version&rdquo; label, set to the Alertmanager version.
* &ldquo;kubectl.kubernetes.io/default-container&rdquo; annotation, set to &ldquo;alertmanager&rdquo;.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#pullpolicy-v1-core">
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
Deprecated: use &lsquo;image&rsquo; instead. The image tag can be specified as part of the image URL.</p>
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
Deprecated: use &lsquo;image&rsquo; instead. The image digest can be specified as part of the image URL.</p>
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
Deprecated: use &lsquo;image&rsquo; instead.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volume-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volumemount-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#resourcerequirements-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#affinity-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#toleration-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#topologyspreadconstraint-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#podsecuritycontext-v1-core">
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
<code>dnsPolicy</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.DNSPolicy">
DNSPolicy
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
<a href="#monitoring.coreos.com/v1.PodDNSConfig">
PodDNSConfig
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
<code>serviceName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The name of the service name used by the underlying StatefulSet(s) as the governing service.
If defined, the Service  must be created before the Alertmanager resource in the same namespace and it must define a selector that matches the pod labels.
If empty, the operator will create and manage a headless service named <code>alertmanager-operated</code> for Alermanager resources.
When deploying multiple Alertmanager resources in the same namespace, it is recommended to specify a different value for each.
See <a href="https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id">https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id</a> for more details.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core">
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
<code>clusterLabel</code><br/>
<em>
string
</em>
</td>
<td>
<p>Defines the identifier that uniquely identifies the Alertmanager cluster.
You should only set it when the Alertmanager cluster includes Alertmanager instances which are external to this Alertmanager resource. In practice, the addresses of the external instances are provided via the <code>.spec.additionalPeers</code> field.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
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
<code>alertmanagerConfigMatcherStrategy</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertmanagerConfigMatcherStrategy">
AlertmanagerConfigMatcherStrategy
</a>
</em>
</td>
<td>
<p>AlertmanagerConfigMatcherStrategy defines how AlertmanagerConfig objects
process incoming alerts.</p>
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
<code>limits</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertmanagerLimitsSpec">
AlertmanagerLimitsSpec
</a>
</em>
</td>
<td>
<p>Defines the limits command line flags when starting Alertmanager.</p>
</td>
</tr>
<tr>
<td>
<code>clusterTLS</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ClusterTLSConfig">
ClusterTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures the mutual TLS configuration for the Alertmanager cluster&rsquo;s gossip protocol.</p>
<p>It requires Alertmanager &gt;= 0.24.0.</p>
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
<em>(Optional)</em>
<p>alertmanagerConfiguration specifies the configuration of Alertmanager.</p>
<p>If defined, it takes precedence over the <code>configSecret</code> field.</p>
<p>This is an <em>experimental feature</em>, it may change in any upcoming release
in a breaking way.</p>
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
<tr>
<td>
<code>enableFeatures</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enable access to Alertmanager feature flags. By default, no features are enabled.
Enabling features which are disabled by default is entirely outside the
scope of what the maintainers will support and by doing so, you accept
that this behaviour may break at any time without notice.</p>
<p>It requires Alertmanager &gt;= 0.27.0.</p>
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
<p>AdditionalArgs allows setting additional arguments for the &lsquo;Alertmanager&rsquo; container.
It is intended for e.g. activating hidden flags which are not supported by
the dedicated configuration options yet. The arguments are passed as-is to the
Alertmanager container which may cause issues if they are invalid or not supported
by the given Alertmanager version.</p>
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
<p>Defaults to 120 seconds.</p>
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
<p>The <code>PodMonitor</code> custom resource definition (CRD) defines how <code>Prometheus</code> and <code>PrometheusAgent</code> can scrape metrics from a group of pods.
Among other things, it allows to specify:
* The pods to scrape via label selectors.
* The container ports to scrape.
* Authentication credentials to use.
* Target and metric relabeling.</p>
<p><code>Prometheus</code> and <code>PrometheusAgent</code> objects select <code>PodMonitor</code> objects using label and namespace selectors.</p>
</div>
<table>
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
<p>The label to use to retrieve the job name from.
<code>jobLabel</code> selects the label from the associated Kubernetes <code>Pod</code>
object which will be used as the <code>job</code> label for all metrics.</p>
<p>For example if <code>jobLabel</code> is set to <code>foo</code> and the Kubernetes <code>Pod</code>
object is labeled with <code>foo: bar</code>, then Prometheus adds the <code>job=&quot;bar&quot;</code>
label to all ingested metrics.</p>
<p>If the value of this field is empty, the <code>job</code> label of the metrics
defaults to the namespace and name of the PodMonitor object (e.g. <code>&lt;namespace&gt;/&lt;name&gt;</code>).</p>
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
<p><code>podTargetLabels</code> defines the labels which are transferred from the
associated Kubernetes <code>Pod</code> object onto the ingested metrics.</p>
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
<em>(Optional)</em>
<p>Defines how to scrape metrics from the selected pods.</p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Label selector to select the Kubernetes <code>Pod</code> objects to scrape metrics from.</p>
</td>
</tr>
<tr>
<td>
<code>selectorMechanism</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SelectorMechanism">
SelectorMechanism
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Mechanism used to select the endpoints to scrape.
By default, the selection process relies on relabel configurations to filter the discovered targets.
Alternatively, you can opt in for role selectors, which may offer better efficiency in large clusters.
Which strategy is best for your use case needs to be carefully evaluated.</p>
<p>It requires Prometheus &gt;= v2.17.0.</p>
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
<p><code>namespaceSelector</code> defines in which namespace(s) Prometheus should discover the pods.
By default, the pods are discovered in the same namespace as the <code>PodMonitor</code> object but it is possible to select pods across different/all namespaces.</p>
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
<p><code>sampleLimit</code> defines a per-scrape limit on the number of scraped samples
that will be accepted.</p>
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
<p><code>targetLimit</code> defines a limit on the number of scraped targets that will
be accepted.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeProtocols</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
[]ScrapeProtocol
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>scrapeProtocols</code> defines the protocols to negotiate during a scrape. It tells clients the
protocols supported by Prometheus in order of preference (from most to least preferred).</p>
<p>If unset, Prometheus uses its default value.</p>
<p>It requires Prometheus &gt;= v2.49.0.</p>
</td>
</tr>
<tr>
<td>
<code>fallbackScrapeProtocol</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
ScrapeProtocol
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
<code>labelLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on number of labels that will be accepted for a sample.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
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
<p>Per-scrape limit on length of labels name that will be accepted for a sample.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
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
<p>Per-scrape limit on length of labels value that will be accepted for a sample.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
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
<code>attachMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AttachMetadata">
AttachMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>attachMetadata</code> defines additional metadata which is added to the
discovered targets.</p>
<p>It requires Prometheus &gt;= v2.35.0.</p>
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
<tr>
<td>
<code>bodySizeLimit</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>When defined, bodySizeLimit specifies a job level limit on the size
of uncompressed response body that will be accepted by Prometheus.</p>
<p>It requires Prometheus &gt;= v2.28.0.</p>
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
<p>The <code>Probe</code> custom resource definition (CRD) defines how to scrape metrics from prober exporters such as the <a href="https://github.com/prometheus/blackbox_exporter">blackbox exporter</a>.</p>
<p>The <code>Probe</code> resource needs 2 pieces of information:
* The list of probed addresses which can be defined statically or by discovering Kubernetes Ingress objects.
* The prober which exposes the availability of probed endpoints (over various protocols such HTTP, TCP, ICMP, &hellip;) as Prometheus metrics.</p>
<p><code>Prometheus</code> and <code>PrometheusAgent</code> objects select <code>Probe</code> objects using label and namespace selectors.</p>
</div>
<table>
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
If not specified, the Prometheus global scrape timeout is used.
The value cannot be greater than the scrape interval otherwise the operator will reject the resource.</p>
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
<p>TLS configuration to use when scraping the endpoint.</p>
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
<code>scrapeProtocols</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
[]ScrapeProtocol
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>scrapeProtocols</code> defines the protocols to negotiate during a scrape. It tells clients the
protocols supported by Prometheus in order of preference (from most to least preferred).</p>
<p>If unset, Prometheus uses its default value.</p>
<p>It requires Prometheus &gt;= v2.49.0.</p>
</td>
</tr>
<tr>
<td>
<code>fallbackScrapeProtocol</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
ScrapeProtocol
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
<h3 id="monitoring.coreos.com/v1.Prometheus">Prometheus
</h3>
<div>
<p>The <code>Prometheus</code> custom resource definition (CRD) defines a desired <a href="https://prometheus.io/docs/prometheus">Prometheus</a> setup to run in a Kubernetes cluster. It allows to specify many options such as the number of replicas, persistent storage, and Alertmanagers where firing alerts should be sent and many more.</p>
<p>For each <code>Prometheus</code> resource, the Operator deploys one or several <code>StatefulSet</code> objects in the same namespace. The number of StatefulSets is equal to the number of shards which is 1 by default.</p>
<p>The resource defines via label and namespace selectors which <code>ServiceMonitor</code>, <code>PodMonitor</code>, <code>Probe</code> and <code>PrometheusRule</code> objects should be associated to the deployed Prometheus instances.</p>
<p>The Operator continuously reconciles the scrape and rules configuration and a sidecar container running in the Prometheus pods triggers a reload of the configuration when needed.</p>
</div>
<table>
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
<p>Number of seconds to wait until a scrape request times out.
The value cannot be greater than the scrape interval otherwise the operator will reject the resource.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeProtocols</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
[]ScrapeProtocol
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
<a href="#monitoring.coreos.com/v1.RemoteWriteMessageVersion">
[]RemoteWriteMessageVersion
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
<a href="#monitoring.coreos.com/v1.EnableFeature">
[]EnableFeature
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
<a href="#monitoring.coreos.com/v1.TopologySpreadConstraint">
[]TopologySpreadConstraint
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
<code>otlp</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OTLPConfig">
OTLPConfig
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
<a href="#monitoring.coreos.com/v1.DNSPolicy">
DNSPolicy
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
<a href="#monitoring.coreos.com/v1.PodDNSConfig">
PodDNSConfig
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
<a href="#monitoring.coreos.com/v1.NameValidationSchemeOptions">
NameValidationSchemeOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the validation scheme for metric and label names.</p>
</td>
</tr>
<tr>
<td>
<code>nameEscapingScheme</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.NameEscapingSchemeOptions">
NameEscapingSchemeOptions
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
<p>Whether to convert all scraped classic histograms into a native histogram with custom buckets.</p>
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
<a href="#monitoring.coreos.com/v1.PrometheusTracingConfig">
PrometheusTracingConfig
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
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
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
<a href="#monitoring.coreos.com/v1.ReloadStrategyType">
ReloadStrategyType
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
<a href="#monitoring.coreos.com/v1.ScrapeClass">
[]ScrapeClass
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
<a href="#monitoring.coreos.com/v1.ServiceDiscoveryRole">
ServiceDiscoveryRole
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
<a href="#monitoring.coreos.com/v1.TSDBSpec">
TSDBSpec
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
<a href="#monitoring.coreos.com/v1.RuntimeConfig">
RuntimeConfig
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
<tr>
<td>
<code>baseImage</code><br/>
<em>
string
</em>
</td>
<td>
<p>Deprecated: use &lsquo;spec.image&rsquo; instead.</p>
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
<p>Deprecated: use &lsquo;spec.image&rsquo; instead. The image&rsquo;s tag can be specified as part of the image name.</p>
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
<p>Deprecated: use &lsquo;spec.image&rsquo; instead. The image&rsquo;s digest can be specified as part of the image name.</p>
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
<code>shardRetentionPolicy</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ShardRetentionPolicy">
ShardRetentionPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ShardRetentionPolicy defines the retention policy for the Prometheus shards.
(Alpha) Using this field requires the &lsquo;PrometheusShardRetentionPolicy&rsquo; feature gate to be enabled.</p>
<p>The final goals for this feature can be seen at <a href="https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/proposals/202310-shard-autoscaling.md#graceful-scale-down-of-prometheus-servers">https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/proposals/202310-shard-autoscaling.md#graceful-scale-down-of-prometheus-servers</a>,
however, the feature is not yet fully implemented in this PR. The limitation being:
* Retention duration is not settable, for now, shards are retained forever.</p>
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
<p>When true, the Prometheus compaction is disabled.
When <code>spec.thanos.objectStorageConfig</code> or <code>spec.objectStorageConfigFile</code> are defined, the operator automatically
disables block compaction to avoid race conditions during block uploads (as the Thanos documentation recommends).</p>
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
Deprecated: use <code>spec.excludedFromEnforcement</code> instead.</p>
</td>
</tr>
<tr>
<td>
<code>ruleSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<p>Deprecated: this flag has no effect for Prometheus &gt;= 2.39.0 where overlapping blocks are enabled by default.</p>
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
<code>ruleQueryOffset</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the offset the rule evaluation timestamp of this particular group by the specified duration into the past.
It requires Prometheus &gt;= v2.53.0.</p>
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
<p>The <code>PrometheusRule</code> custom resource definition (CRD) defines <a href="https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/">alerting</a> and <a href="https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/">recording</a> rules to be evaluated by <code>Prometheus</code> or <code>ThanosRuler</code> objects.</p>
<p><code>Prometheus</code> and <code>ThanosRuler</code> objects select <code>PrometheusRule</code> objects using label and namespace selectors.</p>
</div>
<table>
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
<p>The <code>ServiceMonitor</code> custom resource definition (CRD) defines how <code>Prometheus</code> and <code>PrometheusAgent</code> can scrape metrics from a group of services.
Among other things, it allows to specify:
* The services to scrape via label selectors.
* The container ports to scrape.
* Authentication credentials to use.
* Target and metric relabeling.</p>
<p><code>Prometheus</code> and <code>PrometheusAgent</code> objects select <code>ServiceMonitor</code> objects using label and namespace selectors.</p>
</div>
<table>
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
<p><code>jobLabel</code> selects the label from the associated Kubernetes <code>Service</code>
object which will be used as the <code>job</code> label for all metrics.</p>
<p>For example if <code>jobLabel</code> is set to <code>foo</code> and the Kubernetes <code>Service</code>
object is labeled with <code>foo: bar</code>, then Prometheus adds the <code>job=&quot;bar&quot;</code>
label to all ingested metrics.</p>
<p>If the value of this field is empty or if the label doesn&rsquo;t exist for
the given Service, the <code>job</code> label of the metrics defaults to the name
of the associated Kubernetes <code>Service</code>.</p>
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
<em>(Optional)</em>
<p><code>targetLabels</code> defines the labels which are transferred from the
associated Kubernetes <code>Service</code> object onto the ingested metrics.</p>
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
<p><code>podTargetLabels</code> defines the labels which are transferred from the
associated Kubernetes <code>Pod</code> object onto the ingested metrics.</p>
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
<p>List of endpoints part of this ServiceMonitor.
Defines how to scrape metrics from Kubernetes <a href="https://kubernetes.io/docs/concepts/services-networking/service/#endpoints">Endpoints</a> objects.
In most cases, an Endpoints object is backed by a Kubernetes <a href="https://kubernetes.io/docs/concepts/services-networking/service/">Service</a> object with the same name and labels.</p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Label selector to select the Kubernetes <code>Endpoints</code> objects to scrape metrics from.</p>
</td>
</tr>
<tr>
<td>
<code>selectorMechanism</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SelectorMechanism">
SelectorMechanism
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Mechanism used to select the endpoints to scrape.
By default, the selection process relies on relabel configurations to filter the discovered targets.
Alternatively, you can opt in for role selectors, which may offer better efficiency in large clusters.
Which strategy is best for your use case needs to be carefully evaluated.</p>
<p>It requires Prometheus &gt;= v2.17.0.</p>
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
<p><code>namespaceSelector</code> defines in which namespace(s) Prometheus should discover the services.
By default, the services are discovered in the same namespace as the <code>ServiceMonitor</code> object but it is possible to select pods across different/all namespaces.</p>
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
<p><code>sampleLimit</code> defines a per-scrape limit on the number of scraped samples
that will be accepted.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeProtocols</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
[]ScrapeProtocol
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>scrapeProtocols</code> defines the protocols to negotiate during a scrape. It tells clients the
protocols supported by Prometheus in order of preference (from most to least preferred).</p>
<p>If unset, Prometheus uses its default value.</p>
<p>It requires Prometheus &gt;= v2.49.0.</p>
</td>
</tr>
<tr>
<td>
<code>fallbackScrapeProtocol</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
ScrapeProtocol
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
<code>targetLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>targetLimit</code> defines a limit on the number of scraped targets that will
be accepted.</p>
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
<p>Per-scrape limit on number of labels that will be accepted for a sample.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
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
<p>Per-scrape limit on length of labels name that will be accepted for a sample.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
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
<p>Per-scrape limit on length of labels value that will be accepted for a sample.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
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
<code>attachMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AttachMetadata">
AttachMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>attachMetadata</code> defines additional metadata which is added to the
discovered targets.</p>
<p>It requires Prometheus &gt;= v2.37.0.</p>
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
<tr>
<td>
<code>bodySizeLimit</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>When defined, bodySizeLimit specifies a job level limit on the size
of uncompressed response body that will be accepted by Prometheus.</p>
<p>It requires Prometheus &gt;= v2.28.0.</p>
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
<p>The <code>ThanosRuler</code> custom resource definition (CRD) defines a desired <a href="https://github.com/thanos-io/thanos/blob/main/docs/components/rule.md">Thanos Ruler</a> setup to run in a Kubernetes cluster.</p>
<p>A <code>ThanosRuler</code> instance requires at least one compatible Prometheus API endpoint (either Thanos Querier or Prometheus services).</p>
<p>The resource defines via label and namespace selectors which <code>PrometheusRule</code> objects should be associated to the deployed Thanos Ruler instances.</p>
</div>
<table>
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
<em>(Optional)</em>
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
<em>(Optional)</em>
<p>PodMetadata configures labels and annotations which are propagated to the ThanosRuler pods.</p>
<p>The following items are reserved and cannot be overridden:
* &ldquo;app.kubernetes.io/name&rdquo; label, set to &ldquo;thanos-ruler&rdquo;.
* &ldquo;app.kubernetes.io/managed-by&rdquo; label, set to &ldquo;prometheus-operator&rdquo;.
* &ldquo;app.kubernetes.io/instance&rdquo; label, set to the name of the ThanosRuler instance.
* &ldquo;thanos-ruler&rdquo; label, set to the name of the ThanosRuler instance.
* &ldquo;kubectl.kubernetes.io/default-container&rdquo; annotation, set to &ldquo;thanos-ruler&rdquo;.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#pullpolicy-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
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
<em>(Optional)</em>
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
<em>(Optional)</em>
<p>Define which Nodes the Pods are scheduled on.</p>
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
<p>Resources defines the resource requirements for single Pods.
If not provided, no requests/limits will be set</p>
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
<p>If specified, the pod&rsquo;s scheduling constraints.</p>
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
<p>If specified, the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s topology spread constraints.</p>
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
<a href="#monitoring.coreos.com/v1.DNSPolicy">
DNSPolicy
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
<a href="#monitoring.coreos.com/v1.PodDNSConfig">
PodDNSConfig
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
<code>serviceName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The name of the service name used by the underlying StatefulSet(s) as the governing service.
If defined, the Service  must be created before the ThanosRuler resource in the same namespace and it must define a selector that matches the pod labels.
If empty, the operator will create and manage a headless service named <code>thanos-ruler-operated</code> for ThanosRuler resources.
When deploying multiple ThanosRuler resources in the same namespace, it is recommended to specify a different value for each.
See <a href="https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id">https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id</a> for more details.</p>
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
<em>(Optional)</em>
<p>Storage spec to specify how storage shall be used.</p>
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
<em>(Optional)</em>
<p>Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will
be appended to other volumes that are generated as a result of StorageSpec objects.</p>
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
<em>(Optional)</em>
<p>VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
VolumeMounts specified will be appended to other VolumeMounts in the ruler container,
that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>objectStorageConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures object storage.</p>
<p>The configuration format is defined at <a href="https://thanos.io/tip/thanos/storage.md/#configuring-access-to-object-storage">https://thanos.io/tip/thanos/storage.md/#configuring-access-to-object-storage</a></p>
<p>The operator performs no validation of the configuration.</p>
<p><code>objectStorageConfigFile</code> takes precedence over this field.</p>
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
<p>Configures the path of the object storage configuration file.</p>
<p>The configuration format is defined at <a href="https://thanos.io/tip/thanos/storage.md/#configuring-access-to-object-storage">https://thanos.io/tip/thanos/storage.md/#configuring-access-to-object-storage</a></p>
<p>The operator performs no validation of the configuration file.</p>
<p>This field takes precedence over <code>objectStorageConfig</code>.</p>
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
<em>(Optional)</em>
<p>Configures the list of Thanos Query endpoints from which to query metrics.</p>
<p>For Thanos &gt;= v0.11.0, it is recommended to use <code>queryConfig</code> instead.</p>
<p><code>queryConfig</code> takes precedence over this field.</p>
</td>
</tr>
<tr>
<td>
<code>queryConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures the list of Thanos Query endpoints from which to query metrics.</p>
<p>The configuration format is defined at <a href="https://thanos.io/tip/components/rule.md/#query-api">https://thanos.io/tip/components/rule.md/#query-api</a></p>
<p>It requires Thanos &gt;= v0.11.0.</p>
<p>The operator performs no validation of the configuration.</p>
<p>This field takes precedence over <code>queryEndpoints</code>.</p>
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
<em>(Optional)</em>
<p>Configures the list of Alertmanager endpoints to send alerts to.</p>
<p>For Thanos &gt;= v0.10.0, it is recommended to use <code>alertmanagersConfig</code> instead.</p>
<p><code>alertmanagersConfig</code> takes precedence over this field.</p>
</td>
</tr>
<tr>
<td>
<code>alertmanagersConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures the list of Alertmanager endpoints to send alerts to.</p>
<p>The configuration format is defined at <a href="https://thanos.io/tip/components/rule.md/#alertmanager">https://thanos.io/tip/components/rule.md/#alertmanager</a>.</p>
<p>It requires Thanos &gt;= v0.10.0.</p>
<p>The operator performs no validation of the configuration.</p>
<p>This field takes precedence over <code>alertmanagersUrl</code>.</p>
</td>
</tr>
<tr>
<td>
<code>ruleSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
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
<em>(Optional)</em>
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
<em>(Optional)</em>
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
<p>Time duration ThanosRuler shall retain data for. Default is &lsquo;24h&rsquo;, and
must match the regular expression <code>[0-9]+(ms|s|m|h|d|w|y)</code> (milliseconds
seconds minutes hours days weeks years).</p>
<p>The field has no effect when remote-write is configured since the Ruler
operates in stateless mode.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<em>(Optional)</em>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures tracing.</p>
<p>The configuration format is defined at <a href="https://thanos.io/tip/thanos/tracing.md/#configuration">https://thanos.io/tip/thanos/tracing.md/#configuration</a></p>
<p>This is an <em>experimental feature</em>, it may change in any upcoming release
in a breaking way.</p>
<p>The operator performs no validation of the configuration.</p>
<p><code>tracingConfigFile</code> takes precedence over this field.</p>
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
<em>(Optional)</em>
<p>Configures the path of the tracing configuration file.</p>
<p>The configuration format is defined at <a href="https://thanos.io/tip/thanos/tracing.md/#configuration">https://thanos.io/tip/thanos/tracing.md/#configuration</a></p>
<p>This is an <em>experimental feature</em>, it may change in any upcoming release
in a breaking way.</p>
<p>The operator performs no validation of the configuration file.</p>
<p>This field takes precedence over <code>tracingConfig</code>.</p>
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
<p>Configures the external label pairs of the ThanosRuler resource.</p>
<p>A default replica label <code>thanos_ruler_replica</code> will be always added as a
label with the value of the pod&rsquo;s name.</p>
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
<em>(Optional)</em>
<p>Configures the label names which should be dropped in Thanos Ruler
alerts.</p>
<p>The replica label <code>thanos_ruler_replica</code> will always be dropped from the alerts.</p>
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
<em>(Optional)</em>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures alert relabeling in Thanos Ruler.</p>
<p>Alert relabel configuration must have the form as specified in the
official Prometheus documentation:
<a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs</a></p>
<p>The operator performs no validation of the configuration.</p>
<p><code>alertRelabelConfigFile</code> takes precedence over this field.</p>
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
<em>(Optional)</em>
<p>Configures the path to the alert relabeling configuration file.</p>
<p>Alert relabel configuration must have the form as specified in the
official Prometheus documentation:
<a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs</a></p>
<p>The operator performs no validation of the configuration file.</p>
<p>This field takes precedence over <code>alertRelabelConfig</code>.</p>
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
<em>(Optional)</em>
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
<tr>
<td>
<code>web</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ThanosRulerWebSpec">
ThanosRulerWebSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the configuration of the ThanosRuler web server.</p>
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
<p>When the list isn&rsquo;t empty, the ruler is configured with stateless mode.</p>
<p>It requires Thanos &gt;= 0.24.0.</p>
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
<p>Defaults to 120 seconds.</p>
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
<p>Deprecated: this will be removed in a future release. Prefer using <code>authorization</code>.</p>
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
<p>Deprecated: this will be removed in a future release.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AdditionalLabelSelectors">AdditionalLabelSelectors
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.TopologySpreadConstraint">TopologySpreadConstraint</a>)
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
<tbody><tr><td><p>&#34;OnResource&#34;</p></td>
<td><p>Automatically add a label selector that will select all pods matching the same Prometheus/PrometheusAgent resource (irrespective of their shards).</p>
</td>
</tr><tr><td><p>&#34;OnShard&#34;</p></td>
<td><p>Automatically add a label selector that will select all pods matching the same shard.</p>
</td>
</tr></tbody>
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
<p>Alertmanager endpoints where Prometheus should send alerts to.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AlertmanagerAPIVersion">AlertmanagerAPIVersion
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>)
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
<tbody><tr><td><p>&#34;V1&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;V2&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AlertmanagerConfigMatcherStrategy">AlertmanagerConfigMatcherStrategy
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerSpec">AlertmanagerSpec</a>)
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
<a href="#monitoring.coreos.com/v1.AlertmanagerConfigMatcherStrategyType">
AlertmanagerConfigMatcherStrategyType
</a>
</em>
</td>
<td>
<p>AlertmanagerConfigMatcherStrategyType defines the strategy used by
AlertmanagerConfig objects to match alerts in the routes and inhibition
rules.</p>
<p>The default value is <code>OnNamespace</code>.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AlertmanagerConfigMatcherStrategyType">AlertmanagerConfigMatcherStrategyType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerConfigMatcherStrategy">AlertmanagerConfigMatcherStrategy</a>)
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
<tbody><tr><td><p>&#34;None&#34;</p></td>
<td><p>With <code>None</code>, the route and inhbition rules of an AlertmanagerConfig
object process all incoming alerts.</p>
</td>
</tr><tr><td><p>&#34;OnNamespace&#34;</p></td>
<td><p>With <code>OnNamespace</code>, the route and inhibition rules of an
AlertmanagerConfig object only process alerts that have a <code>namespace</code>
label equal to the namespace of the object.</p>
</td>
</tr></tbody>
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
<em>(Optional)</em>
<p>Namespace of the Endpoints object.</p>
<p>If not set, the object will be discovered in the namespace of the
Prometheus object.</p>
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
<p>Cannot be set at the same time as <code>bearerTokenFile</code>, <code>authorization</code> or <code>sigv4</code>.</p>
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
<p>Cannot be set at the same time as <code>basicAuth</code>, <code>authorization</code>, or <code>sigv4</code>.</p>
<p>Deprecated: this will be removed in a future release. Prefer using <code>authorization</code>.</p>
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
<p>Cannot be set at the same time as <code>basicAuth</code>, <code>bearerTokenFile</code> or <code>sigv4</code>.</p>
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
<p>It requires Prometheus &gt;= v2.48.0.</p>
<p>Cannot be set at the same time as <code>basicAuth</code>, <code>bearerTokenFile</code> or <code>authorization</code>.</p>
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
<code>apiVersion</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertmanagerAPIVersion">
AlertmanagerAPIVersion
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Version of the Alertmanager API that Prometheus uses to send alerts.
It can be &ldquo;V1&rdquo; or &ldquo;V2&rdquo;.
The field has no effect for Prometheus &gt;= v3.0.0 because only the v2 API is supported.</p>
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
<p>Relabel configuration applied to the discovered Alertmanagers.</p>
</td>
</tr>
<tr>
<td>
<code>alertRelabelings</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RelabelConfig">
[]RelabelConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Relabeling configs applied before sending alerts to a specific Alertmanager.
It requires Prometheus &gt;= v2.51.0.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<h3 id="monitoring.coreos.com/v1.AlertmanagerLimitsSpec">AlertmanagerLimitsSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerSpec">AlertmanagerSpec</a>)
</p>
<div>
<p>AlertmanagerLimitsSpec defines the limits command line flags when starting Alertmanager.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>maxSilences</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The maximum number active and pending silences. This corresponds to the
Alertmanager&rsquo;s <code>--silences.max-silences</code> flag.
It requires Alertmanager &gt;= v0.28.0.</p>
</td>
</tr>
<tr>
<td>
<code>maxPerSilenceBytes</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The maximum size of an individual silence as stored on disk. This corresponds to the Alertmanager&rsquo;s
<code>--silences.max-per-silence-bytes</code> flag.
It requires Alertmanager &gt;= v0.28.0.</p>
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
<p>PodMetadata configures labels and annotations which are propagated to the Alertmanager pods.</p>
<p>The following items are reserved and cannot be overridden:
* &ldquo;alertmanager&rdquo; label, set to the name of the Alertmanager instance.
* &ldquo;app.kubernetes.io/instance&rdquo; label, set to the name of the Alertmanager instance.
* &ldquo;app.kubernetes.io/managed-by&rdquo; label, set to &ldquo;prometheus-operator&rdquo;.
* &ldquo;app.kubernetes.io/name&rdquo; label, set to &ldquo;alertmanager&rdquo;.
* &ldquo;app.kubernetes.io/version&rdquo; label, set to the Alertmanager version.
* &ldquo;kubectl.kubernetes.io/default-container&rdquo; annotation, set to &ldquo;alertmanager&rdquo;.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#pullpolicy-v1-core">
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
Deprecated: use &lsquo;image&rsquo; instead. The image tag can be specified as part of the image URL.</p>
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
Deprecated: use &lsquo;image&rsquo; instead. The image digest can be specified as part of the image URL.</p>
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
Deprecated: use &lsquo;image&rsquo; instead.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volume-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volumemount-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#resourcerequirements-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#affinity-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#toleration-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#topologyspreadconstraint-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#podsecuritycontext-v1-core">
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
<code>dnsPolicy</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.DNSPolicy">
DNSPolicy
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
<a href="#monitoring.coreos.com/v1.PodDNSConfig">
PodDNSConfig
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
<code>serviceName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The name of the service name used by the underlying StatefulSet(s) as the governing service.
If defined, the Service  must be created before the Alertmanager resource in the same namespace and it must define a selector that matches the pod labels.
If empty, the operator will create and manage a headless service named <code>alertmanager-operated</code> for Alermanager resources.
When deploying multiple Alertmanager resources in the same namespace, it is recommended to specify a different value for each.
See <a href="https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id">https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id</a> for more details.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core">
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
<code>clusterLabel</code><br/>
<em>
string
</em>
</td>
<td>
<p>Defines the identifier that uniquely identifies the Alertmanager cluster.
You should only set it when the Alertmanager cluster includes Alertmanager instances which are external to this Alertmanager resource. In practice, the addresses of the external instances are provided via the <code>.spec.additionalPeers</code> field.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
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
<code>alertmanagerConfigMatcherStrategy</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertmanagerConfigMatcherStrategy">
AlertmanagerConfigMatcherStrategy
</a>
</em>
</td>
<td>
<p>AlertmanagerConfigMatcherStrategy defines how AlertmanagerConfig objects
process incoming alerts.</p>
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
<code>limits</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AlertmanagerLimitsSpec">
AlertmanagerLimitsSpec
</a>
</em>
</td>
<td>
<p>Defines the limits command line flags when starting Alertmanager.</p>
</td>
</tr>
<tr>
<td>
<code>clusterTLS</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ClusterTLSConfig">
ClusterTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures the mutual TLS configuration for the Alertmanager cluster&rsquo;s gossip protocol.</p>
<p>It requires Alertmanager &gt;= 0.24.0.</p>
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
<em>(Optional)</em>
<p>alertmanagerConfiguration specifies the configuration of Alertmanager.</p>
<p>If defined, it takes precedence over the <code>configSecret</code> field.</p>
<p>This is an <em>experimental feature</em>, it may change in any upcoming release
in a breaking way.</p>
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
<tr>
<td>
<code>enableFeatures</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enable access to Alertmanager feature flags. By default, no features are enabled.
Enabling features which are disabled by default is entirely outside the
scope of what the maintainers will support and by doing so, you accept
that this behaviour may break at any time without notice.</p>
<p>It requires Alertmanager &gt;= 0.27.0.</p>
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
<p>AdditionalArgs allows setting additional arguments for the &lsquo;Alertmanager&rsquo; container.
It is intended for e.g. activating hidden flags which are not supported by
the dedicated configuration options yet. The arguments are passed as-is to the
Alertmanager container which may cause issues if they are invalid or not supported
by the given Alertmanager version.</p>
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
<p>Defaults to 120 seconds.</p>
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
<code>selector</code><br/>
<em>
string
</em>
</td>
<td>
<p>The selector used to match the pods targeted by this Alertmanager object.</p>
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
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerSpec">AlertmanagerSpec</a>, <a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosSpec">ThanosSpec</a>)
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
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PodMonitorSpec">PodMonitorSpec</a>, <a href="#monitoring.coreos.com/v1.ScrapeClass">ScrapeClass</a>, <a href="#monitoring.coreos.com/v1.ServiceMonitorSpec">ServiceMonitorSpec</a>)
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
<p>When set to true, Prometheus attaches node metadata to the discovered
targets.</p>
<p>The Prometheus service account must have the <code>list</code> and <code>watch</code>
permissions on the <code>Nodes</code> objects.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.Authorization">Authorization
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.APIServerConfig">APIServerConfig</a>, <a href="#monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1.ScrapeClass">ScrapeClass</a>)
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<h3 id="monitoring.coreos.com/v1.AzureAD">AzureAD
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>)
</p>
<div>
<p>AzureAD defines the configuration for remote write&rsquo;s azuread parameters.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>cloud</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The Azure Cloud. Options are &lsquo;AzurePublic&rsquo;, &lsquo;AzureChina&rsquo;, or &lsquo;AzureGovernment&rsquo;.</p>
</td>
</tr>
<tr>
<td>
<code>managedIdentity</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ManagedIdentity">
ManagedIdentity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ManagedIdentity defines the Azure User-assigned Managed identity.
Cannot be set at the same time as <code>oauth</code> or <code>sdk</code>.</p>
</td>
</tr>
<tr>
<td>
<code>oauth</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AzureOAuth">
AzureOAuth
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>OAuth defines the oauth config that is being used to authenticate.
Cannot be set at the same time as <code>managedIdentity</code> or <code>sdk</code>.</p>
<p>It requires Prometheus &gt;= v2.48.0 or Thanos &gt;= v0.31.0.</p>
</td>
</tr>
<tr>
<td>
<code>sdk</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AzureSDK">
AzureSDK
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SDK defines the Azure SDK config that is being used to authenticate.
See <a href="https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication">https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication</a>
Cannot be set at the same time as <code>oauth</code> or <code>managedIdentity</code>.</p>
<p>It requires Prometheus &gt;= v2.52.0 or Thanos &gt;= v0.36.0.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AzureOAuth">AzureOAuth
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AzureAD">AzureAD</a>)
</p>
<div>
<p>AzureOAuth defines the Azure OAuth settings.</p>
</div>
<table>
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
string
</em>
</td>
<td>
<p><code>clientID</code> is the clientId of the Azure Active Directory application that is being used to authenticate.</p>
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
<p><code>clientSecret</code> specifies a key of a Secret containing the client secret of the Azure Active Directory application that is being used to authenticate.</p>
</td>
</tr>
<tr>
<td>
<code>tenantId</code><br/>
<em>
string
</em>
</td>
<td>
<p><code>tenantId</code> is the tenant ID of the Azure Active Directory application that is being used to authenticate.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.AzureSDK">AzureSDK
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AzureAD">AzureAD</a>)
</p>
<div>
<p>AzureSDK is used to store azure SDK config values.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>tenantId</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>tenantId</code> is the tenant ID of the azure active directory application that is being used to authenticate.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.BasicAuth">BasicAuth
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.APIServerConfig">APIServerConfig</a>, <a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>, <a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.AzureSDConfig">AzureSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ConsulSDConfig">ConsulSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DockerSDConfig">DockerSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DockerSwarmSDConfig">DockerSwarmSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.EurekaSDConfig">EurekaSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPSDConfig">HTTPSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HetznerSDConfig">HetznerSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.KubernetesSDConfig">KubernetesSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.KumaSDConfig">KumaSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.LightSailSDConfig">LightSailSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.NomadSDConfig">NomadSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.PuppetDBSDConfig">PuppetDBSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>, <a href="#monitoring.coreos.com/v1beta1.HTTPConfig">HTTPConfig</a>)
</p>
<div>
<p>BasicAuth configures HTTP Basic Authentication settings.</p>
</div>
<table>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p><code>username</code> specifies a key of a Secret containing the username for
authentication.</p>
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
<p><code>password</code> specifies a key of a Secret containing the password for
authentication.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ByteSize">ByteSize
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerLimitsSpec">AlertmanagerLimitsSpec</a>, <a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.PodMonitorSpec">PodMonitorSpec</a>, <a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>, <a href="#monitoring.coreos.com/v1.ServiceMonitorSpec">ServiceMonitorSpec</a>)
</p>
<div>
<p>ByteSize is a valid memory size type based on powers-of-2, so 1KB is 1024B.
Supported units: B, KB, KiB, MB, MiB, GB, GiB, TB, TiB, PB, PiB, EB, EiB Ex: <code>512MB</code>.</p>
</div>
<h3 id="monitoring.coreos.com/v1.ClusterTLSConfig">ClusterTLSConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerSpec">AlertmanagerSpec</a>)
</p>
<div>
<p>ClusterTLSConfig defines the mutual TLS configuration for the Alertmanager cluster TLS protocol.</p>
</div>
<table>
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
<a href="#monitoring.coreos.com/v1.WebTLSConfig">
WebTLSConfig
</a>
</em>
</td>
<td>
<p>Server-side configuration for mutual TLS.</p>
</td>
</tr>
<tr>
<td>
<code>client</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
</a>
</em>
</td>
<td>
<p>Client-side configuration for mutual TLS.</p>
</td>
</tr>
</tbody>
</table>
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
<p>Number of seconds to wait until a scrape request times out.
The value cannot be greater than the scrape interval otherwise the operator will reject the resource.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeProtocols</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
[]ScrapeProtocol
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
<a href="#monitoring.coreos.com/v1.RemoteWriteMessageVersion">
[]RemoteWriteMessageVersion
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
<a href="#monitoring.coreos.com/v1.EnableFeature">
[]EnableFeature
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
<a href="#monitoring.coreos.com/v1.TopologySpreadConstraint">
[]TopologySpreadConstraint
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
<code>otlp</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OTLPConfig">
OTLPConfig
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
<a href="#monitoring.coreos.com/v1.DNSPolicy">
DNSPolicy
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
<a href="#monitoring.coreos.com/v1.PodDNSConfig">
PodDNSConfig
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
<a href="#monitoring.coreos.com/v1.NameValidationSchemeOptions">
NameValidationSchemeOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the validation scheme for metric and label names.</p>
</td>
</tr>
<tr>
<td>
<code>nameEscapingScheme</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.NameEscapingSchemeOptions">
NameEscapingSchemeOptions
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
<p>Whether to convert all scraped classic histograms into a native histogram with custom buckets.</p>
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
<a href="#monitoring.coreos.com/v1.PrometheusTracingConfig">
PrometheusTracingConfig
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
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
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
<a href="#monitoring.coreos.com/v1.ReloadStrategyType">
ReloadStrategyType
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
<a href="#monitoring.coreos.com/v1.ScrapeClass">
[]ScrapeClass
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
<a href="#monitoring.coreos.com/v1.ServiceDiscoveryRole">
ServiceDiscoveryRole
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
<a href="#monitoring.coreos.com/v1.TSDBSpec">
TSDBSpec
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
<a href="#monitoring.coreos.com/v1.RuntimeConfig">
RuntimeConfig
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#time-v1-meta">
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
<h3 id="monitoring.coreos.com/v1.CoreV1TopologySpreadConstraint">CoreV1TopologySpreadConstraint
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.TopologySpreadConstraint">TopologySpreadConstraint</a>)
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
<code>maxSkew</code><br/>
<em>
int32
</em>
</td>
<td>
<p>MaxSkew describes the degree to which pods may be unevenly distributed.
When <code>whenUnsatisfiable=DoNotSchedule</code>, it is the maximum permitted difference
between the number of matching pods in the target topology and the global minimum.
The global minimum is the minimum number of matching pods in an eligible domain
or zero if the number of eligible domains is less than MinDomains.
For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same
labelSelector spread as 2/2/1:
In this case, the global minimum is 1.
| zone1 | zone2 | zone3 |
|  P P  |  P P  |   P   |
- if MaxSkew is 1, incoming pod can only be scheduled to zone3 to become 2/2/2;
scheduling it onto zone1(zone2) would make the ActualSkew(3-1) on zone1(zone2)
violate MaxSkew(1).
- if MaxSkew is 2, incoming pod can be scheduled onto any zone.
When <code>whenUnsatisfiable=ScheduleAnyway</code>, it is used to give higher precedence
to topologies that satisfy it.
It&rsquo;s a required field. Default value is 1 and 0 is not allowed.</p>
</td>
</tr>
<tr>
<td>
<code>topologyKey</code><br/>
<em>
string
</em>
</td>
<td>
<p>TopologyKey is the key of node labels. Nodes that have a label with this key
and identical values are considered to be in the same topology.
We consider each <key, value> as a &ldquo;bucket&rdquo;, and try to put balanced number
of pods into each bucket.
We define a domain as a particular instance of a topology.
Also, we define an eligible domain as a domain whose nodes meet the requirements of
nodeAffinityPolicy and nodeTaintsPolicy.
e.g. If TopologyKey is &ldquo;kubernetes.io/hostname&rdquo;, each Node is a domain of that topology.
And, if TopologyKey is &ldquo;topology.kubernetes.io/zone&rdquo;, each zone is a domain of that topology.
It&rsquo;s a required field.</p>
</td>
</tr>
<tr>
<td>
<code>whenUnsatisfiable</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#unsatisfiableconstraintaction-v1-core">
Kubernetes core/v1.UnsatisfiableConstraintAction
</a>
</em>
</td>
<td>
<p>WhenUnsatisfiable indicates how to deal with a pod if it doesn&rsquo;t satisfy
the spread constraint.
- DoNotSchedule (default) tells the scheduler not to schedule it.
- ScheduleAnyway tells the scheduler to schedule the pod in any location,
but giving higher precedence to topologies that would help reduce the
skew.
A constraint is considered &ldquo;Unsatisfiable&rdquo; for an incoming pod
if and only if every possible node assignment for that pod would violate
&ldquo;MaxSkew&rdquo; on some topology.
For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same
labelSelector spread as 3/1/1:
| zone1 | zone2 | zone3 |
| P P P |   P   |   P   |
If WhenUnsatisfiable is set to DoNotSchedule, incoming pod can only be scheduled
to zone2(zone3) to become 3/2/1(3/1/2) as ActualSkew(2-1) on zone2(zone3) satisfies
MaxSkew(1). In other words, the cluster can still be imbalanced, but scheduler
won&rsquo;t make it <em>more</em> imbalanced.
It&rsquo;s a required field.</p>
</td>
</tr>
<tr>
<td>
<code>labelSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>LabelSelector is used to find matching pods.
Pods that match this label selector are counted to determine the number of pods
in their corresponding topology domain.</p>
</td>
</tr>
<tr>
<td>
<code>minDomains</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>MinDomains indicates a minimum number of eligible domains.
When the number of eligible domains with matching topology keys is less than minDomains,
Pod Topology Spread treats &ldquo;global minimum&rdquo; as 0, and then the calculation of Skew is performed.
And when the number of eligible domains with matching topology keys equals or greater than minDomains,
this value has no effect on scheduling.
As a result, when the number of eligible domains is less than minDomains,
scheduler won&rsquo;t schedule more than maxSkew Pods to those domains.
If value is nil, the constraint behaves as if MinDomains is equal to 1.
Valid values are integers greater than 0.
When value is not nil, WhenUnsatisfiable must be DoNotSchedule.</p>
<p>For example, in a 3-zone cluster, MaxSkew is set to 2, MinDomains is set to 5 and pods with the same
labelSelector spread as 2/2/2:
| zone1 | zone2 | zone3 |
|  P P  |  P P  |  P P  |
The number of domains is less than 5(MinDomains), so &ldquo;global minimum&rdquo; is treated as 0.
In this situation, new pod with the same labelSelector cannot be scheduled,
because computed skew will be 3(3 - 0) if new Pod is scheduled to any of the three zones,
it will violate MaxSkew.</p>
</td>
</tr>
<tr>
<td>
<code>nodeAffinityPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#nodeinclusionpolicy-v1-core">
Kubernetes core/v1.NodeInclusionPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeAffinityPolicy indicates how we will treat Pod&rsquo;s nodeAffinity/nodeSelector
when calculating pod topology spread skew. Options are:
- Honor: only nodes matching nodeAffinity/nodeSelector are included in the calculations.
- Ignore: nodeAffinity/nodeSelector are ignored. All nodes are included in the calculations.</p>
<p>If this value is nil, the behavior is equivalent to the Honor policy.</p>
</td>
</tr>
<tr>
<td>
<code>nodeTaintsPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#nodeinclusionpolicy-v1-core">
Kubernetes core/v1.NodeInclusionPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeTaintsPolicy indicates how we will treat node taints when calculating
pod topology spread skew. Options are:
- Honor: nodes without taints, along with tainted nodes for which the incoming pod
has a toleration, are included.
- Ignore: node taints are ignored. All nodes are included.</p>
<p>If this value is nil, the behavior is equivalent to the Ignore policy.</p>
</td>
</tr>
<tr>
<td>
<code>matchLabelKeys</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>MatchLabelKeys is a set of pod label keys to select the pods over which
spreading will be calculated. The keys are used to lookup values from the
incoming pod labels, those key-value labels are ANDed with labelSelector
to select the group of existing pods over which spreading will be calculated
for the incoming pod. The same key is forbidden to exist in both MatchLabelKeys and LabelSelector.
MatchLabelKeys cannot be set when LabelSelector isn&rsquo;t set.
Keys that don&rsquo;t exist in the incoming pod labels will
be ignored. A null or empty list means only match against labelSelector.</p>
<p>This is a beta field and requires the MatchLabelKeysInPodTopologySpread feature gate to be enabled (enabled by default).</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.DNSPolicy">DNSPolicy
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerSpec">AlertmanagerSpec</a>, <a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>)
</p>
<div>
<p>DNSPolicy specifies the DNS policy for the pod.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;ClusterFirst&#34;</p></td>
<td><p>DNSClusterFirst indicates that the pod should use cluster DNS
first unless hostNetwork is true, if it is available, then
fall back on the default (as determined by kubelet) DNS settings.</p>
</td>
</tr><tr><td><p>&#34;ClusterFirstWithHostNet&#34;</p></td>
<td><p>DNSClusterFirstWithHostNet indicates that the pod should use cluster DNS
first, if it is available, then fall back on the default
(as determined by kubelet) DNS settings.</p>
</td>
</tr><tr><td><p>&#34;Default&#34;</p></td>
<td><p>DNSDefault indicates that the pod should use the default (as
determined by kubelet) DNS settings.</p>
</td>
</tr><tr><td><p>&#34;None&#34;</p></td>
<td><p>DNSNone indicates that the pod should use empty DNS settings. DNS
parameters such as nameservers and search paths should be defined via
DNSConfig.</p>
</td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1.Duration">Duration
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>, <a href="#monitoring.coreos.com/v1.AlertmanagerGlobalConfig">AlertmanagerGlobalConfig</a>, <a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.MetadataConfig">MetadataConfig</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1.PrometheusSpec">PrometheusSpec</a>, <a href="#monitoring.coreos.com/v1.PrometheusTracingConfig">PrometheusTracingConfig</a>, <a href="#monitoring.coreos.com/v1.QuerySpec">QuerySpec</a>, <a href="#monitoring.coreos.com/v1.QueueConfig">QueueConfig</a>, <a href="#monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1.RetainConfig">RetainConfig</a>, <a href="#monitoring.coreos.com/v1.Rule">Rule</a>, <a href="#monitoring.coreos.com/v1.RuleGroup">RuleGroup</a>, <a href="#monitoring.coreos.com/v1.TSDBSpec">TSDBSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosSpec">ThanosSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.AzureSDConfig">AzureSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ConsulSDConfig">ConsulSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DNSSDConfig">DNSSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DigitalOceanSDConfig">DigitalOceanSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DockerSDConfig">DockerSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DockerSwarmSDConfig">DockerSwarmSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.EC2SDConfig">EC2SDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.EurekaSDConfig">EurekaSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.FileSDConfig">FileSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.GCESDConfig">GCESDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPSDConfig">HTTPSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HetznerSDConfig">HetznerSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.IonosSDConfig">IonosSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.KumaSDConfig">KumaSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.LightSailSDConfig">LightSailSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.LinodeSDConfig">LinodeSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.NomadSDConfig">NomadSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.OVHCloudSDConfig">OVHCloudSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.OpenStackSDConfig">OpenStackSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.PuppetDBSDConfig">PuppetDBSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.PushoverConfig">PushoverConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ScalewaySDConfig">ScalewaySDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.WebhookConfig">WebhookConfig</a>, <a href="#monitoring.coreos.com/v1beta1.PushoverConfig">PushoverConfig</a>, <a href="#monitoring.coreos.com/v1beta1.WebhookConfig">WebhookConfig</a>)
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#persistentvolumeclaimspec-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#persistentvolumeaccessmode-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volumeresourcerequirements-v1-core">
Kubernetes core/v1.VolumeResourceRequirements
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#persistentvolumemode-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#typedlocalobjectreference-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#typedobjectreference-v1-core">
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
<tr>
<td>
<code>volumeAttributesClassName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>volumeAttributesClassName may be used to set the VolumeAttributesClass used by this claim.
If specified, the CSI driver will create or update the volume with the attributes defined
in the corresponding VolumeAttributesClass. This has a different purpose than storageClassName,
it can be changed after the claim is created. An empty string value means that no VolumeAttributesClass
will be applied to the claim but it&rsquo;s not allowed to reset this field to empty string once it is set.
If unspecified and the PersistentVolumeClaim is unbound, the default VolumeAttributesClass
will be set by the persistentvolume controller if it exists.
If the resource referred to by volumeAttributesClass does not exist, this PersistentVolumeClaim will be
set to a Pending state, as reflected by the modifyVolumeStatus field, until such as a resource
exists.
More info: <a href="https://kubernetes.io/docs/concepts/storage/volume-attributes-classes/">https://kubernetes.io/docs/concepts/storage/volume-attributes-classes/</a>
(Beta) Using this field requires the VolumeAttributesClass feature gate to be enabled (off by default).</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#persistentvolumeclaimstatus-v1-core">
Kubernetes core/v1.PersistentVolumeClaimStatus
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Deprecated: this field is never set.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.EnableFeature">EnableFeature
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>)
</p>
<div>
</div>
<h3 id="monitoring.coreos.com/v1.Endpoint">Endpoint
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ServiceMonitorSpec">ServiceMonitorSpec</a>)
</p>
<div>
<p>Endpoint defines an endpoint serving Prometheus metrics to be scraped by
Prometheus.</p>
</div>
<table>
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
<p>Name of the Service port which this endpoint refers to.</p>
<p>It takes precedence over <code>targetPort</code>.</p>
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
<em>(Optional)</em>
<p>Name or number of the target port of the <code>Pod</code> object behind the
Service. The port must be specified with the container&rsquo;s port property.</p>
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
<p>HTTP path from which to scrape for metrics.</p>
<p>If empty, Prometheus uses the default value (e.g. <code>/metrics</code>).</p>
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
<p><code>http</code> and <code>https</code> are the expected values unless you rewrite the
<code>__scheme__</code> label via relabeling.</p>
<p>If empty, Prometheus uses the default value <code>http</code>.</p>
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
<p>params define optional HTTP URL parameters.</p>
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
<p>Interval at which Prometheus scrapes the metrics from the target.</p>
<p>If empty, Prometheus uses the global scrape interval.</p>
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
<p>Timeout after which Prometheus considers the scrape to be failed.</p>
<p>If empty, Prometheus uses the global scrape timeout unless it is less
than the target&rsquo;s scrape interval value in which the latter is used.
The value cannot be greater than the scrape interval otherwise the operator will reject the resource.</p>
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
<p>TLS configuration to use when scraping the target.</p>
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
<p>File to read bearer token for scraping the target.</p>
<p>Deprecated: use <code>authorization</code> instead.</p>
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
<p><code>bearerTokenSecret</code> specifies a key of a Secret containing the bearer
token for scraping targets. The secret needs to be in the same namespace
as the ServiceMonitor object and readable by the Prometheus Operator.</p>
<p>Deprecated: use <code>authorization</code> instead.</p>
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
<p><code>authorization</code> configures the Authorization header credentials to use when
scraping the target.</p>
<p>Cannot be set at the same time as <code>basicAuth</code>, or <code>oauth2</code>.</p>
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
<p>When true, <code>honorLabels</code> preserves the metric&rsquo;s labels when they collide
with the target&rsquo;s labels.</p>
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
<p><code>honorTimestamps</code> controls whether Prometheus preserves the timestamps
when exposed by the target.</p>
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
<p><code>trackTimestampsStaleness</code> defines whether Prometheus tracks staleness of
the metrics that have an explicit timestamp present in scraped data.
Has no effect if <code>honorTimestamps</code> is false.</p>
<p>It requires Prometheus &gt;= v2.48.0.</p>
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
<p><code>basicAuth</code> configures the Basic Authentication credentials to use when
scraping the target.</p>
<p>Cannot be set at the same time as <code>authorization</code>, or <code>oauth2</code>.</p>
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
<p><code>oauth2</code> configures the OAuth2 settings to use when scraping the target.</p>
<p>It requires Prometheus &gt;= 2.27.0.</p>
<p>Cannot be set at the same time as <code>authorization</code>, or <code>basicAuth</code>.</p>
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
<em>(Optional)</em>
<p><code>metricRelabelings</code> configures the relabeling rules to apply to the
samples before ingestion.</p>
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
<p><code>relabelings</code> configures the relabeling rules to apply the target&rsquo;s
metadata labels.</p>
<p>The Operator automatically adds relabelings for a few standard Kubernetes fields.</p>
<p>The original scrape job&rsquo;s name is available via the <code>__tmp_prometheus_job_name</code> label.</p>
<p>More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config</a></p>
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
<p><code>proxyURL</code> configures the HTTP Proxy URL (e.g.
&ldquo;<a href="http://proxyserver:2195&quot;)">http://proxyserver:2195&rdquo;)</a> to go through when scraping the target.</p>
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
<p><code>followRedirects</code> defines whether the scrape requests should follow HTTP
3xx redirects.</p>
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
<p><code>enableHttp2</code> can be used to disable HTTP2 when scraping the target.</p>
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
<em>(Optional)</em>
<p>When true, the pods which are not running (e.g. either in Failed or
Succeeded state) are dropped during the target discovery.</p>
<p>If unset, the filtering is enabled.</p>
<p>More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase">https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase</a></p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<p>The default TLS configuration for SMTP receivers</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<p>LabelName is a valid Prometheus label name which may only contain ASCII
letters, numbers, as well as underscores.</p>
</div>
<h3 id="monitoring.coreos.com/v1.ManagedIdentity">ManagedIdentity
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AzureAD">AzureAD</a>)
</p>
<div>
<p>ManagedIdentity defines the Azure User-assigned Managed identity.</p>
</div>
<table>
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
string
</em>
</td>
<td>
<p>The client id</p>
</td>
</tr>
</tbody>
</table>
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
<tr>
<td>
<code>maxSamplesPerSend</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>MaxSamplesPerSend is the maximum number of metadata samples per send.</p>
<p>It requires Prometheus &gt;= v2.29.0.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.NameEscapingSchemeOptions">NameEscapingSchemeOptions
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>)
</p>
<div>
<p>Specifies the character escaping scheme that will be requested when scraping
for metric and label names that do not conform to the legacy Prometheus
character set.</p>
<p>Supported values are:
- <code>AllowUTF8NameEscapingScheme</code> for Full UTF-8 support, no escaping needed.
- <code>UnderscoresNameEscapingScheme</code> for Escape all legacy-invalid characters to underscores.
- <code>DotsNameEscapingScheme</code> for Escapes dots to <code>_dot_</code>, underscores to <code>__</code>, and all other
legacy-invalid characters to underscores.
- <code>ValuesNameEscapingScheme</code> for Prepend the name with <code>U__</code> and replace all invalid
characters with their unicode value, surrounded by underscores.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;AllowUTF8&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Dots&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Underscores&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Values&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1.NameValidationSchemeOptions">NameValidationSchemeOptions
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>)
</p>
<div>
<p>Specifies the validation scheme for metric and label names.</p>
<p>Supported values are:
- <code>UTF8NameValidationScheme</code> for UTF-8 support.
- <code>LegacyNameValidationScheme</code> for letters, numbers, colons, and underscores.</p>
<p>Note that <code>LegacyNameValidationScheme</code> cannot be used along with the OpenTelemetry <code>NoUTF8EscapingWithSuffixes</code> translation strategy (if enabled).</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Legacy&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;UTF8&#34;</p></td>
<td></td>
</tr></tbody>
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
<h3 id="monitoring.coreos.com/v1.NativeHistogramConfig">NativeHistogramConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PodMonitorSpec">PodMonitorSpec</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1.ServiceMonitorSpec">ServiceMonitorSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>NativeHistogramConfig extends the native histogram configuration settings.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
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
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.AzureSDConfig">AzureSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ConsulSDConfig">ConsulSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DigitalOceanSDConfig">DigitalOceanSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DockerSDConfig">DockerSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DockerSwarmSDConfig">DockerSwarmSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.EurekaSDConfig">EurekaSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPSDConfig">HTTPSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HetznerSDConfig">HetznerSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.IonosSDConfig">IonosSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.KubernetesSDConfig">KubernetesSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.KumaSDConfig">KumaSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.LightSailSDConfig">LightSailSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.LinodeSDConfig">LinodeSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.NomadSDConfig">NomadSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.PuppetDBSDConfig">PuppetDBSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>, <a href="#monitoring.coreos.com/v1beta1.HTTPConfig">HTTPConfig</a>)
</p>
<div>
<p>OAuth2 configures OAuth2 settings.</p>
</div>
<table>
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
<p><code>clientId</code> specifies a key of a Secret or ConfigMap containing the
OAuth2 client&rsquo;s ID.</p>
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
<p><code>clientSecret</code> specifies a key of a Secret containing the OAuth2
client&rsquo;s secret.</p>
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
<p><code>tokenURL</code> configures the URL to fetch the token from.</p>
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
<p><code>scopes</code> defines the OAuth2 scopes used for the token request.</p>
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
<em>(Optional)</em>
<p><code>endpointParams</code> configures the HTTP parameters to append to the token
URL.</p>
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
<p>TLS configuration to use when connecting to the OAuth2 server.
It requires Prometheus &gt;= v2.43.0.</p>
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
<h3 id="monitoring.coreos.com/v1.OTLPConfig">OTLPConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>)
</p>
<div>
<p>OTLPConfig is the configuration for writing to the OTLP endpoint.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>promoteResourceAttributes</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of OpenTelemetry Attributes that should be promoted to metric labels, defaults to none.</p>
</td>
</tr>
<tr>
<td>
<code>translationStrategy</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TranslationStrategyOption">
TranslationStrategyOption
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures how the OTLP receiver endpoint translates the incoming metrics.</p>
<p>It requires Prometheus &gt;= v3.0.0.</p>
</td>
</tr>
<tr>
<td>
<code>keepIdentifyingResourceAttributes</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enables adding <code>service.name</code>, <code>service.namespace</code> and <code>service.instance.id</code>
resource attributes to the <code>target_info</code> metric, on top of converting them into the <code>instance</code> and <code>job</code> labels.</p>
<p>It requires Prometheus &gt;= v3.1.0.</p>
</td>
</tr>
<tr>
<td>
<code>convertHistogramsToNHCB</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures optional translation of OTLP explicit bucket histograms into native histograms with custom buckets.
It requires Prometheus &gt;= v3.4.0.</p>
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
<h3 id="monitoring.coreos.com/v1.PodDNSConfig">PodDNSConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerSpec">AlertmanagerSpec</a>, <a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>)
</p>
<div>
<p>PodDNSConfig defines the DNS parameters of a pod in addition to
those generated from DNSPolicy.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>nameservers</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>A list of DNS name server IP addresses.
This will be appended to the base nameservers generated from DNSPolicy.</p>
</td>
</tr>
<tr>
<td>
<code>searches</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>A list of DNS search domains for host-name lookup.
This will be appended to the base search paths generated from DNSPolicy.</p>
</td>
</tr>
<tr>
<td>
<code>options</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.PodDNSConfigOption">
[]PodDNSConfigOption
</a>
</em>
</td>
<td>
<p>A list of DNS resolver options.
This will be merged with the base options generated from DNSPolicy.
Resolution options given in Options
will override those that appear in the base DNSPolicy.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.PodDNSConfigOption">PodDNSConfigOption
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PodDNSConfig">PodDNSConfig</a>)
</p>
<div>
<p>PodDNSConfigOption defines DNS resolver options of a pod.</p>
</div>
<table>
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
<p>Name is required and must be unique.</p>
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
<p>Value is optional.</p>
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
<p>PodMetricsEndpoint defines an endpoint serving Prometheus metrics to be scraped by
Prometheus.</p>
</div>
<table>
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
<em>(Optional)</em>
<p>The <code>Pod</code> port name which exposes the endpoint.</p>
<p>It takes precedence over the <code>portNumber</code> and <code>targetPort</code> fields.</p>
</td>
</tr>
<tr>
<td>
<code>portNumber</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The <code>Pod</code> port number which exposes the endpoint.</p>
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
<p>Name or number of the target port of the <code>Pod</code> object behind the Service, the
port must be specified with container port property.</p>
<p>Deprecated: use &lsquo;port&rsquo; or &lsquo;portNumber&rsquo; instead.</p>
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
<p>HTTP path from which to scrape for metrics.</p>
<p>If empty, Prometheus uses the default value (e.g. <code>/metrics</code>).</p>
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
<p><code>http</code> and <code>https</code> are the expected values unless you rewrite the
<code>__scheme__</code> label via relabeling.</p>
<p>If empty, Prometheus uses the default value <code>http</code>.</p>
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
<p><code>params</code> define optional HTTP URL parameters.</p>
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
<p>Interval at which Prometheus scrapes the metrics from the target.</p>
<p>If empty, Prometheus uses the global scrape interval.</p>
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
<p>Timeout after which Prometheus considers the scrape to be failed.</p>
<p>If empty, Prometheus uses the global scrape timeout unless it is less
than the target&rsquo;s scrape interval value in which the latter is used.
The value cannot be greater than the scrape interval otherwise the operator will reject the resource.</p>
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
<p>TLS configuration to use when scraping the target.</p>
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
<p><code>bearerTokenSecret</code> specifies a key of a Secret containing the bearer
token for scraping targets. The secret needs to be in the same namespace
as the PodMonitor object and readable by the Prometheus Operator.</p>
<p>Deprecated: use <code>authorization</code> instead.</p>
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
<p>When true, <code>honorLabels</code> preserves the metric&rsquo;s labels when they collide
with the target&rsquo;s labels.</p>
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
<p><code>honorTimestamps</code> controls whether Prometheus preserves the timestamps
when exposed by the target.</p>
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
<p><code>trackTimestampsStaleness</code> defines whether Prometheus tracks staleness of
the metrics that have an explicit timestamp present in scraped data.
Has no effect if <code>honorTimestamps</code> is false.</p>
<p>It requires Prometheus &gt;= v2.48.0.</p>
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
<p><code>basicAuth</code> configures the Basic Authentication credentials to use when
scraping the target.</p>
<p>Cannot be set at the same time as <code>authorization</code>, or <code>oauth2</code>.</p>
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
<p><code>oauth2</code> configures the OAuth2 settings to use when scraping the target.</p>
<p>It requires Prometheus &gt;= 2.27.0.</p>
<p>Cannot be set at the same time as <code>authorization</code>, or <code>basicAuth</code>.</p>
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
<p><code>authorization</code> configures the Authorization header credentials to use when
scraping the target.</p>
<p>Cannot be set at the same time as <code>basicAuth</code>, or <code>oauth2</code>.</p>
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
<em>(Optional)</em>
<p><code>metricRelabelings</code> configures the relabeling rules to apply to the
samples before ingestion.</p>
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
<p><code>relabelings</code> configures the relabeling rules to apply the target&rsquo;s
metadata labels.</p>
<p>The Operator automatically adds relabelings for a few standard Kubernetes fields.</p>
<p>The original scrape job&rsquo;s name is available via the <code>__tmp_prometheus_job_name</code> label.</p>
<p>More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config</a></p>
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
<p><code>proxyURL</code> configures the HTTP Proxy URL (e.g.
&ldquo;<a href="http://proxyserver:2195&quot;)">http://proxyserver:2195&rdquo;)</a> to go through when scraping the target.</p>
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
<p><code>followRedirects</code> defines whether the scrape requests should follow HTTP
3xx redirects.</p>
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
<p><code>enableHttp2</code> can be used to disable HTTP2 when scraping the target.</p>
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
<em>(Optional)</em>
<p>When true, the pods which are not running (e.g. either in Failed or
Succeeded state) are dropped during the target discovery.</p>
<p>If unset, the filtering is enabled.</p>
<p>More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase">https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase</a></p>
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
<p>The label to use to retrieve the job name from.
<code>jobLabel</code> selects the label from the associated Kubernetes <code>Pod</code>
object which will be used as the <code>job</code> label for all metrics.</p>
<p>For example if <code>jobLabel</code> is set to <code>foo</code> and the Kubernetes <code>Pod</code>
object is labeled with <code>foo: bar</code>, then Prometheus adds the <code>job=&quot;bar&quot;</code>
label to all ingested metrics.</p>
<p>If the value of this field is empty, the <code>job</code> label of the metrics
defaults to the namespace and name of the PodMonitor object (e.g. <code>&lt;namespace&gt;/&lt;name&gt;</code>).</p>
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
<p><code>podTargetLabels</code> defines the labels which are transferred from the
associated Kubernetes <code>Pod</code> object onto the ingested metrics.</p>
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
<em>(Optional)</em>
<p>Defines how to scrape metrics from the selected pods.</p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Label selector to select the Kubernetes <code>Pod</code> objects to scrape metrics from.</p>
</td>
</tr>
<tr>
<td>
<code>selectorMechanism</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SelectorMechanism">
SelectorMechanism
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Mechanism used to select the endpoints to scrape.
By default, the selection process relies on relabel configurations to filter the discovered targets.
Alternatively, you can opt in for role selectors, which may offer better efficiency in large clusters.
Which strategy is best for your use case needs to be carefully evaluated.</p>
<p>It requires Prometheus &gt;= v2.17.0.</p>
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
<p><code>namespaceSelector</code> defines in which namespace(s) Prometheus should discover the pods.
By default, the pods are discovered in the same namespace as the <code>PodMonitor</code> object but it is possible to select pods across different/all namespaces.</p>
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
<p><code>sampleLimit</code> defines a per-scrape limit on the number of scraped samples
that will be accepted.</p>
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
<p><code>targetLimit</code> defines a limit on the number of scraped targets that will
be accepted.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeProtocols</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
[]ScrapeProtocol
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>scrapeProtocols</code> defines the protocols to negotiate during a scrape. It tells clients the
protocols supported by Prometheus in order of preference (from most to least preferred).</p>
<p>If unset, Prometheus uses its default value.</p>
<p>It requires Prometheus &gt;= v2.49.0.</p>
</td>
</tr>
<tr>
<td>
<code>fallbackScrapeProtocol</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
ScrapeProtocol
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
<code>labelLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Per-scrape limit on number of labels that will be accepted for a sample.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
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
<p>Per-scrape limit on length of labels name that will be accepted for a sample.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
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
<p>Per-scrape limit on length of labels value that will be accepted for a sample.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
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
<code>attachMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AttachMetadata">
AttachMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>attachMetadata</code> defines additional metadata which is added to the
discovered targets.</p>
<p>It requires Prometheus &gt;= v2.35.0.</p>
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
<tr>
<td>
<code>bodySizeLimit</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>When defined, bodySizeLimit specifies a job level limit on the size
of uncompressed response body that will be accepted by Prometheus.</p>
<p>It requires Prometheus &gt;= v2.28.0.</p>
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
If not specified, the Prometheus global scrape timeout is used.
The value cannot be greater than the scrape interval otherwise the operator will reject the resource.</p>
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
<p>TLS configuration to use when scraping the endpoint.</p>
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
<code>scrapeProtocols</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
[]ScrapeProtocol
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>scrapeProtocols</code> defines the protocols to negotiate during a scrape. It tells clients the
protocols supported by Prometheus in order of preference (from most to least preferred).</p>
<p>If unset, Prometheus uses its default value.</p>
<p>It requires Prometheus &gt;= v2.49.0.</p>
</td>
</tr>
<tr>
<td>
<code>fallbackScrapeProtocol</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
ScrapeProtocol
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
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
<p>Number of seconds to wait until a scrape request times out.
The value cannot be greater than the scrape interval otherwise the operator will reject the resource.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeProtocols</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
[]ScrapeProtocol
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
<a href="#monitoring.coreos.com/v1.RemoteWriteMessageVersion">
[]RemoteWriteMessageVersion
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
<a href="#monitoring.coreos.com/v1.EnableFeature">
[]EnableFeature
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
<a href="#monitoring.coreos.com/v1.TopologySpreadConstraint">
[]TopologySpreadConstraint
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
<code>otlp</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OTLPConfig">
OTLPConfig
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
<a href="#monitoring.coreos.com/v1.DNSPolicy">
DNSPolicy
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
<a href="#monitoring.coreos.com/v1.PodDNSConfig">
PodDNSConfig
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
<a href="#monitoring.coreos.com/v1.NameValidationSchemeOptions">
NameValidationSchemeOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the validation scheme for metric and label names.</p>
</td>
</tr>
<tr>
<td>
<code>nameEscapingScheme</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.NameEscapingSchemeOptions">
NameEscapingSchemeOptions
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
<p>Whether to convert all scraped classic histograms into a native histogram with custom buckets.</p>
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
<a href="#monitoring.coreos.com/v1.PrometheusTracingConfig">
PrometheusTracingConfig
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
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
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
<a href="#monitoring.coreos.com/v1.ReloadStrategyType">
ReloadStrategyType
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
<a href="#monitoring.coreos.com/v1.ScrapeClass">
[]ScrapeClass
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
<a href="#monitoring.coreos.com/v1.ServiceDiscoveryRole">
ServiceDiscoveryRole
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
<a href="#monitoring.coreos.com/v1.TSDBSpec">
TSDBSpec
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
<a href="#monitoring.coreos.com/v1.RuntimeConfig">
RuntimeConfig
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
<tr>
<td>
<code>baseImage</code><br/>
<em>
string
</em>
</td>
<td>
<p>Deprecated: use &lsquo;spec.image&rsquo; instead.</p>
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
<p>Deprecated: use &lsquo;spec.image&rsquo; instead. The image&rsquo;s tag can be specified as part of the image name.</p>
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
<p>Deprecated: use &lsquo;spec.image&rsquo; instead. The image&rsquo;s digest can be specified as part of the image name.</p>
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
<code>shardRetentionPolicy</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ShardRetentionPolicy">
ShardRetentionPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ShardRetentionPolicy defines the retention policy for the Prometheus shards.
(Alpha) Using this field requires the &lsquo;PrometheusShardRetentionPolicy&rsquo; feature gate to be enabled.</p>
<p>The final goals for this feature can be seen at <a href="https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/proposals/202310-shard-autoscaling.md#graceful-scale-down-of-prometheus-servers">https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/proposals/202310-shard-autoscaling.md#graceful-scale-down-of-prometheus-servers</a>,
however, the feature is not yet fully implemented in this PR. The limitation being:
* Retention duration is not settable, for now, shards are retained forever.</p>
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
<p>When true, the Prometheus compaction is disabled.
When <code>spec.thanos.objectStorageConfig</code> or <code>spec.objectStorageConfigFile</code> are defined, the operator automatically
disables block compaction to avoid race conditions during block uploads (as the Thanos documentation recommends).</p>
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
Deprecated: use <code>spec.excludedFromEnforcement</code> instead.</p>
</td>
</tr>
<tr>
<td>
<code>ruleSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<p>Deprecated: this flag has no effect for Prometheus &gt;= 2.39.0 where overlapping blocks are enabled by default.</p>
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
<code>ruleQueryOffset</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the offset the rule evaluation timestamp of this particular group by the specified duration into the past.
It requires Prometheus &gt;= v2.53.0.</p>
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
<tr>
<td>
<code>shards</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Shards is the most recently observed number of shards.</p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
string
</em>
</td>
<td>
<p>The selector used to match the pods targeted by this Prometheus resource.</p>
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
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/api/resource#Quantity">
k8s.io/apimachinery/pkg/api/resource.Quantity
</a>
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
<h3 id="monitoring.coreos.com/v1.ProxyConfig">ProxyConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>, <a href="#monitoring.coreos.com/v1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1.OAuth2">OAuth2</a>, <a href="#monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.AzureSDConfig">AzureSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ConsulSDConfig">ConsulSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DigitalOceanSDConfig">DigitalOceanSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DockerSDConfig">DockerSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DockerSwarmSDConfig">DockerSwarmSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.EC2SDConfig">EC2SDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.EurekaSDConfig">EurekaSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPSDConfig">HTTPSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HetznerSDConfig">HetznerSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.IonosSDConfig">IonosSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.KubernetesSDConfig">KubernetesSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.KumaSDConfig">KumaSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.LightSailSDConfig">LightSailSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.LinodeSDConfig">LinodeSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.NomadSDConfig">NomadSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.PuppetDBSDConfig">PuppetDBSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ScalewaySDConfig">ScalewaySDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>, <a href="#monitoring.coreos.com/v1beta1.HTTPConfig">HTTPConfig</a>)
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>MinBackoff is the initial retry delay. Gets doubled for every retry.</p>
</td>
</tr>
<tr>
<td>
<code>maxBackoff</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
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
<p>Retry upon receiving a 429 status code from the remote-write storage.</p>
<p>This is an <em>experimental feature</em>, it may change in any upcoming release
in a breaking way.</p>
</td>
</tr>
<tr>
<td>
<code>sampleAgeLimit</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SampleAgeLimit drops samples older than the limit.
It requires Prometheus &gt;= v2.50.0 or Thanos &gt;= v0.32.0.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.RelabelConfig">RelabelConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>, <a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1.ProbeTargetIngress">ProbeTargetIngress</a>, <a href="#monitoring.coreos.com/v1.ProbeTargetStaticConfig">ProbeTargetStaticConfig</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1.ScrapeClass">ScrapeClass</a>, <a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
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
<em>(Optional)</em>
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
<h3 id="monitoring.coreos.com/v1.ReloadStrategyType">ReloadStrategyType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>)
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
<tbody><tr><td><p>&#34;HTTP&#34;</p></td>
<td><p>HTTPReloadStrategyType reloads the configuration using the /-/reload HTTP endpoint.</p>
</td>
</tr><tr><td><p>&#34;ProcessSignal&#34;</p></td>
<td><p>ProcessSignalReloadStrategyType reloads the configuration by sending a SIGHUP signal to the process.</p>
</td>
</tr></tbody>
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
<em>(Optional)</em>
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
<p>Deprecated: this will be removed in a future release. Prefer using <code>authorization</code>.</p>
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
<p>Deprecated: this will be removed in a future release.</p>
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
<h3 id="monitoring.coreos.com/v1.RemoteWriteMessageVersion">RemoteWriteMessageVersion
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>)
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
<tbody><tr><td><p>&#34;V1.0&#34;</p></td>
<td><p>Remote Write message&rsquo;s version 1.0.</p>
</td>
</tr><tr><td><p>&#34;V2.0&#34;</p></td>
<td><p>Remote Write message&rsquo;s version 2.0.</p>
</td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>)
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
<em>(Optional)</em>
<p>The name of the remote write queue, it must be unique if specified. The
name is used in metrics and logging in order to differentiate queues.</p>
<p>It requires Prometheus &gt;= v2.15.0 or Thanos &gt;= 0.24.0.</p>
</td>
</tr>
<tr>
<td>
<code>messageVersion</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RemoteWriteMessageVersion">
RemoteWriteMessageVersion
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The Remote Write message&rsquo;s version to use when writing to the endpoint.</p>
<p><code>Version1.0</code> corresponds to the <code>prometheus.WriteRequest</code> protobuf message introduced in Remote Write 1.0.
<code>Version2.0</code> corresponds to the <code>io.prometheus.write.v2.Request</code> protobuf message introduced in Remote Write 2.0.</p>
<p>When <code>Version2.0</code> is selected, Prometheus will automatically be
configured to append the metadata of scraped metrics to the WAL.</p>
<p>Before setting this field, consult with your remote storage provider
what message version it supports.</p>
<p>It requires Prometheus &gt;= v2.54.0 or Thanos &gt;= v0.37.0.</p>
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
exemplar-storage itself must be enabled using the <code>spec.enableFeatures</code>
option for exemplars to be scraped in the first place.</p>
<p>It requires Prometheus &gt;= v2.27.0 or Thanos &gt;= v0.24.0.</p>
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
<p>It requires Prometheus &gt;= v2.40.0 or Thanos &gt;= v0.30.0.</p>
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
<em>(Optional)</em>
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
<p>It requires Prometheus &gt;= v2.25.0 or Thanos &gt;= v0.24.0.</p>
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
<p>It requires Prometheus &gt;= v2.27.0 or Thanos &gt;= v0.24.0.</p>
<p>Cannot be set at the same time as <code>sigv4</code>, <code>authorization</code>, <code>basicAuth</code>, or <code>azureAd</code>.</p>
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
<p>Cannot be set at the same time as <code>sigv4</code>, <code>authorization</code>, <code>oauth2</code>, or <code>azureAd</code>.</p>
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
<p>Deprecated: this will be removed in a future release. Prefer using <code>authorization</code>.</p>
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
<p>It requires Prometheus &gt;= v2.26.0 or Thanos &gt;= v0.24.0.</p>
<p>Cannot be set at the same time as <code>sigv4</code>, <code>basicAuth</code>, <code>oauth2</code>, or <code>azureAd</code>.</p>
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
<p>It requires Prometheus &gt;= v2.26.0 or Thanos &gt;= v0.24.0.</p>
<p>Cannot be set at the same time as <code>authorization</code>, <code>basicAuth</code>, <code>oauth2</code>, or <code>azureAd</code>.</p>
</td>
</tr>
<tr>
<td>
<code>azureAd</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AzureAD">
AzureAD
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AzureAD for the URL.</p>
<p>It requires Prometheus &gt;= v2.45.0 or Thanos &gt;= v0.31.0.</p>
<p>Cannot be set at the same time as <code>authorization</code>, <code>basicAuth</code>, <code>oauth2</code>, or <code>sigv4</code>.</p>
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
<p>Deprecated: this will be removed in a future release.</p>
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
<p>It requires Prometheus &gt;= v2.26.0 or Thanos &gt;= v0.24.0.</p>
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
<code>roundRobinDNS</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>When enabled:
- The remote-write mechanism will resolve the hostname via DNS.
- It will randomly select one of the resolved IP addresses and connect to it.</p>
<p>When disabled (default behavior):
- The Go standard library will handle hostname resolution.
- It will attempt connections to each resolved IP address sequentially.</p>
<p>Note: The connection timeout applies to the entire resolution and connection process.
If disabled, the timeout is distributed across all connection attempts.</p>
<p>It requires Prometheus &gt;= v3.1.0 or Thanos &gt;= v0.38.0.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.RetainConfig">RetainConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ShardRetentionPolicy">ShardRetentionPolicy</a>)
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
<code>retentionPeriod</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
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
<code>labels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Labels to add or overwrite before storing the result for its rules.
The labels defined at the rule level take precedence.</p>
<p>It requires Prometheus &gt;= 3.0.0.
The field is ignored for Thanos Ruler.</p>
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
<code>query_offset</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the offset the rule evaluation timestamp of this particular group by the specified duration into the past.</p>
<p>It requires Prometheus &gt;= v2.53.0.
It is not supported for ThanosRuler.</p>
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
<h3 id="monitoring.coreos.com/v1.RuntimeConfig">RuntimeConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>)
</p>
<div>
<p>RuntimeConfig configures the values for the process behavior.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>goGC</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The Go garbage collection target percentage. Lowering this number may increase the CPU usage.
See: <a href="https://tip.golang.org/doc/gc-guide#GOGC">https://tip.golang.org/doc/gc-guide#GOGC</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.SafeAuthorization">SafeAuthorization
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>, <a href="#monitoring.coreos.com/v1.Authorization">Authorization</a>, <a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.AzureSDConfig">AzureSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ConsulSDConfig">ConsulSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DigitalOceanSDConfig">DigitalOceanSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DockerSDConfig">DockerSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DockerSwarmSDConfig">DockerSwarmSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.EurekaSDConfig">EurekaSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPSDConfig">HTTPSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HetznerSDConfig">HetznerSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.IonosSDConfig">IonosSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.KubernetesSDConfig">KubernetesSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.KumaSDConfig">KumaSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.LightSailSDConfig">LightSailSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.LinodeSDConfig">LinodeSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.NomadSDConfig">NomadSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.PuppetDBSDConfig">PuppetDBSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>, <a href="#monitoring.coreos.com/v1beta1.HTTPConfig">HTTPConfig</a>)
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ClusterTLSConfig">ClusterTLSConfig</a>, <a href="#monitoring.coreos.com/v1.GlobalSMTPConfig">GlobalSMTPConfig</a>, <a href="#monitoring.coreos.com/v1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1.OAuth2">OAuth2</a>, <a href="#monitoring.coreos.com/v1.PodMetricsEndpoint">PodMetricsEndpoint</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1.TLSConfig">TLSConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.AzureSDConfig">AzureSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ConsulSDConfig">ConsulSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DigitalOceanSDConfig">DigitalOceanSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DockerSDConfig">DockerSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.DockerSwarmSDConfig">DockerSwarmSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.EC2SDConfig">EC2SDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.EmailConfig">EmailConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.EurekaSDConfig">EurekaSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPConfig">HTTPConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HTTPSDConfig">HTTPSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.HetznerSDConfig">HetznerSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.IonosSDConfig">IonosSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.KubernetesSDConfig">KubernetesSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.KumaSDConfig">KumaSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.LightSailSDConfig">LightSailSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.LinodeSDConfig">LinodeSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.NomadSDConfig">NomadSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.OpenStackSDConfig">OpenStackSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.PuppetDBSDConfig">PuppetDBSDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ScalewaySDConfig">ScalewaySDConfig</a>, <a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>, <a href="#monitoring.coreos.com/v1beta1.EmailConfig">EmailConfig</a>, <a href="#monitoring.coreos.com/v1beta1.HTTPConfig">HTTPConfig</a>)
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<em>(Optional)</em>
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
<em>(Optional)</em>
<p>Disable target certificate validation.</p>
</td>
</tr>
<tr>
<td>
<code>minVersion</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TLSVersion">
TLSVersion
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Minimum acceptable TLS version.</p>
<p>It requires Prometheus &gt;= v2.35.0 or Thanos &gt;= v0.28.0.</p>
</td>
</tr>
<tr>
<td>
<code>maxVersion</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TLSVersion">
TLSVersion
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Maximum acceptable TLS version.</p>
<p>It requires Prometheus &gt;= v2.41.0 or Thanos &gt;= v0.31.0.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ScrapeClass">ScrapeClass
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
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the scrape class.</p>
</td>
</tr>
<tr>
<td>
<code>default</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Default indicates that the scrape applies to all scrape objects that
don&rsquo;t configure an explicit scrape class name.</p>
<p>Only one scrape class can be set as the default.</p>
</td>
</tr>
<tr>
<td>
<code>fallbackScrapeProtocol</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
ScrapeProtocol
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The protocol to use if a scrape returns blank, unparseable, or otherwise invalid Content-Type.
It will only apply if the scrape resource doesn&rsquo;t specify any FallbackScrapeProtocol</p>
<p>It requires Prometheus &gt;= v3.0.0.</p>
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
<p>TLSConfig defines the TLS settings to use for the scrape. When the
scrape objects define their own CA, certificate and/or key, they take
precedence over the corresponding scrape class fields.</p>
<p>For now only the <code>caFile</code>, <code>certFile</code> and <code>keyFile</code> fields are supported.</p>
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
<p>Authorization section for the ScrapeClass.
It will only apply if the scrape resource doesn&rsquo;t specify any Authorization.</p>
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
<p>Relabelings configures the relabeling rules to apply to all scrape targets.</p>
<p>The Operator automatically adds relabelings for a few standard Kubernetes fields
like <code>__meta_kubernetes_namespace</code> and <code>__meta_kubernetes_service_name</code>.
Then the Operator adds the scrape class relabelings defined here.
Then the Operator adds the target-specific relabelings defined in the scrape object.</p>
<p>More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config</a></p>
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
<em>(Optional)</em>
<p>MetricRelabelings configures the relabeling rules to apply to all samples before ingestion.</p>
<p>The Operator adds the scrape class metric relabelings defined here.
Then the Operator adds the target-specific metric relabelings defined in ServiceMonitors, PodMonitors, Probes and ScrapeConfigs.
Then the Operator adds namespace enforcement relabeling rule, specified in &lsquo;.spec.enforcedNamespaceLabel&rsquo;.</p>
<p>More info: <a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#metric_relabel_configs">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#metric_relabel_configs</a></p>
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
<em>(Optional)</em>
<p>AttachMetadata configures additional metadata to the discovered targets.
When the scrape object defines its own configuration, it takes
precedence over the scrape class configuration.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ScrapeProtocol">ScrapeProtocol
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>, <a href="#monitoring.coreos.com/v1.PodMonitorSpec">PodMonitorSpec</a>, <a href="#monitoring.coreos.com/v1.ProbeSpec">ProbeSpec</a>, <a href="#monitoring.coreos.com/v1.ScrapeClass">ScrapeClass</a>, <a href="#monitoring.coreos.com/v1.ServiceMonitorSpec">ServiceMonitorSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.ScrapeConfigSpec">ScrapeConfigSpec</a>)
</p>
<div>
<p>ScrapeProtocol represents a protocol used by Prometheus for scraping metrics.
Supported values are:
* <code>OpenMetricsText0.0.1</code>
* <code>OpenMetricsText1.0.0</code>
* <code>PrometheusProto</code>
* <code>PrometheusText0.0.4</code>
* <code>PrometheusText1.0.0</code></p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;OpenMetricsText0.0.1&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;OpenMetricsText1.0.0&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;PrometheusProto&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;PrometheusText0.0.4&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;PrometheusText1.0.0&#34;</p></td>
<td></td>
</tr></tbody>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#configmapkeyselector-v1-core">
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
<h3 id="monitoring.coreos.com/v1.SelectorMechanism">SelectorMechanism
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.PodMonitorSpec">PodMonitorSpec</a>, <a href="#monitoring.coreos.com/v1.ServiceMonitorSpec">ServiceMonitorSpec</a>)
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
<tbody><tr><td><p>&#34;RelabelConfig&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;RoleSelector&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ServiceDiscoveryRole">ServiceDiscoveryRole
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.CommonPrometheusFields">CommonPrometheusFields</a>)
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
<tbody><tr><td><p>&#34;EndpointSlice&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Endpoints&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ServiceMonitorSpec">ServiceMonitorSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ServiceMonitor">ServiceMonitor</a>)
</p>
<div>
<p>ServiceMonitorSpec defines the specification parameters for a ServiceMonitor.</p>
</div>
<table>
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
<p><code>jobLabel</code> selects the label from the associated Kubernetes <code>Service</code>
object which will be used as the <code>job</code> label for all metrics.</p>
<p>For example if <code>jobLabel</code> is set to <code>foo</code> and the Kubernetes <code>Service</code>
object is labeled with <code>foo: bar</code>, then Prometheus adds the <code>job=&quot;bar&quot;</code>
label to all ingested metrics.</p>
<p>If the value of this field is empty or if the label doesn&rsquo;t exist for
the given Service, the <code>job</code> label of the metrics defaults to the name
of the associated Kubernetes <code>Service</code>.</p>
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
<em>(Optional)</em>
<p><code>targetLabels</code> defines the labels which are transferred from the
associated Kubernetes <code>Service</code> object onto the ingested metrics.</p>
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
<p><code>podTargetLabels</code> defines the labels which are transferred from the
associated Kubernetes <code>Pod</code> object onto the ingested metrics.</p>
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
<p>List of endpoints part of this ServiceMonitor.
Defines how to scrape metrics from Kubernetes <a href="https://kubernetes.io/docs/concepts/services-networking/service/#endpoints">Endpoints</a> objects.
In most cases, an Endpoints object is backed by a Kubernetes <a href="https://kubernetes.io/docs/concepts/services-networking/service/">Service</a> object with the same name and labels.</p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Label selector to select the Kubernetes <code>Endpoints</code> objects to scrape metrics from.</p>
</td>
</tr>
<tr>
<td>
<code>selectorMechanism</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SelectorMechanism">
SelectorMechanism
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Mechanism used to select the endpoints to scrape.
By default, the selection process relies on relabel configurations to filter the discovered targets.
Alternatively, you can opt in for role selectors, which may offer better efficiency in large clusters.
Which strategy is best for your use case needs to be carefully evaluated.</p>
<p>It requires Prometheus &gt;= v2.17.0.</p>
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
<p><code>namespaceSelector</code> defines in which namespace(s) Prometheus should discover the services.
By default, the services are discovered in the same namespace as the <code>ServiceMonitor</code> object but it is possible to select pods across different/all namespaces.</p>
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
<p><code>sampleLimit</code> defines a per-scrape limit on the number of scraped samples
that will be accepted.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeProtocols</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
[]ScrapeProtocol
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>scrapeProtocols</code> defines the protocols to negotiate during a scrape. It tells clients the
protocols supported by Prometheus in order of preference (from most to least preferred).</p>
<p>If unset, Prometheus uses its default value.</p>
<p>It requires Prometheus &gt;= v2.49.0.</p>
</td>
</tr>
<tr>
<td>
<code>fallbackScrapeProtocol</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
ScrapeProtocol
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
<code>targetLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>targetLimit</code> defines a limit on the number of scraped targets that will
be accepted.</p>
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
<p>Per-scrape limit on number of labels that will be accepted for a sample.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
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
<p>Per-scrape limit on length of labels name that will be accepted for a sample.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
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
<p>Per-scrape limit on length of labels value that will be accepted for a sample.</p>
<p>It requires Prometheus &gt;= v2.27.0.</p>
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
<code>attachMetadata</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AttachMetadata">
AttachMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>attachMetadata</code> defines additional metadata which is added to the
discovered targets.</p>
<p>It requires Prometheus &gt;= v2.37.0.</p>
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
<tr>
<td>
<code>bodySizeLimit</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>When defined, bodySizeLimit specifies a job level limit on the size
of uncompressed response body that will be accepted by Prometheus.</p>
<p>It requires Prometheus &gt;= v2.28.0.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ShardRetentionPolicy">ShardRetentionPolicy
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
<code>whenScaled</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.WhenScaledRetentionType">
WhenScaledRetentionType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the retention policy when the Prometheus shards are scaled down.
* <code>Delete</code>, the operator will delete the pods from the scaled-down shard(s).
* <code>Retain</code>, the operator will keep the pods from the scaled-down shard(s), so the data can still be queried.</p>
<p>If not defined, the operator assumes the <code>Delete</code> value.</p>
</td>
</tr>
<tr>
<td>
<code>retain</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.RetainConfig">
RetainConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the config for retention when the retention policy is set to <code>Retain</code>.
This field is ineffective as of now.</p>
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
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1alpha1.SNSConfig">SNSConfig</a>, <a href="#monitoring.coreos.com/v1beta1.SNSConfig">SNSConfig</a>)
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<p>Deprecated: subPath usage will be removed in a future release.</p>
</td>
</tr>
<tr>
<td>
<code>emptyDir</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#emptydirvolumesource-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#ephemeralvolumesource-v1-core">
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
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.APIServerConfig">APIServerConfig</a>, <a href="#monitoring.coreos.com/v1.AlertmanagerEndpoints">AlertmanagerEndpoints</a>, <a href="#monitoring.coreos.com/v1.Endpoint">Endpoint</a>, <a href="#monitoring.coreos.com/v1.PrometheusTracingConfig">PrometheusTracingConfig</a>, <a href="#monitoring.coreos.com/v1.RemoteReadSpec">RemoteReadSpec</a>, <a href="#monitoring.coreos.com/v1.RemoteWriteSpec">RemoteWriteSpec</a>, <a href="#monitoring.coreos.com/v1.ScrapeClass">ScrapeClass</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosSpec">ThanosSpec</a>)
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<em>(Optional)</em>
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
<em>(Optional)</em>
<p>Disable target certificate validation.</p>
</td>
</tr>
<tr>
<td>
<code>minVersion</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TLSVersion">
TLSVersion
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Minimum acceptable TLS version.</p>
<p>It requires Prometheus &gt;= v2.35.0 or Thanos &gt;= v0.28.0.</p>
</td>
</tr>
<tr>
<td>
<code>maxVersion</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.TLSVersion">
TLSVersion
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Maximum acceptable TLS version.</p>
<p>It requires Prometheus &gt;= v2.41.0 or Thanos &gt;= v0.31.0.</p>
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
<h3 id="monitoring.coreos.com/v1.TLSVersion">TLSVersion
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.SafeTLSConfig">SafeTLSConfig</a>)
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
<tbody><tr><td><p>&#34;TLS10&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;TLS11&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;TLS12&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;TLS13&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1.TSDBSpec">TSDBSpec
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
<code>outOfOrderTimeWindow</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures how old an out-of-order/out-of-bounds sample can be with
respect to the TSDB max time.</p>
<p>An out-of-order/out-of-bounds sample is ingested into the TSDB as long as
the timestamp of the sample is &gt;= (TSDB.MaxTime - outOfOrderTimeWindow).</p>
<p>This is an <em>experimental feature</em>, it may change in any upcoming release
in a breaking way.</p>
<p>It requires Prometheus &gt;= v2.39.0 or PrometheusAgent &gt;= v2.54.0.</p>
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
<em>(Optional)</em>
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
<em>(Optional)</em>
<p>PodMetadata configures labels and annotations which are propagated to the ThanosRuler pods.</p>
<p>The following items are reserved and cannot be overridden:
* &ldquo;app.kubernetes.io/name&rdquo; label, set to &ldquo;thanos-ruler&rdquo;.
* &ldquo;app.kubernetes.io/managed-by&rdquo; label, set to &ldquo;prometheus-operator&rdquo;.
* &ldquo;app.kubernetes.io/instance&rdquo; label, set to the name of the ThanosRuler instance.
* &ldquo;thanos-ruler&rdquo; label, set to the name of the ThanosRuler instance.
* &ldquo;kubectl.kubernetes.io/default-container&rdquo; annotation, set to &ldquo;thanos-ruler&rdquo;.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#pullpolicy-v1-core">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
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
<em>(Optional)</em>
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
<em>(Optional)</em>
<p>Define which Nodes the Pods are scheduled on.</p>
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
<p>Resources defines the resource requirements for single Pods.
If not provided, no requests/limits will be set</p>
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
<p>If specified, the pod&rsquo;s scheduling constraints.</p>
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
<p>If specified, the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s topology spread constraints.</p>
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
<a href="#monitoring.coreos.com/v1.DNSPolicy">
DNSPolicy
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
<a href="#monitoring.coreos.com/v1.PodDNSConfig">
PodDNSConfig
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
<code>serviceName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The name of the service name used by the underlying StatefulSet(s) as the governing service.
If defined, the Service  must be created before the ThanosRuler resource in the same namespace and it must define a selector that matches the pod labels.
If empty, the operator will create and manage a headless service named <code>thanos-ruler-operated</code> for ThanosRuler resources.
When deploying multiple ThanosRuler resources in the same namespace, it is recommended to specify a different value for each.
See <a href="https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id">https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id</a> for more details.</p>
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
<em>(Optional)</em>
<p>Storage spec to specify how storage shall be used.</p>
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
<em>(Optional)</em>
<p>Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will
be appended to other volumes that are generated as a result of StorageSpec objects.</p>
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
<em>(Optional)</em>
<p>VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
VolumeMounts specified will be appended to other VolumeMounts in the ruler container,
that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>objectStorageConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures object storage.</p>
<p>The configuration format is defined at <a href="https://thanos.io/tip/thanos/storage.md/#configuring-access-to-object-storage">https://thanos.io/tip/thanos/storage.md/#configuring-access-to-object-storage</a></p>
<p>The operator performs no validation of the configuration.</p>
<p><code>objectStorageConfigFile</code> takes precedence over this field.</p>
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
<p>Configures the path of the object storage configuration file.</p>
<p>The configuration format is defined at <a href="https://thanos.io/tip/thanos/storage.md/#configuring-access-to-object-storage">https://thanos.io/tip/thanos/storage.md/#configuring-access-to-object-storage</a></p>
<p>The operator performs no validation of the configuration file.</p>
<p>This field takes precedence over <code>objectStorageConfig</code>.</p>
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
<em>(Optional)</em>
<p>Configures the list of Thanos Query endpoints from which to query metrics.</p>
<p>For Thanos &gt;= v0.11.0, it is recommended to use <code>queryConfig</code> instead.</p>
<p><code>queryConfig</code> takes precedence over this field.</p>
</td>
</tr>
<tr>
<td>
<code>queryConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures the list of Thanos Query endpoints from which to query metrics.</p>
<p>The configuration format is defined at <a href="https://thanos.io/tip/components/rule.md/#query-api">https://thanos.io/tip/components/rule.md/#query-api</a></p>
<p>It requires Thanos &gt;= v0.11.0.</p>
<p>The operator performs no validation of the configuration.</p>
<p>This field takes precedence over <code>queryEndpoints</code>.</p>
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
<em>(Optional)</em>
<p>Configures the list of Alertmanager endpoints to send alerts to.</p>
<p>For Thanos &gt;= v0.10.0, it is recommended to use <code>alertmanagersConfig</code> instead.</p>
<p><code>alertmanagersConfig</code> takes precedence over this field.</p>
</td>
</tr>
<tr>
<td>
<code>alertmanagersConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures the list of Alertmanager endpoints to send alerts to.</p>
<p>The configuration format is defined at <a href="https://thanos.io/tip/components/rule.md/#alertmanager">https://thanos.io/tip/components/rule.md/#alertmanager</a>.</p>
<p>It requires Thanos &gt;= v0.10.0.</p>
<p>The operator performs no validation of the configuration.</p>
<p>This field takes precedence over <code>alertmanagersUrl</code>.</p>
</td>
</tr>
<tr>
<td>
<code>ruleSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
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
<em>(Optional)</em>
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
<em>(Optional)</em>
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
<p>Time duration ThanosRuler shall retain data for. Default is &lsquo;24h&rsquo;, and
must match the regular expression <code>[0-9]+(ms|s|m|h|d|w|y)</code> (milliseconds
seconds minutes hours days weeks years).</p>
<p>The field has no effect when remote-write is configured since the Ruler
operates in stateless mode.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<em>(Optional)</em>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures tracing.</p>
<p>The configuration format is defined at <a href="https://thanos.io/tip/thanos/tracing.md/#configuration">https://thanos.io/tip/thanos/tracing.md/#configuration</a></p>
<p>This is an <em>experimental feature</em>, it may change in any upcoming release
in a breaking way.</p>
<p>The operator performs no validation of the configuration.</p>
<p><code>tracingConfigFile</code> takes precedence over this field.</p>
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
<em>(Optional)</em>
<p>Configures the path of the tracing configuration file.</p>
<p>The configuration format is defined at <a href="https://thanos.io/tip/thanos/tracing.md/#configuration">https://thanos.io/tip/thanos/tracing.md/#configuration</a></p>
<p>This is an <em>experimental feature</em>, it may change in any upcoming release
in a breaking way.</p>
<p>The operator performs no validation of the configuration file.</p>
<p>This field takes precedence over <code>tracingConfig</code>.</p>
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
<p>Configures the external label pairs of the ThanosRuler resource.</p>
<p>A default replica label <code>thanos_ruler_replica</code> will be always added as a
label with the value of the pod&rsquo;s name.</p>
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
<em>(Optional)</em>
<p>Configures the label names which should be dropped in Thanos Ruler
alerts.</p>
<p>The replica label <code>thanos_ruler_replica</code> will always be dropped from the alerts.</p>
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
<em>(Optional)</em>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configures alert relabeling in Thanos Ruler.</p>
<p>Alert relabel configuration must have the form as specified in the
official Prometheus documentation:
<a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs</a></p>
<p>The operator performs no validation of the configuration.</p>
<p><code>alertRelabelConfigFile</code> takes precedence over this field.</p>
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
<em>(Optional)</em>
<p>Configures the path to the alert relabeling configuration file.</p>
<p>Alert relabel configuration must have the form as specified in the
official Prometheus documentation:
<a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs</a></p>
<p>The operator performs no validation of the configuration file.</p>
<p>This field takes precedence over <code>alertRelabelConfig</code>.</p>
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
<em>(Optional)</em>
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
<tr>
<td>
<code>web</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ThanosRulerWebSpec">
ThanosRulerWebSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the configuration of the ThanosRuler web server.</p>
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
<p>When the list isn&rsquo;t empty, the ruler is configured with stateless mode.</p>
<p>It requires Thanos &gt;= 0.24.0.</p>
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
<p>Defaults to 120 seconds.</p>
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
<p>The current state of the ThanosRuler object.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.ThanosRulerWebSpec">ThanosRulerWebSpec
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ThanosRulerSpec">ThanosRulerSpec</a>)
</p>
<div>
<p>ThanosRulerWebSpec defines the configuration of the ThanosRuler web server.</p>
</div>
<table>
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
<p>Deprecated: use &lsquo;image&rsquo; instead. The image&rsquo;s tag can be specified as as part of the image name.</p>
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
<p>Deprecated: use &lsquo;image&rsquo; instead.  The image digest can be specified as part of the image name.</p>
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
<p>Deprecated: use &lsquo;image&rsquo; instead.</p>
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
<p>Defines the resources requests and limits of the Thanos sidecar.</p>
</td>
</tr>
<tr>
<td>
<code>objectStorageConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
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
<p>Deprecated: use <code>grpcListenLocal</code> and <code>httpListenLocal</code> instead.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines the tracing configuration for the Thanos sidecar.</p>
<p><code>tracingConfigFile</code> takes precedence over this field.</p>
<p>More info: <a href="https://thanos.io/tip/thanos/tracing.md/">https://thanos.io/tip/thanos/tracing.md/</a></p>
<p>This is an <em>experimental feature</em>, it may change in any upcoming release
in a breaking way.</p>
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
<p>This field takes precedence over <code>tracingConfig</code>.</p>
<p>More info: <a href="https://thanos.io/tip/thanos/tracing.md/">https://thanos.io/tip/thanos/tracing.md/</a></p>
<p>This is an <em>experimental feature</em>, it may change in any upcoming release
in a breaking way.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volumemount-v1-core">
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
<h3 id="monitoring.coreos.com/v1.TopologySpreadConstraint">TopologySpreadConstraint
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
<code>maxSkew</code><br/>
<em>
int32
</em>
</td>
<td>
<p>MaxSkew describes the degree to which pods may be unevenly distributed.
When <code>whenUnsatisfiable=DoNotSchedule</code>, it is the maximum permitted difference
between the number of matching pods in the target topology and the global minimum.
The global minimum is the minimum number of matching pods in an eligible domain
or zero if the number of eligible domains is less than MinDomains.
For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same
labelSelector spread as 2/2/1:
In this case, the global minimum is 1.
| zone1 | zone2 | zone3 |
|  P P  |  P P  |   P   |
- if MaxSkew is 1, incoming pod can only be scheduled to zone3 to become 2/2/2;
scheduling it onto zone1(zone2) would make the ActualSkew(3-1) on zone1(zone2)
violate MaxSkew(1).
- if MaxSkew is 2, incoming pod can be scheduled onto any zone.
When <code>whenUnsatisfiable=ScheduleAnyway</code>, it is used to give higher precedence
to topologies that satisfy it.
It&rsquo;s a required field. Default value is 1 and 0 is not allowed.</p>
</td>
</tr>
<tr>
<td>
<code>topologyKey</code><br/>
<em>
string
</em>
</td>
<td>
<p>TopologyKey is the key of node labels. Nodes that have a label with this key
and identical values are considered to be in the same topology.
We consider each <key, value> as a &ldquo;bucket&rdquo;, and try to put balanced number
of pods into each bucket.
We define a domain as a particular instance of a topology.
Also, we define an eligible domain as a domain whose nodes meet the requirements of
nodeAffinityPolicy and nodeTaintsPolicy.
e.g. If TopologyKey is &ldquo;kubernetes.io/hostname&rdquo;, each Node is a domain of that topology.
And, if TopologyKey is &ldquo;topology.kubernetes.io/zone&rdquo;, each zone is a domain of that topology.
It&rsquo;s a required field.</p>
</td>
</tr>
<tr>
<td>
<code>whenUnsatisfiable</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#unsatisfiableconstraintaction-v1-core">
Kubernetes core/v1.UnsatisfiableConstraintAction
</a>
</em>
</td>
<td>
<p>WhenUnsatisfiable indicates how to deal with a pod if it doesn&rsquo;t satisfy
the spread constraint.
- DoNotSchedule (default) tells the scheduler not to schedule it.
- ScheduleAnyway tells the scheduler to schedule the pod in any location,
but giving higher precedence to topologies that would help reduce the
skew.
A constraint is considered &ldquo;Unsatisfiable&rdquo; for an incoming pod
if and only if every possible node assignment for that pod would violate
&ldquo;MaxSkew&rdquo; on some topology.
For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same
labelSelector spread as 3/1/1:
| zone1 | zone2 | zone3 |
| P P P |   P   |   P   |
If WhenUnsatisfiable is set to DoNotSchedule, incoming pod can only be scheduled
to zone2(zone3) to become 3/2/1(3/1/2) as ActualSkew(2-1) on zone2(zone3) satisfies
MaxSkew(1). In other words, the cluster can still be imbalanced, but scheduler
won&rsquo;t make it <em>more</em> imbalanced.
It&rsquo;s a required field.</p>
</td>
</tr>
<tr>
<td>
<code>labelSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>LabelSelector is used to find matching pods.
Pods that match this label selector are counted to determine the number of pods
in their corresponding topology domain.</p>
</td>
</tr>
<tr>
<td>
<code>minDomains</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>MinDomains indicates a minimum number of eligible domains.
When the number of eligible domains with matching topology keys is less than minDomains,
Pod Topology Spread treats &ldquo;global minimum&rdquo; as 0, and then the calculation of Skew is performed.
And when the number of eligible domains with matching topology keys equals or greater than minDomains,
this value has no effect on scheduling.
As a result, when the number of eligible domains is less than minDomains,
scheduler won&rsquo;t schedule more than maxSkew Pods to those domains.
If value is nil, the constraint behaves as if MinDomains is equal to 1.
Valid values are integers greater than 0.
When value is not nil, WhenUnsatisfiable must be DoNotSchedule.</p>
<p>For example, in a 3-zone cluster, MaxSkew is set to 2, MinDomains is set to 5 and pods with the same
labelSelector spread as 2/2/2:
| zone1 | zone2 | zone3 |
|  P P  |  P P  |  P P  |
The number of domains is less than 5(MinDomains), so &ldquo;global minimum&rdquo; is treated as 0.
In this situation, new pod with the same labelSelector cannot be scheduled,
because computed skew will be 3(3 - 0) if new Pod is scheduled to any of the three zones,
it will violate MaxSkew.</p>
</td>
</tr>
<tr>
<td>
<code>nodeAffinityPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#nodeinclusionpolicy-v1-core">
Kubernetes core/v1.NodeInclusionPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeAffinityPolicy indicates how we will treat Pod&rsquo;s nodeAffinity/nodeSelector
when calculating pod topology spread skew. Options are:
- Honor: only nodes matching nodeAffinity/nodeSelector are included in the calculations.
- Ignore: nodeAffinity/nodeSelector are ignored. All nodes are included in the calculations.</p>
<p>If this value is nil, the behavior is equivalent to the Honor policy.</p>
</td>
</tr>
<tr>
<td>
<code>nodeTaintsPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#nodeinclusionpolicy-v1-core">
Kubernetes core/v1.NodeInclusionPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeTaintsPolicy indicates how we will treat node taints when calculating
pod topology spread skew. Options are:
- Honor: nodes without taints, along with tainted nodes for which the incoming pod
has a toleration, are included.
- Ignore: node taints are ignored. All nodes are included.</p>
<p>If this value is nil, the behavior is equivalent to the Ignore policy.</p>
</td>
</tr>
<tr>
<td>
<code>matchLabelKeys</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>MatchLabelKeys is a set of pod label keys to select the pods over which
spreading will be calculated. The keys are used to lookup values from the
incoming pod labels, those key-value labels are ANDed with labelSelector
to select the group of existing pods over which spreading will be calculated
for the incoming pod. The same key is forbidden to exist in both MatchLabelKeys and LabelSelector.
MatchLabelKeys cannot be set when LabelSelector isn&rsquo;t set.
Keys that don&rsquo;t exist in the incoming pod labels will
be ignored. A null or empty list means only match against labelSelector.</p>
<p>This is a beta field and requires the MatchLabelKeysInPodTopologySpread feature gate to be enabled (enabled by default).</p>
</td>
</tr>
<tr>
<td>
<code>additionalLabelSelectors</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.AdditionalLabelSelectors">
AdditionalLabelSelectors
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defines what Prometheus Operator managed labels should be added to labelSelector on the topologySpreadConstraint.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.TranslationStrategyOption">TranslationStrategyOption
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.OTLPConfig">OTLPConfig</a>)
</p>
<div>
<p>TranslationStrategyOption represents a translation strategy option for the OTLP endpoint.
Supported values are:
* <code>NoUTF8EscapingWithSuffixes</code>
* <code>UnderscoreEscapingWithSuffixes</code>
* <code>NoTranslation</code></p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;NoTranslation&#34;</p></td>
<td><p>It requires Prometheus &gt;= v3.4.0.</p>
</td>
</tr><tr><td><p>&#34;NoUTF8EscapingWithSuffixes&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;UnderscoreEscapingWithSuffixes&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="monitoring.coreos.com/v1.WebConfigFileFields">WebConfigFileFields
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.AlertmanagerWebSpec">AlertmanagerWebSpec</a>, <a href="#monitoring.coreos.com/v1.PrometheusWebSpec">PrometheusWebSpec</a>, <a href="#monitoring.coreos.com/v1.ThanosRulerWebSpec">ThanosRulerWebSpec</a>)
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
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ClusterTLSConfig">ClusterTLSConfig</a>, <a href="#monitoring.coreos.com/v1.WebConfigFileFields">WebConfigFileFields</a>)
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
<code>cert</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.SecretOrConfigMap">
SecretOrConfigMap
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Secret or ConfigMap containing the TLS certificate for the web server.</p>
<p>Either <code>keySecret</code> or <code>keyFile</code> must be defined.</p>
<p>It is mutually exclusive with <code>certFile</code>.</p>
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
<em>(Optional)</em>
<p>Path to the TLS certificate file in the container for the web server.</p>
<p>Either <code>keySecret</code> or <code>keyFile</code> must be defined.</p>
<p>It is mutually exclusive with <code>cert</code>.</p>
</td>
</tr>
<tr>
<td>
<code>keySecret</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Secret containing the TLS private key for the web server.</p>
<p>Either <code>cert</code> or <code>certFile</code> must be defined.</p>
<p>It is mutually exclusive with <code>keyFile</code>.</p>
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
<em>(Optional)</em>
<p>Path to the TLS private key file in the container for the web server.</p>
<p>If defined, either <code>cert</code> or <code>certFile</code> must be defined.</p>
<p>It is mutually exclusive with <code>keySecret</code>.</p>
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
<em>(Optional)</em>
<p>Secret or ConfigMap containing the CA certificate for client certificate
authentication to the server.</p>
<p>It is mutually exclusive with <code>clientCAFile</code>.</p>
</td>
</tr>
<tr>
<td>
<code>clientCAFile</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Path to the CA certificate file for client certificate authentication to
the server.</p>
<p>It is mutually exclusive with <code>client_ca</code>.</p>
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
<em>(Optional)</em>
<p>The server policy for client TLS authentication.</p>
<p>For more detail on clientAuth options:
<a href="https://golang.org/pkg/crypto/tls/#ClientAuthType">https://golang.org/pkg/crypto/tls/#ClientAuthType</a></p>
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
<em>(Optional)</em>
<p>Minimum TLS version that is acceptable.</p>
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
<em>(Optional)</em>
<p>Maximum TLS version that is acceptable.</p>
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
<em>(Optional)</em>
<p>List of supported cipher suites for TLS versions up to TLS 1.2.</p>
<p>If not defined, the Go default cipher suites are used.
Available cipher suites are documented in the Go documentation:
<a href="https://golang.org/pkg/crypto/tls/#pkg-constants">https://golang.org/pkg/crypto/tls/#pkg-constants</a></p>
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
<em>(Optional)</em>
<p>Controls whether the server selects the client&rsquo;s most preferred cipher
suite, or the server&rsquo;s most preferred cipher suite.</p>
<p>If true then the server&rsquo;s preference, as expressed in
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
<em>(Optional)</em>
<p>Elliptic curves that will be used in an ECDHE handshake, in preference
order.</p>
<p>Available curves are documented in the Go documentation:
<a href="https://golang.org/pkg/crypto/tls/#CurveID">https://golang.org/pkg/crypto/tls/#CurveID</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.coreos.com/v1.WhenScaledRetentionType">WhenScaledRetentionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1.ShardRetentionPolicy">ShardRetentionPolicy</a>)
</p>
<div>
</div>
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
<a href="#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
EmbeddedObjectMetadata
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
<p>Number of seconds to wait until a scrape request times out.
The value cannot be greater than the scrape interval otherwise the operator will reject the resource.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeProtocols</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
[]ScrapeProtocol
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
<a href="#monitoring.coreos.com/v1.RemoteWriteMessageVersion">
[]RemoteWriteMessageVersion
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
<a href="#monitoring.coreos.com/v1.EnableFeature">
[]EnableFeature
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
<a href="#monitoring.coreos.com/v1.TopologySpreadConstraint">
[]TopologySpreadConstraint
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
<code>otlp</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OTLPConfig">
OTLPConfig
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
<a href="#monitoring.coreos.com/v1.DNSPolicy">
DNSPolicy
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
<a href="#monitoring.coreos.com/v1.PodDNSConfig">
PodDNSConfig
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
<a href="#monitoring.coreos.com/v1.NameValidationSchemeOptions">
NameValidationSchemeOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the validation scheme for metric and label names.</p>
</td>
</tr>
<tr>
<td>
<code>nameEscapingScheme</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.NameEscapingSchemeOptions">
NameEscapingSchemeOptions
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
<p>Whether to convert all scraped classic histograms into a native histogram with custom buckets.</p>
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
<a href="#monitoring.coreos.com/v1.PrometheusTracingConfig">
PrometheusTracingConfig
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
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
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
<a href="#monitoring.coreos.com/v1.ReloadStrategyType">
ReloadStrategyType
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
<a href="#monitoring.coreos.com/v1.ScrapeClass">
[]ScrapeClass
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
<a href="#monitoring.coreos.com/v1.ServiceDiscoveryRole">
ServiceDiscoveryRole
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
<a href="#monitoring.coreos.com/v1.TSDBSpec">
TSDBSpec
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
<a href="#monitoring.coreos.com/v1.RuntimeConfig">
RuntimeConfig
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
<code>scrapeInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
[]ScrapeProtocol
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
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
ScrapeProtocol
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
<code>oauth2</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.RelabelConfig">
[]RelabelConfig
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
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
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
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
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
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
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
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
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
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
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<p>Authorization header configuration to authenticate against the Docker API.
Cannot be set at the same time as <code>oauth2</code>.</p>
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
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
<code>oauth2</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
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
<code>oauth2</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
More info: <a href="https://prometheus.io/docs/operating/configuration/#endpoints">https://prometheus.io/docs/operating/configuration/#endpoints</a>
Cannot be set at the same time as <code>authorization</code>, or <code>oAuth2</code>.</p>
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
<p>Authorization header configuration to authenticate against the target HTTP endpoint.
Cannot be set at the same time as <code>oAuth2</code>, or <code>basicAuth</code>.</p>
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
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
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
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
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
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
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
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
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
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
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
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
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
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<code>oauth2</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
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
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
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
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
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
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<code>oauth2</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
EmbeddedObjectMetadata
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
<p>Number of seconds to wait until a scrape request times out.
The value cannot be greater than the scrape interval otherwise the operator will reject the resource.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeProtocols</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
[]ScrapeProtocol
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
<a href="#monitoring.coreos.com/v1.RemoteWriteMessageVersion">
[]RemoteWriteMessageVersion
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
<a href="#monitoring.coreos.com/v1.EnableFeature">
[]EnableFeature
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
<a href="#monitoring.coreos.com/v1.TopologySpreadConstraint">
[]TopologySpreadConstraint
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
<code>otlp</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OTLPConfig">
OTLPConfig
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
<a href="#monitoring.coreos.com/v1.DNSPolicy">
DNSPolicy
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
<a href="#monitoring.coreos.com/v1.PodDNSConfig">
PodDNSConfig
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
<a href="#monitoring.coreos.com/v1.NameValidationSchemeOptions">
NameValidationSchemeOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the validation scheme for metric and label names.</p>
</td>
</tr>
<tr>
<td>
<code>nameEscapingScheme</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.NameEscapingSchemeOptions">
NameEscapingSchemeOptions
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
<p>Whether to convert all scraped classic histograms into a native histogram with custom buckets.</p>
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
<a href="#monitoring.coreos.com/v1.PrometheusTracingConfig">
PrometheusTracingConfig
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
<a href="#monitoring.coreos.com/v1.ByteSize">
ByteSize
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
<a href="#monitoring.coreos.com/v1.ReloadStrategyType">
ReloadStrategyType
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
<a href="#monitoring.coreos.com/v1.ScrapeClass">
[]ScrapeClass
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
<a href="#monitoring.coreos.com/v1.ServiceDiscoveryRole">
ServiceDiscoveryRole
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
<a href="#monitoring.coreos.com/v1.TSDBSpec">
TSDBSpec
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
<a href="#monitoring.coreos.com/v1.RuntimeConfig">
RuntimeConfig
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.BasicAuth">
BasicAuth
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
<a href="#monitoring.coreos.com/v1.SafeAuthorization">
SafeAuthorization
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
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.SafeTLSConfig">
SafeTLSConfig
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<code>scrapeInterval</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
[]ScrapeProtocol
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
<a href="#monitoring.coreos.com/v1.ScrapeProtocol">
ScrapeProtocol
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
<code>oauth2</code><br/>
<em>
<a href="#monitoring.coreos.com/v1.OAuth2">
OAuth2
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
<a href="#monitoring.coreos.com/v1.RelabelConfig">
[]RelabelConfig
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<h2 id="monitoring.coreos.com/v1beta1">monitoring.coreos.com/v1beta1</h2>
Resource Types:
<ul><li>
<a href="#monitoring.coreos.com/v1beta1.AlertmanagerConfig">AlertmanagerConfig</a>
</li></ul>
<h3 id="monitoring.coreos.com/v1beta1.AlertmanagerConfig">AlertmanagerConfig
</h3>
<div>
<p>The <code>AlertmanagerConfig</code> custom resource definition (CRD) defines how <code>Alertmanager</code> objects process Prometheus alerts. It allows to specify alert grouping and routing, notification receivers and inhibition rules.</p>
<p><code>Alertmanager</code> objects select <code>AlertmanagerConfig</code> objects using label and namespace selectors.</p>
</div>
<table>
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
<h3 id="monitoring.coreos.com/v1beta1.DiscordConfig">DiscordConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Receiver">Receiver</a>)
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
<a href="#monitoring.coreos.com/v1beta1.URL">
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
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.DiscordConfig">DiscordConfig</a>, <a href="#monitoring.coreos.com/v1beta1.MSTeamsConfig">MSTeamsConfig</a>, <a href="#monitoring.coreos.com/v1beta1.MSTeamsV2Config">MSTeamsV2Config</a>, <a href="#monitoring.coreos.com/v1beta1.OpsGenieConfig">OpsGenieConfig</a>, <a href="#monitoring.coreos.com/v1beta1.PagerDutyConfig">PagerDutyConfig</a>, <a href="#monitoring.coreos.com/v1beta1.PushoverConfig">PushoverConfig</a>, <a href="#monitoring.coreos.com/v1beta1.SNSConfig">SNSConfig</a>, <a href="#monitoring.coreos.com/v1beta1.SlackConfig">SlackConfig</a>, <a href="#monitoring.coreos.com/v1beta1.TelegramConfig">TelegramConfig</a>, <a href="#monitoring.coreos.com/v1beta1.VictorOpsConfig">VictorOpsConfig</a>, <a href="#monitoring.coreos.com/v1beta1.WeChatConfig">WeChatConfig</a>, <a href="#monitoring.coreos.com/v1beta1.WebexConfig">WebexConfig</a>, <a href="#monitoring.coreos.com/v1beta1.WebhookConfig">WebhookConfig</a>)
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
<h3 id="monitoring.coreos.com/v1beta1.MSTeamsConfig">MSTeamsConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Receiver">Receiver</a>)
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
<h3 id="monitoring.coreos.com/v1beta1.MSTeamsV2Config">MSTeamsV2Config
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Receiver">Receiver</a>)
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
<a href="#monitoring.coreos.com/v1beta1.SecretKeySelector">
SecretKeySelector
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
<a href="#monitoring.coreos.com/v1.Duration">
Duration
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
<code>discordConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1beta1.DiscordConfig">
[]DiscordConfig
</a>
</em>
</td>
<td>
<p>List of Slack configurations.</p>
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
<tr>
<td>
<code>webexConfigs</code><br/>
<em>
<a href="#monitoring.coreos.com/v1beta1.WebexConfig">
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
<a href="#monitoring.coreos.com/v1beta1.MSTeamsConfig">
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
<a href="#monitoring.coreos.com/v1beta1.MSTeamsV2Config">
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
<h3 id="monitoring.coreos.com/v1beta1.URL">URL
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.DiscordConfig">DiscordConfig</a>, <a href="#monitoring.coreos.com/v1beta1.WebexConfig">WebexConfig</a>)
</p>
<div>
<p>URL represents a valid URL</p>
</div>
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
<h3 id="monitoring.coreos.com/v1beta1.WebexConfig">WebexConfig
</h3>
<p>
(<em>Appears on:</em><a href="#monitoring.coreos.com/v1beta1.Receiver">Receiver</a>)
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
<a href="#monitoring.coreos.com/v1beta1.URL">
URL
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The Webex Teams API URL i.e. <a href="https://webexapis.com/v1/messages">https://webexapis.com/v1/messages</a></p>
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
<p>The HTTP client&rsquo;s configuration.
You must use this configuration to supply the bot token as part of the HTTP <code>Authorization</code> header.</p>
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
<p>The maximum time to wait for a webhook request to complete, before failing the
request and allowing it to be retried.
It requires Alertmanager &gt;= v0.28.0.</p>
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
