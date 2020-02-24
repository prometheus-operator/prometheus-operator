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
)

// BuildImagePath builds a container image path based on
// the given parameters.
// baseImage and version are used by default.
// If the tag is specified, we use the tag to identify the container image.
// If the sha is specified, we use the sha to identify the container image,
// as it has even stronger immutable guarantees to identify the image.
func BuildImagePath(baseImage, version, tag, sha string) string {
	image := baseImage
	if version != "" {
		image = fmt.Sprintf("%s:%s", baseImage, version)
	}
	if tag != "" {
		image = fmt.Sprintf("%s:%s", baseImage, tag)
	}
	if sha != "" {
		image = fmt.Sprintf("%s@sha256:%s", baseImage, sha)
	}
	return image
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
