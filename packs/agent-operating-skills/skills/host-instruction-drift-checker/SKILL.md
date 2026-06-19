---
name: "host-instruction-drift-checker"
description: "Check whether host instruction surfaces such as AGENTS.md, CLAUDE.md, Codex agents, Cursor rules, Copilot instructions, and SKILL.md files have drifted apart."
---

# Host Instruction Drift Checker

Use this skill when a repo supports multiple coding-agent hosts or when a baseline update may change agent behavior.

## Process

- Identify the canonical instruction source for the repo.
- Inventory host surfaces:
  - `AGENTS.md`
  - `CLAUDE.md`
  - `AGENT_INSTRUCTIONS.md`
  - `.claude/agents/*`
  - `.codex/agents/*`
  - `.agents/skills/*/SKILL.md`
  - `.claude/skills/*/SKILL.md`
  - `.codex/skills/*/SKILL.md`
  - `.cursor/rules/*.mdc`
  - `.github/copilot-instructions.md`
- Compare only the behavioral contract: scope, permissions, stop rules, evidence, generated-file ownership, and profile/installer semantics.
- Treat host-specific syntax as normal, not drift.
- Classify each mismatch as generated drift, overlay drift, stale doctrine, policy conflict, missing host surface, or intentional host specialization.
- Prefer updating canonical source and regenerating derived surfaces over hand-editing generated agent files.

## Output

### Canonical Source
### Surfaces Checked
### Drift Findings
### Intentional Differences
### Regeneration Path
### Required Manual Review

## Avoid

- forcing every host file to contain identical prose
- editing generated files as if they were canonical
- ignoring Cursor `.mdc` and Copilot surfaces when the repo advertises those hosts
- treating missing evidence gates as minor wording drift
