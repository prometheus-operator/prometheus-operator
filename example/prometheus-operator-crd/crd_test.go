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

package crd

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintAll(t *testing.T) {
	var buf bytes.Buffer
	err := PrintAll(&buf)
	if err != nil {
		t.Fatalf("PrintAll() error = %v", err)
	}

	output := buf.String()

	// Check that output is not empty
	if len(output) == 0 {
		t.Error("PrintAll() returned empty output")
	}

	// Check for expected CRD content markers
	expectedMarkers := []string{
		"apiVersion: apiextensions.k8s.io/v1",
		"kind: CustomResourceDefinition",
		"monitoring.coreos.com",
	}

	for _, marker := range expectedMarkers {
		if !strings.Contains(output, marker) {
			t.Errorf("PrintAll() output missing expected marker: %q", marker)
		}
	}

	// Check that we have multiple CRDs (separated by ---)
	crdCount := strings.Count(output, "kind: CustomResourceDefinition")
	if crdCount < 10 {
		t.Errorf("PrintAll() expected at least 10 CRDs, got %d", crdCount)
	}
}

func TestListNames(t *testing.T) {
	names, err := ListNames()
	if err != nil {
		t.Fatalf("ListNames() error = %v", err)
	}

	// Check that we have the expected number of CRDs
	if len(names) < 10 {
		t.Errorf("ListNames() expected at least 10 CRDs, got %d", len(names))
	}

	// Check for specific expected CRD names
	expectedCRDs := []string{
		"monitoring.coreos.com_prometheuses.yaml",
		"monitoring.coreos.com_alertmanagers.yaml",
		"monitoring.coreos.com_servicemonitors.yaml",
		"monitoring.coreos.com_podmonitors.yaml",
		"monitoring.coreos.com_prometheusrules.yaml",
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}

	for _, expected := range expectedCRDs {
		if !nameSet[expected] {
			t.Errorf("ListNames() missing expected CRD: %q", expected)
		}
	}
}

func TestPrintByName(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "monitoring.coreos.com_prometheuses.yaml",
			want:    "prometheuses.monitoring.coreos.com",
			wantErr: false,
		},
		{
			name:    "monitoring.coreos.com_prometheuses", // without .yaml
			want:    "prometheuses.monitoring.coreos.com",
			wantErr: false,
		},
		{
			name:    "nonexistent.yaml",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintByName(&buf, tt.name)

			if (err != nil) != tt.wantErr {
				t.Errorf("PrintByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !strings.Contains(buf.String(), tt.want) {
				t.Errorf("PrintByName() output missing expected content: %q", tt.want)
			}
		})
	}
}
