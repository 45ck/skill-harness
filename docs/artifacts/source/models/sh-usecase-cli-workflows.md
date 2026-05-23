---
artifactType: model-view
modelId: sh-usecase-cli-workflows
modelKind: use-case
method: uml
notation: mermaid
abstractionLevel: domain
owner: system-modeler
implementationTouchpoints:
  - cmd/skill-harness/main.go
  - README.md
docTouchpoints:
  - docs/developer-artifacts.md
  - AGENT_INSTRUCTIONS.md
evidenceLinks:
  - cmd/skill-harness/main_test.go
  - docs/developer-artifacts.md
reviewRequired: true
updateTriggers:
  - new CLI command or flag
  - setup-project behavior change
  - render or install behavior change
driftVerdict: aligned
---

# CLI Workflow Use Cases

The CLI supports maintainers and agents that need to inspect, install, render, check, and scaffold project capabilities.

## Purpose

Show the user goals that the CLI exposes to repo maintainers and agent operators.

## Scope

Top-level use cases are install full suite, install selected packs, install selected agents, run interactive install, set up a project, validate dependencies, render agents, install Beads worktrees, uninstall agents, and open artifact review. Flag variants are extensions unless they materially change behavior.

## Source Model

```mermaid
flowchart TD
  Maintainer["Maintainer or agent"] --> Install["Install suite"]
  Maintainer --> Render["Render agents"]
  Maintainer --> Check["Check installed skills and templates"]
  Maintainer --> List["List agents and packs"]
  Maintainer --> Setup["Setup project"]
  Setup --> Artifacts["Scaffold developer artifacts"]
  Setup --> Models["Scaffold model review policy"]
  Setup --> Beads["Initialize or respect Beads"]
```

## Use-Case Notes

`setup-project` is the workflow most tightly coupled to developer artifacts. It creates repo-local policy, scripts, proof files, source directories, and review surfaces while respecting flags that skip Beads, agent-docs, Claude settings, or modeling.

## Evidence

Evidence comes from `cmd/skill-harness/main.go`, CLI tests, `README.md`, and `docs/developer-artifacts.md`.

## Freshness

Update this model when commands, setup flags, package scripts, or artifact-opening behavior change.
