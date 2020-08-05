# kube-prometheus

## WARNING: kube-prometheus moved to [prometheus-operator/kube-prometheus](https://github.com/prometheus-operator/kube-prometheus)!

**Why did you move it?**

Even though kube-prometheus is an entirely different project, it was part of the Prometheus Operator repository for the last two years.
Moving kube-prometheus into its own repository is going to allow us to move more independently from the Prometheus Operator.
As an example, we are now going to publish versioned kube-prometheus releases, something that was not possible before.

Take a look at this issue for more information:
https://github.com/prometheus-operator/prometheus-operator/issues/2553


**What do I need to do?**

Simply go to [prometheus-operator/kube-prometheus](https://github.com/prometheus-operator/kube-prometheus) and make use of it the same way you did before.

Users depending on kube-prometheus with jsonnet-bundler, should change this their `jsonnetfile.json` and `jsonnetfile.lock.json` to point to the correct repository.

```diff
             "name": "kube-prometheus",
             "source": {
                 "git": {
-                    "remote": "https://github.com/prometheus-operator/prometheus-operator",
-                    "subdir": "contrib/kube-prometheus/jsonnet/kube-prometheus"
+                    "remote": "https://github.com/prometheus-operator/kube-prometheus",
+                    "subdir": "jsonnet/kube-prometheus"
                 }
             },
             "version": "master"
```

*Note: We needed to merge the two repositories and commit hashes are not the same anymore, when referencing prometheus-operator/kube-prometheus.*

**You still have questions why it moved?**

Feel free to either create an issue or ask in #prometheus-operator on [Kubernetes Slack](http://slack.k8s.io/).
