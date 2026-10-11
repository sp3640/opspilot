package config

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	validConfig := Config{
		AppEnv:                         "production",
		Port:                           "8080",
		DBURL:                          "postgres://opspilot:password@db:5432/opspilot?sslmode=require",
		JWTSecret:                      strings.Repeat("a", 32),
		ClusterCredentialEncryptionKey: base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))),
		CORSAllowedOrigins:             []string{"https://app.example.com"},
		FrontendBaseURL:                "https://app.example.com",
		SMTPHost:                       "smtp.mail.example.com",
		SMTPFrom:                       "OpsPilot <no-reply@example.com>",
	}

	tests := []struct {
		name string
		edit func(*Config)
	}{
		{
			name: "rejects short JWT secret",
			edit: func(cfg *Config) { cfg.JWTSecret = "too-short" },
		},
		{
			name: "rejects invalid environment",
			edit: func(cfg *Config) { cfg.AppEnv = "qa" },
		},
		{
			name: "rejects invalid port",
			edit: func(cfg *Config) { cfg.Port = "70000" },
		},
		{
			name: "rejects non PostgreSQL URL",
			edit: func(cfg *Config) { cfg.DBURL = "mysql://db/opspilot" },
		},
		{
			name: "rejects missing cluster encryption key",
			edit: func(cfg *Config) { cfg.ClusterCredentialEncryptionKey = "" },
		},
		{
			name: "rejects invalid cluster encryption key",
			edit: func(cfg *Config) { cfg.ClusterCredentialEncryptionKey = "not-base64" },
		},
		{
			name: "rejects production without CORS_ALLOWED_ORIGINS",
			edit: func(cfg *Config) { cfg.CORSAllowedOrigins = nil },
		},
		{
			name: "rejects wildcard CORS origin in production",
			edit: func(cfg *Config) { cfg.CORSAllowedOrigins = []string{"*"} },
		},
		{
			name: "rejects a wildcard mixed in with real production origins",
			edit: func(cfg *Config) {
				cfg.CORSAllowedOrigins = []string{"https://app.example.com", "*"}
			},
		},
		{
			name: "rejects production FRONTEND_BASE_URL left at the localhost default",
			edit: func(cfg *Config) { cfg.FrontendBaseURL = defaultLocalOrigin },
		},
		{
			name: "rejects an SMTP_FROM that is not an email address",
			edit: func(cfg *Config) { cfg.SMTPFrom = "OpsPilot no-reply" },
		},
		{
			name: "rejects a non-HTTPS FRONTEND_BASE_URL in production (verification links carry a token)",
			edit: func(cfg *Config) { cfg.FrontendBaseURL = "http://app.example.com" },
		},
		{
			name: "rejects production without SMTP_HOST (nobody could verify an account)",
			edit: func(cfg *Config) { cfg.SMTPHost = "" },
		},
		{
			name: "rejects production without SMTP_FROM",
			edit: func(cfg *Config) { cfg.SMTPFrom = "  " },
		},
		{
			name: "rejects a negative email verification token TTL",
			edit: func(cfg *Config) { cfg.EmailVerificationTokenTTL = -time.Hour },
		},
		{
			name: "rejects a negative HTTP write timeout",
			edit: func(cfg *Config) { cfg.HTTPWriteTimeout = -time.Second },
		},
		{
			name: "rejects an HTTP write timeout equal to the deployment execution timeout",
			edit: func(cfg *Config) { cfg.HTTPWriteTimeout = DeploymentExecutionTimeout },
		},
		{
			name: "rejects an HTTP write timeout shorter than the deployment execution timeout",
			edit: func(cfg *Config) { cfg.HTTPWriteTimeout = DeploymentExecutionTimeout - time.Second },
		},
	}

	if err := validConfig.Validate(); err != nil {
		t.Fatalf("valid configuration was rejected: %v", err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := validConfig
			test.edit(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

// TestConfigValidateAllowsLocalhostOutsideProduction covers the "local
// development still works" half of CORS/FrontendBaseURL production
// validation: development (and test/staging) must accept the localhost
// defaults Load() falls back to when CORS_ALLOWED_ORIGINS/FRONTEND_BASE_URL
// are left unset, since validateProductionOrigins only applies to
// APP_ENV=production.
func TestConfigValidateAllowsLocalhostOutsideProduction(t *testing.T) {
	for _, appEnv := range []string{"development", "test", "staging"} {
		t.Run(appEnv, func(t *testing.T) {
			cfg := Config{
				AppEnv:                         appEnv,
				Port:                           "8080",
				DBHost:                         "localhost",
				DBUser:                         "opspilot",
				DBName:                         "opspilot",
				DBPort:                         "5432",
				JWTSecret:                      strings.Repeat("a", 32),
				ClusterCredentialEncryptionKey: base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))),
				CORSAllowedOrigins:             []string{defaultLocalOrigin},
				FrontendBaseURL:                defaultLocalOrigin,
			}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("expected the localhost default to remain valid outside production, got: %v", err)
			}
		})
	}
}

// TestConfigValidateProductionAcceptsMultipleExplicitOrigins covers
// "multiple configured origins work" for a real production configuration.
func TestConfigValidateProductionAcceptsMultipleExplicitOrigins(t *testing.T) {
	cfg := Config{
		AppEnv:                         "production",
		Port:                           "8080",
		DBURL:                          "postgres://opspilot:password@db:5432/opspilot?sslmode=require",
		JWTSecret:                      strings.Repeat("a", 32),
		ClusterCredentialEncryptionKey: base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))),
		CORSAllowedOrigins:             []string{"https://app.example.com", "https://admin.example.com"},
		FrontendBaseURL:                "https://app.example.com",
		SMTPHost:                       "smtp.mail.example.com",
		SMTPFrom:                       "OpsPilot <no-reply@example.com>",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected multiple explicit production origins to be valid, got: %v", err)
	}
}

func TestParseCORSOrigins(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{
			name: "single origin",
			raw:  "http://localhost:3000",
			want: []string{"http://localhost:3000"},
		},
		{
			name: "multiple comma-separated origins",
			raw:  "https://app.example.com,https://admin.example.com",
			want: []string{"https://app.example.com", "https://admin.example.com"},
		},
		{
			name: "whitespace around origins and separators is trimmed",
			raw:  "  https://app.example.com , https://admin.example.com  ",
			want: []string{"https://app.example.com", "https://admin.example.com"},
		},
		{
			name: "empty entries from stray/trailing commas are dropped",
			raw:  "https://app.example.com,,https://admin.example.com,",
			want: []string{"https://app.example.com", "https://admin.example.com"},
		},
		{
			name: "empty string yields no origins",
			raw:  "",
			want: nil,
		},
		{
			name: "whitespace-only string yields no origins",
			raw:  "   ",
			want: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := parseCORSOrigins(test.raw)
			if len(got) != len(test.want) {
				t.Fatalf("parseCORSOrigins(%q) = %#v, want %#v", test.raw, got, test.want)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Fatalf("parseCORSOrigins(%q) = %#v, want %#v", test.raw, got, test.want)
				}
			}
		})
	}
}

func TestDatabaseDSN(t *testing.T) {
	cfg := Config{
		DBURL: "postgres://opspilot:password@db/opspilot",
	}
	if got := cfg.DatabaseDSN(); got != cfg.DBURL {
		t.Fatalf("DatabaseDSN() = %q, want preferred DBURL %q", got, cfg.DBURL)
	}

	cfg = Config{
		DBHost:     "db",
		DBPort:     "5432",
		DBUser:     "opspilot",
		DBPassword: "password",
		DBName:     "opspilot",
		DBSSLMode:  "require",
	}
	want := "host=db user=opspilot password=password dbname=opspilot port=5432 sslmode=require"
	if got := cfg.DatabaseDSN(); got != want {
		t.Fatalf("DatabaseDSN() = %q, want %q", got, want)
	}
}

func TestConfigValidateDefaultsEmailVerificationTokenTTL(t *testing.T) {
	cfg := Config{
		AppEnv:                         "development",
		Port:                           "8080",
		DBHost:                         "localhost",
		DBUser:                         "opspilot",
		DBName:                         "opspilot",
		DBPort:                         "5432",
		JWTSecret:                      strings.Repeat("a", 32),
		ClusterCredentialEncryptionKey: base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))),
		CORSAllowedOrigins:             []string{defaultLocalOrigin},
		FrontendBaseURL:                defaultLocalOrigin,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if cfg.EmailVerificationTokenTTL != 24*time.Hour {
		t.Fatalf("default verification token TTL = %s, want 24h", cfg.EmailVerificationTokenTTL)
	}
}

// TestConfigValidateDefaultsHTTPWriteTimeout covers the "deployment
// creation/rollback must not be cut off" fix: with HTTP_WRITE_TIMEOUT unset,
// the default must stay greater than DeploymentExecutionTimeout, the executor
// timeout the deployment service runs synchronously in the request.
func TestConfigValidateDefaultsHTTPWriteTimeout(t *testing.T) {
	cfg := Config{
		AppEnv:                         "development",
		Port:                           "8080",
		DBHost:                         "localhost",
		DBUser:                         "opspilot",
		DBName:                         "opspilot",
		DBPort:                         "5432",
		JWTSecret:                      strings.Repeat("a", 32),
		ClusterCredentialEncryptionKey: base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))),
		CORSAllowedOrigins:             []string{defaultLocalOrigin},
		FrontendBaseURL:                defaultLocalOrigin,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if cfg.HTTPWriteTimeout != DefaultHTTPWriteTimeout {
		t.Fatalf("default HTTP write timeout = %s, want %s", cfg.HTTPWriteTimeout, DefaultHTTPWriteTimeout)
	}
	if cfg.HTTPWriteTimeout <= DeploymentExecutionTimeout {
		t.Fatalf("default HTTP write timeout (%s) is not greater than the deployment execution timeout (%s)", cfg.HTTPWriteTimeout, DeploymentExecutionTimeout)
	}
}

// TestConfigValidateAcceptsAnExplicitlyLongerHTTPWriteTimeout covers a real
// production override, e.g. a deliberately generous HTTP_WRITE_TIMEOUT.
func TestConfigValidateAcceptsAnExplicitlyLongerHTTPWriteTimeout(t *testing.T) {
	cfg := Config{
		AppEnv:                         "production",
		Port:                           "8080",
		DBURL:                          "postgres://opspilot:password@db:5432/opspilot?sslmode=require",
		JWTSecret:                      strings.Repeat("a", 32),
		ClusterCredentialEncryptionKey: base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))),
		CORSAllowedOrigins:             []string{"https://app.example.com"},
		FrontendBaseURL:                "https://app.example.com",
		SMTPHost:                       "smtp.mail.example.com",
		SMTPFrom:                       "OpsPilot <no-reply@example.com>",
		HTTPWriteTimeout:               5 * time.Minute,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected a longer explicit HTTP write timeout to be valid, got: %v", err)
	}
	if cfg.HTTPWriteTimeout != 5*time.Minute {
		t.Fatalf("HTTPWriteTimeout was altered: got %s, want 5m", cfg.HTTPWriteTimeout)
	}
}

// TestParseTrustedProxies covers TRUSTED_PROXIES parsing: splitting,
// trimming, empty-entry handling (mirroring parseCORSOrigins), and the extra
// validation step CORS origins don't need - every entry must be exactly what
// gin.Engine.SetTrustedProxies accepts, so a config-time typo is caught here
// rather than only failing (or silently doing nothing) later.
func TestParseTrustedProxies(t *testing.T) {
	okTests := []struct {
		name string
		raw  string
		want []string
	}{
		{name: "empty string yields no proxies", raw: "", want: nil},
		{name: "whitespace-only string yields no proxies", raw: "   ", want: nil},
		{name: "single bare IPv4 address", raw: "203.0.113.9", want: []string{"203.0.113.9"}},
		{name: "single IPv4 CIDR", raw: "10.0.0.0/8", want: []string{"10.0.0.0/8"}},
		{name: "single IPv6 address", raw: "::1", want: []string{"::1"}},
		{name: "single IPv6 CIDR", raw: "2001:db8::/32", want: []string{"2001:db8::/32"}},
		{
			name: "the documented Cloud Run 'trust the immediate peer' configuration",
			raw:  "0.0.0.0/0,::/0",
			want: []string{"0.0.0.0/0", "::/0"},
		},
		{
			name: "whitespace around entries and separators is trimmed",
			raw:  "  10.0.0.0/8 , 192.168.1.1  ",
			want: []string{"10.0.0.0/8", "192.168.1.1"},
		},
		{
			name: "empty entries from stray/trailing commas are dropped",
			raw:  "10.0.0.0/8,,192.168.1.1,",
			want: []string{"10.0.0.0/8", "192.168.1.1"},
		},
	}
	for _, test := range okTests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseTrustedProxies(test.raw)
			if err != nil {
				t.Fatalf("parseTrustedProxies(%q) returned an unexpected error: %v", test.raw, err)
			}
			if len(got) != len(test.want) {
				t.Fatalf("parseTrustedProxies(%q) = %#v, want %#v", test.raw, got, test.want)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Fatalf("parseTrustedProxies(%q) = %#v, want %#v", test.raw, got, test.want)
				}
			}
		})
	}

	invalidTests := []struct {
		name string
		raw  string
	}{
		{name: "not an IP address at all", raw: "not-an-ip"},
		{name: "an out-of-range IPv4 octet", raw: "999.0.0.1"},
		{name: "a CIDR with a malformed prefix", raw: "10.0.0.0/abc"},
		{name: "a CIDR with an out-of-range prefix length", raw: "10.0.0.0/99"},
		{name: "one invalid entry among otherwise valid ones", raw: "10.0.0.0/8,not-an-ip"},
		{name: "a hostname is not accepted (gin.SetTrustedProxies needs IPs/CIDRs)", raw: "proxy.internal.example.com"},
	}
	for _, test := range invalidTests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parseTrustedProxies(test.raw); err == nil {
				t.Fatalf("parseTrustedProxies(%q) did not return an error", test.raw)
			}
		})
	}
}

// TestConfigLoadRejectsInvalidTrustedProxies exercises the full Load() path
// (env var -> parseTrustedProxies -> Validate), confirming a malformed
// TRUSTED_PROXIES fails startup with a clear error rather than being
// silently accepted or only failing later inside gin.
func TestConfigLoadRejectsInvalidTrustedProxies(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("JWT_SECRET", strings.Repeat("a", 32))
	t.Setenv("CLUSTER_CREDENTIAL_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))))
	t.Setenv("DATABASE_URL", "postgres://opspilot:password@db:5432/opspilot?sslmode=disable")
	t.Setenv("TRUSTED_PROXIES", "not-an-ip")

	if _, err := Load(); err == nil {
		t.Fatal("expected Load() to reject an invalid TRUSTED_PROXIES entry")
	} else if !strings.Contains(err.Error(), "TRUSTED_PROXIES") {
		t.Fatalf("expected the error to mention TRUSTED_PROXIES, got: %v", err)
	}
}

// TestConfigLoadDefaultsTrustedProxiesToNone confirms local/development
// behavior is unaffected when TRUSTED_PROXIES is unset: ClientIP() should
// keep resolving to the raw TCP peer (see the ClientIP-level coverage in
// internal/middleware for the behavioral proof).
func TestConfigLoadDefaultsTrustedProxiesToNone(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("JWT_SECRET", strings.Repeat("a", 32))
	t.Setenv("CLUSTER_CREDENTIAL_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))))
	t.Setenv("DATABASE_URL", "postgres://opspilot:password@db:5432/opspilot?sslmode=disable")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected Load() error: %v", err)
	}
	if cfg.TrustedProxies != nil {
		t.Fatalf("TrustedProxies = %#v, want nil when TRUSTED_PROXIES is unset", cfg.TrustedProxies)
	}
}
