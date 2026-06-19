---
name: "task-memory-profile-planner"
description: "Choose the right task-memory and issue-state pattern for a repo: Beads, lightweight files, external tracker integration, or no durable task memory yet."
---

# Task Memory Profile Planner

Use this skill when deciding how a repo should preserve agent task state across sessions, branches, worktrees, and tools.

## Process

- Identify who owns task state: human maintainer, agent team, external tracker, or local repo.
- Decide whether the repo needs durable issue graphs, dependency edges, worktree coordination, persistent memories, or only lightweight plan files.
- Compare options:
  - Beads for repo-local graph issues and agent session continuity
  - lightweight checked-in plan/evidence files
  - external tracker such as GitHub Issues, Linear, Jira, or Azure DevOps
  - no durable task memory yet
- Define sync, conflict, branch, and closeout rules.
- Record what must never be stored: secrets, raw private tickets, broad mailbox data, private traces, or unredacted customer content.
- Tie the selected memory profile to installer flags and setup proof.

## Output

### Task-State Owner
### Required Memory Shape
### Selected Profile
### Installer Flags
### Sync And Closeout Rules
### Privacy Exclusions
### Evidence Path

## Avoid

- enabling Beads only because it is available
- storing private operational context in generated traces
- mixing task memory with long-term model memory without an explicit policy
- leaving closeout rules separate from the repo's normal quality gates
