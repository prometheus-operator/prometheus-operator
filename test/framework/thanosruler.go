// Copyright 2020 The prometheus-operator Authors
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
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/pkg/errors"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/thanos"
)

func (f *Framework) MakeBasicThanosRuler(name string, replicas int32) *monitoringv1.ThanosRuler {
	return &monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: monitoringv1.ThanosRulerSpec{
			Replicas:       &replicas,
			QueryEndpoints: []string{"test"},
			LogLevel:       "debug",
		},
	}
}

func (f *Framework) CreateThanosRulerAndWaitUntilReady(ns string, tr *monitoringv1.ThanosRuler) (*monitoringv1.ThanosRuler, error) {
	result, err := f.MonClientV1.ThanosRulers(ns).Create(context.TODO(), tr, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("creating %v ThanosRuler instances failed (%v): %v", tr.Spec.Replicas, tr.Name, err)
	}

	if err := f.WaitForThanosRulerReady(result, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("waiting for %v Prometheus instances timed out (%v): %v", tr.Spec.Replicas, tr.Name, err)
	}

	return result, nil
}

func (f *Framework) UpdateThanosRulerAndWaitUntilReady(ns string, tr *monitoringv1.ThanosRuler) (*monitoringv1.ThanosRuler, error) {
	result, err := f.MonClientV1.ThanosRulers(ns).Update(context.TODO(), tr, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	if err := f.WaitForThanosRulerReady(result, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("failed to update %d ThanosRuler instances (%v): %v", tr.Spec.Replicas, tr.Name, err)
	}

	return result, nil
}

func (f *Framework) WaitForThanosRulerReady(tr *monitoringv1.ThanosRuler, timeout time.Duration) error {
	var pollErr error

	err := wait.Poll(2*time.Second, timeout, func() (bool, error) {
		st, _, pollErr := thanos.RulerStatus(context.Background(), f.KubeClient, tr)

		if pollErr != nil {
			return false, nil
		}

		if st.UpdatedReplicas == *tr.Spec.Replicas {
			return true, nil
		}

		return false, nil
	})
	return errors.Wrapf(pollErr, "waiting for ThanosRuler %v/%v: %v", tr.Namespace, tr.Name, err)
}

func (f *Framework) DeleteThanosRulerAndWaitUntilGone(ns, name string) error {
	_, err := f.MonClientV1.ThanosRulers(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("requesting ThanosRuler custom resource %v failed", name))
	}

	if err := f.MonClientV1.ThanosRulers(ns).Delete(context.TODO(), name, metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("deleting ThanosRuler custom resource %v failed", name))
	}

	if err := WaitForPodsReady(
		f.KubeClient,
		ns,
		f.DefaultTimeout,
		0,
		thanos.ListOptions(name),
	); err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("waiting for Prometheus custom resource (%s) to vanish timed out", name),
		)
	}

	return nil
}
