package migration

import (
	"flag"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	operatorFramework "github.com/coreos/prometheus-operator/test/framework"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
	tprClient := kclient.ExtensionsV1beta1().ThirdPartyResources()
	crdClient := extClient.ApiextensionsV1beta1().CustomResourceDefinitions()

	if framework, err = operatorFramework.New(
		*ns,
		*kubeconfig,
		"quay.io/coreos/prometheus-operator:v0.10.2",
	); err != nil {
		log.Printf("failed to setup framework: %v\n", err)
		t.Fatal(err)
	}

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns2 := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupPrometheusRBAC(t, ns2, framework.KubeClient)

	// Launch the objects.
	name := "test"
	group := "tpr-migration-test"

	prometheusTPR := framework.MakeBasicPrometheus(ns2, name, name, 1)
	prometheusTPR.Namespace = ns2
	if err := framework.CreatePrometheusAndWaitUntilReady(ns2, prometheusTPR); err != nil {
		t.Fatal(err)
	}

	s := framework.MakeBasicServiceMonitor(group)
	if _, err := framework.MonClient.ServiceMonitors(ns2).Create(s); err != nil {
		t.Fatal("Creating ServiceMonitor failed: ", err)
	}

	if err := framework.CreateAlertmanagerAndWaitUntilReady(ns2, framework.MakeBasicAlertmanager(name, 3)); err != nil {
		t.Fatal(err)
	}

	obj, err := framework.MonClient.Prometheuses(ns2).List(metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	proms := obj.(*v1alpha1.PrometheusList)

	// Get the objects.
	obj, err = framework.MonClient.Alertmanagers(ns2).List(metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	ams := obj.(*v1alpha1.AlertmanagerList)

	obj, err = framework.MonClient.ServiceMonitors(ns2).List(metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	sms := obj.(*v1alpha1.ServiceMonitorList)

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

	// Check if TPRs are deleted.
	tprList, err := tprClient.List(metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(tprList.Items) != 0 {
		t.Fatalf("expected 0 TPRs got %d", len(tprList.Items))
	}

	// Check if CRDs are created.
	crdList, err := crdClient.List(metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(crdList.Items) != 3 {
		t.Fatalf("expected 3 CRDs got %d", len(crdList.Items))
	}

	// Compare old and new objects.
	obj, err = framework.MonClient.Prometheuses(ns2).List(metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	promsNew := obj.(*v1alpha1.PrometheusList)
	if len(promsNew.Items) != len(proms.Items) {
		t.Fatalf("expected %d prometheuses, got %d", len(proms.Items), len(promsNew.Items))
	}

	for i, prom := range proms.Items {
		if !reflect.DeepEqual(prom.Spec, promsNew.Items[i].Spec) {
			t.Fatalf("The prometheus object changed %d", i)
		}
	}

	obj, err = framework.MonClient.Alertmanagers(ns2).List(metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	amsNew := obj.(*v1alpha1.AlertmanagerList)
	if len(amsNew.Items) != len(ams.Items) {
		t.Fatalf("expected %d ams, got %d", len(ams.Items), len(amsNew.Items))
	}

	for i, am := range ams.Items {
		if !reflect.DeepEqual(am.Spec, amsNew.Items[i].Spec) {
			t.Fatalf("The alertmanager object changed %d", i)
		}
	}

	obj, err = framework.MonClient.ServiceMonitors(ns2).List(metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	smsNew := obj.(*v1alpha1.ServiceMonitorList)
	if len(smsNew.Items) != len(sms.Items) {
		t.Fatalf("expected %d ams, got %d", len(sms.Items), len(smsNew.Items))
	}

	for i, sm := range sms.Items {
		if !reflect.DeepEqual(sm.Spec, smsNew.Items[i].Spec) {
			t.Fatalf("The servicemonitor object changed %d", i)
		}
	}

	if err := framework.Teardown(); err != nil {
		t.Fatal(err)
	}
}
