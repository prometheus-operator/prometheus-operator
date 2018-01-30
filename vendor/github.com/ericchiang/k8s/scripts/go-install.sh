#!/bin/bash -e

function usage {
    >&2 echo "./go-install.sh [repo] [repo import path] [tool import path] [rev]"
}

REPO=$1
REPO_ROOT=$2
TOOL=$3
REV=$4

git clone $REPO _output/src/$REPO_ROOT
cd _output/src/$REPO_ROOT
git checkout $REV
cd -
GOPATH=$PWD/_output GOBIN=$PWD/_output/bin go install -v $TOOL
