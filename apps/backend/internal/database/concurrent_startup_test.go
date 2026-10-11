package database

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/sp3640/opspilot/backend/internal/config"
)

// These tests start several application instances at once against the SAME
// PostgreSQL database - what a multi-replica deployment does on its first
// rollout. Each Connect opens its own connection pool, exactly like a
// separate process would, and all of them are released together so their
// migrations genuinely overlap. A sequential double-run cannot reproduce
// the failures being guarded against: the schema check-then-create steps
// (does the table exist? then create it) only race when two instances are
// inside them at the same moment.
//
// Like the other PostgreSQL migration tests they need
// OPSPILOT_TEST_POSTGRES_DSN and are skipped, not faked, without it.

const concurrentInstances = 6

// startInstances runs Connect from n goroutines released at the same
// instant, and returns each instance's error.
func startInstances(t *testing.T, pg *postgresSchema, n int) []error {
	t.Helper()

	var (
		wg      sync.WaitGroup
		release = make(chan struct{})
		errs    = make([]error, n)
		handles = make([]*gorm.DB, n)
	)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-release
			handles[i], errs[i] = Connect(&config.Config{DBURL: pg.url})
		}(i)
	}
	close(release)
	wg.Wait()

	for _, handle := range handles {
		if handle == nil {
			continue
		}
		if sqlDB, err := handle.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	return errs
}

func requireAllStarted(t *testing.T, errs []error) {
	t.Helper()
	for i, err := range errs {
		if err != nil {
			t.Errorf("instance %d failed to start: %v", i, err)
		}
	}
	if t.Failed() {
		t.FailNow()
	}
}

// schemaShape is everything a first-time migration creates, by name, so two
// schemas can be compared for exact equality.
type schemaShape struct {
	Tables      []string
	Columns     []string
	Indexes     []string
	Constraints []string
}

func (s *postgresSchema) shape(t *testing.T) schemaShape {
	t.Helper()

	strings := func(query string, args ...any) []string {
		var values []string
		if err := s.db.Raw(query, args...).Scan(&values).Error; err != nil {
			t.Fatalf("query %q: %v", query, err)
		}
		sort.Strings(values)
		return values
	}

	return schemaShape{
		Tables: strings(`SELECT table_name FROM information_schema.tables WHERE table_schema = ? AND table_type = 'BASE TABLE'`, s.name),
		Columns: strings(`SELECT table_name || '.' || column_name || ':' || data_type || ':' || is_nullable
			FROM information_schema.columns WHERE table_schema = ?`, s.name),
		Indexes: strings(`SELECT tablename || '.' || indexname FROM pg_indexes WHERE schemaname = ?`, s.name),
		Constraints: strings(`SELECT c.conrelid::regclass::text || '.' || c.conname || ':' || c.contype::text
			FROM pg_constraint c JOIN pg_namespace n ON n.oid = c.connamespace WHERE n.nspname = ?`, s.name),
	}
}

func (s schemaShape) diff(other schemaShape) string {
	var problems []string
	compare := func(name string, a, b []string) {
		if reflect.DeepEqual(a, b) {
			return
		}
		missing, extra := setDifference(b, a), setDifference(a, b)
		problems = append(problems, fmt.Sprintf("%s differ: missing %v, unexpected %v", name, missing, extra))
	}
	compare("tables", s.Tables, other.Tables)
	compare("columns", s.Columns, other.Columns)
	compare("indexes", s.Indexes, other.Indexes)
	compare("constraints", s.Constraints, other.Constraints)
	return strings.Join(problems, "\n")
}

func setDifference(a, b []string) []string {
	in := map[string]bool{}
	for _, value := range b {
		in[value] = true
	}
	var out []string
	for _, value := range a {
		if !in[value] {
			out = append(out, value)
		}
	}
	return out
}

// TestConcurrentFirstStartupOnEmptyPostgres is the acceptance criterion:
// several instances starting together on one EMPTY database must all
// initialize successfully and leave exactly the schema a single instance
// would have created.
func TestConcurrentFirstStartupOnEmptyPostgres(t *testing.T) {
	reference := newEmptyPostgresSchema(t)
	reference.connect(t)
	expected := reference.shape(t)
	if got := int64(len(expected.Tables)); got != int64(len(migrationModels())) {
		t.Fatalf("reference schema has %d tables, want %d", got, len(migrationModels()))
	}

	// A single attempt can get lucky, so repeat with a fresh empty database
	// each round.
	for round := 1; round <= 3; round++ {
		t.Run(fmt.Sprintf("round %d", round), func(t *testing.T) {
			pg := newEmptyPostgresSchema(t)
			if got := pg.tableCount(t); got != 0 {
				t.Fatalf("test database must start empty, has %d tables", got)
			}

			requireAllStarted(t, startInstances(t, pg, concurrentInstances))

			if problems := pg.shape(t).diff(expected); problems != "" {
				t.Fatalf("schema after concurrent first start differs from a single-instance start:\n%s", problems)
			}
			pg.assertUserOrganizationForeignKeys(t)
			if got, want := pg.foreignKeyCount(t), reference.foreignKeyCount(t); got != want {
				t.Errorf("foreign keys = %d, want %d", got, want)
			}
		})
	}
}

// TestConcurrentStartupOnLegacyPostgres covers the other multi-instance
// first rollout: an existing database from before organizations and email
// verification. The data migrations (default organization, grandfathering
// existing users) must each happen exactly once, not once per instance.
func TestConcurrentStartupOnLegacyPostgres(t *testing.T) {
	pg := newEmptyPostgresSchema(t)
	pg.seedLegacyUsers(t)

	requireAllStarted(t, startInstances(t, pg, concurrentInstances))

	if got := pg.count(t, `SELECT count(*) FROM users`); got != 3 {
		t.Errorf("legacy users lost or duplicated: %d, want 3", got)
	}
	if got := pg.count(t, `SELECT count(*) FROM organizations`); got != 1 {
		t.Errorf("%d default organizations were created, want exactly 1", got)
	}
	if got := pg.count(t, `SELECT count(*) FROM users u JOIN organizations o ON o.id = u.organization_id`); got != 3 {
		t.Errorf("only %d of 3 legacy users were linked to the default organization", got)
	}
	if got := pg.count(t, `SELECT count(*) FROM users WHERE email_verified_at IS NOT NULL`); got != 3 {
		t.Errorf("only %d of 3 legacy users were grandfathered as verified", got)
	}
	pg.assertUserOrganizationForeignKeys(t)
}

// TestConcurrentStartupOnInitializedPostgresPreservesData is a rolling
// restart of every replica at once against a database that already holds
// data.
func TestConcurrentStartupOnInitializedPostgresPreservesData(t *testing.T) {
	pg := newEmptyPostgresSchema(t)
	pg.connect(t)

	const orgID = "6f1c5f3e-8b0a-4a44-9a55-0d2f5d9c1a11"
	for _, statement := range []string{
		`INSERT INTO users (id, name, email, password_hash, role, email_verified_at, created_at, updated_at) VALUES (1, 'Owner', 'owner@example.com', 'x', 'Platform Admin', now(), now(), now())`,
		`INSERT INTO organizations (id, name, slug, owner_id, created_at, updated_at) VALUES ('` + orgID + `', 'Acme', 'acme', 1, now(), now())`,
		`UPDATE users SET organization_id = '` + orgID + `'`,
	} {
		if err := pg.db.Exec(statement).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	before := pg.shape(t)
	checksum := func() int64 {
		return pg.count(t, `SELECT (SELECT count(*) FROM users) * 1000 + (SELECT count(*) FROM organizations) * 100
			+ (SELECT count(*) FROM users WHERE email_verified_at IS NOT NULL AND organization_id = ?)`, orgID)
	}
	dataBefore := checksum()

	requireAllStarted(t, startInstances(t, pg, concurrentInstances))

	if problems := pg.shape(t).diff(before); problems != "" {
		t.Fatalf("a concurrent restart changed the schema:\n%s", problems)
	}
	if got := checksum(); got != dataBefore {
		t.Fatalf("a concurrent restart changed existing data")
	}
}

// heldMigrationLock takes the startup migration lock on its own connection,
// as a slower instance already in the middle of migrating would hold it,
// and reports when another session is queued behind it.
type heldMigrationLock struct {
	tx *gorm.DB
}

func holdMigrationLock(t *testing.T, pg *postgresSchema) *heldMigrationLock {
	t.Helper()
	tx := pg.db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin: %v", tx.Error)
	}
	if err := tx.Exec(`SELECT pg_advisory_xact_lock(?)`, migrationLockKey).Error; err != nil {
		t.Fatalf("hold migration lock: %v", err)
	}
	held := &heldMigrationLock{tx: tx}
	t.Cleanup(func() { held.release() })
	return held
}

func (h *heldMigrationLock) release() {
	if h.tx != nil {
		_ = h.tx.Rollback().Error
		h.tx = nil
	}
}

func waitingOnMigrationLock(pg *postgresSchema, t *testing.T) int64 {
	return pg.count(t, `SELECT count(*) FROM pg_locks WHERE locktype = 'advisory' AND NOT granted AND classid::bigint = 0 AND objid::bigint = ?
		AND database = (SELECT oid FROM pg_database WHERE datname = current_database())`, migrationLockKey)
}

// TestStartupWaitsForTheMigrationLockBeforeTouchingTheSchema proves the lock
// is taken BEFORE any migration step runs: while another session holds it,
// a starting instance must be queued on it and must not have created a
// single table. Once the holder finishes, the waiter proceeds and succeeds.
func TestStartupWaitsForTheMigrationLockBeforeTouchingTheSchema(t *testing.T) {
	pg := newEmptyPostgresSchema(t)
	held := holdMigrationLock(t, pg)

	done := make(chan error, 1)
	go func() {
		handle, err := Connect(&config.Config{DBURL: pg.url})
		if handle != nil {
			if sqlDB, dbErr := handle.DB(); dbErr == nil {
				_ = sqlDB.Close()
			}
		}
		done <- err
	}()

	// Wait (bounded) until the instance is visibly queued on the lock.
	deadline := time.Now().Add(20 * time.Second)
	for waitingOnMigrationLock(pg, t) == 0 {
		select {
		case err := <-done:
			t.Fatalf("Connect finished without waiting for the held migration lock (err=%v)", err)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatalf("Connect never queued on the migration lock")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if got := pg.tableCount(t); got != 0 {
		t.Fatalf("an instance created %d table(s) before acquiring the migration lock", got)
	}

	held.release()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Connect failed after the lock was released: %v", err)
		}
	case <-time.After(60 * time.Second):
		t.Fatalf("Connect did not proceed after the migration lock was released")
	}
	if got, want := pg.tableCount(t), int64(len(migrationModels())); got != want {
		t.Fatalf("created %d tables, want %d", got, want)
	}
}

// TestStartupGivesUpWhenTheMigrationLockIsNeverReleased makes sure a wedged
// migrator cannot hang every other replica forever: the wait is bounded and
// fails with an error the orchestrator can act on (restart the pod).
func TestStartupGivesUpWhenTheMigrationLockIsNeverReleased(t *testing.T) {
	pg := newEmptyPostgresSchema(t)
	holdMigrationLock(t, pg)

	previous := migrationLockTimeout
	migrationLockTimeout = 300 * time.Millisecond
	t.Cleanup(func() { migrationLockTimeout = previous })

	handle, err := Connect(&config.Config{DBURL: pg.url})
	if handle != nil {
		t.Fatalf("Connect returned a handle although it could not take the migration lock")
	}
	if err == nil || !strings.Contains(err.Error(), "migration lock") {
		t.Fatalf("expected a migration lock timeout error, got %v", err)
	}
	if got := pg.tableCount(t); got != 0 {
		t.Fatalf("a timed-out instance still created %d table(s)", got)
	}
}
