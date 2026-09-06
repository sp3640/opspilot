package services

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/security"
)

const superSecretTestToken = "SUPER_SECRET_TEST_TOKEN"

func testCipher(t *testing.T) security.ClusterCredentialCipher {
	t.Helper()
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	cipher, err := security.NewClusterCredentialCipher(base64.StdEncoding.EncodeToString(key))
	if err != nil {
		t.Fatalf("construct test cipher: %v", err)
	}
	return cipher
}

func TestRedactCredentialsNeverLeaksValues(t *testing.T) {
	redacted := redactCredentials(map[string]string{
		"token":    superSecretTestToken,
		"password": "hunter2",
	})

	for key, value := range redacted {
		if value == superSecretTestToken || value == "hunter2" {
			t.Fatalf("expected redacted value for key %q, got the real secret %q", key, value)
		}
		if value != redactedValue {
			t.Fatalf("expected every redacted value to equal %q, got %q", redactedValue, value)
		}
	}

	if _, ok := redacted["token"]; !ok {
		t.Fatalf("expected redaction to preserve the key name, only replace the value")
	}
}

func TestRedactCredentialsHandlesEmpty(t *testing.T) {
	if redactCredentials(nil) != nil {
		t.Fatalf("expected redacting nil credentials to return nil")
	}
	if redactCredentials(map[string]string{}) != nil {
		t.Fatalf("expected redacting empty credentials to return nil")
	}
}

func TestEncryptDecryptCredentialsRoundTrip(t *testing.T) {
	s := &IntegrationService{cipher: testCipher(t)}

	credentials := map[string]string{"token": superSecretTestToken}
	encrypted, err := s.encryptCredentials(credentials)
	if err != nil {
		t.Fatalf("encrypt credentials: %v", err)
	}
	if encrypted == "" {
		t.Fatalf("expected a non-empty ciphertext")
	}
	if strings.Contains(encrypted, superSecretTestToken) {
		t.Fatalf("expected the ciphertext to never contain the plaintext secret, got %q", encrypted)
	}

	integration := &models.Integration{EncryptedCredentials: encrypted}
	decrypted, err := s.decryptCredentials(integration)
	if err != nil {
		t.Fatalf("decrypt credentials: %v", err)
	}
	if decrypted["token"] != superSecretTestToken {
		t.Fatalf("expected round-tripped token %q, got %q", superSecretTestToken, decrypted["token"])
	}
}

func TestEncryptCredentialsFailsSafelyWithoutCipher(t *testing.T) {
	s := &IntegrationService{}

	if _, err := s.encryptCredentials(map[string]string{"token": superSecretTestToken}); err == nil {
		t.Fatalf("expected encryption to fail when no cipher is configured, never silently store plaintext")
	}
}

func TestEncryptCredentialsNoOpWhenEmpty(t *testing.T) {
	s := &IntegrationService{}

	encrypted, err := s.encryptCredentials(nil)
	if err != nil {
		t.Fatalf("expected no error encrypting empty credentials, got %v", err)
	}
	if encrypted != "" {
		t.Fatalf("expected empty ciphertext for empty credentials, got %q", encrypted)
	}
}

func TestMarshalUnmarshalMetadataRoundTrip(t *testing.T) {
	metadata := map[string]string{"workspace": "acme", "baseUrl": "https://example.com"}
	raw := marshalMetadata(metadata)
	roundTripped := unmarshalMetadata(raw)

	if roundTripped["workspace"] != "acme" || roundTripped["baseUrl"] != "https://example.com" {
		t.Fatalf("expected metadata to round-trip, got %v", roundTripped)
	}
}

func TestUnmarshalMetadataHandlesEmpty(t *testing.T) {
	if len(unmarshalMetadata("")) != 0 {
		t.Fatalf("expected empty metadata for an empty string")
	}
	if len(unmarshalMetadata("not json")) != 0 {
		t.Fatalf("expected empty metadata for invalid JSON, not an error")
	}
}
