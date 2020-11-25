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

package versionutil_test

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/prometheus/common/version"
	"github.com/stretchr/testify/assert"

	"github.com/prometheus-operator/prometheus-operator/pkg/versionutil"
)

func TestShouldPrintVersion(t *testing.T) {
	const program = "foo"

	restore := setAllVersionFieldsTo("test-value")
	defer restore()

	tests := map[string]struct {
		flag        string
		expOutput   string
		program     string
		shouldPrint bool
	}{
		"Should print only version": {
			flag:      "--short-version",
			expOutput: "test-value",
		},
		"Should print full version": {
			flag: "--version",
			expOutput: fmt.Sprintf(
				"%s, version test-value (branch: test-value, revision: test-value)\n"+
					"  build user:       test-value\n"+
					"  build date:       test-value\n"+
					"  go version:       test-value\n"+
					"  platform:         linux/amd64", program),
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			restore := snapshotOSArgsAndFlags()
			defer restore()

			// given
			os.Args = []string{program, tc.flag}
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// when
			versionutil.RegisterParseFlags()
			// then
			assert.Equal(t, true, versionutil.ShouldPrintVersion())

			// when
			var buf bytes.Buffer
			versionutil.Print(&buf, program)
			// then
			assert.Equal(t, tc.expOutput, buf.String())
		})
	}
}

func TestShouldNotPrintVersion(t *testing.T) {
	// given
	restore := snapshotOSArgsAndFlags()
	defer restore()

	// no flags set
	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// when
	versionutil.RegisterParseFlags()

	// then
	assert.False(t, versionutil.ShouldPrintVersion())
}

// snapshotOSArgsAndFlags the os.Args and flag.CommandLine and allows you to restore them.
// inspired by: https://golang.org/src/flag/flag_test.go#L318
func snapshotOSArgsAndFlags() func() {
	oldArgs := os.Args
	oldFlags := flag.CommandLine

	return func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlags
	}
}

// setAllVersionFieldsToTestFixture sets all version fields to a given value.
// Simplifies test cases by ensuring that values are predictable.
func setAllVersionFieldsTo(fixture string) func() {
	oldVersion := version.Version
	oldRevision := version.Revision
	oldBranch := version.Branch
	oldBuildUser := version.BuildUser
	oldBuildDate := version.BuildDate
	oldGoVersion := version.GoVersion

	version.Version = fixture
	version.Revision = fixture
	version.Branch = fixture
	version.BuildUser = fixture
	version.BuildDate = fixture
	version.GoVersion = fixture

	return func() {
		version.Version = oldVersion
		version.Revision = oldRevision
		version.Branch = oldBranch
		version.BuildUser = oldBuildUser
		version.BuildDate = oldBuildDate
		version.GoVersion = oldGoVersion
	}
}
