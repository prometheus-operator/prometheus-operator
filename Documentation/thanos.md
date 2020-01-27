# Thanos and the Prometheus Operator

_Note: This guide is valid for Prometheus Operator v0.28+ and Thanos v0.2+ and above._

[Thanos](https://github.com/thanos-io/thanos/) is a set of components that can be composed into a highly available,
multi Prometheus metric system with potentially unlimited storage capacity, if your Object Storage allows for it.

Before continuing with Prometheus Operator Thanos integration, it is recommended to read more about Thanos in the [documentation](https://thanos.io/getting-started.md/).

## What Prometheus Operator helps with?

Prometheus Operator operates `Prometheus` and optionally `ThanosRuler` components.
Other Thanos components, such as the querier and store gateway, must be configured
separately.  The Thanos system integrates with Prometheus by adding a Thanos
sidecar to each Prometheus instance.  The Thanos sidecar can be configured directly in the `Prometheus` CRD. This Sidecar can hook into the Thanos querying system as well as **optionally** back up your data in object storage.

Each component other than the sidecar and `ThanosRuler` is deployed independently of the Prometheus Operator and its Thanos configuration. The
[kube-thanos](https://github.com/thanos-io/kube-thanos/) project has some starting points for other Thanos components deployments.

In short, for the Thanos integration using the Prometheus Operator to work correctly you will need to have these extra components installed and configured.

## Prometheus Custom Resource with Thanos Sidecar

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
    version: v0.8.1
...
```

Note: If you're using Istio you may need to also set `ListenLocal` on the Thanos spec due to Istio's forwarding of traffic to localhost.

## Configuring Thanos Object Storage

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

And then you can specify this secret inside Thanos part of the Prometheus CRD we mentioned [earlier](#prometheus-custom-resource-with-thanos-sidecar):

```
...
spec:
  ...
  thanos:
    baseImage: quay.io/thanos/thanos
    version: v0.8.1
    objectStorageConfig:
      key: thanos.yaml
      name: thanos-objstore-config
...
```

This will attach Thanos sidecar that will backup all _new blocks_ that Prometheus produces every 2 hours to the object storage.

NOTE: This option will also disable local Prometheus compaction. This means that Thanos compactor is the main singleton component
responsible for compactions on a global, object storage level.

## Thanos Ruler

The [Thanos Ruler](https://github.com/thanos-io/thanos/blob/master/docs/components/rule.md) component allows recording and alerting rules to be processed across
multiple Promtheus instances.  A `ThanosRuler` instance requires at least one `queryEndpoint` which points to the location of Thanos Queriers or Prometheus instances.  The `queryEndpoints` are used to configure the `--query` arguments(s) of the Thanos runtime.

```
...
apiVersion: monitoring.coreos.com/v1
kind: ThanosRuler
metadata:
  name: thanos-ruler-demo
  labels:
    example: thanos-ruler
  namespace: monitoring
spec:
  image: quay.io/thanos/thanos
  ruleSelector:
    matchLabels:
      role: my-thanos-rules
  queryEndpoints:
    - dnssrv+_http._tcp.my-thanos-querier.monitoring.svc.cluster.local
```

The recording and alerting rules used by a `ThanosRuler` component, are configured using the same `PrometheusRule` objects which are used by Prometheus.  In the given example, the rules contained in any `PrometheusRule` object which match the label `role=my-thanos-rules` will be added to the Thanos Ruler POD.


## Other Thanos Components

Deploying the sidecar was the first step towards getting Thanos up and running, but there are more components to be deployed, that complete Thanos:

- [Querier](https://thanos.io/components/query.md/)

Additionally, when object storage backup is desired:

- [Store](https://thanos.io/components/store.md/)
- [Compactor](https://thanos.io/components/compact.md/)

Again, take a look at the Thanos documentation for more details on these components: https://thanos.io/quick-tutorial.md

kube-thanos project has already supported several thanos components.
For more details, please checkout [kube-thanos](https://github.com/thanos-io/kube-thanos/).
