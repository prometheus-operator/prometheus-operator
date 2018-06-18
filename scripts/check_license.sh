#!/usr/bin/env bash

licRes=$(
    find . -type f -iname '*.go' ! -path '*/vendor/*' -exec \
         sh -c 'head -n3 $1 | grep -Eq "(Copyright|generated|GENERATED)" || echo -e  $1' {} {} \;
)

if [ -n "${licRes}" ]; then
	echo -e "license header checking failed:\\n${licRes}"
	exit 255
fi
