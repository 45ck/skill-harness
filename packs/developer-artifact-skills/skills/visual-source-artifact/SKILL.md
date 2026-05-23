---
name: visual-source-artifact
description: Shape product, business, data, research, UX, and mockup artifacts as agent-readable sources with generated visual human review surfaces.
---

# Visual Source Artifact

Use this skill when a human needs to inspect product, business, data, research, UX, or mockup work visually while agents need durable source they can diff, update, and verify.

## Core Rule

Use a source-first pair:

- Canonical source: Markdown, TOON, JSON, YAML, specgraph-compatible docs, model text, or structured data.
- Generated review surface: static HTML, dashboard, state board, prototype, journey map, schema map, evidence board, or comparison page under `generated/review/`.
- For human-facing discovery and planning, the generated review surface should be infographic-style HTML under `generated/review/<family>/`.

Generated visual files are never the only durable source. Edit source first, then regenerate or discard the review surface.

## Default Fidelity

- High-fidelity is the default for UI, product, customer-facing workflow, and mockup review.
- Low-fidelity sketches are scratch artifacts only unless the project explicitly records them as research evidence.
- Use realistic states, dense data, error paths, accessibility cues, assumptions, and evidence strength in the review surface.
- Use scan-friendly metrics, charts, timelines, evidence/freshness panels, and source links for non-UI human review artifacts.
- Use Mermaid for authored diagrams, Vega-Lite as the default chart grammar, Observable Plot for compact exploratory charts, D3 for bespoke static layouts, Graphviz for node-edge maps, ECharts only as generation-time/static output, RAWGraphs for exported design-led SVGs, and Chart.js only through server-rendered/static output or an equivalent static chart.
- Put chart and graph intent in `artifact-infographic` JSON fences or manifest `infographics` arrays when the project has the generic renderer.

## Artifact Families

- Product: PRD, opportunity brief, feature map, roadmap, acceptance matrix.
- Business: business model, pricing assumptions, stakeholder map, operating risk, go-to-market assumptions.
- Data: schema, data dictionary, metric definition, lineage, quality rule, sample records.
- Research: claim-evidence matrix, literature theme map, interview synthesis, assumption register.
- UX: design brief, component state spec, interaction flow, high-fidelity prototype source, journey map.

## Agent Team

- `requirements-analyst` owns product intent, requirements, and acceptance criteria.
- `delivery-manager` owns business viability, stakeholder impact, rollout, and risk.
- `backend-engineer` owns data shape, schemas, integrity, and metrics implementation constraints.
- `research-writer` owns citations, claim strength, research synthesis, and gaps.
- `ux-researcher` owns task evidence, prototype critique, high-fidelity UX review, and accessibility concerns.
- `system-modeler` owns UML/UWE/C4/workflow impact when structure or behavior changes.
- `quality-reviewer` owns freshness, manifest completeness, evidence sufficiency, and readiness risks.

## Process

1. Name the decision the visual artifact supports.
2. Select the artifact family and owner agent.
3. Define the canonical source path under `docs/artifacts/source/<family>/` unless a domain docs path is better.
4. Define the generated review path under `generated/review/<family>/`.
5. List evidence links and label real evidence separately from synthetic user or agent-simulation evidence.
6. Add or update the manifest entry with source, review surface, owner, evidence, status, freshness, and `reviewRequired: true` when a generated HTML surface is expected.
7. Add `artifact-infographic` specs for charts, graphs, timelines, or source-backed infographic panels when useful.
8. Run `node scripts/generate-artifact-review.mjs` when available.
9. Run manifest and HTML policy checks before handoff when the project has them.

## Output

### Artifact Family
### Decision Supported
### Canonical Source
### Generated Review Surface
### Owner Agent
### Evidence Links
### Fidelity Requirement
### Manifest Fields
### Readiness Gates

