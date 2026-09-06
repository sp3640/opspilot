package constants

import "testing"

func TestIsValidIntegrationType(t *testing.T) {
	for _, valid := range ValidIntegrationTypes {
		if !IsValidIntegrationType(valid) {
			t.Fatalf("expected %q to be a valid integration type", valid)
		}
	}

	for _, invalid := range []string{"", "GITHUB", "gitlab", "bitbucket", "unknown"} {
		if IsValidIntegrationType(invalid) {
			t.Fatalf("expected %q to be rejected as an invalid integration type", invalid)
		}
	}
}

func TestIsValidIntegrationStatus(t *testing.T) {
	for _, valid := range ValidIntegrationStatuses {
		if !IsValidIntegrationStatus(valid) {
			t.Fatalf("expected %q to be a valid integration status", valid)
		}
	}

	for _, invalid := range []string{"", "pending", "unknown", "CONNECTING"} {
		if IsValidIntegrationStatus(invalid) {
			t.Fatalf("expected %q to be rejected as an invalid integration status", invalid)
		}
	}
}
