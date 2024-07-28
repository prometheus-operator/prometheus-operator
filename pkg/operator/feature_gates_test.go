// Copyright 2024 The prometheus-operator Authors
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
)

func TestUpdateFeatureGates(t *testing.T) {
	newFg := func() *FeatureGates {
		return &FeatureGates{
			FeatureGateName("Foo"): {
				description: "foo",
				enabled:     true,
			},
			FeatureGateName("Bar"): {
				description: "bar",
				enabled:     false,
			},
		}
	}

	for _, tc := range []struct {
		flags map[string]bool

		err bool
	}{
		{
			flags: map[string]bool{
				"Foo": false,
				"Bar": true,
			},
		},
		{
			flags: map[string]bool{"Foox": false, "Bar": true},
			err:   true,
		},
	} {
		t.Run("", func(t *testing.T) {
			fg := newFg()
			err := fg.UpdateFeatureGates(tc.flags)
			if tc.err {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			for k, v := range tc.flags {
				require.Equal(t, fg.Enabled(FeatureGateName(k)), v)
			}
		})
	}
}
