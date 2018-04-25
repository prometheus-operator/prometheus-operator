# kube-prometheus

> Note that everything in the `contrib/kube-prometheus/` directory is experimental and may change significantly at any time.

This repository collects Kubernetes manifests, [Grafana](http://grafana.com/) dashboards, and [Prometheus rules](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/) combined with documentation and scripts to provide easy to operate end-to-end Kubernetes cluster monitoring with [Prometheus](https://prometheus.io/) using the Prometheus Operator.

The content of this project is written in [jsonnet](http://jsonnet.org/). This project could both be described as a package as well as a library.

Components included in this package:

* The [Prometheus Operator](https://github.com/coreos/prometheus-operator)
* Highly available [Prometheus](https://prometheus.io/)
* Highly available [Alertmanager](https://github.com/prometheus/alertmanager)
* [Prometheus node-exporter](https://github.com/prometheus/node_exporter)
* [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics)
* [Grafana](https://grafana.com/)

This stack is meant for cluster monitoring, so it is pre-configured to collect metrics from all Kubernetes components. In addition to that it delivers a default set of dashboards and alerting rules. Many of the useful dashboards and alerts come from the [kubernetes-mixin project](https://github.com/kubernetes-monitoring/kubernetes-mixin), similar to this project it provides composable jsonnet as a library for users to customize to their needs.

## Prerequisites

You will need a Kubernetes cluster, that's it! By default it is assumed, that the kubelet uses token authN and authZ, as otherwise Prometheus needs a client certificate, which gives it full access to the kubelet, rather than just the metrics. Token authN and authZ allows more fine grained and easier access control.

### minikube

In order to just try out this stack, start minikube with the following command:

```
$ minikube delete && minikube start --kubernetes-version=v1.10.1 --memory=4096 --bootstrapper=kubeadm --extra-config=kubelet.authentication-token-webhook=true --extra-config=kubelet.authorization-mode=Webhook --extra-config=scheduler.address=0.0.0.0 --extra-config=controller-manager.address=0.0.0.0
```

## Quickstart

Although this project is intended to be used as a library, a compiled version of the Kubernetes manifests generated with this library is checked into this repository in order to try the content our quickly.

Simply create the stack:

```
$ kubectl create -f manifests/
```

## Usage

The content of this project consists of a set of [jsonnet](http://jsonnet.org/) files making up a library to be consumed.

Install this library in your own project with [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler#install):

```
$ mkdir my-kube-prometheus; cd my-kube-prometheus
$ jb init
$ jb install github.com/coreos/prometheus-operator/contrib/kube-prometheus/jsonnet/kube-prometheus
```

> `jb` can be installed with `go get github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb`

You may wish to not use ksonnet and simply render the generated manifests to files on disk, this can be done with:

[embedmd]:# (hack/scripts/kube-prometheus-base.jsonnet)
```jsonnet
local kp = (import "kube-prometheus/kube-prometheus.libsonnet") + {
    _config+:: {
        namespace: "monitoring",
    }
};

{["0prometheus-operator-"+name+".yaml"]: std.manifestYamlDoc(kp.prometheusOperator[name]) for name in std.objectFields(kp.prometheusOperator)} +
{["node-exporter-"+name+".yaml"]: std.manifestYamlDoc(kp.nodeExporter[name]) for name in std.objectFields(kp.nodeExporter)} +
{["kube-state-metrics-"+name+".yaml"]: std.manifestYamlDoc(kp.kubeStateMetrics[name]) for name in std.objectFields(kp.kubeStateMetrics)} +
{["alertmanager-"+name+".yaml"]: std.manifestYamlDoc(kp.alertmanager[name]) for name in std.objectFields(kp.alertmanager)} +
{["prometheus-"+name+".yaml"]: std.manifestYamlDoc(kp.prometheus[name]) for name in std.objectFields(kp.prometheus)} +
{["grafana-"+name+".yaml"]: std.manifestYamlDoc(kp.grafana[name]) for name in std.objectFields(kp.grafana)}
```

This renders all manifests in a json structure of `{filename: manifest-content}`. To split this into files on disk use:

> Note you need `jsonnet`, `jq`, `sed`, `tr` and `gojsonyaml` (`go get github.com/brancz/gojsontoyaml`) installed.

```bash
jsonnet -J vendor example.jsonnet > tmp.json

files=$(jq -r 'keys[]' tmp.json)

for file in ${files}; do
	# prepare directory
    dir=$(dirname "${file}")
    path="${dir}"
    mkdir -p ${path}

	# covert file name to snake case with dashes
    fullfile=$(echo ${file} | sed -r 's/([a-z0-9])([A-Z])/\1-\L\2/g' | tr '[:upper:]' '[:lower:]')

	# write each value to the path in key; convert multiple times to prettify yaml
    jq -r ".[\"${file}\"]" tmp.json | gojsontoyaml -yamltojson | gojsontoyaml > "${fullfile}"
done

rm tmp.json
```

## Configuration

A hidden `_config` field is located at the top level of the object this library provides. These are the available fields with their respective default values:

```
{
	_config+:: {
        namespace: "default",

        versions+:: {
            alertmanager: "v0.14.0",
            nodeExporter: "v0.15.2",
            kubeStateMetrics: "v1.3.0",
            kubeRbacProxy: "v0.3.0",
            addonResizer: "1.0",
            prometheusOperator: "v0.18.1",
            prometheus: "v2.2.1",
        },

        imageRepos+:: {
            prometheus: "quay.io/prometheus/prometheus",
            alertmanager: "quay.io/prometheus/alertmanager",
            kubeStateMetrics: "quay.io/coreos/kube-state-metrics",
            kubeRbacProxy: "quay.io/coreos/kube-rbac-proxy",
            addonResizer: "quay.io/coreos/addon-resizer",
            nodeExporter: "quay.io/prometheus/node-exporter",
            prometheusOperator: "quay.io/coreos/prometheus-operator",
        },

        prometheus+:: {
            replicas: 2,
            rules: {},
        },

        alertmanager+:: {
            config: alertmanagerConfig,
            replicas: 3,
        },
	},
}
```

## Customization

Jsonnet is a turing complete language, any logic can be reflected in it. It also has powerful merge functionalities, allowing sophisticated customizations of any kind simply by merging it into the object the library provides.

A common example is that not all Kubernetes clusters are created exactly the same way, meaning the configuration to monitor them may be slightly different. For [kubeadm]() and [bootkube]() clusters there are mixins available to easily configure these:

kubeadm:
[embedmd]:# (examples/kubeadm.jsonnet)

bootkube:
[embedmd]:# (examples/bootkube.jsonnet)

Another mixin that may be useful for exploring the stack is to expose the UIs of Prometheus, Alertmanager and Grafana on NodePorts:

[embedmd]:# (examples/node-ports.jsonnet)

For example the name of the `Prometheus` object provided by this library can be overridden:

[embedmd]:# (examples/prometheus-name-override.jsonnet)
```jsonnet
((import "kube-prometheus/kube-prometheus.libsonnet") + {
	prometheus+: {
		prometheus+: {
			metadata+: {
				name: "my-name",
			}
		}
	}
}).prometheus.prometheus
```

Standard Kubernetes manifests are all written using [ksonnet-lib](https://github.com/ksonnet/ksonnet-lib/), so they can be modified with the mixins supplied by ksonnet-lib. For example to override the namespace of the node-exporter DaemonSet:

[embedmd]:# (examples/ksonnet-example.jsonnet)
```jsonnet
local k = import "ksonnet/ksonnet.beta.3/k.libsonnet";
local daemonset = k.apps.v1beta2.daemonSet;

((import "kube-prometheus/kube-prometheus.libsonnet") + {
	nodeExporter+: {
		daemonset+:
            daemonset.mixin.metadata.withNamespace("my-custom-namespace") +
    }
}).nodeExporter.daemonset
```
