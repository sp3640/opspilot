package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type InvitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) *InvitationRepository {
	return &InvitationRepository{db: db}
}

func (r *InvitationRepository) Create(invitation *models.Invitation) error {
	return r.db.Create(invitation).Error
}

func (r *InvitationRepository) GetByID(id uuid.UUID) (*models.Invitation, error) {
	var invitation models.Invitation
	if err := r.db.Where("id = ?", id).First(&invitation).Error; err != nil {
		return nil, err
	}

	return &invitation, nil
}

func (r *InvitationRepository) GetByToken(token string) (*models.Invitation, error) {
	var invitation models.Invitation
	if err := r.db.Where("token = ?", token).First(&invitation).Error; err != nil {
		return nil, err
	}

	return &invitation, nil
}

func (r *InvitationRepository) GetByEmail(organizationID uuid.UUID, email string) (*models.Invitation, error) {
	var invitation models.Invitation
	if err := r.db.
		Where("organization_id = ? AND email = ? AND status = ?", organizationID, email, models.InvitationStatusPending).
		Order("created_at DESC").
		First(&invitation).Error; err != nil {
		return nil, err
	}

	return &invitation, nil
}

func (r *InvitationRepository) ListByOrganization(req *models.PaginationRequest, organizationID uuid.UUID) ([]models.Invitation, int64, error) {
	if err := req.Validate("created_at", "updated_at", "email", "expires_at", "status"); err != nil {
		return nil, 0, err
	}

	query := r.db.Model(&models.Invitation{}).Where("organization_id = ?", organizationID)
	if req.Search != "" {
		query = query.Where("email ILIKE ?", "%"+req.Search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortField := "created_at"
	switch req.Sort {
	case "updated_at":
		sortField = "updated_at"
	case "email":
		sortField = "email"
	case "expires_at":
		sortField = "expires_at"
	case "status":
		sortField = "status"
	case "created_at":
		sortField = "created_at"
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	items := make([]models.Invitation, 0)
	if err := query.Order(sortField + " " + order).Limit(req.Limit).Offset((req.Page - 1) * req.Limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *InvitationRepository) Update(invitation *models.Invitation) error {
	return r.db.Model(invitation).Updates(invitation).Error
}

func (r *InvitationRepository) Delete(id uuid.UUID, organizationID uuid.UUID) error {
	return r.db.Where("id = ? AND organization_id = ?", id, organizationID).Delete(&models.Invitation{}).Error
}

func (r *InvitationRepository) ExpireOldInvitations(now time.Time) (int64, error) {
	result := r.db.Model(&models.Invitation{}).
		Where("status = ? AND expires_at < ?", models.InvitationStatusPending, now).
		Updates(map[string]any{"status": models.InvitationStatusExpired})
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
