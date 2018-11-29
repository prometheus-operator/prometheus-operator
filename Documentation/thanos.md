# Thanos and the Prometheus Operator

Thanos is a set of components that can be composed into a highly available
metric system with unlimited storage capacity.  The Prometheus Operator provides
integration for allowing Prometheus to connect to Thanos.

These Thanos components include the queriers and stores, which Thanos needs to
be fully functional, and should be deployed independently of the Prometheus
Operator and its Thanos configuration.  The
[kube-prometheus](contrib/kube-prometheus/) project has some experimental
starting points as well as the [thanos
project](https://github.com/improbable-eng/thanos/tree/master/kube/manifests).
In the future there may also be jsonnet configurations for deploying these
additional components.

In short, for the Thanos integration using the Prometheus Operator to work
correctly you will need to have these extra components installed and
configured.

## Prometheus Operator

The `Prometheus` CRD has support for adding a Thanos sidecar to the Prometheus
Pod.  To enable the sidecar, reference the following examples.  These examples
assume that the Thanos components have been configured to use the
`thanos-peers.monitoring.svc:10900` service for querier peers to connect to,
which is important for getting HA to work with Thanos.

### No s3 storage (assumes Thanos querier has been deployed)

This is the simplest configuration change that needs to be made to your
Prometheus Custom Resource, and is the only configuration needed to
provide HA benefits.

```
...
spec:
  ...
  thanos:
    baseImage: improbable/thanos
    peers: thanos-peers.monitoring.svc:10900
...
```

### S3 compatible object storage (assumes Thanos store has been deployed)

Adding the object storage configuration allows Thanos to effeciently store
metrics long term.

```
...
spec:
  ...
  thanos:
    baseImage: improbable/thanos
    peers: thanos-peers.monitoring.svc:10900
    s3:
      accessKey:
        key: access
        name: prometheus-thanos-auth
      bucket: bucket
      endpoint: ams3.digitaloceanspaces.com
      secretKey:
        key: secret
        name: prometheus-thanos-auth
...
```

Note: The `endpoint` key allows for non AWS based object storage to be used.  In
the above example, we are using a Digital Ocean bucket in place of an S3 bucket.
If you use one of the linked deployments you will need to update it to reflect
the cloud provider you are using.

### Extra

Once the sidecars have been configured you will need to make sure to
update any monitoring frontends, e.g. Grafana to connect to the Thanos sidecar
querier, in place of Prometheus.

There is an optional Thanos
[compactor](https://github.com/improbable-eng/thanos/blob/master/docs/components/compact.md)
component, which allows compaction of object storage.

