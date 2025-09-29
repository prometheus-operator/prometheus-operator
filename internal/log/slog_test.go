// Copyright The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestReplaceAttribute validates if all attributes that were replaced are present in the slog.Logger output.
func TestReplaceAttributes(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: replaceSlogAttributes,
	})

	l := slog.New(h)

	l.Info("test")

	var m map[string]any
	err := json.Unmarshal(buf.Bytes(), &m)
	require.NoError(t, err)

	require.Contains(t, m, "level")
	require.Contains(t, m, "msg")
	require.Contains(t, m, "caller")
}

func TestParseFmt(t *testing.T) {
	handler, err := getHandlerFromFormat(FormatJSON, slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})

	require.NoError(t, err)

	wantHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})

	require.Equal(t, wantHandler, handler)

	handler, err = getHandlerFromFormat(FormatLogFmt, slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})

	require.NoError(t, err)

	wantTextHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})

	require.Equal(t, wantTextHandler, handler)
}

func TestParseLevel(t *testing.T) {
	type args struct {
		lvl string
	}
	tests := []struct {
		name    string
		args    args
		want    slog.Level
		wantErr bool
	}{
		{
			name: "all",
			args: args{
				lvl: LevelAll,
			},
			want:    slog.LevelDebug,
			wantErr: false,
		},
		{
			name: "debug",
			args: args{
				lvl: LevelDebug,
			},
			want:    slog.LevelDebug,
			wantErr: false,
		},
		{
			name: "info",
			args: args{
				lvl: LevelInfo,
			},
			want:    slog.LevelInfo,
			wantErr: false,
		},
		{
			name: "warn",
			args: args{
				lvl: LevelWarn,
			},
			want:    slog.LevelWarn,
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				lvl: LevelError,
			},
			want:    slog.LevelError,
			wantErr: false,
		},
		{
			name: "none",
			args: args{
				lvl: LevelNone,
			},
			want:    math.MaxInt,
			wantErr: false,
		},
		{
			name: "unknown",
			args: args{
				lvl: "unknown",
			},
			want:    math.MaxInt,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLevel(tt.args.lvl)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
