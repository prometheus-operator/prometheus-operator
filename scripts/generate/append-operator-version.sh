#!/usr/bin/env bash

if [[ "$OSTYPE" == "darwin"* ]]; then
  find example/prometheus-operator-crd/ example/prometheus-operator-crd-full/ -name '*.yaml' -exec sed -i '' -e "/^    controller-gen.kubebuilder.io.version.*/a\\
    operator.prometheus.io/version: $VERSION" {} +
else
  find example/prometheus-operator-crd/ example/prometheus-operator-crd-full/ -name '*.yaml' -exec sed -i "/^    controller-gen.kubebuilder.io.version.*/a\\    operator.prometheus.io/version: $VERSION" {} +
fi
