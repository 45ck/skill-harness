---
artifactType: model-view
modelId: sh-domain-artifact-governance
modelKind: domain
method: uml
notation: mermaid
abstractionLevel: domain
owner: system-modeler
implementationTouchpoints:
  - .skill-harness/project.json
  - docs/artifacts/artifacts.manifest.json
  - scripts/check-artifact-manifest.mjs
  - scripts/check-model-artifact-policy.mjs
  - scripts/generate-model-review.mjs
docTouchpoints:
  - docs/developer-artifacts.md
  - docs/artifacts/source/models/model-inventory.md
evidenceLinks:
  - docs/artifacts/source/model-to-code-human-artifacts-plan-2026-05-23.md
  - docs/developer-artifacts.md
reviewRequired: true
updateTriggers:
  - artifact manifest schema changes
  - HTML policy changes
  - model metadata contract changes
driftVerdict: aligned
---

# Artifact Governance Domain Model

The artifact domain keeps canonical source, generated review surfaces, and evidence separate so humans can review rich material without losing diffable source control.

## Purpose

Define the vocabulary and invariants for source-backed developer artifacts and model views.

## Scope

This is a governance/domain model, not a Go type map. It covers artifact profile, modeling mode, model inventory, manifest artifacts, review surfaces, setup proof, evidence links, update triggers, and ownership.

## Source Model

```mermaid
classDiagram
  class Artifact {
    id
    type
    status
    source
    reviewSurface
  }
  class ModelView {
    modelId
    modelKind
    method
    notation
    abstractionLevel
  }
  class SourceFile {
    path
    sourceHash
  }
  class EvidenceLink {
    pathOrIssue
  }
  class ReviewSurface {
    htmlPath
    generated
  }
  Artifact <|-- ModelView
  Artifact --> SourceFile
  Artifact --> EvidenceLink
  Artifact --> ReviewSurface
```

## Invariants

- Every ready artifact names a canonical source.
- Every ready artifact has concrete evidence.
- Every model view records method, notation, owner, touchpoints, and freshness.
- HTML review files are generated from source and checked for drift.
- Host-specific opening is transport only: `open-artifact-review.mjs` resolves the review target, while Codex Browser, Claude preview, system browser, or a local HTTP server provides the human viewing surface.

## Evidence

Evidence comes from `docs/developer-artifacts.md`, `.skill-harness/project.json`, the manifest, and the model-to-code planning artifact.

## Freshness

Update this model when artifact manifest schema, model metadata requirements, HTML safety policy, or generated review semantics change.
