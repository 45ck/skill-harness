---
name: "agent-memory-design-reviewer"
description: "Review agent memory, handoff, retrieval, and learning designs for usefulness, provenance, privacy, staleness, and poisoning risk."
---

# Agent Memory Design Reviewer

Use this skill when a project wants agents to remember lessons, retrieve prior work, or improve future runs.

## Review

- Separate durable memory from run-local notes.
- Require source links for durable memory.
- Define freshness and revocation behavior.
- Exclude secrets, credentials, private logs, customer data, raw traces, and unsupported claims.
- Treat self-improvement as a proposal until verified by tests, reviews, or policy checks.
- Check poisoning paths: untrusted docs, generated content, user content, web content, issue comments, and previous agent outputs.

## Memory Verdicts

- `record`: durable, sourced, reusable, safe
- `summarize`: useful but too noisy or sensitive in raw form
- `run-local`: useful only for the current task
- `reject`: stale, unsafe, unsupported, or too specific

## Output

### Verdict
### Memory Items
### Source Links
### Retrieval Rules
### Revocation / Freshness
### Poisoning Risks
### Required Redactions
