#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

SCRIPT_DIR=$(dirname "${BASH_SOURCE[0]}")

"${SCRIPT_DIR}"/create-minikube.sh

# build nsenter
# https://github.com/kubernetes/helm/issues/966
sudo apt-get update
sudo apt-get install libncurses5-dev libslang2-dev gettext zlib1g-dev libselinux1-dev debhelper lsb-release pkg-config po-debconf autoconf automake autopoint libtool
mkdir .tmp || true
wget https://www.kernel.org/pub/linux/utils/util-linux/v2.30/util-linux-2.30.2.tar.gz -qO - | tar -xz -C .tmp/
cd .tmp/util-linux-2.30.2 && ./autogen.sh && ./configure && make nsenter && sudo cp nsenter /usr/local/bin && cd -  

make test-e2e-helm

"${SCRIPT_DIR}"/delete-minikube.sh
