package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sp3640/opspilot/backend/internal/auth"
	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	"github.com/sp3640/opspilot/backend/internal/middleware"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestUserOrganizationRelationshipIntegration verifies that every uninvited
// registrant — the very first user ever, and every one after them — creates
// their own new workspace and becomes its Platform Admin. There is no hidden
// "join an existing organization" fallback for uninvited signups.
func TestUserOrganizationRelationshipIntegration(t *testing.T) {
	db := setupSQLiteUserOrganizationDB(t)

	userRepo := repository.NewUserRepository(db)
	organizationRepo := repository.NewOrganizationRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	testConfig := &config.Config{JWTSecret: "this-is-a-very-long-test-jwt-secret-1234567890"}

	userService := services.NewUserService(userRepo, organizationRepo, invitationRepo, testConfig)
	userHandler := handlers.NewUserHandler(userService)

	if err := userService.Register("Alice", "alice@opspilot.dev", "password123", ""); err != nil {
		t.Fatalf("register first user: %v", err)
	}

	firstUser, err := userRepo.GetByEmail("alice@opspilot.dev")
	if err != nil {
		t.Fatalf("load first user: %v", err)
	}
	if firstUser.Role != models.RolePlatformAdmin {
		t.Fatalf("expected first user role %q, got %q", models.RolePlatformAdmin, firstUser.Role)
	}
	if firstUser.OrganizationID == nil {
		t.Fatalf("expected first user organization id to be assigned")
	}

	firstOrganization, err := organizationRepo.GetByID(*firstUser.OrganizationID)
	if err != nil {
		t.Fatalf("load first user's organization: %v", err)
	}
	if firstOrganization.OwnerID != firstUser.ID {
		t.Fatalf("expected organization owner %d, got %d", firstUser.ID, firstOrganization.OwnerID)
	}
	if firstOrganization.Name != "Alice's Workspace" {
		t.Fatalf("unexpected default organization name: %s", firstOrganization.Name)
	}

	// A second, later, entirely uninvited registrant must get their OWN new
	// workspace — not silently join Alice's — and may name it explicitly.
	if err := userService.Register("Bob", "bob@opspilot.dev", "password123", "Bob's Team"); err != nil {
		t.Fatalf("register second user: %v", err)
	}

	secondUser, err := userRepo.GetByEmail("bob@opspilot.dev")
	if err != nil {
		t.Fatalf("load second user: %v", err)
	}
	if secondUser.Role != models.RolePlatformAdmin {
		t.Fatalf("expected second user role %q, got %q", models.RolePlatformAdmin, secondUser.Role)
	}
	if secondUser.OrganizationID == nil {
		t.Fatalf("expected second user organization id to be assigned")
	}
	if *secondUser.OrganizationID == *firstUser.OrganizationID {
		t.Fatalf("expected second user to create their own organization, not join the first user's")
	}

	secondOrganization, err := organizationRepo.GetByID(*secondUser.OrganizationID)
	if err != nil {
		t.Fatalf("load second user's organization: %v", err)
	}
	if secondOrganization.OwnerID != secondUser.ID {
		t.Fatalf("expected second user to own their own new workspace")
	}
	if secondOrganization.Name != "Bob's Team" {
		t.Fatalf("expected the requested organization name to be honored, got %q", secondOrganization.Name)
	}

	organizationCount, err := organizationRepo.CountOrganizations()
	if err != nil {
		t.Fatalf("count organizations: %v", err)
	}
	if organizationCount != 2 {
		t.Fatalf("expected exactly 2 organizations, got %d", organizationCount)
	}

	token, err := userService.Login("bob@opspilot.dev", "password123")
	if err != nil {
		t.Fatalf("login second user: %v", err)
	}

	claims, err := auth.ValidateToken(token, testConfig.JWTSecret)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if claims.OrganizationID != secondUser.OrganizationID.String() {
		t.Fatalf("expected token organization id %q, got %q", secondUser.OrganizationID.String(), claims.OrganizationID)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	users := router.Group("/api/v1/users")
	users.Use(middleware.AuthMiddleware(testConfig))
	users.GET("/me", userHandler.Me)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected /users/me status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success bool                   `json:"success"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode /users/me payload: %v", err)
	}

	organizationID, _ := payload.Data["organizationId"].(string)
	organizationName, _ := payload.Data["organizationName"].(string)
	organizationSlug, _ := payload.Data["organizationSlug"].(string)
	if organizationID != secondUser.OrganizationID.String() {
		t.Fatalf("expected /users/me organizationId %q, got %q", secondUser.OrganizationID.String(), organizationID)
	}
	if organizationName != secondOrganization.Name {
		t.Fatalf("expected /users/me organizationName %q, got %q", secondOrganization.Name, organizationName)
	}
	if organizationSlug != secondOrganization.Slug {
		t.Fatalf("expected /users/me organizationSlug %q, got %q", secondOrganization.Slug, organizationSlug)
	}
}

func setupSQLiteUserOrganizationDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:user_org_relationship_%d?mode=memory&cache=private", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}

	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}

	migrateIntegrationSchema(t, db)

	return db
}
