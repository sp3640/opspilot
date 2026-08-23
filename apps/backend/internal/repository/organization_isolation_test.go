package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestOrganizationIsolation is a direct, isolated regression test for the
// single most security-critical bug class in this multi-tenant system: an
// org-scoped repository method returning a row that belongs to a different
// organization than the one it was asked to filter by. Every method below
// takes both an id and an organizationID and must return gorm.ErrRecordNotFound
// (not the row) when the id exists but belongs to a different organization.
//
// This is deliberately independent of the HTTP-level integration suite
// (internal/integration/*_test.go), which already covers cross-tenant
// isolation broadly but only exercises whichever repository method happens
// to sit behind the routes it drives. A regression in the WHERE clause of
// one of these methods - the exact place the bug would actually live -
// would not necessarily be caught there if no integration test happens to
// route through that specific method with cross-org fixtures.
func TestOrganizationIsolation(t *testing.T) {
	db := setupOrganizationIsolationDB(t)

	user := mustCreateIsolationUser(t, db, "isolation-user@opspilot.dev")
	orgA := mustCreateIsolationOrganization(t, db, "Org A", user.ID)
	orgB := mustCreateIsolationOrganization(t, db, "Org B", user.ID)

	t.Run("ProjectRepository.GetByIDAndOrganizationID", func(t *testing.T) {
		repo := NewProjectRepository(db)
		project := mustCreateIsolationProject(t, db, orgA, user.ID)

		if _, err := repo.GetByIDAndOrganizationID(project.ID, orgA); err != nil {
			t.Fatalf("expected the owning organization to read its own project, got %v", err)
		}

		_, err := repo.GetByIDAndOrganizationID(project.ID, orgB)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected ErrRecordNotFound reading org A's project as org B, got %v", err)
		}
	})

	t.Run("ApplicationRepository.GetApplication", func(t *testing.T) {
		repo := NewApplicationRepository(db)
		project := mustCreateIsolationProject(t, db, orgA, user.ID)
		application := mustCreateIsolationApplication(t, db, orgA, project.ID)

		if _, err := repo.GetApplication(application.ID, orgA); err != nil {
			t.Fatalf("expected the owning organization to read its own application, got %v", err)
		}

		_, err := repo.GetApplication(application.ID, orgB)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected ErrRecordNotFound reading org A's application as org B, got %v", err)
		}
	})

	t.Run("ClusterRepository.FindByID", func(t *testing.T) {
		repo := NewClusterRepository(db)
		project := mustCreateIsolationProject(t, db, orgA, user.ID)
		cluster := mustCreateIsolationCluster(t, db, orgA, project.ID, user.ID)

		if _, err := repo.FindByID(cluster.ID, orgA); err != nil {
			t.Fatalf("expected the owning organization to read its own cluster, got %v", err)
		}

		_, err := repo.FindByID(cluster.ID, orgB)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected ErrRecordNotFound reading org A's cluster as org B, got %v", err)
		}
	})

	t.Run("IncidentRepository.GetByIDAndOrganizationID", func(t *testing.T) {
		repo := NewIncidentRepository(db)
		project := mustCreateIsolationProject(t, db, orgA, user.ID)
		incident := mustCreateIsolationIncident(t, db, orgA, project.ID, user.ID)

		if _, err := repo.GetByIDAndOrganizationID(incident.ID, orgA); err != nil {
			t.Fatalf("expected the owning organization to read its own incident, got %v", err)
		}

		_, err := repo.GetByIDAndOrganizationID(incident.ID, orgB)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected ErrRecordNotFound reading org A's incident as org B, got %v", err)
		}
	})

	t.Run("AuditRepository.ListByProjectID", func(t *testing.T) {
		repo := NewAuditRepository(db)
		project := mustCreateIsolationProject(t, db, orgA, user.ID)
		mustCreateIsolationAuditLog(t, db, orgA, project.ID, user.ID)

		req := &models.PaginationRequest{Page: 1, Limit: 20}

		itemsA, totalA, err := repo.ListByProjectID(req, project.ID, orgA)
		if err != nil {
			t.Fatalf("list as owning organization: %v", err)
		}
		if totalA != 1 || len(itemsA) != 1 {
			t.Fatalf("expected the owning organization to see its own audit log, got total=%d items=%d", totalA, len(itemsA))
		}

		itemsB, totalB, err := repo.ListByProjectID(req, project.ID, orgB)
		if err != nil {
			t.Fatalf("list as other organization: %v", err)
		}
		if totalB != 0 || len(itemsB) != 0 {
			t.Fatalf("expected org B to see none of org A's audit logs for org A's project, got total=%d items=%d", totalB, len(itemsB))
		}
	})
}

func setupOrganizationIsolationDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:organization_isolation_%s?mode=memory&cache=private", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}

	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.Project{},
		&models.Application{},
		&models.Cluster{},
		&models.Incident{},
		&models.AuditLog{},
	); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	return db
}

func mustCreateIsolationOrganization(t *testing.T, db *gorm.DB, name string, ownerID uint) uuid.UUID {
	t.Helper()

	organization := &models.Organization{Name: name, Slug: name + "-" + uuid.NewString(), OwnerID: ownerID}
	if err := db.Create(organization).Error; err != nil {
		t.Fatalf("create organization %q: %v", name, err)
	}

	return organization.ID
}

func mustCreateIsolationUser(t *testing.T, db *gorm.DB, email string) *models.User {
	t.Helper()

	user := &models.User{Name: "Isolation Test User", Email: email, PasswordHash: "not-a-real-hash", Role: models.RolePlatformAdmin}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user %q: %v", email, err)
	}

	return user
}

func mustCreateIsolationProject(t *testing.T, db *gorm.DB, organizationID uuid.UUID, ownerID uint) *models.Project {
	t.Helper()

	project := &models.Project{
		OrganizationID: organizationID,
		Name:           "Isolation Test Project",
		Slug:           "isolation-test-project-" + uuid.NewString(),
		OwnerID:        ownerID,
	}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}

	return project
}

func mustCreateIsolationApplication(t *testing.T, db *gorm.DB, organizationID, projectID uuid.UUID) *models.Application {
	t.Helper()

	application := &models.Application{
		OrganizationID: organizationID,
		ProjectID:      projectID,
		Name:           "Isolation Test Application",
		Slug:           "isolation-test-application-" + uuid.NewString(),
		Runtime:        "node",
		Port:           8080,
	}
	if err := db.Create(application).Error; err != nil {
		t.Fatalf("create application: %v", err)
	}

	return application
}

func mustCreateIsolationCluster(t *testing.T, db *gorm.DB, organizationID, projectID uuid.UUID, createdBy uint) *models.Cluster {
	t.Helper()

	cluster := &models.Cluster{
		OrganizationID: organizationID,
		ProjectID:      projectID,
		Name:           "Isolation Test Cluster",
		Provider:       "generic",
		Status:         "PENDING",
		CreatedBy:      createdBy,
		Metadata:       json.RawMessage("{}"),
	}
	if err := db.Create(cluster).Error; err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	return cluster
}

func mustCreateIsolationIncident(t *testing.T, db *gorm.DB, organizationID, projectID uuid.UUID, userID uint) *models.Incident {
	t.Helper()

	incident := &models.Incident{
		OrganizationID: organizationID,
		ProjectID:      projectID,
		Title:          "Isolation Test Incident",
		Severity:       "P2",
		Status:         "OPEN",
		UserID:         userID,
	}
	if err := db.Create(incident).Error; err != nil {
		t.Fatalf("create incident: %v", err)
	}

	return incident
}

func mustCreateIsolationAuditLog(t *testing.T, db *gorm.DB, organizationID, projectID uuid.UUID, userID uint) *models.AuditLog {
	t.Helper()

	log := &models.AuditLog{
		OrganizationID: organizationID,
		ProjectID:      &projectID,
		UserID:         userID,
		EntityType:     "project",
		EntityID:       projectID.String(),
		Action:         models.AuditActionCreate,
		Result:         models.AuditResultSuccess,
	}
	if err := db.Create(log).Error; err != nil {
		t.Fatalf("create audit log: %v", err)
	}

	return log
}
