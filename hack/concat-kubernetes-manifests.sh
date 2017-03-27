#!/usr/bin/env bash

# Concatenate all files with "---" because that's how to specify multiple
# Kubernetes manifests in one file. Because the first `awk` also adds "---" in
# the first line, we remove it with the second `awk` call.
awk 'FNR==1{print "---"}1' $@ | awk '{if (NR!=1) {print}}'
