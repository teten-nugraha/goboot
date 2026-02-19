package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ctxKey string

const TraceIDKey ctxKey = "traceId"

type FileLoggerOptions struct {
	Dir        string
	Filename   string
	Level      string // "debug","info","warn","error"
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
	AppName    string
}

// resolveDir memilih direktori yang writeable:
// 1) dari opsi (dir), 2) dari env LOG_DIR, 3) "./logs", 4) $TMPDIR/logs
func resolveDir(dir string) (string, error) {
	candidates := []string{}

	// 1) opsi eksplisit
	if dir != "" {
		candidates = append(candidates, dir)
	}
	// 2) env
	if v := os.Getenv("LOG_DIR"); v != "" {
		candidates = append(candidates, v)
	}
	// 3) default relative
	candidates = append(candidates, "./logs")
	// 4) tmp
	candidates = append(candidates, filepath.Join(os.TempDir(), "logs"))

	var lastErr error
	for _, d := range candidates {
		// Normalisasi: kalau diawali "~", expand ke home
		if strings.HasPrefix(d, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				d = filepath.Join(home, strings.TrimPrefix(d, "~"))
			}
		}
		// Kalau absolut yang mengarah ke root ("/logs"), biarkan tapi akan gagal jika read-only.
		if err := os.MkdirAll(d, 0o755); err != nil {
			lastErr = err
			continue
		}
		// Tes write by creating a small temp file then remove
		testFile := filepath.Join(d, ".write_test")
		if err := os.WriteFile(testFile, []byte("ok"), 0o644); err != nil {
			lastErr = err
			continue
		}
		_ = os.Remove(testFile)
		return d, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no valid log directory candidates")
	}
	return "", lastErr
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zap.DebugLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}

// NewFileLoggerWithOptions membuat zap logger dengan rotasi file dan fallback dir yang aman.
func NewFileLoggerWithOptions(opt FileLoggerOptions) (*zap.Logger, string, error) {
	// Defaults
	if opt.Filename == "" {
		opt.Filename = "app.log"
	}
	if opt.MaxSizeMB <= 0 {
		opt.MaxSizeMB = 50
	}
	if opt.MaxBackups < 0 {
		opt.MaxBackups = 7
	}
	if opt.MaxAgeDays <= 0 {
		opt.MaxAgeDays = 14
	}

	dir, err := resolveDir(opt.Dir)
	if err != nil {
		return nil, "", err
	}
	logPath := filepath.Join(dir, opt.Filename)

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    opt.MaxSizeMB,
		MaxBackups: opt.MaxBackups,
		MaxAge:     opt.MaxAgeDays,
		Compress:   opt.Compress,
	})

	encCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encCfg),
		w,
		parseLevel(opt.Level),
	)

	logger := zap.New(core, zap.AddCaller(), zap.Fields(zap.String("app", opt.AppName)))
	return logger, logPath, nil
}

func FromCtx(ctx context.Context, base *zap.Logger) *zap.Logger {
	if base == nil {
		return zap.NewNop()
	}
	if ctx == nil {
		return base
	}
	if v := ctx.Value(TraceIDKey); v != nil {
		if tid, ok := v.(string); ok && tid != "" {
			return base.With(zap.String("traceId", tid))
		}
	}
	return base
}

func WithFields(l *zap.Logger, fields ...zap.Field) *zap.Logger {
	if l == nil {
		return zap.NewNop()
	}
	return l.With(fields...)
}

func Track(l *zap.Logger, name string) func() {
	start := time.Now()
	return func() {
		if l != nil {
			l.Debug("timing", zap.String("op", name), zap.Duration("elapsed", time.Since(start)))
		}
	}
}
