# Vector Database

## Architecture
### Current Implementation
- Not implemented.

### Future Architecture
- Dedicated vector store with metadata filters: project, environment, incident_id, source_type.

## Embeddings
- Index embedded chunks for incidents, comments, runbooks, and architecture docs.

## Context Retrieval
- ANN search with hybrid lexical reranking.

## Prompt Strategy
- Return chunk text plus metadata for citation-ready prompts.

## Conversation Memory
- Store short-lived conversation vectors for session relevance.

## Incident Analysis
- Similar-incident nearest-neighbor lookup.

## Root Cause Analysis
- Fetch correlated historical failure patterns.

## Future AI Agents
- Index maintenance agent and retrieval quality evaluator.

## Tradeoffs
- High recall indexes can increase false positives without reranking.

## Open Questions
- Vendor choice and tenancy isolation model.
- Reindex cadence for mutable operational data.

## Related
- `docs/ai/RAG.md`
- `docs/database/DATA_MODEL.md`
