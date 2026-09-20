package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/connector"
)

func TestPrometheusServiceQueryInstantParsesVectorResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("query"); got != "up" {
			t.Fatalf("unexpected query %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer super-secret-token" {
			t.Fatalf("unexpected auth header %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"instance":"localhost:9090"},"value":[1700000000,"1"]}]}}`))
	}))
	defer server.Close()

	svc := NewPrometheusService().WithClient(server.Client())
	result, err := svc.QueryInstant(context.Background(), connector.Config{
		Metadata: map[string]string{"api_url": server.URL},
		Credentials: map[string]string{"bearer_token": "super-secret-token"},
	}, "up")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Samples) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(result.Samples))
	}
	if result.Samples[0].Value != 1 {
		t.Fatalf("expected value 1, got %v", result.Samples[0].Value)
	}
	if result.Query != "up" {
		t.Fatalf("expected query to be preserved, got %q", result.Query)
	}
	_ = uuid.New()
}

func TestPrometheusServiceQueryRangeParsesMatrixResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("query"); got != "up" {
			t.Fatalf("unexpected query %q", got)
		}
		if got := r.URL.Query().Get("start"); got == "" {
			t.Fatalf("missing start param")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"instance":"localhost:9090"},"values":[[1700000000,"1"],[1700000060,"2"]]}]}}`))
	}))
	defer server.Close()

	svc := NewPrometheusService().WithClient(server.Client())
	start := time.Unix(1700000000, 0)
	end := time.Unix(1700000060, 0)
	result, err := svc.QueryRange(context.Background(), connector.Config{Metadata: map[string]string{"api_url": server.URL}}, "up", start, end, "60s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Series) != 1 {
		t.Fatalf("expected 1 series, got %d", len(result.Series))
	}
	if len(result.Series[0].Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(result.Series[0].Values))
	}
	if result.Series[0].Values[1].Value != 2 {
		t.Fatalf("expected second sample value 2, got %v", result.Series[0].Values[1].Value)
	}
}
