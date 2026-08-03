package resourcesync

import (
	"bytes"
	"sort"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
)

type SyncPlan struct {
	ResourcesToCreate  []models.Resource
	ResourcesToUpdate  []models.Resource
	ResourcesToDelete  []models.Resource
	ResourcesToRestore []models.Resource
}

func BuildSyncPlan(persistedResources, discoveredResources []models.Resource) SyncPlan {
	persistedByKey := make(map[string]models.Resource, len(persistedResources))
	discoveredByKey := make(map[string]models.Resource, len(discoveredResources))

	for _, resource := range persistedResources {
		persistedByKey[resourceIdentity(resource)] = resource
	}

	for _, resource := range discoveredResources {
		discoveredByKey[resourceIdentity(resource)] = resource
	}

	plan := SyncPlan{
		ResourcesToCreate:  make([]models.Resource, 0),
		ResourcesToUpdate:  make([]models.Resource, 0),
		ResourcesToDelete:  make([]models.Resource, 0),
		ResourcesToRestore: make([]models.Resource, 0),
	}

	for key, discovered := range discoveredByKey {
		persisted, exists := persistedByKey[key]
		if !exists {
			plan.ResourcesToCreate = append(plan.ResourcesToCreate, discovered)
			continue
		}

		merged := mergeResourceForSync(persisted, discovered)

		if persisted.DeletedAt.Valid {
			plan.ResourcesToRestore = append(plan.ResourcesToRestore, merged)
			continue
		}

		if resourcesDiffer(persisted, merged) {
			plan.ResourcesToUpdate = append(plan.ResourcesToUpdate, merged)
		}
	}

	for key, persisted := range persistedByKey {
		if _, exists := discoveredByKey[key]; exists {
			continue
		}

		if persisted.DeletedAt.Valid {
			continue
		}

		plan.ResourcesToDelete = append(plan.ResourcesToDelete, persisted)
	}

	sortResources(plan.ResourcesToCreate)
	sortResources(plan.ResourcesToUpdate)
	sortResources(plan.ResourcesToDelete)
	sortResources(plan.ResourcesToRestore)

	return plan
}

func resourceIdentity(resource models.Resource) string {
	return resource.ProjectID.String() + ":" + resource.Kind + ":" + resource.ExternalID
}

func mergeResourceForSync(persisted, discovered models.Resource) models.Resource {
	merged := discovered
	merged.ID = persisted.ID
	merged.CreatedAt = persisted.CreatedAt
	merged.UpdatedAt = persisted.UpdatedAt
	merged.DeletedAt = persisted.DeletedAt

	if merged.ParentResourceID == nil {
		merged.ParentResourceID = persisted.ParentResourceID
	}

	if merged.CreatedBy == 0 {
		merged.CreatedBy = persisted.CreatedBy
	}

	return merged
}

func resourcesDiffer(left, right models.Resource) bool {
	if left.ProjectID != right.ProjectID ||
		left.Kind != right.Kind ||
		left.Name != right.Name ||
		left.DisplayName != right.DisplayName ||
		left.ExternalID != right.ExternalID ||
		left.Provider != right.Provider ||
		left.Region != right.Region ||
		left.Namespace != right.Namespace ||
		left.Cluster != right.Cluster ||
		left.Status != right.Status ||
		left.Health != right.Health ||
		left.CreatedBy != right.CreatedBy {
		return true
	}

	if !uuidPointersEqual(left.ParentResourceID, right.ParentResourceID) {
		return true
	}

	if !bytes.Equal(normalizeJSON(left.Labels), normalizeJSON(right.Labels)) {
		return true
	}
	if !bytes.Equal(normalizeJSON(left.Annotations), normalizeJSON(right.Annotations)) {
		return true
	}
	if !bytes.Equal(normalizeJSON(left.Metadata), normalizeJSON(right.Metadata)) {
		return true
	}

	return false
}

func normalizeJSON(value []byte) []byte {
	if len(value) == 0 {
		return []byte("{}")
	}

	return value
}

func uuidPointersEqual(left, right *uuid.UUID) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}

	return *left == *right
}

func sortResources(resources []models.Resource) {
	sort.Slice(resources, func(i, j int) bool {
		left := resourceIdentity(resources[i])
		right := resourceIdentity(resources[j])
		return left < right
	})
}
