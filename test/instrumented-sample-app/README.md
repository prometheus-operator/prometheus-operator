This directory contains a sample application instrumented for Prometheus. It is
used by the end-to-end tests to verify that Prometheus can scrape metrics using
different authentication methods such as bearer tokens, mTLS and Basic-Auth.

The program is available at `quay.io/prometheus-operator/instrumented-sample-app:latest`.

## Updating the image on quay.io

The image requires very few updates since the program is very simple and only used for testing.

Pre-requisites:
* Credentials to push the image to `quay.io/prometheus-operator/instrumented-sample-app`.
* Buildah CLI + the `qemu-user-static` package.

Running `make manifest` should be all that is needed.
