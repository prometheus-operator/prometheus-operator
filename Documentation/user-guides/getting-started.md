<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.39.0, Prometheus Operator requires use of Kubernetes v1.16.x and up.<br><br>
This documentation is for an alpha feature. For questions and feedback on the Prometheus OCS Alpha program, email <a href="mailto:tectonic-alpha-feedback@coreos.com">tectonic-alpha-feedback@coreos.com</a>.
</div>

# Prometheus Operator

[Operators](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) were introduced by CoreOS as a class of software that operates other software, putting operational knowledge collected by humans into software.

The Prometheus Operator serves to make running Prometheus on top of Kubernetes as easy as possible, while preserving Kubernetes-native configuration options.

## Example Prometheus Operator manifest

To follow this getting started you will need a Kubernetes cluster you have access to. This [example](../../bundle.yaml) describes a Prometheus Operator Deployment, and its required ClusterRole, ClusterRoleBinding, Service Account and Custom Resource Definitions.

## Related resources

The Prometheus Operator introduces additional resources in Kubernetes to declare the desired state of a Prometheus and Alertmanager cluster as well as the Prometheus configuration. The resources it introduces are:

* `Prometheus`
* `Alertmanager`
* `ServiceMonitor`

> See the [Alerting guide](alerting.md) for more information about the `Alertmanager` resource, or the [Design document](../design.md) for an overview of all resources introduced by the Prometheus Operator.

The Prometheus resource declaratively describes the desired state of a Prometheus deployment, while a ServiceMonitor describes the set of targets to be monitored by Prometheus.

![Prometheus Operator Architecture](images/architecture.png "Prometheus Operator Architecture")

The Prometheus resource includes a field called `serviceMonitorSelector`, which defines a selection of ServiceMonitors to be used. By default and before the version `v0.19.0`, ServiceMonitors must be installed in the same namespace as the Prometheus instance. With the Prometheus Operator `v0.19.0` and above, ServiceMonitors can be selected outside the Prometheus namespace via the `serviceMonitorNamespaceSelector` field of the Prometheus resource.

First, deploy three instances of a simple example application, which listens and exposes metrics on port `8080`.

```yaml mdox-exec="cat example/user-guides/getting-started/example-app-deployment.yaml"
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: example-app
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

The ServiceMonitor has a label selector to select Services and their underlying Endpoint objects. The Service object for the example application selects the Pods by the `app` label having the `example-app` value. The Service object also specifies the port on which the metrics are exposed.

```yaml mdox-exec="cat example/user-guides/getting-started/example-app-service.yaml"
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

This Service object is discovered by a ServiceMonitor, which selects in the same way. The `app` label must have the value `example-app`.

```yaml mdox-exec="cat example/user-guides/getting-started/example-app-service-monitor.yaml"
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

## Enable RBAC rules for Prometheus pods

If [RBAC](https://kubernetes.io/docs/reference/access-authn-authz/authorization/) authorization is activated, you must create RBAC rules for both Prometheus *and* Prometheus Operator. A ClusterRole and a ClusterRoleBinding for the Prometheus Operator were created in the example Prometheus Operator manifest above. The same must be done for the Prometheus Pods.

Create a ClusterRole and ClusterRoleBinding for the Prometheus Pods:

```yaml mdox-exec="cat example/rbac/prometheus/prometheus-service-account.yaml"
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus
```

```yaml mdox-exec="cat example/rbac/prometheus/prometheus-cluster-role.yaml"
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prometheus
rules:
- apiGroups: [""]
  resources:
  - nodes
  - nodes/metrics
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources:
  - configmaps
  verbs: ["get"]
- apiGroups:
  - networking.k8s.io
  resources:
  - ingresses
  verbs: ["get", "list", "watch"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]
```

```yaml mdox-exec="cat example/rbac/prometheus/prometheus-cluster-role-binding.yaml"
apiVersion: rbac.authorization.k8s.io/v1
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

For more information, see the [Prometheus Operator RBAC guide](../rbac.md).

## Include ServiceMonitors

A Prometheus object defines the `serviceMonitorSelector` to specify which ServiceMonitors should be included. Above the label `team: frontend` was specified, so that's what the Prometheus object selects by.

```yaml mdox-exec="cat example/user-guides/getting-started/prometheus-service-monitor.yaml"
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
spec:
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      team: frontend
  resources:
    requests:
      memory: 400Mi
  enableAdminAPI: false
```

> If you have RBAC authorization activated, use the RBAC aware [Prometheus manifest](../../example/rbac/prometheus/prometheus.yaml) instead.

This enables the frontend team to create new ServiceMonitors and Services which allow Prometheus to be dynamically reconfigured.

## Include PodMonitors

Finally, a Prometheus object defines the `podMonitorSelector` to specify which PodMonitors should be included. Above the label `team: frontend` was specified, so that's what the Prometheus object selects by.

```yaml mdox-exec="cat example/user-guides/getting-started/prometheus-pod-monitor.yaml"
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
spec:
  serviceAccountName: prometheus
  podMonitorSelector:
    matchLabels:
      team: frontend
  resources:
    requests:
      memory: 400Mi
  enableAdminAPI: false
```

> If you have RBAC authorization activated, use the RBAC aware [Prometheus manifest](../../example/rbac/prometheus/prometheus.yaml) instead.

This enables the frontend team to create new PodMonitors which allow Prometheus to be dynamically reconfigured.

## Expose the Prometheus instance

To access the Prometheus instance it must be exposed to the outside. This example exposes the instance using a Service of type `NodePort`.

```yaml mdox-exec="cat example/user-guides/getting-started/prometheus-service.yaml"
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

Once this Service is created the Prometheus web UI is available under the node's IP address on port `30900`. The targets page in the web UI now shows that the instances of the example application have successfully been discovered.

> Exposing the Prometheus web UI may not be an applicable solution. Read more about the possibilities of exposing it in the [exposing Prometheus and Alertmanager guide](exposing-prometheus-and-alertmanager.md).

## Expose the Prometheus Admin API

Prometheus Admin API allows access to delete series for a certain time range, cleanup tombstones, capture snapshots, etc. More information about the admin API can be found in [Prometheus official documentation](https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis)
This API access is disabled by default and can be toggled using this boolean flag. The following example exposes the admin API:

> WARNING: Enabling the admin APIs enables mutating endpoints, to delete data,
> shutdown Prometheus, and more. Enabling this should be done with care and the
> user is advised to add additional authentication authorization via a proxy to
> ensure only clients authorized to perform these actions can do so.

```yaml mdox-exec="cat example/user-guides/getting-started/prometheus-admin-api.yaml"
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
spec:
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      team: frontend
  resources:
    requests:
      memory: 400Mi
  enableAdminAPI: true
```

Further reading:

* [Alerting](alerting.md) describes using the Prometheus Operator to manage Alertmanager clusters.
