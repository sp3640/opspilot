package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	// migrationLockKey identifies the startup-migration advisory lock. It is
	// an arbitrary constant shared by every OpsPilot instance; advisory locks
	// are scoped to one database, so it never collides with another
	// database's (or another application's, unless it picked the same
	// number) locks. runEmailVerificationMigration keeps its own key
	// (emailVerificationLockKey) so it is also safe when called on its own.
	migrationLockKey int64 = 7301400
)

// migrationLockTimeout bounds how long an instance waits for another
// instance to finish migrating. It is a variable only so tests can shorten
// it. Waiting forever behind a wedged migrator would hang every replica
// silently; a bounded wait fails startup with a clear error instead, which
// the orchestrator answers by restarting the instance - by which time the
// original holder has normally finished (or its dead connection has
// released the lock).
var migrationLockTimeout = 5 * time.Minute

// runStartupMigrations brings the database schema and data up to date, and
// is safe to run from any number of instances at the same time.
//
// Why a lock is needed: schema migration is check-then-act ("does this table
// exist? no -> create it"), which two instances can interleave. On an empty
// database that produced duplicate-object errors while creating the first
// tables; on an existing one, concurrent ALTER/UPDATE statements deadlocked.
// Nothing in the individual steps is wrong on its own - they are only
// unsafe to overlap.
//
// How it is made safe: the WHOLE sequence runs inside one transaction whose
// first statement takes a PostgreSQL transaction-level advisory lock.
//   - Exactly one instance runs the sequence; the others queue on the lock
//     inside the database, so this works across processes, containers and
//     hosts with no coordination service and no dependence on start order.
//   - Every step runs on the same transaction, hence the same connection
//     that holds the lock, so the lock cannot be held on one connection
//     while the work happens on another.
//   - The lock is released by COMMIT or ROLLBACK - and by the connection
//     dying - so a crashed instance can never leave it held.
//   - A transaction-level lock (unlike a session-level one) also works
//     behind a transaction-pooling proxy such as PgBouncer.
//   - PostgreSQL DDL is transactional, so first-time initialization is
//     all-or-nothing: a failure leaves an empty database, not a half-built
//     schema that the next instance would have to untangle.
//
// Once the winner commits, each waiting instance acquires the lock in turn,
// finds everything already migrated, and every step is a no-op for it (all
// steps are additive and idempotent).
//
// The critical section is exactly the migrations - opening the connection
// pool beforehand and starting the application afterwards happen outside
// it, so the lock is never held during unrelated startup work.
func runStartupMigrations(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := acquireMigrationLock(tx); err != nil {
			return err
		}

		// Order matters; see the comment on each step.
		steps := []struct {
			name string
			run  func(*gorm.DB) error
		}{
			// Must stay first: it detects "users table exists but predates
			// email verification" by the absence of users.email_verified_at,
			// and any later step that migrates the User model would add that
			// column and hide the fact that existing accounts still need
			// grandfathering.
			{"run email verification migration", runEmailVerificationMigration},
			// Creates users/organizations with their mutual foreign keys
			// staged after both tables exist (see
			// MigrateUserOrganizationFoundation), then backfills
			// organizations for pre-existing data.
			{"run organization backfill migration", runOrganizationBackfillMigration},
			{"run role normalization migration", runRoleNormalizationMigration},
			{"run audit entity id migration", runAuditEntityIDMigration},
			// Everything else, including creating every remaining table on
			// first start.
			{"migrate database models", func(tx *gorm.DB) error { return tx.AutoMigrate(migrationModels()...) }},
		}
		for _, step := range steps {
			if err := step.run(tx); err != nil {
				return fmt.Errorf("%s: %w", step.name, err)
			}
		}
		return nil
	})
}

// acquireMigrationLock takes the transaction-level advisory lock, waiting at
// most migrationLockTimeout. It must be called first on the transaction that
// then performs the migrations.
//
// lock_timeout is what bounds an advisory lock wait. It is applied only
// around the acquisition (SET LOCAL, so it dies with the transaction) and
// then reset, so the migration statements themselves behave exactly as they
// did before the lock existed.
func acquireMigrationLock(tx *gorm.DB) error {
	if err := tx.Exec(fmt.Sprintf(`SET LOCAL lock_timeout = %d`, migrationLockTimeout.Milliseconds())).Error; err != nil {
		return fmt.Errorf("set migration lock timeout: %w", err)
	}
	if err := tx.Exec(`SELECT pg_advisory_xact_lock(?)`, migrationLockKey).Error; err != nil {
		return fmt.Errorf("acquire migration lock (waited up to %s for another instance to finish migrating): %w", migrationLockTimeout, err)
	}
	if err := tx.Exec(`RESET lock_timeout`).Error; err != nil {
		return fmt.Errorf("reset lock timeout: %w", err)
	}
	return nil
}
