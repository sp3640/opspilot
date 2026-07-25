# OpsPilot API

OpsPilot API is a Go service for authentication, projects, incidents, comments,
audit history, and dashboard summaries. It uses Gin for HTTP routing, GORM for
PostgreSQL persistence, and Prometheus-compatible metrics for observability.

## Architecture

The application keeps HTTP, business, and persistence concerns separate:

~~~text
cmd/server                 application wiring, lifecycle, graceful shutdown
internal/router            public and protected route registration
internal/middleware        request ID, logging, recovery, security, rate limits, auth
internal/handlers          HTTP parsing and consistent API responses
internal/services          business rules and audit orchestration
internal/repository        PostgreSQL data access through GORM
internal/database          connection and schema migration setup
internal/response          shared success/error response format
~~~

Every response carries X-Request-ID. Request logs include that ID, request
method and path, response status, latency, and client IP. Panics are recovered
without exposing internal details; their stack traces are logged server-side.

## Local setup

Prerequisites: Go 1.25+, PostgreSQL 16+ (or Docker), and a PostgreSQL database.

~~~bash
cd apps/backend
cp .env.example .env
# Edit .env and replace both placeholder secrets.
go mod download
go run ./cmd/server
~~~

The service listens on http://localhost:8080 by default. Configuration is
validated during startup, so an invalid environment, database configuration,
port, or JWT secret stops the process before it starts serving traffic.

### Environment variables

Never commit .env files or production credentials. Use a secret manager or
your platform's secret mechanism in deployed environments.

| Variable | Default | Notes |
| --- | --- | --- |
| APP_NAME | OpsPilot Backend | Log and health endpoint service name. |
| APP_ENV | development | One of development, test, staging, or production. |
| APP_VERSION | dev | Version reported by /health. |
| PORT | 8080 | HTTP listen port; must be a valid TCP port. |
| JWT_SECRET | none | Required. Use a high-entropy secret in every environment. |
| JWT_EXPIRY | 24h | JWT lifetime accepted by Go duration parsing. |
| DATABASE_URL | none | Preferred PostgreSQL connection URL. Takes precedence when set. |
| DB_HOST | none | Required legacy PostgreSQL host when DATABASE_URL is not set. |
	| DB_PORT | 5432 | Legacy PostgreSQL port. |
| DB_USER | none | Required legacy PostgreSQL user when DATABASE_URL is not set. |
	| DB_PASSWORD | none | Legacy PostgreSQL password. |
| DB_NAME | none | Required legacy PostgreSQL database name when DATABASE_URL is not set. |
| DB_SSLMODE | disable | PostgreSQL SSL mode; use require or stronger in production. |
| RATE_LIMIT_REQUESTS | 100 | Maximum requests per IP per rate-limit window. |
| RATE_LIMIT_WINDOW | 1m | Go duration for the rate-limit window. |

Set either DATABASE_URL or the legacy DB_* values. A production URL should use
TLS, for example postgres://user:password@db.example:5432/opspilot?sslmode=require.
URL-encode special characters in the username or password.

## Endpoints

Operational endpoints are unauthenticated so container and orchestration health
checks can reach them:

| Method | Path | Purpose |
| --- | --- | --- |
| GET | /health | Database connectivity, uptime, and application version. |
| GET | /ready | Readiness: application initialized and database reachable. |
| GET | /live | Liveness: process is alive. |
| GET | /metrics | Prometheus metrics: requests, durations, statuses, panics, and active requests. |

API routes are rooted at /api/v1. Protected endpoints require a bearer token.

| Area | Routes |
| --- | --- |
| Authentication | POST /auth/register, POST /auth/login |
| User | GET /users/me |
| Projects | POST, GET /projects; GET, PUT, DELETE /projects/:id; GET /projects/:id/audit-logs |
| Incidents | POST, GET /incidents; GET, PUT, DELETE /incidents/:id; POST, GET /incidents/:id/comments; GET /incidents/:id/audit-logs |
| Comments | PUT, DELETE /comments/:id |
| Dashboard | GET /dashboard/summary, /dashboard/recent-incidents, /dashboard/activity, /dashboard/stats |

All paginated list endpoints enforce page >= 1, default limit=20, and limit <=
100; sorting accepts only endpoint-approved fields and asc or desc order.

## Docker

The included Dockerfile uses a Go build stage and a small Alpine runtime image.
The runtime image contains only the compiled binary, CA certificates, timezone
data, and a non-root opspilot user.

~~~bash
cd apps/backend
docker build -t opspilot-api:local .
docker run --rm -p 8080:8080 -e JWT_SECRET='replace-with-a-long-random-secret' -e DATABASE_URL='postgres://user:password@host:5432/opspilot?sslmode=require' opspilot-api:local
~~~

For a complete local stack, copy .env.example first and replace the placeholder
values, then run:

~~~bash
docker compose up --build
~~~

Compose starts PostgreSQL, waits until it is healthy, then starts the API. The
database port binds to loopback only; use POSTGRES_PORT to choose its host port.
Persistent development data is stored in the postgres-data volume.

## Kubernetes

The k8s/ directory contains a reusable Deployment, Service, ConfigMap, and
Kustomize entrypoint. The Deployment runs two non-root replicas with readiness
(/ready), liveness (/live), and startup probes, resource requests and limits,
graceful termination time, and Prometheus scrape annotations.

Create the required secret through your secret manager or directly for a test
environment. Do not apply secret.example.yaml unchanged.

~~~bash
kubectl -n opspilot create secret generic opspilot-api-secrets --from-literal=DATABASE_URL='postgres://user:password@db:5432/opspilot?sslmode=require' --from-literal=JWT_SECRET='replace-with-a-long-random-secret'
# Replace the image in k8s/deployment.yaml with an immutable image tag first.
kubectl -n opspilot apply -k k8s
kubectl -n opspilot rollout status deployment/opspilot-api
~~~

Create the opspilot namespace before these commands if it does not already
exist. For GitOps environments, keep the ConfigMap in source control and source
the secret from the platform's external-secret integration instead of committing
credentials.

## Production deployment checklist

- Set APP_ENV=production, a release APP_VERSION, and a strong JWT_SECRET.
- Use a managed PostgreSQL database with TLS and provide DATABASE_URL via a
  secret manager.
- Publish and deploy an immutable, scanned container image rather than latest.
- Route public traffic through an ingress or load balancer that terminates TLS.
- Scrape /metrics, retain structured logs, and alert on elevated 5xx, panic,
  latency, and readiness failures.
- Tune replica counts and resource values from observed production load; retain
  the readiness probe so unavailable pods receive no traffic during rollouts.
