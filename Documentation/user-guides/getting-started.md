# Prometheus Operator

Operators were introduced by CoreOS as a class of software that operates other software, putting operational knowledge collected by humans into software. Read more in the [original blog post](https://coreos.com/blog/introducing-operators.html).

The mission of the Prometheus Operator is to make running Prometheus on top of Kubernetes as easy as possible, while preserving configurability as well as making the configuration Kubernetes native.

To follow this getting started you will need a Kubernetes cluster you have access to. Let's give the Prometheus Operator a spin:

[embedmd]:# (../../bundle.yaml)
```yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: prometheus-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus-operator
subjects:
- kind: ServiceAccount
  name: prometheus-operator
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: prometheus-operator
rules:
- apiGroups:
  - extensions
  resources:
  - thirdpartyresources
  verbs:
  - "*"
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - "*"
- apiGroups:
  - monitoring.coreos.com
  resources:
  - alertmanagers
  - prometheuses
  - servicemonitors
  verbs:
  - "*"
- apiGroups:
  - apps
  resources:
  - statefulsets
  verbs: ["*"]
- apiGroups: [""]
  resources:
  - configmaps
  - secrets
  verbs: ["*"]
- apiGroups: [""]
  resources:
  - pods
  verbs: ["list", "delete"]
- apiGroups: [""]
  resources:
  - services
  - endpoints
  verbs: ["get", "create", "update"]
- apiGroups: [""]
  resources:
  - nodes
  verbs: ["list", "watch"]
- apiGroups: [""]
  resources:
  - namespaces
  verbs: ["list"]
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus-operator
---
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

The Prometheus Operator introduces third party resources in Kubernetes to declare the desired state of a Prometheus and Alertmanager cluster as well as the Prometheus configuration. The resources it introduces are:

* `Prometheus`
* `Alertmanager`
* `ServiceMonitor`

> Important for this guide are the `Prometheus` and `ServiceMonitor` resources. Have a look at the [alerting guide](alerting.md) for more information about the `Alertmanager` resource or the [design doc](../design.md) for an overview of all resources introduced by the Prometheus Operator.

The `Prometheus` resource declaratively describes the desired state of a Prometheus deployment, while a `ServiceMonitor` describes the set of targets to be monitored by Prometheus.

![Prometheus Operator Architecture](images/architecture.png "Prometheus Operator Architecture")

The `Prometheus` resource includes a selection of `ServiceMonitors` to be used, this field is called the `serviceMonitorSelector`.

First, deploy three instances of a simple example application, which listens and exposes metrics on port `8080`.

[embedmd]:# (../../example/user-guides/getting-started/example-app-deployment.yaml)
```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: example-app
spec:
  replicas: 3
  template:
    metadata:
      labels:
        app: example-app
    spec:
      containers:
      - name: example-app 
        image: fabxc/instrumented_app
        ports:
        - name: web
          containerPort: 8080
```

The `ServiceMonitor` has a label selector to select `Service`s and the underlying `Endpoints` objects. The `Service` object for the example application selects the `Pod`s by the `app` label having the `example-app` value. In addition to that the `Service` object specifies the port the metrics are exposed on.

[embedmd]:# (../../example/user-guides/getting-started/example-app-service.yaml)
```yaml
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

This `Service` object is discovered by a `ServiceMonitor`, which selects in the same way. The `app` label must have the value `example-app`.

[embedmd]:# (../../example/user-guides/getting-started/example-app-service-monitor.yaml)
```yaml
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

---

>**Important**: If you have [RBAC](https://kubernetes.io/docs/admin/authorization/) authorization activated you need to create RBAC rules for both *Prometheus* and *Prometheus Operator*. We already created a `ClusterRole` and a `ClusterRoleBinding` for the *Prometheus Operator* in the first step. The same has to be done for the *Prometheus* Pods:

[embedmd]:# (../../example/rbac/prometheus/prometheus-service-account.yaml)
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus
```

[embedmd]:# (../../example/rbac/prometheus/prometheus-cluster-role.yaml)
```yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: prometheus
rules:
- apiGroups: [""]
  resources:
  - nodes
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources:
  - configmaps
  verbs: ["get"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]
```

[embedmd]:# (../../example/rbac/prometheus/prometheus-cluster-role-binding.yaml)
```yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: prometheus
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus
subjects:
- kind: ServiceAccount
  name: prometheus
  namespace: default
```

You can find further details in the [*Prometheus Operator RBAC* guide](../rbac.md).

---

Finally, a `Prometheus` object defines the `serviceMonitorSelector` to specify which `ServiceMonitor`s should be included. Above the label `team: frontend` was specified, so that's what the `Prometheus` object selects by.

[embedmd]:# (../../example/user-guides/getting-started/prometheus.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
spec:
  serviceMonitorSelector:
    matchLabels:
      team: frontend
  version: v1.7.0
  resources:
    requests:
      memory: 400Mi
```
> If you have RBAC authorization activated, use the RBAC aware [*Prometheus* manifest](../../example/rbac/prometheus/prometheus-cluster-role-binding.yaml) instead.


This way the frontend team can create new `ServiceMonitor`s and `Service`s resulting in `Prometheus` to be dynamically reconfigured.

To be able to access the Prometheus instance it will have to be exposed to the outside somehow. For demonstration purpose it will be exposed via a `Service` of type `NodePort`.

[embedmd]:# (../../example/user-guides/getting-started/prometheus-service.yaml)
```yaml
apiVersion: v1
kind: Service
metadata:
  name: prometheus
spec:
  type: NodePort
  ports:
  - name: web
    nodePort: 30900
    port: 9090
    protocol: TCP
    targetPort: web
  selector:
    prometheus: prometheus
```

Once this `Service` is created the Prometheus web UI is available under the node's IP address on port `30900`. The targets page in the web UI now shows that the instances of the example application have successfully been discovered.

> Exposing the Prometheus web UI may not be an applicable solution. Read more about the possibilities of exposing it in the [exposing Prometheus and Alertmanager guide](exposing-prometheus-and-alertmanager.md).

Further reading:

* In addition to managing Prometheus deployments the Prometheus Operator can also manage Alertmanager clusters. Learn more in the [alerting guide](alerting.md).

* Monitoring the Kubernetes cluster itself. Learn more in the [Cluster Monitoring guide](cluster-monitoring.md).
