package configmaps

import (
	"fmt"
	"sort"
	"time"

	"github.com/sp3640/opspilot/backend/internal/dto"
	corev1 "k8s.io/api/core/v1"
)

type ConfigMapMapper struct {
	now func() time.Time
}

func NewConfigMapMapper() *ConfigMapMapper {
	return &ConfigMapMapper{now: time.Now}
}

func (m *ConfigMapMapper) CalculateConfigMapAge(createdAt time.Time) string {
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

func (m *ConfigMapMapper) MapConfigMap(configMap corev1.ConfigMap, includeDetails bool) dto.ConfigMapResponse {
	response := dto.ConfigMapResponse{
		Name:              configMap.Name,
		Namespace:         configMap.Namespace,
		DataKeyCount:      len(configMap.Data),
		Labels:            configMap.Labels,
		Annotations:       configMap.Annotations,
		Age:               m.CalculateConfigMapAge(configMap.CreationTimestamp.Time),
		CreationTimestamp: configMap.CreationTimestamp.Time,
	}

	if !includeDetails {
		return response
	}

	response.Data = configMap.Data
	response.Immutable = configMap.Immutable != nil && *configMap.Immutable
	response.UID = string(configMap.UID)
	response.ResourceVersion = configMap.ResourceVersion
	response.BinaryData = makeBinaryDataMetadata(configMap.BinaryData)

	return response
}

func makeBinaryDataMetadata(binaryData map[string][]byte) []dto.ConfigMapBinaryDataMetadataResponse {
	items := make([]dto.ConfigMapBinaryDataMetadataResponse, 0, len(binaryData))
	for key, value := range binaryData {
		items = append(items, dto.ConfigMapBinaryDataMetadataResponse{Key: key, SizeBytes: len(value)})
	}

	sort.Slice(items, func(i int, j int) bool {
		return items[i].Key < items[j].Key
	})

	return items
}
