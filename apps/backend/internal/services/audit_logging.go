package services

import (
	"context"
	"log/slog"

	"github.com/sp3640/opspilot/backend/internal/logger"
)

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
