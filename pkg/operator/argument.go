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

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func intersection(a, b []string) (i []string) {
	m := make(map[string]struct{})

	for _, item := range a {
		m[item] = struct{}{}
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			i = append(i, item)
		}

		negatedItem := strings.TrimPrefix(item, "no-")
		if item == negatedItem {
			negatedItem = fmt.Sprintf("no-%s", item)
		}

		if _, ok := m[negatedItem]; ok {
			i = append(i, item)
		}
	}
	return i
}

func extractArgKeys(args []monitoringv1.Argument) []string {
	var k []string
	for _, arg := range args {
		key := arg.Name
		k = append(k, key)
	}

	return k
}

// BuildArgs takes a list of arguments and a list of additional arguments and returns a []string to use in a container Args
func BuildArgs(args []monitoringv1.Argument, additionalArgs []monitoringv1.Argument) ([]string, error) {
	var containerArgs []string

	argKeys := extractArgKeys(args)
	additionalArgKeys := extractArgKeys(additionalArgs)

	i := intersection(argKeys, additionalArgKeys)
	if len(i) > 0 {
		return nil, fmt.Errorf("can't set arguments which are already managed by the operator: %s", strings.Join(i, ","))
	}

	args = append(args, additionalArgs...)

	for _, arg := range args {
		if arg.Value != "" {
			containerArgs = append(containerArgs, fmt.Sprintf("--%s=%s", arg.Name, arg.Value))
		} else {
			containerArgs = append(containerArgs, fmt.Sprintf("--%s", arg.Name))

		}
	}

	return containerArgs, nil
}
