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

// ExecAction returns an ExecAction that probes the given URL via curl or wget
// under a shell. Used for the config-reloader listenLocal probes and for
// Prometheus versions that predate promtool check healthy|ready (2.44.0).
func ExecAction(u string) *corev1.ExecAction {
	return &corev1.ExecAction{
		Command: []string{
			"sh",
			"-c",
			fmt.Sprintf(
				`if [ -x "$(command -v curl)" ]; then exec %s; elif [ -x "$(command -v wget)" ]; then exec %s; else exit 1; fi`,
				curlProber(u),
				wgetProber(u),
			),
		},
	}
}

// PromtoolExecAction returns an ExecAction that probes the given URL with
// promtool check healthy|ready. promtool ships in official Prometheus images
// (including distroless builds that lack sh/curl/wget). Requires Prometheus
// 2.44.0 or later. See https://github.com/prometheus-operator/prometheus-operator/issues/8605
func PromtoolExecAction(u string) *corev1.ExecAction {
	check := "healthy"
	if strings.Contains(u, "/-/ready") {
		check = "ready"
	}
	// promtool expects the base URL (scheme + host[:port]), not the probe path.
	baseURL, _, _ := strings.Cut(u, "/-/")
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
