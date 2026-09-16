package prometheus

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sp3640/opspilot/backend/internal/connector"
)

func TestPrometheusConnectorTestConnectionSucceedsForValidAPI(t *testing.T) {
	var gotMethod string
	var gotURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotURL = r.URL.String()
		if r.URL.Path != "/api/v1/query" {
			http.Error(w, "unexpected path", http.StatusBadRequest)
			return
		}
		if r.URL.Query().Get("query") != "up" {
			http.Error(w, "missing query", http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if len(body) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data": map[string]any{
				"resultType": "vector",
				"result":     []any{},
			},
		})
	}))
	defer server.Close()

	c := NewConnector()
	result, err := c.TestConnection(context.Background(), connector.Config{Metadata: map[string]string{"api_url": server.URL}, Credentials: map[string]string{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success=true, got %+v", result)
	}
	if gotMethod != http.MethodGet {
		t.Fatalf("expected GET, got %q", gotMethod)
	}
	if !strings.Contains(gotURL, "/api/v1/query") {
		t.Fatalf("expected query endpoint, got %q", gotURL)
	}
}

func TestPrometheusConnectorRejectsInvalidConfiguration(t *testing.T) {
	c := NewConnector()
	if _, err := c.TestConnection(context.Background(), connector.Config{}); err == nil {
		t.Fatal("expected empty config to fail")
	}
	if _, err := c.TestConnection(context.Background(), connector.Config{Metadata: map[string]string{"api_url": "not a url"}}); err == nil {
		t.Fatal("expected invalid URL to fail")
	}
}

func TestPrometheusConnectorSupportsBearerTokenAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer super-secret-token" {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": map[string]any{"resultType": "vector", "result": []any{}}})
	}))
	defer server.Close()

	c := NewConnector()
	result, err := c.TestConnection(context.Background(), connector.Config{
		Metadata: map[string]string{"api_url": server.URL},
		Credentials: map[string]string{"bearer_token": "super-secret-token"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success=true, got %+v", result)
	}
}

func TestPrometheusConnectorReportsCapabilities(t *testing.T) {
	caps := NewConnector().Capabilities()
	if !caps.Implemented || !caps.SupportsTestConnection || !caps.SupportsHealthCheck {
		t.Fatalf("expected full connector capabilities, got %+v", caps)
	}
}
