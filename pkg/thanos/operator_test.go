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

package thanos

import (
	"testing"
)

func TestListOptions(t *testing.T) {
	for i := 0; i < 1000; i++ {
		o := ListOptions("test")
		if o.LabelSelector != "app.kubernetes.io/name=thanos-ruler,thanos-ruler=test" && o.LabelSelector != "thanos-ruler=test,app.kubernetes.io/name=thanos-ruler" {
			t.Fatalf("LabelSelector not computed correctly\n\nExpected: \"app.kubernetes.io/name=thanos-ruler,thanos-ruler=test\"\n\nGot:      %#+v", o.LabelSelector)
		}
	}
}
