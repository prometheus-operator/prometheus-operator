// Copyright The prometheus-operator Authors
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
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// ExecAction returns an ExecAction probing the given URL.
// Prefer promtool (available in official Prometheus images, including distroless
// builds that lack sh/curl/wget) so listenLocal probes work without a shell.
// See https://github.com/prometheus-operator/prometheus-operator/issues/8605
func ExecAction(u string) *corev1.ExecAction {
	check := "healthy"
	if strings.Contains(u, "/-/ready") {
		check = "ready"
	}
	baseURL := u
	if i := strings.Index(u, "/-/"); i >= 0 {
		baseURL = u[:i]
	}
	return &corev1.ExecAction{
		Command: []string{
			"promtool",
			"check",
			check,
			"--url",
			baseURL,
		},
	}
}

func curlProber(u string) string {
	return fmt.Sprintf("curl --fail %s", u)
}

func wgetProber(u string) string {
	return fmt.Sprintf("wget -q -O /dev/null %s", u)
}
