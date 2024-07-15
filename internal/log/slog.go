// Copyright 2021 The prometheus-operator Authors
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
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

// NewLoggerSlog returns a *slog.Logger that prints in the provided format at the
// provided level with a UTC timestamp and the caller of the log entry.
func NewLoggerSlog(c Config) (*slog.Logger, error) {
	lvlOption, err := parseLevel(c.Level)
	if err != nil {
		return nil, err
	}

	handler, err := getHandlerFromFormat(c.Format, slog.HandlerOptions{
		Level:       lvlOption,
		AddSource:   true,
		ReplaceAttr: replaceSlogAttributes,
	})
	if err != nil {
		return nil, err
	}

	return slog.New(handler), nil
}

// replaceSlogAttributes replaces fields that were added by default by slog, but had different
// formats or key names in github.com/go-kit/log. The operator was originally implemented with go-kit/log,
// so we use these replacements to make the migration smoother.
func replaceSlogAttributes(_ []string, a slog.Attr) slog.Attr {
	if a.Key == "time" {
		return slog.Attr{
			Key:   "ts",
			Value: slog.StringValue(a.Value.Time().UTC().Format(time.RFC3339Nano)),
		}
	}

	if a.Key == "level" {
		return slog.Attr{
			Key:   "level",
			Value: slog.StringValue(strings.ToLower(a.Value.String())),
		}
	}

	if a.Key == "source" {
		return slog.Attr{
			Key:   "caller",
			Value: a.Value,
		}
	}

	return a
}

// getHandlerFromFormat returns a slog.Handler based on the provided format and slog options.
func getHandlerFromFormat(format string, opts slog.HandlerOptions) (slog.Handler, error) {
	var handler slog.Handler
	switch strings.ToLower(format) {
	case FormatLogFmt:
		handler = slog.NewTextHandler(os.Stdout, &opts)
		return handler, nil
	case FormatJSON:
		handler = slog.NewJSONHandler(os.Stdout, &opts)
		return handler, nil
	default:
		return nil, fmt.Errorf("log format %s unknown, %v are possible values", format, AvailableLogFormats)
	}
}

// parseLevel returns the slog.Level based on the provided string.
func parseLevel(lvl string) (slog.Level, error) {
	switch strings.ToLower(lvl) {
	case LevelAll:
		return slog.LevelDebug, nil
	case LevelDebug:
		return slog.LevelDebug, nil
	case LevelInfo:
		return slog.LevelInfo, nil
	case LevelWarn:
		return slog.LevelWarn, nil
	case LevelError:
		return slog.LevelError, nil
	case LevelNone:
		return -1, nil
	default:
		return -1, fmt.Errorf("log log_level %s unknown, %v are possible values", lvl, AvailableLogLevels)
	}
}
