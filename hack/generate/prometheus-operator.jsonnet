local k = import "ksonnet/ksonnet.beta.3/k.libsonnet";
local rawVersion = importstr "../../VERSION";

local removeLineBreaks = function(str) std.join("", std.filter(function(c) c != "\n", std.stringChars(str)));
local version = removeLineBreaks(rawVersion);

local deployment = k.apps.v1beta2.deployment;
local container = k.apps.v1beta2.deployment.mixin.spec.template.spec.containersType;
local containerPort = container.portsType;

local targetPort = 8080;
local podLabels = {"k8s-app": "prometheus-operator"};

local operatorContainer =
  container.new("prometheus-operator", "quay.io/coreos/prometheus-operator:v" + version) +
  container.withPorts(containerPort.newNamed("http", targetPort)) +
  container.withArgs(["--kubelet-service=kube-system/kubelet", "--config-reloader-image=quay.io/coreos/configmap-reload:v0.0.1"]) +
  container.mixin.resources.withRequests({cpu: "100m", memory: "50Mi"}) +
  container.mixin.resources.withLimits({cpu: "200m", memory: "100Mi"});

local operatorDeployment = deployment.new("prometheus-operator", 1, operatorContainer, podLabels) +
  deployment.mixin.spec.selector.withMatchLabels(podLabels) +
  deployment.mixin.metadata.withLabels(podLabels) +
  deployment.mixin.spec.template.spec.securityContext.withRunAsNonRoot(true) +
  deployment.mixin.spec.template.spec.securityContext.withRunAsUser(65534);

operatorDeployment
