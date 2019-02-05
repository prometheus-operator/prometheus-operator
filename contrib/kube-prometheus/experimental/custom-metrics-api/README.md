# Custom Metrics API

The custom metrics API allows the HPA v2 to scale based on arbirary metrics.

This directory contains an example deployment which extends the Prometheus Adapter, deployed with kube-prometheus, serve the [Custom Metrics API](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/instrumentation/custom-metrics-api.md) by talking to Prometheus running inside the cluster.

Make sure you have the Prometheus Adapter up and running in the `monitoring` namespace.

You can deploy everything in the `monitoring` namespace using `./deploy.sh`.

When you're done, you can teardown using the `./teardown.sh` script.

### Sample App

Additionally, this directory contains a sample app that uses the [Horizontal Pod Autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) to scale the Deployment's replicas of Pods up and down as needed.  
Deploy this app by running `kubectl apply -f sample-app.yaml`. 
Make the app accessible on your system, for example by using `kubectl port-forward svc/sample-app 8080`. Next you need to put some load on its http endpoints. 

A tool like [hey](https://github.com/rakyll/hey) is helpful for doing so: `hey -c 20 -n 100000000 http://localhost:8080/metrics`

There is an even more detailed information on this sample app at [luxas/kubeadm-workshop](https://github.com/luxas/kubeadm-workshop#deploying-the-prometheus-operator-for-monitoring-services-in-the-cluster).
