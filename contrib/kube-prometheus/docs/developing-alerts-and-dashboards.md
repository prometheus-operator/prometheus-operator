# Developing Prometheus Rules and Grafana Dashboards

`kube-prometheus` ships with a set of default [Prometheus rules](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/) and [Grafana](http://grafana.com/) dashboards. At some point one might like to extend them, the purpose of this document is to explain how to do this.

For both the Prometheus rules and the Grafana dashboards there are Kubernetes `ConfigMap`s, that are generated from content in the `assets/` directory.

The source of truth for the alerts and dashboards are the files in the `assets/` directory. The respective files have to be changed there and then the `make generate` make target is executed to re-generate the Kubernetes manifests.

Note: `make generate` should be executed from kube-prometheus base directory.

## Prometheus Rules

The `ConfigMap` that is generated and holds the Prometheus rule files can be found in `manifests/prometheus/prometheus-k8s-rules.yaml`.

It is generated from all the `*.rules.yaml` files in the `assets/prometheus/rules/` directory.

To extend the rules simply add a new `.rules.yaml` file into the `assets/prometheus/rules/` directory and re-generate the manifests. To modify the existing rules, simply edit the respective `.rules.yaml` file and re-generate the manifest.

Then the generated manifest can be applied against a Kubernetes cluster.

## Dashboards

The generated `ConfigMap`s holding the Grafana dashboard definitions can be found in `manifests/grafana/grafana-dashboards.yaml`.

The dashboards themselves get generated from Python scripts: assets/grafana/\*.dashboard.py.
These scripts are loaded by the [grafanalib](https://github.com/aknuds1/grafanalib)
Grafana dashboard generator, which turns them into dashboards.

Bear in mind that we are for now using a fork of grafanalib as we needed to make extensive
changes to it, in order to be able to generate our dashboards. We are hoping to be able to
consolidate our version with the original.

After changing grafanalib scripts in assets/grafana, or adding your own, you'll have to run
`make generate` in the kube-prometheus root directory in order to re-generate the dashboards
manifest. You can deploy the latter with kubectl similar to the following:

```
kubectl -n monitoring apply -f manifests/grafana/grafana-dashboards.yaml
```

This should cause Grafana to re-load its dashboards automatically.
