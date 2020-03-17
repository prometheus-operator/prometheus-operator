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
	"path/filepath"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"github.com/jsonnet-bundler/jsonnet-bundler/tool/rewrite"
)

func rewriteCommand(dir, vendorDir string) int {
	locks, err := jsonnetfile.Load(filepath.Join(dir, jsonnetfile.LockFile))
	if err != nil {
		kingpin.Fatalf("Failed to load lockFile: %s.\nThe locks are required to compute the new import names. Make sure to run `jb install` first.", err)
	}

	if err := rewrite.Rewrite(dir, vendorDir, locks.Dependencies); err != nil {
		kingpin.FatalIfError(err, "")
	}

	return 0
}
