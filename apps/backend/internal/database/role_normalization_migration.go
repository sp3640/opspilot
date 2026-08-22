package database

import (
	"fmt"

	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

// runRoleNormalizationMigration rewrites the legacy "User" role value to the
// new default role, "Viewer", introduced by the four-role RBAC model.
// Platform Admin rows are never touched. The statement is idempotent: once
// no row has role = 'User' it is a no-op on every subsequent startup, so it
// is safe to run unconditionally.
func runRoleNormalizationMigration(db *gorm.DB) error {
	if !db.Migrator().HasTable("users") {
		return nil
	}
	if !db.Migrator().HasColumn("users", "role") {
		return nil
	}

	if err := db.Exec(
		`UPDATE users SET role = ? WHERE role = ?`,
		models.RoleViewer, models.RoleUser,
	).Error; err != nil {
		return fmt.Errorf("normalize legacy user roles: %w", err)
	}

	return nil
}
