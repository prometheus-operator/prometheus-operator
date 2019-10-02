# Thanos and the Prometheus Operator

_Note: This guide is valid for Prometheus Operator v0.28+ and Thanos v0.2+ and above._

[Thanos](https://github.com/thanos-io/thanos/) is a set of components
that can be composed into a highly available
metric system with unlimited storage capacity, if your Object Storage allows for it.
The Prometheus Operator provides integration for allowing Prometheus to connect to Thanos.

Thanos components include the rulers, compactors, queries and stores, which Thanos needs to
be fully functional, and should be deployed independently of the Prometheus
Operator and its Thanos configuration. The
[kube-thanos](https://github.com/metalmatze/kube-thanos/) project has some experimental
starting points as well as the [thanos
project](https://github.com/thanos-io/thanos/blob/master/tutorials/kubernetes-demo/manifests).

In short, for the Thanos integration using the Prometheus Operator to work
correctly you will need to have these extra components installed and
configured.

Now please take the time to look at the Thanos README and read through the documentation before continuing with the Prometheus Operator integration.

https://github.com/thanos-io/thanos 
https://github.com/thanos-io/thanos/blob/master/docs/getting-started.md

## Prometheus Operator

### Configuring Thanos Object Storage

Beginning with Thanos v0.2 the sidecar assumes an existing Kubernetes Secret containing the Thanos configuration.
Inside this secret you configure how to run Thanos with your object storage.

For more information and examples about the configuration itself, take a look at the Thanos documentation:  
https://github.com/thanos-io/thanos/blob/master/docs/storage.md

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
Pod. To enable the sidecar, reference the following examples.

This is the simplest configuration change that needs to be made to your
Prometheus Custom Resource, after creating the secret.

```
...
spec:
  ...
  thanos:
    baseImage: quay.io/thanos/thanos
    version: v0.2.1
    objectStorageConfig:
      key: thanos.yaml
      name: thanos-objstore-config
...
```
Note: If you're using Istio you may need to also set `ListenLocal` on the Thanos spec due to Istio's forwarding of traffic to localhost.

## Thanos and kube-thanos

Deploying the sidecar was the first step towards getting Thanos up and running, but there are more components to be deployed, that complete Thanos.

* Store
* Querier
* Compactor

Again, take a look at the Thanos documentation for more details on these components:  
https://github.com/thanos-io/thanos/blob/master/docs/getting-started.md#store-api

Although kube-thanos project is still in early stage, it has already supported several thanos components. 
For more details, please checkout [kube-thanos](https://github.com/metalmatze/kube-thanos/).
