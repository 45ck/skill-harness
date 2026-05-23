---
name: "agent-run-evidence-reviewer"
description: "Review agent run traces, eval summaries, tool logs, and handoff evidence before accepting a workflow result or self-improvement proposal."
---

# Agent Run Evidence Reviewer

Use this skill before accepting that an agent run worked, that a workflow improved, or that a learning should be made durable.

## Evidence Checks

- The run has a task or issue id.
- The agent role and tool boundaries are clear.
- Changed files, commands, tests, and validators are named.
- Failures and retries are visible, not edited out.
- Claims are tied to concrete evidence.
- Generated artifacts point back to source.
- Sensitive traces are excluded or redacted.
- Follow-up work is filed instead of hidden in prose.

## Verdicts

- `accept`: evidence supports the result
- `needs-gate`: result may be correct but validation is missing
- `needs-redaction`: evidence is useful but unsafe to store or share
- `needs-follow-up`: result is partial and requires tracked work
- `reject`: evidence contradicts the result or is too weak

## Output

### Verdict
### Evidence Present
### Missing Gates
### Redaction Needs
### Follow-Up Issues
### Durable Learning Decision
