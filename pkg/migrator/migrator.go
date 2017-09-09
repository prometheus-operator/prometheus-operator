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

package migrator

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"

	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
)

type migrationState struct {
	backedUp   bool
	tprDataDel bool
	tprRegDel  bool

	crdCreated bool
	finished   bool
}

const (
	NoMigration = iota
	TPR2CRD
)

type Migrator struct {
	eclient apiextensionsclient.Interface
	kclient kubernetes.Interface
	mclient monitoring.Interface

	migrationState *migrationState

	logger log.Logger

	cmPrometheus     string
	cmServiceMonitor string
	cmAlertmanager   string
}

func (m *Migrator) getMigration() (int, error) {
	mv, err := k8sutil.GetMinorVersion(m.kclient.Discovery())
	if err != nil {
		return NoMigration, err
	}
	if mv >= 7 {
		// See if the TPRs managed by this operator exist.
		tprClient := m.kclient.Extensions().ThirdPartyResources()
		_, err = tprClient.Get("prometheus."+v1alpha1.Group, metav1.GetOptions{})
		if err == nil {
			return TPR2CRD, nil
		}
		_, err = tprClient.Get("service-monitor."+v1alpha1.Group, metav1.GetOptions{})
		if err == nil {
			return TPR2CRD, nil
		}
		_, err = tprClient.Get("alertmanager."+v1alpha1.Group, metav1.GetOptions{})
		if err == nil {
			return TPR2CRD, nil
		}
	}

	return NoMigration, nil
}

func (m *Migrator) RunMigration() error {
	mg, err := m.getMigration()
	if err != nil {
		return err
	}

	switch mg {
	case TPR2CRD:
		err = m.migrateTPR2CRD()
		if err != nil {
			m.rollback()
			return err
		}
	}

	return nil
}

func NewMigrator(cfg *rest.Config, logger log.Logger) (*Migrator, error) {
	kclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating kubernetes client")
	}

	mclient, err := monitoring.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating monitoring client")
	}

	eclient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating apiextensions client")
	}

	configmapSuffix := strconv.FormatInt(time.Now().Unix(), 10)

	return &Migrator{
		migrationState:   &migrationState{},
		logger:           logger,
		cmPrometheus:     "crd-migrate-prometheus-" + configmapSuffix,
		cmServiceMonitor: "crd-migrate-servicemonitor-" + configmapSuffix,
		cmAlertmanager:   "crd-migrate-alertmanager-" + configmapSuffix,
		kclient:          kclient,
		mclient:          mclient,
		eclient:          eclient,
	}, nil
}

func (m *Migrator) migrateTPR2CRD() error {
	m.logger.Log("msg", "Performing TPR to CRD migration.")

	m.logger.Log("msg", "Backing up TPR objects.")
	err := m.backupObjects()
	if err != nil {
		return errors.Wrap(err, "backing up objects failed")
	}

	m.migrationState.backedUp = true

	m.logger.Log("msg", "Deleting TPR objects.")
	err = m.deleteData(
		m.newPrometheusV1alpha1DeleteFunc,
		m.newServiceMonitorV1alpha1DeleteFunc,
		m.newAlertmanagerV1alpha1DeleteFunc,
	)
	if err != nil {
		return errors.Wrap(err, "deleting old objects failed")
	}

	m.migrationState.tprDataDel = true

	m.logger.Log("msg", "Deleting TPRs.")
	err = m.deleteTPRs(
		k8sutil.NewPrometheusTPRDefinition(),
		k8sutil.NewServiceMonitorTPRDefinition(),
		k8sutil.NewAlertmanagerTPRDefinition(),
	)
	if err != nil {
		return errors.Wrap(err, "deleting TPRs failed")
	}

	m.migrationState.tprRegDel = true

	m.logger.Log("msg", "Creating CRDs.")
	err = m.createCRDs(
		k8sutil.NewPrometheusCustomResourceDefinition(),
		k8sutil.NewServiceMonitorCustomResourceDefinition(),
		k8sutil.NewAlertmanagerCustomResourceDefinition(),
	)
	if err != nil {
		return errors.Wrap(err, "creating CRDs failed")
	}

	m.migrationState.crdCreated = true

	m.logger.Log("msg", "Waiting for CRDs to be ready.")
	err = m.waitForCRDsReady()
	if err != nil {
		return errors.Wrap(err, "waiting for CRDs to be ready failed")
	}

	//time.Sleep(time.Minute * 6)

	m.logger.Log("msg", "Restoring backed up objects.")
	err = m.restoreObjects()
	if err != nil {
		return errors.Wrap(err, "failed to restore objects")
	}

	m.migrationState.finished = true

	return nil
}

func (m *Migrator) backupObjects() error {
	err := m.backupObjectsInConfigMap(m.cmPrometheus, m.mclient.MonitoringV1alpha1().Prometheuses(api.NamespaceAll).List)
	if err != nil {
		return errors.Wrap(err, "backing up Prometheus objects failed")
	}

	err = m.backupObjectsInConfigMap(m.cmServiceMonitor, m.mclient.MonitoringV1alpha1().ServiceMonitors(api.NamespaceAll).List)
	if err != nil {
		return errors.Wrap(err, "backing up ServiceMonitor objects failed")
	}

	err = m.backupObjectsInConfigMap(m.cmAlertmanager, m.mclient.MonitoringV1alpha1().Alertmanagers(api.NamespaceAll).List)
	if err != nil {
		return errors.Wrap(err, "backing up Alertmanager objects failed")
	}

	return nil
}

func (m *Migrator) restoreObjects() error {
	err := m.restorePrometheusesFromCM(m.cmPrometheus)
	if err != nil {
		return errors.Wrap(err, "restoring Prometheus objects from ConfigMap failed")
	}

	err = m.restoreServiceMonitorsFromCM(m.cmServiceMonitor)
	if err != nil {
		return errors.Wrap(err, "restoring Prometheus objects from ConfigMap failed")
	}

	err = m.restoreAlertmanagersFromCM(m.cmAlertmanager)
	if err != nil {
		return errors.Wrap(err, "restoring Prometheus objects from ConfigMap failed")
	}

	return nil
}

func (m *Migrator) newPrometheusV1alpha1DeleteFunc(namespace string) func(*metav1.DeleteOptions, metav1.ListOptions) error {
	return m.mclient.MonitoringV1alpha1().Prometheuses(namespace).DeleteCollection
}

func (m *Migrator) newServiceMonitorV1alpha1DeleteFunc(namespace string) func(*metav1.DeleteOptions, metav1.ListOptions) error {
	return m.mclient.MonitoringV1alpha1().ServiceMonitors(namespace).DeleteCollection
}

func (m *Migrator) newAlertmanagerV1alpha1DeleteFunc(namespace string) func(*metav1.DeleteOptions, metav1.ListOptions) error {
	return m.mclient.MonitoringV1alpha1().Alertmanagers(namespace).DeleteCollection
}

func (m *Migrator) newPrometheusV1DeleteFunc(namespace string) func(*metav1.DeleteOptions, metav1.ListOptions) error {
	return m.mclient.MonitoringV1().Prometheuses(namespace).DeleteCollection
}

func (m *Migrator) newServiceMonitorV1DeleteFunc(namespace string) func(*metav1.DeleteOptions, metav1.ListOptions) error {
	return m.mclient.MonitoringV1().ServiceMonitors(namespace).DeleteCollection
}

func (m *Migrator) newAlertmanagerV1DeleteFunc(namespace string) func(*metav1.DeleteOptions, metav1.ListOptions) error {
	return m.mclient.MonitoringV1().Alertmanagers(namespace).DeleteCollection
}

func (m *Migrator) createTPRs(tprs ...*extensionsv1beta1.ThirdPartyResource) error {
	tprClient := m.kclient.ExtensionsV1beta1().ThirdPartyResources()

	for _, tpr := range tprs {
		_, err := tprClient.Create(tpr)
		if err != nil {
			return errors.Wrapf(err, "failed to create %s TPR", tpr.Name)
		}
	}

	return nil
}

func (m *Migrator) deleteTPRs(tprs ...*extensionsv1beta1.ThirdPartyResource) error {
	tprClient := m.kclient.ExtensionsV1beta1().ThirdPartyResources()

	for _, tpr := range tprs {
		err := tprClient.Delete(tpr.Name, &metav1.DeleteOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to remove %s TPR", tpr.Name)
		}
	}

	return nil
}

func (m *Migrator) backupObjectsInConfigMap(configmapName string, listFunc func(opts metav1.ListOptions) (runtime.Object, error)) error {
	res, err := listFunc(metav1.ListOptions{})
	if err != nil {
		return err
	}

	data, err := json.Marshal(res)
	if err != nil {
		return err
	}

	_, err = m.kclient.CoreV1().ConfigMaps("default").Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configmapName,
		},
		Data: map[string]string{
			"backup": string(data),
		},
	})
	return err
}

func (m *Migrator) waitForCRDsReady() error {
	err := k8sutil.WaitForCRDReady(m.mclient.MonitoringV1().Prometheuses(api.NamespaceAll).List)
	if err != nil {
		return errors.Wrap(err, "waiting for Prometheus CRD to be ready failed")
	}

	err = k8sutil.WaitForCRDReady(m.mclient.MonitoringV1().ServiceMonitors(api.NamespaceAll).List)
	if err != nil {
		return errors.Wrap(err, "waiting for ServiceMonitor CRD to be ready failed")
	}

	err = k8sutil.WaitForCRDReady(m.mclient.MonitoringV1().Alertmanagers(api.NamespaceAll).List)
	if err != nil {
		return errors.Wrap(err, "waiting for Alertmanager CRD to be ready failed")
	}

	return nil
}

func (m *Migrator) deleteCRDs(crds ...*extensionsobj.CustomResourceDefinition) error {
	for _, crd := range crds {
		err := m.deleteCRD(crd)
		if err != nil {
			return errors.Wrapf(err, "deleting %s CRD failed", crd.Name)
		}
	}

	return nil
}

func (m *Migrator) deleteCRD(crd *extensionsobj.CustomResourceDefinition) error {
	crdClient := m.eclient.ApiextensionsV1beta1().CustomResourceDefinitions()

	return crdClient.Delete(crd.Name, &metav1.DeleteOptions{})
}

func (m *Migrator) createCRDs(crds ...*extensionsobj.CustomResourceDefinition) error {
	for _, crd := range crds {
		err := m.createCRD(crd)
		if err != nil {
			return errors.Wrapf(err, "creating %s CRD failed", crd.Name)
		}
	}

	return nil
}

func (m *Migrator) createCRD(crd *extensionsobj.CustomResourceDefinition) error {
	crdClient := m.eclient.ApiextensionsV1beta1().CustomResourceDefinitions()

	_, err := crdClient.Create(crd)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrapf(err, "creating %s CRD failed", crd.Spec.Names.Kind)
	}

	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		crdEst, err := crdClient.Get(crd.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crdEst.Status.Conditions {
			switch cond.Type {
			case extensionsobj.Established:
				if cond.Status == extensionsobj.ConditionTrue {
					return true, err
				}
			case extensionsobj.NamesAccepted:
				if cond.Status == extensionsobj.ConditionFalse {
					fmt.Printf("Name conflict: %v\n", cond.Reason)
				}
			}
		}
		return false, err
	})

	return nil
}

func (m *Migrator) readItemsFrom(configmapName string, processFunc func(item *unstructured.Unstructured) error) error {
	cm, err := m.kclient.CoreV1().ConfigMaps("default").Get(configmapName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "creating config map client")
	}

	tprdata := cm.Data["backup"]

	us := &unstructured.UnstructuredList{}
	err = json.Unmarshal([]byte(tprdata), us)
	if err != nil {
		return errors.Wrap(err, "unmarshalling into UnstructuredList")
	}

	for _, item := range us.Items {
		item.SetResourceVersion("")
		item.SetAPIVersion("monitoring.coreos.com/v1")
		err := processFunc(&item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) restorePrometheusesFromCM(configmapName string) error {
	return m.readItemsFrom(configmapName, func(item *unstructured.Unstructured) error {
		p, err := monitoringv1.PrometheusFromUnstructured(item)
		if err != nil {
			return err
		}

		_, err = m.mclient.MonitoringV1().Prometheuses(p.GetNamespace()).Create(p)
		if err != nil {
			return err
		}

		return nil
	})
}

func (m *Migrator) restoreServiceMonitorsFromCM(configmapName string) error {
	return m.readItemsFrom(configmapName, func(item *unstructured.Unstructured) error {
		s, err := monitoringv1.ServiceMonitorFromUnstructured(item)
		if err != nil {
			return err
		}

		_, err = m.mclient.MonitoringV1().ServiceMonitors(s.GetNamespace()).Create(s)
		if err != nil {
			return err
		}

		return nil
	})
}

func (m *Migrator) restoreAlertmanagersFromCM(configmapName string) error {
	return m.readItemsFrom(configmapName, func(item *unstructured.Unstructured) error {
		a, err := monitoringv1.AlertmanagerFromUnstructured(item)
		if err != nil {
			return err
		}

		_, err = m.mclient.MonitoringV1().Alertmanagers(a.GetNamespace()).Create(a)
		if err != nil {
			return err
		}

		return nil
	})
}

func (m *Migrator) deleteData(deleteFuncFactories ...func(namespace string) func(*metav1.DeleteOptions, metav1.ListOptions) error) error {
	nsList, err := m.kclient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to list namespaces")
	}

	for _, ns := range nsList.Items {
		for _, f := range deleteFuncFactories {
			err := f(ns.Name)(&metav1.DeleteOptions{}, metav1.ListOptions{})
			if err != nil {
				return errors.Wrapf(err, "failed to delete tprdata in namespace \"%s\"", ns.Name)
			}
		}
	}

	return nil
}

func (m *Migrator) rollback() error {
	m.logger.Log("msg", "Rolling back migration.")

	ms := m.migrationState
	if ms.finished {
		m.logger.Log("msg", "Deleting CRD objects.")
		err := m.deleteData(
			m.newPrometheusV1DeleteFunc,
			m.newServiceMonitorV1DeleteFunc,
			m.newAlertmanagerV1DeleteFunc,
		)
		if err != nil {
			return errors.Wrapf(err, "deleting the CRD data failed")
		}
	}

	if ms.crdCreated {
		m.logger.Log("msg", "Deleting CRDs.")
		err := m.deleteCRDs(
			k8sutil.NewPrometheusCustomResourceDefinition(),
			k8sutil.NewServiceMonitorCustomResourceDefinition(),
			k8sutil.NewAlertmanagerCustomResourceDefinition(),
		)
		if err != nil {
			return errors.Wrapf(err, "deleting the CRDs failed")
		}
	}

	if ms.tprRegDel {
		m.logger.Log("msg", "Creating TPRs.")
		err := m.createTPRs(
			k8sutil.NewPrometheusTPRDefinition(),
			k8sutil.NewServiceMonitorTPRDefinition(),
			k8sutil.NewAlertmanagerTPRDefinition(),
		)
		if err != nil {
			return errors.Wrapf(err, "recreating the TPRs failed")
		}

		err = m.waitForTPRsReady()
		if err != nil {
			return errors.Wrapf(err, "waiting for TPRs to be ready failed")
		}

	}

	if ms.tprDataDel {
		m.logger.Log("msg", "Restoring TPR objects from backup.")
		err := m.restoreV1alpha1ObjectsFromCM()
		if err != nil {
			return errors.Wrapf(err, "recreating the TPR data for from ConfigMaps failed")
		}
	}

	return nil
}

func (m *Migrator) restoreV1alpha1ObjectsFromCM() error {
	err := m.restorePrometheusesV1alpha1FromCM(m.cmPrometheus)
	if err != nil {
		return errors.Wrap(err, "failed to restore Prometheus objects from configmap")
	}

	err = m.restoreServiceMonitorsV1alpha1FromCM(m.cmServiceMonitor)
	if err != nil {
		return errors.Wrap(err, "failed to restore ServiceMonitor objects from configmap")
	}

	err = m.restoreAlertmanagersV1alpha1FromCM(m.cmAlertmanager)
	if err != nil {
		return errors.Wrap(err, "failed to restore Alertmanager objects from configmap")
	}

	return nil
}

func (m *Migrator) restorePrometheusesV1alpha1FromCM(configmapName string) error {
	return m.readItemsFrom(configmapName, func(item *unstructured.Unstructured) error {
		p, err := v1alpha1.PrometheusFromUnstructured(item)
		if err != nil {
			return err
		}

		_, err = m.mclient.MonitoringV1alpha1().Prometheuses(p.GetNamespace()).Create(p)
		if err != nil {
			return err
		}

		return nil
	})
}

func (m *Migrator) restoreServiceMonitorsV1alpha1FromCM(configmapName string) error {
	return m.readItemsFrom(configmapName, func(item *unstructured.Unstructured) error {
		s, err := v1alpha1.ServiceMonitorFromUnstructured(item)
		if err != nil {
			return err
		}

		_, err = m.mclient.MonitoringV1alpha1().ServiceMonitors(s.GetNamespace()).Create(s)
		if err != nil {
			return err
		}

		return nil
	})
}

func (m *Migrator) restoreAlertmanagersV1alpha1FromCM(configmapName string) error {
	return m.readItemsFrom(configmapName, func(item *unstructured.Unstructured) error {
		a, err := v1alpha1.AlertmanagerFromUnstructured(item)
		if err != nil {
			return err
		}

		_, err = m.mclient.MonitoringV1alpha1().Alertmanagers(a.GetNamespace()).Create(a)
		if err != nil {
			return err
		}

		return nil
	})
}

func (m *Migrator) waitForTPRsReady() error {
	err := k8sutil.WaitForCRDReady(m.mclient.MonitoringV1alpha1().Prometheuses(api.NamespaceAll).List)
	if err != nil {
		return errors.Wrap(err, "waiting for Prometheus TPR to be ready failed")
	}

	err = k8sutil.WaitForCRDReady(m.mclient.MonitoringV1alpha1().ServiceMonitors(api.NamespaceAll).List)
	if err != nil {
		return errors.Wrap(err, "waiting for ServiceMonitor TPR to be ready failed")
	}

	err = k8sutil.WaitForCRDReady(m.mclient.MonitoringV1alpha1().Alertmanagers(api.NamespaceAll).List)
	if err != nil {
		return errors.Wrap(err, "waiting for Alertmanager TPR to be ready failed")
	}

	return nil
}
