---
artifactType: implementation-plan
artifactId: generic-infographic-artifact-workflow-plan-2026-05-24
owner: codex
issue: skill-harness-okh
status: ready
reviewRequired: true
evidenceLinks:
  - AGENTS.md
  - docs/developer-artifacts.md
  - scripts/generate-artifact-review.mjs
  - scripts/check-artifact-manifest.mjs
  - scripts/check-artifact-html-policy.mjs
  - cmd/skill-harness/main.go
freshness:
  generatedAt: 2026-05-24
  sourceFirst: true
---

# Generic Infographic Artifact Workflow Plan

## Purpose

Make source-backed human HTML artifacts a first-class workflow for non-model artifacts, including discovery, planning, product, business, data, research, UX, and mockup review work.

## Decision

Add a generic artifact review generator alongside the existing model review generator. Model review remains specialized for UML/C4/model artifacts. Generic review handles non-model artifacts that opt in through `reviewRequired: true` or the generic renderer metadata.

## Current Gap

The repo already has source-first artifact policy and generated model HTML, but non-model artifacts rely on manual HTML. This lets agents complete Beads discovery and textual handoff without producing the human review surface that the artifact doctrine expects.

## Target Workflow

1. Create or update canonical source under `docs/artifacts/source/<family>/`.
2. Add a manifest entry with `reviewRequired: true`, owner, evidence links, and status.
3. Run `node scripts/generate-artifact-review.mjs`.
4. Run `npm run artifacts:check`.
5. Surface or open the generated HTML review page in the handoff.

## Interaction Policy

The default lane is CSS-only. It permits radio tabs, details/summary, anchor navigation, and inline SVG states. Inline JavaScript remains disabled; the reviewed inline-JS lane is reserved until manifest metadata, human approval, CSP support, and checker coverage for blocked browser APIs are implemented together.

## Generated Surface Standard

Every generic artifact review page should show:

- review verdict and purpose
- artifact family, status, owner, and freshness
- evidence counts and source-depth indicators
- source-to-review flow
- evidence links and update triggers
- canonical source preview
- source hash and renderer metadata

## Implementation Scope

- Add `scripts/generate-artifact-review.mjs`.
- Add `artifacts:generate` and `artifacts:review` package scripts.
- Include the generator in scaffolded project output.
- Extend manifest checks for `reviewRequired: true`.
- Update opener discovery to prefer `generated/review/index.html` and allow artifact ids.
- Update agent and skill guidance so human-facing discovery creates source plus HTML by default.

## Acceptance Criteria

- Non-model artifacts with `reviewRequired: true` generate deterministic HTML.
- `--check` fails when generated HTML or manifest renderer metadata is stale.
- `npm run artifacts:check` includes generic artifact drift checks.
- Generated HTML passes the no-script HTML policy.
- Scaffolded repos receive the same generator, scripts, and guidance.
