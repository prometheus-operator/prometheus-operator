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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
	"github.com/pkg/errors"
)

var (
	JsonnetFile     = "jsonnetfile.json"
	JsonnetLockFile = "jsonnetfile.lock.json"
	VersionMismatch = errors.New("multiple colliding versions specified")
)

func Install(ctx context.Context, isLock bool, dependencySourceIdentifier string, m spec.JsonnetFile, dir string) (lock *spec.JsonnetFile, err error) {
	lock = &spec.JsonnetFile{}
	for _, dep := range m.Dependencies {
		tmp := filepath.Join(dir, ".tmp")
		err = os.MkdirAll(tmp, os.ModePerm)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create general tmp dir")
		}
		tmpDir, err := ioutil.TempDir(tmp, fmt.Sprintf("jsonnetpkg-%s-%s", dep.Name, dep.Version))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create tmp dir")
		}
		defer os.RemoveAll(tmpDir)

		subdir := ""
		var p Interface
		if dep.Source.GitSource != nil {
			p = NewGitPackage(dep.Source.GitSource)
			subdir = dep.Source.GitSource.Subdir
		}

		lockVersion, err := p.Install(ctx, tmpDir, dep.Version)
		if err != nil {
			return nil, errors.Wrap(err, "failed to install package")
		}

		color.Green(">>> Installed %s version %s\n", dep.Name, dep.Version)

		destPath := path.Join(dir, dep.Name)
		if err != nil {
			return nil, errors.Wrap(err, "failed to find destination path for package")
		}

		err = os.MkdirAll(path.Dir(destPath), os.ModePerm)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create parent path")
		}

		err = os.RemoveAll(destPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to clean previous destination path")
		}
		err = os.Rename(path.Join(tmpDir, subdir), destPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to move package")
		}

		lock.Dependencies, err = insertDependency(lock.Dependencies, spec.Dependency{
			Name:      dep.Name,
			Source:    dep.Source,
			Version:   lockVersion,
			DepSource: dependencySourceIdentifier,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to insert dependency to lock dependencies")
		}

		// If dependencies are being installed from a lock file, the transitive
		// dependencies are not questioned, the locked dependencies are just
		// installed.
		if isLock {
			continue
		}

		filepath, isLock, err := ChooseJsonnetFile(destPath)
		if err != nil {
			return nil, err
		}
		depsDeps, err := LoadJsonnetfile(filepath)
		// It is ok for depedencies not to have a JsonnetFile, it just means
		// they do not have transitive dependencies of their own.
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}

		depsInstalledByDependency, err := Install(ctx, isLock, filepath, depsDeps, dir)
		if err != nil {
			return nil, err
		}

		for _, d := range depsInstalledByDependency.Dependencies {
			lock.Dependencies, err = insertDependency(lock.Dependencies, d)
			if err != nil {
				return nil, errors.Wrap(err, "failed to insert dependency to lock dependencies")
			}
		}
	}

	return lock, nil
}

func insertDependency(deps []spec.Dependency, newDep spec.Dependency) ([]spec.Dependency, error) {
	if len(deps) == 0 {
		return []spec.Dependency{newDep}, nil
	}

	res := []spec.Dependency{}
	newDepPreviouslyPresent := false
	for _, d := range deps {
		if d.Name == newDep.Name {
			if d.Version != newDep.Version {
				return nil, fmt.Errorf("multiple colliding versions specified for %s: %s (from %s) and %s (from %s)", d.Name, d.Version, d.DepSource, newDep.Version, newDep.DepSource)
			}
			res = append(res, d)
			newDepPreviouslyPresent = true
		} else {
			res = append(res, d)
		}
	}
	if !newDepPreviouslyPresent {
		res = append(res, newDep)
	}

	return res, nil
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func ChooseJsonnetFile(dir string) (string, bool, error) {
	lockfile := path.Join(dir, JsonnetLockFile)
	jsonnetfile := path.Join(dir, JsonnetFile)
	filename := lockfile
	isLock := true

	lockExists, err := FileExists(filepath.Join(dir, JsonnetLockFile))
	if err != nil {
		return "", false, err
	}

	if !lockExists {
		filename = jsonnetfile
		isLock = false
	}

	return filename, isLock, err
}

func LoadJsonnetfile(filepath string) (spec.JsonnetFile, error) {
	m := spec.JsonnetFile{}

	if _, err := os.Stat(filepath); err != nil {
		return m, err
	}

	f, err := os.Open(filepath)
	if err != nil {
		return m, err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&m)
	if err != nil {
		return m, err
	}

	return m, nil
}
