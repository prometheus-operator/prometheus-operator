package log

import (
	"log/slog"
	"os"
	"reflect"
	"testing"
)

func TestParseFmt(t *testing.T) {
	type args struct {
		lvlOption slog.Level
		format    string
	}
	tests := []struct {
		name    string
		args    args
		want    *slog.Logger
		wantErr bool
	}{
		{
			name: "logfmt",
			args: args{
				lvlOption: slog.LevelDebug,
				format:    FormatLogFmt,
			},
			want:    slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
			wantErr: false,
		},
		{
			name: "json",
			args: args{
				lvlOption: slog.LevelDebug,
				format:    FormatJSON,
			},
			want:    slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFmt(tt.args.lvlOption, tt.args.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFmt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFmt() = %v, want %v", got, tt.want)
			}
		})
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
			got, err := ParseLevel(tt.args.lvl)
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
