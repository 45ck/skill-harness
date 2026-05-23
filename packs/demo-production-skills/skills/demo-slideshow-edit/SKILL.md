---
name: demo-slideshow-edit
description: Plan no-caption slideshow-style MP4s and frame reels from demo-machine screenshots, storyboard frames, selected video spans, or manual QA evidence. Use when Codex needs a polished still-frame walkthrough, hero reel, or reviewable visual summary without recapturing the browser flow.
---

# Demo Slideshow Edit

## Overview

Use this skill when the best output is a curated still-frame or segment-based reel rather than a full narrated playback. The slideshow must remain source-backed and evidence-linked.

## Inputs

- storyboard frames or screenshots
- selected spans from `video.shots.json`, `segment.evidence.json`, or `events.json`
- source `.demo.yaml` or QA flow/report
- target duration, aspect ratio, and audience

## Workflow

1. Select frames or spans that show state changes clearly.
2. Reject frames that expose secrets, private data, or unreadable UI.
3. Define the sequence with durations, crop/zoom notes, and transition notes.
4. Prefer simple transitions and stable framing over decorative effects.
5. Record the source path, evidence links, and output path in the artifact manifest when available.
6. Produce a render plan or command. Do not claim the MP4 is approved until evidence and safety checks pass.

## Output

Return:

- source run directory
- ordered frame/span list
- duration and aspect ratio
- output media path
- review surface path, if generated
- checks needed before handoff

## Guardrails

- Do not fabricate app states that were not captured.
- Do not crop away failures, warnings, or key context.
- Do not use external stock or music unless the user explicitly asks and licensing is recorded.
- Keep generated slideshow media out of git unless explicitly promoted.
