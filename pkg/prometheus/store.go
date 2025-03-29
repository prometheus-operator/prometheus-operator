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

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
)

func AddRemoteWritesToStore(ctx context.Context, store *assets.StoreBuilder, namespace string, remotes []monitoringv1.RemoteWriteSpec) error {
	for i, remote := range remotes {
		if err := validateRemoteWriteSpec(remote); err != nil {
			return fmt.Errorf("remote write %d: %w", i, err)
		}

		if err := store.AddBasicAuth(ctx, namespace, remote.BasicAuth); err != nil {
			return fmt.Errorf("remote write %d: %w", i, err)
		}

		if err := store.AddOAuth2(ctx, namespace, remote.OAuth2); err != nil {
			return fmt.Errorf("remote write %d: %w", i, err)
		}

		if err := store.AddTLSConfig(ctx, namespace, remote.TLSConfig); err != nil {
			return fmt.Errorf("remote write %d: %w", i, err)
		}

		if err := store.AddAuthorizationCredentials(ctx, namespace, remote.Authorization); err != nil {
			return fmt.Errorf("remote write %d: %w", i, err)
		}

		if err := store.AddSigV4(ctx, namespace, remote.Sigv4); err != nil {
			return fmt.Errorf("remote write %d: %w", i, err)
		}

		if err := store.AddAzureOAuth(ctx, namespace, remote.AzureAD); err != nil {
			return fmt.Errorf("remote write %d: %w", i, err)
		}

		if err := store.AddProxyConfig(ctx, namespace, remote.ProxyConfig); err != nil {
			return fmt.Errorf("remote write %d: %w", i, err)
		}
	}

	return nil
}

func AddRemoteReadsToStore(ctx context.Context, store *assets.StoreBuilder, namespace string, remotes []monitoringv1.RemoteReadSpec) error {
	for i, remote := range remotes {
		if err := store.AddBasicAuth(ctx, namespace, remote.BasicAuth); err != nil {
			return fmt.Errorf("remote read %d: %w", i, err)
		}

		if err := store.AddOAuth2(ctx, namespace, remote.OAuth2); err != nil {
			return fmt.Errorf("remote read %d: %w", i, err)
		}

		if err := store.AddTLSConfig(ctx, namespace, remote.TLSConfig); err != nil {
			return fmt.Errorf("remote read %d: %w", i, err)
		}

		if err := store.AddAuthorizationCredentials(ctx, namespace, remote.Authorization); err != nil {
			return fmt.Errorf("remote read %d: %w", i, err)
		}

		if err := remote.Validate(); err != nil {
			return fmt.Errorf("remote read %d: %w", i, err)
		}

		if err := store.AddProxyConfig(ctx, namespace, remote.ProxyConfig); err != nil {
			return fmt.Errorf("remote read %d: %w", i, err)
		}
	}

	return nil
}

func AddAPIServerConfigToStore(ctx context.Context, store *assets.StoreBuilder, namespace string, config *monitoringv1.APIServerConfig) error {
	if config == nil {
		return nil
	}

	if err := store.AddBasicAuth(ctx, namespace, config.BasicAuth); err != nil {
		return fmt.Errorf("apiserver config: %w", err)
	}

	if err := store.AddAuthorizationCredentials(ctx, namespace, config.Authorization); err != nil {
		return fmt.Errorf("apiserver config: %w", err)
	}

	if err := store.AddTLSConfig(ctx, namespace, config.TLSConfig); err != nil {
		return fmt.Errorf("apiserver config: %w", err)
	}

	return nil
}

func AddScrapeClassesToStore(ctx context.Context, store *assets.StoreBuilder, namespace string, scrapeClasses []monitoringv1.ScrapeClass) error {
	for _, scrapeClass := range scrapeClasses {
		if err := store.AddTLSConfig(ctx, namespace, scrapeClass.TLSConfig); err != nil {
			return fmt.Errorf("scrape class %q: %w", scrapeClass.Name, err)
		}
	}
	return nil
}

func addProxyConfigToStore(ctx context.Context, pc monitoringv1.ProxyConfig, store *assets.StoreBuilder, namespace string) error {
	if err := pc.Validate(); err != nil {
		return err
	}

	return store.AddProxyConfig(ctx, namespace, pc)
}
