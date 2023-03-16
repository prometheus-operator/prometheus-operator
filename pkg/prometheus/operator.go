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
	"regexp"
	"strings"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var prometheusKeyInShardStatefulSet = regexp.MustCompile("^(.+)/prometheus-(.+)-shard-[1-9][0-9]*$")
var prometheusKeyInStatefulSet = regexp.MustCompile("^(.+)/prometheus-(.+)$")

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

func KeyToStatefulSetKey(key string, shard int) string {
	keyParts := strings.Split(key, "/")
	return fmt.Sprintf("%s/%s", keyParts[0], statefulSetNameFromPrometheusName(keyParts[1], shard))
}

func statefulSetNameFromPrometheusName(name string, shard int) string {
	if shard == 0 {
		return fmt.Sprintf("prometheus-%s", name)
	}
	return fmt.Sprintf("prometheus-%s-shard-%d", name, shard)
}

func NewTLSAssetSecret(p monitoringv1.PrometheusInterface, labels map[string]string) *v1.Secret {
	objMeta := p.GetObjectMeta()
	typeMeta := p.GetTypeMeta()

	boolTrue := true
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   TLSAssetsSecretName(objMeta.GetName()),
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
