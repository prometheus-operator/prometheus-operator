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

package k8s

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/validation"
)

func TestUniqueVolumeName(t *testing.T) {
	cases := []struct {
		prefix   string
		name     string
		expected string
		err      bool
	}{
		{
			name:     "@$!!@$%!#$%!#$%!#$!#$%%$#@!#",
			expected: "",
			err:      true,
		},
		{
			name:     "NAME",
			expected: "name-4cfd3574",
		},
		{
			name:     "foo--",
			expected: "foo-e705c7c8",
		},
		{
			name:     "foo^%#$bar",
			expected: "foo-bar-f3e212b1",
		},
		{
			name:     "fOo^%#$bar",
			expected: "foo-bar-ee5c3c18",
		},
		{
			name: strings.Repeat("a", validation.DNS1123LabelMaxLength*2),
			expected: strings.Repeat("a", validation.DNS1123LabelMaxLength-9) +
				"-4ed69ce2",
		},
		{
			prefix:   "with-prefix",
			name:     "name",
			expected: "with-prefix-name-6c5f7b2e",
		},
		{
			prefix:   "with-prefix-",
			name:     "name",
			expected: "with-prefix-name-6c5f7b2e",
		},
		{
			prefix:   "with-prefix",
			name:     strings.Repeat("a", validation.DNS1123LabelMaxLength*2),
			expected: "with-prefix-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-4ed69ce2",
		},
	}

	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rn := ResourceNamer{prefix: c.prefix}

			out, err := rn.UniqueDNS1123Label(c.name)
			if c.err {
				if err == nil {
					t.Errorf("expecting error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("expecting no error, got %v", err)
			}

			if c.expected != out {
				t.Errorf("expected test case %d to be %q but got %q", i, c.expected, out)
			}
		})
	}
}

func TestUniqueVolumeNameCollision(t *testing.T) {
	// a<63>-foo
	foo := strings.Repeat("a", validation.DNS1123LabelMaxLength) + "foo"
	// a<63>-bar
	bar := strings.Repeat("a", validation.DNS1123LabelMaxLength) + "bar"

	rn := ResourceNamer{}

	fooSanitized, err := rn.UniqueDNS1123Label(foo)
	if err != nil {
		t.Errorf("expecting no error, got %v", err)
	}

	barSanitized, err := rn.UniqueDNS1123Label(bar)
	if err != nil {
		t.Errorf("expecting no error, got %v", err)
	}

	require.NotEqual(t, fooSanitized, barSanitized, "expected sanitized volume name of %q and %q to be different but got %q", foo, bar, fooSanitized)
}
