package database

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/sp3640/opspilot/backend/internal/models"
)

// userOrganizationConstraints are the two foreign keys that make User and
// Organization mutually dependent:
//
//	users.organization_id  -> organizations.id  (User.Organization)
//	organizations.owner_id -> users.id          (Organization.Owner)
//
// Both are real, permanent constraints. They are only *created* in a
// separate step from the tables, see MigrateUserOrganizationFoundation.
var userOrganizationConstraints = []struct {
	model        any
	relationship string
}{
	{model: &models.User{}, relationship: "Organization"},
	{model: &models.Organization{}, relationship: "Owner"},
}

// MigrateUserOrganizationFoundation creates/updates the users and
// organizations tables and both foreign keys between them.
//
// Why this is not a plain AutoMigrate(&User{}, &Organization{}): the two
// models reference each other, so no creation order satisfies both. GORM
// resolves a cycle by picking one table to create first and embedding its
// foreign key inline in the CREATE TABLE statement - on an empty PostgreSQL
// database that is `CREATE TABLE organizations (... FOREIGN KEY (owner_id)
// REFERENCES users ...)` before `users` exists, which fails with
// `relation "users" does not exist`. SQLite never enforces this, which is
// why SQLite-based tests could not catch it.
//
// The fix is to stage it, never to weaken it:
//  1. Migrate both tables with inline foreign-key creation disabled (on a
//     copy of the connection config - the shared DB handle is untouched).
//     Tables and columns only, so order no longer matters.
//  2. Add each foreign key explicitly once both tables exist, skipping any
//     that are already present.
//
// Exported so other code that builds a PostgreSQL schema itself (e.g. test
// setup) uses the same staged path instead of AutoMigrate-ing the cycle.
//
// Every step is additive and idempotent: an existing database has both
// tables and both constraints already, so nothing is created, altered or
// dropped and no rows are touched. Foreign-key integrity is never disabled
// globally and both constraints exist when this returns.
func MigrateUserOrganizationFoundation(db *gorm.DB) error {
	tablesOnly := db.Session(&gorm.Session{})
	tablesOnly.Config.DisableForeignKeyConstraintWhenMigrating = true
	if err := tablesOnly.AutoMigrate(&models.User{}, &models.Organization{}); err != nil {
		return fmt.Errorf("migrate users and organizations tables: %w", err)
	}

	migrator := db.Migrator()
	for _, constraint := range userOrganizationConstraints {
		if migrator.HasConstraint(constraint.model, constraint.relationship) {
			continue
		}
		if err := migrator.CreateConstraint(constraint.model, constraint.relationship); err != nil {
			return fmt.Errorf("create %T.%s foreign key: %w", constraint.model, constraint.relationship, err)
		}
	}

	return nil
}
