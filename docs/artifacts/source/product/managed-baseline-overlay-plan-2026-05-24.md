---
artifactType: planning-artifact
artifactId: managed-baseline-overlay-plan-2026-05-24
owner: codex
issue: skill-harness-dcj
status: draft
reviewRequired: true
evidenceLinks:
  - README.md
  - docs/artifacts/source/agent-native-bootstrap-update-plan-2026-05-24.md
  - docs/artifacts/source/skill-harness-opportunity-review-2026-05-23.md
  - scripts/bootstrap_dependencies.py
  - scripts/render_codex_agents.py
  - scripts/render_claude_agents.py
  - cmd/skill-harness/main.go
freshness:
  generatedAt: 2026-05-24
  sourceFirst: true
---

# Managed Baseline Overlay Plan

Date: 2026-05-24

Beads issue: `skill-harness-dcj`

Status: draft planning artifact

## Purpose

This plan defines the next product direction for `skill-harness`: a managed baseline that downstream repositories can customize without losing the ability to receive upstream updates.

The current harness is useful as an installer and generator, but active repositories have already grown their own agent rules, local skills, hooks, package workflows, and cost controls. The future model should treat that local ownership as first-class instead of accidental drift.

## Research Signal

A three-agent read-only review compared `skill-harness` with active consumer archetypes:

| Archetype | Signal | Design implication |
|---|---|---|
| Governed control-plane repo | Has local Beads/worktree workflows, stricter tool permissions, hooks, and repo-owned Claude skills. | Executable workflow assets need explicit ownership and should not be overwritten by baseline updates. |
| Large product monorepo | Has mirrored local skills for multiple tools, custom MCP config, project-specific agents, and workstation/cost constraints. | Baseline should support generated mirrors, local overlays, and opt-outs for expensive or irrelevant surfaces. |
| Lightweight internal monorepo | Has lean root guidance and app-level instructions, but does not need the full governance stack by default. | Minimal profile must remain genuinely small and should not install heavy artifacts, hooks, or Beads unless requested. |
| Agent workspace runtime | Uses identity, memory, heartbeat, plugins, local skills, and runtime config rather than app CI workflows. | Harness profiles must support agent workspaces without assuming normal package scripts or repo closeout rules. |

Private project details are intentionally summarized as archetypes. The durable planning value is the pattern: downstream repositories need a baseline plus local overlay model, not blind file replacement.

## Problem Statement

`skill-harness` currently installs and renders shared assets, but it does not yet have a durable contract for repo-local changes to those assets.

That creates three failure modes:

1. **Update fear:** maintainers avoid rerunning setup because it may overwrite local instructions, hooks, settings, or skills.
2. **Silent divergence:** copied files keep drifting until nobody knows whether they are upstream baseline, local customization, or stale generated output.
3. **Unbounded cost:** repos inherit more agents, skills, mirrors, review surfaces, and workflows than they actually use.

## Decision Direction

Adopt a managed baseline overlay model.

`skill-harness` should own upstream baselines, profiles, and renderers. Each target repository should own a small manifest that declares which baseline surfaces it consumes, disables, overrides, or treats as generated output.

This is not a final implementation decision. It is the preferred planning direction for follow-up design, schema, and resolver work.

## Ownership Modes

Every managed surface should have an explicit mode:

| Mode | Meaning | Update behavior |
|---|---|---|
| `generated` | Output derived from baseline plus manifest. | Safe to replace after dry-run because source of truth is elsewhere. |
| `managed-section` | Only named sections are controlled by harness markers. | Sync managed sections, preserve all other local content. |
| `overlay` | Repo adds or removes skills, agents, packs, or sections on top of the baseline. | Resolve overlay after upstream baseline; report conflicts. |
| `owned` | Repo owns the file or directory. | Never overwrite automatically. |
| `ignored` | Repo opts out of this surface. | Do not install, render, or warn except in audit output. |

Safe default: any file or directory without harness provenance is `owned`.

## Candidate Manifest

The exact schema should be designed separately, but the durable concept is:

```json
{
  "version": 1,
  "baseline": {
    "source": "https://github.com/45ck/skill-harness.git",
    "channel": "stable",
    "pin": null
  },
  "profile": "minimal",
  "surfaces": {
    "AGENTS.md": { "mode": "managed-section" },
    "CLAUDE.md": { "mode": "managed-section" },
    ".claude/hooks": { "mode": "owned" },
    ".codex/skills": { "mode": "generated" },
    ".github/skills": { "mode": "ignored" },
    "scripts/beads/bd.mjs": { "mode": "owned" }
  },
  "agents": {
    "enabled": ["requirements-analyst", "software-architect"],
    "disabled": ["pentest-reviewer"]
  },
  "packs": {
    "enabled": ["developer-artifact-skills"],
    "disabled": ["demo-production-skills"]
  },
  "policies": {
    "beads": "optional",
    "closeout": "repo-defined",
    "globalWrites": "ask",
    "packageInstalls": "ask",
    "hookChanges": "ask"
  }
}
```

Potential file names:

- `.skill-harness/baseline.manifest.json`
- `.skill-harness/baseline.lock.json`
- `.skill-harness/update-report.json`

The manifest describes desired state. The lockfile records resolved upstream revisions, rendered outputs, hashes, and effective skill/agent lists.

## Update Flow

1. Read the repo manifest.
2. Read the previous lockfile.
3. Resolve the selected upstream baseline and profile.
4. Apply repo-local overlays, disables, and ownership rules.
5. Compare old resolved state, new resolved state, and current working tree.
6. Classify changes.
7. Produce a dry-run report.
8. Apply only generated outputs and approved managed-section updates.
9. Leave owned and ignored surfaces untouched.
10. Update the lockfile and setup proof.

## Change Classes

| Class | Meaning | Default action |
|---|---|---|
| `baseline-addition` | Upstream added an agent, skill, pack, section, or profile. | Suggest; enable only if profile allows additions. |
| `baseline-fix` | Upstream changed an enabled item without policy-sensitive behavior. | Auto-apply to generated surfaces after dry-run. |
| `baseline-removal` | Upstream removed or renamed an item. | Require review when repo still references it. |
| `overlay-addition` | Repo adds a local skill, pack, or section. | Preserve and record in lockfile. |
| `shadow` | Repo-local item intentionally wins over upstream id. | Preserve, but report clearly. |
| `orphaned-override` | Repo override references missing upstream item. | Block sync until resolved or marked local-only. |
| `policy-sensitive` | Change affects permissions, hooks, CI, security, quality, package installs, global writes, or pushing. | Ask before apply. |
| `detached-generated-file` | Generated file was edited directly. | Preserve in report; fix by moving intent into manifest/overlay. |

## Cost And Scope Controls

The overlay model should reduce cost in three ways:

1. **Context cost:** repos can disable irrelevant agents, skills, packs, and mirrored tool surfaces.
2. **Maintenance cost:** generated mirrors are derived from one canonical source instead of edited in parallel.
3. **Operational cost:** heavy hooks, Beads workflows, model artifacts, media artifacts, and review surfaces are profile-driven instead of defaulting everywhere.

Suggested commands:

```bash
skill-harness repo audit --dir .
skill-harness repo update --dir . --check
skill-harness repo trim --dir . --dry-run
skill-harness repo sync --dir .
skill-harness repo drift --dir .
```

Initial implementation should be read-only where possible. `audit`, `drift`, and `update --check` should exist before mutating `sync`.

## Profiles

Profiles should be defaults, not prisons:

| Profile | Intended use |
|---|---|
| `minimal` | Lightweight repos with root guidance and little or no local skill surface. |
| `product` | Product monorepos with selected agents, repo-local skills, checks, and docs. |
| `governed` | Repos with Beads, worktrees, closeout policy, hooks, and tool permission rules. |
| `desktop-review` | Repos that benefit from generated human HTML review surfaces. |
| `agent-workspace` | Agent runtime workspaces with identity, memory, plugins, heartbeat, and local skills. |

Repos can still disable individual surfaces inside any profile.

## Acceptance Criteria

- A repo can receive upstream changes to managed sections of `AGENTS.md` without losing local sections.
- A repo can mark `.claude/hooks/*` or `scripts/beads/*` as owned and future updates will not overwrite them.
- A repo can keep one canonical local skill source and generate only the enabled mirrors.
- A repo can disable `.github/skills` or other duplicated surfaces to reduce maintenance and context cost.
- A lightweight repo can adopt `skill-harness` without receiving Beads, hooks, model artifacts, media artifacts, or local skill mirrors.
- An agent workspace can adopt baseline guidance without assuming package scripts, app CI, or Beads closeout.
- A dry-run update reports updated, skipped, owned, ignored, conflicted, and policy-sensitive surfaces.
- A no-op update on an unchanged repo produces no file churn.

## Open Questions

- Should ownership be tracked at file, directory, named-section, or individual skill level by default?
- Should `AGENTS.md` and `CLAUDE.md` be partly generated, fully local with includes, or only linted against manifest policy?
- How should canonical local skills generate Claude, Codex, and Copilot-compatible mirrors without losing tool-specific metadata?
- Should `setup-project` evolve into this model, or should `repo init/sync/update` be a separate command family?
- What semantic versioning rules should apply to baseline profiles and skill pack contracts?

## Model Impact

This artifact is planning only. It does not change runtime behavior, CLI commands, installer behavior, setup proof schema, or generated project output.

If implemented later, the following canonical models should be updated:

- `sh-system-context`
- `sh-usecase-cli-workflows`
- `sh-component-scaffold-engine`
- `sh-domain-artifact-governance`
- `sh-state-artifact-freshness`

## Recommended Next Slice

Implement an observe-only audit before any mutating sync behavior:

1. Define the manifest and lockfile schema.
2. Add `skill-harness repo audit --dir .`.
3. Detect existing repo-local agent surfaces and classify them as owned by default.
4. Report likely profile, enabled/disabled surfaces, and collision risks.
5. Pilot on one governed repo, one product monorepo, one lightweight monorepo, and one agent workspace.
