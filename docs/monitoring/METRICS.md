# Metrics Architecture

## Problem Statement
Metrics must provide both API health visibility now and service/resource observability as platform scope expands.

## Collection Architecture
### Current Implementation
- Custom Prometheus text collector exposed at `/metrics`.
- Captures request totals, duration histograms, active requests, recovered panics.

### Why
A lightweight internal collector avoids external dependencies while establishing operational baselines.

## Data Flow
```mermaid
flowchart TD
  Request --> Middleware[Collector Middleware]
  Middleware --> InMemory[(Metrics Maps)]
  InMemory --> Endpoint[/metrics]
  Endpoint --> Prometheus[Prometheus Scrape]
```

## Component Diagram
```mermaid
graph LR
  Gin --> Collector --> Prometheus --> Grafana
```

## Prometheus Integration
- Implemented via exposition format endpoint.
- Network-layer protection required in production.

## OpenTelemetry
- Future Architecture: emit metrics through OTel SDK/exporters.

## Grafana
- Future Architecture: dashboard templates for latency, error rate, panic rate, saturation.

## Azure Monitor
- Future Architecture: mirror key metrics through Azure Managed Prometheus or direct exporters.

## Future Kubernetes Integration
- Add ServiceMonitor annotations and scrape configs.

## Tradeoffs
- In-memory collector is simple but per-instance only.

## Open Questions
- Cardinality limits for route labels.
- Retention and downsampling policy.

## Related
- `docs/monitoring/HEALTH_ENGINE.md`
- `docs/monitoring/ALERT_RULES.md`
