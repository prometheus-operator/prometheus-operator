# kube-prometheus

This is the helm chart equivalent of `contrib/kube-prometheus`.

# Prerequisites

Requires helm >= 2.5.0

# Installing on GKE/EKS/AKS

Since the controlplane is managed in these solutions, make sure you tell prometheus to not monitor the scheduler or controller-manager:

```
deployKubeScheduler: False
deployKubeControllerManager: False
```

Then install your custom values (for example):

```
helm upgrade --install kube-prometheus coreos/kube-prometheus --namespace monitoring -f /root/.helm/kube-prometheus-values.yml"
```
