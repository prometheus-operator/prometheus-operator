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
	"os"
	"reflect"
	"testing"
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

	var m map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := m["level"]; !ok {
		t.Fatalf("key level not found in the JSON")
	}

	if _, ok := m["ts"]; !ok {
		t.Fatalf("key ts not found in the JSON")
	}

	if _, ok := m["caller"]; !ok {
		t.Fatalf("key caller not found in the JSON")
	}
}

func TestParseFmt(t *testing.T) {
	handler, err := getHandlerFromFormat(FormatJSON, slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantJSONHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})

	if !reflect.DeepEqual(handler, wantJSONHandler) {
		t.Fatalf("handler not equal to wantJSONHandler")
	}

	handler, err = getHandlerFromFormat(FormatLogFmt, slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantTextHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})

	if !reflect.DeepEqual(handler, wantTextHandler) {
		t.Fatalf("handler not equal to wantTextHandler")
	}
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
			want:    -1,
			wantErr: false,
		},
		{
			name: "unknown",
			args: args{
				lvl: "unknown",
			},
			want:    -1,
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}
