# Prometheus Controller

The Prometheus Controller for Kubernetes provides easy monitoring definitions
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
      team:       mobile
  # Endpoints of the service or their underlying pods that can be monitored.
  endpoints:
  - port: web            # Name of the service port.
    scrapeInterval: 30s  # Interval at which the service endpoints will be scraped.
  - targetPort: metrics  # Name or number of the target port of a service endpoint.
    path: /varz          # HTTP path that exposes metrics.
    scheme: https        # To use http or https when scraping. 
    scrapeInterval: 60s
```

The controller generates Prometheus `job` names of the pattern `<>`.

### Prometheus

The Prometheus TPR selects ServiceMonitors by their labels and specifies additional
configuration for the deployed Prometheus server instances.
The Controller watches Prometheus objects and deploys actual Prometheus servers
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
  serviceMonitors:
  - selector:
      matchLabels:
        team: mobile
  ruleEvaluationInterval: 30s
```
