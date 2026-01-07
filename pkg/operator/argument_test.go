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

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestBuildArgs(t *testing.T) {
	for _, tc := range []struct {
		a   []v1.Argument
		b   []v1.Argument
		exp []string
		err bool
	}{
		{
			a: []v1.Argument{
				{Name: "test", Value: "value"},
				{Name: "test2-test", Value: "value2"},
				{Name: "test3.test", Value: "value3"},
			},
			b: []v1.Argument{
				{Name: "addtest", Value: "value"},
				{Name: "addtest2-test", Value: "value2"},
				{Name: "addtest3.test", Value: "value3"},
			},
			exp: []string{
				"--test=value",
				"--test2-test=value2",
				"--test3.test=value3",
				"--addtest=value",
				"--addtest2-test=value2",
				"--addtest3.test=value3",
			},
		},
		{
			a: []v1.Argument{
				{Name: "test", Value: "value"},
				{Name: "test2", Value: "value2"},
			},
			b: []v1.Argument{
				{Name: "addtest", Value: "value"},
				{Name: "test2", Value: "value3"},
			},
			err: true,
		},
		{
			a: []v1.Argument{
				{Name: "test", Value: "value"},
				{Name: "test2", Value: ""},
			},
			b: []v1.Argument{
				{Name: "addtest", Value: "value"},
				{Name: "no-test2", Value: ""},
			},
			err: true,
		},
	} {
		t.Run("", func(t *testing.T) {
			args, err := BuildArgs(tc.a, tc.b)
			if tc.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.exp, args)
		})
	}
}
