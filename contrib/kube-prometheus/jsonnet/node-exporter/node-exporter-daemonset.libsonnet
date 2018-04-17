local k = import "ksonnet.beta.3/k.libsonnet";

local daemonset = k.apps.v1beta2.daemonSet;
local container = daemonset.mixin.spec.template.spec.containersType;
local volume = daemonset.mixin.spec.template.spec.volumesType;
local containerPort = container.portsType;
local containerVolumeMount = container.volumeMountsType;
local podSelector = daemonset.mixin.spec.template.spec.selectorType;

local nodeExporterVersion = "v0.15.2";
local kubeRbacProxyVersion = "v0.3.0";
local podLabels = {"app": "node-exporter"};

local procVolumeName = "proc";
local procVolume = volume.fromHostPath(procVolumeName, "/proc");
local procVolumeMount = containerVolumeMount.new(procVolumeName, "/host/proc");

local sysVolumeName = "sys";
local sysVolume = volume.fromHostPath(sysVolumeName, "/sys");
local sysVolumeMount = containerVolumeMount.new(sysVolumeName, "/host/sys");

local nodeExporter =
  container.new("node-exporter", "quay.io/prometheus/node-exporter:" + nodeExporterVersion) +
  container.withArgs([
    "--web.listen-address=127.0.0.1:9101",
    "--path.procfs=/host/proc",
    "--path.sysfs=/host/sys",
  ]) +
  container.withVolumeMounts([procVolumeMount, sysVolumeMount]) +
  container.mixin.resources.withRequests({cpu: "102m", memory: "180Mi"}) +
  container.mixin.resources.withLimits({cpu: "102m", memory: "180Mi"});

local proxy =
  container.new("kube-rbac-proxy", "quay.io/coreos/kube-rbac-proxy:" + kubeRbacProxyVersion) +
  container.withArgs([
	"--secure-listen-address=:9100",
	"--upstream=http://127.0.0.1:9101/",
  ]) +
  container.withPorts(containerPort.newNamed("https", 9100)) +
  container.mixin.resources.withRequests({cpu: "10m", memory: "20Mi"}) +
  container.mixin.resources.withLimits({cpu: "20m", memory: "40Mi"});

local c = [nodeExporter, proxy];

{
    new(namespace)::
        daemonset.new() +
          daemonset.mixin.metadata.withName("node-exporter") +
          daemonset.mixin.metadata.withNamespace(namespace) +
          daemonset.mixin.metadata.withLabels(podLabels) +
          daemonset.mixin.spec.selector.withMatchLabels(podLabels) +
          daemonset.mixin.spec.template.metadata.withLabels(podLabels) +
          daemonset.mixin.spec.template.spec.withContainers(c) +
          daemonset.mixin.spec.template.spec.withVolumes([procVolume, sysVolume]) +
          daemonset.mixin.spec.template.spec.securityContext.withRunAsNonRoot(true) +
          daemonset.mixin.spec.template.spec.securityContext.withRunAsUser(65534) +
          daemonset.mixin.spec.template.spec.withServiceAccountName("node-exporter")
}
