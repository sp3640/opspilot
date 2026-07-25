package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/logger"
)

const healthCheckTimeout = 2 * time.Second

// DatabasePinger is the small dependency required by operational endpoints.
// Keeping it as a function makes health checks straightforward to test and
// avoids coupling handlers to a particular database implementation.
type DatabasePinger func(context.Context) error

// HealthHandler owns the process-level health state exposed to orchestrators.
type HealthHandler struct {
	appName     string
	environment string
	version     string
	startedAt   time.Time
	ping        DatabasePinger
	initialized atomic.Bool
}

// NewHealthHandler creates handlers for /health, /ready, and /live.
func NewHealthHandler(cfg *config.Config, startedAt time.Time, ping DatabasePinger) *HealthHandler {
	if startedAt.IsZero() {
		startedAt = time.Now()
	}

	handler := &HealthHandler{
		startedAt: startedAt,
		ping:      ping,
	}
	if cfg != nil {
		handler.appName = cfg.AppName
		handler.environment = cfg.AppEnv
		handler.version = cfg.Version
	}

	return handler
}

// SetInitialized controls readiness after startup dependencies have completed.
func (h *HealthHandler) SetInitialized(initialized bool) {
	if h == nil {
		return
	}
	h.initialized.Store(initialized)
}

// HealthCheck reports database reachability as well as process metadata.
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	databaseReachable, err := h.databaseReachable(c.Request.Context())
	status := http.StatusOK
	state := "ok"
	databaseState := "ok"
	if !databaseReachable {
		status = http.StatusServiceUnavailable
		state = "degraded"
		databaseState = "unavailable"
		h.logDatabaseFailure(c, err)
	}

	c.JSON(status, gin.H{
		"status":         state,
		"service":        h.appName,
		"environment":    h.environment,
		"version":        h.version,
		"database":       databaseState,
		"uptime_seconds": int64(time.Since(h.startedAt).Seconds()),
		"timestamp":      time.Now().UTC(),
	})
}

// Ready reports whether startup has completed and PostgreSQL is reachable.
func (h *HealthHandler) Ready(c *gin.Context) {
	initialized := h != nil && h.initialized.Load()
	databaseReachable, err := h.databaseReachable(c.Request.Context())
	if !initialized || !databaseReachable {
		if !databaseReachable {
			h.logDatabaseFailure(c, err)
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":      "not_ready",
			"initialized": initialized,
			"database":    databaseStatus(databaseReachable),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "ready",
		"initialized": true,
		"database":    "ok",
	})
}

// Live intentionally does not depend on downstream dependencies: if the HTTP
// process can answer this request, it is alive and should not be restarted.
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *HealthHandler) databaseReachable(parent context.Context) (bool, error) {
	if h == nil || h.ping == nil {
		return false, errors.New("database health check is not configured")
	}

	ctx, cancel := context.WithTimeout(parent, healthCheckTimeout)
	defer cancel()
	if err := h.ping(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (h *HealthHandler) logDatabaseFailure(c *gin.Context, err error) {
	if err == nil {
		return
	}

	logger.Warn(
		c.Request.Context(),
		"health check database dependency unavailable",
		slog.Any("error", err),
	)
}

func databaseStatus(reachable bool) string {
	if reachable {
		return "ok"
	}
	return "unavailable"
}
