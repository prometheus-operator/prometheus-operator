// Create JSON patch that configures the conversion webhook for the Alertmanager CRD.
local alertmanagerConfigCrd = (import 'prometheus-operator/alertmanagerconfigs-crd.json');
local conversion = (import 'prometheus-operator/conversion.libsonnet');
local admissionWebhook = (import 'prometheus-operator/admission-webhook.libsonnet');
local config = (import 'config.jsonnet');

local aw = admissionWebhook(config {
  image: 'quay.io/prometheus-operator/admission-webhook:v' + config.version,
});

{
  apiVersion: alertmanagerConfigCrd.apiVersion,
  kind: alertmanagerConfigCrd.kind,
  metadata: alertmanagerConfigCrd.metadata,
} + conversion({
  name: aw.service.metadata.name,
  namespace: aw.service.metadata.namespace,
})
