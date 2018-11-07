local kp =
  (import 'kube-prometheus/kube-prometheus.libsonnet') +
  (import 'kube-prometheus/kube-prometheus-kubeadm.libsonnet') +
  // Note that NodePort type services is likely not a good idea for your production use case, it is only used for demonstration purposes here.
  (import 'kube-prometheus/kube-prometheus-node-ports.libsonnet') +
  {
    _config+:: {
      namespace: 'monitoring',
      alertmanager+:: {
        config: importstr 'alertmanager-config.yaml',
      },
      grafana+:: {
        config: {  // http://docs.grafana.org/installation/configuration/
          sections: {
            // Do not require grafana users to login/authenticate
            'auth.anonymous': { enabled: true },
          },
        },
      },
    },

    // For simplicity, each of the following values for 'externalUrl':
    //  * assume that `minikube ip` prints "192.168.99.100"
    //  * hard-code the NodePort for each app
    prometheus+:: {
      prometheus+: {
        // Reference info: https://coreos.com/operators/prometheus/docs/latest/api.html#prometheusspec
        spec+: {
          // An e.g. of the purpose of this is so the "Source" links on http://<alert-manager>/#/alerts are valid.
          externalUrl: 'http://192.168.99.100:30900',

          // Reference info: "external_labels" on https://prometheus.io/docs/prometheus/latest/configuration/configuration/
          externalLabels: {
            // This 'cluster' label will be included on every firing prometheus alert. (This is more useful
            // when running multiple clusters in a shared environment (e.g. AWS) with other users.)
            cluster: 'minikube-<INSERT YOUR USERNAME HERE>',
          },
        },
      },
    },
    alertmanager+:: {
      alertmanager+: {
        // Reference info: https://github.com/coreos/prometheus-operator/blob/master/Documentation/api.md#alertmanagerspec
        spec+: {
          externalUrl: 'http://192.168.99.100:30903',

          logLevel: 'debug',  // So firing alerts show up in log
        },
      },
    },
  };

{ ['00namespace-' + name]: kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus) } +
{ ['0prometheus-operator-' + name]: kp.prometheusOperator[name] for name in std.objectFields(kp.prometheusOperator) } +
{ ['node-exporter-' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
{ ['kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
{ ['alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
{ ['prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) } +
{ ['grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) }
