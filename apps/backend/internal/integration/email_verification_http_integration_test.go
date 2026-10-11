package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// HTTP-level behaviour of the verification API on the real router: status
// codes, machine-readable codes, and - just as important - what the
// responses must never contain.

func registerRequest(email string) map[string]any {
	return map[string]any{"name": "Http User", "email": email, "password": "password123"}
}

func TestEmailVerificationAPIRegistrationIsUnverifiedAndLeaksNothing(t *testing.T) {
	t.Parallel()
	app := setupRBACApp(t)
	const email = "http.user@opspilot.dev"

	rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/register", "", registerRequest(email))
	assertStatus(t, rec, http.StatusCreated)
	raw := mailboxFor(t, app.router).latestToken(t, email)
	if strings.Contains(rec.Body.String(), raw) || strings.Contains(strings.ToLower(rec.Body.String()), "token") {
		t.Fatalf("the register response exposes verification material: %s", rec.Body.String())
	}

	// Registering an address again - whether or not it is verified - must
	// answer exactly like the first registration.
	again := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/register", "", registerRequest(email))
	assertStatus(t, again, http.StatusCreated)
	if again.Body.String() != rec.Body.String() {
		t.Fatalf("registering an existing address is distinguishable:\nfirst:  %s\nsecond: %s", rec.Body.String(), again.Body.String())
	}
	verifyEmailViaHTTP(t, app.router, email)
	verifiedAgain := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/register", "", registerRequest(email))
	assertStatus(t, verifiedAgain, http.StatusCreated)
	if verifiedAgain.Body.String() != rec.Body.String() {
		t.Fatalf("registering a verified address is distinguishable: %s", verifiedAgain.Body.String())
	}
}

func TestEmailVerificationAPILoginRequiresVerification(t *testing.T) {
	t.Parallel()
	app := setupRBACApp(t)
	const email = "login.user@opspilot.dev"

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/register", "", registerRequest(email)), http.StatusCreated)

	rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/login", "", map[string]any{"email": email, "password": "password123"})
	assertStatus(t, rec, http.StatusForbidden)
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["code"] != "email_not_verified" || body["success"] != false {
		t.Fatalf("expected machine-readable email_not_verified, got %v", body)
	}
	if _, has := body["data"]; has || strings.Contains(rec.Body.String(), "access_token") {
		t.Fatalf("an unverified login response carries a token or data: %s", rec.Body.String())
	}

	// A wrong password gets the ordinary failure, not the verification hint.
	wrong := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/login", "", map[string]any{"email": email, "password": "not-the-password"})
	assertStatus(t, wrong, http.StatusUnauthorized)

	// After verification the login is unchanged: 200 with a bearer token
	// that authenticates a protected endpoint.
	verifyEmailViaHTTP(t, app.router, email)
	token := loginOnly(t, app.router, email, "password123")
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/users/me", token, nil), http.StatusOK)
}

func TestEmailVerificationAPIVerifyEndpoint(t *testing.T) {
	t.Parallel()
	app := setupRBACApp(t)
	const email = "verify.user@opspilot.dev"
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/register", "", registerRequest(email)), http.StatusCreated)
	mailbox := mailboxFor(t, app.router)
	raw := mailbox.latestToken(t, email)

	verify := func(token string) (int, map[string]any) {
		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/verify-email", "", map[string]any{"token": token})
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		return rec.Code, body
	}
	dataStatus := func(body map[string]any) any {
		data, _ := body["data"].(map[string]any)
		return data["status"]
	}

	if status, body := verify(strings.Repeat("A", 43)); status != http.StatusBadRequest || body["code"] != "verification_token_invalid" {
		t.Fatalf("unknown token = %d %v, want 400 verification_token_invalid", status, body)
	}
	if status, body := verify("short"); status != http.StatusBadRequest || body["code"] != "verification_token_invalid" {
		t.Fatalf("malformed token = %d %v, want 400 verification_token_invalid", status, body)
	}
	if rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/verify-email", "", map[string]any{}); rec.Code != http.StatusBadRequest {
		t.Fatalf("missing token = %d, want 400", rec.Code)
	}

	if err := app.db.Exec(`UPDATE email_verification_tokens SET expires_at = ?`, time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatalf("expire token: %v", err)
	}
	if status, body := verify(raw); status != http.StatusGone || body["code"] != "verification_token_expired" {
		t.Fatalf("expired token = %d %v, want 410 verification_token_expired", status, body)
	}
	if err := app.db.Exec(`UPDATE email_verification_tokens SET expires_at = ?`, time.Now().Add(time.Hour)).Error; err != nil {
		t.Fatalf("restore expiry: %v", err)
	}

	if status, body := verify(raw); status != http.StatusOK || dataStatus(body) != "verified" {
		t.Fatalf("valid token = %d %v, want 200 verified", status, body)
	}
	if status, body := verify(raw); status != http.StatusOK || dataStatus(body) != "already_verified" {
		t.Fatalf("reused token = %d %v, want 200 already_verified", status, body)
	}

	// The link must not work as a GET: verification is a POST so mail
	// scanners that pre-fetch links cannot consume it.
	if rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/auth/verify-email?token="+raw, "", nil); rec.Code == http.StatusOK {
		t.Fatalf("GET /auth/verify-email must not verify anything, got %d", rec.Code)
	}
}

func TestEmailVerificationAPIResendIsGenericAndRateLimited(t *testing.T) {
	t.Parallel()
	app := setupRBACApp(t)
	mailbox := mailboxFor(t, app.router)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/register", "", registerRequest("resend.user@opspilot.dev")), http.StatusCreated)

	resend := func(email string) (int, string) {
		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/resend-verification", "", map[string]any{"email": email})
		return rec.Code, rec.Body.String()
	}

	// A real unverified account (inside its cooldown) and an address nobody
	// registered must get byte-identical answers.
	knownStatus, knownBody := resend("resend.user@opspilot.dev")
	unknownStatus, unknownBody := resend("nobody@opspilot.dev")
	if knownStatus != http.StatusOK || unknownStatus != http.StatusOK || knownBody != unknownBody {
		t.Fatalf("resend answers differ by account:\nknown:   %d %s\nunknown: %d %s", knownStatus, knownBody, unknownStatus, unknownBody)
	}
	if mailbox.sentTo("nobody@opspilot.dev") != 0 {
		t.Fatalf("resend emailed an address that has no account")
	}
	if strings.Contains(knownBody, mailbox.latestToken(t, "resend.user@opspilot.dev")) {
		t.Fatalf("resend response exposes the token")
	}

	// A malformed address is a plain validation error, not an account probe.
	if status, _ := resend("not-an-email"); status != http.StatusBadRequest {
		t.Fatalf("malformed email = %d, want 400", status)
	}

	// Per-IP limit on resend: 5 requests are allowed (two used above: one
	// known, one unknown, and the malformed one is also counted).
	limited := 0
	for i := 0; i < 8; i++ {
		if status, _ := resend(fmt.Sprintf("probe-%d@opspilot.dev", i)); status == http.StatusTooManyRequests {
			limited++
		}
	}
	if limited == 0 {
		t.Fatalf("the resend endpoint was never rate limited")
	}
}

func TestEmailVerificationAPIRegistrationIsRateLimited(t *testing.T) {
	t.Parallel()
	app := setupRBACApp(t)

	limitedAt := 0
	for i := 1; i <= 30; i++ {
		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/register", "", registerRequest(fmt.Sprintf("bulk-%d@opspilot.dev", i)))
		if rec.Code == http.StatusTooManyRequests {
			limitedAt = i
			break
		}
		assertStatus(t, rec, http.StatusCreated)
	}
	if limitedAt == 0 || limitedAt > 25 {
		t.Fatalf("registration was not rate limited by request 25 (limited at %d)", limitedAt)
	}
}

func TestEmailVerificationAPIEmailFailureIsReportedAndLeavesNothing(t *testing.T) {
	t.Parallel()
	app := setupRBACApp(t)
	mailbox := mailboxFor(t, app.router)
	mailbox.setSendError(errors.New("smtp: dial tcp 10.0.0.7:587: connection refused"))

	rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/register", "", registerRequest("down.user@opspilot.dev"))
	assertStatus(t, rec, http.StatusServiceUnavailable)
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["code"] != "email_delivery_unavailable" {
		t.Fatalf("expected email_delivery_unavailable, got %v", body)
	}
	if strings.Contains(rec.Body.String(), "10.0.0.7") || strings.Contains(rec.Body.String(), "smtp") {
		t.Fatalf("the response leaks SMTP internals: %s", rec.Body.String())
	}
	if _, err := app.userRepo.GetByEmail("down.user@opspilot.dev"); err == nil {
		t.Fatalf("a registration whose email could not be sent left an account behind")
	}
}
