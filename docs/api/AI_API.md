# AI API

## Current Implementation
No AI HTTP endpoints are implemented.

## Future Architecture
### Endpoint: POST /api/v1/ai/query
- Purpose: ask operational questions across incidents/resources/telemetry.
- Authentication: required.
- Request: `{ projectId, query, contextHints }`.
- Response: `{ answer, evidence, confidence, actions[] }`.
- Validation: project access and payload limits.
- Errors: 400, 401, 403, 429, 500.
- Future extensions: streaming responses.

### Endpoint: POST /api/v1/ai/incidents/:id/analysis
- Purpose: generate incident-specific analysis and RCA candidates.
- Authentication: required.
- Request: incident id + optional scope controls.
- Response: ranked hypotheses with evidence.
- Future extensions: remediation playbook generation.

### Endpoint: POST /api/v1/ai/feedback
- Purpose: capture operator feedback on AI quality.
- Future extensions: model and retrieval tuning.

## Related
- `docs/ai/AI_ASSISTANT.md`
- `docs/ai/RAG.md`
