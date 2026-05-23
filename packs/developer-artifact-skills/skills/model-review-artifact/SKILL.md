---
name: model-review-artifact
description: Shape source-backed model and diagram review artifacts for architecture, UML-style, C4, dependency, and workflow views.
---

# Model Review Artifact

Use this skill when a human needs to inspect a system model, architecture view, dependency graph, workflow, or UML-style diagram.

## Core Rule

Keep the model source durable and diffable. Use generated HTML only as the review surface.

- Canonical source: Markdown, TOON, specgraph, Mermaid text, PlantUML text, or project docs.
- Review surface: static HTML with pre-rendered inline SVG or static markup.
- Do not load Mermaid, PlantUML, or other diagram runtimes in generated review HTML by default.

## Good Fits

- C4 context, container, component, dynamic, or deployment views.
- UML-style sequence, state, class/domain, activity, and use-case views.
- Architecture-space summaries: ownership, risk, quality attributes, dependencies, and decisions.
- Blast-radius views for proposed agentic changes.
- Generated dependency graphs derived from code, package metadata, specs, or traces.

## Notation Guidance

- Mermaid sequence, state, and class diagrams are good source-backed diagram candidates.
- Mermaid C4 is useful for review, but mark its C4 level explicitly and do not treat the diagram as the only architecture authority.
- UWE UML is a modelling method, not a Mermaid-native contract. Keep UWE concepts in structured source and generate simplified review diagrams when useful.
- Dependency graphs should usually be generated from evidence rather than hand-maintained.

## Required Metadata

- artifact type: `model-view`, `model-diff`, `architecture-view`, `blast-radius`, or `review-dashboard`
- model id: stable identity across revisions when the project uses model policy checks
- model kind: sequence, state, class, domain, context, container, component, dynamic, deployment, dependency, use-case, activity, or architecture-space
- notation
- method when relevant: uml, uwe, or c4
- facets when relevant: UWE content, navigation, presentation, process, access, or adaptation
- abstraction level: domain, design, runtime, deployment, or decision
- canonical source path
- generated review path, if any
- evidence links
- owner
- freshness data: source hash, renderer, renderer version, and generated timestamp when available
- for model diffs: before artifact id, after artifact id, diff method, lineage, and generated review path

## Process

1. Identify the question the model must answer.
2. Pick the model kind and abstraction level.
3. Choose the canonical notation and source path.
4. Name the evidence that backs the model.
5. Decide whether a generated HTML review surface is useful. When `--enable-modeling` is active, pair authored model sources with generated HTML review surfaces under `generated/review/models/`.
6. Record the artifact in `docs/artifacts/artifacts.manifest.json`.
7. If HTML is generated, pre-render diagrams and run the artifact checks. When model policy checks are scaffolded, also run `node scripts/check-model-artifact-policy.mjs`.

## Output

### Model Purpose
### Canonical Source
### Model Kind
### Notation
### Review Surface
### Evidence Links
### Freshness Rule
### Required Checks
