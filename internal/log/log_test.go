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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderedLogger(t *testing.T) {
	tests := []struct {
		name  string
		input []any
		want  []any
	}{
		{
			name:  "level before timestamp",
			input: []any{"level", "info", "ts", "now", "caller", "main.go:1"},
			want:  []any{"ts", "now", "level", "info", "caller", "main.go:1"},
		},
		{
			name:  "no level",
			input: []any{"ts", "now", "caller", "main.go:1"},
			want:  []any{"ts", "now", "caller", "main.go:1"},
		},
		{
			name:  "level without timestamp",
			input: []any{"level", "info", "caller", "main.go:1"},
			want:  []any{"level", "info", "caller", "main.go:1"},
		},
		{
			name:  "fewer than four items",
			input: []any{"level", "info", "ts"},
			want:  []any{"level", "info", "ts"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := &keyvalsLogger{}
			logger := orderedLogger{next: next}

			require.NoError(t, logger.Log(tt.input...))
			require.Equal(t, tt.want, next.keyvals)
		})
	}
}

type keyvalsLogger struct {
	keyvals []any
}

func (l *keyvalsLogger) Log(keyvals ...any) error {
	l.keyvals = append([]any(nil), keyvals...)
	return nil
}
