#!/bin/bash

set -eu

ID=$1
TITLE=$2

echo "🐙 >>> started the creation process for RFD ${ID} titled \"${TITLE}\""

branch_exists=$(git branch -rl *PROPOSAL-${ID})

if [[ -z ${branch_exists} ]]; then
    git checkout -b PROPOSAL-${ID}
    mkdir $ID
    cat ./templates/prototype.md | sed s/ID/$ID/ | sed s/TITLE/"$TITLE"/ > $ID/README.md
    echo "🍀 >>> You are all set! Good luck"
    exit 0
else
    echo "😞 >>> branch with name ${RFD_ID} already exists. Check out with your
    co-workers or pick a different id."
    exit 1
fi

exit 1
