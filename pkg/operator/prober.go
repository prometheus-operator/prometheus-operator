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

	v1 "k8s.io/api/core/v1"
)

// ExecAction returns an ExecAction probing the given URL.
func ExecAction(u string) *v1.ExecAction {
	return &v1.ExecAction{
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

func curlProber(u string) string {
	return fmt.Sprintf("curl --fail %s", u)
}

func wgetProber(u string) string {
	return fmt.Sprintf("wget -q -O /dev/null %s", u)
}
