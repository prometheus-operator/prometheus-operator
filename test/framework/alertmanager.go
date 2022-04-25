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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/pkg/errors"
	"github.com/prometheus-operator/prometheus-operator/pkg/alertmanager"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/textparse"
)

var ValidAlertmanagerConfig = `global:
  resolve_timeout: 5m
route:
  group_by: ['job']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h
  receiver: 'webhook'
receivers:
- name: 'webhook'
  webhook_configs:
  - url: 'http://alertmanagerwh:30500/'
`

func (f *Framework) MakeBasicAlertmanager(name string, replicas int32) *monitoringv1.Alertmanager {
	return &monitoringv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: monitoringv1.AlertmanagerSpec{
			Replicas: &replicas,
			LogLevel: "debug",
		},
	}
}

func (f *Framework) CreateAlertmanagerConfig(ctx context.Context, ns, name string) (*monitoringv1alpha1.AlertmanagerConfig, error) {
	subRoute := monitoringv1alpha1.Route{
		Receiver: "null",
		Matchers: []monitoringv1alpha1.Matcher{
			{
				Name:  "mykey",
				Value: "myvalue-1",
				Regex: false,
			},
		},
	}
	subRouteJSON, err := json.Marshal(subRoute)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal subroute")
	}

	amConfig := &monitoringv1alpha1.AlertmanagerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: monitoringv1alpha1.AlertmanagerConfigSpec{
			InhibitRules: []monitoringv1alpha1.InhibitRule{
				{
					SourceMatch: []monitoringv1alpha1.Matcher{
						{
							Name:  "mykey",
							Value: "myvalue-1",
							Regex: false,
						},
					},
					TargetMatch: []monitoringv1alpha1.Matcher{
						{
							Name:  "mykey",
							Value: "myvalue-2",
							Regex: false,
						},
					},
					Equal: []string{"equalkey"},
				},
			},
			Receivers: []monitoringv1alpha1.Receiver{
				{
					Name: "null",
				},
			},
			Route: &monitoringv1alpha1.Route{
				Receiver: "null",
				Routes: []extv1.JSON{
					{
						Raw: subRouteJSON,
					},
				},
			},
		},
	}

	return f.MonClientV1alpha1.AlertmanagerConfigs(ns).Create(ctx, amConfig, metav1.CreateOptions{})
}

func (f *Framework) MakeAlertmanagerService(name, group string, serviceType v1.ServiceType) *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("alertmanager-%s", name),
			Labels: map[string]string{
				"group": group,
			},
		},
		Spec: v1.ServiceSpec{
			Type: serviceType,
			Ports: []v1.ServicePort{
				{
					Name:       "web",
					Port:       9093,
					TargetPort: intstr.FromString("web"),
				},
			},
			Selector: map[string]string{
				"alertmanager": name,
			},
		},
	}

	return service
}

func (f *Framework) SecretFromYaml(filepath string) (*v1.Secret, error) {
	manifest, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	s := v1.Secret{}
	err = yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (f *Framework) AlertmanagerConfigSecret(ns, name string) (*v1.Secret, error) {
	s, err := f.SecretFromYaml("../../test/framework/resources/alertmanager-main-secret.yaml")
	if err != nil {
		return nil, err
	}

	s.Name = name
	s.Namespace = ns
	return s, nil
}

func (f *Framework) CreateAlertmanagerAndWaitUntilReady(ctx context.Context, ns string, a *monitoringv1.Alertmanager) (*monitoringv1.Alertmanager, error) {
	amConfigSecretName := fmt.Sprintf("alertmanager-%s", a.Name)
	s, err := f.AlertmanagerConfigSecret(ns, amConfigSecretName)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("making alertmanager config secret %v failed", amConfigSecretName))
	}
	_, err = f.KubeClient.CoreV1().Secrets(ns).Create(ctx, s, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("creating alertmanager config secret %v failed", s.Name))
	}

	a, err = f.MonClientV1.Alertmanagers(ns).Create(ctx, a, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("creating alertmanager %v failed", a.Name))
	}

	return a, f.WaitForAlertmanagerReady(ctx, ns, a.Name, int(*a.Spec.Replicas), a.Spec.ForceEnableClusterMode)
}

// WaitForAlertmanagerReady waits for each individual pod as well as the
// cluster as a whole to be ready.
func (f *Framework) WaitForAlertmanagerReady(ctx context.Context, ns, name string, replicas int, forceEnableClusterMode bool) error {
	if err := f.WaitForPodsReady(
		ctx,
		ns,
		5*time.Minute,
		replicas,
		alertmanager.ListOptions(name),
	); err != nil {
		return errors.Wrap(err,
			fmt.Sprintf(
				"failed to wait for an Alertmanager cluster (%s) with %d instances to become ready",
				name, replicas,
			))
	}

	for i := 0; i < replicas; i++ {
		name := fmt.Sprintf("alertmanager-%v-%v", name, strconv.Itoa(i))
		if err := f.WaitForAlertmanagerInitialized(ctx, ns, name, replicas, forceEnableClusterMode); err != nil {
			return errors.Wrap(err,
				fmt.Sprintf(
					"failed to wait for an Alertmanager cluster (%s) with %d instances to become ready",
					name, replicas,
				),
			)
		}
	}

	return nil
}

func (f *Framework) UpdateAlertmanagerAndWaitUntilReady(ctx context.Context, ns string, a *monitoringv1.Alertmanager) (*monitoringv1.Alertmanager, error) {
	a, err := f.MonClientV1.Alertmanagers(ns).Update(ctx, a, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	err = f.WaitForPodsReady(
		ctx,
		ns,
		5*time.Minute,
		int(*a.Spec.Replicas),
		alertmanager.ListOptions(a.Name),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update %d Alertmanager instances (%s): %v", a.Spec.Replicas, a.Name, err)
	}

	return a, nil
}

func (f *Framework) DeleteAlertmanagerAndWaitUntilGone(ctx context.Context, ns, name string) error {
	_, err := f.MonClientV1.Alertmanagers(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("requesting Alertmanager tpr %v failed", name))
	}

	if err := f.MonClientV1.Alertmanagers(ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("deleting Alertmanager tpr %v failed", name))
	}

	if err := f.WaitForPodsReady(
		ctx,
		ns,
		f.DefaultTimeout,
		0,
		alertmanager.ListOptions(name),
	); err != nil {
		return errors.Wrap(err, fmt.Sprintf("waiting for Alertmanager tpr (%s) to vanish timed out", name))
	}

	return f.KubeClient.CoreV1().Secrets(ns).Delete(ctx, fmt.Sprintf("alertmanager-%s", name), metav1.DeleteOptions{})
}

func (f *Framework) WaitForAlertmanagerInitialized(ctx context.Context, ns, name string, amountPeers int, forceEnableClusterMode bool) error {
	var pollError error
	err := wait.Poll(time.Second, time.Minute*5, func() (bool, error) {

		amStatus, err := f.GetAlertmanagerStatus(ctx, ns, name)
		if err != nil {
			pollError = fmt.Errorf("failed to query Alertmanager: %s", err)
			return false, nil
		}

		isAlertmanagerInClusterMode := amountPeers > 1 || forceEnableClusterMode
		if !isAlertmanagerInClusterMode {
			return true, nil
		}

		if amStatus.Cluster == nil {
			pollError = fmt.Errorf("do not have a cluster status")
			return false, nil
		}

		if *amStatus.Cluster.Status != "ready" {
			pollError = fmt.Errorf("failed to get cluster status, expected ready, got %s", *amStatus.Cluster.Status)
			return false, nil
		}

		if len(amStatus.Cluster.Peers) != amountPeers {

			var addrs = make([]string, len(amStatus.Cluster.Peers))
			for i := range amStatus.Cluster.Peers {
				addrs[i] = *amStatus.Cluster.Peers[i].Name
			}
			pollError = fmt.Errorf("failed to get correct amount of peers, expected %d, got %d, addresses %v", amountPeers, len(amStatus.Cluster.Peers), addrs)
			return false, nil

		}
		return true, nil
	})

	if err != nil {
		return fmt.Errorf("failed to wait for initialized alertmanager cluster: %v: %v", err, pollError)
	}

	return nil
}

func (f *Framework) GetAlertmanagerStatus(ctx context.Context, ns, n string) (models.AlertmanagerStatus, error) {
	var amStatus models.AlertmanagerStatus
	request := f.ProxyGetPod(ns, n, "/api/v2/status")
	resp, err := request.DoRaw(ctx)

	if err != nil {
		return amStatus, err
	}

	if err := json.Unmarshal(resp, &amStatus); err != nil {
		return amStatus, err
	}
	return amStatus, nil
}

func (f *Framework) GetAlertmanagerMetrics(ctx context.Context, ns, n string) (textparse.Parser, error) {
	request := f.ProxyGetPod(ns, n, "/metrics")
	resp, err := request.DoRaw(ctx)
	if err != nil {
		return nil, err
	}
	return textparse.NewPromParser(resp), nil
}

func (f *Framework) CreateSilence(ctx context.Context, ns, n string) (string, error) {
	var createSilenceResponse silence.PostSilencesOKBody

	request := f.ProxyPostPod(
		ns, n,
		"/api/v2/silences",
		`{"createdBy":"Max Mustermann","comment":"1234","startsAt":"2030-04-09T09:16:15.114Z","endsAt":"2031-04-09T11:16:15.114Z","matchers":[{"name":"test","value":"123","isRegex":false}]}`,
	)
	resp, err := request.DoRaw(ctx)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(resp, &createSilenceResponse); err != nil {
		return "", err
	}
	return createSilenceResponse.SilenceID, nil
}

// SendAlertToAlertmanager sends an alert to the alertmanager in the given
// namespace (ns) with the given name (n).
func (f *Framework) SendAlertToAlertmanager(ctx context.Context, ns, n string) error {
	alerts := models.PostableAlerts{{
		Alert: models.Alert{
			GeneratorURL: "http://prometheus-test-0:9090/graph?g0.expr=vector%281%29\u0026g0.tab=1",
			Labels: map[string]string{
				"alertname": "ExampleAlert", "prometheus": "my-prometheus",
			},
		},
	}}

	b, err := json.Marshal(alerts)
	if err != nil {
		return err
	}

	request := f.ProxyPostPod(ns, n, "api/v2/alerts", string(b))
	_, err = request.DoRaw(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (f *Framework) GetSilences(ctx context.Context, ns, n string) (models.GettableSilences, error) {
	var getSilencesResponse models.GettableSilences

	request := f.ProxyGetPod(ns, n, "/api/v2/silences")
	resp, err := request.DoRaw(ctx)
	if err != nil {
		return getSilencesResponse, err
	}

	if err := json.Unmarshal(resp, &getSilencesResponse); err != nil {
		return getSilencesResponse, err
	}

	return getSilencesResponse, nil
}

// PollAlertmanagerConfiguration retrieves the Alertmanager configuration via
// the Alertmanager's API and checks that all conditions return without error.
// It will retry every 10 second for 5 minutes before giving up.
func (f *Framework) PollAlertmanagerConfiguration(ctx context.Context, ns, amName string, conditions ...func(config string) error) error {
	var pollError error
	err := wait.Poll(10*time.Second, time.Minute*5, func() (bool, error) {
		amStatus, err := f.GetAlertmanagerStatus(ctx, ns, "alertmanager-"+amName+"-0")

		if err != nil {
			pollError = fmt.Errorf("failed to query Alertmanager: %s", err)
			return false, nil
		}

		for _, c := range conditions {
			pollError = c(*amStatus.Config.Original)
			if pollError != nil {
				return false, nil
			}
		}

		return true, nil
	})

	if err != nil {
		return fmt.Errorf("failed to wait for alertmanager config: %v: %v", err, pollError)
	}

	return nil
}

func (f *Framework) WaitForAlertmanagerConfigToContainString(ctx context.Context, ns, amName, expected string) error {
	return f.PollAlertmanagerConfiguration(ctx, ns, amName, func(config string) error {
		if !strings.Contains(config, expected) {
			return fmt.Errorf("failed to get matching config expected %q but got %q", expected, config)
		}
		return nil
	})
}

func (f *Framework) WaitForAlertmanagerConfigToBeReloaded(ctx context.Context, ns, amName string, previousReloadTimestamp time.Time) error {
	const configReloadMetricName = "alertmanager_config_last_reload_success_timestamp_seconds"
	err := wait.Poll(10*time.Second, time.Minute*5, func() (bool, error) {
		parser, err := f.GetAlertmanagerMetrics(ctx, ns, "alertmanager-"+amName+"-0")
		if err != nil {
			return false, err
		}

		for {
			entry, err := parser.Next()
			if err != nil {
				return false, err
			}
			if entry == textparse.EntryInvalid {
				return false, fmt.Errorf("invalid prometheus metric entry")
			}
			if entry != textparse.EntrySeries {
				continue
			}

			seriesLabels := labels.Labels{}
			parser.Metric(&seriesLabels)

			if seriesLabels.Get("__name__") != configReloadMetricName {
				continue
			}

			_, _, timestampSec := parser.Series()
			timestamp := time.Unix(int64(timestampSec), 0)
			return timestamp.After(previousReloadTimestamp), nil
		}
	})

	if err != nil {
		return fmt.Errorf("failed to wait for alertmanager config to have been reloaded after %v: %v", previousReloadTimestamp, err)
	}

	return nil
}
