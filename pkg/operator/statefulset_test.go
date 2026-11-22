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

package operator

import (
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestIsStatefulSetUpToDate(t *testing.T) {
	const inputHash = "12345"

	owner := &metav1.PartialObjectMetadata{
		ObjectMeta: metav1.ObjectMeta{
			Name: "prometheus",
			UID:  types.UID("new-owner-uid"),
		},
	}

	sset := func(inputHash string, ownerUID types.UID) *appsv1.StatefulSet {
		controller := true
		return &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					InputHashAnnotationKey: inputHash,
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "monitoring.coreos.com/v1",
						Kind:       "Prometheus",
						Name:       "prometheus",
						UID:        ownerUID,
						Controller: &controller,
					},
				},
			},
		}
	}

	for _, tc := range []struct {
		name     string
		sset     *appsv1.StatefulSet
		expected bool
	}{
		{
			name:     "same hash and owner",
			sset:     sset(inputHash, "new-owner-uid"),
			expected: true,
		},
		{
			name:     "same hash and previous owner",
			sset:     sset(inputHash, "previous-owner-uid"),
			expected: false,
		},
		{
			name:     "different hash",
			sset:     sset("67890", "new-owner-uid"),
			expected: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, IsStatefulSetUpToDate(tc.sset, inputHash, owner))
		})
	}
}
