// Copyright 2021 The prometheus-operator Authors
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
	"github.com/alecthomas/units"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
)

func ValidateSizeField(sizeField string) error {
	// To validate if given value is parsable for the acceptable size values
	if _, err := units.ParseBase2Bytes(sizeField); err != nil {
		return err
	}
	return nil
}

func ValidateDurationField(durationField string) error {
	// To validate if given value is parsable for the acceptable duration values
	if _, err := model.ParseDuration(durationField); err != nil {
		return err
	}
	return nil
}

// CompareScrapeTimeoutToScrapeInterval validates value of scrapeTimeout based on scrapeInterval
func CompareScrapeTimeoutToScrapeInterval(scrapeTimeout, scrapeInterval string) error {
	var err error
	var si, st model.Duration

	// since default scrapeInterval set by operator is 30s
	if scrapeInterval == "" {
		scrapeInterval = "30s"
	}

	if si, err = model.ParseDuration(scrapeInterval); err != nil {
		return errors.Wrapf(err, "invalid scrapeInterval %q", scrapeInterval)
	}
	if st, err = model.ParseDuration(scrapeTimeout); err != nil {
		return errors.Wrapf(err, "invalid scrapeTimeout: %q", scrapeTimeout)
	}

	if st > si {
		return errors.Errorf("scrapeTimeout %q greater than scrapeInterval %q", scrapeTimeout, scrapeInterval)
	}

	return nil
}
