// Copyright 2024 The prometheus-operator Authors
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

package sortutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSortKeysEmptyMap(t *testing.T) {
	var emptyMap map[int]string
	keys := SortedKeys(emptyMap)
	target := []int{}
	require.Equal(t, keys, target)
}

func TestSortKeys(t *testing.T) {

	intKeys := SortedKeys(map[int]any{
		-10: 6,
		0:   "",
		5:   []byte(""),
		-1:  -9.56,
	})
	require.Equal(t, []int{-10, -1, 0, 5}, intKeys)

	strKeys := SortedKeys(map[string]any{
		"a": 6,
		"c": "",
		"d": []byte(""),
		"b": -9.56,
	})
	require.Equal(t, []string{"a", "b", "c", "d"}, strKeys)

	int32Keys := SortedKeys(map[int32]any{
		-10: 6,
		0:   "",
		5:   []byte(""),
		-1:  -9.56,
	})
	require.Equal(t, []int32{-10, -1, 0, 5}, int32Keys)
}
