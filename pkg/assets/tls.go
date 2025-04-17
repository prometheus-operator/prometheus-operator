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
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	v1 "k8s.io/api/core/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

type source int

const (
	fromSecret source = iota
	fromConfigMap
)

// tlsAssetKey is a unique key for a TLS asset.
type tlsAssetKey struct {
	from source
	ns   string
	name string
	key  string
}

// tlsAssetKeyFromSecretSelector returns a tlsAssetKey struct from a secret key selector.
func tlsAssetKeyFromSecretSelector(ns string, sel *v1.SecretKeySelector) tlsAssetKey {
	return tlsAssetKeyFromSelector(
		ns,
		monitoringv1.SecretOrConfigMap{
			Secret: sel,
		},
	)
}

// tlsAssetKeyFromSelector returns a TLSAssetKey struct from a secret or configmap key selector.
func tlsAssetKeyFromSelector(ns string, sel monitoringv1.SecretOrConfigMap) tlsAssetKey {
	if sel.Secret != nil {
		return tlsAssetKey{
			from: fromSecret,
			ns:   ns,
			name: sel.Secret.Name,
			key:  sel.Secret.Key,
		}
	}

	return tlsAssetKey{
		from: fromConfigMap,
		ns:   ns,
		name: sel.ConfigMap.Name,
		key:  sel.ConfigMap.Key,
	}
}

func (k tlsAssetKey) toString() string {
	return fmt.Sprintf("%d_%s_%s_%s", k.from, k.ns, k.name, k.key)
}

// addTLSAssets processes the given SafeTLSConfig and adds the referenced CA, certificate and key to the store.
func (s *StoreBuilder) addTLSAssets(ctx context.Context, ns string, tlsConfig monitoringv1.SafeTLSConfig) error {
	var (
		err  error
		ca   string
		cert string
		key  string
	)

	ca, err = s.GetKey(ctx, ns, tlsConfig.CA)
	if err != nil {
		return fmt.Errorf("failed to get ca %q: %w", tlsConfig.CA.String(), err)
	}

	cert, err = s.GetKey(ctx, ns, tlsConfig.Cert)
	if err != nil {
		return fmt.Errorf("failed to get cert %q: %w", tlsConfig.Cert.String(), err)
	}

	if tlsConfig.KeySecret != nil {
		key, err = s.GetSecretKey(ctx, ns, *tlsConfig.KeySecret)
		if err != nil {
			return fmt.Errorf("failed to get key %s/%s: %w", tlsConfig.KeySecret.Name, tlsConfig.KeySecret.Key, err)
		}
	}

	if ca != "" {
		block, _ := pem.Decode([]byte(ca))
		if block == nil {
			return fmt.Errorf("ca %s: failed to decode PEM block", tlsConfig.CA.String())
		}

		_, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("ca %s: failed to parse certificate: %w", tlsConfig.CA.String(), err)
		}

		s.tlsAssetKeys[tlsAssetKeyFromSelector(ns, tlsConfig.CA)] = struct{}{}
	}

	if cert != "" && key != "" {
		_, err = tls.X509KeyPair([]byte(cert), []byte(key))
		if err != nil {
			return fmt.Errorf(
				"cert %s, key <%s/%s>: %w",
				tlsConfig.Cert.String(),
				tlsConfig.KeySecret.Name, tlsConfig.KeySecret.Key,
				err)
		}

		s.tlsAssetKeys[tlsAssetKeyFromSelector(ns, tlsConfig.Cert)] = struct{}{}
		s.tlsAssetKeys[tlsAssetKeyFromSecretSelector(ns, tlsConfig.KeySecret)] = struct{}{}
	}

	return nil
}

// AddSafeTLSConfig validates the given SafeTLSConfig and adds it to the store.
func (s *StoreBuilder) AddSafeTLSConfig(ctx context.Context, ns string, tlsConfig *monitoringv1.SafeTLSConfig) error {
	if tlsConfig == nil {
		return nil
	}

	err := tlsConfig.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate TLS configuration: %w", err)
	}

	return s.addTLSAssets(ctx, ns, *tlsConfig)
}

// AddTLSConfig validates the given TLSConfig and adds it to the store.
func (s *StoreBuilder) AddTLSConfig(ctx context.Context, ns string, tlsConfig *monitoringv1.TLSConfig) error {
	if tlsConfig == nil {
		return nil
	}

	err := tlsConfig.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate TLS configuration: %w", err)
	}

	return s.addTLSAssets(ctx, ns, tlsConfig.SafeTLSConfig)
}

// TLSAssets returns a map of TLS assets (certificates and keys) which have
// been added to the store by AddTLSConfig() and AddSafeTLSConfig().
func (s *StoreBuilder) TLSAssets() map[string][]byte {
	m := make(map[string][]byte, len(s.tlsAssetKeys))

	for tak := range s.tlsAssetKeys {
		obj, found, err := s.objStore.GetByKey(fmt.Sprintf("%d/%s/%s", tak.from, tak.ns, tak.name))
		if !found || err != nil {
			continue
		}

		var b []byte
		switch v := obj.(type) {
		case *v1.ConfigMap:
			b = []byte(v.Data[tak.key])
		case *v1.Secret:
			b = v.Data[tak.key]
		}

		if len(b) > 0 {
			m[tak.toString()] = b
		}
	}

	return m
}
