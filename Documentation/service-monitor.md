# ServiceMonitor

The `ServiceMonitor` third party resource (TPR) allows to declaratively define
how a dynamic set of services should be monitored. Which services are selected
to be monitored with the desired configuration is defined using label
selections. This allows an organization to introduce conventions around how
metrics are exposed, and then following these conventions new services are
automatically discovered, without the need to reconfigure the system.

## Design

For Prometheus to monitor any application within Kubernetes an `Endpoints`
object needs to exist. `Endpoints` objects are essentially lists of IP
addresses. Typically an `Endpoints` object is populated by a `Service` object.
A `Service` object discovers `Pod`s by a label selector and adds those to the
`Endpoints` object.

A `Service` may expose one or more service ports, which are backed by a list of
multiple endpoints that point to a `Pod` in the common case. This is reflected
in the respective `Endpoints` object as well.

The `ServiceMonitor` object introduced by the Prometheus Operator in turn
discovers those `Endpoints` objects and configures Prometheus to monitor those
`Pod`s.

The `endpoints` section of the `ServiceMonitorSpec`, is used to configure which
ports of these `Endpoints` are going to be scraped for metrics, and with which
parameters. For advanced use cases one may want to monitor ports of backing
`Pod`s, which are not directly part of the service endpoints. Therefore when
specifying an endpoint in the `endpoints` section, they are strictly used.

> Note: `endpoints` (lowercase) is the TPR field, while `Endpoints`
> (capitalized) is the Kubernetes object kind.

While `ServiceMonitor`s must live in the same namespace as the `Prometheus`
TPR, discovered targets may come from any namespace. This is important to allow
cross-namespace monitoring use cases, e.g. for meta-monitoring. Using the
`namespaceSelector` of the `ServiceMonitorSpec`, one can restrict the
namespaces the `Endpoints` objects are allowed to be discovered from.

## Specification

### `ServiceMonitor`

| Name | Description | Required | Schema | Default |
| ---- | ----------- | -------- | ------ | ------- |
| spec | Specification of the ServiceMonitor object. | true | ServiceMonitorSpec | |

### `ServiceMonitorSpec`

| Name | Description | Required | Schema | Default |
| ---- | ----------- | -------- | ------ | ------- |
| jobLabel | Service label of which the value is used to assemble a job name of the form `<label value>-<port>`. If no label is specified, the service name is used. | false | string |  |
| selector | Label selector for services the `ServiceMonitor` applies to. | true | [unversioned.LabelSelector](http://kubernetes.io/docs/api-reference/v1/definitions/#_unversioned_labelselector) | |
| namespaceSelector | Namespaces from which services are selected. | false | NamespaceSelector | same namespace only |
| endpoints | The endpoints to be monitored for endpoints of the selected services. | true | Endpoint array | |

### `Endpoint`

| Name | Description | Required | Schema | Default |
| ---- | ----------- | -------- | ------ | ------- |
| port | Name of the service port this endpoint refers to. Mutually exclusive with targetPort. | false | string | |
| targetPort | Name or number of the target port of the endpoint. Mutually exclusive with port. | false | integer or string | |
| path | HTTP path to scrape for metrics. | false | string | /metrics |
| scheme | HTTP scheme to use for scraping | false | string | http |
| interval | Interval at which metrics should be scraped | false | duration | 30s |
| tlsConfig | TLS configuration to use when scraping the endpoint | false | TLSConfig | |
| bearerTokenFile | File to read bearer token for scraping targets. | false | string | |

### `TLSConfig`

| Name | Description | Required | Schema | Default |
| ---- | ----------- | -------- | ------ | ------- |
| caFile | Path to the CA file. | false | string | |
| certFile | Path to client certificate file | false | |
| keyFile | Path to client key file | false | |
| serverName | Server name used to verify host name | |
| insecureSkipVerify | Skip certificate verification | false | bool | false |

### `NamespaceSelector`

| Name | Description | Required | Schema | Default |
| ---- | ----------- | -------- | ------ | ------- |
| any | Match any namespace | false | bool | false |
| matchNames | Explicit list of namespace names to select | false | string array | |

