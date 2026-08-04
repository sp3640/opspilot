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

		fetched, err := organizationService.GetByID(createdOrganizationID, owner.ID)
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
		updated, err := organizationService.Update(createdOrganizationID, owner.ID, dto.UpdateOrganizationRequest{
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
		if err := organizationService.Delete(createdOrganizationID, owner.ID); err != nil {
			t.Fatalf("delete organization: %v", err)
		}

		_, err := organizationService.GetByID(createdOrganizationID, owner.ID)
		if !errors.Is(err, apperrors.ErrOrganizationNotFound) {
			t.Fatalf("expected ErrOrganizationNotFound after delete, got %v", err)
		}
	})
}
