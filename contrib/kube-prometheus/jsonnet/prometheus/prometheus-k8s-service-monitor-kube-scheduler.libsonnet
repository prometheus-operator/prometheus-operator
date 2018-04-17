{
    new(namespace)::
        {
            "apiVersion": "monitoring.coreos.com/v1",
            "kind": "ServiceMonitor",
            "metadata": {
                "name": "kube-scheduler",
                "namespace": namespace,
                "labels": {
                    "k8s-app": "kube-scheduler"
                }
            },
            "spec": {
                "jobLabel": "k8s-app",
                "endpoints": [
                    {
                        "port": "http-metrics",
                        "interval": "30s"
                    }
                ],
                "selector": {
                    "matchLabels": {
                        "k8s-app": "kube-scheduler"
                    }
                },
                "namespaceSelector": {
                    "matchNames": [
                        "kube-system"
                    ]
                }
            }
        }
}
