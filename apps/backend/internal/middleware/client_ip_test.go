package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// These tests exercise gin.Engine.SetTrustedProxies + c.ClientIP() directly,
// with no application wiring beyond what cmd/server/main.go itself does
// (gin.New(), then SetTrustedProxies with a config-supplied list). They
// prove the actual behavior behind the config.Config.TrustedProxies /
// TRUSTED_PROXIES option: what every consumer of ClientIP() in this codebase
// (rate limiting - see rate_limit_test.go - and login audit logging) will
// see for a given trusted-proxy configuration.

func clientIPEngine(t *testing.T, trustedProxies []string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	if err := engine.SetTrustedProxies(trustedProxies); err != nil {
		t.Fatalf("SetTrustedProxies(%v): %v", trustedProxies, err)
	}
	return engine
}

func clientIPFrom(t *testing.T, engine *gin.Engine, remoteAddr, xForwardedFor string) string {
	t.Helper()
	var got string
	engine.GET("/", func(c *gin.Context) {
		got = c.ClientIP()
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = remoteAddr
	if xForwardedFor != "" {
		request.Header.Set("X-Forwarded-For", xForwardedFor)
	}
	engine.ServeHTTP(httptest.NewRecorder(), request)
	return got
}

// TestClientIPWithNoTrustedProxiesIgnoresForwardedHeader is the default,
// local/development behavior (TRUSTED_PROXIES unset -> cfg.TrustedProxies is
// nil): ClientIP() must ignore X-Forwarded-For entirely and return the raw
// TCP peer, exactly as it did before TRUSTED_PROXIES existed. This also
// doubles as the security property the audit asked for: a caller directly
// reachable by this process cannot spoof their IP by sending their own
// X-Forwarded-For header.
func TestClientIPWithNoTrustedProxiesIgnoresForwardedHeader(t *testing.T) {
	engine := clientIPEngine(t, nil)

	got := clientIPFrom(t, engine, "203.0.113.9:54321", "8.8.8.8")
	if got != "203.0.113.9" {
		t.Fatalf("ClientIP() = %q, want the raw TCP peer 203.0.113.9 (X-Forwarded-For must be ignored)", got)
	}
}

// TestClientIPWithNarrowTrustedProxyRecoversRealClientIP is the "explicit,
// narrow trust" production shape: only a specific proxy IP is trusted, and
// only that proxy's forwarded value is honored.
func TestClientIPWithNarrowTrustedProxyRecoversRealClientIP(t *testing.T) {
	engine := clientIPEngine(t, []string{"127.0.0.1/32"})

	got := clientIPFrom(t, engine, "127.0.0.1:54321", "8.8.8.8")
	if got != "8.8.8.8" {
		t.Fatalf("ClientIP() = %q, want the forwarded client IP 8.8.8.8 from the trusted proxy", got)
	}
}

// TestClientIPWithNarrowTrustedProxyIgnoresUntrustedPeer proves the "do not
// blindly trust arbitrary X-Forwarded-For headers" requirement holds even
// with TRUSTED_PROXIES configured: a peer outside the trusted CIDR still
// cannot inject a forwarded IP.
func TestClientIPWithNarrowTrustedProxyIgnoresUntrustedPeer(t *testing.T) {
	engine := clientIPEngine(t, []string{"127.0.0.1/32"})

	got := clientIPFrom(t, engine, "198.51.100.7:54321", "8.8.8.8")
	if got != "198.51.100.7" {
		t.Fatalf("ClientIP() = %q, want the untrusted peer's own address 198.51.100.7 (its X-Forwarded-For must be ignored)", got)
	}
}

// TestClientIPWithCloudRunStyleTrustAllRecoversRealClientIP proves the exact
// configuration documented for Cloud Run in .env.example and
// config.Config.TrustedProxies (TRUSTED_PROXIES=0.0.0.0/0,::/0 - trust the
// immediate peer unconditionally) actually recovers the real client IP,
// regardless of what that peer's own address happens to be.
func TestClientIPWithCloudRunStyleTrustAllRecoversRealClientIP(t *testing.T) {
	engine := clientIPEngine(t, []string{"0.0.0.0/0", "::/0"})

	got := clientIPFrom(t, engine, "169.254.8.129:54321", "203.0.113.42")
	if got != "203.0.113.42" {
		t.Fatalf("ClientIP() = %q, want the forwarded client IP 203.0.113.42", got)
	}
}

// TestSetTrustedProxiesRejectsInvalidEntry documents that gin itself also
// refuses a malformed proxy list - config.parseTrustedProxies validates the
// same shapes at config-load time (see internal/config's TestParseTrustedProxies),
// but this is the second, independent guard cmd/server/main.go relies on
// (its SetTrustedProxies error is still checked, not ignored).
func TestSetTrustedProxiesRejectsInvalidEntry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	if err := engine.SetTrustedProxies([]string{"not-an-ip"}); err == nil {
		t.Fatal("expected SetTrustedProxies to reject a malformed proxy entry")
	}
}
