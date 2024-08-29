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

package k8sutil

import (
	"context"
	"log/slog"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestLoadSecretRef(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret",
			Namespace: "ns",
		},
		Data: map[string][]byte{
			"key1": []byte("val1"),
		},
	}

	sClient := fake.NewSimpleClientset(secret).CoreV1().Secrets("ns")
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		// slog level math.MaxInt means no logging
		// We would like to use the slog buil-in No-op level once it is available
		// More: https://github.com/golang/go/issues/62005
		Level: slog.Level(math.MaxInt),
	}))

	for _, tc := range []struct {
		name     string
		ref      *v1.SecretKeySelector
		expected []byte
		err      bool
	}{
		{
			name: "nil ref",
		},
		{
			name: "valid ref",
			ref: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "secret",
				},
				Key: "key1",
			},
			expected: []byte("val1"),
		},
		{
			name: "missing secret",
			ref: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "secret2",
				},
				Key: "key1",
			},
			err: true,
		},
		{
			name: "missing key",
			ref: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "secret",
				},
				Key: "key2",
			},
			err: true,
		},
		{
			name: "missing optional secret",
			ref: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "secret2",
				},
				Key:      "key1",
				Optional: ptr.To(true),
			},
			expected: nil,
		},
		{
			name: "missing optional key",
			ref: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "secret",
				},
				Key:      "key2",
				Optional: ptr.To(true),
			},
			expected: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, err := LoadSecretRef(context.Background(), logger, sClient, tc.ref)
			if tc.err {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, b)
		})
	}
}
