---
name: demo-story-packager
description: Package completed demo-machine runs, browser capture artifacts, or evidence-backed product demos into source-linked handoff bundles. Use when Codex needs to organize `.demo.yaml`, `events.json`, `quality.json`, screenshots, storyboard outputs, rendered videos, review prompts, or release-ready demo assets without losing provenance.
---

# Demo Story Packager

## Overview

Use this skill to turn a captured demo into a reviewable and reusable package. Keep the durable source first and treat video, posters, frame strips, and HTML as generated outputs.

## Inputs

- `.demo.yaml` or equivalent source flow
- `events.json`, `verification.json`, `quality.json`, `environment.json`
- rendered `output.mp4` and raw `video.webm`
- analyzer outputs such as `review-bundle.json`, `review-prompt.md`, `video.shots.json`, `segment.evidence.json`, `layout-safety.report.json`, and `segment-storyboard/`
- screenshots, Playwright traces, or manual QA reports when they are part of the evidence

## Workflow

1. Identify the canonical source: usually `.demo.yaml`, QA flow JSON, or a source artifact under `docs/artifacts/source/`.
2. Inventory generated media and evidence. Do not infer pass/fail claims from video alone.
3. Separate draft assets from approved assets. Keep generated assets out of git unless the repo intentionally promotes them.
4. Write a handoff summary with audience, source path, run directory, evidence files, known quality findings, and next action.
5. If a review surface is needed, use `demo-review-surface` and record it in `docs/artifacts/artifacts.manifest.json` when the target repo has a manifest.
6. Before handoff, run available checks such as `demo-machine analyze`, `demo-machine doctor`, `node scripts/check-artifact-manifest.mjs`, or `node scripts/check-artifact-html-policy.mjs`.

## Output Contract

Name these paths when available:

- source spec or QA flow
- run directory
- rendered media
- evidence files
- generated review surface
- approval or promotion destination
- commands already run

## Guardrails

- Do not treat generated video as canonical evidence without source files.
- Do not include secrets, tokens, private browser data, or customer records in shareable media.
- Prefer local-first artifacts. Do not upload media to third-party services unless the user explicitly asks.
- If the source and media disagree, mark the package stale and regenerate from source.
