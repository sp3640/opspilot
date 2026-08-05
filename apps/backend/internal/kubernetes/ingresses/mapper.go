package ingresses

import (
	"sort"
	"strconv"
	"strings"

	"github.com/sp3640/opspilot/backend/internal/dto"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IngressMapper struct {
	helper *IngressStatusHelper
}

func NewIngressMapper(helper *IngressStatusHelper) *IngressMapper {
	if helper == nil {
		helper = NewIngressStatusHelper()
	}

	return &IngressMapper{helper: helper}
}

func (m *IngressMapper) MapIngress(ingress networkingv1.Ingress, includeDetails bool) dto.IngressResponse {
	hostSet := map[string]struct{}{}
	pathSet := map[string]struct{}{}
	backendServiceSet := map[string]struct{}{}
	rules := make([]dto.IngressRuleResponse, 0, len(ingress.Spec.Rules))
	httpPaths := make([]dto.IngressPathResponse, 0)

	for _, rule := range ingress.Spec.Rules {
		if strings.TrimSpace(rule.Host) != "" {
			hostSet[rule.Host] = struct{}{}
		}

		mappedPaths := make([]dto.IngressPathResponse, 0)
		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				mapped := mapIngressPath(path)
				mappedPaths = append(mappedPaths, mapped)
				httpPaths = append(httpPaths, mapped)

				if strings.TrimSpace(mapped.Path) != "" {
					pathSet[mapped.Path] = struct{}{}
				}
				if strings.TrimSpace(mapped.BackendService) != "" {
					backendServiceSet[mapped.BackendService] = struct{}{}
				}
			}
		}

		rules = append(rules, dto.IngressRuleResponse{
			Host:  rule.Host,
			Paths: mappedPaths,
		})
	}

	tlsConfig := make([]dto.IngressTLSResponse, 0, len(ingress.Spec.TLS))
	tlsEnabled := len(ingress.Spec.TLS) > 0
	tlsSecret := ""
	for _, tls := range ingress.Spec.TLS {
		tlsConfig = append(tlsConfig, dto.IngressTLSResponse{
			Hosts:      append([]string{}, tls.Hosts...),
			SecretName: tls.SecretName,
		})
		if tlsSecret == "" && strings.TrimSpace(tls.SecretName) != "" {
			tlsSecret = tls.SecretName
		}
	}

	hosts := mapSetToSortedSlice(hostSet)
	paths := mapSetToSortedSlice(pathSet)
	backendServices := mapSetToSortedSlice(backendServiceSet)
	loadBalancerIP, loadBalancerHostname := m.helper.ResolveIngressAddress(ingress)
	status := "pending"
	if loadBalancerIP != "" || loadBalancerHostname != "" {
		status = "active"
	}

	response := dto.IngressResponse{
		Name:                 ingress.Name,
		Namespace:            ingress.Namespace,
		IngressClass:         resolveIngressClass(ingress),
		Hosts:                hosts,
		Paths:                paths,
		BackendServices:      backendServices,
		TLSEnabled:           tlsEnabled,
		TLSSecret:            tlsSecret,
		LoadBalancerIP:       loadBalancerIP,
		LoadBalancerHostname: loadBalancerHostname,
		Labels:               ingress.Labels,
		Annotations:          ingress.Annotations,
		Age:                  m.helper.CalculateIngressAge(ingress.CreationTimestamp.Time),
		CreationTimestamp:    ingress.CreationTimestamp.Time,
	}

	if !includeDetails {
		return response
	}

	defaultBackendService, defaultBackendPort := resolveDefaultBackend(ingress)
	response.Rules = rules
	response.HTTPPaths = httpPaths
	response.DefaultBackend = defaultBackendService
	response.DefaultBackendPort = defaultBackendPort
	response.TLSConfig = tlsConfig
	response.Status = status
	response.OwnerReferences = mapIngressOwnerReferences(ingress.OwnerReferences)
	response.UID = string(ingress.UID)
	response.ResourceVersion = ingress.ResourceVersion

	return response
}

func resolveIngressClass(ingress networkingv1.Ingress) string {
	if ingress.Spec.IngressClassName != nil {
		return strings.TrimSpace(*ingress.Spec.IngressClassName)
	}

	if ingress.Annotations != nil {
		if ingressClass, ok := ingress.Annotations["kubernetes.io/ingress.class"]; ok {
			return strings.TrimSpace(ingressClass)
		}
	}

	return ""
}

func mapIngressPath(path networkingv1.HTTPIngressPath) dto.IngressPathResponse {
	response := dto.IngressPathResponse{
		Path: path.Path,
	}
	if path.PathType != nil {
		response.PathType = string(*path.PathType)
	}
	if path.Backend.Service != nil {
		response.BackendService = path.Backend.Service.Name
		response.ServicePort = path.Backend.Service.Port.Name
		if strings.TrimSpace(response.ServicePort) == "" && path.Backend.Service.Port.Number != 0 {
			response.ServicePort = strconv.Itoa(int(path.Backend.Service.Port.Number))
		}
	}

	return response
}

func resolveDefaultBackend(ingress networkingv1.Ingress) (string, string) {
	if ingress.Spec.DefaultBackend == nil || ingress.Spec.DefaultBackend.Service == nil {
		return "", ""
	}

	service := ingress.Spec.DefaultBackend.Service
	port := service.Port.Name
	if strings.TrimSpace(port) == "" && service.Port.Number != 0 {
		port = strconv.Itoa(int(service.Port.Number))
	}

	return service.Name, port
}

func mapIngressOwnerReferences(owners []metav1.OwnerReference) []dto.OwnerReferenceResponse {
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

func mapSetToSortedSlice(set map[string]struct{}) []string {
	items := make([]string, 0, len(set))
	for value := range set {
		if strings.TrimSpace(value) == "" {
			continue
		}
		items = append(items, value)
	}
	sort.Strings(items)

	return items
}
