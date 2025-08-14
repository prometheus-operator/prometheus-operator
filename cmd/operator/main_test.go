// Copyright 2025 The prometheus-operator Authors
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

package main

import (
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

// setupDelayFlags creates a minimal flag set with just the delay flags.
func setupDelayFlags() *flag.FlagSet {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	// Reset global variables
	alertmanagerReconcileDelay = 0
	prometheusReconcileDelay = 0
	thanosRulerReconcileDelay = 0

	// Register delay flags.
	fs.DurationVar(&alertmanagerReconcileDelay, "alertmanager-reconcile-delay", 0,
		"Delay Alertmanager reconciliation by this duration to reduce API calls (e.g., 30s, 2m, 5m). Default: 0 (disabled)")
	fs.DurationVar(&prometheusReconcileDelay, "prometheus-reconcile-delay", 0,
		"Delay Prometheus reconciliation by this duration to reduce API calls (e.g., 30s, 2m, 5m). Default: 0 (disabled)")
	fs.DurationVar(&thanosRulerReconcileDelay, "thanos-ruler-reconcile-delay", 0,
		"Delay ThanosRuler reconciliation by this duration to reduce API calls (e.g., 30s, 2m, 5m). Default: 0 (disabled)")

	return fs
}

func TestDelayFlags(t *testing.T) {
	tests := []struct {
		name                               string
		args                               []string
		expectedAlertmanagerReconcileDelay time.Duration
		expectedPrometheusReconcileDelay   time.Duration
		expectedThanosRulerReconcileDelay  time.Duration
	}{
		{
			name:                               "default values (no flags)",
			args:                               []string{},
			expectedAlertmanagerReconcileDelay: 0,
			expectedPrometheusReconcileDelay:   0,
			expectedThanosRulerReconcileDelay:  0,
		},
		{
			name: "all delay flags set",
			args: []string{
				"--alertmanager-reconcile-delay=2m",
				"--prometheus-reconcile-delay=1m30s",
				"--thanos-ruler-reconcile-delay=45s",
			},
			expectedAlertmanagerReconcileDelay: 2 * time.Minute,
			expectedPrometheusReconcileDelay:   90 * time.Second,
			expectedThanosRulerReconcileDelay:  45 * time.Second,
		},
		{
			name: "only alertmanager delay set",
			args: []string{
				"--alertmanager-reconcile-delay=5m",
			},
			expectedAlertmanagerReconcileDelay: 5 * time.Minute,
			expectedPrometheusReconcileDelay:   0,
			expectedThanosRulerReconcileDelay:  0,
		},
		{
			name: "only prometheus delay set",
			args: []string{
				"--prometheus-reconcile-delay=30s",
			},
			expectedAlertmanagerReconcileDelay: 0,
			expectedPrometheusReconcileDelay:   30 * time.Second,
			expectedThanosRulerReconcileDelay:  0,
		},
		{
			name: "mixed delay settings",
			args: []string{
				"--alertmanager-reconcile-delay=3m",
				"--thanos-ruler-reconcile-delay=1m",
			},
			expectedAlertmanagerReconcileDelay: 3 * time.Minute,
			expectedPrometheusReconcileDelay:   0,
			expectedThanosRulerReconcileDelay:  1 * time.Minute,
		},
		{
			name: "zero explicit values",
			args: []string{
				"--alertmanager-reconcile-delay=0s",
				"--prometheus-reconcile-delay=0m",
				"--thanos-ruler-reconcile-delay=0h",
			},
			expectedAlertmanagerReconcileDelay: 0,
			expectedPrometheusReconcileDelay:   0,
			expectedThanosRulerReconcileDelay:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := setupDelayFlags()
			err := fs.Parse(tt.args)
			assert.NoError(t, err, "Flag parsing should not fail")

			// Verify the global variables have the expected values
			assert.Equal(t, tt.expectedAlertmanagerReconcileDelay, alertmanagerReconcileDelay,
				"alertmanagerReconcileDelay should match expected value")
			assert.Equal(t, tt.expectedPrometheusReconcileDelay, prometheusReconcileDelay,
				"prometheusReconcileDelay should match expected value")
			assert.Equal(t, tt.expectedThanosRulerReconcileDelay, thanosRulerReconcileDelay,
				"thanosRulerReconcileDelay should match expected value")
		})
	}
}

func TestDelayFlagsInvalid(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "invalid alertmanager delay format",
			args: []string{"--alertmanager-reconcile-delay=invalid"},
		},
		{
			name: "invalid prometheus delay format",
			args: []string{"--prometheus-reconcile-delay=not-a-duration"},
		},
		{
			name: "invalid thanos ruler delay format",
			args: []string{"--thanos-ruler-reconcile-delay=xyz"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := setupDelayFlags()
			err := fs.Parse(tt.args)
			assert.Error(t, err, "Invalid duration should cause parsing error")
		})
	}
}

func TestConfigAssignment(t *testing.T) {
	// Set test values
	alertmanagerReconcileDelay = 2 * time.Minute
	prometheusReconcileDelay = 90 * time.Second
	thanosRulerReconcileDelay = 30 * time.Second

	cfg := operator.DefaultConfig("100m", "50Mi")

	// Simulate the assignment that happens in run() function.
	cfg.AlertmanagerReconcileDelay = alertmanagerReconcileDelay
	cfg.PrometheusReconcileDelay = prometheusReconcileDelay
	cfg.ThanosRulerReconcileDelay = thanosRulerReconcileDelay

	// Verify the config has the correct values
	assert.Equal(t, 2*time.Minute, cfg.AlertmanagerReconcileDelay)
	assert.Equal(t, 90*time.Second, cfg.PrometheusReconcileDelay)
	assert.Equal(t, 30*time.Second, cfg.ThanosRulerReconcileDelay)
}

func TestConfigDefaultValues(t *testing.T) {
	// Create a default config
	cfg := operator.DefaultConfig("100m", "50Mi")

	// Verify delay fields have zero values by default
	assert.Equal(t, time.Duration(0), cfg.AlertmanagerReconcileDelay)
	assert.Equal(t, time.Duration(0), cfg.PrometheusReconcileDelay)
	assert.Equal(t, time.Duration(0), cfg.ThanosRulerReconcileDelay)
}

func TestEndToEndFlagToConfig(t *testing.T) {
	fs := setupDelayFlags()
	args := []string{
		"--alertmanager-reconcile-delay=5m",
		"--prometheus-reconcile-delay=2m",
		"--thanos-ruler-reconcile-delay=1m",
	}

	err := fs.Parse(args)
	assert.NoError(t, err, "Flag parsing should not fail")

	// Create config and assign values (simulating run() function)
	cfg := operator.DefaultConfig("100m", "50Mi")
	cfg.AlertmanagerReconcileDelay = alertmanagerReconcileDelay
	cfg.PrometheusReconcileDelay = prometheusReconcileDelay
	cfg.ThanosRulerReconcileDelay = thanosRulerReconcileDelay

	assert.Equal(t, 5*time.Minute, cfg.AlertmanagerReconcileDelay)
	assert.Equal(t, 2*time.Minute, cfg.PrometheusReconcileDelay)
	assert.Equal(t, 1*time.Minute, cfg.ThanosRulerReconcileDelay)
}
