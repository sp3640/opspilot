package services

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/utils"
	"gorm.io/gorm"
)

type OrganizationService struct {
	repo *repository.OrganizationRepository
}

func NewOrganizationService(repo *repository.OrganizationRepository) *OrganizationService {
	return &OrganizationService{repo: repo}
}

func (s *OrganizationService) Create(ownerID uint, req dto.CreateOrganizationRequest) (*dto.OrganizationResponse, error) {
	name, slug, description, err := normalizeOrganizationInput(req.Name, req.Slug, req.Description)
	if err != nil {
		return nil, err
	}

	if err := s.ensureSlugAvailable(slug, uuid.Nil); err != nil {
		return nil, err
	}

	organization := &models.Organization{
		Name:        name,
		Slug:        slug,
		Description: description,
		OwnerID:     ownerID,
	}

	if err := s.repo.Create(organization); err != nil {
		return nil, err
	}

	response := mapOrganizationResponse(*organization)
	return &response, nil
}

func (s *OrganizationService) List(ownerID uint, req *models.PaginationRequest) (*dto.OrganizationListResponse, error) {
	organizations, total, err := s.repo.List(req, ownerID)
	if err != nil {
		return nil, err
	}

	items := make([]dto.OrganizationResponse, 0, len(organizations))
	for _, organization := range organizations {
		items = append(items, mapOrganizationResponse(organization))
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &dto.OrganizationListResponse{
		Items:      items,
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *OrganizationService) GetByID(id uuid.UUID, _ uint) (*dto.OrganizationResponse, error) {
	organization, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, err
	}

	response := mapOrganizationResponse(*organization)
	return &response, nil
}

func (s *OrganizationService) Update(id uuid.UUID, _ uint, req dto.UpdateOrganizationRequest) (*dto.OrganizationResponse, error) {
	name, slug, description, err := normalizeOrganizationInput(req.Name, req.Slug, req.Description)
	if err != nil {
		return nil, err
	}

	organization, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, err
	}

	if err := s.ensureSlugAvailable(slug, organization.ID); err != nil {
		return nil, err
	}

	organization.Name = name
	organization.Slug = slug
	organization.Description = description

	if err := s.repo.Update(organization); err != nil {
		return nil, err
	}

	response := mapOrganizationResponse(*organization)
	return &response, nil
}

func (s *OrganizationService) Delete(id uuid.UUID, _ uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrOrganizationNotFound
		}
		return err
	}

	return s.repo.Delete(id)
}

func (s *OrganizationService) ensureSlugAvailable(slug string, currentID uuid.UUID) error {
	organization, err := s.repo.GetBySlug(slug)
	if err == nil {
		if currentID == uuid.Nil || organization.ID != currentID {
			return apperrors.ErrOrganizationAlreadyExists
		}
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	return err
}

func normalizeOrganizationInput(name, slug, description string) (string, string, string, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if length := utf8.RuneCountInString(name); length < 3 || length > 100 {
		return "", "", "", apperrors.ErrInvalidOrganizationName
	}

	if utf8.RuneCountInString(description) > 500 {
		return "", "", "", apperrors.ErrInvalidOrganizationDescription
	}

	slug = strings.TrimSpace(slug)
	if slug == "" {
		slug = utils.GenerateSlug(name)
	}
	slug = utils.GenerateSlug(slug)
	if slug == "" {
		return "", "", "", apperrors.ErrInvalidOrganizationSlug
	}
	if length := utf8.RuneCountInString(slug); length < 3 || length > 120 {
		return "", "", "", apperrors.ErrInvalidOrganizationSlug
	}

	return name, slug, description, nil
}

func mapOrganizationResponse(organization models.Organization) dto.OrganizationResponse {
	return dto.OrganizationResponse{
		ID:          organization.ID.String(),
		Name:        organization.Name,
		Slug:        organization.Slug,
		Description: organization.Description,
		OwnerID:     organization.OwnerID,
		CreatedAt:   organization.CreatedAt,
		UpdatedAt:   organization.UpdatedAt,
	}
}
