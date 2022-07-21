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
	"bytes"
	"testing"
)

func TestGzipConfig(t *testing.T) {
	var buf bytes.Buffer
	if err := GzipConfig(&buf, []byte("aaa")); err != nil {
		t.Error("failed to gzip config")
	}
	uncompressed, err := GunzipConfig(buf.Bytes())
	if err != nil {
		t.Error("failed to create string form ungzipped config")
	}

	if uncompressed != "aaa" {
		t.Errorf("gzip data is incorrect - %s", "foo")
	}
}
