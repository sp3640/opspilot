package database

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/sp3640/opspilot/backend/internal/config"
)

// These tests run the real startup migration (Connect) against a genuine
// PostgreSQL server. SQLite cannot reproduce the failures they guard
// against: it does not enforce foreign-key creation order, so the
// "relation users does not exist" startup bug on an empty PostgreSQL
// database passed every SQLite test.
//
// They need a reachable PostgreSQL and are skipped - not faked with SQLite -
// when OPSPILOT_TEST_POSTGRES_DSN is unset (same variable the integration
// tests use), e.g.:
//
//	docker run -d --name pg -e POSTGRES_PASSWORD=pw -p 127.0.0.1:5432:5432 postgres:17-alpine
//	OPSPILOT_TEST_POSTGRES_DSN='postgres://postgres:pw@127.0.0.1:5432/postgres?sslmode=disable' go test ./internal/database/
//
// Each test creates its own uniquely named, initially empty schema and drops
// it afterwards, so nothing else in the database is touched and no
// application tables are ever pre-created.
const postgresMigrationDSNEnv = "OPSPILOT_TEST_POSTGRES_DSN"

type postgresSchema struct {
	name string
	url  string // DSN pinned to the empty schema via search_path
	db   *gorm.DB
}

func newEmptyPostgresSchema(t *testing.T) *postgresSchema {
	t.Helper()

	baseDSN := strings.TrimSpace(os.Getenv(postgresMigrationDSNEnv))
	if baseDSN == "" {
		t.Skipf("skipping PostgreSQL migration test: %s is not set", postgresMigrationDSNEnv)
	}

	admin, err := gorm.Open(postgres.Open(baseDSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("connect to PostgreSQL (%s is set but unreachable): %v", postgresMigrationDSNEnv, err)
	}
	adminSQL, err := admin.DB()
	if err != nil {
		t.Fatalf("PostgreSQL handle: %v", err)
	}

	suffix := make([]byte, 6)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatalf("random schema name: %v", err)
	}
	name := "opspilot_mig_" + hex.EncodeToString(suffix)
	if err := admin.Exec(`CREATE SCHEMA ` + name).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	dsn := baseDSN + " search_path=" + name
	if strings.Contains(baseDSN, "://") {
		separator := "?"
		if strings.Contains(baseDSN, "?") {
			separator = "&"
		}
		dsn = baseDSN + separator + "search_path=" + name
	}

	verify, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("connect to schema %s: %v", name, err)
	}
	verifySQL, _ := verify.DB()

	t.Cleanup(func() {
		_ = Close()
		_ = verifySQL.Close()
		_ = admin.Exec(`DROP SCHEMA ` + name + ` CASCADE`).Error
		_ = adminSQL.Close()
	})

	return &postgresSchema{name: name, url: dsn, db: verify}
}

func (s *postgresSchema) connect(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := Connect(&config.Config{DBURL: s.url})
	if err != nil {
		t.Fatalf("Connect failed against PostgreSQL schema %s: %v", s.name, err)
	}
	return db
}

func (s *postgresSchema) count(t *testing.T, query string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := s.db.Raw(query, args...).Scan(&n).Error; err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return n
}

func (s *postgresSchema) tableCount(t *testing.T) int64 {
	return s.count(t, `SELECT count(*) FROM information_schema.tables WHERE table_schema = ? AND table_type = 'BASE TABLE'`, s.name)
}

func (s *postgresSchema) foreignKeyCount(t *testing.T) int64 {
	return s.count(t, `SELECT count(*) FROM pg_constraint c JOIN pg_namespace n ON n.oid = c.connamespace WHERE c.contype = 'f' AND n.nspname = ?`, s.name)
}

func (s *postgresSchema) hasForeignKey(t *testing.T, name, table, referencedTable string) bool {
	t.Helper()
	return s.count(t, `SELECT count(*) FROM pg_constraint c
		JOIN pg_namespace n ON n.oid = c.connamespace
		WHERE c.contype = 'f' AND n.nspname = ? AND c.conname = ?
		  AND c.conrelid::regclass::text = ? AND c.confrelid::regclass::text = ?
		  AND c.confdeltype = 'r' AND c.confupdtype = 'c'`,
		s.name, name, table, referencedTable) == 1
}

func (s *postgresSchema) assertUserOrganizationForeignKeys(t *testing.T) {
	t.Helper()
	if !s.hasForeignKey(t, "fk_users_organization", "users", "organizations") {
		t.Errorf("missing foreign key users.organization_id -> organizations.id (RESTRICT/CASCADE)")
	}
	if !s.hasForeignKey(t, "fk_organizations_owner", "organizations", "users") {
		t.Errorf("missing foreign key organizations.owner_id -> users.id (RESTRICT/CASCADE)")
	}
}

// TestConnectInitializesEmptyPostgresDatabase is the deployment scenario:
// a brand-new PostgreSQL database with zero application tables.
func TestConnectInitializesEmptyPostgresDatabase(t *testing.T) {
	pg := newEmptyPostgresSchema(t)

	if got := pg.tableCount(t); got != 0 {
		t.Fatalf("test schema must start empty, has %d tables", got)
	}

	db := pg.connect(t)

	for _, model := range migrationModels() {
		if !db.Migrator().HasTable(model) {
			t.Errorf("table for %T was not created", model)
		}
	}
	if got, want := pg.tableCount(t), int64(len(migrationModels())); got != want {
		t.Errorf("created %d tables, want %d", got, want)
	}

	pg.assertUserOrganizationForeignKeys(t)

	for _, index := range []struct{ table, name string }{
		{"users", "idx_users_organization_id"},
		{"organizations", "idx_organizations_owner_id"},
		{"organizations", "idx_organizations_slug"},
	} {
		if pg.count(t, `SELECT count(*) FROM pg_indexes WHERE schemaname = ? AND tablename = ? AND indexname = ?`, pg.name, index.table, index.name) != 1 {
			t.Errorf("missing index %s on %s", index.name, index.table)
		}
	}
	if pg.count(t, `SELECT count(*) FROM pg_indexes WHERE schemaname = ? AND tablename = 'users' AND indexdef ILIKE '%UNIQUE%' AND indexdef ILIKE '%(email)%'`, pg.name) != 1 {
		t.Errorf("missing unique index on users.email")
	}

	// Integrity must actually be enforced, not merely declared.
	if err := pg.db.Exec(`INSERT INTO organizations (id, name, slug, owner_id, created_at, updated_at)
		VALUES (gen_random_uuid(), 'Orphan', 'orphan', 999999, now(), now())`).Error; err == nil {
		t.Errorf("organization with a nonexistent owner was accepted; foreign key is not enforced")
	}
}

// TestConnectIsIdempotentAndPreservesUserOrganizationData runs startup twice
// against an initialized database that holds real User/Organization rows.
func TestConnectIsIdempotentAndPreservesUserOrganizationData(t *testing.T) {
	pg := newEmptyPostgresSchema(t)
	pg.connect(t)

	// Circular data has to be created in two steps: user, then org, then link.
	mustExec := func(query string, args ...any) {
		t.Helper()
		if err := pg.db.Exec(query, args...).Error; err != nil {
			t.Fatalf("%s: %v", query, err)
		}
	}
	const orgID = "6f1c5f3e-8b0a-4a44-9a55-0d2f5d9c1a11"
	mustExec(`INSERT INTO users (id, name, email, password_hash, role, created_at, updated_at) VALUES (1, 'Owner', 'owner@example.com', 'x', 'Platform Admin', now(), now())`)
	mustExec(`INSERT INTO organizations (id, name, slug, owner_id, created_at, updated_at) VALUES (?, 'Acme', 'acme', 1, now(), now())`, orgID)
	mustExec(`UPDATE users SET organization_id = ? WHERE id = 1`, orgID)

	tablesBefore, keysBefore := pg.tableCount(t), pg.foreignKeyCount(t)

	pg.connect(t)
	pg.connect(t)

	if got := pg.tableCount(t); got != tablesBefore {
		t.Errorf("tables changed across restarts: %d -> %d", tablesBefore, got)
	}
	if got := pg.foreignKeyCount(t); got != keysBefore {
		t.Errorf("foreign keys changed across restarts (duplicate or dropped): %d -> %d", keysBefore, got)
	}
	pg.assertUserOrganizationForeignKeys(t)

	if got := pg.count(t, `SELECT count(*) FROM users WHERE id = 1 AND email = 'owner@example.com' AND role = 'Platform Admin' AND organization_id = ?`, orgID); got != 1 {
		t.Errorf("user row or its organization link was altered by a restart")
	}
	if got := pg.count(t, `SELECT count(*) FROM organizations WHERE id = ? AND owner_id = 1 AND slug = 'acme'`, orgID); got != 1 {
		t.Errorf("organization row or its owner link was altered by a restart")
	}
	if got := pg.count(t, `SELECT count(*) FROM organizations`); got != 1 {
		t.Errorf("restart created %d organizations, want the original 1", got)
	}
}

// TestConnectMigratesLegacyPostgresDatabase covers an existing database from
// before organizations existed: users with data, no organizations table.
func TestConnectMigratesLegacyPostgresDatabase(t *testing.T) {
	pg := newEmptyPostgresSchema(t)

	for _, statement := range []string{
		`CREATE TABLE users (
			id bigserial PRIMARY KEY,
			name varchar(100) NOT NULL,
			email varchar(255) NOT NULL,
			password_hash text NOT NULL,
			role varchar(50) NOT NULL DEFAULT 'User',
			created_at timestamptz,
			updated_at timestamptz)`,
		`CREATE UNIQUE INDEX idx_users_email ON users (email)`,
		`INSERT INTO users (name, email, password_hash, role) VALUES ('First', 'first@example.com', 'x', 'User'), ('Second', 'second@example.com', 'x', 'Platform Admin')`,
	} {
		if err := pg.db.Exec(statement).Error; err != nil {
			t.Fatalf("seed legacy schema: %v", err)
		}
	}

	pg.connect(t)

	if got := pg.count(t, `SELECT count(*) FROM users`); got != 2 {
		t.Fatalf("legacy users lost or duplicated: have %d, want 2", got)
	}
	if got := pg.count(t, `SELECT count(*) FROM organizations`); got != 1 {
		t.Fatalf("expected one default organization for legacy users, have %d", got)
	}
	if got := pg.count(t, `SELECT count(*) FROM users u JOIN organizations o ON o.id = u.organization_id AND o.owner_id = (SELECT min(id) FROM users)`); got != 2 {
		t.Errorf("legacy users were not all linked to the default organization owned by the first user")
	}
	if got := pg.count(t, `SELECT count(*) FROM users WHERE email = 'first@example.com' AND role = 'Viewer'`); got != 1 {
		t.Errorf("legacy 'User' role was not normalized to 'Viewer'")
	}
	if got := pg.count(t, `SELECT count(*) FROM users WHERE email = 'second@example.com' AND role = 'Platform Admin'`); got != 1 {
		t.Errorf("Platform Admin role was modified")
	}
	pg.assertUserOrganizationForeignKeys(t)
}
