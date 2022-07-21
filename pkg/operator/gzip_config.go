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

package operator

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
)

func GzipConfig(w io.Writer, conf []byte) error {
	buf := gzip.NewWriter(w)
	defer buf.Close()
	if _, err := buf.Write(conf); err != nil {
		return err
	}
	return nil
}

func GunzipConfig(b []byte) (string, error) {
	buf := bytes.NewBuffer(b)
	reader, err := gzip.NewReader(buf)
	if err != nil {
		return "", err
	}
	uncompressed := new(strings.Builder)
	_, err = io.Copy(uncompressed, reader)
	if err != nil {
		return "", err
	}
	return uncompressed.String(), nil
}
