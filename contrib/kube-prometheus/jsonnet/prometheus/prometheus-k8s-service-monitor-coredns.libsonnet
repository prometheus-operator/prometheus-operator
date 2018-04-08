{
    new(namespace)::
        {
            "apiVersion": "monitoring.coreos.com/v1",
            "kind": "ServiceMonitor",
            "metadata": {
                "name": "coredns",
                "namespace": namespace,
                "labels": {
                    "k8s-app": "coredns"
                },
            },
            "spec": {
                "jobLabel": "k8s-app",
                "selector": {
                    "matchLabels": {
                        "k8s-app": "coredns",
                        "component": "metrics"
                    }
                },
                "namespaceSelector": {
                    "matchNames": [
                        "kube-system"
                    ]
                },
                "endpoints": [
                    {
                        "port": "http-metrics",
                        "interval": "15s",
                        "bearerTokenFile": "/var/run/secrets/kubernetes.io/serviceaccount/token"
                    }
                ]
            }
        }
}
