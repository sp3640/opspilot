package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/dto"
	k8sevents "github.com/sp3640/opspilot/backend/internal/kubernetes/events"
	k8slogs "github.com/sp3640/opspilot/backend/internal/kubernetes/logs"
	k8spods "github.com/sp3640/opspilot/backend/internal/kubernetes/pods"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/rca"
	"github.com/sp3640/opspilot/backend/internal/repository"
)

// rcaLookbackWindow is how far before the incident's creation time evidence
// gathering looks - long enough to catch a deployment or config change that
// only manifested as an incident several minutes later, short enough that
// unrelated history doesn't get pulled in.
const rcaLookbackWindow = 60 * time.Minute

const (
	rcaMaxDeployments   = 10
	rcaMaxConfigChanges = 100
	rcaMaxAlerts        = 100
	rcaMaxMetricPoints  = 500
	rcaMaxPodsForLogs   = 3
	rcaLogTailLines     = int64(200)
)

// RCAService assembles the real, already-available evidence for one
// incident - alerts, metrics, logs, deployments, Kubernetes events, pod
// health, and audit/config-change history - and hands it to rca.Evaluate
// for a deterministic correlation. It never computes a possible cause
// itself; every correlation rule lives in internal/rca, which is
// unit-tested independently of any I/O. The three Kubernetes-backed
// dependencies (pods, events, logs) are optional (via With...()): when
// absent, or when the live cluster call fails, the corresponding evidence
// category simply becomes unavailable rather than failing the analysis.
type RCAService struct {
	incidentRepo    *repository.IncidentRepository
	alertRepo       *repository.AlertRepository
	metricRepo      *repository.MetricRepository
	deploymentRepo  *repository.DeploymentRepository
	auditRepo       *repository.AuditRepository
	applicationRepo *repository.ApplicationRepository

	podService   *k8spods.PodService
	eventService *k8sevents.EventService
	logService   *k8slogs.LogService
}

func NewRCAService(
	incidentRepo *repository.IncidentRepository,
	alertRepo *repository.AlertRepository,
	metricRepo *repository.MetricRepository,
	deploymentRepo *repository.DeploymentRepository,
	auditRepo *repository.AuditRepository,
	applicationRepo *repository.ApplicationRepository,
) *RCAService {
	return &RCAService{
		incidentRepo:    incidentRepo,
		alertRepo:       alertRepo,
		metricRepo:      metricRepo,
		deploymentRepo:  deploymentRepo,
		auditRepo:       auditRepo,
		applicationRepo: applicationRepo,
	}
}

func (s *RCAService) WithPodService(podService *k8spods.PodService) *RCAService {
	s.podService = podService
	return s
}

func (s *RCAService) WithEventService(eventService *k8sevents.EventService) *RCAService {
	s.eventService = eventService
	return s
}

func (s *RCAService) WithLogService(logService *k8slogs.LogService) *RCAService {
	s.logService = logService
	return s
}

// Analyze gathers every real signal available for incidentID right now and
// evaluates it deterministically via rca.Evaluate. A signal this method
// could not obtain is simply reported as unavailable in the rca.Input it
// builds - it never substitutes a guess.
func (s *RCAService) Analyze(ctx context.Context, incidentID uint, organizationID uuid.UUID) (*dto.IncidentRCAResponse, error) {
	incident, err := s.incidentRepo.GetByIDAndOrganizationID(incidentID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrIncidentNotFound
		}
		return nil, err
	}

	windowStart := incident.CreatedAt.Add(-rcaLookbackWindow)
	windowEnd := time.Now().UTC()
	if incident.ResolvedAt != nil {
		windowEnd = *incident.ResolvedAt
	}

	input := rca.Input{
		IncidentID:             incident.ID,
		IncidentTitle:          incident.Title,
		IncidentDescription:    incident.Description,
		IncidentSeverity:       incident.Severity,
		IncidentStatus:         incident.Status,
		IncidentCreatedAt:      incident.CreatedAt,
		IncidentAcknowledgedAt: incident.AcknowledgedAt,
		IncidentResolvedAt:     incident.ResolvedAt,
		WindowStart:            windowStart,
		WindowEnd:              windowEnd,
	}

	s.gatherAlerts(&input, incident.ID, organizationID)
	s.gatherMetrics(&input, incident.ProjectID, organizationID, windowStart, windowEnd)
	s.gatherConfigChanges(&input, incident.ProjectID, organizationID, windowStart, windowEnd)

	var applicationIDStr *string
	if incident.ApplicationID != nil {
		if application, appErr := s.applicationRepo.GetApplication(*incident.ApplicationID, organizationID); appErr == nil {
			idStr := application.ID.String()
			applicationIDStr = &idStr
			input.ApplicationAvailable = true
			input.ApplicationID = idStr
			input.ApplicationName = application.Name

			namespace := s.gatherDeployments(&input, application.ID, organizationID)
			if namespace != "" {
				pods := s.gatherPods(ctx, &input, application.ID, organizationID, namespace)
				s.gatherK8sEvents(ctx, &input, application.ID, organizationID, namespace)
				s.gatherLogs(ctx, &input, application.ID, organizationID, namespace, windowStart, pods)
			}
		}
	}

	result := rca.Evaluate(input)
	response := mapRCAResult(incident.ID, applicationIDStr, windowStart, windowEnd, result)
	return response, nil
}

func (s *RCAService) gatherAlerts(input *rca.Input, incidentID uint, organizationID uuid.UUID) {
	req := &models.PaginationRequest{Page: 1, Limit: rcaMaxAlerts, IncidentID: incidentID}
	alerts, _, err := s.alertRepo.List(req, organizationID)
	if err != nil {
		return
	}

	input.AlertsAvailable = true
	input.Alerts = make([]rca.AlertRecord, 0, len(alerts))
	for _, alert := range alerts {
		input.Alerts = append(input.Alerts, rca.AlertRecord{
			ID:              alert.ID,
			Title:           alert.Title,
			Severity:        alert.Severity,
			Status:          alert.Status,
			FirstSeenAt:     alert.FirstSeenAt,
			LastSeenAt:      alert.LastSeenAt,
			OccurrenceCount: alert.OccurrenceCount,
		})
	}
}

func (s *RCAService) gatherMetrics(input *rca.Input, projectID, organizationID uuid.UUID, windowStart, windowEnd time.Time) {
	metrics, err := s.metricRepo.ListByProject(projectID, "", "", &windowStart, &windowEnd, rcaMaxMetricPoints, organizationID)
	if err != nil {
		return
	}

	input.MetricsAvailable = true
	input.Metrics = make([]rca.MetricPoint, 0, len(metrics))
	for _, metric := range metrics {
		input.Metrics = append(input.Metrics, rca.MetricPoint{
			MetricType: metric.MetricType,
			MetricName: metric.MetricName,
			Unit:       metric.Unit,
			Value:      metric.Value,
			Timestamp:  metric.Timestamp,
		})
	}
}

func (s *RCAService) gatherConfigChanges(input *rca.Input, projectID, organizationID uuid.UUID, windowStart, windowEnd time.Time) {
	req := &models.PaginationRequest{Page: 1, Limit: rcaMaxConfigChanges, DateFrom: &windowStart, DateTo: &windowEnd}
	logs, _, err := s.auditRepo.ListByProjectID(req, projectID, organizationID)
	if err != nil {
		return
	}

	input.ConfigChangesAvailable = true
	input.ConfigChanges = make([]rca.ConfigChangeRecord, 0, len(logs))
	for _, log := range logs {
		input.ConfigChanges = append(input.ConfigChanges, rca.ConfigChangeRecord{
			EntityType: log.EntityType,
			EntityID:   log.EntityID,
			Action:     string(log.Action),
			FieldName:  log.FieldName,
			OldValue:   log.OldValue,
			NewValue:   log.NewValue,
			ChangedAt:  log.CreatedAt,
		})
	}
}

// gatherDeployments loads the application's most recent deployments and
// returns the namespace of the newest one (empty if the application has
// never been deployed), which the caller uses to scope live Kubernetes
// evidence lookups.
func (s *RCAService) gatherDeployments(input *rca.Input, applicationID, organizationID uuid.UUID) string {
	req := &models.PaginationRequest{Page: 1, Limit: rcaMaxDeployments}
	deployments, _, err := s.deploymentRepo.ListByApplication(applicationID, organizationID, req)
	if err != nil {
		return ""
	}

	input.DeploymentsAvailable = true
	input.Deployments = make([]rca.DeploymentRecord, 0, len(deployments))
	namespace := ""
	for i, deployment := range deployments {
		if i == 0 {
			namespace = deployment.Namespace
		}
		input.Deployments = append(input.Deployments, rca.DeploymentRecord{
			ID:          deployment.ID.String(),
			ImageTag:    deployment.ImageTag,
			Status:      deployment.Status,
			Environment: deployment.Environment,
			CreatedAt:   deployment.CreatedAt,
		})
	}

	return namespace
}

// gatherPods fetches live pod status for the application's namespace. The
// returned slice (also used to pick log-sampling targets) is nil when pod
// data could not be obtained.
func (s *RCAService) gatherPods(ctx context.Context, input *rca.Input, applicationID, organizationID uuid.UUID, namespace string) []rca.PodStatus {
	if s.podService == nil {
		return nil
	}

	list, err := s.podService.ListPodsByApplication(ctx, applicationID, organizationID, namespace)
	if err != nil || list == nil {
		return nil
	}

	input.PodsAvailable = true
	input.Pods = make([]rca.PodStatus, 0, len(list.Items))
	for _, pod := range list.Items {
		input.Pods = append(input.Pods, rca.PodStatus{
			Name:         pod.Name,
			Ready:        pod.Ready,
			RestartCount: pod.RestartCount,
			Phase:        pod.Phase,
			Reason:       pod.Reason,
		})
	}

	return input.Pods
}

func (s *RCAService) gatherK8sEvents(ctx context.Context, input *rca.Input, applicationID, organizationID uuid.UUID, namespace string) {
	if s.eventService == nil {
		return
	}

	list, err := s.eventService.ListEventsByApplication(ctx, applicationID, organizationID, namespace)
	if err != nil || list == nil {
		return
	}

	input.K8sEventsAvailable = true
	input.K8sEvents = make([]rca.K8sEvent, 0, len(list.Items))
	for _, event := range list.Items {
		input.K8sEvents = append(input.K8sEvents, rca.K8sEvent{
			Reason:         event.Reason,
			Message:        event.Message,
			Type:           event.Type,
			Count:          event.Count,
			LastTimestamp:  event.LastTimestamp,
			InvolvedObject: event.InvolvedObject,
		})
	}
}

// gatherLogs fetches a bounded tail of log lines for up to rcaMaxPodsForLogs
// pods, preferring pods that are unhealthy (not ready or with restarts) -
// the ones most likely to contain the errors that explain the incident -
// over arbitrarily picking the first N pods returned by the cluster.
func (s *RCAService) gatherLogs(ctx context.Context, input *rca.Input, applicationID, organizationID uuid.UUID, namespace string, windowStart time.Time, pods []rca.PodStatus) {
	if s.logService == nil || len(pods) == 0 {
		return
	}

	sinceSeconds := int64(time.Since(windowStart).Seconds())
	if sinceSeconds < 1 {
		sinceSeconds = 1
	}
	tailLines := rcaLogTailLines

	targets := selectPodsForLogSampling(pods, rcaMaxPodsForLogs)

	lines := make([]rca.LogLine, 0)
	fetchedAny := false
	for _, pod := range targets {
		logResponse, err := s.logService.GetPodLogs(ctx, applicationID, organizationID, namespace, pod.Name, "", &tailLines, &sinceSeconds, false, false)
		if err != nil || logResponse == nil {
			continue
		}
		fetchedAny = true
		for _, line := range strings.Split(logResponse.Log, "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			lines = append(lines, rca.LogLine{PodName: pod.Name, Container: logResponse.Container, Line: line})
		}
	}

	if fetchedAny {
		input.LogsAvailable = true
		input.Logs = lines
	}
}

// selectPodsForLogSampling ranks pods by how likely they are to contain
// evidence (not ready, then higher restart count) and returns up to limit,
// preserving the relative order of equally-ranked pods (a stable
// insertion sort over what is always a small slice).
func selectPodsForLogSampling(pods []rca.PodStatus, limit int) []rca.PodStatus {
	ranked := make([]rca.PodStatus, len(pods))
	copy(ranked, pods)

	for i := 1; i < len(ranked); i++ {
		for j := i; j > 0 && podSamplingRank(ranked[j]) > podSamplingRank(ranked[j-1]); j-- {
			ranked[j], ranked[j-1] = ranked[j-1], ranked[j]
		}
	}

	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	return ranked
}

func podSamplingRank(pod rca.PodStatus) int {
	rank := int(pod.RestartCount)
	if !pod.Ready {
		rank += 1000
	}
	return rank
}

func mapRCAResult(incidentID uint, applicationID *string, windowStart, windowEnd time.Time, result rca.Result) *dto.IncidentRCAResponse {
	response := &dto.IncidentRCAResponse{
		IncidentID:                    incidentID,
		Summary:                       result.Summary,
		AffectedApplicationID:         applicationID,
		AffectedApplicationName:       result.AffectedApplication,
		ApplicationKnown:              result.ApplicationKnown,
		WindowStart:                   windowStart,
		WindowEnd:                     windowEnd,
		Timeline:                      make([]dto.RCATimelineEventResponse, 0, len(result.Timeline)),
		RecentChanges:                 make([]dto.RCAConfigChangeResponse, 0, len(result.RecentChanges)),
		CorrelatedAlerts:              make([]dto.RCAAlertResponse, 0, len(result.CorrelatedAlerts)),
		RelevantMetrics:               make([]dto.RCAMetricSignalResponse, 0, len(result.RelevantMetrics)),
		RelevantLogs:                  make([]dto.RCALogSignalResponse, 0, len(result.RelevantLogs)),
		KubernetesEvidence:            make([]dto.RCAKubernetesEventResponse, 0, len(result.KubernetesEvidence)),
		PodEvidence:                   make([]dto.RCAPodResponse, 0, len(result.PodEvidence)),
		PossibleCauses:                make([]dto.RCAPossibleCauseResponse, 0, len(result.PossibleCauses)),
		RecommendedInvestigationSteps: result.RecommendedInvestigationSteps,
		RecommendedRemediation:        result.RecommendedRemediation,
		EvidenceLevel:                 string(result.EvidenceLevel),
		EvidenceLevelReason:           result.EvidenceLevelReason,
		GeneratedAt:                   time.Now().UTC(),
	}

	for _, event := range result.Timeline {
		response.Timeline = append(response.Timeline, dto.RCATimelineEventResponse{
			Timestamp:   event.Timestamp,
			Type:        string(event.Type),
			Title:       event.Title,
			Description: event.Description,
			Source:      event.Source,
		})
	}

	for _, change := range result.RecentChanges {
		response.RecentChanges = append(response.RecentChanges, dto.RCAConfigChangeResponse{
			EntityType: change.EntityType,
			EntityID:   change.EntityID,
			Action:     change.Action,
			FieldName:  change.FieldName,
			OldValue:   change.OldValue,
			NewValue:   change.NewValue,
			ChangedAt:  change.ChangedAt,
		})
	}

	for _, alert := range result.CorrelatedAlerts {
		response.CorrelatedAlerts = append(response.CorrelatedAlerts, dto.RCAAlertResponse{
			ID:              alert.ID,
			Title:           alert.Title,
			Severity:        alert.Severity,
			Status:          alert.Status,
			FirstSeenAt:     alert.FirstSeenAt,
			LastSeenAt:      alert.LastSeenAt,
			OccurrenceCount: alert.OccurrenceCount,
		})
	}

	for _, metric := range result.RelevantMetrics {
		response.RelevantMetrics = append(response.RelevantMetrics, dto.RCAMetricSignalResponse{
			MetricType:    metric.MetricType,
			MetricName:    metric.MetricName,
			Unit:          metric.Unit,
			Before:        metric.Before,
			After:         metric.After,
			ChangePercent: metric.ChangePercent,
			Direction:     metric.Direction,
			SampleCount:   metric.SampleCount,
		})
	}

	for _, log := range result.RelevantLogs {
		response.RelevantLogs = append(response.RelevantLogs, dto.RCALogSignalResponse{
			PodName:   log.PodName,
			Container: log.Container,
			Line:      log.Line,
			Keyword:   log.Keyword,
		})
	}

	for _, event := range result.KubernetesEvidence {
		response.KubernetesEvidence = append(response.KubernetesEvidence, dto.RCAKubernetesEventResponse{
			Reason:         event.Reason,
			Message:        event.Message,
			Type:           event.Type,
			Count:          event.Count,
			LastTimestamp:  event.LastTimestamp,
			InvolvedObject: event.InvolvedObject,
		})
	}

	for _, pod := range result.PodEvidence {
		response.PodEvidence = append(response.PodEvidence, dto.RCAPodResponse{
			Name:         pod.Name,
			Ready:        pod.Ready,
			RestartCount: pod.RestartCount,
			Phase:        pod.Phase,
			Reason:       pod.Reason,
		})
	}

	for _, cause := range result.PossibleCauses {
		response.PossibleCauses = append(response.PossibleCauses, dto.RCAPossibleCauseResponse{
			Title:      cause.Title,
			Category:   cause.Category,
			Confidence: string(cause.Confidence),
			Evidence:   cause.Evidence,
		})
	}

	return response
}
