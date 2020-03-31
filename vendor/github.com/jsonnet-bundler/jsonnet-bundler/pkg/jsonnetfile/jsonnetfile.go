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

package jsonnetfile

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	v0 "github.com/jsonnet-bundler/jsonnet-bundler/spec/v0"
	v1 "github.com/jsonnet-bundler/jsonnet-bundler/spec/v1"
)

const (
	File     = "jsonnetfile.json"
	LockFile = "jsonnetfile.lock.json"
)

var (
	ErrUpdateJB = errors.New("jsonnetfile version unknown, update jb")
)

// Load reads a jsonnetfile.(lock).json from disk
func Load(filepath string) (v1.JsonnetFile, error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return v1.New(), err
	}

	return Unmarshal(bytes)
}

// Unmarshal creates a spec.JsonnetFile from bytes. Empty bytes
// will create an empty spec.
func Unmarshal(bytes []byte) (v1.JsonnetFile, error) {
	m := v1.New()

	if len(bytes) == 0 {
		return m, nil
	}

	versions := struct {
		Version uint `json:"version"`
	}{}

	err := json.Unmarshal(bytes, &versions)
	if err != nil {
		return m, err
	}

	switch versions.Version {
	case v0.Version:
		var mv0 v0.JsonnetFile
		if err := json.Unmarshal(bytes, &mv0); err != nil {
			return m, errors.Wrap(err, "failed to unmarshal jsonnetfile")
		}
		return v1.FromV0(mv0)
	case v1.Version:
		if err := json.Unmarshal(bytes, &m); err != nil {
			return m, errors.Wrap(err, "failed to unmarshal v1 file")
		}
		return m, nil
	default:
		return m, ErrUpdateJB
	}
}

// Exists returns whether the file at the given path exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
