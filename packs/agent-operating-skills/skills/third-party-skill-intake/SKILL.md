---
name: "third-party-skill-intake"
description: "Review third-party skill, agent, rule, and plugin repos for format fit, provenance, and first-party adoption paths without treating them as direct harness dependencies."
---

# Third-Party Skill Intake

Use this skill when reviewing a public skill, agent, Cursor rule, plugin, MCP, or agent-workflow repository for possible 45ck adoption.

## Process

- Classify the repo type: official reference, curated index, skill pack, rule pack, plugin, MCP/tool pack, workflow framework, task-memory system, or runtime platform.
- Treat stars, forks, and social activity as discovery signals only.
- Verify license, provenance, helper scripts, tool configuration, and install posture before recommending adoption.
- Inventory actual surfaces: `SKILL.md`, `.cursor/rules/*.mdc`, `.claude/agents`, `.codex/agents`, plugin manifests, MCP configs, lockfiles, install scripts, generated metadata, and task-state files.
- Compare the repo against the first-party 45ck catalog before proposing new skills.
- Map every useful idea to one destination:
  - first-party rewrite under `packs/`
  - harness helper or installer feature
  - fixture/test case only
  - reference/monitor only
  - reject/quarantine

## Output

### Repo Type
### Surfaces Found
### License And Provenance
### Risk Flags
### Catalog Overlap
### Adoption Destination
### Rewrite Requirements
### Fixture Recommendation

## Avoid

- copying whole public catalogs into the harness
- treating README claims as proof of safe behavior
- adding third-party repos to `scripts/dependencies.json` without an explicit ownership decision
- installing unreviewed skills globally
- running helper scripts from public repos inside the live install path
