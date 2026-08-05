package executor

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/sp3640/opspilot/backend/internal/models"
)

type DeploymentManifestBuilder struct{}

func NewDeploymentManifestBuilder() *DeploymentManifestBuilder {
	return &DeploymentManifestBuilder{}
}

func (b *DeploymentManifestBuilder) Build(deployment *models.Deployment, application *models.Application) *appsv1.Deployment {
	labels := map[string]string{
		"app.kubernetes.io/name":       sanitizeKubernetesName(application.Slug),
		"app.kubernetes.io/instance":   deployment.ID.String(),
		"app.kubernetes.io/managed-by": "opspilot",
		"opspilot/deployment-id":       deployment.ID.String(),
		"opspilot/application-id":      deployment.ApplicationID.String(),
		"opspilot/project-id":          deployment.ProjectID.String(),
		"opspilot/organization-id":     deployment.OrganizationID.String(),
	}

	replicas := int32(deployment.ReplicaCount)
	containerPort := int32(application.Port)
	manifestName := buildManifestName(application, deployment)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      manifestName,
			Namespace: deployment.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{
				"opspilot/deployment-id": deployment.ID.String(),
			}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            manifestName,
							Image:           strings.TrimSpace(deployment.Image + ":" + deployment.ImageTag),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{{
								Name:          "http",
								ContainerPort: containerPort,
								Protocol:      corev1.ProtocolTCP,
							}},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{
									Path: "/",
									Port: intstr.FromInt(application.Port),
								}},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
						},
					},
				},
			},
		},
	}
}

func buildManifestName(application *models.Application, deployment *models.Deployment) string {
	base := application.Slug
	if strings.TrimSpace(base) == "" {
		base = application.Name
	}
	if strings.TrimSpace(base) == "" {
		base = "opspilot-app"
	}

	shortID := deployment.ID.String()
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}

	return sanitizeKubernetesName(fmt.Sprintf("%s-%s", base, shortID))
}

func sanitizeKubernetesName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "opspilot"
	}

	builder := strings.Builder{}
	for _, ch := range value {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			builder.WriteRune(ch)
			continue
		}
		builder.WriteRune('-')
	}

	output := strings.Trim(builder.String(), "-")
	if output == "" {
		output = "opspilot"
	}
	if len(output) > 63 {
		output = strings.Trim(output[:63], "-")
	}
	if output == "" {
		output = "opspilot"
	}

	return output
}
