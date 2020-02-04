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

package main

import (
	"io/ioutil"
	"path/filepath"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
)

func initCommand(dir string) int {
	exists, err := jsonnetfile.Exists(jsonnetfile.File)
	if err != nil {
		kingpin.Errorf("Failed to check for jsonnetfile.json: %v", err)
		return 1
	}

	if exists {
		kingpin.Errorf("jsonnetfile.json already exists")
		return 1
	}

	filename := filepath.Join(dir, jsonnetfile.File)

	if err := ioutil.WriteFile(filename, []byte("{}\n"), 0644); err != nil {
		kingpin.Errorf("Failed to write new jsonnetfile.json: %v", err)
		return 1
	}

	return 0
}
