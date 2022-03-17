local admissionWebhook = (import 'prometheus-operator/admission-webhook.libsonnet');
local config = (import 'config.jsonnet');
local aw = admissionWebhook(config {
  image: 'quay.io/prometheus-operator/admission-webhook:v' + config.version,
});

{
  'service-account.yaml': aw.serviceAccount,
  'deployment.yaml': aw.deployment,
  'service.yaml': aw.service,
  'service-monitor.yaml': aw.serviceMonitor,
}
