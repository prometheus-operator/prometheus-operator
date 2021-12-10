#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail

if [ "$1" == "operator" ]; then
  # Print the output via usage but remove the line containing the executable info
  go run cmd/operator/main.go --help 2> >(grep -v 'Usage of')
fi
