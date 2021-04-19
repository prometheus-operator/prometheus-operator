# Shards and Replicas

If single prometheus can't hold current targets metrics,user can reshard targets on multiple prometheus servers.
Shards use prometheus `modulus` configuration to implement,which take of the hash of the source label values,split scrape targets based on the number of shards.

Prometheus operator will create  number of `shards` multiplied by `replicas` pods.

Note that scaling down shards will not reshard data onto remaining instances,it must be manually moved. Increasing shards will not reshard data either but it will continue to be available from the same instances. 
To query globally use Thanos sidecar and Thanos querier or remote write data to a central location. Sharding is done on the content of the `__address__` target meta-label.

## Example

The following manifest configure shards replicas field,it will create four pods.
```
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  labels:
    prometheus: prometheus
spec:
  replicas: 2
  shards: 2
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      team: frontend
```

This could by verified by the following command:
  
```bash
> kubectl get pods -n <namespace>
```
  
The output is similar to this:
  
```bash
prometheus-prometheus-0                2/2     Running   1          3m31s
prometheus-prometheus-1                2/2     Running   1          3m31s
prometheus-prometheus-shard-1-0        2/2     Running   1          3m31s
prometheus-prometheus-shard-1-1        2/2     Running   1          3m31s
```
