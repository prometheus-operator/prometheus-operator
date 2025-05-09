# Add BlackboxExporter and BlackboxExporterModule CRDs

* **Owners:**
  * [Pger-Y](https://github.com/Pger-Y)

* **Related Tickets:**
  * [Discussion](https://github.com/prometheus-operator/prometheus-operator/discussions/7134#discussioncomment-11569949)

* **Other docs:**
  * n/a

This document proposes the creation of a Blackbox Operator and associated CRDs, similar to Prometheus-Operator, to manage blackbox-exporter instances and introduce a Module CRD for handling dynamic configuration of blackbox modules in a Kubernetes-native way.



## Why
The Probe CRD can define all the parameters that the ICMP Probe requires and is quite convenient for most use cases. But configuring the HTTP Probe using the Probe CRD is a lot more complex, requiring the user to define plenty of parameters. This makes the Probe CRD an insufficient solution and potentially error-prone for HTTP Probe use cases.

The situation is very similar to Alertmanager/AlertmanagerConfig that prometheus-operator offers. We need a Kubernetes-native way to generate blackbox config dynamically by reading the Probe Module Config.


## Pitfalls of the current solution

The Probe CRD provides a partial solution for handling complex modules like HTTP and gRPC. However, it introduces a separate lifecycle for Blackbox management. Users must pre-define modules containing specific configurations (such as headers and request bodies) by manually updating and reloading the Blackbox configuration, before referencing these modules in their probes.

## Goals

- Provide a way for users to self-service adding module config for their Probe CRD
- Consolidate the module configuration generation logic in a central point for other resources to use

## Audience

* Administrators providing Prometheus monitoring who want to enable self-service probe configuration for their users
* Teams wanting to manage blackbox module configurations declaratively through Kubernetes CRDs
* Users requiring a standardized, Kubernetes-native approach to monitor external targets

## Non-Goals

- This propose don't aim to replace the probe crd

## How

We propose creating two new Custom Resource Definitions (CRDs):
- The BlackboxExporter CRD, which functions similarly to the Prometheus CRD, will manage the deployment configuration and lifecycle of blackbox-exporter instances
- The BlackboxExporterModule CRD will define module specifications and automatically generate blackbox configurations, analogous to how Alertmanager works in Prometheus Operator

## Alternatives
Users may need to update or create a Probe CRD to monitor an HTTP endpoint with specific headers or body content. This often requires updating the Blackbox Exporter configuration, such as creating a new module or modifying an existing one. While users can manage this manually, it is a more cumbersome and error-prone process.

## Action Plan

1. Create the BlackboxExporter CRD which will manage blackbox-exporter instances as either a DaemonSet or Deployment
2. Implement a configuration reloader sidecar container to automatically reload blackbox-exporter when configuration changes, similar to how Prometheus Operator handles configuration updates

## CRDs Draft
### 1. BlackBoxExporter CRD
```yaml
apiVersion: monitoring.zhipu.ai/v1alpha1
kind: BlackboxExporter
metadata:
  name: blackbox-exporter-draft
spec:
  # Deployment mode - how the BlackboxExporter should be deployed 
    # Or Daemonset mode
  deployMode: Deployment
  replicas: 2
  
  # Key configuration options
  image: prom/blackbox-exporter
  timeoutOffset: "0.5"
  logLevelProber: "info"
  
  # Module selection - allows selecting which probe modules to include
  moduleNamespaceSelector:
    matchLabels:
        kubernetes.io/metadata.name: monitoring
  moduleSelector:
    matchLabels:
      type: http-probe
  
  # Resource requirements
  resources:
    limits:
      memory: 200Mi
      cpu: 100m
    requests:
      memory: 100Mi
      cpu: 50m
      
  # Advanced configuration options
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
  
  # Network configuration
  webListenAddresses:
    - ":9115"
  externalURL: "https://blackbox.example.com"
  
status:
  # Status fields will be managed by the controller
  replicas: 2
  availableReplicas: 2
  updatedReplicas: 2
  unavailableReplicas: 0
  conditions:
    - type: Available
      status: "True"
      lastTransitionTime: "2025-02-28T12:00:00Z"
      reason: MinimumReplicasAvailable
      message: "All replicas are available"

```
### 2. BlackBoxExporterModule CRD
```
---
apiVersion: monitoring.zhipu.ai/v1alpha1
kind: BlackboxExporterModule
metadata:
  name: example-http-probe
  labels:
    type: http-probe
spec:
  # HTTP probe configuration
  # Common probe settings
  timeout: 5s
  prober: http
  http:
    # Request definition
    method: POST
    headers:
      Accept: "application/json"
    body: |
      {
        "name": "example",
        "value": 42,
        "active": true,
        "metadata": {
          "source": "kubernetes-crd",
          "version": "1.0"
        }
      }
    # Validation options
    validStatusCodes: [200, 301, 302]
    validHTTPVersions: ["HTTP/1.1", "HTTP/2.0"]
    failIfSSL: false
    failIfNotSSL: true
    # Response validation
    failIfBodyMatchesRegexp:
      - "Error|Failure"
    failIfBodyNotMatchesRegexp:
      - "Success"
```
