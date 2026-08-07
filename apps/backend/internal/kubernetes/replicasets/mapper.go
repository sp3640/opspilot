package replicasets

import (
	"time"

	"github.com/sp3640/opspilot/backend/internal/dto"
	appsv1 "k8s.io/api/apps/v1"
)

type ReplicaSetMapper struct {
	statusHelper *ReplicaSetStatusHelper
}

func NewReplicaSetMapper(
	statusHelper *ReplicaSetStatusHelper,
) *ReplicaSetMapper {
	return &ReplicaSetMapper{
		statusHelper: statusHelper,
	}
}
func (m *ReplicaSetMapper) MapReplicaSet(
	rs appsv1.ReplicaSet,
) dto.ReplicaSet {

	return dto.ReplicaSet{
		Name:              rs.Name,
		Namespace:         rs.Namespace,
		DesiredReplicas:   *rs.Spec.Replicas,
		CurrentReplicas:   rs.Status.Replicas,
		ReadyReplicas:     rs.Status.ReadyReplicas,
		AvailableReplicas: rs.Status.AvailableReplicas,
		Status: m.statusHelper.DetermineStatus(
			*rs.Spec.Replicas,
			rs.Status.ReadyReplicas,
			rs.Status.AvailableReplicas,
		),
		Age:       CalculateReplicaSetAge(rs.CreationTimestamp.Time),
		CreatedAt: rs.CreationTimestamp.Time,
	}
}
func (m *ReplicaSetMapper) MapReplicaSetList(
	items []appsv1.ReplicaSet,
) []dto.ReplicaSet {

	result := make([]dto.ReplicaSet, 0, len(items))

	for _, item := range items {
		result = append(result, m.MapReplicaSet(item))
	}

	return result
}
func CalculateReplicaSetAge(created time.Time) string {
	return time.Since(created).Round(time.Minute).String()
}