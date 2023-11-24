## Scrape Classes

* **Owners:**
  * `nicolastakashi`
  * `eronwright`

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
The downside is obviously the loss of a great feature of the Prometheus Operator. The monitor resources make it possible
to compose the scrape configurations in a Kubernetes way.

## Goals

- Allow for the administrator to define a named, reusable scrape configuration snippet, including unsafe elements such as file references.
- Allow for a user to select a configuration snippet by name in their probe/podmonitor/servicemonitor endpoint configuration.
- Avoid giving the user the ability to use arbirary files within the Prometheus pod.

## Non-Goals

- Allow Prometheus owners to override the configuration defined in the monitoring resources (at least in the first iteration).

### Audience

- Users who serve Prometheus as a service and want to give their customers autonomy in defining monitors, but want to provide a default configuration for scraping.

## How

The proposed solution is to introduce a notion of a *scrape class*, akin to a Kubernetes [storage class](https://kubernetes.io/docs/concepts/storage/storage-classes/).
A scrape class defines TLS settings (and possibly other settings in future, e.g. sensitive authorization settings) to be applied to all scrape configs of that class.

One scrape class may be designated as the default class, in which case that class is applied to any scrape config that doesn't specify a value for `scrapeClass`.

When defining a podmonitor/servicemonitor/probe/scrapeconfig, a user may assign a scrape class via the `scrapeClass` field.
When there's a match, a scrape class is assigned to all the endpoints.

Class names are assumed to be installation-specific. In practice, some common class names like `istio-mtls` are likely to emerge.

### Prometheus Resource

It is proposed that the `Prometheus` and `PrometheusAgent` resources contain a new section for defining scrape classes.

The rationale for defining scrape classes inline is that, in practice, the TLS file paths are closely related to the `volumeMounts`
of the `Prometheus` spec. An alternative is outlined later, of factoring the class definitions into a separate resource.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
spec:
  # define scrape classes for use by the monitors
  # one class may be designated as the default class
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
    - name: istio-certs
      secret:
        secretName: istio-certs
```

Any object references in the scrape class definition are assumed to refer to objects in the namespace of the `Prometheus` object.

### PodMonitor Resource

Allow the user to select a scrape class which applies to all endpoints.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
spec:
  scrapeClass: istio-mtls
  podMetricsEndpoints:
  - port: http
    path: /metrics
```

The proposed behavior is:
1. the `tlsConfig` in the associated scrape class is automatically applied to the scrape configuration of the endpoint.
2. the inline `tlsConfig` (if any) takes precedence over the `tlsConfig` in the scrape class.

### Probe Resource

Allow the user to select a scrape class for the probe.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Probe
spec:
  scrapeClass: istio-mtls
```

### ServiceMonitor Resource

Allow the user to select a scrape class for each endpoint.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
spec:
  scrapeClass: istio-mtls
  endpoints:
  - port: http
    path: /metrics
```

Out-of-scope: deprecation of the unsafe TLS settings in `ServiceMonitor`.

### ScrapeConfig

Allow the user to select a scrape class for the generic scrape configuration.

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: ScrapeConfig
metadata:
  name: scrape-config
spec:
  scrapeClass: istio-mtls
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

* [ ] Change the Operator API to define scrape classes and for endpoint to select a scrape class
* [ ] Update the scrape configuration generator to use the scrape class
* [ ] Update documentation
* [ ] Resolve issues #4121, #3922