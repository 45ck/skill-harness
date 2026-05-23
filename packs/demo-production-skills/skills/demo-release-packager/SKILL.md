---
name: demo-release-packager
description: Assemble approved demo media, canonical source specs, evidence links, provenance, and promotion notes into a release or handoff bundle. Use when Codex needs to prepare demo-machine, manual QA, or generated media outputs for docs, launch pages, changelogs, social previews, or another agent.
---

# Demo Release Packager

## Overview

Use this skill only after source, media, and evidence have been reviewed. It packages approved assets for handoff while excluding sensitive raw capture data by default.

## Required Bundle Contents

- canonical source: `.demo.yaml`, QA flow, or source artifact
- approved media path: MP4, GIF/WebP preview, poster frame, or frame strip
- evidence summary with links to quality, verification, QA, storyboard, or segment files
- status: draft, needs-evidence, approved, rejected, or stale
- owner and intended destination
- commands run and checks passed

## Default Exclusions

Exclude unless explicitly redacted and approved:

- `trace.zip`
- HAR or raw network dumps
- full console logs
- page error dumps with private context
- screenshots containing secrets, tokens, personal data, or customer data
- unreviewed OCR/ASR transcripts

## Workflow

1. Confirm the source and media are not stale.
2. Confirm upstream quality or QA verdicts support the requested status.
3. Apply the exclusion list before creating a handoff bundle.
4. Write a release note or handoff summary with exact paths.
5. Record the bundle in `docs/artifacts/artifacts.manifest.json` when available.

## Guardrails

- Failed or inconclusive QA evidence can produce only a draft or repro package, never an approved product demo.
- Missing analyzer evidence should downgrade status to `needs-evidence`.
- Do not upload, publish, or promote assets without explicit user approval.
