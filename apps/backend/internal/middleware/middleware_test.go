package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/logger"
)

func TestRequestIDSetsHeaderGinContextAndRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/", func(c *gin.Context) {
		requestID := RequestIDFromContext(c)
		if requestID == "" {
			t.Fatal("request ID missing from Gin context")
		}
		if got := logger.RequestID(c.Request.Context()); got != requestID {
			t.Fatalf("request context ID = %q, want %q", got, requestID)
		}
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	requestID := recorder.Header().Get(RequestIDHeader)
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(requestID) {
		t.Fatalf("request ID %q is not an RFC 4122 v4 UUID", requestID)
	}
}

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(SecurityHeaders())
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	for header, want := range map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"Referrer-Policy":         "no-referrer",
		"Content-Security-Policy": contentSecurityPolicy,
	} {
		if got := recorder.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}

func TestRateLimitReturnsStandardTooManyRequestsResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	if err := engine.SetTrustedProxies(nil); err != nil {
		t.Fatalf("disable trusted proxies: %v", err)
	}
	engine.Use(RateLimit(NewIPRateLimiter(2)))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	for attempt := 1; attempt <= 3; attempt++ {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = "203.0.113.9:1234"
		engine.ServeHTTP(recorder, request)

		if attempt <= 2 && recorder.Code != http.StatusNoContent {
			t.Fatalf("attempt %d status = %d, want %d", attempt, recorder.Code, http.StatusNoContent)
		}
		if attempt == 3 {
			if recorder.Code != http.StatusTooManyRequests {
				t.Fatalf("attempt %d status = %d, want %d", attempt, recorder.Code, http.StatusTooManyRequests)
			}
			if got, want := recorder.Body.String(), `{"success":false,"message":"rate limit exceeded"}`; got != want {
				t.Fatalf("rate limit response = %s, want %s", got, want)
			}
			if recorder.Header().Get("Retry-After") == "" {
				t.Fatal("Retry-After header is missing")
			}
		}
	}
}

func TestIPRateLimiterRefillsTokens(t *testing.T) {
	limiter := NewIPRateLimiter(60)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }

	if allowed, _ := limiter.Allow("203.0.113.9"); !allowed {
		t.Fatal("first request should be allowed")
	}
	for attempt := 0; attempt < 59; attempt++ {
		if allowed, _ := limiter.Allow("203.0.113.9"); !allowed {
			t.Fatalf("initial token %d should be allowed", attempt+2)
		}
	}
	if allowed, _ := limiter.Allow("203.0.113.9"); allowed {
		t.Fatal("request after bucket exhaustion should be rejected")
	}

	now = now.Add(time.Second)
	if allowed, _ := limiter.Allow("203.0.113.9"); !allowed {
		t.Fatal("one token should refill after one second at 60 requests/minute")
	}
}

func TestRecoveryReturnsSanitizedResponseAndLogsStack(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(logger.NewJSONLogger(&logs, slog.LevelDebug))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	engine := gin.New()
	engine.Use(RequestID(), RequestLogger(), Recovery())
	engine.GET("/panic", func(*gin.Context) { panic("sensitive implementation detail") })

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if got, want := recorder.Body.String(), `{"success":false,"message":"Internal server error"}`; got != want {
		t.Fatalf("recovery response = %s, want %s", got, want)
	}
	if strings.Contains(recorder.Body.String(), "sensitive implementation detail") {
		t.Fatal("panic detail was exposed in response")
	}
	for _, field := range []string{`"msg":"panic recovered"`, `"stack_trace":`, `"request_id":"`} {
		if !strings.Contains(logs.String(), field) {
			t.Errorf("recovery log does not contain %s: %s", field, logs.String())
		}
	}
}

func TestRequestLoggerIncludesRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(logger.NewJSONLogger(&logs, slog.LevelDebug))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	engine := gin.New()
	if err := engine.SetTrustedProxies(nil); err != nil {
		t.Fatalf("disable trusted proxies: %v", err)
	}
	engine.Use(RequestID(), RequestLogger())
	engine.GET("/requests", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	request := httptest.NewRequest(http.MethodGet, "/requests", nil)
	request.RemoteAddr = "203.0.113.9:1234"
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	for _, field := range []string{
		`"request_id":"`,
		`"method":"GET"`,
		`"path":"/requests"`,
		`"status":204`,
		`"latency":`,
		`"client_ip":"203.0.113.9"`,
	} {
		if !strings.Contains(logs.String(), field) {
			t.Errorf("request log does not contain %s: %s", field, logs.String())
		}
	}
}
