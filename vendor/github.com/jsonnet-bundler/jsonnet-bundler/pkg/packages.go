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

package pkg

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pkg/errors"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
)

var (
	VersionMismatch = errors.New("multiple colliding versions specified")
)

// Ensure receives all direct packages as, the directory to vendor in and all known locks.
// It then makes sure all direct and nested dependencies are present in vendor at the correct version:
//
// If the package is locked and the files in vendor match the sha256 checksum,
// nothing needs to be done. Otherwise, the package is retrieved from the
// upstream source and added into vendor. If previously locked, the sums are
// checked as well.
// In case a (nested) package is already present in the lock,
// the one from the lock takes precedence. This allows the user to set the
// desired version in case by `jb install`ing it.
//
// Finally, all unknown files and directories are removed from vendor/
func Ensure(direct spec.JsonnetFile, vendorDir string, locks map[string]spec.Dependency) (map[string]spec.Dependency, error) {
	// ensure all required files are in vendor
	deps, err := ensure(direct.Dependencies, vendorDir, locks)
	if err != nil {
		return nil, err
	}

	// cleanup unknown dirs from vendor/
	f, err := os.Open(vendorDir)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		if _, ok := deps[name]; !ok {
			dir := filepath.Join(vendorDir, name)
			if err := os.RemoveAll(dir); err != nil {
				return nil, err
			}
			if name != ".tmp" {
				color.Magenta("CLEAN %s", dir)
			}
		}
	}

	// return the final lockfile contents
	return deps, nil
}

func ensure(direct map[string]spec.Dependency, vendorDir string, locks map[string]spec.Dependency) (map[string]spec.Dependency, error) {
	deps := make(map[string]spec.Dependency)

	for _, d := range direct {
		l, present := locks[d.Name]

		// already locked and the integrity is intact
		if present {
			d.Version = locks[d.Name].Version

			if check(l, vendorDir) {
				deps[d.Name] = l
				continue
			}
		}
		expectedSum := locks[d.Name].Sum

		// either not present or not intact: download again
		dir := filepath.Join(vendorDir, d.Name)
		os.RemoveAll(dir)

		locked, err := download(d, vendorDir)
		if err != nil {
			return nil, errors.Wrap(err, "downloading")
		}
		if expectedSum != "" && locked.Sum != expectedSum {
			return nil, fmt.Errorf("checksum mismatch for %s. Expected %s but got %s", d.Name, expectedSum, locked.Sum)
		}
		deps[d.Name] = *locked
		// we settled on a new version, add it to the locks for recursion
		locks[d.Name] = *locked
	}

	for _, d := range deps {
		f, err := jsonnetfile.Load(filepath.Join(vendorDir, d.Name, jsonnetfile.File))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		nested, err := ensure(f.Dependencies, vendorDir, locks)
		if err != nil {
			return nil, err
		}

		for _, d := range nested {
			if _, ok := deps[d.Name]; !ok {
				deps[d.Name] = d
			}
		}
	}

	return deps, nil
}

// download retrieves a package from a remote upstream. The checksum of the
// files is generated afterwards.
func download(d spec.Dependency, vendorDir string) (*spec.Dependency, error) {
	var p Interface
	switch {
	case d.Source.GitSource != nil:
		p = NewGitPackage(d.Source.GitSource)
	case d.Source.LocalSource != nil:
		p = NewLocalPackage(d.Source.LocalSource)
	}

	if p == nil {
		return nil, errors.New("either git or local source is required")
	}

	version, err := p.Install(context.TODO(), d.Name, vendorDir, d.Version)
	if err != nil {
		return nil, err
	}

	var sum string
	if d.Source.LocalSource == nil {
		sum = hashDir(filepath.Join(vendorDir, d.Name))
	}

	return &spec.Dependency{
		Name:    d.Name,
		Source:  d.Source,
		Version: version,
		Sum:     sum,
	}, nil
}

// check returns whether the files present at the vendor/ folder match the
// sha256 sum of the package. local-directory dependencies are not checked as
// their purpose is to change during development where integrity checking would
// be a hindrance.
func check(d spec.Dependency, vendorDir string) bool {
	// assume a local dependency is intact as long as it exists
	if d.Source.LocalSource != nil {
		x, err := jsonnetfile.Exists(filepath.Join(vendorDir, d.Name))
		if err != nil {
			return false
		}
		return x
	}

	if d.Sum == "" {
		// no sum available, need to download
		return false
	}

	dir := filepath.Join(vendorDir, d.Name)
	sum := hashDir(dir)
	return d.Sum == sum
}

// hashDir computes the checksum of a directory by concatenating all files and
// hashing this data using sha256. This can be memory heavy with lots of data,
// but jsonnet files should be fairly small
func hashDir(dir string) string {
	hasher := sha256.New()

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(hasher, f); err != nil {
			return err
		}

		return nil
	})

	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}
