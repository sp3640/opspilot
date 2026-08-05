package services

import (
	"strings"

	"github.com/sp3640/opspilot/backend/internal/dto"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceMapper struct {
	helper *ServiceStatusHelper
}

func NewServiceMapper(helper *ServiceStatusHelper) *ServiceMapper {
	if helper == nil {
		helper = NewServiceStatusHelper()
	}

	return &ServiceMapper{helper: helper}
}

func (m *ServiceMapper) MapService(service corev1.Service, includeDetails bool) dto.ServiceResponse {
	externalIPs, loadBalancerIP := m.helper.ResolveExternalAddress(service)

	ports := make([]dto.ServicePortResponse, 0, len(service.Spec.Ports))
	targetPorts := make([]string, 0, len(service.Spec.Ports))
	protocols := make([]string, 0, len(service.Spec.Ports))
	protocolSet := map[string]struct{}{}
	for _, port := range service.Spec.Ports {
		targetPort := port.TargetPort.String()
		if targetPort == "" {
			targetPort = "0"
		}
		protocol := string(port.Protocol)
		if protocol == "" {
			protocol = string(corev1.ProtocolTCP)
		}
		ports = append(ports, dto.ServicePortResponse{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: targetPort,
			Protocol:   protocol,
			NodePort:   port.NodePort,
		})
		targetPorts = append(targetPorts, targetPort)
		if _, exists := protocolSet[protocol]; !exists {
			protocolSet[protocol] = struct{}{}
			protocols = append(protocols, protocol)
		}
	}

	response := dto.ServiceResponse{
		Name:              service.Name,
		Namespace:         service.Namespace,
		Type:              string(service.Spec.Type),
		ClusterIP:         service.Spec.ClusterIP,
		ExternalIPs:       externalIPs,
		LoadBalancerIP:    loadBalancerIP,
		Ports:             ports,
		TargetPorts:       targetPorts,
		Protocol:          strings.Join(protocols, ","),
		Selector:          service.Spec.Selector,
		Labels:            service.Labels,
		Annotations:       service.Annotations,
		Age:               m.helper.CalculateServiceAge(service.CreationTimestamp.Time),
		CreationTimestamp: service.CreationTimestamp.Time,
	}

	if !includeDetails {
		return response
	}

	if service.Spec.InternalTrafficPolicy != nil {
		response.InternalTrafficPolicy = string(*service.Spec.InternalTrafficPolicy)
	}
	if service.Spec.IPFamilyPolicy != nil {
		response.IPFamilyPolicy = string(*service.Spec.IPFamilyPolicy)
	}

	response.SessionAffinity = string(service.Spec.SessionAffinity)
	response.ExternalTrafficPolicy = string(service.Spec.ExternalTrafficPolicy)
	response.HealthCheckNodePort = service.Spec.HealthCheckNodePort
	response.PublishNotReadyAddresses = service.Spec.PublishNotReadyAddresses
	response.ResourceVersion = service.ResourceVersion
	response.UID = string(service.UID)
	response.OwnerReferences = mapOwnerReferences(service.OwnerReferences)
	response.IPFamilies = mapIPFamilies(service.Spec.IPFamilies)

	return response
}

func mapOwnerReferences(owners []metav1.OwnerReference) []dto.OwnerReferenceResponse {
	responses := make([]dto.OwnerReferenceResponse, 0, len(owners))
	for _, owner := range owners {
		responses = append(responses, dto.OwnerReferenceResponse{
			APIVersion: owner.APIVersion,
			Kind:       owner.Kind,
			Name:       owner.Name,
			UID:        string(owner.UID),
		})
	}

	return responses
}

func mapIPFamilies(families []corev1.IPFamily) []string {
	mapped := make([]string, 0, len(families))
	for _, family := range families {
		mapped = append(mapped, string(family))
	}

	return mapped
}
