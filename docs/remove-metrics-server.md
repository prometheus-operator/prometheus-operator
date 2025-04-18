# ⚠️ Remove `metrics-server` Before Using `kube-prometheus`

Before setting up `kube-prometheus`, ensure that the `metrics-server` is **removed** from your Kubernetes cluster to avoid conflicts.

## Why Remove `metrics-server`?

The `metrics-server` can interfere with the Prometheus Adapter's custom metrics endpoints (`/apis/custom.metrics.k8s.io`). This may lead to issues with metric scraping and cause Prometheus Adapter errors. 

**Steps to Remove `metrics-server`:**

1. Run the following command to uninstall the `metrics-server`:
   ```bash
   kubectl delete deployment metrics-server -n kube-system
