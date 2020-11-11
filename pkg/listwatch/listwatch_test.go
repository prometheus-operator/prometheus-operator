// Copyright 2016 The prometheus-operator Authors
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

package listwatch

import (
	"testing"
)

func TestIdenticalNamespaces(t *testing.T) {
	for _, tc := range []struct {
		a, b map[string]struct{}
		ret  bool
	}{
		{
			a: map[string]struct{}{
				"foo": struct{}{},
			},
			b: map[string]struct{}{
				"foo": struct{}{},
			},
			ret: true,
		},
		{
			a: map[string]struct{}{
				"foo": struct{}{},
			},
			b: map[string]struct{}{
				"bar": struct{}{},
			},
			ret: false,
		},
	} {
		tc := tc
		t.Run("", func(t *testing.T) {
			ret := IdenticalNamespaces(tc.a, tc.b)
			if ret != tc.ret {
				t.Fatalf("expecting IdenticalNamespaces() to return %v, got %v", tc.ret, ret)
			}
		})
	}
}
