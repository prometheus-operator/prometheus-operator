# Prometheus Operator

**Project status: *alpha*** Not all planned features are completed. The API, spec, status 
and other user facing objects are subject to change. We do not support backward-compability 
for the alpha releases.

The Prometheus Operator for Kubernetes provides easy monitoring definitions for Kubernetes
services and deployment and management of Prometheus instances.

Once installed the Prometheus Operator provides the following features:

* **Create/Destroy**: Easily launch a Prometheus instance for your Kubernetes namespace,
  a specific application or team easily using the Operator.

* **Simple Configuration**: Configure the fundamentals of Prometheus like versions, persistence, 
  retention policies, and replicas from a native Kubernetes resource.

* **Target Services via Labels**: Automatically generate monitoring target configurations based
  on familiar Kubernetes label queries; no need to learn a Prometheus specific configuration language.


## Third party resources

The Operator acts on two third party resources (TPRs):

* **[`Prometheus`](./Documentation/prometheus.md)**, which defines a desired Prometheus deployment.
  The Operator ensures at all times that a deployment matching the resource definition is running.

* **[`ServiceMonitor`](./Documentation/service-monitor.md)**, which declaratively specifies how groups
  of services should be monitored. The Operator automatically generates Prometheus scrape configuration
  based on the definition.


## Installation

You can install the Operator inside of your cluster by running the following command:

```
kubectl apply -f deployment.yaml
```

To run the Operator outside of your cluster:

```
make
hack/run-external.sh <kubectl cluster name>
```
