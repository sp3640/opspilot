package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// PrometheusCorrelationResult describes whether a Prometheus sample set can
// be mapped to a concrete resource/application in a deterministic, non-guessing
// way. It is intentionally conservative: if labels are absent, ambiguous, or
// span multiple identities, Availability is false and a clear reason is given.
type PrometheusCorrelationResult struct {
	Available       bool
	Reason          string
	ResourceID      *uuid.UUID
	ApplicationID   *uuid.UUID
	ApplicationName string
}

// PrometheusResourceCorrelationResult is kept as a compatibility alias for
// callers that prefer a more explicit resource-focused naming.
type PrometheusResourceCorrelationResult = PrometheusCorrelationResult

// PrometheusApplicationCorrelationResult is kept as a compatibility alias for
// callers that prefer an application-focused naming.
type PrometheusApplicationCorrelationResult = PrometheusCorrelationResult

func CorrelatePrometheusSamplesToResource(samples []dto.PrometheusSample, resources []models.Resource) PrometheusCorrelationResult {
	if len(samples) == 0 {
		return unavailable("no Prometheus samples were provided for correlation")
	}

	var matched *uuid.UUID
	for _, sample := range samples {
		candidate, err := matchResourceSample(sample, resources)
		if err != nil {
			return unavailable(err.Error())
		}
		if candidate == nil {
			continue
		}
		if matched != nil && *matched != *candidate {
			return unavailable(fmt.Sprintf("Prometheus labels are ambiguous across multiple resources (%s and %s)", matched.String(), candidate.String()))
		}
		matched = candidate
	}

	if matched == nil {
		return unavailable("no deterministic resource match was found for the Prometheus labels")
	}

	return PrometheusCorrelationResult{Available: true, ResourceID: matched}
}

func CorrelatePrometheusSamplesToApplication(samples []dto.PrometheusSample, resources []models.Resource, deployments []models.Deployment, applications []models.Application) PrometheusCorrelationResult {
	if len(samples) == 0 {
		return unavailable("no Prometheus samples were provided for application correlation")
	}

	resourceMatch := CorrelatePrometheusSamplesToResource(samples, resources)
	if resourceMatch.Available && resourceMatch.ResourceID != nil {
		if appID := resolveApplicationIDFromResource(resources, *resourceMatch.ResourceID); appID != nil {
			for _, app := range applications {
				if app.ID == *appID {
					return PrometheusCorrelationResult{Available: true, ResourceID: resourceMatch.ResourceID, ApplicationID: &app.ID, ApplicationName: app.Name}
				}
			}
		}
		if appID := resolveApplicationIDFromResourceSamples(samples, deployments, applications, resources); appID != nil {
			for _, app := range applications {
				if app.ID == *appID {
					return PrometheusCorrelationResult{Available: true, ResourceID: resourceMatch.ResourceID, ApplicationID: &app.ID, ApplicationName: app.Name}
				}
			}
		}
	}

	appCandidates := make(map[uuid.UUID]bool)
	for _, sample := range samples {
		namespace := normalizeSampleLabel(sample.Labels, "namespace")
		if namespace == "" {
			continue
		}
		for _, deployment := range deployments {
			if !equalFold(deployment.Namespace, namespace) {
				continue
			}
			if deployment.ApplicationID == uuid.Nil {
				continue
			}
			appCandidates[deployment.ApplicationID] = true
		}
	}

	if len(appCandidates) == 1 {
		for appID := range appCandidates {
			for _, app := range applications {
				if app.ID == appID {
					return PrometheusCorrelationResult{Available: true, ApplicationID: &app.ID, ApplicationName: app.Name}
				}
			}
		}
	}
	if len(appCandidates) > 1 {
		return unavailable("Prometheus labels match more than one application namespace mapping")
	}

	if resourceMatch.Available && resourceMatch.ResourceID != nil {
		return unavailable("Prometheus sample identity was matched to a resource, but no deterministic application mapping was found")
	}

	return unavailable("no deterministic application match was found for the Prometheus labels")
}

func matchResourceSample(sample dto.PrometheusSample, resources []models.Resource) (*uuid.UUID, error) {
	if len(resources) == 0 {
		return nil, nil
	}

	candidateIDs := make([]uuid.UUID, 0, 1)
	for _, resource := range resources {
		if resourceMatchesLabelSet(resource, sample.Labels) {
			candidateIDs = append(candidateIDs, resource.ID)
		}
	}
	if len(candidateIDs) > 1 {
		return nil, fmt.Errorf("Prometheus labels match multiple resources (%d matches)", len(candidateIDs))
	}
	if len(candidateIDs) == 1 {
		return &candidateIDs[0], nil
	}
	return nil, nil
}

func resourceMatchesLabelSet(resource models.Resource, labels map[string]string) bool {
	if resource.ID == uuid.Nil {
		return false
	}
	if labels == nil {
		return false
	}

	namespace := normalizeSampleLabel(labels, "namespace")
	cluster := normalizeSampleLabel(labels, "cluster")
	podName := normalizeSampleLabel(labels, "pod")
	if podName == "" {
		podName = normalizeSampleLabel(labels, "pod_name")
	}
	deploymentName := normalizeSampleLabel(labels, "deployment")
	if deploymentName == "" {
		deploymentName = normalizeSampleLabel(labels, "workload")
	}

	if resource.Kind == constants.ResourceKindPod {
		if namespace != "" && !equalFold(resource.Namespace, namespace) {
			return false
		}
		if cluster != "" && !equalFold(resource.Cluster, cluster) {
			return false
		}
		if podName != "" && !equalFold(resource.Name, podName) {
			return false
		}
		if podName == "" && deploymentName != "" && !equalFold(resource.Name, deploymentName) {
			return false
		}
		return true
	}

	if resource.Kind == constants.ResourceKindDeployment {
		if namespace != "" && !equalFold(resource.Namespace, namespace) {
			return false
		}
		if cluster != "" && !equalFold(resource.Cluster, cluster) {
			return false
		}
		if deploymentName != "" && !equalFold(resource.Name, deploymentName) {
			return false
		}
		if deploymentName == "" && podName != "" && !equalFold(resource.Name, podName) {
			return false
		}
		return true
	}

	return false
}

func resolveApplicationIDFromResource(resources []models.Resource, resourceID uuid.UUID) *uuid.UUID {
	for _, resource := range resources {
		if resource.ID != resourceID {
			continue
		}
		if len(resource.Labels) == 0 {
			return nil
		}
		var labels map[string]string
		if err := json.Unmarshal(resource.Labels, &labels); err != nil {
			return nil
		}
		value := strings.TrimSpace(labels["opspilot/application-id"])
		if value == "" {
			return nil
		}
		parsed, err := uuid.Parse(value)
		if err != nil {
			return nil
		}
		return &parsed
	}
	return nil
}

func resolveApplicationIDFromResourceSamples(samples []dto.PrometheusSample, deployments []models.Deployment, applications []models.Application, resources []models.Resource) *uuid.UUID {
	if len(samples) == 0 || len(deployments) == 0 {
		return nil
	}
	for _, sample := range samples {
		namespace := normalizeSampleLabel(sample.Labels, "namespace")
		podName := normalizeSampleLabel(sample.Labels, "pod")
		if podName == "" {
			podName = normalizeSampleLabel(sample.Labels, "pod_name")
		}
		for _, deployment := range deployments {
			if namespace != "" && !equalFold(deployment.Namespace, namespace) {
				continue
			}
			if podName != "" && !matchesDeploymentResourceName(deployment, podName, resources) {
				continue
			}
			if deployment.ApplicationID != uuid.Nil {
				for _, app := range applications {
					if app.ID == deployment.ApplicationID {
						return &app.ID
					}
				}
			}
		}
	}
	return nil
}

func matchesDeploymentResourceName(deployment models.Deployment, podName string, resources []models.Resource) bool {
	if deployment.Namespace == "" || podName == "" {
		return false
	}
	for _, resource := range resources {
		if resource.Kind != constants.ResourceKindPod {
			continue
		}
		if equalFold(resource.Namespace, deployment.Namespace) && equalFold(resource.Name, podName) {
			return true
		}
	}
	return false
}

func normalizeSampleLabel(labels map[string]string, key string) string {
	if labels == nil {
		return ""
	}
	value := strings.TrimSpace(labels[key])
	if value == "" {
		for actualKey, actualValue := range labels {
			if strings.EqualFold(actualKey, key) {
				value = strings.TrimSpace(actualValue)
				break
			}
		}
	}
	return strings.ToLower(value)
}

func equalFold(left, right string) bool {
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}

func unavailable(reason string) PrometheusCorrelationResult {
	return PrometheusCorrelationResult{Available: false, Reason: reason}
}
