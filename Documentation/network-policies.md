<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.39.0, Prometheus Operator requires use of Kubernetes v1.16.x and up.
</div>

# Network policies

[Network policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/) allow you easily restrict the ingress traffic between pods using [k8s labels](https://kubernetes.io/docs/user-guide/labels/).
To keep your cluster safer, it's strongly recommended to enable network policies into prometheus namespace.

# Example

This example will close all inbound communication on the namespace monitoring, and allow only necessary traffic.
**This example has only been tested with the calico provider.**

First, follow the instructions to [add Calico to an existing Kubernetes cluster](http://docs.projectcalico.org/v1.5/getting-started/kubernetes/installation/).

Next, use the following configuration to deny all the ingress (inbound) traffic.
```yaml
 apiVersion: networking.k8s.io/v1
 kind: NetworkPolicy
 metadata:
   name: default-deny-all
   namespace: monitoring
 spec:
   podSelector:
     matchLabels:
```
Save the config file as default-deny-all.yaml and apply the configuration to the cluster using

```
kubectl apply -f <path to config file>/default-deny-all.yaml
```

Apply the following network policies to allow the necessary traffic to access ports in the pod:

```
$ kubectl apply -n monitoring -f example/networkpolicies/

networkpolicy "alertmanager-web" configured
networkpolicy "alertmanager-mesh" configured
networkpolicy "grafana" configured
networkpolicy "node-exporter" configured
networkpolicy "prometheus" configured
```

## Explaining the network policies

#### Alertmanager

* Allow inbound tcp dst port 9093 from any source to alertmanager
* Allow inbound tcp & udp dst port 9094 from only alertmanager to alertmanager

[embedmd]:# (../example/networkpolicies/alertmanager.yaml)
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: alertmanager-web
spec:
  ingress:
  - from:
    ports:
    - port: 9093
      protocol: TCP
  podSelector:
    matchLabels:
      alertmanager: main
      app: alertmanager
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: alertmanager-mesh
spec:
  ingress:
  - from:
    - podSelector:
        matchExpressions:
        - key: app
          operator: In
          values:
          - alertmanager
        - key: alertmanager
          operator: In
          values:
          - main
    ports:
    - port: 9094
      protocol: TCP
    - port: 9094
      protocol: UDP
  podSelector:
    matchLabels:
      alertmanager: main
      app: alertmanager

```

#### Grafana

* Allow inbound tcp dst port 3000 from any source to grafana

[embedmd]:# (../example/networkpolicies/grafana.yaml)
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: grafana
spec:
  ingress:
  - ports:
    - port: 3000
      protocol: TCP
  podSelector:
    matchLabels:
      app: grafana
```

#### Prometheus

* Allow inbound tcp dst port 9090 from any source to prometheus

[embedmd]:# (../example/networkpolicies/prometheus.yaml)
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: prometheus
spec:
  ingress:
  - ports:
    - port: 9090
      protocol: TCP
  podSelector:
    matchLabels:
      app: prometheus
      prometheus: k8s
```

#### Node-exporter

* Allow inbound tcp dst port 9100 from only prometheus to node-exporter

[embedmd]:# (../example/networkpolicies/node-exporter.yaml)
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: node-exporter
spec:
  ingress:
  - from:
    - podSelector:
        matchExpressions:
        - key: app
          operator: In
          values:
          - prometheus
        - key: prometheus
          operator: In
          values:
          - k8s
    ports:
    - port: 9100
      protocol: TCP
  podSelector:
    matchLabels:
      app: node-exporter
```

#### Kube-state-metrics

* Allow inbound tcp dst port 8080 from only prometheus to kube-state-metrics

[embedmd]:# (../example/networkpolicies/kube-state-metrics.yaml)
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kube-state-metrics
spec:
  ingress:
  - from:
    - podSelector:
        matchExpressions:
        - key: app
          operator: In
          values:
          - prometheus
        - key: prometheus
          operator: In
          values:
          - k8s
    ports:
    - port: 8080
      protocol: TCP
  podSelector:
    matchLabels:
      app: kube-state-metrics
```
