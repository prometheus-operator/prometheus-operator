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
	"net"
	"net/http"
	"os"
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

		client := createHTTPClient(&timeoutDuration)

		if diff := deep.Equal(client, expectedClient); diff != nil {
			t.Errorf("found differences %v", diff)
		}
	})
}
