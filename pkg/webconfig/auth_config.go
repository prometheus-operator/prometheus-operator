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

package webconfig

import (
	"context"
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
)

var (
	managedKey      = "webconfig/managed/"
	userKey         = "webconfig/userDefine/"
	managedUserName = "prometheus-operator-managed-user"
)

// authConfig
type authConfig struct {
	namespace      string
	basicAuthUsers []monitoringv1.BasicAuth
}

func newAuthConfig(namespace string, basicAuthUsers []monitoringv1.BasicAuth) authConfig {
	return authConfig{
		namespace:      namespace,
		basicAuthUsers: basicAuthUsers,
	}
}

func (a authConfig) addAuthsToStore(ctx context.Context, managedUserPassword string, store *assets.Store) error {
	store.BasicAuthAssets[fmt.Sprintf("%s%s", managedKey, managedUserName)] = assets.BasicAuthCredentials{
		Username: managedUserName,
		Password: managedUserPassword,
	}

	for _, bau := range a.basicAuthUsers {
		if err := store.AddBasicAuth(ctx, a.namespace, &bau, fmt.Sprintf("%s%s", userKey, bau.Username.Key)); err != nil {
			return err
		}
	}

	return nil
}

func (a authConfig) getManagedAuth(store *assets.Store) (string, string, error) {
	v, ok := store.BasicAuthAssets[fmt.Sprintf("%s%s", managedKey, managedUserName)]
	if !ok {
		return "", "", fmt.Errorf("key %s not found", fmt.Sprintf("%s%s", managedKey, managedUserName))
	}

	return v.Username, v.Password, nil
}

func (a authConfig) getUserDefineAuths(store *assets.Store) (map[string]string, error) {
	auths := map[string]string{}

	for _, bau := range a.basicAuthUsers {
		v, ok := store.BasicAuthAssets[fmt.Sprintf("%s%s", userKey, bau.Username.Name)]
		if !ok {
			return nil, fmt.Errorf("key %s not found", fmt.Sprintf("%s%s", userKey, bau.Username.Name))
		}

		auths[v.Username] = v.Password
	}

	return auths, nil
}
