<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.12.0, Prometheus Operator requires use of Kubernetes v1.7.x and up.
</div>

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


> We assume that the kubelet uses token authN and authZ, as otherwise Prometheus needs a client certificate, which gives it full access to the kubelet, rather than just the metrics. Token authN and authZ allows more fine grained and easier access control. Simply start minikube with the following command (you can of course adapt the version and memory to your needs):
>
> $ minikube delete && minikube start --kubernetes-version=v1.9.1 --memory=4096 --bootstrapper=kubeadm --extra-config=kubelet.authentication-token-webhook=true --extra-config=kubelet.authorization-mode=Webhook --extra-config=scheduler.address=0.0.0.0 --extra-config=controller-manager.address=0.0.0.0
>
> In future versions of minikube and kubeadm this will be the default, but for the time being, we will have to configure it ourselves.

[embedmd]:# (../../contrib/kube-prometheus/manifests/0prometheus-operator-deployment.yaml)
```yaml
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  labels:
    k8s-app: prometheus-operator
  name: prometheus-operator
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: prometheus-operator
  template:
    metadata:
      labels:
        k8s-app: prometheus-operator
    spec:
      containers:
      - args:
        - --kubelet-service=kube-system/kubelet
        - --logtostderr=true
        - --config-reloader-image=quay.io/coreos/configmap-reload:v0.0.1
        - --prometheus-config-reloader=quay.io/coreos/prometheus-config-reloader:v0.28.0
        image: quay.io/coreos/prometheus-operator:v0.28.0
        name: prometheus-operator
        ports:
        - containerPort: 8080
          name: http
        resources:
          limits:
            cpu: 200m
            memory: 200Mi
          requests:
            cpu: 100m
            memory: 100Mi
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
      nodeSelector:
        beta.kubernetes.io/os: linux
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
      serviceAccountName: prometheus-operator
```

> Make sure that the `ServiceAccount` called `prometheus-operator` exists and if using RBAC, is bound to the correct role. Read more on [RBAC when using the Prometheus Operator](../rbac.md).

Once started it ensures that all internal IPs of the nodes in the cluster are synchronized into the specified `Endpoints` object. In this case the object is called `kubelet` and is located in the `kube-system` namespace.

By default every Kubernetes cluster has a `Service` for easy access to the API server. This is the `Service` called `kubernetes` in the `default` namespace. A `Service` object automatically synchronizes an `Endpoints` object with the targets it selects. Therefore there is nothing, extra to do for Prometheus to be able to discover the API server.

Aside from the kubelet and the API server the other Kubernetes components all run on top of Kubernetes itself. To discover Kubernetes components that run in a Pod, they simply have to be added to a `Service`.

> Note the `Service` manifests for the scheduler and controller-manager are just examples. They may need to be adapted respective to a cluster.

kube-scheduler:

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

[embedmd]:# (../../contrib/kube-prometheus/manifests/node-exporter-daemonset.yaml)
```yaml
apiVersion: apps/v1beta2
kind: DaemonSet
metadata:
  labels:
    app: node-exporter
  name: node-exporter
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: node-exporter
  template:
    metadata:
      labels:
        app: node-exporter
    spec:
      containers:
      - args:
        - --web.listen-address=127.0.0.1:9100
        - --path.procfs=/host/proc
        - --path.sysfs=/host/sys
        - --path.rootfs=/host/root
        - --collector.filesystem.ignored-mount-points=^/(dev|proc|sys|var/lib/docker/.+)($|/)
        - --collector.filesystem.ignored-fs-types=^(autofs|binfmt_misc|cgroup|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|mqueue|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|sysfs|tracefs)$
        image: quay.io/prometheus/node-exporter:v0.17.0
        name: node-exporter
        resources:
          limits:
            cpu: 250m
            memory: 180Mi
          requests:
            cpu: 102m
            memory: 180Mi
        volumeMounts:
        - mountPath: /host/proc
          name: proc
          readOnly: false
        - mountPath: /host/sys
          name: sys
          readOnly: false
        - mountPath: /host/root
          mountPropagation: HostToContainer
          name: root
          readOnly: true
      - args:
        - --logtostderr
        - --secure-listen-address=$(IP):9100
        - --tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256
        - --upstream=http://127.0.0.1:9100/
        env:
        - name: IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        image: quay.io/coreos/kube-rbac-proxy:v0.4.1
        name: kube-rbac-proxy
        ports:
        - containerPort: 9100
          hostPort: 9100
          name: https
        resources:
          limits:
            cpu: 20m
            memory: 40Mi
          requests:
            cpu: 10m
            memory: 20Mi
      hostNetwork: true
      hostPID: true
      nodeSelector:
        beta.kubernetes.io/os: linux
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
      serviceAccountName: node-exporter
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      volumes:
      - hostPath:
          path: /proc
        name: proc
      - hostPath:
          path: /sys
        name: sys
      - hostPath:
          path: /
        name: root
```

The respective `ServiceAccount` manifest:

[embedmd]:# (../../contrib/kube-prometheus/manifests/node-exporter-serviceAccount.yaml)
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: node-exporter
  namespace: monitoring
```


And the respective `Service` manifest:

[embedmd]:# (../../contrib/kube-prometheus/manifests/node-exporter-service.yaml)
```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    k8s-app: node-exporter
  name: node-exporter
  namespace: monitoring
spec:
  clusterIP: None
  ports:
  - name: https
    port: 9100
    targetPort: https
  selector:
    app: node-exporter
```

And last but not least, kube-state-metrics which collects information about Kubernetes objects themselves as they are accessible from the API. Find more information on what kind of metrics kube-state-metrics exposes in [its repository](https://github.com/kubernetes/kube-state-metrics).

[embedmd]:# (../../contrib/kube-prometheus/manifests/kube-state-metrics-deployment.yaml)
```yaml
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  labels:
    app: kube-state-metrics
  name: kube-state-metrics
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kube-state-metrics
  template:
    metadata:
      labels:
        app: kube-state-metrics
    spec:
      containers:
      - args:
        - --logtostderr
        - --secure-listen-address=:8443
        - --tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256
        - --upstream=http://127.0.0.1:8081/
        image: quay.io/coreos/kube-rbac-proxy:v0.4.1
        name: kube-rbac-proxy-main
        ports:
        - containerPort: 8443
          name: https-main
        resources:
          limits:
            cpu: 20m
            memory: 40Mi
          requests:
            cpu: 10m
            memory: 20Mi
      - args:
        - --logtostderr
        - --secure-listen-address=:9443
        - --tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256
        - --upstream=http://127.0.0.1:8082/
        image: quay.io/coreos/kube-rbac-proxy:v0.4.1
        name: kube-rbac-proxy-self
        ports:
        - containerPort: 9443
          name: https-self
        resources:
          limits:
            cpu: 20m
            memory: 40Mi
          requests:
            cpu: 10m
            memory: 20Mi
      - args:
        - --host=127.0.0.1
        - --port=8081
        - --telemetry-host=127.0.0.1
        - --telemetry-port=8082
        image: quay.io/coreos/kube-state-metrics:v1.5.0
        name: kube-state-metrics
        resources:
          limits:
            cpu: 100m
            memory: 150Mi
          requests:
            cpu: 100m
            memory: 150Mi
      - command:
        - /pod_nanny
        - --container=kube-state-metrics
        - --cpu=100m
        - --extra-cpu=2m
        - --memory=150Mi
        - --extra-memory=30Mi
        - --acceptance-offset=5
        - --deployment=kube-state-metrics
        env:
        - name: MY_POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: MY_POD_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        image: gcr.io/google-containers/addon-resizer-amd64:2.1
        name: addon-resizer
        resources:
          limits:
            cpu: 50m
            memory: 30Mi
          requests:
            cpu: 10m
            memory: 30Mi
      nodeSelector:
        beta.kubernetes.io/os: linux
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
      serviceAccountName: kube-state-metrics
```

> Make sure that the `ServiceAccount` called `kube-state-metrics` exists and if using RBAC, is bound to the correct role. See the kube-state-metrics [repository for RBAC requirements](https://github.com/kubernetes/kube-state-metrics/tree/master/kubernetes).

And the respective `Service` manifest:

[embedmd]:# (../../contrib/kube-prometheus/manifests/kube-state-metrics-service.yaml)
```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    k8s-app: kube-state-metrics
  name: kube-state-metrics
  namespace: monitoring
spec:
  clusterIP: None
  ports:
  - name: https-main
    port: 8443
    targetPort: https-main
  - name: https-self
    port: 9443
    targetPort: https-self
  selector:
    app: kube-state-metrics
```

## Setup Monitoring

Once all the steps in the previous section have been taken there should be `Endpoints` objects containing the IPs of all of the above mentioned Kubernetes components. Now to setup the actual Prometheus and Alertmanager clusters. This manifest assumes that the Alertmanager cluster will be deployed in the `monitoring` namespace.

[embedmd]:# (../../contrib/kube-prometheus/manifests/prometheus-prometheus.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  labels:
    prometheus: k8s
  name: k8s
  namespace: monitoring
spec:
  alerting:
    alertmanagers:
    - name: alertmanager-main
      namespace: monitoring
      port: web
  baseImage: quay.io/prometheus/prometheus
  nodeSelector:
    beta.kubernetes.io/os: linux
  replicas: 2
  resources:
    requests:
      memory: 400Mi
  ruleSelector:
    matchLabels:
      prometheus: k8s
      role: alert-rules
  securityContext:
    fsGroup: 2000
    runAsNonRoot: true
    runAsUser: 1000
  serviceAccountName: prometheus-k8s
  serviceMonitorNamespaceSelector: {}
  serviceMonitorSelector: {}
  version: v2.5.0
```

> Make sure that the `ServiceAccount` called `prometheus-k8s` exists and if using RBAC, is bound to the correct role. Read more on [RBAC when using the Prometheus Operator](../rbac.md).

The expression to match for selecting `ServiceMonitor`s here is that they must have a label which has a key called `k8s-app`. If you look closely at all the `Service` objects described above they all have a label called `k8s-app` and their component name this allows to conveniently select them with `ServiceMonitor`s.

[embedmd]:# (../../contrib/kube-prometheus/manifests/prometheus-serviceMonitorApiserver.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    k8s-app: apiserver
  name: kube-apiserver
  namespace: monitoring
spec:
  endpoints:
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    interval: 30s
    metricRelabelings:
    - action: drop
      regex: etcd_(debugging|disk|request|server).*
      sourceLabels:
      - __name__
    - action: drop
      regex: apiserver_admission_controller_admission_latencies_seconds_.*
      sourceLabels:
      - __name__
    - action: drop
      regex: apiserver_admission_step_admission_latencies_seconds_.*
      sourceLabels:
      - __name__
    port: https
    scheme: https
    tlsConfig:
      caFile: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      serverName: kubernetes
  jobLabel: component
  namespaceSelector:
    matchNames:
    - default
  selector:
    matchLabels:
      component: apiserver
      provider: kubernetes
```

[embedmd]:# (../../contrib/kube-prometheus/manifests/prometheus-serviceMonitorKubelet.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    k8s-app: kubelet
  name: kubelet
  namespace: monitoring
spec:
  endpoints:
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    honorLabels: true
    interval: 30s
    port: https-metrics
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    honorLabels: true
    interval: 30s
    metricRelabelings:
    - action: drop
      regex: container_([a-z_]+);
      sourceLabels:
      - __name__
      - image
    - action: drop
      regex: container_(network_tcp_usage_total|network_udp_usage_total|tasks_state|cpu_load_average_10s)
      sourceLabels:
      - __name__
    path: /metrics/cadvisor
    port: https-metrics
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
  jobLabel: k8s-app
  namespaceSelector:
    matchNames:
    - kube-system
  selector:
    matchLabels:
      k8s-app: kubelet
```

[embedmd]:# (../../contrib/kube-prometheus/manifests/prometheus-serviceMonitorKubeControllerManager.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    k8s-app: kube-controller-manager
  name: kube-controller-manager
  namespace: monitoring
spec:
  endpoints:
  - interval: 30s
    metricRelabelings:
    - action: drop
      regex: etcd_(debugging|disk|request|server).*
      sourceLabels:
      - __name__
    port: http-metrics
  jobLabel: k8s-app
  namespaceSelector:
    matchNames:
    - kube-system
  selector:
    matchLabels:
      k8s-app: kube-controller-manager
```

[embedmd]:# (../../contrib/kube-prometheus/manifests/prometheus-serviceMonitorKubeScheduler.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    k8s-app: kube-scheduler
  name: kube-scheduler
  namespace: monitoring
spec:
  endpoints:
  - interval: 30s
    port: http-metrics
  jobLabel: k8s-app
  namespaceSelector:
    matchNames:
    - kube-system
  selector:
    matchLabels:
      k8s-app: kube-scheduler
```

[embedmd]:# (../../contrib/kube-prometheus/manifests/kube-state-metrics-serviceMonitor.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    k8s-app: kube-state-metrics
  name: kube-state-metrics
  namespace: monitoring
spec:
  endpoints:
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    honorLabels: true
    interval: 30s
    port: https-main
    scheme: https
    scrapeTimeout: 30s
    tlsConfig:
      insecureSkipVerify: true
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    interval: 30s
    port: https-self
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
  jobLabel: k8s-app
  selector:
    matchLabels:
      k8s-app: kube-state-metrics
```

[embedmd]:# (../../contrib/kube-prometheus/manifests/node-exporter-serviceMonitor.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    k8s-app: node-exporter
  name: node-exporter
  namespace: monitoring
spec:
  endpoints:
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    interval: 30s
    port: https
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
  jobLabel: k8s-app
  selector:
    matchLabels:
      k8s-app: node-exporter
```

And the Alertmanager:

[embedmd]:# (../../contrib/kube-prometheus/manifests/alertmanager-alertmanager.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  labels:
    alertmanager: main
  name: main
  namespace: monitoring
spec:
  baseImage: quay.io/prometheus/alertmanager
  nodeSelector:
    beta.kubernetes.io/os: linux
  replicas: 3
  securityContext:
    fsGroup: 2000
    runAsNonRoot: true
    runAsUser: 1000
  serviceAccountName: alertmanager-main
  version: v0.16.0
```

Read more in the [alerting guide](alerting.md) on how to configure the Alertmanager as it will not spin up unless it has a valid configuration mounted through a `Secret`. Note that the `Secret` has to be in the same namespace as the `Alertmanager` resource as well as have the name `alertmanager-<name-of-alertmanager-object>` and the key of the configuration is `alertmanager.yaml`.

## Outlook

Once finished with this guide you have an entire monitoring pipeline for Kubernetes. To now access the web UIs they need to be exposed by the Kubernetes cluster, read through the [exposing Prometheus and Alertmanager guide](exposing-prometheus-and-alertmanager.md) to find out how.

To help get started more quickly with monitoring Kubernetes clusters, [kube-prometheus](https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus) was created. It is a collection of manifests including dashboards and alerting rules that can easily be deployed. It utilizes the Prometheus Operator and all the manifests demonstrated in this guide.
