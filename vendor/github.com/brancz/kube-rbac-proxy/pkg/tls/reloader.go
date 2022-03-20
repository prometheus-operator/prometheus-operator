/*
Copyright 2017 Frederic Branczyk All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tls

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

// CertReloader is the struct that parses a certificate/key pair,
// providing a goroutine safe GetCertificate method to retrieve the parsed content.
//
// The GetCertificate signature is compatible with https://golang.org/pkg/crypto/tls/#Config.GetCertificate
// and can be used to hot-reload a certificate/key pair.
//
// For hot-reloading the Watch method must be started explicitly.
type CertReloader struct {
	certPath, keyPath string
	interval          time.Duration

	mu              sync.RWMutex // protects the fields below
	cert            *tls.Certificate
	certRaw, keyRaw []byte
}

func NewCertReloader(certPath, keyPath string, interval time.Duration) (*CertReloader, error) {
	r := &CertReloader{
		certPath: certPath,
		keyPath:  keyPath,
		interval: interval,
	}

	if err := r.reload(); err != nil {
		return nil, fmt.Errorf("error loading certificates: %v", err)
	}

	return r, nil
}

// Watch watches the configured certificate and key path and blocks the current goroutine
// until the scenario context is done or an error occurred during reloading.
func (r *CertReloader) Watch(ctx context.Context) error {
	t := time.NewTicker(r.interval)

	for {
		select {
		case <-t.C:
		case <-ctx.Done():
			return nil
		}

		if err := r.reload(); err != nil {
			return fmt.Errorf("reloading failed: %v", err)
		}
	}
}

func (r *CertReloader) reload() error {
	certRaw, err := ioutil.ReadFile(r.certPath)
	if err != nil {
		return fmt.Errorf("error loading certificate: %v", err)
	}

	keyRaw, err := ioutil.ReadFile(r.keyPath)
	if err != nil {
		return fmt.Errorf("error loading key: %v", err)
	}

	r.mu.RLock()
	equal := bytes.Equal(keyRaw, r.keyRaw) && bytes.Equal(certRaw, r.certRaw)
	r.mu.RUnlock()

	if equal {
		return nil
	}

	klog.V(4).Info("reloading key ", r.keyPath, " certificate ", r.certPath)

	cert, err := tls.X509KeyPair(certRaw, keyRaw)
	if err != nil {
		return fmt.Errorf("error parsing certificate: %v", err)
	}

	r.mu.Lock()
	r.cert = &cert
	r.certRaw = certRaw
	r.keyRaw = keyRaw
	r.mu.Unlock()

	return nil
}

// GetCertificate returns the current valid certificate.
// The ClientHello message is ignored
// and is just there to be compatible with https://golang.org/pkg/crypto/tls/#Config.GetCertificate.
func (r *CertReloader) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.cert, nil
}
