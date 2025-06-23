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

package assets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestHasRefTo(t *testing.T) {
	c := fake.NewSimpleClientset(
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod",
				Namespace: "ns1",
			},
		},
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret",
				Namespace: "ns1",
			},
			Data: map[string][]byte{
				"key1": []byte("val1"),
			},
		},
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret2",
				Namespace: "ns1",
			},
			Data: map[string][]byte{
				"key1": []byte("val1"),
			},
		},
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cm",
				Namespace: "ns1",
			},
			Data: map[string]string{
				"cmCA":   caPEM,
				"cmCert": certPEM,
				"cmKey":  keyPEM,
			},
		},
	)

	store := NewStoreBuilder(c.CoreV1(), c.CoreV1())

	// This secret reference is valid.
	_, err := store.GetSecretKey(
		context.Background(),
		"ns1",
		v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: "secret",
			},
			Key: "key1",
		})
	require.NoError(t, err)

	// This secret doesn't exist but it should be recorded in the RefTracker.
	_, err = store.GetSecretKey(
		context.Background(),
		"ns1",
		v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: "nosecret",
			},
			Key: "key1",
		})
	require.Error(t, err)

	// This configmap reference is valid.
	_, err = store.GetConfigMapKey(
		context.Background(),
		"ns1",
		v1.ConfigMapKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: "cm",
			},
			Key: "cmCA",
		})
	require.NoError(t, err)

	// This configmap doesn't exist but it should be recorded in the RefTracker.
	_, err = store.GetConfigMapKey(
		context.Background(),
		"ns1",
		v1.ConfigMapKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: "nocm",
			},
			Key: "key1",
		})
	require.Error(t, err)

	refTracker := store.RefTracker()

	secret, err := c.CoreV1().Secrets("ns1").Get(context.Background(), "secret", metav1.GetOptions{})
	require.NoError(t, err)
	require.True(t, refTracker.Has(secret))

	_, err = c.CoreV1().Secrets("ns1").Get(context.Background(), "nosecret", metav1.GetOptions{})
	require.Error(t, err)
	require.True(t, refTracker.Has(
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nosecret",
				Namespace: "ns1",
			},
		}),
	)

	cm, err := c.CoreV1().ConfigMaps("ns1").Get(context.Background(), "cm", metav1.GetOptions{})
	require.NoError(t, err)
	require.True(t, refTracker.Has(cm))

	_, err = c.CoreV1().Secrets("ns1").Get(context.Background(), "nocm", metav1.GetOptions{})
	require.Error(t, err)
	require.True(t, refTracker.Has(
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nocm",
				Namespace: "ns1",
			},
		}),
	)

	// This secret has no reference.
	secret, err = c.CoreV1().Secrets("ns1").Get(context.Background(), "secret2", metav1.GetOptions{})
	require.NoError(t, err)
	require.False(t, refTracker.Has(secret))

	// This object's kind is not supported.
	pod, err := c.CoreV1().Pods("ns1").Get(context.Background(), "pod", metav1.GetOptions{})
	require.NoError(t, err)
	require.False(t, refTracker.Has(pod))
}
