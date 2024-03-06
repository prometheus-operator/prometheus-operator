// Copyright 2023 The prometheus-operator Authors
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

package prometheus

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestStartupProbeTimeoutSeconds(t *testing.T) {
	tests := []struct {
		maximumStartupDurationSeconds   *int32
		expectedStartupPeriodSeconds    int32
		expectedStartupFailureThreshold int32
		expectedMaxStartupDuration      int32
	}{
		{
			maximumStartupDurationSeconds:   nil,
			expectedStartupPeriodSeconds:    15,
			expectedStartupFailureThreshold: 60,
			expectedMaxStartupDuration:      900,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(0)),
			expectedStartupPeriodSeconds:    15,
			expectedStartupFailureThreshold: 60,
			expectedMaxStartupDuration:      900,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(1)),
			expectedStartupPeriodSeconds:    15,
			expectedStartupFailureThreshold: 60,
			expectedMaxStartupDuration:      900,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(60)),
			expectedStartupPeriodSeconds:    60,
			expectedStartupFailureThreshold: 1,
			expectedMaxStartupDuration:      60,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(600)),
			expectedStartupPeriodSeconds:    60,
			expectedStartupFailureThreshold: 10,
			expectedMaxStartupDuration:      600,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(900)),
			expectedStartupPeriodSeconds:    60,
			expectedStartupFailureThreshold: 15,
			expectedMaxStartupDuration:      900,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(1200)),
			expectedStartupPeriodSeconds:    60,
			expectedStartupFailureThreshold: 20,
			expectedMaxStartupDuration:      1200,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(129)),
			expectedStartupPeriodSeconds:    43,
			expectedStartupFailureThreshold: 3,
			expectedMaxStartupDuration:      129,
		},
		{
			maximumStartupDurationSeconds:   ptr.To(int32(322)),
			expectedStartupPeriodSeconds:    54,
			expectedStartupFailureThreshold: 6,
			expectedMaxStartupDuration:      324,
		},
	}

	for _, test := range tests {
		startupPeriodSeconds, startupFailureThreshold := GetStatupProbePeriodSecondsAndFailureThreshold(monitoringv1.CommonPrometheusFields{
			MaximumStartupDurationSeconds: test.maximumStartupDurationSeconds,
		})

		require.Equal(t, test.expectedStartupPeriodSeconds, startupPeriodSeconds)
		require.Equal(t, test.expectedStartupFailureThreshold, startupFailureThreshold)
		require.Equal(t, test.expectedMaxStartupDuration, startupPeriodSeconds*startupFailureThreshold)
	}
}
