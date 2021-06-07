// Copyright 2019 The prometheus-operator Authors
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

package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	certPath = flag.String("cert-path", "", "path to tls certificates for mutual TLS endpoint")
)

func main() {
	flag.Parse()

	if *certPath != "" {
		go func() {
			log.Fatal(mTLSEndpoint())
		}()
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if checkBasicAuth(w, r) {
			promhttp.Handler().ServeHTTP(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="MY REALM"`)
		w.WriteHeader(401)
		w.Write([]byte("401 Unauthorized\n"))
	})

	http.HandleFunc("/bearer-metrics", func(w http.ResponseWriter, r *http.Request) {
		if checkBearerAuth(w, r) {
			promhttp.Handler().ServeHTTP(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", `Bearer realm="MY REALM"`)
		w.WriteHeader(401)
		w.Write([]byte("401 Unauthorized\n"))
	})

	address := ":8080"

	fmt.Printf("listening for metric requests on '%v' protected via basic auth or bearer token\n", address)

	_ = http.ListenAndServe(address, nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, time.Now().String())
	fmt.Fprint(w, "\nAppVersion:"+os.Getenv("VERSION"))
}

func checkBasicAuth(w http.ResponseWriter, r *http.Request) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}

	return pair[0] == "user" && pair[1] == "pass"
}

func checkBearerAuth(w http.ResponseWriter, r *http.Request) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}

	fmt.Println(s[1])

	return s[1] == "abc"
}

func mTLSEndpoint() error {
	certPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile(path.Join(*certPath, "cert.pem"))
	if err != nil {
		return fmt.Errorf("failed to load certificate authority")
	}
	ok := certPool.AppendCertsFromPEM(pem)
	if !ok {
		return fmt.Errorf("failed to add certificate authority to certificate pool")
	}

	tlsConfig := &tls.Config{
		ClientCAs:  certPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	address := ":8081"

	server := &http.Server{
		Addr:      address,
		TLSConfig: tlsConfig,
		Handler:   promhttp.Handler(),
	}

	fmt.Printf("listening for metric requests on '%v' protected via mutual tls\n", address)

	return server.ListenAndServeTLS(path.Join(*certPath, "cert.pem"), path.Join(*certPath, "key.pem"))
}
