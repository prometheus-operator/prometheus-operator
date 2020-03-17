// Copyright 2018 jsonnet-bundler authors
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

// Package rewrite provides a tool that automatically rewrites legacy imports to
// absolute ones
package rewrite

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

var expr = regexp.MustCompile(`(?mU)(import ["'])(.*)(\/.*["'])`)

// Rewrite changes all imports in `dir` from legacy to absolute style
// All files in `vendorDir` are ignored
func Rewrite(dir, vendorDir string, packages map[string]deps.Dependency) error {
	imports := make(map[string]string)
	for _, p := range packages {
		if p.LegacyName() == p.Name() {
			continue
		}

		imports[p.LegacyName()] = p.Name()
	}

	vendorFi, err := os.Stat(filepath.Join(dir, vendorDir))
	if err != nil {
		return err
	}

	// list all Jsonnet files
	files := []string{}
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if os.SameFile(vendorFi, info) {
			return filepath.SkipDir
		}

		if ext := filepath.Ext(path); ext == ".jsonnet" || ext == ".libsonnet" {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return err
	}

	// change the imports
	for _, s := range files {
		if err := replaceFile(s, imports); err != nil {
			return err
		}
	}

	return nil
}

func wrap(s, q string) string {
	return fmt.Sprintf(`import %s%s`, q, s)
}

func replaceFile(name string, imports map[string]string) error {
	raw, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}

	out := replace(string(raw), imports)
	return ioutil.WriteFile(name, out, 0644)
}

func replace(data string, imports map[string]string) []byte {
	contents := strings.Split(string(data), "\n")

	// try to fix imports line by line
	buf := make([]string, 0, len(contents))
	for _, line := range contents {
		match := expr.FindStringSubmatch(line)
		// no import in this line: push unmodified
		if len(match) == 0 {
			buf = append(buf, line)
			continue
		}

		// the legacyName
		matchedName := match[2]

		replaced := false
		for legacy, absolute := range imports {
			// not this import
			if matchedName != legacy {
				continue
			}

			// fix the import
			replaced = true
			buf = append(buf, expr.ReplaceAllString(line, "${1}"+absolute+"${3}"))
		}

		// no matching known import found? push unmodified
		if !replaced {
			buf = append(buf, line)
		}
	}

	return []byte(strings.Join(buf, "\n"))
}
