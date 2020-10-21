// Copyright The prometheus-operator Authors
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

// Holds the CLI related version functions that unifies handling version printing in all CLIs binaries.
package versionutil

import (
	"flag"
	"fmt"
	"io"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/prometheus/common/version"
)

var (
	printVer   bool
	printShort bool
)

// RegisterParseFlags registers and parses version related flags.
func RegisterParseFlags() {
	RegisterFlags()
	flag.Parse()
}

// RegisterFlags registers version related flags to core.
func RegisterFlags() {
	flag.BoolVar(&printVer, "version", false, "Prints current version.")
	flag.BoolVar(&printShort, "short-version", false, "Print just the version number.")
}

// RegisterIntoKingpinFlags registers version related flags in kingpin framework.
func RegisterIntoKingpinFlags(app *kingpin.Application) {
	app.Flag("version", "Prints current version.").Default("false").BoolVar(&printVer)
	app.Flag("short-version", "Print just the version number.").Default("false").BoolVar(&printShort)
}

// ShouldPrintVersion returns true if version should be printed.
// Use Print function to print version information.
func ShouldPrintVersion() bool {
	return printVer || printShort
}

// Print version information to a given out writer.
func Print(out io.Writer, program string) {
	if printShort {
		fmt.Fprint(out, version.Version)
		return
	}
	if printVer {
		fmt.Fprint(out, version.Print(program))
	}
}
