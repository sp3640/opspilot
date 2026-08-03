# Monitoring API

## Current Implementation
Monitoring ingestion/query APIs are not implemented.

## Operational Endpoints (Implemented)
### Endpoint: GET /health
- Purpose: process health including DB reachability and uptime.
- Authentication: none.
- Response: 200 or 503.
- Errors: not applicable (status models degraded states).

### Endpoint: GET /ready
- Purpose: readiness for orchestrators.
- Authentication: none.
- Response: 200 when initialized and DB reachable, else 503.

### Endpoint: GET /live
- Purpose: liveness probe.
- Authentication: none.
- Response: 200 if process alive.

### Endpoint: GET /metrics
- Purpose: Prometheus scrape endpoint.
- Authentication: none (protect at network layer).
- Response: text exposition format.

## Future Architecture
- `/api/v1/monitoring/metrics`
- `/api/v1/monitoring/logs`
- `/api/v1/monitoring/traces`

For future endpoints:
- Authentication: required.
- Validation: project/resource scope and query bounds.
- Errors: 400/401/403/500.
- Future extensions: query templates and saved views.

## Related
- `docs/monitoring/HEALTH_ENGINE.md`
- `docs/monitoring/METRICS.md`
