# skill-harness

<p align="center">
  <img src="logo.svg" alt="skill-harness logo" width="128" height="128" />
</p>

`skill-harness` is the setup repo for the 45ck agent workflow stack. It is primarily an installer and generator: most downstream repos receive copied Claude/Codex agent definitions or project setup changes rather than importing `skill-harness` as an application runtime.

[![quality](https://github.com/45ck/skill-harness/actions/workflows/quality.yml/badge.svg)](https://github.com/45ck/skill-harness/actions/workflows/quality.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

It does five jobs:

- installs the shared dependency-repo suite across pack repos, doctrine repos, and single-skill repos
- installs the shared Claude and Codex agents
- bootstraps project-level tooling with `@45ck/noslop` and `45ck/agent-docs`
- optionally bootstraps Beads, enabled by default in project setup
- hosts embedded packs for suite-local or incubating capabilities

## Copy this into your agent to use it

Open the target repo in Codex, Claude Code, or another coding agent, then paste this:

````md
Use the 45ck skill-harness baseline in this repository.

First inspect the repo and report:
- language, package manager, monorepo layout, CI, tests, docs, and existing agent/tooling files
- the smallest useful skill-harness setup: usually repo profile `minimal`, `team`, or `agent-native`; only add artifact, media, security, or agent-loop capabilities when the repo clearly needs them
- files you expect to write
- package installs, monorepo-root setup, global/home-directory writes, Claude settings or permission changes, Beads install/init, Git hook changes, CI changes, network operations, or destructive actions that need approval

If skill-harness is not already available, fetch or use a local checkout from https://github.com/45ck/skill-harness and build the CLI:

```sh
git clone https://github.com/45ck/skill-harness.git .skill-harness-upstream
cd .skill-harness-upstream
go build -o skill-harness ./cmd/skill-harness
```

Then return to this target repo and run the safe, repo-local bootstrap flow first:

```sh
../.skill-harness-upstream/skill-harness audit-project --dir .
../.skill-harness-upstream/skill-harness bootstrap --dir . --agent-native
../.skill-harness-upstream/skill-harness resolve --dir .
```

Do not run `setup-project`, install packages, write global agent files, change hooks, change Claude permissions, initialize Beads, or change CI until approval is given. After approval, apply the selected setup, run available checks, and leave `.skill-harness/agent-stack.json`, `.skill-harness/agent-stack.lock.json`, and `.skill-harness/setup-proof.json` as evidence for future agents.
````

PowerShell agents should build `skill-harness.exe` and use `..\.skill-harness-upstream\skill-harness.exe` for the same commands.

This prompt is the preferred installer surface for downstream repos: the agent chooses a lean profile, the CLI gives deterministic resolution, and repo-local overlays preserve future updates.

## Two ways to use skill-harness

- **Agent-native bootstrap, recommended for downstream repos:** paste the prompt above into an agent running inside the target repo. The agent inspects first, proposes a minimal profile, asks before side effects, runs the deterministic bootstrap commands, and leaves `.skill-harness/` evidence.
- **Direct CLI/manual setup:** build or obtain the CLI yourself, then run `install`, `setup-project`, `render`, `check`, or repo governance commands directly. This is still useful for maintainers, release bundles, and controlled local installs.

## Project status

This repo is open source under the [MIT License](LICENSE). It is actively evolving, and the default branch is the source of truth until formal release channels are established.

The root [package.json](package.json) is marked `private` because npm is used here for local policy and artifact scripts, not for publishing the CLI as an npm package. Build the Go CLI locally or from a release bundle.

## Trust boundaries

`skill-harness` can write into user-level agent directories, clone or copy skill packs, install project tooling, run setup commands, and create generated review surfaces in target repos. Review scripts and dependency changes as supply-chain-sensitive.

The `setup-project` command runs inside the resolved setup directory. In `--scope auto`, that may be a detected monorepo root rather than the exact directory passed with `--dir`. It can:

- create or update `package.json`
- install `@45ck/noslop` and `github:45ck/agent-docs` as development dependencies
- run `agent-docs init`, `noslop init`, `bd init`, and `agent-docs install-gates --quality`
- install Beads if `bd` is not already available
- write `.claude/settings.json` to allow Claude agent-team use
- install the repo-local Beads worktree wrapper by default
- create `.skill-harness/`, `docs/artifacts/`, `generated/review/`, helper scripts, and package scripts

Beads installation first uses an existing `bd`, then tries `go install github.com/steveyegge/beads/cmd/bd@latest`, then falls back to the upstream Beads install script (`irm ... | iex` on Windows or `curl ... | bash` on Unix). Use `--skip-beads` when you want to install or review Beads manually.

Use `--install-only` to install packages without running initialization commands, `--scope workspace` to avoid monorepo-root lifting, `--skip-claude-settings` to avoid changing Claude tool permissions, `--beads-worktrees=false` to skip the Beads worktree wrapper, and the other `--skip-*` flags to reduce setup side effects.

Dependency source and pinning details are documented in [docs/dependency-provenance.md](docs/dependency-provenance.md).

Generated HTML under `generated/review/` is a human review surface only. Canonical decisions, specs, models, and evidence stay in source files under `docs/`, `packs/`, scripts, or target repo scaffolds. Human-facing discovery and planning artifacts should use source plus generated infographic HTML, not Markdown-only handoff, when the active workflow is a desktop or browser review surface.

## What it can set up

### Shared suite

- remote dependency repos plus embedded local packs under `packs/`
- shared skills synced into `~/.claude/skills/` and `~/.agents/skills/`
- supports repos exposed as `skills/`, packaged `.claude/.agents` mirrors, or a single root `SKILL.md`
- shared Claude agents
- shared Codex agents

### Agent-native bootstrap

Humans do not need to learn the full install flow before using the stack. The intended modern path is prompt-first: open a target repo in Codex, Claude Code, or a similar coding agent and ask it to bootstrap the repo from the `skill-harness` baseline.

Use the copy-paste agent prompt near the top of this README for first contact. See [docs/agent-native-bootstrap.md](docs/agent-native-bootstrap.md) for the longer bootstrap prompt, repo-local overlay model, update flow, and approval boundaries. The durable planning source is [docs/artifacts/source/agent-native-bootstrap-update-plan-2026-05-24.md](docs/artifacts/source/agent-native-bootstrap-update-plan-2026-05-24.md).

Agent-native repos use `.skill-harness/agent-stack.json` as desired state and can be inspected or updated with:

```bash
./skill-harness audit-project --dir ../my-project
./skill-harness resolve --dir ../my-project
./skill-harness bootstrap --dir ../my-project --agent-native
./skill-harness update-project --dir ../my-project --write-lock
./skill-harness render --dir ../my-project
./skill-harness check --dir ../my-project
```

`bootstrap --agent-native` writes the desired-state overlay, `.skill-harness/agent-stack.lock.json`, and `.skill-harness/setup-proof.json` without installing packages or writing global agent files. `install --dir`, `render --dir`, and `check --dir` use the resolved stack and require that overlay to exist. Full three-way reconciliation across an old lockfile, a new upstream baseline, and local overlays remains follow-up work; current update commands provide resolution, lock refresh, and conservative reports.

For ongoing distribution across active projects, use the repo governance commands. They add a repo-local baseline manifest and report how the shared harness should interact with project-owned files, overlays, and generated outputs:

```bash
./skill-harness repo init --dir ../my-project --profile team
./skill-harness repo audit --dir ../my-project
./skill-harness repo sync --dir ../my-project
./skill-harness repo update --dir ../my-project --check
./skill-harness repo trim --dir ../my-project --dry-run
```

`repo init` writes `.skill-harness/baseline.manifest.json`. `repo sync` writes `.skill-harness/baseline.lock.json` and `.skill-harness/update-report.json`. Existing repo-local skills, agents, Beads state, and project instructions are classified as `overlay` or `owned` surfaces so baseline updates can be received without deleting project-specific customization. Update and trim are intentionally observe-only until a repo chooses which surfaces and capabilities to accept, ignore, or opt out of.

### Project tooling

- [`45ck/noslop`](https://github.com/45ck/noslop)
- [`45ck/agent-docs`](https://github.com/45ck/agent-docs)
- [`steveyegge/beads`](https://github.com/steveyegge/beads)

Use the project setup command when you want a repo scaffolded with the 45ck tooling stack:

```bash
./skill-harness setup-project --dir path/to/project
```

That command:

- auto-detects monorepo roots from workspace markers such as `pnpm-workspace.yaml`, `package.json` workspaces, `nx.json`, `turbo.json`, `lerna.json`, and `rush.json`
- auto-detects `npm`, `pnpm`, `yarn`, or `bun` from lockfiles or `packageManager`
- defaults to monorepo-root setup when the target directory is inside a detected monorepo
- creates a `package.json` in the resolved setup directory if one does not exist yet
- installs `@45ck/noslop` and `github:45ck/agent-docs`
- installs the Beads CLI if it is not already present
- runs `agent-docs init`
- runs `noslop init`
- runs `bd init`
- runs `agent-docs install-gates --quality`
- writes `.claude/settings.json` to allow Claude agent-team use unless `--skip-claude-settings` is passed
- installs the repo-local Beads worktree wrapper unless `--beads-worktrees=false` is passed
- scaffolds developer artifact guidance by default, with Markdown/TOON/JSON/YAML/specgraph sources, visual-source-first product/business/data/research/UX/mockup artifacts, UML/UWE/C4/evidence model sources as canonical source, generated HTML as a human review surface, and a manifest for provenance/freshness, including optional governed agent-loop scaffolding
- writes `.skill-harness/setup-proof.json` with setup scope, package manager, selected profiles, tool statuses, check commands, generated paths, skipped capabilities, and Beads mode

## Benchmark results

Controlled experiments measured the toolkit (specgraph + noslop + skills) against raw Claude Code. Full data in [`experiments/RESULTS.md`](experiments/RESULTS.md). Current adoption audit notes are in [`docs/adoption-audit-2026-04-29.md`](docs/adoption-audit-2026-04-29.md).

| Experiment | Toolkit | Baseline | Delta |
|---|:---:|:---:|:---:|
| Small greenfield (×2 runs) | 31–32 / 35 | 19–20 / 35 | +12 |
| Large greenfield (3 modules) | 32 / 35 | 19 / 35 | +13 |
| Maintenance / handoff | 33 / 35 | 19 / 35 | +14 |
| Ambiguous brief | **35 / 35** | **13 / 35** | **+22** |

The gap is driven by scope enforcement and traceability — not code quality (functional output was equal in every experiment). The largest signal came from the ambiguous brief: the toolkit forced scope decisions before code was written; the baseline built a 4-class framework for a task that needed a 3-function library. Treat these as controlled workflow results; downstream repos still need active hook, CI, and agent wiring before those practices are enforced.

## Install the CLI

Prerequisites for local builds and wrapper scripts: Go, Git, and the target shell (`bash` or PowerShell). Node.js and Python are only needed for development and artifact checks.

### Build locally

```bash
git clone https://github.com/45ck/skill-harness.git
cd skill-harness
go build -o skill-harness ./cmd/skill-harness
```

Windows:

```powershell
git clone https://github.com/45ck/skill-harness.git
cd skill-harness
go build -o skill-harness.exe .\cmd\skill-harness
```

### Use wrapper scripts

```bash
bash install.sh
```

```powershell
.\install.ps1
```

The wrapper scripts are `go run` shortcuts. They still require Go and run the checked-out source tree.

### Build a release bundle

Release bundles can ship the binary plus the repo files together so Go is not required.

Build them with:

```bash
python scripts/build_release.py --version v0.1.0
```

## Main commands

### Install the full shared suite

```bash
./skill-harness install --all
```

### Install selected agents

```bash
./skill-harness install --agents=requirements-analyst,system-modeler,security-reviewer
```

### Install selected packs only

```bash
./skill-harness install --packs=business-analysis-skills,documentation-evidence-skills --packs-only
```

### Install doctrine or utility repos only

```bash
./skill-harness install --packs=frontier-agent-playbook,repo-branding-skill --packs-only
```

### Use the interactive installer

```bash
./skill-harness install --interactive
```

### Set up a project with noslop and agent-docs

```bash
./skill-harness setup-project --dir ../my-project
```

Install only the packages and skip initialization:

```bash
./skill-harness setup-project --dir ../my-project --install-only
```

Monorepo auto mode:

```bash
./skill-harness setup-project --dir ../my-monorepo/apps/web
```

Force workspace-local setup instead of lifting to the monorepo root:

```bash
./skill-harness setup-project --dir ../my-monorepo/apps/web --scope workspace
```

Override package manager if auto-detection is not what you want:

```bash
./skill-harness setup-project --dir ../my-monorepo --package-manager pnpm
```

Skip one tool:

```bash
./skill-harness setup-project --dir ../my-project --skip-agent-docs
./skill-harness setup-project --dir ../my-project --skip-noslop
./skill-harness setup-project --dir ../my-project --skip-beads
./skill-harness setup-project --dir ../my-project --skip-claude-settings
./skill-harness setup-project --dir ../my-project --beads-worktrees=false
./skill-harness setup-project --dir ../my-project --skip-artifacts
./skill-harness setup-project --dir ../my-project --skip-developer-artifacts
```

Install the repo-local Beads worktree wrapper (copies `scripts/beads/bd.mjs` + adds `.trees/` to `.gitignore`):

```bash
./skill-harness beads-worktrees --dir ../my-project
```

Audit and pin baseline governance for an existing repo:

```bash
./skill-harness repo audit --dir ../my-project
./skill-harness repo init --dir ../my-project --profile minimal
./skill-harness repo sync --dir ../my-project
./skill-harness repo drift --dir ../my-project
./skill-harness repo update --dir ../my-project --check
./skill-harness repo trim --dir ../my-project --dry-run
```

Profiles choose the default effective agent set: `minimal` keeps the lowest-cost delivery and quality loadout, `team` adds requirements and architecture, and `agent-native` adds workflow orchestration. The manifest can opt agents and packs in or out per repo, while surface modes describe ownership: `generated`, `managed-section`, `overlay`, `owned`, or `ignored`.

Choose a developer artifact profile:

```bash
./skill-harness setup-project --dir ../my-project --developer-artifacts-profile auto
./skill-harness setup-project --dir ../my-project --developer-artifacts-profile codex-app
./skill-harness setup-project --dir ../my-project --developer-artifacts-profile claude-desktop
./skill-harness setup-project --dir ../my-project --developer-artifacts-profile cli
./skill-harness setup-project --dir ../my-project --developer-artifacts-profile tui
./skill-harness setup-project --dir ../my-project --developer-artifacts-profile media
./skill-harness setup-project --dir ../my-project --developer-artifacts-profile agent-loop
./skill-harness setup-project --dir ../my-project --developer-artifacts-profile none
```

Artifact profiles are guidance and scaffold settings, not a separate runtime. `auto` is the default and resolves to `dual`: canonical Markdown/TOON/specgraph sources plus optional generated review surfaces. Use `codex-app` or `claude-desktop` for desktop workflows with file-backed previews, `cli` or `tui` for terminal-heavy projects, `media` for source-backed demo and generated media review workflows, `agent-loop` for governed self-improving agent workflows with trace/eval receipts, and `none`, `--skip-artifacts`, or `--skip-developer-artifacts` for minimal projects. The shorter `--artifact-profile media|agent-loop|markdown|html|dual|none` form remains supported as an alias.

Developer artifact scaffolding also creates `docs/artifacts/artifacts.manifest.json`, `scripts/check-artifact-manifest.mjs`, and `scripts/generate-artifact-review.mjs`. Use the manifest to record source-backed generated views, including product briefs, business dashboards, data dictionaries, metric definitions, research evidence boards, high-fidelity UX prototypes, Mermaid, C4, UML-style, dependency, architecture-space, demo media, and agent-loop artifacts. Generated HTML should be the human artifact: a self-contained review page with clear sections, tabs, diagrams or static previews, screenshots or evidence images, text summaries, source links, and freshness metadata. For non-model human artifacts, set `reviewRequired: true` and run `npm run artifacts:generate` or `node scripts/generate-artifact-review.mjs` to produce infographic-style HTML with metrics, charts, evidence/freshness panels, and source links. The infographic toolkit treats Mermaid, Vega-Lite, Observable Plot, D3, Graphviz, Apache ECharts, RAWGraphs, and Chart.js as source/spec or generation-time renderers; generated HTML embeds only static SVG/HTML/data-url output and does not load browser runtimes. Visual-source-first artifact families use canonical sources under `docs/artifacts/source/product/`, `business/`, `data/`, `research/`, and `ux/`, with generated reviews under matching `generated/review/` subfolders. High-fidelity review is the default for UI, product, customer-facing workflow, and mockup artifacts; low-fidelity sketches are scratch unless captured as evidence. Agents should auto-detect model impact for engineering changes and update the relevant model source, manifest, and human HTML review artifact when code, API, workflow, dependency, deployment, UI structure, or agent behavior changes. The `media` profile also creates `generated/media/` and `docs/artifacts/templates/demo-artifact.md`; generated media stays out of git by default. The `agent-loop` profile creates `generated/agent-runs/`, `docs/artifacts/source/agent-loop-playbook.md`, `docs/artifacts/templates/agent-loop-artifact.md`, and `scripts/check-agent-loop-policy.mjs`.

Modeling mode defaults to `auto`. Fresh developer-artifact setups resolve `auto` to UML-first modeling; existing harnessed repos keep their current behavior unless migrated. Use explicit modes when needed:

```bash
./skill-harness setup-project --dir ../my-project --modeling-mode auto
./skill-harness setup-project --dir ../my-project --modeling-mode uml-first
./skill-harness setup-project --dir ../my-project --modeling-mode baseline
./skill-harness setup-project --dir ../my-project --skip-modeling
```

UML-first mode keeps modeling inside the normal artifact system and adds `docs/artifacts/source/models/`, `docs/artifacts/source/models/model-inventory.md`, `generated/review/models/`, `docs/artifacts/templates/model-diff-artifact.md`, `scripts/generate-model-review.mjs`, `scripts/open-artifact-review.mjs`, `scripts/check-model-artifact-policy.mjs`, model-aware package scripts, and setup-proof check entries. `model-view` still exists in the base scaffold; UML-first adds stricter UML/UWE/C4/evidence policy, `model-diff` support, and generated human HTML review expectations. `--enable-modeling` remains as a legacy alias for `--modeling-mode uml-first`.

Use `npm run artifacts:review` to regenerate non-model infographic HTML plus model review HTML and run policy checks. Use `npm run artifacts:generate` when you only need to refresh generated review files. Use `npm run models:review` to regenerate model review HTML and `npm run models:drift` to fail on stale generated model review files. CI should use the drift checks after generation has been committed or uploaded as an artifact.

Use `npm run models:open` to open the generated model index through the default local surface. Agents and host apps can resolve the same target without launching a browser by running:

```bash
node scripts/open-artifact-review.mjs --json --print
```

In Codex app or Claude desktop, open the resolved artifact with the built-in browser/preview surface when available. If that surface blocks `file://` URLs, serve `generated/review/` through a local static server and open `/index.html` or the target artifact path from `http://127.0.0.1:<port>/`.

Every `setup-project` run writes `.skill-harness/setup-proof.json`. Treat it as machine-readable install evidence: it records the resolved setup directory, monorepo lift, package manager, requested/effective artifact profile, initialized tools, available check commands, generated paths, and skipped capabilities. It is intentionally descriptive; run the recorded check commands for live conformance.

### Validate installed agent dependencies

```bash
./skill-harness check --all
```

## Development

Required local tools:

- Go, using the version in [go.mod](go.mod)
- Node.js 22 or newer for generated artifact checks
- Python 3.12 or newer for suite maintenance scripts

Core checks:

```bash
go test ./...
npm run artifacts:check
python scripts/check_suite_drift.py --check
```

The GitHub Actions workflow in [.github/workflows/quality.yml](.github/workflows/quality.yml) also builds the CLI, validates JSON inputs, compiles Python scripts, checks suite drift, verifies generated model review HTML, and runs a hermetic `setup-project` smoke test.

## Included agents

- `requirements-analyst`
- `requirements-analyst-beads`
- `ux-researcher`
- `system-modeler`
- `system-modeler-beads`
- `software-architect`
- `software-architect-beads`
- `web-engineer`
- `backend-engineer`
- `test-designer`
- `test-designer-beads`
- `qa-automation-engineer`
- `quality-reviewer`
- `security-reviewer`
- `security-reviewer-beads`
- `pentest-reviewer`
- `delivery-manager`
- `delivery-manager-beads`
- `research-writer`
- `workflow-engineer`

Agent-to-skill mapping lives in [docs/agent-loadouts.md](docs/agent-loadouts.md).

## Included dependency repos

- [`45ck/agile-delivery-skills`](https://github.com/45ck/agile-delivery-skills)
- [`45ck/authentication-cryptography-skills`](https://github.com/45ck/authentication-cryptography-skills)
- [`45ck/automation-testing-skills`](https://github.com/45ck/automation-testing-skills)
- [`45ck/backend-persistence-skills`](https://github.com/45ck/backend-persistence-skills)
- [`45ck/business-analysis-skills`](https://github.com/45ck/business-analysis-skills)
- [`45ck/cloud-platform-operations-skills`](https://github.com/45ck/cloud-platform-operations-skills)
- [`45ck/code-review-inspection-skills`](https://github.com/45ck/code-review-inspection-skills)
- [`45ck/data-structures-algorithmic-reasoning-skills`](https://github.com/45ck/data-structures-algorithmic-reasoning-skills)
- [`45ck/deployment-release-skills`](https://github.com/45ck/deployment-release-skills)
- [`45ck/design-for-testability-skills`](https://github.com/45ck/design-for-testability-skills)
- [`45ck/documentation-evidence-skills`](https://github.com/45ck/documentation-evidence-skills)
- [`45ck/enterprise-architecture-integration-skills`](https://github.com/45ck/enterprise-architecture-integration-skills)
- [`45ck/fagan-inspection-skill`](https://github.com/45ck/fagan-inspection-skill)
- [`45ck/frontier-agent-playbook`](https://github.com/45ck/frontier-agent-playbook) - doctrine repo
- [`45ck/hci-review-skill`](https://github.com/45ck/hci-review-skill)
- [`45ck/llm-agent-security-skills`](https://github.com/45ck/llm-agent-security-skills)
- [`45ck/maintenance-evolution-skills`](https://github.com/45ck/maintenance-evolution-skills)
- [`45ck/marketing-product-skills`](https://github.com/45ck/marketing-product-skills)
- [`45ck/non-functional-testing-skills`](https://github.com/45ck/non-functional-testing-skills)
- [`45ck/oop-code-structure-skills`](https://github.com/45ck/oop-code-structure-skills)
- [`45ck/pentest-security-testing-skills`](https://github.com/45ck/pentest-security-testing-skills)
- [`45ck/project-management-skills`](https://github.com/45ck/project-management-skills)
- [`45ck/repo-branding-skill`](https://github.com/45ck/repo-branding-skill) - single-skill repo
- [`45ck/refactoring-code-smells-skills`](https://github.com/45ck/refactoring-code-smells-skills)
- [`45ck/research-literature-review-skills`](https://github.com/45ck/research-literature-review-skills)
- [`45ck/security-engineering-skills`](https://github.com/45ck/security-engineering-skills)
- [`45ck/software-architecture-skills`](https://github.com/45ck/software-architecture-skills)
- [`45ck/software-quality-skills`](https://github.com/45ck/software-quality-skills)
- [`45ck/uml-analysis-modelling-skills`](https://github.com/45ck/uml-analysis-modelling-skills)
- [`45ck/verification-test-design-skills`](https://github.com/45ck/verification-test-design-skills)
- [`45ck/web-engineering-skills`](https://github.com/45ck/web-engineering-skills)
- `coding-workflow-skills` (embedded)
- `design-tooling-skills` (embedded)
- `integration-tooling-skills` (embedded)
- `specgraph-skills` (embedded)
- `noslop-skills` (embedded)
- `developer-artifact-skills` (embedded)
- `demo-production-skills` (embedded)
- `agent-operating-skills` (embedded)

## Shared doctrine companion

[`45ck/frontier-agent-playbook`](https://github.com/45ck/frontier-agent-playbook) is the suite-wide doctrine repo for frontier-capability priors, agentic thinking, anti-fallback checks, and LLM-first architecture.

Use it in two ways:

- install the doctrine skills globally with `./skill-harness install --packs=frontier-agent-playbook --packs-only`
- copy `AGENTS.md`, `CLAUDE.md`, `AGENT_INSTRUCTIONS.md`, and `llms.txt` into a target project when you want repo-local doctrine surfaces

## Codex helper plugin

The optional [skill-harness helpers plugin](plugins/skill-harness-helpers/README.md) packages Codex-oriented helper skills for agent selection, handoff planning, Beads task shaping, and third-party skill intake. It is a local helper bundle, not a replacement for the main `.codex/agents/` loadouts installed by the CLI.

## Tooling repos used here

- [`45ck/skill-harness`](https://github.com/45ck/skill-harness)
- [`45ck/noslop`](https://github.com/45ck/noslop)
- [`45ck/agent-docs`](https://github.com/45ck/agent-docs)
- [`steveyegge/beads`](https://github.com/steveyegge/beads)

## Contributing and support

- Start with [CONTRIBUTING.md](CONTRIBUTING.md) for local setup, validation, scope, and pull request expectations.
- Use [SUPPORT.md](SUPPORT.md) for public support requests.
- Follow [SECURITY.md](SECURITY.md) for vulnerability reports and generated artifact safety expectations.
- Follow [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for collaboration standards.

## Full toolkit

The standard full toolkit for a new project is **specgraph + noslop**. The `setup-project` command installs both automatically. For manual installation steps or to install the matching skill packs (`specgraph-skills`, `noslop-skills`), see [AGENT_INSTRUCTIONS.md](AGENT_INSTRUCTIONS.md).

## For other agents

If another agent needs to install this repo or use it as the setup entrypoint, point it at [AGENT_INSTRUCTIONS.md](AGENT_INSTRUCTIONS.md).

## Important files

- [cmd/skill-harness/main.go](cmd/skill-harness/main.go)
- [AGENT_INSTRUCTIONS.md](AGENT_INSTRUCTIONS.md)
- [CONTRIBUTING.md](CONTRIBUTING.md)
- [SECURITY.md](SECURITY.md)
- [SUPPORT.md](SUPPORT.md)
- [docs/developer-artifacts.md](docs/developer-artifacts.md)
- [docs/agent-native-bootstrap.md](docs/agent-native-bootstrap.md)
- [docs/dependency-provenance.md](docs/dependency-provenance.md)
- [docs/agent-operating-skills.md](docs/agent-operating-skills.md)
- [docs/demo-production-media.md](docs/demo-production-media.md)
- [docs/third-party-skill-intake.md](docs/third-party-skill-intake.md)
- [packs/README.md](packs/README.md)
- [scripts/dependencies.json](scripts/dependencies.json)
- [scripts/external_skill_intake.py](scripts/external_skill_intake.py)
- [scripts/build_release.py](scripts/build_release.py)

## License

[MIT](LICENSE)
