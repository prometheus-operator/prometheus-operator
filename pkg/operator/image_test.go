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

func TestContainerImageURL(t *testing.T) {
	defaultImageSpec := &ImageConfig{
		DefaultBaseImage: "foo/bar",
		DefaultImage:     "foo/baz",
		DefaultVersion:   "0.0.1",
	}
	testImage := "myrepo/myimage"
	testImageWithVersion := "myhost:9090/myrepo/myimage:0.3"
	cases := []struct {
		spec     *ImageConfig
		expected string
	}{
		{
			spec:     &ImageConfig{},
			expected: "foo/baz",
		},
		{
			spec: &ImageConfig{
				Image: &testImage,
			},
			expected: testImage,
		},
		{
			spec: &ImageConfig{
				Image: &testImageWithVersion,
			},
			expected: testImageWithVersion,
		},
		{
			spec: &ImageConfig{
				BaseImage: "foo/baseimage",
			},
			expected: "foo/baseimage:0.0.1",
		},
		{
			spec: &ImageConfig{
				Tag: "release-v1",
			},
			expected: "foo/bar:release-v1",
		},
		{
			spec: &ImageConfig{
				SHA: "12345",
			},
			expected: "foo/bar@sha256:12345",
		},
		{
			spec: &ImageConfig{
				Version: "2.0",
			},
			expected: "foo/baz",
		},
		{
			spec: &ImageConfig{
				Image:   &testImageWithVersion,
				Version: "2.0",
			},
			expected: testImageWithVersion,
		},
	}

	for i, c := range cases {
		testSpec := mergeImageSpecs(defaultImageSpec, c.spec)
		result := ContainerImageURL(testSpec)
		if c.expected != result {
			t.Errorf("expected test case %d to be %q but got %q", i, c.expected, result)
		}
	}
}

func mergeImageSpecs(defaults, overrides *ImageConfig) *ImageConfig {
	spec := *defaults
	if overrides.DefaultBaseImage != "" {
		spec.DefaultBaseImage = overrides.DefaultBaseImage
	}
	if overrides.BaseImage != "" {
		spec.BaseImage = overrides.BaseImage
	}
	if overrides.DefaultImage != "" {
		spec.DefaultImage = overrides.DefaultImage
	}
	if overrides.Image != nil {
		spec.Image = overrides.Image
	}
	if overrides.DefaultVersion != "" {
		spec.DefaultVersion = overrides.DefaultVersion
	}
	if overrides.Tag != "" {
		spec.Tag = overrides.Tag
	}
	if overrides.SHA != "" {
		spec.SHA = overrides.SHA
	}
	return &spec
}

func TestGetImageVersion(t *testing.T) {
	cases := []struct {
		image    string
		expected string
	}{
		{
			image:    "myimage",
			expected: "",
		},
		{
			image:    "myimage:0.2",
			expected: "0.2",
		},
		{
			image:    "myrepo/myimage:0.3",
			expected: "0.3",
		},
		{
			image:    "myhost.com/myrepo/myimage:0.4",
			expected: "0.4",
		},
		{
			image:    "myhost.com:8080/myrepo/myimage",
			expected: "",
		},
		{
			image:    "myhost.com:8080/myrepo/myimage:0.6-beta1",
			expected: "0.6-beta1",
		},
		{
			image:    "myimage@sha256:45b23dee",
			expected: "",
		},
		{
			image:    "myhost.com:8080/myrepo/myimage@sha256:45b23dee0",
			expected: "",
		},
		{
			image:    "#$&*#&56garbage#($@&($&(#(@(^%)",
			expected: "",
		},
	}

	for i, c := range cases {
		result := GetImageVersion(c.image)
		if c.expected != result {
			t.Errorf("expected test case %d to be %q but got %q", i, c.expected, result)
		}
	}
}

func TestApplyImageVersion(t *testing.T) {
	cases := []struct {
		image    string
		version  string
		expected string
	}{
		{
			image:    "myimage",
			version:  "0.1",
			expected: "myimage:0.1",
		},
		{
			image:    "myrepo/myimage",
			version:  "0.2",
			expected: "myrepo/myimage:0.2",
		},
		{
			image:    "myrepo/myimage:v1.0",
			version:  "0.3",
			expected: "myrepo/myimage:0.3",
		},
		{
			image:    "myrepo/myimage:0.4",
			version:  "",
			expected: "myrepo/myimage:0.4",
		},
		{
			image:    "myhost.com/myrepo/myimage:2.0.1",
			version:  "v0.5.1",
			expected: "myhost.com/myrepo/myimage:v0.5.1",
		},
		{
			image:    "myhost.com:8080/myrepo/myimage",
			version:  "",
			expected: "myhost.com:8080/myrepo/myimage",
		},
		{
			image:    "myhost.com:8080/myrepo/myimage",
			version:  "0.6",
			expected: "myhost.com:8080/myrepo/myimage:0.6",
		},
	}

	for i, c := range cases {
		result := ApplyImageVersion(c.image, c.version)
		if c.expected != result {
			t.Errorf("expected test case %d to be %q but got %q", i, c.expected, result)
		}
	}
}
