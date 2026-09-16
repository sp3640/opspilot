package slack

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sp3640/opspilot/backend/internal/connector"
	"github.com/sp3640/opspilot/backend/internal/notification"
)

func TestSlackConnectorTestConnectionSucceedsForValidWebhook(t *testing.T) {
	var gotMethod string
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		gotBody = string(body)
		if err := json.Unmarshal(body, &struct {
			Text string `json:"text"`
		}{}); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewConnector()
	result, err := c.TestConnection(context.Background(), connector.Config{Credentials: map[string]string{"webhook_url": server.URL}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success=true, got %+v", result)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %q", gotMethod)
	}
	if !strings.Contains(gotBody, "OpsPilot Slack connection test succeeded") {
		t.Fatalf("expected payload to include test message, got %q", gotBody)
	}
}

func TestSlackConnectorRejectsInvalidWebhookConfiguration(t *testing.T) {
	c := NewConnector()
	if _, err := c.TestConnection(context.Background(), connector.Config{}); err == nil {
		t.Fatal("expected empty config to fail")
	}
	if _, err := c.TestConnection(context.Background(), connector.Config{Credentials: map[string]string{"webhook_url": "not a url"}}); err == nil {
		t.Fatal("expected invalid URL to fail")
	}
}

func TestSlackConnectorDoesNotLeakWebhookInErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	secret := server.URL + "/services/secret-webhook-value"
	c := NewConnector()
	_, err := c.TestConnection(context.Background(), connector.Config{Credentials: map[string]string{"webhook_url": secret}})
	if err == nil {
		t.Fatal("expected failing webhook response to error")
	}
	if strings.Contains(err.Error(), "secret-webhook-value") {
		t.Fatalf("expected webhook URL to stay out of errors, got %q", err)
	}
}

func TestSlackConnectorSendNotificationSucceeds(t *testing.T) {
	var gotMethod string
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewConnector()
	msg := notification.Message{
		Title:    "HIGH Alert",
		Body:     "Pod crash detected",
		Severity: "HIGH",
		Fields:   map[string]string{"Application": "OpsPilot Backend", "Incident": "INC-123"},
	}
	if err := c.SendNotification(context.Background(), connector.Config{Credentials: map[string]string{"webhook_url": server.URL}}, msg); err != nil {
		t.Fatalf("send notification: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %q", gotMethod)
	}
	text, ok := payload["text"].(string)
	if !ok {
		t.Fatalf("expected text payload, got %v", payload)
	}
	if !strings.Contains(text, "HIGH Alert") || !strings.Contains(text, "Pod crash detected") || !strings.Contains(text, "Application: OpsPilot Backend") {
		t.Fatalf("expected message to include operational context, got %q", text)
	}
}

func TestSlackConnectorReportsCapabilities(t *testing.T) {
	caps := NewConnector().Capabilities()
	if !caps.Implemented || !caps.SupportsTestConnection || !caps.SupportsHealthCheck {
		t.Fatalf("expected full connector capabilities, got %+v", caps)
	}
}
