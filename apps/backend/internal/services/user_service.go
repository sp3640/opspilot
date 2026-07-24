package services

import (
	"errors"
	"strings"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/auth"
	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	repo *repository.UserRepository
	cfg  *config.Config
}

func NewUserService(
	repo *repository.UserRepository,
	cfg *config.Config,
) *UserService {
	return &UserService{
		repo: repo,
		cfg:  cfg,
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

	return s.repo.Create(user)
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
		s.cfg.JWTSecret,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}