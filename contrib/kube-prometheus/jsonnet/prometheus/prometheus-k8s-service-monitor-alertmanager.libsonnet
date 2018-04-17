{
    new(namespace)::
        {
            "apiVersion": "monitoring.coreos.com/v1",
            "kind": "ServiceMonitor",
            "metadata": {
                "name": "alertmanager",
                "namespace": namespace,
                "labels": {
                    "k8s-app": "alertmanager"
                }
            },
            "spec": {
                "selector": {
                    "matchLabels": {
                        "alertmanager": "main"
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
