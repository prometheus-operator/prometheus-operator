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

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

// Store is a store that fetches and caches TLS materials, bearer tokens
// and auth credentials from configmaps and secrets.
// Data can be referenced directly from a Prometheus object or indirectly (for
// instance via ServiceMonitor). In practice a new store is created and used by
// each reconciliation loop.
//
// Store doesn't support concurrent access.
type Store struct {
	cmClient corev1client.ConfigMapsGetter
	sClient  corev1client.SecretsGetter
	objStore cache.Store

	TLSAssets       map[TLSAssetKey]TLSAsset
	TokenAssets     map[string]Token
	BasicAuthAssets map[string]BasicAuthCredentials
	OAuth2Assets    map[string]OAuth2Credentials
	SigV4Assets     map[string]SigV4Credentials
}

// NewStore returns an empty assetStore.
func NewStore(cmClient corev1client.ConfigMapsGetter, sClient corev1client.SecretsGetter) *Store {
	return &Store{
		cmClient:        cmClient,
		sClient:         sClient,
		TLSAssets:       make(map[TLSAssetKey]TLSAsset),
		TokenAssets:     make(map[string]Token),
		BasicAuthAssets: make(map[string]BasicAuthCredentials),
		OAuth2Assets:    make(map[string]OAuth2Credentials),
		SigV4Assets:     make(map[string]SigV4Credentials),
		objStore:        cache.NewStore(assetKeyFunc),
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

// addTLSAssets processes the given SafeTLSConfig and adds the referenced CA, certificate and key to the store.
func (s *Store) addTLSAssets(ctx context.Context, ns string, tlsConfig monitoringv1.SafeTLSConfig) error {
	var (
		err  error
		ca   string
		cert string
		key  string
	)

	ca, err = s.GetKey(ctx, ns, tlsConfig.CA)
	if err != nil {
		return errors.Wrap(err, "failed to get CA")
	}

	cert, err = s.GetKey(ctx, ns, tlsConfig.Cert)
	if err != nil {
		return errors.Wrap(err, "failed to get cert")
	}

	if tlsConfig.KeySecret != nil {
		key, err = s.GetSecretKey(ctx, ns, *tlsConfig.KeySecret)
		if err != nil {
			return errors.Wrap(err, "failed to get key")
		}
	}

	if ca != "" {
		block, _ := pem.Decode([]byte(ca))
		if block == nil {
			return errors.New("failed to decode CA certificate")
		}
		_, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return errors.Wrap(err, "failed to parse CA certificate")
		}
		s.TLSAssets[TLSAssetKeyFromSelector(ns, tlsConfig.CA)] = TLSAsset(ca)
	}

	if cert != "" && key != "" {
		_, err = tls.X509KeyPair([]byte(cert), []byte(key))
		if err != nil {
			return errors.Wrap(err, "failed to load X509 key pair")
		}
		s.TLSAssets[TLSAssetKeyFromSelector(ns, tlsConfig.Cert)] = TLSAsset(cert)
		s.TLSAssets[TLSAssetKeyFromSelector(ns, monitoringv1.SecretOrConfigMap{Secret: tlsConfig.KeySecret})] = TLSAsset(key)
	}

	return nil
}

// AddSafeTLSConfig validates the given SafeTLSConfig and adds it to the store.
func (s *Store) AddSafeTLSConfig(ctx context.Context, ns string, tlsConfig *monitoringv1.SafeTLSConfig) error {
	if tlsConfig == nil {
		return nil
	}

	err := tlsConfig.Validate()
	if err != nil {
		return errors.Wrap(err, "failed to validate TLS configuration")
	}

	return s.addTLSAssets(ctx, ns, *tlsConfig)
}

// AddTLSConfig validates the given TLSConfig and adds it to the store.
func (s *Store) AddTLSConfig(ctx context.Context, ns string, tlsConfig *monitoringv1.TLSConfig) error {
	if tlsConfig == nil {
		return nil
	}

	err := tlsConfig.Validate()
	if err != nil {
		return errors.Wrap(err, "failed to validate TLS configuration")
	}

	return s.addTLSAssets(ctx, ns, tlsConfig.SafeTLSConfig)
}

// AddBasicAuth processes the given *BasicAuth and adds the referenced credentials to the store.
func (s *Store) AddBasicAuth(ctx context.Context, ns string, ba *monitoringv1.BasicAuth, key string) error {
	if ba == nil {
		return nil
	}

	username, err := s.GetSecretKey(ctx, ns, ba.Username)
	if err != nil {
		return errors.Wrap(err, "failed to get basic auth username")
	}

	password, err := s.GetSecretKey(ctx, ns, ba.Password)
	if err != nil {
		return errors.Wrap(err, "failed to get basic auth password")
	}

	s.BasicAuthAssets[key] = BasicAuthCredentials{
		Username: username,
		Password: password,
	}

	return nil
}

// AddOAuth2 processes the given *OAuth2 and adds the referenced credentials to the store.
func (s *Store) AddOAuth2(ctx context.Context, ns string, oauth2 *monitoringv1.OAuth2, key string) error {
	if oauth2 == nil {
		return nil
	}

	if err := oauth2.Validate(); err != nil {
		return err
	}

	clientID, err := s.GetKey(ctx, ns, oauth2.ClientID)
	if err != nil {
		return errors.Wrap(err, "failed to get oauth2 client id")
	}

	clientSecret, err := s.GetSecretKey(ctx, ns, oauth2.ClientSecret)
	if err != nil {
		return errors.Wrap(err, "failed to get oauth2 client secret")
	}

	s.OAuth2Assets[key] = OAuth2Credentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	return nil
}

// AddToken processes the given SecretKeySelector and adds the referenced data to the store.
func (s *Store) addToken(ctx context.Context, ns string, sel v1.SecretKeySelector, key string) error {
	if sel.Name == "" {
		return nil
	}

	token, err := s.GetSecretKey(ctx, ns, sel)
	if err != nil {
		return errors.Wrap(err, "failed to get token from secret")
	}

	s.TokenAssets[key] = Token(token)

	return nil
}

func (s *Store) AddBearerToken(ctx context.Context, ns string, sel v1.SecretKeySelector, key string) error {
	err := s.addToken(ctx, ns, sel, key)
	if err != nil {
		return errors.Wrap(err, "failed to get bearer token")
	}
	return nil
}

func (s *Store) AddSafeAuthorizationCredentials(ctx context.Context, namespace string, auth *monitoringv1.SafeAuthorization, key string) error {
	if auth == nil || auth.Credentials == nil {
		return nil
	}

	if err := auth.Validate(); err != nil {
		return err
	}

	err := s.addToken(ctx, namespace, *auth.Credentials, key)
	if err != nil {
		return errors.Wrapf(err, "failed to get authorization token of type %s", auth.Type)
	}
	return nil
}

func (s *Store) AddAuthorizationCredentials(ctx context.Context, namespace string, auth *monitoringv1.Authorization, key string) error {
	if auth == nil || auth.Credentials == nil {
		return nil
	}

	if err := auth.Validate(); err != nil {
		return err
	}

	err := s.addToken(ctx, namespace, *auth.Credentials, key)
	if err != nil {
		return errors.Wrapf(err, "failed to get authorization token of type %s", auth.Type)
	}
	return nil
}

// AddSigV4 processes the SigV4 SecretKeySelectors and adds the SigV4 data to the store.
func (s *Store) AddSigV4(ctx context.Context, ns string, sigv4 *monitoringv1.Sigv4, key string) error {
	if sigv4 == nil || (sigv4.AccessKey == nil && sigv4.SecretKey == nil) {
		return nil
	}

	if sigv4.AccessKey == nil || sigv4.SecretKey == nil {
		return errors.New("both accessKey and secretKey should be provided")
	}

	sigV4Credentials := SigV4Credentials{}

	accessKey, err := s.GetSecretKey(ctx, ns, *sigv4.AccessKey)
	if err != nil {
		return errors.Wrap(err, "failed to read SigV4 access-key")
	}
	sigV4Credentials.AccessKeyID = accessKey

	secretKey, err := s.GetSecretKey(ctx, ns, *sigv4.SecretKey)
	if err != nil {
		return errors.Wrap(err, "failed to read SigV4 secret-key")
	}
	sigV4Credentials.SecretKeyID = secretKey

	s.SigV4Assets[key] = sigV4Credentials

	return nil
}

// GetKey processes the given SecretOrConfigMap selector and returns the referenced data.
func (s *Store) GetKey(ctx context.Context, namespace string, sel monitoringv1.SecretOrConfigMap) (string, error) {
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
func (s *Store) GetConfigMapKey(ctx context.Context, namespace string, sel v1.ConfigMapKeySelector) (string, error) {
	obj, exists, err := s.objStore.Get(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sel.Name,
			Namespace: namespace,
		},
	})
	if err != nil {
		return "", errors.Wrapf(err, "unexpected store error when getting configmap %q", sel.Name)
	}

	if !exists {
		cm, err := s.cmClient.ConfigMaps(namespace).Get(ctx, sel.Name, metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrapf(err, "unable to get configmap %q", sel.Name)
		}
		if err = s.objStore.Add(cm); err != nil {
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

// GetSecretKey processes the given SecretKeySelector and returns the referenced data.
func (s *Store) GetSecretKey(ctx context.Context, namespace string, sel v1.SecretKeySelector) (string, error) {
	obj, exists, err := s.objStore.Get(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sel.Name,
			Namespace: namespace,
		},
	})
	if err != nil {
		return "", errors.Wrapf(err, "unexpected store error when getting secret %q", sel.Name)
	}

	if !exists {
		secret, err := s.sClient.Secrets(namespace).Get(ctx, sel.Name, metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrapf(err, "unable to get secret %q", sel.Name)
		}
		if err = s.objStore.Add(secret); err != nil {
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
