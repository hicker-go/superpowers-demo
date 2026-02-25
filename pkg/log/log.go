/*
Package log provides a zap-based structured logger with configurable level,
output destination (stdout or file), and file rotation.

Config fields align with configs/settings.yaml:
  - level: debug, info, warn, error
  - output: stdout or file
  - file: path when output is file (required when output=file)
  - max_size: max size in MB before rotation
  - max_age: days to retain rotated logs
  - max_backups: number of rotated files to keep
*/
package log

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config holds logger configuration matching settings.yaml.
type Config struct {
	Level      string `mapstructure:"level"`       // debug, info, warn, error
	Output     string `mapstructure:"output"`      // stdout or file
	File       string `mapstructure:"file"`        // path when output=file
	MaxSize    int    `mapstructure:"max_size"`    // MB before rotation
	MaxAge     int    `mapstructure:"max_age"`     // days to retain
	MaxBackups int    `mapstructure:"max_backups"` // number of rotated files
}

// Logger is a minimal interface for structured logging; satisfied by *zap.Logger.
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Sync() error
}

// ensure *zap.Logger implements Logger
var _ Logger = (*zap.Logger)(nil)

// New builds a zap.Logger from config.
// When output is "file", File must be non-empty and rotation uses max_size, max_age, max_backups.
func New(cfg *Config) (*zap.Logger, error) {
	if cfg == nil {
		cfg = &Config{Level: "info", Output: "stdout"}
	}

	level := parseLevel(cfg.Level)
	encCfg := zap.NewProductionEncoderConfig()
	enc := zapcore.NewJSONEncoder(encCfg)

	var ws zapcore.WriteSyncer
	switch strings.ToLower(strings.TrimSpace(cfg.Output)) {
	case "stdout", "":
		ws = zapcore.AddSync(os.Stdout)
	case "file":
		if cfg.File == "" {
			return nil, fmt.Errorf("log: output is file but file path is empty")
		}
		lj := &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
		}
		ws = zapcore.AddSync(lj)
	default:
		return nil, fmt.Errorf("log: unknown output %q", cfg.Output)
	}

	core := zapcore.NewCore(enc, ws, level)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

func parseLevel(s string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return zapcore.DebugLevel
	case "info", "":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
