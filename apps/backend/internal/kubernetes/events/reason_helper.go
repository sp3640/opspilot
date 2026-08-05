package events

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type ReasonHelper struct{}

func NewReasonHelper() *ReasonHelper {
	return &ReasonHelper{}
}

func (h *ReasonHelper) NormalizeType(eventType string) string {
	value := strings.TrimSpace(eventType)
	if strings.EqualFold(value, corev1.EventTypeWarning) {
		return corev1.EventTypeWarning
	}

	return corev1.EventTypeNormal
}

func (h *ReasonHelper) ResolveInvolvedObject(event corev1.Event) string {
	parts := make([]string, 0, 3)
	if strings.TrimSpace(event.InvolvedObject.Kind) != "" {
		parts = append(parts, event.InvolvedObject.Kind)
	}
	if strings.TrimSpace(event.InvolvedObject.Namespace) != "" {
		parts = append(parts, event.InvolvedObject.Namespace)
	}
	if strings.TrimSpace(event.InvolvedObject.Name) != "" {
		parts = append(parts, event.InvolvedObject.Name)
	}

	return strings.Join(parts, "/")
}

func (h *ReasonHelper) ResolveSource(event corev1.Event) (component string, source string) {
	component = strings.TrimSpace(event.Source.Component)
	host := strings.TrimSpace(event.Source.Host)
	if component == "" {
		component = strings.TrimSpace(event.ReportingController)
	}
	if host == "" {
		host = strings.TrimSpace(event.ReportingInstance)
	}

	source = component
	if component != "" && host != "" {
		source = component + "/" + host
	} else if source == "" {
		source = host
	}

	return component, source
}
