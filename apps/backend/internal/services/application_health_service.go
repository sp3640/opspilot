package services

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sp3640/opspilot/backend/internal/alerting"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/health"
	k8sdeployments "github.com/sp3640/opspilot/backend/internal/kubernetes/deployments"
	k8spods "github.com/sp3640/opspilot/backend/internal/kubernetes/pods"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
)

// activeAlertStatuses/activeIncidentStatuses mirror the frontend's existing
// (best-effort, project-scoped) approximation from Phase 18's
// ApplicationHealth component, now computed precisely server-side.
var activeAlertStatuses = map[string]bool{
	constants.AlertStatusOpen:          true,
	constants.AlertStatusAcknowledged:  true,
	constants.AlertStatusInvestigating: true,
}

var activeIncidentStatuses = map[string]bool{
	constants.StatusOpen:          true,
	constants.StatusInvestigating: true,
	constants.StatusMitigating:    true,
}

// ApplicationHealthService assembles the real, already-available signals for
// one application and hands them to health.Evaluate for a deterministic
// score. It never computes a score itself - every scoring rule lives in
// internal/health, which is unit-tested independently of any I/O. The two
// Kubernetes-backed dependencies (pods, runtime deployments) are optional
// (via With...()): when absent, the corresponding factors simply become
// Unknown rather than the service failing.
type ApplicationHealthService struct {
	applicationRepo    *repository.ApplicationRepository
	deploymentRepo     *repository.DeploymentRepository
	alertRepo          *repository.AlertRepository
	incidentRepo       *repository.IncidentRepository
	podService         *k8spods.PodService
	runtimeDeployments *k8sdeployments.DeploymentRuntimeService
}

func NewApplicationHealthService(
	applicationRepo *repository.ApplicationRepository,
	deploymentRepo *repository.DeploymentRepository,
	alertRepo *repository.AlertRepository,
	incidentRepo *repository.IncidentRepository,
) *ApplicationHealthService {
	return &ApplicationHealthService{
		applicationRepo: applicationRepo,
		deploymentRepo:  deploymentRepo,
		alertRepo:       alertRepo,
		incidentRepo:    incidentRepo,
	}
}

func (s *ApplicationHealthService) WithPodService(podService *k8spods.PodService) *ApplicationHealthService {
	s.podService = podService
	return s
}

func (s *ApplicationHealthService) WithRuntimeDeploymentService(runtimeDeployments *k8sdeployments.DeploymentRuntimeService) *ApplicationHealthService {
	s.runtimeDeployments = runtimeDeployments
	return s
}

// GetApplicationHealth gathers every real signal available for the
// application right now and evaluates it deterministically via
// health.Evaluate. A signal this method could not obtain is simply omitted
// from the health.Input it builds - it never substitutes a guess.
func (s *ApplicationHealthService) GetApplicationHealth(ctx context.Context, applicationID, organizationID uuid.UUID) (*dto.ApplicationHealthResponse, error) {
	application, err := s.getOwnedApplication(applicationID, organizationID)
	if err != nil {
		return nil, err
	}

	input := health.Input{}

	namespace := ""
	deployment, err := s.deploymentRepo.GetLatestDeployment(applicationID, organizationID)
	switch {
	case err == nil:
		status := deployment.Status
		input.DeploymentStatus = &status
		namespace = deployment.Namespace
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Never deployed - DeploymentStatus stays nil, which health.Evaluate
		// reports as Unknown, not a fabricated status.
	default:
		return nil, err
	}

	if namespace != "" && s.runtimeDeployments != nil {
		if runtimeStatus := s.resolveRuntimeAvailability(ctx, applicationID, organizationID, namespace); runtimeStatus != nil {
			input.RuntimeDeploymentStatus = runtimeStatus
		}
	}

	if namespace != "" && s.podService != nil {
		if pods, ok := s.resolvePods(ctx, applicationID, organizationID, namespace); ok {
			input.PodsAvailable = true
			input.Pods = pods
		}
	}

	if alertCount, ok := s.countActiveAlerts(application.ProjectID, organizationID, applicationID); ok {
		input.ActiveAlertCount = &alertCount
	}

	if incidentCount, ok := s.countActiveIncidents(applicationID, organizationID); ok {
		input.ActiveIncidentCount = &incidentCount
	}

	result := health.Evaluate(input)
	return mapApplicationHealth(applicationID, result), nil
}

// resolveRuntimeAvailability finds every live Kubernetes Deployment object
// belonging to this application in its deployed namespace (the runtime
// service already scopes this precisely via a real label selector, not a
// name guess) and returns the single worst status among them. Returns nil
// (Unknown) on any error or when no live Deployment object exists yet.
func (s *ApplicationHealthService) resolveRuntimeAvailability(ctx context.Context, applicationID, organizationID uuid.UUID, namespace string) *string {
	list, err := s.runtimeDeployments.ListDeploymentsByApplication(ctx, applicationID, organizationID, namespace)
	if err != nil || list == nil || len(list.Items) == 0 {
		return nil
	}

	worst := list.Items[0].Status
	for _, item := range list.Items[1:] {
		if runtimeStatusRank(item.Status) > runtimeStatusRank(worst) {
			worst = item.Status
		}
	}

	return &worst
}

// runtimeStatusRank orders runtime deployment statuses from best to worst so
// resolveRuntimeAvailability can deterministically pick the worst one when
// an application has more than one live Deployment object in a namespace.
func runtimeStatusRank(status string) int {
	switch status {
	case "Healthy":
		return 0
	case "Progressing", "Scaling":
		return 1
	case "Unavailable", "Failed":
		return 2
	default:
		return 3 // Unknown/unrecognized ranks worst so it is never hidden by a healthy sibling
	}
}

func (s *ApplicationHealthService) resolvePods(ctx context.Context, applicationID, organizationID uuid.UUID, namespace string) ([]health.PodInput, bool) {
	list, err := s.podService.ListPodsByApplication(ctx, applicationID, organizationID, namespace)
	if err != nil || list == nil {
		return nil, false
	}

	pods := make([]health.PodInput, 0, len(list.Items))
	for _, pod := range list.Items {
		pods = append(pods, health.PodInput{Ready: pod.Ready, RestartCount: pod.RestartCount})
	}

	return pods, true
}

// countActiveAlerts counts alerts attributable to this application via the
// same applicationId metadata key the alerting engine itself writes
// (alerting.ConditionMetadata) - Alert has no applicationId column, so this
// is the only real (non-fabricated) way to attribute an alert to an
// application. Alerts without that metadata (manual/third-party alerts)
// cannot be attributed and are excluded, not guessed at.
func (s *ApplicationHealthService) countActiveAlerts(projectID, organizationID, applicationID uuid.UUID) (int, bool) {
	alerts, err := s.alertRepo.ListByProject(projectID, organizationID)
	if err != nil {
		return 0, false
	}

	count := 0
	target := applicationID.String()
	for _, alert := range alerts {
		if !activeAlertStatuses[alert.Status] {
			continue
		}

		var metadata alerting.ConditionMetadata
		if err := json.Unmarshal(alert.Metadata, &metadata); err != nil {
			continue
		}
		if metadata.ApplicationID == target {
			count++
		}
	}

	return count, true
}

// countActiveIncidents counts incidents scoped to this application via its
// real ApplicationID column - a precise match, not a best-effort one.
func (s *ApplicationHealthService) countActiveIncidents(applicationID, organizationID uuid.UUID) (int, bool) {
	incidents, err := s.incidentRepo.ListByApplicationID(applicationID, organizationID)
	if err != nil {
		return 0, false
	}

	count := 0
	for _, incident := range incidents {
		if activeIncidentStatuses[incident.Status] {
			count++
		}
	}

	return count, true
}

func (s *ApplicationHealthService) getOwnedApplication(id, organizationID uuid.UUID) (*models.Application, error) {
	application, err := s.applicationRepo.GetApplication(id, organizationID)
	if err == nil {
		return application, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if existing, anyErr := s.applicationRepo.FindByIDAnyOrganization(id); anyErr == nil && existing != nil {
			return nil, apperrors.ErrApplicationForbidden
		}
		return nil, apperrors.ErrApplicationNotFound
	}

	return nil, err
}

func mapApplicationHealth(applicationID uuid.UUID, result health.Result) *dto.ApplicationHealthResponse {
	factors := make([]dto.ApplicationHealthFactorResponse, 0, len(result.Factors))
	for _, factor := range result.Factors {
		factors = append(factors, dto.ApplicationHealthFactorResponse{
			Key:    string(factor.Key),
			Label:  factor.Label,
			State:  string(factor.State),
			Score:  factor.Score,
			Weight: factor.Weight,
			Reason: factor.Reason,
		})
	}

	return &dto.ApplicationHealthResponse{
		ApplicationID: applicationID.String(),
		Score:         result.Score,
		State:         string(result.OverallState),
		Factors:       factors,
	}
}
