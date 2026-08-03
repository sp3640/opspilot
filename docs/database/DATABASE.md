# Database Architecture

## Problem Statement
OpsPilot requires a durable, auditable operational data model that supports incident workflows today and telemetry/resource expansion later.

## Current Schema
Database is managed by GORM `AutoMigrate` in backend startup and currently includes:
- `users`
- `projects`
- `incidents`
- `comments`
- `audit_logs`

## Relationships
- users 1:N projects via `projects.owner_id`.
- users 1:N incidents via `incidents.user_id`.
- incidents 1:N comments via `comments.incident_id`.
- projects/incidents reference audit logs via nullable FKs + entity metadata.

## Indexes and Constraints
### users
- unique index on `email`.

### projects
- primary key: UUID.
- unique index `idx_projects_slug` on `slug`.
- indexes: `owner_id`, `environment`, `health`.
- checks: `members >= 0`, `services >= 0`.
- soft delete index on `deleted_at`.

### incidents
- indexes on `project_id`, `user_id`.

### comments
- indexes on `incident_id`, `user_id`.

### audit_logs
- indexes on `user_id`, `project_id`, `incident_id`, `entity_type`, `entity_id`, `action`.

## Scaling Strategy
### Current Implementation
- Single PostgreSQL instance.
- Query pagination and bounded limits (`<=100`).
- Sort field allowlists per endpoint to prevent unsafe dynamic sort.

### Future Architecture
- Read replicas for analytic and dashboard-heavy read paths.
- Project or tenant partition strategy for incident/audit datasets.
- Materialized views for operational dashboards.

## Partitioning Strategy
### Current Implementation
- No partitioning configured.

### Future Architecture
- Partition `audit_logs` by time.
- Partition `incidents` by organization/project once tenancy exists.

## Migration Strategy
### Current Implementation
- Automatic schema migration via `db.AutoMigrate(...)` at process startup.

### Why
Simple for MVP velocity and low operational complexity.

### Future Architecture
- Versioned migration tooling with explicit up/down plans.
- Pre-deploy migration checks and rollback procedures.

## Audit Strategy
### Current Implementation
- Audit service writes create/update/delete events with field-level details.
- Incident deletion clears incident FK references in audit logs before hard delete.

### Future Architecture
- Immutable append-only audit stream with retention policy tiers.
- Tamper-evidence and export pipelines for compliance audits.

## Related
- `docs/database/ER_DIAGRAM.md`
- `docs/database/DATA_MODEL.md`
- `docs/architecture/SECURITY_ARCHITECTURE.md`
