# E2E Testing

End-to-end (e2e) testing is automated testing for real user scenarios.

## Build and run test

Prerequisites:
- a running k8s cluster and kube config. We will need to pass kube config as arguments.
- Have kubeconfig file ready.
- Have prometheus operator image ready.

e2e tests are written as Go test. All go test techniques apply, e.g. picking
what to run, timeout length. Let's say I want to run all tests in "test/e2e/":

```
$ go test -v ./test/e2e/ --kubeconfig "$HOME/.kube/config" --operator-image=quay.io/coreos/prometheus-operator
```
