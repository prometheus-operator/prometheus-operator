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
	"encoding/json"
	"sort"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

const Version uint = 1

// JsonnetFile is the structure of a `.json` file describing a set of jsonnet
// dependencies. It is used for both, the jsonnetFile and the lockFile.
type JsonnetFile struct {
	// List of dependencies
	Dependencies map[string]deps.Dependency

	// Symlink files to old location
	LegacyImports bool
}

// New returns a new JsonnetFile with the dependencies map initialized
func New() JsonnetFile {
	return JsonnetFile{
		Dependencies:  make(map[string]deps.Dependency),
		LegacyImports: true,
	}
}

// jsonFile is the json representation of a JsonnetFile, which is different for
// compatibility reasons.
type jsonFile struct {
	Version       uint              `json:"version"`
	Dependencies  []deps.Dependency `json:"dependencies"`
	LegacyImports bool              `json:"legacyImports"`
}

// UnmarshalJSON unmarshals a `jsonFile`'s json into a JsonnetFile
func (jf *JsonnetFile) UnmarshalJSON(data []byte) error {
	var s jsonFile
	s.LegacyImports = jf.LegacyImports // adpot default

	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	jf.Dependencies = make(map[string]deps.Dependency)
	for _, d := range s.Dependencies {
		jf.Dependencies[d.Name()] = d
	}

	jf.LegacyImports = s.LegacyImports

	return nil
}

// MarshalJSON serializes a JsonnetFile into json of the format of a `jsonFile`
func (jf JsonnetFile) MarshalJSON() ([]byte, error) {
	var s jsonFile

	s.Version = Version
	s.LegacyImports = jf.LegacyImports

	for _, d := range jf.Dependencies {
		s.Dependencies = append(s.Dependencies, d)
	}

	sort.SliceStable(s.Dependencies, func(i int, j int) bool {
		return s.Dependencies[i].Name() < s.Dependencies[j].Name()
	})

	if s.Dependencies == nil {
		s.Dependencies = make([]deps.Dependency, 0, 0)
	}

	return json.Marshal(s)
}
