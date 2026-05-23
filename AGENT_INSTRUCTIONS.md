# Agent Instructions

Use this file when another agent needs to install or use `skill-harness` correctly.

## What this repo is for

`skill-harness` is both:

- the installer for the shared 45ck skill, doctrine, and agent suite
- the setup repo for project-level tooling based on `@45ck/noslop`, `45ck/agent-docs`, and optional Beads integration
- the home for embedded suite-local packs under `packs/`
- the dependency entrypoint for pack repos, doctrine repos, and single-skill repos such as `repo-branding-skill`

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
- install `45ck/agent-docs`
- install the Beads CLI by default if it is not already available
- run `agent-docs init`
- run `noslop init`
- run `bd init`
- run `agent-docs install-gates --quality`
- scaffold developer artifact guidance by default; `auto` resolves to a dual profile with canonical Markdown/TOON/specgraph sources, optional generated review surfaces, and an artifact provenance manifest

Useful variants:

```bash
./skill-harness setup-project --dir path/to/project --install-only
./skill-harness setup-project --dir path/to/project --scope workspace
./skill-harness setup-project --dir path/to/project --scope root
./skill-harness setup-project --dir path/to/project --package-manager pnpm
./skill-harness setup-project --dir path/to/project --skip-agent-docs
./skill-harness setup-project --dir path/to/project --skip-noslop
./skill-harness setup-project --dir path/to/project --skip-beads
./skill-harness setup-project --dir path/to/project --skip-artifacts
./skill-harness setup-project --dir path/to/project --skip-developer-artifacts
./skill-harness setup-project --dir path/to/project --developer-artifacts-profile codex-app
./skill-harness setup-project --dir path/to/project --developer-artifacts-profile claude-desktop
./skill-harness setup-project --dir path/to/project --developer-artifacts-profile cli
./skill-harness setup-project --dir path/to/project --developer-artifacts-profile tui
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

## Full Toolkit Setup

When bootstrapping a new project manually (without the `setup-project` command), install the complete toolkit in this order:

### 1. Install specgraph (agent-docs)

```bash
npm install --save-dev @45ck/agent-docs
npx specgraph init
```

### 2. Install noslop

```bash
npm install -g @45ck/noslop
noslop init
```

### 3. Install skill packs

```bash
./skill-harness install --packs specgraph-skills,noslop-skills --packs-only
```

### What you get

- **specgraph**: spec verification engine, evidence tracking, gap analysis
- **noslop**: quality gates (pre-commit + pre-PR), content-aware config protection
- **Skills**: 5 specgraph workflow skills + 3 noslop quality gate skills

For the fully automated equivalent of the above, use the `setup-project` command described in the previous section — it installs both tools, runs their init commands, and sets up git hooks in one step.

## Developer Artifact Rules

- Keep canonical decisions, specs, handoffs, and model sources in Markdown, TOON, specgraph-compatible sources, or existing project docs.
- Treat generated HTML under `generated/review/` as a human review surface only.
- Record generated review surfaces in `docs/artifacts/artifacts.manifest.json` with source path, artifact type, evidence links, and freshness data when available.
- For Mermaid, C4, UML-style, UWE-inspired, dependency, and architecture-space views, keep the diagram/model source durable and pre-render diagrams into HTML as inline SVG or static markup.
- Do not add browser Mermaid runtimes, external scripts, or external assets to generated review HTML by default.

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

## Reference files

- [README.md](README.md)
- [cmd/skill-harness/main.go](cmd/skill-harness/main.go)
- [docs/developer-artifacts.md](docs/developer-artifacts.md)
- [scripts/dependencies.json](scripts/dependencies.json)
- [scripts/build_release.py](scripts/build_release.py)
