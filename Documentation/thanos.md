# Thanos and the Prometheus Operator

_Note: This guide is valid for Prometheus Operator v0.28+ and Thanos v0.2+ and above._

[Thanos](https://github.com/improbable-eng/thanos/) is a set of components
that can be composed into a highly available
metric system with unlimited storage capacity, if you Object Storage allows for it.
The Prometheus Operator provides integration for allowing Prometheus to connect to Thanos.

These Thanos components include the queriers and stores, which Thanos needs to
be fully functional, and should be deployed independently of the Prometheus
Operator and its Thanos configuration. The
[kube-prometheus](contrib/kube-prometheus/) project has some experimental
starting points as well as the [thanos
project](https://github.com/improbable-eng/thanos/tree/master/kube/manifests).

In short, for the Thanos integration using the Prometheus Operator to work
correctly you will need to have these extra components installed and
configured.

Now please take the time to look at the Thanos README and read through the documentation before continuing with the Prometheus Operator integration.

https://github.com/improbable-eng/thanos/  
https://github.com/improbable-eng/thanos/blob/master/docs/getting_started.md

## Prometheus Operator

### Configuring Thanos Object Storage

Beginning with Thanos v0.2 the sidecar assumes an existing Kubernetes Secret containing the Thanos configuration.
Inside this secret you configure how to run Thanos with your object storage.

For more information and examples about the configuration itself, take a look at the Thanos documentation:  
https://github.com/improbable-eng/thanos/blob/master/docs/storage.md

Once you have written your configuration save it to a file.  
Here's an example:

```yaml
type: s3
config:
  bucket: thanos
  endpoint: ams3.digitaloceanspaces.com
  access_key: XXX
  secret_key: XXX
```

Let's assume you saved this file to `/tmp/thanos-config.yaml`. You can use the following command to create a secret called `thanos-objstore-config` inside your cluster in the `monitoring` namespace.

```
kubectl -n monitoring create secret generic thanos-objstore-config --from-file=thanos.yaml=/tmp/thanos-config.yaml
```

### Prometheus Custom Resource with Thanos Sidecar

The `Prometheus` CRD has support for adding a Thanos sidecar to the Prometheus
Pod. To enable the sidecar, reference the following examples. These examples
assume that the Thanos components have been configured to use the
`thanos-peers.monitoring.svc:10900` service for querier peers to connect to,
which is important for getting high availability to work with Thanos.

This is the simplest configuration change that needs to be made to your
Prometheus Custom Resource, after creating the secret, and is the only configuration needed to
provide high availability benefits.

```
...
spec:
  ...
  thanos:
    baseImage: improbable/thanos
    version: v0.2.1
    peers: thanos-peers.monitoring.svc:10900
      objectStorageConfig:
        key: thanos.yaml
        name: thanos-objstore-config
...
```

## Thanos and kube-prometheus

Deploying the sidecar was the first step towards getting Thanos up and running, but there are more components to be deployed, that complete Thanos.

* Store
* Query
* Compactor

Again, take a look at the Thanos documentation for more details on these componenets:  
https://github.com/improbable-eng/thanos/blob/master/docs/getting_started.md#store-api

kube-prometheus has built-in support for these extra components.
To enabled these, you need to change the [contrib/kube-prometheus/example.jsonnet](https://github.com/coreos/prometheus-operator/blob/master/contrib/kube-prometheus/example.jsonnet)
file slightly:

```diff
diff --git a/contrib/kube-prometheus/example.jsonnet b/contrib/kube-prometheus/example.jsonnet
index fcd2bb01..bcf01c75 100644
--- a/contrib/kube-prometheus/example.jsonnet
+++ b/contrib/kube-prometheus/example.jsonnet
@@ -1,5 +1,6 @@
 local kp =
-  (import 'kube-prometheus/kube-prometheus.libsonnet') + {
+  (import 'kube-prometheus/kube-prometheus.libsonnet') +
+  (import 'kube-prometheus/kube-prometheus-thanos.libsonnet') + {
     _config+:: {
       namespace: 'monitoring',
     },

```

Now you can rebuild the manifests by running `./build.sh` and all necesarry changes will be written to `manifests/`.

`git status -s manifests`:
```
 M manifests/prometheus-prometheus.yaml
?? manifests/prometheus-serviceMonitorThanosPeer.yaml
?? manifests/prometheus-thanosPeerService.yaml
?? manifests/prometheus-thanosQueryDeployment.yaml
?? manifests/prometheus-thanosQueryService.yaml
?? manifests/prometheus-thanosStoreStatefulset.yaml
```

Now you can `kubectl apply -f manifests` like always.
The store will know configured itself with the same `thanos-objstore-config` secret that the sidecar uses.

We also deployed a ServiceMonitor that automatically starts to scrape your Thanos components.
