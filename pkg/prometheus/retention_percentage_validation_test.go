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
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
)

// TestRetentionPercentageRejectsQuantitySuffix is the killer regression for the
// Quantity milli footgun: retentionPercentage: "80m" must not produce any
// Prometheus configuration (especially not percentage: 0.08).
func TestRetentionPercentageRejectsQuantitySuffix(t *testing.T) {
	q := resource.MustParse("80m")
	require.InDelta(t, 0.08, q.AsApproximateFloat64(), 1e-9)

	p := defaultPrometheus()
	p.Spec.CommonPrometheusFields.Version = "v3.11.0"
	p.Spec.RetentionPercentage = &q

	cg := mustNewConfigGenerator(t, p)
	cfg, err := cg.GenerateServerConfiguration(
		p,
		nil,
		nil,
		nil,
		nil,
		&assets.StoreBuilder{},
		nil,
		nil,
		nil,
		nil,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "retentionPercentage")
	require.Empty(t, cfg, "rejected retentionPercentage must not yield a Prometheus config")
	require.NotContains(t, string(cfg), "percentage:")
	require.NotContains(t, string(cfg), "0.08")
}

func TestRetentionPercentageValidationBoundaries(t *testing.T) {
	for _, tc := range []struct {
		raw     string
		wantErr bool
	}{
		// accepted integers
		{raw: "0"},
		{raw: "1"},
		{raw: "99"},
		{raw: "100"},
		{raw: "80"},

		// out of range
		{raw: "-1", wantErr: true},
		{raw: "101", wantErr: true},

		// Quantity suffixes (killer cases)
		{raw: "80m", wantErr: true},
		{raw: "80Mi", wantErr: true},
		{raw: "1Gi", wantErr: true},

		// Non-integer decimal forms are rejected too, because a percentage is
		// expected to be an integer and the controller can't distinguish a
		// decimal such as "80.5" from a milli-suffixed quantity such as "80m".
		{raw: "80.5", wantErr: true},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			q := resource.MustParse(tc.raw)
			err := validateRetentionPercentage(&q)
			if tc.wantErr {
				require.Error(t, err, "Go validation should reject %q (String=%q)", tc.raw, q.String())
				return
			}
			require.NoError(t, err, "Go validation should accept %q (String=%q)", tc.raw, q.String())
		})
	}
}

func TestRetentionPercentageAcceptedIntegersEmitConfig(t *testing.T) {
	for _, v := range []int64{0, 1, 99, 100} {
		t.Run(resource.NewQuantity(v, resource.DecimalSI).String(), func(t *testing.T) {
			q := resource.NewQuantity(v, resource.DecimalSI)
			p := defaultPrometheus()
			p.Spec.CommonPrometheusFields.Version = "v3.11.0"
			p.Spec.RetentionPercentage = q

			cg := mustNewConfigGenerator(t, p)
			cfg, err := cg.GenerateServerConfiguration(
				p,
				nil,
				nil,
				nil,
				nil,
				&assets.StoreBuilder{},
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
			require.NotEmpty(t, cfg)
			require.Contains(t, string(cfg), "percentage:")
			// Killer regression: never emit the milli-quantity footgun value.
			require.NotContains(t, string(cfg), "percentage: 0.08")
			require.Contains(t, string(cfg), "percentage: "+resource.NewQuantity(v, resource.DecimalSI).String())
		})
	}
}
