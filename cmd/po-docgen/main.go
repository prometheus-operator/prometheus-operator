// Copyright 2016 The prometheus-operator Authors
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
	"os"

	"github.com/prometheus-operator/prometheus-operator/pkg/versionutil"
)

func main() {
	versionutil.RegisterParseFlags()
	if versionutil.ShouldPrintVersion() {
		versionutil.Print(os.Stdout, "po-docgen")
		os.Exit(0)
	}

	switch os.Args[1] {
	case "api":
		printAPIDocs(os.Args[2:])
	case "compatibility":
		printCompatMatrixDocs()
	case "operator":
		printOperatorDocs(os.Args[2:])
	}
}
