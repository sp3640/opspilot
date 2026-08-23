package services

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
)

const (
	minSLOTargetPercentage = 1.0
	maxSLOTargetPercentage = 100.0
	minSLOWindowDays       = 1
	maxSLOWindowDays       = 365
	defaultSLOWindowDays   = 30
	secondsPerDay          = 24 * 60 * 60
)

// downtimeSeverities are the incident severities counted as an outage for
// availability/error-budget purposes. P0 (critical, full outage) and P1
// (major, significant degradation) both represent the service being
// unavailable to at least some users; P2-P4 are lesser degradations that
// are still tracked (they count toward MTTR/MTTA) but are not "downtime".
// This is a documented modeling choice, not an arbitrary cutoff - see
// docs/sre-metrics.md.
var downtimeSeverities = map[string]bool{
	constants.SeverityP0: true,
	constants.SeverityP1: true,
}

// Formula strings returned verbatim to API consumers alongside every
// MetricValue, so the frontend never has to hardcode (and risk drifting
// from) an explanation of how a number was derived. See docs/sre-metrics.md
// for the full methodology and rationale.
const (
	availabilityFormula = "Availability % = (observed window seconds − downtime seconds) / observed window seconds × 100. " +
		"Downtime is the merged duration of P0/P1 incidents (by CreatedAt→ResolvedAt, or CreatedAt→now if still open) overlapping the window."
	errorBudgetFormula = "Error budget remaining % = (total budget − downtime consumed) / total budget × 100, " +
		"where total budget = observed window seconds × (1 − target% / 100)."
	mttrFormula          = "MTTR = average(resolvedAt − createdAt) across every resolved incident in the window."
	mttaFormula          = "MTTA = average(acknowledgedAt − createdAt) across every acknowledged incident in the window."
	sloComplianceFormula = "Compliant when Availability % ≥ the configured SLO target %."
)

// SREMetricsService computes SLO/availability/error-budget/MTTR/MTTA for an
// application entirely from real historical data already present in this
// system (Incident rows scoped to the application via its ApplicationID
// column - see docs/sre-metrics.md). Uptime and Availability are reported
// identically: this backend has no separate synthetic uptime-probe data
// source, so presenting a second, independently-computed "uptime" number
// would not be honest.
type SREMetricsService struct {
	applicationRepo *repository.ApplicationRepository
	incidentRepo    *repository.IncidentRepository
	sloRepo         *repository.ApplicationSLORepository
	auditService    *AuditService
}

func NewSREMetricsService(
	applicationRepo *repository.ApplicationRepository,
	incidentRepo *repository.IncidentRepository,
	sloRepo *repository.ApplicationSLORepository,
) *SREMetricsService {
	return &SREMetricsService{
		applicationRepo: applicationRepo,
		incidentRepo:    incidentRepo,
		sloRepo:         sloRepo,
	}
}

func (s *SREMetricsService) WithAuditService(auditService *AuditService) *SREMetricsService {
	s.auditService = auditService
	return s
}

func (s *SREMetricsService) ConfigureSLO(userID uint, applicationID, organizationID uuid.UUID, req dto.ApplicationSLOConfigRequest) (*dto.ApplicationSLOConfigResponse, error) {
	if _, err := s.getOwnedApplication(applicationID, organizationID); err != nil {
		return nil, err
	}
	if req.TargetPercentage < minSLOTargetPercentage || req.TargetPercentage > maxSLOTargetPercentage {
		return nil, apperrors.ErrInvalidSLOTarget
	}
	if req.WindowDays < minSLOWindowDays || req.WindowDays > maxSLOWindowDays {
		return nil, apperrors.ErrInvalidSLOWindow
	}

	existing, err := s.sloRepo.GetByApplicationID(applicationID, organizationID)
	isCreate := errors.Is(err, gorm.ErrRecordNotFound)
	if err != nil && !isCreate {
		return nil, err
	}

	var previousTarget float64
	var previousWindow int
	slo := existing
	if isCreate {
		slo = &models.ApplicationSLO{OrganizationID: organizationID, ApplicationID: applicationID, CreatedBy: userID}
	} else {
		previousTarget = existing.TargetPercentage
		previousWindow = existing.WindowDays
	}
	slo.TargetPercentage = req.TargetPercentage
	slo.WindowDays = req.WindowDays

	if err := s.sloRepo.Upsert(slo); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		action := models.AuditActionCreate
		if !isCreate {
			action = models.AuditActionUpdate
		}
		_ = s.auditService.LogEvent(AuditEventInput{
			UserID:         userID,
			OrganizationID: organizationID,
			ApplicationID:  &applicationID,
			EntityType:     "application_slo",
			EntityID:       slo.ID.String(),
			Action:         action,
			BeforeState:    marshalAuditState(map[string]any{"targetPercentage": previousTarget, "windowDays": previousWindow}),
			AfterState:     marshalAuditState(map[string]any{"targetPercentage": slo.TargetPercentage, "windowDays": slo.WindowDays}),
		})
	}

	return mapSLOConfig(slo), nil
}

func (s *SREMetricsService) GetSLOConfig(applicationID, organizationID uuid.UUID) (*dto.ApplicationSLOConfigResponse, error) {
	if _, err := s.getOwnedApplication(applicationID, organizationID); err != nil {
		return nil, err
	}
	slo, err := s.sloRepo.GetByApplicationID(applicationID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrSLONotConfigured
		}
		return nil, err
	}
	return mapSLOConfig(slo), nil
}

// GetMetrics computes SLO/availability/error-budget/MTTR/MTTA for an
// application. Every value is derived from real Incident rows scoped to
// this application; anything this method cannot honestly compute reports
// Available=false with a Reason, never an invented number. See the formula
// constants above and docs/sre-metrics.md for the complete methodology.
func (s *SREMetricsService) GetMetrics(applicationID, organizationID uuid.UUID) (*dto.ApplicationSREMetricsResponse, error) {
	application, err := s.getOwnedApplication(applicationID, organizationID)
	if err != nil {
		return nil, err
	}

	sloConfig, err := s.sloRepo.GetByApplicationID(applicationID, organizationID)
	sloConfigured := true
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		sloConfigured = false
	}

	windowDays := defaultSLOWindowDays
	if sloConfigured {
		windowDays = sloConfig.WindowDays
	}

	now := time.Now().UTC()
	windowStart := now.AddDate(0, 0, -windowDays)
	// Never claim more history than the application actually has.
	if application.CreatedAt.After(windowStart) {
		windowStart = application.CreatedAt
	}
	observedSeconds := now.Sub(windowStart).Seconds()

	response := &dto.ApplicationSREMetricsResponse{
		ApplicationID: applicationID.String(),
		WindowDays:    windowDays,
		WindowStart:   windowStart,
		WindowEnd:     now,
		ObservedDays:  round2(observedSeconds / secondsPerDay),
		SLOConfigured: sloConfigured,
	}

	if observedSeconds <= 0 {
		insufficient := insufficientMetric("Application has no observation history yet", availabilityFormula)
		response.Availability = insufficient
		response.Uptime = insufficient
		response.MTTR = insufficientMetric("No observation history yet", mttrFormula)
		response.MTTA = insufficientMetric("No observation history yet", mttaFormula)
		response.ErrorBudget = insufficientMetric("No observation history yet", errorBudgetFormula)
		response.SLOCompliance = insufficientMetric("No observation history yet", sloComplianceFormula)
		return response, nil
	}

	incidents, err := s.incidentRepo.ListByApplicationForWindow(applicationID, organizationID, windowStart)
	if err != nil {
		return nil, err
	}
	response.IncidentCount = len(incidents)

	downtimeSeconds := computeDowntimeSeconds(incidents, windowStart, now)
	availabilityPercent := (observedSeconds - downtimeSeconds) / observedSeconds * 100
	if availabilityPercent < 0 {
		availabilityPercent = 0
	}

	availabilityMetric := dto.MetricValue{
		Available: true,
		Value:     round2(availabilityPercent),
		Unit:      "%",
		Formatted: fmt.Sprintf("%.2f%%", availabilityPercent),
		Formula:   availabilityFormula,
	}
	response.Availability = availabilityMetric
	response.Uptime = availabilityMetric

	response.MTTR = computeMTTR(incidents)
	response.MTTA = computeMTTA(incidents)

	if sloConfigured {
		target := sloConfig.TargetPercentage
		response.CurrentSLO = &target

		totalBudgetSeconds := observedSeconds * (1 - target/100)
		if totalBudgetSeconds <= 0 {
			response.ErrorBudget = insufficientMetric("A 100% SLO target leaves no error budget to track", errorBudgetFormula)
		} else {
			remainingPercent := (totalBudgetSeconds - downtimeSeconds) / totalBudgetSeconds * 100
			response.ErrorBudget = dto.MetricValue{
				Available: true,
				Value:     round2(remainingPercent),
				Unit:      "%",
				Formatted: fmt.Sprintf("%.2f%%", remainingPercent),
				Formula:   errorBudgetFormula,
			}
		}

		label := "Breached"
		complianceValue := 0.0
		if availabilityPercent >= target {
			label = "Compliant"
			complianceValue = 1
		}
		response.SLOCompliance = dto.MetricValue{Available: true, Value: complianceValue, Formatted: label, Formula: sloComplianceFormula}
	} else {
		response.ErrorBudget = insufficientMetric("SLO not configured for this application", errorBudgetFormula)
		response.SLOCompliance = insufficientMetric("SLO not configured for this application", sloComplianceFormula)
	}

	return response, nil
}

type downtimeInterval struct {
	start, end time.Time
}

// computeDowntimeSeconds merges overlapping downtime-severity incident
// intervals (so two simultaneous P0/P1 incidents never double-count the
// same downtime) and sums the merged duration, clipped to [windowStart, now].
func computeDowntimeSeconds(incidents []models.Incident, windowStart, now time.Time) float64 {
	intervals := make([]downtimeInterval, 0, len(incidents))
	for _, incident := range incidents {
		if !downtimeSeverities[incident.Severity] {
			continue
		}

		start := incident.CreatedAt
		if start.Before(windowStart) {
			start = windowStart
		}
		end := now
		if incident.ResolvedAt != nil && incident.ResolvedAt.Before(now) {
			end = *incident.ResolvedAt
		}
		if !end.After(start) {
			continue
		}
		intervals = append(intervals, downtimeInterval{start: start, end: end})
	}

	if len(intervals) == 0 {
		return 0
	}

	sort.Slice(intervals, func(i, j int) bool { return intervals[i].start.Before(intervals[j].start) })

	merged := []downtimeInterval{intervals[0]}
	for _, iv := range intervals[1:] {
		last := &merged[len(merged)-1]
		if iv.start.After(last.end) {
			merged = append(merged, iv)
			continue
		}
		if iv.end.After(last.end) {
			last.end = iv.end
		}
	}

	var total float64
	for _, iv := range merged {
		total += iv.end.Sub(iv.start).Seconds()
	}
	return total
}

func computeMTTR(incidents []models.Incident) dto.MetricValue {
	var total float64
	var count int
	for _, incident := range incidents {
		if incident.ResolvedAt == nil {
			continue
		}
		total += incident.ResolvedAt.Sub(incident.CreatedAt).Minutes()
		count++
	}
	if count == 0 {
		return insufficientMetric("No resolved incidents in the selected window", mttrFormula)
	}
	avg := total / float64(count)
	return dto.MetricValue{Available: true, Value: round2(avg), Unit: "minutes", Formatted: formatMinutes(avg), Formula: mttrFormula}
}

func computeMTTA(incidents []models.Incident) dto.MetricValue {
	var total float64
	var count int
	for _, incident := range incidents {
		if incident.AcknowledgedAt == nil {
			continue
		}
		total += incident.AcknowledgedAt.Sub(incident.CreatedAt).Minutes()
		count++
	}
	if count == 0 {
		return insufficientMetric("No acknowledged incidents in the selected window", mttaFormula)
	}
	avg := total / float64(count)
	return dto.MetricValue{Available: true, Value: round2(avg), Unit: "minutes", Formatted: formatMinutes(avg), Formula: mttaFormula}
}

func insufficientMetric(reason, formula string) dto.MetricValue {
	return dto.MetricValue{Available: false, Formatted: "Insufficient data", Reason: reason, Formula: formula}
}

func formatMinutes(minutes float64) string {
	if minutes < 60 {
		return fmt.Sprintf("%.0fm", minutes)
	}
	hours := int(minutes / 60)
	remainder := minutes - float64(hours*60)
	if hours < 24 {
		return fmt.Sprintf("%dh %.0fm", hours, remainder)
	}
	days := hours / 24
	remainingHours := hours % 24
	return fmt.Sprintf("%dd %dh", days, remainingHours)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func mapSLOConfig(slo *models.ApplicationSLO) *dto.ApplicationSLOConfigResponse {
	return &dto.ApplicationSLOConfigResponse{
		ApplicationID:    slo.ApplicationID.String(),
		TargetPercentage: slo.TargetPercentage,
		WindowDays:       slo.WindowDays,
		CreatedAt:        slo.CreatedAt,
		UpdatedAt:        slo.UpdatedAt,
	}
}

func (s *SREMetricsService) getOwnedApplication(id, organizationID uuid.UUID) (*models.Application, error) {
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
