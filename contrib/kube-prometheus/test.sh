#!/usr/bin/env bash
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail


for i in examples/jsonnet-snippets/*.jsonnet; do
    [ -f "$i" ] || break
    echo "Testing: ${i}"
    echo ""
    snippet="local kp = $(<${i});

$(<examples/jsonnet-build-snippet/build-snippet.jsonnet)"
    echo "${snippet}" > "test.jsonnet"
    echo "\`\`\`"
    echo "${snippet}"
    echo "\`\`\`"
    echo ""
    jsonnet -J vendor "test.jsonnet" > /dev/null
    rm -rf "test.jsonnet"
done

for i in examples/*.jsonnet; do
    [ -f "$i" ] || break
    echo "Testing: ${i}"
    echo ""
    echo "\`\`\`"
    echo "$(<${i})"
    echo "\`\`\`"
    echo ""
    jsonnet -J vendor ${i} > /dev/null
done
