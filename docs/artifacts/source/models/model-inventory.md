---
artifactType: model-inventory
owner: system-modeler
reviewRequired: true
evidenceLinks:
  - docs/developer-artifacts.md
  - scripts/agent_loadouts.json
  - scripts/dependencies.json
  - cmd/skill-harness/main.go
  - cmd/skill-harness/main_test.go
---

# Skill Harness Model Inventory

This inventory is the human index for source-backed models owned by the `skill-harness` repo. Each model below has a stable `modelId`, a canonical Markdown source, implementation and documentation touchpoints, evidence links, and an HTML review surface generated under `generated/review/models/`.

## Purpose

Keep the repo's canonical models discoverable and tied to implementation touchpoints so model impact can be checked on every engineering change.

## Scope

This inventory covers the baseline models for the `skill-harness` suite entrypoint. It does not replace external pack documentation or target project models.

| Model ID | Kind | Method | Owner | Source | Touchpoints | Evidence | Review surface |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `sh-system-context` | context | c4 | system-modeler | `docs/artifacts/source/models/sh-system-context.md` | `cmd/skill-harness/main.go`, `scripts/dependencies.json`, `scripts/agent_loadouts.json`, `packs/`, `README.md`, `docs/developer-artifacts.md` | `AGENTS.md`, `scripts/dependencies.json`, `scripts/agent_loadouts.json` | `generated/review/models/sh-system-context.html` |
| `sh-domain-artifact-governance` | domain | uml | system-modeler | `docs/artifacts/source/models/sh-domain-artifact-governance.md` | `.skill-harness/project.json`, `docs/artifacts/artifacts.manifest.json`, `scripts/check-artifact-manifest.mjs`, `scripts/check-model-artifact-policy.mjs`, `scripts/check-artifact-html-policy.mjs`, `scripts/open-artifact-review.mjs`, `scripts/generate-artifact-review.mjs`, `scripts/generate-model-review.mjs`, `docs/developer-artifacts.md`, `docs/artifacts/source/models/model-inventory.md` | `docs/artifacts/source/model-to-code-human-artifacts-plan-2026-05-23.md`, `docs/developer-artifacts.md`, `cmd/skill-harness/main.go` | `generated/review/models/sh-domain-artifact-governance.html` |
| `sh-usecase-cli-workflows` | use-case | uml | system-modeler | `docs/artifacts/source/models/sh-usecase-cli-workflows.md` | `cmd/skill-harness/main.go`, `README.md`, `docs/developer-artifacts.md`, `AGENT_INSTRUCTIONS.md` | `cmd/skill-harness/main_test.go`, `docs/developer-artifacts.md` | `generated/review/models/sh-usecase-cli-workflows.html` |
| `sh-activity-setup-project` | activity | uml | system-modeler | `docs/artifacts/source/models/sh-activity-setup-project.md` | `cmd/skill-harness/main.go`, `cmd/skill-harness/main_test.go`, `docs/developer-artifacts.md` | `cmd/skill-harness/main_test.go`, `.skill-harness/setup-proof.json` | `generated/review/models/sh-activity-setup-project.html` |
| `sh-component-scaffold-engine` | component | c4 | system-modeler | `docs/artifacts/source/models/sh-component-scaffold-engine.md` | `cmd/skill-harness/main.go`, `scripts/suite_graph.py`, `scripts/render_suite_docs.py`, `scripts/check_suite_drift.py`, `scripts/check-artifact-manifest.mjs`, `scripts/check-model-artifact-policy.mjs`, `scripts/check-artifact-html-policy.mjs`, `scripts/generate-artifact-review.mjs`, `scripts/generate-model-review.mjs`, `docs/agent-loadouts.md`, `docs/developer-artifacts.md` | `cmd/skill-harness/main_test.go`, `scripts/dependencies.json`, `scripts/agent_loadouts.json` | `generated/review/models/sh-component-scaffold-engine.html` |
| `sh-state-artifact-freshness` | state | uml | system-modeler | `docs/artifacts/source/models/sh-state-artifact-freshness.md` | `docs/artifacts/artifacts.manifest.json`, `scripts/check-artifact-manifest.mjs`, `scripts/generate-artifact-review.mjs`, `scripts/generate-model-review.mjs`, `scripts/check-artifact-html-policy.mjs`, `scripts/open-artifact-review.mjs`, `.github/workflows/quality.yml`, `docs/developer-artifacts.md`, `docs/artifacts/source/models/model-inventory.md` | `docs/artifacts/source/model-to-code-human-artifacts-plan-2026-05-23.md`, `.skill-harness/setup-proof.json` | `generated/review/models/sh-state-artifact-freshness.html` |
| `sh-uwe-navigation-atlas` | component | uwe | system-modeler | `docs/artifacts/source/models/sh-uwe-navigation-atlas.md` | `cmd/skill-harness/main.go`, `scripts/generate-artifact-review.mjs`, `scripts/check-artifact-manifest.mjs`, `docs/artifacts/artifacts.manifest.json` | `docs/artifacts/source/product/e2e-product-system-atlas-workflow-2026-05-24.md`, `docs/developer-artifacts.md` | `generated/review/models/sh-uwe-navigation-atlas.html` |
| `sample-uwe-navigation` | component | uwe | system-modeler | `docs/artifacts/source/models/sample-uwe-navigation.md` | `docs/artifacts/source/product/sample-uwe-screenshot-atlas.md`, `generated/review/evidence/sample-uwe-atlas/landing.svg`, `generated/review/evidence/sample-uwe-atlas/auth.svg`, `generated/review/evidence/sample-uwe-atlas/dashboard.svg` | `docs/artifacts/source/product/sample-uwe-screenshot-atlas.md` | `generated/review/models/sample-uwe-navigation.html` |

## Evidence

The inventory is backed by the artifact manifest, scaffold code, developer artifact documentation, and suite loadout inputs.

## Freshness

Regenerate model review HTML and rerun artifact checks after any listed touchpoint changes.

## Changelog

| Date | Change | Evidence |
| --- | --- | --- |
| 2026-05-23 | Created self-modeling baseline with six model views. | `docs/artifacts/source/model-to-code-human-artifacts-plan-2026-05-23.md`, `cmd/skill-harness/main_test.go` |
| 2026-05-24 | Added UWE navigation atlas model for screenshot-backed whole-app inspection. | `docs/artifacts/source/product/e2e-product-system-atlas-workflow-2026-05-24.md`, `docs/artifacts/source/models/sh-uwe-navigation-atlas.md` |
| 2026-05-24 | Added synthetic UWE screenshot atlas example. | `docs/artifacts/source/product/sample-uwe-screenshot-atlas.md`, `docs/artifacts/source/models/sample-uwe-navigation.md` |

## Update Rule

When code, scripts, package metadata, agent loadouts, embedded packs, or artifact policy changes, run:

```bash
npm run models:impact -- --changed
npm run models:review
npm run artifacts:check
```

If a changed file maps to a model touchpoint, update that model source or document why the model remains valid in the related issue or change record.
