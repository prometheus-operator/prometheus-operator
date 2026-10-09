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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildImagePath(t *testing.T) {
	for _, tc := range []struct {
		name      string
		baseImage string
		image     string
		version   string
		tag       string
		sha       string
		expected  string
		err       bool
	}{
		{
			name: "error when no image and no base image are provided",
			err:  true,
		},
		{
			name:      "use base image without version",
			baseImage: "example.com/bar",
			err:       true,
		},
		{
			name:      "use base image reference including a tag",
			baseImage: "example.com/bar:v0.0.1",
			expected:  "example.com/bar:v0.0.1",
		},
		{
			name:      "use base image reference including a invalid tag",
			baseImage: "example.com/bar::v0.0.1",
			err:       true,
		},
		{
			name:      "use base image reference including a sha",
			baseImage: "example.com/bar@sha256:b05f4d42b3026e0f80cf6d5f355d6218f739bc5877b2352cd7358da0b8dcb808",
			expected:  "example.com/bar@sha256:b05f4d42b3026e0f80cf6d5f355d6218f739bc5877b2352cd7358da0b8dcb808",
		},
		{
			name:      "use base image reference including an invalid sha",
			baseImage: "example.com/bar@sha256:b05f4",
			err:       true,
		},
		{
			name:      "use base image with version",
			baseImage: "example.com/bar",
			version:   "v0.0.1",
			expected:  "example.com/bar:v0.0.1",
		},
		{
			name:      "use image with tag",
			baseImage: "example.com/bar",
			version:   "v1.0.0",
			tag:       "latest",
			expected:  "example.com/bar:latest",
		},
		{
			name:      "use base image with sha",
			baseImage: "example.com/bar",
			version:   "v1.0.0",
			tag:       "latest",
			sha:       "abcd1234",
			expected:  "example.com/bar@sha256:abcd1234",
		},
		{
			name:      "use valid image reference",
			image:     "example.com/foo",
			baseImage: "example.com/bar",
			version:   "v1.0.0",
			tag:       "latest",
			sha:       "abcd1234",
			expected:  "example.com/foo",
		},
		{
			name:      "use image reference even if invalid",
			image:     "example.com/foo:sha256:abcd1234",
			baseImage: "example.com/bar",
			version:   "v1.0.0",
			tag:       "latest",
			sha:       "abcd1234",
			expected:  "example.com/foo:sha256:abcd1234",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := BuildImagePath(tc.image, tc.baseImage, tc.version, tc.tag, tc.sha)
			if tc.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}
