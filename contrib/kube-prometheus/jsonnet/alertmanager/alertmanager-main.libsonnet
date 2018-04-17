{
    new(namespace)::
        {
          apiVersion: "monitoring.coreos.com/v1",
          kind: "Alertmanager",
          metadata: {
            name: "main",
            namespace: namespace,
            labels: {
              alertmanager: "main",
            },
          },
          spec: {
            replicas: 3,
            version: "v0.14.0",
            serviceAccountName: "alertmanager-main",
          },
        }
}
