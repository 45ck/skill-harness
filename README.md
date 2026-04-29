# skill-harness

<p align="center">
  <img src="logo.svg" alt="skill-harness logo" width="128" height="128" />
</p>

`skill-harness` is the setup repo for the 45ck agent workflow stack. It is primarily an installer and generator: most downstream repos receive copied Claude/Codex agent definitions or project setup changes rather than importing `skill-harness` as an application runtime.

It does five jobs:

- installs the shared dependency-repo suite across pack repos, doctrine repos, and single-skill repos
- installs the shared Claude and Codex agents
- bootstraps project-level tooling with `@45ck/noslop` and `45ck/agent-docs`
- optionally bootstraps Beads, enabled by default in project setup
- hosts embedded packs for suite-local or incubating capabilities

## What it can set up

### Shared suite

- remote dependency repos plus embedded local packs under `packs/`
- shared skills synced into `~/.claude/skills/` and `~/.agents/skills/`
- supports repos exposed as `skills/`, packaged `.claude/.agents` mirrors, or a single root `SKILL.md`
- shared Claude agents
- shared Codex agents

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
- installs `@45ck/noslop` and `45ck/agent-docs`
- installs the Beads CLI if it is not already present
- runs `agent-docs init`
- runs `noslop init`
- runs `bd init`
- runs `agent-docs install-gates --quality`

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

### Download a release bundle

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
```

### Validate installed agent dependencies

```bash
./skill-harness check --all
```

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

## Shared doctrine companion

[`45ck/frontier-agent-playbook`](https://github.com/45ck/frontier-agent-playbook) is the suite-wide doctrine repo for frontier-capability priors, agentic thinking, anti-fallback checks, and LLM-first architecture.

Use it in two ways:

- install the doctrine skills globally with `./skill-harness install --packs=frontier-agent-playbook --packs-only`
- copy `AGENTS.md`, `CLAUDE.md`, `AGENT_INSTRUCTIONS.md`, and `llms.txt` into a target project when you want repo-local doctrine surfaces

## Tooling repos used here

- [`45ck/skill-harness`](https://github.com/45ck/skill-harness)
- [`45ck/noslop`](https://github.com/45ck/noslop)
- [`45ck/agent-docs`](https://github.com/45ck/agent-docs)
- [`steveyegge/beads`](https://github.com/steveyegge/beads)

## Full toolkit

The standard full toolkit for a new project is **specgraph + noslop**. The `setup-project` command installs both automatically. For manual installation steps or to install the matching skill packs (`specgraph-skills`, `noslop-skills`), see [AGENT_INSTRUCTIONS.md](AGENT_INSTRUCTIONS.md).

## For other agents

If another agent needs to install this repo or use it as the setup entrypoint, point it at [AGENT_INSTRUCTIONS.md](AGENT_INSTRUCTIONS.md).

## Important files

- [cmd/skill-harness/main.go](cmd/skill-harness/main.go)
- [AGENT_INSTRUCTIONS.md](AGENT_INSTRUCTIONS.md)
- [docs/third-party-skill-intake.md](docs/third-party-skill-intake.md)
- [packs/README.md](packs/README.md)
- [scripts/dependencies.json](scripts/dependencies.json)
- [scripts/external_skill_intake.py](scripts/external_skill_intake.py)
- [scripts/build_release.py](scripts/build_release.py)

## License

[MIT](LICENSE)
