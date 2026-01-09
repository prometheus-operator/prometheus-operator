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
	"iter"
	"strings"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// ArgumentsIntersection returns the list of arguments which intersect between a and b.
func ArgumentsIntersection(a, b []monitoringv1.Argument) []string {
	m := make(map[string]struct{})
	for item := range argumentNameIter(a) {
		m[item] = struct{}{}
	}

	var intersection []string
	for name := range argumentNameIter(b) {
		if _, ok := m[name]; ok {
			intersection = append(intersection, name)
			continue
		}

		negated, found := strings.CutPrefix(name, "no-")
		if !found {
			negated = fmt.Sprintf("no-%s", name)
		}

		if _, ok := m[negated]; ok {
			intersection = append(intersection, negated)
		}
	}

	return intersection
}

func argumentNameIter(args []monitoringv1.Argument) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, arg := range args {
			if !yield(arg.Name) {
				return
			}
		}
	}
}

// BuildArgs returns the concatenation of the 2 argument lists of arguments.
// It returns an error if the 2 lists intersect.
func BuildArgs(args []monitoringv1.Argument, additionalArgs []monitoringv1.Argument) ([]string, error) {
	if i := ArgumentsIntersection(args, additionalArgs); len(i) > 0 {
		return nil, fmt.Errorf("can't set arguments which are already managed by the operator: %s", strings.Join(i, ","))
	}

	var containerArgs []string
	for _, arg := range append(args, additionalArgs...) {
		if arg.Value != "" {
			containerArgs = append(containerArgs, fmt.Sprintf("--%s=%s", arg.Name, arg.Value))
		} else {
			containerArgs = append(containerArgs, fmt.Sprintf("--%s", arg.Name))
		}
	}

	return containerArgs, nil
}
