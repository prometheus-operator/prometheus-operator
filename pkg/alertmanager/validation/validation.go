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
	"errors"
	"fmt"
	"net"
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

// VictorOpsReservedFields contains the set of field names that are reserved
// and cannot be used in VictorOps custom fields.
// See: https://github.com/prometheus/alertmanager/blob/a7f9fdadbecbb7e692d2cd8d3334e3d6de1602e1/config/notifiers.go#L497
var VictorOpsReservedFields = map[string]struct{}{
	"routing_key":         {},
	"message_type":        {},
	"state_message":       {},
	"entity_display_name": {},
	"monitoring_tool":     {},
	"entity_id":           {},
	"entity_state":        {},
}

// ValidateVictorOpsCustomFields checks that no custom field uses a reserved name.
func ValidateVictorOpsCustomFields(customFieldKeys []string) error {
	for _, key := range customFieldKeys {
		if _, ok := VictorOpsReservedFields[key]; ok {
			return fmt.Errorf("usage of reserved word %q is not allowed in custom fields", key)
		}
	}
	return nil
}

// ValidateEmailHeaders checks for duplicate headers in email configuration.
// Header names are case-insensitive, so we normalize to lowercase for comparison.
func ValidateEmailHeaders(headerKeys []string) error {
	normalizedHeaders := make(map[string]struct{}, len(headerKeys))
	for _, key := range headerKeys {
		normalized := strings.ToLower(key)
		if _, ok := normalizedHeaders[normalized]; ok {
			return fmt.Errorf("duplicate header %q", normalized)
		}
		normalizedHeaders[normalized] = struct{}{}
	}
	return nil
}

// ValidateSmarthost checks that the smarthost string is a valid host:port combination.
func ValidateSmarthost(smarthost string) error {
	_, _, err := net.SplitHostPort(smarthost)
	if err != nil {
		return fmt.Errorf("invalid 'smarthost' %s: %w", smarthost, err)
	}
	return nil
}

// ValidatePushoverConfig checks that either userKey or userKeyFile is configured,
// either token or tokenFile is configured, and that HTML and Monospace options
// are not both enabled (they are mutually exclusive).
func ValidatePushoverConfig(hasUserKey, hasUserKeyFile, hasToken, hasTokenFile, html, monospace bool) error {
	if !hasUserKey && !hasUserKeyFile {
		return fmt.Errorf("one of userKey or userKeyFile must be configured")
	}

	if !hasToken && !hasTokenFile {
		return fmt.Errorf("one of token or tokenFile must be configured")
	}

	if html && monospace {
		return fmt.Errorf("html and monospace options are mutually exclusive")
	}

	return nil
}

// ValidateWebhookConfig checks that either url or urlSecret is configured.
func ValidateWebhookConfig(hasURL, hasURLSecret bool) error {
	if !hasURL && !hasURLSecret {
		return errors.New("one of 'url' or 'urlSecret' must be specified")
	}
	return nil
}

// ValidatePagerDutyConfig checks that either routingKey or serviceKey is configured.
func ValidatePagerDutyConfig(hasRoutingKey, hasServiceKey bool) error {
	if !hasRoutingKey && !hasServiceKey {
		return errors.New("one of 'routingKey' or 'serviceKey' is required")
	}
	return nil
}

// ValidateRouteReceiver checks that a route has a valid receiver reference.
// For top-level routes, the receiver is required. For child routes, it's optional.
func ValidateRouteReceiver(receiver string, receivers map[string]struct{}, isTopLevel bool) error {
	if receiver == "" {
		if isTopLevel {
			return errors.New("root route must define a receiver")
		}
		return nil
	}
	if _, found := receivers[receiver]; !found {
		return fmt.Errorf("receiver %q not found", receiver)
	}
	return nil
}

// ValidateRouteGroupBy checks the groupBy field for a route.
// It ensures no duplicates and that "..." is the sole value if present.
func ValidateRouteGroupBy(groupBy []string) error {
	if len(groupBy) == 0 {
		return nil
	}

	groupedBy := make(map[string]struct{}, len(groupBy))
	for _, str := range groupBy {
		if _, found := groupedBy[str]; found {
			return fmt.Errorf("duplicate values not permitted in route 'groupBy': %v", groupBy)
		}
		groupedBy[str] = struct{}{}
	}
	if _, found := groupedBy["..."]; found && len(groupBy) > 1 {
		return fmt.Errorf("'...' must be a sole value in route 'groupBy': %v", groupBy)
	}
	return nil
}

// ValidateTimeIntervalReference checks that a named time interval exists.
func ValidateTimeIntervalReference(name string, timeIntervals map[string]struct{}, isMute bool) error {
	if _, found := timeIntervals[name]; !found {
		if isMute {
			return fmt.Errorf("mute time interval %q not found", name)
		}
		return fmt.Errorf("time interval %q not found", name)
	}
	return nil
}

// ValidateSNSConfig checks that exactly one of targetARN, topicARN, or phoneNumber is configured.
func ValidateSNSConfig(hasTargetARN, hasTopicARN, hasPhoneNumber bool) error {
	// XOR logic: exactly one must be true
	count := 0
	if hasTargetARN {
		count++
	}
	if hasTopicARN {
		count++
	}
	if hasPhoneNumber {
		count++
	}
	if count != 1 {
		return errors.New("must provide either a targetARN, topicARN, or phoneNumber for SNS config")
	}
	return nil
}

// ValidateTelegramConfig checks that botToken or botTokenFile is configured and chatID is not zero.
func ValidateTelegramConfig(hasBotToken, hasBotTokenFile bool, chatID int64) error {
	if !hasBotToken && !hasBotTokenFile {
		return errors.New("mandatory field botToken or botTokenfile is empty")
	}
	if chatID == 0 {
		return errors.New("mandatory field 'chatID' is empty")
	}
	return nil
}
