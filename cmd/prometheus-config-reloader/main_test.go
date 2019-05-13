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
	"os"
	"testing"
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
			os.Setenv(statefulsetOrdinalFromEnvvarDefault, tt.in)
			s := createOrdinalEnvvar(statefulsetOrdinalFromEnvvarDefault)
			if os.Getenv(statefulsetOrdinalEnvvar) != tt.out {
				t.Errorf("got %v, want %s", s, tt.out)
			}
		})
	}
}
