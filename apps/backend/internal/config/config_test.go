package config

import (
	"strings"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	validConfig := Config{
		AppEnv:    "production",
		Port:      "8080",
		DBURL:     "postgres://opspilot:password@db:5432/opspilot?sslmode=require",
		JWTSecret: strings.Repeat("a", 32),
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
