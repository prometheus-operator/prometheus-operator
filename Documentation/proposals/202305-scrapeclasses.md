## Scrape Classes

* **Owners:**
  * `nicolastakashi`
  * `eronwright`
* **Status:**
  * `Implemented`
* **Related Tickets:**
  * https://github.com/prometheus-operator/prometheus-operator/issues/4121
  * https://github.com/prometheus-operator/prometheus-operator/issues/3922
  * https://github.com/prometheus-operator/prometheus-operator/issues/5947
  * https://github.com/prometheus-operator/prometheus-operator/issues/5948

This proposal introduces the concept of *scrape classes*, enabling users to utilize scrape configuration data provided by the administrator, through scrape objects such as PodMonitor, Probe, ServiceMonitor and ScrapeConfig.

## Why

Sometimes Prometheus administrators needs to provide default configurations for scrape objects, such as the definition of TLS certificates, when running Prometheus to scrape pods in an Istio mesh with strict mTLS as described in [Istio documentation](https://istio.io/latest/docs/ops/integrations/prometheus/#tls-settings).

Another motivation is to improve feature parity amongst the monitor resources. The `PodMonitor` and `Probe` resources aren't at parity with `ServiceMonitor` because the latter allows for unsafe TLS settings.

### Pitfalls of the current solution

The only known solution for use cases where you'd need to use unsafe TLS settings is to use `additionalScrapeConfig`.

The downside is the loss of integration provided by Prometheus Operator through monitor resources to compose the scrape configurations in a Kubernetes way.

## Goals

- Allow for the administrator to define a named, reusable scrape configuration snippet, including unsafe elements such as file references.
- Allow for a user to select a configuration snippet by name in their probe/podmonitor/servicemonitor endpoint configuration.
- Avoid giving the user the ability to exflitrate arbitrary files within the Prometheus pod.

## Non-Goals

- Provide a way for Prometheus administrators to enforce/override settings on scrape configurations.
- Deprecation of the unsafe TLS settings in `ServiceMonitor`.

### Audience

- Users who serve Prometheus as a service and want to give their customers autonomy in defining monitors, but want to provide a default configuration for scraping.

## How

The proposed solution is to introduce a notion of a *scrape class*, akin to a Kubernetes [storage class](https://kubernetes.io/docs/concepts/storage/storage-classes/).
A scrape class defines TLS settings (and possibly other settings in future, e.g. sensitive authorization settings) to be applied to all scrape resources (ServiceMonitor, PodMonitor, Probe and ScrapeConfig) of that class.

One scrape class may be designated as the default class, in which case that class is applied to any scrape resource that doesn't specify a value for `scrapeClassName`.

When defining a podmonitor/servicemonitor/probe/scrapeconfig, a user may assign a scrape class via the `scrapeClassName field.
When there's a match, a scrape class is assigned to all the endpoints.

Class names are assumed to be installation-specific. In practice, some common class names like `istio-mtls` are likely to emerge.

### Prometheus Resource

It is proposed that the `Prometheus` and `PrometheusAgent` resources contain a new field for defining scrape classes.

The rationale for defining scrape classes inline is that, in practice, the TLS file paths are closely related to the `volumeMounts`
of the `Prometheus` spec. An alternative is outlined later, of factoring the class definitions into a separate resource.

One (and only one) scrape class may be designated as the default class.

When a resource defines several default scrape classes, it should fail the reconciliation.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
spec:
  scrapeClasses:
    - name: istio-mtls
      default: true
      tlsConfig:
        caFile: "/etc/istio-certs/root-cert.pem"
        certFile: "/etc/istio-certs/cert-chain.pem"
        keyFile: "/etc/istio-certs/key.pem"
        insecureSkipVerify: true

  # mount the certs from the istio sidecar (shown here for illustration purposes)
  volumeMounts:
    - name: istio-certs
      mountPath: "/etc/istio-certs/"
  volumes:
  - emptyDir:
      medium: Memory
    name: istio-certs
```

Any object references in the scrape class definition are assumed to refer to objects in the namespace of the `Prometheus` object.

### Monitoring Resource

If the monitor resource specifies a scrape class name that isn't defined in the Prometheus/PrometheusAgent object, then the scrape resource is rejected by the operator.

This behavior is consistent with the behavior of monitor resources referencing a non-existing secret for bearer token authentication.

To ensure users will have proper information about the error, the operator may (in the future) emit an event with the error message on the monitor resource and also update the status of the monitor resource with the error message.

### PodMonitor Resource

Allow the user to select a scrape class which applies to all endpoints.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
spec:
  scrapeClassName: istio-mtls
  podMetricsEndpoints:
  - port: http
    path: /metrics
```

If the `Monitor` resource has a `tlsConfig` field defined, the Operator will use a merge strategy to combine the `tlsConfig` fields from the PodMonitor object with the `tlsConfig` fields of the scrape class, the `tlsConfig` fields in the `PodMonitor` resource take precedence.

### Probe Resource

Allow the user to select a scrape class for the prober service.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Probe
spec:
  scrapeClassName: istio-mtls
```

### ServiceMonitor Resource

Allow the user to select a scrape class for all endpoints.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
spec:
  scrapeClassName: istio-mtls
  endpoints:
  - port: http
    path: /metrics
```

### ScrapeConfig

Allow the user to select a scrape class for the whole scrape configuration.

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: ScrapeConfig
metadata:
  name: scrape-config
spec:
  scrapeClassName: istio-mtls
  staticConfigs:
    [...]
  httpSDConfig:
    [...]
  fileSDConfig:
    [...]
```

## Test Plan

1. Regression test; ensure no change to generated configs unless a scrape class is applied.
2. Optionality of the default scrape class; ensure that it is an optional configuration element.
3. Acceptance testing in Istio environment; ensure that the solution is effective for a key use case.

## Alternatives

### Global scrape TLS configuration

An alternative solution would be to apply a default TLS configuration to all monitors.

For example, via a hypothetical field `spec.scrapeTlsConfig`:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
spec:
  scrapeTlsConfig:
    caFile: "/etc/istio-certs/root-cert.pem"
    certFile: "/etc/istio-certs/cert-chain.pem"
    keyFile: "/etc/istio-certs/key.pem"
    insecureSkipVerify: true
```

Objections:
1. A singular default configuration may be too inflexible to effectively scrape a diverse set of pods. For example, in Istio,
   some pods may be in STRICT mode.

### Istio Permissive Mode

An alternative for the Istio use case is to use PERMISSIVE mode (see [documentation](https://istio.io/latest/docs/concepts/security/#permissive-mode)),
or to use `exclude` annotations on the pod such that the metrics endpoint bypasses mTLS.

Objections:
1. This solution is less secure at a transport level unless TLS is implemented at the application layer.
2. Strict mTLS is the basis for [Istio authorization policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/), e.g. to explicitly allow access from prom to the metrics endpoint.

### ScrapeClass Resource

A variant of the proposed solution is to introduce a new custom resource for defining scrape classes.

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: ScrapeClass
metadata:
  name: istio-mtls
spec:
  tlsConfig:
    caFile: "/etc/istio-certs/root-cert.pem"
    certFile: "/etc/istio-certs/cert-chain.pem"
    keyFile: "/etc/istio-certs/key.pem"
    insecureSkipVerify: true
```

An open question is whether the resource would be cluster-scoped or namespace-scoped.

Objections:

1. Since the file paths are dependent on the volume mounts in the server, this approach may not achieve a meaningful decoupling.
2. Extra complexity in defining a new CRD.

### Non-Safe Monitors

Another alternative would be to allow `PodMonitor` and `Probe` to use unsafe TLS settings.

Objections:
1. See [explanation](https://github.com/prometheus-operator/prometheus-operator/issues/3922#issuecomment-802899950) for why
   unsafe settings were disallowed in the first place.

## Action Plan

* [ ] Change the Operator API to define scrape classes and for scrape resources to select a scrape class
* [ ] Update the scrape configuration generator to use the scrape class
* [ ] Update documentation
* [ ] Resolve issues #4121, #3922
