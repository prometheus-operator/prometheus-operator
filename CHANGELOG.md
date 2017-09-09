## 0.12.0 / 2017-08-24

Starting with this release only Kubernetes `v1.7.x` and up is supported as CustomResourceDefinitions are a requirement for the Prometheus Operator and are only available from those versions and up.

Additionally all objects have been promoted from `v1alpha1` to `v1`. On start up of this version of the Prometheus Operator the previously used `ThirdPartyResource`s and the associated `v1alpha1` objects will be automatically migrated to their `v1` equivalent `CustomResourceDefinition`.

* [CHANGE] All manifests created and used by the Prometheus Operator have been promoted from `v1alpha1` to `v1`.
* [CHANGE] Use Kubernetes `CustomResourceDefinition`s instead of `ThirdPartyResource`s.
* [FEATURE] Add ability to set scrape timeout to `ServiceMonitor`.
* [ENHANCEMENT] Use `StatefulSet` rolling deployments.
* [ENHANCEMENT] Properly set `SecurityContext` for Prometheus 2.0 deployments.
* [ENHANCEMENT] Enable web lifecycle APIs for Prometheus 2.0 deployments.

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
* [ENHANCEMENT] Reconcile Serive and Endpoints objects.
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

