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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

func TestMakeRulesConfigMaps(t *testing.T) {
	t.Run("ShouldReturnAtLeastOneConfigMap", shouldReturnAtLeastOneConfigMap)
	t.Run("ShouldSplitUpLargeSmallIntoTwo", shouldSplitUpLargeSmallIntoTwo)
}

// makeRulesConfigMaps should return at least one ConfigMap even if it is empty
// when there are no rules. Otherwise adding a rule to a Prometheus without rules
// would change the statefulset definition and thereby force Prometheus to
// restart.
func shouldReturnAtLeastOneConfigMap(t *testing.T) {
	ruleFiles := map[string]string{}

	configMaps, err := makeRulesConfigMaps(&monitoringv1.Prometheus{ObjectMeta: metav1.ObjectMeta{Name: "test"}}, ruleFiles)
	require.NoError(t, err)

	require.Len(t, configMaps, 1, "expected one ConfigMaps but got %v", len(configMaps))
}

func shouldSplitUpLargeSmallIntoTwo(t *testing.T) {
	ruleFiles := map[string]string{}

	ruleFiles["first"] = strings.Repeat("a", operator.MaxConfigMapDataSize)
	ruleFiles["second"] = "a"

	configMaps, err := makeRulesConfigMaps(&monitoringv1.Prometheus{ObjectMeta: metav1.ObjectMeta{Name: "test"}}, ruleFiles)
	require.NoError(t, err)

	require.Len(t, configMaps, 2, "expected rule files to be split up into two ConfigMaps, but got '%v' instead", len(configMaps))

	require.False(t, (configMaps[0].Data["first"] != ruleFiles["first"] || configMaps[1].Data["second"] != ruleFiles["second"]), "expected ConfigMap data to match rule file content")
}
