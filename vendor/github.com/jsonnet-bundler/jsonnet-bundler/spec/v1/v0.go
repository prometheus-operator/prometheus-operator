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
package spec

import (
	"path/filepath"

	v0 "github.com/jsonnet-bundler/jsonnet-bundler/spec/v0"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

func FromV0(mv0 v0.JsonnetFile) (JsonnetFile, error) {
	m := New()
	m.LegacyImports = true

	for name, old := range mv0.Dependencies {
		var d deps.Dependency

		switch {
		case old.Source.GitSource != nil:
			d = *deps.Parse("", old.Source.GitSource.Remote)

			if old.Source.GitSource.Subdir != "" {
				subdir := filepath.Clean("/" + old.Source.GitSource.Subdir)
				d.Source.GitSource.Subdir = subdir
			}

		case old.Source.LocalSource != nil:
			d = *deps.Parse("", old.Source.LocalSource.Directory)
		}

		d.Sum = old.Sum
		d.Version = old.Version
		d.LegacyNameCompat = name

		m.Dependencies[d.Name()] = d
	}

	return m, nil
}
