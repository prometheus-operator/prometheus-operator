// Copyright 2020 The prometheus-operator Authors
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
Package alertmanager implements the controller logic for reconciling Alertmanager
and AlertmanagerConfig resources in a Kubernetes cluster.

This documentation centralizes key information to guide contributors making changes
to the AlertmanagerConfig CRD.

Overview:

The Prometheus Operator supports two CRDs related to Alertmanager:
1. Alertmanager CRD: Defines the Alertmanager StatefulSet.
2. AlertmanagerConfig CRD: Defines the configuration for Alertmanager instances.

The Alertmanager CRD is in v1, while the AlertmanagerConfig CRD has two versions:
- v1alpha1: The default stored version.
- v1beta1: A beta version convertible to and from v1alpha1.

These CRDs are validated, converted, and used by the operator to generate configurations.

Updating the AlertmanagerConfig CRD:

If you're contributing to the AlertmanagerConfig CRD, you will likely need to work with the following files:

1. pkg/apis/monitoring/v1alpha1/alertmanager_config_types.go
  - Add, update, or delete configurations for the alpha version.

2. pkg/apis/monitoring/v1beta1/alertmanager_config_types.go
  - Add, update, or delete configurations for the beta version.

3. pkg/apis/monitoring/v1beta1/conversion_from.go
  - Logic to convert from the hub version (v1alpha1) to beta.

4. pkg/apis/monitoring/v1beta1/conversion_to.go
  - Logic to convert from beta to the hub version (v1alpha1).

5. pkg/alertmanager/amcfg.go
  - Logic to convert AlertmanagerConfig CRD into a configuration object.

6. pkg/alertmanager/validation/validation.go
  - Core validation methods for the AlertmanagerConfig CRD fields.

7. pkg/alertmanager/validation/v1alpha1/validation.go
  - Version-specific validation for the alpha version.

8. pkg/alertmanager/validation/v1beta1/validation.go
  - Version-specific validation for the beta version.

Example Contribution Flow:

1. Identify whether your changes impact v1alpha1, v1beta1, or both.
2. Update the corresponding alertmanager_config_types.go file(s).
3. Implement conversion logic in conversion_from.go and conversion_to.go if needed.
4. Modify validation logic in validation.go or version-specific validation files.
5. Test your changes to ensure compatibility with existing configurations.
6. Refer to example pull requests for guidance (https://github.com/prometheus-operator/prometheus-operator/pull/5886).

Notes:

- Refer to the Prometheus Operator project status for updates on CRD versions.
- Use the existing GoDoc comments in individual files for additional details.
*/
package alertmanager
