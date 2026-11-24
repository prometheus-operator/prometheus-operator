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

package log

import (
	"io"
	"os"
	"strings"
	"testing"

	loglevel "github.com/go-kit/log/level"
	"github.com/stretchr/testify/require"
)

func TestNewLoggerFieldOrder(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer r.Close()
	defer w.Close()

	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	logger, err := NewLogger(Config{Level: LevelInfo, Format: FormatLogFmt})
	require.NoError(t, err)

	require.NoError(t, loglevel.Info(logger).Log("msg", "with-level"))
	require.NoError(t, logger.Log("msg", "no-level"))
	require.NoError(t, w.Close())

	out, err := io.ReadAll(r)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	require.Len(t, lines, 2)
	require.Regexp(t, `^ts=\S+ level=info caller=\S+ msg=with-level$`, lines[0])
	require.Regexp(t, `^ts=\S+ caller=\S+ msg=no-level$`, lines[1])
}
