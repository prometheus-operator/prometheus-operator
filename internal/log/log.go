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

// nolint
package log

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"log/slog"

	klogv2 "k8s.io/klog/v2"
)

const (
	LevelAll   = "all"
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelNone  = "none"
)

const (
	FormatLogFmt = "logfmt"
	FormatJSON   = "json"
)

type Config struct {
	Level  string
	Format string
}

func RegisterFlags(fs *flag.FlagSet, c *Config) {
	fs.StringVar(&c.Level, "log-level", "info", fmt.Sprintf("Log level to use. Possible values: %s", strings.Join(AvailableLogLevels, ", ")))
	fs.StringVar(&c.Format, "log-format", "logfmt", fmt.Sprintf("Log format to use. Possible values: %s", strings.Join(AvailableLogFormats, ", ")))
}

// NewLogger returns a *slog.Logger that prints in the provided format at the
// provided level with a UTC timestamp and the caller of the log entry.
func NewLogger(c Config) (*slog.Logger, error) {
	var (
		logger    *slog.Logger
		lvlOption slog.Level
		handler   slog.Handler
	)

	lvlOption, err := slogLevelFromString(c.Level)
	if err != nil {
		return nil, err
	}

	handlerOptions := &slog.HandlerOptions{
		Level:       lvlOption,
		AddSource:   true,
		ReplaceAttr: replaceSlogAttributes,
	}

	switch c.Format {
	case FormatLogFmt:
		handler = slog.NewTextHandler(os.Stdout, handlerOptions)
	case FormatJSON:
		handler = slog.NewJSONHandler(os.Stdout, handlerOptions)
	default:
		return nil, fmt.Errorf("log format %s unknown, %v are possible values", c.Format, AvailableLogFormats)
	}
	logger = slog.New(handler)

	klogv2.SetSlogLogger(logger)

	return logger, nil
}

func slogLevelFromString(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case LevelAll:
		return slog.LevelDebug, nil
	case LevelDebug:
		// When the log level is set to debug, we set the klog verbosity level to 6.
		// Above level 6, the k8s client would log bearer tokens in clear-text.
		return slog.LevelDebug, nil
	case LevelInfo:
		return slog.LevelInfo, nil
	case LevelWarn:
		return slog.LevelWarn, nil
	case LevelError:
		return slog.LevelError, nil
	case LevelNone:
		return 1, nil
	default:
		return 1, fmt.Errorf("log log_level %s unknown, %v are possible values", level, AvailableLogLevels)
	}
}

func replaceSlogAttributes(groups []string, a slog.Attr) slog.Attr {
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

// AvailableLogLevels is a list of supported logging levels
var AvailableLogLevels = []string{
	LevelAll,
	LevelDebug,
	LevelInfo,
	LevelWarn,
	LevelError,
	LevelNone,
}

// AvailableLogFormats is a list of supported log formats
var AvailableLogFormats = []string{
	FormatLogFmt,
	FormatJSON,
}
