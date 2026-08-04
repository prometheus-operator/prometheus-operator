// Copyright The prometheus-operator Authors
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
	"iter"

	"k8s.io/client-go/tools/cache"
)

func StoresIter[T any](stores ...cache.Store) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, store := range stores {
			for _, obj := range store.List() {
				v, ok := obj.(T)
				if !ok {
					continue
				}
				if !yield(v) {
					return
				}
			}
		}
	}
}
