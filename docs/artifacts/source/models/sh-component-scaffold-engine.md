---
artifactType: model-view
modelId: sh-component-scaffold-engine
modelKind: component
method: c4
notation: mermaid
abstractionLevel: design
owner: system-modeler
implementationTouchpoints:
  - cmd/skill-harness/main.go
  - scripts/suite_graph.py
  - scripts/render_suite_docs.py
  - scripts/check_suite_drift.py
  - scripts/check-artifact-manifest.mjs
  - scripts/check-model-artifact-policy.mjs
  - scripts/generate-model-review.mjs
docTouchpoints:
  - docs/agent-loadouts.md
  - docs/developer-artifacts.md
evidenceLinks:
  - cmd/skill-harness/main_test.go
  - scripts/dependencies.json
  - scripts/agent_loadouts.json
reviewRequired: true
updateTriggers:
  - scaffold script changes
  - suite graph schema changes
  - agent template rendering changes
driftVerdict: aligned
---

# Scaffold Engine Component View

The scaffold engine combines Go CLI behavior with repo-local scripts. Go writes target project defaults; scripts check and render repo-local evidence.

## Purpose

Show the design-level components that collaborate to scaffold and verify suite outputs.

## Scope

Included components are the CLI command router, dependency/loadout readers, install orchestrator, project setup orchestrator, developer artifact scaffold writer, model policy/review script emitters, Beads worktree wrapper installer, Python helper scripts, agent template sources, target repo filesystem, and external command dependencies.

## Source Model

```mermaid
flowchart LR
  CLI["cmd/skill-harness"] --> Scaffold["Developer artifact scaffold"]
  Scaffold --> ProjectConfig[".skill-harness/project.json"]
  Scaffold --> PackageScripts["package.json scripts"]
  Scaffold --> PolicyScripts["artifact and model policy scripts"]
  PolicyScripts --> Manifest["docs/artifacts/artifacts.manifest.json"]
  PolicyScripts --> ReviewHTML["generated/review/models/*.html"]
  SuiteScripts["suite graph scripts"] --> LoadoutDocs["docs/agent-loadouts.md"]
```

## Responsibility Split

Go owns portable project setup. Node scripts own artifact and HTML checks because target projects commonly already have Node for package scripts. Python scripts own suite graph generation because existing suite maintenance scripts are Python.

## Evidence

Evidence comes from the Go CLI, Node artifact scripts, Python suite scripts, loadout JSON, dependency JSON, and generated agent templates.

## Freshness

Update this model when scaffold writers, suite graph scripts, agent rendering scripts, or target repo output contracts change.
