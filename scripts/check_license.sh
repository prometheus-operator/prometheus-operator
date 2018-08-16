#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

licRes=$(
    find . -type f -iname '*.go' ! -path '*/vendor/*' -exec \
         sh -c 'head -n3 $1 | grep -Eq "(Copyright|generated|GENERATED)" || echo -e  $1' {} {} \;
)

if [ -n "${licRes}" ]; then
	echo -e "license header checking failed:\\n${licRes}"
	exit 255
fi
