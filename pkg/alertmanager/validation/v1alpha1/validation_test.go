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

package v1alpha1

import (
	"strings"
	"testing"

	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

func TestValidateMuteTimeIntervalsDuplicateName(t *testing.T) {
	period := func(day monitoringv1alpha1.WeekdayRange) []monitoringv1alpha1.TimeInterval {
		return []monitoringv1alpha1.TimeInterval{{Weekdays: []monitoringv1alpha1.WeekdayRange{day}}}
	}

	_, err := validateMuteTimeIntervals([]monitoringv1alpha1.MuteTimeInterval{
		{Name: "night", TimeIntervals: period("Saturday")},
		{Name: "night", TimeIntervals: period("Sunday")},
	})
	if err == nil || !strings.Contains(err.Error(), "not unique") {
		t.Fatalf("expected a duplicate name error, got %v", err)
	}

	if _, err := validateMuteTimeIntervals([]monitoringv1alpha1.MuteTimeInterval{
		{Name: "night", TimeIntervals: period("Saturday")},
		{Name: "day", TimeIntervals: period("Sunday")},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
