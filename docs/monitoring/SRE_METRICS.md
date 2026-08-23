# SRE Metrics (SLO, Availability, Error Budget, MTTR, MTTA)

## Problem Statement
Teams need a real, defensible answer to "how reliable is this application" - derived from actual incident history, never a fabricated or placeholder number. Phase 24 adds an SLO configuration per application and a set of metrics computed strictly from that application's own `Incident` rows.

## Source Data
Every metric below is computed from `Incident` records where `ApplicationID` matches the application in question - a real, non-nullable-by-convention foreign key column (`internal/models/incident.go`), not a best-effort attribution. No other data source (synthetic uptime probes, external APM, a third-party status page) is consulted, because none exists in this system. If a metric cannot be computed honestly from this data, the API returns `"available": false` and `"formatted": "Insufficient data"` with a `reason` - it never invents a number.

Two incident-lifecycle timestamps back these metrics:
- `Incident.ResolvedAt` - set when status transitions to `RESOLVED` (existing, Phase 15+).
- `Incident.AcknowledgedAt` - set once, the first time `PATCH /incidents/:id/acknowledge` is called (added in Phase 24 specifically to give MTTA real backing data, mirroring the equivalent field Alert already had).

## Observation Window
Every calculation is scoped to a window of `WindowDays` (from the application's `ApplicationSLO` config, or 30 days by default if none is configured yet), ending now. The window is then clipped to the application's own age:

```
windowStart = max(now - WindowDays, application.CreatedAt)
observedSeconds = now - windowStart
```

This means a 5-day-old application configured with a 30-day SLO window is only ever measured over its real 5 days of history, not a fabricated 30 - `ObservedDays` in the API response reports exactly how much history was actually used.

## Formulas

### Availability / Uptime
```
downtimeSeconds = duration of the union of [CreatedAt, ResolvedAt-or-now] intervals
                   for every incident with Severity P0 or P1, clipped to [windowStart, now]
Availability % = (observedSeconds - downtimeSeconds) / observedSeconds × 100
```
Overlapping incident intervals are merged before summing, so two simultaneous P0/P1 incidents are never double-counted as more downtime than actually elapsed.

**Why P0/P1 only:** P0 (critical) and P1 (major) represent the service being unavailable to at least some users. P2-P4 are lesser degradations - they still count toward MTTR/MTTA, but are not treated as downtime. This is a documented modeling choice (`downtimeSeverities` in `internal/services/sre_metrics_service.go`), not an arbitrary cutoff.

**Uptime is reported identically to Availability.** This system has no separate synthetic uptime-probe data source (no external pinger, no APM). Presenting a second, independently-computed "uptime" number when both would necessarily derive from the same incident data would not be honest, so the API returns the same value under both keys rather than inventing a second methodology.

### Current SLO
The `TargetPercentage` an operator configured via `PUT /applications/:id/slo` (`ApplicationSLOConfigRequest`). A simple passthrough of stored configuration - not a calculation.

### SLO Compliance
```
Compliant when Availability % ≥ TargetPercentage
```
Requires a configured SLO; otherwise reports "SLO not configured" rather than comparing against a fabricated target.

### Error Budget
```
totalBudgetSeconds = observedSeconds × (1 - TargetPercentage / 100)
remaining % = (totalBudgetSeconds - downtimeSeconds) / totalBudgetSeconds × 100
```
Remaining can go negative (over budget). Requires a configured SLO target (the budget's size is defined by it) and a positive budget (a 100% target has none to track).

### MTTR (Mean Time To Resolve)
```
MTTR = average(ResolvedAt - CreatedAt) across every RESOLVED incident in the window
```
Unresolved incidents are excluded entirely - their resolution time is unknown, never estimated as zero or as "now".

### MTTA (Mean Time To Acknowledge)
```
MTTA = average(AcknowledgedAt - CreatedAt) across every acknowledged incident in the window
```
Incidents that were never acknowledged are excluded, not counted as zero.

## API Surface
- `PUT /applications/:id/slo` - configure `target_percentage` (1-100) and `window_days` (1-365). `slo:manage` permission.
- `GET /applications/:id/slo` - current configuration, or 404 if none exists yet. `slo:read` permission.
- `GET /applications/:id/slo/metrics` - the full computed view (`ApplicationSREMetricsResponse`): availability, uptime, current SLO, compliance, error budget, MTTR, MTTA, `incidentCount`, and `observedDays`. Every `MetricValue` in the response carries its own `formula` string verbatim (the exact text in this document's Formulas section), so the frontend never has to hardcode - and risk drifting from - an explanation of how a number was derived.

## Every SLO/notification action is audited
`ConfigureSLO` writes an `application_slo` audit entry (create or update) with before/after target and window - see `docs/audit` conventions from Phase 23.

## Tradeoffs
- Merging overlapping P0/P1 intervals is correct but adds O(n log n) sort cost per request; acceptable given incident volumes are small relative to metric/log data.
- MTTA depends on operators actually using `PATCH /incidents/:id/acknowledge` - an incident acknowledged only by implication (e.g. commented on) has no MTTA signal, and is correctly excluded rather than guessed at.

## Open Questions
- Whether P1 should always count as full downtime, or only when it also breaches a resource-specific severity threshold - deferred until real incident data suggests the current binary rule is too coarse.
- Whether Alert-derived signals (which lack a reliable `ApplicationID` - see `docs/monitoring/HEALTH_ENGINE.md`) should ever feed into availability once Alert gains a real application FK.

## Related
- `docs/monitoring/HEALTH_ENGINE.md` - the real-time health *score* (different in kind: point-in-time, not a windowed rollup)
- `internal/services/sre_metrics_service.go` - implementation and formula constants
- `internal/services/incident_service.go` - `AcknowledgeIncident`/`AssignIncident`
