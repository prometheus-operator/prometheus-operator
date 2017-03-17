#!/bin/bash

kubectl -n prometheus-operator-e2e-tests exec -it prometheus-test-0 -c prometheus '/bin/sh -c "cat /etc/prometheus/rules/rules-0/test.rules"'
