<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.12.0, Prometheus Operator requires use of Kubernetes v1.7.x and up.
</div>

# Network policies

[Network policies](https://kubernetes.io/docs/user-guide/networkpolicies/) allow you easily restrict the ingress traffic between pods using [k8s labels](https://kubernetes.io/docs/user-guide/labels/). 
To keep your cluster safer, it's strongly recommended to enable network policies into prometheus namespace.

# Example

In this example we are closing all the inbound communication on the namespace monitoring and just allowing the necessary traffic.
**This example are only tested with calico provider.**

Follow the steps [here](http://docs.projectcalico.org/v1.5/getting-started/kubernetes/installation/) to install calico, also dont' forget to [enable network policy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) in you k8s cluster.

Once you've done that, you should use the following configuration to deny all the ingress (inbound) traffic.
``` 
 kind: NetworkPolicy
 apiVersion: networking.k8s.io/v1
 metadata:
 name: default-deny-all
 namespace: default
 spec:
 podSelector:
 matchLabels:
```
Save the config file as default-deny-all.yaml and apply the configuration to the cluster using

```kubectl apply -f <path to config file>/default-deny-all.yaml```

In this step you can't reach any port in your pod, so let's apply this network policies examples to allow the necessary traffic.

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
* Allow inbound tcp dst port 6783 from only alertmanager to alertmanager 
 
[embedmd]:# (../example/networkpolicies/alertmanager.yaml)
```yaml
apiVersion: extensions/v1beta1
kind: NetworkPolicy
metadata:
  name: alertmanager-web
spec:
  ingress:
  - from:
    ports:
    - port: 9093
      protocol: tcp
  podSelector:
    matchLabels:
      alertmanager: main
      app: alertmanager
---
apiVersion: extensions/v1beta1
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
    - port: 6783
      protocol: tcp
  podSelector:
    matchLabels:
      alertmanager: main
      app: alertmanager
```

#### Grafana

* Allow inbound tcp dst port 3000 from any source to grafana  

[embedmd]:# (../example/networkpolicies/grafana.yaml)
```yaml
apiVersion: extensions/v1beta1
kind: NetworkPolicy
metadata:
  name: grafana
spec:
  ingress:
  - ports:
    - port: 3000
      protocol: tcp
  podSelector:
    matchLabels:
      app: grafana
```

#### Prometheus

* Allow inbound tcp dst port 9090 from any source to prometheus  

[embedmd]:# (../example/networkpolicies/prometheus.yaml)
```yaml
apiVersion: extensions/v1beta1
kind: NetworkPolicy
metadata:
  name: prometheus
spec:
  ingress:
  - ports:
    - port: 9090
      protocol: tcp
  podSelector:
    matchLabels:
      app: prometheus
      prometheus: k8s
```

#### Node-exporter

* Allow inbound tcp dst port 9100 from only prometheus to node-exporter  

[embedmd]:# (../example/networkpolicies/node-exporter.yaml)
```yaml
apiVersion: extensions/v1beta1
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
      protocol: tcp
  podSelector:
    matchLabels:
      app: node-exporter
```

#### Kube-state-metrics

* Allow inbound tcp dst port 8080 from only prometheus to kube-state-metrics  

[embedmd]:# (../example/networkpolicies/kube-state-metrics.yaml)
```yaml
apiVersion: extensions/v1beta1
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
      protocol: tcp
  podSelector:
    matchLabels:
      app: kube-state-metrics
```
