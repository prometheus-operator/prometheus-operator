package log

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func NewLoggerSlog(c Config) (*slog.Logger, error) {
	var (
		logger *slog.Logger
	)

	lvlOption, err := ParseLevel(c.Level)
	if err != nil {
		return nil, err
	}

	logger, err = ParseFmt(lvlOption, c.Format)
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func ParseFmt(lvlOption slog.Level, format string) (*slog.Logger, error) {
	var logger *slog.Logger
	switch strings.ToLower(format) {
	case FormatLogFmt:
		h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: lvlOption,
		})
		logger = slog.New(h)
		return logger, nil
	case FormatJSON:
		h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: lvlOption,
		})
		logger = slog.New(h)
		return logger, nil
	default:
		return nil, fmt.Errorf("log format %s unknown, %v are possible values", format, AvailableLogFormats)
	}
}

func ParseLevel(lvl string) (slog.Level, error) {
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
