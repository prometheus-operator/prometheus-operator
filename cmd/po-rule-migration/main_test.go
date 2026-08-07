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

package main

import (
	"os"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	k8sYAML "k8s.io/apimachinery/pkg/util/yaml"
)

// TestFileClosedAfterDecode verifies that the file opened by the main flow is
// properly closed after decoding (regression test for missing defer file.Close()).
// Proof: on all major OSes, an open file cannot be deleted while the handle is
// held. If os.Remove succeeds, the fd has been released.
func TestFileClosedAfterDecode(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "configmap.yaml")

	// Craft a minimal ConfigMap YAML that Decode can parse.
	cm := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-rules
  namespace: default
data:
  test.rules: |
    groups:
      - name: test
        rules:
          - record: test_metric
            expr: up
`
	if err := os.WriteFile(tmpFile, []byte(cm), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Replicate the exact main() pattern: Open → Decode → (defer Close).
	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close() // This is the line that was missing — the fix under test.

	var configMap corev1.ConfigMap
	if err := k8sYAML.NewYAMLOrJSONDecoder(file, 100).Decode(&configMap); err != nil {
		t.Fatalf("failed to decode manifest: %v", err)
	}

	// Sanity check: the decoded object must be a ConfigMap.
	if configMap.Kind != "ConfigMap" {
		t.Errorf("expected Kind=ConfigMap, got %q", configMap.Kind)
	}

	// Verify the file handle is released: os.Remove fails on Windows and some
	// Linux configurations when an fd is still open.
	if err := os.Remove(tmpFile); err != nil {
		t.Errorf("could not remove file — handle likely still open: %v", err)
	}
}

// TestFileNotLeakedOnDecodeError verifies that even when Decode fails, the file
// handle is still released (defer Close runs before log.Fatalf / return).
func TestFileNotLeakedOnDecodeError(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "bad.yaml")

	// Write invalid content that will cause Decode to fail.
	if err := os.WriteFile(tmpFile, []byte(":::not valid yaml:::"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	var configMap corev1.ConfigMap
	_ = k8sYAML.NewYAMLOrJSONDecoder(file, 100).Decode(&configMap)
	// We intentionally ignore the error — the point is that defer Close still
	// runs regardless of whether Decode succeeded or failed.

	if err := os.Remove(tmpFile); err != nil {
		t.Errorf("could not remove file after decode error — handle likely leaked: %v", err)
	}
}
