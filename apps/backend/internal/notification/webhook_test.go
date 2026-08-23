package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhookProviderSendsGenericEnvelope(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected JSON content type, got %q", r.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider := NewWebhookProvider()
	msg := Message{Event: "DEPLOYMENT_FAILED", Title: "Deployment failed", Body: "image pull error", Severity: "HIGH"}

	if err := provider.Send(context.Background(), server.URL, msg); err != nil {
		t.Fatalf("send: %v", err)
	}

	if received["event"] != "DEPLOYMENT_FAILED" {
		t.Fatalf("expected event field to round-trip, got %v", received["event"])
	}
	if received["title"] != "Deployment failed" {
		t.Fatalf("expected title field to round-trip, got %v", received["title"])
	}
	if _, ok := received["timestamp"].(string); !ok {
		t.Fatalf("expected a timestamp string field, got %v", received["timestamp"])
	}
}

func TestWebhookProviderRejectsNonHTTPTarget(t *testing.T) {
	provider := NewWebhookProvider()
	if err := provider.Send(context.Background(), "ftp://example.com/hook", Message{Title: "x"}); err == nil {
		t.Fatalf("expected error for non-http(s) target")
	}
}
