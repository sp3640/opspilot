package config

import (
	"encoding/base64"
	"strings"
	"testing"
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
