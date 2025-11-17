// Copyright 2021 The prometheus-operator Authors
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
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"
)

func TestShardedSecret(t *testing.T) {
	const namePrefix = "secret"

	tt := []struct {
		desc         string
		input        map[string][]byte
		expectShards int
	}{
		{
			desc:         "empty data",
			input:        make(map[string][]byte),
			expectShards: 1,
		},
		{
			desc: "one shard",
			input: map[string][]byte{
				"key": []byte("data"),
			},
			expectShards: 1,
		},
		{
			desc: "exactly the size limit",
			input: map[string][]byte{
				"key": make([]byte, MaxSecretDataSizeBytes-3), // -3 because of the key size
			},
			expectShards: 1,
		},
		{
			desc: "slightly over the size limit",
			input: map[string][]byte{
				"key": make([]byte, MaxSecretDataSizeBytes), // max size will push us over the limit because of the key size
			},
			expectShards: 2,
		},
		{
			desc: "three shards",
			input: map[string][]byte{
				"one":   make([]byte, MaxSecretDataSizeBytes-3), // -3 because of the key size
				"two":   make([]byte, MaxSecretDataSizeBytes-3), // -3 because of the key size
				"three": []byte("data"),
			},
			expectShards: 3,
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			template := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: namePrefix,
				},
			}
			s := &ShardedSecret{
				template: template,
				data:     tc.input,
			}

			secrets := s.shard()
			if len(secrets) != tc.expectShards {
				t.Errorf("sharding failed: got %d shards; want %d", len(secrets), tc.expectShards)
			}
		})
	}
}

func TestCleanupExcessSecretShardsSkipsMissing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	namespace := "ns"
	client := fake.NewSimpleClientset(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-0",
			Namespace: namespace,
		},
	})

	s := &ShardedSecret{
		template: &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret",
				Namespace: namespace,
			},
		},
	}

	if err := s.cleanupExcessSecretShards(ctx, client.CoreV1().Secrets(namespace), 0); err != nil {
		t.Fatalf("cleanupExcessSecretShards returned error: %v", err)
	}

	actions := client.Actions()
	if len(actions) != 1 || actions[0].GetVerb() != "get" {
		t.Fatalf("unexpected client actions: %#v", actions)
	}
}

func TestCleanupExcessSecretShardsRemovesExtra(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	namespace := "ns"
	client := fake.NewSimpleClientset(
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret-0",
				Namespace: namespace,
			},
		},
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret-1",
				Namespace: namespace,
			},
		},
	)

	s := &ShardedSecret{
		template: &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret",
				Namespace: namespace,
			},
		},
	}

	if err := s.cleanupExcessSecretShards(ctx, client.CoreV1().Secrets(namespace), 0); err != nil {
		t.Fatalf("cleanupExcessSecretShards returned error: %v", err)
	}

	actions := append([]clientgotesting.Action(nil), client.Actions()...)
	if _, err := client.CoreV1().Secrets(namespace).Get(ctx, "secret-1", metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatalf("expected secret-1 to be deleted, got err=%v", err)
	}

	if len(actions) != 3 {
		t.Fatalf("unexpected number of client actions: %#v", actions)
	}
	if actions[0].GetVerb() != "get" || actions[1].GetVerb() != "delete" || actions[2].GetVerb() != "get" {
		t.Fatalf("unexpected action order: %#v", actions)
	}
}
