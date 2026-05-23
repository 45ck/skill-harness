---
name: qa-to-demo
description: Convert manual QA flows, QA reports, screenshots, network/console evidence, accessibility findings, or bug reproduction steps into demo-machine specs, reproducible demo plans, or short evidence-backed repro clips. Use when Codex needs to turn manual-qa-machine output into a product demo, release proof, or issue reproduction asset.
---

# QA To Demo

## Overview

Use this skill to bridge quality evidence and demo production. The goal is a reproducible spec or clip plan grounded in QA artifacts, not a marketing retelling that hides failures.

## Inputs

- `qa-report.json` and `qa-report.md`
- manual-qa-machine flow JSON
- screenshots, snapshots, console logs, network logs, page errors, accessibility reports, and performance reports
- bug reproduction notes or Beads/GitHub issue context
- target app URL and startup command, if a new demo spec must be drafted

## Conversion Workflow

1. Read the QA verdict and evidence before selecting moments.
2. Choose the target output:
   - product demo for a passing flow
   - repro clip for a failing flow
   - release proof for a certified flow
   - exploratory finding package for triage
3. Map QA steps to demo-machine actions using stable targets first: role, label, text, test id, then CSS only as a fallback.
4. Preserve assertions as demo checkpoints or review notes.
5. Draft a `.demo.yaml` only when the source has enough target and timing detail; otherwise produce a demo plan with missing inputs.
6. If rendering or capture is run, package output with `demo-story-packager`.

## Evidence Rules

- Do not convert `fail` or `inconclusive` QA results into positive product demos.
- Keep console, network, accessibility, and performance warnings visible in the handoff.
- For bug repro media, show the shortest path that proves the issue.
- For release proof media, include the certification attempt or report path.

## Output

Return one of:

- `.demo.yaml` draft plus run command
- repro clip plan plus source evidence
- release proof package plan
- list of missing details blocking a reliable demo

Always include the QA report path, source flow path, selected steps, and evidence files.
