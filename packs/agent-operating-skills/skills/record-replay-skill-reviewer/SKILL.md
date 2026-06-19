---
name: "record-replay-skill-reviewer"
description: "Review skills generated from recorded workflows before they are installed, shared, or treated as reusable automation."
---

# Record/Replay Skill Reviewer

Use this skill when a skill is generated from a browser, desktop, terminal, or app recording.

## Process

- Identify what was recorded, who recorded it, when, and which host generated the skill.
- Separate durable process from incidental UI coordinates, timing, session state, local paths, and private data.
- Check for credentials, personal data, cookies, account identifiers, tokens, screenshots, private URLs, and unreleased business data.
- Replace brittle recorded steps with stable tool/API calls when available.
- Add explicit preconditions, permissions, verification steps, and failure handling.
- Mark the skill as draft until a human reviews source, redaction, and replay evidence.
- Decide whether the generated skill belongs as a private repo-local skill, a fixture, or a rewritten first-party skill.

## Output

### Recording Source
### Sensitive Data Review
### Durable Workflow
### Brittle Steps
### Required Redactions
### Verification Evidence
### Install Decision

## Avoid

- installing generated skills globally by default
- preserving raw recordings, screenshots, HAR files, cookies, or private traces in the repo
- confusing a successful replay with a safe reusable skill
- copying UI-specific steps when an authenticated API or MCP tool is the safer boundary
