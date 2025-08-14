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
	reconcileDelay = 0

	// Register delay flag.
	fs.DurationVar(&reconcileDelay, "reconcile-delay", 0,
		"Delay reconciliation by this duration to reduce API calls (e.g., 30s, 2m, 5m). Default: 0 (disabled)")

	return fs
}

func TestDelayFlag(t *testing.T) {
	tests := []struct {
		name                   string
		args                   []string
		expectedReconcileDelay time.Duration
	}{
		{
			name:                   "default values (no flags)",
			args:                   []string{},
			expectedReconcileDelay: 0,
		},
		{
			name: "delay flag set",
			args: []string{
				"--reconcile-delay=2m",
			},
			expectedReconcileDelay: 2 * time.Minute,
		},

		{
			name: "zero explicit value",
			args: []string{
				"--reconcile-delay=0s",
			},
			expectedReconcileDelay: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := setupDelayFlags()
			err := fs.Parse(tt.args)
			assert.NoError(t, err, "Flag parsing should not fail")
			assert.Equal(t, tt.expectedReconcileDelay, reconcileDelay,
				"reconcileDelay should match expected value")
		})
	}
}

func TestDelayFlagsInvalid(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "invalid delay format",
			args: []string{"--reconcile-delay=xyz"},
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
	reconcileDelay = 2 * time.Minute
	cfg := operator.DefaultConfig("100m", "50Mi")

	// Simulate the assignment that happens in run() function.
	cfg.ReconcileDelay = reconcileDelay

	// Verify the config has the correct values
	assert.Equal(t, 2*time.Minute, cfg.ReconcileDelay)
}

func TestConfigDefaultValues(t *testing.T) {
	// Create a default config
	cfg := operator.DefaultConfig("100m", "50Mi")

	// Verify delay field have zero values by default
	assert.Equal(t, time.Duration(0), cfg.ReconcileDelay)
}

func TestEndToEndFlagToConfig(t *testing.T) {
	fs := setupDelayFlags()
	args := []string{
		"--reconcile-delay=5m",
	}

	err := fs.Parse(args)
	assert.NoError(t, err, "Flag parsing should not fail")

	// Create config and assign values (simulating run() function)
	cfg := operator.DefaultConfig("100m", "50Mi")
	cfg.ReconcileDelay = reconcileDelay

	assert.Equal(t, 5*time.Minute, cfg.ReconcileDelay)
}
