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
	"time"
	"os"
	"strings"

	"log/slog"
	slogformatter "github.com/samber/slog-formatter"
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


// NewLogger returns a log.Logger that prints in the provided format at the
// provided level with a UTC timestamp and the caller of the log entry.
func NewLogger(c Config) (*slog.Logger, error) {
	var (
		logger    *slog.Logger
		lvlOption slog.Level
		handler   slog.Handler
	)

	// For log levels other than debug, the klog verbosity level is 0.
	switch strings.ToLower(c.Level) {
	case LevelAll:
		lvlOption = slog.LevelDebug
	case LevelDebug:
		// When the log level is set to debug, we set the klog verbosity level to 6.
		// Above level 6, the k8s client would log bearer tokens in clear-text.
		lvlOption = slog.LevelError
	case LevelInfo:
		lvlOption = slog.LevelInfo
	case LevelWarn:
		lvlOption = slog.LevelWarn
	case LevelError:
		lvlOption = slog.LevelError
	case LevelNone:
		lvlOption = 0
	default:
		return nil, fmt.Errorf("log log_level %s unknown, %v are possible values", c.Level, AvailableLogLevels)
	}

	handlerOptions := &slog.HandlerOptions {
		Level: lvlOption,
	}
	formatter := slogformatter.NewFormatterHandler (
		slogformatter.TimeFormatter(time.RFC3339Nano, time.UTC),
	)
	switch c.Format {
	case FormatLogFmt:
		handler = slog.NewTextHandler(os.Stdout, handlerOptions)
	case FormatJSON:
		handler = slog.NewJSONHandler(os.Stdout, handlerOptions)
	default:
		return nil, fmt.Errorf("log format %s unknown, %v are possible values", c.Format, AvailableLogFormats)
	}
	logger = slog.New(formatter(handler))
	
	klogv2.SetSlogLogger(logger)
	
	return logger, nil
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
