local k = import "ksonnet.beta.2/k.libsonnet";
local rawVersion = importstr "../../VERSION";

local removeLineBreaks = function(str) std.join("", std.filter(function(c) c != "\n", std.stringChars(str)));
local version = removeLineBreaks(rawVersion);

local deployment = k.extensions.v1beta1.deployment;
local container = k.extensions.v1beta1.deployment.mixin.spec.template.spec.containersType;
local containerPort = container.portsType;

local targetPort = 8080;
local podLabels = {"k8s-app": "prometheus-operator"};

local operatorContainer =
  container.new("prometheus-operator", "quay.io/coreos/prometheus-operator:v" + version) +
  container.ports(containerPort.newNamed("http", targetPort)) +
  container.args("--kubelet-service=kube-system/kubelet") +
  container.args("--config-reloader-image=quay.io/coreos/configmap-reload:v0.0.1") +
  container.mixin.resources.requests({cpu: "100m", memory: "50Mi"}) +
  container.mixin.resources.limits({cpu: "200m", memory: "100Mi"});

local operatorDeployment = deployment.new("prometheus-operator", 1, operatorContainer, podLabels) +
  deployment.mixin.metadata.labels(podLabels);

operatorDeployment
