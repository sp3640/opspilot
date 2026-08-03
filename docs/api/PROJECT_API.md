# Project API

## Endpoint: POST /api/v1/projects
- Purpose: create a project for authenticated owner.
- Authentication: Bearer JWT required.
- Request: `{ name, description }`.
- Response: 201 with `ProjectResponse`.
- Validation: `name` length 3-100, `description` <=300.
- Errors: 400, 401, 409, 500.
- Future extensions: environment, team assignment, tags.

## Endpoint: GET /api/v1/projects
- Purpose: list owner projects.
- Authentication: required.
- Request: query `page`, `limit`, `search`, `sort(name|created_at|updated_at)`, `order(asc|desc)`.
- Response: paginated `ProjectListResponse`.
- Validation: pagination bounds and sort allowlist.
- Errors: 400, 401, 500.
- Future extensions: filter by health/environment.

## Endpoint: GET /api/v1/projects/:id
- Purpose: fetch one owner-scoped project.
- Authentication: required.
- Response: 200 with `ProjectResponse`.
- Errors: 400, 401, 403, 404, 500.

## Endpoint: PUT /api/v1/projects/:id
- Purpose: update owner project name/description.
- Authentication: required.
- Request: `{ name, description }`.
- Response: 200 with updated `ProjectResponse`.
- Validation: same as create.
- Errors: 400, 401, 403, 404, 500.
- Future extensions: patch semantics.

## Endpoint: DELETE /api/v1/projects/:id
- Purpose: soft-delete owner project.
- Authentication: required.
- Response: 200 success envelope.
- Errors: 400, 401, 403, 404, 500.
- Future extensions: restore endpoint.

## Endpoint: GET /api/v1/projects/:id/audit-logs
- Purpose: list project-scoped audit logs.
- Authentication: required.
- Request: pagination + optional `action`, `entityType`, `search`.
- Response: paginated `AuditLog` items.
- Validation: sort allowlist `created_at|entity_type|action`.
- Errors: 400, 401, 403, 404, 500.

## Related
- `docs/database/DATA_MODEL.md`
- `docs/api/API_GUIDELINES.md`
