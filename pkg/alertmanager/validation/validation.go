// Copyright 2022 The prometheus-operator Authors
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

package validation

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/prometheus/alertmanager/config"
	"k8s.io/utils/ptr"
)

// ValidateURLPtr validates a URL string pointer.
// If the pointer is nil, it will return no error.
func ValidateURLPtr(url *string) error {
	return validateStringPtr(url, func(url string) error {
		if _, err := ValidateURL(url); err != nil {
			return err
		}
		return nil
	})
}

// ValidateURL against the config.URL
// This could potentially become a regex and be validated via OpenAPI
// but right now, since we know we need to unmarshal into an upstream type
// after conversion, we validate we don't error when doing so.
func ValidateURL(url string) (*config.URL, error) {
	var u config.URL
	err := json.Unmarshal(fmt.Appendf(nil, `"%s"`, url), &u)
	if err != nil {
		return nil, fmt.Errorf("validate url from string failed for %s: %w", url, err)
	}

	return &u, nil
}

// ValidateTemplateURLPtr validates a URL string pointer which can be a Go
// template.
// If the pointer is nil, it will return no error.
func ValidateTemplateURLPtr(url *string) error {
	return validateStringPtr(url, ValidateTemplateURL)
}

// ValidateTemplateURL validates a URL string against the config.URL.
// If the value is a Go template then the function ensures that the template
// definition is valid.
func ValidateTemplateURL(url string) error {
	if strings.Contains(url, "{{") {
		_, err := template.New("").Parse(url)
		if err != nil {
			return err
		}
		return nil
	}

	// Assume that the URL is a secret for safety.
	return ValidateSecretURL(url)
}

func validateStringPtr(s *string, validFn func(string) error) error {
	if ptr.Deref(s, "") == "" {
		return nil
	}

	return validFn(*s)
}

// ValidateSecretURL against config.URL
// This is for URLs which are retrieved from secrets and should not
// logged as part of the err.
func ValidateSecretURL(url string) error {
	var u config.SecretURL

	err := u.UnmarshalJSON(fmt.Appendf(nil, `"%s"`, url))
	if err != nil {
		return fmt.Errorf("validate url from string failed with error: %w", err)
	}

	return nil
}
