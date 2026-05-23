---
name: "context-engineering-planner"
description: "Plan the context, artifacts, memory, retrieval, and compression surfaces needed for an agent workflow or long-horizon task."
---

# Context Engineering Planner

Use this skill when an agent workflow needs deliberate context design instead of a large undifferentiated prompt.

## Plan

- Identify the decision the agent must make and the evidence it needs.
- Split context into durable source, transient working context, generated evidence, and memory candidates.
- Prefer repo files, issues, manifests, tests, traces, and docs over prose recollection.
- Define what may be retrieved automatically and what must be explicitly requested.
- Define what should be compressed, summarized, or excluded.
- Name stale-context risks and refresh checks.

## Context Classes

- `source`: canonical specs, code, issues, policies, manifests
- `evidence`: tests, logs, screenshots, traces, evaluation results
- `memory`: reusable lessons worth recording with the project memory mechanism
- `working`: temporary notes used only for the current run
- `excluded`: secrets, private data, raw traces, irrelevant history, stale generated output

## Output

### Workflow Goal
### Required Source Context
### Evidence Inputs
### Retrieval Plan
### Compression Plan
### Memory Candidates
### Exclusions
### Refresh Checks

