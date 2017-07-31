# Migration from TPR to CRDs

ThirdPartyResources (TPRs) are now being deprecated in favor of CustomResourceDefinitions (CRDs) and the Prometheus Operator automatically migrates TPRs to CRDs when it can. If it encounters an error during the migration, it will rollback the changes done until that point. To deal with edge cases, like the operator dying in the middle of a rollback, we store the data safely so that it can always be restored manually. This doc outlines the process of migration and the manual recovery.

Migration happens in the following steps:

1. Take all TPR data and store it in a ConfigMap
2. Delete the TPR data
3. Delete the TPR registration
4. Create the CRD
5. Restore data from ConfigMap


### ConfigMap layout
We store the data in 3 configmaps:
1. crd-migrate-alertmanager-1501479108
2. crd-migrate-prometheus-1501479108
3. crd-migrate-servicemonitor-1501479108

Where the number at the end is the timestamp at which the migration ran.

The data is stored as stringified json inside the `backup` field in the config map.
For example:
`kubectl get cm crd-migrate-prometheus-1501479108 -oyaml`
```yaml
apiVersion: v1
data:
  backup: '{"kind":"PrometheusList","items":[{"apiVersion":"monitoring.coreos.com/v1alpha1","kind":"Prometheus","metadata":{"name":"main","namespace":"default","selfLink":"/apis/monitoring.coreos.com/v1alpha1/namespaces/default/prometheuses/main","uid":"7da0c9d5-75b1-11e7-859a-080027cd0f64","resourceVersion":"1997","creationTimestamp":"2017-07-31T05:31:23Z","labels":{"prometheus":"main"}},"spec":{"alerting":{"alertmanagers":[{"name":"alertmanager","namespace":"default","port":"web"}]},"replicas":2,"resources":{"requests":{"memory":"400Mi"}},"serviceMonitorSelector":{"matchExpressions":[{"key":"app","operator":"In","values":["node-exporter","example-app"]}]},"version":"v1.5.2"}}],"metadata":{"selfLink":"/apis/monitoring.coreos.com/v1alpha1/prometheuses","resourceVersion":"2023"},"apiVersion":"monitoring.coreos.com/v1alpha1"}'
kind: ConfigMap
metadata:
  creationTimestamp: 2017-07-31T05:31:47Z
  name: crd-migrate-prometheus-1501479108
  namespace: default
  resourceVersion: "2024"
  selfLink: /api/v1/namespaces/default/configmaps/crd-migrate-prometheus-1501479108
  uid: 8be1e132-75b1-11e7-859a-080027cd0f64
```

### Manual restore
1. Check if crds are created. If yes, delete them via:
  ```
  kubectl delete crd alertmanagers.monitoring.coreos.com
  kubectl delete crd prometheuses.monitoring.coreos.com
  kubectl delete crd servicemonitors.monitoring.coreos.com
  ```
1. Recreate the TPRs:
  `kubectl create -f -` and paste the following TPR config in each command.

  ```
  apiVersion: extensions/v1beta1
  description: Managed Prometheus server
  kind: ThirdPartyResource
  metadata:
    name: prometheus.monitoring.coreos.com
  versions:
  - name: v1alpha1
  ```
```
  apiVersion: extensions/v1beta1
  description: Prometheus monitoring for a service
  kind: ThirdPartyResource
  metadata:
    name: service-monitor.monitoring.coreos.com
  versions:
  - name: v1alpha1
```
```
apiVersion: extensions/v1beta1
description: Managed Alertmanager cluster
kind: ThirdPartyResource
metadata:
  name: alertmanager.monitoring.coreos.com
versions:
- name: v1alpha1
```

1. Restore the data:

  * `kubectl get cm crd-migrate-servicemonitor-1501479108 -ojson | jq '.data.backup | fromjson '  > servicemonitor.json; kubectl create -f servicemonitor.json`

  * `kubectl get cm crd-migrate-prometheus-1501479108 -ojson | jq '.data.backup | fromjson '  > prometheus.json; kubectl create -f prometheus.json`

  * `kubectl get cm crd-migrate-alertmanager-1501479108 -ojson | jq '.data.backup | fromjson '  > alertmanager.json; kubectl create -f alertmanager.json`
