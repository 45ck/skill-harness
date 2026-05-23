---
name: artifact-evidence-gate
description: Verify that developer artifacts are grounded, current, and safe to hand off.
---

# Artifact Evidence Gate

Use this skill before treating a developer artifact as ready for handoff, review, or implementation.

## Checks

- The artifact has a clear purpose and owner.
- The canonical source is named.
- Generated views link back to source.
- Source-backed generated views are listed in `docs/artifacts/artifacts.manifest.json` when the project uses the scaffold.
- Evidence is concrete: files, issues, tests, screenshots, logs, traces, or citations.
- Claims are separated from assumptions.
- Stale generated outputs are identified.
- HTML artifacts contain no secrets and no external script or asset dependency unless explicitly approved.

## Verdicts

- `ready` - evidence and source links are sufficient.
- `needs-source` - generated view exists but canonical source is missing or stale.
- `needs-evidence` - claims are under-supported.
- `unsafe` - artifact leaks sensitive data or executes unapproved code/network behavior.

## Output

### Verdict
### Blocking Gaps
### Evidence Present
### Required Fixes

