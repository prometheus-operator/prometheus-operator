This directory contains a very simple program that can receive alert
notifications from Alertmanager. It is used by the end-to-end tests to verify
that Alertmanager works as expected.

The program is available at `quay.io/prometheus-operator/prometheus-alertmanager-test-webhook:latest`.

## Updating the image on quay.io

The image requires very few updates since the program is very simple and only used for testing.

Pre-requisites:
* Credentials to push the image to `quay.io/prometheus-operator/prometheus-alertmanager-test-webhook`.
* Buildah CLI + the `qemu-user-static` package.

Running `make manifest` should be all that is needed.
