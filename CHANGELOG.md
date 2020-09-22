## Next release

## 0.42.1 / 2020-09-21

* [BUGFIX] Bump client-go to fix watch bug 

## 0.42.0 / 2020-09-09

The Prometheus Operator now lives in its own indepent GitHub organization.  
We have also added a governance (#3398).

* [FEATURE] Move API types out into their own module (#3395)
* [FEATURE] Create a monitoring mixin for prometheus-operator (#3333)
* [ENHANCEMENT] Remove multilistwatcher and denylistfilter (#3440)
* [ENHANCEMENT] Instrument client-go requests (#3465)
* [ENHANCEMENT] pkg/prometheus: skip invalid service monitors (#3445)
* [ENHANCEMENT] pkg/alertmanager: Use lower value for --cluster.reconnect-timeout (#3436)
* [ENHANCEMENT] pkg/alertmanager: cleanup resources via OwnerReferences (#3423)
* [ENHANCEMENT] Add prometheus_operator_reconcile_operations_total metric (#3415)
* [ENHANCEMENT] pkg/operator/image.go: Adjust image path building (#3392)
* [ENHANCEMENT] Specify timeouts per Alertmanager target when sending alerts. (#3385)
* [ENHANCEMENT] Push container images to Quay into coreos and prometheus-operator orgs (#3390)
* [ENHANCEMENT] Run single replica Alertmanager in HA cluster mode (#3382)
* [BUGFIX] Fix validation logic for SecretOrConfigMap (#3413)
* [BUGFIX] Don't overwrite __param_target (#3377)

## 0.41.1 / 2020-08-12

* [BUGFIX] Fix image url logic (#3402)

## 0.41.0 / 2020-07-29

* [CHANGE] Configmap-reload: Update to v0.4.0 (#3334)
* [CHANGE] Update prometheus compatibility matrix to v2.19.2 (#3316)
* [FEATURE] Add Synthetic Probes support. This includes support for job names. (#2832, #3318, #3312, #3306)
* [FEATURE] Support Prometheus vertical compaction (#3281)
* [ENHANCEMENT] pkg: Instrument resources being tracked by the operator (#3360)
* [ENHANCEMENT] Add SecretListWatchSelector to reduce memory and CPU footprint (#3355)
* [ENHANCEMENT] Added support for configuring CA, cert, and key via secret or configmap. (#3249)
* [ENHANCEMENT] Consolidate image url logic, deprecating `baseImage`, `sha`, and `tag` in favor of `image` field in CRDs. (#3103, #3358)
* [ENHANCEMENT] Decouple alertmanager pod labels from selector labels (#3317)
* [ENHANCEMENT] pkg/prometheus: Ensure relabeling of container label in ServiceMonitors (#3315)
* [ENHANCEMENT] misc: Remove v1beta1 crd remainings (#3311)
* [ENHANCEMENT] Normalize default durations (#3308)
* [ENHANCEMENT] pkg/prometheus: Allow enforcing namespace label in Probe configs (#3304)
* [BUGFIX] Revert "Normalize default durations" (#3364)
* [BUGFIX] Reload alertmanager on configmap/secret change (#3319)
* [BUGFIX] listwatch: Do not duplicate resource versions (#3373)

## 0.40.0 / 2020-06-17

* [CHANGE] Update dependencies to prometheus 2.18 (#3231)
* [CHANGE] Add support for new prometheus versions (v2.18 & v2.19) (#3284)
* [CHANGE] bump Alertmanager default version to v0.21.0 (#3286) 
* [FEATURE] Automatically disable high availability mode for 1 replica alertmanager (#3233)
* [FEATURE] thanos-sidecar: Add minTime arg (#3253)
* [FEATURE] Add scrapeTimeout as global configurable parameter (#3250) 
* [FEATURE] Add EnforcedSampleLimit which enforces a global sample limit (#3276) 
* [FEATURE] add ability to exclude rules from namespace label enforcement (#3207) 
* [BUGFIX] thanos sidecar: log flags double definition (#3242)
* [BUGFIX] Mutate rule labels, annotations to strings (#3230)

## 0.39.0 / 2020-05-06

* [CHANGE] Introduce release schedule (#3135)
* [CHANGE] Remove options configuring CRD management (manage-crds, crd-kinds, with-validation) (#3155)
* [CHANGE] Add CRD definitions to bundle.yaml (#3171)
* [CHANGE] Switch to apiextensions.k8s.io/v1 CRD and require kubernetes v1.16 or newer (#3175, #3187)
* [FEATURE] Add support prometheus query log file (#3116)
* [FEATURE] Add support for watching specified rules directory by config-relader (#3128)
* [FEATURE] Add TLS support for operator web server (#3134, #3157)
* [FEATURE] Allow to set address for operator http endpoint (#3098)
* [FEATURE] Allow setting the alertmanagers cluster.advertiseAddress (#3160)
* [FEATURE] Build operator images for ARM and ARM64 platforms (#3177)
* [ENHANCEMENT] Allow setting log level and format for thanos sidecar (#3112)
* [ENHANCEMENT] Support naming of remote write queues (#3144)
* [ENHANCEMENT] Allow disabling mount subPath for volumes (#3143)
* [ENHANCEMENT] Update k8s libraries to v1.18 (#3154)
* [ENHANCEMENT] Create separate namespace informers when needed (#3182)
* [BUGFIX] Tolerate version strings which aren't following semver (#3101)
* [BUGFIX] Retain metadata for embedded PVC objects (#3115)
* [BUGFIX] Fix definition of thanos-ruler-operated service (#3126)
* [BUGFIX] Allow setting the cluster domain (#3138)
* [BUGFIX] Allow matching only PodMonitors (#3173)
* [BUGFIX] Fix typo in statefulset informer (#3179)

## 0.38.1 / 2020-04-16

* [BUGFIX] Fix definition of web service port for Alertmanager (#3125)
* [BUGFIX] Support external alert query URL for THanos Ruler (#3129)
* [BUGFIX] Do not modify the PrometheusRule cache object (#3105)

## 0.38.0 / 2020-03-20

* [CHANGE] Changed ThanosRuler custom resource field alertmanagersURL type from string to []string (#3067)
* [CHANGE] Deprecate PodMonitor targetPort field (#3071, #3078)
* [FEATURE] Add queryConfig field to ThanosRuler spec (#3068)
* [FEATURE] GRPC TLS config for Thanos Ruler and Sidecar (#3059)
* [FEATURE] MergePatch Alertmanager containers (#3080)
* [FEATURE] Add VolumeMounts to Prometheus custom resources (#2871)
* [ENHANCEMENT] Update Thanos to v0.11.0 (#3066)
* [ENHANCEMENT] Clarify that Endpoint.targetPort is pod port (#3064)
* [BUGFIX] ThanosRuler restarts instead of reloading with new PrometheusRules (#3056)
* [BUGFIX] Omit QueryEndpoints if empty (#3091)

## 0.37.0 / 2020-03-02

* [FEATURE] Add routePrefix to ThanosRuler spec (#3023)
* [FEATURE] Add externalPrefix to ThanosRuler spec (#3058)
* [FEATURE] Add pod template fields to ThanosRuler custom resource (#3034)
* [ENHANCEMENT] Make ports on kubelet service and endpoints match (#3039)
* [ENHANCEMENT] Update kubernetes API dependencies to v0.17.3/1.17.3 (#3042)
* [ENHANCEMENT] Simplify multi-arch building (#3035)
* [ENHANCEMENT] Default to Prometheus v2.16.0 (#3050, #3051)
* [BUGFIX] Fix stateful set being pruned by kubectl (#3029, #3030)
* [BUGFIX] Fix prometheus rule validator admitting rules with invalid label types (#2727,#2962)
* [BUGFIX] Fix flaky test in Thanos ruler (#3038)
* [BUGFIX] Fix ThanosRuler status reporting (#3045)
* [BUGFIX] Preserve pod labels and annotations in custom resources (#3041, #3043)
* [BUGFIX] Prevent stateful set update loop for alertmanager and thonos types (#3048, #3049)

## 0.36.0 / 2020-02-10

* [CHANGE] Rename binary `lint` to `po-lint` (#2964)
* [CHANGE] Restrict api extension RBAC rules (#2974)
* [FEATURE] Add operator for Thanos Ruler resources (#2943)
* [FEATURE] Thanos Ruler Improvements (#2986, #2991, #2993, #2994, #3018, #3019)
* [FEATURE] Add additional printer columns for custom resources (#2922)
* [ENHANCEMENT] Set config-reloader containers resources (#2958)
* [ENHANCEMENT] Fix broken links and remove spec.version in examples (#2961)
* [ENHANCEMENT] Reduce deprecation warning verbosity (#2978)
* [ENHANCEMENT] Build tooling improvements (#2979, #2982, #2983)
* [ENHANCEMENT] Update Prometheus compatible version list (#2998)
* [ENHANCEMENT] Fix broken links in documentation (#3005)
* [ENHANCEMENT] Update default container image versions (#3007)
* [BUGFIX] Fix statefulset crash loop in Kube 1.17 (#2987)

## 0.35.0 / 2020-01-13

* [CHANGE] Deprecate baseImage, tag, and sha fields in custom resources (#2914)
* [FEATURE] Add APIVersion field to Prometheus.Spec.Alerting (#2884)
* [FEATURE] Add an option to disable compaction (#2886)
* [FEATURE] Allow configuring PVC access mode (#978)
* [ENHANCEMENT] Do not disable compaction on sidecar without object storage configuration (#2845)
* [ENHANCEMENT] Fix StatefulSet being needlessly re-created (#2857)
* [ENHANCEMENT] Add metric for statefulset re-create (#2859)
* [ENHANCEMENT] Rename `mesh` ports to be prefixed with protocol (#2863)
* [ENHANCEMENT] Use kubebuilder controller-gen for creating CRDs (#2855)
* [ENHANCEMENT] Add metrics about node endpoints synchronization (#2887)
* [ENHANCEMENT] Turn off preserveUnknownFields to enable kubectl explain (#2903)
* [ENHANCEMENT] Instrument operator's list and watch operations (#2893)
* [BUGFIX] Modified prometheus wget probe when listenLocal=true (#2929)
* [BUGFIX] Fix generated statefulset being pruned by kubectl (#2944)

## 0.34.0 / 2019-10-31

* [CHANGE] Make arbitraryFSAccessThroughSMs optional (#2797)
* [FEATURE] Add [prometheus,alertmanager]-instance-namespaces cmdline parameter (#2783)
* [FEATURE] Add configSecret element to the AlertmanagerSpec (#2827)
* [FEATURE] Add enforcedNamespaceLabel to Prometheus CRD (#2820)
* [FEATURE] Add exec probes against localhost:9090 to Prometheus if listenLocal is set to true (#2763)
* [FEATURE] Add honorTimestamps field to Prometheus, Podmonitor, and ServiceMonitor CRD (#2800)
* [FEATURE] Add ignoreNamespaceSelectors field to Prometheus CRD (#2816)
* [FEATURE] Add local configuration options in jsonnet/prometheus-operator (#2794)
* [FEATURE] Add overrideHonorLabels to Prometheus CRD (#2806)
* [FEATURE] Reference secrets instead of local files (#2716)
* [ENHANCEMENT] Add missing json struct tag for ArbitraryFSAccessThroughSMsConfig (#2808)
* [ENHANCEMENT] Ensure containers have "FallbackToLogsOnError" termination policy (#2819)
* [ENHANCEMENT] Improve detection of StatefulSet changes (#2801)
* [ENHANCEMENT] Only append relabelings if EnforcedNamespaceLabel value is set (#2830)
* [ENHANCEMENT] Remove unneeded Ownership change in prometheus-config-reloader (#2761)
* [ENHANCEMENT] Update k8s client to 1.16 (#2778)
* [BUGFIX] Update prometheus dependency to v2.12.0 to fix validation failure for .externalLabels admission webhook (#2779)

## 0.33.0 / 2019-09-12

* [FEATURE] Add Thanos service port to governing service (#2754)
* [FEATURE] Add VolumeMounts to Alertmanager (#2755)
* [ENHANCEMENT] Bump default thanos image and version (#2746)

## 0.32.0 / 2019-08-30

* [CHANGE] Change PodManagement policy to parallel in Alertmanager (#2676)
* [FEATURE] Adding label selector for Alertmanager objects discovery filtering (#2662)
* [FEATURE] Provide option to turn on WAL compression (#2683)
* [FEATURE] Support namespace denylist for listwatcher (#2710)
* [FEATURE] Add support for Volumes to Prometheus Custom Resource (#2734)
* [FEATURE] Add support for Volumes to Alertmanager Custom Resource (#2737)
* [FEATURE] Add support for InitContainers to Prometheus Custom Resource (#2522)

## 0.31.1 / 2019-06-25
* [BUGFIX] Increase terminationGracePeriod for alertmanager statefulSet as it cannot be 0. (#2657)

## 0.31.0 / 2019-06-20

* [CHANGE] Remove gossip configuration from Thanos sidecar. This means only non-gossip configurations can be used going forward. (#2623, #2629)
* [FEATURE] Add PodMonitor, allowing monitoring pods directly without the necessity to go through a Endpoints of a Service, this is an experimental feature, it may break at any time without notice. (#2566)
* [FEATURE] Add admission webhook to validate `PrometheusRule` objects with Prometheus' promtool linting. (#2551)
* [FEATURE] Add ability to select subset of Prometheus objects to reconcile against, configurable with `--prometheus-instance-selector` flag. (#2615)
* [FEATURE] Add ability to configure size based retention on Prometheus. (#2608)
* [FEATURE] Add ability to use StatefulSet ordinal in external labels. (#2591)
* [ENHANCEMENT] Use /-/healthy and /-/ready for probes in Alertmanager. (#2600)

## 0.30.0 / 2019-05-10

Note: Both kube-prometheus (#2554) and the Helm Chart (#2416) have been removed from this repository.
kube-prometheus is not hosted as github.com/coroes/kube-prometheus and the helm chart is available at https://github.com/helm/charts/tree/master/stable/prometheus-operator

* [CHANGE] Drop support for Alertmanager < v0.15.0 (#2568)
* [FEATURE] Add Prometheus Config Reloader CPU and Memory flags (#2466)
* [FEATURE] Support `--max-samples` flag in QuerySpec (#2505)
* [FEATURE] Adding kustomization files for remote bases (#2497)
* [FEATURE] Allow disabling limits on sidecars (#2560)
* [FEATURE] Modify arbitrary parts of the operator generated containers (#2445)
* [ENHANCEMENT] Add proper Operator labels as recommended by SIG-Apps (#2427)
* [ENHANCEMENT] Watch ConfigMaps having the prometheus-name selector (#2454)
* [ENHANCEMENT] Add prometheusExternalLabelName field to Prometheus object (#2430)
* [ENHANCEMENT] Optional secret in scrapeconfig (#2511)
* [ENHANCEMENT] Update PodSecurityContext docs (#2569)
* [ENHANCEMENT] Update Kubernetes client libraries to 1.14.0 (#2570)
* [ENHANCEMENT] Use Go modules with Kubernetes 1.14 (#2571)
* [ENHANCEMENT] Update to Alertmanager v0.17.0 (#2587)
* [ENHANCEMENT] Add support for setting Log Format for Alertmanager (#2577)
* [ENHANCEMENT] Switch Deployments and StatefulSets from apps/v1beta to apps/v1 (#2593)
* [ENHANCEMENT] Add Service and Servicemonitor to bundle.yaml (#2595)
* [BUGFIX] Fix startup nodeSyncEndpoints (#2475)
* [BUGIFX] Update Thanos vendoring to include config reloader fixes (#2504)

## 0.29.0 / 2019-02-19

* [FEATURE] Thanos sidecar supports external Thanos clusters (#2412)
* [FEATURE] Make replicas external label name configurable (#2411)
* [FEATURE] Flags for config reloader memory and cpu limits (#2403)
* [ENHANCEMENT] Update to Prometheus v2.7.1 as default (#2374)
* [ENHANCEMENT] Update to Alertmanager v0.16.1 as default (#2362)

## 0.28.0 / 2019-01-24

* [FEATURE] CLI tool to lint YAML against CRD definitions (#2269)
* [FEATURE] Support Thanos v0.2 arbitrary object storage configuration (#2264)
* [ENHANCEMENT] Update Alertmanager to v0.16.0 (#2145)
* [ENHANCEMENT] Added AlertResendDelay to Prometheus resource (#2265)
* [ENHANCEMENT] Support min_shards configuration of the queueConfig (#2284)
* [ENHANCEMENT] Write compressed Prometheus config into Kubernetes Secret (#2243)
* [ENHANCEMENT] Add flag to enable Prometheus web admin API (#2300)
* [ENHANCEMENT] Add logFormat support for Prometheus (#2307)
* [ENHANCEMENT] Configure Thanos sidecar with route prefix (#2345)
* [BUGFIX] Fix omitting source_labels where they are unnecessary (#2292)
* [BUGFIX] Guard against nil targetPort (#2318)

## 0.27.0 / 2019-01-08

* [FEATURE] Add `image` field to specify full Prometheus, Alertmanager and Thanos images.
* [FEATURE] Add prometheus query options (lookback-delta, max-concurrency, timeout).

## 0.26.0 / 2018-11-30

* [CHANGE] Remove attempting to set "secure" security context (#2109).
* [CHANGE] Remove deprecated StorageSpec fields (#2132).
* [ENHANCEMENT] Better handling for pod/node labels from ServiceMonitors (#2089).
* [ENHANCEMENT] Update to Prometheus v2.5.0 as default (#2101).
* [ENHANCEMENT] Update to Alertmanager v0.15.3 as default (#2128).
* [ENHANCEMENT] Increase CPU limits for small containers to not being throttled as much (#2144).
* [BUGFIX] Sanitize thanos secret volume mount name (#2159).
* [BUGFIX] Fix racy Kubernetes multi watch (#2177).

## 0.25.0 / 2018-10-24

* [FEATURE] Allow passing additional alert relabel configs in Prometheus custom resource (#2022)
* [FEATURE] Add ability to mount custom ConfigMaps into Alertmanager and Prometheus (#2028)

## 0.24.0 / 2018-10-11

This release has a breaking changes for `prometheus_operator_.*` metrics.

`prometheus_operator_alertmanager_reconcile_errors_total` and `prometheus_operator_prometheus_reconcile_errors_total`
are now combined and called `prometheus_operator_reconcile_errors_total`.
Instead the metric has a "controller" label which indicates the errors from the Prometheus or Alertmanager controller.

The same happened with `prometheus_operator_alertmanager_spec_replicas` and `prometheus_operator_prometheus_spec_replicas`
which is now called `prometheus_operator_spec_replicas` and also has the "controller" label.

The `prometheus_operator_triggered_total` metric now has a "controller" label as well and finally instruments the
Alertmanager controller.

For a full description see: https://github.com/coreos/prometheus-operator/pull/1984#issue-221139702

In order to support multiple namespaces, the `--namespace` flag changed to `--namespaces`
and accepts and comma-separated list of namespaces as a string.

* [CHANGE] Default to Node Exporter v0.16.0 (#1812)
* [CHANGE] Update to Go 1.11 (#1855)
* [CHANGE] Default to Prometheus v2.4.3 (#1929) (#1983)
* [CHANGE] Default to Thanos v0.1.0 (#1954)
* [CHANGE] Overhaul metrics while adding triggerBy metric for Alertmanager (#1984)
* [CHANGE] Add multi namespace support (#1813)
* [FEATURE] Add SHA field to Prometheus, Alertmanager and Thanos for images (#1847) (#1854)
* [FEATURE] Add configuration for priority class to be assigned to Pods (#1875)
* [FEATURE] Configure sampleLimit per ServiceMonitor (#1895)
* [FEATURE] Add additionalPeers to Alertmanager (#1878)
* [FEATURE] Add podTargetLabels to ServiceMonitors (#1880)
* [FEATURE] Relabel target name for Pods (#1896)
* [FEATURE] Allow configuration of relabel_configs per ServiceMonitor (#1879)
* [FEATURE] Add illegal update reconciliation by deleting StatefulSet (#1931)
* [ENHANCEMENT] Set Thanos cluster and grpc ip from pod.ip (#1836)
* [BUGFIX] Add square brackets around pod IPs for IPv6 support (#1881)
* [BUGFIX] Allow periods in secret name (#1907)
* [BUGFIX] Add BearerToken in generateRemoteReadConfig (#1956)

## 0.23.2 / 2018-08-23

* [BUGFIX] Do not abort kubelet endpoints update due to nodes without IP addresses defined (#1816)

## 0.23.1 / 2018-08-13

* [BUGFIX] Fix high CPU usage of Prometheus Operator when annotating Prometheus resource (#1785)

## 0.23.0 / 2018-08-06

* [CHANGE] Deprecate specification of Prometheus rules via ConfigMaps in favor of `PrometheusRule` CRDs
* [FEATURE] Introduce new flag to control logging format (#1475)
* [FEATURE] Ensure Prometheus Operator container runs as `nobody` user by default (#1393)
* [BUGFIX] Fix reconciliation of Prometheus StatefulSets due to ServiceMonitors and PrometheusRules changes when a single namespace is being watched (#1749)

## 0.22.2 / 2018-07-24

[BUGFIX] Do not migrate rule config map for Prometheus statefulset on rule config map to PrometheusRule migration (#1679)

## 0.22.1 / 2018-07-19

* [ENHANCEMENT] Enable operation when CRDs are created externally (#1640)
* [BUGFIX] Do not watch for new namespaces if a specific namespace has been selected (#1640)

## 0.22.0 / 2018-07-09

* [FEATURE] Allow setting volume name via volumetemplateclaimtemplate in prom and alertmanager (#1538)
* [FEATURE] Allow setting custom tags of container images (#1584) 
* [ENHANCEMENT] Update default Thanos to v0.1.0-rc.2 (#1585)
* [ENHANCEMENT] Split rule config map mounted into Prometheus if it exceeds Kubernetes config map limit (#1562)
* [BUGFIX] Mount Prometheus data volume into Thanos sidecar & pass correct path to Thanos sidecar (#1583)

## 0.21.0 / 2018-06-28

* [CHANGE] Default to Prometheus v2.3.1.
* [CHANGE] Default to Alertmanager v0.15.0.
* [FEATURE] Make remote write queue configurations configurable.
* [FEATURE] Add Thanos integration (experimental).
* [BUGFIX] Fix usage of console templates and libraries.

## 0.20.0 / 2018-06-05

With this release we introduce a new Custom Resource Definition - the
`PrometheusRule` CRD. It addresses the need for rule syntax validation and rule
selection across namespaces. `PrometheusRule` replaces the configuration of
Prometheus rules via K8s ConfigMaps. There are two migration paths:

1. Automated live migration: If the Prometheus Operator finds Kubernetes
   ConfigMaps that match the `RuleSelector` in a `Prometheus` specification, it
   will convert them to matching `PrometheusRule` resources.

2. Manual migration: We provide a basic CLI tool to convert Kubernetes
   ConfigMaps to `PrometheusRule` resources.

```bash
go get -u github.com/coreos/prometheus-operator/cmd/po-rule-migration
po-rule-migration \
--rule-config-map=<path-to-config-map> \
--rule-crds-destination=<path-to-rule-crd-destination>
```

* [FEATURE] Add leveled logging to Prometheus Operator (#1277)
* [FEATURE] Allow additional Alertmanager configuration in Prometheus CRD (#1338)
* [FEATURE] Introduce `PrometheusRule` Custom Resource Definition (#1333)
* [ENHANCEMENT] Allow Prometheus to consider all namespaces to find ServiceMonitors (#1278)
* [BUGFIX] Do not attempt to set default memory request for Prometheus 2.0 (#1275)

## 0.19.0 / 2018-04-25

* [FEATURE] Allow specifying additional Prometheus scrape configs via secret (#1246)
* [FEATURE] Enable Thanos sidecar (#1219)
* [FEATURE] Make AM log level configurable (#1192)
* [ENHANCEMENT] Enable Prometheus to select Service Monitors outside own namespace (#1227)
* [ENHANCEMENT] Enrich Prometheus operator CRD registration error handling (#1208)
* [BUGFIX] Allow up to 10m for Prometheus pod on startup for data recovery (#1232)

## 0.18.1 / 2018-04-09

* [BUGFIX] Fix alertmanager >=0.15.0 cluster gossip communication (#1193)

## 0.18.0 / 2018-03-04

From this release onwards only Kubernetes versions v1.8 and higher are supported. If you have an older version of Kubernetes and the Prometheus Operator running, we recommend upgrading Kubernetes first and then the Prometheus Operator.

While multiple validation issues have been fixed, it will remain a beta feature in this release. If you want to update validations, you need to either apply the CustomResourceDefinitions located in `example/prometheus-operator-crd` or delete all CRDs and restart the Prometheus Operator.

Some changes cause Prometheus and Alertmanager clusters to be redeployed. If you do not have persistent storage backing your data, this means you will loose the amount of data equal to your retention time.

* [CHANGE] Use canonical `/prometheus` and `/alertmanager` as data dirs in containers.
* [FEATURE] Allow configuring Prometheus and Alertmanager servers to listen on loopback interface, allowing proxies to be the ingress point of those Pods.
* [FEATURE] Allow configuring additional containers in Prometheus and Alertmanager Pods.
* [FEATURE] Add ability to whitelist Kubernetes labels to become Prometheus labels.
* [FEATURE] Allow specifying additional secrets for Alertmanager Pods to mount.
* [FEATURE] Allow specifying `bearer_token_file` for Alertmanger configurations of Prometheus objects in order to authenticate with Alertmanager.
* [FEATURE] Allow specifying TLS configuration for Alertmanger configurations of Prometheus objects.
* [FEATURE] Add metrics for reconciliation errors: `prometheus_operator_alertmanager_reconcile_errors_total` and `prometheus_operator_prometheus_reconcile_errors_total`.
* [FEATURE] Support `read_recent` and `required_matchers` fields for remote read configurations.
* [FEATURE] Allow disabling any defaults of `SecurityContext` fields of Pods.
* [BUGFIX] Handle Alertmanager >=v0.15.0 breaking changes correctly.
* [BUGFIX] Fix invalid validations for metric relabeling fields.
* [BUGFIX] Fix validations for `AlertingSpec`.
* [BUGFIX] Fix validations for deprecated storage fields.
* [BUGFIX] Fix remote read and write basic auth support.
* [BUGFIX] Fix properly propagate errors of Prometheus config reloader.

## 0.17.0 / 2018-02-15

This release adds validations as a beta feature. It will only be installed on new clusters, existing CRD definitions will not be updated, this will be done in a future release. Please try out this feature and give us feedback!

* [CHANGE] Default Prometheus version v2.2.0-rc.0.
* [CHANGE] Default Alertmanager version v0.14.0.
* [FEATURE] Generate and add CRD validations.
* [FEATURE] Add ability to set `serviceAccountName` for Alertmanager Pods.
* [FEATURE] Add ability to specify custom `securityContext` for Alertmanager Pods.
* [ENHANCEMENT] Default to non-root security context for Alertmanager Pods.

## 0.16.1 / 2018-01-16

* [CHANGE] Default to Alertmanager v0.13.0.
* [BUGFIX] Alertmanager flags must be double dashed starting v0.13.0.

## 0.16.0 / 2018-01-11

* [FEATURE] Add support for specifying remote storage configurations.
* [FEATURE] Add ability to specify log level.
* [FEATURE] Add support for dropping metrics at scrape time.
* [ENHANCEMENT] Ensure that resource limit can't make Pods unschedulable.
* [ENHANCEMENT] Allow configuring emptyDir volumes
* [BUGFIX] Use `--storage.tsdb.no-lockfile` for Prometheus 2.0.
* [BUGFIX] Fix Alertmanager default storage.path.

## 0.15.0 / 2017-11-22

* [CHANGE] Default Prometheus version v2.0.0
* [BUGFIX] Generate ExternalLabels deterministically
* [BUGFIX] Fix incorrect mount path of Alertmanager data volume
* [EXPERIMENTAL] Add ability to specify CRD Kind name

## 0.14.1 / 2017-11-01

* [BUGFIX] Ignore illegal change of PodManagementPolicy to StatefulSet.

## 0.14.0 / 2017-10-19

* [CHANGE] Default Prometheus version v2.0.0-rc.1.
* [CHANGE] Default Alertmanager version v0.9.1.
* [BUGFIX] Set StatefulSet replicas to 0 if 0 is specified in Alertmanager/Prometheus object.
* [BUGFIX] Glob for all files in a ConfigMap as rule files.
* [FEATURE] Add ability to run Prometheus Operator for a single namespace.
* [FEATURE] Add ability to specify CRD api group.
* [FEATURE] Use readiness and health endpoints of Prometheus 1.8+.
* [ENHANCEMENT] Add OwnerReferences to managed objects.
* [ENHANCEMENT] Use parallel pod creation strategy for Prometheus StatefulSets.

## 0.13.0 / 2017-09-21

After a long period of not having broken any functionality in the Prometheus Operator, we have decided to promote the status of this project to beta.

Compatibility guarantees and migration strategies continue to be the same as for the `v0.12.0` release.

* [CHANGE] Remove analytics collection.
* [BUGFIX] Fix memory leak in kubelet endpoints sync.
* [FEATURE] Allow setting global default `scrape_interval`.
* [FEATURE] Allow setting Pod objectmeta to Prometheus and Alertmanger objects.
* [FEATURE] Allow setting tolerations and affinity for Prometheus and Alertmanager objects.

## 0.12.0 / 2017-08-24

Starting with this release only Kubernetes `v1.7.x` and up is supported as CustomResourceDefinitions are a requirement for the Prometheus Operator and are only available from those versions and up.

Additionally all objects have been promoted from `v1alpha1` to `v1`. On start up of this version of the Prometheus Operator the previously used `ThirdPartyResource`s and the associated `v1alpha1` objects will be automatically migrated to their `v1` equivalent `CustomResourceDefinition`.

* [CHANGE] All manifests created and used by the Prometheus Operator have been promoted from `v1alpha1` to `v1`.
* [CHANGE] Use Kubernetes `CustomResourceDefinition`s instead of `ThirdPartyResource`s.
* [FEATURE] Add ability to set scrape timeout to `ServiceMonitor`.
* [ENHANCEMENT] Use `StatefulSet` rolling deployments.
* [ENHANCEMENT] Properly set `SecurityContext` for Prometheus 2.0 deployments.
* [ENHANCEMENT] Enable web lifecycle APIs for Prometheus 2.0 deployments.

## 0.11.2 / 2017-09-21

* [BUGFIX] Fix memory leak in kubelet endpoints sync.

## 0.11.1 / 2017-07-28

* [ENHANCEMENT] Add profiling endpoints.
* [BUGFIX] Adapt Alertmanager storage usage to not use deprecated storage definition.

## 0.11.0 / 2017-07-20

Warning: This release deprecates the previously used storage definition in favor of upstream PersistentVolumeClaim templates. While this should not have an immediate effect on a running cluster, Prometheus object definitions that have storage configured need to be adapted. The previously existing fields are still there, but have no effect anymore.

* [FEATURE] Add Prometheus 2.0 alpha3 support.
* [FEATURE] Use PVC templates instead of custom storage definition.
* [FEATURE] Add cAdvisor port to kubelet sync.
* [FEATURE] Allow default base images to be configurable.
* [FEATURE] Configure Prometheus to only use necessary namespaces.
* [ENHANCEMENT] Improve rollout detection for Alertmanager clusters.
* [BUGFIX] Fix targetPort relabeling.

## 0.10.2 / 2017-06-21

* [BUGFIX] Use computed route prefix instead of directly from manifest.

## 0.10.1 / 2017-06-13

Attention: if the basic auth feature was previously used, the `key` and `name`
fields need to be switched. This was not intentional, and the bug is not fixed,
but causes this change.

* [CHANGE] Prometheus default version v1.7.1.
* [CHANGE] Alertmanager default version v0.7.1.
* [BUGFIX] Fix basic auth secret key selector `key` and `name` switched.
* [BUGFIX] Fix route prefix flag not always being set for Prometheus.
* [BUGFIX] Fix nil panic on replica metrics.
* [FEATURE] Add ability to specify Alertmanager path prefix for Prometheus.

## 0.10.0 / 2017-06-09

* [CHANGE] Prometheus route prefix defaults to root.
* [CHANGE] Default to Prometheus v1.7.0.
* [CHANGE] Default to Alertmanager v0.7.0.
* [FEATURE] Add route prefix support to Alertmanager resource.
* [FEATURE] Add metrics on expected replicas.
* [FEATURE] Support for runing Alertmanager v0.7.0.
* [BUGFIX] Fix sensitive rollout triggering.

## 0.9.1 / 2017-05-18

* [FEATURE] Add experimental Prometheus 2.0 support.
* [FEATURE] Add support for setting Prometheus external labels.
* [BUGFIX] Fix non-deterministic config generation.

## 0.9.0 / 2017-05-09

* [CHANGE] The `kubelet-object` flag has been renamed to `kubelet-service`.
* [CHANGE] Remove automatic relabelling of Pod and Service labels onto targets.
* [CHANGE] Remove "non-namespaced" alpha annotation in favor of `honor_labels`.
* [FEATURE] Add ability make use of the Prometheus `honor_labels` configuration option.
* [FEATURE] Add ability to specify image pull secrets for Prometheus and Alertmanager pods.
* [FEATURE] Add basic auth configuration option through ServiceMonitor.
* [ENHANCEMENT] Add liveness and readiness probes to Prometheus and Alertmanger pods.
* [ENHANCEMENT] Add default resource requests for Alertmanager pods.
* [ENHANCEMENT] Fallback to ExternalIPs when InternalIPs are not available in kubelet sync.
* [ENHANCEMENT] Improved change detection to trigger Prometheus rollout.
* [ENHANCEMENT] Do not delete unmanaged Prometheus configuration Secret.

## 0.8.2 / 2017-04-20

* [ENHANCEMENT] Use new Prometheus 1.6 storage flags and make it default.

## 0.8.1 / 2017-04-13

* [ENHANCEMENT] Include kubelet insecure port in kubelet Enpdoints object.

## 0.8.0 / 2017-04-07

* [FEATURE] Add ability to mount custom secrets into Prometheus Pods. Note that
  secrets cannot be modified after creation, if the list if modified after
  creation it will not effect the Prometheus Pods.
* [FEATURE] Attach pod and service name as labels to Pod targets.

## 0.7.0 / 2017-03-17

This release introduces breaking changes to the generated StatefulSet's
PodTemplate, which cannot be modified for StatefulSets. The Prometheus and
Alertmanager objects have to be deleted and recreated for the StatefulSets to
be created properly.

* [CHANGE] Use Secrets instead of ConfigMaps for configurations.
* [FEATURE] Allow ConfigMaps containing rules to be selected via label selector.
* [FEATURE] `nodeSelector` added to the Alertmanager kind.
* [ENHANCEMENT] Use Prometheus v2 chunk encoding by default.
* [BUGFIX] Fix Alertmanager cluster mesh initialization.

## 0.6.0 / 2017-02-28

* [FEATURE] Allow not tagging targets with the `namespace` label.
* [FEATURE] Allow specifying `ServiceAccountName` to be used by Prometheus pods.
* [ENHANCEMENT] Label governing services to uniquely identify them.
* [ENHANCEMENT] Reconcile Service and Endpoints objects.
* [ENHANCEMENT] General stability improvements.
* [BUGFIX] Hostname cannot be fqdn when syncing kubelets into Endpoints object.

## 0.5.1 / 2017-02-17

* [BUGFIX] Use correct governing `Service` for Prometheus `StatefulSet`.

## 0.5.0 / 2017-02-15

* [FEATURE] Allow synchronizing kubelets into an `Endpoints` object.
* [FEATURE] Allow specifying custom configmap-reload image

## 0.4.0 / 2017-02-02

* [CHANGE] Split endpoint and job in separate labels instead of a single
  concatenated one.
* [BUGFIX] Properly exit on errors communicating with the apiserver.

## 0.3.0 / 2017-01-31

This release introduces breaking changes to the underlying naming schemes. It
is recommended to destroy existing Prometheus and Alertmanager objects and
recreate them with new namings.

With this release support for `v1.4.x` clusters is dropped. Changes will not be
backported to the `0.1.x` release series anymore.

* [CHANGE] Prefixed StatefulSet namings based on managing resource
* [FEATURE] Pass labels and annotations through to StatefulSets
* [FEATURE] Add tls config to use for Prometheus target scraping
* [FEATURE] Add configurable `routePrefix` for Prometheus
* [FEATURE] Add node selector to Prometheus TPR
* [ENHANCEMENT] Stability improvements

## 0.2.3 / 2017-01-05

* [BUGFIX] Fix config reloading when using external url.

## 0.1.3 / 2017-01-05

The `0.1.x` releases are backport releases with Kubernetes `v1.4.x` compatibility.

* [BUGFIX] Fix config reloading when using external url.

## 0.2.2 / 2017-01-03

* [FEATURE] Add ability to set the external url the Prometheus/Alertmanager instances will be available under.

## 0.1.2 / 2017-01-03

The `0.1.x` releases are backport releases with Kubernetes `v1.4.x` compatibility.

* [FEATURE] Add ability to set the external url the Prometheus/Alertmanager instances will be available under.

## 0.2.1 / 2016-12-23

* [BUGFIX] Fix `subPath` behavior when not using storage provisioning

## 0.2.0 / 2016-12-20

This release requires a Kubernetes cluster >=1.5.0. See the readme for
instructions on how to upgrade if you are currently running on a lower version
with the operator.

* [CHANGE] Use StatefulSet instead of PetSet
* [BUGFIX] Fix Prometheus config generation for labels containing "-"
