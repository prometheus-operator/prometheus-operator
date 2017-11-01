# Adding kube-prometheus to [KOPS](https://github.com/kubernetes/kops) on AWS 1.5.x


## Prerequisites

A running Kubernetes cluster created with [KOPS](https://github.com/kubernetes/kops).
 
These instructions have currently been tested with  **topology=public** on AWS with KOPS 1.7.1 and Kubernetes 1.7.x

Following the instructions in the [README](https://github.com/coreos/prometheus-operator/blob/master/contrib/kube-prometheus/README.md):

Example:

```bash
git clone -b master https://github.com/coreos/prometheus-operator.git prometheus-operator-temp;
cd prometheus-operator-temp/contrib/kube-prometheus
./hack/cluster-monitoring/deploy
kubectl -n kube-system create -f manifests/k8s/self-hosted/
cd -
rm -rf prometheus-operator-temp
```
