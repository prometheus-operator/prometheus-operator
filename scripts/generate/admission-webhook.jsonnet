local admissionWebhook = (import 'prometheus-operator/admission-webhook.libsonnet');
local config = (import 'config.jsonnet');
local aw = admissionWebhook(config {
  image: 'quay.io/prometheus-operator/admission-webhook:v' + config.version,
  tlsSecretName: 'admission-webhook-certs',
});

{
  'service-account.yaml': aw.serviceAccount,
  'deployment.yaml': aw.deployment,
  'service.yaml': aw.service,
  'service-monitor.yaml': aw.serviceMonitor,
  'pod-disruption-budget.yaml': aw.podDisruptionBudget,
}
