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

package crd

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func TestPrintAll(t *testing.T) {
	var buf bytes.Buffer
	err := PrintAll(&buf)
	require.NoError(t, err)

	crds := parseCRDs(t, buf.Bytes())
	require.GreaterOrEqual(t, len(crds), 10, "expected at least 10 CRDs")

	// Verify all are valid CRDs
	for _, crd := range crds {
		require.Equal(t, "CustomResourceDefinition", crd.Kind)
		require.Contains(t, crd.Name, "monitoring.coreos.com")
	}
}

func TestPrintAllFull(t *testing.T) {
	var buf bytes.Buffer
	err := PrintAllFull(&buf)
	require.NoError(t, err)

	crds := parseCRDs(t, buf.Bytes())
	require.GreaterOrEqual(t, len(crds), 10, "expected at least 10 full CRDs")

	// Verify all are valid CRDs
	for _, crd := range crds {
		require.Equal(t, "CustomResourceDefinition", crd.Kind)
		require.Contains(t, crd.Name, "monitoring.coreos.com")
	}
}

// parseCRDs parses multi-document YAML into CRD objects.
func parseCRDs(t *testing.T, data []byte) []apiextensionsv1.CustomResourceDefinition {
	t.Helper()

	var crds []apiextensionsv1.CustomResourceDefinition
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)

	for {
		var crd apiextensionsv1.CustomResourceDefinition
		err := decoder.Decode(&crd)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		crds = append(crds, crd)
	}

	return crds
}
