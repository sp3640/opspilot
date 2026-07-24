package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/logger"
)

const (
	// RequestIDContextKey is the Gin context key that holds the current request
	// ID. Use RequestIDFromContext rather than accessing it directly.
	RequestIDContextKey = "request_id"
	// RequestIDHeader is returned on every HTTP response.
	RequestIDHeader = "X-Request-ID"
)

var requestIDFallbackCounter atomic.Uint64

// RequestID assigns a fresh UUID to each request, makes it available to
// handlers, adds it to the request context used by logging, and returns it in
// the response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := NewRequestID()

		c.Set(RequestIDContextKey, requestID)
		c.Header(RequestIDHeader, requestID)
		c.Request = c.Request.WithContext(logger.WithRequestID(c.Request.Context(), requestID))

		c.Next()
	}
}

// RequestIDFromContext returns the request ID set by RequestID middleware.
func RequestIDFromContext(c *gin.Context) string {
	if c == nil {
		return ""
	}

	requestID, _ := c.Get(RequestIDContextKey)
	value, _ := requestID.(string)
	return value
}

// NewRequestID returns an RFC 4122 version 4 UUID. The fallback still
// preserves the UUID shape if the operating system random source is
// unavailable, which lets request handling continue during a degraded host
// condition.
func NewRequestID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		fallback := uint64(time.Now().UnixNano()) ^ requestIDFallbackCounter.Add(1)
		for index := 0; index < 8; index++ {
			value[index] = byte(fallback >> (index * 8))
		}
		for index := 8; index < len(value); index++ {
			value[index] = byte(fallback>>((index-8)*8)) ^ byte(index*31)
		}
	}

	// Set the version and variant bits required by RFC 4122 UUID v4.
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80

	encoded := make([]byte, 36)
	hex.Encode(encoded[0:8], value[0:4])
	encoded[8] = '-'
	hex.Encode(encoded[9:13], value[4:6])
	encoded[13] = '-'
	hex.Encode(encoded[14:18], value[6:8])
	encoded[18] = '-'
	hex.Encode(encoded[19:23], value[8:10])
	encoded[23] = '-'
	hex.Encode(encoded[24:36], value[10:16])

	return string(encoded)
}
