# Agent-Native Bootstrap

`skill-harness` should be usable without a human manually learning the install flow. The expected modern workflow is that a human opens a target repo in Codex, Claude Code, or a similar coding agent and gives it a bootstrap prompt. The agent then inspects the repo, chooses a small useful stack, applies setup through deterministic harness commands, and leaves evidence for future agents.

The detailed planning source is [agent-native-bootstrap-update-plan-2026-05-24.md](artifacts/source/agent-native-bootstrap-update-plan-2026-05-24.md).

Implementation note: the current CLI implements the repo-local desired-state contract, lock/proof scaffolding, and resolved install/render/check behavior. Full three-way update reconciliation across old lock, new baseline, and local overlay remains follow-up work.

## Core Idea

Use `skill-harness` as an upstream baseline, not as files that downstream repos permanently fork.

Target repos should carry repo-local intent in `.skill-harness/`:

- `agent-stack.json`: desired baseline, profile, opt-outs, and local overrides
- `agent-stack.lock.json`: resolved baseline revision, effective packs, agents, skills, and hashes
- `setup-proof.json`: what setup did, what it skipped, what checks ran, and what needs follow-up

Generated agent files and installed skills are derived from those sources. If a repo needs to modify an agent or skill loadout, it should express that as an overlay rather than editing generated files directly.

## Bootstrap Prompt

Paste this into Codex or Claude Code from the target repo:

````md
Use the 45ck skill-harness baseline in this repository.

Inspect the repo first. Choose the smallest useful setup. For repo governance use `minimal`, `team`, or `agent-native`; for agent-stack overlays use `minimal`, `default`, or `security`; only add artifact, media, security, or agent-loop capabilities when the repo clearly needs them. Prefer repo-local overlays over editing generated agent or skill files directly. Preserve the ability to receive upstream skill-harness updates.

Before making changes, report:
- detected language, package manager, monorepo layout, CI, tests, docs, and existing agent/tooling files
- proposed profile and skipped capabilities
- local files you expect to write
- package installs, monorepo-root setup, global/home-directory writes, Claude settings or permission changes, Beads install/init, Git hook changes, CI changes, network operations, or destructive actions that need approval

If skill-harness is not already available, fetch or use a local checkout from https://github.com/45ck/skill-harness and build the CLI. Run the safe repo-local bootstrap flow before `setup-project` or other side-effecting setup:

```sh
skill-harness audit-project --dir .
skill-harness bootstrap --dir . --agent-native
skill-harness resolve --dir .
```

After approval, configure the repo, run available checks, and leave setup evidence in `.skill-harness/`.
````

For a lean setup, add:

```md
Keep the setup minimal. Do not enable media, demo, pentest, or agent-loop capabilities unless this repo clearly needs them. Use generated HTML review surfaces only when the active workflow benefits from desktop/browser review.
```

For a governed agent-loop repo, add:

```md
Use the governed agent-loop profile only if the repo needs recurring agent improvement, trace receipts, eval summaries, and explicit learning proposals.
```

## Agent Workflow

The agent should follow this sequence:

1. Inspect repo shape and existing tooling.
2. Decide whether the repo is unmanaged, generated-only, already harnessed, or agent-native.
3. Propose the smallest useful setup choices and opt-outs.
4. Locate or fetch `skill-harness`.
5. Create or update `.skill-harness/agent-stack.json`.
6. Run a read-only resolution before mutating setup.
7. Ask before package installs, global writes, hooks, CI changes, permission changes, publishing, or destructive actions.
8. Apply setup and render derived files after approval.
9. Run available checks.
10. Write lock/proof evidence and update docs for future agents.

## CLI Support

The prompt-first flow is backed by deterministic commands:

```sh
skill-harness audit-project --dir .
skill-harness bootstrap --dir . --agent-native
skill-harness resolve --dir .
skill-harness update-project --dir . --write-lock
skill-harness install --dir .
skill-harness render --dir .
skill-harness check --dir .
```

`resolve` is read-only. It computes effective packs, agents, skills, opt-outs, repo-local additions, and policy-sensitive changes from the upstream baseline plus `.skill-harness/agent-stack.json`.

`bootstrap --agent-native` scaffolds the repo-local overlay when it is missing, writes `.skill-harness/agent-stack.lock.json`, writes `.skill-harness/setup-proof.json`, and prints the next safe action. It does not silently install packages or write global agent state.

`install --dir`, `render --dir`, and `check --dir` consume the resolved agent stack when the user has not supplied explicit `--agents` or `--packs` flags. These side-effecting commands require `.skill-harness/agent-stack.json`; run `bootstrap --agent-native` first in a fresh repo.

`update-project` reads the overlay, reports resolved agent-stack state, and can write `.skill-harness/agent-stack.lock.json` with `--write-lock`. It is intentionally dry-run unless asked to write the lockfile.

`audit-project` tells an agent whether a repo is unmanaged, generated-only, agent-native, or conflicted. `repo audit`, `repo drift`, `repo sync`, `repo update --check`, and `repo trim --dry-run` expose the lower-level baseline-manifest governance surface. `repo sync` writes `.skill-harness/baseline.lock.json`, not the agent-stack lock.

## Overlay Rules

Use structured operations first:

- enable or disable packs
- enable or disable agents
- add a skill to an agent
- remove a skill from an agent
- replace a skill with a repo-local skill
- shadow a baseline skill only when explicitly recorded
- pin or change baseline channel

Avoid line-level patching of upstream `SKILL.md` files. Prompt prose is not a good merge target. If a repo needs different behavior, prefer a repo-local companion skill or explicit replacement.

## Update Flow

Future updates should be reconciliation, not blind reinstall.

1. Read `.skill-harness/agent-stack.json`.
2. Read `.skill-harness/agent-stack.lock.json`.
3. Fetch or locate the requested baseline channel.
4. Resolve latest baseline plus repo overlay.
5. Compare old lockfile, new baseline, and local overlay.
6. Classify changes.
7. Print a dry-run update report.
8. Ask before approval-sensitive changes.
9. Re-render generated outputs.
10. Run checks.
11. Update lockfile and setup proof.
12. File follow-up issues for unresolved conflicts.

Change classes:

- `baseline-addition`
- `baseline-removal`
- `overlay-addition`
- `overlay-removal`
- `shadow`
- `orphaned-override`
- `policy-sensitive`
- `detached-generated-file`
- `clean`

## Approval Boundaries

Agents may usually do these after the user asks for bootstrap or planning:

- inspect files
- draft `.skill-harness/agent-stack.json`
- run read-only resolution
- write local docs or setup evidence
- run local checks

Agents should ask before:

- package installs
- global writes under user home
- Git hook changes
- CI workflow changes
- Beads initialization in a target repo
- enabling agent teams or tool permissions
- changing security, privacy, legal, or compliance policy

Agents need explicit approval for:

- pushing
- publishing
- deployment
- production data access
- credential handling
- destructive filesystem, database, or infrastructure actions
- permission expansion

## Profile Direction

Initial setup choices should be conservative and use current CLI vocabulary:

| Surface | Values | Intent |
| --- | --- |
| Agent-stack profile | `minimal`, `default`, `security` | Controls effective agents and packs in `.skill-harness/agent-stack.json`. |
| Repo governance profile | `minimal`, `team`, `agent-native` | Controls baseline manifest defaults for managed downstream repos. |
| Developer artifact profile | `auto`, `codex-app`, `claude-desktop`, `cli`, `tui`, `media`, `agent-loop`, `none` | Controls source-first docs, generated review surfaces, media/demo scaffolds, or governed agent-loop scaffolds. |
| Modeling mode | `auto`, `off`, `baseline`, `uml-first` | Controls model source and generated model review policy. |

Profiles describe defaults. Repos can still opt out of specific packs or agents through the overlay.

## Implementation Order

1. Document the bootstrap prompt and update model.
2. Define `.skill-harness/agent-stack.json`.
3. Implement read-only resolution.
4. Add tests and fixture baselines.
5. Extend `setup-project` to scaffold overlay config.
6. Teach install/render/check to use resolved state.
7. Implement `update-project` reconciliation.
8. Add `audit-project`.
9. Update canonical models and generated review surfaces.

Current implementation status:

- `agent-stack.json` schema and default scaffold exist.
- `resolve` computes effective agents, packs, skills, opt-outs, overlays, and diagnostics.
- `setup-project` writes `.skill-harness/agent-stack.json` without overwriting an existing stack.
- `bootstrap --agent-native` writes the overlay, resolved agent-stack lock, and setup proof without installing packages or writing global agent files.
- `install --dir`, `render --dir`, and `check --dir` consume resolved state and fail fast when a target repo has not been bootstrapped with `agent-stack.json`.
- `audit-project`, `repo audit`, `repo drift`, `repo sync`, `bootstrap --agent-native`, and `update-project --write-lock` provide the first audit/update surfaces.

## Done Criteria

This capability is ready when an agent can enter a fresh repo, receive the bootstrap prompt, and:

- inspect the repo
- select a lean profile
- create repo-local overlay config
- run read-only resolution
- ask before side effects
- apply setup
- render effective agents
- skip irrelevant packs
- leave setup proof and lockfile evidence
- update later without losing repo-local customization
- report conflicts clearly when automatic update is unsafe
