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
	v1 "k8s.io/api/core/v1"
)

func TestMap(t *testing.T) {
	var m Map

	require.Equal(t, "", m.String())

	require.Equal(t, map[string]string{"foo": "xxx", "foo3": "bar3"}, m.Merge(map[string]string{"foo": "xxx", "foo3": "bar3"}))

	require.NoError(t, m.Set("foo2=bar2,foo=bar"))
	require.Len(t, m, 2)
	require.Equal(t, []string{"foo", "foo2"}, m.SortedKeys())
	require.Equal(t, "foo=bar,foo2=bar2", m.String())

	require.Equal(t, map[string]string{"foo": "bar", "foo2": "bar2", "foo3": "bar3"}, m.Merge(map[string]string{"foo": "xxx", "foo3": "bar3"}))
}

func TestFieldSelector(t *testing.T) {
	for _, tc := range []struct {
		value string
		fail  bool
	}{
		{
			value: "",
		},
		{
			value: "foo = bar",
		},
		{
			value: "foo",
			fail:  true,
		},
	} {
		t.Run(tc.value, func(t *testing.T) {
			fs := new(FieldSelector)

			err := fs.Set(tc.value)
			if tc.fail {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestLabelSelector(t *testing.T) {
	for _, tc := range []struct {
		value string
		fail  bool
	}{
		{
			value: "",
		},
		{
			value: "foo in (bar)",
		},
		{
			value: "foo in",
			fail:  true,
		},
	} {
		t.Run(tc.value, func(t *testing.T) {
			ls := new(LabelSelector)

			err := ls.Set(tc.value)
			if tc.fail {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestNodeAddressPriority(t *testing.T) {
	p := new(NodeAddressPriority)
	require.Equal(t, "internal", p.String())

	require.NoError(t, p.Set("internal"))
	require.Equal(t, "internal", p.String())

	require.NoError(t, p.Set("external"))
	require.Equal(t, "external", p.String())

	require.Error(t, p.Set("foo"))
}

func TestStringSet(t *testing.T) {
	var s StringSet

	require.Error(t, s.Set("a,b,c"))

	s = StringSet{}

	require.NoError(t, s.Set("a,b,c"))
	require.Len(t, s, 3)
	require.Equal(t, "a,b,c", s.String())
	for _, k := range []string{"a", "b", "c"} {
		_, found := s[k]
		require.True(t, found)
	}
}

func TestNamespaceFinalize(t *testing.T) {
	for _, tc := range []struct {
		name string
		ns   Namespaces

		exp Namespaces
		err bool
	}{
		{
			name: "allow all",
			ns:   Namespaces{},
			exp: Namespaces{
				AllowList: StringSet{
					"": struct{}{},
				},
				PrometheusAllowList: StringSet{
					"": struct{}{},
				},
				AlertmanagerAllowList: StringSet{
					"": struct{}{},
				},
				AlertmanagerConfigAllowList: StringSet{
					"": struct{}{},
				},
				ThanosRulerAllowList: StringSet{
					"": struct{}{},
				},
			},
		},
		{
			name: "allow one",
			ns: Namespaces{
				AllowList: StringSet{
					"foo": struct{}{},
				},
			},
			exp: Namespaces{
				AllowList: StringSet{
					"foo": struct{}{},
				},
				PrometheusAllowList: StringSet{
					"foo": struct{}{},
				},
				AlertmanagerAllowList: StringSet{
					"foo": struct{}{},
				},
				AlertmanagerConfigAllowList: StringSet{
					"foo": struct{}{},
				},
				ThanosRulerAllowList: StringSet{
					"foo": struct{}{},
				},
			},
		},
		{
			name: "deny one",
			ns: Namespaces{
				DenyList: StringSet{
					"foo": struct{}{},
				},
			},
			exp: Namespaces{
				AllowList: StringSet{
					"": struct{}{},
				},
				DenyList: StringSet{
					"foo": struct{}{},
				},
				PrometheusAllowList: StringSet{
					"": struct{}{},
				},
				AlertmanagerAllowList: StringSet{
					"": struct{}{},
				},
				AlertmanagerConfigAllowList: StringSet{
					"": struct{}{},
				},
				ThanosRulerAllowList: StringSet{
					"": struct{}{},
				},
			},
		},
		{
			name: "allow and deny forbidden",
			ns: Namespaces{
				AllowList: StringSet{
					"bar": struct{}{},
				},
				DenyList: StringSet{
					"foo": struct{}{},
				},
			},
			err: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.ns.Finalize()
			if tc.err {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.exp, tc.ns)
		})
	}
}

func TestMergeAllowLists(t *testing.T) {
	for _, tc := range []struct {
		name string
		a    StringSet
		b    StringSet

		exp StringSet
	}{
		{
			a: StringSet{
				v1.NamespaceAll: struct{}{},
			},
			b: StringSet{
				"foo": struct{}{},
			},
			exp: StringSet{
				v1.NamespaceAll: struct{}{},
			},
		},
		{
			a: StringSet{
				"foo": struct{}{},
			},
			b: StringSet{
				v1.NamespaceAll: struct{}{},
			},
			exp: StringSet{
				v1.NamespaceAll: struct{}{},
			},
		},
		{
			a: StringSet{
				"foo": struct{}{},
			},
			b: StringSet{
				"bar": struct{}{},
			},
			exp: StringSet{
				"foo": struct{}{},
				"bar": struct{}{},
			},
		},
		{
			a: StringSet{
				"foo": struct{}{},
				"bar": struct{}{},
			},
			b: StringSet{
				"bar":  struct{}{},
				"fred": struct{}{},
			},
			exp: StringSet{
				"foo":  struct{}{},
				"bar":  struct{}{},
				"fred": struct{}{},
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			got := MergeAllowLists(tc.a, tc.b)
			require.Equal(t, tc.exp, got)
		})
	}
}
