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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/thanos"
)

func (f *Framework) MakeBasicThanosRuler(name string, replicas int32, queryEndpoint string) *monitoringv1.ThanosRuler {
	return &monitoringv1.ThanosRuler{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: monitoringv1.ThanosRulerSpec{
			Replicas:       &replicas,
			QueryEndpoints: []string{queryEndpoint},
			LogLevel:       "debug",
			RuleSelector:   &metav1.LabelSelector{},
		},
	}
}

func (f *Framework) CreateThanosRulerAndWaitUntilReady(ctx context.Context, ns string, tr *monitoringv1.ThanosRuler) (*monitoringv1.ThanosRuler, error) {
	result, err := f.MonClientV1.ThanosRulers(ns).Create(ctx, tr, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("creating %v ThanosRuler instances failed (%v): %v", tr.Spec.Replicas, tr.Name, err)
	}

	if err := f.WaitForThanosRulerReady(ctx, ns, result, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("waiting for %v Prometheus instances timed out (%v): %v", tr.Spec.Replicas, tr.Name, err)
	}

	return result, nil
}

func (f *Framework) PatchThanosRulerAndWaitUntilReady(ctx context.Context, name, ns string, spec monitoringv1.ThanosRulerSpec) (*monitoringv1.ThanosRuler, error) {
	tr, err := f.PatchThanosRuler(ctx, name, ns, spec)
	if err != nil {
		return nil, err
	}

	if err := f.WaitForThanosRulerReady(ctx, ns, tr, 5*time.Minute); err != nil {
		return nil, err
	}

	return tr, nil
}

func (f *Framework) PatchThanosRuler(ctx context.Context, name, ns string, spec monitoringv1.ThanosRulerSpec) (*monitoringv1.ThanosRuler, error) {
	b, err := json.Marshal(
		&monitoringv1.ThanosRuler{
			TypeMeta: metav1.TypeMeta{
				Kind:       monitoringv1.ThanosRulerKind,
				APIVersion: schema.GroupVersion{Group: monitoring.GroupName, Version: monitoringv1.Version}.String(),
			},
			Spec: spec,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal ThanosRuler spec")
	}

	tr, err := f.MonClientV1.ThanosRulers(ns).Patch(
		ctx,
		name,
		types.ApplyPatchType,
		b,
		metav1.PatchOptions{
			Force:        func(b bool) *bool { return &b }(true),
			FieldManager: "e2e-test",
		},
	)

	if err != nil {
		return nil, err
	}

	return tr, nil
}

func (f *Framework) WaitForThanosRulerReady(ctx context.Context, ns string, tr *monitoringv1.ThanosRuler, timeout time.Duration) error {
	expected := *tr.Spec.Replicas

	if err := f.WaitForResourceAvailable(
		ctx,
		func(ctx context.Context) (resourceStatus, error) {
			current, err := f.MonClientV1.ThanosRulers(ns).Get(ctx, tr.Name, metav1.GetOptions{})
			if err != nil {
				return resourceStatus{}, err
			}
			return resourceStatus{
				expectedReplicas: expected,
				generation:       current.Generation,
				replicas:         current.Status.UpdatedReplicas,
				conditions:       current.Status.Conditions,
			}, nil
		},
		timeout,
	); err != nil {
		return errors.Wrapf(err, "thanos ruler %v/%v failed to become available", tr.Namespace, tr.Name)
	}

	return nil
}

func (f *Framework) MakeThanosRulerService(name, group string, serviceType v1.ServiceType) *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("thanos-ruler-%s", name),
			Labels: map[string]string{
				"group": group,
			},
		},
		Spec: v1.ServiceSpec{
			Type: serviceType,
			Ports: []v1.ServicePort{
				{
					Name:       "web",
					Port:       9090,
					TargetPort: intstr.FromString("web"),
				},
			},
			Selector: map[string]string{
				"thanos-ruler": name,
			},
		},
	}
	return service
}

func (f *Framework) WaitForThanosFiringAlert(ctx context.Context, ns, svcName, alertName string) error {
	var loopError error

	err := wait.PollUntilContextTimeout(ctx, time.Second, 5*f.DefaultTimeout, false, func(ctx context.Context) (bool, error) {
		var firing bool
		firing, loopError = f.CheckThanosFiringAlert(ctx, ns, svcName, alertName)
		return firing, nil
	})

	if err != nil {
		return errors.Errorf(
			"waiting for alert '%v' to fire: %v: %v",
			alertName,
			err,
			loopError,
		)
	}
	return nil
}

func (f *Framework) CheckThanosFiringAlert(ctx context.Context, ns, svcName, alertName string) (bool, error) {
	response, err := f.ThanosSVCGetRequest(
		ctx,
		ns,
		svcName,
		"/api/v1/alerts",
		nil,
	)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get Thanos service %s/%s", ns, svcName)
	}

	apiResponse := ThanosAlertsAPIResponse{}
	if err := json.NewDecoder(bytes.NewBuffer(response)).Decode(&apiResponse); err != nil {
		return false, errors.Wrap(err, "failed to decode alerts from Thanos ruler API")
	}

	for _, alert := range apiResponse.Data.Alerts {
		if alert.State != "firing" {
			continue
		}
		if alert.Labels["alertname"] == alertName {
			return true, nil
		}
	}

	return false, fmt.Errorf("failed to find %q alert in the list of %d alerts", alertName, len(apiResponse.Data.Alerts))
}

func (f *Framework) ThanosSVCGetRequest(ctx context.Context, ns, svcName, endpoint string, query map[string]string) ([]byte, error) {
	ProxyGet := f.KubeClient.CoreV1().Services(ns).ProxyGet
	request := ProxyGet("", svcName, "web", endpoint, query)
	return request.DoRaw(ctx)
}

func (f *Framework) DeleteThanosRulerAndWaitUntilGone(ctx context.Context, ns, name string) error {
	_, err := f.MonClientV1.ThanosRulers(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("requesting ThanosRuler custom resource %v failed", name))
	}

	if err := f.MonClientV1.ThanosRulers(ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("deleting ThanosRuler custom resource %v failed", name))
	}

	if err := f.WaitForPodsReady(
		ctx,
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

type ThanosAlert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	State       string            `json:"state"`
	ActiveAt    time.Time         `json:"activeAt"`
	Value       string            `json:"value"`
}

type ThanosAlertsData struct {
	Alerts []ThanosAlert `json:"alerts"`
}

type ThanosAlertsAPIResponse struct {
	Status string            `json:"status"`
	Data   *ThanosAlertsData `json:"data"`
}
