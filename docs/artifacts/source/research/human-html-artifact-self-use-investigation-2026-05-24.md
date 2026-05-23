---
artifactType: research-synthesis
artifactId: human-html-artifact-self-use-investigation-2026-05-24
owner: codex
issue: skill-harness-b04
status: ready
reviewSurface: generated/review/research/human-html-artifact-self-use-investigation-2026-05-24.html
evidenceLinks:
  - .skill-harness/project.json
  - docs/developer-artifacts.md
  - scripts/generate-model-review.mjs
  - scripts/check-artifact-html-policy.mjs
  - packs/developer-artifact-skills/skills/html-review-artifact/SKILL.md
  - packs/developer-artifact-skills/skills/visual-source-artifact/SKILL.md
freshness:
  generatedAt: 2026-05-24
  sourceFirst: true
---

# Human HTML Artifact Self-Use Investigation

## Purpose

Investigate why a maintainer-facing discovery about infographic-style human artifacts did not itself produce a human HTML review artifact, and define the self-use behavior skill-harness should expect.

## Short Answer

The repo is partially self-scaffolded for developer artifacts. It has source-first configuration, a manifest, generated model review HTML, HTML safety checks, and committed generated review files. The gap is that the current generator is model-specific: `scripts/generate-model-review.mjs` only renders model artifacts. There is no generic research/product/business/data/UX human artifact renderer that turns ordinary discovery work into an infographic-style HTML surface.

The assistant also stopped at Beads discovery because the task was interpreted as planning and issue creation. That was behaviorally incomplete for this repository's own doctrine: when the user asks for human artifacts, the agent should create or update a source-backed HTML artifact and surface it in the handoff.

## Evidence Snapshot

| Signal | Current State | Implication |
| --- | --- | --- |
| Source-first doctrine | `docs/developer-artifacts.md` says HTML is never the only durable source. | Canonical source must be created before generated HTML. |
| Review surface config | `.skill-harness/project.json` points review output to `generated/review` and commits generated review files. | This repo expects generated human surfaces to be durable enough to review. |
| Model generator | `scripts/generate-model-review.mjs` filters for model artifacts. | Model views get HTML; research/discovery artifacts do not. |
| HTML policy | `allowInlineJavaScript` is false and CSP uses `script-src 'none'`. | Default interactivity must be CSS/HTML-only. |
| Skills guidance | `html-review-artifact` and `visual-source-artifact` call for rich visual review surfaces. | Agent behavior should go beyond Markdown for human-facing discovery. |

## Why The Previous Response Did Not Show HTML

1. The work was treated as issue-tracked discovery, not as a source-backed review artifact.
2. The repo has no command like `npm run artifacts:research:generate` or generic `artifacts:generate` that would make the expected output obvious.
3. The existing generated review pipeline covers `model-view` and `model-diff`, not `research-synthesis`, `planning-artifact`, or other visual-source-first families.
4. The final response summarized Beads and checks, but did not include a generated review path or open the HTML review surface.

## What Self-Use Should Mean

For skill-harness itself, a human-facing discovery should produce:

1. A canonical source file under `docs/artifacts/source/<family>/`.
2. A generated review surface under `generated/review/<family>/`.
3. A manifest entry linking source, review surface, owner, evidence, status, freshness, and source hash.
4. A successful `npm run artifacts:check`.
5. A final response that links the HTML artifact and, in Codex app, opens or offers the best preview surface.

## Infographic Review Standard

Default pages should be self-contained static HTML with:

- decision summary and review question
- evidence coverage metrics
- freshness and policy status
- inline SVG diagrams, flow charts, timelines, or bar charts
- comparison panels for current behavior, desired behavior, and gaps
- CSS-only tabs, accordions, filters, or details panels when interaction helps review
- links back to canonical sources and Beads issue IDs

JavaScript should remain a separate policy lane. The current no-script policy is compatible with useful interactivity through radio tabs, details/summary, anchors, and SVG/CSS states.

## Recommended Implementation Path

1. Define the interactive HTML policy lane first.
2. Add a generic infographic renderer for non-model artifacts.
3. Wire package scripts such as `artifacts:generate` and `artifacts:review`.
4. Extend scaffold templates so downstream repos get the same behavior.
5. Add tests proving non-model artifacts generate HTML and pass policy checks.
6. Pilot the renderer with this investigation and existing model artifacts.

## Open Risks

- If generated HTML is too manual, agents will skip it under time pressure.
- If inline JavaScript is allowed without a manifest-controlled lane, the safety checker becomes ambiguous.
- If generated review files are committed without drift checks, they become stale review surfaces.
- If the final response does not link or open the artifact, humans still experience a Markdown-only workflow.
