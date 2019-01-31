# Kubelet / cAdvisor special configuration updates for GKE

Prior to GKE 1.11, the kubelet does not support token
authentication. Until it does, Prometheus must use HTTP (not HTTPS)
for scraping.

You can configure this behavior through kube-prometheus with:
```
local kp = (import 'kube-prometheus/kube-prometheus.libsonnet') +
    (import 'kube-prometheus/kube-prometheus-insecure-kubelet.libsonnet') +
	{
        _config+:: {
		# ... config here
		}
    };
```

Or, you can patch and re-apply your existing manifests with:

On linux:

```
sed -i -e 's/https/http/g' manifests/prometheus-serviceMonitorKubelet.yaml
```

On MacOs:

```
sed -i '' -e 's/https/http/g' manifests/prometheus-serviceMonitorKubelet.yaml
```

After you have modified the yaml file please run

```
kubectl apply -f manifests/prometheus-serviceMonitorKubelet.yaml
```
