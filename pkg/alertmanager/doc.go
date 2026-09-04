// Copyright The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package alertmanager implements the controller which reconciles the
Alertmanager and AlertmanagerConfig resources in a Kubernetes cluster.

This documentation centralizes key information to guide contributors making
changes to the Alertmanager configuration support.

# Overview

The Prometheus Operator supports two Custom Resource Definitions (CRD) related
to Alertmanager:

 1. The Alertmanager CRD which defines the Alertmanager StatefulSet.
 2. The AlertmanagerConfig CRD which defines the configuration for
    Alertmanager instances.

The Alertmanager CRD version is v1 while the AlertmanagerConfig CRD exists in
two versions: v1alpha1 (the stored version and conversion hub) and v1beta1.
There's a conversion webhook service which can convert custom resources
between the 2 versions.

The operator builds the final Alertmanager configuration by merging the global
configuration (a Kubernetes Secret referenced by spec.configSecret) with the
AlertmanagerConfig objects selected by the Alertmanager resource.

# Updating the AlertmanagerConfig CRD

If you're contributing a change to the AlertmanagerConfig CRD (for instance
adding a new field to a receiver), you will likely need to work with the
following files:

 1. pkg/apis/monitoring/v1alpha1/alertmanager_config_types.go: type
    definitions for the alpha version.
 2. pkg/apis/monitoring/v1beta1/alertmanager_config_types.go: type
    definitions for the beta version.
 3. pkg/apis/monitoring/v1beta1/conversion_from.go and
    pkg/apis/monitoring/v1beta1/conversion_to.go: conversion logic between
    the v1alpha1 (hub) and v1beta1 versions.
 4. pkg/alertmanager/validation/v1alpha1/validation.go and
    pkg/alertmanager/validation/v1beta1/validation.go: version-specific
    semantic validation, used by both the operator and the admission webhook
    (pkg/admission).
 5. pkg/alertmanager/amcfg.go: logic converting AlertmanagerConfig objects
    into the final Alertmanager configuration. Fields which aren't supported
    by the running Alertmanager version have to be dropped (see the various
    sanitize() methods).

After updating the API types, run "make generate" to refresh all generated
assets (deepcopy functions, CRD manifests, bundle.yaml and API documentation)
and commit the changes.

# Supporting new fields of the native Alertmanager configuration

The operator parses the global configuration secret using its own mirror of
the upstream Alertmanager configuration types, defined in
pkg/alertmanager/types.go. The local types exist to avoid the obfuscation of
secret values when marshalling the configuration (see
https://github.com/prometheus/alertmanager/issues/1985 for details).

Because the parsing is strict, a configuration field added upstream must also
be added to the mirrored types, otherwise a valid configuration using it is
rejected with an error such as "field xxx not found in type
alertmanager.alertmanagerConfig".

# Testing

Most of the configuration generation logic is covered by unit tests in
pkg/alertmanager/amcfg_test.go which compare the generated configuration
against golden files stored in pkg/alertmanager/testdata. When a change
affects many golden files at once, run "make test-unit-update-golden" to
refresh them. See TESTING.md at the repository root for more details.

# Example contribution flow

 1. Identify whether your change impacts v1alpha1, v1beta1 or both.
 2. Update the corresponding alertmanager_config_types.go file(s) and the
    conversion functions, then run "make generate".
 3. Update the validation logic if needed.
 4. Update the configuration generation in amcfg.go, making sure that the
    field is dropped for Alertmanager versions which don't support it.
 5. Add or update the unit tests and run "make test-unit".
 6. Run "make check" to lint the code and the API definitions.

For reference, https://github.com/prometheus-operator/prometheus-operator/pull/5886
is an example of a pull request adding new fields to the AlertmanagerConfig
CRD.
*/
package alertmanager
