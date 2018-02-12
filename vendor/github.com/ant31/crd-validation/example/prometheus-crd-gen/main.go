// Copyright 2018
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
	v1 "github.com/ant31/crd-validation/example/prometheus-crd-gen/v1"
	crdutils "github.com/ant31/crd-validation/pkg"
	"os"
)

var (
	cfg crdutils.Config
)

func init() {
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagset := crdutils.InitFlags(&cfg, flagset)
	flagset.Parse(os.Args[1:])
}

func main() {
	cfg.GetOpenAPIDefinitions = v1.GetOpenAPIDefinitions
	crd := crdutils.NewCustomResourceDefinition(cfg)
	crdutils.MarshallCrd(crd, cfg.OutputFormat)
	os.Exit(0)
}
