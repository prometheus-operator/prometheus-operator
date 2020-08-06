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

package operator

import (
	"testing"
)

type ImageSpec struct {
	Image   string
	Version string
	Tag     string
	SHA     string
}

func TestBuildImagePath(t *testing.T) {
	defaultImageSpec := &ImageSpec{
		Image:   "foo.com/bar",
		Version: "0.0.1",
	}
	// imageWithoutVersion := "myrepo/myimage:123"
	// imageWithVersion := "myhost:9090/myrepo/myimage:0.2"
	// imageWithTag := "myhost:9090/myrepo/myimage:latest"
	// imageWithSHA := "foo/bar@sha256:12345"
	cases := []struct {
		spec     *ImageSpec
		expected string
	}{
		{
			spec:     &ImageSpec{},
			expected: "",
		},
		{
			spec:     defaultImageSpec,
			expected: defaultImageSpec.Image + ":" + defaultImageSpec.Version,
		},
		{
			spec:     &ImageSpec{"myrepo.com/foo", "1.0", "", ""},
			expected: "myrepo.com/foo:1.0",
		},
		{
			spec:     &ImageSpec{"myrepo.com/foo", "1.0", "latest", ""},
			expected: "myrepo.com/foo:latest",
		},
		{
			spec:     &ImageSpec{"myrepo.com/foo", "1.0", "latest", "abcd1234"},
			expected: "myrepo.com/foo@sha256:abcd1234",
		},
	}

	for i, c := range cases {
		result, _ := BuildImagePath(c.spec.Image, c.spec.Version, c.spec.Tag, c.spec.SHA)
		if c.expected != result {
			t.Errorf("expected test case %d to be %q but got %q", i, c.expected, result)
		}
	}
}
