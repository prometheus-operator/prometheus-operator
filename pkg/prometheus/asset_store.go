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

package prometheus

import (
	"context"
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

// tlsAssetKey is a key for a TLS asset.
type tlsAssetKey struct {
	from string
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

// tlsAssetKeyFromSelector returns a tlsAssetKey struct from a secret or configmap key selector.
func tlsAssetKeyFromSelector(ns string, sel monitoringv1.SecretOrConfigMap) tlsAssetKey {
	if sel.Secret != nil {
		return tlsAssetKey{
			from: "secret",
			ns:   ns,
			name: sel.Secret.Name,
			key:  sel.Secret.Key,
		}
	}
	return tlsAssetKey{
		from: "configmap",
		ns:   ns,
		name: sel.ConfigMap.Name,
		key:  sel.ConfigMap.Key,
	}
}

// String implements the fmt.Stringer interface.
func (k tlsAssetKey) String() string {
	return fmt.Sprintf("%s_%s_%s_%s", k.from, k.ns, k.name, k.key)
}

// assetStore is a store that fetches and caches TLS materials, bearer tokens
// and auth credentials from configmaps and secrets.
// Data can be referenced directly from a Prometheus object or indirectly (for
// instance via ServiceMonitor). In practice a new store is created and used by
// each reconciliation loop.
//
// assetStore doesn't support concurrent access.
type assetStore struct {
	cmClient corev1client.ConfigMapsGetter
	sClient  corev1client.SecretsGetter
	objStore cache.Store

	tlsAssets         map[tlsAssetKey]TLSAsset
	bearerTokenAssets map[string]BearerToken
	basicAuthAssets   map[string]BasicAuthCredentials
}

// newAssetStore returns an empty assetStore.
func newAssetStore(cmClient corev1client.ConfigMapsGetter, sClient corev1client.SecretsGetter) *assetStore {
	return &assetStore{
		cmClient:          cmClient,
		sClient:           sClient,
		tlsAssets:         make(map[tlsAssetKey]TLSAsset),
		bearerTokenAssets: make(map[string]BearerToken),
		basicAuthAssets:   make(map[string]BasicAuthCredentials),
		objStore:          cache.NewStore(assetKeyFunc),
	}
}

func assetKeyFunc(obj interface{}) (string, error) {
	switch v := obj.(type) {
	case *v1.ConfigMap:
		return fmt.Sprintf("0/%s/%s", v.GetNamespace(), v.GetName()), nil
	case *v1.Secret:
		return fmt.Sprintf("1/%s/%s", v.GetNamespace(), v.GetName()), nil
	}
	return "", errors.Errorf("unsupported type: %T", obj)
}

// addTLSConfig processes the given *TLSConfig and adds the referenced CA, certifcate and key to the store.
func (a *assetStore) addTLSConfig(ctx context.Context, ns string, tlsConfig *monitoringv1.TLSConfig) error {
	if tlsConfig == nil {
		return nil
	}

	if tlsConfig.CA != (monitoringv1.SecretOrConfigMap{}) {
		var (
			ca  string
			err error
		)

		switch {
		case tlsConfig.CA.Secret != nil:
			ca, err = a.getSecretKey(ctx, ns, *tlsConfig.CA.Secret)

		case tlsConfig.CA.ConfigMap != nil:
			ca, err = a.getConfigMapKey(ctx, ns, *tlsConfig.CA.ConfigMap)
		}

		if err != nil {
			return errors.Wrap(err, "failed to get CA")
		}

		a.tlsAssets[tlsAssetKeyFromSelector(ns, tlsConfig.CA)] = TLSAsset(ca)
	}

	if tlsConfig.Cert != (monitoringv1.SecretOrConfigMap{}) {
		var (
			cert string
			err  error
		)

		switch {
		case tlsConfig.Cert.Secret != nil:
			cert, err = a.getSecretKey(ctx, ns, *tlsConfig.Cert.Secret)

		case tlsConfig.Cert.ConfigMap != nil:
			cert, err = a.getConfigMapKey(ctx, ns, *tlsConfig.Cert.ConfigMap)
		}

		if err != nil {
			return errors.Wrap(err, "failed to get cert")
		}

		a.tlsAssets[tlsAssetKeyFromSelector(ns, tlsConfig.Cert)] = TLSAsset(cert)
	}

	if tlsConfig.KeySecret != nil {
		key, err := a.getSecretKey(ctx, ns, *tlsConfig.KeySecret)
		if err != nil {
			return errors.Wrap(err, "failed to get key")
		}
		a.tlsAssets[tlsAssetKeyFromSelector(ns, monitoringv1.SecretOrConfigMap{Secret: tlsConfig.KeySecret})] = TLSAsset(key)
	}

	return nil
}

// addTLSConfig processes the given *BasicAuth and adds the referenced credentials to the store.
func (a *assetStore) addBasicAuth(ctx context.Context, ns string, ba *monitoringv1.BasicAuth, key string) error {
	if ba == nil {
		return nil
	}

	username, err := a.getSecretKey(ctx, ns, ba.Username)
	if err != nil {
		return errors.Wrap(err, "failed to get basic auth username")
	}

	password, err := a.getSecretKey(ctx, ns, ba.Password)
	if err != nil {
		return errors.Wrap(err, "failed to get basic auth password")
	}

	a.basicAuthAssets[key] = BasicAuthCredentials{
		username: username,
		password: password,
	}

	return nil
}

// addTLSConfig processes the given SecretKeySelector and adds the referenced data to the store.
func (a *assetStore) addBearerToken(ctx context.Context, ns string, sel v1.SecretKeySelector, key string) error {
	if sel.Name == "" {
		return nil
	}

	bearerToken, err := a.getSecretKey(ctx, ns, sel)
	if err != nil {
		return errors.Wrap(err, "failed to get bearer token")
	}

	a.bearerTokenAssets[key] = BearerToken(bearerToken)

	return nil
}

func (a *assetStore) getConfigMapKey(ctx context.Context, namespace string, sel v1.ConfigMapKeySelector) (string, error) {
	obj, exists, err := a.objStore.Get(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sel.Name,
			Namespace: namespace,
		},
	})
	if err != nil {
		return "", errors.Wrapf(err, "unexpected store error when getting configmap %q", sel.Name)
	}

	if !exists {
		cm, err := a.cmClient.ConfigMaps(namespace).Get(ctx, sel.Name, metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrapf(err, "unable to get configmap %q", sel.Name)
		}
		if err = a.objStore.Add(cm); err != nil {
			return "", errors.Wrapf(err, "unexpected store error when adding configmap %q", sel.Name)
		}
		obj = cm
	}

	cm := obj.(*v1.ConfigMap)
	if _, found := cm.Data[sel.Key]; !found {
		return "", errors.Errorf("key %q in configmap %q not found", sel.Key, sel.Name)
	}

	return cm.Data[sel.Key], nil
}

func (a *assetStore) getSecretKey(ctx context.Context, namespace string, sel v1.SecretKeySelector) (string, error) {
	obj, exists, err := a.objStore.Get(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sel.Name,
			Namespace: namespace,
		},
	})
	if err != nil {
		return "", errors.Wrapf(err, "unexpected store error when getting secret %q", sel.Name)
	}

	if !exists {
		secret, err := a.sClient.Secrets(namespace).Get(ctx, sel.Name, metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrapf(err, "unable to get secret %q", sel.Name)
		}
		if err = a.objStore.Add(secret); err != nil {
			return "", errors.Wrapf(err, "unexpected store error when adding secret %q", sel.Name)
		}
		obj = secret
	}

	secret := obj.(*v1.Secret)
	if _, found := secret.Data[sel.Key]; !found {
		return "", errors.Errorf("key %q in secret %q not found", sel.Key, sel.Name)
	}

	return string(secret.Data[sel.Key]), nil
}
