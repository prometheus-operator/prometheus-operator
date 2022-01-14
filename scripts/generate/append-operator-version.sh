#!/usr/bin/env bash

if [[ "$OSTYPE" == "darwin"* ]]; then
  find example/prometheus-operator-crd/ -name '*.yaml' -exec sed -i '' -e "/^    controller-gen.kubebuilder.io.version.*/a\\
    prometheus-operator.dev/version: $VERSION" {} +
else
  find example/prometheus-operator-crd/ -name '*.yaml' -exec sed -i "/^    controller-gen.kubebuilder.io.version.*/a\\    prometheus-operator.dev/version: $VERSION" {} +
fi
