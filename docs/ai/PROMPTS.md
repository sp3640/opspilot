# Prompt Strategy

## Architecture
### Current Implementation
- No production prompt orchestration.

### Future Architecture
- Prompt templates by workflow: triage, timeline summary, RCA, action planning.

## Embeddings
- Prompt context will include retrieved snippets with source metadata.

## Context Retrieval
- Hard project scoping and recency constraints.

## Prompt Strategy
- System prompt: safety, scope, and action constraints.
- Task prompt: question-specific instructions.
- Context prompt: evidence snippets with provenance.

## Conversation Memory
- Session memory includes prior operator questions and accepted recommendations.

## Incident Analysis
- Prompt includes severity/status/timeline snapshots.

## Root Cause Analysis
- Prompt asks for hypotheses, evidence, and uncertainty explicitly.

## Future AI Agents
- Prompt quality evaluator and regression test agent.

## Tradeoffs
- More constraints improve reliability but can reduce creativity.

## Open Questions
- Prompt versioning and rollout mechanism.

## Related
- `docs/ai/AI_ASSISTANT.md`
- `docs/ai/MEMORY.md`
