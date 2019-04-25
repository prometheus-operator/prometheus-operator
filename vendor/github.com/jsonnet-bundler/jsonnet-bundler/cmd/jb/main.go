/*
Copyright 2018 jsonnet-bundler authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	installActionName = "install"
	updateActionName  = "update"
	initActionName    = "init"
	basePath          = ".jsonnetpkg"
	srcDirName        = "src"
)

var (
	availableSubcommands = []string{
		initActionName,
		installActionName,
	}
	gitSSHRegex                   = regexp.MustCompile("git\\+ssh://git@([^:]+):([^/]+)/([^/]+).git")
	gitSSHWithVersionRegex        = regexp.MustCompile("git\\+ssh://git@([^:]+):([^/]+)/([^/]+).git@(.*)")
	gitSSHWithPathRegex           = regexp.MustCompile("git\\+ssh://git@([^:]+):([^/]+)/([^/]+).git/(.*)")
	gitSSHWithPathAndVersionRegex = regexp.MustCompile("git\\+ssh://git@([^:]+):([^/]+)/([^/]+).git/(.*)@(.*)")

	githubSlugRegex                   = regexp.MustCompile("github.com/([-_a-zA-Z0-9]+)/([-_a-zA-Z0-9]+)")
	githubSlugWithVersionRegex        = regexp.MustCompile("github.com/([-_a-zA-Z0-9]+)/([-_a-zA-Z0-9]+)@(.*)")
	githubSlugWithPathRegex           = regexp.MustCompile("github.com/([-_a-zA-Z0-9]+)/([-_a-zA-Z0-9]+)/(.*)")
	githubSlugWithPathAndVersionRegex = regexp.MustCompile("github.com/([-_a-zA-Z0-9]+)/([-_a-zA-Z0-9]+)/(.*)@(.*)")
)

func main() {
	os.Exit(Main())
}

func Main() int {
	cfg := struct {
		JsonnetHome string
	}{}

	a := kingpin.New(filepath.Base(os.Args[0]), "A jsonnet package manager")
	a.HelpFlag.Short('h')

	a.Flag("jsonnetpkg-home", "The directory used to cache packages in.").
		Default("vendor").StringVar(&cfg.JsonnetHome)

	initCmd := a.Command(initActionName, "Initialize a new empty jsonnetfile")

	installCmd := a.Command(installActionName, "Install all dependencies or install specific ones")
	installCmdURLs := installCmd.Arg("packages", "URLs to package to install").URLList()

	updateCmd := a.Command(updateActionName, "Update all dependencies.")

	command, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing commandline arguments"))
		a.Usage(os.Args[1:])
		return 2
	}

	switch command {
	case initCmd.FullCommand():
		return initCommand()
	case installCmd.FullCommand():
		return installCommand(cfg.JsonnetHome, *installCmdURLs...)
	case updateCmd.FullCommand():
		return updateCommand(cfg.JsonnetHome)
	default:
		installCommand(cfg.JsonnetHome)
	}

	return 0
}

func initCommand() int {
	exists, err := pkg.FileExists(pkg.JsonnetFile)
	if err != nil {
		kingpin.Errorf("Failed to check for jsonnetfile.json: %v", err)
		return 1
	}

	if exists {
		kingpin.Errorf("jsonnetfile.json already exists")
		return 1
	}

	if err := ioutil.WriteFile(pkg.JsonnetFile, []byte("{}\n"), 0644); err != nil {
		kingpin.Errorf("Failed to write new jsonnetfile.json: %v", err)
		return 1
	}

	return 0
}

func parseDepedency(urlString string) *spec.Dependency {
	if spec := parseGitSSHDependency(urlString); spec != nil {
		return spec
	}

	if spec := parseGithubDependency(urlString); spec != nil {
		return spec
	}

	return nil
}

func parseGitSSHDependency(urlString string) *spec.Dependency {
	if !gitSSHRegex.MatchString(urlString) {
		return nil
	}

	subdir := ""
	host := ""
	org := ""
	repo := ""
	version := "master"

	if gitSSHWithPathAndVersionRegex.MatchString(urlString) {
		matches := gitSSHWithPathAndVersionRegex.FindStringSubmatch(urlString)
		host = matches[1]
		org = matches[2]
		repo = matches[3]
		subdir = matches[4]
		version = matches[5]
	} else if gitSSHWithPathRegex.MatchString(urlString) {
		matches := gitSSHWithPathRegex.FindStringSubmatch(urlString)
		host = matches[1]
		org = matches[2]
		repo = matches[3]
		subdir = matches[4]
	} else if gitSSHWithVersionRegex.MatchString(urlString) {
		matches := gitSSHWithVersionRegex.FindStringSubmatch(urlString)
		host = matches[1]
		org = matches[2]
		repo = matches[3]
		version = matches[4]
	} else {
		matches := gitSSHRegex.FindStringSubmatch(urlString)
		host = matches[1]
		org = matches[2]
		repo = matches[3]
	}

	return &spec.Dependency{
		Name: repo,
		Source: spec.Source{
			GitSource: &spec.GitSource{
				Remote: fmt.Sprintf("git@%s:%s/%s", host, org, repo),
				Subdir: subdir,
			},
		},
		Version: version,
	}
}

func parseGithubDependency(urlString string) *spec.Dependency {
	if !githubSlugRegex.MatchString(urlString) {
		return nil
	}

	name := ""
	user := ""
	repo := ""
	subdir := ""
	version := "master"

	if githubSlugWithPathRegex.MatchString(urlString) {
		if githubSlugWithPathAndVersionRegex.MatchString(urlString) {
			matches := githubSlugWithPathAndVersionRegex.FindStringSubmatch(urlString)
			user = matches[1]
			repo = matches[2]
			subdir = matches[3]
			version = matches[4]
			name = path.Base(subdir)
		} else {
			matches := githubSlugWithPathRegex.FindStringSubmatch(urlString)
			user = matches[1]
			repo = matches[2]
			subdir = matches[3]
			name = path.Base(subdir)
		}
	} else {
		if githubSlugWithVersionRegex.MatchString(urlString) {
			matches := githubSlugWithVersionRegex.FindStringSubmatch(urlString)
			user = matches[1]
			repo = matches[2]
			name = repo
			version = matches[3]
		} else {
			matches := githubSlugRegex.FindStringSubmatch(urlString)
			user = matches[1]
			repo = matches[2]
			name = repo
		}
	}

	return &spec.Dependency{
		Name: name,
		Source: spec.Source{
			GitSource: &spec.GitSource{
				Remote: fmt.Sprintf("https://github.com/%s/%s", user, repo),
				Subdir: subdir,
			},
		},
		Version: version,
	}
}

func installCommand(jsonnetHome string, urls ...*url.URL) int {
	workdir := "."

	jsonnetfile, isLock, err := pkg.ChooseJsonnetFile(workdir)
	if err != nil {
		kingpin.Fatalf("failed to choose jsonnetfile: %v", err)
		return 1
	}

	m, err := pkg.LoadJsonnetfile(jsonnetfile)
	if err != nil {
		kingpin.Fatalf("failed to load jsonnetfile: %v", err)
		return 1
	}

	if len(urls) > 0 {
		for _, url := range urls {
			// install package specified in command
			// $ jsonnetpkg install ksonnet git@github.com:ksonnet/ksonnet-lib
			// $ jsonnetpkg install grafonnet git@github.com:grafana/grafonnet-lib grafonnet
			// $ jsonnetpkg install github.com/grafana/grafonnet-lib/grafonnet
			//
			// github.com/(slug)/(dir)

			urlString := url.String()
			newDep := parseDepedency(urlString)
			if newDep == nil {
				kingpin.Errorf("ignoring unrecognized url: %s", url)
				continue
			}

			oldDeps := m.Dependencies
			newDeps := []spec.Dependency{}
			oldDepReplaced := false
			for _, d := range oldDeps {
				if d.Name == newDep.Name {
					newDeps = append(newDeps, *newDep)
					oldDepReplaced = true
				} else {
					newDeps = append(newDeps, d)
				}
			}

			if !oldDepReplaced {
				newDeps = append(newDeps, *newDep)
			}

			m.Dependencies = newDeps
		}
	}

	srcPath := filepath.Join(jsonnetHome)
	err = os.MkdirAll(srcPath, os.ModePerm)
	if err != nil {
		kingpin.Fatalf("failed to create jsonnet home path: %v", err)
		return 3
	}

	lock, err := pkg.Install(context.TODO(), isLock, jsonnetfile, m, jsonnetHome)
	if err != nil {
		kingpin.Fatalf("failed to install: %v", err)
		return 3
	}

	// If installing from lock file there is no need to write any files back.
	if !isLock {
		b, err := json.MarshalIndent(m, "", "    ")
		if err != nil {
			kingpin.Fatalf("failed to encode jsonnet file: %v", err)
			return 3
		}
		b = append(b, []byte("\n")...)

		err = ioutil.WriteFile(pkg.JsonnetFile, b, 0644)
		if err != nil {
			kingpin.Fatalf("failed to write jsonnet file: %v", err)
			return 3
		}

		b, err = json.MarshalIndent(lock, "", "    ")
		if err != nil {
			kingpin.Fatalf("failed to encode jsonnet file: %v", err)
			return 3
		}
		b = append(b, []byte("\n")...)

		err = ioutil.WriteFile(pkg.JsonnetLockFile, b, 0644)
		if err != nil {
			kingpin.Fatalf("failed to write lock file: %v", err)
			return 3
		}
	}

	return 0
}

func updateCommand(jsonnetHome string, urls ...*url.URL) int {
	jsonnetfile := pkg.JsonnetFile

	m, err := pkg.LoadJsonnetfile(jsonnetfile)
	if err != nil {
		kingpin.Fatalf("failed to load jsonnetfile: %v", err)
		return 1
	}

	err = os.MkdirAll(jsonnetHome, os.ModePerm)
	if err != nil {
		kingpin.Fatalf("failed to create jsonnet home path: %v", err)
		return 3
	}

	// When updating, the lockfile is explicitly ignored.
	isLock := false
	lock, err := pkg.Install(context.TODO(), isLock, jsonnetfile, m, jsonnetHome)
	if err != nil {
		kingpin.Fatalf("failed to install: %v", err)
		return 3
	}

	b, err := json.MarshalIndent(lock, "", "    ")
	if err != nil {
		kingpin.Fatalf("failed to encode jsonnet file: %v", err)
		return 3
	}
	b = append(b, []byte("\n")...)

	err = ioutil.WriteFile(pkg.JsonnetLockFile, b, 0644)
	if err != nil {
		kingpin.Fatalf("failed to write lock file: %v", err)
		return 3
	}

	return 0
}
