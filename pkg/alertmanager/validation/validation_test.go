// Copyright 2021 The prometheus-operator Authors
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

package validation

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/prometheus/alertmanager/config"
)

func TestValidateUrl(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		expectErr    bool
		expectResult func() *config.URL
	}{
		{
			name:      "Test invalid url returns error",
			in:        "https://!^invalid.com",
			expectErr: true,
		},
		{
			name:      "Test missing scheme returns error",
			in:        "is.normally.valid",
			expectErr: true,
		},
		{
			name: "Test happy path",
			in:   "https://u:p@is.compliant.with.upstream.unmarshal",
			expectResult: func() *config.URL {
				u, _ := url.Parse("https://u:p@is.compliant.with.upstream.unmarshal")
				return &config.URL{URL: u}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := ValidateURL(tc.in)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			res := tc.expectResult()
			if !reflect.DeepEqual(u, res) {
				t.Fatalf("wanted %v but got %v", res, u)
			}
		})
	}
}
