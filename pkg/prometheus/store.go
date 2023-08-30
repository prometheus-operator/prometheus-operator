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

package prometheus

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
)

func AddRemoteWritesToStore(ctx context.Context, store *assets.Store, namespace string, remotes []monv1.RemoteWriteSpec) error {

	for i, remote := range remotes {
		if err := ValidateRemoteWriteSpec(remote); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		key := fmt.Sprintf("remoteWrite/%d", i)
		if err := store.AddBasicAuth(ctx, namespace, remote.BasicAuth, key); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		if err := store.AddOAuth2(ctx, namespace, remote.OAuth2, key); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		if err := store.AddTLSConfig(ctx, namespace, remote.TLSConfig); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		if err := store.AddAuthorizationCredentials(ctx, namespace, remote.Authorization, fmt.Sprintf("remoteWrite/auth/%d", i)); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		if err := store.AddSigV4(ctx, namespace, remote.Sigv4, key); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
	}
	return nil
}

func AddRemoteReadsToStore(ctx context.Context, store *assets.Store, namespace string, remotes []monv1.RemoteReadSpec) error {

	for i, remote := range remotes {
		if err := store.AddBasicAuth(ctx, namespace, remote.BasicAuth, fmt.Sprintf("remoteRead/%d", i)); err != nil {
			return errors.Wrapf(err, "remote read %d", i)
		}
		if err := store.AddOAuth2(ctx, namespace, remote.OAuth2, fmt.Sprintf("remoteRead/%d", i)); err != nil {
			return errors.Wrapf(err, "remote read %d", i)
		}
		if err := store.AddTLSConfig(ctx, namespace, remote.TLSConfig); err != nil {
			return errors.Wrapf(err, "remote read %d", i)
		}
		if err := store.AddAuthorizationCredentials(ctx, namespace, remote.Authorization, fmt.Sprintf("remoteRead/auth/%d", i)); err != nil {
			return errors.Wrapf(err, "remote read %d", i)
		}
	}
	return nil
}

func AddAPIServerConfigToStore(ctx context.Context, store *assets.Store, namespace string, config *monv1.APIServerConfig) error {
	if config == nil {
		return nil
	}

	if err := store.AddBasicAuth(ctx, namespace, config.BasicAuth, "apiserver"); err != nil {
		return errors.Wrap(err, "apiserver config")
	}
	if err := store.AddAuthorizationCredentials(ctx, namespace, config.Authorization, "apiserver/auth"); err != nil {
		return errors.Wrapf(err, "apiserver config")
	}
	return nil
}
