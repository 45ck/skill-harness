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
  - scripts/generate-artifact-review.mjs
  - scripts/check-artifact-html-policy.mjs
  - scripts/open-artifact-review.mjs
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
  - generic artifact review generation changes
  - generated HTML or visual review changes
  - visual-source-first policy changes
  - CI quality gate changes
driftVerdict: aligned
---

# Artifact Freshness State Model

Artifact freshness is a state machine over source, manifest metadata, generated visual review surfaces, and evidence.

## Purpose

Define how a source-backed artifact moves between draft, ready, stale, evidence-deficient, unsafe, and inconclusive states.

## Scope

This state model applies to source-backed artifacts listed in `docs/artifacts/artifacts.manifest.json`, including UML-first model views and visual-source-first product, E2E product system atlas, business, data, research, UX, planning, discovery, and mockup artifacts.

## Source Model

```mermaid
stateDiagram-v2
  [*] --> Draft
  Draft --> Ready: source + evidence + policy pass
  Ready --> StaleSource: source changes without sourceHash update
  Ready --> StaleReview: generated visual review differs from source
  StaleSource --> Ready: update manifest hash
  StaleReview --> Ready: regenerate visual review
  Ready --> NeedsReviewSurface: reviewRequired true without generated HTML
  NeedsReviewSurface --> Ready: generate artifact review surface
  Ready --> NeedsEvidence: evidence removed or weakened
  NeedsEvidence --> Ready: add concrete evidence
  Ready --> NeedsFidelity: UI/product review lacks high-fidelity surface
  NeedsFidelity --> Ready: generate high-fidelity review
  Ready --> NeedsNavigationEvidence: UWE atlas lacks node/action evidence
  NeedsNavigationEvidence --> Ready: add screenshots, QA verdicts, side effects
  Ready --> Unsafe: review leaks secret or private data
  Unsafe --> Ready: redact and rerun policy checks
```

## CI Gate

CI runs manifest checks, model policy checks, generated review drift checks, HTML safety checks, suite drift checks, Go tests, Python syntax checks, and a hermetic setup-project smoke.

## Evidence

Evidence comes from manifest source hashes, generic artifact review drift checks, generated model review drift checks, model policy checks, visual-source policy metadata, HTML policy checks, and CI workflow results.

## Freshness

Update this model when readiness verdicts, sourceHash behavior, generic artifact review generation, model review generation, visual-source-first policy, E2E atlas evidence policy, or CI gate ordering changes.

| Lifecycle state | Derived drift verdict |
| --- | --- |
| Draft | `needs-source` or `inconclusive` |
| Ready | `aligned` |
| Stale | `source-missing`, `mapping-missing`, `evidence-stale`, or `review-stale` |
| NeedsEvidence | `needs-evidence` |
| NeedsFidelity | `mapping-missing` or `review-stale` |
| NeedsReviewSurface | `review-stale` |
| NeedsNavigationEvidence | `needs-evidence` |
| Unsafe | `unsafe` |
| Inconclusive | `inconclusive` |
