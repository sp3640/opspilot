Implement PHASE 25: OpsPilot Intelligence / Root Cause Analysis.

IMPORTANT:
This is the final intelligence layer.
Do not implement it by simply generating generic AI text.

The system should first collect structured evidence:

Incident
 ↓
Alerts
 ↓
Metrics
 ↓
Logs
 ↓
Deployments
 ↓
Kubernetes events
 ↓
Pod health
 ↓
Recent configuration changes
 ↓
Audit events

Build a deterministic evidence/correlation layer first.

For an incident, produce:

1. Summary
2. Affected application
3. Timeline
4. Recent changes
5. Correlated alerts
6. Relevant metrics
7. Relevant logs
8. Kubernetes evidence
9. Possible causes
10. Recommended investigation steps
11. Recommended remediation
12. Confidence/evidence level

Example:

Possible cause:

Deployment v1.8.2 may be related to the incident.

Evidence:
- deployment occurred 7 minutes before degradation
- error rate increased afterward
- memory increased
- pods restarted
- logs contain OOM-related errors

Recommendation:
Investigate or rollback v1.8.2.

IMPORTANT:
Never claim certainty when the evidence only shows correlation.

Use:
Possible cause
Potentially related
Evidence suggests

instead of:
Root cause confirmed

If an LLM/AI provider is introduced:
- isolate it behind a service interface
- never send secrets
- never send credentials
- minimize sensitive data
- log AI recommendations
- clearly label AI-generated recommendations
- never allow autonomous destructive remediation

The final UI should present:

Incident Intelligence

Summary
Evidence
Timeline
Possible Causes
Recommended Actions

Human engineers remain responsible for approving remediation.

Add deterministic tests for evidence correlation.

Do not make AI-generated recommendations the authorization mechanism.