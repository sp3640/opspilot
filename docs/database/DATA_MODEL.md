# Data Model

## Problem Statement
The model must support secure per-user operations now while evolving to organization-level operations later.

## Current Data Model (Implemented)
### User
- Identity: `id`, `name`, `email`.
- Credentials: `password_hash`.

### Project
- Scope root for most operational entities.
- Ownership enforced via `owner_id`.

### Incident
- Core ops event with `severity`, `status`, `project_id`, `user_id`.

### Comment
- Investigation collaboration records on incidents.

### AuditLog
- Operational trace for all mutable core entities.

## Future Data Model (Future Architecture)
- Organization and team hierarchy.
- Resource graph entities.
- Alert and monitoring signal entities.
- AI context memory and retrieval metadata.

## Constraints
### Current
- Input validation in handlers/services.
- Enum-like constraints for severity/status are service-level, not DB-level.

### Why
Service-level validation kept schema flexible during MVP and reduced migration burden.

### Future
- Introduce DB-level check constraints for critical enums.
- Add stricter referential actions and cascade policies where safe.

## Indexing Strategy
### Current
- Resource-specific indexes for list filters and ownership lookups.

### Future
- Composite indexes for common dashboard queries.
- Time-series indexes for monitoring tables.

## Migration Strategy
- Current: startup AutoMigrate.
- Future: explicit migration files, semantic versions, canary migrations.

## Audit Model
- Current: `audit_logs` captures user + action + field deltas.
- Future: include correlation IDs and source event IDs for cross-system traceability.

## Related
- `docs/database/DATABASE.md`
- `docs/api/API_GUIDELINES.md`
