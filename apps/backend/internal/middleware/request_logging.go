package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/logger"
)

// RequestLogger emits one structured entry for every completed request. It is
// intended to run outside Recovery so that recovered panics are recorded with
// their final 500 status.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		attributes := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(startedAt)),
			slog.String("client_ip", c.ClientIP()),
		}
		if len(c.Errors) > 0 {
			attributes = append(attributes, slog.String("errors", c.Errors.String()))
		}

		switch {
		case c.Writer.Status() >= 500:
			logger.Error(c.Request.Context(), "http request completed", attributes...)
		case c.Writer.Status() >= 400:
			logger.Warn(c.Request.Context(), "http request completed", attributes...)
		default:
			logger.Info(c.Request.Context(), "http request completed", attributes...)
		}
	}
}
