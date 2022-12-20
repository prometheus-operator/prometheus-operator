---
weight: 208
toc: true
title: Thanos
menu:
    docs:
        parent: operator
lead: Thanos and the Prometheus Operator.
images: []
draft: false
description: Thanos and the Prometheus Operator.
---

[Thanos](https://github.com/thanos-io/thanos/) is a set of components that can be composed into a highly available,
multi Prometheus metric system with potentially unlimited storage capacity, if your Object Storage allows for it.

Before continuing with Prometheus Operator Thanos integration, it is recommended to read more about Thanos in the official [documentation](https://thanos.io/tip/thanos/getting-started.md/).

## What does the Prometheus Operator help with?

Prometheus Operator can manage:
* the Thanos sidecar component with the `Prometheus` custom resource definition. Deployed within the Prometheus pod, it can hook into the Thanos querying system as well as **optionally** back up your data to object storage.
* Thanos Ruler instances with the `ThanosRuler` custom resource definition.

Other Thanos components such the Querier, the Receiver, the Compactor and the Store Gateway should be deployed independently of the Prometheus Operator and its Thanos configuration. The
[kube-thanos](https://github.com/thanos-io/kube-thanos/) project has some starting points for other Thanos components deployments.

## Prometheus Custom Resource with Thanos Sidecar

The `Prometheus` CRD has support for adding a Thanos sidecar to the Prometheus
Pod. To enable the sidecar, the `thanos` section must be set to a non empty value.
For example, the simplest configuration is to just set a valid thanos container image url.

```yaml
...
spec:
  ...
  thanos:
    image: quay.io/thanos/thanos:v0.28.1
...
```

### Configuring Thanos Object Storage

You can configure the Thanos sidecar to upload TSDB blocks to object storage by providing a Kubernetes `Secret` containing the required configuration.

For more information and examples about the configuration itself, take a look at the [Thanos documentation](https://github.com/thanos-io/thanos/blob/main/docs/storage.md).

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

```sh
kubectl -n monitoring create secret generic thanos-objstore-config --from-file=thanos.yaml=/tmp/thanos-config.yaml
```

Then you can specify this secret inside the Thanos field of the Prometheus spec as mentioned [earlier](#prometheus-custom-resource-with-thanos-sidecar):

```yaml
...
spec:
  ...
  thanos:
    image: quay.io/thanos/thanos:v0.28.1
    objectStorageConfig:
      key: thanos.yaml
      name: thanos-objstore-config
...
```

This will attach Thanos sidecar that will backup all *new blocks* that Prometheus produces every 2 hours to the object storage.

NOTE: This option will also disable the local Prometheus compaction. This means that Thanos compactor is the main singleton component
responsible for compactions on a global, object storage level.

## Thanos Ruler

The [Thanos Ruler](https://github.com/thanos-io/thanos/blob/main/docs/components/rule.md) component allows recording and alerting rules to be processed across
multiple Prometheus instances. A `ThanosRuler` instance requires at least one `queryEndpoint` which points to the location of Thanos Queriers or Prometheus instances. The `queryEndpoints` are used to configure the `--query` arguments(s) of the Thanos rulers.

```yaml
...
apiVersion: monitoring.coreos.com/v1
kind: ThanosRuler
metadata:
  name: thanos-ruler-demo
  labels:
    example: thanos-ruler
  namespace: monitoring
spec:
  image: quay.io/thanos/thanos:v0.28.1
  ruleSelector:
    matchLabels:
      role: my-thanos-rules
  queryEndpoints:
    - dnssrv+_http._tcp.my-thanos-querier.monitoring.svc.cluster.local
```

The recording and alerting rules used by a `ThanosRuler` component, are configured using the same `PrometheusRule` objects which are used by Prometheus. In the given example, the rules contained in any `PrometheusRule` object which match the label `role=my-thanos-rules` will be loaded by the Thanos Ruler pods.

## Other Thanos Components

Deploying the sidecar was the first step towards getting Thanos up and running, but there are more components to be deployed to get a complete Thanos setup.

- [Querier](https://thanos.io/tip/components/query.md/)

Additionally, when object storage backup is desired:

- [Store](https://thanos.io/tip/components/store.md/)
- [Compactor](https://thanos.io/tip/components/compact.md/)

kube-thanos project has already supported several thanos components.
For more details, please checkout [kube-thanos](https://github.com/thanos-io/kube-thanos/).
