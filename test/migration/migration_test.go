// Copyright 2017 The prometheus-operator Authors
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

package migration

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/coreos/prometheus-operator/pkg/alertmanager"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	monitoringv1alpha1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/coreos/prometheus-operator/pkg/prometheus"
	operatorFramework "github.com/coreos/prometheus-operator/test/framework"
	"github.com/pkg/errors"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	crdc "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/tools/clientcmd"
)

var framework *operatorFramework.Framework
var kubeconfig, opImage, ns, ip *string

func TestMain(m *testing.M) {
	kubeconfig = flag.String("kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	opImage = flag.String("operator-image", "", "operator image, e.g. quay.io/coreos/prometheus-operator")
	ns = flag.String("namespace", "prometheus-operator-migration-tests", "migration test namespace")
	flag.Parse()

	os.Exit(m.Run())
}

func TestMigration(t *testing.T) {
	// Create all the clients required.
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		t.Fatal(err)
	}

	kclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatal(err)
	}

	mv, err := k8sutil.GetMinorVersion(kclient.Discovery())
	if err != nil {
		t.Fatal(err)
	}

	if mv < 7 {
		t.Skip("lower than 1.7 version")
		return
	}

	extClient, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		t.Fatal(err)
	}

	mclientV1alpha1, err := monitoringv1alpha1.NewForConfig(config)
	if err != nil {
		t.Fatal(err)
	}

	tprClient := kclient.ExtensionsV1beta1().ThirdPartyResources()
	crdClient := extClient.ApiextensionsV1beta1().CustomResourceDefinitions()

	if framework, err = operatorFramework.New(
		*ns,
		*kubeconfig,
		"quay.io/coreos/prometheus-operator:v0.11.1",
	); err != nil {
		log.Printf("failed to setup framework: %v\n", err)
		t.Fatal(err)
	}

	err = k8sutil.WaitForCRDReady(mclientV1alpha1.Prometheuses(api.NamespaceAll).List)
	if err != nil {
		t.Fatal(err)
	}

	err = k8sutil.WaitForCRDReady(mclientV1alpha1.ServiceMonitors(api.NamespaceAll).List)
	if err != nil {
		t.Fatal(err)
	}

	err = k8sutil.WaitForCRDReady(mclientV1alpha1.Alertmanagers(api.NamespaceAll).List)
	if err != nil {
		t.Fatal(err)
	}

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns2 := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns2, framework.KubeClient)

	// Launch the objects.
	name := "test"
	group := "tpr-migration-test"

	p := framework.MakeBasicPrometheusV1alpha1(ns2, name, name, 1)
	if _, err := mclientV1alpha1.Prometheuses(ns2).Create(p); err != nil {
		t.Fatal("Creating Prometheus failed: ", err)
	}

	if err := operatorFramework.WaitForPodsReady(kclient, ns2, time.Minute*3, 1, prometheus.ListOptions(p.Name)); err != nil {
		t.Fatal("Waiting for Prometheus pods to be ready failed: ", err)
	}

	if _, err := mclientV1alpha1.ServiceMonitors(ns2).Create(framework.MakeBasicServiceMonitorV1alpha1(group)); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	a := framework.MakeBasicAlertmanagerV1alpha1(name, 3)
	if _, err := mclientV1alpha1.Alertmanagers(ns2).Create(a); err != nil {
		t.Fatal("Creating Alertmanager failed: ", err)
	}

	amConfigSecretName := fmt.Sprintf("alertmanager-%s", a.Name)
	s, err := framework.AlertmanagerConfigSecret(amConfigSecretName)
	if err != nil {
		t.Fatal(errors.Wrap(err, fmt.Sprintf("making alertmanager config secret %v failed", amConfigSecretName)))
	}
	_, err = framework.KubeClient.CoreV1().Secrets(ns2).Create(s)
	if err != nil {
		t.Fatal(errors.Wrap(err, fmt.Sprintf("creating alertmanager config secret %v failed", s.Name)))
	}

	if err := operatorFramework.WaitForPodsReady(kclient, ns2, time.Minute*3, 3, alertmanager.ListOptions(a.Name)); err != nil {
		t.Fatal("Waiting for Alertmanager pods to be ready failed: ", err)
	}

	obj, err := mclientV1alpha1.Prometheuses(ns2).List(metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	proms := obj.(*monitoringv1alpha1.PrometheusList)

	// Get the objects.
	obj, err = mclientV1alpha1.Alertmanagers(ns2).List(metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	ams := obj.(*monitoringv1alpha1.AlertmanagerList)

	obj, err = mclientV1alpha1.ServiceMonitors(ns2).List(metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	sms := obj.(*monitoringv1alpha1.ServiceMonitorList)

	// Delete and launch new operator.
	if err := operatorFramework.DeleteDeployment(
		framework.KubeClient,
		framework.Namespace.Name,
		"prometheus-operator",
	); err != nil {
		t.Fatal(err)
	}

	if err := operatorFramework.WaitUntilDeploymentGone(
		framework.KubeClient,
		framework.Namespace.Name,
		"prometheus-operator",
		framework.DefaultTimeout,
	); err != nil {
		t.Fatal(err)
	}

	if err := framework.Setup(*opImage); err != nil {
		t.Fatal(err)
	}

	err = wait.Poll(time.Second*20, time.Minute*10, func() (bool, error) {
		return checkMigrationDone(ns2, framework, tprClient, crdClient, proms, sms, ams)
	})
	if err != nil {
		log.Println("Waiting for the migration to succeed failed: ", err)
		t.Fail()
	}

	if err := framework.Teardown(); err != nil {
		log.Println("Framework teardown failed: ", err)
		t.Fail()
	}
}

func checkMigrationDone(ns2 string, framework *operatorFramework.Framework, tprClient v1beta1.ThirdPartyResourceInterface, crdClient crdc.CustomResourceDefinitionInterface, proms *monitoringv1alpha1.PrometheusList, sms *monitoringv1alpha1.ServiceMonitorList, ams *monitoringv1alpha1.AlertmanagerList) (bool, error) {
	// Check if TPRs are deleted.
	_, err := tprClient.Get(k8sutil.NewPrometheusTPRDefinition().Name, metav1.GetOptions{})
	if err == nil {
		fmt.Println("Expected Prometheus TPR definition to be deleted, but it still exists.")
		return false, nil
	}
	if err != nil && !apierrors.IsNotFound(err) {
		return false, err
	}

	_, err = tprClient.Get(k8sutil.NewAlertmanagerTPRDefinition().Name, metav1.GetOptions{})
	if err == nil {
		fmt.Println("Expected Alertmanager TPR definition to be deleted, but it still exists.")
		return false, nil
	}
	if err != nil && !apierrors.IsNotFound(err) {
		return false, err
	}

	_, err = tprClient.Get(k8sutil.NewServiceMonitorTPRDefinition().Name, metav1.GetOptions{})
	if err == nil {
		fmt.Println("Expected ServiceMonitor TPR definition to be deleted, but it still exists.")
		return false, nil
	}
	if err != nil && !apierrors.IsNotFound(err) {
		return false, err
	}

	// Check if CRDs are created.
	_, err = crdClient.Get(k8sutil.NewPrometheusCustomResourceDefinition().Name, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		fmt.Println("Expected Prometheus CRD definition to be created, but it does not exists.")
		return false, nil
	}
	if err != nil {
		return false, err
	}

	_, err = crdClient.Get(k8sutil.NewAlertmanagerCustomResourceDefinition().Name, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		fmt.Println("Expected Alertmanager CRD definition to be created, but it does not exists.")
		return false, nil
	}
	if err != nil {
		return false, err
	}

	_, err = crdClient.Get(k8sutil.NewServiceMonitorCustomResourceDefinition().Name, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		fmt.Println("Expected ServiceMonitor CRD definition to be created, but it does not exists.")
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Compare old and new objects.
	obj, err := framework.MonClient.Prometheuses(ns2).List(metav1.ListOptions{})
	if err != nil {
		fmt.Printf("failed to list prometheus objects: \n", err)
		return false, nil
	}
	promsNew := obj.(*monitoringv1.PrometheusList)
	if len(promsNew.Items) != len(proms.Items) {
		fmt.Printf("expected %d prometheuses, got %d\n", len(proms.Items), len(promsNew.Items))
		return false, nil
	}

	for i, prom := range proms.Items {
		if promsNew.Items[i].GroupVersionKind().Version != monitoringv1.Version {
			return false, fmt.Errorf("expected version %f got %f", monitoringv1.Version, promsNew.Items[i].GroupVersionKind().Version)
		}

		oldb, err := json.Marshal(prom.Spec)
		if err != nil {
			return false, err
		}

		newb, err := json.Marshal(promsNew.Items[i].Spec)
		if err != nil {
			return false, err
		}

		if string(oldb) != string(newb) {
			return false, fmt.Errorf("The prometheus object changed %d", i)
		}
	}

	obj, err = framework.MonClient.Alertmanagers(ns2).List(metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	amsNew := obj.(*monitoringv1.AlertmanagerList)
	if len(amsNew.Items) != len(ams.Items) {
		fmt.Printf("expected %d ams, got %d\n", len(ams.Items), len(amsNew.Items))
		return false, nil
	}

	for i, am := range ams.Items {
		if amsNew.Items[i].GroupVersionKind().Version != monitoringv1.Version {
			return false, fmt.Errorf("expected version %f got %f", monitoringv1.Version, amsNew.Items[i].GroupVersionKind().Version)
		}

		oldb, err := json.Marshal(am.Spec)
		if err != nil {
			return false, err
		}

		newb, err := json.Marshal(amsNew.Items[i].Spec)
		if err != nil {
			return false, err
		}

		if string(oldb) != string(newb) {
			fmt.Errorf("The alertmanager object changed %d", i)
		}
	}

	obj, err = framework.MonClient.ServiceMonitors(ns2).List(metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	smsNew := obj.(*monitoringv1.ServiceMonitorList)
	if len(smsNew.Items) != len(sms.Items) {
		fmt.Printf("expected %d ams, got %d\n", len(sms.Items), len(smsNew.Items))
		return false, nil
	}

	for i, sm := range sms.Items {
		oldb, err := json.Marshal(sm.Spec)
		if err != nil {
			return false, err
		}

		newb, err := json.Marshal(smsNew.Items[i].Spec)
		if err != nil {
			return false, err
		}

		if string(oldb) != string(newb) {
			return false, fmt.Errorf("The servicemonitor object changed %d", i)
		}
	}

	return true, nil
}
