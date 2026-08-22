package database

import (
	"testing"

	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupRoleNormalizationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("migrate users table: %v", err)
	}

	return db
}

func TestRoleNormalizationMigrationRewritesLegacyUserRole(t *testing.T) {
	db := setupRoleNormalizationTestDB(t)

	legacy := &models.User{Name: "Legacy", Email: "legacy@opspilot.dev", PasswordHash: "hash", Role: models.RoleUser}
	if err := db.Create(legacy).Error; err != nil {
		t.Fatalf("seed legacy user: %v", err)
	}

	if err := runRoleNormalizationMigration(db); err != nil {
		t.Fatalf("run role normalization migration: %v", err)
	}

	var migrated models.User
	if err := db.First(&migrated, legacy.ID).Error; err != nil {
		t.Fatalf("reload legacy user: %v", err)
	}
	if migrated.Role != models.RoleViewer {
		t.Fatalf("expected legacy 'User' role normalized to %q, got %q", models.RoleViewer, migrated.Role)
	}
}

func TestRoleNormalizationMigrationLeavesOtherRolesUntouched(t *testing.T) {
	db := setupRoleNormalizationTestDB(t)

	admin := &models.User{Name: "Admin", Email: "admin@opspilot.dev", PasswordHash: "hash", Role: models.RolePlatformAdmin}
	if err := db.Create(admin).Error; err != nil {
		t.Fatalf("seed admin user: %v", err)
	}
	devops := &models.User{Name: "DevOps", Email: "devops@opspilot.dev", PasswordHash: "hash", Role: models.RoleDevOpsEngineer}
	if err := db.Create(devops).Error; err != nil {
		t.Fatalf("seed devops user: %v", err)
	}

	if err := runRoleNormalizationMigration(db); err != nil {
		t.Fatalf("run role normalization migration: %v", err)
	}

	var reloadedAdmin models.User
	if err := db.First(&reloadedAdmin, admin.ID).Error; err != nil {
		t.Fatalf("reload admin user: %v", err)
	}
	if reloadedAdmin.Role != models.RolePlatformAdmin {
		t.Fatalf("expected Platform Admin role untouched, got %q", reloadedAdmin.Role)
	}

	var reloadedDevOps models.User
	if err := db.First(&reloadedDevOps, devops.ID).Error; err != nil {
		t.Fatalf("reload devops user: %v", err)
	}
	if reloadedDevOps.Role != models.RoleDevOpsEngineer {
		t.Fatalf("expected DevOps Engineer role untouched, got %q", reloadedDevOps.Role)
	}
}

func TestRoleNormalizationMigrationIsIdempotent(t *testing.T) {
	db := setupRoleNormalizationTestDB(t)

	legacy := &models.User{Name: "Legacy", Email: "legacy@opspilot.dev", PasswordHash: "hash", Role: models.RoleUser}
	if err := db.Create(legacy).Error; err != nil {
		t.Fatalf("seed legacy user: %v", err)
	}

	for i := 0; i < 2; i++ {
		if err := runRoleNormalizationMigration(db); err != nil {
			t.Fatalf("run role normalization migration (pass %d): %v", i, err)
		}
	}

	var migrated models.User
	if err := db.First(&migrated, legacy.ID).Error; err != nil {
		t.Fatalf("reload legacy user: %v", err)
	}
	if migrated.Role != models.RoleViewer {
		t.Fatalf("expected role %q after repeated runs, got %q", models.RoleViewer, migrated.Role)
	}
}
