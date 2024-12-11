---
weight: 502
toc: true
title: Testing
menu:
    docs:
        parent: community
lead: ""
images: []
draft: false
---

When contributing code to Prometheus-Operator, you'll notice that every Pull Request will run against an extensive test suite. Among an extensive list of benefits that tests brings to the Project's overall health and reliability, it can be the reviewer's and contributors's best friend during development:

* Test cases serve as documentation, providing insights into the expected behavior of the software.
* Testing can prevent regressions by verifying that new changes don't break existing functionality.
* Running tests locally accelerate the feedback loop, removing the dependency that contributors might have on CI when working on a Pull Request.

This document will focus on teaching you about the different test suites that we currently have and how to run different scenarios to help your development experience!

# Test categories

## Unit tests

Unit tests are used to test particular code snippets in isolation. They are your best ally when looking for quick feedback loops in a particular function.

Imagine you're working on a PR that adds a new field to the ScrapeConfig CRD and you want to test if your change is reflected to the configmap. Instead of creating a full Kubernetes cluster, installing all the CRDs, running the Prometheus-Operator, deploying a Prometheus resource with ScrapeConfigSelectors and finally check if your change made it to the live object, you could simply write or extend a unit test for the configmap generation.

Here is an [example test](https://github.com/prometheus-operator/prometheus-operator/blob/20cc1a012515ed5a23fb0cf566318a26f2fecc31/pkg/prometheus/promcfg_test.go#L5866-L5879) that checks if the config generated from ScrapeConfigs are equal to an expected file.

Unit tests can be run with:

```shell
make test-unit
```

They can also be run for particular packages:

```shell
go test ./pkg/prometheus/server
```

Or even particular functions:

```shell
go test -run ^TestPodLabelsAnnotations$ ./pkg/prometheus/server
```

### Testing multi line string comparison - Golden files

[Golden files](https://pkg.go.dev/gotest.tools/v3/golden) are plain-text documents designed to facilitate the validation of lengthy strings. They come in handy when, for instance, you need to test a Prometheus configuration that's generated using Go structures. You can marshal this configuration into YAML and then compare it against a static reference to ensure a match. Golden files offer an elegant solution to this challenge, sparing you the need to hard-code the static configuration directly into your test code.

In the [example](https://github.com/prometheus-operator/prometheus-operator/blob/aeceb0b4fadc8307a44dc55afdceca0bea50bbb0/pkg/prometheus/promcfg_test.go#L102-L277), we're generating the Prometheus configuration (which can easily have 100+ lines for each individual test) and comparing it against a golden file.

If not for golden files, the test above, instead of ~150 lines, would easily require around ~1000 lines. The usage of golden files help us maintain test suites with several multi line strings comparison without sacrificing test readability.

### Updating Golden Files

There are contributions, e.g. adding a new required field to an existing configuration, that require to update several golden files at once. This can easily be done with the command below:

```shell
make test-unit-update-golden
```

## End-to-End tests

Sometimes, running tests in isolation is not enough and we really want test the behavior of Prometheus-Operator when running in a working Kubernetes cluster. For those occasions, end-to-end tests are our choice.

To run e2e-tests locally, first start a Kubernetes cluster.

We recommend [KinD](https://kind.sigs.k8s.io/) because it is lightweight (it can run on small notebooks) and this is what the project's CI uses. [MiniKube](https://minikube.sigs.k8s.io/docs/start/) is also another option.

### Setting up the local cluster

#### Kind

```shell
kind create cluster --name e2e
```

##### Building images and loading them into your kind cluster

###### Using docker with kind

Before running automated end-to-end tests, you need run the following command to make images and load it in your local cluster:

```shell
KIND_CONTEXT=e2e make test-e2e-images
```

###### Using podman with kind

When running Kind on MacOS using [podman](https://podman.io), it is recommended to create podman machine with `4` CPUs and `8 GiB` memory. Less resources might cause end to end tests to fail because of lack of resources in the cluster.

```shell
podman machine init --cpus=4 --memory=8192 --rootful --now
```

Before running automated end-to-end tests, you need to run the following command to make images and load it in your local cluster:

```shell
CONTAINER_CLI=podman KIND_CONTEXT=e2e make test-e2e-images
```

> [!TIP]
> If you are using [podman-desktop](https://podman-desktop.io/docs/podman/creating-a-podman-machine), it is quite easy to create the podman machine or update the podman machine resource settings. It also has a nice [kind extension](https://podman-desktop.io/docs/kind) to make the intergation with kind easier.

#### Minikube

Create the minikube cluster.

```shell
minikube start --embed-certs --profile e2e
```

Above command will use the default driver. If you want to use a different driver, use `--driver` flag
in the command. Refer [minikube documentation](https://minikube.sigs.k8s.io/docs/drivers/) for more details.

> [!NOTE]
> `--embed-certs` option is needed to embed the certs in kubeconfig like in Kind and extract it later for
> running the script `./scripts/run-external.sh`

##### Building images and loading them into your minikube cluster

###### Using docker with minikube

Before running automated end-to-end tests, you need run the following command to make images and load it in your local cluster:

```shell
make test-e2e-images-minikube
```

###### Using podman with minikube

Follow the same instructions from [using podman with kind](#using-podman-with-kind) for creating podman machine.

Before running automated end-to-end tests, you need to run the following command to make images and load it in your local cluster:

```shell
CONTAINER_CLI=podman make test-e2e-images-minikube
```

> [!TIP]
> If you are using [podman-desktop](https://podman-desktop.io/docs/podman/creating-a-podman-machine), it is quite easy to create the podman machine or update the podman machine resource settings. It also has a nice [minikube extension](https://podman-desktop.io/docs/minikube) to make the intergation with minikube easier.

### Troubleshooting

If you get errors while running the `make` commands, most probably running `make clean` and then executing commands will fix
the issues.

#### Manual Testing

For manual testing, you can use the utility script [scripts/run-external.sh](scripts/run-external.sh), it will check all the requirements and run your local version of the Prometheus Operator on your cluster:

If the default context already points to the test cluster created, you can simply run:

```shell
./scripts/run-external.sh -c
```

else run:

```shell
./scripts/run-external.sh <cluster context name>
```

### Running the automated E2E Tests

To run the automated end-to-end tests, run the following command:

```shell
make test-e2e
```

`make test-e2e` will run the complete end-to-end test suite. Those are the same tests we run in Pull Requests pipelines and it will make sure all features requirements amongst ***all*** controllers are working.

When working on a contribution though, it's rare that you'll need to make a change that impacts all controllers at once. Running the complete test suite takes a ***long time***, so you might want to run only the tests that are relevant to your change while developing it.

### Skipping test suites

Although we always run all the tests in CI, skipping irrelevant tests are great during development as they shorten the feedback loop.

The following [Makefile](Makefile) targets can run specific end-to-end tests:

* `make test-e2e-alertmanager` - Will run Alertmanager tests.
* `make test-e2e-thanos-ruler` - Will run Thanos-Ruler tests.
* `make test-e2e-prometheus` - Will run Prometheus tests with limited namespace permissions.
* `make test-e2e-prometheus-all-namespaces` - Will run regular Prometheus tests.
* `make test-e2e-operator-upgrade` - Will validate that a monitoring stack managed by the previous version of Prometheus-Operator will continue to work after an upgrade to the current version.
* `make test-e2e-prometheus-upgrade` - Will validate that a series of Prometheus versions can be sequentially upgraded.
* `make test-e2e-feature-gates` - Will validate the features behind a gate.

### Running only one end-to-end test

The test suites can easily take some dozens of minutes, even when running on your top-notch laptop. If you're debugging a particular test, it might be advantageous to run only this specific test. For example, the following command will only run the `TestPrometheusRuleCRDValidation/valid-rule-names` sub-test:

```shell
TEST_RUN_ARGS="-run TestPrometheusRuleCRDValidation/valid-rule-names" make test-e2e-prometheus
```
