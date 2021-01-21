# Prometheus Operator
[![Build Status](https://github.com/prometheus-operator/prometheus-operator/workflows/ci/badge.svg)](https://github.com/prometheus-operator/prometheus-operator/actions)
[![Go Report Card](https://goreportcard.com/badge/prometheus-operator/prometheus-operator "Go Report Card")](https://goreportcard.com/report/prometheus-operator/prometheus-operator)
[![Slack](https://img.shields.io/badge/join%20slack-%23prometheus--operator-brightgreen.svg)](http://slack.k8s.io/)

**Project status: *beta*** Not all planned features are completed. The API, spec, status and other user facing objects may change, but in a backward compatible way.

Note: Project was previously known as coreos/prometheus-operator.

## Overview

The Prometheus Operator provides [Kubernetes](https://kubernetes.io/) native deployment and management of
[Prometheus](https://prometheus.io/) and related monitoring components.  The purpose of this project is to
simplify and automate the configuration of a Prometheus based monitoring stack for Kubernetes clusters.

The Prometheus operator includes, but is not limited to, the following features:

* **Kubernetes Custom Resources**: Use Kubernetes custom resources to deploy and manage Prometheus, Alertmanager,
  and related components.

* **Simplified Deployment Configuration**: Configure the fundamentals of Prometheus like versions, persistence,
  retention policies, and replicas from a native Kubernetes resource.

* **Prometheus Target Configuration**: Automatically generate monitoring target configurations based
  on familiar Kubernetes label queries; no need to learn a Prometheus specific configuration language.

For an introduction to the Prometheus Operator, see the [getting started](https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/user-guides/getting-started.md) guide.

## Prometheus Operator vs. kube-prometheus vs. community helm chart

### Prometheus Operator
The Prometheus Operator uses Kubernetes [custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) to simplify the deployment and configuration of Prometheus, Alertmanager, and related monitoring components.

### kube-prometheus
[kube-prometheus](https://github.com/prometheus-operator/kube-prometheus) provides example configurations for a complete cluster monitoring
stack based on Prometheus and the Prometheus Operator.  This includes deployment of multiple Prometheus and Alertmanager instances,
metrics exporters such as the node_exporter for gathering node metrics, scrape target configuration linking Prometheus to various
metrics endpoints, and example alerting rules for notification of potential issues in the cluster.

### helm chart
The [prometheus-community/kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)
helm chart provides a similar feature set to kube-prometheus. This chart is maintained by the Prometheus community.
For more information, please see the [chart's readme](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack#kube-prometheus-stack)

## Prerequisites

Version `>=0.39.0` of the Prometheus Operator requires a Kubernetes
cluster of version `>=1.16.0`. If you are just starting out with the
Prometheus Operator, it is highly recommended to use the latest version.

If you have an older version of Kubernetes and the Prometheus Operator running,
we recommend upgrading Kubernetes first and then the Prometheus Operator.

## CustomResourceDefinitions

A core feature of the Prometheus Operator is to monitor the Kubernetes API server for changes
to specific objects and ensure that the current Prometheus deployments match these objects.
The Operator acts on the following [custom resource definitions (CRDs)](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/):

* **`Prometheus`**, which defines a desired Prometheus deployment.

* **`Alertmanager`**, which defines a desired Alertmanager deployment.

* **`ThanosRuler`**, which defines a desired Thanos Ruler deployment.

* **`ServiceMonitor`**, which declaratively specifies how groups of Kubernetes services should be monitored.
  The Operator automatically generates Prometheus scrape configuration based on the current state of the objects in the API server.

* **`PodMonitor`**, which declaratively specifies how group of pods should be monitored.
  The Operator automatically generates Prometheus scrape configuration based on the current state of the objects in the API server.

* **`Probe`**, which declaratively specifies how groups
  of ingresses or static targets should be monitored. The Operator automatically generates Prometheus scrape configuration
  based on the definition.

* **`PrometheusRule`**, which defines a desired set of Prometheus alerting and/or recording rules.
  The Operator generates a rule file, which can be used by Prometheus instances.

* **`AlertmanagerConfig`**, which declaratively specifies subsections of the Alertmanager configuration, allowing
  routing of alerts to custom receivers, and setting inhibit rules.

The Prometheus operator automatically detects changes in the Kubernetes API server to any of the above objects, and ensures that
matching deployments and configurations are kept in sync.

To learn more about the CRDs introduced by the Prometheus Operator have a look
at the [design doc](Documentation/design.md).

To automate validation of your CRD configuration files see about [linting](Documentation/user-guides/linting.md).

## Dynamic Admission Control

To prevent invalid Prometheus alerting and recording rules from causing failures in a deployed Prometheus instance,
an [admission webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
is provided to validate `PrometheusRule` resources upon initial creation or update.

For more information on this feature, see the [user guide](Documentation/user-guides/webhook.md).

## Quickstart

**Note:** this quickstart does not provision an entire monitoring stack; if that is what you are looking for,
see the [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus) project.  If you want the whole stack,
but have already applied the `bundle.yaml`, delete the bundle first (`kubectl delete -f bundle.yaml`).

To quickly try out _just_ the Prometheus Operator inside a cluster, **choose a release** and run the following command:

```sh
kubectl apply -f bundle.yaml
```

> Note: make sure to adapt the namespace in the ClusterRoleBinding if deploying in a namespace other than the default namespace.

To run the Operator outside of a cluster:

```sh
make
scripts/run-external.sh <kubectl cluster name>
```

## Removal

To remove the operator and Prometheus, first delete any custom resources you created in each namespace. The
operator will automatically shut down and remove Prometheus and Alertmanager pods, and associated ConfigMaps.

```sh
for n in $(kubectl get namespaces -o jsonpath={..metadata.name}); do
  kubectl delete --all --namespace=$n prometheus,servicemonitor,podmonitor,alertmanager
done
```

After a couple of minutes you can go ahead and remove the operator itself.

```sh
kubectl delete -f bundle.yaml
```

The operator automatically creates services in each namespace where you created a Prometheus or Alertmanager resources,
and defines three custom resource definitions. You can clean these up now.

```sh
for n in $(kubectl get namespaces -o jsonpath={..metadata.name}); do
  kubectl delete --ignore-not-found --namespace=$n service prometheus-operated alertmanager-operated
done

kubectl delete --ignore-not-found customresourcedefinitions \
  prometheuses.monitoring.coreos.com \
  servicemonitors.monitoring.coreos.com \
  podmonitors.monitoring.coreos.com \
  alertmanagers.monitoring.coreos.com \
  prometheusrules.monitoring.coreos.com
```

## Development

### Prerequisites

- golang environment
- docker (used for creating container images, etc.)
- kind (optional)

### Testing

> Ensure that you're running tests in the following path:
> `$GOPATH/src/github.com/prometheus-operator/prometheus-operator` as tests expect paths to
> match. If you're working from a fork, just add the forked repo as a remote and
> pull against your local prometheus-operator checkout before running tests.

#### Running *unit tests*:

`make test-unit`

#### Running *end-to-end* tests on local kind cluster:

1. `kind create cluster --image=kindest/node:<latest>`. e.g `v1.16.2` version.
2. `export KUBECONFIG="$(kind get kubeconfig-path --name="kind")"`
3. `make image` - build Prometheus Operator  docker image locally.
4. `for n in "operator" "config-reloader"; do kind load docker-image "quay.io/prometheus-operator/prometheus-$n:$(git rev-parse --short HEAD)"; done` - publish
built locally images to be accessible inside kind.
5. `make test-e2e`

#### Running *end-to-end* tests on local minikube cluster:

1. `minikube start --kubernetes-version=v1.10.0 --memory=4096
   --extra-config=apiserver.authorization-mode=RBAC`
2. `eval $(minikube docker-env) && make image` - build Prometheus Operator
   docker image on minikube's docker
3. `make test-e2e`

## Contributing

Many files (documentation, manifests, ...) in this repository are
auto-generated. E.g. `bundle.yaml` originates from the _Jsonnet_ files in
`/jsonnet/prometheus-operator`. Before proposing a pull request:

1. Commit your changes.
2. Run `make generate`.
3. Commit the generated changes.


## Security

If you find a security vulnerability related to the Prometheus Operator, please
do not report it by opening a GitHub issue, but instead please send an e-mail to
the maintainers of the project found in the [MAINTAINERS.md](MAINTAINERS.md) file.

[operator-vs-kube]: https://github.com/prometheus-operator/prometheus-operator/issues/2510#issuecomment-476692399

## Troubleshooting

Check the [troubleshooting documentation](Documentation/troubleshooting.md) for 
common issues and frequently asked questions (FAQ).

## Acknowledgements

prometheus-operator organization logo was created and contributed by [Bianca Cheng Costanzo](https://github.com/bia).
