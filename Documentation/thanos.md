# Thanos and the Prometheus Operator

_Note: This guide is valid for Prometheus Operator v0.28+ and Thanos v0.2+ and above._

[Thanos](https://github.com/thanos-io/thanos/) is a set of components that can be composed into a highly available, multi
Prometheus metric system with potentially unlimited storage capacity, if your Object Storage allows for it.

## What Prometheus Operator helps with?

Prometheus Operator operates Prometheus, not Thanos. However Thanos system integrates with existing setup by adding 
sidecar to each Prometheus running in the system.
 
Please before continuing with Prometheus Operator Thanos integration, read more about Thanos in the [documentation](https://thanos.io/getting-started.md/).

Prometheus Operator allows your to optionally add Thanos sidecar to Prometheus. Sidecar allows to hook into Thanos
querying system as well as **optionally** back up your data in object storage.

Thanos system includes other components like queriers or rulers. To get the advantage of object storage it also requires compactors and stores.

All beside the sidecar should be deployed independently of the Prometheus Operator and its Thanos configuration. The
[kube-thanos](https://github.com/thanos-io/kube-thanos/) project has some starting points for other Thanos components deployments.

In short, for the Thanos integration using the Prometheus Operator to work correctly you will need to have these extra
components installed and configured.

## Prometheus Operator

Let's walk through the process of adding Thanos sidecar to Prometheus Operator.

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
...
```
Note: If you're using Istio you may need to also set `ListenLocal` on the Thanos spec due to Istio's forwarding of traffic to localhost.

### Optional: Configuring Thanos Object Storage

If you want sidecar to be able to upload blocks to object storage you need to tell Prometheus Operator about it.

In this mode, sidecar assumes an existing Kubernetes Secret containing the Thanos configuration.
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

And then you can specify this secret inside Thanose part of the Prometheus CRD we mentioned [earlier](#prometheus-custom-resource-with-thanos-sidecar):

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

This will attach Thanos sidecar that will backup all *new blocks* that Prometheus produces every 2 hours to the object storage.

NOTE: This option will also disable local Prometheus compaction. This means that Thanos compactor is the main singleton component 
responsible for compactions on a global, object storage level.

## Thanos and kube-thanos

Deploying the sidecar was the first step towards getting Thanos up and running, but there are more components to be deployed, that complete Thanos:

* [Querier](https://thanos.io/components/query.md/)

Additionally, when object storage backup is desired:

* [Store](https://thanos.io/components/store.md/)
* [Compactor](https://thanos.io/components/compact.md/)

Again, take a look at the Thanos documentation for more details on these components: https://thanos.io/quick-tutorial.md

kube-thanos project has already supported several thanos components. 
For more details, please checkout [kube-thanos](https://github.com/thanos-io/kube-thanos/).
