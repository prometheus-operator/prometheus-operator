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

package e2e

import (
	"flag"
	"log"
	"os"
	"testing"

	operatorFramework "github.com/coreos/prometheus-operator/test/e2e/framework"
)

var framework *operatorFramework.Framework

// Basic set of e2e tests for the operator:
// - config reload (with and without external url)

func TestMain(m *testing.M) {
	kubeconfig := flag.String("kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	opImage := flag.String("operator-image", "", "operator image, e.g. quay.io/coreos/prometheus-operator")
	ns := flag.String("namespace", "prometheus-operator-e2e-tests", "e2e test namespace")
	ip := flag.String("cluster-ip", "", "ip of the kubernetes cluster to use for external requests")
	flag.Parse()

	var (
		err  error
		code int = 0
	)

	if framework, err = operatorFramework.New(*ns, *kubeconfig, *opImage, *ip); err != nil {
		log.Printf("failed to setup framework: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if err := framework.Teardown(); err != nil {
			log.Printf("failed to teardown framework: %v\n", err)
			os.Exit(1)
		}
		os.Exit(code)
	}()

	code = m.Run()
}
