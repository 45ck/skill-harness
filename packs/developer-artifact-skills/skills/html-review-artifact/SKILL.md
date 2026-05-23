---
name: html-review-artifact
description: Produce safe, self-contained HTML review artifacts from canonical project sources.
---

# HTML Review Artifact

Use this skill when a human needs to scan, compare, inspect, or interact with a rich generated view.

## Default Constraints

- Generate one self-contained `.html` file.
- Use inline CSS only.
- Use inline JavaScript only when interaction materially improves review.
- Do not load external scripts, fonts, images, analytics, or network resources.
- For Mermaid, C4, UML-style, and architecture diagrams, embed pre-rendered inline SVG or static markup by default.
- For product, business, data, research, UX, and mockup review, render the source into a visual surface such as a dashboard, evidence board, state board, journey map, schema map, or high-fidelity prototype.
- For UI and customer-facing workflow review, prefer high-fidelity states with realistic copy, data density, errors, loading states, and accessibility affordances.
- Label synthetic user, simulated customer, or agent-generated evidence separately from real user or customer evidence.
- Do not include secrets, tokens, credentials, private logs, or customer data.
- Link to the canonical source artifact and issue.
- Use semantic headings, landmarks, meaningful link text, and alt text for embedded images.

## Good Fits

- option comparisons
- PR walkthroughs
- architecture explainers
- incident and investigation reports
- visual evidence reviews
- prototype controls for design tuning
- product briefs and feature maps
- business assumption dashboards
- data dictionaries, schema maps, and metric definitions
- research claim-evidence boards
- high-fidelity UX mockups and interaction state boards

## Poor Fits

- canonical decisions
- source-controlled specs intended for diffs
- short handoff notes
- CLI-only workflows
- artifacts that would need live credentials or authenticated pages

## Output Checklist

- source path and generated path are named
- artifact family and owner agent are named
- visual hierarchy supports scanning
- layout works at mobile and desktop widths
- source, evidence strength, assumptions, and freshness are visible
- all external links are explicit
- no external runtime dependency exists
