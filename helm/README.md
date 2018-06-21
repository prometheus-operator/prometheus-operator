# TL;DR

```
helm repo add coreos https://s3-eu-west-1.amazonaws.com/coreos-charts/stable/
helm install coreos/prometheus-operator --name prometheus-operator --namespace monitoring
helm install coreos/kube-prometheus --name kube-prometheus --namespace monitoring
````

# How to contribue?

1. Fork the project
2. Make	 the changes in the helm charts
3. Bump the version in Chart.yaml for each modified chart
4. Update [kube-prometheus/requirements.yaml](kube-prometheus/requirements.yaml) file with the dependencies
5. Bump the [kube-prometheus/Chart.yaml](kube-prometheus/Chart.yaml)
6. [Test locally](#how-to-test)
7. Push the changes

# How to test?


```
# From top directory i.e. prometheus-operator
helm install helm/prometheus-operator --name prometheus-operator --namespace monitoring
mkdir -p helm/kube-prometheus/charts
helm package -d helm/kube-prometheus/charts helm/alertmanager helm/grafana helm/prometheus  helm/exporter-kube-dns \
helm/exporter-kube-scheduler helm/exporter-kubelets helm/exporter-node helm/exporter-kube-controller-manager \
helm/exporter-kube-etcd helm/exporter-kube-state helm/exporter-coredns helm/exporter-kubernetes
helm install helm/kube-prometheus --name kube-prometheus --namespace monitoring

```
