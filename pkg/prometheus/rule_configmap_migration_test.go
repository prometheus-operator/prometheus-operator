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
	"os"
	"path/filepath"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/pkg/errors"
)

func TestCMToRuleFiles(t *testing.T) {
	cm, err := parseConfigMapYaml("../../contrib/kube-prometheus/manifests/prometheus-rules.yaml")
	if err != nil {
		t.Fatal(err)
	}

	_, err = CMToRule(cm)
	if err != nil {
		t.Fatal(err)
	}
}

// ParseConfigMapYaml takes a path to a yaml file and returns a Kubernetes
// ConfigMap
func parseConfigMapYaml(relativePath string) (*v1.ConfigMap, error) {
	absolutPath, err := filepath.Abs(relativePath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed generate absolut file path of %s", relativePath))
	}

	manifest, err := os.Open(absolutPath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to open file %s", absolutPath))
	}

	configMap := v1.ConfigMap{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&configMap); err != nil {
		return nil, err
	}

	return &configMap, nil

}
