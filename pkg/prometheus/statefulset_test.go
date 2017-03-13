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
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
)

var (
	defaultTestConfig = &Config{
		ConfigReloaderImage: "quay.io/coreos/configmap-reload:latest",
	}
)

func TestStatefulSetLabelingAndAnnotations(t *testing.T) {
	labels := map[string]string{
		"testlabel": "testlabelvalue",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	sset := makeStatefulSet(v1alpha1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
	}, nil, defaultTestConfig, []*v1.ConfigMap{})

	if !reflect.DeepEqual(labels, sset.Labels) || !reflect.DeepEqual(annotations, sset.Annotations) {
		t.Fatal("Labels or Annotations are not properly being propagated to the StatefulSet")
	}
}
