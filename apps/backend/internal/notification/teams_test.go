package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTeamsProviderSendsMessageCard(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider := NewTeamsProvider()
	msg := Message{
		Title:    "SEV-1 incident: Checkout down",
		Body:     "Checkout service is returning 500s",
		Severity: "P0",
		Fields:   map[string]string{"Project": "proj-1"},
	}

	if err := provider.Send(context.Background(), server.URL, msg); err != nil {
		t.Fatalf("send: %v", err)
	}

	if received["@type"] != "MessageCard" {
		t.Fatalf("expected @type MessageCard, got %v", received["@type"])
	}
	if received["title"] != msg.Title {
		t.Fatalf("expected title %q, got %v", msg.Title, received["title"])
	}
	if received["themeColor"] != "D93F3F" {
		t.Fatalf("expected critical theme color for P0, got %v", received["themeColor"])
	}
}

func TestTeamsProviderRequiresTarget(t *testing.T) {
	provider := NewTeamsProvider()
	if err := provider.Send(context.Background(), "", Message{Title: "x"}); err == nil {
		t.Fatalf("expected error for empty target")
	}
}
