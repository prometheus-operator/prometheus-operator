package migrator

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"

	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	extensionsobjtypes "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	extensionsobjoldtypes "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	extensionsobjold "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
)

const (
	cmPrometheus     = "crd-migrate-prometheus"
	cmServiceMonitor = "crd-migrate-servicemonitor"
	cmAlertmanager   = "crd-migrate-alertmanager"
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
	cmClient      corev1.ConfigMapInterface
	dynClient     *dynamic.Client
	dynClientTPR  *dynamic.Client
	restClient    rest.Interface
	restClientTPR rest.Interface
	crdClient     extensionsobjtypes.CustomResourceDefinitionInterface
	tprClient     extensionsobjoldtypes.ThirdPartyResourceInterface
	nsClient      corev1.NamespaceInterface

	migrated map[string]*migrationState

	cmSuffix string
	logger   log.Logger
}

func GetMigration(cfg *rest.Config, logger log.Logger) (int, error) {
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return NoMigration, err
	}

	mv, err := k8sutil.GetMinorVersion(client.Discovery())
	if err != nil {
		return NoMigration, err
	}
	if mv >= 7 {
		// See if the TPRs managed by this operator exist.
		tprClient := client.Extensions().ThirdPartyResources()
		_, err = tprClient.Get("prometheus."+monitoringv1.Group, metav1.GetOptions{})
		if err == nil {
			return TPR2CRD, nil
		}
		_, err = tprClient.Get("service-monitor."+monitoringv1.Group, metav1.GetOptions{})
		if err == nil {
			return TPR2CRD, nil
		}
		_, err = tprClient.Get("alertmanager."+monitoringv1.Group, metav1.GetOptions{})
		if err == nil {
			return TPR2CRD, nil
		}
	}

	return NoMigration, nil
}

func NewMigrator(cfg *rest.Config, logger log.Logger) (*Migrator, error) {
	m := &Migrator{
		migrated: make(map[string]*migrationState),
		logger:   logger,
		cmSuffix: strconv.FormatInt(time.Now().Unix(), 10),
	}

	tprCfg := *cfg
	tprCfg.GroupVersion = &schema.GroupVersion{
		Group:   monitoringv1.Group,
		Version: "v1alpha1",
	}

	kclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating kubernetes client")
	}

	restClient, err := rest.RESTClientFor(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating rest client")
	}
	restClientTPR, err := rest.RESTClientFor(&tprCfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating rest client")
	}
	cmClient := kclient.CoreV1().ConfigMaps("default")
	dynClient, err := dynamic.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating a dynamic client")
	}
	dynClientTPR, err := dynamic.NewClient(&tprCfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating a dynamic client")
	}
	extClient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating apiextensions client")
	}
	crdClient := extClient.ApiextensionsV1beta1().CustomResourceDefinitions()
	tprClient := kclient.ExtensionsV1beta1().ThirdPartyResources()

	m.dynClient = dynClient
	m.dynClientTPR = dynClientTPR
	m.crdClient = crdClient
	m.restClient = restClient
	m.restClientTPR = restClientTPR
	m.tprClient = tprClient
	m.cmClient = cmClient
	m.nsClient = kclient.CoreV1().Namespaces()

	return m, nil
}

func MigrateTPR2CRD(cfg *rest.Config, logger log.Logger) error {
	cfg.GroupVersion = &schema.GroupVersion{
		Group:   monitoringv1.Group,
		Version: monitoringv1.Version,
	}
	cfg.APIPath = "/apis"
	cfg.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}

	m, err := NewMigrator(cfg, logger)
	if err != nil {
		return errors.Wrap(err, "creating the migrator")
	}

	_, err = m.tprClient.Get("prometheus."+monitoringv1.Group, metav1.GetOptions{})
	if err == nil {
		if err := m.MigratePrometheus(); err != nil {
			logger.Log("error", err)
			return m.Rollback()
		}
	}

	_, err = m.tprClient.Get("service-monitor."+monitoringv1.Group, metav1.GetOptions{})
	if err == nil {
		if err := m.MigrateServiceMonitor(); err != nil {
			logger.Log("error", err)
			return m.Rollback()
		}
	}

	_, err = m.tprClient.Get("alertmanager."+monitoringv1.Group, metav1.GetOptions{})
	if err == nil {
		if err := m.MigrateAlertmanager(); err != nil {
			logger.Log("error", err)
			return m.Rollback()
		}
	}

	return nil
}

func (m *Migrator) migrateTPR(tprKind, tprName, tprOldReg, configmap string, shortNames []string) error {
	m.logger.Log("msg", fmt.Sprintf("migrating TPR \"%s\"", tprKind))

	m.migrated[tprKind] = new(migrationState)

	err := m.moveTPRDataToCM(tprName, configmap)
	if err != nil {
		return errors.Wrapf(err, "error moving \"%s\" TPR data to config-map \"%s\"", tprKind, configmap)
	}
	m.migrated[tprKind].backedUp = true

	err = m.deleteTPRData(tprKind, tprName)
	if err != nil {
		return errors.Wrapf(err, "deleting TPR Data for \"%s\"", tprKind)
	}
	m.migrated[tprKind].tprDataDel = true

	err = m.deleteTPR(tprOldReg + "." + monitoringv1.Group)
	if err != nil {
		return errors.Wrapf(err, "deleting TPR registration for \"%s\"", tprKind)
	}
	m.migrated[tprKind].tprRegDel = true

	crd := &extensionsobj.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: tprName + "." + monitoringv1.Group,
		},
		Spec: extensionsobj.CustomResourceDefinitionSpec{
			Group:   monitoringv1.Group,
			Version: monitoringv1.Version,
			Scope:   extensionsobj.NamespaceScoped,
			Names: extensionsobj.CustomResourceDefinitionNames{
				Plural:     tprName,
				Kind:       tprKind,
				ShortNames: shortNames,
			},
		},
	}
	err = m.createCRD(crd)
	if err != nil {
		return errors.Wrapf(err, "error creating CRD registration for \"%s\"", tprKind)
	}
	m.migrated[tprKind].crdCreated = true

	if err := k8sutil.WaitForCRDReady(m.restClient, monitoringv1.Group, monitoringv1.Version, monitoringv1.PrometheusName); err != nil {
		return errors.Wrapf(err, "error waiting for ready CRD registration for \"%s\"", tprKind, configmap)
	}

	err = m.restoreFromCM(tprKind, tprName, configmap)
	if err != nil {
		return errors.Wrapf(err, "error restoring \"%s\" CRs from \"%s\"", tprKind, configmap)
	}

	m.migrated[tprKind].finished = true

	return nil
}

func (m *Migrator) MigratePrometheus() error {
	return m.migrateTPR(monitoringv1.PrometheusesKind, monitoringv1.PrometheusName, "prometheus", cmPrometheus, []string{monitoringv1.PrometheusShort})
}

func (m *Migrator) MigrateServiceMonitor() error {
	return m.migrateTPR(monitoringv1.ServiceMonitorsKind, monitoringv1.ServiceMonitorName, "service-monitor", cmServiceMonitor, nil)
}

func (m *Migrator) MigrateAlertmanager() error {
	return m.migrateTPR(monitoringv1.AlertmanagersKind, monitoringv1.AlertmanagerName, "alertmanager", cmAlertmanager, nil)
}

func (m *Migrator) Rollback() error {
	oldVersion := "v1alpha1"

	if err := m.rollback(
		monitoringv1.PrometheusesKind,
		monitoringv1.PrometheusName,
		cmPrometheus,
		&extensionsobjold.ThirdPartyResource{
			ObjectMeta: metav1.ObjectMeta{
				Name: "prometheus." + monitoringv1.Group,
			},
			Versions: []extensionsobjold.APIVersion{
				{Name: oldVersion},
			},
			Description: "Managed Prometheus server",
		},
	); err != nil {
		return errors.Wrap(err, "rollback for Prometheus TPR")
	}

	if err := m.rollback(
		monitoringv1.ServiceMonitorsKind,
		monitoringv1.ServiceMonitorName,
		cmServiceMonitor,
		&extensionsobjold.ThirdPartyResource{
			ObjectMeta: metav1.ObjectMeta{
				Name: "service-monitor." + monitoringv1.Group,
			},
			Versions: []extensionsobjold.APIVersion{
				{Name: oldVersion},
			},
			Description: "Prometheus monitoring for a service",
		},
	); err != nil {
		return errors.Wrap(err, "rollback for ServiceMonitor TPR")
	}

	if err := m.rollback(
		monitoringv1.AlertmanagersKind,
		monitoringv1.AlertmanagerName,
		cmAlertmanager,
		&extensionsobjold.ThirdPartyResource{
			ObjectMeta: metav1.ObjectMeta{
				Name: "alertmanager." + monitoringv1.Group,
			},
			Versions: []extensionsobjold.APIVersion{
				{Name: oldVersion},
			},
			Description: "Managed Alertmanager cluster",
		},
	); err != nil {
		return errors.Wrap(err, "rollback for Alertmanager TPR")
	}

	return nil
}

func (m *Migrator) rollback(tprKind, tprName, configmapName string, tpr *extensionsobjold.ThirdPartyResource) error {
	if ms, ok := m.migrated[tprKind]; ok {
		if ms.finished {
			err := m.deleteTPRData(tprKind, tprName)
			if err != nil {
				return errors.Wrapf(err, "deleting the CRD data for \"%s\"", tprKind)
			}
		}

		if ms.crdCreated {
			err := m.deleteCRD(tprName + "." + monitoringv1.Group)
			if err != nil {
				return errors.Wrapf(err, "deleting the CRD registration for \"%s\"", tprKind)
			}
		}

		if ms.tprRegDel {
			_, err := m.tprClient.Create(tpr)
			if err != nil {
				return errors.Wrapf(err, "recreating the TPR registration for \"%s\"", tprKind)
			}
		}

		if ms.tprDataDel {
			err := m.restoreFromCM(tprKind, tprName, configmapName)
			if err != nil {
				return errors.Wrapf(
					err,
					"recreating the TPR data for \"%s\" from config-map \"%s\"",
					tprKind, configmapName)
			}
		}
	}

	return nil
}

func (m *Migrator) moveTPRDataToCM(tprName, configmapName string) error {
	if m.cmSuffix != "" {
		configmapName = configmapName + "-" + m.cmSuffix
	}

	req := m.restClientTPR.Get().
		Namespace(api.NamespaceAll).
		Resource(tprName).
		FieldsSelectorParam(nil)

	tprdata, err := req.DoRaw()
	if err != nil {
		return nil
	}

	cm := v1.ConfigMap{}

	cm.ObjectMeta.SetNamespace("default")
	cm.ObjectMeta.SetName(configmapName)

	cm.Data = map[string]string{"backup": string(tprdata)}

	_, err = m.cmClient.Create(&cm)
	return err
}

func (m *Migrator) deleteTPR(tprName string) error {
	return m.tprClient.Delete(tprName, &metav1.DeleteOptions{})
}

func (m *Migrator) deleteCRD(crdName string) error {
	return m.crdClient.Delete(crdName, &metav1.DeleteOptions{})
}

func (m *Migrator) createCRD(crd *extensionsobj.CustomResourceDefinition) error {
	_, err := m.crdClient.Create(crd)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrapf(err, "Creating CRD: %s", crd.Spec.Names.Kind)
	}

	return nil
}

func (m *Migrator) restoreFromCM(tprKind, tprName, configmapName string) error {
	if m.cmSuffix != "" {
		configmapName = configmapName + "-" + m.cmSuffix
	}

	cm, err := m.cmClient.Get(configmapName, metav1.GetOptions{})
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
		item.SetAPIVersion(monitoringv1.Group + "/" + monitoringv1.Version)
		// Namespace issues
		resClient := m.dynClient.Resource(
			&metav1.APIResource{
				Kind:       tprKind,
				Name:       tprName,
				Namespaced: true,
			},
			item.GetNamespace(),
		)
		_, err = resClient.Create(&item)
		if err != nil {
			return errors.Wrap(err, "restoring tprdata")
		}
	}

	return nil
}

func (m *Migrator) deleteTPRData(tprKind, tprName string) error {
	nsList, err := m.nsClient.List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "listing namspaces")
	}

	for _, ns := range nsList.Items {
		resClient := m.dynClientTPR.Resource(
			&metav1.APIResource{
				Kind:       tprKind,
				Name:       tprName,
				Namespaced: true,
			},
			ns.ObjectMeta.Name,
		)

		err := resClient.DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{})
		if err != nil {
			return errors.Wrapf(err, "deleting tprdata in namespace \"%s\"", ns.ObjectMeta.Name)
		}
	}

	return nil
}
