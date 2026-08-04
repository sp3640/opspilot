package repository

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type MetricRepository struct {
	db *gorm.DB
}

func NewMetricRepository(db *gorm.DB) *MetricRepository {
	return &MetricRepository{db: db}
}

func (r *MetricRepository) Create(metric *models.Metric) error {
	return r.db.Create(metric).Error
}

func (r *MetricRepository) BulkCreate(metrics []models.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	return r.db.Create(&metrics).Error
}

func (r *MetricRepository) List(req *models.PaginationRequest, organizationID uuid.UUID) ([]models.Metric, int64, error) {
	if err := req.Validate("created_at", "timestamp", "metric_type", "metric_name", "resource_kind", "value"); err != nil {
		return nil, 0, err
	}

	query := r.baseOwnedQuery(organizationID)

	if req.Search != "" {
		query = query.Where(
			"metrics.metric_name ILIKE ? OR metrics.resource_kind ILIKE ? OR metrics.unit ILIKE ?",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
		)
	}

	if req.ProjectID != uuid.Nil {
		query = query.Where("metrics.project_id = ?", req.ProjectID)
	}

	if req.Kind != "" {
		query = query.Where("metrics.resource_kind = ?", req.Kind)
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
		case "timestamp":
			sortField = "timestamp"
		case "metric_type":
			sortField = "metric_type"
		case "metric_name":
			sortField = "metric_name"
		case "resource_kind":
			sortField = "resource_kind"
		case "value":
			sortField = "value"
		}
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	items := make([]models.Metric, 0)
	err := query.
		Order("metrics." + sortField + " " + order).
		Limit(req.Limit).
		Offset((req.Page - 1) * req.Limit).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *MetricRepository) ListByResource(resourceID uuid.UUID, metricType, metricName string, startTime, endTime *time.Time, limit int, organizationID uuid.UUID) ([]models.Metric, error) {
	query := r.baseOwnedQuery(organizationID).Where("metrics.resource_id = ?", resourceID)
	query = applyMetricFilters(query, metricType, metricName, startTime, endTime)

	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	items := make([]models.Metric, 0)
	if err := query.Order("metrics.timestamp DESC").Limit(limit).Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (r *MetricRepository) ListByCluster(clusterID uuid.UUID, metricType, metricName string, startTime, endTime *time.Time, limit int, organizationID uuid.UUID) ([]models.Metric, error) {
	query := r.baseOwnedQuery(organizationID).Where("metrics.cluster_id = ?", clusterID)
	query = applyMetricFilters(query, metricType, metricName, startTime, endTime)

	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	items := make([]models.Metric, 0)
	if err := query.Order("metrics.timestamp DESC").Limit(limit).Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (r *MetricRepository) ListByProject(projectID uuid.UUID, metricType, metricName string, startTime, endTime *time.Time, limit int, organizationID uuid.UUID) ([]models.Metric, error) {
	query := r.baseOwnedQuery(organizationID).Where("metrics.project_id = ?", projectID)
	query = applyMetricFilters(query, metricType, metricName, startTime, endTime)

	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	items := make([]models.Metric, 0)
	if err := query.Order("metrics.timestamp DESC").Limit(limit).Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

// DeleteOlderThan is reserved for scheduled retention jobs that prune stale metrics.
func (r *MetricRepository) DeleteOlderThan(projectID uuid.UUID, cutoff time.Time) (int64, error) {
	result := r.db.Where("project_id = ? AND timestamp < ?", projectID, cutoff).Delete(&models.Metric{})
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (r *MetricRepository) Aggregate(projectID uuid.UUID, metricType, metricName string, startTime, endTime time.Time, interval string, organizationID uuid.UUID) ([]models.MetricAggregatePoint, error) {
	interval = normalizeInterval(interval)

	query := r.baseOwnedQuery(organizationID).
		Where("metrics.project_id = ?", projectID).
		Where("metrics.timestamp >= ? AND metrics.timestamp <= ?", startTime, endTime)

	if metricType != "" {
		query = query.Where("metrics.metric_type = ?", metricType)
	}
	if metricName != "" {
		query = query.Where("metrics.metric_name = ?", metricName)
	}

	type metricAggregateRow struct {
		Bucket time.Time
		Value  float64
		Count  int64
	}

	rows := make([]metricAggregateRow, 0)
	err := query.
		Select("date_trunc(?, metrics.timestamp) AS bucket, AVG(metrics.value) AS value, COUNT(*) AS count", interval).
		Group("bucket").
		Order("bucket ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	points := make([]models.MetricAggregatePoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, models.MetricAggregatePoint{
			Bucket: row.Bucket,
			Value:  row.Value,
			Count:  row.Count,
		})
	}

	return points, nil
}

func (r *MetricRepository) Latest(projectID uuid.UUID, clusterID *uuid.UUID, resourceID *uuid.UUID, metricType, metricName string, organizationID uuid.UUID) (*models.Metric, error) {
	query := r.baseOwnedQuery(organizationID).Where("metrics.project_id = ?", projectID)

	if clusterID != nil {
		query = query.Where("metrics.cluster_id = ?", *clusterID)
	}
	if resourceID != nil {
		query = query.Where("metrics.resource_id = ?", *resourceID)
	}
	if metricType != "" {
		query = query.Where("metrics.metric_type = ?", metricType)
	}
	if metricName != "" {
		query = query.Where("metrics.metric_name = ?", metricName)
	}

	var metric models.Metric
	if err := query.Order("metrics.timestamp DESC").Order("metrics.created_at DESC").First(&metric).Error; err != nil {
		return nil, err
	}

	return &metric, nil
}

func (r *MetricRepository) ProjectBelongsToOrganization(projectID, organizationID uuid.UUID) (bool, error) {
	var project models.Project

	if err := r.db.Where("id = ? AND organization_id = ?", projectID, organizationID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *MetricRepository) GetProjectOrganizationID(projectID uuid.UUID) (uuid.UUID, error) {
	var project models.Project
	if err := r.db.Select("organization_id").Where("id = ?", projectID).First(&project).Error; err != nil {
		return uuid.Nil, err
	}

	return project.OrganizationID, nil
}

func (r *MetricRepository) baseOwnedQuery(organizationID uuid.UUID) *gorm.DB {
	return r.db.Model(&models.Metric{}).
		Where("metrics.organization_id = ?", organizationID)
}

func applyMetricFilters(query *gorm.DB, metricType, metricName string, startTime, endTime *time.Time) *gorm.DB {
	if metricType != "" {
		query = query.Where("metrics.metric_type = ?", strings.TrimSpace(strings.ToUpper(metricType)))
	}
	if metricName != "" {
		query = query.Where("metrics.metric_name = ?", strings.TrimSpace(metricName))
	}
	if startTime != nil && !startTime.IsZero() {
		query = query.Where("metrics.timestamp >= ?", *startTime)
	}
	if endTime != nil && !endTime.IsZero() {
		query = query.Where("metrics.timestamp <= ?", *endTime)
	}

	return query
}

func normalizeInterval(interval string) string {
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "minute":
		return "minute"
	case "day":
		return "day"
	default:
		return "hour"
	}
}
