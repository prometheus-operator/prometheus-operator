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

package assets

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// RefTracker is a set-based struct which records the references to secrets and
// configmaps.
type RefTracker map[string]struct{}

// insert records the object in the tracker.
func (r RefTracker) insert(obj interface{}) {
	key, err := assetKeyFunc(obj)
	if err != nil {
		return
	}
	r[key] = struct{}{}
}

// Has returns true if the tracker knows about the given object.
func (r RefTracker) Has(obj runtime.Object) bool {
	key, err := assetKeyFunc(obj)
	if err != nil {
		return false
	}

	_, found := r[key]
	return found
}
