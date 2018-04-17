local k = import "ksonnet.beta.3/k.libsonnet";
local service = k.core.v1.service;
local servicePort = k.core.v1.service.mixin.spec.portsType;
local kubePrometheus = import "kube-prometheus.libsonnet";

local namespace = "monitoring";

local controllerManagerService = service.new("kube-controller-manager-prometheus-discovery", {"k8s-app": "kube-controller-manager"}, servicePort.newNamed("http-metrics", 10252, 10252)) +
  service.mixin.metadata.withNamespace("kube-system") +
  service.mixin.metadata.withLabels({"k8s-app": "kube-controller-manager"});

local schedulerService = service.new("kube-scheduler-prometheus-discovery", {"k8s-app": "kube-scheduler"}, servicePort.newNamed("http-metrics", 10251, 10251)) +
  service.mixin.metadata.withNamespace("kube-system") +
  service.mixin.metadata.withLabels({"k8s-app": "kube-scheduler"});

local kubeDNSService = service.new("kube-dns-prometheus-discovery", {"k8s-app": "kube-dns"}, [servicePort.newNamed("http-metrics-skydns", 10055, 10055), servicePort.newNamed("http-metrics-dnsmasq", 10054, 10054)]) +
  service.mixin.metadata.withNamespace("kube-system") +
  service.mixin.metadata.withLabels({"k8s-app": "kube-dns"});

local objects = kubePrometheus.new(namespace) +
    {
        "prometheus-k8s/prometheus-k8s-service.yaml"+:
            service.mixin.spec.withPorts(servicePort.newNamed("web", 9090, "web") + servicePort.withNodePort(30900)) +
            service.mixin.spec.withType("NodePort"),
        "alertmanager-main/alertmanager-main-service.yaml"+:
            service.mixin.spec.withPorts(servicePort.newNamed("web", 9093, "web") + servicePort.withNodePort(30903)) +
            service.mixin.spec.withType("NodePort"),
        "grafana/grafana-service.yaml"+:
            service.mixin.spec.withPorts(servicePort.newNamed("http", 3000, "http") + servicePort.withNodePort(30902)) +
            service.mixin.spec.withType("NodePort"),
        "prometheus-k8s/kube-controller-manager-prometheus-discovery-service.yaml": controllerManagerService,
        "prometheus-k8s/kube-scheduler-prometheus-discovery-service.yaml": schedulerService,
        "prometheus-k8s/kube-dns-prometheus-discovery-service.yaml": kubeDNSService,
    };

{[path]: std.manifestYamlDoc(objects[path]) for path in std.objectFields(objects)}
