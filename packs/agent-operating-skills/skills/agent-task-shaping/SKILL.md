---
name: "agent-task-shaping"
description: "Shape ambiguous or frontier-agent work into a bounded agent task with outcome, scope, context, tools, autonomy level, and verification evidence."
---

# Agent Task Shaping

Use this skill before launching an agentic implementation, research, review, or long-running workflow from a vague request.

## Process

- State the user-visible outcome in one sentence.
- Name the smallest reversible slice that can produce evidence.
- Separate required context from optional background.
- Choose the agent loadout or role that can own the slice end to end.
- Identify allowed tools, blocked tools, and any permission expansion that needs approval.
- Define the autonomy level: act directly, ask before mutation, ask before external side effects, or propose only.
- Define verification evidence before work starts.

## Output

### Task
### Scope
### Context Needed
### Agent / Loadout
### Tool Boundaries
### Autonomy Level
### Verification Evidence
### Follow-Up Capture

## Avoid

- turning broad strategy into unbounded execution
- starting from tool availability instead of task outcome
- asking a human for decisions the agent can resolve from repository evidence
- adopting learnings without a traceable issue, test, policy check, or run receipt
