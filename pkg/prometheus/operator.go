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

package prometheus

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	ResyncPeriod = 5 * time.Minute
)

var prometheusKeyInShardStatefulSet = regexp.MustCompile("^(.+)/prometheus-(.+)-shard-[1-9][0-9]*$")
var prometheusKeyInStatefulSet = regexp.MustCompile("^(.+)/prometheus-(.+)$")

type CommonConfig struct {
	ClusterConfig *rest.Config
	KClient       *kubernetes.Clientset
	MClient       *monitoringclient.Clientset

	PromInfs   *informers.ForResource
	PromAgInfs *informers.ForResource
	SmonInfs   *informers.ForResource
	PmonInfs   *informers.ForResource
	ProbeInfs  *informers.ForResource
	CmapInfs   *informers.ForResource
	SecrInfs   *informers.ForResource
	SsetInfs   *informers.ForResource
}

func StatefulSetKeyToPrometheusKey(key string) (bool, string) {
	r := prometheusKeyInStatefulSet
	if prometheusKeyInShardStatefulSet.MatchString(key) {
		r = prometheusKeyInShardStatefulSet
	}

	matches := r.FindAllStringSubmatch(key, 2)
	if len(matches) != 1 {
		return false, ""
	}
	if len(matches[0]) != 3 {
		return false, ""
	}
	return true, matches[0][1] + "/" + matches[0][2]
}

func KeyToStatefulSetKey(p monitoringv1.PrometheusInterface, key string, shard int) string {
	keyParts := strings.Split(key, "/")
	return fmt.Sprintf("%s/%s", keyParts[0], statefulSetNameFromPrometheusName(p, keyParts[1], shard))
}

func statefulSetNameFromPrometheusName(p monitoringv1.PrometheusInterface, name string, shard int) string {
	if shard == 0 {
		return fmt.Sprintf("%s-%s", prefix(p), name)
	}
	return fmt.Sprintf("%s-%s-shard-%d", prefix(p), name, shard)
}

func NewTLSAssetSecret(p monitoringv1.PrometheusInterface, labels map[string]string) *v1.Secret {
	objMeta := p.GetObjectMeta()
	typeMeta := p.GetTypeMeta()

	boolTrue := true
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   TLSAssetsSecretName(p),
			Labels: labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         typeMeta.APIVersion,
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               typeMeta.Kind,
					Name:               objMeta.GetName(),
					UID:                objMeta.GetUID(),
				},
			},
		},
		Data: map[string][]byte{},
	}
}

// ValidateRemoteWriteSpec checks that mutually exclusive configurations are not
// included in the Prometheus remoteWrite configuration section.
// Reference:
// https://github.com/prometheus/prometheus/blob/main/docs/configuration/configuration.md#remote_write
func ValidateRemoteWriteSpec(spec monitoringv1.RemoteWriteSpec) error {
	var nonNilFields []string
	for k, v := range map[string]interface{}{
		"basicAuth":     spec.BasicAuth,
		"oauth2":        spec.OAuth2,
		"authorization": spec.Authorization,
		"sigv4":         spec.Sigv4,
	} {
		if reflect.ValueOf(v).IsNil() {
			continue
		}
		nonNilFields = append(nonNilFields, fmt.Sprintf("%q", k))
	}

	if len(nonNilFields) > 1 {
		return errors.Errorf("%s can't be set at the same time, at most one of them must be defined", strings.Join(nonNilFields, " and "))
	}

	return nil
}

func InitCommonConfig(c operator.Config) (*CommonConfig, error) {
	var cm CommonConfig
	var err error

	cm.ClusterConfig, err = k8sutil.NewClusterConfig(c.Host, c.TLSInsecure, &c.TLSConfig)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating cluster config failed")
	}

	cm.KClient, err = kubernetes.NewForConfig(cm.ClusterConfig)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating kubernetes client failed")
	}

	cm.MClient, err = monitoringclient.NewForConfig(cm.ClusterConfig)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating monitoring client failed")
	}

	secretListWatchSelector, err := fields.ParseSelector(c.SecretListWatchSelector)
	if err != nil {
		return nil, errors.Wrap(err, "can not parse secrets selector value")
	}
	// init promInfs
	cm.PromInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.Namespaces.PrometheusAllowList,
			c.Namespaces.DenyList,
			cm.MClient,
			ResyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = c.PromSelector.String()
			},
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PrometheusName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating prometheus informers")
	}

	// init promAgInfs
	cm.PromAgInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.Namespaces.PrometheusAllowList,
			c.Namespaces.DenyList,
			cm.MClient,
			ResyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = c.PromSelector.String()
			},
		),
		monitoringv1alpha1.SchemeGroupVersion.WithResource(monitoringv1alpha1.PrometheusAgentName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating prometheus-agent informers")
	}

	// init smonInfs
	cm.SmonInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.Namespaces.AllowList,
			c.Namespaces.DenyList,
			cm.MClient,
			ResyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ServiceMonitorName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating servicemonitor informers")
	}

	// init pmonInfs
	cm.PmonInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.Namespaces.AllowList,
			c.Namespaces.DenyList,
			cm.MClient,
			ResyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PodMonitorName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating podmonitor informers")
	}

	// init probeInfs
	cm.ProbeInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.Namespaces.AllowList,
			c.Namespaces.DenyList,
			cm.MClient,
			ResyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ProbeName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating probe informers")
	}

	// init cmapInfs
	cm.CmapInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			c.Namespaces.PrometheusAllowList,
			c.Namespaces.DenyList,
			cm.KClient,
			ResyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = LabelPrometheusName
			},
		),
		v1.SchemeGroupVersion.WithResource(string(v1.ResourceConfigMaps)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating configmap informers")
	}

	// init secrInfs
	cm.SecrInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			c.Namespaces.PrometheusAllowList,
			c.Namespaces.DenyList,
			cm.KClient,
			ResyncPeriod,
			func(options *metav1.ListOptions) {
				options.FieldSelector = secretListWatchSelector.String()
			},
		),
		v1.SchemeGroupVersion.WithResource(string(v1.ResourceSecrets)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating secrets informers")
	}

	// init ssetInfs
	cm.SsetInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			c.Namespaces.PrometheusAllowList,
			c.Namespaces.DenyList,
			cm.KClient,
			ResyncPeriod,
			nil,
		),
		appsv1.SchemeGroupVersion.WithResource("statefulsets"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating statefulset informers")
	}

	return &cm, nil
}
