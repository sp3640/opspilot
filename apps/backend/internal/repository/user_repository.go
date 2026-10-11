package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// WithTx returns a repository whose queries run inside tx.
func (r *UserRepository) WithTx(tx *gorm.DB) *UserRepository {
	return &UserRepository{db: tx}
}

// Transaction runs fn in a database transaction, committing when it returns
// nil and rolling back on any error.
func (r *UserRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User

	err := r.db.Where("email = ?", email).First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User

	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByIDWithOrganization(id uint) (*models.User, error) {
	var user models.User

	err := r.db.Preload("Organization").First(&user, id).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) CountUsers() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *UserRepository) AssignOrganization(userID uint, organizationID uuid.UUID) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("organization_id", organizationID).Error
}

func (r *UserRepository) AssignOrganizationAndRole(userID uint, organizationID uuid.UUID, role string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]any{
		"organization_id": organizationID,
		"role":            role,
	}).Error
}

// GetByEmailForUpdate loads the user and locks the row until the surrounding
// transaction ends, serializing concurrent verification/resend/register
// work on the same account.
func (r *UserRepository) GetByEmailForUpdate(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByIDForUpdate(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// MarkEmailVerified stamps the account verified.
func (r *UserRepository) MarkEmailVerified(userID uint, at time.Time) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("email_verified_at", at).Error
}

// UpdateUnverifiedCredentials replaces the name and password hash of an
// account that has never been verified. The email_verified_at IS NULL guard
// is part of the statement, so it can never touch a verified account.
func (r *UserRepository) UpdateUnverifiedCredentials(userID uint, name, passwordHash string) error {
	return r.db.Model(&models.User{}).
		Where("id = ? AND email_verified_at IS NULL", userID).
		Updates(map[string]any{"name": name, "password_hash": passwordHash}).Error
}
