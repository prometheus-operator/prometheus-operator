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
	"fmt"
	"sort"

	"github.com/pkg/errors"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// MaxSecretDataSizeBytes is the maximum data size that a single secret shard
// may use. This is lower than v1.MaxSecretSize in order to reserve space for
// metadata and the rest of the secret k8s object.
const MaxSecretDataSizeBytes = v1.MaxSecretSize - 50_000

// ShardedSecret is k8s secret data that is sharded across multiple enumerated
// k8s secrets. This is used to circumvent the size limitation of k8s secrets.
type ShardedSecret struct {
	namePrefix   string
	template     *v1.Secret
	data         map[string][]byte
	secretShards []*v1.Secret
}

// NewShardedSecret takes a v1.Secret as template and a secret name prefix and
// returns a new ShardedSecret.
func NewShardedSecret(template *v1.Secret, namePrefix string) *ShardedSecret {
	return &ShardedSecret{
		namePrefix: namePrefix,
		template:   template,
		data:       make(map[string][]byte),
	}
}

// AppendData adds data to the secrets data portion. Already existing keys get
// overwritten.
func (s *ShardedSecret) AppendData(key string, data []byte) {
	if s == nil {
		return
	}
	s.data[key] = data
}

// StoreSecrets creates the individual secret shards and stores it via sClient.
func (s *ShardedSecret) StoreSecrets(ctx context.Context, sClient corev1.SecretInterface) error {
	if s == nil {
		return nil
	}
	secrets := s.shard()
	for _, secret := range secrets {
		err := k8sutil.CreateOrUpdateSecret(ctx, sClient, secret)
		if err != nil {
			return errors.Wrapf(err, "failed to create secret shard %q", secret.Name)
		}
	}
	return s.cleanupExcessSecretShards(ctx, sClient, len(secrets)-1, s.namePrefix)
}

// shard does the in-memory sharding of the secret data.
func (s *ShardedSecret) shard() []*v1.Secret {
	// we need to iterate over the data in a stable order to avoid multiple
	// re-shardings followed by secret updates just because the order of data
	// items in the map changed and would now end up in another secret shard.
	var keys []string
	for k := range s.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// reset s.secretShards to ensure it's empty in case shard() is called multiple times
	s.secretShards = []*v1.Secret{}
	curSecretIndex := 0
	curSecretSize := 0
	curSecret := s.newSecretShard(curSecretIndex)

	for _, key := range keys {
		v := s.data[key]
		vSize := len(key) + len(v)
		if curSecretSize+vSize > MaxSecretDataSizeBytes {
			s.secretShards = append(s.secretShards, curSecret)
			curSecretIndex++
			curSecretSize = 0
			curSecret = s.newSecretShard(curSecretIndex)
		}
		curSecretSize += vSize
		curSecret.Data[key] = v
	}
	s.secretShards = append(s.secretShards, curSecret)
	return s.secretShards
}

// newSecretShard allocates a new secret shard according to the secret template
// suffixed by index.
func (s *ShardedSecret) newSecretShard(index int) *v1.Secret {
	newShardSecret := s.template.DeepCopy()
	newShardSecret.Data = make(map[string][]byte)
	newShardSecret.Name = makeShardSecretName(newShardSecret.Name, index)
	return newShardSecret
}

// cleanupExcessSecretShards removes excess secret shards that are no longer in use.
// It also tries to remove a non-sharded secret that exactly matches the name
// prefix in order to make sure that operator version upgrades run smoothly.
func (s *ShardedSecret) cleanupExcessSecretShards(ctx context.Context, sClient corev1.SecretInterface, lastSecretIndex int, sNamePrefix string) error {
	for i := lastSecretIndex + 1; ; i++ {
		secretName := makeShardSecretName(sNamePrefix, i)
		err := sClient.Delete(ctx, secretName, metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			// we reached the end of existing secrets
			break
		}
		if err != nil {
			return errors.Wrapf(err, "failed to delete excess secret shard %q", secretName)
		}
	}
	// Cleanup possibly existing secret of older non-sharded secret versions.
	// TODO: remove this in future versions to save the unnecessary API calls.
	err := sClient.Delete(ctx, sNamePrefix, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "failed to delete non-sharded secret %q from older controller version", sNamePrefix)
	}
	return nil
}

func makeShardSecretName(prefix string, index int) string {
	return fmt.Sprintf("%s-%d", prefix, index)
}

// ShardNames returns the names of the secret shards. This only returns
// something after StoreSecrets was called and the actual sharding took place.
func (s *ShardedSecret) ShardNames() []string {
	if s == nil {
		return []string{}
	}
	var names []string
	for i := 0; i < len(s.secretShards); i++ {
		names = append(names, makeShardSecretName(s.namePrefix, i))
	}
	return names
}
