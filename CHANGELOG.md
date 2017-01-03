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

