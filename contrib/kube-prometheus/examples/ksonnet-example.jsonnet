local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';
local daemonset = k.apps.v1beta2.daemonSet;

((import 'kube-prometheus/kube-prometheus.libsonnet') + {
   nodeExporter+: {
     daemonset+:
       daemonset.mixin.metadata.withNamespace('my-custom-namespace'),
   },
 }).nodeExporter.daemonset
