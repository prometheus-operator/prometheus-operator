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
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

const (
	installActionName = "install"
	updateActionName  = "update"
	initActionName    = "init"
	rewriteActionName = "rewrite"
)

func main() {
	os.Exit(Main())
}

func Main() int {
	cfg := struct {
		JsonnetHome string
	}{}

	color.Output = color.Error

	a := newApp()
	a.HelpFlag.Short('h')

	a.Flag("jsonnetpkg-home", "The directory used to cache packages in.").
		Default("vendor").StringVar(&cfg.JsonnetHome)

	initCmd := a.Command(initActionName, "Initialize a new empty jsonnetfile")

	installCmd := a.Command(installActionName, "Install all dependencies or install specific ones")
	installCmdURIs := installCmd.Arg("uris", "URIs to packages to install, URLs or file paths").Strings()

	updateCmd := a.Command(updateActionName, "Update all dependencies.")

	rewriteCmd := a.Command(rewriteActionName, "Automatically rewrite legacy imports to absolute ones")

	command, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing commandline arguments"))
		a.Usage(os.Args[1:])
		return 2
	}

	workdir, err := os.Getwd()
	if err != nil {
		return 1
	}

	cfg.JsonnetHome = filepath.Clean(cfg.JsonnetHome)

	switch command {
	case initCmd.FullCommand():
		return initCommand(workdir)
	case installCmd.FullCommand():
		return installCommand(workdir, cfg.JsonnetHome, *installCmdURIs)
	case updateCmd.FullCommand():
		return updateCommand(workdir, cfg.JsonnetHome)
	case rewriteCmd.FullCommand():
		return rewriteCommand(workdir, cfg.JsonnetHome)
	default:
		installCommand(workdir, cfg.JsonnetHome, []string{})
	}

	return 0
}
