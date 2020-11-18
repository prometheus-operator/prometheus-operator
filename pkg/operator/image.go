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
	"fmt"
	"strings"

	dockerref "github.com/docker/distribution/reference"
)

// BuildImagePath builds a container image path based on
// the given parameters.
// Return specImage if not empty.
// If image contains a tag or digest then image will be returned.
// Otherwise, return image suffixed by either SHA, tag or version(in that order).
// Inspired by kubernetes code handling of image building.
func BuildImagePath(specImage, image, version, tag, sha string) (string, error) {
	if strings.TrimSpace(specImage) != "" {
		return specImage, nil
	}
	named, err := dockerref.ParseNormalizedNamed(image)
	if err != nil {
		return "", fmt.Errorf("couldn't parse image reference %q: %v", image, err)
	}
	_, isTagged := named.(dockerref.Tagged)
	_, isDigested := named.(dockerref.Digested)
	if isTagged || isDigested {
		return image, nil
	}

	if sha != "" {
		return fmt.Sprintf("%s@sha256:%s", image, sha), nil
	} else if tag != "" {
		imageTag, err := dockerref.WithTag(named, tag)
		if err != nil {
			return "", err
		}
		return imageTag.String(), nil
	} else if version != "" {
		return image + ":" + version, nil
	}

	return image, nil
}

// StringValOrDefault returns the default val if the
// given string is empty/whitespace.
// Otherwise returns the value of the string..
func StringValOrDefault(val, defaultVal string) string {
	if strings.TrimSpace(val) == "" {
		return defaultVal
	}
	return val
}

// StringPtrValOrDefault returns the default val if the
// given string pointer is nil points to an empty/whitespace string.
// Otherwise returns the value of the string.
func StringPtrValOrDefault(val *string, defaultVal string) string {
	if val == nil {
		return defaultVal
	}
	return StringValOrDefault(*val, defaultVal)
}
