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
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
)

func TestShardedSecret(t *testing.T) {
	const namePrefix = "secret"

	tt := []struct {
		desc             string
		input            map[string][]byte
		expectShards     int
		expectShardNames []string
	}{
		{
			desc:             "empty data",
			input:            make(map[string][]byte),
			expectShards:     1,
			expectShardNames: []string{namePrefix + "-0"},
		},
		{
			desc: "one shard",
			input: map[string][]byte{
				"key": []byte("data"),
			},
			expectShards:     1,
			expectShardNames: []string{namePrefix + "-0"},
		},
		{
			desc: "exactly the size limit",
			input: map[string][]byte{
				"key": make([]byte, MaxSecretDataSizeBytes-3), // -3 because of the key size
			},
			expectShards:     1,
			expectShardNames: []string{namePrefix + "-0"},
		},
		{
			desc: "slightly over the size limit",
			input: map[string][]byte{
				"key": make([]byte, MaxSecretDataSizeBytes), // max size will push us over the limit because of the key size
			},
			expectShards:     2,
			expectShardNames: []string{namePrefix + "-0", namePrefix + "-1"},
		},
		{
			desc: "three shards",
			input: map[string][]byte{
				"one":   make([]byte, MaxSecretDataSizeBytes-3), // -3 because of the key size
				"two":   make([]byte, MaxSecretDataSizeBytes-3), // -3 because of the key size
				"three": []byte("data"),
			},
			expectShards:     3,
			expectShardNames: []string{namePrefix + "-0", namePrefix + "-1", namePrefix + "-2"},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			template := &v1.Secret{}
			s := NewShardedSecret(template, namePrefix)
			for k, v := range tc.input {
				s.AppendData(k, v)
			}
			secrets := s.shard()
			if len(secrets) != tc.expectShards {
				t.Errorf("sharding failed: got %d shards; want %d", len(secrets), tc.expectShards)
			}
			if diff := cmp.Diff(tc.expectShardNames, s.ShardNames()); diff != "" {
				t.Errorf("ShardNames() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
