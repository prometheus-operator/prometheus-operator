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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

type fakeOwner struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

var _ = Owner(&fakeOwner{})

func TestUpdateObject(t *testing.T) {
	for _, tc := range []struct {
		opts []ObjectOption
		o    *v1.Secret

		exp *v1.Secret
	}{
		{
			o: &v1.Secret{},
			exp: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"managed-by": "prometheus-operator"},
				},
			},
		},
		{
			opts: []ObjectOption{
				WithLabels(map[string]string{"label1": "val1"}),
				WithLabels(map[string]string{"label3": "val1"}),
				WithLabels(map[string]string{"label3": "val3"}),
				WithAnnotations(map[string]string{"annotation1": "val1"}),
				WithManagingOwner(
					&fakeOwner{
						metav1.TypeMeta{
							Kind:       "Prometheus",
							APIVersion: "monitoring.coreos.com/v1",
						},
						metav1.ObjectMeta{
							Name: "bar",
							UID:  "456",
						},
					},
				),
			},
			o: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"annotation1": "val1", "annotation2": "val2"},
					Labels:      map[string]string{"managed-by": "prometheus-operator2", "label2": "val2"},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "example.com/v1",
							Kind:       "Foo",
							Name:       "foo",
							UID:        "123",
						},
					},
				},
			},
			exp: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"annotation1": "val1",
						"annotation2": "val2",
					},
					Labels: map[string]string{
						"label1":     "val1",
						"label2":     "val2",
						"label3":     "val3",
						"managed-by": "prometheus-operator",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "example.com/v1",
							Kind:       "Foo",
							Name:       "foo",
							UID:        "123",
						},
						{
							APIVersion:         "monitoring.coreos.com/v1",
							Kind:               "Prometheus",
							Name:               "bar",
							UID:                "456",
							Controller:         ptr.To(true),
							BlockOwnerDeletion: ptr.To(true),
						},
					},
				},
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			UpdateObject(tc.o, tc.opts...)

			require.Equal(t, tc.exp, tc.o)
		})
	}
}
