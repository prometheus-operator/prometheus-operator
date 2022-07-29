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
	"path/filepath"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/pkg/errors"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

const (
	governingServiceName            = "prometheus-operated"
	defaultRetention                = "24h"
	defaultReplicaExternalLabelName = "prometheus_replica"
	storageDir                      = "/prometheus"
	confDir                         = "/etc/prometheus/config"
	confOutDir                      = "/etc/prometheus/config_out"
	webConfigDir                    = "/etc/prometheus/web_config"
	tlsAssetsDir                    = "/etc/prometheus/certs"
	rulesDir                        = "/etc/prometheus/rules"
	secretsDir                      = "/etc/prometheus/secrets/"
	configmapsDir                   = "/etc/prometheus/configmaps/"
	configFilename                  = "prometheus.yaml.gz"
	configEnvsubstFilename          = "prometheus.env.yaml"
	sSetInputHashName               = "prometheus-operator-input-hash"
	defaultPortName                 = "web"
	defaultQueryLogDirectory        = "/var/log/prometheus"
	defaultQueryLogVolume           = "query-log-file"
)

var (
	minShards                   int32 = 1
	minReplicas                 int32 = 1
	defaultMaxConcurrency       int32 = 20
	managedByOperatorLabel            = "managed-by"
	managedByOperatorLabelValue       = "prometheus-operator"
	managedByOperatorLabels           = map[string]string{
		managedByOperatorLabel: managedByOperatorLabelValue,
	}
	shardLabelName                = "operator.prometheus.io/shard"
	prometheusNameLabelName       = "operator.prometheus.io/name"
	probeTimeoutSeconds     int32 = 3
)

func expectedStatefulSetShardNames(
	p *monitoringv1.Prometheus,
) []string {
	res := []string{}
	shards := minShards
	if p.Spec.Shards != nil && *p.Spec.Shards > 1 {
		shards = *p.Spec.Shards
	}

	for i := int32(0); i < shards; i++ {
		res = append(res, prometheusNameByShard(p.Name, i))
	}

	return res
}

func prometheusNameByShard(name string, shard int32) string {
	base := prefixedName(name)
	if shard == 0 {
		return base
	}
	return fmt.Sprintf("%s-shard-%d", base, shard)
}

func makeEmptyConfigurationSecret(p *monitoringv1.Prometheus, config operator.Config) (*v1.Secret, error) {
	s := makeConfigSecret(p, config)

	s.ObjectMeta.Annotations = map[string]string{
		"empty": "true",
	}

	return s, nil
}

func makeConfigSecret(p *monitoringv1.Prometheus, config operator.Config) *v1.Secret {
	boolTrue := true
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   configSecretName(p.Name),
			Labels: config.Labels.Merge(managedByOperatorLabels),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         p.APIVersion,
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               p.Kind,
					Name:               p.Name,
					UID:                p.UID,
				},
			},
		},
		Data: map[string][]byte{
			configFilename: {},
		},
	}
}

func makeStatefulSetService(p *monitoringv1.Prometheus, config operator.Config) *v1.Service {
	p = p.DeepCopy()

	if p.Spec.PortName == "" {
		p.Spec.PortName = defaultPortName
	}

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: governingServiceName,
			OwnerReferences: []metav1.OwnerReference{
				{
					Name:       p.GetName(),
					Kind:       p.Kind,
					APIVersion: p.APIVersion,
					UID:        p.GetUID(),
				},
			},
			Labels: config.Labels.Merge(map[string]string{
				"operated-prometheus": "true",
			}),
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				{
					Name:       p.Spec.PortName,
					Port:       9090,
					TargetPort: intstr.FromString(p.Spec.PortName),
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name": "prometheus",
			},
		},
	}

	if p.Spec.Thanos != nil {
		svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
			Name:       "grpc",
			Port:       10901,
			TargetPort: intstr.FromString("grpc"),
		})
	}

	return svc
}

func configSecretName(name string) string {
	return prefixedName(name)
}

func tlsAssetsSecretName(name string) string {
	return fmt.Sprintf("%s-tls-assets", prefixedName(name))
}

func webConfigSecretName(name string) string {
	return fmt.Sprintf("%s-web-config", prefixedName(name))
}

func volumeName(name string) string {
	return fmt.Sprintf("%s-db", prefixedName(name))
}

func prefixedName(name string) string {
	return fmt.Sprintf("prometheus-%s", name)
}

func subPathForStorage(s *monitoringv1.StorageSpec) string {
	//nolint:staticcheck // Ignore SA1019 this field is marked as deprecated.
	if s == nil || s.DisableMountSubPath {
		return ""
	}

	return "prometheus-db"
}

func usesDefaultQueryLogVolume(p *monitoringv1.Prometheus) bool {
	return p.Spec.QueryLogFile != "" && filepath.Dir(p.Spec.QueryLogFile) == "."
}

func queryLogFileVolumeMount(p *monitoringv1.Prometheus) (v1.VolumeMount, bool) {
	if !usesDefaultQueryLogVolume(p) {
		return v1.VolumeMount{}, false
	}

	return v1.VolumeMount{
		Name:      defaultQueryLogVolume,
		ReadOnly:  false,
		MountPath: defaultQueryLogDirectory,
	}, true
}

func queryLogFileVolume(p *monitoringv1.Prometheus) (v1.Volume, bool) {
	if !usesDefaultQueryLogVolume(p) {
		return v1.Volume{}, false
	}

	return v1.Volume{
		Name: defaultQueryLogVolume,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}, true
}

func queryLogFilePath(p *monitoringv1.Prometheus) string {
	if !usesDefaultQueryLogVolume(p) {
		return p.Spec.QueryLogFile
	}

	return filepath.Join(defaultQueryLogDirectory, p.Spec.QueryLogFile)
}

func intersection(a, b []string) (i []string) {
	m := make(map[string]struct{})

	for _, item := range a {
		m[item] = struct{}{}
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			i = append(i, item)
		}

		negatedItem := strings.TrimPrefix(item, "no-")
		if item == negatedItem {
			negatedItem = fmt.Sprintf("no-%s", item)
		}

		if _, ok := m[negatedItem]; ok {
			i = append(i, item)
		}
	}
	return i
}

func extractArgKeys(args []monitoringv1.Argument) []string {
	var k []string
	for _, arg := range args {
		key := arg.Name
		k = append(k, key)
	}

	return k
}

func buildArgs(args []monitoringv1.Argument, additionalArgs []monitoringv1.Argument) ([]string, error) {
	var containerArgs []string

	argKeys := extractArgKeys(args)
	additionalArgKeys := extractArgKeys(additionalArgs)

	i := intersection(argKeys, additionalArgKeys)
	if len(i) > 0 {
		return nil, errors.Errorf("can't set arguments which are already managed by the operator: %s", strings.Join(i, ","))
	}

	args = append(args, additionalArgs...)

	for _, arg := range args {
		if arg.Value != "" {
			containerArgs = append(containerArgs, fmt.Sprintf("--%s=%s", arg.Name, arg.Value))
		} else {
			containerArgs = append(containerArgs, fmt.Sprintf("--%s", arg.Name))

		}
	}

	return containerArgs, nil
}
