package repository

import (
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type NotificationChannelRepository struct {
	db *gorm.DB
}

func NewNotificationChannelRepository(db *gorm.DB) *NotificationChannelRepository {
	return &NotificationChannelRepository{db: db}
}

func (r *NotificationChannelRepository) Create(channel *models.NotificationChannel) error {
	return r.db.Create(channel).Error
}

func (r *NotificationChannelRepository) GetByID(id, organizationID uuid.UUID) (*models.NotificationChannel, error) {
	var channel models.NotificationChannel
	if err := r.db.Where("id = ? AND organization_id = ?", id, organizationID).First(&channel).Error; err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *NotificationChannelRepository) Update(channel *models.NotificationChannel) error {
	return r.db.Save(channel).Error
}

func (r *NotificationChannelRepository) Delete(id, organizationID uuid.UUID) error {
	return r.db.Where("id = ? AND organization_id = ?", id, organizationID).Delete(&models.NotificationChannel{}).Error
}

// List returns a paginated, optionally team/type-filtered set of channels
// for organizationID. teamID == uuid.Nil / channelType == "" mean "no filter".
func (r *NotificationChannelRepository) List(req *models.PaginationRequest, organizationID, teamID uuid.UUID, channelType string) ([]models.NotificationChannel, int64, error) {
	if err := req.Validate("created_at", "name", "type"); err != nil {
		return nil, 0, err
	}

	query := r.db.Model(&models.NotificationChannel{}).Where("organization_id = ?", organizationID)
	if teamID != uuid.Nil {
		query = query.Where("team_id = ?", teamID)
	}
	if channelType != "" {
		query = query.Where("type = ?", channelType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	var channels []models.NotificationChannel
	err := query.
		Order(req.Sort + " " + order).
		Limit(req.Limit).
		Offset((req.Page - 1) * req.Limit).
		Find(&channels).Error
	if err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

// ListEnabledForDispatch returns every enabled organization-level channel
// (TeamID IS NULL) plus, when teamID is non-nil, every enabled channel
// scoped to that specific team - the full fan-out set for one event.
// Event-subscription filtering happens in the service layer, since Events
// is an opaque JSON blob to the database.
func (r *NotificationChannelRepository) ListEnabledForDispatch(organizationID uuid.UUID, teamID *uuid.UUID) ([]models.NotificationChannel, error) {
	query := r.db.Where("organization_id = ? AND enabled = ?", organizationID, true)
	if teamID != nil {
		query = query.Where("team_id IS NULL OR team_id = ?", *teamID)
	} else {
		query = query.Where("team_id IS NULL")
	}

	var channels []models.NotificationChannel
	if err := query.Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}
