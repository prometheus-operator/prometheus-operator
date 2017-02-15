#!/bin/bash

set -e

AMOUNT_NODES=$1

# Wait for kubernetes cluster to become available
until kubectl cluster-info
do
    sleep 10
done

function getAmountReadyNodes {
    kubectl get nodes -ojson | jq '[.items[].status.conditions[] | select( .type=="Ready" and .status=="True")] | length'
}

# Wait for all nodes to become ready
until [[ $(getAmountReadyNodes) == $AMOUNT_NODES ]]
do
    echo "Waiting for nodes to become ready: $(getAmountReadyNodes) / $AMOUNT_NODES are ready."
    sleep 10
done
