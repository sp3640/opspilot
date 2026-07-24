package middleware

import (
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/response"
)

const (
	// DefaultRequestsPerMinute is the production-safe default per client IP.
	DefaultRequestsPerMinute = 100
	maxTrackedClients        = 10000
	clientIdleTTL            = 2 * time.Minute
	cleanupInterval          = time.Minute
)

type rateLimitClient struct {
	tokens   float64
	lastSeen time.Time
}

// IPRateLimiter is an in-memory token bucket limiter keyed by client IP. It
// has no background goroutine and periodically removes idle clients while
// handling requests, so it is safe to use during graceful shutdown.
type IPRateLimiter struct {
	mu          sync.Mutex
	clients     map[string]*rateLimitClient
	perMinute   int
	now         func() time.Time
	lastCleanup time.Time
}

// NewIPRateLimiter creates a limiter that permits requestsPerMinute requests
// per client IP. Non-positive values fall back to the documented default.
func NewIPRateLimiter(requestsPerMinute int) *IPRateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = DefaultRequestsPerMinute
	}

	return &IPRateLimiter{
		clients:   make(map[string]*rateLimitClient),
		perMinute: requestsPerMinute,
		now:       time.Now,
	}
}

// Allow reports whether an IP can make one request now and returns the number
// of seconds a rejected client should wait before retrying.
func (l *IPRateLimiter) Allow(clientIP string) (allowed bool, retryAfterSeconds int) {
	if clientIP == "" {
		clientIP = "unknown"
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.cleanup(now)

	client, found := l.clients[clientIP]
	if !found {
		l.makeRoom()
		client = &rateLimitClient{
			tokens:   float64(l.perMinute),
			lastSeen: now,
		}
		l.clients[clientIP] = client
	}

	elapsed := now.Sub(client.lastSeen).Seconds()
	if elapsed > 0 {
		client.tokens = math.Min(
			float64(l.perMinute),
			client.tokens+elapsed*float64(l.perMinute)/time.Minute.Seconds(),
		)
	}
	client.lastSeen = now

	if client.tokens >= 1 {
		client.tokens--
		return true, 0
	}

	retryAfter := int(math.Ceil((1 - client.tokens) * time.Minute.Seconds() / float64(l.perMinute)))
	if retryAfter < 1 {
		retryAfter = 1
	}
	return false, retryAfter
}

func (l *IPRateLimiter) cleanup(now time.Time) {
	if !l.lastCleanup.IsZero() && now.Sub(l.lastCleanup) < cleanupInterval {
		return
	}

	for clientIP, client := range l.clients {
		if now.Sub(client.lastSeen) > clientIdleTTL {
			delete(l.clients, clientIP)
		}
	}
	l.lastCleanup = now
}

func (l *IPRateLimiter) makeRoom() {
	if len(l.clients) < maxTrackedClients {
		return
	}

	var oldestClientIP string
	var oldestSeen time.Time
	for clientIP, client := range l.clients {
		if oldestSeen.IsZero() || client.lastSeen.Before(oldestSeen) {
			oldestClientIP = clientIP
			oldestSeen = client.lastSeen
		}
	}
	if oldestClientIP != "" {
		delete(l.clients, oldestClientIP)
	}
}

// RateLimit wraps an IPRateLimiter in Gin middleware. A 429 response uses the
// same API response shape as the rest of the application.
func RateLimit(limiter *IPRateLimiter) gin.HandlerFunc {
	if limiter == nil {
		limiter = NewIPRateLimiter(DefaultRequestsPerMinute)
	}

	return func(c *gin.Context) {
		allowed, retryAfterSeconds := limiter.Allow(c.ClientIP())
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(retryAfterSeconds))
			response.Error(c, http.StatusTooManyRequests, "rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitFromEnvironment builds the global IP limiter from
// RATE_LIMIT_REQUESTS_PER_MINUTE. RATE_LIMIT_REQUESTS and
// RATE_LIMIT_PER_MINUTE are accepted as aliases. Invalid values use the safe
// default.
func RateLimitFromEnvironment() gin.HandlerFunc {
	return RateLimit(NewIPRateLimiter(RateLimitPerMinuteFromEnvironment()))
}

// RateLimitPerMinuteFromEnvironment returns the configured limit or the
// documented default when the setting is absent or invalid.
func RateLimitPerMinuteFromEnvironment() int {
	value := strings.TrimSpace(os.Getenv("RATE_LIMIT_REQUESTS_PER_MINUTE"))
	if value == "" {
		value = strings.TrimSpace(os.Getenv("RATE_LIMIT_REQUESTS"))
	}
	if value == "" {
		value = strings.TrimSpace(os.Getenv("RATE_LIMIT_PER_MINUTE"))
	}

	requestsPerMinute, err := strconv.Atoi(value)
	if err != nil || requestsPerMinute <= 0 {
		return DefaultRequestsPerMinute
	}

	return requestsPerMinute
}
