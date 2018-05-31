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

package framework

import (
	"fmt"
	"time"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (f *Framework) MakeBasicRuleFile(ns, name string, groups []monitoringv1.RuleGroup) monitoringv1.RuleFile {
	return monitoringv1.RuleFile{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels: map[string]string{
				"role": "rulefile",
			},
		},
		Spec: monitoringv1.RuleFileSpec{
			Groups: groups,
		},
	}
}

func (f *Framework) CreateRuleFile(ns string, ar monitoringv1.RuleFile) error {
	_, err := f.MonClientV1.RuleFiles(ns).Create(&ar)
	if err != nil {
		return fmt.Errorf("creating %v RuleFile failed: %v", ar.Name, err)
	}

	return nil
}

func (f *Framework) MakeAndCreateFiringRuleFile(ns, name, alertName string) (monitoringv1.RuleFile, error) {
	groups := []monitoringv1.RuleGroup{
		monitoringv1.RuleGroup{
			Name: alertName,
			Rules: []monitoringv1.Rule{
				monitoringv1.Rule{
					Alert: alertName,
					Expr:  "vector(1)",
				},
			},
		},
	}
	file := f.MakeBasicRuleFile(ns, name, groups)

	err := f.CreateRuleFile(ns, file)
	if err != nil {
		return file, err
	}

	return file, nil
}

// WaitForRuleFile waits for a rule file with a given name to exist in a given
// namespace.
func (f *Framework) WaitForRuleFile(ns, name string) error {
	return wait.Poll(time.Second, f.DefaultTimeout, func() (bool, error) {
		_, err := f.MonClientV1.RuleFiles(ns).Get(name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}
		return true, nil
	})
}

func (f *Framework) UpdateRuleFile(ns string, ar monitoringv1.RuleFile) error {
	_, err := f.MonClientV1.RuleFiles(ns).Update(&ar)
	if err != nil {
		return fmt.Errorf("updating %v RuleFile failed: %v", ar.Name, err)
	}

	return nil
}
