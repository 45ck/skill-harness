---
name: "self-improving-agent-loop"
description: "Design governed self-improving agent loops that sense work, model failure modes, act in small slices, gate evidence, and turn learnings into tracked proposals."
---

# Self-Improving Agent Loop

Use this skill when a project wants agents to improve a workflow, skill, prompt, loadout, checker, or operating policy over time.

## Loop

1. Sense: collect issues, diffs, tests, traces, artifacts, user feedback, and handoff notes.
2. Model: identify the task type, capability assumption, failure mode, and quality bar.
3. Plan: choose one reversible improvement with explicit evidence.
4. Act: change the smallest useful surface using existing project patterns.
5. Gate: run tests, artifact checks, permission checks, and review gates.
6. Learn: record a follow-up issue, durable memory, skill update, or checker proposal only when evidence supports it.

## Governance

- Treat frontier models as capable of planning, synthesis, review, and context work when the digital surfaces are available.
- Keep deterministic scaffolding for repeatable validation, policy checks, manifests, and install wiring.
- Treat generated traces and self-assessments as evidence candidates, not proof.
- Require human approval for permission expansion, destructive actions, production data, publishing, merge, deployment, and policy changes.
- Keep domain-specific loops out of the core pack unless the pattern generalizes across projects.

## Output

### Loop Goal
### Sensors
### Failure Model
### Reversible Action
### Gates
### Human Approval Boundaries
### Learning Output
### Next Issue Or Memory
