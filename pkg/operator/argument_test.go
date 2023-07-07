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
	"testing"

	"golang.org/x/exp/slices"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestBuildArgs(t *testing.T) {
	args := []v1.Argument{
		{Name: "test", Value: "value"},
		{Name: "test2-test", Value: "value2"},
		{Name: "test3.test", Value: "value3"},
	}

	additionalArgs := []v1.Argument{
		{Name: "addtest", Value: "value"},
		{Name: "addtest2-test", Value: "value2"},
		{Name: "addtest3.test", Value: "value3"},
	}

	containerArgs, err := BuildArgs(args, additionalArgs)
	if err != nil {
		t.Errorf("BuildArgs returned an error: %s", err.Error())
	}

	for _, arg := range args {
		argString := fmt.Sprintf("--%s=%s", arg.Name, arg.Value)
		if !slices.Contains(containerArgs, argString) {
			t.Fatalf("expected containerArgs to contain arg %v, got %v", argString, containerArgs)
		}
	}

	for _, arg := range additionalArgs {
		argString := fmt.Sprintf("--%s=%s", arg.Name, arg.Value)
		if !slices.Contains(containerArgs, argString) {
			t.Fatalf("expected containerArgs to contain additionalArg %v, got %v", argString, containerArgs)
		}
	}
}
