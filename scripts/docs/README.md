This directory contains the configuration and templates for generating the
HTML/MarkDown documentation of the Prometheus operator's custom resource
definitions. It uses the
[gen-crd-api-reference-docs](https://github.com/ahmetb/gen-crd-api-reference-docs)
project.

## Building

From the project's top directory, run:

```console
make --always-make generate-docs
```
