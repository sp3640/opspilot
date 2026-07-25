package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/config"
)

func TestHealthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	health := NewHealthHandler(
		&config.Config{AppName: "OpsPilot", AppEnv: "test", Version: "test-version"},
		time.Now().Add(-time.Minute),
		func(context.Context) error { return nil },
	)
	health.SetInitialized(true)

	engine := gin.New()
	engine.GET("/health", health.HealthCheck)
	engine.GET("/ready", health.Ready)
	engine.GET("/live", health.Live)

	for _, test := range []struct {
		path       string
		wantStatus int
		contains   string
	}{
		{path: "/health", wantStatus: http.StatusOK, contains: `"database":"ok"`},
		{path: "/ready", wantStatus: http.StatusOK, contains: `"status":"ready"`},
		{path: "/live", wantStatus: http.StatusOK, contains: `"status":"ok"`},
	} {
		t.Run(test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if body := recorder.Body.String(); !strings.Contains(body, test.contains) {
				t.Fatalf("body = %s, want it to contain %s", body, test.contains)
			}
		})
	}
}

func TestHealthReportsDatabaseFailureWithoutFailingLiveness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	health := NewHealthHandler(
		&config.Config{AppName: "OpsPilot", AppEnv: "test", Version: "test-version"},
		time.Now(),
		func(context.Context) error { return errors.New("database unavailable") },
	)
	health.SetInitialized(true)

	engine := gin.New()
	engine.GET("/health", health.HealthCheck)
	engine.GET("/ready", health.Ready)
	engine.GET("/live", health.Live)

	for _, test := range []struct {
		path       string
		wantStatus int
	}{
		{path: "/health", wantStatus: http.StatusServiceUnavailable},
		{path: "/ready", wantStatus: http.StatusServiceUnavailable},
		{path: "/live", wantStatus: http.StatusOK},
	} {
		t.Run(test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
		})
	}
}
