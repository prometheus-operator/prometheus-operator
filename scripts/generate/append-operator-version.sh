#!/usr/bin/env bash

# Appends the operator.prometheus.io/version annotation to the generated CRDs.
#
# When the controller-gen.kubebuilder.io/version annotation is present
# (OMIT_TOOL_VERSION=false), the operator version is inserted right after it.
# Otherwise (OMIT_TOOL_VERSION=true, the default), an annotations block is
# created right below the top-level metadata field.

find example/prometheus-operator-crd/ example/prometheus-operator-crd-full/ -name '*.yaml' -print0 | while IFS= read -r -d '' f; do
  if grep -q '^    controller-gen\.kubebuilder\.io/version' "$f"; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' -e "/^    controller-gen.kubebuilder.io.version.*/a\\
    operator.prometheus.io/version: $VERSION" "$f"
    else
      sed -i "/^    controller-gen.kubebuilder.io.version.*/a\\    operator.prometheus.io/version: $VERSION" "$f"
    fi
  else
    awk -v ver="$VERSION" '
      !inserted && /^metadata:/ {
        print
        print "  annotations:"
        print "    operator.prometheus.io/version: " ver
        inserted = 1
        next
      }
      { print }
    ' "$f" > "$f.tmp" && mv "$f.tmp" "$f"
  fi
done
