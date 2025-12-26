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

//go:build promtool

package prometheus

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPromtoolGoldenFiles(t *testing.T) {
	goldenFiles, err := filepath.Glob("testdata/*.golden")
	if err != nil {
		t.Fatalf("failed to list golden files: %v", err)
	}

	if len(goldenFiles) == 0 {
		t.Fatal("no golden files found")
	}

	for _, file := range goldenFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			// Skip partial configs that are not complete Prometheus configs.
			if !bytes.HasPrefix(bytes.TrimSpace(content), []byte("global:")) {
				t.Skipf("skipping partial config: %s", file)
				return
			}

			tmpFile, err := os.CreateTemp("", "promtool-*.yaml")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.Write(content); err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}
			tmpFile.Close()

			cmd := exec.Command("promtool", "check", "config", tmpFile.Name())
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("promtool validation failed:\n%s", string(output))
			}
		})
	}
}
