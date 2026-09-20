package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

func TestCorrelatePrometheusSamplesToResource_SucceedsWithNamespaceAndPod(t *testing.T) {
	projectID := uuid.New()
	orgID := uuid.New()
	resourceID := uuid.New()
	resources := []models.Resource{{
		ID:             resourceID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Kind:           constants.ResourceKindPod,
		Name:           "checkout-7d6f5b9d8f-k7j7t",
		Namespace:      "checkout",
		Cluster:        "prod-cluster",
		Labels:         []byte(`{"pod":"checkout-7d6f5b9d8f-k7j7t","namespace":"checkout"}`),
	}}

	result := CorrelatePrometheusSamplesToResource([]dto.PrometheusSample{{
		Timestamp: time.Now().UTC(),
		Value:     1,
		Labels: map[string]string{
			"namespace": "checkout",
			"pod":       "checkout-7d6f5b9d8f-k7j7t",
			"cluster":   "prod-cluster",
		},
	}}, resources)

	if !result.Available {
		t.Fatal("expected resource correlation to be available")
	}
	if result.ResourceID == nil || *result.ResourceID != resourceID {
		t.Fatalf("expected resource ID %s, got %#v", resourceID, result.ResourceID)
	}
	if result.ApplicationID != nil {
		t.Fatalf("expected no application correlation from resource-only lookup, got %#v", result.ApplicationID)
	}
}

func TestCorrelatePrometheusSamplesToResource_FailsWhenLabelsAreAmbiguous(t *testing.T) {
	projectID := uuid.New()
	orgID := uuid.New()
	resources := []models.Resource{
		{ID: uuid.New(), OrganizationID: orgID, ProjectID: projectID, Kind: constants.ResourceKindPod, Name: "api-1", Namespace: "api", Cluster: "prod-cluster"},
		{ID: uuid.New(), OrganizationID: orgID, ProjectID: projectID, Kind: constants.ResourceKindPod, Name: "api-2", Namespace: "api", Cluster: "prod-cluster"},
	}

	result := CorrelatePrometheusSamplesToResource([]dto.PrometheusSample{{
		Timestamp: time.Now().UTC(),
		Value:     1,
		Labels: map[string]string{
			"namespace": "api",
			"pod":       "api",
		},
	}}, resources)

	if result.Available {
		t.Fatal("expected ambiguous labels to be rejected")
	}
	if result.Reason == "" {
		t.Fatal("expected a reason when correlation is unavailable")
	}
}

func TestCorrelatePrometheusSamplesToApplication_UsesDeploymentRuntimeIdentity(t *testing.T) {
	projectID := uuid.New()
	orgID := uuid.New()
	appID := uuid.New()
	resourceID := uuid.New()
	app := models.Application{ID: appID, OrganizationID: orgID, ProjectID: projectID, Name: "checkout-service", Slug: "checkout-service"}
	resources := []models.Resource{{
		ID:             resourceID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Kind:           constants.ResourceKindPod,
		Name:           "checkout-7d6f5b9d8f-k7j7t",
		Namespace:      "checkout",
		Cluster:        "prod-cluster",
	}}
	deployments := []models.Deployment{{
		ID:              uuid.New(),
		ApplicationID:   appID,
		ProjectID:       projectID,
		OrganizationID:  orgID,
		Namespace:       "checkout",
		TargetClusterID: uuid.New(),
		CreatedAt:       time.Now().UTC().Add(-10 * time.Minute),
	}}

	result := CorrelatePrometheusSamplesToApplication([]dto.PrometheusSample{{
		Timestamp: time.Now().UTC(),
		Value:     1,
		Labels: map[string]string{
			"namespace": "checkout",
			"pod":       "checkout-7d6f5b9d8f-k7j7t",
			"cluster":   "prod-cluster",
		},
	}}, resources, deployments, []models.Application{app})

	if !result.Available {
		t.Fatal("expected application correlation to be available")
	}
	if result.ApplicationID == nil || *result.ApplicationID != appID {
		t.Fatalf("expected app ID %s, got %#v", appID, result.ApplicationID)
	}
	if result.ApplicationName != "checkout-service" {
		t.Fatalf("expected application name checkout-service, got %q", result.ApplicationName)
	}
}
