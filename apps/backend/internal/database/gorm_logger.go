package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/sp3640/opspilot/backend/internal/logger"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// structuredGormLogger prevents GORM from bypassing the application's JSON
// logger. It intentionally avoids logging SQL text because bound values can
// contain credentials or user-provided content.
type structuredGormLogger struct {
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

func newStructuredGormLogger() gormlogger.Interface {
	return structuredGormLogger{
		level:         gormlogger.Error,
		slowThreshold: time.Second,
	}
}

func (l structuredGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	l.level = level
	return l
}

func (l structuredGormLogger) Info(ctx context.Context, message string, values ...interface{}) {
	if l.level < gormlogger.Info {
		return
	}
	logger.Info(ctx, "gorm", slog.String("message", fmt.Sprintf(message, values...)))
}

func (l structuredGormLogger) Warn(ctx context.Context, message string, values ...interface{}) {
	if l.level < gormlogger.Warn {
		return
	}
	logger.Warn(ctx, "gorm", slog.String("message", fmt.Sprintf(message, values...)))
}

func (l structuredGormLogger) Error(ctx context.Context, message string, values ...interface{}) {
	if l.level < gormlogger.Error {
		return
	}
	logger.Error(ctx, "gorm", slog.String("message", fmt.Sprintf(message, values...)))
}

func (l structuredGormLogger) Trace(ctx context.Context, startedAt time.Time, fc func() (string, int64), err error) {
	if l.level == gormlogger.Silent {
		return
	}

	elapsed := time.Since(startedAt)
	_, rows := fc()
	attributes := []slog.Attr{
		slog.Duration("latency", elapsed),
		slog.Int64("rows", rows),
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		if l.level >= gormlogger.Error {
			attributes = append(attributes, slog.Any("error", err))
			logger.Error(ctx, "database query failed", attributes...)
		}
		return
	}

	if l.slowThreshold > 0 && elapsed > l.slowThreshold && l.level >= gormlogger.Warn {
		logger.Warn(ctx, "slow database query", attributes...)
	}
}
