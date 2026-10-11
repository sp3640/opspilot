package database

import (
	"fmt"

	"gorm.io/gorm"
)

const (
	usersTable               = "users"
	emailVerifiedAtColumn    = "email_verified_at"
	emailVerificationLockKey = 7301401 // arbitrary, fixed: serializes concurrent starters
)

// runEmailVerificationMigration introduces users.email_verified_at without
// locking anyone out.
//
// Login now requires a verified email. Every account that already exists
// when this feature first reaches a database was created without any
// verification step, so it is grandfathered: the column is added and every
// existing row is stamped as verified, in ONE transaction (so a crash can
// never leave the column present but unstamped, which would lock all
// existing users out).
//
// "First time" is defined precisely as "the users table exists but has no
// email_verified_at column". That is why this must run before anything that
// AutoMigrates the User model (which would add the column empty and make
// the check below skip the stamp), and why it never re-stamps later: once
// the column exists this is a no-op, so accounts registered afterwards stay
// unverified until their owner follows the emailed link.
//
// A fresh database has no users table yet and is skipped; AutoMigrate then
// creates the column empty and every account there registers normally.
//
// Concurrent instances starting on the same legacy database are serialized
// by an advisory lock; the loser re-checks inside the lock, finds the
// column, and does nothing - so it can never stamp a user that registered
// after the winner finished.
func runEmailVerificationMigration(db *gorm.DB) error {
	if db.Dialector.Name() != postgresDialectName {
		return nil
	}
	if !db.Migrator().HasTable(usersTable) || db.Migrator().HasColumn(usersTable, emailVerifiedAtColumn) {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`SELECT pg_advisory_xact_lock(?)`, emailVerificationLockKey).Error; err != nil {
			return fmt.Errorf("lock email verification migration: %w", err)
		}
		if tx.Migrator().HasColumn(usersTable, emailVerifiedAtColumn) {
			return nil
		}

		if err := tx.Exec(`ALTER TABLE users ADD COLUMN email_verified_at timestamptz`).Error; err != nil {
			return fmt.Errorf("add users.email_verified_at: %w", err)
		}
		if err := tx.Exec(`UPDATE users SET email_verified_at = now() WHERE email_verified_at IS NULL`).Error; err != nil {
			return fmt.Errorf("grandfather existing users as verified: %w", err)
		}
		return nil
	})
}
