// Package logger provides the application's structured logging helpers.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type requestIDContextKey struct{}

// Configure installs a JSON logger as the process-wide slog default. It is
// intentionally safe to call during application startup before any requests
// are served.
func Configure() *slog.Logger {
	configuredLogger := NewJSONLogger(os.Stdout, levelFromEnvironment())
	slog.SetDefault(configuredLogger)
	return configuredLogger
}

// NewJSONLogger creates a JSON logger. It is exported so applications and
// tests can direct structured logs to a chosen writer without changing the
// process-wide logger.
func NewJSONLogger(output io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: level,
	}))
}

// WithRequestID attaches a request ID to a context. All logging helpers in
// this package emit the value as request_id.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, requestIDContextKey{}, requestID)
}

// RequestID returns the request ID associated with ctx, if one exists.
func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

// Log emits a structured entry and guarantees that request_id is present,
// including for process-level logs where it is an empty string.
func Log(ctx context.Context, level slog.Level, message string, attributes ...slog.Attr) {
	if ctx == nil {
		ctx = context.Background()
	}

	attributes = append([]slog.Attr{slog.String("request_id", RequestID(ctx))}, attributes...)
	slog.Default().LogAttrs(ctx, level, message, attributes...)
}

// Debug emits a debug-level structured log entry.
func Debug(ctx context.Context, message string, attributes ...slog.Attr) {
	Log(ctx, slog.LevelDebug, message, attributes...)
}

// Info emits an info-level structured log entry.
func Info(ctx context.Context, message string, attributes ...slog.Attr) {
	Log(ctx, slog.LevelInfo, message, attributes...)
}

// Warn emits a warning-level structured log entry.
func Warn(ctx context.Context, message string, attributes ...slog.Attr) {
	Log(ctx, slog.LevelWarn, message, attributes...)
}

// Error emits an error-level structured log entry.
func Error(ctx context.Context, message string, attributes ...slog.Attr) {
	Log(ctx, slog.LevelError, message, attributes...)
}

func levelFromEnvironment() slog.Level {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL"))) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
