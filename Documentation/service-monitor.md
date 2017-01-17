# ServiceMonitor

The `ServiceMonitor` third party resource (TPR) allows to declaratively define
how a dynamic set of services should be monitored. Which services are selected
to be monitored with the desired configuration is defined using label selections.
This allows to dynamically express monitoring without having to update additional
configuration for services that follow known monitoring patterns.

A service may expose one or more service ports, which are backed by a list
of multiple endpoints that point to a pod in the common case.

In the `endpoints` section of the TPR, we can configure which ports of these
endpoints we want to scrape for metrics and with which paramters. For advanced use
cases one may want to monitor ports of backing pods, which are not directly part
of the service endpoints. This is also made possible by the Prometheus Operator.


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

### `TLSConfig`

| Name | Description | Required | Schema | Default |
| ---- | ----------- | -------- | ------ | ------- |
| caFile | Path to the CA file. | false | string | |
| certFile | Path to client certificate file | false | |
| keyFile | Path to client key file | false | |
| serverName | Server name used to verify host name | |
| insecureSkipVerify | Skip certificate verification | false | bool | false |


## Current state and roadmap

### Namespaces

While `ServiceMonitor`s must live in the same namespace as the `Prometheus` TPR,
discovered targets may come from any namespace. This is important to allow cross-namespace
monitoring use cases, e.g. for meta-monitoring.

Currently, targets are always discovered from all namespaces. In the future, the
`ServiceMonitor` should allow to restrict this to one or more namespaces.
How such a configuration would look like, i.e. explicit namespaces, selection by labels,
or both, and what the default behavior should be is still up for discussion.
