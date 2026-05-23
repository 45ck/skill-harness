# UML-First Default Review

Date: 2026-05-23

Beads issue: `skill-harness-0e9`

Status: review complete, implementation not yet safe as a naive default flip

## Purpose

This note synthesizes a six-agent review of the request to make UML/model artifacts the default posture for developer artifacts, especially for human engineering review and continuous human/agent system understanding.

## Consensus

UML-first should become the default for fresh projects that enable developer artifacts, but not by simply treating the current `--enable-modeling` flag as always-on.

The safer product shape is:

- modeling remains inside developer artifacts, not a new profile family
- source-backed models become the default comprehension surface
- generated HTML remains a human review surface, not canonical truth
- existing `none`, `--skip-artifacts`, and `--skip-developer-artifacts` flows must keep working
- existing repos must not be silently migrated into stricter model checks
- terminal-first profiles still get canonical models, but HTML review should be lazy or profile-sensitive

## Recommended Direction

Replace the current boolean with an explicit modeling mode:

```text
modeling-mode: off | baseline | uml-first
```

Recommended behavior:

- `baseline`: source-first model inventory and durable model taxonomy, without strict human HTML review requirements
- `uml-first`: strict UML/UWE/C4 model policy, model-diff support, model checker, and expected human review surfaces
- `off`: developer artifacts remain enabled, but model-specific scaffold/checks are skipped

Keep `--enable-modeling` as a backward-compatible alias for `--modeling-mode uml-first`.

Add an opt-out flag such as `--skip-modeling` or `--modeling-mode off`.

## Fresh Vs Existing Repos

Fresh setup:

- default to `uml-first` when developer artifacts are enabled
- write explicit modeling mode into `.skill-harness/project.json`
- write requested/effective modeling mode into `.skill-harness/setup-proof.json`
- scaffold model source and review directories
- scaffold model inventory and model-diff template
- wire model checks into package scripts and setup proof

Existing setup:

- do not silently tighten model checks on a repo that already has `.skill-harness/project.json` without modeling mode
- preserve legacy behavior unless the user explicitly requests migration
- add a migration path that updates config, scripts, local artifact README, checker files, and setup proof coherently

## Model Inventory

The default scaffold needs a canonical model inventory, not just an empty folder.

Minimum recommended inventory:

- `system-context`: system boundary, actors, external systems, ownership
- `domain-language`: domain terms, entities, relationships, invariants
- `use-case-index`: actors, goals, important workflows
- update matrix: which model kinds must be revisited for each change type

Durable model defaults:

- required baseline: `context`, `domain`, `use-case`
- conditional durable models: `activity`, `sequence`, `state`, `container`, `component`, `deployment`
- durable only when intentionally authored: `class`, `dynamic`
- generated evidence by default: `dependency`, reverse-engineered class/dynamic views, runtime snapshots

UWE remains metadata/facets, not a separate top-level model family:

- `content`
- `navigation`
- `presentation`
- `process`
- `access`
- `adaptation`

## Feature Update Rules

These rules should move from the research note into durable setup guidance:

- user goal or workflow changes -> update `use-case`, `activity`, and maybe `sequence`
- business concepts or rules change -> update `domain` and maybe `state`
- service/component boundaries change -> update `context`, `container`, `component`, and maybe `deployment`
- internal refactor only -> usually regenerate dependency evidence, not durable UML
- generated review is stale after source changes until checks regenerate or explicitly mark it stale

Every Beads issue should record model impact:

- `none`
- `evidence-only`
- `canonical-model-update-required`

## HTML Human Artifacts

Generated HTML is still important, but should be profile-sensitive.

Recommended defaults:

- `model-view`: can be ready without HTML in CLI/TUI workflows if source, evidence, owner, and freshness are complete
- `model-diff`: requires generated HTML when entering human review or PR review
- Codex app, Claude Desktop, media, and agent-loop workflows should treat generated model review surfaces as expected
- generated HTML stays under `generated/review/models/`
- generated HTML remains out of git by default unless the target repo opts in

Minimum HTML review surface expectations:

- source link
- issue link
- evidence links
- freshness/staleness state
- changed-elements summary
- before/after renders for model diffs
- responsive side-by-side and stacked compare layouts
- skip link, landmarks, local table of contents
- table captions and header scopes where tables are used

## Checker Gaps

Before default-on, resolve these contract issues:

- `dependency` and `architecture-space` are advertised model kinds but do not currently fit the strict `uml|uwe|c4` method mapping
- the model template currently permits `method: none`, but the checker does not
- strict checker requirements could break legacy loose `model-view` manifests
- `artifacts:check` must converge on rerun and include modeling checks only when the repo's modeling mode requires them
- generated HTML safety checks do not yet enforce the minimum structural accessibility contract

## PR And Handoff Requirements

UML-first PRs should include:

- Beads issue link
- changed canonical model sources
- generated review surface, if present
- before/after meaning
- evidence links
- stale-risk note
- model verdict: `ready`, `needs-evidence`, or `inconclusive`
- human approval requirement before `gh pr create` in shared repos where policy is unclear

## Implementation Slices

1. Add explicit modeling mode:
   - `off | baseline | uml-first`
   - default fresh artifact setups to `uml-first`
   - preserve legacy reruns unless migration is explicit
   - keep `--enable-modeling` as an alias
   - add `--skip-modeling` or `--modeling-mode off`

2. Add model inventory scaffold:
   - source model inventory
   - update matrix
   - PR/handoff checklist for model-backed changes

3. Fix checker contract:
   - handle `dependency`, `architecture-space`, and method-less/generated evidence cases
   - make HTML review required by artifact type/profile/mode, not always by `ready`
   - add fixtures for default, off, legacy, migration, and strict modes

4. Improve human HTML artifacts:
   - model review generator
   - model-diff comparison surface
   - local artifact index/navigation
   - accessibility structure checks

5. Add migration support:
   - explicit rerun mode for already scaffolded repos
   - deterministic package script convergence
   - local README/checker update policy

## Decision

Proceed, but as a migration-safe default-mode redesign rather than a quick default flip.

The next implementation should start with explicit modeling modes and scaffold convergence tests, then add inventory and human review improvements.
