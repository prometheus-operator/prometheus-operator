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
	"flag"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"k8s.io/api/core/v1"

	"os"
)

var (
	cfg Config
)

func init() {
	cfg.CrdKinds = monitoringv1.DefaultCrdKinds
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagset.Var(&cfg.Labels, "labels", "Labels to be add to all resources created by the operator")
	flagset.BoolVar(&cfg.EnableValidation, "with-validation", false, "Include the validation spec")
	flagset.StringVar(&cfg.CrdGroup, "crd-apigroup", monitoringv1.Group, "prometheus CRD  API group name")
	flagset.Var(&cfg.CrdKinds, "crd-kinds", "customize CRD kind names")
	flagset.StringVar(&cfg.Namespace, "namespace", v1.NamespaceAll, "Namespace to scope the interaction of the Prometheus Operator and the apiserver.")
	flagset.Parse(os.Args[1:])
}

func main() {
	PrintCrd(cfg)
	os.Exit(0)
}
