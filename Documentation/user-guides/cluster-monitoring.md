# Cluster Monitoring

This guide is intended to give an introduction to all the parts required to start monitoring a Kubernetes cluster with Prometheus using the Prometheus Operator.

This guide assumes you have a basic understanding of how to use the functionality the Prometheus Operator implements. If you haven't yet, we recommend reading through the [getting started guide](getting-started.md) as well as the [alerting guide](alerting.md).

## Metric Sources

Monitoring a Kubernetes cluster with Prometheus is a natural choice as Kubernetes components themselves are instrumented with Prometheus metrics, therefore those components simply have to be discovered by Prometheus and most of the cluster is monitored.

Metrics that are rather about cluster state than a single component's metrics is exposed by the add-on component [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics).

Additionally, to have an overview of cluster nodes' resources the Prometheus [node_exporter](https://github.com/prometheus/node_exporter) is used. The node_exporter allows monitoring a node's resources: CPU, memory and disk utilization and more.

Once you complete this guide you will monitor the following:

* cluster state via kube-state-metrics
* nodes via the node_exporter
* kubelets
* apiserver
* kube-scheduler
* kube-controller-manager

## Preparing Kubernetes Components

The manifests used here use the [Prometheus Operator](https://github.com/coreos/prometheus-operator), which manages Prometheus servers and their configuration in a cluster. Prometheus discovers targets through `Endpoints` objects, which means all targets that are running as `Pod`s in the Kubernetes cluster are easily monitored. Many Kubernetes components can be [self-hosted](https://coreos.com/blog/self-hosted-kubernetes.html) today. The kubelet, however, is not. Therefore the Prometheus Operator implements a functionality to synchronize the kubelets into an `Endpoints` object. To make use of that functionality the `--kubelet-service` argument must be passed to the Prometheus Operator when running it.

[embedmd]:# (../../contrib/kube-prometheus/manifests/prometheus-operator/prometheus-operator.yaml)
```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    k8s-app: prometheus-operator
  name: prometheus-operator
spec:
  replicas: 1
  template:
    metadata:
      labels:
        k8s-app: prometheus-operator
    spec:
      containers:
      - args:
        - --kubelet-service=kube-system/kubelet
        - --config-reloader-image=quay.io/coreos/configmap-reload:v0.0.1
        image: quay.io/coreos/prometheus-operator:v0.12.0
        name: prometheus-operator
        ports:
        - containerPort: 8080
          name: http
        resources:
          limits:
            cpu: 200m
            memory: 100Mi
          requests:
            cpu: 100m
            memory: 50Mi
      serviceAccountName: prometheus-operator
```

> Make sure that the `ServiceAccount` called `prometheus-operator` exists and if using RBAC, is bound to the correct role. Read more on [RBAC when using the Prometheus Operator](../rbac.md).

Once started it ensures that all internal IPs of the nodes in the cluster are synchronized into the specified `Endpoints` object. In this case the object is called `kubelet` and is located in the `kube-system` namespace.

By default every Kubernetes cluster has a `Service` for easy access to the API server. This is the `Service` called `kubernetes` in the `default` namespace. A `Service` object automatically synchronizes an `Endpoints` object with the targets it selects. Therefore there is nothing, extra to do for Prometheus to be able to discover the API server.

Aside from the kubelet and the API server the other Kubernetes components all run on top of Kubernetes itself. To discover Kubernetes components that run in a Pod, they simply have to be added to a `Service`.

kube-scheduler:

[embedmd]:# (../../contrib/kube-prometheus/manifests/k8s/self-hosted/kube-scheduler.yaml)
```yaml
apiVersion: v1
kind: Service
metadata:
  namespace: kube-system
  name: kube-scheduler-prometheus-discovery
  labels:
    k8s-app: kube-scheduler
spec:
  selector:
    k8s-app: kube-scheduler
  type: ClusterIP
  clusterIP: None
  ports:
  - name: http-metrics
    port: 10251
    targetPort: 10251
    protocol: TCP
```

kube-controller-manager:

[embedmd]:# (../../contrib/kube-prometheus/manifests/k8s/self-hosted/kube-controller-manager.yaml)
```yaml
apiVersion: v1
kind: Service
metadata:
  namespace: kube-system
  name: kube-controller-manager-prometheus-discovery
  labels:
    k8s-app: kube-controller-manager
spec:
  selector:
    k8s-app: kube-controller-manager
  type: ClusterIP
  clusterIP: None
  ports:
  - name: http-metrics
    port: 10252
    targetPort: 10252
    protocol: TCP
```

## Exporters

Unrelated to Kubernetes itself, but still important is to gather various metrics about the actual nodes. Typical metrics are CPU, memory, disk and network utilization, all of these metrics can be gathered using the node_exporter.

[embedmd]:# (../../contrib/kube-prometheus/manifests/node-exporter/node-exporter-daemonset.yaml)
```yaml
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: node-exporter
spec:
  template:
    metadata:
      labels:
        app: node-exporter
      name: node-exporter
    spec:
      hostNetwork: true
      hostPID: true
      containers:
      - image:  quay.io/prometheus/node-exporter:v0.14.0
        args:
        - "-collector.procfs=/host/proc"
        - "-collector.sysfs=/host/sys"
        name: node-exporter
        ports:
        - containerPort: 9100
          hostPort: 9100
          name: scrape
        resources:
          requests:
            memory: 30Mi
            cpu: 100m
          limits:
            memory: 50Mi
            cpu: 200m
        volumeMounts:
        - name: proc
          readOnly:  true
          mountPath: /host/proc
        - name: sys
          readOnly: true
          mountPath: /host/sys
      volumes:
      - name: proc
        hostPath:
          path: /proc
      - name: sys
        hostPath:
          path: /sys

```

And the respective `Service` manifest:

[embedmd]:# (../../contrib/kube-prometheus/manifests/node-exporter/node-exporter-service.yaml)
```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: node-exporter
    k8s-app: node-exporter
  name: node-exporter
spec:
  type: ClusterIP
  clusterIP: None
  ports:
  - name: http-metrics
    port: 9100
    protocol: TCP
  selector:
    app: node-exporter

```

And last but not least, kube-state-metrics which collects information about Kubernetes objects themselves as they are accessible from the API. Find more information on what kind of metrics kube-state-metrics exposes in [its repository](https://github.com/kubernetes/kube-state-metrics).

[embedmd]:# (../../contrib/kube-prometheus/manifests/kube-state-metrics/kube-state-metrics-deployment.yaml)
```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: kube-state-metrics
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: kube-state-metrics
    spec:
      serviceAccountName: kube-state-metrics
      containers:
      - name: kube-state-metrics
        image: quay.io/coreos/kube-state-metrics:v1.0.0
        ports:
        - name: metrics
          containerPort: 8080
        resources:
          requests:
            memory: 100Mi
            cpu: 100m
          limits:
            memory: 200Mi
            cpu: 200m

```

> Make sure that the `ServiceAccount` called `kube-state-metrics` exists and if using RBAC, is bound to the correct role. See the kube-state-metrics [repository for RBAC requirements](https://github.com/kubernetes/kube-state-metrics/tree/master/kubernetes).

And the respective `Service` manifest:

[embedmd]:# (../../contrib/kube-prometheus/manifests/kube-state-metrics/kube-state-metrics-service.yaml)
```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: kube-state-metrics
    k8s-app: kube-state-metrics
  name: kube-state-metrics
spec:
  ports:
  - name: http-metrics
    port: 8080
    targetPort: metrics
    protocol: TCP
  selector:
    app: kube-state-metrics

```

## Setup Monitoring

Once all the steps in the previous section have been taken there should be `Endpoints` objects containing the IPs of all of the above mentioned Kubernetes components. Now to setup the actual Prometheus and Alertmanager clusters. This manifest assumes that the Alertmanager cluster will be deployed in the `monitoring` namespace.

[embedmd]:# (../../contrib/kube-prometheus/manifests/prometheus/prometheus-k8s.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: k8s
  labels:
    prometheus: k8s
spec:
  replicas: 2
  version: v1.7.0
  serviceAccountName: prometheus-k8s
  serviceMonitorSelector:
    matchExpressions:
    - {key: k8s-app, operator: Exists}
  ruleSelector:
    matchLabels:
      role: prometheus-rulefiles
      prometheus: k8s
  resources:
    requests:
      # 2Gi is default, but won't schedule if you don't have a node with >2Gi
      # memory. Modify based on your target and time-series count for
      # production use. This value is mainly meant for demonstration/testing
      # purposes.
      memory: 400Mi
  alerting:
    alertmanagers:
    - namespace: monitoring
      name: alertmanager-main
      port: web
```

> Make sure that the `ServiceAccount` called `prometheus-k8s` exists and if using RBAC, is bound to the correct role. Read more on [RBAC when using the Prometheus Operator](../rbac.md).

The expression to match for selecting `ServiceMonitor`s here is that they must have a label which has a key called `k8s-apps`. If you look closely at all the `Service` objects described above they all have a label called `k8s-app` and their component name this allows to conveniently select them with `ServiceMonitor`s.

[embedmd]:# (../../contrib/kube-prometheus/manifests/prometheus/prometheus-k8s-service-monitor-apiserver.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kube-apiserver
  labels:
    k8s-app: apiserver
spec:
  jobLabel: component
  selector:
    matchLabels:
      component: apiserver
      provider: kubernetes
  namespaceSelector:
    matchNames:
    - default
  endpoints:
  - port: https
    interval: 30s
    scheme: https
    tlsConfig:
      caFile: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      serverName: kubernetes
    bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
```

[embedmd]:# (../../contrib/kube-prometheus/manifests/prometheus/prometheus-k8s-service-monitor-kubelet.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kubelet
  labels:
    k8s-app: kubelet
spec:
  jobLabel: k8s-app
  endpoints:
  - port: http-metrics
    interval: 30s
  - port: cadvisor
    interval: 30s
    honorLabels: true
  selector:
    matchLabels:
      k8s-app: kubelet
  namespaceSelector:
    matchNames:
    - kube-system
```

[embedmd]:# (../../contrib/kube-prometheus/manifests/prometheus/prometheus-k8s-service-monitor-kube-controller-manager.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kube-controller-manager
  labels:
    k8s-app: kube-controller-manager
spec:
  jobLabel: k8s-app
  endpoints:
  - port: http-metrics
    interval: 30s
  selector:
    matchLabels:
      k8s-app: kube-controller-manager
  namespaceSelector:
    matchNames:
    - kube-system
```

[embedmd]:# (../../contrib/kube-prometheus/manifests/prometheus/prometheus-k8s-service-monitor-kube-scheduler.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kube-scheduler
  labels:
    k8s-app: kube-scheduler
spec:
  jobLabel: k8s-app
  endpoints:
  - port: http-metrics
    interval: 30s
  selector:
    matchLabels:
      k8s-app: kube-scheduler
  namespaceSelector:
    matchNames:
    - kube-system
```

[embedmd]:# (../../contrib/kube-prometheus/manifests/prometheus/prometheus-k8s-service-monitor-kube-state-metrics.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kube-state-metrics
  labels:
    k8s-app: kube-state-metrics
spec:
  jobLabel: k8s-app
  selector:
    matchLabels:
      k8s-app: kube-state-metrics
  namespaceSelector:
    matchNames:
    - monitoring
  endpoints:
  - port: http-metrics
    interval: 30s
    honorLabels: true
```

[embedmd]:# (../../contrib/kube-prometheus/manifests/prometheus/prometheus-k8s-service-monitor-node-exporter.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: node-exporter
  labels:
    k8s-app: node-exporter
spec:
  jobLabel: k8s-app
  selector:
    matchLabels:
      k8s-app: node-exporter
  namespaceSelector:
    matchNames:
    - monitoring
  endpoints:
  - port: http-metrics
    interval: 30s
```

And the Alertmanager:

[embedmd]:# (../../contrib/kube-prometheus/manifests/alertmanager/alertmanager.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: main
  labels:
    alertmanager: main
spec:
  replicas: 3
  version: v0.7.1
```

Read more in the [alerting guide](alerting.md) on how to configure the Alertmanager as it will not spin up unless it has a valid configuration mounted through a `Secret`. Note that the `Secret` has to be in the same namespace as the `Alertmanager` resource as well as have the name `alertmanager-<name-of-alertmanager-object` and the key of the configuration is `alertmanager.yaml`.

## Outlook

Once finished with this guide you have an entire monitoring pipeline for Kubernetes. To now access the web UIs they need to be exposed by the Kubernetes cluster, read through the [exposing Prometheus and Alertmanager guide](exposing-prometheus-and-alertmanager.md) to find out how.

To help get started more quickly with monitoring Kubernetes clusters, [kube-prometheus](https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus) was created. It is a collection of manifests including dashboards and alerting rules that can easily be deployed. It utilizes the Prometheus Operator and all the manifests demonstrated in this guide.
