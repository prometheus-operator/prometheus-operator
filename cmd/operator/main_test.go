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

package main

import (
	"strings"
	"testing"
)

func TestNamespacesType(t *testing.T) {
	var ns namespaces
	if ns.String() != "" {
		t.Errorf("incorrect string value for nil namespaces, want: empty string, got %v", ns.String())
	}

	val := "a,b,c"
	err := ns.Set(val)
	if err == nil {
		t.Error("expected error for nil namespaces")
	}

	ns = namespaces{}
	ns.Set(val)
	if len(ns) != 3 {
		t.Errorf("incorrect length of namespaces, want: %v, got: %v", 3, len(ns))
	}

	for _, next := range strings.Split(val, ",") {
		if _, ok := ns[next]; !ok {
			t.Errorf("namespace not in map, want: %v, not in map: %v", next, map[string]struct{}(ns))
		}
	}

}
