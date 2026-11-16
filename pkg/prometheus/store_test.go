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
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
)

func TestAddRemoteReadsToStore(t *testing.T) {
	for _, tc := range []struct {
		name      string
		names     []string
		expectErr bool
	}{
		{
			name:  "empty names",
			names: []string{"", ""},
		},
		{
			name:  "unique names",
			names: []string{"a", "b", ""},
		},
		{
			name:      "duplicate names",
			names:     []string{"a", "", "a"},
			expectErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cs := fake.NewClientset()
			store := assets.NewStoreBuilder(cs.CoreV1(), cs.CoreV1())

			rrs := make([]monitoringv1.RemoteReadSpec, 0, len(tc.names))
			for _, name := range tc.names {
				rrs = append(rrs, monitoringv1.RemoteReadSpec{
					URL:  "http://example.com",
					Name: name,
				})
			}

			err := AddRemoteReadsToStore(context.Background(), store, "default", rrs)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
