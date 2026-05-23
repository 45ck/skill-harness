---
name: demo-social-cut
description: Plan short silent demo cuts, slideshow edits, hero loops, and social clips from existing demo-machine or QA evidence. Use when Codex needs no-caption polished videos, frame-based reels, poster frames, or audience-specific excerpts derived from `events.json`, screenshots, storyboard evidence, or rendered demo runs.
---

# Demo Social Cut

## Overview

Use this skill to derive a polished short edit from a real captured flow. The edit plan must point back to the source run and evidence; it is not a free-form video editing brief.

## Decision Rules

- Use `demo-machine edit` or a project-native render path when the run has `events.json`.
- Use frame and screenshot based slideshow cuts when the strongest evidence is screenshots or storyboard output.
- Use FFmpeg/Remotion only as render mechanisms, not as the source of truth.
- Prefer no captions when the user asks for silent, hero, slideshow, or in-browser preview output.
- Keep narration and captions optional derivatives, never implicit defaults for a silent cut.

## Edit Plan

Produce a concise plan with:

- source run directory and source spec
- target audience: review, docs, launch, social, portfolio, issue repro, or release note
- aspect ratio and duration target
- selected moments with evidence references
- output assets: MP4, GIF/WebP preview, poster frame, frame strip, or HTML review page
- render command or next command to run
- quality checks and known risks

## Recommended Cut Types

- `hero-loop`: 6-15 second silent loop for README, landing page, or app preview
- `feature-cut`: 20-45 second focused clip of one capability
- `repro-cut`: short evidence clip showing a bug or fix path
- `slideshow-reel`: still frames, zooms, and transitions from screenshots/storyboards
- `release-recap`: selected before/after or workflow moments for changelog review

## Quality Checks

- Verify the source run is complete before editing.
- Check that UI text remains readable at the target aspect ratio.
- Avoid cuts that hide failed assertions, console/network failures, or layout safety warnings.
- For silent cuts, make state changes visually obvious through framing, cursor, timing, or title cards.
- If external assets are used, record license and source; otherwise keep the cut local-first.

## Handoff

Return the edit plan and any generated paths. If rendering is not run, state the exact command and prerequisites.
