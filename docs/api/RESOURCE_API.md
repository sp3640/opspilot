# Resource API

## Current Implementation
No resource inventory endpoints are implemented.

## Future Architecture
Planned endpoint family under `/api/v1/resources`.

### Endpoint: GET /resources
- Purpose: list resources by project, kind, health.
- Authentication: required.
- Request: query filters (`projectId`, `kind`, `namespace`, `health`, pagination).
- Response: paginated canonical resource objects.
- Validation: allowlisted filter/sort fields.
- Errors: 400/401/403/500.
- Future extensions: graph expansion and dependency traversal.

### Endpoint: GET /resources/:id
- Purpose: fetch resource details and relationships.
- Authentication: required.
- Future extensions: include last telemetry snapshot.

### Endpoint: POST /resources/discovery-runs
- Purpose: trigger manual discovery.
- Authentication: required.
- Future extensions: async job tracking and webhooks.

## Related
- `docs/architecture/RESOURCE_INVENTORY.md`
- `docs/architecture/DISCOVERY_ENGINE.md`
