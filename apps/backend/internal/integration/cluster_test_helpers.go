package integration

import (
	"encoding/base64"
	"testing"

	"github.com/sp3640/opspilot/backend/internal/security"
)

func testClusterCredentialCipher(t testing.TB) security.ClusterCredentialCipher {
	t.Helper()

	cipher, err := security.NewClusterCredentialCipher(testClusterEncryptionKey())
	if err != nil {
		t.Fatalf("create test cluster credential cipher: %v", err)
	}

	return cipher
}

func testClusterEncryptionKey() string {
	return base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
}
