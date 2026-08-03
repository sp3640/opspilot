package resourcesync

import (
	"context"

	"github.com/sp3640/opspilot/backend/internal/models"
)

type ResourceStore interface {
	ListResources(ctx context.Context, projectID string) ([]models.Resource, error)
	CreateResources(ctx context.Context, resources []models.Resource) error
	UpdateResources(ctx context.Context, resources []models.Resource) error
	SoftDeleteResources(ctx context.Context, resources []models.Resource) error
	RestoreResources(ctx context.Context, resources []models.Resource) error
}
