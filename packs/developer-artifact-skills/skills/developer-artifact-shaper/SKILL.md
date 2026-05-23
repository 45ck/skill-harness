---
name: developer-artifact-shaper
description: Choose the right developer artifact type, source of truth, and review surface for a task.
---

# Developer Artifact Shaper

Use this skill when a task needs a plan, design note, investigation, review report, or handoff artifact.

## Core Rule

Pick the artifact format from the job it must do:

- Markdown or TOON for canonical source, git diffs, CLI/TUI handoff, and long-lived docs.
- HTML for generated human review surfaces, dense comparisons, visual diagrams, prototypes, dashboards, or desktop app previews.
- Dual when a durable source and a rich review surface are both useful.
- Model review artifacts use durable text sources and generated static review views; do not make rendered diagrams the only source.
- Product, business, data, research, UX, and mockup artifacts use visual-source-first pairs when humans need visual inspection: agent-readable source plus generated visual review surface.
- High-fidelity is the default for UI, customer-facing workflow, product, and mockup review. Low-fidelity sketches are scratch only unless explicitly captured as evidence.

Do not make generated HTML the only durable source for a decision.

## Process

1. Identify the artifact purpose: decision, plan, investigation, review, report, runbook, product brief, business case, data definition, research synthesis, UX flow, handoff, or prototype.
2. Identify the canonical source location.
3. Decide the review surface: markdown, html, or dual.
4. Name the evidence that must be linked.
5. For visual artifacts, choose the family: product, business, data, research, or UX.
6. Decide whether the artifact should be listed in `docs/artifacts/artifacts.manifest.json`.
7. State the update rule: edit source first, regenerate review surface second.

## Output

### Artifact Type
### Canonical Source
### Review Surface
### Artifact Family
### Evidence Links
### Manifest Entry
### Update Rule

