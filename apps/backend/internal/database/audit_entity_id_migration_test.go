package database

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/config"
)

// The audit_logs.entity_id column used to be varchar(36), sized for UUIDs,
// while metric snapshots ("<uuid>:<time>", 57 chars) and resource-sync
// fallbacks (up to 357 chars) are longer. PostgreSQL rejects those inserts;
// SQLite does not, so only a real PostgreSQL run can prove the fix. These
// tests are skipped without OPSPILOT_TEST_POSTGRES_DSN, see
// postgres_migration_test.go.

const (
	auditEntityIDColumnQuery = `SELECT COALESCE(character_maximum_length, -1) FROM information_schema.columns
		WHERE table_schema = ? AND table_name = 'audit_logs' AND column_name = 'entity_id'`
	auditEntityIDIndexCountQuery = `SELECT count(*) FROM pg_indexes
		WHERE schemaname = ? AND tablename = 'audit_logs' AND indexdef ILIKE '%(entity_id)%'`
	auditEntityIndexQuery = `SELECT count(*) FROM pg_indexes
		WHERE schemaname = ? AND tablename = 'audit_logs' AND indexname = 'idx_audit_logs_entity'`
)

func (s *postgresSchema) insertAuditLog(entityType, entityID string) error {
	return s.db.Exec(`INSERT INTO audit_logs (organization_id, user_id, entity_type, entity_id, action, result, created_at)
		VALUES (?, 1, ?, ?, 'CREATE', 'SUCCESS', now())`, uuid.New(), entityType, entityID).Error
}

// widestResourceEntityID mirrors the resource sync fallback identifier
// "<project uuid>:<kind (64)>:<external id (255)>".
func widestResourceEntityID() string {
	return uuid.NewString() + ":" + strings.Repeat("k", 64) + ":" + strings.Repeat("e", 255)
}

func TestConnectCreatesAuditEntityIDWideEnoughOnEmptyPostgres(t *testing.T) {
	pg := newEmptyPostgresSchema(t)
	pg.connect(t)

	if got := pg.count(t, auditEntityIDColumnQuery, pg.name); got < 357 {
		t.Fatalf("audit_logs.entity_id holds %d chars, want at least 357 (widest composite identifier)", got)
	}

	for name, id := range map[string]string{
		"uuid":             uuid.NewString(),
		"numeric":          "18446744073709551615",
		"metric snapshot":  uuid.NewString() + ":2026-09-20T07:30:45Z",
		"resource sync":    widestResourceEntityID(),
		"integration:none": "integration:none",
	} {
		if err := pg.insertAuditLog("entity", id); err != nil {
			t.Errorf("%s entity id (%d chars) rejected: %v", name, len(id), err)
		}
	}
}

// TestConnectWidensLegacyAuditEntityID starts from the schema every existing
// deployment has - a varchar(36) entity_id holding real audit history - and
// checks the startup migration widens it in place without touching data.
func TestConnectWidensLegacyAuditEntityID(t *testing.T) {
	pg := newEmptyPostgresSchema(t)
	pg.connect(t)

	// Recreate the legacy column type on an initialized schema: same table,
	// same indexes, only the original width.
	if err := pg.db.Exec(`ALTER TABLE audit_logs ALTER COLUMN entity_id TYPE varchar(36)`).Error; err != nil {
		t.Fatalf("recreate legacy entity_id width: %v", err)
	}
	legacyIDs := []string{uuid.NewString(), uuid.NewString(), "42", "integration:none"}
	for _, id := range legacyIDs {
		if err := pg.insertAuditLog("legacy", id); err != nil {
			t.Fatalf("seed legacy audit row: %v", err)
		}
	}
	if err := pg.insertAuditLog("legacy", widestResourceEntityID()); err == nil {
		t.Fatalf("test setup is wrong: legacy varchar(36) accepted a 357-char id")
	}

	if got := pg.count(t, auditEntityIDColumnQuery, pg.name); got != 36 {
		t.Fatalf("legacy setup has entity_id width %d, want 36", got)
	}
	rowsBefore := pg.count(t, `SELECT count(*) FROM audit_logs`)
	indexesBefore := pg.count(t, auditEntityIDIndexCountQuery, pg.name)
	checksumBefore := pg.count(t, `SELECT COALESCE(sum(hashtext(id::text || entity_type || entity_id || action || result || created_at::text)), 0) FROM audit_logs`)

	pg.connect(t)

	if got := pg.count(t, auditEntityIDColumnQuery, pg.name); got < 357 {
		t.Errorf("legacy entity_id not widened: holds %d chars", got)
	}
	if got := pg.count(t, `SELECT count(*) FROM audit_logs`); got != rowsBefore {
		t.Errorf("audit rows changed by migration: %d -> %d", rowsBefore, got)
	}
	if got := pg.count(t, `SELECT COALESCE(sum(hashtext(id::text || entity_type || entity_id || action || result || created_at::text)), 0) FROM audit_logs`); got != checksumBefore {
		t.Errorf("audit row contents changed by migration")
	}
	for _, id := range legacyIDs {
		if pg.count(t, `SELECT count(*) FROM audit_logs WHERE entity_type = 'legacy' AND entity_id = ?`, id) != 1 {
			t.Errorf("legacy audit row with entity id %q lost", id)
		}
	}
	if got := pg.count(t, auditEntityIDIndexCountQuery, pg.name); got != indexesBefore || got == 0 {
		t.Errorf("entity_id indexes changed: %d -> %d", indexesBefore, got)
	}
	if pg.count(t, auditEntityIndexQuery, pg.name) != 1 {
		t.Errorf("composite idx_audit_logs_entity missing after migration")
	}

	if err := pg.insertAuditLog("resource", widestResourceEntityID()); err != nil {
		t.Errorf("widest identifier still rejected after migration: %v", err)
	}
}

// TestConnectAuditEntityIDMigrationIsRepeatable restarts several times and
// checks the widening never narrows, duplicates indexes, or loses rows.
func TestConnectAuditEntityIDMigrationIsRepeatable(t *testing.T) {
	pg := newEmptyPostgresSchema(t)
	pg.connect(t)
	if err := pg.insertAuditLog("project", uuid.NewString()); err != nil {
		t.Fatalf("seed audit row: %v", err)
	}
	if err := pg.insertAuditLog("resource", widestResourceEntityID()); err != nil {
		t.Fatalf("seed wide audit row: %v", err)
	}
	width := pg.count(t, auditEntityIDColumnQuery, pg.name)
	indexes := pg.count(t, `SELECT count(*) FROM pg_indexes WHERE schemaname = ? AND tablename = 'audit_logs'`, pg.name)

	pg.connect(t)
	pg.connect(t)

	if got := pg.count(t, auditEntityIDColumnQuery, pg.name); got != width {
		t.Errorf("entity_id width changed across restarts: %d -> %d", width, got)
	}
	if got := pg.count(t, `SELECT count(*) FROM pg_indexes WHERE schemaname = ? AND tablename = 'audit_logs'`, pg.name); got != indexes {
		t.Errorf("audit_logs indexes changed across restarts: %d -> %d", indexes, got)
	}
	if got := pg.count(t, `SELECT count(*) FROM audit_logs`); got != 2 {
		t.Errorf("audit rows after restarts = %d, want 2", got)
	}
}

// TestConnectNeverTruncatesAuditEntityID covers a column an operator already
// widened by hand (say, as an emergency workaround for the original bug).
// AutoMigrate alone would narrow it back to the model width with an explicit
// cast, which silently truncates. Startup must either leave every value
// intact or refuse to start - never cut audit history short.
func TestConnectNeverTruncatesAuditEntityID(t *testing.T) {
	t.Run("wider column with values that fit is normalized without loss", func(t *testing.T) {
		pg := newEmptyPostgresSchema(t)
		pg.connect(t)
		if err := pg.db.Exec(`ALTER TABLE audit_logs ALTER COLUMN entity_id TYPE text`).Error; err != nil {
			t.Fatalf("widen entity_id manually: %v", err)
		}
		id := widestResourceEntityID()
		if err := pg.insertAuditLog("manual", id); err != nil {
			t.Fatalf("seed audit row: %v", err)
		}

		pg.connect(t)

		if pg.count(t, `SELECT count(*) FROM audit_logs WHERE entity_type = 'manual' AND entity_id = ?`, id) != 1 {
			t.Errorf("audit row was altered by a restart")
		}
		if got := pg.count(t, auditEntityIDColumnQuery, pg.name); got < 357 {
			t.Errorf("entity_id ended up %d wide, want at least 357", got)
		}
	})

	t.Run("values longer than the model width stop startup and are preserved", func(t *testing.T) {
		pg := newEmptyPostgresSchema(t)
		pg.connect(t)
		if err := pg.db.Exec(`ALTER TABLE audit_logs ALTER COLUMN entity_id TYPE text`).Error; err != nil {
			t.Fatalf("widen entity_id manually: %v", err)
		}
		long := uuid.NewString() + ":" + strings.Repeat("x", 1500)
		if err := pg.insertAuditLog("manual", long); err != nil {
			t.Fatalf("seed long audit row: %v", err)
		}

		if _, err := Connect(&config.Config{DBURL: pg.url}); err == nil {
			t.Fatalf("Connect succeeded although narrowing entity_id would truncate audit history")
		} else if !strings.Contains(err.Error(), "refusing to narrow") {
			t.Errorf("unexpected error: %v", err)
		}

		if pg.count(t, `SELECT count(*) FROM audit_logs WHERE entity_type = 'manual' AND entity_id = ?`, long) != 1 {
			t.Errorf("long audit row was truncated or lost")
		}
	})
}

// TestAuditEntityIDMigrationWidensWithoutAutoMigrate proves the explicit
// migration step does the widening itself, rather than the test passing
// only because AutoMigrate happens to alter the column afterwards.
func TestAuditEntityIDMigrationWidensWithoutAutoMigrate(t *testing.T) {
	pg := newEmptyPostgresSchema(t)
	pg.connect(t)
	if err := pg.db.Exec(`ALTER TABLE audit_logs ALTER COLUMN entity_id TYPE varchar(36)`).Error; err != nil {
		t.Fatalf("recreate legacy entity_id width: %v", err)
	}
	if err := pg.insertAuditLog("legacy", uuid.NewString()); err != nil {
		t.Fatalf("seed legacy audit row: %v", err)
	}

	for run := 1; run <= 2; run++ {
		if err := runAuditEntityIDMigration(pg.db); err != nil {
			t.Fatalf("migration run %d: %v", run, err)
		}
		if got := pg.count(t, auditEntityIDColumnQuery, pg.name); got != 512 {
			t.Fatalf("after run %d entity_id holds %d chars, want 512", run, got)
		}
	}
	if got := pg.count(t, `SELECT count(*) FROM audit_logs`); got != 1 {
		t.Errorf("audit rows = %d, want 1", got)
	}
}
