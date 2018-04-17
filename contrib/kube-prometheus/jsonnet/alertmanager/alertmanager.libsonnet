{
    config:: import "alertmanager-main-secret.libsonnet",
    serviceAccount:: import "alertmanager-main-service-account.libsonnet",
    service:: import "alertmanager-main-service.libsonnet",
    alertmanager:: import "alertmanager-main.libsonnet",
}
