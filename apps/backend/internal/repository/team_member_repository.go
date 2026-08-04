package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type TeamMemberRepository struct {
	db *gorm.DB
}

func NewTeamMemberRepository(db *gorm.DB) *TeamMemberRepository {
	return &TeamMemberRepository{db: db}
}

func (r *TeamMemberRepository) AddMember(member *models.TeamMember) error {
	if err := r.db.Create(member).Error; err != nil {
		if isDuplicateTeamMembershipError(err) {
			return apperrors.ErrTeamMemberAlreadyExists
		}
		return err
	}

	return nil
}

func (r *TeamMemberRepository) RemoveMember(teamID uuid.UUID, userID uint) error {
	return r.db.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&models.TeamMember{}).Error
}

func (r *TeamMemberRepository) ListMembers(teamID uuid.UUID) ([]models.TeamMember, error) {
	items := make([]models.TeamMember, 0)
	if err := r.db.Preload("User").Where("team_id = ?", teamID).Order("created_at ASC").Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (r *TeamMemberRepository) IsMember(teamID uuid.UUID, userID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&models.TeamMember{}).Where("team_id = ? AND user_id = ?", teamID, userID).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *TeamMemberRepository) CountMembers(teamID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&models.TeamMember{}).Where("team_id = ?", teamID).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func isDuplicateTeamMembershipError(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		postgresError.ConstraintName == "idx_team_members_team_user"
}
