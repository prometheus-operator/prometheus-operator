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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	k8sflag "k8s.io/component-base/cli/flag"
)

func TestDefaultFeatureGates(t *testing.T) {
	m := AvailableFeatureGates()

	require.GreaterOrEqual(t, len(m), 1)
}

func TestMapToString(t *testing.T) {
	m := mapToString(defaultFeatureGates)

	require.True(t, strings.Contains(m, "auto-gomemlimit=true"))
}

func TestValidateFeatureGates(t *testing.T) {
	m, err := ValidateFeatureGates(k8sflag.NewMapStringBool(&map[string]bool{"aa": true, "bb": false}))
	require.Error(t, err)
	require.Equal(t, "", m)

	m, err = ValidateFeatureGates(k8sflag.NewMapStringBool(&map[string]bool{"auto-gomemlimit": false}))
	require.NoError(t, err)
	require.True(t, strings.Contains(m, "auto-gomemlimit=false"))
}
