package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildAuthorizeURLIncludesState(t *testing.T) {
	cfg := NewOAuthConfig("client-id", "client-secret", "https://ops.example.com/callback")
	authorizeURL := cfg.BuildAuthorizeURL("signed-state-token")

	if !strings.Contains(authorizeURL, "client_id=client-id") {
		t.Fatalf("expected client_id in authorize URL, got %s", authorizeURL)
	}
	if !strings.Contains(authorizeURL, "state=signed-state-token") {
		t.Fatalf("expected state in authorize URL, got %s", authorizeURL)
	}
	if strings.Contains(authorizeURL, "client_secret") {
		t.Fatalf("client secret must never appear in the authorize URL, got %s", authorizeURL)
	}
}

func TestIsConfiguredRequiresBothIDAndSecret(t *testing.T) {
	if (&OAuthConfig{}).IsConfigured() {
		t.Fatalf("expected an empty OAuthConfig to be unconfigured")
	}
	if (&OAuthConfig{ClientID: "x"}).IsConfigured() {
		t.Fatalf("expected a config missing a client secret to be unconfigured")
	}
	if !(&OAuthConfig{ClientID: "x", ClientSecret: "y"}).IsConfigured() {
		t.Fatalf("expected a config with both fields set to be configured")
	}
}

func TestExchangeCodeSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.FormValue("code") != "one-time-code" {
			t.Fatalf("expected the code to be forwarded, got %q", r.FormValue("code"))
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "SUPER_SECRET_GITHUB_TOKEN", "token_type": "bearer"})
	}))
	defer server.Close()

	cfg := NewOAuthConfig("client-id", "client-secret", "https://ops.example.com/callback")
	cfg.AccessTokenURL = server.URL

	token, err := cfg.ExchangeCode(context.Background(), "one-time-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "SUPER_SECRET_GITHUB_TOKEN" {
		t.Fatalf("expected the exchanged token to round-trip, got %q", token)
	}
}

func TestExchangeCodeRejectsGitHubError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "bad_verification_code", "error_description": "expired"})
	}))
	defer server.Close()

	cfg := NewOAuthConfig("client-id", "client-secret", "https://ops.example.com/callback")
	cfg.AccessTokenURL = server.URL

	if _, err := cfg.ExchangeCode(context.Background(), "stale-code"); err != ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestExchangeCodeFailsWhenUnconfigured(t *testing.T) {
	cfg := &OAuthConfig{}
	if _, err := cfg.ExchangeCode(context.Background(), "any-code"); err == nil {
		t.Fatalf("expected an error when OAuth is not configured")
	}
}
