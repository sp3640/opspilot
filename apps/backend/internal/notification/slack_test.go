package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSlackProviderSendsFormattedText(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider := NewSlackProvider()
	msg := Message{
		Title:  "Critical alert: CPU spike",
		Body:   "CPU usage exceeded threshold",
		Fields: map[string]string{"Project": "proj-1", "Source": "KUBERNETES"},
	}

	if err := provider.Send(context.Background(), server.URL, msg); err != nil {
		t.Fatalf("send: %v", err)
	}

	text, _ := received["text"].(string)
	if !strings.Contains(text, "Critical alert: CPU spike") {
		t.Fatalf("expected text to contain title, got %q", text)
	}
	if !strings.Contains(text, "Project: proj-1") || !strings.Contains(text, "Source: KUBERNETES") {
		t.Fatalf("expected text to contain sorted fields, got %q", text)
	}
}

func TestSlackProviderRequiresTarget(t *testing.T) {
	provider := NewSlackProvider()
	if err := provider.Send(context.Background(), "", Message{Title: "x"}); err == nil {
		t.Fatalf("expected error for empty target")
	}
}

func TestSlackProviderFailsOnNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := NewSlackProvider()
	if err := provider.Send(context.Background(), server.URL, Message{Title: "x"}); err == nil {
		t.Fatalf("expected error for 500 response")
	}
}
