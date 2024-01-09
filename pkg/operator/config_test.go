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
	"errors"
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

func TestStringSet(t *testing.T) {
	var s StringSet

	require.Error(t, s.Set("a,b,c"))

	s = StringSet{}

	require.NoError(t, s.Set("a,b,c"))
	require.Equal(t, len(s), 3)
	require.Equal(t, s.String(), "a,b,c")
	for _, k := range []string{"a", "b", "c"} {
		_, found := s[k]
		require.True(t, found)
	}
}

func TestImageVersion(t *testing.T) {
	for _, tc := range []struct {
		name            string
		reloaderConfig  ContainerConfig
		expectedVersion string
		err             error
	}{
		{
			name: "latest version",
			reloaderConfig: ContainerConfig{
				Image: "quay.io/prometheus-operator/prometheus-config-reloader:latest",
			},
			expectedVersion: "latest",
			err:             nil,
		},
		{
			name: "normal version",
			reloaderConfig: ContainerConfig{
				Image: "quay.io/prometheus-operator/prometheus-config-reloader:v0.69.0",
			},
			expectedVersion: "v0.69.0",
			err:             nil,
		},
		{
			name: "empty version",
			reloaderConfig: ContainerConfig{
				Image: "",
			},
			expectedVersion: "",
			err:             errors.New("invalid reference format"),
		},
		{
			name: "illegal image version",
			reloaderConfig: ContainerConfig{
				Image: "adasdf",
			},
			expectedVersion: "",
			err:             errors.New("cannot parse image tag"),
		},
		{
			name: "image sha256 version",
			reloaderConfig: ContainerConfig{
				Image: "quay.io/prometheus-operator/prometheus-config-reloader@sha256:21852ba2d7876259999d8e54561ece2907696f3b499644b1c3cb6d9e07c9a317",
			},
			expectedVersion: "",
			err:             errors.New("cannot parse image tag"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			version, err := tc.reloaderConfig.ImageVersion()
			require.Equal(t, tc.expectedVersion, version)
			require.Equal(t, tc.err, err)
		})
	}
}
