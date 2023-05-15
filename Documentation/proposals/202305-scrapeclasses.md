## Scrape Classes

* **Owners:**
  * `eronwright`

* **Related Tickets:**
  * https://github.com/prometheus-operator/prometheus-operator/issues/4121
  * https://github.com/prometheus-operator/prometheus-operator/issues/3922

This proposal aims to introduce the concept of *scrape classes*, to enable users to leverage scrape configuration data
that is provided by the administrator. For example, to allow pod monitors and probes to safely use
unsafe TLS configuration fields (e.g. `keyFile`).

## Why

Today, there are some scenarios that are difficult to support in a safe way, notably the need to reference
a local TLS file in a scrape configuration. For example, to scrape a pod in an Istio mesh with strict mTLS would require that
the scrape configuration use a TLS keyfile that is provided by the Prometheus server's Istio sidecar.
See [Istio documentation](https://istio.io/latest/docs/ops/integrations/prometheus/#tls-settings) for more details.

Another motivation is to improve feature parity amongst the monitor resources. The `PodMonitor` and `Probe` resources
aren't at parity with `ServiceMonitor` because the latter allows for unsafe TLS settings.

Yet another motivation is to decouple the monitor spec from low-level infrastructure details like TLS certificates.

### Pitfalls of the current solution

The only known solution for use cases where you'd need to use unsafe TLS settings is to use `additionalScrapeConfig`.
The downside is obviously the loss of a great feature of the Prometheus Operator. The monitor resources make it possible
to compose the scrape configurations in a Kubernetes way.

## Goals

- Allow for the administrator to define a named, reusable scrape configuration snippet, including unsafe elements such as file references.
- Allow for a user to select a configuration snippet by name in their probe/podmonitor/servicemonitor endpoint configuration.
- Avoid giving the user the ability to use arbirary files within the Prometheus pod.

### Audience

- Users who serve Prometheus as a service and want to give their customers autonomy in defining monitors.
- Users who want to scrape pods within an Istio mesh.

## How

The proposed solution is to introduce a notion of a *scrape class*, akin to a Kubernetes [storage class](https://kubernetes.io/docs/concepts/storage/storage-classes/).
A scrape class defines TLS settings (and possibly other settings in future) to be applied to all scrape endpoints of that class.

When defining a probe/podmonitor/servicemonitor, a user may optionally assign a class to each endpoint.
A default class may be defined by the administrator.

Class names are assumed to be installation-specific. In practice, some common class names like `istio-mtls` are likely to emerge.

### Prometheus Resource

It is proposed that the `Prometheus` resource contain a new section for defining scrape classes.

The rationale for defining scrape classes inline is that, in practice, the TLS file paths are closely related to the `volumeMounts`
of the `Prometheus` spec. An alternative is outlined later, of factoring the class definitions into a separate resource.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
spec:
  # optional: set a default scrape class
  defaultScrapeClass: istio-mtls

  # define scrape classes for use by the monitors
  scrapeClasses:
    - name: istio-mtls
      tlsConfig:
        caFile: "/etc/istio-certs/root-cert.pem"
        certFile: "/etc/istio-certs/cert-chain.pem"
        keyFile: "/etc/istio-certs/key.pem"
        insecureSkipVerify: true

  # mount the certs from the istio sidecar (shown here for illustration purposes)
  volumeMounts:
    - name: istio-certs
      mountPath: "/etc/istio-certs/"
```

Any object references in the scrape class definition are assumed to refer to objects in the namespace of the `Prometheus` object.

### PodMonitor Resource

Allow the user to select a scrape class for each endpoint.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
spec:
  podMetricsEndpoints:
  - port: http
    path: /metrics
    scrapeClass: istio-mtls
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
  endpoints:
  - port: http
    path: /metrics
    scrapeClass: istio-mtls
```

Deprecate the unsafe TLS settings in `ServiceMonitor`.

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
