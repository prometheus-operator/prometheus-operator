{
    new(namespace)::
        {
            "apiVersion": "monitoring.coreos.com/v1",
            "kind": "ServiceMonitor",
            "metadata": {
                "name": "kube-apiserver",
                "namespace": namespace,
                "labels": {
                    "k8s-app": "apiserver"
                }
            },
            "spec": {
                "jobLabel": "component",
                "selector": {
                    "matchLabels": {
                        "component": "apiserver",
                        "provider": "kubernetes"
                    }
                },
                "namespaceSelector": {
                    "matchNames": [
                        "default"
                    ]
                },
                "endpoints": [
                    {
                        "port": "https",
                        "interval": "30s",
                        "scheme": "https",
                        "tlsConfig": {
                            "caFile": "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
                            "serverName": "kubernetes"
                        },
                        "bearerTokenFile": "/var/run/secrets/kubernetes.io/serviceaccount/token"
                    }
                ]
            }
        }
}
