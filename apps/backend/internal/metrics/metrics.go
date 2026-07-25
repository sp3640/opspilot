// Package metrics exposes a small, dependency-free Prometheus collector for
// the API's HTTP lifecycle metrics.
package metrics

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var durationBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

type requestKey struct {
	method string
	path   string
	status string
}

type durationKey struct {
	method string
	path   string
}

type durationHistogram struct {
	count   uint64
	sum     float64
	buckets []uint64
}

// Collector holds the metrics scoped to one HTTP server.
type Collector struct {
	activeRequests atomic.Int64
	panicCount     atomic.Uint64

	mu             sync.RWMutex
	requestCounts  map[requestKey]uint64
	durationValues map[durationKey]*durationHistogram
}

// NewCollector creates an independent collector suitable for one application
// instance or a test server.
func NewCollector() *Collector {
	return &Collector{
		requestCounts:  make(map[requestKey]uint64),
		durationValues: make(map[durationKey]*durationHistogram),
	}
}

// Middleware records each completed HTTP response. Use it outside recovery so
// panics are observed with their final 500 response status.
func (c *Collector) Middleware() gin.HandlerFunc {
	if c == nil {
		return func(context *gin.Context) { context.Next() }
	}

	return func(context *gin.Context) {
		startedAt := time.Now()
		c.activeRequests.Add(1)
		defer c.activeRequests.Add(-1)

		context.Next()

		status := context.Writer.Status()
		if status == 0 {
			status = http.StatusOK
		}
		path := context.FullPath()
		if path == "" {
			path = "unmatched"
		}
		c.Observe(context.Request.Method, path, status, time.Since(startedAt))
	}
}

// RecordPanic increments the recovered-panic counter.
func (c *Collector) RecordPanic() {
	if c != nil {
		c.panicCount.Add(1)
	}
}

// Observe records a completed request. It is exported for non-Gin HTTP
// integrations and focused tests.
func (c *Collector) Observe(method, path string, status int, duration time.Duration) {
	if c == nil {
		return
	}

	request := requestKey{
		method: method,
		path:   path,
		status: strconv.Itoa(status),
	}
	durationMetric := durationKey{method: method, path: path}
	durationSeconds := duration.Seconds()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.requestCounts[request]++
	histogram := c.durationValues[durationMetric]
	if histogram == nil {
		histogram = &durationHistogram{buckets: make([]uint64, len(durationBuckets))}
		c.durationValues[durationMetric] = histogram
	}
	histogram.count++
	histogram.sum += durationSeconds
	for index, boundary := range durationBuckets {
		if durationSeconds <= boundary {
			histogram.buckets[index]++
		}
	}
}

// Handler writes Prometheus' text exposition format. It intentionally has no
// authentication so a local Prometheus server can scrape it; protect access at
// the network layer in production.
func (c *Collector) Handler(context *gin.Context) {
	if c == nil {
		context.String(http.StatusServiceUnavailable, "metrics collector unavailable\n")
		return
	}

	context.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", c.prometheusText())
}

func (c *Collector) prometheusText() []byte {
	var buffer bytes.Buffer

	buffer.WriteString("# HELP opspilot_http_requests_total Total completed HTTP requests by route and status.\n")
	buffer.WriteString("# TYPE opspilot_http_requests_total counter\n")
	buffer.WriteString("# HELP opspilot_http_request_duration_seconds HTTP request duration in seconds.\n")
	buffer.WriteString("# TYPE opspilot_http_request_duration_seconds histogram\n")

	c.mu.RLock()
	requestKeys := make([]requestKey, 0, len(c.requestCounts))
	for key := range c.requestCounts {
		requestKeys = append(requestKeys, key)
	}
	sort.Slice(requestKeys, func(left, right int) bool {
		if requestKeys[left].method != requestKeys[right].method {
			return requestKeys[left].method < requestKeys[right].method
		}
		if requestKeys[left].path != requestKeys[right].path {
			return requestKeys[left].path < requestKeys[right].path
		}
		return requestKeys[left].status < requestKeys[right].status
	})
	for _, key := range requestKeys {
		fmt.Fprintf(
			&buffer,
			"opspilot_http_requests_total{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
			prometheusLabel(key.method),
			prometheusLabel(key.path),
			prometheusLabel(key.status),
			c.requestCounts[key],
		)
	}

	durationKeys := make([]durationKey, 0, len(c.durationValues))
	for key := range c.durationValues {
		durationKeys = append(durationKeys, key)
	}
	sort.Slice(durationKeys, func(left, right int) bool {
		if durationKeys[left].method != durationKeys[right].method {
			return durationKeys[left].method < durationKeys[right].method
		}
		return durationKeys[left].path < durationKeys[right].path
	})
	for _, key := range durationKeys {
		histogram := c.durationValues[key]
		for index, boundary := range durationBuckets {
			fmt.Fprintf(
				&buffer,
				"opspilot_http_request_duration_seconds_bucket{method=\"%s\",path=\"%s\",le=\"%s\"} %d\n",
				prometheusLabel(key.method),
				prometheusLabel(key.path),
				strconv.FormatFloat(boundary, 'g', -1, 64),
				histogram.buckets[index],
			)
		}
		fmt.Fprintf(
			&buffer,
			"opspilot_http_request_duration_seconds_bucket{method=\"%s\",path=\"%s\",le=\"+Inf\"} %d\n",
			prometheusLabel(key.method),
			prometheusLabel(key.path),
			histogram.count,
		)
		fmt.Fprintf(
			&buffer,
			"opspilot_http_request_duration_seconds_sum{method=\"%s\",path=\"%s\"} %s\n",
			prometheusLabel(key.method),
			prometheusLabel(key.path),
			strconv.FormatFloat(histogram.sum, 'g', -1, 64),
		)
		fmt.Fprintf(
			&buffer,
			"opspilot_http_request_duration_seconds_count{method=\"%s\",path=\"%s\"} %d\n",
			prometheusLabel(key.method),
			prometheusLabel(key.path),
			histogram.count,
		)
	}
	c.mu.RUnlock()

	buffer.WriteString("# HELP opspilot_http_active_requests Current active HTTP requests.\n")
	buffer.WriteString("# TYPE opspilot_http_active_requests gauge\n")
	fmt.Fprintf(&buffer, "opspilot_http_active_requests %d\n", c.activeRequests.Load())
	buffer.WriteString("# HELP opspilot_http_panics_total Total recovered HTTP panics.\n")
	buffer.WriteString("# TYPE opspilot_http_panics_total counter\n")
	fmt.Fprintf(&buffer, "opspilot_http_panics_total %d\n", c.panicCount.Load())

	return buffer.Bytes()
}

func prometheusLabel(value string) string {
	return strings.NewReplacer(
		"\\", "\\\\",
		"\n", "\\n",
		"\"", "\\\"",
	).Replace(value)
}
