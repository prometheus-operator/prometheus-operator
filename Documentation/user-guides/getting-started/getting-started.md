# Prometheus Operator

The mission of the Prometheus Operator is to make running Prometheus on top of
Kubernetes as easy as possible, while preserving configurability as well as
making the configuration Kubernetes native.

To follow this getting started you will need a Kubernetes cluster you have
access to. Let's give the Prometheus Operator a spin:

[embedmd]:# (../../../deployment.yaml)
```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: prometheus-operator
  labels:
    operator: prometheus
spec:
  replicas: 1
  template:
    metadata:
      labels:
        operator: prometheus
    spec:
      containers:
       - name: prometheus-operator
         image: quay.io/coreos/prometheus-operator:v0.6.0
         resources:
           requests:
             cpu: 100m
             memory: 50Mi
           limits:
             cpu: 200m
             memory: 100Mi
```

The Prometheus Operator introduces third party resources in Kubernetes to
declare the desired state of a Prometheus and Alertmanager cluster as well as
the Prometheus configuration. The resources it introduces are:

* [`Prometheus`](../../prometheus.md)
* [`Alertmanager`](../../alertmanager.md)
* [`ServiceMonitor`](../../service-monitor.md)

> Important for this guide are the `Prometheus` and `ServiceMonitor` resources.
> Have a look at the [Alerting guide](../alerting/alerting.md) for more
> information about the `Alertmanager` resource.

The Prometheus resource includes fields such as the desired version of
Prometheus to run, the number of replicas, as well as a number of parameters to
configure Prometheus itself. The connection of the Prometheus resource to the
`ServiceMonitor` is established through the `serviceMonitorSelector`, which
selects which `ServiceMonitor`s are to be used to generate the configuration
file for Prometheus.

We will walk through an example application that could look like this:

[embedmd]:# (examples/example-app-deployment.yaml)
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

Essentially the `ServiceMonitor` has a label selector to select `Service`s and
the underlying `Endpoints` objects. For example one might have a `Service` that
looks like the following:

[embedmd]:# (examples/example-app-service.yaml)
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

This `Service` object could be discovered by a `ServiceMonitor`.

[embedmd]:# (examples/example-app-service-monitor.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1alpha1
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

Finally a `Prometheus` object defines the `serviceMonitorSelector` to specify
which `ServiceMonitor`s should be included when generating the Prometheus
configuration.

[embedmd]:# (examples/prometheus-example.yaml)
```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: Prometheus
metadata:
  name: example
spec:
  serviceMonitorSelector:
    matchLabels:
      team: frontend
  resources:
    requests:
      memory: 400Mi
```

To be able to access the Prometheus instance it will have to be exposed to the
outside somehow. Purely for demonstration purpose we will expose it via a
`Service` of type `NodePort`.

[embedmd]:# (examples/prometheus-example-service.yaml)
```yaml
apiVersion: v1
kind: Service
metadata:
  name: prometheus-example
spec:
  type: NodePort
  ports:
  - name: web
    nodePort: 30900
    port: 9090
    protocol: TCP
    targetPort: web
  selector:
    prometheus: example
```

Once this `Service` is created the Prometheus web UI is available under the
node's IP address on port `30900`. The targets page in the web UI now shows
that the instances of the example application have successfully been
discovered.

> Exposing the Prometheus web UI may not be an applicable solution. Read more
> about the possibilities of exposing it in the [exposing Prometheus and
> Alertmanager guide](../exposing-prometheus-and-alertmanager/exposing-prometheus-and-alertmanager.md).

Further reading:

* In addition to managing Prometheus clusters the Prometheus Operator can also
  manage Alertmanager clusters. Learn more in the [Alerting
  guide](../alerting/alerting.md).

* Monitoring the Kubernetes cluster itself. Learn more in the [Cluster
  Monitoring guide](../cluster-monitoring/cluster-monitoring.md)
