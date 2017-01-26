#!/bin/bash

kubectl get pods --all-namespaces -l app=$1 -ojsonpath=\{\.items\[\*\]\.spec\.containers\[\?\(\@.name==\"$1\"\)\].image\}
