package repository

import (
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
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
