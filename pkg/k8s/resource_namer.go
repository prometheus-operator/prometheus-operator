// Copyright 2026 The prometheus-operator Authors
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
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/cespare/xxhash/v2"
	"k8s.io/apimachinery/pkg/util/validation"
)

var invalidDNS1123Characters = regexp.MustCompile("[^-a-z0-9]+")

// ResourceNamer knows how to generate valid names for various Kubernetes resources.
type ResourceNamer struct {
	prefix string
}

// NewResourceNamerWithPrefix returns a ResourceNamer that adds a prefix
// followed by an hyphen character to all resource names.
func NewResourceNamerWithPrefix(p string) ResourceNamer {
	return ResourceNamer{prefix: p}
}

func (rn ResourceNamer) sanitizedLabel(name string) string {
	if rn.prefix != "" {
		name = strings.TrimRight(rn.prefix, "-") + "-" + name
	}

	name = strings.ToLower(name)
	name = invalidDNS1123Characters.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")

	return name
}

func isValidDNS1123Label(name string) error {
	if errs := validation.IsDNS1123Label(name); len(errs) > 0 {
		return errors.New(strings.Join(errs, ","))
	}

	return nil
}

// UniqueDNS1123Label returns a name that is a valid DNS-1123 label.
// The returned name has a hash-based suffix to ensure uniqueness in case the
// input name exceeds the 63-chars limit.
func (rn ResourceNamer) UniqueDNS1123Label(name string) (string, error) {
	// Hash the name and append the 8 first characters of the hash
	// value to the resulting name to ensure that 2 names longer than
	// DNS1123LabelMaxLength return unique names.
	// E.g. long-63-chars-abc, long-63-chars-XYZ may be added to
	// name since they are trimmed at long-63-chars, there will be 2
	// resource entries with the same name.
	// In practice, the hash is computed for the full name then trimmed to
	// the first 8 chars and added to the end:
	// * long-63-chars-abc -> first-54-chars-deadbeef
	// * long-63-chars-XYZ -> first-54-chars-d3adb33f
	xxh := xxhash.New()
	if _, err := xxh.Write([]byte(name)); err != nil {
		return "", err
	}

	h := fmt.Sprintf("-%x", xxh.Sum64())
	h = h[:9]

	name = rn.sanitizedLabel(name)

	if len(name) > validation.DNS1123LabelMaxLength-9 {
		name = name[:validation.DNS1123LabelMaxLength-9]
	}

	name = name + h
	if errs := validation.IsDNS1123Label(name); len(errs) > 0 {
		return "", errors.New(strings.Join(errs, ","))
	}

	return name, isValidDNS1123Label(name)
}

// DNS1123Label returns a name that is a valid DNS-1123 label.
// It will sanitize a name, removing invalid characters and if
// the name is bigger than 63 chars it will truncate it.
func (rn ResourceNamer) DNS1123Label(name string) (string, error) {
	name = rn.sanitizedLabel(name)

	if len(name) > validation.DNS1123LabelMaxLength {
		name = name[:validation.DNS1123LabelMaxLength]
	}

	return name, isValidDNS1123Label(name)
}
