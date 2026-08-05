package ingresses

import (
	"fmt"
	"time"

	networkingv1 "k8s.io/api/networking/v1"
)

type IngressStatusHelper struct {
	now func() time.Time
}

func NewIngressStatusHelper() *IngressStatusHelper {
	return &IngressStatusHelper{now: time.Now}
}

func (h *IngressStatusHelper) CalculateIngressAge(createdAt time.Time) string {
	if createdAt.IsZero() {
		return ""
	}

	delta := h.now().Sub(createdAt)
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

func (h *IngressStatusHelper) ResolveIngressAddress(ingress networkingv1.Ingress) (string, string) {
	for _, address := range ingress.Status.LoadBalancer.Ingress {
		if address.IP != "" {
			return address.IP, address.Hostname
		}
		if address.Hostname != "" {
			return "", address.Hostname
		}
	}

	return "", ""
}
