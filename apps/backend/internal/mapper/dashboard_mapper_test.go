package mapper

import (
	"testing"
	"time"

	"github.com/sp3640/opspilot/backend/internal/models"
)

func TestGenerateGreeting(t *testing.T) {
	// This test is time-dependent, so we'll just verify it returns a non-empty string
	greeting := GenerateGreeting()
	if greeting == "" {
		t.Error("GenerateGreeting() returned empty string")
	}

	validGreetings := map[string]bool{
		"Good morning":   true,
		"Good afternoon": true,
		"Good evening":   true,
		"Good night":     true,
	}

	if !validGreetings[greeting] {
		t.Errorf("GenerateGreeting() returned unexpected value: %s", greeting)
	}
}

func TestHumanizeEntityType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PROJECT", "Project"},
		{"INCIDENT", "Incident"},
		{"USER", "User"},
		{"COMMENT", "Comment"},
		{"AUDIT_LOG", "Audit Log"},
		// Lowercase variants stored by the audit logging service
		{"project", "Project"},
		{"incident", "Incident"},
		{"user", "User"},
		{"", "Item"},
		{"UNKNOWN", "Unknown"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := HumanizeEntityType(test.input)
			if result != test.expected {
				t.Errorf("HumanizeEntityType(%q) = %q, want %q", test.input, result, test.expected)
			}
		})
	}
}

func TestMapAuditActionToStatus(t *testing.T) {
	tests := []struct {
		input    models.AuditAction
		expected string
	}{
		{models.AuditActionCreate, "success"},
		{models.AuditActionUpdate, "info"},
		{models.AuditActionDelete, "warning"},
		{models.AuditAction("UNKNOWN"), "info"},
	}

	for _, test := range tests {
		t.Run(string(test.input), func(t *testing.T) {
			result := MapAuditActionToStatus(test.input)
			if result != test.expected {
				t.Errorf("MapAuditActionToStatus(%q) = %q, want %q", test.input, result, test.expected)
			}
		})
	}
}

func TestGenerateActivityTitleAndDescription(t *testing.T) {
	tests := []struct {
		name          string
		log           models.AuditLog
		expectedTitle string
		expectedDesc  string
	}{
		{
			name: "Create Project",
			log: models.AuditLog{
				EntityType: "PROJECT",
				Action:     models.AuditActionCreate,
				FieldName:  "",
				OldValue:   "",
				NewValue:   "",
			},
			expectedTitle: "Project Created",
			expectedDesc:  "Project created.",
		},
		{
			name: "Update Incident with field",
			log: models.AuditLog{
				EntityType: "INCIDENT",
				Action:     models.AuditActionUpdate,
				FieldName:  "status",
				OldValue:   "OPEN",
				NewValue:   "RESOLVED",
			},
			expectedTitle: "Incident Updated",
			expectedDesc:  "Incident status was updated.",
		},
		{
			name: "Update Incident without field",
			log: models.AuditLog{
				EntityType: "INCIDENT",
				Action:     models.AuditActionUpdate,
				FieldName:  "",
				OldValue:   "",
				NewValue:   "",
			},
			expectedTitle: "Incident Updated",
			expectedDesc:  "Incident updated.",
		},
		{
			name: "Delete User",
			log: models.AuditLog{
				EntityType: "USER",
				Action:     models.AuditActionDelete,
				FieldName:  "",
				OldValue:   "",
				NewValue:   "",
			},
			expectedTitle: "User Deleted",
			expectedDesc:  "User deleted.",
		},
		{
			name: "Unknown Action",
			log: models.AuditLog{
				EntityType: "PROJECT",
				Action:     models.AuditAction("UNKNOWN"),
				FieldName:  "",
				OldValue:   "",
				NewValue:   "",
			},
			expectedTitle: "Project Modified",
			expectedDesc:  "Project was modified.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			title, desc := GenerateActivityTitleAndDescription(test.log)
			if title != test.expectedTitle {
				t.Errorf("GenerateActivityTitleAndDescription() title = %q, want %q", title, test.expectedTitle)
			}
			if desc != test.expectedDesc {
				t.Errorf("GenerateActivityTitleAndDescription() description = %q, want %q", desc, test.expectedDesc)
			}
		})
	}
}

func TestMapIncidentsForDashboard(t *testing.T) {
	incidents := []models.Incident{
		{
			ID:        1,
			Title:     "API Down",
			Severity:  "P1",
			Status:    "OPEN",
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			Title:     "Database Slow",
			Severity:  "P2",
			Status:    "IN_PROGRESS",
			CreatedAt: time.Now(),
		},
	}

	result := MapIncidentsForDashboard(incidents)

	if len(result) != len(incidents) {
		t.Errorf("MapIncidentsForDashboard() returned %d items, want %d", len(result), len(incidents))
	}

	for i, incident := range result {
		if incident.ID != incidents[i].ID {
			t.Errorf("MapIncidentsForDashboard()[%d].ID = %d, want %d", i, incident.ID, incidents[i].ID)
		}
		if incident.Title != incidents[i].Title {
			t.Errorf("MapIncidentsForDashboard()[%d].Title = %q, want %q", i, incident.Title, incidents[i].Title)
		}
		if incident.Severity != incidents[i].Severity {
			t.Errorf("MapIncidentsForDashboard()[%d].Severity = %q, want %q", i, incident.Severity, incidents[i].Severity)
		}
		if incident.Status != incidents[i].Status {
			t.Errorf("MapIncidentsForDashboard()[%d].Status = %q, want %q", i, incident.Status, incidents[i].Status)
		}
	}
}

func TestMapIncidentsForDashboardEmpty(t *testing.T) {
	incidents := []models.Incident{}
	result := MapIncidentsForDashboard(incidents)

	if len(result) != 0 {
		t.Errorf("MapIncidentsForDashboard(empty) returned %d items, want 0", len(result))
	}

	// Ensure result is not nil (even for empty input)
	if result == nil {
		t.Error("MapIncidentsForDashboard(empty) returned nil, want empty slice")
	}
}

func TestMapAuditLogs(t *testing.T) {
	logs := []models.AuditLog{
		{
			ID:         1,
			EntityType: "PROJECT",
			Action:     models.AuditActionCreate,
			FieldName:  "",
			OldValue:   "",
			NewValue:   "",
			CreatedAt:  time.Now(),
		},
		{
			ID:         2,
			EntityType: "INCIDENT",
			Action:     models.AuditActionDelete,
			FieldName:  "",
			OldValue:   "",
			NewValue:   "",
			CreatedAt:  time.Now(),
		},
	}

	result := MapAuditLogs(logs)

	if len(result) != len(logs) {
		t.Errorf("MapAuditLogs() returned %d items, want %d", len(result), len(logs))
	}

	for i, activity := range result {
		if activity.ID != logs[i].ID {
			t.Errorf("MapAuditLogs()[%d].ID = %d, want %d", i, activity.ID, logs[i].ID)
		}
		if activity.Type != logs[i].EntityType {
			t.Errorf("MapAuditLogs()[%d].Type = %q, want %q", i, activity.Type, logs[i].EntityType)
		}
		if activity.Status == "" {
			t.Errorf("MapAuditLogs()[%d].Status is empty", i)
		}
		if activity.Title == "" {
			t.Errorf("MapAuditLogs()[%d].Title is empty", i)
		}
		if activity.Description == "" {
			t.Errorf("MapAuditLogs()[%d].Description is empty", i)
		}
	}
}

func TestMapAuditLogsEmpty(t *testing.T) {
	logs := []models.AuditLog{}
	result := MapAuditLogs(logs)

	if len(result) != 0 {
		t.Errorf("MapAuditLogs(empty) returned %d items, want 0", len(result))
	}

	// Ensure result is not nil (even for empty input)
	if result == nil {
		t.Error("MapAuditLogs(empty) returned nil, want empty slice")
	}
}

func TestHumanizeEntityTypeEdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "Item"},
		{"A", "A"},
		{"PROJECT", "Project"},
		{"project", "Project"},   // lowercase variant
		{"incident", "Incident"}, // lowercase variant
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := HumanizeEntityType(test.input)
			if result != test.expected {
				t.Errorf("HumanizeEntityType(%q) = %q, want %q", test.input, result, test.expected)
			}
		})
	}
}
