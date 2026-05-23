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

Do not make generated HTML the only durable source for a decision.

## Process

1. Identify the artifact purpose: decision, plan, investigation, review, report, runbook, handoff, or prototype.
2. Identify the canonical source location.
3. Decide the review surface: markdown, html, or dual.
4. Name the evidence that must be linked.
5. Decide whether the artifact should be listed in `docs/artifacts/artifacts.manifest.json`.
6. State the update rule: edit source first, regenerate review surface second.

## Output

### Artifact Type
### Canonical Source
### Review Surface
### Evidence Links
### Manifest Entry
### Update Rule

