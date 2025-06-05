// Copyright 2018 The prometheus-operator Authors
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

package promtool

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPromtoolGoldenFiles(t *testing.T) {
	goldenFiles, err := filepath.Glob("../../../pkg/prometheus/testdata/*.golden")
	if err != nil {
		t.Fatalf("Failed to list golden files: %v", err)
	}

	for _, file := range goldenFiles {
		t.Run(file, func(t *testing.T) {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			cmd := exec.Command("promtool", "check", "config", "-")
			cmd.Stdin = stringReader(string(content))

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("promtool validation failed for %s: %s", file, string(output))
			}
		})
	}
}

func stringReader(content string) *os.File {
	tmpFile, err := os.CreateTemp("", "promtool-*.yaml")
	if err != nil {
		panic(err)
	}
	tmpFile.WriteString(content)
	tmpFile.Seek(0, 0)
	return tmpFile
}
