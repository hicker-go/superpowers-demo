/*
Package log provides a zap-based structured logger with configurable level,
output destination (stdout, file, or both), and file rotation.

Config fields align with configs/settings.yaml:
  - level: debug, info, warn, error
  - output: stdout, file, or both (stdout+file)
  - file: path for file output (required when output includes file)
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
	Level      string `mapstructure:"level"`        // debug, info, warn, error
	Output     string `mapstructure:"output"`       // stdout, file, or both
	File       string `mapstructure:"file"`        // path when output includes file
	MaxSize    int    `mapstructure:"max_size"`     // MB before rotation
	MaxAge     int    `mapstructure:"max_age"`      // days to retain
	MaxBackups int    `mapstructure:"max_backups"`  // number of rotated files
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
// Output: "stdout" (console only), "file" (file only), "both" (console and file).
// When output includes file, File must be non-empty; rotation uses max_size, max_age, max_backups.
func New(cfg *Config) (*zap.Logger, error) {
	if cfg == nil {
		cfg = &Config{Level: "info", Output: "stdout"}
	}

	level := parseLevel(cfg.Level)
	encCfg := zap.NewProductionEncoderConfig()
	enc := zapcore.NewJSONEncoder(encCfg)

	out := strings.ToLower(strings.TrimSpace(cfg.Output))
	if out == "" {
		out = "stdout"
	}

	var writers []zapcore.WriteSyncer

	// Console output
	if out == "stdout" || out == "both" {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	// File output
	if out == "file" || out == "both" {
		if cfg.File == "" {
			return nil, fmt.Errorf("log: output requires file but file path is empty")
		}
		lj := &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
		}
		writers = append(writers, zapcore.AddSync(lj))
	}

	if len(writers) == 0 {
		return nil, fmt.Errorf("log: unknown output %q (use stdout, file, or both)", cfg.Output)
	}

	ws := zapcore.NewMultiWriteSyncer(writers...)
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
