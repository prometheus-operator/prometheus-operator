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
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg"
	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	v1 "github.com/jsonnet-bundler/jsonnet-bundler/spec/v1"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

func updateCommand(dir, jsonnetHome string, urls ...*url.URL) int {
	if dir == "" {
		dir = "."
	}

	jsonnetFile, err := jsonnetfile.Load(filepath.Join(dir, jsonnetfile.File))
	kingpin.FatalIfError(err, "failed to load jsonnetfile")

	kingpin.FatalIfError(
		os.MkdirAll(filepath.Join(dir, jsonnetHome, ".tmp"), os.ModePerm),
		"creating vendor folder")

	// When updating, locks are ignored.
	locks := map[string]deps.Dependency{}
	locked, err := pkg.Ensure(jsonnetFile, jsonnetHome, locks)
	kingpin.FatalIfError(err, "failed to install packages")

	kingpin.FatalIfError(
		writeJSONFile(filepath.Join(dir, jsonnetfile.File), jsonnetFile),
		"updating jsonnetfile.json")
	kingpin.FatalIfError(
		writeJSONFile(filepath.Join(dir, jsonnetfile.LockFile), v1.JsonnetFile{Dependencies: locked}),
		"updating jsonnetfile.lock.json")
	return 0
}
