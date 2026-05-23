---
name: "tool-permission-planner"
description: "Design least-privilege tool access for agent workflows, MCP servers, connectors, browser automation, shell commands, and external side effects."
---

# Tool Permission Planner

Use this skill when an agent needs tools beyond local read-only reasoning.

## Process

- Start from user task outcomes, not endpoint completeness.
- List each tool, its side effects, and the data it can expose.
- Choose read-only or dry-run modes first when they can prove the next step.
- Define command, path, network, account, and write boundaries.
- Name approval gates for destructive or externally visible actions.
- Define audit evidence: logs, issue comments, manifests, screenshots, traces, or command output.

## Permission Tiers

- `read`: inspect files, issues, docs, metadata, screenshots
- `write-local`: edit repo files or generated artifacts
- `execute-local`: run tests, builds, formatters, or local validators
- `external-read`: call APIs or remote systems without mutation
- `external-write`: create, update, publish, deploy, merge, or notify
- `destructive`: delete, revoke, reset, drop, overwrite, or irreversibly mutate

## Output

### Tool Surface
### Minimum Required Tier
### Boundaries
### Approval Gates
### Audit Evidence
### Denied Capabilities
