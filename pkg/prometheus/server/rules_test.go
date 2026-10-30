// Copyright The prometheus-operator Authors
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
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

func TestCreateOrUpdateRuleConfigMapsMinCount(t *testing.T) {
	for _, tc := range []struct {
		name     string
		minCount *int32
		want     int
	}{
		{name: "default", want: 3},
		{name: "no reservation", minCount: new(int32(0)), want: 1},
		{name: "one mount", minCount: new(int32(1)), want: 1},
		{name: "below default", minCount: new(int32(2)), want: 2},
		{name: "above default", minCount: new(int32(10)), want: 10},
		{name: "large reservation", minCount: new(int32(100)), want: 100},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			p := &monitoringv1.Prometheus{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "monitoring"},
				Spec: monitoringv1.PrometheusSpec{
					RuleConfigMapMinCount: tc.minCount,
				},
			}
			o := Operator{kclient: fake.NewClientset()}

			names, err := o.createOrUpdateRuleConfigMaps(ctx, p, operator.PrometheusRuleSelection{}, slog.Default())
			require.NoError(t, err)
			require.Len(t, names, tc.want)
			for i, name := range names {
				require.Equal(t, fmt.Sprintf("prometheus-test-rulefiles-%d", i), name)
			}

			// Reserved mounts must not create empty ConfigMap objects.
			cms, err := o.kclient.CoreV1().ConfigMaps(p.Namespace).List(ctx, metav1.ListOptions{})
			require.NoError(t, err)
			require.Len(t, cms.Items, 1)
			require.Equal(t, names[0], cms.Items[0].Name)

			volumes, mounts := appendServerVolumes(p, nil, nil, names)
			require.Len(t, volumes, tc.want)
			require.Len(t, mounts, tc.want)
			for _, volume := range volumes {
				require.NotNil(t, volume.ConfigMap)
				require.Equal(t, new(true), volume.ConfigMap.Optional)
			}
		})
	}
}
