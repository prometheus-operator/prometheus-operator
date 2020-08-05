---
name: Bug
about: Report a bug related to the Prometheus Operator
labels: kind/bug
---

<!--

Feel free to ask questions in #prometheus-operator on Kubernetes Slack!

Note: This repository is about prometheus-operator itself, if you have questions about:
- helm installation, go to https://github.com/helm/charts repository
- kube-prometheus setup, go to https://github.com/prometheus-operator/kube-prometheus

-->

**What happened?**

**Did you expect to see something different?**

**How to reproduce it (as minimally and precisely as possible)**:

**Environment**

* Prometheus Operator version:

    `Insert image tag or Git SHA here`
    <!-- Try kubectl -n monitoring describe deployment prometheus-operator -->

* Kubernetes version information:

    `kubectl version`
    <!-- Replace the command with its output above -->

* Kubernetes cluster kind:

    insert how you created your cluster: kops, bootkube, etc.

* Manifests:

```
insert manifests relevant to the issue
```

* Prometheus Operator Logs:

```
insert Prometheus Operator logs relevant to the issue here
```

**Anything else we need to know?**:
