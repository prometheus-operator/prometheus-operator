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
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/log"
	loglevel "github.com/go-kit/log/level"
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

// NewLogger returns a log.Logger that prints in the provided format at the
// provided level with a UTC timestamp and the caller of the log entry.
func NewLogger(level string, format string) (log.Logger, error) {
	var (
		logger    log.Logger
		lvlOption loglevel.Option
	)

	switch strings.ToLower(level) {
	case LevelAll:
		lvlOption = loglevel.AllowAll()
	case LevelDebug:
		lvlOption = loglevel.AllowDebug()
	case LevelInfo:
		lvlOption = loglevel.AllowInfo()
	case LevelWarn:
		lvlOption = loglevel.AllowWarn()
	case LevelError:
		lvlOption = loglevel.AllowError()
	case LevelNone:
		lvlOption = loglevel.AllowNone()
	default:
		return nil, fmt.Errorf("log log_level %s unknown, %v are possible values", level, AvailableLogLevels)
	}

	switch format {
	case FormatLogFmt:
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	case FormatJSON:
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	default:
		return nil, fmt.Errorf("log format %s unknown, %v are possible values", format, AvailableLogFormats)
	}

	logger = loglevel.NewFilter(logger, lvlOption)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
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
