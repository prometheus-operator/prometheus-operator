---
title: "Elastic API reference"
description: "Prometheus operator generated API reference docs"
draft: false
images: []
menu: "operator"
weight: 152
toc: true
---

# API Reference

## Packages
- [monitoring.coreos.com/v1](#monitoringcoreoscomv1)
- [monitoring.coreos.com/v1alpha1](#monitoringcoreoscomv1alpha1)
- [monitoring.coreos.com/v1beta1](#monitoringcoreoscomv1beta1)


## monitoring.coreos.com/v1


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.



#### APIServerConfig



APIServerConfig defines how the Prometheus server connects to the Kubernetes API server.

More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config



_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ | Kubernetes API address consisting of a hostname or IP address followed<br />by an optional port number. |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth configuration for the API server.<br />Cannot be set at the same time as `authorization`, `bearerToken`, or<br />`bearerTokenFile`. |  |  |
| `bearerTokenFile` _string_ | File to read bearer token for accessing apiserver.<br />Cannot be set at the same time as `basicAuth`, `authorization`, or `bearerToken`.<br />Deprecated: this will be removed in a future release. Prefer using `authorization`. |  |  |
| `tlsConfig` _[TLSConfig](#tlsconfig)_ | TLS Config to use for the API server. |  |  |
| `authorization` _[Authorization](#authorization)_ | Authorization section for the API server.<br />Cannot be set at the same time as `basicAuth`, `bearerToken`, or<br />`bearerTokenFile`. |  |  |
| `bearerToken` _string_ | *Warning: this field shouldn't be used because the token value appears<br />in clear-text. Prefer using `authorization`.*<br />Deprecated: this will be removed in a future release. |  |  |


#### AdditionalLabelSelectors

_Underlying type:_ _string_



_Validation:_
- Enum: [OnResource OnShard]

_Appears in:_
- [TopologySpreadConstraint](#topologyspreadconstraint)

| Field | Description |
| --- | --- |
| `OnResource` | Automatically add a label selector that will select all pods matching the same Prometheus/PrometheusAgent resource (irrespective of their shards).<br /> |
| `OnShard` | Automatically add a label selector that will select all pods matching the same shard.<br /> |


#### AlertingSpec



AlertingSpec defines parameters for alerting configuration of Prometheus servers.



_Appears in:_
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `alertmanagers` _[AlertmanagerEndpoints](#alertmanagerendpoints) array_ | Alertmanager endpoints where Prometheus should send alerts to. |  |  |


#### Alertmanager



The `Alertmanager` custom resource definition (CRD) defines a desired [Alertmanager](https://prometheus.io/docs/alerting) setup to run in a Kubernetes cluster. It allows to specify many options such as the number of replicas, persistent storage and many more.

For each `Alertmanager` resource, the Operator deploys a `StatefulSet` in the same namespace. When there are two or more configured replicas, the Operator runs the Alertmanager instances in high-availability mode.

The resource defines via label and namespace selectors which `AlertmanagerConfig` objects should be associated to the deployed Alertmanager instances.



_Appears in:_
- [AlertmanagerList](#alertmanagerlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[AlertmanagerSpec](#alertmanagerspec)_ | Specification of the desired behavior of the Alertmanager cluster. More info:<br />https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status |  |  |
| `status` _[AlertmanagerStatus](#alertmanagerstatus)_ | Most recent observed status of the Alertmanager cluster. Read-only.<br />More info:<br />https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status |  |  |


#### AlertmanagerAPIVersion

_Underlying type:_ _string_



_Validation:_
- Enum: [v1 V1 v2 V2]

_Appears in:_
- [AlertmanagerEndpoints](#alertmanagerendpoints)



#### AlertmanagerConfigMatcherStrategy







_Appears in:_
- [AlertmanagerSpec](#alertmanagerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[AlertmanagerConfigMatcherStrategyType](#alertmanagerconfigmatcherstrategytype)_ | AlertmanagerConfigMatcherStrategyType defines the strategy used by<br />AlertmanagerConfig objects to match alerts in the routes and inhibition<br />rules.<br />The default value is `OnNamespace`. | OnNamespace | Enum: [OnNamespace None] <br /> |


#### AlertmanagerConfigMatcherStrategyType

_Underlying type:_ _string_





_Appears in:_
- [AlertmanagerConfigMatcherStrategy](#alertmanagerconfigmatcherstrategy)

| Field | Description |
| --- | --- |
| `OnNamespace` | With `OnNamespace`, the route and inhibition rules of an<br />AlertmanagerConfig object only process alerts that have a `namespace`<br />label equal to the namespace of the object.<br /> |
| `None` | With `None`, the route and inhbition rules of an AlertmanagerConfig<br />object process all incoming alerts.<br /> |


#### AlertmanagerConfiguration



AlertmanagerConfiguration defines the Alertmanager configuration.



_Appears in:_
- [AlertmanagerSpec](#alertmanagerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | The name of the AlertmanagerConfig resource which is used to generate the Alertmanager configuration.<br />It must be defined in the same namespace as the Alertmanager object.<br />The operator will not enforce a `namespace` label for routes and inhibition rules. |  | MinLength: 1 <br /> |
| `global` _[AlertmanagerGlobalConfig](#alertmanagerglobalconfig)_ | Defines the global parameters of the Alertmanager configuration. |  |  |
| `templates` _[SecretOrConfigMap](#secretorconfigmap) array_ | Custom notification templates. |  |  |


#### AlertmanagerEndpoints



AlertmanagerEndpoints defines a selection of a single Endpoints object
containing Alertmanager IPs to fire alerts against.



_Appears in:_
- [AlertingSpec](#alertingspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `namespace` _string_ | Namespace of the Endpoints object.<br />If not set, the object will be discovered in the namespace of the<br />Prometheus object. |  | MinLength: 1 <br /> |
| `name` _string_ | Name of the Endpoints object in the namespace. |  | MinLength: 1 <br /> |
| `port` _[IntOrString](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#intorstring-intstr-util)_ | Port on which the Alertmanager API is exposed. |  |  |
| `scheme` _string_ | Scheme to use when firing alerts. |  |  |
| `pathPrefix` _string_ | Prefix for the HTTP path alerts are pushed to. |  |  |
| `tlsConfig` _[TLSConfig](#tlsconfig)_ | TLS Config to use for Alertmanager. |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth configuration for Alertmanager.<br />Cannot be set at the same time as `bearerTokenFile`, `authorization` or `sigv4`. |  |  |
| `bearerTokenFile` _string_ | File to read bearer token for Alertmanager.<br />Cannot be set at the same time as `basicAuth`, `authorization`, or `sigv4`.<br />Deprecated: this will be removed in a future release. Prefer using `authorization`. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization section for Alertmanager.<br />Cannot be set at the same time as `basicAuth`, `bearerTokenFile` or `sigv4`. |  |  |
| `sigv4` _[Sigv4](#sigv4)_ | Sigv4 allows to configures AWS's Signature Verification 4 for the URL.<br />It requires Prometheus >= v2.48.0.<br />Cannot be set at the same time as `basicAuth`, `bearerTokenFile` or `authorization`. |  |  |
| `apiVersion` _[AlertmanagerAPIVersion](#alertmanagerapiversion)_ | Version of the Alertmanager API that Prometheus uses to send alerts.<br />It can be "V1" or "V2".<br />The field has no effect for Prometheus >= v3.0.0 because only the v2 API is supported. |  | Enum: [v1 V1 v2 V2] <br /> |
| `timeout` _[Duration](#duration)_ | Timeout is a per-target Alertmanager timeout when pushing alerts. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `enableHttp2` _boolean_ | Whether to enable HTTP2. |  |  |
| `relabelings` _[RelabelConfig](#relabelconfig) array_ | Relabel configuration applied to the discovered Alertmanagers. |  |  |
| `alertRelabelings` _[RelabelConfig](#relabelconfig) array_ | Relabeling configs applied before sending alerts to a specific Alertmanager.<br />It requires Prometheus >= v2.51.0. |  |  |


#### AlertmanagerGlobalConfig



AlertmanagerGlobalConfig configures parameters that are valid in all other configuration contexts.
See https://prometheus.io/docs/alerting/latest/configuration/#configuration-file



_Appears in:_
- [AlertmanagerConfiguration](#alertmanagerconfiguration)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `smtp` _[GlobalSMTPConfig](#globalsmtpconfig)_ | Configures global SMTP parameters. |  |  |
| `resolveTimeout` _[Duration](#duration)_ | ResolveTimeout is the default value used by alertmanager if the alert does<br />not include EndsAt, after this time passes it can declare the alert as resolved if it has not been updated.<br />This has no impact on alerts from Prometheus, as they always include EndsAt. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |
| `slackApiUrl` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The default Slack API URL. |  |  |
| `opsGenieApiUrl` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The default OpsGenie API URL. |  |  |
| `opsGenieApiKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The default OpsGenie API Key. |  |  |
| `pagerdutyUrl` _string_ | The default Pagerduty URL. |  |  |


#### AlertmanagerLimitsSpec



AlertmanagerLimitsSpec defines the limits command line flags when starting Alertmanager.



_Appears in:_
- [AlertmanagerSpec](#alertmanagerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `maxSilences` _integer_ | The maximum number active and pending silences. This corresponds to the<br />Alertmanager's `--silences.max-silences` flag.<br />It requires Alertmanager >= v0.28.0. |  | Minimum: 0 <br /> |
| `maxPerSilenceBytes` _[ByteSize](#bytesize)_ | The maximum size of an individual silence as stored on disk. This corresponds to the Alertmanager's<br />`--silences.max-per-silence-bytes` flag.<br />It requires Alertmanager >= v0.28.0. |  | Pattern: `(^0\|([0-9]*[.])?[0-9]+((K\|M\|G\|T\|E\|P)i?)?B)$` <br /> |




#### AlertmanagerSpec



AlertmanagerSpec is a specification of the desired behavior of the Alertmanager cluster. More info:
https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status



_Appears in:_
- [Alertmanager](#alertmanager)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `podMetadata` _[EmbeddedObjectMetadata](#embeddedobjectmetadata)_ | PodMetadata configures labels and annotations which are propagated to the Alertmanager pods.<br />The following items are reserved and cannot be overridden:<br />* "alertmanager" label, set to the name of the Alertmanager instance.<br />* "app.kubernetes.io/instance" label, set to the name of the Alertmanager instance.<br />* "app.kubernetes.io/managed-by" label, set to "prometheus-operator".<br />* "app.kubernetes.io/name" label, set to "alertmanager".<br />* "app.kubernetes.io/version" label, set to the Alertmanager version.<br />* "kubectl.kubernetes.io/default-container" annotation, set to "alertmanager". |  |  |
| `image` _string_ | Image if specified has precedence over baseImage, tag and sha<br />combinations. Specifying the version is still necessary to ensure the<br />Prometheus Operator knows what version of Alertmanager is being<br />configured. |  |  |
| `imagePullPolicy` _[PullPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#pullpolicy-v1-core)_ | Image pull policy for the 'alertmanager', 'init-config-reloader' and 'config-reloader' containers.<br />See https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy for more details. |  | Enum: [ Always Never IfNotPresent] <br /> |
| `version` _string_ | Version the cluster should be on. |  |  |
| `tag` _string_ | Tag of Alertmanager container image to be deployed. Defaults to the value of `version`.<br />Version is ignored if Tag is set.<br />Deprecated: use 'image' instead. The image tag can be specified as part of the image URL. |  |  |
| `sha` _string_ | SHA of Alertmanager container image to be deployed. Defaults to the value of `version`.<br />Similar to a tag, but the SHA explicitly deploys an immutable container image.<br />Version and Tag are ignored if SHA is set.<br />Deprecated: use 'image' instead. The image digest can be specified as part of the image URL. |  |  |
| `baseImage` _string_ | Base image that is used to deploy pods, without tag.<br />Deprecated: use 'image' instead. |  |  |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#localobjectreference-v1-core) array_ | An optional list of references to secrets in the same namespace<br />to use for pulling prometheus and alertmanager images from registries<br />see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod |  |  |
| `secrets` _string array_ | Secrets is a list of Secrets in the same namespace as the Alertmanager<br />object, which shall be mounted into the Alertmanager Pods.<br />Each Secret is added to the StatefulSet definition as a volume named `secret-<secret-name>`.<br />The Secrets are mounted into `/etc/alertmanager/secrets/<secret-name>` in the 'alertmanager' container. |  |  |
| `configMaps` _string array_ | ConfigMaps is a list of ConfigMaps in the same namespace as the Alertmanager<br />object, which shall be mounted into the Alertmanager Pods.<br />Each ConfigMap is added to the StatefulSet definition as a volume named `configmap-<configmap-name>`.<br />The ConfigMaps are mounted into `/etc/alertmanager/configmaps/<configmap-name>` in the 'alertmanager' container. |  |  |
| `configSecret` _string_ | ConfigSecret is the name of a Kubernetes Secret in the same namespace as the<br />Alertmanager object, which contains the configuration for this Alertmanager<br />instance. If empty, it defaults to `alertmanager-<alertmanager-name>`.<br />The Alertmanager configuration should be available under the<br />`alertmanager.yaml` key. Additional keys from the original secret are<br />copied to the generated secret and mounted into the<br />`/etc/alertmanager/config` directory in the `alertmanager` container.<br />If either the secret or the `alertmanager.yaml` key is missing, the<br />operator provisions a minimal Alertmanager configuration with one empty<br />receiver (effectively dropping alert notifications). |  |  |
| `logLevel` _string_ | Log level for Alertmanager to be configured with. |  | Enum: [ debug info warn error] <br /> |
| `logFormat` _string_ | Log format for Alertmanager to be configured with. |  | Enum: [ logfmt json] <br /> |
| `replicas` _integer_ | Size is the expected size of the alertmanager cluster. The controller will<br />eventually make the size of the running cluster equal to the expected<br />size. |  |  |
| `retention` _[GoDuration](#goduration)_ | Time duration Alertmanager shall retain data for. Default is '120h',<br />and must match the regular expression `[0-9]+(ms\|s\|m\|h)` (milliseconds seconds minutes hours). | 120h | Pattern: `^(0\|(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `storage` _[StorageSpec](#storagespec)_ | Storage is the definition of how storage will be used by the Alertmanager<br />instances. |  |  |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volume-v1-core) array_ | Volumes allows configuration of additional volumes on the output StatefulSet definition.<br />Volumes specified will be appended to other volumes that are generated as a result of<br />StorageSpec objects. |  |  |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volumemount-v1-core) array_ | VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.<br />VolumeMounts specified will be appended to other VolumeMounts in the alertmanager container,<br />that are generated as a result of StorageSpec objects. |  |  |
| `persistentVolumeClaimRetentionPolicy` _[StatefulSetPersistentVolumeClaimRetentionPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#statefulsetpersistentvolumeclaimretentionpolicy-v1-apps)_ | The field controls if and how PVCs are deleted during the lifecycle of a StatefulSet.<br />The default behavior is all PVCs are retained.<br />This is an alpha field from kubernetes 1.23 until 1.26 and a beta field from 1.26.<br />It requires enabling the StatefulSetAutoDeletePVC feature gate. |  |  |
| `externalUrl` _string_ | The external URL the Alertmanager instances will be available under. This is<br />necessary to generate correct URLs. This is necessary if Alertmanager is not<br />served from root of a DNS name. |  |  |
| `routePrefix` _string_ | The route prefix Alertmanager registers HTTP handlers for. This is useful,<br />if using ExternalURL and a proxy is rewriting HTTP routes of a request,<br />and the actual ExternalURL is still true, but the server serves requests<br />under a different route prefix. For example for use with `kubectl proxy`. |  |  |
| `paused` _boolean_ | If set to true all actions on the underlying managed objects are not<br />goint to be performed, except for delete actions. |  |  |
| `nodeSelector` _object (keys:string, values:string)_ | Define which Nodes the Pods are scheduled on. |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#resourcerequirements-v1-core)_ | Define resources requests and limits for single Pods. |  |  |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#affinity-v1-core)_ | If specified, the pod's scheduling constraints. |  |  |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#toleration-v1-core) array_ | If specified, the pod's tolerations. |  |  |
| `topologySpreadConstraints` _[TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#topologyspreadconstraint-v1-core) array_ | If specified, the pod's topology spread constraints. |  |  |
| `securityContext` _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#podsecuritycontext-v1-core)_ | SecurityContext holds pod-level security attributes and common container settings.<br />This defaults to the default PodSecurityContext. |  |  |
| `dnsPolicy` _[DNSPolicy](#dnspolicy)_ | Defines the DNS policy for the pods. |  | Enum: [ClusterFirstWithHostNet ClusterFirst Default None] <br /> |
| `dnsConfig` _[PodDNSConfig](#poddnsconfig)_ | Defines the DNS configuration for the pods. |  |  |
| `enableServiceLinks` _boolean_ | Indicates whether information about services should be injected into pod's environment variables |  |  |
| `serviceName` _string_ | The name of the service name used by the underlying StatefulSet(s) as the governing service.<br />If defined, the Service  must be created before the Alertmanager resource in the same namespace and it must define a selector that matches the pod labels.<br />If empty, the operator will create and manage a headless service named `alertmanager-operated` for Alermanager resources.<br />When deploying multiple Alertmanager resources in the same namespace, it is recommended to specify a different value for each.<br />See https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id for more details. |  | MinLength: 1 <br /> |
| `serviceAccountName` _string_ | ServiceAccountName is the name of the ServiceAccount to use to run the<br />Prometheus Pods. |  |  |
| `listenLocal` _boolean_ | ListenLocal makes the Alertmanager server listen on loopback, so that it<br />does not bind against the Pod IP. Note this is only for the Alertmanager<br />UI, not the gossip communication. |  |  |
| `containers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core) array_ | Containers allows injecting additional containers. This is meant to<br />allow adding an authentication proxy to an Alertmanager pod.<br />Containers described here modify an operator generated container if they<br />share the same name and modifications are done via a strategic merge<br />patch. The current container names are: `alertmanager` and<br />`config-reloader`. Overriding containers is entirely outside the scope<br />of what the maintainers will support and by doing so, you accept that<br />this behaviour may break at any time without notice. |  |  |
| `initContainers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core) array_ | InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.<br />fetch secrets for injection into the Alertmanager configuration from external sources. Any<br />errors during the execution of an initContainer will lead to a restart of the Pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/<br />InitContainers described here modify an operator<br />generated init containers if they share the same name and modifications are<br />done via a strategic merge patch. The current init container name is:<br />`init-config-reloader`. Overriding init containers is entirely outside the<br />scope of what the maintainers will support and by doing so, you accept that<br />this behaviour may break at any time without notice. |  |  |
| `priorityClassName` _string_ | Priority class assigned to the Pods |  |  |
| `additionalPeers` _string array_ | AdditionalPeers allows injecting a set of additional Alertmanagers to peer with to form a highly available cluster. |  |  |
| `clusterAdvertiseAddress` _string_ | ClusterAdvertiseAddress is the explicit address to advertise in cluster.<br />Needs to be provided for non RFC1918 [1] (public) addresses.<br />[1] RFC1918: https://tools.ietf.org/html/rfc1918 |  |  |
| `clusterGossipInterval` _[GoDuration](#goduration)_ | Interval between gossip attempts. |  | Pattern: `^(0\|(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `clusterLabel` _string_ | Defines the identifier that uniquely identifies the Alertmanager cluster.<br />You should only set it when the Alertmanager cluster includes Alertmanager instances which are external to this Alertmanager resource. In practice, the addresses of the external instances are provided via the `.spec.additionalPeers` field. |  |  |
| `clusterPushpullInterval` _[GoDuration](#goduration)_ | Interval between pushpull attempts. |  | Pattern: `^(0\|(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `clusterPeerTimeout` _[GoDuration](#goduration)_ | Timeout for cluster peering. |  | Pattern: `^(0\|(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `portName` _string_ | Port name used for the pods and governing service.<br />Defaults to `web`. | web |  |
| `forceEnableClusterMode` _boolean_ | ForceEnableClusterMode ensures Alertmanager does not deactivate the cluster mode when running with a single replica.<br />Use case is e.g. spanning an Alertmanager cluster across Kubernetes clusters with a single replica in each. |  |  |
| `alertmanagerConfigSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | AlertmanagerConfigs to be selected for to merge and configure Alertmanager with. |  |  |
| `alertmanagerConfigNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to be selected for AlertmanagerConfig discovery. If nil, only<br />check own namespace. |  |  |
| `alertmanagerConfigMatcherStrategy` _[AlertmanagerConfigMatcherStrategy](#alertmanagerconfigmatcherstrategy)_ | AlertmanagerConfigMatcherStrategy defines how AlertmanagerConfig objects<br />process incoming alerts. |  |  |
| `minReadySeconds` _integer_ | Minimum number of seconds for which a newly created pod should be ready<br />without any of its container crashing for it to be considered available.<br />Defaults to 0 (pod will be considered available as soon as it is ready)<br />This is an alpha field from kubernetes 1.22 until 1.24 which requires enabling the StatefulSetMinReadySeconds feature gate. |  |  |
| `hostAliases` _[HostAlias](#hostalias) array_ | Pods' hostAliases configuration |  |  |
| `web` _[AlertmanagerWebSpec](#alertmanagerwebspec)_ | Defines the web command line flags when starting Alertmanager. |  |  |
| `limits` _[AlertmanagerLimitsSpec](#alertmanagerlimitsspec)_ | Defines the limits command line flags when starting Alertmanager. |  |  |
| `clusterTLS` _[ClusterTLSConfig](#clustertlsconfig)_ | Configures the mutual TLS configuration for the Alertmanager cluster's gossip protocol.<br />It requires Alertmanager >= 0.24.0. |  |  |
| `alertmanagerConfiguration` _[AlertmanagerConfiguration](#alertmanagerconfiguration)_ | alertmanagerConfiguration specifies the configuration of Alertmanager.<br />If defined, it takes precedence over the `configSecret` field.<br />This is an *experimental feature*, it may change in any upcoming release<br />in a breaking way. |  |  |
| `automountServiceAccountToken` _boolean_ | AutomountServiceAccountToken indicates whether a service account token should be automatically mounted in the pod.<br />If the service account has `automountServiceAccountToken: true`, set the field to `false` to opt out of automounting API credentials. |  |  |
| `enableFeatures` _string array_ | Enable access to Alertmanager feature flags. By default, no features are enabled.<br />Enabling features which are disabled by default is entirely outside the<br />scope of what the maintainers will support and by doing so, you accept<br />that this behaviour may break at any time without notice.<br />It requires Alertmanager >= 0.27.0. |  |  |
| `additionalArgs` _[Argument](#argument) array_ | AdditionalArgs allows setting additional arguments for the 'Alertmanager' container.<br />It is intended for e.g. activating hidden flags which are not supported by<br />the dedicated configuration options yet. The arguments are passed as-is to the<br />Alertmanager container which may cause issues if they are invalid or not supported<br />by the given Alertmanager version. |  |  |
| `terminationGracePeriodSeconds` _integer_ | Optional duration in seconds the pod needs to terminate gracefully.<br />Value must be non-negative integer. The value zero indicates stop immediately via<br />the kill signal (no opportunity to shut down) which may lead to data corruption.<br />Defaults to 120 seconds. |  | Minimum: 0 <br /> |


#### AlertmanagerStatus



AlertmanagerStatus is the most recent observed status of the Alertmanager cluster. Read-only.
More info:
https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status



_Appears in:_
- [Alertmanager](#alertmanager)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `paused` _boolean_ | Represents whether any actions on the underlying managed objects are<br />being performed. Only delete actions will be performed. |  |  |
| `replicas` _integer_ | Total number of non-terminated pods targeted by this Alertmanager<br />object (their labels match the selector). |  |  |
| `updatedReplicas` _integer_ | Total number of non-terminated pods targeted by this Alertmanager<br />object that have the desired version spec. |  |  |
| `availableReplicas` _integer_ | Total number of available pods (ready for at least minReadySeconds)<br />targeted by this Alertmanager cluster. |  |  |
| `unavailableReplicas` _integer_ | Total number of unavailable pods targeted by this Alertmanager object. |  |  |
| `selector` _string_ | The selector used to match the pods targeted by this Alertmanager object. |  |  |
| `conditions` _[Condition](#condition) array_ | The current state of the Alertmanager object. |  |  |


#### AlertmanagerWebSpec



AlertmanagerWebSpec defines the web command line flags when starting Alertmanager.



_Appears in:_
- [AlertmanagerSpec](#alertmanagerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `tlsConfig` _[WebTLSConfig](#webtlsconfig)_ | Defines the TLS parameters for HTTPS. |  |  |
| `httpConfig` _[WebHTTPConfig](#webhttpconfig)_ | Defines HTTP parameters for web server. |  |  |
| `getConcurrency` _integer_ | Maximum number of GET requests processed concurrently. This corresponds to the<br />Alertmanager's `--web.get-concurrency` flag. |  |  |
| `timeout` _integer_ | Timeout for HTTP requests. This corresponds to the Alertmanager's<br />`--web.timeout` flag. |  |  |


#### ArbitraryFSAccessThroughSMsConfig



ArbitraryFSAccessThroughSMsConfig enables users to configure, whether
a service monitor selected by the Prometheus instance is allowed to use
arbitrary files on the file system of the Prometheus container. This is the case
when e.g. a service monitor specifies a BearerTokenFile in an endpoint. A
malicious user could create a service monitor selecting arbitrary secret files
in the Prometheus container. Those secrets would then be sent with a scrape
request by Prometheus to a malicious target. Denying the above would prevent the
attack, users can instead use the BearerTokenSecret field.



_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `deny` _boolean_ |  |  |  |


#### Argument



Argument as part of the AdditionalArgs list.



_Appears in:_
- [AlertmanagerSpec](#alertmanagerspec)
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)
- [ThanosRulerSpec](#thanosrulerspec)
- [ThanosSpec](#thanosspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the argument, e.g. "scrape.discovery-reload-interval". |  | MinLength: 1 <br /> |
| `value` _string_ | Argument value, e.g. 30s. Can be empty for name-only arguments (e.g. --storage.tsdb.no-lockfile) |  |  |


#### AttachMetadata







_Appears in:_
- [PodMonitorSpec](#podmonitorspec)
- [ScrapeClass](#scrapeclass)
- [ServiceMonitorSpec](#servicemonitorspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `node` _boolean_ | When set to true, Prometheus attaches node metadata to the discovered<br />targets.<br />The Prometheus service account must have the `list` and `watch`<br />permissions on the `Nodes` objects. |  |  |


#### Authorization







_Appears in:_
- [APIServerConfig](#apiserverconfig)
- [RemoteReadSpec](#remotereadspec)
- [RemoteWriteSpec](#remotewritespec)
- [ScrapeClass](#scrapeclass)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _string_ | Defines the authentication type. The value is case-insensitive.<br />"Basic" is not a supported value.<br />Default: "Bearer" |  |  |
| `credentials` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Selects a key of a Secret in the namespace that contains the credentials for authentication. |  |  |
| `credentialsFile` _string_ | File to read a secret from, mutually exclusive with `credentials`. |  |  |




#### AzureAD



AzureAD defines the configuration for remote write's azuread parameters.



_Appears in:_
- [RemoteWriteSpec](#remotewritespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `cloud` _string_ | The Azure Cloud. Options are 'AzurePublic', 'AzureChina', or 'AzureGovernment'. |  | Enum: [AzureChina AzureGovernment AzurePublic] <br /> |
| `managedIdentity` _[ManagedIdentity](#managedidentity)_ | ManagedIdentity defines the Azure User-assigned Managed identity.<br />Cannot be set at the same time as `oauth` or `sdk`. |  |  |
| `oauth` _[AzureOAuth](#azureoauth)_ | OAuth defines the oauth config that is being used to authenticate.<br />Cannot be set at the same time as `managedIdentity` or `sdk`.<br />It requires Prometheus >= v2.48.0 or Thanos >= v0.31.0. |  |  |
| `sdk` _[AzureSDK](#azuresdk)_ | SDK defines the Azure SDK config that is being used to authenticate.<br />See https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication<br />Cannot be set at the same time as `oauth` or `managedIdentity`.<br />It requires Prometheus >= v2.52.0 or Thanos >= v0.36.0. |  |  |


#### AzureOAuth



AzureOAuth defines the Azure OAuth settings.



_Appears in:_
- [AzureAD](#azuread)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `clientId` _string_ | `clientID` is the clientId of the Azure Active Directory application that is being used to authenticate. |  | MinLength: 1 <br /> |
| `clientSecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | `clientSecret` specifies a key of a Secret containing the client secret of the Azure Active Directory application that is being used to authenticate. |  |  |
| `tenantId` _string_ | `tenantId` is the tenant ID of the Azure Active Directory application that is being used to authenticate. |  | MinLength: 1 <br />Pattern: `^[0-9a-zA-Z-.]+$` <br /> |


#### AzureSDK



AzureSDK is used to store azure SDK config values.



_Appears in:_
- [AzureAD](#azuread)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `tenantId` _string_ | `tenantId` is the tenant ID of the azure active directory application that is being used to authenticate. |  | Pattern: `^[0-9a-zA-Z-.]+$` <br /> |


#### BasicAuth



BasicAuth configures HTTP Basic Authentication settings.



_Appears in:_
- [APIServerConfig](#apiserverconfig)
- [AlertmanagerEndpoints](#alertmanagerendpoints)
- [AzureSDConfig](#azuresdconfig)
- [ConsulSDConfig](#consulsdconfig)
- [DockerSDConfig](#dockersdconfig)
- [DockerSwarmSDConfig](#dockerswarmsdconfig)
- [Endpoint](#endpoint)
- [EurekaSDConfig](#eurekasdconfig)
- [HTTPConfig](#httpconfig)
- [HTTPConfig](#httpconfig)
- [HTTPConfig](#httpconfig)
- [HTTPSDConfig](#httpsdconfig)
- [HetznerSDConfig](#hetznersdconfig)
- [KubernetesSDConfig](#kubernetessdconfig)
- [KumaSDConfig](#kumasdconfig)
- [LightSailSDConfig](#lightsailsdconfig)
- [NomadSDConfig](#nomadsdconfig)
- [PodMetricsEndpoint](#podmetricsendpoint)
- [ProbeSpec](#probespec)
- [PuppetDBSDConfig](#puppetdbsdconfig)
- [RemoteReadSpec](#remotereadspec)
- [RemoteWriteSpec](#remotewritespec)
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `username` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | `username` specifies a key of a Secret containing the username for<br />authentication. |  |  |
| `password` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | `password` specifies a key of a Secret containing the password for<br />authentication. |  |  |


#### ByteSize

_Underlying type:_ _string_

ByteSize is a valid memory size type based on powers-of-2, so 1KB is 1024B.
Supported units: B, KB, KiB, MB, MiB, GB, GiB, TB, TiB, PB, PiB, EB, EiB Ex: `512MB`.

_Validation:_
- Pattern: `(^0|([0-9]*[.])?[0-9]+((K|M|G|T|E|P)i?)?B)$`

_Appears in:_
- [AlertmanagerLimitsSpec](#alertmanagerlimitsspec)
- [CommonPrometheusFields](#commonprometheusfields)
- [PodMonitorSpec](#podmonitorspec)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)
- [ServiceMonitorSpec](#servicemonitorspec)



#### ClusterTLSConfig



ClusterTLSConfig defines the mutual TLS configuration for the Alertmanager cluster TLS protocol.



_Appears in:_
- [AlertmanagerSpec](#alertmanagerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `server` _[WebTLSConfig](#webtlsconfig)_ | Server-side configuration for mutual TLS. |  |  |
| `client` _[SafeTLSConfig](#safetlsconfig)_ | Client-side configuration for mutual TLS. |  |  |


#### CommonPrometheusFields



CommonPrometheusFields are the options available to both the Prometheus server and agent.



_Appears in:_
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `podMetadata` _[EmbeddedObjectMetadata](#embeddedobjectmetadata)_ | PodMetadata configures labels and annotations which are propagated to the Prometheus pods.<br />The following items are reserved and cannot be overridden:<br />* "prometheus" label, set to the name of the Prometheus object.<br />* "app.kubernetes.io/instance" label, set to the name of the Prometheus object.<br />* "app.kubernetes.io/managed-by" label, set to "prometheus-operator".<br />* "app.kubernetes.io/name" label, set to "prometheus".<br />* "app.kubernetes.io/version" label, set to the Prometheus version.<br />* "operator.prometheus.io/name" label, set to the name of the Prometheus object.<br />* "operator.prometheus.io/shard" label, set to the shard number of the Prometheus object.<br />* "kubectl.kubernetes.io/default-container" annotation, set to "prometheus". |  |  |
| `serviceMonitorSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | ServiceMonitors to be selected for target discovery. An empty label<br />selector matches all objects. A null label selector matches no objects.<br />If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`<br />and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.<br />The Prometheus operator will ensure that the Prometheus configuration's<br />Secret exists, but it is the responsibility of the user to provide the raw<br />gzipped Prometheus configuration under the `prometheus.yaml.gz` key.<br />This behavior is *deprecated* and will be removed in the next major version<br />of the custom resource definition. It is recommended to use<br />`spec.additionalScrapeConfigs` instead. |  |  |
| `serviceMonitorNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to match for ServicedMonitors discovery. An empty label selector<br />matches all namespaces. A null label selector (default value) matches the current<br />namespace only. |  |  |
| `podMonitorSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | PodMonitors to be selected for target discovery. An empty label selector<br />matches all objects. A null label selector matches no objects.<br />If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`<br />and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.<br />The Prometheus operator will ensure that the Prometheus configuration's<br />Secret exists, but it is the responsibility of the user to provide the raw<br />gzipped Prometheus configuration under the `prometheus.yaml.gz` key.<br />This behavior is *deprecated* and will be removed in the next major version<br />of the custom resource definition. It is recommended to use<br />`spec.additionalScrapeConfigs` instead. |  |  |
| `podMonitorNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to match for PodMonitors discovery. An empty label selector<br />matches all namespaces. A null label selector (default value) matches the current<br />namespace only. |  |  |
| `probeSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Probes to be selected for target discovery. An empty label selector<br />matches all objects. A null label selector matches no objects.<br />If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`<br />and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.<br />The Prometheus operator will ensure that the Prometheus configuration's<br />Secret exists, but it is the responsibility of the user to provide the raw<br />gzipped Prometheus configuration under the `prometheus.yaml.gz` key.<br />This behavior is *deprecated* and will be removed in the next major version<br />of the custom resource definition. It is recommended to use<br />`spec.additionalScrapeConfigs` instead. |  |  |
| `probeNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to match for Probe discovery. An empty label<br />selector matches all namespaces. A null label selector matches the<br />current namespace only. |  |  |
| `scrapeConfigSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | ScrapeConfigs to be selected for target discovery. An empty label<br />selector matches all objects. A null label selector matches no objects.<br />If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`<br />and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.<br />The Prometheus operator will ensure that the Prometheus configuration's<br />Secret exists, but it is the responsibility of the user to provide the raw<br />gzipped Prometheus configuration under the `prometheus.yaml.gz` key.<br />This behavior is *deprecated* and will be removed in the next major version<br />of the custom resource definition. It is recommended to use<br />`spec.additionalScrapeConfigs` instead.<br />Note that the ScrapeConfig custom resource definition is currently at Alpha level. |  |  |
| `scrapeConfigNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to match for ScrapeConfig discovery. An empty label selector<br />matches all namespaces. A null label selector matches the current<br />namespace only.<br />Note that the ScrapeConfig custom resource definition is currently at Alpha level. |  |  |
| `version` _string_ | Version of Prometheus being deployed. The operator uses this information<br />to generate the Prometheus StatefulSet + configuration files.<br />If not specified, the operator assumes the latest upstream version of<br />Prometheus available at the time when the version of the operator was<br />released. |  |  |
| `paused` _boolean_ | When a Prometheus deployment is paused, no actions except for deletion<br />will be performed on the underlying objects. |  |  |
| `image` _string_ | Container image name for Prometheus. If specified, it takes precedence<br />over the `spec.baseImage`, `spec.tag` and `spec.sha` fields.<br />Specifying `spec.version` is still necessary to ensure the Prometheus<br />Operator knows which version of Prometheus is being configured.<br />If neither `spec.image` nor `spec.baseImage` are defined, the operator<br />will use the latest upstream version of Prometheus available at the time<br />when the operator was released. |  |  |
| `imagePullPolicy` _[PullPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#pullpolicy-v1-core)_ | Image pull policy for the 'prometheus', 'init-config-reloader' and 'config-reloader' containers.<br />See https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy for more details. |  | Enum: [ Always Never IfNotPresent] <br /> |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#localobjectreference-v1-core) array_ | An optional list of references to Secrets in the same namespace<br />to use for pulling images from registries.<br />See http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod |  |  |
| `replicas` _integer_ | Number of replicas of each shard to deploy for a Prometheus deployment.<br />`spec.replicas` multiplied by `spec.shards` is the total number of Pods<br />created.<br />Default: 1 |  |  |
| `shards` _integer_ | Number of shards to distribute the scraped targets onto.<br />`spec.replicas` multiplied by `spec.shards` is the total number of Pods<br />being created.<br />When not defined, the operator assumes only one shard.<br />Note that scaling down shards will not reshard data onto the remaining<br />instances, it must be manually moved. Increasing shards will not reshard<br />data either but it will continue to be available from the same<br />instances. To query globally, use either<br />* Thanos sidecar + querier for query federation and Thanos Ruler for rules.<br />* Remote-write to send metrics to a central location.<br />By default, the sharding of targets is performed on:<br />* The `__address__` target's metadata label for PodMonitor,<br />ServiceMonitor and ScrapeConfig resources.<br />* The `__param_target__` label for Probe resources.<br />Users can define their own sharding implementation by setting the<br />`__tmp_hash` label during the target discovery with relabeling<br />configuration (either in the monitoring resources or via scrape class).<br />You can also disable sharding on a specific target by setting the<br />`__tmp_disable_sharding` label with relabeling configuration. When<br />the label value isn't empty, all Prometheus shards will scrape the target. |  |  |
| `replicaExternalLabelName` _string_ | Name of Prometheus external label used to denote the replica name.<br />The external label will _not_ be added when the field is set to the<br />empty string (`""`).<br />Default: "prometheus_replica" |  |  |
| `prometheusExternalLabelName` _string_ | Name of Prometheus external label used to denote the Prometheus instance<br />name. The external label will _not_ be added when the field is set to<br />the empty string (`""`).<br />Default: "prometheus" |  |  |
| `logLevel` _string_ | Log level for Prometheus and the config-reloader sidecar. |  | Enum: [ debug info warn error] <br /> |
| `logFormat` _string_ | Log format for Log level for Prometheus and the config-reloader sidecar. |  | Enum: [ logfmt json] <br /> |
| `scrapeInterval` _[Duration](#duration)_ | Interval between consecutive scrapes.<br />Default: "30s" | 30s | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `scrapeTimeout` _[Duration](#duration)_ | Number of seconds to wait until a scrape request times out.<br />The value cannot be greater than the scrape interval otherwise the operator will reject the resource. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `scrapeProtocols` _[ScrapeProtocol](#scrapeprotocol) array_ | The protocols to negotiate during a scrape. It tells clients the<br />protocols supported by Prometheus in order of preference (from most to least preferred).<br />If unset, Prometheus uses its default value.<br />It requires Prometheus >= v2.49.0.<br />`PrometheusText1.0.0` requires Prometheus >= v3.0.0. |  | Enum: [PrometheusProto OpenMetricsText0.0.1 OpenMetricsText1.0.0 PrometheusText0.0.4 PrometheusText1.0.0] <br /> |
| `externalLabels` _object (keys:string, values:string)_ | The labels to add to any time series or alerts when communicating with<br />external systems (federation, remote storage, Alertmanager).<br />Labels defined by `spec.replicaExternalLabelName` and<br />`spec.prometheusExternalLabelName` take precedence over this list. |  |  |
| `enableRemoteWriteReceiver` _boolean_ | Enable Prometheus to be used as a receiver for the Prometheus remote<br />write protocol.<br />WARNING: This is not considered an efficient way of ingesting samples.<br />Use it with caution for specific low-volume use cases.<br />It is not suitable for replacing the ingestion via scraping and turning<br />Prometheus into a push-based metrics collection system.<br />For more information see https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver<br />It requires Prometheus >= v2.33.0. |  |  |
| `enableOTLPReceiver` _boolean_ | Enable Prometheus to be used as a receiver for the OTLP Metrics protocol.<br />Note that the OTLP receiver endpoint is automatically enabled if `.spec.otlpConfig` is defined.<br />It requires Prometheus >= v2.47.0. |  |  |
| `remoteWriteReceiverMessageVersions` _[RemoteWriteMessageVersion](#remotewritemessageversion) array_ | List of the protobuf message versions to accept when receiving the<br />remote writes.<br />It requires Prometheus >= v2.54.0. |  | Enum: [V1.0 V2.0] <br />MinItems: 1 <br /> |
| `enableFeatures` _[EnableFeature](#enablefeature) array_ | Enable access to Prometheus feature flags. By default, no features are enabled.<br />Enabling features which are disabled by default is entirely outside the<br />scope of what the maintainers will support and by doing so, you accept<br />that this behaviour may break at any time without notice.<br />For more information see https://prometheus.io/docs/prometheus/latest/feature_flags/ |  | MinLength: 1 <br /> |
| `externalUrl` _string_ | The external URL under which the Prometheus service is externally<br />available. This is necessary to generate correct URLs (for instance if<br />Prometheus is accessible behind an Ingress resource). |  |  |
| `routePrefix` _string_ | The route prefix Prometheus registers HTTP handlers for.<br />This is useful when using `spec.externalURL`, and a proxy is rewriting<br />HTTP routes of a request, and the actual ExternalURL is still true, but<br />the server serves requests under a different route prefix. For example<br />for use with `kubectl proxy`. |  |  |
| `storage` _[StorageSpec](#storagespec)_ | Storage defines the storage used by Prometheus. |  |  |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volume-v1-core) array_ | Volumes allows the configuration of additional volumes on the output<br />StatefulSet definition. Volumes specified will be appended to other<br />volumes that are generated as a result of StorageSpec objects. |  |  |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volumemount-v1-core) array_ | VolumeMounts allows the configuration of additional VolumeMounts.<br />VolumeMounts will be appended to other VolumeMounts in the 'prometheus'<br />container, that are generated as a result of StorageSpec objects. |  |  |
| `persistentVolumeClaimRetentionPolicy` _[StatefulSetPersistentVolumeClaimRetentionPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#statefulsetpersistentvolumeclaimretentionpolicy-v1-apps)_ | The field controls if and how PVCs are deleted during the lifecycle of a StatefulSet.<br />The default behavior is all PVCs are retained.<br />This is an alpha field from kubernetes 1.23 until 1.26 and a beta field from 1.26.<br />It requires enabling the StatefulSetAutoDeletePVC feature gate. |  |  |
| `web` _[PrometheusWebSpec](#prometheuswebspec)_ | Defines the configuration of the Prometheus web server. |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#resourcerequirements-v1-core)_ | Defines the resources requests and limits of the 'prometheus' container. |  |  |
| `nodeSelector` _object (keys:string, values:string)_ | Defines on which Nodes the Pods are scheduled. |  |  |
| `serviceAccountName` _string_ | ServiceAccountName is the name of the ServiceAccount to use to run the<br />Prometheus Pods. |  |  |
| `automountServiceAccountToken` _boolean_ | AutomountServiceAccountToken indicates whether a service account token should be automatically mounted in the pod.<br />If the field isn't set, the operator mounts the service account token by default.<br />**Warning:** be aware that by default, Prometheus requires the service account token for Kubernetes service discovery.<br />It is possible to use strategic merge patch to project the service account token into the 'prometheus' container. |  |  |
| `secrets` _string array_ | Secrets is a list of Secrets in the same namespace as the Prometheus<br />object, which shall be mounted into the Prometheus Pods.<br />Each Secret is added to the StatefulSet definition as a volume named `secret-<secret-name>`.<br />The Secrets are mounted into /etc/prometheus/secrets/<secret-name> in the 'prometheus' container. |  |  |
| `configMaps` _string array_ | ConfigMaps is a list of ConfigMaps in the same namespace as the Prometheus<br />object, which shall be mounted into the Prometheus Pods.<br />Each ConfigMap is added to the StatefulSet definition as a volume named `configmap-<configmap-name>`.<br />The ConfigMaps are mounted into /etc/prometheus/configmaps/<configmap-name> in the 'prometheus' container. |  |  |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#affinity-v1-core)_ | Defines the Pods' affinity scheduling rules if specified. |  |  |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#toleration-v1-core) array_ | Defines the Pods' tolerations if specified. |  |  |
| `topologySpreadConstraints` _[TopologySpreadConstraint](#topologyspreadconstraint) array_ | Defines the pod's topology spread constraints if specified. |  |  |
| `remoteWrite` _[RemoteWriteSpec](#remotewritespec) array_ | Defines the list of remote write configurations. |  |  |
| `otlp` _[OTLPConfig](#otlpconfig)_ | Settings related to the OTLP receiver feature.<br />It requires Prometheus >= v2.55.0. |  |  |
| `securityContext` _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#podsecuritycontext-v1-core)_ | SecurityContext holds pod-level security attributes and common container settings.<br />This defaults to the default PodSecurityContext. |  |  |
| `dnsPolicy` _[DNSPolicy](#dnspolicy)_ | Defines the DNS policy for the pods. |  | Enum: [ClusterFirstWithHostNet ClusterFirst Default None] <br /> |
| `dnsConfig` _[PodDNSConfig](#poddnsconfig)_ | Defines the DNS configuration for the pods. |  |  |
| `listenLocal` _boolean_ | When true, the Prometheus server listens on the loopback address<br />instead of the Pod IP's address. |  |  |
| `enableServiceLinks` _boolean_ | Indicates whether information about services should be injected into pod's environment variables |  |  |
| `containers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core) array_ | Containers allows injecting additional containers or modifying operator<br />generated containers. This can be used to allow adding an authentication<br />proxy to the Pods or to change the behavior of an operator generated<br />container. Containers described here modify an operator generated<br />container if they share the same name and modifications are done via a<br />strategic merge patch.<br />The names of containers managed by the operator are:<br />* `prometheus`<br />* `config-reloader`<br />* `thanos-sidecar`<br />Overriding containers is entirely outside the scope of what the<br />maintainers will support and by doing so, you accept that this behaviour<br />may break at any time without notice. |  |  |
| `initContainers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core) array_ | InitContainers allows injecting initContainers to the Pod definition. Those<br />can be used to e.g.  fetch secrets for injection into the Prometheus<br />configuration from external sources. Any errors during the execution of<br />an initContainer will lead to a restart of the Pod. More info:<br />https://kubernetes.io/docs/concepts/workloads/pods/init-containers/<br />InitContainers described here modify an operator generated init<br />containers if they share the same name and modifications are done via a<br />strategic merge patch.<br />The names of init container name managed by the operator are:<br />* `init-config-reloader`.<br />Overriding init containers is entirely outside the scope of what the<br />maintainers will support and by doing so, you accept that this behaviour<br />may break at any time without notice. |  |  |
| `additionalScrapeConfigs` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | AdditionalScrapeConfigs allows specifying a key of a Secret containing<br />additional Prometheus scrape configurations. Scrape configurations<br />specified are appended to the configurations generated by the Prometheus<br />Operator. Job configurations specified must have the form as specified<br />in the official Prometheus documentation:<br />https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config.<br />As scrape configs are appended, the user is responsible to make sure it<br />is valid. Note that using this feature may expose the possibility to<br />break upgrades of Prometheus. It is advised to review Prometheus release<br />notes to ensure that no incompatible scrape configs are going to break<br />Prometheus after the upgrade. |  |  |
| `apiserverConfig` _[APIServerConfig](#apiserverconfig)_ | APIServerConfig allows specifying a host and auth methods to access the<br />Kuberntees API server.<br />If null, Prometheus is assumed to run inside of the cluster: it will<br />discover the API servers automatically and use the Pod's CA certificate<br />and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/. |  |  |
| `priorityClassName` _string_ | Priority class assigned to the Pods. |  |  |
| `portName` _string_ | Port name used for the pods and governing service.<br />Default: "web" | web |  |
| `arbitraryFSAccessThroughSMs` _[ArbitraryFSAccessThroughSMsConfig](#arbitraryfsaccessthroughsmsconfig)_ | When true, ServiceMonitor, PodMonitor and Probe object are forbidden to<br />reference arbitrary files on the file system of the 'prometheus'<br />container.<br />When a ServiceMonitor's endpoint specifies a `bearerTokenFile` value<br />(e.g.  '/var/run/secrets/kubernetes.io/serviceaccount/token'), a<br />malicious target can get access to the Prometheus service account's<br />token in the Prometheus' scrape request. Setting<br />`spec.arbitraryFSAccessThroughSM` to 'true' would prevent the attack.<br />Users should instead provide the credentials using the<br />`spec.bearerTokenSecret` field. |  |  |
| `overrideHonorLabels` _boolean_ | When true, Prometheus resolves label conflicts by renaming the labels in the scraped data<br /> to “exported_” for all targets created from ServiceMonitor, PodMonitor and<br />ScrapeConfig objects. Otherwise the HonorLabels field of the service or pod monitor applies.<br />In practice,`overrideHonorLaels:true` enforces `honorLabels:false`<br />for all ServiceMonitor, PodMonitor and ScrapeConfig objects. |  |  |
| `overrideHonorTimestamps` _boolean_ | When true, Prometheus ignores the timestamps for all the targets created<br />from service and pod monitors.<br />Otherwise the HonorTimestamps field of the service or pod monitor applies. |  |  |
| `ignoreNamespaceSelectors` _boolean_ | When true, `spec.namespaceSelector` from all PodMonitor, ServiceMonitor<br />and Probe objects will be ignored. They will only discover targets<br />within the namespace of the PodMonitor, ServiceMonitor and Probe<br />object. |  |  |
| `enforcedNamespaceLabel` _string_ | When not empty, a label will be added to:<br />1. All metrics scraped from `ServiceMonitor`, `PodMonitor`, `Probe` and `ScrapeConfig` objects.<br />2. All metrics generated from recording rules defined in `PrometheusRule` objects.<br />3. All alerts generated from alerting rules defined in `PrometheusRule` objects.<br />4. All vector selectors of PromQL expressions defined in `PrometheusRule` objects.<br />The label will not added for objects referenced in `spec.excludedFromEnforcement`.<br />The label's name is this field's value.<br />The label's value is the namespace of the `ServiceMonitor`,<br />`PodMonitor`, `Probe`, `PrometheusRule` or `ScrapeConfig` object. |  |  |
| `enforcedSampleLimit` _integer_ | When defined, enforcedSampleLimit specifies a global limit on the number<br />of scraped samples that will be accepted. This overrides any<br />`spec.sampleLimit` set by ServiceMonitor, PodMonitor, Probe objects<br />unless `spec.sampleLimit` is greater than zero and less than<br />`spec.enforcedSampleLimit`.<br />It is meant to be used by admins to keep the overall number of<br />samples/series under a desired limit.<br />When both `enforcedSampleLimit` and `sampleLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined sampleLimit value will inherit the global sampleLimit value (Prometheus >= 2.45.0) or the enforcedSampleLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedSampleLimit` is greater than the `sampleLimit`, the `sampleLimit` will be set to `enforcedSampleLimit`.<br />* Scrape objects with a sampleLimit value less than or equal to enforcedSampleLimit keep their specific value.<br />* Scrape objects with a sampleLimit value greater than enforcedSampleLimit are set to enforcedSampleLimit. |  |  |
| `enforcedTargetLimit` _integer_ | When defined, enforcedTargetLimit specifies a global limit on the number<br />of scraped targets. The value overrides any `spec.targetLimit` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.targetLimit` is<br />greater than zero and less than `spec.enforcedTargetLimit`.<br />It is meant to be used by admins to to keep the overall number of<br />targets under a desired limit.<br />When both `enforcedTargetLimit` and `targetLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined targetLimit value will inherit the global targetLimit value (Prometheus >= 2.45.0) or the enforcedTargetLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedTargetLimit` is greater than the `targetLimit`, the `targetLimit` will be set to `enforcedTargetLimit`.<br />* Scrape objects with a targetLimit value less than or equal to enforcedTargetLimit keep their specific value.<br />* Scrape objects with a targetLimit value greater than enforcedTargetLimit are set to enforcedTargetLimit. |  |  |
| `enforcedLabelLimit` _integer_ | When defined, enforcedLabelLimit specifies a global limit on the number<br />of labels per sample. The value overrides any `spec.labelLimit` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.labelLimit` is<br />greater than zero and less than `spec.enforcedLabelLimit`.<br />It requires Prometheus >= v2.27.0.<br />When both `enforcedLabelLimit` and `labelLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined labelLimit value will inherit the global labelLimit value (Prometheus >= 2.45.0) or the enforcedLabelLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedLabelLimit` is greater than the `labelLimit`, the `labelLimit` will be set to `enforcedLabelLimit`.<br />* Scrape objects with a labelLimit value less than or equal to enforcedLabelLimit keep their specific value.<br />* Scrape objects with a labelLimit value greater than enforcedLabelLimit are set to enforcedLabelLimit. |  |  |
| `enforcedLabelNameLengthLimit` _integer_ | When defined, enforcedLabelNameLengthLimit specifies a global limit on the length<br />of labels name per sample. The value overrides any `spec.labelNameLengthLimit` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.labelNameLengthLimit` is<br />greater than zero and less than `spec.enforcedLabelNameLengthLimit`.<br />It requires Prometheus >= v2.27.0.<br />When both `enforcedLabelNameLengthLimit` and `labelNameLengthLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined labelNameLengthLimit value will inherit the global labelNameLengthLimit value (Prometheus >= 2.45.0) or the enforcedLabelNameLengthLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedLabelNameLengthLimit` is greater than the `labelNameLengthLimit`, the `labelNameLengthLimit` will be set to `enforcedLabelNameLengthLimit`.<br />* Scrape objects with a labelNameLengthLimit value less than or equal to enforcedLabelNameLengthLimit keep their specific value.<br />* Scrape objects with a labelNameLengthLimit value greater than enforcedLabelNameLengthLimit are set to enforcedLabelNameLengthLimit. |  |  |
| `enforcedLabelValueLengthLimit` _integer_ | When not null, enforcedLabelValueLengthLimit defines a global limit on the length<br />of labels value per sample. The value overrides any `spec.labelValueLengthLimit` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.labelValueLengthLimit` is<br />greater than zero and less than `spec.enforcedLabelValueLengthLimit`.<br />It requires Prometheus >= v2.27.0.<br />When both `enforcedLabelValueLengthLimit` and `labelValueLengthLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined labelValueLengthLimit value will inherit the global labelValueLengthLimit value (Prometheus >= 2.45.0) or the enforcedLabelValueLengthLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedLabelValueLengthLimit` is greater than the `labelValueLengthLimit`, the `labelValueLengthLimit` will be set to `enforcedLabelValueLengthLimit`.<br />* Scrape objects with a labelValueLengthLimit value less than or equal to enforcedLabelValueLengthLimit keep their specific value.<br />* Scrape objects with a labelValueLengthLimit value greater than enforcedLabelValueLengthLimit are set to enforcedLabelValueLengthLimit. |  |  |
| `enforcedKeepDroppedTargets` _integer_ | When defined, enforcedKeepDroppedTargets specifies a global limit on the number of targets<br />dropped by relabeling that will be kept in memory. The value overrides<br />any `spec.keepDroppedTargets` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.keepDroppedTargets` is<br />greater than zero and less than `spec.enforcedKeepDroppedTargets`.<br />It requires Prometheus >= v2.47.0.<br />When both `enforcedKeepDroppedTargets` and `keepDroppedTargets` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined keepDroppedTargets value will inherit the global keepDroppedTargets value (Prometheus >= 2.45.0) or the enforcedKeepDroppedTargets value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedKeepDroppedTargets` is greater than the `keepDroppedTargets`, the `keepDroppedTargets` will be set to `enforcedKeepDroppedTargets`.<br />* Scrape objects with a keepDroppedTargets value less than or equal to enforcedKeepDroppedTargets keep their specific value.<br />* Scrape objects with a keepDroppedTargets value greater than enforcedKeepDroppedTargets are set to enforcedKeepDroppedTargets. |  |  |
| `enforcedBodySizeLimit` _[ByteSize](#bytesize)_ | When defined, enforcedBodySizeLimit specifies a global limit on the size<br />of uncompressed response body that will be accepted by Prometheus.<br />Targets responding with a body larger than this many bytes will cause<br />the scrape to fail.<br />It requires Prometheus >= v2.28.0.<br />When both `enforcedBodySizeLimit` and `bodySizeLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined bodySizeLimit value will inherit the global bodySizeLimit value (Prometheus >= 2.45.0) or the enforcedBodySizeLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedBodySizeLimit` is greater than the `bodySizeLimit`, the `bodySizeLimit` will be set to `enforcedBodySizeLimit`.<br />* Scrape objects with a bodySizeLimit value less than or equal to enforcedBodySizeLimit keep their specific value.<br />* Scrape objects with a bodySizeLimit value greater than enforcedBodySizeLimit are set to enforcedBodySizeLimit. |  | Pattern: `(^0\|([0-9]*[.])?[0-9]+((K\|M\|G\|T\|E\|P)i?)?B)$` <br /> |
| `nameValidationScheme` _[NameValidationSchemeOptions](#namevalidationschemeoptions)_ | Specifies the validation scheme for metric and label names.<br />It requires Prometheus >= v2.55.0. |  | Enum: [UTF8 Legacy] <br /> |
| `nameEscapingScheme` _[NameEscapingSchemeOptions](#nameescapingschemeoptions)_ | Specifies the character escaping scheme that will be requested when scraping<br />for metric and label names that do not conform to the legacy Prometheus<br />character set.<br />It requires Prometheus >= v3.4.0. |  | Enum: [AllowUTF8 Underscores Dots Values] <br /> |
| `convertClassicHistogramsToNHCB` _boolean_ | Whether to convert all scraped classic histograms into a native<br />histogram with custom buckets.<br />It requires Prometheus >= v3.4.0. |  |  |
| `minReadySeconds` _integer_ | Minimum number of seconds for which a newly created Pod should be ready<br />without any of its container crashing for it to be considered available.<br />Defaults to 0 (pod will be considered available as soon as it is ready)<br />This is an alpha field from kubernetes 1.22 until 1.24 which requires<br />enabling the StatefulSetMinReadySeconds feature gate. |  |  |
| `hostAliases` _[HostAlias](#hostalias) array_ | Optional list of hosts and IPs that will be injected into the Pod's<br />hosts file if specified. |  |  |
| `additionalArgs` _[Argument](#argument) array_ | AdditionalArgs allows setting additional arguments for the 'prometheus' container.<br />It is intended for e.g. activating hidden flags which are not supported by<br />the dedicated configuration options yet. The arguments are passed as-is to the<br />Prometheus container which may cause issues if they are invalid or not supported<br />by the given Prometheus version.<br />In case of an argument conflict (e.g. an argument which is already set by the<br />operator itself) or when providing an invalid argument, the reconciliation will<br />fail and an error will be logged. |  |  |
| `walCompression` _boolean_ | Configures compression of the write-ahead log (WAL) using Snappy.<br />WAL compression is enabled by default for Prometheus >= 2.20.0<br />Requires Prometheus v2.11.0 and above. |  |  |
| `excludedFromEnforcement` _[ObjectReference](#objectreference) array_ | List of references to PodMonitor, ServiceMonitor, Probe and PrometheusRule objects<br />to be excluded from enforcing a namespace label of origin.<br />It is only applicable if `spec.enforcedNamespaceLabel` set to true. |  |  |
| `hostNetwork` _boolean_ | Use the host's network namespace if true.<br />Make sure to understand the security implications if you want to enable<br />it (https://kubernetes.io/docs/concepts/configuration/overview/).<br />When hostNetwork is enabled, this will set the DNS policy to<br />`ClusterFirstWithHostNet` automatically (unless `.spec.DNSPolicy` is set<br />to a different value). |  |  |
| `podTargetLabels` _string array_ | PodTargetLabels are appended to the `spec.podTargetLabels` field of all<br />PodMonitor and ServiceMonitor objects. |  |  |
| `tracingConfig` _[PrometheusTracingConfig](#prometheustracingconfig)_ | TracingConfig configures tracing in Prometheus.<br />This is an *experimental feature*, it may change in any upcoming release<br />in a breaking way. |  |  |
| `bodySizeLimit` _[ByteSize](#bytesize)_ | BodySizeLimit defines per-scrape on response body size.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedBodySizeLimit. |  | Pattern: `(^0\|([0-9]*[.])?[0-9]+((K\|M\|G\|T\|E\|P)i?)?B)$` <br /> |
| `sampleLimit` _integer_ | SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedSampleLimit. |  |  |
| `targetLimit` _integer_ | TargetLimit defines a limit on the number of scraped targets that will be accepted.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedTargetLimit. |  |  |
| `labelLimit` _integer_ | Per-scrape limit on number of labels that will be accepted for a sample.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelLimit. |  |  |
| `labelNameLengthLimit` _integer_ | Per-scrape limit on length of labels name that will be accepted for a sample.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelNameLengthLimit. |  |  |
| `labelValueLengthLimit` _integer_ | Per-scrape limit on length of labels value that will be accepted for a sample.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelValueLengthLimit. |  |  |
| `keepDroppedTargets` _integer_ | Per-scrape limit on the number of targets dropped by relabeling<br />that will be kept in memory. 0 means no limit.<br />It requires Prometheus >= v2.47.0.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedKeepDroppedTargets. |  |  |
| `reloadStrategy` _[ReloadStrategyType](#reloadstrategytype)_ | Defines the strategy used to reload the Prometheus configuration.<br />If not specified, the configuration is reloaded using the /-/reload HTTP endpoint. |  | Enum: [HTTP ProcessSignal] <br /> |
| `maximumStartupDurationSeconds` _integer_ | Defines the maximum time that the `prometheus` container's startup probe will wait before being considered failed. The startup probe will return success after the WAL replay is complete.<br />If set, the value should be greater than 60 (seconds). Otherwise it will be equal to 600 seconds (15 minutes). |  | Minimum: 60 <br /> |
| `scrapeClasses` _[ScrapeClass](#scrapeclass) array_ | List of scrape classes to expose to scraping objects such as<br />PodMonitors, ServiceMonitors, Probes and ScrapeConfigs.<br />This is an *experimental feature*, it may change in any upcoming release<br />in a breaking way. |  |  |
| `serviceDiscoveryRole` _[ServiceDiscoveryRole](#servicediscoveryrole)_ | Defines the service discovery role used to discover targets from<br />`ServiceMonitor` objects and Alertmanager endpoints.<br />If set, the value should be either "Endpoints" or "EndpointSlice".<br />If unset, the operator assumes the "Endpoints" role. |  | Enum: [Endpoints EndpointSlice] <br /> |
| `tsdb` _[TSDBSpec](#tsdbspec)_ | Defines the runtime reloadable configuration of the timeseries database(TSDB).<br />It requires Prometheus >= v2.39.0 or PrometheusAgent >= v2.54.0. |  |  |
| `scrapeFailureLogFile` _string_ | File to which scrape failures are logged.<br />Reloading the configuration will reopen the file.<br />If the filename has an empty path, e.g. 'file.log', The Prometheus Pods<br />will mount the file into an emptyDir volume at `/var/log/prometheus`.<br />If a full path is provided, e.g. '/var/log/prometheus/file.log', you<br />must mount a volume in the specified directory and it must be writable.<br />It requires Prometheus >= v2.55.0. |  | MinLength: 1 <br /> |
| `serviceName` _string_ | The name of the service name used by the underlying StatefulSet(s) as the governing service.<br />If defined, the Service  must be created before the Prometheus/PrometheusAgent resource in the same namespace and it must define a selector that matches the pod labels.<br />If empty, the operator will create and manage a headless service named `prometheus-operated` for Prometheus resources,<br />or `prometheus-agent-operated` for PrometheusAgent resources.<br />When deploying multiple Prometheus/PrometheusAgent resources in the same namespace, it is recommended to specify a different value for each.<br />See https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id for more details. |  | MinLength: 1 <br /> |
| `runtime` _[RuntimeConfig](#runtimeconfig)_ | RuntimeConfig configures the values for the Prometheus process behavior |  |  |
| `terminationGracePeriodSeconds` _integer_ | Optional duration in seconds the pod needs to terminate gracefully.<br />Value must be non-negative integer. The value zero indicates stop immediately via<br />the kill signal (no opportunity to shut down) which may lead to data corruption.<br />Defaults to 600 seconds. |  | Minimum: 0 <br /> |


#### Condition



Condition represents the state of the resources associated with the
Prometheus, Alertmanager or ThanosRuler resource.



_Appears in:_
- [AlertmanagerStatus](#alertmanagerstatus)
- [PrometheusStatus](#prometheusstatus)
- [ThanosRulerStatus](#thanosrulerstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[ConditionType](#conditiontype)_ | Type of the condition being reported. |  | MinLength: 1 <br /> |
| `status` _[ConditionStatus](#conditionstatus)_ | Status of the condition. |  | MinLength: 1 <br /> |
| `lastTransitionTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#time-v1-meta)_ | lastTransitionTime is the time of the last update to the current status property. |  |  |
| `reason` _string_ | Reason for the condition's last transition. |  |  |
| `message` _string_ | Human-readable message indicating details for the condition's last transition. |  |  |
| `observedGeneration` _integer_ | ObservedGeneration represents the .metadata.generation that the<br />condition was set based upon. For instance, if `.metadata.generation` is<br />currently 12, but the `.status.conditions[].observedGeneration` is 9, the<br />condition is out of date with respect to the current state of the<br />instance. |  |  |


#### ConditionStatus

_Underlying type:_ _string_



_Validation:_
- MinLength: 1

_Appears in:_
- [Condition](#condition)

| Field | Description |
| --- | --- |
| `True` |  |
| `Degraded` |  |
| `False` |  |
| `Unknown` |  |


#### ConditionType

_Underlying type:_ _string_



_Validation:_
- MinLength: 1

_Appears in:_
- [Condition](#condition)

| Field | Description |
| --- | --- |
| `Available` | Available indicates whether enough pods are ready to provide the<br />service.<br />The possible status values for this condition type are:<br />- True: all pods are running and ready, the service is fully available.<br />- Degraded: some pods aren't ready, the service is partially available.<br />- False: no pods are running, the service is totally unavailable.<br />- Unknown: the operator couldn't determine the condition status.<br /> |
| `Reconciled` | Reconciled indicates whether the operator has reconciled the state of<br />the underlying resources with the object's spec.<br />The possible status values for this condition type are:<br />- True: the reconciliation was successful.<br />- False: the reconciliation failed.<br />- Unknown: the operator couldn't determine the condition status.<br /> |


#### CoreV1TopologySpreadConstraint

_Underlying type:_ _[TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#topologyspreadconstraint-v1-core)_





_Appears in:_
- [TopologySpreadConstraint](#topologyspreadconstraint)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `maxSkew` _integer_ | MaxSkew describes the degree to which pods may be unevenly distributed.<br />When `whenUnsatisfiable=DoNotSchedule`, it is the maximum permitted difference<br />between the number of matching pods in the target topology and the global minimum.<br />The global minimum is the minimum number of matching pods in an eligible domain<br />or zero if the number of eligible domains is less than MinDomains.<br />For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same<br />labelSelector spread as 2/2/1:<br />In this case, the global minimum is 1.<br />\| zone1 \| zone2 \| zone3 \|<br />\|  P P  \|  P P  \|   P   \|<br />- if MaxSkew is 1, incoming pod can only be scheduled to zone3 to become 2/2/2;<br />scheduling it onto zone1(zone2) would make the ActualSkew(3-1) on zone1(zone2)<br />violate MaxSkew(1).<br />- if MaxSkew is 2, incoming pod can be scheduled onto any zone.<br />When `whenUnsatisfiable=ScheduleAnyway`, it is used to give higher precedence<br />to topologies that satisfy it.<br />It's a required field. Default value is 1 and 0 is not allowed. |  |  |
| `topologyKey` _string_ | TopologyKey is the key of node labels. Nodes that have a label with this key<br />and identical values are considered to be in the same topology.<br />We consider each <key, value> as a "bucket", and try to put balanced number<br />of pods into each bucket.<br />We define a domain as a particular instance of a topology.<br />Also, we define an eligible domain as a domain whose nodes meet the requirements of<br />nodeAffinityPolicy and nodeTaintsPolicy.<br />e.g. If TopologyKey is "kubernetes.io/hostname", each Node is a domain of that topology.<br />And, if TopologyKey is "topology.kubernetes.io/zone", each zone is a domain of that topology.<br />It's a required field. |  |  |
| `whenUnsatisfiable` _[UnsatisfiableConstraintAction](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#unsatisfiableconstraintaction-v1-core)_ | WhenUnsatisfiable indicates how to deal with a pod if it doesn't satisfy<br />the spread constraint.<br />- DoNotSchedule (default) tells the scheduler not to schedule it.<br />- ScheduleAnyway tells the scheduler to schedule the pod in any location,<br />  but giving higher precedence to topologies that would help reduce the<br />  skew.<br />A constraint is considered "Unsatisfiable" for an incoming pod<br />if and only if every possible node assignment for that pod would violate<br />"MaxSkew" on some topology.<br />For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same<br />labelSelector spread as 3/1/1:<br />\| zone1 \| zone2 \| zone3 \|<br />\| P P P \|   P   \|   P   \|<br />If WhenUnsatisfiable is set to DoNotSchedule, incoming pod can only be scheduled<br />to zone2(zone3) to become 3/2/1(3/1/2) as ActualSkew(2-1) on zone2(zone3) satisfies<br />MaxSkew(1). In other words, the cluster can still be imbalanced, but scheduler<br />won't make it *more* imbalanced.<br />It's a required field. |  |  |
| `labelSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | LabelSelector is used to find matching pods.<br />Pods that match this label selector are counted to determine the number of pods<br />in their corresponding topology domain. |  |  |
| `minDomains` _integer_ | MinDomains indicates a minimum number of eligible domains.<br />When the number of eligible domains with matching topology keys is less than minDomains,<br />Pod Topology Spread treats "global minimum" as 0, and then the calculation of Skew is performed.<br />And when the number of eligible domains with matching topology keys equals or greater than minDomains,<br />this value has no effect on scheduling.<br />As a result, when the number of eligible domains is less than minDomains,<br />scheduler won't schedule more than maxSkew Pods to those domains.<br />If value is nil, the constraint behaves as if MinDomains is equal to 1.<br />Valid values are integers greater than 0.<br />When value is not nil, WhenUnsatisfiable must be DoNotSchedule.<br />For example, in a 3-zone cluster, MaxSkew is set to 2, MinDomains is set to 5 and pods with the same<br />labelSelector spread as 2/2/2:<br />\| zone1 \| zone2 \| zone3 \|<br />\|  P P  \|  P P  \|  P P  \|<br />The number of domains is less than 5(MinDomains), so "global minimum" is treated as 0.<br />In this situation, new pod with the same labelSelector cannot be scheduled,<br />because computed skew will be 3(3 - 0) if new Pod is scheduled to any of the three zones,<br />it will violate MaxSkew. |  |  |
| `nodeAffinityPolicy` _[NodeInclusionPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#nodeinclusionpolicy-v1-core)_ | NodeAffinityPolicy indicates how we will treat Pod's nodeAffinity/nodeSelector<br />when calculating pod topology spread skew. Options are:<br />- Honor: only nodes matching nodeAffinity/nodeSelector are included in the calculations.<br />- Ignore: nodeAffinity/nodeSelector are ignored. All nodes are included in the calculations.<br />If this value is nil, the behavior is equivalent to the Honor policy. |  |  |
| `nodeTaintsPolicy` _[NodeInclusionPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#nodeinclusionpolicy-v1-core)_ | NodeTaintsPolicy indicates how we will treat node taints when calculating<br />pod topology spread skew. Options are:<br />- Honor: nodes without taints, along with tainted nodes for which the incoming pod<br />has a toleration, are included.<br />- Ignore: node taints are ignored. All nodes are included.<br />If this value is nil, the behavior is equivalent to the Ignore policy. |  |  |
| `matchLabelKeys` _string array_ | MatchLabelKeys is a set of pod label keys to select the pods over which<br />spreading will be calculated. The keys are used to lookup values from the<br />incoming pod labels, those key-value labels are ANDed with labelSelector<br />to select the group of existing pods over which spreading will be calculated<br />for the incoming pod. The same key is forbidden to exist in both MatchLabelKeys and LabelSelector.<br />MatchLabelKeys cannot be set when LabelSelector isn't set.<br />Keys that don't exist in the incoming pod labels will<br />be ignored. A null or empty list means only match against labelSelector.<br />This is a beta field and requires the MatchLabelKeysInPodTopologySpread feature gate to be enabled (enabled by default). |  |  |


#### DNSPolicy

_Underlying type:_ _string_

DNSPolicy specifies the DNS policy for the pod.

_Validation:_
- Enum: [ClusterFirstWithHostNet ClusterFirst Default None]

_Appears in:_
- [AlertmanagerSpec](#alertmanagerspec)
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)
- [ThanosRulerSpec](#thanosrulerspec)

| Field | Description |
| --- | --- |
| `ClusterFirstWithHostNet` | DNSClusterFirstWithHostNet indicates that the pod should use cluster DNS<br />first, if it is available, then fall back on the default<br />(as determined by kubelet) DNS settings.<br /> |
| `ClusterFirst` | DNSClusterFirst indicates that the pod should use cluster DNS<br />first unless hostNetwork is true, if it is available, then<br />fall back on the default (as determined by kubelet) DNS settings.<br /> |
| `Default` | DNSDefault indicates that the pod should use the default (as<br />determined by kubelet) DNS settings.<br /> |
| `None` | DNSNone indicates that the pod should use empty DNS settings. DNS<br />parameters such as nameservers and search paths should be defined via<br />DNSConfig.<br /> |


#### Duration

_Underlying type:_ _string_

Duration is a valid time duration that can be parsed by Prometheus model.ParseDuration() function.
Supported units: y, w, d, h, m, s, ms
Examples: `30s`, `1m`, `1h20m15s`, `15d`

_Validation:_
- Pattern: `^(0|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$`

_Appears in:_
- [AlertmanagerEndpoints](#alertmanagerendpoints)
- [AlertmanagerGlobalConfig](#alertmanagerglobalconfig)
- [AzureSDConfig](#azuresdconfig)
- [CommonPrometheusFields](#commonprometheusfields)
- [ConsulSDConfig](#consulsdconfig)
- [DNSSDConfig](#dnssdconfig)
- [DigitalOceanSDConfig](#digitaloceansdconfig)
- [DockerSDConfig](#dockersdconfig)
- [DockerSwarmSDConfig](#dockerswarmsdconfig)
- [EC2SDConfig](#ec2sdconfig)
- [Endpoint](#endpoint)
- [EurekaSDConfig](#eurekasdconfig)
- [FileSDConfig](#filesdconfig)
- [GCESDConfig](#gcesdconfig)
- [HTTPSDConfig](#httpsdconfig)
- [HetznerSDConfig](#hetznersdconfig)
- [IonosSDConfig](#ionossdconfig)
- [KumaSDConfig](#kumasdconfig)
- [LightSailSDConfig](#lightsailsdconfig)
- [LinodeSDConfig](#linodesdconfig)
- [MetadataConfig](#metadataconfig)
- [NomadSDConfig](#nomadsdconfig)
- [OVHCloudSDConfig](#ovhcloudsdconfig)
- [OpenStackSDConfig](#openstacksdconfig)
- [PodMetricsEndpoint](#podmetricsendpoint)
- [ProbeSpec](#probespec)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)
- [PrometheusTracingConfig](#prometheustracingconfig)
- [PuppetDBSDConfig](#puppetdbsdconfig)
- [PushoverConfig](#pushoverconfig)
- [PushoverConfig](#pushoverconfig)
- [QuerySpec](#queryspec)
- [QueueConfig](#queueconfig)
- [RemoteReadSpec](#remotereadspec)
- [RemoteWriteSpec](#remotewritespec)
- [RetainConfig](#retainconfig)
- [Rule](#rule)
- [RuleGroup](#rulegroup)
- [ScalewaySDConfig](#scalewaysdconfig)
- [ScrapeConfigSpec](#scrapeconfigspec)
- [TSDBSpec](#tsdbspec)
- [ThanosRulerSpec](#thanosrulerspec)
- [ThanosSpec](#thanosspec)
- [WebhookConfig](#webhookconfig)
- [WebhookConfig](#webhookconfig)



#### EmbeddedObjectMetadata



EmbeddedObjectMetadata contains a subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta
Only fields which are relevant to embedded resources are included.



_Appears in:_
- [AlertmanagerSpec](#alertmanagerspec)
- [CommonPrometheusFields](#commonprometheusfields)
- [EmbeddedPersistentVolumeClaim](#embeddedpersistentvolumeclaim)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)
- [ThanosRulerSpec](#thanosrulerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name must be unique within a namespace. Is required when creating resources, although<br />some resources may allow a client to request the generation of an appropriate name<br />automatically. Name is primarily intended for creation idempotence and configuration<br />definition.<br />Cannot be updated.<br />More info: http://kubernetes.io/docs/user-guide/identifiers#names |  |  |
| `labels` _object (keys:string, values:string)_ | Map of string keys and values that can be used to organize and categorize<br />(scope and select) objects. May match selectors of replication controllers<br />and services.<br />More info: http://kubernetes.io/docs/user-guide/labels |  |  |
| `annotations` _object (keys:string, values:string)_ | Annotations is an unstructured key value map stored with a resource that may be<br />set by external tools to store and retrieve arbitrary metadata. They are not<br />queryable and should be preserved when modifying objects.<br />More info: http://kubernetes.io/docs/user-guide/annotations |  |  |


#### EmbeddedPersistentVolumeClaim



EmbeddedPersistentVolumeClaim is an embedded version of k8s.io/api/core/v1.PersistentVolumeClaim.
It contains TypeMeta and a reduced ObjectMeta.



_Appears in:_
- [StorageSpec](#storagespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metadata` _[EmbeddedObjectMetadata](#embeddedobjectmetadata)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[PersistentVolumeClaimSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#persistentvolumeclaimspec-v1-core)_ | Defines the desired characteristics of a volume requested by a pod author.<br />More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims |  |  |
| `status` _[PersistentVolumeClaimStatus](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#persistentvolumeclaimstatus-v1-core)_ | Deprecated: this field is never set. |  |  |


#### EnableFeature

_Underlying type:_ _string_



_Validation:_
- MinLength: 1

_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)



#### Endpoint



Endpoint defines an endpoint serving Prometheus metrics to be scraped by
Prometheus.



_Appears in:_
- [ServiceMonitorSpec](#servicemonitorspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `port` _string_ | Name of the Service port which this endpoint refers to.<br />It takes precedence over `targetPort`. |  |  |
| `targetPort` _[IntOrString](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#intorstring-intstr-util)_ | Name or number of the target port of the `Pod` object behind the<br />Service. The port must be specified with the container's port property. |  |  |
| `path` _string_ | HTTP path from which to scrape for metrics.<br />If empty, Prometheus uses the default value (e.g. `/metrics`). |  |  |
| `scheme` _string_ | HTTP scheme to use for scraping.<br />`http` and `https` are the expected values unless you rewrite the<br />`__scheme__` label via relabeling.<br />If empty, Prometheus uses the default value `http`. |  | Enum: [http https] <br /> |
| `params` _object (keys:string, values:string array)_ | params define optional HTTP URL parameters. |  |  |
| `interval` _[Duration](#duration)_ | Interval at which Prometheus scrapes the metrics from the target.<br />If empty, Prometheus uses the global scrape interval. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `scrapeTimeout` _[Duration](#duration)_ | Timeout after which Prometheus considers the scrape to be failed.<br />If empty, Prometheus uses the global scrape timeout unless it is less<br />than the target's scrape interval value in which the latter is used.<br />The value cannot be greater than the scrape interval otherwise the operator will reject the resource. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `tlsConfig` _[TLSConfig](#tlsconfig)_ | TLS configuration to use when scraping the target. |  |  |
| `bearerTokenFile` _string_ | File to read bearer token for scraping the target.<br />Deprecated: use `authorization` instead. |  |  |
| `bearerTokenSecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | `bearerTokenSecret` specifies a key of a Secret containing the bearer<br />token for scraping targets. The secret needs to be in the same namespace<br />as the ServiceMonitor object and readable by the Prometheus Operator.<br />Deprecated: use `authorization` instead. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | `authorization` configures the Authorization header credentials to use when<br />scraping the target.<br />Cannot be set at the same time as `basicAuth`, or `oauth2`. |  |  |
| `honorLabels` _boolean_ | When true, `honorLabels` preserves the metric's labels when they collide<br />with the target's labels. |  |  |
| `honorTimestamps` _boolean_ | `honorTimestamps` controls whether Prometheus preserves the timestamps<br />when exposed by the target. |  |  |
| `trackTimestampsStaleness` _boolean_ | `trackTimestampsStaleness` defines whether Prometheus tracks staleness of<br />the metrics that have an explicit timestamp present in scraped data.<br />Has no effect if `honorTimestamps` is false.<br />It requires Prometheus >= v2.48.0. |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | `basicAuth` configures the Basic Authentication credentials to use when<br />scraping the target.<br />Cannot be set at the same time as `authorization`, or `oauth2`. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | `oauth2` configures the OAuth2 settings to use when scraping the target.<br />It requires Prometheus >= 2.27.0.<br />Cannot be set at the same time as `authorization`, or `basicAuth`. |  |  |
| `metricRelabelings` _[RelabelConfig](#relabelconfig) array_ | `metricRelabelings` configures the relabeling rules to apply to the<br />samples before ingestion. |  |  |
| `relabelings` _[RelabelConfig](#relabelconfig) array_ | `relabelings` configures the relabeling rules to apply the target's<br />metadata labels.<br />The Operator automatically adds relabelings for a few standard Kubernetes fields.<br />The original scrape job's name is available via the `__tmp_prometheus_job_name` label.<br />More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config |  |  |
| `proxyUrl` _string_ | `proxyURL` configures the HTTP Proxy URL (e.g.<br />"http://proxyserver:2195") to go through when scraping the target. |  |  |
| `followRedirects` _boolean_ | `followRedirects` defines whether the scrape requests should follow HTTP<br />3xx redirects. |  |  |
| `enableHttp2` _boolean_ | `enableHttp2` can be used to disable HTTP2 when scraping the target. |  |  |
| `filterRunning` _boolean_ | When true, the pods which are not running (e.g. either in Failed or<br />Succeeded state) are dropped during the target discovery.<br />If unset, the filtering is enabled.<br />More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase |  |  |


#### Exemplars







_Appears in:_
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `maxSize` _integer_ | Maximum number of exemplars stored in memory for all series.<br />exemplar-storage itself must be enabled using the `spec.enableFeature`<br />option for exemplars to be scraped in the first place.<br />If not set, Prometheus uses its default value. A value of zero or less<br />than zero disables the storage. |  |  |


#### GlobalSMTPConfig



GlobalSMTPConfig configures global SMTP parameters.
See https://prometheus.io/docs/alerting/latest/configuration/#configuration-file



_Appears in:_
- [AlertmanagerGlobalConfig](#alertmanagerglobalconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `from` _string_ | The default SMTP From header field. |  |  |
| `smartHost` _[HostPort](#hostport)_ | The default SMTP smarthost used for sending emails. |  |  |
| `hello` _string_ | The default hostname to identify to the SMTP server. |  |  |
| `authUsername` _string_ | SMTP Auth using CRAM-MD5, LOGIN and PLAIN. If empty, Alertmanager doesn't authenticate to the SMTP server. |  |  |
| `authPassword` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | SMTP Auth using LOGIN and PLAIN. |  |  |
| `authIdentity` _string_ | SMTP Auth using PLAIN |  |  |
| `authSecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | SMTP Auth using CRAM-MD5. |  |  |
| `requireTLS` _boolean_ | The default SMTP TLS requirement.<br />Note that Go does not support unencrypted connections to remote SMTP endpoints. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | The default TLS configuration for SMTP receivers |  |  |


#### GoDuration

_Underlying type:_ _string_

GoDuration is a valid time duration that can be parsed by Go's time.ParseDuration() function.
Supported units: h, m, s, ms
Examples: `45ms`, `30s`, `1m`, `1h20m15s`

_Validation:_
- Pattern: `^(0|(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$`

_Appears in:_
- [AlertmanagerSpec](#alertmanagerspec)



#### HTTPConfig



HTTPConfig defines a client HTTP configuration.
See https://prometheus.io/docs/alerting/latest/configuration/#http_config



_Appears in:_
- [AlertmanagerGlobalConfig](#alertmanagerglobalconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header configuration for the client.<br />This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+. |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth for the client.<br />This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | OAuth2 client credentials used to fetch a token for the targets. |  |  |
| `bearerTokenSecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the bearer token to be used by the client<br />for authentication.<br />The secret needs to be in the same namespace as the Alertmanager<br />object and accessible by the Prometheus Operator. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration for the client. |  |  |
| `followRedirects` _boolean_ | FollowRedirects specifies whether the client should follow HTTP 3xx redirects. |  |  |


#### HostAlias



HostAlias holds the mapping between IP and hostnames that will be injected as an entry in the
pod's hosts file.



_Appears in:_
- [AlertmanagerSpec](#alertmanagerspec)
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)
- [ThanosRulerSpec](#thanosrulerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `ip` _string_ | IP address of the host file entry. |  | Required: \{\} <br /> |
| `hostnames` _string array_ | Hostnames for the above IP address. |  | Required: \{\} <br /> |


#### HostPort



HostPort represents a "host:port" network address.



_Appears in:_
- [GlobalSMTPConfig](#globalsmtpconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ | Defines the host's address, it can be a DNS name or a literal IP address. |  | MinLength: 1 <br /> |
| `port` _string_ | Defines the host's port, it can be a literal port number or a port name. |  | MinLength: 1 <br /> |


#### LabelName

_Underlying type:_ _string_

LabelName is a valid Prometheus label name which may only contain ASCII
letters, numbers, as well as underscores.

_Validation:_
- Pattern: `^[a-zA-Z_][a-zA-Z0-9_]*$`

_Appears in:_
- [RelabelConfig](#relabelconfig)



#### ManagedIdentity



ManagedIdentity defines the Azure User-assigned Managed identity.



_Appears in:_
- [AzureAD](#azuread)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `clientId` _string_ | The client id |  |  |


#### MetadataConfig



MetadataConfig configures the sending of series metadata to the remote storage.



_Appears in:_
- [RemoteWriteSpec](#remotewritespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `send` _boolean_ | Defines whether metric metadata is sent to the remote storage or not. |  |  |
| `sendInterval` _[Duration](#duration)_ | Defines how frequently metric metadata is sent to the remote storage. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `maxSamplesPerSend` _integer_ | MaxSamplesPerSend is the maximum number of metadata samples per send.<br />It requires Prometheus >= v2.29.0. |  | Minimum: -1 <br /> |


#### NameEscapingSchemeOptions

_Underlying type:_ _string_

Specifies the character escaping scheme that will be applied when scraping
for metric and label names that do not conform to the legacy Prometheus
character set.

Supported values are:

  - `AllowUTF8`, full UTF-8 support, no escaping needed.
  - `Underscores`, legacy-invalid characters are escaped to underscores.
  - `Dots`, dot characters are escaped to `_dot_`, underscores to `__`, and
    all other legacy-invalid characters to underscores.
  - `Values`, the string is prefixed by `U__` and all invalid characters are
    escaped to their unicode value, surrounded by underscores.

_Validation:_
- Enum: [AllowUTF8 Underscores Dots Values]

_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description |
| --- | --- |
| `AllowUTF8` |  |
| `Underscores` |  |
| `Dots` |  |
| `Values` |  |


#### NameValidationSchemeOptions

_Underlying type:_ _string_

Specifies the validation scheme for metric and label names.

Supported values are:
  - `UTF8NameValidationScheme` for UTF-8 support.
  - `LegacyNameValidationScheme` for letters, numbers, colons, and underscores.

Note that `LegacyNameValidationScheme` cannot be used along with the
OpenTelemetry `NoUTF8EscapingWithSuffixes` translation strategy (if
enabled).

_Validation:_
- Enum: [UTF8 Legacy]

_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description |
| --- | --- |
| `UTF8` |  |
| `Legacy` |  |


#### NamespaceSelector



NamespaceSelector is a selector for selecting either all namespaces or a
list of namespaces.
If `any` is true, it takes precedence over `matchNames`.
If `matchNames` is empty and `any` is false, it means that the objects are
selected from the current namespace.



_Appears in:_
- [PodMonitorSpec](#podmonitorspec)
- [ProbeTargetIngress](#probetargetingress)
- [ServiceMonitorSpec](#servicemonitorspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `any` _boolean_ | Boolean describing whether all namespaces are selected in contrast to a<br />list restricting them. |  |  |
| `matchNames` _string array_ | List of namespace names to select from. |  |  |


#### NativeHistogramConfig



NativeHistogramConfig extends the native histogram configuration settings.



_Appears in:_
- [PodMonitorSpec](#podmonitorspec)
- [ProbeSpec](#probespec)
- [ScrapeConfigSpec](#scrapeconfigspec)
- [ServiceMonitorSpec](#servicemonitorspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `scrapeClassicHistograms` _boolean_ | Whether to scrape a classic histogram that is also exposed as a native histogram.<br />It requires Prometheus >= v2.45.0. |  |  |
| `nativeHistogramBucketLimit` _integer_ | If there are more than this many buckets in a native histogram,<br />buckets will be merged to stay within the limit.<br />It requires Prometheus >= v2.45.0. |  |  |
| `nativeHistogramMinBucketFactor` _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#quantity-resource-api)_ | If the growth factor of one bucket to the next is smaller than this,<br />buckets will be merged to increase the factor sufficiently.<br />It requires Prometheus >= v2.50.0. |  |  |
| `convertClassicHistogramsToNHCB` _boolean_ | Whether to convert all scraped classic histograms into a native histogram with custom buckets.<br />It requires Prometheus >= v3.0.0. |  |  |


#### NonEmptyDuration

_Underlying type:_ _string_

NonEmptyDuration is a valid time duration that can be parsed by Prometheus model.ParseDuration() function.
Compared to Duration,  NonEmptyDuration enforces a minimum length of 1.
Supported units: y, w, d, h, m, s, ms
Examples: `30s`, `1m`, `1h20m15s`, `15d`

_Validation:_
- MinLength: 1
- Pattern: `^(0|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$`

_Appears in:_
- [Rule](#rule)



#### OAuth2



OAuth2 configures OAuth2 settings.



_Appears in:_
- [AzureSDConfig](#azuresdconfig)
- [ConsulSDConfig](#consulsdconfig)
- [DigitalOceanSDConfig](#digitaloceansdconfig)
- [DockerSDConfig](#dockersdconfig)
- [DockerSwarmSDConfig](#dockerswarmsdconfig)
- [Endpoint](#endpoint)
- [EurekaSDConfig](#eurekasdconfig)
- [HTTPConfig](#httpconfig)
- [HTTPConfig](#httpconfig)
- [HTTPConfig](#httpconfig)
- [HTTPSDConfig](#httpsdconfig)
- [HetznerSDConfig](#hetznersdconfig)
- [IonosSDConfig](#ionossdconfig)
- [KubernetesSDConfig](#kubernetessdconfig)
- [KumaSDConfig](#kumasdconfig)
- [LightSailSDConfig](#lightsailsdconfig)
- [LinodeSDConfig](#linodesdconfig)
- [NomadSDConfig](#nomadsdconfig)
- [PodMetricsEndpoint](#podmetricsendpoint)
- [ProbeSpec](#probespec)
- [PuppetDBSDConfig](#puppetdbsdconfig)
- [RemoteReadSpec](#remotereadspec)
- [RemoteWriteSpec](#remotewritespec)
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `clientId` _[SecretOrConfigMap](#secretorconfigmap)_ | `clientId` specifies a key of a Secret or ConfigMap containing the<br />OAuth2 client's ID. |  |  |
| `clientSecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | `clientSecret` specifies a key of a Secret containing the OAuth2<br />client's secret. |  |  |
| `tokenUrl` _string_ | `tokenURL` configures the URL to fetch the token from. |  | MinLength: 1 <br /> |
| `scopes` _string array_ | `scopes` defines the OAuth2 scopes used for the token request. |  |  |
| `endpointParams` _object (keys:string, values:string)_ | `endpointParams` configures the HTTP parameters to append to the token<br />URL. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration to use when connecting to the OAuth2 server.<br />It requires Prometheus >= v2.43.0. |  |  |




#### OTLPConfig



OTLPConfig is the configuration for writing to the OTLP endpoint.



_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `promoteResourceAttributes` _string array_ | List of OpenTelemetry Attributes that should be promoted to metric labels, defaults to none. |  | MinItems: 1 <br />items:MinLength: 1 <br /> |
| `translationStrategy` _[TranslationStrategyOption](#translationstrategyoption)_ | Configures how the OTLP receiver endpoint translates the incoming metrics.<br />It requires Prometheus >= v3.0.0. |  | Enum: [NoUTF8EscapingWithSuffixes UnderscoreEscapingWithSuffixes NoTranslation] <br /> |
| `keepIdentifyingResourceAttributes` _boolean_ | Enables adding `service.name`, `service.namespace` and `service.instance.id`<br />resource attributes to the `target_info` metric, on top of converting them into the `instance` and `job` labels.<br />It requires Prometheus >= v3.1.0. |  |  |
| `convertHistogramsToNHCB` _boolean_ | Configures optional translation of OTLP explicit bucket histograms into native histograms with custom buckets.<br />It requires Prometheus >= v3.4.0. |  |  |


#### ObjectReference



ObjectReference references a PodMonitor, ServiceMonitor, Probe or PrometheusRule object.



_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)
- [ThanosRulerSpec](#thanosrulerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _string_ | Group of the referent. When not specified, it defaults to `monitoring.coreos.com` | monitoring.coreos.com | Enum: [monitoring.coreos.com] <br /> |
| `resource` _string_ | Resource of the referent. |  | Enum: [prometheusrules servicemonitors podmonitors probes scrapeconfigs] <br />Required: \{\} <br /> |
| `namespace` _string_ | Namespace of the referent.<br />More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/ |  | MinLength: 1 <br />Required: \{\} <br /> |
| `name` _string_ | Name of the referent. When not set, all resources in the namespace are matched. |  |  |


#### PodDNSConfig



PodDNSConfig defines the DNS parameters of a pod in addition to
those generated from DNSPolicy.



_Appears in:_
- [AlertmanagerSpec](#alertmanagerspec)
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)
- [ThanosRulerSpec](#thanosrulerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `nameservers` _string array_ | A list of DNS name server IP addresses.<br />This will be appended to the base nameservers generated from DNSPolicy. |  | Optional: \{\} <br />items:MinLength: 1 <br /> |
| `searches` _string array_ | A list of DNS search domains for host-name lookup.<br />This will be appended to the base search paths generated from DNSPolicy. |  | Optional: \{\} <br />items:MinLength: 1 <br /> |
| `options` _[PodDNSConfigOption](#poddnsconfigoption) array_ | A list of DNS resolver options.<br />This will be merged with the base options generated from DNSPolicy.<br />Resolution options given in Options<br />will override those that appear in the base DNSPolicy. |  | Optional: \{\} <br /> |


#### PodDNSConfigOption



PodDNSConfigOption defines DNS resolver options of a pod.



_Appears in:_
- [PodDNSConfig](#poddnsconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is required and must be unique. |  | MinLength: 1 <br /> |
| `value` _string_ | Value is optional. |  | Optional: \{\} <br /> |


#### PodMetricsEndpoint



PodMetricsEndpoint defines an endpoint serving Prometheus metrics to be scraped by
Prometheus.



_Appears in:_
- [PodMonitorSpec](#podmonitorspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `port` _string_ | The `Pod` port name which exposes the endpoint.<br />It takes precedence over the `portNumber` and `targetPort` fields. |  |  |
| `portNumber` _integer_ | The `Pod` port number which exposes the endpoint. |  | Maximum: 65535 <br />Minimum: 1 <br /> |
| `targetPort` _[IntOrString](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#intorstring-intstr-util)_ | Name or number of the target port of the `Pod` object behind the Service, the<br />port must be specified with container port property.<br />Deprecated: use 'port' or 'portNumber' instead. |  |  |
| `path` _string_ | HTTP path from which to scrape for metrics.<br />If empty, Prometheus uses the default value (e.g. `/metrics`). |  |  |
| `scheme` _string_ | HTTP scheme to use for scraping.<br />`http` and `https` are the expected values unless you rewrite the<br />`__scheme__` label via relabeling.<br />If empty, Prometheus uses the default value `http`. |  | Enum: [http https] <br /> |
| `params` _object (keys:string, values:string array)_ | `params` define optional HTTP URL parameters. |  |  |
| `interval` _[Duration](#duration)_ | Interval at which Prometheus scrapes the metrics from the target.<br />If empty, Prometheus uses the global scrape interval. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `scrapeTimeout` _[Duration](#duration)_ | Timeout after which Prometheus considers the scrape to be failed.<br />If empty, Prometheus uses the global scrape timeout unless it is less<br />than the target's scrape interval value in which the latter is used.<br />The value cannot be greater than the scrape interval otherwise the operator will reject the resource. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration to use when scraping the target. |  |  |
| `bearerTokenSecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | `bearerTokenSecret` specifies a key of a Secret containing the bearer<br />token for scraping targets. The secret needs to be in the same namespace<br />as the PodMonitor object and readable by the Prometheus Operator.<br />Deprecated: use `authorization` instead. |  |  |
| `honorLabels` _boolean_ | When true, `honorLabels` preserves the metric's labels when they collide<br />with the target's labels. |  |  |
| `honorTimestamps` _boolean_ | `honorTimestamps` controls whether Prometheus preserves the timestamps<br />when exposed by the target. |  |  |
| `trackTimestampsStaleness` _boolean_ | `trackTimestampsStaleness` defines whether Prometheus tracks staleness of<br />the metrics that have an explicit timestamp present in scraped data.<br />Has no effect if `honorTimestamps` is false.<br />It requires Prometheus >= v2.48.0. |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | `basicAuth` configures the Basic Authentication credentials to use when<br />scraping the target.<br />Cannot be set at the same time as `authorization`, or `oauth2`. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | `oauth2` configures the OAuth2 settings to use when scraping the target.<br />It requires Prometheus >= 2.27.0.<br />Cannot be set at the same time as `authorization`, or `basicAuth`. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | `authorization` configures the Authorization header credentials to use when<br />scraping the target.<br />Cannot be set at the same time as `basicAuth`, or `oauth2`. |  |  |
| `metricRelabelings` _[RelabelConfig](#relabelconfig) array_ | `metricRelabelings` configures the relabeling rules to apply to the<br />samples before ingestion. |  |  |
| `relabelings` _[RelabelConfig](#relabelconfig) array_ | `relabelings` configures the relabeling rules to apply the target's<br />metadata labels.<br />The Operator automatically adds relabelings for a few standard Kubernetes fields.<br />The original scrape job's name is available via the `__tmp_prometheus_job_name` label.<br />More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config |  |  |
| `proxyUrl` _string_ | `proxyURL` configures the HTTP Proxy URL (e.g.<br />"http://proxyserver:2195") to go through when scraping the target. |  |  |
| `followRedirects` _boolean_ | `followRedirects` defines whether the scrape requests should follow HTTP<br />3xx redirects. |  |  |
| `enableHttp2` _boolean_ | `enableHttp2` can be used to disable HTTP2 when scraping the target. |  |  |
| `filterRunning` _boolean_ | When true, the pods which are not running (e.g. either in Failed or<br />Succeeded state) are dropped during the target discovery.<br />If unset, the filtering is enabled.<br />More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase |  |  |


#### PodMonitor



The `PodMonitor` custom resource definition (CRD) defines how `Prometheus` and `PrometheusAgent` can scrape metrics from a group of pods.
Among other things, it allows to specify:
* The pods to scrape via label selectors.
* The container ports to scrape.
* Authentication credentials to use.
* Target and metric relabeling.

`Prometheus` and `PrometheusAgent` objects select `PodMonitor` objects using label and namespace selectors.



_Appears in:_
- [PodMonitorList](#podmonitorlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[PodMonitorSpec](#podmonitorspec)_ | Specification of desired Pod selection for target discovery by Prometheus. |  |  |




#### PodMonitorSpec



PodMonitorSpec contains specification parameters for a PodMonitor.



_Appears in:_
- [PodMonitor](#podmonitor)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `jobLabel` _string_ | The label to use to retrieve the job name from.<br />`jobLabel` selects the label from the associated Kubernetes `Pod`<br />object which will be used as the `job` label for all metrics.<br />For example if `jobLabel` is set to `foo` and the Kubernetes `Pod`<br />object is labeled with `foo: bar`, then Prometheus adds the `job="bar"`<br />label to all ingested metrics.<br />If the value of this field is empty, the `job` label of the metrics<br />defaults to the namespace and name of the PodMonitor object (e.g. `<namespace>/<name>`). |  |  |
| `podTargetLabels` _string array_ | `podTargetLabels` defines the labels which are transferred from the<br />associated Kubernetes `Pod` object onto the ingested metrics. |  |  |
| `podMetricsEndpoints` _[PodMetricsEndpoint](#podmetricsendpoint) array_ | Defines how to scrape metrics from the selected pods. |  |  |
| `selector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Label selector to select the Kubernetes `Pod` objects to scrape metrics from. |  |  |
| `selectorMechanism` _[SelectorMechanism](#selectormechanism)_ | Mechanism used to select the endpoints to scrape.<br />By default, the selection process relies on relabel configurations to filter the discovered targets.<br />Alternatively, you can opt in for role selectors, which may offer better efficiency in large clusters.<br />Which strategy is best for your use case needs to be carefully evaluated.<br />It requires Prometheus >= v2.17.0. |  | Enum: [RelabelConfig RoleSelector] <br /> |
| `namespaceSelector` _[NamespaceSelector](#namespaceselector)_ | `namespaceSelector` defines in which namespace(s) Prometheus should discover the pods.<br />By default, the pods are discovered in the same namespace as the `PodMonitor` object but it is possible to select pods across different/all namespaces. |  |  |
| `sampleLimit` _integer_ | `sampleLimit` defines a per-scrape limit on the number of scraped samples<br />that will be accepted. |  |  |
| `targetLimit` _integer_ | `targetLimit` defines a limit on the number of scraped targets that will<br />be accepted. |  |  |
| `scrapeProtocols` _[ScrapeProtocol](#scrapeprotocol) array_ | `scrapeProtocols` defines the protocols to negotiate during a scrape. It tells clients the<br />protocols supported by Prometheus in order of preference (from most to least preferred).<br />If unset, Prometheus uses its default value.<br />It requires Prometheus >= v2.49.0. |  | Enum: [PrometheusProto OpenMetricsText0.0.1 OpenMetricsText1.0.0 PrometheusText0.0.4 PrometheusText1.0.0] <br /> |
| `fallbackScrapeProtocol` _[ScrapeProtocol](#scrapeprotocol)_ | The protocol to use if a scrape returns blank, unparseable, or otherwise invalid Content-Type.<br />It requires Prometheus >= v3.0.0. |  | Enum: [PrometheusProto OpenMetricsText0.0.1 OpenMetricsText1.0.0 PrometheusText0.0.4 PrometheusText1.0.0] <br /> |
| `labelLimit` _integer_ | Per-scrape limit on number of labels that will be accepted for a sample.<br />It requires Prometheus >= v2.27.0. |  |  |
| `labelNameLengthLimit` _integer_ | Per-scrape limit on length of labels name that will be accepted for a sample.<br />It requires Prometheus >= v2.27.0. |  |  |
| `labelValueLengthLimit` _integer_ | Per-scrape limit on length of labels value that will be accepted for a sample.<br />It requires Prometheus >= v2.27.0. |  |  |
| `scrapeClassicHistograms` _boolean_ | Whether to scrape a classic histogram that is also exposed as a native histogram.<br />It requires Prometheus >= v2.45.0. |  |  |
| `nativeHistogramBucketLimit` _integer_ | If there are more than this many buckets in a native histogram,<br />buckets will be merged to stay within the limit.<br />It requires Prometheus >= v2.45.0. |  |  |
| `nativeHistogramMinBucketFactor` _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#quantity-resource-api)_ | If the growth factor of one bucket to the next is smaller than this,<br />buckets will be merged to increase the factor sufficiently.<br />It requires Prometheus >= v2.50.0. |  |  |
| `convertClassicHistogramsToNHCB` _boolean_ | Whether to convert all scraped classic histograms into a native histogram with custom buckets.<br />It requires Prometheus >= v3.0.0. |  |  |
| `keepDroppedTargets` _integer_ | Per-scrape limit on the number of targets dropped by relabeling<br />that will be kept in memory. 0 means no limit.<br />It requires Prometheus >= v2.47.0. |  |  |
| `attachMetadata` _[AttachMetadata](#attachmetadata)_ | `attachMetadata` defines additional metadata which is added to the<br />discovered targets.<br />It requires Prometheus >= v2.35.0. |  |  |
| `scrapeClass` _string_ | The scrape class to apply. |  | MinLength: 1 <br /> |
| `bodySizeLimit` _[ByteSize](#bytesize)_ | When defined, bodySizeLimit specifies a job level limit on the size<br />of uncompressed response body that will be accepted by Prometheus.<br />It requires Prometheus >= v2.28.0. |  | Pattern: `(^0\|([0-9]*[.])?[0-9]+((K\|M\|G\|T\|E\|P)i?)?B)$` <br /> |


#### Probe



The `Probe` custom resource definition (CRD) defines how to scrape metrics from prober exporters such as the [blackbox exporter](https://github.com/prometheus/blackbox_exporter).

The `Probe` resource needs 2 pieces of information:
* The list of probed addresses which can be defined statically or by discovering Kubernetes Ingress objects.
* The prober which exposes the availability of probed endpoints (over various protocols such HTTP, TCP, ICMP, ...) as Prometheus metrics.

`Prometheus` and `PrometheusAgent` objects select `Probe` objects using label and namespace selectors.



_Appears in:_
- [ProbeList](#probelist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[ProbeSpec](#probespec)_ | Specification of desired Ingress selection for target discovery by Prometheus. |  |  |




#### ProbeSpec



ProbeSpec contains specification parameters for a Probe.



_Appears in:_
- [Probe](#probe)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `jobName` _string_ | The job name assigned to scraped metrics by default. |  |  |
| `prober` _[ProberSpec](#proberspec)_ | Specification for the prober to use for probing targets.<br />The prober.URL parameter is required. Targets cannot be probed if left empty. |  |  |
| `module` _string_ | The module to use for probing specifying how to probe the target.<br />Example module configuring in the blackbox exporter:<br />https://github.com/prometheus/blackbox_exporter/blob/master/example.yml |  |  |
| `targets` _[ProbeTargets](#probetargets)_ | Targets defines a set of static or dynamically discovered targets to probe. |  |  |
| `interval` _[Duration](#duration)_ | Interval at which targets are probed using the configured prober.<br />If not specified Prometheus' global scrape interval is used. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `scrapeTimeout` _[Duration](#duration)_ | Timeout for scraping metrics from the Prometheus exporter.<br />If not specified, the Prometheus global scrape timeout is used.<br />The value cannot be greater than the scrape interval otherwise the operator will reject the resource. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration to use when scraping the endpoint. |  |  |
| `bearerTokenSecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Secret to mount to read bearer token for scraping targets. The secret<br />needs to be in the same namespace as the probe and accessible by<br />the Prometheus Operator. |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth allow an endpoint to authenticate over basic authentication.<br />More info: https://prometheus.io/docs/operating/configuration/#endpoint |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer. |  |  |
| `metricRelabelings` _[RelabelConfig](#relabelconfig) array_ | MetricRelabelConfigs to apply to samples before ingestion. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization section for this endpoint |  |  |
| `sampleLimit` _integer_ | SampleLimit defines per-scrape limit on number of scraped samples that will be accepted. |  |  |
| `targetLimit` _integer_ | TargetLimit defines a limit on the number of scraped targets that will be accepted. |  |  |
| `scrapeProtocols` _[ScrapeProtocol](#scrapeprotocol) array_ | `scrapeProtocols` defines the protocols to negotiate during a scrape. It tells clients the<br />protocols supported by Prometheus in order of preference (from most to least preferred).<br />If unset, Prometheus uses its default value.<br />It requires Prometheus >= v2.49.0. |  | Enum: [PrometheusProto OpenMetricsText0.0.1 OpenMetricsText1.0.0 PrometheusText0.0.4 PrometheusText1.0.0] <br /> |
| `fallbackScrapeProtocol` _[ScrapeProtocol](#scrapeprotocol)_ | The protocol to use if a scrape returns blank, unparseable, or otherwise invalid Content-Type.<br />It requires Prometheus >= v3.0.0. |  | Enum: [PrometheusProto OpenMetricsText0.0.1 OpenMetricsText1.0.0 PrometheusText0.0.4 PrometheusText1.0.0] <br /> |
| `labelLimit` _integer_ | Per-scrape limit on number of labels that will be accepted for a sample.<br />Only valid in Prometheus versions 2.27.0 and newer. |  |  |
| `labelNameLengthLimit` _integer_ | Per-scrape limit on length of labels name that will be accepted for a sample.<br />Only valid in Prometheus versions 2.27.0 and newer. |  |  |
| `labelValueLengthLimit` _integer_ | Per-scrape limit on length of labels value that will be accepted for a sample.<br />Only valid in Prometheus versions 2.27.0 and newer. |  |  |
| `scrapeClassicHistograms` _boolean_ | Whether to scrape a classic histogram that is also exposed as a native histogram.<br />It requires Prometheus >= v2.45.0. |  |  |
| `nativeHistogramBucketLimit` _integer_ | If there are more than this many buckets in a native histogram,<br />buckets will be merged to stay within the limit.<br />It requires Prometheus >= v2.45.0. |  |  |
| `nativeHistogramMinBucketFactor` _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#quantity-resource-api)_ | If the growth factor of one bucket to the next is smaller than this,<br />buckets will be merged to increase the factor sufficiently.<br />It requires Prometheus >= v2.50.0. |  |  |
| `convertClassicHistogramsToNHCB` _boolean_ | Whether to convert all scraped classic histograms into a native histogram with custom buckets.<br />It requires Prometheus >= v3.0.0. |  |  |
| `keepDroppedTargets` _integer_ | Per-scrape limit on the number of targets dropped by relabeling<br />that will be kept in memory. 0 means no limit.<br />It requires Prometheus >= v2.47.0. |  |  |
| `scrapeClass` _string_ | The scrape class to apply. |  | MinLength: 1 <br /> |


#### ProbeTargetIngress



ProbeTargetIngress defines the set of Ingress objects considered for probing.
The operator configures a target for each host/path combination of each ingress object.



_Appears in:_
- [ProbeTargets](#probetargets)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `selector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Selector to select the Ingress objects. |  |  |
| `namespaceSelector` _[NamespaceSelector](#namespaceselector)_ | From which namespaces to select Ingress objects. |  |  |
| `relabelingConfigs` _[RelabelConfig](#relabelconfig) array_ | RelabelConfigs to apply to the label set of the target before it gets<br />scraped.<br />The original ingress address is available via the<br />`__tmp_prometheus_ingress_address` label. It can be used to customize the<br />probed URL.<br />The original scrape job's name is available via the `__tmp_prometheus_job_name` label.<br />More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config |  |  |


#### ProbeTargetStaticConfig



ProbeTargetStaticConfig defines the set of static targets considered for probing.



_Appears in:_
- [ProbeTargets](#probetargets)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `static` _string array_ | The list of hosts to probe. |  |  |
| `labels` _object (keys:string, values:string)_ | Labels assigned to all metrics scraped from the targets. |  |  |
| `relabelingConfigs` _[RelabelConfig](#relabelconfig) array_ | RelabelConfigs to apply to the label set of the targets before it gets<br />scraped.<br />More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config |  |  |


#### ProbeTargets



ProbeTargets defines how to discover the probed targets.
One of the `staticConfig` or `ingress` must be defined.
If both are defined, `staticConfig` takes precedence.



_Appears in:_
- [ProbeSpec](#probespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `staticConfig` _[ProbeTargetStaticConfig](#probetargetstaticconfig)_ | staticConfig defines the static list of targets to probe and the<br />relabeling configuration.<br />If `ingress` is also defined, `staticConfig` takes precedence.<br />More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#static_config. |  |  |
| `ingress` _[ProbeTargetIngress](#probetargetingress)_ | ingress defines the Ingress objects to probe and the relabeling<br />configuration.<br />If `staticConfig` is also defined, `staticConfig` takes precedence. |  |  |




#### ProberSpec



ProberSpec contains specification parameters for the Prober used for probing.



_Appears in:_
- [ProbeSpec](#probespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `url` _string_ | Mandatory URL of the prober. |  |  |
| `scheme` _string_ | HTTP scheme to use for scraping.<br />`http` and `https` are the expected values unless you rewrite the `__scheme__` label via relabeling.<br />If empty, Prometheus uses the default value `http`. |  | Enum: [http https] <br /> |
| `path` _string_ | Path to collect metrics from.<br />Defaults to `/probe`. | /probe |  |
| `proxyUrl` _string_ | Optional ProxyURL. |  |  |


#### Prometheus



The `Prometheus` custom resource definition (CRD) defines a desired [Prometheus](https://prometheus.io/docs/prometheus) setup to run in a Kubernetes cluster. It allows to specify many options such as the number of replicas, persistent storage, and Alertmanagers where firing alerts should be sent and many more.

For each `Prometheus` resource, the Operator deploys one or several `StatefulSet` objects in the same namespace. The number of StatefulSets is equal to the number of shards which is 1 by default.

The resource defines via label and namespace selectors which `ServiceMonitor`, `PodMonitor`, `Probe` and `PrometheusRule` objects should be associated to the deployed Prometheus instances.

The Operator continuously reconciles the scrape and rules configuration and a sidecar container running in the Prometheus pods triggers a reload of the configuration when needed.



_Appears in:_
- [PrometheusList](#prometheuslist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[PrometheusSpec](#prometheusspec)_ | Specification of the desired behavior of the Prometheus cluster. More info:<br />https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status |  |  |
| `status` _[PrometheusStatus](#prometheusstatus)_ | Most recent observed status of the Prometheus cluster. Read-only.<br />More info:<br />https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status |  |  |






#### PrometheusRule



The `PrometheusRule` custom resource definition (CRD) defines [alerting](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/) and [recording](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/) rules to be evaluated by `Prometheus` or `ThanosRuler` objects.

`Prometheus` and `ThanosRuler` objects select `PrometheusRule` objects using label and namespace selectors.



_Appears in:_
- [PrometheusRuleList](#prometheusrulelist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[PrometheusRuleSpec](#prometheusrulespec)_ | Specification of desired alerting rule definitions for Prometheus. |  |  |


#### PrometheusRuleExcludeConfig



PrometheusRuleExcludeConfig enables users to configure excluded
PrometheusRule names and their namespaces to be ignored while enforcing
namespace label for alerts and metrics.



_Appears in:_
- [PrometheusSpec](#prometheusspec)
- [ThanosRulerSpec](#thanosrulerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `ruleNamespace` _string_ | Namespace of the excluded PrometheusRule object. |  |  |
| `ruleName` _string_ | Name of the excluded PrometheusRule object. |  |  |




#### PrometheusRuleSpec



PrometheusRuleSpec contains specification parameters for a Rule.



_Appears in:_
- [PrometheusRule](#prometheusrule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `groups` _[RuleGroup](#rulegroup) array_ | Content of Prometheus rule file |  |  |


#### PrometheusSpec



PrometheusSpec is a specification of the desired behavior of the Prometheus cluster. More info:
https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status



_Appears in:_
- [Prometheus](#prometheus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `podMetadata` _[EmbeddedObjectMetadata](#embeddedobjectmetadata)_ | PodMetadata configures labels and annotations which are propagated to the Prometheus pods.<br />The following items are reserved and cannot be overridden:<br />* "prometheus" label, set to the name of the Prometheus object.<br />* "app.kubernetes.io/instance" label, set to the name of the Prometheus object.<br />* "app.kubernetes.io/managed-by" label, set to "prometheus-operator".<br />* "app.kubernetes.io/name" label, set to "prometheus".<br />* "app.kubernetes.io/version" label, set to the Prometheus version.<br />* "operator.prometheus.io/name" label, set to the name of the Prometheus object.<br />* "operator.prometheus.io/shard" label, set to the shard number of the Prometheus object.<br />* "kubectl.kubernetes.io/default-container" annotation, set to "prometheus". |  |  |
| `serviceMonitorSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | ServiceMonitors to be selected for target discovery. An empty label<br />selector matches all objects. A null label selector matches no objects.<br />If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`<br />and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.<br />The Prometheus operator will ensure that the Prometheus configuration's<br />Secret exists, but it is the responsibility of the user to provide the raw<br />gzipped Prometheus configuration under the `prometheus.yaml.gz` key.<br />This behavior is *deprecated* and will be removed in the next major version<br />of the custom resource definition. It is recommended to use<br />`spec.additionalScrapeConfigs` instead. |  |  |
| `serviceMonitorNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to match for ServicedMonitors discovery. An empty label selector<br />matches all namespaces. A null label selector (default value) matches the current<br />namespace only. |  |  |
| `podMonitorSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | PodMonitors to be selected for target discovery. An empty label selector<br />matches all objects. A null label selector matches no objects.<br />If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`<br />and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.<br />The Prometheus operator will ensure that the Prometheus configuration's<br />Secret exists, but it is the responsibility of the user to provide the raw<br />gzipped Prometheus configuration under the `prometheus.yaml.gz` key.<br />This behavior is *deprecated* and will be removed in the next major version<br />of the custom resource definition. It is recommended to use<br />`spec.additionalScrapeConfigs` instead. |  |  |
| `podMonitorNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to match for PodMonitors discovery. An empty label selector<br />matches all namespaces. A null label selector (default value) matches the current<br />namespace only. |  |  |
| `probeSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Probes to be selected for target discovery. An empty label selector<br />matches all objects. A null label selector matches no objects.<br />If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`<br />and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.<br />The Prometheus operator will ensure that the Prometheus configuration's<br />Secret exists, but it is the responsibility of the user to provide the raw<br />gzipped Prometheus configuration under the `prometheus.yaml.gz` key.<br />This behavior is *deprecated* and will be removed in the next major version<br />of the custom resource definition. It is recommended to use<br />`spec.additionalScrapeConfigs` instead. |  |  |
| `probeNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to match for Probe discovery. An empty label<br />selector matches all namespaces. A null label selector matches the<br />current namespace only. |  |  |
| `scrapeConfigSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | ScrapeConfigs to be selected for target discovery. An empty label<br />selector matches all objects. A null label selector matches no objects.<br />If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`<br />and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.<br />The Prometheus operator will ensure that the Prometheus configuration's<br />Secret exists, but it is the responsibility of the user to provide the raw<br />gzipped Prometheus configuration under the `prometheus.yaml.gz` key.<br />This behavior is *deprecated* and will be removed in the next major version<br />of the custom resource definition. It is recommended to use<br />`spec.additionalScrapeConfigs` instead.<br />Note that the ScrapeConfig custom resource definition is currently at Alpha level. |  |  |
| `scrapeConfigNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to match for ScrapeConfig discovery. An empty label selector<br />matches all namespaces. A null label selector matches the current<br />namespace only.<br />Note that the ScrapeConfig custom resource definition is currently at Alpha level. |  |  |
| `version` _string_ | Version of Prometheus being deployed. The operator uses this information<br />to generate the Prometheus StatefulSet + configuration files.<br />If not specified, the operator assumes the latest upstream version of<br />Prometheus available at the time when the version of the operator was<br />released. |  |  |
| `paused` _boolean_ | When a Prometheus deployment is paused, no actions except for deletion<br />will be performed on the underlying objects. |  |  |
| `image` _string_ | Container image name for Prometheus. If specified, it takes precedence<br />over the `spec.baseImage`, `spec.tag` and `spec.sha` fields.<br />Specifying `spec.version` is still necessary to ensure the Prometheus<br />Operator knows which version of Prometheus is being configured.<br />If neither `spec.image` nor `spec.baseImage` are defined, the operator<br />will use the latest upstream version of Prometheus available at the time<br />when the operator was released. |  |  |
| `imagePullPolicy` _[PullPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#pullpolicy-v1-core)_ | Image pull policy for the 'prometheus', 'init-config-reloader' and 'config-reloader' containers.<br />See https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy for more details. |  | Enum: [ Always Never IfNotPresent] <br /> |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#localobjectreference-v1-core) array_ | An optional list of references to Secrets in the same namespace<br />to use for pulling images from registries.<br />See http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod |  |  |
| `replicas` _integer_ | Number of replicas of each shard to deploy for a Prometheus deployment.<br />`spec.replicas` multiplied by `spec.shards` is the total number of Pods<br />created.<br />Default: 1 |  |  |
| `shards` _integer_ | Number of shards to distribute the scraped targets onto.<br />`spec.replicas` multiplied by `spec.shards` is the total number of Pods<br />being created.<br />When not defined, the operator assumes only one shard.<br />Note that scaling down shards will not reshard data onto the remaining<br />instances, it must be manually moved. Increasing shards will not reshard<br />data either but it will continue to be available from the same<br />instances. To query globally, use either<br />* Thanos sidecar + querier for query federation and Thanos Ruler for rules.<br />* Remote-write to send metrics to a central location.<br />By default, the sharding of targets is performed on:<br />* The `__address__` target's metadata label for PodMonitor,<br />ServiceMonitor and ScrapeConfig resources.<br />* The `__param_target__` label for Probe resources.<br />Users can define their own sharding implementation by setting the<br />`__tmp_hash` label during the target discovery with relabeling<br />configuration (either in the monitoring resources or via scrape class).<br />You can also disable sharding on a specific target by setting the<br />`__tmp_disable_sharding` label with relabeling configuration. When<br />the label value isn't empty, all Prometheus shards will scrape the target. |  |  |
| `replicaExternalLabelName` _string_ | Name of Prometheus external label used to denote the replica name.<br />The external label will _not_ be added when the field is set to the<br />empty string (`""`).<br />Default: "prometheus_replica" |  |  |
| `prometheusExternalLabelName` _string_ | Name of Prometheus external label used to denote the Prometheus instance<br />name. The external label will _not_ be added when the field is set to<br />the empty string (`""`).<br />Default: "prometheus" |  |  |
| `logLevel` _string_ | Log level for Prometheus and the config-reloader sidecar. |  | Enum: [ debug info warn error] <br /> |
| `logFormat` _string_ | Log format for Log level for Prometheus and the config-reloader sidecar. |  | Enum: [ logfmt json] <br /> |
| `scrapeInterval` _[Duration](#duration)_ | Interval between consecutive scrapes.<br />Default: "30s" | 30s | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `scrapeTimeout` _[Duration](#duration)_ | Number of seconds to wait until a scrape request times out.<br />The value cannot be greater than the scrape interval otherwise the operator will reject the resource. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `scrapeProtocols` _[ScrapeProtocol](#scrapeprotocol) array_ | The protocols to negotiate during a scrape. It tells clients the<br />protocols supported by Prometheus in order of preference (from most to least preferred).<br />If unset, Prometheus uses its default value.<br />It requires Prometheus >= v2.49.0.<br />`PrometheusText1.0.0` requires Prometheus >= v3.0.0. |  | Enum: [PrometheusProto OpenMetricsText0.0.1 OpenMetricsText1.0.0 PrometheusText0.0.4 PrometheusText1.0.0] <br /> |
| `externalLabels` _object (keys:string, values:string)_ | The labels to add to any time series or alerts when communicating with<br />external systems (federation, remote storage, Alertmanager).<br />Labels defined by `spec.replicaExternalLabelName` and<br />`spec.prometheusExternalLabelName` take precedence over this list. |  |  |
| `enableRemoteWriteReceiver` _boolean_ | Enable Prometheus to be used as a receiver for the Prometheus remote<br />write protocol.<br />WARNING: This is not considered an efficient way of ingesting samples.<br />Use it with caution for specific low-volume use cases.<br />It is not suitable for replacing the ingestion via scraping and turning<br />Prometheus into a push-based metrics collection system.<br />For more information see https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver<br />It requires Prometheus >= v2.33.0. |  |  |
| `enableOTLPReceiver` _boolean_ | Enable Prometheus to be used as a receiver for the OTLP Metrics protocol.<br />Note that the OTLP receiver endpoint is automatically enabled if `.spec.otlpConfig` is defined.<br />It requires Prometheus >= v2.47.0. |  |  |
| `remoteWriteReceiverMessageVersions` _[RemoteWriteMessageVersion](#remotewritemessageversion) array_ | List of the protobuf message versions to accept when receiving the<br />remote writes.<br />It requires Prometheus >= v2.54.0. |  | Enum: [V1.0 V2.0] <br />MinItems: 1 <br /> |
| `enableFeatures` _[EnableFeature](#enablefeature) array_ | Enable access to Prometheus feature flags. By default, no features are enabled.<br />Enabling features which are disabled by default is entirely outside the<br />scope of what the maintainers will support and by doing so, you accept<br />that this behaviour may break at any time without notice.<br />For more information see https://prometheus.io/docs/prometheus/latest/feature_flags/ |  | MinLength: 1 <br /> |
| `externalUrl` _string_ | The external URL under which the Prometheus service is externally<br />available. This is necessary to generate correct URLs (for instance if<br />Prometheus is accessible behind an Ingress resource). |  |  |
| `routePrefix` _string_ | The route prefix Prometheus registers HTTP handlers for.<br />This is useful when using `spec.externalURL`, and a proxy is rewriting<br />HTTP routes of a request, and the actual ExternalURL is still true, but<br />the server serves requests under a different route prefix. For example<br />for use with `kubectl proxy`. |  |  |
| `storage` _[StorageSpec](#storagespec)_ | Storage defines the storage used by Prometheus. |  |  |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volume-v1-core) array_ | Volumes allows the configuration of additional volumes on the output<br />StatefulSet definition. Volumes specified will be appended to other<br />volumes that are generated as a result of StorageSpec objects. |  |  |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volumemount-v1-core) array_ | VolumeMounts allows the configuration of additional VolumeMounts.<br />VolumeMounts will be appended to other VolumeMounts in the 'prometheus'<br />container, that are generated as a result of StorageSpec objects. |  |  |
| `persistentVolumeClaimRetentionPolicy` _[StatefulSetPersistentVolumeClaimRetentionPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#statefulsetpersistentvolumeclaimretentionpolicy-v1-apps)_ | The field controls if and how PVCs are deleted during the lifecycle of a StatefulSet.<br />The default behavior is all PVCs are retained.<br />This is an alpha field from kubernetes 1.23 until 1.26 and a beta field from 1.26.<br />It requires enabling the StatefulSetAutoDeletePVC feature gate. |  |  |
| `web` _[PrometheusWebSpec](#prometheuswebspec)_ | Defines the configuration of the Prometheus web server. |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#resourcerequirements-v1-core)_ | Defines the resources requests and limits of the 'prometheus' container. |  |  |
| `nodeSelector` _object (keys:string, values:string)_ | Defines on which Nodes the Pods are scheduled. |  |  |
| `serviceAccountName` _string_ | ServiceAccountName is the name of the ServiceAccount to use to run the<br />Prometheus Pods. |  |  |
| `automountServiceAccountToken` _boolean_ | AutomountServiceAccountToken indicates whether a service account token should be automatically mounted in the pod.<br />If the field isn't set, the operator mounts the service account token by default.<br />**Warning:** be aware that by default, Prometheus requires the service account token for Kubernetes service discovery.<br />It is possible to use strategic merge patch to project the service account token into the 'prometheus' container. |  |  |
| `secrets` _string array_ | Secrets is a list of Secrets in the same namespace as the Prometheus<br />object, which shall be mounted into the Prometheus Pods.<br />Each Secret is added to the StatefulSet definition as a volume named `secret-<secret-name>`.<br />The Secrets are mounted into /etc/prometheus/secrets/<secret-name> in the 'prometheus' container. |  |  |
| `configMaps` _string array_ | ConfigMaps is a list of ConfigMaps in the same namespace as the Prometheus<br />object, which shall be mounted into the Prometheus Pods.<br />Each ConfigMap is added to the StatefulSet definition as a volume named `configmap-<configmap-name>`.<br />The ConfigMaps are mounted into /etc/prometheus/configmaps/<configmap-name> in the 'prometheus' container. |  |  |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#affinity-v1-core)_ | Defines the Pods' affinity scheduling rules if specified. |  |  |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#toleration-v1-core) array_ | Defines the Pods' tolerations if specified. |  |  |
| `topologySpreadConstraints` _[TopologySpreadConstraint](#topologyspreadconstraint) array_ | Defines the pod's topology spread constraints if specified. |  |  |
| `remoteWrite` _[RemoteWriteSpec](#remotewritespec) array_ | Defines the list of remote write configurations. |  |  |
| `otlp` _[OTLPConfig](#otlpconfig)_ | Settings related to the OTLP receiver feature.<br />It requires Prometheus >= v2.55.0. |  |  |
| `securityContext` _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#podsecuritycontext-v1-core)_ | SecurityContext holds pod-level security attributes and common container settings.<br />This defaults to the default PodSecurityContext. |  |  |
| `dnsPolicy` _[DNSPolicy](#dnspolicy)_ | Defines the DNS policy for the pods. |  | Enum: [ClusterFirstWithHostNet ClusterFirst Default None] <br /> |
| `dnsConfig` _[PodDNSConfig](#poddnsconfig)_ | Defines the DNS configuration for the pods. |  |  |
| `listenLocal` _boolean_ | When true, the Prometheus server listens on the loopback address<br />instead of the Pod IP's address. |  |  |
| `enableServiceLinks` _boolean_ | Indicates whether information about services should be injected into pod's environment variables |  |  |
| `containers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core) array_ | Containers allows injecting additional containers or modifying operator<br />generated containers. This can be used to allow adding an authentication<br />proxy to the Pods or to change the behavior of an operator generated<br />container. Containers described here modify an operator generated<br />container if they share the same name and modifications are done via a<br />strategic merge patch.<br />The names of containers managed by the operator are:<br />* `prometheus`<br />* `config-reloader`<br />* `thanos-sidecar`<br />Overriding containers is entirely outside the scope of what the<br />maintainers will support and by doing so, you accept that this behaviour<br />may break at any time without notice. |  |  |
| `initContainers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core) array_ | InitContainers allows injecting initContainers to the Pod definition. Those<br />can be used to e.g.  fetch secrets for injection into the Prometheus<br />configuration from external sources. Any errors during the execution of<br />an initContainer will lead to a restart of the Pod. More info:<br />https://kubernetes.io/docs/concepts/workloads/pods/init-containers/<br />InitContainers described here modify an operator generated init<br />containers if they share the same name and modifications are done via a<br />strategic merge patch.<br />The names of init container name managed by the operator are:<br />* `init-config-reloader`.<br />Overriding init containers is entirely outside the scope of what the<br />maintainers will support and by doing so, you accept that this behaviour<br />may break at any time without notice. |  |  |
| `additionalScrapeConfigs` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | AdditionalScrapeConfigs allows specifying a key of a Secret containing<br />additional Prometheus scrape configurations. Scrape configurations<br />specified are appended to the configurations generated by the Prometheus<br />Operator. Job configurations specified must have the form as specified<br />in the official Prometheus documentation:<br />https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config.<br />As scrape configs are appended, the user is responsible to make sure it<br />is valid. Note that using this feature may expose the possibility to<br />break upgrades of Prometheus. It is advised to review Prometheus release<br />notes to ensure that no incompatible scrape configs are going to break<br />Prometheus after the upgrade. |  |  |
| `apiserverConfig` _[APIServerConfig](#apiserverconfig)_ | APIServerConfig allows specifying a host and auth methods to access the<br />Kuberntees API server.<br />If null, Prometheus is assumed to run inside of the cluster: it will<br />discover the API servers automatically and use the Pod's CA certificate<br />and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/. |  |  |
| `priorityClassName` _string_ | Priority class assigned to the Pods. |  |  |
| `portName` _string_ | Port name used for the pods and governing service.<br />Default: "web" | web |  |
| `arbitraryFSAccessThroughSMs` _[ArbitraryFSAccessThroughSMsConfig](#arbitraryfsaccessthroughsmsconfig)_ | When true, ServiceMonitor, PodMonitor and Probe object are forbidden to<br />reference arbitrary files on the file system of the 'prometheus'<br />container.<br />When a ServiceMonitor's endpoint specifies a `bearerTokenFile` value<br />(e.g.  '/var/run/secrets/kubernetes.io/serviceaccount/token'), a<br />malicious target can get access to the Prometheus service account's<br />token in the Prometheus' scrape request. Setting<br />`spec.arbitraryFSAccessThroughSM` to 'true' would prevent the attack.<br />Users should instead provide the credentials using the<br />`spec.bearerTokenSecret` field. |  |  |
| `overrideHonorLabels` _boolean_ | When true, Prometheus resolves label conflicts by renaming the labels in the scraped data<br /> to “exported_” for all targets created from ServiceMonitor, PodMonitor and<br />ScrapeConfig objects. Otherwise the HonorLabels field of the service or pod monitor applies.<br />In practice,`overrideHonorLaels:true` enforces `honorLabels:false`<br />for all ServiceMonitor, PodMonitor and ScrapeConfig objects. |  |  |
| `overrideHonorTimestamps` _boolean_ | When true, Prometheus ignores the timestamps for all the targets created<br />from service and pod monitors.<br />Otherwise the HonorTimestamps field of the service or pod monitor applies. |  |  |
| `ignoreNamespaceSelectors` _boolean_ | When true, `spec.namespaceSelector` from all PodMonitor, ServiceMonitor<br />and Probe objects will be ignored. They will only discover targets<br />within the namespace of the PodMonitor, ServiceMonitor and Probe<br />object. |  |  |
| `enforcedNamespaceLabel` _string_ | When not empty, a label will be added to:<br />1. All metrics scraped from `ServiceMonitor`, `PodMonitor`, `Probe` and `ScrapeConfig` objects.<br />2. All metrics generated from recording rules defined in `PrometheusRule` objects.<br />3. All alerts generated from alerting rules defined in `PrometheusRule` objects.<br />4. All vector selectors of PromQL expressions defined in `PrometheusRule` objects.<br />The label will not added for objects referenced in `spec.excludedFromEnforcement`.<br />The label's name is this field's value.<br />The label's value is the namespace of the `ServiceMonitor`,<br />`PodMonitor`, `Probe`, `PrometheusRule` or `ScrapeConfig` object. |  |  |
| `enforcedSampleLimit` _integer_ | When defined, enforcedSampleLimit specifies a global limit on the number<br />of scraped samples that will be accepted. This overrides any<br />`spec.sampleLimit` set by ServiceMonitor, PodMonitor, Probe objects<br />unless `spec.sampleLimit` is greater than zero and less than<br />`spec.enforcedSampleLimit`.<br />It is meant to be used by admins to keep the overall number of<br />samples/series under a desired limit.<br />When both `enforcedSampleLimit` and `sampleLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined sampleLimit value will inherit the global sampleLimit value (Prometheus >= 2.45.0) or the enforcedSampleLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedSampleLimit` is greater than the `sampleLimit`, the `sampleLimit` will be set to `enforcedSampleLimit`.<br />* Scrape objects with a sampleLimit value less than or equal to enforcedSampleLimit keep their specific value.<br />* Scrape objects with a sampleLimit value greater than enforcedSampleLimit are set to enforcedSampleLimit. |  |  |
| `enforcedTargetLimit` _integer_ | When defined, enforcedTargetLimit specifies a global limit on the number<br />of scraped targets. The value overrides any `spec.targetLimit` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.targetLimit` is<br />greater than zero and less than `spec.enforcedTargetLimit`.<br />It is meant to be used by admins to to keep the overall number of<br />targets under a desired limit.<br />When both `enforcedTargetLimit` and `targetLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined targetLimit value will inherit the global targetLimit value (Prometheus >= 2.45.0) or the enforcedTargetLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedTargetLimit` is greater than the `targetLimit`, the `targetLimit` will be set to `enforcedTargetLimit`.<br />* Scrape objects with a targetLimit value less than or equal to enforcedTargetLimit keep their specific value.<br />* Scrape objects with a targetLimit value greater than enforcedTargetLimit are set to enforcedTargetLimit. |  |  |
| `enforcedLabelLimit` _integer_ | When defined, enforcedLabelLimit specifies a global limit on the number<br />of labels per sample. The value overrides any `spec.labelLimit` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.labelLimit` is<br />greater than zero and less than `spec.enforcedLabelLimit`.<br />It requires Prometheus >= v2.27.0.<br />When both `enforcedLabelLimit` and `labelLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined labelLimit value will inherit the global labelLimit value (Prometheus >= 2.45.0) or the enforcedLabelLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedLabelLimit` is greater than the `labelLimit`, the `labelLimit` will be set to `enforcedLabelLimit`.<br />* Scrape objects with a labelLimit value less than or equal to enforcedLabelLimit keep their specific value.<br />* Scrape objects with a labelLimit value greater than enforcedLabelLimit are set to enforcedLabelLimit. |  |  |
| `enforcedLabelNameLengthLimit` _integer_ | When defined, enforcedLabelNameLengthLimit specifies a global limit on the length<br />of labels name per sample. The value overrides any `spec.labelNameLengthLimit` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.labelNameLengthLimit` is<br />greater than zero and less than `spec.enforcedLabelNameLengthLimit`.<br />It requires Prometheus >= v2.27.0.<br />When both `enforcedLabelNameLengthLimit` and `labelNameLengthLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined labelNameLengthLimit value will inherit the global labelNameLengthLimit value (Prometheus >= 2.45.0) or the enforcedLabelNameLengthLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedLabelNameLengthLimit` is greater than the `labelNameLengthLimit`, the `labelNameLengthLimit` will be set to `enforcedLabelNameLengthLimit`.<br />* Scrape objects with a labelNameLengthLimit value less than or equal to enforcedLabelNameLengthLimit keep their specific value.<br />* Scrape objects with a labelNameLengthLimit value greater than enforcedLabelNameLengthLimit are set to enforcedLabelNameLengthLimit. |  |  |
| `enforcedLabelValueLengthLimit` _integer_ | When not null, enforcedLabelValueLengthLimit defines a global limit on the length<br />of labels value per sample. The value overrides any `spec.labelValueLengthLimit` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.labelValueLengthLimit` is<br />greater than zero and less than `spec.enforcedLabelValueLengthLimit`.<br />It requires Prometheus >= v2.27.0.<br />When both `enforcedLabelValueLengthLimit` and `labelValueLengthLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined labelValueLengthLimit value will inherit the global labelValueLengthLimit value (Prometheus >= 2.45.0) or the enforcedLabelValueLengthLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedLabelValueLengthLimit` is greater than the `labelValueLengthLimit`, the `labelValueLengthLimit` will be set to `enforcedLabelValueLengthLimit`.<br />* Scrape objects with a labelValueLengthLimit value less than or equal to enforcedLabelValueLengthLimit keep their specific value.<br />* Scrape objects with a labelValueLengthLimit value greater than enforcedLabelValueLengthLimit are set to enforcedLabelValueLengthLimit. |  |  |
| `enforcedKeepDroppedTargets` _integer_ | When defined, enforcedKeepDroppedTargets specifies a global limit on the number of targets<br />dropped by relabeling that will be kept in memory. The value overrides<br />any `spec.keepDroppedTargets` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.keepDroppedTargets` is<br />greater than zero and less than `spec.enforcedKeepDroppedTargets`.<br />It requires Prometheus >= v2.47.0.<br />When both `enforcedKeepDroppedTargets` and `keepDroppedTargets` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined keepDroppedTargets value will inherit the global keepDroppedTargets value (Prometheus >= 2.45.0) or the enforcedKeepDroppedTargets value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedKeepDroppedTargets` is greater than the `keepDroppedTargets`, the `keepDroppedTargets` will be set to `enforcedKeepDroppedTargets`.<br />* Scrape objects with a keepDroppedTargets value less than or equal to enforcedKeepDroppedTargets keep their specific value.<br />* Scrape objects with a keepDroppedTargets value greater than enforcedKeepDroppedTargets are set to enforcedKeepDroppedTargets. |  |  |
| `enforcedBodySizeLimit` _[ByteSize](#bytesize)_ | When defined, enforcedBodySizeLimit specifies a global limit on the size<br />of uncompressed response body that will be accepted by Prometheus.<br />Targets responding with a body larger than this many bytes will cause<br />the scrape to fail.<br />It requires Prometheus >= v2.28.0.<br />When both `enforcedBodySizeLimit` and `bodySizeLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined bodySizeLimit value will inherit the global bodySizeLimit value (Prometheus >= 2.45.0) or the enforcedBodySizeLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedBodySizeLimit` is greater than the `bodySizeLimit`, the `bodySizeLimit` will be set to `enforcedBodySizeLimit`.<br />* Scrape objects with a bodySizeLimit value less than or equal to enforcedBodySizeLimit keep their specific value.<br />* Scrape objects with a bodySizeLimit value greater than enforcedBodySizeLimit are set to enforcedBodySizeLimit. |  | Pattern: `(^0\|([0-9]*[.])?[0-9]+((K\|M\|G\|T\|E\|P)i?)?B)$` <br /> |
| `nameValidationScheme` _[NameValidationSchemeOptions](#namevalidationschemeoptions)_ | Specifies the validation scheme for metric and label names.<br />It requires Prometheus >= v2.55.0. |  | Enum: [UTF8 Legacy] <br /> |
| `nameEscapingScheme` _[NameEscapingSchemeOptions](#nameescapingschemeoptions)_ | Specifies the character escaping scheme that will be requested when scraping<br />for metric and label names that do not conform to the legacy Prometheus<br />character set.<br />It requires Prometheus >= v3.4.0. |  | Enum: [AllowUTF8 Underscores Dots Values] <br /> |
| `convertClassicHistogramsToNHCB` _boolean_ | Whether to convert all scraped classic histograms into a native<br />histogram with custom buckets.<br />It requires Prometheus >= v3.4.0. |  |  |
| `minReadySeconds` _integer_ | Minimum number of seconds for which a newly created Pod should be ready<br />without any of its container crashing for it to be considered available.<br />Defaults to 0 (pod will be considered available as soon as it is ready)<br />This is an alpha field from kubernetes 1.22 until 1.24 which requires<br />enabling the StatefulSetMinReadySeconds feature gate. |  |  |
| `hostAliases` _[HostAlias](#hostalias) array_ | Optional list of hosts and IPs that will be injected into the Pod's<br />hosts file if specified. |  |  |
| `additionalArgs` _[Argument](#argument) array_ | AdditionalArgs allows setting additional arguments for the 'prometheus' container.<br />It is intended for e.g. activating hidden flags which are not supported by<br />the dedicated configuration options yet. The arguments are passed as-is to the<br />Prometheus container which may cause issues if they are invalid or not supported<br />by the given Prometheus version.<br />In case of an argument conflict (e.g. an argument which is already set by the<br />operator itself) or when providing an invalid argument, the reconciliation will<br />fail and an error will be logged. |  |  |
| `walCompression` _boolean_ | Configures compression of the write-ahead log (WAL) using Snappy.<br />WAL compression is enabled by default for Prometheus >= 2.20.0<br />Requires Prometheus v2.11.0 and above. |  |  |
| `excludedFromEnforcement` _[ObjectReference](#objectreference) array_ | List of references to PodMonitor, ServiceMonitor, Probe and PrometheusRule objects<br />to be excluded from enforcing a namespace label of origin.<br />It is only applicable if `spec.enforcedNamespaceLabel` set to true. |  |  |
| `hostNetwork` _boolean_ | Use the host's network namespace if true.<br />Make sure to understand the security implications if you want to enable<br />it (https://kubernetes.io/docs/concepts/configuration/overview/).<br />When hostNetwork is enabled, this will set the DNS policy to<br />`ClusterFirstWithHostNet` automatically (unless `.spec.DNSPolicy` is set<br />to a different value). |  |  |
| `podTargetLabels` _string array_ | PodTargetLabels are appended to the `spec.podTargetLabels` field of all<br />PodMonitor and ServiceMonitor objects. |  |  |
| `tracingConfig` _[PrometheusTracingConfig](#prometheustracingconfig)_ | TracingConfig configures tracing in Prometheus.<br />This is an *experimental feature*, it may change in any upcoming release<br />in a breaking way. |  |  |
| `bodySizeLimit` _[ByteSize](#bytesize)_ | BodySizeLimit defines per-scrape on response body size.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedBodySizeLimit. |  | Pattern: `(^0\|([0-9]*[.])?[0-9]+((K\|M\|G\|T\|E\|P)i?)?B)$` <br /> |
| `sampleLimit` _integer_ | SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedSampleLimit. |  |  |
| `targetLimit` _integer_ | TargetLimit defines a limit on the number of scraped targets that will be accepted.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedTargetLimit. |  |  |
| `labelLimit` _integer_ | Per-scrape limit on number of labels that will be accepted for a sample.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelLimit. |  |  |
| `labelNameLengthLimit` _integer_ | Per-scrape limit on length of labels name that will be accepted for a sample.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelNameLengthLimit. |  |  |
| `labelValueLengthLimit` _integer_ | Per-scrape limit on length of labels value that will be accepted for a sample.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelValueLengthLimit. |  |  |
| `keepDroppedTargets` _integer_ | Per-scrape limit on the number of targets dropped by relabeling<br />that will be kept in memory. 0 means no limit.<br />It requires Prometheus >= v2.47.0.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedKeepDroppedTargets. |  |  |
| `reloadStrategy` _[ReloadStrategyType](#reloadstrategytype)_ | Defines the strategy used to reload the Prometheus configuration.<br />If not specified, the configuration is reloaded using the /-/reload HTTP endpoint. |  | Enum: [HTTP ProcessSignal] <br /> |
| `maximumStartupDurationSeconds` _integer_ | Defines the maximum time that the `prometheus` container's startup probe will wait before being considered failed. The startup probe will return success after the WAL replay is complete.<br />If set, the value should be greater than 60 (seconds). Otherwise it will be equal to 600 seconds (15 minutes). |  | Minimum: 60 <br /> |
| `scrapeClasses` _[ScrapeClass](#scrapeclass) array_ | List of scrape classes to expose to scraping objects such as<br />PodMonitors, ServiceMonitors, Probes and ScrapeConfigs.<br />This is an *experimental feature*, it may change in any upcoming release<br />in a breaking way. |  |  |
| `serviceDiscoveryRole` _[ServiceDiscoveryRole](#servicediscoveryrole)_ | Defines the service discovery role used to discover targets from<br />`ServiceMonitor` objects and Alertmanager endpoints.<br />If set, the value should be either "Endpoints" or "EndpointSlice".<br />If unset, the operator assumes the "Endpoints" role. |  | Enum: [Endpoints EndpointSlice] <br /> |
| `tsdb` _[TSDBSpec](#tsdbspec)_ | Defines the runtime reloadable configuration of the timeseries database(TSDB).<br />It requires Prometheus >= v2.39.0 or PrometheusAgent >= v2.54.0. |  |  |
| `scrapeFailureLogFile` _string_ | File to which scrape failures are logged.<br />Reloading the configuration will reopen the file.<br />If the filename has an empty path, e.g. 'file.log', The Prometheus Pods<br />will mount the file into an emptyDir volume at `/var/log/prometheus`.<br />If a full path is provided, e.g. '/var/log/prometheus/file.log', you<br />must mount a volume in the specified directory and it must be writable.<br />It requires Prometheus >= v2.55.0. |  | MinLength: 1 <br /> |
| `serviceName` _string_ | The name of the service name used by the underlying StatefulSet(s) as the governing service.<br />If defined, the Service  must be created before the Prometheus/PrometheusAgent resource in the same namespace and it must define a selector that matches the pod labels.<br />If empty, the operator will create and manage a headless service named `prometheus-operated` for Prometheus resources,<br />or `prometheus-agent-operated` for PrometheusAgent resources.<br />When deploying multiple Prometheus/PrometheusAgent resources in the same namespace, it is recommended to specify a different value for each.<br />See https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id for more details. |  | MinLength: 1 <br /> |
| `runtime` _[RuntimeConfig](#runtimeconfig)_ | RuntimeConfig configures the values for the Prometheus process behavior |  |  |
| `terminationGracePeriodSeconds` _integer_ | Optional duration in seconds the pod needs to terminate gracefully.<br />Value must be non-negative integer. The value zero indicates stop immediately via<br />the kill signal (no opportunity to shut down) which may lead to data corruption.<br />Defaults to 600 seconds. |  | Minimum: 0 <br /> |
| `baseImage` _string_ | Deprecated: use 'spec.image' instead. |  |  |
| `tag` _string_ | Deprecated: use 'spec.image' instead. The image's tag can be specified as part of the image name. |  |  |
| `sha` _string_ | Deprecated: use 'spec.image' instead. The image's digest can be specified as part of the image name. |  |  |
| `retention` _[Duration](#duration)_ | How long to retain the Prometheus data.<br />Default: "24h" if `spec.retention` and `spec.retentionSize` are empty. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `retentionSize` _[ByteSize](#bytesize)_ | Maximum number of bytes used by the Prometheus data. |  | Pattern: `(^0\|([0-9]*[.])?[0-9]+((K\|M\|G\|T\|E\|P)i?)?B)$` <br /> |
| `shardRetentionPolicy` _[ShardRetentionPolicy](#shardretentionpolicy)_ | ShardRetentionPolicy defines the retention policy for the Prometheus shards.<br />(Alpha) Using this field requires the 'PrometheusShardRetentionPolicy' feature gate to be enabled.<br />The final goals for this feature can be seen at https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/proposals/202310-shard-autoscaling.md#graceful-scale-down-of-prometheus-servers,<br />however, the feature is not yet fully implemented in this PR. The limitation being:<br />* Retention duration is not settable, for now, shards are retained forever. |  |  |
| `disableCompaction` _boolean_ | When true, the Prometheus compaction is disabled.<br />When `spec.thanos.objectStorageConfig` or `spec.objectStorageConfigFile` are defined, the operator automatically<br />disables block compaction to avoid race conditions during block uploads (as the Thanos documentation recommends). |  |  |
| `rules` _[Rules](#rules)_ | Defines the configuration of the Prometheus rules' engine. |  |  |
| `prometheusRulesExcludedFromEnforce` _[PrometheusRuleExcludeConfig](#prometheusruleexcludeconfig) array_ | Defines the list of PrometheusRule objects to which the namespace label<br />enforcement doesn't apply.<br />This is only relevant when `spec.enforcedNamespaceLabel` is set to true.<br />Deprecated: use `spec.excludedFromEnforcement` instead. |  |  |
| `ruleSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | PrometheusRule objects to be selected for rule evaluation. An empty<br />label selector matches all objects. A null label selector matches no<br />objects. |  |  |
| `ruleNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to match for PrometheusRule discovery. An empty label selector<br />matches all namespaces. A null label selector matches the current<br />namespace only. |  |  |
| `query` _[QuerySpec](#queryspec)_ | QuerySpec defines the configuration of the Promethus query service. |  |  |
| `alerting` _[AlertingSpec](#alertingspec)_ | Defines the settings related to Alertmanager. |  |  |
| `additionalAlertRelabelConfigs` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | AdditionalAlertRelabelConfigs specifies a key of a Secret containing<br />additional Prometheus alert relabel configurations. The alert relabel<br />configurations are appended to the configuration generated by the<br />Prometheus Operator. They must be formatted according to the official<br />Prometheus documentation:<br />https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs<br />The user is responsible for making sure that the configurations are valid<br />Note that using this feature may expose the possibility to break<br />upgrades of Prometheus. It is advised to review Prometheus release notes<br />to ensure that no incompatible alert relabel configs are going to break<br />Prometheus after the upgrade. |  |  |
| `additionalAlertManagerConfigs` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | AdditionalAlertManagerConfigs specifies a key of a Secret containing<br />additional Prometheus Alertmanager configurations. The Alertmanager<br />configurations are appended to the configuration generated by the<br />Prometheus Operator. They must be formatted according to the official<br />Prometheus documentation:<br />https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config<br />The user is responsible for making sure that the configurations are valid<br />Note that using this feature may expose the possibility to break<br />upgrades of Prometheus. It is advised to review Prometheus release notes<br />to ensure that no incompatible AlertManager configs are going to break<br />Prometheus after the upgrade. |  |  |
| `remoteRead` _[RemoteReadSpec](#remotereadspec) array_ | Defines the list of remote read configurations. |  |  |
| `thanos` _[ThanosSpec](#thanosspec)_ | Defines the configuration of the optional Thanos sidecar. |  |  |
| `queryLogFile` _string_ | queryLogFile specifies where the file to which PromQL queries are logged.<br />If the filename has an empty path, e.g. 'query.log', The Prometheus Pods<br />will mount the file into an emptyDir volume at `/var/log/prometheus`.<br />If a full path is provided, e.g. '/var/log/prometheus/query.log', you<br />must mount a volume in the specified directory and it must be writable.<br />This is because the prometheus container runs with a read-only root<br />filesystem for security reasons.<br />Alternatively, the location can be set to a standard I/O stream, e.g.<br />`/dev/stdout`, to log query information to the default Prometheus log<br />stream. |  |  |
| `allowOverlappingBlocks` _boolean_ | AllowOverlappingBlocks enables vertical compaction and vertical query<br />merge in Prometheus.<br />Deprecated: this flag has no effect for Prometheus >= 2.39.0 where overlapping blocks are enabled by default. |  |  |
| `exemplars` _[Exemplars](#exemplars)_ | Exemplars related settings that are runtime reloadable.<br />It requires to enable the `exemplar-storage` feature flag to be effective. |  |  |
| `evaluationInterval` _[Duration](#duration)_ | Interval between rule evaluations.<br />Default: "30s" | 30s | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `ruleQueryOffset` _[Duration](#duration)_ | Defines the offset the rule evaluation timestamp of this particular group by the specified duration into the past.<br />It requires Prometheus >= v2.53.0. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `enableAdminAPI` _boolean_ | Enables access to the Prometheus web admin API.<br />WARNING: Enabling the admin APIs enables mutating endpoints, to delete data,<br />shutdown Prometheus, and more. Enabling this should be done with care and the<br />user is advised to add additional authentication authorization via a proxy to<br />ensure only clients authorized to perform these actions can do so.<br />For more information:<br />https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis |  |  |


#### PrometheusStatus



PrometheusStatus is the most recent observed status of the Prometheus cluster.
More info:
https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status



_Appears in:_
- [Prometheus](#prometheus)
- [PrometheusAgent](#prometheusagent)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `paused` _boolean_ | Represents whether any actions on the underlying managed objects are<br />being performed. Only delete actions will be performed. |  |  |
| `replicas` _integer_ | Total number of non-terminated pods targeted by this Prometheus deployment<br />(their labels match the selector). |  |  |
| `updatedReplicas` _integer_ | Total number of non-terminated pods targeted by this Prometheus deployment<br />that have the desired version spec. |  |  |
| `availableReplicas` _integer_ | Total number of available pods (ready for at least minReadySeconds)<br />targeted by this Prometheus deployment. |  |  |
| `unavailableReplicas` _integer_ | Total number of unavailable pods targeted by this Prometheus deployment. |  |  |
| `conditions` _[Condition](#condition) array_ | The current state of the Prometheus deployment. |  |  |
| `shardStatuses` _[ShardStatus](#shardstatus) array_ | The list has one entry per shard. Each entry provides a summary of the shard status. |  |  |
| `shards` _integer_ | Shards is the most recently observed number of shards. |  |  |
| `selector` _string_ | The selector used to match the pods targeted by this Prometheus resource. |  |  |


#### PrometheusTracingConfig







_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `clientType` _string_ | Client used to export the traces. Supported values are `http` or `grpc`. |  | Enum: [http grpc] <br /> |
| `endpoint` _string_ | Endpoint to send the traces to. Should be provided in format <host>:<port>. |  | MinLength: 1 <br /> |
| `samplingFraction` _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#quantity-resource-api)_ | Sets the probability a given trace will be sampled. Must be a float from 0 through 1. |  |  |
| `insecure` _boolean_ | If disabled, the client will use a secure connection. |  |  |
| `headers` _object (keys:string, values:string)_ | Key-value pairs to be used as headers associated with gRPC or HTTP requests. |  |  |
| `compression` _string_ | Compression key for supported compression types. The only supported value is `gzip`. |  | Enum: [gzip] <br /> |
| `timeout` _[Duration](#duration)_ | Maximum time the exporter will wait for each batch export. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `tlsConfig` _[TLSConfig](#tlsconfig)_ | TLS Config to use when sending traces. |  |  |


#### PrometheusWebSpec



PrometheusWebSpec defines the configuration of the Prometheus web server.



_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `tlsConfig` _[WebTLSConfig](#webtlsconfig)_ | Defines the TLS parameters for HTTPS. |  |  |
| `httpConfig` _[WebHTTPConfig](#webhttpconfig)_ | Defines HTTP parameters for web server. |  |  |
| `pageTitle` _string_ | The prometheus web page title. |  |  |
| `maxConnections` _integer_ | Defines the maximum number of simultaneous connections<br />A zero value means that Prometheus doesn't accept any incoming connection. |  | Minimum: 0 <br /> |


#### ProxyConfig

_Underlying type:_ _[struct{ProxyURL *string "json:\"proxyUrl,omitempty\""; NoProxy *string "json:\"noProxy,omitempty\""; ProxyFromEnvironment *bool "json:\"proxyFromEnvironment,omitempty\""; ProxyConnectHeader map[string][]k8s.io/api/core/v1.SecretKeySelector "json:\"proxyConnectHeader,omitempty\""}](#struct{proxyurl-*string-"json:\"proxyurl,omitempty\"";-noproxy-*string-"json:\"noproxy,omitempty\"";-proxyfromenvironment-*bool-"json:\"proxyfromenvironment,omitempty\"";-proxyconnectheader-map[string][]k8sioapicorev1secretkeyselector-"json:\"proxyconnectheader,omitempty\""})_





_Appears in:_
- [AlertmanagerEndpoints](#alertmanagerendpoints)
- [AzureSDConfig](#azuresdconfig)
- [ConsulSDConfig](#consulsdconfig)
- [DigitalOceanSDConfig](#digitaloceansdconfig)
- [DockerSDConfig](#dockersdconfig)
- [DockerSwarmSDConfig](#dockerswarmsdconfig)
- [EC2SDConfig](#ec2sdconfig)
- [EurekaSDConfig](#eurekasdconfig)
- [HTTPConfig](#httpconfig)
- [HTTPConfig](#httpconfig)
- [HTTPConfig](#httpconfig)
- [HTTPSDConfig](#httpsdconfig)
- [HetznerSDConfig](#hetznersdconfig)
- [IonosSDConfig](#ionossdconfig)
- [KubernetesSDConfig](#kubernetessdconfig)
- [KumaSDConfig](#kumasdconfig)
- [LightSailSDConfig](#lightsailsdconfig)
- [LinodeSDConfig](#linodesdconfig)
- [NomadSDConfig](#nomadsdconfig)
- [OAuth2](#oauth2)
- [PuppetDBSDConfig](#puppetdbsdconfig)
- [RemoteReadSpec](#remotereadspec)
- [RemoteWriteSpec](#remotewritespec)
- [ScalewaySDConfig](#scalewaysdconfig)
- [ScrapeConfigSpec](#scrapeconfigspec)



#### QuerySpec



QuerySpec defines the query command line flags when starting Prometheus.



_Appears in:_
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `lookbackDelta` _string_ | The delta difference allowed for retrieving metrics during expression evaluations. |  |  |
| `maxConcurrency` _integer_ | Number of concurrent queries that can be run at once. |  | Minimum: 1 <br /> |
| `maxSamples` _integer_ | Maximum number of samples a single query can load into memory. Note that<br />queries will fail if they would load more samples than this into memory,<br />so this also limits the number of samples a query can return. |  |  |
| `timeout` _[Duration](#duration)_ | Maximum time a query may take before being aborted. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |


#### QueueConfig



QueueConfig allows the tuning of remote write's queue_config parameters.
This object is referenced in the RemoteWriteSpec object.



_Appears in:_
- [RemoteWriteSpec](#remotewritespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `capacity` _integer_ | Capacity is the number of samples to buffer per shard before we start<br />dropping them. |  |  |
| `minShards` _integer_ | MinShards is the minimum number of shards, i.e. amount of concurrency. |  |  |
| `maxShards` _integer_ | MaxShards is the maximum number of shards, i.e. amount of concurrency. |  |  |
| `maxSamplesPerSend` _integer_ | MaxSamplesPerSend is the maximum number of samples per send. |  |  |
| `batchSendDeadline` _[Duration](#duration)_ | BatchSendDeadline is the maximum time a sample will wait in buffer. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `maxRetries` _integer_ | MaxRetries is the maximum number of times to retry a batch on recoverable errors. |  |  |
| `minBackoff` _[Duration](#duration)_ | MinBackoff is the initial retry delay. Gets doubled for every retry. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `maxBackoff` _[Duration](#duration)_ | MaxBackoff is the maximum retry delay. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `retryOnRateLimit` _boolean_ | Retry upon receiving a 429 status code from the remote-write storage.<br />This is an *experimental feature*, it may change in any upcoming release<br />in a breaking way. |  |  |
| `sampleAgeLimit` _[Duration](#duration)_ | SampleAgeLimit drops samples older than the limit.<br />It requires Prometheus >= v2.50.0 or Thanos >= v0.32.0. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |


#### RelabelConfig



RelabelConfig allows dynamic rewriting of the label set for targets, alerts,
scraped samples and remote write samples.

More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config



_Appears in:_
- [AlertmanagerEndpoints](#alertmanagerendpoints)
- [Endpoint](#endpoint)
- [PodMetricsEndpoint](#podmetricsendpoint)
- [ProbeSpec](#probespec)
- [ProbeTargetIngress](#probetargetingress)
- [ProbeTargetStaticConfig](#probetargetstaticconfig)
- [RemoteWriteSpec](#remotewritespec)
- [ScrapeClass](#scrapeclass)
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sourceLabels` _[LabelName](#labelname) array_ | The source labels select values from existing labels. Their content is<br />concatenated using the configured Separator and matched against the<br />configured regular expression. |  | Pattern: `^[a-zA-Z_][a-zA-Z0-9_]*$` <br /> |
| `separator` _string_ | Separator is the string between concatenated SourceLabels. |  |  |
| `targetLabel` _string_ | Label to which the resulting string is written in a replacement.<br />It is mandatory for `Replace`, `HashMod`, `Lowercase`, `Uppercase`,<br />`KeepEqual` and `DropEqual` actions.<br />Regex capture groups are available. |  |  |
| `regex` _string_ | Regular expression against which the extracted value is matched. |  |  |
| `modulus` _integer_ | Modulus to take of the hash of the source label values.<br />Only applicable when the action is `HashMod`. |  |  |
| `replacement` _string_ | Replacement value against which a Replace action is performed if the<br />regular expression matches.<br />Regex capture groups are available. |  |  |
| `action` _string_ | Action to perform based on the regex matching.<br />`Uppercase` and `Lowercase` actions require Prometheus >= v2.36.0.<br />`DropEqual` and `KeepEqual` actions require Prometheus >= v2.41.0.<br />Default: "Replace" | replace | Enum: [replace Replace keep Keep drop Drop hashmod HashMod labelmap LabelMap labeldrop LabelDrop labelkeep LabelKeep lowercase Lowercase uppercase Uppercase keepequal KeepEqual dropequal DropEqual] <br /> |


#### ReloadStrategyType

_Underlying type:_ _string_



_Validation:_
- Enum: [HTTP ProcessSignal]

_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)

| Field | Description |
| --- | --- |
| `HTTP` | HTTPReloadStrategyType reloads the configuration using the /-/reload HTTP endpoint.<br /> |
| `ProcessSignal` | ProcessSignalReloadStrategyType reloads the configuration by sending a SIGHUP signal to the process.<br /> |


#### RemoteReadSpec



RemoteReadSpec defines the configuration for Prometheus to read back samples
from a remote endpoint.



_Appears in:_
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `url` _string_ | The URL of the endpoint to query from. |  |  |
| `name` _string_ | The name of the remote read queue, it must be unique if specified. The<br />name is used in metrics and logging in order to differentiate read<br />configurations.<br />It requires Prometheus >= v2.15.0. |  |  |
| `requiredMatchers` _object (keys:string, values:string)_ | An optional list of equality matchers which have to be present<br />in a selector to query the remote read endpoint. |  |  |
| `remoteTimeout` _[Duration](#duration)_ | Timeout for requests to the remote read endpoint. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `headers` _object (keys:string, values:string)_ | Custom HTTP headers to be sent along with each remote read request.<br />Be aware that headers that are set by Prometheus itself can't be overwritten.<br />Only valid in Prometheus versions 2.26.0 and newer. |  |  |
| `readRecent` _boolean_ | Whether reads should be made for queries for time ranges that<br />the local storage should have complete data for. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | OAuth2 configuration for the URL.<br />It requires Prometheus >= v2.27.0.<br />Cannot be set at the same time as `authorization`, or `basicAuth`. |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth configuration for the URL.<br />Cannot be set at the same time as `authorization`, or `oauth2`. |  |  |
| `bearerTokenFile` _string_ | File from which to read the bearer token for the URL.<br />Deprecated: this will be removed in a future release. Prefer using `authorization`. |  |  |
| `authorization` _[Authorization](#authorization)_ | Authorization section for the URL.<br />It requires Prometheus >= v2.26.0.<br />Cannot be set at the same time as `basicAuth`, or `oauth2`. |  |  |
| `bearerToken` _string_ | *Warning: this field shouldn't be used because the token value appears<br />in clear-text. Prefer using `authorization`.*<br />Deprecated: this will be removed in a future release. |  |  |
| `tlsConfig` _[TLSConfig](#tlsconfig)_ | TLS Config to use for the URL. |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects.<br />It requires Prometheus >= v2.26.0. |  |  |
| `filterExternalLabels` _boolean_ | Whether to use the external labels as selectors for the remote read endpoint.<br />It requires Prometheus >= v2.34.0. |  |  |


#### RemoteWriteMessageVersion

_Underlying type:_ _string_



_Validation:_
- Enum: [V1.0 V2.0]

_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)
- [RemoteWriteSpec](#remotewritespec)



#### RemoteWriteSpec



RemoteWriteSpec defines the configuration to write samples from Prometheus
to a remote endpoint.



_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)
- [ThanosRulerSpec](#thanosrulerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `url` _string_ | The URL of the endpoint to send samples to. |  | MinLength: 1 <br /> |
| `name` _string_ | The name of the remote write queue, it must be unique if specified. The<br />name is used in metrics and logging in order to differentiate queues.<br />It requires Prometheus >= v2.15.0 or Thanos >= 0.24.0. |  |  |
| `messageVersion` _[RemoteWriteMessageVersion](#remotewritemessageversion)_ | The Remote Write message's version to use when writing to the endpoint.<br />`Version1.0` corresponds to the `prometheus.WriteRequest` protobuf message introduced in Remote Write 1.0.<br />`Version2.0` corresponds to the `io.prometheus.write.v2.Request` protobuf message introduced in Remote Write 2.0.<br />When `Version2.0` is selected, Prometheus will automatically be<br />configured to append the metadata of scraped metrics to the WAL.<br />Before setting this field, consult with your remote storage provider<br />what message version it supports.<br />It requires Prometheus >= v2.54.0 or Thanos >= v0.37.0. |  | Enum: [V1.0 V2.0] <br /> |
| `sendExemplars` _boolean_ | Enables sending of exemplars over remote write. Note that<br />exemplar-storage itself must be enabled using the `spec.enableFeatures`<br />option for exemplars to be scraped in the first place.<br />It requires Prometheus >= v2.27.0 or Thanos >= v0.24.0. |  |  |
| `sendNativeHistograms` _boolean_ | Enables sending of native histograms, also known as sparse histograms<br />over remote write.<br />It requires Prometheus >= v2.40.0 or Thanos >= v0.30.0. |  |  |
| `remoteTimeout` _[Duration](#duration)_ | Timeout for requests to the remote write endpoint. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `headers` _object (keys:string, values:string)_ | Custom HTTP headers to be sent along with each remote write request.<br />Be aware that headers that are set by Prometheus itself can't be overwritten.<br />It requires Prometheus >= v2.25.0 or Thanos >= v0.24.0. |  |  |
| `writeRelabelConfigs` _[RelabelConfig](#relabelconfig) array_ | The list of remote write relabel configurations. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | OAuth2 configuration for the URL.<br />It requires Prometheus >= v2.27.0 or Thanos >= v0.24.0.<br />Cannot be set at the same time as `sigv4`, `authorization`, `basicAuth`, or `azureAd`. |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth configuration for the URL.<br />Cannot be set at the same time as `sigv4`, `authorization`, `oauth2`, or `azureAd`. |  |  |
| `bearerTokenFile` _string_ | File from which to read bearer token for the URL.<br />Deprecated: this will be removed in a future release. Prefer using `authorization`. |  |  |
| `authorization` _[Authorization](#authorization)_ | Authorization section for the URL.<br />It requires Prometheus >= v2.26.0 or Thanos >= v0.24.0.<br />Cannot be set at the same time as `sigv4`, `basicAuth`, `oauth2`, or `azureAd`. |  |  |
| `sigv4` _[Sigv4](#sigv4)_ | Sigv4 allows to configures AWS's Signature Verification 4 for the URL.<br />It requires Prometheus >= v2.26.0 or Thanos >= v0.24.0.<br />Cannot be set at the same time as `authorization`, `basicAuth`, `oauth2`, or `azureAd`. |  |  |
| `azureAd` _[AzureAD](#azuread)_ | AzureAD for the URL.<br />It requires Prometheus >= v2.45.0 or Thanos >= v0.31.0.<br />Cannot be set at the same time as `authorization`, `basicAuth`, `oauth2`, or `sigv4`. |  |  |
| `bearerToken` _string_ | *Warning: this field shouldn't be used because the token value appears<br />in clear-text. Prefer using `authorization`.*<br />Deprecated: this will be removed in a future release. |  |  |
| `tlsConfig` _[TLSConfig](#tlsconfig)_ | TLS Config to use for the URL. |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects.<br />It requires Prometheus >= v2.26.0 or Thanos >= v0.24.0. |  |  |
| `queueConfig` _[QueueConfig](#queueconfig)_ | QueueConfig allows tuning of the remote write queue parameters. |  |  |
| `metadataConfig` _[MetadataConfig](#metadataconfig)_ | MetadataConfig configures the sending of series metadata to the remote storage. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2. |  |  |
| `roundRobinDNS` _boolean_ | When enabled:<br />    - The remote-write mechanism will resolve the hostname via DNS.<br />    - It will randomly select one of the resolved IP addresses and connect to it.<br />When disabled (default behavior):<br />    - The Go standard library will handle hostname resolution.<br />    - It will attempt connections to each resolved IP address sequentially.<br />Note: The connection timeout applies to the entire resolution and connection process.<br />      If disabled, the timeout is distributed across all connection attempts.<br />It requires Prometheus >= v3.1.0 or Thanos >= v0.38.0. |  |  |


#### RetainConfig







_Appears in:_
- [ShardRetentionPolicy](#shardretentionpolicy)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `retentionPeriod` _[Duration](#duration)_ |  |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |


#### Rule



Rule describes an alerting or recording rule
See Prometheus documentation: [alerting](https://www.prometheus.io/docs/prometheus/latest/configuration/alerting_rules/) or [recording](https://www.prometheus.io/docs/prometheus/latest/configuration/recording_rules/#recording-rules) rule



_Appears in:_
- [RuleGroup](#rulegroup)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `record` _string_ | Name of the time series to output to. Must be a valid metric name.<br />Only one of `record` and `alert` must be set. |  |  |
| `alert` _string_ | Name of the alert. Must be a valid label value.<br />Only one of `record` and `alert` must be set. |  |  |
| `expr` _[IntOrString](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#intorstring-intstr-util)_ | PromQL expression to evaluate. |  |  |
| `for` _[Duration](#duration)_ | Alerts are considered firing once they have been returned for this long. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `keep_firing_for` _[NonEmptyDuration](#nonemptyduration)_ | KeepFiringFor defines how long an alert will continue firing after the condition that triggered it has cleared. |  | MinLength: 1 <br />Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `labels` _object (keys:string, values:string)_ | Labels to add or overwrite. |  |  |
| `annotations` _object (keys:string, values:string)_ | Annotations to add to each alert.<br />Only valid for alerting rules. |  |  |


#### RuleGroup



RuleGroup is a list of sequentially evaluated recording and alerting rules.



_Appears in:_
- [PrometheusRuleSpec](#prometheusrulespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the rule group. |  | MinLength: 1 <br /> |
| `labels` _object (keys:string, values:string)_ | Labels to add or overwrite before storing the result for its rules.<br />The labels defined at the rule level take precedence.<br />It requires Prometheus >= 3.0.0.<br />The field is ignored for Thanos Ruler. |  |  |
| `interval` _[Duration](#duration)_ | Interval determines how often rules in the group are evaluated. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `query_offset` _[Duration](#duration)_ | Defines the offset the rule evaluation timestamp of this particular group by the specified duration into the past.<br />It requires Prometheus >= v2.53.0.<br />It is not supported for ThanosRuler. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `rules` _[Rule](#rule) array_ | List of alerting and recording rules. |  |  |
| `partial_response_strategy` _string_ | PartialResponseStrategy is only used by ThanosRuler and will<br />be ignored by Prometheus instances.<br />More info: https://github.com/thanos-io/thanos/blob/main/docs/components/rule.md#partial-response |  | Pattern: `^(?i)(abort\|warn)?$` <br /> |
| `limit` _integer_ | Limit the number of alerts an alerting rule and series a recording<br />rule can produce.<br />Limit is supported starting with Prometheus >= 2.31 and Thanos Ruler >= 0.24. |  |  |


#### Rules







_Appears in:_
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `alert` _[RulesAlert](#rulesalert)_ | Defines the parameters of the Prometheus rules' engine.<br />Any update to these parameters trigger a restart of the pods. |  |  |


#### RulesAlert







_Appears in:_
- [Rules](#rules)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `forOutageTolerance` _string_ | Max time to tolerate prometheus outage for restoring 'for' state of<br />alert. |  |  |
| `forGracePeriod` _string_ | Minimum duration between alert and restored 'for' state.<br />This is maintained only for alerts with a configured 'for' time greater<br />than the grace period. |  |  |
| `resendDelay` _string_ | Minimum amount of time to wait before resending an alert to<br />Alertmanager. |  |  |


#### RuntimeConfig



RuntimeConfig configures the values for the process behavior.



_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `goGC` _integer_ | The Go garbage collection target percentage. Lowering this number may increase the CPU usage.<br />See: https://tip.golang.org/doc/gc-guide#GOGC |  | Minimum: -1 <br /> |


#### SafeAuthorization



SafeAuthorization specifies a subset of the Authorization struct, that is
safe for use because it doesn't provide access to the Prometheus container's
filesystem.



_Appears in:_
- [AlertmanagerEndpoints](#alertmanagerendpoints)
- [Authorization](#authorization)
- [AzureSDConfig](#azuresdconfig)
- [ConsulSDConfig](#consulsdconfig)
- [DigitalOceanSDConfig](#digitaloceansdconfig)
- [DockerSDConfig](#dockersdconfig)
- [DockerSwarmSDConfig](#dockerswarmsdconfig)
- [Endpoint](#endpoint)
- [EurekaSDConfig](#eurekasdconfig)
- [HTTPConfig](#httpconfig)
- [HTTPConfig](#httpconfig)
- [HTTPConfig](#httpconfig)
- [HTTPSDConfig](#httpsdconfig)
- [HetznerSDConfig](#hetznersdconfig)
- [IonosSDConfig](#ionossdconfig)
- [KubernetesSDConfig](#kubernetessdconfig)
- [KumaSDConfig](#kumasdconfig)
- [LightSailSDConfig](#lightsailsdconfig)
- [LinodeSDConfig](#linodesdconfig)
- [NomadSDConfig](#nomadsdconfig)
- [PodMetricsEndpoint](#podmetricsendpoint)
- [ProbeSpec](#probespec)
- [PuppetDBSDConfig](#puppetdbsdconfig)
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _string_ | Defines the authentication type. The value is case-insensitive.<br />"Basic" is not a supported value.<br />Default: "Bearer" |  |  |
| `credentials` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Selects a key of a Secret in the namespace that contains the credentials for authentication. |  |  |


#### SafeTLSConfig



SafeTLSConfig specifies safe TLS configuration parameters.



_Appears in:_
- [AzureSDConfig](#azuresdconfig)
- [ClusterTLSConfig](#clustertlsconfig)
- [ConsulSDConfig](#consulsdconfig)
- [DigitalOceanSDConfig](#digitaloceansdconfig)
- [DockerSDConfig](#dockersdconfig)
- [DockerSwarmSDConfig](#dockerswarmsdconfig)
- [EC2SDConfig](#ec2sdconfig)
- [EmailConfig](#emailconfig)
- [EmailConfig](#emailconfig)
- [EurekaSDConfig](#eurekasdconfig)
- [GlobalSMTPConfig](#globalsmtpconfig)
- [HTTPConfig](#httpconfig)
- [HTTPConfig](#httpconfig)
- [HTTPConfig](#httpconfig)
- [HTTPSDConfig](#httpsdconfig)
- [HetznerSDConfig](#hetznersdconfig)
- [IonosSDConfig](#ionossdconfig)
- [KubernetesSDConfig](#kubernetessdconfig)
- [KumaSDConfig](#kumasdconfig)
- [LightSailSDConfig](#lightsailsdconfig)
- [LinodeSDConfig](#linodesdconfig)
- [NomadSDConfig](#nomadsdconfig)
- [OAuth2](#oauth2)
- [OpenStackSDConfig](#openstacksdconfig)
- [PodMetricsEndpoint](#podmetricsendpoint)
- [ProbeSpec](#probespec)
- [PuppetDBSDConfig](#puppetdbsdconfig)
- [ScalewaySDConfig](#scalewaysdconfig)
- [ScrapeConfigSpec](#scrapeconfigspec)
- [TLSConfig](#tlsconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `ca` _[SecretOrConfigMap](#secretorconfigmap)_ | Certificate authority used when verifying server certificates. |  |  |
| `cert` _[SecretOrConfigMap](#secretorconfigmap)_ | Client certificate to present when doing client-authentication. |  |  |
| `keySecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Secret containing the client key file for the targets. |  |  |
| `serverName` _string_ | Used to verify the hostname for the targets. |  |  |
| `insecureSkipVerify` _boolean_ | Disable target certificate validation. |  |  |
| `minVersion` _[TLSVersion](#tlsversion)_ | Minimum acceptable TLS version.<br />It requires Prometheus >= v2.35.0 or Thanos >= v0.28.0. |  | Enum: [TLS10 TLS11 TLS12 TLS13] <br /> |
| `maxVersion` _[TLSVersion](#tlsversion)_ | Maximum acceptable TLS version.<br />It requires Prometheus >= v2.41.0 or Thanos >= v0.31.0. |  | Enum: [TLS10 TLS11 TLS12 TLS13] <br /> |


#### ScrapeClass







_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the scrape class. |  | MinLength: 1 <br /> |
| `default` _boolean_ | Default indicates that the scrape applies to all scrape objects that<br />don't configure an explicit scrape class name.<br />Only one scrape class can be set as the default. |  |  |
| `fallbackScrapeProtocol` _[ScrapeProtocol](#scrapeprotocol)_ | The protocol to use if a scrape returns blank, unparseable, or otherwise invalid Content-Type.<br />It will only apply if the scrape resource doesn't specify any FallbackScrapeProtocol<br />It requires Prometheus >= v3.0.0. |  | Enum: [PrometheusProto OpenMetricsText0.0.1 OpenMetricsText1.0.0 PrometheusText0.0.4 PrometheusText1.0.0] <br /> |
| `tlsConfig` _[TLSConfig](#tlsconfig)_ | TLSConfig defines the TLS settings to use for the scrape. When the<br />scrape objects define their own CA, certificate and/or key, they take<br />precedence over the corresponding scrape class fields.<br />For now only the `caFile`, `certFile` and `keyFile` fields are supported. |  |  |
| `authorization` _[Authorization](#authorization)_ | Authorization section for the ScrapeClass.<br />It will only apply if the scrape resource doesn't specify any Authorization. |  |  |
| `relabelings` _[RelabelConfig](#relabelconfig) array_ | Relabelings configures the relabeling rules to apply to all scrape targets.<br />The Operator automatically adds relabelings for a few standard Kubernetes fields<br />like `__meta_kubernetes_namespace` and `__meta_kubernetes_service_name`.<br />Then the Operator adds the scrape class relabelings defined here.<br />Then the Operator adds the target-specific relabelings defined in the scrape object.<br />More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config |  |  |
| `metricRelabelings` _[RelabelConfig](#relabelconfig) array_ | MetricRelabelings configures the relabeling rules to apply to all samples before ingestion.<br />The Operator adds the scrape class metric relabelings defined here.<br />Then the Operator adds the target-specific metric relabelings defined in ServiceMonitors, PodMonitors, Probes and ScrapeConfigs.<br />Then the Operator adds namespace enforcement relabeling rule, specified in '.spec.enforcedNamespaceLabel'.<br />More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#metric_relabel_configs |  |  |
| `attachMetadata` _[AttachMetadata](#attachmetadata)_ | AttachMetadata configures additional metadata to the discovered targets.<br />When the scrape object defines its own configuration, it takes<br />precedence over the scrape class configuration. |  |  |


#### ScrapeProtocol

_Underlying type:_ _string_

ScrapeProtocol represents a protocol used by Prometheus for scraping metrics.
Supported values are:
* `OpenMetricsText0.0.1`
* `OpenMetricsText1.0.0`
* `PrometheusProto`
* `PrometheusText0.0.4`
* `PrometheusText1.0.0`

_Validation:_
- Enum: [PrometheusProto OpenMetricsText0.0.1 OpenMetricsText1.0.0 PrometheusText0.0.4 PrometheusText1.0.0]

_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PodMonitorSpec](#podmonitorspec)
- [ProbeSpec](#probespec)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)
- [ScrapeClass](#scrapeclass)
- [ScrapeConfigSpec](#scrapeconfigspec)
- [ServiceMonitorSpec](#servicemonitorspec)

| Field | Description |
| --- | --- |
| `PrometheusProto` |  |
| `PrometheusText0.0.4` |  |
| `PrometheusText1.0.0` |  |
| `OpenMetricsText0.0.1` |  |
| `OpenMetricsText1.0.0` |  |


#### SecretOrConfigMap



SecretOrConfigMap allows to specify data as a Secret or ConfigMap. Fields are mutually exclusive.



_Appears in:_
- [AlertmanagerConfiguration](#alertmanagerconfiguration)
- [OAuth2](#oauth2)
- [SafeTLSConfig](#safetlsconfig)
- [TLSConfig](#tlsconfig)
- [WebTLSConfig](#webtlsconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Secret containing data to use for the targets. |  |  |
| `configMap` _[ConfigMapKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#configmapkeyselector-v1-core)_ | ConfigMap containing data to use for the targets. |  |  |


#### SelectorMechanism

_Underlying type:_ _string_



_Validation:_
- Enum: [RelabelConfig RoleSelector]

_Appears in:_
- [PodMonitorSpec](#podmonitorspec)
- [ServiceMonitorSpec](#servicemonitorspec)

| Field | Description |
| --- | --- |
| `RelabelConfig` |  |
| `RoleSelector` |  |


#### ServiceDiscoveryRole

_Underlying type:_ _string_



_Validation:_
- Enum: [Endpoints EndpointSlice]

_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)

| Field | Description |
| --- | --- |
| `Endpoints` |  |
| `EndpointSlice` |  |


#### ServiceMonitor



The `ServiceMonitor` custom resource definition (CRD) defines how `Prometheus` and `PrometheusAgent` can scrape metrics from a group of services.
Among other things, it allows to specify:
* The services to scrape via label selectors.
* The container ports to scrape.
* Authentication credentials to use.
* Target and metric relabeling.

`Prometheus` and `PrometheusAgent` objects select `ServiceMonitor` objects using label and namespace selectors.



_Appears in:_
- [ServiceMonitorList](#servicemonitorlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[ServiceMonitorSpec](#servicemonitorspec)_ | Specification of desired Service selection for target discovery by<br />Prometheus. |  |  |




#### ServiceMonitorSpec



ServiceMonitorSpec defines the specification parameters for a ServiceMonitor.



_Appears in:_
- [ServiceMonitor](#servicemonitor)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `jobLabel` _string_ | `jobLabel` selects the label from the associated Kubernetes `Service`<br />object which will be used as the `job` label for all metrics.<br />For example if `jobLabel` is set to `foo` and the Kubernetes `Service`<br />object is labeled with `foo: bar`, then Prometheus adds the `job="bar"`<br />label to all ingested metrics.<br />If the value of this field is empty or if the label doesn't exist for<br />the given Service, the `job` label of the metrics defaults to the name<br />of the associated Kubernetes `Service`. |  |  |
| `targetLabels` _string array_ | `targetLabels` defines the labels which are transferred from the<br />associated Kubernetes `Service` object onto the ingested metrics. |  |  |
| `podTargetLabels` _string array_ | `podTargetLabels` defines the labels which are transferred from the<br />associated Kubernetes `Pod` object onto the ingested metrics. |  |  |
| `endpoints` _[Endpoint](#endpoint) array_ | List of endpoints part of this ServiceMonitor.<br />Defines how to scrape metrics from Kubernetes [Endpoints](https://kubernetes.io/docs/concepts/services-networking/service/#endpoints) objects.<br />In most cases, an Endpoints object is backed by a Kubernetes [Service](https://kubernetes.io/docs/concepts/services-networking/service/) object with the same name and labels. |  |  |
| `selector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Label selector to select the Kubernetes `Endpoints` objects to scrape metrics from. |  |  |
| `selectorMechanism` _[SelectorMechanism](#selectormechanism)_ | Mechanism used to select the endpoints to scrape.<br />By default, the selection process relies on relabel configurations to filter the discovered targets.<br />Alternatively, you can opt in for role selectors, which may offer better efficiency in large clusters.<br />Which strategy is best for your use case needs to be carefully evaluated.<br />It requires Prometheus >= v2.17.0. |  | Enum: [RelabelConfig RoleSelector] <br /> |
| `namespaceSelector` _[NamespaceSelector](#namespaceselector)_ | `namespaceSelector` defines in which namespace(s) Prometheus should discover the services.<br />By default, the services are discovered in the same namespace as the `ServiceMonitor` object but it is possible to select pods across different/all namespaces. |  |  |
| `sampleLimit` _integer_ | `sampleLimit` defines a per-scrape limit on the number of scraped samples<br />that will be accepted. |  |  |
| `scrapeProtocols` _[ScrapeProtocol](#scrapeprotocol) array_ | `scrapeProtocols` defines the protocols to negotiate during a scrape. It tells clients the<br />protocols supported by Prometheus in order of preference (from most to least preferred).<br />If unset, Prometheus uses its default value.<br />It requires Prometheus >= v2.49.0. |  | Enum: [PrometheusProto OpenMetricsText0.0.1 OpenMetricsText1.0.0 PrometheusText0.0.4 PrometheusText1.0.0] <br /> |
| `fallbackScrapeProtocol` _[ScrapeProtocol](#scrapeprotocol)_ | The protocol to use if a scrape returns blank, unparseable, or otherwise invalid Content-Type.<br />It requires Prometheus >= v3.0.0. |  | Enum: [PrometheusProto OpenMetricsText0.0.1 OpenMetricsText1.0.0 PrometheusText0.0.4 PrometheusText1.0.0] <br /> |
| `targetLimit` _integer_ | `targetLimit` defines a limit on the number of scraped targets that will<br />be accepted. |  |  |
| `labelLimit` _integer_ | Per-scrape limit on number of labels that will be accepted for a sample.<br />It requires Prometheus >= v2.27.0. |  |  |
| `labelNameLengthLimit` _integer_ | Per-scrape limit on length of labels name that will be accepted for a sample.<br />It requires Prometheus >= v2.27.0. |  |  |
| `labelValueLengthLimit` _integer_ | Per-scrape limit on length of labels value that will be accepted for a sample.<br />It requires Prometheus >= v2.27.0. |  |  |
| `scrapeClassicHistograms` _boolean_ | Whether to scrape a classic histogram that is also exposed as a native histogram.<br />It requires Prometheus >= v2.45.0. |  |  |
| `nativeHistogramBucketLimit` _integer_ | If there are more than this many buckets in a native histogram,<br />buckets will be merged to stay within the limit.<br />It requires Prometheus >= v2.45.0. |  |  |
| `nativeHistogramMinBucketFactor` _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#quantity-resource-api)_ | If the growth factor of one bucket to the next is smaller than this,<br />buckets will be merged to increase the factor sufficiently.<br />It requires Prometheus >= v2.50.0. |  |  |
| `convertClassicHistogramsToNHCB` _boolean_ | Whether to convert all scraped classic histograms into a native histogram with custom buckets.<br />It requires Prometheus >= v3.0.0. |  |  |
| `keepDroppedTargets` _integer_ | Per-scrape limit on the number of targets dropped by relabeling<br />that will be kept in memory. 0 means no limit.<br />It requires Prometheus >= v2.47.0. |  |  |
| `attachMetadata` _[AttachMetadata](#attachmetadata)_ | `attachMetadata` defines additional metadata which is added to the<br />discovered targets.<br />It requires Prometheus >= v2.37.0. |  |  |
| `scrapeClass` _string_ | The scrape class to apply. |  | MinLength: 1 <br /> |
| `bodySizeLimit` _[ByteSize](#bytesize)_ | When defined, bodySizeLimit specifies a job level limit on the size<br />of uncompressed response body that will be accepted by Prometheus.<br />It requires Prometheus >= v2.28.0. |  | Pattern: `(^0\|([0-9]*[.])?[0-9]+((K\|M\|G\|T\|E\|P)i?)?B)$` <br /> |


#### ShardRetentionPolicy







_Appears in:_
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `whenScaled` _[WhenScaledRetentionType](#whenscaledretentiontype)_ | Defines the retention policy when the Prometheus shards are scaled down.<br />* `Delete`, the operator will delete the pods from the scaled-down shard(s).<br />* `Retain`, the operator will keep the pods from the scaled-down shard(s), so the data can still be queried.<br />If not defined, the operator assumes the `Delete` value. |  | Enum: [Retain Delete] <br /> |
| `retain` _[RetainConfig](#retainconfig)_ | Defines the config for retention when the retention policy is set to `Retain`.<br />This field is ineffective as of now. |  |  |


#### ShardStatus







_Appears in:_
- [PrometheusStatus](#prometheusstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `shardID` _string_ | Identifier of the shard. |  |  |
| `replicas` _integer_ | Total number of pods targeted by this shard. |  |  |
| `updatedReplicas` _integer_ | Total number of non-terminated pods targeted by this shard<br />that have the desired spec. |  |  |
| `availableReplicas` _integer_ | Total number of available pods (ready for at least minReadySeconds)<br />targeted by this shard. |  |  |
| `unavailableReplicas` _integer_ | Total number of unavailable pods targeted by this shard. |  |  |


#### Sigv4



Sigv4 optionally configures AWS's Signature Verification 4 signing process to
sign requests.



_Appears in:_
- [AlertmanagerEndpoints](#alertmanagerendpoints)
- [RemoteWriteSpec](#remotewritespec)
- [SNSConfig](#snsconfig)
- [SNSConfig](#snsconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `region` _string_ | Region is the AWS region. If blank, the region from the default credentials chain used. |  |  |
| `accessKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | AccessKey is the AWS API key. If not specified, the environment variable<br />`AWS_ACCESS_KEY_ID` is used. |  |  |
| `secretKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | SecretKey is the AWS API secret. If not specified, the environment<br />variable `AWS_SECRET_ACCESS_KEY` is used. |  |  |
| `profile` _string_ | Profile is the named AWS profile used to authenticate. |  |  |
| `roleArn` _string_ | RoleArn is the named AWS profile used to authenticate. |  |  |


#### StorageSpec



StorageSpec defines the configured storage for a group Prometheus servers.
If no storage option is specified, then by default an [EmptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) will be used.

If multiple storage options are specified, priority will be given as follows:
 1. emptyDir
 2. ephemeral
 3. volumeClaimTemplate



_Appears in:_
- [AlertmanagerSpec](#alertmanagerspec)
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)
- [ThanosRulerSpec](#thanosrulerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `disableMountSubPath` _boolean_ | Deprecated: subPath usage will be removed in a future release. |  |  |
| `emptyDir` _[EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#emptydirvolumesource-v1-core)_ | EmptyDirVolumeSource to be used by the StatefulSet.<br />If specified, it takes precedence over `ephemeral` and `volumeClaimTemplate`.<br />More info: https://kubernetes.io/docs/concepts/storage/volumes/#emptydir |  |  |
| `ephemeral` _[EphemeralVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#ephemeralvolumesource-v1-core)_ | EphemeralVolumeSource to be used by the StatefulSet.<br />This is a beta field in k8s 1.21 and GA in 1.15.<br />For lower versions, starting with k8s 1.19, it requires enabling the GenericEphemeralVolume feature gate.<br />More info: https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/#generic-ephemeral-volumes |  |  |
| `volumeClaimTemplate` _[EmbeddedPersistentVolumeClaim](#embeddedpersistentvolumeclaim)_ | Defines the PVC spec to be used by the Prometheus StatefulSets.<br />The easiest way to use a volume that cannot be automatically provisioned<br />is to use a label selector alongside manually created PersistentVolumes. |  |  |


#### TLSConfig



TLSConfig extends the safe TLS configuration with file parameters.



_Appears in:_
- [APIServerConfig](#apiserverconfig)
- [AlertmanagerEndpoints](#alertmanagerendpoints)
- [Endpoint](#endpoint)
- [PrometheusTracingConfig](#prometheustracingconfig)
- [RemoteReadSpec](#remotereadspec)
- [RemoteWriteSpec](#remotewritespec)
- [ScrapeClass](#scrapeclass)
- [ThanosRulerSpec](#thanosrulerspec)
- [ThanosSpec](#thanosspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `ca` _[SecretOrConfigMap](#secretorconfigmap)_ | Certificate authority used when verifying server certificates. |  |  |
| `cert` _[SecretOrConfigMap](#secretorconfigmap)_ | Client certificate to present when doing client-authentication. |  |  |
| `keySecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Secret containing the client key file for the targets. |  |  |
| `serverName` _string_ | Used to verify the hostname for the targets. |  |  |
| `insecureSkipVerify` _boolean_ | Disable target certificate validation. |  |  |
| `minVersion` _[TLSVersion](#tlsversion)_ | Minimum acceptable TLS version.<br />It requires Prometheus >= v2.35.0 or Thanos >= v0.28.0. |  | Enum: [TLS10 TLS11 TLS12 TLS13] <br /> |
| `maxVersion` _[TLSVersion](#tlsversion)_ | Maximum acceptable TLS version.<br />It requires Prometheus >= v2.41.0 or Thanos >= v0.31.0. |  | Enum: [TLS10 TLS11 TLS12 TLS13] <br /> |
| `caFile` _string_ | Path to the CA cert in the Prometheus container to use for the targets. |  |  |
| `certFile` _string_ | Path to the client cert file in the Prometheus container for the targets. |  |  |
| `keyFile` _string_ | Path to the client key file in the Prometheus container for the targets. |  |  |


#### TLSVersion

_Underlying type:_ _string_



_Validation:_
- Enum: [TLS10 TLS11 TLS12 TLS13]

_Appears in:_
- [SafeTLSConfig](#safetlsconfig)
- [TLSConfig](#tlsconfig)

| Field | Description |
| --- | --- |
| `TLS10` |  |
| `TLS11` |  |
| `TLS12` |  |
| `TLS13` |  |


#### TSDBSpec







_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `outOfOrderTimeWindow` _[Duration](#duration)_ | Configures how old an out-of-order/out-of-bounds sample can be with<br />respect to the TSDB max time.<br />An out-of-order/out-of-bounds sample is ingested into the TSDB as long as<br />the timestamp of the sample is >= (TSDB.MaxTime - outOfOrderTimeWindow).<br />This is an *experimental feature*, it may change in any upcoming release<br />in a breaking way.<br />It requires Prometheus >= v2.39.0 or PrometheusAgent >= v2.54.0. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |


#### ThanosRuler



The `ThanosRuler` custom resource definition (CRD) defines a desired [Thanos Ruler](https://github.com/thanos-io/thanos/blob/main/docs/components/rule.md) setup to run in a Kubernetes cluster.

A `ThanosRuler` instance requires at least one compatible Prometheus API endpoint (either Thanos Querier or Prometheus services).

The resource defines via label and namespace selectors which `PrometheusRule` objects should be associated to the deployed Thanos Ruler instances.



_Appears in:_
- [ThanosRulerList](#thanosrulerlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[ThanosRulerSpec](#thanosrulerspec)_ | Specification of the desired behavior of the ThanosRuler cluster. More info:<br />https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status |  |  |
| `status` _[ThanosRulerStatus](#thanosrulerstatus)_ | Most recent observed status of the ThanosRuler cluster. Read-only.<br />More info:<br />https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status |  |  |




#### ThanosRulerSpec



ThanosRulerSpec is a specification of the desired behavior of the ThanosRuler. More info:
https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status



_Appears in:_
- [ThanosRuler](#thanosruler)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `version` _string_ | Version of Thanos to be deployed. |  |  |
| `podMetadata` _[EmbeddedObjectMetadata](#embeddedobjectmetadata)_ | PodMetadata configures labels and annotations which are propagated to the ThanosRuler pods.<br />The following items are reserved and cannot be overridden:<br />* "app.kubernetes.io/name" label, set to "thanos-ruler".<br />* "app.kubernetes.io/managed-by" label, set to "prometheus-operator".<br />* "app.kubernetes.io/instance" label, set to the name of the ThanosRuler instance.<br />* "thanos-ruler" label, set to the name of the ThanosRuler instance.<br />* "kubectl.kubernetes.io/default-container" annotation, set to "thanos-ruler". |  |  |
| `image` _string_ | Thanos container image URL. |  |  |
| `imagePullPolicy` _[PullPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#pullpolicy-v1-core)_ | Image pull policy for the 'thanos', 'init-config-reloader' and 'config-reloader' containers.<br />See https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy for more details. |  | Enum: [ Always Never IfNotPresent] <br /> |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#localobjectreference-v1-core) array_ | An optional list of references to secrets in the same namespace<br />to use for pulling thanos images from registries<br />see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod |  |  |
| `paused` _boolean_ | When a ThanosRuler deployment is paused, no actions except for deletion<br />will be performed on the underlying objects. |  |  |
| `replicas` _integer_ | Number of thanos ruler instances to deploy. |  |  |
| `nodeSelector` _object (keys:string, values:string)_ | Define which Nodes the Pods are scheduled on. |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#resourcerequirements-v1-core)_ | Resources defines the resource requirements for single Pods.<br />If not provided, no requests/limits will be set |  |  |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#affinity-v1-core)_ | If specified, the pod's scheduling constraints. |  |  |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#toleration-v1-core) array_ | If specified, the pod's tolerations. |  |  |
| `topologySpreadConstraints` _[TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#topologyspreadconstraint-v1-core) array_ | If specified, the pod's topology spread constraints. |  |  |
| `securityContext` _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#podsecuritycontext-v1-core)_ | SecurityContext holds pod-level security attributes and common container settings.<br />This defaults to the default PodSecurityContext. |  |  |
| `dnsPolicy` _[DNSPolicy](#dnspolicy)_ | Defines the DNS policy for the pods. |  | Enum: [ClusterFirstWithHostNet ClusterFirst Default None] <br /> |
| `dnsConfig` _[PodDNSConfig](#poddnsconfig)_ | Defines the DNS configuration for the pods. |  |  |
| `enableServiceLinks` _boolean_ | Indicates whether information about services should be injected into pod's environment variables |  |  |
| `priorityClassName` _string_ | Priority class assigned to the Pods |  |  |
| `serviceName` _string_ | The name of the service name used by the underlying StatefulSet(s) as the governing service.<br />If defined, the Service  must be created before the ThanosRuler resource in the same namespace and it must define a selector that matches the pod labels.<br />If empty, the operator will create and manage a headless service named `thanos-ruler-operated` for ThanosRuler resources.<br />When deploying multiple ThanosRuler resources in the same namespace, it is recommended to specify a different value for each.<br />See https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id for more details. |  | MinLength: 1 <br /> |
| `serviceAccountName` _string_ | ServiceAccountName is the name of the ServiceAccount to use to run the<br />Thanos Ruler Pods. |  |  |
| `storage` _[StorageSpec](#storagespec)_ | Storage spec to specify how storage shall be used. |  |  |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volume-v1-core) array_ | Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will<br />be appended to other volumes that are generated as a result of StorageSpec objects. |  |  |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volumemount-v1-core) array_ | VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.<br />VolumeMounts specified will be appended to other VolumeMounts in the ruler container,<br />that are generated as a result of StorageSpec objects. |  |  |
| `objectStorageConfig` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Configures object storage.<br />The configuration format is defined at https://thanos.io/tip/thanos/storage.md/#configuring-access-to-object-storage<br />The operator performs no validation of the configuration.<br />`objectStorageConfigFile` takes precedence over this field. |  |  |
| `objectStorageConfigFile` _string_ | Configures the path of the object storage configuration file.<br />The configuration format is defined at https://thanos.io/tip/thanos/storage.md/#configuring-access-to-object-storage<br />The operator performs no validation of the configuration file.<br />This field takes precedence over `objectStorageConfig`. |  |  |
| `listenLocal` _boolean_ | ListenLocal makes the Thanos ruler listen on loopback, so that it<br />does not bind against the Pod IP. |  |  |
| `queryEndpoints` _string array_ | Configures the list of Thanos Query endpoints from which to query metrics.<br />For Thanos >= v0.11.0, it is recommended to use `queryConfig` instead.<br />`queryConfig` takes precedence over this field. |  |  |
| `queryConfig` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Configures the list of Thanos Query endpoints from which to query metrics.<br />The configuration format is defined at https://thanos.io/tip/components/rule.md/#query-api<br />It requires Thanos >= v0.11.0.<br />The operator performs no validation of the configuration.<br />This field takes precedence over `queryEndpoints`. |  |  |
| `alertmanagersUrl` _string array_ | Configures the list of Alertmanager endpoints to send alerts to.<br />For Thanos >= v0.10.0, it is recommended to use `alertmanagersConfig` instead.<br />`alertmanagersConfig` takes precedence over this field. |  |  |
| `alertmanagersConfig` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Configures the list of Alertmanager endpoints to send alerts to.<br />The configuration format is defined at https://thanos.io/tip/components/rule.md/#alertmanager.<br />It requires Thanos >= v0.10.0.<br />The operator performs no validation of the configuration.<br />This field takes precedence over `alertmanagersUrl`. |  |  |
| `ruleSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | PrometheusRule objects to be selected for rule evaluation. An empty<br />label selector matches all objects. A null label selector matches no<br />objects. |  |  |
| `ruleNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to be selected for Rules discovery. If unspecified, only<br />the same namespace as the ThanosRuler object is in is used. |  |  |
| `enforcedNamespaceLabel` _string_ | EnforcedNamespaceLabel enforces adding a namespace label of origin for each alert<br />and metric that is user created. The label value will always be the namespace of the object that is<br />being created. |  |  |
| `excludedFromEnforcement` _[ObjectReference](#objectreference) array_ | List of references to PrometheusRule objects<br />to be excluded from enforcing a namespace label of origin.<br />Applies only if enforcedNamespaceLabel set to true. |  |  |
| `prometheusRulesExcludedFromEnforce` _[PrometheusRuleExcludeConfig](#prometheusruleexcludeconfig) array_ | PrometheusRulesExcludedFromEnforce - list of Prometheus rules to be excluded from enforcing<br />of adding namespace labels. Works only if enforcedNamespaceLabel set to true.<br />Make sure both ruleNamespace and ruleName are set for each pair<br />Deprecated: use excludedFromEnforcement instead. |  |  |
| `logLevel` _string_ | Log level for ThanosRuler to be configured with. |  | Enum: [ debug info warn error] <br /> |
| `logFormat` _string_ | Log format for ThanosRuler to be configured with. |  | Enum: [ logfmt json] <br /> |
| `portName` _string_ | Port name used for the pods and governing service.<br />Defaults to `web`. | web |  |
| `evaluationInterval` _[Duration](#duration)_ | Interval between consecutive evaluations. | 15s | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `retention` _[Duration](#duration)_ | Time duration ThanosRuler shall retain data for. Default is '24h', and<br />must match the regular expression `[0-9]+(ms\|s\|m\|h\|d\|w\|y)` (milliseconds<br />seconds minutes hours days weeks years).<br />The field has no effect when remote-write is configured since the Ruler<br />operates in stateless mode. | 24h | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `containers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core) array_ | Containers allows injecting additional containers or modifying operator generated<br />containers. This can be used to allow adding an authentication proxy to a ThanosRuler pod or<br />to change the behavior of an operator generated container. Containers described here modify<br />an operator generated container if they share the same name and modifications are done via a<br />strategic merge patch. The current container names are: `thanos-ruler` and `config-reloader`.<br />Overriding containers is entirely outside the scope of what the maintainers will support and by doing<br />so, you accept that this behaviour may break at any time without notice. |  |  |
| `initContainers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core) array_ | InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.<br />fetch secrets for injection into the ThanosRuler configuration from external sources. Any<br />errors during the execution of an initContainer will lead to a restart of the Pod.<br />More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/<br />Using initContainers for any use case other then secret fetching is entirely outside the scope<br />of what the maintainers will support and by doing so, you accept that this behaviour may break<br />at any time without notice. |  |  |
| `tracingConfig` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Configures tracing.<br />The configuration format is defined at https://thanos.io/tip/thanos/tracing.md/#configuration<br />This is an *experimental feature*, it may change in any upcoming release<br />in a breaking way.<br />The operator performs no validation of the configuration.<br />`tracingConfigFile` takes precedence over this field. |  |  |
| `tracingConfigFile` _string_ | Configures the path of the tracing configuration file.<br />The configuration format is defined at https://thanos.io/tip/thanos/tracing.md/#configuration<br />This is an *experimental feature*, it may change in any upcoming release<br />in a breaking way.<br />The operator performs no validation of the configuration file.<br />This field takes precedence over `tracingConfig`. |  |  |
| `labels` _object (keys:string, values:string)_ | Configures the external label pairs of the ThanosRuler resource.<br />A default replica label `thanos_ruler_replica` will be always added as a<br />label with the value of the pod's name. |  |  |
| `alertDropLabels` _string array_ | Configures the label names which should be dropped in Thanos Ruler<br />alerts.<br />The replica label `thanos_ruler_replica` will always be dropped from the alerts. |  |  |
| `externalPrefix` _string_ | The external URL the Thanos Ruler instances will be available under. This is<br />necessary to generate correct URLs. This is necessary if Thanos Ruler is not<br />served from root of a DNS name. |  |  |
| `routePrefix` _string_ | The route prefix ThanosRuler registers HTTP handlers for. This allows thanos UI to be served on a sub-path. |  |  |
| `grpcServerTlsConfig` _[TLSConfig](#tlsconfig)_ | GRPCServerTLSConfig configures the gRPC server from which Thanos Querier reads<br />recorded rule data.<br />Note: Currently only the CAFile, CertFile, and KeyFile fields are supported.<br />Maps to the '--grpc-server-tls-*' CLI args. |  |  |
| `alertQueryUrl` _string_ | The external Query URL the Thanos Ruler will set in the 'Source' field<br />of all alerts.<br />Maps to the '--alert.query-url' CLI arg. |  |  |
| `minReadySeconds` _integer_ | Minimum number of seconds for which a newly created pod should be ready<br />without any of its container crashing for it to be considered available.<br />Defaults to 0 (pod will be considered available as soon as it is ready)<br />This is an alpha field from kubernetes 1.22 until 1.24 which requires enabling the StatefulSetMinReadySeconds feature gate. |  |  |
| `alertRelabelConfigs` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Configures alert relabeling in Thanos Ruler.<br />Alert relabel configuration must have the form as specified in the<br />official Prometheus documentation:<br />https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs<br />The operator performs no validation of the configuration.<br />`alertRelabelConfigFile` takes precedence over this field. |  |  |
| `alertRelabelConfigFile` _string_ | Configures the path to the alert relabeling configuration file.<br />Alert relabel configuration must have the form as specified in the<br />official Prometheus documentation:<br />https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs<br />The operator performs no validation of the configuration file.<br />This field takes precedence over `alertRelabelConfig`. |  |  |
| `hostAliases` _[HostAlias](#hostalias) array_ | Pods' hostAliases configuration |  |  |
| `additionalArgs` _[Argument](#argument) array_ | AdditionalArgs allows setting additional arguments for the ThanosRuler container.<br />It is intended for e.g. activating hidden flags which are not supported by<br />the dedicated configuration options yet. The arguments are passed as-is to the<br />ThanosRuler container which may cause issues if they are invalid or not supported<br />by the given ThanosRuler version.<br />In case of an argument conflict (e.g. an argument which is already set by the<br />operator itself) or when providing an invalid argument the reconciliation will<br />fail and an error will be logged. |  |  |
| `web` _[ThanosRulerWebSpec](#thanosrulerwebspec)_ | Defines the configuration of the ThanosRuler web server. |  |  |
| `remoteWrite` _[RemoteWriteSpec](#remotewritespec) array_ | Defines the list of remote write configurations.<br />When the list isn't empty, the ruler is configured with stateless mode.<br />It requires Thanos >= 0.24.0. |  |  |
| `terminationGracePeriodSeconds` _integer_ | Optional duration in seconds the pod needs to terminate gracefully.<br />Value must be non-negative integer. The value zero indicates stop immediately via<br />the kill signal (no opportunity to shut down) which may lead to data corruption.<br />Defaults to 120 seconds. |  | Minimum: 0 <br /> |


#### ThanosRulerStatus



ThanosRulerStatus is the most recent observed status of the ThanosRuler. Read-only.
More info:
https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status



_Appears in:_
- [ThanosRuler](#thanosruler)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `paused` _boolean_ | Represents whether any actions on the underlying managed objects are<br />being performed. Only delete actions will be performed. |  |  |
| `replicas` _integer_ | Total number of non-terminated pods targeted by this ThanosRuler deployment<br />(their labels match the selector). |  |  |
| `updatedReplicas` _integer_ | Total number of non-terminated pods targeted by this ThanosRuler deployment<br />that have the desired version spec. |  |  |
| `availableReplicas` _integer_ | Total number of available pods (ready for at least minReadySeconds)<br />targeted by this ThanosRuler deployment. |  |  |
| `unavailableReplicas` _integer_ | Total number of unavailable pods targeted by this ThanosRuler deployment. |  |  |
| `conditions` _[Condition](#condition) array_ | The current state of the ThanosRuler object. |  |  |


#### ThanosRulerWebSpec



ThanosRulerWebSpec defines the configuration of the ThanosRuler web server.



_Appears in:_
- [ThanosRulerSpec](#thanosrulerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `tlsConfig` _[WebTLSConfig](#webtlsconfig)_ | Defines the TLS parameters for HTTPS. |  |  |
| `httpConfig` _[WebHTTPConfig](#webhttpconfig)_ | Defines HTTP parameters for web server. |  |  |


#### ThanosSpec



ThanosSpec defines the configuration of the Thanos sidecar.



_Appears in:_
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | Container image name for Thanos. If specified, it takes precedence over<br />the `spec.thanos.baseImage`, `spec.thanos.tag` and `spec.thanos.sha`<br />fields.<br />Specifying `spec.thanos.version` is still necessary to ensure the<br />Prometheus Operator knows which version of Thanos is being configured.<br />If neither `spec.thanos.image` nor `spec.thanos.baseImage` are defined,<br />the operator will use the latest upstream version of Thanos available at<br />the time when the operator was released. |  |  |
| `version` _string_ | Version of Thanos being deployed. The operator uses this information<br />to generate the Prometheus StatefulSet + configuration files.<br />If not specified, the operator assumes the latest upstream release of<br />Thanos available at the time when the version of the operator was<br />released. |  |  |
| `tag` _string_ | Deprecated: use 'image' instead. The image's tag can be specified as as part of the image name. |  |  |
| `sha` _string_ | Deprecated: use 'image' instead.  The image digest can be specified as part of the image name. |  |  |
| `baseImage` _string_ | Deprecated: use 'image' instead. |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#resourcerequirements-v1-core)_ | Defines the resources requests and limits of the Thanos sidecar. |  |  |
| `objectStorageConfig` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Defines the Thanos sidecar's configuration to upload TSDB blocks to object storage.<br />More info: https://thanos.io/tip/thanos/storage.md/<br />objectStorageConfigFile takes precedence over this field. |  |  |
| `objectStorageConfigFile` _string_ | Defines the Thanos sidecar's configuration file to upload TSDB blocks to object storage.<br />More info: https://thanos.io/tip/thanos/storage.md/<br />This field takes precedence over objectStorageConfig. |  |  |
| `listenLocal` _boolean_ | Deprecated: use `grpcListenLocal` and `httpListenLocal` instead. |  |  |
| `grpcListenLocal` _boolean_ | When true, the Thanos sidecar listens on the loopback interface instead<br />of the Pod IP's address for the gRPC endpoints.<br />It has no effect if `listenLocal` is true. |  |  |
| `httpListenLocal` _boolean_ | When true, the Thanos sidecar listens on the loopback interface instead<br />of the Pod IP's address for the HTTP endpoints.<br />It has no effect if `listenLocal` is true. |  |  |
| `tracingConfig` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Defines the tracing configuration for the Thanos sidecar.<br />`tracingConfigFile` takes precedence over this field.<br />More info: https://thanos.io/tip/thanos/tracing.md/<br />This is an *experimental feature*, it may change in any upcoming release<br />in a breaking way. |  |  |
| `tracingConfigFile` _string_ | Defines the tracing configuration file for the Thanos sidecar.<br />This field takes precedence over `tracingConfig`.<br />More info: https://thanos.io/tip/thanos/tracing.md/<br />This is an *experimental feature*, it may change in any upcoming release<br />in a breaking way. |  |  |
| `grpcServerTlsConfig` _[TLSConfig](#tlsconfig)_ | Configures the TLS parameters for the gRPC server providing the StoreAPI.<br />Note: Currently only the `caFile`, `certFile`, and `keyFile` fields are supported. |  |  |
| `logLevel` _string_ | Log level for the Thanos sidecar. |  | Enum: [ debug info warn error] <br /> |
| `logFormat` _string_ | Log format for the Thanos sidecar. |  | Enum: [ logfmt json] <br /> |
| `minTime` _string_ | Defines the start of time range limit served by the Thanos sidecar's StoreAPI.<br />The field's value should be a constant time in RFC3339 format or a time<br />duration relative to current time, such as -1d or 2h45m. Valid duration<br />units are ms, s, m, h, d, w, y. |  |  |
| `blockSize` _[Duration](#duration)_ | BlockDuration controls the size of TSDB blocks produced by Prometheus.<br />The default value is 2h to match the upstream Prometheus defaults.<br />WARNING: Changing the block duration can impact the performance and<br />efficiency of the entire Prometheus/Thanos stack due to how it interacts<br />with memory and Thanos compactors. It is recommended to keep this value<br />set to a multiple of 120 times your longest scrape or rule interval. For<br />example, 30s * 120 = 1h. | 2h | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `readyTimeout` _[Duration](#duration)_ | ReadyTimeout is the maximum time that the Thanos sidecar will wait for<br />Prometheus to start. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `getConfigInterval` _[Duration](#duration)_ | How often to retrieve the Prometheus configuration. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `getConfigTimeout` _[Duration](#duration)_ | Maximum time to wait when retrieving the Prometheus configuration. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volumemount-v1-core) array_ | VolumeMounts allows configuration of additional VolumeMounts for Thanos.<br />VolumeMounts specified will be appended to other VolumeMounts in the<br />'thanos-sidecar' container. |  |  |
| `additionalArgs` _[Argument](#argument) array_ | AdditionalArgs allows setting additional arguments for the Thanos container.<br />The arguments are passed as-is to the Thanos container which may cause issues<br />if they are invalid or not supported the given Thanos version.<br />In case of an argument conflict (e.g. an argument which is already set by the<br />operator itself) or when providing an invalid argument, the reconciliation will<br />fail and an error will be logged. |  |  |


#### TopologySpreadConstraint







_Appears in:_
- [CommonPrometheusFields](#commonprometheusfields)
- [PrometheusAgentSpec](#prometheusagentspec)
- [PrometheusSpec](#prometheusspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `additionalLabelSelectors` _[AdditionalLabelSelectors](#additionallabelselectors)_ | Defines what Prometheus Operator managed labels should be added to labelSelector on the topologySpreadConstraint. |  | Enum: [OnResource OnShard] <br /> |


#### TranslationStrategyOption

_Underlying type:_ _string_

TranslationStrategyOption represents a translation strategy option for the OTLP endpoint.
Supported values are:
* `NoUTF8EscapingWithSuffixes`
* `UnderscoreEscapingWithSuffixes`
* `NoTranslation`

_Validation:_
- Enum: [NoUTF8EscapingWithSuffixes UnderscoreEscapingWithSuffixes NoTranslation]

_Appears in:_
- [OTLPConfig](#otlpconfig)

| Field | Description |
| --- | --- |
| `NoUTF8EscapingWithSuffixes` |  |
| `UnderscoreEscapingWithSuffixes` |  |
| `NoTranslation` | It requires Prometheus >= v3.4.0.<br /> |


#### WebConfigFileFields



WebConfigFileFields defines the file content for --web.config.file flag.



_Appears in:_
- [AlertmanagerWebSpec](#alertmanagerwebspec)
- [PrometheusWebSpec](#prometheuswebspec)
- [ThanosRulerWebSpec](#thanosrulerwebspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `tlsConfig` _[WebTLSConfig](#webtlsconfig)_ | Defines the TLS parameters for HTTPS. |  |  |
| `httpConfig` _[WebHTTPConfig](#webhttpconfig)_ | Defines HTTP parameters for web server. |  |  |


#### WebHTTPConfig



WebHTTPConfig defines HTTP parameters for web server.



_Appears in:_
- [AlertmanagerWebSpec](#alertmanagerwebspec)
- [PrometheusWebSpec](#prometheuswebspec)
- [ThanosRulerWebSpec](#thanosrulerwebspec)
- [WebConfigFileFields](#webconfigfilefields)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `http2` _boolean_ | Enable HTTP/2 support. Note that HTTP/2 is only supported with TLS.<br />When TLSConfig is not configured, HTTP/2 will be disabled.<br />Whenever the value of the field changes, a rolling update will be triggered. |  |  |
| `headers` _[WebHTTPHeaders](#webhttpheaders)_ | List of headers that can be added to HTTP responses. |  |  |


#### WebHTTPHeaders



WebHTTPHeaders defines the list of headers that can be added to HTTP responses.



_Appears in:_
- [WebHTTPConfig](#webhttpconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `contentSecurityPolicy` _string_ | Set the Content-Security-Policy header to HTTP responses.<br />Unset if blank. |  |  |
| `xFrameOptions` _string_ | Set the X-Frame-Options header to HTTP responses.<br />Unset if blank. Accepted values are deny and sameorigin.<br />https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options |  | Enum: [ Deny SameOrigin] <br /> |
| `xContentTypeOptions` _string_ | Set the X-Content-Type-Options header to HTTP responses.<br />Unset if blank. Accepted value is nosniff.<br />https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options |  | Enum: [ NoSniff] <br /> |
| `xXSSProtection` _string_ | Set the X-XSS-Protection header to all responses.<br />Unset if blank.<br />https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-XSS-Protection |  |  |
| `strictTransportSecurity` _string_ | Set the Strict-Transport-Security header to HTTP responses.<br />Unset if blank.<br />Please make sure that you use this with care as this header might force<br />browsers to load Prometheus and the other applications hosted on the same<br />domain and subdomains over HTTPS.<br />https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security |  |  |


#### WebTLSConfig



WebTLSConfig defines the TLS parameters for HTTPS.



_Appears in:_
- [AlertmanagerWebSpec](#alertmanagerwebspec)
- [ClusterTLSConfig](#clustertlsconfig)
- [PrometheusWebSpec](#prometheuswebspec)
- [ThanosRulerWebSpec](#thanosrulerwebspec)
- [WebConfigFileFields](#webconfigfilefields)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `cert` _[SecretOrConfigMap](#secretorconfigmap)_ | Secret or ConfigMap containing the TLS certificate for the web server.<br />Either `keySecret` or `keyFile` must be defined.<br />It is mutually exclusive with `certFile`. |  |  |
| `certFile` _string_ | Path to the TLS certificate file in the container for the web server.<br />Either `keySecret` or `keyFile` must be defined.<br />It is mutually exclusive with `cert`. |  |  |
| `keySecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Secret containing the TLS private key for the web server.<br />Either `cert` or `certFile` must be defined.<br />It is mutually exclusive with `keyFile`. |  |  |
| `keyFile` _string_ | Path to the TLS private key file in the container for the web server.<br />If defined, either `cert` or `certFile` must be defined.<br />It is mutually exclusive with `keySecret`. |  |  |
| `client_ca` _[SecretOrConfigMap](#secretorconfigmap)_ | Secret or ConfigMap containing the CA certificate for client certificate<br />authentication to the server.<br />It is mutually exclusive with `clientCAFile`. |  |  |
| `clientCAFile` _string_ | Path to the CA certificate file for client certificate authentication to<br />the server.<br />It is mutually exclusive with `client_ca`. |  |  |
| `clientAuthType` _string_ | The server policy for client TLS authentication.<br />For more detail on clientAuth options:<br />https://golang.org/pkg/crypto/tls/#ClientAuthType |  |  |
| `minVersion` _string_ | Minimum TLS version that is acceptable. |  |  |
| `maxVersion` _string_ | Maximum TLS version that is acceptable. |  |  |
| `cipherSuites` _string array_ | List of supported cipher suites for TLS versions up to TLS 1.2.<br />If not defined, the Go default cipher suites are used.<br />Available cipher suites are documented in the Go documentation:<br />https://golang.org/pkg/crypto/tls/#pkg-constants |  |  |
| `preferServerCipherSuites` _boolean_ | Controls whether the server selects the client's most preferred cipher<br />suite, or the server's most preferred cipher suite.<br />If true then the server's preference, as expressed in<br />the order of elements in cipherSuites, is used. |  |  |
| `curvePreferences` _string array_ | Elliptic curves that will be used in an ECDHE handshake, in preference<br />order.<br />Available curves are documented in the Go documentation:<br />https://golang.org/pkg/crypto/tls/#CurveID |  |  |


#### WhenScaledRetentionType

_Underlying type:_ _string_





_Appears in:_
- [ShardRetentionPolicy](#shardretentionpolicy)




## monitoring.coreos.com/v1alpha1




#### AlertmanagerConfig



AlertmanagerConfig configures the Prometheus Alertmanager,
specifying how alerts should be grouped, inhibited and notified to external systems.



_Appears in:_
- [AlertmanagerConfigList](#alertmanagerconfiglist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[AlertmanagerConfigSpec](#alertmanagerconfigspec)_ |  |  |  |




#### AlertmanagerConfigSpec



AlertmanagerConfigSpec is a specification of the desired behavior of the
Alertmanager configuration.
By default, the Alertmanager configuration only applies to alerts for which
the `namespace` label is equal to the namespace of the AlertmanagerConfig
resource (see the `.spec.alertmanagerConfigMatcherStrategy` field of the
Alertmanager CRD).



_Appears in:_
- [AlertmanagerConfig](#alertmanagerconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `route` _[Route](#route)_ | The Alertmanager route definition for alerts matching the resource's<br />namespace. If present, it will be added to the generated Alertmanager<br />configuration as a first-level route. |  |  |
| `receivers` _[Receiver](#receiver) array_ | List of receivers. |  |  |
| `inhibitRules` _[InhibitRule](#inhibitrule) array_ | List of inhibition rules. The rules will only apply to alerts matching<br />the resource's namespace. |  |  |
| `muteTimeIntervals` _[MuteTimeInterval](#mutetimeinterval) array_ | List of MuteTimeInterval specifying when the routes should be muted. |  |  |


#### AttachMetadata







_Appears in:_
- [KubernetesSDConfig](#kubernetessdconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `node` _boolean_ | Attaches node metadata to discovered targets.<br />When set to true, Prometheus must have the `get` permission on the<br />`Nodes` objects.<br />Only valid for Pod, Endpoint and Endpointslice roles. |  |  |


#### AuthenticationMethodType

_Underlying type:_ _string_



_Validation:_
- Enum: [OAuth ManagedIdentity SDK]

_Appears in:_
- [AzureSDConfig](#azuresdconfig)

| Field | Description |
| --- | --- |
| `OAuth` |  |
| `ManagedIdentity` |  |
| `SDK` |  |


#### AzureSDConfig



AzureSDConfig allow retrieving scrape targets from Azure VMs.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#azure_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `environment` _string_ | The Azure environment. |  | MinLength: 1 <br /> |
| `authenticationMethod` _[AuthenticationMethodType](#authenticationmethodtype)_ | # The authentication method, either `OAuth` or `ManagedIdentity` or `SDK`.<br />See https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview<br />SDK authentication method uses environment variables by default.<br />See https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication |  | Enum: [OAuth ManagedIdentity SDK] <br /> |
| `subscriptionID` _string_ | The subscription ID. Always required. |  | MinLength: 1 <br /> |
| `tenantID` _string_ | Optional tenant ID. Only required with the OAuth authentication method. |  | MinLength: 1 <br /> |
| `clientID` _string_ | Optional client ID. Only required with the OAuth authentication method. |  | MinLength: 1 <br /> |
| `clientSecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Optional client secret. Only required with the OAuth authentication method. |  |  |
| `resourceGroup` _string_ | Optional resource group name. Limits discovery to this resource group.<br />Requires  Prometheus v2.35.0 and above |  | MinLength: 1 <br /> |
| `refreshInterval` _[Duration](#duration)_ | RefreshInterval configures the refresh interval at which Prometheus will re-read the instance list. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `port` _integer_ | The port to scrape metrics from. If using the public IP address, this must<br />instead be specified in the relabeling rule. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth information to authenticate against the target HTTP endpoint.<br />More info: https://prometheus.io/docs/operating/configuration/#endpoints<br />Cannot be set at the same time as `authorization`, or `oAuth2`. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header configuration to authenticate against the target HTTP endpoint.<br />Cannot be set at the same time as `oAuth2`, or `basicAuth`. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Optional OAuth 2.0 configuration to authenticate against the target HTTP endpoint.<br />Cannot be set at the same time as `authorization`, or `basicAuth`. |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration applying to the target HTTP endpoint. |  |  |


#### ConsulSDConfig



ConsulSDConfig defines a Consul service discovery configuration
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#consul_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `server` _string_ | Consul server address. A valid string consisting of a hostname or IP followed by an optional port number. |  | MinLength: 1 <br /> |
| `pathPrefix` _string_ | Prefix for URIs for when consul is behind an API gateway (reverse proxy).<br />It requires Prometheus >= 2.45.0. |  | MinLength: 1 <br /> |
| `tokenRef` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Consul ACL TokenRef, if not provided it will use the ACL from the local Consul Agent. |  |  |
| `datacenter` _string_ | Consul Datacenter name, if not provided it will use the local Consul Agent Datacenter. |  | MinLength: 1 <br /> |
| `namespace` _string_ | Namespaces are only supported in Consul Enterprise.<br />It requires Prometheus >= 2.28.0. |  | MinLength: 1 <br /> |
| `partition` _string_ | Admin Partitions are only supported in Consul Enterprise. |  | MinLength: 1 <br /> |
| `scheme` _string_ | HTTP Scheme default "http" |  | Enum: [HTTP HTTPS] <br /> |
| `services` _string array_ | A list of services for which targets are retrieved. If omitted, all services are scraped. |  |  |
| `tags` _string array_ | An optional list of tags used to filter nodes for a given service. Services must contain all tags in the list.<br />Starting with Consul 1.14, it is recommended to use `filter` with the `ServiceTags` selector instead. |  |  |
| `tagSeparator` _string_ | The string by which Consul tags are joined into the tag label.<br />If unset, Prometheus uses its default value. |  | MinLength: 1 <br /> |
| `nodeMeta` _object (keys:string, values:string)_ | Node metadata key/value pairs to filter nodes for a given service.<br />Starting with Consul 1.14, it is recommended to use `filter` with the `NodeMeta` selector instead. |  |  |
| `filter` _string_ | Filter expression used to filter the catalog results.<br />See https://www.consul.io/api-docs/catalog#list-services<br />It requires Prometheus >= 3.0.0. |  | MinLength: 1 <br /> |
| `allowStale` _boolean_ | Allow stale Consul results (see https://www.consul.io/api/features/consistency.html ). Will reduce load on Consul.<br />If unset, Prometheus uses its default value. |  |  |
| `refreshInterval` _[Duration](#duration)_ | The time after which the provided names are refreshed.<br />On large setup it might be a good idea to increase this value because the catalog will change all the time.<br />If unset, Prometheus uses its default value. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `basicAuth` _[BasicAuth](#basicauth)_ | Optional BasicAuth information to authenticate against the Consul Server.<br />More info: https://prometheus.io/docs/operating/configuration/#endpoints<br />Cannot be set at the same time as `authorization`, or `oauth2`. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Optional Authorization header configuration to authenticate against the Consul Server.<br />Cannot be set at the same time as `basicAuth`, or `oauth2`. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Optional OAuth2.0 configuration.<br />Cannot be set at the same time as `basicAuth`, or `authorization`. |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects.<br />If unset, Prometheus uses its default value. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2.<br />If unset, Prometheus uses its default value. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration to connect to the Consul API. |  |  |


#### DNSRecordType

_Underlying type:_ _string_



_Validation:_
- Enum: [A AAAA MX NS SRV]

_Appears in:_
- [DNSSDConfig](#dnssdconfig)

| Field | Description |
| --- | --- |
| `A` |  |
| `SRV` |  |
| `AAAA` |  |
| `MX` |  |
| `NS` |  |


#### DNSSDConfig



DNSSDConfig allows specifying a set of DNS domain names which are periodically queried to discover a list of targets.
The DNS servers to be contacted are read from /etc/resolv.conf.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#dns_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `names` _string array_ | A list of DNS domain names to be queried. |  | MinItems: 1 <br />items:MinLength: 1 <br /> |
| `refreshInterval` _[Duration](#duration)_ | RefreshInterval configures the time after which the provided names are refreshed.<br />If not set, Prometheus uses its default value. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `type` _[DNSRecordType](#dnsrecordtype)_ | The type of DNS query to perform. One of SRV, A, AAAA, MX or NS.<br />If not set, Prometheus uses its default value.<br />When set to NS, it requires Prometheus >= v2.49.0.<br />When set to MX, it requires Prometheus >= v2.38.0 |  | Enum: [A AAAA MX NS SRV] <br /> |
| `port` _integer_ | The port number used if the query type is not SRV<br />Ignored for SRV records |  | Maximum: 65535 <br />Minimum: 0 <br /> |


#### DayOfMonthRange



DayOfMonthRange is an inclusive range of days of the month beginning at 1



_Appears in:_
- [TimeInterval](#timeinterval)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `start` _integer_ | Start of the inclusive range |  | Maximum: 31 <br />Minimum: -31 <br /> |
| `end` _integer_ | End of the inclusive range |  | Maximum: 31 <br />Minimum: -31 <br /> |


#### DigitalOceanSDConfig



DigitalOceanSDConfig allow retrieving scrape targets from DigitalOcean's Droplets API.
This service discovery uses the public IPv4 address by default, by that can be changed with relabeling
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#digitalocean_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header configuration to authenticate against the DigitalOcean API.<br />Cannot be set at the same time as `oauth2`. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Optional OAuth 2.0 configuration.<br />Cannot be set at the same time as `authorization`. |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration applying to the target HTTP endpoint. |  |  |
| `port` _integer_ | The port to scrape metrics from. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `refreshInterval` _[Duration](#duration)_ | Refresh interval to re-read the instance list. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |


#### DiscordConfig



DiscordConfig configures notifications via Discord.
See https://prometheus.io/docs/alerting/latest/configuration/#discord_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `apiURL` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the Discord webhook URL.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `title` _string_ | The template of the message's title. |  |  |
| `message` _string_ | The template of the message's body. |  |  |
| `content` _string_ | The template of the content's body. |  | MinLength: 1 <br /> |
| `username` _string_ | The username of the message sender. |  | MinLength: 1 <br /> |
| `avatarURL` _[URL](#url)_ | The avatar url of the message sender. |  | Pattern: `^https?://.+$` <br /> |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### DockerSDConfig



Docker SD configurations allow retrieving scrape targets from Docker Engine hosts.
This SD discovers "containers" and will create a target for each network IP and
port the container is configured to expose.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#docker_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ | Address of the docker daemon |  | MinLength: 1 <br /> |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration applying to the target HTTP endpoint. |  |  |
| `port` _integer_ | The port to scrape metrics from. |  |  |
| `hostNetworkingHost` _string_ | The host to use if the container is in host networking mode. |  |  |
| `matchFirstNetwork` _boolean_ | Configure whether to match the first network if the container has multiple networks defined.<br />If unset, Prometheus uses true by default.<br />It requires Prometheus >= v2.54.1. |  |  |
| `filters` _[Filters](#filters)_ | Optional filters to limit the discovery process to a subset of the available resources. |  |  |
| `refreshInterval` _[Duration](#duration)_ | Time after which the container is refreshed. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth information to use on every scrape request. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header configuration to authenticate against the Docker API.<br />Cannot be set at the same time as `oauth2`. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Optional OAuth 2.0 configuration.<br />Cannot be set at the same time as `authorization`. |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2. |  |  |


#### DockerSwarmSDConfig



DockerSwarmSDConfig configurations allow retrieving scrape targets from Docker Swarm engine.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#dockerswarm_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ | Address of the Docker daemon |  | Pattern: `^[a-zA-Z][a-zA-Z0-9+.-]*://.+$` <br /> |
| `role` _string_ | Role of the targets to retrieve. Must be `Services`, `Tasks`, or `Nodes`. |  | Enum: [Services Tasks Nodes] <br /> |
| `port` _integer_ | The port to scrape metrics from, when `role` is nodes, and for discovered<br />tasks and services that don't have published ports. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `filters` _[Filters](#filters)_ | Optional filters to limit the discovery process to a subset of available<br />resources.<br />The available filters are listed in the upstream documentation:<br />Services: https://docs.docker.com/engine/api/v1.40/#operation/ServiceList<br />Tasks: https://docs.docker.com/engine/api/v1.40/#operation/TaskList<br />Nodes: https://docs.docker.com/engine/api/v1.40/#operation/NodeList |  |  |
| `refreshInterval` _[Duration](#duration)_ | The time after which the service discovery data is refreshed. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `basicAuth` _[BasicAuth](#basicauth)_ | Optional HTTP basic authentication information. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header configuration to authenticate against the target HTTP endpoint. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Optional OAuth 2.0 configuration.<br />Cannot be set at the same time as `authorization`, or `basicAuth`. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration to use on every scrape request |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2. |  |  |


#### EC2SDConfig



EC2SDConfig allow retrieving scrape targets from AWS EC2 instances.
The private IP address is used by default, but may be changed to the public IP address with relabeling.
The IAM credentials used must have the ec2:DescribeInstances permission to discover scrape targets
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#ec2_sd_config

The EC2 service discovery requires AWS API keys or role ARN for authentication.
BasicAuth, Authorization and OAuth2 fields are not present on purpose.



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `region` _string_ | The AWS region. |  | MinLength: 1 <br /> |
| `accessKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | AccessKey is the AWS API key. |  |  |
| `secretKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | SecretKey is the AWS API secret. |  |  |
| `roleARN` _string_ | AWS Role ARN, an alternative to using AWS API keys. |  | MinLength: 1 <br /> |
| `port` _integer_ | The port to scrape metrics from. If using the public IP address, this must<br />instead be specified in the relabeling rule. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `refreshInterval` _[Duration](#duration)_ | RefreshInterval configures the refresh interval at which Prometheus will re-read the instance list. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `filters` _[Filters](#filters)_ | Filters can be used optionally to filter the instance list by other criteria.<br />Available filter criteria can be found here:<br />https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html<br />Filter API documentation: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Filter.html<br />It requires Prometheus >= v2.3.0 |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration to connect to the AWS EC2 API.<br />It requires Prometheus >= v2.41.0 |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects.<br />It requires Prometheus >= v2.41.0 |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2.<br />It requires Prometheus >= v2.41.0 |  |  |


#### EmailConfig



EmailConfig configures notifications via Email.



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `to` _string_ | The email address to send notifications to. |  |  |
| `from` _string_ | The sender address. |  |  |
| `hello` _string_ | The hostname to identify to the SMTP server. |  |  |
| `smarthost` _string_ | The SMTP host and port through which emails are sent. E.g. example.com:25 |  |  |
| `authUsername` _string_ | The username to use for authentication. |  |  |
| `authPassword` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the password to use for authentication.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `authSecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the CRAM-MD5 secret.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `authIdentity` _string_ | The identity to use for authentication. |  |  |
| `headers` _[KeyValue](#keyvalue) array_ | Further headers email header key/value pairs. Overrides any headers<br />previously set by the notification implementation. |  |  |
| `html` _string_ | The HTML body of the email notification. |  |  |
| `text` _string_ | The text body of the email notification. |  |  |
| `requireTLS` _boolean_ | The SMTP TLS requirement.<br />Note that Go does not support unencrypted connections to remote SMTP endpoints. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration |  |  |


#### EurekaSDConfig



Eureka SD configurations allow retrieving scrape targets using the Eureka REST API.
Prometheus will periodically check the REST endpoint and create a target for every app instance.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#eureka_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `server` _string_ | The URL to connect to the Eureka server. |  | MinLength: 1 <br /> |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth information to use on every scrape request. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header to use on every scrape request. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Optional OAuth 2.0 configuration.<br />Cannot be set at the same time as `authorization` or `basic_auth`. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration applying to the target HTTP endpoint. |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2. |  |  |
| `refreshInterval` _[Duration](#duration)_ | Refresh interval to re-read the instance list. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |


#### FileSDConfig



FileSDConfig defines a Prometheus file service discovery configuration
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#file_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `files` _[SDFile](#sdfile) array_ | List of files to be used for file discovery. Recommendation: use absolute paths. While relative paths work, the<br />prometheus-operator project makes no guarantees about the working directory where the configuration file is<br />stored.<br />Files must be mounted using Prometheus.ConfigMaps or Prometheus.Secrets. |  | MinItems: 1 <br />Pattern: `^[^*]*(\*[^/]*)?\.(json\|yml\|yaml\|JSON\|YML\|YAML)$` <br /> |
| `refreshInterval` _[Duration](#duration)_ | RefreshInterval configures the refresh interval at which Prometheus will reload the content of the files. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |


#### Filter



Filter name and value pairs to limit the discovery process to a subset of available resources.



_Appears in:_
- [Filters](#filters)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the Filter. |  |  |
| `values` _string array_ | Value to filter on. |  | MinItems: 1 <br />items:MinLength: 1 <br /> |


#### Filters

_Underlying type:_ _[Filter](#filter)_





_Appears in:_
- [DockerSDConfig](#dockersdconfig)
- [DockerSwarmSDConfig](#dockerswarmsdconfig)
- [EC2SDConfig](#ec2sdconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the Filter. |  |  |
| `values` _string array_ | Value to filter on. |  | MinItems: 1 <br />items:MinLength: 1 <br /> |


#### GCESDConfig



GCESDConfig configures scrape targets from GCP GCE instances.
The private IP address is used by default, but may be changed to
the public IP address with relabeling.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#gce_sd_config

The GCE service discovery will load the Google Cloud credentials
from the file specified by the GOOGLE_APPLICATION_CREDENTIALS environment variable.
See https://cloud.google.com/kubernetes-engine/docs/tutorials/authenticating-to-cloud-platform

A pre-requisite for using GCESDConfig is that a Secret containing valid
Google Cloud credentials is mounted into the Prometheus or PrometheusAgent
pod via the `.spec.secrets` field and that the GOOGLE_APPLICATION_CREDENTIALS
environment variable is set to /etc/prometheus/secrets/<secret-name>/<credentials-filename.json>.



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `project` _string_ | The Google Cloud Project ID |  | MinLength: 1 <br /> |
| `zone` _string_ | The zone of the scrape targets. If you need multiple zones use multiple GCESDConfigs. |  | MinLength: 1 <br /> |
| `filter` _string_ | Filter can be used optionally to filter the instance list by other criteria<br />Syntax of this filter is described in the filter query parameter section:<br />https://cloud.google.com/compute/docs/reference/latest/instances/list |  | MinLength: 1 <br /> |
| `refreshInterval` _[Duration](#duration)_ | RefreshInterval configures the refresh interval at which Prometheus will re-read the instance list. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `port` _integer_ | The port to scrape metrics from. If using the public IP address, this must<br />instead be specified in the relabeling rule. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `tagSeparator` _string_ | The tag separator is used to separate the tags on concatenation |  | MinLength: 1 <br /> |


#### HTTPConfig



HTTPConfig defines a client HTTP configuration.
See https://prometheus.io/docs/alerting/latest/configuration/#http_config



_Appears in:_
- [DiscordConfig](#discordconfig)
- [MSTeamsConfig](#msteamsconfig)
- [MSTeamsV2Config](#msteamsv2config)
- [OpsGenieConfig](#opsgenieconfig)
- [PagerDutyConfig](#pagerdutyconfig)
- [PushoverConfig](#pushoverconfig)
- [SNSConfig](#snsconfig)
- [SlackConfig](#slackconfig)
- [TelegramConfig](#telegramconfig)
- [VictorOpsConfig](#victoropsconfig)
- [WeChatConfig](#wechatconfig)
- [WebexConfig](#webexconfig)
- [WebhookConfig](#webhookconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header configuration for the client.<br />This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+. |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth for the client.<br />This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | OAuth2 client credentials used to fetch a token for the targets. |  |  |
| `bearerTokenSecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the bearer token to be used by the client<br />for authentication.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration for the client. |  |  |
| `proxyURL` _string_ | Optional proxy URL.<br />If defined, this field takes precedence over `proxyUrl`. |  |  |
| `followRedirects` _boolean_ | FollowRedirects specifies whether the client should follow HTTP 3xx redirects. |  |  |


#### HTTPSDConfig



HTTPSDConfig defines a prometheus HTTP service discovery configuration
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#http_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `url` _string_ | URL from which the targets are fetched. |  | MinLength: 1 <br />Pattern: `^http(s)?://.+$` <br /> |
| `refreshInterval` _[Duration](#duration)_ | RefreshInterval configures the refresh interval at which Prometheus will re-query the<br />endpoint to update the target list. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth information to authenticate against the target HTTP endpoint.<br />More info: https://prometheus.io/docs/operating/configuration/#endpoints<br />Cannot be set at the same time as `authorization`, or `oAuth2`. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header configuration to authenticate against the target HTTP endpoint.<br />Cannot be set at the same time as `oAuth2`, or `basicAuth`. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Optional OAuth 2.0 configuration to authenticate against the target HTTP endpoint.<br />Cannot be set at the same time as `authorization`, or `basicAuth`. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration applying to the target HTTP endpoint. |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2. |  |  |


#### HetznerSDConfig



HetznerSDConfig allow retrieving scrape targets from Hetzner Cloud API and Robot API.
This service discovery uses the public IPv4 address by default, but that can be changed with relabeling
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#hetzner_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `role` _string_ | The Hetzner role of entities that should be discovered. |  | Enum: [hcloud Hcloud robot Robot] <br /> |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth information to use on every scrape request, required when role is robot.<br />Role hcloud does not support basic auth. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header configuration, required when role is hcloud.<br />Role robot does not support bearer token authentication. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Optional OAuth 2.0 configuration.<br />Cannot be used at the same time as `basic_auth` or `authorization`. |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration to use on every scrape request. |  |  |
| `port` _integer_ | The port to scrape metrics from. |  |  |
| `refreshInterval` _[Duration](#duration)_ | The time after which the servers are refreshed. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |


#### InhibitRule



InhibitRule defines an inhibition rule that allows to mute alerts when other
alerts are already firing.
See https://prometheus.io/docs/alerting/latest/configuration/#inhibit_rule



_Appears in:_
- [AlertmanagerConfigSpec](#alertmanagerconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `targetMatch` _[Matcher](#matcher) array_ | Matchers that have to be fulfilled in the alerts to be muted. The<br />operator enforces that the alert matches the resource's namespace. |  |  |
| `sourceMatch` _[Matcher](#matcher) array_ | Matchers for which one or more alerts have to exist for the inhibition<br />to take effect. The operator enforces that the alert matches the<br />resource's namespace. |  |  |
| `equal` _string array_ | Labels that must have an equal value in the source and target alert for<br />the inhibition to take effect. |  |  |


#### IonosSDConfig



IonosSDConfig configurations allow retrieving scrape targets from IONOS resources.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#ionos_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `datacenterID` _string_ | The unique ID of the IONOS data center. |  | MinLength: 1 <br /> |
| `port` _integer_ | Port to scrape the metrics from. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `refreshInterval` _[Duration](#duration)_ | Refresh interval to re-read the list of resources. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization` header configuration, required when using IONOS. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration to use when connecting to the IONOS API. |  |  |
| `followRedirects` _boolean_ | Configure whether the HTTP requests should follow HTTP 3xx redirects. |  |  |
| `enableHTTP2` _boolean_ | Configure whether to enable HTTP2. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Configure whether to enable OAuth2. |  |  |


#### K8SSelectorConfig



K8SSelectorConfig is Kubernetes Selector Config



_Appears in:_
- [KubernetesSDConfig](#kubernetessdconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `role` _[KubernetesRole](#kubernetesrole)_ | Role specifies the type of Kubernetes resource to limit the service discovery to.<br />Accepted values are: Node, Pod, Endpoints, EndpointSlice, Service, Ingress. |  | Enum: [Pod Endpoints Ingress Service Node EndpointSlice] <br /> |
| `label` _string_ | An optional label selector to limit the service discovery to resources with specific labels and label values.<br />e.g: `node.kubernetes.io/instance-type=master` |  | MinLength: 1 <br /> |
| `field` _string_ | An optional field selector to limit the service discovery to resources which have fields with specific values.<br />e.g: `metadata.name=foobar` |  | MinLength: 1 <br /> |


#### KeyValue



KeyValue defines a (key, value) tuple.



_Appears in:_
- [EmailConfig](#emailconfig)
- [OpsGenieConfig](#opsgenieconfig)
- [PagerDutyConfig](#pagerdutyconfig)
- [VictorOpsConfig](#victoropsconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `key` _string_ | Key of the tuple. |  | MinLength: 1 <br /> |
| `value` _string_ | Value of the tuple. |  |  |


#### KubernetesRole

_Underlying type:_ _string_



_Validation:_
- Enum: [Pod Endpoints Ingress Service Node EndpointSlice]

_Appears in:_
- [K8SSelectorConfig](#k8sselectorconfig)
- [KubernetesSDConfig](#kubernetessdconfig)

| Field | Description |
| --- | --- |
| `Pod` |  |
| `Endpoints` |  |
| `Ingress` |  |
| `Service` |  |
| `Node` |  |
| `EndpointSlice` |  |


#### KubernetesSDConfig



KubernetesSDConfig allows retrieving scrape targets from Kubernetes' REST API.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiServer` _string_ | The API server address consisting of a hostname or IP address followed<br />by an optional port number.<br />If left empty, Prometheus is assumed to run inside<br />of the cluster. It will discover API servers automatically and use the pod's<br />CA certificate and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/. |  | MinLength: 1 <br /> |
| `role` _[KubernetesRole](#kubernetesrole)_ | Role of the Kubernetes entities that should be discovered.<br />Role `Endpointslice` requires Prometheus >= v2.21.0 |  | Enum: [Pod Endpoints Ingress Service Node EndpointSlice] <br /> |
| `namespaces` _[NamespaceDiscovery](#namespacediscovery)_ | Optional namespace discovery. If omitted, Prometheus discovers targets across all namespaces. |  |  |
| `attachMetadata` _[AttachMetadata](#attachmetadata)_ | Optional metadata to attach to discovered targets.<br />It requires Prometheus >= v2.35.0 when using the `Pod` role and<br />Prometheus >= v2.37.0 for `Endpoints` and `Endpointslice` roles. |  |  |
| `selectors` _[K8SSelectorConfig](#k8sselectorconfig) array_ | Selector to select objects.<br />It requires Prometheus >= v2.17.0 |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth information to use on every scrape request.<br />Cannot be set at the same time as `authorization`, or `oauth2`. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header to use on every scrape request.<br />Cannot be set at the same time as `basicAuth`, or `oauth2`. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Optional OAuth 2.0 configuration.<br />Cannot be set at the same time as `authorization`, or `basicAuth`. |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration to connect to the Kubernetes API. |  |  |


#### KumaSDConfig



KumaSDConfig allow retrieving scrape targets from Kuma's control plane.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kuma_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `server` _string_ | Address of the Kuma Control Plane's MADS xDS server. |  | MinLength: 1 <br /> |
| `clientID` _string_ | Client id is used by Kuma Control Plane to compute Monitoring Assignment for specific Prometheus backend. |  |  |
| `refreshInterval` _[Duration](#duration)_ | The time to wait between polling update requests. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `fetchTimeout` _[Duration](#duration)_ | The time after which the monitoring assignments are refreshed. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration to use on every scrape request |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth information to use on every scrape request. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header to use on every scrape request. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Optional OAuth 2.0 configuration.<br />Cannot be set at the same time as `authorization`, or `basicAuth`. |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2. |  |  |


#### LightSailSDConfig



LightSailSDConfig configurations allow retrieving scrape targets from AWS Lightsail instances.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#lightsail_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `region` _string_ | The AWS region. |  | MinLength: 1 <br /> |
| `accessKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | AccessKey is the AWS API key. |  |  |
| `secretKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | SecretKey is the AWS API secret. |  |  |
| `roleARN` _string_ | AWS Role ARN, an alternative to using AWS API keys. |  |  |
| `endpoint` _string_ | Custom endpoint to be used. |  | MinLength: 1 <br /> |
| `refreshInterval` _[Duration](#duration)_ | Refresh interval to re-read the list of instances. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `port` _integer_ | Port to scrape the metrics from.<br />If using the public IP address, this must instead be specified in the relabeling rule. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `basicAuth` _[BasicAuth](#basicauth)_ | Optional HTTP basic authentication information.<br />Cannot be set at the same time as `authorization`, or `oauth2`. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Optional `authorization` HTTP header configuration.<br />Cannot be set at the same time as `basicAuth`, or `oauth2`. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Optional OAuth2.0 configuration.<br />Cannot be set at the same time as `basicAuth`, or `authorization`. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration to connect to the Puppet DB. |  |  |
| `followRedirects` _boolean_ | Configure whether the HTTP requests should follow HTTP 3xx redirects. |  |  |
| `enableHTTP2` _boolean_ | Configure whether to enable HTTP2. |  |  |


#### LinodeSDConfig



LinodeSDConfig configurations allow retrieving scrape targets from Linode's Linode APIv4.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#linode_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `region` _string_ | Optional region to filter on. |  | MinLength: 1 <br /> |
| `port` _integer_ | Default port to scrape metrics from. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `tagSeparator` _string_ | The string by which Linode Instance tags are joined into the tag label. |  | MinLength: 1 <br /> |
| `refreshInterval` _[Duration](#duration)_ | Time after which the linode instances are refreshed. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header configuration. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Optional OAuth 2.0 configuration.<br />Cannot be used at the same time as `authorization`. |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration applying to the target HTTP endpoint. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2. |  |  |


#### MSTeamsConfig



MSTeamsConfig configures notifications via Microsoft Teams.
It requires Alertmanager >= 0.26.0.



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether to notify about resolved alerts. |  |  |
| `webhookUrl` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | MSTeams webhook URL. |  | Required: \{\} <br /> |
| `title` _string_ | Message title template. |  |  |
| `summary` _string_ | Message summary template.<br />It requires Alertmanager >= 0.27.0. |  |  |
| `text` _string_ | Message body template. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### MSTeamsV2Config



MSTeamsV2Config configures notifications via Microsoft Teams using the new message format with adaptive cards as required by flows
See https://prometheus.io/docs/alerting/latest/configuration/#msteamsv2_config
It requires Alertmanager >= 0.28.0.



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether to notify about resolved alerts. |  |  |
| `webhookURL` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | MSTeams incoming webhook URL. |  |  |
| `title` _string_ | Message title template. |  | MinLength: 1 <br /> |
| `text` _string_ | Message body template. |  | MinLength: 1 <br /> |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### MatchType

_Underlying type:_ _string_

MatchType is a comparison operator on a Matcher



_Appears in:_
- [Matcher](#matcher)

| Field | Description |
| --- | --- |
| `=` |  |
| `!=` |  |
| `=~` |  |
| `!~` |  |


#### Matcher



Matcher defines how to match on alert's labels.



_Appears in:_
- [InhibitRule](#inhibitrule)
- [Route](#route)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Label to match. |  | MinLength: 1 <br /> |
| `value` _string_ | Label value to match. |  |  |
| `matchType` _[MatchType](#matchtype)_ | Match operation available with AlertManager >= v0.22.0 and<br />takes precedence over Regex (deprecated) if non-empty. |  | Enum: [!= = =~ !~] <br /> |
| `regex` _boolean_ | Whether to match on equality (false) or regular-expression (true).<br />Deprecated: for AlertManager >= v0.22.0, `matchType` should be used instead. |  |  |




#### MonthRange

_Underlying type:_ _string_

MonthRange is an inclusive range of months of the year beginning in January
Months can be specified by name (e.g 'January') by numerical month (e.g '1') or as an inclusive range (e.g 'January:March', '1:3', '1:March')

_Validation:_
- Pattern: `^((?i)january|february|march|april|may|june|july|august|september|october|november|december|1[0-2]|[1-9])(?:((:((?i)january|february|march|april|may|june|july|august|september|october|november|december|1[0-2]|[1-9]))$)|$)`

_Appears in:_
- [TimeInterval](#timeinterval)



#### MuteTimeInterval



MuteTimeInterval specifies the periods in time when notifications will be muted



_Appears in:_
- [AlertmanagerConfigSpec](#alertmanagerconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the time interval |  | Required: \{\} <br /> |
| `timeIntervals` _[TimeInterval](#timeinterval) array_ | TimeIntervals is a list of TimeInterval |  |  |


#### NamespaceDiscovery



NamespaceDiscovery is the configuration for discovering
Kubernetes namespaces.



_Appears in:_
- [KubernetesSDConfig](#kubernetessdconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `ownNamespace` _boolean_ | Includes the namespace in which the Prometheus pod runs to the list of watched namespaces. |  |  |
| `names` _string array_ | List of namespaces where to watch for resources.<br />If empty and `ownNamespace` isn't true, Prometheus watches for resources in all namespaces. |  |  |


#### NomadSDConfig



NomadSDConfig configurations allow retrieving scrape targets from Nomad's Service API.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#nomad_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `allowStale` _boolean_ | The information to access the Nomad API. It is to be defined<br />as the Nomad documentation requires. |  |  |
| `namespace` _string_ |  |  |  |
| `refreshInterval` _[Duration](#duration)_ |  |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `region` _string_ |  |  |  |
| `server` _string_ |  |  | MinLength: 1 <br /> |
| `tagSeparator` _string_ |  |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth information to use on every scrape request. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header to use on every scrape request. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Optional OAuth 2.0 configuration.<br />Cannot be set at the same time as `authorization` or `basic_auth`. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration applying to the target HTTP endpoint. |  |  |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2. |  |  |


#### OVHCloudSDConfig



OVHCloudSDConfig configurations allow retrieving scrape targets from OVHcloud's dedicated servers and VPS using their API.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#ovhcloud_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `applicationKey` _string_ | Access key to use. https://api.ovh.com. |  | MinLength: 1 <br /> |
| `applicationSecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ |  |  |  |
| `consumerKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ |  |  |  |
| `service` _[OVHService](#ovhservice)_ | Service of the targets to retrieve. Must be `VPS` or `DedicatedServer`. |  | Enum: [VPS DedicatedServer] <br /> |
| `endpoint` _string_ | Custom endpoint to be used. |  | MinLength: 1 <br /> |
| `refreshInterval` _[Duration](#duration)_ | Refresh interval to re-read the resources list. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |


#### OVHService

_Underlying type:_ _string_

Service of the targets to retrieve. Must be `VPS` or `DedicatedServer`.

_Validation:_
- Enum: [VPS DedicatedServer]

_Appears in:_
- [OVHCloudSDConfig](#ovhcloudsdconfig)

| Field | Description |
| --- | --- |
| `VPS` |  |
| `DedicatedServer` |  |


#### OpenStackRole

_Underlying type:_ _string_



_Validation:_
- Enum: [Instance Hypervisor LoadBalancer]

_Appears in:_
- [OpenStackSDConfig](#openstacksdconfig)

| Field | Description |
| --- | --- |
| `Instance` |  |
| `Hypervisor` |  |
| `LoadBalancer` |  |


#### OpenStackSDConfig



OpenStackSDConfig allow retrieving scrape targets from OpenStack Nova instances.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#openstack_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `role` _[OpenStackRole](#openstackrole)_ | The OpenStack role of entities that should be discovered.<br />Note: The `LoadBalancer` role requires Prometheus >= v3.2.0. |  | Enum: [Instance Hypervisor LoadBalancer] <br /> |
| `region` _string_ | The OpenStack Region. |  | MinLength: 1 <br /> |
| `identityEndpoint` _string_ | IdentityEndpoint specifies the HTTP endpoint that is required to work with<br />the Identity API of the appropriate version. |  | Pattern: `^http(s)?:\/\/.+$` <br /> |
| `username` _string_ | Username is required if using Identity V2 API. Consult with your provider's<br />control panel to discover your account's username.<br />In Identity V3, either userid or a combination of username<br />and domainId or domainName are needed |  | MinLength: 1 <br /> |
| `userid` _string_ | UserID |  | MinLength: 1 <br /> |
| `password` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Password for the Identity V2 and V3 APIs. Consult with your provider's<br />control panel to discover your account's preferred method of authentication. |  |  |
| `domainName` _string_ | At most one of domainId and domainName must be provided if using username<br />with Identity V3. Otherwise, either are optional. |  | MinLength: 1 <br /> |
| `domainID` _string_ | DomainID |  | MinLength: 1 <br /> |
| `projectName` _string_ | The ProjectId and ProjectName fields are optional for the Identity V2 API.<br />Some providers allow you to specify a ProjectName instead of the ProjectId.<br />Some require both. Your provider's authentication policies will determine<br />how these fields influence authentication. |  | MinLength: 1 <br /> |
| `projectID` _string_ |  ProjectID |  | MinLength: 1 <br /> |
| `applicationCredentialName` _string_ | The ApplicationCredentialID or ApplicationCredentialName fields are<br />required if using an application credential to authenticate. Some providers<br />allow you to create an application credential to authenticate rather than a<br />password. |  | MinLength: 1 <br /> |
| `applicationCredentialId` _string_ | ApplicationCredentialID |  |  |
| `applicationCredentialSecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The applicationCredentialSecret field is required if using an application<br />credential to authenticate. |  |  |
| `allTenants` _boolean_ | Whether the service discovery should list all instances for all projects.<br />It is only relevant for the 'instance' role and usually requires admin permissions. |  |  |
| `refreshInterval` _[Duration](#duration)_ | Refresh interval to re-read the instance list. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `port` _integer_ | The port to scrape metrics from. If using the public IP address, this must<br />instead be specified in the relabeling rule. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `availability` _string_ | Availability of the endpoint to connect to. |  | Enum: [Public public Admin admin Internal internal] <br /> |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration applying to the target HTTP endpoint. |  |  |


#### OpsGenieConfig



OpsGenieConfig configures notifications via OpsGenie.
See https://prometheus.io/docs/alerting/latest/configuration/#opsgenie_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `apiKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the OpsGenie API key.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `apiURL` _string_ | The URL to send OpsGenie API requests to. |  |  |
| `message` _string_ | Alert text limited to 130 characters. |  |  |
| `description` _string_ | Description of the incident. |  |  |
| `source` _string_ | Backlink to the sender of the notification. |  |  |
| `tags` _string_ | Comma separated list of tags attached to the notifications. |  |  |
| `note` _string_ | Additional alert note. |  |  |
| `priority` _string_ | Priority level of alert. Possible values are P1, P2, P3, P4, and P5. |  |  |
| `updateAlerts` _boolean_ | Whether to update message and description of the alert in OpsGenie if it already exists<br />By default, the alert is never updated in OpsGenie, the new message only appears in activity log. |  |  |
| `details` _[KeyValue](#keyvalue) array_ | A set of arbitrary key/value pairs that provide further detail about the incident. |  |  |
| `responders` _[OpsGenieConfigResponder](#opsgenieconfigresponder) array_ | List of responders responsible for notifications. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |
| `entity` _string_ | Optional field that can be used to specify which domain alert is related to. |  |  |
| `actions` _string_ | Comma separated list of actions that will be available for the alert. |  |  |


#### OpsGenieConfigResponder



OpsGenieConfigResponder defines a responder to an incident.
One of `id`, `name` or `username` has to be defined.



_Appears in:_
- [OpsGenieConfig](#opsgenieconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `id` _string_ | ID of the responder. |  |  |
| `name` _string_ | Name of the responder. |  |  |
| `username` _string_ | Username of the responder. |  |  |
| `type` _string_ | Type of responder. |  | MinLength: 1 <br /> |


#### PagerDutyConfig



PagerDutyConfig configures notifications via PagerDuty.
See https://prometheus.io/docs/alerting/latest/configuration/#pagerduty_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `routingKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the PagerDuty integration key (when using<br />Events API v2). Either this field or `serviceKey` needs to be defined.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `serviceKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the PagerDuty service key (when using<br />integration type "Prometheus"). Either this field or `routingKey` needs to<br />be defined.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `url` _string_ | The URL to send requests to. |  |  |
| `client` _string_ | Client identification. |  |  |
| `clientURL` _string_ | Backlink to the sender of notification. |  |  |
| `description` _string_ | Description of the incident. |  |  |
| `severity` _string_ | Severity of the incident. |  |  |
| `class` _string_ | The class/type of the event. |  |  |
| `group` _string_ | A cluster or grouping of sources. |  |  |
| `component` _string_ | The part or component of the affected system that is broken. |  |  |
| `details` _[KeyValue](#keyvalue) array_ | Arbitrary key/value pairs that provide further detail about the incident. |  |  |
| `pagerDutyImageConfigs` _[PagerDutyImageConfig](#pagerdutyimageconfig) array_ | A list of image details to attach that provide further detail about an incident. |  |  |
| `pagerDutyLinkConfigs` _[PagerDutyLinkConfig](#pagerdutylinkconfig) array_ | A list of link details to attach that provide further detail about an incident. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |
| `source` _string_ | Unique location of the affected system. |  |  |


#### PagerDutyImageConfig



PagerDutyImageConfig attaches images to an incident



_Appears in:_
- [PagerDutyConfig](#pagerdutyconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `src` _string_ | Src of the image being attached to the incident |  |  |
| `href` _string_ | Optional URL; makes the image a clickable link. |  |  |
| `alt` _string_ | Alt is the optional alternative text for the image. |  |  |


#### PagerDutyLinkConfig



PagerDutyLinkConfig attaches text links to an incident



_Appears in:_
- [PagerDutyConfig](#pagerdutyconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `href` _string_ | Href is the URL of the link to be attached |  |  |
| `alt` _string_ | Text that describes the purpose of the link, and can be used as the link's text. |  |  |




#### PrometheusAgent



The `PrometheusAgent` custom resource definition (CRD) defines a desired [Prometheus Agent](https://prometheus.io/blog/2021/11/16/agent/) setup to run in a Kubernetes cluster.

The CRD is very similar to the `Prometheus` CRD except for features which aren't available in agent mode like rule evaluation, persistent storage and Thanos sidecar.



_Appears in:_
- [PrometheusAgentList](#prometheusagentlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[PrometheusAgentSpec](#prometheusagentspec)_ | Specification of the desired behavior of the Prometheus agent. More info:<br />https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status |  |  |
| `status` _[PrometheusStatus](#prometheusstatus)_ | Most recent observed status of the Prometheus cluster. Read-only.<br />More info:<br />https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status |  |  |




#### PrometheusAgentMode

_Underlying type:_ _string_



_Validation:_
- Enum: [StatefulSet DaemonSet]

_Appears in:_
- [PrometheusAgentSpec](#prometheusagentspec)

| Field | Description |
| --- | --- |
| `DaemonSet` | Deploys PrometheusAgent as DaemonSet.<br /> |
| `StatefulSet` | Deploys PrometheusAgent as StatefulSet.<br /> |


#### PrometheusAgentSpec



PrometheusAgentSpec is a specification of the desired behavior of the Prometheus agent. More info:
https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status



_Appears in:_
- [PrometheusAgent](#prometheusagent)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `mode` _[PrometheusAgentMode](#prometheusagentmode)_ | Mode defines how the Prometheus operator deploys the PrometheusAgent pod(s).<br />(Alpha) Using this field requires the `PrometheusAgentDaemonSet` feature gate to be enabled. |  | Enum: [StatefulSet DaemonSet] <br /> |
| `podMetadata` _[EmbeddedObjectMetadata](#embeddedobjectmetadata)_ | PodMetadata configures labels and annotations which are propagated to the Prometheus pods.<br />The following items are reserved and cannot be overridden:<br />* "prometheus" label, set to the name of the Prometheus object.<br />* "app.kubernetes.io/instance" label, set to the name of the Prometheus object.<br />* "app.kubernetes.io/managed-by" label, set to "prometheus-operator".<br />* "app.kubernetes.io/name" label, set to "prometheus".<br />* "app.kubernetes.io/version" label, set to the Prometheus version.<br />* "operator.prometheus.io/name" label, set to the name of the Prometheus object.<br />* "operator.prometheus.io/shard" label, set to the shard number of the Prometheus object.<br />* "kubectl.kubernetes.io/default-container" annotation, set to "prometheus". |  |  |
| `serviceMonitorSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | ServiceMonitors to be selected for target discovery. An empty label<br />selector matches all objects. A null label selector matches no objects.<br />If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`<br />and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.<br />The Prometheus operator will ensure that the Prometheus configuration's<br />Secret exists, but it is the responsibility of the user to provide the raw<br />gzipped Prometheus configuration under the `prometheus.yaml.gz` key.<br />This behavior is *deprecated* and will be removed in the next major version<br />of the custom resource definition. It is recommended to use<br />`spec.additionalScrapeConfigs` instead. |  |  |
| `serviceMonitorNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to match for ServicedMonitors discovery. An empty label selector<br />matches all namespaces. A null label selector (default value) matches the current<br />namespace only. |  |  |
| `podMonitorSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | PodMonitors to be selected for target discovery. An empty label selector<br />matches all objects. A null label selector matches no objects.<br />If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`<br />and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.<br />The Prometheus operator will ensure that the Prometheus configuration's<br />Secret exists, but it is the responsibility of the user to provide the raw<br />gzipped Prometheus configuration under the `prometheus.yaml.gz` key.<br />This behavior is *deprecated* and will be removed in the next major version<br />of the custom resource definition. It is recommended to use<br />`spec.additionalScrapeConfigs` instead. |  |  |
| `podMonitorNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to match for PodMonitors discovery. An empty label selector<br />matches all namespaces. A null label selector (default value) matches the current<br />namespace only. |  |  |
| `probeSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Probes to be selected for target discovery. An empty label selector<br />matches all objects. A null label selector matches no objects.<br />If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`<br />and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.<br />The Prometheus operator will ensure that the Prometheus configuration's<br />Secret exists, but it is the responsibility of the user to provide the raw<br />gzipped Prometheus configuration under the `prometheus.yaml.gz` key.<br />This behavior is *deprecated* and will be removed in the next major version<br />of the custom resource definition. It is recommended to use<br />`spec.additionalScrapeConfigs` instead. |  |  |
| `probeNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to match for Probe discovery. An empty label<br />selector matches all namespaces. A null label selector matches the<br />current namespace only. |  |  |
| `scrapeConfigSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | ScrapeConfigs to be selected for target discovery. An empty label<br />selector matches all objects. A null label selector matches no objects.<br />If `spec.serviceMonitorSelector`, `spec.podMonitorSelector`, `spec.probeSelector`<br />and `spec.scrapeConfigSelector` are null, the Prometheus configuration is unmanaged.<br />The Prometheus operator will ensure that the Prometheus configuration's<br />Secret exists, but it is the responsibility of the user to provide the raw<br />gzipped Prometheus configuration under the `prometheus.yaml.gz` key.<br />This behavior is *deprecated* and will be removed in the next major version<br />of the custom resource definition. It is recommended to use<br />`spec.additionalScrapeConfigs` instead.<br />Note that the ScrapeConfig custom resource definition is currently at Alpha level. |  |  |
| `scrapeConfigNamespaceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#labelselector-v1-meta)_ | Namespaces to match for ScrapeConfig discovery. An empty label selector<br />matches all namespaces. A null label selector matches the current<br />namespace only.<br />Note that the ScrapeConfig custom resource definition is currently at Alpha level. |  |  |
| `version` _string_ | Version of Prometheus being deployed. The operator uses this information<br />to generate the Prometheus StatefulSet + configuration files.<br />If not specified, the operator assumes the latest upstream version of<br />Prometheus available at the time when the version of the operator was<br />released. |  |  |
| `paused` _boolean_ | When a Prometheus deployment is paused, no actions except for deletion<br />will be performed on the underlying objects. |  |  |
| `image` _string_ | Container image name for Prometheus. If specified, it takes precedence<br />over the `spec.baseImage`, `spec.tag` and `spec.sha` fields.<br />Specifying `spec.version` is still necessary to ensure the Prometheus<br />Operator knows which version of Prometheus is being configured.<br />If neither `spec.image` nor `spec.baseImage` are defined, the operator<br />will use the latest upstream version of Prometheus available at the time<br />when the operator was released. |  |  |
| `imagePullPolicy` _[PullPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#pullpolicy-v1-core)_ | Image pull policy for the 'prometheus', 'init-config-reloader' and 'config-reloader' containers.<br />See https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy for more details. |  | Enum: [ Always Never IfNotPresent] <br /> |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#localobjectreference-v1-core) array_ | An optional list of references to Secrets in the same namespace<br />to use for pulling images from registries.<br />See http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod |  |  |
| `replicas` _integer_ | Number of replicas of each shard to deploy for a Prometheus deployment.<br />`spec.replicas` multiplied by `spec.shards` is the total number of Pods<br />created.<br />Default: 1 |  |  |
| `shards` _integer_ | Number of shards to distribute the scraped targets onto.<br />`spec.replicas` multiplied by `spec.shards` is the total number of Pods<br />being created.<br />When not defined, the operator assumes only one shard.<br />Note that scaling down shards will not reshard data onto the remaining<br />instances, it must be manually moved. Increasing shards will not reshard<br />data either but it will continue to be available from the same<br />instances. To query globally, use either<br />* Thanos sidecar + querier for query federation and Thanos Ruler for rules.<br />* Remote-write to send metrics to a central location.<br />By default, the sharding of targets is performed on:<br />* The `__address__` target's metadata label for PodMonitor,<br />ServiceMonitor and ScrapeConfig resources.<br />* The `__param_target__` label for Probe resources.<br />Users can define their own sharding implementation by setting the<br />`__tmp_hash` label during the target discovery with relabeling<br />configuration (either in the monitoring resources or via scrape class).<br />You can also disable sharding on a specific target by setting the<br />`__tmp_disable_sharding` label with relabeling configuration. When<br />the label value isn't empty, all Prometheus shards will scrape the target. |  |  |
| `replicaExternalLabelName` _string_ | Name of Prometheus external label used to denote the replica name.<br />The external label will _not_ be added when the field is set to the<br />empty string (`""`).<br />Default: "prometheus_replica" |  |  |
| `prometheusExternalLabelName` _string_ | Name of Prometheus external label used to denote the Prometheus instance<br />name. The external label will _not_ be added when the field is set to<br />the empty string (`""`).<br />Default: "prometheus" |  |  |
| `logLevel` _string_ | Log level for Prometheus and the config-reloader sidecar. |  | Enum: [ debug info warn error] <br /> |
| `logFormat` _string_ | Log format for Log level for Prometheus and the config-reloader sidecar. |  | Enum: [ logfmt json] <br /> |
| `scrapeInterval` _[Duration](#duration)_ | Interval between consecutive scrapes.<br />Default: "30s" | 30s | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `scrapeTimeout` _[Duration](#duration)_ | Number of seconds to wait until a scrape request times out.<br />The value cannot be greater than the scrape interval otherwise the operator will reject the resource. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `scrapeProtocols` _[ScrapeProtocol](#scrapeprotocol) array_ | The protocols to negotiate during a scrape. It tells clients the<br />protocols supported by Prometheus in order of preference (from most to least preferred).<br />If unset, Prometheus uses its default value.<br />It requires Prometheus >= v2.49.0.<br />`PrometheusText1.0.0` requires Prometheus >= v3.0.0. |  | Enum: [PrometheusProto OpenMetricsText0.0.1 OpenMetricsText1.0.0 PrometheusText0.0.4 PrometheusText1.0.0] <br /> |
| `externalLabels` _object (keys:string, values:string)_ | The labels to add to any time series or alerts when communicating with<br />external systems (federation, remote storage, Alertmanager).<br />Labels defined by `spec.replicaExternalLabelName` and<br />`spec.prometheusExternalLabelName` take precedence over this list. |  |  |
| `enableRemoteWriteReceiver` _boolean_ | Enable Prometheus to be used as a receiver for the Prometheus remote<br />write protocol.<br />WARNING: This is not considered an efficient way of ingesting samples.<br />Use it with caution for specific low-volume use cases.<br />It is not suitable for replacing the ingestion via scraping and turning<br />Prometheus into a push-based metrics collection system.<br />For more information see https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver<br />It requires Prometheus >= v2.33.0. |  |  |
| `enableOTLPReceiver` _boolean_ | Enable Prometheus to be used as a receiver for the OTLP Metrics protocol.<br />Note that the OTLP receiver endpoint is automatically enabled if `.spec.otlpConfig` is defined.<br />It requires Prometheus >= v2.47.0. |  |  |
| `remoteWriteReceiverMessageVersions` _[RemoteWriteMessageVersion](#remotewritemessageversion) array_ | List of the protobuf message versions to accept when receiving the<br />remote writes.<br />It requires Prometheus >= v2.54.0. |  | Enum: [V1.0 V2.0] <br />MinItems: 1 <br /> |
| `enableFeatures` _[EnableFeature](#enablefeature) array_ | Enable access to Prometheus feature flags. By default, no features are enabled.<br />Enabling features which are disabled by default is entirely outside the<br />scope of what the maintainers will support and by doing so, you accept<br />that this behaviour may break at any time without notice.<br />For more information see https://prometheus.io/docs/prometheus/latest/feature_flags/ |  | MinLength: 1 <br /> |
| `externalUrl` _string_ | The external URL under which the Prometheus service is externally<br />available. This is necessary to generate correct URLs (for instance if<br />Prometheus is accessible behind an Ingress resource). |  |  |
| `routePrefix` _string_ | The route prefix Prometheus registers HTTP handlers for.<br />This is useful when using `spec.externalURL`, and a proxy is rewriting<br />HTTP routes of a request, and the actual ExternalURL is still true, but<br />the server serves requests under a different route prefix. For example<br />for use with `kubectl proxy`. |  |  |
| `storage` _[StorageSpec](#storagespec)_ | Storage defines the storage used by Prometheus. |  |  |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volume-v1-core) array_ | Volumes allows the configuration of additional volumes on the output<br />StatefulSet definition. Volumes specified will be appended to other<br />volumes that are generated as a result of StorageSpec objects. |  |  |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#volumemount-v1-core) array_ | VolumeMounts allows the configuration of additional VolumeMounts.<br />VolumeMounts will be appended to other VolumeMounts in the 'prometheus'<br />container, that are generated as a result of StorageSpec objects. |  |  |
| `persistentVolumeClaimRetentionPolicy` _[StatefulSetPersistentVolumeClaimRetentionPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#statefulsetpersistentvolumeclaimretentionpolicy-v1-apps)_ | The field controls if and how PVCs are deleted during the lifecycle of a StatefulSet.<br />The default behavior is all PVCs are retained.<br />This is an alpha field from kubernetes 1.23 until 1.26 and a beta field from 1.26.<br />It requires enabling the StatefulSetAutoDeletePVC feature gate. |  |  |
| `web` _[PrometheusWebSpec](#prometheuswebspec)_ | Defines the configuration of the Prometheus web server. |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#resourcerequirements-v1-core)_ | Defines the resources requests and limits of the 'prometheus' container. |  |  |
| `nodeSelector` _object (keys:string, values:string)_ | Defines on which Nodes the Pods are scheduled. |  |  |
| `serviceAccountName` _string_ | ServiceAccountName is the name of the ServiceAccount to use to run the<br />Prometheus Pods. |  |  |
| `automountServiceAccountToken` _boolean_ | AutomountServiceAccountToken indicates whether a service account token should be automatically mounted in the pod.<br />If the field isn't set, the operator mounts the service account token by default.<br />**Warning:** be aware that by default, Prometheus requires the service account token for Kubernetes service discovery.<br />It is possible to use strategic merge patch to project the service account token into the 'prometheus' container. |  |  |
| `secrets` _string array_ | Secrets is a list of Secrets in the same namespace as the Prometheus<br />object, which shall be mounted into the Prometheus Pods.<br />Each Secret is added to the StatefulSet definition as a volume named `secret-<secret-name>`.<br />The Secrets are mounted into /etc/prometheus/secrets/<secret-name> in the 'prometheus' container. |  |  |
| `configMaps` _string array_ | ConfigMaps is a list of ConfigMaps in the same namespace as the Prometheus<br />object, which shall be mounted into the Prometheus Pods.<br />Each ConfigMap is added to the StatefulSet definition as a volume named `configmap-<configmap-name>`.<br />The ConfigMaps are mounted into /etc/prometheus/configmaps/<configmap-name> in the 'prometheus' container. |  |  |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#affinity-v1-core)_ | Defines the Pods' affinity scheduling rules if specified. |  |  |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#toleration-v1-core) array_ | Defines the Pods' tolerations if specified. |  |  |
| `topologySpreadConstraints` _[TopologySpreadConstraint](#topologyspreadconstraint) array_ | Defines the pod's topology spread constraints if specified. |  |  |
| `remoteWrite` _[RemoteWriteSpec](#remotewritespec) array_ | Defines the list of remote write configurations. |  |  |
| `otlp` _[OTLPConfig](#otlpconfig)_ | Settings related to the OTLP receiver feature.<br />It requires Prometheus >= v2.55.0. |  |  |
| `securityContext` _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#podsecuritycontext-v1-core)_ | SecurityContext holds pod-level security attributes and common container settings.<br />This defaults to the default PodSecurityContext. |  |  |
| `dnsPolicy` _[DNSPolicy](#dnspolicy)_ | Defines the DNS policy for the pods. |  | Enum: [ClusterFirstWithHostNet ClusterFirst Default None] <br /> |
| `dnsConfig` _[PodDNSConfig](#poddnsconfig)_ | Defines the DNS configuration for the pods. |  |  |
| `listenLocal` _boolean_ | When true, the Prometheus server listens on the loopback address<br />instead of the Pod IP's address. |  |  |
| `enableServiceLinks` _boolean_ | Indicates whether information about services should be injected into pod's environment variables |  |  |
| `containers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core) array_ | Containers allows injecting additional containers or modifying operator<br />generated containers. This can be used to allow adding an authentication<br />proxy to the Pods or to change the behavior of an operator generated<br />container. Containers described here modify an operator generated<br />container if they share the same name and modifications are done via a<br />strategic merge patch.<br />The names of containers managed by the operator are:<br />* `prometheus`<br />* `config-reloader`<br />* `thanos-sidecar`<br />Overriding containers is entirely outside the scope of what the<br />maintainers will support and by doing so, you accept that this behaviour<br />may break at any time without notice. |  |  |
| `initContainers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#container-v1-core) array_ | InitContainers allows injecting initContainers to the Pod definition. Those<br />can be used to e.g.  fetch secrets for injection into the Prometheus<br />configuration from external sources. Any errors during the execution of<br />an initContainer will lead to a restart of the Pod. More info:<br />https://kubernetes.io/docs/concepts/workloads/pods/init-containers/<br />InitContainers described here modify an operator generated init<br />containers if they share the same name and modifications are done via a<br />strategic merge patch.<br />The names of init container name managed by the operator are:<br />* `init-config-reloader`.<br />Overriding init containers is entirely outside the scope of what the<br />maintainers will support and by doing so, you accept that this behaviour<br />may break at any time without notice. |  |  |
| `additionalScrapeConfigs` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | AdditionalScrapeConfigs allows specifying a key of a Secret containing<br />additional Prometheus scrape configurations. Scrape configurations<br />specified are appended to the configurations generated by the Prometheus<br />Operator. Job configurations specified must have the form as specified<br />in the official Prometheus documentation:<br />https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config.<br />As scrape configs are appended, the user is responsible to make sure it<br />is valid. Note that using this feature may expose the possibility to<br />break upgrades of Prometheus. It is advised to review Prometheus release<br />notes to ensure that no incompatible scrape configs are going to break<br />Prometheus after the upgrade. |  |  |
| `apiserverConfig` _[APIServerConfig](#apiserverconfig)_ | APIServerConfig allows specifying a host and auth methods to access the<br />Kuberntees API server.<br />If null, Prometheus is assumed to run inside of the cluster: it will<br />discover the API servers automatically and use the Pod's CA certificate<br />and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/. |  |  |
| `priorityClassName` _string_ | Priority class assigned to the Pods. |  |  |
| `portName` _string_ | Port name used for the pods and governing service.<br />Default: "web" | web |  |
| `arbitraryFSAccessThroughSMs` _[ArbitraryFSAccessThroughSMsConfig](#arbitraryfsaccessthroughsmsconfig)_ | When true, ServiceMonitor, PodMonitor and Probe object are forbidden to<br />reference arbitrary files on the file system of the 'prometheus'<br />container.<br />When a ServiceMonitor's endpoint specifies a `bearerTokenFile` value<br />(e.g.  '/var/run/secrets/kubernetes.io/serviceaccount/token'), a<br />malicious target can get access to the Prometheus service account's<br />token in the Prometheus' scrape request. Setting<br />`spec.arbitraryFSAccessThroughSM` to 'true' would prevent the attack.<br />Users should instead provide the credentials using the<br />`spec.bearerTokenSecret` field. |  |  |
| `overrideHonorLabels` _boolean_ | When true, Prometheus resolves label conflicts by renaming the labels in the scraped data<br /> to “exported_” for all targets created from ServiceMonitor, PodMonitor and<br />ScrapeConfig objects. Otherwise the HonorLabels field of the service or pod monitor applies.<br />In practice,`overrideHonorLaels:true` enforces `honorLabels:false`<br />for all ServiceMonitor, PodMonitor and ScrapeConfig objects. |  |  |
| `overrideHonorTimestamps` _boolean_ | When true, Prometheus ignores the timestamps for all the targets created<br />from service and pod monitors.<br />Otherwise the HonorTimestamps field of the service or pod monitor applies. |  |  |
| `ignoreNamespaceSelectors` _boolean_ | When true, `spec.namespaceSelector` from all PodMonitor, ServiceMonitor<br />and Probe objects will be ignored. They will only discover targets<br />within the namespace of the PodMonitor, ServiceMonitor and Probe<br />object. |  |  |
| `enforcedNamespaceLabel` _string_ | When not empty, a label will be added to:<br />1. All metrics scraped from `ServiceMonitor`, `PodMonitor`, `Probe` and `ScrapeConfig` objects.<br />2. All metrics generated from recording rules defined in `PrometheusRule` objects.<br />3. All alerts generated from alerting rules defined in `PrometheusRule` objects.<br />4. All vector selectors of PromQL expressions defined in `PrometheusRule` objects.<br />The label will not added for objects referenced in `spec.excludedFromEnforcement`.<br />The label's name is this field's value.<br />The label's value is the namespace of the `ServiceMonitor`,<br />`PodMonitor`, `Probe`, `PrometheusRule` or `ScrapeConfig` object. |  |  |
| `enforcedSampleLimit` _integer_ | When defined, enforcedSampleLimit specifies a global limit on the number<br />of scraped samples that will be accepted. This overrides any<br />`spec.sampleLimit` set by ServiceMonitor, PodMonitor, Probe objects<br />unless `spec.sampleLimit` is greater than zero and less than<br />`spec.enforcedSampleLimit`.<br />It is meant to be used by admins to keep the overall number of<br />samples/series under a desired limit.<br />When both `enforcedSampleLimit` and `sampleLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined sampleLimit value will inherit the global sampleLimit value (Prometheus >= 2.45.0) or the enforcedSampleLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedSampleLimit` is greater than the `sampleLimit`, the `sampleLimit` will be set to `enforcedSampleLimit`.<br />* Scrape objects with a sampleLimit value less than or equal to enforcedSampleLimit keep their specific value.<br />* Scrape objects with a sampleLimit value greater than enforcedSampleLimit are set to enforcedSampleLimit. |  |  |
| `enforcedTargetLimit` _integer_ | When defined, enforcedTargetLimit specifies a global limit on the number<br />of scraped targets. The value overrides any `spec.targetLimit` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.targetLimit` is<br />greater than zero and less than `spec.enforcedTargetLimit`.<br />It is meant to be used by admins to to keep the overall number of<br />targets under a desired limit.<br />When both `enforcedTargetLimit` and `targetLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined targetLimit value will inherit the global targetLimit value (Prometheus >= 2.45.0) or the enforcedTargetLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedTargetLimit` is greater than the `targetLimit`, the `targetLimit` will be set to `enforcedTargetLimit`.<br />* Scrape objects with a targetLimit value less than or equal to enforcedTargetLimit keep their specific value.<br />* Scrape objects with a targetLimit value greater than enforcedTargetLimit are set to enforcedTargetLimit. |  |  |
| `enforcedLabelLimit` _integer_ | When defined, enforcedLabelLimit specifies a global limit on the number<br />of labels per sample. The value overrides any `spec.labelLimit` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.labelLimit` is<br />greater than zero and less than `spec.enforcedLabelLimit`.<br />It requires Prometheus >= v2.27.0.<br />When both `enforcedLabelLimit` and `labelLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined labelLimit value will inherit the global labelLimit value (Prometheus >= 2.45.0) or the enforcedLabelLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedLabelLimit` is greater than the `labelLimit`, the `labelLimit` will be set to `enforcedLabelLimit`.<br />* Scrape objects with a labelLimit value less than or equal to enforcedLabelLimit keep their specific value.<br />* Scrape objects with a labelLimit value greater than enforcedLabelLimit are set to enforcedLabelLimit. |  |  |
| `enforcedLabelNameLengthLimit` _integer_ | When defined, enforcedLabelNameLengthLimit specifies a global limit on the length<br />of labels name per sample. The value overrides any `spec.labelNameLengthLimit` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.labelNameLengthLimit` is<br />greater than zero and less than `spec.enforcedLabelNameLengthLimit`.<br />It requires Prometheus >= v2.27.0.<br />When both `enforcedLabelNameLengthLimit` and `labelNameLengthLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined labelNameLengthLimit value will inherit the global labelNameLengthLimit value (Prometheus >= 2.45.0) or the enforcedLabelNameLengthLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedLabelNameLengthLimit` is greater than the `labelNameLengthLimit`, the `labelNameLengthLimit` will be set to `enforcedLabelNameLengthLimit`.<br />* Scrape objects with a labelNameLengthLimit value less than or equal to enforcedLabelNameLengthLimit keep their specific value.<br />* Scrape objects with a labelNameLengthLimit value greater than enforcedLabelNameLengthLimit are set to enforcedLabelNameLengthLimit. |  |  |
| `enforcedLabelValueLengthLimit` _integer_ | When not null, enforcedLabelValueLengthLimit defines a global limit on the length<br />of labels value per sample. The value overrides any `spec.labelValueLengthLimit` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.labelValueLengthLimit` is<br />greater than zero and less than `spec.enforcedLabelValueLengthLimit`.<br />It requires Prometheus >= v2.27.0.<br />When both `enforcedLabelValueLengthLimit` and `labelValueLengthLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined labelValueLengthLimit value will inherit the global labelValueLengthLimit value (Prometheus >= 2.45.0) or the enforcedLabelValueLengthLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedLabelValueLengthLimit` is greater than the `labelValueLengthLimit`, the `labelValueLengthLimit` will be set to `enforcedLabelValueLengthLimit`.<br />* Scrape objects with a labelValueLengthLimit value less than or equal to enforcedLabelValueLengthLimit keep their specific value.<br />* Scrape objects with a labelValueLengthLimit value greater than enforcedLabelValueLengthLimit are set to enforcedLabelValueLengthLimit. |  |  |
| `enforcedKeepDroppedTargets` _integer_ | When defined, enforcedKeepDroppedTargets specifies a global limit on the number of targets<br />dropped by relabeling that will be kept in memory. The value overrides<br />any `spec.keepDroppedTargets` set by<br />ServiceMonitor, PodMonitor, Probe objects unless `spec.keepDroppedTargets` is<br />greater than zero and less than `spec.enforcedKeepDroppedTargets`.<br />It requires Prometheus >= v2.47.0.<br />When both `enforcedKeepDroppedTargets` and `keepDroppedTargets` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined keepDroppedTargets value will inherit the global keepDroppedTargets value (Prometheus >= 2.45.0) or the enforcedKeepDroppedTargets value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedKeepDroppedTargets` is greater than the `keepDroppedTargets`, the `keepDroppedTargets` will be set to `enforcedKeepDroppedTargets`.<br />* Scrape objects with a keepDroppedTargets value less than or equal to enforcedKeepDroppedTargets keep their specific value.<br />* Scrape objects with a keepDroppedTargets value greater than enforcedKeepDroppedTargets are set to enforcedKeepDroppedTargets. |  |  |
| `enforcedBodySizeLimit` _[ByteSize](#bytesize)_ | When defined, enforcedBodySizeLimit specifies a global limit on the size<br />of uncompressed response body that will be accepted by Prometheus.<br />Targets responding with a body larger than this many bytes will cause<br />the scrape to fail.<br />It requires Prometheus >= v2.28.0.<br />When both `enforcedBodySizeLimit` and `bodySizeLimit` are defined and greater than zero, the following rules apply:<br />* Scrape objects without a defined bodySizeLimit value will inherit the global bodySizeLimit value (Prometheus >= 2.45.0) or the enforcedBodySizeLimit value (Prometheus < v2.45.0).<br />  If Prometheus version is >= 2.45.0 and the `enforcedBodySizeLimit` is greater than the `bodySizeLimit`, the `bodySizeLimit` will be set to `enforcedBodySizeLimit`.<br />* Scrape objects with a bodySizeLimit value less than or equal to enforcedBodySizeLimit keep their specific value.<br />* Scrape objects with a bodySizeLimit value greater than enforcedBodySizeLimit are set to enforcedBodySizeLimit. |  | Pattern: `(^0\|([0-9]*[.])?[0-9]+((K\|M\|G\|T\|E\|P)i?)?B)$` <br /> |
| `nameValidationScheme` _[NameValidationSchemeOptions](#namevalidationschemeoptions)_ | Specifies the validation scheme for metric and label names.<br />It requires Prometheus >= v2.55.0. |  | Enum: [UTF8 Legacy] <br /> |
| `nameEscapingScheme` _[NameEscapingSchemeOptions](#nameescapingschemeoptions)_ | Specifies the character escaping scheme that will be requested when scraping<br />for metric and label names that do not conform to the legacy Prometheus<br />character set.<br />It requires Prometheus >= v3.4.0. |  | Enum: [AllowUTF8 Underscores Dots Values] <br /> |
| `convertClassicHistogramsToNHCB` _boolean_ | Whether to convert all scraped classic histograms into a native<br />histogram with custom buckets.<br />It requires Prometheus >= v3.4.0. |  |  |
| `minReadySeconds` _integer_ | Minimum number of seconds for which a newly created Pod should be ready<br />without any of its container crashing for it to be considered available.<br />Defaults to 0 (pod will be considered available as soon as it is ready)<br />This is an alpha field from kubernetes 1.22 until 1.24 which requires<br />enabling the StatefulSetMinReadySeconds feature gate. |  |  |
| `hostAliases` _[HostAlias](#hostalias) array_ | Optional list of hosts and IPs that will be injected into the Pod's<br />hosts file if specified. |  |  |
| `additionalArgs` _[Argument](#argument) array_ | AdditionalArgs allows setting additional arguments for the 'prometheus' container.<br />It is intended for e.g. activating hidden flags which are not supported by<br />the dedicated configuration options yet. The arguments are passed as-is to the<br />Prometheus container which may cause issues if they are invalid or not supported<br />by the given Prometheus version.<br />In case of an argument conflict (e.g. an argument which is already set by the<br />operator itself) or when providing an invalid argument, the reconciliation will<br />fail and an error will be logged. |  |  |
| `walCompression` _boolean_ | Configures compression of the write-ahead log (WAL) using Snappy.<br />WAL compression is enabled by default for Prometheus >= 2.20.0<br />Requires Prometheus v2.11.0 and above. |  |  |
| `excludedFromEnforcement` _[ObjectReference](#objectreference) array_ | List of references to PodMonitor, ServiceMonitor, Probe and PrometheusRule objects<br />to be excluded from enforcing a namespace label of origin.<br />It is only applicable if `spec.enforcedNamespaceLabel` set to true. |  |  |
| `hostNetwork` _boolean_ | Use the host's network namespace if true.<br />Make sure to understand the security implications if you want to enable<br />it (https://kubernetes.io/docs/concepts/configuration/overview/).<br />When hostNetwork is enabled, this will set the DNS policy to<br />`ClusterFirstWithHostNet` automatically (unless `.spec.DNSPolicy` is set<br />to a different value). |  |  |
| `podTargetLabels` _string array_ | PodTargetLabels are appended to the `spec.podTargetLabels` field of all<br />PodMonitor and ServiceMonitor objects. |  |  |
| `tracingConfig` _[PrometheusTracingConfig](#prometheustracingconfig)_ | TracingConfig configures tracing in Prometheus.<br />This is an *experimental feature*, it may change in any upcoming release<br />in a breaking way. |  |  |
| `bodySizeLimit` _[ByteSize](#bytesize)_ | BodySizeLimit defines per-scrape on response body size.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedBodySizeLimit. |  | Pattern: `(^0\|([0-9]*[.])?[0-9]+((K\|M\|G\|T\|E\|P)i?)?B)$` <br /> |
| `sampleLimit` _integer_ | SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedSampleLimit. |  |  |
| `targetLimit` _integer_ | TargetLimit defines a limit on the number of scraped targets that will be accepted.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedTargetLimit. |  |  |
| `labelLimit` _integer_ | Per-scrape limit on number of labels that will be accepted for a sample.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelLimit. |  |  |
| `labelNameLengthLimit` _integer_ | Per-scrape limit on length of labels name that will be accepted for a sample.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelNameLengthLimit. |  |  |
| `labelValueLengthLimit` _integer_ | Per-scrape limit on length of labels value that will be accepted for a sample.<br />Only valid in Prometheus versions 2.45.0 and newer.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedLabelValueLengthLimit. |  |  |
| `keepDroppedTargets` _integer_ | Per-scrape limit on the number of targets dropped by relabeling<br />that will be kept in memory. 0 means no limit.<br />It requires Prometheus >= v2.47.0.<br />Note that the global limit only applies to scrape objects that don't specify an explicit limit value.<br />If you want to enforce a maximum limit for all scrape objects, refer to enforcedKeepDroppedTargets. |  |  |
| `reloadStrategy` _[ReloadStrategyType](#reloadstrategytype)_ | Defines the strategy used to reload the Prometheus configuration.<br />If not specified, the configuration is reloaded using the /-/reload HTTP endpoint. |  | Enum: [HTTP ProcessSignal] <br /> |
| `maximumStartupDurationSeconds` _integer_ | Defines the maximum time that the `prometheus` container's startup probe will wait before being considered failed. The startup probe will return success after the WAL replay is complete.<br />If set, the value should be greater than 60 (seconds). Otherwise it will be equal to 600 seconds (15 minutes). |  | Minimum: 60 <br /> |
| `scrapeClasses` _[ScrapeClass](#scrapeclass) array_ | List of scrape classes to expose to scraping objects such as<br />PodMonitors, ServiceMonitors, Probes and ScrapeConfigs.<br />This is an *experimental feature*, it may change in any upcoming release<br />in a breaking way. |  |  |
| `serviceDiscoveryRole` _[ServiceDiscoveryRole](#servicediscoveryrole)_ | Defines the service discovery role used to discover targets from<br />`ServiceMonitor` objects and Alertmanager endpoints.<br />If set, the value should be either "Endpoints" or "EndpointSlice".<br />If unset, the operator assumes the "Endpoints" role. |  | Enum: [Endpoints EndpointSlice] <br /> |
| `tsdb` _[TSDBSpec](#tsdbspec)_ | Defines the runtime reloadable configuration of the timeseries database(TSDB).<br />It requires Prometheus >= v2.39.0 or PrometheusAgent >= v2.54.0. |  |  |
| `scrapeFailureLogFile` _string_ | File to which scrape failures are logged.<br />Reloading the configuration will reopen the file.<br />If the filename has an empty path, e.g. 'file.log', The Prometheus Pods<br />will mount the file into an emptyDir volume at `/var/log/prometheus`.<br />If a full path is provided, e.g. '/var/log/prometheus/file.log', you<br />must mount a volume in the specified directory and it must be writable.<br />It requires Prometheus >= v2.55.0. |  | MinLength: 1 <br /> |
| `serviceName` _string_ | The name of the service name used by the underlying StatefulSet(s) as the governing service.<br />If defined, the Service  must be created before the Prometheus/PrometheusAgent resource in the same namespace and it must define a selector that matches the pod labels.<br />If empty, the operator will create and manage a headless service named `prometheus-operated` for Prometheus resources,<br />or `prometheus-agent-operated` for PrometheusAgent resources.<br />When deploying multiple Prometheus/PrometheusAgent resources in the same namespace, it is recommended to specify a different value for each.<br />See https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#stable-network-id for more details. |  | MinLength: 1 <br /> |
| `runtime` _[RuntimeConfig](#runtimeconfig)_ | RuntimeConfig configures the values for the Prometheus process behavior |  |  |
| `terminationGracePeriodSeconds` _integer_ | Optional duration in seconds the pod needs to terminate gracefully.<br />Value must be non-negative integer. The value zero indicates stop immediately via<br />the kill signal (no opportunity to shut down) which may lead to data corruption.<br />Defaults to 600 seconds. |  | Minimum: 0 <br /> |


#### PuppetDBSDConfig



PuppetDBSDConfig configurations allow retrieving scrape targets from PuppetDB resources.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#puppetdb_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `url` _string_ | The URL of the PuppetDB root query endpoint. |  | MinLength: 1 <br />Pattern: `^http(s)?://.+$` <br /> |
| `query` _string_ | Puppet Query Language (PQL) query. Only resources are supported.<br />https://puppet.com/docs/puppetdb/latest/api/query/v4/pql.html |  | MinLength: 1 <br /> |
| `includeParameters` _boolean_ | Whether to include the parameters as meta labels.<br />Note: Enabling this exposes parameters in the Prometheus UI and API. Make sure<br />that you don't have secrets exposed as parameters if you enable this. |  |  |
| `refreshInterval` _[Duration](#duration)_ | Refresh interval to re-read the list of resources. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `port` _integer_ | Port to scrape the metrics from. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `basicAuth` _[BasicAuth](#basicauth)_ | Optional HTTP basic authentication information.<br />Cannot be set at the same time as `authorization`, or `oauth2`. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Optional `authorization` HTTP header configuration.<br />Cannot be set at the same time as `basicAuth`, or `oauth2`. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | Optional OAuth2.0 configuration.<br />Cannot be set at the same time as `basicAuth`, or `authorization`. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration to connect to the Puppet DB. |  |  |
| `followRedirects` _boolean_ | Configure whether the HTTP requests should follow HTTP 3xx redirects. |  |  |
| `enableHTTP2` _boolean_ | Configure whether to enable HTTP2. |  |  |


#### PushoverConfig



PushoverConfig configures notifications via Pushover.
See https://prometheus.io/docs/alerting/latest/configuration/#pushover_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `userKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the recipient user's user key.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator.<br />Either `userKey` or `userKeyFile` is required. |  |  |
| `userKeyFile` _string_ | The user key file that contains the recipient user's user key.<br />Either `userKey` or `userKeyFile` is required.<br />It requires Alertmanager >= v0.26.0. |  |  |
| `token` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the registered application's API token, see https://pushover.net/apps.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator.<br />Either `token` or `tokenFile` is required. |  |  |
| `tokenFile` _string_ | The token file that contains the registered application's API token, see https://pushover.net/apps.<br />Either `token` or `tokenFile` is required.<br />It requires Alertmanager >= v0.26.0. |  |  |
| `title` _string_ | Notification title. |  |  |
| `message` _string_ | Notification message. |  |  |
| `url` _string_ | A supplementary URL shown alongside the message. |  |  |
| `urlTitle` _string_ | A title for supplementary URL, otherwise just the URL is shown |  |  |
| `ttl` _[Duration](#duration)_ | The time to live definition for the alert notification |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `device` _string_ | The name of a device to send the notification to |  |  |
| `sound` _string_ | The name of one of the sounds supported by device clients to override the user's default sound choice |  |  |
| `priority` _string_ | Priority, see https://pushover.net/api#priority |  |  |
| `retry` _string_ | How often the Pushover servers will send the same notification to the user.<br />Must be at least 30 seconds. |  | Pattern: `^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$` <br /> |
| `expire` _string_ | How long your notification will continue to be retried for, unless the user<br />acknowledges the notification. |  | Pattern: `^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$` <br /> |
| `html` _boolean_ | Whether notification message is HTML or plain text. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### Receiver



Receiver defines one or more notification integrations.



_Appears in:_
- [AlertmanagerConfigSpec](#alertmanagerconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the receiver. Must be unique across all items from the list. |  | MinLength: 1 <br /> |
| `opsgenieConfigs` _[OpsGenieConfig](#opsgenieconfig) array_ | List of OpsGenie configurations. |  |  |
| `pagerdutyConfigs` _[PagerDutyConfig](#pagerdutyconfig) array_ | List of PagerDuty configurations. |  |  |
| `discordConfigs` _[DiscordConfig](#discordconfig) array_ | List of Discord configurations. |  |  |
| `slackConfigs` _[SlackConfig](#slackconfig) array_ | List of Slack configurations. |  |  |
| `webhookConfigs` _[WebhookConfig](#webhookconfig) array_ | List of webhook configurations. |  |  |
| `wechatConfigs` _[WeChatConfig](#wechatconfig) array_ | List of WeChat configurations. |  |  |
| `emailConfigs` _[EmailConfig](#emailconfig) array_ | List of Email configurations. |  |  |
| `victoropsConfigs` _[VictorOpsConfig](#victoropsconfig) array_ | List of VictorOps configurations. |  |  |
| `pushoverConfigs` _[PushoverConfig](#pushoverconfig) array_ | List of Pushover configurations. |  |  |
| `snsConfigs` _[SNSConfig](#snsconfig) array_ | List of SNS configurations |  |  |
| `telegramConfigs` _[TelegramConfig](#telegramconfig) array_ | List of Telegram configurations. |  |  |
| `webexConfigs` _[WebexConfig](#webexconfig) array_ | List of Webex configurations. |  |  |
| `msteamsConfigs` _[MSTeamsConfig](#msteamsconfig) array_ | List of MSTeams configurations.<br />It requires Alertmanager >= 0.26.0. |  |  |
| `msteamsv2Configs` _[MSTeamsV2Config](#msteamsv2config) array_ | List of MSTeamsV2 configurations.<br />It requires Alertmanager >= 0.28.0. |  |  |


#### Route



Route defines a node in the routing tree.



_Appears in:_
- [AlertmanagerConfigSpec](#alertmanagerconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `receiver` _string_ | Name of the receiver for this route. If not empty, it should be listed in<br />the `receivers` field. |  |  |
| `groupBy` _string array_ | List of labels to group by.<br />Labels must not be repeated (unique list).<br />Special label "..." (aggregate by all possible labels), if provided, must be the only element in the list. |  |  |
| `groupWait` _string_ | How long to wait before sending the initial notification.<br />Must match the regular expression`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$`<br />Example: "30s" |  |  |
| `groupInterval` _string_ | How long to wait before sending an updated notification.<br />Must match the regular expression`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$`<br />Example: "5m" |  |  |
| `repeatInterval` _string_ | How long to wait before repeating the last notification.<br />Must match the regular expression`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$`<br />Example: "4h" |  |  |
| `matchers` _[Matcher](#matcher) array_ | List of matchers that the alert's labels should match. For the first<br />level route, the operator removes any existing equality and regexp<br />matcher on the `namespace` label and adds a `namespace: <object<br />namespace>` matcher. |  |  |
| `continue` _boolean_ | Boolean indicating whether an alert should continue matching subsequent<br />sibling nodes. It will always be overridden to true for the first-level<br />route by the Prometheus operator. |  |  |
| `routes` _[JSON](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#json-v1-apiextensions-k8s-io) array_ | Child routes. |  |  |
| `muteTimeIntervals` _string array_ | Note: this comment applies to the field definition above but appears<br />below otherwise it gets included in the generated manifest.<br />CRD schema doesn't support self-referential types for now (see<br />https://github.com/kubernetes/kubernetes/issues/62872). We have to use<br />an alternative type to circumvent the limitation. The downside is that<br />the Kube API can't validate the data beyond the fact that it is a valid<br />JSON representation.<br />MuteTimeIntervals is a list of MuteTimeInterval names that will mute this route when matched, |  |  |
| `activeTimeIntervals` _string array_ | ActiveTimeIntervals is a list of MuteTimeInterval names when this route should be active. |  |  |


#### SDFile

_Underlying type:_ _string_

SDFile represents a file used for service discovery

_Validation:_
- Pattern: `^[^*]*(\*[^/]*)?\.(json|yml|yaml|JSON|YML|YAML)$`

_Appears in:_
- [FileSDConfig](#filesdconfig)



#### SNSConfig



SNSConfig configures notifications via AWS SNS.
See https://prometheus.io/docs/alerting/latest/configuration/#sns_configs



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `apiURL` _string_ | The SNS API URL i.e. https://sns.us-east-2.amazonaws.com.<br />If not specified, the SNS API URL from the SNS SDK will be used. |  |  |
| `sigv4` _[Sigv4](#sigv4)_ | Configures AWS's Signature Verification 4 signing process to sign requests. |  |  |
| `topicARN` _string_ | SNS topic ARN, i.e. arn:aws:sns:us-east-2:698519295917:My-Topic<br />If you don't specify this value, you must specify a value for the PhoneNumber or TargetARN. |  |  |
| `subject` _string_ | Subject line when the message is delivered to email endpoints. |  |  |
| `phoneNumber` _string_ | Phone number if message is delivered via SMS in E.164 format.<br />If you don't specify this value, you must specify a value for the TopicARN or TargetARN. |  |  |
| `targetARN` _string_ | The  mobile platform endpoint ARN if message is delivered via mobile notifications.<br />If you don't specify this value, you must specify a value for the topic_arn or PhoneNumber. |  |  |
| `message` _string_ | The message content of the SNS notification. |  |  |
| `attributes` _object (keys:string, values:string)_ | SNS message attributes. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### ScalewayRole

_Underlying type:_ _string_

Role of the targets to retrieve. Must be `Instance` or `Baremetal`.

_Validation:_
- Enum: [Instance Baremetal]

_Appears in:_
- [ScalewaySDConfig](#scalewaysdconfig)

| Field | Description |
| --- | --- |
| `Instance` |  |
| `Baremetal` |  |


#### ScalewaySDConfig



ScalewaySDConfig configurations allow retrieving scrape targets from Scaleway instances and baremetal services.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scaleway_sd_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `accessKey` _string_ | Access key to use. https://console.scaleway.com/project/credentials |  | MinLength: 1 <br /> |
| `secretKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Secret key to use when listing targets. |  |  |
| `projectID` _string_ | Project ID of the targets. |  | MinLength: 1 <br /> |
| `role` _[ScalewayRole](#scalewayrole)_ | Service of the targets to retrieve. Must be `Instance` or `Baremetal`. |  | Enum: [Instance Baremetal] <br /> |
| `port` _integer_ | The port to scrape metrics from. |  | Maximum: 65535 <br />Minimum: 0 <br /> |
| `apiURL` _string_ | API URL to use when doing the server listing requests. |  | Pattern: `^http(s)?://.+$` <br /> |
| `zone` _string_ | Zone is the availability zone of your targets (e.g. fr-par-1). |  | MinLength: 1 <br /> |
| `nameFilter` _string_ | NameFilter specify a name filter (works as a LIKE) to apply on the server listing request. |  | MinLength: 1 <br /> |
| `tagsFilter` _string array_ | TagsFilter specify a tag filter (a server needs to have all defined tags to be listed) to apply on the server listing request. |  | MinItems: 1 <br />items:MinLength: 1 <br /> |
| `refreshInterval` _[Duration](#duration)_ | Refresh interval to re-read the list of instances. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `followRedirects` _boolean_ | Configure whether HTTP requests follow HTTP 3xx redirects. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration to use on every scrape request |  |  |


#### ScrapeConfig



ScrapeConfig defines a namespaced Prometheus scrape_config to be aggregated across
multiple namespaces into the Prometheus configuration.



_Appears in:_
- [ScrapeConfigList](#scrapeconfiglist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[ScrapeConfigSpec](#scrapeconfigspec)_ |  |  |  |




#### ScrapeConfigSpec



ScrapeConfigSpec is a specification of the desired configuration for a scrape configuration.



_Appears in:_
- [ScrapeConfig](#scrapeconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `jobName` _string_ | The value of the `job` label assigned to the scraped metrics by default.<br />The `job_name` field in the rendered scrape configuration is always controlled by the<br />operator to prevent duplicate job names, which Prometheus does not allow. Instead the<br />`job` label is set by means of relabeling configs. |  | MinLength: 1 <br /> |
| `staticConfigs` _[StaticConfig](#staticconfig) array_ | StaticConfigs defines a list of static targets with a common label set. |  |  |
| `fileSDConfigs` _[FileSDConfig](#filesdconfig) array_ | FileSDConfigs defines a list of file service discovery configurations. |  |  |
| `httpSDConfigs` _[HTTPSDConfig](#httpsdconfig) array_ | HTTPSDConfigs defines a list of HTTP service discovery configurations. |  |  |
| `kubernetesSDConfigs` _[KubernetesSDConfig](#kubernetessdconfig) array_ | KubernetesSDConfigs defines a list of Kubernetes service discovery configurations. |  |  |
| `consulSDConfigs` _[ConsulSDConfig](#consulsdconfig) array_ | ConsulSDConfigs defines a list of Consul service discovery configurations. |  |  |
| `dnsSDConfigs` _[DNSSDConfig](#dnssdconfig) array_ | DNSSDConfigs defines a list of DNS service discovery configurations. |  |  |
| `ec2SDConfigs` _[EC2SDConfig](#ec2sdconfig) array_ | EC2SDConfigs defines a list of EC2 service discovery configurations. |  |  |
| `azureSDConfigs` _[AzureSDConfig](#azuresdconfig) array_ | AzureSDConfigs defines a list of Azure service discovery configurations. |  |  |
| `gceSDConfigs` _[GCESDConfig](#gcesdconfig) array_ | GCESDConfigs defines a list of GCE service discovery configurations. |  |  |
| `openstackSDConfigs` _[OpenStackSDConfig](#openstacksdconfig) array_ | OpenStackSDConfigs defines a list of OpenStack service discovery configurations. |  |  |
| `digitalOceanSDConfigs` _[DigitalOceanSDConfig](#digitaloceansdconfig) array_ | DigitalOceanSDConfigs defines a list of DigitalOcean service discovery configurations. |  |  |
| `kumaSDConfigs` _[KumaSDConfig](#kumasdconfig) array_ | KumaSDConfigs defines a list of Kuma service discovery configurations. |  |  |
| `eurekaSDConfigs` _[EurekaSDConfig](#eurekasdconfig) array_ | EurekaSDConfigs defines a list of Eureka service discovery configurations. |  |  |
| `dockerSDConfigs` _[DockerSDConfig](#dockersdconfig) array_ | DockerSDConfigs defines a list of Docker service discovery configurations. |  |  |
| `linodeSDConfigs` _[LinodeSDConfig](#linodesdconfig) array_ | LinodeSDConfigs defines a list of Linode service discovery configurations. |  |  |
| `hetznerSDConfigs` _[HetznerSDConfig](#hetznersdconfig) array_ | HetznerSDConfigs defines a list of Hetzner service discovery configurations. |  |  |
| `nomadSDConfigs` _[NomadSDConfig](#nomadsdconfig) array_ | NomadSDConfigs defines a list of Nomad service discovery configurations. |  |  |
| `dockerSwarmSDConfigs` _[DockerSwarmSDConfig](#dockerswarmsdconfig) array_ | DockerswarmSDConfigs defines a list of Dockerswarm service discovery configurations. |  |  |
| `puppetDBSDConfigs` _[PuppetDBSDConfig](#puppetdbsdconfig) array_ | PuppetDBSDConfigs defines a list of PuppetDB service discovery configurations. |  |  |
| `lightSailSDConfigs` _[LightSailSDConfig](#lightsailsdconfig) array_ | LightsailSDConfigs defines a list of Lightsail service discovery configurations. |  |  |
| `ovhcloudSDConfigs` _[OVHCloudSDConfig](#ovhcloudsdconfig) array_ | OVHCloudSDConfigs defines a list of OVHcloud service discovery configurations. |  |  |
| `scalewaySDConfigs` _[ScalewaySDConfig](#scalewaysdconfig) array_ | ScalewaySDConfigs defines a list of Scaleway instances and baremetal service discovery configurations. |  |  |
| `ionosSDConfigs` _[IonosSDConfig](#ionossdconfig) array_ | IonosSDConfigs defines a list of IONOS service discovery configurations. |  |  |
| `relabelings` _[RelabelConfig](#relabelconfig) array_ | RelabelConfigs defines how to rewrite the target's labels before scraping.<br />Prometheus Operator automatically adds relabelings for a few standard Kubernetes fields.<br />The original scrape job's name is available via the `__tmp_prometheus_job_name` label.<br />More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config |  | MinItems: 1 <br /> |
| `metricsPath` _string_ | MetricsPath HTTP path to scrape for metrics. If empty, Prometheus uses the default value (e.g. /metrics). |  | MinLength: 1 <br /> |
| `scrapeInterval` _[Duration](#duration)_ | ScrapeInterval is the interval between consecutive scrapes. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `scrapeTimeout` _[Duration](#duration)_ | ScrapeTimeout is the number of seconds to wait until a scrape request times out.<br />The value cannot be greater than the scrape interval otherwise the operator will reject the resource. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `scrapeProtocols` _[ScrapeProtocol](#scrapeprotocol) array_ | The protocols to negotiate during a scrape. It tells clients the<br />protocols supported by Prometheus in order of preference (from most to least preferred).<br />If unset, Prometheus uses its default value.<br />It requires Prometheus >= v2.49.0. |  | Enum: [PrometheusProto OpenMetricsText0.0.1 OpenMetricsText1.0.0 PrometheusText0.0.4 PrometheusText1.0.0] <br />MinItems: 1 <br /> |
| `fallbackScrapeProtocol` _[ScrapeProtocol](#scrapeprotocol)_ | The protocol to use if a scrape returns blank, unparseable, or otherwise invalid Content-Type.<br />It requires Prometheus >= v3.0.0. |  | Enum: [PrometheusProto OpenMetricsText0.0.1 OpenMetricsText1.0.0 PrometheusText0.0.4 PrometheusText1.0.0] <br /> |
| `honorTimestamps` _boolean_ | HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data. |  |  |
| `trackTimestampsStaleness` _boolean_ | TrackTimestampsStaleness whether Prometheus tracks staleness of<br />the metrics that have an explicit timestamp present in scraped data.<br />Has no effect if `honorTimestamps` is false.<br />It requires Prometheus >= v2.48.0. |  |  |
| `honorLabels` _boolean_ | HonorLabels chooses the metric's labels on collisions with target labels. |  |  |
| `params` _object (keys:string, values:string array)_ | Optional HTTP URL parameters |  |  |
| `scheme` _string_ | Configures the protocol scheme used for requests.<br />If empty, Prometheus uses HTTP by default. |  | Enum: [HTTP HTTPS] <br /> |
| `enableCompression` _boolean_ | When false, Prometheus will request uncompressed response from the scraped target.<br />It requires Prometheus >= v2.49.0.<br />If unset, Prometheus uses true by default. |  |  |
| `enableHTTP2` _boolean_ | Whether to enable HTTP2. |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth information to use on every scrape request. |  |  |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header to use on every scrape request. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | OAuth2 configuration to use on every scrape request. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration to use on every scrape request |  |  |
| `sampleLimit` _integer_ | SampleLimit defines per-scrape limit on number of scraped samples that will be accepted. |  |  |
| `targetLimit` _integer_ | TargetLimit defines a limit on the number of scraped targets that will be accepted. |  |  |
| `labelLimit` _integer_ | Per-scrape limit on number of labels that will be accepted for a sample.<br />Only valid in Prometheus versions 2.27.0 and newer. |  |  |
| `labelNameLengthLimit` _integer_ | Per-scrape limit on length of labels name that will be accepted for a sample.<br />Only valid in Prometheus versions 2.27.0 and newer. |  |  |
| `labelValueLengthLimit` _integer_ | Per-scrape limit on length of labels value that will be accepted for a sample.<br />Only valid in Prometheus versions 2.27.0 and newer. |  |  |
| `scrapeClassicHistograms` _boolean_ | Whether to scrape a classic histogram that is also exposed as a native histogram.<br />It requires Prometheus >= v2.45.0. |  |  |
| `nativeHistogramBucketLimit` _integer_ | If there are more than this many buckets in a native histogram,<br />buckets will be merged to stay within the limit.<br />It requires Prometheus >= v2.45.0. |  |  |
| `nativeHistogramMinBucketFactor` _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#quantity-resource-api)_ | If the growth factor of one bucket to the next is smaller than this,<br />buckets will be merged to increase the factor sufficiently.<br />It requires Prometheus >= v2.50.0. |  |  |
| `convertClassicHistogramsToNHCB` _boolean_ | Whether to convert all scraped classic histograms into a native histogram with custom buckets.<br />It requires Prometheus >= v3.0.0. |  |  |
| `keepDroppedTargets` _integer_ | Per-scrape limit on the number of targets dropped by relabeling<br />that will be kept in memory. 0 means no limit.<br />It requires Prometheus >= v2.47.0. |  |  |
| `metricRelabelings` _[RelabelConfig](#relabelconfig) array_ | MetricRelabelConfigs to apply to samples before ingestion. |  | MinItems: 1 <br /> |
| `nameValidationScheme` _[NameValidationSchemeOptions](#namevalidationschemeoptions)_ | Specifies the validation scheme for metric and label names.<br />It requires Prometheus >= v3.0.0. |  | Enum: [UTF8 Legacy] <br /> |
| `nameEscapingScheme` _[NameEscapingSchemeOptions](#nameescapingschemeoptions)_ | Metric name escaping mode to request through content negotiation.<br />It requires Prometheus >= v3.4.0. |  | Enum: [AllowUTF8 Underscores Dots Values] <br /> |
| `scrapeClass` _string_ | The scrape class to apply. |  | MinLength: 1 <br /> |


#### SlackAction



SlackAction configures a single Slack action that is sent with each
notification.
See https://api.slack.com/docs/message-attachments#action_fields and
https://api.slack.com/docs/message-buttons for more information.



_Appears in:_
- [SlackConfig](#slackconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _string_ |  |  | MinLength: 1 <br /> |
| `text` _string_ |  |  | MinLength: 1 <br /> |
| `url` _string_ |  |  |  |
| `style` _string_ |  |  |  |
| `name` _string_ |  |  |  |
| `value` _string_ |  |  |  |
| `confirm` _[SlackConfirmationField](#slackconfirmationfield)_ |  |  |  |


#### SlackConfig



SlackConfig configures notifications via Slack.
See https://prometheus.io/docs/alerting/latest/configuration/#slack_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `apiURL` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the Slack webhook URL.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `channel` _string_ | The channel or user to send notifications to. |  |  |
| `username` _string_ |  |  |  |
| `color` _string_ |  |  |  |
| `title` _string_ |  |  |  |
| `titleLink` _string_ |  |  |  |
| `pretext` _string_ |  |  |  |
| `text` _string_ |  |  |  |
| `fields` _[SlackField](#slackfield) array_ | A list of Slack fields that are sent with each notification. |  |  |
| `shortFields` _boolean_ |  |  |  |
| `footer` _string_ |  |  |  |
| `fallback` _string_ |  |  |  |
| `callbackId` _string_ |  |  |  |
| `iconEmoji` _string_ |  |  |  |
| `iconURL` _string_ |  |  |  |
| `imageURL` _string_ |  |  |  |
| `thumbURL` _string_ |  |  |  |
| `linkNames` _boolean_ |  |  |  |
| `mrkdwnIn` _string array_ |  |  |  |
| `actions` _[SlackAction](#slackaction) array_ | A list of Slack actions that are sent with each notification. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### SlackConfirmationField



SlackConfirmationField protect users from destructive actions or
particularly distinguished decisions by asking them to confirm their button
click one more time.
See https://api.slack.com/docs/interactive-message-field-guide#confirmation_fields
for more information.



_Appears in:_
- [SlackAction](#slackaction)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `text` _string_ |  |  | MinLength: 1 <br /> |
| `title` _string_ |  |  |  |
| `okText` _string_ |  |  |  |
| `dismissText` _string_ |  |  |  |


#### SlackField



SlackField configures a single Slack field that is sent with each notification.
Each field must contain a title, value, and optionally, a boolean value to indicate if the field
is short enough to be displayed next to other fields designated as short.
See https://api.slack.com/docs/message-attachments#fields for more information.



_Appears in:_
- [SlackConfig](#slackconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `title` _string_ |  |  | MinLength: 1 <br /> |
| `value` _string_ |  |  | MinLength: 1 <br /> |
| `short` _boolean_ |  |  |  |


#### StaticConfig



StaticConfig defines a Prometheus static configuration.
See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config



_Appears in:_
- [ScrapeConfigSpec](#scrapeconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `targets` _[Target](#target) array_ | List of targets for this static configuration. |  | MinItems: 1 <br /> |
| `labels` _object (keys:string, values:string)_ | Labels assigned to all metrics scraped from the targets. |  |  |


#### Target

_Underlying type:_ _string_

Target represents a target for Prometheus to scrape
kubebuilder:validation:MinLength:=1



_Appears in:_
- [StaticConfig](#staticconfig)



#### TelegramConfig



TelegramConfig configures notifications via Telegram.
See https://prometheus.io/docs/alerting/latest/configuration/#telegram_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether to notify about resolved alerts. |  |  |
| `apiURL` _string_ | The Telegram API URL i.e. https://api.telegram.org.<br />If not specified, default API URL will be used. |  |  |
| `botToken` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | Telegram bot token. It is mutually exclusive with `botTokenFile`.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator.<br />Either `botToken` or `botTokenFile` is required. |  |  |
| `botTokenFile` _string_ | File to read the Telegram bot token from. It is mutually exclusive with `botToken`.<br />Either `botToken` or `botTokenFile` is required.<br />It requires Alertmanager >= v0.26.0. |  |  |
| `chatID` _integer_ | The Telegram chat ID. |  |  |
| `messageThreadID` _integer_ | The Telegram Group Topic ID.<br />It requires Alertmanager >= 0.26.0. |  |  |
| `message` _string_ | Message template |  |  |
| `disableNotifications` _boolean_ | Disable telegram notifications |  |  |
| `parseMode` _string_ | Parse mode for telegram message |  | Enum: [MarkdownV2 Markdown HTML] <br /> |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### Time

_Underlying type:_ _string_

Time defines a time in 24hr format

_Validation:_
- Pattern: `^((([01][0-9])|(2[0-3])):[0-5][0-9])$|(^24:00$)`

_Appears in:_
- [TimeRange](#timerange)



#### TimeInterval



TimeInterval describes intervals of time



_Appears in:_
- [MuteTimeInterval](#mutetimeinterval)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `times` _[TimeRange](#timerange) array_ | Times is a list of TimeRange |  |  |
| `weekdays` _[WeekdayRange](#weekdayrange) array_ | Weekdays is a list of WeekdayRange |  | Pattern: `^((?i)sun\|mon\|tues\|wednes\|thurs\|fri\|satur)day(?:((:(sun\|mon\|tues\|wednes\|thurs\|fri\|satur)day)$)\|$)` <br /> |
| `daysOfMonth` _[DayOfMonthRange](#dayofmonthrange) array_ | DaysOfMonth is a list of DayOfMonthRange |  |  |
| `months` _[MonthRange](#monthrange) array_ | Months is a list of MonthRange |  | Pattern: `^((?i)january\|february\|march\|april\|may\|june\|july\|august\|september\|october\|november\|december\|1[0-2]\|[1-9])(?:((:((?i)january\|february\|march\|april\|may\|june\|july\|august\|september\|october\|november\|december\|1[0-2]\|[1-9]))$)\|$)` <br /> |
| `years` _[YearRange](#yearrange) array_ | Years is a list of YearRange |  | Pattern: `^2\d\{3\}(?::2\d\{3\}\|$)` <br /> |


#### TimeRange



TimeRange defines a start and end time in 24hr format



_Appears in:_
- [TimeInterval](#timeinterval)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `startTime` _[Time](#time)_ | StartTime is the start time in 24hr format. |  | Pattern: `^((([01][0-9])\|(2[0-3])):[0-5][0-9])$\|(^24:00$)` <br /> |
| `endTime` _[Time](#time)_ | EndTime is the end time in 24hr format. |  | Pattern: `^((([01][0-9])\|(2[0-3])):[0-5][0-9])$\|(^24:00$)` <br /> |


#### URL

_Underlying type:_ _string_

URL represents a valid URL

_Validation:_
- Pattern: `^https?://.+$`

_Appears in:_
- [DiscordConfig](#discordconfig)
- [WebexConfig](#webexconfig)



#### VictorOpsConfig



VictorOpsConfig configures notifications via VictorOps.
See https://prometheus.io/docs/alerting/latest/configuration/#victorops_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `apiKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the API key to use when talking to the VictorOps API.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `apiUrl` _string_ | The VictorOps API URL. |  |  |
| `routingKey` _string_ | A key used to map the alert to a team. |  |  |
| `messageType` _string_ | Describes the behavior of the alert (CRITICAL, WARNING, INFO). |  |  |
| `entityDisplayName` _string_ | Contains summary of the alerted problem. |  |  |
| `stateMessage` _string_ | Contains long explanation of the alerted problem. |  |  |
| `monitoringTool` _string_ | The monitoring tool the state message is from. |  |  |
| `customFields` _[KeyValue](#keyvalue) array_ | Additional custom fields for notification. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | The HTTP client's configuration. |  |  |


#### WeChatConfig



WeChatConfig configures notifications via WeChat.
See https://prometheus.io/docs/alerting/latest/configuration/#wechat_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `apiSecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the WeChat API key.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `apiURL` _string_ | The WeChat API URL. |  |  |
| `corpID` _string_ | The corp id for authentication. |  |  |
| `agentID` _string_ |  |  |  |
| `toUser` _string_ |  |  |  |
| `toParty` _string_ |  |  |  |
| `toTag` _string_ |  |  |  |
| `message` _string_ | API request data as defined by the WeChat API. |  |  |
| `messageType` _string_ |  |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### WebexConfig



WebexConfig configures notification via Cisco Webex
See https://prometheus.io/docs/alerting/latest/configuration/#webex_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether to notify about resolved alerts. |  |  |
| `apiURL` _[URL](#url)_ | The Webex Teams API URL i.e. https://webexapis.com/v1/messages<br />Provide if different from the default API URL. |  | Pattern: `^https?://.+$` <br /> |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | The HTTP client's configuration.<br />You must supply the bot token via the `httpConfig.authorization` field. |  |  |
| `message` _string_ | Message template |  |  |
| `roomID` _string_ | ID of the Webex Teams room where to send the messages. |  | MinLength: 1 <br /> |


#### WebhookConfig



WebhookConfig configures notifications via a generic receiver supporting the webhook payload.
See https://prometheus.io/docs/alerting/latest/configuration/#webhook_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `url` _string_ | The URL to send HTTP POST requests to. `urlSecret` takes precedence over<br />`url`. One of `urlSecret` and `url` should be defined. |  |  |
| `urlSecret` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the webhook URL to send HTTP requests to.<br />`urlSecret` takes precedence over `url`. One of `urlSecret` and `url`<br />should be defined.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |
| `maxAlerts` _integer_ | Maximum number of alerts to be sent per webhook message. When 0, all alerts are included. |  | Minimum: 0 <br /> |
| `timeout` _[Duration](#duration)_ | The maximum time to wait for a webhook request to complete, before failing the<br />request and allowing it to be retried.<br />It requires Alertmanager >= v0.28.0. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |




#### WeekdayRange

_Underlying type:_ _string_

WeekdayRange is an inclusive range of days of the week beginning on Sunday
Days can be specified by name (e.g 'Sunday') or as an inclusive range (e.g 'Monday:Friday')

_Validation:_
- Pattern: `^((?i)sun|mon|tues|wednes|thurs|fri|satur)day(?:((:(sun|mon|tues|wednes|thurs|fri|satur)day)$)|$)`

_Appears in:_
- [TimeInterval](#timeinterval)



#### YearRange

_Underlying type:_ _string_

YearRange is an inclusive range of years

_Validation:_
- Pattern: `^2\d{3}(?::2\d{3}|$)`

_Appears in:_
- [TimeInterval](#timeinterval)




## monitoring.coreos.com/v1beta1




#### AlertmanagerConfig



The `AlertmanagerConfig` custom resource definition (CRD) defines how `Alertmanager` objects process Prometheus alerts. It allows to specify alert grouping and routing, notification receivers and inhibition rules.

`Alertmanager` objects select `AlertmanagerConfig` objects using label and namespace selectors.



_Appears in:_
- [AlertmanagerConfigList](#alertmanagerconfiglist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[AlertmanagerConfigSpec](#alertmanagerconfigspec)_ |  |  |  |




#### AlertmanagerConfigSpec



AlertmanagerConfigSpec is a specification of the desired behavior of the Alertmanager configuration.
By definition, the Alertmanager configuration only applies to alerts for which
the `namespace` label is equal to the namespace of the AlertmanagerConfig resource.



_Appears in:_
- [AlertmanagerConfig](#alertmanagerconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `route` _[Route](#route)_ | The Alertmanager route definition for alerts matching the resource's<br />namespace. If present, it will be added to the generated Alertmanager<br />configuration as a first-level route. |  |  |
| `receivers` _[Receiver](#receiver) array_ | List of receivers. |  |  |
| `inhibitRules` _[InhibitRule](#inhibitrule) array_ | List of inhibition rules. The rules will only apply to alerts matching<br />the resource's namespace. |  |  |
| `timeIntervals` _[TimeInterval](#timeinterval) array_ | List of TimeInterval specifying when the routes should be muted or active. |  |  |


#### DayOfMonthRange



DayOfMonthRange is an inclusive range of days of the month beginning at 1



_Appears in:_
- [TimePeriod](#timeperiod)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `start` _integer_ | Start of the inclusive range |  | Maximum: 31 <br />Minimum: -31 <br /> |
| `end` _integer_ | End of the inclusive range |  | Maximum: 31 <br />Minimum: -31 <br /> |


#### DiscordConfig



DiscordConfig configures notifications via Discord.
See https://prometheus.io/docs/alerting/latest/configuration/#discord_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `apiURL` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | The secret's key that contains the Discord webhook URL.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `title` _string_ | The template of the message's title. |  |  |
| `message` _string_ | The template of the message's body. |  |  |
| `content` _string_ | The template of the content's body. |  | MinLength: 1 <br /> |
| `username` _string_ | The username of the message sender. |  | MinLength: 1 <br /> |
| `avatarURL` _[URL](#url)_ | The avatar url of the message sender. |  | Pattern: `^https?://.+$` <br /> |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### EmailConfig



EmailConfig configures notifications via Email.



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `to` _string_ | The email address to send notifications to. |  |  |
| `from` _string_ | The sender address. |  |  |
| `hello` _string_ | The hostname to identify to the SMTP server. |  |  |
| `smarthost` _string_ | The SMTP host and port through which emails are sent. E.g. example.com:25 |  |  |
| `authUsername` _string_ | The username to use for authentication. |  |  |
| `authPassword` _[SecretKeySelector](#secretkeyselector)_ | The secret's key that contains the password to use for authentication.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `authSecret` _[SecretKeySelector](#secretkeyselector)_ | The secret's key that contains the CRAM-MD5 secret.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `authIdentity` _string_ | The identity to use for authentication. |  |  |
| `headers` _[KeyValue](#keyvalue) array_ | Further headers email header key/value pairs. Overrides any headers<br />previously set by the notification implementation. |  |  |
| `html` _string_ | The HTML body of the email notification. |  |  |
| `text` _string_ | The text body of the email notification. |  |  |
| `requireTLS` _boolean_ | The SMTP TLS requirement.<br />Note that Go does not support unencrypted connections to remote SMTP endpoints. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration |  |  |


#### HTTPConfig



HTTPConfig defines a client HTTP configuration.
See https://prometheus.io/docs/alerting/latest/configuration/#http_config



_Appears in:_
- [DiscordConfig](#discordconfig)
- [MSTeamsConfig](#msteamsconfig)
- [MSTeamsV2Config](#msteamsv2config)
- [OpsGenieConfig](#opsgenieconfig)
- [PagerDutyConfig](#pagerdutyconfig)
- [PushoverConfig](#pushoverconfig)
- [SNSConfig](#snsconfig)
- [SlackConfig](#slackconfig)
- [TelegramConfig](#telegramconfig)
- [VictorOpsConfig](#victoropsconfig)
- [WeChatConfig](#wechatconfig)
- [WebexConfig](#webexconfig)
- [WebhookConfig](#webhookconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `authorization` _[SafeAuthorization](#safeauthorization)_ | Authorization header configuration for the client.<br />This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+. |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth for the client.<br />This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence. |  |  |
| `oauth2` _[OAuth2](#oauth2)_ | OAuth2 client credentials used to fetch a token for the targets. |  |  |
| `bearerTokenSecret` _[SecretKeySelector](#secretkeyselector)_ | The secret's key that contains the bearer token to be used by the client<br />for authentication.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `tlsConfig` _[SafeTLSConfig](#safetlsconfig)_ | TLS configuration for the client. |  |  |
| `proxyURL` _string_ | Optional proxy URL.<br />If defined, this field takes precedence over `proxyUrl`. |  |  |
| `followRedirects` _boolean_ | FollowRedirects specifies whether the client should follow HTTP 3xx redirects. |  |  |


#### InhibitRule



InhibitRule defines an inhibition rule that allows to mute alerts when other
alerts are already firing.
See https://prometheus.io/docs/alerting/latest/configuration/#inhibit_rule



_Appears in:_
- [AlertmanagerConfigSpec](#alertmanagerconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `targetMatch` _[Matcher](#matcher) array_ | Matchers that have to be fulfilled in the alerts to be muted. The<br />operator enforces that the alert matches the resource's namespace. |  |  |
| `sourceMatch` _[Matcher](#matcher) array_ | Matchers for which one or more alerts have to exist for the inhibition<br />to take effect. The operator enforces that the alert matches the<br />resource's namespace. |  |  |
| `equal` _string array_ | Labels that must have an equal value in the source and target alert for<br />the inhibition to take effect. |  |  |


#### KeyValue



KeyValue defines a (key, value) tuple.



_Appears in:_
- [EmailConfig](#emailconfig)
- [OpsGenieConfig](#opsgenieconfig)
- [PagerDutyConfig](#pagerdutyconfig)
- [VictorOpsConfig](#victoropsconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `key` _string_ | Key of the tuple. |  | MinLength: 1 <br /> |
| `value` _string_ | Value of the tuple. |  |  |


#### MSTeamsConfig



MSTeamsConfig configures notifications via Microsoft Teams.
It requires Alertmanager >= 0.26.0.



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether to notify about resolved alerts. |  |  |
| `webhookUrl` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | MSTeams webhook URL. |  | Required: \{\} <br /> |
| `title` _string_ | Message title template. |  |  |
| `summary` _string_ | Message summary template.<br />It requires Alertmanager >= 0.27.0. |  |  |
| `text` _string_ | Message body template. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### MSTeamsV2Config



MSTeamsV2Config configures notifications via Microsoft Teams using the new message format with adaptive cards as required by flows
See https://prometheus.io/docs/alerting/latest/configuration/#msteamsv2_config
It requires Alertmanager >= 0.28.0.



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether to notify about resolved alerts. |  |  |
| `webhookURL` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#secretkeyselector-v1-core)_ | MSTeams incoming webhook URL. |  |  |
| `title` _string_ | Message title template. |  | MinLength: 1 <br /> |
| `text` _string_ | Message body template. |  | MinLength: 1 <br /> |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### MatchType

_Underlying type:_ _string_

MatchType is a comparison operator on a Matcher



_Appears in:_
- [Matcher](#matcher)

| Field | Description |
| --- | --- |
| `=` |  |
| `!=` |  |
| `=~` |  |
| `!~` |  |


#### Matcher



Matcher defines how to match on alert's labels.



_Appears in:_
- [InhibitRule](#inhibitrule)
- [Route](#route)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Label to match. |  | MinLength: 1 <br /> |
| `value` _string_ | Label value to match. |  |  |
| `matchType` _[MatchType](#matchtype)_ | Match operator, one of `=` (equal to), `!=` (not equal to), `=~` (regex<br />match) or `!~` (not regex match).<br />Negative operators (`!=` and `!~`) require Alertmanager >= v0.22.0. |  | Enum: [!= = =~ !~] <br /> |




#### MonthRange

_Underlying type:_ _string_

MonthRange is an inclusive range of months of the year beginning in January
Months can be specified by name (e.g 'January') by numerical month (e.g '1') or as an inclusive range (e.g 'January:March', '1:3', '1:March')

_Validation:_
- Pattern: `^((?i)january|february|march|april|may|june|july|august|september|october|november|december|1[0-2]|[1-9])(?:((:((?i)january|february|march|april|may|june|july|august|september|october|november|december|1[0-2]|[1-9]))$)|$)`

_Appears in:_
- [TimePeriod](#timeperiod)



#### OpsGenieConfig



OpsGenieConfig configures notifications via OpsGenie.
See https://prometheus.io/docs/alerting/latest/configuration/#opsgenie_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `apiKey` _[SecretKeySelector](#secretkeyselector)_ | The secret's key that contains the OpsGenie API key.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `apiURL` _string_ | The URL to send OpsGenie API requests to. |  |  |
| `message` _string_ | Alert text limited to 130 characters. |  |  |
| `description` _string_ | Description of the incident. |  |  |
| `source` _string_ | Backlink to the sender of the notification. |  |  |
| `tags` _string_ | Comma separated list of tags attached to the notifications. |  |  |
| `note` _string_ | Additional alert note. |  |  |
| `priority` _string_ | Priority level of alert. Possible values are P1, P2, P3, P4, and P5. |  |  |
| `details` _[KeyValue](#keyvalue) array_ | A set of arbitrary key/value pairs that provide further detail about the incident. |  |  |
| `responders` _[OpsGenieConfigResponder](#opsgenieconfigresponder) array_ | List of responders responsible for notifications. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |
| `entity` _string_ | Optional field that can be used to specify which domain alert is related to. |  |  |
| `actions` _string_ | Comma separated list of actions that will be available for the alert. |  |  |


#### OpsGenieConfigResponder



OpsGenieConfigResponder defines a responder to an incident.
One of `id`, `name` or `username` has to be defined.



_Appears in:_
- [OpsGenieConfig](#opsgenieconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `id` _string_ | ID of the responder. |  |  |
| `name` _string_ | Name of the responder. |  |  |
| `username` _string_ | Username of the responder. |  |  |
| `type` _string_ | Type of responder. |  | Enum: [team teams user escalation schedule] <br />MinLength: 1 <br /> |


#### PagerDutyConfig



PagerDutyConfig configures notifications via PagerDuty.
See https://prometheus.io/docs/alerting/latest/configuration/#pagerduty_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `routingKey` _[SecretKeySelector](#secretkeyselector)_ | The secret's key that contains the PagerDuty integration key (when using<br />Events API v2). Either this field or `serviceKey` needs to be defined.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `serviceKey` _[SecretKeySelector](#secretkeyselector)_ | The secret's key that contains the PagerDuty service key (when using<br />integration type "Prometheus"). Either this field or `routingKey` needs to<br />be defined.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `url` _string_ | The URL to send requests to. |  |  |
| `client` _string_ | Client identification. |  |  |
| `clientURL` _string_ | Backlink to the sender of notification. |  |  |
| `description` _string_ | Description of the incident. |  |  |
| `severity` _string_ | Severity of the incident. |  |  |
| `class` _string_ | The class/type of the event. |  |  |
| `group` _string_ | A cluster or grouping of sources. |  |  |
| `component` _string_ | The part or component of the affected system that is broken. |  |  |
| `details` _[KeyValue](#keyvalue) array_ | Arbitrary key/value pairs that provide further detail about the incident. |  |  |
| `pagerDutyImageConfigs` _[PagerDutyImageConfig](#pagerdutyimageconfig) array_ | A list of image details to attach that provide further detail about an incident. |  |  |
| `pagerDutyLinkConfigs` _[PagerDutyLinkConfig](#pagerdutylinkconfig) array_ | A list of link details to attach that provide further detail about an incident. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |
| `source` _string_ | Unique location of the affected system. |  |  |


#### PagerDutyImageConfig



PagerDutyImageConfig attaches images to an incident



_Appears in:_
- [PagerDutyConfig](#pagerdutyconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `src` _string_ | Src of the image being attached to the incident |  |  |
| `href` _string_ | Optional URL; makes the image a clickable link. |  |  |
| `alt` _string_ | Alt is the optional alternative text for the image. |  |  |


#### PagerDutyLinkConfig



PagerDutyLinkConfig attaches text links to an incident



_Appears in:_
- [PagerDutyConfig](#pagerdutyconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `href` _string_ | Href is the URL of the link to be attached |  |  |
| `alt` _string_ | Text that describes the purpose of the link, and can be used as the link's text. |  |  |




#### PushoverConfig



PushoverConfig configures notifications via Pushover.
See https://prometheus.io/docs/alerting/latest/configuration/#pushover_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `userKey` _[SecretKeySelector](#secretkeyselector)_ | The secret's key that contains the recipient user's user key.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator.<br />Either `userKey` or `userKeyFile` is required. |  |  |
| `userKeyFile` _string_ | The user key file that contains the recipient user's user key.<br />Either `userKey` or `userKeyFile` is required.<br />It requires Alertmanager >= v0.26.0. |  |  |
| `token` _[SecretKeySelector](#secretkeyselector)_ | The secret's key that contains the registered application's API token, see https://pushover.net/apps.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator.<br />Either `token` or `tokenFile` is required. |  |  |
| `tokenFile` _string_ | The token file that contains the registered application's API token, see https://pushover.net/apps.<br />Either `token` or `tokenFile` is required.<br />It requires Alertmanager >= v0.26.0. |  |  |
| `title` _string_ | Notification title. |  |  |
| `message` _string_ | Notification message. |  |  |
| `url` _string_ | A supplementary URL shown alongside the message. |  |  |
| `urlTitle` _string_ | A title for supplementary URL, otherwise just the URL is shown |  |  |
| `ttl` _[Duration](#duration)_ | The time to live definition for the alert notification |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |
| `device` _string_ | The name of a device to send the notification to |  |  |
| `sound` _string_ | The name of one of the sounds supported by device clients to override the user's default sound choice |  |  |
| `priority` _string_ | Priority, see https://pushover.net/api#priority |  |  |
| `retry` _string_ | How often the Pushover servers will send the same notification to the user.<br />Must be at least 30 seconds. |  | Pattern: `^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$` <br /> |
| `expire` _string_ | How long your notification will continue to be retried for, unless the user<br />acknowledges the notification. |  | Pattern: `^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$` <br /> |
| `html` _boolean_ | Whether notification message is HTML or plain text. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### Receiver



Receiver defines one or more notification integrations.



_Appears in:_
- [AlertmanagerConfigSpec](#alertmanagerconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the receiver. Must be unique across all items from the list. |  | MinLength: 1 <br /> |
| `opsgenieConfigs` _[OpsGenieConfig](#opsgenieconfig) array_ | List of OpsGenie configurations. |  |  |
| `pagerdutyConfigs` _[PagerDutyConfig](#pagerdutyconfig) array_ | List of PagerDuty configurations. |  |  |
| `discordConfigs` _[DiscordConfig](#discordconfig) array_ | List of Slack configurations. |  |  |
| `slackConfigs` _[SlackConfig](#slackconfig) array_ | List of Slack configurations. |  |  |
| `webhookConfigs` _[WebhookConfig](#webhookconfig) array_ | List of webhook configurations. |  |  |
| `wechatConfigs` _[WeChatConfig](#wechatconfig) array_ | List of WeChat configurations. |  |  |
| `emailConfigs` _[EmailConfig](#emailconfig) array_ | List of Email configurations. |  |  |
| `victoropsConfigs` _[VictorOpsConfig](#victoropsconfig) array_ | List of VictorOps configurations. |  |  |
| `pushoverConfigs` _[PushoverConfig](#pushoverconfig) array_ | List of Pushover configurations. |  |  |
| `snsConfigs` _[SNSConfig](#snsconfig) array_ | List of SNS configurations |  |  |
| `telegramConfigs` _[TelegramConfig](#telegramconfig) array_ | List of Telegram configurations. |  |  |
| `webexConfigs` _[WebexConfig](#webexconfig) array_ | List of Webex configurations. |  |  |
| `msteamsConfigs` _[MSTeamsConfig](#msteamsconfig) array_ | List of MSTeams configurations.<br />It requires Alertmanager >= 0.26.0. |  |  |
| `msteamsv2Configs` _[MSTeamsV2Config](#msteamsv2config) array_ | List of MSTeamsV2 configurations.<br />It requires Alertmanager >= 0.28.0. |  |  |


#### Route



Route defines a node in the routing tree.



_Appears in:_
- [AlertmanagerConfigSpec](#alertmanagerconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `receiver` _string_ | Name of the receiver for this route. If not empty, it should be listed in<br />the `receivers` field. |  |  |
| `groupBy` _string array_ | List of labels to group by.<br />Labels must not be repeated (unique list).<br />Special label "..." (aggregate by all possible labels), if provided, must be the only element in the list. |  |  |
| `groupWait` _string_ | How long to wait before sending the initial notification.<br />Must match the regular expression`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$`<br />Example: "30s" |  |  |
| `groupInterval` _string_ | How long to wait before sending an updated notification.<br />Must match the regular expression`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$`<br />Example: "5m" |  |  |
| `repeatInterval` _string_ | How long to wait before repeating the last notification.<br />Must match the regular expression`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$`<br />Example: "4h" |  |  |
| `matchers` _[Matcher](#matcher) array_ | List of matchers that the alert's labels should match. For the first<br />level route, the operator removes any existing equality and regexp<br />matcher on the `namespace` label and adds a `namespace: <object<br />namespace>` matcher. |  |  |
| `continue` _boolean_ | Boolean indicating whether an alert should continue matching subsequent<br />sibling nodes. It will always be overridden to true for the first-level<br />route by the Prometheus operator. |  |  |
| `routes` _[JSON](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#json-v1-apiextensions-k8s-io) array_ | Child routes. |  |  |
| `muteTimeIntervals` _string array_ | Note: this comment applies to the field definition above but appears<br />below otherwise it gets included in the generated manifest.<br />CRD schema doesn't support self-referential types for now (see<br />https://github.com/kubernetes/kubernetes/issues/62872). We have to use<br />an alternative type to circumvent the limitation. The downside is that<br />the Kube API can't validate the data beyond the fact that it is a valid<br />JSON representation.<br />MuteTimeIntervals is a list of TimeInterval names that will mute this route when matched. |  |  |
| `activeTimeIntervals` _string array_ | ActiveTimeIntervals is a list of TimeInterval names when this route should be active. |  |  |


#### SNSConfig



SNSConfig configures notifications via AWS SNS.
See https://prometheus.io/docs/alerting/latest/configuration/#sns_configs



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `apiURL` _string_ | The SNS API URL i.e. https://sns.us-east-2.amazonaws.com.<br />If not specified, the SNS API URL from the SNS SDK will be used. |  |  |
| `sigv4` _[Sigv4](#sigv4)_ | Configures AWS's Signature Verification 4 signing process to sign requests. |  |  |
| `topicARN` _string_ | SNS topic ARN, i.e. arn:aws:sns:us-east-2:698519295917:My-Topic<br />If you don't specify this value, you must specify a value for the PhoneNumber or TargetARN. |  |  |
| `subject` _string_ | Subject line when the message is delivered to email endpoints. |  |  |
| `phoneNumber` _string_ | Phone number if message is delivered via SMS in E.164 format.<br />If you don't specify this value, you must specify a value for the TopicARN or TargetARN. |  |  |
| `targetARN` _string_ | The  mobile platform endpoint ARN if message is delivered via mobile notifications.<br />If you don't specify this value, you must specify a value for the topic_arn or PhoneNumber. |  |  |
| `message` _string_ | The message content of the SNS notification. |  |  |
| `attributes` _object (keys:string, values:string)_ | SNS message attributes. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### SecretKeySelector



SecretKeySelector selects a key of a Secret.



_Appears in:_
- [EmailConfig](#emailconfig)
- [HTTPConfig](#httpconfig)
- [OpsGenieConfig](#opsgenieconfig)
- [PagerDutyConfig](#pagerdutyconfig)
- [PushoverConfig](#pushoverconfig)
- [SlackConfig](#slackconfig)
- [TelegramConfig](#telegramconfig)
- [VictorOpsConfig](#victoropsconfig)
- [WeChatConfig](#wechatconfig)
- [WebhookConfig](#webhookconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | The name of the secret in the object's namespace to select from. |  | MinLength: 1 <br />Required: \{\} <br /> |
| `key` _string_ | The key of the secret to select from.  Must be a valid secret key. |  | MinLength: 1 <br />Required: \{\} <br /> |


#### SlackAction



SlackAction configures a single Slack action that is sent with each
notification.
See https://api.slack.com/docs/message-attachments#action_fields and
https://api.slack.com/docs/message-buttons for more information.



_Appears in:_
- [SlackConfig](#slackconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _string_ |  |  | MinLength: 1 <br /> |
| `text` _string_ |  |  | MinLength: 1 <br /> |
| `url` _string_ |  |  |  |
| `style` _string_ |  |  |  |
| `name` _string_ |  |  |  |
| `value` _string_ |  |  |  |
| `confirm` _[SlackConfirmationField](#slackconfirmationfield)_ |  |  |  |


#### SlackConfig



SlackConfig configures notifications via Slack.
See https://prometheus.io/docs/alerting/latest/configuration/#slack_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `apiURL` _[SecretKeySelector](#secretkeyselector)_ | The secret's key that contains the Slack webhook URL.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `channel` _string_ | The channel or user to send notifications to. |  |  |
| `username` _string_ |  |  |  |
| `color` _string_ |  |  |  |
| `title` _string_ |  |  |  |
| `titleLink` _string_ |  |  |  |
| `pretext` _string_ |  |  |  |
| `text` _string_ |  |  |  |
| `fields` _[SlackField](#slackfield) array_ | A list of Slack fields that are sent with each notification. |  |  |
| `shortFields` _boolean_ |  |  |  |
| `footer` _string_ |  |  |  |
| `fallback` _string_ |  |  |  |
| `callbackId` _string_ |  |  |  |
| `iconEmoji` _string_ |  |  |  |
| `iconURL` _string_ |  |  |  |
| `imageURL` _string_ |  |  |  |
| `thumbURL` _string_ |  |  |  |
| `linkNames` _boolean_ |  |  |  |
| `mrkdwnIn` _string array_ |  |  |  |
| `actions` _[SlackAction](#slackaction) array_ | A list of Slack actions that are sent with each notification. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### SlackConfirmationField



SlackConfirmationField protect users from destructive actions or
particularly distinguished decisions by asking them to confirm their button
click one more time.
See https://api.slack.com/docs/interactive-message-field-guide#confirmation_fields
for more information.



_Appears in:_
- [SlackAction](#slackaction)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `text` _string_ |  |  | MinLength: 1 <br /> |
| `title` _string_ |  |  |  |
| `okText` _string_ |  |  |  |
| `dismissText` _string_ |  |  |  |


#### SlackField



SlackField configures a single Slack field that is sent with each notification.
Each field must contain a title, value, and optionally, a boolean value to indicate if the field
is short enough to be displayed next to other fields designated as short.
See https://api.slack.com/docs/message-attachments#fields for more information.



_Appears in:_
- [SlackConfig](#slackconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `title` _string_ |  |  | MinLength: 1 <br /> |
| `value` _string_ |  |  | MinLength: 1 <br /> |
| `short` _boolean_ |  |  |  |


#### TelegramConfig



TelegramConfig configures notifications via Telegram.
See https://prometheus.io/docs/alerting/latest/configuration/#telegram_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether to notify about resolved alerts. |  |  |
| `apiURL` _string_ | The Telegram API URL i.e. https://api.telegram.org.<br />If not specified, default API URL will be used. |  |  |
| `botToken` _[SecretKeySelector](#secretkeyselector)_ | Telegram bot token. It is mutually exclusive with `botTokenFile`.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator.<br />Either `botToken` or `botTokenFile` is required. |  |  |
| `botTokenFile` _string_ | File to read the Telegram bot token from. It is mutually exclusive with `botToken`.<br />Either `botToken` or `botTokenFile` is required.<br />It requires Alertmanager >= v0.26.0. |  |  |
| `chatID` _integer_ | The Telegram chat ID. |  |  |
| `messageThreadID` _integer_ | The Telegram Group Topic ID.<br />It requires Alertmanager >= 0.26.0. |  |  |
| `message` _string_ | Message template |  |  |
| `disableNotifications` _boolean_ | Disable telegram notifications |  |  |
| `parseMode` _string_ | Parse mode for telegram message |  | Enum: [MarkdownV2 Markdown HTML] <br /> |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### Time

_Underlying type:_ _string_

Time defines a time in 24hr format

_Validation:_
- Pattern: `^((([01][0-9])|(2[0-3])):[0-5][0-9])$|(^24:00$)`

_Appears in:_
- [TimeRange](#timerange)



#### TimeInterval



TimeInterval specifies the periods in time when notifications will be muted or active.



_Appears in:_
- [AlertmanagerConfigSpec](#alertmanagerconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the time interval. |  | Required: \{\} <br /> |
| `timeIntervals` _[TimePeriod](#timeperiod) array_ | TimeIntervals is a list of TimePeriod. |  |  |


#### TimePeriod



TimePeriod describes periods of time.



_Appears in:_
- [TimeInterval](#timeinterval)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `times` _[TimeRange](#timerange) array_ | Times is a list of TimeRange |  |  |
| `weekdays` _[WeekdayRange](#weekdayrange) array_ | Weekdays is a list of WeekdayRange |  | Pattern: `^((?i)sun\|mon\|tues\|wednes\|thurs\|fri\|satur)day(?:((:(sun\|mon\|tues\|wednes\|thurs\|fri\|satur)day)$)\|$)` <br /> |
| `daysOfMonth` _[DayOfMonthRange](#dayofmonthrange) array_ | DaysOfMonth is a list of DayOfMonthRange |  |  |
| `months` _[MonthRange](#monthrange) array_ | Months is a list of MonthRange |  | Pattern: `^((?i)january\|february\|march\|april\|may\|june\|july\|august\|september\|october\|november\|december\|1[0-2]\|[1-9])(?:((:((?i)january\|february\|march\|april\|may\|june\|july\|august\|september\|october\|november\|december\|1[0-2]\|[1-9]))$)\|$)` <br /> |
| `years` _[YearRange](#yearrange) array_ | Years is a list of YearRange |  | Pattern: `^2\d\{3\}(?::2\d\{3\}\|$)` <br /> |


#### TimeRange



TimeRange defines a start and end time in 24hr format



_Appears in:_
- [TimePeriod](#timeperiod)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `startTime` _[Time](#time)_ | StartTime is the start time in 24hr format. |  | Pattern: `^((([01][0-9])\|(2[0-3])):[0-5][0-9])$\|(^24:00$)` <br /> |
| `endTime` _[Time](#time)_ | EndTime is the end time in 24hr format. |  | Pattern: `^((([01][0-9])\|(2[0-3])):[0-5][0-9])$\|(^24:00$)` <br /> |


#### URL

_Underlying type:_ _string_

URL represents a valid URL

_Validation:_
- Pattern: `^https?://.+$`

_Appears in:_
- [DiscordConfig](#discordconfig)
- [WebexConfig](#webexconfig)



#### VictorOpsConfig



VictorOpsConfig configures notifications via VictorOps.
See https://prometheus.io/docs/alerting/latest/configuration/#victorops_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `apiKey` _[SecretKeySelector](#secretkeyselector)_ | The secret's key that contains the API key to use when talking to the VictorOps API.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `apiUrl` _string_ | The VictorOps API URL. |  |  |
| `routingKey` _string_ | A key used to map the alert to a team. |  |  |
| `messageType` _string_ | Describes the behavior of the alert (CRITICAL, WARNING, INFO). |  |  |
| `entityDisplayName` _string_ | Contains summary of the alerted problem. |  |  |
| `stateMessage` _string_ | Contains long explanation of the alerted problem. |  |  |
| `monitoringTool` _string_ | The monitoring tool the state message is from. |  |  |
| `customFields` _[KeyValue](#keyvalue) array_ | Additional custom fields for notification. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | The HTTP client's configuration. |  |  |


#### WeChatConfig



WeChatConfig configures notifications via WeChat.
See https://prometheus.io/docs/alerting/latest/configuration/#wechat_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `apiSecret` _[SecretKeySelector](#secretkeyselector)_ | The secret's key that contains the WeChat API key.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `apiURL` _string_ | The WeChat API URL. |  |  |
| `corpID` _string_ | The corp id for authentication. |  |  |
| `agentID` _string_ |  |  |  |
| `toUser` _string_ |  |  |  |
| `toParty` _string_ |  |  |  |
| `toTag` _string_ |  |  |  |
| `message` _string_ | API request data as defined by the WeChat API. |  |  |
| `messageType` _string_ |  |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |


#### WebexConfig



WebexConfig configures notification via Cisco Webex
See https://prometheus.io/docs/alerting/latest/configuration/#webex_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether to notify about resolved alerts. |  |  |
| `apiURL` _[URL](#url)_ | The Webex Teams API URL i.e. https://webexapis.com/v1/messages |  | Pattern: `^https?://.+$` <br /> |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | The HTTP client's configuration.<br />You must use this configuration to supply the bot token as part of the HTTP `Authorization` header. |  |  |
| `message` _string_ | Message template |  |  |
| `roomID` _string_ | ID of the Webex Teams room where to send the messages. |  | MinLength: 1 <br /> |


#### WebhookConfig



WebhookConfig configures notifications via a generic receiver supporting the webhook payload.
See https://prometheus.io/docs/alerting/latest/configuration/#webhook_config



_Appears in:_
- [Receiver](#receiver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sendResolved` _boolean_ | Whether or not to notify about resolved alerts. |  |  |
| `url` _string_ | The URL to send HTTP POST requests to. `urlSecret` takes precedence over<br />`url`. One of `urlSecret` and `url` should be defined. |  |  |
| `urlSecret` _[SecretKeySelector](#secretkeyselector)_ | The secret's key that contains the webhook URL to send HTTP requests to.<br />`urlSecret` takes precedence over `url`. One of `urlSecret` and `url`<br />should be defined.<br />The secret needs to be in the same namespace as the AlertmanagerConfig<br />object and accessible by the Prometheus Operator. |  |  |
| `httpConfig` _[HTTPConfig](#httpconfig)_ | HTTP client configuration. |  |  |
| `maxAlerts` _integer_ | Maximum number of alerts to be sent per webhook message. When 0, all alerts are included. |  | Minimum: 0 <br /> |
| `timeout` _[Duration](#duration)_ | The maximum time to wait for a webhook request to complete, before failing the<br />request and allowing it to be retried.<br />It requires Alertmanager >= v0.28.0. |  | Pattern: `^(0\|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$` <br /> |




#### WeekdayRange

_Underlying type:_ _string_

WeekdayRange is an inclusive range of days of the week beginning on Sunday
Days can be specified by name (e.g 'Sunday') or as an inclusive range (e.g 'Monday:Friday')

_Validation:_
- Pattern: `^((?i)sun|mon|tues|wednes|thurs|fri|satur)day(?:((:(sun|mon|tues|wednes|thurs|fri|satur)day)$)|$)`

_Appears in:_
- [TimePeriod](#timeperiod)



#### YearRange

_Underlying type:_ _string_

YearRange is an inclusive range of years

_Validation:_
- Pattern: `^2\d{3}(?::2\d{3}|$)`

_Appears in:_
- [TimePeriod](#timeperiod)



