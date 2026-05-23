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

## Poor Fits

- canonical decisions
- source-controlled specs intended for diffs
- short handoff notes
- CLI-only workflows
- artifacts that would need live credentials or authenticated pages

## Output Checklist

- source path and generated path are named
- visual hierarchy supports scanning
- layout works at mobile and desktop widths
- all external links are explicit
- no external runtime dependency exists
