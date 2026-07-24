package services

import (
	"errors"
	"log"
	"strings"

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

func (s *CommentService) CreateComment(content string, incidentID, userID uint) (*models.Comment, error) {
	trimmedContent := strings.TrimSpace(content)
	if !isValidCommentContent(trimmedContent) {
		return nil, apperrors.ErrInvalidCommentContent
	}

	if _, err := s.incidentRepo.GetByIDAndUserID(incidentID, userID); err != nil {
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
		incidentIDValue := incidentID
		incidentIDPtr := &incidentIDValue
		if err := s.auditRepo.LogCreate(userID, "comment", comment.ID, nil, incidentIDPtr); err != nil {
			log.Printf("audit create failed: %v", err)
		}
	}

	return comment, nil
}

func (s *CommentService) GetCommentsByIncidentID(incidentID, userID uint) ([]models.Comment, error) {
	if _, err := s.incidentRepo.GetByIDAndUserID(incidentID, userID); err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrIncidentNotFound
		}
		return nil, err
	}

	return s.commentRepo.GetByIncidentID(incidentID)
}

func (s *CommentService) ListCommentsByIncidentID(incidentID, userID uint, req *models.PaginationRequest) (*models.PaginationResponse, error) {
	if _, err := s.incidentRepo.GetByIDAndUserID(incidentID, userID); err != nil {
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

func (s *CommentService) UpdateComment(id, userID uint, content string) (*models.Comment, error) {
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
		if err := s.auditRepo.LogUpdate(userID, "comment", comment.ID, nil, incidentIDPtr, "content", previousContent, trimmedContent); err != nil {
			log.Printf("audit update failed: %v", err)
		}
	}

	return comment, nil
}

func (s *CommentService) DeleteComment(id, userID uint) error {
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
		if err := s.auditRepo.LogDelete(userID, "comment", comment.ID, nil, incidentIDPtr); err != nil {
			log.Printf("audit delete failed: %v", err)
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
