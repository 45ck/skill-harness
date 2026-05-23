---
name: "multi-agent-workflow-reviewer"
description: "Review or design multi-agent workflows with clear ownership, handoffs, evidence gates, conflict handling, and minimal coordination overhead."
---

# Multi-Agent Workflow Reviewer

Use this skill when work crosses real boundaries that one agent loadout should not own alone.

## Review

- Name the reason multiple agents are needed.
- Assign each agent an owned outcome, not just a topic.
- Keep handoffs artifact-backed: issue id, files, tests, traces, decisions, and unresolved risks.
- Define conflict resolution when agents disagree.
- Avoid parallel agents that inspect the same context and produce duplicate prose.
- Gate adoption with tests, policy checks, review evidence, or human approval where needed.

## Useful Split Patterns

- research/model agent then implementation agent
- implementation agent then adversarial reviewer
- security reviewer beside product or architecture owner
- QA evidence collector then demo/story packager
- workflow engineer then delivery manager for closeout

## Output

### Workflow Goal
### Why Multiple Agents
### Agent Responsibilities
### Handoff Artifacts
### Conflict Rules
### Evidence Gates
### Coordination Risks

