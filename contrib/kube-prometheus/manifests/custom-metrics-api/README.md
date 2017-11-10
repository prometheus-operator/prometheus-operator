# Custom Metrics API

The custom metrics API allows the HPA v2 to scale on arbirary metrics.

This directory contains an example deployment of the custom metrics API adapter using Prometheus as the backing monitoring system.

In order to deploy the custom metrics adapter for Prometheus you need to generate TLS certficates used to serve the API. An example of how these could be generated can be found in `./gencerts.sh`, note that this is _not_ recommended to be used in production. You need to employ a secure PKI strategy, this is merely an example to get started and try it out quickly.

Once the generated `Secret` with the certificates is in place, you can deploy everything in the `monitoring` namespace using `./deploy.sh`.

When you're done, you can teardown using the `./teardown.sh` script.
