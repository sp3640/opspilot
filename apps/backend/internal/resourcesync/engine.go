package resourcesync

import (
	"context"
	"time"

	"github.com/sp3640/opspilot/backend/internal/models"
)

type SyncEngine struct {
	store ResourceStore
}

func NewSyncEngine(store ResourceStore) *SyncEngine {
	return &SyncEngine{store: store}
}

func (e *SyncEngine) Plan(ctx context.Context, projectID string, discoveredResources []models.Resource) (*SyncPlan, error) {
	if e == nil || e.store == nil {
		plan := BuildSyncPlan(nil, discoveredResources)
		return &plan, nil
	}

	persistedResources, err := e.store.ListResources(ctx, projectID)
	if err != nil {
		return nil, err
	}

	plan := BuildSyncPlan(persistedResources, discoveredResources)
	return &plan, nil
}

func (e *SyncEngine) Execute(ctx context.Context, plan *SyncPlan) (*SyncResult, error) {
	startedAt := time.Now()
	result := &SyncResult{
		Errors: make([]string, 0),
	}

	if plan == nil {
		result.Duration = time.Since(startedAt)
		return result, nil
	}

	if e == nil || e.store == nil {
		result.Duration = time.Since(startedAt)
		return result, nil
	}

	if len(plan.ResourcesToCreate) > 0 {
		if err := e.store.CreateResources(ctx, plan.ResourcesToCreate); err != nil {
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.Created = len(plan.ResourcesToCreate)
		}
	}

	if len(plan.ResourcesToUpdate) > 0 {
		if err := e.store.UpdateResources(ctx, plan.ResourcesToUpdate); err != nil {
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.Updated = len(plan.ResourcesToUpdate)
		}
	}

	if len(plan.ResourcesToDelete) > 0 {
		if err := e.store.SoftDeleteResources(ctx, plan.ResourcesToDelete); err != nil {
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.Deleted = len(plan.ResourcesToDelete)
		}
	}

	if len(plan.ResourcesToRestore) > 0 {
		if err := e.store.RestoreResources(ctx, plan.ResourcesToRestore); err != nil {
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.Restored = len(plan.ResourcesToRestore)
		}
	}

	result.Duration = time.Since(startedAt)
	return result, nil
}
