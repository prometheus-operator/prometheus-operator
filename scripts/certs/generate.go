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

// Program to generate TLS materials for remote write end-to-end tests.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	certDirectory  = "../test/e2e/remote_write_certs"
	validityPeriod = 10 * 365 * 24 * time.Hour
)

func writePEMFile(file, typ string, b []byte) error {
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %v", file, err)
	}

	if err := pem.Encode(f, &pem.Block{Type: typ, Bytes: b}); err != nil {
		return fmt.Errorf("failed to write data to %s: %v", file, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing %s: %v", file, err)
	}

	log.Printf("wrote %s\n", file)
	return nil
}

func writeCert(name string, b []byte) error {
	return writePEMFile(filepath.Join(certDirectory, fmt.Sprintf("%s.crt", name)), "CERTIFICATE", b)
}

func writeKey(name string, priv *rsa.PrivateKey) error {
	b, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %v", err)
	}
	return writePEMFile(filepath.Join(certDirectory, fmt.Sprintf("%s.key", name)), "PRIVATE KEY", b)
}

type certificate struct {
	name    string
	subject string
}

func generateMaterials(ca, client certificate) error {
	now := time.Now()
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate CA serial number: %v", err)
	}

	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate CA private key: %v", err)
	}

	clientKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate client private key: %v", err)
	}

	// Generate the CA.
	caCert := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: ca.subject,
		},
		NotBefore:             now,
		NotAfter:              now.Add(validityPeriod),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		BasicConstraintsValid: true,
		DNSNames:              []string{ca.subject},
	}

	b, err := x509.CreateCertificate(rand.Reader, &caCert, &caCert, &caKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %v", err)
	}

	if err := writeCert(ca.name, b); err != nil {
		return fmt.Errorf("failed to write CA certificate: %v", err)
	}

	if err := writeKey(ca.name, caKey); err != nil {
		return fmt.Errorf("failed to write CA key: %v", err)
	}

	// Generate the client certificate/key.
	clientCert := x509.Certificate{
		SerialNumber: serialNumber.Add(serialNumber, big.NewInt(1)),
		Subject: pkix.Name{
			CommonName: client.subject,
		},
		NotBefore:             now,
		NotAfter:              now.Add(validityPeriod),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		BasicConstraintsValid: true,
		DNSNames:              []string{client.subject},
	}

	b, err = x509.CreateCertificate(rand.Reader, &clientCert, &caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create client certificate: %v", err)
	}

	if err := writeCert(client.name, b); err != nil {
		return fmt.Errorf("failed to write client certificate: %v", err)
	}

	if err := writeKey(client.name, clientKey); err != nil {
		return fmt.Errorf("failed to write client key: %v", err)
	}

	return nil
}

func main() {
	log.Println("Generating remote write certificates in", certDirectory)

	err := generateMaterials(
		certificate{name: "ca", subject: "caandserver.com"},
		certificate{name: "client", subject: "PrometheusRemoteWriteClient"},
	)
	if err != nil {
		log.Fatal(err)
	}

	err = generateMaterials(
		certificate{name: "bad_ca", subject: "badcaandserver.com"},
		certificate{name: "bad_client", subject: "BadPrometheusRemoteWriteClient"},
	)
	if err != nil {
		log.Fatal(err)
	}
}
