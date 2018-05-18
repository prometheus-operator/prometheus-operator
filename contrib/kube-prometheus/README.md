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

Although this project is intended to be used as a library, a compiled version of the Kubernetes manifests generated with this library is checked into this repository in order to try the content out quickly.

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

[embedmd]:# (example.jsonnet)
```jsonnet
local kp = (import 'kube-prometheus/kube-prometheus.libsonnet') + {
  _config+:: {
    namespace: 'monitoring',
  },
};

{ ['00namespace-' + name]: kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus) } +
{ ['0prometheus-operator-' + name]: kp.prometheusOperator[name] for name in std.objectFields(kp.prometheusOperator) } +
{ ['node-exporter-' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
{ ['kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
{ ['alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
{ ['prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) } +
{ ['grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) }
```

This renders all manifests in a json structure of `{filename: manifest-content}`.

### Compiling

To compile the above and get each manifest in a separate file on disk use the following script:

[embedmd]:# (build.sh)
```sh
#!/usr/bin/env bash
set -e
set -x
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail

# Make sure to start with a clean 'manifests' dir
rm -rf manifests
mkdir manifests

                                               # optional, but we would like to generate yaml, not json
jsonnet -J vendor -m manifests ${1-example.jsonnet} | xargs -I{} sh -c 'cat $1 | gojsontoyaml > $1.yaml; rm -f $1' -- {}

```

> Note you need `jsonnet` and `gojsonyaml` (`go get github.com/brancz/gojsontoyaml`) installed. If you just want json output, not yaml, then you can skip the pipe and everything afterwards.

This script reads each key of the generated json and uses that as the file name, and writes the value of that key to that file.

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

[embedmd]:# (examples/jsonnet-snippets/kubeadm.jsonnet)
```jsonnet
(import 'kube-prometheus/kube-prometheus.libsonnet') +
(import 'kube-prometheus/kube-prometheus-kubeadm.libsonnet')
```

bootkube:

[embedmd]:# (examples/jsonnet-snippets/bootkube.jsonnet)
```jsonnet
(import 'kube-prometheus/kube-prometheus.libsonnet') +
(import 'kube-prometheus/kube-prometheus-bootkube.libsonnet')
```

kops:

[embedmd]:# (examples/jsonnet-snippets/kops.jsonnet)
```jsonnet
(import 'kube-prometheus/kube-prometheus.libsonnet') +
(import 'kube-prometheus/kube-prometheus-kops.libsonnet')
```

Another mixin that may be useful for exploring the stack is to expose the UIs of Prometheus, Alertmanager and Grafana on NodePorts:

[embedmd]:# (examples/jsonnet-snippets/node-ports.jsonnet)
```jsonnet
(import 'kube-prometheus/kube-prometheus.libsonnet') +
(import 'kube-prometheus/kube-prometheus-node-ports.libsonnet')
```

For example the name of the `Prometheus` object provided by this library can be overridden:

[embedmd]:# (examples/prometheus-name-override.jsonnet)
```jsonnet
((import 'kube-prometheus/kube-prometheus.libsonnet') + {
   prometheus+: {
     prometheus+: {
       metadata+: {
         name: 'my-name',
       },
     },
   },
 }).prometheus.prometheus
```

Standard Kubernetes manifests are all written using [ksonnet-lib](https://github.com/ksonnet/ksonnet-lib/), so they can be modified with the mixins supplied by ksonnet-lib. For example to override the namespace of the node-exporter DaemonSet:

[embedmd]:# (examples/ksonnet-example.jsonnet)
```jsonnet
local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';
local daemonset = k.apps.v1beta2.daemonSet;

((import 'kube-prometheus/kube-prometheus.libsonnet') + {
   nodeExporter+: {
     daemonset+:
       daemonset.mixin.metadata.withNamespace('my-custom-namespace'),
   },
 }).nodeExporter.daemonset
```

### Customizing Prometheus alerting/recording rules and Grafana dashboards

See [developing Prometheus rules and Grafana dashboards](docs/developing-prometheus-rules-and-grafana-dashboards.md) guide.

## Example

To use an easy to reproduce example, let's take the minikube setup as demonstrated in [prerequisites](#Prerequisites). It is a kubeadm cluster (as we use the kubeadm bootstrapper) and because we would like easy access to our Prometheus, Alertmanager and Grafana UI we want the services to be exposed as NodePort type services:

> Note that NodePort type services is likely not a good idea for your production use case, it is only used for demonstration purposes here.

[embedmd]:# (examples/minikube.jsonnet)
```jsonnet
local kp =
  (import 'kube-prometheus/kube-prometheus.libsonnet') +
  (import 'kube-prometheus/kube-prometheus-kubeadm.libsonnet') +
  (import 'kube-prometheus/kube-prometheus-node-ports.libsonnet') +
  {
    _config+:: {
      namespace: 'monitoring',
    },
  };

{ ['00namespace-' + name]: kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus) } +
{ ['0prometheus-operator-' + name]: kp.prometheusOperator[name] for name in std.objectFields(kp.prometheusOperator) } +
{ ['node-exporter-' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
{ ['kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
{ ['alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
{ ['prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) } +
{ ['grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) }
```
