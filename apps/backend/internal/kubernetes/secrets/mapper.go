package secrets

import (
	"fmt"
	"sort"
	"time"

	"github.com/sp3640/opspilot/backend/internal/dto"
	corev1 "k8s.io/api/core/v1"
)

type SecretMapper struct {
	now func() time.Time
}

func NewSecretMapper() *SecretMapper {
	return &SecretMapper{now: time.Now}
}

func (m *SecretMapper) CalculateSecretAge(createdAt time.Time) string {
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

func (m *SecretMapper) MapSecret(secret corev1.Secret) dto.SecretResponse {
	return dto.SecretResponse{
		Name:              secret.Name,
		Namespace:         secret.Namespace,
		Type:              string(secret.Type),
		DataKeyCount:      len(secret.Data),
		Labels:            secret.Labels,
		Annotations:       secret.Annotations,
		Age:               m.CalculateSecretAge(secret.CreationTimestamp.Time),
		CreationTimestamp: secret.CreationTimestamp.Time,
	}
}

func (m *SecretMapper) MapSecretDetail(secret corev1.Secret) dto.SecretDetailResponse {
	return dto.SecretDetailResponse{
		Name:            secret.Name,
		Namespace:       secret.Namespace,
		Type:            string(secret.Type),
		Keys:            secretKeys(secret),
		Immutable:       secret.Immutable != nil && *secret.Immutable,
		Labels:          secret.Labels,
		Annotations:     secret.Annotations,
		UID:             string(secret.UID),
		ResourceVersion: secret.ResourceVersion,
		Age:             m.CalculateSecretAge(secret.CreationTimestamp.Time),
	}
}

func secretKeys(secret corev1.Secret) []string {
	keySet := make(map[string]struct{}, len(secret.Data)+len(secret.StringData))
	for key := range secret.Data {
		keySet[key] = struct{}{}
	}
	for key := range secret.StringData {
		keySet[key] = struct{}{}
	}

	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}
