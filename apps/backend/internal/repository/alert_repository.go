package repository

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type AlertRepository struct {
	db *gorm.DB
}

func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

func (r *AlertRepository) Create(alert *models.Alert) error {
	return r.db.Create(alert).Error
}

func (r *AlertRepository) Update(alert *models.Alert) error {
	return r.db.Model(alert).Updates(alert).Error
}

func (r *AlertRepository) Delete(id uint) error {
	return r.db.Delete(&models.Alert{}, id).Error
}

func (r *AlertRepository) FindByID(id uint) (*models.Alert, error) {
	var alert models.Alert
	if err := r.db.First(&alert, id).Error; err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *AlertRepository) FindByFingerprint(projectID uuid.UUID, fingerprint string) (*models.Alert, error) {
	var alert models.Alert
	if err := r.db.Where("project_id = ? AND fingerprint = ?", projectID, fingerprint).First(&alert).Error; err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *AlertRepository) List(req *models.PaginationRequest, userID uint) ([]models.Alert, int64, error) {
	if err := req.Validate("created_at", "updated_at", "severity", "status", "last_seen_at"); err != nil {
		return nil, 0, err
	}

	query := r.db.Model(&models.Alert{}).
		Joins("JOIN projects ON projects.id = alerts.project_id").
		Where("projects.owner_id = ?", userID)

	if req.Search != "" {
		query = query.Where(
			"alerts.title ILIKE ? OR alerts.description ILIKE ? OR alerts.resource_id ILIKE ?",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
		)
	}

	if req.Status != "" {
		query = query.Where("alerts.status = ?", req.Status)
	}

	if req.Severity != "" {
		query = query.Where("alerts.severity = ?", req.Severity)
	}

	if req.Source != "" {
		query = query.Where("alerts.source = ?", req.Source)
	}

	if req.ProjectID != uuid.Nil {
		query = query.Where("alerts.project_id = ?", req.ProjectID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortField := "created_at"
	if req.Sort != "" {
		switch req.Sort {
		case "created_at":
			sortField = "created_at"
		case "updated_at":
			sortField = "updated_at"
		case "severity":
			sortField = "severity"
		case "status":
			sortField = "status"
		case "last_seen_at":
			sortField = "last_seen_at"
		}
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	var alerts []models.Alert
	err := query.
		Order("alerts." + sortField + " " + order).
		Limit(req.Limit).
		Offset((req.Page - 1) * req.Limit).
		Find(&alerts).Error
	if err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}

func (r *AlertRepository) IncrementOccurrence(id uint, lastSeenAt time.Time, metadata json.RawMessage) error {
	updates := map[string]interface{}{
		"occurrence_count": gorm.Expr("occurrence_count + 1"),
		"last_seen_at":     lastSeenAt,
	}

	if len(metadata) > 0 {
		updates["metadata"] = metadata
	}

	return r.db.Model(&models.Alert{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AlertRepository) Resolve(id uint, resolvedAt time.Time) error {
	return r.db.Model(&models.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      constants.AlertStatusResolved,
		"resolved_at": resolvedAt,
	}).Error
}

func (r *AlertRepository) Acknowledge(id uint, acknowledgedAt time.Time) error {
	return r.db.Model(&models.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":          constants.AlertStatusAcknowledged,
		"acknowledged_at": acknowledgedAt,
	}).Error
}

func (r *AlertRepository) AttachIncident(id uint, incidentID uint) error {
	return r.db.Model(&models.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"incident_id": incidentID,
		"status":      constants.AlertStatusInvestigating,
	}).Error
}

func (r *AlertRepository) Reopen(id uint, lastSeenAt time.Time) error {
	return r.db.Model(&models.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       constants.AlertStatusOpen,
		"resolved_at":  nil,
		"last_seen_at": lastSeenAt,
	}).Error
}

// RefreshLastSeen is reserved for background deduplication/heartbeat jobs that touch alert freshness.
func (r *AlertRepository) RefreshLastSeen(id uint, lastSeenAt time.Time) error {
	return r.db.Model(&models.Alert{}).Where("id = ?", id).Update("last_seen_at", lastSeenAt).Error
}

func (r *AlertRepository) ProjectBelongsToUser(projectID uuid.UUID, userID uint) (bool, error) {
	var project models.Project

	if err := r.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	return project.OwnerID == userID, nil
}
