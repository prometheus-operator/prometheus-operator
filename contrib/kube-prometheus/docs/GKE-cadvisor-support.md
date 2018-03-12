# Kubelet / cAdvisor special configuration updates for GKE 

In order to allow Prometheus to access the endpoints provided by the kubelet/cAdvisor on GKE we have to downgrade the scheme to HTTP (from HTTPS).


On linux:

```
sed -i -e 's/https/http/g' \
contrib/kube-prometheus/manifests/prometheus/prometheus-k8s-service-monitor-kubelet.yaml
```

On MacOs: 

``` 
sed -i '' -e 's/https/http/g' \
contrib/kube-prometheus/manifests/prometheus/prometheus-k8s-service-monitor-kubelet.yaml
```

After you have modified the yaml file please run

```
kubectl apply -f contrib/kube-prometheus/manifests/prometheus/prometheus-k8s-service-monitor-kubelet.yaml
```
