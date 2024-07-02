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
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// StoreBuilder is a store that fetches and caches TLS materials, bearer tokens
// and auth credentials from configmaps and secrets.
//
// Data can be referenced directly from a Prometheus object or indirectly (for
// instance via ServiceMonitor). In practice a new store is created and used by
// each reconciliation loop.
//
// StoreBuilder doesn't support concurrent access.
type StoreBuilder struct {
	cmClient corev1client.ConfigMapsGetter
	sClient  corev1client.SecretsGetter
	objStore cache.Store

	tlsAssetKeys map[tlsAssetKey]struct{}
}

// NewTestStoreBuilder returns a *StoreBuilder already initialized with the
// provided objects. It is only used in tests.
func NewTestStoreBuilder(objects ...interface{}) *StoreBuilder {
	sb := &StoreBuilder{
		objStore: cache.NewStore(assetKeyFunc),
	}

	for _, o := range objects {
		if err := sb.objStore.Add(o); err != nil {
			panic(err)
		}
	}

	return sb
}

// NewStoreBuilder returns an object that can fetch data from ConfigMaps and Secrets.
func NewStoreBuilder(cmClient corev1client.ConfigMapsGetter, sClient corev1client.SecretsGetter) *StoreBuilder {
	return &StoreBuilder{
		cmClient:     cmClient,
		sClient:      sClient,
		tlsAssetKeys: make(map[tlsAssetKey]struct{}),
		objStore:     cache.NewStore(assetKeyFunc),
	}
}

// assetKeyFunc returns a unique key for a ConfigMap or Secret object.
func assetKeyFunc(obj interface{}) (string, error) {
	switch v := obj.(type) {
	case *v1.ConfigMap:
		return fmt.Sprintf("%d/%s/%s", fromConfigMap, v.GetNamespace(), v.GetName()), nil
	case *v1.Secret:
		return fmt.Sprintf("%d/%s/%s", fromSecret, v.GetNamespace(), v.GetName()), nil
	}

	return "", fmt.Errorf("unsupported type: %T", obj)
}

// AddBasicAuth processes the given *BasicAuth and adds the referenced credentials to the store.
func (s *StoreBuilder) AddBasicAuth(ctx context.Context, ns string, ba *monitoringv1.BasicAuth) error {
	if ba == nil {
		return nil
	}

	_, err := s.GetSecretKey(ctx, ns, ba.Username)
	if err != nil {
		return fmt.Errorf("failed to get basic auth username: %w", err)
	}

	_, err = s.GetSecretKey(ctx, ns, ba.Password)
	if err != nil {
		return fmt.Errorf("failed to get basic auth password: %w", err)
	}

	return nil
}

// AddProxyConfig processes the given *ProxyConfig and adds the referenced credentials to the store.
func (s *StoreBuilder) AddProxyConfig(ctx context.Context, ns string, pc monitoringv1.ProxyConfig) error {
	if len(pc.ProxyConnectHeader) <= 0 {
		return nil
	}

	for k, v := range pc.ProxyConnectHeader {
		for _, v1 := range v {
			_, err := s.GetSecretKey(ctx, ns, v1)
			if err != nil {
				return fmt.Errorf("failed to get proxy config connect header: %s %w", k, err)
			}
		}
	}

	return nil
}

// AddOAuth2 processes the given *OAuth2 and adds the referenced credentials to the store.
func (s *StoreBuilder) AddOAuth2(ctx context.Context, ns string, oauth2 *monitoringv1.OAuth2) error {
	if oauth2 == nil {
		return nil
	}

	if err := oauth2.Validate(); err != nil {
		return err
	}

	_, err := s.GetKey(ctx, ns, oauth2.ClientID)
	if err != nil {
		return fmt.Errorf("failed to get oauth2 client id: %w", err)
	}

	_, err = s.GetSecretKey(ctx, ns, oauth2.ClientSecret)
	if err != nil {
		return fmt.Errorf("failed to get oauth2 client secret: %w", err)
	}

	return nil
}

func (s *StoreBuilder) AddSafeAuthorizationCredentials(ctx context.Context, namespace string, auth *monitoringv1.SafeAuthorization) error {
	if auth == nil || auth.Credentials == nil {
		return nil
	}

	if err := auth.Validate(); err != nil {
		return err
	}

	if auth.Credentials.Name != "" {
		if _, err := s.GetSecretKey(ctx, namespace, *auth.Credentials); err != nil {
			return fmt.Errorf("failed to get authorization token of type %q: %w", auth.Type, err)
		}
	}

	return nil
}

func (s *StoreBuilder) AddAuthorizationCredentials(ctx context.Context, namespace string, auth *monitoringv1.Authorization) error {
	if auth == nil || auth.Credentials == nil {
		return nil
	}

	if err := auth.Validate(); err != nil {
		return err
	}

	if auth.Credentials != nil && auth.Credentials.Name != "" {
		if _, err := s.GetSecretKey(ctx, namespace, *auth.Credentials); err != nil {
			return fmt.Errorf("failed to get authorization token of type %q: %w", auth.Type, err)
		}
	}

	return nil
}

// AddSigV4 processes the SigV4 SecretKeySelectors and adds the SigV4 data to the store.
func (s *StoreBuilder) AddSigV4(ctx context.Context, ns string, sigv4 *monitoringv1.Sigv4) error {
	if sigv4 == nil || (sigv4.AccessKey == nil && sigv4.SecretKey == nil) {
		return nil
	}

	if sigv4.AccessKey == nil || sigv4.SecretKey == nil {
		return errors.New("both accessKey and secretKey should be provided")
	}

	_, err := s.GetSecretKey(ctx, ns, *sigv4.AccessKey)
	if err != nil {
		return fmt.Errorf("failed to read SigV4 access-key: %w", err)
	}

	_, err = s.GetSecretKey(ctx, ns, *sigv4.SecretKey)
	if err != nil {
		return fmt.Errorf("failed to read SigV4 secret-key: %w", err)
	}

	return nil
}

// AddAzureOAuth processes the AzureOAuth SecretKeySelectors and adds the AzureOAuth data to the store.
func (s *StoreBuilder) AddAzureOAuth(ctx context.Context, ns string, azureAD *monitoringv1.AzureAD) error {
	if azureAD == nil || azureAD.OAuth == nil {
		return nil
	}

	_, err := s.GetSecretKey(ctx, ns, azureAD.OAuth.ClientSecret)
	if err != nil {
		return fmt.Errorf("failed to read AzureOAuth clientSecret: %w", err)
	}

	return nil
}

// GetKey processes the given SecretOrConfigMap selector and returns the referenced data.
func (s *StoreBuilder) GetKey(ctx context.Context, namespace string, sel monitoringv1.SecretOrConfigMap) (string, error) {
	switch {
	case sel.Secret != nil:
		return s.GetSecretKey(ctx, namespace, *sel.Secret)
	case sel.ConfigMap != nil:
		return s.GetConfigMapKey(ctx, namespace, *sel.ConfigMap)
	default:
		return "", nil
	}
}

// GetConfigMapKey processes the given ConfigMapKeySelector and returns the referenced data.
func (s *StoreBuilder) GetConfigMapKey(ctx context.Context, namespace string, sel v1.ConfigMapKeySelector) (string, error) {
	if namespace == "" {
		return "", errors.New("namespace cannot be empty")
	}

	obj, exists, err := s.objStore.Get(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sel.Name,
			Namespace: namespace,
		},
	})
	if err != nil {
		return "", fmt.Errorf("unexpected store error when getting configmap %q: %w", sel.Name, err)
	}

	if !exists {
		cm, err := s.cmClient.ConfigMaps(namespace).Get(ctx, sel.Name, metav1.GetOptions{})
		if err != nil {
			return "", fmt.Errorf("unable to get configmap %q: %w", sel.Name, err)
		}
		if err = s.objStore.Add(cm); err != nil {
			return "", fmt.Errorf("unexpected store error when adding configmap %q: %w", sel.Name, err)
		}
		obj = cm
	}

	cm := obj.(*v1.ConfigMap)
	if _, found := cm.Data[sel.Key]; !found {
		return "", fmt.Errorf("key %q in configmap %q not found", sel.Key, sel.Name)
	}

	return cm.Data[sel.Key], nil
}

// GetSecretKey processes the given SecretKeySelector and returns the referenced data.
func (s *StoreBuilder) GetSecretKey(ctx context.Context, namespace string, sel v1.SecretKeySelector) (string, error) {
	if namespace == "" {
		return "", errors.New("namespace cannot be empty")
	}

	obj, exists, err := s.objStore.Get(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sel.Name,
			Namespace: namespace,
		},
	})
	if err != nil {
		return "", fmt.Errorf("unexpected store error when getting secret %q: %w", sel.Name, err)
	}

	if !exists {
		secret, err := s.sClient.Secrets(namespace).Get(ctx, sel.Name, metav1.GetOptions{})
		if err != nil {
			return "", fmt.Errorf("unable to get secret %q: %w", sel.Name, err)
		}
		if err = s.objStore.Add(secret); err != nil {
			return "", fmt.Errorf("unexpected store error when adding secret %q: %w", sel.Name, err)
		}
		obj = secret
	}

	secret := obj.(*v1.Secret)
	if _, found := secret.Data[sel.Key]; !found {
		return "", fmt.Errorf("key %q in secret %q not found", sel.Key, sel.Name)
	}

	return string(secret.Data[sel.Key]), nil
}

// ForNamespace returns a StoreGetter scoped to the given namespace.
// It reads data only from the cache which needs to be populated beforehand.
// The namespace argument can't be empty.
func (s *StoreBuilder) ForNamespace(namespace string) StoreGetter {
	if namespace == "" {
		panic("namespace can't be empty")
	}
	return &cacheOnlyStore{
		ns: namespace,
		c:  s.objStore,
	}
}

type cacheOnlyStore struct {
	ns string
	c  cache.Store
}

var _ = StoreGetter(&cacheOnlyStore{})

func (cos *cacheOnlyStore) GetConfigMapKey(sel v1.ConfigMapKeySelector) (string, error) {
	obj, exists, err := cos.c.Get(&v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: sel.Name, Namespace: cos.ns}})
	if err != nil {
		return "", fmt.Errorf("failed to get configmap %s/%s: %w", cos.ns, sel.Name, err)
	}

	if !exists {
		return "", fmt.Errorf("configmap %s/%s not found", cos.ns, sel.Name)
	}

	cm := obj.(*v1.ConfigMap)
	if _, found := cm.Data[sel.Key]; !found {
		return "", fmt.Errorf("key %q in configmap %s/%s not found", sel.Key, cos.ns, sel.Name)
	}

	return cm.Data[sel.Key], nil
}

func (cos *cacheOnlyStore) GetSecretKey(sel v1.SecretKeySelector) ([]byte, error) {
	obj, exists, err := cos.c.Get(&v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: sel.Name, Namespace: cos.ns}})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s/%s: %w", cos.ns, sel.Name, err)
	}

	if !exists {
		return nil, fmt.Errorf("secret %s/%s not found", cos.ns, sel.Name)
	}

	s := obj.(*v1.Secret)
	if _, found := s.Data[sel.Key]; !found {
		return nil, fmt.Errorf("key %q in secret %s/%s not found", sel.Key, cos.ns, sel.Name)
	}

	return s.Data[sel.Key], nil
}

func (cos *cacheOnlyStore) GetSecretOrConfigMapKey(key monitoringv1.SecretOrConfigMap) (string, error) {
	switch {
	case key.Secret != nil:
		b, err := cos.GetSecretKey(*key.Secret)
		if err != nil {
			return "", err
		}
		return string(b), nil

	case key.ConfigMap != nil:
		return cos.GetConfigMapKey(*key.ConfigMap)

	default:
		return "", nil
	}
}

func (cos *cacheOnlyStore) GetNamespace() string {
	return cos.ns
}
