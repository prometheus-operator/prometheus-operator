---
weight: 209
toc: true
title: Storage
menu:
    docs:
        parent: operator
lead: ""
images: []
draft: false
description: Storage considerations
---

By default, the operator configures Pods to store data on `emptyDir` volumes
which aren't persisted when the Pods are redeployed. To maintain data across
deployments and version upgrades, you can configure persistent storage for
Prometheus, Alertmanager and ThanosRuler resources.

Kubernetes supports several kinds of storage volumes. The Prometheus Operator
works with PersistentVolumeClaims, which support the underlying
PersistentVolume to be provisioned when requested.

This document assumes a basic understanding of PersistentVolumes,
PersistentVolumeClaims, and their
[provisioning](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#provisioning).

## Storage Provisioning on AWS

Automatic provisioning of storage requires a `StorageClass`.

```yaml mdox-exec="cat example/storage/storageclass.yaml"
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: ssd
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp2
```

> Note: Make sure that AWS as a cloud provider is properly configured with your cluster, or storage provisioning will not work.

For best results, use volumes that have high I/O throughput. These examples use
SSD EBS volumes. Read the Kubernetes [Persistent
Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#provisioning)
documentation to adapt this `StorageClass` to your needs.

The `StorageClass` that was created can be specified in the `storage` section
in the `Prometheus` resource (note that if you're using
[kube-prometheus](https://github.com/prometheus-operator/kube-prometheus), then
instead of making the following change to your `Prometheus` resource, see the
[prometheus-pvc.jsonnet](https://github.com/prometheus-operator/kube-prometheus/blob/main/examples/prometheus-pvc.jsonnet)
example).

```yaml mdox-exec="cat example/storage/persisted-prometheus.yaml"
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

> The full documentation of the `storage` field can be found in the [API reference]({{< ref "api" >}}).

When creating the Prometheus object, a PersistentVolumeClaim is used for each
Pod in the StatefulSet, and the storage should automatically be provisioned,
mounted and used.

The same approach should work with other cloud providers (GCP, Azure, ...) and
any Kubernetes storage provider supporting dynamic provisioning.

## Manual storage provisioning

The Prometheus CRD specification allows you to support arbitrary storage
through a PersistentVolumeClaim.

The easiest way to use a volume that cannot be automatically provisioned (for
whatever reason) is to use a label selector alongside a manually created
PersistentVolume.

For example, using an NFS volume might be accomplished with the following
manifests:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: my-example-prometheus-name
  labels:
    prometheus: example
spec:
  replicas: 1
  storage:
    volumeClaimTemplate:
      spec:
        selector:
          matchLabels:
            app.kubernetes.io/name: my-example-prometheus
        resources:
          requests:
            storage: 50Gi
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: my-pv-name
  labels:
    app.kubernetes.io/name: my-example-prometheus
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

To manually provision volumes (as of Kubernetes 1.6.0), you may need to disable
the default StorageClass that is automatically created for certain Cloud
Providers. Default StorageClasses are pre-installed on Azure, AWS, GCE,
OpenStack, and vSphere.

The default StorageClass behavior will override manual storage provisioning,
preventing PersistentVolumeClaims from automatically binding to manually
created PersistentVolumes.

To override this behavior, you must explicitly create the same resource, but
set it to *not* be default (see the
[changelog](https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.6.md#volumes)
for more information.)

For example, to disable default StorageClasses on a Google Container Engine cluster, create the following StorageClass:

```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: standard
  annotations:
    # disable this default storage class by setting this annotation to false.
    storageclass.kubernetes.io/is-default-class: "false"
provisioner: kubernetes.io/gce-pd
parameters:
  type: pd-ssd
  zone: us-east1-d
```

## Resizing volumes

Even if the StorageClass supports resizing, Kubernetes doesn't support (yet)
volume expansion through StatefulSets. This means that when you update the
storage requests in the `spec.storage` field of a custom resource such as
Prometheus, the operator has to delete/recreate the underlying StatefulSet and
the associated PVCs aren't expanded (more details in the [KEP
issue](https://github.com/kubernetes/enhancements/issues/661)).

It is still possible to fix the situation manually.

First check that the storage class allows volume expansion:

```bash
$ kubectl get storageclass -o custom-columns=NAME:.metadata.name,ALLOWVOLUMEEXPANSION:.allowVolumeExpansion
NAME      ALLOWVOLUMEEXPANSION
gp2-csi   true
gp3-csi   true
```

Next, update the `spec.paused` field to `true` (to prevent the operator from recreating the StatefulSet) and update the storage request in the `spec.storage` field of the custom resource. Assuming a Prometheus resource named `example` for which you want to increase the storage size to 10Gi:

```bash
kubectl patch prometheus/example --patch '{"spec": {"paused": true, "storage": {"volumeClaimTemplate": {"spec": {"resources": {"requests": {"storage":"10Gi"}}}}}}}' --type merge
```

Next, patch every PVC with the updated storage request (10Gi in this example):

```bash
for p in $(kubectl get pvc -l operator.prometheus.io/name=example -o jsonpath='{range .items[*]}{.metadata.name} {end}'); do \
  kubectl patch pvc/${p} --patch '{"spec": {"resources": {"requests": {"storage":"10Gi"}}}}'; \
done
```

Next, delete the underlying StatefulSet using the `orphan` deletion strategy:

```bash
kubectl delete statefulset -l operator.prometheus.io/name=example --cascade=orphan
```

Last, change `spec.paused` field of the custom resource back to `false`.

```bash
kubectl patch prometheus/example --patch '{"spec": {"paused": false}}' --type merge
```

The operator should recreate the StatefulSet immediately, there will be no
service disruption thanks to the `orphan` strategy and the volumes mounted in
the Pods should have the updated size.
