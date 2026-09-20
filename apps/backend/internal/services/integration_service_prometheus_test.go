package services

import (
	"testing"

	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/repository"
)

func TestCorrelatePrometheusSamplesUnavailableWithoutRepos(t *testing.T) {
	s := &IntegrationService{}

	result := s.CorrelatePrometheusSamples(uuid.New(), []dto.PrometheusSample{{Value: 1}})

	if result.Available {
		t.Fatalf("expected correlation to be unavailable when no repos are wired, got %+v", result)
	}
	if result.Reason == "" {
		t.Fatalf("expected a human-readable reason when correlation is unavailable")
	}
	if result.Resource != nil || result.Application != nil {
		t.Fatalf("expected no resource/application when correlation is unavailable, got %+v", result)
	}
}

func TestCorrelatePrometheusSamplesUnavailableWithoutSamples(t *testing.T) {
	// Non-nil but unusable repos: with an empty sample set,
	// CorrelatePrometheusSamples must short-circuit before ever touching
	// them, so a zero-value (DB-less) repo is safe to pass here.
	s := &IntegrationService{
		resourceRepo:    &repository.ResourceRepository{},
		deploymentRepo:  &repository.DeploymentRepository{},
		applicationRepo: &repository.ApplicationRepository{},
	}

	result := s.CorrelatePrometheusSamples(uuid.New(), nil)

	if result.Available {
		t.Fatalf("expected correlation to be unavailable for an empty sample set, got %+v", result)
	}
}
