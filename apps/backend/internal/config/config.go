package config

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName         string
	AppEnv          string
	Port            string
	Version         string
	DBURL           string
	RateLimit       int
	RateLimitWindow time.Duration

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

	cfg := &Config{
		AppName:         getEnv("APP_NAME", "OpsPilot Backend"),
		AppEnv:          getEnv("APP_ENV", "development"),
		Port:            getEnv("PORT", "8080"),
		Version:         getEnv("APP_VERSION", "dev"),
		DBURL:           strings.TrimSpace(getEnv("DATABASE_URL", "")),
		RateLimit:       rateLimit,
		RateLimitWindow: rateLimitWindow,

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

		GitHubClientID:         strings.TrimSpace(getEnv("GITHUB_CLIENT_ID", "")),
		GitHubClientSecret:     strings.TrimSpace(getEnv("GITHUB_CLIENT_SECRET", "")),
		GitHubOAuthRedirectURL: strings.TrimSpace(getEnv("GITHUB_OAUTH_REDIRECT_URL", "")),
		FrontendBaseURL:        strings.TrimSpace(getEnv("FRONTEND_BASE_URL", "http://localhost:3000")),
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
