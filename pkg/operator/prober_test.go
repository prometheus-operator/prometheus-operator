// Copyright 2022 The prometheus-operator Authors
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

package operator

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
)

func TestProbers(t *testing.T) {
	for _, tc := range []struct {
		code int
		err  bool
	}{
		{
			code: http.StatusOK,
			err:  false,
		},
		{
			code: http.StatusServiceUnavailable,
			err:  true,
		},
	} {
		for _, p := range []struct {
			name   string
			prober func(string) string
		}{
			{
				name:   "curl",
				prober: CurlProber,
			},
			{
				name:   "wget",
				prober: WgetProber,
			},
		} {
			t.Run(fmt.Sprintf("%d-%s", tc.code, p.name), func(t *testing.T) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.code)
				}))
				defer ts.Close()

				args := strings.Split(p.prober(ts.URL), " ")
				if _, err := exec.LookPath(args[0]); err != nil {
					t.Skipf("%s: %v", args[0], err)
				}

				cmd := exec.Command(args[0], args[1:]...)
				b, err := cmd.CombinedOutput()
				if tc.err {
					if err == nil {
						t.Logf("%s: %s", strings.Join(args, " "), string(b))
						t.Fatal("expecting error but got nil")
					}
					return
				}

				if err != nil {
					t.Logf("%s: %s", strings.Join(args, " "), string(b))
					t.Fatalf("expecting no error but got %v", err)
				}
			})
		}

	}
}
