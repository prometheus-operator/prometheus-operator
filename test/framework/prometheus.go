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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/pkg/errors"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

const (
	SECRET    = 0
	CONFIGMAP = 1
)

type Key struct {
	Filename   string
	SecretName string
}

type Cert struct {
	Filename     string
	ResourceName string
	ResourceType int
}

type PromRemoteWriteTestConfig struct {
	Name               string
	ClientKey          Key
	ClientCert         Cert
	CA                 Cert
	InsecureSkipVerify bool
	ShouldSuccess      bool
}

func (f *Framework) MakeBasicPrometheus(ns, name, group string, replicas int32) *monitoringv1.Prometheus {
	return &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: map[string]string{},
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Replicas: &replicas,
				Version:  operator.DefaultPrometheusVersion,
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": group,
					},
				},
				PodMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"group": group,
					},
				},
				ServiceAccountName: "prometheus",
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("400Mi"),
					},
				},
			},
			RuleSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"role": "rulefile",
				},
			},
		},
	}
}

func (f *Framework) AddRemoteWriteWithTLSToPrometheus(p *monitoringv1.Prometheus,
	url string, prwtc PromRemoteWriteTestConfig) {

	p.Spec.RemoteWrite = []monitoringv1.RemoteWriteSpec{{
		URL: url,
	}}

	if (prwtc.ClientKey.SecretName != "" && prwtc.ClientCert.ResourceName != "") || prwtc.CA.ResourceName != "" {

		p.Spec.RemoteWrite[0].TLSConfig = &monitoringv1.TLSConfig{
			SafeTLSConfig: monitoringv1.SafeTLSConfig{
				ServerName: "caandserver.com",
			},
		}

		if prwtc.ClientKey.SecretName != "" && prwtc.ClientCert.ResourceName != "" {
			p.Spec.RemoteWrite[0].TLSConfig.KeySecret = &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: prwtc.ClientKey.SecretName,
				},
				Key: "key.pem",
			}
			p.Spec.RemoteWrite[0].TLSConfig.Cert = monitoringv1.SecretOrConfigMap{}

			if prwtc.ClientCert.ResourceType == SECRET {
				p.Spec.RemoteWrite[0].TLSConfig.Cert.Secret = &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: prwtc.ClientCert.ResourceName,
					},
					Key: "cert.pem",
				}
			} else { //certType == CONFIGMAP
				p.Spec.RemoteWrite[0].TLSConfig.Cert.ConfigMap = &v1.ConfigMapKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: prwtc.ClientCert.ResourceName,
					},
					Key: "cert.pem",
				}
			}
		}

		if prwtc.CA.ResourceName != "" {
			p.Spec.RemoteWrite[0].TLSConfig.CA = monitoringv1.SecretOrConfigMap{}
			if prwtc.CA.ResourceType == SECRET {
				p.Spec.RemoteWrite[0].TLSConfig.CA.Secret = &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: prwtc.CA.ResourceName,
					},
					Key: "ca.pem",
				}
			} else { //caType == CONFIGMAP
				p.Spec.RemoteWrite[0].TLSConfig.CA.ConfigMap = &v1.ConfigMapKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: prwtc.CA.ResourceName,
					},
					Key: "ca.pem",
				}
			}
		} else if prwtc.InsecureSkipVerify {
			p.Spec.RemoteWrite[0].TLSConfig.InsecureSkipVerify = true
		}
	}
}

func (f *Framework) AddRemoteReceiveWithWebTLSToPrometheus(p *monitoringv1.Prometheus, prwtc PromRemoteWriteTestConfig) {
	p.Spec.EnableFeatures = []string{"remote-write-receiver"}

	p.Spec.Web = &monitoringv1.WebSpec{}
	p.Spec.Web.TLSConfig = &monitoringv1.WebTLSConfig{
		ClientCA: monitoringv1.SecretOrConfigMap{
			Secret: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "server-tls-ca",
				},
				Key: "ca.pem",
			},
		},
		Cert: monitoringv1.SecretOrConfigMap{
			Secret: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "server-tls",
				},
				Key: "cert.pem",
			},
		},
		KeySecret: v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: "server-tls",
			},
			Key: "key.pem",
		},
		ClientAuthType: "VerifyClientCertIfGiven",
	}

}

func (f *Framework) AddAlertingToPrometheus(p *monitoringv1.Prometheus, ns, name string) {
	p.Spec.Alerting = &monitoringv1.AlertingSpec{
		Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
			{
				Namespace: ns,
				Name:      fmt.Sprintf("alertmanager-%s", name),
				Port:      intstr.FromString("web"),
			},
		},
	}
}

func (f *Framework) MakeBasicServiceMonitor(name string) *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"group": name,
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group": name,
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "web",
					Interval: "30s",
				},
			},
		},
	}
}

func (f *Framework) MakeBasicPodMonitor(name string) *monitoringv1.PodMonitor {
	return &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"group": name,
			},
		},
		Spec: monitoringv1.PodMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group": name,
				},
			},
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Port:     "web",
					Interval: "30s",
				},
			},
		},
	}
}

func (f *Framework) MakePrometheusService(name, group string, serviceType v1.ServiceType) *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("prometheus-%s", name),
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
				"prometheus": name,
			},
		},
	}
	return service
}

func (f *Framework) MakeThanosQuerierService(name string) *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "http-query",
					Port:       10902,
					TargetPort: intstr.FromString("http"),
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name": "thanos-query",
			},
		},
	}
	return service
}

func (f *Framework) CreatePrometheusAndWaitUntilReady(ctx context.Context, ns string, p *monitoringv1.Prometheus) (*monitoringv1.Prometheus, error) {
	result, err := f.MonClientV1.Prometheuses(ns).Create(ctx, p, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("creating %v Prometheus instances failed (%v): %v", p.Spec.Replicas, p.Name, err)
	}

	if err := f.WaitForPrometheusReady(ctx, result, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("waiting for %v Prometheus instances timed out (%v): %v", p.Spec.Replicas, p.Name, err)
	}

	return result, nil
}

func (f *Framework) UpdatePrometheusAndWaitUntilReady(ctx context.Context, ns string, p *monitoringv1.Prometheus) (*monitoringv1.Prometheus, error) {
	result, err := f.MonClientV1.Prometheuses(ns).Update(ctx, p, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	if err := f.WaitForPrometheusReady(ctx, result, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("failed to update %d Prometheus instances (%v): %v", p.Spec.Replicas, p.Name, err)
	}

	return result, nil
}

func (f *Framework) WaitForPrometheusReady(ctx context.Context, p *monitoringv1.Prometheus, timeout time.Duration) error {
	var pollErr error

	err := wait.Poll(2*time.Second, timeout, func() (bool, error) {
		st, _, pollErr := prometheus.Status(ctx, f.KubeClient, p)

		if pollErr != nil {
			return false, nil
		}

		shards := p.Spec.Shards
		defaultShards := int32(1)
		if shards == nil {
			shards = &defaultShards
		}
		if st.UpdatedReplicas == (*p.Spec.Replicas * *shards) {
			return true, nil
		}

		return false, nil
	})
	return errors.Wrapf(pollErr, "waiting for Prometheus %v/%v: %v", p.Namespace, p.Name, err)
}

func (f *Framework) DeletePrometheusAndWaitUntilGone(ctx context.Context, ns, name string) error {
	_, err := f.MonClientV1.Prometheuses(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("requesting Prometheus custom resource %v failed", name))
	}

	if err := f.MonClientV1.Prometheuses(ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("deleting Prometheus custom resource %v failed", name))
	}

	if err := f.WaitForPodsReady(
		ctx,
		ns,
		f.DefaultTimeout,
		0,
		prometheus.ListOptions(name),
	); err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("waiting for Prometheus custom resource (%s) to vanish timed out", name),
		)
	}

	return nil
}

func (f *Framework) WaitForPrometheusRunImageAndReady(ctx context.Context, ns string, p *monitoringv1.Prometheus) error {
	if err := f.WaitForPodsRunImage(ctx, ns, int(*p.Spec.Replicas), promImage(p.Spec.Version), prometheus.ListOptions(p.Name)); err != nil {
		return err
	}
	return f.WaitForPodsReady(
		ctx,
		ns,
		f.DefaultTimeout,
		int(*p.Spec.Replicas),
		prometheus.ListOptions(p.Name),
	)
}

func promImage(version string) string {
	return fmt.Sprintf("quay.io/prometheus/prometheus:%s", version)
}

// WaitForActiveTargets waits for a number of targets to be configured.
func (f *Framework) WaitForActiveTargets(ctx context.Context, ns, svcName string, amount int) error {
	var targets []*Target

	if err := wait.Poll(time.Second, time.Minute*5, func() (bool, error) {
		var err error
		targets, err = f.GetActiveTargets(ctx, ns, svcName)
		if err != nil {
			return false, err
		}

		if len(targets) == amount {
			return true, nil
		}

		return false, nil
	}); err != nil {
		return fmt.Errorf("waiting for active targets timed out. %v of %v active targets found. %v", len(targets), amount, err)
	}

	return nil
}

// WaitForHealthyTargets waits for a number of targets to be configured and
// healthy.
func (f *Framework) WaitForHealthyTargets(ctx context.Context, ns, svcName string, amount int) error {
	var targets []*Target

	if err := wait.Poll(time.Second, time.Minute*5, func() (bool, error) {
		var err error
		targets, err = f.GetHealthyTargets(ctx, ns, svcName)
		if err != nil {
			return false, err
		}

		if len(targets) == amount {
			return true, nil
		}

		return false, nil
	}); err != nil {
		return fmt.Errorf("waiting for healthy targets timed out. %v of %v healthy targets found. %v", len(targets), amount, err)
	}

	return nil
}

func (f *Framework) WaitForDiscoveryWorking(ctx context.Context, ns, svcName, prometheusName string) error {
	var loopErr error

	err := wait.Poll(time.Second, 5*f.DefaultTimeout, func() (bool, error) {
		pods, loopErr := f.KubeClient.CoreV1().Pods(ns).List(ctx, prometheus.ListOptions(prometheusName))
		if loopErr != nil {
			return false, loopErr
		}
		if 1 != len(pods.Items) {
			return false, nil
		}
		podIP := pods.Items[0].Status.PodIP
		expectedTargets := []string{fmt.Sprintf("http://%s:9090/metrics", podIP)}

		activeTargets, loopErr := f.GetActiveTargets(ctx, ns, svcName)
		if loopErr != nil {
			return false, loopErr
		}

		if loopErr = assertExpectedTargets(activeTargets, expectedTargets); loopErr != nil {
			return false, nil
		}

		working, loopErr := f.basicQueryWorking(ctx, ns, svcName)
		if loopErr != nil {
			return false, loopErr
		}
		if !working {
			return false, nil
		}

		return true, nil
	})

	if err != nil {
		return fmt.Errorf("waiting for Prometheus to discover targets failed: %v: %v", err, loopErr)
	}

	return nil
}

func (f *Framework) basicQueryWorking(ctx context.Context, ns, svcName string) (bool, error) {
	response, err := f.PrometheusSVCGetRequest(ctx, ns, svcName, "http", "/api/v1/query", map[string]string{"query": "up"})
	if err != nil {
		return false, err
	}

	rq := PrometheusQueryAPIResponse{}
	if err := json.NewDecoder(bytes.NewBuffer(response)).Decode(&rq); err != nil {
		return false, err
	}

	if rq.Status != "success" && rq.Data.Result[0].Value[1] == "1" {
		fmt.Printf("Query Response not successful.")
		return false, nil
	}

	return true, nil
}

func assertExpectedTargets(targets []*Target, expectedTargets []string) error {
	existingTargets := []string{}

	for _, t := range targets {
		existingTargets = append(existingTargets, t.ScrapeURL)
	}

	sort.Strings(expectedTargets)
	sort.Strings(existingTargets)

	if !reflect.DeepEqual(expectedTargets, existingTargets) {
		return fmt.Errorf(
			"expected targets %q but got %q", strings.Join(expectedTargets, ","),
			strings.Join(existingTargets, ","),
		)
	}

	return nil
}

func (f *Framework) PrometheusSVCGetRequest(ctx context.Context, ns, svcName, scheme, endpoint string, query map[string]string) ([]byte, error) {
	ProxyGet := f.KubeClient.CoreV1().Services(ns).ProxyGet
	request := ProxyGet(scheme, svcName, "web", endpoint, query)
	return request.DoRaw(ctx)
}

func (f *Framework) GetActiveTargets(ctx context.Context, ns, svcName string) ([]*Target, error) {
	response, err := f.PrometheusSVCGetRequest(ctx, ns, svcName, "http", "/api/v1/targets", map[string]string{})
	if err != nil {
		return nil, err
	}

	rt := prometheusTargetAPIResponse{}
	if err := json.NewDecoder(bytes.NewBuffer(response)).Decode(&rt); err != nil {
		return nil, err
	}

	return rt.Data.ActiveTargets, nil
}

func (f *Framework) GetHealthyTargets(ctx context.Context, ns, svcName string) ([]*Target, error) {
	targets, err := f.GetActiveTargets(ctx, ns, svcName)
	if err != nil {
		return nil, err
	}

	healthyTargets := make([]*Target, 0, len(targets))
	for _, target := range targets {
		switch target.Health {
		case healthGood:
			healthyTargets = append(healthyTargets, target)
		case healthBad:
			return nil, errors.Errorf("target %q: %s", target.ScrapeURL, target.LastError)
		}
	}

	return healthyTargets, nil
}

func (f *Framework) CheckPrometheusFiringAlert(ctx context.Context, ns, svcName, alertName string) (bool, error) {
	response, err := f.PrometheusSVCGetRequest(ctx, ns, svcName, "http", "/api/v1/query", map[string]string{"query": fmt.Sprintf(`ALERTS{alertname="%v",alertstate="firing"}`, alertName)})
	if err != nil {
		return false, err
	}

	q := PrometheusQueryAPIResponse{}
	if err := json.NewDecoder(bytes.NewBuffer(response)).Decode(&q); err != nil {
		return false, err
	}

	if len(q.Data.Result) != 1 {
		return false, errors.Errorf("expected 1 query result but got %v", len(q.Data.Result))
	}

	return true, nil
}

func (f *Framework) PrometheusQuery(ns, svcName, scheme, query string) ([]PrometheusQueryResult, error) {
	response, err := f.PrometheusSVCGetRequest(context.Background(), ns, svcName, scheme, "/api/v1/query", map[string]string{"query": query})
	if err != nil {
		return nil, err
	}

	q := PrometheusQueryAPIResponse{}
	if err := json.NewDecoder(bytes.NewBuffer(response)).Decode(&q); err != nil {
		return nil, err
	}

	if q.Status != "success" {
		return nil, fmt.Errorf("expecting status to be 'success', got %q instead", q.Status)
	}

	return q.Data.Result, nil
}

// PrintPrometheusLogs prints the logs for each Prometheus replica.
func (f *Framework) PrintPrometheusLogs(ctx context.Context, t *testing.T, p *monitoringv1.Prometheus) {
	if p == nil {
		return
	}

	replicas := int(*p.Spec.Replicas)
	for i := 0; i < replicas; i++ {
		l, err := f.GetLogs(ctx, p.Namespace, fmt.Sprintf("prometheus-%s-%d", p.Name, i), "prometheus")
		if err != nil {
			t.Logf("failed to retrieve logs for replica[%d]: %v", i, err)
			continue
		}
		t.Logf("Prometheus #%d replica logs:", i)
		t.Logf("%s", l)
	}
}

func (f *Framework) WaitForPrometheusFiringAlert(ctx context.Context, ns, svcName, alertName string) error {
	var loopError error

	err := wait.Poll(time.Second, 5*f.DefaultTimeout, func() (bool, error) {
		var firing bool
		firing, loopError = f.CheckPrometheusFiringAlert(ctx, ns, svcName, alertName)
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

type targetHealth string

const (
	healthGood targetHealth = "up"
	healthBad  targetHealth = "down"
)

type Target struct {
	ScrapeURL string            `json:"scrapeUrl"`
	Labels    map[string]string `json:"labels"`
	LastError string            `json:"lastError"`
	Health    targetHealth      `json:"health"`
}

type targetDiscovery struct {
	ActiveTargets []*Target `json:"activeTargets"`
}

type prometheusTargetAPIResponse struct {
	Status string           `json:"status"`
	Data   *targetDiscovery `json:"data"`
}

type PrometheusQueryResult struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}

type PrometheusQueryData struct {
	ResultType string                  `json:"resultType"`
	Result     []PrometheusQueryResult `json:"result"`
}

type PrometheusQueryAPIResponse struct {
	Status string               `json:"status"`
	Data   *PrometheusQueryData `json:"data"`
}
