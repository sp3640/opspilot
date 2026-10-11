package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/sp3640/opspilot/backend/internal/models"
)

type EmailVerificationTokenRepository struct {
	db *gorm.DB
}

func NewEmailVerificationTokenRepository(db *gorm.DB) *EmailVerificationTokenRepository {
	return &EmailVerificationTokenRepository{db: db}
}

// WithTx returns a repository whose queries run inside tx.
func (r *EmailVerificationTokenRepository) WithTx(tx *gorm.DB) *EmailVerificationTokenRepository {
	return &EmailVerificationTokenRepository{db: tx}
}

// ReplaceForUser makes token the user's only unused verification token:
// every previous unused token is deleted and token is inserted. Spent
// tokens are left alone (they only answer "already verified"). Callers run
// it inside the same transaction as the user-row lock so two concurrent
// resends cannot leave two live tokens.
func (r *EmailVerificationTokenRepository) ReplaceForUser(token *models.EmailVerificationToken) error {
	if err := r.db.Where("user_id = ? AND used_at IS NULL", token.UserID).
		Delete(&models.EmailVerificationToken{}).Error; err != nil {
		return err
	}
	return r.db.Omit(clause.Associations).Create(token).Error
}

// GetByHashForUpdate loads the token with the given hash and locks its row
// until the surrounding transaction ends, so two requests presenting the
// same link cannot both consume it.
func (r *EmailVerificationTokenRepository) GetByHashForUpdate(hash string) (*models.EmailVerificationToken, error) {
	var token models.EmailVerificationToken
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_hash = ?", hash).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// LatestCreatedAt returns when the user's most recent token (used or not)
// was issued, and false when they have never had one.
func (r *EmailVerificationTokenRepository) LatestCreatedAt(userID uint) (time.Time, bool, error) {
	var token models.EmailVerificationToken
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, err
	}
	return token.CreatedAt, true, nil
}

func (r *EmailVerificationTokenRepository) MarkUsed(id uint, at time.Time) error {
	return r.db.Model(&models.EmailVerificationToken{}).Where("id = ?", id).Update("used_at", at).Error
}

// DeleteUnusedForUser drops every unused token of the user.
func (r *EmailVerificationTokenRepository) DeleteUnusedForUser(userID uint) error {
	return r.db.Where("user_id = ? AND used_at IS NULL", userID).Delete(&models.EmailVerificationToken{}).Error
}
