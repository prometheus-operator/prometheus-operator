# Linting

This document describes how to use the standalone linting tool to validate your Prometheus Operator [CRD-based](../design.md) configuration files.

## Getting linter

To use the linter either get it with `go get -u github.com/prometheus-operator/prometheus-operator/cmd/po-lint` and executable is `$GOPATH/bin/po-lint`, or use the container image from `quay.io/coreos/po-tooling` and executable is `/go/bin/po-lint`.

## Using linter

The `po-lint` executable takes a list of yaml files to check as command arguments. It will output any errors to stderr and returns with exit code `1` on errors, `0` otherwise.

## Example

Here is an example script to lint a `src` sub-directory full of Prometheus Operator CRD files with ether local `po-lint` or Dockerized version:

```sh
#!/bin/sh

LINTER="quay.io/coreos/po-tooling"

lint_files() {
  if [ -x "$(command -v po-lint)" ]; then
    echo "Linting '${2}' files in directory '${1}'..."
    had_errors=0
    for file in $(find "${1}" -name "${2}"); do
      echo "${file}"
      po-lint "${file}"
      retval=$?
      if [ $retval -ne 0 ]; then
        had_errors=1
      fi
    done
    exit ${had_errors}
  elif [ -x "$(command -v docker)" ]; then
    echo "Using Dockerized linter."
    docker run --rm --volume "$PWD:/data:ro" --workdir /data ${LINTER} \
    /bin/bash -c "/go/bin/po-lint $1/$2"
  else
    echo "Linter executable not found."
    exit 1
  fi
}

lint_files "./src" "*.yaml"
```
