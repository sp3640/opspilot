package database

import (
	"sync"
	"testing"
)

// Email verification adds users.email_verified_at and the
// email_verification_tokens table. The behaviour that matters most is what
// happens to accounts that already exist: login now requires a verified
// email, so a migration that left them unverified would lock every existing
// user out of a deployed system. These tests run the real startup
// migration (Connect) against genuine PostgreSQL and are skipped, not
// faked, when OPSPILOT_TEST_POSTGRES_DSN is unset (see
// postgres_migration_test.go).

const legacyUsersWithoutVerification = `CREATE TABLE users (
	id bigserial PRIMARY KEY,
	name varchar(100) NOT NULL,
	email varchar(255) NOT NULL,
	password_hash text NOT NULL,
	role varchar(50) NOT NULL DEFAULT 'Viewer',
	organization_id uuid,
	created_at timestamptz,
	updated_at timestamptz)`

func (s *postgresSchema) seedLegacyUsers(t *testing.T) {
	t.Helper()
	for _, statement := range []string{
		legacyUsersWithoutVerification,
		`CREATE UNIQUE INDEX idx_users_email ON users (email)`,
		`INSERT INTO users (name, email, password_hash, role, created_at, updated_at) VALUES
			('First', 'first@example.com', 'x', 'Viewer', now(), now()),
			('Second', 'second@example.com', 'x', 'Platform Admin', now(), now()),
			('Third', 'third@example.com', 'x', 'Developer', now(), now())`,
	} {
		if err := s.db.Exec(statement).Error; err != nil {
			t.Fatalf("seed legacy users: %v", err)
		}
	}
}

func TestConnectCreatesEmailVerificationSchemaOnEmptyPostgres(t *testing.T) {
	pg := newEmptyPostgresSchema(t)
	pg.connect(t)

	if pg.count(t, `SELECT count(*) FROM information_schema.columns WHERE table_schema = ? AND table_name = 'users'
		AND column_name = 'email_verified_at' AND is_nullable = 'YES' AND data_type = 'timestamp with time zone'`, pg.name) != 1 {
		t.Fatalf("users.email_verified_at is missing or is not a nullable timestamptz")
	}
	if pg.count(t, `SELECT count(*) FROM pg_indexes WHERE schemaname = ? AND tablename = 'email_verification_tokens'
		AND indexdef ILIKE '%UNIQUE%' AND indexdef ILIKE '%(token_hash)%'`, pg.name) != 1 {
		t.Fatalf("token_hash must carry a unique index")
	}
	if pg.count(t, `SELECT count(*) FROM pg_constraint c JOIN pg_namespace n ON n.oid = c.connamespace
		WHERE c.contype = 'f' AND n.nspname = ? AND c.conrelid::regclass::text = 'email_verification_tokens'
		AND c.confrelid::regclass::text = 'users' AND c.confdeltype = 'c'`, pg.name) != 1 {
		t.Fatalf("email_verification_tokens.user_id must reference users with ON DELETE CASCADE")
	}
	pg.assertUserOrganizationForeignKeys(t)

	// New accounts start unverified, and stay that way across restarts.
	// (The first user becomes owner of the default organization created by
	// the startup backfill, so the cascade check below uses the second.)
	if err := pg.db.Exec(`INSERT INTO users (name, email, password_hash, role, created_at, updated_at) VALUES
		('Owner', 'owner@example.com', 'x', 'Viewer', now(), now()),
		('New', 'new@example.com', 'x', 'Viewer', now(), now())`).Error; err != nil {
		t.Fatalf("insert users: %v", err)
	}
	pg.connect(t)
	pg.connect(t)
	if pg.count(t, `SELECT count(*) FROM users WHERE email_verified_at IS NULL`) != 2 {
		t.Fatalf("a restart verified an account nobody verified")
	}

	// Deleting a user takes its tokens with it instead of being blocked.
	if err := pg.db.Exec(`INSERT INTO email_verification_tokens (user_id, token_hash, expires_at, created_at)
		SELECT id, repeat('a', 64), now() + interval '1 day', now() FROM users WHERE email = 'new@example.com'`).Error; err != nil {
		t.Fatalf("insert token: %v", err)
	}
	if err := pg.db.Exec(`DELETE FROM users WHERE email = 'new@example.com'`).Error; err != nil {
		t.Fatalf("delete user with a token: %v", err)
	}
	if pg.count(t, `SELECT count(*) FROM email_verification_tokens`) != 0 {
		t.Fatalf("token survived its user")
	}
}

// TestConnectGrandfathersExistingUsersAsVerified is the compatibility
// policy: every account that exists when the feature arrives is stamped
// verified, exactly once, so nobody is locked out - while accounts created
// afterwards are not stamped by later restarts.
func TestConnectGrandfathersExistingUsersAsVerified(t *testing.T) {
	pg := newEmptyPostgresSchema(t)
	pg.seedLegacyUsers(t)

	pg.connect(t)

	if got := pg.count(t, `SELECT count(*) FROM users`); got != 3 {
		t.Fatalf("legacy users lost or duplicated: have %d, want 3", got)
	}
	if got := pg.count(t, `SELECT count(*) FROM users WHERE email_verified_at IS NOT NULL`); got != 3 {
		t.Fatalf("%d of 3 existing users were left unverified and would be locked out", 3-got)
	}
	stamps := pg.count(t, `SELECT COALESCE(sum(extract(epoch FROM email_verified_at))::bigint, 0) FROM users`)

	// An account registered after the migration is unverified ...
	if err := pg.db.Exec(`INSERT INTO users (name, email, password_hash, role, organization_id, created_at, updated_at)
		VALUES ('Fresh', 'fresh@example.com', 'x', 'Viewer', (SELECT id FROM organizations LIMIT 1), now(), now())`).Error; err != nil {
		t.Fatalf("insert post-migration user: %v", err)
	}

	// ... and later startups must neither verify it nor restamp anyone.
	pg.connect(t)
	pg.connect(t)

	if pg.count(t, `SELECT count(*) FROM users WHERE email = 'fresh@example.com' AND email_verified_at IS NULL`) != 1 {
		t.Fatalf("a later startup verified an unverified account (the grandfathering ran again)")
	}
	if got := pg.count(t, `SELECT COALESCE(sum(extract(epoch FROM email_verified_at))::bigint, 0) FROM users`); got != stamps {
		t.Fatalf("existing users' verification timestamps changed on restart")
	}
}

// TestConnectGrandfathersUsersInAnInitializedDatabase covers a deployed,
// fully migrated database from before the feature (all tables present,
// real data), reproduced by removing exactly what the feature adds.
func TestConnectGrandfathersUsersInAnInitializedDatabase(t *testing.T) {
	pg := newEmptyPostgresSchema(t)
	pg.connect(t)

	mustExec := func(query string, args ...any) {
		t.Helper()
		if err := pg.db.Exec(query, args...).Error; err != nil {
			t.Fatalf("%s: %v", query, err)
		}
	}
	const orgID = "6f1c5f3e-8b0a-4a44-9a55-0d2f5d9c1a11"
	mustExec(`INSERT INTO users (id, name, email, password_hash, role, created_at, updated_at) VALUES (1, 'Owner', 'owner@example.com', 'x', 'Platform Admin', now(), now()), (2, 'Member', 'member@example.com', 'x', 'Developer', now(), now())`)
	mustExec(`INSERT INTO organizations (id, name, slug, owner_id, created_at, updated_at) VALUES (?, 'Acme', 'acme', 1, now(), now())`, orgID)
	mustExec(`UPDATE users SET organization_id = ?`, orgID)

	// Roll the schema back to how a pre-feature deployment looks.
	mustExec(`DROP TABLE email_verification_tokens`)
	mustExec(`ALTER TABLE users DROP COLUMN email_verified_at`)
	dataBefore := pg.count(t, `SELECT (SELECT count(*) FROM users) * 1000 + (SELECT count(*) FROM organizations) * 100`)

	pg.connect(t)

	if got := pg.count(t, `SELECT count(*) FROM users WHERE email_verified_at IS NOT NULL`); got != 2 {
		t.Fatalf("%d of 2 existing users were left unverified", 2-got)
	}
	if got := pg.count(t, `SELECT (SELECT count(*) FROM users) * 1000 + (SELECT count(*) FROM organizations) * 100`); got != dataBefore {
		t.Fatalf("existing data changed by the migration")
	}
	if pg.count(t, `SELECT count(*) FROM users WHERE email = 'owner@example.com' AND role = 'Platform Admin' AND organization_id = ?`, orgID) != 1 {
		t.Fatalf("existing user was altered beyond the new column")
	}
	if !pg.db.Migrator().HasTable("email_verification_tokens") {
		t.Fatalf("email_verification_tokens was not created")
	}
}

// TestEmailVerificationMigrationIsSafeUnderConcurrentStartup runs the
// migration from several instances at once on the same legacy database (a
// rolling deploy): all must succeed, every legacy user must end up verified,
// and no instance may stamp an account registered after another finished.
func TestEmailVerificationMigrationIsSafeUnderConcurrentStartup(t *testing.T) {
	pg := newEmptyPostgresSchema(t)
	pg.seedLegacyUsers(t)

	const instances = 6
	errs := make(chan error, instances)
	var wg sync.WaitGroup
	for i := 0; i < instances; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- runEmailVerificationMigration(pg.db)
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent migration failed: %v", err)
		}
	}

	if got := pg.count(t, `SELECT count(*) FROM users WHERE email_verified_at IS NOT NULL`); got != 3 {
		t.Fatalf("%d of 3 legacy users verified after concurrent startup", got)
	}
}

func TestRunEmailVerificationMigrationIsIdempotent(t *testing.T) {
	pg := newEmptyPostgresSchema(t)
	pg.seedLegacyUsers(t)

	for run := 1; run <= 3; run++ {
		if err := runEmailVerificationMigration(pg.db); err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
	}
	if got := pg.count(t, `SELECT count(*) FROM users WHERE email_verified_at IS NOT NULL`); got != 3 {
		t.Fatalf("expected the 3 existing users stamped, have %d", got)
	}
	if err := pg.db.Exec(`INSERT INTO users (name, email, password_hash, role, created_at, updated_at) VALUES ('Later', 'later@example.com', 'x', 'Viewer', now(), now())`).Error; err != nil {
		t.Fatalf("insert later user: %v", err)
	}
	if err := runEmailVerificationMigration(pg.db); err != nil {
		t.Fatalf("rerun: %v", err)
	}
	if pg.count(t, `SELECT count(*) FROM users WHERE email = 'later@example.com' AND email_verified_at IS NULL`) != 1 {
		t.Fatalf("re-running the migration verified a later, unverified account")
	}
}
