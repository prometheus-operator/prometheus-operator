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
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	certutil "k8s.io/client-go/util/cert"

	"google.golang.org/protobuf/proto"

	"github.com/prometheus-operator/prometheus-operator/pkg/alertmanager"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	"github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"

	"github.com/kylelemons/godebug/pretty"
	"github.com/pkg/errors"
)

var (
	certsDir = "../../test/e2e/remote_write_certs/"
)

func createK8sResources(t *testing.T, ns, certsDir string, cKey testFramework.Key, cCert, ca testFramework.Cert) {
	var clientKey, clientCert, serverKey, serverCert, caCert []byte
	var err error

	if cKey.Filename != "" {
		clientKey, err = ioutil.ReadFile(certsDir + cKey.Filename)
		if err != nil {
			t.Fatalf("failed to load %s: %v", cKey.Filename, err)
		}
	}

	if cCert.Filename != "" {
		clientCert, err = ioutil.ReadFile(certsDir + cCert.Filename)
		if err != nil {
			t.Fatalf("failed to load %s: %v", cCert.Filename, err)
		}
	}

	if ca.Filename != "" {
		caCert, err = ioutil.ReadFile(certsDir + ca.Filename)
		if err != nil {
			t.Fatalf("failed to load %s: %v", ca.Filename, err)
		}
	}

	serverKey, err = ioutil.ReadFile(certsDir + "ca.key")
	if err != nil {
		t.Fatalf("failed to load %s: %v", "ca.key", err)
	}

	serverCert, err = ioutil.ReadFile(certsDir + "ca.crt")
	if err != nil {
		t.Fatalf("failed to load %s: %v", "ca.crt", err)
	}

	scrapingKey, err := ioutil.ReadFile(certsDir + "client.key")
	if err != nil {
		t.Fatalf("failed to load %s: %v", "client.key", err)
	}

	scrapingCert, err := ioutil.ReadFile(certsDir + "client.crt")
	if err != nil {
		t.Fatalf("failed to load %s: %v", "client.crt", err)
	}

	var s *v1.Secret
	var cm *v1.ConfigMap
	secrets := []*v1.Secret{}
	configMaps := []*v1.ConfigMap{}

	s = testFramework.MakeSecretWithCert(ns, "scraping-tls",
		[]string{"key.pem", "cert.pem"}, [][]byte{scrapingKey, scrapingCert})
	secrets = append(secrets, s)

	s = testFramework.MakeSecretWithCert(ns, "server-tls",
		[]string{"key.pem", "cert.pem"}, [][]byte{serverKey, serverCert})
	secrets = append(secrets, s)

	s = testFramework.MakeSecretWithCert(ns, "server-tls-ca",
		[]string{"ca.pem"}, [][]byte{serverCert})
	secrets = append(secrets, s)

	if cKey.Filename != "" && cCert.Filename != "" {
		s = testFramework.MakeSecretWithCert(ns, cKey.SecretName,
			[]string{"key.pem"}, [][]byte{clientKey})
		secrets = append(secrets, s)

		if cCert.ResourceType == testFramework.SECRET {
			if cCert.ResourceName == cKey.SecretName {
				s.Data["cert.pem"] = clientCert
			} else {
				s = testFramework.MakeSecretWithCert(ns, cCert.ResourceName,
					[]string{"cert.pem"}, [][]byte{clientCert})
				secrets = append(secrets, s)
			}
		} else if cCert.ResourceType == testFramework.CONFIGMAP {
			cm = testFramework.MakeConfigMapWithCert(framework.KubeClient, ns, cCert.ResourceName,
				"", "cert.pem", "", nil, clientCert, nil)
			configMaps = append(configMaps, cm)
		} else {
			t.Fatal("cert must be a Secret or a ConfigMap")
		}
	}

	if ca.Filename != "" {
		if ca.ResourceType == testFramework.SECRET {
			if ca.ResourceName == cKey.SecretName {
				secrets[3].Data["ca.pem"] = caCert
			} else if ca.ResourceName == cCert.ResourceName {
				s.Data["ca.pem"] = caCert
			} else {
				s = testFramework.MakeSecretWithCert(ns, ca.ResourceName,
					[]string{"ca.pem"}, [][]byte{caCert})
				secrets = append(secrets, s)
			}
		} else if ca.ResourceType == testFramework.CONFIGMAP {
			if ca.ResourceName == cCert.ResourceName {
				cm.Data["ca.pem"] = string(caCert)
			} else {
				cm = testFramework.MakeConfigMapWithCert(framework.KubeClient, ns, ca.ResourceName,
					"", "", "ca.pem", nil, nil, caCert)
				configMaps = append(configMaps, cm)
			}
		} else {
			t.Fatal("cert must be a Secret or a ConfigMap")
		}
	}

	for _, s = range secrets {
		_, err := framework.KubeClient.CoreV1().Secrets(s.ObjectMeta.Namespace).Create(context.Background(), s, metav1.CreateOptions{})
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, cm = range configMaps {
		_, err := framework.KubeClient.CoreV1().ConfigMaps(ns).Create(context.Background(), cm, metav1.CreateOptions{})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func createK8sSampleApp(t *testing.T, name, ns string) {
	simple, err := testFramework.MakeDeployment("../../test/framework/resources/basic-auth-app-deployment.yaml")
	if err != nil {
		t.Fatal(err)
	}

	simple.Spec.Template.Spec.Containers[0].Args = []string{"--cert-path=/etc/certs"}

	simple.Spec.Template.Spec.Volumes = []v1.Volume{
		{
			Name: "tls-certs",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: "server-tls",
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

	if err := framework.CreateDeployment(context.Background(), ns, simple); err != nil {
		t.Fatal("Creating simple basic auth app failed: ", err)
	}

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

	if _, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
		t.Fatal(err)
	}

	_, err = framework.KubeClient.CoreV1().Services(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func createK8sAppMonitoring(name, ns string, prwtc testFramework.PromRemoteWriteTestConfig) (prometheus *monitoringv1.Prometheus, prometheusRecieverSvc string, err error) {

	sm := framework.MakeBasicServiceMonitor(name)
	sm.Spec.Endpoints = []monitoringv1.Endpoint{
		{
			Port:     "mtls",
			Interval: "30s",
			Scheme:   "https",
			TLSConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					InsecureSkipVerify: true,
					Cert: monitoringv1.SecretOrConfigMap{
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "scraping-tls",
							},
							Key: "cert.pem",
						},
					},
					KeySecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "scraping-tls",
						},
						Key: "key.pem",
					},
				},
			},
		},
	}

	if _, err = framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), sm, metav1.CreateOptions{}); err != nil {
		return nil, prometheusRecieverSvc, errors.Wrap(err, "creating ServiceMonitor failed")
	}

	// Create prometheus receiver for remote writes
	receiverName := fmt.Sprintf("%s-%s", name, "receiver")
	prometheusReceiverCRD := framework.MakeBasicPrometheus(ns, receiverName, receiverName, 1)
	framework.AddRemoteReceiveWithWebTLSToPrometheus(prometheusReceiverCRD, prwtc)

	if _, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusReceiverCRD); err != nil {
		return nil, "", err
	}
	prometheusReceiverSvc := framework.MakePrometheusService(receiverName, receiverName, v1.ServiceTypeClusterIP)
	if _, err = framework.CreateServiceAndWaitUntilReady(context.Background(), ns, prometheusReceiverSvc); err != nil {
		return nil, "", err
	}
	prometheusReceiverURL := "https://" + prometheusReceiverSvc.Name + ":9090/api/v1/write"

	// Create prometheus for scraping app metrics with remote prometheus as write target
	prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)
	framework.AddRemoteWriteWithTLSToPrometheus(prometheusCRD, prometheusReceiverURL, prwtc)
	if _, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD); err != nil {
		return nil, "", err
	}

	promSVC := framework.MakePrometheusService(prometheusCRD.Name, name, v1.ServiceTypeClusterIP)
	if _, err = framework.CreateServiceAndWaitUntilReady(context.Background(), ns, promSVC); err != nil {
		return nil, "", err
	}

	return prometheusCRD, prometheusReceiverSvc.Name, nil
}

func testPromRemoteWriteWithTLS(t *testing.T) {
	t.Parallel()
	// can't extend the names since ns cannot be created with more than 63 characters
	var tests = []testFramework.PromRemoteWriteTestConfig{
		// working configurations
		{
			Name: "variant-1",
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
			ShouldSuccess:      true,
		},
		{
			Name: "variant-2",
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
			ShouldSuccess:      true,
		},
		{
			Name: "variant-3",
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
			ShouldSuccess:      true,
		},
		{
			Name: "variant-4",
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
			ShouldSuccess:      true,
		},
		{
			Name: "variant-5",
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
			ShouldSuccess:      true,
		},
		{
			Name: "variant-6",
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
			ShouldSuccess:      true,
		},
		{
			Name: "variant-7",
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
			ShouldSuccess:      true,
		},
		{
			Name: "variant-8",
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
			ShouldSuccess:      true,
		},
		{
			Name: "variant-9",
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
			ShouldSuccess:      true,
		},
		{
			Name: "variant-10",
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
			ShouldSuccess:      true,
		},
		{
			Name: "variant-11",
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
			ShouldSuccess:      true,
		},
		{
			Name: "variant-12",
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
			ShouldSuccess:      true,
		},
		// non working configurations
		// we will check it only for one configuration for simplicity - only one Secret
		{
			Name: "variant-13",
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
			ShouldSuccess:      false,
		},
		{
			Name: "variant-14",
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
			ShouldSuccess:      false,
		},
		{
			Name: "variant-15",
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
			ShouldSuccess:      false,
		},
		{
			Name: "variant-16",
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
			ShouldSuccess:      false,
		},
		{
			Name: "variant-17",
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
			ShouldSuccess:      false,
		},
		{
			Name: "variant-18",
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
			ShouldSuccess:      false,
		},
		{
			Name: "variant-19",
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
			ShouldSuccess:      false,
		},
		// Had to change the success flag to True, because prometheus receiver is running in VerifyClientCertIfGiven mode. Details here - https://github.com/prometheus-operator/prometheus-operator/pull/4337#discussion_r735064646
		{
			Name: "variant-20",
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
			ShouldSuccess:      true,
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			testCtx := framework.NewTestCtx(t)
			defer testCtx.Cleanup(t)

			ns := framework.CreateNamespace(context.Background(), t, testCtx)
			framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)
			name := "test"

			// apply authorized certificate and key to k8s as a Secret
			createK8sResources(t, ns, certsDir, test.ClientKey, test.ClientCert, test.CA)

			// Setup a sample-app which supports mTLS therefore will play 2 roles:
			// 	1. app scraped by prometheus
			// 	2. TLS receiver for prometheus remoteWrite
			createK8sSampleApp(t, name, ns)

			// Setup monitoring.
			prometheusCRD, prometheusRecieverSvc, err := createK8sAppMonitoring(name, ns, test)
			if err != nil {
				t.Fatal(err)
			}

			// Check for proper scraping.
			promSVC := framework.MakePrometheusService(name, name, v1.ServiceTypeClusterIP)
			if err := framework.WaitForHealthyTargets(context.Background(), ns, promSVC.Name, 1); err != nil {
				framework.PrintPrometheusLogs(context.Background(), t, prometheusCRD)
				t.Fatal(err)
			}

			//TODO: make it wait by poll, there are some examples in other tests
			// use wait.Poll() in k8s.io/apimachinery@v0.18.3/pkg/util/wait/wait.go
			time.Sleep(45 * time.Second)

			response, err := framework.PrometheusQuery(ns, prometheusRecieverSvc, "https", "up{container = 'example-app'}")
			if test.ShouldSuccess {
				if err != nil {
					t.Logf("test with (%s, %s, %s) failed with error %s", test.ClientKey.Filename, test.ClientCert.Filename, test.CA.Filename, err.Error())
				}
				if response[0].Value[1] != "1" {
					framework.PrintPrometheusLogs(context.Background(), t, prometheusCRD)
					t.Fatalf("test with (%s, %s, %s) failed\nReciever Prometheus does not have the instrumented app metrics",
						test.ClientKey.Filename, test.ClientCert.Filename, test.CA.Filename)
				}
			} else {
				if err != nil {
					framework.PrintPrometheusLogs(context.Background(), t, prometheusCRD)
					t.Fatalf("test with (%s, %s, %s) failed with error %s", test.ClientKey.Filename, test.ClientCert.Filename, test.CA.Filename, err.Error())
				}
				if len(response) != 0 {
					t.Fatalf("test with (%s, %s, %s) failed\nExpeted reciever prometheus to not have the instrumented app metrics",
						test.ClientKey.Filename, test.ClientCert.Filename, test.CA.Filename)
				}
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

	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD); err != nil {
		t.Fatal(err)
	}

	if err := framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, name); err != nil {
		t.Fatal(err)
	}
}

func testPromScaleUpDownCluster(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"

	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, framework.MakeBasicPrometheus(ns, name, name, 1))
	if err != nil {
		t.Fatal(err)
	}

	p.Spec.Replicas = proto.Int32(3)
	p, err = framework.UpdatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}

	p.Spec.Replicas = proto.Int32(2)
	_, err = framework.UpdatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}
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
	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p); err != nil {
		t.Fatal(err)
	}
}

func testPromVersionMigration(t *testing.T) {
	t.Parallel()
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
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range compatibilityMatrix {
		p.Spec.Version = v
		p, err = framework.UpdatePrometheusAndWaitUntilReady(context.Background(), ns, p)
		if err != nil {
			t.Fatal(err)
		}
		if err := framework.WaitForPrometheusRunImageAndReady(context.Background(), ns, p); err != nil {
			t.Fatal(err)
		}
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
	if err != nil {
		t.Fatal(err)
	}

	pods, err := framework.KubeClient.CoreV1().Pods(ns).List(context.Background(), prometheus.ListOptions(name))
	if err != nil {
		t.Fatal(err)
	}
	res := pods.Items[0].Spec.Containers[0].Resources

	if !reflect.DeepEqual(res, p.Spec.Resources) {
		t.Fatalf("resources don't match. Has %#+v, want %#+v", res, p.Spec.Resources)
	}

	p.Spec.Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("200Mi"),
		},
	}
	p, err = framework.MonClientV1.Prometheuses(ns).Update(context.Background(), p, metav1.UpdateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	err = wait.Poll(5*time.Second, 2*time.Minute, func() (bool, error) {
		pods, err := framework.KubeClient.CoreV1().Pods(ns).List(context.Background(), prometheus.ListOptions(name))
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

	if err != nil {
		t.Fatal(err)
	}
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
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse("200Mi"),
					},
				},
			},
		},
	}

	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}

	if val := p.Spec.Storage.VolumeClaimTemplate.Labels["test-label"]; val != "foo" {
		t.Errorf("incorrect volume claim label, want: %v, got: %v", "foo", val)
	}
	if val := p.Spec.Storage.VolumeClaimTemplate.Annotations["test-annotation"]; val != "bar" {
		t.Errorf("incorrect volume claim annotation, want: %v, got: %v", "bar", val)
	}

	err = wait.Poll(5*time.Second, 2*time.Minute, func() (bool, error) {
		sts, err := framework.KubeClient.AppsV1().StatefulSets(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		if len(sts.Items) < 1 {
			return false, nil
		}

		for _, vct := range sts.Items[0].Spec.VolumeClaimTemplates {
			if vct.Name == "prometheus-"+name+"-db" {
				if val := vct.Labels["test-label"]; val != "foo" {
					return false, errors.Errorf("incorrect volume claim label on sts, want: %v, got: %v", "foo", val)
				}
				if val := vct.Annotations["test-annotation"]; val != "bar" {
					return false, errors.Errorf("incorrect volume claim annotation on sts, want: %v, got: %v", "bar", val)
				}
				return true, nil
			}
		}

		return false, nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func testPromStorageUpdate(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)

	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}

	p.Spec.Storage = &monitoringv1.StorageSpec{
		VolumeClaimTemplate: monitoringv1.EmbeddedPersistentVolumeClaim{
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse("200Mi"),
					},
				},
			},
		},
	}
	_, err = framework.MonClientV1.Prometheuses(ns).Update(context.Background(), p, metav1.UpdateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	err = wait.Poll(5*time.Second, 2*time.Minute, func() (bool, error) {
		pods, err := framework.KubeClient.CoreV1().Pods(ns).List(context.Background(), prometheus.ListOptions(name))
		if err != nil {
			return false, err
		}

		if len(pods.Items) != 1 {
			return false, nil
		}

		for _, volume := range pods.Items[0].Spec.Volumes {
			if volume.Name == "prometheus-"+name+"-db" && volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName != "" {
				return true, nil
			}
		}

		return false, nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func testPromReloadConfig(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	name := "test"
	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.ServiceMonitorSelector = nil
	p.Spec.PodMonitorSelector = nil

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
	if err := gzipConfig(&bufOne, []byte(firstConfig)); err != nil {
		t.Fatal(err)
	}
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

	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), cfg, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
		t.Fatal(err)
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	if err := framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 1); err != nil {
		t.Fatal(err)
	}

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
	if err := gzipConfig(&bufTwo, []byte(secondConfig)); err != nil {
		t.Fatal(err)
	}
	secondConfigCompressed := bufTwo.Bytes()

	cfg, err := framework.KubeClient.CoreV1().Secrets(ns).Get(context.Background(), cfg.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(errors.Wrap(err, "could not retrieve previous secret"))
	}

	cfg.Data["prometheus.yaml.gz"] = secondConfigCompressed
	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Update(context.Background(), cfg, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}

	if err := framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 2); err != nil {
		t.Fatal(err)
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
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{}); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

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
	_, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), &secret, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	p.Spec.AdditionalScrapeConfigs = &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: "additional-scrape-configs",
		},
		Key: "prometheus-additional.yaml",
	}
	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	// Wait for ServiceMonitor target, as well as additional-config target
	if err := framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 2); err != nil {
		t.Fatal(err)
	}
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
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{}); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

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
	_, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), &secret, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	p.Spec.AdditionalAlertManagerConfigs = &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: "additional-alert-configs",
		},
		Key: "prometheus-additional.yaml",
	}
	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	// Wait for ServiceMonitor target
	if err := framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 1); err != nil {
		t.Fatal(err)
	}

	err = wait.Poll(time.Second, 5*time.Minute, func() (done bool, err error) {
		response, err := framework.PrometheusSVCGetRequest(context.Background(), ns, svc.Name, "http", "/api/v1/alertmanagers", map[string]string{})
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

	if err != nil {
		t.Fatal(errors.Wrap(err, "validating Prometheus Alertmanager configuration failed"))
	}
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
	if err != nil {
		t.Fatal(err)
	}

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	p, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}

	pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, pSVC); err != nil {
		t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	err = framework.WaitForPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, firtAlertName)
	if err != nil {
		t.Fatal(err)
	}

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
	if err != nil {
		t.Fatal(err)
	}

	err = framework.WaitForPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, secondAlertName)
	if err != nil {
		t.Fatal(err)
	}
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
		if err != nil {
			t.Fatal(err)
		}
	}

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}

	pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, pSVC); err != nil {
		t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	for _, alertName := range alertNames {
		err := framework.WaitForPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, alertName)
		if err != nil {
			t.Fatal(err)
		}
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
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, file := range ruleFiles {
		_, err := framework.MakeAndCreateFiringRule(context.Background(), file.ns, file.alertName, file.alertName)
		if err != nil {
			t.Fatal(err)
		}
	}

	p := framework.MakeBasicPrometheus(rootNS, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	p.Spec.RuleNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: ruleFilesNamespaceSelector,
	}
	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), rootNS, p)
	if err != nil {
		t.Fatal(err)
	}

	pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), rootNS, pSVC); err != nil {
		t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	for _, file := range ruleFiles {
		err := framework.WaitForPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, file.alertName)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Remove the selecting label from the namespaces holding PrometheusRules
	// and wait until the rules are removed from Prometheus.
	// See https://github.com/prometheus-operator/prometheus-operator/issues/3847
	for _, file := range ruleFiles {
		if err := framework.RemoveLabelsFromNamespace(context.Background(), file.ns, "monitored"); err != nil {
			t.Fatal(err)
		}
	}

	for _, file := range ruleFiles {
		var loopError error
		err = wait.Poll(time.Second, 5*framework.DefaultTimeout, func() (bool, error) {
			var firing bool
			firing, loopError = framework.CheckPrometheusFiringAlert(context.Background(), file.ns, pSVC.Name, file.alertName)
			return !firing, nil
		})

		if err != nil {
			t.Fatalf("waiting for alert %q in namespace %s to stop firing: %v: %v", file.alertName, file.ns, err, loopError)
		}
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
		if err != nil {
			t.Fatal(err)
		}
		prometheusRules = append(prometheusRules, rule)
	}

	name := "test"

	p := framework.MakeBasicPrometheus(ns, name, name, 1)
	p.Spec.EvaluationInterval = "1s"
	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if t.Failed() {
			if err := framework.PrintPodLogs(context.Background(), ns, "prometheus-"+p.Name+"-0"); err != nil {
				t.Fatal(err)
			}
		}
	}()

	pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, pSVC); err != nil {
		t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	for i := range prometheusRules {
		_, err := framework.WaitForConfigMapExist(context.Background(), ns, "prometheus-"+p.Name+"-rulefiles-"+strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
	}

	// Make sure both rule files ended up in the Prometheus Pod
	for i := range prometheusRules {
		err := framework.WaitForPrometheusFiringAlert(context.Background(), ns, pSVC.Name, "my-alert-"+strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
	}

	err = framework.DeleteRule(context.Background(), ns, prometheusRules[1].Name)
	if err != nil {
		t.Fatal(err)
	}

	_, err = framework.WaitForConfigMapExist(context.Background(), ns, "prometheus-"+p.Name+"-rulefiles-0")
	if err != nil {
		t.Fatal(err)
	}
	err = framework.WaitForConfigMapNotExist(context.Background(), ns, "prometheus-"+p.Name+"-rulefiles-1")
	if err != nil {
		t.Fatal(err)
	}

	err = framework.WaitForPrometheusFiringAlert(context.Background(), ns, pSVC.Name, "my-alert-0")
	if err != nil {
		t.Fatal(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}

	rule, err := framework.GetRule(context.Background(), ns, name)
	if err != nil {
		t.Fatal(err)
	}

	val, ok := rule.Annotations["prometheus-operator-validated"]
	if !ok {
		t.Fatal("Expected prometheusrule to be annotated")
	}
	if val != "true" {
		t.Fatal("Expected prometheusrule annotation to be 'true'")
	}
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
	if err == nil {
		t.Fatal("Expected invalid prometheusrule to be rejected")
	}
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
// and the Prometheus rules configmap on relevant changes
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
			Name: "prometheus",
			Getter: func(prometheusName string) (versionedResource, error) {
				return framework.
					MonClientV1.
					Prometheuses(ns).
					Get(context.Background(), prometheusName, metav1.GetOptions{})
			},
			MaxExpectedChanges: 1,
		},
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
			Getter: func(prometheusName string) (versionedResource, error) {
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
	if _, err := framework.MakeAndCreateFiringRule(context.Background(), ns, "my-prometheus-rule", alertName); err != nil {
		t.Fatal(err)
	}

	prometheus, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheus)
	if err != nil {
		t.Fatal(err)
	}

	pSVC := framework.MakePrometheusService(prometheus.Name, name, v1.ServiceTypeClusterIP)
	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, pSVC); err != nil {
		t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	s := framework.MakeBasicServiceMonitor(name)
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{}); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	err = framework.WaitForPrometheusFiringAlert(context.Background(), prometheus.Namespace, pSVC.Name, alertName)
	if err != nil {
		t.Fatal(err)
	}

	err = framework.WaitForDiscoveryWorking(context.Background(), ns, pSVC.Name, prometheus.Name)
	if err != nil {
		t.Fatal(errors.Wrap(err, "validating Prometheus target discovery failed"))
	}

	if err := framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, name); err != nil {
		t.Fatal(err)
	}

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
				fmt.Println(pretty.Compare(previous, version))
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
	if err != nil {
		t.Fatal(err)
	}

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
		if err != nil {
			t.Fatal(err)
		}

		updateObjectLabels(res, updatedLabels)
		updateObjectAnnotations(res, updatedAnnotations)

		_, err = rConf.update(res)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Ensure resource reconciles
	prometheusCRD.Spec.Replicas = proto.Int32(2)
	_, err = framework.UpdatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD)
	if err != nil {
		t.Fatal(err)
	}

	// Assert labels preserved
	for _, rConf := range resourceConfigs {
		res, err := rConf.get()
		if err != nil {
			t.Fatal(err)
		}

		labels := res.GetLabels()
		if !containsValues(labels, updatedLabels) {
			t.Errorf("%s: labels do not contain updated labels, found: %q, should contain: %q", rConf.name, labels, updatedLabels)
		}

		annotations := res.GetAnnotations()
		if !containsValues(annotations, updatedAnnotations) {
			t.Fatalf("%s: annotations do not contain updated annotations, found: %q, should contain: %q", rConf.name, annotations, updatedAnnotations)
		}
	}

	// Cleanup
	if err := framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, name); err != nil {
		t.Fatal(err)
	}
}

func asService(t *testing.T, object metav1.Object) *v1.Service {
	svc, ok := object.(*v1.Service)
	if !ok {
		t.Fatalf("expected service got %T", object)
	}
	return svc
}

func asEndpoints(t *testing.T, object metav1.Object) *v1.Endpoints {
	endpoints, ok := object.(*v1.Endpoints)
	if !ok {
		t.Fatalf("expected endpoints got %T", object)
	}
	return endpoints
}

func asStatefulSet(t *testing.T, object metav1.Object) *appsv1.StatefulSet {
	sset, ok := object.(*appsv1.StatefulSet)
	if !ok {
		t.Fatalf("expected stateful set got %T", object)
	}
	return sset
}

func asSecret(t *testing.T, object metav1.Object) *v1.Secret {
	sec, ok := object.(*v1.Secret)
	if !ok {
		t.Fatalf("expected secret set got %T", object)
	}
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
	if err != nil {
		t.Fatal(err)
	}

	configMapName := fmt.Sprintf("prometheus-%v-rulefiles-0", p.Name)

	_, err = framework.WaitForConfigMapExist(context.Background(), ns, configMapName)
	if err != nil {
		t.Fatal(err)
	}

	// Waits for Prometheus pods to vanish
	err = framework.DeletePrometheusAndWaitUntilGone(context.Background(), ns, p.Name)
	if err != nil {
		t.Fatal(err)
	}

	err = framework.WaitForConfigMapNotExist(context.Background(), ns, configMapName)
	if err != nil {
		t.Fatal(err)
	}
}

func testPromDiscovery(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	prometheusName := "test"
	group := "servicediscovery-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{}); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	_, err = framework.KubeClient.CoreV1().Secrets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
	if err != nil {
		t.Fatal("Generated Secret could not be retrieved: ", err)
	}

	err = framework.WaitForDiscoveryWorking(context.Background(), ns, svc.Name, prometheusName)
	if err != nil {
		t.Fatal(errors.Wrap(err, "validating Prometheus target discovery failed"))
	}
}

func testPromSharedResourcesReconciliation(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	s := framework.MakeBasicServiceMonitor("reconcile-test")
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Creating ServiceMonitor failed: %v", err)
	}

	// Create 2 Prometheus different Prometheus instances that watch the service monitor created above.
	for _, prometheusName := range []string{"test", "test2"} {
		p := framework.MakeBasicPrometheus(ns, prometheusName, "reconcile-test", 1)
		_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
		if err != nil {
			t.Fatal(err)
		}

		svc := framework.MakePrometheusService(prometheusName, fmt.Sprintf("reconcile-%s", prometheusName), v1.ServiceTypeClusterIP)
		if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
			t.Fatal(err)
		} else {
			testCtx.AddFinalizerFn(finalizerFn)
		}

		_, err = framework.KubeClient.CoreV1().Secrets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Generated Secret could not be retrieved for %s: %v", prometheusName, err)
		}

		err = framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 1)
		if err != nil {
			t.Fatalf("Validating Prometheus active targets failed for %s: %v", prometheusName, err)
		}
	}

	if err := framework.MonClientV1.ServiceMonitors(ns).Delete(context.Background(), "reconcile-test", metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Deleting ServiceMonitor failed: %v", err)
	}

	// Delete the service monitors and check that both Prometheus instances are updated.
	for _, prometheusName := range []string{"test", "test2"} {
		svc := framework.MakePrometheusService(prometheusName, fmt.Sprintf("reconcile-%s", prometheusName), v1.ServiceTypeClusterIP)

		if err := framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 0); err != nil {
			t.Fatalf("Validating Prometheus active targets failed for %s: %v", prometheusName, err)
		}
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
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{}); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	shards := int32(2)
	p.Spec.Shards = &shards
	_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

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
		stdout, _, err := framework.ExecWithOptions(testFramework.ExecOptions{
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
		if err != nil {
			t.Fatalf("Failed to read config from pod %q: %v", p.pod, err)
		}
		if !strings.Contains(stdout, p.expectedShardConfigSnippet) {
			t.Fatalf("Expected shard config to be present for %v but not found in config:\n\n%s\n\nexpected to find:\n\n%s", p.pod, stdout, p.expectedShardConfigSnippet)
		}
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
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{}); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	p, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	shards := int32(2)
	p.Spec.Shards = &shards
	p, err = framework.UpdatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}

	_, err = framework.KubeClient.AppsV1().StatefulSets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", p.Name), metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = framework.KubeClient.AppsV1().StatefulSets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s-shard-1", p.Name), metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	shards = int32(1)
	p.Spec.Shards = &shards
	p, err = framework.UpdatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}

	_, err = framework.KubeClient.AppsV1().StatefulSets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", p.Name), metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	err = wait.Poll(time.Second, 1*time.Minute, func() (bool, error) {
		_, err = framework.KubeClient.AppsV1().StatefulSets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s-shard-1", p.Name), metav1.GetOptions{})
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
	_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
	if err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Creating ServiceMonitor failed: %v", err)
	}

	_, err = framework.KubeClient.CoreV1().Secrets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Generated Secret could not be retrieved: %v", err)
	}

	if _, err := framework.CreateAlertmanagerAndWaitUntilReady(context.Background(), ns, framework.MakeBasicAlertmanager(alertmanagerName, 3)); err != nil {
		t.Fatal(err)
	}

	if _, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, amsvc); err != nil {
		t.Fatal(errors.Wrap(err, "creating Alertmanager service failed"))
	}

	err = wait.Poll(time.Second, 18*time.Minute, isAlertmanagerDiscoveryWorking(context.Background(), ns, svc.Name, alertmanagerName))
	if err != nil {
		t.Fatal(errors.Wrap(err, "validating Prometheus Alertmanager discovery failed"))
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

	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, basicPrometheus); err != nil {
		t.Fatal("Creating prometheus failed: ", err)
	}

	if _, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, service); err != nil {
		t.Fatal("Creating prometheus service failed: ", err)
	}

	ProxyGet := framework.KubeClient.CoreV1().Services(ns).ProxyGet
	request := ProxyGet("", service.Name, "web", "/metrics", make(map[string]string))
	_, err := request.DoRaw(context.Background())
	if err != nil {
		t.Fatal(err)
	}
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
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), sm, metav1.CreateOptions{}); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	_, err := framework.KubeClient.CoreV1().Secrets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
	if err != nil {
		t.Fatal("Generated Secret could not be retrieved: ", err)
	}

	err = framework.WaitForDiscoveryWorking(context.Background(), ns, svc.Name, prometheusName)
	if err != nil {
		t.Fatal(errors.Wrap(err, "validating Prometheus target discovery failed"))
	}
}

func testPromOpMatchPromAndServMonInDiffNSs(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	prometheusNSName := framework.CreateNamespace(context.Background(), t, testCtx)
	serviceMonitorNSName := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, prometheusNSName)

	if err := framework.AddLabelsToNamespace(
		context.Background(),
		serviceMonitorNSName,
		map[string]string{"team": "frontend"},
	); err != nil {
		t.Fatal(err)
	}

	group := "sample-app"

	prometheusJobName := serviceMonitorNSName + "/" + group

	prometheusName := "test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	s := framework.MakeBasicServiceMonitor(group)

	if _, err := framework.MonClientV1.ServiceMonitors(serviceMonitorNSName).Create(context.Background(), s, metav1.CreateOptions{}); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	p := framework.MakeBasicPrometheus(prometheusNSName, prometheusName, group, 1)
	p.Spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"team": "frontend",
		},
	}
	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), prometheusNSName, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), prometheusNSName, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	resp, err := framework.PrometheusSVCGetRequest(context.Background(), prometheusNSName, svc.Name, "http", "/api/v1/status/config", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Count(string(resp), prometheusJobName) != 1 {
		t.Fatalf("expected Prometheus operator to configure Prometheus in ns '%v' to scrape the service monitor in ns '%v'", prometheusNSName, serviceMonitorNSName)
	}
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
	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prom); err != nil {
		t.Fatal("Creating prometheus failed: ", err)
	}

	promSvc := framework.MakePrometheusService(prom.Name, "test-group", v1.ServiceTypeClusterIP)
	if _, err := framework.KubeClient.CoreV1().Services(ns).Create(context.Background(), promSvc, metav1.CreateOptions{}); err != nil {
		t.Fatal("Creating prometheus service failed: ", err)
	}

	svcMon := framework.MakeBasicServiceMonitor("test-group")
	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), svcMon, metav1.CreateOptions{}); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	qryDep, err := testFramework.MakeDeployment("../../example/thanos/query-deployment.yaml")
	if err != nil {
		t.Fatal("Making thanos query deployment failed: ", err)
	}
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
	if err := framework.CreateDeployment(context.Background(), ns, qryDep); err != nil {
		t.Fatal("Creating Thanos query deployment failed: ", err)
	}

	qrySvc := framework.MakeThanosQuerierService(qryDep.Name)
	if _, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, qrySvc); err != nil {
		t.Fatal("Creating Thanos query service failed: ", err)
	}

	err = wait.Poll(5*time.Second, 5*time.Minute, func() (bool, error) {
		proxyGet := framework.KubeClient.CoreV1().Services(ns).ProxyGet
		request := proxyGet("http", qrySvc.Name, "http-query", "/api/v1/query", map[string]string{
			"query": "prometheus_build_info",
			"dedup": "false",
		})
		b, err := request.DoRaw(context.Background())
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
	if err != nil {
		t.Fatal("Failed to get correct result from Thanos query: ", err)
	}
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
				sm.Spec.Endpoints[0].BearerTokenSecret = v1.SecretKeySelector{
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

			maptest := make(map[string]string)
			maptest["tc"] = ns
			prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)
			prometheusCRD.Spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
				MatchLabels: maptest,
			}
			prometheusCRD.Spec.ScrapeInterval = "1s"

			if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD); err != nil {
				t.Fatal(err)
			}
			testNamespace := framework.CreateNamespace(context.Background(), t, testCtx)

			err := framework.AddLabelsToNamespace(context.Background(), testNamespace, maptest)
			if err != nil {
				t.Fatal(err)
			}

			simple, err := testFramework.MakeDeployment("../../test/framework/resources/basic-auth-app-deployment.yaml")
			if err != nil {
				t.Fatal(err)
			}

			if err := framework.CreateDeployment(context.Background(), testNamespace, simple); err != nil {
				t.Fatal("Creating simple basic auth app failed: ", err)
			}

			authSecret := test.secret
			if _, err := framework.KubeClient.CoreV1().Secrets(testNamespace).Create(context.Background(), authSecret, metav1.CreateOptions{}); err != nil {
				t.Fatal(err)
			}

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
			if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), testNamespace, svc); err != nil {
				t.Fatal(err)
			} else {
				testCtx.AddFinalizerFn(finalizerFn)
			}

			if _, err := framework.MonClientV1.ServiceMonitors(testNamespace).Create(context.Background(), sm, metav1.CreateOptions{}); err != nil {
				t.Fatal("Creating ServiceMonitor failed: ", err)
			}

			if err := framework.WaitForHealthyTargets(context.Background(), ns, "prometheus-operated", 1); err != nil {
				t.Fatal(err)
			}
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
			if err != nil {
				t.Fatal(err)
			}
		}

		// Prometheus Operator only watches single namespace mainNS, not arbitraryNS.
		_, err := framework.CreatePrometheusOperator(context.Background(), operatorNS, *opImage, []string{mainNS}, nil, nil, nil, false, true)
		if err != nil {
			t.Fatal(err)
		}

		ruleDef := []struct {
			NSName    string
			AlertName string
		}{{arbitraryNS, secondAlertName}, {mainNS, firtAlertName}}

		for _, r := range ruleDef {
			_, err := framework.MakeAndCreateFiringRule(context.Background(), r.NSName, name, r.AlertName)
			if err != nil {
				t.Fatal(err)
			}
		}

		p := framework.MakeBasicPrometheus(mainNS, name, name, 1)
		p.Spec.RuleNamespaceSelector = &metav1.LabelSelector{
			MatchLabels: prometheusNamespaceSelector,
		}
		p.Spec.EvaluationInterval = "1s"
		p, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), mainNS, p)
		if err != nil {
			t.Fatal(err)
		}

		pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
		if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), mainNS, pSVC); err != nil {
			t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
		} else {
			testCtx.AddFinalizerFn(finalizerFn)
		}

		err = framework.WaitForPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, firtAlertName)
		if err != nil {
			t.Fatal(err)
		}

		firing, err := framework.CheckPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, secondAlertName)
		if err != nil && !strings.Contains(err.Error(), "expected 1 query result but got 0") {
			t.Fatal(err)
		}

		if firing {
			t.Fatalf("expected alert %q not to fire", secondAlertName)
		}
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
			if err != nil {
				t.Fatal(err)
			}
		}

		// Prometheus Operator only watches prometheusNS and ruleNS, not arbitraryNS.
		_, err := framework.CreatePrometheusOperator(context.Background(), operatorNS, *opImage, []string{prometheusNS, ruleNS}, nil, nil, nil, false, true)
		if err != nil {
			t.Fatal(err)
		}

		ruleDef := []struct {
			NSName    string
			AlertName string
		}{{arbitraryNS, secondAlertName}, {ruleNS, firtAlertName}}

		for _, r := range ruleDef {
			_, err := framework.MakeAndCreateFiringRule(context.Background(), r.NSName, name, r.AlertName)
			if err != nil {
				t.Fatal(err)
			}
		}

		p := framework.MakeBasicPrometheus(prometheusNS, name, name, 1)
		p.Spec.RuleNamespaceSelector = &metav1.LabelSelector{
			MatchLabels: prometheusNamespaceSelector,
		}
		p.Spec.EvaluationInterval = "1s"
		p, err = framework.CreatePrometheusAndWaitUntilReady(context.Background(), prometheusNS, p)
		if err != nil {
			t.Fatal(err)
		}

		pSVC := framework.MakePrometheusService(p.Name, "not-relevant", v1.ServiceTypeClusterIP)
		if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), prometheusNS, pSVC); err != nil {
			t.Fatal(errors.Wrap(err, "creating Prometheus service failed"))
		} else {
			testCtx.AddFinalizerFn(finalizerFn)
		}

		err = framework.WaitForPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, firtAlertName)
		if err != nil {
			t.Fatal(err)
		}

		firing, err := framework.CheckPrometheusFiringAlert(context.Background(), p.Namespace, pSVC.Name, secondAlertName)
		if err != nil && !strings.Contains(err.Error(), "expected 1 query result but got 0") {
			t.Fatal(err)
		}

		if firing {
			t.Fatalf("expected alert %q not to fire", secondAlertName)
		}
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
				BearerTokenSecret: v1.SecretKeySelector{
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
						InsecureSkipVerify: true,
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
						InsecureSkipVerify: true,
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
			cert, err := ioutil.ReadFile("../../test/instrumented-sample-app/certs/cert.pem")
			if err != nil {
				t.Fatalf("failed to load cert.pem: %v", err)
			}

			key, err := ioutil.ReadFile("../../test/instrumented-sample-app/certs/key.pem")
			if err != nil {
				t.Fatalf("failed to load key.pem: %v", err)
			}

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

			if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), tlsCertsSecret, metav1.CreateOptions{}); err != nil {
				t.Fatal(err)
			}

			tlsCertsConfigMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Data: map[string]string{
					"cert.pem": string(cert),
				},
			}

			if _, err := framework.KubeClient.CoreV1().ConfigMaps(ns).Create(context.Background(), tlsCertsConfigMap, metav1.CreateOptions{}); err != nil {
				t.Fatal(err)
			}

			s := framework.MakeBasicServiceMonitor(name)
			s.Spec.Endpoints[0] = test.endpoint
			if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{}); err != nil {
				t.Fatal("creating ServiceMonitor failed: ", err)
			}

			prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)
			prometheusCRD.Namespace = ns
			prometheusCRD.Spec.ArbitraryFSAccessThroughSMs = test.arbitraryFSAccessThroughSMsConfig

			if strings.HasSuffix(test.name, "-file") {
				mountTLSFiles(prometheusCRD, name)
			}

			if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD); err != nil {
				t.Fatal(err)
			}

			svc := framework.MakePrometheusService(prometheusCRD.Name, name, v1.ServiceTypeClusterIP)
			if _, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
				t.Fatal(err)
			}

			if test.expectTargets {
				if err := framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 1); err != nil {
					t.Fatal(err)
				}

				return
			}

			// Make sure Prometheus has enough time to reload.
			time.Sleep(2 * time.Minute)
			if err := framework.WaitForActiveTargets(context.Background(), ns, svc.Name, 0); err != nil {
				t.Fatal(err)
			}
		})
	}

}

// mountTLSFiles is a helper to manually mount TLS certificate files
// into the prometheus container
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

	cert, err := ioutil.ReadFile("../../test/instrumented-sample-app/certs/cert.pem")
	if err != nil {
		t.Fatalf("failed to load cert.pem: %v", err)
	}

	key, err := ioutil.ReadFile("../../test/instrumented-sample-app/certs/key.pem")
	if err != nil {
		t.Fatalf("failed to load key.pem: %v", err)
	}

	tlsCertsSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string][]byte{
			"cert.pem": cert,
			"key.pem":  key,
		},
	}

	if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), tlsCertsSecret, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	simple, err := testFramework.MakeDeployment("../../test/framework/resources/basic-auth-app-deployment.yaml")
	if err != nil {
		t.Fatal(err)
	}

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

	if err := framework.CreateDeployment(context.Background(), ns, simple); err != nil {
		t.Fatal("Creating simple basic auth app failed: ", err)
	}

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

	if _, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
		t.Fatal(err)
	}

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
					InsecureSkipVerify: true,
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

	if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), sm, metav1.CreateOptions{}); err != nil {
		t.Fatal("creating ServiceMonitor failed: ", err)
	}

	prometheusCRD := framework.MakeBasicPrometheus(ns, name, name, 1)

	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prometheusCRD); err != nil {
		t.Fatal(err)
	}

	promSVC := framework.MakePrometheusService(prometheusCRD.Name, name, v1.ServiceTypeClusterIP)

	if _, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, promSVC); err != nil {
		t.Fatal(err)
	}

	//
	// Check for proper scraping.
	//

	if err := framework.WaitForHealthyTargets(context.Background(), ns, promSVC.Name, 1); err != nil {
		t.Fatal(err)
	}

	// TODO: Do a poll instead, should speed up things.
	time.Sleep(30 * time.Second)

	response, err := framework.PrometheusSVCGetRequest(context.Background(), ns, promSVC.Name, "http", "/api/v1/query", map[string]string{"query": fmt.Sprintf(`up{job="%v",endpoint="%v"}`, name, sm.Spec.Endpoints[0].Port)})
	if err != nil {
		t.Fatal(err)
	}

	q := testFramework.PrometheusQueryAPIResponse{}
	if err := json.NewDecoder(bytes.NewBuffer(response)).Decode(&q); err != nil {
		t.Fatal(err)
	}

	if q.Status != "success" {
		t.Fatalf("expected query status to be 'success' but got %v", q.Status)
	}

	if q.Data.Result[0].Value[1] != "1" {
		t.Fatalf("expected query result to be '1' but got %v", q.Data.Result[0].Value[1])
	}
}

func testPromStaticProbe(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	blackboxExporterName := "blackbox-exporter"
	if err := framework.CreateBlackBoxExporterAndWaitUntilReady(context.Background(), ns, blackboxExporterName); err != nil {
		t.Fatal("Creating blackbox exporter failed: ", err)
	}

	blackboxSvc := framework.MakeBlackBoxExporterService(ns, blackboxExporterName)
	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, blackboxSvc); err != nil {
		t.Fatal("creating blackbox exporter service failed ", err)
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	prometheusName := "test"
	group := "probe-test"
	svc := framework.MakePrometheusService(prometheusName, group, v1.ServiceTypeClusterIP)

	proberURL := blackboxExporterName + ":9115"
	targets := []string{svc.Name + ":9090"}

	probe := framework.MakeBasicStaticProbe(group, proberURL, targets)
	if _, err := framework.MonClientV1.Probes(ns).Create(context.Background(), probe, metav1.CreateOptions{}); err != nil {
		t.Fatal("Creating Probe failed: ", err)
	}

	p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
	p.Spec.ProbeSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"group": group,
		},
	}
	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p); err != nil {
		t.Fatal(err)
	}

	if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
		t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
	} else {
		testCtx.AddFinalizerFn(finalizerFn)
	}

	expectedURL := url.URL{Host: proberURL, Scheme: "http", Path: "/probe"}
	q := expectedURL.Query()
	q.Set("module", "http_2xx")
	q.Set("target", targets[0])
	expectedURL.RawQuery = q.Encode()

	if err := wait.Poll(time.Second, time.Minute*5, func() (bool, error) {
		activeTargets, err := framework.GetActiveTargets(context.Background(), ns, svc.Name)
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
	}); err != nil {
		t.Fatal("waiting for static probe targets timed out.")
	}
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
				TLSConfig: &monitoringv1.PodMetricsEndpointTLSConfig{
					SafeTLSConfig: monitoringv1.SafeTLSConfig{
						InsecureSkipVerify: true,
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
				Path: "/",
			},
		},
		{
			name: "tls-configmap",
			endpoint: monitoringv1.PodMetricsEndpoint{
				Port:   "mtls",
				Scheme: "https",
				TLSConfig: &monitoringv1.PodMetricsEndpointTLSConfig{
					SafeTLSConfig: monitoringv1.SafeTLSConfig{
						InsecureSkipVerify: true,
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
			cert, err := ioutil.ReadFile("../../test/instrumented-sample-app/certs/cert.pem")
			if err != nil {
				t.Fatalf("failed to load cert.pem: %v", err)
			}

			key, err := ioutil.ReadFile("../../test/instrumented-sample-app/certs/key.pem")
			if err != nil {
				t.Fatalf("failed to load key.pem: %v", err)
			}

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

			if _, err := framework.KubeClient.CoreV1().Secrets(ns).Create(context.Background(), secret, metav1.CreateOptions{}); err != nil {
				t.Fatal(err)
			}

			tlsCertsConfigMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Data: map[string]string{
					"cert.pem": string(cert),
				},
			}

			if _, err := framework.KubeClient.CoreV1().ConfigMaps(ns).Create(context.Background(), tlsCertsConfigMap, metav1.CreateOptions{}); err != nil {
				t.Fatal(err)
			}

			prom := framework.MakeBasicPrometheus(ns, name, name, 1)
			prom.Namespace = ns

			if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prom); err != nil {
				t.Fatal(err)
			}

			simple, err := testFramework.MakeDeployment("../../test/framework/resources/basic-auth-app-deployment.yaml")
			if err != nil {
				t.Fatal(err)
			}

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

			if err := framework.CreateDeployment(context.Background(), ns, simple); err != nil {
				t.Fatal("failed to create simple basic auth app: ", err)
			}

			pm := framework.MakeBasicPodMonitor(name)
			pm.Spec.PodMetricsEndpoints[0] = test.endpoint

			if _, err := framework.MonClientV1.PodMonitors(ns).Create(context.Background(), pm, metav1.CreateOptions{}); err != nil {
				t.Fatal("failed to create PodMonitor: ", err)
			}

			if err := framework.WaitForHealthyTargets(context.Background(), ns, "prometheus-operated", 1); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func testPromWebTLS(t *testing.T) {
	t.Parallel()
	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	ns := framework.CreateNamespace(context.Background(), t, testCtx)
	framework.SetupPrometheusRBAC(context.Background(), t, testCtx, ns)

	host := fmt.Sprintf("%s.%s.svc", "basic-prometheus", ns)
	certBytes, keyBytes, err := certutil.GenerateSelfSignedCertKey(host, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	kubeClient := framework.KubeClient
	if err := framework.CreateSecretWithCert(context.Background(), certBytes, keyBytes, ns, "web-tls"); err != nil {
		t.Fatal(err)
	}

	prom := framework.MakeBasicPrometheus(ns, "basic-prometheus", "test-group", 1)
	prom.Spec.Web = &monitoringv1.WebSpec{
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
	}
	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prom); err != nil {
		t.Fatal("Creating prometheus failed: ", err)
	}

	promPods, err := kubeClient.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(promPods.Items) == 0 {
		t.Fatalf("No prometheus pods found in namespace %s", ns)
	}

	cfg := framework.RestConfig
	podName := promPods.Items[0].Name
	if err := testFramework.StartPortForward(cfg, "https", podName, ns, "9090"); err != nil {
		return
	}

	// The prometheus certificate is issued to <pod>.<namespace>.svc,
	// but port-forwarding is done through localhost.
	// This is why we use an http client which skips the TLS verification.
	// In the test we will verify the TLS certificate manually to make sure
	// the prometheus instance is configured properly.
	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := httpClient.Get("https://localhost:9090")
	if err != nil {
		t.Fatal(err)
	}

	receivedCertBytes, err := certutil.EncodeCertificates(resp.TLS.PeerCertificates...)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(receivedCertBytes, certBytes) {
		t.Fatal("Certificate received from prometheus instance does not match the one which is configured")
	}
}

func testPromMinReadySeconds(t *testing.T) {
	runFeatureGatedTests(t)
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
	if err != nil {
		t.Fatal("Creating prometheus failed: ", err)
	}

	promSS, err := kubeClient.AppsV1().StatefulSets(ns).Get(context.Background(), "prometheus-basic-prometheus", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if promSS.Spec.MinReadySeconds != int32(setMinReadySecondsInitial) {
		t.Fatalf("expected MinReadySeconds to be %d but got %d", setMinReadySecondsInitial, promSS.Spec.MinReadySeconds)
	}

	var updated uint32 = 10
	var got int32
	prom.Spec.MinReadySeconds = &updated
	if _, err = framework.UpdatePrometheusAndWaitUntilReady(context.Background(), ns, prom); err != nil {
		t.Fatal("Updating prometheus failed: ", err)
	}

	err = wait.Poll(time.Second, time.Minute*5, func() (bool, error) {
		promSS, err := kubeClient.AppsV1().StatefulSets(ns).Get(context.Background(), "prometheus-basic-prometheus", metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if promSS.Spec.MinReadySeconds != int32(updated) {
			got = promSS.Spec.MinReadySeconds
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		t.Fatalf("expected MinReadySeconds to be %d but got %d", updated, got)
	}
}

// testPromEnforcedNamespaceLabel checks that the enforcedNamespaceLabel field
// is honored even if a user tries to bypass the enforcement.
func testPromEnforcedNamespaceLabel(t *testing.T) {
	t.Parallel()

	for i, tc := range []struct {
		relabelConfigs       []*monitoringv1.RelabelConfig
		metricRelabelConfigs []*monitoringv1.RelabelConfig
	}{
		{
			// override label using the labeldrop action.
			relabelConfigs: []*monitoringv1.RelabelConfig{
				{
					Regex:  "namespace",
					Action: "labeldrop",
				},
			},
			metricRelabelConfigs: []*monitoringv1.RelabelConfig{
				{
					Regex:  "namespace",
					Action: "labeldrop",
				},
			},
		},
		{
			// override label using the replace action.
			relabelConfigs: []*monitoringv1.RelabelConfig{
				{
					TargetLabel: "namespace",
					Replacement: "ns1",
				},
			},
			metricRelabelConfigs: []*monitoringv1.RelabelConfig{
				{
					TargetLabel: "namespace",
					Replacement: "ns1",
				},
			},
		},
		{
			// override label using the labelmap action.
			relabelConfigs: []*monitoringv1.RelabelConfig{
				{
					TargetLabel: "temp_namespace",
					Replacement: "ns1",
				},
			},
			metricRelabelConfigs: []*monitoringv1.RelabelConfig{
				{
					Action:      "labelmap",
					Regex:       "temp_namespace",
					Replacement: "namespace",
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
			if _, err := framework.MonClientV1.ServiceMonitors(ns).Create(context.Background(), s, metav1.CreateOptions{}); err != nil {
				t.Fatal("Creating ServiceMonitor failed: ", err)
			}

			p := framework.MakeBasicPrometheus(ns, prometheusName, group, 1)
			p.Spec.EnforcedNamespaceLabel = "namespace"
			_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p)
			if err != nil {
				t.Fatal(err)
			}

			if finalizerFn, err := framework.CreateServiceAndWaitUntilReady(context.Background(), ns, svc); err != nil {
				t.Fatal(errors.Wrap(err, "creating prometheus service failed"))
			} else {
				ctx.AddFinalizerFn(finalizerFn)
			}

			_, err = framework.KubeClient.CoreV1().Secrets(ns).Get(context.Background(), fmt.Sprintf("prometheus-%s", prometheusName), metav1.GetOptions{})
			if err != nil {
				t.Fatal("Generated Secret could not be retrieved: ", err)
			}

			err = framework.WaitForDiscoveryWorking(context.Background(), ns, svc.Name, prometheusName)
			if err != nil {
				t.Fatal(errors.Wrap(err, "validating Prometheus target discovery failed"))
			}

			// Check that the namespace label is enforced to the correct value.
			var (
				loopErr        error
				namespaceLabel string
			)

			err = wait.Poll(5*time.Second, 1*time.Minute, func() (bool, error) {
				loopErr = nil
				res, err := framework.PrometheusQuery(ns, svc.Name, "http", "prometheus_build_info")
				if err != nil {
					loopErr = errors.Wrap(err, "failed to query Prometheus")
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

			if err != nil {
				t.Fatalf("%v: %v", err, loopErr)
			}

			if namespaceLabel != ns {
				t.Fatalf("expecting 'namespace' label value to be %q but got %q instead", ns, namespaceLabel)
			}
		})
	}
}

func testPrometheusCRDValidation(t *testing.T) {
	t.Parallel()
	name := "test"
	replicas := int32(1)
	commonFields := monitoringv1.CommonPrometheusFields{
		Replicas:           &replicas,
		Version:            operator.DefaultPrometheusVersion,
		ServiceAccountName: "prometheus",
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("400Mi"),
			},
		},
	}

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
				CommonPrometheusFields: commonFields,
				RetentionSize:          "0",
			},
		},
		{
			name: "legacy-unit",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: commonFields,
				RetentionSize:          "1.5GB",
			},
		},
		{
			name: "iec-unit",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: commonFields,
				RetentionSize:          "100MiB",
			},
		},
		{
			name: "legacy-missing-symbol",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: commonFields,
				RetentionSize:          "10M",
			},
			expectedError: true,
		},
		{
			name: "legacy-missing-unit",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: commonFields,
				RetentionSize:          "1000",
			},
			expectedError: true,
		},
		{
			name: "iec-missing-symbol",
			prometheusSpec: monitoringv1.PrometheusSpec{
				CommonPrometheusFields: commonFields,
				RetentionSize:          "15Gi",
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
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if !apierrors.IsInvalid(err) {
					t.Fatalf("expected Invalid error but got %v", err)
				}
				return
			}

			_, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, prom)
			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}
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
	if _, err := framework.CreatePrometheusAndWaitUntilReady(context.Background(), ns, p); err != nil {
		t.Fatal(err)
	}
}

func isAlertmanagerDiscoveryWorking(ctx context.Context, ns, promSVCName, alertmanagerName string) func() (bool, error) {
	return func() (bool, error) {
		pods, err := framework.KubeClient.CoreV1().Pods(ns).List(context.Background(), alertmanager.ListOptions(alertmanagerName))
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

func gzipConfig(buf *bytes.Buffer, conf []byte) error {
	w := gzip.NewWriter(buf)
	defer w.Close()
	if _, err := w.Write(conf); err != nil {
		return err
	}
	return nil
}
