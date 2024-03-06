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
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prometheus "github.com/prometheus-operator/prometheus-operator/pkg/prometheus/server"
)

const (
	SECRET = iota
	CONFIGMAP
)

const (
	ScrapingTLSSecret = "scraping-tls"
	ServerTLSSecret   = "server-tls"
	ServerCASecret    = "server-tls-ca"

	CAKey      = "ca.pem"
	CertKey    = "cert.pem"
	PrivateKey = "key.pem"
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
	ClientKey          Key
	ClientCert         Cert
	CA                 Cert
	InsecureSkipVerify bool
}

func (f *Framework) CreateCertificateResources(namespace, certsDir string, prwtc PromRemoteWriteTestConfig) error {
	var (
		clientKey, clientCert, serverKey, serverCert, caCert []byte
		err                                                  error
	)

	if prwtc.ClientKey.Filename != "" {
		clientKey, err = os.ReadFile(certsDir + prwtc.ClientKey.Filename)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", prwtc.ClientKey.Filename, err)
		}
	}

	if prwtc.ClientCert.Filename != "" {
		clientCert, err = os.ReadFile(certsDir + prwtc.ClientCert.Filename)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", prwtc.ClientCert.Filename, err)
		}
	}

	if prwtc.CA.Filename != "" {
		caCert, err = os.ReadFile(certsDir + prwtc.CA.Filename)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", prwtc.CA.Filename, err)
		}
	}

	serverKey, err = os.ReadFile(certsDir + "ca.key")
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", "ca.key", err)
	}

	serverCert, err = os.ReadFile(certsDir + "ca.crt")
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", "ca.crt", err)
	}

	scrapingKey, err := os.ReadFile(certsDir + "client.key")
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", "client.key", err)
	}

	scrapingCert, err := os.ReadFile(certsDir + "client.crt")
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", "client.crt", err)
	}

	var (
		secrets    = map[string]*v1.Secret{}
		configMaps = map[string]*v1.ConfigMap{}
	)

	secrets[ScrapingTLSSecret] = MakeSecretWithCert(namespace, ScrapingTLSSecret, []string{PrivateKey, CertKey, CAKey}, [][]byte{scrapingKey, scrapingCert, serverCert})
	secrets[ServerTLSSecret] = MakeSecretWithCert(namespace, ServerTLSSecret, []string{PrivateKey, CertKey}, [][]byte{serverKey, serverCert})
	secrets[ServerCASecret] = MakeSecretWithCert(namespace, ServerCASecret, []string{CAKey}, [][]byte{serverCert})

	if len(clientKey) > 0 && len(clientCert) > 0 {
		secrets[prwtc.ClientKey.SecretName] = MakeSecretWithCert(namespace, prwtc.ClientKey.SecretName, []string{PrivateKey}, [][]byte{clientKey})

		if prwtc.ClientCert.ResourceType == CONFIGMAP {
			configMaps[prwtc.ClientCert.ResourceName] = MakeConfigMapWithCert(namespace, prwtc.ClientCert.ResourceName, "", CertKey, "", nil, clientCert, nil)
		} else {
			if _, found := secrets[prwtc.ClientCert.ResourceName]; found {
				secrets[prwtc.ClientCert.ResourceName].Data[CertKey] = clientCert
			} else {
				secrets[prwtc.ClientCert.ResourceName] = MakeSecretWithCert(namespace, prwtc.ClientCert.ResourceName, []string{CertKey}, [][]byte{clientCert})
			}
		}
	}

	if len(caCert) > 0 {
		if prwtc.CA.ResourceType == CONFIGMAP {
			if _, found := configMaps[prwtc.CA.ResourceName]; found {
				configMaps[prwtc.CA.ResourceName].Data[CAKey] = string(caCert)
			} else {
				configMaps[prwtc.CA.ResourceName] = MakeConfigMapWithCert(namespace, prwtc.CA.ResourceName, "", "", CAKey, nil, nil, caCert)
			}
		} else {
			if _, found := secrets[prwtc.CA.ResourceName]; found {
				secrets[prwtc.CA.ResourceName].Data[CAKey] = caCert
			} else {
				secrets[prwtc.CA.ResourceName] = MakeSecretWithCert(namespace, prwtc.CA.ResourceName, []string{CAKey}, [][]byte{caCert})
			}
		}
	}

	for k := range secrets {
		_, err := f.KubeClient.CoreV1().Secrets(namespace).Create(context.Background(), secrets[k], metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create secret: %w", err)
		}
	}

	for k := range configMaps {
		_, err := f.KubeClient.CoreV1().ConfigMaps(namespace).Create(context.Background(), configMaps[k], metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create configmap: %w", err)
		}
	}

	return nil
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

// AddRemoteWriteWithTLSToPrometheus configures Prometheus to send samples to the remote-write endpoint.
func (prwtc PromRemoteWriteTestConfig) AddRemoteWriteWithTLSToPrometheus(p *monitoringv1.Prometheus, url string) {
	p.Spec.RemoteWrite = []monitoringv1.RemoteWriteSpec{{
		URL: url,
		QueueConfig: &monitoringv1.QueueConfig{
			BatchSendDeadline: (*monitoringv1.Duration)(ptr.To("1s")),
		},
	}}

	if (prwtc.ClientKey.SecretName == "" || prwtc.ClientCert.ResourceName == "") && prwtc.CA.ResourceName == "" {
		return
	}

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
			Key: PrivateKey,
		}
		p.Spec.RemoteWrite[0].TLSConfig.Cert = monitoringv1.SecretOrConfigMap{}

		if prwtc.ClientCert.ResourceType == SECRET {
			p.Spec.RemoteWrite[0].TLSConfig.Cert.Secret = &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: prwtc.ClientCert.ResourceName,
				},
				Key: CertKey,
			}
		} else { //certType == CONFIGMAP
			p.Spec.RemoteWrite[0].TLSConfig.Cert.ConfigMap = &v1.ConfigMapKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: prwtc.ClientCert.ResourceName,
				},
				Key: CertKey,
			}
		}
	}

	switch {
	case prwtc.CA.ResourceName != "":
		p.Spec.RemoteWrite[0].TLSConfig.CA = monitoringv1.SecretOrConfigMap{}
		switch prwtc.CA.ResourceType {
		case SECRET:
			p.Spec.RemoteWrite[0].TLSConfig.CA.Secret = &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: prwtc.CA.ResourceName,
				},
				Key: CAKey,
			}
		case CONFIGMAP:
			p.Spec.RemoteWrite[0].TLSConfig.CA.ConfigMap = &v1.ConfigMapKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: prwtc.CA.ResourceName,
				},
				Key: CAKey,
			}
		}

	case prwtc.InsecureSkipVerify:
		p.Spec.RemoteWrite[0].TLSConfig.InsecureSkipVerify = true
	}
}

func (f *Framework) EnableRemoteWriteReceiverWithTLS(p *monitoringv1.Prometheus) {
	p.Spec.EnableFeatures = []string{"remote-write-receiver"}

	p.Spec.Web = &monitoringv1.PrometheusWebSpec{
		WebConfigFileFields: monitoringv1.WebConfigFileFields{
			TLSConfig: &monitoringv1.WebTLSConfig{
				ClientCA: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: ServerCASecret,
						},
						Key: CAKey,
					},
				},
				Cert: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: ServerTLSSecret,
						},
						Key: CertKey,
					},
				},
				KeySecret: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: ServerTLSSecret,
					},
					Key: PrivateKey,
				},
				// Liveness/readiness probes don't work when using "RequireAndVerifyClientCert".
				ClientAuthType: "VerifyClientCertIfGiven",
			},
		},
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
					Port:              "web",
					Interval:          "30s",
					BearerTokenSecret: &v1.SecretKeySelector{},
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
		return nil, fmt.Errorf("creating %d Prometheus instances failed (%v): %v", ptr.Deref(p.Spec.Replicas, 1), p.Name, err)
	}

	result, err = f.WaitForPrometheusReady(ctx, result, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("waiting for %d Prometheus instances timed out (%v): %v", ptr.Deref(p.Spec.Replicas, 1), p.Name, err)
	}

	return result, nil
}

func (f *Framework) UpdatePrometheusReplicasAndWaitUntilReady(ctx context.Context, name, ns string, replicas int32) (*monitoringv1.Prometheus, error) {
	return f.PatchPrometheusAndWaitUntilReady(
		ctx,
		name,
		ns,
		monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Replicas: ptr.To(replicas),
			},
		},
	)
}

func (f *Framework) ScalePrometheusAndWaitUntilReady(ctx context.Context, name, ns string, shards int32) (*monitoringv1.Prometheus, error) {
	promClient := f.MonClientV1.Prometheuses(ns)
	scale, err := promClient.GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Prometheus %s/%s scale: %w", ns, name, err)
	}
	scale.Spec.Replicas = shards

	_, err = promClient.UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to scale Prometheus %s/%s: %w", ns, name, err)
	}
	p, err := promClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Prometheus %s/%s: %w", ns, name, err)
	}
	return f.WaitForPrometheusReady(ctx, p, 5*time.Minute)
}

func (f *Framework) PatchPrometheus(ctx context.Context, name, ns string, spec monitoringv1.PrometheusSpec) (*monitoringv1.Prometheus, error) {
	b, err := json.Marshal(
		&monitoringv1.Prometheus{
			TypeMeta: metav1.TypeMeta{
				Kind:       monitoringv1.PrometheusesKind,
				APIVersion: schema.GroupVersion{Group: monitoring.GroupName, Version: monitoringv1.Version}.String(),
			},
			Spec: spec,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Prometheus spec: %w", err)
	}

	p, err := f.MonClientV1.Prometheuses(ns).Patch(
		ctx,
		name,
		types.ApplyPatchType,
		b,
		metav1.PatchOptions{
			Force:        ptr.To(true),
			FieldManager: "e2e-test",
		},
	)

	if err != nil {
		return nil, err
	}

	return p, nil
}

func (f *Framework) PatchPrometheusAndWaitUntilReady(ctx context.Context, name, ns string, spec monitoringv1.PrometheusSpec) (*monitoringv1.Prometheus, error) {
	p, err := f.PatchPrometheus(ctx, name, ns, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to patch Prometheus %s/%s: %w", ns, name, err)
	}

	p, err = f.WaitForPrometheusReady(ctx, p, 5*time.Minute)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (f *Framework) WaitForPrometheusReady(ctx context.Context, p *monitoringv1.Prometheus, timeout time.Duration) (*monitoringv1.Prometheus, error) {
	expected := *p.Spec.Replicas
	if p.Spec.Shards != nil && *p.Spec.Shards > 0 {
		expected = expected * *p.Spec.Shards
	}

	var current *monitoringv1.Prometheus
	var getErr error
	if err := f.WaitForResourceAvailable(
		ctx,
		func(ctx context.Context) (resourceStatus, error) {
			current, getErr = f.MonClientV1.Prometheuses(p.Namespace).Get(ctx, p.Name, metav1.GetOptions{})
			if getErr != nil {
				return resourceStatus{}, getErr
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
		return nil, fmt.Errorf("prometheus %v/%v failed to become available: %w", p.Namespace, p.Name, err)
	}

	return current, nil
}

func (f *Framework) DeletePrometheusAndWaitUntilGone(ctx context.Context, ns, name string) error {
	_, err := f.MonClientV1.Prometheuses(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("requesting Prometheus custom resource %v failed: %w", name, err)
	}

	if err := f.MonClientV1.Prometheuses(ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("deleting Prometheus custom resource %v failed: %w", name, err)
	}

	if err := f.WaitForPodsReady(
		ctx,
		ns,
		f.DefaultTimeout,
		0,
		prometheus.ListOptions(name),
	); err != nil {
		return fmt.Errorf("waiting for Prometheus custom resource (%s) to vanish timed out: %w", name, err)
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

	if err := wait.PollUntilContextTimeout(ctx, time.Second, time.Minute*5, false, func(ctx context.Context) (bool, error) {
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
	var loopErr error

	err := wait.PollUntilContextTimeout(ctx, time.Second, time.Minute*1, true, func(ctx context.Context) (bool, error) {
		var targets []*Target
		targets, loopErr = f.GetHealthyTargets(ctx, ns, svcName)
		if loopErr != nil {
			return false, nil
		}

		if len(targets) == amount {
			return true, nil
		}

		loopErr = fmt.Errorf("expected %d, found %d healthy targets", amount, len(targets))
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("%s: waiting for healthy targets failed: %v: %v", svcName, err, loopErr)
	}

	return nil
}

func (f *Framework) WaitForDiscoveryWorking(ctx context.Context, ns, svcName, prometheusName string) error {
	var loopErr error

	err := wait.PollUntilContextTimeout(ctx, time.Second, 5*f.DefaultTimeout, false, func(ctx context.Context) (bool, error) {
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
			return nil, fmt.Errorf("target %q: %s", target.ScrapeURL, target.LastError)
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
		return false, fmt.Errorf("expected 1 query result but got %v", len(q.Data.Result))
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
		t.Logf("Prometheus %q/%q (replica #%d) logs:", p.Namespace, p.Name, i)
		t.Logf("%s", l)
	}
}

func (f *Framework) WaitForPrometheusFiringAlert(ctx context.Context, ns, svcName, alertName string) error {
	var loopError error

	err := wait.PollUntilContextTimeout(ctx, time.Second, 5*f.DefaultTimeout, false, func(ctx context.Context) (bool, error) {
		var firing bool
		firing, loopError = f.CheckPrometheusFiringAlert(ctx, ns, svcName, alertName)
		return firing, nil
	})

	if err != nil {
		return fmt.Errorf(
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
