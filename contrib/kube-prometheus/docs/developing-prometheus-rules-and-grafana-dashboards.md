# Developing Prometheus Rules and Grafana Dashboards

`kube-prometheus` ships with a set of default [Prometheus rules](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/) and [Grafana](http://grafana.com/) dashboards. At some point one might like to extend them, the purpose of this document is to explain how to do this.

All manifests of kube-prometheus are generated using [jsonnet](https://jsonnet.org/) and Prometheus rules and Grafana dashboards in specific follow the [Prometheus Monitoring Mixins proposal](https://docs.google.com/document/d/1A9xvzwqnFVSOZ5fD3blKODXfsat5fg6ZhnKu9LK3lB4/).

For both the Prometheus rules and the Grafana dashboards Kubernetes `ConfigMap`s are generated within kube-prometheus. In order to add additional rules and dashboards simply merge them onto the existing json objects. This document illustrates examples for rules as well as dashboards.

As a basis, all examples in this guide are based on the base example of the kube-prometheus [readme](../README.md):

[embedmd]:# (../example.jsonnet)
```jsonnet
local kp =
  (import 'kube-prometheus/kube-prometheus.libsonnet') + {
    _config+:: {
      namespace: 'monitoring',
    },
  };

{ ['00namespace-' + name]: kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus) } +
{ ['0prometheus-operator-' + name]: kp.prometheusOperator[name] for name in std.objectFields(kp.prometheusOperator) } +
{ ['node-exporter-' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
{ ['kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
{ ['alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
{ ['prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) } +
{ ['prometheus-adapter-' + name]: kp.prometheusAdapter[name] for name in std.objectFields(kp.prometheusAdapter) } +
{ ['grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) }

```

## Prometheus rules

### Alerting rules

According to the [Prometheus Monitoring Mixins proposal](https://docs.google.com/document/d/1A9xvzwqnFVSOZ5fD3blKODXfsat5fg6ZhnKu9LK3lB4/) Prometheus alerting rules are under the key `prometheusAlerts` in the top level object, so in order to add an additional alerting rule, we can simply merge an extra rule into the existing object.

The format is exactly the Prometheus format, so there should be no changes necessary should you have existing rules that you want to include.

> Note that alerts can just as well be included into this file, using the jsonnet `import` function. In this example it is just inlined in order to demonstrate their use in a single file.

[embedmd]:# (../examples/prometheus-additional-alert-rule-example.jsonnet)
```jsonnet
local kp = (import 'kube-prometheus/kube-prometheus.libsonnet') + {
  _config+:: {
    namespace: 'monitoring',
  },
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'example-group',
        rules: [
          {
            alert: 'Watchdog',
            expr: 'vector(1)',
            labels: {
              severity: 'none',
            },
            annotations: {
              description: 'This is a Watchdog meant to ensure that the entire alerting pipeline is functional.',
            },
          },
        ],
      },
    ],
  },
};

{ ['00namespace-' + name]: kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus) } +
{ ['0prometheus-operator-' + name]: kp.prometheusOperator[name] for name in std.objectFields(kp.prometheusOperator) } +
{ ['node-exporter-' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
{ ['kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
{ ['alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
{ ['prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) } +
{ ['grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) }
```

### Recording rules

In order to add a recording rule, simply do the same with the `prometheusRules` field.

> Note that rules can just as well be included into this file, using the jsonnet `import` function. In this example it is just inlined in order to demonstrate their use in a single file.

[embedmd]:# (../examples/prometheus-additional-recording-rule-example.jsonnet)
```jsonnet
local kp = (import 'kube-prometheus/kube-prometheus.libsonnet') + {
  _config+:: {
    namespace: 'monitoring',
  },
  prometheusRules+:: {
    groups+: [
      {
        name: 'example-group',
        rules: [
          {
            record: 'some_recording_rule_name',
            expr: 'vector(1)',
          },
        ],
      },
    ],
  },
};

{ ['00namespace-' + name]: kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus) } +
{ ['0prometheus-operator-' + name]: kp.prometheusOperator[name] for name in std.objectFields(kp.prometheusOperator) } +
{ ['node-exporter-' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
{ ['kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
{ ['alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
{ ['prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) } +
{ ['grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) }
```

### Pre-rendered rules

We acknowledge, that users may need to transition existing rules, and therefore allow an option to add additional pre-rendered rules. Luckily the yaml and json formats are very close so the yaml rules just need to be converted to json without any manual interaction needed. Just a tool to convert yaml to json is needed:

```
go get -u -v github.com/brancz/gojsontoyaml
```

And convert the existing rule file:

```
cat existingrule.yaml | gojsontoyaml -yamltojson > existingrule.json
```

Then import it in jsonnet:

[embedmd]:# (../examples/prometheus-additional-rendered-rule-example.jsonnet)
```jsonnet
local kp = (import 'kube-prometheus/kube-prometheus.libsonnet') + {
  prometheusAlerts+:: (import 'existingrule.json'),
};

{ ['00namespace-' + name]: kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus) } +
{ ['0prometheus-operator-' + name]: kp.prometheusOperator[name] for name in std.objectFields(kp.prometheusOperator) } +
{ ['node-exporter-' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
{ ['kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
{ ['alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
{ ['prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) } +
{ ['grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) }
```
### Changing default rules

Along with adding additional rules, we give the user the option to filter or adjust the existing rules imported by `kube-prometheus/kube-prometheus.libsonnet`. The recording rules can be found in [kube-prometheus/rules](https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus/jsonnet/kube-prometheus/rules) and [kubernetes-mixin/rules](https://github.com/kubernetes-monitoring/kubernetes-mixin/tree/master/rules) while the alerting rules can be found in [kube-prometheus/alerts](https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus/jsonnet/kube-prometheus/alerts) and [kubernetes-mixin/alerts](https://github.com/kubernetes-monitoring/kubernetes-mixin/tree/master/alerts).

Knowing which rules to change, the user can now use functions from the [Jsonnet standard library](https://jsonnet.org/ref/stdlib.html) to make these changes. Below are examples of both a filter and an adjustment being made to the default rules. These changes can be assigned to a local variable and then added to the `local kp` object as seen in the examples above.  

#### Filter
Here the alert `KubeStatefulSetReplicasMismatch` is being filtered out of the group `kubernetes-apps`. The default rule can be seen [here](https://github.com/kubernetes-monitoring/kubernetes-mixin/blob/master/alerts/apps_alerts.libsonnet).
```jsonnet
local filter = {
  prometheusAlerts+:: {
    groups: std.map(
      function(group)
        if group.name == 'kubernetes-apps' then
          group {
            rules: std.filter(function(rule)
              rule.alert != "KubeStatefulSetReplicasMismatch",
              group.rules
            )
          }
        else
          group,
      super.groups
    ),
  },
};
```
#### Adjustment
Here the expression for the alert used above is updated from its previous value. The default rule can be seen [here](https://github.com/kubernetes-monitoring/kubernetes-mixin/blob/master/alerts/apps_alerts.libsonnet).
```jsonnet
local update = {
  prometheusAlerts+:: {
    groups: std.map(
      function(group)
        if group.name == 'kubernetes-apps' then
          group {
            rules: std.map(
              function(rule)
                if rule.alert == "KubeStatefulSetReplicasMismatch" then
                  rule {
                    expr: "kube_statefulset_status_replicas_ready{job=\"kube-state-metrics\",statefulset!=\"vault\"} != kube_statefulset_status_replicas{job=\"kube-state-metrics\",statefulset!=\"vault\"}"
                  }
                else
                  rule,
                group.rules
            )
          }
        else
          group,
      super.groups
    ),
  },
};
```
Using the example from above about adding in pre-rendered rules, the new local vaiables can be added in as follows:
```jsonnet
local kp = (import 'kube-prometheus/kube-prometheus.libsonnet') + filter + update + {
    prometheusAlerts+:: (import 'existingrule.json'),
};

{ ['00namespace-' + name]: kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus) } +
{ ['0prometheus-operator-' + name]: kp.prometheusOperator[name] for name in std.objectFields(kp.prometheusOperator) } +
{ ['node-exporter-' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
{ ['kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
{ ['alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
{ ['prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) } +
{ ['prometheus-adapter-' + name]: kp.prometheusAdapter[name] for name in std.objectFields(kp.prometheusAdapter) } +
{ ['grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) }
``` 
## Dashboards

Dashboards can either be added using jsonnet or simply a pre-rendered json dashboard.

### Jsonnet dashboard

We recommend using the [grafonnet](https://github.com/grafana/grafonnet-lib/) library for jsonnet, which gives you a simple DSL to generate Grafana dashboards. Following the [Prometheus Monitoring Mixins proposal](https://docs.google.com/document/d/1A9xvzwqnFVSOZ5fD3blKODXfsat5fg6ZhnKu9LK3lB4/) additional dashboards are added to the `grafanaDashboards` key, located in the top level object. To add new jsonnet dashboards, simply add one.

> Note that dashboards can just as well be included into this file, using the jsonnet `import` function. In this example it is just inlined in order to demonstrate their use in a single file.

[embedmd]:# (../examples/grafana-additional-jsonnet-dashboard-example.jsonnet)
```jsonnet
local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local row = grafana.row;
local prometheus = grafana.prometheus;
local template = grafana.template;
local graphPanel = grafana.graphPanel;

local kp = (import 'kube-prometheus/kube-prometheus.libsonnet') + {
  _config+:: {
    namespace: 'monitoring',
  },
  grafanaDashboards+:: {
    'my-dashboard.json':
      dashboard.new('My Dashboard')
      .addTemplate(
        {
          current: {
            text: 'Prometheus',
            value: 'Prometheus',
          },
          hide: 0,
          label: null,
          name: 'datasource',
          options: [],
          query: 'prometheus',
          refresh: 1,
          regex: '',
          type: 'datasource',
        },
      )
      .addRow(
        row.new()
        .addPanel(graphPanel.new('My Panel', span=6, datasource='$datasource')
                  .addTarget(prometheus.target('vector(1)')))
      ),
  },
};

{ ['00namespace-' + name]: kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus) } +
{ ['0prometheus-operator-' + name]: kp.prometheusOperator[name] for name in std.objectFields(kp.prometheusOperator) } +
{ ['node-exporter-' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
{ ['kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
{ ['alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
{ ['prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) } +
{ ['grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) }
```

### Pre-rendered Grafana dashboards

As jsonnet is a superset of json, the jsonnet `import` function can be used to include Grafana dashboard json blobs. In this example we are importing a [provided example dashboard](../examples/example-grafana-dashboard.json).

[embedmd]:# (../examples/grafana-additional-rendered-dashboard-example.jsonnet)
```jsonnet
local kp = (import 'kube-prometheus/kube-prometheus.libsonnet') + {
  _config+:: {
    namespace: 'monitoring',
  },
  grafanaDashboards+:: {
    'my-dashboard.json': (import 'example-grafana-dashboard.json'),
  },
};

{ ['00namespace-' + name]: kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus) } +
{ ['0prometheus-operator-' + name]: kp.prometheusOperator[name] for name in std.objectFields(kp.prometheusOperator) } +
{ ['node-exporter-' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
{ ['kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
{ ['alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
{ ['prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) } +
{ ['grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) }
```
