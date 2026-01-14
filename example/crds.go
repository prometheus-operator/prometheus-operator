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

// Package crd provides access to the embedded CRD manifests.
package crd

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"slices"
	"strings"
)

//go:embed prometheus-operator-crd/*.yaml
var embeddedCRDs embed.FS

//go:embed prometheus-operator-crd-full/*.yaml
var embeddedFullCRDs embed.FS

const (
	crdsDir     = "prometheus-operator-crd"
	fullCRDsDir = "prometheus-operator-crd-full"
)

// PrintAll prints all standard CRDs to the given writer.
func PrintAll(w io.Writer) error {
	return printCRDs(w, embeddedCRDs, crdsDir)
}

// PrintAllFull prints all full CRDs to the given writer.
func PrintAllFull(w io.Writer) error {
	return printCRDs(w, embeddedFullCRDs, fullCRDsDir)
}

func printCRDs(w io.Writer, fsys embed.FS, dir string) error {
	files, err := fsys.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read embedded CRDs directory: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no CRD files found in embedded directory")
	}

	slices.SortFunc(files, func(a, b fs.DirEntry) int {
		return strings.Compare(a.Name(), b.Name())
	})

	for i, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		if i > 0 {
			fmt.Fprintln(w, "---")
		}

		content, err := fsys.ReadFile(dir + "/" + file.Name())
		if err != nil {
			return fmt.Errorf("failed to read CRD %s: %w", file.Name(), err)
		}

		if _, err := w.Write(content); err != nil {
			return fmt.Errorf("failed to write CRD %s: %w", file.Name(), err)
		}
	}

	return nil
}
