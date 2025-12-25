// Copyright 2025 The prometheus-operator Authors
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
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestGenerateServiceMonitorConfigWithServiceName(t *testing.T) {
	cg := &ConfigGenerator{
		logger:  nil,
		version: semver.MustParse("2.45.0"),
	}

	serviceName := "my-service"
	namespace := "my-namespace"

	// Test addSafeTLStoYaml
	tlsConfig := &monitoringv1.SafeTLSConfig{
		ServiceName: &serviceName,
	}

	store := &MockStore{}

	// Call addSafeTLStoYaml
	cfg := yaml.MapSlice{}
	cfg = cg.addSafeTLStoYaml(cfg, namespace, store, tlsConfig)

	// Verify "server_name" inside "tls_config"
	var tlsMap yaml.MapSlice
	foundTLS := false
	for _, item := range cfg {
		if item.Key == "tls_config" {
			tlsMap = item.Value.(yaml.MapSlice)
			foundTLS = true
			break
		}
	}
	require.True(t, foundTLS, "tls_config not found in generated config")

	foundServerName := false
	for _, item := range tlsMap {
		if item.Key == "server_name" {
			foundServerName = true
			require.Equal(t, serviceName+"."+namespace+".svc", item.Value)
		}
	}
	require.True(t, foundServerName, "server_name not found in nested tls_config")
}

// Minimal mock for StoreGetter.
type MockStore struct{}

func (s *MockStore) GetSecretKey(key v1.SecretKeySelector) ([]byte, error) {
	return []byte("secret"), nil
}

func (s *MockStore) GetSecretOrConfigMapKey(key monitoringv1.SecretOrConfigMap) (string, error) {
	return "secret", nil
}

func (s *MockStore) GetConfigMapKey(key v1.ConfigMapKeySelector) (string, error) {
	return "config", nil
}

func (s *MockStore) TLSAsset(key any) string {
	return ""
}
