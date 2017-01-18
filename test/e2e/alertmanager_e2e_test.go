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
	"testing"
)

func TestAlertmanagerCreateDeleteCluster(t *testing.T) {
	name := "alertmanager-test"

	defer func() {
		if err := framework.DeleteAlertmanagerAndWaitUntilGone(name); err != nil {
			t.Fatal(err)
		}
	}()

	if err := framework.CreateAlertmanagerAndWaitUntilReady(framework.MakeBasicAlertmanager(name, 3)); err != nil {
		t.Fatal(err)
	}
}

func TestAlertmanagerScaling(t *testing.T) {
	name := "alertmanager-test"

	defer func() {
		if err := framework.DeleteAlertmanagerAndWaitUntilGone(name); err != nil {
			t.Fatal(err)
		}
	}()

	if err := framework.CreateAlertmanagerAndWaitUntilReady(framework.MakeBasicAlertmanager(name, 3)); err != nil {
		t.Fatal(err)
	}

	if err := framework.UpdateAlertmanagerAndWaitUntilReady(framework.MakeBasicAlertmanager(name, 5)); err != nil {
		t.Fatal(err)
	}

	if err := framework.UpdateAlertmanagerAndWaitUntilReady(framework.MakeBasicAlertmanager(name, 3)); err != nil {
		t.Fatal(err)
	}
}

func TestAlertmanagerVersionMigration(t *testing.T) {
	name := "alertmanager-test"

	defer func() {
		if err := framework.DeleteAlertmanagerAndWaitUntilGone(name); err != nil {
			t.Fatal(err)
		}
	}()

	am := framework.MakeBasicAlertmanager(name, 3)
	am.Spec.Version = "v0.5.0"
	if err := framework.CreateAlertmanagerAndWaitUntilReady(am); err != nil {
		t.Fatal(err)
	}

	am.Spec.Version = "v0.5.1"
	if err := framework.UpdateAlertmanagerAndWaitUntilReady(am); err != nil {
		t.Fatal(err)
	}

	am.Spec.Version = "v0.5.0"
	if err := framework.UpdateAlertmanagerAndWaitUntilReady(am); err != nil {
		t.Fatal(err)
	}
}
