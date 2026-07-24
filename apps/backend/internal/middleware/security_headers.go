package middleware

import "github.com/gin-gonic/gin"

const contentSecurityPolicy = "default-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

// SecurityHeaders adds defensive HTTP response headers for every endpoint.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Content-Security-Policy", contentSecurityPolicy)

		c.Next()
	}
}
