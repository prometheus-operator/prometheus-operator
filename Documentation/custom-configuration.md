<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> Starting with v0.12.0, Prometheus Operator requires use of Kubernetes v1.7.x and up.
</div>


**Deprecation Warning:** The _custom configuration_ option of the Prometheus Operator will be deprecated in favor of the [_additional scrape config_](./additional-scrape-config.md) option.


# Custom Configuration

There are a few reasons, why one may want to provide a custom configuration to Prometheus instances, instead of having the Prometheus Operator generate the configuration based on `ServiceMonitor` objects.

> Note that custom configurations are not the primary goal the Prometheus Operator is solving. Changes made in the Prometheus Operator, may affect custom configurations in a breaking way. Additionally the Prometheus Operator attempts to generate configurations in a forward and backward compatible way, with custom configurations, this is up to the user to manage gracefully.

Use cases include:

* The necessity to use a service discovery mechanism other than the Kubernetes service discovery, such as AWS SD, Azure SD, etc.
* Cases that are not (yet) very well supported by the Prometheus Operator, such as performing blackbox probes.

Note that because the Prometheus Operator does not generate the Prometheus configuration in this case, any fields of the Prometheus resource, which influence the configuration will have no effect, and one has to specify this explicitly. The features that will not be supported, meaning they will have to be configured manually:

* `serviceMonitorSelector`: Auto-generating Prometheus configuration from `ServiceMonitor` objects. This means, that creating `ServiceMonitor` objects is not how a Prometheus instance is configured, but rather the raw configuration has to be written.
* `alerting`: Alertmanager discovery as available in the Prometheus object is translated to the Prometheus configuration, meaning this configuration has to be done manually.
* `scrapeInterval`
* `evaluationInterval`
* `externalLabels`

In order to enable to specify a custom configuration, the `serviceMonitorSelector` field has to be left empty. When the `serviceMonitorSelector` field is empty, the Prometheus Operator will not attempt to manage the `Secret`, that contains the Prometheus configuration. The `Secret`, that contains the Prometheus configuration is called `prometheus-<name-of-prometheus-object>`, in the same namespace as the Prometheus object. Within this `Secret`, the key that contains the Prometheus configuration is called `prometheus.yaml`.
