# Storage

To keep data cross deployments and version upgrades the data must be persisted to some volume other than `emptyDir`, to be able to reuse it by `Pod`s after an upgrade.

There are various kinds of volumes supported by Kubernetes. The Prometheus Operator works with `PersistentVolumeClaim`s. `PersistentVolumeClaims` are especially useful, because they support the underlying `PersistentVolume` to be provisioned when requested.

This document assumes you have a basic understanding of `PersisentVolume`s, `PersisentVolumeClaim`s, and their [provisioning](https://kubernetes.io/docs/user-guide/persistent-volumes/#provisioning).


## Storage Provisioning on AWS

For automatic provisioning of storage a `StorageClass` is required.

```yaml
apiVersion: storage.k8s.io/v1beta1
kind: StorageClass
metadata:
  name: ssd
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp2
```

> Make sure that AWS as a cloud provider is properly configured with your cluster, as otherwise storage provisioning will not work.

It is recommended to use volumes that have high I/O throughput therefore we're using SSD EBS volumes here. Make sure to read the [documentation](https://kubernetes.io/docs/user-guide/persistent-volumes/#aws) to adapt this `StorageClass` to your needs.

The `StorageClass` that was created can be specified in the `storage` section in the `Prometheus` resource.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: persisted
spec:
  replicas: 1
  resources:
  storage:
    volumeClaimTemplate:
      metadata:
        annotations:
          annotation1: foo
      spec:
        resources:
          requests:
            storage: 1Gi
```

> The full documentation of the `storage` field can be found in the [spec documentation](../api.md#storagespec).

When now creating the `Prometheus` object a `PersistentVolumeClaim` is used for each `Pod` in the `StatefulSet` and the storage should automatically be provisioned, mounted and used.


## Manual storage provisioning

The Prometheus TPR specification allows you to support arbitrary storage, via a PersistentVolumeClaim.

The easiest way to use a volume that cannot be automatically provisioned (for whatever reason) is to use a label selector alongside a manually created PersistentVolume.

For example, using an NFS volume might be accomplished with the following specifications:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: my-example-prometheus-name
  labels:
    prometheus: example
spec:
  ...
  storage:
    volumeClaimTemplate:
      spec:
        selector:
          matchLabels:
            app: my-example-prometheus
        resources:
          requests:
            storage: 50Gi

---

apiVersion: v1
kind: PersistentVolume
metadata:
  name: my-pv-name
  labels:
    app: my-example-prometheus
spec:
  capacity:
    storage: 50Gi
  accessModes:
  - ReadWriteOnce # required
  nfs:
    server: myServer
    path: "/path/to/prom/db"
```

### Disabling Default StorageClasses

In order to manually provoision volumes, as of Kubernetes 1.6.0, you may need to disable the default `StorageClass` that is automatically created for certain Cloud Providers. Default StorageClasses are pre-installed on Azure, AWS, GCE, OpenStack, and vSphere.

The default `StorageClass`s behavior will override manual storage provisioning, causing `PerisistentVolumeClaim`s not to bind manually created `PersistentVolume`s automatically.

To override this behavior, you must explicitely create the same resource, but set it to *not* be default ([see changelog for details](https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG.md#volumes)).

To accomplish this on a Google Container Engine cluster, create the following `StorageClass`:

```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1beta1
metadata:
  name: standard
  annotations:
    # disable this default storage class by setting this annotation to false.
    storageclass.beta.kubernetes.io/is-default-class: "false"
provisioner: kubernetes.io/gce-pd
parameters:
  type: pd-ssd
  zone: us-east1-d
```
