package services

import (
	"errors"
	"strings"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"gorm.io/gorm"
)

type CommentService struct {
	commentRepo  *repository.CommentRepository
	incidentRepo *repository.IncidentRepository
}

func NewCommentService(commentRepo *repository.CommentRepository, incidentRepo *repository.IncidentRepository) *CommentService {
	return &CommentService{
		commentRepo:  commentRepo,
		incidentRepo: incidentRepo,
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

	comment.Content = trimmedContent
	if err := s.commentRepo.Update(comment); err != nil {
		return nil, err
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

	return s.commentRepo.Delete(comment.ID)
}

func isValidCommentContent(content string) bool {
	if strings.TrimSpace(content) == "" {
		return false
	}
	return len([]rune(strings.TrimSpace(content))) <= 2000
}
