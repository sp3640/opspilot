package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/mapper"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"gorm.io/gorm"
)

type MetricService struct {
	repo      *repository.MetricRepository
	auditRepo *AuditService
}

func NewMetricService(repo *repository.MetricRepository, auditService *AuditService) *MetricService {
	return &MetricService{
		repo:      repo,
		auditRepo: auditService,
	}
}

func (s *MetricService) StoreMetrics(ctx context.Context, projectID uuid.UUID, userID uint, requests []dto.CreateMetricRequest) ([]dto.MetricResponse, error) {
	organizationID, err := s.repo.GetProjectOrganizationID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrInvalidProject
		}
		return nil, err
	}

	if len(requests) == 0 {
		return []dto.MetricResponse{}, nil
	}

	now := time.Now().UTC()
	items := make([]models.Metric, 0, len(requests))

	for _, request := range requests {
		metric, err := s.buildMetricModel(projectID, organizationID, request, now)
		if err != nil {
			return nil, err
		}
		items = append(items, metric)
	}

	if err := s.repo.BulkCreate(items); err != nil {
		return nil, err
	}

	return mapper.MapMetrics(items), nil
}

func (s *MetricService) StoreSnapshot(ctx context.Context, projectID uuid.UUID, userID uint, requests []dto.CreateMetricRequest) ([]dto.MetricResponse, error) {
	organizationID := uuid.Nil
	stored, err := s.StoreMetrics(ctx, projectID, userID, requests)
	if err != nil {
		return nil, err
	}

	if organizationID == uuid.Nil {
		organizationID, err = s.repo.GetProjectOrganizationID(projectID)
		if err != nil {
			organizationID = uuid.Nil
		}
	}

	if s.auditRepo != nil && len(stored) > 0 {
		snapshotID := projectID.String() + ":" + time.Now().UTC().Format(time.RFC3339)
		if organizationID != uuid.Nil {
			if err := s.auditRepo.LogCreate(userID, organizationID, "metric_snapshot", snapshotID, &projectID, nil); err != nil {
				logAuditFailure(ctx, "snapshot_store", "metric_snapshot", 0, err)
			}
		}
	}

	return stored, nil
}

func (s *MetricService) GetMetrics(organizationID uuid.UUID, req *models.PaginationRequest) (*dto.MetricListResponse, error) {
	items, total, err := s.repo.List(req, organizationID)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	return &dto.MetricListResponse{
		Items:      mapper.MapMetrics(items),
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *MetricService) GetLatest(organizationID uuid.UUID, projectID uuid.UUID, clusterID, resourceID *uuid.UUID, metricType, metricName string) (*dto.MetricResponse, error) {
	if err := s.validateProjectOwnership(projectID, organizationID); err != nil {
		return nil, err
	}

	normalizedType, err := normalizeMetricType(metricType)
	if err != nil {
		return nil, err
	}

	metric, err := s.repo.Latest(projectID, clusterID, resourceID, normalizedType, strings.TrimSpace(metricName), organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	response := mapper.MapMetric(*metric)
	return &response, nil
}

func (s *MetricService) GetHistory(organizationID uuid.UUID, projectID uuid.UUID, clusterID, resourceID *uuid.UUID, metricType, metricName string, startTime, endTime *time.Time, limit int) ([]dto.MetricResponse, error) {
	if err := s.validateProjectOwnership(projectID, organizationID); err != nil {
		return nil, err
	}

	normalizedType, err := normalizeMetricType(metricType)
	if err != nil {
		return nil, err
	}

	metricName = strings.TrimSpace(metricName)

	var items []models.Metric
	if resourceID != nil {
		items, err = s.repo.ListByResource(*resourceID, normalizedType, metricName, startTime, endTime, limit, organizationID)
	} else if clusterID != nil {
		items, err = s.repo.ListByCluster(*clusterID, normalizedType, metricName, startTime, endTime, limit, organizationID)
	} else {
		items, err = s.repo.ListByProject(projectID, normalizedType, metricName, startTime, endTime, limit, organizationID)
	}
	if err != nil {
		return nil, err
	}

	return mapper.MapMetrics(items), nil
}

func (s *MetricService) Aggregate(organizationID uuid.UUID, projectID uuid.UUID, metricType, metricName, interval string, startTime, endTime time.Time) (*dto.MetricAggregateResponse, error) {
	if err := s.validateProjectOwnership(projectID, organizationID); err != nil {
		return nil, err
	}

	normalizedType, err := normalizeMetricType(metricType)
	if err != nil {
		return nil, err
	}

	if startTime.IsZero() {
		startTime = time.Now().UTC().Add(-24 * time.Hour)
	}
	if endTime.IsZero() {
		endTime = time.Now().UTC()
	}
	if endTime.Before(startTime) {
		return nil, apperrors.ErrInvalidTimeRange
	}

	metricName = strings.TrimSpace(metricName)
	points, err := s.repo.Aggregate(projectID, normalizedType, metricName, startTime, endTime, interval, organizationID)
	if err != nil {
		return nil, err
	}

	var count int64
	var sum float64
	min := 0.0
	max := 0.0
	for index, point := range points {
		count += point.Count
		sum += point.Value * float64(point.Count)
		if index == 0 || point.Value < min {
			min = point.Value
		}
		if index == 0 || point.Value > max {
			max = point.Value
		}
	}

	average := 0.0
	if count > 0 {
		average = sum / float64(count)
	}

	unit := ""
	latest, latestErr := s.repo.Latest(projectID, nil, nil, normalizedType, metricName, organizationID)
	if latestErr == nil {
		unit = latest.Unit
	}

	return &dto.MetricAggregateResponse{
		ProjectID:  projectID.String(),
		MetricType: normalizedType,
		MetricName: metricName,
		Unit:       unit,
		Interval:   strings.TrimSpace(strings.ToLower(interval)),
		StartTime:  startTime,
		EndTime:    endTime,
		Count:      count,
		Average:    average,
		Minimum:    min,
		Maximum:    max,
		Sum:        sum,
		Points:     mapper.MapMetricAggregatePoints(points),
	}, nil
}

func (s *MetricService) buildMetricModel(projectID, organizationID uuid.UUID, request dto.CreateMetricRequest, now time.Time) (models.Metric, error) {
	if request.ProjectID != uuid.Nil && request.ProjectID != projectID {
		return models.Metric{}, apperrors.ErrInvalidProject
	}

	metricType, err := normalizeMetricType(request.MetricType)
	if err != nil {
		return models.Metric{}, err
	}

	resourceKind := strings.TrimSpace(request.ResourceKind)
	if !constants.IsValidResourceKind(resourceKind) {
		return models.Metric{}, apperrors.ErrInvalidResourceKind
	}

	metricName := strings.TrimSpace(request.MetricName)
	if metricName == "" {
		return models.Metric{}, fmt.Errorf("metric name is required")
	}

	labels := request.Labels
	if len(labels) == 0 {
		labels = json.RawMessage(`{}`)
	}

	metadata := request.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}

	timestamp := request.Timestamp.UTC()
	if timestamp.IsZero() {
		timestamp = now
	}

	if request.ClusterID == uuid.Nil {
		return models.Metric{}, fmt.Errorf("cluster id is required")
	}
	if request.ResourceID == uuid.Nil {
		return models.Metric{}, fmt.Errorf("resource id is required")
	}

	return models.Metric{
		OrganizationID: organizationID,
		ProjectID:      projectID,
		ClusterID:      request.ClusterID,
		ResourceID:     request.ResourceID,
		ResourceKind:   resourceKind,
		MetricType:     metricType,
		MetricName:     metricName,
		Value:          request.Value,
		Unit:           strings.TrimSpace(request.Unit),
		Timestamp:      timestamp,
		Labels:         labels,
		Metadata:       metadata,
	}, nil
}

func (s *MetricService) validateProjectOwnership(projectID, organizationID uuid.UUID) error {
	belongs, err := s.repo.ProjectBelongsToOrganization(projectID, organizationID)
	if err != nil {
		return err
	}
	if !belongs {
		return apperrors.ErrInvalidProject
	}

	return nil
}

func normalizeMetricType(metricType string) (string, error) {
	normalized := strings.TrimSpace(strings.ToUpper(metricType))
	if !constants.IsValidMetricType(normalized) {
		return "", apperrors.ErrInvalidMetricType
	}

	return normalized, nil
}
