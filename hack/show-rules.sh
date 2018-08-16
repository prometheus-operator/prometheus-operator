#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

kubectl -n prometheus-operator-e2e-tests exec -it prometheus-test-0 -c prometheus '/bin/sh -c "cat /etc/prometheus/rules/rules-0/test.rules"'
