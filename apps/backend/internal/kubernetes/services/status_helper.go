package services

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
)

type ServiceStatusHelper struct {
	now func() time.Time
}

func NewServiceStatusHelper() *ServiceStatusHelper {
	return &ServiceStatusHelper{now: time.Now}
}

func (h *ServiceStatusHelper) CalculateServiceAge(createdAt time.Time) string {
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

func (h *ServiceStatusHelper) ResolveExternalAddress(service corev1.Service) (externalIPs []string, loadBalancerIP string) {
	externalIPs = make([]string, 0, len(service.Spec.ExternalIPs))
	externalIPs = append(externalIPs, service.Spec.ExternalIPs...)

	loadBalancerIP = service.Spec.LoadBalancerIP
	for _, ingress := range service.Status.LoadBalancer.Ingress {
		if ingress.IP != "" {
			externalIPs = append(externalIPs, ingress.IP)
			if loadBalancerIP == "" {
				loadBalancerIP = ingress.IP
			}
			continue
		}
		if ingress.Hostname != "" {
			externalIPs = append(externalIPs, ingress.Hostname)
			if loadBalancerIP == "" {
				loadBalancerIP = ingress.Hostname
			}
		}
	}

	return externalIPs, loadBalancerIP
}
