package database

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

const organizationIDColumn = "organization_id"

type organizationScopedTable struct {
	name         string
	backfillSQLs []string
}

var organizationScopedTables = []organizationScopedTable{
	{
		name: "projects",
		backfillSQLs: []string{
			`UPDATE projects p
			 SET organization_id = u.organization_id
			 FROM users u
			 WHERE p.organization_id IS NULL
			   AND u.id = p.owner_id
			   AND u.organization_id IS NOT NULL`,
		},
	},
	{
		name: "clusters",
		backfillSQLs: []string{
			`UPDATE clusters c
			 SET organization_id = p.organization_id
			 FROM projects p
			 WHERE c.organization_id IS NULL
			   AND p.id = c.project_id`,
		},
	},
	{
		name: "resources",
		backfillSQLs: []string{
			`UPDATE resources r
			 SET organization_id = p.organization_id
			 FROM projects p
			 WHERE r.organization_id IS NULL
			   AND p.id = r.project_id`,
		},
	},
	{
		name: "incidents",
		backfillSQLs: []string{
			`UPDATE incidents i
			 SET organization_id = p.organization_id
			 FROM projects p
			 WHERE i.organization_id IS NULL
			   AND p.id = i.project_id`,
			`UPDATE incidents i
			 SET organization_id = u.organization_id
			 FROM users u
			 WHERE i.organization_id IS NULL
			   AND u.id = i.user_id
			   AND u.organization_id IS NOT NULL`,
		},
	},
	{
		name: "alerts",
		backfillSQLs: []string{
			`UPDATE alerts a
			 SET organization_id = p.organization_id
			 FROM projects p
			 WHERE a.organization_id IS NULL
			   AND p.id = a.project_id`,
		},
	},
	{
		name: "metrics",
		backfillSQLs: []string{
			`UPDATE metrics m
			 SET organization_id = p.organization_id
			 FROM projects p
			 WHERE m.organization_id IS NULL
			   AND p.id = m.project_id`,
		},
	},
	{
		name: "audit_logs",
		backfillSQLs: []string{
			`UPDATE audit_logs a
			 SET organization_id = p.organization_id
			 FROM projects p
			 WHERE a.organization_id IS NULL
			   AND a.project_id = p.id`,
			`UPDATE audit_logs a
			 SET organization_id = i.organization_id
			 FROM incidents i
			 WHERE a.organization_id IS NULL
			   AND a.incident_id = i.id`,
			`UPDATE audit_logs a
			 SET organization_id = u.organization_id
			 FROM users u
			 WHERE a.organization_id IS NULL
			   AND a.user_id = u.id
			   AND u.organization_id IS NOT NULL`,
		},
	},
}

func runOrganizationBackfillMigration(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&models.User{}, &models.Organization{}); err != nil {
			return fmt.Errorf("migrate user and organization foundation: %w", err)
		}

		defaultOrganizationID, err := ensureDefaultOrganization(tx)
		if err != nil {
			return err
		}

		if err := ensureUserOrganizationIDs(tx, defaultOrganizationID); err != nil {
			return err
		}

		for _, table := range organizationScopedTables {
			if err := ensureOrganizationColumnBackfilled(tx, table, defaultOrganizationID); err != nil {
				return err
			}
		}

		return nil
	})
}

func ensureDefaultOrganization(tx *gorm.DB) (uuid.UUID, error) {
	if !tx.Migrator().HasTable("organizations") {
		return uuid.Nil, errors.New("organizations table does not exist")
	}

	var existingIDValue string
	err := tx.Raw(`SELECT id::text FROM organizations ORDER BY created_at ASC, id ASC LIMIT 1`).Scan(&existingIDValue).Error
	if err == nil && existingIDValue != "" {
		existingID, parseErr := uuid.Parse(existingIDValue)
		if parseErr != nil {
			return uuid.Nil, fmt.Errorf("parse existing organization id %q: %w", existingIDValue, parseErr)
		}
		if existingID != uuid.Nil {
			return existingID, nil
		}
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, fmt.Errorf("query existing organization: %w", err)
	}

	ownerID, hasOwner, err := findDefaultOrganizationOwnerID(tx)
	if err != nil {
		return uuid.Nil, err
	}
	if !hasOwner {
		return uuid.Nil, nil
	}

	defaultID := uuid.New()
	if err := tx.Exec(`
		INSERT INTO organizations (id, name, slug, description, owner_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`, defaultID, "Default Organization", "default-organization", "Auto-created during organization migration", ownerID).Error; err != nil {
		return uuid.Nil, fmt.Errorf("create default organization: %w", err)
	}

	return defaultID, nil
}

func findDefaultOrganizationOwnerID(tx *gorm.DB) (uint, bool, error) {
	var ownerID uint
	if err := tx.Raw(`SELECT id FROM users ORDER BY id ASC LIMIT 1`).Scan(&ownerID).Error; err != nil {
		return 0, false, fmt.Errorf("select default organization owner: %w", err)
	}
	if ownerID == 0 {
		return 0, false, nil
	}

	return ownerID, true, nil
}

func ensureUserOrganizationIDs(tx *gorm.DB, defaultOrganizationID uuid.UUID) error {
	if !tx.Migrator().HasTable("users") {
		return nil
	}

	if !tx.Migrator().HasColumn("users", organizationIDColumn) {
		if err := tx.Exec(`ALTER TABLE users ADD COLUMN organization_id uuid`).Error; err != nil {
			return fmt.Errorf("add users.organization_id: %w", err)
		}
	}

	if defaultOrganizationID == uuid.Nil {
		return nil
	}

	if err := tx.Exec(`UPDATE users SET organization_id = ? WHERE organization_id IS NULL`, defaultOrganizationID).Error; err != nil {
		return fmt.Errorf("backfill users.organization_id: %w", err)
	}

	return nil
}

func ensureOrganizationColumnBackfilled(tx *gorm.DB, table organizationScopedTable, defaultOrganizationID uuid.UUID) error {
	if !tx.Migrator().HasTable(table.name) {
		return nil
	}

	if !tx.Migrator().HasColumn(table.name, organizationIDColumn) {
		if err := tx.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN organization_id uuid`, table.name)).Error; err != nil {
			return fmt.Errorf("add %s.organization_id: %w", table.name, err)
		}
	}

	for _, statement := range table.backfillSQLs {
		if err := tx.Exec(statement).Error; err != nil {
			return fmt.Errorf("backfill %s.organization_id: %w", table.name, err)
		}
	}

	if defaultOrganizationID != uuid.Nil {
		if err := tx.Exec(fmt.Sprintf(`UPDATE %s SET organization_id = ? WHERE organization_id IS NULL`, table.name), defaultOrganizationID).Error; err != nil {
			return fmt.Errorf("fallback backfill %s.organization_id: %w", table.name, err)
		}
	}

	if err := assertNoNullOrganizationIDs(tx, table.name); err != nil {
		return err
	}

	if err := tx.Exec(fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN organization_id SET NOT NULL`, table.name)).Error; err != nil {
		return fmt.Errorf("set %s.organization_id not null: %w", table.name, err)
	}

	return nil
}

func assertNoNullOrganizationIDs(tx *gorm.DB, tableName string) error {
	var nullCount int64
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE organization_id IS NULL`, tableName)
	if err := tx.Raw(query).Scan(&nullCount).Error; err != nil {
		return fmt.Errorf("count null organization ids for %s: %w", tableName, err)
	}
	if nullCount > 0 {
		return fmt.Errorf("cannot enforce not null on %s.organization_id: %d rows still null", tableName, nullCount)
	}

	return nil
}
