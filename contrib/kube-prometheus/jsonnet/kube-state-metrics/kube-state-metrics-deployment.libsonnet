local k = import "ksonnet.beta.3/k.libsonnet";
local deployment = k.apps.v1beta2.deployment;

local deployment = k.apps.v1beta2.deployment;
local container = k.apps.v1beta2.deployment.mixin.spec.template.spec.containersType;
local volume = k.apps.v1beta2.deployment.mixin.spec.template.spec.volumesType;
local containerPort = container.portsType;
local containerVolumeMount = container.volumeMountsType;
local podSelector = deployment.mixin.spec.template.spec.selectorType;

local kubeStateMetricsVersion = "v1.3.0";
local kubeRbacProxyVersion = "v0.3.0";
local addonResizerVersion = "1.0";
local podLabels = {"app": "kube-state-metrics"};

local proxyClusterMetrics =
  container.new("kube-rbac-proxy-main", "quay.io/coreos/kube-rbac-proxy:" + kubeRbacProxyVersion) +
  container.withArgs([
	"--secure-listen-address=:8443",
	"--upstream=http://127.0.0.1:8081/",
  ]) +
  container.withPorts(containerPort.newNamed("https-main", 8443)) +
  container.mixin.resources.withRequests({cpu: "10m", memory: "20Mi"}) +
  container.mixin.resources.withLimits({cpu: "20m", memory: "40Mi"});

local proxySelfMetrics =
  container.new("kube-rbac-proxy-self", "quay.io/coreos/kube-rbac-proxy:" + kubeRbacProxyVersion) +
  container.withArgs([
	"--secure-listen-address=:9443",
	"--upstream=http://127.0.0.1:8082/",
  ]) +
  container.withPorts(containerPort.newNamed("https-self", 9443)) +
  container.mixin.resources.withRequests({cpu: "10m", memory: "20Mi"}) +
  container.mixin.resources.withLimits({cpu: "20m", memory: "40Mi"});

local kubeStateMetrics =
  container.new("kube-state-metrics", "quay.io/coreos/kube-state-metrics:" + kubeStateMetricsVersion) +
  container.withArgs([
	"--host=127.0.0.1",
	"--port=8081",
	"--telemetry-host=127.0.0.1",
	"--telemetry-port=8082",
  ]) +
  container.mixin.resources.withRequests({cpu: "102m", memory: "180Mi"}) +
  container.mixin.resources.withLimits({cpu: "102m", memory: "180Mi"});

local addonResizer =
  container.new("addon-resizer", "quay.io/coreos/addon-resizer:" + addonResizerVersion) +
  container.withCommand([
	"/pod_nanny",
	"--container=kube-state-metrics",
	"--cpu=100m",
	"--extra-cpu=2m",
	"--memory=150Mi",
	"--extra-memory=30Mi",
	"--threshold=5",
	"--deployment=kube-state-metrics",
  ]) +
  container.withEnv([
	{
	  name: "MY_POD_NAME",
      valueFrom: {
		fieldRef: {apiVersion: "v1", fieldPath: "metadata.name"}
      }
	}, {
	  name: "MY_POD_NAMESPACE",
      valueFrom: {
		fieldRef: {apiVersion: "v1", fieldPath: "metadata.namespace"}
      }
	}
  ]) +
  container.mixin.resources.withRequests({cpu: "10m", memory: "30Mi"}) +
  container.mixin.resources.withLimits({cpu: "10m", memory: "30Mi"});

local c = [proxyClusterMetrics, proxySelfMetrics, kubeStateMetrics, addonResizer];

{
    new(namespace)::
        deployment.new("kube-state-metrics", 1, c, podLabels) +
          deployment.mixin.metadata.withNamespace(namespace) +
          deployment.mixin.metadata.withLabels(podLabels) +
          deployment.mixin.spec.selector.withMatchLabels(podLabels) +
          deployment.mixin.spec.template.spec.securityContext.withRunAsNonRoot(true) +
          deployment.mixin.spec.template.spec.securityContext.withRunAsUser(65534) +
          deployment.mixin.spec.template.spec.withServiceAccountName("kube-state-metrics")
}
