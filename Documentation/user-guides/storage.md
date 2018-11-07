<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.12.0, Prometheus Operator requires use of Kubernetes v1.7.x and up.
</div>

# Storage

To maintain data across deployments and version upgrades, the data must be persisted to some volume other than `emptyDir`, allowing it to be reused by Pods after an upgrade.

Kubernetes supports several kinds of storage volumes. The Prometheus Operator works with PersistentVolumeClaims, which support the underlying PersistentVolume to be provisioned when requested.

This document assumes a basic understanding of PersistentVolumes, PersistentVolumeClaims, and their [provisioning][pv-provisioning].

## Storage Provisioning on AWS

Automatic provisioning of storage requires a `StorageClass`.

[embedmd]:# (../../example/storage/storageclass.yaml)
```yaml
apiVersion: storage.k8s.io/v1beta1
kind: StorageClass
metadata:
  name: ssd
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp2
```

> Make sure that AWS as a cloud provider is properly configured with your cluster, or storage provisioning will not work.

For best results, use volumes that have high I/O throughput. These examples use SSD EBS volumes. Read the Kubernetes [Persistent Volumes][persistent-volumes] documentation to adapt this `StorageClass` to your needs.

The `StorageClass` that was created can be specified in the `storage` section in the `Prometheus` resource (note that if you're using [kube-prometheus](../../contrib/kube-prometheus/), then instead of making the following change to your `Prometheus` resource, see the [prometheus-pvc.jsonnet](../../contrib/kube-prometheus/examples/prometheus-pvc.jsonnet) example).

[embedmd]:# (../../example/storage/persisted-prometheus.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: persisted
spec:
  storage:
    volumeClaimTemplate:
      spec:
        storageClassName: ssd
        resources:
          requests:
            storage: 40Gi
```

> The full documentation of the `storage` field can be found in the [API documentation][api-doc].

When creating the Prometheus object, a PersistentVolumeClaim is used for each Pod in the StatefulSet, and the storage should automatically be provisioned, mounted and used.

## Manual storage provisioning

The Prometheus CRD specification allows you to support arbitrary storage through a PersistentVolumeClaim.

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

To manually provision volumes (as of Kubernetes 1.6.0), you may need to disable the default StorageClass that is automatically created for certain Cloud Providers. Default StorageClasses are pre-installed on Azure, AWS, GCE, OpenStack, and vSphere.

The default StorageClass behavior will override manual storage provisioning, preventing PersistentVolumeClaims from automatically binding to manually created PersistentVolumes.

To override this behavior, you must explicitly create the same resource, but set it to *not* be default. (See the [changelog][volumes-changelog] for more information.)

For example, to disable default StorageClasses on a Google Container Engine cluster, create the following StorageClass:

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


[volumes-changelog]: https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG.md#volumes
[api-doc]: ../api.md#storagespec
[pv-provisioning]: https://kubernetes.io/docs/user-guide/persistent-volumes/#provisioning
[persistent-volumes]: https://kubernetes.io/docs/user-guide/persistent-volumes/#aws
