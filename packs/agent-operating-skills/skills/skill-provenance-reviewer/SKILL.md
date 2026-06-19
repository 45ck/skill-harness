---
name: "skill-provenance-reviewer"
description: "Review skills, agents, rules, plugins, and MCP/tool packs for license, source, script, permission, and supply-chain risk before adoption."
---

# Skill Provenance Reviewer

Use this skill before a public or generated skill becomes a first-party harness asset, fixture, or downstream install surface.

## Process

- Identify the source repo, owner, commit or release, license, and whether the license covers the specific copied content.
- Separate protected expression from reusable ideas, formats, and behavior.
- Inspect helper scripts, package manifests, MCP configs, plugin manifests, and install instructions for side effects.
- Flag user-global writes, credential handling, shell pipelines, postinstall hooks, network calls, destructive commands, approval bypass, sandbox disabling, and secret exposure patterns.
- Decide whether the material can be copied with attribution, must be rewritten, should become a fixture only, or should be rejected.
- Record attribution and provenance requirements for anything copied or adapted.

## Output

### Source
### License Position
### Copied Or Rewritten Content
### Tool And Script Surfaces
### Permission Risks
### Required Attribution
### Adoption Decision
### Evidence

## Avoid

- relying on repo popularity as a trust signal
- assuming a repo-wide license covers every embedded asset
- burying MCP or plugin permissions inside implementation detail
- approving generated or recorded skills without provenance and redaction review
