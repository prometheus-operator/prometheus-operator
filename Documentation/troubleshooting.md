# FAQ / Troubleshooting

### RBAC on Google Container Engine (GKE)

When you try to create `ClusterRole` (`kube-state-metrics`, `prometheus` `prometheus-operator`, etc.) on GKE Kubernetes cluster running 1.6 version, you will probably run into permission errors:

```
<....>
Error from server (Forbidden): error when creating 
"manifests/prometheus-operator/prometheus-operator-cluster-role.yaml": 
clusterroles.rbac.authorization.k8s.io "prometheus-operator" is forbidden: attempt to grant extra privileges:
<....>
````

This is due to the way Container Engine checks permissions. From [Google Container Engine docs](https://cloud.google.com/container-engine/docs/role-based-access-control):

> Because of the way Container Engine checks permissions when you create a Role or ClusterRole, you must first create a RoleBinding that grants you all of the permissions included in the role you want to create.
> An example workaround is to create a RoleBinding that gives your Google identity a cluster-admin role before attempting to create additional Role or ClusterRole permissions.
> This is a known issue in the Beta release of Role-Based Access Control in Kubernetes and Container Engine version 1.6.

To overcome this, you must grant your current Google identity `cluster-admin` Role:

```console
# get current google identity
$ gcloud info | grep Account
Account: [myname@example.org]

# grant cluster-admin to your current identity
$ kubectl create clusterrolebinding myname-cluster-admin-binding --clusterrole=cluster-admin --user=myname@example.org
Clusterrolebinding "myname-cluster-admin-binding" created
```
