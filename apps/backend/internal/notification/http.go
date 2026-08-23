package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// httpTimeout bounds every outbound webhook-style call so a slow or
// unresponsive endpoint (Slack, Teams, or a third-party Webhook receiver)
// can never block the caller indefinitely. Dispatch is always best-effort
// and already runs off the critical path of the triggering operation, but a
// bound is still required so it eventually gives up.
const httpTimeout = 10 * time.Second

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: httpTimeout}
}

// postJSON posts payload as a JSON body to target and treats any non-2xx/3xx
// response as a failed delivery. target must be an absolute http(s) URL -
// NotificationService validates this at channel-configuration time, but the
// check is repeated here since a Provider must never be trusted to receive
// only pre-validated input.
func postJSON(ctx context.Context, client *http.Client, target string, payload any) error {
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		return fmt.Errorf("invalid webhook target url")
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode notification payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("build notification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send notification request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("notification endpoint returned status %d", resp.StatusCode)
	}

	return nil
}
