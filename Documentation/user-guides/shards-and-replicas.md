# Shards and Replicas

If a single Prometheus can't hold the current target's metrics, one can reshard targets onto multiple Prometheus servers.

Shards use the Prometheus `modulus` configuration which takes the hash of the source label values in order to split scrape
targets based on the number of shards. Prometheus operator will create number of `shards` multiplied by `replicas` pods.

Note that scaling down shards will not reshard data onto remaining instances. It must be manually moved. Increasing
shards will not reshard data either. It will continue to be available from the same instances.
To query globally, use the Thanos sidecar and Thanos querier. Alternatively, remote
write to a central location. Sharding is done on the contents of the `__address__` target meta-label.

## Example

View the complete [Shards manifests](../../example/shards).

The following manifest creates a Prometheus server with two replicas:

```yaml
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
  serviceMonitorSelector:
    matchLabels:
      team: frontend
```

This can be verified with the following command:

```bash
> kubectl get pods -n <namespace>
```

The output is similar to this:

```bash
prometheus-prometheus-0                2/2     Running   1          10s
prometheus-prometheus-1                1/2     Running   1          10s
```

Deploy the example application and monitor it:

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
        image: quay.io/brancz/prometheus-example-app:v0.5.0
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

We find the prometheus server scrapes three targets.

### Reshard Targets and Expand Prometheus

Expand prometheus to two shards as shown below:

```yaml mdox-exec="cat example/shards/prometheus.yaml"
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  labels:
    prometheus: shards
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

This can be verified with the following command:

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

Explore one of the monitoring Prometheus instances added for sharding:

```bash
> kubectl port-forward prometheus-prometheus-shard-1-0 9091:9090
```

We find two targets are being scraped. The original Prometheus instance scrapes one target.

To query globally, we must use the Thanos sidecar, since the original data in Prometheus will not be rebalanced.
