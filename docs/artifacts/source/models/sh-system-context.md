---
artifactType: model-view
modelId: sh-system-context
modelKind: context
method: c4
notation: mermaid
abstractionLevel: runtime
owner: system-modeler
implementationTouchpoints:
  - cmd/skill-harness/main.go
  - scripts/dependencies.json
  - scripts/agent_loadouts.json
  - scripts/external_skill_intake.py
  - scripts/check_external_skill_intake.py
  - plugins/skill-harness-helpers/
  - packs/
docTouchpoints:
  - README.md
  - docs/developer-artifacts.md
  - docs/agent-operating-skills.md
  - docs/third-party-skill-intake.md
evidenceLinks:
  - AGENTS.md
  - scripts/dependencies.json
  - scripts/agent_loadouts.json
  - tests/fixtures/external-skill-intake/
reviewRequired: true
updateTriggers:
  - CLI command surface changes
  - agent loadout or pack boundary changes
  - developer artifact capability changes
  - external skill intake policy or scanner changes
driftVerdict: aligned
---

# Skill Harness System Context

`skill-harness` is the suite entrypoint for installing and rendering the 45ck agent and skill stack into target environments. It coordinates local pack metadata, external dependency references, Codex and Claude agent templates, optional Codex helper-plugin skills, Beads-aware project setup, agent-native bootstrap overlays, external skill ecosystem intake, and source-backed developer artifacts, including visual-source-first product, business, data, research, UX, and model review surfaces.

## Purpose

Show the system boundary around `skill-harness` as an installer, renderer, and scaffold engine.

## Scope

Included actors and externals are maintainers, agents, target repos, package managers, `agent-docs`, `noslop`, `bd`, external pack repos, public skill/rule/plugin/MCP ecosystems, published loop catalogs, embedded packs, optional helper plugins, home agent directories, repo-local `.skill-harness/` state, test fixtures, and generated artifact directories. Embedded packs include core toolkit packs such as `specgraph-skills` and `noslop-skills` plus suite-local workflow packs. Target repo runtime behavior is out of scope.

## Source Model

```mermaid
flowchart LR
  Human["Human maintainer"] --> CLI["skill-harness CLI"]
  Agent["Codex or Claude agent"] --> CLI
  CLI --> Config["Repo configuration and pack metadata"]
  CLI --> Target["Target project workspace"]
  PublicEcosystems["Public skill, rule, plugin, MCP, and task-memory repos"] --> Intake["External skill intake scanner and fixtures"]
  LoopCatalogs["Published loop catalogs"] --> Helpers["Optional Codex helper plugin skills"]
  Helpers --> Agent
  Intake --> Config
  Target --> Stack[".skill-harness agent stack overlay, lock, and setup proof"]
  Stack --> CLI
  Config --> Agents["Rendered agent loadouts"]
  Config --> Artifacts["Developer artifact policy"]
  Target --> Review["Human review surfaces"]
```

## Boundary

The harness owns suite setup, rendering, resolved agent-stack locks, setup proof, repo-local artifact policy, optional helper-plugin guidance, and synthetic external ecosystem fixtures. Target projects own their canonical product, business, data, research, UX, model, and generated evidence sources. Generated HTML is review material, not canonical source. Public third-party ecosystems and published loop catalogs remain outside the live dependency path until a reviewed first-party rewrite or explicit fixture decision exists; catalog prompts are reference data and do not grant runtime authority.

## Evidence

Evidence comes from `AGENTS.md`, `scripts/dependencies.json`, `scripts/agent_loadouts.json`, `plugins/skill-harness-helpers/`, the intake scanner fixtures, and the setup-project implementation.

## Freshness

Update this model when CLI command boundaries, pack dependencies, helper plugin behavior, agent rendering behavior, external ecosystem intake behavior, or developer artifact policy changes.
