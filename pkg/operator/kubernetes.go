// Copyright 2025 The prometheus-operator Authors
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

package operator

const (
	// InputHashAnnotationKey is the name of the annotation used to store the
	// operator's computed hash value.
	InputHashAnnotationKey = "prometheus-operator-input-hash"

	// The label and annotation keys defined below come from
	// https://kubernetes.io/docs/reference/labels-annotations-taints/

	// ApplicationNameLabelKey is the name of the application.
	ApplicationNameLabelKey = "app.kubernetes.io/name"

	// ManagedByLabelKey is the tool managing the application (e.g. this operator).
	ManagedByLabelKey = "app.kubernetes.io/managed-by"

	// ManagedByLabelValue is the name of this operator.
	ManagedByLabelValue = "prometheus-operator"

	// ApplicationInstanceLabelKey is the unique name identifying the application's instance.
	ApplicationInstanceLabelKey = "app.kubernetes.io/instance"

	// ApplicationVersionLabelKey is the version of the application.
	ApplicationVersionLabelKey = "app.kubernetes.io/version"

	// DefaultContainerAnnotationKey is the annotation defining the default container of the pod.
	DefaultContainerAnnotationKey = "kubectl.kubernetes.io/default-container"
)
