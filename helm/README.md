# THIS CHART IS DEPRECATED
The chart has been replaced by [stable/prometheus-operator](https://github.com/helm/charts/tree/master/stable/prometheus-operator) and is no longer maintained. Therefore we are currently not accepting any pull requests. Thank you for your understanding.

You can check out the tickets for this change [here](https://github.com/coreos/prometheus-operator/issues/592) and [here](https://github.com/helm/charts/pull/6765)

---

#### How to install this chart

```
# Install helm https://docs.helm.sh/using_helm/ then run:
helm repo add coreos https://s3-eu-west-1.amazonaws.com/coreos-charts/stable/
helm install coreos/prometheus-operator --name prometheus-operator --namespace monitoring
helm install coreos/kube-prometheus --name kube-prometheus --namespace monitoring
```