# Incident API

## Endpoint: POST /api/v1/incidents
- Purpose: create an incident within owned project.
- Authentication: Bearer JWT required.
- Request: `{ title, description, severity, status, project_id }`.
- Response: 201 with `IncidentResponse`.
- Validation: severity/status enum checks; project ownership required.
- Errors: 400, 401, 403, 500.
- Future extensions: assignment, labels, runbook link.

## Endpoint: GET /api/v1/incidents
- Purpose: list owner incidents.
- Authentication: required.
- Request: query `page`, `limit`, `search`, `sort`, `order`, `projectID`, `status`, `severity`.
- Response: paginated `IncidentListResponse`.
- Validation: sort allowlist and pagination bounds.
- Errors: 400, 401, 500.

## Endpoint: GET /api/v1/incidents/:id
- Purpose: fetch owner incident by numeric id.
- Authentication: required.
- Response: 200 with `IncidentResponse`.
- Errors: 400, 401, 403, 404, 500.

## Endpoint: PUT /api/v1/incidents/:id
- Purpose: update incident fields and project association.
- Authentication: required.
- Request: same shape as create.
- Response: 200 with updated `IncidentResponse`.
- Validation: severity/status + project ownership.
- Errors: 400, 401, 403, 404, 500.

## Endpoint: DELETE /api/v1/incidents/:id
- Purpose: delete incident and dependent comments.
- Authentication: required.
- Response: 200 success envelope.
- Errors: 400, 401, 403, 404, 500.
- Future extensions: soft delete with restore.

## Endpoint: GET /api/v1/incidents/:id/audit-logs
- Purpose: list incident audit history.
- Authentication: required.
- Request: pagination + optional `action`, `entityType`, `search`.
- Response: paginated `AuditLog` items.
- Errors: 400, 401, 403, 404, 500.

## Endpoint: POST /api/v1/incidents/:id/comments
- Purpose: add comment to incident.
- Authentication: required.
- Request: `{ content }`.
- Response: 201 with `Comment` model payload.
- Validation: non-empty, <=2000 chars.
- Errors: 400, 401, 403, 404, 500.

## Endpoint: GET /api/v1/incidents/:id/comments
- Purpose: list incident comments.
- Authentication: required.
- Request: pagination + optional search.
- Response: paginated comment list.
- Errors: 400, 401, 403, 404, 500.

## Endpoint: PUT /api/v1/comments/:id
- Purpose: update comment content by comment owner.
- Authentication: required.
- Request: `{ content }`.
- Response: 200 with updated `Comment`.
- Errors: 400, 401, 403, 404, 500.

## Endpoint: DELETE /api/v1/comments/:id
- Purpose: delete comment by owner.
- Authentication: required.
- Response: 200 success envelope.
- Errors: 400, 401, 403, 404, 500.

## Related
- `docs/architecture/INCIDENT_ENGINE.md`
- `docs/database/ER_DIAGRAM.md`
