// Copyright 2023 The prometheus-operator Authors
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

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
)

func environmentIntersection(a, b []v1.EnvVar) []string {
	m := make(map[string]struct{}, len(a))
	for _, e := range a {
		m[e.Name] = struct{}{}
	}

	var intersection []string
	for _, e := range b {
		if _, ok := m[e.Name]; ok {
			intersection = append(intersection, e.Name)
		}
	}

	return intersection
}

// BuildEnvironment returns the concatenation of the 2 argument lists of environment variables.
// It returns an error if the 2 lists intersect.
func BuildEnvironment(env, additionalEnv []v1.EnvVar) ([]v1.EnvVar, error) {
	if i := environmentIntersection(env, additionalEnv); len(i) > 0 {
		return nil, fmt.Errorf("can't set environment variables which are already managed by the operator: %s", strings.Join(i, ","))
	}

	return append(env, additionalEnv...), nil
}
