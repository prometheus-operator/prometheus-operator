local kp = (import 'kube-prometheus/kube-prometheus.libsonnet') +
           (import 'kube-prometheus/kube-prometheus-static-etcd.libsonnet') + {
  _config+:: {
    namespace: 'monitoring',

    // Reference info: https://github.com/coreos/prometheus-operator/blob/master/contrib/kube-prometheus/README.md#static-etcd-configuration
    etcd+:: {
      // Configure this to be the IP(s) to scrape - i.e. your etcd node(s) (use commas to separate multiple values).
      ips: ['127.0.0.1'],

      // Reference info:
      //  * https://github.com/coreos/prometheus-operator/blob/master/Documentation/api.md#servicemonitorspec (has endpoints)
      //  * https://github.com/coreos/prometheus-operator/blob/master/Documentation/api.md#endpoint (has tlsConfig)
      //  * https://github.com/coreos/prometheus-operator/blob/master/Documentation/api.md#tlsconfig (has: caFile, certFile, keyFile, serverName, & insecureSkipVerify)

      // Set these three variables to the fully qualified directory path on your work machine to the certificate files that are valid to scrape etcd metrics with (check the apiserver container).
      // Most likely these certificates are generated somewhere in an infrastructure repository, so using the jsonnet `importstr` function can
      // be useful here. (Kube-aws stores these three files inside the credential folder.)
      // All the sensitive information on the certificates will end up in a Kubernetes Secret.
      clientCA: importstr 'etcd-client-ca.crt',
      clientKey: importstr 'etcd-client.key',
      clientCert: importstr 'etcd-client.crt',

      // A valid name (DNS or Subject Alternative Name) that the client (i.e. prometheus) will use to verify the etcd TLS certificate.
      serverName: 'etcd.my-cluster.local', // a real-life e.g. value is "etcd.kube-system.svc.cluster.local"

      // TODO: prometheus-operator issue #1755 ("kube-prometheus: re-introduce insecureSkipVerify for etcd monitoring") will enhance kube-prometheus-static-etcd.libsonnet
      //       to allow 'insecureSkipVerify: true' to be specified here (as an alternative to specifying a value for 'serverName').
      // Note that once that issue #1755 is resolved, insecureSkipVerify is only to be used if you cannot use a Subject Alternative Name.

      // In case you have generated the etcd certificate with kube-aws:
      //  * If you only have one etcd node, you can use the value from 'etcd.internalDomainName' (specified in your kube-aws cluster.yaml) as the value for 'serverName'.
      //  * But if you have multiple etcd nodes, you will need to use 'insecureSkipVerify: true' (if using default certificate generators method), as the valid certificate domain
      //    will be different for each etcd node. (kube-aws default certificates are not valid against the IP - they were created for the DNS.)
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
