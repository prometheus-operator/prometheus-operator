// Copyright 2023 The prometheus-operator Authors
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

package operator

import (
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestBuildEnvironment(t *testing.T) {
	for _, tc := range []struct {
		a   []v1.EnvVar
		b   []v1.EnvVar
		exp []v1.EnvVar
		err bool
	}{
		{
			a: []v1.EnvVar{
				{Name: "ENV1", Value: "value"},
				{Name: "ENV2", Value: "value2"},
			},
			b: []v1.EnvVar{
				{Name: "ENV12", Value: "value"},
				{Name: "ENV21", Value: "value2"},
			},
			exp: []v1.EnvVar{
				{Name: "ENV1", Value: "value"},
				{Name: "ENV2", Value: "value2"},
				{Name: "ENV12", Value: "value"},
				{Name: "ENV21", Value: "value2"},
			},
		},
		{
			a: []v1.EnvVar{
				{Name: "TEST", Value: "value"},
				{Name: "TEST2", Value: "value2"},
			},
			b: []v1.EnvVar{
				{Name: "ENV", Value: "value"},
				{Name: "TEST2", Value: "value3"},
			},
			err: true,
		},
	} {
		t.Run("", func(t *testing.T) {
			args, err := BuildEnvironment(tc.a, tc.b)
			if tc.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.exp, args)
		})
	}
}
