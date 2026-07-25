package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/logger"
	"github.com/sp3640/opspilot/backend/internal/response"
)

// PanicRecorder receives a notification when Recovery catches a panic. The
// metrics collector implements this interface without creating a package cycle.
type PanicRecorder interface {
	RecordPanic()
}

// Recovery catches unhandled panics, logs diagnostic details internally, and
// sends the application's standard error response without exposing the panic
// or stack trace to clients.
func Recovery(recorders ...PanicRecorder) gin.HandlerFunc {
	var recorder PanicRecorder
	if len(recorders) > 0 {
		recorder = recorders[0]
	}

	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if recorder != nil {
					recorder.RecordPanic()
				}
				logger.Error(
					c.Request.Context(),
					"panic recovered",
					slog.String("panic", fmt.Sprint(recovered)),
					slog.String("stack_trace", string(debug.Stack())),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.String("client_ip", c.ClientIP()),
				)

				if !c.Writer.Written() {
					response.Error(c, http.StatusInternalServerError, "Internal server error")
				}

				c.Abort()
			}
		}()

		c.Next()
	}
}
