package namespaces

import (
	"fmt"
	"time"

	"github.com/sp3640/opspilot/backend/internal/dto"
	corev1 "k8s.io/api/core/v1"
)

type NamespaceMapper struct {
	now func() time.Time
}

func NewNamespaceMapper() *NamespaceMapper {
	return &NamespaceMapper{now: time.Now}
}

func (m *NamespaceMapper) CalculateNamespaceAge(createdAt time.Time) string {
	if createdAt.IsZero() {
		return ""
	}

	delta := m.now().Sub(createdAt)
	if delta < time.Minute {
		return fmt.Sprintf("%ds", int(delta.Seconds()))
	}
	if delta < time.Hour {
		return fmt.Sprintf("%dm", int(delta.Minutes()))
	}
	if delta < 24*time.Hour {
		return fmt.Sprintf("%dh", int(delta.Hours()))
	}

	return fmt.Sprintf("%dd", int(delta.Hours()/24))
}

func (m *NamespaceMapper) MapNamespace(namespace corev1.Namespace) dto.NamespaceResponse {
	return dto.NamespaceResponse{
		Name:              namespace.Name,
		Phase:             string(namespace.Status.Phase),
		Labels:            namespace.Labels,
		Annotations:       namespace.Annotations,
		Age:               m.CalculateNamespaceAge(namespace.CreationTimestamp.Time),
		CreationTimestamp: namespace.CreationTimestamp.Time,
	}
}

func (m *NamespaceMapper) MapNamespaceList(items []corev1.Namespace) []dto.NamespaceResponse {
	responses := make([]dto.NamespaceResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, m.MapNamespace(item))
	}

	return responses
}
