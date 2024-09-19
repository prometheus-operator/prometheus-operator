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
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/proto"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/utils/ptr"

	"github.com/prometheus-operator/prometheus-operator/pkg/alertmanager"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prometheus "github.com/prometheus-operator/prometheus-operator/pkg/prometheus/server"
	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
)

var (
	certsDir = "../../test/e2e/remote_write_certs/"
)

func createMutualTLSSecret(t *testing.T, secretName, ns string) {
	serverCert, err := os.ReadFile(certsDir + "ca.crt")
	require.NoError(t, err)

	scrapingKey, err := os.ReadFile(certsDir + "client.key")
	require.NoError(t, err)

	scrapingCert, err := os.ReadFile(certsDir + "client.crt")
	require.NoError(t, err)

	s := testFramework.MakeSecretWithCert(ns, secretName,
		[]string{"key.pem", "cert.pem", "ca.crt"}, [][]byte{scrapingKey, scrapingCert, serverCert})

	_, err = framework.KubeClient.CoreV1().Secrets(s.ObjectMeta.Namespace).Create(context.Background(), s, metav1.CreateOptions{})
	require.NoError(t, err)
}

func deployInstrumentedApplicationWithTLS(name, ns string) error {
	dep, err := testFramework.MakeDeployment("../../test/framework/resources/basic-auth-app-deployment.yaml")
	if err != nil {
		return err
	}

	dep.Spec.Template.Spec.Containers[0].Args = []string{"--cert-path=/etc/certs"}
	dep.Spec.Template.Spec.Volumes = []v1.Volume{{
		Name: "tls-certs",
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: testFramework.ServerTLSSecret,
			},
		},
	}}

	dep.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
		{
			Name:      dep.Spec.Template.Spec.Volumes[0].Name,
			MountPath: "/etc/certs",
		},
	}

	if err := framework.CreateDeployment(context.Background(), ns, dep); err != nil {
		return fmt.Errorf("failed to create app deployment: %w", err)
	}

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: dep.Name,
			Labels: map[string]string{
				"group": name,
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				{
					Name: "mtls",
					Port: 8081,
				},
			},
			Selector: map[string]string{
				"group": name,
			},
		},
	}

	if _, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
		return fmt.Errorf("failed to create app service: %w", err)
	}

	sm := framework.MakeBasicServiceMonitor(name)
	sm.Spec.Endpoints = []monitoringv1.Endpoint{
		{
			Port:     "mtls",
			Interval: "1s",
			Scheme:   "https",
			TLSConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					ServerName: ptr.To("caandserver.com"),
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: testFramework.ScrapingTLSSecret,
							},
							Key: testFramework.CAKey,
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: testFramework.ScrapingTLSSecret,
							},
							Key: testFramework.CertKey,
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: testFramework.ScrapingTLSSecret,
						},
						Key: testFramework.PrivateKey,
					},
				},
			},
		},
	}

	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), sm, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("failed to create ServiceMonitor: %w", err)
	}

	return nil
}

// createRemoteWriteStack creates a pair of Prometheus objects with the first
// instance scraping targets and remote-writing samples to the second one.
// The 1st and 2nd returned values are the scraping Prometheus object and its service.
// The 3rd and 4th returned values are the receiver Prometheus object and its service.
func createRemoteWriteStack(name, ns string, prwtc testFramework.PromRemoteWriteTestConfig) (*monitoringv1.Prometheus, *v1.Service, *monitoringv1.Prometheus, *v1.Service, error) {
	// Prometheus instance with remote-write receiver enabled.
	receiverName := fmt.Sprintf("%s-%s", name, "receiver")
	rwReceiver := framework.MakeBasicPrometheus(ns, receiverName, receiverName, 1)
	framework.EnableRemoteWriteReceiverWithTLS(rwReceiver)

	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, rwReceiver); err != nil {
		return nil, nil, nil, nil, err
	}

	rwReceiverService := framework.MakePrometheusService(receiverName, receiverName, v1.ServiceTypeClusterIP)
	if _, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, rwReceiverService); err != nil {
		return nil, nil, nil, nil, err
	}

	// Prometheus instance scraping targets.
	prometheus := framework.MakeBasicPrometheus(ns, name, name, 1)
	prwtc.AddRemoteWriteWithTLSToPrometheus(prometheus, "https://"+rwReceiverService.Name+":9090/api/v1/write")
	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheus); err != nil {
		return nil, nil, nil, nil, err
	}

	prometheusService := framework.MakePrometheusService(name, name, v1.ServiceTypeClusterIP)
	if _, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, prometheusService); err != nil {
		return nil, nil, nil, nil, err
	}

	return prometheus, prometheusService, rwReceiver, rwReceiverService, nil
}

func createServiceAccountSecret(t *testing.T, saName, ns string) {
	// Create the secret object
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName + "-sa-secret",
			Namespace: ns,
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": saName,
			},
		},
		Type: v1.SecretTypeServiceAccountToken,
	}

	// Create the secret
	_, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), secret, metav1.CreateOptions{})
	require.NoError(t, err)
	if err != nil {
		fmt.Printf("Failed to create secret: %v\n", err)
		os.Exit(1)
	}
}

func testPromRemoteWriteWithTLS(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		rwConfig testFramework.PromRemoteWriteTestConfig

		success bool
	}{
		{
			// All TLS materials in one secret.
			name: "variant-1",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "client.key",
					SecretName: "client-tls-key-cert-ca",
				},
				ClientCert: testFramework.Cert{
					Filename:     "client.crt",
					ResourceName: "client-tls-key-cert-ca",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "ca.crt",
					ResourceName: "client-tls-key-cert-ca",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: true,
		},
		{
			// TLS materials split into individual secrets.
			name: "variant-2",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "client.key",
					SecretName: "client-tls-key",
				},
				ClientCert: testFramework.Cert{
					Filename:     "client.crt",
					ResourceName: "client-tls-cert",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "ca.crt",
					ResourceName: "client-tls-ca",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: true,
		},
		{
			// client cert/key and CA in different secrets.
			name: "variant-3",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "client.key",
					SecretName: "client-tls-key-cert",
				},
				ClientCert: testFramework.Cert{
					Filename:     "client.crt",
					ResourceName: "client-tls-key-cert",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "ca.crt",
					ResourceName: "client-tls-ca",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: true,
		},
		{
			// client key and client cert/CA in different secrets.
			name: "variant-4",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "client.key",
					SecretName: "client-tls-key",
				},
				ClientCert: testFramework.Cert{
					Filename:     "client.crt",
					ResourceName: "client-tls-cert-ca",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "ca.crt",
					ResourceName: "client-tls-cert-ca",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: true,
		},
		{
			// client cert and client key/CA in different secrets.
			name: "variant-5",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "client.key",
					SecretName: "client-tls-key-ca",
				},
				ClientCert: testFramework.Cert{
					Filename:     "client.crt",
					ResourceName: "client-tls-cert",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "ca.crt",
					ResourceName: "client-tls-key-ca",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: true,
		},
		{
			// client key in secret and client cert/CA in configmap.
			name: "variant-6",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "client.key",
					SecretName: "client-tls-key",
				},
				ClientCert: testFramework.Cert{
					Filename:     "client.crt",
					ResourceName: "client-tls-cert-ca",
					ResourceType: testFramework.CONFIGMAP,
				},
				CA: testFramework.Cert{
					Filename:     "ca.crt",
					ResourceName: "client-tls-cert-ca",
					ResourceType: testFramework.CONFIGMAP,
				},
				InsecureSkipVerify: false,
			},
			success: true,
		},
		{
			// client key in secret and dedicated configmaps for client cert and CA.
			name: "variant-7",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "client.key",
					SecretName: "client-tls-key",
				},
				ClientCert: testFramework.Cert{
					Filename:     "client.crt",
					ResourceName: "client-tls-cert",
					ResourceType: testFramework.CONFIGMAP,
				},
				CA: testFramework.Cert{
					Filename:     "ca.crt",
					ResourceName: "client-tls-ca",
					ResourceType: testFramework.CONFIGMAP,
				},
				InsecureSkipVerify: false,
			},
			success: true,
		},
		{
			// client key/cert in secret and CA in configmap.
			name: "variant-8",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "client.key",
					SecretName: "client-tls-key-cert",
				},
				ClientCert: testFramework.Cert{
					Filename:     "client.crt",
					ResourceName: "client-tls-key-cert",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "ca.crt",
					ResourceName: "client-tls-ca",
					ResourceType: testFramework.CONFIGMAP,
				},
				InsecureSkipVerify: false,
			},
			success: true,
		},
		{
			// client key and cert in dedicated secrets and CA in configmap.
			name: "variant-9",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "client.key",
					SecretName: "client-tls-key",
				},
				ClientCert: testFramework.Cert{
					Filename:     "client.crt",
					ResourceName: "client-tls-cert",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "ca.crt",
					ResourceName: "client-tls-ca",
					ResourceType: testFramework.CONFIGMAP,
				},
				InsecureSkipVerify: false,
			},
			success: true,
		},
		{
			// client key in secret, cert in configmap and CA in secret.
			name: "variant-10",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "client.key",
					SecretName: "client-tls-key-ca",
				},
				ClientCert: testFramework.Cert{
					Filename:     "client.crt",
					ResourceName: "client-tls-cert",
					ResourceType: testFramework.CONFIGMAP,
				},
				CA: testFramework.Cert{
					Filename:     "ca.crt",
					ResourceName: "client-tls-key-ca",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: true,
		},
		{
			// client key in secret, cert in configmap and CA in secret.
			name: "variant-11",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "client.key",
					SecretName: "client-tls-key",
				},
				ClientCert: testFramework.Cert{
					Filename:     "client.crt",
					ResourceName: "client-tls-cert",
					ResourceType: testFramework.CONFIGMAP,
				},
				CA: testFramework.Cert{
					Filename:     "ca.crt",
					ResourceName: "client-tls-ca",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: true,
		},
		{
			// client key/cert in secret and no CA.
			name: "variant-12",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "client.key",
					SecretName: "client-tls-key-cert",
				},
				ClientCert: testFramework.Cert{
					Filename:     "client.crt",
					ResourceName: "client-tls-key-cert",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "",
					ResourceName: "",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: true,
			},
			success: true,
		},
		// non working configurations
		// we will check it only for one configuration for simplicity - only one Secret
		{
			// Invalid CA.
			name: "variant-13",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "client.key",
					SecretName: "client-tls-key-cert-ca",
				},
				ClientCert: testFramework.Cert{
					Filename:     "client.crt",
					ResourceName: "client-tls-key-cert-ca",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "bad_ca.crt",
					ResourceName: "client-tls-key-cert-ca",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: false,
		},
		{
			// Missing CA.
			name: "variant-14",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "client.key",
					SecretName: "client-tls-key-cert",
				},
				ClientCert: testFramework.Cert{
					Filename:     "client.crt",
					ResourceName: "client-tls-key-cert",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "",
					ResourceName: "",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: false,
		},
		{
			// Invalid cert/key + CA.
			name: "variant-15",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "bad_client.key",
					SecretName: "client-tls-key-cert-ca",
				},
				ClientCert: testFramework.Cert{
					Filename:     "bad_client.crt",
					ResourceName: "client-tls-key-cert-ca",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "bad_ca.crt",
					ResourceName: "client-tls-key-cert-ca",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: false,
		},
		{
			// Invalid cert + missing CA.
			name: "variant-16",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "bad_client.key",
					SecretName: "client-tls-key-cert",
				},
				ClientCert: testFramework.Cert{
					Filename:     "bad_client.crt",
					ResourceName: "client-tls-key-cert",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "",
					ResourceName: "",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: false,
		},
		{
			// Missing cert/key + invalid CA.
			name: "variant-17",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "",
					SecretName: "",
				},
				ClientCert: testFramework.Cert{
					Filename:     "",
					ResourceName: "",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "bad_ca.crt",
					ResourceName: "client-tls-ca",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: false,
		},
		{
			// Missing cert/key + CA.
			name: "variant-18",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "",
					SecretName: "",
				},
				ClientCert: testFramework.Cert{
					Filename:     "",
					ResourceName: "",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "",
					ResourceName: "",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: false,
		},
		{
			// Invalid cert/key.
			name: "variant-19",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "bad_client.key",
					SecretName: "client-tls-key-cert-ca",
				},
				ClientCert: testFramework.Cert{
					Filename:     "bad_client.crt",
					ResourceName: "client-tls-key-cert-ca",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "ca.crt",
					ResourceName: "client-tls-key-cert-ca",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: false,
		},
		// Had to change the success flag to True, because prometheus receiver is running in VerifyClientCertIfGiven mode. Details here - https://github.com/prometheus-operator/prometheus-operator/pull/4337#discussion_r735064646
		{
			// Valid CA without cert/key.
			name: "variant-20",
			rwConfig: testFramework.PromRemoteWriteTestConfig{
				ClientKey: testFramework.Key{
					Filename:   "",
					SecretName: "",
				},
				ClientCert: testFramework.Cert{
					Filename:     "",
					ResourceName: "",
					ResourceType: testFramework.SECRET,
				},
				CA: testFramework.Cert{
					Filename:     "ca.crt",
					ResourceName: "client-tls-ca",
					ResourceType: testFramework.SECRET,
				},
				InsecureSkipVerify: false,
			},
			success: true,
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			// The sub-test deploys the following setup:
			//
			// [example app] <---scrapes--- [Prometheus] ---remote-writes---> [Prometheus receiver]
			//
			// When the test expects a success, it should find the samples in the Prometheus receiver.
			// Otherwise the samples should always be found in the scraping Prometheus.
			t.Parallel()

			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)

			ns := framework.CreateNamespace(context.Background(), t, testCtx)
			framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
			name := "test"

			// Create the secrets/configmaps storing the TLS certificates.
			err := framework.CreateCertificateResources(ns, certsDir, tc.rwConfig)
			require.NoError(t, err)

			err = deployInstrumentedApplicationWithTLS(name, ns)
			require.NoError(t, err)

			prometheus, svc, receiver, receiverSvc, err := createRemoteWriteStack(name, ns, tc.rwConfig)
			require.NoError(t, err)

			// Wait for the instrumented application to be scraped.
			if err := framework.WaitForHealthyTargets(context.Background(), ns, svc.Name, 1); err != nil {
				framework.PrintPrometheusLogs(context.Background(), t, prometheus)
				t.Fatal(err)
			}

			// Query metrics from the scraping Prometheus.
			q := "up{container='example-app'} == 1"
			response, err := framework.PrometheusQuery(ns, svc.Name, "http", q)
			require.NoError(t, err)
			require.Len(t, response, 1, "Prometheus does not have the instrumented app metrics: %v", response)

			if !tc.success {
				q = "absent(up)"
			}

			// Query metrics from the remote-write receiver.
			response, err = framework.PrometheusQuery(ns, receiverSvc.Name, "https", q)
			require.NoError(t, err, "(%s, %s, %s): query %q failed: %s", tc.rwConfig.ClientKey.Filename, tc.rwConfig.ClientCert.Filename, tc.rwConfig.CA.Filename, q, err.Error())

			if len(response) != 1 {
				framework.PrintPrometheusLogs(context.Background(), t, prometheus)
				framework.PrintPrometheusLogs(context.Background(), t, receiver)
				t.Fatalf("(%s, %s, %s): query %q failed: %v", tc.rwConfig.ClientKey.Filename, tc.rwConfig.ClientCert.Filename, tc.rwConfig.CA.Filename, q, response)
			}
		})
	}
}

func testPromCreateDeleteCluster(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"

	prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)
	prometheusCRD.Namespace = ns

	_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD)
	require.NoError(t, err)

	err = framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)
}

func testPromScaleUpDownReplicas(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"

	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, framework.MakeBasicPrometheus(ns, name, name, 1))
	require.NoError(t, err)

	p, err = framework.UpdatePrometheusReplicasAndWaitUntilReady(context.Background(), p.Name, ns, *p.Spec.Replicas+1)
	require.NoError(t, err)

	_, err = framework.UpdatePrometheusReplicasAndWaitUntilReady(context.Background(), p.Name, ns, *p.Spec.Replicas-1)
	require.NoError(t, err)
}

func testPromNoServiceMonitorSelector(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"
	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.ServiceMonitorSelector = nil
	_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)
}

func testPromVersionMigration(t *testing.T) {
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"
	startVersion := operator.PrometheusCompatibilityMatrix[0]
	compatibilityMatrix := operator.PrometheusCompatibilityMatrix[1:]

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.Version = startVersion
	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	for _, v := range compatibilityMatrix {
		p, err = framework.PatchPrometheusAndWaitUntilReady(
			context.Background(),
			p.Name,
			ns,
			monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Version: v,
				},
			},
		)
		require.NoError(t, err)
		err := framework.WaitForPrometheusRunImageAndReady(context.Background(), ns, p)
		require.NoError(t, err)
	}
}

func testPromResourceUpdate(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)

	p.Spec.Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("100Mi"),
		},
	}
	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	pods, err := framework.KubeClient.CoreV1().Pods(ns).List(context.Background(), prometheus.ListOptions(name))
	require.NoError(t, err)
	res := pods.Items[0].Spec.Containers[0].Resources

	require.Equal(t, res, p.Spec.Resources, "resources don't match. Has %#+v, want %#+v", res, p.Spec.Resources)

	p, err = framework.PatchPrometheusAndWaitUntilReady(
		context.Background(),
		p.Name,
		ns,
		monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("200Mi"),
					},
				},
			},
		},
	)
	require.NoError(t, err)

	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 2*time.Minute, false, func(ctx context.Context) (bool, error) {
		pods, err := framework.KubeClient.CoreV1().Pods(ns).List(ctx, prometheus.ListOptions(name))
		if err != nil {
			return false, err
		}

		if len(pods.Items) != 1 {
			return false, nil
		}

		res = pods.Items[0].Spec.Containers[0].Resources
		if !reflect.DeepEqual(res, p.Spec.Resources) {
			return false, nil
		}

		return true, nil
	})

	require.NoError(t, err)
}

func testPromStorageLabelsAnnotations(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)

	p.Spec.Storage = &monitoringv1.StorageSpec{
		VolumeClaimTemplate: monitoringv1.EmbeddedPersistentVolumeClaim{
			EmbeddedObjectMetadata: monitoringv1.EmbeddedObjectMetadata{
				Labels: map[string]string{
					"test-label": "foo",
				},
				Annotations: map[string]string{
					"test-annotation": "bar",
				},
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
				Resources: v1.VolumeResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse("200Mi"),
					},
				},
			},
		},
	}

	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	if val := p.Spec.Storage.VolumeClaimTemplate.Labels["test-label"]; val != "foo" {
		t.Errorf("incorrect volume claim label, want: %v, got: %v", "foo", val)
	}
	if val := p.Spec.Storage.VolumeClaimTemplate.Annotations["test-annotation"]; val != "bar" {
		t.Errorf("incorrect volume claim annotation, want: %v, got: %v", "bar", val)
	}

	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 2*time.Minute, false, func(ctx context.Context) (bool, error) {
		sts, err := framework.KubeClient.AppsV1().StatefulSets(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		if len(sts.Items) < 1 {
			return false, nil
		}

		for _, vct := range sts.Items[0].Spec.VolumeClaimTemplates {
			if vct.Name == "prometheus-"+name+"-db" {
				if val := vct.Labels["test-label"]; val != "foo" {
					return false, fmt.Errorf("incorrect volume claim label on sts, want: %v, got: %v", "foo", val)
				}
				if val := vct.Annotations["test-annotation"]; val != "bar" {
					return false, fmt.Errorf("incorrect volume claim annotation on sts, want: %v, got: %v", "bar", val)
				}
				return true, nil
			}
		}

		return false, nil
	})

	require.NoError(t, err)
}

func testPromStorageUpdate(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	p := framework.MakeBasicPrometheus(ns, "test", "", 1)

	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	p, err = framework.PatchPrometheusAndWaitUntilReady(
		context.Background(),
		p.Name,
		ns,
		monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Storage: &monitoringv1.StorageSpec{
					VolumeClaimTemplate: monitoringv1.EmbeddedPersistentVolumeClaim{
						Spec: v1.PersistentVolumeClaimSpec{
							AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
							Resources: v1.VolumeResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceStorage: resource.MustParse("200Mi"),
								},
							},
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)

	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 2*time.Minute, false, func(ctx context.Context) (bool, error) {
		pods, err := framework.KubeClient.CoreV1().Pods(ns).List(ctx, prometheus.ListOptions(p.Name))
		if err != nil {
			return false, err
		}

		if len(pods.Items) != 1 {
			return false, nil
		}

		for _, volume := range pods.Items[0].Spec.Volumes {
			if volume.Name == "prometheus-"+p.Name+"-db" && volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName != "" {
				return true, nil
			}
		}

		return false, nil
	})

	require.NoError(t, err)

	// Invalid storageclass e2e test

	_, err = framework.PatchPrometheus(
		context.Background(),
		p.Name,
		ns,
		monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Storage: &monitoringv1.StorageSpec{
					VolumeClaimTemplate: monitoringv1.EmbeddedPersistentVolumeClaim{
						Spec: v1.PersistentVolumeClaimSpec{
							StorageClassName: ptr.To("unknown-storage-class"),
							Resources: v1.VolumeResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceStorage: resource.MustParse("200Mi"),
								},
							},
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)

	var loopError error
	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, framework.DefaultTimeout, true, func(ctx context.Context) (bool, error) {
		current, err := framework.MonClientV1.Prometheuses(ns).Get(ctx, p.Name, metav1.GetOptions{})
		if err != nil {
			loopError = fmt.Errorf("failed to get object: %w", err)
			return false, nil
		}

		if err := framework.AssertCondition(current.Status.Conditions, monitoringv1.Reconciled, monitoringv1.ConditionFalse); err == nil {
			return true, nil
		}

		return false, nil
	})

	require.NoError(t, err, "%v: %v", err, loopError)
}

func testPromReloadConfig(t *testing.T) {
	for _, tc := range []struct {
		reloadStrategy monitoringv1.ReloadStrategyType
	}{
		{
			reloadStrategy: monitoringv1.HTTPReloadStrategyType,
		},
		{
			reloadStrategy: monitoringv1.ProcessSignalReloadStrategyType,
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("%s reload strategy", tc.reloadStrategy), func(t *testing.T) {
			t.Parallel()
			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)
			ns := framework.CreateNamespace(context.Background(), t, testCtx)
			framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

			name := "test"
			p := framework.MakeBasicPrometheus(ns, name, name, 1)
			p.Spec.ServiceMonitorSelector = nil
			p.Spec.PodMonitorSelector = nil
			p.Spec.ReloadStrategy = ptr.To(tc.reloadStrategy)

			firstConfig := `
global:
  scrape_interval: 1m
scrape_configs:
  - job_name: testReloadConfig
    metrics_path: /metrics
    static_configs:
      - targets:
        - 111.111.111.111:9090
`

			var bufOne bytes.Buffer
			err := operator.GzipConfig(&bufOne, []byte(firstConfig))
			require.NoError(t, err)
			firstConfigCompressed := bufOne.Bytes()

			cfg := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("prometheus-%s", name),
				},
				Data: map[string][]byte{
					"prometheus.yaml.gz": firstConfigCompressed,
					"configmaps.json":    []byte("{}"),
				},
			}

			svc := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)

			_, err = framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), cfg, metav1.CreateOptions{})
			require.NoError(t, err)
			_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
			require.NoError(t, err)

			finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
			require.NoError(t, err)
			testCtx.AddFinalizerFn(finalizerFn)

			err = framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 1)
			require.NoError(t, err)

			secondConfig := `
global:
  scrape_interval: 1m
scrape_configs:
  - job_name: testReloadConfig
    metrics_path: /metrics
    static_configs:
      - targets:
        - 111.111.111.111:9090
        - 111.111.111.112:9090
`

			var bufTwo bytes.Buffer
			err = operator.GzipConfig(&bufTwo, []byte(secondConfig))
			require.NoError(t, err)
			secondConfigCompressed := bufTwo.Bytes()

			cfg, err = framework.KubeClient.CoreV1().Secrets(ns).Get(context.Background(), cfg.Name, metav1.GetOptions{})
			require.NoError(t, err)

			cfg.Data["prometheus.yaml.gz"] = secondConfigCompressed
			_, err = framework.KubeClient.CoreV1().Secrets(ns).Update(context.Background(), cfg, metav1.UpdateOptions{})
			require.NoError(t, err)

			err = framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 2)
			require.NoError(t, err)
		})
	}
}

func testPromAdditionalScrapeConfig(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	prometheusName := "test"
	group := "additional-config-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	s := framework.MakeBasicServiceMonitor(group)
	_, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{})
	require.NoError(t, err)

	additionalConfig := `
- job_name: "prometheus"
  static_configs:
  - targets: ["localhost:9090"]
`
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "additional-scrape-configs",
		},
		Data: map[string][]byte{
			"prometheus-additional.yaml": []byte(additionalConfig),
		},
	}
	_, err = framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), &secret, metav1.CreateOptions{})
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	p.Spec.AdditionalScrapeConfigs = &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: "additional-scrape-configs",
		},
		Key: "prometheus-additional.yaml",
	}
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
	require.NoError(t, err)

	testCtx.AddFinalizerFn(finalizerFn)

	// Wait for ServiceMonitor target, as well as additional-config target
	err = framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 2)
	require.NoError(t, err)
}

func testPromAdditionalAlertManagerConfig(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	prometheusName := "test"
	group := "additional-alert-config-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	s := framework.MakeBasicServiceMonitor(group)
	_, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{})
	require.NoError(t, err)

	additionalConfig := `
- path_prefix: /
  scheme: http
  static_configs:
  - targets: ["localhost:9093"]
`
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "additional-alert-configs",
		},
		Data: map[string][]byte{
			"prometheus-additional.yaml": []byte(additionalConfig),
		},
	}
	_, err = framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), &secret, metav1.CreateOptions{})
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	p.Spec.AdditionalAlertManagerConfigs = &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: "additional-alert-configs",
		},
		Key: "prometheus-additional.yaml",
	}
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
	require.NoError(t, err)

	testCtx.AddFinalizerFn(finalizerFn)

	// Wait for ServiceMonitor target
	err = framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 1)
	require.NoError(t, err)

	err = wait.PollUntilContextTimeout(context.Background(), time.Second, 5*time.Minute, false, func(ctx context.Context) (done bool, err error) {
		response, err := framework.PrometheusSVCGetRequest(ctx, ns, svc.Name, "http", "/api/v1/alertmanagers", map[string]string{})
		if err != nil {
			return true, err
		}

		ra := prometheusAlertmanagerAPIResponse{}
		if err := json.NewDecoder(bytes.NewBuffer(response)).Decode(&ra); err != nil {
			return true, err
		}

		if ra.Status == "success" && len(ra.Data.ActiveAlertmanagers) == 1 {
			return true, nil
		}

		return false, nil
	})

	require.NoError(t, err)
}

func testPromReloadRules(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"
	firtAlertName := "firstAlert"
	secondAlertName := "secondAlert"

	ruleFile, err := framework.MakeAndCreateFiringRule(context.Background(), ns, name, firtAlertName)
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	p, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
	finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, pSVC)
	require.NoError(t, err)

	testCtx.AddFinalizerFn(finalizerFn)

	err = framework.WaitForPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, firtAlertName)
	require.NoError(t, err)

	ruleFile.Spec.Groups = []monitoringv1.RuleGroup{
		{
			Name: "my-alerting-group",
			Rules: []monitoringv1.Rule{
				{
					Alert: secondAlertName,
					Expr:  intstr.FromString("vector(1)"),
				},
			},
		},
	}
	_, err = framework.UpdateRule(context.Background(), ns, ruleFile)
	require.NoError(t, err)

	err = framework.WaitForPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, secondAlertName)
	require.NoError(t, err)
}

func testPromMultiplePrometheusRulesSameNS(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"
	alertNames := []string{"first-alert", "second-alert"}

	for _, alertName := range alertNames {
		_, err := framework.MakeAndCreateFiringRule(context.Background(), ns, alertName, alertName)
		require.NoError(t, err)
	}

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
	finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, pSVC)
	require.NoError(t, err)

	testCtx.AddFinalizerFn(finalizerFn)

	for _, alertName := range alertNames {
		err := framework.WaitForPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, alertName)
		require.NoError(t, err)
	}
}

func testPromMultiplePrometheusRulesDifferentNS(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	rootNS := framework.CreateNamespace(context.Background(), t, testCtx)
	alertNSOne := framework.CreateNamespace(context.Background(), t, testCtx)
	alertNSTwo := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, rootNS)

	name := "test"
	ruleFiles := []struct {
		alertName string
		ns        string
	}{{"first-alert", alertNSOne}, {"second-alert", alertNSTwo}}

	ruleFilesNamespaceSelector := map[string]string{"monitored": "true"}

	for _, file := range ruleFiles {
		err := framework.AddLabelsToNamespace(context.Background(), file.ns, ruleFilesNamespaceSelector)
		require.NoError(t, err)
	}

	for _, file := range ruleFiles {
		_, err := framework.MakeAndCreateFiringRule(context.Background(), file.ns, file.alertName, file.alertName)
		require.NoError(t, err)
	}

	p := framework.MakeBasicPrometheus(rootNS, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	p.Spec.RuleNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: ruleFilesNamespaceSelector,
	}
	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), rootNS, p)
	require.NoError(t, err)

	pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
	finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), rootNS, pSVC)
	require.NoError(t, err)

	testCtx.AddFinalizerFn(finalizerFn)

	for _, file := range ruleFiles {
		err := framework.WaitForPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, file.alertName)
		require.NoError(t, err)
	}

	// Remove the selecting label from the namespaces holding PrometheusRules
	// and wait until the rules are removed from Prometheus.
	// See https://github.com/prometheus-operator/prometheus-operator/issues/3847
	for _, file := range ruleFiles {
		err := framework.RemoveLabelsFromNamespace(context.Background(), file.ns, "monitored")
		require.NoError(t, err)
	}

	for _, file := range ruleFiles {
		var loopError error
		err = wait.PollUntilContextTimeout(context.Background(), time.Second, 5*framework.DefaultTimeout, false, func(ctx context.Context) (bool, error) {
			var firing bool
			firing, loopError = framework.CheckPrometheusFiringAlert(ctx, file.ns, pSVC.Name, file.alertName)
			return !firing, nil
		})

		require.NoError(t, err, "waiting for alert %q in namespace %s to stop firing: %v: %v", file.alertName, file.ns, err, loopError)
	}
}

func testPromRulesExceedingConfigMapLimit(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	prometheusRules := []*monitoringv1.PrometheusRule{}
	for i := 0; i < 2; i++ {
		rule := generateHugePrometheusRule(ns, strconv.Itoa(i))
		rule, err := framework.CreateRule(context.Background(), ns, rule)
		require.NoError(t, err)
		prometheusRules = append(prometheusRules, rule)
	}

	name := "test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	defer func() {
		if !t.Failed() {
			return
		}

		b := &bytes.Buffer{}
		if err := framework.WritePodLogs(context.Background(), b, ns, "prometheus-"+p.Name+"-0", testFramework.LogOptions{}); err != nil {
			t.Logf("failed to get logs: %v", err)
		}
		t.Log(b.String())
	}()

	pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
	finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, pSVC)
	require.NoError(t, err)

	testCtx.AddFinalizerFn(finalizerFn)

	for i := range prometheusRules {
		_, err := framework.WaitForConfigMapExist(context.Background(), ns, "prometheus-"+p.Name+"-rulefiles-"+strconv.Itoa(i))
		require.NoError(t, err)
	}

	// Make sure both rule files ended up in the Prometheus Pod
	for i := range prometheusRules {
		err := framework.WaitForPrometheusFiringAlert(context.Background(), ns, pSVC.Name, "my-alert-"+strconv.Itoa(i))
		require.NoError(t, err)
	}

	err = framework.DeleteRule(context.Background(), ns, prometheusRules[1].Name)
	require.NoError(t, err)

	_, err = framework.WaitForConfigMapExist(context.Background(), ns, "prometheus-"+p.Name+"-rulefiles-0")
	require.NoError(t, err)
	err = framework.WaitForConfigMapNotExist(context.Background(), ns, "prometheus-"+p.Name+"-rulefiles-1")
	require.NoError(t, err)

	err = framework.WaitForPrometheusFiringAlert(context.Background(), ns, pSVC.Name, "my-alert-0")
	require.NoError(t, err)
}

func testPromRulesMustBeAnnotated(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "admission"
	admissionAlert := "admissionAlert"

	_, err := framework.MakeAndCreateFiringRule(context.Background(), ns, name, admissionAlert)
	require.NoError(t, err)

	rule, err := framework.GetRule(context.Background(), ns, name)
	require.NoError(t, err)

	val, ok := rule.Annotations["prometheus-operator-validated"]
	require.True(t, ok, "Expected prometheusrule to be annotated")
	require.Equal(t, "true", val, "Expected prometheusrule annotation to be 'true'")
}

func testInvalidRulesAreRejected(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "admission"
	admissionAlert := "admissionAlert"

	_, err := framework.MakeAndCreateInvalidRule(context.Background(), ns, name, admissionAlert)
	require.Error(t, err, "Expected invalid prometheusrule to be rejected")
}

func testPromReconcileStatusWhenInvalidRuleCreated(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
	ruleFilesNamespaceSelector := map[string]string{"excludeFromWebhook": "true", "role": "rulefile"}

	err := framework.AddLabelsToNamespace(context.Background(), ns, ruleFilesNamespaceSelector)
	require.NoError(t, err)

	ruleName := "invalidrule"
	alertName := "invalidalert"

	_, err = framework.MakeAndCreateInvalidRule(context.Background(), ns, ruleName, alertName)
	require.NoError(t, err, "expected invalid rule to be created in namespace %v", err)

	prom := framework.MakeBasicPrometheus(ns, "basic-prometheus", "test-group", 1)
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prom)
	require.NoError(t, err)
}

// generateHugePrometheusRule returns a Prometheus rule instance that would fill
// more than half of the space of a Kubernetes ConfigMap.
func generateHugePrometheusRule(ns, identifier string) *monitoringv1.PrometheusRule {
	alertName := "my-alert"
	groups := []monitoringv1.RuleGroup{
		{
			Name:  alertName,
			Rules: []monitoringv1.Rule{},
		},
	}
	// One rule marshaled as yaml is ~34 bytes long, the max is ~524288 bytes.
	for i := 0; i < 12000; i++ {
		groups[0].Rules = append(groups[0].Rules, monitoringv1.Rule{
			Alert: alertName + "-" + identifier,
			Expr:  intstr.FromString("vector(1)"),
		})
	}
	rule := framework.MakeBasicRule(ns, "prometheus-rule-"+identifier, groups)

	return rule
}

// Make sure the Prometheus operator only updates the Prometheus config secret
// and the Prometheus rules configmap on relevant changes.
func testPromOnlyUpdatedOnRelevantChanges(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"
	prometheus := framework.MakeBasicPrometheus(ns, name, name, 1)

	// Adding an annotation to Prometheus lead to high CPU usage in the past
	// updating the Prometheus StatefulSet in a loop (See
	// https://github.com/prometheus-operator/prometheus-operator/issues/1659). Added here to
	// prevent a regression.
	prometheus.Annotations["test-annotation"] = "test-value"

	ctx, cancel := context.WithCancel(context.Background())

	type versionedResource interface {
		GetResourceVersion() string
	}

	resourceDefinitions := []struct {
		Name               string
		Getter             func(prometheusName string) (versionedResource, error)
		Versions           map[string]interface{}
		MaxExpectedChanges int
	}{
		{
			Name: "rulesConfigMap",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					KubeClient.
					CoreV1().
					ConfigMaps(ns).
					Get(context.Background(), "prometheus-"+prometheusName+"-rulefiles-0", metav1.GetOptions{})
			},
			// The Prometheus Operator first creates the ConfigMap for the
			// given Prometheus stateful set and then updates it with the matching
			// Prometheus rules.
			MaxExpectedChanges: 2,
		},
		{
			Name: "configurationSecret",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					KubeClient.
					CoreV1().
					Secrets(ns).
					Get(context.Background(), "prometheus-"+prometheusName, metav1.GetOptions{})
			},
			MaxExpectedChanges: 2,
		},
		{
			Name: "tlsAssetSecret",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					KubeClient.
					CoreV1().
					Secrets(ns).
					Get(context.Background(), "prometheus-"+prometheusName+"-tls-assets-0", metav1.GetOptions{})
			},
			MaxExpectedChanges: 2,
		},
		{
			Name: "statefulset",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					KubeClient.
					AppsV1().
					StatefulSets(ns).
					Get(context.Background(), "prometheus-"+prometheusName, metav1.GetOptions{})
			},
			// First is the creation of the StatefulSet itself, following is the
			// update of e.g. the ReadyReplicas status field
			MaxExpectedChanges: 3,
		},
		{
			Name: "service-operated",
			Getter: func(_ string) (versionedResource, error) {
				return framework.
					KubeClient.
					CoreV1().
					Services(ns).
					Get(context.Background(), "prometheus-operated", metav1.GetOptions{})
			},
			MaxExpectedChanges: 1,
		},
		{
			Name: "serviceMonitor",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					MonClientV1.
					ServiceMonitors(ns).
					Get(context.Background(), prometheusName, metav1.GetOptions{})
			},
			MaxExpectedChanges: 1,
		},
	}

	// Init Versions maps
	for i := range resourceDefinitions {
		resourceDefinitions[i].Versions = map[string]interface{}{}
	}

	errc := make(chan error, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(10 * time.Millisecond)

				for i, resourceDef := range resourceDefinitions {
					resource, err := resourceDef.Getter(prometheus.Name)
					if apierrors.IsNotFound(err) {
						continue
					}
					if err != nil {
						cancel()
						errc <- err
						return
					}

					resourceDefinitions[i].Versions[resource.GetResourceVersion()] = resource
				}
			}
		}
	}()

	alertName := "my-alert"
	_, err := framework.MakeAndCreateFiringRule(context.Background(), ns, "my-prometheus-rule", alertName)
	require.NoError(t, err)

	prometheus, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheus)
	require.NoError(t, err)

	pSVC := framework.MakePrometheusService(prometheus.Name, name, v1.ServiceTypeClusterIP)
	finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, pSVC)
	require.NoError(t, err)

	testCtx.AddFinalizerFn(finalizerFn)

	s := framework.MakeBasicServiceMonitor(name)
	_, err = framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{})
	require.NoError(t, err)

	err = framework.WaitForPrometheusFiringAlert(context.Background(), prometheus.Namespace, pSVC.Name, alertName)
	require.NoError(t, err)

	err = framework.WaitForDiscoveryWorking(context.Background(), ns, pSVC.Name, prometheus.Name)
	require.NoError(t, err)

	err = framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)

	cancel()

	select {
	case err := <-errc:
		t.Fatal(err)
	default:
	}

	for _, resource := range resourceDefinitions {
		if len(resource.Versions) > resource.MaxExpectedChanges || len(resource.Versions) < 1 {
			var previous interface{}
			for _, version := range resource.Versions {
				if previous == nil {
					previous = version
					continue
				}
				t.Log(pretty.Compare(previous, version))
				previous = version
			}

			t.Fatalf(
				"expected resource %v to be created/updated %v times, but saw %v instead",
				resource.Name,
				resource.MaxExpectedChanges,
				len(resource.Versions),
			)
		}
	}
}

func testPromPreserveUserAddedMetadata(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"

	prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)
	prometheusCRD.Namespace = ns

	prometheusCRD, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD)
	require.NoError(t, err)

	updatedLabels := map[string]string{
		"user-defined-label": "custom-label-value",
	}
	updatedAnnotations := map[string]string{
		"user-defined-annotation": "custom-annotation-val",
	}

	svcClient := framework.KubeClient.CoreV1().Services(ns)
	endpointsClient := framework.KubeClient.CoreV1().Endpoints(ns)
	ssetClient := framework.KubeClient.AppsV1().StatefulSets(ns)
	secretClient := framework.KubeClient.CoreV1().Secrets(ns)

	resourceConfigs := []struct {
		name   string
		get    func() (metav1.Object, error)
		update func(object metav1.Object) (metav1.Object, error)
	}{
		{
			name: "prometheus-operated service",
			get: func() (metav1.Object, error) {
				return svcClient.Get(context.Background(), "prometheus-operated", metav1.GetOptions{})
			},
			update: func(object metav1.Object) (metav1.Object, error) {
				return svcClient.Update(context.Background(), asService(t, object), metav1.UpdateOptions{})
			},
		},
		{
			name: "prometheus stateful set",
			get: func() (metav1.Object, error) {
				return ssetClient.Get(context.Background(), "prometheus-test", metav1.GetOptions{})
			},
			update: func(object metav1.Object) (metav1.Object, error) {
				return ssetClient.Update(context.Background(), asStatefulSet(t, object), metav1.UpdateOptions{})
			},
		},
		{
			name: "prometheus-operated endpoints",
			get: func() (metav1.Object, error) {
				return endpointsClient.Get(context.Background(), "prometheus-operated", metav1.GetOptions{})
			},
			update: func(object metav1.Object) (metav1.Object, error) {
				return endpointsClient.Update(context.Background(), asEndpoints(t, object), metav1.UpdateOptions{})
			},
		},
		{
			name: "prometheus secret",
			get: func() (metav1.Object, error) {
				return secretClient.Get(context.Background(), "prometheus-test", metav1.GetOptions{})
			},
			update: func(object metav1.Object) (metav1.Object, error) {
				return secretClient.Update(context.Background(), asSecret(t, object), metav1.UpdateOptions{})
			},
		},
	}

	for _, rConf := range resourceConfigs {
		res, err := rConf.get()
		require.NoError(t, err)

		updateObjectLabels(res, updatedLabels)
		updateObjectAnnotations(res, updatedAnnotations)

		_, err = rConf.update(res)
		require.NoError(t, err)
	}

	// Ensure resource reconciles
	_, err = framework.UpdatePrometheusReplicasAndWaitUntilReady(context.Background(), prometheusCRD.Name, ns, *prometheusCRD.Spec.Replicas+1)
	require.NoError(t, err)

	// Assert labels preserved
	for _, rConf := range resourceConfigs {
		res, err := rConf.get()
		require.NoError(t, err)

		labels := res.GetLabels()
		if !containsValues(labels, updatedLabels) {
			t.Errorf("%s: labels do not contain updated labels, found: %q, should contain: %q", rConf.name, labels, updatedLabels)
		}

		annotations := res.GetAnnotations()
		require.True(t, containsValues(annotations, updatedAnnotations))
	}

	// Cleanup
	err = framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, name)
	require.NoError(t, err)
}

func asService(t *testing.T, object metav1.Object) *v1.Service {
	svc, ok := object.(*v1.Service)
	require.True(t, ok, "expected service got %T", object)
	return svc
}

func asEndpoints(t *testing.T, object metav1.Object) *v1.Endpoints {
	endpoints, ok := object.(*v1.Endpoints)
	require.True(t, ok, "expected endpoints got %T", object)
	return endpoints
}

func asStatefulSet(t *testing.T, object metav1.Object) *appsv1.StatefulSet {
	sset, ok := object.(*appsv1.StatefulSet)
	require.True(t, ok, "expected stateful set got %T", object)
	return sset
}

func asSecret(t *testing.T, object metav1.Object) *v1.Secret {
	sec, ok := object.(*v1.Secret)
	require.True(t, ok, "expected secret set got %T", object)
	return sec
}

func containsValues(got, expected map[string]string) bool {
	for k, v := range expected {
		if got[k] != v {
			return false
		}
	}
	return true
}

func updateObjectLabels(object metav1.Object, labels map[string]string) {
	current := object.GetLabels()
	current = mergeMap(current, labels)
	object.SetLabels(current)
}

func updateObjectAnnotations(object metav1.Object, annotations map[string]string) {
	current := object.GetAnnotations()
	current = mergeMap(current, annotations)
	object.SetAnnotations(current)
}

func mergeMap(a, b map[string]string) map[string]string {
	if a == nil {
		a = make(map[string]string, len(b))
	}
	for k, v := range b {
		a[k] = v
	}
	return a
}

func testPromWhenDeleteCRDCleanUpViaOwnerRef(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	configMapName := fmt.Sprintf("prometheus-%v-rulefiles-0", p.Name)

	_, err = framework.WaitForConfigMapExist(context.Background(), ns, configMapName)
	require.NoError(t, err)

	// Waits for Prometheus pods to vanish
	err = framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, p.Name)
	require.NoError(t, err)

	err = framework.WaitForConfigMapNotExist(context.Background(), ns, configMapName)
	require.NoError(t, err)
}

func testPromDiscovery(t *testing.T) {
	for _, tc := range []struct {
		role *monitoringv1.ServiceDiscoveryRole
	}{
		{
			role: nil,
		},
		{
			role: ptr.To(monitoringv1.EndpointsRole),
		},
		{
			role: ptr.To(monitoringv1.EndpointSliceRole),
		},
	} {
		t.Run(fmt.Sprintf("role=%s", ptr.Deref(tc.role, "<nil>")), func(t *testing.T) {
			t.Parallel()
			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)
			ns := framework.CreateNamespace(context.Background(), t, testCtx)
			framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

			prometheusName := "test"
			group := "servicediscovery-test"
			svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

			s := framework.MakeBasicServiceMonitor(group)
			_, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{})
			require.NoError(t, err)

			p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
			p.Spec.ServiceDiscoveryRole = tc.role
			_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
			require.NoError(t, err)

			finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
			require.NoError(t, err)

			testCtx.AddFinalizerFn(finalizerFn)

			_, err = framework.KubeClient.CoreV1().Secrets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
			require.NoError(t, err)

			err = framework.WaitForDiscoveryWorking(context.Background(), ns, svc.Name, prometheusName)
			require.NoError(t, err)
		})
	}
}

func testPromSharedResourcesReconciliation(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	s := framework.MakeBasicServiceMonitor("reconcile-test")
	_, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{})
	require.NoError(t, err)

	// Create 2 Prometheus different Prometheus instances that watch the service monitor created above.
	for _, prometheusName := range []string{"test", "test2"} {
		p := framework.MakeBasicPrometheus(ns, prometheusName, "reconcile-test", 1)
		_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
		require.NoError(t, err)

		svc := framework.MakePrometheusService(prometheusName, fmt.Sprintf("reconcile-%s", prometheusName), v1.ServiceTypeClusterIP)
		finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
		require.NoError(t, err)

		testCtx.AddFinalizerFn(finalizerFn)

		_, err = framework.KubeClient.CoreV1().Secrets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
		require.NoError(t, err, "Generated Secret could not be retrieved for %s: %v", prometheusName, err)

		err = framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 1)
		require.NoError(t, err)
	}

	err = framework.MonClientV1.ServiceMonitors(ns).Delete(context.Background(), "reconcile-test", metav1.DeleteOptions{})
	require.NoError(t, err)

	// Delete the service monitors and check that both Prometheus instances are updated.
	for _, prometheusName := range []string{"test", "test2"} {
		svc := framework.MakePrometheusService(prometheusName, fmt.Sprintf("reconcile-%s", prometheusName), v1.ServiceTypeClusterIP)

		err := framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 0)
		require.NoError(t, err)
	}
}

func testShardingProvisioning(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	prometheusName := "test"
	group := "servicediscovery-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	s := framework.MakeBasicServiceMonitor(group)
	_, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{})
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	shards := int32(2)
	p.Spec.Shards = &shards
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
	require.NoError(t, err)

	testCtx.AddFinalizerFn(finalizerFn)

	pods := []struct {
		pod                        string
		expectedShardConfigSnippet string
	}{
		{
			pod: "prometheus-test-0",
			expectedShardConfigSnippet: `
  - source_labels:
    - __tmp_hash
    regex: 0
    action: keep`,
		}, {
			pod: "prometheus-test-shard-1-0",
			expectedShardConfigSnippet: `
  - source_labels:
    - __tmp_hash
    regex: 1
    action: keep`,
		},
	}

	for _, p := range pods {
		stdout, _, err := framework.ExecWithOptions(context.Background(), testFramework.ExecOptions{
			Command: []string{
				"/bin/sh", "-c", "cat /etc/prometheus/config_out/prometheus.env.yaml",
			},
			Namespace:     ns,
			PodName:       p.pod,
			ContainerName: "prometheus",
			CaptureStdout: true,
			CaptureStderr: true,
			Stdin:         nil,
		})
		require.NoError(t, err)
		require.Contains(t, p.expectedShardConfigSnippet, stdout)
	}
}

func testResharding(t *testing.T) {
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	prometheusName := "test"
	group := "servicediscovery-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	s := framework.MakeBasicServiceMonitor(group)
	_, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{})
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	p, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
	require.NoError(t, err)

	testCtx.AddFinalizerFn(finalizerFn)

	shards := int32(2)
	p, err = framework.PatchPrometheusAndWaitUntilReady(
		context.Background(),
		p.Name,
		ns,
		monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Shards: &shards,
			},
		},
	)
	require.NoError(t, err)

	_, err = framework.KubeClient.AppsV1().StatefulSets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", p.Name), metav1.GetOptions{})
	require.NoError(t, err)

	_, err = framework.KubeClient.AppsV1().StatefulSets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s-shard-1", p.Name), metav1.GetOptions{})
	require.NoError(t, err)

	shards = int32(1)
	p, err = framework.PatchPrometheusAndWaitUntilReady(
		context.Background(),
		p.Name,
		ns,
		monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Shards: &shards,
			},
		},
	)
	require.NoError(t, err)

	_, err = framework.KubeClient.AppsV1().StatefulSets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", p.Name), metav1.GetOptions{})
	require.NoError(t, err)

	err = wait.PollUntilContextTimeout(context.Background(), time.Second, 1*time.Minute, false, func(ctx context.Context) (bool, error) {
		_, err = framework.KubeClient.AppsV1().StatefulSets(ns).Get(ctx, fmt.Sprintf("prometheus-%s-shard-1", p.Name), metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return false, err
		}
		if err == nil {
			// StatefulSet still exists.
			return false, nil
		}
		// StatefulSet not found.
		return true, nil
	})
}

func testPromAlertmanagerDiscovery(t *testing.T) {
	for _, tc := range []struct {
		sdRole monitoringv1.ServiceDiscoveryRole
	}{
		{
			sdRole: monitoringv1.EndpointsRole,
		},
		{
			sdRole: monitoringv1.EndpointSliceRole,
		},
	} {
		t.Run(string(tc.sdRole), func(t *testing.T) {
			t.Parallel()
			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)
			ns := framework.CreateNamespace(context.Background(), t, testCtx)
			framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

			prometheusName := "test"
			alertmanagerName := "test"
			group := "servicediscovery-test"
			svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)
			amsvc := framework.MakeAlertmanagerService(alertmanagerName, group, v1.ServiceTypeClusterIP)

			p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
			framework.AddAlertingToPrometheus(p, ns, alertmanagerName)
			p.Spec.ServiceDiscoveryRole = ptr.To(tc.sdRole)
			_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
			require.NoError(t, err)

			finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
			require.NoError(t, err)

			testCtx.AddFinalizerFn(finalizerFn)

			s := framework.MakeBasicServiceMonitor(group)
			_, err = framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{})
			require.NoError(t, err)

			_, err = framework.KubeClient.CoreV1().Secrets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
			require.NoError(t, err)

			_, err = framework.CreateAlertmanagerAndWaitUntilReady(context.Background(), framework.MakeBasicAlertmanager(ns, alertmanagerName, 3))
			require.NoError(t, err)

			_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, amsvc)
			require.NoError(t, err)

			err = wait.PollUntilContextTimeout(context.Background(), time.Second, 5*time.Minute, false, isAlertmanagerDiscoveryWorking(ns, svc.Name, alertmanagerName))
			require.NoError(t, err)
		})
	}
}

func testPromExposingWithKubernetesAPI(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	basicPrometheus := framework.MakeBasicPrometheus(ns, "basic-prometheus", "test-group", 1)
	service := framework.MakePrometheusService(basicPrometheus.Name, "test-group", v1.ServiceTypeClusterIP)

	_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, basicPrometheus)
	require.NoError(t, err)

	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, service)
	require.NoError(t, err)

	ProxyGet := framework.KubeClient.CoreV1().Services(ns).ProxyGet
	request := ProxyGet("", service.Name, "web", "/metrics", make(map[string]string))
	_, err = request.DoRaw(context.Background())
	require.NoError(t, err)
}

func testPromDiscoverTargetPort(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	prometheusName := "test"
	group := "servicediscovery-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	targetPort := intstr.FromInt(9090)
	sm := &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name: prometheusName,
			Labels: map[string]string{
				"group": group,
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group": group,
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					TargetPort: &targetPort,
					Interval:   "5s",
				},
			},
		},
	}
	_, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), sm, metav1.CreateOptions{})
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
	require.NoError(t, err)

	testCtx.AddFinalizerFn(finalizerFn)

	_, err = framework.KubeClient.CoreV1().Secrets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
	require.NoError(t, err)

	err = framework.WaitForDiscoveryWorking(context.Background(), ns, svc.Name, prometheusName)
	require.NoError(t, err)
}

func testPromOpMatchPromAndServMonInDiffNSs(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	prometheusNSName := framework.CreateNamespace(context.Background(), t, testCtx)
	serviceMonitorNSName := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, prometheusNSName)

	err := framework.AddLabelsToNamespace(
		context.Background(),
		serviceMonitorNSName,
		map[string]string{"team": "frontend"},
	)
	require.NoError(t, err)

	group := "sample-app"

	prometheusJobName := serviceMonitorNSName + "/" + group

	prometheusName := "test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	s := framework.MakeBasicServiceMonitor(group)

	_, err = framework.MonClientV1.ServiceMonitors(serviceMonitorNSName).Create(context.Background(), s, metav1.CreateOptions{})
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(prometheusNSName, prometheusName, group, 1)
	p.Spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"team": "frontend",
		},
	}
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), prometheusNSName, p)
	require.NoError(t, err)

	finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), prometheusNSName, svc)
	require.NoError(t, err)

	testCtx.AddFinalizerFn(finalizerFn)

	resp, err := framework.PrometheusSVCGetRequest(context.Background(), prometheusNSName, svc.Name, "http", "/api/v1/status/config", map[string]string{})
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(string(resp), prometheusJobName), "expected Prometheus operator to configure Prometheus in ns '%v' to scrape the service monitor in ns '%v'", prometheusNSName, serviceMonitorNSName)
}

func testThanos(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	version := operator.DefaultThanosVersion

	prom := framework.MakeBasicPrometheus(ns, "basic-prometheus", "test-group", 1)
	prom.Spec.Replicas = proto.Int32(2)
	prom.Spec.Thanos = &monitoringv1.ThanosSpec{
		Version: &version,
	}
	_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prom)
	require.NoError(t, err)

	promSvc := framework.MakePrometheusService(prom.Name, "test-group", v1.ServiceTypeClusterIP)
	_, err = framework.KubeClient.CoreV1().Services(ns).Create(context.Background(), promSvc, metav1.CreateOptions{})
	require.NoError(t, err)

	svcMon := framework.MakeBasicServiceMonitor("test-group")
	_, err = framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), svcMon, metav1.CreateOptions{})
	require.NoError(t, err)

	qryDep, err := testFramework.MakeDeployment("../../example/thanos/query-deployment.yaml")
	require.NoError(t, err)
	// override image
	qryImage := "quay.io/thanos/thanos:" + version
	t.Log("setting up query with image: ", qryImage)
	qryDep.Spec.Template.Spec.Containers[0].Image = qryImage
	// override args
	qryArgs := []string{
		"query",
		"--log.level=debug",
		"--query.replica-label=prometheus_replica",
		fmt.Sprintf("--store=dnssrv+_grpc._tcp.prometheus-operated.%s.svc.cluster.local", ns),
	}
	t.Log("setting up query with args: ", qryArgs)
	qryDep.Spec.Template.Spec.Containers[0].Args = qryArgs
	err = framework.CreateDeployment(context.Background(), ns, qryDep)
	require.NoError(t, err)

	qrySvc := framework.MakeThanosQuerierService(qryDep.Name)
	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, qrySvc)
	require.NoError(t, err)

	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 5*time.Minute, false, func(ctx context.Context) (bool, error) {
		proxyGet := framework.KubeClient.CoreV1().Services(ns).ProxyGet
		request := proxyGet("http", qrySvc.Name, "http-query", "/api/v1/query", map[string]string{
			"query": "prometheus_build_info",
			"dedup": "false",
		})
		b, err := request.DoRaw(ctx)
		if err != nil {
			t.Logf("Error performing request against Thanos query: %v\n\nretrying...", err)
			return false, nil
		}

		d := struct {
			Data struct {
				Result []map[string]interface{} `json:"result"`
			} `json:"data"`
		}{}

		err = json.Unmarshal(b, &d)
		if err != nil {
			return false, err
		}

		result := len(d.Data.Result)
		// We're expecting 4 results as we are requesting the
		// `prometheus_build_info` metric, which is collected for both
		// Prometheus replicas by both replicas.
		expected := 4
		if result != expected {
			t.Logf("Unexpected number of results from query. Got %d, expected %d. retrying...", result, expected)
			return false, nil
		}
		return true, nil
	})
	require.NoError(t, err)
}

func testPromGetAuthSecret(t *testing.T) {
	t.Parallel()
	name := "test"

	tests := []struct {
		name           string
		secret         *v1.Secret
		serviceMonitor func() *monitoringv1.ServiceMonitor
	}{
		{
			name: "basic-auth",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Data: map[string][]byte{
					"user":     []byte("user"),
					"password": []byte("pass"),
				},
			},
			serviceMonitor: func() *monitoringv1.ServiceMonitor {
				sm := framework.MakeBasicServiceMonitor(name)
				sm.Spec.Endpoints[0].BasicAuth = &monitoringv1.BasicAuth{
					Username: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						},
						Key: "user",
					},
					Password: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						},
						Key: "password",
					},
				}

				return sm
			},
		},
		{
			name: "bearer-token",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Data: map[string][]byte{
					"bearertoken": []byte("abc"),
				},
			},
			serviceMonitor: func() *monitoringv1.ServiceMonitor {
				sm := framework.MakeBasicServiceMonitor(name)
				sm.Spec.Endpoints[0].BearerTokenSecret = &v1.SecretKeySelector{ //nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
					LocalObjectReference: v1.LocalObjectReference{
						Name: name,
					},
					Key: "bearertoken",
				}
				sm.Spec.Endpoints[0].Path = "/bearer-metrics"
				return sm
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)
			ns := framework.CreateNamespace(context.Background(), t, testCtx)
			framework.SetupPrometheusRBACGlobal(context.Background(), t, testCtx, ns)

			matchLabels := map[string]string{
				"tc": ns,
			}
			testNamespace := framework.CreateNamespace(context.Background(), t, testCtx)

			err := framework.AddLabelsToNamespace(context.Background(), testNamespace, matchLabels)
			require.NoError(t, err)

			simple, err := testFramework.MakeDeployment("../../test/framework/resources/basic-auth-app-deployment.yaml")
			require.NoError(t, err)

			err = framework.CreateDeployment(context.Background(), testNamespace, simple)
			require.NoError(t, err)

			authSecret := test.secret
			_, err = framework.KubeClient.CoreV1().Secrets(testNamespace).Create(context.Background(), authSecret, metav1.CreateOptions{})
			require.NoError(t, err)

			svc := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						"group": name,
					},
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
					Ports: []v1.ServicePort{
						{
							Name: "web",
							Port: 8080,
						},
					},
					Selector: map[string]string{
						"group": name,
					},
				},
			}

			sm := test.serviceMonitor()
			finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), testNamespace, svc)
			require.NoError(t, err)

			testCtx.AddFinalizerFn(finalizerFn)

			_, err = framework.MonClientV1.ServiceMonitors(testNamespace).Create(context.Background(), sm, metav1.CreateOptions{})
			require.NoError(t, err)

			prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)
			prometheusCRD.Spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
				MatchLabels: matchLabels,
			}
			prometheusCRD.Spec.ScrapeInterval = "1s"
			_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD)
			require.NoError(t, err)

			err = framework.WaitForHealthyTargets(context.Background(), ns, "prometheus-operated", 1)
			require.NoError(t, err)
		})
	}
}

// testOperatorNSScope tests the multi namespace feature of the Prometheus Operator.
// It checks whether it ignores rules that are not in the watched namespaces of the
// Prometheus Operator. The Prometheus Operator internally treats watching a
// single namespace different than watching multiple namespaces, hence the two
// sub-tests.
func testOperatorNSScope(t *testing.T) {
	name := "test"
	firtAlertName := "firstAlert"
	secondAlertName := "secondAlert"

	t.Run("SingleNS", func(t *testing.T) {

		testCtx := framework.NewTestCtx(t)
		defer testCtx.Cleanup(t)

		operatorNS := framework.CreateNamespace(context.Background(), t, testCtx)
		mainNS := framework.CreateNamespace(context.Background(), t, testCtx)
		arbitraryNS := framework.CreateNamespace(context.Background(), t, testCtx)

		framework.SetupPrometheusRBAC(context.Background(), t, testCtx, mainNS)

		prometheusNamespaceSelector := map[string]string{"prometheus": mainNS}

		// Add labels to namespaces for Prometheus RuleNamespaceSelector.
		for _, ns := range []string{mainNS, arbitraryNS} {
			err := framework.AddLabelsToNamespace(context.Background(), ns, prometheusNamespaceSelector)
			require.NoError(t, err)
		}

		// Prometheus Operator only watches single namespace mainNS, not arbitraryNS.
		_, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), operatorNS, []string{mainNS}, nil, nil, nil, false, true, true)
		require.NoError(t, err)

		ruleDef := []struct {
			NSName    string
			AlertName string
		}{{arbitraryNS, secondAlertName}, {mainNS, firtAlertName}}

		for _, r := range ruleDef {
			_, err := framework.MakeAndCreateFiringRule(context.Background(), r.NSName, name, r.AlertName)
			require.NoError(t, err)
		}

		p := framework.MakeBasicPrometheus(mainNS, name, name, 1)
		p.Spec.RuleNamespaceSelector = &metav1.LabelSelector{
			MatchLabels: prometheusNamespaceSelector,
		}
		p.Spec.EvaluationInterval = "1s"
		p, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), mainNS, p)
		require.NoError(t, err)

		pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
		finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), mainNS, pSVC)
		require.NoError(t, err)

		testCtx.AddFinalizerFn(finalizerFn)

		err = framework.WaitForPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, firtAlertName)
		require.NoError(t, err)

		firing, err := framework.CheckPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, secondAlertName)
		if err != nil && !strings.Contains(err.Error(), "expected 1 query result but got 0") {
			t.Fatal(err)
		}

		require.False(t, firing, "expected alert %q not to fire", secondAlertName)
	})

	t.Run("MultiNS", func(t *testing.T) {

		testCtx := framework.NewTestCtx(t)
		defer testCtx.Cleanup(t)

		operatorNS := framework.CreateNamespace(context.Background(), t, testCtx)
		prometheusNS := framework.CreateNamespace(context.Background(), t, testCtx)
		ruleNS := framework.CreateNamespace(context.Background(), t, testCtx)
		arbitraryNS := framework.CreateNamespace(context.Background(), t, testCtx)

		framework.SetupPrometheusRBAC(context.Background(), t, testCtx, prometheusNS)

		prometheusNamespaceSelector := map[string]string{"prometheus": prometheusNS}

		for _, ns := range []string{ruleNS, arbitraryNS} {
			err := framework.AddLabelsToNamespace(context.Background(), ns, prometheusNamespaceSelector)
			require.NoError(t, err)
		}

		// Prometheus Operator only watches prometheusNS and ruleNS, not arbitraryNS.
		_, err := framework.CreateOrUpdatePrometheusOperator(context.Background(), operatorNS, []string{prometheusNS, ruleNS}, nil, nil, nil, false, true, true)
		require.NoError(t, err)

		ruleDef := []struct {
			NSName    string
			AlertName string
		}{{arbitraryNS, secondAlertName}, {ruleNS, firtAlertName}}

		for _, r := range ruleDef {
			_, err := framework.MakeAndCreateFiringRule(context.Background(), r.NSName, name, r.AlertName)
			require.NoError(t, err)
		}

		p := framework.MakeBasicPrometheus(prometheusNS, name, name, 1)
		p.Spec.RuleNamespaceSelector = &metav1.LabelSelector{
			MatchLabels: prometheusNamespaceSelector,
		}
		p.Spec.EvaluationInterval = "1s"
		p, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), prometheusNS, p)
		require.NoError(t, err)

		pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
		finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), prometheusNS, pSVC)
		require.NoError(t, err)

		testCtx.AddFinalizerFn(finalizerFn)

		err = framework.WaitForPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, firtAlertName)
		require.NoError(t, err)

		firing, err := framework.CheckPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, secondAlertName)
		if err != nil && !strings.Contains(err.Error(), "expected 1 query result but got 0") {
			t.Fatal(err)
		}

		require.False(t, firing, "expected alert %q not to fire", secondAlertName)
	})
}

// testPromArbitraryFSAcc tests the
// github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.PrometheusSpec.ArbitraryFSAccessThroughSMs
// configuration with the service monitor bearer token and tls assets option.
func testPromArbitraryFSAcc(t *testing.T) {
	t.Parallel()
	name := "test"

	tests := []struct {
		name                              string
		arbitraryFSAccessThroughSMsConfig monitoringv1.ArbitraryFSAccessThroughSMsConfig
		endpoint                          monitoringv1.Endpoint
		expectTargets                     bool
	}{
		//
		// Bearer tokens:
		//
		{
			name: "allowed-bearer-file",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: false,
			},
			endpoint: monitoringv1.Endpoint{
				Port:            "web",
				BearerTokenFile: "/etc/ca-certificates/bearer-token",
			},
			expectTargets: true,
		},
		{
			name: "denied-bearer-file",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: true,
			},
			endpoint: monitoringv1.Endpoint{
				Port:            "web",
				BearerTokenFile: "/etc/ca-certificates/bearer-token",
			},
			expectTargets: false,
		},
		{
			name: "denied-bearer-secret",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: true,
			},
			endpoint: monitoringv1.Endpoint{
				Port: "web",
				BearerTokenSecret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: name,
					},
					Key: "bearer-token",
				},
			},
			expectTargets: true,
		},
		//
		// TLS assets:
		//
		{
			name: "allowed-tls-file",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: false,
			},
			endpoint: monitoringv1.Endpoint{
				Port: "web",
				TLSConfig: &monitoringv1.TLSConfig{
					CAFile:   "/etc/ca-certificates/cert.pem",
					CertFile: "/etc/ca-certificates/cert.pem",
					KeyFile:  "/etc/ca-certificates/key.pem",
				},
			},
			expectTargets: true,
		},
		{
			name: "denied-tls-file",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: true,
			},
			endpoint: monitoringv1.Endpoint{
				Port: "web",
				TLSConfig: &monitoringv1.TLSConfig{
					CAFile:   "/etc/ca-certificates/cert.pem",
					CertFile: "/etc/ca-certificates/cert.pem",
					KeyFile:  "/etc/ca-certificates/key.pem",
				},
			},
			expectTargets: false,
		},
		{
			name: "denied-tls-secret",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: true,
			},
			endpoint: monitoringv1.Endpoint{
				Port: "web",
				TLSConfig: &monitoringv1.TLSConfig{
					SafeTLSConfig: monitoringv1.SafeTLSConfig{
						InsecureSkipVerify: ptr.To(true),
						CA: monitoringv1.SecretOrConfigMap{
							Secret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: name,
								},
								Key: "cert.pem",
							},
						},
						Cert: monitoringv1.SecretOrConfigMap{
							Secret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: name,
								},
								Key: "cert.pem",
							},
						},
						KeySecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: name,
							},
							Key: "key.pem",
						},
					},
				},
			},
			expectTargets: true,
		},
		{
			name: "denied-tls-configmap",
			arbitraryFSAccessThroughSMsConfig: monitoringv1.ArbitraryFSAccessThroughSMsConfig{
				Deny: true,
			},
			endpoint: monitoringv1.Endpoint{
				Port: "web",
				TLSConfig: &monitoringv1.TLSConfig{
					SafeTLSConfig: monitoringv1.SafeTLSConfig{
						InsecureSkipVerify: ptr.To(true),
						CA: monitoringv1.SecretOrConfigMap{
							ConfigMap: &v1.ConfigMapKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: name,
								},
								Key: "cert.pem",
							},
						},
						Cert: monitoringv1.SecretOrConfigMap{
							ConfigMap: &v1.ConfigMapKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: name,
								},
								Key: "cert.pem",
							},
						},
						KeySecret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: name,
							},
							Key: "key.pem",
						},
					},
				},
			},
			expectTargets: true,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)

			ns := framework.CreateNamespace(context.Background(), t, testCtx)
			framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

			// Create secret either used by bearer token secret key ref, tls
			// asset key ref or tls configmap key ref.
			cert, err := os.ReadFile("../../test/instrumented-sample-app/certs/cert.pem")
			require.NoError(t, err)

			key, err := os.ReadFile("../../test/instrumented-sample-app/certs/key.pem")
			require.NoError(t, err)

			tlsCertsSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Data: map[string][]byte{
					"cert.pem":     cert,
					"key.pem":      key,
					"bearer-token": []byte("abc"),
				},
			}

			_, err = framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), tlsCertsSecret, metav1.CreateOptions{})
			require.NoError(t, err)

			tlsCertsConfigMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Data: map[string]string{
					"cert.pem": string(cert),
				},
			}

			_, err = framework.KubeClient.CoreV1().ConfigMaps(ns).Create(context.Background(), tlsCertsConfigMap, metav1.CreateOptions{})
			require.NoError(t, err)

			s := framework.MakeBasicServiceMonitor(name)
			s.Spec.Endpoints[0] = test.endpoint
			_, err = framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{})
			require.NoError(t, err)

			prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)
			prometheusCRD.Namespace = ns
			prometheusCRD.Spec.ArbitraryFSAccessThroughSMs = test.arbitraryFSAccessThroughSMsConfig

			if strings.HasSuffix(test.name, "-file") {
				mountTLSFiles(prometheusCRD, name)
			}

			_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD)
			require.NoError(t, err)

			svc := framework.MakePrometheusService(prometheusCRD.Name, name, v1.ServiceTypeClusterIP)
			_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
			require.NoError(t, err)

			if test.expectTargets {
				err := framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 1)
				require.NoError(t, err)

				return
			}

			// Make sure Prometheus has enough time to reload.
			time.Sleep(2 * time.Minute)
			err = framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 0)
			require.NoError(t, err)
		})
	}

}

// mountTLSFiles is a helper to manually mount TLS certificate files
// into the prometheus container.
func mountTLSFiles(p *monitoringv1.Prometheus, secretName string) {
	volumeName := secretName
	p.Spec.Volumes = append(p.Spec.Volumes,
		v1.Volume{
			Name: volumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		})
	p.Spec.Containers = []v1.Container{
		{
			Name: "prometheus",
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      volumeName,
					MountPath: "/etc/ca-certificates",
				},
			},
		},
	}
}

// testPromTLSConfigViaSecret tests the service monitor endpoint option to load
// certificate assets via Kubernetes secrets into the Prometheus container.
func testPromTLSConfigViaSecret(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
	name := "test"

	//
	// Setup sample app.
	//

	cert, err := os.ReadFile("../../test/instrumented-sample-app/certs/cert.pem")
	require.NoError(t, err)

	key, err := os.ReadFile("../../test/instrumented-sample-app/certs/key.pem")
	require.NoError(t, err)

	tlsCertsSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string][]byte{
			"cert.pem": cert,
			"key.pem":  key,
		},
	}

	_, err = framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), tlsCertsSecret, metav1.CreateOptions{})
	require.NoError(t, err)

	simple, err := testFramework.MakeDeployment("../../test/framework/resources/basic-auth-app-deployment.yaml")
	require.NoError(t, err)

	simple.Spec.Template.Spec.Containers[0].Args = []string{"--cert-path=/etc/certs"}

	simple.Spec.Template.Spec.Volumes = []v1.Volume{
		{
			Name: "tls-certs",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: tlsCertsSecret.Name,
				},
			},
		},
	}

	simple.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
		{
			Name:      simple.Spec.Template.Spec.Volumes[0].Name,
			MountPath: "/etc/certs",
		},
	}

	err = framework.CreateDeployment(context.Background(), ns, simple)
	require.NoError(t, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"group": name,
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				{
					Name: "web",
					Port: 8080,
				},
				{
					Name: "mtls",
					Port: 8081,
				},
			},
			Selector: map[string]string{
				"group": name,
			},
		},
	}

	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
	require.NoError(t, err)

	//
	// Setup monitoring.
	//

	sm := framework.MakeBasicServiceMonitor(name)
	sm.Spec.Endpoints = []monitoringv1.Endpoint{
		{
			Port:     "mtls",
			Interval: "30s",
			Scheme:   "https",
			TLSConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					InsecureSkipVerify: ptr.To(true),
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: tlsCertsSecret.Name,
							},
							Key: "cert.pem",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: tlsCertsSecret.Name,
						},
						Key: "key.pem",
					},
				},
			},
		},
	}

	_, err = framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), sm, metav1.CreateOptions{})
	require.NoError(t, err)

	prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)

	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD)
	require.NoError(t, err)

	promSVC := framework.MakePrometheusService(prometheusCRD.Name, name, v1.ServiceTypeClusterIP)

	_, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, promSVC)
	require.NoError(t, err)

	//
	// Check for proper scraping.
	//

	err = framework.WaitForHealthyTargets(context.Background(), ns, promSVC.Name, 1)
	require.NoError(t, err)

	// TODO: Do a poll instead, should speed up things.
	time.Sleep(30 * time.Second)

	response, err := framework.PrometheusSVCGetRequest(context.Background(), ns, promSVC.Name, "http", "/api/v1/query", map[string]string{"query": fmt.Sprintf(`up{job="%v",endpoint="%v"}`, name, sm.Spec.Endpoints[0].Port)})
	require.NoError(t, err)

	q := testFramework.PrometheusQueryAPIResponse{}
	err = json.NewDecoder(bytes.NewBuffer(response)).Decode(&q)
	require.NoError(t, err)

	require.Equal(t, "success", q.Status)

	require.Equal(t, "1", q.Data.Result[0].Value[1])
}

func testPromStaticProbe(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	blackboxExporterName := "blackbox-exporter"
	err := framework.CreateBlackBoxExporterAndWaitUntilReady(context.Background(), ns, blackboxExporterName)
	require.NoError(t, err)

	blackboxSvc := framework.MakeBlackBoxExporterService(ns, blackboxExporterName)
	finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, blackboxSvc)
	require.NoError(t, err)

	testCtx.AddFinalizerFn(finalizerFn)

	prometheusName := "test"
	group := "probe-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	proberURL := blackboxExporterName + ":9115"
	targets := []string{svc.Name + ":9090"}

	probe := framework.MakeBasicStaticProbe(group, proberURL, targets)
	_, err = framework.MonClientV1.Probes(ns).Create(context.Background(), probe, metav1.CreateOptions{})
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	p.Spec.ProbeSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": group,
		},
	}
	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	finalizerFn, err = framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
	require.NoError(t, err)

	testCtx.AddFinalizerFn(finalizerFn)

	expectedURL := url.URL{Host: proberURL, Scheme: "http", Path: "/probe"}
	q := expectedURL.Query()
	q.Set("module", "http_2xx")
	q.Set("target", targets[0])
	expectedURL.RawQuery = q.Encode()

	err = wait.PollUntilContextTimeout(context.Background(), time.Second, time.Minute*5, false, func(ctx context.Context) (bool, error) {
		activeTargets, err := framework.GetActiveTargets(ctx, ns, svc.Name)
		if err != nil {
			return false, err
		}

		if len(activeTargets) != 1 {
			return false, nil
		}

		exp := expectedURL.String()
		if activeTargets[0].ScrapeURL != exp {
			return false, nil
		}

		if value, ok := activeTargets[0].Labels["instance"]; !ok || value != targets[0] {
			return false, nil
		}

		return true, nil
	})
	require.NoError(t, err)
}

func testPromSecurePodMonitor(t *testing.T) {
	t.Parallel()
	name := "test"

	tests := []struct {
		name     string
		endpoint monitoringv1.PodMetricsEndpoint
	}{
		//
		// Basic auth:
		//
		{
			name: "basic-auth-secret",
			endpoint: monitoringv1.PodMetricsEndpoint{
				Port: "web",
				BasicAuth: &monitoringv1.BasicAuth{
					Username: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						},
						Key: "user",
					},
					Password: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						},
						Key: "password",
					},
				},
			},
		},
		//
		// Bearer tokens:
		//
		{
			name: "bearer-secret",
			endpoint: monitoringv1.PodMetricsEndpoint{
				Port: "web",
				BearerTokenSecret: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: name,
					},
					Key: "bearer-token",
				},
				Path: "/bearer-metrics",
			},
		},
		//
		// TLS assets:
		//
		{
			name: "tls-secret",
			endpoint: monitoringv1.PodMetricsEndpoint{
				Port:   "mtls",
				Scheme: "https",
				TLSConfig: &monitoringv1.SafeTLSConfig{
					InsecureSkipVerify: ptr.To(true),
					CA: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: name,
							},
							Key: "cert.pem",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: name,
							},
							Key: "cert.pem",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						},
						Key: "key.pem",
					},
				},
				Path: "/",
			},
		},
		{
			name: "tls-configmap",
			endpoint: monitoringv1.PodMetricsEndpoint{
				Port:   "mtls",
				Scheme: "https",
				TLSConfig: &monitoringv1.SafeTLSConfig{
					InsecureSkipVerify: ptr.To(true),
					CA: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: name,
							},
							Key: "cert.pem",
						},
					},
					Cert: monitoringv1.SecretOrConfigMap{
						ConfigMap: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: name,
							},
							Key: "cert.pem",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						},
						Key: "key.pem",
					},
				},
				Path: "/",
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)
			ns := framework.CreateNamespace(context.Background(), t, testCtx)
			framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

			// Create secret either used by bearer token secret key ref, tls
			// asset key ref or tls configmap key ref.
			cert, err := os.ReadFile("../../test/instrumented-sample-app/certs/cert.pem")
			require.NoError(t, err)

			key, err := os.ReadFile("../../test/instrumented-sample-app/certs/key.pem")
			require.NoError(t, err)

			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Data: map[string][]byte{
					"user":         []byte("user"),
					"password":     []byte("pass"),
					"bearer-token": []byte("abc"),
					"cert.pem":     cert,
					"key.pem":      key,
				},
			}

			_, err = framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), secret, metav1.CreateOptions{})
			require.NoError(t, err)

			tlsCertsConfigMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Data: map[string]string{
					"cert.pem": string(cert),
				},
			}

			_, err = framework.KubeClient.CoreV1().ConfigMaps(ns).Create(context.Background(), tlsCertsConfigMap, metav1.CreateOptions{})
			require.NoError(t, err)

			simple, err := testFramework.MakeDeployment("../../test/framework/resources/basic-auth-app-deployment.yaml")
			require.NoError(t, err)

			simple.Spec.Template.Spec.Volumes = []v1.Volume{
				{
					Name: name,
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: name,
						},
					},
				},
			}

			simple.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
				{
					Name:      name,
					MountPath: "/etc/ca-certificates",
				},
			}

			if test.endpoint.Port == "mtls" {
				simple.Spec.Template.Spec.Containers[0].Args = []string{"--cert-path=/etc/ca-certificates"}
			}

			err = framework.CreateDeployment(context.Background(), ns, simple)
			require.NoError(t, err)

			pm := framework.MakeBasicPodMonitor(name)
			pm.Spec.PodMetricsEndpoints = []monitoringv1.PodMetricsEndpoint{test.endpoint}

			_, err = framework.MonClientV1.PodMonitors(ns).Create(context.Background(), pm, metav1.CreateOptions{})
			require.NoError(t, err)

			prom := framework.MakeBasicPrometheus(ns, name, name, 1)
			prom.Spec.ScrapeInterval = "1s"
			_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prom)
			require.NoError(t, err)

			err = framework.WaitForHealthyTargets(context.Background(), ns, "prometheus-operated", 1)
			require.NoError(t, err)
		})
	}
}

func testPromWebWithThanosSidecar(t *testing.T) {
	t.Parallel()
	trueVal := true
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	host := fmt.Sprintf("%s.%s.svc", "basic-prometheus", ns)
	certBytes, keyBytes, err := certutil.GenerateSelfSignedCertKey(host, nil, nil)
	require.NoError(t, err)

	kubeClient := framework.KubeClient
	err = framework.CreateOrUpdateSecretWithCert(context.Background(), certBytes, keyBytes, ns, "web-tls")
	require.NoError(t, err)

	version := operator.DefaultThanosVersion
	prom := framework.MakeBasicPrometheus(ns, "basic-prometheus", "test-group", 1)
	prom.Spec.Web = &monitoringv1.PrometheusWebSpec{
		WebConfigFileFields: monitoringv1.WebConfigFileFields{
			TLSConfig: &monitoringv1.WebTLSConfig{
				KeySecret: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "web-tls",
					},
					Key: "tls.key",
				},
				Cert: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "web-tls",
						},
						Key: "tls.crt",
					},
				},
			},
			HTTPConfig: &monitoringv1.WebHTTPConfig{
				HTTP2: &trueVal,
				Headers: &monitoringv1.WebHTTPHeaders{
					ContentSecurityPolicy:   "default-src 'self'",
					XFrameOptions:           "Deny",
					XContentTypeOptions:     "NoSniff",
					XXSSProtection:          "1; mode=block",
					StrictTransportSecurity: "max-age=31536000; includeSubDomains",
				},
			},
		},
	}
	prom.Spec.Thanos = &monitoringv1.ThanosSpec{
		Version: &version,
	}

	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prom)
	require.NoError(t, err)

	var pollErr error
	err = wait.PollUntilContextTimeout(context.Background(), time.Second, time.Minute, false, func(ctx context.Context) (bool, error) {
		promPods, err := kubeClient.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			pollErr = err
			return false, nil
		}

		if len(promPods.Items) == 0 {
			pollErr = fmt.Errorf("No prometheus pods found in namespace %s", ns)
			return false, nil
		}

		cfg := framework.RestConfig
		podName := promPods.Items[0].Name

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		closer, err := testFramework.StartPortForward(ctx, cfg, "https", podName, ns, "9090")
		if err != nil {
			pollErr = fmt.Errorf("failed to start port forwarding: %v", err)
			t.Log(pollErr)
			return false, nil
		}
		defer closer()

		req, err := http.NewRequestWithContext(ctx, "GET", "https://localhost:9090", nil)
		if err != nil {
			pollErr = err
			return false, nil
		}

		// The prometheus certificate is issued to <pod>.<namespace>.svc,
		// but port-forwarding is done through localhost.
		// This is why we use an http client which skips the TLS verification.
		// In the test we will verify the TLS certificate manually to make sure
		// the prometheus instance is configured properly.
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		err = http2.ConfigureTransport(transport)
		if err != nil {
			pollErr = err
			return false, nil
		}

		httpClient := http.Client{
			Transport: transport,
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			pollErr = err
			return false, nil
		}

		if resp.ProtoMajor != 2 {
			pollErr = fmt.Errorf("expected ProtoMajor to be 2 but got %d", resp.ProtoMajor)
			return false, nil
		}

		receivedCertBytes, err := certutil.EncodeCertificates(resp.TLS.PeerCertificates...)
		if err != nil {
			pollErr = err
			return false, nil
		}

		if !bytes.Equal(receivedCertBytes, certBytes) {
			pollErr = fmt.Errorf("certificate received from prometheus instance does not match the one which is configured")
			return false, nil
		}

		expectedHeaders := map[string]string{
			"Content-Security-Policy":   "default-src 'self'",
			"X-Frame-Options":           "deny",
			"X-Content-Type-Options":    "nosniff",
			"X-XSS-Protection":          "1; mode=block",
			"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		}

		for k, v := range expectedHeaders {
			rv := resp.Header.Get(k)

			if rv != v {
				pollErr = fmt.Errorf("expected header %s value to be %s but got %s", k, v, rv)
				return false, nil
			}
		}

		reloadSuccessTimestamp, err := framework.GetMetricVal(context.Background(), "https", ns, podName, "8080", "reloader_last_reload_success_timestamp_seconds")
		if err != nil {
			pollErr = err
			return false, nil
		}

		if reloadSuccessTimestamp == 0 {
			pollErr = fmt.Errorf("config reloader failed to reload once")
			return false, nil
		}

		thanosSidecarPrometheusUp, err := framework.GetMetricVal(context.Background(), "http", ns, podName, "10902", "thanos_sidecar_prometheus_up")
		if err != nil {
			pollErr = err
			return false, nil
		}

		if thanosSidecarPrometheusUp == 0 {
			pollErr = fmt.Errorf("thanos sidecar failed to connect prometheus")
			return false, nil
		}

		return true, nil
	})

	require.NoError(t, err, "poll function execution error: %v: %v", err, pollErr)

	// Simulate a certificate renewal and check that the new certificate is in place
	certBytesNew, keyBytesNew, err := certutil.GenerateSelfSignedCertKey(host, nil, nil)
	require.NoError(t, err)

	err = framework.CreateOrUpdateSecretWithCert(context.Background(), certBytesNew, keyBytesNew, ns, "web-tls")
	require.NoError(t, err)

	err = wait.PollUntilContextTimeout(context.Background(), time.Second, 2*time.Minute, false, func(ctx context.Context) (bool, error) {
		promPods, err := kubeClient.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			pollErr = err
			return false, nil
		}

		if len(promPods.Items) == 0 {
			pollErr = fmt.Errorf("No prometheus pods found in namespace %s", ns)
			return false, nil
		}

		cfg := framework.RestConfig
		podName := promPods.Items[0].Name

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		closer, err := testFramework.StartPortForward(ctx, cfg, "https", podName, ns, "9090")
		if err != nil {
			pollErr = fmt.Errorf("failed to start port forwarding: %v", err)
			t.Log(pollErr)
			return false, nil
		}
		defer closer()

		req, err := http.NewRequestWithContext(ctx, "GET", "https://localhost:9090", nil)
		if err != nil {
			pollErr = err
			return false, nil
		}

		// The prometheus certificate is issued to <pod>.<namespace>.svc,
		// but port-forwarding is done through localhost.
		// This is why we use an http client which skips the TLS verification.
		// In the test we will verify the TLS certificate manually to make sure
		// the prometheus instance is configured properly.
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		err = http2.ConfigureTransport(transport)
		if err != nil {
			pollErr = err
			return false, nil
		}

		httpClient := http.Client{
			Transport: transport,
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			pollErr = err
			return false, nil
		}

		if resp.ProtoMajor != 2 {
			pollErr = fmt.Errorf("expected ProtoMajor to be 2 but got %d", resp.ProtoMajor)
			return false, nil
		}

		receivedCertBytes, err := certutil.EncodeCertificates(resp.TLS.PeerCertificates...)
		if err != nil {
			pollErr = err
			return false, nil
		}

		if !bytes.Equal(receivedCertBytes, certBytesNew) {
			pollErr = fmt.Errorf("certificate received from prometheus instance does not match the one which is configured after certificate renewal")
			return false, nil
		}

		return true, nil
	})

	require.NoError(t, err, "poll function execution error: %v: %v", err, pollErr)
}

func testPromMinReadySeconds(t *testing.T) {
	t.Parallel()

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	kubeClient := framework.KubeClient

	var setMinReadySecondsInitial uint32 = 5
	prom := framework.MakeBasicPrometheus(ns, "basic-prometheus", "test-group", 1)
	prom.Spec.MinReadySeconds = &setMinReadySecondsInitial

	prom, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prom)
	require.NoError(t, err)

	promSS, err := kubeClient.AppsV1().StatefulSets(ns).Get(context.Background(), "prometheus-basic-prometheus", metav1.GetOptions{})
	require.NoError(t, err)

	require.Equal(t, promSS.Spec.MinReadySeconds, int32(setMinReadySecondsInitial))

	var updated uint32 = 10
	var got int32
	_, err = framework.PatchPrometheusAndWaitUntilReady(
		context.Background(),
		prom.Name,
		ns,
		monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				MinReadySeconds: &updated,
			},
		},
	)
	require.NoError(t, err)

	err = wait.PollUntilContextTimeout(context.Background(), time.Second, time.Minute*5, false, func(ctx context.Context) (bool, error) {
		promSS, err := kubeClient.AppsV1().StatefulSets(ns).Get(ctx, "prometheus-basic-prometheus", metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if promSS.Spec.MinReadySeconds != int32(updated) {
			got = promSS.Spec.MinReadySeconds
			return false, nil
		}
		return true, nil
	})

	require.NoError(t, err, "expected MinReadySeconds to be %d but got %d", updated, got)
}

// testPromEnforcedNamespaceLabel checks that the enforcedNamespaceLabel field
// is honored even if a user tries to bypass the enforcement.
func testPromEnforcedNamespaceLabel(t *testing.T) {
	t.Parallel()

	for i, tc := range []struct {
		relabelConfigs       []monitoringv1.RelabelConfig
		metricRelabelConfigs []monitoringv1.RelabelConfig
	}{
		{
			// override label using the labeldrop action.
			relabelConfigs: []monitoringv1.RelabelConfig{
				{
					Regex:  "namespace",
					Action: "labeldrop",
				},
			},
			metricRelabelConfigs: []monitoringv1.RelabelConfig{
				{
					Regex:  "namespace",
					Action: "labeldrop",
				},
			},
		},
		{
			// override label using the replace action.
			relabelConfigs: []monitoringv1.RelabelConfig{
				{
					TargetLabel: "namespace",
					Replacement: ptr.To("ns1"),
				},
			},
			metricRelabelConfigs: []monitoringv1.RelabelConfig{
				{
					TargetLabel: "namespace",
					Replacement: ptr.To("ns1"),
				},
			},
		},
		{
			// override label using the replace action with empty replacement.
			relabelConfigs: []monitoringv1.RelabelConfig{
				{
					TargetLabel: "namespace",
					Replacement: ptr.To(""),
				},
			},
			metricRelabelConfigs: []monitoringv1.RelabelConfig{
				{
					TargetLabel: "namespace",
					Replacement: ptr.To(""),
				},
			},
		},
		{
			// override label using the labelmap action.
			relabelConfigs: []monitoringv1.RelabelConfig{
				{
					TargetLabel: "temp_namespace",
					Replacement: ptr.To("ns1"),
				},
			},
			metricRelabelConfigs: []monitoringv1.RelabelConfig{
				{
					Action:      "labelmap",
					Regex:       "temp_namespace",
					Replacement: ptr.To("namespace"),
				},
				{
					Action: "labeldrop",
					Regex:  "temp_namespace",
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ctx := framework.NewTestCtx(t)
			defer ctx.Cleanup(t)
			ns := framework.CreateNamespace(context.Background(), t, ctx)
			framework.SetupPrometheusRBAC(context.Background(), t, ctx, ns)

			prometheusName := "test"
			group := "servicediscovery-test"
			svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

			s := framework.MakeBasicServiceMonitor(group)
			s.Spec.Endpoints[0].RelabelConfigs = tc.relabelConfigs
			s.Spec.Endpoints[0].MetricRelabelConfigs = tc.metricRelabelConfigs
			_, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{})
			require.NoError(t, err)

			p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
			p.Spec.EnforcedNamespaceLabel = "namespace"
			_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
			require.NoError(t, err)

			finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
			require.NoError(t, err)

			ctx.AddFinalizerFn(finalizerFn)

			_, err = framework.KubeClient.CoreV1().Secrets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
			require.NoError(t, err)

			err = framework.WaitForDiscoveryWorking(context.Background(), ns, svc.Name, prometheusName)
			require.NoError(t, err)

			// Check that the namespace label is enforced to the correct value.
			var (
				loopErr        error
				namespaceLabel string
			)

			err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 1*time.Minute, false, func(_ context.Context) (bool, error) {
				loopErr = nil
				res, err := framework.PrometheusQuery(ns, svc.Name, "http", "prometheus_build_info")
				if err != nil {
					loopErr = fmt.Errorf("failed to query Prometheus: %w", err)
					return false, nil
				}

				if len(res) != 1 {
					loopErr = fmt.Errorf("expecting 1 item but got %d", len(res))
					return false, nil
				}

				for k, v := range res[0].Metric {
					if k == "namespace" {
						namespaceLabel = v
						return true, nil
					}
				}

				loopErr = fmt.Errorf("expecting to find 'namespace' label in %v", res[0].Metric)
				return false, nil
			})

			require.NoError(t, err, "%v:%v", err, loopErr)
			require.Equal(t, namespaceLabel, ns)
		})
	}
}

// testPromNamespaceEnforcementExclusion checks that the enforcedNamespaceLabel field
// is not enforced on objects defined in ExcludedFromEnforcement.
func testPromNamespaceEnforcementExclusion(t *testing.T) {
	t.Parallel()

	for i, tc := range []struct {
		relabelConfigs       []monitoringv1.RelabelConfig
		metricRelabelConfigs []monitoringv1.RelabelConfig
		expectedNamespace    string
	}{
		{
			// override label using the labeldrop action.
			relabelConfigs: []monitoringv1.RelabelConfig{
				{
					Regex:  "namespace",
					Action: "labeldrop",
				},
			},
			metricRelabelConfigs: []monitoringv1.RelabelConfig{
				{
					Regex:  "namespace",
					Action: "labeldrop",
				},
			},
			expectedNamespace: "",
		},
		{
			// override label using the replace action.
			relabelConfigs: []monitoringv1.RelabelConfig{
				{
					TargetLabel: "namespace",
					Replacement: ptr.To("ns1"),
				},
			},
			metricRelabelConfigs: []monitoringv1.RelabelConfig{
				{
					TargetLabel: "namespace",
					Replacement: ptr.To("ns1"),
				},
			},
			expectedNamespace: "ns1",
		},
		{
			// override label using the labelmap action.
			relabelConfigs: []monitoringv1.RelabelConfig{
				{
					TargetLabel: "temp_namespace",
					Replacement: ptr.To("ns1"),
				},
			},
			metricRelabelConfigs: []monitoringv1.RelabelConfig{
				{
					Action:      "labelmap",
					Regex:       "temp_namespace",
					Replacement: ptr.To("namespace"),
				},
				{
					Action: "labeldrop",
					Regex:  "temp_namespace",
				},
			},
			expectedNamespace: "ns1",
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ctx := framework.NewTestCtx(t)
			defer ctx.Cleanup(t)
			ns := framework.CreateNamespace(context.Background(), t, ctx)
			framework.SetupPrometheusRBAC(context.Background(), t, ctx, ns)

			prometheusName := "test"
			group := "servicediscovery-test"
			svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

			s := framework.MakeBasicServiceMonitor(group)
			s.Spec.Endpoints[0].RelabelConfigs = tc.relabelConfigs
			s.Spec.Endpoints[0].MetricRelabelConfigs = tc.metricRelabelConfigs
			_, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{})
			require.NoError(t, err)

			p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
			p.Spec.EnforcedNamespaceLabel = "namespace"
			p.Spec.ExcludedFromEnforcement = []monitoringv1.ObjectReference{
				{
					Namespace: ns,
					Group:     "monitoring.coreos.com",
					Resource:  monitoringv1.ServiceMonitorName,
				},
			}
			_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
			require.NoError(t, err)

			finalizerFn, err := framework.CreateOrUpdateServiceAndWaitUntilReady(context.Background(), ns, svc)
			require.NoError(t, err)

			ctx.AddFinalizerFn(finalizerFn)

			_, err = framework.KubeClient.CoreV1().Secrets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
			require.NoError(t, err)

			err = framework.WaitForDiscoveryWorking(context.Background(), ns, svc.Name, prometheusName)
			require.NoError(t, err)

			// Check that the namespace label isn't enforced.
			var (
				loopErr        error
				namespaceLabel string
			)

			err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 1*time.Minute, false, func(_ context.Context) (bool, error) {
				loopErr = nil
				res, err := framework.PrometheusQuery(ns, svc.Name, "http", "prometheus_build_info")
				if err != nil {
					loopErr = fmt.Errorf("failed to query Prometheus: %w", err)
					return false, nil
				}

				if len(res) != 1 {
					loopErr = fmt.Errorf("expecting 1 item but got %d", len(res))
					return false, nil
				}

				for k, v := range res[0].Metric {
					if k == "namespace" {
						namespaceLabel = v
						break
					}
				}

				return true, nil
			})

			require.NoError(t, err, "%v:%v", err, loopErr)
			require.Equal(t, tc.expectedNamespace, namespaceLabel)
		})
	}
}

func testPrometheusCRDValidation(t *testing.T) {
	t.Parallel()
	name := "test"
	replicas := int32(1)

	tests := []struct {
		name           string
		prometheusSpec monitoringv1.PrometheusSpec
		expectedError  bool
	}{
		//
		// RetentionSize Validation:
		//
		{
			name: "zero-size-without-unit",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
				},
				RetentionSize: "0",
			},
		},
		{
			name: "legacy-unit",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
				},
				RetentionSize: "1.5GB",
			},
		},
		{
			name: "iec-unit",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
				},
				RetentionSize: "100MiB",
			},
		},
		{
			name: "legacy-missing-symbol",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
				},
				RetentionSize: "10M",
			},
			expectedError: true,
		},
		{
			name: "legacy-missing-unit",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
				},
				RetentionSize: "1000",
			},
			expectedError: true,
		},
		{
			name: "iec-missing-symbol",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
				},
				RetentionSize: "15Gi",
			},
			expectedError: true,
		},
		//
		// ScrapeInterval validation
		{
			name: "zero-time-without-unit",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
					ScrapeInterval: "0",
				},
			},
		},
		{
			name: "time-in-seconds-unit",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
					ScrapeInterval: "30s",
				},
			},
		},
		{
			name: "complex-time-unit",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
					ScrapeInterval: "1h30m15s",
				},
			},
		},
		{
			name: "time-missing-symbols",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
					ScrapeInterval: "600",
				},
			},
			expectedError: true,
		},
		{
			name: "time-unit-misspelled",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
					ScrapeInterval: "60ss",
				},
			},
			expectedError: true,
		},
		{
			name: "max-connections",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
					Web: &monitoringv1.PrometheusWebSpec{
						MaxConnections: proto.Int32(100),
					},
				},
			},
		},
		{
			name: "max-connections-negative-value",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
					Web: &monitoringv1.PrometheusWebSpec{
						MaxConnections: proto.Int32(-1),
					},
				},
			},
			expectedError: true,
		},
		{
			name: "max-concurrency",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
				},
				Query: &monitoringv1.QuerySpec{
					MaxConcurrency: ptr.To(int32(100)),
				},
			},
		},
		{
			name: "max-concurrency-zero-value",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
				},
				Query: &monitoringv1.QuerySpec{
					MaxConcurrency: ptr.To(int32(0)),
				},
			},
			expectedError: true,
		},
		//
		// Alertmanagers-Endpoints tests
		{
			name: "no-endpoint-namespace",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
				},
				Alerting: &monitoringv1.AlertingSpec{
					Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
						{
							Name:            "test",
							Port:            intstr.FromInt(9797),
							Scheme:          "https",
							PathPrefix:      "/alerts",
							BearerTokenFile: "/file",
							APIVersion:      "v1",
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "endpoint-namespace",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
				},
				Alerting: &monitoringv1.AlertingSpec{
					Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
						{
							Name:            "test",
							Namespace:       ptr.To("default"),
							Port:            intstr.FromInt(9797),
							Scheme:          "https",
							PathPrefix:      "/alerts",
							BearerTokenFile: "/file",
							APIVersion:      "v1",
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "no-endpoint-name",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas:           &replicas,
					Version:            operator.DefaultPrometheusVersion,
					ServiceAccountName: "prometheus",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("400Mi"),
						},
					},
				},
				Alerting: &monitoringv1.AlertingSpec{
					Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
						{
							Namespace:       ptr.To("default"),
							Port:            intstr.FromInt(9797),
							Scheme:          "https",
							PathPrefix:      "/alerts",
							BearerTokenFile: "/file",
							APIVersion:      "v1",
						},
					},
				},
			},
			expectedError: true,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)
			ns := framework.CreateNamespace(context.Background(), t, testCtx)
			framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
			prom := &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{
					Name:        name,
					Namespace:   ns,
					Annotations: map[string]string{},
				},
				Spec: test.prometheusSpec,
			}

			if test.expectedError {
				_, err := framework.MonClientV1.Prometheuses(ns).Create(context.Background(), prom, metav1.CreateOptions{})
				require.Error(t, err)
				require.True(t, apierrors.IsInvalid(err))
				return
			}

			_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prom)
			require.NoError(t, err)
		})
	}
}

func testRelabelConfigCRDValidation(t *testing.T) {
	t.Parallel()
	name := "test"
	tests := []struct {
		scenario       string
		relabelConfigs []monitoringv1.RelabelConfig
		expectedError  bool
	}{
		{
			scenario: "no-explicit-sep",
			relabelConfigs: []monitoringv1.RelabelConfig{
				{
					SourceLabels: []monitoringv1.LabelName{"__address__"},
					Action:       "replace",
					Regex:        "([^:]+)(?::\\d+)?",
					Replacement:  ptr.To("$1:80"),
					TargetLabel:  "__address__",
				},
			},
		},
		{
			scenario: "no-explicit-action",
			relabelConfigs: []monitoringv1.RelabelConfig{
				{
					SourceLabels: []monitoringv1.LabelName{"__address__"},
					Separator:    ptr.To(","),
					Regex:        "([^:]+)(?::\\d+)?",
					Replacement:  ptr.To("$1:80"),
					TargetLabel:  "__address__",
				},
			},
		},
		{
			scenario: "empty-separator",
			relabelConfigs: []monitoringv1.RelabelConfig{
				{
					Separator: ptr.To(""),
				},
			},
		},
		{
			scenario: "invalid-action",
			relabelConfigs: []monitoringv1.RelabelConfig{
				{
					Action: "replacee",
				},
			},
			expectedError: true,
		},
		{
			scenario: "empty-source-lbl",
			relabelConfigs: []monitoringv1.RelabelConfig{
				{
					SourceLabels: []monitoringv1.LabelName{""},
				},
			},
			expectedError: true,
		},
		{
			scenario: "invalid-source-lbl",
			relabelConfigs: []monitoringv1.RelabelConfig{
				{
					SourceLabels: []monitoringv1.LabelName{"metric%)"},
				},
			},
			expectedError: true,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.scenario, func(t *testing.T) {
			t.Parallel()
			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)
			ns := framework.CreateNamespace(context.Background(), t, testCtx)
			framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

			p := framework.MakeBasicPrometheus(ns, name, "", 1)

			_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
			require.NoError(t, err)

			s := framework.MakeBasicServiceMonitor(name)
			s.Spec.Endpoints[0].RelabelConfigs = test.relabelConfigs
			_, err = framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{})

			if err == nil {
				require.False(t, test.expectedError)
				return
			}

			require.True(t, apierrors.IsInvalid(err))
		})
	}
}

func testPromQueryLogFile(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	p := framework.MakeBasicPrometheus(ns, "test", "", 1)
	p.Spec.QueryLogFile = "query.log"
	_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)
}

func testPromUnavailableConditionStatus(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	p := framework.MakeBasicPrometheus(ns, "test", "", 2)
	// A non-existing service account prevents the creation of the statefulset's pods.
	p.Spec.ServiceAccountName = "does-not-exist"

	_, err := framework.MonClientV1.Prometheuses(ns).Create(context.Background(), p, metav1.CreateOptions{})
	require.NoError(t, err)

	var pollErr error
	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 5*time.Minute, false, func(ctx context.Context) (bool, error) {
		var current *monitoringv1.Prometheus
		current, pollErr = framework.MonClientV1.Prometheuses(p.Namespace).Get(ctx, p.Name, metav1.GetOptions{})
		if pollErr != nil {
			return false, nil
		}

		for _, cond := range current.Status.Conditions {
			if cond.Type != monitoringv1.Available {
				continue
			}

			if cond.Status != monitoringv1.ConditionFalse {
				pollErr = fmt.Errorf(
					"expected Available condition to be 'False', got %q (reason %s, %q)",
					cond.Status,
					cond.Reason,
					cond.Message,
				)
				return false, nil
			}

			if cond.Reason != "NoPodReady" {
				pollErr = fmt.Errorf(
					"expected Available condition's reason to be 'NoPodReady',  got %s (message %q)",
					cond.Reason,
					cond.Message,
				)
				return false, nil
			}

			return true, nil
		}

		pollErr = fmt.Errorf("failed to find Available condition in status subresource")
		return false, nil
	})

	require.NoError(t, err)
}

func testPromDegradedConditionStatus(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	p := framework.MakeBasicPrometheus(ns, "test", "", 2)
	_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)

	// Roll out a new version of the Prometheus object that references a
	// non-existing container image which should trigger the
	// "Available=Degraded" condition.
	p, err = framework.PatchPrometheus(
		context.Background(),
		p.Name,
		ns,
		monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Containers: []v1.Container{{
					Name:  "bad-image",
					Image: "quay.io/prometheus-operator/invalid-image",
				}},
			},
		},
	)
	require.NoError(t, err)

	var pollErr error
	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 5*time.Minute, false, func(ctx context.Context) (bool, error) {
		var current *monitoringv1.Prometheus
		current, pollErr = framework.MonClientV1.Prometheuses(p.Namespace).Get(ctx, p.Name, metav1.GetOptions{})
		if pollErr != nil {
			return false, nil
		}

		for _, cond := range current.Status.Conditions {
			if cond.Type != monitoringv1.Available {
				continue
			}

			if cond.Status != monitoringv1.ConditionDegraded {
				pollErr = fmt.Errorf(
					"expected Available condition to be 'Degraded', got %q (reason %s, %q)",
					cond.Status,
					cond.Reason,
					cond.Message,
				)
				return false, nil
			}

			if cond.Reason != "SomePodsNotReady" {
				pollErr = fmt.Errorf(
					"expected Available condition's reason to be 'SomePodsNotReady',  got %s (message %q)",
					cond.Reason,
					cond.Message,
				)
				return false, nil
			}

			if !strings.Contains(cond.Message, "bad-image") {
				pollErr = fmt.Errorf(
					"expected Available condition's message to contain 'bad-image', got %q",
					cond.Message,
				)
				return false, nil
			}

			return true, nil
		}

		pollErr = fmt.Errorf("failed to find Available condition in status subresource")
		return false, nil
	})

	require.NoError(t, err)
}

func testPromStrategicMergePatch(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "secret", Namespace: ns},
		Type:       v1.SecretType("Opaque"),
		Data:       map[string][]byte{},
	}
	_, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), secret, metav1.CreateOptions{})
	require.NoError(t, err)

	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "configmap", Namespace: ns},
		Data:       map[string]string{},
	}
	_, err = framework.KubeClient.CoreV1().ConfigMaps(ns).Create(context.Background(), configmap, metav1.CreateOptions{})
	require.NoError(t, err)

	p := framework.MakeBasicPrometheus(ns, "test", "", 1)
	p.Spec.Secrets = []string{secret.Name}
	p.Spec.ConfigMaps = []string{configmap.Name}
	p.Spec.Containers = []v1.Container{{
		Name:  "sidecar",
		Image: "nginx",
		// Ensure that the sidecar container can mount the additional secret and configmap.
		VolumeMounts: []v1.VolumeMount{{
			Name:      "secret-" + secret.Name,
			MountPath: "/tmp/secret",
		}, {
			Name:      "configmap-" + configmap.Name,
			MountPath: "/tmp/configmap",
		}},
	}}

	_, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	require.NoError(t, err)
}

func testPrometheusWithStatefulsetCreationFailure(t *testing.T) {
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	p := framework.MakeBasicPrometheus(ns, "test", "", 1)
	// Invalid spec which prevents the creation of the statefulset
	p.Spec.Web = &monitoringv1.PrometheusWebSpec{
		WebConfigFileFields: monitoringv1.WebConfigFileFields{
			TLSConfig: &monitoringv1.WebTLSConfig{
				Cert: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{},
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "tls-cert",
						},
						Key: "tls.crt",
					},
				},
				KeySecret: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "tls-cert",
					},
					Key: "tls.key",
				},
			},
		},
	}
	_, err := framework.MonClientV1.Prometheuses(p.Namespace).Create(ctx, p, metav1.CreateOptions{})
	require.NoError(t, err)

	var loopError error
	err = wait.PollUntilContextTimeout(ctx, time.Second, framework.DefaultTimeout, true, func(ctx context.Context) (bool, error) {
		current, err := framework.MonClientV1.Prometheuses(ns).Get(ctx, "test", metav1.GetOptions{})
		if err != nil {
			loopError = fmt.Errorf("failed to get object: %w", err)
			return false, nil
		}

		if err := framework.AssertCondition(current.Status.Conditions, monitoringv1.Reconciled, monitoringv1.ConditionFalse); err != nil {
			loopError = err
			return false, nil
		}

		if err := framework.AssertCondition(current.Status.Conditions, monitoringv1.Available, monitoringv1.ConditionFalse); err != nil {
			loopError = err
			return false, nil
		}

		return true, nil
	})

	require.NotErrorIs(t, err, loopError)

	require.NoError(t, framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, "test"))
}

func testPrometheusStatusScale(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(ctx, t, testCtx)
	framework.SetupPrometheusRBAC(ctx, t, testCtx, ns)
	name := "test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.CommonPrometheusFields.Shards = proto.Int32(1)

	p, err := framework.CreatePrometheusAndWaitUntilReady(ctx, ns, p)
	require.NoError(t, err)

	require.Equal(t, int32(1), p.Status.Shards)

	p, err = framework.ScalePrometheusAndWaitUntilReady(ctx, name, ns, 2)
	require.NoError(t, err)

	require.Equal(t, int32(2), p.Status.Shards)
}

func isAlertmanagerDiscoveryWorking(ns, promSVCName, alertmanagerName string) func(ctx context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		pods, err := framework.KubeClient.CoreV1().Pods(ns).List(ctx, alertmanager.ListOptions(alertmanagerName))
		if err != nil {
			return false, err
		}
		if 3 != len(pods.Items) {
			return false, nil
		}
		expectedAlertmanagerTargets := []string{}
		for _, p := range pods.Items {
			expectedAlertmanagerTargets = append(expectedAlertmanagerTargets, fmt.Sprintf("http://%s:9093/api/v2/alerts", p.Status.PodIP))
		}

		response, err := framework.PrometheusSVCGetRequest(context.Background(), ns, promSVCName, "http", "/api/v1/alertmanagers", map[string]string{})
		if err != nil {
			return false, err
		}

		ra := prometheusAlertmanagerAPIResponse{}
		if err := json.NewDecoder(bytes.NewBuffer(response)).Decode(&ra); err != nil {
			return false, err
		}

		if assertExpectedAlertmanagerTargets(ra.Data.ActiveAlertmanagers, expectedAlertmanagerTargets) {
			return true, nil
		}

		return false, nil
	}
}

func assertExpectedAlertmanagerTargets(ams []*alertmanagerTarget, expectedTargets []string) bool {
	log.Printf("Expected Alertmanager Targets: %#+v\n", expectedTargets)

	existingTargets := []string{}

	for _, am := range ams {
		existingTargets = append(existingTargets, am.URL)
	}

	sort.Strings(expectedTargets)
	sort.Strings(existingTargets)

	if !reflect.DeepEqual(expectedTargets, existingTargets) {
		log.Printf("Existing Alertmanager Targets: %#+v\n", existingTargets)
		return false
	}

	return true
}

type alertmanagerTarget struct {
	URL string `json:"url"`
}

type alertmanagerDiscovery struct {
	ActiveAlertmanagers []*alertmanagerTarget `json:"activeAlertmanagers"`
}

type prometheusAlertmanagerAPIResponse struct {
	Status string                 `json:"status"`
	Data   *alertmanagerDiscovery `json:"data"`
}
