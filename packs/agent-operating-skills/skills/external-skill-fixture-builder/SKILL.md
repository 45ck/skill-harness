---
name: "external-skill-fixture-builder"
description: "Turn external skill, rule, plugin, MCP, task-memory, and multi-agent workspace patterns into safe fixture coverage rather than live dependencies."
---

# External Skill Fixture Builder

Use this skill when a third-party ecosystem pattern should be tested by `skill-harness` but should not become a live dependency.

## Process

- Choose the fixture lane by install posture, not by brand:
  - pure `SKILL.md` pack
  - plugin manifest
  - Cursor `.mdc` rules
  - Codex/Claude/Cursor subagents
  - MCP/tool config
  - generated record/replay skill
  - Beads or task-memory repo
  - multi-agent workspace layout
- Use minimal synthetic fixtures when possible.
- Use pinned external snapshots only when real structure matters and license/provenance allows it.
- Keep fixtures out of global user directories and live install caches.
- Make the fixture assert one behavior: parse, classify, trust recommendation, drift detection, render target, or installer opt-in.
- Add a command or check that proves the fixture is exercised.

## Output

### Fixture Lane
### Source Or Synthetic Basis
### Files Included
### Expected Classification
### Risk Flags Expected
### Check Command
### Follow-Up Adoption Question

## Avoid

- making fixtures large mirrors of public repos
- using fixtures as a back door for third-party adoption
- mixing fixture coverage with default installer behavior
- storing credentials, real traces, browser recordings, or private task state in fixtures
