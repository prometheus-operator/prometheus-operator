// Copyright 2020 The prometheus-operator Authors
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

package thanos

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

func TestCreateOrUpdateRulerConfigSecret(t *testing.T) {
	for _, tc := range []struct {
		name        string
		version     string
		remoteWrite []monitoringv1.RemoteWriteSpec
		golden      string
		expectErr   bool
	}{
		{
			name:    "empty config",
			version: operator.DefaultThanosVersion,
			golden:  "empty_remote_write_config.golden",
		},
		{
			name:    "default version",
			version: operator.DefaultThanosVersion,
			remoteWrite: []monitoringv1.RemoteWriteSpec{
				{
					URL:                  "http://example.com",
					MessageVersion:       ptr.To(monitoringv1.RemoteWriteMessageVersion2_0),
					SendNativeHistograms: ptr.To(true),
					RoundRobinDNS:        ptr.To(true),
				},
			},
			golden: "default_remote_write_config.golden",
		},
		{
			name:    "with v0.24.0",
			version: "v0.24.0",
			remoteWrite: []monitoringv1.RemoteWriteSpec{
				{
					URL: "http://example.com",
				},
			},
			golden: "v0.24.0_remote_write_config.golden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cs := fake.NewClientset()
			o := &Operator{kclient: cs}
			tr := &monitoringv1.ThanosRuler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Spec: monitoringv1.ThanosRulerSpec{
					Version:     ptr.To(tc.version),
					RemoteWrite: tc.remoteWrite,
				},
			}
			sb := &assets.StoreBuilder{}

			err := o.createOrUpdateRulerConfigSecret(context.Background(), sb, tr)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			sec, err := cs.CoreV1().Secrets(tr.Namespace).Get(context.Background(), "thanos-ruler-foo-config", metav1.GetOptions{})
			require.NoError(t, err)
			golden.Assert(t, string(sec.Data[rwConfigFile]), tc.golden)
		})
	}
}
