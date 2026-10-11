package config

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// defaultLocalOrigin is the local frontend origin every default in this file
// falls back to when unset - both CORSAllowedOrigins and FrontendBaseURL, so
// local development keeps working exactly as before this file introduced
// production-specific validation for either one.
const defaultLocalOrigin = "http://localhost:3000"

// DefaultEmailVerificationTokenTTL is the lifetime of a verification link
// when EMAIL_VERIFICATION_TOKEN_TTL is unset.
const DefaultEmailVerificationTokenTTL = 24 * time.Hour

// DeploymentExecutionTimeout bounds how long one synchronous Kubernetes
// deployment rollout may run
// (internal/kubernetes/executor.DeploymentExecutor, wired up in
// cmd/server/main.go). It is not itself environment-configurable - changing
// how long a deployment is allowed to run is an application change, not a
// per-deployment tuning knob - but DefaultHTTPWriteTimeout is derived from
// it below, and Validate rejects an explicitly configured HTTPWriteTimeout
// that isn't longer than it. That keeps the two values from silently
// drifting apart the way they did before this constant existed: a
// synchronous request that legitimately runs for up to this long must never
// be cut off by the HTTP server's own write timeout first.
const DeploymentExecutionTimeout = 90 * time.Second

// httpWriteTimeoutMargin covers the non-executor work sharing the same
// request as a deployment execution (repository writes, audit logging,
// response encoding) on top of DeploymentExecutionTimeout.
const httpWriteTimeoutMargin = 30 * time.Second

// DefaultHTTPWriteTimeout is used when HTTP_WRITE_TIMEOUT is unset. It must
// stay greater than DeploymentExecutionTimeout - see the comment there.
const DefaultHTTPWriteTimeout = DeploymentExecutionTimeout + httpWriteTimeoutMargin

type Config struct {
	AppName         string
	AppEnv          string
	Port            string
	Version         string
	DBURL           string
	RateLimit       int
	RateLimitWindow time.Duration

	// CORSAllowedOrigins is the explicit allow-list of origins the API's
	// CORS middleware accepts, parsed from a comma-separated
	// CORS_ALLOWED_ORIGINS. Defaults to the local frontend origin so local
	// development is unaffected when unset; production must set this
	// explicitly (enforced in Validate) rather than silently keeping the
	// localhost default.
	CORSAllowedOrigins []string

	// TrustedProxies is the explicit set of proxy IPs/CIDR blocks (IPv4 or
	// IPv6) whose X-Forwarded-For/X-Real-IP headers Gin's c.ClientIP() may
	// trust (gin.Engine.SetTrustedProxies) - used for per-IP rate limiting
	// (general, registration, resend-verification) and for the IP address
	// recorded on login audit events.
	//
	// Defaults to none: with TRUSTED_PROXIES unset, ClientIP() falls back to
	// the raw TCP peer address, exactly as it did before this option
	// existed. That is the safe choice whenever this process can be reached
	// directly (a raw port exposed to the internet), where a caller could
	// otherwise forge X-Forwarded-For to spoof their IP.
	//
	// Behind a reverse proxy or load balancer that is the ONLY way to reach
	// this process - Google Cloud Run's request-routing layer, for example
	// - the immediate TCP peer is always that proxy, and trusting its
	// X-Forwarded-For is what recovers the real client IP. Cloud Run does
	// not publish a fixed, documented CIDR range for that layer, so there is
	// no single narrow CIDR to hardcode here as "the Cloud Run range" -
	// operators who have confirmed their own deployment topology set this
	// explicitly instead, e.g.:
	//
	//	TRUSTED_PROXIES=0.0.0.0/0,::/0
	//
	// (trust the immediate peer unconditionally - appropriate ONLY when
	// nothing but that proxy can ever reach this process directly, which is
	// the case for a Cloud Run service, but is an explicit, operator-made
	// decision rather than an application default).
	TrustedProxies []string

	// HTTPWriteTimeout bounds how long the HTTP server will spend writing a
	// single response (net/http.Server.WriteTimeout). It must stay longer
	// than DeploymentExecutionTimeout - see its comment - since deployment
	// creation/rollback run the executor synchronously in the request
	// goroutine. Configurable via HTTP_WRITE_TIMEOUT; defaults to
	// DefaultHTTPWriteTimeout when unset. This is the only server timeout
	// this option affects: ReadHeaderTimeout, ReadTimeout, and IdleTimeout
	// are unrelated to how long a handler takes to produce its response and
	// are left as fixed values in cmd/server/main.go.
	HTTPWriteTimeout time.Duration

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret string
	JWTExpiry string

	ClusterCredentialEncryptionKey string

	// SMTP settings for the Email notification provider. All optional: an
	// unset SMTPHost simply means email notifications are not configured
	// (the provider reports this plainly rather than failing startup) -
	// Slack/Teams/Webhook notifications never depend on these. Never
	// hardcoded: sourced only from environment configuration.
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// EmailVerificationTokenTTL is how long an emailed verification link
	// stays valid (EMAIL_VERIFICATION_TOKEN_TTL, default 24h).
	EmailVerificationTokenTTL time.Duration

	// GitHub OAuth App credentials (Sprint 28). All optional: an unset
	// GitHubClientID/Secret simply means "connect with GitHub" is
	// unavailable (the OAuth-start endpoint reports this plainly rather
	// than failing startup) - organizations can still connect GitHub by
	// pasting a personal access token through the generic Sprint 27
	// integration-create endpoint. Never hardcoded: sourced only from
	// environment configuration, and the resulting access token is
	// encrypted with the same ClusterCredentialEncryptionKey used by every
	// other integration - no second encryption key is introduced.
	GitHubClientID         string
	GitHubClientSecret     string
	GitHubOAuthRedirectURL string
	// FrontendBaseURL is where the GitHub OAuth callback redirects the
	// browser back to after completing the exchange - the same origin
	// already allow-listed in CORS.
	FrontendBaseURL string
}

// parseCORSOrigins splits a comma-separated origin list, trimming whitespace
// and dropping empty entries - e.g. "https://a.example, https://b.example"
// or a trailing/stray comma both parse cleanly.
func parseCORSOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		origins = append(origins, origin)
	}
	return origins
}

// parseTrustedProxies splits TRUSTED_PROXIES the same way CORS_ALLOWED_ORIGINS
// is split (comma-separated, trimmed, empty entries dropped), then validates
// each entry is exactly what gin.Engine.SetTrustedProxies accepts (a bare
// IPv4/IPv6 address, or an IPv4/IPv6 CIDR block). Validating here, at
// config-load time, means a typo produces a clear "TRUSTED_PROXIES entry ...
// is invalid" error immediately at startup rather than being silently
// accepted here and only failing (or not) later when SetTrustedProxies is
// actually called in cmd/server/main.go.
func parseTrustedProxies(raw string) ([]string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	parts := strings.Split(trimmed, ",")
	proxies := make([]string, 0, len(parts))
	for _, part := range parts {
		proxy := strings.TrimSpace(part)
		if proxy == "" {
			continue
		}
		if err := validateTrustedProxyEntry(proxy); err != nil {
			return nil, fmt.Errorf("TRUSTED_PROXIES entry %q is invalid: %w", proxy, err)
		}
		proxies = append(proxies, proxy)
	}
	if len(proxies) == 0 {
		return nil, nil
	}
	return proxies, nil
}

// validateTrustedProxyEntry mirrors gin's own prepareTrustedCIDRs parsing
// exactly (a bare IP is treated as a /32 or /128, otherwise the entry must
// be a valid CIDR block), so a value this function accepts can never later
// be rejected by SetTrustedProxies, and vice versa.
func validateTrustedProxyEntry(proxy string) error {
	if !strings.Contains(proxy, "/") {
		if net.ParseIP(proxy) == nil {
			return fmt.Errorf("not a valid IP address or CIDR block")
		}
		return nil
	}
	if _, _, err := net.ParseCIDR(proxy); err != nil {
		return fmt.Errorf("not a valid CIDR block: %w", err)
	}
	return nil
}

// Load reads environment configuration and validates all startup-critical values.
// DATABASE_URL is preferred, while the legacy DB_* variables remain supported for
// backwards compatibility.
func Load() (*Config, error) {
	// A missing .env file is normal in containers and production. Environment
	// variables that are already set always take precedence over values in .env.
	_ = godotenv.Load()

	rateLimit, err := getPositiveIntEnv([]string{
		"RATE_LIMIT_REQUESTS_PER_MINUTE",
		"RATE_LIMIT_REQUESTS",
		"RATE_LIMIT_PER_MINUTE",
	}, 100)
	if err != nil {
		return nil, err
	}
	rateLimitWindow, err := getDurationEnv("RATE_LIMIT_WINDOW", time.Minute)
	if err != nil {
		return nil, err
	}

	emailVerificationTTL, err := getDurationEnv("EMAIL_VERIFICATION_TOKEN_TTL", DefaultEmailVerificationTokenTTL)
	if err != nil {
		return nil, err
	}

	trustedProxies, err := parseTrustedProxies(getEnv("TRUSTED_PROXIES", ""))
	if err != nil {
		return nil, err
	}

	httpWriteTimeout, err := getDurationEnv("HTTP_WRITE_TIMEOUT", DefaultHTTPWriteTimeout)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		AppName:         getEnv("APP_NAME", "OpsPilot Backend"),
		AppEnv:          getEnv("APP_ENV", "development"),
		Port:            getEnv("PORT", "8080"),
		Version:         getEnv("APP_VERSION", "dev"),
		DBURL:           strings.TrimSpace(getEnv("DATABASE_URL", "")),
		RateLimit:       rateLimit,
		RateLimitWindow: rateLimitWindow,

		CORSAllowedOrigins: parseCORSOrigins(getEnv("CORS_ALLOWED_ORIGINS", defaultLocalOrigin)),
		TrustedProxies:     trustedProxies,
		HTTPWriteTimeout:   httpWriteTimeout,

		DBHost:     strings.TrimSpace(getEnv("DB_HOST", "")),
		DBPort:     strings.TrimSpace(getEnv("DB_PORT", "5432")),
		DBUser:     strings.TrimSpace(getEnv("DB_USER", "")),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     strings.TrimSpace(getEnv("DB_NAME", "")),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:                      strings.TrimSpace(getEnv("JWT_SECRET", "")),
		JWTExpiry:                      getEnv("JWT_EXPIRY", "24h"),
		ClusterCredentialEncryptionKey: strings.TrimSpace(getEnv("CLUSTER_CREDENTIAL_ENCRYPTION_KEY", "")),

		SMTPHost:     strings.TrimSpace(getEnv("SMTP_HOST", "")),
		SMTPPort:     strings.TrimSpace(getEnv("SMTP_PORT", "587")),
		SMTPUsername: strings.TrimSpace(getEnv("SMTP_USERNAME", "")),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     strings.TrimSpace(getEnv("SMTP_FROM", "")),

		EmailVerificationTokenTTL: emailVerificationTTL,

		GitHubClientID:         strings.TrimSpace(getEnv("GITHUB_CLIENT_ID", "")),
		GitHubClientSecret:     strings.TrimSpace(getEnv("GITHUB_CLIENT_SECRET", "")),
		GitHubOAuthRedirectURL: strings.TrimSpace(getEnv("GITHUB_OAUTH_REDIRECT_URL", "")),
		FrontendBaseURL:        strings.TrimSpace(getEnv("FRONTEND_BASE_URL", defaultLocalOrigin)),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate fails fast for configuration that would otherwise cause an unsafe or
// unavailable service at runtime.
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("configuration is required")
	}

	c.AppEnv = strings.ToLower(strings.TrimSpace(c.AppEnv))
	switch c.AppEnv {
	case "development", "test", "staging", "production":
	default:
		return fmt.Errorf("APP_ENV must be one of development, test, staging, or production")
	}

	if _, err := validatePort(c.Port, "PORT"); err != nil {
		return err
	}
	if c.RateLimit == 0 {
		c.RateLimit = 100
	}
	if c.RateLimit < 1 {
		return fmt.Errorf("RATE_LIMIT_REQUESTS_PER_MINUTE must be a positive integer")
	}
	if c.RateLimitWindow == 0 {
		c.RateLimitWindow = time.Minute
	}
	if c.RateLimitWindow < 0 {
		return fmt.Errorf("RATE_LIMIT_WINDOW must be a positive duration")
	}

	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be set and contain at least 32 characters")
	}
	if err := validateClusterCredentialEncryptionKey(c.ClusterCredentialEncryptionKey); err != nil {
		return err
	}

	if c.EmailVerificationTokenTTL == 0 {
		c.EmailVerificationTokenTTL = DefaultEmailVerificationTokenTTL
	}
	if c.EmailVerificationTokenTTL < 0 {
		return fmt.Errorf("EMAIL_VERIFICATION_TOKEN_TTL must be a positive duration")
	}

	if c.HTTPWriteTimeout == 0 {
		c.HTTPWriteTimeout = DefaultHTTPWriteTimeout
	}
	if c.HTTPWriteTimeout < 0 {
		return fmt.Errorf("HTTP_WRITE_TIMEOUT must be a positive duration")
	}
	if c.HTTPWriteTimeout <= DeploymentExecutionTimeout {
		return fmt.Errorf(
			"HTTP_WRITE_TIMEOUT (%s) must be greater than the deployment execution timeout (%s), or a legitimately long-running deployment request would be cut off by the HTTP server before it can finish",
			c.HTTPWriteTimeout, DeploymentExecutionTimeout,
		)
	}

	if err := validateProductionEmail(c.AppEnv, c.SMTPHost, c.SMTPFrom, c.FrontendBaseURL); err != nil {
		return err
	}

	if err := validateProductionOrigins(c.AppEnv, c.CORSAllowedOrigins, c.FrontendBaseURL); err != nil {
		return err
	}

	if c.DBURL != "" {
		if err := validateDatabaseURL(c.DBURL); err != nil {
			return err
		}
		return nil
	}

	if c.DBHost == "" || c.DBUser == "" || c.DBName == "" {
		return fmt.Errorf("DATABASE_URL or DB_HOST, DB_USER, and DB_NAME must be configured")
	}
	if _, err := validatePort(c.DBPort, "DB_PORT"); err != nil {
		return err
	}
	return nil
}

// validateProductionOrigins enforces that production deployments configure
// their real origins explicitly rather than silently keeping the localhost
// defaults meant for local development. Development/test/staging are left
// alone - localhost remains a valid value there, and nothing here fails
// startup for those environments.
func validateProductionOrigins(appEnv string, corsAllowedOrigins []string, frontendBaseURL string) error {
	if appEnv != "production" {
		return nil
	}

	if len(corsAllowedOrigins) == 0 {
		return fmt.Errorf("CORS_ALLOWED_ORIGINS must be set to at least one explicit origin when APP_ENV=production")
	}
	for _, origin := range corsAllowedOrigins {
		if strings.Contains(origin, "*") {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS must not contain a wildcard (%q) when APP_ENV=production", origin)
		}
	}

	if frontendBaseURL == defaultLocalOrigin {
		return fmt.Errorf("FRONTEND_BASE_URL must be set to the real frontend URL when APP_ENV=production")
	}

	return nil
}

// validateProductionEmail refuses to start a production deployment that
// cannot send email: account registration requires a verification email, so
// without SMTP nobody could ever activate an account. Other environments
// may run without SMTP (registration then reports email as unavailable).
//
// Verification links carry a secret token, so in production they must be
// HTTPS links: FRONTEND_BASE_URL must use https://.
func validateProductionEmail(appEnv, smtpHost, smtpFrom, frontendBaseURL string) error {
	if appEnv != "production" {
		return nil
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(frontendBaseURL)), "https://") {
		return fmt.Errorf("FRONTEND_BASE_URL must be an https:// URL when APP_ENV=production: verification emails link to it")
	}
	if strings.TrimSpace(smtpHost) == "" {
		return fmt.Errorf("SMTP_HOST must be set when APP_ENV=production: account registration requires email verification")
	}
	if strings.TrimSpace(smtpFrom) == "" {
		return fmt.Errorf("SMTP_FROM must be set when APP_ENV=production: account registration requires email verification")
	}
	if _, err := mail.ParseAddress(smtpFrom); err != nil {
		return fmt.Errorf("SMTP_FROM must be an email address such as no-reply@example.com or OpsPilot <no-reply@example.com>: %w", err)
	}
	return nil
}

func validateClusterCredentialEncryptionKey(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("CLUSTER_CREDENTIAL_ENCRYPTION_KEY must be configured")
	}

	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return fmt.Errorf("CLUSTER_CREDENTIAL_ENCRYPTION_KEY must be valid base64")
	}
	if len(decoded) != 32 {
		return fmt.Errorf("CLUSTER_CREDENTIAL_ENCRYPTION_KEY must decode to 32 bytes")
	}

	return nil
}

func validatePort(value, name string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("%s must be a valid port between 1 and 65535", name)
	}
	return port, nil
}

func validateDatabaseURL(value string) error {
	databaseURL, err := url.ParseRequestURI(value)
	if err != nil || databaseURL.Host == "" {
		return fmt.Errorf("DATABASE_URL must be a valid PostgreSQL connection URL")
	}
	if databaseURL.Scheme != "postgres" && databaseURL.Scheme != "postgresql" {
		return fmt.Errorf("DATABASE_URL must use the postgres or postgresql scheme")
	}
	if port := databaseURL.Port(); port != "" {
		if _, err := validatePort(port, "DATABASE_URL port"); err != nil {
			return err
		}
	}
	return nil
}

func getPositiveIntEnv(keys []string, fallback int) (int, error) {
	value := ""
	key := keys[0]
	for _, candidate := range keys {
		if configured := strings.TrimSpace(getEnv(candidate, "")); configured != "" {
			value = configured
			key = candidate
			break
		}
	}
	if value == "" {
		value = strconv.Itoa(fallback)
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return parsed, nil
}

func getDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(getEnv(key, fallback.String()))
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}
	return parsed, nil
}

// DatabaseDSN returns the preferred database URL or a PostgreSQL DSN composed
// from the backwards-compatible DB_* configuration fields.
func (c *Config) DatabaseDSN() string {
	if c.DBURL != "" {
		return c.DBURL
	}

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.DBHost,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBPort,
		c.DBSSLMode,
	)
}
