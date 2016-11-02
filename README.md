# Prometheus Operator

The Prometheus Operator for Kubernetes provides easy monitoring definitions
for Kubernetes services and deployment and management of Prometheus instances.

## Third party resources

The controller acts on two third party resources (TPRs).

### ServiceMonitor

A service monitor provides definition about how a selection of services
should be monitored.

Example of a ServiceMonitor that specifies how all services of the mobile team
written with the fictional "gokit" framework should be monitored.

```yaml
apiVersion: "prometheus.coreos.com/v1alpha1"
kind: "ServiceMonitor"
metadata:
  name: "example-app"
  labels:
    team: mobile
spec:
  # Selection of services the monitoring definition applies to and rule
  # file ConfigMaps used for the service.
  selector:
    matchLabels:
      framework: gokit
      team:      mobile
  # Endpoints of the service or their underlying pods that can be monitored.
  endpoints:
  - port: web            # Name of the service port.
    interval: 30s        # Interval at which the service endpoints will be scraped.
  - targetPort: metrics  # Name or number of the target port of a service endpoint.
    path: /varz          # HTTP path that exposes metrics.
    scheme: https        # To use http or https when scraping. 
    interval: 60s
```

The controller generates Prometheus `job` names of the pattern `<>`.

### Prometheus

The Prometheus TPR selects ServiceMonitors by their labels and specifies additional
configuration for the deployed Prometheus server instances.
The Operator watches Prometheus objects and deploys actual Prometheus servers
configured to match the ServiceMonitor definitions the object selects.

Example of defining a Prometheus server deployment, that monitors all services that
were specified with ServiceMonitors with the `team=mobile` label.

```yaml
apiVersion: "prometheus.coreos.com/v1alpha1"
kind: "Prometheus"
metadata:
  name: "prometheus-mobile"
  labels:
    prometheus: "mobile"
spec:
  baseImage: quay.io/prometheus/prometheus # default
  version: v1.3.0                          # default, must match image tag
  replicas: 2                              # defaults to 1
  serviceMonitors:
  - selector:
      matchLabels:
        team: mobile
  evaluationInterval: 30s
  alerting:
    # namespace/name of Alertmanager services.
    alertmanagers:
    - namespace: monitoring
      name: alertmanager
  storage:
    # Options for persistent volume claims created for each instance
    class: "ssd"    # storage class name
    resources:
      requests: 20Gi
```

## Installation

You can install the controller inside of your cluster by running

```
kubectl apply -f example/prometheus-controller.yaml
```

To run the controller outside of your cluster:

```
make
hack/controller-external.sh <kubectl cluster name>
```

### Roadmap / Ideas

Roughly in order of importance:

* Namespace configuration/limitation of discovered monitoring targets
* Global desired Prometheus version with optional pinning
* Dynamic mounting of recording/alerting rule ConfigMaps
* Configuring receiving AlertManagers 
* Persistent volume mounts for time series data
* Resource limits for deployed servers and auto-tuned storage flags based on them
* Retention configuration; potentially auto-adapting to remaining storage on persistent volume
* Automatic horizontal sharding

