## 0.65.1 / 2022-05-05

* [BUGFIX] Fix panic when ScrapeConfig CRD is not installed. #5550

## 0.65.0 / 2022-05-04

The main change introduced by this release is the new v1alpha1 `ScrapeConfig` CRD.
This implements the [proposal](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/proposals/202212-scrape-config.md)
documented in [#5279](https://github.com/prometheus-operator/prometheus-operator/pull/5279)
and provides a Kubernetes native API to create and manage additional scrape configurations.

To try it, follow the following steps:
1. Install the new CRD in the cluster (see
   `example/prometheus-operator-crd/monitoring.coreos.com_scrapeconfigs.yaml`).
2. Update the Prometheus operator's RBAC permissions to manage `ScrapeConfig` resources
   (see `example/rbac/prometheus-operator/prometheus-operator-cluster-role.yaml`).

**NOTE**: if these conditions aren't met, the operator will start but it won't
be able to reconcile the `ScrapeConfig` resources.

* [FEATURE] Add the `status` subresource for the `ThanosRuler` CRD. #5520
* [FEATURE] Add `spec.web.timeout` and `spec.web.getConcurrency` to the `Alertmanager` CRD. #5478
* [FEATURE] Add `spec.groups[].limit` to the `Prometheus` CRD. #4999
* [FEATURE] Add ScrapeConfig CRD. #5335
* [ENHANCEMENT] Set a default for `seccompProfile` on the operator and webhook Deployments to `RuntimeDefault`. #5477
* [ENHANCEMENT] Add optional liveness and readiness probes to `prometheus-config-reloader`. This can be enabled via the `--enable-config-reloader-probes` CLI flag. #5449
* [BUGFIX] Don't start the `PrometheusAgent` controller if the CRD isn't present or the operator lacks permissions. #5476
* [BUGFIX] Declare `spec.rules` optional in `PrometheusRule` CRD. #5481
* [BUGFIX] Fix incorrect metric counter value for failed sync status. #5533

## 0.64.1 / 2023-04-24

* [BUGFIX] Fix panic when scraping `/metrics` with PrometheusAgent resources declared. #5511

## 0.64.0 / 2023-03-29

This release provides first-class support for running Prometheus in agent mode
with the new `PrometheusAgent` CRD. As the v1alpha1 version tells it, we don't
recommend using it in production but we're eager to hear all possible feedback.

To try it, follow the following steps:
1. Install the new CRD in the cluster (see
   `example/prometheus-operator-crd/monitoring.coreos.com_prometheusagents.yaml`).
2. Update the Prometheus operator's RBAC permissions to manage PrometheusAgents resources
   (see `example/rbac/prometheus-operator/prometheus-operator-cluster-role.yaml`).

**NOTE**: if these conditions aren't met, the operator will start but it won't
be able to reconcile the PrometheusAgent resources.

For the first time, the container images associated to this release are signed
using [sigstore](https://www.sigstore.dev/).

* [CHANGE] Remove the `/apis` endpoints from the operator's web server. #5396
* [CHANGE] Set default default value of `spec.portName` to `web`. #5350
* [FEATURE] Add v1alpha1 `PrometheusAgent` CRD to run Prometheus in agent mode. #5385
* [FEATURE] Add `--reload-timeout` argument to the config-reloader binary which controls how long the program will wait for the reload operation to complete (default: 30s). #5349
* [ENHANCEMENT] Set web server's `ReadTimeout` and `ReadHeaderTimeout` to 30s for Prometheus operator and config-reloader to avoid potential slowloris attacks. #5340
* [ENHANCEMENT] Add support for `DropEqual` and `KeepEqual` relabeling actions. #5368
* [ENHANCEMENT] Drop invalid `PrometheusRule` objects instead of failing the reconciliation of Prometheus and ThanosRuler objects. #5221
* [ENHANCEMENT] Add `spec.thanos.blockSize` field to the `Prometheus` CRD. #5360
* [ENHANCEMENT] Add `spec.thanos.configTimeout` and `spec.thanos.configInterval` to the Prometheus CRD. #5399
* [ENHANCEMENT] Add `spec.alertmanagerConfiguration.global.slackApiUrl` field to the `Alertmanager` CRD. #5383
* [ENHANCEMENT] Add `spec.alertmanagerConfiguration.global.opsGenieApiUrl` and `spec.alertmanagerConfiguration.global.opsGenieApiKey` fields to the `Alertmanager` CRD. #5422
* [ENHANCEMENT] Reduce the operator's memory usage by using metadata informers for Kubernetes secrets and configmaps. #5424 #5448
* [BUGFIX] Add `init-config-reloader` init container to avoid a restart of the Alertmanager's `config-reloader` container when the pod starts. #5358

## 0.63.0 / 2023-02-08

* [CHANGE] Use `tmpfs` to store `Prometheus` and `Alertmanager` configuration. #5311
* [FEATURE] Add `status` subresource to the `Alertmanager` CRD. #5270
* [FEATURE] Add `spec.additionalArgs` to the `ThanosRuler` CRD. #5293
* [ENHANCEMENT] Add `spec.web.maxConnections` to the `Prometheus` CRD. #5175
* [BUGFIX] Fix unsupported types in Alertmanager route sanitizer log lines. #5296
* [BUGFIX] Fix `ThanosRuler` StatefulSet re-creation bug when labels are specified. #5318

## 0.62.0 / 2023-01-04

* [CHANGE] Use `spec.version` of the Prometheus object as the image's tag if the image name is untagged. #5171
* [FEATURE] Generate "apply configuration" types. #5243
* [FEATURE] Add `spec.podTargetLabels` field to the Prometheus CRD for adding pod target labels to every PodMonitor/ServiceMonitor. #5206
* [FEATURE] Add `spec.version` field to the ThanosRuler CRD. #5177
* [ENHANCEMENT] Add `basicAuth` field to the Prometheus CRD for alerting configuration of Prometheus. #5170
* [ENHANCEMENT] Add `spec.imagePullPolicy` to Prometheus, Alertmanager and ThanosRuler CRDs. #5203
* [ENHANCEMENT] Add `activeTimeIntervals` field to AlertmanagerConfig CRD. #5198
* [ENHANCEMENT] Support `time_intervals` and `active_time_intervals` in the Alertmanager configurations. #5135
* [ENHANCEMENT] Support new fields in the Alertmanager v0.25.0 configuration. #5254 #5263

## 0.61.1 / 2022-11-24

* [BUGFIX] Fixed a regression that caused the ThanosRuler statefulsets to be stuck after upgrading the operator to v0.61.0. #5183

## 0.61.0 / 2022-11-16

* [CHANGE] Updated `RuleGroup` description and add validation for the CRD. #5072
* [CHANGE] Removed validations in the operator that are already covered at the CRD level. #5108
* [CHANGE] jsonnet: Enforced existence of the TLS secret for the admission webhook deployment. #5112
* [CHANGE] jsonnet: Changed default port of the admission webhook service from 8443 to 443. #5112
* [CHANGE] Added a filter for non-running pods in the `ServiceMonitor` CRD. #5149
* [FEATURE] Added `spec.attachMetadata.node` in the `ServiceMonitor` CRD. #5147
* [ENHANCEMENT] Updated `ProbeTLSConfig` and `SafeTLSConfig` description. #5081
* [ENHANCEMENT] Updated admission webhook deployment's jsonnet to avoid down-time on updates. #5099
* [ENHANCEMENT] Added the `filterExternalLabels` field to the remote read configuration of the `Prometheus` CRD. #5142
* [ENHANCEMENT] Added `enableHttp2` field to `AlertingEndpoints` #5152
* [ENHANCEMENT] Updated `ThanosRuler` arguments (`QueryConfig`, `AlertManagerConfig`, `ObjectStorageConfig` and `TracingConfig`) to be directly read from secrets instead of using ENV vars. #5122
* [ENHANCEMENT] Add `alertmanagerConfigMatcherStrategy` to `Alertmanager` CRD in order to disable auto-generated namespace matchers. #5084
* [BUGFIX] Ignore `PartialResponseStrategy` in the `Prometheus` CRD. This field is only applicable for the Thanos Ruler. #5125

## 0.60.1 / 2022-10-10

* [BUGFIX] Fixed configuration when `spec.tsdb.outOfOrderTimeWindow` is set in the Prometheus CRD. #5078

## 0.60.0 / 2022-10-06

* [CHANGE] Added `filterRunning` field to the PodMonitor CRD. By default, non-running pods are dropped by the Prometheus service discovery. To preserve the old behavior and keep pods which aren't running, set `filterRunning: false`. #5049
* [FEATURE] Added `grpcListenLocal` and `httpListenLocal` fields to the Thanos sidecar configuration of the Prometheus CRD. #5045
* [FEATURE] Added `hostNetwork` field to the Prometheus CRD. #5010
* [FEATURE] Added `spec.tsdb.outOfOrderTimeWindow` field to the Prometheus CRD to allow out-of-order samples in TSDB. #5071
* [ENHANCEMENT] Added columns showing the Prometheus conditions to the output of `kubectl get prometheus`. #5055
* [ENHANCEMENT] Added `observedGeneration` field to the Prometheus status conditions. #5005

## 0.59.2 / 2022-09-20

* [CHANGE/BUGFIX] Removed `FOWNER` capability from the Thanos sidecar. #5030

## 0.59.1 / 2022-09-12

* [BUGFIX] Fixed secret and configmap volume names that need to be mounted in additional containers. #5000
* [BUGFIX] Removed `CAP_FOWNER` capability for the Thanos sidecar when not required. #5004
* [BUGFIX] Removed the `CAP_` prefix of the `FOWNER` capability on Thanos sidecar. #5014

## 0.59.0 / 2022-09-02

* [FEATURE] Added validations for timeout and time settings of alertmanager at CRD level. #4898
* [FEATURE] Added support for global `resolveTimeout` and `httpConfig` in Alertmanager CRD. #4622
* [FEATURE] Added support for `additionalArgs` field to the Prometheus CRD for Prometheus, Alertmanager and Thanos sidecar. #4863
* [ENHANCEMENT] Added `tracingConfigFile` option to ThanosRuler CRD. #4962
* [BUGFIX] Fixed compress alertmanager secret to circumvent maximum size limit of 1048576 bytes. #4906
* [BUGFIX] Fixed namespace enforcement exclusion on newly created Prometheus objects. #4915
* [BUGFIX] Fixed `CAP_FOWNER` capability to Thanos sidecar container. #4931
* [BUGFIX] Fixed `spec.query.maxSamples` and `spec.query.maxConcurrency` fields of Prometheus CRD. #4951
* [BUGFIX] Fixed Thanos sidecar connectivity issue when Prometheus TLS is enabled. #4954
* [BUGFIX] Fixed Prometheus and Alertmanager Pods not created when Secret name exceeds 63 characters. #4988

## 0.58.0 / 2022-07-19

* [FEATURE] Add validations for timeout and time settings of alertmanager at CRD level. #4827, #4881
* [FEATURE] Extend the PrometheusSpec to allow to configure the `max_exemplars`. #4834
* [FEATURE] Add support for web TLS configuration for Alertmanager CRD. #4868
* [ENHANCEMENT] Add support for `uppercase`, `lowercase`, and `CamelCase` relabel actions. #4840, #4873
* [ENHANCEMENT] Added support for `enable_http2` in endpoint scrape configuration. #4836
* [BUGFIX] Fixed missing conversion of the `followRedirects` field in HTTP configuration for AlertmanagerConfig v1beta1. #4854
* [BUGFIX] fix AlertmanagerConfig.Spec.Route nil panic. #4853
* [BUGFIX] Optimise warning log message during sanitization of OpsGenie configuration. #4833

## 0.57.0 / 2022-06-02

The main change introduced by this release is a new v1beta1 API version for the
AlertmanagerConfig CRD.

Changes compared to the v1alpha1 API:
* Renamed `spec.muteTimeIntervals` field to `to spec.timeIntervals`.
* Removed `regex` field from the `Matcher` type.
* Replaced all `v1.SecretKeySelector` types by the `SecretKeySelector` type
  * Removed `optional` field.
  * `name` and `key` fields are required.

As a pre-requisite, you need to deploy the admission webhook and configure the
conversion webhook in the AlertmanagerConfig CRD object so that users can use
both v1alpha1 and v1beta1 versions at the same time. There are more details in
`Documentation/user-guides/webhook.md` about the webhook configuration.

Because of the conversion webhook requirement, the new version is an opt-in
feature: the `bundle.yaml` file and the manifests from
`example/prometheus-operator-crd` don't deploy the new API version (the
manifests to enable the v1beta1 version are under the
`example/prometheus-operator-crd-full` directory). We will wait until v0.59.0
(at least) before enabling the new API version by default.

* [CHANGE] Added validations at the API level for the time-based fields of the ThanosRuler CRD. #4815
* [CHANGE] Added validations at the API level for the OpsGenie's `responders` field of the AlertmanagerConfig CRD. #4725
* [FEATURE] Added v1beta1 version for AlertmanagerConfig CRD. #4709
* [FEATURE] Added support for Telegram receiver in the AlertmanagerConfig CRD. #4726
* [FEATURE] Added `updateAlerts` field to the OpsGenie configuration of the AlertmanagerConfig CRD. #4726
* [FEATURE] Added `hostAliases` field to the the Alertmanager, Prometheus and ThanosRuler CRDs. #4787
* [ENHANCEMENT] Added configuration option in the jsonnet mixins to specify the aggregation labels. #4814
* [ENHANCEMENT] Added `attachMetadata` field to the PodMonitor CRD. #4792
* [BUGFIX] Fixed the curl command for exec probes when `listenLocal` is set to true in the Prometheus object. It avoids temporary service outage due to long WAL replays. #4804

## 0.56.3 / 2022-05-23

* [BUGFIX] Fixed errors for Alertmanager configurations using the new `entity`, `actions` and `opsgenie_api_key_file` fields. #4797
* [BUGFIX] Fixed high CPU usage by reducing the number of number of reconciliations on Prometheus objects. #4798 #4806

## 0.56.2 / 2022-05-09

* [BUGFIX] Fix StatefulSet spec's generation to be determistic when `spec.containers` is not empty. #4772

## 0.56.1 / 2022-05-03

* [BUGFIX] Avoid unnecessary updates of the Prometheus StatefulSet object. #4762

## 0.56.0 / 2022-04-20

* [CHANGE] Added validation at the API level for size-based fields of the Prometheus CRD. #4661
* [CHANGE] Added validation at the API level for log level and format fields of the Alertmanager, Prometheus and ThanosRuler CRDs. #4638
* [CHANGE] Added validation at the API level for duration and time-based fields of the Prometheus CRD. #4684
* [CHANGE] Added shortnames for custom resources (`amcfg` for AlertmanagerConfig, `am` for Alertmanager, `pmon` for PodMonitor, `prb` for Probe, `prom` for Prometheus, `smon` for ServiceMonitor, `ruler` for Thanos Ruler). #4680
* [FEATURE] Added `status` subresource to the Prometheus CRD. #4580
* [ENHANCEMENT] Added `excludedFromEnforce` field to the Prometheus CRD. It allows to define PodMonitor, ServiceMonitor, Probe or PrometheusRule objects for which the namespace label enforcement (if enabled) should not be applied. This deprecates `prometheusRulesExcludedFromEnforce` which is still supported but users are encouraged to migrate to the new field. #4397
* [ENHANCEMENT] Added `enableRemoteWriteReceiver` field to the Prometheus CRD. #4633
* [ENHANCEMENT] Added `entity` and `actions` fields for the OpsGenie receiver to the AlertmanagerConfig CRD. #4697
* [ENHANCEMENT] Added `prometheus_operator_reconcile_duration_seconds` histogram metric. #4706
* [BUGFIX] Added support for `opsgenie_api_key_file` and `api_key_file` in the generated Alertmanager configuration. #4666 #4738

## 0.55.1 / 2022-03-26

* [BUGFIX] Fixed Prometheus configuration when `spec.queryLogFile` has no path ("query.log" for instance). #4683

## 0.55.0 / 2022-03-09

* [CHANGE] Enabled read-only root filesystem for containers generated from the Prometheus, Alertmanager and ThanosRuler objects. #4552
* [CHANGE] Disabled privilege escalation for the containers generated from Prometheus, Alertmanager and ThanosRuler objects. #4552
* [CHANGE] Dropped all capabilities for the containers generated from Prometheus, Alertmanager and ThanosRuler objects. #4552
* [CHANGE] Added `emptyDir` volume to the Prometheus statefulset when `spec.queryLogFile` is only a base filename (e.g. `query.log` as opposed to `/tmp/query.log`). When the path contains a full path, a volume + volume mount should be explicitly given in the Prometheus spec since the root file system is now read-only. #4566
* [CHANGE/BUGFIX] Added skip TLS verify for the config-reloader HTTP client when informing Prometheus/Alertmanager on a config reload (partial fix for #4273). #4592
* [CHANGE] Switched using the `endpointslice` role for Prometheus by default if it is supported by the Kubernetes API. #4535
* [FEATURE] Added standalone admission webhook. #4494
* [FEATURE] Support the definition of Alertmanager configuration via AlertManagerConfig instead of Kubernetes secret (EXPERIMENTAL). #4220
* [FEATURE] Support sharding for Probe objects. #4587
* [ENHANCEMENT] Restore Prometheus StatefulSet liveness probe so that deadlocks are detected and recovered from. #4387, #4534
* [ENHANCEMENT] Added `-alertmanager-config-namespaces` CLI flag to the operator. #4455, #4619
* [ENHANCEMENT] `remoteWrite` and `remoteRead` fields of the Prometheus CRD not considered experimental anymore. #4555
* [ENHANCEMENT] Added support for follow_redirects in endpoint scrape configuration. #4563
* [ENHANCEMENT] Added support for Oauth2 in AlertmanagerConfig CRD. #4501
* [ENHANCEMENT] Improved logging when the given Prometheus version doesn't support some CR fields. #4571
* [ENHANCEMENT] Added `__tmp_ingress_address` label to preserve the initial host address of the ingress object. #4603
* [ENHANCEMENT] Fixed potential name collisions in Alertmanager configuration when merging AlertmanagerConfig objects. #4626
* [BUGFIX] Fixed panic when validating `Probe`. #4541
* [BUGFIX] Added validation for sourceLabels in relabel configuration. #4568
* [BUGFIX] Allow retention to be set only by size. #4590

## 0.54.1 / 2022-02-24

* [BUGFIX] Updated relabelConfig validation to accept Prometheus default config on labeldrop relabelConfig. #4579
* [BUGFIX] Fixed relabelConfigs for labelmap action. #4574

## 0.54.0 / 2022-01-26

* [FEATURE] Support SNS Receiver in AlertmanagerConfig CR. #4468
* [ENHANCEMENT] Specify SA token automounting on pod-level for operator and prometheus operand. #4514
* [ENHANCEMENT] Support following redirects and Oauth2 in HTTP Client config in raw alertmanager config secret. #4499
* [ENHANCEMENT] Add Replicas column for Thanos Ruler. #4496
* [ENHANCEMENT] Set User-Agent for the kubernetes client. #4506
* [BUGFIX] Avoid race during recreation of StatefulSet(s). #4504
* [BUGFIX] Add validation for proberSpec `url` field in `ProbeSpec`. #4483
* [BUGFIX] Add validation for relabel configs. #4429
* [BUGFIX] Add validation for scrapeTimeout validation. #4491

## 0.53.1 / 2021-12-20

* [BUGFIX] Fixed the validation pattern for the `february` month in the AlertManagerConfig CRD. #4458

## 0.53.0 / 2021-12-16

* [CHANGE] Added startup probe to Prometheus. #4433 #4369
* [FEATURE] Added support for mute time intervals in the AlertManagerConfig CRD. #4388
* [FEATURE] Added support for new matching syntax in the routes configuration of the AlertmanagerConfig CRD. #4332
* [FEATURE] Added support for new matching syntax in the inhibit rules configuration of the AlertmanagerConfig CRD. #4329
* [FEATURE] Added `headers` in the Remote Read configuration of the Prometheus CRD. #4323
* [FEATURE] Added `retryOnRateLimit` in the Remote Write configuration of the Prometheus CRD. #4420
* [FEATURE] Added support for PagerDuty links and images in the AlertManagerConfig CR. #4425
* [ENHANCEMENT] Optimized the generated Prometheus configuration to make it compatible with the new Prometheus agent mode. #4417
* [BUGFIX] Improved validation for the Alertmanager CRD to match with upstream Alertmanager. #4434
* [BUGFIX] Fixed propagation of annotations from spec.podMetadata to the StatefulSet's pods. #4422
* [BUGFIX] Ensured that `group_by` values are unique in the generated Alertmanager configuration. #4413
* [BUGFIX] Fixed generated secrets being larger than the maximum size limit of 1048576 bytes. #4427 #4449

## 0.52.1 / 2021-11-16

* [BUGFIX] Fixed regex in relabel_configs. #4395

## 0.52.0 / 2021-11-03

* [CHANGE] Extend sharding capabilities to `additionalScrapeConfigs`. #4324
* [CHANGE] Remove `app` label from Prometheus, Alertmanager and Thanos Ruler statefulsets/pods. #4350
* [FEATURE] Add `alertRelabelConfigs` field to the Thanos Ruler CRD for configuring Prometheus alert relabeling features. #4303
* [FEATURE] Add support for updated matching syntax in Alertmanager's raw config for `inhibit_rules` and `route`. #4307, #4309
* [FEATURE] Add validating webhook for AlertManagerConfig. #4338
* [FEATURE] Adds support for Sigv4 when configuring RemoteWrite. #3994
* [ENHANCEMENT] Add "generic ephemeral storage" as a data storage option for Alertmanager, Prometheus and Thanos Ruler. #4326
* [ENHANCEMENT] Improve docs and error message for "smarthost" field. #4299
* [ENHANCEMENT] Add alerts for config reloader sidecars. #4294
* [ENHANCEMENT] Add validations for duration and size fields for Prometheus, Alertmanager, and Thanos Ruler resources #4308, #4352
* [ENHANCEMENT] Add s390x support to docker images. #4351
* [ENHANCEMENT] Only load alertmanager configuration when writing configuration. #4333
* [BUGFIX] Fix `matchLabels` selector to have empty label values in ServiceMonitor, PodMonitor, and Probe. #4327
* [BUGFIX] Prevent rule file name collision. #4347
* [BUGFIX] Update native kubernetes fields used in prometheus-operator CRDs. #4221

## 0.51.2 / 2021-10-04

* [BUGFIX] Validated the value of the `EnforcedBodySizeLimit` field to avoid Prometheus crash. #4285

## 0.51.1 / 2021-09-27

No change since v0.51.0.

*The CI automation failed to build the v0.51.0 images so we had to create a new patch release.*

## 0.51.0 / 2021-09-24

* [FEATURE] Added `metricRelabelings` field to the Probe CRD for configuring the metric relabel configs. #4226
* [FEATURE] Added `volumeMounts` field to the Prometheus CRD for configuring the volume mounts of the thanos-sidecar container. #4238
* [FEATURE] Added `enforcedBodySizeLimit` field to the Prometheus CRD. #4275
* [FEATURE] Added `authorization` field to all HTTP configurations in the AlertmanagerConfig CRD. #4110
* [FEATURE] Added `minReadySeconds` field to AlertManager, Prometheus and ThanosRuler CRDs. #4246
* [FEATURE] Added support for Slack webhook URL via file path in the Alertmanager configuration secret. #4234
* [FEATURE] Added support the `authorization` field for all HTTP configurations in the Alertmanager configuration secret. #4234
* [ENHANCEMENT] Improved detection and rollback of manual changes to Alertmanager statefulsets. #4228
* [BUGFIX] Invalid probes are discarded instead of stopping after the first error when reconciling probes. #4248
* [BUGFIX] Empty basic auth username is allowed in the AlertmanagerConfig CRD. #4260
* [BUGFIX] Update conflicts for secrets are handled properly which avoids overwriting user-defined metadata. #4235
* [BUGFIX] The namespace label is always enforced with metricRelabelConfigs. #4272

## 0.50.0 / 2021-08-17

* [CHANGE] Remove deprecated flags `--config-reloader-memory` and `--config-reloader-cpu` in favor of `--config-reloader-memory-limit`, `--config-reloader-memory-request`, `--config-reloader-cpu-limit`, and `--config-reloader-cpu-request`. #3884
* [CHANGE] Remove use of Kubernetes API versions being removed in v1.22. #4171
* [FEATURE] Added support for OAuth2 authentication in remote read and remote write configuration. #4113
* [FEATURE] Added OAuth2 configuration for ServiceMonitor, PodMonitor and Probe. #4170
* [FEATURE] Added `prometheus_operator_spec_shards` metric for exposing the number of shards set on prometheus operator spec. #4173
* [FEATURE] Support for `Authorization` section in various prometheus sections. #4180
* [FEATURE] Validate prometheus rules when generating rule file content. #4184
* [FEATURE] Support `label_limit`, `label_name_length_limit` and `label_value_length_limit` configuration fields at the Prometheus CRD level as well as support individual limits per ServiceMonitor, PodMonitor and Probe resources. #4195
* [FEATURE] Added sample and target limits to Probe. #4207
* [FEATURE] Added `send_exemplars` field to the `remote_write` configuration in Prometheus. #4215 #4160
* [ENHANCEMENT] Support loading ClusterConfig from concatenated KUBECONFIG env. #4154
* [ENHANCEMENT] Include PrometheusRule in prometheus-operator CRD category. #4213
* [ENHANCEMENT] Preserve annotations set by kubectl. #4185
* [BUGFIX] Thanos: listen to all available IP addresses instead of `POD_IP`, simplifies istio management. #4038
* [BUGFIX] Add port name mapping to ConfigReloader to avoid reloader-web probe failure. #4187
* [BUGFIX] Handle Thanos rules `partial_response_strategy` field in validating admission webhook. #4217

## 0.49.0 / 2021-07-06

* [CHANGE] Flag "storage.tsdb.no-lockfile" will now default to false. #4066
* [CHANGE] Remove `app.kubernetes.io/version` label selector from Prometheus and Alertmanager statefulsets. #4093
* [CHANGE] Exit if the informers cache synchronization doesn't complete after 10 minutes. #4143, #4149
* [FEATURE] Added web TLS configuration support for Prometheus. #4025
* [FEATURE] Add proxy_url support for Probes. #4043
* [ENHANCEMENT] Set proper build context in version package. #4019
* [ENHANCEMENT] Publish images on GitHub Container Registry. #4060
* [ENHANCEMENT] Automatically generate document for operator executable. #4112
* [ENHANCEMENT] Adds configuration to set the Prometheus ready timeout to Thanos sidecar #4118
* [BUGFIX] Fixed bug in Alertmanager config where URLS that are taken from Kubernetes secrets might contain whitespace or newline characters. #4068
* [BUGFIX] Generate correct scraping configuration for Probes with empty or unset `module` parameter. #4074
* [BUGFIX] Operator does not generate `max_retries` option in `remote_write` for Prometheus version 2.11.0 and higher. #4103
* [BUGFIX] Preserve the dual-stack immutable fields on service sync. #4119

## 0.48.1 / 2021-06-01

* [BUGFIX] Added an `app` label on Prometheus pods. #4055

## 0.48.0 / 2021-05-19

Deprecation notice:
app labels will be removed in v0.50.

* [CHANGE] Replace app label names with app.kubernetes.io/name. #3939
* [CHANGE] Drop ksonnet as a dependency in jsonnetfile.json. #4002
* [ENHANCEMENT] Add default container annotation to Alertmanager pod. #3978
* [ENHANCEMENT] Add default container annotation to Thanos ruler pod. #3981
* [ENHANCEMENT] Optimize asset secret update logic. #3986
* [ENHANCEMENT] jsonnet: set default container in prometheus-operator pod. #3979
* [BUGFIX] Watch configmaps from the Prometheus allowed namespaces only. #3992
* [BUGFIX] Reconcile resources on namespace updates when using privileged lister/watcher. #3879
* [BUGFIX] Don't generate broken Alertmanager configuration when `http_config` is defined in the global parameters. #4041

## 0.47.1 / 2021-04-30

* [BUGFIX] Avoid reconciliations for Alertmanager statefulset on resource version changes. #3948

## 0.47.0 / 2021-04-13

The `--config-reloader-cpu` and `--config-reloader-memory` flags are deprecated
and will be removed in v0.49.0. They are replaced respectively by the
`--config-reloader-cpu-request`/`--config-reloader-cpu-limit` and
`config-reloader-memory-request`/`config-reloader-memory-limit` flags.

* [FEATURE] Add `enableFeatures` field to the Prometheus CRD for enabling feature flags. #3878
* [FEATURE] Add `metadataConfig` field to the Prometheus CRD for configuring how remote-write sends metadata information. #3915
* [FEATURE] Add support for TLS and authentication configuration to the Probe CRD. #3876
* [ENHANCEMENT] Allow CPU requests/limits and memory requests/limits of the config reloader to be set independently. #3826
* [ENHANCEMENT] Add rules validation to `po-lint`. #3894
* [ENHANCEMENT] Add common Kubernetes labels to statefulset objects managed by the Prometheus operator. #3841
* [ENHANCEMENT] Avoid unneeded synchronizations on Alertmanager updates. #3943
* [ENHANCEMENT] Retain the original job's name `__tmp_prometheus_job_name label` in the generated scrape configurations. #3828
* [BUGFIX] Fix `app.kubernetes.io/managed-by` label on kubelet endpoint object. #3902
* [BUGFIX] Avoid name collisions in the generated Prometheus configuration. #3913
* [BUGFIX] Restore `prometheus_operator_spec_replicas` metrics for Alertmanager and Thanos Ruler controllers. #3924
* [BUGFIX] Allow `smtp_require_tls` to be false in Alertmanager configuration. #3960

## 0.46.0 / 2021-02-24

* [CHANGE] Drop support for Prometheus 1.x #2822
* [FEATURE] Add relabelingConfigs to ProbeTargetStaticConfig #3817
* [ENHANCEMENT] Add custom HTTP headers in remoteWrite-config #3851
* [ENHANCEMENT] CRDs are now part of prometheus-operator group allowing `kubectl get prometheus-operator` operation #3843
* [ENHANCEMENT] Support app.kubernetes.io/managed-by label to kubelet service and endpoint objects. #3834
* [BUGFIX] Fix the loss of the `headers` key in AlertmanagerConfig. #3856
* [BUGFIX] Preserve user-added labels and annotations on Service, Endpoint and StatefulSet resources managed by the operator. #3810
* [BUGFIX] Do not require child routes in AlertmanagerConfig to have a receiver. #3749

## 0.45.0 / 2021-01-13

* [CHANGE] Add schema validations to AlertmanagerConfig CRD. #3742
* [CHANGE] Refactored jsonnet library to remove ksonnet and align with kube-prometheus #3781
* [ENHANCEMENT] Add `app.kubernetes.io/name` label to Kubelet Service/Endpoints object. #3768
* [ENHANCEMENT] Improve HTTP server's logging #3772
* [ENHANCEMENT] Add namespace label to static probe metrics #3752
* [ENHANCEMENT] Add `TracingConfigFile` field into thanos configuration. #3762
* [BUGFIX] Fix log messages when syncing node endpoints. #3758
* [BUGFIX] fix discovery of `AlertmanagerConfig` resources when `--alertmanager-instance-namespaces` is defined. #3759

## 0.44.1 / 2020-12-09

* [BUGFIX] Fix Alertmanager configuration for OpsGenie receiver. #3728

## 0.44.0 / 2020-12-02

* [CHANGE] Fix child routes support in AlertmanagerConfig. #3703
* [FEATURE] Add Slack receiver type to AlertmanagerConfig CRD. #3618
* [FEATURE] Add WeChat receiver type to AlertmanagerConfig CRD. #3619
* [FEATURE] Add Email receiver type to AlertmanagerConfig CRD. #3692
* [FEATURE] Add Pushover receiver type to AlertmanagerConfig CRD. #3697
* [FEATURE] Add VictorOps receiver type to AlertmanagerConfig CRD. #3701
* [FEATURE] Add sharding support for prometheus cluster. #3241
* [ENHANCEMENT] Add option to allow configuring object storage for Thanos. #3668
* [ENHANCEMENT] Add TLS support for remote read. #3714
* [ENHANCEMENT] Include EnforcedSampleLimit as a metric. #3617
* [ENHANCEMENT] Adjust config reloader memory requests and limits. #3660
* [ENHANCEMENT] Add `clusterGossipInterval`, `clusterPushpullInterval` and `clusterPeerTimeout` fields to Alertmanager CRD. #3663
* [BUGFIX] Handle all possible receiver types in AlertmanagerConfig. #3689
* [BUGFIX] Fix operator crashing on empty Probe targets. #3637
* [BUGFIX] Fix usage of `--prometheus-default-base-image`, `--alertmanager-default-base-image`, and `--thanos-default-base-image` flags. #3642
* [BUGFIX] Fix matching labels with empty values when using Exists/NotExists operators. #3686

## 0.43.2 / 2020-11-06

* [BUGFIX] Fix issue with additional data from the Alertmanager config's secret not being kept. #3647

## 0.43.1 / 2020-11-04

* [BUGFIX] Fix Alertmanager controller to wait for all informers to be synced before reconciling. #3641

## 0.43.0 / 2020-10-26

This release introduces a new `AlertmanagerConfig` CRD that allows to split the
Alertmanager configuration in different objects. For now the CRD only supports
the PagerDuty, OpsGenie and webhook receivers, [other
integrations](https://github.com/prometheus-operator/prometheus-operator/issues?q=is%3Aissue+is%3Aopen+%22receiver+type%22)
will follow in future releases of the operator. The current version of the CRD
is `v1alpha1` meaning that testing/feedback is encouraged and welcome but the
feature is not yet considered stable and the API is subject to change in the
future.

* [CHANGE] Use a single reloader sidecar (instead of 2) for Prometheus. The `--config-reloader-image` flag is deprecated and will be removed in a future release (not before v0.45.0). *Make sure to start the operator with a version of `--prometheus-config-reloader` that is at least `v0.43.0` otherwise the Prometheus pods will fail to start.* #3457
* [FEATURE] Add `targetLimit` and `enforcedTargetLimit` to the Prometheus CRD. #3571
* [FEATURE] Add initial support for `AlertmanagerConfig` CRD. #3451
* [FEATURE] Add support for Pod Topology Spread Constraints to Prometheus, Alertmanager, and ThanosRuler CRDs. #3598
* [ENHANCEMENT] Allow customization of the Prometheus web page title. #3525
* [ENHANCEMENT] Add metrics for selected/rejected resources and synchronization status. #3421
* [ENHANCEMENT] Configure Thanos sidecar for uploads only when needed. #3485
* [ENHANCEMENT] Add `--version` flag to all binaries + `prometheus_operator_build_info` metric. #359
* [ENHANCEMENT] Add `prometheus_operator_prometheus_enforced_sample_limit` metric. #3617
* [BUGFIX] Remove liveness probes to avoid killing Prometheus during the replay of the WAL. #3502
* [BUGFIX] Fix `spec.ipFamily: Invalid value: "null": field is immutable` error when updating governing services. #3526
* [BUGFIX] Generate more unique job names for Probes. #3481
* [BUGFIX] Don't block when the operator is configured to watch namespaces that don't exist yet. #3545
* [BUGFIX] Use `exec` in readiness probes to reduce the chance of leaking zombie processes. #3567
* [BUGFIX] Fix broken AdmissionReview. #3574
* [BUGFIX] Fix reconciliation when 2 resources share the same secret. #3590
* [BUGFIX] Discard invalid TLS configurations. #3578

## 0.42.1 / 2020-09-21

* [BUGFIX] Bump client-go to fix watch bug

## 0.42.0 / 2020-09-09

The Prometheus Operator now lives in its own independent GitHub organization.

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
* [FEATURE] Support for running Alertmanager v0.7.0.
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
