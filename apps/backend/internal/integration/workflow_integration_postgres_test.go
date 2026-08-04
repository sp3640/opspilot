package integration

import (
	"os"
	"strings"
	"testing"

	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const postgresIntegrationDSNEnv = "OPSPILOT_TEST_POSTGRES_DSN"

func TestOpsPilotCompleteWorkflowIntegrationPostgres(t *testing.T) {
	t.Parallel()

	dsn := strings.TrimSpace(os.Getenv(postgresIntegrationDSNEnv))
	if dsn == "" {
		t.Skipf("skipping postgres integration test: %s is not set", postgresIntegrationDSNEnv)
	}

	db := setupPostgresIntegrationDB(t, dsn)
	runCompleteWorkflowIntegration(t, db)
}

func setupPostgresIntegrationDB(t *testing.T, dsn string) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("skipping postgres integration test: open failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Skipf("skipping postgres integration test: database handle failed: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		t.Skipf("skipping postgres integration test: ping failed: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	resetPostgresIntegrationSchema(t, db)
	migrateIntegrationSchema(t, db)

	return db
}

func resetPostgresIntegrationSchema(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.Migrator().DropTable(
		&models.AuditLog{},
		&models.Comment{},
		&models.Metric{},
		&models.Resource{},
		&models.Cluster{},
		&models.Alert{},
		&models.Incident{},
		&models.TeamMember{},
		&models.Team{},
		&models.Project{},
		&models.Invitation{},
		&models.Organization{},
		&models.User{},
	); err != nil {
		t.Fatalf("reset postgres schema: %v", err)
	}
}
