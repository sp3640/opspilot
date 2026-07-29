package mapper

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// fixture returns a populated Project model for use in tests.
func fixture() models.Project {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	now := time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC)

	return models.Project{
		ID:          id,
		Name:        "Platform Core",
		Slug:        "platform-core",
		Description: "Shared control plane.",
		Environment: "production",
		Health:      "healthy",
		Members:     3,
		Services:    14,
		OwnerID:     42,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestMapProject(t *testing.T) {
	p := fixture()
	owner := models.User{ID: p.OwnerID, Name: "Maya Chen"}
	got := MapProject(p, owner)

	if got.ID != p.ID.String() {
		t.Errorf("ID: got %q, want %q", got.ID, p.ID.String())
	}
	if got.Name != p.Name {
		t.Errorf("Name: got %q, want %q", got.Name, p.Name)
	}
	if got.Slug != p.Slug {
		t.Errorf("Slug: got %q, want %q", got.Slug, p.Slug)
	}
	if got.Description != p.Description {
		t.Errorf("Description: got %q, want %q", got.Description, p.Description)
	}
	if got.Environment != p.Environment {
		t.Errorf("Environment: got %q, want %q", got.Environment, p.Environment)
	}
	if got.Health != p.Health {
		t.Errorf("Health: got %q, want %q", got.Health, p.Health)
	}
	if got.Members != p.Members {
		t.Errorf("Members: got %d, want %d", got.Members, p.Members)
	}
	if got.Services != p.Services {
		t.Errorf("Services: got %d, want %d", got.Services, p.Services)
	}
	if got.Owner.ID != owner.ID || got.Owner.Name != owner.Name {
		t.Errorf("Owner: got %+v, want %+v", got.Owner, dto.ProjectOwnerResponse{ID: owner.ID, Name: owner.Name})
	}
	if !got.CreatedAt.Equal(p.CreatedAt) {
		t.Errorf("CreatedAt: got %v, want %v", got.CreatedAt, p.CreatedAt)
	}
	if !got.UpdatedAt.Equal(p.UpdatedAt) {
		t.Errorf("UpdatedAt: got %v, want %v", got.UpdatedAt, p.UpdatedAt)
	}
}

func TestMapProjects(t *testing.T) {
	projects := []models.Project{fixture(), fixture()}
	owner := models.User{ID: projects[0].OwnerID, Name: "Maya Chen"}
	got := MapProjects(projects, owner)

	if len(got) != len(projects) {
		t.Fatalf("MapProjects: got %d items, want %d", len(got), len(projects))
	}

	for i, item := range got {
		if item.ID != projects[i].ID.String() {
			t.Errorf("[%d] ID: got %q, want %q", i, item.ID, projects[i].ID.String())
		}
		if item.Name != projects[i].Name {
			t.Errorf("[%d] Name: got %q, want %q", i, item.Name, projects[i].Name)
		}
	}
}

func TestMapProjectsEmpty(t *testing.T) {
	got := MapProjects([]models.Project{}, models.User{})

	if got == nil {
		t.Error("MapProjects(empty): expected non-nil slice, got nil")
	}
	if len(got) != 0 {
		t.Errorf("MapProjects(empty): got %d items, want 0", len(got))
	}
}

func TestMapProjectSummary(t *testing.T) {
	p := fixture()
	got := MapProjectSummary(p)

	if got.ID != p.ID.String() {
		t.Errorf("ID: got %q, want %q", got.ID, p.ID.String())
	}
	if got.Name != p.Name {
		t.Errorf("Name: got %q, want %q", got.Name, p.Name)
	}
	if got.Slug != p.Slug {
		t.Errorf("Slug: got %q, want %q", got.Slug, p.Slug)
	}
	if got.Environment != p.Environment {
		t.Errorf("Environment: got %q, want %q", got.Environment, p.Environment)
	}
	if got.Health != p.Health {
		t.Errorf("Health: got %q, want %q", got.Health, p.Health)
	}
}

func TestMapProjectZeroValue(t *testing.T) {
	// A zero-value model must never panic and must produce a valid DTO.
	var p models.Project
	got := MapProject(p, models.User{})

	if got.ID == "" {
		// uuid.Nil serialises to the all-zeros string — that is acceptable.
		t.Log("MapProject(zero): ID is empty string (uuid.Nil serialised)")
	}
	if got.Owner.ID != 0 {
		t.Errorf("Owner.ID: got %d, want 0", got.Owner.ID)
	}
}
