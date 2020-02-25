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
	"regexp"
	"strings"

	"github.com/blang/semver"
)

// ImageConfig contains config for creating a container image URL
type ImageConfig struct {
	Image            *string
	DefaultImage     string
	BaseImage        string
	DefaultBaseImage string
	Version          string
	DefaultVersion   string
	Tag              string
	SHA              string
}

// ContainerImageURL gets or generates the URL to download a container image.
// If the `Image` field is not empty, this should override any other setting.
// If no fields are specified, the default image URL is returned.
// Otherwise, use a combination of the default base image
// and the deprecated `Version`, `Tag`, and `SHA`, fields.
// Returns the image URL and the version of the component.
func ContainerImageURL(spec *ImageConfig) string {
	if spec.Image != nil && *spec.Image != "" {
		return ApplyImageVersion(*spec.Image, spec.Version)
	}
	if spec.BaseImage == "" && spec.Version == "" &&
		spec.Tag == "" && spec.SHA == "" {
		return spec.DefaultImage
	}

	// Version is used by default.
	// If the tag is specified, we use the tag to identify the container image.
	// If the sha is specified, we use the sha to identify the container image,
	// as it has even stronger immutable guarantees to identify the image.
	baseImage := spec.DefaultBaseImage
	if spec.BaseImage != "" {
		baseImage = spec.BaseImage
	}
	version := spec.DefaultVersion
	if spec.Version != "" {
		version = spec.Version
	}
	image := fmt.Sprintf("%s:%s", baseImage, version)
	if spec.Tag != "" {
		image = fmt.Sprintf("%s:%s", baseImage, spec.Tag)
	}
	if spec.SHA != "" {
		image = fmt.Sprintf("%s@sha256:%s", baseImage, spec.SHA)
	}
	return image
}

// ParseVersion tries to parse the given version.
// If the version is blank, will try to parse the version
// in the imageURL.
// Otherwise will parse the defaultVersion.
func ParseVersion(version, imageURL, defaultVersion string) (semver.Version, error) {
	if version != "" {
		return semver.ParseTolerant(version)
	}
	imageVersion := GetImageVersion(imageURL)
	if imageVersion != "" {
		v, err := semver.ParseTolerant(imageVersion)
		if err == nil {
			return v, nil
		}
	}
	return semver.ParseTolerant(defaultVersion)
}

var (
	containerURLRegex = regexp.MustCompile(`^([\w-.:]+/)?([\w-.]+/)?[\w-.]+([:]([\w-.]+))?$`)
)

// GetImageVersion gets the version/tag portion of a container
// image URL.  If there is no version specified, or if the url
// specifies a digest instead of a version or tag, an empty string
// is returned.
func GetImageVersion(imageURL string) string {
	match := containerURLRegex.FindStringSubmatch(imageURL)
	if len(match) < 5 {
		return ""
	}
	return match[4]
}

// ApplyImageVersion applies the given version to the imageURL.\
// If the URL does not contain a version, the given version will
// be appended.
// If the URL already includes valid semver version, it will be replaced.
// If the URL contains an invalid semver version such as a tag or SHA,
// the original URL will be returned unmodified.
func ApplyImageVersion(imageURL, version string) string {
	if strings.TrimSpace(version) == "" {
		return imageURL
	}
	oldVersion := GetImageVersion(imageURL)
	if oldVersion == "" {
		if !strings.HasSuffix(imageURL, ":") {
			imageURL = imageURL + ":"
		}
		return imageURL + version
	}
	_, err := semver.ParseTolerant(oldVersion)
	if err == nil {
		versionIndex := len(imageURL) - len(oldVersion)
		return imageURL[:versionIndex] + version
	}
	return imageURL
}
