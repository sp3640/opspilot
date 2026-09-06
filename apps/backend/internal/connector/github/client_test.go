package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthenticatedUserSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer good-token" {
			t.Fatalf("expected bearer token header, got %q", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/user" {
			t.Fatalf("expected /user, got %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(User{Login: "octocat", Name: "The Octocat"})
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	user, err := client.AuthenticatedUser(context.Background(), "good-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Login != "octocat" {
		t.Fatalf("expected login octocat, got %q", user.Login)
	}
}

func TestAuthenticatedUserRequiresToken(t *testing.T) {
	client := NewClient("http://example.invalid", nil)
	if _, err := client.AuthenticatedUser(context.Background(), ""); err != ErrTokenRequired {
		t.Fatalf("expected ErrTokenRequired, got %v", err)
	}
}

func TestAuthenticatedUserUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	if _, err := client.AuthenticatedUser(context.Background(), "bad-token"); err != ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestListRepositoriesRateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	if _, err := client.ListRepositories(context.Background(), "token", 1, 30); err != ErrRateLimited {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

func TestListRepositoriesPlainForbiddenIsNotRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No X-RateLimit-Remaining header - a permissions problem, not a rate limit.
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	if _, err := client.ListRepositories(context.Background(), "token", 1, 30); err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestGetCommitNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	_, err := client.GetCommit(context.Background(), "token", "acme", "widgets", "deadbeef")
	if !IsNotFound(err) {
		t.Fatalf("expected IsNotFound(err) to be true, got %v", err)
	}
}

func TestClientTimesOutRatherThanBlockingForever(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.AuthenticatedUser(ctx, "token")
	if err != ErrTimeout {
		t.Fatalf("expected ErrTimeout when the context deadline is exceeded, got %v", err)
	}
}

func TestListCommitsBoundsPerPage(t *testing.T) {
	var capturedPerPage string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPerPage = r.URL.Query().Get("per_page")
		_ = json.NewEncoder(w).Encode([]Commit{})
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	if _, err := client.ListCommits(context.Background(), "token", "acme", "widgets", "main", 9999); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedPerPage != "100" {
		t.Fatalf("expected per_page to be clamped to 100, got %q", capturedPerPage)
	}
}

func TestUnavailableOnServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	if _, err := client.AuthenticatedUser(context.Background(), "token"); err != ErrUnavailable {
		t.Fatalf("expected ErrUnavailable for a 500, got %v", err)
	}
}
