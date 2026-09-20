package services

import (
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm/schema"

	"github.com/sp3640/opspilot/backend/internal/models"
)

// AuditLog.EntityID is not always a UUID. Writers across the backend store:
//
//	uuid            projects, clusters, applications, teams, ... (36 chars)
//	numeric id      incidents, alerts, comments, users
//	fixed token     "integration:none"
//	composite       metric snapshots "<project uuid>:<RFC3339 time>" and the
//	                resource sync fallback "<project uuid>:<kind>:<external id>"
//
// PostgreSQL enforces varchar length (SQLite does not), so a column sized
// for UUIDs silently rejects every composite audit write in production.
// These tests pin the column to the widest identifier the code can build.

func auditEntityIDColumnSize(t *testing.T) int {
	t.Helper()
	field := gormField(t, &models.AuditLog{}, "EntityID")
	return field.Size
}

func gormField(t *testing.T, model any, name string) *schema.Field {
	t.Helper()
	parsed, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse %T schema: %v", model, err)
	}
	field := parsed.LookUpField(name)
	if field == nil {
		t.Fatalf("%T has no field %s", model, name)
	}
	return field
}

func TestMetricSnapshotEntityIDFormat(t *testing.T) {
	projectID := uuid.MustParse("6f1c5f3e-8b0a-4a44-9a55-0d2f5d9c1a11")
	at := time.Date(2026, 9, 20, 12, 30, 45, 0, time.FixedZone("x", 5*3600))

	got := metricSnapshotEntityID(projectID, at)

	if want := "6f1c5f3e-8b0a-4a44-9a55-0d2f5d9c1a11:2026-09-20T07:30:45Z"; got != want {
		t.Errorf("snapshot entity id = %q, want %q (format is part of stored audit history)", got, want)
	}
}

func TestAuditEntityIDColumnFitsEveryIdentifierFormat(t *testing.T) {
	size := auditEntityIDColumnSize(t)

	kindSize := gormField(t, &models.Resource{}, "Kind").Size
	externalIDSize := gormField(t, &models.Resource{}, "ExternalID").Size
	widestResource := resourceEntityID(models.Resource{
		ProjectID:  uuid.New(),
		Kind:       strings.Repeat("k", kindSize),
		ExternalID: strings.Repeat("e", externalIDSize),
	})

	identifiers := map[string]string{
		"uuid":                     uuid.NewString(),
		"numeric id":               strconv.FormatUint(^uint64(0), 10),
		"integration placeholder":  "integration:none",
		"metric snapshot":          metricSnapshotEntityID(uuid.New(), time.Now()),
		"widest resource fallback": widestResource,
	}
	for name, id := range identifiers {
		if len(id) > size {
			t.Errorf("%s entity id is %d chars but audit_logs.entity_id holds %d: PostgreSQL would reject the audit row", name, len(id), size)
		}
	}
	if len(widestResource) <= 36 {
		t.Fatalf("test is not exercising a composite identifier (%d chars)", len(widestResource))
	}
}

func TestAuditEntityIDColumnStaysIndexable(t *testing.T) {
	// entity_id sits in two btree indexes; PostgreSQL refuses index rows
	// above roughly a third of a page (about 2.7KB). Keep wide but bounded.
	if size := auditEntityIDColumnSize(t); size <= 0 || size > 1024 {
		t.Errorf("audit_logs.entity_id size = %d, want a bounded varchar (1..1024)", size)
	}
}
