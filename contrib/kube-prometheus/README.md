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

## Table of contents

* [Prerequisites](#prerequisites)
    * [minikube](#minikube)
* [Quickstart](#quickstart)
* [Usage](#usage)
    * [Installing](#installing)
    * [Compiling](#compiling)
    * [Containerized Installing and Compiling](#containerized-installing-and-compiling)
* [Configuration](#configuration)
* [Customization](#customization)
    * [Cluster Creation Tools](#cluster-creation-tools)
    * [NodePorts](#nodeports)
    * [Prometheus Object Name](#prometheus-object-name)
    * [node-exporter DaemonSet namespace](#node-exporter-daemonset-namespace)
    * [Alertmanager configuration](#alertmanager-configuration)
    * [Static etcd configuration](#static-etcd-configuration)
    * [Customizing Prometheus alerting/recording rules and Grafana dashboards](#customizing-prometheus-alertingrecording-rules-and-grafana-dashboards)
    * [Exposing Prometheus/Alermanager/Grafana via Ingress](#exposing-prometheusalermanagergrafana-via-ingress)
* [Minikube Example](#minikube-example)
* [Troubleshooting](#troubleshooting)
    * [Error retrieving kubelet metrics](#error-retrieving-kubelet-metrics)
    * [kube-state-metrics resource usage](#kube-state-metrics-resource-usage)

## Prerequisites

You will need a Kubernetes cluster, that's it! By default it is assumed, that the kubelet uses token authN and authZ, as otherwise Prometheus needs a client certificate, which gives it full access to the kubelet, rather than just the metrics. Token authN and authZ allows more fine grained and easier access control.

This means the kubelet configuration must contain these flags:

* `--authentication-token-webhook=true` This flag enables, that a `ServiceAccount` token can be used to authenticate against the kubelet(s).
* `--authorization-mode=Webhook` This flag enables, that the kubelet will perform an RBAC request with the API to determine, whether the requesting entity (Prometheus in this case) is allow to access a resource, in specific for this project the `/metrics` endpoint.

### minikube

In order to just try out this stack, start minikube with the following command:

```
$ minikube delete && minikube start --kubernetes-version=v1.10.1 --memory=4096 --bootstrapper=kubeadm --extra-config=kubelet.authentication-token-webhook=true --extra-config=kubelet.authorization-mode=Webhook --extra-config=scheduler.address=0.0.0.0 --extra-config=controller-manager.address=0.0.0.0
```

## Quickstart

This project is intended to be used as a library (i.e. the intent is not for you to create your own customized copy of this repository).

Though a compiled version of the Kubernetes manifests generated with this library is checked into this repository in order to try the content out quickly.
 * Simply create the stack:
```
$ kubectl create -f manifests/ || true
$ kubectl create -f manifests/ 2>/dev/null || true  # This command sometimes may need to be done twice
```
 * And to teardown the stack:
```
$ kubectl delete -f manifests/ || true
```

## Usage

### Installing

The content of this project consists of a set of [jsonnet](http://jsonnet.org/) files making up a library to be consumed.

Install this library in your own project with [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler#install):
```
$ mkdir my-kube-prometheus; cd my-kube-prometheus
$ jb init  # Creates the initial/empty `jsonnetfile.json`
$ jb install github.com/coreos/prometheus-operator/contrib/kube-prometheus/jsonnet/kube-prometheus  # Creates `vendor/` & `jsonnetfile.lock.json`, and fills in `jsonnetfile.json`
```

> `jb` can be installed with `go get github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb`

> An e.g. of how to install a given version of this library: `jb install github.com/coreos/prometheus-operator/contrib/kube-prometheus/jsonnet/kube-prometheus/@v0.22.0`

### Compiling

You may wish to not use ksonnet/jsonnet and simply render the generated manifests to files on disk (in which case you still don't need a copy of this entire repository, but rather only a copy of a few select files).
This can be done e.g. by running `./build.sh example.jsonnet`.

Here's [example.jsonnet](example.jsonnet):

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

And here's the [build.sh](build.sh) script (this renders all manifests in a json structure of `{filename: manifest-content}`):

[embedmd]:# (build.sh)
```sh
#!/usr/bin/env bash

# This script uses arg $1 (name of *.jsonnet file to use) to generate the manifests/*.yaml files.

set -e
set -x
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail

# Make sure to start with a clean 'manifests' dir
rm -rf manifests
mkdir manifests

                                               # optional, but we would like to generate yaml, not json
jsonnet -J vendor -m manifests "${1-example.jsonnet}" | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml; rm -f {}' -- {}

```

> Note you need `jsonnet` and `gojsontoyaml` (`go get github.com/brancz/gojsontoyaml`) installed. If you just want json output, not yaml, then you can skip the pipe and everything afterwards.

This script reads each key of the generated json and uses that as the file name, and writes the value of that key to that file.

### Containerized Installing and Compiling

If you don't care to have `jb` or `jsonnet` or `gojsontoyaml` installed, then build the `po-jsonnet` Docker image (this is something you'll need a copy of this repository for). Do the following from this `kube-prometheus` directory:
```
$ make ../../hack/jsonnet-docker-image
```

Then you can do commands such as the following:
```
docker run \
	--rm \
	-v `pwd`:`pwd` \
	--workdir `pwd` \
	po-jsonnet jb init

docker run \
	--rm \
	-v `pwd`:`pwd` \
	--workdir `pwd` \
	po-jsonnet jb install github.com/coreos/prometheus-operator/contrib/kube-prometheus/jsonnet/kube-prometheus

docker run \
	--rm \
	-v `pwd`:`pwd` \
	--workdir `pwd` \
	po-jsonnet ./build.sh example.jsonnet
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

The grafana definition is located in a different project (https://github.com/brancz/kubernetes-grafana), but needed configuration can be customized from the same file. F.e. to allow anonymous access to grafana, add the `_config` section:

```
      grafana+:: {
        config: {
          sections: {
            "auth.anonymous": {enabled: true},
          },
        },
      },
```


## Customization

Jsonnet is a turing complete language, any logic can be reflected in it. It also has powerful merge functionalities, allowing sophisticated customizations of any kind simply by merging it into the object the library provides.

### Cluster Creation Tools

A common example is that not all Kubernetes clusters are created exactly the same way, meaning the configuration to monitor them may be slightly different. For [kubeadm](examples/jsonnet-snippets/kubeadm.jsonnet) and [bootkube](examples/jsonnet-snippets/bootkube.jsonnet) and [kops](examples/jsonnet-snippets/kops.jsonnet) clusters there are mixins available to easily configure these:

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

### NodePorts

Another mixin that may be useful for exploring the stack is to expose the UIs of Prometheus, Alertmanager and Grafana on NodePorts:

[embedmd]:# (examples/jsonnet-snippets/node-ports.jsonnet)
```jsonnet
(import 'kube-prometheus/kube-prometheus.libsonnet') +
(import 'kube-prometheus/kube-prometheus-node-ports.libsonnet')
```

### Prometheus Object Name

To give another customization example, the name of the `Prometheus` object provided by this library can be overridden:

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

### node-exporter DaemonSet namespace

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

### Alertmanager configuration

The Alertmanager configuration is located in the `_config.alertmanager.config` configuration field. In order to set a custom Alertmanager configuration simply set this field.

[embedmd]:# (examples/alertmanager-config.jsonnet)
```jsonnet
((import 'kube-prometheus/kube-prometheus.libsonnet') + {
   _config+:: {
     alertmanager+: {
       config: |||
         global:
           resolve_timeout: 10m
         route:
           group_by: ['job']
           group_wait: 30s
           group_interval: 5m
           repeat_interval: 12h
           receiver: 'null'
           routes:
           - match:
               alertname: DeadMansSwitch
             receiver: 'null'
         receivers:
         - name: 'null'
       |||,
     },
   },
 }).alertmanager.secret
```

In the above example the configuration has been inlined, but can just as well be an external file imported in jsonnet via the `importstr` function.

[embedmd]:# (examples/alertmanager-config-external.jsonnet)
```jsonnet
((import 'kube-prometheus/kube-prometheus.libsonnet') + {
   _config+:: {
     alertmanager+: {
       config: importstr 'alertmanager-config.yaml',
     },
   },
 }).alertmanager.secret
```
### Static etcd configuration

In order to configure a static etcd cluster to scrape there is a simple mixin prepared, so only the IPs and certificate information need to be configured. Simply append the `kube-prometheus/kube-prometheus-static-etcd.libsonnet` mixin to the rest of the configuration, and configure the `ips` to be the IPs to scrape, and the `clientCA`, `clientKey` and `clientCert` to values that are valid to scrape etcd metrics with.

Most likely these certificates are generated somewhere in an infrastructure repository, so using the jsonnet `importstr` function can be useful here. All the sensitive information on the certificates will end up in a Kubernetes Secret.

[embedmd]:# (examples/etcd.jsonnet)
```jsonnet
local kp = (import 'kube-prometheus/kube-prometheus.libsonnet') +
           (import 'kube-prometheus/kube-prometheus-static-etcd.libsonnet') + {
  _config+:: {
    namespace: 'monitoring',

    etcd+:: {
      ips: ['127.0.0.1'],
      clientCA: importstr 'etcd-client-ca.crt',
      clientKey: importstr 'etcd-client.key',
      clientCert: importstr 'etcd-client.crt',
      serverName: 'etcd.my-cluster.local',
    },
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

### Customizing Prometheus alerting/recording rules and Grafana dashboards

See [developing Prometheus rules and Grafana dashboards](docs/developing-prometheus-rules-and-grafana-dashboards.md) guide.

### Exposing Prometheus/Alermanager/Grafana via Ingress

See [exposing Prometheus/Alertmanager/Grafana](docs/exposing-prometheus-alertmanager-grafana-ingress.md) guide.

## Minikube Example

To use an easy to reproduce example, see [minikube.jsonnet](examples/minikube.jsonnet), which uses the minikube setup as demonstrated in [Prerequisites](#prerequisites). Because we would like easy access to our Prometheus, Alertmanager and Grafana UIs, `minikube.jsonnet` exposes the services as NodePort type services.

> Note that NodePort type services is likely not a good idea for your production use case, it is only used for demonstration purposes here.

## Troubleshooting

### Error retrieving kubelet metrics

Should the Prometheus `/targets` page show kubelet targets, but not able to successfully scrape the metrics, then most likely it is a problem with the authentication and authorization setup of the kubelets.

As described in the [Prerequisites](#prerequisites) section, in order to retrieve metrics from the kubelet token authentication and authorization must be enabled. Some Kubernetes setup tools do not enable this by default.

If you are using Google's GKE product, see [docs/GKE-cadvisor-support.md].

#### Authentication problem

The Prometheus `/targets` page will show the kubelet job with the error `403 Unauthorized`, when token authentication is not enabled. Ensure, that the `--authentication-token-webhook=true` flag is enabled on all kubelet configurations.

#### Authorization problem

The Prometheus `/targets` page will show the kubelet job with the error `401 Unauthorized`, when token authorization is not enabled. Ensure that the `--authorization-mode=Webhook` flag is enabled on all kubelet configurations.

### kube-state-metrics resource usage

In some environments, kube-state-metrics may need additional
resources. One driver for more resource needs, is a high number of
namespaces. There may be others.

kube-state-metrics resource allocation is managed by
[addon-resizer](https://github.com/kubernetes/autoscaler/tree/master/addon-resizer/nanny)
You can control it's parameters by setting variables in the
config. They default to:

``` jsonnet
    kubeStateMetrics+:: {
      baseCPU: '100m',
      cpuPerNode: '2m',
      baseMemory: '150Mi',
      memoryPerNode: '30Mi',
    }
```
