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

