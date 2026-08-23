package services

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/sp3640/opspilot/backend/internal/logger"
)

// marshalAuditState serializes v (typically a small, hand-built map or
// struct literal the caller controls explicitly) into the JSON text stored
// in AuditLog.BeforeState/AfterState. Callers must only pass values they
// have already vetted to exclude secrets (password hashes, invitation
// tokens, kubeconfigs, API credentials) - this function does not redact
// anything itself. A marshal failure yields an empty string rather than an
// error, since audit snapshots are always best-effort.
func marshalAuditState(v any) string {
	if v == nil {
		return ""
	}

	encoded, err := json.Marshal(v)
	if err != nil {
		return ""
	}

	return string(encoded)
}

// logAuditFailure keeps audit persistence best-effort: a failure is visible to
// operators, correlated to the originating request, and never rolls back a
// completed business operation.
func logAuditFailure(ctx context.Context, operation, entityType string, entityID uint, err error) {
	if err == nil {
		return
	}

	logger.Error(
		ctx,
		"audit logging failed",
		slog.String("operation", operation),
		slog.String("entity_type", entityType),
		slog.Uint64("entity_id", uint64(entityID)),
		slog.Any("error", err),
	)
}
