#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print commands
set -x

MFLAGS=${*:-} # MFLAGS are the parent make call's flags (see Makefile)
SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

# Detect selinux and set docker run volume :Z option so we can write the generated files
VOLUME_OPTIONS=""
if hash getenforce 2> /dev/null && getenforce | grep 'Enforcing' > /dev/null; then
  VOLUME_OPTIONS=":Z"
fi

# shellcheck disable=SC2068
docker run \
    --rm \
    -u="$(id -u "$USER")":"$(id -g "$USER")" \
    -v "${SCRIPTDIR}/..:/go/src/github.com/coreos/prometheus-operator${VOLUME_OPTIONS}" \
    po-jsonnet make ${MFLAGS[@]} generate
