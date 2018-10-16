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

package k8sutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/validation"
)

func Test_SanitizeVolumeName(t *testing.T) {
	cases := []struct {
		name     string
		expected string
	}{
		{
			name:     "@$!!@$%!#$%!#$%!#$!#$%%$#@!#",
			expected: "",
		},
		{
			name:     "NAME",
			expected: "name",
		},
		{
			name:     "foo--",
			expected: "foo",
		},
		{
			name:     "foo^%#$bar",
			expected: "foo-bar",
		},
		{
			name:     "fOo^%#$bar",
			expected: "foo-bar",
		},
		{
			name:     strings.Repeat("a", validation.DNS1123LabelMaxLength*2),
			expected: strings.Repeat("a", validation.DNS1123LabelMaxLength),
		},
	}

	for i, c := range cases {
		out := SanitizeVolumeName(c.name)
		if c.expected != out {
			t.Errorf("expected test case %d to be %q but got %q", i, c.expected, out)
		}
	}
}

func TestTruncateVolumeName(t *testing.T) {

	require.Equal(t, "name", TruncateVolumeName("name"))

	{
		// name without duplicates substrings
		val := TruncateVolumeName(strings.Repeat("a", validation.DNS1123LabelMaxLength+1))
		require.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", val)
		require.True(t, len(val) <= validation.DNS1123LabelMaxLength)
	}

	{
		// name with duplicate substrings
		val := TruncateVolumeName(`glusterfs-dynamic-prometheus-release-prometheus-db-prometheus-release-prometheus-0`)
		require.Equal(t, "glusterfs-dynamic-prometheus-release-db-0", val)
		require.True(t, len(val) <= validation.DNS1123LabelMaxLength)
	}

	{
		// name with duplicate substrings in the end of string
		var src string
		for i := 0; i < 5; i++ {
			if len(src) > 0 {
				src += "-"
			}
			src += strings.Repeat("a", 10)
		}

		val := TruncateVolumeName(src)
		require.Equal(t, "aaaaaaaaaa", val)
		require.True(t, len(val) <= validation.DNS1123LabelMaxLength)
	}

	{
		// '-' in the end of the name
		val := TruncateVolumeName(`glusterfs-dynamic-prometheus-release-prometheus-db-prometheus-release-prometheus-0-`)
		require.Equal(t, "glusterfs-dynamic-prometheus-release-db-0", val)
		require.True(t, len(val) <= validation.DNS1123LabelMaxLength)
	}

	// remove duplicates delimiters
	for _, substr := range []string{
		"--",
		"---",
		"----",
		"-----",
		"------",
	} {
		{
			// name with duplicates delimiters
			src := `glusterfs` + substr + `dynamic-prometheus-release-veryveryveryveryverylongname-db` + substr + `0`
			val := TruncateVolumeName(src)
			require.Equal(t, "glusterfs-dynamic-prometheus-release-veryveryveryveryverylongna", val)
			require.True(t, len(val) <= validation.DNS1123LabelMaxLength)
		}

		{
			// '-' in the end of the name
			part1 := strings.Repeat("a", validation.DNS1123LabelMaxLength/2)
			part2 := strings.Repeat("b", validation.DNS1123LabelMaxLength/2)

			src := part1 + substr + part2 + substr // example: "a--b--"
			val := TruncateVolumeName(src)
			require.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", val)
			require.True(t, len(val) <= validation.DNS1123LabelMaxLength)
		}
	}
}
