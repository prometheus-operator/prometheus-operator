# Exposing Prometheus and Alertmanager

The Prometheus Operator takes care of operating Prometheus and Alertmanagers clusters, however, there are many ways in Kubernetes to expose these to the outside world. This document outlines best practices and caveats to do so in various ways.

## NodePort

The easiest way to expose Prometheus or Alertmanager is to use a `Service` of type `NodePort`.

Let's create a simple `Prometheus` object with one replica.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: main
spec:
  replicas: 1
  version: v1.7.1
  resources:
    requests:
      memory: 400Mi
```

All Prometheus `Pod`s are labeled with `prometheus: <prometheus-name>`, as the Prometheus object's name is `main`, the selector ends up being `prometheus: main`. Meaning, the respective manifest for the `Service` needs the selector to be `prometheus: main`.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: prometheus-main
spec:
  type: NodePort
  ports:
  - name: web
    nodePort: 30900
    port: 9090
    protocol: TCP
    targetPort: web
  selector:
    prometheus: main
```

After creating a `Service` with the above manifest, the web UI of Prometheus will be accessible by browsing to any of the worker nodes using `http://<node-ip>:30900/`.

Exposing the Alertmanager works in the same fashion, the only difference being, that the selector is `alertmanager: <alertmanager-name>`.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: main
spec:
  replicas: 3
  version: v0.7.1
  resources:
    requests:
      memory: 400Mi
```

And the `Service`.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: alertmanager-main
spec:
  type: NodePort
  ports:
  - name: web
    nodePort: 30903
    port: 9093
    protocol: TCP
    targetPort: web
  selector:
    alertmanager: main
```

And the Alertmanager web UI will be available at `http://<node-ip>:30903/`.

## Kubernetes API

The Kubernetes API has a feature of forwarding requests from the API to a cluster internal `Service`. The general URL scheme to access these is:

```
http(s)://master-host/api/v1/proxy/namespaces/<namespace>/services/<service-name>:<port-name-or-number>/
```

> Note for ease of use, you can use `kubectl proxy`, it proxies requests from a local address to the Kubernetes API server and handles authentication for you.

To be able to do so, create a `Service` of type `ClusterIP`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: prometheus-main
spec:
  type: ClusterIP
  ports:
  - name: web
    port: 9090
    protocol: TCP
    targetPort: web
  selector:
    prometheus: main
```

A caveat about this is that Prometheus and Alertmanager need to be configured with the full URL they are going to be exposed at. Therefore the `Prometheus` manifest will need an entry for `externalUrl`:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: main
spec:
  replicas: 1
  version: v1.7.1
  externalUrl: http://127.0.0.1:8001/api/v1/proxy/namespaces/default/services/prometheus-main:web/
  resources:
    requests:
      memory: 400Mi
```

> Note the `externalUrl` used there uses the host `127.0.0.1:8001`, which is how `kubectl proxy` exposes the Kubernetes API by default.

Once the Prometheus `Pod`s are running they are reachable under the specified `externalUrl`.

The Alertmanager works with the same approach so the manifest for the `Alertmanager` object is simply the following:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: main
spec:
  replicas: 3
  version: v0.7.1
  externalUrl: http://127.0.0.1:8001/api/v1/proxy/namespaces/default/services/alertmanager-main:web/
  resources:
    requests:
      memory: 400Mi
```

And the respective `Service`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: alertmanager-main
spec:
  type: ClusterIP
  ports:
  - name: web
    port: 9093
    protocol: TCP
    targetPort: web
  selector:
    alertmanager: main
```

Then it will be available under http://127.0.0.1:8001/api/v1/proxy/namespaces/default/services/alertmanager-main:web/.

> Note the URL used there uses the host `127.0.0.1:8001`, which is how `kubectl proxy` exposes the Kubernetes API by default.

## Ingress

Exposing the Prometheus or Alertmanager web UI via an `Ingress` object is requires a running ingress controller. If you are unfamiliar with Ingress, have a look at the [documentation](https://kubernetes.io/docs/user-guide/ingress/).

This example was tested with the [nginx ingress controller](https://github.com/kubernetes/ingress/tree/master/controllers/nginx).  For a quick-start for running the nginx ingress controller run:

```
kubectl create -f https://raw.githubusercontent.com/kubernetes/ingress/master/examples/deployment/nginx/nginx-ingress-controller.yaml
```

> It is highly recommended to compare the available ingress controllers for a production environment. The nginx ingress controller may or may not be what is suitable for your production environment. Also have a look at HA Proxy, Træfɪk, GCE, AWS, and more.

> WARNING: If you ingress is exposed to the internet, everyone can have full acesss on your resources. It's strongly recommend to enable an [external authentication](https://github.com/kubernetes/ingress/blob/858e3ff2354fb0f5066a88774b904b2427fb9433/examples/external-auth/nginx/README.md) or [whitelisting ip address](https://github.com/kubernetes/ingress/blob/7ca7652ab26e1a5775f3066f53f28d5ea5eb3bb7/controllers/nginx/configuration.md#whitelist-source-range)

An `Ingress` object also requires a `Service` to be setup as the requests are simply routed from the ingress endpoint to the internal `Service`.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: prometheus-main
spec:
  type: ClusterIP
  ports:
  - name: web
    port: 9090
    protocol: TCP
    targetPort: web
  selector:
    prometheus: main
---
apiVersion: v1
kind: Service
metadata:
  name: alertmanager-main
spec:
  type: ClusterIP
  ports:
  - name: web
    port: 9090
    protocol: TCP
    targetPort: web
  selector:
    alertmanager: main
```

Then a corresponding `Ingress` object would be as follows.

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: monitoring
  annotations:
    ingress.kubernetes.io/whitelist-source-range: 10.0.0.0/16 # change this range to admin ips
    ingress.kubernetes.io/rewrite-target: "/"
spec:
  rules:
  - http:
      paths:
      - backend:
          serviceName: prometheus-main
          servicePort: 9090
        path: /prometheus
      - backend:
          serviceName: alertmanager-main
          servicePort: 9093
        path: /alertmanager
```

Lastly the `Prometheus` and `Alertmanager` objects need to be created with the `externalUrl` they are going to be browsed to.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: main
spec:
  replicas: 1
  version: v1.7.1
  externalUrl: http://monitoring.my.systems/prometheus
  resources:
    requests:
      memory: 400Mi
---
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: main
spec:
  replicas: 3
  version: v0.7.1
  externalUrl: http://monitoring.my.systems/alertmanager
  resources:
    requests:
      memory: 400Mi
```

> Note that there is the path `/prometheus` at the end of the `externalUrl`, as that is what was specified in the `Ingress` object.
