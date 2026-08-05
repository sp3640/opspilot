package executor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/logger"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
)

type deploymentHistoryWriter interface {
	CreateHistoryFromDeployment(deployment *models.Deployment, changeSummary string, triggeredBy uint) error
}

type auditLogger interface {
	LogUpdate(userID uint, organizationID uuid.UUID, entityType string, entityID string, projectID *uuid.UUID, incidentID *uint, fieldName string, oldValue string, newValue string) error
}

type DeploymentStatusUpdater struct {
	deploymentRepo *repository.DeploymentRepository
	historyWriter  deploymentHistoryWriter
	auditLogger    auditLogger
}

func NewDeploymentStatusUpdater(
	deploymentRepo *repository.DeploymentRepository,
	historyWriter deploymentHistoryWriter,
	auditLogger auditLogger,
) *DeploymentStatusUpdater {
	return &DeploymentStatusUpdater{
		deploymentRepo: deploymentRepo,
		historyWriter:  historyWriter,
		auditLogger:    auditLogger,
	}
}

func (u *DeploymentStatusUpdater) MarkRunning(ctx context.Context, deployment *models.Deployment, userID uint) error {
	now := time.Now().UTC()
	previousStatus := deployment.Status

	if err := u.deploymentRepo.UpdateStatus(deployment.ID, deployment.OrganizationID, constants.DeploymentStatusRunning, &now, nil, userID); err != nil {
		return err
	}

	deployment.Status = constants.DeploymentStatusRunning
	deployment.StartedAt = &now
	deployment.CompletedAt = nil
	deployment.UpdatedBy = userID

	if u.historyWriter != nil {
		if err := u.historyWriter.CreateHistoryFromDeployment(deployment, "Deployment execution started", userID); err != nil {
			return err
		}
	}

	u.logStatusAuditBestEffort(ctx, deployment, userID, previousStatus, constants.DeploymentStatusRunning)
	return nil
}

func (u *DeploymentStatusUpdater) MarkSucceeded(ctx context.Context, deployment *models.Deployment, userID uint) error {
	now := time.Now().UTC()
	previousStatus := deployment.Status

	if err := u.deploymentRepo.UpdateStatus(deployment.ID, deployment.OrganizationID, constants.DeploymentStatusSucceeded, nil, &now, userID); err != nil {
		return err
	}

	deployment.Status = constants.DeploymentStatusSucceeded
	deployment.CompletedAt = &now
	deployment.UpdatedBy = userID

	if u.historyWriter != nil {
		if err := u.historyWriter.CreateHistoryFromDeployment(deployment, "Deployment execution succeeded", userID); err != nil {
			return err
		}
	}

	u.logStatusAuditBestEffort(ctx, deployment, userID, previousStatus, constants.DeploymentStatusSucceeded)
	return nil
}

func (u *DeploymentStatusUpdater) MarkFailed(ctx context.Context, deployment *models.Deployment, userID uint, reason string) error {
	now := time.Now().UTC()
	previousStatus := deployment.Status

	if err := u.deploymentRepo.UpdateStatus(deployment.ID, deployment.OrganizationID, constants.DeploymentStatusFailed, nil, &now, userID); err != nil {
		return err
	}

	deployment.Status = constants.DeploymentStatusFailed
	deployment.CompletedAt = &now
	deployment.UpdatedBy = userID

	if u.historyWriter != nil {
		summary := "Deployment execution failed"
		if reason != "" {
			summary = fmt.Sprintf("Deployment execution failed: %s", reason)
		}
		if err := u.historyWriter.CreateHistoryFromDeployment(deployment, summary, userID); err != nil {
			return err
		}
	}

	u.logStatusAuditBestEffort(ctx, deployment, userID, previousStatus, constants.DeploymentStatusFailed)
	return nil
}

func (u *DeploymentStatusUpdater) logStatusAuditBestEffort(ctx context.Context, deployment *models.Deployment, userID uint, oldStatus, newStatus string) {
	if u.auditLogger == nil {
		return
	}

	if err := u.auditLogger.LogUpdate(
		userID,
		deployment.OrganizationID,
		"deployment",
		deployment.ID.String(),
		&deployment.ProjectID,
		nil,
		"status",
		oldStatus,
		newStatus,
	); err != nil {
		logger.Error(
			ctx,
			"audit logging failed",
			slog.String("operation", "update"),
			slog.String("entity_type", "deployment"),
			slog.String("entity_id", deployment.ID.String()),
			slog.Any("error", err),
		)
	}
}
