package integration

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/services"
)

func TestOrganizationFoundationIntegration(t *testing.T) {
	db := setupSQLiteIntegrationDB(t)

	userRepo := repository.NewUserRepository(db)
	organizationRepo := repository.NewOrganizationRepository(db)
	organizationService := services.NewOrganizationService(organizationRepo)

	owner := &models.User{
		Name:         "Org Owner",
		Email:        "org-owner@opspilot.dev",
		PasswordHash: "hashed-password",
		Role:         models.RoleUser,
	}
	if err := userRepo.Create(owner); err != nil {
		t.Fatalf("create owner user: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		created, err := organizationService.Create(owner.ID, dto.CreateOrganizationRequest{
			Name:        "Acme Corporation",
			Description: "Primary organization for integration tests",
		})
		if err != nil {
			t.Fatalf("create organization: %v", err)
		}

		if created.ID == "" {
			t.Fatalf("expected organization id")
		}
		if created.OwnerID != owner.ID {
			t.Fatalf("unexpected owner id: got %d want %d", created.OwnerID, owner.ID)
		}
		if created.Slug != "acme-corporation" {
			t.Fatalf("unexpected generated slug: %s", created.Slug)
		}
	})

	t.Run("Duplicate slug", func(t *testing.T) {
		_, err := organizationService.Create(owner.ID, dto.CreateOrganizationRequest{
			Name: "Acme Corporation Duplicate",
			Slug: "acme-corporation",
		})
		if !errors.Is(err, apperrors.ErrOrganizationAlreadyExists) {
			t.Fatalf("expected ErrOrganizationAlreadyExists, got %v", err)
		}
	})

	var createdOrganizationID uuid.UUID
	t.Run("Get", func(t *testing.T) {
		created, err := organizationService.Create(owner.ID, dto.CreateOrganizationRequest{
			Name: "Northwind",
			Slug: "northwind",
		})
		if err != nil {
			t.Fatalf("create organization for get: %v", err)
		}

		createdOrganizationID, err = uuid.Parse(created.ID)
		if err != nil {
			t.Fatalf("parse organization id: %v", err)
		}

		fetched, err := organizationService.GetByID(createdOrganizationID, createdOrganizationID)
		if err != nil {
			t.Fatalf("get organization: %v", err)
		}
		if fetched.Slug != "northwind" {
			t.Fatalf("unexpected slug: %s", fetched.Slug)
		}
	})

	t.Run("List", func(t *testing.T) {
		result, err := organizationService.List(owner.ID, &models.PaginationRequest{
			Page:  1,
			Limit: 10,
			Sort:  "created_at",
			Order: "desc",
		})
		if err != nil {
			t.Fatalf("list organizations: %v", err)
		}
		if len(result.Items) < 2 {
			t.Fatalf("expected at least 2 organizations, got %d", len(result.Items))
		}
		if result.Total < 2 {
			t.Fatalf("expected total >= 2, got %d", result.Total)
		}
	})

	t.Run("Update", func(t *testing.T) {
		updated, err := organizationService.Update(createdOrganizationID, createdOrganizationID, dto.UpdateOrganizationRequest{
			Name:        "Northwind Labs",
			Description: "Updated description",
		})
		if err != nil {
			t.Fatalf("update organization: %v", err)
		}

		if updated.Name != "Northwind Labs" {
			t.Fatalf("unexpected name: %s", updated.Name)
		}
		if updated.Slug != "northwind-labs" {
			t.Fatalf("expected slug regenerated from name, got %s", updated.Slug)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := organizationService.Delete(createdOrganizationID, createdOrganizationID); err != nil {
			t.Fatalf("delete organization: %v", err)
		}

		_, err := organizationService.GetByID(createdOrganizationID, createdOrganizationID)
		if !errors.Is(err, apperrors.ErrOrganizationNotFound) {
			t.Fatalf("expected ErrOrganizationNotFound after delete, got %v", err)
		}
	})

	t.Run("Cross-organization access is forbidden", func(t *testing.T) {
		orgA, err := organizationService.Create(owner.ID, dto.CreateOrganizationRequest{
			Name: "Org A Isolation",
			Slug: "org-a-isolation",
		})
		if err != nil {
			t.Fatalf("create org A: %v", err)
		}
		orgAID, err := uuid.Parse(orgA.ID)
		if err != nil {
			t.Fatalf("parse org A id: %v", err)
		}

		orgB, err := organizationService.Create(owner.ID, dto.CreateOrganizationRequest{
			Name: "Org B Isolation",
			Slug: "org-b-isolation",
		})
		if err != nil {
			t.Fatalf("create org B: %v", err)
		}
		orgBID, err := uuid.Parse(orgB.ID)
		if err != nil {
			t.Fatalf("parse org B id: %v", err)
		}

		// A caller whose own organization is A must not be able to read,
		// rename, or delete organization B merely by knowing its id.
		if _, err := organizationService.GetByID(orgBID, orgAID); !errors.Is(err, apperrors.ErrOrganizationForbidden) {
			t.Fatalf("expected ErrOrganizationForbidden reading org B as org A, got %v", err)
		}
		if _, err := organizationService.Update(orgBID, orgAID, dto.UpdateOrganizationRequest{
			Name: "Hijacked Name",
		}); !errors.Is(err, apperrors.ErrOrganizationForbidden) {
			t.Fatalf("expected ErrOrganizationForbidden updating org B as org A, got %v", err)
		}
		if err := organizationService.Delete(orgBID, orgAID); !errors.Is(err, apperrors.ErrOrganizationForbidden) {
			t.Fatalf("expected ErrOrganizationForbidden deleting org B as org A, got %v", err)
		}

		// Org B must be unaffected by the rejected attempts.
		stillThere, err := organizationService.GetByID(orgBID, orgBID)
		if err != nil {
			t.Fatalf("expected org B to still be readable by its own caller: %v", err)
		}
		if stillThere.Name != "Org B Isolation" {
			t.Fatalf("expected org B name unaffected by rejected cross-org update, got %q", stillThere.Name)
		}
	})
}
