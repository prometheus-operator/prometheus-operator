#!/bin/bash

set -eu

ID=$1
TITLE=$2

echo "🐙 >>> started the creation process for RFD ${ID} titled \"${TITLE}\""

proposal_exists=$(find . -maxdepth 1  -name ${ID})

if [[ -z ${proposal_exists} ]]; then
    git checkout -b PROPOSAL-${ID}
    mkdir $ID
    cat ./templates/prototype.md | sed s/ID/$ID/ | sed s/TITLE/"$TITLE"/ > $ID/README.md
    echo "🍀 >>> You are all set! Good luck"
    exit 0
else
    echo "😞 >>> proposal with name ${ID} already exists. Check out with your
    co-workers or pick a different id."
    exit 1
fi

exit 1
