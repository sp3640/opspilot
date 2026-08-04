package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/auth"
	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	repo             *repository.UserRepository
	organizationRepo *repository.OrganizationRepository
	cfg              *config.Config
}

func NewUserService(
	repo *repository.UserRepository,
	organizationRepo *repository.OrganizationRepository,
	cfg *config.Config,
) *UserService {
	return &UserService{
		repo:             repo,
		organizationRepo: organizationRepo,
		cfg:              cfg,
	}
}

// Register creates a new user
func (s *UserService) Register(name, email, password string) error {

	// Normalize user input
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))

	_, err := s.repo.GetByEmail(email)

	if err == nil {
		return apperrors.ErrEmailAlreadyExists
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}

	userCount, err := s.repo.CountUsers()
	if err != nil {
		return err
	}

	if userCount == 0 {
		user.Role = models.RolePlatformAdmin
	} else {
		user.Role = models.RoleUser
	}

	if err := s.repo.Create(user); err != nil {
		return err
	}

	if userCount == 0 {
		organization := &models.Organization{
			Name:        fmt.Sprintf("%s's Workspace", user.Name),
			Slug:        utils.GenerateSlug(fmt.Sprintf("%s's Workspace", user.Name)),
			Description: "",
			OwnerID:     user.ID,
		}

		if err := s.organizationRepo.Create(organization); err != nil {
			return err
		}

		return s.repo.AssignOrganization(user.ID, organization.ID)
	}

	organization, err := s.organizationRepo.GetFirst()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrOrganizationNotFound
		}
		return err
	}

	return s.repo.AssignOrganization(user.ID, organization.ID)
}

// Login authenticates a user and returns a JWT token
func (s *UserService) Login(email, password string) (string, error) {

	// Normalize email before searching
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return "", apperrors.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		return "", apperrors.ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(
		user.ID,
		user.Email,
		resolveUserRole(user.Role),
		resolveOrganizationID(user.OrganizationID),
		s.cfg.JWTSecret,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}
func (s *UserService) GetCurrentUser(id uint) (*models.User, error) {

	user, err := s.repo.GetByIDWithOrganization(id)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	return user, nil
}

func resolveUserRole(role string) string {
	if strings.TrimSpace(role) == "" {
		return models.RoleUser
	}

	return role
}

func resolveOrganizationID(organizationID *uuid.UUID) string {
	if organizationID == nil {
		return ""
	}

	return organizationID.String()
}
