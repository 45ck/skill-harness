# Contributing

Thanks for helping improve `skill-harness`.

This repository is the install and setup entrypoint for the 45ck agent workflow stack. Changes can affect user-global agent installs, target repo scaffolds, generated developer artifacts, and embedded skill packs, so keep contributions small, explicit, and evidence-backed.

## Before You Start

- Read [README.md](README.md) for the CLI surface and included packs.
- Read [AGENTS.md](AGENTS.md) if you are working with an agent in this repo.
- Check existing work in GitHub issues or the repo Beads issue tracker if you have `bd` available.
- For third-party skill ideas, follow [docs/third-party-skill-intake.md](docs/third-party-skill-intake.md) before proposing an import.

Maintainers use Beads for local task tracking. External contributors can still open GitHub issues and pull requests; maintainers will bridge accepted work into Beads when needed.

## Development Setup

Required tools:

- Go, using the version in [go.mod](go.mod)
- Node.js 22 or newer for artifact policy scripts
- Python 3.12 or newer for suite maintenance scripts

Build and test:

```bash
go build -o skill-harness ./cmd/skill-harness
go test ./...
npm run artifacts:check
python scripts/check_suite_drift.py --check
```

On Windows:

```powershell
go build -o skill-harness.exe .\cmd\skill-harness
go test ./...
npm run artifacts:check
python scripts/check_suite_drift.py --check
```

## Change Scope

Use the smallest change that solves the problem:

- CLI behavior belongs under [cmd/skill-harness](cmd/skill-harness).
- Shared dependency and agent mapping changes belong in [scripts/dependencies.json](scripts/dependencies.json), [scripts/agent_loadouts.json](scripts/agent_loadouts.json), and generated suite docs.
- Embedded pack changes belong under [packs](packs).
- Developer artifact policy changes usually require [docs/developer-artifacts.md](docs/developer-artifacts.md), model sources, manifest entries, and generated review HTML to stay aligned.

Do not vendor third-party skill repositories into this repo. Propose curated, first-party rewrites under `packs/` when an idea is worth adopting.

## Model And Artifact Impact

For every engineering change, check whether it affects code, APIs, workflows, dependencies, deployment behavior, UI structure, or agent behavior. If it does, update the relevant canonical source under [docs/artifacts/source/models](docs/artifacts/source/models) or explain why no model change is required in the issue or PR.

Generated HTML, screenshots, SVGs, PNGs, and comparison pages are review surfaces only. Source files and model diffs remain canonical.

## Pull Requests

Before opening a PR:

- Run the relevant tests and checks.
- Include the commands you ran and their results.
- Link the issue or explain the motivation.
- Call out model/artifact impact.
- Note any skipped checks with a reason.

Keep PRs focused. If a change spans CLI behavior, packs, documentation, and generated artifacts, explain the relationship clearly in the PR body.

