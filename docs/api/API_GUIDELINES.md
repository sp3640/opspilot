# API Guidelines

## Design Principles
- Stable path prefix: `/api/v1`.
- Consistent envelope: `{ success, message, data }`.
- Authenticated endpoints use Bearer JWT.
- Pagination defaults: `page=1`, `limit=20`, `limit<=100`.
- Sort values are allowlisted per endpoint.

## Authentication
- Public endpoints: `/auth/register`, `/auth/login`, health/metrics endpoints.
- Protected endpoints: all project/incident/comment/dashboard/user routes.

## Validation Rules
- Request bodies validated by Gin binding tags and service-layer checks.
- Domain constraints (severity/status/project ownership) validated in services.

## Error Contract
- 400: invalid request payload/query.
- 401: missing/invalid bearer token.
- 403: ownership/authorization violation.
- 404: resource not found.
- 409: conflict (duplicate user email/project slug).
- 429: rate limit exceeded (`Retry-After` header present).
- 500: internal server error (generic message returned).

## Versioning
### Current Implementation
- Single `v1` route namespace.

### Future Architecture
- Add additive evolution policy and deprecation headers.
- Publish OpenAPI spec and changelog.

## Idempotency
### Current
- Not implemented.

### Future Architecture
- Add idempotency keys for create/update operations exposed to automation clients.

## Observability
- All responses include `X-Request-ID` header.
- Request/latency/status captured by structured logs and metrics.

## Related
- `docs/api/PROJECT_API.md`
- `docs/api/INCIDENT_API.md`
- `docs/architecture/SECURITY_ARCHITECTURE.md`
