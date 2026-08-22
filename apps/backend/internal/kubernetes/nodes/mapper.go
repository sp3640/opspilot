package nodes

import (
	"fmt"
	"time"

	"github.com/sp3640/opspilot/backend/internal/dto"
	corev1 "k8s.io/api/core/v1"
)

type NodeMapper struct {
	now func() time.Time
}

func NewNodeMapper() *NodeMapper {
	return &NodeMapper{now: time.Now}
}

func (m *NodeMapper) CalculateNodeAge(createdAt time.Time) string {
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

func (m *NodeMapper) MapNode(node corev1.Node) dto.NodeResponse {
	return dto.NodeResponse{
		Name:              node.Name,
		Ready:             nodeReady(node),
		Unschedulable:     node.Spec.Unschedulable,
		KubeletVersion:    node.Status.NodeInfo.KubeletVersion,
		OperatingSystem:   node.Status.NodeInfo.OperatingSystem,
		Architecture:      node.Status.NodeInfo.Architecture,
		InternalIP:        nodeInternalIP(node),
		Labels:            node.Labels,
		Annotations:       node.Annotations,
		Age:               m.CalculateNodeAge(node.CreationTimestamp.Time),
		CreationTimestamp: node.CreationTimestamp.Time,
	}
}

func (m *NodeMapper) MapNodeList(items []corev1.Node) []dto.NodeResponse {
	responses := make([]dto.NodeResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, m.MapNode(item))
	}

	return responses
}

func nodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}

	return false
}

func nodeInternalIP(node corev1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			return address.Address
		}
	}

	return ""
}
