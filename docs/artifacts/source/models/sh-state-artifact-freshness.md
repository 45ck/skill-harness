---
artifactType: model-view
modelId: sh-state-artifact-freshness
modelKind: state
method: uml
notation: mermaid
abstractionLevel: decision
owner: system-modeler
implementationTouchpoints:
  - docs/artifacts/artifacts.manifest.json
  - scripts/check-artifact-manifest.mjs
  - scripts/generate-model-review.mjs
  - .github/workflows/quality.yml
docTouchpoints:
  - docs/developer-artifacts.md
  - docs/artifacts/source/models/model-inventory.md
evidenceLinks:
  - docs/artifacts/source/model-to-code-human-artifacts-plan-2026-05-23.md
  - .skill-harness/setup-proof.json
reviewRequired: true
updateTriggers:
  - sourceHash policy changes
  - generated HTML review changes
  - CI quality gate changes
driftVerdict: aligned
---

# Artifact Freshness State Model

Artifact freshness is a state machine over source, manifest metadata, generated HTML, and evidence.

## Purpose

Define how a model-backed artifact moves between draft, ready, stale, evidence-deficient, unsafe, and inconclusive states.

## Scope

This state model applies to source-backed artifacts listed in `docs/artifacts/artifacts.manifest.json`, especially UML-first model views.

## Source Model

```mermaid
stateDiagram-v2
  [*] --> Draft
  Draft --> Ready: source + evidence + policy pass
  Ready --> StaleSource: source changes without sourceHash update
  Ready --> StaleReview: generated HTML differs from source
  StaleSource --> Ready: update manifest hash
  StaleReview --> Ready: regenerate review HTML
  Ready --> NeedsEvidence: evidence removed or weakened
  NeedsEvidence --> Ready: add concrete evidence
```

## CI Gate

CI runs manifest checks, model policy checks, generated HTML drift checks, HTML safety checks, suite drift checks, Go tests, Python syntax checks, and a hermetic setup-project smoke.

## Evidence

Evidence comes from manifest source hashes, generated review drift checks, model policy checks, HTML policy checks, and CI workflow results.

## Freshness

Update this model when readiness verdicts, sourceHash behavior, model review generation, or CI gate ordering changes.

| Lifecycle state | Derived drift verdict |
| --- | --- |
| Draft | `needs-source` or `inconclusive` |
| Ready | `aligned` |
| Stale | `source-missing`, `mapping-missing`, `evidence-stale`, or `review-stale` |
| NeedsEvidence | `needs-evidence` |
| Unsafe | `unsafe` |
| Inconclusive | `inconclusive` |
