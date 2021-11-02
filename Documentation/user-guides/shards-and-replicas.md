# Shards and Replicas

If single Prometheus can't hold current targets metrics,user can reshard targets on multiple Prometheus servers. Shards
use Prometheus `modulus` configuration to implement, which take of the hash of the source label values, split scrape
targets based on the number of shards. When a single Prometheus can no longer cater for additional metrics, a user can
reshard targets on multiple Prometheus servers.

Shards use Prometheus `modulus` configuration which take the hash of the source label values in order to split scrape
targets based on the number of shards. Prometheus operator will create number of `shards` multiplied by `replicas` pods.

Note that scaling down shards will not reshard data onto remaining instances, it must be manually moved. Increasing
shards will not reshard data either but it will continue to be available from the same instances. To query globally use
Thanos sidecar and Thanos querier or remote write data to a central location. Sharding is done on the content of
the `__address__` target meta-label. To query globally, use Thanos sidecar and Thanos querier. Alternatively, remote
write to a central location. Sharding is done on the content of the `__address__` target meta-label.

## Example

View the complete [Shards manifests](../../example/shards).

The following manifest creates a Prometheus server with two replicas:

```yaml mdox-exec="cat example/shards/prometheus.yaml"
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  labels:
    prometheus: prometheus
  name: prometheus
  namespace: default
spec:
  serviceAccountName: prometheus
  replicas: 2
  shards: 2
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
prometheus-prometheus-0                2/2     Running   1          10s
prometheus-prometheus-1                1/2     Running   1          10s
```

Deploy example application and monitor it:

```yaml mdox-exec="cat example/shards/example-app-deployment.yaml"
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: example-app
  template:
    metadata:
      labels:
        app: example-app
    spec:
      containers:
      - name: example-app
        image: fabxc/instrumented_app:latest
        ports:
        - name: web
          containerPort: 8080
```

```yaml mdox-exec="cat example/shards/example-app-service.yaml"
kind: Service
apiVersion: v1
metadata:
  name: example-app
  labels:
    app: example-app
spec:
  selector:
    app: example-app
  ports:
  - name: web
    port: 8080
```

```yaml mdox-exec="cat example/shards/example-app-service-monitor.yaml"
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: example-app
  labels:
    team: frontend
spec:
  selector:
    matchLabels:
      app: example-app
  endpoints:
  - port: web
```

Explore one of the monitoring Prometheus instances:

```bash
> kubectl port-forward pod/prometheus-prometheus-0 9090:9090
```

We can find the prometheus server scrape three targets.

### Reshard targets and Expand Prometheus

Expand prometheus to two shards like below:

```yaml mdox-exec="cat example/shards/prometheus.yaml"
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  labels:
    prometheus: prometheus
  name: prometheus
  namespace: default
spec:
  serviceAccountName: prometheus
  replicas: 2
  shards: 2
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
prometheus-prometheus-0                2/2     Running   1          11m
prometheus-prometheus-1                2/2     Running   1          11m
prometheus-prometheus-shard-1-0        2/2     Running   1          12s
prometheus-prometheus-shard-1-1        2/2     Running   1          12s
```

Explore one of expand monitoring Prometheus instances:

```bash
> kubectl port-forward prometheus-prometheus-shard-1-0  9091:9090
```

We find two targets in scraping. The origin Prometheus instance scrapes one target.

To query globally, we must use Thanos sidecar, since the original data in Prometheus will not be rebalanced.
