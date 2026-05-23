---
name: demo-review-surface
description: Create safe static review surfaces for demo and QA media. Use when Codex needs an HTML or Markdown review page comparing demo videos, source specs, QA reports, screenshots, storyboards, quality findings, and evidence links from demo-machine, manual-qa-machine, Playwright, or video analysis outputs.
---

# Demo Review Surface

## Overview

Use this skill when a human needs to inspect demo media and its evidence together. Prefer generated HTML for desktop/app-browser review, and Markdown/TOON for CLI-only handoff.

## Required Sections

- summary: what the demo proves and what it does not prove
- source: `.demo.yaml`, QA flow, spec, issue, or source artifact
- media: rendered video, raw capture, poster, frame strip, or slideshow output
- evidence: events, quality, verification, QA reports, screenshots, traces, analyzer bundles
- findings: failed checks, warnings, stale artifacts, missing evidence
- next action: approve, regenerate, cut down, fix source, or discard

## HTML Constraints

- Use one self-contained static `.html` file.
- Use inline CSS only.
- Do not load external scripts, fonts, images, media, analytics, or network resources.
- Do not include `<script>`, iframe, object, embed, form, meta refresh, external `src`, external `href`, or browser storage/network APIs.
- Embed small images only when safe; prefer relative links to large local media.
- Link or name the canonical source artifact and issue.

## Provenance

If the target repo has `docs/artifacts/artifacts.manifest.json`, add or update an entry with:

- `id`
- `type`: usually `review-dashboard` or `evidence-pack`
- `source`
- `sourceHash` when available
- `reviewSurface`
- `evidenceLinks`
- `status`

Then run:

```bash
node scripts/check-artifact-manifest.mjs
node scripts/check-artifact-html-policy.mjs
```

## Review Defaults

- Desktop agent apps: generate HTML under `generated/review/`.
- CLI/TUI: produce Markdown summary first; generate HTML only when useful.
- CI: produce JSON/Markdown evidence indexes and avoid large media unless promoted.

## Output

Name the generated path, source path, evidence paths, checks run, and any missing evidence.
