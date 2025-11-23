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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	sortutil "github.com/prometheus-operator/prometheus-operator/internal/sortutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
)

// MaxSecretDataSizeBytes is the maximum data size that a single secret shard
// may use. This is lower than v1.MaxSecretSize in order to reserve space for
// metadata and the rest of the secret k8s object.
const MaxSecretDataSizeBytes = corev1.MaxSecretSize - 50_000

// ShardedSecret can shard Secret data across multiple k8s Secrets.
// This is used to circumvent the size limitation of k8s Secrets.
type ShardedSecret struct {
	template     *corev1.Secret
	data         map[string][]byte
	secretShards []*corev1.Secret
}

// updateSecrets updates the concrete Secrets from the stored data.
func (s *ShardedSecret) updateSecrets(ctx context.Context, sClient typedcorev1.SecretInterface) error {
	secrets := s.shard()

	for _, secret := range secrets {
		err := k8sutil.CreateOrUpdateSecret(ctx, sClient, secret)
		if err != nil {
			return fmt.Errorf("failed to create secret %q: %w", secret.Name, err)
		}
	}

	return s.cleanupExcessSecretShards(ctx, sClient, len(secrets)-1)
}

// shard does the in-memory sharding of the secret data.
func (s *ShardedSecret) shard() []*corev1.Secret {
	s.secretShards = []*corev1.Secret{}

	currentIndex := 0
	secretSize := 0
	currentSecret := s.newSecretAt(currentIndex)

	for _, key := range sortutil.SortedKeys(s.data) {
		v := s.data[key]
		vSize := len(key) + len(v)
		if secretSize+vSize > MaxSecretDataSizeBytes {
			s.secretShards = append(s.secretShards, currentSecret)
			currentIndex++
			secretSize = 0
			currentSecret = s.newSecretAt(currentIndex)
		}

		secretSize += vSize
		currentSecret.Data[key] = v
	}
	s.secretShards = append(s.secretShards, currentSecret)

	return s.secretShards
}

// newSecretAt creates a new Kubernetes object at the given shard index.
func (s *ShardedSecret) newSecretAt(index int) *corev1.Secret {
	newShardSecret := s.template.DeepCopy()
	newShardSecret.Name = s.secretNameAt(index)
	newShardSecret.Data = make(map[string][]byte)

	return newShardSecret
}

// cleanupExcessSecretShards removes excess secret shards that are no longer in use.
// It also tries to remove a non-sharded secret that exactly matches the name
// prefix in order to make sure that operator version upgrades run smoothly.
func (s *ShardedSecret) cleanupExcessSecretShards(ctx context.Context, sClient typedcorev1.SecretInterface, lastSecretIndex int) error {
	for i := lastSecretIndex + 1; ; i++ {
		secretName := s.secretNameAt(i)
		err := sClient.Delete(ctx, secretName, metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			// we reached the end of existing secrets
			break
		}

		if err != nil {
			return fmt.Errorf("failed to delete secret %q: %w", secretName, err)
		}
	}

	return nil
}

func (s *ShardedSecret) secretNameAt(index int) string {
	return fmt.Sprintf("%s-%d", s.template.Name, index)
}

// Hash implements the Hashable interface from github.com/mitchellh/hashstructure.
func (s *ShardedSecret) Hash() (uint64, error) {
	return uint64(len(s.secretShards)), nil
}

// Volume returns a v1.Volume object with all TLS assets ready to be mounted in a container.
// It must be called after UpdateSecrets().
func (s *ShardedSecret) Volume(name string) corev1.Volume {
	volume := corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{},
			},
		},
	}

	for i := 0; i < len(s.secretShards); i++ {
		volume.Projected.Sources = append(volume.Projected.Sources,
			corev1.VolumeProjection{
				Secret: &corev1.SecretProjection{
					LocalObjectReference: corev1.LocalObjectReference{Name: s.secretNameAt(i)},
				},
			})
	}

	return volume
}

func ReconcileShardedSecret(ctx context.Context, data map[string][]byte, client kubernetes.Interface, template *corev1.Secret) (*ShardedSecret, error) {
	shardedSecret := &ShardedSecret{
		template: template,
		data:     data,
	}

	if err := shardedSecret.updateSecrets(ctx, client.CoreV1().Secrets(template.Namespace)); err != nil {
		return nil, fmt.Errorf("failed to update the TLS secrets: %w", err)
	}

	return shardedSecret, nil
}
