# Grafana Watcher

The Grafaner Watcher was built for [the kube-prometheus collection](https://github.com/coreos/kube-prometheus) in order to run Grafana in an easily replicable manner without the need to run a complicated database, and rather provision dashboards from configs off of files. It subscribes to filesystem changes in a given directory, reads files matching `*-datasource.json` and `*-dashboard.json` and imports the datasources and dashboards to a given Grafana instance via Grafana's REST API.

## How to use

A Docker container is provided on `quay.io/coreos/grafana-watcher` or can be built by your self. Run the container and make sure Grafana is reachable from Grafana Watcher. Pass source directory and Grafana URL to start command, e.g:

- --grafana-url=http://localhost:3000
- --watch-dir=/var/grafana-dashboards

When running on Kubernetes, make sure to mount a volume containing the desired dashboards and datasources to `watch-dir`, e.g. as ConfigMap, as showed below.
Minimal Kubernetes example configuration:

deployment.yaml
```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: grafana
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: grafana
    spec:
      containers:
      - name: grafana
        image: grafana/grafana:4.4.3
      - name: grafana-watcher
        image: quay.io/coreos/grafana-watcher:v0.0.8
        args:
        - '--watch-dir=/var/grafana-dashboards'
        - '--grafana-url=http://localhost:3000'
        volumeMounts:
        - name: grafana-dashboards
          mountPath: /var/grafana-dashboards
    volumes:
    - name: grafana-dashboards
      configMap:
        name: grafana-dashboards
```

configmap.yaml
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards
data:
  prometheus-datasource.json: |
    {"access":"proxy","basicAuth":true, ... }

  overview-dashboard.json: |
    {"dashboard":{"annotations":{"list":[]}, ... }}
```

When the dashboard JSON is exported via the Grafana web-ui, it has to be wrapped in `{"dashboard": {}}`, variables, marked by `${}`, have to be replaced and the fields `__input` and `__requires` have to be removed, for the dashboard JSON to be parseable by Grafana via the REST API used by the Grafana Watcher.

Alternatively use `make generate` as described in [Developing Alerts and Dashboards](../kube-prometheus/docs/developing-alerts-and-dashboards.md) to create the ConfigMap.
