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

package v1

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestValidateTSDBSpec(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config *TSDBSpec
		err    bool
	}{
		{
			name:   "TSDBSpec nil",
			config: nil,
			err:    false,
		},
		{
			name: "StaleSeriesCompactionThreshold nil",
			config: &TSDBSpec{
				StaleSeriesCompactionThreshold: nil,
			},
			err: false,
		},
		{
			name: "StaleSeriesCompactionThreshold simple value",
			config: &TSDBSpec{
				StaleSeriesCompactionThreshold: func(v resource.Quantity) *resource.Quantity { return &v }(resource.MustParse("0.5")),
			},
			err: false,
		},
		{
			name: "StaleSeriesCompactionThreshold > 1",
			config: &TSDBSpec{
				StaleSeriesCompactionThreshold: resource.NewQuantity(10, resource.DecimalSI),
			},
			err: true,
		},
		{
			name: "StaleSeriesCompactionThreshold < 0",
			config: &TSDBSpec{
				StaleSeriesCompactionThreshold: resource.NewQuantity(-1, resource.DecimalSI),
			},
			err: true,
		},
		{
			name: "StaleSeriesCompactionThreshold == 0",
			config: &TSDBSpec{
				StaleSeriesCompactionThreshold: resource.NewQuantity(0, resource.DecimalSI),
			},
			err: false,
		},
		{
			name: "StaleSeriesCompactionThreshold == 1",
			config: &TSDBSpec{
				StaleSeriesCompactionThreshold: resource.NewQuantity(1, resource.DecimalSI),
			},
			err: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.err {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got: %s", err)
			}
		})
	}
}
