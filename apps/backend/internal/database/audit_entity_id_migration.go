package database

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/sp3640/opspilot/backend/internal/models"
)

const (
	auditLogsTable          = "audit_logs"
	auditEntityIDColumn     = "entity_id"
	postgresDialectName     = "postgres"
	postgresVarcharDataType = "character varying"
)

// runAuditEntityIDMigration brings audit_logs.entity_id to the width the
// AuditLog model declares, widening only.
//
// The column used to be varchar(36), sized for UUIDs, but metric snapshots
// ("<project uuid>:<time>") and resource-sync fallbacks
// ("<project uuid>:<kind>:<external id>") are longer, and PostgreSQL rejects
// those inserts. Existing databases therefore need the column widened in
// place. It is done explicitly, before AutoMigrate, because AutoMigrate's
// own alter converts with an explicit cast and would silently truncate
// values if the column were ever wider than the model.
//
//   - Column narrower than the model: ALTER COLUMN ... TYPE varchar(n) with
//     no USING clause. Growing a varchar limit is a catalog-only change in
//     PostgreSQL (no table rewrite, no index rebuild, no data touched) and
//     the plain assignment cast can only error, never truncate.
//   - Column already the right width: nothing to do, so every later startup
//     is a no-op.
//   - Column wider than the model (or unbounded text, e.g. widened by hand
//     as an emergency workaround): left alone if every stored value fits the
//     model width, which lets AutoMigrate normalise it without loss.
//     If any stored value is longer, startup stops with an error rather than
//     let a cast cut audit history short.
//
// A fresh database has no audit_logs table yet and is skipped; AutoMigrate
// creates it at the model width. No rows are ever modified.
func runAuditEntityIDMigration(db *gorm.DB) error {
	if db.Dialector.Name() != postgresDialectName {
		return nil
	}
	if !db.Migrator().HasTable(auditLogsTable) || !db.Migrator().HasColumn(auditLogsTable, auditEntityIDColumn) {
		return nil
	}

	statement := &gorm.Statement{DB: db}
	if err := statement.Parse(&models.AuditLog{}); err != nil {
		return fmt.Errorf("parse audit log model: %w", err)
	}
	field := statement.Schema.LookUpField("EntityID")
	if field == nil || field.Size <= 0 {
		return fmt.Errorf("audit log model has no bounded EntityID column")
	}
	target := field.Size

	var column struct {
		DataType string
		MaxLen   *int
	}
	if err := db.Raw(`SELECT data_type AS data_type, character_maximum_length AS max_len
		FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = ? AND column_name = ?`,
		auditLogsTable, auditEntityIDColumn).Scan(&column).Error; err != nil {
		return fmt.Errorf("inspect audit_logs.entity_id: %w", err)
	}

	if column.DataType == postgresVarcharDataType && column.MaxLen != nil && *column.MaxLen < target {
		if err := db.Exec(fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN %s TYPE varchar(%d)`, auditLogsTable, auditEntityIDColumn, target)).Error; err != nil {
			return fmt.Errorf("widen audit_logs.entity_id to varchar(%d): %w", target, err)
		}
		return nil
	}

	if column.DataType == postgresVarcharDataType && column.MaxLen != nil && *column.MaxLen == target {
		return nil
	}

	var tooLong int64
	if err := db.Raw(fmt.Sprintf(`SELECT count(*) FROM %s WHERE char_length(%s) > ?`, auditLogsTable, auditEntityIDColumn), target).Scan(&tooLong).Error; err != nil {
		return fmt.Errorf("check audit_logs.entity_id lengths: %w", err)
	}
	if tooLong > 0 {
		return fmt.Errorf("audit_logs.entity_id is wider than the model's varchar(%d) and %d stored value(s) exceed it; refusing to narrow the column and truncate audit history", target, tooLong)
	}

	return nil
}
