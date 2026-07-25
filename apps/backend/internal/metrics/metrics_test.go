package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCollectorExposesPrometheusHTTPAndPanicMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	collector := NewCollector()
	collector.RecordPanic()

	engine := gin.New()
	engine.Use(collector.Middleware())
	engine.GET("/work", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	engine.GET("/metrics", collector.Handler)

	workRecorder := httptest.NewRecorder()
	engine.ServeHTTP(workRecorder, httptest.NewRequest(http.MethodGet, "/work", nil))
	if workRecorder.Code != http.StatusNoContent {
		t.Fatalf("work status = %d, want %d", workRecorder.Code, http.StatusNoContent)
	}

	metricsRecorder := httptest.NewRecorder()
	engine.ServeHTTP(metricsRecorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if metricsRecorder.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, want %d", metricsRecorder.Code, http.StatusOK)
	}
	if contentType := metricsRecorder.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/plain") {
		t.Fatalf("Content-Type = %q, want Prometheus text format", contentType)
	}

	for _, metric := range []string{
		`opspilot_http_requests_total{method="GET",path="/work",status="204"} 1`,
		`opspilot_http_request_duration_seconds_bucket{method="GET",path="/work",le="+Inf"} 1`,
		"opspilot_http_active_requests",
		"opspilot_http_panics_total 1",
	} {
		if !strings.Contains(metricsRecorder.Body.String(), metric) {
			t.Errorf("metrics output does not contain %q:\n%s", metric, metricsRecorder.Body.String())
		}
	}
}
