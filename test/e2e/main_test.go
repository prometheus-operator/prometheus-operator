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

	operatorFramework "github.com/coreos/prometheus-operator/test/framework"
)

var framework *operatorFramework.Framework

// Basic set of e2e tests for the operator:
// - config reload (with and without external url)

func TestMain(m *testing.M) {
	kubeconfig := flag.String("kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	opImage := flag.String("operator-image", "", "operator image, e.g. quay.io/coreos/prometheus-operator")
	ns := flag.String("namespace", "prometheus-operator-e2e-tests", "e2e test namespace")
	flag.Parse()

	var (
		err      error
		exitCode int
	)

	if framework, err = operatorFramework.New(*ns, *kubeconfig, *opImage); err != nil {
		log.Printf("failed to setup framework: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if exitCode != 0 {
			if err := framework.PrintEvents(); err != nil {
				log.Printf("failed to print events: %v", err)
			}
			if err := framework.PrintPodLogs(framework.Namespace.Name, framework.OperatorPod.Name); err != nil {
				log.Printf("failed to print Prometheus Operator logs: %v", err)
			}
		}

		if err := framework.Teardown(); err != nil {
			log.Printf("failed to teardown framework: %v\n", err)
			exitCode = 1
		}

		os.Exit(exitCode)
	}()

	exitCode = m.Run()

	// Check if Prometheus Operator ever restarted.
	restarts, err := framework.GetRestartCount(framework.Namespace.Name, framework.OperatorPod.Name)
	if err != nil {
		log.Printf("failed to retrieve restart count of Prometheus Operator pod: %v", err)
		exitCode = 1
	}
	if len(restarts) != 1 {
		log.Printf("expected to have 1 container but got %d", len(restarts))
		exitCode = 1
	}
	for _, restart := range restarts {
		if restart != 0 {
			log.Printf(
				"expected Prometheus Operator to never restart during entire test execution but got %d restarts",
				restart,
			)
			exitCode = 1
		}
	}
}
