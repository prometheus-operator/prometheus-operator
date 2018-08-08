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

      // Note that you should specify a value EITHER for 'serverName' OR for 'insecureSkipVerify'. (Don't specify a value for both of them, and don't specify a value for neither of them.)
      // * Specifying serverName: Ideally you should provide a valid value for serverName (and then insecureSkipVerify should be left as false - so that serverName gets used).
      // * Specifying insecureSkipVerify: insecureSkipVerify is only to be used (i.e. set to true) if you cannot (based on how your etcd certificates were created) use a Subject Alternative Name.
      // * If you specify a value:
      //     ** for both of these variables: When 'insecureSkipVerify: true' is specified, then also specifying a value for serverName won't hurt anything but it will be ignored.
      //     ** for neither of these variables: then you'll get authentication errors on the prom '/targets' page with your etcd targets.

      // A valid name (DNS or Subject Alternative Name) that the client (i.e. prometheus) will use to verify the etcd TLS certificate.
      //  * Note that doing `nslookup etcd.kube-system.svc.cluster.local` (on a pod in a K8s cluster where kube-prometheus has been installed) shows that kube-prometheus sets up this hostname.
      //  * `openssl x509 -noout -text -in etcd-client.pem` will print the Subject Alternative Names.
      serverName: 'etcd.kube-system.svc.cluster.local',

      // When insecureSkipVerify isn't specified, the default value is "false".
      //insecureSkipVerify: true,

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
