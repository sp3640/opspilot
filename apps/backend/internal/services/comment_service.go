package services

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"gorm.io/gorm"
)

type CommentService struct {
	commentRepo  *repository.CommentRepository
	incidentRepo *repository.IncidentRepository
	auditRepo    *AuditService
}

func NewCommentService(commentRepo *repository.CommentRepository, incidentRepo *repository.IncidentRepository, auditService *AuditService) *CommentService {
	return &CommentService{
		commentRepo:  commentRepo,
		incidentRepo: incidentRepo,
		auditRepo:    auditService,
	}
}

func (s *CommentService) CreateComment(ctx context.Context, content string, incidentID, userID uint, organizationID uuid.UUID) (*models.Comment, error) {
	trimmedContent := strings.TrimSpace(content)
	if !isValidCommentContent(trimmedContent) {
		return nil, apperrors.ErrInvalidCommentContent
	}

	if _, err := s.incidentRepo.GetByIDAndOrganizationID(incidentID, organizationID); err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrIncidentNotFound
		}
		return nil, err
	}

	comment := &models.Comment{
		Content:    trimmedContent,
		IncidentID: incidentID,
		UserID:     userID,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		incidentIDValue := comment.IncidentID
		incidentIDPtr := &incidentIDValue
		if err := s.auditRepo.LogCreate(userID, organizationID, "comment", strconv.FormatUint(uint64(comment.ID), 10), nil, incidentIDPtr); err != nil {
			logAuditFailure(ctx, "create", "comment", comment.ID, err)
		}
	}

	return comment, nil
}

func (s *CommentService) ListCommentsByIncidentID(incidentID uint, organizationID uuid.UUID, req *models.PaginationRequest) (*models.PaginationResponse, error) {
	if _, err := s.incidentRepo.GetByIDAndOrganizationID(incidentID, organizationID); err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrIncidentNotFound
		}
		return nil, err
	}

	items, total, err := s.commentRepo.ListByIncidentID(req, incidentID)
	if err != nil {
		return nil, err
	}

	return &models.PaginationResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: int((total + int64(req.Limit) - 1) / int64(req.Limit)),
		Items:      items,
	}, nil
}

func (s *CommentService) UpdateComment(ctx context.Context, id, userID uint, organizationID uuid.UUID, content string) (*models.Comment, error) {
	comment, err := s.commentRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrCommentNotFound
		}
		return nil, err
	}

	if comment.UserID != userID {
		return nil, apperrors.ErrCommentForbidden
	}

	trimmedContent := strings.TrimSpace(content)
	if !isValidCommentContent(trimmedContent) {
		return nil, apperrors.ErrInvalidCommentContent
	}

	previousContent := comment.Content
	comment.Content = trimmedContent
	if err := s.commentRepo.Update(comment); err != nil {
		return nil, err
	}

	if s.auditRepo != nil && previousContent != trimmedContent {
		incidentIDValue := comment.IncidentID
		incidentIDPtr := &incidentIDValue
		if err := s.auditRepo.LogUpdate(userID, organizationID, "comment", strconv.FormatUint(uint64(comment.ID), 10), nil, incidentIDPtr, "content", previousContent, trimmedContent); err != nil {
			logAuditFailure(ctx, "update", "comment", comment.ID, err)
		}
	}

	return comment, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, id, userID uint, organizationID uuid.UUID) error {
	comment, err := s.commentRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrCommentNotFound
		}
		return err
	}

	if comment.UserID != userID {
		return apperrors.ErrCommentForbidden
	}

	if err := s.commentRepo.Delete(comment.ID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		incidentIDValue := comment.IncidentID
		incidentIDPtr := &incidentIDValue
		if err := s.auditRepo.LogDelete(userID, organizationID, "comment", strconv.FormatUint(uint64(comment.ID), 10), nil, incidentIDPtr); err != nil {
			logAuditFailure(ctx, "delete", "comment", comment.ID, err)
		}
	}

	return nil
}

func isValidCommentContent(content string) bool {
	if strings.TrimSpace(content) == "" {
		return false
	}
	return len([]rune(strings.TrimSpace(content))) <= 2000
}
