---
name: "autonomy-boundary-checker"
description: "Review an agent workflow for where it may act autonomously, where it must ask, and where human approval is required."
---

# Autonomy Boundary Checker

Use this skill before allowing an agent to mutate files, run tools, publish, deploy, approve, or learn from its own output.

## Checks

- The intended autonomous actions are named.
- Reversible actions are separated from irreversible actions.
- External side effects are explicit.
- Tool and data permissions match the task.
- Human approval gates are named before risky actions.
- The agent has a stop condition and escalation path.
- Learning outputs are proposals until validated.

## Human Approval Required

- secret or credential handling
- production data access
- permission expansion
- destructive filesystem, database, or infrastructure actions
- network-visible publishing
- autonomous merge, deployment, or account changes
- security, privacy, legal, or compliance policy changes

## Output

### Verdict
Use `safe-to-act`, `ask-before-action`, `proposal-only`, or `blocked`.

### Autonomous Actions
### Ask-First Actions
### Human-Approval Actions
### Stop Conditions
### Required Policy Changes

