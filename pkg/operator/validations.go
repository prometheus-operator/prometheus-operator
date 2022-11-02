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
