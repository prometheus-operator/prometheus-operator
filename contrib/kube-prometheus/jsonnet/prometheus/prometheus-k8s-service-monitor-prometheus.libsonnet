{
    new(namespace)::
        {
            "apiVersion": "monitoring.coreos.com/v1",
            "kind": "ServiceMonitor",
            "metadata": {
                "name": "prometheus",
                "namespace": namespace,
                "labels": {
                    "k8s-app": "prometheus"
                }
            },
            "spec": {
                "selector": {
                    "matchLabels": {
                        "prometheus": "k8s"
                    }
                },
                "namespaceSelector": {
                    "matchNames": [
                        "monitoring"
                    ]
                },
                "endpoints": [
                    {
                        "port": "web",
                        "interval": "30s"
                    }
                ]
            }
        }
}
