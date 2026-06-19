# Agent Instructions

Use this file when another agent needs to install or use `skill-harness` correctly.

## What this repo is for

`skill-harness` is both:

- the installer for the shared 45ck skill, doctrine, and agent suite
- the setup repo for project-level tooling based on `@45ck/noslop`, `45ck/agent-docs`, and optional Beads integration
- the home for embedded suite-local packs under `packs/`
- the dependency entrypoint for pack repos, doctrine repos, and single-skill repos such as `repo-branding-skill`

For public contribution, support, and security expectations, use [CONTRIBUTING.md](CONTRIBUTING.md), [SUPPORT.md](SUPPORT.md), and [SECURITY.md](SECURITY.md). This file is the operational guide for agents and maintainers.

## Agent-native bootstrap

The preferred downstream setup path is prompt-first. A human should be able to open a target repo in Codex, Claude Code, or another coding agent and ask the agent to bootstrap the repo from the `skill-harness` baseline. The agent should inspect the repo, choose the smallest useful setup, use repo-local overlay config for customizations, ask before approval-sensitive side effects, and leave setup/update evidence in `.skill-harness/`.

Use [docs/agent-native-bootstrap.md](docs/agent-native-bootstrap.md) for the copyable bootstrap prompt and operational rules. The planning source is [docs/artifacts/source/agent-native-bootstrap-update-plan-2026-05-24.md](docs/artifacts/source/agent-native-bootstrap-update-plan-2026-05-24.md).

Useful commands:

```bash
./skill-harness audit-project --dir path/to/project
./skill-harness resolve --dir path/to/project
./skill-harness bootstrap --dir path/to/project --agent-native
./skill-harness install --dir path/to/project
./skill-harness render --dir path/to/project
./skill-harness check --dir path/to/project
./skill-harness update-project --dir path/to/project --write-lock
```

`bootstrap --agent-native` creates the repo-local overlay, resolved lock, and setup proof without package installs or global writes. After that, `install --dir`, `render --dir`, and `check --dir` use only the resolved effective stack unless explicit agents or packs are supplied.

Ask before running `setup-project`, package installs, monorepo-root setup, global or home-directory writes, `.claude/settings.json` permission changes, Beads install/init, Git hook changes, CI changes, publishing, deployment, or destructive actions.

## Shared suite install

Run this from the repo root when the goal is to install the shared packs and agents:

```bash
go build -o skill-harness ./cmd/skill-harness
./skill-harness install --all
```

Windows:

```powershell
go build -o skill-harness.exe .\cmd\skill-harness
.\skill-harness.exe install --all
```

Selective install examples:

```bash
./skill-harness install --agents=requirements-analyst,system-modeler
./skill-harness install --packs=business-analysis-skills,documentation-evidence-skills --packs-only
./skill-harness install --packs=frontier-agent-playbook,repo-branding-skill --packs-only
./skill-harness install --interactive
```

## Project setup

Run this when the goal is to bootstrap a target repo with the 45ck project tooling stack:

```bash
./skill-harness setup-project --dir path/to/project
```

Default behavior:

- auto-detect monorepo roots from workspace markers and default to the monorepo root when the target path is inside one
- auto-detect `npm`, `pnpm`, `yarn`, or `bun` from lockfiles or `packageManager`
- create `package.json` in the resolved setup directory if missing
- install `@45ck/noslop`
- install `github:45ck/agent-docs`
- install the Beads CLI by default if it is not already available
- run `agent-docs init`
- run `noslop init`
- run `bd init`
- run `agent-docs install-gates --quality`
- write `.claude/settings.json` to allow Claude agent-team use unless `--skip-claude-settings` is passed
- install the repo-local Beads worktree wrapper unless `--beads-worktrees=false` is passed
- scaffold developer artifact guidance by default; `auto` resolves to a dual profile with canonical Markdown/TOON/specgraph sources, generated HTML human review surfaces, and an artifact provenance manifest
- scaffold visual-source-first artifact guidance for product, business, data, research, UX, and mockup work, with agent-readable sources and generated visual human review surfaces
- use `--developer-artifacts-profile agent-loop` for governed self-improving agent workflows with a loop playbook, trace/eval receipt directory, and policy checker
- use `--modeling-mode auto` by default; it preserves legacy repos and resolves fresh developer-artifact setups to UML-first modeling with `docs/artifacts/source/models/`, `generated/review/models/`, `model-diff` manifest entries, HTML model review generation, and model policy checks
- use `--modeling-mode off|baseline|uml-first`, `--skip-modeling`, or legacy alias `--enable-modeling` only when the target mode is intentional
- write `.skill-harness/setup-proof.json` as machine-readable evidence of selected profile, package manager, initialized tools, skipped capabilities, generated paths, Beads mode, and available check commands

Useful variants:

```bash
./skill-harness setup-project --dir path/to/project --install-only
./skill-harness setup-project --dir path/to/project --scope workspace
./skill-harness setup-project --dir path/to/project --scope root
./skill-harness setup-project --dir path/to/project --package-manager pnpm
./skill-harness setup-project --dir path/to/project --skip-agent-docs
./skill-harness setup-project --dir path/to/project --skip-noslop
./skill-harness setup-project --dir path/to/project --skip-beads
./skill-harness setup-project --dir path/to/project --skip-claude-settings
./skill-harness setup-project --dir path/to/project --beads-worktrees=false
./skill-harness setup-project --dir path/to/project --skip-artifacts
./skill-harness setup-project --dir path/to/project --skip-developer-artifacts
./skill-harness setup-project --dir path/to/project --developer-artifacts-profile codex-app
./skill-harness setup-project --dir path/to/project --developer-artifacts-profile claude-desktop
./skill-harness setup-project --dir path/to/project --developer-artifacts-profile cli
./skill-harness setup-project --dir path/to/project --developer-artifacts-profile tui
./skill-harness setup-project --dir path/to/project --developer-artifacts-profile media
./skill-harness setup-project --dir path/to/project --developer-artifacts-profile agent-loop
./skill-harness setup-project --dir path/to/project --modeling-mode uml-first
./skill-harness setup-project --dir path/to/project --modeling-mode baseline
./skill-harness setup-project --dir path/to/project --skip-modeling
./skill-harness setup-project --dir path/to/project --enable-modeling
```

## Frontier doctrine companion

Use [`45ck/frontier-agent-playbook`](https://github.com/45ck/frontier-agent-playbook) when the target repo should carry shared frontier-agent doctrine in addition to installed skills.

Install its skills globally through `skill-harness`:

```bash
./skill-harness install --packs=frontier-agent-playbook --packs-only
```

For repo-local doctrine files, copy these into the target project after setup:

- `AGENTS.md`
- `CLAUDE.md`
- `AGENT_INSTRUCTIONS.md`
- `llms.txt`

Use the embedded `agent-operating-skills` pack when frontier doctrine needs to become an operating workflow: task shaping, context engineering, autonomy boundaries, tool permissions, memory review, multi-agent handoffs, and agent run evidence review.

```bash
./skill-harness install --packs=agent-operating-skills --packs-only
```

## Full Toolkit Setup

When bootstrapping a new project manually (without the `setup-project` command), install the complete toolkit in this order:

### 1. Install agent-docs

```bash
npm install --save-dev github:45ck/agent-docs
npx agent-docs init
```

### 2. Install noslop

```bash
npm install -g @45ck/noslop
noslop init
```

### 3. Install skill packs

```bash
./skill-harness install --packs=specgraph-skills,noslop-skills --packs-only
```

### What you get

- **agent-docs/specgraph**: spec verification engine, evidence tracking, gap analysis
- **noslop**: quality gates (pre-commit + pre-PR), content-aware config protection
- **Skills**: 5 specgraph workflow skills + 3 noslop quality gate skills

For the fully automated equivalent of the above, use the `setup-project` command described in the previous section — it installs both tools, runs their init commands, and sets up git hooks in one step.

## Developer Artifact Rules

- Keep canonical decisions, specs, handoffs, and model sources in Markdown, TOON, specgraph-compatible sources, or existing project docs.
- For product, business, data, research, UX, and mockup work, keep canonical agent-readable sources in `docs/artifacts/source/<family>/` unless a domain-specific docs path is better, then generate visual human review surfaces under `generated/review/<family>/`.
- High-fidelity HTML/prototype review is the default for UI, product, customer-facing workflow, and mockup artifacts. Low-fidelity sketches are scratch only unless explicitly captured as research evidence.
- Use the infographic toolkit for human artifact charts and graphs: Mermaid, Vega-Lite, Observable Plot, D3, Graphviz, Apache ECharts, RAWGraphs, and Chart.js are source/spec or generation-time choices only. Generated HTML must stay static and must not load their browser runtimes.
- Prefer `artifact-infographic` JSON fences or manifest `infographics` arrays for non-model charts and graphs so generated HTML can be refreshed deterministically.
- For multi-lane planning, launch-readiness, project-management, chat-history, and "what next" artifacts, put a static Kanban/status board near the top with Done / Doing / Planned / Blocked / Idle lanes, familiar status icons, task links or IDs, and one-sentence acceptance evidence per card.
- Use specialist agent teams when artifact ownership crosses real boundaries: requirements for product, delivery for business, backend for data, research for evidence, UX for high-fidelity review, system-modeler for structural impact, and quality-reviewer for readiness gates.
- Treat generated HTML under `generated/review/` as a human review surface only.
- Open generated HTML with the best available human surface. In Codex app, use the Browser plugin for local HTML when available. In Claude desktop, use the built-in browser or preview when available. In CLI-only contexts, use `node scripts/open-artifact-review.mjs` to open the system default browser or print the file URL in headless/CI contexts.
- Use `node scripts/open-artifact-review.mjs --json --print` when an app or agent needs a machine-readable target path, URL, and preferred host action before opening the review surface.
- Record generated review surfaces in `docs/artifacts/artifacts.manifest.json` with source path, artifact type, evidence links, and freshness data when available.
- Label synthetic user, simulated customer, or agent-generated evidence separately from real user/customer evidence in source artifacts, manifests, and review surfaces.
- For Mermaid, C4, UML-style, UWE-inspired, dependency, and architecture-space views, keep the diagram/model source durable and pre-render diagrams into HTML as inline SVG or static markup.
- Auto-detect model impact for every engineering change. If code, API, workflow, dependency, deployment, UI structure, or agent behavior changes, update the relevant canonical model source or record why no model change is needed.
- In UML-first or baseline modeling mode, keep authored model sources in repo-relative text files, prefer `docs/artifacts/source/models/` when there is no better domain docs path, update `docs/artifacts/source/models/model-inventory.md`, put generated human HTML model reviews under `generated/review/models/`, run `node scripts/generate-model-review.mjs`, and validate with `node scripts/check-model-artifact-policy.mjs`.
- For human-facing discovery, planning, research, product, business, data, UX, and mockup artifacts, create canonical source under `docs/artifacts/source/<family>/`, add a manifest entry with `reviewRequired: true`, run `node scripts/generate-artifact-review.mjs`, and hand off the generated infographic HTML under `generated/review/<family>/`. Do not substitute Markdown-only output when a human review surface was requested.
- Do not add browser Mermaid runtimes, external scripts, or external assets to generated review HTML by default.
- Use `--developer-artifacts-profile agent-loop` only for governed self-improving agent workflows; generated trace receipts belong under `generated/agent-runs/` and should stay out of git unless summarized and redacted.

## Rules

- Run commands from the repo root unless the command explicitly targets another directory.
- Prefer the CLI over manual copying.
- Do not assume dependency repos are already installed.
- Do not assume `noslop` or `agent-docs` are already present in the target project.
- Do not assume repo-local doctrine files are already present unless they were copied from `frontier-agent-playbook`.
- Use `setup-project` for project scaffolding instead of inventing a custom sequence.

## Verify

After shared-suite installation:

```bash
./skill-harness check --all
```

After project setup:

- confirm `package.json` exists in the resolved setup directory
- confirm `@45ck/noslop` and `agent-docs` were installed
- confirm the initialization commands completed without error
- inspect `.skill-harness/setup-proof.json` for the resolved scope, tool statuses, skipped capabilities, and check commands

## Reference files

- [README.md](README.md)
- [CONTRIBUTING.md](CONTRIBUTING.md)
- [SECURITY.md](SECURITY.md)
- [SUPPORT.md](SUPPORT.md)
- [cmd/skill-harness/main.go](cmd/skill-harness/main.go)
- [docs/developer-artifacts.md](docs/developer-artifacts.md)
- [docs/dependency-provenance.md](docs/dependency-provenance.md)
- [docs/agent-operating-skills.md](docs/agent-operating-skills.md)
- [docs/demo-production-media.md](docs/demo-production-media.md)
- [scripts/dependencies.json](scripts/dependencies.json)
- [scripts/build_release.py](scripts/build_release.py)
