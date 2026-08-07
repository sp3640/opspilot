package replicasets

type ReplicaSetStatusHelper struct {
}

func NewReplicaSetStatusHelper() *ReplicaSetStatusHelper {
	return &ReplicaSetStatusHelper{}
}

func (h *ReplicaSetStatusHelper) DetermineStatus(
	desired int32,
	ready int32,
	available int32,
) string {

	if desired == 0 {
		return "ScaledDown"
	}

	if ready == desired && available == desired {
		return "Ready"
	}

	if ready > 0 {
		return "Progressing"
	}

	return "Unavailable"
}
