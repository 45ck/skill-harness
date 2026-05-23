---
artifactType: model-view
modelId: sh-activity-setup-project
modelKind: activity
method: uml
notation: mermaid
abstractionLevel: runtime
owner: system-modeler
implementationTouchpoints:
  - cmd/skill-harness/main.go
  - cmd/skill-harness/main_test.go
docTouchpoints:
  - docs/developer-artifacts.md
evidenceLinks:
  - cmd/skill-harness/main_test.go
  - .skill-harness/setup-proof.json
reviewRequired: true
updateTriggers:
  - setup-project flag behavior changes
  - generated scaffold file changes
  - setup proof schema changes
driftVerdict: aligned
---

# Setup Project Activity

This activity model tracks the high-level execution path for `skill-harness setup-project`.

## Purpose

Capture the concrete setup-project control flow, including early exits and optional capability branches.

## Scope

The activity includes target resolution, monorepo scope handling, package manager detection, artifact profile and modeling mode resolution, package metadata creation, package installation, Beads setup, Claude settings, developer artifact/model scaffolding, optional `agent-docs` and `noslop` initialization, quality gate installation, setup proof writing, and the install-only early-exit path.

## Source Model

```mermaid
flowchart TD
  Start([Start]) --> Resolve["Resolve target and operation directories"]
  Resolve --> Package["Detect or create package metadata"]
  Package --> Beads{"Beads enabled?"}
  Beads -->|yes| InitBeads["Initialize or record Beads"]
  Beads -->|no| SkipBeads["Record skipped Beads"]
  InitBeads --> Agents
  SkipBeads --> Agents["Render or skip agent settings"]
  Agents --> Artifacts{"Developer artifacts enabled?"}
  Artifacts -->|yes| Scaffold["Write artifact and model scaffold"]
  Artifacts -->|no| Proof["Write setup proof"]
  Scaffold --> Proof
  Proof --> End([Done])
```

## Evidence

The unit tests exercise scaffold creation, package scripts, policy checks, model review generation, and browser opener discovery in temporary repositories.

## Freshness

Update this model when setup-project adds or removes branches, changes default profiles, or changes setup proof contents.
