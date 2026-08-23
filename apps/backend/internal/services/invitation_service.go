package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/rbac"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"gorm.io/gorm"
)

const defaultInvitationExpiry = 7 * 24 * time.Hour

type InvitationService struct {
	repo             *repository.InvitationRepository
	userRepo         *repository.UserRepository
	organizationRepo *repository.OrganizationRepository
	auditService     *AuditService
}

func NewInvitationService(
	repo *repository.InvitationRepository,
	userRepo *repository.UserRepository,
	organizationRepo *repository.OrganizationRepository,
) *InvitationService {
	return &InvitationService{repo: repo, userRepo: userRepo, organizationRepo: organizationRepo}
}

func (s *InvitationService) WithAuditService(auditService *AuditService) *InvitationService {
	s.auditService = auditService
	return s
}

func (s *InvitationService) InviteUser(invitedBy uint, inviterRole string, organizationID uuid.UUID, req dto.InviteRequest) (*dto.InvitationResponse, error) {
	if !isPlatformAdminRole(inviterRole) {
		return nil, apperrors.ErrInvitationForbidden
	}

	email, err := normalizeEmail(req.Email)
	if err != nil {
		return nil, err
	}

	role, err := normalizeInvitationRole(req.Role)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.ExpireOldInvitations(time.Now().UTC())
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByEmail(organizationID, email)
	if err == nil {
		if existing.Status == models.InvitationStatusPending && existing.ExpiresAt.After(time.Now().UTC()) {
			return nil, apperrors.ErrInvitationAlreadyExists
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	token, err := generateInvitationToken(32)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	invitation := &models.Invitation{
		OrganizationID: organizationID,
		Email:          email,
		Role:           role,
		Token:          token,
		Status:         models.InvitationStatusPending,
		InvitedBy:      invitedBy,
		ExpiresAt:      now.Add(defaultInvitationExpiry),
	}

	if err := s.repo.Create(invitation); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		// AfterState deliberately excludes invitation.Token - it is a
		// bearer credential for accepting the invite and must never appear
		// in an audit record.
		_ = s.auditService.LogEvent(AuditEventInput{
			UserID:         invitedBy,
			OrganizationID: organizationID,
			EntityType:     "invitation",
			EntityID:       invitation.ID.String(),
			Action:         models.AuditActionCreate,
			AfterState:     marshalAuditState(map[string]string{"email": invitation.Email, "role": invitation.Role}),
		})
	}

	response := mapInvitationResponse(*invitation)
	return &response, nil
}

func (s *InvitationService) AcceptInvitation(userID uint, userEmail string, req dto.AcceptInvitationRequest) (*dto.InvitationResponse, error) {
	invitation, err := s.resolveValidInvitation(strings.TrimSpace(req.Token), userEmail)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}

	if !strings.EqualFold(strings.TrimSpace(user.Email), invitation.Email) {
		return nil, apperrors.ErrInvitationEmailMismatch
	}

	if err := s.userRepo.AssignOrganizationAndRole(user.ID, invitation.OrganizationID, invitation.Role); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	previousStatus := invitation.Status
	invitation.Status = models.InvitationStatusAccepted
	invitation.AcceptedAt = &now
	if err := s.repo.Update(invitation); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		_ = s.auditService.LogUpdate(userID, invitation.OrganizationID, "invitation", invitation.ID.String(), nil, nil, "status", string(previousStatus), string(invitation.Status))
	}

	response := mapInvitationResponse(*invitation)
	return &response, nil
}

// ValidateInvitation resolves an invitation for display before acceptance. It
// performs the same checks as AcceptInvitation but never mutates state, so
// opening the acceptance page cannot consume an invitation.
func (s *InvitationService) ValidateInvitation(userEmail, token string) (*dto.ValidateInvitationResponse, error) {
	invitation, err := s.resolveValidInvitation(strings.TrimSpace(token), userEmail)
	if err != nil {
		return nil, err
	}

	organization, err := s.organizationRepo.GetByID(invitation.OrganizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, err
	}

	return &dto.ValidateInvitationResponse{
		Email:            invitation.Email,
		Role:             invitation.Role,
		OrganizationID:   invitation.OrganizationID.String(),
		OrganizationName: organization.Name,
		Status:           invitation.Status,
		ExpiresAt:        invitation.ExpiresAt,
	}, nil
}

// resolveValidInvitation loads the invitation for token and confirms it is
// still pending, unexpired, and addressed to userEmail. It is shared by
// AcceptInvitation and ValidateInvitation so both apply identical validity
// and authorization rules; only AcceptInvitation goes on to mutate state.
func (s *InvitationService) resolveValidInvitation(token, userEmail string) (*models.Invitation, error) {
	if token == "" {
		return nil, apperrors.ErrInvalidInvitationToken
	}

	normalizedUserEmail, err := normalizeEmail(userEmail)
	if err != nil {
		return nil, err
	}

	if _, err := s.repo.ExpireOldInvitations(time.Now().UTC()); err != nil {
		return nil, err
	}

	invitation, err := s.repo.GetByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrInvitationNotFound
		}
		return nil, err
	}

	if invitation.Status != models.InvitationStatusPending {
		if invitation.Status == models.InvitationStatusExpired {
			return nil, apperrors.ErrInvitationExpired
		}
		return nil, apperrors.ErrInvitationNotPending
	}

	now := time.Now().UTC()
	if invitation.ExpiresAt.Before(now) {
		invitation.Status = models.InvitationStatusExpired
		if updateErr := s.repo.Update(invitation); updateErr != nil {
			return nil, updateErr
		}
		return nil, apperrors.ErrInvitationExpired
	}

	if normalizedUserEmail != invitation.Email {
		return nil, apperrors.ErrInvitationEmailMismatch
	}

	return invitation, nil
}

func (s *InvitationService) RevokeInvitation(invitationID uuid.UUID, actorRole string, userID uint, organizationID uuid.UUID) error {
	if !isPlatformAdminRole(actorRole) {
		return apperrors.ErrInvitationForbidden
	}

	invitation, err := s.repo.GetByID(invitationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrInvitationNotFound
		}
		return err
	}

	if invitation.OrganizationID != organizationID {
		return apperrors.ErrInvitationForbidden
	}

	if invitation.Status != models.InvitationStatusPending {
		if invitation.Status == models.InvitationStatusExpired {
			return apperrors.ErrInvitationExpired
		}
		return apperrors.ErrInvitationNotPending
	}

	now := time.Now().UTC()
	if invitation.ExpiresAt.Before(now) {
		invitation.Status = models.InvitationStatusExpired
		if err := s.repo.Update(invitation); err != nil {
			return err
		}
		return apperrors.ErrInvitationExpired
	}

	previousStatus := invitation.Status
	invitation.Status = models.InvitationStatusRevoked
	if err := s.repo.Update(invitation); err != nil {
		return err
	}

	if s.auditService != nil {
		_ = s.auditService.LogUpdate(userID, organizationID, "invitation", invitation.ID.String(), nil, nil, "status", string(previousStatus), string(invitation.Status))
	}

	return nil
}

func (s *InvitationService) ListInvitations(actorRole string, organizationID uuid.UUID, req *models.PaginationRequest) (*dto.InvitationListResponse, error) {
	if !isPlatformAdminRole(actorRole) {
		return nil, apperrors.ErrInvitationForbidden
	}

	_, err := s.repo.ExpireOldInvitations(time.Now().UTC())
	if err != nil {
		return nil, err
	}

	items, total, err := s.repo.ListByOrganization(req, organizationID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.InvitationResponse, 0, len(items))
	for _, invitation := range items {
		responses = append(responses, mapInvitationResponse(invitation))
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	return &dto.InvitationListResponse{
		Items:      responses,
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func normalizeEmail(value string) (string, error) {
	email := strings.TrimSpace(strings.ToLower(value))
	if email == "" {
		return "", apperrors.ErrInvalidInvitationEmail
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "", apperrors.ErrInvalidInvitationEmail
	}

	return email, nil
}

// normalizeInvitationRole validates that role is one of the four RBAC
// roles. It never trusts the raw string from the frontend as-is: only an
// exact (case/whitespace-insensitive) match against rbac.AllRoles passes.
func normalizeInvitationRole(value string) (string, error) {
	role := strings.TrimSpace(value)
	if !rbac.IsValidRole(role) {
		return "", apperrors.ErrInvalidInvitationRole
	}

	return role, nil
}

func isPlatformAdminRole(role string) bool {
	return strings.EqualFold(strings.TrimSpace(role), models.RolePlatformAdmin)
}

func generateInvitationToken(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func mapInvitationResponse(invitation models.Invitation) dto.InvitationResponse {
	return dto.InvitationResponse{
		ID:             invitation.ID.String(),
		OrganizationID: invitation.OrganizationID.String(),
		Email:          invitation.Email,
		Role:           invitation.Role,
		Token:          invitation.Token,
		Status:         invitation.Status,
		InvitedBy:      invitation.InvitedBy,
		ExpiresAt:      invitation.ExpiresAt,
		AcceptedAt:     invitation.AcceptedAt,
		CreatedAt:      invitation.CreatedAt,
		UpdatedAt:      invitation.UpdatedAt,
	}
}
