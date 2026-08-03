# AI Memory Model

## Architecture
### Current Implementation
- No AI memory subsystem in app runtime.

### Future Architecture
- Session memory for active investigations.
- Long-term memory for validated incident learnings.

## Embeddings
- Memory entries indexed for retrieval by incident and project scope.

## Context Retrieval
- Prioritize recent and high-confidence memories.

## Prompt Strategy
- Include compact memory summaries with provenance and timestamps.

## Conversation Memory
- Session-scoped thread memory with TTL and deletion controls.

## Incident Analysis
- Reuse prior remediation outcomes and failure signatures.

## Root Cause Analysis
- Compare current signal pattern against historical RCAs.

## Future AI Agents
- Memory curator agent to prune stale or low-quality memories.

## Tradeoffs
- Long memory improves continuity but risks stale guidance.

## Open Questions
- Memory approval workflow and human override model.

## Related
- `docs/ai/RAG.md`
- `docs/ai/PROMPTS.md`
