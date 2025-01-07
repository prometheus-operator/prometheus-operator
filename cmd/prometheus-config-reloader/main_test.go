// Copyright 2016 The prometheus-operator Authors
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
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"

	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

var cases = []struct {
	in  string
	out string
}{
	{"prometheus-0", "0"},
	{"prometheus-1", "1"},
	{"prometheus-10", "10"},
	{"prometheus-10a", ""},
	{"prometheus1", "1"},
	{"pro-10-metheus", ""},
}

func TestCreateOrdinalEnvVar(t *testing.T) {
	for _, tt := range cases {
		t.Run(tt.in, func(t *testing.T) {
			os.Setenv(operator.PodNameEnvVar, tt.in)
			s := createOrdinalEnvvar(operator.PodNameEnvVar)
			if os.Getenv(statefulsetOrdinalEnvvar) != tt.out {
				t.Errorf("got %v, want %s", s, tt.out)
			}
		})
	}
}

func TestCreateHTTPClient(t *testing.T) {
	t.Run("http-client-is-created-correctly", func(t *testing.T) {
		transport := (http.DefaultTransport.(*http.Transport)).Clone()

		transport.DialContext = (&net.Dialer{
			Timeout:   -1,
			KeepAlive: -1,
		}).DialContext

		transport.DisableKeepAlives = true
		transport.MaxConnsPerHost = transport.MaxIdleConnsPerHost

		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		expectedClient := http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		}

		timeoutDuration := 30 * time.Second

		client := createHTTPClient(timeoutDuration, nil, nil)

		if diff := deep.Equal(client, expectedClient); diff != nil {
			t.Errorf("found differences %v", diff)
		}
	})
}

func TestCreateHTTPClientWithBasicAuth(t *testing.T) {
	t.Run("http-client-created-without-basic-credential", func(t *testing.T) {
		// Given
		thirtySecondDuration := 30 * time.Second

		password := "password"
		username := "username"

		l, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			t.Errorf("fail to construct testing mock http server %v", err)
			return
		}

		srv := &http.Server{}
		srv.Handler = http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				t.Errorf("missing Authorization header")
				return
			}
			basicBase64 := strings.TrimPrefix(authHeader, "Basic ")
			decodeString, err := base64.StdEncoding.DecodeString(basicBase64)
			if err != nil {
				t.Errorf("invalid base64 basic authorization in header")
				return
			}
			userPasswordHeader := strings.Split(string(decodeString), ":")
			if len(userPasswordHeader) != 2 {
				t.Errorf("invalid base64 basic authorization in header, does not respect user:password format, %s", decodeString)
				return
			}

			if userPasswordHeader[0] != username {
				t.Errorf("invalid username in base64 basic authorization in header %s", userPasswordHeader[0])
				return
			}
			if userPasswordHeader[1] != password {
				t.Errorf("invalid password in base64 basic authorization in header %s", userPasswordHeader[1])
				return
			}
		})
		go func() { _ = srv.Serve(l) }()
		defer srv.Close()

		// When
		client := createHTTPClient(thirtySecondDuration, &username, &password)

		// Then
		URL := fmt.Sprintf("http://%s", l.Addr().String())
		_, err = client.Get(URL)
		if err != nil {
			t.Errorf("fail to execute mock http request %v", err)
			return
		}

		if reflect.TypeOf(client.Transport).String() != "*config.basicAuthRoundTripper" {
			t.Errorf("round tipper for basic auth not populated current instance type is %s", reflect.TypeOf(client.Transport))
			return
		}
	})
}
