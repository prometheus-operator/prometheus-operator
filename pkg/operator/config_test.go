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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	var m Map

	require.Equal(t, "", m.String())

	require.Equal(t, map[string]string{"foo": "xxx", "foo3": "bar3"}, m.Merge(map[string]string{"foo": "xxx", "foo3": "bar3"}))

	require.NoError(t, m.Set("foo2=bar2,foo=bar"))
	require.Equal(t, 2, len(m))
	require.Equal(t, []string{"foo", "foo2"}, m.SortedKeys())
	require.Equal(t, "foo=bar,foo2=bar2", m.String())

	require.Equal(t, map[string]string{"foo": "bar", "foo2": "bar2", "foo3": "bar3"}, m.Merge(map[string]string{"foo": "xxx", "foo3": "bar3"}))
}
