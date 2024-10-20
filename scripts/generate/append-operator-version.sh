#!/usr/bin/env bash

set -xuo pipefail

# shellcheck disable=SC2209
sed=sed
if [[ "$OSTYPE" == "darwin"* ]]; then
  if ! command -v gsed &> /dev/null; then
    echo 'gsed could not be found. Please install gsed.'
    exit 1
  fi
  sed=gsed
fi

find example/prometheus-operator-crd/ -name '*.yaml' -exec $sed -i "/^    controller-gen.kubebuilder.io.version.*/a\\    operator.prometheus.io/version: $VERSION" {} +
